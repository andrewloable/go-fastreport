// gs1_helper_internal_test.go — internal tests for GS1 helper functions.
//
// Verifies gs1FindAIIndex, gs1GetCode, and gs1ParseGS1 against the
// C# GS1Helper.cs algorithms and spot-checks the AI table for correctness.
//
// C# reference: FastReport.Base/Barcode/GS1Helper.cs
package barcode

import (
	"strings"
	"testing"
)

// ── gs1AICodes table verification ─────────────────────────────────────────────

// TestGS1AICodes_TableSize verifies the table has exactly 89 entries,
// matching the C# GS1Helper.cs AICodes list (counted via `new AI(` occurrences).
func TestGS1AICodes_TableSize(t *testing.T) {
	const want = 89
	if got := len(gs1AICodes); got != want {
		t.Errorf("gs1AICodes length = %d, want %d", got, want)
	}
}

// TestGS1AICodes_FirstEntry verifies the first entry is AI "00", fixed-length 18.
func TestGS1AICodes_FirstEntry(t *testing.T) {
	ai := gs1AICodes[0]
	if ai.numAI != "00" || ai.numAILength != 2 || ai.minDataLength != 18 || ai.maxDataLength != 18 || ai.useFNC1 {
		t.Errorf("first entry = %+v, want {00 2 18 18 false}", ai)
	}
}

// TestGS1AICodes_LastEntry verifies the last entry is AI "9X", variable-length 90.
func TestGS1AICodes_LastEntry(t *testing.T) {
	ai := gs1AICodes[len(gs1AICodes)-1]
	if ai.numAI != "9X" || ai.numAILength != 2 || ai.minDataLength != 1 || ai.maxDataLength != 90 || !ai.useFNC1 {
		t.Errorf("last entry = %+v, want {9X 2 1 90 true}", ai)
	}
}

