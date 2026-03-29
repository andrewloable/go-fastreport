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

	// CellObjectEval, when non-nil, is called for each cloned cell during copyCells
	// to evaluate embedded objects (e.g. PictureObject DataColumn bindings).
	CellObjectEval func(cell *TableCell)

	// pageBreak is set by PageBreak() and consumed by the next PrintRow/PrintColumn.
	// Mirrors TableHelper.pageBreak (TableHelper.cs line 22).
	pageBreak bool

	// AutoSpans tracking — mirrors C# TableHelper columnSpans/rowSpans lists
	// (TableHelper.cs lines 20-21). These track cells with ColSpan/RowSpan > 1
	// so that repeated prints of the same template column/row automatically
	// expand the span in the result table.
	columnSpans []spanData
	rowSpans    []spanData
}

// spanData tracks a cell with ColSpan or RowSpan > 1 during AutoSpans processing.
// Mirrors C# TableHelper.SpanData (TableHelper.cs lines 365-372).
type spanData struct {
	originalCell       *TableCell // the template cell (has the original ColSpan/RowSpan)
	resultCell         *TableCell // the result cell whose span is being expanded
	originalCellOrigin [2]int     // [col, row] in the template table
	resultCellOrigin   [2]int     // [col, row] in the result table
	finishFlag         bool       // set when the last col/row of the original span is reached
}

type nowPrinting int

const (
	npNone   nowPrinting = iota
	npRow
	npColumn
)

