package redact

import (
	"regexp"
)

var (
	// Matches formatted CPF (000.000.000-00) or CNPJ (00.000.000/0000-00 or alphanumeric AA.AAA.AAA/AAAA-00)
	docFormattedRegex = regexp.MustCompile(`(?i)\b[0-9A-Z]{2,3}\.[0-9A-Z]{3}\.[0-9A-Z]{3}[/-][0-9A-Z]{2,4}(?:-[0-9]{2})?\b`)
	// Matches 11 or 14 contiguous digits that look like unformatted CPF/CNPJ
	docDigitsRegex = regexp.MustCompile(`\b(?:\d{11}|\d{14})\b`)
	// Matches formatted BR phone numbers like (11) 98765-4321 or (51) 3541-6420
	phoneFormattedRegex = regexp.MustCompile(`\(\d{2}\)\s*\d{4,5}-\d{4}`)
)

// SanitizeLogMessage removes formatted/unformatted CPF, CNPJ, and phone numbers
// from error strings and log messages to comply with LGPD data minimization.
func SanitizeLogMessage(msg string) string {
	if msg == "" {
		return ""
	}
	s := docFormattedRegex.ReplaceAllString(msg, "[REDACTED_DOC]")
	s = phoneFormattedRegex.ReplaceAllString(s, "[REDACTED_PHONE]")
	s = docDigitsRegex.ReplaceAllString(s, "[REDACTED_DOC]")
	return s
}
