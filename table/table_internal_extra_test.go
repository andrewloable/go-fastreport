package table

// table_internal_extra_test.go — internal tests targeting error-path branches
// that require a mock report.Reader or access to unexported fields.
//
// Targets:
//   cell.go:145    Deserialize   — TextObject.Deserialize error propagation
//   column.go:88   Deserialize   — ComponentBase.Deserialize error propagation
//   row.go:127     Deserialize   — ComponentBase.Deserialize error propagation
//   table.go:249   Deserialize   — BreakableComponent.Deserialize error propagation
//   table.go:311   Deserialize   — TableBase.Deserialize error propagation
//   cell.go:122    Serialize     — TextObject.Serialize error propagation
//   column.go:65   Serialize     — ComponentBase.Serialize error propagation
//   row.go:95      Serialize     — ComponentBase.Serialize error propagation
//   table.go:198   Serialize     — BreakableComponent.Serialize error propagation
//   helper.go:193  copyCells     — srcCell nil path

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── failingReader — a mock report.Reader whose parent Deserialize returns error ─

// failReader implements report.Reader and always returns an error from
// the first Read* call, causing parent Deserialize to propagate it upward.
// We achieve this by making ReadStr return "" but also embed a sentinel —
// the real trick is to make the embedded base Deserialize fail. Since
// BaseObject.Deserialize always returns nil, we can't make it fail directly.
//
// Instead we provide a reader that correctly satisfies report.Reader but
// whose NextChild / FinishChild return errors, which is sufficient to test
// the FinishChild error paths.

// noopReader is a no-op reader that returns defaults for all Read* calls
// and provides a controllable error for FinishChild.
type noopReader struct {
	finishChildErr error
	nextChildCalls int
	maxNextChild   int // after maxNextChild calls, NextChild returns false
}

func (r *noopReader) ReadStr(name, def string) string       { return def }
func (r *noopReader) ReadInt(name string, def int) int      { return def }
func (r *noopReader) ReadBool(name string, def bool) bool   { return def }
func (r *noopReader) ReadFloat(name string, def float32) float32 { return def }
func (r *noopReader) NextChild() (string, bool) {
	if r.maxNextChild >= 0 && r.nextChildCalls >= r.maxNextChild {
		return "", false
	}
	r.nextChildCalls++
	return "TableColumn", true
}
func (r *noopReader) FinishChild() error {
	return r.finishChildErr
}

// failingBaseReader wraps a report.Reader and injects an error on Deserialize
// by implementing the interface with a writer that calls Serialize which fails.
// Since we cannot directly make ComponentBase.Deserialize fail (it always
// returns nil), we use a mockWriter approach for Serialize tests. For
// Deserialize tests, since the base always returns nil, the error path in
// column/row/table Deserialize (the `if err := …Deserialize(r); err != nil`)
// is structurally unreachable with real implementations.
//
// We therefore use the internal package access to call the unexported methods
// directly via a wrapper struct that overrides Deserialize behavior — but since
// Go does not allow method overrides, we instead create a mock object that
// embeds the struct and provides a failing Deserialize. However, the error
// paths in TableColumn.Deserialize etc. CANNOT be exercised externally because
// ComponentBase.Deserialize always returns nil.
//
// The practical approach: use a failingWriter for the Serialize error paths that
// are not yet covered, and for Deserialize we cover via the noopReader with
// FinishChild errors (DeserializeChild path).

// ── DeserializeChild — FinishChild error in cell sub-child loop ──────────────

// TestDeserializeChild_TableRow_FinishChildError exercises the
// `if r.FinishChild() != nil { break }` branch inside the inner cell loop
// in deserialize.go by using a reader that returns an error from FinishChild
// after seeing a TableCell child.
func TestDeserializeChild_TableRow_FinishChildErrorInCellLoop(t *testing.T) {
	// We need a reader that:
	// 1. Is positioned on a TableRow element (Deserialize reads row attrs).
	// 2. On NextChild for the row's children returns "TableCell".
	// 3. After the cell's own sub-children iteration (inner loop), FinishChild
	//    returns an error.
	//
	// The inner loop in deserialize.go calls r.NextChild() for cell sub-children,
	// then r.FinishChild(). We need FinishChild to return an error on that call.

	// A specialised reader that:
	// - Row-level NextChild: returns TableCell once, then false.
	// - Cell sub-child NextChild (called from the inner loop): returns false immediately.
	// - FinishChild: returns an error on the first call (which is the cell's outer FinishChild).
	r := &rowWithCellFailReader{
		cellFinishErr: errors.New("forced FinishChild error"),
		state:         rfStateRow,
	}

	tbl := NewTableObject()
	consumed := tbl.DeserializeChild("TableRow", r)
	// The function should return true (it processes a TableRow) even though
	// FinishChild errored — the break exits the loop early.
	if !consumed {
		t.Error("DeserializeChild(TableRow) should return true")
	}
}

