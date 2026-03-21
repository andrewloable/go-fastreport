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

// --- nil guards ---

func TestObjectCollection_Add_Nil(t *testing.T) {
	c := report.NewObjectCollection()
	c.Add(nil) // must not panic
	if c.Len() != 0 {
		t.Fatalf("Add(nil) should not increase Len, got %d", c.Len())
	}
}

func TestObjectCollection_Insert_Nil(t *testing.T) {
	c := report.NewObjectCollection()
	c.Add(newSO("a"))
	c.Insert(0, nil) // must not panic
	if c.Len() != 1 {
		t.Fatalf("Insert(0, nil) should not increase Len, got %d", c.Len())
	}
}

// --- AddRangeCollection ---

func TestObjectCollection_AddRangeCollection_Basic(t *testing.T) {
	src := report.NewObjectCollection()
	a, b := newSO("a"), newSO("b")
	src.Add(a)
	src.Add(b)

	dst := report.NewObjectCollection()
	dst.AddRangeCollection(src)
	if dst.Len() != 2 {
		t.Fatalf("AddRangeCollection: Len() = %d, want 2", dst.Len())
	}
	if dst.Get(0) != a || dst.Get(1) != b {
		t.Error("AddRangeCollection: element mismatch")
	}
}

func TestObjectCollection_AddRangeCollection_Nil(t *testing.T) {
	dst := report.NewObjectCollection()
	dst.AddRangeCollection(nil) // must not panic
	if dst.Len() != 0 {
		t.Fatalf("AddRangeCollection(nil) should leave collection empty, got %d", dst.Len())
	}
}

func TestObjectCollection_AddRangeCollection_Appends(t *testing.T) {
	first := report.NewObjectCollection()
	first.Add(newSO("x"))

	second := report.NewObjectCollection()
	second.Add(newSO("y"))
	second.Add(newSO("z"))

	first.AddRangeCollection(second)
	if first.Len() != 3 {
		t.Fatalf("after AddRangeCollection: Len() = %d, want 3", first.Len())
	}
}

// --- Equals ---

func TestObjectCollection_Equals_Equal(t *testing.T) {
	a, b := newSO("a"), newSO("b")
	c1 := report.NewObjectCollection()
	c1.Add(a)
	c1.Add(b)

	c2 := report.NewObjectCollection()
	c2.Add(a)
	c2.Add(b)

	if !c1.Equals(c2) {
		t.Error("Equals should return true for collections with same elements in same order")
	}
}

func TestObjectCollection_Equals_DifferentOrder(t *testing.T) {
	a, b := newSO("a"), newSO("b")
	c1 := report.NewObjectCollection()
	c1.Add(a)
	c1.Add(b)

	c2 := report.NewObjectCollection()
	c2.Add(b)
	c2.Add(a)

	if c1.Equals(c2) {
		t.Error("Equals should return false for different element order")
	}
}

func TestObjectCollection_Equals_DifferentLength(t *testing.T) {
	c1 := report.NewObjectCollection()
	c1.Add(newSO("a"))

	c2 := report.NewObjectCollection()
	c2.Add(newSO("a"))
	c2.Add(newSO("b"))

	if c1.Equals(c2) {
		t.Error("Equals should return false for different lengths")
	}
}

func TestObjectCollection_Equals_BothEmpty(t *testing.T) {
	c1 := report.NewObjectCollection()
	c2 := report.NewObjectCollection()
	if !c1.Equals(c2) {
		t.Error("Equals should return true for two empty collections")
	}
}

func TestObjectCollection_Equals_Nil(t *testing.T) {
	c := report.NewObjectCollection()
	if c.Equals(nil) {
		t.Error("Equals(nil) should return false")
	}
}

// --- CopyTo ---

func TestObjectCollection_CopyTo_Basic(t *testing.T) {
	a, b := newSO("a"), newSO("b")
	src := report.NewObjectCollection()
	src.Add(a)
	src.Add(b)

	dst := report.NewObjectCollection()
	dst.Add(newSO("old")) // will be overwritten
	src.CopyTo(dst)

	if dst.Len() != 2 {
		t.Fatalf("CopyTo: dst.Len() = %d, want 2", dst.Len())
	}
	if dst.Get(0) != a || dst.Get(1) != b {
		t.Error("CopyTo: element mismatch")
	}
}

func TestObjectCollection_CopyTo_EmptySource(t *testing.T) {
	src := report.NewObjectCollection()
	dst := report.NewObjectCollection()
	dst.Add(newSO("x"))
	src.CopyTo(dst)
	if dst.Len() != 0 {
		t.Fatalf("CopyTo from empty source: dst.Len() = %d, want 0", dst.Len())
	}
}