// TestGS1AICodes_SpotCheck verifies a selection of AI entries against the C# table.
// Each case is: (numAI, numAILength, minDataLength, maxDataLength, useFNC1).
func TestGS1AICodes_SpotCheck(t *testing.T) {
	// Map of expected entries keyed by numAI; used to find & verify each AI.
	type expected struct {
		numAILength   int
		minDataLength int
		maxDataLength int
		useFNC1       bool
	}
	wants := map[string]expected{
		// C# line 27-29: basic fixed-length AIs
		"00": {2, 18, 18, false},
		"01": {2, 14, 14, false},
		"02": {2, 14, 14, false},
		// C# line 30: first variable-length AI
		"10": {2, 1, 20, true},
		// C# line 31-36: date AIs (no 14)
		"11": {2, 6, 6, false},
		"12": {2, 6, 6, false},
		"13": {2, 6, 6, false},
		"15": {2, 6, 6, false},
		"16": {2, 6, 6, false},
		"17": {2, 6, 6, false},
		// C# line 37-39: product AIs
		"20": {2, 2, 2, false},
		"21": {2, 1, 20, true},
		"22": {2, 1, 20, true},
		// C# line 40-48: 3-digit AIs
		"240": {3, 1, 30, true},
		"241": {3, 1, 30, true},
		"242": {3, 1, 6, true},
		"243": {3, 1, 20, true},
		"250": {3, 1, 30, true},
		"251": {3, 1, 30, true},
		"253": {3, 13, 30, true},
		"254": {3, 1, 20, true},
		"255": {3, 13, 25, true},
		// C# line 49: count AI
		"30": {2, 1, 8, true},
		// C# line 50-55: measurement AIs with wildcard
		"31XX": {4, 6, 6, false},
		"32XX": {4, 6, 6, false},
		"33XX": {4, 6, 6, false},
		"34XX": {4, 6, 6, false},
		"35XX": {4, 6, 6, false},
		"36XX": {4, 6, 6, false},
		// C# line 56: quantity
		"37": {2, 1, 8, true},
		// C# line 57-61: monetary AIs
		"390X": {4, 1, 15, true},
		"391X": {4, 3, 18, true},
		"392X": {4, 1, 15, true},
		"393X": {4, 3, 18, true},
		"394X": {4, 4, 4, true},
		// C# line 62-65: order/routing AIs
		"400": {3, 1, 30, true},
		"401": {3, 1, 30, true},
		"402": {3, 17, 17, true},
		"403": {3, 1, 30, true},
		// C# line 66: ship-to AI
		"41X": {3, 13, 13, false},
		// C# line 67-73: location AIs
		"420": {3, 1, 20, true},
		"421": {3, 3, 12, true},
		"422": {3, 3, 3, true},
		"423": {3, 3, 15, true},
		"424": {3, 3, 3, true},
		"425": {3, 3, 15, true},
		"426": {3, 3, 3, true},
		// C# line 74-88: 7xxx AIs
		"7001": {4, 13, 13, true},
		"7002": {4, 1, 30, true},
		"7003": {4, 10, 10, true},
		"7004": {4, 1, 4, true},
		"7005": {4, 1, 12, true},
		"7006": {4, 6, 6, true},
		"7007": {4, 6, 12, true},
		"7008": {4, 1, 3, true},
		"7009": {4, 1, 10, true},
		"7010": {4, 1, 2, true},
		"7020": {4, 1, 20, true},
		"7021": {4, 1, 20, true},
		"7022": {4, 1, 20, true},
		"7023": {4, 1, 30, true},
		"703X": {4, 3, 30, true},
		// C# line 89-93: 71x AIs
		"710": {3, 1, 20, true},
		"711": {3, 1, 20, true},
		"712": {3, 1, 20, true},
		"713": {3, 1, 20, true},
		"71X": {3, 1, 20, true},
		// C# line 94-113: 8xxx AIs
		"8001": {4, 14, 14, true},
		"8002": {4, 1, 20, true},
		"8003": {4, 14, 30, true},
		"8004": {4, 1, 30, true},
		"8005": {4, 6, 6, true},
		"8006": {4, 18, 18, true},
		"8007": {4, 1, 34, true},
		"8008": {4, 8, 12, true},
		"8010": {4, 1, 30, true},
		"8011": {4, 1, 12, true},
		"8012": {4, 1, 20, true},
		"8013": {4, 1, 30, true},
		"8017": {4, 18, 18, true},
		"8018": {4, 18, 18, true},
		"8019": {4, 1, 10, true},
		"8020": {4, 1, 25, true},
		"8110": {4, 1, 70, true},
		"8111": {4, 4, 4, true},
		"8112": {4, 1, 70, true},
		"8200": {4, 1, 70, true},
		// C# line 114-115: internal-use AIs
		"90": {2, 1, 30, true},
		"9X": {2, 1, 90, true},
	}

	// Build a lookup from the Go table.
	got := make(map[string]gs1AICode, len(gs1AICodes))
	for _, ai := range gs1AICodes {
		got[ai.numAI] = ai
	}

	for numAI, w := range wants {
		ai, ok := got[numAI]
		if !ok {
			t.Errorf("AI %q missing from gs1AICodes", numAI)
			continue
		}
		if ai.numAILength != w.numAILength ||
			ai.minDataLength != w.minDataLength ||
			ai.maxDataLength != w.maxDataLength ||
			ai.useFNC1 != w.useFNC1 {
			t.Errorf("AI %q: got {len=%d min=%d max=%d fnc1=%v}, want {len=%d min=%d max=%d fnc1=%v}",
				numAI,
				ai.numAILength, ai.minDataLength, ai.maxDataLength, ai.useFNC1,
				w.numAILength, w.minDataLength, w.maxDataLength, w.useFNC1)
		}
	}
}

// TestGS1AICodes_NumAILengthMatchesStringLength verifies that for every entry
// the stored numAILength equals len(numAI).
// C# FindAIIndex uses AICodes[i].numAI.Length (string length), not numAILength.
// These must be equal for the Go port to behave identically.
func TestGS1AICodes_NumAILengthMatchesStringLength(t *testing.T) {
	for i, ai := range gs1AICodes {
		if len(ai.numAI) != ai.numAILength {
			t.Errorf("entry[%d] AI=%q: numAILength=%d != len(numAI)=%d",
				i, ai.numAI, ai.numAILength, len(ai.numAI))
		}
	}
}

