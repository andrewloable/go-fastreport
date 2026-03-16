package table

// table_finishchild_test.go — internal test targeting the inner-loop break
// in deserialize.go:35 when FinishChild fails while draining cell sub-children.
//
// Target: deserialize.go:35 `if r.FinishChild() != nil { break }` — the break
// branch inside the inner sub-child loop of the TableCell processing code.

import (
	"errors"
	"testing"
)

// cellSubChildFailReader simulates a reader positioned on a TableRow that
// contains one TableCell, and that TableCell has one sub-child. FinishChild
// fails when called inside the inner sub-child drain loop (line 35).
//
// State machine:
//   state 0 (outer loop): NextChild → "TableCell"; transition to state 1
//   state 1 (inner loop, cell sub-children): NextChild → "SubChild" (moreKids=true); transition to state 2
//   state 2 (done): NextChild → "", false
//   FinishChild at state 2 (inner loop after seeing sub-child): returns error
type cellSubChildFailReader struct {
	state         int // 0=outer row level, 1=cell sub-child level, 2=done
	finishErr     error
	finishCalls   int
	failOnFinish  int // which FinishChild call (0-based) to fail
}

func (r *cellSubChildFailReader) ReadStr(name, def string) string       { return def }
func (r *cellSubChildFailReader) ReadInt(name string, def int) int      { return def }
func (r *cellSubChildFailReader) ReadBool(name string, def bool) bool   { return def }
func (r *cellSubChildFailReader) ReadFloat(name string, def float32) float32 { return def }

func (r *cellSubChildFailReader) NextChild() (string, bool) {
	switch r.state {
	case 0:
		// Outer loop: return TableCell once.
		r.state = 1
		return "TableCell", true
	case 1:
		// Inner loop (cell sub-children): return one sub-child.
		r.state = 2
		return "SubChild", true
	default:
		// No more children.
		return "", false
	}
}

func (r *cellSubChildFailReader) FinishChild() error {
	n := r.finishCalls
	r.finishCalls++
	if n == r.failOnFinish {
		return r.finishErr
	}
	return nil
}

// TestDeserializeChild_CellSubChildFinishChildError exercises the inner
// `if r.FinishChild() != nil { break }` on deserialize.go line 35.
// It requires: a TableCell child that itself has a sub-child, and FinishChild
// fails when called inside the inner drain loop for that sub-child.
func TestDeserializeChild_CellSubChildFinishChildError(t *testing.T) {
	r := &cellSubChildFailReader{
		finishErr:    errors.New("forced inner FinishChild error"),
		failOnFinish: 0, // fail on the first FinishChild call (inner loop)
	}

	tbl := NewTableObject()
	consumed := tbl.DeserializeChild("TableRow", r)
	if !consumed {
		t.Error("DeserializeChild(TableRow) should return true")
	}
	// The row is still appended even after FinishChild error (break exits inner loop).
	if tbl.RowCount() != 1 {
		t.Errorf("RowCount: got %d, want 1", tbl.RowCount())
	}
	// The cell was added before the inner drain loop tried to finish the sub-child.
	row := tbl.Row(0)
	if row.CellCount() != 1 {
		t.Errorf("CellCount: got %d, want 1", row.CellCount())
	}
}
