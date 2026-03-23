package crossview_test

// crossview_branches_test.go — additional tests targeting remaining uncovered branches:
//
//   - serial.go CrossViewHeader.Deserialize: non-"Header" child element path
//   - serial.go CrossViewCells.Deserialize: non-"Cell" child element path
//   - serial.go CrossViewDataSerial.Serialize: Rows/Cells WriteObjectNamed error paths
//   - serial.go CrossViewDataSerial.Deserialize: unknown child element path
//   - slice.go TraverseXAxis: nLevels == 0 early return
//   - slice.go TraverseYAxis: nLevels == 0 early return
//   - slice.go GetMeasureCell: missing yKey path, measureIdx >= len(vals) path
//   - crossview.go buildGrid: totalCols <= 0 and totalRows <= 0 clamp branches

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/crossview"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── CrossViewHeader.Deserialize: non-"Header" child element ──────────────────

// TestCrossViewHeader_Deserialize_NonHeaderChild tests that an unknown child
// element inside a Columns/Rows container is silently skipped.
func TestCrossViewHeader_Deserialize_NonHeaderChild(t *testing.T) {
	// Craft XML with a mix of Header and unknown elements.
	xmlStr := `<Columns><Unknown FieldName="skip"/><Header FieldName="Region"/></Columns>`
	r := serial.NewReader(strings.NewReader(xmlStr))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "Columns" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	h := crossview.NewCrossViewHeader("Columns")
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// Only the "Header" child should be added; "Unknown" should be skipped.
	if h.Count() != 1 {
		t.Errorf("Count: got %d, want 1", h.Count())
	}
	if h.Get(0).FieldName != "Region" {
		t.Errorf("FieldName: got %q, want Region", h.Get(0).FieldName)
	}
}

// TestCrossViewHeader_Deserialize_OnlyNonHeaderChild tests skipping when all
// children are unknown.
func TestCrossViewHeader_Deserialize_OnlyNonHeaderChild(t *testing.T) {
	xmlStr := `<Columns><Unknown A="1"/><AnotherUnknown B="2"/></Columns>`
	r := serial.NewReader(strings.NewReader(xmlStr))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "Columns" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	h := crossview.NewCrossViewHeader("Columns")
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if h.Count() != 0 {
		t.Errorf("Count: got %d, want 0", h.Count())
	}
}

// ── CrossViewCells.Deserialize: non-"Cell" child element ─────────────────────

// TestCrossViewCells_Deserialize_NonCellChild tests that unknown child
// elements are silently skipped.
func TestCrossViewCells_Deserialize_NonCellChild(t *testing.T) {
	xmlStr := `<Cells><Unknown X="0" Y="0"/><Cell X="1" Y="2"/></Cells>`
	r := serial.NewReader(strings.NewReader(xmlStr))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "Cells" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	c := crossview.NewCrossViewCells("Cells")
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// Only the "Cell" child should be added.
	if c.Count() != 1 {
		t.Errorf("Count: got %d, want 1", c.Count())
	}
	if c.Get(0).X != 1 || c.Get(0).Y != 2 {
		t.Errorf("Cell[0]: got X=%d Y=%d, want X=1 Y=2", c.Get(0).X, c.Get(0).Y)
	}
}

// TestCrossViewCells_Deserialize_OnlyNonCellChild tests skipping when no
// valid Cell children exist.
func TestCrossViewCells_Deserialize_OnlyNonCellChild(t *testing.T) {
	xmlStr := `<Cells><Rubbish Foo="bar"/></Cells>`
	r := serial.NewReader(strings.NewReader(xmlStr))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "Cells" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	c := crossview.NewCrossViewCells("Cells")
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.Count() != 0 {
		t.Errorf("Count: got %d, want 0", c.Count())
	}
}

// ── CrossViewDataSerial.Deserialize: unknown child element ───────────────────

