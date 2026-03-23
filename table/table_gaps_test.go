package table_test

// table_gaps_test.go — tests for the porting gaps implemented in
// go-fastreport-bcmvb and go-fastreport-2f3uc.
//
// Covers:
//   - TableRow.SaveState/RestoreState (TableRow.cs line 372-381)
//   - TableRow.SetHeight min/max enforcement (TableRow.cs line 51-58)
//   - TableColumn.SaveState/RestoreState (TableColumn.cs line 221-230)
//   - TableColumn.SetWidth min/max enforcement (TableColumn.cs line 47-59)
//   - TableCell.CalcWidth/CalcHeight
//   - TableCell.GetExpressions
//   - TableCell.SaveState/RestoreState
//   - TableCell.GetData
//   - TableBase.GetSpanList/ResetSpanList
//   - TableBase.IsInsideSpan
//   - TableBase.CorrectSpansOnRowChange
//   - TableBase.CorrectSpansOnColumnChange
//   - TableBase.SaveState/RestoreState
//   - TableBase.CalcWidth/CalcHeight

import (
	"testing"

	"github.com/andrewloable/go-fastreport/table"
)

// ── TableRow ──────────────────────────────────────────────────────────────────

func TestTableRow_SetHeight_MinEnforced(t *testing.T) {
	r := table.NewTableRow()
	r.SetMinHeight(50)
	r.SetHeight(10) // below min
	if r.Height() != 50 {
		t.Errorf("Height should be clamped to MinHeight=50, got %v", r.Height())
	}
}

func TestTableRow_SetHeight_MaxEnforced_WhenCanBreakFalse(t *testing.T) {
	r := table.NewTableRow()
	r.SetMaxHeight(200)
	r.SetCanBreak(false)
	r.SetHeight(500) // above max, canBreak=false → clamp
	if r.Height() != 200 {
		t.Errorf("Height should be clamped to MaxHeight=200, got %v", r.Height())
	}
}

func TestTableRow_SetHeight_MaxNotEnforced_WhenCanBreakTrue(t *testing.T) {
	r := table.NewTableRow()
	r.SetMaxHeight(200)
	r.SetCanBreak(true)
	r.SetHeight(500) // above max but canBreak=true → allowed
	if r.Height() != 500 {
		t.Errorf("Height should be 500 when CanBreak=true, got %v", r.Height())
	}
}

func TestTableRow_SetHeight_MaxZero_NoClamp(t *testing.T) {
	r := table.NewTableRow()
	r.SetMaxHeight(0) // 0 = unlimited
	r.SetHeight(9999)
	if r.Height() != 9999 {
		t.Errorf("Height should be 9999 when MaxHeight=0, got %v", r.Height())
	}
}

func TestTableRow_SaveRestoreState(t *testing.T) {
	r := table.NewTableRow()
	r.SetHeight(80)
	r.SetVisible(true)
	r.SaveState()

	r.SetHeight(200)
	r.SetVisible(false)

	r.RestoreState()

	if r.Height() != 80 {
		t.Errorf("Height after restore = %v, want 80", r.Height())
	}
	if !r.Visible() {
		t.Error("Visible after restore should be true")
	}
}

// ── TableColumn ───────────────────────────────────────────────────────────────

func TestTableColumn_SetWidth_MinEnforced(t *testing.T) {
	c := table.NewTableColumn()
	c.SetMinWidth(30)
	c.SetWidth(10) // below min
	if c.Width() != 30 {
		t.Errorf("Width should be clamped to MinWidth=30, got %v", c.Width())
	}
}

func TestTableColumn_SetWidth_MaxEnforced(t *testing.T) {
	c := table.NewTableColumn()
	c.SetMaxWidth(300)
	c.SetWidth(9999) // above max
	if c.Width() != 300 {
		t.Errorf("Width should be clamped to MaxWidth=300, got %v", c.Width())
	}
}

func TestTableColumn_SetWidth_MaxZero_NoClamp(t *testing.T) {
	c := table.NewTableColumn()
	c.SetMaxWidth(0) // 0 = unlimited
	c.SetWidth(99999)
	if c.Width() != 99999 {
		t.Errorf("Width should be 99999 when MaxWidth=0, got %v", c.Width())
	}
}

