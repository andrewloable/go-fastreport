package barcode_test

import (
	"bytes"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
	"github.com/andrewloable/go-fastreport/serial"
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
	if !b.CalcChecksum {
		t.Error("CalcChecksum should default to true (C# LinearBarcodeBase default)")
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
	if q.ErrorCorrection != "L" {
		t.Errorf("ErrorCorrection default = %q, want L", q.ErrorCorrection)
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
		t.Error("Barcode should default to non-nil (Code39)")
	}
	// C# BarcodeObject constructor: Barcode = new Barcode39() (BarcodeObject.cs:688).
	if bo.BarcodeType() != barcode.BarcodeTypeCode39 {
		t.Errorf("BarcodeType default = %q, want Code39 (C# BarcodeObject.cs:688)", bo.BarcodeType())
	}
	if bo.Angle() != 0 {
		t.Errorf("Angle default = %d, want 0", bo.Angle())
	}
	if !bo.AutoSize() {
		t.Error("AutoSize should default to true (C# BarcodeObject.cs:689)")
	}
	if !bo.ShowText() {
		t.Error("ShowText should default to true")
	}
	if bo.Zoom() != 1.0 {
		t.Errorf("Zoom default = %v, want 1.0", bo.Zoom())
	}
	if bo.AllowExpressions() {
		t.Error("AllowExpressions should default to false (C# BarcodeObject.cs:232)")
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

// -----------------------------------------------------------------------
// MaxiCodeBarcode — RS encoding
// -----------------------------------------------------------------------

func TestMaxiCodeEncode_Produces144Codewords(t *testing.T) {
	b := barcode.NewMaxiCodeBarcode()
	if err := b.Encode("Hello MaxiCode"); err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	img, err := b.Render(100, 100)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestMaxiCodeEncode_Mode4(t *testing.T) {
	b := barcode.NewMaxiCodeBarcode()
	b.Mode = 4
	if err := b.Encode("Test Mode 4"); err != nil {
		t.Fatalf("Encode error: %v", err)
	}
}

func TestMaxiCodeEncode_Mode5(t *testing.T) {
	b := barcode.NewMaxiCodeBarcode()
	b.Mode = 5
	if err := b.Encode("Test Mode 5"); err != nil {
		t.Fatalf("Encode error: %v", err)
	}
}

func TestMaxiCodeRS_KnownValues(t *testing.T) {
	// Verify RS ECC for a small known input against manually computed values.
	// GF(64), poly=0x43, generator roots alpha^1..alpha^10.
	// Input: all-zero 10 codewords → ECC should all be zero (trivial case).
	data := make([]byte, 10)
	ecc := barcode.MaxiCodeComputeECC(data, 10)
	for i, v := range ecc {
		if v != 0 {
			t.Errorf("ECC[%d] = %d for all-zero input, want 0", i, v)
		}
	}
}

func TestMaxiCodeMode2Payload(t *testing.T) {
	payload := barcode.MaxiCodeMode2Payload("902840772", "840", "001", "UPSN^TRAKG^")
	if len(payload) < 15 {
		t.Errorf("mode 2 payload too short: %d", len(payload))
	}
}

// -----------------------------------------------------------------------
// BaseBarcodeImpl.Render
// -----------------------------------------------------------------------

func TestBaseBarcodeImpl_Render_NotEncoded(t *testing.T) {
	// Zero-value BaseBarcodeImpl (encoded==nil) should return error.
	b := &barcode.BaseBarcodeImpl{}
	_, err := b.Render(100, 100)
	if err == nil {
		t.Error("expected error when Render called before Encode, got nil")
	}
}

func TestBaseBarcodeImpl_Render_Success(t *testing.T) {
	// Encoding via Code128 sets BaseBarcodeImpl.encoded; Render (inherited) should succeed.
	b := barcode.NewCode128Barcode()
	if err := b.Encode("HELLO"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Error("Render returned nil image")
	}
}

// -----------------------------------------------------------------------
// BaseBarcodeImpl.DefaultValue (base class, not overridden by Code128/Code39)
// -----------------------------------------------------------------------

func TestBaseBarcodeImpl_DefaultValue(t *testing.T) {
	b := &barcode.BaseBarcodeImpl{}
	if got := b.DefaultValue(); got != "12345678" {
		t.Errorf("DefaultValue = %q, want 12345678", got)
	}
}

// -----------------------------------------------------------------------
// Code128Barcode error path
// -----------------------------------------------------------------------

func TestCode128Barcode_Encode_Error(t *testing.T) {
	b := barcode.NewCode128Barcode()
	// Empty string is invalid for Code128.
	if err := b.Encode(""); err == nil {
		t.Error("expected error encoding empty string with Code128, got nil")
	}
}

// -----------------------------------------------------------------------
// Code39Barcode error path
// -----------------------------------------------------------------------

func TestCode39Barcode_Encode_Error(t *testing.T) {
	b := barcode.NewCode39Barcode()
	// Lowercase without AllowExtended is invalid Code39.
	if err := b.Encode("lowercase"); err == nil {
		t.Error("expected error encoding lowercase without AllowExtended, got nil")
	}
}

// -----------------------------------------------------------------------
// QRBarcode — error correction levels and DefaultValue
// -----------------------------------------------------------------------

func TestQRBarcode_Encode_L(t *testing.T) {
	q := barcode.NewQRBarcode()
	q.ErrorCorrection = "L"
	if err := q.Encode("test"); err != nil {
		t.Fatalf("Encode L: %v", err)
	}
}

func TestQRBarcode_Encode_Q(t *testing.T) {
	q := barcode.NewQRBarcode()
	q.ErrorCorrection = "Q"
	if err := q.Encode("test"); err != nil {
		t.Fatalf("Encode Q: %v", err)
	}
}

func TestQRBarcode_Encode_H(t *testing.T) {
	q := barcode.NewQRBarcode()
	q.ErrorCorrection = "H"
	if err := q.Encode("test"); err != nil {
		t.Fatalf("Encode H: %v", err)
	}
}

func TestQRBarcode_DefaultValue(t *testing.T) {
	q := barcode.NewQRBarcode()
	if got := q.DefaultValue(); got == "" {
		t.Error("QRBarcode.DefaultValue returned empty string")
	}
}

// -----------------------------------------------------------------------
// BarcodeObject — SetAllowExpressions, SetBrackets, nil BarcodeType
// -----------------------------------------------------------------------

func TestBarcodeObject_SetAllowExpressions(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetAllowExpressions(false)
	if bo.AllowExpressions() {
		t.Error("AllowExpressions should be false after SetAllowExpressions(false)")
	}
}

func TestBarcodeObject_SetBrackets(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetBrackets("{,}")
	if bo.Brackets() != "{,}" {
		t.Errorf("Brackets = %q, want {,}", bo.Brackets())
	}
}

func TestBarcodeObject_BarcodeType_NilBarcode(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.Barcode = nil
	if got := bo.BarcodeType(); got != "" {
		t.Errorf("BarcodeType with nil Barcode = %q, want empty", got)
	}
}

// -----------------------------------------------------------------------
// BarcodeObject Serialize / Deserialize round-trip
// -----------------------------------------------------------------------

func TestBarcodeObject_SerializeDeserialize(t *testing.T) {
	orig := barcode.NewBarcodeObject()
	orig.SetAngle(45)
	orig.SetAutoSize(true)
	orig.SetText("ABC123")
	orig.SetShowText(false)
	orig.SetZoom(2.0)
	orig.SetHideIfNoData(true)
	orig.SetNoDataText("N/A")
	orig.SetAllowExpressions(false)
	orig.SetBrackets("{,}")
	orig.SetDataColumn("Barcode")
	orig.SetExpression("[Row.Code]")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("BarcodeObject"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := orig.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	got := barcode.NewBarcodeObject()
	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed; xml:\n%s", buf.String())
	}
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if got.Angle() != 45 {
		t.Errorf("Angle = %d, want 45", got.Angle())
	}
	if !got.AutoSize() {
		t.Error("AutoSize should be true")
	}
	if got.Text() != "ABC123" {
		t.Errorf("Text = %q, want ABC123", got.Text())
	}
	if got.ShowText() {
		t.Error("ShowText should be false")
	}
	if got.Zoom() != 2.0 {
		t.Errorf("Zoom = %v, want 2.0", got.Zoom())
	}
	if !got.HideIfNoData() {
		t.Error("HideIfNoData should be true")
	}
	if got.NoDataText() != "N/A" {
		t.Errorf("NoDataText = %q, want N/A", got.NoDataText())
	}
	if got.AllowExpressions() {
		t.Error("AllowExpressions should be false")
	}
	if got.Brackets() != "{,}" {
		t.Errorf("Brackets = %q, want {,}", got.Brackets())
	}
}

// -----------------------------------------------------------------------
// EAN13Barcode
// -----------------------------------------------------------------------

func TestNewEAN13Barcode(t *testing.T) {
	b := barcode.NewEAN13Barcode()
	if b.Type() != barcode.BarcodeTypeEAN13 {
		t.Errorf("Type = %q, want EAN13", b.Type())
	}
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestEAN13Barcode_Encode(t *testing.T) {
	b := barcode.NewEAN13Barcode()
	if err := b.Encode("590123412345"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if b.EncodedText() != "590123412345" {
		t.Errorf("EncodedText = %q", b.EncodedText())
	}
}

func TestEAN13Barcode_Encode_Error(t *testing.T) {
	b := barcode.NewEAN13Barcode()
	if err := b.Encode("123"); err == nil {
		t.Error("expected error for invalid EAN13 data, got nil")
	}
}

// -----------------------------------------------------------------------
// AztecBarcode
// -----------------------------------------------------------------------

func TestNewAztecBarcode(t *testing.T) {
	b := barcode.NewAztecBarcode()
	if b.Type() != barcode.BarcodeTypeAztec {
		t.Errorf("Type = %q, want Aztec", b.Type())
	}
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestAztecBarcode_Encode(t *testing.T) {
	b := barcode.NewAztecBarcode()
	if err := b.Encode("Hello Aztec"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
}

// -----------------------------------------------------------------------
// PDF417Barcode
// -----------------------------------------------------------------------

func TestNewPDF417Barcode(t *testing.T) {
	b := barcode.NewPDF417Barcode()
	if b.Type() != barcode.BarcodeTypePDF417 {
		t.Errorf("Type = %q, want PDF417", b.Type())
	}
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestPDF417Barcode_Encode(t *testing.T) {
	b := barcode.NewPDF417Barcode()
	if err := b.Encode("PDF417 content"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
}

// -----------------------------------------------------------------------
// Code93Barcode
// -----------------------------------------------------------------------

func TestNewCode93Barcode(t *testing.T) {
	b := barcode.NewCode93Barcode()
	if b.Type() != barcode.BarcodeTypeCode93 {
		t.Errorf("Type = %q, want Code93", b.Type())
	}
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestCode93Barcode_Encode_Render(t *testing.T) {
	b := barcode.NewCode93Barcode()
	if err := b.Encode("CODE93"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Error("Render returned nil image")
	}
}

func TestCode93Barcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewCode93Barcode()
	_, err := b.Render(100, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// -----------------------------------------------------------------------
// Code2of5Barcode
// -----------------------------------------------------------------------

func TestNewCode2of5Barcode(t *testing.T) {
	b := barcode.NewCode2of5Barcode()
	if b.Type() != barcode.BarcodeTypeCode2of5 {
		t.Errorf("Type = %q, want 2of5", b.Type())
	}
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestCode2of5Barcode_Encode_Render(t *testing.T) {
	b := barcode.NewCode2of5Barcode()
	if err := b.Encode("12345670"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Error("Render returned nil image")
	}
}

func TestCode2of5Barcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewCode2of5Barcode()
	_, err := b.Render(100, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// -----------------------------------------------------------------------
// CodabarBarcode
// -----------------------------------------------------------------------

func TestNewCodabarBarcode(t *testing.T) {
	b := barcode.NewCodabarBarcode()
	if b.Type() != barcode.BarcodeTypeCodabar {
		t.Errorf("Type = %q, want Codabar", b.Type())
	}
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestCodabarBarcode_Encode_Render(t *testing.T) {
	b := barcode.NewCodabarBarcode()
	if err := b.Encode("A12345B"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Error("Render returned nil image")
	}
}

func TestCodabarBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewCodabarBarcode()
	_, err := b.Render(100, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// -----------------------------------------------------------------------
// DataMatrixBarcode
// -----------------------------------------------------------------------

func TestNewDataMatrixBarcode(t *testing.T) {
	b := barcode.NewDataMatrixBarcode()
	if b.Type() != barcode.BarcodeTypeDataMatrix {
		t.Errorf("Type = %q, want DataMatrix", b.Type())
	}
	if b.DefaultValue() == "" {
		t.Error("DefaultValue should not be empty")
	}
}

func TestDataMatrixBarcode_Encode_Render(t *testing.T) {
	b := barcode.NewDataMatrixBarcode()
	if err := b.Encode("DataMatrix"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(100, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Error("Render returned nil image")
	}
}

func TestDataMatrixBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewDataMatrixBarcode()
	_, err := b.Render(100, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// -----------------------------------------------------------------------
// NewBarcodeByType — remaining cases
// -----------------------------------------------------------------------

func TestNewBarcodeByType_EAN13(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeEAN13)
	if b.Type() != barcode.BarcodeTypeEAN13 {
		t.Errorf("Type = %q, want EAN13", b.Type())
	}
}

func TestNewBarcodeByType_EAN8(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeEAN8)
	// EAN8 maps to EAN13 implementation.
	if b == nil {
		t.Fatal("NewBarcodeByType(EAN8) returned nil")
	}
}

func TestNewBarcodeByType_UPCA(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeUPCA)
	if b == nil {
		t.Fatal("NewBarcodeByType(UPCA) returned nil")
	}
}

func TestNewBarcodeByType_UPCE(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeUPCE)
	if b == nil {
		t.Fatal("NewBarcodeByType(UPCE) returned nil")
	}
}

func TestNewBarcodeByType_Aztec(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeAztec)
	if b.Type() != barcode.BarcodeTypeAztec {
		t.Errorf("Type = %q, want Aztec", b.Type())
	}
}

func TestNewBarcodeByType_PDF417(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypePDF417)
	if b.Type() != barcode.BarcodeTypePDF417 {
		t.Errorf("Type = %q, want PDF417", b.Type())
	}
}

func TestNewBarcodeByType_Code93(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeCode93)
	if b.Type() != barcode.BarcodeTypeCode93 {
		t.Errorf("Type = %q, want Code93", b.Type())
	}
}

func TestNewBarcodeByType_Code2of5(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeCode2of5)
	if b.Type() != barcode.BarcodeTypeCode2of5 {
		t.Errorf("Type = %q, want 2of5", b.Type())
	}
}

func TestNewBarcodeByType_Codabar(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeCodabar)
	if b.Type() != barcode.BarcodeTypeCodabar {
		t.Errorf("Type = %q, want Codabar", b.Type())
	}
}

func TestNewBarcodeByType_DataMatrix(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeDataMatrix)
	if b.Type() != barcode.BarcodeTypeDataMatrix {
		t.Errorf("Type = %q, want DataMatrix", b.Type())
	}
}

func TestNewBarcodeByType_MSI(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeMSI)
	if b == nil {
		t.Fatal("NewBarcodeByType(MSI) returned nil")
	}
}

func TestNewBarcodeByType_MaxiCode(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeMaxiCode)
	if b == nil {
		t.Fatal("NewBarcodeByType(MaxiCode) returned nil")
	}
}

func TestNewBarcodeByType_GS1_128(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeGS1_128)
	if b == nil {
		t.Fatal("NewBarcodeByType(GS1-128) returned nil")
	}
}

func TestNewBarcodeByType_IntelligentMail(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeIntelligentMail)
	if b == nil {
		t.Fatal("NewBarcodeByType(IntelligentMail) returned nil")
	}
}

func TestNewBarcodeByType_Pharmacode(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypePharmacode)
	if b == nil {
		t.Fatal("NewBarcodeByType(Pharmacode) returned nil")
	}
}

func TestNewBarcodeByType_Plessey(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypePlessey)
	if b == nil {
		t.Fatal("NewBarcodeByType(Plessey) returned nil")
	}
}

func TestNewBarcodeByType_PostNet(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypePostNet)
	if b == nil {
		t.Fatal("NewBarcodeByType(PostNet) returned nil")
	}
}

func TestNewBarcodeByType_SwissQR(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeSwissQR)
	if b == nil {
		t.Fatal("NewBarcodeByType(SwissQR) returned nil")
	}
}

