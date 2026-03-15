// Package table implements the table component for go-fastreport.
// It provides TableBase, TableObject, TableRow, TableColumn, and TableCell.
package table

import (
	"github.com/andrewloable/go-fastreport/report"
)

// TableLayout specifies the layout used when printing a large table across pages.
type TableLayout int

const (
	// TableLayoutAcrossThenDown prints the table across pages then down.
	TableLayoutAcrossThenDown TableLayout = iota
	// TableLayoutDownThenAcross prints the table down then across pages.
	TableLayoutDownThenAcross
	// TableLayoutWrapped wraps the table to a new row of pages.
	TableLayoutWrapped
)

// TableBase is the base for table-type report objects.
// It is the Go equivalent of FastReport.Table.TableBase.
type TableBase struct {
	report.BreakableComponent

	// rows is the ordered list of table rows.
	rows []*TableRow
	// columns is the ordered list of table columns.
	columns []*TableColumn

	// fixedRows is the number of header rows repeated on each page.
	fixedRows int
	// fixedColumns is the number of header columns repeated on each page.
	fixedColumns int
	// layout controls how the table is paginated.
	layout TableLayout
	// printOnParent prints the table on its parent band instead of a separate band.
	printOnParent bool
	// wrappedGap is the gap in pixels between wrapped table sections.
	wrappedGap float32
	// repeatHeaders repeats both row and column headers on each page.
	repeatHeaders bool
	// repeatRowHeaders repeats row headers on each page.
	repeatRowHeaders bool
	// repeatColumnHeaders repeats column headers on each page.
	repeatColumnHeaders bool
	// adjustSpannedCellsWidth adjusts column widths to fit spanned cells.
	adjustSpannedCellsWidth bool

	// Event names.
	ManualBuildEvent string
}

// NewTableBase creates a TableBase with defaults matching the C# constructor.
func NewTableBase() *TableBase {
	return &TableBase{
		BreakableComponent: *report.NewBreakableComponent(),
		repeatHeaders:      true, // C# default is true
	}
}

// --- Rows ---

// Rows returns the table rows.
func (t *TableBase) Rows() []*TableRow { return t.rows }

// RowCount returns the number of rows.
func (t *TableBase) RowCount() int { return len(t.rows) }

// Row returns the row at index i, or nil if out of range.
func (t *TableBase) Row(i int) *TableRow {
	if i < 0 || i >= len(t.rows) {
		return nil
	}
	return t.rows[i]
}

// AddRow appends a row and ensures it has the correct number of cells.
func (t *TableBase) AddRow(r *TableRow) {
	// Ensure the row has one cell per existing column.
	for len(r.cells) < len(t.columns) {
		c := NewTableCell()
		r.cells = append(r.cells, c)
	}
	t.rows = append(t.rows, r)
}

// NewRow creates a new TableRow, adds it to the table, and returns it.
func (t *TableBase) NewRow() *TableRow {
	r := NewTableRow()
	t.AddRow(r)
	return r
}

// --- Columns ---

// Columns returns the table columns.
func (t *TableBase) Columns() []*TableColumn { return t.columns }

// ColumnCount returns the number of columns.
func (t *TableBase) ColumnCount() int { return len(t.columns) }

// Column returns the column at index i, or nil if out of range.
func (t *TableBase) Column(i int) *TableColumn {
	if i < 0 || i >= len(t.columns) {
		return nil
	}
	return t.columns[i]
}

// AddColumn appends a column and adds a new cell to every existing row.
func (t *TableBase) AddColumn(c *TableColumn) {
	t.columns = append(t.columns, c)
	for _, r := range t.rows {
		if len(r.cells) < len(t.columns) {
			r.cells = append(r.cells, NewTableCell())
		}
	}
}

// NewColumn creates a new TableColumn, adds it to the table, and returns it.
func (t *TableBase) NewColumn() *TableColumn {
	c := NewTableColumn()
	t.AddColumn(c)
	return c
}

// --- Cell access ---

// Cell returns the cell at (row, col), or nil if out of range.
func (t *TableBase) Cell(row, col int) *TableCell {
	r := t.Row(row)
	if r == nil {
		return nil
	}
	return r.Cell(col)
}

// --- Properties ---

// FixedRows returns the number of fixed header rows.
func (t *TableBase) FixedRows() int { return t.fixedRows }

// SetFixedRows sets the fixed row count.
func (t *TableBase) SetFixedRows(v int) { t.fixedRows = v }

// FixedColumns returns the number of fixed header columns.
func (t *TableBase) FixedColumns() int { return t.fixedColumns }

// SetFixedColumns sets the fixed column count.
func (t *TableBase) SetFixedColumns(v int) { t.fixedColumns = v }

// Layout returns the table pagination layout.
func (t *TableBase) Layout() TableLayout { return t.layout }

// SetLayout sets the table pagination layout.
func (t *TableBase) SetLayout(l TableLayout) { t.layout = l }

// PrintOnParent returns whether the table prints on its parent band.
func (t *TableBase) PrintOnParent() bool { return t.printOnParent }

// SetPrintOnParent sets the print-on-parent flag.
func (t *TableBase) SetPrintOnParent(v bool) { t.printOnParent = v }

// WrappedGap returns the gap between wrapped table sections.
func (t *TableBase) WrappedGap() float32 { return t.wrappedGap }

// SetWrappedGap sets the gap between wrapped sections.
func (t *TableBase) SetWrappedGap(v float32) { t.wrappedGap = v }

