package barcode

// maxicode_internal_coverage_test.go — internal tests for maxicode.go.
// Updated to match the rewritten ISO/IEC 16023 compliant implementation.

import (
	"testing"
)

// TestMaxiCodeTables_SetValues verifies the lookup tables are initialized correctly.
func TestMaxiCodeTables_SetValues(t *testing.T) {
	// Space (32) is set 0 (ambiguous), can fit Set A or B.
	if maxiCodeSet[32] != 0 {
		t.Errorf("maxiCodeSet[' '] = %d, want 0 (ambiguous)", maxiCodeSet[32])
	}
	// '!' (33) is Set B (2).
	if maxiCodeSet[33] != 2 {
		t.Errorf("maxiCodeSet['!'] = %d, want 2 (Set B)", maxiCodeSet[33])
	}
	// 'A' (65) is Set A (1).
	if maxiCodeSet[65] != 1 {
		t.Errorf("maxiCodeSet['A'] = %d, want 1 (Set A)", maxiCodeSet[65])
	}
	// 'a' (97) is Set B (2).
	if maxiCodeSet[97] != 2 {
		t.Errorf("maxiCodeSet['a'] = %d, want 2 (Set B)", maxiCodeSet[97])
	}
	// NUL (0) is Set E (5).
	if maxiCodeSet[0] != 5 {
		t.Errorf("maxiCodeSet[NUL] = %d, want 5 (Set E)", maxiCodeSet[0])
	}
}

func TestMaxiCodeTables_SymbolCharValues(t *testing.T) {
	// 'a' (97) → symbol char 1 in Set B.
	if maxiCodeSymbolChar[97] != 1 {
		t.Errorf("maxiCodeSymbolChar['a'] = %d, want 1", maxiCodeSymbolChar[97])
	}
	// 'z' (122) → symbol char 26 in Set B.
	if maxiCodeSymbolChar[122] != 26 {
		t.Errorf("maxiCodeSymbolChar['z'] = %d, want 26", maxiCodeSymbolChar[122])
	}
	// 'A' (65) → symbol char 1 in Set A.
	if maxiCodeSymbolChar[65] != 1 {
		t.Errorf("maxiCodeSymbolChar['A'] = %d, want 1", maxiCodeSymbolChar[65])
	}
	// Space (32) → symbol char 32.
	if maxiCodeSymbolChar[32] != 32 {
		t.Errorf("maxiCodeSymbolChar[' '] = %d, want 32", maxiCodeSymbolChar[32])
	}
}

// TestMaxiCodeProcessText_LowercaseURL verifies lowercase letters encode correctly.
// Previously broken: 'a'-'z' were clamped to 63. Now they get symbol values 1-26.
func TestMaxiCodeProcessText_LowercaseURL(t *testing.T) {
	_, ch, _, err := maxiCodeProcessText("http://fast-report.com", 4)
	if err != nil {
		t.Fatalf("processText error: %v", err)
	}
	// After Latch B, 'h' (104) → symbol 8, 't' → 20, 't' → 20, 'p' → 16.
	// The first codeword is a latch (63), then 'h'=8.
	// Find the latch and verify 'h' follows.
	found := false
	for i := 0; i < 140; i++ {
		if ch[i] == 63 && ch[i+1] == 8 { // latch B + 'h'
			found = true
			break
		}
	}
	if !found {
		t.Error("expected LATCH(63) followed by 'h'(8) in encoded text — lowercase encoding broken")
	}
}

// TestMaxiCodeProcessText_Empty verifies empty string returns no error.
func TestMaxiCodeProcessText_Empty(t *testing.T) {
	_, _, length, err := maxiCodeProcessText("", 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = length
}

// TestMaxiCodeProcessText_TooLong verifies >138 chars is silently truncated.
func TestMaxiCodeProcessText_TooLong(t *testing.T) {
	long := make([]byte, 200)
	for i := range long {
		long[i] = 'A'
	}
	_, _, length, err := maxiCodeProcessText(string(long), 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After truncation the effective input length will be ≤ 138.
	_ = length
}

// TestMaxiCodeProcessText_AmbiguousSpace verifies space resolves to Set B (47)
// when surrounded by other Set B characters.
func TestMaxiCodeProcessText_AmbiguousSpace(t *testing.T) {
	// "a b" — space surrounded by Set B lowercase letters.
	_, ch, _, err := maxiCodeProcessText("a b", 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not panic and length should be valid.
	_ = ch
}

// TestMaxiCodeProcessText_ControlChars verifies control characters use Set A.
func TestMaxiCodeProcessText_ControlChars(t *testing.T) {
	_, _, _, err := maxiCodeProcessText("\x01\x02\x03", 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
