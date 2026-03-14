// Package crossview implements a cross-tab (pivot table) object for go-fastreport.
// It is the Go equivalent of FastReport.CrossView.CrossViewObject.
//
// Architecture:
//
//	CubeSourceBase  — data source interface (provides axis fields and measure values)
//	CrossViewData   — descriptor model (header and cell descriptors for the grid)
//	CrossViewObject — the renderable object that integrates source + data + layout
//	CrossViewResult — computed grid ready for export/preview
package crossview

import "fmt"

// ── Axis traversal callbacks ──────────────────────────────────────────────────

// AxisDrawCell is the data passed to a TraverseXAxis or TraverseYAxis callback
// for each header cell encountered.
type AxisDrawCell struct {
	Text      string
	Cell      int // position along the data axis
	Level     int // nesting depth (0 = outermost)
	SizeCell  int // span in the data direction
	SizeLevel int // span in the level direction
}

// MeasureCell carries the value returned by GetMeasureCell.
type MeasureCell struct {
	Text string
}

// AxisTraverseFunc is the callback type used by TraverseXAxis and TraverseYAxis.
type AxisTraverseFunc func(cell AxisDrawCell)

// ── CubeSourceBase ────────────────────────────────────────────────────────────

// CubeSourceBase is the data source interface for a CrossViewObject.
// Concrete implementations provide axis metadata and measure values.
type CubeSourceBase interface {
	// Axis field counts.
	XAxisFieldsCount() int
	YAxisFieldsCount() int
	MeasuresCount() int

	// Field names.
	GetXAxisFieldName(i int) string
	GetYAxisFieldName(i int) string
	GetMeasureName(j int) string

	// Result grid dimensions (excluding headers).
	DataColumnCount() int
	DataRowCount() int

	// Axis traversal. The handler is called once per header cell.
	TraverseXAxis(fn AxisTraverseFunc)
	TraverseYAxis(fn AxisTraverseFunc)

	// Cell value retrieval. x/y are 0-based data-grid coordinates.
	GetMeasureCell(x, y int) MeasureCell
}

// ── Descriptor base ───────────────────────────────────────────────────────────

// Descriptor is the base type for CrossView descriptors.
type Descriptor struct {
	// Expression is evaluated to fill the descriptor cell.
	Expression string
}

// ── Header descriptor ─────────────────────────────────────────────────────────

// HeaderDescriptor describes a single header cell on the X or Y axis.
type HeaderDescriptor struct {
	Descriptor

	FieldName   string
	MeasureName string

	IsGrandTotal bool
	IsTotal      bool
	IsMeasure    bool

	// Layout geometry.
	Level     int // nesting depth (0 = top)
	Cell      int // position along the data axis
	LevelSize int // span in the level direction
	CellSize  int // span in the data direction
}

// ── Cell descriptor ───────────────────────────────────────────────────────────

// CellDescriptor describes a single data cell in the cross-tab grid.
type CellDescriptor struct {
	Descriptor

	XFieldName  string
	YFieldName  string
	MeasureName string

	IsXTotal      bool
	IsYTotal      bool
	IsXGrandTotal bool
	IsYGrandTotal bool

	// Grid coordinates.
	X int
	Y int
}

// ── CrossViewData ─────────────────────────────────────────────────────────────

// CrossViewData holds the descriptor model for the cross-tab grid.
type CrossViewData struct {
	Columns []*HeaderDescriptor
	Rows    []*HeaderDescriptor
	Cells   []*CellDescriptor
}

// AddColumn appends a column header descriptor.
func (d *CrossViewData) AddColumn(h *HeaderDescriptor) { d.Columns = append(d.Columns, h) }

// AddRow appends a row header descriptor.
func (d *CrossViewData) AddRow(h *HeaderDescriptor) { d.Rows = append(d.Rows, h) }

// AddCell appends a cell descriptor.
func (d *CrossViewData) AddCell(c *CellDescriptor) { d.Cells = append(d.Cells, c) }

// Clear removes all descriptors.
func (d *CrossViewData) Clear() {
	d.Columns = d.Columns[:0]
	d.Rows = d.Rows[:0]
	d.Cells = d.Cells[:0]
}

// ── ResultCell ────────────────────────────────────────────────────────────────

// ResultCell is a single cell in the computed cross-tab grid.
type ResultCell struct {
	Text       string
	IsHeader   bool
	IsRowLabel bool // true for Y-axis label cells
	IsColLabel bool // true for X-axis label cells
	ColSpan    int  // ≥1
	RowSpan    int  // ≥1
}

// ResultGrid is the computed cross-tab grid (rows of cells).
type ResultGrid struct {
	Rows [][]ResultCell
	// Widths are the column widths in pixels (one per grid column).
	ColCount int
	RowCount int
}

// Cell returns the cell at (row, col), or a zero ResultCell if out of range.
func (g *ResultGrid) Cell(row, col int) ResultCell {
	if row < 0 || row >= len(g.Rows) || col < 0 || col >= len(g.Rows[row]) {
		return ResultCell{}
	}
	return g.Rows[row][col]
}

// ── CrossViewObject ───────────────────────────────────────────────────────────

// CrossViewObject is the main cross-tab object.
// It combines a CubeSourceBase data source with descriptor metadata and produces
// a ResultGrid for rendering.
type CrossViewObject struct {
	Name   string
	Source CubeSourceBase

	Data CrossViewData

	// Display options.
	ShowTitle              bool
	ShowXAxisFieldsCaption bool
	ShowYAxisFieldsCaption bool
}

