// Package table implements the table component for go-fastreport.
// It provides TableBase, TableObject, TableRow, TableColumn, and TableCell.
package table

import (
	"image"

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
	// styles deduplicates cell styles used during rendering.
	// See FastReport.Table.TableBase.styles (TableBase.cs line 39).
	styles *TableStyleCollection

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

	// lockCorrectSpans prevents CorrectSpansOnRowChange/CorrectSpansOnColumnChange
	// from modifying cell spans.
	lockCorrectSpans bool

	// spanList is a cached list of span rectangles; nil means not yet computed.
	// Each rectangle covers (col, row)–(col+colSpan, row+rowSpan).
	spanList []image.Rectangle

	// Event names.
	ManualBuildEvent string
}

// NewTableBase creates a TableBase with defaults matching the C# constructor.
// See FastReport.Table.TableBase constructor (TableBase.cs line 1384).
func NewTableBase() *TableBase {
	return &TableBase{
		BreakableComponent: *report.NewBreakableComponent(),
		styles:             NewTableStyleCollection(),
		repeatHeaders:      true, // C# default is true
	}
}

// Styles returns the style deduplication collection owned by this table.
// It is the Go equivalent of the internal FastReport.Table.TableBase.Styles property
// (TableBase.cs line 77).
func (t *TableBase) Styles() *TableStyleCollection { return t.styles }

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

// InsertRow inserts r at index idx (0-based), shifting subsequent rows down.
// It corrects cell spans for all cells whose row span crosses idx.
// Mirrors C# TableRowCollection.OnInsert → CorrectSpansOnRowChange(index, +1).
func (t *TableBase) InsertRow(idx int, r *TableRow) {
	if idx < 0 {
		idx = 0
	}
	if idx > len(t.rows) {
		idx = len(t.rows)
	}
	// Ensure the row has one cell per existing column.
	for len(r.cells) < len(t.columns) {
		r.cells = append(r.cells, NewTableCell())
	}
	t.rows = append(t.rows, nil)
	copy(t.rows[idx+1:], t.rows[idx:])
	t.rows[idx] = r
	// Correct spans that extend across the inserted row index.
	t.CorrectSpansOnRowChange(idx, +1)
}