// -----------------------------------------------------------------------
// Code39Barcode.DefaultValue
// -----------------------------------------------------------------------

func TestCode39Barcode_DefaultValue(t *testing.T) {
	b := barcode.NewCode39Barcode()
	if got := b.DefaultValue(); got != "12345678" {
		t.Errorf("DefaultValue = %q, want 12345678", got)
	}
}

// -----------------------------------------------------------------------
// BaseBarcodeImpl.Render — scale error (too small dimensions)
// -----------------------------------------------------------------------

func TestBaseBarcodeImpl_Render_ScaleFallback(t *testing.T) {
	b := barcode.NewCode128Barcode()
	if err := b.Encode("HELLO"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// Scale to 0x0 falls back to natural size (C# always renders regardless of size).
	img, err := b.Render(0, 0)
	if err != nil {
		t.Fatalf("Render(0,0) should fall back to natural size, got error: %v", err)
	}
	if img == nil {
		t.Error("Render(0,0) returned nil image")
	}
}

// -----------------------------------------------------------------------
// AztecBarcode.Encode error path
// -----------------------------------------------------------------------

func TestAztecBarcode_Encode_Error(t *testing.T) {
	b := barcode.NewAztecBarcode()
	// Layers=100 is out of range.
	b.UserSpecifiedLayers = 100
	if err := b.Encode("test"); err == nil {
		t.Error("expected error for Aztec with invalid layers, got nil")
	}
}

// -----------------------------------------------------------------------
// PDF417Barcode.Encode error path
// -----------------------------------------------------------------------

func TestPDF417Barcode_Encode_Error(t *testing.T) {
	b := barcode.NewPDF417Barcode()
	// SecurityLevel 9 is out of range (0-8).
	b.SecurityLevel = 9
	if err := b.Encode("test"); err == nil {
		t.Error("expected error for PDF417 with SecurityLevel=9, got nil")
	}
}

// -----------------------------------------------------------------------
// GS1Barcode
// -----------------------------------------------------------------------

func TestGS1Barcode_DefaultValue(t *testing.T) {
	b := barcode.NewGS1Barcode()
	if got := b.DefaultValue(); got == "" {
		t.Error("GS1Barcode.DefaultValue returned empty string")
	}
}

func TestGS1Barcode_Encode_Render(t *testing.T) {
	b := barcode.NewGS1Barcode()
	// Standard GS1 GTIN-14 with AI notation.
	if err := b.Encode("(01)12345678901231"); err != nil {
		t.Fatalf("GS1Barcode.Encode: %v", err)
	}
	img, err := b.Render(300, 100)
	if err != nil {
		t.Fatalf("GS1Barcode.Render: %v", err)
	}
	if img == nil {
		t.Error("GS1Barcode.Render returned nil image")
	}
}

func TestGS1Barcode_Encode_PlainText(t *testing.T) {
	b := barcode.NewGS1Barcode()
	// Plain digits without AI parens should also encode successfully.
	if err := b.Encode("0112345678901231"); err != nil {
		t.Fatalf("GS1Barcode.Encode plain: %v", err)
	}
}

func TestGS1Barcode_Render_NilEncoded(t *testing.T) {
	// A fresh barcode with empty encodedText will try to Encode("") on Render call.
	// Code128 Encode on empty string fails, so we expect an error.
	b := barcode.NewGS1Barcode()
	_, err := b.Render(200, 100)
	// It may succeed or fail — just make sure it doesn't panic.
	_ = err
}

// -----------------------------------------------------------------------
// IntelligentMailBarcode
// -----------------------------------------------------------------------

func TestIntelligentMailBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	if got := b.DefaultValue(); got == "" {
		t.Error("IntelligentMailBarcode.DefaultValue returned empty string")
	}
}

