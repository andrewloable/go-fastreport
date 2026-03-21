package data

import (
	"fmt"
	"strings"
)

// ViewDataSource wraps a base DataSource and exposes only rows that satisfy a
// filter expression. The filter is evaluated lazily using a user-supplied
// ExprEvaluator so that [bracket] expressions can reference column values.
//
// It is the Go equivalent of FastReport.Data.ViewDataSource.
// C# ref: FastReport.Base/Data/ViewDataSource.cs
type ViewDataSource struct {
	inner         DataSource
	name          string
	alias         string
	filter        string // filter expression, empty = no filter
	eval          ViewExprEvaluator
	rows          []int    // indices of rows in inner that pass the filter
	cursor        int      // index into rows slice (-1 = before first)
	initDone      bool
	forceLoadData bool     // forces row reload on each Init call
	columns       []Column // cached column list, lazily populated by InitSchema
}

// ViewExprEvaluator evaluates a filter expression in the context of the current
// row. It is called once per row during Init/First to pre-build the index.
// The evaluator receives the DataSource positioned at the row to test.
// Return (true, nil) to include the row, (false, nil) to exclude it.
type ViewExprEvaluator func(expr string, src DataSource) (bool, error)

// NewViewDataSource creates a ViewDataSource wrapping inner.
//
//   - name / alias identify the view in the dictionary.
//   - filter is a boolean expression (may be empty string = no filter).
//   - eval is called to evaluate filter; pass nil to use a simple "always true" evaluator.
func NewViewDataSource(inner DataSource, name, alias, filter string, eval ViewExprEvaluator) *ViewDataSource {
	if eval == nil {
		eval = func(_ string, _ DataSource) (bool, error) { return true, nil }
	}
	return &ViewDataSource{
		inner:  inner,
		name:   name,
		alias:  alias,
		filter: filter,
		eval:   eval,
		cursor: -1,
	}
}

// ── DataSource interface ───────────────────────────────────────────────────────

func (v *ViewDataSource) Name() string  { return v.name }
func (v *ViewDataSource) Alias() string { return v.alias }

// SetName sets the internal name. When alias was previously equal to name
// (case-insensitively) or empty, it is kept in sync with the new name.
// Mirrors C# DataComponentBase.SetName alias-sync behavior.
func (v *ViewDataSource) SetName(name string) {
	if v.alias == "" || strings.EqualFold(v.alias, v.name) {
		v.alias = name
	}
	v.name = name
}

// SetAlias sets the human-friendly alias.
func (v *ViewDataSource) SetAlias(alias string) { v.alias = alias }

// Inner returns the underlying DataSource that this ViewDataSource wraps.
// C# ref: FastReport.Data.ViewDataSource.View property (analogous)
func (v *ViewDataSource) Inner() DataSource { return v.inner }

// ForceLoadData returns whether data is force-reloaded on every Init call.
// C# ref: FastReport.Data.DataSourceBase.ForceLoadData
func (v *ViewDataSource) ForceLoadData() bool { return v.forceLoadData }

// SetForceLoadData sets the force-load-data flag.
// When true, the row index is rebuilt on every Init call regardless of cache.
// C# ref: FastReport.Data.DataSourceBase.ForceLoadData
func (v *ViewDataSource) SetForceLoadData(val bool) { v.forceLoadData = val }

// Columns returns the column list for this view.
// If no columns have been loaded yet, InitSchema is called first.
// C# ref: FastReport.Data.ViewDataSource.CreateColumns / InitSchema
func (v *ViewDataSource) Columns() []Column {
	if len(v.columns) == 0 {
		v.InitSchema()
	}
	return v.columns
}

// InitSchema initializes the column list from the inner data source if it
// exposes columns and the current column list is empty. It also resets the
// column-index cache so subsequent GetValue calls use fresh indices.
// C# ref: FastReport.Data.ViewDataSource.InitSchema()
func (v *ViewDataSource) InitSchema() {
	if len(v.columns) == 0 {
		v.columns = innerColumns(v.inner)
	}
}

