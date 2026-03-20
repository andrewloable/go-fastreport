package barcode_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ── GS1 DataBar Omnidirectional ───────────────────────────────────────────────

func TestGS1DataBarOmni_Encode_SetsCanonicalText(t *testing.T) {
	b := barcode.NewGS1DataBarOmniBarcode()
	if err := b.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// C# GetValue always adds "(01)" prefix and appends check digit.
	want := "(01)12345678901231"
	if got := b.EncodedText(); got != want {
		t.Errorf("EncodedText = %q, want %q", got, want)
	}
}

func TestGS1DataBarOmni_Encode_WithPrefix(t *testing.T) {
	b := barcode.NewGS1DataBarOmniBarcode()
	if err := b.Encode("(01)1234567890123"); err != nil {
		t.Fatalf("Encode with prefix: %v", err)
	}
	want := "(01)12345678901231"
	if got := b.EncodedText(); got != want {
		t.Errorf("EncodedText = %q, want %q", got, want)
	}
}

func TestGS1DataBarOmni_CalcBounds(t *testing.T) {
	b := barcode.NewGS1DataBarOmniBarcode()
	if err := b.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	w, h := b.CalcBounds()
	if h != 0 {
		t.Errorf("CalcBounds h = %.2f, want 0 (linear)", h)
	}
	// C# omni barWidth = 96 modules * 2.0 wideBarRatio = 192; CalcBounds = 192 * 1.25 = 240.
	const wantW = 240.0
	if math.Abs(float64(w-wantW)) > 0.5 {
		t.Errorf("CalcBounds w = %.2f, want %.2f", w, wantW)
	}
}

func TestGS1DataBarOmni_InvalidInput(t *testing.T) {
	b := barcode.NewGS1DataBarOmniBarcode()
	if err := b.Encode("notanumber"); err == nil {
		t.Error("Encode non-numeric: expected error, got nil")
	}
	if err := b.Encode("123456789012"); err == nil { // only 12 digits
		t.Error("Encode 12 digits: expected error, got nil")
	}
}

// ── GS1 DataBar Stacked ───────────────────────────────────────────────────────

func TestGS1DataBarStacked_Encode_SetsCanonicalText(t *testing.T) {
	b := barcode.NewGS1DataBarStackedBarcode()
	if err := b.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	want := "(01)12345678901231"
	if got := b.EncodedText(); got != want {
		t.Errorf("EncodedText = %q, want %q", got, want)
	}
}

func TestGS1DataBarStacked_CalcBounds_BarOnly(t *testing.T) {
	b := barcode.NewGS1DataBarStackedBarcode()
	if err := b.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// CalcBounds returns bar-only width: row0 sums 50 modules * 2.0 * 1.25 = 125.
	w, h := b.CalcBounds()
	if h != 0 {
		t.Errorf("CalcBounds h = %.2f, want 0 (linear)", h)
	}
	const wantW = 125.0
	if math.Abs(float64(w-wantW)) > 0.5 {
		t.Errorf("CalcBounds w = %.2f, want %.2f", w, wantW)
	}
}

func TestGS1DataBarStacked_UpdateAutoSize_ShowText(t *testing.T) {
	// BarcodeObject.UpdateAutoSize with showText=true should add extra padding
	// when the display text "(01)12345678901235" (18 chars) is wider than the
	// bar area (100px), matching C# LinearBarcodeBase.CalcBounds behaviour.
	// Expected: extra = (104 - 100)/2 + 2 = 4 → w = (100 + 8) * 1.25 = 135.
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewGS1DataBarStackedBarcode()
	if err := obj.Barcode.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 135.0
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UpdateAutoSize width = %.2f, want %.2f (C# expected)", got, wantW)
	}
}

func TestGS1DataBarStacked_UpdateAutoSize_NoShowText(t *testing.T) {
	// With showText=false, no extra padding should be added: w = 125.
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewGS1DataBarStackedBarcode()
	obj.SetShowText(false)
	if err := obj.Barcode.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 125.0
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 0.5 {
		t.Errorf("UpdateAutoSize (no ShowText) width = %.2f, want %.2f", got, wantW)
	}
}

// ── GS1 DataBar Stacked Omni ─────────────────────────────────────────────────

func TestGS1DataBarStackedOmni_Encode_SetsCanonicalText(t *testing.T) {
	b := barcode.NewGS1DataBarStackedOmniBarcode()
	if err := b.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	want := "(01)12345678901231"
	if got := b.EncodedText(); got != want {
		t.Errorf("EncodedText = %q, want %q", got, want)
	}
}

// ── GS1 DataBar Limited ───────────────────────────────────────────────────────

func TestGS1DataBarLimited_Encode_SetsCanonicalText(t *testing.T) {
	b := barcode.NewGS1DataBarLimitedBarcode()
	if err := b.Encode("0123456789012"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got := b.EncodedText()
	if len(got) == 0 {
		t.Error("EncodedText is empty after Encode")
	}
}

// TestGS1DataBarLimited_UpdateAutoSize_1234567890123 asserts final width=185px
// matching C# Barcode.frx HTML output (Barcode53, Text="1234567890123").
func TestGS1DataBarLimited_UpdateAutoSize_1234567890123(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewGS1DataBarLimitedBarcode()
	if err := obj.Barcode.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 185.0
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("GS1DataBarLimited UpdateAutoSize width = %.2f, want %.2f (C# Barcode.frx Barcode53)", got, wantW)
	}
}

// TestGS1DataBarLimited_Encode_SetsCanonical asserts EncodedText includes "(01)" prefix.
func TestGS1DataBarLimited_Encode_CanonicalText(t *testing.T) {
	b := barcode.NewGS1DataBarLimitedBarcode()
	if err := b.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	want := "(01)12345678901231"
	if got := b.EncodedText(); got != want {
		t.Errorf("EncodedText = %q, want %q", got, want)
	}
}

// ── GS1 DataBar Stacked Omnidirectional ──────────────────────────────────────

// TestGS1DataBarStackedOmni_UpdateAutoSize_1234567890123 asserts final width=135px
// matching C# Barcode.frx HTML output (Barcode54, Text="1234567890123").
func TestGS1DataBarStackedOmni_UpdateAutoSize_1234567890123(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewGS1DataBarStackedOmniBarcode()
	if err := obj.Barcode.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 135.0
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("GS1DataBarStackedOmni UpdateAutoSize width = %.2f, want %.2f (C# Barcode.frx Barcode54)", got, wantW)
	}
}
