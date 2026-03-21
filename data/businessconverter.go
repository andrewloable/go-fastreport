package data

import (
	"reflect"
)

// PropertyKind specifies the kind of a reflected property.
// C# ref: FastReport.Data.PropertyKind
type PropertyKind int

const (
	// PropertyKindSimple indicates a property of a simple (value) type such as
	// int, string, bool, float64, []byte, or time.Time.
	PropertyKindSimple PropertyKind = iota

	// PropertyKindComplex indicates a complex property — a struct or pointer-to-struct
	// whose own fields should be reflected recursively.
	PropertyKindComplex

	// PropertyKindEnumerable indicates a property that is a slice, array, or other
	// iterable sequence of objects (equivalent to C# IEnumerable).
	PropertyKindEnumerable
)

// GetPropertyKindEventArgs carries the context passed to the OnGetPropertyKind
// callback. The callback can override Kind before BusinessObjectConverter uses it.
// C# ref: FastReport.Data.GetPropertyKindEventArgs
type GetPropertyKindEventArgs struct {
	// Name is the field/property name being classified.
	Name string
	// Type is the reflect.Type of the field.
	Type reflect.Type
	// Kind is the default classification; the callback may change it.
	Kind PropertyKind
}

// FilterPropertiesEventArgs carries the context passed to the OnFilterProperties
// callback. Set Skip = true to exclude the field from the generated schema.
// C# ref: FastReport.Data.FilterPropertiesEventArgs
type FilterPropertiesEventArgs struct {
	// FieldName is the struct field name.
	FieldName string
	// FieldType is the reflect.Type of the field.
	FieldType reflect.Type
	// Skip controls whether this field should be omitted. Default false.
	Skip bool
}

// BusinessObjectConverter builds and updates a DataColumn schema tree from a Go
// struct type using reflection. It is the Go port of the internal C# class
// FastReport.Base/Data/BusinessObjectConverter.cs.
//
// Typical usage:
//
//	type Order struct {
//	    ID       int
//	    Customer string
//	    Lines    []OrderLine
//	}
//	root := data.NewDataColumn("Orders")
//	root.DataType = "[]Order"
//	conv := data.NewBusinessObjectConverter()
//	conv.CreateInitialObjects(root, reflect.TypeOf(Order{}), 2)
type BusinessObjectConverter struct {
	// MaxNestingLevel is the maximum recursion depth for schema generation.
	// Defaults to 1 (one level of nesting).
	MaxNestingLevel int

	// OnGetPropertyKind is an optional callback invoked for each field to allow
	// callers to override the automatic PropertyKind classification.
	// C# ref: Config.ReportSettings.OnGetBusinessObjectPropertyKind
	OnGetPropertyKind func(args *GetPropertyKindEventArgs)

	// OnFilterProperties is an optional callback invoked for each field to allow
	// callers to skip (exclude) specific fields from the generated schema.
	// C# ref: Config.ReportSettings.OnFilterBusinessObjectProperties
	OnFilterProperties func(args *FilterPropertiesEventArgs)

	// nestingLevel tracks the current recursion depth during schema generation.
	nestingLevel int
}

// NewBusinessObjectConverter creates a BusinessObjectConverter with default
// settings (MaxNestingLevel = 1, no callbacks).
func NewBusinessObjectConverter() *BusinessObjectConverter {
	return &BusinessObjectConverter{
		MaxNestingLevel: 1,
	}
}

// GetPropertyKind classifies a reflect.Type as Simple, Complex, or Enumerable.
// If OnGetPropertyKind is set, the callback can override the default classification.
// C# ref: BusinessObjectConverter.GetPropertyKind(string, Type)
func (c *BusinessObjectConverter) GetPropertyKind(name string, t reflect.Type) PropertyKind {
	if t == nil {
		return PropertyKindSimple
	}

	// Dereference pointers to get the underlying type for classification.
	base := t
	for base.Kind() == reflect.Ptr {
		base = base.Elem()
	}

	kind := PropertyKindComplex

	switch base.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		// Value types and string → Simple
		kind = PropertyKindSimple

	case reflect.Slice, reflect.Array:
		// []byte is treated as a simple binary blob (matches C# byte[]).
		if base.Elem().Kind() == reflect.Uint8 {
			kind = PropertyKindSimple
		} else {
			kind = PropertyKindEnumerable
		}

	case reflect.Map:
		// Maps are treated as Enumerable (similar to IEnumerable in C#).
		kind = PropertyKindEnumerable

	default:
		// Struct, interface, etc. → Complex unless overridden.
		kind = PropertyKindComplex
	}

	// Allow the caller to override via callback.
	if c.OnGetPropertyKind != nil {
		args := &GetPropertyKindEventArgs{Name: name, Type: t, Kind: kind}
		c.OnGetPropertyKind(args)
		kind = args.Kind
	}

	return kind
}

