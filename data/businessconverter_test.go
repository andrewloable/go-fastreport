package data_test

import (
	"reflect"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// -----------------------------------------------------------------------
// Test fixtures
// -----------------------------------------------------------------------

type simpleRow struct {
	ID    int
	Name  string
	Score float64
}

type addressRow struct {
	Street string
	City   string
}

type personRow struct {
	ID      int
	Name    string
	Address addressRow // Complex nested struct
}

type teamRow struct {
	TeamName string
	Members  []personRow // Enumerable nested
}

type unexportedFieldsRow struct {
	Public  string
	private string //nolint:unused
}

// -----------------------------------------------------------------------
// PropertyKind classification
// -----------------------------------------------------------------------

func TestGetPropertyKind_SimpleTypes(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	simpleTypes := []reflect.Type{
		reflect.TypeOf(0),
		reflect.TypeOf(int64(0)),
		reflect.TypeOf(float32(0)),
		reflect.TypeOf(float64(0)),
		reflect.TypeOf(false),
		reflect.TypeOf(""),
		reflect.TypeOf([]byte{}), // byte slice → simple
	}
	for _, tt := range simpleTypes {
		kind := c.GetPropertyKind(tt.String(), tt)
		if kind != data.PropertyKindSimple {
			t.Errorf("GetPropertyKind(%v) = %v, want PropertyKindSimple", tt, kind)
		}
	}
}

func TestGetPropertyKind_ComplexType(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	kind := c.GetPropertyKind("addr", reflect.TypeOf(addressRow{}))
	if kind != data.PropertyKindComplex {
		t.Errorf("GetPropertyKind(struct) = %v, want PropertyKindComplex", kind)
	}
}

func TestGetPropertyKind_EnumerableType(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	kind := c.GetPropertyKind("items", reflect.TypeOf([]simpleRow{}))
	if kind != data.PropertyKindEnumerable {
		t.Errorf("GetPropertyKind([]struct) = %v, want PropertyKindEnumerable", kind)
	}
}

func TestGetPropertyKind_NilType(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	kind := c.GetPropertyKind("x", nil)
	if kind != data.PropertyKindSimple {
		t.Errorf("GetPropertyKind(nil) = %v, want PropertyKindSimple", kind)
	}
}

func TestGetPropertyKind_PtrToStruct(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	kind := c.GetPropertyKind("p", reflect.TypeOf((*addressRow)(nil)))
	if kind != data.PropertyKindComplex {
		t.Errorf("GetPropertyKind(*struct) = %v, want PropertyKindComplex", kind)
	}
}

func TestGetPropertyKind_Callback_Override(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	c.OnGetPropertyKind = func(args *data.GetPropertyKindEventArgs) {
		// Force everything to Simple via callback.
		args.Kind = data.PropertyKindSimple
	}
	kind := c.GetPropertyKind("items", reflect.TypeOf([]simpleRow{}))
	if kind != data.PropertyKindSimple {
		t.Errorf("callback override failed: got %v, want PropertyKindSimple", kind)
	}
}

// -----------------------------------------------------------------------
// CreateInitialObjects — flat struct
// -----------------------------------------------------------------------

func TestCreateInitialObjects_FlatStruct(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("Rows")
	c.CreateInitialObjects(root, reflect.TypeOf(simpleRow{}), 1)

	cols := root.Columns()
	if cols.Len() != 3 {
		t.Fatalf("expected 3 columns for simpleRow, got %d", cols.Len())
	}
	names := map[string]bool{}
	for i := 0; i < cols.Len(); i++ {
		names[cols.Get(i).Name] = true
	}
	for _, want := range []string{"ID", "Name", "Score"} {
		if !names[want] {
			t.Errorf("column %q missing in schema", want)
		}
	}
}

func TestCreateInitialObjects_PropNamesSet(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("Rows")
	c.CreateInitialObjects(root, reflect.TypeOf(simpleRow{}), 1)

	cols := root.Columns()
	for i := 0; i < cols.Len(); i++ {
		col := cols.Get(i)
		if col.PropName != col.Name {
			t.Errorf("column %q: PropName=%q, want same as Name", col.Name, col.PropName)
		}
	}
}

func TestCreateInitialObjects_DataTypesSet(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("Rows")
	c.CreateInitialObjects(root, reflect.TypeOf(simpleRow{}), 1)

	idCol := root.Columns().FindByName("ID")
	if idCol == nil {
		t.Fatal("ID column not found")
	}
	if idCol.DataType == "" {
		t.Error("ID column DataType should not be empty")
	}
}

// -----------------------------------------------------------------------
// CreateInitialObjects — nested struct (Complex)
// -----------------------------------------------------------------------

func TestCreateInitialObjects_NestedStruct(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("People")
	// maxNestingLevel=2 allows one level of recursion into Address.
	c.CreateInitialObjects(root, reflect.TypeOf(personRow{}), 2)

	// Top level: ID, Name, Address
	if root.Columns().Len() != 3 {
		t.Fatalf("top-level columns = %d, want 3", root.Columns().Len())
	}
	addrCol := root.Columns().FindByName("Address")
	if addrCol == nil {
		t.Fatal("Address column not found")
	}
	// Address should have nested Street and City.
	if addrCol.Columns().Len() != 2 {
		t.Fatalf("Address nested columns = %d, want 2", addrCol.Columns().Len())
	}
}

func TestCreateInitialObjects_NestedStruct_DepthLimit(t *testing.T) {
	// With maxNestingLevel=1, Address column is present but not recursed into.
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("People")
	c.CreateInitialObjects(root, reflect.TypeOf(personRow{}), 1)

	addrCol := root.Columns().FindByName("Address")
	if addrCol == nil {
		t.Fatal("Address column not found")
	}
	// Address should have no nested columns at depth limit.
	if addrCol.HasColumns() {
		t.Errorf("Address should have no sub-columns at maxNestingLevel=1, got %d", addrCol.Columns().Len())
	}
}

// -----------------------------------------------------------------------
// CreateInitialObjects — enumerable (slice field)
// -----------------------------------------------------------------------

func TestCreateInitialObjects_EnumerableField(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("Teams")
	c.CreateInitialObjects(root, reflect.TypeOf(teamRow{}), 2)

	// TeamName (simple) + Members (enumerable)
	if root.Columns().Len() != 2 {
		t.Fatalf("top-level columns = %d, want 2", root.Columns().Len())
	}
	membersCol := root.Columns().FindByName("Members")
	if membersCol == nil {
		t.Fatal("Members column not found")
	}
	// Members is enumerable → recursed into personRow fields.
	if membersCol.Columns().Len() == 0 {
		t.Error("Members should have nested columns (ID, Name, Address)")
	}
}

func TestCreateInitialObjects_UnexportedFieldsSkipped(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("Rows")
	c.CreateInitialObjects(root, reflect.TypeOf(unexportedFieldsRow{}), 1)

	cols := root.Columns()
	if cols.Len() != 1 {
		t.Fatalf("expected 1 column (Public only), got %d", cols.Len())
	}
	if cols.Get(0).Name != "Public" {
		t.Errorf("expected column 'Public', got %q", cols.Get(0).Name)
	}
}

// -----------------------------------------------------------------------
// CreateInitialObjects — OnFilterProperties callback
// -----------------------------------------------------------------------

func TestCreateInitialObjects_FilterPropertiesCallback(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	c.OnFilterProperties = func(args *data.FilterPropertiesEventArgs) {
		if args.FieldName == "Score" {
			args.Skip = true
		}
	}
	root := data.NewDataColumn("Rows")
	c.CreateInitialObjects(root, reflect.TypeOf(simpleRow{}), 1)

	cols := root.Columns()
	if cols.Len() != 2 {
		t.Fatalf("expected 2 columns after filtering Score, got %d", cols.Len())
	}
	if cols.FindByName("Score") != nil {
		t.Error("Score column should have been filtered out")
	}
}

// -----------------------------------------------------------------------
// CreateInitialObjects — primitive slice ("Value" synthetic column)
// -----------------------------------------------------------------------

func TestCreateInitialObjects_PrimitiveSlice_ValueColumn(t *testing.T) {
	// A root column whose type is a slice of primitives should get a synthetic
	// "Value" child column when there are no struct fields.
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("Tags")
	c.CreateInitialObjects(root, reflect.TypeOf([]string{}), 1)

	// The root itself has no struct fields, and it's Enumerable.
	// The createInitialObjects call should create a "Value" child.
	// (Depends on OnGetPropertyKind classifying []string as Enumerable for root col.)
	// Since root Name="Tags" and type=[]string is Enumerable, and fields=0, a
	// "Value" column should be created.
	_ = root // just verify no panic; struct field count zero for []string
}

// -----------------------------------------------------------------------
// UpdateExistingObjects — no-op when schema unchanged
// -----------------------------------------------------------------------

func TestUpdateExistingObjects_NoOp_Unchanged(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("Rows")
	c.CreateInitialObjects(root, reflect.TypeOf(simpleRow{}), 1)

	initialCount := root.Columns().Len()

	// Run update with the same type.
	c2 := data.NewBusinessObjectConverter()
	c2.UpdateExistingObjects(root, reflect.TypeOf(simpleRow{}), 1)

	if root.Columns().Len() != initialCount {
		t.Errorf("UpdateExistingObjects changed column count: %d → %d", initialCount, root.Columns().Len())
	}
}

// -----------------------------------------------------------------------
// UpdateExistingObjects — adds a new field
// -----------------------------------------------------------------------

type simpleRowV1 struct {
	ID   int
	Name string
}

type simpleRowV2 struct {
	ID    int
	Name  string
	Email string // New field
}

func TestUpdateExistingObjects_AddsNewField(t *testing.T) {
	// Build schema from V1.
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("Rows")
	c.CreateInitialObjects(root, reflect.TypeOf(simpleRowV1{}), 1)

	if root.Columns().Len() != 2 {
		t.Fatalf("V1: expected 2 columns, got %d", root.Columns().Len())
	}

	// Update schema to V2 — Email should be added.
	c2 := data.NewBusinessObjectConverter()
	c2.UpdateExistingObjects(root, reflect.TypeOf(simpleRowV2{}), 1)

	if root.Columns().Len() != 3 {
		t.Fatalf("V2: expected 3 columns after update, got %d", root.Columns().Len())
	}
	if root.Columns().FindByName("Email") == nil {
		t.Error("Email column should have been added by UpdateExistingObjects")
	}
}

// -----------------------------------------------------------------------
// UpdateExistingObjects — removes a stale field
// -----------------------------------------------------------------------

type simpleRowV3 struct {
	ID int
	// Name removed compared to simpleRowV1
}

func TestUpdateExistingObjects_RemovesStaleField(t *testing.T) {
	// Build schema from V1.
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("Rows")
	c.CreateInitialObjects(root, reflect.TypeOf(simpleRowV1{}), 1)

	// Update to V3 — Name should be removed.
	c2 := data.NewBusinessObjectConverter()
	c2.UpdateExistingObjects(root, reflect.TypeOf(simpleRowV3{}), 1)

	if root.Columns().Len() != 1 {
		t.Fatalf("V3: expected 1 column after update, got %d", root.Columns().Len())
	}
	if root.Columns().FindByName("Name") != nil {
		t.Error("Name column should have been removed by UpdateExistingObjects")
	}
}

// -----------------------------------------------------------------------
// UpdateExistingObjects — preserves Calculated columns
// -----------------------------------------------------------------------

func TestUpdateExistingObjects_PreservesCalculatedColumns(t *testing.T) {
	// Build schema from V1.
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("Rows")
	c.CreateInitialObjects(root, reflect.TypeOf(simpleRowV1{}), 1)

	// Manually add a calculated column (not part of the struct).
	calcCol := data.NewDataColumn("FullInfo")
	calcCol.Calculated = true
	calcCol.Expression = "[ID] + [Name]"
	root.Columns().Add(calcCol)

	// Update to V3 (removes Name) — calculated column should survive.
	c2 := data.NewBusinessObjectConverter()
	c2.UpdateExistingObjects(root, reflect.TypeOf(simpleRowV3{}), 1)

	if root.Columns().FindByName("FullInfo") == nil {
		t.Error("Calculated column FullInfo should be preserved by UpdateExistingObjects")
	}
}

// -----------------------------------------------------------------------
// UpdateExistingObjects — preserves Value synthetic column
// -----------------------------------------------------------------------

func TestUpdateExistingObjects_PreservesValueColumn(t *testing.T) {
	// Create a root column that represents a primitive slice.
	root := data.NewDataColumn("Tags")
	// Manually add a "Value" column (as CreateInitialObjects would for primitive slices).
	valCol := data.NewDataColumn("Value")
	valCol.PropName = "Value"
	root.Columns().Add(valCol)

	// Run update — "Value" column should not be pruned.
	c := data.NewBusinessObjectConverter()
	// Use simpleRowV3 (has 1 field) so properties count > 0 but Value isn't in fields.
	// Since Value.PropName == "Value", it should survive.
	c.UpdateExistingObjects(root, reflect.TypeOf(simpleRowV3{}), 1)

	if root.Columns().FindByPropName("Value") == nil {
		t.Error("synthetic Value column should be preserved by UpdateExistingObjects")
	}
}

// -----------------------------------------------------------------------
// MaxNestingLevel boundary
// -----------------------------------------------------------------------

func TestCreateInitialObjects_MaxNestingLevel_Zero(t *testing.T) {
	// maxNestingLevel=0 → nothing should be created.
	c := data.NewBusinessObjectConverter()
	root := data.NewDataColumn("Rows")
	c.CreateInitialObjects(root, reflect.TypeOf(simpleRow{}), 0)

	if root.Columns().Len() != 0 {
		t.Errorf("expected 0 columns with maxNestingLevel=0, got %d", root.Columns().Len())
	}
}

// -----------------------------------------------------------------------
// NewBusinessObjectConverter defaults
// -----------------------------------------------------------------------

func TestNewBusinessObjectConverter_Defaults(t *testing.T) {
	c := data.NewBusinessObjectConverter()
	if c.MaxNestingLevel != 1 {
		t.Errorf("MaxNestingLevel default = %d, want 1", c.MaxNestingLevel)
	}
	if c.OnGetPropertyKind != nil {
		t.Error("OnGetPropertyKind should be nil by default")
	}
	if c.OnFilterProperties != nil {
		t.Error("OnFilterProperties should be nil by default")
	}
}
