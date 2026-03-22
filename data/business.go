package data

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/andrewloable/go-fastreport/report"
)

// BusinessObjectDataSource binds a Go slice (or any value implementing
// []T where T is a struct or map) to a report band at run time.
// It is the Go equivalent of FastReport.Data.BusinessObjectDataSource.
//
// Usage:
//
//	type Order struct{ ID int; Customer string; Amount float64 }
//	orders := []Order{{1,"Alice",99.9},{2,"Bob",50.0}}
//	ds := data.NewBusinessObjectDataSource("Orders", orders)
//	ds.Init()     // reflects columns from Order struct
//	ds.First()
//	for !ds.EOF() {
//	    v, _ := ds.GetValue("Customer")
//	    ds.Next()
//	}
type BusinessObjectDataSource struct {
	name    string
	alias   string
	enabled bool
	rows    []reflect.Value // each element is one row (struct, map, or primitive)
	rowIdx  int
	inited  bool
	columns []Column
	// rawValue is the original slice/array value provided by the caller.
	rawValue any
	// propName is the property/field name that maps this source to its parent.
	// C# ref: FastReport.Data.DataSourceBase.PropName
	propName string
	// referenceName is the legacy infrastructure reference name from FRX.
	// C# ref: FastReport.Data.DataComponentBase.ReferenceName
	referenceName string
	// LoadBusinessObject is called before data is loaded, enabling load-on-demand.
	LoadBusinessObject func(ds *BusinessObjectDataSource)
}

// NewBusinessObjectDataSource creates a BusinessObjectDataSource bound to the
// given Go value. value must be a slice, array, or a single struct/map.
func NewBusinessObjectDataSource(name string, value any) *BusinessObjectDataSource {
	return &BusinessObjectDataSource{
		name:     name,
		alias:    name,
		enabled:  true,
		rawValue: value,
	}
}

// Name returns the data source name.
func (b *BusinessObjectDataSource) Name() string { return b.name }

// Alias returns the data source alias.
func (b *BusinessObjectDataSource) Alias() string { return b.alias }

// SetAlias sets the alias.
func (b *BusinessObjectDataSource) SetAlias(a string) { b.alias = a }

// Init reflects the bound value to build row data and column metadata.
func (b *BusinessObjectDataSource) Init() error {
	if b.LoadBusinessObject != nil {
		b.LoadBusinessObject(b)
	}
	if b.rawValue == nil {
		b.rows = nil
		b.inited = true
		return nil
	}
	rv := reflect.ValueOf(b.rawValue)
	// dereference pointer
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			b.rows = nil
			b.inited = true
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		b.rows = make([]reflect.Value, rv.Len())
		for i := range b.rows {
			b.rows[i] = rv.Index(i)
		}
	default:
		// Treat a single value as a one-row source.
		b.rows = []reflect.Value{rv}
	}

	// Build column metadata from the first row, if available.
	if len(b.rows) > 0 {
		b.columns = columnsFor(b.rows[0])
	}
	b.rowIdx = -1
	b.inited = true
	return nil
}

// First positions at the first row.
// Returns ErrEOF when the data source has no rows (consistent with other
// data source implementations so that RunDataBandFull can detect empty sources).
func (b *BusinessObjectDataSource) First() error {
	if !b.inited {
		return ErrNotInitialized
	}
	b.rowIdx = 0
	if len(b.rows) == 0 {
		return ErrEOF
	}
	return nil
}

// Next advances to the next row.
func (b *BusinessObjectDataSource) Next() error {
	b.rowIdx++
	if b.rowIdx >= len(b.rows) {
		return ErrEOF
	}
	return nil
}

// EOF returns true when all rows have been consumed.
func (b *BusinessObjectDataSource) EOF() bool {
	return b.rowIdx >= len(b.rows)
}

// RowCount returns the total number of rows.
func (b *BusinessObjectDataSource) RowCount() int { return len(b.rows) }

// CurrentRowNo returns the 0-based current row index.
func (b *BusinessObjectDataSource) CurrentRowNo() int { return b.rowIdx }

// GetValue returns the value of the named column in the current row.
// For struct rows, name is the field name (case-insensitive).
// For map[string]any rows, name is the map key.
func (b *BusinessObjectDataSource) GetValue(column string) (any, error) {
	if b.EOF() || len(b.rows) == 0 {
		return nil, fmt.Errorf("data source %q: no current row", b.name)
	}
	row := b.rows[b.rowIdx]
	return fieldValue(row, column)
}

// Columns returns the column metadata (populated after Init).
func (b *BusinessObjectDataSource) Columns() []Column { return b.columns }

// Close is a no-op for in-memory data sources.
func (b *BusinessObjectDataSource) Close() error { return nil }

// SetData replaces the bound value and marks the source as not initialized.
func (b *BusinessObjectDataSource) SetData(value any) {
	b.rawValue = value
	b.inited = false
	b.rows = nil
}

// Enabled returns whether this data source is active during report execution.
// C# ref: FastReport.Data.DataComponentBase.Enabled
func (b *BusinessObjectDataSource) Enabled() bool { return b.enabled }

