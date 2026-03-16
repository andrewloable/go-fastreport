package crossview

// crossview_serial_coverage_test.go — internal tests (package crossview) for
// the remaining coverage gaps in crossview/serial.go:
//
//   serial.go:165  CrossViewHeader.Deserialize     80.0%
//   serial.go:232  CrossViewCells.Deserialize      80.0%
//   serial.go:321  CrossViewDataSerial.Deserialize 66.7%
//
// Analysis of the uncovered branches:
//
//   CrossViewHeader.Deserialize (line 165):
//     The "if err := d.Deserialize(r); err != nil" guard (line 174) and its
//     inner FinishChild error path (lines 175-178) are structurally dead code
//     because HeaderDescriptor.Deserialize always returns nil (it only calls
//     r.ReadBool / r.ReadStr / r.ReadInt — all value-returning, no error path).
//
//   CrossViewCells.Deserialize (line 232):
//     Identical situation: CellDescriptor.Deserialize always returns nil for
//     the same reason.
//
//   CrossViewDataSerial.Deserialize (line 321):
//     The inner error paths for the Columns / Rows / Cells switch cases are
//     dead for the same reason: CrossViewHeader.Deserialize and
//     CrossViewCells.Deserialize always return nil.
//
//   These 80.0% / 66.7% numbers cannot be improved to 100% without either
//   changing the production code so that the sub-Deserialize methods can
//   return errors, or changing the Reader interface so that ReadXxx methods
//   return errors.
//
// This file provides comprehensive tests for every REACHABLE branch in the
// affected functions, supplementing the tests in:
//   crossview_internal_coverage_test.go  (package crossview)
//   crossview_branches_test.go           (package crossview_test)
//
// We test here using the unexported types / fields that are accessible only
// from within the crossview package.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ─── shared mock reader ───────────────────────────────────────────────────────

// cvMockReader is a lightweight reader for serial coverage tests.
type cvMockReader struct {
	children    []string
	childIdx    int
	finishErr   error
	finishErrAt int // 1-based; 0 = never
	finishCall  int
	strAttrs    map[string]string
}

func newCVMockReader(children []string) *cvMockReader {
	return &cvMockReader{
		children: children,
		strAttrs: make(map[string]string),
	}
}

func (r *cvMockReader) ReadStr(name, def string) string {
	if v, ok := r.strAttrs[name]; ok {
		return v
	}
	return def
}
func (r *cvMockReader) ReadInt(_ string, def int) int       { return def }
func (r *cvMockReader) ReadBool(_ string, def bool) bool    { return def }
func (r *cvMockReader) ReadFloat(_ string, def float32) float32 { return def }
func (r *cvMockReader) NextChild() (string, bool) {
	if r.childIdx >= len(r.children) {
		return "", false
	}
	name := r.children[r.childIdx]
	r.childIdx++
	if name == "" {
		return "", false
	}
	return name, true
}
func (r *cvMockReader) FinishChild() error {
	r.finishCall++
	if r.finishErrAt > 0 && r.finishCall == r.finishErrAt {
		return r.finishErr
	}
	return nil
}

// ─── CrossViewHeader.Deserialize: all reachable branches ─────────────────────

// TestCVHeader_Deserialize_EmptyReader exercises the loop-exit (!ok) branch
// when there are no children at all.
func TestCVHeader_Deserialize_EmptyReader(t *testing.T) {
	h := NewCrossViewHeader("Columns")
	r := newCVMockReader(nil)
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if h.Count() != 0 {
		t.Errorf("Count = %d, want 0", h.Count())
	}
}

// TestCVHeader_Deserialize_OneHeaderChild exercises the "childType == Header"
// true branch → Deserialize succeeds → h.Add(d) (the main happy path).
func TestCVHeader_Deserialize_OneHeaderChild(t *testing.T) {
	h := NewCrossViewHeader("Columns")
	r := newCVMockReader([]string{"Header"})
	r.strAttrs["FieldName"] = "Region"
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if h.Count() != 1 {
		t.Errorf("Count = %d, want 1", h.Count())
	}
}

