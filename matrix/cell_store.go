package matrix

// cell_store.go implements the indexed cell-value store used by the runtime
// matrix pipeline.
//
// C# source: FastReport.Base/Matrix/MatrixCells.cs
//            FastReport.Base/Matrix/MatrixData.cs (AddValue / GetValue / SetValue)
//
// Gaps implemented (issues go-fastreport-29qpl and go-fastreport-y439i):
//   - CellStore type — a 3-D value store [cellIndex][rowIndex][colIndex]
//     matching the C# rows[][] layout in MatrixCells.
//   - MatrixData.AddValue(colVals, rowVals, cellVals) — public C# API
//   - MatrixData.GetValue(colIdx, rowIdx, cellIdx) — index-based lookup
//   - MatrixData.SetValue(colIdx, rowIdx, cellVal) — index-based write
//   - MatrixData.GetValues(colIdx, rowIdx, cellIdx) — returns all raw values
//   - MatrixData.SetValues(colIdx, rowIdx, cellVals) — replaces values
//   - MatrixData.IsEmpty — true until first AddValue call
//
// The CellStore maps [cellIndex][rowIndex][colIndex] → value or []any.
// Matches the C# layout: rows is a slice of row-slices, one per cell descriptor.

// CellStore is the 3-D sparse value store for matrix data cells.
// Each outer index corresponds to one CellDescriptor.
// C# source: MatrixCells rows field (List<ArrayList>[]).
type CellStore struct {
	// data[cellIdx][rowIdx][colIdx] holds the accumulated value(s).
	// A cell entry is nil (no data), a single any value, or []any for multiple values.
	data [][][]any // [cellIdx][rowIdx][colIdx]
	// nCells is the number of cell descriptors (set on first use).
	nCells int
}

// newCellStore creates an empty CellStore for nCells cell descriptors.
func newCellStore(nCells int) *CellStore {
	return &CellStore{
		nCells: nCells,
		data:   make([][][]any, nCells),
	}
}

// isEmpty reports whether no values have been added to the store.
// C# source: MatrixCells.IsEmpty (rows == null).
func (cs *CellStore) isEmpty() bool {
	for _, cellRows := range cs.data {
		if cellRows != nil {
			return false
		}
	}
	return true
}

// ensureIndices grows the data slice for cellIdx so that rowIdx and colIdx are
// accessible. Mirrors the C# CheckIndices method in MatrixCells.
func (cs *CellStore) ensureIndices(cellIdx, rowIdx, colIdx int) {
	// Ensure cellIdx row-slice exists.
	if cs.data[cellIdx] == nil {
		cs.data[cellIdx] = make([][]any, rowIdx+1)
	}
	for rowIdx >= len(cs.data[cellIdx]) {
		cs.data[cellIdx] = append(cs.data[cellIdx], nil)
	}
	// Ensure colIdx within that row.
	if cs.data[cellIdx][rowIdx] == nil {
		cs.data[cellIdx][rowIdx] = make([]any, colIdx+1)
	}
	for colIdx >= len(cs.data[cellIdx][rowIdx]) {
		cs.data[cellIdx][rowIdx] = append(cs.data[cellIdx][rowIdx], nil)
	}
}

// addRaw accumulates a single raw value into the cell at [cellIdx][rowIdx][colIdx].
// Mirrors the private addValue method in C# MatrixCells.
// Null (nil) values are ignored.
func (cs *CellStore) addRaw(cellIdx, rowIdx, colIdx int, value any) {
	if value == nil {
		return
	}
	cs.ensureIndices(cellIdx, rowIdx, colIdx)
	existing := cs.data[cellIdx][rowIdx][colIdx]
	switch prev := existing.(type) {
	case nil:
		cs.data[cellIdx][rowIdx][colIdx] = value
	case []any:
		cs.data[cellIdx][rowIdx][colIdx] = append(prev, value)
	default:
		cs.data[cellIdx][rowIdx][colIdx] = []any{prev, value}
	}
}

