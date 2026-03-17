package table

// ManualBuildFunc is the callback type for the TableObject.ManualBuild event.
// The callback receives a TableHelper that is used to select which template
// rows and columns to print (and in what order).
//
// Example:
//
//	tbl.ManualBuild = func(h *TableHelper) {
//	    h.PrintRow(0)    // header row
//	    h.PrintColumns() // emit all template columns for that row
//	    for i := range data {
//	        h.PrintRow(1)    // data row template
//	        h.PrintColumns() // emit all columns
//	    }
//	}
type ManualBuildFunc func(h *TableHelper)

// TableHelper assists in programmatic table construction during a ManualBuild
// callback. It accumulates selected template rows and columns into a result
// TableBase that the engine renders instead of the design-time template.
//
// It is the Go equivalent of FastReport.Table.TableHelper.
//
// Usage pattern (rows-first):
//
//	h.PrintRow(index)    — select a template row, start a new result row
//	h.PrintColumn(index) — copy cells from that column into the current row
//
// Usage pattern (columns-first):
//
//	h.PrintColumn(index) — select a template column
//	h.PrintRow(index)    — copy cells from that row into the current column
//
// Shortcut helpers:
//
//	h.PrintRows()    — calls PrintRow + PrintColumns for every template row
//	h.PrintColumns() — calls PrintColumn for every template column
type TableHelper struct {
	src    *TableObject // template table
	result *TableBase   // accumulated result

	// State machine tracking whether rows or columns take priority.
	nowPrinting  nowPrinting
	rowsPriority bool

	// Current printing indices in the result table.
	printRowIdx int
	printColIdx int

	// Original (template) indices set by the last PrintRow/PrintColumn call.
	origRowIdx int
	origColIdx int

	// CellTextEval, when non-nil, is applied to each cell's text during copyCells.
	// Set before PrintColumn/PrintRow and clear it after PrintRows/PrintColumns.
	CellTextEval func(text string) string
}

type nowPrinting int

const (
	npNone   nowPrinting = iota
	npRow
	npColumn
)

// newTableHelper creates a TableHelper backed by the given template table.
func newTableHelper(src *TableObject) *TableHelper {
	return &TableHelper{
		src:         src,
		result:      NewTableBase(),
		nowPrinting: npNone,
	}
}

// NewTableHelper creates an exported TableHelper backed by the given template table.
// Use this when building a table programmatically from outside the table package.
func NewTableHelper(src *TableObject) *TableHelper {
	return newTableHelper(src)
}

// Result returns the accumulated result table produced by PrintRow/PrintColumn calls.
func (h *TableHelper) Result() *TableBase { return h.result }

// PrintRow selects template row at index and appends it to the result.
// Call PrintColumn after PrintRow to populate the cells for that row.
// Calling PrintRows() is a shortcut for PrintRow + PrintColumns on all rows.
func (h *TableHelper) PrintRow(index int) {
	if index < 0 || index >= h.src.RowCount() {
		return
	}
	h.origRowIdx = index
	srcRow := h.src.rows[index]

	if h.nowPrinting == npNone {
		// First call — rows take priority.
		h.rowsPriority = true
	}

	if h.rowsPriority {
		// Add a new result row copied from the template row.
		switch h.nowPrinting {
		case npNone:
			h.printRowIdx = 0
		case npColumn:
			h.printRowIdx++
		// npRow: two sequential PrintRow calls — printRowIdx advances on next PrintColumn
		}

		row := cloneRow(srcRow)
		h.result.rows = append(h.result.rows, row)
		h.nowPrinting = npRow
	} else {
		// Columns have priority — this PrintRow fills cells into existing columns.
		switch h.nowPrinting {
		case npColumn:
			// First row inside a column group — reset row index.
			h.printRowIdx = 0
		default:
			h.printRowIdx++
		}

		// Ensure result has enough rows.
		for len(h.result.rows) <= h.printRowIdx {
			h.result.rows = append(h.result.rows, cloneRow(srcRow))
		}

		// Copy cells for all result columns at this row.
		h.copyCells(h.origColIdx, h.origRowIdx, h.printColIdx, h.printRowIdx)
		h.nowPrinting = npRow
	}
}

// PrintRows calls PrintRow for every template row. In row-first mode (when
// PrintRow was the first call, or row priority is already set) it also calls
// PrintColumns after each row, replicating the full grid shortcut. In
// column-first mode it only calls PrintRow for each row, filling cells into
// the currently selected column — matching the C# TableObject.PrintRows()
// semantics used after PrintColumn.
func (h *TableHelper) PrintRows() {
	for i := 0; i < h.src.RowCount(); i++ {
		h.PrintRow(i)
		if h.rowsPriority {
			h.PrintColumns()
		}
	}
}

