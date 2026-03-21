package data

import (
	"fmt"
	"sort"
	"strings"
)

// DataConnectionCollection provides ordered, named access to a set of
// DataConnectionBase objects. It is the Go equivalent of
// FastReport.Data.DataConnectionCollection.
type DataConnectionCollection struct {
	items []*DataConnectionBase
}

// NewDataConnectionCollection creates an empty collection.
func NewDataConnectionCollection() *DataConnectionCollection {
	return &DataConnectionCollection{}
}

// Add appends a connection to the collection.
func (c *DataConnectionCollection) Add(conn *DataConnectionBase) {
	c.items = append(c.items, conn)
}

// Remove removes a connection by reference.
func (c *DataConnectionCollection) Remove(conn *DataConnectionBase) {
	for i, v := range c.items {
		if v == conn {
			c.items = append(c.items[:i], c.items[i+1:]...)
			return
		}
	}
}

// Count returns the number of connections.
func (c *DataConnectionCollection) Count() int { return len(c.items) }

// Get returns the connection at index i.
func (c *DataConnectionCollection) Get(i int) *DataConnectionBase { return c.items[i] }

// All returns a copy of the internal slice.
func (c *DataConnectionCollection) All() []*DataConnectionBase {
	out := make([]*DataConnectionBase, len(c.items))
	copy(out, c.items)
	return out
}

