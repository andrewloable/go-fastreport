package data_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// --- stub DictionaryLookup for tests ---

type stubDict struct {
	sources   map[string]data.DataSource
	relations []*data.Relation
	params    []*data.Parameter
	sysVars   []*data.Parameter
	totals    []*data.Total
}

func (d *stubDict) FindDataSourceByAlias(alias string) data.DataSource {
	return d.sources[alias]
}
func (d *stubDict) FindDataSourceByName(name string) data.DataSource {
	for _, ds := range d.sources {
		if ds.Name() == name {
			return ds
		}
	}
	return nil
}
func (d *stubDict) Relations() []*data.Relation        { return d.relations }
func (d *stubDict) Parameters() []*data.Parameter      { return d.params }
func (d *stubDict) SystemVariables() []*data.Parameter { return d.sysVars }
func (d *stubDict) Totals() []*data.Total              { return d.totals }

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

func TestGetDataSource_ByNameFallback(t *testing.T) {
	dict := newStubDict()
	ds := data.NewBaseDataSource("Orders")
	ds.SetAlias("SalesOrders")
	dict.sources["SalesOrders"] = ds

	got := data.GetDataSource(dict, "Orders")
	if got != ds {
		t.Error("GetDataSource should fall back to datasource name")
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

// --- Column helpers ---

func TestGetColumn_FlatColumn(t *testing.T) {
	dict := newStubDict()
	ds := data.NewBaseDataSource("Orders")
	ds.AddColumn(data.Column{Name: "CustomerID", Alias: "CustomerID", DataType: "int"})
	dict.sources["Orders"] = ds

	got := data.GetColumn(dict, "Orders.CustomerID")
	if got == nil {
		t.Fatal("GetColumn should resolve a flat datasource column")
	}
	if got.Name != "CustomerID" || got.DataType != "int" {
		t.Fatalf("GetColumn = %+v", got)
	}
}

func TestGetColumn_RelationTraversal(t *testing.T) {
	dict := newStubDict()
	customers := data.NewBaseDataSource("Customers")
	customers.AddColumn(data.Column{Name: "Name", Alias: "Name", DataType: "string"})
	orders := data.NewBaseDataSource("Orders")
	dict.sources["Customers"] = customers
	dict.sources["Orders"] = orders
	dict.relations = []*data.Relation{{
		Alias:            "Customers",
		ParentDataSource: customers,
		ChildDataSource:  orders,
	}}

	got := data.GetColumn(dict, "Orders.Customers.Name")
	if got == nil {
		t.Fatal("GetColumn should traverse relations to parent datasource columns")
	}
	if got.Name != "Name" {
		t.Errorf("GetColumn relation traversal name = %q, want Name", got.Name)
	}
}

func TestIsValidColumn_AndType(t *testing.T) {
	dict := newStubDict()
	ds := data.NewBaseDataSource("Orders")
	ds.AddColumn(data.Column{Name: "Amount", Alias: "Amount", DataType: "decimal"})
	dict.sources["Orders"] = ds

	if !data.IsValidColumn(dict, "Orders.Amount") {
		t.Fatal("IsValidColumn should return true for an existing column")
	}
	if got := data.GetColumnType(dict, "Orders.Amount"); got != "decimal" {
		t.Fatalf("GetColumnType = %q, want decimal", got)
	}
	if data.IsValidColumn(dict, "Orders.Missing") {
		t.Fatal("IsValidColumn should return false for a missing column")
	}
}

func TestIsSimpleColumn(t *testing.T) {
	dict := newStubDict()
	ds := data.NewBaseDataSource("Orders")
	ds.AddColumn(data.Column{Name: "Amount", Alias: "Amount", DataType: "decimal"})
	dict.sources["Orders"] = ds

	if !data.IsSimpleColumn(dict, "Orders.Amount") {
		t.Fatal("IsSimpleColumn should be true for direct datasource columns")
	}
	if data.IsSimpleColumn(dict, "Orders.Customers.Amount") {
		t.Fatal("IsSimpleColumn should be false for multi-hop column paths")
	}
}

func TestCreateParameter(t *testing.T) {
	dict := data.NewDictionary()
	got := data.CreateParameter(dict, "Filters.Range.Min")
	if got == nil {
		t.Fatal("CreateParameter returned nil")
	}
	if got.Name != "Min" {
		t.Fatalf("CreateParameter leaf = %q, want Min", got.Name)
	}
	root := dict.FindParameter("Filters.Range.Min")
	if root != got {
		t.Fatal("CreateParameter should create a resolvable nested parameter chain")
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
