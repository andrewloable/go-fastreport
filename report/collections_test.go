package report_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// newSO creates a simpleObject with the given name for use in collection tests.
func newSO(name string) *simpleObject {
	return &simpleObject{name: name}
}

// --- ObjectCollection ---

func TestNewObjectCollection_Empty(t *testing.T) {
	c := report.NewObjectCollection()
	if c.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", c.Len())
	}
}

func TestObjectCollection_Add(t *testing.T) {
	c := report.NewObjectCollection()
	a := newSO("a")
	c.Add(a)
	if c.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", c.Len())
	}
	if c.Get(0) != a {
		t.Error("Get(0) should return the added element")
	}
}

func TestObjectCollection_AddRange(t *testing.T) {
	c := report.NewObjectCollection()
	a, b := newSO("a"), newSO("b")
	c.AddRange([]report.Base{a, b})
	if c.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", c.Len())
	}
	if c.Get(0) != a || c.Get(1) != b {
		t.Error("AddRange order mismatch")
	}
}

func TestObjectCollection_AddRange_Empty(t *testing.T) {
	c := report.NewObjectCollection()
	c.AddRange(nil)
	if c.Len() != 0 {
		t.Fatalf("Len() after AddRange(nil) = %d, want 0", c.Len())
	}
}

func TestObjectCollection_Insert_Beginning(t *testing.T) {
	c := report.NewObjectCollection()
	a, b := newSO("a"), newSO("b")
	c.Add(a)
	c.Insert(0, b)
	if c.Get(0) != b {
		t.Error("Insert at 0: first element should be b")
	}
	if c.Get(1) != a {
		t.Error("Insert at 0: second element should be a")
	}
}

func TestObjectCollection_Insert_Middle(t *testing.T) {
	c := report.NewObjectCollection()
	a, b, mid := newSO("a"), newSO("b"), newSO("mid")
	c.Add(a)
	c.Add(b)
	c.Insert(1, mid)
	if c.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", c.Len())
	}
	if c.Get(1) != mid {
		t.Error("Insert at 1: element at index 1 should be mid")
	}
	if c.Get(2) != b {
		t.Error("Insert at 1: element at index 2 should be b")
	}
}

func TestObjectCollection_Insert_End(t *testing.T) {
	c := report.NewObjectCollection()
	a := newSO("a")
	c.Add(a)
	b := newSO("b")
	c.Insert(1, b)
	if c.Get(1) != b {
		t.Error("Insert at end: element at index 1 should be b")
	}
}

func TestObjectCollection_Remove_Found(t *testing.T) {
	c := report.NewObjectCollection()
	a, b := newSO("a"), newSO("b")
	c.Add(a)
	c.Add(b)
	removed := c.Remove(a)
	if !removed {
		t.Error("Remove should return true when element is found")
	}
	if c.Len() != 1 {
		t.Fatalf("Len() after Remove = %d, want 1", c.Len())
	}
	if c.Get(0) != b {
		t.Error("After removing a, first element should be b")
	}
}

func TestObjectCollection_Remove_NotFound(t *testing.T) {
	c := report.NewObjectCollection()
	a := newSO("a")
	c.Add(a)
	notIn := newSO("x")
	removed := c.Remove(notIn)
	if removed {
		t.Error("Remove should return false when element is not found")
	}
	if c.Len() != 1 {
		t.Fatalf("Len() should remain 1 after failed Remove")
	}
}

func TestObjectCollection_Remove_Empty(t *testing.T) {
	c := report.NewObjectCollection()
	removed := c.Remove(newSO("x"))
	if removed {
		t.Error("Remove on empty collection should return false")
	}
}

func TestObjectCollection_RemoveAt(t *testing.T) {
	c := report.NewObjectCollection()
	a, b, cr := newSO("a"), newSO("b"), newSO("c")
	c.Add(a)
	c.Add(b)
	c.Add(cr)
	c.RemoveAt(1) // remove b
	if c.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", c.Len())
	}
	if c.Get(0) != a || c.Get(1) != cr {
		t.Error("RemoveAt(1) left wrong elements")
	}
}

func TestObjectCollection_Clear(t *testing.T) {
	c := report.NewObjectCollection()
	c.Add(newSO("a"))
	c.Add(newSO("b"))
	c.Clear()
	if c.Len() != 0 {
		t.Fatalf("Len() after Clear = %d, want 0", c.Len())
	}
}

func TestObjectCollection_Contains_True(t *testing.T) {
	c := report.NewObjectCollection()
	a := newSO("a")
	c.Add(a)
	if !c.Contains(a) {
		t.Error("Contains should return true for existing element")
	}
}

func TestObjectCollection_Contains_False(t *testing.T) {
	c := report.NewObjectCollection()
	c.Add(newSO("a"))
	if c.Contains(newSO("x")) {
		t.Error("Contains should return false for absent element")
	}
}

func TestObjectCollection_IndexOf_Found(t *testing.T) {
	c := report.NewObjectCollection()
	a, b := newSO("a"), newSO("b")
	c.Add(a)
	c.Add(b)
	if c.IndexOf(a) != 0 {
		t.Errorf("IndexOf(a) = %d, want 0", c.IndexOf(a))
	}
	if c.IndexOf(b) != 1 {
		t.Errorf("IndexOf(b) = %d, want 1", c.IndexOf(b))
	}
}

func TestObjectCollection_IndexOf_NotFound(t *testing.T) {
	c := report.NewObjectCollection()
	c.Add(newSO("a"))
	if c.IndexOf(newSO("x")) != -1 {
		t.Error("IndexOf should return -1 for absent element")
	}
}

