package data

import (
	"strings"

	"github.com/andrewloable/go-fastreport/report"
)

// Relation describes a master-detail relationship between two data sources.
// It is the Go equivalent of FastReport.Data.Relation.
type Relation struct {
	// Name is the relation's unique name.
	Name string
	// Alias is the human-friendly display name (defaults to Name).
	Alias string
	// Enabled controls whether this relation is active.
	Enabled bool
	// ParentDataSource is the parent (master) data source (resolved at prepare time).
	ParentDataSource DataSource
	// ChildDataSource is the child (detail) data source (resolved at prepare time).
	ChildDataSource DataSource
	// ParentSourceName is the FRX-level data source name for the parent.
	// The engine resolves this to ParentDataSource via the dictionary.
	ParentSourceName string
	// ChildSourceName is the FRX-level data source name for the child.
	ChildSourceName string
	// ParentColumns are the join column names in the parent data source.
	// Populated from ParentColumnNames after dictionary resolution.
	ParentColumns []string
	// ChildColumns are the join column names in the child data source.
	// Populated from ChildColumnNames after dictionary resolution.
	ChildColumns []string
	// ParentColumnNames holds the raw comma-split column names from FRX.
	ParentColumnNames []string
	// ChildColumnNames holds the raw comma-split column names from FRX.
	ChildColumnNames []string
}

// NewRelation creates a Relation with Enabled=true.
func NewRelation() *Relation {
	return &Relation{Enabled: true}
}

// Serialize writes the relation's properties to w.
func (r *Relation) Serialize(w report.Writer) error {
	parentName := ""
	if r.ParentDataSource != nil {
		parentName = r.ParentDataSource.Name()
	} else {
		parentName = r.ParentSourceName
	}
	childName := ""
	if r.ChildDataSource != nil {
		childName = r.ChildDataSource.Name()
	} else {
		childName = r.ChildSourceName
	}
	if parentName != "" {
		w.WriteStr("ParentDataSource", parentName)
	}
	if childName != "" {
		w.WriteStr("ChildDataSource", childName)
	}
	if len(r.ParentColumns) > 0 {
		w.WriteStr("ParentColumns", strings.Join(r.ParentColumns, ","))
	}
	if len(r.ChildColumns) > 0 {
		w.WriteStr("ChildColumns", strings.Join(r.ChildColumns, ","))
	}
	w.WriteBool("Enabled", r.Enabled)
	return nil
}

// Deserialize reads the relation's properties from rdr.
func (r *Relation) Deserialize(rdr report.Reader) error {
	r.ParentSourceName = rdr.ReadStr("ParentDataSource", "")
	r.ChildSourceName = rdr.ReadStr("ChildDataSource", "")
	if cols := rdr.ReadStr("ParentColumns", ""); cols != "" {
		r.ParentColumnNames = strings.Split(cols, ",")
	}
	if cols := rdr.ReadStr("ChildColumns", ""); cols != "" {
		r.ChildColumnNames = strings.Split(cols, ",")
	}
	r.Enabled = rdr.ReadBool("Enabled", true)
	return nil
}