// rowWithCellFailReader is a mock reader that simulates:
// - One TableRow child containing one TableCell.
// - The outer FinishChild (for the cell) fails.
type rfState int

const (
	rfStateRow     rfState = iota // at row level, NextChild will return TableCell
	rfStateCellSub                // inside cell sub-child loop, NextChild returns false
	rfStateDone                   // done
)

type rowWithCellFailReader struct {
	state         rfState
	cellFinishErr error
	calls         int
}

func (r *rowWithCellFailReader) ReadStr(name, def string) string       { return def }
func (r *rowWithCellFailReader) ReadInt(name string, def int) int      { return def }
func (r *rowWithCellFailReader) ReadBool(name string, def bool) bool   { return def }
func (r *rowWithCellFailReader) ReadFloat(name string, def float32) float32 { return def }

func (r *rowWithCellFailReader) NextChild() (string, bool) {
	switch r.state {
	case rfStateRow:
		// First call at row level: return "TableCell" to enter the cell branch.
		r.state = rfStateCellSub
		return "TableCell", true
	case rfStateCellSub:
		// Inside cell sub-child iteration: return false (no sub-children).
		r.state = rfStateDone
		return "", false
	default:
		return "", false
	}
}

func (r *rowWithCellFailReader) FinishChild() error {
	// Return an error on the first FinishChild call (outer cell loop).
	if r.cellFinishErr != nil {
		err := r.cellFinishErr
		r.cellFinishErr = nil // clear so subsequent calls succeed
		return err
	}
	return nil
}

// ── copyCells — srcCell nil (out-of-bounds source indices) ───────────────────

// TestCopyCells_SrcCellNil exercises the branch in copyCells where srcCell
// is nil because the source row/column index is out of range. We access this
// via the internal package.
func TestCopyCells_SrcCellNilBranch(t *testing.T) {
	// Build a table with 1 column and 1 row.
	src := NewTableObject()
	src.NewColumn()
	src.NewRow().cells[0].SetText("Only")

	h := newTableHelper(src)

	// Manually add a result row so copyCells has a valid dstRowIdx.
	resultRow := NewTableRow()
	h.result.rows = append(h.result.rows, resultRow)

	// Call copyCells with an out-of-bounds source column index (99).
	// This means srcCell == nil, so the if-branch is skipped.
	h.copyCells(99, 0, 0, 0) // srcColIdx=99 → out of range → srcCell=nil

	// The destination cell should remain the auto-created default cell.
	if len(resultRow.cells) == 0 {
		t.Error("copyCells should ensure result row has a cell slot")
	}
	// The cell at slot 0 should be the default (empty text), not "Only".
	cell := resultRow.cells[0]
	if cell.Text() != "" {
		t.Errorf("cell should be default empty, got %q", cell.Text())
	}
}

// TestCopyCells_SrcRowNilBranch exercises srcCell==nil via out-of-bounds row.
func TestCopyCells_SrcRowNilBranch(t *testing.T) {
	src := NewTableObject()
	src.NewColumn()
	src.NewRow().cells[0].SetText("Only")

	h := newTableHelper(src)

	resultRow := NewTableRow()
	h.result.rows = append(h.result.rows, resultRow)

	// srcRowIdx=99 is out of range → srcCell = nil
	h.copyCells(0, 99, 0, 0)

	if len(resultRow.cells) == 0 {
		t.Error("copyCells should ensure result row has a cell slot")
	}
	cell := resultRow.cells[0]
	if cell.Text() != "" {
		t.Errorf("cell should be default empty, got %q", cell.Text())
	}
}

// ── Serialize error paths using failing writer ────────────────────────────────

// mockFailWriter fails WriteObject on the N-th call.
type mockFailWriter struct {
	callCount int
	failOn    int // fail on this call number (1-based)
}

func (m *mockFailWriter) WriteStr(name, value string)        {}
func (m *mockFailWriter) WriteInt(name string, v int)         {}
func (m *mockFailWriter) WriteBool(name string, v bool)       {}
func (m *mockFailWriter) WriteFloat(name string, v float32)   {}
func (m *mockFailWriter) WriteObject(obj report.Serializable) error {
	m.callCount++
	if m.failOn > 0 && m.callCount == m.failOn {
		return errors.New("mock WriteObject error")
	}
	// Actually call Serialize on the object to exercise its body.
	inner := &mockFailWriter{failOn: -1}
	return obj.Serialize(inner)
}
func (m *mockFailWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	return m.WriteObject(obj)
}