func TestObjectCollection_All(t *testing.T) {
	c := report.NewObjectCollection()
	a, b := newSO("a"), newSO("b")
	c.Add(a)
	c.Add(b)

	got := make(map[int]report.Base)
	for i, obj := range c.All() {
		got[i] = obj
	}
	if len(got) != 2 {
		t.Fatalf("All() yielded %d pairs, want 2", len(got))
	}
	if got[0] != a {
		t.Error("All(): index 0 mismatch")
	}
	if got[1] != b {
		t.Error("All(): index 1 mismatch")
	}
}

func TestObjectCollection_All_EarlyStop(t *testing.T) {
	c := report.NewObjectCollection()
	c.Add(newSO("a"))
	c.Add(newSO("b"))
	c.Add(newSO("c"))

	count := 0
	for range c.All() {
		count++
		break // stop after first element
	}
	if count != 1 {
		t.Errorf("early break: iterated %d times, want 1", count)
	}
}

func TestObjectCollection_All_Empty(t *testing.T) {
	c := report.NewObjectCollection()
	count := 0
	for range c.All() {
		count++
	}
	if count != 0 {
		t.Errorf("All() on empty collection yielded %d items", count)
	}
}

func TestObjectCollection_Slice(t *testing.T) {
	c := report.NewObjectCollection()
	a, b := newSO("a"), newSO("b")
	c.Add(a)
	c.Add(b)

	s := c.Slice()
	if len(s) != 2 {
		t.Fatalf("Slice() len = %d, want 2", len(s))
	}
	if s[0] != a || s[1] != b {
		t.Error("Slice() element mismatch")
	}
	// Verify it is a copy: mutating the slice does not affect the collection.
	s[0] = newSO("mutated")
	if c.Get(0) != a {
		t.Error("Slice() should return a copy, not a reference to internal slice")
	}
}

func TestObjectCollection_Slice_Empty(t *testing.T) {
	c := report.NewObjectCollection()
	s := c.Slice()
	if len(s) != 0 {
		t.Errorf("Slice() on empty collection returned len %d", len(s))
	}
}

func TestObjectCollection_FindByName_Found(t *testing.T) {
	c := report.NewObjectCollection()
	a := newSO("Alpha")
	b := newSO("Beta")
	c.Add(a)
	c.Add(b)

	found := c.FindByName("Beta")
	if found != b {
		t.Error("FindByName should return the element with matching name")
	}
}

func TestObjectCollection_FindByName_NotFound(t *testing.T) {
	c := report.NewObjectCollection()
	c.Add(newSO("Alpha"))

	found := c.FindByName("Missing")
	if found != nil {
		t.Error("FindByName should return nil when name not found")
	}
}

func TestObjectCollection_FindByName_Empty(t *testing.T) {
	c := report.NewObjectCollection()
	found := c.FindByName("X")
	if found != nil {
		t.Error("FindByName on empty collection should return nil")
	}
}

func TestObjectCollection_FindByName_FirstMatch(t *testing.T) {
	c := report.NewObjectCollection()
	a := newSO("dup")
	b := newSO("dup")
	c.Add(a)
	c.Add(b)
	// Should return first occurrence.
	found := c.FindByName("dup")
	if found != a {
		t.Error("FindByName should return first match")
	}
}

// --- TypedCollection ---

func TestTypedCollection_AddAndGet(t *testing.T) {
	c := &report.TypedCollection[*simpleObject]{}
	a := newSO("a")
	c.Add(a)
	if c.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", c.Len())
	}
	got := c.Get(0)
	if got != a {
		t.Error("TypedCollection.Get(0) should return the added element")
	}
}

func TestTypedCollection_MultipleElements(t *testing.T) {
	c := &report.TypedCollection[*simpleObject]{}
	a, b, cr := newSO("a"), newSO("b"), newSO("c")
	c.Add(a)
	c.Add(b)
	c.Add(cr)
	if c.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", c.Len())
	}
	if c.Get(2) != cr {
		t.Error("Get(2) should return c")
	}
}

func TestTypedCollection_InheritedMethods(t *testing.T) {
	c := &report.TypedCollection[*simpleObject]{}
	a := newSO("a")
	b := newSO("b")
	c.Add(a)
	c.Add(b)

	// Test inherited methods from ObjectCollection.
	if !c.Contains(a) {
		t.Error("Contains should find a")
	}
	if c.IndexOf(b) != 1 {
		t.Errorf("IndexOf(b) = %d, want 1", c.IndexOf(b))
	}
	found := c.FindByName("b")
	if found != b {
		t.Error("FindByName should find b")
	}
	removed := c.Remove(a)
	if !removed {
		t.Error("Remove should succeed")
	}
	if c.Len() != 1 {
		t.Fatalf("Len() after Remove = %d, want 1", c.Len())
	}
}

func TestTypedCollection_All(t *testing.T) {
	c := &report.TypedCollection[*simpleObject]{}
	a, b := newSO("a"), newSO("b")
	c.Add(a)
	c.Add(b)

	var collected []report.Base
	for _, obj := range c.All() {
		collected = append(collected, obj)
	}
	if len(collected) != 2 {
		t.Fatalf("All() yielded %d items, want 2", len(collected))
	}
}

func TestTypedCollection_Clear(t *testing.T) {
	c := &report.TypedCollection[*simpleObject]{}
	c.Add(newSO("x"))
	c.Clear()
	if c.Len() != 0 {
		t.Fatalf("Len() after Clear = %d, want 0", c.Len())
	}
}

// Compile-time check: TypedCollection embeds ObjectCollection properly.
var _ interface{ Len() int } = (*report.TypedCollection[*simpleObject])(nil)
