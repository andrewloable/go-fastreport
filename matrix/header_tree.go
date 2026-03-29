package matrix

import (
	"fmt"

	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/table"
	"github.com/andrewloable/go-fastreport/utils"
)

// HeaderItem is one node in a multi-level header tree.
// It mirrors FastReport.Matrix.MatrixHeaderItem.
//
// The tree is built from paths supplied to AddDataMultiLevel; each unique
// path segment becomes a child node.  Leaf nodes carry CellSize == 1 and
// are the actual column/row headers in the rendered table.
type HeaderItem struct {
	// Value is the display text for this header cell.
	Value string
	// Children are the next-level nested items.
	Children []*HeaderItem
	// childIndex maps value → child position for fast lookup.
	childIndex map[string]int

	// CellSize is the number of leaf columns/rows this item spans
	// (computed after all data is added).
	CellSize int
	// LevelSize is the number of header levels in the sub-tree rooted here
	// (computed after all data is added; 1 for leaf nodes).
	LevelSize int
}

// newHeaderItem creates a leaf node.
func newHeaderItem(value string) *HeaderItem {
	return &HeaderItem{
		Value:      value,
		childIndex: make(map[string]int),
	}
}

// ensureChild returns the child with the given value, creating it if absent.
func (h *HeaderItem) ensureChild(value string) *HeaderItem {
	if idx, ok := h.childIndex[value]; ok {
		return h.Children[idx]
	}
	child := newHeaderItem(value)
	h.childIndex[value] = len(h.Children)
	h.Children = append(h.Children, child)
	return child
}

// isLeaf reports whether this item has no children.
func (h *HeaderItem) isLeaf() bool { return len(h.Children) == 0 }

// computeSizes sets CellSize and LevelSize for the entire sub-tree.
// Call on the root item once all paths have been inserted.
func (h *HeaderItem) computeSizes() {
	if h.isLeaf() {
		h.CellSize = 1
		h.LevelSize = 1
		return
	}
	totalCells := 0
	maxLevel := 0
	for _, child := range h.Children {
		child.computeSizes()
		totalCells += child.CellSize
		if child.LevelSize > maxLevel {
			maxLevel = child.LevelSize
		}
	}
	h.CellSize = totalCells
	h.LevelSize = maxLevel + 1
}

// leaves returns all leaf items in left-to-right order.
func (h *HeaderItem) leaves() []*HeaderItem {
	if h.isLeaf() {
		return []*HeaderItem{h}
	}
	var out []*HeaderItem
	for _, child := range h.Children {
		out = append(out, child.leaves()...)
	}
	return out
}

// ── MultiLevel additions to MatrixObject ─────────────────────────────────────

// MultiLevelKey identifies a cell by a slice of row-path segments, a slice of
// col-path segments, and a cell descriptor index.
type multiLevelKey struct {
	row     string // joined path string
	col     string // joined path string
	cellIdx int
}

// Initialise multi-level state on first use.
func (m *MatrixObject) ensureMultiLevel() {
	if m.rowRoot == nil {
		m.rowRoot = newHeaderItem("") // virtual root
		m.colRoot = newHeaderItem("") // virtual root
		m.mlAccumulators = make(map[multiLevelKey]*accumulator)
	}
}

// AddDataMultiLevel processes one logical row with hierarchical dimension paths.
//
// rowPath contains the value for each row-header level, outermost first
// (e.g. []string{"2024", "Q1", "Jan"}).  colPath is the same for columns.
// values are the cell values in descriptor order.
//
// This method and AddData can be used independently; they populate separate
// internal state and BuildTemplateMultiLevel / BuildTemplate are called respectively.
func (m *MatrixObject) AddDataMultiLevel(rowPath, colPath []string, values []any) {
	m.ensureMultiLevel()

	// Insert into trees.
	rNode := m.rowRoot
	for _, seg := range rowPath {
		rNode = rNode.ensureChild(seg)
	}
	cNode := m.colRoot
	for _, seg := range colPath {
		cNode = cNode.ensureChild(seg)
	}

	// Build composite keys from path segments.
	rKey := joinPath(rowPath)
	cKey := joinPath(colPath)

	for i, v := range values {
		if i >= len(m.Data.Cells) {
			break
		}
		k := multiLevelKey{row: rKey, col: cKey, cellIdx: i}
		acc, exists := m.mlAccumulators[k]
		if !exists {
			acc = newAccumulator(m.Data.Cells[i].Function)
			m.mlAccumulators[k] = acc
		}
		raw := v
		f := toFloat(v)
		acc.add(f, raw)
		_ = rNode
		_ = cNode
	}
}

// CellResultMultiLevel returns the aggregate for a cell identified by leaf paths.
func (m *MatrixObject) CellResultMultiLevel(rowPath, colPath []string, cellIdx int) (float64, error) {
	if m.mlAccumulators == nil {
		return 0, fmt.Errorf("matrix: no multi-level data")
	}
	k := multiLevelKey{row: joinPath(rowPath), col: joinPath(colPath), cellIdx: cellIdx}
	acc, ok := m.mlAccumulators[k]
	if !ok {
		return 0, fmt.Errorf("matrix: no data for row=%v col=%v cell=%d", rowPath, colPath, cellIdx)
	}
	return acc.result(), nil
}

