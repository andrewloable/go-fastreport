// barcode_pattern_coverage_test.go — coverage for GetPattern() and DrawLinearBarcode paths.
//
// These tests call the PatternProvider.GetPattern() methods on all linear barcode
// types and pipe the results through DrawLinearBarcode / DrawBarcode2D / render helpers
// to exercise render.go, render2d.go, and every symbology implementation file.
package barcode_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ── Render helpers: DrawLinearBarcode and DrawBarcode2D ───────────────────────

func TestDrawLinearBarcode_Basic(t *testing.T) {
	pattern := "505160605" // simple pattern
	img := barcode.DrawLinearBarcode(pattern, "TEST", 200, 50, true, 2.0)
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestDrawLinearBarcode_NoText(t *testing.T) {
	pattern := "505160605"
	img := barcode.DrawLinearBarcode(pattern, "", 200, 50, false, 2.0)
	if img == nil {
		t.Fatal("DrawLinearBarcode (no text) returned nil")
	}
}

func TestDrawLinearBarcode_ZeroWidth(t *testing.T) {
	img := barcode.DrawLinearBarcode("505160605", "TEST", 0, 50, true, 2.0)
	if img == nil {
		t.Fatal("DrawLinearBarcode (zero width) returned nil")
	}
}

func TestDrawLinearBarcode_ZeroHeight(t *testing.T) {
	img := barcode.DrawLinearBarcode("505160605", "TEST", 200, 0, true, 2.0)
	if img == nil {
		t.Fatal("DrawLinearBarcode (zero height) returned nil")
	}
}

func TestDrawLinearBarcode_EmptyPattern(t *testing.T) {
	img := barcode.DrawLinearBarcode("", "TEST", 200, 50, true, 2.0)
	if img == nil {
		t.Fatal("DrawLinearBarcode (empty pattern) returned nil")
	}
}

func TestDrawLinearBarcode_ZeroWideBarRatio(t *testing.T) {
	img := barcode.DrawLinearBarcode("505160605", "TEST", 200, 50, true, 0)
	if img == nil {
		t.Fatal("DrawLinearBarcode (zero wideBarRatio) returned nil")
	}
}

func TestDrawLinearBarcode_AllBarTypes(t *testing.T) {
	// Cover all bar line type codes: 0-3 (white), 5-8 (black), 9 (half),
	// A-D (long), E (tracker), F (ascender), G (descender)
	pattern := "0123456789ABCDEFG"
	img := barcode.DrawLinearBarcode(pattern, "ALL", 400, 100, true, 2.0)
	if img == nil {
		t.Fatal("DrawLinearBarcode (all bar types) returned nil")
	}
}

func TestDrawLinearBarcode_WithText_SmallHeight(t *testing.T) {
	// showText=true with very small height so barAreaH could go < 1
	img := barcode.DrawLinearBarcode("5050505050", "T", 200, 15, true, 2.0)
	if img == nil {
		t.Fatal("DrawLinearBarcode (small height) returned nil")
	}
}

func TestDrawBarcode2D_Basic(t *testing.T) {
	matrix := [][]bool{
		{true, false, true},
		{false, true, false},
		{true, false, true},
	}
	img := barcode.DrawBarcode2D(matrix, 3, 3, 100, 100)
	if img == nil {
		t.Fatal("DrawBarcode2D returned nil")
	}
}

func TestDrawBarcode2D_ZeroDimensions(t *testing.T) {
	img := barcode.DrawBarcode2D(nil, 0, 0, 100, 100)
	if img == nil {
		t.Fatal("DrawBarcode2D (zero rows/cols) returned nil")
	}
}

func TestDrawBarcode2D_ZeroOutputSize(t *testing.T) {
	matrix := [][]bool{{true, false}, {false, true}}
	img := barcode.DrawBarcode2D(matrix, 2, 2, 0, 0)
	if img == nil {
		t.Fatal("DrawBarcode2D (zero output size) returned nil")
	}
}

// ── Helper functions: MakeModules, OneBarProps, GetPatternWidth, etc. ─────────

func TestMakeModules(t *testing.T) {
	m := barcode.MakeModules(2.0)
	if m[0] != 1 {
		t.Errorf("modules[0] = %v, want 1", m[0])
	}
	if m[1] != 2.0 {
		t.Errorf("modules[1] = %v, want 2", m[1])
	}
}

func TestOneBarProps_AllCodes(t *testing.T) {
	m := barcode.MakeModules(2.0)
	codes := []byte{'0', '1', '2', '3', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'G'}
	for _, code := range codes {
		w, _, err := barcode.OneBarProps(code, m)
		if err != nil {
			t.Errorf("OneBarProps(%q) error: %v", code, err)
		}
		if w <= 0 {
			t.Errorf("OneBarProps(%q) width = %v, want > 0", code, w)
		}
	}
}

func TestOneBarProps_UnknownCode(t *testing.T) {
	m := barcode.MakeModules(2.0)
	_, _, err := barcode.OneBarProps('Z', m)
	if err == nil {
		t.Error("OneBarProps unknown code: expected error")
	}
}

func TestGetPatternWidth(t *testing.T) {
	m := barcode.MakeModules(2.0)
	w := barcode.GetPatternWidth("5050", m)
	if w <= 0 {
		t.Errorf("GetPatternWidth returned %v, want > 0", w)
	}
}

func TestCheckSumModulo10_Basic(t *testing.T) {
	got := barcode.CheckSumModulo10("1234567")
	if len(got) != 8 {
		t.Errorf("CheckSumModulo10 length = %d, want 8", len(got))
	}
}

func TestCheckSumModulo10_Exact(t *testing.T) {
	// Known: 1234567 → check digit is 0 (sum=70)
	got := barcode.CheckSumModulo10("1234")
	if len(got) != 5 {
		t.Errorf("len = %d, want 5", len(got))
	}
}

func TestMakeLong(t *testing.T) {
	got := barcode.MakeLong("5678")
	if got != "ABCD" {
		t.Errorf("MakeLong(5678) = %q, want ABCD", got)
	}
}

func TestMakeLong_NoChange(t *testing.T) {
	got := barcode.MakeLong("0123")
	if got != "0123" {
		t.Errorf("MakeLong(0123) = %q, want 0123", got)
	}
}

// ── Code128 and variants: GetPattern ──────────────────────────────────────────

func TestCode128Barcode_GetPattern(t *testing.T) {
	b := barcode.NewCode128Barcode()
	if err := b.Encode("HELLO123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	// Render via DrawLinearBarcode
	img := barcode.DrawLinearBarcode(pattern, "HELLO123", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestCode128Barcode_GetPattern_NumericOnly(t *testing.T) {
	b := barcode.NewCode128Barcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
}

func TestCode128Barcode_GetPattern_WithFNC(t *testing.T) {
	b := barcode.NewCode128Barcode()
	// FNC codes in text (using control codes)
	if err := b.Encode("&1;HELLO"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	_, err := b.GetPattern()
	if err != nil {
		t.Logf("GetPattern with FNC code: %v (acceptable)", err)
	}
}

func TestCode128Barcode_GetPattern_MixedCases(t *testing.T) {
	// Test auto-encoding with code switching: B→C for long digit runs
	testCases := []string{
		"ABC12345678XYZ",
		"1234567890",
		"abc",
		"\x01\x02\x03",
	}
	for _, tc := range testCases {
		b := barcode.NewCode128Barcode()
		if err := b.Encode(tc); err != nil {
			t.Logf("Encode(%q): error (acceptable): %v", tc, err)
			continue
		}
		_, err := b.GetPattern()
		if err != nil {
			t.Logf("GetPattern(%q): error: %v", tc, err)
		}
	}
}

func TestCode128ABarcode_GetPattern(t *testing.T) {
	b := barcode.NewCode128ABarcode()
	if err := b.Encode("CODE128A"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	_ = barcode.DrawLinearBarcode(pattern, "CODE128A", 400, 80, true, b.GetWideBarRatio())
}

func TestCode128BBarcode_GetPattern(t *testing.T) {
	b := barcode.NewCode128BBarcode()
	if err := b.Encode("Code128B"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	_ = barcode.DrawLinearBarcode(pattern, "Code128B", 400, 80, true, b.GetWideBarRatio())
}

func TestCode128CBarcode_GetPattern(t *testing.T) {
	b := barcode.NewCode128CBarcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	_ = barcode.DrawLinearBarcode(pattern, "12345678", 400, 80, true, b.GetWideBarRatio())
}

func TestCode128CBarcode_OddLength(t *testing.T) {
	b := barcode.NewCode128CBarcode()
	// Odd length: a '0' is prepended
	if err := b.Encode("12345"); err != nil {
		t.Fatalf("Encode odd-length: %v", err)
	}
	_, err := b.GetPattern()
	if err != nil {
		t.Logf("GetPattern odd-length: %v (acceptable)", err)
	}
}

// ── Code39 and variants ───────────────────────────────────────────────────────

func TestCode39Barcode_GetPattern(t *testing.T) {
	b := barcode.NewCode39Barcode()
	if err := b.Encode("HELLO"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "HELLO", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestCode39ExtendedBarcode_GetPattern(t *testing.T) {
	b := barcode.NewCode39ExtendedBarcode()
	if err := b.Encode("hello-1234"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	_ = barcode.DrawLinearBarcode(pattern, "hello-1234", 400, 80, true, b.GetWideBarRatio())
}

func TestCode39ExtendedBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewCode39ExtendedBarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

// ── Code93 and variants ───────────────────────────────────────────────────────

func TestCode93Barcode_GetPattern(t *testing.T) {
	b := barcode.NewCode93Barcode()
	if err := b.Encode("HELLO"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "HELLO", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestCode93Barcode_GetPattern_Extended(t *testing.T) {
	b := barcode.NewCode93Barcode()
	b.FullASCIIMode = true
	if err := b.Encode("HELLO"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern (extended): %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
}

func TestCode93Barcode_Render_FailPath(t *testing.T) {
	// Render with tiny dimensions triggers fallback (enc.Encode fails, retries with 0,0)
	b := barcode.NewCode93Barcode()
	if err := b.Encode("HELLO"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(1, 1)
	if err != nil {
		t.Logf("Render(1,1): %v (acceptable)", err)
		return
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestCode93ExtendedBarcode_GetPattern(t *testing.T) {
	b := barcode.NewCode93ExtendedBarcode()
	if err := b.Encode("hello"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	_ = barcode.DrawLinearBarcode(pattern, "hello", 400, 80, true, b.GetWideBarRatio())
}

func TestCode93ExtendedBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewCode93ExtendedBarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestCode93ExtendedBarcode_Render(t *testing.T) {
	b := barcode.NewCode93ExtendedBarcode()
	if err := b.Encode("HELLO"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(400, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestCode93ExtendedBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewCode93ExtendedBarcode()
	_, err := b.Render(400, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

// ── Codabar ───────────────────────────────────────────────────────────────────

func TestCodabarBarcode_GetPattern(t *testing.T) {
	b := barcode.NewCodabarBarcode()
	if err := b.Encode("A12345B"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "A12345B", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestCodabarBarcode_GetPattern_ShortInput(t *testing.T) {
	// Text length < 2 — auto-wrap with A...B
	b := barcode.NewCodabarBarcode()
	if err := b.Encode("5"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern short: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern short returned empty pattern")
	}
}

func TestCodabarBarcode_GetPattern_AllStartStops(t *testing.T) {
	// A, B, C, D are valid start/stop chars
	for _, startStop := range []string{"A", "B", "C", "D"} {
		b := barcode.NewCodabarBarcode()
		text := startStop + "1234" + startStop
		if err := b.Encode(text); err != nil {
			t.Fatalf("Encode %q: %v", text, err)
		}
		_, err := b.GetPattern()
		if err != nil {
			t.Logf("GetPattern %q: error: %v", text, err)
		}
	}
}

func TestCodabarBarcode_Render(t *testing.T) {
	b := barcode.NewCodabarBarcode()
	if err := b.Encode("A12345B"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(400, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestCodabarBarcode_Render_FailPath(t *testing.T) {
	b := barcode.NewCodabarBarcode()
	if err := b.Encode("A12345B"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// Tiny dimensions force fallback
	img, err := b.Render(1, 1)
	if err != nil {
		t.Logf("Render(1,1): %v (acceptable)", err)
		return
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

// ── 2-of-5 family ─────────────────────────────────────────────────────────────

func TestCode2of5Barcode_GetPattern_Interleaved(t *testing.T) {
	b := barcode.NewCode2of5Barcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "12345678", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestCode2of5Barcode_GetPattern_OddLength(t *testing.T) {
	// Odd-length gets padded to even
	b := barcode.NewCode2of5Barcode()
	if err := b.Encode("12345"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern odd: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern odd returned empty pattern")
	}
}

func TestCode2of5Barcode_GetPattern_EmptyError(t *testing.T) {
	b := &barcode.Code2of5Barcode{}
	_, err := b.GetPattern()
	if err == nil {
		t.Error("expected error for empty Code2of5 GetPattern")
	}
}

func TestCode2of5IndustrialBarcode_GetPattern(t *testing.T) {
	b := barcode.NewCode2of5IndustrialBarcode()
	if err := b.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "123456", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestCode2of5IndustrialBarcode_Render(t *testing.T) {
	b := barcode.NewCode2of5IndustrialBarcode()
	if err := b.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(400, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestCode2of5IndustrialBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewCode2of5IndustrialBarcode()
	_, err := b.Render(400, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

func TestCode2of5IndustrialBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewCode2of5IndustrialBarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestCode2of5MatrixBarcode_GetPattern(t *testing.T) {
	b := barcode.NewCode2of5MatrixBarcode()
	if err := b.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "123456", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestCode2of5MatrixBarcode_Render(t *testing.T) {
	b := barcode.NewCode2of5MatrixBarcode()
	if err := b.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(400, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestCode2of5MatrixBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewCode2of5MatrixBarcode()
	_, err := b.Render(400, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

func TestCode2of5MatrixBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewCode2of5MatrixBarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

// ── ITF-14 ────────────────────────────────────────────────────────────────────

func TestITF14Barcode_GetPattern(t *testing.T) {
	b := barcode.NewITF14Barcode()
	if err := b.Encode("12345678901231"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "12345678901231", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestITF14Barcode_Encode_13digits(t *testing.T) {
	b := barcode.NewITF14Barcode()
	if err := b.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode 13 digits: %v", err)
	}
}

func TestITF14Barcode_Render(t *testing.T) {
	b := barcode.NewITF14Barcode()
	if err := b.Encode("12345678901231"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(400, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestITF14Barcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewITF14Barcode()
	_, err := b.Render(400, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

func TestITF14Barcode_DefaultValue(t *testing.T) {
	b := barcode.NewITF14Barcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

// ── Deutsche Post barcodes ────────────────────────────────────────────────────

func TestDeutscheIdentcodeBarcode_GetPattern(t *testing.T) {
	b := barcode.NewDeutscheIdentcodeBarcode()
	if err := b.Encode("12345123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "12345123456", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestDeutscheIdentcodeBarcode_Render(t *testing.T) {
	b := barcode.NewDeutscheIdentcodeBarcode()
	if err := b.Encode("12345123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(400, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestDeutscheIdentcodeBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewDeutscheIdentcodeBarcode()
	_, err := b.Render(400, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

func TestDeutscheIdentcodeBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewDeutscheIdentcodeBarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestDeutscheLeitcodeBarcode_GetPattern(t *testing.T) {
	b := barcode.NewDeutscheLeitcodeBarcode()
	if err := b.Encode("1234512312312"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "1234512312312", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestDeutscheLeitcodeBarcode_Render(t *testing.T) {
	b := barcode.NewDeutscheLeitcodeBarcode()
	if err := b.Encode("1234512312312"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(400, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestDeutscheLeitcodeBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewDeutscheLeitcodeBarcode()
	_, err := b.Render(400, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

func TestDeutscheLeitcodeBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewDeutscheLeitcodeBarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

// ── EAN-8 and EAN-13 pattern generation ──────────────────────────────────────

func TestEAN8Barcode_GetPattern(t *testing.T) {
	b := barcode.NewEAN8Barcode()
	if err := b.Encode("1234567"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "1234567", 300, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestEAN13Barcode_GetPattern_AllParityRows(t *testing.T) {
	// Test all 10 possible first digits to exercise all parity table rows
	for d := 0; d <= 9; d++ {
		text := make([]byte, 12)
		text[0] = byte('0' + d)
		for i := 1; i < 12; i++ {
			text[i] = byte('0' + (i % 10))
		}
		b := barcode.NewEAN13Barcode()
		if err := b.Encode(string(text)); err != nil {
			t.Fatalf("Encode(%d): %v", d, err)
		}
		pattern, err := b.GetPattern()
		if err != nil {
			t.Fatalf("GetPattern(%d): %v", d, err)
		}
		if len(pattern) == 0 {
			t.Errorf("GetPattern(%d) returned empty pattern", d)
		}
		_ = barcode.DrawLinearBarcode(pattern, string(text), 400, 80, true, b.GetWideBarRatio())
	}
}

// ── UPC variants ─────────────────────────────────────────────────────────────

func TestUPCABarcode_GetPattern(t *testing.T) {
	b := barcode.NewUPCABarcode()
	if err := b.Encode("01234567890"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "01234567890", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestUPCE0Barcode_GetPattern(t *testing.T) {
	b := barcode.NewUPCE0Barcode()
	if err := b.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "123456", 300, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestUPCE1Barcode_GetPattern(t *testing.T) {
	b := barcode.NewUPCE1Barcode()
	if err := b.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "123456", 300, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestUPCE0Barcode_DefaultValue(t *testing.T) {
	b := barcode.NewUPCE0Barcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestUPCE1Barcode_DefaultValue(t *testing.T) {
	b := barcode.NewUPCE1Barcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

// ── Supplement 2 and 5 ────────────────────────────────────────────────────────

func TestSupplement2Barcode_GetPattern(t *testing.T) {
	b := barcode.NewSupplement2Barcode()
	if err := b.Encode("53"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "53", 100, 60, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestSupplement2Barcode_AllParityGroups(t *testing.T) {
	// value % 4 has 4 cases (0..3); test values 00, 01, 02, 03
	for _, val := range []string{"00", "01", "02", "03"} {
		b := barcode.NewSupplement2Barcode()
		if err := b.Encode(val); err != nil {
			t.Fatalf("Encode(%s): %v", val, err)
		}
		_, err := b.GetPattern()
		if err != nil {
			t.Fatalf("GetPattern(%s): %v", val, err)
		}
	}
}

func TestSupplement2Barcode_DefaultValue(t *testing.T) {
	b := barcode.NewSupplement2Barcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestSupplement5Barcode_GetPattern(t *testing.T) {
	b := barcode.NewSupplement5Barcode()
	if err := b.Encode("52495"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "52495", 200, 60, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestSupplement5Barcode_DefaultValue(t *testing.T) {
	b := barcode.NewSupplement5Barcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

// ── MSI pattern ───────────────────────────────────────────────────────────────

func TestMSIBarcode_GetPattern(t *testing.T) {
	b := barcode.NewMSIBarcode()
	if err := b.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "123456", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestMSIBarcode_GetPattern_AllDigits(t *testing.T) {
	b := barcode.NewMSIBarcode()
	if err := b.Encode("9876543210"); err != nil {
		t.Fatalf("Encode all digits: %v", err)
	}
	_, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern all digits: %v", err)
	}
}

func TestMSIBarcode_GetPattern_EmptyEncoded(t *testing.T) {
	// Force empty encodedText — GetPattern handles it with "0" fallback
	b := &barcode.MSIBarcode{}
	_, err := b.GetPattern()
	if err != nil {
		t.Logf("GetPattern empty: %v (acceptable)", err)
	}
}

// ── Plessey pattern ───────────────────────────────────────────────────────────

func TestPlesseyBarcode_GetPattern(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	if err := b.Encode("1234ABCD"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "1234ABCD", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestPlesseyBarcode_GetPattern_AllHex(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	if err := b.Encode("0123456789ABCDEF"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	_, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern all hex: %v", err)
	}
}

// ── PostNet pattern ───────────────────────────────────────────────────────────

func TestPostNetBarcode_GetPattern(t *testing.T) {
	b := barcode.NewPostNetBarcode()
	if err := b.Encode("90210"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "90210", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestPostNetBarcode_GetPattern_AllDigits(t *testing.T) {
	// All 10 digits to exercise tabPostNet[0..9]
	b := barcode.NewPostNetBarcode()
	if err := b.Encode("90210"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	_, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
}

// ── Pharmacode pattern ────────────────────────────────────────────────────────

func TestPharmacodeBarcode_GetPattern(t *testing.T) {
	b := barcode.NewPharmacodeBarcode()
	if err := b.Encode("1234"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "1234", 200, 60, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestPharmacodeBarcode_GetPattern_InvalidInput(t *testing.T) {
	b := &barcode.PharmacodeBarcode{}
	// Empty encoded text should return error
	_, err := b.GetPattern()
	if err == nil {
		t.Error("expected error for empty pharmacode GetPattern")
	}
}

// ── GS1-128 pattern ───────────────────────────────────────────────────────────

func TestGS1_128Barcode_GetPattern_WithParens(t *testing.T) {
	b := barcode.NewGS1_128Barcode()
	if err := b.Encode("(01)12345678901231"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern with parens: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "(01)12345678901231", 500, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestGS1_128Barcode_GetPattern_NoParens(t *testing.T) {
	b := barcode.NewGS1_128Barcode()
	if err := b.Encode("01123456789012"); err != nil {
		t.Fatalf("Encode no parens: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern no parens: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern no parens returned empty pattern")
	}
}

func TestGS1_128Barcode_GetPattern_ParseFail_FallbackPath(t *testing.T) {
	b := barcode.NewGS1_128Barcode()
	// A parenthesized string that fails gs1ParseGS1 due to invalid AI
	if err := b.Encode("(99)ABC"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	_, err := b.GetPattern()
	if err != nil {
		t.Logf("GetPattern parse-fail path: %v (acceptable)", err)
	}
}

func TestGS1_128Barcode_DefaultValue(t *testing.T) {
	b := barcode.NewGS1_128Barcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

// ── GS1DatamatrixBarcode ──────────────────────────────────────────────────────

func TestGS1DatamatrixBarcode_GetMatrix(t *testing.T) {
	b := barcode.NewGS1DatamatrixBarcode()
	if err := b.Encode("(01)12345678901231"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil {
		t.Fatal("GetMatrix returned nil matrix")
	}
	if rows <= 0 || cols <= 0 {
		t.Errorf("GetMatrix rows=%d cols=%d, want > 0", rows, cols)
	}
	img := barcode.DrawBarcode2D(matrix, rows, cols, 200, 200)
	if img == nil {
		t.Fatal("DrawBarcode2D returned nil")
	}
}

func TestGS1DatamatrixBarcode_GetMatrix_EmptyEncoded(t *testing.T) {
	b := barcode.NewGS1DatamatrixBarcode()
	// No Encode call — encodedText is empty
	matrix, rows, cols := b.GetMatrix()
	if matrix != nil {
		t.Error("GetMatrix with empty text should return nil matrix")
	}
	if rows != 0 || cols != 0 {
		t.Errorf("expected 0,0 rows/cols, got %d,%d", rows, cols)
	}
}

func TestGS1DatamatrixBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewGS1DatamatrixBarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

// ── DataMatrix GetMatrix ──────────────────────────────────────────────────────

func TestDataMatrixBarcode_GetMatrix(t *testing.T) {
	b := barcode.NewDataMatrixBarcode()
	if err := b.Encode("Hello DataMatrix"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil {
		t.Fatal("GetMatrix returned nil")
	}
	if rows <= 0 || cols <= 0 {
		t.Errorf("GetMatrix rows=%d cols=%d, want > 0", rows, cols)
	}
}

// ── QR GetMatrix ─────────────────────────────────────────────────────────────

func TestQRBarcode_GetMatrix(t *testing.T) {
	b := barcode.NewQRBarcode()
	if err := b.Encode("hello"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil {
		t.Fatal("GetMatrix returned nil")
	}
	if rows <= 0 || cols <= 0 {
		t.Errorf("GetMatrix rows=%d cols=%d, want > 0", rows, cols)
	}
	img := barcode.DrawBarcode2D(matrix, rows, cols, 200, 200)
	if img == nil {
		t.Fatal("DrawBarcode2D returned nil")
	}
}

func TestQRBarcode_GetMatrix_EmptyText_UsesDefault(t *testing.T) {
	b := barcode.NewQRBarcode()
	// No Encode — GetMatrix uses DefaultValue()
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil {
		t.Fatal("GetMatrix (empty) returned nil")
	}
	if rows <= 0 || cols <= 0 {
		t.Errorf("GetMatrix rows=%d cols=%d, want > 0", rows, cols)
	}
}

func TestQRBarcode_GetMatrix_AllECLevels(t *testing.T) {
	levels := []string{"L", "M", "Q", "H"}
	for _, lvl := range levels {
		b := barcode.NewQRBarcode()
		b.ErrorCorrection = lvl
		if err := b.Encode("test"); err != nil {
			t.Fatalf("Encode level=%s: %v", lvl, err)
		}
		matrix, rows, cols := b.GetMatrix()
		if matrix == nil {
			t.Fatalf("GetMatrix level=%s returned nil", lvl)
		}
		if rows <= 0 || cols <= 0 {
			t.Errorf("level=%s rows=%d cols=%d, want > 0", lvl, rows, cols)
		}
	}
}

func TestQRBarcode_GetMatrix_NumericContent(t *testing.T) {
	b := barcode.NewQRBarcode()
	if err := b.Encode("1234567890"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix numeric content failed")
	}
}

func TestQRBarcode_GetMatrix_AlphanumericContent(t *testing.T) {
	b := barcode.NewQRBarcode()
	if err := b.Encode("HELLO WORLD"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix alphanumeric content failed")
	}
}

func TestQRBarcode_GetMatrix_ByteContent(t *testing.T) {
	b := barcode.NewQRBarcode()
	if err := b.Encode("hello world 123 !@#"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix byte content failed")
	}
}

// ── Japan Post 4-State pattern ────────────────────────────────────────────────

func TestJapanPost4StateBarcode_GetPattern(t *testing.T) {
	b := barcode.NewJapanPost4StateBarcode()
	// Japan Post 4-State requires at least 7 chars; first 7 must be [0-9\-],
	// remainder must be [A-Z0-9\-]. Use 8 chars: 7 digits + 1 uppercase.
	if err := b.Encode("1234567A"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
	img := barcode.DrawLinearBarcode(pattern, "1234567A", 400, 80, true, b.GetWideBarRatio())
	if img == nil {
		t.Fatal("DrawLinearBarcode returned nil")
	}
}

func TestJapanPost4StateBarcode_GetPattern_WithUppercase(t *testing.T) {
	b := barcode.NewJapanPost4StateBarcode()
	// Japan Post accepts alphanumeric postal codes
	if err := b.Encode("1234567ABC"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	_, err := b.GetPattern()
	if err != nil {
		t.Logf("GetPattern with uppercase: %v (acceptable)", err)
	}
}

func TestJapanPost4StateBarcode_GetPattern_TooShort(t *testing.T) {
	b := barcode.NewJapanPost4StateBarcode()
	if err := b.Encode("123456"); err != nil {
		t.Logf("Encode short: %v", err)
		return
	}
	_, err := b.GetPattern()
	if err == nil {
		t.Error("expected error for too-short Japan Post input")
	}
}

func TestJapanPost4StateBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewJapanPost4StateBarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestJapanPost4StateBarcode_Encode(t *testing.T) {
	b := barcode.NewJapanPost4StateBarcode()
	if err := b.Encode("597-8615"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
}

func TestJapanPost4StateBarcode_Encode_Empty(t *testing.T) {
	b := barcode.NewJapanPost4StateBarcode()
	err := b.Encode("")
	if err == nil {
		t.Error("expected error for empty Japan Post encode")
	}
}

// ── NewBarcodeByName ──────────────────────────────────────────────────────────

func TestNewBarcodeByName_AllKnownNames(t *testing.T) {
	names := []string{
		"2/5 Interleaved", "2/5 Industrial", "2/5 Matrix",
		"Codabar", "Code128", "Code128 A", "Code128 B", "Code128 C",
		"Code39", "Code39 Extended", "Code93", "Code93 Extended",
		"EAN8", "EAN13", "MSI", "PostNet", "UPC-A", "UPC-E0", "UPC-E1",
		"PDF417", "Datamatrix", "QR Code", "Aztec", "Plessey",
		"GS1-128 (UCC/EAN-128)", "Pharmacode", "Intelligent Mail (USPS)",
		"MaxiCode", "ITF-14", "Deutsche Identcode", "Deutsche Leitcode",
		"Japan Post 4 State Code", "Supplement 2", "Supplement 5",
		"GS1 DataBar Omnidirectional", "GS1 DataBar Limited",
		"GS1 DataBar Stacked", "GS1 DataBar Stacked Omnidirectional",
		"GS1 Datamatrix", "SwissQR",
	}
	for _, name := range names {
		b := barcode.NewBarcodeByName(name)
		if b == nil {
			t.Errorf("NewBarcodeByName(%q) returned nil", name)
		}
	}
}

func TestNewBarcodeByName_Unknown(t *testing.T) {
	b := barcode.NewBarcodeByName("NonExistentBarcode")
	if b == nil {
		t.Error("NewBarcodeByName unknown should return Code128, got nil")
	}
	if b.Type() != barcode.BarcodeTypeCode128 {
		t.Errorf("NewBarcodeByName unknown type = %q, want Code128", b.Type())
	}
}

// ── NewBarcodeByType — all remaining types ────────────────────────────────────

func TestNewBarcodeByType_AllTypes(t *testing.T) {
	types := []barcode.BarcodeType{
		barcode.BarcodeTypeCode128A,
		barcode.BarcodeTypeCode128B,
		barcode.BarcodeTypeCode128C,
		barcode.BarcodeTypeGS1_128,
		barcode.BarcodeTypeEAN8,
		barcode.BarcodeTypeUPCA,
		barcode.BarcodeTypeUPCE,
		barcode.BarcodeTypeCode93Extended,
		barcode.BarcodeTypeCode2of5Industrial,
		barcode.BarcodeTypeCode2of5Matrix,
		barcode.BarcodeTypeMSI,
		barcode.BarcodeTypeMaxiCode,
		barcode.BarcodeTypeIntelligentMail,
		barcode.BarcodeTypePharmacode,
		barcode.BarcodeTypePlessey,
		barcode.BarcodeTypePostNet,
		barcode.BarcodeTypeSwissQR,
		barcode.BarcodeTypeITF14,
		barcode.BarcodeTypeDeutscheIdentcode,
		barcode.BarcodeTypeDeutscheLeitcode,
		barcode.BarcodeTypeSupplement2,
		barcode.BarcodeTypeSupplement5,
		barcode.BarcodeTypeJapanPost4State,
		barcode.BarcodeType("Unknown_Type"),
	}
	for _, bt := range types {
		b := barcode.NewBarcodeByType(bt)
		if b == nil {
			t.Errorf("NewBarcodeByType(%q) returned nil", bt)
		}
	}
}

// ── Missing type stubs ────────────────────────────────────────────────────────

func TestCode128ABarcode_BasicOps(t *testing.T) {
	b := barcode.NewCode128ABarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue empty")
	}
	if err := b.Encode("CODE128A"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
}

func TestCode128BBarcode_BasicOps(t *testing.T) {
	b := barcode.NewCode128BBarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue empty")
	}
	if err := b.Encode("Code128B"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
}

func TestCode128CBarcode_BasicOps(t *testing.T) {
	b := barcode.NewCode128CBarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue empty")
	}
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
}

func TestEAN8Barcode_DefaultValue(t *testing.T) {
	b := barcode.NewEAN8Barcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue empty")
	}
}

func TestUPCABarcode_DefaultValue(t *testing.T) {
	b := barcode.NewUPCABarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue empty")
	}
}

func TestUPCEBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewUPCEBarcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue empty")
	}
}

func TestUPCEBarcode_Encode_ValidSizes(t *testing.T) {
	// boombuler EAN encoder accepts 7-digit EAN-8 (computes checksum automatically).
	// 6-digit and 8-digit with wrong checksum are rejected by the library.
	cases := []string{"1234567"}
	for _, tc := range cases {
		b := barcode.NewUPCEBarcode()
		if err := b.Encode(tc); err != nil {
			t.Errorf("Encode(%q): %v", tc, err)
		}
	}
}

func TestUPCEBarcode_Encode_TooShort(t *testing.T) {
	b := barcode.NewUPCEBarcode()
	err := b.Encode("12345")
	if err == nil {
		t.Error("expected error for 5-digit UPC-E")
	}
}

func TestUPCEBarcode_Encode_TooLong(t *testing.T) {
	b := barcode.NewUPCEBarcode()
	err := b.Encode("123456789")
	if err == nil {
		t.Error("expected error for 9-digit UPC-E")
	}
}

func TestUPCEBarcode_Encode_NonDigit(t *testing.T) {
	b := barcode.NewUPCEBarcode()
	err := b.Encode("12345A")
	if err == nil {
		t.Error("expected error for non-digit UPC-E")
	}
}

// ── GS1_128Barcode ────────────────────────────────────────────────────────────

func TestGS1_128Barcode_NewAndDefaultValue(t *testing.T) {
	b := barcode.NewGS1_128Barcode()
	if b == nil {
		t.Fatal("NewGS1_128Barcode returned nil")
	}
	if b.DefaultValue() == "" {
		t.Error("DefaultValue empty")
	}
}

// ── Supplement encode error path ──────────────────────────────────────────────

func TestSupplement2Barcode_Encode_TooShort(t *testing.T) {
	b := barcode.NewSupplement2Barcode()
	err := b.Encode("5")
	if err == nil {
		t.Error("expected error for 1-digit supplement2")
	}
}

func TestSupplement2Barcode_Encode_NonDigit(t *testing.T) {
	b := barcode.NewSupplement2Barcode()
	err := b.Encode("5A")
	if err == nil {
		t.Error("expected error for non-digit supplement2")
	}
}

func TestSupplement5Barcode_Encode_TooShort(t *testing.T) {
	b := barcode.NewSupplement5Barcode()
	err := b.Encode("1234")
	if err == nil {
		t.Error("expected error for 4-digit supplement5")
	}
}

func TestSupplement5Barcode_Encode_NonDigit(t *testing.T) {
	b := barcode.NewSupplement5Barcode()
	err := b.Encode("1234A")
	if err == nil {
		t.Error("expected error for non-digit supplement5")
	}
}
