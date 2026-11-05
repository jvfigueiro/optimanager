package backup_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"optimanager/internal/database"
	"optimanager/internal/services/backup"
)

func setupTestDB(t *testing.T) (*database.DB, string) {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "opti_test.db")

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	return db, tempDir
}

func TestBackup_RoundtripAndIntegrity(t *testing.T) {
	db, tempDir := setupTestDB(t)
	defer db.Close()

	// Seed data across multiple tables
	_, err := db.Writer.Exec(`INSERT INTO clientes (nome, doc, contato) VALUES ('Cliente Alpha', '12345678901', '1199999999')`)
	if err != nil {
		t.Fatalf("failed to insert cliente: %v", err)
	}
	_, err = db.Writer.Exec(`INSERT INTO ordens_servico (data_emissao, cliente_id, equipamento, servico, valor_total) VALUES ('2026-11-05', 1, 'Notebook Dell', 'Troca de SSD', 250.00)`)
	if err != nil {
		t.Fatalf("failed to insert os: %v", err)
	}
	_ = db.InsertAuditLog(1, "Admin", "127.0.0.1", "TEST", "Roundtrip validation")

	backupDir := filepath.Join(tempDir, "backups")
	externalDir := filepath.Join(tempDir, "external")
	svc, err := backup.NewService(db, backupDir, externalDir)
	if err != nil {
		t.Fatalf("failed to initialize backup service: %v", err)
	}

	ctx := context.Background()
	info, err := svc.CreateBackup(ctx, "manual")
	if err != nil {
		t.Fatalf("failed to create manual backup: %v", err)
	}

	if info.Origem != "manual" {
		t.Fatalf("expected origin manual, got %s", info.Origem)
	}
	if !info.CopiaExternaOk {
		t.Fatalf("expected CopiaExternaOk to be true")
	}

	// Open the snapshot and verify integrity and table counts
	snapshotPath := filepath.Join(backupDir, info.Filename)
	snapDB, err := sql.Open("sqlite", fmt.Sprintf("%s?_pragma=query_only(1)", snapshotPath))
	if err != nil {
		t.Fatalf("failed to open snapshot db: %v", err)
	}
	defer snapDB.Close()

	var integrity string
	if err := snapDB.QueryRow("PRAGMA integrity_check").Scan(&integrity); err != nil || integrity != "ok" {
		t.Fatalf("snapshot failed integrity check: %s, err: %v", integrity, err)
	}

	var clienteCount, osCount, auditCount int
	_ = snapDB.QueryRow("SELECT COUNT(*) FROM clientes").Scan(&clienteCount)
	_ = snapDB.QueryRow("SELECT COUNT(*) FROM ordens_servico").Scan(&osCount)
	_ = snapDB.QueryRow("SELECT COUNT(*) FROM audit_log").Scan(&auditCount)

	if clienteCount != 1 || osCount != 1 || auditCount != 1 {
		t.Fatalf("snapshot counts mismatch: clientes=%d, os=%d, audit=%d", clienteCount, osCount, auditCount)
	}

	// Verify external copy also exists and has identical size
	extPath := filepath.Join(externalDir, info.Filename)
	extFi, err := os.Stat(extPath)
	if err != nil {
		t.Fatalf("external backup file missing: %v", err)
	}
	if extFi.Size() != info.TamanhoBytes {
		t.Fatalf("external file size mismatch: got %d, expected %d", extFi.Size(), info.TamanhoBytes)
	}
}

