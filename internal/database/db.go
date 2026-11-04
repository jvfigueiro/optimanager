package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/001_initial_schema.sql
var initialSchemaSQL string

// DB holds the SQLite database connection pools.
type DB struct {
	Writer *sql.DB
	Reader *sql.DB
	Path   string
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

	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=cache_size(-8000)", dbPath)

	// Writer pool: strictly 1 connection to prevent concurrent write contention and locks
	writer, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open write connection: %w", err)
	}
	writer.SetMaxOpenConns(1)
	writer.SetMaxIdleConns(1)

	// Reader pool: up to 8 concurrent read connections
	reader, err := sql.Open("sqlite", dsn)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to open read connection: %w", err)
	}
	reader.SetMaxOpenConns(8)
	reader.SetMaxIdleConns(8)

	// Verify connectivity
	if err := writer.Ping(); err != nil {
		writer.Close()
		reader.Close()
		return nil, fmt.Errorf("ping failed on write connection: %w", err)
	}
	if err := reader.Ping(); err != nil {
		writer.Close()
		reader.Close()
		return nil, fmt.Errorf("ping failed on read connection: %w", err)
	}

	instance := &DB{
		Writer: writer,
		Reader: reader,
		Path:   dbPath,
	}

	// Run initial schema creation and incremental migrations
	if err := instance.migrate(); err != nil {
		instance.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	log.Printf("[DB] SQLite successfully initialized at %s (WAL mode, Writer: 1 conn, Reader: 8 conns)", dbPath)
	return instance, nil
}

func (d *DB) migrate() error {
	// Execute schema creation with IF NOT EXISTS
	if _, err := d.Writer.Exec(initialSchemaSQL); err != nil {
		return err
	}

	// Incremental migrations: check existing table columns in 'config' table before inserting or altering
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

		// Add missing columns dynamically if upgrading from an older schema version
		if !columns["email_contato"] {
			if _, err := d.Writer.Exec(`ALTER TABLE config ADD COLUMN email_contato TEXT NOT NULL DEFAULT ''`); err != nil {
				return fmt.Errorf("failed to add email_contato column: %w", err)
			}
			columns["email_contato"] = true
		}
		if !columns["endereco"] {
			if _, err := d.Writer.Exec(`ALTER TABLE config ADD COLUMN endereco TEXT NOT NULL DEFAULT ''`); err != nil {
				return fmt.Errorf("failed to add endereco column: %w", err)
			}
			columns["endereco"] = true
		}
		if !columns["termo_responsabilidade"] {
			if _, err := d.Writer.Exec(`ALTER TABLE config ADD COLUMN termo_responsabilidade TEXT NOT NULL DEFAULT 'Garantia legal de 90 dias conforme Art. 26 do CDC. Equipamentos não retirados em 90 dias estarão sujeitos a custas de guarda.'`); err != nil {
				return fmt.Errorf("failed to add termo_responsabilidade column: %w", err)
			}
			columns["termo_responsabilidade"] = true
		}
		if !columns["obs_padrao"] {
			if _, err := d.Writer.Exec(`ALTER TABLE config ADD COLUMN obs_padrao TEXT NOT NULL DEFAULT 'Aparelho recebido com marcas normais de uso. Backup de dados sob responsabilidade do cliente.'`); err != nil {
				return fmt.Errorf("failed to add obs_padrao column: %w", err)
			}
			columns["obs_padrao"] = true
		}
		if !columns["custom_presets"] {
			if _, err := d.Writer.Exec(`ALTER TABLE config ADD COLUMN custom_presets TEXT NOT NULL DEFAULT ''`); err != nil {
				return fmt.Errorf("failed to add custom_presets column: %w", err)
			}
			columns["custom_presets"] = true

			defaultPresets := `{"marcas":["Asus","Acer","Dell","Lenovo","HP","Apple","Samsung","LG","Positivo"],"modelos":["Notebook","Computador Desktop","All-in-One","Servidor","Monitor","Smartphone","Tablet","Outro"],"servicos":["Formatação e Instalação de Sistema","Limpeza Preventiva e Troca de Pasta Térmica","Troca de SSD / Upgrade de Memória RAM","Reparo de Placa-Mãe / Conector","Troca de Tela / Display","Remoção de Vírus e Otimização"]}`
			_, _ = d.Writer.Exec(`UPDATE config SET custom_presets = ? WHERE custom_presets = '' OR custom_presets IS NULL`, defaultPresets)
		}

		// Ensure default company row exists
		var configRowCount int
		_ = d.Writer.QueryRow(`SELECT COUNT(*) FROM config WHERE id = 1`).Scan(&configRowCount)
		if configRowCount == 0 {
			insertQuery := `INSERT OR IGNORE INTO config (id, fantasia, razao_social, cnpj, responsavel, fones, email_contato, endereco, cidade, vt_valor_padrao, termo_garantia, logo_base64, cor_primaria, cor_secundaria)
			VALUES (1, 'Opti Tecnologia', 'João Vítor Streb Figueiró', '38.136.905/0001-66', 'João Figueiró', '(55) 99127-2332 | (55) 99674-5762', '', '', 'Santa Maria - RS', 80.00, 'Garantia legal de 90 dias para serviços executados e peças aplicadas.', '', '#0B5694', '#083f6e')`
			if _, err := d.Writer.Exec(insertQuery); err != nil {
				return fmt.Errorf("failed to insert default config record: %w", err)
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