func TestObjectCollection_CopyTo_NilDst(t *testing.T) {
	src := report.NewObjectCollection()
	src.Add(newSO("a"))
	src.CopyTo(nil) // must not panic
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

// --- SortByTop ---

// sortBoundedObj wraps ComponentBase so we can construct objects with a
// known Top value for SortByTop tests.
type sortBoundedObj struct {
	*report.ComponentBase
	name string
}

func (b *sortBoundedObj) Name() string               { return b.name }
func (b *sortBoundedObj) SetName(n string)           { b.name = n }
func (b *sortBoundedObj) BaseName() string           { return "Bounded" }
func (b *sortBoundedObj) Parent() report.Parent      { return nil }
func (b *sortBoundedObj) SetParent(_ report.Parent)  {}
func (b *sortBoundedObj) Serialize(_ report.Writer) error   { return nil }
func (b *sortBoundedObj) Deserialize(_ report.Reader) error { return nil }

func newBounded(name string, top float32) *sortBoundedObj {
	cb := report.NewComponentBase()
	cb.SetTop(top)
	return &sortBoundedObj{ComponentBase: cb, name: name}
}

func TestObjectCollection_SortByTop_BasicOrder(t *testing.T) {
	c := report.NewObjectCollection()
	a := newBounded("a", 30)
	b := newBounded("b", 10)
	cr := newBounded("c", 20)
	c.Add(a)
	c.Add(b)
	c.Add(cr)

	sorted := c.SortByTop()
	if len(sorted) != 3 {
		t.Fatalf("SortByTop len = %d, want 3", len(sorted))
	}
	// Expected order: b(10), c(20), a(30).
	if sorted[0] != b {
		t.Errorf("sorted[0] = %v, want b (top=10)", sorted[0])
	}
	if sorted[1] != cr {
		t.Errorf("sorted[1] = %v, want c (top=20)", sorted[1])
	}
	if sorted[2] != a {
		t.Errorf("sorted[2] = %v, want a (top=30)", sorted[2])
	}
}

func TestObjectCollection_SortByTop_AlreadySorted(t *testing.T) {
	c := report.NewObjectCollection()
	a := newBounded("a", 5)
	b := newBounded("b", 10)
	c.Add(a)
	c.Add(b)

	sorted := c.SortByTop()
	if sorted[0] != a || sorted[1] != b {
		t.Error("SortByTop: already-sorted collection should remain in same order")
	}
}

func TestObjectCollection_SortByTop_EqualTop_StableOrder(t *testing.T) {
	c := report.NewObjectCollection()
	a := newBounded("a", 10)
	b := newBounded("b", 10)
	cr := newBounded("c", 10)
	c.Add(a)
	c.Add(b)
	c.Add(cr)

	sorted := c.SortByTop()
	// All have the same Top; stable sort should preserve insertion order.
	if sorted[0] != a || sorted[1] != b || sorted[2] != cr {
		t.Error("SortByTop with equal Top values should preserve insertion order (stable sort)")
	}
}

func TestObjectCollection_SortByTop_Empty(t *testing.T) {
	c := report.NewObjectCollection()
	sorted := c.SortByTop()
	if len(sorted) != 0 {
		t.Errorf("SortByTop on empty collection returned %d elements", len(sorted))
	}
}

func TestObjectCollection_SortByTop_SingleElement(t *testing.T) {
	c := report.NewObjectCollection()
	a := newBounded("a", 42)
	c.Add(a)

	sorted := c.SortByTop()
	if len(sorted) != 1 || sorted[0] != a {
		t.Error("SortByTop with single element should return that element")
	}
}

func TestObjectCollection_SortByTop_DoesNotMutateOriginal(t *testing.T) {
	c := report.NewObjectCollection()
	a := newBounded("a", 30)
	b := newBounded("b", 10)
	c.Add(a)
	c.Add(b)

	_ = c.SortByTop()
	// The original collection should still be in insertion order.
	if c.Get(0) != a {
		t.Error("SortByTop should not mutate the original collection")
	}
	if c.Get(1) != b {
		t.Error("SortByTop should not mutate the original collection")
	}
}

func TestObjectCollection_SortByTop_NoBoundsObjects(t *testing.T) {
	// simpleObject does not implement Bounds(), so Top defaults to 0.
	c := report.NewObjectCollection()
	a := newSO("a")
	b := newSO("b")
	c.Add(a)
	c.Add(b)

	sorted := c.SortByTop()
	// Both have effective Top=0; stable sort preserves insertion order.
	if len(sorted) != 2 {
		t.Fatalf("SortByTop len = %d, want 2", len(sorted))
	}
	if sorted[0] != a || sorted[1] != b {
		t.Error("SortByTop: objects without Bounds() should sort as Top=0 (stable)")
	}
}

func TestObjectCollection_SortByTop_MixedBoundedAndUnbounded(t *testing.T) {
	// Mix bounded (Top=5) and unbounded (Top=0).
	c := report.NewObjectCollection()
	high := newBounded("high", 5)
	plain := newSO("plain") // no Bounds, effective Top=0
	c.Add(high)
	c.Add(plain)

	sorted := c.SortByTop()
	if len(sorted) != 2 {
		t.Fatalf("SortByTop len = %d, want 2", len(sorted))
	}
	// plain has Top=0, high has Top=5 → plain first.
	if sorted[0] != plain {
		t.Errorf("sorted[0] should be plain (Top=0), got %v", sorted[0])
	}
	if sorted[1] != high {
		t.Errorf("sorted[1] should be high (Top=5), got %v", sorted[1])
	}
}
