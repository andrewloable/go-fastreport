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
// Mirrors CrossViewAxisDrawCell (BaseCubeLink.cs).
type AxisDrawCell struct {
	Text         string
	Cell         int // position along the data axis (0-based data column or row index)
	Level        int // nesting depth (0 = outermost)
	SizeCell     int // span in the data direction
	SizeLevel    int // span in the level direction
	MeasureIndex int // index of the measure for IsMeasure cells (0 for field-level cells)
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
// It mirrors FastReport.CrossView.CrossViewData (CrossViewData.cs).
type CrossViewData struct {
	Columns []*HeaderDescriptor
	Rows    []*HeaderDescriptor
	Cells   []*CellDescriptor

	// cubeSource is the bound data source; nil when no source is assigned.
	// Mirrors the private cubeSource field in CrossViewData.cs.
	cubeSource CubeSourceBase

	// columnTerminalIndexes holds indices into Columns for leaf-level descriptors
	// (the descriptors that map directly to data-grid columns).
	// Mirrors CrossViewData.columnTerminalIndexes (CrossViewData.cs line 23).
	columnTerminalIndexes []int
	// rowTerminalIndexes holds indices into Rows for leaf-level descriptors.
	// Mirrors CrossViewData.rowTerminalIndexes (CrossViewData.cs line 24).
	rowTerminalIndexes []int

	// rowDescriptorsIndexes holds indices into Columns for each X-axis level.
	// Used by GetRowDescriptor to find the column header at a given X-axis level.
	// Mirrors CrossViewData.rowDescriptorsIndexes (CrossViewData.cs line 21).
	rowDescriptorsIndexes []int
	// columnDescriptorsIndexes holds indices into Rows for each Y-axis level.
	// Used by GetColumnDescriptor to find the row header at a given Y-axis level.
	// Mirrors CrossViewData.columnDescriptorsIndexes (CrossViewData.cs line 22).
	columnDescriptorsIndexes []int
}

// ── CrossViewData FastCube convenience properties ──────────────────────────────
// These delegate to the bound CubeSourceBase, returning zero values when no
// source is assigned. They mirror the FastCube properties in CrossViewData.cs.

// SourceAssigned returns true when a CubeSourceBase has been bound.
// Mirrors CrossViewData.SourceAssigned (CrossViewData.cs line 63).
func (d *CrossViewData) SourceAssigned() bool { return d.cubeSource != nil }

// XAxisFieldsCount returns the number of X-axis fields from the bound source,
// or 0 when no source is assigned.
// Mirrors CrossViewData.XAxisFieldsCount (CrossViewData.cs line 31).
func (d *CrossViewData) XAxisFieldsCount() int {
	if d.cubeSource == nil {
		return 0
	}
	return d.cubeSource.XAxisFieldsCount()
}

// YAxisFieldsCount returns the number of Y-axis fields from the bound source,
// or 0 when no source is assigned.
// Mirrors CrossViewData.YAxisFieldsCount (CrossViewData.cs line 35).
func (d *CrossViewData) YAxisFieldsCount() int {
	if d.cubeSource == nil {
		return 0
	}
	return d.cubeSource.YAxisFieldsCount()
}

// MeasuresCount returns the number of measures from the bound source,
// or 0 when no source is assigned.
// Mirrors CrossViewData.MeasuresCount (CrossViewData.cs line 39).
func (d *CrossViewData) MeasuresCount() int {
	if d.cubeSource == nil {
		return 0
	}
	return d.cubeSource.MeasuresCount()
}

// MeasuresLevel returns the measures nesting level from the bound source,
// or 0 when no source is assigned.
// Mirrors CrossViewData.MeasuresLevel (CrossViewData.cs line 43).
func (d *CrossViewData) MeasuresLevel() int {
	if d.cubeSource == nil {
		return 0
	}
	return d.cubeSource.MeasuresLevel()
}

// MeasuresInXAxis returns whether measures appear on the X axis,
// or false when no source is assigned.
// Mirrors CrossViewData.MeasuresInXAxis (CrossViewData.cs line 47).
func (d *CrossViewData) MeasuresInXAxis() bool {
	if d.cubeSource == nil {
		return false
	}
	return d.cubeSource.MeasuresInXAxis()
}

// MeasuresInYAxis returns whether measures appear on the Y axis,
// or false when no source is assigned.
// Mirrors CrossViewData.MeasuresInYAxis (CrossViewData.cs line 51).
func (d *CrossViewData) MeasuresInYAxis() bool {
	if d.cubeSource == nil {
		return false
	}
	return d.cubeSource.MeasuresInYAxis()
}

// DataColumnCount returns the number of data columns from the bound source,
// or 0 when no source is assigned.
// Mirrors CrossViewData.DataColumnCount (CrossViewData.cs line 55).
func (d *CrossViewData) DataColumnCount() int {
	if d.cubeSource == nil {
		return 0
	}
	return d.cubeSource.DataColumnCount()
}

// DataRowCount returns the number of data rows from the bound source,
// or 0 when no source is assigned.
// Mirrors CrossViewData.DataRowCount (CrossViewData.cs line 59).
func (d *CrossViewData) DataRowCount() int {
	if d.cubeSource == nil {
		return 0
	}
	return d.cubeSource.DataRowCount()
}

// SetCubeSource binds or unbinds a CubeSourceBase.
// Mirrors the internal CubeSource setter in CrossViewData.cs (line 139-148).
func (d *CrossViewData) SetCubeSource(src CubeSourceBase) {
	if d.cubeSource != src {
		d.cubeSource = src
	}
}

// CubeSource returns the currently bound CubeSourceBase (may be nil).
func (d *CrossViewData) CubeSource() CubeSourceBase { return d.cubeSource }

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

// RowDescriptorsIndexes returns the indices into Columns for each X-axis level.
// Mirrors CrossViewData.rowDescriptorsIndexes (CrossViewData.cs line 21).
func (d *CrossViewData) RowDescriptorsIndexes() []int { return d.rowDescriptorsIndexes }

// ColumnDescriptorsIndexes returns the indices into Rows for each Y-axis level.
// Mirrors CrossViewData.columnDescriptorsIndexes (CrossViewData.cs line 22).
func (d *CrossViewData) ColumnDescriptorsIndexes() []int { return d.columnDescriptorsIndexes }

// ColumnDescriptorsIndexesStr returns columnDescriptorsIndexes as a comma-separated string.
// Mirrors CrossViewData.ColumnDescriptorsIndexes getter (CrossViewData.cs lines 90-98).
func (d *CrossViewData) ColumnDescriptorsIndexesStr() string {
	return FormatIndexArray(d.columnDescriptorsIndexes)
}

// SetColumnDescriptorsIndexesStr parses a comma-separated string into columnDescriptorsIndexes.
// Mirrors CrossViewData.ColumnDescriptorsIndexes setter (CrossViewData.cs lines 90-98).
func (d *CrossViewData) SetColumnDescriptorsIndexesStr(s string) {
	d.columnDescriptorsIndexes = ParseIndexArray(s)
}

// RowDescriptorsIndexesStr returns rowDescriptorsIndexes as a comma-separated string.
// Mirrors CrossViewData.RowDescriptorsIndexes getter (CrossViewData.cs lines 103-112).
func (d *CrossViewData) RowDescriptorsIndexesStr() string {
	return FormatIndexArray(d.rowDescriptorsIndexes)
}

// SetRowDescriptorsIndexesStr parses a comma-separated string into rowDescriptorsIndexes.
// Mirrors CrossViewData.RowDescriptorsIndexes setter (CrossViewData.cs lines 103-112).
func (d *CrossViewData) SetRowDescriptorsIndexesStr(s string) {
	d.rowDescriptorsIndexes = ParseIndexArray(s)
}

// ColumnTerminalIndexesStr returns columnTerminalIndexes as a comma-separated string.
// Mirrors CrossViewData.ColumnTerminalIndexes getter (CrossViewData.cs lines 116-124).
func (d *CrossViewData) ColumnTerminalIndexesStr() string {
	return FormatIndexArray(d.columnTerminalIndexes)
}

// SetColumnTerminalIndexesStr parses a comma-separated string into columnTerminalIndexes.
// Mirrors CrossViewData.ColumnTerminalIndexes setter (CrossViewData.cs lines 116-124).
func (d *CrossViewData) SetColumnTerminalIndexesStr(s string) {
	d.columnTerminalIndexes = ParseIndexArray(s)
}

// RowTerminalIndexesStr returns rowTerminalIndexes as a comma-separated string.
// Mirrors CrossViewData.RowTerminalIndexes getter (CrossViewData.cs lines 129-137).
func (d *CrossViewData) RowTerminalIndexesStr() string {
	return FormatIndexArray(d.rowTerminalIndexes)
}

// SetRowTerminalIndexesStr parses a comma-separated string into rowTerminalIndexes.
// Mirrors CrossViewData.RowTerminalIndexes setter (CrossViewData.cs lines 129-137).
func (d *CrossViewData) SetRowTerminalIndexesStr(s string) {
	d.rowTerminalIndexes = ParseIndexArray(s)
}

// GetRowDescriptor returns the header descriptor at the given display-row index.
// For index < XAxisFieldsCount: returns the column header at rowDescriptorsIndexes[index].
// For index >= XAxisFieldsCount: returns the row descriptor at rowTerminalIndexes[index - XAxisFieldsCount].
// Mirrors CrossViewData.GetRowDescriptor() (CrossViewData.cs lines 488-499).
func (d *CrossViewData) GetRowDescriptor(index int) *HeaderDescriptor {
	xCount := 1
	if d.SourceAssigned() {
		xCount = d.XAxisFieldsCount()
	}
	if index < xCount {
		if index < len(d.rowDescriptorsIndexes) {
			idx := d.rowDescriptorsIndexes[index]
			if idx >= 0 && idx < len(d.Columns) {
				return d.Columns[idx]
			}
		}
		return nil
	}
	rowIdx := index - xCount
	if rowIdx < len(d.rowTerminalIndexes) {
		idx := d.rowTerminalIndexes[rowIdx]
		if idx >= 0 && idx < len(d.Rows) {
			return d.Rows[idx]
		}
	}
	return nil
}

// GetColumnDescriptor returns the header descriptor at the given display-column index.
// For index < YAxisFieldsCount: returns the row header at columnDescriptorsIndexes[index].
// For index >= YAxisFieldsCount: returns the column descriptor at columnTerminalIndexes[index - YAxisFieldsCount].
// Mirrors CrossViewData.GetColumnDescriptor() (CrossViewData.cs lines 501-512).
func (d *CrossViewData) GetColumnDescriptor(index int) *HeaderDescriptor {
	yCount := 1
	if d.SourceAssigned() {
		yCount = d.YAxisFieldsCount()
	}
	if index < yCount {
		if index < len(d.columnDescriptorsIndexes) {
			idx := d.columnDescriptorsIndexes[index]
			if idx >= 0 && idx < len(d.Rows) {
				return d.Rows[idx]
			}
		}
		return nil
	}
	colIdx := index - yCount
	if colIdx < len(d.columnTerminalIndexes) {
		idx := d.columnTerminalIndexes[colIdx]
		if idx >= 0 && idx < len(d.Columns) {
			return d.Columns[idx]
		}
	}
	return nil
}

// Clear removes all descriptors.
func (d *CrossViewData) Clear() {
	d.Columns = d.Columns[:0]
	d.Rows = d.Rows[:0]
	d.Cells = d.Cells[:0]
	d.columnTerminalIndexes = d.columnTerminalIndexes[:0]
	d.rowTerminalIndexes = d.rowTerminalIndexes[:0]
	d.rowDescriptorsIndexes = d.rowDescriptorsIndexes[:0]
	d.columnDescriptorsIndexes = d.columnDescriptorsIndexes[:0]
}

// CreateDescriptorsFromSource rebuilds descriptors using the bound cubeSource.
// It is a no-op when no source has been assigned (mirrors the C# guard
// "if (!SourceAssigned) return;" in CrossViewData.CreateDescriptors()).
// Mirrors C# CrossViewData.CreateDescriptors() (CrossViewData.cs line 150).
func (d *CrossViewData) CreateDescriptorsFromSource() {
	if d.cubeSource == nil {
		d.Clear()
		return
	}
	d.CreateDescriptors(d.cubeSource)
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
				// Record the column index for this X-axis field (mirrors C# CrossViewData.cs
				// rowDescriptorsIndexes population inside CreateDescriptors).
				d.rowDescriptorsIndexes = append(d.rowDescriptorsIndexes, len(d.Columns))
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
				// Record the row index for this Y-axis field (mirrors C# CrossViewData.cs
				// columnDescriptorsIndexes population inside CreateDescriptors).
				d.columnDescriptorsIndexes = append(d.columnDescriptorsIndexes, len(d.Rows))
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
// Mirrors FastReport.CrossView.CrossViewObject (CrossViewObject.cs).
type CrossViewObject struct {
	Name   string
	Source CubeSourceBase

	Data CrossViewData

	// Display options.
	ShowTitle              bool
	ShowXAxisFieldsCaption bool
	ShowYAxisFieldsCaption bool

	// Style is the name of the matrix style applied to this CrossView.
	// Mirrors CrossViewObject.Style (CrossViewObject.cs line 22 / line 100-107).
	Style string

	// ModifyResultEvent holds the name of the script method to invoke when
	// the result table is ready. Mirrors CrossViewObject.ModifyResultEvent
	// (CrossViewObject.cs lines 117-121).
	ModifyResultEvent string

	// ModifyResultHandler is an optional Go callback fired when the result
	// table is ready. It is the Go equivalent of the C# ModifyResult event
	// (CrossViewObject.cs line 34).
	ModifyResultHandler func(*CrossViewObject)

	// helper holds the layout helper used by BuildTemplate.
	// Mirrors CrossViewObject.helper (CrossViewObject.cs line 25 / line 501).
	helper *CrossViewHelper
}

// NewCrossViewObject creates a CrossViewObject with defaults.
// Mirrors the CrossViewObject constructor (CrossViewObject.cs lines 495-509).
func NewCrossViewObject() *CrossViewObject {
	cv := &CrossViewObject{
		ShowTitle:              true,
		ShowXAxisFieldsCaption: true,
		ShowYAxisFieldsCaption: true,
	}
	cv.helper = NewCrossViewHelper(cv)
	return cv
}

// BuildTemplate creates or updates the matrix template.
// Call this method after modifying the Data descriptor collections.
// Mirrors CrossViewObject.BuildTemplate() (CrossViewObject.cs lines 401-404).
func (cv *CrossViewObject) BuildTemplate() {
	if cv.helper == nil {
		cv.helper = NewCrossViewHelper(cv)
	}
	cv.helper.BuildTemplate()
}

// SetSource binds a CubeSourceBase and rebuilds the descriptor model.
// It also updates cv.Data so that its FastCube convenience properties work.
// Mirrors CrossViewObject.CubeSource setter (CrossViewObject.cs lines 165-191).
func (cv *CrossViewObject) SetSource(src CubeSourceBase) {
	cv.Source = src
	cv.Data.SetCubeSource(src)
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

// OnModifyResult fires the ModifyResultHandler callback.
// Mirrors CrossViewObject.OnModifyResult(EventArgs e) (CrossViewObject.cs line 483-488).
func (cv *CrossViewObject) OnModifyResult() {
	if cv.ModifyResultHandler != nil {
		cv.ModifyResultHandler(cv)
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
