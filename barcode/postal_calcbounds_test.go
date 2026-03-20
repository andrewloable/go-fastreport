// postal_calcbounds_test.go validates that PostNet and JapanPost barcode types
// produce widths matching C# FastReport HTML output.
//
// C# reference: BarcodePostNet.cs
//   LinearBarcodeBase.cs:636 WideBarRatio = 2 (default)
//   C# BarcodePostNet does not validate digit count — accepts any digit string.
//   Go PostNet validates: only 5, 9, or 11 digits accepted (USPS standard).
//   BarcodeIntelligentMail.cs:52 overrides GetDefaultValue()
//
// FRX reference (Barcode.frx):
//   Barcode33: PostNet no Text → C# default "12345678" Width=155.0px
//   Barcode51: JapanPost Text="597-8615-5-7-6" Width=332.5px
//   Barcode3:  IntelligentMail Text="12345678901234567890" Width=402.5px
//
// PostNet width for "12345" (5 digits):
//   start "51" = bar(1)+space(2)=3 mod, each digit=5 bars×1+5 spaces×2=15 mod,
//   stop "5" = 1 mod. Total = 3+5×15+1 = 79 modules × 1.25 = 98.75px.
//
// Note: C# Barcode33 uses default "12345678" (8 digits). Go PostNet only accepts
// 5/9/11 digits per USPS standard, so we test with "12345" (5-digit ZIP).
// Width formula is identical: 3+N×15+1 = (4+15N) mod × 1.25.
//
// JapanPost pattern: start "61G1" + 20 encoded chars×(5+1) + check(5) + stop "1G16"
//   = 4+120+5+4=133 chars. Each char = 2 modules (E/F/G/6 all use modules[1]=WBR=2).
//   Total = 133×2 = 266 modules × 1.25 = 332.5px.
//
// go-fastreport-emcu: Test PostNet/IntelligentMail/JapanPost CalcBounds
package barcode_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ── PostNet ────────────────────────────────────────────────────────────────

// TestPostNet_GetWideBarRatio asserts WideBarRatio=2.
func TestPostNet_GetWideBarRatio(t *testing.T) {
	b := barcode.NewPostNetBarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("PostNet GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestPostNet_UpdateAutoSize_12345 asserts final width=98.75px for a 5-digit ZIP.
// C# BarcodePostNet encodes any digit count; Go validates 5/9/11 digits (USPS).
// Width: start(3)+5×15+stop(1) = 79 modules × 1.25 = 98.75px.
func TestPostNet_UpdateAutoSize_12345(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewPostNetBarcode()
	if err := obj.Barcode.Encode("12345"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 98.75
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("PostNet UpdateAutoSize width = %.4f, want %.4f (C# width formula: 3+5×15+1=79 mod×1.25)", got, wantW)
	}
}

// TestPostNet_Pattern_12345 validates the PostNet bar pattern for a 5-digit ZIP.
// C# BarcodePostNet.cs: start = "51" (bar+space), stop = "5" (final bar).
// Pattern length: start(2)+5×10+stop(1) = 53 chars.
func TestPostNet_Pattern_12345(t *testing.T) {
	b := barcode.NewPostNetBarcode()
	if err := b.Encode("12345"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantLen = 53
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d", len(pattern), wantLen)
	}
	if len(pattern) >= 2 && pattern[:2] != "51" {
		t.Errorf("pattern start = %q, want \"51\" (C# PostNet start)", pattern[:2])
	}
	if len(pattern) >= 1 && pattern[len(pattern)-1] != '5' {
		t.Errorf("pattern stop = %q, want \"5\" (C# PostNet stop)", pattern[len(pattern)-1:])
	}
}

// ── Japan Post 4-State ─────────────────────────────────────────────────────

// TestJapanPost_GetWideBarRatio asserts WideBarRatio=2.
func TestJapanPost_GetWideBarRatio(t *testing.T) {
	b := barcode.NewJapanPost4StateBarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("JapanPost GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestJapanPost_UpdateAutoSize_597_8615_5_7_6 asserts final width=332.5px matching C#
// Barcode.frx HTML output (Barcode51, Text="597-8615-5-7-6").
// Width: start(4)+20 encoded chars×(5+1)(120)+check(5)+stop(4)=133 elements × 2 mod each × 1.25 = 332.5px.
func TestJapanPost_UpdateAutoSize_597_8615_5_7_6(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewJapanPost4StateBarcode()
	if err := obj.Barcode.Encode("597-8615-5-7-6"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 332.5
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("JapanPost UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode51)", got, wantW)
	}
}

// TestJapanPost_Pattern_597_8615_5_7_6 validates the Japan Post 4-state bar pattern.
// C# BarcodeJapanPost4StateCode.cs: start = "61G1", stop = "1G16".
// Pattern length: start(4)+20×6(120)+check(5)+stop(4) = 133 chars.
func TestJapanPost_Pattern_597_8615_5_7_6(t *testing.T) {
	b := barcode.NewJapanPost4StateBarcode()
	if err := b.Encode("597-8615-5-7-6"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	const wantLen = 133
	if len(pattern) != wantLen {
		t.Errorf("pattern length = %d, want %d", len(pattern), wantLen)
	}
	if len(pattern) >= 4 && pattern[:4] != "61G1" {
		t.Errorf("pattern start = %q, want \"61G1\" (C# JapanPost start bar)", pattern[:4])
	}
	if len(pattern) >= 4 && pattern[len(pattern)-4:] != "1G16" {
		t.Errorf("pattern stop = %q, want \"1G16\" (C# JapanPost stop bar)", pattern[len(pattern)-4:])
	}
}