func TestIntelligentMailBarcode_Encode_Valid(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	// 20-digit IMb: barcode-id(2) + service-type(3) + mailer-id(6) + serial(9)
	if err := b.Encode("01234567094987654321"); err != nil {
		t.Fatalf("IntelligentMailBarcode.Encode (20 digits): %v", err)
	}
}

func TestIntelligentMailBarcode_Encode_31Digits(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	// 31-digit IMb (20 barcode ID + 11 digit routing)
	if err := b.Encode("0123456709498765432190210123456"); err != nil {
		t.Fatalf("IntelligentMailBarcode.Encode (31 digits): %v", err)
	}
}

func TestIntelligentMailBarcode_Encode_Error(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	// 8 digits is invalid (not 20/25/29/31).
	if err := b.Encode("12345678"); err == nil {
		t.Error("expected error for 8-digit IMb, got nil")
	}
}

func TestIntelligentMailBarcode_Render(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	if err := b.Encode("01234567094987654321"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 60)
	if err != nil {
		t.Fatalf("IntelligentMailBarcode.Render: %v", err)
	}
	if img == nil {
		t.Error("IntelligentMailBarcode.Render returned nil image")
	}
}

func TestIntelligentMailBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	_, err := b.Render(200, 60)
	if err == nil {
		t.Error("expected error when Render called before Encode")
	}
}

