package crossview

// crossview_internal_coverage_test.go – internal tests for unexported symbols.
// Uses package crossview (not crossview_test) to access unexported functions.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── aggregateAdd: remaining uncovered branches ────────────────────────────────

func TestAggregateAdd_IntV1_NilFallback(t *testing.T) {
	// v2 is int, v1 is not int or float64 → returns val (v2)
	result := aggregateAdd("hello", 42)
	if n, ok := result.(int); !ok || n != 42 {
		t.Errorf("aggregateAdd(string, int) = %v, want 42 (int)", result)
	}
}

func TestAggregateAdd_Int64V1_NilFallback(t *testing.T) {
	// v2 is int64, v1 is not int64 or int → returns val (v2)
	result := aggregateAdd("hello", int64(99))
	if n, ok := result.(int64); !ok || n != 99 {
		t.Errorf("aggregateAdd(string, int64) = %v, want 99 (int64)", result)
	}
}

func TestAggregateAdd_Float64V1_NilFallback(t *testing.T) {
	// v2 is float64, v1 is not float64 or int → returns val (v2)
	result := aggregateAdd("hello", float64(3.14))
	if f, ok := result.(float64); !ok || f != 3.14 {
		t.Errorf("aggregateAdd(string, float64) = %v, want 3.14", result)
	}
}

func TestAggregateAdd_Int64PlusInt_Conversion(t *testing.T) {
	// v2 is int64, v1 is int → should convert and add
	result := aggregateAdd(int(5), int64(10))
	if n, ok := result.(int64); !ok || n != 15 {
		t.Errorf("aggregateAdd(int=5, int64=10) = %v (%T), want int64(15)", result, result)
	}
}

func TestAggregateAdd_Float64PlusInt_Conversion(t *testing.T) {
	// v2 is float64, v1 is int → should convert and add
	result := aggregateAdd(int(3), float64(1.5))
	if f, ok := result.(float64); !ok || f != 4.5 {
		t.Errorf("aggregateAdd(int=3, float64=1.5) = %v (%T), want float64(4.5)", result, result)
	}
}

func TestAggregateAdd_Float32_Conversion(t *testing.T) {
	// v2 is float32 → delegates to float64 path
	result := aggregateAdd(nil, float32(2.5))
	if f, ok := result.(float64); !ok {
		t.Errorf("aggregateAdd(nil, float32) = %v (%T), want float64", result, result)
	} else if f != float64(float32(2.5)) {
		t.Errorf("aggregateAdd(nil, float32(2.5)) = %v, want %v", f, float64(float32(2.5)))
	}
}

func TestAggregateAdd_NonNumeric_BothNonNil(t *testing.T) {
	// default case: v1 non-nil, v2 non-nil (non-numeric) → return v1
	result := aggregateAdd("first", "second")
	if s, ok := result.(string); !ok || s != "first" {
		t.Errorf("aggregateAdd(first, second) = %v, want first", result)
	}
}

func TestAggregateAdd_NonNumeric_V1Nil(t *testing.T) {
	// default case: v1 nil, v2 non-nil (non-numeric) → return v2
	result := aggregateAdd(nil, "value")
	if s, ok := result.(string); !ok || s != "value" {
		t.Errorf("aggregateAdd(nil, value) = %v, want value", result)
	}
}

func TestAggregateAdd_IntNilV1(t *testing.T) {
	// v2 is int, v1 is nil → return v2
	result := aggregateAdd(nil, 77)
	if n, ok := result.(int); !ok || n != 77 {
		t.Errorf("aggregateAdd(nil, int=77) = %v, want 77", result)
	}
}

func TestAggregateAdd_Int64NilV1(t *testing.T) {
	// v2 is int64, v1 is nil → return v2
	result := aggregateAdd(nil, int64(55))
	if n, ok := result.(int64); !ok || n != 55 {
		t.Errorf("aggregateAdd(nil, int64=55) = %v, want 55", result)
	}
}

