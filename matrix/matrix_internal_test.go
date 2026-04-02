package matrix

// Internal tests for unexported types and functions.
// These tests are in the same package so they can access unexported types.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── errorWriter: a report.Writer that always fails on WriteObjectNamed ─────────

type errorWriter struct {
	failOn string // element name that triggers failure
}

func (e *errorWriter) WriteStr(name, value string)             {}
func (e *errorWriter) WriteInt(name string, value int)          {}
func (e *errorWriter) WriteBool(name string, value bool)        {}
func (e *errorWriter) WriteFloat(name string, value float32)    {}
func (e *errorWriter) WriteObject(obj report.Serializable) error { return nil }
func (e *errorWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	if name == e.failOn {
		return fmt.Errorf("simulated error on %s", name)
	}
	return nil
}

// ── newHeaderItem ─────────────────────────────────────────────────────────────

func TestNewHeaderItem(t *testing.T) {
	h := newHeaderItem("hello")
	if h.Value != "hello" {
		t.Errorf("Value = %q, want hello", h.Value)
	}
	if h.childIndex == nil {
		t.Error("childIndex should be initialized")
	}
	if len(h.Children) != 0 {
		t.Errorf("Children should be empty, got %d", len(h.Children))
	}
}

// ── ensureChild ───────────────────────────────────────────────────────────────

func TestEnsureChild_CreatesNew(t *testing.T) {
	h := newHeaderItem("root")
	child := h.ensureChild("A")
	if child.Value != "A" {
		t.Errorf("child.Value = %q, want A", child.Value)
	}
	if len(h.Children) != 1 {
		t.Errorf("Children len = %d, want 1", len(h.Children))
	}
}

func TestEnsureChild_ReturnsExisting(t *testing.T) {
	h := newHeaderItem("root")
	c1 := h.ensureChild("A")
	c2 := h.ensureChild("A")
	if c1 != c2 {
		t.Error("ensureChild should return the same child for the same value")
	}
	if len(h.Children) != 1 {
		t.Errorf("Children len = %d, want 1 (no duplicate)", len(h.Children))
	}
}

func TestEnsureChild_MultipleChildren(t *testing.T) {
	h := newHeaderItem("root")
	h.ensureChild("A")
	h.ensureChild("B")
	h.ensureChild("C")
	if len(h.Children) != 3 {
		t.Errorf("Children len = %d, want 3", len(h.Children))
	}
}

// ── isLeaf ────────────────────────────────────────────────────────────────────

func TestIsLeaf_NoChildren(t *testing.T) {
	h := newHeaderItem("leaf")
	if !h.isLeaf() {
		t.Error("item with no children should be leaf")
	}
}

func TestIsLeaf_WithChildren(t *testing.T) {
	h := newHeaderItem("root")
	h.ensureChild("child")
	if h.isLeaf() {
		t.Error("item with children should not be leaf")
	}
}

// ── computeSizes ──────────────────────────────────────────────────────────────

func TestComputeSizes_Leaf(t *testing.T) {
	h := newHeaderItem("leaf")
	h.computeSizes()
	if h.CellSize != 1 {
		t.Errorf("CellSize = %d, want 1", h.CellSize)
	}
	if h.LevelSize != 1 {
		t.Errorf("LevelSize = %d, want 1", h.LevelSize)
	}
}

func TestComputeSizes_OneLevel(t *testing.T) {
	root := newHeaderItem("")
	root.ensureChild("A")
	root.ensureChild("B")
	root.computeSizes()
	if root.CellSize != 2 {
		t.Errorf("CellSize = %d, want 2", root.CellSize)
	}
	if root.LevelSize != 2 {
		t.Errorf("LevelSize = %d, want 2", root.LevelSize)
	}
}

func TestComputeSizes_TwoLevels(t *testing.T) {
	root := newHeaderItem("")
	a := root.ensureChild("A")
	a.ensureChild("A1")
	a.ensureChild("A2")
	root.ensureChild("B")
	root.computeSizes()
	// A has 2 leaves (A1, A2). B has 1 leaf. Total = 3.
	if root.CellSize != 3 {
		t.Errorf("root.CellSize = %d, want 3", root.CellSize)
	}
	// root level: 1 (its children) + 1 (their children) + 1 (root) = LevelSize 3.
	if root.LevelSize != 3 {
		t.Errorf("root.LevelSize = %d, want 3", root.LevelSize)
	}
	if a.CellSize != 2 {
		t.Errorf("a.CellSize = %d, want 2", a.CellSize)
	}
}

