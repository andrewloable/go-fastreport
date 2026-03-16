package matrix

// matrix_gaps_internal_test.go covers the remaining uncovered branches in
// matrix.go that require mock reader/writer implementations:
//
//   1. MatrixObject.Serialize line 472:
//      `if err := m.TableBase.Serialize(w); err != nil { return err }`
//      Triggered when the writer's WriteObject call fails while TableBase
//      serializes its label column.
//
//   2. DeserializeChild lines 578, 590, 607:
//      `if r.FinishChild() != nil { break }`
//      Triggered via a mock reader that returns an error from FinishChild
//      inside the MatrixRows, MatrixColumns, and MatrixCells inner loops.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── writeObjectFailWriter ─────────────────────────────────────────────────────
// writeObjectFailWriter is a report.Writer whose WriteObject always returns an
// error. This causes TableBase.Serialize to fail (it calls WriteObject for each
// column and row), which propagates to MatrixObject.Serialize.

type writeObjectFailWriter struct{}

func (w *writeObjectFailWriter) WriteStr(name, value string)             {}
func (w *writeObjectFailWriter) WriteInt(name string, value int)          {}
func (w *writeObjectFailWriter) WriteBool(name string, value bool)        {}
func (w *writeObjectFailWriter) WriteFloat(name string, value float32)    {}
func (w *writeObjectFailWriter) WriteObject(_ report.Serializable) error {
	return errors.New("simulated WriteObject failure")
}
func (w *writeObjectFailWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	return obj.Serialize(w)
}

// TestMatrixObject_Serialize_TableBaseError exercises the
// `if err := m.TableBase.Serialize(w); err != nil { return err }` branch.
//
// BuildTemplate populates TableBase columns and rows via AddData. When the
// writer's WriteObject returns an error for those columns, TableBase.Serialize
// fails and MatrixObject.Serialize propagates the error at line 472.
func TestMatrixObject_Serialize_TableBaseError(t *testing.T) {
	m := New()
	m.Data.AddCell(NewCellDescriptor("[V]", AggregateFunctionSum))
	m.AddData("r", "c", []any{1.0})
	m.BuildTemplate() // populates TableBase columns and rows

	w := &writeObjectFailWriter{}
	err := m.Serialize(w)
	if err == nil {
		t.Error("expected error when TableBase.Serialize fails (WriteObject error), got nil")
	}
}

// ── mockDeserializeChildReader ────────────────────────────────────────────────
// mockDeserializeChildReader is a minimal report.Reader implementation that
// drives DeserializeChild with fine-grained control over NextChild and
// FinishChild return values.

type mockDeserializeChildReader struct {
	// nextCalls is the sequence of (typeName, ok) values returned by NextChild.
	nextCalls []struct {
		name string
		ok   bool
	}
	nextIdx int

	// finishErrs maps FinishChild call index (0-based) to the error to return.
	finishErrs  map[int]error
	finishCalls int

	// attrs holds attribute values for ReadStr/ReadInt/ReadBool/ReadFloat.
	attrs map[string]string
}

func (r *mockDeserializeChildReader) ReadStr(name, def string) string {
	if v, ok := r.attrs[name]; ok {
		return v
	}
	return def
}
func (r *mockDeserializeChildReader) ReadInt(name string, def int) int       { return def }
func (r *mockDeserializeChildReader) ReadBool(name string, def bool) bool    { return def }
func (r *mockDeserializeChildReader) ReadFloat(name string, def float32) float32 { return def }

func (r *mockDeserializeChildReader) NextChild() (string, bool) {
	if r.nextIdx >= len(r.nextCalls) {
		return "", false
	}
	c := r.nextCalls[r.nextIdx]
	r.nextIdx++
	return c.name, c.ok
}

func (r *mockDeserializeChildReader) FinishChild() error {
	n := r.finishCalls
	r.finishCalls++
	if r.finishErrs != nil {
		if err, ok := r.finishErrs[n]; ok {
			return err
		}
	}
	return nil
}

// ── DeserializeChild MatrixRows: FinishChild error branch ─────────────────────