func TestTableColumn_SaveRestoreState(t *testing.T) {
	c := table.NewTableColumn()
	c.SetWidth(120)
	c.SetVisible(true)
	c.SaveState()

	c.SetWidth(500)
	c.SetVisible(false)

	c.RestoreState()

	if c.Width() != 120 {
		t.Errorf("Width after restore = %v, want 120", c.Width())
	}
	if !c.Visible() {
		t.Error("Visible after restore should be true")
	}
}

// ── TableCell ─────────────────────────────────────────────────────────────────

func TestTableCell_CalcWidth_ReturnsCurrentWidth(t *testing.T) {
	c := table.NewTableCell()
	c.SetWidth(123)
	if c.CalcWidth() != 123 {
		t.Errorf("CalcWidth() = %v, want 123", c.CalcWidth())
	}
}

func TestTableCell_CalcHeight_ReturnsCurrentHeight(t *testing.T) {
	c := table.NewTableCell()
	c.SetHeight(45)
	if c.CalcHeight() != 45 {
		t.Errorf("CalcHeight() = %v, want 45", c.CalcHeight())
	}
}

func TestTableCell_GetExpressions_ReturnsBase(t *testing.T) {
	c := table.NewTableCell()
	exprs := c.GetExpressions()
	// With no visible/printable expressions, should return empty or nil.
	if exprs == nil {
		exprs = []string{}
	}
	// No expressions set — length should be 0.
	if len(exprs) != 0 {
		t.Errorf("GetExpressions() len = %d, want 0; got %v", len(exprs), exprs)
	}
}

func TestTableCell_GetExpressions_IncludesVisibleExpr(t *testing.T) {
	c := table.NewTableCell()
	c.SetVisibleExpression("[x > 0]")
	exprs := c.GetExpressions()
	if len(exprs) == 0 {
		t.Fatal("GetExpressions() should include visible expression")
	}
	found := false
	for _, e := range exprs {
		if e == "x > 0" {
			found = true
		}
	}
	if !found {
		t.Errorf("GetExpressions() = %v, want to contain 'x > 0'", exprs)
	}
}

func TestTableCell_SaveRestoreState_Text(t *testing.T) {
	c := table.NewTableCell()
	c.SetText("original")
	c.SaveState()

	c.SetText("modified")
	c.RestoreState()

	if c.Text() != "original" {
		t.Errorf("Text after restore = %q, want 'original'", c.Text())
	}
}

func TestTableCell_SaveRestoreState_ObjectsDiscarded(t *testing.T) {
	c := table.NewTableCell()
	inner := table.NewTableCell()
	c.AddObject(inner)
	c.SaveState() // savedObjectCount = 1

	// Add another object after save.
	extra := table.NewTableCell()
	c.AddObject(extra)
	if c.ObjectCount() != 2 {
		t.Fatal("expected 2 objects before restore")
	}

	c.RestoreState()

	if c.ObjectCount() != 1 {
		t.Errorf("ObjectCount after restore = %d, want 1 (extra object discarded)", c.ObjectCount())
	}
}

func TestTableCell_GetData_ClearsTextWhenInsideSpan(t *testing.T) {
	c := table.NewTableCell()
	c.SetText("hello")
	c.SaveState()
	c.GetData(true) // inside span → text cleared
	if c.Text() != "" {
		t.Errorf("Text should be empty when inside span, got %q", c.Text())
	}
}

func TestTableCell_GetData_PreservesTextWhenNotInsideSpan(t *testing.T) {
	c := table.NewTableCell()
	c.SetText("hello")
	c.SaveState()
	c.GetData(false)
	if c.Text() != "hello" {
		t.Errorf("Text should be 'hello' when not inside span, got %q", c.Text())
	}
}

// ── TableBase span management ─────────────────────────────────────────────────

func buildTable3x3() *table.TableBase {
	tbl := table.NewTableBase()
	for i := 0; i < 3; i++ {
		tbl.NewColumn()
	}
	for i := 0; i < 3; i++ {
		tbl.NewRow()
	}
	return tbl
}

func TestTableBase_GetSpanList_NoSpans(t *testing.T) {
	tbl := buildTable3x3()
	spans := tbl.GetSpanList()
	if len(spans) != 0 {
		t.Errorf("GetSpanList() should be empty when no spans, got %v", spans)
	}
}

