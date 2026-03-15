package crossview_test

// crossview_coverage_test.go — additional tests to push coverage to 85%+
//
// Covers:
//   - serial.go: HeaderDescriptor, CellDescriptor, CrossViewHeader, CrossViewCells,
//     CrossViewDataSerial Serialize/Deserialize round-trips, ParseIndexArray,
//     FormatIndexArray
//   - slice.go: SetMeasuresInXAxis, SetMeasuresLevel, AddRow, out-of-range
//     getters, DataColumnCount/DataRowCount with measuresInX/measuresInY,
//     TraverseXAxis/TraverseYAxis multi-level + empty, GetMeasureCell edge cases,
//     aggregateAdd (int, int64, float64, float32, string)
//   - crossview.go: buildDescriptors with nil source, CreateDescriptors measuresInY

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/crossview"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── serialization helpers ─────────────────────────────────────────────────────

// serializeHeaderDescriptor serializes a HeaderDescriptor and returns the XML.
func serializeHeaderDescriptor(t *testing.T, h *crossview.HeaderDescriptor) string {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Header"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := h.Serialize(w); err != nil {
		t.Fatalf("HeaderDescriptor.Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.String()
}

func deserializeHeaderDescriptor(t *testing.T, xml string) *crossview.HeaderDescriptor {
	t.Helper()
	r := serial.NewReader(strings.NewReader(xml))
	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed; xml=%q", xml)
	}
	if typeName != "Header" {
		t.Fatalf("got typeName=%q, want Header", typeName)
	}
	h := &crossview.HeaderDescriptor{}
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("HeaderDescriptor.Deserialize: %v", err)
	}
	return h
}

func serializeCellDescriptor(t *testing.T, c *crossview.CellDescriptor) string {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Cell"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := c.Serialize(w); err != nil {
		t.Fatalf("CellDescriptor.Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.String()
}

func deserializeCellDescriptor(t *testing.T, xmlStr string) *crossview.CellDescriptor {
	t.Helper()
	r := serial.NewReader(strings.NewReader(xmlStr))
	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed")
	}
	if typeName != "Cell" {
		t.Fatalf("got typeName=%q, want Cell", typeName)
	}
	c := &crossview.CellDescriptor{}
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("CellDescriptor.Deserialize: %v", err)
	}
	return c
}

// ── HeaderDescriptor round-trip ───────────────────────────────────────────────

func TestHeaderDescriptor_RoundTrip_Full(t *testing.T) {
	orig := &crossview.HeaderDescriptor{
		Descriptor:   crossview.Descriptor{Expression: "expr1"},
		FieldName:    "Category",
		MeasureName:  "Sales",
		IsGrandTotal: true,
		IsTotal:      true,
		IsMeasure:    true,
		Level:        2,
		Cell:         5,
		LevelSize:    3,
		CellSize:     4,
	}

	xml := serializeHeaderDescriptor(t, orig)

	// Verify key attributes appear.
	for _, want := range []string{"FieldName", "Category", "MeasureName", "Sales",
		"IsGrandTotal", "IsTotal", "IsMeasure", "Expression", "expr1"} {
		if !strings.Contains(xml, want) {
			t.Errorf("serialized XML missing %q:\n%s", want, xml)
		}
	}

	got := deserializeHeaderDescriptor(t, xml)

	if got.FieldName != orig.FieldName {
		t.Errorf("FieldName: got %q, want %q", got.FieldName, orig.FieldName)
	}
	if got.MeasureName != orig.MeasureName {
		t.Errorf("MeasureName: got %q, want %q", got.MeasureName, orig.MeasureName)
	}
	if got.IsGrandTotal != orig.IsGrandTotal {
		t.Errorf("IsGrandTotal: got %v, want %v", got.IsGrandTotal, orig.IsGrandTotal)
	}
	if got.IsTotal != orig.IsTotal {
		t.Errorf("IsTotal: got %v, want %v", got.IsTotal, orig.IsTotal)
	}
	if got.IsMeasure != orig.IsMeasure {
		t.Errorf("IsMeasure: got %v, want %v", got.IsMeasure, orig.IsMeasure)
	}
	if got.Level != orig.Level {
		t.Errorf("Level: got %d, want %d", got.Level, orig.Level)
	}
	if got.Cell != orig.Cell {
		t.Errorf("Cell: got %d, want %d", got.Cell, orig.Cell)
	}
	if got.LevelSize != orig.LevelSize {
		t.Errorf("LevelSize: got %d, want %d", got.LevelSize, orig.LevelSize)
	}
	if got.CellSize != orig.CellSize {
		t.Errorf("CellSize: got %d, want %d", got.CellSize, orig.CellSize)
	}
	if got.Expression != orig.Expression {
		t.Errorf("Expression: got %q, want %q", got.Expression, orig.Expression)
	}
}

func TestHeaderDescriptor_RoundTrip_ZeroValues(t *testing.T) {
	orig := &crossview.HeaderDescriptor{}
	xml := serializeHeaderDescriptor(t, orig)
	got := deserializeHeaderDescriptor(t, xml)
	if got.FieldName != "" || got.MeasureName != "" || got.IsTotal || got.IsGrandTotal || got.IsMeasure {
		t.Errorf("zero-value round-trip failed: got %+v", got)
	}
}

