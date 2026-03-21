package object_test

// barcode_obj_test.go — additional coverage tests for barcode.go
// Targets: BarcodeObject.Serialize/Deserialize and ZipCodeObject.Serialize/Deserialize
// to cover the remaining uncovered branches (default field values, non-default paths).

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── BarcodeObject: all default values (no attributes written) ────────────────

func TestBarcodeObject_Serialize_Defaults(t *testing.T) {
	orig := object.NewBarcodeObject()
	// defaults: showText=true, autoSize=true, allowExpressions=false, text="", barcodeType=""

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("BarcodeObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	// None of these should be written since they match defaults.
	if strings.Contains(xml, `ShowText=`) {
		t.Errorf("unexpected ShowText in default XML:\n%s", xml)
	}
	if strings.Contains(xml, `AutoSize=`) {
		t.Errorf("unexpected AutoSize in default XML:\n%s", xml)
	}
	if strings.Contains(xml, `AllowExpressions=`) {
		t.Errorf("unexpected AllowExpressions in default XML:\n%s", xml)
	}
}

// ── BarcodeObject: round-trip with all non-default values ────────────────────

func TestBarcodeObject_RoundTrip_AllFields(t *testing.T) {
	orig := object.NewBarcodeObject()
	orig.SetText("HelloBarcode")
	orig.SetBarcodeType("QR Code")
	orig.SetShowText(false)
	orig.SetAutoSize(false)
	orig.SetAllowExpressions(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("BarcodeObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewBarcodeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if got.Text() != "HelloBarcode" {
		t.Errorf("Text: got %q, want HelloBarcode", got.Text())
	}
	if got.BarcodeType() != "QR Code" {
		t.Errorf("BarcodeType: got %q, want QR Code", got.BarcodeType())
	}
	if got.ShowText() {
		t.Error("ShowText: want false")
	}
	if got.AutoSize() {
		t.Error("AutoSize: want false")
	}
	if !got.AllowExpressions() {
		t.Error("AllowExpressions: want true")
	}
}

// ── BarcodeObject: Deserialize with explicit default values written ───────────

func TestBarcodeObject_Deserialize_ExplicitDefaults(t *testing.T) {
	// Write an XML with explicit ShowText=true, AutoSize=true to hit those read paths.
	xmlStr := `<BarcodeObject ShowText="true" AutoSize="true" AllowExpressions="false" Text="abc" Barcode="Code128"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewBarcodeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Text() != "abc" {
		t.Errorf("Text: got %q, want abc", got.Text())
	}
	if got.BarcodeType() != "Code128" {
		t.Errorf("BarcodeType: got %q, want Code128", got.BarcodeType())
	}
	if !got.ShowText() {
		t.Error("ShowText: want true")
	}
	if !got.AutoSize() {
		t.Error("AutoSize: want true")
	}
	if got.AllowExpressions() {
		t.Error("AllowExpressions: want false")
	}
}

// ── ZipCodeObject: all fields set (non-default) ──────────────────────────────

func TestZipCodeObject_RoundTrip_AllFields(t *testing.T) {
	orig := object.NewZipCodeObject()
	orig.SetText("654321")  // non-default (default = "123456")
	orig.SetDataColumn("ZipCol")
	orig.SetExpression("[ZipCode]")
	orig.SetSegmentWidth(5.5)
	orig.SetSegmentHeight(12.0)
	orig.SetSpacing(2.0)
	orig.SetSegmentCount(5) // non-default (default = 6)
	orig.SetShowMarkers(false)
	orig.SetShowGrid(false)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ZipCodeObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	for _, attr := range []string{"Text=", "DataColumn=", "Expression=", "SegmentWidth=",
		"SegmentHeight=", "Spacing=", "SegmentCount=", "ShowMarkers=", "ShowGrid="} {
		if !strings.Contains(xml, attr) {
			t.Errorf("expected %q in XML:\n%s", attr, xml)
		}
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewZipCodeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if got.Text() != "654321" {
		t.Errorf("Text: got %q, want 654321", got.Text())
	}
	if got.DataColumn() != "ZipCol" {
		t.Errorf("DataColumn: got %q, want ZipCol", got.DataColumn())
	}
	if got.Expression() != "[ZipCode]" {
		t.Errorf("Expression: got %q, want [ZipCode]", got.Expression())
	}
	if got.SegmentWidth() != 5.5 {
		t.Errorf("SegmentWidth: got %v, want 5.5", got.SegmentWidth())
	}
	if got.SegmentHeight() != 12.0 {
		t.Errorf("SegmentHeight: got %v, want 12.0", got.SegmentHeight())
	}
	if got.Spacing() != 2.0 {
		t.Errorf("Spacing: got %v, want 2.0", got.Spacing())
	}
	if got.SegmentCount() != 5 {
		t.Errorf("SegmentCount: got %d, want 5", got.SegmentCount())
	}
	if got.ShowMarkers() {
		t.Error("ShowMarkers: want false")
	}
	if got.ShowGrid() {
		t.Error("ShowGrid: want false")
	}
}

// ── ZipCodeObject: Serialize with default values (none written) ───────────────

func TestZipCodeObject_Serialize_Defaults(t *testing.T) {
	orig := object.NewZipCodeObject()
	// defaults: text="123456", segmentCount=6, showMarkers=true, showGrid=true,
	// segmentWidth=18.9, segmentHeight=37.8, spacing=34.02 (C# Units.Centimeters)
	// — none of these should be written to XML (diff-based serialization)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ZipCodeObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if strings.Contains(xml, `SegmentCount=`) {
		t.Errorf("unexpected SegmentCount in default XML:\n%s", xml)
	}
	if strings.Contains(xml, `ShowMarkers=`) {
		t.Errorf("unexpected ShowMarkers in default XML:\n%s", xml)
	}
	if strings.Contains(xml, `ShowGrid=`) {
		t.Errorf("unexpected ShowGrid in default XML:\n%s", xml)
	}
	// text="123456" is the C# default — must not appear in XML
	if strings.Contains(xml, `Text=`) {
		t.Errorf("unexpected Text= in default XML:\n%s", xml)
	}
	// segmentWidth/Height/spacing at C# defaults — must not appear in XML
	if strings.Contains(xml, `SegmentWidth=`) {
		t.Errorf("unexpected SegmentWidth= in default XML:\n%s", xml)
	}
	if strings.Contains(xml, `SegmentHeight=`) {
		t.Errorf("unexpected SegmentHeight= in default XML:\n%s", xml)
	}
	if strings.Contains(xml, `Spacing=`) {
		t.Errorf("unexpected Spacing= in default XML:\n%s", xml)
	}
}

// ── ZipCodeObject: Deserialize with explicit default values ──────────────────

func TestZipCodeObject_Deserialize_ExplicitDefaults(t *testing.T) {
	xmlStr := `<ZipCodeObject ShowMarkers="true" ShowGrid="true" SegmentCount="6"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewZipCodeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !got.ShowMarkers() {
		t.Error("ShowMarkers: want true")
	}
	if !got.ShowGrid() {
		t.Error("ShowGrid: want true")
	}
	if got.SegmentCount() != 6 {
		t.Errorf("SegmentCount: got %d, want 6", got.SegmentCount())
	}
}
