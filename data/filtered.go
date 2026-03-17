package data

import "fmt"

// filterKey pairs a child column name with the parent value it must equal.
type filterKey struct {
	column string
	value  any
}

// FilteredDataSource wraps a DataSource and iterates only those rows whose
// specified columns equal specified parent values. It is used by the engine
// to implement master-detail relation filtering: before rendering each parent
// row, the engine creates (or updates) a FilteredDataSource for each child
// DataBand so that the child only sees rows whose join-key columns match the
// current parent row values.
type FilteredDataSource struct {
	inner  DataSource
	keys   []filterKey
	rows   []int // indices into inner's rows that pass the filter
	cursor int   // position in rows slice
}

// NewFilteredDataSource creates a FilteredDataSource wrapping inner.
// keys maps child column names → required values (equality check).
//
// The constructor calls inner.First() internally to scan all rows, so inner
// must already be initialized and positioned at start (or First() called).
func NewFilteredDataSource(inner DataSource, childColumns, parentValues []string) (*FilteredDataSource, error) {
	f := &FilteredDataSource{
		inner:  inner,
		cursor: -1,
	}
	for i, col := range childColumns {
		var val any
		if i < len(parentValues) {
			val = parentValues[i]
		}
		f.keys = append(f.keys, filterKey{column: col, value: val})
	}
	if err := f.rebuildIndex(); err != nil {
		return nil, err
	}
	return f, nil
}

// rebuildIndex scans all rows of inner and records indices of matching rows.
func (f *FilteredDataSource) rebuildIndex() error {
	f.rows = f.rows[:0]
	f.cursor = -1

	if err := f.inner.First(); err != nil {
		// Empty data source — no rows.
		return nil
	}

	for !f.inner.EOF() {
		if f.rowMatches() {
			f.rows = append(f.rows, f.inner.CurrentRowNo())
		}
		if err := f.inner.Next(); err != nil {
			break
		}
	}
	return nil
}

// rowMatches returns true when the current inner row satisfies all key filters.
func (f *FilteredDataSource) rowMatches() bool {
	for _, k := range f.keys {
		val, err := f.inner.GetValue(k.column)
		if err != nil {
			return false
		}
		// Compare as strings for simplicity (covers int/float cases too).
		if fmt.Sprint(val) != fmt.Sprint(k.value) {
			return false
		}
	}
	return true
}

// seekInner positions inner at the row index stored in f.rows[f.cursor].
func (f *FilteredDataSource) seekInner() error {
	if f.cursor < 0 || f.cursor >= len(f.rows) {
		return nil
	}
	target := f.rows[f.cursor]
	if err := f.inner.First(); err != nil {
		return err
	}
	for f.inner.CurrentRowNo() < target {
		if err := f.inner.Next(); err != nil {
			return err
		}
	}
	return nil
}

// ── DataSource interface ───────────────────────────────────────────────────────

// Inner returns the underlying (unfiltered) DataSource that this
// FilteredDataSource wraps. Used by relation-matching code to unwrap the
// filter layer when comparing data source pointers.
func (f *FilteredDataSource) Inner() DataSource { return f.inner }

// Name delegates to inner.
func (f *FilteredDataSource) Name() string { return f.inner.Name() }

// Alias delegates to inner.
func (f *FilteredDataSource) Alias() string { return f.inner.Alias() }

// Init is a no-op for FilteredDataSource since the index is built at
// construction time. Calling inner.Init() would reset the inner source
// and lose the cursor position needed for row-level access.
func (f *FilteredDataSource) Init() error { return nil }

// First positions at the first matching row.
func (f *FilteredDataSource) First() error {
	if len(f.rows) == 0 {
		f.cursor = 0
		return ErrEOF
	}
	f.cursor = 0
	return f.seekInner()
}

// Next advances to the next matching row.
func (f *FilteredDataSource) Next() error {
	f.cursor++
	if f.cursor >= len(f.rows) {
		return ErrEOF
	}
	return f.seekInner()
}

// EOF returns true when all matching rows have been consumed.
func (f *FilteredDataSource) EOF() bool { return f.cursor >= len(f.rows) }

// RowCount returns the number of rows that pass the filter.
func (f *FilteredDataSource) RowCount() int { return len(f.rows) }

// CurrentRowNo returns the cursor position within the filtered rows (0-based).
func (f *FilteredDataSource) CurrentRowNo() int { return f.cursor }

// GetValue delegates to inner for the current row.
func (f *FilteredDataSource) GetValue(column string) (any, error) {
	return f.inner.GetValue(column)
}

// Columns returns the column list from inner if it exposes one.
func (f *FilteredDataSource) Columns() []Column {
	type hasColumns interface{ Columns() []Column }
	if c, ok := f.inner.(hasColumns); ok {
		return c.Columns()
	}
	return nil
}

// Close delegates to inner.
func (f *FilteredDataSource) Close() error { return f.inner.Close() }
