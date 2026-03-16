package functions_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/functions"
)

// ── Dutch (nl) ────────────────────────────────────────────────────────────────

func TestNumToWordsNl_Zero(t *testing.T) {
	if got := functions.NumToWordsNl(0); got != "nul" {
		t.Errorf("NumToWordsNl(0) = %q, want 'nul'", got)
	}
}

func TestNumToWordsNl_Negative(t *testing.T) {
	got := functions.NumToWordsNl(-5)
	if got == "" {
		t.Error("NumToWordsNl(-5) returned empty")
	}
}

func TestNumToWordsNl_Values(t *testing.T) {
	values := []int64{1, 2, 3, 10, 11, 20, 21, 22, 23, 32, 33, 99, 100, 500, 999, 1000, 2000, 1_000_000, 1_000_000_000, 1_000_000_000_000}
	for _, n := range values {
		got := functions.NumToWordsNl(n)
		if got == "" {
			t.Errorf("NumToWordsNl(%d) returned empty", n)
		}
	}
}

func TestNumToWordsNlFloat_Basic(t *testing.T) {
	cases := []float64{0, 1.0, 1.5, -2.25, 100.99}
	for _, v := range cases {
		got := functions.NumToWordsNlFloat(v)
		if got == "" {
			t.Errorf("NumToWordsNlFloat(%v) returned empty", v)
		}
	}
}

func TestNumToWordsNlFloat_Negative(t *testing.T) {
	got := functions.NumToWordsNlFloat(-3.50)
	if got == "" {
		t.Error("NumToWordsNlFloat(-3.50) returned empty")
	}
}

// ── Polish (pl) ───────────────────────────────────────────────────────────────

func TestNumToWordsPl_Zero(t *testing.T) {
	if got := functions.NumToWordsPl(0); got != "zero" {
		t.Errorf("NumToWordsPl(0) = %q, want 'zero'", got)
	}
}

func TestNumToWordsPl_Negative(t *testing.T) {
	got := functions.NumToWordsPl(-7)
	if got == "" {
		t.Error("NumToWordsPl(-7) returned empty")
	}
}

func TestNumToWordsPl_Values(t *testing.T) {
	// Include values exercising gender forms, scale words, teens
	values := []int64{
		1, 2, 3, 4, 5, 10, 11, 12, 14, 20, 21, 22, 100, 999,
		1000, 1001, 2000, 4000, 5000, 1_000_000, 2_000_000,
		1_000_000_000, 1_000_000_000_000,
	}
	for _, n := range values {
		got := functions.NumToWordsPl(n)
		if got == "" {
			t.Errorf("NumToWordsPl(%d) returned empty", n)
		}
	}
}

func TestNumToWordsPlFloat_Basic(t *testing.T) {
	cases := []float64{0, 1.0, 1.5, -2.25, 100.50}
	for _, v := range cases {
		got := functions.NumToWordsPlFloat(v)
		if got == "" {
			t.Errorf("NumToWordsPlFloat(%v) returned empty", v)
		}
	}
}

func TestNumToWordsPlFloat_Negative(t *testing.T) {
	got := functions.NumToWordsPlFloat(-5.75)
	if got == "" {
		t.Error("NumToWordsPlFloat(-5.75) returned empty")
	}
}

// ── Russian (ru) ──────────────────────────────────────────────────────────────

func TestNumToWordsRu_Zero(t *testing.T) {
	got := functions.NumToWordsRu(0)
	if got == "" {
		t.Error("NumToWordsRu(0) returned empty")
	}
}

func TestNumToWordsRu_Negative(t *testing.T) {
	got := functions.NumToWordsRu(-5)
	if got == "" {
		t.Error("NumToWordsRu(-5) returned empty")
	}
}

func TestNumToWordsRu_Values(t *testing.T) {
	// Include values exercising gender forms, scale words, teens (11–19 → "many")
	values := []int64{
		1, 2, 3, 4, 5, 10, 11, 12, 13, 14, 20, 21, 22, 100, 200, 999,
		1000, 1001, 2000, 5000, 11000, 21000, 1_000_000, 2_000_000, 5_000_000,
		1_000_000_000, 2_000_000_000, 1_000_000_000_000,
	}
	for _, n := range values {
		got := functions.NumToWordsRu(n)
		if got == "" {
			t.Errorf("NumToWordsRu(%d) returned empty", n)
		}
	}
}