// isSimpleType returns true when the type is classified as PropertyKindSimple.
func (c *BusinessObjectConverter) isSimpleType(name string, t reflect.Type) bool {
	return c.GetPropertyKind(name, t) == PropertyKindSimple
}

// isEnumerable returns true when the type is classified as PropertyKindEnumerable.
func (c *BusinessObjectConverter) isEnumerable(name string, t reflect.Type) bool {
	return c.GetPropertyKind(name, t) == PropertyKindEnumerable
}

// isLoop detects whether adding a column of the given reflect.Type would create
// an infinite recursion cycle by checking *ancestor* columns (not the column
// itself). This mirrors C# IsLoop which walks column.Parent, not column.
// C# ref: BusinessObjectConverter.IsLoop(Column, Type)
func (c *BusinessObjectConverter) isLoop(col *DataColumn, t reflect.Type) bool {
	// Start from the parent to avoid falsely flagging col itself.
	cur := col.parent
	for cur != nil {
		if cur.Tag != nil {
			if curType, ok := cur.Tag.(reflect.Type); ok && curType == t {
				return true
			}
		}
		cur = cur.parent
	}
	return false
}

// getStructFields returns the exported fields of a struct type, applying the
// OnFilterProperties callback when set.
// C# ref: BusinessObjectConverter.GetProperties(Column)
func (c *BusinessObjectConverter) getStructFields(t reflect.Type) []reflect.StructField {
	// Unwrap pointer and slice/array to reach the struct.
	base := t
	for base.Kind() == reflect.Ptr {
		base = base.Elem()
	}
	if base.Kind() == reflect.Slice || base.Kind() == reflect.Array {
		base = base.Elem()
		for base.Kind() == reflect.Ptr {
			base = base.Elem()
		}
	}
	if base.Kind() != reflect.Struct {
		return nil
	}

	out := make([]reflect.StructField, 0, base.NumField())
	for i := 0; i < base.NumField(); i++ {
		f := base.Field(i)
		if !f.IsExported() {
			continue
		}
		if c.OnFilterProperties != nil {
			args := &FilterPropertiesEventArgs{FieldName: f.Name, FieldType: f.Type}
			c.OnFilterProperties(args)
			if args.Skip {
				continue
			}
		}
		out = append(out, f)
	}
	return out
}

// elemTypeOf returns the element type for slice/array types; for other kinds it
// returns t unchanged.
func elemTypeOf(t reflect.Type) reflect.Type {
	base := t
	for base.Kind() == reflect.Ptr {
		base = base.Elem()
	}
	if base.Kind() == reflect.Slice || base.Kind() == reflect.Array {
		elem := base.Elem()
		for elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		return elem
	}
	return base
}

// CreateInitialObjects builds a full schema tree on col using structType as the
// root Go type. maxNestingLevel limits recursion depth (1 = one level of children).
//
// C# ref: BusinessObjectConverter.CreateInitialObjects(Column, int)
func (c *BusinessObjectConverter) CreateInitialObjects(col *DataColumn, structType reflect.Type, maxNestingLevel int) {
	c.MaxNestingLevel = maxNestingLevel
	c.nestingLevel = 0
	c.createInitialObjects(col, structType)
}

