// Package matrix implements the pivot/matrix table component for go-fastreport.
// It is the Go equivalent of FastReport.Matrix.MatrixObject.
package matrix

import (
	"fmt"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/table"
)

// ── Aggregate function ────────────────────────────────────────────────────────

// AggregateFunction determines how cell values are aggregated.
type AggregateFunction int

const (
	AggregateFunctionNone          AggregateFunction = iota
	AggregateFunctionSum                              // Sum all values.
	AggregateFunctionMin                              // Minimum value.
	AggregateFunctionMax                              // Maximum value.
	AggregateFunctionAvg                              // Average value.
	AggregateFunctionCount                            // Count of rows.
	AggregateFunctionCountDistinct                    // Count distinct values.
)

// ── EvenStylePriority ─────────────────────────────────────────────────────────

// EvenStylePriority controls which axis the even/odd style alternates on.
type EvenStylePriority int

const (
	EvenStylePriorityRows    EvenStylePriority = iota // alternate on rows
	EvenStylePriorityColumns                          // alternate on columns
)

// ── Descriptors ───────────────────────────────────────────────────────────────

// Descriptor is the base for column/row/cell descriptors.
type Descriptor struct {
	// Expression is a report expression (e.g. "[DataSource.Field]").
	Expression string
}

// HeaderDescriptor describes a column or row header group.
type HeaderDescriptor struct {
	Descriptor

	// Sort controls sort direction ("" = none, "asc", "desc").
	Sort string
	// ShowTotal adds a total column/row for this header level.
	ShowTotal bool
	// TotalText is the caption for the total cell.
	TotalText string
	// InterleaveRows interlays rows between this header.
	InterleaveRows bool
}

// NewHeaderDescriptor creates a HeaderDescriptor with the given expression.
func NewHeaderDescriptor(expr string) *HeaderDescriptor {
	return &HeaderDescriptor{
		Descriptor: Descriptor{Expression: expr},
		TotalText:  "Total",
	}
}

// CellDescriptor describes a data cell with an aggregate function.
type CellDescriptor struct {
	Descriptor

	// Function is the aggregate function applied to the cell values.
	Function AggregateFunction
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
type MatrixData struct {
	// Columns are the column header descriptors (from outermost to innermost).
	Columns []*HeaderDescriptor
	// Rows are the row header descriptors (from outermost to innermost).
	Rows []*HeaderDescriptor
	// Cells are the data cell descriptors.
	Cells []*CellDescriptor
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

	// AutoSize enables automatic column/row sizing from content.
	AutoSize bool
	// CellsSideBySide renders multiple cell descriptors side by side.
	CellsSideBySide bool
	// EvenStylePriority controls even-style alternation axis.
	EvenStylePriority EvenStylePriority
	// Style is the named style to apply (e.g. "Green").
	Style string

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
}

// New creates a MatrixObject with defaults.
func New() *MatrixObject {
	return &MatrixObject{
		TableBase:    *table.NewTableBase(),
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
	if hw.h.Sort != "" {
		w.WriteStr("Sort", hw.h.Sort)
	}
	if hw.h.ShowTotal {
		w.WriteBool("ShowTotal", true)
	}
	if hw.h.TotalText != "" && hw.h.TotalText != "Total" {
		w.WriteStr("TotalText", hw.h.TotalText)
	}
	if hw.h.InterleaveRows {
		w.WriteBool("InterleaveRows", true)
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
	if cw.c.Function != AggregateFunctionNone {
		w.WriteInt("Function", int(cw.c.Function))
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
	if m.AutoSize {
		w.WriteBool("AutoSize", true)
	}
	if m.CellsSideBySide {
		w.WriteBool("CellsSideBySide", true)
	}
	if m.EvenStylePriority != EvenStylePriorityRows {
		w.WriteInt("EvenStylePriority", int(m.EvenStylePriority))
	}
	if m.Style != "" {
		w.WriteStr("Style", m.Style)
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
	m.AutoSize = r.ReadBool("AutoSize", false)
	m.CellsSideBySide = r.ReadBool("CellsSideBySide", false)
	m.EvenStylePriority = EvenStylePriority(r.ReadInt("EvenStylePriority", 0))
	m.Style = r.ReadStr("Style", "")
	return nil
}

// DeserializeChild handles matrix-specific child elements (MatrixRows, MatrixColumns, MatrixCells).
func (m *MatrixObject) DeserializeChild(childType string, r report.Reader) bool {
	switch childType {
	case "MatrixRows":
		for {
			ct, ok := r.NextChild()
			if !ok {
				break
			}
			if ct == "Header" {
				hd := &HeaderDescriptor{
					Descriptor:     Descriptor{Expression: r.ReadStr("Expression", "")},
					Sort:           r.ReadStr("Sort", ""),
					ShowTotal:      r.ReadBool("ShowTotal", false),
					TotalText:      r.ReadStr("TotalText", "Total"),
					InterleaveRows: r.ReadBool("InterleaveRows", false),
				}
				m.Data.Rows = append(m.Data.Rows, hd)
			}
			_ = r.FinishChild()
		}
		return true
	case "MatrixColumns":
		for {
			ct, ok := r.NextChild()
			if !ok {
				break
			}
			if ct == "Header" {
				hd := &HeaderDescriptor{
					Descriptor:     Descriptor{Expression: r.ReadStr("Expression", "")},
					Sort:           r.ReadStr("Sort", ""),
					ShowTotal:      r.ReadBool("ShowTotal", false),
					TotalText:      r.ReadStr("TotalText", "Total"),
					InterleaveRows: r.ReadBool("InterleaveRows", false),
				}
				m.Data.Columns = append(m.Data.Columns, hd)
			}
			_ = r.FinishChild()
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
					Function:   AggregateFunction(r.ReadInt("Function", 0)),
				}
				m.Data.Cells = append(m.Data.Cells, cd)
			}
			_ = r.FinishChild()
		}
		return true
	}
	return false
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
	}
	return 0
}
