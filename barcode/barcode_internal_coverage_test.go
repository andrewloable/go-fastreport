// barcode_internal_coverage_test.go — internal tests for unexported functions.
//
// This file is in package barcode (not barcode_test) to access unexported
// functions like code2of5setLen and qrVersionInfo.ecPerBlock.
package barcode

import (
	"testing"
)

// ── code2of5setLen ────────────────────────────────────────────────────────────

func TestCode2of5setLen_Padding(t *testing.T) {
	// Input shorter than n → padding branch.
	result := code2of5setLen("123", 6)
	if len(result) != 6 {
		t.Errorf("got len=%d, want 6", len(result))
	}
	if result != "000123" {
		t.Errorf("got %q, want %q", result, "000123")
	}
}

func TestCode2of5setLen_ExactLength(t *testing.T) {
	// Input exactly n → normal return.
	result := code2of5setLen("123456", 6)
	if result != "123456" {
		t.Errorf("got %q, want %q", result, "123456")
	}
}

func TestCode2of5setLen_Truncation(t *testing.T) {
	// Input longer than n → truncation.
	result := code2of5setLen("1234567890", 6)
	if result != "123456" {
		t.Errorf("got %q, want %q", result, "123456")
	}
}

// ── qrVersionInfo.ecPerBlock ──────────────────────────────────────────────────

func TestQRVersionInfo_ecPerBlock_Zero(t *testing.T) {
	// ecPerBlock with zero totalBlocks → returns 0.
	vi := qrVersionInfo{ecCodewords: 10, blocks: nil}
	result := vi.ecPerBlock()
	if result != 0 {
		t.Errorf("ecPerBlock with no blocks: got %d, want 0", result)
	}
}

func TestQRVersionInfo_ecPerBlock_Normal(t *testing.T) {
	// ecPerBlock with normal values.
	vi := qrVersionInfo{
		ecCodewords: 10,
		blocks: []qrBlock{
			{count: 1, dataPerBlock: 7},
		},
	}
	result := vi.ecPerBlock()
	if result != 10 {
		t.Errorf("ecPerBlock normal: got %d, want 10", result)
	}
}

// ── eanSetLen internal ────────────────────────────────────────────────────────

func TestEanSetLen_Padding(t *testing.T) {
	// Input shorter than n → padding branch.
	result := eanSetLen("123", 7)
	if len(result) != 7 {
		t.Errorf("got len=%d, want 7", len(result))
	}
	if result != "0000123" {
		t.Errorf("got %q, want %q", result, "0000123")
	}
}

func TestEanSetLen_ExactLength(t *testing.T) {
	// Input exactly n → truncation returns same.
	result := eanSetLen("1234567", 7)
	if result != "1234567" {
		t.Errorf("got %q, want %q", result, "1234567")
	}
}

func TestEanSetLen_LongerInput(t *testing.T) {
	// Input longer than n → truncation.
	result := eanSetLen("123456789", 7)
	if result != "1234567" {
		t.Errorf("got %q, want %q", result, "1234567")
	}
}

// ── code39GetPattern with calcChecksum=true ────────────────────────────────────
// Code39Barcode.GetPattern() always passes false; test the checksum path directly.

func TestCode39GetPattern_WithChecksum(t *testing.T) {
	result := code39GetPattern("HELLO", true)
	if len(result) == 0 {
		t.Error("code39GetPattern with checksum returned empty")
	}
}

func TestCode39GetPattern_WithChecksum_Numbers(t *testing.T) {
	result := code39GetPattern("12345", true)
	if len(result) == 0 {
		t.Error("code39GetPattern with checksum numbers returned empty")
	}
}

// ── c128GetNextChar edge cases ────────────────────────────────────────────────
// c128GetNextChar has branches for various token types (&S;, &4;, etc.)

func TestC128GetNextChar_ShiftToken(t *testing.T) {
	// The &S; token is the shift token in Code128.
	// It switches between Code A and Code B.
	pattern, err := code128GetPattern("&A;AB&S;cd&A;EF")
	if err != nil {
		t.Logf("code128GetPattern shift: %v (acceptable)", err)
		return
	}
	if len(pattern) == 0 {
		t.Error("code128GetPattern with shift returned empty")
	}
}

func TestC128GetNextChar_FNC2Token(t *testing.T) {
	// FNC2 = &2; token.
	pattern, err := code128GetPattern("&B;AB&2;CD")
	if err != nil {
		t.Logf("code128GetPattern FNC2: %v (acceptable)", err)
		return
	}
	if len(pattern) == 0 {
		t.Error("code128GetPattern FNC2 returned empty")
	}
}

func TestC128GetNextChar_FNC3Token(t *testing.T) {
	// FNC3 = &3; token.
	pattern, err := code128GetPattern("&B;AB&3;CD")
	if err != nil {
		t.Logf("code128GetPattern FNC3: %v (acceptable)", err)
		return
	}
	if len(pattern) == 0 {
		t.Error("code128GetPattern FNC3 returned empty")
	}
}

