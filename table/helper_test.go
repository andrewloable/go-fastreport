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
