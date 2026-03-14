package data_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// -----------------------------------------------------------------------
// Test fixtures
// -----------------------------------------------------------------------

type product struct {
	ID    int
	Name  string
	Price float64
}

var testProducts = []product{
	{1, "Widget", 9.99},
	{2, "Gadget", 24.99},
	{3, "Doohickey", 4.99},
}

// -----------------------------------------------------------------------
// Construction and metadata
// -----------------------------------------------------------------------

func TestNewBusinessObjectDataSource_Name(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Products", testProducts)
	if ds.Name() != "Products" {
		t.Errorf("Name = %q, want Products", ds.Name())
	}
	if ds.Alias() != "Products" {
		t.Errorf("Alias default = %q, want Products", ds.Alias())
	}
}

func TestBusinessObjectDataSource_SetAlias(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Products", testProducts)
	ds.SetAlias("Prods")
	if ds.Alias() != "Prods" {
		t.Errorf("Alias = %q, want Prods", ds.Alias())
	}
}

// -----------------------------------------------------------------------
// Init and column reflection
// -----------------------------------------------------------------------

func TestBusinessObjectDataSource_Init_Struct(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Products", testProducts)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if ds.RowCount() != 3 {
		t.Errorf("RowCount = %d, want 3", ds.RowCount())
	}
	cols := ds.Columns()
	if len(cols) != 3 {
		t.Fatalf("Columns len = %d, want 3 (ID, Name, Price)", len(cols))
	}
	names := map[string]bool{}
	for _, c := range cols {
		names[c.Name] = true
	}
	for _, want := range []string{"ID", "Name", "Price"} {
		if !names[want] {
			t.Errorf("Column %q not found in metadata", want)
		}
	}
}

func TestBusinessObjectDataSource_Init_EmptySlice(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Empty", []product{})
	if err := ds.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if ds.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0", ds.RowCount())
	}
}

func TestBusinessObjectDataSource_Init_Nil(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Nil", nil)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if ds.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0", ds.RowCount())
	}
}

// -----------------------------------------------------------------------
// Iteration
// -----------------------------------------------------------------------

func TestBusinessObjectDataSource_Iteration(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Products", testProducts)
	_ = ds.Init()
	_ = ds.First()

	count := 0
	for !ds.EOF() {
		count++
		if err := ds.Next(); err != nil && err != data.ErrEOF {
			t.Fatalf("Next error: %v", err)
		}
	}
	if count != 3 {
		t.Errorf("iterated %d rows, want 3", count)
	}
}

func TestBusinessObjectDataSource_EOF_BeforeFirst(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Products", testProducts)
	_ = ds.Init()
	// Before First(), rowIdx = -1, which is < len(rows), so not EOF.
	// After First(), rowIdx = 0.
	_ = ds.First()
	if ds.EOF() {
		t.Error("should not be EOF after First() on non-empty source")
	}
}

func TestBusinessObjectDataSource_CurrentRowNo(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Products", testProducts)
	_ = ds.Init()
	_ = ds.First()
	if ds.CurrentRowNo() != 0 {
		t.Errorf("CurrentRowNo after First = %d, want 0", ds.CurrentRowNo())
	}
	_ = ds.Next()
	if ds.CurrentRowNo() != 1 {
		t.Errorf("CurrentRowNo after Next = %d, want 1", ds.CurrentRowNo())
	}
}

// -----------------------------------------------------------------------
// GetValue — struct rows
// -----------------------------------------------------------------------

func TestBusinessObjectDataSource_GetValue_Struct(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Products", testProducts)
	_ = ds.Init()
	_ = ds.First()

	id, err := ds.GetValue("ID")
	if err != nil {
		t.Fatalf("GetValue(ID) error: %v", err)
	}
	if id != 1 {
		t.Errorf("GetValue(ID) = %v, want 1", id)
	}

	name, err := ds.GetValue("Name")
	if err != nil {
		t.Fatalf("GetValue(Name) error: %v", err)
	}
	if name != "Widget" {
		t.Errorf("GetValue(Name) = %v, want Widget", name)
	}
}

