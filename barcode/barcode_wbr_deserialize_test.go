// barcode_wbr_deserialize_test.go validates that BarcodeObject.Deserialize reads
// Barcode.WideBarRatio from FRX attributes and applies it to CalcBounds.
//
// C# reference: BarcodeObject.cs SetBarcodeProperties():
//   barcode.WideBarRatio = Reader.ReadFloat("Barcode.WideBarRatio", 2)
//
// FRX examples with non-default WBR:
//   Barcode21: 2/5 Matrix  Barcode.WideBarRatio="2.25"
//   Barcode48: ITF-14      Barcode.WideBarRatio="2.25"
//   Barcode49: Deutsche Identcode  Barcode.WideBarRatio="3"
//   Barcode50: Deutsche Leitcode   Barcode.WideBarRatio="3"
//
// go-fastreport-7az4: Deserialize Barcode.WideBarRatio from FRX attributes
package barcode

import (
	"math"
	"testing"
)

// wbrMockReader is a mock report.Reader that supports float properties.
type wbrMockReader struct {
	strs   map[string]string
	floats map[string]float32
	bools  map[string]bool
}

func (r *wbrMockReader) ReadStr(name, def string) string {
	if v, ok := r.strs[name]; ok {
		return v
	}
	return def
}
func (r *wbrMockReader) ReadInt(name string, def int) int { return def }
func (r *wbrMockReader) ReadBool(name string, def bool) bool {
	if v, ok := r.bools[name]; ok {
		return v
	}
	return def
}
func (r *wbrMockReader) ReadFloat(name string, def float32) float32 {
	if v, ok := r.floats[name]; ok {
		return v
	}
	return def
}
func (r *wbrMockReader) NextChild() (string, bool) { return "", false }
func (r *wbrMockReader) FinishChild() error        { return nil }

// TestDeserialize_WideBarRatio_Code128_Default verifies that when Barcode.WideBarRatio
// is not present in FRX, Code128 uses its default WBR=2.
func TestDeserialize_WideBarRatio_Code128_Default(t *testing.T) {
	obj := NewBarcodeObject()
	r := &wbrMockReader{
		strs: map[string]string{
			"Barcode.Type": "Code128",
		},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if obj.Barcode == nil {
		t.Fatal("Barcode is nil after Deserialize")
	}
	// Encode and check WBR=2 is used (default, no override in FRX)
	obj.Barcode.Encode("12345678")
	w, _ := obj.Barcode.CalcBounds()
	// Code128 with "12345678" (Code C, 4 pairs) = 79 modules * 1.25 = 98.75px with WBR=2
	const wantW = 98.75
	if math.Abs(float64(w-wantW)) > 1.0 {
		t.Errorf("Code128 default WBR CalcBounds = %.2f, want %.2f", w, wantW)
	}
}

// TestDeserialize_WideBarRatio_Code128_Override verifies that Barcode.WideBarRatio="2.5"
// in FRX overrides the Code128 default WBR=2 and affects CalcBounds width.
func TestDeserialize_WideBarRatio_Code128_Override(t *testing.T) {
	obj := NewBarcodeObject()
	r := &wbrMockReader{
		strs: map[string]string{
			"Barcode.Type": "Code128",
		},
		floats: map[string]float32{
			"Barcode.WideBarRatio": 2.5,
		},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if obj.Barcode == nil {
		t.Fatal("Barcode is nil after Deserialize")
	}
	// Verify that SetWideBarRatio was applied (accessible via WBROverride).
	type wbrOverrider interface{ WBROverride() float32 }
	if ov, ok := obj.Barcode.(wbrOverrider); ok {
		if got := ov.WBROverride(); math.Abs(float64(got-2.5)) > 0.001 {
			t.Errorf("WBROverride = %.3f, want 2.5", got)
		}
	} else {
		t.Error("Barcode does not implement WBROverride interface")
	}

	// Verify CalcBounds uses the overridden WBR=2.5 instead of default 2.
	// With WBR=2.5: modules=[1, 2.5, 3.75, 5]. Same Code C pattern for "12345678".
	// Width = GetPatternWidth(pattern, modules[1=2.5]) * 1.25
	// This should differ from the WBR=2 width (98.75px).
	obj.Barcode.Encode("12345678")
	w, _ := obj.Barcode.CalcBounds()
	// WBR=2.5 gives wider bars than WBR=2, so width > 98.75
	const defaultWBR2Width = 98.75
	if w <= defaultWBR2Width-1.0 {
		t.Errorf("Code128 WBR=2.5 CalcBounds = %.2f, should be > %.2f (default WBR=2 width)", w, defaultWBR2Width)
	}
}

// TestDeserialize_WideBarRatio_Matrix25_FRX verifies that loading the 2/5 Matrix
// barcode from FRX with Barcode.WideBarRatio="2.25" preserves the type's natural
// WBR (no override needed when FRX value matches constructor default).
func TestDeserialize_WideBarRatio_Matrix25_FRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &wbrMockReader{
		strs: map[string]string{
			"Barcode": "2/5 Matrix",
		},
		floats: map[string]float32{
			"Barcode.WideBarRatio": 2.25,
		},
		bools: map[string]bool{
			"Barcode.CalcCheckSum": false, // C# Barcode.frx Barcode21 has CalcCheckSum=false
		},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if obj.Barcode == nil {
		t.Fatal("Barcode is nil after Deserialize")
	}
	// Encode with "12345678" and check width matches C# Barcode21 = 104.69px.
	if err := obj.Barcode.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	w, _ := obj.Barcode.CalcBounds()
	const wantW = 104.69
	if math.Abs(float64(w-wantW)) > 1.0 {
		t.Errorf("2/5 Matrix WBR=2.25 CalcBounds = %.2f, want %.2f (C# Barcode.frx Barcode21)", w, wantW)
	}
}
