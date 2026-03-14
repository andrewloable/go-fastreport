package data_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

func TestNewDataColumn(t *testing.T) {
	col := data.NewDataColumn("OrderID")
	if col.Name != "OrderID" {
		t.Errorf("Name = %q, want OrderID", col.Name)
	}
	if col.Alias != "OrderID" {
		t.Errorf("Alias default should equal Name, got %q", col.Alias)
	}
	if !col.Enabled {
		t.Error("Enabled should default to true")
	}
	if col.Calculated {
		t.Error("Calculated should default to false")
	}
}

func TestDataColumnHasColumns(t *testing.T) {
	col := data.NewDataColumn("Customer")
	if col.HasColumns() {
		t.Error("new column should have no nested columns")
	}
	col.Columns().Add(data.NewDataColumn("Name"))
	if !col.HasColumns() {
		t.Error("column should have nested columns after add")
	}
}

func TestColumnCollectionAdd(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("A"))
	cc.Add(data.NewDataColumn("B"))
	if cc.Len() != 2 {
		t.Errorf("Len = %d, want 2", cc.Len())
	}
}

func TestColumnCollectionGet(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("X"))
	if cc.Get(0).Name != "X" {
		t.Errorf("Get(0).Name = %q, want X", cc.Get(0).Name)
	}
}

func TestColumnCollectionRemove(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("A"))
	cc.Add(data.NewDataColumn("B"))
	ok := cc.Remove("A")
	if !ok {
		t.Error("Remove should return true for existing column")
	}
	if cc.Len() != 1 {
		t.Errorf("Len = %d, want 1", cc.Len())
	}
	if cc.Get(0).Name != "B" {
		t.Errorf("remaining column = %q, want B", cc.Get(0).Name)
	}
	ok = cc.Remove("nonexistent")
	if ok {
		t.Error("Remove of nonexistent should return false")
	}
}

func TestColumnCollectionClear(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("A"))
	cc.Clear()
	if cc.Len() != 0 {
		t.Errorf("after Clear Len = %d, want 0", cc.Len())
	}
}

func TestColumnCollectionFindByName(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("OrderID"))
	cc.Add(data.NewDataColumn("CustomerName"))

	col := cc.FindByName("CustomerName")
	if col == nil {
		t.Fatal("FindByName returned nil for existing column")
	}
	if col.Name != "CustomerName" {
		t.Errorf("found %q, want CustomerName", col.Name)
	}
	if cc.FindByName("missing") != nil {
		t.Error("FindByName should return nil for missing column")
	}
}

func TestColumnCollectionFindByNameNested(t *testing.T) {
	cc := data.NewColumnCollection()
	parent := data.NewDataColumn("Address")
	parent.Columns().Add(data.NewDataColumn("City"))
	parent.Columns().Add(data.NewDataColumn("ZipCode"))
	cc.Add(parent)

	col := cc.FindByName("City")
	if col == nil {
		t.Fatal("FindByName should find nested column")
	}
	if col.Name != "City" {
		t.Errorf("found %q, want City", col.Name)
	}
}

func TestColumnCollectionFindByAlias(t *testing.T) {
	cc := data.NewColumnCollection()
	col := data.NewDataColumn("cust_name")
	col.Alias = "Customer Name"
	cc.Add(col)

	found := cc.FindByAlias("Customer Name")
	if found == nil {
		t.Fatal("FindByAlias returned nil")
	}
	if found.Name != "cust_name" {
		t.Errorf("found %q, want cust_name", found.Name)
	}
	if cc.FindByAlias("nonexistent") != nil {
		t.Error("FindByAlias should return nil for missing alias")
	}
}

func TestColumnCollectionFindByAliasNested(t *testing.T) {
	cc := data.NewColumnCollection()
	parent := data.NewDataColumn("Location")
	child := data.NewDataColumn("zip")
	child.Alias = "Zip Code"
	parent.Columns().Add(child)
	cc.Add(parent)

	found := cc.FindByAlias("Zip Code")
	if found == nil {
		t.Fatal("FindByAlias should find nested column by alias")
	}
	if found.Name != "zip" {
		t.Errorf("found %q, want zip", found.Name)
	}
}

func TestColumnCollectionAll(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("A"))
	cc.Add(data.NewDataColumn("B"))
	cc.Add(data.NewDataColumn("C"))

	sum := 0
	for i := range cc.All() {
		sum += i
	}
	if sum != 3 { // 0+1+2
		t.Errorf("All indices sum = %d, want 3", sum)
	}
}

func TestColumnCollectionAllEarlyStop(t *testing.T) {
	cc := data.NewColumnCollection()
	for i := 0; i < 5; i++ {
		cc.Add(data.NewDataColumn("col"))
	}
	count := 0
	for range cc.All() {
		count++
		if count == 2 {
			break
		}
	}
	if count != 2 {
		t.Errorf("early stop count = %d, want 2", count)
	}
}

func TestColumnCollectionSlice(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("A"))
	cc.Add(data.NewDataColumn("B"))
	s := cc.Slice()
	if len(s) != 2 {
		t.Fatalf("Slice len = %d, want 2", len(s))
	}
	// Modifying the slice should not affect the collection
	s[0] = data.NewDataColumn("Z")
	if cc.Get(0).Name != "A" {
		t.Error("Slice is not a copy")
	}
}

func TestColumnFormatConstants(t *testing.T) {
	formats := []data.ColumnFormat{
		data.ColumnFormatAuto,
		data.ColumnFormatGeneral,
		data.ColumnFormatNumber,
		data.ColumnFormatCurrency,
		data.ColumnFormatDate,
		data.ColumnFormatTime,
		data.ColumnFormatPercent,
		data.ColumnFormatBoolean,
	}
	seen := make(map[data.ColumnFormat]bool)
	for _, f := range formats {
		if seen[f] {
			t.Errorf("duplicate ColumnFormat value %d", f)
		}
		seen[f] = true
	}
}
