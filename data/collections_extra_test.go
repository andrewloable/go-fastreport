package data_test

// collections_extra_test.go — additional tests for DataSourceCollection,
// TotalCollection, and TableCollection covering previously-uncovered paths.
// DataConnectionCollection and DataSourceCollection.Remove are already covered
// in data_coverage_test.go, so only new unique tests are added here.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// ── DataSourceCollection.Sort ─────────────────────────────────────────────────

func TestDataSourceCollection_Sort(t *testing.T) {
	c := data.NewDataSourceCollection()

	ds1 := data.NewBaseDataSource("Zebra")
	ds1.SetAlias("Zebra")
	ds2 := data.NewBaseDataSource("Apple")
	ds2.SetAlias("Apple")
	ds3 := data.NewBaseDataSource("Mango")
	ds3.SetAlias("Mango")

	c.Add(ds1)
	c.Add(ds2)
	c.Add(ds3)

	c.Sort()

	// After sort, aliases should be in ascending order.
	want := []string{"Apple", "Mango", "Zebra"}
	for i, alias := range want {
		got := c.Get(i).Alias()
		if got != alias {
			t.Errorf("Sort: index %d: got %q, want %q", i, got, alias)
		}
	}
}

func TestDataSourceCollection_Sort_Empty(t *testing.T) {
	c := data.NewDataSourceCollection()
	// Must not panic.
	c.Sort()
	if c.Count() != 0 {
		t.Errorf("Count = %d, want 0", c.Count())
	}
}

func TestDataSourceCollection_Sort_SingleElement(t *testing.T) {
	c := data.NewDataSourceCollection()
	ds := data.NewBaseDataSource("Only")
	c.Add(ds)
	c.Sort() // must not panic
	if c.Count() != 1 {
		t.Errorf("Count = %d, want 1", c.Count())
	}
}

func TestDataSourceCollection_Sort_AlreadySorted(t *testing.T) {
	c := data.NewDataSourceCollection()
	ds1 := data.NewBaseDataSource("Alpha")
	ds1.SetAlias("Alpha")
	ds2 := data.NewBaseDataSource("Beta")
	ds2.SetAlias("Beta")
	c.Add(ds1)
	c.Add(ds2)

	c.Sort()
	if c.Get(0).Alias() != "Alpha" || c.Get(1).Alias() != "Beta" {
		t.Error("Sort: already-sorted collection should remain in same order")
	}
}

// ── TotalCollection ───────────────────────────────────────────────────────────

func TestTotalCollection_AddGetCount(t *testing.T) {
	c := data.NewTotalCollection()
	if c.Count() != 0 {
		t.Errorf("Count = %d, want 0", c.Count())
	}

	t1 := &data.Total{Name: "Total1"}
	t2 := &data.Total{Name: "Total2"}
	c.Add(t1)
	c.Add(t2)

	if c.Count() != 2 {
		t.Errorf("Count = %d, want 2", c.Count())
	}
	if c.Get(0) != t1 {
		t.Error("Get(0) should return t1")
	}
	if c.Get(1) != t2 {
		t.Error("Get(1) should return t2")
	}
}

func TestTotalCollection_All(t *testing.T) {
	c := data.NewTotalCollection()
	t1 := &data.Total{Name: "A"}
	t2 := &data.Total{Name: "B"}
	c.Add(t1)
	c.Add(t2)

	all := c.All()
	if len(all) != 2 {
		t.Errorf("All len = %d, want 2", len(all))
	}
	if all[0] != t1 || all[1] != t2 {
		t.Error("All returned wrong elements")
	}
}

func TestTotalCollection_Remove(t *testing.T) {
	c := data.NewTotalCollection()
	t1 := &data.Total{Name: "T1"}
	t2 := &data.Total{Name: "T2"}
	c.Add(t1)
	c.Add(t2)

	c.Remove(t1)
	if c.Count() != 1 {
		t.Errorf("Count = %d, want 1 after Remove", c.Count())
	}
	if c.Get(0) != t2 {
		t.Error("After removing t1, t2 should be at index 0")
	}
}

