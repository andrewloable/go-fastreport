package functions_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/functions"
)

// ── German (de) ───────────────────────────────────────────────────────────────

func TestNumToWordsDe_Zero(t *testing.T) {
	if got := functions.NumToWordsDe(0); got != "null" {
		t.Errorf("NumToWordsDe(0) = %q, want 'null'", got)
	}
}

func TestNumToWordsDe_Negative(t *testing.T) {
	got := functions.NumToWordsDe(-5)
	if got == "" {
		t.Error("NumToWordsDe(-5) returned empty")
	}
}

func TestNumToWordsDe_Values(t *testing.T) {
	// Include values that exercise:
	// - ones/teens (1, 11, 12)
	// - compound tens (21 → "einundzwanzig")
	// - hundreds, thousands (exactly 1000 → "eintausend", 2000+)
	// - millions (1 → "Million", 2+ → "Millionen")
	// - billions (1 → "Billion", 2+ → "Billionen")
	// - the female=false path (top-level call)
	values := []int64{
		1, 2, 10, 11, 12, 20, 21, 22, 31, 99,
		100, 101, 200, 999,
		1000, 1001, 2000, 5000,
		1_000_000, 2_000_000,
		1_000_000_000_000, 2_000_000_000_000,
	}
	for _, n := range values {
		got := functions.NumToWordsDe(n)
		if got == "" {
			t.Errorf("NumToWordsDe(%d) returned empty", n)
		}
	}
}

func TestNumToWordsDeFloat_Basic(t *testing.T) {
	cases := []float64{0, 1.0, 1.5, -2.25, 100.99}
	for _, v := range cases {
		got := functions.NumToWordsDeFloat(v)
		if got == "" {
			t.Errorf("NumToWordsDeFloat(%v) returned empty", v)
		}
	}
}

func TestNumToWordsDeFloat_Negative(t *testing.T) {
	got := functions.NumToWordsDeFloat(-3.50)
	if got == "" {
		t.Error("NumToWordsDeFloat(-3.50) returned empty")
	}
}

// ── English GB (en_gb) ────────────────────────────────────────────────────────

func TestNumToWordsEnGb_Zero(t *testing.T) {
	if got := functions.NumToWordsEnGb(0); got != "zero" {
		t.Errorf("NumToWordsEnGb(0) = %q, want 'zero'", got)
	}
}

func TestNumToWordsEnGb_Negative(t *testing.T) {
	got := functions.NumToWordsEnGb(-5)
	if got == "" {
		t.Error("NumToWordsEnGb(-5) returned empty")
	}
}

func TestNumToWordsEnGb_Values(t *testing.T) {
	// Include milliard (10^9) and billion (10^12) — British long scale
	values := []int64{
		1, 2, 10, 11, 21, 100, 999,
		1000, 1_000_000,
		1_000_000_000,   // one milliard
		2_000_000_000,   // two milliards (milliard + remainder branch)
		1_000_000_000_000, // one billion
		2_000_000_000_000, // two billions (billion + remainder branch)
	}
	for _, n := range values {
		got := functions.NumToWordsEnGb(n)
		if got == "" {
			t.Errorf("NumToWordsEnGb(%d) returned empty", n)
		}
	}
}

func TestNumToWordsEnGbFloat_Basic(t *testing.T) {
	cases := []float64{0, 1.0, 1.5, -2.25, 100.50}
	for _, v := range cases {
		got := functions.NumToWordsEnGbFloat(v)
		if got == "" {
			t.Errorf("NumToWordsEnGbFloat(%v) returned empty", v)
		}
	}
}

func TestNumToWordsEnGbFloat_Negative(t *testing.T) {
	got := functions.NumToWordsEnGbFloat(-5.75)
	if got == "" {
		t.Error("NumToWordsEnGbFloat(-5.75) returned empty")
	}
}

// ── Spanish (es) ──────────────────────────────────────────────────────────────

func TestNumToWordsEs_Zero(t *testing.T) {
	if got := functions.NumToWordsEs(0); got != "cero" {
		t.Errorf("NumToWordsEs(0) = %q, want 'cero'", got)
	}
}

func TestNumToWordsEs_Negative(t *testing.T) {
	got := functions.NumToWordsEs(-5)
	if got == "" {
		t.Error("NumToWordsEs(-5) returned empty")
	}
}