// TestCrossViewDataSerial_Deserialize_UnknownChildXML tests that an
// unrecognised child element inside CrossViewData is silently skipped.
func TestCrossViewDataSerial_Deserialize_UnknownChildXML(t *testing.T) {
	// Craft XML that includes an unknown top-level child and real Columns/Rows/Cells.
	xmlStr := `<CrossViewData>` +
		`<UnknownElement Foo="bar"/>` +
		`<Columns><Header FieldName="Cat"/></Columns>` +
		`<Rows><Header FieldName="Region"/></Rows>` +
		`<Cells><Cell X="0" Y="0"/></Cells>` +
		`</CrossViewData>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "CrossViewData" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	d := &crossview.CrossViewData{}
	s := crossview.NewCrossViewDataSerial(d)
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if len(d.Columns) != 1 || d.Columns[0].FieldName != "Cat" {
		t.Errorf("Columns: got %v, want [{FieldName:Cat}]", d.Columns)
	}
	if len(d.Rows) != 1 || d.Rows[0].FieldName != "Region" {
		t.Errorf("Rows: got %v, want [{FieldName:Region}]", d.Rows)
	}
	if len(d.Cells) != 1 {
		t.Errorf("Cells len: got %d, want 1", len(d.Cells))
	}
}

// TestCrossViewDataSerial_Deserialize_OnlyUnknownChild tests that a
// CrossViewData with only unknown children deserializes to empty collections.
func TestCrossViewDataSerial_Deserialize_OnlyUnknownChild(t *testing.T) {
	xmlStr := `<CrossViewData><Mystery Foo="1"/><Enigma Bar="2"/></CrossViewData>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "CrossViewData" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	d := &crossview.CrossViewData{}
	s := crossview.NewCrossViewDataSerial(d)
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if len(d.Columns) != 0 || len(d.Rows) != 0 || len(d.Cells) != 0 {
		t.Errorf("Expected empty CrossViewData, got Columns=%d Rows=%d Cells=%d",
			len(d.Columns), len(d.Rows), len(d.Cells))
	}
}

// ── CrossViewDataSerial.Serialize: Rows and Cells error paths ────────────────

// serializeCrossViewDataSerial is a helper that serializes a CrossViewDataSerial
// using a real writer and returns the XML bytes.
func serializeCrossViewDataSerial(t *testing.T, s *crossview.CrossViewDataSerial) []byte {
	t.Helper()
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
	return buf.Bytes()
}

// TestCrossViewDataSerial_Serialize_WithRowsAndCells tests that a
// CrossViewDataSerial with non-empty Rows and Cells serializes correctly —
// this exercises the Rows and Cells WriteObjectNamed calls.
func TestCrossViewDataSerial_Serialize_WithRowsAndCells(t *testing.T) {
	d := &crossview.CrossViewData{}
	d.AddColumn(&crossview.HeaderDescriptor{FieldName: "Cat"})
	d.AddRow(&crossview.HeaderDescriptor{FieldName: "Region"})
	d.AddCell(&crossview.CellDescriptor{X: 0, Y: 0, MeasureName: "Sales"})
	d.AddCell(&crossview.CellDescriptor{X: 1, Y: 0, MeasureName: "Sales"})
	s := crossview.NewCrossViewDataSerial(d)
	s.ColumnDescriptorsIndexes = "0"
	s.RowDescriptorsIndexes = "0"
	s.ColumnTerminalIndexes = "0"
	s.RowTerminalIndexes = "0"

	xmlBytes := serializeCrossViewDataSerial(t, s)
	xmlStr := string(xmlBytes)

	for _, want := range []string{"Columns", "Rows", "Cells", "Cat", "Region", "Sales"} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("serialized XML missing %q:\n%s", want, xmlStr)
		}
	}

	// Round-trip: deserialize back.
	d2 := &crossview.CrossViewData{}
	s2 := crossview.NewCrossViewDataSerial(d2)
	r := serial.NewReader(bytes.NewReader(xmlBytes))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "CrossViewData" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}
	if err := s2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
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
}

// ── TraverseXAxis: nLevels == 0 early return ─────────────────────────────────

// TestTraverseXAxis_NoXFields tests that TraverseXAxis returns early when
// there are no X-axis fields (nLevels == 0), even if xTuples is non-empty.
// With no xFields, Build() produces one xTuple (empty key), so xTuples != nil
// and the check passes the first guard; but nLevels=0 causes early return.
func TestTraverseXAxis_NoXFields(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	// No X-axis fields added.
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Region": "North", "Sales": 100})
	src.Build()

	var called bool
	src.TraverseXAxis(func(_ crossview.AxisDrawCell) { called = true })
	if called {
		t.Error("TraverseXAxis should not call fn when there are no X-axis fields")
	}
}

