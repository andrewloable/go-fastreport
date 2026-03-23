package crossview_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/crossview"
)

// ── mock CubeSource ───────────────────────────────────────────────────────────

// mockCube is a simple in-memory CubeSourceBase.
// It models a 2D table: Product × Region → Sales.
type mockCube struct {
	xFields  []string   // X-axis field names (columns)
	yFields  []string   // Y-axis field names (rows)
	measures []string   // measure names
	xValues  []string   // column header values (ordered by position)
	yValues  []string   // row header values (ordered by position)
	cells    [][]string // cells[row][col]
}

func newMockCube() *mockCube {
	return &mockCube{
		xFields:  []string{"Region"},
		yFields:  []string{"Product"},
		measures: []string{"Sales"},
		xValues:  []string{"North", "South"},
		yValues:  []string{"Apples", "Bananas"},
		cells: [][]string{
			{"100", "200"}, // Apples: North=100, South=200
			{"150", "250"}, // Bananas: North=150, South=250
		},
	}
}

func (m *mockCube) XAxisFieldsCount() int          { return len(m.xFields) }
func (m *mockCube) YAxisFieldsCount() int          { return len(m.yFields) }
func (m *mockCube) MeasuresCount() int             { return len(m.measures) }
func (m *mockCube) GetXAxisFieldName(i int) string { return m.xFields[i] }
func (m *mockCube) GetYAxisFieldName(i int) string { return m.yFields[i] }
func (m *mockCube) GetMeasureName(j int) string    { return m.measures[j] }
func (m *mockCube) DataColumnCount() int           { return len(m.xValues) }
func (m *mockCube) DataRowCount() int              { return len(m.yValues) }
func (m *mockCube) MeasuresInXAxis() bool          { return false }
func (m *mockCube) MeasuresInYAxis() bool          { return false }
func (m *mockCube) MeasuresLevel() int             { return -1 }
func (m *mockCube) SourceAssigned() bool           { return len(m.cells) > 0 }

func (m *mockCube) TraverseXAxis(fn crossview.AxisTraverseFunc) {
	for i, v := range m.xValues {
		fn(crossview.AxisDrawCell{
			Text: v, Cell: i, Level: 0, SizeCell: 1, SizeLevel: 1,
		})
	}
}

func (m *mockCube) TraverseYAxis(fn crossview.AxisTraverseFunc) {
	for i, v := range m.yValues {
		fn(crossview.AxisDrawCell{
			Text: v, Cell: i, Level: 0, SizeCell: 1, SizeLevel: 1,
		})
	}
}

func (m *mockCube) GetMeasureCell(x, y int) crossview.MeasureCell {
	if y >= 0 && y < len(m.cells) && x >= 0 && x < len(m.cells[y]) {
		return crossview.MeasureCell{Text: m.cells[y][x]}
	}
	return crossview.MeasureCell{}
}

// ── CrossViewObject ───────────────────────────────────────────────────────────

func TestNewCrossViewObject_Defaults(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	if !cv.ShowTitle {
		t.Error("ShowTitle should default to true")
	}
	if !cv.ShowXAxisFieldsCaption {
		t.Error("ShowXAxisFieldsCaption should default to true")
	}
	if !cv.ShowYAxisFieldsCaption {
		t.Error("ShowYAxisFieldsCaption should default to true")
	}
}