// ── gs1FindAIIndex ─────────────────────────────────────────────────────────────

// TestGS1FindAIIndex_NegativeIndex verifies that a negative index returns -1.
// C# checks `if (index == -1) return -1` before bounds-checking the string.
func TestGS1FindAIIndex_NegativeIndex(t *testing.T) {
	if got := gs1FindAIIndex("(01)12345678901231", -1); got != -1 {
		t.Errorf("gs1FindAIIndex(code, -1) = %d, want -1", got)
	}
}

// TestGS1FindAIIndex_OutOfBounds verifies that index >= len(code) returns -1.
// Go adds this guard (C# would throw IndexOutOfRangeException).
func TestGS1FindAIIndex_OutOfBounds(t *testing.T) {
	code := "(01)12345678901231"
	if got := gs1FindAIIndex(code, len(code)); got != -1 {
		t.Errorf("gs1FindAIIndex(code, len(code)) = %d, want -1", got)
	}
}

// TestGS1FindAIIndex_NotOpenParen verifies that a non-'(' character at index
// returns -1 immediately (C# GS1Helper.cs:126-127).
func TestGS1FindAIIndex_NotOpenParen(t *testing.T) {
	if got := gs1FindAIIndex("01)12345678901231", 0); got != -1 {
		t.Errorf("gs1FindAIIndex: non-'(' at index 0 = %d, want -1", got)
	}
}

// TestGS1FindAIIndex_TooShort verifies that codeLen < 3 returns -1
// (C# GS1Helper.cs:130-131).
func TestGS1FindAIIndex_TooShort(t *testing.T) {
	// "(0" has len=2 from index 0, so codeLen=2 < 3.
	if got := gs1FindAIIndex("(0", 0); got != -1 {
		t.Errorf("gs1FindAIIndex: codeLen<3 = %d, want -1", got)
	}
}

// TestGS1FindAIIndex_AI00 verifies finding AI "00" at position 0.
func TestGS1FindAIIndex_AI00(t *testing.T) {
	code := "(00)123456789012345678"
	got := gs1FindAIIndex(code, 0)
	if got < 0 {
		t.Fatalf("gs1FindAIIndex AI=00: returned %d, want >= 0", got)
	}
	if gs1AICodes[got].numAI != "00" {
		t.Errorf("gs1FindAIIndex AI=00: found %q, want 00", gs1AICodes[got].numAI)
	}
}

// TestGS1FindAIIndex_AI01 verifies finding AI "01" (GTIN-14).
func TestGS1FindAIIndex_AI01(t *testing.T) {
	code := "(01)12345678901231"
	got := gs1FindAIIndex(code, 0)
	if got < 0 || gs1AICodes[got].numAI != "01" {
		t.Errorf("gs1FindAIIndex AI=01: got index %d (%q), want AI=01",
			got, func() string {
				if got >= 0 {
					return gs1AICodes[got].numAI
				}
				return "<not found>"
			}())
	}
}

// TestGS1FindAIIndex_AI31XX verifies that wildcard AI "31XX" matches "3100",
// "3199", "3157" etc. per C# matching logic (AICodes[i].numAI[j] != 'X').
func TestGS1FindAIIndex_AI31XX(t *testing.T) {
	for _, variant := range []string{"3100", "3150", "3199"} {
		code := "(" + variant + ")123456"
		got := gs1FindAIIndex(code, 0)
		if got < 0 {
			t.Errorf("gs1FindAIIndex AI=%q: not found", variant)
			continue
		}
		if gs1AICodes[got].numAI != "31XX" {
			t.Errorf("gs1FindAIIndex AI=%q: found %q, want 31XX", variant, gs1AICodes[got].numAI)
		}
	}
}

