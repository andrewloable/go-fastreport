// Package matrix implements the pivot/matrix table component for go-fastreport.
// It is the Go equivalent of FastReport.Matrix.MatrixObject.
package matrix

import (
	"fmt"
	"strconv"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/table"
)

// ── Enumerations ──────────────────────────────────────────────────────────────

// AggregateFunction determines how cell values are aggregated.
// Matches FastReport.Matrix.MatrixAggregateFunction.
type AggregateFunction int

const (
	AggregateFunctionNone          AggregateFunction = iota
	AggregateFunctionSum                              // Sum all values.
	AggregateFunctionMin                              // Minimum value.
	AggregateFunctionMax                              // Maximum value.
	AggregateFunctionAvg                              // Average value.
	AggregateFunctionCount                            // Count of rows.
	AggregateFunctionCountDistinct                    // Count distinct values.
	AggregateFunctionCustom                           // Custom (script-driven) aggregation.
)

// MatrixPercent controls how a cell value is expressed as a percentage.
// Matches FastReport.Matrix.MatrixPercent.
type MatrixPercent int

const (
	MatrixPercentNone        MatrixPercent = iota // Raw value (no percentage).
	MatrixPercentColumnTotal                      // As % of column total.
	MatrixPercentRowTotal                         // As % of row total.
	MatrixPercentGrandTotal                       // As % of grand total.
)

// SortOrder controls sort direction for header descriptors.
// Matches FastReport.SortOrder.
type SortOrder int

const (
	SortOrderNone       SortOrder = iota
	SortOrderAscending            // Sort ascending (default).
	SortOrderDescending           // Sort descending.
)

// MatrixEvenStylePriority controls which axis the even/odd style alternates on.
// Matches FastReport.Matrix.MatrixEvenStylePriority.
type MatrixEvenStylePriority int

const (
	MatrixEvenStylePriorityRows    MatrixEvenStylePriority = iota // alternate on rows (default)
	MatrixEvenStylePriorityColumns                                // alternate on columns
)

// EvenStylePriority* aliases for test and user convenience.
const (
	EvenStylePriorityRows    = MatrixEvenStylePriorityRows
	EvenStylePriorityColumns = MatrixEvenStylePriorityColumns
)

// ── Descriptors ───────────────────────────────────────────────────────────────

// Descriptor is the base for column/row/cell descriptors.
// Matches FastReport.Matrix.MatrixDescriptor.
type Descriptor struct {
	// Expression is a report expression (e.g. "[DataSource.Field]").
	Expression string
}

// HeaderDescriptor describes a column or row header group.
// Matches FastReport.Matrix.MatrixHeaderDescriptor.
type HeaderDescriptor struct {
	Descriptor

	// Sort controls sort direction (default: Ascending).
	Sort SortOrder
	// Totals adds a total column/row for this header level (default: true).
	Totals bool
	// TotalsFirst places the total before data rows/columns (default: false).
	TotalsFirst bool
	// PageBreak inserts a page break after each group value (default: false).
	PageBreak bool
	// SuppressTotals suppresses the total when there is only one value (default: false).
	SuppressTotals bool
	// TotalText is the label shown in the totals row/column (default: "Total").
	TotalText string
}

// NewHeaderDescriptor creates a HeaderDescriptor with the given expression and C# defaults.
func NewHeaderDescriptor(expr string) *HeaderDescriptor {
	return &HeaderDescriptor{
		Descriptor: Descriptor{Expression: expr},
		Sort:       SortOrderAscending,
		Totals:     true,
		TotalText:  "Total",
	}
}

// CellDescriptor describes a data cell with an aggregate function.
// Matches FastReport.Matrix.MatrixCellDescriptor.
type CellDescriptor struct {
	Descriptor

	// Function is the aggregate function applied to the cell values (default: Sum).
	Function AggregateFunction
	// Percent controls percentage display (default: None).
	Percent MatrixPercent
}