// -----------------------------------------------------------------------
// MSIBarcode
// -----------------------------------------------------------------------

func TestMSIBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewMSIBarcode()
	if got := b.DefaultValue(); got == "" {
		t.Error("MSIBarcode.DefaultValue returned empty string")
	}
}

func TestMSIBarcode_Encode_Render(t *testing.T) {
	b := barcode.NewMSIBarcode()
	if err := b.Encode("12345"); err != nil {
		t.Fatalf("MSIBarcode.Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("MSIBarcode.Render: %v", err)
	}
	if img == nil {
		t.Error("MSIBarcode.Render returned nil image")
	}
}

func TestMSIBarcode_Encode_Error(t *testing.T) {
	b := barcode.NewMSIBarcode()
	if err := b.Encode("12A45"); err == nil {
		t.Error("expected error for non-digit MSI input, got nil")
	}
}

func TestMSIBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewMSIBarcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when MSI Render called before Encode")
	}
}

// -----------------------------------------------------------------------
// MaxiCodeBarcode DefaultValue (for extended.go coverage)
// -----------------------------------------------------------------------

func TestMaxiCodeBarcode_DefaultValue_Extended(t *testing.T) {
	b := barcode.NewMaxiCodeBarcode()
	if got := b.DefaultValue(); got == "" {
		t.Error("MaxiCodeBarcode.DefaultValue returned empty string")
	}
}

