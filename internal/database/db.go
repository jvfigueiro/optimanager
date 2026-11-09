package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/001_initial_schema.sql
var migration001SQL string

//go:embed migrations/002_backup_and_audit.sql
var migration002SQL string

type Migration struct {
	Version int
	Name    string
	SQL     string
}

var migrations = []Migration{
	{Version: 1, Name: "001_initial_schema", SQL: migration001SQL},
	{Version: 2, Name: "002_backup_and_audit", SQL: migration002SQL},
}

// LatestMigrationVersion returns the highest supported schema version.
func LatestMigrationVersion() int {
	if len(migrations) == 0 {
		return 0
	}
	return migrations[len(migrations)-1].Version
}

// DB holds the SQLite database connection pools and metadata.
type DB struct {
	Writer           *sql.DB
	Reader           *sql.DB
	Path             string
	PreMigrationHook func() error // Used in tests to simulate pre-migration backup failure
}

// New initializes the SQLite database with WAL mode and dedicated writer/reader pools.
func New(customPath string) (*DB, error) {
	dbPath := customPath
	if dbPath == "" {
		dbPath = os.Getenv("DB_PATH")
	}
	if dbPath == "" {
		if info, err := os.Stat("/data"); err == nil && info.IsDir() {
			dbPath = "/data/optimanager.db"
		} else {
			dbPath = "./data/optimanager.db"
		}
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for database (%s): %w", dir, err)
	}

	writer, reader, err := openPools(dbPath)
	if err != nil {
		return nil, err
	}

	instance := &DB{
		Writer: writer,
		Reader: reader,
		Path:   dbPath,
	}

	// Run initial schema creation and sequential incremental migrations
	if err := instance.migrate(); err != nil {
		instance.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	log.Printf("[DB] SQLite successfully initialized at %s (WAL mode, Writer: 1 conn, Reader: 8 conns)", dbPath)
	return instance, nil
}

func openPools(dbPath string) (*sql.DB, *sql.DB, error) {
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=cache_size(-8000)", dbPath)

	// Writer pool: strictly 1 connection to prevent concurrent write contention and locks
	writer, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open write connection: %w", err)
	}
	writer.SetMaxOpenConns(1)
	writer.SetMaxIdleConns(1)

	// Reader pool: up to 8 concurrent read connections
	reader, err := sql.Open("sqlite", dsn)
	if err != nil {
		writer.Close()
		return nil, nil, fmt.Errorf("failed to open read connection: %w", err)
	}
	reader.SetMaxOpenConns(8)
	reader.SetMaxIdleConns(8)

	// Verify connectivity
	if err := writer.Ping(); err != nil {
		writer.Close()
		reader.Close()
		return nil, nil, fmt.Errorf("ping failed on write connection: %w", err)
	}
	if err := reader.Ping(); err != nil {
		writer.Close()
		reader.Close()
		return nil, nil, fmt.Errorf("ping failed on read connection: %w", err)
	}

	return writer, reader, nil
}

// ReplaceDatabase safely swaps the active SQLite database file with a validated backup file,
// removes any stale WAL/SHM files, reopens the connection pools, and runs any pending migrations.
func (d *DB) ReplaceDatabase(srcPath string) error {
	// 1. Copy srcPath to a temporary swap file in the same directory as d.Path
	swapTmp := d.Path + ".swap.tmp"
	_ = os.Remove(swapTmp)

	in, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source backup file: %w", err)
	}
	out, err := os.OpenFile(swapTmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		in.Close()
		return fmt.Errorf("failed to create temporary swap file: %w", err)
	}
	if _, err := io.Copy(out, in); err != nil {
		in.Close()
		out.Close()
		_ = os.Remove(swapTmp)
		return fmt.Errorf("failed to copy backup data to swap file: %w", err)
	}
	in.Close()
	if err := out.Sync(); err != nil {
		out.Close()
		_ = os.Remove(swapTmp)
		return fmt.Errorf("failed to fsync swap file: %w", err)
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(swapTmp)
		return fmt.Errorf("failed to close swap file: %w", err)
	}

	// 2. Close active connection pools
	d.Close()

	// 3. Remove stale WAL and SHM files so old journal pages cannot corrupt the newly restored database
	_ = os.Remove(d.Path + "-wal")
	_ = os.Remove(d.Path + "-shm")

	// 4. Atomically replace the active database file
	if err := os.Rename(swapTmp, d.Path); err != nil {
		_ = os.Remove(swapTmp)
		return fmt.Errorf("failed to rename swap file over active database: %w", err)
	}

	// 5. Reopen connection pools
	writer, reader, err := openPools(d.Path)
	if err != nil {
		return fmt.Errorf("failed to reopen database pools after restore: %w", err)
	}
	d.Writer = writer
	d.Reader = reader

	// 6. Run migrations in case the restored backup is from an older schema version
	if err := d.migrate(); err != nil {
		return fmt.Errorf("failed to migrate restored database: %w", err)
	}

	log.Printf("[DB] Database successfully replaced and reopened from %s", filepath.Base(srcPath))
	return nil
}

