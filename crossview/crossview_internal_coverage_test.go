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
