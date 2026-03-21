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

	// Measures placement. MeasuresInXAxis returns true when the measure
	// headers should appear as an extra level on the X (column) axis; false
	// means they go on the Y (row) axis.
	// MeasuresInYAxis is the inverse convenience accessor.
	// MeasuresLevel returns the nesting depth at which measure headers are
	// inserted (-1 means innermost / deepest level).
	MeasuresInXAxis() bool
	MeasuresInYAxis() bool
	MeasuresLevel() int
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

// GetName returns the display name for this header descriptor.
// Mirrors CrossViewHeaderDescriptor.GetName() in CrossViewHeaderDescriptor.cs.
func (h *HeaderDescriptor) GetName() string {
	if h.IsGrandTotal {
		return "GrandTotal"
	}
	if h.IsMeasure {
		return h.MeasureName
	}
	if h.IsTotal {
		return "Total of " + h.FieldName
	}
	return h.FieldName
}

// Assign copies all fields from src into h.
// Mirrors CrossViewHeaderDescriptor.Assign() in CrossViewHeaderDescriptor.cs.
func (h *HeaderDescriptor) Assign(src *HeaderDescriptor) {
	if src == nil {
		return
	}
	h.Expression = src.Expression
	h.FieldName = src.FieldName
	h.MeasureName = src.MeasureName
	h.IsGrandTotal = src.IsGrandTotal
	h.IsTotal = src.IsTotal
	h.IsMeasure = src.IsMeasure
	h.Level = src.Level
	h.LevelSize = src.LevelSize
	h.Cell = src.Cell
	h.CellSize = src.CellSize
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

// Assign copies all fields from src into c.
// Mirrors CrossViewCellDescriptor.Assign() in CrossViewCellDescriptor.cs.
func (c *CellDescriptor) Assign(src *CellDescriptor) {
	if src == nil {
		return
	}
	c.Expression = src.Expression
	c.XFieldName = src.XFieldName
	c.YFieldName = src.YFieldName
	c.MeasureName = src.MeasureName
	c.IsXTotal = src.IsXTotal
	c.IsYTotal = src.IsYTotal
	c.IsXGrandTotal = src.IsXGrandTotal
	c.IsYGrandTotal = src.IsYGrandTotal
	c.X = src.X
	c.Y = src.Y
}

// ── CrossViewData ─────────────────────────────────────────────────────────────

// CrossViewData holds the descriptor model for the cross-tab grid.
type CrossViewData struct {
	Columns []*HeaderDescriptor
	Rows    []*HeaderDescriptor
	Cells   []*CellDescriptor

	// columnTerminalIndexes holds indices into Columns for leaf-level descriptors
	// (the descriptors that map directly to data-grid columns).
	columnTerminalIndexes []int
	// rowTerminalIndexes holds indices into Rows for leaf-level descriptors.
	rowTerminalIndexes []int
}

// AddColumn appends a column header descriptor.
func (d *CrossViewData) AddColumn(h *HeaderDescriptor) { d.Columns = append(d.Columns, h) }

// AddRow appends a row header descriptor.
func (d *CrossViewData) AddRow(h *HeaderDescriptor) { d.Rows = append(d.Rows, h) }

// AddCell appends a cell descriptor.
func (d *CrossViewData) AddCell(c *CellDescriptor) { d.Cells = append(d.Cells, c) }

// ColumnTerminalIndexes returns the indices of leaf column descriptors.
func (d *CrossViewData) ColumnTerminalIndexes() []int { return d.columnTerminalIndexes }

// RowTerminalIndexes returns the indices of leaf row descriptors.
func (d *CrossViewData) RowTerminalIndexes() []int { return d.rowTerminalIndexes }

// Clear removes all descriptors.
func (d *CrossViewData) Clear() {
	d.Columns = d.Columns[:0]
	d.Rows = d.Rows[:0]
	d.Cells = d.Cells[:0]
	d.columnTerminalIndexes = d.columnTerminalIndexes[:0]
	d.rowTerminalIndexes = d.rowTerminalIndexes[:0]
}

// CreateDescriptors builds the descriptor model from a CubeSourceBase.
//
// Column descriptors: one per X-axis field level. When MeasuresCount > 1 and
// MeasuresInXAxis is true, a measure-level descriptor is inserted at the
// MeasuresLevel position.
//
// Row descriptors: analogous for Y-axis / MeasuresInYAxis.
//
// Cell descriptors: one per unique (terminal-column × terminal-row) combination,
// matching the data grid cell at (x, y).
//
// Terminal indexes are computed after descriptors are created.
func (d *CrossViewData) CreateDescriptors(src CubeSourceBase) {
	d.Clear()

	xFields := src.XAxisFieldsCount()
	yFields := src.YAxisFieldsCount()
	measures := src.MeasuresCount()
	measuresInX := src.MeasuresInXAxis() && measures > 1
	measuresInY := src.MeasuresInYAxis() && measures > 1
	measLevel := src.MeasuresLevel()

	// ── Column (X-axis) descriptors ───────────────────────────────────────────
	// We create one descriptor per field level. If measures belong to the X
	// axis, we insert an extra level at measLevel (or innermost if -1).
	xLevels := xFields
	if measuresInX {
		xLevels++ // extra level for measures
	}
	xMeasureLevel := measLevel
	if measuresInX && xMeasureLevel < 0 {
		xMeasureLevel = xLevels - 1 // innermost
	}

	xFieldIdx := 0
	for level := 0; level < xLevels; level++ {
		if measuresInX && level == xMeasureLevel {
			// Insert measure-level descriptors.
			for j := 0; j < measures; j++ {
				hd := &HeaderDescriptor{
					MeasureName: src.GetMeasureName(j),
					IsMeasure:   true,
					Level:       level,
				}
				d.columnTerminalIndexes = append(d.columnTerminalIndexes, len(d.Columns))
				d.AddColumn(hd)
			}
		} else {
			if xFieldIdx < xFields {
				hd := &HeaderDescriptor{
					FieldName: src.GetXAxisFieldName(xFieldIdx),
					Level:     level,
				}
				// If this is the deepest non-measure field and no measure level,
				// mark as terminal.
				if !measuresInX && level == xFields-1 {
					d.columnTerminalIndexes = append(d.columnTerminalIndexes, len(d.Columns))
				}
				d.AddColumn(hd)
				xFieldIdx++
			}
		}
	}

	// ── Row (Y-axis) descriptors ───────────────────────────────────────────────
	yLevels := yFields
	if measuresInY {
		yLevels++
	}
	yMeasureLevel := measLevel
	if measuresInY && yMeasureLevel < 0 {
		yMeasureLevel = yLevels - 1
	}

	yFieldIdx := 0
	for level := 0; level < yLevels; level++ {
		if measuresInY && level == yMeasureLevel {
			for j := 0; j < measures; j++ {
				hd := &HeaderDescriptor{
					MeasureName: src.GetMeasureName(j),
					IsMeasure:   true,
					Level:       level,
				}
				d.rowTerminalIndexes = append(d.rowTerminalIndexes, len(d.Rows))
				d.AddRow(hd)
			}
		} else {
			if yFieldIdx < yFields {
				hd := &HeaderDescriptor{
					FieldName: src.GetYAxisFieldName(yFieldIdx),
					Level:     level,
				}
				if !measuresInY && level == yFields-1 {
					d.rowTerminalIndexes = append(d.rowTerminalIndexes, len(d.Rows))
				}
				d.AddRow(hd)
				yFieldIdx++
			}
		}
	}

	// ── Cell descriptors ──────────────────────────────────────────────────────
	// One cell descriptor per data-grid position (data cols × data rows).
	// When multiple measures exist on an axis, the data-col/row count already
	// includes the measure dimension via DataColumnCount / DataRowCount.
	dataCols := src.DataColumnCount()
	dataRows := src.DataRowCount()
	for y := 0; y < dataRows; y++ {
		for x := 0; x < dataCols; x++ {
			cd := &CellDescriptor{
				X: x,
				Y: y,
			}
			// Attach measure name when a single measure axis exists.
			if measures == 1 {
				cd.MeasureName = src.GetMeasureName(0)
			}
			d.AddCell(cd)
		}
	}
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
	if cv.Source == nil {
		cv.Data.Clear()
		return
	}
	cv.Data.CreateDescriptors(cv.Source)
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

	dataCols := src.DataColumnCount()
	dataRows := src.DataRowCount()

	// Number of header rows = X-axis field levels.
	// When MeasuresInXAxis, there is an extra level for measures.
	xHeaderRows := src.XAxisFieldsCount()
	if src.MeasuresInXAxis() && src.MeasuresCount() > 1 {
		xHeaderRows++
	}

	// Number of header columns = Y-axis field levels.
	// When MeasuresInYAxis, there is an extra level for measures.
	yHeaderCols := src.YAxisFieldsCount()
	if src.MeasuresInYAxis() && src.MeasuresCount() > 1 {
		yHeaderCols++
	}

	// Grid dimensions:
	//  cols = yHeaderCols (Y-axis labels) + dataCols
	//  rows = xHeaderRows (X-axis header rows) + dataRows
	totalCols := yHeaderCols + dataCols
	totalRows := xHeaderRows + dataRows

	// Alias for the rest of the function (replaces old xFieldCount/yFieldCount).
	xFieldCount := xHeaderRows
	yFieldCount := yHeaderCols

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