func TestTotalCollection_Remove_NotFound(t *testing.T) {
	c := data.NewTotalCollection()
	t1 := &data.Total{Name: "T1"}
	other := &data.Total{Name: "Other"}
	c.Add(t1)

	c.Remove(other) // must be a no-op
	if c.Count() != 1 {
		t.Errorf("Count = %d, want 1 after removing absent item", c.Count())
	}
}

func TestTotalCollection_FindByName(t *testing.T) {
	c := data.NewTotalCollection()
	t1 := &data.Total{Name: "SumTotal"}
	c.Add(t1)

	found := c.FindByName("SumTotal")
	if found != t1 {
		t.Error("FindByName should return matching total")
	}
	if c.FindByName("missing") != nil {
		t.Error("FindByName should return nil for missing name")
	}
}

func TestTotalCollection_FindByName_CaseInsensitive(t *testing.T) {
	c := data.NewTotalCollection()
	t1 := &data.Total{Name: "GrandTotal"}
	c.Add(t1)

	found := c.FindByName("grandtotal")
	if found != t1 {
		t.Error("FindByName should be case-insensitive")
	}
}

func TestTotalCollection_CreateUniqueName(t *testing.T) {
	c := data.NewTotalCollection()
	c.Add(&data.Total{Name: "Total"})
	c.Add(&data.Total{Name: "Total1"})

	name := c.CreateUniqueName("Total")
	if name == "Total" || name == "Total1" {
		t.Errorf("CreateUniqueName should avoid existing names, got %q", name)
	}
	if name != "Total2" {
		t.Errorf("CreateUniqueName = %q, want Total2", name)
	}
}

func TestTotalCollection_CreateUniqueName_NoConflict(t *testing.T) {
	c := data.NewTotalCollection()
	name := c.CreateUniqueName("SumTotal")
	if name != "SumTotal" {
		t.Errorf("CreateUniqueName with no conflict = %q, want SumTotal", name)
	}
}

func TestTotalCollection_GetValue(t *testing.T) {
	c := data.NewTotalCollection()
	tot := &data.Total{Name: "Sales"}
	tot.Value = 42.0
	c.Add(tot)

	val, err := c.GetValue("Sales")
	if err != nil {
		t.Fatalf("GetValue: %v", err)
	}
	if val != 42.0 {
		t.Errorf("GetValue = %v, want 42.0", val)
	}
}

func TestTotalCollection_GetValue_NotFound(t *testing.T) {
	c := data.NewTotalCollection()
	_, err := c.GetValue("missing")
	if err == nil {
		t.Error("GetValue should return error for missing total")
	}
}

// ── TableCollection ───────────────────────────────────────────────────────────

func TestTableCollection_AddGetCount(t *testing.T) {
	c := data.NewTableCollection()
	if c.Count() != 0 {
		t.Errorf("Count = %d, want 0", c.Count())
	}

	tds1 := data.NewTableDataSource("Orders")
	tds2 := data.NewTableDataSource("Customers")
	c.Add(tds1)
	c.Add(tds2)

	if c.Count() != 2 {
		t.Errorf("Count = %d, want 2", c.Count())
	}
	if c.Get(0) != tds1 {
		t.Error("Get(0) should return tds1")
	}
	if c.Get(1) != tds2 {
		t.Error("Get(1) should return tds2")
	}
}

func TestTableCollection_All(t *testing.T) {
	c := data.NewTableCollection()
	tds1 := data.NewTableDataSource("T1")
	tds2 := data.NewTableDataSource("T2")
	c.Add(tds1)
	c.Add(tds2)

	all := c.All()
	if len(all) != 2 {
		t.Errorf("All len = %d, want 2", len(all))
	}
	if all[0] != tds1 || all[1] != tds2 {
		t.Error("All returned wrong elements")
	}
}

