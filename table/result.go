package table

// TableResult is the computed layout of a table prepared for rendering.
// It extends TableBase with rendering callbacks and skip logic used by
// the report engine.
//
// Do not use this type directly — it is created internally by the engine
// when rendering a TableObject (or when InvokeManualBuild returns a result).
//
// It is the Go equivalent of FastReport.Table.TableResult.
type TableResult struct {
	TableBase

	// Skip instructs the engine to skip rendering this result
	// (e.g. when the table is invisible or already rendered).
	Skip bool

	// AfterCalcBounds is called after the engine computes the table bounds.
	// You may use this callback to adjust automatically calculated
	// row/column sizes (e.g. to fit the table on a page).
	AfterCalcBounds func(result *TableResult)
}

// NewTableResult creates an empty TableResult.
func NewTableResult() *TableResult {
	return &TableResult{
		TableBase: *NewTableBase(),
	}
}

// NewTableResultFrom creates a TableResult pre-populated from a source TableBase.
// The rows and columns are copied by reference (shallow copy of the slice headers).
func NewTableResultFrom(src *TableBase) *TableResult {
	r := NewTableResult()
	// Shallow-copy rows and columns so the result can grow independently.
	r.rows = append(r.rows, src.rows...)
	r.columns = append(r.columns, src.columns...)
	r.fixedRows = src.fixedRows
	r.fixedColumns = src.fixedColumns
	r.layout = src.layout
	r.repeatHeaders = src.repeatHeaders
	r.repeatRowHeaders = src.repeatRowHeaders
	r.repeatColumnHeaders = src.repeatColumnHeaders
	r.adjustSpannedCellsWidth = src.adjustSpannedCellsWidth
	return r
}

// CalcBounds computes the total width and height of the result table in pixels.
// It calls AfterCalcBounds (if set) after computing the dimensions.
// Returns (totalWidth, totalHeight).
func (r *TableResult) CalcBounds() (float32, float32) {
	var totalW, totalH float32
	for _, col := range r.columns {
		totalW += col.Width()
	}
	for _, row := range r.rows {
		totalH += row.Height()
	}
	if r.AfterCalcBounds != nil {
		r.AfterCalcBounds(r)
	}
	return totalW, totalH
}

// RowsToSerialize returns all rows that should be included in the output.
// In this implementation all rows are returned; the engine may filter further.
func (r *TableResult) RowsToSerialize() []*TableRow { return r.rows }

// ColumnsToSerialize returns all columns that should be included in the output.
func (r *TableResult) ColumnsToSerialize() []*TableColumn { return r.columns }

// ── TableStyleCollection ──────────────────────────────────────────────────────

// TableStyleCollection deduplicates TableCell style entries used during
// table rendering. It is the Go equivalent of FastReport.Table.TableStyleCollection.
//
// It is an internal helper — do not use it directly.
type TableStyleCollection struct {
	styles []*TableCell
	def    *TableCell
}

// NewTableStyleCollection creates an empty collection with a default cell style.
func NewTableStyleCollection() *TableStyleCollection {
	return &TableStyleCollection{
		def: NewTableCell(),
	}
}

// DefaultStyle returns the default cell style.
func (sc *TableStyleCollection) DefaultStyle() *TableCell { return sc.def }

// Add ensures the given cell style exists in the collection and returns the
// canonical copy. If an equivalent style already exists it is returned;
// otherwise a clone is stored and returned.
func (sc *TableStyleCollection) Add(src *TableCell) *TableCell {
	for _, existing := range sc.styles {
		if cellStylesEqual(existing, src) {
			return existing
		}
	}
	clone := cloneCell(src)
	sc.styles = append(sc.styles, clone)
	return clone
}

// Count returns the number of unique styles.
func (sc *TableStyleCollection) Count() int { return len(sc.styles) }

// Get returns the style at index i.
func (sc *TableStyleCollection) Get(i int) *TableCell {
	if i < 0 || i >= len(sc.styles) {
		return nil
	}
	return sc.styles[i]
}

// cellStylesEqual returns true when two cells have equivalent visual styling
// (border, fill, font). Text/span fields are not compared.
func cellStylesEqual(a, b *TableCell) bool {
	if a.HorzAlign() != b.HorzAlign() || a.VertAlign() != b.VertAlign() {
		return false
	}
	// Compare font names (simple equivalence for deduplication).
	fa, fb := a.Font(), b.Font()
	if fa.Name != fb.Name || fa.Size != fb.Size || fa.Style != fb.Style {
		return false
	}
	return true
}
