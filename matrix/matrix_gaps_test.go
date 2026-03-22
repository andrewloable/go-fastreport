package matrix_test

// matrix_gaps_test.go — tests for the porting gaps implemented in this iteration:
//
//  1. DescriptorExt / HeaderDescriptorExt fields (TemplateColumn/Row/Cell,
//     TemplateTotalColumn/Row/Cell) accessible via HeaderExt() and CellExt().
//  2. HeaderDescriptor.Assign()
//  3. MatrixData.Clear()
//  4. MatrixData collection API for Columns: IndexOfColumn, ContainsColumn,
//     InsertColumn, RemoveColumn, ColumnsToArray
//  5. MatrixData collection API for Rows: IndexOfRow, ContainsRow,
//     InsertRow, RemoveRow, RowsToArray
//  6. MatrixData collection API for Cells: IndexOfCell, ContainsCell,
//     InsertCell, RemoveCell, CellsToArray

import (
	"testing"

	"github.com/andrewloable/go-fastreport/matrix"
	"github.com/andrewloable/go-fastreport/table"
)

// ── HeaderDescriptor template fields via HeaderExt() ─────────────────────────

func TestHeaderDescriptor_TemplateFields(t *testing.T) {
	h := matrix.NewHeaderDescriptor("[Col]")
	ext := h.HeaderExt()

	col := table.NewTableColumn()
	row := table.NewTableRow()
	cell := table.NewTableCell()

	ext.TemplateColumn = col
	ext.TemplateRow = row
	ext.TemplateCell = cell

	got := h.HeaderExt()
	if got.TemplateColumn != col {
		t.Error("TemplateColumn not set correctly")
	}
	if got.TemplateRow != row {
		t.Error("TemplateRow not set correctly")
	}
	if got.TemplateCell != cell {
		t.Error("TemplateCell not set correctly")
	}
}

func TestHeaderDescriptor_TemplateTotalFields(t *testing.T) {
	h := matrix.NewHeaderDescriptor("[Year]")
	ext := h.HeaderExt()

	col := table.NewTableColumn()
	row := table.NewTableRow()
	cell := table.NewTableCell()

	ext.TemplateTotalColumn = col
	ext.TemplateTotalRow = row
	ext.TemplateTotalCell = cell

	got := h.HeaderExt()
	if got.TemplateTotalColumn != col {
		t.Error("TemplateTotalColumn not set correctly")
	}
	if got.TemplateTotalRow != row {
		t.Error("TemplateTotalRow not set correctly")
	}
	if got.TemplateTotalCell != cell {
		t.Error("TemplateTotalCell not set correctly")
	}
}

// ── HeaderExt is idempotent ───────────────────────────────────────────────────

func TestHeaderExt_Idempotent(t *testing.T) {
	h := matrix.NewHeaderDescriptor("[X]")
	e1 := h.HeaderExt()
	e2 := h.HeaderExt()
	if e1 != e2 {
		t.Error("HeaderExt() should return the same pointer on repeated calls")
	}
}

// ── CellExt ───────────────────────────────────────────────────────────────────

func TestCellExt_Fields(t *testing.T) {
	c := matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum)
	ext := c.CellExt()

	col := table.NewTableColumn()
	ext.TemplateColumn = col

	if c.CellExt().TemplateColumn != col {
		t.Error("TemplateColumn not set on CellExt")
	}
}

func TestCellExt_Idempotent(t *testing.T) {
	c := matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum)
	e1 := c.CellExt()
	e2 := c.CellExt()
	if e1 != e2 {
		t.Error("CellExt() should return the same pointer on repeated calls")
	}
}

// ── HeaderDescriptor.Assign ───────────────────────────────────────────────────