// NewCellDescriptor creates a CellDescriptor with the given expression and function.
func NewCellDescriptor(expr string, fn AggregateFunction) *CellDescriptor {
	return &CellDescriptor{
		Descriptor: Descriptor{Expression: expr},
		Function:   fn,
	}
}

// ── MatrixData ────────────────────────────────────────────────────────────────

// MatrixData holds the descriptor collections for columns, rows and cells.
// Matches FastReport.Matrix.MatrixData (container only, not serialized directly).
type MatrixData struct {
	// Columns are the column header descriptors (from outermost to innermost).
	Columns []*HeaderDescriptor
	// Rows are the row header descriptors (from outermost to innermost).
	Rows []*HeaderDescriptor
	// Cells are the data cell descriptors.
	Cells []*CellDescriptor
	// rt holds the runtime header trees and cell store (lazily initialized).
	rt *matrixDataRuntime
}

// AddColumn appends a column header descriptor.
func (d *MatrixData) AddColumn(h *HeaderDescriptor) { d.Columns = append(d.Columns, h) }

// AddRow appends a row header descriptor.
func (d *MatrixData) AddRow(h *HeaderDescriptor) { d.Rows = append(d.Rows, h) }

// AddCell appends a cell descriptor.
func (d *MatrixData) AddCell(c *CellDescriptor) { d.Cells = append(d.Cells, c) }

// ── Aggregator ────────────────────────────────────────────────────────────────

// accumulator holds running aggregate state for one cell key.
type accumulator struct {
	fn       AggregateFunction
	sum      float64
	count    int64
	min      float64
	max      float64
	minSet   bool
	maxSet   bool
	distinct map[any]struct{}
}

func newAccumulator(fn AggregateFunction) *accumulator {
	a := &accumulator{fn: fn}
	if fn == AggregateFunctionCountDistinct {
		a.distinct = make(map[any]struct{})
	}
	return a
}

func (a *accumulator) add(v float64, raw any) {
	a.sum += v
	a.count++
	if !a.minSet || v < a.min {
		a.min = v
		a.minSet = true
	}
	if !a.maxSet || v > a.max {
		a.max = v
		a.maxSet = true
	}
	if a.distinct != nil {
		a.distinct[raw] = struct{}{}
	}
}

func (a *accumulator) result() float64 {
	switch a.fn {
	case AggregateFunctionSum:
		return a.sum
	case AggregateFunctionMin:
		return a.min
	case AggregateFunctionMax:
		return a.max
	case AggregateFunctionAvg:
		if a.count == 0 {
			return 0
		}
		return a.sum / float64(a.count)
	case AggregateFunctionCount:
		return float64(a.count)
	case AggregateFunctionCountDistinct:
		return float64(len(a.distinct))
	}
	return 0
}

// cellKey uniquely identifies a matrix cell by (rowKey, colKey, cellIdx).
type cellKey struct {
	row     string
	col     string
	cellIdx int
}

// ── MatrixObject ──────────────────────────────────────────────────────────────