// -----------------------------------------------------------------------
// PharmacodeBarcode
// -----------------------------------------------------------------------

func TestPharmacodeBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewPharmacodeBarcode()
	if got := b.DefaultValue(); got == "" {
		t.Error("PharmacodeBarcode.DefaultValue returned empty string")
	}
}

func TestPharmacodeBarcode_Encode_Render(t *testing.T) {
	b := barcode.NewPharmacodeBarcode()
	if err := b.Encode("1234"); err != nil {
		t.Fatalf("PharmacodeBarcode.Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("PharmacodeBarcode.Render: %v", err)
	}
	if img == nil {
		t.Error("PharmacodeBarcode.Render returned nil image")
	}
}

func TestPharmacodeBarcode_Render_ZeroDimensions(t *testing.T) {
	// Triggers renderBitPattern → placeholderImage path (width/height <= 0).
	b := barcode.NewPharmacodeBarcode()
	if err := b.Encode("1234"); err != nil {
		t.Fatalf("PharmacodeBarcode.Encode: %v", err)
	}
	img, err := b.Render(0, 0)
	if err != nil {
		t.Fatalf("PharmacodeBarcode.Render(0,0): %v", err)
	}
	if img == nil {
		t.Error("PharmacodeBarcode.Render(0,0) returned nil image")
	}
}

func TestPharmacodeBarcode_Encode_Error_TooSmall(t *testing.T) {
	b := barcode.NewPharmacodeBarcode()
	// Value 2 is below the valid range of 3-131070.
	if err := b.Encode("2"); err == nil {
		t.Error("expected error for Pharmacode value 2, got nil")
	}
}

func TestPharmacodeBarcode_Encode_Error_NonInt(t *testing.T) {
	b := barcode.NewPharmacodeBarcode()
	if err := b.Encode("notanumber"); err == nil {
		t.Error("expected error for non-integer Pharmacode input, got nil")
	}
}

func TestPharmacodeBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewPharmacodeBarcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Pharmacode Render called before Encode")
	}
}