func TestHeaderDescriptor_Assign_CopiesAllFields(t *testing.T) {
	src := matrix.NewHeaderDescriptor("[Sales]")
	src.Sort = matrix.SortOrderDescending
	src.Totals = false
	src.TotalsFirst = true
	src.PageBreak = true
	src.SuppressTotals = true

	totalCell := table.NewTableCell()
	src.HeaderExt().TemplateTotalCell = totalCell
	baseCell := table.NewTableCell()
	src.HeaderExt().TemplateCell = baseCell

	dst := matrix.NewHeaderDescriptor("")
	dst.Assign(src)

	if dst.Expression != "[Sales]" {
		t.Errorf("Expression = %q, want [Sales]", dst.Expression)
	}
	if dst.Sort != matrix.SortOrderDescending {
		t.Errorf("Sort = %v, want Descending", dst.Sort)
	}
	if dst.Totals != false {
		t.Error("Totals should be false after Assign")
	}
	if dst.TotalsFirst != true {
		t.Error("TotalsFirst should be true after Assign")
	}
	if dst.PageBreak != true {
		t.Error("PageBreak should be true after Assign")
	}
	if dst.SuppressTotals != true {
		t.Error("SuppressTotals should be true after Assign")
	}
	if dst.HeaderExt().TemplateTotalCell != totalCell {
		t.Error("TemplateTotalCell not copied by Assign")
	}
	if dst.HeaderExt().TemplateCell != baseCell {
		t.Error("TemplateCell not copied by Assign")
	}
}

func TestHeaderDescriptor_Assign_NoSrcExt(t *testing.T) {
	// Assign from a descriptor whose HeaderExt has never been called
	// (so no entry in the ext map). Must not panic.
	src := matrix.NewHeaderDescriptor("[SrcNoExt]")
	src.Sort = matrix.SortOrderDescending

	dst := matrix.NewHeaderDescriptor("")
	dst.Assign(src) // must not panic

	if dst.Sort != matrix.SortOrderDescending {
		t.Errorf("Sort = %v, want Descending", dst.Sort)
	}
}

// ── MatrixData.Clear ──────────────────────────────────────────────────────────

func TestMatrixData_Clear(t *testing.T) {
	var d matrix.MatrixData
	d.AddColumn(matrix.NewHeaderDescriptor("[c1]"))
	d.AddRow(matrix.NewHeaderDescriptor("[r1]"))
	d.AddCell(matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum))

	d.Clear()

	if len(d.Columns) != 0 {
		t.Errorf("Columns not cleared: len=%d", len(d.Columns))
	}
	if len(d.Rows) != 0 {
		t.Errorf("Rows not cleared: len=%d", len(d.Rows))
	}
	if len(d.Cells) != 0 {
		t.Errorf("Cells not cleared: len=%d", len(d.Cells))
	}
}

func TestMatrixData_Clear_EmptyIsNoop(t *testing.T) {
	var d matrix.MatrixData
	d.Clear() // must not panic on empty data
	if len(d.Columns) != 0 || len(d.Rows) != 0 || len(d.Cells) != 0 {
		t.Error("Clear on empty MatrixData should leave all slices empty")
	}
}

func TestMatrixData_Clear_ThenReAdd(t *testing.T) {
	var d matrix.MatrixData
	d.AddColumn(matrix.NewHeaderDescriptor("[c]"))
	d.AddRow(matrix.NewHeaderDescriptor("[r]"))
	d.AddCell(matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum))

	d.Clear()

	// Re-add after clear.
	h := matrix.NewHeaderDescriptor("[newcol]")
	d.AddColumn(h)

	if len(d.Columns) != 1 || d.Columns[0] != h {
		t.Errorf("After Clear+AddColumn: Columns = %v, want [newcol]", d.Columns)
	}
}

// ── MatrixData Columns collection API ─────────────────────────────────────────

func TestMatrixData_IndexOfColumn_Found(t *testing.T) {
	var d matrix.MatrixData
	h1 := matrix.NewHeaderDescriptor("[c1]")
	h2 := matrix.NewHeaderDescriptor("[c2]")
	d.AddColumn(h1)
	d.AddColumn(h2)

	if got := d.IndexOfColumn(h1); got != 0 {
		t.Errorf("IndexOfColumn(h1) = %d, want 0", got)
	}
	if got := d.IndexOfColumn(h2); got != 1 {
		t.Errorf("IndexOfColumn(h2) = %d, want 1", got)
	}
}

func TestMatrixData_IndexOfColumn_NotFound(t *testing.T) {
	var d matrix.MatrixData
	h := matrix.NewHeaderDescriptor("[c]")
	if got := d.IndexOfColumn(h); got != -1 {
		t.Errorf("IndexOfColumn (absent) = %d, want -1", got)
	}
}

func TestMatrixData_ContainsColumn(t *testing.T) {
	var d matrix.MatrixData
	h := matrix.NewHeaderDescriptor("[c]")
	if d.ContainsColumn(h) {
		t.Error("ContainsColumn should return false before adding")
	}
	d.AddColumn(h)
	if !d.ContainsColumn(h) {
		t.Error("ContainsColumn should return true after adding")
	}
}