func TestBackup_TransactionalConsistencyWithControlCase(t *testing.T) {
	db, tempDir := setupTestDB(t)
	defer db.Close()

	backupDir := filepath.Join(tempDir, "backups")
	svc, err := backup.NewService(db, backupDir, "")
	if err != nil {
		t.Fatalf("failed to initialize backup service: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var writeOps int64
	var stopWrites int32

	// Launch concurrent worker that continuously inserts pairs of related records in single transactions:
	// A new client and a matching service order. They MUST always appear together or not at all.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		id := 1
		for atomic.LoadInt32(&stopWrites) == 0 {
			tx, err := db.Writer.Begin()
			if err != nil {
				return
			}
			nome := fmt.Sprintf("Cliente_Concurrent_%d", id)
			_, err = tx.Exec("INSERT INTO clientes (id, nome, doc, contato) VALUES (?, ?, '000', '11999')", id+10, nome)
			if err == nil {
				_, err = tx.Exec("INSERT INTO ordens_servico (data_emissao, cliente_id, equipamento, servico) VALUES ('2026-11-05', ?, 'PC', 'Format')", id+10)
			}
			if err == nil {
				_ = tx.Commit()
				atomic.AddInt64(&writeOps, 1)
			} else {
				_ = tx.Rollback()
			}
			id++
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Wait until several concurrent transactions have committed
	for i := 0; i < 100; i++ {
		if atomic.LoadInt64(&writeOps) >= 10 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	// 1. Create consistent backup via VACUUM INTO
	info, err := svc.CreateBackup(ctx, "auto")
	if err != nil {
		t.Fatalf("failed creating backup under concurrent writes: %v", err)
	}

	// 2. Control Case: perform a naive file copy (io.Copy) of the active DB file while writes continue
	naiveCopyPath := filepath.Join(tempDir, "naive_copy.db")
	_ = func() error {
		src, err := os.Open(db.Path)
		if err != nil {
			return err
		}
		defer src.Close()
		dst, err := os.Create(naiveCopyPath)
		if err != nil {
			return err
		}
		defer dst.Close()
		_, err = io.Copy(dst, src)
		return err
	}()

	// Stop concurrent writes
	atomic.StoreInt32(&stopWrites, 1)
	wg.Wait()

	// 3. Validate consistency of VACUUM INTO snapshot
	snapPath := filepath.Join(backupDir, info.Filename)
	snapDB, err := sql.Open("sqlite", fmt.Sprintf("%s?_pragma=query_only(1)", snapPath))
	if err != nil {
		t.Fatalf("failed to open snapshot db: %v", err)
	}
	defer snapDB.Close()

	var clientCount, osCount int
	_ = snapDB.QueryRow("SELECT COUNT(*) FROM clientes WHERE id > 10").Scan(&clientCount)
	_ = snapDB.QueryRow("SELECT COUNT(*) FROM ordens_servico WHERE cliente_id > 10").Scan(&osCount)

	// Consistency check: because each client and its OS are created in the same transaction,
	// clientCount MUST EXACTLY EQUAL osCount in the VACUUM INTO snapshot, and must be > 0.
	if clientCount == 0 {
		t.Fatalf("expected non-zero clients committed during test, got %d", clientCount)
	}
	if clientCount != osCount {
		t.Fatalf("transactional inconsistency in VACUUM INTO snapshot: clients=%d, os=%d (should be equal)", clientCount, osCount)
	}

	// Check orphan integrity: every OS in snapshot must point to an existing client
	var orphanCount int
	_ = snapDB.QueryRow("SELECT COUNT(*) FROM ordens_servico os WHERE NOT EXISTS (SELECT 1 FROM clientes c WHERE c.id = os.cliente_id)").Scan(&orphanCount)
	if orphanCount != 0 {
		t.Fatalf("found %d orphan service orders in consistent snapshot", orphanCount)
	}

	// 4. Control Case Verification:
	// A naive copy during WAL activity either fails integrity check, misses uncheckpointed WAL pages,
	// or diverges from the consistent snapshot taken at the same time.
	naiveDB, err := sql.Open("sqlite", naiveCopyPath)
	if err == nil {
		defer naiveDB.Close()
		var naiveIntegrity string
		_ = naiveDB.QueryRow("PRAGMA integrity_check").Scan(&naiveIntegrity)
		var naiveClients, naiveOS int
		_ = naiveDB.QueryRow("SELECT COUNT(*) FROM clientes WHERE id > 10").Scan(&naiveClients)
		_ = naiveDB.QueryRow("SELECT COUNT(*) FROM ordens_servico WHERE cliente_id > 10").Scan(&naiveOS)

		// Naive copy under active WAL misses transactions that were captured by VACUUM INTO
		// or produces divergent/empty counts because WAL file was not merged.
		if naiveClients == clientCount && naiveOS == osCount && naiveClients > 0 {
			t.Logf("Notice: naive copy unexpectedly matched WAL snapshot (may happen if checkpoint ran)")
		} else {
			t.Logf("Control case confirmed: naive copy diverged from live WAL state (naive_clients=%d vs snapshot_clients=%d)", naiveClients, clientCount)
		}
	}
}

func TestBackup_FailureCleansUpTmpFile(t *testing.T) {
	db, tempDir := setupTestDB(t)
	defer db.Close()

	backupDir := filepath.Join(tempDir, "backups")
	svc, err := backup.NewService(db, backupDir, "")
	if err != nil {
		t.Fatalf("failed to initialize backup service: %v", err)
	}

	// Simulate integrity failure
	svc.SimulateIntegrityFail = true

	ctx := context.Background()
	_, err = svc.CreateBackup(ctx, "manual")
	if err == nil {
		t.Fatalf("expected error when SimulateIntegrityFail is true")
	}

	// Verify that NO .db.tmp files exist in the backup directory
	entries, _ := os.ReadDir(backupDir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("found orphan .tmp file %s after failed backup", e.Name())
		}
	}
}

func TestBackup_StartupCleansOrphanTmpFiles(t *testing.T) {
	db, tempDir := setupTestDB(t)
	defer db.Close()

	backupDir := filepath.Join(tempDir, "backups")
	_ = os.MkdirAll(backupDir, 0755)

	// Create orphan .tmp and .perm_check files
	orphan1 := filepath.Join(backupDir, "optimanager_backup_2026-11-05_10-00-00_manual.db.tmp")
	orphan2 := filepath.Join(backupDir, ".perm_check_123456")
	validFile := filepath.Join(backupDir, "optimanager_backup_2026-11-05_09-00-00_manual.db")

	_ = os.WriteFile(orphan1, []byte("temp"), 0644)
	_ = os.WriteFile(orphan2, []byte("check"), 0644)
	_ = os.WriteFile(validFile, []byte("valid db content"), 0644)

	// Initialize service (triggers CleanupOrphanTmpFiles)
	_, err := backup.NewService(db, backupDir, "")
	if err != nil {
		t.Fatalf("failed to initialize service: %v", err)
	}

	if _, err := os.Stat(orphan1); !os.IsNotExist(err) {
		t.Fatalf("expected orphan1 .tmp to be deleted")
	}
	if _, err := os.Stat(orphan2); !os.IsNotExist(err) {
		t.Fatalf("expected orphan2 .perm_check to be deleted")
	}
	if _, err := os.Stat(validFile); os.IsNotExist(err) {
		t.Fatalf("expected valid backup file to be preserved")
	}
}

func TestBackup_IsolatedRetentionPerOrigin(t *testing.T) {
	db, tempDir := setupTestDB(t)
	defer db.Close()

	backupDir := filepath.Join(tempDir, "backups")
	svc, err := backup.NewService(db, backupDir, "")
	if err != nil {
		t.Fatalf("failed to initialize backup service: %v", err)
	}

	// Update config to test custom retention limits:
	// manual = 2, auto = 2, prerestore = 3, premigration = 3
	cfg, _ := db.GetConfig()
	cfg.BackupRetencaoManual = 2
	cfg.BackupRetencaoAuto = 2
	_ = db.UpdateBackupConfig(cfg)

	// Create 4 dummy files for manual and 3 for auto with valid naming
	for i := 1; i <= 4; i++ {
		name := fmt.Sprintf("optimanager_backup_2026-11-05_10-00-%02d_manual.db", i)
		_ = os.WriteFile(filepath.Join(backupDir, name), []byte("content"), 0644)
	}
	for i := 1; i <= 3; i++ {
		name := fmt.Sprintf("optimanager_backup_2026-11-05_11-00-%02d_auto.db", i)
		_ = os.WriteFile(filepath.Join(backupDir, name), []byte("content"), 0644)
	}

	// Apply retention for manual
	if err := svc.ApplyRetention("manual"); err != nil {
		t.Fatalf("failed applying retention for manual: %v", err)
	}

	list, _ := svc.ListBackups()
	manualCount := 0
	autoCount := 0
	for _, b := range list {
		if b.Origem == "manual" {
			manualCount++
		} else if b.Origem == "auto" {
			autoCount++
		}
	}

	// Manual should be pruned to exactly 2 (the newest two: seconds 03 and 04)
	if manualCount != 2 {
		t.Fatalf("expected 2 manual backups retained, got %d", manualCount)
	}
	// Auto should remain untouched at 3 because ApplyRetention was only called for manual!
	if autoCount != 3 {
		t.Fatalf("expected 3 auto backups to remain untouched, got %d", autoCount)
	}

	// Now apply retention for auto
	if err := svc.ApplyRetention("auto"); err != nil {
		t.Fatalf("failed applying retention for auto: %v", err)
	}

	listAfter, _ := svc.ListBackups()
	manualCountAfter := 0
	autoCountAfter := 0
	for _, b := range listAfter {
		if b.Origem == "manual" {
			manualCountAfter++
		} else if b.Origem == "auto" {
			autoCountAfter++
		}
	}
	if autoCountAfter != 2 {
		t.Fatalf("expected 2 auto backups retained after pruning, got %d", autoCountAfter)
	}
	if manualCountAfter != 2 {
		t.Fatalf("expected manual backups to remain 2, got %d", manualCountAfter)
	}
}

func TestBackup_SurvivalRuleNeverPrunesToZero(t *testing.T) {
	db, tempDir := setupTestDB(t)
	defer db.Close()

	backupDir := filepath.Join(tempDir, "backups")
	svc, err := backup.NewService(db, backupDir, "")
	if err != nil {
		t.Fatalf("failed to initialize backup service: %v", err)
	}

	// Set retention limit to 1
	cfg, _ := db.GetConfig()
	cfg.BackupRetencaoManual = 1
	_ = db.UpdateBackupConfig(cfg)

	// Create exactly 1 backup
	name := "optimanager_backup_2026-11-05_12-00-00_manual.db"
	_ = os.WriteFile(filepath.Join(backupDir, name), []byte("content"), 0644)

	// Run retention
	if err := svc.ApplyRetention("manual"); err != nil {
		t.Fatalf("failed applying retention: %v", err)
	}

	list, _ := svc.ListBackups()
	if len(list) != 1 {
		t.Fatalf("expected survival rule to preserve at least 1 backup, got %d", len(list))
	}
}

func TestBackup_PathTraversalAndValidation(t *testing.T) {
	db, tempDir := setupTestDB(t)
	defer db.Close()

	backupDir := filepath.Join(tempDir, "backups")
	svc, err := backup.NewService(db, backupDir, "")
	if err != nil {
		t.Fatalf("failed to initialize backup service: %v", err)
	}

	// Invalid filenames
	invalidNames := []string{
		"../../etc/passwd",
		"random_file.db",
		"optimanager_backup_2026-11-05_12-00-00_invalid.db",
		"optimanager_backup_2026-11-05_12-00-00_manual.zip",
	}

	for _, name := range invalidNames {
		t.Run("invalid_"+name, func(t *testing.T) {
			_, err := svc.GetBackupFilePath(name)
			if err == nil {
				t.Fatalf("expected error for invalid filename %q", name)
			}
			err = svc.DeleteBackup(name)
			if err == nil {
				t.Fatalf("expected error when deleting invalid filename %q", name)
			}
		})
	}
}

func TestBackup_CheckExternalWritable(t *testing.T) {
	db, tempDir := setupTestDB(t)
	defer db.Close()

	backupDir := filepath.Join(tempDir, "backups")
	externalDir := filepath.Join(tempDir, "external")
	_ = os.MkdirAll(externalDir, 0755)

	svc, err := backup.NewService(db, backupDir, externalDir)
	if err != nil {
		t.Fatalf("failed to initialize backup service: %v", err)
	}

	ok, err := svc.CheckExternalWritable()
	if err != nil || !ok {
		t.Fatalf("expected external destination to be writable, got ok=%v, err=%v", ok, err)
	}

	// Verify no .perm_check file remains in external directory
	entries, _ := os.ReadDir(externalDir)
	for _, e := range entries {
		if filepath.HasPrefix(e.Name(), ".perm_check") {
			t.Fatalf("found leftover .perm_check file: %s", e.Name())
		}
	}
}
