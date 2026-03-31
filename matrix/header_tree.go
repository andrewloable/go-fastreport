package matrix

import (
	"fmt"
	"math"
	"sort"

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
	// Save position and identity — TableBase reset destroys the embedded ComponentBase,
	// clearing Name, Left, Top, Width, Height, and Border.
	// C# ref: MatrixHelper creates a new ResultTable but keeps the MatrixObject identity.
	savedName := m.Name()
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

	// Reset table and restore position and identity.
	m.TableBase = *table.NewTableBase()
	m.SetName(savedName) // restore name cleared by TableBase reset
	m.SetLeft(savedLeft)
	m.SetTop(savedTop)
	m.SetWidth(savedWidth)
	m.SetHeight(savedHeight)
	m.SetBorder(savedBorder)
	m.TableBase.SetFixedRows(savedFixedRows)
	m.TableBase.SetFixedColumns(savedFixedCols)

	// Determine if we need a Grand Total row (hasTotals used for grand total row below).
	// Total column leaves are already included in colLeaves after AddTotalItems in
	// SyncRuntimeToMultiLevel, so no extra nCols addition is needed here.
	hasTotals := len(m.Data.Rows) > 0 && len(m.Data.Columns) > 0

	// Add columns using template column widths.
	// Template layout: [rowHdr] [dataCol×nCells] [perGroupTotal×nCells] [grandTotal×nCells]
	// C# ref: MatrixHelper.UpdateCellDescriptors — template col offsets are:
	//   HeaderWidth+ci (data), HeaderWidth+CellCount+ci (group total),
	//   HeaderWidth+CellCount*2+ci (grand total).
	dataCellTemplateCol := func(cellIdx int) int { return nRowHeaderCols + cellIdx }
	totalCellTemplateCol := func(cellIdx int) int { return nRowHeaderCols + nCells + cellIdx }
	grandTotalCellTemplateCol := func(cellIdx int) int { return nRowHeaderCols + nCells*2 + cellIdx }

	// templateColForLeaf returns the template column for colLeaves[leafIdx], cell ci.
	// Grand Total leaves (parent == colRoot) use grandTotalCellTemplateCol.
	// Other Total leaves use totalCellTemplateCol.  Data leaves use dataCellTemplateCol.
	templateColForLeaf := func(leafIdx, ci int) int {
		if leafIdx < 0 || leafIdx >= len(colLeaves) {
			return dataCellTemplateCol(ci)
		}
		leaf := colLeaves[leafIdx]
		if !ItemIsTotal(leaf) {
			return dataCellTemplateCol(ci)
		}
		if ItemParent(leaf) == m.colRoot {
			return grandTotalCellTemplateCol(ci)
		}
		return totalCellTemplateCol(ci)
	}

	for i := 0; i < nCols; i++ {
		col := table.NewTableColumn()
		if i < nRowHeaderCols && i < len(savedCols) {
			// Fixed row-header columns: use template widths directly.
			col.SetMaxWidth(savedCols[i].MaxWidth())
			col.SetWidth(savedCols[i].Width())
			col.SetAutoSize(savedCols[i].AutoSize())
			col.SetVisible(savedCols[i].Visible())
		} else if i >= nRowHeaderCols {
			// Data / Total columns: select template width by leaf type.
			leafIdx := (i - nRowHeaderCols) / nCells
			cellOffset := (i - nRowHeaderCols) % nCells
			tIdx := templateColForLeaf(leafIdx, cellOffset)
			if tIdx < len(savedCols) {
				col.SetMaxWidth(savedCols[tIdx].MaxWidth())
				col.SetWidth(savedCols[tIdx].Width())
				col.SetAutoSize(savedCols[tIdx].AutoSize())
			}
		}
		m.TableBase.AddColumn(col)
	}

	// ── Build column-header rows ──────────────────────────────────────────────
	// Pre-compute per-node layout info for the column header rows.
	// Each entry records the node's output-column start/span, depth, and whether
	// it is a leaf.  Leaf nodes at depth < origColHeaderRows-1 (e.g. the Grand Total
	// column at depth=0 in a 2-level hierarchy) need RowSpan in the header; their
	// column positions at deeper levels need nil placeholders.
	type colHdrEmit struct {
		item        *HeaderItem
		outputStart int // 0-based output column index among data columns
		span        int // CellSize (number of data columns this node covers)
		depth       int
		isLeaf      bool
	}
	var allColEmits []colHdrEmit
	var buildColEmits func(node *HeaderItem, depth, start int)
	buildColEmits = func(node *HeaderItem, depth, start int) {
		allColEmits = append(allColEmits, colHdrEmit{
			item:        node,
			outputStart: start,
			span:        node.CellSize,
			depth:       depth,
			isLeaf:      len(node.Children) == 0,
		})
		cs := start
		for _, child := range node.Children {
			buildColEmits(child, depth+1, cs)
			cs += child.CellSize
		}
	}
	{
		cs := 0
		for _, child := range m.colRoot.Children {
			buildColEmits(child, 0, cs)
			cs += child.CellSize
		}
	}

	// hasTotals and nCols already set above.

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
		// Use the template row matching this header level for correct styling.
		// C# ref: PrintColumnHeader uses template row (titleOffset+level) for each level.
		levelHdrTemplate := titleOffset + level
		row := newRow(levelHdrTemplate)
		// Corner cells: C# PrintCorner places the row descriptor name at level=0
		// with RowSpan covering all nColHeaderRows (including the cell-descriptor row
		// when CellsSideBySide).  Levels > 0 emit nil placeholders for the span.
		for c := 0; c < nRowHeaderCols; c++ {
			if level == 0 {
				cell := newCell(levelHdrTemplate, c)
				if c < len(m.Data.Rows) {
					if len(m.savedTemplateRows) > levelHdrTemplate {
						templateRow := m.savedTemplateRows[levelHdrTemplate]
						if templateRow != nil && c < len(templateRow.Cells()) && templateRow.Cells()[c] != nil {
							cell.SetText(templateRow.Cells()[c].Text())
						}
					}
				}
				if nColHeaderRows > 1 {
					cell.SetRowSpan(nColHeaderRows)
				}
				row.AddCell(cell)
			} else {
				row.AddCell(nil) // spanned by corner at level 0
			}
		}

		// Collect column-header cells for this level, in output-column order.
		//   nodes at depth==level        → emit actual cell
		//   leaf nodes at depth<level    → emit nil placeholder(s) for spanned position
		//   nodes at depth>level         → skipped (will be emitted at their own level)
		type levelCell struct {
			outputStart int
			span        int
			emit        *colHdrEmit // nil ⇒ emit nil placeholders
		}
		var levelCells []levelCell
		for i := range allColEmits {
			e := &allColEmits[i]
			switch {
			case e.depth == level:
				levelCells = append(levelCells, levelCell{e.outputStart, e.span, e})
			case e.isLeaf && e.depth < level:
				levelCells = append(levelCells, levelCell{e.outputStart, e.span, nil})
			}
		}
		sort.Slice(levelCells, func(i, j int) bool {
			return levelCells[i].outputStart < levelCells[j].outputStart
		})

		evenIdx := 0
		for _, lc := range levelCells {
			span := lc.span * nCells
			if lc.emit == nil {
				// Nil placeholders for a leaf spanned from a shallower level.
				for s := 0; s < span; s++ {
					row.AddCell(nil)
				}
				continue
			}
			e := lc.emit
			// Choose template column based on item type.
			// Grand Total items (parent == colRoot) use grandTotalCellTemplateCol;
			// other Total items use totalColTemplate; data items use dataColTemplate.
			tCol := dataColTemplate
			if ItemIsTotal(e.item) {
				if ItemParent(e.item) == m.colRoot {
					tCol = grandTotalCellTemplateCol(0)
				} else {
					tCol = totalColTemplate
				}
			}
			cell := newCell(levelHdrTemplate, tCol)
			// For Total items, use the template cell text ("Total") rather than
			// item.Value (which holds the parent's key like "2011").
			// C# ref: PrintColumnHeader sets the cell text to the template cell text for totals.
			if ItemIsTotal(e.item) {
				if tc := templateCell(levelHdrTemplate, tCol); tc != nil && tc.Text() != "" {
					cell.SetText(tc.Text())
				} else {
					cell.SetText("Total")
				}
			} else {
				cell.SetText(e.item.Value)
			}
			if evenIdx%2 != 0 && colHdrEvenFill != nil {
				cell.SetFill(colHdrEvenFill)
			}
			evenIdx++
			if span > 1 {
				cell.SetColSpan(span)
			}
			// Leaf items at non-deepest levels need RowSpan to span remaining header rows.
			// C# ref: PrintColumnHeader — Total/leaf items at level<origColHeaderRows-1
			// receive RowSpan so they span the remaining column-header rows.
			if e.isLeaf && level < origColHeaderRows-1 {
				rowSpan := origColHeaderRows - level
				if sideBySide {
					rowSpan++ // also covers the cell-descriptor row
				}
				cell.SetRowSpan(rowSpan)
			}
			row.AddCell(cell)
			for s := 1; s < span; s++ {
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
		// Cell descriptor labels for each column leaf (including Total leaves).
		for leafIdx, cLeaf := range colLeaves {
			for ci := 0; ci < nCells; ci++ {
				_ = cLeaf
				tCol := templateColForLeaf(leafIdx, ci)
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

	// rowIdx tracks the flat row index across data rows, subtotals, and grand total.
	// Mirrors C# Matrix.RowIndex incremented in PrintData for every rendered row leaf.
	rowIdx := 0

	for ri, rLeaf := range rowLeaves {
		// Detect level-0 group boundary (e.g. year change) and insert subtotal row.
		curLevel0Node := findNodeAtLevel(m.rowRoot, rowLeafPaths[rLeaf], 0)
		if hasTotals && nRowHeaderCols > 1 && prevLevel0Node != nil && curLevel0Node != prevLevel0Node {
			// Set RowIndex before addSubtotalRow so highlight conditions are evaluated
			// with the correct index.
			m.RowIndex = rowIdx
			m.addSubtotalRow(prevLevel0Node, groupLeaves, colLeaves, nRowHeaderCols, nCells)
			rowIdx++
			groupLeaves = nil
		}
		prevLevel0Node = curLevel0Node
		groupLeaves = append(groupLeaves, rLeaf)

		// Set RowIndex for this data row before building cells.
		// C# ref: Matrix.RowIndex is set in PrintData before templateCell.GetData().
		m.RowIndex = rowIdx
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
		// For Total column leaves (added by AddTotalItems), aggregate non-Total siblings.
		// C# ref: MatrixHelper.PrintData — when CellsSideBySide, iterates cell
		// descriptors for each column, placing them in consecutive columns.
		for leafIdx, cLeaf := range colLeaves {
			cPath := colLeafPaths[cLeaf]
			isTotalCol := ItemIsTotal(cLeaf)
			for ci := 0; ci < nCells; ci++ {
				tCol := templateColForLeaf(leafIdx, ci)
				cell := newCell(dataTemplate, tCol)
				var rawNumVal float64
				if len(m.Data.Cells) > 0 {
					var cellText string
					var found bool
					rt := m.Data.Runtime()
					if rt != nil {
						if isTotalCol {
							// Aggregate across all non-Total data leaves under cLeaf's parent.
							// C# ref: MatrixHelper computes totals by summing terminal items.
							parent := ItemParent(cLeaf)
							if parent != nil {
								for _, dl := range getNonTotalTerminalItems(parent) {
									raw := rt.GetCellValue(ItemIndex(dl), ItemIndex(rLeaf), ci)
									if raw != nil {
										rawNumVal += toFloat(raw)
									}
								}
							}
							if rawNumVal != 0 {
								cellText = fmtVal(cell, rawNumVal)
								found = true
							}
						} else {
							raw := rt.GetCellValue(ItemIndex(cLeaf), ItemIndex(rLeaf), ci)
							if raw != nil {
								rawNumVal = toFloat(raw)
								if rawNumVal != 0 {
									cellText = fmtVal(cell, rawNumVal)
									found = true
								}
							}
						}
					}
					// Fall back to mlAccumulators (populated by AddDataMultiLevel).
					// Total leaves have no mlAccumulators entry; skip fallback for them.
					if !found && !isTotalCol {
						val, err := m.CellResultMultiLevel(rPath, cPath, ci)
						if err == nil {
							rawNumVal = toFloat(val)
							cellText = fmtVal(cell, val)
							found = true
						}
					}
					if found {
						cell.SetText(cellText)
					}
				}
				// Apply highlight conditions with the raw numeric value as context.
				m.CurrentCellValue = rawNumVal
				m.applyHighlights(cell)
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

		// Apply highlight conditions for each cell in this row.
		// C# ref: templateCell.GetData() in PrintData evaluates highlights with current RowIndex.
		for _, cell := range tableRow.Cells() {
			if cell != nil {
				m.applyHighlights(cell)
			}
		}

		m.TableBase.AddRow(tableRow)
		rowIdx++
	}

	// Insert subtotal for the last group.
	if hasTotals && nRowHeaderCols > 1 && prevLevel0Node != nil {
		m.RowIndex = rowIdx
		m.addSubtotalRow(prevLevel0Node, groupLeaves, colLeaves, nRowHeaderCols, nCells)
		rowIdx++
	}

	// Grand total row (sum per column + grand total).
	// C# ref: MatrixHelper uses the last template row (Row7) for the grand total.
	// When CellsSideBySide, emits nCells cells per column leaf.
	if hasTotals {
		// Set RowIndex for the grand total row.
		m.RowIndex = rowIdx
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

		// One cell per column leaf (data or Total).
		// For Total column leaves, aggregate non-Total siblings.
		for leafIdx, cLeaf := range colLeaves {
			for ci := 0; ci < nCells; ci++ {
				tCol := templateColForLeaf(leafIdx, ci)
				cell := newCell(grandTotalTemplateRow, tCol)
				colSum := 0.0
				rt := m.Data.Runtime()
				if rt != nil {
					if ItemIsTotal(cLeaf) {
						// Aggregate across non-Total data leaves under parent.
						parent := ItemParent(cLeaf)
						if parent != nil {
							for _, dl := range getNonTotalTerminalItems(parent) {
								for _, rLeaf := range rowLeaves {
									raw := rt.GetCellValue(ItemIndex(dl), ItemIndex(rLeaf), ci)
									if raw != nil {
										colSum += toFloat(raw)
									}
								}
							}
						}
					} else {
						for _, rLeaf := range rowLeaves {
							raw := rt.GetCellValue(ItemIndex(cLeaf), ItemIndex(rLeaf), ci)
							if raw != nil {
								colSum += toFloat(raw)
							}
						}
					}
				}
				if colSum != 0 {
					cell.SetText(fmtVal(cell, colSum))
				}
				totalRow.AddCell(cell)
			}
		}
		// Apply highlight conditions for the grand total row.
		for _, cell := range totalRow.Cells() {
			if cell != nil {
				m.applyHighlights(cell)
			}
		}
		m.TableBase.AddRow(totalRow)
	}

	// Auto-size columns based on text content width.
	// Mirrors C# TableBase.CalcWidth which uses GDI+ MeasureString (first pass:
	// non-spanned cells only; second pass: spanned cells expand last column).
	// C# formula: column.Width = max(cell.CalcWidth())
	//   where CalcWidth = MeasureString(text) + Padding.Horizontal + 1
	//   and MeasureString is GDI+ GenericDefault (includes trailing whitespace pad).
	// We approximate GDI+ MeasureString with measureStringApprox() and round to
	// the nearest integer, matching C# integer column widths.
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
					cell := row.Cells()[ci]
					// C# CalcWidth first pass: skip cells with ColSpan > 1 (handled in second pass).
					if cell.ColSpan() > 1 {
						continue
					}
					text := cell.Text()
					if text != "" {
						cellFont := templateFont
						if cell.Font().Name != "" {
							cellFont = cell.Font()
						}
						// measureStringApprox approximates C# GDI+ MeasureString output.
						// fontScale = size/8 since glyph widths are calibrated at Tahoma 8pt.
						// C# ref: TextObject.CalcSize → g.MeasureString(text, font, SizeF).
						fontScale := float64(cellFont.Size) / 8.0
						w := measureStringApprox(text, fontScale, cellFont.Style)
						if w > maxW {
							maxW = w
						}
					}
				}
			}
			if maxW > 0 {
				// C# CalcWidth: MeasureString + Padding.Horizontal + 1 = approx + 4 + 1 = approx + 5.
				// GDI+ MeasureString includes ~0.6px trailing whitespace not in our glyph widths,
				// so we use +5.5 instead of +5 to match C# column widths before border adjustment.
				// Empirically: data cols C#=56 (CSS 55), name col C#=87 (CSS 86), total C#=62 (CSS 61).
				cols[ci].SetWidth(float32(math.Round(float64(maxW) + 5.5)))
			}
			// Prevent table.CalcWidth() from overriding the matrix-computed widths
			// with macOS system font metrics (which differ from Windows GDI+).
			// C# CalcWidth already ran here; mark AutoSize=false so the engine's
			// subsequent CalcWidth() call preserves the widths we just set.
			cols[ci].SetAutoSize(false)
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
	// Uppercase — calibrated against C# GDI+ MeasureString for Tahoma 8pt.
	// Values are ~1.06× larger than raw font advances to match GDI+ leading/trailing
	// space included by Graphics.MeasureString with StringFormat.GenericDefault.
	// Calibration: "Steven Buchanan" must total 81.0 so that round(81+5)=86 matches C#.
	glyphWidths['A'] = 6.9
	glyphWidths['B'] = 6.6
	glyphWidths['C'] = 6.4
	glyphWidths['D'] = 7.1
	glyphWidths['E'] = 6.15
	glyphWidths['F'] = 5.6
	glyphWidths['G'] = 7.1
	glyphWidths['H'] = 7.1
	glyphWidths['I'] = 3.8
	glyphWidths['J'] = 4.6
	glyphWidths['K'] = 6.4
	glyphWidths['L'] = 5.5
	glyphWidths['M'] = 8.2
	glyphWidths['N'] = 7.1
	glyphWidths['O'] = 7.3
	glyphWidths['P'] = 6.4
	glyphWidths['Q'] = 7.3
	glyphWidths['R'] = 6.6
	glyphWidths['S'] = 6.15
	glyphWidths['T'] = 6.0
	glyphWidths['U'] = 7.1
	glyphWidths['V'] = 6.6
	glyphWidths['W'] = 9.1
	glyphWidths['X'] = 6.4
	glyphWidths['Y'] = 6.0
	glyphWidths['Z'] = 6.0
	// Lowercase
	glyphWidths['a'] = 5.4
	glyphWidths['b'] = 5.8
	glyphWidths['c'] = 5.0
	glyphWidths['d'] = 5.8
	glyphWidths['e'] = 5.6
	glyphWidths['f'] = 3.7
	glyphWidths['g'] = 5.8
	glyphWidths['h'] = 5.8
	glyphWidths['i'] = 2.8
	glyphWidths['j'] = 3.3
	glyphWidths['k'] = 5.4
	glyphWidths['l'] = 2.8
	glyphWidths['m'] = 8.8
	glyphWidths['n'] = 5.8
	glyphWidths['o'] = 5.8
	glyphWidths['p'] = 5.8
	glyphWidths['q'] = 5.8
	glyphWidths['r'] = 3.9
	glyphWidths['s'] = 4.8
	glyphWidths['t'] = 3.9
	glyphWidths['u'] = 5.8
	glyphWidths['v'] = 5.4
	glyphWidths['w'] = 7.5
	glyphWidths['x'] = 5.4
	glyphWidths['y'] = 5.4
	glyphWidths['z'] = 5.0
	// Digits and punctuation
	// Calibrated against C# GDI+ MeasureString for Tahoma 8pt on Windows.
	// Digits: "₱1,900.00" must total 50.0 so that round(50+5)=55 matches C#.
	glyphWidths['0'] = 6.0
	glyphWidths['1'] = 6.0
	glyphWidths['2'] = 6.0
	glyphWidths['3'] = 6.0
	glyphWidths['4'] = 6.0
	glyphWidths['5'] = 6.0
	glyphWidths['6'] = 6.0
	glyphWidths['7'] = 6.0
	glyphWidths['8'] = 6.0
	glyphWidths['9'] = 6.0
	glyphWidths[' '] = 3.0
	glyphWidths['.'] = 3.0
	glyphWidths[','] = 3.0
	glyphWidths[':'] = 3.0
	glyphWidths['$'] = 5.8

	// Non-ASCII glyph widths (currency symbols, etc.)
	// Calibrated so that measureStringApprox + 5 matches C# CalcWidth output.
	var nonASCII = map[rune]float64{
		'₱': 8.4, // Philippine Peso
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
	// Bold adds ~14% width (calibrated against C# GDI+ Tahoma Bold MeasureString).
	if fontStyle&style.FontStyleBold != 0 {
		total *= 1.14
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
// border, alignment, format, and highlight conditions.
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
		// Copy highlight conditions so they can be evaluated during BuildTemplateMultiLevel.
		// C# ref: RunTimeAssign copies highlight conditions; GetData() evaluates them.
		if h := tc.Highlights(); len(h) > 0 {
			hCopy := make([]style.HighlightCondition, len(h))
			copy(hCopy, h)
			c.SetHighlights(hCopy)
		}
	}
	return c
}

// applyHighlights evaluates the highlight conditions on a cell using the current
// RowIndex/ColumnIndex. The first matching condition is applied (fill, text color,
// font). Mirrors C# templateCell.GetData() called in PrintData/PrintRowHeader.
// The highlights are cleared after evaluation since the style is pre-baked.
func (m *MatrixObject) applyHighlights(cell *table.TableCell) {
	if m.HighlightCalc == nil {
		return
	}
	highlights := cell.Highlights()
	if len(highlights) == 0 {
		return
	}
	for _, cond := range highlights {
		result, err := m.HighlightCalc(cond.Expression)
		if err != nil {
			continue
		}
		matched, _ := result.(bool)
		if !matched {
			continue
		}
		if !cond.Visible {
			break
		}
		if cond.ApplyFill && cond.Fill != nil {
			cell.SetFill(cond.Fill.Clone())
		}
		if cond.ApplyTextFill {
			cell.SetTextColor(cond.TextFillColor)
		}
		if cond.ApplyFont {
			cell.SetFont(cond.Font)
		}
		break
	}
	// Clear highlights — the result is pre-baked into the cell's visual properties.
	cell.SetHighlights(nil)
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
	// Apply highlight conditions using the current m.RowIndex (set by the caller).
	// C# ref: PrintRowHeader calls templateCell.GetData() for subtotal rows too.
	for _, cell := range row.Cells() {
		if cell != nil {
			m.applyHighlights(cell)
		}
	}
	m.TableBase.AddRow(row)
}

// joinPath concatenates path segments with a zero-byte separator for a unique key.
// getNonTotalTerminalItems returns all leaf items under node that are NOT IsTotal.
// Used to compute aggregates for Total column/row leaves: a Total leaf's value is
// the sum of its non-Total siblings' terminal items.
// C# ref: MatrixHelper aggregates by summing GetTerminalItems that are not IsTotal.
func getNonTotalTerminalItems(node *HeaderItem) []*HeaderItem {
	var result []*HeaderItem
	collectNonTotalTerminals(node, &result)
	return result
}

func collectNonTotalTerminals(node *HeaderItem, out *[]*HeaderItem) {
	if len(node.Children) == 0 {
		if !ItemIsTotal(node) {
			*out = append(*out, node)
		}
		return
	}
	for _, child := range node.Children {
		collectNonTotalTerminals(child, out)
	}
}

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