// ── CellDescriptor round-trip ─────────────────────────────────────────────────

func TestCellDescriptor_RoundTrip_Full(t *testing.T) {
	orig := &crossview.CellDescriptor{
		Descriptor:    crossview.Descriptor{Expression: "expr2"},
		XFieldName:    "Region",
		YFieldName:    "Product",
		MeasureName:   "Revenue",
		IsXTotal:      true,
		IsYTotal:      true,
		IsXGrandTotal: false,
		IsYGrandTotal: false,
		X:             3,
		Y:             7,
	}

	xmlStr := serializeCellDescriptor(t, orig)
	got := deserializeCellDescriptor(t, xmlStr)

	if got.XFieldName != orig.XFieldName {
		t.Errorf("XFieldName: got %q, want %q", got.XFieldName, orig.XFieldName)
	}
	if got.YFieldName != orig.YFieldName {
		t.Errorf("YFieldName: got %q, want %q", got.YFieldName, orig.YFieldName)
	}
	if got.MeasureName != orig.MeasureName {
		t.Errorf("MeasureName: got %q, want %q", got.MeasureName, orig.MeasureName)
	}
	if got.IsXTotal != orig.IsXTotal {
		t.Errorf("IsXTotal: got %v, want %v", got.IsXTotal, orig.IsXTotal)
	}
	if got.IsYTotal != orig.IsYTotal {
		t.Errorf("IsYTotal: got %v, want %v", got.IsYTotal, orig.IsYTotal)
	}
	if got.X != orig.X {
		t.Errorf("X: got %d, want %d", got.X, orig.X)
	}
	if got.Y != orig.Y {
		t.Errorf("Y: got %d, want %d", got.Y, orig.Y)
	}
	if got.Expression != orig.Expression {
		t.Errorf("Expression: got %q, want %q", got.Expression, orig.Expression)
	}
}

func TestCellDescriptor_RoundTrip_XGrandTotal(t *testing.T) {
	// When IsXGrandTotal is true, XFieldName should be suppressed.
	orig := &crossview.CellDescriptor{
		XFieldName:    "ShouldBeOmitted",
		YFieldName:    "Product",
		IsXGrandTotal: true,
		IsYGrandTotal: true,
	}
	xmlStr := serializeCellDescriptor(t, orig)

	// XFieldName should not appear because IsXGrandTotal clears it.
	if strings.Contains(xmlStr, "ShouldBeOmitted") {
		t.Errorf("XFieldName should be suppressed when IsXGrandTotal=true; got: %s", xmlStr)
	}

	got := deserializeCellDescriptor(t, xmlStr)
	if got.IsXGrandTotal != true {
		t.Errorf("IsXGrandTotal: got %v, want true", got.IsXGrandTotal)
	}
	if got.IsYGrandTotal != true {
		t.Errorf("IsYGrandTotal: got %v, want true", got.IsYGrandTotal)
	}
	// XFieldName will be empty since it was suppressed.
	if got.XFieldName != "" {
		t.Errorf("XFieldName: got %q, want empty (was suppressed)", got.XFieldName)
	}
}

// ── CrossViewHeader ───────────────────────────────────────────────────────────

func TestCrossViewHeader_CRUD(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	if h == nil {
		t.Fatal("NewCrossViewHeader returned nil")
	}
	if h.Count() != 0 {
		t.Errorf("Count: got %d, want 0", h.Count())
	}

	d1 := &crossview.HeaderDescriptor{FieldName: "A"}
	d2 := &crossview.HeaderDescriptor{FieldName: "B"}
	h.Add(d1)
	h.Add(d2)

	if h.Count() != 2 {
		t.Errorf("Count after Add: got %d, want 2", h.Count())
	}
	if h.Get(0) != d1 {
		t.Error("Get(0) should return d1")
	}
	if h.Get(1) != d2 {
		t.Error("Get(1) should return d2")
	}
	if h.Get(-1) != nil {
		t.Error("Get(-1) should return nil")
	}
	if h.Get(99) != nil {
		t.Error("Get(99) should return nil")
	}

	h.Clear()
	if h.Count() != 0 {
		t.Errorf("Count after Clear: got %d, want 0", h.Count())
	}
}

func TestCrossViewHeader_RoundTrip(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	h.Add(&crossview.HeaderDescriptor{FieldName: "Region", Level: 0, Cell: 1, CellSize: 2})
	h.Add(&crossview.HeaderDescriptor{FieldName: "Category", Level: 1, IsMeasure: true, MeasureName: "Sales"})

	// Serialize.
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Columns"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := h.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	// Deserialize.
	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "Columns" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	h2 := crossview.NewCrossViewHeader("Columns")
	if err := h2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if h2.Count() != 2 {
		t.Fatalf("Count: got %d, want 2", h2.Count())
	}
	if h2.Get(0).FieldName != "Region" {
		t.Errorf("Item[0].FieldName: got %q, want Region", h2.Get(0).FieldName)
	}
	if h2.Get(1).MeasureName != "Sales" {
		t.Errorf("Item[1].MeasureName: got %q, want Sales", h2.Get(1).MeasureName)
	}
	if !h2.Get(1).IsMeasure {
		t.Error("Item[1].IsMeasure should be true")
	}
}

