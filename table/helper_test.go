package table_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/table"
)

// buildTemplate creates a 3-row × 2-column template table with labelled cells.
func buildTemplate() *table.TableObject {
	tbl := table.NewTableObject()
	tbl.NewColumn()
	tbl.NewColumn()
	row0 := tbl.NewRow()
	row0.Cell(0).SetText("H1")
	row0.Cell(1).SetText("H2")
	row1 := tbl.NewRow()
	row1.Cell(0).SetText("D1")
	row1.Cell(1).SetText("D2")
	row2 := tbl.NewRow()
	row2.Cell(0).SetText("F1")
	row2.Cell(1).SetText("F2")
	return tbl
}

func TestIsManualBuild_False(t *testing.T) {
	tbl := table.NewTableObject()
	if tbl.IsManualBuild() {
		t.Error("IsManualBuild should be false when no callback or event is set")
	}
}

func TestIsManualBuild_True(t *testing.T) {
	tbl := table.NewTableObject()
	tbl.ManualBuild = func(h *table.TableHelper) {}
	if !tbl.IsManualBuild() {
		t.Error("IsManualBuild should be true when ManualBuild callback is set")
	}
}

func TestInvokeManualBuild_NilCallback(t *testing.T) {
	tbl := buildTemplate()
	if r := tbl.InvokeManualBuild(); r != nil {
		t.Error("InvokeManualBuild should return nil when no callback is set")
	}
}

func TestTableHelper_PrintRows(t *testing.T) {
	tbl := buildTemplate()
	var result *table.TableBase
	tbl.ManualBuild = func(h *table.TableHelper) {
		h.PrintRows()
		result = h.Result()
	}
	tbl.InvokeManualBuild()

	if result == nil {
		t.Fatal("Result should not be nil")
	}
	if result.RowCount() != 3 {
		t.Errorf("RowCount: got %d, want 3", result.RowCount())
	}
	if result.ColumnCount() != 2 {
		t.Errorf("ColumnCount: got %d, want 2", result.ColumnCount())
	}
	// Cells should have same text as template.
	checkCell := func(row, col int, want string) {
		t.Helper()
		c := result.Cell(row, col)
		if c == nil {
			t.Errorf("Cell(%d,%d) is nil", row, col)
			return
		}
		if c.Text() != want {
			t.Errorf("Cell(%d,%d).Text = %q, want %q", row, col, c.Text(), want)
		}
	}
	checkCell(0, 0, "H1")
	checkCell(0, 1, "H2")
	checkCell(1, 0, "D1")
	checkCell(1, 1, "D2")
	checkCell(2, 0, "F1")
	checkCell(2, 1, "F2")
}

func TestTableHelper_SelectiveRows(t *testing.T) {
	tbl := buildTemplate()
	var result *table.TableBase
	tbl.ManualBuild = func(h *table.TableHelper) {
		// Print header (row 0) and footer (row 2) but skip data row.
		h.PrintRow(0)
		h.PrintColumns()
		h.PrintRow(2)
		h.PrintColumns()
		result = h.Result()
	}
	tbl.InvokeManualBuild()

	if result == nil {
		t.Fatal("Result should not be nil")
	}
	if result.RowCount() != 2 {
		t.Errorf("RowCount: got %d, want 2", result.RowCount())
	}
	c := result.Cell(0, 0)
	if c == nil || c.Text() != "H1" {
		t.Errorf("Cell(0,0).Text = %q, want H1", c.Text())
	}
	c = result.Cell(1, 0)
	if c == nil || c.Text() != "F1" {
		t.Errorf("Cell(1,0).Text = %q, want F1", c.Text())
	}
}

func TestTableHelper_RepeatDataRow(t *testing.T) {
	tbl := buildTemplate()
	dataCount := 3
	var result *table.TableBase
	tbl.ManualBuild = func(h *table.TableHelper) {
		h.PrintRow(0) // header
		h.PrintColumns()
		for i := 0; i < dataCount; i++ {
			h.PrintRow(1) // data row repeated
			h.PrintColumns()
		}
		result = h.Result()
	}
	tbl.InvokeManualBuild()

	if result == nil {
		t.Fatal("Result should not be nil")
	}
	wantRows := 1 + dataCount // header + 3 data
	if result.RowCount() != wantRows {
		t.Errorf("RowCount: got %d, want %d", result.RowCount(), wantRows)
	}
}

