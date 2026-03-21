// code2of5_calcbounds_test.go validates that 2-of-5 family barcodes produce
// widths matching C# FastReport HTML output.
//
// C# reference:
//   Barcode2of5Interleaved.cs, Barcode2of5Industrial.cs, Barcode2of5Matrix.cs
//   BarcodeDeutscheIdentcode.cs, BarcodeDeutscheLeitcode.cs
//   LinearBarcodeBase.cs:636 — WideBarRatio=2 default
//   Barcode2of5.cs:487 — WideBarRatio=2.25F for Matrix 2/5
//   Barcode2of5.cs:566 — WideBarRatio=2.25F for ITF-14
//   Barcode2of5.cs:193 — WideBarRatio=3F for Deutsche Identcode
//   Barcode2of5.cs:320 — WideBarRatio=3F for Deutsche Leitcode
//
// C# default text (BarcodeBase.GetDefaultValue() = "12345678") is used for
// barcodes without explicit Text= in Barcode.frx:
//   Barcode20 (Industrial), Barcode21 (Matrix WBR=2.25) both default to "12345678".
//
// go-fastreport-gnsn: Test 2/5 family CalcBounds against C# expected widths
package barcode_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ── 2/5 Interleaved ───────────────────────────────────────────────────────────

// TestCode2of5_DefaultValue asserts DefaultValue="12345670".
func TestCode2of5_DefaultValue(t *testing.T) {
	b := barcode.NewCode2of5Barcode()
	got := b.DefaultValue()
	if got != "12345670" {
		t.Errorf("DefaultValue = %q, want \"12345670\"", got)
	}
}

// TestCode2of5_GetWideBarRatio asserts WideBarRatio=2.
func TestCode2of5_GetWideBarRatio(t *testing.T) {
	b := barcode.NewCode2of5Barcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("Code2of5 GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestCode2of5_UpdateAutoSize_12345670 asserts final width=80px matching C#
// Barcode.frx HTML output (Barcode19).
func TestCode2of5_UpdateAutoSize_12345670(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewCode2of5Barcode()
	if err := obj.Barcode.Encode("12345670"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 80.0
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode19)", got, wantW)
	}
}

// TestCode2of5_Pattern_12345670 validates the bar pattern for "12345670".
// C# Barcode2of5Interleaved.GetPattern() produces an interleaved pattern
// starting with "5050" (start) and ending with "605" (stop).
func TestCode2of5_Pattern_12345670(t *testing.T) {
	b := barcode.NewCode2of5Barcode()
	if err := b.Encode("12345670"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Fatal("GetPattern returned empty string")
	}
	// Pattern must start with "5050" (C# Interleaved start) and end with "605"
	if len(pattern) < 7 || pattern[:4] != "5050" {
		t.Errorf("pattern start = %q, want \"5050...\"", pattern[:min4(len(pattern), 4)])
	}
	if len(pattern) < 3 || pattern[len(pattern)-3:] != "605" {
		t.Errorf("pattern end = %q, want \"...605\"", pattern[max0(len(pattern)-3):])
	}
}

// ── 2/5 Industrial ────────────────────────────────────────────────────────────

// TestCode2of5Industrial_DefaultValue asserts DefaultValue="12345678" (C# BarcodeBase.GetDefaultValue).
func TestCode2of5Industrial_DefaultValue(t *testing.T) {
	b := barcode.NewCode2of5IndustrialBarcode()
	got := b.DefaultValue()
	if got != "12345678" {
		t.Errorf("DefaultValue = %q, want \"12345678\"", got)
	}
}

// TestCode2of5Industrial_GetWideBarRatio asserts WideBarRatio=2.
func TestCode2of5Industrial_GetWideBarRatio(t *testing.T) {
	b := barcode.NewCode2of5IndustrialBarcode()
	got := b.GetWideBarRatio()
	if got != 2.0 {
		t.Errorf("Code2of5Industrial GetWideBarRatio = %.2f, want 2.0", got)
	}
}

// TestCode2of5Industrial_UpdateAutoSize_12345678 asserts final width=140px matching C#
// Barcode.frx HTML output (Barcode20). C# uses default text "12345678"
// (BarcodeBase.GetDefaultValue()="12345678") with Barcode.CalcCheckSum="false".
// 8 digits × 12 module units + start(8) + stop(8) = 112 module units × 1.25 = 140px.
func TestCode2of5Industrial_UpdateAutoSize_12345678(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	ind := barcode.NewCode2of5IndustrialBarcode()
	ind.CalcChecksum = false // C# Barcode.frx Barcode20 has CalcCheckSum=false
	obj.Barcode = ind
	if err := obj.Barcode.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 140.0
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode20)", got, wantW)
	}
}

// TestCode2of5Industrial_Pattern_12345678 validates bar pattern for "12345678".
// C# Barcode2of5Industrial.GetPattern() start="606050", stop="605060".
func TestCode2of5Industrial_Pattern_12345678(t *testing.T) {
	b := barcode.NewCode2of5IndustrialBarcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if !startsWith(pattern, "606050") {
		t.Errorf("pattern start = %q, want \"606050...\"", pattern[:min4(len(pattern), 6)])
	}
	if !endsWith(pattern, "605060") {
		t.Errorf("pattern end = %q, want \"...605060\"", pattern[max0(len(pattern)-6):])
	}
}

