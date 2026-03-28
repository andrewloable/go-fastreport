// Package data provides data binding types for go-fastreport.
// It is the Go equivalent of FastReport.Data namespace.
package data

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// DataSource is the primary interface for all data providers.
// It is the Go equivalent of DataSourceBase with a simplified Go-idiomatic API.
type DataSource interface {
	// Name returns the data source name.
	Name() string
	// Alias returns the human-friendly alias.
	Alias() string
	// Init initializes the data source and loads data.
	Init() error
	// First positions at the first row (equivalent to CurrentRowNo = 0).
	First() error
	// Next advances to the next row. Returns ErrEOF at end.
	Next() error
	// EOF returns true when all rows have been consumed.
	EOF() bool
	// RowCount returns the total number of rows.
	RowCount() int
	// CurrentRowNo returns the 0-based index of the current row.
	CurrentRowNo() int
	// GetValue returns the value of the named column in the current row.
	// Returns nil if the column does not exist.
	GetValue(column string) (any, error)
	// Close releases resources held by the data source.
	Close() error
}

// ErrEOF is returned by Next() when the data source has no more rows.
var ErrEOF = fmt.Errorf("data source: no more rows (EOF)")

// ErrNotInitialized is returned when accessing a datasource that hasn't been Init()-ed.
var ErrNotInitialized = fmt.Errorf("data source: not initialized")

// Column describes a single column in a data source.
type Column struct {
	// Name is the column name.
	Name string
	// Alias is the human-friendly display name.
	Alias string
	// DataType is a string describing the column type (e.g. "string", "int", "float64").
	DataType string
}

// AdditionalFilterPredicate is a function that returns true if the row value passes
// the filter for a specific column. Mirrors C# DataSourceFilter.ValueMatch().
type AdditionalFilterPredicate func(value any) bool

// BaseDataSource provides a reusable implementation of the DataSource interface
// backed by an in-memory slice of row maps. Concrete data sources can embed it
// and override GetValue to supply their own row storage.
type BaseDataSource struct {
	name             string
	alias            string
	columns          []Column
	rows             []map[string]any
	currentRow       int
	initialized      bool
	additionalFilter map[string]AdditionalFilterPredicate
}

// NewBaseDataSource creates a BaseDataSource with the given name.
func NewBaseDataSource(name string) *BaseDataSource {
	return &BaseDataSource{
		name:       name,
		alias:      name,
		currentRow: -1,
	}
}

// Name returns the data source name.
func (ds *BaseDataSource) Name() string { return ds.name }

// SetName sets the data source name. When alias was previously equal to name
// (case-insensitively) or empty, it is kept in sync with the new name.
// Mirrors C# DataComponentBase.SetName alias-sync behavior.
func (ds *BaseDataSource) SetName(n string) {
	if ds.alias == "" || strings.EqualFold(ds.alias, ds.name) {
		ds.alias = n
	}
	ds.name = n
}

// Alias returns the display alias.
func (ds *BaseDataSource) Alias() string { return ds.alias }

// SetAlias sets the display alias.
func (ds *BaseDataSource) SetAlias(a string) { ds.alias = a }

// Columns returns the column descriptors.
func (ds *BaseDataSource) Columns() []Column { return ds.columns }

// AddColumn adds a column descriptor.
func (ds *BaseDataSource) AddColumn(col Column) {
	ds.columns = append(ds.columns, col)
}

// AddRow appends a row to the internal row store.
func (ds *BaseDataSource) AddRow(row map[string]any) {
	ds.rows = append(ds.rows, row)
}

// Init marks the data source as initialized and resets the position.
func (ds *BaseDataSource) Init() error {
	ds.initialized = true
	ds.currentRow = -1
	return nil
}

// First positions at the first row.
func (ds *BaseDataSource) First() error {
	if !ds.initialized {
		return ErrNotInitialized
	}
	if len(ds.rows) == 0 {
		ds.currentRow = -1
		return ErrEOF
	}
	ds.currentRow = 0
	return nil
}

// Next advances to the next row.
func (ds *BaseDataSource) Next() error {
	if !ds.initialized {
		return ErrNotInitialized
	}
	ds.currentRow++
	if ds.currentRow >= len(ds.rows) {
		ds.currentRow = len(ds.rows)
		return ErrEOF
	}
	return nil
}

// EOF returns true when the cursor is past the last row.
func (ds *BaseDataSource) EOF() bool {
	return ds.currentRow >= len(ds.rows)
}

// BOF returns true when the cursor is before the first row (not yet positioned).
// Mirrors C# DataSourceBase.BOF property (DataSourceBase.cs):
// CurrentRowNo < 0.
func (ds *BaseDataSource) BOF() bool {
	return ds.currentRow < 0
}

// HasMoreRows returns true when the cursor is positioned at a valid row and
// there are more rows to consume (CurrentRowNo < RowCount).
// Mirrors C# DataSourceBase.HasMoreRows property (DataSourceBase.cs lines 90-93):
// CurrentRowNo < RowCount.
func (ds *BaseDataSource) HasMoreRows() bool {
	return ds.currentRow >= 0 && ds.currentRow < len(ds.rows)
}

// RowCount returns the number of rows.
func (ds *BaseDataSource) RowCount() int { return len(ds.rows) }

// CurrentRowNo returns the 0-based current row index, or -1 if not positioned.
func (ds *BaseDataSource) CurrentRowNo() int { return ds.currentRow }

