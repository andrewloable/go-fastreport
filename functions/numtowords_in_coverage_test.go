package functions_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/functions"
)

// TestNumToWordsIn_in3digitsHundreds exercises the n >= 100 branch of in3digits.
//
// in3digits is called from inPositive as in3digits(g) where g = rem / scale.value.
// For the nil scale (10^13), g can be >= 100 when n >= 100 * 10^13 = 10^15.
// int64 max is ~9.2e18, so values up to 999 * 10^13 are valid.
func TestNumToWordsIn_in3digitsHundreds(t *testing.T) {
	const nil_ int64 = 10_000_000_000_000 // 10^13

	cases := []struct {
		n    int64
		desc string
	}{
		// g = 100: "one hundred nil" — hits n>=100 branch with rem==0
		{100 * nil_, "100 nil: in3digits(100) with no remainder"},
		// g = 101: "one hundred one nil" — hits n>=100 branch with rem>0
		{101 * nil_, "101 nil: in3digits(101) with remainder 1"},
		// g = 150: "one hundred fifty nil" — hits n>=100 branch with rem=50
		{150 * nil_, "150 nil: in3digits(150) with remainder 50"},
		// g = 200: "two hundred nil"
		{200 * nil_, "200 nil: in3digits(200) with hundreds=2, rem=0"},
		// g = 999: "nine hundred ninety-nine nil" — maximum 3-digit group
		{999 * nil_, "999 nil: in3digits(999) maximum"},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := functions.NumToWordsIn(tc.n)
			if got == "" {
				t.Errorf("NumToWordsIn(%d) [%s] returned empty", tc.n, tc.desc)
			}
		})
	}
}

// TestNumToWordsIn_in3digitsHundredsWithKharab exercises in3digits(g>=100) via
// the kharab scale (10^11). kharab g can be 1-99 for n < nil (10^13), so only
// reaches 3-digit group if n is at the nil scale level. The nil-level tests
// above cover this. Additionally verify expected output for a known value.
func TestNumToWordsIn_in3digitsExactOutput(t *testing.T) {
	const nil_ int64 = 10_000_000_000_000

	// 100 nil should produce "one hundred nil"
	got := functions.NumToWordsIn(100 * nil_)
	expected := "one hundred nil"
	if got != expected {
		t.Errorf("NumToWordsIn(%d) = %q, want %q", 100*nil_, got, expected)
	}

	// 150 nil should produce "one hundred fifty nil"
	got2 := functions.NumToWordsIn(150 * nil_)
	expected2 := "one hundred fifty nil"
	if got2 != expected2 {
		t.Errorf("NumToWordsIn(%d) = %q, want %q", 150*nil_, got2, expected2)
	}

	// 101 nil should produce "one hundred one nil"
	got3 := functions.NumToWordsIn(101 * nil_)
	expected3 := "one hundred one nil"
	if got3 != expected3 {
		t.Errorf("NumToWordsIn(%d) = %q, want %q", 101*nil_, got3, expected3)
	}
}

// TestNumToWordsIn_ScaleCoverage exercises all Indian number scales to ensure
// full coverage of inPositive and its lakh/crore/arab/kharab/nil paths.
func TestNumToWordsIn_ScaleCoverage(t *testing.T) {
	cases := []struct {
		n    int64
		desc string
	}{
		{1_00_000, "one lakh"},
		{2_00_000, "two lakh"},
		{10_00_000, "ten lakh"},
		{99_00_000, "ninety-nine lakh"},
		{1_00_00_000, "one crore"},
		{5_00_00_000, "five crore"},
		{1_00_00_00_000, "one arab"},
		{99_00_00_00_000, "ninety-nine arab"},
		{1_00_00_00_00_000, "one kharab"},
		{50_00_00_00_00_000, "fifty kharab"},
		{10_00_00_00_00_000_00, "one nil"},
		// Numbers combining multiple scales
		{1_00_000 + 1, "one lakh one"},
		{1_00_00_000 + 1_00_000 + 1, "one crore one lakh one"},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := functions.NumToWordsIn(tc.n)
			if got == "" {
				t.Errorf("NumToWordsIn(%d) [%s] returned empty", tc.n, tc.desc)
			}
		})
	}
}