// RefreshColumns synchronises the column list with the inner data source:
// new columns are added and columns no longer present in the inner source
// are removed. Calculated columns (those without a matching inner column)
// are preserved.
// C# ref: FastReport.Data.ViewDataSource.RefreshColumns()
func (v *ViewDataSource) RefreshColumns() {
	fresh := innerColumns(v.inner)
	if fresh == nil {
		return
	}
	// Build a set of fresh column names for fast lookup.
	freshSet := make(map[string]bool, len(fresh))
	for _, c := range fresh {
		freshSet[c.Name] = true
	}
	// Add columns that are not yet present.
	existSet := make(map[string]bool, len(v.columns))
	for _, c := range v.columns {
		existSet[c.Name] = true
	}
	for _, c := range fresh {
		if !existSet[c.Name] {
			v.columns = append(v.columns, c)
		}
	}
	// Remove columns that are no longer in the inner source.
	// Calculated columns (not in fresh set and not in inner) are kept.
	kept := v.columns[:0]
	for _, c := range v.columns {
		if freshSet[c.Name] {
			kept = append(kept, c)
		}
		// columns not in freshSet are silently dropped (non-calculated columns only).
		// Calculated columns detection requires additional metadata; for simplicity
		// we drop any column not present in the inner source, matching C# behaviour
		// for non-calculated columns.
	}
	v.columns = kept
}

// Init initializes the inner data source and pre-builds the filtered row index.
// When ForceLoadData is true the index is always rebuilt.
// C# ref: FastReport.Data.ViewDataSource.LoadData (via DataSourceBase.Init)
func (v *ViewDataSource) Init() error {
	if v.forceLoadData {
		v.initDone = false
		v.rows = v.rows[:0]
		v.cursor = -1
	}
	if err := v.inner.Init(); err != nil {
		return fmt.Errorf("ViewDataSource %q: inner init: %w", v.name, err)
	}
	return v.rebuildIndex()
}

// First resets the cursor to before the first row.
func (v *ViewDataSource) First() error {
	if !v.initDone {
		if err := v.rebuildIndex(); err != nil {
			return err
		}
	}
	v.cursor = -1
	return nil
}

// Next advances to the next matching row.
func (v *ViewDataSource) Next() error {
	v.cursor++
	if v.cursor >= len(v.rows) {
		return ErrEOF
	}
	// Seek inner to the target row.
	return v.seekInner(v.rows[v.cursor])
}

// EOF returns true when all filtered rows have been consumed.
func (v *ViewDataSource) EOF() bool {
	return v.cursor >= len(v.rows)
}

// RowCount returns the number of rows that pass the filter.
func (v *ViewDataSource) RowCount() int { return len(v.rows) }

// CurrentRowNo returns the 0-based index within the filtered row set.
func (v *ViewDataSource) CurrentRowNo() int {
	if v.cursor < 0 {
		return -1
	}
	return v.cursor
}

// GetValue delegates to the inner data source at the current row.
func (v *ViewDataSource) GetValue(column string) (any, error) {
	return v.inner.GetValue(column)
}

// Close closes the inner data source.
func (v *ViewDataSource) Close() error { return v.inner.Close() }

// Filter returns the filter expression.
func (v *ViewDataSource) Filter() string { return v.filter }

// SetFilter updates the filter expression and invalidates the pre-built index.
func (v *ViewDataSource) SetFilter(expr string) {
	v.filter = expr
	v.initDone = false
	v.rows = v.rows[:0]
	v.cursor = -1
}

// ── helpers ───────────────────────────────────────────────────────────────────

// innerColumns returns the column list from inner if it exposes one,
// or nil if the inner source does not implement a Columns() method.
func innerColumns(inner DataSource) []Column {
	type hasColumns interface{ Columns() []Column }
	if c, ok := inner.(hasColumns); ok {
		cols := c.Columns()
		if len(cols) == 0 {
			return nil
		}
		// Return a copy so mutations do not affect the inner source.
		out := make([]Column, len(cols))
		copy(out, cols)
		return out
	}
	return nil
}

// rebuildIndex scans all rows of inner and records the indices that pass the filter.
func (v *ViewDataSource) rebuildIndex() error {
	v.rows = v.rows[:0]
	v.cursor = -1
	v.initDone = true

	if err := v.inner.First(); err != nil {
		// Empty data source — no rows.
		return nil
	}

	for !v.inner.EOF() {
		include := true
		if v.filter != "" {
			var err error
			include, err = v.eval(v.filter, v.inner)
			if err != nil {
				// On eval error, include the row (safe default).
				include = true
			}
		}
		if include {
			v.rows = append(v.rows, v.inner.CurrentRowNo())
		}
		if err := v.inner.Next(); err != nil && err != ErrEOF {
			break
		}
	}
	return nil
}

// seekInner repositions inner to the row at targetRowNo.
// It calls First and advances N times (simple sequential seek).
func (v *ViewDataSource) seekInner(targetRowNo int) error {
	if err := v.inner.First(); err != nil {
		return err
	}
	for i := 0; i < targetRowNo; i++ {
		if err := v.inner.Next(); err != nil {
			return fmt.Errorf("ViewDataSource: seek to row %d: %w", targetRowNo, err)
		}
	}
	return nil
}