// TestGS1FindAIIndex_AI41X verifies that wildcard AI "41X" matches "410"–"416".
func TestGS1FindAIIndex_AI41X(t *testing.T) {
	for _, variant := range []string{"410", "411", "416"} {
		code := "(" + variant + ")1234567890123"
		got := gs1FindAIIndex(code, 0)
		if got < 0 {
			t.Errorf("gs1FindAIIndex AI=%q: not found", variant)
			continue
		}
		if gs1AICodes[got].numAI != "41X" {
			t.Errorf("gs1FindAIIndex AI=%q: found %q, want 41X", variant, gs1AICodes[got].numAI)
		}
	}
}

// TestGS1FindAIIndex_AI9X verifies that wildcard AI "9X" matches "91"–"99".
func TestGS1FindAIIndex_AI9X(t *testing.T) {
	for _, variant := range []string{"91", "95", "99"} {
		code := "(" + variant + ")ABCDEF"
		got := gs1FindAIIndex(code, 0)
		if got < 0 {
			t.Errorf("gs1FindAIIndex AI=%q: not found", variant)
			continue
		}
		if gs1AICodes[got].numAI != "9X" {
			t.Errorf("gs1FindAIIndex AI=%q: found %q, want 9X", variant, gs1AICodes[got].numAI)
		}
	}
}

// TestGS1FindAIIndex_AI703X verifies wildcard AI "703X" matches "7030"–"7039".
func TestGS1FindAIIndex_AI703X(t *testing.T) {
	for _, variant := range []string{"7030", "7035", "7039"} {
		code := "(" + variant + ")ABCDEF"
		got := gs1FindAIIndex(code, 0)
		if got < 0 {
			t.Errorf("gs1FindAIIndex AI=%q: not found", variant)
			continue
		}
		if gs1AICodes[got].numAI != "703X" {
			t.Errorf("gs1FindAIIndex AI=%q: found %q, want 703X", variant, gs1AICodes[got].numAI)
		}
	}
}

// TestGS1FindAIIndex_UnknownAI verifies that an unrecognised AI (e.g. "99")
// does NOT match "9X" when data length constraint prevents variable-length match.
// Actually "99" SHOULD match "9X" — this test verifies that.
func TestGS1FindAIIndex_AI99_MatchesWildcard9X(t *testing.T) {
	code := "(99)HELLO"
	got := gs1FindAIIndex(code, 0)
	if got < 0 {
		t.Fatalf("gs1FindAIIndex AI=99: not found (should match 9X wildcard)")
	}
	if gs1AICodes[got].numAI != "9X" {
		t.Errorf("gs1FindAIIndex AI=99: found %q, want 9X", gs1AICodes[got].numAI)
	}
}

// TestGS1FindAIIndex_AtOffset verifies FindAIIndex works at a non-zero offset,
// matching how C# uses it after parsing prior segments.
func TestGS1FindAIIndex_AtOffset(t *testing.T) {
	// The second AI starts at offset 18.
	code := "(01)12345678901231(10)LOT1"
	offset := 18 // position of '(' for AI 10
	got := gs1FindAIIndex(code, offset)
	if got < 0 {
		t.Fatalf("gs1FindAIIndex at offset 18: not found")
	}
	if gs1AICodes[got].numAI != "10" {
		t.Errorf("gs1FindAIIndex at offset 18: found %q, want 10", gs1AICodes[got].numAI)
	}
}

// ── gs1GetCode ─────────────────────────────────────────────────────────────────

// TestGS1GetCode_FixedLength_AI01 verifies the fixed-length path (useFNC1=false)
// for AI "01" with 14-digit data. C# GS1Helper.cs:174-178.
func TestGS1GetCode_FixedLength_AI01(t *testing.T) {
	code := "(01)12345678901231"
	seg, newIdx := gs1GetCode(code, 0)
	// Expected: AI digits "01" + data (14 chars) = "0112345678901231"
	want := "0112345678901231"
	if seg != want {
		t.Errorf("gs1GetCode AI=01: seg = %q, want %q", seg, want)
	}
	// newIdx should be at end of string (4 + 14 = 18)
	if newIdx != 18 {
		t.Errorf("gs1GetCode AI=01: newIdx = %d, want 18", newIdx)
	}
}