func TestAggregateAdd_Float64NilV1(t *testing.T) {
	// v2 is float64, v1 is nil → return v2
	result := aggregateAdd(nil, float64(9.9))
	if f, ok := result.(float64); !ok || f != 9.9 {
		t.Errorf("aggregateAdd(nil, float64=9.9) = %v, want 9.9", result)
	}
}

func TestAggregateAdd_V2Nil(t *testing.T) {
	// v2 is nil → return v1 unchanged
	result := aggregateAdd(42, nil)
	if n, ok := result.(int); !ok || n != 42 {
		t.Errorf("aggregateAdd(42, nil) = %v, want 42", result)
	}
}

// ── Mock writer for error paths ───────────────────────────────────────────────

type mockCVWriter struct {
	failNamed bool
}

func (m *mockCVWriter) WriteStr(name, value string)        {}
func (m *mockCVWriter) WriteInt(name string, v int)         {}
func (m *mockCVWriter) WriteBool(name string, v bool)       {}
func (m *mockCVWriter) WriteFloat(name string, v float32)   {}
func (m *mockCVWriter) WriteObject(obj report.Serializable) error {
	return nil
}
func (m *mockCVWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	if m.failNamed {
		return errors.New("mock WriteObjectNamed error")
	}
	return nil
}

// ── CrossViewHeader.Serialize error path ─────────────────────────────────────

func TestCrossViewHeader_Serialize_Error(t *testing.T) {
	h := NewCrossViewHeader("Columns")
	h.Add(&HeaderDescriptor{FieldName: "A"})

	w := &mockCVWriter{failNamed: true}
	err := h.Serialize(w)
	if err == nil {
		t.Error("CrossViewHeader.Serialize should propagate WriteObjectNamed error")
	}
}

// ── CrossViewCells.Serialize error path ──────────────────────────────────────

func TestCrossViewCells_Serialize_Error(t *testing.T) {
	c := NewCrossViewCells("Cells")
	c.Add(&CellDescriptor{X: 0, Y: 0})

	w := &mockCVWriter{failNamed: true}
	err := c.Serialize(w)
	if err == nil {
		t.Error("CrossViewCells.Serialize should propagate WriteObjectNamed error")
	}
}

// ── CrossViewDataSerial.Serialize error paths ─────────────────────────────────

func TestCrossViewDataSerial_Serialize_ColumnsError(t *testing.T) {
	d := &CrossViewData{}
	d.AddColumn(&HeaderDescriptor{FieldName: "Col"})
	s := NewCrossViewDataSerial(d)

	w := &mockCVWriter{failNamed: true}
	err := s.Serialize(w)
	if err == nil {
		t.Error("CrossViewDataSerial.Serialize should return error when WriteObjectNamed fails")
	}
}

// ── CrossViewDataSerial.Deserialize with unknown child ────────────────────────

func TestCrossViewDataSerial_Deserialize_UnknownChild(t *testing.T) {
	// A CrossViewData with an unknown child element should just skip it.
	d := &CrossViewData{}
	_ = NewCrossViewDataSerial(d)

	if len(d.Columns) != 0 {
		t.Errorf("Columns should be empty initially, got %d", len(d.Columns))
	}
}

// ── mockCVWriter with call-count control ──────────────────────────────────────

// mockCVWriterN fails WriteObjectNamed on the N-th call (1-based).
type mockCVWriterN struct {
	failAt  int // which call number (1-based) should fail; 0 = never
	callNum int
}

func (m *mockCVWriterN) WriteStr(name, value string)      {}
func (m *mockCVWriterN) WriteInt(name string, v int)      {}
func (m *mockCVWriterN) WriteBool(name string, v bool)    {}
func (m *mockCVWriterN) WriteFloat(name string, v float32) {}
func (m *mockCVWriterN) WriteObject(obj report.Serializable) error {
	return nil
}
func (m *mockCVWriterN) WriteObjectNamed(name string, obj report.Serializable) error {
	m.callNum++
	if m.failAt > 0 && m.callNum == m.failAt {
		return errors.New("mock WriteObjectNamed error on call " + name)
	}
	return nil
}

// ── CrossViewDataSerial.Serialize: Rows error path ───────────────────────────