// createInitialObjects is the recursive inner implementation.
// C# ref: BusinessObjectConverter.CreateInitialObjects(Column) — private
func (c *BusinessObjectConverter) createInitialObjects(col *DataColumn, t reflect.Type) {
	if c.nestingLevel >= c.MaxNestingLevel {
		return
	}
	c.nestingLevel++

	fields := c.getStructFields(t)
	for _, f := range fields {
		fieldType := f.Type

		isSimple := c.isSimpleType(f.Name, fieldType)
		isEnum := c.isEnumerable(f.Name, fieldType)

		child := NewDataColumn(f.Name)
		child.Alias = f.Name
		child.DataType = fieldType.String()
		child.PropName = f.Name
		// Tag stores the reflect.Type for loop detection.
		child.Tag = fieldType
		child.Enabled = !isEnum || c.nestingLevel < c.MaxNestingLevel

		col.Columns().Add(child)

		if !isSimple {
			if isEnum {
				// Recurse into the element type of the slice/array.
				elemType := elemTypeOf(fieldType)
				if !c.isLoop(child, elemType) {
					c.createInitialObjects(child, elemType)
				}
			} else {
				// Complex (struct): recurse into the struct itself.
				base := fieldType
				for base.Kind() == reflect.Ptr {
					base = base.Elem()
				}
				if !c.isLoop(child, base) {
					c.createInitialObjects(child, base)
				}
			}
		}
	}

	// If this is an enumerable with no discoverable struct fields, create a
	// synthetic "Value" column for the element type.
	// C# ref: BusinessObjectConverter.CreateListValueColumn(Column)
	if c.isEnumerable(col.Name, t) && len(fields) == 0 {
		valueType := elemTypeOf(t)
		value := NewDataColumn("Value")
		value.DataType = valueType.String()
		value.PropName = "Value"
		value.Enabled = c.isSimpleType("Value", valueType)
		col.Columns().Add(value)
	}

	c.nestingLevel--
}

// UpdateExistingObjects performs a delta update of an already-built schema tree.
// New fields are added, removed fields are deleted, and existing entries have
// their metadata refreshed. maxNestingLevel limits recursion depth.
//
// C# ref: BusinessObjectConverter.UpdateExistingObjects(Column, int)
func (c *BusinessObjectConverter) UpdateExistingObjects(col *DataColumn, structType reflect.Type, maxNestingLevel int) {
	c.MaxNestingLevel = maxNestingLevel
	c.nestingLevel = 0
	c.updateExistingObjects(col, structType)
}

// updateExistingObjects is the recursive inner implementation.
// C# ref: BusinessObjectConverter.UpdateExistingObjects(Column) — private
func (c *BusinessObjectConverter) updateExistingObjects(col *DataColumn, t reflect.Type) {
	c.nestingLevel++

	cols := col.Columns()

	// Reset Tag on every child so we can detect stale (removed) columns after
	// the property walk. We reuse Tag as a "seen" flag here: nil means unseen.
	// C# ref: reset PropDescriptor → null before the loop.
	for i := 0; i < cols.Len(); i++ {
		cols.Get(i).Tag = nil
	}

	fields := c.getStructFields(t)
	if len(fields) > 0 {
		for _, f := range fields {
			fieldType := f.Type

			isSimple := c.isSimpleType(f.Name, fieldType)
			isEnum := c.isEnumerable(f.Name, fieldType)

			// Find an existing child column by PropName.
			child := cols.FindByPropName(f.Name)

			if child == nil {
				// New field: create and add.
				child = NewDataColumn(f.Name)
				child.Alias = f.Name
				if isEnum {
					child.Name = cols.CreateUniqueName(f.Name)
				}
				// Enable when simple or when we haven't exhausted nesting levels.
				child.Enabled = isSimple || c.nestingLevel < c.MaxNestingLevel
				cols.Add(child)
			}

			// Update mutable metadata.
			child.DataType = fieldType.String()
			child.PropName = f.Name
			// Mark as seen by storing the type in Tag.
			child.Tag = fieldType

			if child.Enabled && !isSimple {
				if isEnum {
					elemType := elemTypeOf(fieldType)
					if !c.isLoop(child, elemType) {
						c.updateExistingObjects(child, elemType)
					}
				} else {
					base := fieldType
					for base.Kind() == reflect.Ptr {
						base = base.Elem()
					}
					if !c.isLoop(child, base) {
						c.updateExistingObjects(child, base)
					}
				}
			}
		}

		// Remove stale columns: those whose Tag was not refreshed above (nil)
		// and that are not calculated and not the synthetic "Value" column.
		// C# ref: remove columns where PropDescriptor == null && !Calculated && PropName != "Value"
		for i := 0; i < cols.Len(); i++ {
			child := cols.Get(i)
			if child.Tag == nil && !child.Calculated && child.PropName != "Value" {
				cols.removeAt(i)
				i--
			}
		}
	} else if c.isEnumerable(col.Name, t) {
		// No struct fields → ensure the "Value" synthetic column exists.
		// C# ref: CreateListValueColumn(column)
		if cols.FindByPropName("Value") == nil {
			valueType := elemTypeOf(t)
			value := NewDataColumn("Value")
			value.DataType = valueType.String()
			value.PropName = "Value"
			value.Enabled = c.isSimpleType("Value", valueType)
			cols.Add(value)
		}
	}

	c.nestingLevel--
}
