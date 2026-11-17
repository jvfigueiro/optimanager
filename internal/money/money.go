package money

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalidAmount indicates a malformed monetary or quantity string.
var ErrInvalidAmount = errors.New("invalid monetary amount")

// ErrInvalidInstallments indicates an invalid installment count (must be >= 1).
var ErrInvalidInstallments = errors.New("installment count must be at least 1")

// ReaisToCentavos converts a float64 amount in Reais to int64 centavos using
// decimal string representation and half-up rounding to avoid IEEE-754 drift
// (e.g., 19.99 -> 1999, 1.005 -> 101).
func ReaisToCentavos(reais float64) int64 {
	s := strconv.FormatFloat(reais, 'f', 6, 64)
	c, err := ParseBRLToCentavos(s)
	if err != nil {
		if reais >= 0 {
			return int64(reais*100.0 + 0.5)
		}
		return int64(reais*100.0 - 0.5)
	}
	return c
}

// CentavosToReais converts int64 centavos back to a float64 in Reais.
func CentavosToReais(centavos int64) float64 {
	return float64(centavos) / 100.0
}

// ParseBRLToCentavos parses a monetary string (e.g., "R$ 1.234,56", "19.99", "10.0000", "80")
// into int64 centavos using exact integer arithmetic and half-up rounding when more than
// 2 decimal places are present (such as legacy N14.4 fields).
func ParseBRLToCentavos(input string) (int64, error) {
	s := strings.TrimSpace(input)
	s = strings.TrimPrefix(s, "R$")
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrInvalidAmount
	}

	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = strings.TrimSpace(s[1:])
	} else if strings.HasPrefix(s, "+") {
		s = strings.TrimSpace(s[1:])
	}
	s = strings.TrimPrefix(s, "R$")
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrInvalidAmount
	}

	// Determine decimal separator:
	// If both '.' and ',' exist, the last one is the decimal separator and the other is thousands separator.
	// If only ',' exists, ',' is the decimal separator.
	// If only '.' exists, '.' is the decimal separator (unless multiple '.' exist).
	lastComma := strings.LastIndexByte(s, ',')
	lastDot := strings.LastIndexByte(s, '.')

	var intPartStr, fracPartStr string
	switch {
	case lastComma != -1 && lastDot != -1:
		if lastComma > lastDot {
			// Brazilian format: 1.234,56
			intPartStr = strings.ReplaceAll(s[:lastComma], ".", "")
			fracPartStr = s[lastComma+1:]
		} else {
			// US format: 1,234.56
			intPartStr = strings.ReplaceAll(s[:lastDot], ",", "")
			fracPartStr = s[lastDot+1:]
		}
	case lastComma != -1:
		if strings.Count(s, ",") > 1 {
			return 0, ErrInvalidAmount
		}
		intPartStr = s[:lastComma]
		fracPartStr = s[lastComma+1:]
	case lastDot != -1:
		if strings.Count(s, ".") > 1 {
			intPartStr = strings.ReplaceAll(s, ".", "")
			fracPartStr = ""
		} else {
			intPartStr = s[:lastDot]
			fracPartStr = s[lastDot+1:]
		}
	default:
		intPartStr = s
		fracPartStr = ""
	}

	if intPartStr == "" {
		intPartStr = "0"
	}

	for i := 0; i < len(intPartStr); i++ {
		if intPartStr[i] < '0' || intPartStr[i] > '9' {
			return 0, ErrInvalidAmount
		}
	}
	for i := 0; i < len(fracPartStr); i++ {
		if fracPartStr[i] < '0' || fracPartStr[i] > '9' {
			return 0, ErrInvalidAmount
		}
	}

	intVal, err := strconv.ParseInt(intPartStr, 10, 64)
	if err != nil {
		return 0, ErrInvalidAmount
	}

	// Pad or round fractional digits to 2 decimal places (half-up)
	var fracCentavos int64
	var roundUp bool
	switch len(fracPartStr) {
	case 0:
		fracCentavos = 0
	case 1:
		fracCentavos = int64(fracPartStr[0]-'0') * 10
	case 2:
		fracCentavos = int64(fracPartStr[0]-'0')*10 + int64(fracPartStr[1]-'0')
	default:
		fracCentavos = int64(fracPartStr[0]-'0')*10 + int64(fracPartStr[1]-'0')
		if fracPartStr[2] >= '5' {
			roundUp = true
		}
	}

	total := intVal*100 + fracCentavos
	if roundUp {
		total++
	}
	if neg {
		total = -total
	}
	return total, nil
}