func TestCrossViewHeader_Empty_RoundTrip(t *testing.T) {
	h := crossview.NewCrossViewHeader("Rows")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Rows"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := h.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "Rows" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	h2 := crossview.NewCrossViewHeader("Rows")
	if err := h2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if h2.Count() != 0 {
		t.Errorf("Count: got %d, want 0", h2.Count())
	}
}

// ── CrossViewCells ────────────────────────────────────────────────────────────

func TestCrossViewCells_CRUD(t *testing.T) {
	c := crossview.NewCrossViewCells("Cells")
	if c == nil {
		t.Fatal("NewCrossViewCells returned nil")
	}
	if c.Count() != 0 {
		t.Errorf("Count: got %d, want 0", c.Count())
	}

	d1 := &crossview.CellDescriptor{X: 0, Y: 0}
	d2 := &crossview.CellDescriptor{X: 1, Y: 1}
	c.Add(d1)
	c.Add(d2)

	if c.Count() != 2 {
		t.Errorf("Count after Add: got %d, want 2", c.Count())
	}
	if c.Get(0) != d1 {
		t.Error("Get(0) should return d1")
	}
	if c.Get(1) != d2 {
		t.Error("Get(1) should return d2")
	}
	if c.Get(-1) != nil {
		t.Error("Get(-1) should return nil")
	}
	if c.Get(99) != nil {
		t.Error("Get(99) should return nil")
	}

	c.Clear()
	if c.Count() != 0 {
		t.Errorf("Count after Clear: got %d, want 0", c.Count())
	}
}

func TestCrossViewCells_RoundTrip(t *testing.T) {
	cells := crossview.NewCrossViewCells("Cells")
	cells.Add(&crossview.CellDescriptor{X: 0, Y: 0, MeasureName: "Sales", IsXTotal: true})
	cells.Add(&crossview.CellDescriptor{X: 1, Y: 2, XFieldName: "Region", YFieldName: "Product"})

	// Serialize.
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Cells"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := cells.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	// Deserialize.
	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "Cells" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	cells2 := crossview.NewCrossViewCells("Cells")
	if err := cells2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if cells2.Count() != 2 {
		t.Fatalf("Count: got %d, want 2", cells2.Count())
	}
	if cells2.Get(0).MeasureName != "Sales" {
		t.Errorf("Item[0].MeasureName: got %q, want Sales", cells2.Get(0).MeasureName)
	}
	if !cells2.Get(0).IsXTotal {
		t.Error("Item[0].IsXTotal should be true")
	}
	if cells2.Get(1).XFieldName != "Region" {
		t.Errorf("Item[1].XFieldName: got %q, want Region", cells2.Get(1).XFieldName)
	}
	if cells2.Get(1).Y != 2 {
		t.Errorf("Item[1].Y: got %d, want 2", cells2.Get(1).Y)
	}
}

func TestCrossViewCells_Empty_RoundTrip(t *testing.T) {
	cells := crossview.NewCrossViewCells("Cells")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("Cells"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := cells.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "Cells" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}
	cells2 := crossview.NewCrossViewCells("Cells")
	if err := cells2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if cells2.Count() != 0 {
		t.Errorf("Count: got %d, want 0", cells2.Count())
	}
}

// ── CrossViewDataSerial ───────────────────────────────────────────────────────

func buildCrossViewData() *crossview.CrossViewData {
	d := &crossview.CrossViewData{}
	d.AddColumn(&crossview.HeaderDescriptor{FieldName: "Region", Level: 0})
	d.AddRow(&crossview.HeaderDescriptor{FieldName: "Product", Level: 0})
	d.AddCell(&crossview.CellDescriptor{X: 0, Y: 0, MeasureName: "Sales"})
	d.AddCell(&crossview.CellDescriptor{X: 1, Y: 0, MeasureName: "Sales"})
	return d
}

func TestCrossViewDataSerial_RoundTrip(t *testing.T) {
	d := buildCrossViewData()
	s := crossview.NewCrossViewDataSerial(d)
	s.ColumnDescriptorsIndexes = "0,1"
	s.RowDescriptorsIndexes = "0"
	s.ColumnTerminalIndexes = "0,1"
	s.RowTerminalIndexes = "0"

	// Serialize.
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("CrossViewData"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := s.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	xmlStr := buf.String()

	// Check key attributes are present.
	for _, want := range []string{"ColumnDescriptorsIndexes", "RowDescriptorsIndexes",
		"Columns", "Rows", "Cells", "Region", "Product", "Sales"} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("XML missing %q:\n%s", want, xmlStr)
		}
	}

	// Deserialize into a new object.
	d2 := &crossview.CrossViewData{}
	s2 := crossview.NewCrossViewDataSerial(d2)

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "CrossViewData" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}
	if err := s2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if s2.ColumnDescriptorsIndexes != "0,1" {
		t.Errorf("ColumnDescriptorsIndexes: got %q, want 0,1", s2.ColumnDescriptorsIndexes)
	}
	if s2.RowDescriptorsIndexes != "0" {
		t.Errorf("RowDescriptorsIndexes: got %q, want 0", s2.RowDescriptorsIndexes)
	}
	if len(d2.Columns) != 1 {
		t.Errorf("Columns len: got %d, want 1", len(d2.Columns))
	}
	if len(d2.Rows) != 1 {
		t.Errorf("Rows len: got %d, want 1", len(d2.Rows))
	}
	if len(d2.Cells) != 2 {
		t.Errorf("Cells len: got %d, want 2", len(d2.Cells))
	}
	if d2.Columns[0].FieldName != "Region" {
		t.Errorf("Columns[0].FieldName: got %q, want Region", d2.Columns[0].FieldName)
	}
	if d2.Rows[0].FieldName != "Product" {
		t.Errorf("Rows[0].FieldName: got %q, want Product", d2.Rows[0].FieldName)
	}
}

