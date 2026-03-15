package data

import "strings"

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