// setRaw replaces the value at [cellIdx][rowIdx][colIdx].
// Mirrors the private setValue method in C# MatrixCells.
func (cs *CellStore) setRaw(cellIdx, rowIdx, colIdx int, value any) {
	cs.ensureIndices(cellIdx, rowIdx, colIdx)
	cs.data[cellIdx][rowIdx][colIdx] = value
}

// GetValue returns the raw value (single or []any) for the cell.
// Returns nil if the cell is empty or indices are out of range.
// C# source: MatrixCells.GetValue(columnIndex, rowIndex, cellIndex).
func (cs *CellStore) GetValue(colIdx, rowIdx, cellIdx int) any {
	if cellIdx >= len(cs.data) || cs.data[cellIdx] == nil {
		return nil
	}
	if rowIdx >= len(cs.data[cellIdx]) || cs.data[cellIdx][rowIdx] == nil {
		return nil
	}
	row := cs.data[cellIdx][rowIdx]
	if colIdx >= len(row) {
		return nil
	}
	return row[colIdx]
}

// GetValues returns all accumulated values for the cell as []any.
// Returns nil if the cell is empty.
// C# source: MatrixCells.GetValues(columnIndex, rowIndex, cellIndex).
func (cs *CellStore) GetValues(colIdx, rowIdx, cellIdx int) []any {
	v := cs.GetValue(colIdx, rowIdx, cellIdx)
	if v == nil {
		return nil
	}
	if vs, ok := v.([]any); ok {
		return vs
	}
	return []any{v}
}

// AddValues accumulates one set of cell values (one per cell descriptor).
// Nil and "empty" values are ignored (matching C# DBNull / null check).
// C# source: MatrixCells.AddValue(int columnIndex, int rowIndex, object[] value).
func (cs *CellStore) AddValues(colIdx, rowIdx int, values []any) {
	if len(values) != cs.nCells {
		return // mismatch — ignore rather than panic (matches silent C# null handling)
	}
	for i, v := range values {
		cs.addRaw(i, rowIdx, colIdx, v)
	}
}

// SetValues replaces the values for all cell descriptors at the given indices.
// All-nil value sets are ignored.
// C# source: MatrixCells.SetValues(int columnIndex, int rowIndex, object[] cellValues).
func (cs *CellStore) SetValues(colIdx, rowIdx int, values []any) {
	// Ignore all-nil.
	allNil := true
	for _, v := range values {
		if v != nil {
			allNil = false
			break
		}
	}
	if allNil {
		return
	}
	for i, v := range values {
		cs.setRaw(i, rowIdx, colIdx, v)
	}
}

// ── MatrixData runtime additions ──────────────────────────────────────────────

// matrixDataRuntime holds the runtime state for MatrixData: two MatrixHeader
// trees and the cell store. It is initialised lazily on the first AddValue call.
// We keep it separate so it does not conflict with the existing MatrixData fields.
type matrixDataRuntime struct {
	columns *MatrixHeader
	rows    *MatrixHeader
	cells   *CellStore
}

// runtime returns the matrixDataRuntime for d, creating it if needed.
// The Descriptors slices are synchronised from d.Columns and d.Rows.
func (d *MatrixData) runtime() *matrixDataRuntime {
	if d.rt == nil {
		d.rt = &matrixDataRuntime{
			columns: NewMatrixHeader(),
			rows:    NewMatrixHeader(),
		}
	}
	// Always sync Descriptors so Find uses the current sort/totals config.
	d.rt.columns.Descriptors = d.Columns
	d.rt.rows.Descriptors = d.Rows
	if d.rt.cells == nil || d.rt.cells.nCells != len(d.Cells) {
		d.rt.cells = newCellStore(len(d.Cells))
	}
	return d.rt
}