// ── leaves ────────────────────────────────────────────────────────────────────

func TestLeaves_SingleLeaf(t *testing.T) {
	h := newHeaderItem("only")
	ls := h.leaves()
	if len(ls) != 1 || ls[0] != h {
		t.Errorf("leaves = %v, want [only]", ls)
	}
}

func TestLeaves_Tree(t *testing.T) {
	root := newHeaderItem("")
	a := root.ensureChild("A")
	a.ensureChild("A1")
	a.ensureChild("A2")
	root.ensureChild("B")

	ls := root.leaves()
	// Should be [A1, A2, B] in order.
	if len(ls) != 3 {
		t.Fatalf("leaves len = %d, want 3", len(ls))
	}
	if ls[0].Value != "A1" || ls[1].Value != "A2" || ls[2].Value != "B" {
		t.Errorf("leaves = [%s %s %s], want [A1 A2 B]", ls[0].Value, ls[1].Value, ls[2].Value)
	}
}

// ── findNodeAtLevel ───────────────────────────────────────────────────────────

func TestFindNodeAtLevel_NilRoot(t *testing.T) {
	result := findNodeAtLevel(nil, []string{"A"}, 0)
	if result != nil {
		t.Error("expected nil for nil root")
	}
}

func TestFindNodeAtLevel_LevelBeyondPath(t *testing.T) {
	root := newHeaderItem("")
	root.ensureChild("A")
	result := findNodeAtLevel(root, []string{"A"}, 5)
	if result != nil {
		t.Error("expected nil when level >= len(path)")
	}
}

func TestFindNodeAtLevel_Found(t *testing.T) {
	root := newHeaderItem("")
	a := root.ensureChild("A")
	a.ensureChild("A1")

	// Level 0: should return "A".
	n0 := findNodeAtLevel(root, []string{"A", "A1"}, 0)
	if n0 == nil || n0.Value != "A" {
		t.Errorf("level 0: got %v, want A", n0)
	}
	// Level 1: should return "A1".
	n1 := findNodeAtLevel(root, []string{"A", "A1"}, 1)
	if n1 == nil || n1.Value != "A1" {
		t.Errorf("level 1: got %v, want A1", n1)
	}
}

func TestFindNodeAtLevel_NotFound(t *testing.T) {
	root := newHeaderItem("")
	root.ensureChild("A")
	// "B" doesn't exist.
	result := findNodeAtLevel(root, []string{"B"}, 0)
	if result != nil {
		t.Error("expected nil for missing path segment")
	}
}

// ── joinPath ──────────────────────────────────────────────────────────────────

func TestJoinPath_Empty(t *testing.T) {
	result := joinPath(nil)
	if result != "" {
		t.Errorf("joinPath(nil) = %q, want empty", result)
	}
}

func TestJoinPath_Single(t *testing.T) {
	result := joinPath([]string{"hello"})
	if result != "hello" {
		t.Errorf("joinPath([hello]) = %q, want hello", result)
	}
}

func TestJoinPath_Multiple(t *testing.T) {
	// Verify segments are separated and different combinations produce different keys.
	r1 := joinPath([]string{"ab", "c"})
	r2 := joinPath([]string{"a", "bc"})
	if r1 == r2 {
		t.Error("joinPath should distinguish ab+c from a+bc")
	}
}

// ── Deserialize stubs on unexported types ──────────────────────────────────────

func TestHeaderHolder_Deserialize(t *testing.T) {
	h := &headerHolder{}
	// Deserialize is a no-op stub — just verify it returns nil.
	r := serial.NewReader(strings.NewReader(""))
	if err := h.Deserialize(r); err != nil {
		t.Errorf("headerHolder.Deserialize: %v", err)
	}
}

func TestHeaderDescriptorWriter_Deserialize(t *testing.T) {
	hw := &headerDescriptorWriter{h: NewHeaderDescriptor("[V]")}
	r := serial.NewReader(strings.NewReader(""))
	if err := hw.Deserialize(r); err != nil {
		t.Errorf("headerDescriptorWriter.Deserialize: %v", err)
	}
}