func TestTableCollection_Remove(t *testing.T) {
	c := data.NewTableCollection()
	tds1 := data.NewTableDataSource("T1")
	tds2 := data.NewTableDataSource("T2")
	c.Add(tds1)
	c.Add(tds2)

	c.Remove(tds1)
	if c.Count() != 1 {
		t.Errorf("Count = %d, want 1 after Remove", c.Count())
	}
	if c.Get(0) != tds2 {
		t.Error("After removing tds1, tds2 should be at index 0")
	}
}

func TestTableCollection_Remove_NotFound(t *testing.T) {
	c := data.NewTableCollection()
	tds1 := data.NewTableDataSource("T1")
	other := data.NewTableDataSource("Other")
	c.Add(tds1)

	c.Remove(other) // must be no-op
	if c.Count() != 1 {
		t.Errorf("Count = %d, want 1 after removing absent item", c.Count())
	}
}

func TestTableCollection_Sort(t *testing.T) {
	c := data.NewTableCollection()

	z := data.NewTableDataSource("Zebra")
	z.SetAlias("Zebra")
	a := data.NewTableDataSource("Apple")
	a.SetAlias("Apple")
	m := data.NewTableDataSource("Mango")
	m.SetAlias("Mango")

	c.Add(z)
	c.Add(a)
	c.Add(m)

	c.Sort()

	want := []string{"Apple", "Mango", "Zebra"}
	for i, alias := range want {
		got := c.Get(i).Alias()
		if got != alias {
			t.Errorf("Sort: index %d: got %q, want %q", i, got, alias)
		}
	}
}

func TestTableCollection_Sort_Empty(t *testing.T) {
	c := data.NewTableCollection()
	c.Sort() // must not panic
}

// ── TotalCollection.Contains / ClearValues ────────────────────────────────────

// TestTotalCollection_Contains verifies that Contains returns true only for
// totals that were actually added by pointer.
// C# ref: FRCollectionBase.Contains used in Total.AddValue (Total.cs:386).
func TestTotalCollection_Contains(t *testing.T) {
	c := data.NewTotalCollection()
	t1 := &data.Total{Name: "T1"}
	t2 := &data.Total{Name: "T2"}
	c.Add(t1)

	if !c.Contains(t1) {
		t.Error("Contains(t1) should be true after Add")
	}
	if c.Contains(t2) {
		t.Error("Contains(t2) should be false before Add")
	}

	c.Add(t2)
	if !c.Contains(t2) {
		t.Error("Contains(t2) should be true after second Add")
	}
}

func TestTotalCollection_Contains_Empty(t *testing.T) {
	c := data.NewTotalCollection()
	tot := &data.Total{Name: "Orphan"}
	if c.Contains(tot) {
		t.Error("Contains should return false for empty collection")
	}
}

// TestTotalCollection_ClearValues verifies that ClearValues resets every
// total's runtime Value to nil.
// C# ref: FastReport.Data.TotalCollection.ClearValues (TotalCollection.cs:79).
func TestTotalCollection_ClearValues(t *testing.T) {
	c := data.NewTotalCollection()
	t1 := &data.Total{Name: "T1"}
	t1.Value = 100.0
	t2 := &data.Total{Name: "T2"}
	t2.Value = "some string"
	c.Add(t1)
	c.Add(t2)

	c.ClearValues()

	if t1.Value != nil {
		t.Errorf("T1.Value after ClearValues = %v, want nil", t1.Value)
	}
	if t2.Value != nil {
		t.Errorf("T2.Value after ClearValues = %v, want nil", t2.Value)
	}
}

func TestTotalCollection_ClearValues_Empty(t *testing.T) {
	c := data.NewTotalCollection()
	// Must not panic on empty collection.
	c.ClearValues()
}