func TestNumToWordsRuFloat_Basic(t *testing.T) {
	cases := []float64{0, 1.0, 1.5, -2.25, 100.99}
	for _, v := range cases {
		got := functions.NumToWordsRuFloat(v)
		if got == "" {
			t.Errorf("NumToWordsRuFloat(%v) returned empty", v)
		}
	}
}

func TestNumToWordsRuFloat_Negative(t *testing.T) {
	got := functions.NumToWordsRuFloat(-3.50)
	if got == "" {
		t.Error("NumToWordsRuFloat(-3.50) returned empty")
	}
}

// ── Ukrainian (uk) ────────────────────────────────────────────────────────────

func TestNumToWordsUk_Zero(t *testing.T) {
	got := functions.NumToWordsUk(0)
	if got == "" {
		t.Error("NumToWordsUk(0) returned empty")
	}
}

func TestNumToWordsUk_Negative(t *testing.T) {
	got := functions.NumToWordsUk(-3)
	if got == "" {
		t.Error("NumToWordsUk(-3) returned empty")
	}
}

func TestNumToWordsUk_Values(t *testing.T) {
	values := []int64{
		1, 2, 3, 4, 5, 10, 11, 12, 14, 20, 21, 22, 100, 200, 999,
		1000, 1001, 2000, 5000, 11000, 21000, 1_000_000, 2_000_000, 5_000_000,
		1_000_000_000, 2_000_000_000, 1_000_000_000_000,
	}
	for _, n := range values {
		got := functions.NumToWordsUk(n)
		if got == "" {
			t.Errorf("NumToWordsUk(%d) returned empty", n)
		}
	}
}

func TestNumToWordsUkFloat_Basic(t *testing.T) {
	cases := []float64{0, 1.0, 1.5, -2.25, 100.50}
	for _, v := range cases {
		got := functions.NumToWordsUkFloat(v)
		if got == "" {
			t.Errorf("NumToWordsUkFloat(%v) returned empty", v)
		}
	}
}

func TestNumToWordsUkFloat_Negative(t *testing.T) {
	got := functions.NumToWordsUkFloat(-5.75)
	if got == "" {
		t.Error("NumToWordsUkFloat(-5.75) returned empty")
	}
}

// ── Spanish-alt (sp) ──────────────────────────────────────────────────────────

func TestNumToWordsSp_Zero(t *testing.T) {
	if got := functions.NumToWordsSp(0); got != "cero" {
		t.Errorf("NumToWordsSp(0) = %q, want 'cero'", got)
	}
}

func TestNumToWordsSp_Negative(t *testing.T) {
	got := functions.NumToWordsSp(-5)
	if got == "" {
		t.Error("NumToWordsSp(-5) returned empty")
	}
}

func TestNumToWordsSp_Values(t *testing.T) {
	// Include 1, 21-29 (using spFixed), 100+ (hundreds), 1000+, millions, billards
	values := []int64{
		1, 2, 10, 11, 21, 22, 25, 29, 30, 31, 99, 100, 101, 200, 999,
		1000, 1001, 2000, 1_000_000, 1_000_000_000, 2_000_000_000,
		1_000_000_000_000, 2_000_000_000_000,
	}
	for _, n := range values {
		got := functions.NumToWordsSp(n)
		if got == "" {
			t.Errorf("NumToWordsSp(%d) returned empty", n)
		}
	}
}

func TestNumToWordsSpFloat_Basic(t *testing.T) {
	cases := []float64{0, 1.0, 1.5, -2.25, 100.99}
	for _, v := range cases {
		got := functions.NumToWordsSpFloat(v)
		if got == "" {
			t.Errorf("NumToWordsSpFloat(%v) returned empty", v)
		}
	}
}

func TestNumToWordsSpFloat_Negative(t *testing.T) {
	got := functions.NumToWordsSpFloat(-3.50)
	if got == "" {
		t.Error("NumToWordsSpFloat(-3.50) returned empty")
	}
}

// ── Farsi/Persian (fa) ────────────────────────────────────────────────────────

func TestNumToWordsFa_Zero(t *testing.T) {
	got := functions.NumToWordsFa(0)
	if got == "" {
		t.Error("NumToWordsFa(0) returned empty")
	}
}

func TestNumToWordsFa_Negative(t *testing.T) {
	got := functions.NumToWordsFa(-5)
	if got == "" {
		t.Error("NumToWordsFa(-5) returned empty")
	}
}

