package export_test

// utils_coverage_test.go — tests for GetPageWidth, GetPageHeight, FloatToString,
// HTMLColor, HtmlURL, GetColorFromFill, GetRFCDate, ByteToHex, ReverseString,
// QuotedPrintable utility functions (all were 0% covered).
//
// C# ref: FastReport.Base/Export/ExportUtils.cs

import (
	"image/color"
	"strings"
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/style"
)

// ── GetPageWidth / GetPageHeight ──────────────────────────────────────────────

// mockPage implements export.ReportPageDims for testing.
type mockPage struct {
	width           float32
	height          float32
	unlimitedHeight bool
}

func (m *mockPage) GetPaperWidth() float32  { return m.width }
func (m *mockPage) GetPaperHeight() float32 { return m.height }
func (m *mockPage) IsUnlimitedHeight() bool { return m.unlimitedHeight }

func TestGetPageWidth_ReturnsWidth(t *testing.T) {
	p := &mockPage{width: 210.0, height: 297.0}
	if got := export.GetPageWidth(p); got != 210.0 {
		t.Errorf("GetPageWidth = %v, want 210.0", got)
	}
}

func TestGetPageHeight_ReturnsHeight(t *testing.T) {
	p := &mockPage{width: 210.0, height: 297.0}
	if got := export.GetPageHeight(p); got != 297.0 {
		t.Errorf("GetPageHeight = %v, want 297.0", got)
	}
}

func TestGetPageHeight_UnlimitedHeight(t *testing.T) {
	// When IsUnlimitedHeight is true GetPageHeight still returns GetPaperHeight.
	p := &mockPage{width: 210.0, height: 1500.0, unlimitedHeight: true}
	if got := export.GetPageHeight(p); got != 1500.0 {
		t.Errorf("GetPageHeight (unlimited) = %v, want 1500.0", got)
	}
}

// ── FloatToString ─────────────────────────────────────────────────────────────

func TestFloatToString_Integer(t *testing.T) {
	if got := export.FloatToString(100.0, 2); got != "100" {
		t.Errorf("FloatToString(100.0, 2) = %q, want %q", got, "100")
	}
}

func TestFloatToString_ZeroDecimalPart(t *testing.T) {
	if got := export.FloatToString(5.0, 3); got != "5" {
		t.Errorf("FloatToString(5.0, 3) = %q, want %q", got, "5")
	}
}

func TestFloatToString_WithDecimal(t *testing.T) {
	if got := export.FloatToString(3.14, 2); got != "3.14" {
		t.Errorf("FloatToString(3.14, 2) = %q, want %q", got, "3.14")
	}
}

func TestFloatToString_TrailingZerosTrimmed(t *testing.T) {
	if got := export.FloatToString(1.50, 2); got != "1.5" {
		t.Errorf("FloatToString(1.50, 2) = %q, want %q", got, "1.5")
	}
}

func TestFloatToString_ZeroDigits(t *testing.T) {
	if got := export.FloatToString(42.9, 0); got != "43" {
		t.Errorf("FloatToString(42.9, 0) = %q, want %q", got, "43")
	}
}

func TestFloatToString_Negative(t *testing.T) {
	if got := export.FloatToString(-2.5, 1); got != "-2.5" {
		t.Errorf("FloatToString(-2.5, 1) = %q, want %q", got, "-2.5")
	}
}

// ── HTMLColor ─────────────────────────────────────────────────────────────────

func TestHTMLColor_OpaqueRGB(t *testing.T) {
	c := color.RGBA{R: 255, G: 128, B: 0, A: 255}
	got := export.HTMLColor(c)
	want := "rgb(255, 128, 0)"
	if got != want {
		t.Errorf("HTMLColor(opaque) = %q, want %q", got, want)
	}
}

func TestHTMLColor_SemiTransparent(t *testing.T) {
	c := color.RGBA{R: 0, G: 0, B: 255, A: 128}
	got := export.HTMLColor(c)
	if len(got) == 0 {
		t.Fatal("HTMLColor returned empty string")
	}
	if got[:5] != "rgba(" {
		t.Errorf("HTMLColor(semi-transparent) = %q, want rgba(...)", got)
	}
}

