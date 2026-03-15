package data

import "fmt"

// ViewDataSource wraps a base DataSource and exposes only rows that satisfy a
// filter expression. The filter is evaluated lazily using a user-supplied
// ExprEvaluator so that [bracket] expressions can reference column values.
//
// It is the Go equivalent of FastReport.Data.ViewDataSource.
type ViewDataSource struct {
	inner     DataSource
	name      string
	alias     string
	filter    string // filter expression, empty = no filter
	eval      ViewExprEvaluator
	rows      []int // indices of rows in inner that pass the filter
	cursor    int   // index into rows slice (-1 = before first)
	initDone  bool
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

// Init initializes the inner data source and pre-builds the filtered row index.
func (v *ViewDataSource) Init() error {
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