func TestCellHolder_Deserialize(t *testing.T) {
	c := &cellHolder{}
	r := serial.NewReader(strings.NewReader(""))
	if err := c.Deserialize(r); err != nil {
		t.Errorf("cellHolder.Deserialize: %v", err)
	}
}

func TestCellDescriptorWriter_Deserialize(t *testing.T) {
	cw := &cellDescriptorWriter{c: NewCellDescriptor("[V]", AggregateFunctionSum)}
	r := serial.NewReader(strings.NewReader(""))
	if err := cw.Deserialize(r); err != nil {
		t.Errorf("cellDescriptorWriter.Deserialize: %v", err)
	}
}

// ── accumulator.result — zero count Avg path ──────────────────────────────────

func TestAccumulatorResult_AvgZeroCount(t *testing.T) {
	// Create a fresh accumulator for Avg without any adds (count=0).
	a := newAccumulator(AggregateFunctionAvg)
	result := a.result()
	if result != 0 {
		t.Errorf("Avg with count=0: result = %v, want 0", result)
	}
}

// ── toFloat — all branches ────────────────────────────────────────────────────

func TestToFloat_Float64(t *testing.T) {
	if got := toFloat(float64(3.14)); got != 3.14 {
		t.Errorf("float64: %v", got)
	}
}

func TestToFloat_Float32(t *testing.T) {
	var f float32 = 2.5
	if got := toFloat(f); got != float64(f) {
		t.Errorf("float32: %v", got)
	}
}

func TestToFloat_Int(t *testing.T) {
	if got := toFloat(int(7)); got != 7 {
		t.Errorf("int: %v", got)
	}
}

func TestToFloat_Int64(t *testing.T) {
	if got := toFloat(int64(100)); got != 100 {
		t.Errorf("int64: %v", got)
	}
}

func TestToFloat_Int32(t *testing.T) {
	if got := toFloat(int32(42)); got != 42 {
		t.Errorf("int32: %v", got)
	}
}

func TestToFloat_Unknown(t *testing.T) {
	if got := toFloat("not-a-number"); got != 0 {
		t.Errorf("unknown: %v, want 0", got)
	}
}

// ── AddDataMultiLevel — cells guard (i >= len) ─────────────────────────────────

func TestAddDataMultiLevel_NoCellDescriptors(t *testing.T) {
	m := New()
	// No cell descriptors added — loop in AddDataMultiLevel breaks immediately.
	m.AddDataMultiLevel([]string{"r"}, []string{"c"}, []any{1.0, 2.0, 3.0})
	// Should not panic; CellResultMultiLevel should error (no accumulator created).
	_, err := m.CellResultMultiLevel([]string{"r"}, []string{"c"}, 0)
	if err == nil {
		t.Error("expected error since no cell descriptors, accumulator not created")
	}
}

// ── AddData — cells guard (i >= len) ──────────────────────────────────────────

func TestAddData_ValuesExceedDescriptors(t *testing.T) {
	// Values slice has more entries than cell descriptors — excess should be ignored.
	m := New()
	m.Data.AddCell(NewCellDescriptor("[V]", AggregateFunctionSum))
	// Provide 3 values but only 1 descriptor.
	m.AddData("r", "c", []any{5.0, 99.0, 999.0})
	result, err := m.CellResult("r", "c", 0)
	if err != nil {
		t.Fatalf("CellResult: %v", err)
	}
	if result != 5 {
		t.Errorf("result = %v, want 5 (extra values ignored)", result)
	}
}

// ── headerDescriptorWriter.Serialize — all branches ───────────────────────────

func TestHeaderDescriptorWriter_Serialize_AllFlags(t *testing.T) {
	// Exercise all the conditional branches in headerDescriptorWriter.Serialize.
	m := New()
	h := NewHeaderDescriptor("[V]")
	h.Sort = SortOrderNone     // != Ascending → writes Sort
	h.Totals = false           // != true → writes Totals=false
	h.TotalsFirst = true       // → writes TotalsFirst=true
	h.PageBreak = true         // → writes PageBreak=true
	h.SuppressTotals = true    // → writes SuppressTotals=true
	m.Data.AddRow(h)

	var buf strings.Builder
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("MatrixObject", m); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"Sort=", "Totals=", "TotalsFirst=", "PageBreak=", "SuppressTotals="} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}

