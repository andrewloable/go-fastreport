package functions

// Internal white-box tests for unexported helpers that have dead-code branches
// unreachable from the public API. These tests use the internal package to
// directly call unexported functions and cover the remaining branches.

import "testing"

// TestNumToWordsPositive_Zero covers the n==0 early-return branch (line 70)
// which is unreachable from NumToWords (that handles 0 before calling this).
func TestNumToWordsPositive_Zero(t *testing.T) {
	got := numToWordsPositive(0)
	if got != "" {
		t.Errorf("numToWordsPositive(0) = %q, want empty string", got)
	}
}

// TestFr2Digits_Female covers the female=true branches (lines 137-138, 167-168)
// which are unreachable from NumToWordsFr (always passes female=false).
func TestFr2Digits_FemaleOne(t *testing.T) {
	// n < 20, n == 1, female == true → returns "une"
	got := fr2Digits(1, true, false)
	if got != "une" {
		t.Errorf("fr2Digits(1, true, false) = %q, want 'une'", got)
	}
}

func TestFr2Digits_FemaleTensEtUne(t *testing.T) {
	// o == 1, female == true → oStr = "une", returns "vingt et une"
	got := fr2Digits(21, true, false)
	if got != "vingt et une" {
		t.Errorf("fr2Digits(21, true, false) = %q, want 'vingt et une'", got)
	}
}
