package data_test

// column_collection_extra_test.go — tests for ColumnCollection methods
// that were missing coverage: CreateUniqueName, CreateUniqueAlias, Sort.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// ── ColumnCollection.CreateUniqueName ─────────────────────────────────────────

func TestColumnCollection_CreateUniqueName_NoConflict(t *testing.T) {
	cc := data.NewColumnCollection()
	name := cc.CreateUniqueName("OrderID")
	if name != "OrderID" {
		t.Errorf("CreateUniqueName with no conflict = %q, want OrderID", name)
	}
}

func TestColumnCollection_CreateUniqueName_OneConflict(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("Amount"))
	name := cc.CreateUniqueName("Amount")
	if name == "Amount" {
		t.Error("CreateUniqueName should generate a unique name when base exists")
	}
	if name != "Amount1" {
		t.Errorf("CreateUniqueName = %q, want Amount1", name)
	}
}

func TestColumnCollection_CreateUniqueName_MultipleConflicts(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("Field"))
	cc.Add(data.NewDataColumn("Field1"))
	cc.Add(data.NewDataColumn("Field2"))
	name := cc.CreateUniqueName("Field")
	if name != "Field3" {
		t.Errorf("CreateUniqueName = %q, want Field3", name)
	}
}

func TestColumnCollection_CreateUniqueName_EmptyCollection(t *testing.T) {
	cc := data.NewColumnCollection()
	name := cc.CreateUniqueName("X")
	if name != "X" {
		t.Errorf("CreateUniqueName on empty collection = %q, want X", name)
	}
}

// ── ColumnCollection.CreateUniqueAlias ────────────────────────────────────────

func TestColumnCollection_CreateUniqueAlias_NoConflict(t *testing.T) {
	cc := data.NewColumnCollection()
	alias := cc.CreateUniqueAlias("Total Amount")
	if alias != "Total Amount" {
		t.Errorf("CreateUniqueAlias with no conflict = %q, want 'Total Amount'", alias)
	}
}

func TestColumnCollection_CreateUniqueAlias_OneConflict(t *testing.T) {
	cc := data.NewColumnCollection()
	col := data.NewDataColumn("amt")
	col.Alias = "Amount"
	cc.Add(col)

	alias := cc.CreateUniqueAlias("Amount")
	if alias == "Amount" {
		t.Error("CreateUniqueAlias should generate a unique alias when base exists")
	}
	if alias != "Amount1" {
		t.Errorf("CreateUniqueAlias = %q, want Amount1", alias)
	}
}

func TestColumnCollection_CreateUniqueAlias_MultipleConflicts(t *testing.T) {
	cc := data.NewColumnCollection()
	for _, alias := range []string{"Label", "Label1", "Label2"} {
		col := data.NewDataColumn(alias)
		col.Alias = alias
		cc.Add(col)
	}
	alias := cc.CreateUniqueAlias("Label")
	if alias != "Label3" {
		t.Errorf("CreateUniqueAlias = %q, want Label3", alias)
	}
}

// ── ColumnCollection.Sort ─────────────────────────────────────────────────────

func TestColumnCollection_Sort_Ascending(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("Zebra"))
	cc.Add(data.NewDataColumn("Apple"))
	cc.Add(data.NewDataColumn("Mango"))

	cc.Sort()

	want := []string{"Apple", "Mango", "Zebra"}
	for i, name := range want {
		if cc.Get(i).Name != name {
			t.Errorf("Sort index %d = %q, want %q", i, cc.Get(i).Name, name)
		}
	}
}

func TestColumnCollection_Sort_Empty(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Sort() // must not panic
	if cc.Len() != 0 {
		t.Errorf("Len = %d after sort of empty, want 0", cc.Len())
	}
}

func TestColumnCollection_Sort_SingleElement(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("Only"))
	cc.Sort()
	if cc.Get(0).Name != "Only" {
		t.Errorf("Sort single element = %q, want Only", cc.Get(0).Name)
	}
}

func TestColumnCollection_Sort_AlreadySorted(t *testing.T) {
	cc := data.NewColumnCollection()
	cc.Add(data.NewDataColumn("Alpha"))
	cc.Add(data.NewDataColumn("Beta"))
	cc.Add(data.NewDataColumn("Gamma"))
	cc.Sort()
	want := []string{"Alpha", "Beta", "Gamma"}
	for i, name := range want {
		if cc.Get(i).Name != name {
			t.Errorf("Sort already-sorted index %d = %q, want %q", i, cc.Get(i).Name, name)
		}
	}
}

func TestColumnCollection_Sort_Stable(t *testing.T) {
	// Two columns with same name — stable sort should preserve insertion order.
	cc := data.NewColumnCollection()
	a := data.NewDataColumn("dup")
	b := data.NewDataColumn("dup")
	cc.Add(a)
	cc.Add(b)
	cc.Sort()
	if cc.Get(0) != a {
		t.Error("Sort should be stable for equal names")
	}
}
