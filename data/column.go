package data

import (
	"iter"
	"sort"
	"strconv"
	"strings"

	"github.com/andrewloable/go-fastreport/report"
)

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

// columnFormatNames maps ColumnFormat values to their FRX string representation
// for serialization. The index matches the iota value.
var columnFormatNames = [...]string{
	"Auto", "General", "Number", "Currency",
	"Date", "Time", "Percent", "Boolean",
}

// parseColumnFormat converts an FRX string to a ColumnFormat value.
func parseColumnFormat(s string) ColumnFormat {
	for i, name := range columnFormatNames {
		if name == s {
			return ColumnFormat(i)
		}
	}
	return ColumnFormatAuto
}

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
	// When empty it defaults to the same value as Name.
	PropName string
	// Tag holds arbitrary runtime metadata (not serialized).
	// C# ref: FastReport.Data.Column.Tag (internal)
	Tag any
	// parent points to the owning DataColumn when this is a nested child.
	// Used by FullName() to build the dot-separated qualified name.
	parent *DataColumn
	// columns holds nested child columns (for hierarchical data).
	columns *ColumnCollection
}

// NewDataColumn creates a DataColumn with the given name and enabled=true.
// PropName is initialised to name (matching C# Column constructor + SetName).
func NewDataColumn(name string) *DataColumn {
	return &DataColumn{
		Name:     name,
		Alias:    name,
		PropName: name,
		Enabled:  true,
	}
}

// SetName sets the column's Name and, when PropName was previously synced with
// Name (or empty), keeps PropName in sync. This mirrors C# Column.SetName.
func (c *DataColumn) SetName(name string) {
	// Sync PropName when it was previously identical to Name (case-insensitive)
	// or when it was empty. Matches C# Column.SetName.
	if c.PropName == "" || strings.EqualFold(c.PropName, c.Name) {
		c.PropName = name
	}
	// Sync Alias when it was previously identical to Name (case-insensitive)
	// or when it was empty. Matches C# DataComponentBase.SetName alias-sync logic.
	if c.Alias == "" || strings.EqualFold(c.Alias, c.Name) {
		c.Alias = name
	}
	c.Name = name
}

// Parent returns the owning DataColumn for nested columns, or nil for
// top-level columns.
func (c *DataColumn) Parent() *DataColumn { return c.parent }

// SetParent sets the owning parent column.
func (c *DataColumn) SetParent(p *DataColumn) { c.parent = p }

// FullName returns the dot-separated qualified name of this column,
// walking up through parent columns. For a top-level column this is just
// the Alias; for nested columns it is "Parent.Child".
// C# ref: FastReport.Data.Column.FullName
func (c *DataColumn) FullName() string {
	if c.parent != nil {
		return c.parent.FullName() + "." + c.Alias
	}
	return c.Alias
}

// GetExpressions returns the list of expressions used by this column.
// For calculated columns it returns a single-element slice containing the
// Expression; for regular columns it returns nil.
// C# ref: FastReport.Data.Column.GetExpressions()
func (c *DataColumn) GetExpressions() []string {
	if c.Calculated && c.Expression != "" {
		return []string{c.Expression}
	}
	return nil
}

// Serialize writes the column's non-default properties to w.
// C# ref: FastReport.Data.Column.Serialize(FRWriter)
func (c *DataColumn) Serialize(w report.Writer) error {
	// Name is written as an XML attribute by the caller (FRWriter framework);
	// we write the additional properties.
	if c.Alias != "" && c.Alias != c.Name {
		w.WriteStr("Alias", c.Alias)
	}
	if !c.Enabled {
		w.WriteBool("Enabled", false)
	}
	if c.DataType != "" {
		w.WriteStr("DataType", c.DataType)
	}
	if c.PropName != "" && c.PropName != c.Name {
		w.WriteStr("PropName", c.PropName)
	}
	if c.Format != ColumnFormatAuto {
		w.WriteStr("Format", columnFormatNames[c.Format])
	}
	if c.Calculated {
		w.WriteBool("Calculated", true)
		w.WriteStr("Expression", c.Expression)
	}
	return nil
}

