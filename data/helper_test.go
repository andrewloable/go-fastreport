package data_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// --- stub DictionaryLookup for tests ---

type stubDict struct {
	sources    map[string]data.DataSource
	relations  []*data.Relation
	params     []*data.Parameter
	sysVars    []*data.Parameter
	totals     []*data.Total
}

func (d *stubDict) FindDataSourceByAlias(alias string) data.DataSource {
	return d.sources[alias]
}
func (d *stubDict) Relations() []*data.Relation     { return d.relations }
func (d *stubDict) Parameters() []*data.Parameter   { return d.params }
func (d *stubDict) SystemVariables() []*data.Parameter { return d.sysVars }
func (d *stubDict) Totals() []*data.Total            { return d.totals }

func newStubDict() *stubDict {
	return &stubDict{sources: make(map[string]data.DataSource)}
}

// --- GetDataSource ---

func TestGetDataSource_Found(t *testing.T) {
	dict := newStubDict()
	ds := data.NewBaseDataSource("Orders")
	dict.sources["Orders"] = ds

	got := data.GetDataSource(dict, "Orders")
	if got != ds {
		t.Error("GetDataSource should return the matching datasource")
	}
}

func TestGetDataSource_NotFound(t *testing.T) {
	dict := newStubDict()
	got := data.GetDataSource(dict, "Missing")
	if got != nil {
		t.Error("GetDataSource should return nil for unknown alias")
	}
}

func TestGetDataSource_Empty(t *testing.T) {
	dict := newStubDict()
	got := data.GetDataSource(dict, "")
	if got != nil {
		t.Error("GetDataSource should return nil for empty name")
	}
}

// --- FindRelation ---

func TestFindRelation_Found(t *testing.T) {
	dict := newStubDict()
	parent := data.NewBaseDataSource("Customers")
	child := data.NewBaseDataSource("Orders")
	rel := &data.Relation{
		Name:             "CustomersOrders",
		ParentDataSource: parent,
		ChildDataSource:  child,
	}
	dict.relations = []*data.Relation{rel}

	got := data.FindRelation(dict, parent, child)
	if got != rel {
		t.Error("FindRelation should return the matching relation")
	}
}

func TestFindRelation_NotFound(t *testing.T) {
	dict := newStubDict()
	parent := data.NewBaseDataSource("A")
	child := data.NewBaseDataSource("B")

	got := data.FindRelation(dict, parent, child)
	if got != nil {
		t.Error("FindRelation should return nil when no match")
	}
}

func TestFindRelation_WrongDirection(t *testing.T) {
	dict := newStubDict()
	parent := data.NewBaseDataSource("P")
	child := data.NewBaseDataSource("C")
	dict.relations = []*data.Relation{{ParentDataSource: parent, ChildDataSource: child}}

	// Reversed — should not match.
	got := data.FindRelation(dict, child, parent)
	if got != nil {
		t.Error("FindRelation should not match when parent/child are reversed")
	}
}

// --- GetParameter ---

func TestGetParameter_TopLevel(t *testing.T) {
	dict := newStubDict()
	p := &data.Parameter{Name: "StartDate", Value: "2024-01-01"}
	dict.params = []*data.Parameter{p}

	got := data.GetParameter(dict, "StartDate")
	if got != p {
		t.Error("GetParameter should return top-level parameter")
	}
}

func TestGetParameter_Nested(t *testing.T) {
	dict := newStubDict()
	child := &data.Parameter{Name: "Min", Value: 10}
	parent := &data.Parameter{Name: "Range"}
	parent.AddParameter(child)
	dict.params = []*data.Parameter{parent}

	got := data.GetParameter(dict, "Range.Min")
	if got != child {
		t.Error("GetParameter should resolve nested parameter")
	}
}

func TestGetParameter_SystemVariable(t *testing.T) {
	dict := newStubDict()
	sv := &data.Parameter{Name: "PageNumber", Value: 1}
	dict.sysVars = []*data.Parameter{sv}

	got := data.GetParameter(dict, "PageNumber")
	if got != sv {
		t.Error("GetParameter should fall back to system variables")
	}
}

func TestGetParameter_NotFound(t *testing.T) {
	dict := newStubDict()
	got := data.GetParameter(dict, "Missing")
	if got != nil {
		t.Error("GetParameter should return nil for unknown name")
	}
}

func TestGetParameter_Empty(t *testing.T) {
	dict := newStubDict()
	got := data.GetParameter(dict, "")
	if got != nil {
		t.Error("GetParameter should return nil for empty name")
	}
}

// --- IsValidParameter ---

func TestIsValidParameter_True(t *testing.T) {
	dict := newStubDict()
	dict.params = []*data.Parameter{{Name: "P1"}}
	if !data.IsValidParameter(dict, "P1") {
		t.Error("IsValidParameter should return true for existing parameter")
	}
}

func TestIsValidParameter_False(t *testing.T) {
	dict := newStubDict()
	if data.IsValidParameter(dict, "Missing") {
		t.Error("IsValidParameter should return false for missing parameter")
	}
}

// --- GetTotal ---

func TestGetTotal_Found(t *testing.T) {
	dict := newStubDict()
	dict.totals = []*data.Total{{Name: "GrandTotal", Value: 999.99}}
	v := data.GetTotal(dict, "GrandTotal")
	if v != 999.99 {
		t.Errorf("GetTotal = %v, want 999.99", v)
	}
}

func TestGetTotal_NotFound(t *testing.T) {
	dict := newStubDict()
	v := data.GetTotal(dict, "Missing")
	if v != nil {
		t.Error("GetTotal should return nil for unknown total")
	}
}

// --- IsValidTotal ---

func TestIsValidTotal_True(t *testing.T) {
	dict := newStubDict()
	dict.totals = []*data.Total{{Name: "Sum1"}}
	if !data.IsValidTotal(dict, "Sum1") {
		t.Error("IsValidTotal should return true for existing total")
	}
}

func TestIsValidTotal_False(t *testing.T) {
	dict := newStubDict()
	if data.IsValidTotal(dict, "Missing") {
		t.Error("IsValidTotal should return false for missing total")
	}
}

// --- Parameter helpers ---

func TestFindParameterByName_Found(t *testing.T) {
	params := []*data.Parameter{{Name: "A"}, {Name: "B"}}
	got := data.FindParameterByName(params, "B")
	if got == nil || got.Name != "B" {
		t.Errorf("FindParameterByName: got %v, want B", got)
	}
}

func TestFindParameterByName_NotFound(t *testing.T) {
	params := []*data.Parameter{{Name: "A"}}
	got := data.FindParameterByName(params, "Z")
	if got != nil {
		t.Error("FindParameterByName should return nil for missing name")
	}
}

func TestParameter_AddParameter(t *testing.T) {
	parent := &data.Parameter{Name: "Parent"}
	child := &data.Parameter{Name: "Child"}
	parent.AddParameter(child)
	if len(parent.Parameters()) != 1 {
		t.Errorf("Parameters len = %d, want 1", len(parent.Parameters()))
	}
	if parent.Parameters()[0] != child {
		t.Error("AddParameter did not append child")
	}
}