// buildColumnTemplate creates a 2-row × 3-column template:
//   Col 0 (header): "Label1", "Label2"
//   Col 1 (data):   "[DS.F1]", "[DS.F2]"
//   Col 2 (footer): "Total1", "Total2"
func buildColumnTemplate() *table.TableObject {
	tbl := table.NewTableObject()
	tbl.SetFixedColumns(1)
	tbl.NewColumn() // col 0: header
	tbl.NewColumn() // col 1: data
	tbl.NewColumn() // col 2: footer
	row0 := tbl.NewRow()
	row0.Cell(0).SetText("Label1")
	row0.Cell(1).SetText("[DS.F1]")
	row0.Cell(2).SetText("Total1")
	row1 := tbl.NewRow()
	row1.Cell(0).SetText("Label2")
	row1.Cell(1).SetText("[DS.F2]")
	row1.Cell(2).SetText("Total2")
	return tbl
}

// TestTableHelper_ColumnFirstPrintRows verifies that PrintRows in column-first
// mode (after PrintColumn) only fills rows for the current column and does not
// add extra columns (matches C# TableObject.PrintRows() semantics).
func TestTableHelper_ColumnFirstPrintRows(t *testing.T) {
	tbl := buildColumnTemplate()
	h := table.NewTableHelper(tbl)

	// Column-first: header col, then 2 data cols, then footer col.
	h.PrintColumn(0)
	h.PrintRows() // fill 2 rows for col 0

	h.PrintColumn(1)
	h.PrintRows() // fill 2 rows for col 1 (data instance 1)

	h.PrintColumn(1)
	h.PrintRows() // fill 2 rows for col 1 (data instance 2)

	h.PrintColumn(2)
	h.PrintRows() // fill 2 rows for col 2 (footer)

	result := h.Result()
	if result.RowCount() != 2 {
		t.Fatalf("RowCount: got %d, want 2", result.RowCount())
	}
	// Header col + 2 data cols + footer col = 4 columns.
	if result.ColumnCount() != 4 {
		t.Fatalf("ColumnCount: got %d, want 4", result.ColumnCount())
	}
	// Header cells come from template col 0.
	c := result.Cell(0, 0)
	if c == nil || c.Text() != "Label1" {
		t.Errorf("Cell(0,0) = %q, want Label1", c.Text())
	}
	// Data col 1 (result col 1) comes from template col 1.
	c = result.Cell(0, 1)
	if c == nil || c.Text() != "[DS.F1]" {
		t.Errorf("Cell(0,1) = %q, want [DS.F1]", c.Text())
	}
	// Data col 2 (result col 2) also from template col 1.
	c = result.Cell(0, 2)
	if c == nil || c.Text() != "[DS.F1]" {
		t.Errorf("Cell(0,2) = %q, want [DS.F1]", c.Text())
	}
	// Footer col (result col 3) from template col 2.
	c = result.Cell(0, 3)
	if c == nil || c.Text() != "Total1" {
		t.Errorf("Cell(0,3) = %q, want Total1", c.Text())
	}
}

// TestTableHelper_CellTextEval verifies that CellTextEval is applied to cells
// during column-first PrintRows, enabling per-row expression evaluation.
func TestTableHelper_CellTextEval(t *testing.T) {
	tbl := buildColumnTemplate()
	h := table.NewTableHelper(tbl)

	dataRows := []string{"Alice", "Bob"}

	// Header column.
	h.PrintColumn(0)
	h.PrintRows()

	// Data columns with per-row evaluation.
	for _, name := range dataRows {
		captured := name
		h.CellTextEval = func(text string) string {
			if text == "[DS.F1]" {
				return captured
			}
			return text
		}
		h.PrintColumn(1)
		h.PrintRows()
		h.CellTextEval = nil
	}

	// Footer column.
	h.PrintColumn(2)
	h.PrintRows()

	result := h.Result()
	if result.RowCount() != 2 {
		t.Fatalf("RowCount: got %d, want 2", result.RowCount())
	}
	// 1 header + 2 data + 1 footer = 4 columns
	if result.ColumnCount() != 4 {
		t.Fatalf("ColumnCount: got %d, want 4", result.ColumnCount())
	}
	// First data column (col 1) should have "Alice".
	c := result.Cell(0, 1)
	if c == nil || c.Text() != "Alice" {
		t.Errorf("Cell(0,1) = %q, want Alice", c.Text())
	}
	// Second data column (col 2) should have "Bob".
	c = result.Cell(0, 2)
	if c == nil || c.Text() != "Bob" {
		t.Errorf("Cell(0,2) = %q, want Bob", c.Text())
	}
}

// TestTableHelper_NewTableHelper_IsExported ensures the exported constructor works.
func TestTableHelper_NewTableHelper_IsExported(t *testing.T) {
	tbl := buildTemplate()
	h := table.NewTableHelper(tbl)
	if h == nil {
		t.Fatal("NewTableHelper returned nil")
	}
	if h.Result() == nil {
		t.Fatal("Result() returned nil on fresh helper")
	}
}
