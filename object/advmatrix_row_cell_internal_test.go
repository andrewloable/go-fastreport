package object

// advmatrix_row_cell_internal_test.go — internal tests to cover the remaining
// error branches in AdvMatrixRow.Serialize and AdvMatrixCell.Serialize that
// require a failing report.Writer.
//
// Coverage targets:
//   - AdvMatrixRow.Serialize   line 158: return err from WriteObjectNamed (cell loop)
//   - AdvMatrixCell.Serialize  line 185: return err from WriteObjectNamed (button loop)
//   - drainAdvChildren         line 480: FinishChild() != nil break
//   - readAdvDescriptor        line 492: FinishChild() != nil break

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// noopWriter is a report.Writer where all write methods are no-ops and
// WriteObjectNamed returns nil (success).  It is used as the base for the
// subtypes below.
type noopWriter struct{}

func (n *noopWriter) WriteStr(name, value string)       {}
func (n *noopWriter) WriteInt(name string, value int)   {}
func (n *noopWriter) WriteBool(name string, value bool) {}
func (n *noopWriter) WriteFloat(name string, value float32) {}
func (n *noopWriter) WriteObject(obj report.Serializable) error { return nil }
func (n *noopWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	return nil
}

// cellErrWriter returns an error on the first WriteObjectNamed call for "TableCell".
type cellErrWriter struct {
	noopWriter
	triggered bool
}

func (c *cellErrWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	if name == "TableCell" && !c.triggered {
		c.triggered = true
		return errors.New("cell write error")
	}
	return nil
}