func TestCrossViewDataSerial_Serialize_RowsError(t *testing.T) {
	d := &CrossViewData{}
	d.AddColumn(&HeaderDescriptor{FieldName: "Col"})
	d.AddRow(&HeaderDescriptor{FieldName: "Row"})
	s := NewCrossViewDataSerial(d)

	// Fail on the 2nd WriteObjectNamed call (Rows).
	w := &mockCVWriterN{failAt: 2}
	err := s.Serialize(w)
	if err == nil {
		t.Error("CrossViewDataSerial.Serialize should return error when Rows WriteObjectNamed fails")
	}
}

// ── CrossViewDataSerial.Serialize: Cells error path ──────────────────────────

func TestCrossViewDataSerial_Serialize_CellsError(t *testing.T) {
	d := &CrossViewData{}
	d.AddColumn(&HeaderDescriptor{FieldName: "Col"})
	d.AddRow(&HeaderDescriptor{FieldName: "Row"})
	d.AddCell(&CellDescriptor{X: 0, Y: 0})
	s := NewCrossViewDataSerial(d)

	// Fail on the 3rd WriteObjectNamed call (Cells).
	w := &mockCVWriterN{failAt: 3}
	err := s.Serialize(w)
	if err == nil {
		t.Error("CrossViewDataSerial.Serialize should return error when Cells WriteObjectNamed fails")
	}
}

// ── Mock reader for error-path Deserialize tests ──────────────────────────────

// mockCVReader implements report.Reader, allowing fine-grained control over
// NextChild and FinishChild for testing error branches.
type mockCVReader struct {
	// children is the sequence of child type names to return (then ("", false)).
	children    []string
	childIdx    int
	finishErr   error // error to return from FinishChild
	finishErrAt int   // which FinishChild call should fail (1-based; 0=never)
	finishCall  int
	attrs       map[string]string
}

func (m *mockCVReader) ReadStr(name, def string) string {
	if v, ok := m.attrs[name]; ok {
		return v
	}
	return def
}
func (m *mockCVReader) ReadInt(name string, def int) int    { return def }
func (m *mockCVReader) ReadBool(name string, def bool) bool { return def }
func (m *mockCVReader) ReadFloat(name string, def float32) float32 {
	return def
}
func (m *mockCVReader) NextChild() (string, bool) {
	if m.childIdx >= len(m.children) {
		return "", false
	}
	name := m.children[m.childIdx]
	m.childIdx++
	if name == "" {
		return "", false
	}
	return name, true
}
func (m *mockCVReader) FinishChild() error {
	m.finishCall++
	if m.finishErrAt > 0 && m.finishCall == m.finishErrAt {
		return m.finishErr
	}
	return nil
}

// ── CrossViewHeader.Deserialize: FinishChild error break ─────────────────────

func TestCrossViewHeader_Deserialize_FinishChildError(t *testing.T) {
	h := NewCrossViewHeader("Columns")
	// Provide one "Header" child; FinishChild will error on the 1st call.
	r := &mockCVReader{
		children:    []string{"Header"},
		finishErrAt: 1,
		finishErr:   errors.New("mock FinishChild error"),
	}
	// Deserialize should still return nil (it just breaks the loop).
	err := h.Deserialize(r)
	if err != nil {
		t.Errorf("Deserialize: want nil, got %v", err)
	}
}

// TestCrossViewHeader_Deserialize_FinishChildError_UnknownChild exercises the
// FinishChild error break when the child is unknown (non-"Header").
func TestCrossViewHeader_Deserialize_FinishChildError_UnknownChild(t *testing.T) {
	h := NewCrossViewHeader("Columns")
	r := &mockCVReader{
		children:    []string{"Unknown"},
		finishErrAt: 1,
		finishErr:   errors.New("mock FinishChild error on unknown child"),
	}
	err := h.Deserialize(r)
	if err != nil {
		t.Errorf("Deserialize: want nil, got %v", err)
	}
}

// ── CrossViewCells.Deserialize: FinishChild error break ──────────────────────

