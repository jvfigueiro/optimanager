package dbf

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// AllowedDomainTables defines the strict allowlist of non-PII vocabulary/domain tables
// whose distinct values are safe to include in the profile-dbf output.
// Customer, user, order, financial, credential, and free-text tables are NEVER allowlisted.
var AllowedDomainTables = map[string]bool{
	"SITUA":    true,
	"FORMA":    true,
	"GRUPOS":   true,
	"MARCA":    true,
	"MARCAS":   true,
	"MODELO":   true,
	"MODELOCA": true,
	"EQUIPA":   true,
	"INFO":     true,
	"TIPO":     true,
	"PROCEDE":  true,
	"NIVEL":    true,
	"FONTES":   true,
}

// FieldProfile summarizes structure and non-empty fill rate for a single DBF column.
// DistinctValues is populated ONLY when the parent table is in AllowedDomainTables.
type FieldProfile struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Length         uint8    `json:"length"`
	Decimals       uint8    `json:"decimals"`
	FilledCount    int      `json:"filled_count"`
	FillRatePct    float64  `json:"fill_rate_pct"`
	DistinctValues []string `json:"distinct_values,omitempty"`
}

// TableProfile summarizes a single DBF file without exposing personal data.
type TableProfile struct {
	TableName        string         `json:"table_name"`
	FileName         string         `json:"file_name"`
	VersionHex       string         `json:"version_hex"`
	CodePageHex      string         `json:"code_page_hex"`
	ResolvedEncoding string         `json:"resolved_encoding"`
	MemoFile         string         `json:"memo_file,omitempty"`
	HeaderRecords    uint32         `json:"header_records"`
	ActiveRecords    int            `json:"active_records"`
	DeletedRecords   int            `json:"deleted_records"`
	TruncatedRecords int            `json:"truncated_records"`
	CountMismatch    bool           `json:"count_mismatch"`
	IsDomainTable    bool           `json:"is_domain_table"`
	Fields           []FieldProfile `json:"fields"`
}

// DirectoryProfile holds the complete profile report for a legacy database folder.
type DirectoryProfile struct {
	TotalTables int            `json:"total_tables"`
	Tables      []TableProfile `json:"tables"`
}

// ProfileDirectory scans all .dbf files (case-insensitive) in dirPath and generates
// a privacy-safe structural and statistical report.
func ProfileDirectory(dirPath string, opts Options) (*DirectoryProfile, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var dbfFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".dbf") {
			dbfFiles = append(dbfFiles, filepath.Join(dirPath, e.Name()))
		}
	}
	sort.Slice(dbfFiles, func(i, j int) bool {
		return strings.ToUpper(filepath.Base(dbfFiles[i])) < strings.ToUpper(filepath.Base(dbfFiles[j]))
	})

	report := &DirectoryProfile{
		TotalTables: len(dbfFiles),
		Tables:      make([]TableProfile, 0, len(dbfFiles)),
	}

	for _, fpath := range dbfFiles {
		tp, err := ProfileTable(fpath, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to profile %s: %w", filepath.Base(fpath), err)
		}
		report.Tables = append(report.Tables, *tp)
	}

	return report, nil
}

