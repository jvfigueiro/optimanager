package redact_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"optimanager/internal/redact"
)

func TestSanitizeLogMessage_RemovesPII(t *testing.T) {
	raw := "failed to insert client CPF 123.456.789-01 or 12345678901 and CNPJ 12.345.678/0001-90 or AB.CDE.123/0001-45 phone (51) 99139-5089"
	got := redact.SanitizeLogMessage(raw)

	canaries := []string{
		"123.456.789-01",
		"12345678901",
		"12.345.678/0001-90",
		"AB.CDE.123/0001-45",
		"(51) 99139-5089",
	}
	for _, c := range canaries {
		if strings.Contains(got, c) {
			t.Errorf("expected %q to be redacted in %q", c, got)
		}
	}
}

func TestCheckForbiddenFilesScript_BlocksForbiddenExtensions(t *testing.T) {
	scriptPath, err := filepath.Abs("../../scripts/check-forbidden-files.sh")
	if err != nil {
		t.Fatalf("failed to resolve script path: %v", err)
	}

	forbiddenSamples := []string{
		"CLIENTES.DBF",
		"clientes.Dbf",
		"MEMO.DBT",
		"CLIENTE.fpt",
		"INDCLI.CDX",
		"index.ntx",
		"table.mdx",
		"table.idx",
		"backup.bak",
		"dump.db",
		"export.csv",
		"DADOS.zip",
		"data/secret.txt",
		"legacy-data/test.txt",
	}

	for _, sample := range forbiddenSamples {
		tmpDir := t.TempDir()
		runGit(t, tmpDir, "init", "-q")

		fullPath := filepath.Join(tmpDir, sample)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("dummy"), 0644); err != nil {
			t.Fatalf("write failed: %v", err)
		}
		runGit(t, tmpDir, "add", "-f", sample)

		cmd := exec.Command(scriptPath, "--staged")
		cmd.Dir = tmpDir
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Errorf("expected check-forbidden-files.sh to fail for %q, but it succeeded", sample)
		}
		if !strings.Contains(string(out), "SECURITY ERROR") {
			t.Errorf("expected SECURITY ERROR in output for %q, got: %s", sample, string(out))
		}
	}

	// Verify clean repository passes
	cleanDir := t.TempDir()
	runGit(t, cleanDir, "init", "-q")
	if err := os.WriteFile(filepath.Join(cleanDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, cleanDir, "add", "main.go")
	cmd := exec.Command(scriptPath, "--staged")
	cmd.Dir = cleanDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("expected clean repo to pass, got error: %v (%s)", err, string(out))
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v (%s)", args, err, string(out))
	}
}
