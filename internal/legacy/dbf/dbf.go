package dbf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

// FieldDesc describes a single column in a DBF table header.
type FieldDesc struct {
	Name     string `json:"name"`
	Type     byte   `json:"type"` // C, N, F, D, L, M
	Length   uint8  `json:"length"`
	Decimals uint8  `json:"decimals"`
}

// Header holds metadata from the 32-byte DBF file header and field descriptors.
type Header struct {
	VersionByte       byte        `json:"version_byte"`
	LastUpdateYear    int         `json:"last_update_year"`
	LastUpdateMonth   int         `json:"last_update_month"`
	LastUpdateDay     int         `json:"last_update_day"`
	HeaderRecordCount uint32      `json:"header_record_count"`
	HeaderLength      uint16      `json:"header_length"`
	RecordLength      uint16      `json:"record_length"`
	CodePageByte      byte        `json:"code_page_byte"`
	Fields            []FieldDesc `json:"fields"`
}

// Options configures how a DBF file is opened and decoded.
type Options struct {
	// Encoding overrides automatic code page detection when the DBF header has code page 0x00.
	// Supported values: "cp1252", "cp850", "cp437", "cp866", "utf-8".
	Encoding string
}

// TableStats reports physical record counts after scanning a DBF file.
type TableStats struct {
	HeaderRecords    uint32 `json:"header_records"`
	ActiveRecords    int    `json:"active_records"`
	DeletedRecords   int    `json:"deleted_records"`
	TruncatedRecords int    `json:"truncated_records"`
	CountMismatch    bool   `json:"count_mismatch"`
	ResolvedEncoding string `json:"resolved_encoding"`
	MemoFile         string `json:"memo_file,omitempty"`
}

// Record represents an active (non-deleted) row decoded from a DBF table.
type Record struct {
	PhysicalIndex int               // 0-based physical index in the file
	Values        map[string]string // Decoded and trimmed string values by field name
}

// Table represents an opened read-only DBF table and its optional companion memo file.
type Table struct {
	Header   Header
	Stats    TableStats
	Records  []Record
	filePath string
}

// Open reads a DBF file (and its companion .DBT or .FPT memo file if present) in read-only mode.
// It never trusts HeaderRecordCount to stop reading; instead it scans until EOF or 0x1A marker,
// skips deleted records (0x2A), and tolerates count mismatches or trailing truncated bytes.
func Open(path string, opts Options) (*Table, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open DBF file: %w", err)
	}
	if len(raw) < 32 {
		return nil, fmt.Errorf("file too short to be a valid DBF header (%d bytes)", len(raw))
	}

	hdr, err := parseHeader(raw)
	if err != nil {
		return nil, err
	}

	memoPath, memoKind := findCompanionMemo(path)
	var memoData []byte
	if memoPath != "" {
		memoData, _ = os.ReadFile(memoPath)
	}

	// Sample bytes from body to assist encoding detection when CodePageByte == 0x00
	var sample []byte
	if int(hdr.HeaderLength) < len(raw) {
		body := raw[hdr.HeaderLength:]
		if len(body) > 16384 {
			sample = body[:16384]
		} else {
			sample = body
		}
	}
	encName, cmap := resolveCharset(hdr.CodePageByte, opts.Encoding, sample)

	t := &Table{
		Header:   hdr,
		filePath: path,
		Stats: TableStats{
			HeaderRecords:    hdr.HeaderRecordCount,
			ResolvedEncoding: encName,
		},
	}
	if memoPath != "" {
		t.Stats.MemoFile = filepath.Base(memoPath)
	}

	if hdr.RecordLength == 0 || int(hdr.HeaderLength) >= len(raw) {
		t.Stats.CountMismatch = hdr.HeaderRecordCount != 0
		return t, nil
	}

	pos := int(hdr.HeaderLength)
	recLen := int(hdr.RecordLength)
	physIdx := 0

	for pos < len(raw) {
		// 0x1A is the standard xBase EOF marker byte
		if raw[pos] == 0x1A {
			break
		}
		if pos+recLen > len(raw) {
			t.Stats.TruncatedRecords++
			break
		}

		recBytes := raw[pos : pos+recLen]
		pos += recLen
		delFlag := recBytes[0]
		if delFlag == 0x2A { // '*' marks a deleted record
			t.Stats.DeletedRecords++
			physIdx++
			continue
		}

		t.Stats.ActiveRecords++
		values := make(map[string]string, len(hdr.Fields))
		offset := 1
		for _, f := range hdr.Fields {
			end := offset + int(f.Length)
			if end > len(recBytes) {
				break
			}
			fieldRaw := recBytes[offset:end]
			offset = end

			if f.Type == 'M' {
				blockNum := parseMemoBlockNum(fieldRaw)
				if blockNum > 0 && len(memoData) > 0 {
					memoBytes := readMemoBlock(memoData, memoKind, blockNum)
					values[f.Name] = strings.TrimSpace(decodeBytes(memoBytes, cmap))
				} else {
					values[f.Name] = ""
				}
			} else {
				values[f.Name] = strings.TrimSpace(decodeBytes(fieldRaw, cmap))
			}
		}

		t.Records = append(t.Records, Record{
			PhysicalIndex: physIdx,
			Values:        values,
		})
		physIdx++
	}

	totalScanned := t.Stats.ActiveRecords + t.Stats.DeletedRecords
	t.Stats.CountMismatch = totalScanned != int(hdr.HeaderRecordCount)

	return t, nil
}