func TestMatrixData_InsertColumn_AtFront(t *testing.T) {
	var d matrix.MatrixData
	h1 := matrix.NewHeaderDescriptor("[c1]")
	h2 := matrix.NewHeaderDescriptor("[c2]")
	d.AddColumn(h1)
	d.InsertColumn(0, h2)

	if d.Columns[0] != h2 {
		t.Error("InsertColumn(0) should place h2 at index 0")
	}
	if d.Columns[1] != h1 {
		t.Error("h1 should shift to index 1 after InsertColumn(0)")
	}
}

func TestMatrixData_InsertColumn_AtEnd(t *testing.T) {
	var d matrix.MatrixData
	h1 := matrix.NewHeaderDescriptor("[c1]")
	h2 := matrix.NewHeaderDescriptor("[c2]")
	d.AddColumn(h1)
	d.InsertColumn(1, h2)

	if d.Columns[1] != h2 {
		t.Error("InsertColumn(1) should place h2 at index 1")
	}
}

func TestMatrixData_RemoveColumn_Present(t *testing.T) {
	var d matrix.MatrixData
	h1 := matrix.NewHeaderDescriptor("[c1]")
	h2 := matrix.NewHeaderDescriptor("[c2]")
	d.AddColumn(h1)
	d.AddColumn(h2)
	d.RemoveColumn(h1)

	if len(d.Columns) != 1 {
		t.Errorf("Columns len = %d, want 1", len(d.Columns))
	}
	if d.Columns[0] != h2 {
		t.Error("h2 should remain after removing h1")
	}
}

func TestMatrixData_RemoveColumn_Absent(t *testing.T) {
	var d matrix.MatrixData
	h := matrix.NewHeaderDescriptor("[c]")
	d.AddColumn(matrix.NewHeaderDescriptor("[other]"))
	d.RemoveColumn(h) // must not panic

	if len(d.Columns) != 1 {
		t.Errorf("Columns len = %d, want 1 (absent remove is noop)", len(d.Columns))
	}
}

func TestMatrixData_ColumnsToArray(t *testing.T) {
	var d matrix.MatrixData
	h1 := matrix.NewHeaderDescriptor("[c1]")
	h2 := matrix.NewHeaderDescriptor("[c2]")
	d.AddColumn(h1)
	d.AddColumn(h2)

	arr := d.ColumnsToArray()
	if len(arr) != 2 {
		t.Fatalf("ColumnsToArray len = %d, want 2", len(arr))
	}
	if arr[0] != h1 || arr[1] != h2 {
		t.Error("ColumnsToArray elements mismatch")
	}
	// Verify it's a copy: modifying arr does not affect d.Columns.
	arr[0] = nil
	if d.Columns[0] != h1 {
		t.Error("ColumnsToArray should return an independent copy")
	}
}

// ── MatrixData Rows collection API ────────────────────────────────────────────

func TestMatrixData_IndexOfRow_Found(t *testing.T) {
	var d matrix.MatrixData
	h := matrix.NewHeaderDescriptor("[r1]")
	d.AddRow(h)
	if got := d.IndexOfRow(h); got != 0 {
		t.Errorf("IndexOfRow = %d, want 0", got)
	}
}

func TestMatrixData_IndexOfRow_NotFound(t *testing.T) {
	var d matrix.MatrixData
	h := matrix.NewHeaderDescriptor("[r]")
	if got := d.IndexOfRow(h); got != -1 {
		t.Errorf("IndexOfRow (absent) = %d, want -1", got)
	}
}

func TestMatrixData_ContainsRow(t *testing.T) {
	var d matrix.MatrixData
	h := matrix.NewHeaderDescriptor("[r]")
	if d.ContainsRow(h) {
		t.Error("ContainsRow should return false before adding")
	}
	d.AddRow(h)
	if !d.ContainsRow(h) {
		t.Error("ContainsRow should return true after adding")
	}
}

func TestMatrixData_InsertRow(t *testing.T) {
	var d matrix.MatrixData
	h1 := matrix.NewHeaderDescriptor("[r1]")
	h2 := matrix.NewHeaderDescriptor("[r2]")
	d.AddRow(h1)
	d.InsertRow(0, h2)

	if d.Rows[0] != h2 {
		t.Error("InsertRow(0) should place h2 at index 0")
	}
}

