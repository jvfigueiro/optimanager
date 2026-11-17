package dbf_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/text/encoding/charmap"

	"optimanager/internal/legacy/dbf"
)

func TestOpen_SyntheticDBF_HeaderMismatchDeletedAndEncodings(t *testing.T) {
	tmpDir := t.TempDir()
	dbfPath := filepath.Join(tmpDir, "SITUA.DBF")

	fields := []dbf.FieldDesc{
		{Name: "CODIGO", Type: 'N', Length: 6},
		{Name: "SITUACAO", Type: 'C', Length: 30},
	}

	rows := []struct {
		Deleted bool
		Cells   map[string]string
	}{
		{Deleted: false, Cells: map[string]string{"CODIGO": "1", "SITUACAO": "Aguardando Aprovação"}},
		{Deleted: true, Cells: map[string]string{"CODIGO": "2", "SITUACAO": "Excluído Antigo"}},
		{Deleted: false, Cells: map[string]string{"CODIGO": "3", "SITUACAO": "Concluído / Manutenção"}},
	}

	var buf bytes.Buffer
	// Write header claiming 99 records while actual physical records = 3 (2 active + 1 deleted)
	err := dbf.WriteSyntheticDBF(&buf, 0x03, 0x00, 99, fields, rows, charmap.Windows1252)
	if err != nil {
		t.Fatalf("WriteSyntheticDBF failed: %v", err)
	}
	if err := os.WriteFile(dbfPath, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	tbl, err := dbf.Open(dbfPath, dbf.Options{})
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if tbl.Stats.ActiveRecords != 2 {
		t.Fatalf("expected 2 active records, got %d", tbl.Stats.ActiveRecords)
	}
	if tbl.Stats.DeletedRecords != 1 {
		t.Fatalf("expected 1 deleted record, got %d", tbl.Stats.DeletedRecords)
	}
	if !tbl.Stats.CountMismatch {
		t.Fatalf("expected CountMismatch=true when header has 99 and file has 3")
	}
	if tbl.Records[0].Values["SITUACAO"] != "Aguardando Aprovação" {
		t.Errorf("unexpected decoded value: %q", tbl.Records[0].Values["SITUACAO"])
	}
	if tbl.Records[1].Values["SITUACAO"] != "Concluído / Manutenção" {
		t.Errorf("unexpected decoded value: %q", tbl.Records[1].Values["SITUACAO"])
	}
}

func TestOpen_SyntheticDBT_And_FPT_Memos(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Test dBase III .DBT memo (512-byte blocks)
	dbtTablePath := filepath.Join(tmpDir, "OBSORD.DBF")
	dbtMemoPath := filepath.Join(tmpDir, "OBSORD.DBT")

	dbtFields := []dbf.FieldDesc{
		{Name: "CODIGO", Type: 'N', Length: 6},
		{Name: "OBSGERAL", Type: 'M', Length: 10},
	}
	dbtRows := []struct {
		Deleted bool
		Cells   map[string]string
	}{
		{Deleted: false, Cells: map[string]string{"CODIGO": "1", "OBSGERAL": "         1"}},
	}

	var dbtBuf bytes.Buffer
	if err := dbf.WriteSyntheticDBF(&dbtBuf, 0x83, 0x03, 1, dbtFields, dbtRows, charmap.Windows1252); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dbtTablePath, dbtBuf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	// Build 2 blocks of 512 bytes: block 0 = header, block 1 = memo text + 0x1A
	dbtData := make([]byte, 1024)
	memoText, _ := charmap.Windows1252.NewEncoder().Bytes([]byte("Observação técnica do bloco DBT"))
	copy(dbtData[512:], memoText)
	dbtData[512+len(memoText)] = 0x1A
	if err := os.WriteFile(dbtMemoPath, dbtData, 0644); err != nil {
		t.Fatal(err)
	}

	tblDBT, err := dbf.Open(dbtTablePath, dbf.Options{})
	if err != nil {
		t.Fatalf("Open DBT table failed: %v", err)
	}
	if got := tblDBT.Records[0].Values["OBSGERAL"]; got != "Observação técnica do bloco DBT" {
		t.Errorf("DBT memo mismatch: got %q", got)
	}

	// 2. Test FoxPro .FPT memo (custom block size 64 in header bytes 6..7)
	fptTablePath := filepath.Join(tmpDir, "CLIENTE.DBF")
	fptMemoPath := filepath.Join(tmpDir, "CLIENTE.fpt")

	fptFields := []dbf.FieldDesc{
		{Name: "COD", Type: 'N', Length: 6},
		{Name: "NOTAS", Type: 'M', Length: 10},
	}
	fptRows := []struct {
		Deleted bool
		Cells   map[string]string
	}{
		{Deleted: false, Cells: map[string]string{"COD": "10", "NOTAS": "         8"}},
	}
	var fptBuf bytes.Buffer
	if err := dbf.WriteSyntheticDBF(&fptBuf, 0xF5, 0x02, 1, fptFields, fptRows, charmap.CodePage850); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fptTablePath, fptBuf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	// Build FPT: 512-byte header (block size = 64 => block 8 is at offset 512)
	fptData := make([]byte, 640)
	binary.BigEndian.PutUint16(fptData[6:8], 64)
	fptContent, _ := charmap.CodePage850.NewEncoder().Bytes([]byte("Histórico FoxPro FPT com acentuação"))
	binary.BigEndian.PutUint32(fptData[512:516], 1) // type 1 = text
	binary.BigEndian.PutUint32(fptData[516:520], uint32(len(fptContent)))
	copy(fptData[520:], fptContent)
	if err := os.WriteFile(fptMemoPath, fptData, 0644); err != nil {
		t.Fatal(err)
	}

	tblFPT, err := dbf.Open(fptTablePath, dbf.Options{})
	if err != nil {
		t.Fatalf("Open FPT table failed: %v", err)
	}
	if got := tblFPT.Records[0].Values["NOTAS"]; got != "Histórico FoxPro FPT com acentuação" {
		t.Errorf("FPT memo mismatch: got %q", got)
	}
}

func TestProfileDirectory_ZeroPIICanaryGuarantee(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Create a sensitive non-domain table (CLIENTES.DBF) containing canary PII
	cliFields := []dbf.FieldDesc{
		{Name: "COD", Type: 'N', Length: 5},
		{Name: "NOM", Type: 'C', Length: 50},
		{Name: "CPF", Type: 'C', Length: 14},
		{Name: "TEL1", Type: 'C', Length: 14},
	}
	cliRows := []struct {
		Deleted bool
		Cells   map[string]string
	}{
		{
			Deleted: false,
			Cells: map[string]string{
				"COD":  "1",
				"NOM":  "CANARY_SECRET_CUSTOMER_NAME_SILVA",
				"CPF":  "987.654.321-09",
				"TEL1": "(51) 99887-6655",
			},
		},
	}
	var cliBuf bytes.Buffer
	_ = dbf.WriteSyntheticDBF(&cliBuf, 0x03, 0x00, 1, cliFields, cliRows, charmap.Windows1252)
	if err := os.WriteFile(filepath.Join(tmpDir, "CLIENTES.DBF"), cliBuf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Create an allowlisted domain table (FORMA.DBF)
	formaFields := []dbf.FieldDesc{
		{Name: "CODIGO", Type: 'N', Length: 6},
		{Name: "TIPO", Type: 'C', Length: 20},
	}
	formaRows := []struct {
		Deleted bool
		Cells   map[string]string
	}{
		{Deleted: false, Cells: map[string]string{"CODIGO": "1", "TIPO": "Dinheiro"}},
		{Deleted: false, Cells: map[string]string{"CODIGO": "2", "TIPO": "PIX"}},
	}
	var formaBuf bytes.Buffer
	_ = dbf.WriteSyntheticDBF(&formaBuf, 0x03, 0x00, 2, formaFields, formaRows, charmap.Windows1252)
	if err := os.WriteFile(filepath.Join(tmpDir, "FORMA.DBF"), formaBuf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	prof, err := dbf.ProfileDirectory(tmpDir, dbf.Options{})
	if err != nil {
		t.Fatalf("ProfileDirectory failed: %v", err)
	}

	for _, formatJSON := range []bool{false, true} {
		var out bytes.Buffer
		if err := dbf.WriteReport(&out, prof, formatJSON); err != nil {
			t.Fatalf("WriteReport(json=%v) failed: %v", formatJSON, err)
		}
		reportStr := out.String()

		// Canaries from CLIENTES.DBF MUST NEVER appear in the output
		canaries := []string{
			"CANARY_SECRET_CUSTOMER_NAME_SILVA",
			"987.654.321-09",
			"(51) 99887-6655",
			"99887-6655",
		}
		for _, c := range canaries {
			if strings.Contains(reportStr, c) {
				t.Fatalf("PII LEAK DETECTED in profile-dbf output (json=%v): found canary %q in:\n%s", formatJSON, c, reportStr)
			}
		}

		// Domain values from FORMA.DBF MUST appear
		if !strings.Contains(reportStr, "Dinheiro") || !strings.Contains(reportStr, "PIX") {
			t.Errorf("expected allowlisted FORMA domain values in output (json=%v), got:\n%s", formatJSON, reportStr)
		}
	}
}