// FormatBRL formats int64 centavos into Brazilian currency format (e.g. "R$ 1.234,56").
func FormatBRL(centavos int64) string {
	neg := centavos < 0
	if neg {
		centavos = -centavos
	}
	reais := centavos / 100
	cents := centavos % 100

	intStr := strconv.FormatInt(reais, 10)
	var b strings.Builder
	n := len(intStr)
	for i, ch := range intStr {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(ch)
	}

	if neg {
		return fmt.Sprintf("-R$ %s,%02d", b.String(), cents)
	}
	return fmt.Sprintf("R$ %s,%02d", b.String(), cents)
}

// MultiplyQtyPrice computes (qtyMilesimos * unitCentavos) / 1000 with half-up rounding.
// 1000 milesimos = 1 unit.
func MultiplyQtyPrice(qtyMilesimos int64, unitCentavos int64) int64 {
	prod := qtyMilesimos * unitCentavos
	if prod >= 0 {
		return (prod + 500) / 1000
	}
	return (prod - 500) / 1000
}

// SplitInstallments divides totalCentavos into n installments.
// Any division remainder is added to the first installment (index 0),
// guaranteeing that the sum of installments equals totalCentavos.
func SplitInstallments(totalCentavos int64, n int) ([]int64, error) {
	if n < 1 {
		return nil, ErrInvalidInstallments
	}
	count := int64(n)
	base := totalCentavos / count
	rem := totalCentavos - base*count

	installments := make([]int64, n)
	for i := 0; i < n; i++ {
		installments[i] = base
	}
	installments[0] += rem
	return installments, nil
}

// ParseQtyToMilesimos parses a quantity string (up to 3+ decimals) into milesimos (1000 = 1.000).
func ParseQtyToMilesimos(input string) (int64, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return 0, ErrInvalidAmount
	}

	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = strings.TrimSpace(s[1:])
	} else if strings.HasPrefix(s, "+") {
		s = strings.TrimSpace(s[1:])
	}

	s = strings.ReplaceAll(s, ",", ".")
	if strings.Count(s, ".") > 1 {
		return 0, ErrInvalidAmount
	}

	parts := strings.SplitN(s, ".", 2)
	intStr := parts[0]
	if intStr == "" {
		intStr = "0"
	}
	fracStr := ""
	if len(parts) == 2 {
		fracStr = parts[1]
	}

	for i := 0; i < len(intStr); i++ {
		if intStr[i] < '0' || intStr[i] > '9' {
			return 0, ErrInvalidAmount
		}
	}
	for i := 0; i < len(fracStr); i++ {
		if fracStr[i] < '0' || fracStr[i] > '9' {
			return 0, ErrInvalidAmount
		}
	}

	intVal, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return 0, ErrInvalidAmount
	}

	var fracMil int64
	var roundUp bool
	switch len(fracStr) {
	case 0:
		fracMil = 0
	case 1:
		fracMil = int64(fracStr[0]-'0') * 100
	case 2:
		fracMil = int64(fracStr[0]-'0')*100 + int64(fracStr[1]-'0')*10
	case 3:
		fracMil = int64(fracStr[0]-'0')*100 + int64(fracStr[1]-'0')*10 + int64(fracStr[2]-'0')
	default:
		fracMil = int64(fracStr[0]-'0')*100 + int64(fracStr[1]-'0')*10 + int64(fracStr[2]-'0')
		if fracStr[3] >= '5' {
			roundUp = true
		}
	}

	total := intVal*1000 + fracMil
	if roundUp {
		total++
	}
	if neg {
		total = -total
	}
	return total, nil
}