func TestMatrixData_RemoveRow_Present(t *testing.T) {
	var d matrix.MatrixData
	h1 := matrix.NewHeaderDescriptor("[r1]")
	h2 := matrix.NewHeaderDescriptor("[r2]")
	d.AddRow(h1)
	d.AddRow(h2)
	d.RemoveRow(h1)

	if len(d.Rows) != 1 || d.Rows[0] != h2 {
		t.Errorf("RemoveRow: Rows len=%d, want 1 with h2", len(d.Rows))
	}
}

func TestMatrixData_RemoveRow_Absent(t *testing.T) {
	var d matrix.MatrixData
	absent := matrix.NewHeaderDescriptor("[x]")
	d.AddRow(matrix.NewHeaderDescriptor("[r]"))
	d.RemoveRow(absent) // must not panic

	if len(d.Rows) != 1 {
		t.Errorf("Rows len = %d, want 1", len(d.Rows))
	}
}

func TestMatrixData_RowsToArray(t *testing.T) {
	var d matrix.MatrixData
	h := matrix.NewHeaderDescriptor("[r]")
	d.AddRow(h)

	arr := d.RowsToArray()
	if len(arr) != 1 || arr[0] != h {
		t.Error("RowsToArray mismatch")
	}
	// Independence check.
	arr[0] = nil
	if d.Rows[0] != h {
		t.Error("RowsToArray should return an independent copy")
	}
}

// ── MatrixData Cells collection API ───────────────────────────────────────────

func TestMatrixData_IndexOfCell_Found(t *testing.T) {
	var d matrix.MatrixData
	c := matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum)
	d.AddCell(c)
	if got := d.IndexOfCell(c); got != 0 {
		t.Errorf("IndexOfCell = %d, want 0", got)
	}
}

func TestMatrixData_IndexOfCell_NotFound(t *testing.T) {
	var d matrix.MatrixData
	c := matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum)
	if got := d.IndexOfCell(c); got != -1 {
		t.Errorf("IndexOfCell (absent) = %d, want -1", got)
	}
}

func TestMatrixData_ContainsCell(t *testing.T) {
	var d matrix.MatrixData
	c := matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum)
	if d.ContainsCell(c) {
		t.Error("ContainsCell should return false before adding")
	}
	d.AddCell(c)
	if !d.ContainsCell(c) {
		t.Error("ContainsCell should return true after adding")
	}
}

func TestMatrixData_InsertCell(t *testing.T) {
	var d matrix.MatrixData
	c1 := matrix.NewCellDescriptor("[v1]", matrix.AggregateFunctionSum)
	c2 := matrix.NewCellDescriptor("[v2]", matrix.AggregateFunctionCount)
	d.AddCell(c1)
	d.InsertCell(0, c2)

	if d.Cells[0] != c2 {
		t.Error("InsertCell(0) should place c2 at index 0")
	}
	if d.Cells[1] != c1 {
		t.Error("c1 should shift to index 1 after InsertCell(0)")
	}
}

func TestMatrixData_RemoveCell_Present(t *testing.T) {
	var d matrix.MatrixData
	c1 := matrix.NewCellDescriptor("[v1]", matrix.AggregateFunctionSum)
	c2 := matrix.NewCellDescriptor("[v2]", matrix.AggregateFunctionCount)
	d.AddCell(c1)
	d.AddCell(c2)
	d.RemoveCell(c1)

	if len(d.Cells) != 1 || d.Cells[0] != c2 {
		t.Errorf("RemoveCell: Cells len=%d, want 1 with c2", len(d.Cells))
	}
}

func TestMatrixData_RemoveCell_Absent(t *testing.T) {
	var d matrix.MatrixData
	absent := matrix.NewCellDescriptor("[x]", matrix.AggregateFunctionSum)
	d.AddCell(matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum))
	d.RemoveCell(absent) // must not panic

	if len(d.Cells) != 1 {
		t.Errorf("Cells len = %d, want 1", len(d.Cells))
	}
}

func TestMatrixData_CellsToArray(t *testing.T) {
	var d matrix.MatrixData
	c := matrix.NewCellDescriptor("[v]", matrix.AggregateFunctionSum)
	d.AddCell(c)

	arr := d.CellsToArray()
	if len(arr) != 1 || arr[0] != c {
		t.Error("CellsToArray mismatch")
	}
	// Independence check.
	arr[0] = nil
	if d.Cells[0] != c {
		t.Error("CellsToArray should return an independent copy")
	}
}