// GetUserVersion reads the current PRAGMA user_version from SQLite.
func (d *DB) GetUserVersion() (int, error) {
	var ver int
	err := d.Writer.QueryRow("PRAGMA user_version").Scan(&ver)
	if err != nil {
		return 0, fmt.Errorf("failed to read PRAGMA user_version: %w", err)
	}
	return ver, nil
}

func (d *DB) migrate() error {
	currentVersion, err := d.GetUserVersion()
	if err != nil {
		return err
	}

	// If currentVersion == 0, check if this is an existing database from before user_version tracking
	if currentVersion == 0 {
		var tableCount int
		_ = d.Writer.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='config'`).Scan(&tableCount)
		if tableCount > 0 {
			// Legacy unversioned database: bring to v1 baseline and set user_version = 1
			if err := d.applyLegacyBaseline(); err != nil {
				return fmt.Errorf("failed to apply legacy baseline: %w", err)
			}
			if _, err := d.Writer.Exec("PRAGMA user_version = 1"); err != nil {
				return fmt.Errorf("failed to set baseline user_version = 1: %w", err)
			}
			currentVersion = 1
		}
	}

	targetVersion := migrations[len(migrations)-1].Version

	// If there are migrations to apply on an existing database (currentVersion >= 1),
	// create an automatic pre-migration backup before applying any new migrations.
	if currentVersion > 0 && currentVersion < targetVersion {
		if err := d.createPreMigrationBackup(); err != nil {
			return fmt.Errorf("pre-migration backup failed: %w; migrations aborted to protect database", err)
		}
	}

	// Sequentially execute pending migrations
	for _, m := range migrations {
		if m.Version > currentVersion {
			if err := d.applyMigration(m); err != nil {
				return err
			}
			currentVersion = m.Version
		}
	}

	return nil
}

func (d *DB) applyMigration(m Migration) error {
	tx, err := d.Writer.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for migration %d (%s): %w", m.Version, m.Name, err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(m.SQL); err != nil {
		return fmt.Errorf("failed to execute SQL for migration %d (%s): %w", m.Version, m.Name, err)
	}

	if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", m.Version)); err != nil {
		return fmt.Errorf("failed to set user_version to %d for migration (%s): %w", m.Version, m.Name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %d (%s): %w", m.Version, m.Name, err)
	}

	log.Printf("[DB] Successfully applied migration %03d_%s (user_version = %d)", m.Version, m.Name, m.Version)
	return nil
}

func (d *DB) createPreMigrationBackup() error {
	if d.PreMigrationHook != nil {
		if err := d.PreMigrationHook(); err != nil {
			return err
		}
	}

	// Skip if memory database
	if d.Path == "" || d.Path == ":memory:" {
		return nil
	}

	backupDir := filepath.Join(filepath.Dir(d.Path), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory (%s): %w", backupDir, err)
	}

	now := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("optimanager_backup_%s_premigration.db", now)
	tmpPath := filepath.Join(backupDir, filename+".tmp")
	finalPath := filepath.Join(backupDir, filename)

	defer func() {
		if _, err := os.Stat(tmpPath); err == nil {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := d.Writer.Exec(fmt.Sprintf("VACUUM INTO '%s'", tmpPath)); err != nil {
		return fmt.Errorf("pre-migration VACUUM INTO failed: %w", err)
	}

	checkDB, err := sql.Open("sqlite", fmt.Sprintf("%s?_pragma=query_only(1)", tmpPath))
	if err != nil {
		return fmt.Errorf("failed to open pre-migration snapshot for integrity check: %w", err)
	}
	var checkResult string
	err = checkDB.QueryRow("PRAGMA integrity_check").Scan(&checkResult)
	checkDB.Close()
	if err != nil {
		return fmt.Errorf("pre-migration integrity check query failed: %w", err)
	}
	if checkResult != "ok" {
		return fmt.Errorf("pre-migration integrity check failed with result: %s", checkResult)
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("failed to rename pre-migration backup to final: %w", err)
	}

	log.Printf("[DB] Successfully created pre-migration backup: %s", finalPath)
	return nil
}

func (d *DB) applyLegacyBaseline() error {
	// Incremental migrations for legacy databases
	var tableCount int
	err := d.Writer.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='config'`).Scan(&tableCount)
	if err != nil {
		return fmt.Errorf("failed to inspect sqlite_master: %w", err)
	}

	if tableCount > 0 {
		columns, err := d.getTableColumns("config")
		if err != nil {
			return fmt.Errorf("failed to get columns for config table: %w", err)
		}

		if !columns["email_contato"] {
			if _, err := d.Writer.Exec(`ALTER TABLE config ADD COLUMN email_contato TEXT NOT NULL DEFAULT ''`); err != nil {
				return fmt.Errorf("failed to add email_contato column: %w", err)
			}
		}
		if !columns["endereco"] {
			if _, err := d.Writer.Exec(`ALTER TABLE config ADD COLUMN endereco TEXT NOT NULL DEFAULT ''`); err != nil {
				return fmt.Errorf("failed to add endereco column: %w", err)
			}
		}
		if !columns["termo_responsabilidade"] {
			if _, err := d.Writer.Exec(`ALTER TABLE config ADD COLUMN termo_responsabilidade TEXT NOT NULL DEFAULT 'Garantia legal de 90 dias conforme Art. 26 do CDC. Equipamentos não retirados em 90 dias estarão sujeitos a custas de guarda.'`); err != nil {
				return fmt.Errorf("failed to add termo_responsabilidade column: %w", err)
			}
		}
		if !columns["obs_padrao"] {
			if _, err := d.Writer.Exec(`ALTER TABLE config ADD COLUMN obs_padrao TEXT NOT NULL DEFAULT 'Aparelho recebido com marcas normais de uso. Backup de dados sob responsabilidade do cliente.'`); err != nil {
				return fmt.Errorf("failed to add obs_padrao column: %w", err)
			}
		}
		if !columns["custom_presets"] {
			if _, err := d.Writer.Exec(`ALTER TABLE config ADD COLUMN custom_presets TEXT NOT NULL DEFAULT ''`); err != nil {
				return fmt.Errorf("failed to add custom_presets column: %w", err)
			}
			defaultPresets := `{"marcas":["Asus","Acer","Dell","Lenovo","HP","Apple","Samsung","LG","Positivo"],"modelos":["Notebook","Computador Desktop","All-in-One","Servidor","Monitor","Smartphone","Tablet","Outro"],"servicos":["Formatação e Instalação de Sistema","Limpeza Preventiva e Troca de Pasta Térmica","Troca de SSD / Upgrade de Memória RAM","Reparo de Placa-Mãe / Conector","Troca de Tela / Display","Remoção de Vírus e Otimização"]}`
			_, _ = d.Writer.Exec(`UPDATE config SET custom_presets = ? WHERE custom_presets = '' OR custom_presets IS NULL`, defaultPresets)
		}

		// Ensure default company row exists
		var configRowCount int
		_ = d.Writer.QueryRow(`SELECT COUNT(*) FROM config WHERE id = 1`).Scan(&configRowCount)
		if configRowCount == 0 {
			insertQuery := `INSERT OR IGNORE INTO config (id, fantasia, razao_social, cnpj, responsavel, fones, email_contato, endereco, cidade, vt_valor_padrao, termo_garantia, logo_base64, cor_primaria, cor_secundaria)
			VALUES (1, 'Minha Empresa Tech', 'Soluções em Tecnologia & Serviços Ltda', '00.000.000/0001-00', 'Administrador', '(11) 98765-4321', 'contato@empresa.com.br', 'Av. Paulista, 1000 - Bela Vista', 'São Paulo - SP', 80.00, 'Garantia legal de 90 dias para serviços executados e peças aplicadas.', '', '#0B5694', '#083f6e')`
			if _, err := d.Writer.Exec(insertQuery); err != nil {
				return fmt.Errorf("failed to insert default config record: %w", err)
			}
		}
	}

	// Incremental migrations for 'ordens_servico' table
	var osTableCount int
	_ = d.Writer.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='ordens_servico'`).Scan(&osTableCount)
	if osTableCount > 0 {
		osCols, err := d.getTableColumns("ordens_servico")
		if err != nil {
			return fmt.Errorf("failed to get columns for ordens_servico table: %w", err)
		}

		if !osCols["hora_emissao"] {
			if _, err := d.Writer.Exec(`ALTER TABLE ordens_servico ADD COLUMN hora_emissao TEXT NOT NULL DEFAULT ''`); err != nil {
				return fmt.Errorf("failed to add hora_emissao column: %w", err)
			}
		}
		if !osCols["data_entrega"] {
			if _, err := d.Writer.Exec(`ALTER TABLE ordens_servico ADD COLUMN data_entrega TEXT NOT NULL DEFAULT ''`); err != nil {
				return fmt.Errorf("failed to add data_entrega column: %w", err)
			}
		}
		if !osCols["hora_entrega"] {
			if _, err := d.Writer.Exec(`ALTER TABLE ordens_servico ADD COLUMN hora_entrega TEXT NOT NULL DEFAULT ''`); err != nil {
				return fmt.Errorf("failed to add hora_entrega column: %w", err)
			}
		}

		// Remove legacy CHECK constraints on status and pagamento_forma if present
		var osCreateSQL string
		_ = d.Writer.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='ordens_servico'`).Scan(&osCreateSQL)
		if strings.Contains(osCreateSQL, "CHECK (status IN") || strings.Contains(osCreateSQL, "CHECK (pagamento_forma IN") {
			rebuildSQL := `
				PRAGMA foreign_keys = OFF;
				BEGIN TRANSACTION;
				CREATE TABLE ordens_servico_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					data_emissao DATE NOT NULL,
					hora_emissao TEXT NOT NULL DEFAULT '',
					data_entrega TEXT NOT NULL DEFAULT '',
					hora_entrega TEXT NOT NULL DEFAULT '',
					status TEXT NOT NULL DEFAULT 'Avaliação',
					cliente_id INTEGER NOT NULL,
					equipamento TEXT NOT NULL DEFAULT 'N/A',
					servico TEXT NOT NULL,
					obs_tecnicas TEXT NOT NULL DEFAULT '',
					valor_total REAL NOT NULL DEFAULT 0.00,
					pagamento_status TEXT NOT NULL DEFAULT 'Pendente' CHECK (pagamento_status IN ('Pendente', 'Pago')),
					pagamento_forma TEXT NOT NULL DEFAULT 'Pix',
					pago_em DATETIME,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (cliente_id) REFERENCES clientes(id) ON DELETE RESTRICT
				);
				INSERT INTO ordens_servico_new (
					id, data_emissao, hora_emissao, data_entrega, hora_entrega, status, cliente_id,
					equipamento, servico, obs_tecnicas, valor_total,
					pagamento_status, pagamento_forma, pago_em, created_at
				)
				SELECT
					id, data_emissao, hora_emissao, data_entrega, hora_entrega, status, cliente_id,
					equipamento, servico, obs_tecnicas, valor_total,
					pagamento_status, pagamento_forma, pago_em, created_at
				FROM ordens_servico;
				DROP TABLE ordens_servico;
				ALTER TABLE ordens_servico_new RENAME TO ordens_servico;
				CREATE INDEX IF NOT EXISTS idx_os_data_emissao ON ordens_servico(data_emissao);
				CREATE INDEX IF NOT EXISTS idx_os_status ON ordens_servico(status);
				CREATE INDEX IF NOT EXISTS idx_os_pagamento_status ON ordens_servico(pagamento_status);
				CREATE INDEX IF NOT EXISTS idx_os_cliente_id ON ordens_servico(cliente_id);
				COMMIT;
				PRAGMA foreign_keys = ON;
			`
			if _, err := d.Writer.Exec(rebuildSQL); err != nil {
				return fmt.Errorf("failed to rebuild ordens_servico table: %w", err)
			}
		}
	}

	return nil
}

// getTableColumns inspects SQLite PRAGMA table_info to return a map of column names.
func (d *DB) getTableColumns(tableName string) (map[string]bool, error) {
	rows, err := d.Writer.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		cols[strings.ToLower(name)] = true
	}
	return cols, rows.Err()
}

// Close gracefully closes both write and read database pools.
func (d *DB) Close() {
	if d.Writer != nil {
		d.Writer.Close()
	}
	if d.Reader != nil {
		d.Reader.Close()
	}
}