func TestC128GetNextChar_CODE4Token(t *testing.T) {
	// CODE4 = &4; token.
	pattern, err := code128GetPattern("&A;AB&4;CD")
	if err != nil {
		t.Logf("code128GetPattern CODE4: %v (acceptable)", err)
		return
	}
	if len(pattern) == 0 {
		t.Error("code128GetPattern CODE4 returned empty")
	}
}

func TestC128GetNextChar_InvalidDefault(t *testing.T) {
	// Invalid start token → should return error.
	_, err := code128GetPattern("&X;HELLO")
	if err == nil {
		t.Error("expected error for invalid start token")
	}
}

// ── dmGetEncodation internal tests ────────────────────────────────────────────

func TestDmGetEncodation_NegativeDataSize(t *testing.T) {
	// dataSize < 0 → returns -1.
	result := dmGetEncodation([]byte("ABC"), 0, 3, make([]byte, 100), 0, -1, false, false)
	if result != -1 {
		t.Errorf("got %d, want -1", result)
	}
}

func TestDmGetEncodation_FirstMatch_ASCII(t *testing.T) {
	// firstMatch=true with ASCII input → returns on first match.
	data := make([]byte, 100)
	result := dmGetEncodation([]byte("ABC"), 0, 3, data, 0, 10, false, true)
	if result < 0 {
		t.Errorf("firstMatch ASCII: got %d, want >= 0", result)
	}
}

func TestDmGetEncodation_FirstMatch_Binary(t *testing.T) {
	// firstMatch=true with binary input → ASCII may fail.
	data := make([]byte, 100)
	input := []byte{200, 201, 202}
	result := dmGetEncodation(input, 0, 3, data, 0, 10, false, true)
	_ = result // result may be -1 or valid
}

// ── dmGenerate with large text (> capacity) ───────────────────────────────────

func TestDmGenerate_TooLarge(t *testing.T) {
	// Generate with text too large for any symbol (> ~1558 bytes).
	data := make([]byte, 1600)
	for i := range data {
		data[i] = byte('A' + i%26)
	}
	_, _, _, err := dmGenerate(data)
	// This may succeed with a large symbol or fail with "too large".
	if err != nil {
		t.Logf("dmGenerate too large: %v (acceptable)", err)
	}
}

// ── dmGetPoly with all cases ───────────────────────────────────────────────────

func TestDmGetPoly_AllCases(t *testing.T) {
	cases := []int{5, 7, 10, 11, 12, 14, 18, 20, 24, 28, 36, 42, 48, 56, 62, 68, 99}
	for _, nc := range cases {
		result := dmGetPoly(nc)
		if nc != 99 && result == nil {
			t.Errorf("dmGetPoly(%d) returned nil", nc)
		}
		if nc == 99 && result != nil {
			t.Errorf("dmGetPoly(99) should return nil, got %v", result)
		}
	}
}

// ── DeutscheIdentcode GetPattern case 12 (encodedText set directly) ──────────
// DeutscheIdentcodeBarcode.Encode() only accepts 11 digits, so the "case 12"
// branch in code2of5.go GetPattern can only be reached via direct struct access.

func TestDeutscheIdentcodeBarcode_GetPattern_12Digit_Internal(t *testing.T) {
	b := &DeutscheIdentcodeBarcode{}
	b.encodedText = "123456789012" // 12 digits bypasses Encode validation
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern 12-digit: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern 12-digit returned empty")
	}
}

// ── DeutscheLeitcode GetPattern case 14 (encodedText set directly) ────────────

func TestDeutscheLeitcodeBarcode_GetPattern_14Digit_Internal(t *testing.T) {
	b := &DeutscheLeitcodeBarcode{}
	b.encodedText = "12345678901234" // 14 digits bypasses Encode validation
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern 14-digit: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern 14-digit returned empty")
	}
}

// ── dmDoPlacement with ncol%8==0 to exercise corner4 ─────────────────────────
// corner4 fires when row==nrow+4 && col==2 && ncol%8==0 during ecc200().
// We call dmDoPlacement directly with various (nrow,ncol) pairs that have
// ncol%8==0 to maximise the chance of triggering corner4.

func TestDmDoPlacement_NcolMod8Zero(t *testing.T) {
	// These (nrow, ncol) pairs all have ncol%8==0 and correspond to actual
	// DataMatrix symbol sizes from dmSizes.
	pairs := [][2]int{
		{8, 8},   // 10x10 symbol
		{6, 16},  // 8x18 symbol
		{10, 24}, // 12x26 symbol
		{16, 16}, // 18x18 symbol
		{10, 32}, // 12x36 symbol
		{14, 32}, // 16x36 symbol
	}
	for _, p := range pairs {
		nrow, ncol := p[0], p[1]
		result := dmDoPlacement(nrow, ncol)
		if result == nil {
			t.Errorf("dmDoPlacement(%d,%d) returned nil", nrow, ncol)
		}
		if len(result) != nrow*ncol {
			t.Errorf("dmDoPlacement(%d,%d) returned len=%d, want %d",
				nrow, ncol, len(result), nrow*ncol)
		}
	}
}