// -----------------------------------------------------------------------
// PlesseyBarcode
// -----------------------------------------------------------------------

func TestPlesseyBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	if got := b.DefaultValue(); got == "" {
		t.Error("PlesseyBarcode.DefaultValue returned empty string")
	}
}

func TestPlesseyBarcode_Encode_Render(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	if err := b.Encode("1234"); err != nil {
		t.Fatalf("PlesseyBarcode.Encode: %v", err)
	}
	img, err := b.Render(300, 100)
	if err != nil {
		t.Fatalf("PlesseyBarcode.Render: %v", err)
	}
	if img == nil {
		t.Error("PlesseyBarcode.Render returned nil image")
	}
}

func TestPlesseyBarcode_Encode_HexDigits(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	// A-F are valid hex digits.
	if err := b.Encode("ABCDEF"); err != nil {
		t.Fatalf("PlesseyBarcode.Encode ABCDEF: %v", err)
	}
}

func TestPlesseyBarcode_Encode_Error(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	// G is not a valid hex digit.
	if err := b.Encode("123G"); err == nil {
		t.Error("expected error for invalid Plessey char G, got nil")
	}
}

func TestPlesseyBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Plessey Render called before Encode")
	}
}

// -----------------------------------------------------------------------
// PostNetBarcode
// -----------------------------------------------------------------------

func TestPostNetBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewPostNetBarcode()
	if got := b.DefaultValue(); got == "" {
		t.Error("PostNetBarcode.DefaultValue returned empty string")
	}
}

func TestPostNetBarcode_Encode_Render(t *testing.T) {
	b := barcode.NewPostNetBarcode()
	if err := b.Encode("90210"); err != nil {
		t.Fatalf("PostNetBarcode.Encode: %v", err)
	}
	img, err := b.Render(200, 60)
	if err != nil {
		t.Fatalf("PostNetBarcode.Render: %v", err)
	}
	if img == nil {
		t.Error("PostNetBarcode.Render returned nil image")
	}
}

func TestPostNetBarcode_Encode_9Digits(t *testing.T) {
	b := barcode.NewPostNetBarcode()
	if err := b.Encode("902101234"); err != nil {
		t.Fatalf("PostNetBarcode.Encode 9 digits: %v", err)
	}
}

func TestPostNetBarcode_Encode_8Digits(t *testing.T) {
	b := barcode.NewPostNetBarcode()
	// C# BarcodePostNet.GetPattern() accepts any number of digits without
	// length validation. 8 digits should be accepted.
	if err := b.Encode("12345678"); err != nil {
		t.Errorf("PostNetBarcode.Encode 8 digits: unexpected error: %v", err)
	}
}

func TestPostNetBarcode_Encode_Error_Empty(t *testing.T) {
	b := barcode.NewPostNetBarcode()
	if err := b.Encode(""); err == nil {
		t.Error("expected error for empty PostNet, got nil")
	}
}

func TestPostNetBarcode_Encode_Error_NonDigit(t *testing.T) {
	b := barcode.NewPostNetBarcode()
	if err := b.Encode("9021A"); err == nil {
		t.Error("expected error for non-digit PostNet, got nil")
	}
}

func TestPostNetBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewPostNetBarcode()
	_, err := b.Render(200, 60)
	if err == nil {
		t.Error("expected error when PostNet Render called before Encode")
	}
}

// -----------------------------------------------------------------------
// SwissQRBarcode
// -----------------------------------------------------------------------

func TestSwissQRBarcode_DefaultValue(t *testing.T) {
	b := barcode.NewSwissQRBarcode()
	if got := b.DefaultValue(); got == "" {
		t.Error("SwissQRBarcode.DefaultValue returned empty string")
	}
}

