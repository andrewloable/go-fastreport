package barcode_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// -----------------------------------------------------------------------
// BarcodeType constants
// -----------------------------------------------------------------------

func TestBarcodeTypeConstants(t *testing.T) {
	types := []barcode.BarcodeType{
		barcode.BarcodeTypeCode128,
		barcode.BarcodeTypeCode39,
		barcode.BarcodeTypeCode93,
		barcode.BarcodeTypeQR,
		barcode.BarcodeTypeDataMatrix,
		barcode.BarcodeTypeAztec,
	}
	seen := map[barcode.BarcodeType]bool{}
	for _, bt := range types {
		if seen[bt] {
			t.Errorf("duplicate BarcodeType %q", bt)
		}
		seen[bt] = true
	}
}

// -----------------------------------------------------------------------
// Code128Barcode
// -----------------------------------------------------------------------

func TestNewCode128Barcode(t *testing.T) {
	b := barcode.NewCode128Barcode()
	if b == nil {
		t.Fatal("NewCode128Barcode returned nil")
	}
	if b.Type() != barcode.BarcodeTypeCode128 {
		t.Errorf("Type = %q, want Code128", b.Type())
	}
}

func TestCode128Barcode_Encode(t *testing.T) {
	b := barcode.NewCode128Barcode()
	if err := b.Encode("HELLO123"); err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if b.EncodedText() != "HELLO123" {
		t.Errorf("EncodedText = %q, want HELLO123", b.EncodedText())
	}
}

func TestCode128Barcode_DefaultValue(t *testing.T) {
	b := barcode.NewCode128Barcode()
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

// -----------------------------------------------------------------------
// Code39Barcode
// -----------------------------------------------------------------------

func TestNewCode39Barcode(t *testing.T) {
	b := barcode.NewCode39Barcode()
	if b.Type() != barcode.BarcodeTypeCode39 {
		t.Errorf("Type = %q, want Code39", b.Type())
	}
	if b.AllowExtended {
		t.Error("AllowExtended should default to false")
	}
	if b.CalcChecksum {
		t.Error("CalcChecksum should default to false")
	}
}

func TestCode39Barcode_Encode(t *testing.T) {
	b := barcode.NewCode39Barcode()
	if err := b.Encode("CODE39"); err != nil {
		t.Fatalf("Encode error: %v", err)
	}
}

func TestCode39Barcode_Fields(t *testing.T) {
	b := barcode.NewCode39Barcode()
	b.AllowExtended = true
	b.CalcChecksum = true
	if !b.AllowExtended || !b.CalcChecksum {
		t.Error("Code39 fields should be settable")
	}
}

// -----------------------------------------------------------------------
// QRBarcode
// -----------------------------------------------------------------------

func TestNewQRBarcode(t *testing.T) {
	q := barcode.NewQRBarcode()
	if q == nil {
		t.Fatal("NewQRBarcode returned nil")
	}
	if q.Type() != barcode.BarcodeTypeQR {
		t.Errorf("Type = %q, want QR", q.Type())
	}
	if q.ErrorCorrection != "M" {
		t.Errorf("ErrorCorrection default = %q, want M", q.ErrorCorrection)
	}
}

func TestQRBarcode_Encode(t *testing.T) {
	q := barcode.NewQRBarcode()
	if err := q.Encode("https://example.com"); err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if q.EncodedText() != "https://example.com" {
		t.Errorf("EncodedText = %q", q.EncodedText())
	}
}

// -----------------------------------------------------------------------
// NewBarcodeByType
// -----------------------------------------------------------------------

func TestNewBarcodeByType_Code128(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeCode128)
	if b.Type() != barcode.BarcodeTypeCode128 {
		t.Errorf("Type = %q, want Code128", b.Type())
	}
}

func TestNewBarcodeByType_Code39(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeCode39)
	if b.Type() != barcode.BarcodeTypeCode39 {
		t.Errorf("Type = %q, want Code39", b.Type())
	}
}

func TestNewBarcodeByType_QR(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeQR)
	if b.Type() != barcode.BarcodeTypeQR {
		t.Errorf("Type = %q, want QR", b.Type())
	}
}

