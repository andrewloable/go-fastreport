package data

import "strings"

// Dictionary is the central registry for all data components in a report:
// data sources, connections, relations, parameters, system variables, and totals.
// It is the Go equivalent of FastReport.Data.Dictionary.
type Dictionary struct {
	connections     []*DataConnectionBase
	dataSources     []DataSource
	relations       []*Relation
	parameters      []*Parameter
	systemVariables []*Parameter
	totals          []*Total
	// aggregateTotals holds the richer aggregate-total definitions used by the engine.
	aggregateTotals []*AggregateTotal
}

// NewDictionary creates an empty Dictionary.
func NewDictionary() *Dictionary {
	return &Dictionary{}
}

// -----------------------------------------------------------------------
// Connections
// -----------------------------------------------------------------------

// AddConnection registers a database connection.
func (d *Dictionary) AddConnection(c *DataConnectionBase) {
	d.connections = append(d.connections, c)
}

// RemoveConnection removes a connection by reference.
func (d *Dictionary) RemoveConnection(c *DataConnectionBase) {
	for i, conn := range d.connections {
		if conn == c {
			d.connections = append(d.connections[:i], d.connections[i+1:]...)
			return
		}
	}
}

// Connections returns all registered database connections.
func (d *Dictionary) Connections() []*DataConnectionBase { return d.connections }

// FindConnectionByName returns the connection with the given name (case-insensitive),
// or nil if not found.
func (d *Dictionary) FindConnectionByName(name string) *DataConnectionBase {
	for _, c := range d.connections {
		if strings.EqualFold(c.Name(), name) {
			return c
		}
	}
	return nil
}

// -----------------------------------------------------------------------
// Data sources
// -----------------------------------------------------------------------

// AddDataSource registers a data source.
func (d *Dictionary) AddDataSource(ds DataSource) {
	d.dataSources = append(d.dataSources, ds)
}

// RemoveDataSource removes a data source by reference.
func (d *Dictionary) RemoveDataSource(ds DataSource) {
	for i, s := range d.dataSources {
		if s == ds {
			d.dataSources = append(d.dataSources[:i], d.dataSources[i+1:]...)
			return
		}
	}
}

// DataSources returns all registered data sources.
func (d *Dictionary) DataSources() []DataSource { return d.dataSources }

// FindDataSourceByAlias returns the data source with the given alias,
// or nil if not found.
func (d *Dictionary) FindDataSourceByAlias(alias string) DataSource {
	for _, ds := range d.dataSources {
		if strings.EqualFold(ds.Alias(), alias) {
			return ds
		}
	}
	return nil
}

// FindDataSourceByName returns the data source with the given name,
// or nil if not found.
func (d *Dictionary) FindDataSourceByName(name string) DataSource {
	for _, ds := range d.dataSources {
		if strings.EqualFold(ds.Name(), name) {
			return ds
		}
	}
	return nil
}

// RegisterData is the high-level method for binding Go values to the dictionary.
// It creates a BusinessObjectDataSource with the given name and registers it.
func (d *Dictionary) RegisterData(value any, name string) DataSource {
	ds := NewBusinessObjectDataSource(name, value)
	d.AddDataSource(ds)
	return ds
}

// -----------------------------------------------------------------------
// Relations
// -----------------------------------------------------------------------

// AddRelation registers a master-detail relation.
func (d *Dictionary) AddRelation(r *Relation) {
	d.relations = append(d.relations, r)
}

// RemoveRelation removes a relation by reference.
func (d *Dictionary) RemoveRelation(r *Relation) {
	for i, rel := range d.relations {
		if rel == r {
			d.relations = append(d.relations[:i], d.relations[i+1:]...)
			return
		}
	}
}

// Relations returns all registered relations.
func (d *Dictionary) Relations() []*Relation { return d.relations }

// -----------------------------------------------------------------------
// Parameters
// -----------------------------------------------------------------------

// AddParameter registers a report parameter.
func (d *Dictionary) AddParameter(p *Parameter) {
	d.parameters = append(d.parameters, p)
}

// RemoveParameter removes a parameter by reference.
func (d *Dictionary) RemoveParameter(p *Parameter) {
	for i, param := range d.parameters {
		if param == p {
			d.parameters = append(d.parameters[:i], d.parameters[i+1:]...)
			return
		}
	}
}

// Parameters returns all report parameters.
func (d *Dictionary) Parameters() []*Parameter { return d.parameters }

// FindParameter returns the parameter with the given name (dot-separated for nested),
// or nil if not found.
func (d *Dictionary) FindParameter(name string) *Parameter {
	return GetParameter(d, name)
}

// -----------------------------------------------------------------------
// System variables
// -----------------------------------------------------------------------

// AddSystemVariable registers a system variable.
func (d *Dictionary) AddSystemVariable(p *Parameter) {
	d.systemVariables = append(d.systemVariables, p)
}