// BuildTemplateMultiLevel constructs the table from the multi-level header trees.
//
// The resulting table has:
//   - LevelSize(colRoot) header rows (one per column-header level)
//   - LevelSize(rowRoot) header columns (one per row-header level)
//   - One data row per col-leaf × one data column per row-leaf
func (m *MatrixObject) BuildTemplateMultiLevel() {
	m.ensureMultiLevel()
	// Save template rows/columns before resetting TableBase.
	m.savedTemplateRows = m.TableBase.Rows()
	savedCols := m.TableBase.Columns()
	// Save position — TableBase reset destroys the embedded ComponentBase.
	savedLeft := m.Left()
	savedTop := m.Top()
	savedWidth := m.Width()
	savedHeight := m.Height()
	savedBorder := m.Border()
	savedFixedRows := m.TableBase.FixedRows()
	savedFixedCols := m.TableBase.FixedColumns()

	m.rowRoot.computeSizes()
	m.colRoot.computeSizes()

	rowLeaves := m.rowRoot.leaves()
	colLeaves := m.colRoot.leaves()

	// LevelSize includes the root node itself; subtract 1 to get the number
	// of actual header columns/rows (matching C# MatrixHelper header count).
	nRowHeaderCols := m.rowRoot.LevelSize - 1 // columns used for row-header labels
	if nRowHeaderCols < 1 {
		nRowHeaderCols = 1
	}
	nColHeaderRows := m.colRoot.LevelSize - 1 // rows used for column-header labels
	if nColHeaderRows < 1 {
		nColHeaderRows = 1
	}

	// C# ref: MatrixHelper.UpdateTemplateSizes — CellsSideBySide with multiple
	// cell descriptors adds an extra column header row for cell descriptor names
	// (e.g. "Items Sold", "Revenue") and multiplies body columns by nCells.
	nCells := 1
	sideBySide := m.CellsSideBySide && len(m.Data.Cells) > 1
	origColHeaderRows := nColHeaderRows
	if sideBySide {
		nCells = len(m.Data.Cells)
		nColHeaderRows++ // extra row for cell descriptor names
	}

	// C# ref: MatrixHelper.UpdateTemplateSizes — when ShowTitle, headerHeight++.
	// The title row occupies row 0 in the result table and shifts everything else down.
	titleOffset := 0
	if m.ShowTitle {
		titleOffset = 1
	}

	// Total table columns = row-header columns + (col-leaves × nCells).
	nCols := nRowHeaderCols + len(colLeaves)*nCells
	// Total table rows = title (if any) + col-header rows + one per row-leaf.
	nRows := titleOffset + nColHeaderRows + len(rowLeaves)

	// Reset table and restore position.
	m.TableBase = *table.NewTableBase()
	m.SetLeft(savedLeft)
	m.SetTop(savedTop)
	m.SetWidth(savedWidth)
	m.SetHeight(savedHeight)
	m.SetBorder(savedBorder)
	m.TableBase.SetFixedRows(savedFixedRows)
	m.TableBase.SetFixedColumns(savedFixedCols)

	// Determine if we need a Total column.
	hasTotals := len(m.Data.Rows) > 0 && len(m.Data.Columns) > 0
	if hasTotals {
		nCols += nCells // +nCells for Total columns (1 per cell descriptor)
	}

	// Add columns using template column widths.
	// Template layout: [fixedCols...] [dataColTemplate(×nCells)] [totalColTemplate(×nCells)]
	// Result layout:   [fixedCols...] [dataCols × N × nCells]    [totalCols × nCells]
	//
	// C# MatrixHelper uses AutoSize to measure content for each data column.
	// Go uses the template data column width (Column3) for all expanded data
	// columns, and the template total column width (Column4) for the Total column.
	// Template column helpers for side-by-side.
	// C# ref: MatrixHelper.UpdateCellDescriptors — when CellsSideBySide, each
	// cell descriptor occupies consecutive template columns starting at HeaderWidth.
	// dataCellTemplateCol(i) returns the template column for cell descriptor i.
	// totalCellTemplateCol(i) returns the template column for total cell descriptor i.
	dataCellTemplateCol := func(cellIdx int) int { return nRowHeaderCols + cellIdx }
	totalCellTemplateCol := func(cellIdx int) int { return nRowHeaderCols + nCells + cellIdx }

	lastDataColIdx := nRowHeaderCols + len(colLeaves)*nCells - 1
	totalColStartIdx := nRowHeaderCols + len(colLeaves)*nCells

	for i := 0; i < nCols; i++ {
		col := table.NewTableColumn()
		switch {
		case i < nRowHeaderCols && i < len(savedCols):
			// Fixed row-header columns: use template widths directly.
			col.SetMaxWidth(savedCols[i].MaxWidth())
			col.SetWidth(savedCols[i].Width())
			col.SetAutoSize(savedCols[i].AutoSize())
			col.SetVisible(savedCols[i].Visible())
		case i >= nRowHeaderCols && i <= lastDataColIdx:
			// Data columns: each cell descriptor uses its own template column.
			cellOffset := (i - nRowHeaderCols) % nCells
			tIdx := dataCellTemplateCol(cellOffset)
			if tIdx < len(savedCols) {
				col.SetMaxWidth(savedCols[tIdx].MaxWidth())
				col.SetWidth(savedCols[tIdx].Width())
				col.SetAutoSize(savedCols[tIdx].AutoSize())
			}
		case hasTotals && i >= totalColStartIdx:
			// Total columns: each cell descriptor uses its own total template.
			cellOffset := (i - totalColStartIdx) % nCells
			tIdx := totalCellTemplateCol(cellOffset)
			if tIdx < len(savedCols) {
				col.SetMaxWidth(savedCols[tIdx].MaxWidth())
				col.SetWidth(savedCols[tIdx].Width())
				col.SetAutoSize(savedCols[tIdx].AutoSize())
			}
		}
		m.TableBase.AddColumn(col)
	}

	// ── Build column-header rows ──────────────────────────────────────────────
	// For each column-header level, emit one table row.
	// The left nRowHeaderCols cells are corner cells (empty except for last row).
	// Then one cell per colRoot child at that level (spanning CellSize columns).
	type colBFS struct {
		item  *HeaderItem
		depth int // 0 = colRoot children (level 0)
	}
	// Collect nodes per level via BFS.
	levelNodes := make([][]*HeaderItem, m.colRoot.LevelSize)
	queue := []colBFS{}
	for _, child := range m.colRoot.Children {
		queue = append(queue, colBFS{child, 0})
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		levelNodes[cur.depth] = append(levelNodes[cur.depth], cur.item)
		for _, child := range cur.item.Children {
			queue = append(queue, colBFS{child, cur.depth + 1})
		}
	}

	// hasTotals and nCols already set above with the Total column included.

	savedRows := m.savedTemplateRows

	// Use the method-level helpers for template cell creation.
	templateCell := m.templateCellAt
	fmtVal := m.fmtCellVal
	newCell := m.newStyledCell

	// Helper: create a row with template height/autosize when available.
	newRow := func(templateRowIdx int) *table.TableRow {
		r := table.NewTableRow()
		if templateRowIdx >= 0 && templateRowIdx < len(savedRows) {
			r.SetHeight(savedRows[templateRowIdx].Height())
			r.SetAutoSize(savedRows[templateRowIdx].AutoSize())
		}
		return r
	}

	// Template column indices:
	//   0..nRowHeaderCols-1                   = row-header columns (corner / row labels)
	//   nRowHeaderCols..nRowHeaderCols+nCells-1 = data cell columns (one per cell descriptor)
	//   nRowHeaderCols+nCells..                 = total cell columns (one per cell descriptor)
	// C# ref: MatrixHelper.UpdateDescriptors maps template cells by these offsets.
	dataColTemplate := nRowHeaderCols
	totalColTemplate := nRowHeaderCols + nCells

	// ── Build title row (when ShowTitle is true) ─────────────────────────────
	// C# ref: MatrixHelper.PrintTitle — row 0 contains the title cell spanning
	// all data columns, plus a corner cell spanning all row-header columns.
	if m.ShowTitle {
		titleRow := newRow(0) // title uses template row 0
		// Corner cell(s) of the title row.
		cornerCell := newCell(0, 0) // template: Row0/Col0
		if nRowHeaderCols > 1 {
			cornerCell.SetColSpan(nRowHeaderCols)
		}
		titleRow.AddCell(cornerCell)
		for c := 1; c < nRowHeaderCols; c++ {
			titleRow.AddCell(nil) // spanned placeholders
		}
		// Title cell spanning all data + total columns.
		// C# ref: PrintTitle sets ColSpan = ResultTable.ColumnCount - HeaderWidth.
		titleCell := newCell(0, dataColTemplate)
		if tc := m.templateCellAt(0, dataColTemplate); tc != nil {
			titleCell.SetText(tc.Text())
		}
		titleSpan := nCols - nRowHeaderCols
		if titleSpan > 1 {
			titleCell.SetColSpan(titleSpan)
		}
		titleRow.AddCell(titleCell)
		for c := 1; c < titleSpan; c++ {
			titleRow.AddCell(nil) // spanned placeholders
		}
		m.TableBase.AddRow(titleRow)
	}

	// ── Build column-header rows ──────────────────────────────────────────────
	// C# ref: PrintColumnHeader starts at row ShowTitle ? 1 : 0.
	// Template row index for column headers = titleOffset (0 without title, 1 with).
	colHdrTemplate := titleOffset // template row index for column header cells

	// Resolve EvenStyle fill for column header cells.
	// C# ref: PrintHeaderCell calls templateCell.ApplyEvenStyle() for even-indexed headers.
	var colHdrEvenFill style.Fill
	if tc := templateCell(colHdrTemplate, dataColTemplate); tc != nil && tc.EvenStyleName() != "" && m.StyleLookup != nil {
		se := m.StyleLookup.FindStyle(tc.EvenStyleName())
		if se != nil {
			if se.Fill != nil {
				colHdrEvenFill = se.Fill
			} else if se.FillColor.A > 0 {
				colHdrEvenFill = style.NewSolidFill(se.FillColor)
			}
		}
	}

	for level := 0; level < origColHeaderRows; level++ {
		row := newRow(colHdrTemplate) // header row uses template row after title
		// Corner cells with row descriptor names (e.g. "Employee").
		// C# ref: PrintCorner — corner cell at the last column-header level gets the
		// row descriptor name. When CellsSideBySide, RowSpan covers the cell header row.
		for c := 0; c < nRowHeaderCols; c++ {
			cell := newCell(colHdrTemplate, c) // template: column header row corner
			if level == origColHeaderRows-1 && c < len(m.Data.Rows) {
				if len(m.savedTemplateRows) > colHdrTemplate {
					templateRow := m.savedTemplateRows[colHdrTemplate]
					if templateRow != nil && c < len(templateRow.Cells()) && templateRow.Cells()[c] != nil {
						cell.SetText(templateRow.Cells()[c].Text())
					}
				}
				// C# ref: PrintCorner — RowSpan is clamped to corner area height.
				// When CellsSideBySide, the corner spans the cell header row too.
				if nColHeaderRows-level > 1 {
					cell.SetRowSpan(nColHeaderRows - level)
				}
			}
			row.AddCell(cell)
		}
		// Column header cells for this level.
		// C# ref: PrintColumnHeader — when CellsSideBySide, span is multiplied by nCells.
		evenIdx := 0
		for _, item := range levelNodes[level] {
			cell := newCell(colHdrTemplate, dataColTemplate) // template: column header data cell
			cell.SetText(item.Value)
			// Apply EvenStyle fill for even-indexed column headers.
			if evenIdx%2 != 0 && colHdrEvenFill != nil {
				cell.SetFill(colHdrEvenFill)
			}
			evenIdx++
			span := item.CellSize * nCells // multiply by nCells for side-by-side
			if span > 1 {
				cell.SetColSpan(span)
			}
			row.AddCell(cell)
			for s := 1; s < span; s++ {
				row.AddCell(nil)
			}
		}
		// Total column header.
		// C# ref: PrintColumnHeader — total header at last column level.
		if hasTotals && level == origColHeaderRows-1 {
			cell := newCell(colHdrTemplate, totalColTemplate) // template: total column header
			cell.SetText("Total")
			// Apply EvenStyle to Total header if index is even.
			if evenIdx%2 != 0 && colHdrEvenFill != nil {
				cell.SetFill(colHdrEvenFill)
			}
			if nCells > 1 {
				cell.SetColSpan(nCells)
			}
			row.AddCell(cell)
			for s := 1; s < nCells; s++ {
				row.AddCell(nil)
			}
		}
		m.TableBase.AddRow(row)
	}

	// ── Build cell header row (when CellsSideBySide) ─────────────────────────
	// C# ref: PrintColumnHeader — at leaf level, emits cell descriptor names
	// (e.g. "Items Sold", "Revenue") for each column leaf and total.
	if sideBySide {
		cellHdrTemplate := titleOffset + origColHeaderRows // template row with cell labels
		cellHdrRow := newRow(cellHdrTemplate)
		// Corner cells: nil placeholders (spanned by RowSpan of corner cell above).
		for c := 0; c < nRowHeaderCols; c++ {
			cellHdrRow.AddCell(nil)
		}
		// Cell descriptor labels for each column leaf.
		for range colLeaves {
			for ci := 0; ci < nCells; ci++ {
				tCol := dataCellTemplateCol(ci)
				cell := newCell(cellHdrTemplate, tCol)
				if tc := templateCell(cellHdrTemplate, tCol); tc != nil {
					cell.SetText(tc.Text())
				}
				cellHdrRow.AddCell(cell)
			}
		}
		// Cell descriptor labels for total columns.
		if hasTotals {
			for ci := 0; ci < nCells; ci++ {
				tCol := totalCellTemplateCol(ci)
				cell := newCell(cellHdrTemplate, tCol)
				if tc := templateCell(cellHdrTemplate, tCol); tc != nil {
					cell.SetText(tc.Text())
				}
				cellHdrRow.AddCell(cell)
			}
		}
		m.TableBase.AddRow(cellHdrRow)
	}

	// ── Build data rows ───────────────────────────────────────────────────────
	// Each row-leaf gets one table row.
	// The first nRowHeaderCols cells are row-header label cells.
	// Then one cell per col-leaf.

	// For row-header labels, collect nodes per row-level via BFS on rowRoot.
	type rowBFS struct {
		item  *HeaderItem
		depth int
	}
	rowLevelNodes := make([][]*HeaderItem, m.rowRoot.LevelSize)
	rowQueue := []rowBFS{}
	for _, child := range m.rowRoot.Children {
		rowQueue = append(rowQueue, rowBFS{child, 0})
	}
	for len(rowQueue) > 0 {
		cur := rowQueue[0]
		rowQueue = rowQueue[1:]
		rowLevelNodes[cur.depth] = append(rowLevelNodes[cur.depth], cur.item)
		for _, child := range cur.item.Children {
			rowQueue = append(rowQueue, rowBFS{child, cur.depth + 1})
		}
	}
	// Map leaf → its ancestor path for data lookup.
	buildLeafPaths := func(root *HeaderItem) map[*HeaderItem][]string {
		paths := make(map[*HeaderItem][]string)
		var walk func(node *HeaderItem, path []string)
		walk = func(node *HeaderItem, path []string) {
			p := append(append([]string{}, path...), node.Value)
			if node.isLeaf() {
				paths[node] = p
			}
			for _, child := range node.Children {
				walk(child, p)
			}
		}
		for _, child := range root.Children {
			walk(child, nil)
		}
		return paths
	}
	rowLeafPaths := buildLeafPaths(m.rowRoot)
	colLeafPaths := buildLeafPaths(m.colRoot)

	_ = nRows // total rows already computed

	// Build a per-row-leaf, per-level map so we know which level to emit
	// (and its RowSpan) for grouped row-header cells.
	// For each leaf and each ancestor level, compute how many consecutive
	// leaf rows share the same ancestor node (= RowSpan of that header cell).
	// We track which nodes have already been emitted so we only emit the
	// first occurrence; subsequent cells in the span are nil placeholders.
	emittedRowNode := make(map[*HeaderItem]bool)

	// Track the current level-0 group for subtotal row insertion.
	var prevLevel0Node *HeaderItem
	// groupLeaves collects leaves in the current level-0 group for subtotals.
	var groupLeaves []*HeaderItem

	// Template row index for data rows (after title + column headers).
	// C# ref: UpdateCellDescriptors uses Rows[HeaderHeight] for the data template.
	dataTemplate := titleOffset + nColHeaderRows // typically 1 (no title) or 2 (with title)
	if dataTemplate >= len(savedRows) && len(savedRows) > 0 {
		dataTemplate = len(savedRows) - 2 // fallback: second-to-last row
		if dataTemplate < 0 {
			dataTemplate = 0
		}
	}

	// Resolve EvenStyle fill for alternating data rows.
	// C# ref: MatrixHelper.PrintData applies EvenStyle to even rows.
	var evenFill style.Fill
	if tc := templateCell(dataTemplate, 0); tc != nil && tc.EvenStyleName() != "" && m.StyleLookup != nil {
		se := m.StyleLookup.FindStyle(tc.EvenStyleName())
		if se != nil {
			if se.Fill != nil {
				evenFill = se.Fill
			} else if se.FillColor.A > 0 {
				// Legacy: Fill.Color stored as FillColor (not a Fill interface).
				evenFill = style.NewSolidFill(se.FillColor)
			}
		}
	}

	for ri, rLeaf := range rowLeaves {
		// Detect level-0 group boundary (e.g. year change) and insert subtotal row.
		curLevel0Node := findNodeAtLevel(m.rowRoot, rowLeafPaths[rLeaf], 0)
		if hasTotals && nRowHeaderCols > 1 && prevLevel0Node != nil && curLevel0Node != prevLevel0Node {
			m.addSubtotalRow(prevLevel0Node, groupLeaves, colLeaves, nRowHeaderCols, nCells)
			groupLeaves = nil
		}
		prevLevel0Node = curLevel0Node
		groupLeaves = append(groupLeaves, rLeaf)

		_ = ri
		tableRow := newRow(dataTemplate) // data rows use template row after headers
		rPath := rowLeafPaths[rLeaf]

		// Row-header columns with RowSpan support.
		for level := 0; level < nRowHeaderCols; level++ {
			if level >= len(rPath) {
				tableRow.AddCell(newCell(dataTemplate, level))
				continue
			}
			// Find the HeaderItem at this level in the path.
			node := findNodeAtLevel(m.rowRoot, rPath, level)
			if node == nil || emittedRowNode[node] {
				// Already emitted this header span — insert nil placeholder.
				tableRow.AddCell(nil)
				continue
			}
			// First occurrence — emit the cell with proper RowSpan.
			cell := newCell(dataTemplate, level) // template: row header cell at this column
			cell.SetText(rPath[level])
			span := node.CellSize
			// Add 1 for the subtotal row that follows this group.
			// Applies to all level-0 groups, even those with a single leaf,
			// because the subtotal row is always emitted.
			if hasTotals && nRowHeaderCols > 1 && level == 0 {
				span++
			}
			if span > 1 {
				cell.SetRowSpan(span)
			}
			emittedRowNode[node] = true
			tableRow.AddCell(cell)
		}

		// Data cells — try runtime cell store first, fall back to mlAccumulators.
		// C# ref: MatrixHelper.PrintData — when CellsSideBySide, iterates cell
		// descriptors for each column, placing them in consecutive columns.
		for _, cLeaf := range colLeaves {
			cPath := colLeafPaths[cLeaf]
			for ci := 0; ci < nCells; ci++ {
				tCol := dataCellTemplateCol(ci)
				cell := newCell(dataTemplate, tCol) // template: data cell for descriptor ci
				if len(m.Data.Cells) > 0 {
					var cellText string
					var found bool
					// Try runtime store (populated by GetDataWithCalc).
					rt := m.Data.Runtime()
					if rt != nil {
						raw := rt.GetCellValue(ItemIndex(cLeaf), ItemIndex(rLeaf), ci)
						if raw != nil {
							f := toFloat(raw)
							if f != 0 { // suppress zero values (C# shows empty)
								cellText = fmtVal(cell, f)
								found = true
							}
						}
					}
					// Fall back to mlAccumulators (populated by AddDataMultiLevel).
					if !found {
						val, err := m.CellResultMultiLevel(rPath, cPath, ci)
						if err == nil {
							cellText = fmtVal(cell, val)
							found = true
						}
					}
					if found {
						cell.SetText(cellText)
					}
				}
				tableRow.AddCell(cell)
			}
		}

		// Row total cells (sum across all columns for this row, per cell descriptor).
		// C# ref: PrintData — when CellsSideBySide, emits nCells total cells.
		if hasTotals {
			for ci := 0; ci < nCells; ci++ {
				tCol := totalCellTemplateCol(ci)
				cell := newCell(dataTemplate, tCol) // template: row total for descriptor ci
				rowSum := 0.0
				rt := m.Data.Runtime()
				if rt != nil {
					for _, cLeaf := range colLeaves {
						raw := rt.GetCellValue(ItemIndex(cLeaf), ItemIndex(rLeaf), ci)
						if raw != nil {
							rowSum += toFloat(raw)
						}
					}
				}
				if rowSum != 0 {
					cell.SetText(fmtVal(cell, rowSum))
				}
				tableRow.AddCell(cell)
			}
		}

		// Apply EvenStyle fill to even-indexed data rows.
		// C# ref: MatrixHelper.PrintData applies EvenStyle on even rows.
		if ri%2 == 1 && evenFill != nil {
			for _, cell := range tableRow.Cells() {
				if cell != nil {
					cell.SetFill(evenFill)
				}
			}
		}

		m.TableBase.AddRow(tableRow)
	}

	// Insert subtotal for the last group.
	if hasTotals && nRowHeaderCols > 1 && prevLevel0Node != nil {
		m.addSubtotalRow(prevLevel0Node, groupLeaves, colLeaves, nRowHeaderCols, nCells)
	}

	// Grand total row (sum per column + grand total).
	// C# ref: MatrixHelper uses the last template row (Row7) for the grand total.
	// When CellsSideBySide, emits nCells cells per column leaf.
	if hasTotals {
		grandTotalTemplateRow := len(savedRows) - 1
		totalRow := newRow(grandTotalTemplateRow) // grand total uses last template row
		// Total label spanning all row-header columns.
		labelCell := newCell(grandTotalTemplateRow, 0) // template: grand total label cell
		labelCell.SetText("Total")
		if nRowHeaderCols > 1 {
			labelCell.SetColSpan(nRowHeaderCols)
		}
		totalRow.AddCell(labelCell)
		for c := 1; c < nRowHeaderCols; c++ {
			totalRow.AddCell(nil) // spanned placeholders
		}

		grandTotals := make([]float64, nCells)
		for _, cLeaf := range colLeaves {
			for ci := 0; ci < nCells; ci++ {
				tCol := dataCellTemplateCol(ci)
				cell := newCell(grandTotalTemplateRow, tCol) // template: grand total data cell
				colSum := 0.0
				rt := m.Data.Runtime()
				if rt != nil {
					for _, rLeaf := range rowLeaves {
						raw := rt.GetCellValue(ItemIndex(cLeaf), ItemIndex(rLeaf), ci)
						if raw != nil {
							colSum += toFloat(raw)
						}
					}
				}
				grandTotals[ci] += colSum
				if colSum != 0 {
					cell.SetText(fmtVal(cell, colSum))
				}
				totalRow.AddCell(cell)
			}
		}
		// Grand total cells (one per cell descriptor).
		for ci := 0; ci < nCells; ci++ {
			tCol := totalCellTemplateCol(ci)
			gtCell := newCell(grandTotalTemplateRow, tCol) // template: grand total cell
			if grandTotals[ci] != 0 {
				gtCell.SetText(fmtVal(gtCell, grandTotals[ci]))
			}
			totalRow.AddCell(gtCell)
		}
		m.TableBase.AddRow(totalRow)
	}

	// Auto-size columns based on text content width.
	// Mirrors C# MatrixHelper which uses GDI+ MeasureString to auto-size.
	// Go uses utils.MeasureText which measures text using font.Face metrics.
	if m.AutoSize {
		templateFont := style.DefaultFont()
		if len(savedRows) > 0 && len(savedRows[0].Cells()) > 0 {
			if tc := savedRows[0].Cells()[0]; tc != nil {
				templateFont = tc.Font()
			}
		}

		cols := m.TableBase.Columns()
		allRows := m.TableBase.Rows()
		for ci := range cols {
			maxW := float32(0)
			for _, row := range allRows {
				if ci < len(row.Cells()) && row.Cells()[ci] != nil {
					text := row.Cells()[ci].Text()
					if text != "" {
						cellFont := templateFont
						if tc := row.Cells()[ci]; tc != nil && tc.Font().Name != "" {
							cellFont = tc.Font()
						}
						w, _ := utils.MeasureText(text, cellFont, 0)
						// MeasureText includes GDI+ MeasureString padding
						// (≈ fontSize*96/72/3 = 1/3 em). Subtract it since we
						// add Padding.Horizontal (4px) separately below.
						// C# ref: CalcWidth = MeasureString.Width + Padding.Horizontal.
						// MeasureString already includes GDI pad; we must not double-count.
						gdiPad := cellFont.Size * 96.0 / 72.0 / 3.0
						w -= gdiPad
						if w > maxW {
							maxW = w
						}
					}
				}
			}
			if maxW > 0 {
				// Add cell padding: Padding.Horizontal = left(2) + right(2) = 4.
				// C# CalcWidth: MeasureString width + Padding.Horizontal.
				cols[ci].SetWidth(maxW + 4)
			}
		}

		// Auto-size row heights based on text content.
		// C# CalcHeight measures text at cell width, adds Padding.Vertical + 1.
		// C# ref: TableBase.CalcHeight() (two-pass for non-spanned, then spanned cells).
		cols = m.TableBase.Columns()
		for _, row := range m.TableBase.Rows() {
			if !row.AutoSize() {
				continue
			}
			bestH := float32(0)
			for ci, cell := range row.Cells() {
				if cell == nil {
					continue
				}
				text := cell.Text()
				if text == "" {
					continue
				}
				cellFont := templateFont
				if cell.Font().Name != "" {
					cellFont = cell.Font()
				}
				cellW := float32(0)
				if ci < len(cols) {
					cellW = cols[ci].Width()
				}
				// Compute how many lines are needed. Use MeasureText for
				// the full width including GDI padding, and compare against
				// the full column width (which also includes padding space).
				// This avoids double-counting padding in the comparison.
				fullW, lineH := utils.MeasureText(text, cellFont, 0)
				nLines := 1
				if cellW > 0 && fullW > cellW {
					nLines = int(fullW/cellW) + 1
				}
				h := lineH * float32(nLines)
				// Add padding matching C# CalcSize: Padding.Vertical(2 for TableCell) + 1.
				// The HTML exporter subtracts border widths (bTop/2 + bBottom/2 = 1 for
				// 1px "All" border), so we add an extra 1 to compensate.
				// C# ref: TextObject.CalcSize → size.Height + Padding.Vertical + 1.
				// C# TableCell default padding = (2,1,2,1), Vertical = 2.
				h += 3
				if h > bestH {
					bestH = h
				}
			}
			if bestH > 0 {
				row.SetHeight(bestH)
			}
		}
	}
}