func TestCrossViewDataSerial_Empty_RoundTrip(t *testing.T) {
	d := &crossview.CrossViewData{}
	s := crossview.NewCrossViewDataSerial(d)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("CrossViewData"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := s.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	d2 := &crossview.CrossViewData{}
	s2 := crossview.NewCrossViewDataSerial(d2)

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "CrossViewData" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}
	if err := s2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if len(d2.Columns) != 0 {
		t.Errorf("Columns: got %d, want 0", len(d2.Columns))
	}
}

// ── ParseIndexArray / FormatIndexArray ────────────────────────────────────────

func TestParseIndexArray(t *testing.T) {
	tests := []struct {
		input string
		want  []int
	}{
		{"", nil},
		{"0", []int{0}},
		{"1,2,3", []int{1, 2, 3}},
		{"0,10,20,30", []int{0, 10, 20, 30}},
		{" 1 , 2 , 3 ", []int{1, 2, 3}}, // trimmed
		{"abc,1,xyz,2", []int{1, 2}},     // non-numeric entries skipped
		{"0,0,0", []int{0, 0, 0}},
	}
	for _, tt := range tests {
		got := crossview.ParseIndexArray(tt.input)
		if !intSliceEqual(got, tt.want) {
			t.Errorf("ParseIndexArray(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestFormatIndexArray(t *testing.T) {
	tests := []struct {
		input []int
		want  string
	}{
		{nil, ""},
		{[]int{}, ""},
		{[]int{0}, "0"},
		{[]int{1, 2, 3}, "1,2,3"},
		{[]int{0, 10, 20, 30}, "0,10,20,30"},
	}
	for _, tt := range tests {
		got := crossview.FormatIndexArray(tt.input)
		if got != tt.want {
			t.Errorf("FormatIndexArray(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseFormatRoundTrip(t *testing.T) {
	orig := []int{0, 1, 5, 10, 100}
	formatted := crossview.FormatIndexArray(orig)
	parsed := crossview.ParseIndexArray(formatted)
	if !intSliceEqual(parsed, orig) {
		t.Errorf("ParseFormatRoundTrip: got %v, want %v", parsed, orig)
	}
}

// intSliceEqual compares two int slices.
func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ── SliceCubeSource – SetMeasuresInXAxis, SetMeasuresLevel, AddRow ────────────

func TestSliceCubeSource_SetMeasuresInXAxis(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	// default is true
	if !src.MeasuresInXAxis() {
		t.Error("default MeasuresInXAxis should be true")
	}
	src.SetMeasuresInXAxis(false)
	if src.MeasuresInXAxis() {
		t.Error("MeasuresInXAxis should be false after SetMeasuresInXAxis(false)")
	}
	if !src.MeasuresInYAxis() {
		t.Error("MeasuresInYAxis should be true when MeasuresInXAxis=false")
	}
}

func TestSliceCubeSource_SetMeasuresLevel(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	// default is -1
	if src.MeasuresLevel() != -1 {
		t.Errorf("default MeasuresLevel: got %d, want -1", src.MeasuresLevel())
	}
	src.SetMeasuresLevel(0)
	if src.MeasuresLevel() != 0 {
		t.Errorf("MeasuresLevel after Set: got %d, want 0", src.MeasuresLevel())
	}
	src.SetMeasuresLevel(3)
	if src.MeasuresLevel() != 3 {
		t.Errorf("MeasuresLevel after Set: got %d, want 3", src.MeasuresLevel())
	}
}

func TestSliceCubeSource_AddRow(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")

	// Use AddRow (not AddRows).
	src.AddRow(map[string]any{"Cat": "A", "Region": "North", "Sales": 100})
	src.AddRow(map[string]any{"Cat": "B", "Region": "South", "Sales": 200})
	src.Build()

	if src.DataColumnCount() != 2 {
		t.Errorf("DataColumnCount: got %d, want 2", src.DataColumnCount())
	}
	if src.DataRowCount() != 2 {
		t.Errorf("DataRowCount: got %d, want 2", src.DataRowCount())
	}
}

// ── Out-of-range field name getters ──────────────────────────────────────────

func TestSliceCubeSource_Getters_OutOfRange(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("X")
	src.AddYAxisField("Y")
	src.AddMeasure("M")

	// Valid indexes.
	if src.GetXAxisFieldName(0) != "X" {
		t.Errorf("GetXAxisFieldName(0): got %q, want X", src.GetXAxisFieldName(0))
	}
	if src.GetYAxisFieldName(0) != "Y" {
		t.Errorf("GetYAxisFieldName(0): got %q, want Y", src.GetYAxisFieldName(0))
	}
	if src.GetMeasureName(0) != "M" {
		t.Errorf("GetMeasureName(0): got %q, want M", src.GetMeasureName(0))
	}

	// Out-of-range: negative.
	if src.GetXAxisFieldName(-1) != "" {
		t.Errorf("GetXAxisFieldName(-1): got %q, want empty", src.GetXAxisFieldName(-1))
	}
	if src.GetYAxisFieldName(-1) != "" {
		t.Errorf("GetYAxisFieldName(-1): got %q, want empty", src.GetYAxisFieldName(-1))
	}
	if src.GetMeasureName(-1) != "" {
		t.Errorf("GetMeasureName(-1): got %q, want empty", src.GetMeasureName(-1))
	}

	// Out-of-range: too large.
	if src.GetXAxisFieldName(99) != "" {
		t.Errorf("GetXAxisFieldName(99): got %q, want empty", src.GetXAxisFieldName(99))
	}
	if src.GetYAxisFieldName(99) != "" {
		t.Errorf("GetYAxisFieldName(99): got %q, want empty", src.GetYAxisFieldName(99))
	}
	if src.GetMeasureName(99) != "" {
		t.Errorf("GetMeasureName(99): got %q, want empty", src.GetMeasureName(99))
	}
}

// ── DataColumnCount / DataRowCount with multiple measures ────────────────────

func TestSliceCubeSource_DataColumnCount_MultiMeasuresInX(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddMeasure("Sales")
	src.AddMeasure("Qty")
	// measuresInX = true (default), 2 measures
	src.AddRow(map[string]any{"Cat": "A", "Sales": 100, "Qty": 5})
	src.AddRow(map[string]any{"Cat": "B", "Sales": 200, "Qty": 10})
	src.Build()

	// 2 xTuples × 2 measures = 4
	if src.DataColumnCount() != 4 {
		t.Errorf("DataColumnCount: got %d, want 4", src.DataColumnCount())
	}
	// measuresInX so DataRowCount = len(yTuples); no Y fields added so rows have
	// the empty-key tuple → yTuples has 1 entry (the empty-key row).
	// DataRowCount returns len(yTuples) = 1 when measuresInX=true.
	if src.DataRowCount() != 1 {
		t.Errorf("DataRowCount: got %d, want 1 (one yTuple for the empty Y key)", src.DataRowCount())
	}
}

func TestSliceCubeSource_DataRowCount_MultiMeasuresInY(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddMeasure("Qty")
	src.SetMeasuresInXAxis(false) // measures in Y
	src.AddRow(map[string]any{"Region": "North", "Sales": 100, "Qty": 5})
	src.AddRow(map[string]any{"Region": "South", "Sales": 200, "Qty": 10})
	src.Build()

	// 2 yTuples × 2 measures = 4
	if src.DataRowCount() != 4 {
		t.Errorf("DataRowCount: got %d, want 4", src.DataRowCount())
	}
	// measuresInY so DataColumnCount = len(xTuples); no X fields added so rows
	// have the empty-key tuple → xTuples has 1 entry.
	if src.DataColumnCount() != 1 {
		t.Errorf("DataColumnCount: got %d, want 1 (one xTuple for empty X key)", src.DataColumnCount())
	}
}

// ── TraverseXAxis / TraverseYAxis edge cases ──────────────────────────────────

func TestTraverseXAxis_Empty(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.Build() // no rows → xTuples is empty

	var called bool
	src.TraverseXAxis(func(cell crossview.AxisDrawCell) { called = true })
	if called {
		t.Error("TraverseXAxis should not call fn when xTuples is empty")
	}
}

func TestTraverseYAxis_Empty(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddYAxisField("Region")
	src.Build() // no rows → yTuples is empty

	var called bool
	src.TraverseYAxis(func(cell crossview.AxisDrawCell) { called = true })
	if called {
		t.Error("TraverseYAxis should not call fn when yTuples is empty")
	}
}

func TestTraverseXAxis_MultiLevel_Grouping(t *testing.T) {
	// Two X fields: Year, Quarter
	// Data: 2023/Q1, 2023/Q2, 2024/Q1
	// Expected X-axis: level 0 → "2023"(span=2), "2024"(span=1); level 1 → "Q1", "Q2", "Q1"
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Year")
	src.AddXAxisField("Quarter")
	src.AddYAxisField("Product")
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Year": "2023", "Quarter": "Q1", "Product": "A", "Sales": 100})
	src.AddRow(map[string]any{"Year": "2023", "Quarter": "Q2", "Product": "A", "Sales": 200})
	src.AddRow(map[string]any{"Year": "2024", "Quarter": "Q1", "Product": "A", "Sales": 300})
	src.Build()

	type cellInfo struct {
		text  string
		level int
		size  int
	}
	var cells []cellInfo
	src.TraverseXAxis(func(ac crossview.AxisDrawCell) {
		cells = append(cells, cellInfo{ac.Text, ac.Level, ac.SizeCell})
	})

	// Level 0: 2023 (span=2), 2024 (span=1)
	// Level 1: Q1 (span=1), Q2 (span=1), Q1 (span=1)
	if len(cells) != 5 {
		t.Fatalf("TraverseXAxis cells: got %d, want 5; cells=%v", len(cells), cells)
	}

	// Level 0 checks.
	if cells[0].text != "2023" || cells[0].level != 0 || cells[0].size != 2 {
		t.Errorf("cell[0]: got {%q,%d,%d}, want {2023,0,2}", cells[0].text, cells[0].level, cells[0].size)
	}
	if cells[1].text != "2024" || cells[1].level != 0 || cells[1].size != 1 {
		t.Errorf("cell[1]: got {%q,%d,%d}, want {2024,0,1}", cells[1].text, cells[1].level, cells[1].size)
	}
	// Level 1 checks.
	if cells[2].text != "Q1" || cells[2].level != 1 {
		t.Errorf("cell[2]: got {%q,%d}, want {Q1,1}", cells[2].text, cells[2].level)
	}
	if cells[3].text != "Q2" || cells[3].level != 1 {
		t.Errorf("cell[3]: got {%q,%d}, want {Q2,1}", cells[3].text, cells[3].level)
	}
	if cells[4].text != "Q1" || cells[4].level != 1 {
		t.Errorf("cell[4]: got {%q,%d}, want {Q1,1}", cells[4].text, cells[4].level)
	}
}

func TestTraverseYAxis_MultiLevel_Grouping(t *testing.T) {
	// Two Y fields: Category, SubCat
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Region")
	src.AddYAxisField("Category")
	src.AddYAxisField("SubCat")
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Region": "N", "Category": "Fruit", "SubCat": "Apple", "Sales": 10})
	src.AddRow(map[string]any{"Region": "N", "Category": "Fruit", "SubCat": "Banana", "Sales": 20})
	src.AddRow(map[string]any{"Region": "N", "Category": "Veg", "SubCat": "Carrot", "Sales": 30})
	src.Build()

	type cellInfo struct {
		text  string
		level int
		size  int
	}
	var cells []cellInfo
	src.TraverseYAxis(func(ac crossview.AxisDrawCell) {
		cells = append(cells, cellInfo{ac.Text, ac.Level, ac.SizeCell})
	})

	// Level 0: Fruit (span=2), Veg (span=1) → 2 cells
	// Level 1: Apple (span=1), Banana (span=1), Carrot (span=1) → 3 cells
	if len(cells) != 5 {
		t.Fatalf("TraverseYAxis cells: got %d, want 5; cells=%v", len(cells), cells)
	}
	if cells[0].text != "Fruit" || cells[0].size != 2 || cells[0].level != 0 {
		t.Errorf("cell[0]: got {%q,%d,%d}, want {Fruit,0,2}", cells[0].text, cells[0].level, cells[0].size)
	}
	if cells[1].text != "Veg" || cells[1].size != 1 || cells[1].level != 0 {
		t.Errorf("cell[1]: got {%q,%d,%d}, want {Veg,0,1}", cells[1].text, cells[1].level, cells[1].size)
	}
}

// ── GetMeasureCell edge cases ─────────────────────────────────────────────────

func TestGetMeasureCell_OutOfRange(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 100})
	src.Build()

	// Out-of-range x.
	mc := src.GetMeasureCell(99, 0)
	if mc.Text != "" {
		t.Errorf("GetMeasureCell(99,0): got %q, want empty", mc.Text)
	}
	// Out-of-range y.
	mc = src.GetMeasureCell(0, 99)
	if mc.Text != "" {
		t.Errorf("GetMeasureCell(0,99): got %q, want empty", mc.Text)
	}
}

func TestGetMeasureCell_MultiMeasuresInX(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddMeasure("Qty")
	// measuresInX = true (default)
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 100, "Qty": 5})
	src.Build()

	// x=0 → xTupleIdx=0/2=0, measureIdx=0%2=0 → Sales
	mc0 := src.GetMeasureCell(0, 0)
	if mc0.Text != "100" {
		t.Errorf("GetMeasureCell(0,0): got %q, want 100", mc0.Text)
	}
	// x=1 → xTupleIdx=1/2=0, measureIdx=1%2=1 → Qty
	mc1 := src.GetMeasureCell(1, 0)
	if mc1.Text != "5" {
		t.Errorf("GetMeasureCell(1,0): got %q, want 5", mc1.Text)
	}
}

func TestGetMeasureCell_MultiMeasuresInY(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddMeasure("Qty")
	src.SetMeasuresInXAxis(false) // measures in Y
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 200, "Qty": 10})
	src.Build()

	// y=0 → yTupleIdx=0/2=0, measureIdx=0%2=0 → Sales
	mc0 := src.GetMeasureCell(0, 0)
	if mc0.Text != "200" {
		t.Errorf("GetMeasureCell(0,0): got %q, want 200", mc0.Text)
	}
	// y=1 → yTupleIdx=1/2=0, measureIdx=1%2=1 → Qty
	mc1 := src.GetMeasureCell(0, 1)
	if mc1.Text != "10" {
		t.Errorf("GetMeasureCell(0,1): got %q, want 10", mc1.Text)
	}
}

