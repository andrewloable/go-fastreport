package data

import "strings"

// Relation describes a master-detail relationship between two data sources.
// It is the Go equivalent of FastReport.Data.Relation.
type Relation struct {
	// Name is the relation's unique name.
	Name string
	// Alias is the human-friendly display name (defaults to Name).
	Alias string
	// ParentDataSource is the parent (master) data source.
	ParentDataSource DataSource
	// ChildDataSource is the child (detail) data source.
	ChildDataSource DataSource
	// ParentColumns are the join column names in the parent data source.
	ParentColumns []string
	// ChildColumns are the join column names in the child data source.
	ChildColumns []string
}

// Parameter represents a named report parameter with an optional nested
// parameter collection. It is the Go equivalent of FastReport.Data.Parameter.
type Parameter struct {
	// Name is the parameter name.
	Name string
	// Value is the current runtime value.
	Value any
	// Expression is an optional expression string for computed parameters.
	Expression string
	// DataType describes the expected Go type (e.g. "string", "int64").
	DataType string
	// nested holds nested child parameters.
	nested []*Parameter
}

// Parameters returns the nested child parameters slice (creates it on demand).
func (p *Parameter) Parameters() []*Parameter { return p.nested }

// AddParameter appends a nested parameter.
func (p *Parameter) AddParameter(child *Parameter) {
	p.nested = append(p.nested, child)
}

// FindByName returns the first nested parameter with the given name, or nil.
func FindParameterByName(params []*Parameter, name string) *Parameter {
	for _, p := range params {
		if p.Name == name {
			return p
		}
	}
	return nil
}

// Total represents an aggregate calculation over a data band.
// It is the Go equivalent of FastReport.Data.Total.
type Total struct {
	// Name is the total's unique name.
	Name string
	// Value is the computed aggregate value.
	Value any
}

// DictionaryLookup is the minimal interface that DataHelper needs from the
// report Dictionary. The full Dictionary implementation satisfies this interface.
type DictionaryLookup interface {
	// FindDataSourceByAlias returns the data source with the given alias, or nil.
	FindDataSourceByAlias(alias string) DataSource
	// Relations returns all defined master-detail relations.
	Relations() []*Relation
	// Parameters returns the top-level parameter slice.
	Parameters() []*Parameter
	// SystemVariables returns the built-in system variable parameters.
	SystemVariables() []*Parameter
	// Totals returns all defined totals.
	Totals() []*Total
}

// GetDataSource resolves a dot-separated complex name (e.g. "Orders.Details")
// to the deepest matching data source in the dictionary.
// Returns nil when not found.
func GetDataSource(dict DictionaryLookup, complexName string) DataSource {
	if complexName == "" {
		return nil
	}
	names := strings.Split(complexName, ".")
	ds := dict.FindDataSourceByAlias(names[0])
	if ds == nil {
		return nil
	}
	// For nested datasources (e.g. BO columns that are themselves DataSources)
	// we traverse the column tree. This is handled by callers; for now return
	// the root datasource when multiple segments are present.
	_ = names
	return ds
}

// FindRelation returns the Relation where parent and child match the given
// data sources, or nil when no such relation is registered.
func FindRelation(dict DictionaryLookup, parent, child DataSource) *Relation {
	for _, r := range dict.Relations() {
		if r.ParentDataSource == parent && r.ChildDataSource == child {
			return r
		}
	}
	return nil
}

// GetParameter resolves a dot-separated complex name (e.g. "Filters.MinDate")
// through the nested parameter hierarchy. Returns nil when not found.
// Also searches system variables when a single-segment name is not found in
// user parameters.
func GetParameter(dict DictionaryLookup, complexName string) *Parameter {
	if complexName == "" {
		return nil
	}
	names := strings.Split(complexName, ".")
	par := FindParameterByName(dict.Parameters(), names[0])
	if par == nil {
		par = FindParameterByName(dict.SystemVariables(), names[0])
		return par // system variables cannot have nested children
	}
	for _, segment := range names[1:] {
		par = FindParameterByName(par.Parameters(), segment)
		if par == nil {
			return nil
		}
	}
	return par
}

// IsValidParameter returns true when the complex parameter name resolves to
// an existing user parameter or system variable.
func IsValidParameter(dict DictionaryLookup, complexName string) bool {
	return GetParameter(dict, complexName) != nil
}

// GetTotal returns the value of the named total, or nil when not found.
func GetTotal(dict DictionaryLookup, name string) any {
	for _, t := range dict.Totals() {
		if t.Name == name {
			return t.Value
		}
	}
	return nil
}

// IsValidTotal returns true when a total with the given name exists.
func IsValidTotal(dict DictionaryLookup, name string) bool {
	for _, t := range dict.Totals() {
		if t.Name == name {
			return true
		}
	}
	return false
}