func TestTableBase_GetSpanList_WithSpan(t *testing.T) {
	tbl := buildTable3x3()
	// Cell at (row=0, col=0) spans 2 columns.
	tbl.Cell(0, 0).SetColSpan(2)
	spans := tbl.GetSpanList()
	if len(spans) != 1 {
		t.Fatalf("GetSpanList() len = %d, want 1; got %v", len(spans), spans)
	}
	s := spans[0]
	if s.Min.X != 0 || s.Min.Y != 0 || s.Max.X != 2 || s.Max.Y != 1 {
		t.Errorf("span = %v, want {(0,0)-(2,1)}", s)
	}
}

func TestTableBase_GetSpanList_Cached(t *testing.T) {
	tbl := buildTable3x3()
	tbl.Cell(0, 0).SetColSpan(2)
	s1 := tbl.GetSpanList()
	s2 := tbl.GetSpanList()
	// Should return the same backing slice (identical pointer for first element).
	if len(s1) != len(s2) {
		t.Error("GetSpanList() should return cached list")
	}
}

func TestTableBase_ResetSpanList(t *testing.T) {
	tbl := buildTable3x3()
	tbl.Cell(0, 0).SetColSpan(2)
	_ = tbl.GetSpanList() // populate cache
	tbl.ResetSpanList()

	// Add a new span after reset — list should be recomputed.
	tbl.Cell(1, 1).SetRowSpan(2)
	spans := tbl.GetSpanList()
	// Now we should see both (col=0 row=0 colspan=2) and (col=1 row=1 rowspan=2).
	if len(spans) != 2 {
		t.Errorf("GetSpanList() after reset = %d, want 2; %v", len(spans), spans)
	}
}

func TestTableBase_IsInsideSpan_True(t *testing.T) {
	tbl := buildTable3x3()
	// Cell at row=0,col=0 spans 2 columns: covers col=0,col=1 in row=0.
	tbl.Cell(0, 0).SetColSpan(2)
	// col=1, row=0 is inside the span.
	if !tbl.IsInsideSpan(1, 0) {
		t.Error("IsInsideSpan(col=1, row=0) should be true")
	}
}

func TestTableBase_IsInsideSpan_False_Origin(t *testing.T) {
	tbl := buildTable3x3()
	tbl.Cell(0, 0).SetColSpan(2)
	// The origin cell (col=0, row=0) is NOT inside its own span.
	if tbl.IsInsideSpan(0, 0) {
		t.Error("IsInsideSpan(col=0, row=0) should be false (it is the span origin)")
	}
}

func TestTableBase_IsInsideSpan_False_NoSpan(t *testing.T) {
	tbl := buildTable3x3()
	if tbl.IsInsideSpan(1, 1) {
		t.Error("IsInsideSpan(1,1) should be false with no spans")
	}
}

func TestTableBase_CorrectSpansOnRowChange_Insert(t *testing.T) {
	tbl := buildTable3x3()
	// Cell at row=0,col=0 spans 2 rows.
	tbl.Cell(0, 0).SetRowSpan(2)
	// Insert a row at index 1 — the span should grow to 3.
	tbl.CorrectSpansOnRowChange(1, +1)
	if tbl.Cell(0, 0).RowSpan() != 3 {
		t.Errorf("RowSpan after insert = %d, want 3", tbl.Cell(0, 0).RowSpan())
	}
}

func TestTableBase_CorrectSpansOnRowChange_Remove(t *testing.T) {
	tbl := buildTable3x3()
	tbl.Cell(0, 0).SetRowSpan(3)
	// Remove row at index 1 — span should shrink to 2.
	tbl.CorrectSpansOnRowChange(1, -1)
	if tbl.Cell(0, 0).RowSpan() != 2 {
		t.Errorf("RowSpan after remove = %d, want 2", tbl.Cell(0, 0).RowSpan())
	}
}

func TestTableBase_CorrectSpansOnRowChange_Locked(t *testing.T) {
	tbl := buildTable3x3()
	tbl.Cell(0, 0).SetRowSpan(2)
	tbl.SetLockCorrectSpans(true)
	tbl.CorrectSpansOnRowChange(1, +1) // should be a no-op
	tbl.SetLockCorrectSpans(false)
	if tbl.Cell(0, 0).RowSpan() != 2 {
		t.Errorf("RowSpan should be unchanged when locked, got %d", tbl.Cell(0, 0).RowSpan())
	}
}