func TestNumToWordsEs_Values(t *testing.T) {
	// Exercise all branches: 1–29 fixed, 30-99 compound, 100 (cien/ciento),
	// 1000 (mil, prefix empty), 2000+ (prefix), millions, billions
	values := []int64{
		1, 2, 10, 11, 21, 22, 29, 30, 31, 99,
		100, 101, 200, 999,
		1000, 1001, 2000,
		1_000_000, 2_000_000,
		1_000_000_000_000, 2_000_000_000_000,
	}
	for _, n := range values {
		got := functions.NumToWordsEs(n)
		if got == "" {
			t.Errorf("NumToWordsEs(%d) returned empty", n)
		}
	}
}

func TestNumToWordsEsFloat_Basic(t *testing.T) {
	cases := []float64{0, 1.0, 1.5, -2.25, 100.99}
	for _, v := range cases {
		got := functions.NumToWordsEsFloat(v)
		if got == "" {
			t.Errorf("NumToWordsEsFloat(%v) returned empty", v)
		}
	}
}

func TestNumToWordsEsFloat_Negative(t *testing.T) {
	got := functions.NumToWordsEsFloat(-3.50)
	if got == "" {
		t.Error("NumToWordsEsFloat(-3.50) returned empty")
	}
}

// ── French (fr) ───────────────────────────────────────────────────────────────

func TestNumToWordsFr_Zero(t *testing.T) {
	if got := functions.NumToWordsFr(0); got != "zéro" {
		t.Errorf("NumToWordsFr(0) = %q, want 'zéro'", got)
	}
}

func TestNumToWordsFr_Negative(t *testing.T) {
	got := functions.NumToWordsFr(-5)
	if got == "" {
		t.Error("NumToWordsFr(-5) returned empty")
	}
}

func TestNumToWordsFr_Values(t *testing.T) {
	// French is complex: 70-79 (soixante-dix...), 80 (quatre-vingts),
	// 90-99 (quatre-vingt-dix...), cent/cents plural, mille, millions, milliards, billions
	values := []int64{
		1, 2, 10, 11, 16, 20, 21, 30, 31,
		70, 71, 72, 79,   // seventies
		80, 81, 90, 91, 99, // 80s and nineties
		100, 200, 201, 300, // hundreds with/without plural "cents"
		1000, 1001, 2000, // thousands: prefix "" (1000) vs prefix (2000)
		1_000_000, 2_000_000,
		1_000_000_000, 2_000_000_000, // milliards
		1_000_000_000_000, 2_000_000_000_000, // billions
	}
	for _, n := range values {
		got := functions.NumToWordsFr(n)
		if got == "" {
			t.Errorf("NumToWordsFr(%d) returned empty", n)
		}
	}
}

func TestNumToWordsFrFloat_Basic(t *testing.T) {
	cases := []float64{0, 1.0, 1.5, -2.25, 100.50}
	for _, v := range cases {
		got := functions.NumToWordsFrFloat(v)
		if got == "" {
			t.Errorf("NumToWordsFrFloat(%v) returned empty", v)
		}
	}
}

func TestNumToWordsFrFloat_Negative(t *testing.T) {
	got := functions.NumToWordsFrFloat(-5.75)
	if got == "" {
		t.Error("NumToWordsFrFloat(-5.75) returned empty")
	}
}

// ── numToWordsPositive (numtowords.go) — boundary coverage ───────────────────
// The 88.9% coverage gap is likely the zero return "" branch in enGbPositive.
// We cover it by calling NumToWordsEnGb with 0 (handled by the outer func)
// and by triggering the internal path via milliard/billion with no remainder.

func TestNumToWordsEnGb_MilliardExact(t *testing.T) {
	// rem == 0 branch: returns high + " milliard"
	got := functions.NumToWordsEnGb(1_000_000_000)
	if got == "" {
		t.Error("NumToWordsEnGb(1_000_000_000) returned empty")
	}
}

func TestNumToWordsEnGb_BillionExact(t *testing.T) {
	// rem == 0 branch: returns high + " billion"
	got := functions.NumToWordsEnGb(1_000_000_000_000)
	if got == "" {
		t.Error("NumToWordsEnGb(1_000_000_000_000) returned empty")
	}
}

func TestNumToWordsEnGb_MilliardWithRemainder(t *testing.T) {
	// rem > 0 branch: returns high + " milliard " + enGbPositive(rem)
	got := functions.NumToWordsEnGb(1_000_000_001)
	if got == "" {
		t.Error("NumToWordsEnGb(1_000_000_001) returned empty")
	}
}

func TestNumToWordsEnGb_BillionWithRemainder(t *testing.T) {
	// rem > 0 branch: returns high + " billion " + enGbPositive(rem)
	got := functions.NumToWordsEnGb(1_000_000_000_001)
	if got == "" {
		t.Error("NumToWordsEnGb(1_000_000_000_001) returned empty")
	}
}
