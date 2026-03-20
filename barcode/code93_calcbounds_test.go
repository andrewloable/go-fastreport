// code93_calcbounds_test.go validates that Code93 and Code93 Extended
// CalcBounds and UpdateAutoSize produce widths matching C# FastReport HTML output.
//
// C# reference: BarcodeCode93.cs — GetPattern for Code93 and Code93 Extended
// LinearBarcodeBase.cs:636 WideBarRatio = 2 (default)
// BarcodeBase.GetDefaultValue() = "12345678" (no override in Barcode93.cs)
// Test data from Barcode.frx C# HTML output:
//   Barcode24: Barcode="Code93" Width="136.25" (no Text → default "12345678")
//   Barcode36: Barcode="Code93 Extended" Width="136.25" (no Text → default "12345678")
//
// Width calculation: start(9) + 8 digits×9 + 2 checksums×9 + stop(10) = 109 modules × 1.25 = 136.25px
//
// Pattern validation (DoConvert output):
//   start "111141" at positions 0-5 → "505080"  (4-wide bar encoded as '8')
//   stop  "1111411" at positions 66-72 → "5050805"
//
// go-fastreport-l0l6: Test Code93/Code93 Extended CalcBounds against C# expected widths
package barcode_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// TestCode93_DefaultValue asserts DefaultValue returns "12345678" matching C#
// BarcodeBase.GetDefaultValue() (Barcode93.cs has no override).
func TestCode93_DefaultValue(t *testing.T) {
	b := barcode.NewCode93Barcode()
	got := b.DefaultValue()
	if got != "12345678" {
		t.Errorf("DefaultValue = %q, want \"12345678\" (C# BarcodeBase.GetDefaultValue)", got)
	}
}

// TestCode93_GetWideBarRatio asserts WideBarRatio=2.
func TestCode93_GetWideBarRatio(t *testing.T) {
	b := barcode.NewCode93Barcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("Code93 GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestCode93_UpdateAutoSize_12345678 asserts final width=136.25px matching C#
// Barcode.frx HTML output (Barcode24). C# uses default text "12345678"
// (BarcodeBase.GetDefaultValue()="12345678", Barcode93.cs has no override).
// Width: start(9)+8×9+2×9+stop(10)=109 module units × 1.25 = 136.25px.
func TestCode93_UpdateAutoSize_12345678(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewCode93Barcode()
	if err := obj.Barcode.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 136.25
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode24)", got, wantW)
	}
}

// TestCode93_Pattern_12345678 validates bar pattern for "12345678".
// C# Barcode93.cs:GetPattern starts with "111141" (start), stops with "1111411".
// After DoConvert: start → "505080" (positions 0-5), stop → "5050805" (positions 66-72).
func TestCode93_Pattern_12345678(t *testing.T) {
	b := barcode.NewCode93Barcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantLen = 73
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d", len(pattern), wantLen)
	}
	if len(pattern) >= 6 && pattern[:6] != "505080" {
		t.Errorf("pattern start = %q, want \"505080\" (C# start \"111141\" after DoConvert)", pattern[:6])
	}
	if len(pattern) >= 7 && pattern[len(pattern)-7:] != "5050805" {
		t.Errorf("pattern stop = %q, want \"5050805\" (C# stop \"1111411\" after DoConvert)", pattern[len(pattern)-7:])
	}
}

// TestCode93Extended_GetWideBarRatio asserts WideBarRatio=2.
func TestCode93Extended_GetWideBarRatio(t *testing.T) {
	b := barcode.NewCode93ExtendedBarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("Code93Extended GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestCode93Extended_UpdateAutoSize_12345678 asserts final width=136.25px matching C#
// Barcode.frx HTML output (Barcode36). Digits encode directly as Code93 symbols,
// so Code93 Extended "12345678" = Code93 "12345678" → 136.25px.
func TestCode93Extended_UpdateAutoSize_12345678(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewCode93ExtendedBarcode()
	if err := obj.Barcode.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 136.25
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode36)", got, wantW)
	}
}

// TestCode93Extended_Pattern_12345678 validates bar pattern for Code93 Extended "12345678".
// Digits encode directly, so the pattern is identical to Code93 "12345678".
func TestCode93Extended_Pattern_12345678(t *testing.T) {
	b := barcode.NewCode93ExtendedBarcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantLen = 73
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d", len(pattern), wantLen)
	}
	if len(pattern) >= 6 && pattern[:6] != "505080" {
		t.Errorf("pattern start = %q, want \"505080\"", pattern[:6])
	}
	if len(pattern) >= 7 && pattern[len(pattern)-7:] != "5050805" {
		t.Errorf("pattern stop = %q, want \"5050805\"", pattern[len(pattern)-7:])
	}
}