// TestGS1GetCode_VariableLength_AI10_WithNextAI verifies the variable-length path
// (useFNC1=true) for AI "10" followed by another AI.
// The function should detect the next '(' and append "&1;" separator.
// C# GS1Helper.cs:180-198.
func TestGS1GetCode_VariableLength_AI10_WithNextAI(t *testing.T) {
	code := "(10)LOT123(21)SERIAL456"
	seg, newIdx := gs1GetCode(code, 0)
	// AI "10" data = "LOT123" (6 chars); followed by '(' at offset 10.
	// maxLen = 10 - (index after ')') = 10 - 4 = 6.
	// 6 >= minDataLength(1) and <= maxDataLength(20) → valid.
	// result = "10" + "LOT123" + "&1;" (because maxLen < codeLen).
	want := "10LOT123&1;"
	if seg != want {
		t.Errorf("gs1GetCode AI=10 with next AI: seg = %q, want %q", seg, want)
	}
	// newIdx should be at position of '(' for next AI.
	if newIdx != 10 {
		t.Errorf("gs1GetCode AI=10 with next AI: newIdx = %d, want 10", newIdx)
	}
}

// TestGS1GetCode_VariableLength_AI10_Last verifies that the last AI in a chain
// does NOT append "&1;" when there is no following '('.
func TestGS1GetCode_VariableLength_AI10_Last(t *testing.T) {
	code := "(10)LOT123"
	seg, _ := gs1GetCode(code, 0)
	// No next paren → maxLen == codeLen → no FNC1 appended.
	want := "10LOT123"
	if seg != want {
		t.Errorf("gs1GetCode AI=10 last: seg = %q, want %q", seg, want)
	}
	if strings.Contains(seg, "&1;") {
		t.Errorf("gs1GetCode AI=10 last: should not contain FNC1 separator")
	}
}

// TestGS1GetCode_NoClosingParen verifies that missing ')' returns empty.
// C# GS1Helper.cs:168-169.
func TestGS1GetCode_NoClosingParen(t *testing.T) {
	code := "(01" // no closing paren
	seg, _ := gs1GetCode(code, 0)
	if seg != "" {
		t.Errorf("gs1GetCode no ')': seg = %q, want empty", seg)
	}
}

// TestGS1GetCode_DataTooShort verifies that data shorter than minDataLength returns empty.
// C# GS1Helper.cs:191 checks maxLen >= minDataLength.
func TestGS1GetCode_DataTooShort(t *testing.T) {
	// AI "00" requires exactly 18 digits. With only 5 digits, fixed-length path
	// requires codeLen >= maxDataLength (18). Since codeLen=5 < 18, returns empty.
	code := "(00)12345"
	seg, _ := gs1GetCode(code, 0)
	if seg != "" {
		t.Errorf("gs1GetCode data too short: seg = %q, want empty", seg)
	}
}

// TestGS1GetCode_AI21_VariableLengthExact verifies AI "21" with max-length data.
func TestGS1GetCode_AI21_VariableLengthExact(t *testing.T) {
	code := "(21)12345678901234567890" // 20 chars = maxDataLength
	seg, _ := gs1GetCode(code, 0)
	// maxLen = 20 == maxDataLength → no FNC1.
	want := "2112345678901234567890"
	if seg != want {
		t.Errorf("gs1GetCode AI=21 exact max: seg = %q, want %q", seg, want)
	}
}

// ── gs1ParseGS1 ───────────────────────────────────────────────────────────────

// TestGS1ParseGS1_TooShort verifies that input shorter than 3 chars returns "".
// C# GS1Helper.cs:207-208.
func TestGS1ParseGS1_TooShort(t *testing.T) {
	for _, s := range []string{"", "(", "(0"} {
		got := gs1ParseGS1(s)
		if got != "" {
			t.Errorf("gs1ParseGS1(%q) = %q, want empty", s, got)
		}
	}
}