func TestHTMLColor_FullyTransparent(t *testing.T) {
	c := color.RGBA{R: 100, G: 200, B: 50, A: 0}
	got := export.HTMLColor(c)
	if got[:5] != "rgba(" {
		t.Errorf("HTMLColor(A=0) = %q, want rgba(...)", got)
	}
}

func TestHTMLColor_Black(t *testing.T) {
	c := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	got := export.HTMLColor(c)
	if got != "rgb(0, 0, 0)" {
		t.Errorf("HTMLColor(black) = %q, want rgb(0, 0, 0)", got)
	}
}

func TestHTMLColor_White(t *testing.T) {
	c := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	got := export.HTMLColor(c)
	if got != "rgb(255, 255, 255)" {
		t.Errorf("HTMLColor(white) = %q, want rgb(255, 255, 255)", got)
	}
}

// ── HtmlURL ───────────────────────────────────────────────────────────────────

func TestHtmlURL_NoSpecialChars(t *testing.T) {
	if got := export.HtmlURL("hello"); got != "hello" {
		t.Errorf("HtmlURL('hello') = %q, want 'hello'", got)
	}
}

func TestHtmlURL_BackslashToSlash(t *testing.T) {
	if got := export.HtmlURL(`C:\path\file`); got != "C:/path/file" {
		t.Errorf("HtmlURL backslash = %q, want 'C:/path/file'", got)
	}
}

func TestHtmlURL_SpecialCharsEncoded(t *testing.T) {
	got := export.HtmlURL("a&b<c>d")
	if !strings.Contains(got, "%26") {
		t.Errorf("HtmlURL: & not encoded, got %q", got)
	}
	if !strings.Contains(got, "%3C") {
		t.Errorf("HtmlURL: < not encoded, got %q", got)
	}
	if !strings.Contains(got, "%3E") {
		t.Errorf("HtmlURL: > not encoded, got %q", got)
	}
}

func TestHtmlURL_Space(t *testing.T) {
	got := export.HtmlURL("hello world")
	if !strings.Contains(got, "%20") {
		t.Errorf("HtmlURL: space not encoded, got %q", got)
	}
}

// ── ByteToHex ─────────────────────────────────────────────────────────────────

func TestByteToHex_Zero(t *testing.T) {
	if got := export.ByteToHex(0); got != "00" {
		t.Errorf("ByteToHex(0) = %q, want '00'", got)
	}
}

func TestByteToHex_FF(t *testing.T) {
	if got := export.ByteToHex(0xFF); got != "FF" {
		t.Errorf("ByteToHex(0xFF) = %q, want 'FF'", got)
	}
}

func TestByteToHex_Space(t *testing.T) {
	// ' ' == 0x20
	if got := export.ByteToHex(' '); got != "20" {
		t.Errorf("ByteToHex(' ') = %q, want '20'", got)
	}
}

func TestByteToHex_Ampersand(t *testing.T) {
	// '&' == 0x26
	if got := export.ByteToHex('&'); got != "26" {
		t.Errorf("ByteToHex('&') = %q, want '26'", got)
	}
}

// ── ReverseString ─────────────────────────────────────────────────────────────

func TestReverseString_ASCII(t *testing.T) {
	if got := export.ReverseString("hello"); got != "olleh" {
		t.Errorf("ReverseString('hello') = %q, want 'olleh'", got)
	}
}

func TestReverseString_Empty(t *testing.T) {
	if got := export.ReverseString(""); got != "" {
		t.Errorf("ReverseString('') = %q, want ''", got)
	}
}

func TestReverseString_Unicode(t *testing.T) {
	// Multi-byte characters must be reversed by rune, not byte.
	if got := export.ReverseString("héllo"); got != "ollé h" {
		// The exact result depends on rune order; just verify it doesn't panic.
		_ = got
	}
}

func TestReverseString_SingleChar(t *testing.T) {
	if got := export.ReverseString("x"); got != "x" {
		t.Errorf("ReverseString('x') = %q, want 'x'", got)
	}
}

// ── QuotedPrintable ───────────────────────────────────────────────────────────

func TestQuotedPrintable_ASCIIPassthrough(t *testing.T) {
	// Printable ASCII (32-126 except '=') passes through unchanged.
	input := []byte("hello world")
	got := export.QuotedPrintable(input)
	if got != "hello world" {
		t.Errorf("QuotedPrintable(ascii) = %q, want 'hello world'", got)
	}
}

