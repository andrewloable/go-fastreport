package data_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// -----------------------------------------------------------------------
// Data sources
// -----------------------------------------------------------------------

func TestDictionary_AddAndFindDataSource(t *testing.T) {
	d := data.NewDictionary()
	ds := data.NewBusinessObjectDataSource("Orders", nil)
	ds.SetAlias("Orders")
	d.AddDataSource(ds)

	found := d.FindDataSourceByAlias("Orders")
	if found == nil {
		t.Fatal("FindDataSourceByAlias returned nil")
	}
	if found.Name() != "Orders" {
		t.Errorf("found.Name = %q, want Orders", found.Name())
	}
}

func TestDictionary_FindDataSourceByAlias_CaseInsensitive(t *testing.T) {
	d := data.NewDictionary()
	ds := data.NewBusinessObjectDataSource("Sales", nil)
	d.AddDataSource(ds)
	if d.FindDataSourceByAlias("sales") == nil {
		t.Error("FindDataSourceByAlias should be case-insensitive")
	}
}

func TestDictionary_FindDataSourceByName(t *testing.T) {
	d := data.NewDictionary()
	ds := data.NewBusinessObjectDataSource("Products", nil)
	d.AddDataSource(ds)
	found := d.FindDataSourceByName("products")
	if found == nil {
		t.Error("FindDataSourceByName returned nil")
	}
}

func TestDictionary_FindDataSourceByAlias_NotFound(t *testing.T) {
	d := data.NewDictionary()
	if d.FindDataSourceByAlias("Nope") != nil {
		t.Error("FindDataSourceByAlias should return nil for unknown alias")
	}
}

func TestDictionary_RemoveDataSource(t *testing.T) {
	d := data.NewDictionary()
	ds := data.NewBusinessObjectDataSource("DS", nil)
	d.AddDataSource(ds)
	d.RemoveDataSource(ds)
	if len(d.DataSources()) != 0 {
		t.Errorf("DataSources len = %d after remove, want 0", len(d.DataSources()))
	}
}

func TestDictionary_RegisterData(t *testing.T) {
	d := data.NewDictionary()
	type row struct{ X int }
	ds := d.RegisterData([]row{{1}, {2}}, "Rows")
	if ds == nil {
		t.Fatal("RegisterData returned nil")
	}
	if len(d.DataSources()) != 1 {
		t.Errorf("DataSources len = %d, want 1", len(d.DataSources()))
	}
	_ = ds.Init()
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
}

// -----------------------------------------------------------------------
// Relations
// -----------------------------------------------------------------------

func TestDictionary_AddAndRemoveRelation(t *testing.T) {
	d := data.NewDictionary()
	r := &data.Relation{Name: "Orders_Products"}
	d.AddRelation(r)
	if len(d.Relations()) != 1 {
		t.Fatalf("Relations len = %d, want 1", len(d.Relations()))
	}
	d.RemoveRelation(r)
	if len(d.Relations()) != 0 {
		t.Errorf("Relations len = %d after remove, want 0", len(d.Relations()))
	}
}

// -----------------------------------------------------------------------
// Parameters
// -----------------------------------------------------------------------

func TestDictionary_AddAndFindParameter(t *testing.T) {
	d := data.NewDictionary()
	p := &data.Parameter{Name: "StartDate", Value: "2024-01-01"}
	d.AddParameter(p)

	found := d.FindParameter("StartDate")
	if found == nil {
		t.Fatal("FindParameter returned nil")
	}
	if found.Value != "2024-01-01" {
		t.Errorf("found.Value = %v, want 2024-01-01", found.Value)
	}
}

func TestDictionary_FindParameter_NestedDot(t *testing.T) {
	d := data.NewDictionary()
	parent := &data.Parameter{Name: "Filter"}
	child := &data.Parameter{Name: "Region", Value: "EMEA"}
	parent.AddParameter(child)
	d.AddParameter(parent)

	found := d.FindParameter("Filter.Region")
	if found == nil {
		t.Fatal("FindParameter nested returned nil")
	}
	if found.Value != "EMEA" {
		t.Errorf("found.Value = %v, want EMEA", found.Value)
	}
}