// PrintColumn selects template column at index and appends it to the result.
// Call PrintRow after PrintColumn to populate the cells for that column.
func (h *TableHelper) PrintColumn(index int) {
	if index < 0 || index >= h.src.ColumnCount() {
		return
	}
	h.origColIdx = index
	srcCol := h.src.columns[index]

	if h.nowPrinting == npNone {
		// First call — columns take priority.
		h.rowsPriority = false
	}

	if h.rowsPriority {
		// Rows have priority — PrintColumn fills cells into the current row.
		switch h.nowPrinting {
		case npRow:
			// First column inside a row group — reset col index.
			h.printColIdx = 0
		default:
			h.printColIdx++
		}

		// Ensure result columns are wide enough.
		for len(h.result.columns) <= h.printColIdx {
			h.result.columns = append(h.result.columns, cloneColumn(srcCol))
		}

		// Copy cells into the last result row.
		rowIdx := len(h.result.rows) - 1
		if rowIdx >= 0 {
			h.copyCells(h.origColIdx, h.origRowIdx, h.printColIdx, rowIdx)
		}
		h.nowPrinting = npColumn
	} else {
		// Columns take priority — add a new result column.
		switch h.nowPrinting {
		case npNone:
			h.printColIdx = 0
		case npRow:
			h.printColIdx++
		}

		h.result.columns = append(h.result.columns, cloneColumn(srcCol))
		h.nowPrinting = npColumn
	}
}

// PrintColumns calls PrintColumn for every template column.
func (h *TableHelper) PrintColumns() {
	for i := 0; i < h.src.ColumnCount(); i++ {
		h.PrintColumn(i)
	}
}

// copyCells copies a cell from the template at (srcColIdx, srcRowIdx) into
// the result at (dstColIdx, dstRowIdx), applying auto-span logic when
// ManualBuildAutoSpans is set.
func (h *TableHelper) copyCells(srcColIdx, srcRowIdx, dstColIdx, dstRowIdx int) {
	if dstRowIdx < 0 || dstRowIdx >= len(h.result.rows) {
		return
	}
	row := h.result.rows[dstRowIdx]
	// Ensure the row has enough cell slots.
	for len(row.cells) <= dstColIdx {
		row.cells = append(row.cells, NewTableCell())
	}

	var srcCell *TableCell
	if srcRowIdx >= 0 && srcRowIdx < h.src.RowCount() &&
		srcColIdx >= 0 && srcColIdx < h.src.ColumnCount() {
		srcCell = h.src.Cell(srcRowIdx, srcColIdx)
	}

	if srcCell != nil {
		dst := cloneCell(srcCell)
		if h.src.ManualBuildAutoSpans {
			// Reset spans — auto-span is resolved during rendering.
			dst.SetColSpan(1)
			dst.SetRowSpan(1)
		}
		if h.CellTextEval != nil {
			dst.SetText(h.CellTextEval(dst.Text()))
		}
		row.cells[dstColIdx] = dst
	}
}

// ── Clone helpers ─────────────────────────────────────────────────────────────

func cloneRow(src *TableRow) *TableRow {
	r := NewTableRow()
	r.SetHeight(src.Height())
	r.SetMinHeight(src.MinHeight())
	r.SetMaxHeight(src.MaxHeight())
	r.SetAutoSize(src.AutoSize())
	r.SetPageBreak(src.PageBreak())
	return r
}

func cloneColumn(src *TableColumn) *TableColumn {
	c := NewTableColumn()
	c.SetWidth(src.Width())
	c.SetMinWidth(src.MinWidth())
	c.SetMaxWidth(src.MaxWidth())
	c.SetAutoSize(src.AutoSize())
	c.SetPageBreak(src.PageBreak())
	return c
}

func cloneCell(src *TableCell) *TableCell {
	dst := NewTableCell()
	dst.SetColSpan(src.ColSpan())
	dst.SetRowSpan(src.RowSpan())
	dst.SetText(src.Text())
	dst.SetHorzAlign(src.HorzAlign())
	dst.SetVertAlign(src.VertAlign())
	dst.SetWordWrap(src.WordWrap())
	// Copy border and fill by value.
	b := src.Border()
	dst.SetBorder(b)
	dst.SetFill(src.Fill())
	dst.SetFont(src.Font())
	return dst
}