// MatrixObject is a pivot table component.
// It builds its output from descriptor collections connected to a data source.
// It is the Go equivalent of FastReport.Matrix.MatrixObject.
type MatrixObject struct {
	table.TableBase

	// Data holds the descriptor collections.
	Data MatrixData

	// DataSource is the data source to iterate for populating the matrix.
	DataSource data.DataSource

	// DataSourceName is the string name of the bound data source, preserved
	// from the FRX and used to resolve DataSource at engine run time.
	DataSourceName string

	// Filter is a boolean expression for filtering data rows (default: "").
	Filter string

	// AutoSize enables automatic column/row sizing from content (default: true).
	AutoSize bool
	// CellsSideBySide renders multiple cell descriptors side by side (default: false).
	CellsSideBySide bool
	// KeepCellsSideBySide keeps side-by-side cells together on page breaks (default: false).
	KeepCellsSideBySide bool
	// EvenStylePriority controls even-style alternation axis (default: Rows).
	EvenStylePriority MatrixEvenStylePriority
	// Style is the named built-in style to apply (e.g. "Green").
	Style string
	// ShowTitle shows the data source title row (default: false).
	ShowTitle bool
	// SplitRows allows data rows to split across pages (default: false).
	SplitRows bool
	// PrintIfEmpty prints the matrix even when there is no data (default: true).
	PrintIfEmpty bool

	// Event names for script hooks.
	ManualBuildEvent  string
	ModifyResultEvent string
	AfterTotalsEvent  string

	// accumulators holds aggregate state per cell key.
	accumulators map[cellKey]*accumulator
	// rowValues stores unique ordered row header values.
	rowValues []string
	// colValues stores unique ordered col header values.
	colValues []string
	// rowIndex maps row value → index.
	rowIndex map[string]int
	// colIndex maps col value → index.
	colIndex map[string]int

	// Multi-level header support (populated by AddDataMultiLevel).
	rowRoot        *HeaderItem
	colRoot        *HeaderItem
	mlAccumulators map[multiLevelKey]*accumulator

	// Runtime state for engine lifecycle (matrix_lifecycle.go).
	// C# source: MatrixObject private fields populated by Helper.AddDataRow.
	ColumnValues []any // current column expression values
	RowValues2   []any // current row expression values (RowValues is reserved)
	CellValues   []any // current cell values
	ColumnIndex  int   // current flat column index
	RowIndex     int   // current flat row index

	// Event callbacks (set by engine or user code before running).
	ManualBuildHandler  func(*MatrixObject) // C#: MatrixObject.ManualBuild event
	ModifyResultHandler func(*MatrixObject) // C#: MatrixObject.ModifyResult event
	AfterTotalsHandler  func(*MatrixObject) // C#: MatrixObject.AfterTotals event

	// StyleLookup resolves named styles (e.g. EvenStyle) during template building.
	// Set by the engine before calling BuildTemplateMultiLevel.
	StyleLookup report.StyleLookup
	// HighlightCalc evaluates a highlight condition expression in the current matrix
	// context (with RowIndex/ColumnIndex available). Set by the engine before calling
	// BuildTemplateMultiLevel.
	// C# ref: templateCell.GetData() evaluates highlight conditions in PrintData.
	HighlightCalc func(expr string) (any, error)
	// CurrentCellValue holds the raw numeric value of the current data cell being
	// processed, so that HighlightCalc can expose it as "Value" in the expression
	// context. Mirrors C# TextObject.Value passed to Report.Calc(expr, varValue).
	CurrentCellValue any
	// cellFormat is applied to data cell values during BuildTemplateMultiLevel.
	// Populated from the template data cell's Format (e.g. Currency).
	cellFormat format.Format
	// savedTemplateRows preserves template rows before BuildTemplateMultiLevel
	// resets the TableBase. Used to extract header labels and formatting.
	savedTemplateRows []*table.TableRow

	// saveVisible holds the pre-run Visible state for RestoreState.
	saveVisible bool
	// visible shadows ComponentBase.visible so SaveState/RestoreState work
	// without depending on the exact embedding depth.
	visible bool
}

// New creates a MatrixObject with defaults matching the C# constructor.
func New() *MatrixObject {
	return &MatrixObject{
		TableBase:    *table.NewTableBase(),
		AutoSize:     true,
		PrintIfEmpty: true,
		accumulators: make(map[cellKey]*accumulator),
		rowIndex:     make(map[string]int),
		colIndex:     make(map[string]int),
	}
}

// TypeName returns "MatrixObject".
func (m *MatrixObject) TypeName() string { return "MatrixObject" }

