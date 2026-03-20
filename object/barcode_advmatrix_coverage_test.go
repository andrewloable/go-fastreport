package object

// barcode_advmatrix_coverage_test.go — internal tests to cover the remaining
// unreachable `return err` branches in BarcodeObject.Serialize/Deserialize,
// ZipCodeObject.Serialize/Deserialize, and AdvMatrixObject.Serialize/Deserialize.
//
// All targeted branches guard calls to ReportComponentBase.Serialize/Deserialize,
// which always returns nil — making these branches dead code through the
// normal writer/reader interface. The tests below exercise the full reachable
// function bodies with no-error mocks, and document the field-reading paths
// using custom readers that return specific values.

import (
	"testing"
)

// ── BarcodeObject: Serialize with no-error writer ────────────────────────────

// TestBarcodeObject_Serialize_NegativeBranch exercises BarcodeObject.Serialize
// via an errBaseWriter{err: nil} (no-op, never errors). This registers the
// function body as covered, leaving only the dead-code `return err` line at
// line 66 as the sole uncovered statement.
func TestBarcodeObject_Serialize_NoopWriter(t *testing.T) {
	b := NewBarcodeObject()
	b.SetText("12345678")
	b.SetBarcodeType("Code128")
	b.SetShowText(false)
	b.SetAutoSize(false)
	b.SetAllowExpressions(true)

	w := &errBaseWriter{err: nil}
	if err := b.Serialize(w); err != nil {
		t.Fatalf("BarcodeObject.Serialize unexpected error: %v", err)
	}
}

// TestBarcodeObject_Serialize_Defaults_NoopWriter exercises the branch where
// text and barcodeType are empty (skipped) and showText/autoSize are defaults.
func TestBarcodeObject_Serialize_Defaults_NoopWriter(t *testing.T) {
	b := NewBarcodeObject()
	// defaults: showText=true, autoSize=true, allowExpressions=false, text="", barcodeType=""

	w := &errBaseWriter{err: nil}
	if err := b.Serialize(w); err != nil {
		t.Fatalf("BarcodeObject.Serialize unexpected error: %v", err)
	}
}

// ── BarcodeObject: Deserialize with default and custom readers ────────────────

// TestBarcodeObject_Deserialize_DefaultReader verifies Deserialize returns nil
// with a default reader (all fields at defaults).
func TestBarcodeObject_Deserialize_DefaultReader(t *testing.T) {
	b := NewBarcodeObject()
	r := &defaultReader{}
	if err := b.Deserialize(r); err != nil {
		t.Fatalf("BarcodeObject.Deserialize unexpected error: %v", err)
	}
	if b.text != "" {
		t.Errorf("text: got %q, want empty", b.text)
	}
	if b.barcodeType != "" {
		t.Errorf("barcodeType: got %q, want empty", b.barcodeType)
	}
	// showText default = true, autoSize default = true (passed as def to ReadBool)
	if !b.showText {
		t.Error("showText should be true (default)")
	}
	if !b.autoSize {
		t.Error("autoSize should be true (default)")
	}
	if b.allowExpressions {
		t.Error("allowExpressions should be false (default)")
	}
}

// TestBarcodeObject_Deserialize_AllFields verifies Deserialize reads all fields.
func TestBarcodeObject_Deserialize_AllFields(t *testing.T) {
	b := NewBarcodeObject()
	r := &barcodeAllFieldsReader{
		text:             "HelloWorld",
		barcodeType:      "QR Code",
		showText:         false,
		autoSize:         false,
		allowExpressions: true,
	}
	if err := b.Deserialize(r); err != nil {
		t.Fatalf("BarcodeObject.Deserialize unexpected error: %v", err)
	}
	if b.text != "HelloWorld" {
		t.Errorf("text: got %q, want HelloWorld", b.text)
	}
	if b.barcodeType != "QR Code" {
		t.Errorf("barcodeType: got %q, want QR Code", b.barcodeType)
	}
	if b.showText {
		t.Error("showText should be false")
	}
	if b.autoSize {
		t.Error("autoSize should be false")
	}
	if !b.allowExpressions {
		t.Error("allowExpressions should be true")
	}
}

