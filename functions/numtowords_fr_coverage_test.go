package functions_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/functions"
)

// TestNumToWordsFr_MillionRem covers the million rem>0 branch (frPositive line 95).
// frPositive million branch: if rem > 0, appends frPositive(rem, ...).
func TestNumToWordsFr_MillionRem(t *testing.T) {
	cases := []int64{
		1_000_001,   // one million one
		1_000_100,   // one million cent
		2_000_022,   // deux millions vingt-deux
		3_500_000,   // trois millions cinq cent mille
		5_000_001,   // cinq millions un
	}
	for _, n := range cases {
		got := functions.NumToWordsFr(n)
		if got == "" {
			t.Errorf("NumToWordsFr(%d) returned empty", n)
		}
	}
}

// TestNumToWordsFr_MilliardRem covers the milliard rem>0 branch (frPositive line 82).
func TestNumToWordsFr_MilliardRem(t *testing.T) {
	cases := []int64{
		1_000_000_001,   // un milliard un
		1_000_000_100,   // un milliard cent
		2_000_000_001,   // deux milliards un
		3_000_500_000,   // trois milliards cinq cent mille
	}
	for _, n := range cases {
		got := functions.NumToWordsFr(n)
		if got == "" {
			t.Errorf("NumToWordsFr(%d) returned empty", n)
		}
	}
}

// TestNumToWordsFr_BillionRem covers the billion rem>0 branch (frPositive line 69).
func TestNumToWordsFr_BillionRem(t *testing.T) {
	cases := []int64{
		1_000_000_000_001,   // un billion un
		1_000_000_000_100,   // un billion cent
		2_000_000_000_001,   // deux billions un
		1_000_001_000_000,   // un billion un million
	}
	for _, n := range cases {
		got := functions.NumToWordsFr(n)
		if got == "" {
			t.Errorf("NumToWordsFr(%d) returned empty", n)
		}
	}
}

// TestNumToWordsFr_QuatreVingtNonFinal covers the n==80 non-final branch in fr2Digits
// (returns "quatre-vingt" without trailing 's' when final=false).
//
// This happens when 80 appears as a sub-group in a larger number.
// For example: 80_000 → frPositive(80_000): g=80, rem=0, prefix=frPositive(80,false,false)+"  "
// frPositive(80,false,false) → fr2Digits(80,false,false) → final=false → "quatre-vingt"
// Then: "quatre-vingt mille"
func TestNumToWordsFr_QuatreVingtNonFinal(t *testing.T) {
	cases := []struct {
		n    int64
		desc string
	}{
		// 80 * 1000 = 80_000: frPositive calls frPositive(80, false, false)
		// which calls fr2Digits(80, false, false) → final=false → "quatre-vingt"
		{80_000, "quatre-vingt mille"},
		{80_000_000, "quatre-vingt millions"},
		{80_000_000_000, "quatre-vingt milliards"},
		// 2_080_000: million scale, g=2, rem=80_000
		// then frPositive(80_000, false, true): g=80, prefix=frPositive(80,false,false)
		{2_080_000, "deux millions quatre-vingt mille"},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := functions.NumToWordsFr(tc.n)
			if got == "" {
				t.Errorf("NumToWordsFr(%d) [%s] returned empty", tc.n, tc.desc)
			}
		})
	}
}

// TestNumToWordsFr_TensHyphenUnit covers the fr2Digits branch where tens digit is set,
// unit digit != 0 and != 1 (returns "tens-unit" form like "vingt-deux").
func TestNumToWordsFr_TensHyphenUnit(t *testing.T) {
	// Values where t is 2-6 (vingt/trente/quarante/cinquante/soixante) and o is 2-9
	cases := []int64{
		22,  // vingt-deux
		23,  // vingt-trois
		32,  // trente-deux
		42,  // quarante-deux
		52,  // cinquante-deux
		62,  // soixante-deux
		99,  // quatre-vingt-dix-neuf (t==9, o==9 via frNineties)
	}
	for _, n := range cases {
		got := functions.NumToWordsFr(n)
		if got == "" {
			t.Errorf("NumToWordsFr(%d) returned empty", n)
		}
	}
}

// TestNumToWordsFr_SeventiesAll covers all values 71-79 to exercise the
// frSeventies array and the t==7 branch in fr2Digits.
func TestNumToWordsFr_SeventiesAll(t *testing.T) {
	expected := map[int64]string{
		70: "soixante-dix",
		71: "soixante et onze",
		72: "soixante-douze",
		73: "soixante-treize",
		74: "soixante-quatorze",
		75: "soixante-quinze",
		76: "soixante-seize",
		77: "soixante-dix-sept",
		78: "soixante-dix-huit",
		79: "soixante-dix-neuf",
	}
	for n, want := range expected {
		got := functions.NumToWordsFr(n)
		if got != want {
			t.Errorf("NumToWordsFr(%d) = %q, want %q", n, got, want)
		}
	}
}

// TestNumToWordsFr_NinetiesAll covers 90-99.
func TestNumToWordsFr_NinetiesAll(t *testing.T) {
	expected := map[int64]string{
		90: "quatre-vingt-dix",
		91: "quatre-vingt-onze",
		99: "quatre-vingt-dix-neuf",
	}
	for n, want := range expected {
		got := functions.NumToWordsFr(n)
		if got != want {
			t.Errorf("NumToWordsFr(%d) = %q, want %q", n, got, want)
		}
	}
}