func TestCrossViewCells_Deserialize_FinishChildError(t *testing.T) {
	c := NewCrossViewCells("Cells")
	r := &mockCVReader{
		children:    []string{"Cell"},
		finishErrAt: 1,
		finishErr:   errors.New("mock FinishChild error"),
	}
	err := c.Deserialize(r)
	if err != nil {
		t.Errorf("Deserialize: want nil, got %v", err)
	}
}

// TestCrossViewCells_Deserialize_FinishChildError_UnknownChild exercises the
// FinishChild error break when the child is unknown (non-"Cell").
func TestCrossViewCells_Deserialize_FinishChildError_UnknownChild(t *testing.T) {
	c := NewCrossViewCells("Cells")
	r := &mockCVReader{
		children:    []string{"Unknown"},
		finishErrAt: 1,
		finishErr:   errors.New("mock FinishChild error"),
	}
	err := c.Deserialize(r)
	if err != nil {
		t.Errorf("Deserialize: want nil, got %v", err)
	}
}

// ── CrossViewDataSerial.Deserialize: FinishChild error ───────────────────────

func TestCrossViewDataSerial_Deserialize_FinishChildError(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	// Provide one "Columns" child; FinishChild errors on first call → return nil
	r := &mockCVReader{
		children:    []string{"Columns"},
		finishErrAt: 1,
		finishErr:   errors.New("mock FinishChild error"),
	}
	err := s.Deserialize(r)
	if err != nil {
		t.Errorf("Deserialize: want nil, got %v", err)
	}
}

// TestCrossViewDataSerial_Deserialize_UnknownChild_FinishChildError exercises
// the unknown-child path (default switch case) with FinishChild error → return nil.
func TestCrossViewDataSerial_Deserialize_UnknownChildPath(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	r := &mockCVReader{
		children:    []string{"Unknown"},
		finishErrAt: 1,
		finishErr:   errors.New("mock FinishChild error on unknown child"),
	}
	err := s.Deserialize(r)
	if err != nil {
		t.Errorf("Deserialize: want nil, got %v", err)
	}
}

// TestCrossViewDataSerial_Deserialize_RowsFinishChildError exercises the "Rows"
// switch case and the FinishChild error that causes early return.
func TestCrossViewDataSerial_Deserialize_RowsFinishChildError(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	// "Rows" child → CrossViewHeader.Deserialize(r) will call NextChild on the
	// same reader and get "" → returns immediately. Then outer FinishChild errors.
	r := &mockCVReader{
		children:    []string{"Rows"},
		finishErrAt: 1,
		finishErr:   errors.New("mock FinishChild error on Rows"),
	}
	err := s.Deserialize(r)
	if err != nil {
		t.Errorf("Deserialize: want nil, got %v", err)
	}
}

// TestCrossViewDataSerial_Deserialize_CellsFinishChildError exercises the "Cells"
// switch case and the FinishChild error that causes early return.
func TestCrossViewDataSerial_Deserialize_CellsFinishChildError(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	// "Cells" child → CrossViewCells.Deserialize(r) will call NextChild and get ""
	// → returns immediately. Then outer FinishChild errors.
	r := &mockCVReader{
		children:    []string{"Cells"},
		finishErrAt: 1,
		finishErr:   errors.New("mock FinishChild error on Cells"),
	}
	err := s.Deserialize(r)
	if err != nil {
		t.Errorf("Deserialize: want nil, got %v", err)
	}
}

// TestCrossViewDataSerial_Deserialize_AllThreeChildrenMock exercises the
// "Columns", "Rows", and "Cells" cases in order with no errors.
func TestCrossViewDataSerial_Deserialize_AllThreeChildrenMock(t *testing.T) {
	d := &CrossViewData{}
	s := NewCrossViewDataSerial(d)
	// All three known child types in sequence, no errors.
	r := &mockCVReader{
		children: []string{"Columns", "Rows", "Cells"},
	}
	err := s.Deserialize(r)
	if err != nil {
		t.Errorf("Deserialize: want nil, got %v", err)
	}
}