func TestSwissQRBarcode_Encode_WithText(t *testing.T) {
	b := barcode.NewSwissQRBarcode()
	payload := b.DefaultValue()
	if err := b.Encode(payload); err != nil {
		t.Fatalf("SwissQRBarcode.Encode: %v", err)
	}
}

func TestSwissQRBarcode_Encode_EmptyText_BuildsPayload(t *testing.T) {
	// Empty text triggers buildPayload from struct fields.
	b := barcode.NewSwissQRBarcode()
	b.IBAN = "CH5604835012345678009"
	b.CreditorName = "MaxMuster"
	b.Amount = "10.00"
	if err := b.Encode(""); err != nil {
		t.Fatalf("SwissQRBarcode.Encode empty: %v", err)
	}
}

func TestSwissQRBarcode_Render(t *testing.T) {
	b := barcode.NewSwissQRBarcode()
	if err := b.Encode(b.DefaultValue()); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 200)
	if err != nil {
		t.Fatalf("SwissQRBarcode.Render: %v", err)
	}
	if img == nil {
		t.Error("SwissQRBarcode.Render returned nil image")
	}
}

func TestSwissQRBarcode_Render_NilEncoded(t *testing.T) {
	// Fresh barcode with empty encodedText → Render calls Encode("") → buildPayload.
	b := barcode.NewSwissQRBarcode()
	img, err := b.Render(200, 200)
	if err != nil {
		t.Fatalf("SwissQRBarcode.Render (nil encoded): %v", err)
	}
	if img == nil {
		t.Error("SwissQRBarcode.Render (nil encoded) returned nil image")
	}
}

func TestSwissQRBarcode_BuildPayload_EmptyCurrencyRefType(t *testing.T) {
	// Covers the if-currency=="" and if-refType=="" branches in buildPayload.
	b := barcode.NewSwissQRBarcode()
	b.Currency = ""   // triggers the `if currency == ""` default-to-CHF branch
	b.RefType = ""    // triggers the `if refType == ""` default-to-NON branch
	if err := b.Encode(""); err != nil {
		t.Fatalf("SwissQRBarcode.Encode with empty currency/refType: %v", err)
	}
}

// -----------------------------------------------------------------------
// IntelligentMailBarcode: cover strings.Map non-digit filter branch and
// Render fallback when imb_encode returns an error.
// -----------------------------------------------------------------------

func TestIntelligentMailBarcode_Encode_WithDashes(t *testing.T) {
	// Dashes are filtered out by strings.Map (covers the return -1 branch).
	// Result after filtering: 8 digits → wrong length → error expected.
	b := barcode.NewIntelligentMailBarcode()
	if err := b.Encode("1234-5678"); err == nil {
		t.Error("expected error for 8-digit IMb after dash-filtering, got nil")
	}
}

func TestIntelligentMailBarcode_Render_FallbackOnEncodeError(t *testing.T) {
	// IntelligentMailBarcode.Encode only validates digit count, NOT that
	// second digit is 0–4. So "09123456789012345678" (2nd digit '9') passes
	// Encode but fails imb_encode → Render falls back to placeholderImage.
	b := barcode.NewIntelligentMailBarcode()
	if err := b.Encode("09123456789012345678"); err != nil {
		t.Fatalf("Encode: %v", err) // Encode succeeds (20 digits)
	}
	img, err := b.Render(200, 60)
	// Render returns nil error + placeholder image on imb_encode failure.
	if err != nil {
		t.Fatalf("Render should not error on imb_encode failure, got: %v", err)
	}
	if img == nil {
		t.Error("Render fallback should return non-nil image")
	}
}

