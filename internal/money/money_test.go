package money_test

import (
	"testing"

	"optimanager/internal/money"
)

func TestReaisToCentavos_HalfUpAndFloatPrecision(t *testing.T) {
	cases := []struct {
		in   float64
		want int64
	}{
		{0, 0},
		{19.99, 1999},
		{19.90, 1990},
		{80.00, 8000},
		{1.005, 101},
		{1.004, 100},
		{1234.56, 123456},
		{-19.99, -1999},
		{-1.005, -101},
	}

	for _, tc := range cases {
		got := money.ReaisToCentavos(tc.in)
		if got != tc.want {
			t.Errorf("ReaisToCentavos(%v) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestParseBRLToCentavos(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"0", 0},
		{"80", 8000},
		{"19.99", 1999},
		{"19,99", 1999},
		{"R$ 1.234,56", 123456},
		{"1,234.56", 123456},
		{"10.0000", 1000},
		{"10.1250", 1013}, // half-up on 3rd decimal
		{"10.1249", 1012},
		{"-45,50", -4550},
	}

	for _, tc := range cases {
		got, err := money.ParseBRLToCentavos(tc.in)
		if err != nil {
			t.Fatalf("ParseBRLToCentavos(%q) unexpected error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("ParseBRLToCentavos(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}

	invalid := []string{"", "abc", "R$", "12,34,56"}
	for _, s := range invalid {
		if _, err := money.ParseBRLToCentavos(s); err == nil {
			t.Errorf("ParseBRLToCentavos(%q) expected error, got nil", s)
		}
	}
}

func TestFormatBRL(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "R$ 0,00"},
		{5, "R$ 0,05"},
		{1999, "R$ 19,99"},
		{123456, "R$ 1.234,56"},
		{100000000, "R$ 1.000.000,00"},
		{-2550, "-R$ 25,50"},
	}
	for _, tc := range cases {
		if got := money.FormatBRL(tc.in); got != tc.want {
			t.Errorf("FormatBRL(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestMultiplyQtyPrice(t *testing.T) {
	// 1.000 unit * R$ 22,00 = R$ 22,00 (2200 centavos)
	if got := money.MultiplyQtyPrice(1000, 2200); got != 2200 {
		t.Errorf("got %d, want 2200", got)
	}
	// 1.500 units * R$ 10,01 = 1501.5 centavos -> 1502 centavos (half-up)
	if got := money.MultiplyQtyPrice(1500, 1001); got != 1502 {
		t.Errorf("got %d, want 1502", got)
	}
	// 250 milesimos (0.25) * R$ 10,00 = 250 centavos
	if got := money.MultiplyQtyPrice(250, 1000); got != 250 {
		t.Errorf("got %d, want 250", got)
	}
}

func TestSplitInstallments_PropertyRemainderInFirst(t *testing.T) {
	if _, err := money.SplitInstallments(1000, 0); err == nil {
		t.Fatalf("expected error for 0 installments")
	}

	// Check explicitly 100 centavos in 3 installments -> [34, 33, 33]
	p3, err := money.SplitInstallments(100, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p3) != 3 || p3[0] != 34 || p3[1] != 33 || p3[2] != 33 {
		t.Fatalf("expected [34, 33, 33], got %v", p3)
	}

	// Property check across totals and installment counts 1..36
	totals := []int64{0, 1, 2, 7, 99, 100, 101, 1999, 10000, 1234567}
	for _, total := range totals {
		for n := 1; n <= 36; n++ {
			parts, err := money.SplitInstallments(total, n)
			if err != nil {
				t.Fatalf("SplitInstallments(%d, %d) error: %v", total, n, err)
			}
			if len(parts) != n {
				t.Fatalf("expected %d parts, got %d", n, len(parts))
			}
			var sum int64
			for i, v := range parts {
				sum += v
				if i > 1 && v != parts[1] {
					t.Fatalf("expected all tail installments to be equal: %v", parts)
				}
			}
			if sum != total {
				t.Fatalf("sum(%v) = %d, want %d (n=%d)", parts, sum, total, n)
			}
			if n > 1 {
				rem := total - (total/int64(n))*int64(n)
				if parts[0]-parts[1] != rem {
					t.Fatalf("expected first installment to carry remainder %d: %v", rem, parts)
				}
			}
		}
	}
}

func TestParseQtyToMilesimos(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"1", 1000},
		{"1.5", 1500},
		{"1,25", 1250},
		{"100.000", 100000},
		{"0.001", 1},
		{"0.0015", 2},
	}
	for _, tc := range cases {
		got, err := money.ParseQtyToMilesimos(tc.in)
		if err != nil {
			t.Fatalf("ParseQtyToMilesimos(%q) error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("ParseQtyToMilesimos(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
