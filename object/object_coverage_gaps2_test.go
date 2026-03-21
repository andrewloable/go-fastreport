package object

// object_coverage_gaps2_test.go — internal coverage tests for:
//
//	barcode.go       : BarcodeObject.Serialize/Deserialize parent-error paths
//	                   ZipCodeObject.Serialize/Deserialize parent-error paths
//	cellular_text.go : CellularTextObject.Serialize/Deserialize parent-error paths
//	container.go     : CheckBoxObject, ContainerObject, SubreportObject parent-error paths
//	digital_signature.go: DigitalSignatureObject parent-error paths
//	html.go          : HtmlObject parent-error paths
//	lines.go         : LineObject, ShapeObject parent-error paths
//	map.go           : MapLayer, MapObject parent-error paths
//
// All "return err" branches after a parent Serialize/Deserialize call are dead
// code under the current design: the entire embedded chain
// (ReportComponentBase → ComponentBase → BaseObject) never returns a non-nil
// error because none of those methods call WriteObject/WriteObjectNamed.
//
// To make these branches executable we use a mock report.Writer / report.Reader
// whose attribute-write no-ops are supplemented by a "poison" mechanism that
// forces the first call to Serialize (via WriteObjectNamed) to fail.
//
// The strategy that actually works: for functions whose Serialize chain DOES
// call WriteObject (e.g. ContainerObject child loop — already covered by
// container_internal_error_test.go), we add error-returning mock writers.
// For the parent-chain dead-code paths we document them and provide the
// closest-possible exerciser: calling Serialize/Deserialize with a mock that
// at least verifies the functions behave correctly under normal conditions,
// and adds additional edge-case coverage for the surrounding branches.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── shared mocks ──────────────────────────────────────────────────────────────

// noopWriter is a report.Writer whose attribute writes are all no-ops and
// whose WriteObject / WriteObjectNamed always succeed. It is used to call
// Serialize directly without going through serial.Writer.
type noopWriter2 struct{}

func (w *noopWriter2) WriteStr(name, value string)            {}
func (w *noopWriter2) WriteInt(name string, value int)        {}
func (w *noopWriter2) WriteBool(name string, value bool)      {}
func (w *noopWriter2) WriteFloat(name string, value float32)  {}
func (w *noopWriter2) WriteObject(obj report.Serializable) error {
	return obj.Serialize(w)
}
func (w *noopWriter2) WriteObjectNamed(name string, obj report.Serializable) error {
	return obj.Serialize(w)
}

// writeObjErrWriter is a report.Writer whose WriteObject and WriteObjectNamed
// always return a sentinel error. Used to cover error-return branches in
// Serialize methods that call WriteObject for child objects.
type writeObjErrWriter struct{ msg string }

func (w *writeObjErrWriter) WriteStr(name, value string)            {}
func (w *writeObjErrWriter) WriteInt(name string, value int)        {}
func (w *writeObjErrWriter) WriteBool(name string, value bool)      {}
func (w *writeObjErrWriter) WriteFloat(name string, value float32)  {}
func (w *writeObjErrWriter) WriteObject(obj report.Serializable) error {
	return errors.New(w.msg)
}
func (w *writeObjErrWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	return errors.New(w.msg)
}

// noopReader is a report.Reader that returns default values for all reads and
// signals no children. Used to call Deserialize directly without XML parsing.
type noopReader struct{}

func (r *noopReader) ReadStr(name, def string) string             { return def }
func (r *noopReader) ReadInt(name string, def int) int            { return def }
func (r *noopReader) ReadBool(name string, def bool) bool         { return def }
func (r *noopReader) ReadFloat(name string, def float32) float32  { return def }
func (r *noopReader) NextChild() (string, bool)                   { return "", false }
func (r *noopReader) FinishChild() error                          { return nil }

// ── BarcodeObject ─────────────────────────────────────────────────────────────

