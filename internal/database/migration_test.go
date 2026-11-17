package database_test

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"optimanager/internal/database"
	"optimanager/internal/models"
)

func TestMigration_FreshDatabaseAppliesSequentially(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "fresh_opti.db")

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize fresh database: %v", err)
	}
	defer db.Close()

	ver, err := db.GetUserVersion()
	if err != nil {
		t.Fatalf("failed to get user_version: %v", err)
	}

	if ver != database.LatestMigrationVersion() {
		t.Fatalf("expected user_version to be %d on fresh database, got %d", database.LatestMigrationVersion(), ver)
	}

	// Verify that table audit_log exists
	var auditCount int
	err = db.Writer.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='audit_log'").Scan(&auditCount)
	if err != nil || auditCount != 1 {
		t.Fatalf("expected audit_log table to exist, got count %d, err: %v", auditCount, err)
	}

	// Verify that backup and foundation columns exist in config
	cfg, err := db.GetConfig()
	if err != nil {
		t.Fatalf("failed to get config: %v", err)
	}
	if cfg.BackupHorario != "02:00" {
		t.Fatalf("expected default backup_horario to be '02:00', got %q", cfg.BackupHorario)
	}
	if cfg.BackupRetencaoAuto != 7 {
		t.Fatalf("expected default backup_retencao_auto to be 7, got %d", cfg.BackupRetencaoAuto)
	}
	if cfg.BackupRetencaoManual != 5 {
		t.Fatalf("expected default backup_retencao_manual to be 5, got %d", cfg.BackupRetencaoManual)
	}
	if cfg.VTValorPadraoCentavos != 8000 {
		t.Fatalf("expected default vt_valor_padrao_centavos to be 8000, got %d", cfg.VTValorPadraoCentavos)
	}
	if cfg.CNPJ != "" {
		t.Fatalf("expected placeholder CNPJ to be cleared on migration 003, got %q", cfg.CNPJ)
	}
	if cfg.Versao != 1 {
		t.Fatalf("expected initial config versao to be 1, got %d", cfg.Versao)
	}
}

func TestMigration_IdempotentOnReopen(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "idempotent_opti.db")

	// First open: runs migrations to latest version
	db1, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed first open: %v", err)
	}
	db1.Close()

	// Second open: should detect latest version and skip migrations cleanly
	db2, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed second open: %v", err)
	}
	defer db2.Close()

	ver, err := db2.GetUserVersion()
	if err != nil {
		t.Fatalf("failed to get user_version: %v", err)
	}
	if ver != database.LatestMigrationVersion() {
		t.Fatalf("expected user_version to remain %d, got %d", database.LatestMigrationVersion(), ver)
	}
}

func TestMigration_PreMigrationBackupFailureAbortsBoot(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "premigration_abort.db")

	// 1. Create a version 1 database directly
	rawDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open raw sqlite db: %v", err)
	}
	initV1 := `
		CREATE TABLE config (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			fantasia TEXT NOT NULL DEFAULT 'Tech',
			razao_social TEXT NOT NULL DEFAULT 'Tech Ltda',
			cnpj TEXT NOT NULL DEFAULT '00.000.000/0001-00',
			responsavel TEXT NOT NULL DEFAULT 'Admin',
			fones TEXT NOT NULL DEFAULT '1199999999',
			email_contato TEXT NOT NULL DEFAULT '',
			endereco TEXT NOT NULL DEFAULT '',
			cidade TEXT NOT NULL DEFAULT 'SP',
			vt_valor_padrao REAL NOT NULL DEFAULT 80.00,
			termo_garantia TEXT NOT NULL DEFAULT '',
			termo_responsabilidade TEXT NOT NULL DEFAULT '',
			obs_padrao TEXT NOT NULL DEFAULT '',
			custom_presets TEXT NOT NULL DEFAULT '',
			logo_base64 TEXT NOT NULL DEFAULT '',
			cor_primaria TEXT NOT NULL DEFAULT '#0B5694',
			cor_secundaria TEXT NOT NULL DEFAULT '#083f6e'
		);
		INSERT INTO config (id) VALUES (1);
		PRAGMA user_version = 1;
	`
	if _, err := rawDB.Exec(initV1); err != nil {
		t.Fatalf("failed to prepare v1 database: %v", err)
	}
	rawDB.Close()

	// 2. Simulate pre-migration backup failure by creating a regular file at the path of 'backups'
	backupDir := filepath.Join(tempDir, "backups")
	if err := os.WriteFile(backupDir, []byte("blocking-file"), 0644); err != nil {
		t.Fatalf("failed to write blocking file: %v", err)
	}

	// database.New should fail because createPreMigrationBackup fails to create the directory
	failedDB, err := database.New(dbPath)
	if err == nil {
		failedDB.Close()
		t.Fatalf("expected database.New to fail when pre-migration backup directory creation fails")
	}

	// 3. Verify that migration 002/003 were NOT applied and user_version remains 1
	verifyDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open verify db: %v", err)
	}
	defer verifyDB.Close()

	var ver int
	if err := verifyDB.QueryRow("PRAGMA user_version").Scan(&ver); err != nil {
		t.Fatalf("failed to read user_version: %v", err)
	}
	if ver != 1 {
		t.Fatalf("expected user_version to remain 1 after aborted migration, got %d", ver)
	}

	var auditCount int
	_ = verifyDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='audit_log'").Scan(&auditCount)
	if auditCount != 0 {
		t.Fatalf("expected audit_log to not exist after aborted migration, got count %d", auditCount)
	}
}

