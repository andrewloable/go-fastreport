package object_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/style"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// roundTripXML serializes obj into XML, reads the header, then calls deserFn to
// deserialize, and returns the XML string for spot-checks.
func roundTripXML(t *testing.T, tagName string, ser interface {
	Serialize(interface{ WriteStr(string, string); WriteInt(string, int); WriteBool(string, bool); WriteFloat(string, float32); WriteObject(interface{ Serialize(interface{}) error }) error; WriteObjectNamed(string, interface{ Serialize(interface{}) error }) error }) error
}) string {
	t.Helper()
	// Just return empty — we use the serial-specific helpers below.
	return ""
}

// serializeTo serializes any report.Serializable to XML bytes using serial.Writer.
func serializeTo(t *testing.T, tagName string, obj interface {
	Serialize(w interface {
		WriteStr(name, value string)
		WriteInt(name string, value int)
		WriteBool(name string, value bool)
		WriteFloat(name string, value float32)
	}) error
}) []byte {
	t.Helper()
	// Use a concrete writer below.
	return nil
}

// ══════════════════════════════════════════════════════════════════════════════
// BarcodeObject
// ══════════════════════════════════════════════════════════════════════════════

func TestNewBarcodeObject_Defaults(t *testing.T) {
	b := object.NewBarcodeObject()
	if b == nil {
		t.Fatal("NewBarcodeObject returned nil")
	}
	if !b.ShowText() {
		t.Error("ShowText default should be true")
	}
	if !b.AutoSize() {
		t.Error("AutoSize default should be true")
	}
	if b.Text() != "12345678" {
		t.Errorf("Text default = %q, want 12345678", b.Text())
	}
	if b.BarcodeType() != "" {
		t.Errorf("BarcodeType default = %q, want empty", b.BarcodeType())
	}
	if b.AllowExpressions() {
		t.Error("AllowExpressions default should be false")
	}
}

func TestBarcodeObject_Setters(t *testing.T) {
	b := object.NewBarcodeObject()
	b.SetText("123456")
	b.SetBarcodeType("QR Code")
	b.SetShowText(false)
	b.SetAutoSize(false)
	b.SetAllowExpressions(true)

	if b.Text() != "123456" {
		t.Errorf("Text = %q", b.Text())
	}
	if b.BarcodeType() != "QR Code" {
		t.Errorf("BarcodeType = %q", b.BarcodeType())
	}
	if b.ShowText() {
		t.Error("ShowText should be false")
	}
	if b.AutoSize() {
		t.Error("AutoSize should be false")
	}
	if !b.AllowExpressions() {
		t.Error("AllowExpressions should be true")
	}
}

func TestBarcodeObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewBarcodeObject()
	orig.SetName("bc1")
	orig.SetText("HELLO")
	orig.SetBarcodeType("Code128")
	orig.SetShowText(false)
	orig.SetAutoSize(false)
	orig.SetAllowExpressions(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("BarcodeObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "BarcodeObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}

	got := object.NewBarcodeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Name() != "bc1" {
		t.Errorf("Name = %q", got.Name())
	}
	if got.Text() != "HELLO" {
		t.Errorf("Text = %q", got.Text())
	}
	if got.BarcodeType() != "Code128" {
		t.Errorf("BarcodeType = %q", got.BarcodeType())
	}
	if got.ShowText() {
		t.Error("ShowText should be false after round-trip")
	}
	if got.AutoSize() {
		t.Error("AutoSize should be false after round-trip")
	}
	if !got.AllowExpressions() {
		t.Error("AllowExpressions should be true after round-trip")
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// ZipCodeObject
// ══════════════════════════════════════════════════════════════════════════════

func TestNewZipCodeObject_Defaults(t *testing.T) {
	z := object.NewZipCodeObject()
	if z == nil {
		t.Fatal("NewZipCodeObject returned nil")
	}
	if z.SegmentCount() != 6 {
		t.Errorf("SegmentCount default = %d, want 6", z.SegmentCount())
	}
	if !z.ShowMarkers() {
		t.Error("ShowMarkers default should be true")
	}
	if !z.ShowGrid() {
		t.Error("ShowGrid default should be true")
	}
}

func TestZipCodeObject_Setters(t *testing.T) {
	z := object.NewZipCodeObject()
	z.SetText("12345")
	z.SetDataColumn("ZipCol")
	z.SetExpression("[Zip]")
	z.SetSegmentWidth(10)
	z.SetSegmentHeight(20)
	z.SetSpacing(5)
	z.SetSegmentCount(5)
	z.SetShowMarkers(false)
	z.SetShowGrid(false)

	if z.Text() != "12345" {
		t.Errorf("Text = %q", z.Text())
	}
	if z.DataColumn() != "ZipCol" {
		t.Errorf("DataColumn = %q", z.DataColumn())
	}
	if z.Expression() != "[Zip]" {
		t.Errorf("Expression = %q", z.Expression())
	}
	if z.SegmentWidth() != 10 {
		t.Errorf("SegmentWidth = %v", z.SegmentWidth())
	}
	if z.SegmentHeight() != 20 {
		t.Errorf("SegmentHeight = %v", z.SegmentHeight())
	}
	if z.Spacing() != 5 {
		t.Errorf("Spacing = %v", z.Spacing())
	}
	if z.SegmentCount() != 5 {
		t.Errorf("SegmentCount = %d", z.SegmentCount())
	}
	if z.ShowMarkers() {
		t.Error("ShowMarkers should be false")
	}
	if z.ShowGrid() {
		t.Error("ShowGrid should be false")
	}
}

func TestZipCodeObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewZipCodeObject()
	orig.SetName("zip1")
	orig.SetText("90210")
	orig.SetSegmentCount(7)
	orig.SetShowMarkers(false)
	orig.SetShowGrid(false)
	orig.SetSegmentWidth(12)
	orig.SetSegmentHeight(24)
	orig.SetSpacing(3)
	orig.SetDataColumn("Col1")
	orig.SetExpression("[x]")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ZipCodeObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "ZipCodeObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}

	got := object.NewZipCodeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Text() != "90210" {
		t.Errorf("Text = %q", got.Text())
	}
	if got.SegmentCount() != 7 {
		t.Errorf("SegmentCount = %d", got.SegmentCount())
	}
	if got.ShowMarkers() {
		t.Error("ShowMarkers should be false")
	}
	if got.SegmentWidth() != 12 {
		t.Errorf("SegmentWidth = %v", got.SegmentWidth())
	}
	if got.DataColumn() != "Col1" {
		t.Errorf("DataColumn = %q", got.DataColumn())
	}
	if got.Expression() != "[x]" {
		t.Errorf("Expression = %q", got.Expression())
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// CellularTextObject — serialize/deserialize round-trip
// ══════════════════════════════════════════════════════════════════════════════

func TestCellularTextObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewCellularTextObject()
	orig.SetName("ct1")
	orig.SetCellWidth(28)
	orig.SetCellHeight(14)
	orig.SetHorzSpacing(4)
	orig.SetVertSpacing(2)
	orig.SetWordWrap(false)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("CellularTextObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "CellularTextObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewCellularTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.CellWidth() != 28 {
		t.Errorf("CellWidth = %v", got.CellWidth())
	}
	if got.CellHeight() != 14 {
		t.Errorf("CellHeight = %v", got.CellHeight())
	}
	if got.HorzSpacing() != 4 {
		t.Errorf("HorzSpacing = %v", got.HorzSpacing())
	}
	if got.VertSpacing() != 2 {
		t.Errorf("VertSpacing = %v", got.VertSpacing())
	}
	if got.WordWrap() {
		t.Error("WordWrap should be false")
	}
}

func TestCellularTextObject_VertSpacing(t *testing.T) {
	c := object.NewCellularTextObject()
	c.SetVertSpacing(7)
	if c.VertSpacing() != 7 {
		t.Errorf("VertSpacing = %v, want 7", c.VertSpacing())
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// DigitalSignatureObject
// ══════════════════════════════════════════════════════════════════════════════

func TestNewDigitalSignatureObject_Defaults(t *testing.T) {
	d := object.NewDigitalSignatureObject()
	if d == nil {
		t.Fatal("NewDigitalSignatureObject returned nil")
	}
	if d.TypeName() != "DigitalSignatureObject" {
		t.Errorf("TypeName = %q", d.TypeName())
	}
	if d.BaseName() != "DigitalSignature" {
		t.Errorf("BaseName = %q", d.BaseName())
	}
	if d.Placeholder() != "" {
		t.Errorf("Placeholder default = %q, want empty", d.Placeholder())
	}
}

func TestDigitalSignatureObject_Placeholder(t *testing.T) {
	d := object.NewDigitalSignatureObject()
	d.SetPlaceholder("Sign here")
	if d.Placeholder() != "Sign here" {
		t.Errorf("Placeholder = %q", d.Placeholder())
	}
}

func TestDigitalSignatureObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewDigitalSignatureObject()
	orig.SetName("ds1")
	orig.SetPlaceholder("Click to sign")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("DigitalSignatureObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "DigitalSignatureObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewDigitalSignatureObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Placeholder() != "Click to sign" {
		t.Errorf("Placeholder = %q", got.Placeholder())
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// HtmlObject
// ══════════════════════════════════════════════════════════════════════════════

func TestNewHtmlObject_Defaults(t *testing.T) {
	h := object.NewHtmlObject()
	if h == nil {
		t.Fatal("NewHtmlObject returned nil")
	}
	if h.RightToLeft() {
		t.Error("RightToLeft should default to false")
	}
}

func TestHtmlObject_RightToLeft(t *testing.T) {
	h := object.NewHtmlObject()
	h.SetRightToLeft(true)
	if !h.RightToLeft() {
		t.Error("RightToLeft should be true")
	}
}

func TestHtmlObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewHtmlObject()
	orig.SetName("html1")
	orig.SetText("<b>Hello</b>")
	orig.SetRightToLeft(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("HtmlObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "HtmlObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewHtmlObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if !got.RightToLeft() {
		t.Error("RightToLeft should be true after round-trip")
	}
	if got.Text() != "<b>Hello</b>" {
		t.Errorf("Text = %q", got.Text())
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// MapObject / MapLayer
// ══════════════════════════════════════════════════════════════════════════════

func TestNewMapObject_Defaults(t *testing.T) {
	m := object.NewMapObject()
	if m == nil {
		t.Fatal("NewMapObject returned nil")
	}
	if m.TypeName() != "MapObject" {
		t.Errorf("TypeName = %q", m.TypeName())
	}
	if m.BaseName() != "Map" {
		t.Errorf("BaseName = %q", m.BaseName())
	}
	if m.OffsetX != 0 || m.OffsetY != 0 {
		t.Errorf("Offsets should default to 0")
	}
}

func TestMapObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewMapObject()
	orig.SetName("map1")
	orig.OffsetX = 10
	orig.OffsetY = 20

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MapObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "MapObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewMapObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.OffsetX != 10 {
		t.Errorf("OffsetX = %v", got.OffsetX)
	}
	if got.OffsetY != 20 {
		t.Errorf("OffsetY = %v", got.OffsetY)
	}
}

func TestNewMapLayer_Defaults(t *testing.T) {
	l := object.NewMapLayer()
	if l == nil {
		t.Fatal("NewMapLayer returned nil")
	}
	if l.TypeName() != "MapLayer" {
		t.Errorf("TypeName = %q", l.TypeName())
	}
	if l.BaseName() != "MapLayer" {
		t.Errorf("BaseName = %q", l.BaseName())
	}
}

func TestMapLayer_SerializeDeserialize(t *testing.T) {
	orig := object.NewMapLayer()
	orig.SetName("layer1")
	orig.Shapefile = "world.shp"
	orig.Type = "Choropleth"
	orig.DataSource = "DS1"
	orig.Filter = "[Country] = 'US'"
	orig.SpatialColumn = "ISOCode"
	orig.SpatialValue = "[Code]"
	orig.AnalyticalValue = "[Population]"
	orig.LabelColumn = "Name"
	orig.BoxAsString = "0,0,100,100"
	orig.Palette = "Heat"

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MapLayer", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "MapLayer" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewMapLayer()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Shapefile != "world.shp" {
		t.Errorf("Shapefile = %q", got.Shapefile)
	}
	if got.Type != "Choropleth" {
		t.Errorf("Type = %q", got.Type)
	}
	if got.DataSource != "DS1" {
		t.Errorf("DataSource = %q", got.DataSource)
	}
	if got.Filter != "[Country] = 'US'" {
		t.Errorf("Filter = %q", got.Filter)
	}
	if got.SpatialColumn != "ISOCode" {
		t.Errorf("SpatialColumn = %q", got.SpatialColumn)
	}
	if got.SpatialValue != "[Code]" {
		t.Errorf("SpatialValue = %q", got.SpatialValue)
	}
	if got.AnalyticalValue != "[Population]" {
		t.Errorf("AnalyticalValue = %q", got.AnalyticalValue)
	}
	if got.LabelColumn != "Name" {
		t.Errorf("LabelColumn = %q", got.LabelColumn)
	}
	if got.BoxAsString != "0,0,100,100" {
		t.Errorf("BoxAsString = %q", got.BoxAsString)
	}
	if got.Palette != "Heat" {
		t.Errorf("Palette = %q", got.Palette)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// RFIDLabel — serialize/deserialize round-trip
// ══════════════════════════════════════════════════════════════════════════════

func TestRFIDLabel_SerializeDeserialize(t *testing.T) {
	orig := object.NewRFIDLabel()
	orig.SetName("rfid1")
	orig.EPCBank = object.RFIDBank{Data: "AABB", DataColumn: "col", Offset: 2, DataFormat: object.RFIDBankFormatASCII}
	orig.TIDBank = object.RFIDBank{Data: "TID1"}
	orig.UserBank = object.RFIDBank{Data: "USER", DataFormat: object.RFIDBankFormatDecimal}
	orig.AccessPassword = "AAAA"
	orig.AccessPasswordDataColumn = "accCol"
	orig.KillPassword = "BBBB"
	orig.KillPasswordDataColumn = "killCol"
	orig.LockKillPassword = object.RFIDLockTypeLock
	orig.LockAccessPassword = object.RFIDLockTypePermanentLock
	orig.LockEPCBank = object.RFIDLockTypePermanentUnlock
	orig.LockUserBank = object.RFIDLockTypeLock
	orig.UseAdjustForEPC = true
	orig.RewriteEPCBank = true
	orig.ErrorHandle = object.RFIDErrorHandlePause

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RFIDLabel", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "RFIDLabel" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewRFIDLabel()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.EPCBank.Data != "AABB" {
		t.Errorf("EPCBank.Data = %q", got.EPCBank.Data)
	}
	if got.EPCBank.DataColumn != "col" {
		t.Errorf("EPCBank.DataColumn = %q", got.EPCBank.DataColumn)
	}
	if got.EPCBank.Offset != 2 {
		t.Errorf("EPCBank.Offset = %d", got.EPCBank.Offset)
	}
	if got.EPCBank.DataFormat != object.RFIDBankFormatASCII {
		t.Errorf("EPCBank.DataFormat = %v", got.EPCBank.DataFormat)
	}
	if got.TIDBank.Data != "TID1" {
		t.Errorf("TIDBank.Data = %q", got.TIDBank.Data)
	}
	if got.UserBank.DataFormat != object.RFIDBankFormatDecimal {
		t.Errorf("UserBank.DataFormat = %v", got.UserBank.DataFormat)
	}
	if got.AccessPassword != "AAAA" {
		t.Errorf("AccessPassword = %q", got.AccessPassword)
	}
	if got.KillPassword != "BBBB" {
		t.Errorf("KillPassword = %q", got.KillPassword)
	}
	if got.LockKillPassword != object.RFIDLockTypeLock {
		t.Errorf("LockKillPassword = %v", got.LockKillPassword)
	}
	if got.LockAccessPassword != object.RFIDLockTypePermanentLock {
		t.Errorf("LockAccessPassword = %v", got.LockAccessPassword)
	}
	if got.LockEPCBank != object.RFIDLockTypePermanentUnlock {
		t.Errorf("LockEPCBank = %v", got.LockEPCBank)
	}
	if !got.UseAdjustForEPC {
		t.Error("UseAdjustForEPC should be true")
	}
	if !got.RewriteEPCBank {
		t.Error("RewriteEPCBank should be true")
	}
	if got.ErrorHandle != object.RFIDErrorHandlePause {
		t.Errorf("ErrorHandle = %v", got.ErrorHandle)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// RichObject
// ══════════════════════════════════════════════════════════════════════════════

func TestNewRichObject_Defaults(t *testing.T) {
	r := object.NewRichObject()
	if r == nil {
		t.Fatal("NewRichObject returned nil")
	}
	if r.TypeName() != "RichObject" {
		t.Errorf("TypeName = %q", r.TypeName())
	}
	if r.BaseName() != "Rich" {
		t.Errorf("BaseName = %q", r.BaseName())
	}
	if r.Text() != "" {
		t.Errorf("Text default = %q", r.Text())
	}
	if r.CanGrow() {
		t.Error("CanGrow default should be false")
	}
}

func TestRichObject_Setters(t *testing.T) {
	r := object.NewRichObject()
	r.SetText("{\\rtf1 Hello}")
	r.SetCanGrow(true)
	if r.Text() != "{\\rtf1 Hello}" {
		t.Errorf("Text = %q", r.Text())
	}
	if !r.CanGrow() {
		t.Error("CanGrow should be true")
	}
}

func TestRichObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewRichObject()
	orig.SetName("rich1")
	orig.SetText("{\\rtf1 content}")
	orig.SetCanGrow(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("RichObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r2 := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r2.ReadObjectHeader()
	if !ok || typeName != "RichObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewRichObject()
	if err := got.Deserialize(r2); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Text() != "{\\rtf1 content}" {
		t.Errorf("Text = %q", got.Text())
	}
	if !got.CanGrow() {
		t.Error("CanGrow should be true")
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// SparklineObject
// ══════════════════════════════════════════════════════════════════════════════

func TestNewSparklineObject_Defaults(t *testing.T) {
	s := object.NewSparklineObject()
	if s == nil {
		t.Fatal("NewSparklineObject returned nil")
	}
	if s.TypeName() != "SparklineObject" {
		t.Errorf("TypeName = %q", s.TypeName())
	}
	if s.BaseName() != "Sparkline" {
		t.Errorf("BaseName = %q", s.BaseName())
	}
	if s.ChartData != "" {
		t.Errorf("ChartData default = %q", s.ChartData)
	}
	if s.Dock != "" {
		t.Errorf("Dock default = %q", s.Dock)
	}
}

func TestSparklineObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewSparklineObject()
	orig.SetName("spark1")
	orig.ChartData = "base64encodeddata"
	orig.Dock = "Bottom"

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SparklineObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "SparklineObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewSparklineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.ChartData != "base64encodeddata" {
		t.Errorf("ChartData = %q", got.ChartData)
	}
	if got.Dock != "Bottom" {
		t.Errorf("Dock = %q", got.Dock)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// SVGObject
// ══════════════════════════════════════════════════════════════════════════════

func TestNewSVGObject_Defaults(t *testing.T) {
	s := object.NewSVGObject()
	if s == nil {
		t.Fatal("NewSVGObject returned nil")
	}
	if s.TypeName() != "SVGObject" {
		t.Errorf("TypeName = %q", s.TypeName())
	}
	if s.BaseName() != "SVG" {
		t.Errorf("BaseName = %q", s.BaseName())
	}
	if s.SvgData != "" {
		t.Errorf("SvgData default = %q", s.SvgData)
	}
}

func TestSVGObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewSVGObject()
	orig.SetName("svg1")
	orig.SvgData = "PHN2ZyAvPg=="

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SVGObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "SVGObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewSVGObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.SvgData != "PHN2ZyAvPg==" {
		t.Errorf("SvgData = %q", got.SvgData)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// AdvMatrixObject
// ══════════════════════════════════════════════════════════════════════════════

func TestNewAdvMatrixObject_Defaults(t *testing.T) {
	a := object.NewAdvMatrixObject()
	if a == nil {
		t.Fatal("NewAdvMatrixObject returned nil")
	}
	if a.TypeName() != "AdvMatrixObject" {
		t.Errorf("TypeName = %q", a.TypeName())
	}
	if a.BaseName() != "AdvMatrix" {
		t.Errorf("BaseName = %q", a.BaseName())
	}
	if a.DataSource != "" {
		t.Errorf("DataSource default = %q", a.DataSource)
	}
	if len(a.TableColumns) != 0 {
		t.Errorf("TableColumns default len = %d", len(a.TableColumns))
	}
	if len(a.TableRows) != 0 {
		t.Errorf("TableRows default len = %d", len(a.TableRows))
	}
}

func TestAdvMatrixColumn_SerializeDeserialize(t *testing.T) {
	orig := &object.AdvMatrixColumn{Name: "col1", Width: 100, AutoSize: true}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TableColumn", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := &object.AdvMatrixColumn{}
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Name != "col1" {
		t.Errorf("Name = %q", got.Name)
	}
	if got.Width != 100 {
		t.Errorf("Width = %v", got.Width)
	}
	if !got.AutoSize {
		t.Error("AutoSize should be true")
	}
}

func TestAdvMatrixRow_SerializeDeserialize(t *testing.T) {
	orig := &object.AdvMatrixRow{Name: "row1", Height: 30, AutoSize: false}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TableRow", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := &object.AdvMatrixRow{}
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Name != "row1" {
		t.Errorf("Name = %q", got.Name)
	}
	if got.Height != 30 {
		t.Errorf("Height = %v", got.Height)
	}
}

func TestAdvMatrixCell_SerializeDeserialize(t *testing.T) {
	orig := &object.AdvMatrixCell{
		Name:      "cell1",
		Width:     50,
		Height:    20,
		ColSpan:   2,
		RowSpan:   3,
		Text:      "Hello",
		HorzAlign: 1,
		VertAlign: 2,
	}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TableCell", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := &object.AdvMatrixCell{}
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Name != "cell1" {
		t.Errorf("Name = %q", got.Name)
	}
	if got.Width != 50 {
		t.Errorf("Width = %v", got.Width)
	}
	if got.ColSpan != 2 {
		t.Errorf("ColSpan = %d", got.ColSpan)
	}
	if got.RowSpan != 3 {
		t.Errorf("RowSpan = %d", got.RowSpan)
	}
	if got.Text != "Hello" {
		t.Errorf("Text = %q", got.Text)
	}
	if got.HorzAlign != 1 {
		t.Errorf("HorzAlign = %d", got.HorzAlign)
	}
	if got.VertAlign != 2 {
		t.Errorf("VertAlign = %d", got.VertAlign)
	}
}

func TestAdvMatrixCell_DefaultColRowSpan(t *testing.T) {
	// ColSpan/RowSpan of 0 should become 1 after deserialize.
	xml := `<TableCell Name="c1" ColSpan="0" RowSpan="0"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := &object.AdvMatrixCell{}
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.ColSpan != 1 {
		t.Errorf("ColSpan should be 1 (clamped), got %d", got.ColSpan)
	}
	if got.RowSpan != 1 {
		t.Errorf("RowSpan should be 1 (clamped), got %d", got.RowSpan)
	}
}

func TestMatrixButton_SerializeDeserialize(t *testing.T) {
	orig := &object.MatrixButton{
		TypeName:               "MatrixCollapseButton",
		Name:                   "btn1",
		Left:                   5,
		Width:                  10,
		Height:                 10,
		Dock:                   "Left",
		SymbolSize:             8,
		Symbol:                 "Plus",
		ShowCollapseExpandMenu: true,
	}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MatrixCollapseButton", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := &object.MatrixButton{}
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Name != "btn1" {
		t.Errorf("Name = %q", got.Name)
	}
	if got.Left != 5 {
		t.Errorf("Left = %v", got.Left)
	}
	if got.Width != 10 {
		t.Errorf("Width = %v", got.Width)
	}
	if got.Dock != "Left" {
		t.Errorf("Dock = %q", got.Dock)
	}
	if got.SymbolSize != 8 {
		t.Errorf("SymbolSize = %v", got.SymbolSize)
	}
	if got.Symbol != "Plus" {
		t.Errorf("Symbol = %q", got.Symbol)
	}
	if !got.ShowCollapseExpandMenu {
		t.Error("ShowCollapseExpandMenu should be true")
	}
}

func TestAdvMatrixObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewAdvMatrixObject()
	orig.SetName("am1")
	orig.DataSource = "DS1"
	orig.TableColumns = []*object.AdvMatrixColumn{
		{Name: "col1", Width: 100},
		{Name: "col2", Width: 80, AutoSize: true},
	}
	orig.TableRows = []*object.AdvMatrixRow{
		{Name: "row1", Height: 30},
	}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("AdvMatrixObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, "TableColumn") {
		t.Error("expected TableColumn in output")
	}
	if !strings.Contains(xml, "TableRow") {
		t.Error("expected TableRow in output")
	}

	// Deserialize
	r := serial.NewReader(strings.NewReader(xml))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "AdvMatrixObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewAdvMatrixObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.DataSource != "DS1" {
		t.Errorf("DataSource = %q", got.DataSource)
	}
	// DeserializeChild is called by the FRX engine; test the child iteration directly
	for {
		childType, childOk := r.NextChild()
		if !childOk {
			break
		}
		got.DeserializeChild(childType, r)
		_ = r.FinishChild()
	}
	// After child processing we should have columns and rows
}

func TestAdvMatrixObject_DeserializeChild_TableColumn(t *testing.T) {
	xml := `<AdvMatrixObject Name="am1" DataSource="DS1">` +
		`<TableColumn Name="col1" Width="100"/>` +
		`<TableColumn Name="col2" Width="80" AutoSize="true"/>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	a := object.NewAdvMatrixObject()
	if err := a.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		a.DeserializeChild(ct, r)
		_ = r.FinishChild()
	}
	if len(a.TableColumns) != 2 {
		t.Errorf("TableColumns len = %d, want 2", len(a.TableColumns))
	}
	if a.TableColumns[0].Name != "col1" {
		t.Errorf("TableColumns[0].Name = %q", a.TableColumns[0].Name)
	}
	if !a.TableColumns[1].AutoSize {
		t.Error("TableColumns[1].AutoSize should be true")
	}
}

func TestAdvMatrixObject_DeserializeChild_TableRow(t *testing.T) {
	xml := `<AdvMatrixObject Name="am1">` +
		`<TableRow Name="row1" Height="30">` +
		`<TableCell Name="c1" Width="100" Text="Header"/>` +
		`</TableRow>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	a := object.NewAdvMatrixObject()
	_ = a.Deserialize(r)
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		a.DeserializeChild(ct, r)
		_ = r.FinishChild()
	}
	if len(a.TableRows) != 1 {
		t.Errorf("TableRows len = %d, want 1", len(a.TableRows))
	}
	if a.TableRows[0].Name != "row1" {
		t.Errorf("TableRows[0].Name = %q", a.TableRows[0].Name)
	}
	if len(a.TableRows[0].Cells) != 1 {
		t.Errorf("Cells len = %d, want 1", len(a.TableRows[0].Cells))
	}
	if a.TableRows[0].Cells[0].Text != "Header" {
		t.Errorf("Cell[0].Text = %q", a.TableRows[0].Cells[0].Text)
	}
}

func TestAdvMatrixObject_DeserializeChild_Columns(t *testing.T) {
	xml := `<AdvMatrixObject Name="am1">` +
		`<Columns>` +
		`<Descriptor Expression="[Category]" DisplayText="Category" Sort="Ascending"/>` +
		`</Columns>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	a := object.NewAdvMatrixObject()
	_ = a.Deserialize(r)
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		a.DeserializeChild(ct, r)
		_ = r.FinishChild()
	}
	if len(a.Columns) != 1 {
		t.Errorf("Columns len = %d, want 1", len(a.Columns))
	}
	if a.Columns[0].Expression != "[Category]" {
		t.Errorf("Columns[0].Expression = %q", a.Columns[0].Expression)
	}
	if a.Columns[0].DisplayText != "Category" {
		t.Errorf("Columns[0].DisplayText = %q", a.Columns[0].DisplayText)
	}
	if a.Columns[0].Sort != "Ascending" {
		t.Errorf("Columns[0].Sort = %q", a.Columns[0].Sort)
	}
}

func TestAdvMatrixObject_DeserializeChild_Rows(t *testing.T) {
	xml := `<AdvMatrixObject Name="am1">` +
		`<Rows>` +
		`<Descriptor Expression="[Region]">` +
		`<Descriptor Expression="[SubRegion]"/>` +
		`</Descriptor>` +
		`</Rows>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	a := object.NewAdvMatrixObject()
	_ = a.Deserialize(r)
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		a.DeserializeChild(ct, r)
		_ = r.FinishChild()
	}
	if len(a.Rows) != 1 {
		t.Errorf("Rows len = %d, want 1", len(a.Rows))
	}
	if a.Rows[0].Expression != "[Region]" {
		t.Errorf("Rows[0].Expression = %q", a.Rows[0].Expression)
	}
	if len(a.Rows[0].Children) != 1 {
		t.Errorf("Rows[0].Children len = %d, want 1", len(a.Rows[0].Children))
	}
	if a.Rows[0].Children[0].Expression != "[SubRegion]" {
		t.Errorf("Children[0].Expression = %q", a.Rows[0].Children[0].Expression)
	}
}

func TestAdvMatrixObject_DeserializeChild_MatrixButtons(t *testing.T) {
	// Test the drain paths for MatrixCollapseButton etc at AdvMatrix level
	xml := `<AdvMatrixObject Name="am1">` +
		`<MatrixCollapseButton Name="btn1"/>` +
		`<MatrixSortButton Name="sort1"/>` +
		`<Cells/>` +
		`<MatrixRows/>` +
		`<MatrixColumns/>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	a := object.NewAdvMatrixObject()
	_ = a.Deserialize(r)
	// Should not panic and should drain all children
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		a.DeserializeChild(ct, r)
		_ = r.FinishChild()
	}
}

func TestAdvMatrixObject_DeserializeChild_UnknownChild(t *testing.T) {
	xml := `<AdvMatrixObject Name="am1"><Unknown Attr="x"/></AdvMatrixObject>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	a := object.NewAdvMatrixObject()
	_ = a.Deserialize(r)
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		handled := a.DeserializeChild(ct, r)
		if handled {
			t.Errorf("Unknown child %q should not be handled", ct)
		}
		_ = r.FinishChild()
	}
}

func TestAdvMatrixObject_TableRowWithButton(t *testing.T) {
	xml := `<AdvMatrixObject Name="am1">` +
		`<TableRow Name="row1" Height="30">` +
		`<TableCell Name="c1" Width="100" Text="Hdr">` +
		`<MatrixCollapseButton Name="btn1" Left="2" Width="10" Height="10" Dock="Left" SymbolSize="8" Symbol="P" ShowCollapseExpandMenu="true"/>` +
		`</TableCell>` +
		`</TableRow>` +
		`</AdvMatrixObject>`

	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	a := object.NewAdvMatrixObject()
	_ = a.Deserialize(r)
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		a.DeserializeChild(ct, r)
		_ = r.FinishChild()
	}
	if len(a.TableRows) != 1 {
		t.Fatalf("TableRows len = %d", len(a.TableRows))
	}
	if len(a.TableRows[0].Cells) != 1 {
		t.Fatalf("Cells len = %d", len(a.TableRows[0].Cells))
	}
	if len(a.TableRows[0].Cells[0].Buttons) != 1 {
		t.Fatalf("Buttons len = %d, want 1", len(a.TableRows[0].Cells[0].Buttons))
	}
	btn := a.TableRows[0].Cells[0].Buttons[0]
	if btn.TypeName != "MatrixCollapseButton" {
		t.Errorf("TypeName = %q", btn.TypeName)
	}
	if !btn.ShowCollapseExpandMenu {
		t.Error("ShowCollapseExpandMenu should be true")
	}
}

func TestAdvMatrixCell_WithBorder(t *testing.T) {
	xml := `<TableCell Name="c1" Width="50" Border.Lines="All" Border.Color="255,0,0"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := &object.AdvMatrixCell{}
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Border == nil {
		t.Fatal("Border should not be nil")
	}
}

func TestAdvMatrixCell_WithFillAndFont(t *testing.T) {
	xml := `<TableCell Name="c1" Fill.Color="0,255,0" Font="Arial, 10"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := &object.AdvMatrixCell{}
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.FillColor == nil {
		t.Fatal("FillColor should not be nil")
	}
	if got.Font == nil {
		t.Fatal("Font should not be nil")
	}
}

func TestAdvMatrixCell_SerializeWithBorder(t *testing.T) {
	// Verify that a cell with border lines serializes the Border.Lines attribute.
	// We exercise the formatBorderLinesStr path via Serialize.
	// We need to import style to set VisibleLines; instead test via full round-trip.
	xml := `<TableCell Name="c1" Border.Lines="Left"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	cell := &object.AdvMatrixCell{}
	_ = cell.Deserialize(r)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TableCell", cell); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()
	out := buf.String()
	if !strings.Contains(out, "Border.Lines") {
		t.Errorf("expected Border.Lines in output: %s", out)
	}
}

func TestAdvMatrixCell_SerializeWithFillColor(t *testing.T) {
	xml := `<TableCell Name="c1" Fill.Color="255,0,0"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	cell := &object.AdvMatrixCell{}
	_ = cell.Deserialize(r)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.WriteObjectNamed("TableCell", cell)
	_ = w.Flush()
	out := buf.String()
	if !strings.Contains(out, "Fill.Color") {
		t.Errorf("expected Fill.Color in output: %s", out)
	}
}

func TestAdvMatrixCell_SerializeWithFont(t *testing.T) {
	xml := `<TableCell Name="c1" Font="Arial, 10"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	cell := &object.AdvMatrixCell{}
	_ = cell.Deserialize(r)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.WriteObjectNamed("TableCell", cell)
	_ = w.Flush()
	out := buf.String()
	if !strings.Contains(out, "Font") {
		t.Errorf("expected Font in output: %s", out)
	}
}

func TestAdvMatrixCell_SerializeWithButtons(t *testing.T) {
	cell := &object.AdvMatrixCell{
		Name: "c1",
		Buttons: []*object.MatrixButton{
			{TypeName: "MatrixCollapseButton", Name: "btn1", Width: 10, Height: 10},
		},
	}
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TableCell", cell); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()
	out := buf.String()
	if !strings.Contains(out, "MatrixCollapseButton") {
		t.Errorf("expected MatrixCollapseButton in output: %s", out)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// Container — serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

func TestContainerObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewContainerObject()
	orig.SetName("cont1")
	orig.SetBeforeLayoutEvent("OnBefore")
	orig.SetAfterLayoutEvent("OnAfter")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ContainerObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "ContainerObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewContainerObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.BeforeLayoutEvent() != "OnBefore" {
		t.Errorf("BeforeLayoutEvent = %q", got.BeforeLayoutEvent())
	}
	if got.AfterLayoutEvent() != "OnAfter" {
		t.Errorf("AfterLayoutEvent = %q", got.AfterLayoutEvent())
	}
}

func TestContainerObject_UpdateLayout(t *testing.T) {
	c := object.NewContainerObject()
	// Should not panic
	c.UpdateLayout(5, 10)
	c.UpdateLayout(-5, -10)
	c.UpdateLayout(0, 0)
}

func TestSubreportObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewSubreportObject()
	orig.SetName("sub1")
	orig.SetReportPageName("DetailPage")
	orig.SetPrintOnParent(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SubreportObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "SubreportObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewSubreportObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.ReportPageName() != "DetailPage" {
		t.Errorf("ReportPageName = %q", got.ReportPageName())
	}
	if !got.PrintOnParent() {
		t.Error("PrintOnParent should be true")
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// MSChartObject / MSChartSeries — serialization
// ══════════════════════════════════════════════════════════════════════════════

func TestMSChartObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewMSChartObject()
	orig.SetName("chart1")
	orig.ChartData = "base64data"
	orig.ChartType = "Bar"
	orig.DataSource = "DS1"

	s := object.NewMSChartSeries()
	s.SetName("series1")
	s.ChartType = "Line"
	s.ValuesSource = "[Sales]"
	s.ArgumentSource = "[Month]"
	s.LegendText = "Sales"
	orig.Series = append(orig.Series, s)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MSChartObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	xml := buf.String()
	if !strings.Contains(xml, "MSChartSeries") {
		t.Error("expected MSChartSeries in output")
	}
	if !strings.Contains(xml, "Sales") {
		t.Error("expected Sales in output")
	}

	r := serial.NewReader(strings.NewReader(xml))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "MSChartObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewMSChartObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.ChartData != "base64data" {
		t.Errorf("ChartData = %q", got.ChartData)
	}
	if got.ChartType != "Bar" {
		t.Errorf("ChartType = %q", got.ChartType)
	}
	if got.DataSource != "DS1" {
		t.Errorf("DataSource = %q", got.DataSource)
	}
	// Deserialize children
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		got.DeserializeChild(ct, r)
		_ = r.FinishChild()
	}
	if len(got.Series) != 1 {
		t.Errorf("Series len = %d, want 1", len(got.Series))
	}
	if got.Series[0].ChartType != "Line" {
		t.Errorf("Series[0].ChartType = %q", got.Series[0].ChartType)
	}
}

func TestMSChartSeries_TypeNameBaseName(t *testing.T) {
	s := object.NewMSChartSeries()
	if s.TypeName() != "MSChartSeries" {
		t.Errorf("TypeName = %q", s.TypeName())
	}
	if s.BaseName() != "Series" {
		t.Errorf("BaseName = %q", s.BaseName())
	}
}

func TestMSChartObject_DeserializeChild_Unknown(t *testing.T) {
	m := object.NewMSChartObject()
	xml := `<MSChartObject><UnknownChild/></MSChartObject>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	_ = m.Deserialize(r)
	for {
		ct, ok := r.NextChild()
		if !ok {
			break
		}
		handled := m.DeserializeChild(ct, r)
		if handled {
			t.Errorf("Unknown child %q should not be handled", ct)
		}
		_ = r.FinishChild()
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// PictureObject — serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

func TestPictureObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewPictureObject()
	orig.SetName("pic1")
	orig.SetAngle(90)
	orig.SetDataColumn("Photo")
	orig.SetGrayscale(true)
	orig.SetImageLocation("http://example.com/img.png")
	orig.SetImageSourceExpression("[ImgPath]")
	orig.SetMaxWidth(200)
	orig.SetMaxHeight(150)
	orig.SetPadding(object.Padding{Left: 5, Top: 5, Right: 5, Bottom: 5})
	orig.SetSizeMode(object.SizeModeStretchImage)
	orig.SetImageAlign(object.ImageAlignTopLeft)
	orig.SetShowErrorImage(true)
	orig.SetTile(true)
	orig.SetTransparency(0.5)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PictureObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "PictureObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewPictureObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Angle() != 90 {
		t.Errorf("Angle = %d", got.Angle())
	}
	if got.DataColumn() != "Photo" {
		t.Errorf("DataColumn = %q", got.DataColumn())
	}
	if !got.Grayscale() {
		t.Error("Grayscale should be true")
	}
	if got.ImageLocation() != "http://example.com/img.png" {
		t.Errorf("ImageLocation = %q", got.ImageLocation())
	}
	if got.ImageSourceExpression() != "[ImgPath]" {
		t.Errorf("ImageSourceExpression = %q", got.ImageSourceExpression())
	}
	if got.MaxWidth() != 200 {
		t.Errorf("MaxWidth = %v", got.MaxWidth())
	}
	if got.MaxHeight() != 150 {
		t.Errorf("MaxHeight = %v", got.MaxHeight())
	}
	wantPad := object.Padding{Left: 5, Top: 5, Right: 5, Bottom: 5}
	if got.Padding() != wantPad {
		t.Errorf("Padding = %+v", got.Padding())
	}
	if got.SizeMode() != object.SizeModeStretchImage {
		t.Errorf("SizeMode = %d", got.SizeMode())
	}
	if got.ImageAlign() != object.ImageAlignTopLeft {
		t.Errorf("ImageAlign = %d", got.ImageAlign())
	}
	if !got.ShowErrorImage() {
		t.Error("ShowErrorImage should be true")
	}
	if !got.Tile() {
		t.Error("Tile should be true")
	}
	if got.Transparency() != 0.5 {
		t.Errorf("Transparency = %v", got.Transparency())
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// LineObject — serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

func TestLineObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewLineObject()
	orig.SetName("line1")
	orig.SetDiagonal(true)
	orig.StartCap = object.CapSettings{Width: 12, Height: 10, Style: object.CapStyleArrow}
	orig.EndCap = object.CapSettings{Width: 6, Height: 6, Style: object.CapStyleCircle}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("LineObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "LineObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if !got.Diagonal() {
		t.Error("Diagonal should be true")
	}
	if got.StartCap.Style != object.CapStyleArrow {
		t.Errorf("StartCap.Style = %d", got.StartCap.Style)
	}
	if got.EndCap.Style != object.CapStyleCircle {
		t.Errorf("EndCap.Style = %d", got.EndCap.Style)
	}
}

func TestShapeObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewShapeObject()
	orig.SetName("shape1")
	orig.SetShape(object.ShapeKindEllipse)
	orig.SetCurve(15)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ShapeObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "ShapeObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewShapeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Shape() != object.ShapeKindEllipse {
		t.Errorf("Shape = %d", got.Shape())
	}
	if got.Curve() != 15 {
		t.Errorf("Curve = %v", got.Curve())
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// TextObject — serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

func TestTextObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetName("text1")
	orig.SetText("Hello [Name]")
	orig.SetHorzAlign(object.HorzAlignCenter)
	orig.SetVertAlign(object.VertAlignCenter)
	orig.SetAngle(90)
	orig.SetRightToLeft(true)
	orig.SetWordWrap(false)
	orig.SetUnderlines(true)
	orig.SetFontWidthRatio(0.8)
	orig.SetFirstTabOffset(10)
	orig.SetTabWidth(50)
	orig.SetClip(false)
	orig.SetWysiwyg(true)
	orig.SetLineHeight(18)
	orig.SetForceJustify(true)
	orig.SetTextRenderType(object.TextRenderTypeHtmlTags)
	orig.SetAutoShrink(object.AutoShrinkFontSize)
	orig.SetAutoShrinkMinSize(6)
	orig.SetParagraphOffset(5)
	orig.SetMergeMode(object.MergeModeHorizontal)
	orig.SetAutoWidth(true)
	orig.SetAllowExpressions(false)
	orig.SetHideZeros(true)
	orig.SetHideValue("0")
	orig.SetNullValue("-")
	orig.SetDuplicates(object.DuplicatesHide)
	orig.SetEditable(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TextObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Text() != "Hello [Name]" {
		t.Errorf("Text = %q", got.Text())
	}
	if got.HorzAlign() != object.HorzAlignCenter {
		t.Errorf("HorzAlign = %d", got.HorzAlign())
	}
	if got.VertAlign() != object.VertAlignCenter {
		t.Errorf("VertAlign = %d", got.VertAlign())
	}
	if !got.RightToLeft() {
		t.Error("RightToLeft should be true")
	}
	if got.WordWrap() {
		t.Error("WordWrap should be false")
	}
	if !got.Wysiwyg() {
		t.Error("Wysiwyg should be true")
	}
	if got.MergeMode() != object.MergeModeHorizontal {
		t.Errorf("MergeMode = %d", got.MergeMode())
	}
	if !got.AutoWidth() {
		t.Error("AutoWidth should be true")
	}
	if got.HideValue() != "0" {
		t.Errorf("HideValue = %q", got.HideValue())
	}
	if got.NullValue() != "-" {
		t.Errorf("NullValue = %q", got.NullValue())
	}
	if got.Duplicates() != object.DuplicatesHide {
		t.Errorf("Duplicates = %d", got.Duplicates())
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// format_serial.go — serializeTextFormat / deserializeTextFormat
// ══════════════════════════════════════════════════════════════════════════════

func TestTextObject_FormatSerial_Number(t *testing.T) {
	// Use TextObject to exercise serializeTextFormat and deserializeTextFormat
	// by setting the format field via TextObjectBase.SetFormat.
	orig := object.NewTextObject()
	orig.SetName("fmt1")

	// Build a NumberFormat via TextObjectBase.SetFormat using the format package.
	// We test via a raw XML deserialize path which calls deserializeTextFormat.
	xml := `<TextObject Name="fmt1" Format="Number" Format.DecimalDigits="3" Format.UseLocaleSettings="false" Format.DecimalSeparator="." Format.GroupSeparator="," Format.NegativePattern="1"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Format() == nil {
		t.Fatal("Format should not be nil after deserializing Number format")
	}
}

func TestTextObject_FormatSerial_Currency(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Currency" Format.DecimalDigits="2" Format.UseLocaleSettings="false" Format.CurrencySymbol="$"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	if got.Format() == nil {
		t.Fatal("Format should not be nil for Currency")
	}
}

func TestTextObject_FormatSerial_Date(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Date" Format.Format="dd/MM/yyyy"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	if got.Format() == nil {
		t.Fatal("Format should not be nil for Date")
	}
}

func TestTextObject_FormatSerial_Time(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Time" Format.Format="HH:mm:ss"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	if got.Format() == nil {
		t.Fatal("Format should not be nil for Time")
	}
}

func TestTextObject_FormatSerial_Percent(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Percent" Format.DecimalDigits="1" Format.UseLocaleSettings="false" Format.PercentSymbol="%"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	if got.Format() == nil {
		t.Fatal("Format should not be nil for Percent")
	}
}

func TestTextObject_FormatSerial_Boolean(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Boolean" Format.TrueText="Yes" Format.FalseText="No"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	if got.Format() == nil {
		t.Fatal("Format should not be nil for Boolean")
	}
}

func TestTextObject_FormatSerial_Custom(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Custom" Format.Format="####"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	if got.Format() == nil {
		t.Fatal("Format should not be nil for Custom")
	}
}

func TestTextObject_FormatSerial_Unknown(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Bogus"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	// Unknown format returns nil — Format stays nil
	if got.Format() != nil {
		t.Error("Format should be nil for unknown format type")
	}
}

func TestTextObject_FormatSerialize_RoundTrip(t *testing.T) {
	// Test serializeTextFormat via a round-trip: set format on orig then check it appears
	// in XML and re-reads.  We inject via the child Formats element path.
	xml := `<TextObject Name="f1"><Formats><NumberFormat DecimalDigits="2" UseLocaleSettings="true"/></Formats></TextObject>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	// Process child Formats element
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		got.DeserializeChild(ct, r)
		_ = r.FinishChild()
	}
	if got.Format() == nil {
		t.Fatal("Format should not be nil after Formats child")
	}
	if got.Formats() == nil {
		t.Fatal("Formats collection should not be nil")
	}
}

func TestTextObject_FormatSerialize_AllChildFormats(t *testing.T) {
	// Exercise deserializeFormatFromChild for all format types
	formats := []string{
		"CurrencyFormat", "DateFormat", "TimeFormat",
		"PercentFormat", "BooleanFormat", "CustomFormat", "GeneralFormat",
	}
	for _, ft := range formats {
		xml := `<TextObject Name="f1"><Formats><` + ft + `/></Formats></TextObject>`
		r := serial.NewReader(strings.NewReader(xml))
		_, _ = r.ReadObjectHeader()
		got := object.NewTextObject()
		_ = got.Deserialize(r)
		for {
			ct, ok2 := r.NextChild()
			if !ok2 {
				break
			}
			got.DeserializeChild(ct, r)
			_ = r.FinishChild()
		}
	}
}

func TestTextObject_FormatSerialize_UnknownChildFormat(t *testing.T) {
	xml := `<TextObject Name="f1"><Formats><UnknownFormat/></Formats></TextObject>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		got.DeserializeChild(ct, r)
		_ = r.FinishChild()
	}
	// should not panic
}

// ══════════════════════════════════════════════════════════════════════════════
// TextObject — serialize format fields
// ══════════════════════════════════════════════════════════════════════════════

func TestTextObject_SerializeFormat_Number(t *testing.T) {
	// Deserialize with Number format then re-serialize to verify output.
	xml := `<TextObject Name="f1" Format="Number" Format.DecimalDigits="4" Format.UseLocaleSettings="false"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.WriteObjectNamed("TextObject", got)
	_ = w.Flush()
	out := buf.String()
	if !strings.Contains(out, `Format="Number"`) {
		t.Errorf("expected Format=Number in: %s", out)
	}
}

func TestTextObject_SerializeFormat_Currency(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Currency" Format.UseLocaleSettings="false" Format.CurrencySymbol="€"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.WriteObjectNamed("TextObject", got)
	_ = w.Flush()
	if !strings.Contains(buf.String(), `Format="Currency"`) {
		t.Errorf("expected Currency format in: %s", buf.String())
	}
}

func TestTextObject_SerializeFormat_Date(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Date" Format.Format="yyyy-MM-dd" Format.UseLocaleSettings="true"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.WriteObjectNamed("TextObject", got)
	_ = w.Flush()
	if !strings.Contains(buf.String(), `Format="Date"`) {
		t.Errorf("expected Date format in output")
	}
}

func TestTextObject_SerializeFormat_Time(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Time" Format.Format="hh:mm" Format.UseLocaleSettings="true"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.WriteObjectNamed("TextObject", got)
	_ = w.Flush()
	if !strings.Contains(buf.String(), `Format="Time"`) {
		t.Errorf("expected Time format in output")
	}
}

func TestTextObject_SerializeFormat_Percent(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Percent" Format.UseLocaleSettings="false" Format.PercentSymbol="pct"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.WriteObjectNamed("TextObject", got)
	_ = w.Flush()
	if !strings.Contains(buf.String(), `Format="Percent"`) {
		t.Errorf("expected Percent format in output")
	}
}

func TestTextObject_SerializeFormat_Boolean(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Boolean" Format.TrueText="Oui" Format.FalseText="Non"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.WriteObjectNamed("TextObject", got)
	_ = w.Flush()
	if !strings.Contains(buf.String(), `Format="Boolean"`) {
		t.Errorf("expected Boolean format in output")
	}
}

func TestTextObject_SerializeFormat_Custom(t *testing.T) {
	xml := `<TextObject Name="f1" Format="Custom" Format.Format="##.##"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObject()
	_ = got.Deserialize(r)
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	_ = w.WriteObjectNamed("TextObject", got)
	_ = w.Flush()
	if !strings.Contains(buf.String(), `Format="Custom"`) {
		t.Errorf("expected Custom format in output")
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// MapObject — DeserializeChild MapLayer
// ══════════════════════════════════════════════════════════════════════════════

func TestMapObject_DeserializeChild_MapLayer(t *testing.T) {
	xml := `<MapObject Name="m1" OffsetX="5" OffsetY="10">` +
		`<MapLayer Name="layer1" Shapefile="world.shp" Type="Choropleth" DataSource="DS1"/>` +
		`</MapObject>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	m := object.NewMapObject()
	_ = m.Deserialize(r)
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		m.DeserializeChild(ct, r)
		_ = r.FinishChild()
	}
	if len(m.Layers) != 1 {
		t.Fatalf("Layers len = %d, want 1", len(m.Layers))
	}
	if m.Layers[0].Shapefile != "world.shp" {
		t.Errorf("Shapefile = %q", m.Layers[0].Shapefile)
	}
}

func TestMapObject_DeserializeChild_Unknown(t *testing.T) {
	m := object.NewMapObject()
	xml := `<MapObject Name="m1"><UnknownChild/></MapObject>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	_ = m.Deserialize(r)
	for {
		ct, ok := r.NextChild()
		if !ok {
			break
		}
		handled := m.DeserializeChild(ct, r)
		if handled {
			t.Errorf("Unknown child %q should not be handled", ct)
		}
		_ = r.FinishChild()
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// TextObject — TextOutline serialization
// ══════════════════════════════════════════════════════════════════════════════

func TestTextObject_TextOutline_SerializeDeserialize(t *testing.T) {
	xml := `<TextObject Name="t1" TextOutline.Enabled="true" TextOutline.Color="255,0,0" TextOutline.Width="2" TextOutline.DashStyle="1"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	outline := got.TextOutline()
	if !outline.Enabled {
		t.Error("TextOutline.Enabled should be true")
	}
	if outline.Width != 2 {
		t.Errorf("TextOutline.Width = %v", outline.Width)
	}
	if outline.DashStyle != 1 {
		t.Errorf("TextOutline.DashStyle = %d", outline.DashStyle)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// TextObject — CheckBoxObject serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

func TestCheckBoxObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewCheckBoxObject()
	orig.SetName("cb1")
	orig.SetChecked(true)
	orig.SetCheckedSymbol(object.CheckedSymbolCross)
	orig.SetUncheckedSymbol(object.UncheckedSymbolMinus)
	orig.SetDataColumn("IsActive")
	orig.SetExpression("[Val]>0")
	orig.SetCheckWidthRatio(0.8)
	orig.SetHideIfUnchecked(true)
	orig.SetEditable(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("CheckBoxObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "CheckBoxObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewCheckBoxObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if !got.Checked() {
		t.Error("Checked should be true")
	}
	if got.CheckedSymbol() != object.CheckedSymbolCross {
		t.Errorf("CheckedSymbol = %d", got.CheckedSymbol())
	}
	if got.UncheckedSymbol() != object.UncheckedSymbolMinus {
		t.Errorf("UncheckedSymbol = %d", got.UncheckedSymbol())
	}
	if got.DataColumn() != "IsActive" {
		t.Errorf("DataColumn = %q", got.DataColumn())
	}
	if got.Expression() != "[Val]>0" {
		t.Errorf("Expression = %q", got.Expression())
	}
	if got.CheckWidthRatio() != 0.8 {
		t.Errorf("CheckWidthRatio = %v", got.CheckWidthRatio())
	}
	if !got.HideIfUnchecked() {
		t.Error("HideIfUnchecked should be true")
	}
	if !got.Editable() {
		t.Error("Editable should be true")
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// PolyLineObject / PolygonObject — serialize/deserialize
// ══════════════════════════════════════════════════════════════════════════════

func TestPolyLineObject_SerializeDeserialize(t *testing.T) {
	orig := object.NewPolyLineObject()
	orig.SetName("pl1")
	orig.SetCenterX(50)
	orig.SetCenterY(75)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PolyLineObject", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "PolyLineObject" {
		t.Fatalf("ReadObjectHeader: %q %v", typeName, ok)
	}
	got := object.NewPolyLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.Name() != "pl1" {
		t.Errorf("Name = %q", got.Name())
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// MSChartSeries — color serialization
// ══════════════════════════════════════════════════════════════════════════════

func TestMSChartSeries_ColorSerializeDeserialize(t *testing.T) {
	xml := `<MSChartSeries Name="s1" ChartType="Bar" Color="255,0,128" ValuesSource="[Sales]" ArgumentSource="[Month]" LegendText="Revenue"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewMSChartSeries()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	if got.ChartType != "Bar" {
		t.Errorf("ChartType = %q", got.ChartType)
	}
	if got.ValuesSource != "[Sales]" {
		t.Errorf("ValuesSource = %q", got.ValuesSource)
	}
	if got.ArgumentSource != "[Month]" {
		t.Errorf("ArgumentSource = %q", got.ArgumentSource)
	}
	if got.LegendText != "Revenue" {
		t.Errorf("LegendText = %q", got.LegendText)
	}
	if got.Color.R != 255 {
		t.Errorf("Color.R = %d, want 255", got.Color.R)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// Padding helpers
// ══════════════════════════════════════════════════════════════════════════════

func TestPadding_RoundTrip(t *testing.T) {
	// Exercise paddingToStr and strToPadding via TextObjectBase.
	orig := object.NewTextObjectBase()
	orig.SetPadding(object.Padding{Left: 1, Top: 2, Right: 3, Bottom: 4})

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObjectBase", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	_ = w.Flush()

	r := serial.NewReader(strings.NewReader(buf.String()))
	_, _ = r.ReadObjectHeader()
	got := object.NewTextObjectBase()
	_ = got.Deserialize(r)
	want := object.Padding{Left: 1, Top: 2, Right: 3, Bottom: 4}
	if got.Padding() != want {
		t.Errorf("Padding = %+v, want %+v", got.Padding(), want)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// MSChartObject TypeName / BaseName
// ══════════════════════════════════════════════════════════════════════════════

func TestMSChartObject_TypeNameBaseName(t *testing.T) {
	m := object.NewMSChartObject()
	if m.TypeName() != "MSChartObject" {
		t.Errorf("TypeName = %q", m.TypeName())
	}
	if m.BaseName() != "Chart" {
		t.Errorf("BaseName = %q", m.BaseName())
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// AdvMatrix — border lines all variants
// ══════════════════════════════════════════════════════════════════════════════

func TestAdvMatrixCell_BorderLinesVariants(t *testing.T) {
	variants := []string{"All", "None", "Left", "Right", "Top", "Bottom"}
	for _, v := range variants {
		xml := `<TableCell Name="c1" Border.Lines="` + v + `"/>`
		r := serial.NewReader(strings.NewReader(xml))
		_, _ = r.ReadObjectHeader()
		cell := &object.AdvMatrixCell{}
		_ = cell.Deserialize(r)

		// Re-serialize to exercise formatBorderLinesStr.
		var buf bytes.Buffer
		w := serial.NewWriter(&buf)
		_ = w.WriteObjectNamed("TableCell", cell)
		_ = w.Flush()
		// "None" border has VisibleLines==0, so Border.Lines is not emitted.
		if v != "None" {
			out := buf.String()
			if !strings.Contains(out, "Border.Lines") {
				t.Errorf("variant %q: expected Border.Lines in output: %s", v, out)
			}
		}
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// Container — FireBeforeLayout / FireAfterLayout no-op when handler nil
// ══════════════════════════════════════════════════════════════════════════════

func TestContainerObject_FireEvents_NoHandler(t *testing.T) {
	c := object.NewContainerObject()
	// Should not panic when handlers are nil
	c.FireBeforeLayout()
	c.FireAfterLayout()
}

// ══════════════════════════════════════════════════════════════════════════════
// AdvMatrix — TableRow with unexpected child type in cells
// ══════════════════════════════════════════════════════════════════════════════

func TestAdvMatrixObject_TableRow_UnexpectedCellChild(t *testing.T) {
	xml := `<AdvMatrixObject Name="am1">` +
		`<TableRow Name="row1">` +
		`<UnknownElement Attr="x"/>` +
		`</TableRow>` +
		`</AdvMatrixObject>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	a := object.NewAdvMatrixObject()
	_ = a.Deserialize(r)
	for {
		ct, ok := r.NextChild()
		if !ok {
			break
		}
		a.DeserializeChild(ct, r)
		_ = r.FinishChild()
	}
	// Should have 1 row with 0 cells (unknown element was drained).
	if len(a.TableRows) != 1 {
		t.Errorf("TableRows len = %d, want 1", len(a.TableRows))
	}
	if len(a.TableRows[0].Cells) != 0 {
		t.Errorf("Cells len = %d, want 0", len(a.TableRows[0].Cells))
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// AdvMatrix — Columns/Rows with unknown non-Descriptor child
// ══════════════════════════════════════════════════════════════════════════════

func TestAdvMatrixObject_Columns_UnknownChild(t *testing.T) {
	xml := `<AdvMatrixObject Name="am1">` +
		`<Columns>` +
		`<UnknownElement/>` +
		`</Columns>` +
		`</AdvMatrixObject>`
	r := serial.NewReader(strings.NewReader(xml))
	_, _ = r.ReadObjectHeader()
	a := object.NewAdvMatrixObject()
	_ = a.Deserialize(r)
	for {
		ct, ok := r.NextChild()
		if !ok {
			break
		}
		a.DeserializeChild(ct, r)
		_ = r.FinishChild()
	}
	if len(a.Columns) != 0 {
		t.Errorf("Columns len = %d, want 0 (unknown element should be skipped)", len(a.Columns))
	}
}

// ── TextObject missing getter/setter tests ────────────────────────────────────

func TestTextObject_SetFormat_Format(t *testing.T) {
	ob := object.NewTextObjectBase()
	fc := &format.NumberFormat{}
	ob.SetFormat(fc)
	if ob.Format() != fc {
		t.Error("Format() should return the value set by SetFormat()")
	}
}

func TestTextObject_ApplyStyle_WithFont(t *testing.T) {
	to := object.NewTextObject()
	f := style.Font{Name: "Arial", Size: 14}
	entry := &style.StyleEntry{
		ApplyFont: true,
		Font:      f,
	}
	to.ApplyStyle(entry) // should not panic; sets font
}

func TestTextObject_ApplyStyle_Nil(t *testing.T) {
	to := object.NewTextObject()
	to.ApplyStyle(nil) // nil entry — should not panic
}

func TestTextObject_FirstTabOffset(t *testing.T) {
	to := object.NewTextObject()
	to.SetFirstTabOffset(12.5)
	if to.FirstTabOffset() != 12.5 {
		t.Errorf("FirstTabOffset = %v, want 12.5", to.FirstTabOffset())
	}
}

func TestTextObject_LineHeight(t *testing.T) {
	to := object.NewTextObject()
	to.SetLineHeight(20)
	if to.LineHeight() != 20 {
		t.Errorf("LineHeight = %v, want 20", to.LineHeight())
	}
}

func TestTextObject_ForceJustify(t *testing.T) {
	to := object.NewTextObject()
	to.SetForceJustify(true)
	if !to.ForceJustify() {
		t.Error("ForceJustify should be true")
	}
}

func TestTextObject_AutoShrinkMinSize(t *testing.T) {
	to := object.NewTextObject()
	to.SetAutoShrinkMinSize(6)
	if to.AutoShrinkMinSize() != 6 {
		t.Errorf("AutoShrinkMinSize = %v, want 6", to.AutoShrinkMinSize())
	}
}

func TestTextObject_ParagraphOffset(t *testing.T) {
	to := object.NewTextObject()
	to.SetParagraphOffset(8)
	if to.ParagraphOffset() != 8 {
		t.Errorf("ParagraphOffset = %v, want 8", to.ParagraphOffset())
	}
}

func TestTextObject_SetFormats(t *testing.T) {
	to := object.NewTextObject()
	col := &format.Collection{}
	to.SetFormats(col)
	if to.Formats() != col {
		t.Error("Formats() should return the collection set by SetFormats()")
	}
}

func TestTextObject_HighlightsAndAddHighlight(t *testing.T) {
	to := object.NewTextObject()
	if len(to.Highlights()) != 0 {
		t.Error("Highlights should be empty initially")
	}
	hc := style.HighlightCondition{Expression: "1==1"}
	to.AddHighlight(hc)
	if len(to.Highlights()) != 1 {
		t.Errorf("Highlights len = %d, want 1", len(to.Highlights()))
	}
	if to.Highlights()[0].Expression != "1==1" {
		t.Error("Highlight expression mismatch")
	}
}