// measureStringApprox approximates C# GDI+ MeasureString for proportional fonts.
// It uses per-glyph width estimates for Tahoma at 8pt (96 DPI), scaled by fontScale.
// The values are calibrated against C# GDI+ MeasureString output.
func measureStringApprox(text string, fontScale float64, fontStyle style.FontStyle) float32 {
	// Per-glyph widths for Tahoma 8pt at 96 DPI, derived from C# GDI+ measurements.
	// Uppercase, lowercase, digits, and common punctuation.
	var glyphWidths = [256]float64{
		// Default: 5.0 for unknown characters
	}
	// Initialize all to a reasonable default
	for i := range glyphWidths {
		glyphWidths[i] = 5.0
	}
	// Uppercase — calibrated against C# GDI+ MeasureString for Tahoma 8pt
	glyphWidths['A'] = 6.5
	glyphWidths['B'] = 6.2
	glyphWidths['C'] = 6.0
	glyphWidths['D'] = 6.7
	glyphWidths['E'] = 5.8
	glyphWidths['F'] = 5.3
	glyphWidths['G'] = 6.7
	glyphWidths['H'] = 6.7
	glyphWidths['I'] = 3.6
	glyphWidths['J'] = 4.3
	glyphWidths['K'] = 6.0
	glyphWidths['L'] = 5.2
	glyphWidths['M'] = 7.7
	glyphWidths['N'] = 6.7
	glyphWidths['O'] = 6.9
	glyphWidths['P'] = 6.0
	glyphWidths['Q'] = 6.9
	glyphWidths['R'] = 6.2
	glyphWidths['S'] = 5.8
	glyphWidths['T'] = 5.7
	glyphWidths['U'] = 6.7
	glyphWidths['V'] = 6.2
	glyphWidths['W'] = 8.6
	glyphWidths['X'] = 6.0
	glyphWidths['Y'] = 5.7
	glyphWidths['Z'] = 5.7
	// Lowercase
	glyphWidths['a'] = 5.1
	glyphWidths['b'] = 5.5
	glyphWidths['c'] = 4.7
	glyphWidths['d'] = 5.5
	glyphWidths['e'] = 5.3
	glyphWidths['f'] = 3.5
	glyphWidths['g'] = 5.5
	glyphWidths['h'] = 5.5
	glyphWidths['i'] = 2.6
	glyphWidths['j'] = 3.1
	glyphWidths['k'] = 5.1
	glyphWidths['l'] = 2.6
	glyphWidths['m'] = 8.3
	glyphWidths['n'] = 5.5
	glyphWidths['o'] = 5.5
	glyphWidths['p'] = 5.5
	glyphWidths['q'] = 5.5
	glyphWidths['r'] = 3.7
	glyphWidths['s'] = 4.5
	glyphWidths['t'] = 3.7
	glyphWidths['u'] = 5.5
	glyphWidths['v'] = 5.1
	glyphWidths['w'] = 7.1
	glyphWidths['x'] = 5.1
	glyphWidths['y'] = 5.1
	glyphWidths['z'] = 4.7
	// Digits and punctuation
	glyphWidths['0'] = 5.2
	glyphWidths['1'] = 5.2
	glyphWidths['2'] = 5.2
	glyphWidths['3'] = 5.2
	glyphWidths['4'] = 5.2
	glyphWidths['5'] = 5.2
	glyphWidths['6'] = 5.2
	glyphWidths['7'] = 5.2
	glyphWidths['8'] = 5.2
	glyphWidths['9'] = 5.2
	glyphWidths[' '] = 2.8
	glyphWidths['.'] = 2.8
	glyphWidths[','] = 2.8
	glyphWidths[':'] = 2.8
	glyphWidths['$'] = 5.5

	// Non-ASCII glyph widths (currency symbols, etc.)
	var nonASCII = map[rune]float64{
		'₱': 7.0, // Philippine Peso
		'€': 7.0, // Euro
		'¥': 6.5, // Yen/Yuan
		'£': 6.0, // Pound
		'₹': 7.0, // Rupee
		'₽': 7.0, // Ruble
		'₩': 8.0, // Won
		'₺': 6.5, // Turkish Lira
		'₫': 7.5, // Dong
		'₴': 7.0, // Hryvnia
	}

	total := 0.0
	for _, ch := range text {
		idx := int(ch)
		if idx >= 0 && idx < 256 {
			total += glyphWidths[idx]
		} else if w, ok := nonASCII[ch]; ok {
			total += w
		} else {
			total += 5.5 // non-ASCII default
		}
	}
	total *= fontScale
	// Bold adds ~10% width
	if fontStyle&style.FontStyleBold != 0 {
		total *= 1.1
	}
	return float32(total)
}