// AddData processes one logical row from the data source using the descriptors.
// rowExpr is the string value for the row header; colExpr is for the column header;
// values are the raw cell values in the order matching Data.Cells.
//
// For multi-level headers, provide the outermost value as rowExpr/colExpr.
// (This simplified API handles single-level headers; nested headers are future work.)
func (m *MatrixObject) AddData(rowExpr, colExpr string, values []any) {
	// Track unique row/col values in insertion order.
	if _, ok := m.rowIndex[rowExpr]; !ok {
		m.rowIndex[rowExpr] = len(m.rowValues)
		m.rowValues = append(m.rowValues, rowExpr)
	}
	if _, ok := m.colIndex[colExpr]; !ok {
		m.colIndex[colExpr] = len(m.colValues)
		m.colValues = append(m.colValues, colExpr)
	}

	for i, v := range values {
		if i >= len(m.Data.Cells) {
			break
		}
		k := cellKey{row: rowExpr, col: colExpr, cellIdx: i}
		acc, exists := m.accumulators[k]
		if !exists {
			acc = newAccumulator(m.Data.Cells[i].Function)
			m.accumulators[k] = acc
		}
		raw := v
		f := toFloat(v)
		acc.add(f, raw)
	}
}

// CellResult returns the aggregated result for the cell at (rowKey, colKey, cellIdx).
// Returns 0 and an error if no data has been accumulated for that key.
func (m *MatrixObject) CellResult(rowKey, colKey string, cellIdx int) (float64, error) {
	k := cellKey{row: rowKey, col: colKey, cellIdx: cellIdx}
	acc, ok := m.accumulators[k]
	if !ok {
		return 0, fmt.Errorf("matrix: no data for row=%q col=%q cell=%d", rowKey, colKey, cellIdx)
	}
	return acc.result(), nil
}

// RowValues returns unique row header values in insertion order.
func (m *MatrixObject) RowValues() []string { return m.rowValues }

// ColValues returns unique column header values in insertion order.
func (m *MatrixObject) ColValues() []string { return m.colValues }

// BuildTemplate constructs the table skeleton based on current descriptors and data.
// It populates rows/columns from the accumulated rowValues and colValues.
func (m *MatrixObject) BuildTemplate() {
	// Reset table.
	m.TableBase = *table.NewTableBase()

	// Nothing to build if there are no data rows.
	if len(m.rowValues) == 0 && len(m.colValues) == 0 {
		return
	}

	// Add columns: one label column + one per col value.
	m.TableBase.AddColumn(table.NewTableColumn()) // label column
	for range m.colValues {
		m.TableBase.AddColumn(table.NewTableColumn())
	}

	// Header row: corner cell + column header values.
	headerRow := table.NewTableRow()
	corner := table.NewTableCell()
	corner.SetText("") // top-left corner cell
	headerRow.AddCell(corner)
	for _, cv := range m.colValues {
		cell := table.NewTableCell()
		cell.SetText(cv)
		headerRow.AddCell(cell)
	}
	m.TableBase.AddRow(headerRow)

	// Data rows.
	for _, rv := range m.rowValues {
		row := table.NewTableRow()
		labelCell := table.NewTableCell()
		labelCell.SetText(rv)
		row.AddCell(labelCell)
		for _, cv := range m.colValues {
			dataCell := table.NewTableCell()
			if len(m.Data.Cells) > 0 {
				val, err := m.CellResult(rv, cv, 0)
				if err == nil {
					dataCell.SetText(fmt.Sprintf("%g", val))
				}
			}
			row.AddCell(dataCell)
		}
		m.TableBase.AddRow(row)
	}
}

// ── Descriptor serialization helpers ─────────────────────────────────────────

// headerHolder wraps a HeaderDescriptor slice for writing as a named child block.
type headerHolder struct {
	headers []*HeaderDescriptor
}

func (h *headerHolder) Serialize(w report.Writer) error {
	for _, hd := range h.headers {
		if err := w.WriteObjectNamed("Header", &headerDescriptorWriter{hd}); err != nil {
			return err
		}
	}
	return nil
}

func (h *headerHolder) Deserialize(_ report.Reader) error { return nil }

// headerDescriptorWriter serializes a single HeaderDescriptor.
type headerDescriptorWriter struct{ h *HeaderDescriptor }