// SystemVariables returns all system variables.
func (d *Dictionary) SystemVariables() []*Parameter { return d.systemVariables }

// SetSystemVariable sets the value of an existing system variable by name,
// or creates it if it does not exist.
func (d *Dictionary) SetSystemVariable(name string, value any) {
	for _, sv := range d.systemVariables {
		if strings.EqualFold(sv.Name, name) {
			sv.Value = value
			return
		}
	}
	d.systemVariables = append(d.systemVariables, &Parameter{Name: name, Value: value})
}

// -----------------------------------------------------------------------
// Totals
// -----------------------------------------------------------------------

// AddTotal registers an aggregate total.
func (d *Dictionary) AddTotal(t *Total) {
	d.totals = append(d.totals, t)
}

// RemoveTotal removes a total by reference.
func (d *Dictionary) RemoveTotal(t *Total) {
	for i, tot := range d.totals {
		if tot == t {
			d.totals = append(d.totals[:i], d.totals[i+1:]...)
			return
		}
	}
}

// Totals returns all registered totals.
func (d *Dictionary) Totals() []*Total { return d.totals }

// AggregateTotals returns all aggregate total definitions.
func (d *Dictionary) AggregateTotals() []*AggregateTotal { return d.aggregateTotals }

// AddAggregateTotal registers an aggregate total definition.
// It also ensures a corresponding simple Total entry exists in Totals() so that
// the expression evaluator can reference the current accumulated value by name.
func (d *Dictionary) AddAggregateTotal(at *AggregateTotal) {
	d.aggregateTotals = append(d.aggregateTotals, at)
	// Ensure a matching simple Total placeholder exists.
	for _, t := range d.totals {
		if strings.EqualFold(t.Name, at.Name) {
			return
		}
	}
	d.totals = append(d.totals, &Total{Name: at.Name})
}

// FindTotal returns the total with the given name, or nil.
func (d *Dictionary) FindTotal(name string) *Total {
	for _, t := range d.totals {
		if strings.EqualFold(t.Name, name) {
			return t
		}
	}
	return nil
}

// -----------------------------------------------------------------------
// DictionaryLookup implementation
// -----------------------------------------------------------------------

// Dictionary implements the DictionaryLookup interface defined in helper.go.
var _ DictionaryLookup = (*Dictionary)(nil)

// Merge merges data sources, connections, parameters, and totals from source
// into this dictionary, skipping any entry whose name already exists.
// Mirrors C# Dictionary.Merge(source) (Dictionary.cs:725-780) — the Go port
// uses a simple name-deduplication approach instead of the C# clone-and-fixup.
func (d *Dictionary) Merge(source *Dictionary) {
	for _, c := range source.connections {
		if d.FindConnectionByName(c.Name()) == nil {
			d.AddConnection(c)
		}
	}
	for _, ds := range source.dataSources {
		if d.FindDataSourceByName(ds.Name()) == nil {
			d.AddDataSource(ds)
		}
	}
	for _, rel := range source.relations {
		found := false
		for _, r := range d.relations {
			if strings.EqualFold(r.Name, rel.Name) {
				found = true
				break
			}
		}
		if !found {
			d.relations = append(d.relations, rel)
		}
	}
	for _, p := range source.parameters {
		exists := false
		for _, ep := range d.parameters {
			if strings.EqualFold(ep.Name, p.Name) {
				exists = true
				break
			}
		}
		if !exists {
			d.parameters = append(d.parameters, p)
		}
	}
	for _, t := range source.totals {
		if d.FindTotal(t.Name) == nil {
			d.totals = append(d.totals, t)
		}
	}
}

// ResolveRelations resolves ParentDataSource and ChildDataSource for each
// relation from their string names, and also splits ParentColumnNames /
// ChildColumnNames into the ParentColumns / ChildColumns slices.
// Call this after all data sources have been registered (typically at Prepare time).
func (d *Dictionary) ResolveRelations() {
	for _, rel := range d.relations {
		if rel.ParentDataSource == nil && rel.ParentSourceName != "" {
			rel.ParentDataSource = d.FindDataSourceByAlias(rel.ParentSourceName)
			if rel.ParentDataSource == nil {
				rel.ParentDataSource = d.FindDataSourceByName(rel.ParentSourceName)
			}
		}
		if rel.ChildDataSource == nil && rel.ChildSourceName != "" {
			rel.ChildDataSource = d.FindDataSourceByAlias(rel.ChildSourceName)
			if rel.ChildDataSource == nil {
				rel.ChildDataSource = d.FindDataSourceByName(rel.ChildSourceName)
			}
		}
		// Resolve column name slices if not already set.
		if len(rel.ParentColumns) == 0 {
			rel.ParentColumns = rel.ParentColumnNames
		}
		if len(rel.ChildColumns) == 0 {
			rel.ChildColumns = rel.ChildColumnNames
		}
	}
}
