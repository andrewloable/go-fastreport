// pharmacode_plessey_calcbounds_test.go validates that Pharmacode and Plessey
// barcode types produce widths and patterns matching C# FastReport HTML output.
//
// C# reference: BarcodePharmacodeBase.cs, BarcodePlessey.cs
//   LinearBarcodeBase.cs:636 WideBarRatio = 2 (default)
//
// FRX reference (Barcode.frx):
//   Barcode2:  Pharmacode Text="123456" Width=98.75px
//   Barcode40: Plessey    Text="123456" Width=153.75px
//
// Pharmacode width for "123456":
//   val = 123457 → binary drop leading 1 → "1110001001000001" (16 bits)
//   Pattern: "2"(quiet) + per bit: '1'→"72"(3+3=6mod), '0'→"52"(1+3=4mod)
//   6 wide bars: 6×6=36 mod, 10 narrow bars: 10×4=40 mod, 17 spaces: 17×3=51mod
//   Wait: '7'=3mod bar, '5'=1mod bar, '2'=3mod space
//   Total = 17×3(spaces) + 6×3(wide bars) + 10×1(narrow bars) = 51+18+10 = 79 mod × 1.25 = 98.75px
//
// Plessey width for "123456":
//   start "606050060" (12 mod) + 6 digits × 12 mod each (72 mod) +
//   8 CRC bits × 3 mod each (24 mod) + end "70050050606" (15 mod)
//   = 12+72+24+15 = 123 modules × 1.25 = 153.75px
//
// go-fastreport-biq6: Test Pharmacode/Plessey CalcBounds against C# expected widths
package barcode_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ── Pharmacode ─────────────────────────────────────────────────────────────

// TestPharmacode_GetWideBarRatio asserts WideBarRatio=2.
func TestPharmacode_GetWideBarRatio(t *testing.T) {
	b := barcode.NewPharmacodeBarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("Pharmacode GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestPharmacode_UpdateAutoSize_123456 asserts final width=98.75px matching C#
// Barcode.frx HTML output (Barcode2, Text="123456").
// Width: 79 modules × 1.25 = 98.75px.
func TestPharmacode_UpdateAutoSize_123456(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewPharmacodeBarcode()
	if err := obj.Barcode.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 98.75
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("Pharmacode UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode2)", got, wantW)
	}
}

// TestPharmacode_Pattern_123456 validates the Pharmacode bar pattern for "123456".
// val=123457 binary drop leading '1' → 16-bit "1110001001000001".
// Pattern: "2" quiet + per bit: '1'→"72", '0'→"52". Length = 1 + 16×2 = 33 chars.
// Starts with "2" (leading quiet zone), ends with "72" (last wide bar + space).
func TestPharmacode_Pattern_123456(t *testing.T) {
	b := barcode.NewPharmacodeBarcode()
	if err := b.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantLen = 33
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d", len(pattern), wantLen)
	}
	if len(pattern) >= 1 && pattern[:1] != "2" {
		t.Errorf("pattern start = %q, want \"2\" (Pharmacode leading quiet zone)", pattern[:1])
	}
	if len(pattern) >= 2 && pattern[len(pattern)-2:] != "72" {
		t.Errorf("pattern end = %q, want \"72\" (Pharmacode last wide bar + space)", pattern[len(pattern)-2:])
	}
}

// ── Plessey ────────────────────────────────────────────────────────────────

// TestPlessey_GetWideBarRatio asserts WideBarRatio=2.
func TestPlessey_GetWideBarRatio(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("Plessey GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestPlessey_UpdateAutoSize_123456 asserts final width=153.75px matching C#
// Barcode.frx HTML output (Barcode40, Text="123456").
// Width: start(12)+6×12(72)+8 CRC bits×3(24)+end(15) = 123 modules × 1.25 = 153.75px.
func TestPlessey_UpdateAutoSize_123456(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewPlesseyBarcode()
	if err := obj.Barcode.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 153.75
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("Plessey UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode40)", got, wantW)
	}
}

// TestPlessey_Pattern_123456 validates the Plessey bar pattern for "123456".
// C# BarcodePlessey.cs: start = "606050060", end = "70050050606".
// Pattern length: start(9)+6×data+8 CRC bits+end(11) = 104 chars.
func TestPlessey_Pattern_123456(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	if err := b.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantLen = 104
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d", len(pattern), wantLen)
	}
	if len(pattern) >= 9 && pattern[:9] != "606050060" {
		t.Errorf("pattern start = %q, want \"606050060\" (C# Plessey start)", pattern[:9])
	}
	if len(pattern) >= 11 && pattern[len(pattern)-11:] != "70050050606" {
		t.Errorf("pattern end = %q, want \"70050050606\" (C# Plessey end)", pattern[len(pattern)-11:])
	}
}
