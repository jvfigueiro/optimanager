package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDataDirAndUploadsDir(t *testing.T) {
	// 1. When DB_PATH is set, ResolveDataDir returns its parent directory
	customDB := filepath.Join(t.TempDir(), "custom", "my.db")
	t.Setenv("DB_PATH", customDB)
	if got := ResolveDataDir(); got != filepath.Dir(customDB) {
		t.Fatalf("expected ResolveDataDir() = %q, got %q", filepath.Dir(customDB), got)
	}

	// 2. When DB_PATH is empty, ResolveDataDir returns "/data" if it exists,
	// or "<executable_dir>/data" otherwise.
	t.Setenv("DB_PATH", "")
	got := ResolveDataDir()
	if info, err := os.Stat("/data"); err == nil && info.IsDir() {
		if got != "/data" {
			t.Fatalf("expected /data when /data exists, got %q", got)
		}
	} else {
		exePath, err := os.Executable()
		if err != nil {
			t.Fatalf("os.Executable failed: %v", err)
		}
		if resolved, err := filepath.EvalSymlinks(exePath); err == nil {
			exePath = resolved
		}
		expected := filepath.Join(filepath.Dir(exePath), "data")
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	}

	// 3. UploadsDir on a DB instance resolves alongside the DB file
	dbDir := t.TempDir()
	db := &DB{Path: filepath.Join(dbDir, "optimanager.db")}
	expectedUploads := filepath.Join(dbDir, "uploads")
	if gotUploads := db.UploadsDir(); gotUploads != expectedUploads {
		t.Fatalf("expected UploadsDir() = %q, got %q", expectedUploads, gotUploads)
	}
}
