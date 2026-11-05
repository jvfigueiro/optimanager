package database_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"optimanager/internal/database"
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

	if ver != 2 {
		t.Fatalf("expected user_version to be 2 on fresh database, got %d", ver)
	}

	// Verify that table audit_log exists
	var auditCount int
	err = db.Writer.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='audit_log'").Scan(&auditCount)
	if err != nil || auditCount != 1 {
		t.Fatalf("expected audit_log table to exist, got count %d, err: %v", auditCount, err)
	}

	// Verify that backup columns exist in config
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
}

func TestMigration_IdempotentOnReopen(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "idempotent_opti.db")

	// First open: runs migrations to version 2
	db1, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed first open: %v", err)
	}
	db1.Close()

	// Second open: should detect version 2 and skip migrations cleanly
	db2, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed second open: %v", err)
	}
	defer db2.Close()

	ver, err := db2.GetUserVersion()
	if err != nil {
		t.Fatalf("failed to get user_version: %v", err)
	}
	if ver != 2 {
		t.Fatalf("expected user_version to remain 2, got %d", ver)
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
	// Create minimal config table and set user_version = 1
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

	// 2. Now attempt to open with database.New, but simulate pre-migration backup failure
	// We can simulate this by making the backup directory unwritable or using a read-only directory
	backupDir := filepath.Join(tempDir, "backups")
	if err := os.MkdirAll(backupDir, 0555); err != nil {
		t.Fatalf("failed to create read-only backups dir: %v", err)
	}
	// If the user is running as root or permissions allow, make a file with the backupDir name
	_ = os.RemoveAll(backupDir)
	// Create a regular file at the path of the 'backups' directory so MkdirAll fails!
	if err := os.WriteFile(backupDir, []byte("blocking-file"), 0644); err != nil {
		t.Fatalf("failed to write blocking file: %v", err)
	}

	// database.New should fail because createPreMigrationBackup fails to create the directory
	failedDB, err := database.New(dbPath)
	if err == nil {
		failedDB.Close()
		t.Fatalf("expected database.New to fail when pre-migration backup directory creation fails")
	}

	// 3. Verify that migration 002 was NOT applied and user_version remains 1
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

	// Verify that table audit_log was NOT created
	var auditCount int
	_ = verifyDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='audit_log'").Scan(&auditCount)
	if auditCount != 0 {
		t.Fatalf("expected audit_log to not exist after aborted migration, got count %d", auditCount)
	}
}

func TestMigration_PreMigrationBackupCreatedSuccessfullyOnUpgrade(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "premigration_success.db")

	// 1. Create a version 1 database
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

	// 2. Open with database.New to trigger upgrade from 1 to 2
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("expected successful upgrade, got error: %v", err)
	}
	defer db.Close()

	// 3. Verify user_version is now 2
	ver, err := db.GetUserVersion()
	if err != nil || ver != 2 {
		t.Fatalf("expected user_version 2, got %d, err: %v", ver, err)
	}

	// 4. Verify that pre-migration backup file was created in backups/ with origin premigration
	backupDir := filepath.Join(tempDir, "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("failed to read backup dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected pre-migration backup file in %s, found none", backupDir)
	}

	foundPremigration := false
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".db" && len(e.Name()) > 10 {
			foundPremigration = true
		}
	}
	if !foundPremigration {
		t.Fatalf("expected to find premigration .db file in backup directory")
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

	// Insert audit log entries
	err = db.InsertAuditLog(1, "Admin Test", "192.168.1.100", "BACKUP_DOWNLOAD", "optimanager_backup_2026-11-05_12-00-00_manual.db")
	if err != nil {
		t.Fatalf("failed to insert audit log: %v", err)
	}

	err = db.InsertAuditLog(2, "Operador Test", "10.0.0.50", "BACKUP_DELETE", "optimanager_backup_2026-11-05_11-00-00_auto.db")
	if err != nil {
		t.Fatalf("failed to insert second audit log: %v", err)
	}

	// List audit logs
	logs, err := db.ListAuditLogs(10)
	if err != nil {
		t.Fatalf("failed to list audit logs: %v", err)
	}

	if len(logs) != 2 {
		t.Fatalf("expected 2 audit logs, got %d", len(logs))
	}

	// Verify order is DESC (newest first)
	if logs[0].Acao != "BACKUP_DELETE" || logs[0].UsuarioNome != "Operador Test" {
		t.Fatalf("expected newest log first, got: %+v", logs[0])
	}
	if logs[1].Acao != "BACKUP_DOWNLOAD" || logs[1].IP != "192.168.1.100" {
		t.Fatalf("expected second log, got: %+v", logs[1])
	}
}