// newTableHelper creates a TableHelper backed by the given template table.
func newTableHelper(src *TableObject) *TableHelper {
	res := NewTableBase()
	// Propagate layout settings from the source table so that the result
	// table renders with the same wrapping/pagination behaviour.
	// Mirrors C# TableHelper which inherits Layout/WrappedGap/FixedColumns.
	res.SetLayout(src.Layout())
	res.SetWrappedGap(src.WrappedGap())
	res.SetFixedColumns(src.FixedColumns())
	res.SetBorder(src.Border()) // C# propagates table border to result
	return &TableHelper{
		src:         src,
		result:      res,
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

// PageBreak marks that the next PrintRow or PrintColumn should set a page-break
// flag on the resulting row or column.
// Mirrors TableHelper.PageBreak() (TableHelper.cs line 180).
func (h *TableHelper) PageBreak() { h.pageBreak = true }

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
		row.SetOriginalIndex(index)
		row.SetPageBreak(h.pageBreak)
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
			r := cloneRow(srcRow)
			r.SetOriginalIndex(index)
			h.result.rows = append(h.result.rows, r)
		}

		// Apply page-break flag to the target row (new or existing).
		h.result.rows[h.printRowIdx].SetPageBreak(h.pageBreak)

		// Copy cells for all result columns at this row.
		h.copyCells(h.origColIdx, h.origRowIdx, h.printColIdx, h.printRowIdx)
		h.nowPrinting = npRow
	}
	h.pageBreak = false
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
			c := cloneColumn(srcCol)
			c.SetOriginalIndex(index)
			h.result.columns = append(h.result.columns, c)
		}

		// Apply page-break flag to the target column (new or existing).
		h.result.columns[h.printColIdx].SetPageBreak(h.pageBreak)

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

		col := cloneColumn(srcCol)
		col.SetOriginalIndex(index)
		col.SetPageBreak(h.pageBreak)
		h.result.columns = append(h.result.columns, col)
		h.nowPrinting = npColumn
	}
	h.pageBreak = false
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
//
// Mirrors C# TableHelper.CopyCells (TableHelper.cs lines 185-383).
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
	if srcCell == nil {
		return
	}

	needData := true

	if h.src.ManualBuildAutoSpans && h.rowsPriority {
		// ── Column span tracking (row-priority path) ──
		// Mirrors C# TableHelper.CopyCells AutoSpans rowsPriority block
		// (TableHelper.cs lines 193-279).
		if len(h.columnSpans) > 0 {
			sd := &h.columnSpans[0]
			// Check if we reached the last column of the original span.
			lastCol := sd.originalCellOrigin[0] + sd.originalCell.ColSpan() - 1
			if srcColIdx == lastCol {
				sd.finishFlag = true
			}
			// Clear span if we've wrapped back to the start after finishing,
			// or if we're outside the original span range.
			if (sd.finishFlag && srcColIdx == sd.originalCellOrigin[0]) ||
				srcColIdx < sd.originalCellOrigin[0] || srcColIdx > lastCol {
				h.columnSpans = h.columnSpans[:0]
			} else {
				sd.resultCell.SetColSpan(sd.resultCell.ColSpan() + 1)
				needData = false
			}
		}

		// Start tracking a new column span if the cell has ColSpan > 1.
		if srcCell.ColSpan() > 1 && len(h.columnSpans) == 0 {
			// The result cell will be created below when needData is true.
			// We store a placeholder and update it after the cell is placed.
			h.columnSpans = append(h.columnSpans, spanData{
				originalCell:       srcCell,
				originalCellOrigin: [2]int{srcColIdx, srcRowIdx},
				resultCellOrigin:   [2]int{dstColIdx, dstRowIdx},
				// resultCell is set below after the cell is placed in the row.
			})
		}

		// ── Row span tracking ──
		// Check row spans once per row (at the first column).
		if dstColIdx == 0 {
			for i := len(h.rowSpans) - 1; i >= 0; i-- {
				sd := &h.rowSpans[i]
				lastRow := sd.originalCellOrigin[1] + sd.originalCell.RowSpan() - 1
				if srcRowIdx == lastRow {
					sd.finishFlag = true
				}
				if (sd.finishFlag && srcRowIdx == sd.originalCellOrigin[1]) ||
					srcRowIdx < sd.originalCellOrigin[1] || srcRowIdx > lastRow {
					h.rowSpans = append(h.rowSpans[:i], h.rowSpans[i+1:]...)
				} else {
					sd.resultCell.SetRowSpan(sd.resultCell.RowSpan() + 1)
				}
			}
		}

		// Skip cell if it falls inside a row span.
		for i := range h.rowSpans {
			sd := &h.rowSpans[i]
			if dstColIdx >= sd.resultCellOrigin[0] &&
				dstColIdx <= sd.resultCellOrigin[0]+sd.resultCell.ColSpan()-1 &&
				dstRowIdx >= sd.resultCellOrigin[1] &&
				dstRowIdx <= sd.resultCellOrigin[1]+sd.resultCell.RowSpan() {
				needData = false
				break
			}
		}

		// Start tracking a new row span.
		if srcCell.RowSpan() > 1 && needData {
			h.rowSpans = append(h.rowSpans, spanData{
				originalCell:       srcCell,
				originalCellOrigin: [2]int{srcColIdx, srcRowIdx},
				resultCellOrigin:   [2]int{dstColIdx, dstRowIdx},
				// resultCell is set below.
			})
		}
	} else if h.src.ManualBuildAutoSpans && !h.rowsPriority {
		// ── Column-priority AutoSpans path ──
		// Mirrors C# TableHelper.CopyCells column-priority block
		// (TableHelper.cs lines 281-366).
		if len(h.rowSpans) > 0 {
			sd := &h.rowSpans[0]
			lastRow := sd.originalCellOrigin[1] + sd.originalCell.RowSpan() - 1
			if srcRowIdx == lastRow {
				sd.finishFlag = true
			}
			if (sd.finishFlag && srcRowIdx == sd.originalCellOrigin[1]) ||
				srcRowIdx < sd.originalCellOrigin[1] || srcRowIdx > lastRow {
				h.rowSpans = h.rowSpans[:0]
			} else {
				sd.resultCell.SetRowSpan(sd.resultCell.RowSpan() + 1)
				needData = false
			}
		}

		if srcCell.RowSpan() > 1 && len(h.rowSpans) == 0 {
			h.rowSpans = append(h.rowSpans, spanData{
				originalCell:       srcCell,
				originalCellOrigin: [2]int{srcColIdx, srcRowIdx},
				resultCellOrigin:   [2]int{dstColIdx, dstRowIdx},
			})
		}

		if dstRowIdx == 0 {
			for i := len(h.columnSpans) - 1; i >= 0; i-- {
				sd := &h.columnSpans[i]
				lastCol := sd.originalCellOrigin[0] + sd.originalCell.ColSpan() - 1
				if srcColIdx == lastCol {
					sd.finishFlag = true
				}
				if (sd.finishFlag && srcColIdx == sd.originalCellOrigin[0]) ||
					srcColIdx < sd.originalCellOrigin[0] || srcColIdx > lastCol {
					h.columnSpans = append(h.columnSpans[:i], h.columnSpans[i+1:]...)
				} else {
					sd.resultCell.SetColSpan(sd.resultCell.ColSpan() + 1)
				}
			}
		}

		for i := range h.columnSpans {
			sd := &h.columnSpans[i]
			if dstColIdx >= sd.resultCellOrigin[0] &&
				dstColIdx <= sd.resultCellOrigin[0]+sd.resultCell.ColSpan()-1 &&
				dstRowIdx >= sd.resultCellOrigin[1] &&
				dstRowIdx <= sd.resultCellOrigin[1]+sd.resultCell.RowSpan()-1 {
				needData = false
				break
			}
		}

		if srcCell.ColSpan() > 1 && needData {
			h.columnSpans = append(h.columnSpans, spanData{
				originalCell:       srcCell,
				originalCellOrigin: [2]int{srcColIdx, srcRowIdx},
				resultCellOrigin:   [2]int{dstColIdx, dstRowIdx},
			})
		}
	}

	if needData {
		dst := cloneCell(srcCell)
		dst.SetOriginalCellName(srcCell.Name())
		// When AutoSpans is active, result cell starts with ColSpan/RowSpan = 1;
		// the span tracking increments them. Without AutoSpans, preserve the
		// template spans directly.
		if h.src.ManualBuildAutoSpans {
			dst.SetColSpan(1)
			dst.SetRowSpan(1)
		}
		if h.CellTextEval != nil {
			dst.SetText(h.CellTextEval(dst.Text()))
		}
		if h.CellObjectEval != nil {
			h.CellObjectEval(dst)
		}
		row.cells[dstColIdx] = dst

		// Update spanData resultCell references to point to the placed cell.
		for i := range h.columnSpans {
			if h.columnSpans[i].resultCell == nil &&
				h.columnSpans[i].resultCellOrigin == [2]int{dstColIdx, dstRowIdx} {
				h.columnSpans[i].resultCell = dst
			}
		}
		for i := range h.rowSpans {
			if h.rowSpans[i].resultCell == nil &&
				h.rowSpans[i].resultCellOrigin == [2]int{dstColIdx, dstRowIdx} {
				h.rowSpans[i].resultCell = dst
			}
		}
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
	// Clone embedded objects (e.g. PictureObjects).
	for _, obj := range src.Objects() {
		dst.AddObject(obj)
	}
	return dst
}
