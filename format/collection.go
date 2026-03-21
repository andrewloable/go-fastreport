package format

// Collection is an ordered list of Format values attached to a TextObject or
// RichObject. It mirrors the C# FormatCollection class.
//
// The first element is the primary format; subsequent elements can be used for
// multi-expression scenarios (one format per expression placeholder).
type Collection struct {
	items []Format
}

// NewCollection returns an empty Collection.
func NewCollection() *Collection { return &Collection{} }

// Add appends f to the collection. No-op if f is nil.
// Returns the zero-based index of the newly added element, or -1 if f is nil.
func (c *Collection) Add(f Format) int {
	if f == nil {
		return -1
	}
	c.items = append(c.items, f)
	return len(c.items) - 1
}

// Insert inserts f at position index. No-op if f is nil.
func (c *Collection) Insert(index int, f Format) {
	if f == nil {
		return
	}
	c.items = append(c.items, nil)
	copy(c.items[index+1:], c.items[index:])
	c.items[index] = f
}

// Remove removes the first occurrence of f from the collection.
func (c *Collection) Remove(f Format) {
	for i, item := range c.items {
		if item == f {
			c.items = append(c.items[:i], c.items[i+1:]...)
			return
		}
	}
}

// Get returns the element at index i.
func (c *Collection) Get(i int) Format { return c.items[i] }

// Count returns the number of formats in the collection.
func (c *Collection) Count() int { return len(c.items) }

// Clear removes all elements from the collection.
func (c *Collection) Clear() { c.items = c.items[:0] }

// Contains reports whether f is in the collection.
func (c *Collection) Contains(f Format) bool {
	for _, item := range c.items {
		if item == f {
			return true
		}
	}
	return false
}

// IndexOf returns the zero-based index of f, or -1 if not found.
func (c *Collection) IndexOf(f Format) int {
	for i, item := range c.items {
		if item == f {
			return i
		}
	}
	return -1
}

// Assign clears this collection and copies all formats from src via Clone.
// Each Format is deep-copied via the Cloner interface if available, otherwise
// the same value is reused.
func (c *Collection) Assign(src *Collection) {
	c.Clear()
	if src == nil {
		return
	}
	for _, f := range src.items {
		if cl, ok := f.(interface{ Clone() Format }); ok {
			c.items = append(c.items, cl.Clone())
		} else {
			c.items = append(c.items, f)
		}
	}
}

// Primary returns the first format in the collection, or nil if empty.
// This is the format applied when a single Format is expected.
func (c *Collection) Primary() Format {
	if len(c.items) == 0 {
		return nil
	}
	return c.items[0]
}

// FormatValue formats v using the primary format (first element).
// Returns fmt.Sprint(v) when the collection is empty.
func (c *Collection) FormatValue(v any) string {
	if len(c.items) == 0 {
		return (&GeneralFormat{}).FormatValue(v)
	}
	return c.items[0].FormatValue(v)
}

// All returns a copy of the underlying slice for iteration.
func (c *Collection) All() []Format {
	out := make([]Format, len(c.items))
	copy(out, c.items)
	return out
}

// Equals reports whether c and other contain the same number of formats
// and each format at the same index is equal (via the Equaler interface if
// available, otherwise pointer identity).
// Mirrors C# FormatCollection.Equals().
func (c *Collection) Equals(other *Collection) bool {
	if other == nil || len(c.items) != len(other.items) {
		return false
	}
	for i, f := range c.items {
		o := other.items[i]
		if eq, ok := f.(interface{ Equals(Format) bool }); ok {
			if !eq.Equals(o) {
				return false
			}
		} else if f != o {
			return false
		}
	}
	return true
}