func TestNumToWordsFa_Values(t *testing.T) {
	values := []int64{
		1, 2, 10, 11, 20, 21, 30, 99, 100, 200, 999,
		1000, 1001, 1_000_000, 1_000_000_000, 1_000_000_000_000,
	}
	for _, n := range values {
		got := functions.NumToWordsFa(n)
		if got == "" {
			t.Errorf("NumToWordsFa(%d) returned empty", n)
		}
	}
}

func TestNumToWordsFaFloat_Basic(t *testing.T) {
	cases := []float64{0, 1.0, 1.5, -2.25, 100.50}
	for _, v := range cases {
		got := functions.NumToWordsFaFloat(v)
		if got == "" {
			t.Errorf("NumToWordsFaFloat(%v) returned empty", v)
		}
	}
}

func TestNumToWordsFaFloat_Negative(t *testing.T) {
	got := functions.NumToWordsFaFloat(-5.75)
	if got == "" {
		t.Error("NumToWordsFaFloat(-5.75) returned empty")
	}
}

// ── Indian English (in) ───────────────────────────────────────────────────────

func TestNumToWordsIn_Zero(t *testing.T) {
	if got := functions.NumToWordsIn(0); got != "zero" {
		t.Errorf("NumToWordsIn(0) = %q, want 'zero'", got)
	}
}

func TestNumToWordsIn_Negative(t *testing.T) {
	got := functions.NumToWordsIn(-5)
	if got == "" {
		t.Error("NumToWordsIn(-5) returned empty")
	}
}

func TestNumToWordsIn_Values(t *testing.T) {
	// Indian system: thousand, lakh (1e5), crore (1e7), arab (1e9), kharab (1e11), nil (1e13)
	values := []int64{
		1, 2, 10, 11, 19, 20, 21, 99, 100, 101, 999,
		1000, 1001, 99_999,
		100_000,    // one lakh
		1_000_000,  // ten lakh
		10_000_000, // one crore
		1_000_000_000,    // one arab
		100_000_000_000,  // one kharab
		10_000_000_000_000, // one nil
		// 3-digit arab groups to exercise in3digits n>=100 branch:
		// For arab scale (10^9): g = n/10^9 can be 100-999 if n < kharab (10^11)
		100_000_000_000 + 500, // kharab with remainder → g=1 (kharab scale), then arab of remainder
		// Direct arab > 100: 500 arab = 500*10^9 < kharab, so g=500 for arab scale
		500_000_000_000, // > kharab: g = 500_000_000_000/10^11 = 5 (kharab)
		// For in3digits n>=100: need g >= 100 for a scale
		// Arab max without kharab: up to 99*10^9; kharab starts at 10^11
		// Crore (10^7) max g < 10 (since > 10*10^7 = lakh)
		// Wait: lakh=10^5, crore=10^7 → crore g = n/10^7 can be 1-9 (since n < arab=10^9 means g < 100)
		// arab=10^9, g can be 1-99 (since n < kharab=10^11 means g < 100)
		// kharab=10^11, g can be 1-99 (since n < nil=10^13 means g < 100)
		// thousand=10^3, g can be 1-99 (since n < lakh=10^5 means g < 100)
		// Actually for kharab: g = n/10^11, max g when n < 10^13 is 99. So g is still < 100.
		// For in3digits to hit n>=100, we need g >= 100 which means n >= 100 * scale_value.
		// That only happens if n >= 100*kharab = 10^13 = nil scale.
		// Actually nil is 10^13, kharab is 10^11: 100*kharab = 10^13 = nil. So they're handled separately.
		// Conclusion: in3digits(n>=100) is unreachable from normal Indian numbering flows.
		// Testing the function directly via NumToWordsIn exercising all standard paths.
	}
	for _, n := range values {
		got := functions.NumToWordsIn(n)
		if got == "" {
			t.Errorf("NumToWordsIn(%d) returned empty", n)
		}
	}
}

func TestNumToWordsInFloat_Basic(t *testing.T) {
	cases := []float64{0, 1.0, 1.5, -2.25, 100.99}
	for _, v := range cases {
		got := functions.NumToWordsInFloat(v)
		if got == "" {
			t.Errorf("NumToWordsInFloat(%v) returned empty", v)
		}
	}
}

func TestNumToWordsInFloat_Negative(t *testing.T) {
	got := functions.NumToWordsInFloat(-3.50)
	if got == "" {
		t.Error("NumToWordsInFloat(-3.50) returned empty")
	}
}