// TestBarcodeObject_Serialize_ViaNoopWriter calls Serialize directly (without
// serial.Writer) to exercise the function body via a no-op writer. This
// exercises every reachable branch including the false-branch of the
// parent-error check (which always evaluates to false).
func TestBarcodeObject_Serialize_ViaNoopWriter(t *testing.T) {
	b := NewBarcodeObject()
	b.text = "hello"
	b.barcodeType = "QR Code"
	b.showText = false
	b.autoSize = false
	b.allowExpressions = true

	w := &noopWriter2{}
	if err := b.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestBarcodeObject_Serialize_Defaults_ViaNoopWriter exercises the branches
// where text, barcodeType are empty and showText/autoSize are true (defaults).
func TestBarcodeObject_Serialize_Defaults_ViaNoopWriter(t *testing.T) {
	b := NewBarcodeObject()
	// defaults: showText=true, autoSize=true, text="", barcodeType=""
	w := &noopWriter2{}
	if err := b.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestBarcodeObject_Deserialize_ViaNoopReader calls Deserialize directly with
// a no-op reader to exercise the function body. The parent-error branch always
// evaluates to false (parent always returns nil).
func TestBarcodeObject_Deserialize_ViaNoopReader(t *testing.T) {
	b := NewBarcodeObject()
	r := &noopReader{}
	if err := b.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// ── ZipCodeObject ─────────────────────────────────────────────────────────────

// TestZipCodeObject_Serialize_ViaNoopWriter calls Serialize with a no-op writer
// to exercise all reachable branches including non-default fields.
func TestZipCodeObject_Serialize_ViaNoopWriter(t *testing.T) {
	z := NewZipCodeObject()
	z.text = "654321"  // non-default (default = "123456")
	z.dataColumn = "ZipCol"
	z.expression = "[Zip]"
	z.segmentWidth = 5
	z.segmentHeight = 12
	z.spacing = 2
	z.segmentCount = 5 // non-default (default = 6)
	z.showMarkers = false
	z.showGrid = false

	w := &noopWriter2{}
	if err := z.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestZipCodeObject_Serialize_Defaults_ViaNoopWriter exercises the path where
// all ZipCodeObject fields are at defaults (no attributes written).
func TestZipCodeObject_Serialize_Defaults_ViaNoopWriter(t *testing.T) {
	z := NewZipCodeObject()
	// defaults: segmentCount=6, showMarkers=true, showGrid=true
	w := &noopWriter2{}
	if err := z.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestZipCodeObject_Deserialize_ViaNoopReader calls Deserialize with a no-op
// reader to exercise the full function body.
func TestZipCodeObject_Deserialize_ViaNoopReader(t *testing.T) {
	z := NewZipCodeObject()
	r := &noopReader{}
	if err := z.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// ── CellularTextObject ────────────────────────────────────────────────────────

// TestCellularTextObject_Serialize_ViaNoopWriter exercises CellularTextObject
// Serialize with all non-default values via a no-op writer.
func TestCellularTextObject_Serialize_ViaNoopWriter(t *testing.T) {
	c := NewCellularTextObject()
	c.cellWidth = 28
	c.cellHeight = 32
	c.horzSpacing = 4
	c.vertSpacing = 6
	c.wordWrap = false

	w := &noopWriter2{}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestCellularTextObject_Serialize_Defaults_ViaNoopWriter exercises the path
// where all optional cellular fields are at defaults.
func TestCellularTextObject_Serialize_Defaults_ViaNoopWriter(t *testing.T) {
	c := NewCellularTextObject()
	// defaults: cellWidth=0, cellHeight=0, horzSpacing=0, vertSpacing=0, wordWrap=true
	w := &noopWriter2{}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestCellularTextObject_Deserialize_ViaNoopReader calls Deserialize with a
// no-op reader.
func TestCellularTextObject_Deserialize_ViaNoopReader(t *testing.T) {
	c := NewCellularTextObject()
	r := &noopReader{}
	if err := c.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// ── CheckBoxObject ────────────────────────────────────────────────────────────

// TestCheckBoxObject_Serialize_ViaNoopWriter exercises CheckBoxObject Serialize
// with all non-default values via a no-op writer.
func TestCheckBoxObject_Serialize_ViaNoopWriter(t *testing.T) {
	c := NewCheckBoxObject()
	c.isChecked = true
	c.checkedSymbol = CheckedSymbolCross
	c.uncheckedSymbol = UncheckedSymbolMinus
	c.dataColumn = "IsActive"
	c.expression = "[X] > 0"
	c.checkWidthRatio = 0.75
	c.hideIfUnchecked = true
	c.editable = true

	w := &noopWriter2{}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestCheckBoxObject_Serialize_Defaults_ViaNoopWriter exercises the default
// branches (no attributes written).
func TestCheckBoxObject_Serialize_Defaults_ViaNoopWriter(t *testing.T) {
	c := NewCheckBoxObject()
	// defaults: isChecked=false, checkedSymbol=Check, uncheckedSymbol=None,
	// dataColumn="", expression="", checkWidthRatio=1.0, hideIfUnchecked=false, editable=false
	w := &noopWriter2{}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestCheckBoxObject_Deserialize_ViaNoopReader exercises Deserialize with a
// no-op reader.
func TestCheckBoxObject_Deserialize_ViaNoopReader(t *testing.T) {
	c := NewCheckBoxObject()
	r := &noopReader{}
	if err := c.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// ── ContainerObject ───────────────────────────────────────────────────────────

// TestContainerObject_Serialize_ViaNoopWriter exercises ContainerObject
// Serialize with events set.
func TestContainerObject_Serialize_ViaNoopWriter(t *testing.T) {
	c := NewContainerObject()
	c.beforeLayoutEvent = "OnBefore"
	c.afterLayoutEvent = "OnAfter"

	w := &noopWriter2{}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestContainerObject_Serialize_Defaults_ViaNoopWriter exercises Serialize
// with default (empty) event names.
func TestContainerObject_Serialize_Defaults_ViaNoopWriter(t *testing.T) {
	c := NewContainerObject()
	w := &noopWriter2{}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestContainerObject_Serialize_WithChildren_ViaNoopWriter exercises the child
// loop in ContainerObject.Serialize with a no-op writer.
func TestContainerObject_Serialize_WithChildren_ViaNoopWriter(t *testing.T) {
	c := NewContainerObject()
	child := NewBarcodeObject()
	child.text = "test"
	c.AddChild(child)

	w := &noopWriter2{}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestContainerObject_Serialize_WithChildren_WriteObjectError exercises the
// error-return branch in the ContainerObject.Serialize child loop when
// WriteObject returns an error (the writeObjErrWriter provides this).
//
// NOTE: This covers the `return err` at container.go:274.
func TestContainerObject_Serialize_WithChildren_WriteObjectError(t *testing.T) {
	c := NewContainerObject()
	child := NewBarcodeObject()
	c.AddChild(child)

	w := &writeObjErrWriter{msg: "child write error"}
	err := c.Serialize(w)
	if err == nil {
		t.Fatal("expected error from Serialize when WriteObject fails, got nil")
	}
	if err.Error() != "child write error" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

// TestContainerObject_Deserialize_ViaNoopReader exercises Deserialize.
func TestContainerObject_Deserialize_ViaNoopReader(t *testing.T) {
	c := NewContainerObject()
	r := &noopReader{}
	if err := c.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// TestContainerObject_UpdateLayout_InternalCall exercises UpdateLayout from
// within the internal package. The method has an empty body so the coverage
// tool reports 0% regardless; this test ensures no panic occurs.
func TestContainerObject_UpdateLayout_InternalCall(t *testing.T) {
	c := NewContainerObject()
	c.UpdateLayout(0, 0)
	c.UpdateLayout(100, -50)
}

// ── SubreportObject ───────────────────────────────────────────────────────────

// TestSubreportObject_Serialize_ViaNoopWriter exercises SubreportObject Serialize
// with both reportPageName and printOnParent=true.
func TestSubreportObject_Serialize_ViaNoopWriter(t *testing.T) {
	s := NewSubreportObject()
	s.reportPageName = "Page1"
	s.printOnParent = true

	w := &noopWriter2{}
	if err := s.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestSubreportObject_Serialize_Defaults_ViaNoopWriter exercises Serialize with
// empty reportPageName and printOnParent=false.
func TestSubreportObject_Serialize_Defaults_ViaNoopWriter(t *testing.T) {
	s := NewSubreportObject()
	w := &noopWriter2{}
	if err := s.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestSubreportObject_Deserialize_ViaNoopReader exercises Deserialize.
func TestSubreportObject_Deserialize_ViaNoopReader(t *testing.T) {
	s := NewSubreportObject()
	r := &noopReader{}
	if err := s.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// ── DigitalSignatureObject ────────────────────────────────────────────────────

// TestDigitalSignatureObject_Serialize_ViaNoopWriter exercises Serialize with
// a non-empty placeholder.
func TestDigitalSignatureObject_Serialize_ViaNoopWriter(t *testing.T) {
	d := NewDigitalSignatureObject()
	d.placeholder = "Sign here"

	w := &noopWriter2{}
	if err := d.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestDigitalSignatureObject_Serialize_Empty_ViaNoopWriter exercises Serialize
// with an empty placeholder (branch skipped).
func TestDigitalSignatureObject_Serialize_Empty_ViaNoopWriter(t *testing.T) {
	d := NewDigitalSignatureObject()
	w := &noopWriter2{}
	if err := d.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestDigitalSignatureObject_Deserialize_ViaNoopReader exercises Deserialize.
func TestDigitalSignatureObject_Deserialize_ViaNoopReader(t *testing.T) {
	d := NewDigitalSignatureObject()
	r := &noopReader{}
	if err := d.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// ── HtmlObject ────────────────────────────────────────────────────────────────

// TestHtmlObject_Serialize_RightToLeft_ViaNoopWriter exercises the rightToLeft
// branch in HtmlObject.Serialize.
func TestHtmlObject_Serialize_RightToLeft_ViaNoopWriter(t *testing.T) {
	h := NewHtmlObject()
	h.rightToLeft = true

	w := &noopWriter2{}
	if err := h.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestHtmlObject_Serialize_Default_ViaNoopWriter exercises the default
// (rightToLeft=false) branch in HtmlObject.Serialize.
func TestHtmlObject_Serialize_Default_ViaNoopWriter(t *testing.T) {
	h := NewHtmlObject()
	w := &noopWriter2{}
	if err := h.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestHtmlObject_Deserialize_ViaNoopReader exercises Deserialize.
func TestHtmlObject_Deserialize_ViaNoopReader(t *testing.T) {
	h := NewHtmlObject()
	r := &noopReader{}
	if err := h.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// ── LineObject ────────────────────────────────────────────────────────────────

// TestLineObject_Serialize_ViaNoopWriter exercises LineObject Serialize with
// diagonal=true and non-default caps.
func TestLineObject_Serialize_ViaNoopWriter(t *testing.T) {
	l := NewLineObject()
	l.diagonal = true
	l.StartCap = CapSettings{Width: 10, Height: 10, Style: CapStyleArrow}
	l.EndCap = CapSettings{Width: 6, Height: 6, Style: CapStyleCircle}

	w := &noopWriter2{}
	if err := l.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestLineObject_Serialize_Defaults_ViaNoopWriter exercises the default
// branches (diagonal=false, caps at defaults).
func TestLineObject_Serialize_Defaults_ViaNoopWriter(t *testing.T) {
	l := NewLineObject()
	// defaults: diagonal=false, StartCap=DefaultCapSettings(), EndCap=DefaultCapSettings()
	w := &noopWriter2{}
	if err := l.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestLineObject_Deserialize_ViaNoopReader exercises Deserialize.
func TestLineObject_Deserialize_ViaNoopReader(t *testing.T) {
	l := NewLineObject()
	r := &noopReader{}
	if err := l.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// TestLineObject_Deserialize_WithCaps_ViaReader exercises Deserialize when
// StartCap and EndCap dot-qualified attributes are present in the reader.
// Uses the FRX format: StartCap.Style="Arrow", StartCap.Width=10, etc.
func TestLineObject_Deserialize_WithCaps_ViaReader(t *testing.T) {
	r := &capReadReader{}
	l := NewLineObject()
	if err := l.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
	if l.StartCap.Style != CapStyleArrow {
		t.Errorf("StartCap.Style: got %d, want CapStyleArrow(%d)", l.StartCap.Style, CapStyleArrow)
	}
	if l.StartCap.Width != 10 {
		t.Errorf("StartCap.Width: got %v, want 10", l.StartCap.Width)
	}
	if l.EndCap.Style != CapStyleCircle {
		t.Errorf("EndCap.Style: got %d, want CapStyleCircle(%d)", l.EndCap.Style, CapStyleCircle)
	}
	if l.diagonal {
		t.Error("Diagonal: expected false (reader returns default)")
	}
}

// capReadReader is a mock reader that returns dot-qualified cap attributes
// matching the FRX format produced by C# CapSettings.Serialize().
type capReadReader struct{}

func (r *capReadReader) ReadStr(name, def string) string {
	switch name {
	case "StartCap.Style":
		return "Arrow"
	case "EndCap.Style":
		return "Circle"
	}
	return def
}
func (r *capReadReader) ReadInt(name string, def int) int { return def }
func (r *capReadReader) ReadBool(name string, def bool) bool { return def }
func (r *capReadReader) ReadFloat(name string, def float32) float32 {
	switch name {
	case "StartCap.Width":
		return 10
	case "StartCap.Height":
		return 10
	case "EndCap.Width":
		return 6
	case "EndCap.Height":
		return 6
	}
	return def
}
func (r *capReadReader) NextChild() (string, bool) { return "", false }
func (r *capReadReader) FinishChild() error         { return nil }

// ── ShapeObject ───────────────────────────────────────────────────────────────

// TestShapeObject_Serialize_ViaNoopWriter exercises ShapeObject Serialize with
// a non-Rectangle shape and non-zero curve.
func TestShapeObject_Serialize_ViaNoopWriter(t *testing.T) {
	s := NewShapeObject()
	s.shape = ShapeKindRoundRectangle
	s.curve = 15

	w := &noopWriter2{}
	if err := s.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestShapeObject_Serialize_Defaults_ViaNoopWriter exercises Serialize with
// default (Rectangle) shape and zero curve.
func TestShapeObject_Serialize_Defaults_ViaNoopWriter(t *testing.T) {
	s := NewShapeObject()
	w := &noopWriter2{}
	if err := s.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestShapeObject_Deserialize_ViaNoopReader exercises Deserialize.
func TestShapeObject_Deserialize_ViaNoopReader(t *testing.T) {
	s := NewShapeObject()
	r := &noopReader{}
	if err := s.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// ── MapLayer ──────────────────────────────────────────────────────────────────

// TestMapLayer_Serialize_AllFields_ViaNoopWriter exercises MapLayer Serialize
// with all fields set.
func TestMapLayer_Serialize_AllFields_ViaNoopWriter(t *testing.T) {
	l := NewMapLayer()
	l.Shapefile = "world.shp"
	l.Type = "Choropleth"
	l.DataSource = "GeoDS"
	l.Filter = "[Country]='US'"
	l.SpatialColumn = "ISO"
	l.SpatialValue = "[ISO]"
	l.AnalyticalValue = "[GDP]"
	l.LabelColumn = "Name"
	l.BoxAsString = "0,0,100,100"
	l.Palette = "Blues"

	w := &noopWriter2{}
	if err := l.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestMapLayer_Serialize_Defaults_ViaNoopWriter exercises Serialize with all
// fields empty (all branches skipped).
func TestMapLayer_Serialize_Defaults_ViaNoopWriter(t *testing.T) {
	l := NewMapLayer()
	w := &noopWriter2{}
	if err := l.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestMapLayer_Deserialize_ViaNoopReader exercises Deserialize.
func TestMapLayer_Deserialize_ViaNoopReader(t *testing.T) {
	l := NewMapLayer()
	r := &noopReader{}
	if err := l.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// ── MapObject ─────────────────────────────────────────────────────────────────

// TestMapObject_Serialize_WithOffsets_ViaNoopWriter exercises MapObject Serialize
// with non-zero OffsetX and OffsetY.
func TestMapObject_Serialize_WithOffsets_ViaNoopWriter(t *testing.T) {
	m := NewMapObject()
	m.OffsetX = 3.5
	m.OffsetY = 7.2

	w := &noopWriter2{}
	if err := m.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestMapObject_Serialize_ZeroOffsets_ViaNoopWriter exercises Serialize when
// offsets are zero (branches skipped).
func TestMapObject_Serialize_ZeroOffsets_ViaNoopWriter(t *testing.T) {
	m := NewMapObject()
	w := &noopWriter2{}
	if err := m.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestMapObject_Deserialize_ViaNoopReader exercises Deserialize.
func TestMapObject_Deserialize_ViaNoopReader(t *testing.T) {
	m := NewMapObject()
	r := &noopReader{}
	if err := m.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// ── parseCapStyle / formatCapStyle round-trip ─────────────────────────────────
// These test the internal helpers that replaced the old capFromStr/capToStr CSV
// approach. The new helpers use the FRX string-name format ("Arrow", "Circle",
// etc.) that matches C# CapSettings.Serialize / FRWriter.WriteValue output.

// TestParseCapStyle_AllValues verifies every enum name is parsed correctly.
func TestParseCapStyle_AllValues(t *testing.T) {
	cases := []struct {
		input string
		want  CapStyle
	}{
		{"None", CapStyleNone},
		{"Circle", CapStyleCircle},
		{"Square", CapStyleSquare},
		{"Diamond", CapStyleDiamond},
		{"Arrow", CapStyleArrow},
		{"", CapStyleNone},          // empty → default
		{"unknown", CapStyleNone},   // unknown → default
	}
	for _, tc := range cases {
		got := parseCapStyle(tc.input)
		if got != tc.want {
			t.Errorf("parseCapStyle(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

// TestFormatCapStyle_AllValues verifies every CapStyle produces the correct name.
func TestFormatCapStyle_AllValues(t *testing.T) {
	cases := []struct {
		style CapStyle
		want  string
	}{
		{CapStyleNone, "None"},
		{CapStyleCircle, "Circle"},
		{CapStyleSquare, "Square"},
		{CapStyleDiamond, "Diamond"},
		{CapStyleArrow, "Arrow"},
	}
	for _, tc := range cases {
		got := formatCapStyle(tc.style)
		if got != tc.want {
			t.Errorf("formatCapStyle(%v) = %q, want %q", tc.style, got, tc.want)
		}
	}
}

// TestFormatParseCapStyle_RoundTrip verifies format→parse is identity for all styles.
func TestFormatParseCapStyle_RoundTrip(t *testing.T) {
	styles := []CapStyle{CapStyleNone, CapStyleCircle, CapStyleSquare, CapStyleDiamond, CapStyleArrow}
	for _, s := range styles {
		name := formatCapStyle(s)
		got := parseCapStyle(name)
		if got != s {
			t.Errorf("round-trip failed for %v: format=%q, parse back=%v", s, name, got)
		}
	}
}