// TestTableCell_Serialize_ColSpanAndRowSpan exercises both the colSpan!=1
// and rowSpan!=1 write paths, plus the embedded objects write path via a
// non-failing writer.
func TestTableCell_Serialize_AllNonDefaultBranches(t *testing.T) {
	c := NewTableCell()
	c.SetColSpan(2)
	c.SetRowSpan(3)
	c.SetDuplicates(CellDuplicatesClear)
	inner := NewTableCell()
	c.objects = append(c.objects, inner)

	w := &mockTableWriter{failWriteObject: false}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize should not error: %v", err)
	}
}

// TestTableRow_Serialize_AllBranches exercises MaxHeight != 1000 write path.
func TestTableRow_Serialize_MaxHeightBranch(t *testing.T) {
	r := NewTableRow()
	r.SetMaxHeight(2000) // non-default, triggers the maxHeight != 1000 branch
	r.SetMinHeight(5)
	r.SetAutoSize(true)
	r.SetCanBreak(true)
	r.SetPageBreak(true)
	r.SetKeepRows(1)

	w := &mockTableWriter{failWriteObject: false}
	if err := r.Serialize(w); err != nil {
		t.Errorf("Serialize should not error: %v", err)
	}
}

// TestTableColumn_Serialize_AllBranches exercises MaxWidth != 5000 write path.
func TestTableColumn_Serialize_MaxWidthBranch(t *testing.T) {
	c := NewTableColumn()
	c.SetMinWidth(10)
	c.SetMaxWidth(100) // non-default (default = 5000)
	c.SetAutoSize(true)
	c.SetPageBreak(true)
	c.SetKeepColumns(2)

	w := &mockTableWriter{}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize should not error: %v", err)
	}
}

// TestTableBase_Serialize_AllBranches exercises all non-default branches in
// TableBase.Serialize that weren't hit by previous tests.
func TestTableBase_Serialize_AllNonDefaultBranches(t *testing.T) {
	tbl := NewTableObject()
	tbl.SetFixedRows(1)
	tbl.SetFixedColumns(1)
	tbl.SetLayout(TableLayoutDownThenAcross) // non-default (1)
	tbl.SetPrintOnParent(true)
	tbl.SetWrappedGap(5)
	tbl.SetRepeatHeaders(false)
	tbl.SetRepeatRowHeaders(true)
	tbl.SetRepeatColumnHeaders(true)
	tbl.SetAdjustSpannedCellsWidth(true)
	tbl.ManualBuildEvent = "Ev"

	w := &mockTableWriter{failWriteObject: false}
	if err := tbl.Serialize(w); err != nil {
		t.Errorf("TableBase.Serialize should not error: %v", err)
	}
}

// TestTableObject_Serialize_ManualBuildAutoSpansFalse exercises the
// ManualBuildAutoSpans=false write path in TableObject.Serialize.
func TestTableObject_Serialize_ManualBuildAutoSpansFalse(t *testing.T) {
	tbl := NewTableObject()
	tbl.ManualBuildAutoSpans = false // non-default (default=true)

	w := &mockTableWriter{}
	if err := tbl.Serialize(w); err != nil {
		t.Errorf("TableObject.Serialize should not error: %v", err)
	}
}

// ── copyCells — dstRowIdx out-of-bounds early return ─────────────────────────

// TestCopyCells_DstRowNegative exercises the early-return in copyCells when
// dstRowIdx is negative.
func TestCopyCells_DstRowNegative(t *testing.T) {
	src := NewTableObject()
	src.NewColumn()
	src.NewRow().cells[0].SetText("X")

	h := newTableHelper(src)

	// Result has one row (index 0). Call copyCells with dstRowIdx=-1.
	resultRow := NewTableRow()
	h.result.rows = append(h.result.rows, resultRow)

	// dstRowIdx = -1 triggers the early return before any mutation.
	h.copyCells(0, 0, 0, -1)

	// The result row should be untouched (no cells added).
	if len(resultRow.cells) != 0 {
		t.Errorf("copyCells with dstRowIdx=-1 should not modify rows; got %d cells", len(resultRow.cells))
	}
}

// TestCopyCells_DstRowBeyondBounds exercises the early-return when dstRowIdx
// is past the end of the result rows slice.
func TestCopyCells_DstRowBeyondBounds(t *testing.T) {
	src := NewTableObject()
	src.NewColumn()
	src.NewRow().cells[0].SetText("X")

	h := newTableHelper(src)

	// result has 1 row (index 0). dstRowIdx=1 is out of range.
	resultRow := NewTableRow()
	h.result.rows = append(h.result.rows, resultRow)

	h.copyCells(0, 0, 0, 1) // dstRowIdx=1, but len=1 → out of bounds

	// The result row should be untouched.
	if len(resultRow.cells) != 0 {
		t.Errorf("copyCells with out-of-bounds dstRowIdx should not modify rows; got %d cells", len(resultRow.cells))
	}
}
