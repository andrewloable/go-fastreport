package report

import (
	"iter"
	"sort"
)

// ObjectCollection is a typed collection of Base objects.
// It is the Go equivalent of FastReport.ObjectCollection and is used throughout
// the report model to hold child objects.
type ObjectCollection struct {
	items []Base
}

// NewObjectCollection returns an empty ObjectCollection.
func NewObjectCollection() *ObjectCollection {
	return &ObjectCollection{}
}

// Add appends obj to the end of the collection.
// If obj is nil the call is a no-op, matching C# FRCollectionBase.Add nil guard
// (FRCollectionBase.cs).
func (c *ObjectCollection) Add(obj Base) {
	if obj == nil {
		return
	}
	c.items = append(c.items, obj)
}

// AddRange appends all elements of objs to the collection.
func (c *ObjectCollection) AddRange(objs []Base) {
	c.items = append(c.items, objs...)
}

// Insert inserts obj at the given index, shifting subsequent elements right.
// If obj is nil the call is a no-op, matching the C# FRCollectionBase.Insert nil
// guard (FRCollectionBase.cs).
// Panics if index is out of range [0, Len()].
func (c *ObjectCollection) Insert(index int, obj Base) {
	if obj == nil {
		return
	}
	c.items = append(c.items, nil)
	copy(c.items[index+1:], c.items[index:])
	c.items[index] = obj
}

// Remove removes the first occurrence of obj from the collection.
// Returns true if an element was removed, false if obj was not found.
func (c *ObjectCollection) Remove(obj Base) bool {
	for i, item := range c.items {
		if item == obj {
			c.removeAt(i)
			return true
		}
	}
	return false
}

// RemoveAt removes the element at index from the collection.
// Panics if index is out of range.
func (c *ObjectCollection) RemoveAt(index int) {
	c.removeAt(index)
}

func (c *ObjectCollection) removeAt(index int) {
	c.items = append(c.items[:index], c.items[index+1:]...)
}

// Get returns the element at index.
// Panics if index is out of range.
func (c *ObjectCollection) Get(index int) Base {
	return c.items[index]
}

// Len returns the number of elements in the collection.
func (c *ObjectCollection) Len() int {
	return len(c.items)
}

// Clear removes all elements from the collection.
func (c *ObjectCollection) Clear() {
	c.items = c.items[:0]
}

// Contains reports whether obj is present in the collection.
func (c *ObjectCollection) Contains(obj Base) bool {
	return c.IndexOf(obj) >= 0
}

// IndexOf returns the index of the first occurrence of obj, or -1 if not found.
func (c *ObjectCollection) IndexOf(obj Base) int {
	for i, item := range c.items {
		if item == obj {
			return i
		}
	}
	return -1
}

// All returns an iterator over index-element pairs (Go 1.23 iter.Seq2).
func (c *ObjectCollection) All() iter.Seq2[int, Base] {
	return func(yield func(int, Base) bool) {
		for i, item := range c.items {
			if !yield(i, item) {
				return
			}
		}
	}
}

// Slice returns a shallow copy of the underlying slice.
func (c *ObjectCollection) Slice() []Base {
	out := make([]Base, len(c.items))
	copy(out, c.items)
	return out
}

// FindByName searches the collection for an object whose Name() equals name.
// Returns the first match, or nil if not found.
func (c *ObjectCollection) FindByName(name string) Base {
	for _, item := range c.items {
		if item.Name() == name {
			return item
		}
	}
	return nil
}

// SortByTop returns a new slice of objects sorted by their Top coordinate.
// It is the Go equivalent of ReportComponentCollection.SortByTop().
// Objects that do not expose Bounds are sorted as having Top = 0.
func (c *ObjectCollection) SortByTop() []Base {
	out := c.Slice()
	sort.SliceStable(out, func(i, j int) bool {
		var topI, topJ float32
		type bounded interface{ Bounds() Rect }
		if ci, ok := out[i].(bounded); ok {
			topI = ci.Bounds().Top
		}
		if cj, ok := out[j].(bounded); ok {
			topJ = cj.Bounds().Top
		}
		return topI < topJ
	})
	return out
}

// AddRangeCollection appends all elements of other to the collection.
// It is the ObjectCollection-overload of AddRange, matching C#
// FRCollectionBase.AddRange(ObjectCollection) (FRCollectionBase.cs).
func (c *ObjectCollection) AddRangeCollection(other *ObjectCollection) {
	if other == nil {
		return
	}
	c.items = append(c.items, other.items...)
}

// Equals reports whether c and other contain the same elements in the same
// order.  It is the Go equivalent of FRCollectionBase.Equals (FRCollectionBase.cs).
func (c *ObjectCollection) Equals(other *ObjectCollection) bool {
	if other == nil {
		return false
	}
	if len(c.items) != len(other.items) {
		return false
	}
	for i := range c.items {
		if c.items[i] != other.items[i] {
			return false
		}
	}
	return true
}

// CopyTo replaces the contents of dst with a shallow copy of c.
// It is the Go equivalent of FRCollectionBase.CopyTo (FRCollectionBase.cs).
func (c *ObjectCollection) CopyTo(dst *ObjectCollection) {
	if dst == nil {
		return
	}
	dst.items = dst.items[:0]
	dst.items = append(dst.items, c.items...)
}

// TypedCollection is a generic typed collection whose element type T must
// implement Base. It embeds ObjectCollection and adds type-safe Add and Get
// methods that shadow the untyped equivalents.
type TypedCollection[T Base] struct {
	ObjectCollection
}

// Add appends obj to the collection.
func (c *TypedCollection[T]) Add(obj T) {
	c.ObjectCollection.Add(obj)
}

// Get returns the element at index as T.
// Panics if index is out of range or if the stored element is not T.
func (c *TypedCollection[T]) Get(index int) T {
	return c.ObjectCollection.Get(index).(T)
}