func TestSetSource_BuildsDescriptors(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.SetSource(newMockCube())

	if len(cv.Data.Columns) != 1 {
		t.Errorf("Columns len = %d, want 1 (one X-axis field)", len(cv.Data.Columns))
	}
	if len(cv.Data.Rows) != 1 {
		t.Errorf("Rows len = %d, want 1 (one Y-axis field)", len(cv.Data.Rows))
	}
	// CreateDescriptors creates one CellDescriptor per data-grid position
	// (dataCols × dataRows = 2 × 2 = 4).
	if len(cv.Data.Cells) != 4 {
		t.Errorf("Cells len = %d, want 4 (dataCols×dataRows)", len(cv.Data.Cells))
	}
	if cv.Data.Columns[0].FieldName != "Region" {
		t.Errorf("Column[0].FieldName = %q, want Region", cv.Data.Columns[0].FieldName)
	}
	if cv.Data.Rows[0].FieldName != "Product" {
		t.Errorf("Row[0].FieldName = %q, want Product", cv.Data.Rows[0].FieldName)
	}
	if cv.Data.Cells[0].MeasureName != "Sales" {
		t.Errorf("Cell[0].MeasureName = %q, want Sales", cv.Data.Cells[0].MeasureName)
	}
}

func TestBuild_NoSource_ReturnsError(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.Name = "TestCV"
	_, err := cv.Build()
	if err == nil {
		t.Error("Build with no source should return error")
	}
}

func TestBuild_GridDimensions(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.SetSource(newMockCube())
	// xFields=1, yFields=1, dataCols=2, dataRows=2
	// totalCols = 1+2=3, totalRows = 1+2=3

	grid, err := cv.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if grid.ColCount != 3 {
		t.Errorf("ColCount = %d, want 3", grid.ColCount)
	}
	if grid.RowCount != 3 {
		t.Errorf("RowCount = %d, want 3", grid.RowCount)
	}
}

func TestBuild_XAxisHeaders(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.SetSource(newMockCube())

	grid, err := cv.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Row 0 (X-axis header row), col 1 = first X value "North"
	north := grid.Cell(0, 1)
	if north.Text != "North" {
		t.Errorf("Cell(0,1).Text = %q, want 'North'", north.Text)
	}
	if !north.IsColLabel {
		t.Error("Cell(0,1) should be IsColLabel=true")
	}

	south := grid.Cell(0, 2)
	if south.Text != "South" {
		t.Errorf("Cell(0,2).Text = %q, want 'South'", south.Text)
	}
}

func TestBuild_YAxisHeaders(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.SetSource(newMockCube())

	grid, err := cv.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Col 0, row 1 = first Y value "Apples"
	apples := grid.Cell(1, 0)
	if apples.Text != "Apples" {
		t.Errorf("Cell(1,0).Text = %q, want 'Apples'", apples.Text)
	}
	if !apples.IsRowLabel {
		t.Error("Cell(1,0) should be IsRowLabel=true")
	}

	bananas := grid.Cell(2, 0)
	if bananas.Text != "Bananas" {
		t.Errorf("Cell(2,0).Text = %q, want 'Bananas'", bananas.Text)
	}
}

func TestBuild_DataCells(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.SetSource(newMockCube())

	grid, err := cv.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Data region starts at (row=1, col=1):
	// [1][1] = Apples/North = "100"
	// [1][2] = Apples/South = "200"
	// [2][1] = Bananas/North = "150"
	// [2][2] = Bananas/South = "250"
	cases := []struct {
		row, col int
		want     string
	}{
		{1, 1, "100"},
		{1, 2, "200"},
		{2, 1, "150"},
		{2, 2, "250"},
	}
	for _, tc := range cases {
		c := grid.Cell(tc.row, tc.col)
		if c.Text != tc.want {
			t.Errorf("Cell(%d,%d).Text = %q, want %q", tc.row, tc.col, c.Text, tc.want)
		}
	}
}

func TestBuild_CornerCell_ShowsFieldName(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.SetSource(newMockCube())

	grid, err := cv.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Corner cell (0,0): should show Y-axis field name "Product"
	corner := grid.Cell(0, 0)
	if corner.Text != "Product" {
		t.Errorf("Corner cell text = %q, want 'Product'", corner.Text)
	}
}

// ── CrossViewData ─────────────────────────────────────────────────────────────

