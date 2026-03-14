package utils

import "iter"

// Namer is the minimal interface that collection items must implement.
// Any report object that has a Name satisfies this.
type Namer interface {
	Name() string
}

// Collection is a generic ordered collection of report objects.
// It is the Go equivalent of FRCollectionBase.
type Collection[T any] struct {
	items []T
}

// Len returns the number of items in the collection.
func (c *Collection[T]) Len() int {
	return len(c.items)
}

// Get returns the item at index i.
// Panics if i is out of range.
func (c *Collection[T]) Get(i int) T {
	return c.items[i]
}

// Add appends item to the collection.
func (c *Collection[T]) Add(item T) {
	c.items = append(c.items, item)
}

// Insert inserts item at position index, shifting subsequent items right.
func (c *Collection[T]) Insert(index int, item T) {
	c.items = append(c.items, item) // grow by one
	copy(c.items[index+1:], c.items[index:])
	c.items[index] = item
}

// Remove removes the first occurrence of item from the collection using
// the provided equality function eq.
func (c *Collection[T]) Remove(item T, eq func(a, b T) bool) bool {
	for i, v := range c.items {
		if eq(v, item) {
			c.items = append(c.items[:i], c.items[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveAt removes the item at index i.
func (c *Collection[T]) RemoveAt(i int) {
	c.items = append(c.items[:i], c.items[i+1:]...)
}

// IndexOf returns the index of item using eq, or -1 if not found.
func (c *Collection[T]) IndexOf(item T, eq func(a, b T) bool) int {
	for i, v := range c.items {
		if eq(v, item) {
			return i
		}
	}
	return -1
}

// Contains returns true if item is in the collection using eq.
func (c *Collection[T]) Contains(item T, eq func(a, b T) bool) bool {
	return c.IndexOf(item, eq) >= 0
}

// Clear removes all items from the collection.
func (c *Collection[T]) Clear() {
	c.items = c.items[:0]
}

// All returns an iterator over all items (Go 1.23 range-over-func).
func (c *Collection[T]) All() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range c.items {
			if !yield(i, v) {
				return
			}
		}
	}
}

// Slice returns a copy of the underlying slice.
func (c *Collection[T]) Slice() []T {
	result := make([]T, len(c.items))
	copy(result, c.items)
	return result
}

// SetOrder moves item to the specified position in the collection.
// Uses eq to find the item.
func (c *Collection[T]) SetOrder(item T, order int, eq func(a, b T) bool) {
	idx := c.IndexOf(item, eq)
	if idx < 0 || idx == order {
		return
	}
	// Remove from current position
	c.items = append(c.items[:idx], c.items[idx+1:]...)
	// Clamp order to valid range
	if order > len(c.items) {
		order = len(c.items)
	}
	if order < 0 {
		order = 0
	}
	// Insert at new position
	c.items = append(c.items, item) // grow by one
	copy(c.items[order+1:], c.items[order:])
	c.items[order] = item
}