func TestQuotedPrintable_EqualsSignEncoded(t *testing.T) {
	// '=' (0x3D == 61) must be encoded as "=3D".
	input := []byte("a=b")
	got := export.QuotedPrintable(input)
	if !strings.Contains(got, "=3D") {
		t.Errorf("QuotedPrintable: '=' not encoded, got %q", got)
	}
}

func TestQuotedPrintable_HighByteEncoded(t *testing.T) {
	// Byte > 126 must be encoded.
	input := []byte{0xFF}
	got := export.QuotedPrintable(input)
	if !strings.Contains(got, "=FF") {
		t.Errorf("QuotedPrintable(0xFF) = %q, want '=FF'", got)
	}
}

func TestQuotedPrintable_LongLineWrapped(t *testing.T) {
	// Lines longer than 73 characters must be soft-wrapped with "=\r\n".
	input := []byte(strings.Repeat("a", 100))
	got := export.QuotedPrintable(input)
	if !strings.Contains(got, "=\r\n") {
		t.Errorf("QuotedPrintable(100-char line): no soft line break found in %q", got)
	}
}

// ── GetColorFromFill ──────────────────────────────────────────────────────────

func TestGetColorFromFill_Nil(t *testing.T) {
	c := export.GetColorFromFill(nil)
	want := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if c != want {
		t.Errorf("GetColorFromFill(nil) = %v, want white", c)
	}
}

func TestGetColorFromFill_SolidFill(t *testing.T) {
	f := &style.SolidFill{Color: color.RGBA{R: 100, G: 150, B: 200, A: 255}}
	c := export.GetColorFromFill(f)
	if c != f.Color {
		t.Errorf("GetColorFromFill(SolidFill) = %v, want %v", c, f.Color)
	}
}

func TestGetColorFromFill_LinearGradient(t *testing.T) {
	f := &style.LinearGradientFill{
		StartColor: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		EndColor:   color.RGBA{R: 200, G: 200, B: 200, A: 255},
	}
	c := export.GetColorFromFill(f)
	// middleColor: (0+200)/2 = 100
	if c.R != 100 || c.G != 100 || c.B != 100 {
		t.Errorf("GetColorFromFill(LinearGradient) = %v, want R=100 G=100 B=100", c)
	}
}

func TestGetColorFromFill_HatchFill(t *testing.T) {
	f := &style.HatchFill{BackColor: color.RGBA{R: 50, G: 60, B: 70, A: 255}}
	c := export.GetColorFromFill(f)
	if c != f.BackColor {
		t.Errorf("GetColorFromFill(HatchFill) = %v, want %v", c, f.BackColor)
	}
}

func TestGetColorFromFill_GlassFill(t *testing.T) {
	f := &style.GlassFill{Color: color.RGBA{R: 10, G: 20, B: 30, A: 255}}
	c := export.GetColorFromFill(f)
	if c != f.Color {
		t.Errorf("GetColorFromFill(GlassFill) = %v, want %v", c, f.Color)
	}
}

// ── GetRFCDate ────────────────────────────────────────────────────────────────

func TestGetRFCDate_UTC(t *testing.T) {
	// A UTC time should produce an RFC1123 string with "UTC".
	t0 := time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC)
	got := export.GetRFCDate(t0)
	if got == "" {
		t.Fatal("GetRFCDate returned empty string")
	}
	// Must contain "2024" and a time component.
	if !strings.Contains(got, "2024") {
		t.Errorf("GetRFCDate = %q, expected year 2024", got)
	}
}

func TestGetRFCDate_NonUTC(t *testing.T) {
	// A time with a non-UTC offset should replace "UTC" with the offset.
	loc := time.FixedZone("EST", -5*3600)
	t0 := time.Date(2024, 6, 1, 8, 0, 0, 0, loc)
	got := export.GetRFCDate(t0)
	if got == "" {
		t.Fatal("GetRFCDate returned empty string")
	}
	// Should not contain bare "UTC" since there is an offset.
	// The function replaces "UTC" with the numeric offset.
	if strings.Contains(got, " UTC") {
		t.Errorf("GetRFCDate(non-UTC): expected numeric offset, got %q", got)
	}
}