// SetEnabled enables or disables this data source.
func (b *BusinessObjectDataSource) SetEnabled(v bool) { b.enabled = v }

// SetName sets the data source name.  When alias was previously equal to
// name (case-insensitively) or empty, it is kept in sync with the new name.
// Mirrors C# DataComponentBase.SetName alias-sync behaviour.
func (b *BusinessObjectDataSource) SetName(n string) {
	if b.alias == "" || strings.EqualFold(b.alias, b.name) {
		b.alias = n
	}
	b.name = n
}

// PropName returns the property/field name used to bind this source to its parent.
// C# ref: FastReport.Data.DataSourceBase.PropName
func (b *BusinessObjectDataSource) PropName() string { return b.propName }

// SetPropName sets the property name used to bind this source to its parent.
func (b *BusinessObjectDataSource) SetPropName(p string) { b.propName = p }

// ReferenceName returns the FRX infrastructure reference name.
// C# ref: FastReport.Data.DataComponentBase.ReferenceName
func (b *BusinessObjectDataSource) ReferenceName() string { return b.referenceName }

// SetReferenceName sets the FRX infrastructure reference name.
func (b *BusinessObjectDataSource) SetReferenceName(n string) { b.referenceName = n }

// Serialize writes the data source's FRX properties to w.
// C# ref: FastReport.Data.BusinessObjectDataSource — inherits DataSourceBase.Serialize.
func (b *BusinessObjectDataSource) Serialize(w report.Writer) error {
	if b.name != "" {
		w.WriteStr("Name", b.name)
	}
	if b.alias != "" && !strings.EqualFold(b.alias, b.name) {
		w.WriteStr("Alias", b.alias)
	}
	if !b.enabled {
		w.WriteBool("Enabled", false)
	}
	if b.referenceName != "" {
		w.WriteStr("ReferenceName", b.referenceName)
	}
	if b.propName != "" {
		w.WriteStr("PropName", b.propName)
	}
	return nil
}

// Deserialize reads the data source's FRX properties from r and handles
// legacy compatibility:
//
//   - If ReferenceName contains a dot (legacy .NET format), the last
//     dot-separated segment becomes PropName and ReferenceName is cleared.
//     C# ref: FastReport.Data.BusinessObjectDataSource.Deserialize line 159-164.
func (b *BusinessObjectDataSource) Deserialize(r report.Reader) error {
	b.name = r.ReadStr("Name", b.name)
	b.alias = r.ReadStr("Alias", b.name)
	b.enabled = r.ReadBool("Enabled", true)
	b.referenceName = r.ReadStr("ReferenceName", "")
	b.propName = r.ReadStr("PropName", "")

	// Compatibility with old reports: if ReferenceName contains a dot,
	// use the last part as PropName and clear ReferenceName.
	// C# ref: BusinessObjectDataSource.Deserialize lines 159-164.
	if b.referenceName != "" && strings.Contains(b.referenceName, ".") {
		parts := strings.Split(b.referenceName, ".")
		if b.propName == "" {
			b.propName = parts[len(parts)-1]
		}
		b.referenceName = ""
	}
	return nil
}

// -----------------------------------------------------------------------
// Reflection helpers
// -----------------------------------------------------------------------

// columnsFor returns column metadata for a single row value.
func columnsFor(row reflect.Value) []Column {
	// dereference pointer
	for row.Kind() == reflect.Ptr {
		if row.IsNil() {
			return nil
		}
		row = row.Elem()
	}
	switch row.Kind() {
	case reflect.Struct:
		t := row.Type()
		cols := make([]Column, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			cols = append(cols, Column{Name: f.Name, Alias: f.Name, DataType: f.Type.String()})
		}
		return cols
	case reflect.Map:
		cols := make([]Column, 0, row.Len())
		for _, k := range row.MapKeys() {
			name := fmt.Sprintf("%v", k.Interface())
			cols = append(cols, Column{Name: name, Alias: name})
		}
		return cols
	default:
		return []Column{{Name: "Value", Alias: "Value", DataType: row.Type().String()}}
	}
}

// fieldValue extracts the named field value from a reflect.Value row.
func fieldValue(row reflect.Value, name string) (any, error) {
	// dereference pointer
	for row.Kind() == reflect.Ptr {
		if row.IsNil() {
			return nil, nil
		}
		row = row.Elem()
	}
	switch row.Kind() {
	case reflect.Struct:
		t := row.Type()
		nameLower := strings.ToLower(name)
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			if strings.ToLower(f.Name) == nameLower {
				return row.Field(i).Interface(), nil
			}
		}
		return nil, fmt.Errorf("field %q not found", name)
	case reflect.Map:
		v := row.MapIndex(reflect.ValueOf(name))
		if !v.IsValid() {
			return nil, fmt.Errorf("key %q not found in map", name)
		}
		return v.Interface(), nil
	default:
		// Single-value row: ignore the name.
		return row.Interface(), nil
	}
}