// RemoveRow removes the row at index idx and corrects cell spans.
// Mirrors C# TableRowCollection.OnRemove → CorrectSpansOnRowChange(index, -1).
func (t *TableBase) RemoveRow(idx int) {
	if idx < 0 || idx >= len(t.rows) {
		return
	}
	t.rows = append(t.rows[:idx], t.rows[idx+1:]...)
	// Correct spans that extended across the removed row index.
	t.CorrectSpansOnRowChange(idx, -1)
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

// InsertColumn inserts c at index idx (0-based), shifting subsequent columns right.
// It corrects cell spans for all cells whose column span crosses idx, and inserts
// a new empty cell at position idx in every existing row.
// Mirrors C# TableColumnCollection.OnInsert → CorrectSpansOnColumnChange(index, +1).
func (t *TableBase) InsertColumn(idx int, c *TableColumn) {
	if idx < 0 {
		idx = 0
	}
	if idx > len(t.columns) {
		idx = len(t.columns)
	}
	t.columns = append(t.columns, nil)
	copy(t.columns[idx+1:], t.columns[idx:])
	t.columns[idx] = c
	// Insert a new empty cell at idx in every row.
	for _, r := range t.rows {
		r.cells = append(r.cells, nil)
		copy(r.cells[idx+1:], r.cells[idx:])
		r.cells[idx] = NewTableCell()
	}
	// Correct spans that extend across the inserted column index.
	t.CorrectSpansOnColumnChange(idx, +1)
}

// RemoveColumn removes the column at index idx, removes the corresponding cell
// from every row, and corrects cell spans.
// Mirrors C# TableColumnCollection.OnRemove → CorrectSpansOnColumnChange(index, -1).
func (t *TableBase) RemoveColumn(idx int) {
	if idx < 0 || idx >= len(t.columns) {
		return
	}
	t.columns = append(t.columns[:idx], t.columns[idx+1:]...)
	for _, r := range t.rows {
		if idx < len(r.cells) {
			r.cells = append(r.cells[:idx], r.cells[idx+1:]...)
		}
	}
	// Correct spans that extended across the removed column index.
	t.CorrectSpansOnColumnChange(idx, -1)
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

// LockCorrectSpans returns the lock flag for span correction operations.
func (t *TableBase) LockCorrectSpans() bool { return t.lockCorrectSpans }

// SetLockCorrectSpans sets the lock flag; when true,
// CorrectSpansOnRowChange and CorrectSpansOnColumnChange are no-ops.
func (t *TableBase) SetLockCorrectSpans(v bool) { t.lockCorrectSpans = v }

// GetSpanList returns a cached list of span rectangles for all cells with
// ColSpan > 1 or RowSpan > 1. Each image.Rectangle is (Min={col,row},
// Max={col+colSpan, row+rowSpan}). The list is rebuilt lazily after
// ResetSpanList. Mirrors C# TableBase.GetSpanList (TableBase.cs).
func (t *TableBase) GetSpanList() []image.Rectangle {
	if t.spanList != nil {
		return t.spanList
	}
	t.spanList = []image.Rectangle{}
	for row := range t.rows {
		for col := range t.columns {
			c := t.Cell(row, col)
			if c == nil {
				continue
			}
			if c.ColSpan() > 1 || c.RowSpan() > 1 {
				t.spanList = append(t.spanList, image.Rect(col, row, col+c.ColSpan(), row+c.RowSpan()))
			}
		}
	}
	return t.spanList
}

// ResetSpanList clears the span cache so the next GetSpanList call recomputes it.
// Mirrors C# TableBase.ResetSpanList (TableBase.cs).
func (t *TableBase) ResetSpanList() { t.spanList = nil }

// IsInsideSpan reports whether the cell at (col, row) is covered by another
// cell's span — i.e. it is not the origin of a span but lies within one.
// Mirrors C# TableBase.IsInsideSpan (TableBase.cs).
func (t *TableBase) IsInsideSpan(col, row int) bool {
	for _, span := range t.GetSpanList() {
		// The origin cell (span.Min) is not "inside" its own span.
		if col == span.Min.X && row == span.Min.Y {
			continue
		}
		if col >= span.Min.X && col < span.Max.X && row >= span.Min.Y && row < span.Max.Y {
			return true
		}
	}
	return false
}

// ProcessDuplicates applies the CellDuplicates deduplication logic to all
// eligible cells in the table.
//
// For each cell where CellDuplicates != CellDuplicatesShow the method finds
// adjacent cells (right and down) that share the same Name() and Text(), then:
//   - CellDuplicatesClear: blanks the text of all duplicate cells except the origin
//   - CellDuplicatesMerge: expands the origin cell's ColSpan/RowSpan
//   - CellDuplicatesMergeNonEmpty: same as Merge but only when text is non-empty
//
// Mirrors C# TableResult.ProcessDuplicates() lines 165-245
// (FastReport.Base/Table/TableResult.cs).
func (t *TableBase) ProcessDuplicates() {
	type span struct{ x, y, w, h int }
	var processed []span

	// isInProcessed reports whether (col, row) falls inside any already-processed span.
	// Mirrors C# local IsInsideSpan(cell, list) using cell.Address.
	isInProcessed := func(col, row int) bool {
		for _, s := range processed {
			if col >= s.x && col < s.x+s.w && row >= s.y && row < s.y+s.h {
				return true
			}
		}
		return false
	}

	rowCount := t.RowCount()
	colCount := t.ColumnCount()

	// Outer loops: col first, row second — mirrors C# for(x) { for(y) }.
	for col := 0; col < colCount; col++ {
		for row := 0; row < rowCount; row++ {
			cell := t.Cell(row, col)
			if cell == nil || cell.Duplicates() == CellDuplicatesShow {
				continue
			}
			if isInProcessed(col, row) {
				continue
			}

			cellAlias := cell.Name()
			cellText := cell.Text()
			cellDups := cell.Duplicates()
			designColSpan := cell.ColSpan() // design-time ColSpan — mirrors cellData.ColSpan

			// countInRow counts consecutive cells starting at (col, targetRow) that
			// share the same name and text, stopping at already-processed spans.
			// Mirrors C# Func<int,int> func = (row) => { ... }
			countInRow := func(targetRow int) int {
				count := 0
				for xi := col; xi < colCount; xi++ {
					if isInProcessed(xi, targetRow) {
						break
					}
					c := t.Cell(targetRow, xi)
					if c == nil || c.Name() != cellAlias || c.Text() != cellText {
						break
					}
					count++
				}
				return count
			}

			colSpan := countInRow(row)
			rowSpan := 1
			for yi := row + 1; yi < rowCount; yi++ {
				if countInRow(yi) < designColSpan {
					break
				}
				rowSpan++
			}

			switch cellDups {
			case CellDuplicatesClear:
				// Blank all duplicate cells except the origin (0,0).
				for dx := 0; dx < colSpan; dx++ {
					for dy := 0; dy < rowSpan; dy++ {
						if dx == 0 && dy == 0 {
							continue
						}
						if c := t.Cell(row+dy, col+dx); c != nil {
							c.SetText("")
						}
					}
				}
			case CellDuplicatesMerge:
				cell.SetColSpan(colSpan)
				cell.SetRowSpan(rowSpan)
			case CellDuplicatesMergeNonEmpty:
				if cellText != "" {
					cell.SetColSpan(colSpan)
					cell.SetRowSpan(rowSpan)
				}
			}

			processed = append(processed, span{col, row, colSpan, rowSpan})
		}
	}
}

// CorrectSpansOnRowChange adjusts ColSpan/RowSpan values when a row is
// inserted (delta=+1) or removed (delta=-1) at rowIdx.
// Mirrors C# TableBase.CorrectSpansOnRowChange (TableBase.cs).
// A no-op when LockCorrectSpans is true.
func (t *TableBase) CorrectSpansOnRowChange(rowIdx, delta int) {
	if t.lockCorrectSpans {
		return
	}
	t.ResetSpanList()
	for row := 0; row < len(t.rows); row++ {
		for col := 0; col < len(t.columns); col++ {
			c := t.Cell(row, col)
			if c == nil || c.RowSpan() <= 1 {
				continue
			}
			spanEnd := row + c.RowSpan()
			if rowIdx > row && rowIdx < spanEnd {
				c.SetRowSpan(c.RowSpan() + delta)
				if c.RowSpan() < 1 {
					c.SetRowSpan(1)
				}
			}
		}
	}
	// Handle structural changes for insertion.
	if delta > 0 {
		newRow := NewTableRow()
		newRow.SetHeight(t.rows[0].Height())
		for range t.columns {
			newRow.cells = append(newRow.cells, NewTableCell())
		}
		t.rows = append(t.rows[:rowIdx], append([]*TableRow{newRow}, t.rows[rowIdx:]...)...)
	} else if delta < 0 && rowIdx < len(t.rows) {
		t.rows = append(t.rows[:rowIdx], t.rows[rowIdx+1:]...)
	}
}

// CorrectSpansOnColumnChange adjusts ColSpan/RowSpan values when a column is
// inserted (delta=+1) or removed (delta=-1) at colIdx.
// Mirrors C# TableBase.CorrectSpansOnColumnChange (TableBase.cs).
// A no-op when LockCorrectSpans is true.
func (t *TableBase) CorrectSpansOnColumnChange(colIdx, delta int) {
	if t.lockCorrectSpans {
		return
	}
	t.ResetSpanList()
	for row := 0; row < len(t.rows); row++ {
		for col := 0; col < len(t.columns); col++ {
			c := t.Cell(row, col)
			if c == nil || c.ColSpan() <= 1 {
				continue
			}
			spanEnd := col + c.ColSpan()
			if colIdx > col && colIdx < spanEnd {
				c.SetColSpan(c.ColSpan() + delta)
				if c.ColSpan() < 1 {
					c.SetColSpan(1)
				}
			}
		}
	}
	// Handle structural changes.
	if delta > 0 {
		newCol := NewTableColumn()
		t.columns = append(t.columns[:colIdx], append([]*TableColumn{newCol}, t.columns[colIdx:]...)...)
		for _, row := range t.rows {
			newCell := NewTableCell()
			row.cells = append(row.cells[:colIdx], append([]*TableCell{newCell}, row.cells[colIdx:]...)...)
		}
	} else if delta < 0 && colIdx < len(t.columns) {
		t.columns = append(t.columns[:colIdx], t.columns[colIdx+1:]...)
		for _, row := range t.rows {
			if colIdx < len(row.cells) {
				row.cells = append(row.cells[:colIdx], row.cells[colIdx+1:]...)
			}
		}
	}
}

// GetExpressions collects expression strings from the table itself and from
// every cell in every row.  This mirrors the behaviour of C# TableCell.GetExpressions
// (TableCell.cs lines 336-350) applied at the table level: the engine needs
// all expressions declared in the table's cells to pre-compile them.
//
// The base component expressions (VisibleExpression, PrintableExpression, etc.)
// are collected first via the embedded BreakableComponent, then each cell's
// GetExpressions result is appended.
func (t *TableBase) GetExpressions() []string {
	exprs := t.BreakableComponent.GetExpressions()
	for _, row := range t.rows {
		for _, cell := range row.cells {
			exprs = append(exprs, cell.GetExpressions()...)
		}
	}
	return exprs
}

// Assign copies all TableBase properties from src.
// Mirrors C# TableBase.Assign (TableBase.cs:473-489).
// Note: rows/columns/cells are not copied — they are structural and managed
// separately by the engine (as in C# where Assign does not copy child collections).
func (t *TableBase) Assign(src *TableBase) {
	if src == nil {
		return
	}
	t.BreakableComponent.Assign(&src.BreakableComponent)
	t.fixedRows = src.fixedRows
	t.fixedColumns = src.fixedColumns
	t.printOnParent = src.printOnParent
	t.repeatHeaders = src.repeatHeaders
	t.repeatRowHeaders = src.repeatRowHeaders
	t.repeatColumnHeaders = src.repeatColumnHeaders
	t.layout = src.layout
	t.wrappedGap = src.wrappedGap
	t.adjustSpannedCellsWidth = src.adjustSpannedCellsWidth
	t.ManualBuildEvent = src.ManualBuildEvent
}

// SaveState saves the current state of all rows, columns, and cells, then
// sets CanGrow=true and CanShrink=true (matching C# TableBase.SaveState).
// Mirrors C# TableBase.SaveState (TableBase.cs).
func (t *TableBase) SaveState() {
	for _, row := range t.rows {
		row.SaveState()
		for _, cell := range row.cells {
			cell.SaveState()
		}
	}
	for _, col := range t.columns {
		col.SaveState()
	}
	t.SetCanGrow(true)
	t.SetCanShrink(true)
}

// RestoreState restores all rows, columns, and cells to the state saved by
// the most recent SaveState call.
// Mirrors C# TableBase.RestoreState (TableBase.cs).
func (t *TableBase) RestoreState() {
	for _, row := range t.rows {
		row.RestoreState()
		for _, cell := range row.cells {
			cell.RestoreState()
		}
	}
	for _, col := range t.columns {
		col.RestoreState()
	}
	t.ResetSpanList()
}

// CalcSpans auto-detects RowSpan for ManualBuild tables when
// ManualBuildAutoSpans is true. For each column, consecutive cells with
// identical text are merged: the first cell gets RowSpan = count, the rest
// get RowSpan = 0 (marking them as spanned/covered).
// Mirrors C# TableBase.CalcSpans (TableBase.cs).
func (t *TableBase) CalcSpans() {
	nCols := len(t.columns)
	nRows := len(t.rows)
	for ci := 0; ci < nCols; ci++ {
		ri := 0
		for ri < nRows {
			cell := t.Cell(ri, ci)
			if cell == nil {
				ri++
				continue
			}
			// Count consecutive cells with the same text in this column.
			span := 1
			for ri+span < nRows {
				next := t.Cell(ri+span, ci)
				if next == nil || next.Text() != cell.Text() {
					break
				}
				span++
			}
			if span > 1 {
				cell.SetRowSpan(span)
				// Mark covered cells with RowSpan=0 so IsInsideSpan skips them.
				for k := 1; k < span; k++ {
					if c := t.Cell(ri+k, ci); c != nil {
						c.SetRowSpan(0)
					}
				}
			}
			ri += span
		}
	}
	t.ResetSpanList()
}

// CalcWidth returns the total width of the table by summing visible column
// widths, applying AutoSize expansion/contraction from cell widths.
// When AutoSize=true and MinWidth > 0, the column contracts to MinWidth
// if content doesn't require the declared width (no text measurement
// available in Go, so MinWidth is used as the contracted width).
// Mirrors C# TableBase.CalcWidth (TableBase.cs).
func (t *TableBase) CalcWidth() float32 {
	if len(t.rows) > 0 {
		for ci, col := range t.columns {
			if col.AutoSize() {
				cell := t.Cell(0, ci)
				if cell != nil && cell.Width() > col.Width() {
					// Expand to fit content.
					col.SetWidth(cell.Width())
				} else if col.MinWidth() > 0 && col.Width() > col.MinWidth() {
					// Contract to MinWidth when content fits.
					// C# measures actual text; Go approximates with MinWidth.
				col.SetWidth(col.MinWidth())
				}
			}
		}
	}
	var total float32
	for _, col := range t.columns {
		if col.Visible() {
			total += col.Width()
		}
	}
	// C# ref: TableBase.CalcWidth sets Width = total (line 983).
	// This propagates the actual rendered width to the component so that
	// sibling objects can be repositioned (e.g. two matrices side-by-side).
	t.SetWidth(total)
	return total
}

// CalcHeight returns the total height of the table by summing visible row
// heights, applying AutoSize expansion from the tallest cell in each row.
// Mirrors C# TableBase.CalcHeight (TableBase.cs).
func (t *TableBase) CalcHeight() float32 {
	for ri, row := range t.rows {
		if row.AutoSize() {
			var maxH float32
			for ci := range t.columns {
				cell := t.Cell(ri, ci)
				if cell != nil && cell.Height() > maxH {
					maxH = cell.Height()
				}
			}
			if maxH > row.Height() {
				row.ComponentBase.SetHeight(maxH)
			}
		}
	}
	var total float32
	for _, row := range t.rows {
		if row.Visible() {
			total += row.Height()
		}
	}
	return total
}

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
	// Layout can be stored as an int (0/1/2) or a string enum name
	// ("AcrossThenDown"/"DownThenAcross"/"Wrapped") in FRX files.
	layoutStr := r.ReadStr("Layout", "")
	switch layoutStr {
	case "Wrapped":
		t.layout = TableLayoutWrapped
	case "DownThenAcross":
		t.layout = TableLayoutDownThenAcross
	default:
		t.layout = TableLayout(r.ReadInt("Layout", 0))
	}
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

	// ManualBuild is an optional callback invoked during report preparation
	// to programmatically construct the table instead of using the static
	// design-time rows/columns. Use the provided TableHelper to call
	// PrintRow/PrintColumn in the order you want them rendered.
	//
	// When ManualBuild is set, the engine ignores the template rows/columns
	// and renders the result table produced by the helper.
	ManualBuild ManualBuildFunc
}

// TypeName returns the FRX element name.
func (t *TableObject) TypeName() string { return "TableObject" }

// Assign copies all TableObject properties from src, including the TableBase fields.
// Mirrors C# TableObject.Assign (TableObject.cs:226-233).
func (t *TableObject) Assign(src *TableObject) {
	if src == nil {
		return
	}
	t.TableBase.Assign(&src.TableBase)
	t.ManualBuildAutoSpans = src.ManualBuildAutoSpans
	t.ManualBuild = src.ManualBuild
}

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

// IsManualBuild returns true when a ManualBuild callback or ManualBuildEvent
// script name is set, meaning the engine should call InvokeManualBuild instead
// of rendering the static template rows/columns.
func (t *TableObject) IsManualBuild() bool {
	return t.ManualBuild != nil || t.ManualBuildEvent != ""
}

// InvokeManualBuild calls the ManualBuild callback (if set), passing a fresh
// TableHelper backed by this TableObject. It returns the result TableBase
// built by the callback, or nil when neither callback nor event is set.
//
// The returned TableBase is ready for rendering: its rows/columns/cells are
// copies of the selected template elements (no references to the template).
func (t *TableObject) InvokeManualBuild() *TableBase {
	if t.ManualBuild == nil {
		return nil
	}
	h := newTableHelper(t)
	t.ManualBuild(h)
	return h.Result()
}
