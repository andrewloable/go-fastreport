package matrix_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/matrix"
)

// ── Constructor ────────────────────────────────────────────────────────────────

func TestNew(t *testing.T) {
	m := matrix.New()
	if m == nil {
		t.Fatal("New returned nil")
	}
	if m.TypeName() != "MatrixObject" {
		t.Errorf("TypeName = %q, want MatrixObject", m.TypeName())
	}
}

// ── Descriptors ───────────────────────────────────────────────────────────────

func TestNewHeaderDescriptor(t *testing.T) {
	h := matrix.NewHeaderDescriptor("[Sales.Year]")
	if h.Expression != "[Sales.Year]" {
		t.Errorf("Expression = %q", h.Expression)
	}
	if h.TotalText != "Total" {
		t.Errorf("TotalText = %q, want Total", h.TotalText)
	}
}

func TestNewCellDescriptor(t *testing.T) {
	c := matrix.NewCellDescriptor("[Sales.Revenue]", matrix.AggregateFunctionSum)
	if c.Expression != "[Sales.Revenue]" {
		t.Errorf("Expression = %q", c.Expression)
	}
	if c.Function != matrix.AggregateFunctionSum {
		t.Errorf("Function = %v, want Sum", c.Function)
	}
}

// ── MatrixData ────────────────────────────────────────────────────────────────

func TestMatrixData_AddColumn(t *testing.T) {
	var d matrix.MatrixData
	d.AddColumn(matrix.NewHeaderDescriptor("[col1]"))
	d.AddColumn(matrix.NewHeaderDescriptor("[col2]"))
	if len(d.Columns) != 2 {
		t.Errorf("Columns len = %d, want 2", len(d.Columns))
	}
}

func TestMatrixData_AddRow(t *testing.T) {
	var d matrix.MatrixData
	d.AddRow(matrix.NewHeaderDescriptor("[row1]"))
	if len(d.Rows) != 1 {
		t.Errorf("Rows len = %d, want 1", len(d.Rows))
	}
}

func TestMatrixData_AddCell(t *testing.T) {
	var d matrix.MatrixData
	d.AddCell(matrix.NewCellDescriptor("[val]", matrix.AggregateFunctionSum))
	if len(d.Cells) != 1 {
		t.Errorf("Cells len = %d, want 1", len(d.Cells))
	}
}

// ── AddData and aggregation ───────────────────────────────────────────────────

func TestAddData_Sum(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Revenue]", matrix.AggregateFunctionSum))

	m.AddData("Alice", "Q1", []any{100.0})
	m.AddData("Alice", "Q1", []any{200.0})

	result, err := m.CellResult("Alice", "Q1", 0)
	if err != nil {
		t.Fatalf("CellResult: %v", err)
	}
	if result != 300 {
		t.Errorf("Sum = %v, want 300", result)
	}
}

func TestAddData_Count(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Count]", matrix.AggregateFunctionCount))

	m.AddData("Alice", "Q1", []any{1.0})
	m.AddData("Alice", "Q1", []any{2.0})
	m.AddData("Alice", "Q1", []any{3.0})

	result, _ := m.CellResult("Alice", "Q1", 0)
	if result != 3 {
		t.Errorf("Count = %v, want 3", result)
	}
}

func TestAddData_Avg(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Score]", matrix.AggregateFunctionAvg))

	m.AddData("Bob", "Jan", []any{10.0})
	m.AddData("Bob", "Jan", []any{20.0})
	m.AddData("Bob", "Jan", []any{30.0})

	result, _ := m.CellResult("Bob", "Jan", 0)
	if result != 20 {
		t.Errorf("Avg = %v, want 20", result)
	}
}

func TestAddData_Min(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Val]", matrix.AggregateFunctionMin))

	m.AddData("row", "col", []any{5.0})
	m.AddData("row", "col", []any{3.0})
	m.AddData("row", "col", []any{7.0})

	result, _ := m.CellResult("row", "col", 0)
	if result != 3 {
		t.Errorf("Min = %v, want 3", result)
	}
}

func TestAddData_Max(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Val]", matrix.AggregateFunctionMax))

	m.AddData("row", "col", []any{5.0})
	m.AddData("row", "col", []any{3.0})
	m.AddData("row", "col", []any{7.0})

	result, _ := m.CellResult("row", "col", 0)
	if result != 7 {
		t.Errorf("Max = %v, want 7", result)
	}
}

func TestAddData_CountDistinct(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Val]", matrix.AggregateFunctionCountDistinct))

	m.AddData("row", "col", []any{"A"})
	m.AddData("row", "col", []any{"B"})
	m.AddData("row", "col", []any{"A"}) // duplicate

	result, _ := m.CellResult("row", "col", 0)
	if result != 2 {
		t.Errorf("CountDistinct = %v, want 2", result)
	}
}