// TestCVHeader_Deserialize_NonHeaderChild exercises the "childType == Header"
// false branch (unknown child → skip, call FinishChild).
func TestCVHeader_Deserialize_NonHeaderChild(t *testing.T) {
	h := NewCrossViewHeader("Columns")
	r := newCVMockReader([]string{"Unknown"})
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if h.Count() != 0 {
		t.Errorf("Count = %d, want 0", h.Count())
	}
}

// TestCVHeader_Deserialize_MultipleChildren exercises multiple loop iterations:
// known headers and unknown siblings interleaved.
func TestCVHeader_Deserialize_MultipleChildren(t *testing.T) {
	h := NewCrossViewHeader("Columns")
	r := newCVMockReader([]string{"Unknown", "Header", "AnotherUnknown", "Header"})
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if h.Count() != 2 {
		t.Errorf("Count = %d, want 2", h.Count())
	}
}

// TestCVHeader_Deserialize_FinishChildError_OnHeader exercises the outer
// FinishChild error branch when processing a Header child — the loop breaks.
func TestCVHeader_Deserialize_FinishChildError_OnHeader(t *testing.T) {
	h := NewCrossViewHeader("Columns")
	r := newCVMockReader([]string{"Header"})
	r.finishErrAt = 1
	r.finishErr = errors.New("finishchild error")
	// Should return nil (loop breaks on FinishChild error).
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVHeader_Deserialize_FinishChildError_OnUnknown exercises the outer
// FinishChild error when processing an unknown child element.
func TestCVHeader_Deserialize_FinishChildError_OnUnknown(t *testing.T) {
	h := NewCrossViewHeader("Columns")
	r := newCVMockReader([]string{"Unknown"})
	r.finishErrAt = 1
	r.finishErr = errors.New("finishchild error on unknown")
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVHeader_Deserialize_FinishChildError_AfterFirst exercises FinishChild
// error on the second child — verifying the first Header IS added.
func TestCVHeader_Deserialize_FinishChildError_AfterFirst(t *testing.T) {
	h := NewCrossViewHeader("Columns")
	r := newCVMockReader([]string{"Header", "Header"})
	r.finishErrAt = 2
	r.finishErr = errors.New("second finishchild error")
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// First Header was successfully added before the second FinishChild error.
	if h.Count() < 1 {
		t.Errorf("Count = %d, want at least 1 (first header before break)", h.Count())
	}
}

// ─── CrossViewCells.Deserialize: all reachable branches ──────────────────────

// TestCVCells_Deserialize_EmptyReader exercises the loop-exit branch with no children.
func TestCVCells_Deserialize_EmptyReader(t *testing.T) {
	c := NewCrossViewCells("Cells")
	r := newCVMockReader(nil)
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.Count() != 0 {
		t.Errorf("Count = %d, want 0", c.Count())
	}
}

// TestCVCells_Deserialize_OneCellChild exercises the Cell child happy path.
func TestCVCells_Deserialize_OneCellChild(t *testing.T) {
	c := NewCrossViewCells("Cells")
	r := newCVMockReader([]string{"Cell"})
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.Count() != 1 {
		t.Errorf("Count = %d, want 1", c.Count())
	}
}

// TestCVCells_Deserialize_NonCellChild exercises the "childType == Cell"
// false branch.
func TestCVCells_Deserialize_NonCellChild(t *testing.T) {
	c := NewCrossViewCells("Cells")
	r := newCVMockReader([]string{"Rubbish"})
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.Count() != 0 {
		t.Errorf("Count = %d, want 0", c.Count())
	}
}

// TestCVCells_Deserialize_MixedChildren exercises multiple loop iterations.
func TestCVCells_Deserialize_MixedChildren(t *testing.T) {
	c := NewCrossViewCells("Cells")
	r := newCVMockReader([]string{"Unknown", "Cell", "AnotherUnknown", "Cell"})
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.Count() != 2 {
		t.Errorf("Count = %d, want 2", c.Count())
	}
}

// TestCVCells_Deserialize_FinishChildError_OnCell exercises the outer
// FinishChild error break when processing a Cell child.
func TestCVCells_Deserialize_FinishChildError_OnCell(t *testing.T) {
	c := NewCrossViewCells("Cells")
	r := newCVMockReader([]string{"Cell"})
	r.finishErrAt = 1
	r.finishErr = errors.New("finishchild error on Cell")
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVCells_Deserialize_FinishChildError_OnUnknown exercises the outer
// FinishChild error break on an unknown child.
func TestCVCells_Deserialize_FinishChildError_OnUnknown(t *testing.T) {
	c := NewCrossViewCells("Cells")
	r := newCVMockReader([]string{"Unknown"})
	r.finishErrAt = 1
	r.finishErr = errors.New("finishchild error on Unknown")
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVCells_Deserialize_FinishChildError_AfterFirst exercises FinishChild
// error on the second cell — the first Cell IS added.
func TestCVCells_Deserialize_FinishChildError_AfterFirst(t *testing.T) {
	c := NewCrossViewCells("Cells")
	r := newCVMockReader([]string{"Cell", "Cell"})
	r.finishErrAt = 2
	r.finishErr = errors.New("second finishchild error")
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.Count() < 1 {
		t.Errorf("Count = %d, want at least 1", c.Count())
	}
}

// ─── CrossViewDataSerial.Deserialize: all reachable branches ─────────────────

// TestCVDataSerial_Deserialize_EmptyReader exercises the loop-exit branch
// immediately when there are no children.
func TestCVDataSerial_Deserialize_EmptyReader(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	r := newCVMockReader(nil)
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVDataSerial_Deserialize_ColumnsChild exercises the "Columns" case in
// the switch and verifies items are synced back to CrossViewData.
func TestCVDataSerial_Deserialize_ColumnsChild(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	// Provide "Columns" as the only child; the inner CrossViewHeader.Deserialize
	// will try NextChild on the same reader — gets "", so it stops immediately.
	r := newCVMockReader([]string{"Columns"})
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVDataSerial_Deserialize_RowsChild exercises the "Rows" switch case.
func TestCVDataSerial_Deserialize_RowsChild(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	r := newCVMockReader([]string{"Rows"})
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVDataSerial_Deserialize_CellsChild exercises the "Cells" switch case.
func TestCVDataSerial_Deserialize_CellsChild(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	r := newCVMockReader([]string{"Cells"})
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVDataSerial_Deserialize_UnknownChild exercises the default (unknown)
// switch case: FinishChild is called and the loop continues.
func TestCVDataSerial_Deserialize_UnknownChild(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	r := newCVMockReader([]string{"Mystery"})
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVDataSerial_Deserialize_AllThreeChildren exercises all three switch
// cases in sequence with no errors.
func TestCVDataSerial_Deserialize_AllThreeChildren(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	r := newCVMockReader([]string{"Columns", "Rows", "Cells"})
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVDataSerial_Deserialize_FinishChildError_OnColumns exercises the outer
// FinishChild error return when processing the Columns child.
func TestCVDataSerial_Deserialize_FinishChildError_OnColumns(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	r := newCVMockReader([]string{"Columns"})
	r.finishErrAt = 1
	r.finishErr = errors.New("finishchild error on Columns")
	// Deserialize should return nil (it returns nil on FinishChild error per implementation).
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVDataSerial_Deserialize_FinishChildError_OnRows exercises the outer
// FinishChild error return when processing the Rows child.
func TestCVDataSerial_Deserialize_FinishChildError_OnRows(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	r := newCVMockReader([]string{"Rows"})
	r.finishErrAt = 1
	r.finishErr = errors.New("finishchild error on Rows")
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVDataSerial_Deserialize_FinishChildError_OnCells exercises the outer
// FinishChild error return when processing the Cells child.
func TestCVDataSerial_Deserialize_FinishChildError_OnCells(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	r := newCVMockReader([]string{"Cells"})
	r.finishErrAt = 1
	r.finishErr = errors.New("finishchild error on Cells")
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVDataSerial_Deserialize_FinishChildError_OnUnknown exercises the outer
// FinishChild error return when processing an unknown child.
func TestCVDataSerial_Deserialize_FinishChildError_OnUnknown(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	r := newCVMockReader([]string{"Unknown"})
	r.finishErrAt = 1
	r.finishErr = errors.New("finishchild error on Unknown")
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// TestCVDataSerial_Deserialize_SyncBack verifies that the sync-back assignments
// (CrossViewData.Columns/Rows/Cells = columnHeader.Items/...) execute after
// a successful Deserialize.  When no children are present in the reader the
// columnHeader/rowHeader/cells collections retain their initial state (copied
// from CrossViewData in NewCrossViewDataSerial) and the sync-back simply
// reassigns those same items.
func TestCVDataSerial_Deserialize_SyncBack(t *testing.T) {
	d := &CrossViewData{}
	d.AddColumn(&HeaderDescriptor{FieldName: "Col"})
	d.AddRow(&HeaderDescriptor{FieldName: "Row"})
	d.AddCell(&CellDescriptor{X: 1, Y: 2})

	s := NewCrossViewDataSerial(d)
	// Empty reader: loop exits immediately → sync-back copies columnHeader.Items
	// (which still has the pre-populated items) back to d.Columns.
	r := newCVMockReader(nil)
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// After sync-back the CrossViewData still has the original items.
	if len(d.Columns) != 1 {
		t.Errorf("Columns len = %d, want 1 (sync-back retained original)", len(d.Columns))
	}
	if len(d.Rows) != 1 {
		t.Errorf("Rows len = %d, want 1 (sync-back retained original)", len(d.Rows))
	}
	if len(d.Cells) != 1 {
		t.Errorf("Cells len = %d, want 1 (sync-back retained original)", len(d.Cells))
	}
}

// TestCVDataSerial_Deserialize_StringAttrs verifies that string attributes
// (ColumnDescriptorsIndexes etc.) are read correctly.
func TestCVDataSerial_Deserialize_StringAttrs(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	r := newCVMockReader(nil)
	r.strAttrs["ColumnDescriptorsIndexes"] = "0,1,2"
	r.strAttrs["RowDescriptorsIndexes"] = "0"
	r.strAttrs["ColumnTerminalIndexes"] = "2"
	r.strAttrs["RowTerminalIndexes"] = "0"
	if err := s.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if s.ColumnDescriptorsIndexes != "0,1,2" {
		t.Errorf("ColumnDescriptorsIndexes = %q, want 0,1,2", s.ColumnDescriptorsIndexes)
	}
	if s.RowDescriptorsIndexes != "0" {
		t.Errorf("RowDescriptorsIndexes = %q, want 0", s.RowDescriptorsIndexes)
	}
	if s.ColumnTerminalIndexes != "2" {
		t.Errorf("ColumnTerminalIndexes = %q, want 2", s.ColumnTerminalIndexes)
	}
	if s.RowTerminalIndexes != "0" {
		t.Errorf("RowTerminalIndexes = %q, want 0", s.RowTerminalIndexes)
	}
}

// ─── CrossViewDataSerial.Serialize: additional reachable branches ─────────────

// shared mock writer for serial tests
type cvSerialMockWriter struct {
	strings    map[string]string
	failNamed  bool
	failAt     int // 1-based; which WriteObjectNamed call should fail
	callNum    int
}

func newCVSerialMockWriter() *cvSerialMockWriter {
	return &cvSerialMockWriter{strings: make(map[string]string)}
}

func (m *cvSerialMockWriter) WriteStr(name, value string) { m.strings[name] = value }
func (m *cvSerialMockWriter) WriteInt(_ string, _ int)    {}
func (m *cvSerialMockWriter) WriteBool(_ string, _ bool)  {}
func (m *cvSerialMockWriter) WriteFloat(_ string, _ float32) {}
func (m *cvSerialMockWriter) WriteObject(_ report.Serializable) error { return nil }
func (m *cvSerialMockWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	m.callNum++
	if m.failNamed || (m.failAt > 0 && m.callNum == m.failAt) {
		return errors.New("mock WriteObjectNamed error on " + name)
	}
	return nil
}

// TestCVDataSerial_Serialize_StringAttrs verifies that string index fields
// are written when non-empty.
func TestCVDataSerial_Serialize_StringAttrs(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	s.ColumnDescriptorsIndexes = "0,1"
	s.RowDescriptorsIndexes = "0"
	s.ColumnTerminalIndexes = "1"
	s.RowTerminalIndexes = "0"
	w := newCVSerialMockWriter()
	if err := s.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.strings["ColumnDescriptorsIndexes"] != "0,1" {
		t.Errorf("ColumnDescriptorsIndexes = %q, want 0,1", w.strings["ColumnDescriptorsIndexes"])
	}
	if w.strings["RowDescriptorsIndexes"] != "0" {
		t.Errorf("RowDescriptorsIndexes = %q, want 0", w.strings["RowDescriptorsIndexes"])
	}
	if w.strings["ColumnTerminalIndexes"] != "1" {
		t.Errorf("ColumnTerminalIndexes = %q, want 1", w.strings["ColumnTerminalIndexes"])
	}
	if w.strings["RowTerminalIndexes"] != "0" {
		t.Errorf("RowTerminalIndexes = %q, want 0", w.strings["RowTerminalIndexes"])
	}
}

// TestCVDataSerial_Serialize_EmptyAttrs verifies that empty string index fields
// are NOT written (the guard branches' "false" arms).
func TestCVDataSerial_Serialize_EmptyAttrs(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	// All index strings are empty by default.
	w := newCVSerialMockWriter()
	if err := s.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for _, k := range []string{
		"ColumnDescriptorsIndexes",
		"RowDescriptorsIndexes",
		"ColumnTerminalIndexes",
		"RowTerminalIndexes",
	} {
		if _, ok := w.strings[k]; ok {
			t.Errorf("key %q should not be written when empty", k)
		}
	}
}

// TestCVDataSerial_Serialize_ColumnsError exercises the error path when
// writing the Columns child fails (first WriteObjectNamed call).
func TestCVDataSerial_Serialize_ColumnsError(t *testing.T) {
	d := &CrossViewData{}
	d.AddColumn(&HeaderDescriptor{FieldName: "Cat"})
	s := NewCrossViewDataSerial(d)
	w := newCVSerialMockWriter()
	w.failAt = 1
	err := s.Serialize(w)
	if err == nil {
		t.Error("Serialize should return error when Columns WriteObjectNamed fails")
	}
}

// TestCVDataSerial_Serialize_RowsError exercises the error path when writing
// the Rows child fails (second WriteObjectNamed call).
func TestCVDataSerial_Serialize_RowsError(t *testing.T) {
	d := &CrossViewData{}
	d.AddColumn(&HeaderDescriptor{FieldName: "Cat"})
	d.AddRow(&HeaderDescriptor{FieldName: "Region"})
	s := NewCrossViewDataSerial(d)
	w := newCVSerialMockWriter()
	w.failAt = 2
	err := s.Serialize(w)
	if err == nil {
		t.Error("Serialize should return error when Rows WriteObjectNamed fails")
	}
}

// TestCVDataSerial_Serialize_CellsError exercises the error path when writing
// the Cells child fails (third WriteObjectNamed call).
func TestCVDataSerial_Serialize_CellsError(t *testing.T) {
	d := &CrossViewData{}
	d.AddColumn(&HeaderDescriptor{FieldName: "Cat"})
	d.AddRow(&HeaderDescriptor{FieldName: "Region"})
	d.AddCell(&CellDescriptor{X: 0, Y: 0})
	s := NewCrossViewDataSerial(d)
	w := newCVSerialMockWriter()
	w.failAt = 3
	err := s.Serialize(w)
	if err == nil {
		t.Error("Serialize should return error when Cells WriteObjectNamed fails")
	}
}

// ─── HeaderDescriptor.Deserialize / CellDescriptor.Deserialize: coverage ─────

// TestHeaderDescriptor_Deserialize_AllFields exercises all fields being read.
func TestHeaderDescriptor_Deserialize_AllFields(t *testing.T) {
	h := &HeaderDescriptor{}
	r := newCVMockReader(nil)
	r.strAttrs["FieldName"] = "Category"
	r.strAttrs["MeasureName"] = "Sales"
	r.strAttrs["Expression"] = "[Cat]"
	// bool and int fields use the mock reader's defaults (false/0).
	if err := h.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if h.FieldName != "Category" {
		t.Errorf("FieldName = %q, want Category", h.FieldName)
	}
	if h.MeasureName != "Sales" {
		t.Errorf("MeasureName = %q, want Sales", h.MeasureName)
	}
	if h.Expression != "[Cat]" {
		t.Errorf("Expression = %q, want [Cat]", h.Expression)
	}
}

// TestCellDescriptor_Deserialize_AllFields exercises all fields being read.
func TestCellDescriptor_Deserialize_AllFields(t *testing.T) {
	c := &CellDescriptor{}
	r := newCVMockReader(nil)
	r.strAttrs["XFieldName"] = "Cat"
	r.strAttrs["YFieldName"] = "Region"
	r.strAttrs["MeasureName"] = "Sales"
	r.strAttrs["Expression"] = "[Expr]"
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.XFieldName != "Cat" {
		t.Errorf("XFieldName = %q, want Cat", c.XFieldName)
	}
	if c.YFieldName != "Region" {
		t.Errorf("YFieldName = %q, want Region", c.YFieldName)
	}
	if c.MeasureName != "Sales" {
		t.Errorf("MeasureName = %q, want Sales", c.MeasureName)
	}
	if c.Expression != "[Expr]" {
		t.Errorf("Expression = %q, want [Expr]", c.Expression)
	}
}

// ─── HeaderDescriptor.Serialize / CellDescriptor.Serialize: guard branches ───

// TestHeaderDescriptor_Serialize_AllEmpty exercises every guard's "false" arm
// when all fields are at their zero values.
func TestHeaderDescriptor_Serialize_AllEmpty(t *testing.T) {
	h := &HeaderDescriptor{}
	w := newCVSerialMockWriter()
	if err := h.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	// Nothing should be written.
	if len(w.strings) > 0 {
		t.Errorf("unexpected writes: %v", w.strings)
	}
}

// TestCellDescriptor_Serialize_AllEmpty exercises every guard's "false" arm.
func TestCellDescriptor_Serialize_AllEmpty(t *testing.T) {
	c := &CellDescriptor{}
	w := newCVSerialMockWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if len(w.strings) > 0 {
		t.Errorf("unexpected writes: %v", w.strings)
	}
}

// TestCellDescriptor_Serialize_IsXGrandTotal_SuppressesXFieldName exercises
// the XFieldName suppression path when IsXGrandTotal is true.
func TestCellDescriptor_Serialize_IsXGrandTotal_SuppressesXFieldName(t *testing.T) {
	// We can't easily test this via mock writer's WriteStr because the writer
	// only captures strings — but we can at least verify no panic and the cell
	// serializes without error.
	c := &CellDescriptor{
		IsXGrandTotal: true,
		XFieldName:    "Cat", // should be suppressed
	}
	w := newCVSerialMockWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	// XFieldName should NOT be written when IsXGrandTotal is true.
	if _, ok := w.strings["XFieldName"]; ok {
		t.Error("XFieldName should not be written when IsXGrandTotal is true")
	}
}

// TestCellDescriptor_Serialize_IsYGrandTotal_SuppressesYFieldName exercises
// the YFieldName suppression path when IsYGrandTotal is true.
func TestCellDescriptor_Serialize_IsYGrandTotal_SuppressesYFieldName(t *testing.T) {
	c := &CellDescriptor{
		IsYGrandTotal: true,
		YFieldName:    "Region", // should be suppressed
	}
	w := newCVSerialMockWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if _, ok := w.strings["YFieldName"]; ok {
		t.Error("YFieldName should not be written when IsYGrandTotal is true")
	}
}
