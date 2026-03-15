package crossview_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/crossview"
)

// buildSalesSource creates a test SliceCubeSource with:
//
//	X axis: Category
//	Y axis: Region
//	Measure: Sales
//
// Data:
//
//	Category=A, Region=North, Sales=100
//	Category=A, Region=South, Sales=200
//	Category=B, Region=North, Sales=300
//	Category=B, Region=South, Sales=400
func buildSalesSource() *crossview.SliceCubeSource {
	src := crossview.NewSliceCubeSource()
	src.AddXAxisField("Category")
	src.AddYAxisField("Region")
	src.AddMeasure("Sales")
	src.AddRows([]map[string]any{
		{"Category": "A", "Region": "North", "Sales": 100},
		{"Category": "A", "Region": "South", "Sales": 200},
		{"Category": "B", "Region": "North", "Sales": 300},
		{"Category": "B", "Region": "South", "Sales": 400},
	})
	src.Build()
	return src
}

func TestSliceCubeSource_FieldCounts(t *testing.T) {
	src := buildSalesSource()

	if src.XAxisFieldsCount() != 1 {
		t.Errorf("XAxisFieldsCount: got %d, want 1", src.XAxisFieldsCount())
	}
	if src.YAxisFieldsCount() != 1 {
		t.Errorf("YAxisFieldsCount: got %d, want 1", src.YAxisFieldsCount())
	}
	if src.MeasuresCount() != 1 {
		t.Errorf("MeasuresCount: got %d, want 1", src.MeasuresCount())
	}
	if src.GetXAxisFieldName(0) != "Category" {
		t.Errorf("GetXAxisFieldName(0): got %q, want Category", src.GetXAxisFieldName(0))
	}
	if src.GetYAxisFieldName(0) != "Region" {
		t.Errorf("GetYAxisFieldName(0): got %q, want Region", src.GetYAxisFieldName(0))
	}
	if src.GetMeasureName(0) != "Sales" {
		t.Errorf("GetMeasureName(0): got %q, want Sales", src.GetMeasureName(0))
	}
}

func TestSliceCubeSource_DataCounts(t *testing.T) {
	src := buildSalesSource()

	if src.DataColumnCount() != 2 {
		t.Errorf("DataColumnCount: got %d, want 2 (A and B)", src.DataColumnCount())
	}
	if src.DataRowCount() != 2 {
		t.Errorf("DataRowCount: got %d, want 2 (North and South)", src.DataRowCount())
	}
}

func TestSliceCubeSource_TraverseXAxis(t *testing.T) {
	src := buildSalesSource()

	var cells []crossview.AxisDrawCell
	src.TraverseXAxis(func(c crossview.AxisDrawCell) {
		cells = append(cells, c)
	})

	if len(cells) != 2 {
		t.Fatalf("TraverseXAxis: got %d cells, want 2", len(cells))
	}
	// Cells should be "A" at col 0 and "B" at col 1.
	if cells[0].Text != "A" || cells[0].Cell != 0 {
		t.Errorf("X cell 0: got {%q, %d}, want {A, 0}", cells[0].Text, cells[0].Cell)
	}
	if cells[1].Text != "B" || cells[1].Cell != 1 {
		t.Errorf("X cell 1: got {%q, %d}, want {B, 1}", cells[1].Text, cells[1].Cell)
	}
}

func TestSliceCubeSource_TraverseYAxis(t *testing.T) {
	src := buildSalesSource()

	var cells []crossview.AxisDrawCell
	src.TraverseYAxis(func(c crossview.AxisDrawCell) {
		cells = append(cells, c)
	})

	if len(cells) != 2 {
		t.Fatalf("TraverseYAxis: got %d cells, want 2", len(cells))
	}
	if cells[0].Text != "North" || cells[0].Cell != 0 {
		t.Errorf("Y cell 0: got {%q, %d}", cells[0].Text, cells[0].Cell)
	}
	if cells[1].Text != "South" || cells[1].Cell != 1 {
		t.Errorf("Y cell 1: got {%q, %d}", cells[1].Text, cells[1].Cell)
	}
}

func TestSliceCubeSource_GetMeasureCell(t *testing.T) {
	src := buildSalesSource()

	tests := []struct {
		x, y    int
		wantText string
	}{
		{0, 0, "100"}, // A, North
		{0, 1, "200"}, // A, South
		{1, 0, "300"}, // B, North
		{1, 1, "400"}, // B, South
	}
	for _, tt := range tests {
		mc := src.GetMeasureCell(tt.x, tt.y)
		if mc.Text != tt.wantText {
			t.Errorf("GetMeasureCell(%d,%d) = %q, want %q", tt.x, tt.y, mc.Text, tt.wantText)
		}
	}
}

func TestSliceCubeSource_IntegrationWithCrossView(t *testing.T) {
	src := buildSalesSource()

	cv := crossview.NewCrossViewObject()
	cv.SetSource(src)

	grid, err := cv.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Grid should be 3×3:
	//  [R0C0=corner] [R0C1=A] [R0C2=B]
	//  [R1C0=North]  [R1C1=100] [R1C2=300]
	//  [R2C0=South]  [R2C1=200] [R2C2=400]
	if grid.RowCount != 3 {
		t.Errorf("grid.RowCount = %d, want 3", grid.RowCount)
	}
	if grid.ColCount != 3 {
		t.Errorf("grid.ColCount = %d, want 3", grid.ColCount)
	}

	// Data cells.
	dataTests := []struct{ row, col int; text string }{
		{1, 1, "100"},
		{1, 2, "300"},
		{2, 1, "200"},
		{2, 2, "400"},
	}
	for _, dt := range dataTests {
		c := grid.Cell(dt.row, dt.col)
		if c.Text != dt.text {
			t.Errorf("Cell(%d,%d).Text = %q, want %q", dt.row, dt.col, c.Text, dt.text)
		}
	}
}

func TestSliceCubeSource_ImplementsInterface(t *testing.T) {
	var _ crossview.CubeSourceBase = (*crossview.SliceCubeSource)(nil)
}