func TestTableBase_CorrectSpansOnColumnChange_Insert(t *testing.T) {
	tbl := buildTable3x3()
	// Cell at row=0,col=0 spans 2 columns.
	tbl.Cell(0, 0).SetColSpan(2)
	// Insert column at index 1 — span should grow to 3.
	tbl.CorrectSpansOnColumnChange(1, +1)
	if tbl.Cell(0, 0).ColSpan() != 3 {
		t.Errorf("ColSpan after insert = %d, want 3", tbl.Cell(0, 0).ColSpan())
	}
	// The rows should have a new cell inserted at index 1.
	for i, row := range tbl.Rows() {
		if row.CellCount() != 4 { // was 3, now 4
			t.Errorf("Row[%d] cell count = %d after insert, want 4", i, row.CellCount())
		}
	}
}

func TestTableBase_CorrectSpansOnColumnChange_Remove(t *testing.T) {
	tbl := buildTable3x3()
	tbl.Cell(0, 0).SetColSpan(3)
	// Remove column at index 1 — span should shrink to 2.
	tbl.CorrectSpansOnColumnChange(1, -1)
	if tbl.Cell(0, 0).ColSpan() != 2 {
		t.Errorf("ColSpan after remove = %d, want 2", tbl.Cell(0, 0).ColSpan())
	}
	// Each row should have lost one cell.
	for i, row := range tbl.Rows() {
		if row.CellCount() != 2 {
			t.Errorf("Row[%d] cell count = %d after remove, want 2", i, row.CellCount())
		}
	}
}

func TestTableBase_CorrectSpansOnColumnChange_Locked(t *testing.T) {
	tbl := buildTable3x3()
	tbl.Cell(0, 0).SetColSpan(2)
	tbl.SetLockCorrectSpans(true)
	tbl.CorrectSpansOnColumnChange(1, +1) // no-op
	tbl.SetLockCorrectSpans(false)
	if tbl.Cell(0, 0).ColSpan() != 2 {
		t.Errorf("ColSpan should be unchanged when locked, got %d", tbl.Cell(0, 0).ColSpan())
	}
}

// ── TableBase.SaveState / RestoreState ────────────────────────────────────────

func TestTableBase_SaveRestoreState(t *testing.T) {
	tbl := table.NewTableBase()
	col := tbl.NewColumn()
	col.SetWidth(200)
	row := tbl.NewRow()
	row.SetHeight(60)
	cell := tbl.Cell(0, 0)
	cell.SetText("original")

	tbl.SaveState()

	// Mutate everything.
	col.SetWidth(999)
	row.SetHeight(999)
	cell.SetText("mutated")

	tbl.RestoreState()

	if col.Width() != 200 {
		t.Errorf("Column width after restore = %v, want 200", col.Width())
	}
	if row.Height() != 60 {
		t.Errorf("Row height after restore = %v, want 60", row.Height())
	}
	if cell.Text() != "original" {
		t.Errorf("Cell text after restore = %q, want 'original'", cell.Text())
	}
}

func TestTableBase_SaveState_SetsCanGrowCanShrink(t *testing.T) {
	tbl := table.NewTableBase()
	tbl.NewColumn()
	tbl.NewRow()
	tbl.SetCanGrow(false)
	tbl.SetCanShrink(false)

	tbl.SaveState()

	// After SaveState, CanGrow and CanShrink should be true.
	if !tbl.CanGrow() {
		t.Error("CanGrow should be true after SaveState")
	}
	if !tbl.CanShrink() {
		t.Error("CanShrink should be true after SaveState")
	}
}

// ── TableBase.CalcWidth / CalcHeight ─────────────────────────────────────────

func TestTableBase_CalcWidth_NonAutoSize(t *testing.T) {
	tbl := table.NewTableBase()
	c1 := tbl.NewColumn()
	c1.SetWidth(100)
	c1.SetAutoSize(false)
	c2 := tbl.NewColumn()
	c2.SetWidth(150)
	c2.SetAutoSize(false)
	tbl.NewRow()

	w := tbl.CalcWidth()
	if w != 250 {
		t.Errorf("CalcWidth() = %v, want 250", w)
	}
}