// NewCrossViewObject creates a CrossViewObject with defaults.
func NewCrossViewObject() *CrossViewObject {
	return &CrossViewObject{
		ShowTitle:              true,
		ShowXAxisFieldsCaption: true,
		ShowYAxisFieldsCaption: true,
	}
}

// SetSource binds a CubeSourceBase and rebuilds the descriptor model.
func (cv *CrossViewObject) SetSource(src CubeSourceBase) {
	cv.Source = src
	cv.buildDescriptors()
}

// buildDescriptors rebuilds the descriptor model from the current source.
func (cv *CrossViewObject) buildDescriptors() {
	cv.Data.Clear()
	if cv.Source == nil {
		return
	}
	src := cv.Source

	// Column (X-axis) header descriptors — one per X-axis field.
	for i := 0; i < src.XAxisFieldsCount(); i++ {
		cv.Data.AddColumn(&HeaderDescriptor{
			FieldName: src.GetXAxisFieldName(i),
			Level:     i,
		})
	}

	// Row (Y-axis) header descriptors — one per Y-axis field.
	for i := 0; i < src.YAxisFieldsCount(); i++ {
		cv.Data.AddRow(&HeaderDescriptor{
			FieldName: src.GetYAxisFieldName(i),
			Level:     i,
		})
	}

	// Cell descriptors — one per combination of terminal column × terminal row.
	// For simplicity we create one per measure per data cell.
	for j := 0; j < src.MeasuresCount(); j++ {
		cv.Data.AddCell(&CellDescriptor{
			MeasureName: src.GetMeasureName(j),
		})
	}
}

// Build computes the ResultGrid by traversing the source axes and filling cells.
// Returns an error if no source is set.
func (cv *CrossViewObject) Build() (*ResultGrid, error) {
	if cv.Source == nil {
		return nil, fmt.Errorf("crossview: no CubeSource set on %q", cv.Name)
	}
	return buildGrid(cv)
}

// ── Grid builder ──────────────────────────────────────────────────────────────

// buildGrid constructs the full ResultGrid from a CrossViewObject.
func buildGrid(cv *CrossViewObject) (*ResultGrid, error) {
	src := cv.Source

	xFieldCount := src.XAxisFieldsCount()
	yFieldCount := src.YAxisFieldsCount()
	dataCols := src.DataColumnCount()
	dataRows := src.DataRowCount()

	// Grid dimensions:
	//  cols = yFieldCount (Y-axis labels) + dataCols
	//  rows = xFieldCount (X-axis header rows) + dataRows
	// An optional title row at the very top is not included here for simplicity.
	totalCols := yFieldCount + dataCols
	totalRows := xFieldCount + dataRows

	if totalCols <= 0 {
		totalCols = 1
	}
	if totalRows <= 0 {
		totalRows = 1
	}

	grid := &ResultGrid{
		ColCount: totalCols,
		RowCount: totalRows,
		Rows:     make([][]ResultCell, totalRows),
	}
	for i := range grid.Rows {
		grid.Rows[i] = make([]ResultCell, totalCols)
		for j := range grid.Rows[i] {
			grid.Rows[i][j] = ResultCell{ColSpan: 1, RowSpan: 1}
		}
	}

	// Fill X-axis header rows (top-left corner shows Y-axis field names).
	for i := 0; i < xFieldCount && i < totalRows; i++ {
		// Corner: Y-axis field labels (if ShowYAxisFieldsCaption).
		for j := 0; j < yFieldCount && j < totalCols; j++ {
			if cv.ShowXAxisFieldsCaption && j < len(cv.Data.Rows) {
				grid.Rows[i][j] = ResultCell{
					Text:     cv.Data.Rows[j].FieldName,
					IsHeader: true,
					ColSpan:  1,
					RowSpan:  1,
				}
			}
		}
	}

	// Fill X-axis column headers via TraverseXAxis.
	src.TraverseXAxis(func(ac AxisDrawCell) {
		row := ac.Level
		col := yFieldCount + ac.Cell
		if row < totalRows && col < totalCols {
			grid.Rows[row][col] = ResultCell{
				Text:       ac.Text,
				IsHeader:   true,
				IsColLabel: true,
				ColSpan:    ac.SizeCell,
				RowSpan:    ac.SizeLevel,
			}
		}
	})

	// Fill Y-axis row headers via TraverseYAxis.
	src.TraverseYAxis(func(ac AxisDrawCell) {
		row := xFieldCount + ac.Cell
		col := ac.Level
		if row < totalRows && col < totalCols {
			grid.Rows[row][col] = ResultCell{
				Text:       ac.Text,
				IsHeader:   true,
				IsRowLabel: true,
				ColSpan:    ac.SizeLevel,
				RowSpan:    ac.SizeCell,
			}
		}
	})

	// Fill data cells.
	for r := 0; r < dataRows; r++ {
		for c := 0; c < dataCols; c++ {
			mc := src.GetMeasureCell(c, r)
			gridRow := xFieldCount + r
			gridCol := yFieldCount + c
			if gridRow < totalRows && gridCol < totalCols {
				grid.Rows[gridRow][gridCol] = ResultCell{
					Text:    mc.Text,
					ColSpan: 1,
					RowSpan: 1,
				}
			}
		}
	}

	return grid, nil
}
