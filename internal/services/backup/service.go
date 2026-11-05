package backup

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "time/tzdata"

	"optimanager/internal/database"
)

var (
	// Strict regex for backup files: optimanager_backup_YYYY-MM-DD_HH-mm-ss_[manual|auto|prerestore|premigration].db
	backupFilenameRegex = regexp.MustCompile(`^optimanager_backup_\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}_(manual|auto|prerestore|premigration)\.db$`)

	ErrInvalidFilename   = errors.New("invalid backup filename format")
	ErrPathTraversal     = errors.New("path traversal detected")
	ErrBackupNotFound    = errors.New("backup file not found")
	ErrInsufficientSpace = errors.New("insufficient disk space for backup")
	ErrInvalidOrigin     = errors.New("invalid backup origin")
)

type BackupInfo struct {
	Filename       string    `json:"filename"`
	Origem         string    `json:"origem"`
	TamanhoBytes   int64     `json:"tamanho_bytes"`
	TamanhoFmt     string    `json:"tamanho_formatado"`
	DataCriacao    time.Time `json:"data_criacao"`
	DataCriacaoFmt string    `json:"data_criacao_formatada"`
	CopiaExternaOk bool      `json:"copia_externa_ok"`
}

type Service struct {
	db          *database.DB
	localDir    string
	externalDir string
	mu          sync.Mutex
	tz          *time.Location

	// Test hooks to simulate conditions
	SimulateVacuumFail    bool
	SimulateIntegrityFail bool
}

// NewService initializes the backup service, ensures local directories exist and cleans up orphan .tmp files.
func NewService(db *database.DB, localDir, externalDir string) (*Service, error) {
	if localDir == "" {
		if db.Path != "" && db.Path != ":memory:" {
			localDir = filepath.Join(filepath.Dir(db.Path), "backups")
		} else {
			localDir = "./data/backups"
		}
	}

	if externalDir == "" {
		externalDir = strings.TrimSpace(os.Getenv("BACKUP_EXTERNAL_DIR"))
	}

	if err := os.MkdirAll(localDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local backup directory (%s): %w", localDir, err)
	}

	if externalDir != "" {
		_ = os.MkdirAll(externalDir, 0755)
	}

	tz, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		tz = time.Local
	}

	s := &Service{
		db:          db,
		localDir:    localDir,
		externalDir: externalDir,
		tz:          tz,
	}

	// Clean up any orphan *.db.tmp or .perm_check files left by abrupt shutdowns
	if err := s.CleanupOrphanTmpFiles(); err != nil {
		log.Printf("[BACKUP WARNING] Failed cleaning orphan tmp files on startup: %v", err)
	}

	return s, nil
}

// LocalDir returns the configured primary backup directory.
func (s *Service) LocalDir() string {
	return s.localDir
}

// ExternalDir returns the configured external backup directory.
func (s *Service) ExternalDir() string {
	return s.externalDir
}

// CleanupOrphanTmpFiles removes any leftover *.db.tmp and .perm_check* files.
func (s *Service) CleanupOrphanTmpFiles() error {
	dirs := []string{s.localDir}
	if s.externalDir != "" {
		dirs = append(dirs, s.externalDir)
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasSuffix(name, ".db.tmp") || strings.HasPrefix(name, ".perm_check") {
				p := filepath.Join(dir, name)
				if err := os.Remove(p); err == nil {
					log.Printf("[BACKUP CLEANUP] Removed orphan temporary file: %s", p)
				}
			}
		}
	}
	return nil
}

// CheckExternalWritable executes a single test write using a unique filename (.perm_check_<timestamp>_<rand>).
func (s *Service) CheckExternalWritable() (bool, error) {
	if s.externalDir == "" {
		return false, nil
	}

	randBytes := make([]byte, 4)
	_, _ = rand.Read(randBytes)
	uniqueName := fmt.Sprintf(".perm_check_%d_%x", time.Now().UnixNano(), randBytes)
	testFile := filepath.Join(s.externalDir, uniqueName)

	f, err := os.OpenFile(testFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		nowStr := time.Now().In(s.tz).Format(time.RFC3339)
		_ = s.db.SetBackupExternalWarning(fmt.Sprintf("Permissão de escrita negada no destino externo: %v", err), nowStr)
		return false, err
	}
	_, _ = f.WriteString("opti_check\n")
	_ = f.Close()
	defer os.Remove(testFile)

	// Writing succeeded, clear previous external warning
	_ = s.db.SetBackupExternalWarning("", "")
	return true, nil
}