func parseHeader(raw []byte) (Header, error) {
	var h Header
	h.VersionByte = raw[0]
	yy := int(raw[1])
	if yy < 70 {
		h.LastUpdateYear = 2000 + yy
	} else {
		h.LastUpdateYear = 1900 + yy
	}
	h.LastUpdateMonth = int(raw[2])
	h.LastUpdateDay = int(raw[3])
	h.HeaderRecordCount = binary.LittleEndian.Uint32(raw[4:8])
	h.HeaderLength = binary.LittleEndian.Uint16(raw[8:10])
	h.RecordLength = binary.LittleEndian.Uint16(raw[10:12])
	h.CodePageByte = raw[29]

	if h.HeaderLength < 32 || int(h.HeaderLength) > len(raw) {
		return h, fmt.Errorf("invalid DBF header length: %d (file size %d)", h.HeaderLength, len(raw))
	}

	pos := 32
	limit := int(h.HeaderLength)
	for pos+32 <= limit {
		if raw[pos] == 0x0D {
			break
		}
		desc := raw[pos : pos+32]
		pos += 32

		nameBytes := desc[:11]
		if idx := bytes.IndexByte(nameBytes, 0); idx >= 0 {
			nameBytes = nameBytes[:idx]
		}
		name := strings.TrimSpace(string(nameBytes))
		if name == "" {
			continue
		}

		h.Fields = append(h.Fields, FieldDesc{
			Name:     name,
			Type:     desc[11],
			Length:   desc[16],
			Decimals: desc[17],
		})
	}

	return h, nil
}

func findCompanionMemo(dbfPath string) (string, string) {
	dir := filepath.Dir(dbfPath)
	base := filepath.Base(dbfPath)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", ""
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		eExt := filepath.Ext(name)
		eStem := strings.TrimSuffix(name, eExt)
		if !strings.EqualFold(eStem, stem) {
			continue
		}
		switch strings.ToLower(eExt) {
		case ".dbt":
			return filepath.Join(dir, name), "dbt"
		case ".fpt":
			return filepath.Join(dir, name), "fpt"
		}
	}
	return "", ""
}

