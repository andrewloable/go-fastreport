package matrix

import (
	"fmt"

	"github.com/andrewloable/go-fastreport/table"
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
	m.rowRoot.computeSizes()
	m.colRoot.computeSizes()

	rowLeaves := m.rowRoot.leaves()
	colLeaves := m.colRoot.leaves()

	nRowHeaderCols := m.rowRoot.LevelSize // columns used for row-header labels
	nColHeaderRows := m.colRoot.LevelSize // rows used for column-header labels

	// Total table columns = row-header columns + one per col-leaf.
	nCols := nRowHeaderCols + len(colLeaves)
	// Total table rows = col-header rows + one per row-leaf.
	nRows := nColHeaderRows + len(rowLeaves)

	// Reset table.
	m.TableBase = *table.NewTableBase()

	// Add columns.
	for i := 0; i < nCols; i++ {
		m.TableBase.AddColumn(table.NewTableColumn())
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

	for level := 0; level < nColHeaderRows; level++ {
		row := table.NewTableRow()
		// Corner cells.
		for c := 0; c < nRowHeaderCols; c++ {
			cell := table.NewTableCell()
			if level == nColHeaderRows-1 && c == 0 {
				cell.SetText("") // corner
			}
			row.AddCell(cell)
		}
		// Column header cells for this level.
		for _, item := range levelNodes[level] {
			cell := table.NewTableCell()
			cell.SetText(item.Value)
			// TODO: apply ColSpan = item.CellSize when table supports colspan.
			row.AddCell(cell)
			// Fill remaining span cells with empty (for flat table rendering).
			for s := 1; s < item.CellSize; s++ {
				row.AddCell(table.NewTableCell())
			}
		}
		m.TableBase.AddRow(row)
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

	_ = rowLevelNodes // used for colspan annotation in a future iteration
	_ = nRows         // total rows already computed

	for _, rLeaf := range rowLeaves {
		tableRow := table.NewTableRow()
		rPath := rowLeafPaths[rLeaf]

		// Row-header columns: emit only the leaf path values for now.
		// For multi-level rendering, emit ancestor labels in earlier rows
		// (merging is a renderer concern; here we fill the deepest label).
		for level := 0; level < nRowHeaderCols; level++ {
			cell := table.NewTableCell()
			if level < len(rPath) {
				if level == len(rPath)-1 {
					cell.SetText(rPath[level])
				}
			}
			tableRow.AddCell(cell)
		}

		// Data cells.
		for _, cLeaf := range colLeaves {
			cPath := colLeafPaths[cLeaf]
			cell := table.NewTableCell()
			if len(m.Data.Cells) > 0 {
				val, err := m.CellResultMultiLevel(rPath, cPath, 0)
				if err == nil {
					cell.SetText(fmt.Sprintf("%g", val))
				}
			}
			tableRow.AddCell(cell)
		}

		m.TableBase.AddRow(tableRow)
	}
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