func (hw *headerDescriptorWriter) Serialize(w report.Writer) error {
	if hw.h.Expression != "" {
		w.WriteStr("Expression", hw.h.Expression)
	}
	// Sort default is Ascending (1); only write when != Ascending.
	if hw.h.Sort != SortOrderAscending {
		w.WriteInt("Sort", int(hw.h.Sort))
	}
	// Totals default is true; write when false.
	if !hw.h.Totals {
		w.WriteBool("Totals", false)
	}
	if hw.h.TotalsFirst {
		w.WriteBool("TotalsFirst", true)
	}
	if hw.h.PageBreak {
		w.WriteBool("PageBreak", true)
	}
	if hw.h.SuppressTotals {
		w.WriteBool("SuppressTotals", true)
	}
	return nil
}

func (hw *headerDescriptorWriter) Deserialize(_ report.Reader) error { return nil }

// cellHolder wraps a CellDescriptor slice for writing as a named child block.
type cellHolder struct {
	cells []*CellDescriptor
}

func (c *cellHolder) Serialize(w report.Writer) error {
	for _, cd := range c.cells {
		if err := w.WriteObjectNamed("Cell", &cellDescriptorWriter{cd}); err != nil {
			return err
		}
	}
	return nil
}

func (c *cellHolder) Deserialize(_ report.Reader) error { return nil }

// cellDescriptorWriter serializes a single CellDescriptor.
type cellDescriptorWriter struct{ c *CellDescriptor }

func (cw *cellDescriptorWriter) Serialize(w report.Writer) error {
	if cw.c.Expression != "" {
		w.WriteStr("Expression", cw.c.Expression)
	}
	// Function default is Sum (1); write when != Sum.
	if cw.c.Function != AggregateFunctionSum {
		w.WriteInt("Function", int(cw.c.Function))
	}
	if cw.c.Percent != MatrixPercentNone {
		w.WriteInt("Percent", int(cw.c.Percent))
	}
	return nil
}

func (cw *cellDescriptorWriter) Deserialize(_ report.Reader) error { return nil }

// ── MatrixObject Serialize / Deserialize ──────────────────────────────────────