func parseMemoBlockNum(raw []byte) int {
	if len(raw) == 4 {
		// Binary 4-byte integer pointer (Visual FoxPro)
		return int(binary.LittleEndian.Uint32(raw))
	}
	s := strings.TrimSpace(string(bytes.TrimRight(raw, "\x00")))
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func readMemoBlock(memoData []byte, kind string, blockNum int) []byte {
	if blockNum <= 0 {
		return nil
	}
	switch kind {
	case "fpt":
		if len(memoData) < 512 {
			return nil
		}
		blockSize := int(binary.BigEndian.Uint16(memoData[6:8]))
		if blockSize <= 0 {
			blockSize = 64
		}
		offset := blockNum * blockSize
		if offset+8 > len(memoData) {
			return nil
		}
		length := int(binary.BigEndian.Uint32(memoData[offset+4 : offset+8]))
		if length <= 0 {
			return nil
		}
		start := offset + 8
		end := start + length
		if end > len(memoData) {
			end = len(memoData)
		}
		return bytes.TrimRight(memoData[start:end], "\x00\x1a")

	case "dbt":
		const blockSize = 512
		offset := blockNum * blockSize
		if offset >= len(memoData) {
			return nil
		}
		rest := memoData[offset:]
		// Check for dBase IV block header (0xFF 0xFF 0x08 0x00 + 4-byte length)
		if len(rest) >= 8 && rest[0] == 0xFF && rest[1] == 0xFF && rest[2] == 0x08 && rest[3] == 0x00 {
			totalLen := int(binary.LittleEndian.Uint32(rest[4:8]))
			if totalLen > 8 {
				end := totalLen
				if end > len(rest) {
					end = len(rest)
				}
				return bytes.TrimRight(rest[8:end], "\x00\x1a")
			}
		}
		// dBase III: terminated by 0x1A (or 0x1A 0x1A)
		if idx := bytes.IndexByte(rest, 0x1A); idx >= 0 {
			return bytes.TrimRight(rest[:idx], "\x00")
		}
		return bytes.TrimRight(rest, "\x00")
	}
	return nil
}

func resolveCharset(codePageByte byte, explicit string, sample []byte) (string, *charmap.Charmap) {
	// 1. Honor explicit code page byte in header if present
	switch codePageByte {
	case 0x01:
		return "cp437 (header 0x01)", charmap.CodePage437
	case 0x02:
		return "cp850 (header 0x02)", charmap.CodePage850
	case 0x03, 0x57, 0xC8:
		return fmt.Sprintf("cp1252 (header 0x%02x)", codePageByte), charmap.Windows1252
	case 0x65:
		return "cp866 (header 0x65)", charmap.CodePage866
	}

	// 2. Honor explicit caller override
	switch strings.ToLower(strings.TrimSpace(explicit)) {
	case "cp1252", "windows-1252", "latin1", "iso-8859-1":
		return "cp1252 (explicit)", charmap.Windows1252
	case "cp850", "ibm850":
		return "cp850 (explicit)", charmap.CodePage850
	case "cp437", "ibm437":
		return "cp437 (explicit)", charmap.CodePage437
	case "cp866":
		return "cp866 (explicit)", charmap.CodePage866
	case "utf-8", "utf8":
		return "utf-8 (explicit)", nil
	}

	// 3. Heuristic auto-detection between cp1252 and cp850 on high bytes
	var score1252, score850 int
	for _, b := range sample {
		if b < 0x80 {
			continue
		}
		// Common Portuguese bytes in Windows-1252:
		// Á(C1) Â(C2) Ã(C3) Ç(C7) É(C9) Ê(CA) Í(CD) Ó(D3) Ô(D4) Õ(D5) Ú(DA)
		// á(E1) â(E2) ã(E3) ç(E7) é(E9) ê(EA) í(ED) ó(F3) ô(F4) õ(F5) ú(FA)
		switch b {
		case 0xC1, 0xC2, 0xC3, 0xC7, 0xC9, 0xCA, 0xCD, 0xD3, 0xD4, 0xD5, 0xDA,
			0xE1, 0xE2, 0xE3, 0xE7, 0xE9, 0xEA, 0xED, 0xF3, 0xF4, 0xF5, 0xFA:
			score1252 += 2
		}
		// Common Portuguese bytes in CP850 that are control/rare in CP1252:
		// Ç(80) ü(81) é(82) â(83) ç(87) ê(88) í(A1) ó(A2) ú(A3) ã(C6) Ã(C7) õ(E4) Õ(E5)
		switch b {
		case 0x80, 0x82, 0x83, 0x87, 0x88, 0xA0, 0xA1, 0xA2, 0xA3, 0xC6:
			score850 += 3
		}
	}

	if score850 > score1252 {
		return "cp850 (auto)", charmap.CodePage850
	}
	return "cp1252 (auto)", charmap.Windows1252
}

func decodeBytes(src []byte, cmap *charmap.Charmap) string {
	src = bytes.TrimRight(src, "\x00")
	if len(src) == 0 {
		return ""
	}
	if cmap == nil {
		if utf8.Valid(src) {
			return string(src)
		}
		cmap = charmap.Windows1252
	}
	var b strings.Builder
	b.Grow(len(src))
	for _, c := range src {
		if c < 0x80 {
			b.WriteByte(c)
		} else {
			r := cmap.DecodeByte(c)
			if r != utf8.RuneError {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// WriteSyntheticDBF is a test helper that builds a valid binary DBF file in memory or on disk.
// Used exclusively for unit testing without committing real customer DBF files.
func WriteSyntheticDBF(w io.Writer, version byte, codePage byte, headerCountOverride int, fields []FieldDesc, rows []struct {
	Deleted bool
	Cells   map[string]string
}, cmap *charmap.Charmap) error {
	headerLen := uint16(32 + len(fields)*32 + 1)
	recLen := uint16(1)
	for _, f := range fields {
		recLen += uint16(f.Length)
	}

	headerCount := uint32(len(rows))
	if headerCountOverride >= 0 {
		headerCount = uint32(headerCountOverride)
	}

	hdr := make([]byte, 32)
	hdr[0] = version
	hdr[1] = 26 // 2026
	hdr[2] = 11
	hdr[3] = 17
	binary.LittleEndian.PutUint32(hdr[4:8], headerCount)
	binary.LittleEndian.PutUint16(hdr[8:10], headerLen)
	binary.LittleEndian.PutUint16(hdr[10:12], recLen)
	hdr[29] = codePage

	if _, err := w.Write(hdr); err != nil {
		return err
	}

	for _, f := range fields {
		desc := make([]byte, 32)
		copy(desc[:11], f.Name)
		desc[11] = f.Type
		desc[16] = f.Length
		desc[17] = f.Decimals
		if _, err := w.Write(desc); err != nil {
			return err
		}
	}

	// Header terminator 0x0D
	if _, err := w.Write([]byte{0x0D}); err != nil {
		return err
	}

	for _, r := range rows {
		buf := bytes.Repeat([]byte{' '}, int(recLen))
		if r.Deleted {
			buf[0] = 0x2A
		} else {
			buf[0] = 0x20
		}
		offset := 1
		for _, f := range fields {
			val := r.Cells[f.Name]
			encoded := encodeString(val, cmap)
			if len(encoded) > int(f.Length) {
				encoded = encoded[:f.Length]
			}
			copy(buf[offset:offset+int(f.Length)], encoded)
			offset += int(f.Length)
		}
		if _, err := w.Write(buf); err != nil {
			return err
		}
	}

	// EOF marker 0x1A
	_, err := w.Write([]byte{0x1A})
	return err
}

func encodeString(s string, cmap *charmap.Charmap) []byte {
	if cmap == nil {
		return []byte(s)
	}
	var out []byte
	for _, r := range s {
		if b, ok := cmap.EncodeRune(r); ok {
			out = append(out, b)
		} else if r < 128 {
			out = append(out, byte(r))
		} else {
			out = append(out, '?')
		}
	}
	return out
}