func TestHeaderDescriptorWriter_Serialize_EmptyExpression(t *testing.T) {
	// When Expression is "", it should NOT be written.
	m := New()
	h := NewHeaderDescriptor("")
	m.Data.AddRow(h)

	var buf strings.Builder
	w := serial.NewWriter(&buf)
	_ = w.WriteObjectNamed("MatrixObject", m)
	_ = w.Flush()
	out := buf.String()
	if strings.Contains(out, "Expression=") {
		t.Errorf("expected no Expression attr for empty expression:\n%s", out)
	}
}

func TestCellDescriptorWriter_Serialize_EmptyExpression(t *testing.T) {
	// When Expression is "", it should NOT be written.
	m := New()
	m.Data.AddCell(NewCellDescriptor("", AggregateFunctionSum))

	var buf strings.Builder
	w := serial.NewWriter(&buf)
	_ = w.WriteObjectNamed("MatrixObject", m)
	_ = w.Flush()
	out := buf.String()
	// The MatrixCells block should exist but no Expression attr.
	if strings.Contains(out, "Expression=") {
		t.Errorf("expected no Expression attr:\n%s", out)
	}
}

// ── headerHolder.Serialize — error path ───────────────────────────────────────

func TestHeaderHolder_Serialize_ErrorPath(t *testing.T) {
	h := &headerHolder{headers: []*HeaderDescriptor{NewHeaderDescriptor("[V]")}}
	ew := &errorWriter{failOn: "Header"}
	err := h.Serialize(ew)
	if err == nil {
		t.Error("expected error when WriteObjectNamed fails")
	}
}

// ── cellHolder.Serialize — error path ─────────────────────────────────────────

func TestCellHolder_Serialize_ErrorPath(t *testing.T) {
	c := &cellHolder{cells: []*CellDescriptor{NewCellDescriptor("[V]", AggregateFunctionSum)}}
	ew := &errorWriter{failOn: "Cell"}
	err := c.Serialize(ew)
	if err == nil {
		t.Error("expected error when WriteObjectNamed fails")
	}
}

// ── MatrixObject.Serialize — WriteObjectNamed error paths ─────────────────────

// fullErrorWriter is a report.Writer that calls Serialize on obj for non-failing
// elements (to exercise the full code path), but returns error for specific names.
type fullErrorWriter struct {
	failOn string
}

func (f *fullErrorWriter) WriteStr(name, value string)          {}
func (f *fullErrorWriter) WriteInt(name string, value int)       {}
func (f *fullErrorWriter) WriteBool(name string, value bool)     {}
func (f *fullErrorWriter) WriteFloat(name string, value float32) {}
func (f *fullErrorWriter) WriteObject(obj report.Serializable) error {
	return obj.Serialize(f)
}
func (f *fullErrorWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	if name == f.failOn {
		return fmt.Errorf("simulated error on %s", name)
	}
	return obj.Serialize(f)
}

func TestMatrixObject_Serialize_MatrixRowsError(t *testing.T) {
	m := New()
	m.Data.AddRow(NewHeaderDescriptor("[V]"))
	ew := &fullErrorWriter{failOn: "MatrixRows"}
	err := m.Serialize(ew)
	if err == nil {
		t.Error("expected error when MatrixRows WriteObjectNamed fails")
	}
}

func TestMatrixObject_Serialize_MatrixColumnsError(t *testing.T) {
	m := New()
	m.Data.AddColumn(NewHeaderDescriptor("[V]"))
	ew := &fullErrorWriter{failOn: "MatrixColumns"}
	err := m.Serialize(ew)
	if err == nil {
		t.Error("expected error when MatrixColumns WriteObjectNamed fails")
	}
}

func TestMatrixObject_Serialize_MatrixCellsError(t *testing.T) {
	m := New()
	m.Data.AddCell(NewCellDescriptor("[V]", AggregateFunctionSum))
	ew := &fullErrorWriter{failOn: "MatrixCells"}
	err := m.Serialize(ew)
	if err == nil {
		t.Error("expected error when MatrixCells WriteObjectNamed fails")
	}
}

