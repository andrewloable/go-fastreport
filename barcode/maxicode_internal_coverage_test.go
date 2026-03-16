package barcode

// maxicode_internal_coverage_test.go — internal tests to cover remaining
// uncovered branches in maxicode.go that require internal (package-level) access.
//
// Uncovered items from the coverage profile:
//
//  1. maxicode.go:67 init 88.9%
//     Line 81-83: `if ch > 0x7F { ch = 0x7F }` inside the Set B init loop.
//     For i=1..63, ch = 0x1F+i ranges from 0x20 to 0x5E, never exceeds 0x7F.
//     This branch is dead code in the current loop bounds — cannot be reached.
//     We verify the init ran correctly to exercise as much of the loop as possible.
//
//  2. maxicode.go:138 maxiCodeEncodeText 96.0%
//     Line 161-163: `if idx > 63 { idx = 63 }` — triggered by DEL (0x7F):
//     idx = 0x7F - 0x1F = 96 > 63 → clamped.
//     Line 158-160: `if idx < 1 { idx = 1 }` — idx = int(ch) - 0x1F < 1 means
//     ch < 0x20, but ch < 0x20 is handled by the if branch, not the else.
//     So idx < 1 is also dead code in the else branch.
//     We directly call maxiCodeEncodeText to trigger the idx > 63 branch.

import (
	"testing"
)

// TestMaxiCodeInit_SetBTableValues verifies that the init() function correctly
// populates maxiCodeSetB. At i=63, ch = 0x1F+63 = 0x5E ('~'), which is ≤ 0x7F,
// so the `if ch > 0x7F` branch at line 81 is dead code (never true for i<64).
// We verify the table was initialized correctly for boundary values.
func TestMaxiCodeInit_SetBTableValues(t *testing.T) {
	// maxiCodeSetB[0] should be 0 (LATB sentinel).
	if maxiCodeSetB[0] != 0 {
		t.Errorf("maxiCodeSetB[0] = %d, want 0 (LATB sentinel)", maxiCodeSetB[0])
	}
	// maxiCodeSetB[1] should be SP = 0x20.
	if maxiCodeSetB[1] != 0x20 {
		t.Errorf("maxiCodeSetB[1] = 0x%02X, want 0x20 (SP)", maxiCodeSetB[1])
	}
	// maxiCodeSetB[63] should be 0x5E ('^') = 0x1F+63.
	// The `if ch > 0x7F { ch = 0x7F }` at line 81 is never true for i<=63.
	expected := byte(0x1F + 63)
	if maxiCodeSetB[63] != expected {
		t.Errorf("maxiCodeSetB[63] = 0x%02X, want 0x%02X", maxiCodeSetB[63], expected)
	}
	// maxiCodeSetA[63] should be 0 (LATA sentinel).
	if maxiCodeSetA[63] != 0 {
		t.Errorf("maxiCodeSetA[63] = %d, want 0 (LATA sentinel)", maxiCodeSetA[63])
	}
	// maxiCodeSetA[0] should be 0 (NUL).
	if maxiCodeSetA[0] != 0 {
		t.Errorf("maxiCodeSetA[0] = %d, want 0 (NUL)", maxiCodeSetA[0])
	}
	// maxiCodeSetA[31] should be 0x1F (US, unit separator) — last control char.
	if maxiCodeSetA[31] != 0x1F {
		t.Errorf("maxiCodeSetA[31] = 0x%02X, want 0x1F", maxiCodeSetA[31])
	}
}

// TestMaxiCodeEncodeText_DelCharacter exercises the `if idx > 63 { idx = 63 }`
// branch at line 161 of maxicode.go.
//
// DEL (0x7F) is a printable ASCII boundary character:
//   - r = 0x7F, not > 0x7F, so it is not substituted.
//   - ch = 0x7F, not < 0x20, so it enters the else branch.
//   - idx = int(0x7F) - 0x1F = 127 - 31 = 96, which is > 63 → clamped to 63.
func TestMaxiCodeEncodeText_DelCharacter(t *testing.T) {
	// Text containing DEL (0x7F) — triggers the idx > 63 clamp.
	text := "ABC\x7FDEF"
	maxCW := 68
	cw := maxiCodeEncodeText(text, maxCW)
	if len(cw) != maxCW {
		t.Errorf("maxiCodeEncodeText: len=%d, want %d", len(cw), maxCW)
	}
	// The DEL character maps to idx=96 → clamped to 63.
	// Position 3 in cw (0-indexed) corresponds to DEL.
	if cw[3] != 63 {
		t.Errorf("cw[3] = %d, want 63 (DEL clamped to max Set B index)", cw[3])
	}
}

// TestMaxiCodeEncodeText_AllPrintableASCII encodes a string with all printable
// ASCII characters including DEL to exercise all branches in maxiCodeEncodeText.
func TestMaxiCodeEncodeText_AllPrintableASCII(t *testing.T) {
	// Build a string: space (0x20) to DEL (0x7F).
	// This exercises the normal Set B path and the idx>63 clamp.
	var text []byte
	for ch := 0x20; ch <= 0x7F; ch++ {
		text = append(text, byte(ch))
	}
	maxCW := 84
	cw := maxiCodeEncodeText(string(text), maxCW)
	if len(cw) != maxCW {
		t.Errorf("maxiCodeEncodeText: len=%d, want %d", len(cw), maxCW)
	}
}

// TestMaxiCodeEncodeText_ControlThenPrintableThenControl exercises multiple
// LATA/LATB switches including the idx>63 path for DEL within a Set B run.
func TestMaxiCodeEncodeText_ControlThenDelThenControl(t *testing.T) {
	// Sequence: control char (Set A) → DEL (Set B, idx clamped) → control char (Set A).
	text := "\x01\x7F\x02"
	maxCW := 68
	cw := maxiCodeEncodeText(text, maxCW)
	if len(cw) != maxCW {
		t.Errorf("maxiCodeEncodeText: len=%d, want %d", len(cw), maxCW)
	}
}

// TestMaxiCodeEncodeText_EmptyString exercises the padding loop at line 170.
// An empty input causes all codewords to be zero-padded.
func TestMaxiCodeEncodeText_EmptyString(t *testing.T) {
	cw := maxiCodeEncodeText("", 10)
	if len(cw) != 10 {
		t.Errorf("empty string: len=%d, want 10", len(cw))
	}
	for i, v := range cw {
		if v != 0 {
			t.Errorf("cw[%d] = %d, want 0 (pad)", i, v)
		}
	}
}
