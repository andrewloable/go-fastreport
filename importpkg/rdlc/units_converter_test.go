package rdlc

// Internal-package tests for units_converter.go helper functions.
// These tests verify the RDL/RDLC unit conversion helpers directly.

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/style"
)

// ── sizeToPixels ──────────────────────────────────────────────────────────────

func TestSizeToPixels_Millimeters(t *testing.T) {
	// 10mm * 3.78 px/mm = 37.8 px
	got := sizeToPixels("10mm")
	if got < 37.7 || got > 37.9 {
		t.Fatalf("sizeToPixels(10mm) = %v, want ~37.8", got)
	}
}

func TestSizeToPixels_Centimeters(t *testing.T) {
	// 1cm * 37.8 px/cm = 37.8 px
	got := sizeToPixels("1cm")
	if got < 37.7 || got > 37.9 {
		t.Fatalf("sizeToPixels(1cm) = %v, want ~37.8", got)
	}
}

func TestSizeToPixels_Inches(t *testing.T) {
	// 1in * 96 px/in = 96 px
	got := sizeToPixels("1in")
	if got < 95.9 || got > 96.1 {
		t.Fatalf("sizeToPixels(1in) = %v, want 96", got)
	}
}

func TestSizeToPixels_Points(t *testing.T) {
	// 1pt = 0.3528mm * 3.78 px/mm ≈ 1.334 px
	got := sizeToPixels("1pt")
	if got < 1.3 || got > 1.4 {
		t.Fatalf("sizeToPixels(1pt) = %v, want ~1.334", got)
	}
}

func TestSizeToPixels_Pica(t *testing.T) {
	// 1pc = 4.2336mm * 3.78 px/mm ≈ 16.003 px
	got := sizeToPixels("1pc")
	if got < 15.9 || got > 16.1 {
		t.Fatalf("sizeToPixels(1pc) = %v, want ~16.0", got)
	}
}

func TestSizeToPixels_Unknown_ReturnsZero(t *testing.T) {
	got := sizeToPixels("10em")
	if got != 0 {
		t.Fatalf("sizeToPixels(10em) = %v, want 0", got)
	}
}

// ── sizeToMillimeters ─────────────────────────────────────────────────────────

func TestSizeToMillimeters_Millimeters(t *testing.T) {
	got := sizeToMillimeters("100mm")
	if got < 99.9 || got > 100.1 {
		t.Fatalf("sizeToMillimeters(100mm) = %v, want 100", got)
	}
}

func TestSizeToMillimeters_Centimeters(t *testing.T) {
	// 10cm = 100mm
	got := sizeToMillimeters("10cm")
	if got < 99.9 || got > 100.1 {
		t.Fatalf("sizeToMillimeters(10cm) = %v, want 100", got)
	}
}

func TestSizeToMillimeters_Inches(t *testing.T) {
	// 1in = 25.4mm
	got := sizeToMillimeters("1in")
	if got < 25.3 || got > 25.5 {
		t.Fatalf("sizeToMillimeters(1in) = %v, want 25.4", got)
	}
}

func TestSizeToMillimeters_Points(t *testing.T) {
	// 1pt = 0.3528mm
	got := sizeToMillimeters("1pt")
	if got < 0.34 || got > 0.37 {
		t.Fatalf("sizeToMillimeters(1pt) = %v, want ~0.3528", got)
	}
}

func TestSizeToMillimeters_Unknown_ReturnsZero(t *testing.T) {
	got := sizeToMillimeters("10em")
	if got != 0 {
		t.Fatalf("sizeToMillimeters(10em) = %v, want 0", got)
	}
}

// ── booleanToBool ─────────────────────────────────────────────────────────────

func TestBooleanToBool_True(t *testing.T) {
	for _, s := range []string{"true", "True", "TRUE"} {
		if !booleanToBool(s) {
			t.Errorf("booleanToBool(%q) = false, want true", s)
		}
	}
}

func TestBooleanToBool_False(t *testing.T) {
	for _, s := range []string{"false", "False", "FALSE", "no", "0", ""} {
		if booleanToBool(s) {
			t.Errorf("booleanToBool(%q) = true, want false", s)
		}
	}
}

// ── convertColor ─────────────────────────────────────────────────────────────

func TestConvertColor_Red(t *testing.T) {
	// Named color "Red"
	got := convertColor("Red")
	want := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if got != want {
		t.Fatalf("convertColor(Red) = %v, want %v", got, want)
	}
}

func TestConvertColor_HexColor(t *testing.T) {
	// Hex "#FF8000"
	got := convertColor("#FF8000")
	if got.R != 255 || got.G != 128 || got.B != 0 || got.A != 255 {
		t.Fatalf("convertColor(#FF8000) = %v, want R=255 G=128 B=0 A=255", got)
	}
}