func TestDictionary_FindParameter_NotFound(t *testing.T) {
	d := data.NewDictionary()
	if d.FindParameter("X") != nil {
		t.Error("FindParameter should return nil for unknown name")
	}
}

func TestDictionary_RemoveParameter(t *testing.T) {
	d := data.NewDictionary()
	p := &data.Parameter{Name: "P"}
	d.AddParameter(p)
	d.RemoveParameter(p)
	if len(d.Parameters()) != 0 {
		t.Errorf("Parameters len = %d after remove, want 0", len(d.Parameters()))
	}
}

// -----------------------------------------------------------------------
// System variables
// -----------------------------------------------------------------------

func TestDictionary_SetSystemVariable_New(t *testing.T) {
	d := data.NewDictionary()
	d.SetSystemVariable("PageNumber", 1)
	svs := d.SystemVariables()
	if len(svs) != 1 {
		t.Fatalf("SystemVariables len = %d, want 1", len(svs))
	}
	if svs[0].Value != 1 {
		t.Errorf("PageNumber value = %v, want 1", svs[0].Value)
	}
}

func TestDictionary_SetSystemVariable_Update(t *testing.T) {
	d := data.NewDictionary()
	d.SetSystemVariable("Page", 1)
	d.SetSystemVariable("page", 5) // case-insensitive update
	if len(d.SystemVariables()) != 1 {
		t.Fatalf("should not duplicate; len = %d", len(d.SystemVariables()))
	}
	if d.SystemVariables()[0].Value != 5 {
		t.Errorf("Page value = %v, want 5", d.SystemVariables()[0].Value)
	}
}

func TestDictionary_AddSystemVariable(t *testing.T) {
	d := data.NewDictionary()
	sv := &data.Parameter{Name: "TotalPages", Value: 0}
	d.AddSystemVariable(sv)
	if len(d.SystemVariables()) != 1 {
		t.Errorf("SystemVariables len = %d, want 1", len(d.SystemVariables()))
	}
}

// -----------------------------------------------------------------------
// Totals
// -----------------------------------------------------------------------

func TestDictionary_AddAndFindTotal(t *testing.T) {
	d := data.NewDictionary()
	tot := &data.Total{Name: "GrandTotal", Value: 0.0}
	d.AddTotal(tot)

	found := d.FindTotal("GrandTotal")
	if found == nil {
		t.Fatal("FindTotal returned nil")
	}
	if found != tot {
		t.Error("FindTotal should return same pointer")
	}
}

func TestDictionary_FindTotal_CaseInsensitive(t *testing.T) {
	d := data.NewDictionary()
	d.AddTotal(&data.Total{Name: "Sum1", Value: 100})
	if d.FindTotal("sum1") == nil {
		t.Error("FindTotal should be case-insensitive")
	}
}

func TestDictionary_FindTotal_NotFound(t *testing.T) {
	d := data.NewDictionary()
	if d.FindTotal("X") != nil {
		t.Error("FindTotal should return nil for unknown total")
	}
}

func TestDictionary_RemoveTotal(t *testing.T) {
	d := data.NewDictionary()
	tot := &data.Total{Name: "T"}
	d.AddTotal(tot)
	d.RemoveTotal(tot)
	if len(d.Totals()) != 0 {
		t.Errorf("Totals len = %d after remove, want 0", len(d.Totals()))
	}
}

// -----------------------------------------------------------------------
// DictionaryLookup interface satisfaction
// -----------------------------------------------------------------------

func TestDictionary_ImplementsDictionaryLookup(t *testing.T) {
	d := data.NewDictionary()
	var _ data.DictionaryLookup = d // compile-time check
	_ = d
}
