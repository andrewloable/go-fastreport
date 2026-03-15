package data

// VirtualDataSource is a data source that produces a fixed number of rows
// with no actual column data. It is the Go equivalent of
// FastReport.Data.VirtualDataSource.
//
// Use it when a band needs to iterate N times and values are computed by
// report expressions (not read from a real data source). Each row is an
// empty position — all GetValue calls return nil.
//
// Example:
//
//	ds := data.NewVirtualDataSource("MyVirtual", 10)
//	dict.RegisterData(ds, "MyVirtual")
type VirtualDataSource struct {
	name       string
	alias      string
	rowCount   int
	currentRow int
	initialized bool
}

// NewVirtualDataSource creates a VirtualDataSource with the given name that
// will produce rowCount virtual rows.
func NewVirtualDataSource(name string, rowCount int) *VirtualDataSource {
	return &VirtualDataSource{
		name:     name,
		alias:    name,
		rowCount: rowCount,
	}
}

// RowsCount returns the number of virtual rows.
func (ds *VirtualDataSource) RowsCount() int { return ds.rowCount }

// SetRowsCount sets the number of virtual rows.
func (ds *VirtualDataSource) SetRowsCount(n int) { ds.rowCount = n }

// Name returns the data source name.
func (ds *VirtualDataSource) Name() string { return ds.name }

// Alias returns the human-friendly alias.
func (ds *VirtualDataSource) Alias() string { return ds.alias }

// SetAlias sets the display alias.
func (ds *VirtualDataSource) SetAlias(a string) { ds.alias = a }

// Init marks the data source as initialized and resets position.
func (ds *VirtualDataSource) Init() error {
	ds.initialized = true
	ds.currentRow = -1
	return nil
}

// First positions at the first virtual row.
func (ds *VirtualDataSource) First() error {
	if !ds.initialized {
		return ErrNotInitialized
	}
	if ds.rowCount <= 0 {
		ds.currentRow = ds.rowCount
		return ErrEOF
	}
	ds.currentRow = 0
	return nil
}

// Next advances to the next virtual row.
func (ds *VirtualDataSource) Next() error {
	if !ds.initialized {
		return ErrNotInitialized
	}
	ds.currentRow++
	if ds.currentRow >= ds.rowCount {
		return ErrEOF
	}
	return nil
}

// EOF returns true when all virtual rows have been consumed.
func (ds *VirtualDataSource) EOF() bool {
	return ds.currentRow >= ds.rowCount
}

// RowCount returns the total number of virtual rows.
func (ds *VirtualDataSource) RowCount() int { return ds.rowCount }

// CurrentRowNo returns the 0-based index of the current virtual row.
func (ds *VirtualDataSource) CurrentRowNo() int { return ds.currentRow }

// GetValue always returns nil — virtual rows carry no real column data.
// Values are expected to be computed by report expressions.
func (ds *VirtualDataSource) GetValue(column string) (any, error) {
	return nil, nil
}

// Close is a no-op for a virtual data source.
func (ds *VirtualDataSource) Close() error { return nil }