func TestNewBarcodeByType_Unknown(t *testing.T) {
	b := barcode.NewBarcodeByType("Unknown")
	// falls back to Code128
	if b.Type() != barcode.BarcodeTypeCode128 {
		t.Errorf("Unknown type falls back to Code128, got %q", b.Type())
	}
}

// -----------------------------------------------------------------------
// BarcodeObject
// -----------------------------------------------------------------------

func TestNewBarcodeObject_Defaults(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	if bo == nil {
		t.Fatal("NewBarcodeObject returned nil")
	}
	if bo.Barcode == nil {
		t.Error("Barcode should default to non-nil (Code128)")
	}
	if bo.BarcodeType() != barcode.BarcodeTypeCode128 {
		t.Errorf("BarcodeType default = %q, want Code128", bo.BarcodeType())
	}
	if bo.Angle() != 0 {
		t.Errorf("Angle default = %d, want 0", bo.Angle())
	}
	if bo.AutoSize() {
		t.Error("AutoSize should default to false")
	}
	if !bo.ShowText() {
		t.Error("ShowText should default to true")
	}
	if bo.Zoom() != 1.0 {
		t.Errorf("Zoom default = %v, want 1.0", bo.Zoom())
	}
	if !bo.AllowExpressions() {
		t.Error("AllowExpressions should default to true")
	}
	if bo.Brackets() != "[,]" {
		t.Errorf("Brackets default = %q, want [,]", bo.Brackets())
	}
	if bo.HideIfNoData() {
		t.Error("HideIfNoData should default to false")
	}
}

func TestBarcodeObject_Angle(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetAngle(90)
	if bo.Angle() != 90 {
		t.Errorf("Angle = %d, want 90", bo.Angle())
	}
}

func TestBarcodeObject_AutoSize(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetAutoSize(true)
	if !bo.AutoSize() {
		t.Error("AutoSize should be true")
	}
}

func TestBarcodeObject_Text(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetText("1234567890")
	if bo.Text() != "1234567890" {
		t.Errorf("Text = %q", bo.Text())
	}
}

func TestBarcodeObject_DataColumn(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetDataColumn("SKU")
	if bo.DataColumn() != "SKU" {
		t.Errorf("DataColumn = %q, want SKU", bo.DataColumn())
	}
}

func TestBarcodeObject_Expression(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetExpression("[Product.Barcode]")
	if bo.Expression() != "[Product.Barcode]" {
		t.Errorf("Expression = %q", bo.Expression())
	}
}

func TestBarcodeObject_ShowText(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetShowText(false)
	if bo.ShowText() {
		t.Error("ShowText should be false")
	}
}

func TestBarcodeObject_Zoom(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetZoom(2.0)
	if bo.Zoom() != 2.0 {
		t.Errorf("Zoom = %v, want 2.0", bo.Zoom())
	}
}

func TestBarcodeObject_HideIfNoData(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetHideIfNoData(true)
	if !bo.HideIfNoData() {
		t.Error("HideIfNoData should be true")
	}
}

func TestBarcodeObject_NoDataText(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetNoDataText("N/A")
	if bo.NoDataText() != "N/A" {
		t.Errorf("NoDataText = %q, want N/A", bo.NoDataText())
	}
}

func TestBarcodeObject_Padding(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	p := barcode.Padding{Left: 5, Top: 5, Right: 5, Bottom: 5}
	bo.SetPadding(p)
	if bo.Padding() != p {
		t.Errorf("Padding = %+v, want %+v", bo.Padding(), p)
	}
}

func TestBarcodeObject_SwitchBarcode(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.Barcode = barcode.NewQRBarcode()
	if bo.BarcodeType() != barcode.BarcodeTypeQR {
		t.Errorf("BarcodeType after switch = %q, want QR", bo.BarcodeType())
	}
}

func TestBarcodeObject_InheritsVisible(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	if !bo.Visible() {
		t.Error("BarcodeObject should inherit Visible=true")
	}
}