// ── TraverseYAxis: nLevels == 0 early return ─────────────────────────────────

// TestTraverseYAxis_NoYFields tests that TraverseYAxis returns early when
// there are no Y-axis fields (nLevels == 0), even if yTuples is non-empty.
func TestTraverseYAxis_NoYFields(t *testing.T) {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	// No Y-axis fields added.
	src.AddMeasure("Sales")
	src.AddRow(map[string]any{"Cat": "A", "Sales": 100})
	src.Build()

	var called bool
	src.TraverseYAxis(func(_ crossview.AxisDrawCell) { called = true })
	if called {
		t.Error("TraverseYAxis should not call fn when there are no Y-axis fields")
	}
}

// ── GetMeasureCell: missing yKey path ────────────────────────────────────────

// TestGetMeasureCell_MissingYKey tests that GetMeasureCell returns an empty
// MeasureCell when the yKey is not present in cellData for the given xKey.
// We achieve this by building a source then looking up a y-tuple index that
// exists in yTuples but has no data for the given xKey.
func TestGetMeasureCell_MissingYKey(t *testing.T) {
	// Build a source where (Cat=A, Region=North) exists but (Cat=B, Region=South) does not.
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Cat")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	// Only add (A, North) and (B, North) — South never appears.
	// But we need yTuples to contain South somehow without data for Cat=A.
	// Strategy: add (A, North) and (B, South); then for x=0 (Cat=A), y=1 (South) has no data.
	src.AddRow(map[string]any{"Cat": "A", "Region": "North", "Sales": 100})
	src.AddRow(map[string]any{"Cat": "B", "Region": "South", "Sales": 200})
	src.Build()

	// GetMeasureCell(x=0, y=1): Cat=A has no data for Region=South.
	mc := src.GetMeasureCell(0, 1)
	if mc.Text != "" {
		t.Errorf("GetMeasureCell(0,1): got %q, want empty (missing yKey)", mc.Text)
	}

	// GetMeasureCell(x=1, y=0): Cat=B has no data for Region=North.
	mc2 := src.GetMeasureCell(1, 0)
	if mc2.Text != "" {
		t.Errorf("GetMeasureCell(1,0): got %q, want empty (missing yKey)", mc2.Text)
	}
}

// ── buildGrid: totalCols <= 0 and totalRows <= 0 clamp ───────────────────────

// zeroColSource is a CubeSourceBase that reports 0 Y-axis fields and 0 data
// columns, which forces totalCols = 0 in buildGrid (triggers the clamp to 1).
type zeroColSource struct{}

func (z *zeroColSource) XAxisFieldsCount() int          { return 1 }
func (z *zeroColSource) YAxisFieldsCount() int          { return 0 } // zero header cols
func (z *zeroColSource) MeasuresCount() int             { return 1 }
func (z *zeroColSource) GetXAxisFieldName(i int) string { return "Cat" }
func (z *zeroColSource) GetYAxisFieldName(i int) string { return "" }
func (z *zeroColSource) GetMeasureName(j int) string    { return "Sales" }
func (z *zeroColSource) DataColumnCount() int           { return 0 } // zero data cols
func (z *zeroColSource) DataRowCount() int              { return 1 }
func (z *zeroColSource) MeasuresInXAxis() bool          { return false }
func (z *zeroColSource) MeasuresInYAxis() bool          { return false }
func (z *zeroColSource) MeasuresLevel() int             { return -1 }
func (z *zeroColSource) SourceAssigned() bool           { return true }
func (z *zeroColSource) TraverseXAxis(fn crossview.AxisTraverseFunc) {
	fn(crossview.AxisDrawCell{Text: "A", Cell: 0, Level: 0, SizeCell: 1, SizeLevel: 1})
}
func (z *zeroColSource) TraverseYAxis(fn crossview.AxisTraverseFunc) {}
func (z *zeroColSource) GetMeasureCell(x, y int) crossview.MeasureCell {
	return crossview.MeasureCell{Text: "42"}
}

// TestBuildGrid_TotalColsClampedToOne tests that when totalCols would be 0
// (yHeaderCols=0 and dataCols=0), buildGrid clamps it to 1.
func TestBuildGrid_TotalColsClampedToOne(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.SetSource(&zeroColSource{})

	grid, err := cv.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if grid.ColCount < 1 {
		t.Errorf("ColCount = %d, want >= 1 (should be clamped)", grid.ColCount)
	}
}