func TestConvertColor_Invalid_ReturnsBlack(t *testing.T) {
	// Should return a non-panicking default
	got := convertColor("")
	// An empty string from utils.ParseColor returns an error; we default to black
	if got.A != 255 {
		t.Fatalf("convertColor('') alpha = %d, want 255", got.A)
	}
}

// ── convertFontStyle ─────────────────────────────────────────────────────────

func TestConvertFontStyle_Italic(t *testing.T) {
	got := convertFontStyle("Italic")
	if got != style.FontStyleItalic {
		t.Fatalf("convertFontStyle(Italic) = %v, want FontStyleItalic", got)
	}
}

func TestConvertFontStyle_Other_ReturnsRegular(t *testing.T) {
	for _, s := range []string{"Normal", "Bold", "", "Oblique"} {
		got := convertFontStyle(s)
		if got != style.FontStyleRegular {
			t.Errorf("convertFontStyle(%q) = %v, want FontStyleRegular", s, got)
		}
	}
}

// ── convertFontSize ───────────────────────────────────────────────────────────

func TestConvertFontSize_WithPt(t *testing.T) {
	got := convertFontSize("12pt")
	if got < 11.9 || got > 12.1 {
		t.Fatalf("convertFontSize(12pt) = %v, want 12", got)
	}
}

func TestConvertFontSize_Numeric(t *testing.T) {
	got := convertFontSize("14")
	if got < 13.9 || got > 14.1 {
		t.Fatalf("convertFontSize(14) = %v, want 14", got)
	}
}

// ── convertTextAlign ─────────────────────────────────────────────────────────

func TestConvertTextAlign(t *testing.T) {
	tests := []struct {
		in   string
		want object.HorzAlign
	}{
		{"Center", object.HorzAlignCenter},
		{"Right", object.HorzAlignRight},
		{"Left", object.HorzAlignLeft},
		{"General", object.HorzAlignLeft},
		{"", object.HorzAlignLeft},
	}
	for _, tc := range tests {
		got := convertTextAlign(tc.in)
		if got != tc.want {
			t.Errorf("convertTextAlign(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// ── convertVerticalAlign ─────────────────────────────────────────────────────

func TestConvertVerticalAlign(t *testing.T) {
	tests := []struct {
		in   string
		want object.VertAlign
	}{
		{"Middle", object.VertAlignCenter},
		{"Bottom", object.VertAlignBottom},
		{"Top", object.VertAlignTop},
		{"", object.VertAlignTop},
	}
	for _, tc := range tests {
		got := convertVerticalAlign(tc.in)
		if got != tc.want {
			t.Errorf("convertVerticalAlign(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// ── convertWritingMode ────────────────────────────────────────────────────────

func TestConvertWritingMode_TbRl(t *testing.T) {
	got := convertWritingMode("tb-rl")
	if got != 90 {
		t.Fatalf("convertWritingMode(tb-rl) = %v, want 90", got)
	}
}

func TestConvertWritingMode_Other(t *testing.T) {
	got := convertWritingMode("lr-tb")
	if got != 0 {
		t.Fatalf("convertWritingMode(lr-tb) = %v, want 0", got)
	}
}

// ── convertBorderStyle ────────────────────────────────────────────────────────

func TestConvertBorderStyle(t *testing.T) {
	tests := []struct {
		in   string
		want style.LineStyle
	}{
		{"Dotted", style.LineStyleDot},
		{"Dashed", style.LineStyleDash},
		{"Double", style.LineStyleDouble},
		{"Solid", style.LineStyleSolid},
		{"", style.LineStyleSolid},
	}
	for _, tc := range tests {
		got := convertBorderStyle(tc.in)
		if got != tc.want {
			t.Errorf("convertBorderStyle(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// ── convertSizing ─────────────────────────────────────────────────────────────

func TestConvertSizing(t *testing.T) {
	tests := []struct {
		in   string
		want object.SizeMode
	}{
		{"AutoSize", object.SizeModeAutoSize},
		{"Fit", object.SizeModeStretchImage},
		{"Clip", object.SizeModeNormal},
		{"ProportionalFit", object.SizeModeZoom},
		{"", object.SizeModeZoom},
	}
	for _, tc := range tests {
		got := convertSizing(tc.in)
		if got != tc.want {
			t.Errorf("convertSizing(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// ── sizeToInt ─────────────────────────────────────────────────────────────────

func TestSizeToInt_Points(t *testing.T) {
	got := sizeToInt("4pt", sizeUnitPt)
	if got != 4 {
		t.Fatalf("sizeToInt(4pt, pt) = %v, want 4", got)
	}
}

func TestSizeToInt_NoUnit(t *testing.T) {
	got := sizeToInt("5", "pt")
	if got != 5 {
		t.Fatalf("sizeToInt(5, pt) = %v, want 5", got)
	}
}