// TestAdvMatrixRow_Serialize_WriteObjectNamedErrorOnCell verifies that
// AdvMatrixRow.Serialize returns the error from WriteObjectNamed when iterating
// the Cells slice (the WriteObjectNamed call at line 169 of advmatrix.go).
func TestAdvMatrixRow_Serialize_WriteObjectNamedErrorOnCell(t *testing.T) {
	row := &AdvMatrixRow{
		Name:   "R1",
		Height: 25,
		Cells:  []*AdvMatrixCell{{Name: "c1", Text: "hello"}},
	}

	w := &cellErrWriter{}
	err := row.Serialize(w)
	if err == nil {
		t.Fatal("expected error from AdvMatrixRow.Serialize when WriteObjectNamed fails for TableCell, got nil")
	}
	if err.Error() != "cell write error" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

// buttonErrWriter returns an error on the first WriteObjectNamed call for any
// button type name (MatrixCollapseButton / MatrixSortButton).
type buttonErrWriter struct {
	noopWriter
	triggered bool
}

func (b *buttonErrWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	if (name == "MatrixCollapseButton" || name == "MatrixSortButton") && !b.triggered {
		b.triggered = true
		return errors.New("button write error")
	}
	return nil
}

// TestAdvMatrixCell_Serialize_WriteObjectNamedErrorOnButton verifies that
// AdvMatrixCell.Serialize returns the error from WriteObjectNamed when iterating
// the Buttons slice (the WriteObjectNamed call inside the btn loop).
func TestAdvMatrixCell_Serialize_WriteObjectNamedErrorOnButton(t *testing.T) {
	cell := &AdvMatrixCell{
		Name: "c1",
		Buttons: []*MatrixButton{
			{TypeName: "MatrixCollapseButton", Name: "btn1"},
		},
	}

	w := &buttonErrWriter{}
	err := cell.Serialize(w)
	if err == nil {
		t.Fatal("expected error from AdvMatrixCell.Serialize when WriteObjectNamed fails for button, got nil")
	}
	if err.Error() != "button write error" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

// ── FinishChild error paths ───────────────────────────────────────────────────
// drainAdvChildren calls r.FinishChild() and breaks if it errors.
// readAdvDescriptor similarly calls r.FinishChild() and breaks if it errors.
//
// We use a mock report.Reader whose FinishChild returns an error after the
// first call to NextChild, simulating a broken XML reader.

// finishErrReader is a mock report.Reader.
// NextChild returns one fake child on the first call, then signals done.
// FinishChild always returns an error.
type finishErrReader struct {
	called int
}

func (r *finishErrReader) ReadStr(name, def string) string        { return def }
func (r *finishErrReader) ReadInt(name string, def int) int       { return def }
func (r *finishErrReader) ReadBool(name string, def bool) bool    { return def }
func (r *finishErrReader) ReadFloat(name string, def float32) float32 { return def }
func (r *finishErrReader) NextChild() (string, bool) {
	r.called++
	if r.called == 1 {
		return "SomeChild", true
	}
	return "", false
}
func (r *finishErrReader) FinishChild() error {
	return errors.New("finish child error")
}

// TestDrainAdvChildren_FinishChildError verifies that drainAdvChildren breaks
// out of its loop when FinishChild returns an error.
func TestDrainAdvChildren_FinishChildError(t *testing.T) {
	r := &finishErrReader{}
	// Should not panic; breaks on FinishChild error.
	drainAdvChildren(r)
	if r.called < 1 {
		t.Error("expected NextChild to be called at least once")
	}
}

// descriptorFinishErrReader is like finishErrReader but yields a "Descriptor"
// child type on the first call so that readAdvDescriptor recurses into it
// before hitting the FinishChild error.
type descriptorFinishErrReader struct {
	called int
}

func (r *descriptorFinishErrReader) ReadStr(name, def string) string        { return def }
func (r *descriptorFinishErrReader) ReadInt(name string, def int) int       { return def }
func (r *descriptorFinishErrReader) ReadBool(name string, def bool) bool    { return def }
func (r *descriptorFinishErrReader) ReadFloat(name string, def float32) float32 { return def }
func (r *descriptorFinishErrReader) NextChild() (string, bool) {
	r.called++
	if r.called == 1 {
		return "Descriptor", true
	}
	return "", false
}
func (r *descriptorFinishErrReader) FinishChild() error {
	return errors.New("finish child error")
}

// TestReadAdvDescriptor_FinishChildError verifies that readAdvDescriptor breaks
// out of its child-scanning loop when FinishChild returns an error (covering the
// `if r.FinishChild() != nil { break }` branch at the end of its for loop).
func TestReadAdvDescriptor_FinishChildError(t *testing.T) {
	r := &descriptorFinishErrReader{}
	d := readAdvDescriptor(r)
	if d == nil {
		t.Fatal("readAdvDescriptor returned nil")
	}
	// We yielded one "Descriptor" child, so d.Children should have exactly one entry.
	if len(d.Children) != 1 {
		t.Errorf("expected 1 child descriptor, got %d", len(d.Children))
	}
}

// ── DeserializeChild FinishChild error paths ─────────────────────────────────
// Lines 427 (button loop) and 434 (cell/row loop) in DeserializeChild both call
// `if r.FinishChild() != nil { break }`. We need a reader that returns:
//   - a "TableCell" child on the first NextChild call
//   - a "MatrixCollapseButton" child on the second NextChild call
//   - then an error from FinishChild on the first call
// to cover the button-level break (line 427).

// btnFinishErrReader simulates: TableRow → TableCell → MatrixCollapseButton,
// then FinishChild returns an error (triggers button-level break at line 427).
type btnFinishErrReader struct {
	nextCount    int
	finishCount  int
}

func (r *btnFinishErrReader) ReadStr(name, def string) string        { return def }
func (r *btnFinishErrReader) ReadInt(name string, def int) int       { return def }
func (r *btnFinishErrReader) ReadBool(name string, def bool) bool    { return def }
func (r *btnFinishErrReader) ReadFloat(name string, def float32) float32 { return def }
func (r *btnFinishErrReader) NextChild() (string, bool) {
	r.nextCount++
	switch r.nextCount {
	case 1:
		return "TableCell", true
	case 2:
		return "MatrixCollapseButton", true
	}
	return "", false
}
func (r *btnFinishErrReader) FinishChild() error {
	r.finishCount++
	return errors.New("finish error")
}

// TestDeserializeChild_TableRow_ButtonFinishChildError verifies that
// DeserializeChild for "TableRow" exits the button-scanning loop when
// FinishChild returns an error (covers the break at line 427).
func TestDeserializeChild_TableRow_ButtonFinishChildError(t *testing.T) {
	a := NewAdvMatrixObject()
	r := &btnFinishErrReader{}
	handled := a.DeserializeChild("TableRow", r)
	if !handled {
		t.Error("expected DeserializeChild to return true for TableRow")
	}
	if len(a.TableRows) != 1 {
		t.Errorf("expected 1 row, got %d", len(a.TableRows))
	}
	// The cell was added before the break; the button was appended inside the cell.
	if len(a.TableRows[0].Cells) != 1 {
		t.Errorf("expected 1 cell (added before FinishChild break), got %d", len(a.TableRows[0].Cells))
	}
}

// cellFinishErrReader simulates: TableRow → (first NextChild=TableCell), then
// FinishChild errors — triggering the outer cell-level break at line 434.
// The cell's inner button loop never fires because NextChild returns "" for the
// button level (no buttons).
type cellFinishErrReader struct {
	outerNext int
	innerNext int
	outerDone bool
}

func (r *cellFinishErrReader) ReadStr(name, def string) string        { return def }
func (r *cellFinishErrReader) ReadInt(name string, def int) int       { return def }
func (r *cellFinishErrReader) ReadBool(name string, def bool) bool    { return def }
func (r *cellFinishErrReader) ReadFloat(name string, def float32) float32 { return def }
func (r *cellFinishErrReader) NextChild() (string, bool) {
	if !r.outerDone {
		// First call at the outer (row) level → returns TableCell, then done.
		r.outerNext++
		if r.outerNext == 1 {
			return "TableCell", true
		}
		return "", false
	}
	// Inner (button) level — no buttons.
	r.innerNext++
	return "", false
}
func (r *cellFinishErrReader) FinishChild() error {
	// Mark outer as done so subsequent NextChild calls go to inner path.
	r.outerDone = true
	return errors.New("outer finish error")
}

// TestDeserializeChild_TableRow_CellFinishChildError verifies that
// DeserializeChild for "TableRow" exits the cell-scanning loop when
// FinishChild returns an error at the cell level (covers line 434).
func TestDeserializeChild_TableRow_CellFinishChildError(t *testing.T) {
	a := NewAdvMatrixObject()
	r := &cellFinishErrReader{}
	handled := a.DeserializeChild("TableRow", r)
	if !handled {
		t.Error("expected DeserializeChild to return true for TableRow")
	}
	if len(a.TableRows) != 1 {
		t.Errorf("expected 1 row, got %d", len(a.TableRows))
	}
}

// ── DeserializeChild "Columns" / "Rows" FinishChild error paths ──────────────
// Lines 451 and 467 in DeserializeChild both call
// `if r.FinishChild() != nil { break }` inside the Columns and Rows loops.

// columnsFinishErrReader yields a "Descriptor" child on the first NextChild
// call, then returns an error from FinishChild (triggering line 451).
type columnsFinishErrReader struct {
	called int
}

func (r *columnsFinishErrReader) ReadStr(name, def string) string        { return def }
func (r *columnsFinishErrReader) ReadInt(name string, def int) int       { return def }
func (r *columnsFinishErrReader) ReadBool(name string, def bool) bool    { return def }
func (r *columnsFinishErrReader) ReadFloat(name string, def float32) float32 { return def }
func (r *columnsFinishErrReader) NextChild() (string, bool) {
	r.called++
	if r.called == 1 {
		return "Descriptor", true
	}
	return "", false
}
func (r *columnsFinishErrReader) FinishChild() error {
	return errors.New("columns finish error")
}

// TestDeserializeChild_Columns_FinishChildError covers the FinishChild != nil
// break inside the "Columns" case loop (line 451).
func TestDeserializeChild_Columns_FinishChildError(t *testing.T) {
	a := NewAdvMatrixObject()
	r := &columnsFinishErrReader{}
	handled := a.DeserializeChild("Columns", r)
	if !handled {
		t.Error("expected DeserializeChild to return true for Columns")
	}
	if len(a.Columns) != 1 {
		t.Errorf("expected 1 column descriptor, got %d", len(a.Columns))
	}
}

// rowsFinishErrReader yields a "Descriptor" child on the first NextChild call,
// then errors from FinishChild (triggering line 467).
type rowsFinishErrReader struct {
	called int
}

func (r *rowsFinishErrReader) ReadStr(name, def string) string        { return def }
func (r *rowsFinishErrReader) ReadInt(name string, def int) int       { return def }
func (r *rowsFinishErrReader) ReadBool(name string, def bool) bool    { return def }
func (r *rowsFinishErrReader) ReadFloat(name string, def float32) float32 { return def }
func (r *rowsFinishErrReader) NextChild() (string, bool) {
	r.called++
	if r.called == 1 {
		return "Descriptor", true
	}
	return "", false
}
func (r *rowsFinishErrReader) FinishChild() error {
	return errors.New("rows finish error")
}

// TestDeserializeChild_Rows_FinishChildError covers the FinishChild != nil
// break inside the "Rows" case loop (line 467).
func TestDeserializeChild_Rows_FinishChildError(t *testing.T) {
	a := NewAdvMatrixObject()
	r := &rowsFinishErrReader{}
	handled := a.DeserializeChild("Rows", r)
	if !handled {
		t.Error("expected DeserializeChild to return true for Rows")
	}
	if len(a.Rows) != 1 {
		t.Errorf("expected 1 row descriptor, got %d", len(a.Rows))
	}
}

// ── formatBorderLinesStr: None case (BorderLinesNone = 0) ─────────────────────
// The outer condition in AdvMatrixCell.Serialize requires VisibleLines != 0
// before calling formatBorderLinesStr. Therefore the "None" case in the switch
// can only be reached if it is called directly.
// This test exercises it by calling the private function directly.

func TestFormatBorderLinesStr_NoneCase(t *testing.T) {
	result := formatBorderLinesStr(0) // BorderLinesNone = 0
	if result != "None" {
		t.Errorf("expected \"None\" for BorderLinesNone, got %q", result)
	}
}
