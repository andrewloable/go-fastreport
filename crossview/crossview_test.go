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
	if len(cv.Data.Cells) != 1 {
		t.Errorf("Cells len = %d, want 1 (one measure)", len(cv.Data.Cells))
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

// ── ResultGrid helper ─────────────────────────────────────────────────────────

func TestResultGrid_CellOutOfRange(t *testing.T) {
	var g crossview.ResultGrid
	c := g.Cell(99, 99)
	if c.Text != "" {
		t.Error("out-of-range Cell should return zero ResultCell")
	}
}