// ── 2/5 Matrix ────────────────────────────────────────────────────────────────

// TestCode2of5Matrix_GetWideBarRatio asserts WideBarRatio=2.25.
// C# Barcode2of5.cs:487 sets WideBarRatio=2.25F in the Matrix 2/5 constructor.
func TestCode2of5Matrix_GetWideBarRatio(t *testing.T) {
	b := barcode.NewCode2of5MatrixBarcode()
	got := b.GetWideBarRatio()
	if math.Abs(float64(got-2.25)) > 0.001 {
		t.Errorf("Code2of5Matrix GetWideBarRatio = %.4f, want 2.25 (C# Barcode2of5.cs:487)", got)
	}
}

// TestCode2of5Matrix_UpdateAutoSize_12345678 asserts final width=104.69px matching C#
// Barcode.frx HTML output (Barcode21, WideBarRatio=2.25). C# uses default text
// "12345678" with Barcode.CalcCheckSum="false".
// 8 digits × 8.5 module units + start(8.375) + stop(7.375) = 83.75 × 1.25 = 104.69px.
func TestCode2of5Matrix_UpdateAutoSize_12345678(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	mat := barcode.NewCode2of5MatrixBarcode()
	mat.CalcChecksum = false // C# Barcode.frx Barcode21 has CalcCheckSum=false
	obj.Barcode = mat
	if err := obj.Barcode.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 104.69
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode21)", got, wantW)
	}
}

// TestCode2of5Matrix_Pattern_12345678 validates bar pattern for "12345678".
// C# Barcode2of5Matrix.GetPattern() start="705050", stop="70505".
func TestCode2of5Matrix_Pattern_12345678(t *testing.T) {
	b := barcode.NewCode2of5MatrixBarcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if !startsWith(pattern, "705050") {
		t.Errorf("pattern start = %q, want \"705050...\"", pattern[:min4(len(pattern), 6)])
	}
	if !endsWith(pattern, "70505") {
		t.Errorf("pattern end = %q, want \"...70505\"", pattern[max0(len(pattern)-5):])
	}
}

// ── ITF-14 ────────────────────────────────────────────────────────────────────

// TestITF14_GetWideBarRatio asserts WideBarRatio=2.25.
// C# Barcode2of5.cs:566 sets WideBarRatio=2.25f in the ITF-14 constructor.
func TestITF14_GetWideBarRatio(t *testing.T) {
	b := barcode.NewITF14Barcode()
	got := b.GetWideBarRatio()
	if math.Abs(float64(got-2.25)) > 0.001 {
		t.Errorf("ITF14 GetWideBarRatio = %.4f, want 2.25 (C# Barcode2of5.cs:566)", got)
	}
}

// ── Deutsche Identcode ────────────────────────────────────────────────────────

// TestDeutscheIdentcode_GetWideBarRatio asserts WideBarRatio=3.
// C# Barcode2of5.cs:193 sets WideBarRatio=3F.
func TestDeutscheIdentcode_GetWideBarRatio(t *testing.T) {
	b := barcode.NewDeutscheIdentcodeBarcode()
	got := b.GetWideBarRatio()
	if got != 3.0 {
		t.Errorf("DeutscheIdentcode GetWideBarRatio = %.2f, want 3.0", got)
	}
}

// TestDeutscheIdentcode_UpdateAutoSize_12345123456 asserts final width=146.25px
// matching C# Barcode.frx HTML output (Barcode49).
func TestDeutscheIdentcode_UpdateAutoSize_12345123456(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewDeutscheIdentcodeBarcode()
	if err := obj.Barcode.Encode("12345123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 146.25
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode49)", got, wantW)
	}
}

// ── Deutsche Leitcode ─────────────────────────────────────────────────────────

// TestDeutscheLeitcode_GetWideBarRatio asserts WideBarRatio=3.
// C# Barcode2of5.cs:320 sets WideBarRatio=3F.
func TestDeutscheLeitcode_GetWideBarRatio(t *testing.T) {
	b := barcode.NewDeutscheLeitcodeBarcode()
	got := b.GetWideBarRatio()
	if got != 3.0 {
		t.Errorf("DeutscheLeitcode GetWideBarRatio = %.2f, want 3.0", got)
	}
}

// TestDeutscheLeitcode_UpdateAutoSize_1234512312312 asserts final width=168.75px
// matching C# Barcode.frx HTML output (Barcode50).
func TestDeutscheLeitcode_UpdateAutoSize_1234512312312(t *testing.T) {
	obj := barcode.NewBarcodeObject()
	obj.Barcode = barcode.NewDeutscheLeitcodeBarcode()
	if err := obj.Barcode.Encode("1234512312312"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	obj.UpdateAutoSize()
	const wantW = 168.75
	got := obj.Width()
	if math.Abs(float64(got-wantW)) > 1.0 {
		t.Errorf("UpdateAutoSize width = %.4f, want %.4f (C# Barcode.frx Barcode50)", got, wantW)
	}
}

// helpers used by pattern tests
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func min4(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}