type barcodeAllFieldsReader struct {
	defaultReader
	text, barcodeType    string
	showText, autoSize   bool
	allowExpressions     bool
}

func (r *barcodeAllFieldsReader) ReadStr(name, def string) string {
	switch name {
	case "Text":
		return r.text
	case "Barcode":
		return r.barcodeType
	}
	return def
}

func (r *barcodeAllFieldsReader) ReadBool(name string, def bool) bool {
	switch name {
	case "ShowText":
		return r.showText
	case "AutoSize":
		return r.autoSize
	case "AllowExpressions":
		return r.allowExpressions
	}
	return def
}

// ── ZipCodeObject: Serialize with no-error writer ────────────────────────────

// TestZipCodeObject_Serialize_NoopWriter exercises ZipCodeObject.Serialize
// via a no-op writer with all non-default field values set.
func TestZipCodeObject_Serialize_NoopWriter(t *testing.T) {
	z := NewZipCodeObject()
	z.text = "123456"
	z.dataColumn = "ZipCol"
	z.expression = "[Zip]"
	z.segmentWidth = 4.5
	z.segmentHeight = 10.0
	z.spacing = 1.5
	z.segmentCount = 5 // non-default
	z.showMarkers = false
	z.showGrid = false

	w := &errBaseWriter{err: nil}
	if err := z.Serialize(w); err != nil {
		t.Fatalf("ZipCodeObject.Serialize unexpected error: %v", err)
	}
}

// TestZipCodeObject_Serialize_Defaults_NoopWriter exercises the default-value
// branches (none of the optional attributes are written).
func TestZipCodeObject_Serialize_Defaults_NoopWriter(t *testing.T) {
	z := NewZipCodeObject()
	// defaults: segmentCount=6, showMarkers=true, showGrid=true, all others zero

	w := &errBaseWriter{err: nil}
	if err := z.Serialize(w); err != nil {
		t.Fatalf("ZipCodeObject.Serialize unexpected error: %v", err)
	}
}

// ── ZipCodeObject: Deserialize with default and custom readers ────────────────

// TestZipCodeObject_Deserialize_DefaultReader verifies Deserialize returns nil
// with a default reader (all fields at defaults).
func TestZipCodeObject_Deserialize_DefaultReader(t *testing.T) {
	z := NewZipCodeObject()
	r := &defaultReader{}
	if err := z.Deserialize(r); err != nil {
		t.Fatalf("ZipCodeObject.Deserialize unexpected error: %v", err)
	}
	if z.text != "" {
		t.Errorf("text: got %q, want empty", z.text)
	}
	if z.segmentCount != 6 {
		t.Errorf("segmentCount: got %d, want 6 (default)", z.segmentCount)
	}
	if !z.showMarkers {
		t.Error("showMarkers should be true (default)")
	}
	if !z.showGrid {
		t.Error("showGrid should be true (default)")
	}
}

// TestZipCodeObject_Deserialize_AllFields verifies Deserialize reads all fields.
func TestZipCodeObject_Deserialize_AllFields(t *testing.T) {
	z := NewZipCodeObject()
	r := &zipCodeAllFieldsReader{
		text:          "654321",
		dataColumn:    "Col1",
		expression:    "[ZipExpr]",
		segmentWidth:  3.0,
		segmentHeight: 8.0,
		spacing:       2.0,
		segmentCount:  4,
		showMarkers:   false,
		showGrid:      false,
	}
	if err := z.Deserialize(r); err != nil {
		t.Fatalf("ZipCodeObject.Deserialize unexpected error: %v", err)
	}
	if z.text != "654321" {
		t.Errorf("text: got %q, want 654321", z.text)
	}
	if z.dataColumn != "Col1" {
		t.Errorf("dataColumn: got %q, want Col1", z.dataColumn)
	}
	if z.expression != "[ZipExpr]" {
		t.Errorf("expression: got %q, want [ZipExpr]", z.expression)
	}
	if z.segmentWidth != 3.0 {
		t.Errorf("segmentWidth: got %v, want 3.0", z.segmentWidth)
	}
	if z.segmentHeight != 8.0 {
		t.Errorf("segmentHeight: got %v, want 8.0", z.segmentHeight)
	}
	if z.spacing != 2.0 {
		t.Errorf("spacing: got %v, want 2.0", z.spacing)
	}
	if z.segmentCount != 4 {
		t.Errorf("segmentCount: got %d, want 4", z.segmentCount)
	}
	if z.showMarkers {
		t.Error("showMarkers should be false")
	}
	if z.showGrid {
		t.Error("showGrid should be false")
	}
}

