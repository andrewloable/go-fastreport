// ean_upc_calcbounds_test.go validates that EAN/UPC family barcodes produce
// widths matching C# FastReport HTML output.
//
// C# reference: BarcodeEAN.cs, BarcodeUPC.cs
//   LinearBarcodeBase.cs:636 WideBarRatio = 2 (default)
//   BarcodeBase.GetDefaultValue() = "12345678" (no override in EAN/UPC classes)
//
// FRX reference (Barcode.frx):
//   Barcode27: EAN-13 Text="1234567890123" Width=128.75
//   Barcode26: EAN-8  no Text (C# default "12345678") Width=83.75
//   Barcode28: UPC-A  Text="123456789012"  Width=138.75
//   Barcode29: UPC-E0 Text="1234567"       Width=83.75
//
// EAN-8 module count: start(3)+4×7+center(5)+4×7+stop(3) = 67 modules × 1.25 = 83.75px
//
// Pattern format: guard bars encoded as 'A'/'B' (BarLineBlackLong).
// EAN-8/13 start/stop = "A0A", center = "0A0A0".
//
// go-fastreport-mv43: Test EAN/UPC family CalcBounds against C# expected widths
package barcode_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ── EAN-13 ─────────────────────────────────────────────────────────────────

// TestEAN13_GetWideBarRatio asserts WideBarRatio=2.
func TestEAN13_GetWideBarRatio(t *testing.T) {
	b := barcode.NewEAN13Barcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("EAN13 GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestEAN13_UpdateAutoSize_1234567890123 asserts width matching C# Barcode.frx
// Barcode27 (Text="1234567890123", Width=128.75px).
func TestEAN13_UpdateAutoSize_1234567890123(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewEAN13Barcode()
	if err := obj.Barcode.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 128.75
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("EAN13 UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode27)", got, wantW)
	}
}

// TestEAN13_Pattern_1234567890123 validates the EAN-13 bar pattern.
// EAN-13 pattern starts with "A0A" (start guard) and ends with "A0A" (stop guard).
func TestEAN13_Pattern_1234567890123(t *testing.T) {
	b := barcode.NewEAN13Barcode()
	if err := b.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Fatal("pattern is empty")
	}
	// EAN-13 start guard: "A0A"
	if len(pattern) >= 3 && pattern[:3] != "A0A" {
		t.Errorf("pattern start = %q, want \"A0A\" (EAN-13 start guard)", pattern[:3])
	}
	// EAN-13 stop guard: "A0A"
	if len(pattern) >= 3 && pattern[len(pattern)-3:] != "A0A" {
		t.Errorf("pattern stop = %q, want \"A0A\" (EAN-13 stop guard)", pattern[len(pattern)-3:])
	}
}

// ── EAN-8 ──────────────────────────────────────────────────────────────────

// TestEAN8_GetWideBarRatio asserts WideBarRatio=2.
func TestEAN8_GetWideBarRatio(t *testing.T) {
	b := barcode.NewEAN8Barcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("EAN8 GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestEAN8_UpdateAutoSize_12345678 asserts final width=83.75px matching C#
// Barcode.frx Barcode26 (no Text → C# default "12345678", Width=83.75).
// Width: start(3)+4×7+center(5)+4×7+stop(3) = 67 modules × 1.25 = 83.75px.
// EAN-8 uses first 7 digits and computes the 8th as checksum.
func TestEAN8_UpdateAutoSize_12345678(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewEAN8Barcode()
	if err := obj.Barcode.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 83.75
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("EAN8 UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode26)", got, wantW)
	}
}

// TestEAN8_Pattern_12345678 validates the EAN-8 bar pattern.
// EAN-8 pattern starts with "A0A" (start guard) and ends with "A0A" (stop guard).
// C# GetPattern: start(3)+4 left digits×4 chars(16)+center(5)+4 right digits×4 chars(16)+stop(3) = 43 chars.
func TestEAN8_Pattern_12345678(t *testing.T) {
	b := barcode.NewEAN8Barcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantLen = 43
	if len(pattern) != wantLen {
		t.Errorf("EAN8 pattern length = %d, want %d", len(pattern), wantLen)
	}
	if len(pattern) >= 3 && pattern[:3] != "A0A" {
		t.Errorf("EAN8 pattern start = %q, want \"A0A\"", pattern[:3])
	}
	if len(pattern) >= 3 && pattern[len(pattern)-3:] != "A0A" {
		t.Errorf("EAN8 pattern stop = %q, want \"A0A\"", pattern[len(pattern)-3:])
	}
}

// ── UPC-A ──────────────────────────────────────────────────────────────────

// TestUPCA_GetWideBarRatio asserts WideBarRatio=2.
func TestUPCA_GetWideBarRatio(t *testing.T) {
	b := barcode.NewUPCABarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("UPC-A GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestUPCA_UpdateAutoSize_123456789012 asserts width matching C# Barcode.frx
// Barcode28 (Text="123456789012", Width=138.75px).
func TestUPCA_UpdateAutoSize_123456789012(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewUPCABarcode()
	if err := obj.Barcode.Encode("123456789012"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 138.75
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UPC-A UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode28)", got, wantW)
	}
}

// TestUPCA_Pattern_123456789012 validates UPC-A bar pattern.
// UPC-A starts with "A0A" (start guard) and ends with "A0A" (stop guard).
func TestUPCA_Pattern_123456789012(t *testing.T) {
	b := barcode.NewUPCABarcode()
	if err := b.Encode("123456789012"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Fatal("pattern is empty")
	}
	if len(pattern) >= 3 && pattern[:3] != "A0A" {
		t.Errorf("UPC-A pattern start = %q, want \"A0A\"", pattern[:3])
	}
	if len(pattern) >= 3 && pattern[len(pattern)-3:] != "A0A" {
		t.Errorf("UPC-A pattern stop = %q, want \"A0A\"", pattern[len(pattern)-3:])
	}
}

// ── UPC-E0 ─────────────────────────────────────────────────────────────────

// TestUPCE0_GetWideBarRatio asserts WideBarRatio=2.
func TestUPCE0_GetWideBarRatio(t *testing.T) {
	b := barcode.NewUPCE0Barcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("UPC-E0 GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestUPCE0_UpdateAutoSize_1234567 asserts width matching C# Barcode.frx
// Barcode29 (Text="1234567", Width=83.75px).
func TestUPCE0_UpdateAutoSize_1234567(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewUPCE0Barcode()
	if err := obj.Barcode.Encode("1234567"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 83.75
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UPC-E0 UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode29)", got, wantW)
	}
}