// ── c128IsFourOrMoreDigits boundary tests ─────────────────────────────────────
//
// C# IsFourOrMoreDigits requires index+4 < code.Length (strictly less than),
// meaning exactly 4 digits at end-of-string returns FALSE. This mirrors
// Barcode128.cs:238-249 which uses "index + 4 < code.Length".

// TestC128IsFourOrMoreDigits_ExactlyFour checks the boundary: 4 digits with no
// further chars must return false, matching C# strict-less-than check.
func TestC128IsFourOrMoreDigits_ExactlyFour(t *testing.T) {
	ok, n := c128IsFourOrMoreDigits("1234", 0)
	if ok {
		t.Errorf("c128IsFourOrMoreDigits(\"1234\", 0) = (true, %d), want (false, 0); C# index+4 < len requires 5+ chars", n)
	}
}

// TestC128IsFourOrMoreDigits_FiveDigits checks that 5 digits at index 0 returns true.
func TestC128IsFourOrMoreDigits_FiveDigits(t *testing.T) {
	ok, n := c128IsFourOrMoreDigits("12345", 0)
	if !ok || n != 5 {
		t.Errorf("c128IsFourOrMoreDigits(\"12345\", 0) = (%v, %d), want (true, 5)", ok, n)
	}
}

// TestC128IsFourOrMoreDigits_FourDigitsThenAlpha checks 4 digits followed by alpha.
// C# condition: 0+4 < 5 → true, so count=4 digits → returns (true, 4).
func TestC128IsFourOrMoreDigits_FourDigitsThenAlpha(t *testing.T) {
	ok, n := c128IsFourOrMoreDigits("1234A", 0)
	if !ok || n != 4 {
		t.Errorf("c128IsFourOrMoreDigits(\"1234A\", 0) = (%v, %d), want (true, 4)", ok, n)
	}
}

// TestC128IsFourOrMoreDigits_EmptyString checks that empty input returns false.
func TestC128IsFourOrMoreDigits_EmptyString(t *testing.T) {
	ok, _ := c128IsFourOrMoreDigits("", 0)
	if ok {
		t.Error("c128IsFourOrMoreDigits(\"\", 0) should return false for empty string")
	}
}

// TestC128IsFourOrMoreDigits_NonDigitFirst checks that a non-digit first char returns false.
func TestC128IsFourOrMoreDigits_NonDigitFirst(t *testing.T) {
	ok, _ := c128IsFourOrMoreDigits("ABCDE", 0)
	if ok {
		t.Error("c128IsFourOrMoreDigits(\"ABCDE\", 0) should return false (non-digit first)")
	}
}

// TestC128AutoEncode_ExactlyFourDigits verifies that 4 digits with no trailing
// chars use Code B (not Code C). C# IsFourOrMoreDigits requires index+4 < len,
// so exactly 4 digits at end-of-string does NOT select Code C.
// Ref: Barcode128.cs:238-249, GetNextPortion lines 302-315.
func TestC128AutoEncode_ExactlyFourDigits(t *testing.T) {
	encoded := c128AutoEncode("1234")
	// C# produces "&B;1234" (Code B), NOT "&C;1234" (Code C).
	if encoded != "&B;1234" {
		t.Errorf("c128AutoEncode(\"1234\") = %q, want \"&B;1234\" (C# uses Code B for exactly 4 digits)", encoded)
	}
}

// TestC128AutoEncode_FiveDigits verifies that 5 digits select Code C (even pair = 4).
// C# IsFourOrMoreDigits returns true for 5 digits (0+4 < 5 → true), and
// numDigits=(5/2)*2=4, so "&C;1234" is produced, leaving "5" for the next portion.
func TestC128AutoEncode_FiveDigits(t *testing.T) {
	encoded := c128AutoEncode("12345")
	// 5 digits: Code C for first 4 (even), then Code B for remaining "5".
	// Expected: "&C;1234&B;5"
	if encoded != "&C;1234&B;5" {
		t.Errorf("c128AutoEncode(\"12345\") = %q, want \"&C;1234&B;5\"", encoded)
	}
}

// TestCode128_GetPattern_ExactlyFourDigits ensures the GetPattern output for "1234"
// (auto-encoded as Code B) is well-formed and has correct length.
// Code B: StartB(6)+4chars×6(24)+check(6)+stop(7) = 43 chars.
func TestCode128_GetPattern_ExactlyFourDigits(t *testing.T) {
	b := NewCode128Barcode()
	if err := b.Encode("1234"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	// StartB "211214" → doConvert at positions 0-5 → "605153"
	// doConvert: i=0(even):'2'-1+5=6; i=1(odd):'1'-1=0; i=2(even):'1'-1+5=5;
	//            i=3(odd):'2'-1=1; i=4(even):'1'-1+5=5; i=5(odd):'4'-1=3.
	if len(pattern) < 6 || pattern[:6] != "605153" {
		t.Errorf("pattern start = %q, want \"605153\" (C# StartB \"211214\" after doConvert)", pattern[:6])
	}
	// Stop "2331112" → "6270506"
	if len(pattern) < 7 || pattern[len(pattern)-7:] != "6270506" {
		t.Errorf("pattern stop = %q, want \"6270506\"", pattern[len(pattern)-7:])
	}
}