type zipCodeAllFieldsReader struct {
	defaultReader
	text, dataColumn, expression string
	segmentWidth, segmentHeight  float32
	spacing                      float32
	segmentCount                 int
	showMarkers, showGrid        bool
}

func (r *zipCodeAllFieldsReader) ReadStr(name, def string) string {
	switch name {
	case "Text":
		return r.text
	case "DataColumn":
		return r.dataColumn
	case "Expression":
		return r.expression
	}
	return def
}

func (r *zipCodeAllFieldsReader) ReadFloat(name string, def float32) float32 {
	switch name {
	case "SegmentWidth":
		return r.segmentWidth
	case "SegmentHeight":
		return r.segmentHeight
	case "Spacing":
		return r.spacing
	}
	return def
}

func (r *zipCodeAllFieldsReader) ReadInt(name string, def int) int {
	if name == "SegmentCount" {
		return r.segmentCount
	}
	return def
}

func (r *zipCodeAllFieldsReader) ReadBool(name string, def bool) bool {
	switch name {
	case "ShowMarkers":
		return r.showMarkers
	case "ShowGrid":
		return r.showGrid
	}
	return def
}

// ── AdvMatrixObject: Serialize and Deserialize with no-error mock ─────────────

// TestAdvMatrixObject_Serialize_NoopWriter_AllFields exercises all branches in
// AdvMatrixObject.Serialize via a no-op writer with DataSource set.
func TestAdvMatrixObject_Serialize_NoopWriter_AllFields(t *testing.T) {
	a := NewAdvMatrixObject()
	a.DataSource = "TestDS"
	a.TableColumns = append(a.TableColumns, &AdvMatrixColumn{Name: "C1", Width: 80})
	a.TableRows = append(a.TableRows, &AdvMatrixRow{
		Name:   "R1",
		Height: 25,
		Cells:  []*AdvMatrixCell{{Name: "cell1", Text: "data"}},
	})

	w := &errBaseWriter{err: nil}
	if err := a.Serialize(w); err != nil {
		t.Fatalf("AdvMatrixObject.Serialize unexpected error: %v", err)
	}
}

// TestAdvMatrixObject_Serialize_NoopWriter_NoDataSource exercises the branch
// where DataSource is empty (the if-block is skipped).
func TestAdvMatrixObject_Serialize_NoopWriter_NoDataSource(t *testing.T) {
	a := NewAdvMatrixObject()
	// DataSource="" — the if-block at line 115 is skipped.

	w := &errBaseWriter{err: nil}
	if err := a.Serialize(w); err != nil {
		t.Fatalf("AdvMatrixObject.Serialize unexpected error: %v", err)
	}
}

// TestAdvMatrixObject_Deserialize_AllFields_NoopReader verifies Deserialize
// reads DataSource via a custom reader that supplies it.
func TestAdvMatrixObject_Deserialize_AllFields_NoopReader(t *testing.T) {
	a := NewAdvMatrixObject()
	r := &advMatrixAllFieldsReader{dataSource: "ProdDB"}
	if err := a.Deserialize(r); err != nil {
		t.Fatalf("AdvMatrixObject.Deserialize unexpected error: %v", err)
	}
	if a.DataSource != "ProdDB" {
		t.Errorf("DataSource: got %q, want ProdDB", a.DataSource)
	}
}

type advMatrixAllFieldsReader struct {
	defaultReader
	dataSource string
}

func (r *advMatrixAllFieldsReader) ReadStr(name, def string) string {
	if name == "DataSource" {
		return r.dataSource
	}
	return def
}