// TestDeserializeChild_MatrixRows_FinishChildError exercises the
// `if r.FinishChild() != nil { break }` branch inside the MatrixRows loop.
//
// Sequence: NextChild returns "Header" (ok=true), then the loop calls
// FinishChild which returns an error → break exits the loop.
func TestDeserializeChild_MatrixRows_FinishChildError(t *testing.T) {
	m := New()
	r := &mockDeserializeChildReader{
		nextCalls: []struct {
			name string
			ok   bool
		}{
			{name: "Header", ok: true}, // first child: a Header
			// After FinishChild fails, the loop breaks — NextChild not called again.
		},
		finishErrs: map[int]error{
			0: errors.New("forced FinishChild error in MatrixRows"),
		},
		attrs: map[string]string{
			"Expression": "[RowsFinishFail]",
		},
	}

	consumed := m.DeserializeChild("MatrixRows", r)
	if !consumed {
		t.Error("DeserializeChild(MatrixRows) should return true")
	}
	// The Header was added before FinishChild was called.
	if len(m.Data.Rows) != 1 {
		t.Errorf("Data.Rows = %d, want 1 (added before FinishChild error)", len(m.Data.Rows))
	}
}

// ── DeserializeChild MatrixColumns: FinishChild error branch ──────────────────

// TestDeserializeChild_MatrixColumns_FinishChildError exercises the
// `if r.FinishChild() != nil { break }` branch inside the MatrixColumns loop.
func TestDeserializeChild_MatrixColumns_FinishChildError(t *testing.T) {
	m := New()
	r := &mockDeserializeChildReader{
		nextCalls: []struct {
			name string
			ok   bool
		}{
			{name: "Header", ok: true},
		},
		finishErrs: map[int]error{
			0: errors.New("forced FinishChild error in MatrixColumns"),
		},
		attrs: map[string]string{
			"Expression": "[ColsFinishFail]",
		},
	}

	consumed := m.DeserializeChild("MatrixColumns", r)
	if !consumed {
		t.Error("DeserializeChild(MatrixColumns) should return true")
	}
	if len(m.Data.Columns) != 1 {
		t.Errorf("Data.Columns = %d, want 1 (added before FinishChild error)", len(m.Data.Columns))
	}
}

// ── DeserializeChild MatrixCells: FinishChild error branch ────────────────────

// TestDeserializeChild_MatrixCells_FinishChildError exercises the
// `if r.FinishChild() != nil { break }` branch inside the MatrixCells loop.
func TestDeserializeChild_MatrixCells_FinishChildError(t *testing.T) {
	m := New()
	r := &mockDeserializeChildReader{
		nextCalls: []struct {
			name string
			ok   bool
		}{
			{name: "Cell", ok: true},
		},
		finishErrs: map[int]error{
			0: errors.New("forced FinishChild error in MatrixCells"),
		},
		attrs: map[string]string{
			"Expression": "[CellsFinishFail]",
		},
	}

	consumed := m.DeserializeChild("MatrixCells", r)
	if !consumed {
		t.Error("DeserializeChild(MatrixCells) should return true")
	}
	if len(m.Data.Cells) != 1 {
		t.Errorf("Data.Cells = %d, want 1 (added before FinishChild error)", len(m.Data.Cells))
	}
}

// ── DeserializeChild: non-Header/Cell child with FinishChild error ────────────

// TestDeserializeChild_MatrixRows_NonHeaderFinishChildError exercises the
// `if r.FinishChild() != nil { break }` branch when the child is NOT a "Header"
// (so the Header is not added, but FinishChild is still called and fails).
func TestDeserializeChild_MatrixRows_NonHeaderFinishChildError(t *testing.T) {
	m := New()
	r := &mockDeserializeChildReader{
		nextCalls: []struct {
			name string
			ok   bool
		}{
			{name: "Noise", ok: true}, // not a Header
		},
		finishErrs: map[int]error{
			0: errors.New("forced FinishChild error for non-Header in MatrixRows"),
		},
	}

	consumed := m.DeserializeChild("MatrixRows", r)
	if !consumed {
		t.Error("DeserializeChild(MatrixRows) should return true")
	}
	// No Header was added since child type was "Noise".
	if len(m.Data.Rows) != 0 {
		t.Errorf("Data.Rows = %d, want 0 (Noise child is not a Header)", len(m.Data.Rows))
	}
}