func TestCrossViewData_Clear(t *testing.T) {
	var d crossview.CrossViewData
	d.AddColumn(&crossview.HeaderDescriptor{FieldName: "A"})
	d.AddRow(&crossview.HeaderDescriptor{FieldName: "B"})
	d.AddCell(&crossview.CellDescriptor{MeasureName: "C"})

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

// ── Multi-level X-axis ────────────────────────────────────────────────────────

// multiLevelCube models Year → Quarter layout on X-axis.
type multiLevelCube struct{}

func (m *multiLevelCube) XAxisFieldsCount() int          { return 2 } // Year, Quarter
func (m *multiLevelCube) YAxisFieldsCount() int          { return 1 }
func (m *multiLevelCube) MeasuresCount() int             { return 1 }
func (m *multiLevelCube) GetXAxisFieldName(i int) string { return []string{"Year", "Quarter"}[i] }
func (m *multiLevelCube) GetYAxisFieldName(i int) string { return "Product" }
func (m *multiLevelCube) GetMeasureName(j int) string    { return "Sales" }
func (m *multiLevelCube) DataColumnCount() int           { return 2 } // Q1, Q2
func (m *multiLevelCube) DataRowCount() int              { return 1 }
func (m *multiLevelCube) MeasuresInXAxis() bool          { return false }
func (m *multiLevelCube) MeasuresInYAxis() bool          { return false }
func (m *multiLevelCube) MeasuresLevel() int             { return -1 }
func (m *multiLevelCube) SourceAssigned() bool           { return true }

func (m *multiLevelCube) TraverseXAxis(fn crossview.AxisTraverseFunc) {
	// Level 0: 2024 spanning both quarters
	fn(crossview.AxisDrawCell{Text: "2024", Cell: 0, Level: 0, SizeCell: 2, SizeLevel: 1})
	// Level 1: Q1, Q2
	fn(crossview.AxisDrawCell{Text: "Q1", Cell: 0, Level: 1, SizeCell: 1, SizeLevel: 1})
	fn(crossview.AxisDrawCell{Text: "Q2", Cell: 1, Level: 1, SizeCell: 1, SizeLevel: 1})
}
func (m *multiLevelCube) TraverseYAxis(fn crossview.AxisTraverseFunc) {
	fn(crossview.AxisDrawCell{Text: "Apples", Cell: 0, Level: 0, SizeCell: 1, SizeLevel: 1})
}
func (m *multiLevelCube) GetMeasureCell(x, y int) crossview.MeasureCell {
	vals := []string{"500", "300"}
	if x < len(vals) {
		return crossview.MeasureCell{Text: vals[x]}
	}
	return crossview.MeasureCell{}
}

func TestBuild_MultiLevelXAxis(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.SetSource(&multiLevelCube{})

	grid, err := cv.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// totalCols = yFields(1) + dataCols(2) = 3
	// totalRows = xFields(2) + dataRows(1) = 3
	if grid.ColCount != 3 {
		t.Errorf("ColCount = %d, want 3", grid.ColCount)
	}
	if grid.RowCount != 3 {
		t.Errorf("RowCount = %d, want 3", grid.RowCount)
	}

	// Row 0 (level 0): "2024" at col 1 with ColSpan=2
	cell2024 := grid.Cell(0, 1)
	if cell2024.Text != "2024" {
		t.Errorf("Cell(0,1).Text = %q, want '2024'", cell2024.Text)
	}
	if cell2024.ColSpan != 2 {
		t.Errorf("Cell(0,1).ColSpan = %d, want 2", cell2024.ColSpan)
	}

	// Row 1 (level 1): Q1, Q2
	q1 := grid.Cell(1, 1)
	if q1.Text != "Q1" {
		t.Errorf("Cell(1,1).Text = %q, want 'Q1'", q1.Text)
	}
}

// ── Multi-measure X-axis ──────────────────────────────────────────────────────

// multiMeasureCube: Region (X) × Product (Y), two measures: Sales + Qty.
// MeasuresInXAxis = true, so each leaf column has two sub-columns.
type multiMeasureCube struct{}

func (m *multiMeasureCube) XAxisFieldsCount() int          { return 1 }
func (m *multiMeasureCube) YAxisFieldsCount() int          { return 1 }
func (m *multiMeasureCube) MeasuresCount() int             { return 2 }
func (m *multiMeasureCube) GetXAxisFieldName(i int) string { return "Region" }
func (m *multiMeasureCube) GetYAxisFieldName(i int) string { return "Product" }
func (m *multiMeasureCube) GetMeasureName(j int) string    { return []string{"Sales", "Qty"}[j] }
func (m *multiMeasureCube) DataColumnCount() int           { return 4 } // North/Sales, North/Qty, South/Sales, South/Qty
func (m *multiMeasureCube) DataRowCount() int              { return 2 }
func (m *multiMeasureCube) MeasuresInXAxis() bool          { return true }
func (m *multiMeasureCube) MeasuresInYAxis() bool          { return false }
func (m *multiMeasureCube) MeasuresLevel() int             { return -1 } // innermost
func (m *multiMeasureCube) SourceAssigned() bool           { return true }

func (m *multiMeasureCube) TraverseXAxis(fn crossview.AxisTraverseFunc) {
	// Level 0: Region values (each spans 2 measure columns)
	fn(crossview.AxisDrawCell{Text: "North", Cell: 0, Level: 0, SizeCell: 2, SizeLevel: 1})
	fn(crossview.AxisDrawCell{Text: "South", Cell: 2, Level: 0, SizeCell: 2, SizeLevel: 1})
	// Level 1 (measure): Sales, Qty, Sales, Qty
	fn(crossview.AxisDrawCell{Text: "Sales", Cell: 0, Level: 1, SizeCell: 1, SizeLevel: 1})
	fn(crossview.AxisDrawCell{Text: "Qty", Cell: 1, Level: 1, SizeCell: 1, SizeLevel: 1})
	fn(crossview.AxisDrawCell{Text: "Sales", Cell: 2, Level: 1, SizeCell: 1, SizeLevel: 1})
	fn(crossview.AxisDrawCell{Text: "Qty", Cell: 3, Level: 1, SizeCell: 1, SizeLevel: 1})
}

func (m *multiMeasureCube) TraverseYAxis(fn crossview.AxisTraverseFunc) {
	fn(crossview.AxisDrawCell{Text: "Apples", Cell: 0, Level: 0, SizeCell: 1, SizeLevel: 1})
	fn(crossview.AxisDrawCell{Text: "Bananas", Cell: 1, Level: 0, SizeCell: 1, SizeLevel: 1})
}

func (m *multiMeasureCube) GetMeasureCell(x, y int) crossview.MeasureCell {
	// [row][col]: 4 cols (N/Sales, N/Qty, S/Sales, S/Qty), 2 rows
	data := [][]string{
		{"100", "10", "200", "20"},
		{"150", "15", "250", "25"},
	}
	if y >= 0 && y < len(data) && x >= 0 && x < len(data[y]) {
		return crossview.MeasureCell{Text: data[y][x]}
	}
	return crossview.MeasureCell{}
}

func TestCreateDescriptors_MultiMeasureInXAxis(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.SetSource(&multiMeasureCube{})

	// Columns: Region(level0) + Sales(level1,measure) + Qty(level1,measure) = 3
	if len(cv.Data.Columns) != 3 {
		t.Errorf("Columns len = %d, want 3", len(cv.Data.Columns))
	}
	// Terminal column indexes = measure descriptors (indices 1 and 2)
	terminals := cv.Data.ColumnTerminalIndexes()
	if len(terminals) != 2 {
		t.Errorf("ColumnTerminalIndexes len = %d, want 2", len(terminals))
	}
	// Measure descriptors should have IsMeasure=true
	for _, idx := range terminals {
		if !cv.Data.Columns[idx].IsMeasure {
			t.Errorf("Columns[%d].IsMeasure = false, want true", idx)
		}
	}
	// Row: Product = 1 descriptor, 1 terminal
	if len(cv.Data.Rows) != 1 {
		t.Errorf("Rows len = %d, want 1", len(cv.Data.Rows))
	}
	if len(cv.Data.RowTerminalIndexes()) != 1 {
		t.Errorf("RowTerminalIndexes len = %d, want 1", len(cv.Data.RowTerminalIndexes()))
	}
	// Cells: dataCols(4) × dataRows(2) = 8
	if len(cv.Data.Cells) != 8 {
		t.Errorf("Cells len = %d, want 8", len(cv.Data.Cells))
	}
}

func TestBuild_MultiMeasureInXAxis_GridDimensions(t *testing.T) {
	cv := crossview.NewCrossViewObject()
	cv.SetSource(&multiMeasureCube{})

	grid, err := cv.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// totalCols = yFields(1) + dataCols(4) = 5
	// totalRows = xFields(2) + dataRows(2) = 4  (xFields includes measure level)
	if grid.ColCount != 5 {
		t.Errorf("ColCount = %d, want 5", grid.ColCount)
	}
	if grid.RowCount != 4 {
		t.Errorf("RowCount = %d, want 4", grid.RowCount)
	}
	// Row 0 (level 0): "North" at col 1 with ColSpan=2
	north := grid.Cell(0, 1)
	if north.Text != "North" {
		t.Errorf("Cell(0,1).Text = %q, want 'North'", north.Text)
	}
	if north.ColSpan != 2 {
		t.Errorf("Cell(0,1).ColSpan = %d, want 2", north.ColSpan)
	}
	// Data cell at (row=2, col=1) = Apples/North/Sales = "100"
	c := grid.Cell(2, 1)
	if c.Text != "100" {
		t.Errorf("Cell(2,1).Text = %q, want '100'", c.Text)
	}
}

// ── ResultGrid helper ─────────────────────────────────────────────────────────

func TestResultGrid_CellOutOfRange(t *testing.T) {
	var g crossview.ResultGrid
	c := g.Cell(99, 99)
	if c.Text != "" {
		t.Error("out-of-range Cell should return zero ResultCell")
	}
}

// ── HeaderDescriptor.GetName ──────────────────────────────────────────────────

// TestHeaderDescriptor_GetName verifies the GetName() method against the C#
// CrossViewHeaderDescriptor.GetName() logic in CrossViewHeaderDescriptor.cs.
func TestHeaderDescriptor_GetName(t *testing.T) {
	cases := []struct {
		desc *crossview.HeaderDescriptor
		want string
	}{
		{
			desc: &crossview.HeaderDescriptor{IsGrandTotal: true, FieldName: "Category"},
			want: "GrandTotal",
		},
		{
			desc: &crossview.HeaderDescriptor{IsMeasure: true, MeasureName: "Sales"},
			want: "Sales",
		},
		{
			desc: &crossview.HeaderDescriptor{IsTotal: true, FieldName: "Region"},
			want: "Total of Region",
		},
		{
			desc: &crossview.HeaderDescriptor{FieldName: "Product"},
			want: "Product",
		},
	}
	for _, tc := range cases {
		got := tc.desc.GetName()
		if got != tc.want {
			t.Errorf("GetName() = %q, want %q", got, tc.want)
		}
	}
}

// ── HeaderDescriptor.Assign ───────────────────────────────────────────────────

// TestHeaderDescriptor_Assign verifies field copying against C# Assign().
func TestHeaderDescriptor_Assign(t *testing.T) {
	src := &crossview.HeaderDescriptor{
		FieldName:    "Region",
		MeasureName:  "Sales",
		IsGrandTotal: true,
		IsTotal:      false,
		IsMeasure:    false,
		Level:        2,
		LevelSize:    3,
		Cell:         5,
		CellSize:     4,
	}
	src.Expression = "[Region]"

	var dst crossview.HeaderDescriptor
	dst.Assign(src)

	if dst.FieldName != src.FieldName {
		t.Errorf("FieldName: got %q, want %q", dst.FieldName, src.FieldName)
	}
	if dst.MeasureName != src.MeasureName {
		t.Errorf("MeasureName: got %q, want %q", dst.MeasureName, src.MeasureName)
	}
	if dst.IsGrandTotal != src.IsGrandTotal {
		t.Errorf("IsGrandTotal: got %v, want %v", dst.IsGrandTotal, src.IsGrandTotal)
	}
	if dst.Level != src.Level {
		t.Errorf("Level: got %d, want %d", dst.Level, src.Level)
	}
	if dst.LevelSize != src.LevelSize {
		t.Errorf("LevelSize: got %d, want %d", dst.LevelSize, src.LevelSize)
	}
	if dst.Cell != src.Cell {
		t.Errorf("Cell: got %d, want %d", dst.Cell, src.Cell)
	}
	if dst.CellSize != src.CellSize {
		t.Errorf("CellSize: got %d, want %d", dst.CellSize, src.CellSize)
	}
	if dst.Expression != src.Expression {
		t.Errorf("Expression: got %q, want %q", dst.Expression, src.Expression)
	}
}

// TestHeaderDescriptor_Assign_Nil verifies Assign with nil src is a no-op.
func TestHeaderDescriptor_Assign_Nil(t *testing.T) {
	var dst crossview.HeaderDescriptor
	dst.FieldName = "Original"
	dst.Assign(nil) // should not panic or change dst
	if dst.FieldName != "Original" {
		t.Errorf("Assign(nil) changed FieldName to %q", dst.FieldName)
	}
}

// ── CellDescriptor.Assign ─────────────────────────────────────────────────────

// TestCellDescriptor_Assign verifies field copying against C# CrossViewCellDescriptor.Assign().
func TestCellDescriptor_Assign(t *testing.T) {
	src := &crossview.CellDescriptor{
		XFieldName:    "Region",
		YFieldName:    "Product",
		MeasureName:   "Sales",
		IsXTotal:      true,
		IsYTotal:      false,
		IsXGrandTotal: false,
		IsYGrandTotal: true,
		X:             3,
		Y:             7,
	}
	src.Expression = "[Sales]"

	var dst crossview.CellDescriptor
	dst.Assign(src)

	if dst.XFieldName != src.XFieldName {
		t.Errorf("XFieldName: got %q, want %q", dst.XFieldName, src.XFieldName)
	}
	if dst.YFieldName != src.YFieldName {
		t.Errorf("YFieldName: got %q, want %q", dst.YFieldName, src.YFieldName)
	}
	if dst.MeasureName != src.MeasureName {
		t.Errorf("MeasureName: got %q, want %q", dst.MeasureName, src.MeasureName)
	}
	if dst.IsXTotal != src.IsXTotal {
		t.Errorf("IsXTotal: got %v, want %v", dst.IsXTotal, src.IsXTotal)
	}
	if dst.IsYGrandTotal != src.IsYGrandTotal {
		t.Errorf("IsYGrandTotal: got %v, want %v", dst.IsYGrandTotal, src.IsYGrandTotal)
	}
	if dst.X != src.X {
		t.Errorf("X: got %d, want %d", dst.X, src.X)
	}
	if dst.Y != src.Y {
		t.Errorf("Y: got %d, want %d", dst.Y, src.Y)
	}
	if dst.Expression != src.Expression {
		t.Errorf("Expression: got %q, want %q", dst.Expression, src.Expression)
	}
}

// TestCellDescriptor_Assign_Nil verifies Assign with nil src is a no-op.
func TestCellDescriptor_Assign_Nil(t *testing.T) {
	var dst crossview.CellDescriptor
	dst.XFieldName = "Original"
	dst.Assign(nil) // should not panic or change dst
	if dst.XFieldName != "Original" {
		t.Errorf("Assign(nil) changed XFieldName to %q", dst.XFieldName)
	}
}
