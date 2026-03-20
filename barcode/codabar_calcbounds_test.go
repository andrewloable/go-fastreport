// codabar_calcbounds_test.go validates that Codabar barcode produces widths
// matching C# FastReport HTML output.
//
// C# reference: BarcodeCodabar.cs
//   LinearBarcodeBase.cs:636 WideBarRatio = 2 (default)
//   BarcodeBase.GetDefaultValue() = "12345678" (no override in BarcodeCodabar.cs)
//   C# StartChar=A, StopChar=B (default properties) — added separately from text
//
// FRX reference (Barcode.frx):
//   Barcode22: Codabar no Text → C# encodes as A + "12345678" + B
//   Width=126.25px
//
// Width calculation (A start, "12345678" data, B stop):
//   Each Codabar char = 9 modules + 1 separator = 10 per data char
//   Start A: 10+1 sep = 11 mod, Data 8×10 = 80 mod, Stop B: 10 mod
//   Total = 11+80+10 = 101 modules × 1.25 = 126.25px
//
// Pattern starts with start-char A data "5061515", ends with stop-char B data "5151506".
//
// go-fastreport-7xd5: Test Codabar CalcBounds against C# expected width
package barcode_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// TestCodabar_GetWideBarRatio asserts WideBarRatio=2.
func TestCodabar_GetWideBarRatio(t *testing.T) {
	b := barcode.NewCodabarBarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("Codabar GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestCodabar_UpdateAutoSize_A12345678B asserts final width=126.25px matching C#
// Barcode.frx HTML output (Barcode22). C# uses default text "12345678" plus
// StartChar=A and StopChar=B (set as separate properties), equivalent to "A12345678B"
// in Go's input format.
// Width: StartA(11)+8×data(80)+StopB(10) = 101 modules × 1.25 = 126.25px.
func TestCodabar_UpdateAutoSize_A12345678B(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewCodabarBarcode()
	if err := obj.Barcode.Encode("A12345678B"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 126.25
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("Codabar UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode22)", got, wantW)
	}
}

// TestCodabar_Pattern_A12345678B validates the Codabar bar pattern.
// C# BarcodeCodabar.cs: start char A → data "5061515", stop char B → data "5151506".
// Pattern starts with "5061515" (start A) and ends with "5151506" (stop B).
// Pattern length: start(7+1)+8×data(8×8)+stop(7) = 8+64+7 = 79 chars.
func TestCodabar_Pattern_A12345678B(t *testing.T) {
	b := barcode.NewCodabarBarcode()
	if err := b.Encode("A12345678B"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantLen = 79
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d", len(pattern), wantLen)
	}
	// Start char A: "5061515" (first 7 chars)
	if len(pattern) >= 7 && pattern[:7] != "5061515" {
		t.Errorf("pattern start = %q, want \"5061515\" (C# Codabar char A)", pattern[:7])
	}
	// Stop char B: "5151506" (last 7 chars)
	if len(pattern) >= 7 && pattern[len(pattern)-7:] != "5151506" {
		t.Errorf("pattern stop = %q, want \"5151506\" (C# Codabar char B)", pattern[len(pattern)-7:])
	}
}