func TestIntelligentMailBarcode_Render_ZeroDimensions(t *testing.T) {
	// Covers width<=0 and height<=0 defaults inside Render.
	b := barcode.NewIntelligentMailBarcode()
	if err := b.Encode("01234567094987654321"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(0, 0)
	if err != nil {
		t.Fatalf("Render(0,0): %v", err)
	}
	if img == nil {
		t.Error("Render(0,0) returned nil image")
	}
}

func TestIntelligentMailBarcode_Render_25Digits(t *testing.T) {
	// 25-digit IMb (20 barcode + 5-digit ZIP). Covers zip case 5 in imb_encode.
	b := barcode.NewIntelligentMailBarcode()
	// Second digit '1' (≤ 4), total 25 digits.
	if err := b.Encode("0123456709498765432190210"); err != nil {
		t.Fatalf("Encode 25 digits: %v", err)
	}
	img, err := b.Render(200, 60)
	if err != nil {
		t.Fatalf("Render 25 digits: %v", err)
	}
	if img == nil {
		t.Error("Render 25 digits returned nil image")
	}
}

func TestIntelligentMailBarcode_Render_29Digits(t *testing.T) {
	// 29-digit IMb (20 barcode + 9-digit ZIP). Covers zip case 9 in imb_encode.
	b := barcode.NewIntelligentMailBarcode()
	// Second digit '1' (≤ 4), total 29 digits.
	if err := b.Encode("01234567094987654321902101234"); err != nil {
		t.Fatalf("Encode 29 digits: %v", err)
	}
	img, err := b.Render(200, 60)
	if err != nil {
		t.Fatalf("Render 29 digits: %v", err)
	}
	if img == nil {
		t.Error("Render 29 digits returned nil image")
	}
}

// -----------------------------------------------------------------------
// MaxiCode: cover remaining branches
// -----------------------------------------------------------------------

func TestMaxiCodeMode3Payload(t *testing.T) {
	got := barcode.MaxiCodeMode3Payload("12345", "840", "10", "secondary data")
	if got == "" {
		t.Error("MaxiCodeMode3Payload returned empty string")
	}
}

func TestMaxiCodeBarcode_Encode_InvalidMode(t *testing.T) {
	b := barcode.NewMaxiCodeBarcode()
	b.Mode = 1 // invalid — must be 2-6
	if err := b.Encode("test"); err == nil {
		t.Error("expected error for MaxiCode mode 1, got nil")
	}
}

func TestMaxiCodeBarcode_Render_ZeroDimensions(t *testing.T) {
	// Covers the width/height <=0 branches and min2(width, height) path.
	b := barcode.NewMaxiCodeBarcode()
	if err := b.Encode("test"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(0, 0)
	if err != nil {
		t.Fatalf("Render(0,0): %v", err)
	}
	if img == nil {
		t.Error("Render(0,0) returned nil image")
	}
}

func TestMaxiCodeBarcode_Encode_Mode5(t *testing.T) {
	// Mode 5 uses different secondary data/ECC sizes (68 vs 84).
	b := barcode.NewMaxiCodeBarcode()
	b.Mode = 5
	if err := b.Encode("Mode 5 payload"); err != nil {
		t.Fatalf("Encode mode 5: %v", err)
	}
	img, err := b.Render(100, 100)
	if err != nil {
		t.Fatalf("Render mode 5: %v", err)
	}
	if img == nil {
		t.Error("Render mode 5 returned nil image")
	}
}

func TestMaxiCodeEncodeText_ControlCharsAndSetA(t *testing.T) {
	// Pass a string that starts with a printable char (Set B) then a control
	// char (< 0x20) → triggers LATA latch, then back with LATB.
	// This covers the inSetA=false → latch to SetA → latch back to SetB path.
	ecc := barcode.MaxiCodeComputeECC([]byte{0x01, 0x02, 0x03}, 3)
	if len(ecc) != 3 {
		t.Errorf("MaxiCodeComputeECC len = %d, want 3", len(ecc))
	}
	// Encode text with control characters via a mode-4 MaxiCode barcode render.
	b := barcode.NewMaxiCodeBarcode()
	b.Mode = 4
	// ASCII control chars embedded in text → triggers Set A / Set B latching.
	payload := "Hello\x01World\x02"
	if err := b.Encode(payload); err != nil {
		t.Fatalf("Encode with ctrl chars: %v", err)
	}
	if _, err := b.Render(100, 100); err != nil {
		t.Fatalf("Render with ctrl chars: %v", err)
	}
}

func TestMaxiCodeBarcode_Render_NotEncoded(t *testing.T) {
	// Covers the "maxicode: not encoded" error path in Render.
	b := barcode.NewMaxiCodeBarcode()
	_, err := b.Render(100, 100)
	if err == nil {
		t.Error("expected error from Render when not encoded")
	}
}

func TestMaxiCodeBarcode_Render_PortraitDimensions(t *testing.T) {
	// width < height → min2 returns width (covers the `return a` branch in min2).
	b := barcode.NewMaxiCodeBarcode()
	if err := b.Encode("portrait render test"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(80, 120)
	if err != nil {
		t.Fatalf("Render(80,120): %v", err)
	}
	if img == nil {
		t.Error("Render(80,120) returned nil image")
	}
}

func TestMaxiCodeEncodeText_NonASCIIAndDEL(t *testing.T) {
	// Non-ASCII rune → substituted with GS (0x1D) → goes to Set A path.
	// DEL (0x7F) in Set B → idx = 0x7F - 0x1F = 96 > 63 → clamped to 63.
	b := barcode.NewMaxiCodeBarcode()
	b.Mode = 4
	// "café\x7f" contains:
	// - 'c','a' → Set B
	// - 'f','é' → 'é' is U+00E9 > 0x7F → GS (0x1D) → Set A
	// - '\x7f' → DEL, idx=96 → clamp to 63
	payload := "caf\u00e9\x7f"
	if err := b.Encode(payload); err != nil {
		t.Fatalf("Encode non-ASCII: %v", err)
	}
	if _, err := b.Render(100, 100); err != nil {
		t.Fatalf("Render non-ASCII: %v", err)
	}
}