func TestAddData_None(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Val]", matrix.AggregateFunctionNone))

	m.AddData("row", "col", []any{42.0})
	result, _ := m.CellResult("row", "col", 0)
	if result != 0 {
		t.Errorf("None = %v, want 0", result)
	}
}

func TestAddData_AvgZeroDivide(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Val]", matrix.AggregateFunctionAvg))
	// No AddData calls — accumulator not created.
	_, err := m.CellResult("row", "col", 0)
	if err == nil {
		t.Error("expected error for missing cell data")
	}
}

// ── RowValues / ColValues ──────────────────────────────────────────────────────

func TestRowColValues_InsertionOrder(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Val]", matrix.AggregateFunctionSum))

	m.AddData("Alice", "Q1", []any{1.0})
	m.AddData("Bob", "Q2", []any{2.0})
	m.AddData("Alice", "Q3", []any{3.0})

	rows := m.RowValues()
	if len(rows) != 2 || rows[0] != "Alice" || rows[1] != "Bob" {
		t.Errorf("RowValues = %v, want [Alice Bob]", rows)
	}
	cols := m.ColValues()
	if len(cols) != 3 || cols[0] != "Q1" || cols[1] != "Q2" || cols[2] != "Q3" {
		t.Errorf("ColValues = %v, want [Q1 Q2 Q3]", cols)
	}
}

// ── BuildTemplate ─────────────────────────────────────────────────────────────

func TestBuildTemplate_Basic(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Revenue]", matrix.AggregateFunctionSum))

	m.AddData("Alice", "Q1", []any{100.0})
	m.AddData("Bob", "Q1", []any{200.0})
	m.AddData("Alice", "Q2", []any{50.0})

	m.BuildTemplate()

	// Expect: 1 header row + 2 data rows = 3 rows total.
	if m.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", m.RowCount())
	}
	// Expect: 1 label col + 2 data cols = 3 cols per row.
	if m.ColumnCount() != 3 {
		t.Errorf("ColumnCount = %d, want 3", m.ColumnCount())
	}
}

func TestBuildTemplate_Empty(t *testing.T) {
	m := matrix.New()
	m.BuildTemplate()
	if m.RowCount() != 0 {
		t.Errorf("empty BuildTemplate RowCount = %d, want 0", m.RowCount())
	}
}

func TestBuildTemplate_CellText(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Val]", matrix.AggregateFunctionSum))

	m.AddData("Alice", "Q1", []any{42.0})
	m.BuildTemplate()

	// Row 0 is the header row. Row 1 is "Alice".
	// Column 0 is label, column 1 is Q1 value.
	cell := m.Cell(1, 1)
	if cell == nil {
		t.Fatal("Cell(1,1) is nil")
	}
	if cell.Text() != "42" {
		t.Errorf("Cell(1,1).Text = %q, want 42", cell.Text())
	}
}

// ── MultipleCells ─────────────────────────────────────────────────────────────

func TestAddData_MultipleCellDescriptors(t *testing.T) {
	m := matrix.New()
	m.Data.AddCell(matrix.NewCellDescriptor("[Revenue]", matrix.AggregateFunctionSum))
	m.Data.AddCell(matrix.NewCellDescriptor("[Count]", matrix.AggregateFunctionCount))

	m.AddData("Alice", "Q1", []any{100.0, 1.0})
	m.AddData("Alice", "Q1", []any{200.0, 1.0})

	rev, _ := m.CellResult("Alice", "Q1", 0)
	cnt, _ := m.CellResult("Alice", "Q1", 1)
	if rev != 300 {
		t.Errorf("Revenue Sum = %v, want 300", rev)
	}
	if cnt != 2 {
		t.Errorf("Count = %v, want 2", cnt)
	}
}

// ── EvenStylePriority / flags ─────────────────────────────────────────────────

func TestMatrixObject_Fields(t *testing.T) {
	m := matrix.New()
	m.AutoSize = true
	m.CellsSideBySide = true
	m.EvenStylePriority = matrix.EvenStylePriorityColumns
	m.Style = "Green"

	if !m.AutoSize {
		t.Error("AutoSize should be true")
	}
	if !m.CellsSideBySide {
		t.Error("CellsSideBySide should be true")
	}
	if m.EvenStylePriority != matrix.EvenStylePriorityColumns {
		t.Error("EvenStylePriority mismatch")
	}
	if m.Style != "Green" {
		t.Errorf("Style = %q, want Green", m.Style)
	}
}
