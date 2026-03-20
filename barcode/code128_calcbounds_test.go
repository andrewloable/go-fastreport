// code128_calcbounds_test.go validates that Code128 and GS1-128 CalcBounds
// and UpdateAutoSize produce widths matching C# FastReport HTML output.
//
// C# reference:
//   LinearBarcodeBase.cs:636 WideBarRatio = 2 (default)
//   BarcodeBase.GetDefaultValue() = "12345678" (no override in Barcode128.cs)
//   Barcode.frx Barcode25: Barcode="Code128" Barcode.AutoEncode="true" Width="98.75"
//     → no Text attr → C# uses "12345678" default
//     → auto-encode selects Code C (8 digits even) → 4 pairs
//     → StartC(11)+4×11+check(11)+stop(13)=79 modules × 1.25 = 98.75px
//   Barcode.frx Barcode1: Barcode="GS1-128" Text="(10)123456" Width="126.25"
//
// Pattern validation (DoConvert output for Code C start/stop):
//   StartC "211232" at positions 0-5   → "605171"
//   stop   "2331112" at positions 36-42 → "6270506"
//   Total pattern length = 43 chars
//
// go-fastreport-j5n6: Test Code128 CalcBounds against C# expected widths
// go-fastreport-dntv: Test GS1-128 CalcBounds against C# expected width
package barcode_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ── Code128 (Auto-encode) ─────────────────────────────────────────────────────

// TestCode128_DefaultValue asserts DefaultValue returns "12345678" matching C#
// BarcodeBase.GetDefaultValue() (no override in Barcode128.cs).
func TestCode128_DefaultValue(t *testing.T) {
	b := barcode.NewCode128Barcode()
	got := b.DefaultValue()
	if got != "12345678" {
		t.Errorf("Code128 DefaultValue = %q, want \"12345678\" (C# BarcodeBase.GetDefaultValue)", got)
	}
}

// TestCode128_GetWideBarRatio asserts that Code128 uses WideBarRatio=2
// matching the C# LinearBarcodeBase default (LinearBarcodeBase.cs:636).
func TestCode128_GetWideBarRatio(t *testing.T) {
	b := barcode.NewCode128Barcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("Code128 GetWideBarRatio = %.2f, want 2.0 (C# LinearBarcodeBase.cs:636)", got)
	}
}

// TestCode128_UpdateAutoSize_12345678 asserts that Code128 auto-encode "12345678"
// produces a final width of 98.75px — matching C# Barcode.frx HTML output (Barcode25).
// Auto-encode selects Code C (8 digits, even): 4 pairs.
// StartC(11)+4×11(44)+check(11)+stop(13) = 79 modules × 1.25 = 98.75px.
func TestCode128_UpdateAutoSize_12345678(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewCode128Barcode()
	if err := obj.Barcode.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 98.75
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode25)", got, wantW)
	}
}

// TestCode128_Pattern_12345678 validates the bar pattern for "12345678".
// Auto-encode selects Code C. StartC data "211232" → DoConvert at positions 0-5 → "605171".
// Stop data "2331112" → DoConvert at positions 36-42 → "6270506".
// Total pattern length = StartC(6)+4 pairs×6(24)+check(6)+stop(7) = 43 chars.
func TestCode128_Pattern_12345678(t *testing.T) {
	b := barcode.NewCode128Barcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantLen = 43
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d", len(pattern), wantLen)
	}
	// StartC "211232" after DoConvert → "605171"
	if len(pattern) >= 6 && pattern[:6] != "605171" {
		t.Errorf("pattern start = %q, want \"605171\" (C# StartC \"211232\" after DoConvert)", pattern[:6])
	}
	// Stop "2331112" after DoConvert → "6270506"
	if len(pattern) >= 7 && pattern[len(pattern)-7:] != "6270506" {
		t.Errorf("pattern stop = %q, want \"6270506\" (C# stop \"2331112\" after DoConvert)", pattern[len(pattern)-7:])
	}
}

// TestCode128A_GetWideBarRatio checks Code128A variant.
func TestCode128A_GetWideBarRatio(t *testing.T) {
	b := barcode.NewCode128ABarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("Code128A GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestCode128B_GetWideBarRatio checks Code128B variant.
func TestCode128B_GetWideBarRatio(t *testing.T) {
	b := barcode.NewCode128BBarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("Code128B GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestCode128C_GetWideBarRatio checks Code128C variant.
func TestCode128C_GetWideBarRatio(t *testing.T) {
	b := barcode.NewCode128CBarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("Code128C GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// ── GS1-128 ───────────────────────────────────────────────────────────────────

// TestGS1_128_GetWideBarRatio asserts WideBarRatio=2 for GS1-128.
// C# inherits from LinearBarcodeBase which sets WideBarRatio=2.
func TestGS1_128_GetWideBarRatio(t *testing.T) {
	b := barcode.NewGS1_128Barcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("GS1_128 GetWideBarRatio = %.2f, want 2.0 (C# LinearBarcodeBase.cs:636)", got)
	}
}

// TestGS1_128_UpdateAutoSize_10_123456 asserts that GS1-128 with "(10)123456"
// produces a final width of 126.25px — matching C# Barcode.frx HTML output (Barcode1).
func TestGS1_128_UpdateAutoSize_10_123456(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewGS1_128Barcode()
	if err := obj.Barcode.Encode("(10)123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 126.25
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode1)", got, wantW)
	}
}