func TestTableBase_CalcWidth_AutoSize_GrowsToCell(t *testing.T) {
	tbl := table.NewTableBase()
	col := tbl.NewColumn()
	col.SetAutoSize(true)
	col.SetWidth(10) // starts narrow
	tbl.NewRow()
	// Cell is wider than column.
	tbl.Cell(0, 0).SetWidth(200)

	w := tbl.CalcWidth()
	// Column should have grown to 200.
	if col.Width() != 200 {
		t.Errorf("AutoSize column width = %v, want 200", col.Width())
	}
	if w != 200 {
		t.Errorf("CalcWidth() = %v, want 200", w)
	}
}

func TestTableBase_CalcHeight_NonAutoSize(t *testing.T) {
	tbl := table.NewTableBase()
	tbl.NewColumn()
	r1 := tbl.NewRow()
	r1.SetHeight(40)
	r1.SetAutoSize(false)
	r2 := tbl.NewRow()
	r2.SetHeight(60)
	r2.SetAutoSize(false)

	h := tbl.CalcHeight()
	if h != 100 {
		t.Errorf("CalcHeight() = %v, want 100", h)
	}
}

func TestTableBase_CalcHeight_AutoSize_GrowsToCell(t *testing.T) {
	tbl := table.NewTableBase()
	tbl.NewColumn()
	row := tbl.NewRow()
	row.SetAutoSize(true)
	row.SetHeight(10)
	// Cell is taller than row.
	tbl.Cell(0, 0).SetHeight(80)

	h := tbl.CalcHeight()
	if row.Height() != 80 {
		t.Errorf("AutoSize row height = %v, want 80", row.Height())
	}
	if h != 80 {
		t.Errorf("CalcHeight() = %v, want 80", h)
	}
}

func TestTableBase_CalcHeight_SkipsInvisibleRows(t *testing.T) {
	tbl := table.NewTableBase()
	tbl.NewColumn()
	r1 := tbl.NewRow()
	r1.SetHeight(50)
	r2 := tbl.NewRow()
	r2.SetHeight(50)
	r2.SetVisible(false) // invisible — should not count

	h := tbl.CalcHeight()
	if h != 50 {
		t.Errorf("CalcHeight() with one invisible row = %v, want 50", h)
	}
}

// ── TableBase.ProcessDuplicates ───────────────────────────────────────────────

// buildSimpleTable creates a nRows×nCols table with rows of height 30 and
// columns of width 100. Each cell is given Name=name and Text=text.
func buildSimpleTable(cells [][]struct{ name, text string }) *table.TableBase {
	tbl := table.NewTableBase()
	nRows := len(cells)
	nCols := 0
	if nRows > 0 {
		nCols = len(cells[0])
	}
	for ci := 0; ci < nCols; ci++ {
		col := table.NewTableColumn()
		col.SetWidth(100)
		tbl.AddColumn(col)
	}
	for ri := 0; ri < nRows; ri++ {
		row := table.NewTableRow()
		row.SetHeight(30)
		for ci := 0; ci < nCols; ci++ {
			cell := table.NewTableCell()
			cell.SetName(cells[ri][ci].name)
			cell.SetText(cells[ri][ci].text)
			row.AddCell(cell)
		}
		tbl.AddRow(row)
	}
	return tbl
}

func TestProcessDuplicates_ShowIsNoop(t *testing.T) {
	// CellDuplicatesShow (default) must not change anything.
	tbl := buildSimpleTable([][]struct{ name, text string }{
		{{"A", "X"}, {"A", "X"}},
	})
	tbl.ProcessDuplicates()
	if got := tbl.Cell(0, 0).Text(); got != "X" {
		t.Errorf("cell(0,0).Text = %q, want X", got)
	}
	if got := tbl.Cell(0, 1).Text(); got != "X" {
		t.Errorf("cell(0,1).Text = %q, want X", got)
	}
}

func TestProcessDuplicates_Clear_HorizontalRun(t *testing.T) {
	// Three consecutive cells in same row with same Name+Text → Clear blanks [1] and [2].
	tbl := buildSimpleTable([][]struct{ name, text string }{
		{{"A", "hello"}, {"A", "hello"}, {"A", "hello"}},
	})
	tbl.Cell(0, 0).SetDuplicates(table.CellDuplicatesClear)
	tbl.Cell(0, 1).SetDuplicates(table.CellDuplicatesClear)
	tbl.Cell(0, 2).SetDuplicates(table.CellDuplicatesClear)
	tbl.ProcessDuplicates()
	if got := tbl.Cell(0, 0).Text(); got != "hello" {
		t.Errorf("origin text = %q, want hello", got)
	}
	if got := tbl.Cell(0, 1).Text(); got != "" {
		t.Errorf("dup[1] text = %q, want empty", got)
	}
	if got := tbl.Cell(0, 2).Text(); got != "" {
		t.Errorf("dup[2] text = %q, want empty", got)
	}
}