// FindByName returns the connection with the given name (case-insensitive),
// or nil if not found.
func (c *DataConnectionCollection) FindByName(name string) *DataConnectionBase {
	for _, v := range c.items {
		if strings.EqualFold(v.Name(), name) {
			return v
		}
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────

// DataSourceCollection provides ordered, named access to DataSource objects.
// It is the Go equivalent of FastReport.Data.DataSourceCollection.
type DataSourceCollection struct {
	items []DataSource
}

// NewDataSourceCollection creates an empty collection.
func NewDataSourceCollection() *DataSourceCollection {
	return &DataSourceCollection{}
}

// Add appends a data source to the collection.
func (c *DataSourceCollection) Add(ds DataSource) {
	c.items = append(c.items, ds)
}

// Remove removes a data source by reference.
func (c *DataSourceCollection) Remove(ds DataSource) {
	for i, v := range c.items {
		if v == ds {
			c.items = append(c.items[:i], c.items[i+1:]...)
			return
		}
	}
}

// Count returns the number of data sources.
func (c *DataSourceCollection) Count() int { return len(c.items) }

// Get returns the data source at index i.
func (c *DataSourceCollection) Get(i int) DataSource { return c.items[i] }

// All returns a copy of the internal slice.
func (c *DataSourceCollection) All() []DataSource {
	out := make([]DataSource, len(c.items))
	copy(out, c.items)
	return out
}

// FindByName returns the data source with the given name (case-insensitive),
// or nil if not found.
func (c *DataSourceCollection) FindByName(name string) DataSource {
	for _, v := range c.items {
		if strings.EqualFold(v.Name(), name) {
			return v
		}
	}
	return nil
}

// FindByAlias returns the data source with the given alias (case-insensitive),
// or nil if not found.
func (c *DataSourceCollection) FindByAlias(alias string) DataSource {
	for _, v := range c.items {
		if strings.EqualFold(v.Alias(), alias) {
			return v
		}
	}
	return nil
}

// Sort sorts data sources by their aliases (ascending).
// C# ref: FastReport.Data.DataSourceCollection.Sort()
func (c *DataSourceCollection) Sort() {
	sort.SliceStable(c.items, func(i, j int) bool {
		return c.items[i].Alias() < c.items[j].Alias()
	})
}

// ──────────────────────────────────────────────────────────────────────────────

// TotalCollection manages a set of Total objects used during report execution.
// It is the Go equivalent of FastReport.Data.TotalCollection.
type TotalCollection struct {
	items []*Total
}

// NewTotalCollection creates an empty collection.
func NewTotalCollection() *TotalCollection {
	return &TotalCollection{}
}

// Add appends a total to the collection.
func (c *TotalCollection) Add(t *Total) {
	c.items = append(c.items, t)
}

// Remove removes a total by reference.
func (c *TotalCollection) Remove(t *Total) {
	for i, v := range c.items {
		if v == t {
			c.items = append(c.items[:i], c.items[i+1:]...)
			return
		}
	}
}

// Count returns the number of totals.
func (c *TotalCollection) Count() int { return len(c.items) }

// Get returns the total at index i.
func (c *TotalCollection) Get(i int) *Total { return c.items[i] }

// All returns a copy of the internal slice.
func (c *TotalCollection) All() []*Total {
	out := make([]*Total, len(c.items))
	copy(out, c.items)
	return out
}

// FindByName returns the first total whose name matches (case-insensitive),
// or nil if not found.
// C# ref: FastReport.Data.TotalCollection.FindByName
func (c *TotalCollection) FindByName(name string) *Total {
	for _, v := range c.items {
		if v.Name == name || strings.EqualFold(v.Name, name) {
			return v
		}
	}
	return nil
}

// CreateUniqueName returns a unique total name derived from name by appending
// an incrementing integer suffix until no collision is found.
// C# ref: FastReport.Data.TotalCollection.CreateUniqueName
func (c *TotalCollection) CreateUniqueName(name string) string {
	base := name
	i := 1
	for c.FindByName(name) != nil {
		name = fmt.Sprintf("%s%d", base, i)
		i++
	}
	return name
}

// Contains reports whether the given Total pointer is already in the collection.
// C# ref: FastReport.Data.TotalCollection.Contains (via FRCollectionBase)
func (c *TotalCollection) Contains(t *Total) bool {
	for _, v := range c.items {
		if v == t {
			return true
		}
	}
	return false
}

// GetValue returns the current value of the named total.
// Returns an error when the total is not found.
// C# ref: FastReport.Data.TotalCollection.GetValue (internal)
func (c *TotalCollection) GetValue(name string) (any, error) {
	t := c.FindByName(name)
	if t == nil {
		return nil, fmt.Errorf("TotalCollection: total %q not found", name)
	}
	return t.Value, nil
}

// ClearValues resets the accumulated Value on all totals in the collection.
// C# ref: FastReport.Data.TotalCollection.ClearValues (internal)
func (c *TotalCollection) ClearValues() {
	for _, t := range c.items {
		t.Value = nil
	}
}

// ──────────────────────────────────────────────────────────────────────────────

// TableCollection provides ordered access to TableDataSource objects.
// It is the Go equivalent of FastReport.Data.TableCollection.
type TableCollection struct {
	items []*TableDataSource
}

// NewTableCollection creates an empty collection.
func NewTableCollection() *TableCollection {
	return &TableCollection{}
}

// Add appends a table data source to the collection.
func (c *TableCollection) Add(t *TableDataSource) {
	c.items = append(c.items, t)
}

// Remove removes a table data source by reference.
func (c *TableCollection) Remove(t *TableDataSource) {
	for i, v := range c.items {
		if v == t {
			c.items = append(c.items[:i], c.items[i+1:]...)
			return
		}
	}
}

// Count returns the number of table data sources.
func (c *TableCollection) Count() int { return len(c.items) }

// Get returns the table data source at index i.
func (c *TableCollection) Get(i int) *TableDataSource { return c.items[i] }

// All returns a copy of the internal slice.
func (c *TableCollection) All() []*TableDataSource {
	out := make([]*TableDataSource, len(c.items))
	copy(out, c.items)
	return out
}

// Sort sorts table data sources by their aliases (ascending).
// C# ref: FastReport.Data.TableCollection.Sort()
func (c *TableCollection) Sort() {
	sort.SliceStable(c.items, func(i, j int) bool {
		return c.items[i].Alias() < c.items[j].Alias()
	})
}