// findNodeAtLevel traverses root to find the *HeaderItem node at the given
// depth that matches the path prefix path[0..level]. Returns nil if not found.
func findNodeAtLevel(root *HeaderItem, path []string, level int) *HeaderItem {
	if level >= len(path) || root == nil {
		return nil
	}
	// Walk from root through the first 'level' path segments.
	cur := root
	for d := 0; d <= level; d++ {
		found := false
		for _, child := range cur.Children {
			if child.Value == path[d] {
				cur = child
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return cur
}

// templateCellAt returns the template cell at (rowIdx, colIdx) from savedTemplateRows,
// or nil if out of bounds.
func (m *MatrixObject) templateCellAt(rowIdx, colIdx int) *table.TableCell {
	if rowIdx >= 0 && rowIdx < len(m.savedTemplateRows) {
		row := m.savedTemplateRows[rowIdx]
		if row != nil && colIdx < len(row.Cells()) {
			return row.Cells()[colIdx]
		}
	}
	return nil
}

// newStyledCell creates a new table cell with visual properties copied from the
// template cell at (rowIdx, colIdx). Mirrors C# RunTimeAssign: copies font, fill,
// border, alignment, format.
func (m *MatrixObject) newStyledCell(rowIdx, colIdx int) *table.TableCell {
	c := table.NewTableCell()
	if tc := m.templateCellAt(rowIdx, colIdx); tc != nil {
		c.SetFont(tc.Font())
		c.SetFill(tc.Fill())
		c.SetBorder(tc.Border())
		c.SetHorzAlign(tc.HorzAlign())
		c.SetVertAlign(tc.VertAlign())
		c.SetTextColor(tc.TextColor())
		if tc.Format() != nil {
			c.SetFormat(tc.Format())
		}
	}
	return c
}

// fmtCellVal formats a value using the cell's format, falling back to m.cellFormat.
// When there are multiple cell descriptors, each descriptor's template cell carries
// its own format (or no format), so the blanket m.cellFormat fallback is suppressed
// to avoid applying e.g. Currency to an "Items Sold" column that has no format.
func (m *MatrixObject) fmtCellVal(cell *table.TableCell, v any) string {
	if f := cell.Format(); f != nil {
		if _, isGen := f.(*format.GeneralFormat); !isGen {
			return f.FormatValue(v)
		}
	}
	// Only use the blanket cellFormat fallback for single-cell matrices where
	// all data cells share one template. For multi-cell matrices, the per-cell
	// template already has the correct format (or no format for plain numbers).
	if m.cellFormat != nil && len(m.Data.Cells) <= 1 {
		return m.cellFormat.FormatValue(v)
	}
	return fmt.Sprintf("%g", v)
}

// addSubtotalRow inserts a subtotal row for a level-0 group (e.g. per-year subtotal).
// It sums the cell values for all leaves in the group across all column leaves.
// nCells is the number of cell descriptors (>1 when CellsSideBySide).
// Template row index 2 (Row5) provides styling for the subtotal cells.
// C# ref: MatrixHelper.PrintRowHeader — total items use TemplateTotalRow cells.
func (m *MatrixObject) addSubtotalRow(_ *HeaderItem, groupLeaves, colLeaves []*HeaderItem, nRowHeaderCols, nCells int) {
	row := table.NewTableRow()
	// Use subtotal template row height (Row5 = index 2 in template).
	if len(m.savedTemplateRows) > 2 {
		row.SetHeight(m.savedTemplateRows[2].Height())
		row.SetAutoSize(m.savedTemplateRows[2].AutoSize())
	}
	// First cell: empty (year cell already has RowSpan covering this row).
	row.AddCell(nil)
	// Second cell: "Total" label — styled from template Cell6 (row 2, col 1).
	if nRowHeaderCols > 1 {
		cell := m.newStyledCell(2, 1) // template: Row5/Cell6
		cell.SetText("Total")
		row.AddCell(cell)
	}
	for c := 2; c < nRowHeaderCols; c++ {
		row.AddCell(m.newStyledCell(2, c))
	}
	// Template column helpers for subtotal cells.
	dataCellCol := func(ci int) int { return nRowHeaderCols + ci }
	totalCellCol := func(ci int) int { return nRowHeaderCols + nCells + ci }
	// Subtotal per column — one cell per cell descriptor per column leaf.
	subtotalGrands := make([]float64, nCells)
	rt := m.Data.Runtime()
	for _, cLeaf := range colLeaves {
		for ci := 0; ci < nCells; ci++ {
			cell := m.newStyledCell(2, dataCellCol(ci))
			colSum := 0.0
			if rt != nil {
				for _, rLeaf := range groupLeaves {
					raw := rt.GetCellValue(ItemIndex(cLeaf), ItemIndex(rLeaf), ci)
					if raw != nil {
						colSum += toFloat(raw)
					}
				}
			}
			subtotalGrands[ci] += colSum
			if colSum != 0 {
				cell.SetText(m.fmtCellVal(cell, colSum))
			}
			row.AddCell(cell)
		}
	}
	// Subtotal grand total — one per cell descriptor.
	for ci := 0; ci < nCells; ci++ {
		gtCell := m.newStyledCell(2, totalCellCol(ci))
		if subtotalGrands[ci] != 0 {
			gtCell.SetText(m.fmtCellVal(gtCell, subtotalGrands[ci]))
		}
		row.AddCell(gtCell)
	}
	m.TableBase.AddRow(row)
}

// joinPath concatenates path segments with a zero-byte separator for a unique key.
func joinPath(path []string) string {
	var buf []byte
	for i, seg := range path {
		if i > 0 {
			buf = append(buf, 0)
		}
		buf = append(buf, seg...)
	}
	return string(buf)
}
