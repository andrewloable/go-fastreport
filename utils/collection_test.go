package utils_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/utils"
)

func eqInt(a, b int) bool { return a == b }

func TestCollectionAddLen(t *testing.T) {
	c := &utils.Collection[int]{}
	if c.Len() != 0 {
		t.Fatalf("empty collection should have len 0, got %d", c.Len())
	}
	c.Add(1)
	c.Add(2)
	c.Add(3)
	if c.Len() != 3 {
		t.Fatalf("expected len 3, got %d", c.Len())
	}
}

func TestCollectionGet(t *testing.T) {
	c := &utils.Collection[string]{}
	c.Add("a")
	c.Add("b")
	if c.Get(0) != "a" {
		t.Errorf("Get(0) = %v, want a", c.Get(0))
	}
	if c.Get(1) != "b" {
		t.Errorf("Get(1) = %v, want b", c.Get(1))
	}
}

func TestCollectionInsert(t *testing.T) {
	c := &utils.Collection[int]{}
	c.Add(1)
	c.Add(3)
	c.Insert(1, 2)
	if c.Len() != 3 {
		t.Fatalf("expected 3, got %d", c.Len())
	}
	if c.Get(1) != 2 {
		t.Errorf("Insert: Get(1) = %v, want 2", c.Get(1))
	}
	if c.Get(2) != 3 {
		t.Errorf("Insert: Get(2) = %v, want 3", c.Get(2))
	}
}

func TestCollectionRemove(t *testing.T) {
	c := &utils.Collection[int]{}
	c.Add(1)
	c.Add(2)
	c.Add(3)
	ok := c.Remove(2, eqInt)
	if !ok {
		t.Fatal("Remove should return true")
	}
	if c.Len() != 2 {
		t.Fatalf("expected len 2, got %d", c.Len())
	}
	if c.Get(0) != 1 || c.Get(1) != 3 {
		t.Errorf("unexpected items after remove: %v, %v", c.Get(0), c.Get(1))
	}
}

func TestCollectionRemoveMissing(t *testing.T) {
	c := &utils.Collection[int]{}
	c.Add(1)
	ok := c.Remove(99, eqInt)
	if ok {
		t.Error("Remove of missing item should return false")
	}
	if c.Len() != 1 {
		t.Error("collection should be unchanged")
	}
}

func TestCollectionRemoveAt(t *testing.T) {
	c := &utils.Collection[int]{}
	c.Add(10)
	c.Add(20)
	c.Add(30)
	c.RemoveAt(1)
	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}
	if c.Get(0) != 10 || c.Get(1) != 30 {
		t.Errorf("unexpected after RemoveAt: %v, %v", c.Get(0), c.Get(1))
	}
}

func TestCollectionIndexOf(t *testing.T) {
	c := &utils.Collection[int]{}
	c.Add(10)
	c.Add(20)
	c.Add(30)
	if c.IndexOf(20, eqInt) != 1 {
		t.Errorf("IndexOf(20) = %d, want 1", c.IndexOf(20, eqInt))
	}
	if c.IndexOf(99, eqInt) != -1 {
		t.Errorf("IndexOf(99) should be -1")
	}
}

func TestCollectionContains(t *testing.T) {
	c := &utils.Collection[int]{}
	c.Add(5)
	if !c.Contains(5, eqInt) {
		t.Error("Contains(5) should be true")
	}
	if c.Contains(6, eqInt) {
		t.Error("Contains(6) should be false")
	}
}

func TestCollectionClear(t *testing.T) {
	c := &utils.Collection[int]{}
	c.Add(1)
	c.Add(2)
	c.Clear()
	if c.Len() != 0 {
		t.Errorf("after Clear, Len = %d, want 0", c.Len())
	}
}

func TestCollectionAll(t *testing.T) {
	c := &utils.Collection[int]{}
	c.Add(1)
	c.Add(2)
	c.Add(3)
	sum := 0
	for _, v := range c.All() {
		sum += v
	}
	if sum != 6 {
		t.Errorf("All sum = %d, want 6", sum)
	}
}

func TestCollectionAllEarlyStop(t *testing.T) {
	c := &utils.Collection[int]{}
	for i := 0; i < 10; i++ {
		c.Add(i)
	}
	count := 0
	for _, _ = range c.All() {
		count++
		if count == 3 {
			break
		}
	}
	if count != 3 {
		t.Errorf("early stop: count = %d, want 3", count)
	}
}

func TestCollectionSlice(t *testing.T) {
	c := &utils.Collection[int]{}
	c.Add(1)
	c.Add(2)
	s := c.Slice()
	if len(s) != 2 {
		t.Fatalf("Slice len = %d, want 2", len(s))
	}
	// Modifying the slice should not affect the collection
	s[0] = 99
	if c.Get(0) != 1 {
		t.Error("Slice returned a reference, not a copy")
	}
}

func TestCollectionSetOrder(t *testing.T) {
	c := &utils.Collection[int]{}
	c.Add(1)
	c.Add(2)
	c.Add(3)
	c.SetOrder(1, 2, eqInt) // move 1 to end
	if c.Get(2) != 1 {
		t.Errorf("SetOrder: Get(2) = %v, want 1", c.Get(2))
	}
	if c.Get(0) != 2 || c.Get(1) != 3 {
		t.Errorf("SetOrder: unexpected order: %v, %v", c.Get(0), c.Get(1))
	}
}

func TestCollectionSetOrderSamePos(t *testing.T) {
	c := &utils.Collection[int]{}
	c.Add(1)
	c.Add(2)
	c.SetOrder(1, 0, eqInt) // already at position 0
	if c.Get(0) != 1 || c.Get(1) != 2 {
		t.Errorf("SetOrder same position changed order: %v, %v", c.Get(0), c.Get(1))
	}
}