// zeroRowSource is a CubeSourceBase that reports 0 X-axis fields and 0 data
// rows, which forces totalRows = 0 in buildGrid (triggers the clamp to 1).
type zeroRowSource struct{}

func (z *zeroRowSource) XAxisFieldsCount() int          { return 0 } // zero header rows
func (z *zeroRowSource) YAxisFieldsCount() int          { return 1 }
func (z *zeroRowSource) MeasuresCount() int             { return 1 }
func (z *zeroRowSource) GetXAxisFieldName(i int) string { return "" }
func (z *zeroRowSource) GetYAxisFieldName(i int) string { return "Region" }
func (z *zeroRowSource) GetMeasureName(j int) string    { return "Sales" }
func (z *zeroRowSource) DataColumnCount() int           { return 1 }
func (z *zeroRowSource) DataRowCount() int              { return 0 } // zero data rows
func (z *zeroRowSource) MeasuresInXAxis() bool          { return false }
func (z *zeroRowSource) MeasuresInYAxis() bool          { return false }
func (z *zeroRowSource) MeasuresLevel() int             { return -1 }
func (z *zeroRowSource) SourceAssigned() bool           { return true }
func (z *zeroRowSource) TraverseXAxis(fn crossview.AxisTraverseFunc) {}
func (z *zeroRowSource) TraverseYAxis(fn crossview.AxisTraverseFunc) {
	fn(crossview.AxisDrawCell{Text: "North", Cell: 0, Level: 0, SizeCell: 1, SizeLevel: 1})
}
func (z *zeroRowSource) GetMeasureCell(x, y int) crossview.MeasureCell {
	return crossview.MeasureCell{}
}

// TestBuildGrid_TotalRowsClampedToOne tests that when totalRows would be 0
// (xHeaderRows=0 and dataRows=0), buildGrid clamps it to 1.
func TestBuildGrid_TotalRowsClampedToOne(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.SetSource(&zeroRowSource{})

	grid, err := cv.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if grid.RowCount < 1 {
		t.Errorf("RowCount = %d, want >= 1 (should be clamped)", grid.RowCount)
	}
}

// zeroColsAndRowsSource reports 0 for everything, so both totalCols and
// totalRows are forced to 0, exercising both clamp branches in one call.
type zeroColsAndRowsSource struct{}

func (z *zeroColsAndRowsSource) XAxisFieldsCount() int          { return 0 }
func (z *zeroColsAndRowsSource) YAxisFieldsCount() int          { return 0 }
func (z *zeroColsAndRowsSource) MeasuresCount() int             { return 0 }
func (z *zeroColsAndRowsSource) GetXAxisFieldName(i int) string { return "" }
func (z *zeroColsAndRowsSource) GetYAxisFieldName(i int) string { return "" }
func (z *zeroColsAndRowsSource) GetMeasureName(j int) string    { return "" }
func (z *zeroColsAndRowsSource) DataColumnCount() int           { return 0 }
func (z *zeroColsAndRowsSource) DataRowCount() int              { return 0 }
func (z *zeroColsAndRowsSource) MeasuresInXAxis() bool          { return false }
func (z *zeroColsAndRowsSource) MeasuresInYAxis() bool          { return false }
func (z *zeroColsAndRowsSource) MeasuresLevel() int             { return -1 }
func (z *zeroColsAndRowsSource) SourceAssigned() bool           { return false }
func (z *zeroColsAndRowsSource) TraverseXAxis(_ crossview.AxisTraverseFunc) {}
func (z *zeroColsAndRowsSource) TraverseYAxis(_ crossview.AxisTraverseFunc) {}
func (z *zeroColsAndRowsSource) GetMeasureCell(x, y int) crossview.MeasureCell {
	return crossview.MeasureCell{}
}

// TestBuildGrid_BothClamped tests that when both totalCols and totalRows
// would be 0, buildGrid clamps both to 1.
func TestBuildGrid_BothClamped(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.SetSource(&zeroColsAndRowsSource{})

	grid, err := cv.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if grid.ColCount < 1 {
		t.Errorf("ColCount = %d, want >= 1", grid.ColCount)
	}
	if grid.RowCount < 1 {
		t.Errorf("RowCount = %d, want >= 1", grid.RowCount)
	}
}