func TestGetMeasureCell_MissingXKey(t *testing.T) {
	// Build a source with data, then look for a combo that should not exist.
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 100})
	src.Build()

	// Only 1 column (Cat=A), so x=1 is out of range for xTuples.
	mc := src.GetMeasureCell(1, 0)
	if mc.Text != "" {
		t.Errorf("GetMeasureCell(1,0): got %q, want empty (xTuple out of range)", mc.Text)
	}
}

// ── aggregateAdd — via multiple rows ─────────────────────────────────────────
// aggregateAdd is an internal function; we test it via the SliceCubeSource
// accumulation behavior in GetMeasureCell.

func TestAggregateAdd_Int(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 100})
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 200}) // same key → sum
	src.Build()

	mc := src.GetMeasureCell(0, 0)
	if mc.Text != "300" {
		t.Errorf("aggregateAdd int: got %q, want 300", mc.Text)
	}
}

func TestAggregateAdd_Float64(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Price")
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Price": 1.5})
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Price": 2.5})
	src.Build()

	mc := src.GetMeasureCell(0, 0)
	if mc.Text != "4" && mc.Text != "4.0" {
		t.Errorf("aggregateAdd float64: got %q, want 4 (or 4.0)", mc.Text)
	}
}

func TestAggregateAdd_Int64(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Count")
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Count": int64(10)})
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Count": int64(20)})
	src.Build()

	mc := src.GetMeasureCell(0, 0)
	if mc.Text != "30" {
		t.Errorf("aggregateAdd int64: got %q, want 30", mc.Text)
	}
}