func TestMigration_PreMigrationBackupCreatedSuccessfullyOnUpgrade(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "premigration_success.db")

	// 1. Create a version 1 database with sample records
	rawDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open raw sqlite db: %v", err)
	}
	initV1 := `
		CREATE TABLE config (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			fantasia TEXT NOT NULL DEFAULT 'Tech',
			razao_social TEXT NOT NULL DEFAULT 'Tech Ltda',
			cnpj TEXT NOT NULL DEFAULT '00.000.000/0001-00',
			responsavel TEXT NOT NULL DEFAULT 'Admin',
			fones TEXT NOT NULL DEFAULT '1199999999',
			email_contato TEXT NOT NULL DEFAULT '',
			endereco TEXT NOT NULL DEFAULT '',
			cidade TEXT NOT NULL DEFAULT 'SP',
			vt_valor_padrao REAL NOT NULL DEFAULT 80.00,
			termo_garantia TEXT NOT NULL DEFAULT '',
			termo_responsabilidade TEXT NOT NULL DEFAULT '',
			obs_padrao TEXT NOT NULL DEFAULT '',
			custom_presets TEXT NOT NULL DEFAULT '',
			logo_base64 TEXT NOT NULL DEFAULT '',
			cor_primaria TEXT NOT NULL DEFAULT '#0B5694',
			cor_secundaria TEXT NOT NULL DEFAULT '#083f6e'
		);
		INSERT INTO config (id, vt_valor_padrao) VALUES (1, 19.99);

		CREATE TABLE clientes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nome TEXT NOT NULL UNIQUE COLLATE NOCASE,
			doc TEXT NOT NULL DEFAULT '',
			contato TEXT NOT NULL DEFAULT '',
			desde_ano INTEGER NOT NULL DEFAULT 2024,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO clientes (id, nome, doc, contato) VALUES (1, 'Cliente V1', '', '11988887777');

		CREATE TABLE ordens_servico (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			data_emissao DATE NOT NULL,
			hora_emissao TEXT NOT NULL DEFAULT '',
			data_entrega TEXT NOT NULL DEFAULT '',
			hora_entrega TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'Concluído',
			cliente_id INTEGER NOT NULL,
			equipamento TEXT NOT NULL DEFAULT 'Notebook',
			servico TEXT NOT NULL,
			obs_tecnicas TEXT NOT NULL DEFAULT '',
			valor_total REAL NOT NULL DEFAULT 0.00,
			pagamento_status TEXT NOT NULL DEFAULT 'Pago',
			pagamento_forma TEXT NOT NULL DEFAULT 'Pix',
			pago_em DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (cliente_id) REFERENCES clientes(id) ON DELETE RESTRICT
		);
		INSERT INTO ordens_servico (id, data_emissao, status, cliente_id, servico, valor_total, pagamento_status)
		VALUES (1000, '2026-11-17', 'Concluído', 1, 'Limpeza', 19.99, 'Pago');

		CREATE TABLE usuarios (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nome TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE COLLATE NOCASE,
			senha_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'operador',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO usuarios (id, nome, email, senha_hash, role) VALUES (1, 'Op V1', 'op@test.com', 'hash', 'operador');

		PRAGMA user_version = 1;
	`
	if _, err := rawDB.Exec(initV1); err != nil {
		t.Fatalf("failed to prepare v1 database: %v", err)
	}
	rawDB.Close()

	// 2. Open with database.New to trigger upgrade from 1 to LatestMigrationVersion() (3)
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("expected successful upgrade, got error: %v", err)
	}
	defer db.Close()

	// 3. Verify user_version is now LatestMigrationVersion() (3)
	ver, err := db.GetUserVersion()
	if err != nil || ver != database.LatestMigrationVersion() {
		t.Fatalf("expected user_version %d, got %d, err: %v", database.LatestMigrationVersion(), ver, err)
	}

	// 4. Verify centavos conversion on config and ordens_servico
	cfg, err := db.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if cfg.VTValorPadraoCentavos != 1999 {
		t.Fatalf("expected vt_valor_padrao_centavos = 1999, got %d", cfg.VTValorPadraoCentavos)
	}
	if cfg.CNPJ != "" {
		t.Fatalf("expected fake placeholder CNPJ to be cleared, got %q", cfg.CNPJ)
	}

	osRec, err := db.GetOrdemByID(1000)
	if err != nil || osRec == nil {
		t.Fatalf("GetOrdemByID(1000) failed: %v", err)
	}
	if osRec.ValorTotalCentavos != 1999 || osRec.ValorTotal != 19.99 {
		t.Fatalf("expected order centavos=1999 and float=19.99, got %d / %v", osRec.ValorTotalCentavos, osRec.ValorTotal)
	}

	// 5. Verify operator user migrated to 'atendente' profile with permissions
	u, err := db.GetUsuarioByID(1)
	if err != nil || u == nil {
		t.Fatalf("GetUsuarioByID(1) failed: %v", err)
	}
	if u.Perfil != "atendente" || !u.Ativo || !u.HasPermission(models.PermVerValores) {
		t.Fatalf("expected active atendente with ver_valores, got %+v", u)
	}

	// 6. Verify inactive legacy technician creation (no email, no password)
	tech, err := db.CreateInactiveLegacyTechnician("Técnico Histórico Legado")
	if err != nil {
		t.Fatalf("CreateInactiveLegacyTechnician failed: %v", err)
	}
	if tech.Ativo || tech.Email != "" || tech.Perfil != "tecnico" {
		t.Fatalf("expected inactive tecnico with empty email, got %+v", tech)
	}

	// 7. Verify optimistic concurrency conflict on OrdemServico
	_, err = db.UpdateOrdem(1000, &models.UpdateOrdemRequest{
		Servico:            "Limpeza Atualizada 1",
		ValorTotalCentavos: 2500,
		Versao:             1,
	})
	if err != nil {
		t.Fatalf("first update with versao=1 failed: %v", err)
	}
	_, err = db.UpdateOrdem(1000, &models.UpdateOrdemRequest{
		Servico:            "Limpeza Stale",
		ValorTotalCentavos: 3000,
		Versao:             1, // stale version!
	})
	if !errors.Is(err, database.ErrVersionConflict) {
		t.Fatalf("expected ErrVersionConflict on stale update, got %v", err)
	}

	// 8. Verify pre-migration backup file was created in backups/
	backupDir := filepath.Join(tempDir, "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("expected pre-migration backup file in %s, err: %v", backupDir, err)
	}
}