// CreateBackup creates a consistent snapshot of the active SQLite database using VACUUM INTO.
func (s *Service) CreateBackup(ctx context.Context, origem string) (*BackupInfo, error) {
	origem = strings.ToLower(strings.TrimSpace(origem))
	if origem != "manual" && origem != "auto" && origem != "prerestore" && origem != "premigration" {
		return nil, fmt.Errorf("%w: %s", ErrInvalidOrigin, origem)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Check local free space (require >= 2x active DB size)
	dbSize := s.getActiveDatabaseSize()
	minRequired := dbSize * 2
	if minRequired < 10*1024*1024 {
		minRequired = 10 * 1024 * 1024 // At least 10 MB minimum
	}

	freeSpaceLocal, err := getAvailableDiskSpace(s.localDir)
	if err == nil && freeSpaceLocal > 0 && freeSpaceLocal < minRequired {
		return nil, fmt.Errorf("%w: destino local possui %s disponíveis, necessário no mínimo %s",
			ErrInsufficientSpace, formatBytes(int64(freeSpaceLocal)), formatBytes(int64(minRequired)))
	}

	// 2. Check external destination free space if configured
	if s.externalDir != "" {
		freeSpaceExt, err := getAvailableDiskSpace(s.externalDir)
		nowStr := time.Now().In(s.tz).Format(time.RFC3339)
		if err == nil && freeSpaceExt > 0 && freeSpaceExt < minRequired {
			_ = s.db.SetBackupExternalWarning(fmt.Sprintf("Espaço insuficiente no destino externo (%s disponíveis, necessário %s)",
				formatBytes(int64(freeSpaceExt)), formatBytes(int64(minRequired))), nowStr)
		}
	}

	// 3. Generate filename with collision resolution (up to 3 retries with context cancellation)
	var filename string
	for attempt := 0; attempt < 3; attempt++ {
		now := time.Now().In(s.tz)
		candidate := fmt.Sprintf("optimanager_backup_%s_%s.db", now.Format("2006-01-02_15-04-05"), origem)
		finalPath := filepath.Join(s.localDir, candidate)

		if _, err := os.Stat(finalPath); os.IsNotExist(err) {
			filename = candidate
			break
		}

		// Wait 1 second respecting context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}

	if filename == "" {
		return nil, errors.New("failed to resolve unique backup filename after 3 attempts")
	}

	tmpPath := filepath.Join(s.localDir, filename+".tmp")
	finalPath := filepath.Join(s.localDir, filename)

	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpPath)
		}
	}()

	// 4. Execute VACUUM INTO
	if s.SimulateVacuumFail {
		return nil, errors.New("simulated VACUUM INTO failure")
	}

	vacuumSQL := fmt.Sprintf("VACUUM INTO '%s'", tmpPath)
	if _, err := s.db.Writer.ExecContext(ctx, vacuumSQL); err != nil {
		nowStr := time.Now().In(s.tz).Format(time.RFC3339)
		_ = s.db.SetBackupError(fmt.Sprintf("Falha ao gerar snapshot: %v", err), nowStr)
		return nil, fmt.Errorf("failed to execute VACUUM INTO: %w", err)
	}

	// 5. Verify integrity with PRAGMA integrity_check
	if s.SimulateIntegrityFail {
		return nil, errors.New("simulated integrity check failure")
	}

	checkDB, err := sql.Open("sqlite", fmt.Sprintf("%s?_pragma=query_only(1)&_pragma=busy_timeout(5000)", tmpPath))
	if err != nil {
		nowStr := time.Now().In(s.tz).Format(time.RFC3339)
		_ = s.db.SetBackupError(fmt.Sprintf("Falha ao abrir snapshot para validação: %v", err), nowStr)
		return nil, fmt.Errorf("failed to open snapshot for integrity check: %w", err)
	}

	var checkResult string
	err = checkDB.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&checkResult)
	checkDB.Close()
	if err != nil {
		nowStr := time.Now().In(s.tz).Format(time.RFC3339)
		_ = s.db.SetBackupError(fmt.Sprintf("Falha na consulta de integridade: %v", err), nowStr)
		return nil, fmt.Errorf("integrity check query failed: %w", err)
	}
	if checkResult != "ok" {
		nowStr := time.Now().In(s.tz).Format(time.RFC3339)
		_ = s.db.SetBackupError(fmt.Sprintf("Snapshot corrompido: %s", checkResult), nowStr)
		return nil, fmt.Errorf("integrity check failed with status: %s", checkResult)
	}

	// 6. Atomically rename .tmp to final
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return nil, fmt.Errorf("failed to finalize backup file: %w", err)
	}
	success = true

	// 7. Copy to external destination if configured
	externalOk := false
	if s.externalDir != "" {
		if err := s.copyToExternal(finalPath, filename); err != nil {
			nowStr := time.Now().In(s.tz).Format(time.RFC3339)
			_ = s.db.SetBackupExternalWarning(fmt.Sprintf("Falha ao copiar para destino externo: %v", err), nowStr)
			log.Printf("[BACKUP WARNING] Failed copying backup to external dir: %v", err)
		} else {
			externalOk = true
			_ = s.db.SetBackupExternalWarning("", "")
		}
	}

	// 8. Update database status
	nowStr := time.Now().In(s.tz).Format(time.RFC3339)
	if origem == "auto" {
		_ = s.db.SetBackupSuccess(nowStr)
	}

	// 9. Apply retention policy separately for this origin
	if err := s.ApplyRetention(origem); err != nil {
		log.Printf("[BACKUP WARNING] Failed applying retention for origin %s: %v", origem, err)
	}

	fi, err := os.Stat(finalPath)
	size := int64(0)
	modTime := time.Now().In(s.tz)
	if err == nil {
		size = fi.Size()
		modTime = fi.ModTime().In(s.tz)
	}

	return &BackupInfo{
		Filename:       filename,
		Origem:         origem,
		TamanhoBytes:   size,
		TamanhoFmt:     formatBytes(size),
		DataCriacao:    modTime,
		DataCriacaoFmt: modTime.Format("02/01/2006 15:04:05"),
		CopiaExternaOk: externalOk,
	}, nil
}