// Serialize writes MatrixObject properties that differ from defaults,
// including MatrixRows, MatrixColumns, and MatrixCells child blocks.
func (m *MatrixObject) Serialize(w report.Writer) error {
	if err := m.TableBase.Serialize(w); err != nil {
		return err
	}
	if m.DataSourceName != "" {
		w.WriteStr("DataSource", m.DataSourceName)
	}
	if m.Filter != "" {
		w.WriteStr("Filter", m.Filter)
	}
	// AutoSize default is true; write when false.
	if !m.AutoSize {
		w.WriteBool("AutoSize", false)
	}
	if m.CellsSideBySide {
		w.WriteBool("CellsSideBySide", true)
	}
	if m.KeepCellsSideBySide {
		w.WriteBool("KeepCellsSideBySide", true)
	}
	if m.EvenStylePriority != MatrixEvenStylePriorityRows {
		w.WriteInt("MatrixEvenStylePriority", int(m.EvenStylePriority))
	}
	if m.Style != "" {
		w.WriteStr("Style", m.Style)
	}
	if m.ShowTitle {
		w.WriteBool("ShowTitle", true)
	}
	if m.SplitRows {
		w.WriteBool("SplitRows", true)
	}
	// PrintIfEmpty default is true; write when false.
	if !m.PrintIfEmpty {
		w.WriteBool("PrintIfEmpty", false)
	}
	if m.ManualBuildEvent != "" {
		w.WriteStr("ManualBuildEvent", m.ManualBuildEvent)
	}
	if m.ModifyResultEvent != "" {
		w.WriteStr("ModifyResultEvent", m.ModifyResultEvent)
	}
	if m.AfterTotalsEvent != "" {
		w.WriteStr("AfterTotalsEvent", m.AfterTotalsEvent)
	}
	// Write descriptor child blocks.
	if len(m.Data.Rows) > 0 {
		if err := w.WriteObjectNamed("MatrixRows", &headerHolder{m.Data.Rows}); err != nil {
			return err
		}
	}
	if len(m.Data.Columns) > 0 {
		if err := w.WriteObjectNamed("MatrixColumns", &headerHolder{m.Data.Columns}); err != nil {
			return err
		}
	}
	if len(m.Data.Cells) > 0 {
		if err := w.WriteObjectNamed("MatrixCells", &cellHolder{m.Data.Cells}); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads MatrixObject properties.
func (m *MatrixObject) Deserialize(r report.Reader) error {
	if err := m.TableBase.Deserialize(r); err != nil {
		return err
	}
	m.DataSourceName = r.ReadStr("DataSource", "")
	m.Filter = r.ReadStr("Filter", "")
	m.AutoSize = r.ReadBool("AutoSize", true)
	m.CellsSideBySide = r.ReadBool("CellsSideBySide", false)
	m.KeepCellsSideBySide = r.ReadBool("KeepCellsSideBySide", false)
	m.EvenStylePriority = MatrixEvenStylePriority(r.ReadInt("MatrixEvenStylePriority", 0))
	m.Style = r.ReadStr("Style", "")
	m.ShowTitle = r.ReadBool("ShowTitle", false)
	m.SplitRows = r.ReadBool("SplitRows", false)
	m.PrintIfEmpty = r.ReadBool("PrintIfEmpty", true)
	m.ManualBuildEvent = r.ReadStr("ManualBuildEvent", "")
	m.ModifyResultEvent = r.ReadStr("ModifyResultEvent", "")
	m.AfterTotalsEvent = r.ReadStr("AfterTotalsEvent", "")
	// Note: cellFormat extraction is deferred to ExtractCellFormat() which
	// is called by the engine after children (template rows/cells) are loaded.
	return nil
}

// DeserializeChild handles matrix-specific child elements (MatrixRows, MatrixColumns, MatrixCells).
func (m *MatrixObject) DeserializeChild(childType string, r report.Reader) bool {
	readHeader := func() *HeaderDescriptor {
		return &HeaderDescriptor{
			Descriptor:     Descriptor{Expression: r.ReadStr("Expression", "")},
			Sort:           SortOrder(r.ReadInt("Sort", int(SortOrderAscending))),
			Totals:         r.ReadBool("Totals", true),
			TotalsFirst:    r.ReadBool("TotalsFirst", false),
			PageBreak:      r.ReadBool("PageBreak", false),
			SuppressTotals: r.ReadBool("SuppressTotals", false),
		}
	}
	switch childType {
	case "MatrixRows":
		for {
			ct, ok := r.NextChild()
			if !ok {
				break
			}
			if ct == "Header" {
				m.Data.Rows = append(m.Data.Rows, readHeader())
			}
			if r.FinishChild() != nil { break }
		}
		return true
	case "MatrixColumns":
		for {
			ct, ok := r.NextChild()
			if !ok {
				break
			}
			if ct == "Header" {
				m.Data.Columns = append(m.Data.Columns, readHeader())
			}
			if r.FinishChild() != nil { break }
		}
		return true
	case "MatrixCells":
		for {
			ct, ok := r.NextChild()
			if !ok {
				break
			}
			if ct == "Cell" {
				cd := &CellDescriptor{
					Descriptor: Descriptor{Expression: r.ReadStr("Expression", "")},
					Function:   AggregateFunction(r.ReadInt("Function", int(AggregateFunctionSum))),
					Percent:    MatrixPercent(r.ReadInt("Percent", 0)),
				}
				m.Data.Cells = append(m.Data.Cells, cd)
			}
			if r.FinishChild() != nil { break }
		}
		return true
	}
	// Delegate to TableBase for TableRow/TableColumn children.
	return m.TableBase.DeserializeChild(childType, r)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func toFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case int32:
		return float64(x)
	case string:
		f, err := strconv.ParseFloat(x, 64)
		if err == nil {
			return f
		}
	case []any:
		// Accumulated values: sum all elements.
		sum := 0.0
		for _, elem := range x {
			sum += toFloat(elem)
		}
		return sum
	}
	return 0
}