func TestAggregateAdd_Float32(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Val")
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Val": float32(1.0)})
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Val": float32(2.0)})
	src.Build()

	mc := src.GetMeasureCell(0, 0)
	// float32 → converted to float64 internally
	if mc.Text != "3" && mc.Text != "3.0" && mc.Text != "3.000000119..." {
		// Accept any representation that starts with "3"
		if !strings.HasPrefix(mc.Text, "3") {
			t.Errorf("aggregateAdd float32: got %q, want ~3", mc.Text)
		}
	}
}

func TestAggregateAdd_String_NonNumeric(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Label")
	// Non-numeric: aggregateAdd returns first non-nil value.
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Label": "first"})
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Label": "second"})
	src.Build()

	mc := src.GetMeasureCell(0, 0)
	if mc.Text != "first" {
		t.Errorf("aggregateAdd string: got %q, want first (first non-nil wins)", mc.Text)
	}
}

func TestAggregateAdd_NilFirstValue(t *testing.T) {
	// When v1 is nil, aggregateAdd should return v2.
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	// Only one row so v1 starts as nil, then gets the first value.
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 42})
	src.Build()

	mc := src.GetMeasureCell(0, 0)
	if mc.Text != "42" {
		t.Errorf("aggregateAdd nil+int: got %q, want 42", mc.Text)
	}
}

