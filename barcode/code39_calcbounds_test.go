// code39_calcbounds_test.go validates that Code39 Extended CalcBounds and
// UpdateAutoSize produce widths matching C# FastReport HTML output.
//
// C# reference: BarcodeCode39.cs — Code39 Extended GetPattern
// LinearBarcodeBase.cs:636 WideBarRatio = 2 (default)
// Test data from Barcode.frx C# HTML output.
//
// go-fastreport-6wqc: Test Code39 Extended CalcBounds against C# expected width
package barcode_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// TestCode39Extended_GetWideBarRatio asserts that Code39Extended uses WideBarRatio=2.
func TestCode39Extended_GetWideBarRatio(t *testing.T) {
	b := barcode.NewCode39ExtendedBarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("Code39Extended GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestCode39Extended_Encode succeeds for "123456".
func TestCode39Extended_Encode(t *testing.T) {
	b := barcode.NewCode39ExtendedBarcode()
	if err := b.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if b.EncodedText() == "" {
		t.Error("EncodedText is empty after Encode")
	}
}

// TestCode39Extended_UpdateAutoSize_123456 asserts that Code39Extended "123456"
// produces a final width of 145px — matching C# Barcode.frx HTML output (Barcode37).
func TestCode39Extended_UpdateAutoSize_123456(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewCode39ExtendedBarcode()
	if err := obj.Barcode.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 145.0
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode37)", got, wantW)
	}
}
