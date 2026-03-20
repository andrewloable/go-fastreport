// msi_calcbounds_test.go validates that MSI barcode produces widths matching
// C# FastReport HTML output.
//
// C# reference: BarcodeMSI.cs
//   LinearBarcodeBase.cs:636 WideBarRatio = 2 (default)
//   BarcodeBase.GetDefaultValue() = "12345678" (no override in BarcodeMSI.cs)
//
// FRX reference (Barcode.frx):
//   Barcode32: MSI no Text → C# default "12345678", Width=143.75px
//
// Width calculation for "12345678" (8 data digits, 1 check digit):
//   start "60" = 3 mod, each digit = 12 mod (8 chars each), stop "515" = 4 mod
//   Total = 3 + 9×12 + 4 = 3+108+4 = 115 modules × 1.25 = 143.75px
//
// Pattern starts with "60" (start) and ends with "515" (stop).
// Pattern length: start(2)+9×8(72)+stop(3) = 77 chars
//
// go-fastreport-d2ik: Test MSI CalcBounds against C# expected width
package barcode_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// TestMSI_GetWideBarRatio asserts WideBarRatio=2.
func TestMSI_GetWideBarRatio(t *testing.T) {
	b := barcode.NewMSIBarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("MSI GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestMSI_UpdateAutoSize_12345678 asserts final width=143.75px matching C#
// Barcode.frx HTML output (Barcode32). C# uses default text "12345678"
// (BarcodeBase.GetDefaultValue()="12345678", BarcodeMSI.cs has no override).
// Width: start(3)+8 data digits×12+checksum×12+stop(4) = 3+96+12+4=115 modules × 1.25=143.75px.
func TestMSI_UpdateAutoSize_12345678(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewMSIBarcode()
	if err := obj.Barcode.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 143.75
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("MSI UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode32)", got, wantW)
	}
}

// TestMSI_Pattern_12345678 validates the MSI bar pattern.
// C# BarcodeMSI.cs: start = "60", stop = "515".
// Pattern: start(2)+8 data digits×8+checksum×8+stop(3) = 2+64+8+3 = 77 chars.
func TestMSI_Pattern_12345678(t *testing.T) {
	b := barcode.NewMSIBarcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantLen = 77
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d", len(pattern), wantLen)
	}
	if len(pattern) >= 2 && pattern[:2] != "60" {
		t.Errorf("pattern start = %q, want \"60\" (C# BarcodeMSI start)", pattern[:2])
	}
	if len(pattern) >= 3 && pattern[len(pattern)-3:] != "515" {
		t.Errorf("pattern stop = %q, want \"515\" (C# BarcodeMSI stop)", pattern[len(pattern)-3:])
	}
}