// Equals returns true when both relations have the same parent/child data
// sources and the same column sets.
func (r *Relation) Equals(other *Relation) bool {
	return r.ParentDataSource == other.ParentDataSource &&
		r.ChildDataSource == other.ChildDataSource &&
		strings.Join(r.ParentColumns, ",") == strings.Join(other.ParentColumns, ",") &&
		strings.Join(r.ChildColumns, ",") == strings.Join(other.ChildColumns, ",")
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
	// Description is a human-readable description of this parameter.
	Description string
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

// enumParameters recursively collects all parameters from root into list,
// keyed by their dot-separated full name. Mirrors C# ParameterCollection.EnumParameters.
func enumParameters(root []*Parameter, prefix string, list map[string]*Parameter) {
	for _, p := range root {
		fullName := p.Name
		if prefix != "" {
			fullName = prefix + "." + p.Name
		}
		if _, exists := list[fullName]; !exists {
			list[fullName] = p
			enumParameters(p.nested, fullName, list)
		}
	}
}

// AssignValues copies parameter values from src parameters into dst parameters
// by matching full dot-separated names. Mirrors C# ParameterCollection.AssignValues.
func AssignValues(dst []*Parameter, src []*Parameter) {
	dstMap := make(map[string]*Parameter)
	enumParameters(dst, "", dstMap)
	srcMap := make(map[string]*Parameter)
	enumParameters(src, "", srcMap)
	for fullName, srcParam := range srcMap {
		if dstParam, ok := dstMap[fullName]; ok {
			dstParam.Value = srcParam.Value
		}
	}
}

// Total represents an aggregate calculation over a data band.
// It is the Go equivalent of FastReport.Data.Total.
type Total struct {
	// Name is the total's unique name.
	Name string
	// Value is the computed aggregate value (set at report run time).
	Value any
	// Expression is the value expression evaluated per row (empty for Count).
	Expression string
	// TotalType is the aggregate function (Sum, Min, Max, Avg, Count, etc.).
	TotalType TotalType
	// Evaluator is the name of the DataBand that drives this total.
	Evaluator string
	// PrintOn is the name of the band where the total is printed/reset.
	PrintOn string
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

type dataSourceByNameLookup interface {
	FindDataSourceByName(name string) DataSource
}

type dataSourceColumns interface {
	Columns() []Column
}

func findDataSource(dict DictionaryLookup, name string) DataSource {
	if dict == nil || name == "" {
		return nil
	}
	if ds := dict.FindDataSourceByAlias(name); ds != nil {
		return ds
	}
	if named, ok := dict.(dataSourceByNameLookup); ok {
		return named.FindDataSourceByName(name)
	}
	return nil
}

func findColumn(columns []Column, name string) *Column {
	for i := range columns {
		if strings.EqualFold(columns[i].Alias, name) || strings.EqualFold(columns[i].Name, name) {
			return &columns[i]
		}
	}
	return nil
}

// GetDataSource resolves a dot-separated complex name (e.g. "Orders.Details")
// to the deepest matching data source in the dictionary.
// Returns nil when not found.
func GetDataSource(dict DictionaryLookup, complexName string) DataSource {
	if complexName == "" {
		return nil
	}
	names := strings.Split(complexName, ".")
	ds := findDataSource(dict, names[0])
	if ds == nil {
		return nil
	}
	for _, segment := range names[1:] {
		next := findDataSource(dict, segment)
		if next == nil {
			break
		}
		ds = next
	}
	return ds
}

// unwrapDS unwraps any FilteredDataSource layers to return the underlying
// base DataSource. This is needed for pointer-based relation matching because
// the engine wraps child data sources in FilteredDataSource during master-detail
// rendering, and those wrappers should still match relations registered against
// the original data source pointer.
func unwrapDS(ds DataSource) DataSource {
	type hasInner interface{ Inner() DataSource }
	for {
		if u, ok := ds.(hasInner); ok {
			ds = u.Inner()
		} else {
			return ds
		}
	}
}

// FindRelation returns the Relation where parent and child match the given
// data sources, or nil when no such relation is registered.
// It first compares resolved DataSource pointers; if those are nil it falls
// back to matching by data source name.
// FilteredDataSource wrappers are unwrapped before pointer comparison so that
// nested master-detail bands (whose parent DS is already wrapped) still find
// the correct relation for their own children.
func FindRelation(dict DictionaryLookup, parent, child DataSource) *Relation {
	// Unwrap any FilteredDataSource layers for pointer comparison.
	parentBase := unwrapDS(parent)
	childBase := unwrapDS(child)

	for _, r := range dict.Relations() {
		if r.ParentDataSource != nil && r.ChildDataSource != nil {
			if r.ParentDataSource == parentBase && r.ChildDataSource == childBase {
				return r
			}
			continue
		}
		// Fallback: match by name when pointers have not been resolved yet.
		parentName := ""
		childName := ""
		if parent != nil {
			parentName = parent.Name()
		}
		if child != nil {
			childName = child.Name()
		}
		if strings.EqualFold(r.ParentSourceName, parentName) &&
			strings.EqualFold(r.ChildSourceName, childName) {
			return r
		}
	}
	return nil
}

// GetColumn resolves a dot-separated column reference such as
// "Orders.CustomerID" or relation-based paths like "Orders.Customers.Name".
// It mirrors the C# helper at a flat-column level using the Go datasource model.
func GetColumn(dict DictionaryLookup, complexName string) *DataColumn {
	if complexName == "" {
		return nil
	}
	names := strings.Split(complexName, ".")
	data := findDataSource(dict, names[0])
	return GetColumnFromParts(dict, data, names, false)
}

// GetColumnFromParts resolves a column reference starting from an already
// resolved datasource. When initRelation is true, relation traversal is allowed
// but row initialization side effects from the C# implementation are not needed
// in the current Go engine, so it behaves the same as false.
func GetColumnFromParts(dict DictionaryLookup, data DataSource, names []string, initRelation bool) *DataColumn {
	_ = initRelation
	if data == nil || len(names) < 2 {
		return nil
	}

	i := 1
	for ; i < len(names); i++ {
		found := false
		for _, r := range dict.Relations() {
			parentAlias := ""
			parentName := ""
			if r.ParentDataSource != nil {
				parentAlias = r.ParentDataSource.Alias()
				parentName = r.ParentDataSource.Name()
			}
			if r.ChildDataSource == unwrapDS(data) &&
				(strings.EqualFold(parentAlias, names[i]) || strings.EqualFold(parentName, names[i]) || strings.EqualFold(r.Alias, names[i])) {
				data = r.ParentDataSource
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	withCols, ok := data.(dataSourceColumns)
	if !ok {
		return nil
	}
	cols := withCols.Columns()

	fullName := strings.Join(names[i:], ".")
	if col := findColumn(cols, fullName); col != nil {
		return &DataColumn{Name: col.Name, Alias: col.Alias, DataType: col.DataType, Enabled: true, PropName: col.Name}
	}
	if col := findColumn(cols, names[len(names)-1]); col != nil {
		return &DataColumn{Name: col.Name, Alias: col.Alias, DataType: col.DataType, Enabled: true, PropName: col.Name}
	}
	return nil
}

// IsValidColumn returns true when the complex column name resolves successfully.
func IsValidColumn(dict DictionaryLookup, complexName string) bool {
	return GetColumn(dict, complexName) != nil
}

// IsSimpleColumn reports whether the reference is a direct datasource column
// without relation traversal.
func IsSimpleColumn(dict DictionaryLookup, complexName string) bool {
	if complexName == "" {
		return false
	}
	names := strings.Split(complexName, ".")
	if len(names) != 2 {
		return false
	}
	ds := findDataSource(dict, names[0])
	if ds == nil {
		return false
	}
	withCols, ok := ds.(dataSourceColumns)
	if !ok {
		return false
	}
	col := findColumn(withCols.Columns(), names[1])
	return col != nil
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

// CreateParameter resolves or creates a nested parameter chain.
func CreateParameter(dict *Dictionary, complexName string) *Parameter {
	if dict == nil || complexName == "" {
		return nil
	}
	names := strings.Split(complexName, ".")
	parameters := &dict.parameters
	var par *Parameter
	for _, name := range names {
		par = FindParameterByName(*parameters, name)
		if par == nil {
			par = &Parameter{Name: name}
			*parameters = append(*parameters, par)
		}
		parameters = &par.nested
	}
	return par
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

// GetColumnType returns the datatype string for a resolved column, or empty.
func GetColumnType(dict DictionaryLookup, complexName string) string {
	col := GetColumn(dict, complexName)
	if col == nil {
		return ""
	}
	return col.DataType
}

// RelationCollection is an ordered collection of Relation objects.
// It is the Go equivalent of FastReport.Data.RelationCollection.
type RelationCollection struct {
	items []*Relation
}

// NewRelationCollection creates an empty RelationCollection.
func NewRelationCollection() *RelationCollection {
	return &RelationCollection{}
}

// Add appends a relation to the collection.
func (c *RelationCollection) Add(rel *Relation) {
	c.items = append(c.items, rel)
}

// Remove removes a relation by reference.
func (c *RelationCollection) Remove(rel *Relation) {
	for i, v := range c.items {
		if v == rel {
			c.items = append(c.items[:i], c.items[i+1:]...)
			return
		}
	}
}

// Count returns the number of relations.
func (c *RelationCollection) Count() int { return len(c.items) }

// Get returns the relation at index i.
func (c *RelationCollection) Get(i int) *Relation { return c.items[i] }

// All returns a copy of the internal slice.
func (c *RelationCollection) All() []*Relation {
	out := make([]*Relation, len(c.items))
	copy(out, c.items)
	return out
}

// FindByName returns the relation with the given name, or nil if not found.
func (c *RelationCollection) FindByName(name string) *Relation {
	for _, v := range c.items {
		if v.Name == name {
			return v
		}
	}
	return nil
}

// FindByAlias returns the relation with the given alias, or nil if not found.
func (c *RelationCollection) FindByAlias(alias string) *Relation {
	for _, v := range c.items {
		if v.Alias == alias {
			return v
		}
	}
	return nil
}

// FindEqual returns the first relation that is equal to rel, or nil if not found.
func (c *RelationCollection) FindEqual(rel *Relation) *Relation {
	for _, v := range c.items {
		if v.Equals(rel) {
			return v
		}
	}
	return nil
}