// ApplyRetention applies the origin-specific retention limit.
// It never prunes if remaining files for that origin would drop below 1.
func (s *Service) ApplyRetention(origem string) error {
	cfg, err := s.db.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to read config for retention: %w", err)
	}

	limit := 7
	switch origem {
	case "manual":
		limit = cfg.BackupRetencaoManual
		if limit < 1 || limit > 60 {
			limit = 5
		}
	case "auto":
		limit = cfg.BackupRetencaoAuto
		if limit < 1 || limit > 60 {
			limit = 7
		}
	case "prerestore":
		limit = cfg.BackupRetencaoPrerestore
		if limit < 3 || limit > 60 {
			limit = 3
		}
	case "premigration":
		limit = cfg.BackupRetencaoPremigration
		if limit < 3 || limit > 60 {
			limit = 3
		}
	}

	entries, err := os.ReadDir(s.localDir)
	if err != nil {
		return err
	}

	suffix := fmt.Sprintf("_%s.db", origem)
	var matchedFiles []os.FileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if backupFilenameRegex.MatchString(name) && strings.HasSuffix(name, suffix) {
			if fi, err := e.Info(); err == nil {
				matchedFiles = append(matchedFiles, fi)
			}
		}
	}

	// Sort chronologically ascending (oldest first)
	sort.Slice(matchedFiles, func(i, j int) bool {
		return matchedFiles[i].Name() < matchedFiles[j].Name()
	})

	// Rule: never prune if total <= 1 or count <= limit
	if len(matchedFiles) <= 1 || len(matchedFiles) <= limit {
		return nil
	}

	numToDelete := len(matchedFiles) - limit
	// Ensure at least 1 survives
	if len(matchedFiles)-numToDelete < 1 {
		numToDelete = len(matchedFiles) - 1
	}

	for i := 0; i < numToDelete; i++ {
		toRemove := matchedFiles[i].Name()
		localFile := filepath.Join(s.localDir, toRemove)
		if err := os.Remove(localFile); err == nil {
			log.Printf("[BACKUP RETENTION] Pruned old %s backup: %s", origem, toRemove)
		}

		if s.externalDir != "" {
			extFile := filepath.Join(s.externalDir, toRemove)
			_ = os.Remove(extFile)
		}
	}

	return nil
}