// Deserialize reads the column's properties from r.
// C# ref: FastReport.Data.Column — properties auto-read by FRReader;
// Go requires explicit reads.
func (c *DataColumn) Deserialize(r report.Reader) error {
	c.Alias = r.ReadStr("Alias", c.Name)
	c.Enabled = r.ReadBool("Enabled", true)
	c.DataType = r.ReadStr("DataType", c.DataType)
	propName := r.ReadStr("PropName", "")
	if propName != "" {
		c.PropName = propName
	}
	fmtStr := r.ReadStr("Format", "")
	if fmtStr != "" {
		c.Format = parseColumnFormat(fmtStr)
	}
	c.Calculated = r.ReadBool("Calculated", false)
	if c.Calculated {
		c.Expression = r.ReadStr("Expression", "")
	} else {
		// Expression can exist even when Calculated is false (e.g. saved but
		// toggled off); read it to avoid losing data on round-trip.
		c.Expression = r.ReadStr("Expression", c.Expression)
	}
	return nil
}

// ColumnFormatString returns the FRX string representation of a ColumnFormat.
func ColumnFormatString(f ColumnFormat) string {
	if int(f) >= 0 && int(f) < len(columnFormatNames) {
		return columnFormatNames[f]
	}
	return strconv.Itoa(int(f))
}

// Columns returns the nested column collection, creating it on first access.
// The collection's owner is set to c so that added children get their parent
// pointer set automatically.
func (c *DataColumn) Columns() *ColumnCollection {
	if c.columns == nil {
		c.columns = &ColumnCollection{owner: c}
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
	// owner is the DataColumn that owns this collection (for setting parent pointers).
	// Nil for top-level (free-standing) collections.
	owner *DataColumn
}

// NewColumnCollection creates an empty ColumnCollection with no owner.
func NewColumnCollection() *ColumnCollection {
	return &ColumnCollection{}
}

// Add appends col to the collection. If the collection has an owner (i.e. it
// belongs to a DataColumn's Columns()), the child's parent pointer is set.
func (cc *ColumnCollection) Add(col *DataColumn) {
	if cc.owner != nil {
		col.parent = cc.owner
	}
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

// FindByName returns the first column whose Name matches (case-insensitive), or nil.
func (cc *ColumnCollection) FindByName(name string) *DataColumn {
	for _, col := range cc.items {
		if strings.EqualFold(col.Name, name) {
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

// FindByAlias returns the first column whose Alias matches (case-insensitive), or nil.
func (cc *ColumnCollection) FindByAlias(alias string) *DataColumn {
	for _, col := range cc.items {
		if strings.EqualFold(col.Alias, alias) {
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

// FindByPropName returns the first column whose PropName matches, or nil.
// C# ref: FastReport.Data.Column.FindByPropName(string)
func (cc *ColumnCollection) FindByPropName(propName string) *DataColumn {
	for _, col := range cc.items {
		if col.PropName == propName {
			return col
		}
	}
	return nil
}

// removeAt removes the column at index i (bounds must be valid).
func (cc *ColumnCollection) removeAt(i int) {
	cc.items = append(cc.items[:i], cc.items[i+1:]...)
}

// CreateUniqueName returns a unique column name based on the given base name.
func (cc *ColumnCollection) CreateUniqueName(name string) string {
	base := name
	i := 1
	for cc.FindByName(name) != nil {
		name = base + strconv.Itoa(i)
		i++
	}
	return name
}

// CreateUniqueAlias returns a unique column alias based on the given base alias.
func (cc *ColumnCollection) CreateUniqueAlias(alias string) string {
	base := alias
	i := 1
	for cc.FindByAlias(alias) != nil {
		alias = base + strconv.Itoa(i)
		i++
	}
	return alias
}

// Sort sorts the collection of columns by Name (ascending, case-sensitive).
func (cc *ColumnCollection) Sort() {
	sort.SliceStable(cc.items, func(i, j int) bool {
		return cc.items[i].Name < cc.items[j].Name
	})
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
