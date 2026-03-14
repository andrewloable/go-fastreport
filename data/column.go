package data

import "iter"

// ColumnFormat specifies how a column's value is formatted.
type ColumnFormat int

const (
	// ColumnFormatAuto determines format automatically from data type.
	ColumnFormatAuto ColumnFormat = iota
	// ColumnFormatGeneral applies no formatting.
	ColumnFormatGeneral
	// ColumnFormatNumber applies number formatting.
	ColumnFormatNumber
	// ColumnFormatCurrency applies currency formatting.
	ColumnFormatCurrency
	// ColumnFormatDate applies date formatting.
	ColumnFormatDate
	// ColumnFormatTime applies time formatting.
	ColumnFormatTime
	// ColumnFormatPercent applies percent formatting.
	ColumnFormatPercent
	// ColumnFormatBoolean applies boolean formatting.
	ColumnFormatBoolean
)

// DataColumn represents a single data column in a data source.
// It is the Go equivalent of FastReport.Data.Column.
type DataColumn struct {
	// Name is the column's programmatic name.
	Name string
	// Alias is the human-friendly display name.
	Alias string
	// DataType is a string describing the column's value type (e.g. "string", "int64").
	DataType string
	// Format specifies the display format.
	Format ColumnFormat
	// Calculated is true when the column value comes from an expression.
	Calculated bool
	// Expression is the formula used for calculated columns.
	Expression string
	// Enabled controls whether the column is active.
	Enabled bool
	// PropName is the bound business-object property name.
	PropName string
	// columns holds nested child columns (for hierarchical data).
	columns *ColumnCollection
}

// NewDataColumn creates a DataColumn with the given name and enabled=true.
func NewDataColumn(name string) *DataColumn {
	return &DataColumn{
		Name:    name,
		Alias:   name,
		Enabled: true,
	}
}

// Columns returns the nested column collection, creating it on first access.
func (c *DataColumn) Columns() *ColumnCollection {
	if c.columns == nil {
		c.columns = NewColumnCollection()
	}
	return c.columns
}

// HasColumns returns true if this column has nested child columns.
func (c *DataColumn) HasColumns() bool {
	return c.columns != nil && c.columns.Len() > 0
}

// ColumnCollection is an ordered collection of DataColumns.
type ColumnCollection struct {
	items []*DataColumn
}

// NewColumnCollection creates an empty ColumnCollection.
func NewColumnCollection() *ColumnCollection {
	return &ColumnCollection{}
}

// Add appends col to the collection.
func (cc *ColumnCollection) Add(col *DataColumn) {
	cc.items = append(cc.items, col)
}

// Len returns the number of columns.
func (cc *ColumnCollection) Len() int { return len(cc.items) }

// Get returns the column at index i.
func (cc *ColumnCollection) Get(i int) *DataColumn { return cc.items[i] }

// Remove removes the first column with the given name.
// Returns true if found and removed.
func (cc *ColumnCollection) Remove(name string) bool {
	for i, col := range cc.items {
		if col.Name == name {
			cc.items = append(cc.items[:i], cc.items[i+1:]...)
			return true
		}
	}
	return false
}

// Clear removes all columns.
func (cc *ColumnCollection) Clear() {
	cc.items = cc.items[:0]
}

// FindByName returns the first column whose Name matches, or nil.
func (cc *ColumnCollection) FindByName(name string) *DataColumn {
	for _, col := range cc.items {
		if col.Name == name {
			return col
		}
		// Recurse into nested columns
		if col.HasColumns() {
			if found := col.Columns().FindByName(name); found != nil {
				return found
			}
		}
	}
	return nil
}

// FindByAlias returns the first column whose Alias matches, or nil.
func (cc *ColumnCollection) FindByAlias(alias string) *DataColumn {
	for _, col := range cc.items {
		if col.Alias == alias {
			return col
		}
		if col.HasColumns() {
			if found := col.Columns().FindByAlias(alias); found != nil {
				return found
			}
		}
	}
	return nil
}

// All returns an iterator over all columns (Go 1.23 range-over-func).
func (cc *ColumnCollection) All() iter.Seq2[int, *DataColumn] {
	return func(yield func(int, *DataColumn) bool) {
		for i, col := range cc.items {
			if !yield(i, col) {
				return
			}
		}
	}
}

// Slice returns a copy of the underlying slice.
func (cc *ColumnCollection) Slice() []*DataColumn {
	result := make([]*DataColumn, len(cc.items))
	copy(result, cc.items)
	return result
}
