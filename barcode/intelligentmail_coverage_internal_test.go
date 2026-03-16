package barcode

// intelligentmail_coverage_internal_test.go — internal package tests to cover
// the remaining reachable branches in intelligentmail.go.
//
// Uncovered blocks (from cover profile):
//   - Line 150.22,155.3: imbMathAdd loop body — dead code (l = x|65535, never == 1)
//   - Line 270.17,272.4: imb_encode bad zip parse error — dead code (zip already validated)
//   - Line 355.25,357.4: imb_encode codeword out-of-range — dead code (always in range)
//   - Line 421.3,421.12: Render strings.Map non-digit filter returns -1 — REACHABLE
//   - Line 457.17,459.4: Render x1 clamping — potentially reachable with specific widths
//   - Line 475.11,477.15: Render default bar-type case — dead code (bars always 0-3)
//
// This file covers the reachable blocks via internal (package-level) access.

import (
	"testing"
)

// TestIntelligentMailBarcode_Render_NonDigitEncodedText exercises the
// strings.Map non-digit return path at line 421. By setting encodedText
// directly (bypassing Encode validation), the Map function encounters
// non-digit characters and returns -1 for each, exercising that branch.
// The remaining 20 digits in "01-234-567-09-4987654321" are valid, so
// imb_encode succeeds after stripping the dashes.
func TestIntelligentMailBarcode_Render_NonDigitEncodedText(t *testing.T) {
	b := NewIntelligentMailBarcode()
	// Set encodedText to a 22-char string with dashes: after strings.Map
	// filters non-digits, the result is the 20-digit "01234567094987654321".
	// This exercises the `return -1` branch in the Map function at line 421.
	b.encodedText = "01-234-567-09-4987654321"
	img, err := b.Render(130, 60)
	if err != nil {
		t.Fatalf("Render with dashes in encodedText: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image for formatted encodedText")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 130 || bounds.Dy() != 60 {
		t.Errorf("expected 130x60, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestIntelligentMailBarcode_Render_NonDigitEncodedText_WithSpaces is a variant
// that uses spaces and dots in encodedText to further exercise the -1 return path.
func TestIntelligentMailBarcode_Render_NonDigitEncodedText_WithSpaces(t *testing.T) {
	b := NewIntelligentMailBarcode()
	// "01 234 567094987654321" — spaces stripped by Map → 20 valid digits.
	b.encodedText = "01 234 567094987654321"
	img, err := b.Render(200, 60)
	if err != nil {
		t.Fatalf("Render with spaces in encodedText: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image for encodedText with spaces")
	}
}

// TestIntelligentMailBarcode_Render_X1Clamping attempts to trigger the
// `if x1 > width { x1 = width }` guard at line 457 by using widths
// where floating-point rounding of 65*barW could exceed the given width.
// We try a range of widths that are not clean multiples of 65 to maximise
// the chance of hitting the IEEE 754 boundary condition.
func TestIntelligentMailBarcode_Render_X1Clamping(t *testing.T) {
	b := NewIntelligentMailBarcode()
	// "00000000000000000000" — all-zeros, valid 20-digit IMb.
	if err := b.Encode("00000000000000000000"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// Try widths that are not multiples of 65 — the last bar's x1 may round
	// to width+1 due to floating-point arithmetic.
	widths := []int{1, 2, 3, 4, 7, 13, 64, 66, 100, 130, 131, 200, 259, 260, 261}
	for _, w := range widths {
		img, err := b.Render(w, 60)
		if err != nil {
			t.Fatalf("Render(width=%d): %v", w, err)
		}
		if img == nil {
			t.Fatalf("Render(width=%d) returned nil", w)
		}
		// Verify the image has the requested width.
		got := img.Bounds().Dx()
		if got != w {
			t.Errorf("Render(width=%d): image width=%d", w, got)
		}
	}
}

// TestImbMathAdd_AllReachableStatements exercises all statements in imbMathAdd
// that can actually execute (the for-body at line 151-155 is dead code:
// l = x|65535 ensures l >= 65535, so l==1 is never true).
// We call imbMathAdd with representative inputs to maximise coverage of the
// 5 reachable pre-loop statements.
func TestImbMathAdd_AllReachableStatements(t *testing.T) {
	tests := []struct {
		name    string
		arr     []int
		j       int
		want12  int
		want11  int
	}{
		// j = 0: x = arr[12] | (arr[11]<<8) + 0; no change
		{"zero-j", makeArr(0x42, 0x01), 0, 0x42, 0x01},
		// j = 1, no carry into byte 11
		{"small-j", makeArr(5, 0), 3, 8, 0},
		// j produces carry from byte 12 into byte 11
		{"cross-byte-carry", makeArr(0xFF, 0x00), 1, 0x00, 0x01},
		// Large j that fills multiple bytes
		{"large-j", makeArr(0x00, 0x00), 0xFFFF, 0xFF, 0xFF},
		// Already large arr[11] value
		{"large-arr11", makeArr(0x00, 0xFE), 0x0100, 0x00, 0xFF},
		// All zeros
		{"all-zeros", makeArr(0, 0), 0, 0, 0},
		// Max byte values
		{"max-bytes", makeArr(0xFF, 0xFF), 0, 0xFF, 0xFF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arr := tt.arr
			imbMathAdd(arr, tt.j)
			if arr[12] != tt.want12 {
				t.Errorf("arr[12] = 0x%02x, want 0x%02x", arr[12], tt.want12)
			}
			if arr[11] != tt.want11 {
				t.Errorf("arr[11] = 0x%02x, want 0x%02x", arr[11], tt.want11)
			}
		})
	}
}

// makeArr creates a 13-element int slice with arr[12]=lo and arr[11]=hi.
func makeArr(lo, hi int) []int {
	a := make([]int, 13)
	a[12] = lo
	a[11] = hi
	return a
}

// TestImbEncode_HighFCSPath exercises the `fcs >> 10 != 0` branch (line 344)
// by using an input that produces fcs >= 1024. The input "00000000000000000000"
// produces fcs=1760 (bit 10 set), confirmed by discovery testing.
func TestImbEncode_HighFCSPath(t *testing.T) {
	bars, err := imb_encode("00000000000000000000")
	if err != nil {
		t.Fatalf("imb_encode (high FCS): %v", err)
	}
	if len(bars) != 65 {
		t.Errorf("expected 65 bars, got %d", len(bars))
	}
}

// TestImbEncode_LowFCSPath exercises the `fcs >> 10 == 0` branch (line 344)
// by using an input that produces fcs < 1024. "01234567094987654321" → fcs=81.
func TestImbEncode_LowFCSPath(t *testing.T) {
	bars, err := imb_encode("01234567094987654321")
	if err != nil {
		t.Fatalf("imb_encode (low FCS): %v", err)
	}
	if len(bars) != 65 {
		t.Errorf("expected 65 bars, got %d", len(bars))
	}
}

// TestImbEncode_AllFourZipLengths exercises all three zip-length branches plus
// the no-zip case (len(zip)==0) in imb_encode.
func TestImbEncode_AllFourZipLengths(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{"no-zip (20)", "01234567094987654321"},
		{"zip-5 (25)", "0123456709498765432190210"},
		{"zip-9 (29)", "01234567094987654321902101234"},
		{"zip-11 (31)", "0123456709498765432112345678901"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bars, err := imb_encode(tc.text)
			if err != nil {
				t.Fatalf("imb_encode(%s): %v", tc.name, err)
			}
			if len(bars) != 65 {
				t.Errorf("expected 65 bars, got %d", len(bars))
			}
		})
	}
}

// TestIntelligentMailBarcode_Render_NotEncodedReturnsError verifies that calling
// Render without calling Encode returns an error (line 413 in Render).
func TestIntelligentMailBarcode_Render_NotEncodedReturnsError(t *testing.T) {
	b := NewIntelligentMailBarcode()
	// encodedText is "" — Render should return a non-nil error.
	_, err := b.Render(130, 60)
	if err == nil {
		t.Error("Render without Encode: expected error, got nil")
	}
}