// GetValue returns the value of the named column in the current row.
func (ds *BaseDataSource) GetValue(column string) (any, error) {
	if !ds.initialized {
		return nil, ErrNotInitialized
	}
	if ds.currentRow < 0 || ds.currentRow >= len(ds.rows) {
		return nil, fmt.Errorf("data source %q: no current row (position %d)", ds.name, ds.currentRow)
	}
	row := ds.rows[ds.currentRow]
	val, ok := row[column]
	if !ok {
		return nil, nil // column not found → nil value (same as .NET behaviour)
	}
	return val, nil
}

// Close is a no-op for in-memory data sources.
func (ds *BaseDataSource) Close() error {
	ds.initialized = false
	ds.currentRow = -1
	return nil
}

// Prior moves the cursor one row backwards.
// Mirrors C# DataSourceBase.Prior() (DataSourceBase.cs:724): CurrentRowNo--.
// No lower-bound check — callers must ensure position validity.
func (ds *BaseDataSource) Prior() {
	ds.currentRow--
}

// SetCurrentRowNo directly positions the cursor to the given 0-based row index.
// Mirrors C# DataSourceBase.CurrentRowNo setter used in ReportEngine.Groups.cs line 226.
func (ds *BaseDataSource) SetCurrentRowNo(n int) {
	ds.currentRow = n
}

// EnsureInit initialises the data source if it has not been initialised yet.
// Mirrors C# DataSourceBase.EnsureInit() lazy-init pattern.
func (ds *BaseDataSource) EnsureInit() error {
	if ds.initialized {
		return nil
	}
	return ds.Init()
}

// GetDisplayName returns the human-readable display name.
// Returns Alias if it is non-empty, otherwise returns Name.
// Mirrors C# DataComponentBase.GetDisplayName() behaviour.
func (ds *BaseDataSource) GetDisplayName() string {
	if ds.alias != "" {
		return ds.alias
	}
	return ds.name
}

// SetAdditionalFilter adds or replaces a column-level filter predicate.
// When rows are filtered (ApplyAdditionalFilter), only rows where pred returns true
// for the named column's value are kept.
// Mirrors C# DataSourceBase.AdditionalFilter Hashtable (DataSourceBase.cs:249-251).
func (ds *BaseDataSource) SetAdditionalFilter(column string, pred AdditionalFilterPredicate) {
	if ds.additionalFilter == nil {
		ds.additionalFilter = make(map[string]AdditionalFilterPredicate)
	}
	ds.additionalFilter[column] = pred
}

// ClearAdditionalFilter removes all additional filter predicates.
// Mirrors C# DataSourceBase.ClearData() → additionalFilter.Clear() (DataSourceBase.cs:754).
func (ds *BaseDataSource) ClearAdditionalFilter() {
	ds.additionalFilter = nil
}

// ApplyAdditionalFilter removes any rows that fail the additional filter predicates.
// Mirrors C# DataSourceBase.ApplyAdditionalFilter() (DataSourceBase.cs:325-341).
// Call after Init() when additional column filters are needed.
func (ds *BaseDataSource) ApplyAdditionalFilter() {
	if len(ds.additionalFilter) == 0 {
		return
	}
	filtered := ds.rows[:0]
	for _, row := range ds.rows {
		keep := true
		for col, pred := range ds.additionalFilter {
			if !pred(row[col]) {
				keep = false
				break
			}
		}
		if keep {
			filtered = append(filtered, row)
		}
	}
	ds.rows = filtered
}

// SortSpec describes one sort key for SortRows.
type SortSpec struct {
	Column     string
	Descending bool
}

// SortRows reorders the internal rows slice according to specs.
// Only string, int64, float64, and bool column values are compared;
// other types fall back to fmt.Sprintf comparison.
func (ds *BaseDataSource) SortRows(specs []SortSpec) {
	if len(specs) == 0 {
		return
	}
	slices.SortStableFunc(ds.rows, func(a, b map[string]any) int {
		for _, spec := range specs {
			av := a[spec.Column]
			bv := b[spec.Column]
			c := compareAny(av, bv)
			if spec.Descending {
				c = -c
			}
			if c != 0 {
				return c
			}
		}
		return 0
	})
}

// Sortable is implemented by data sources that support in-memory row sorting.
type Sortable interface {
	SortRows(specs []SortSpec)
}

// unicodeCollator provides culture-aware string comparison that matches C#'s
// default CurrentCulture sort order. Accented characters sort near their base
// characters (e.g. â sorts as a variant of 'a'), matching System.String.Compare.
var unicodeCollator = collate.New(language.Und)

// compareAny compares two values of arbitrary type.
func compareAny(a, b any) int {
	switch av := a.(type) {
	case int:
		bv, _ := b.(int)
		return cmp.Compare(av, bv)
	case int64:
		bv, _ := b.(int64)
		return cmp.Compare(av, bv)
	case float64:
		bv, _ := b.(float64)
		return cmp.Compare(av, bv)
	case float32:
		bv, _ := b.(float32)
		return cmp.Compare(av, bv)
	case string:
		bv, _ := b.(string)
		// Use Unicode collation to match C# culture-sensitive string comparison.
		// This ensures accented characters (â, é, ö) sort near their base characters.
		return unicodeCollator.CompareString(av, bv)
	case bool:
		if av == b.(bool) {
			return 0
		}
		if !av {
			return -1
		}
		return 1
	default:
		return unicodeCollator.CompareString(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
	}
}

// Verify BaseDataSource satisfies DataSource at compile time.
var _ DataSource = (*BaseDataSource)(nil)
