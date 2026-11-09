package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
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
	// Strict regex for backup files: optimanager_backup_YYYY-MM-DD_HH-mm-ss_[manual|auto|prerestore|premigration].[zip|db]
	backupFilenameRegex = regexp.MustCompile(`^optimanager_backup_\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}_(manual|auto|prerestore|premigration)\.(zip|db)$`)

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
			if strings.HasSuffix(name, ".db.tmp") || strings.HasSuffix(name, ".zip.tmp") || strings.HasPrefix(name, ".perm_check") || (strings.HasPrefix(name, ".restore_") && strings.HasSuffix(name, ".tmp")) {
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

	return s.createBackupLocked(ctx, origem)
}

func (s *Service) createBackupLocked(ctx context.Context, origem string) (*BackupInfo, error) {
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
		candidate := fmt.Sprintf("optimanager_backup_%s_%s.zip", now.Format("2006-01-02_15-04-05"), origem)
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

	tmpDBPath := filepath.Join(s.localDir, filename+".db.tmp")
	tmpZipPath := filepath.Join(s.localDir, filename+".tmp")
	finalPath := filepath.Join(s.localDir, filename)

	success := false
	defer func() {
		_ = os.Remove(tmpDBPath)
		if !success {
			_ = os.Remove(tmpZipPath)
		}
	}()

	// 4. Execute VACUUM INTO temporary SQLite snapshot
	if s.SimulateVacuumFail {
		return nil, errors.New("simulated VACUUM INTO failure")
	}

	vacuumSQL := fmt.Sprintf("VACUUM INTO '%s'", tmpDBPath)
	if _, err := s.db.Writer.ExecContext(ctx, vacuumSQL); err != nil {
		nowStr := time.Now().In(s.tz).Format(time.RFC3339)
		_ = s.db.SetBackupError(fmt.Sprintf("Falha ao gerar snapshot: %v", err), nowStr)
		return nil, fmt.Errorf("failed to execute VACUUM INTO: %w", err)
	}

	// 5. Verify integrity with PRAGMA integrity_check and read company logo from snapshot
	if s.SimulateIntegrityFail {
		return nil, errors.New("simulated integrity check failure")
	}

	checkDB, err := sql.Open("sqlite", fmt.Sprintf("%s?_pragma=query_only(1)&_pragma=busy_timeout(5000)", tmpDBPath))
	if err != nil {
		nowStr := time.Now().In(s.tz).Format(time.RFC3339)
		_ = s.db.SetBackupError(fmt.Sprintf("Falha ao abrir snapshot para validação: %v", err), nowStr)
		return nil, fmt.Errorf("failed to open snapshot for integrity check: %w", err)
	}

	var checkResult string
	err = checkDB.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&checkResult)
	var logoDataURI string
	_ = checkDB.QueryRowContext(ctx, "SELECT logo_base64 FROM config WHERE id = 1").Scan(&logoDataURI)
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

	// 6. Package database snapshot and company logo into a compressed .zip archive
	if err := createBackupZipArchive(tmpZipPath, tmpDBPath, logoDataURI, origem); err != nil {
		nowStr := time.Now().In(s.tz).Format(time.RFC3339)
		_ = s.db.SetBackupError(fmt.Sprintf("Falha ao compactar arquivo de backup (.zip): %v", err), nowStr)
		return nil, fmt.Errorf("failed to create zip backup archive: %w", err)
	}
	_ = os.Remove(tmpDBPath)

	// 7. Atomically rename .zip.tmp to final .zip
	if err := os.Rename(tmpZipPath, finalPath); err != nil {
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

	suffixZip := fmt.Sprintf("_%s.zip", origem)
	suffixDB := fmt.Sprintf("_%s.db", origem)
	var matchedFiles []os.FileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if backupFilenameRegex.MatchString(name) && (strings.HasSuffix(name, suffixZip) || strings.HasSuffix(name, suffixDB)) {
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

const (
	sqliteMagicHeader     = "SQLite format 3\x00"
	maxUncompressedDBSize = 250 << 20 // 250 MB safety cap against zip bombs
)

type backupMetadata struct {
	App       string `json:"app"`
	Version   string `json:"version"`
	Origin    string `json:"origin"`
	CreatedAt string `json:"created_at"`
	HasLogo   bool   `json:"has_logo"`
	LogoFile  string `json:"logo_file,omitempty"`
}

func createBackupZipArchive(destZipPath, srcDBPath, logoDataURI, origem string) error {
	zipFile, err := os.OpenFile(destZipPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)

	// 1. Add optimanager.db to the zip archive with Deflate compression
	dbIn, err := os.Open(srcDBPath)
	if err != nil {
		_ = zw.Close()
		return err
	}
	dbStat, err := dbIn.Stat()
	if err != nil {
		dbIn.Close()
		_ = zw.Close()
		return err
	}

	dbHeader, err := zip.FileInfoHeader(dbStat)
	if err != nil {
		dbIn.Close()
		_ = zw.Close()
		return err
	}
	dbHeader.Name = "optimanager.db"
	dbHeader.Method = zip.Deflate

	dbWriter, err := zw.CreateHeader(dbHeader)
	if err != nil {
		dbIn.Close()
		_ = zw.Close()
		return err
	}
	if _, err := io.Copy(dbWriter, dbIn); err != nil {
		dbIn.Close()
		_ = zw.Close()
		return err
	}
	dbIn.Close()

	// 2. If company logo is configured, include both the decoded image file (assets/logo.<ext>)
	// and the raw data URI (assets/logo_data_uri.txt) inside the zip archive.
	hasLogo := false
	logoFilename := ""
	trimmedLogo := strings.TrimSpace(logoDataURI)
	if trimmedLogo != "" {
		rawBytes, ext := decodeDataURIImage(trimmedLogo)
		if len(rawBytes) > 0 {
			hasLogo = true
			logoFilename = "assets/logo." + ext
			logoWriter, err := zw.CreateHeader(&zip.FileHeader{
				Name:     logoFilename,
				Method:   zip.Deflate,
				Modified: time.Now(),
			})
			if err == nil {
				_, _ = logoWriter.Write(rawBytes)
			}
		}

		uriWriter, err := zw.CreateHeader(&zip.FileHeader{
			Name:     "assets/logo_data_uri.txt",
			Method:   zip.Deflate,
			Modified: time.Now(),
		})
		if err == nil {
			_, _ = uriWriter.Write([]byte(trimmedLogo))
		}
	}

	// 3. Add backup_metadata.json manifest
	meta := backupMetadata{
		App:       "OptiManager",
		Version:   "3.0.0",
		Origin:    origem,
		CreatedAt: time.Now().Format(time.RFC3339),
		HasLogo:   hasLogo,
		LogoFile:  logoFilename,
	}
	if metaBytes, err := json.MarshalIndent(meta, "", "  "); err == nil {
		if mw, err := zw.CreateHeader(&zip.FileHeader{
			Name:     "backup_metadata.json",
			Method:   zip.Deflate,
			Modified: time.Now(),
		}); err == nil {
			_, _ = mw.Write(metaBytes)
		}
	}

	if err := zw.Close(); err != nil {
		return err
	}
	return zipFile.Sync()
}

func decodeDataURIImage(dataURI string) ([]byte, string) {
	ext := "png"
	b64Part := dataURI
	if strings.HasPrefix(dataURI, "data:") {
		commaIdx := strings.IndexByte(dataURI, ',')
		if commaIdx != -1 {
			meta := strings.ToLower(dataURI[:commaIdx])
			b64Part = dataURI[commaIdx+1:]
			switch {
			case strings.Contains(meta, "image/jpeg") || strings.Contains(meta, "image/jpg"):
				ext = "jpg"
			case strings.Contains(meta, "image/webp"):
				ext = "webp"
			case strings.Contains(meta, "image/svg"):
				ext = "svg"
			case strings.Contains(meta, "image/gif"):
				ext = "gif"
			default:
				ext = "png"
			}
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64Part))
	if err != nil {
		return nil, ext
	}
	return decoded, ext
}

func encodeImageBytesToDataURI(imgBytes []byte, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mime := "image/png"
	switch ext {
	case ".jpg", ".jpeg":
		mime = "image/jpeg"
	case ".webp":
		mime = "image/webp"
	case ".svg":
		mime = "image/svg+xml"
	case ".gif":
		mime = "image/gif"
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(imgBytes))
}

func isZipFile(filePath string) bool {
	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()
	magic := make([]byte, 4)
	n, err := io.ReadFull(f, magic)
	if err != nil || n < 4 {
		return false
	}
	return bytes.Equal(magic, []byte{'P', 'K', 0x03, 0x04})
}

// ExtractDatabaseFromBackup extracts the SQLite database file from a .zip backup archive
// (or copies it directly if backupPath is a raw .db file) into destDBPath.
func ExtractDatabaseFromBackup(backupPath, destDBPath string) error {
	_, err := extractBackupArchive(backupPath, destDBPath)
	return err
}

func extractBackupArchive(srcPath, destDBPath string) (string, error) {
	if !isZipFile(srcPath) {
		in, err := os.Open(srcPath)
		if err != nil {
			return "", err
		}
		defer in.Close()
		out, err := os.OpenFile(destDBPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return "", err
		}
		defer out.Close()
		if _, err := io.Copy(out, in); err != nil {
			return "", err
		}
		return "", out.Sync()
	}

	zr, err := zip.OpenReader(srcPath)
	if err != nil {
		return "", fmt.Errorf("falha ao abrir arquivo compactado (.zip): %w", err)
	}
	defer zr.Close()

	var dbEntry *zip.File
	var logoDataURI string
	var fallbackLogoBytes []byte
	var fallbackLogoName string

	for _, f := range zr.File {
		cleanName := filepath.ToSlash(filepath.Clean(f.Name))
		if strings.HasPrefix(cleanName, "../") || strings.HasPrefix(cleanName, "/") {
			continue
		}
		if cleanName == "optimanager.db" || (dbEntry == nil && strings.HasSuffix(strings.ToLower(cleanName), ".db")) {
			dbEntry = f
		} else if cleanName == "assets/logo_data_uri.txt" && f.UncompressedSize64 < 5<<20 {
			if rc, err := f.Open(); err == nil {
				if data, err := io.ReadAll(io.LimitReader(rc, 5<<20)); err == nil {
					logoDataURI = strings.TrimSpace(string(data))
				}
				rc.Close()
			}
		} else if strings.HasPrefix(cleanName, "assets/logo.") && f.UncompressedSize64 < 5<<20 {
			if rc, err := f.Open(); err == nil {
				if data, err := io.ReadAll(io.LimitReader(rc, 5<<20)); err == nil && len(data) > 0 {
					fallbackLogoBytes = data
					fallbackLogoName = cleanName
				}
				rc.Close()
			}
		}
	}

	if dbEntry == nil {
		return "", errors.New("o arquivo .zip enviado não contém um banco de dados (.db) do OptiManager")
	}

	rc, err := dbEntry.Open()
	if err != nil {
		return "", fmt.Errorf("falha ao extrair banco de dados do arquivo .zip: %w", err)
	}
	defer rc.Close()

	out, err := os.OpenFile(destDBPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(out, io.LimitReader(rc, maxUncompressedDBSize)); err != nil {
		out.Close()
		return "", fmt.Errorf("falha ao gravar banco de dados extraído do .zip: %w", err)
	}
	if err := out.Sync(); err != nil {
		out.Close()
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}

	if logoDataURI == "" && len(fallbackLogoBytes) > 0 {
		logoDataURI = encodeImageBytesToDataURI(fallbackLogoBytes, fallbackLogoName)
	}

	return logoDataURI, nil
}

// ValidateBackupFile checks that filePath is a valid, uncorrupted SQLite database file
// compatible with OptiManager's schema and optionally containing at least one administrator account.
func (s *Service) ValidateBackupFile(ctx context.Context, filePath string, requireAdmin bool) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("falha ao abrir arquivo de backup: %w", err)
	}
	header := make([]byte, 16)
	n, err := io.ReadFull(f, header)
	f.Close()
	if err != nil || n < 16 || string(header) != sqliteMagicHeader {
		return errors.New("o arquivo enviado não é um banco de dados SQLite válido")
	}

	checkDB, err := sql.Open("sqlite", fmt.Sprintf("%s?mode=ro&_pragma=query_only(1)&_pragma=busy_timeout(5000)", filePath))
	if err != nil {
		return fmt.Errorf("falha ao inspecionar banco de dados de backup: %w", err)
	}
	defer checkDB.Close()

	var integrity string
	if err := checkDB.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil {
		return fmt.Errorf("falha ao verificar integridade do arquivo SQLite: %w", err)
	}
	if integrity != "ok" {
		return fmt.Errorf("o arquivo de backup está corrompido (integrity_check: %s)", integrity)
	}

	var userVer int
	if err := checkDB.QueryRowContext(ctx, "PRAGMA user_version").Scan(&userVer); err != nil {
		return fmt.Errorf("falha ao verificar versão de esquema do backup: %w", err)
	}
	maxVer := database.LatestMigrationVersion()
	if userVer < 0 || userVer > maxVer {
		return fmt.Errorf("versão de banco de dados incompatível (versão do arquivo: v%d, versão máxima suportada por esta instância: v%d)", userVer, maxVer)
	}

	requiredTables := []string{"config", "usuarios", "clientes", "ordens_servico"}
	for _, tbl := range requiredTables {
		var count int
		err := checkDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, tbl).Scan(&count)
		if err != nil || count == 0 {
			return fmt.Errorf("o arquivo SQLite não pertence ao OptiManager (tabela obrigatória ausente: %s)", tbl)
		}
	}

	if requireAdmin {
		var adminCount int
		err := checkDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM usuarios WHERE role = 'admin'`).Scan(&adminCount)
		if err != nil || adminCount == 0 {
			return errors.New("o backup selecionado não possui nenhum usuário administrador cadastrado e não pode ser restaurado")
		}
	}

	return nil
}

// RestoreFromReader validates an uploaded SQLite stream, optionally creates a pre-restore snapshot,
// atomically replaces the active database, and logs the operation.
func (s *Service) RestoreFromReader(ctx context.Context, r io.Reader, sourceName string, createPreRestore bool, usuarioID int64, usuarioNome, ip string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	randBytes := make([]byte, 4)
	_, _ = rand.Read(randBytes)
	tmpPath := filepath.Join(s.localDir, fmt.Sprintf(".restore_%d_%x.tmp", time.Now().UnixNano(), randBytes))
	defer os.Remove(tmpPath)

	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo temporário para restauração: %w", err)
	}
	if _, err := io.Copy(out, r); err != nil {
		out.Close()
		return fmt.Errorf("falha ao gravar arquivo de backup enviado: %w", err)
	}
	if err := out.Sync(); err != nil {
		out.Close()
		return fmt.Errorf("falha ao sincronizar arquivo temporário em disco: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("falha ao fechar arquivo temporário: %w", err)
	}

	return s.restoreFromValidatedPathLocked(ctx, tmpPath, sourceName, createPreRestore, usuarioID, usuarioNome, ip)
}

// RestoreFromFilename restores the active database from an existing backup file in the local backup directory.
func (s *Service) RestoreFromFilename(ctx context.Context, filename string, createPreRestore bool, usuarioID int64, usuarioNome, ip string) error {
	targetPath, err := s.GetBackupFilePath(filename)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.restoreFromValidatedPathLocked(ctx, targetPath, filename, createPreRestore, usuarioID, usuarioNome, ip)
}

func (s *Service) restoreFromValidatedPathLocked(ctx context.Context, srcPath, sourceLabel string, createPreRestore bool, usuarioID int64, usuarioNome, ip string) error {
	randBytes := make([]byte, 4)
	_, _ = rand.Read(randBytes)
	extractedDBPath := filepath.Join(s.localDir, fmt.Sprintf(".restore_db_%d_%x.tmp", time.Now().UnixNano(), randBytes))
	defer os.Remove(extractedDBPath)

	restoredLogoURI, err := extractBackupArchive(srcPath, extractedDBPath)
	if err != nil {
		return err
	}

	if err := s.ValidateBackupFile(ctx, extractedDBPath, true); err != nil {
		return err
	}

	var preRestoreFilename string
	if createPreRestore {
		info, err := s.createBackupLocked(ctx, "prerestore")
		if err != nil {
			return fmt.Errorf("restauração abortada: falha ao criar snapshot de segurança pré-restauração: %w", err)
		}
		preRestoreFilename = info.Filename
	}

	if err := s.db.ReplaceDatabase(extractedDBPath); err != nil {
		return fmt.Errorf("falha crítica ao substituir banco de dados ativo: %w", err)
	}

	// If the .zip archive contained an extracted company logo asset, ensure config.logo_base64 is populated
	if restoredLogoURI != "" {
		if cfg, err := s.db.GetConfig(); err == nil && cfg != nil && strings.TrimSpace(cfg.LogoBase64) == "" {
			cfg.LogoBase64 = restoredLogoURI
			_ = s.db.UpdateConfig(cfg)
		}
	}

	// Record persistent file-based restore log + SQLite audit_log in the newly active database
	detail := fmt.Sprintf("Banco de dados restaurado a partir de '%s'", sourceLabel)
	if preRestoreFilename != "" {
		detail += fmt.Sprintf(" (snapshot pré-restauração: %s)", preRestoreFilename)
	}
	_ = s.db.InsertAuditLog(usuarioID, usuarioNome, ip, "BACKUP_RESTORE", detail)
	s.appendExternalRestoreLog(usuarioNome, ip, detail)

	return nil
}

func (s *Service) appendExternalRestoreLog(usuarioNome, ip, detail string) {
	logPath := filepath.Join(filepath.Dir(s.db.Path), "restore_audit.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	line := fmt.Sprintf("[%s] user=%s ip=%s %s\n", time.Now().In(s.tz).Format(time.RFC3339), usuarioNome, ip, detail)
	_, _ = f.WriteString(line)
}

// StartScheduler starts a lightweight background goroutine that checks once per minute
// whether an automated backup is due based on the configuration stored in SQLite.
func (s *Service) StartScheduler(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		var lastTriggeredDate string

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cfg, err := s.db.GetConfig()
				if err != nil || cfg == nil || cfg.BackupAutoEnabled != 1 {
					continue
				}

				now := time.Now().In(s.tz)
				currentHM := now.Format("15:04")
				todayKey := now.Format("2006-01-02") + "_" + currentHM

				if strings.TrimSpace(cfg.BackupHorario) != currentHM {
					continue
				}

				if cfg.BackupFrequencia == "semanal" && int(now.Weekday()) != cfg.BackupDiaSemana {
					continue
				}

				if lastTriggeredDate == todayKey {
					continue
				}

				// Also check if BackupUltimoSucesso was already recorded in the same minute
				if strings.HasPrefix(cfg.BackupUltimoSucesso, now.Format("2006-01-02T15:04")) {
					lastTriggeredDate = todayKey
					continue
				}

				lastTriggeredDate = todayKey
				backupCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
				info, err := s.CreateBackup(backupCtx, "auto")
				cancel()
				if err != nil {
					log.Printf("[BACKUP SCHEDULER] Automated backup failed: %v", err)
				} else {
					log.Printf("[BACKUP SCHEDULER] Automated backup created: %s (%s)", info.Filename, info.TamanhoFmt)
					_ = s.db.InsertAuditLog(0, "sistema", "127.0.0.1", "BACKUP_AUTO", fmt.Sprintf("Backup automático gerado: %s", info.Filename))
				}
			}
		}
	}()
}