func TestBusinessObjectDataSource_GetValue_CaseInsensitive(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Products", testProducts)
	_ = ds.Init()
	_ = ds.First()

	v, err := ds.GetValue("name") // lowercase
	if err != nil {
		t.Fatalf("GetValue(name) error: %v", err)
	}
	if v != "Widget" {
		t.Errorf("GetValue(name) = %v, want Widget", v)
	}
}

func TestBusinessObjectDataSource_GetValue_UnknownField(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Products", testProducts)
	_ = ds.Init()
	_ = ds.First()
	_, err := ds.GetValue("NoSuchField")
	if err == nil {
		t.Error("expected error for unknown field, got nil")
	}
}

// -----------------------------------------------------------------------
// GetValue — map rows
// -----------------------------------------------------------------------

func TestBusinessObjectDataSource_GetValue_Map(t *testing.T) {
	rows := []map[string]any{
		{"city": "Paris", "pop": 2161000},
		{"city": "Berlin", "pop": 3669491},
	}
	ds := data.NewBusinessObjectDataSource("Cities", rows)
	_ = ds.Init()
	_ = ds.First()

	city, err := ds.GetValue("city")
	if err != nil {
		t.Fatalf("GetValue(city) error: %v", err)
	}
	if city != "Paris" {
		t.Errorf("GetValue(city) = %v, want Paris", city)
	}
}

func TestBusinessObjectDataSource_GetValue_Map_MissingKey(t *testing.T) {
	rows := []map[string]any{{"a": 1}}
	ds := data.NewBusinessObjectDataSource("M", rows)
	_ = ds.Init()
	_ = ds.First()
	_, err := ds.GetValue("z")
	if err == nil {
		t.Error("expected error for missing map key, got nil")
	}
}

// -----------------------------------------------------------------------
// GetValue — slice of primitives
// -----------------------------------------------------------------------

func TestBusinessObjectDataSource_GetValue_Primitives(t *testing.T) {
	nums := []int{10, 20, 30}
	ds := data.NewBusinessObjectDataSource("Nums", nums)
	_ = ds.Init()
	_ = ds.First()

	v, err := ds.GetValue("Value")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if v != 10 {
		t.Errorf("GetValue = %v, want 10", v)
	}
}

// -----------------------------------------------------------------------
// Pointer slice
// -----------------------------------------------------------------------

func TestBusinessObjectDataSource_PointerSlice(t *testing.T) {
	rows := []*product{
		{4, "Thing", 1.5},
		{5, "Stuff", 2.5},
	}
	ds := data.NewBusinessObjectDataSource("Ptrs", rows)
	_ = ds.Init()
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
	_ = ds.First()
	v, err := ds.GetValue("Name")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if v != "Thing" {
		t.Errorf("GetValue(Name) = %v, want Thing", v)
	}
}

// -----------------------------------------------------------------------
// LoadBusinessObject callback
// -----------------------------------------------------------------------

func TestBusinessObjectDataSource_LoadCallback(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Lazy", nil)
	called := false
	ds.LoadBusinessObject = func(d *data.BusinessObjectDataSource) {
		called = true
		d.SetData([]product{{99, "Loaded", 0.01}})
	}
	_ = ds.Init()
	if !called {
		t.Error("LoadBusinessObject callback should have been called")
	}
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1 after lazy load", ds.RowCount())
	}
}

// -----------------------------------------------------------------------
// SetData re-init
// -----------------------------------------------------------------------

func TestBusinessObjectDataSource_SetData(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("P", testProducts)
	_ = ds.Init()
	ds.SetData([]product{{7, "New", 5.0}})
	_ = ds.Init()
	if ds.RowCount() != 1 {
		t.Errorf("RowCount after SetData = %d, want 1", ds.RowCount())
	}
}

// -----------------------------------------------------------------------
// Close
// -----------------------------------------------------------------------

func TestBusinessObjectDataSource_Close(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Products", testProducts)
	if err := ds.Close(); err != nil {
		t.Errorf("Close error: %v", err)
	}
}

// -----------------------------------------------------------------------
// NotInitialized
// -----------------------------------------------------------------------

func TestBusinessObjectDataSource_First_NotInited(t *testing.T) {
	ds := data.NewBusinessObjectDataSource("Products", testProducts)
	err := ds.First()
	if err == nil {
		t.Error("expected ErrNotInitialized, got nil")
	}
}