func TestAuditLog_InsertAndList(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "audit_test.db")

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize db: %v", err)
	}
	defer db.Close()

	// Insert audit log entries (including PII canary that must be sanitized)
	err = db.InsertAuditLogWithEntity(1, "Admin Test", "192.168.1.100", "BACKUP_DOWNLOAD", "optimanager_backup_2026-11-05_12-00-00_manual.db", "backup", "1")
	if err != nil {
		t.Fatalf("failed to insert audit log: %v", err)
	}

	err = db.InsertAuditLog(2, "Operador Test", "10.0.0.50", "BACKUP_DELETE", "optimanager_backup_2026-11-05_11-00-00_auto.db")
	if err != nil {
		t.Fatalf("failed to insert second audit log: %v", err)
	}

	logs, err := db.ListAuditLogs(10)
	if err != nil {
		t.Fatalf("failed to list audit logs: %v", err)
	}

	if len(logs) != 2 {
		t.Fatalf("expected 2 audit logs, got %d", len(logs))
	}

	if logs[0].Acao != "BACKUP_DELETE" || logs[0].UsuarioNome != "Operador Test" {
		t.Fatalf("expected newest log first, got: %+v", logs[0])
	}
	if logs[1].Acao != "BACKUP_DOWNLOAD" || logs[1].Entidade != "backup" || logs[1].EntidadeID != "1" {
		t.Fatalf("expected second log with entity metadata, got: %+v", logs[1])
	}
}