// RepeatHeaders returns whether all headers are repeated on each page.
func (t *TableBase) RepeatHeaders() bool { return t.repeatHeaders }

// SetRepeatHeaders sets the repeat-headers flag.
func (t *TableBase) SetRepeatHeaders(v bool) { t.repeatHeaders = v }

// RepeatRowHeaders returns whether row headers are repeated on each page.
func (t *TableBase) RepeatRowHeaders() bool { return t.repeatRowHeaders }

// SetRepeatRowHeaders sets the repeat-row-headers flag.
func (t *TableBase) SetRepeatRowHeaders(v bool) { t.repeatRowHeaders = v }

// RepeatColumnHeaders returns whether column headers are repeated on each page.
func (t *TableBase) RepeatColumnHeaders() bool { return t.repeatColumnHeaders }

// SetRepeatColumnHeaders sets the repeat-column-headers flag.
func (t *TableBase) SetRepeatColumnHeaders(v bool) { t.repeatColumnHeaders = v }

// AdjustSpannedCellsWidth returns whether column widths adjust for spanned cells.
func (t *TableBase) AdjustSpannedCellsWidth() bool { return t.adjustSpannedCellsWidth }

// SetAdjustSpannedCellsWidth sets the adjust-spanned-cells-width flag.
func (t *TableBase) SetAdjustSpannedCellsWidth(v bool) { t.adjustSpannedCellsWidth = v }

// --- Serialization ---

// Serialize writes TableBase properties that differ from defaults.
func (t *TableBase) Serialize(w report.Writer) error {
	if err := t.BreakableComponent.Serialize(w); err != nil {
		return err
	}
	if t.fixedRows != 0 {
		w.WriteInt("FixedRows", t.fixedRows)
	}
	if t.fixedColumns != 0 {
		w.WriteInt("FixedColumns", t.fixedColumns)
	}
	if t.layout != TableLayoutAcrossThenDown {
		w.WriteInt("Layout", int(t.layout))
	}
	if t.printOnParent {
		w.WriteBool("PrintOnParent", true)
	}
	if t.wrappedGap != 0 {
		w.WriteFloat("WrappedGap", t.wrappedGap)
	}
	// RepeatHeaders default is true; write when false.
	if !t.repeatHeaders {
		w.WriteBool("RepeatHeaders", false)
	}
	if t.repeatRowHeaders {
		w.WriteBool("RepeatRowHeaders", true)
	}
	if t.repeatColumnHeaders {
		w.WriteBool("RepeatColumnHeaders", true)
	}
	if t.adjustSpannedCellsWidth {
		w.WriteBool("AdjustSpannedCellsWidth", true)
	}
	if t.ManualBuildEvent != "" {
		w.WriteStr("ManualBuildEvent", t.ManualBuildEvent)
	}
	// Serialize columns.
	for _, col := range t.columns {
		if err := w.WriteObject(col); err != nil {
			return err
		}
	}
	// Serialize rows (each row serializes its cells).
	for _, row := range t.rows {
		if err := w.WriteObject(row); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads TableBase properties.
func (t *TableBase) Deserialize(r report.Reader) error {
	if err := t.BreakableComponent.Deserialize(r); err != nil {
		return err
	}
	t.fixedRows = r.ReadInt("FixedRows", 0)
	t.fixedColumns = r.ReadInt("FixedColumns", 0)
	t.layout = TableLayout(r.ReadInt("Layout", 0))
	t.printOnParent = r.ReadBool("PrintOnParent", false)
	t.wrappedGap = r.ReadFloat("WrappedGap", 0)
	t.repeatHeaders = r.ReadBool("RepeatHeaders", true) // C# default is true
	t.repeatRowHeaders = r.ReadBool("RepeatRowHeaders", false)
	t.repeatColumnHeaders = r.ReadBool("RepeatColumnHeaders", false)
	t.adjustSpannedCellsWidth = r.ReadBool("AdjustSpannedCellsWidth", false)
	t.ManualBuildEvent = r.ReadStr("ManualBuildEvent", "")
	return nil
}

// ── TableObject ───────────────────────────────────────────────────────────────

// TableObject is the concrete table report component.
// It is the Go equivalent of FastReport.Table.TableObject.
type TableObject struct {
	TableBase

	// ManualBuildAutoSpans propagates ColSpan/RowSpan automatically
	// during a ManualBuild event (default: true).
	ManualBuildAutoSpans bool
}

// TypeName returns the FRX element name.
func (t *TableObject) TypeName() string { return "TableObject" }

// NewTableObject creates a TableObject with defaults.
func NewTableObject() *TableObject {
	return &TableObject{
		TableBase:            *NewTableBase(),
		ManualBuildAutoSpans: true, // C# default is true
	}
}

// Serialize writes TableObject properties.
func (t *TableObject) Serialize(w report.Writer) error {
	if err := t.TableBase.Serialize(w); err != nil {
		return err
	}
	// ManualBuildAutoSpans default is true; write when false.
	if !t.ManualBuildAutoSpans {
		w.WriteBool("ManualBuildAutoSpans", false)
	}
	return nil
}

// Deserialize reads TableObject properties.
func (t *TableObject) Deserialize(r report.Reader) error {
	if err := t.TableBase.Deserialize(r); err != nil {
		return err
	}
	t.ManualBuildAutoSpans = r.ReadBool("ManualBuildAutoSpans", true)
	return nil
}