func TestAggregateAdd_NilV2(t *testing.T) {
	// When measure field is missing (nil), existing accumulation unchanged.
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 50})
	src.AddRow(map[string]any{"Cat": "A", "Region": "N"}) // Sales absent → nil
	src.Build()

	mc := src.GetMeasureCell(0, 0)
	if mc.Text != "50" {
		t.Errorf("aggregateAdd nil v2: got %q, want 50", mc.Text)
	}
}

func TestAggregateAdd_IntPlusFloat64(t *testing.T) {
	// First value is int, second is float64 → result is float64.
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 10})      // int
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": float64(5.5)}) // float64
	src.Build()

	mc := src.GetMeasureCell(0, 0)
	// 10 + 5.5 = 15.5
	if !strings.Contains(mc.Text, "15.5") {
		t.Errorf("aggregateAdd int+float64: got %q, want 15.5", mc.Text)
	}
}

func TestAggregateAdd_Int64PlusInt(t *testing.T) {
	// First value is int64, second is int.
	// aggregateAdd switches on v2's type. v2=int(50) → case int branch.
	// In the int case: v1=int64(100) doesn't match int or float64, so returns val=50.
	// This tests the "type mismatch fallback" path in the int case of aggregateAdd.
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": int64(100)})
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 50})
	src.Build()

	mc := src.GetMeasureCell(0, 0)
	// Due to type mismatch in aggregateAdd (v1=int64, v2=int), result is just int(50).
	if mc.Text != "50" {
		t.Errorf("aggregateAdd int64+int type mismatch: got %q, want 50", mc.Text)
	}
}

func TestAggregateAdd_Float64PlusInt(t *testing.T) {
	// First value is float64, second is int.
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": float64(3.5)})
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 2})
	src.Build()

	mc := src.GetMeasureCell(0, 0)
	if !strings.Contains(mc.Text, "5.5") {
		t.Errorf("aggregateAdd float64+int: got %q, want 5.5", mc.Text)
	}
}

// ── buildDescriptors with nil source ─────────────────────────────────────────

func TestBuildDescriptors_NilSource(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	// SetSource with nil should clear descriptors without panicking.
	cv.SetSource(nil)
	if len(cv.Data.Columns) != 0 {
		t.Errorf("Columns after nil source: got %d, want 0", len(cv.Data.Columns))
	}
	if len(cv.Data.Rows) != 0 {
		t.Errorf("Rows after nil source: got %d, want 0", len(cv.Data.Rows))
	}
	if len(cv.Data.Cells) != 0 {
		t.Errorf("Cells after nil source: got %d, want 0", len(cv.Data.Cells))
	}
}

// ── CreateDescriptors — measuresInY path ─────────────────────────────────────