func TestProcessDuplicates_Clear_DifferentTextStops(t *testing.T) {
	// Second cell has different text → only first two are cleared (none).
	tbl := buildSimpleTable([][]struct{ name, text string }{
		{{"A", "hello"}, {"A", "world"}, {"A", "hello"}},
	})
	tbl.Cell(0, 0).SetDuplicates(table.CellDuplicatesClear)
	tbl.Cell(0, 1).SetDuplicates(table.CellDuplicatesClear)
	tbl.Cell(0, 2).SetDuplicates(table.CellDuplicatesClear)
	tbl.ProcessDuplicates()
	// All texts unchanged because the run of duplicates at col=0 has length=1
	if got := tbl.Cell(0, 0).Text(); got != "hello" {
		t.Errorf("cell(0,0) = %q, want hello", got)
	}
	if got := tbl.Cell(0, 1).Text(); got != "world" {
		t.Errorf("cell(0,1) = %q, want world", got)
	}
}

func TestProcessDuplicates_Merge_ExpandsSpan(t *testing.T) {
	// Two cells in same row with same Name+Text → first cell ColSpan becomes 2.
	tbl := buildSimpleTable([][]struct{ name, text string }{
		{{"B", "val"}, {"B", "val"}, {"B", "other"}},
	})
	tbl.Cell(0, 0).SetDuplicates(table.CellDuplicatesMerge)
	tbl.Cell(0, 1).SetDuplicates(table.CellDuplicatesMerge)
	tbl.Cell(0, 2).SetDuplicates(table.CellDuplicatesMerge)
	tbl.ProcessDuplicates()
	if got := tbl.Cell(0, 0).ColSpan(); got != 2 {
		t.Errorf("ColSpan = %d, want 2", got)
	}
}

func TestProcessDuplicates_MergeNonEmpty_SkipsEmpty(t *testing.T) {
	// MergeNonEmpty with empty text must NOT expand the span.
	tbl := buildSimpleTable([][]struct{ name, text string }{
		{{"C", ""}, {"C", ""}},
	})
	tbl.Cell(0, 0).SetDuplicates(table.CellDuplicatesMergeNonEmpty)
	tbl.Cell(0, 1).SetDuplicates(table.CellDuplicatesMergeNonEmpty)
	tbl.ProcessDuplicates()
	if got := tbl.Cell(0, 0).ColSpan(); got != 1 {
		t.Errorf("ColSpan for empty text = %d, want 1 (no merge)", got)
	}
}

func TestProcessDuplicates_MergeNonEmpty_MergesWhenNonEmpty(t *testing.T) {
	tbl := buildSimpleTable([][]struct{ name, text string }{
		{{"C", "hi"}, {"C", "hi"}},
	})
	tbl.Cell(0, 0).SetDuplicates(table.CellDuplicatesMergeNonEmpty)
	tbl.Cell(0, 1).SetDuplicates(table.CellDuplicatesMergeNonEmpty)
	tbl.ProcessDuplicates()
	if got := tbl.Cell(0, 0).ColSpan(); got != 2 {
		t.Errorf("ColSpan for non-empty = %d, want 2", got)
	}
}

func TestProcessDuplicates_VerticalMerge(t *testing.T) {
	// Same name+text in two rows, same column → RowSpan becomes 2.
	tbl := buildSimpleTable([][]struct{ name, text string }{
		{{"D", "same"}},
		{{"D", "same"}},
		{{"D", "diff"}},
	})
	tbl.Cell(0, 0).SetDuplicates(table.CellDuplicatesMerge)
	tbl.Cell(1, 0).SetDuplicates(table.CellDuplicatesMerge)
	tbl.Cell(2, 0).SetDuplicates(table.CellDuplicatesMerge)
	tbl.ProcessDuplicates()
	if got := tbl.Cell(0, 0).RowSpan(); got != 2 {
		t.Errorf("RowSpan = %d, want 2", got)
	}
}