// ProfileTable profiles a single DBF file.
func ProfileTable(path string, opts Options) (*TableProfile, error) {
	tbl, err := Open(path, opts)
	if err != nil {
		return nil, err
	}

	base := filepath.Base(path)
	stem := strings.ToUpper(strings.TrimSuffix(base, filepath.Ext(base)))
	isDomain := AllowedDomainTables[stem]

	fields := make([]FieldProfile, 0, len(tbl.Header.Fields))
	for _, f := range tbl.Header.Fields {
		filled := 0
		var distinctSet map[string]struct{}
		var distinctList []string
		if isDomain && f.Type == 'C' {
			distinctSet = make(map[string]struct{})
		}

		for _, r := range tbl.Records {
			val := r.Values[f.Name]
			if isMeaningfullyFilled(val, f.Type) {
				filled++
				if distinctSet != nil && len(distinctList) < 100 {
					if _, exists := distinctSet[val]; !exists {
						distinctSet[val] = struct{}{}
						distinctList = append(distinctList, val)
					}
				}
			}
		}

		var pct float64
		if tbl.Stats.ActiveRecords > 0 {
			pct = math.Round((float64(filled)/float64(tbl.Stats.ActiveRecords))*1000.0) / 10.0
		}

		fields = append(fields, FieldProfile{
			Name:           f.Name,
			Type:           string(f.Type),
			Length:         f.Length,
			Decimals:       f.Decimals,
			FilledCount:    filled,
			FillRatePct:    pct,
			DistinctValues: distinctList,
		})
	}

	return &TableProfile{
		TableName:        stem,
		FileName:         base,
		VersionHex:       fmt.Sprintf("0x%02x", tbl.Header.VersionByte),
		CodePageHex:      fmt.Sprintf("0x%02x", tbl.Header.CodePageByte),
		ResolvedEncoding: tbl.Stats.ResolvedEncoding,
		MemoFile:         tbl.Stats.MemoFile,
		HeaderRecords:    tbl.Stats.HeaderRecords,
		ActiveRecords:    tbl.Stats.ActiveRecords,
		DeletedRecords:   tbl.Stats.DeletedRecords,
		TruncatedRecords: tbl.Stats.TruncatedRecords,
		CountMismatch:    tbl.Stats.CountMismatch,
		IsDomainTable:    isDomain,
		Fields:           fields,
	}, nil
}

func isMeaningfullyFilled(val string, fieldType byte) bool {
	val = strings.TrimSpace(val)
	if val == "" {
		return false
	}
	switch fieldType {
	case 'L':
		return strings.EqualFold(val, "T") || strings.EqualFold(val, "Y") || val == "1"
	case 'N', 'F':
		cleaned := strings.Trim(val, "0.,-+ ")
		return cleaned != ""
	case 'D':
		cleaned := strings.Trim(val, "0/-: ")
		return cleaned != ""
	default:
		// Ignore empty formatting masks like "(  )     -    " or "  /  /    "
		cleaned := strings.Trim(val, " ./-:_()")
		return cleaned != ""
	}
}

// WriteReport renders the DirectoryProfile in JSON or human-readable text format.
func WriteReport(w io.Writer, prof *DirectoryProfile, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(prof)
	}

	fmt.Fprintf(w, "OptiManager Legacy DBF Profile Report (Zero-PII Mode)\n")
	fmt.Fprintf(w, "Total DBF tables scanned: %d\n", prof.TotalTables)
	fmt.Fprintf(w, "========================================================================\n")

	for _, t := range prof.Tables {
		mismatchFlag := ""
		if t.CountMismatch {
			mismatchFlag = " [HEADER_COUNT_MISMATCH]"
		}
		memoInfo := ""
		if t.MemoFile != "" {
			memoInfo = fmt.Sprintf(" memo=%s", t.MemoFile)
		}
		fmt.Fprintf(w, "\nTable: %s (%s) ver=%s cp=%s enc=%q%s\n",
			t.TableName, t.FileName, t.VersionHex, t.CodePageHex, t.ResolvedEncoding, memoInfo)
		fmt.Fprintf(w, "  Records: active=%d, deleted=%d, header=%d, truncated=%d%s\n",
			t.ActiveRecords, t.DeletedRecords, t.HeaderRecords, t.TruncatedRecords, mismatchFlag)
		fmt.Fprintf(w, "  Fields (%d):\n", len(t.Fields))
		for _, f := range t.Fields {
			typeSpec := fmt.Sprintf("%s(%d)", f.Type, f.Length)
			if f.Decimals > 0 {
				typeSpec = fmt.Sprintf("%s(%d.%d)", f.Type, f.Length, f.Decimals)
			}
			fmt.Fprintf(w, "    - %-12s %-10s filled=%d/%d (%.1f%%)\n",
				f.Name, typeSpec, f.FilledCount, t.ActiveRecords, f.FillRatePct)
			if t.IsDomainTable && len(f.DistinctValues) > 0 {
				fmt.Fprintf(w, "      distinct: %s\n", strings.Join(f.DistinctValues, " | "))
			}
		}
	}
	return nil
}