func TestCreateDescriptors_MeasuresInYAxis(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddMeasure("Qty")
	src.SetMeasuresInXAxis(false) // measures on Y axis

	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 10, "Qty": 1})
	src.AddRow(map[string]any{"Cat": "A", "Region": "S", "Sales": 20, "Qty": 2})
	src.Build()

	cv := crossview.NewCrossViewObject()
	cv.SetSource(src)

	// X: 1 field (Cat) → 1 column descriptor.
	if len(cv.Data.Columns) != 1 {
		t.Errorf("Columns len: got %d, want 1", len(cv.Data.Columns))
	}
	// Y: 1 field (Region) + 2 measures → 3 row descriptors.
	if len(cv.Data.Rows) != 3 {
		t.Errorf("Rows len: got %d, want 3 (Region + Sales + Qty)", len(cv.Data.Rows))
	}
	// At least 2 of the row descriptors should be measures.
	measureCount := 0
	for _, r := range cv.Data.Rows {
		if r.IsMeasure {
			measureCount++
		}
	}
	if measureCount != 2 {
		t.Errorf("IsMeasure rows: got %d, want 2", measureCount)
	}
}

// ── tupleValues nil field ─────────────────────────────────────────────────────

func TestTupleValues_NilField(t *testing.T) {
	// tupleValues is tested indirectly; a nil field value should produce "".
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Cat": nil, "Region": "N", "Sales": 10})
	src.Build()

	// Category is nil → key is empty string; should still produce 1 xTuple.
	if src.DataColumnCount() != 1 {
		t.Errorf("DataColumnCount with nil Cat: got %d, want 1", src.DataColumnCount())
	}
}

// ── samePrefixUpTo — via TraverseXAxis with identical outer level ─────────────

func TestSamePrefixUpTo_AllSameOuter(t *testing.T) {
	// 2023/Q1, 2023/Q2: same outer → single outer span of 2.
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Year")
	src.AddXAxisField("Quarter")
	src.AddYAxisField("Product")
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Year": "2023", "Quarter": "Q1", "Product": "A", "Sales": 1})
	src.AddRow(map[string]any{"Year": "2023", "Quarter": "Q2", "Product": "A", "Sales": 2})
	src.Build()

	var level0cells []crossview.AxisDrawCell
	src.TraverseXAxis(func(ac crossview.AxisDrawCell) {
		if ac.Level == 0 {
			level0cells = append(level0cells, ac)
		}
	})
	if len(level0cells) != 1 {
		t.Fatalf("level 0 cells: got %d, want 1 (2023 spans both quarters)", len(level0cells))
	}
	if level0cells[0].SizeCell != 2 {
		t.Errorf("level0[0].SizeCell: got %d, want 2", level0cells[0].SizeCell)
	}
}

// ── Integration: MeasuresInY with full build ──────────────────────────────────

func TestBuild_MeasuresInYAxis_Grid(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddMeasure("Qty")
	src.SetMeasuresInXAxis(false)
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 100, "Qty": 10})
	src.AddRow(map[string]any{"Cat": "B", "Region": "N", "Sales": 200, "Qty": 20})
	src.Build()

	cv := crossview.NewCrossViewObject()
	cv.SetSource(src)

	grid, err := cv.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Only 1 unique Region "N", so yTuples=1; DataRowCount = 1*2 = 2.
	// xHeaderRows = XAxisFieldsCount(1), measuresInX=false → 1
	// yHeaderCols = YAxisFieldsCount(1) + measuresInY(2 measures > 1) → 2
	// totalCols = 2 + DataColumnCount(2 xTuples) = 4
	// totalRows = 1 + DataRowCount(2) = 3
	if grid.ColCount != 4 {
		t.Errorf("ColCount: got %d, want 4", grid.ColCount)
	}
	if grid.RowCount != 3 {
		t.Errorf("RowCount: got %d, want 3", grid.RowCount)
	}
	_ = fmt.Sprintf("grid dims: %dx%d", grid.ColCount, grid.RowCount)
}

// ── CrossViewData — CreateDescriptors with single measure, no measure axis ───

func TestCreateDescriptors_SingleMeasure_NoMeasureAxis(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales") // single measure → no extra axis level
	src.AddRow(map[string]any{"Cat": "A", "Region": "N", "Sales": 10})
	src.AddRow(map[string]any{"Cat": "B", "Region": "S", "Sales": 20})
	src.Build()

	cv := crossview.NewCrossViewObject()
	cv.SetSource(src)

	// 1 X field → 1 column descriptor.
	if len(cv.Data.Columns) != 1 {
		t.Errorf("Columns len: got %d, want 1", len(cv.Data.Columns))
	}
	// 1 Y field → 1 row descriptor.
	if len(cv.Data.Rows) != 1 {
		t.Errorf("Rows len: got %d, want 1", len(cv.Data.Rows))
	}
	// dataCols × dataRows = 2 × 2 = 4 cells, each with MeasureName = "Sales".
	if len(cv.Data.Cells) != 4 {
		t.Errorf("Cells len: got %d, want 4", len(cv.Data.Cells))
	}
	for i, c := range cv.Data.Cells {
		if c.MeasureName != "Sales" {
			t.Errorf("Cell[%d].MeasureName: got %q, want Sales", i, c.MeasureName)
		}
	}
}
