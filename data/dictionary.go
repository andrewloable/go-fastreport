package data

import "strings"

// Dictionary is the central registry for all data components in a report:
// data sources, connections, relations, parameters, system variables, and totals.
// It is the Go equivalent of FastReport.Data.Dictionary.
type Dictionary struct {
	dataSources     []DataSource
	relations       []*Relation
	parameters      []*Parameter
	systemVariables []*Parameter
	totals          []*Total
}

// NewDictionary creates an empty Dictionary.
func NewDictionary() *Dictionary {
	return &Dictionary{}
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
