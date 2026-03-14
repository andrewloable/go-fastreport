// Package data provides data binding types for go-fastreport.
// It is the Go equivalent of FastReport.Data namespace.
package data

import "fmt"

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

// BaseDataSource provides a reusable implementation of the DataSource interface
// backed by an in-memory slice of row maps. Concrete data sources can embed it
// and override GetValue to supply their own row storage.
type BaseDataSource struct {
	name        string
	alias       string
	columns     []Column
	rows        []map[string]any
	currentRow  int
	initialized bool
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

// SetName sets the data source name.
func (ds *BaseDataSource) SetName(n string) { ds.name = n }

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

// Verify BaseDataSource satisfies DataSource at compile time.
var _ DataSource = (*BaseDataSource)(nil)