// AddValue adds a set of values to the matrix using typed object arrays.
// This is the C# MatrixData.AddValue(object[] columnValues, object[] rowValues,
// object[] cellValues) public API.
//
// columnValues length must match Data.Columns; rowValues must match Data.Rows;
// cellValues must match Data.Cells.
//
// C# source: FastReport.Base/Matrix/MatrixData.cs, AddValue overload (calls with dataRowNo=0).
func (d *MatrixData) AddValue(columnValues, rowValues, cellValues []any) {
	d.AddValueAt(columnValues, rowValues, cellValues, 0)
}

// AddValueAt is the equivalent of the C# AddValue(columnValues, rowValues,
// cellValues, dataRowNo) overload. dataRowNo is the data source row number
// stored on newly created header items for SaveState/RestoreState engine use.
//
// C# source: FastReport.Base/Matrix/MatrixData.cs, AddValue(object[], object[], object[], int).
func (d *MatrixData) AddValueAt(columnValues, rowValues, cellValues []any, dataRowNo int) {
	rt := d.runtime()

	col := rt.columns.Find(columnValues, true, dataRowNo)
	row := rt.rows.Find(rowValues, true, dataRowNo)
	if col == nil || row == nil {
		return
	}
	rt.cells.AddValues(ItemIndex(col), ItemIndex(row), cellValues)
}

// GetValue returns the aggregated value for the cell at (colIdx, rowIdx, cellIdx).
// C# source: FastReport.Base/Matrix/MatrixData.cs, GetValue(int, int, int).
func (d *MatrixData) GetValue(colIdx, rowIdx, cellIdx int) any {
	if d.rt == nil || d.rt.cells == nil {
		return nil
	}
	return d.rt.cells.GetValue(colIdx, rowIdx, cellIdx)
}

// GetValues returns all raw values for the cell at (colIdx, rowIdx, cellIdx).
// C# source: FastReport.Base/Matrix/MatrixData.cs, GetValues (internal).
func (d *MatrixData) GetValues(colIdx, rowIdx, cellIdx int) []any {
	if d.rt == nil || d.rt.cells == nil {
		return nil
	}
	return d.rt.cells.GetValues(colIdx, rowIdx, cellIdx)
}

// SetValue replaces the value for the first cell descriptor at (colIdx, rowIdx).
// C# source: FastReport.Base/Matrix/MatrixData.cs, SetValue(int, int, object).
func (d *MatrixData) SetValue(colIdx, rowIdx int, cellValue any) {
	if d.rt == nil {
		return
	}
	d.rt.cells.SetValues(colIdx, rowIdx, []any{cellValue})
}

// SetValues replaces the values for all cell descriptors at (colIdx, rowIdx).
// C# source: FastReport.Base/Matrix/MatrixData.cs, SetValues (internal).
func (d *MatrixData) SetValues(colIdx, rowIdx int, cellValues []any) {
	if d.rt == nil {
		return
	}
	d.rt.cells.SetValues(colIdx, rowIdx, cellValues)
}

// IsEmpty reports whether no data has been added yet.
// C# source: MatrixData.IsEmpty (delegates to MatrixCells.IsEmpty).
func (d *MatrixData) IsEmpty() bool {
	if d.rt == nil || d.rt.cells == nil {
		return true
	}
	return d.rt.cells.isEmpty()
}

// RuntimeReset resets the runtime header trees and cell store, clearing all
// accumulated data. Call this before a new data-fill pass (StartPrint).
// C# source: MatrixData.Clear() calls Columns.Reset(), Rows.Reset(), Cells.Reset().
func (d *MatrixData) RuntimeReset() {
	if d.rt == nil {
		return
	}
	d.rt.columns.Reset()
	d.rt.rows.Reset()
	d.rt.cells = newCellStore(len(d.Cells))
}

// ColumnHeader returns the MatrixHeader for the column axis (created on demand).
func (d *MatrixData) ColumnHeader() *MatrixHeader {
	return d.runtime().columns
}

// RowHeader returns the MatrixHeader for the row axis (created on demand).
func (d *MatrixData) RowHeader() *MatrixHeader {
	return d.runtime().rows
}