// TestGS1ParseGS1_SingleAI_Fixed verifies a single fixed-length AI parses correctly.
// C# ParseGS1: prepends "&1;" and concatenates AI+data.
func TestGS1ParseGS1_SingleAI_Fixed(t *testing.T) {
	code := "(01)12345678901231"
	got := gs1ParseGS1(code)
	// C# output: "&1;" + "01" + "12345678901231"
	want := "&1;0112345678901231"
	if got != want {
		t.Errorf("gs1ParseGS1 single fixed AI: got %q, want %q", got, want)
	}
}

// TestGS1ParseGS1_TwoAIs_VariableLength verifies two variable-length AIs produce
// FNC1 separator between them and no FNC1 after the last.
func TestGS1ParseGS1_TwoAIs_VariableLength(t *testing.T) {
	code := "(10)LOT123(21)SERIAL456"
	got := gs1ParseGS1(code)
	// Expected: "&1;" + "10LOT123&1;" + "21SERIAL456"
	want := "&1;10LOT123&1;21SERIAL456"
	if got != want {
		t.Errorf("gs1ParseGS1 two variable AIs: got %q, want %q", got, want)
	}
}

// TestGS1ParseGS1_InvalidAI returns "" when a segment cannot be parsed.
// C# GS1Helper.cs:219-222: sets result="" and returns.
func TestGS1ParseGS1_InvalidAI(t *testing.T) {
	// "(99)ABC" — AI "99" matches "9X" wildcard, but data "ABC" has len=3
	// which is within 1..90, so this actually SUCCEEDS.
	// Use an AI with data that violates minDataLength instead:
	// AI "7001" requires exactly 13 digits. With only 3 chars, fails.
	code := "(7001)ABC"
	got := gs1ParseGS1(code)
	if got != "" {
		t.Errorf("gs1ParseGS1 invalid data length: got %q, want empty", got)
	}
}

// TestGS1ParseGS1_StartsWithFNC1 verifies that the output always starts with "&1;".
// C# ParseGS1 line 212: result = "&1;" before the loop.
func TestGS1ParseGS1_StartsWithFNC1(t *testing.T) {
	for _, code := range []string{
		"(01)12345678901231",
		"(10)LOTABC",
		"(21)SERIALXYZ",
	} {
		got := gs1ParseGS1(code)
		if got != "" && !strings.HasPrefix(got, "&1;") {
			t.Errorf("gs1ParseGS1(%q) = %q: should start with &1;", code, got)
		}
	}
}

// TestGS1ParseGS1_AI00_18Digits verifies SSCC (AI 00) with 18 digits.
func TestGS1ParseGS1_AI00_18Digits(t *testing.T) {
	code := "(00)123456789012345678"
	got := gs1ParseGS1(code)
	want := "&1;00123456789012345678"
	if got != want {
		t.Errorf("gs1ParseGS1 AI=00: got %q, want %q", got, want)
	}
}

// TestGS1ParseGS1_AI30_VariableLength verifies count AI with variable length.
func TestGS1ParseGS1_AI30_VariableLength(t *testing.T) {
	code := "(30)12345678"
	got := gs1ParseGS1(code)
	want := "&1;3012345678"
	if got != want {
		t.Errorf("gs1ParseGS1 AI=30: got %q, want %q", got, want)
	}
}

// TestGS1ParseGS1_MultipleFixedAndVariable verifies a chain of mixed AIs.
func TestGS1ParseGS1_MultipleFixedAndVariable(t *testing.T) {
	// AI 01 (fixed 14) + AI 10 (variable, ends at next '(') + AI 17 (fixed 6)
	code := "(01)12345678901231(10)LOT01(17)261231"
	got := gs1ParseGS1(code)
	// AI 01: "0112345678901231"
	// AI 10: "10LOT01&1;" (followed by another AI)
	// AI 17: "17261231"
	want := "&1;0112345678901231" + "10LOT01&1;" + "17261231"
	if got != want {
		t.Errorf("gs1ParseGS1 mixed: got %q, want %q", got, want)
	}
}