// ListBackups lists all valid backups in the local directory, ignoring unmatching files.
func (s *Service) ListBackups() ([]BackupInfo, error) {
	entries, err := os.ReadDir(s.localDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var list []BackupInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		matches := backupFilenameRegex.FindStringSubmatch(name)
		if matches == nil {
			// Ignore files that don't match strict backup regex
			continue
		}

		fi, err := e.Info()
		if err != nil {
			continue
		}

		origem := matches[1]
		modTime := fi.ModTime().In(s.tz)

		// Check if external copy exists
		extOk := false
		if s.externalDir != "" {
			extPath := filepath.Join(s.externalDir, name)
			if _, err := os.Stat(extPath); err == nil {
				extOk = true
			}
		}

		list = append(list, BackupInfo{
			Filename:       name,
			Origem:         origem,
			TamanhoBytes:   fi.Size(),
			TamanhoFmt:     formatBytes(fi.Size()),
			DataCriacao:    modTime,
			DataCriacaoFmt: modTime.Format("02/01/2006 15:04:05"),
			CopiaExternaOk: extOk,
		})
	}

	// Sort descending (newest first)
	sort.Slice(list, func(i, j int) bool {
		return list[i].Filename > list[j].Filename
	})

	return list, nil
}

// DeleteBackup removes a backup file from local and external destinations.
func (s *Service) DeleteBackup(filename string) error {
	filename = filepath.Base(filename)
	if !backupFilenameRegex.MatchString(filename) {
		return ErrInvalidFilename
	}

	localPath := filepath.Join(s.localDir, filename)
	rel, err := filepath.Rel(s.localDir, localPath)
	if err != nil || strings.HasPrefix(rel, "..") || rel != filename {
		return ErrPathTraversal
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return ErrBackupNotFound
	}

	if err := os.Remove(localPath); err != nil {
		return fmt.Errorf("failed to delete local backup file: %w", err)
	}

	if s.externalDir != "" {
		extPath := filepath.Join(s.externalDir, filename)
		_ = os.Remove(extPath)
	}

	return nil
}

// GetBackupFilePath validates and returns the safe absolute path to a backup file.
func (s *Service) GetBackupFilePath(filename string) (string, error) {
	filename = filepath.Base(filename)
	if !backupFilenameRegex.MatchString(filename) {
		return "", ErrInvalidFilename
	}

	targetPath := filepath.Join(s.localDir, filename)
	rel, err := filepath.Rel(s.localDir, targetPath)
	if err != nil || strings.HasPrefix(rel, "..") || rel != filename {
		return "", ErrPathTraversal
	}

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return "", ErrBackupNotFound
	}

	return targetPath, nil
}

func (s *Service) copyToExternal(srcPath, filename string) error {
	if s.externalDir == "" {
		return nil
	}

	destPath := filepath.Join(s.externalDir, filename)
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	if _, err := io.Copy(dest, src); err != nil {
		_ = os.Remove(destPath)
		return err
	}

	return dest.Sync()
}

func (s *Service) getActiveDatabaseSize() uint64 {
	if s.db.Path == "" || s.db.Path == ":memory:" {
		return 1024 * 1024 // 1 MB default
	}
	fi, err := os.Stat(s.db.Path)
	if err != nil {
		return 1024 * 1024
	}
	return uint64(fi.Size())
}

func getAvailableDiskSpace(dir string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return 0, err
	}
	return stat.Bavail * uint64(stat.Bsize), nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

