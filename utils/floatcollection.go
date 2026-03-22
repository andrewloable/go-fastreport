package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// FloatCollection is a serializable slice of float32 values.
// It is the Go equivalent of FastReport.Utils.FloatCollection.
//
// The string format is a comma-separated list of decimal values, e.g. "2,4,2,4".
// Leading/trailing whitespace around values is trimmed during parsing.
type FloatCollection []float32

// String returns the comma-separated string representation.
func (fc FloatCollection) String() string {
	if len(fc) == 0 {
		return ""
	}
	parts := make([]string, len(fc))
	for i, v := range fc {
		parts[i] = strconv.FormatFloat(float64(v), 'f', -1, 32)
	}
	return strings.Join(parts, ",")
}

// ParseFloatCollection parses a comma-separated string into a FloatCollection.
// Empty string returns an empty collection without error.
func ParseFloatCollection(s string) (FloatCollection, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return FloatCollection{}, nil
	}
	parts := strings.Split(s, ",")
	fc := make(FloatCollection, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseFloat(p, 32)
		if err != nil {
			return nil, fmt.Errorf("FloatCollection: parse %q: %w", p, err)
		}
		fc = append(fc, float32(v))
	}
	return fc, nil
}

// MustParseFloatCollection parses s and panics on error (use in tests/constants).
func MustParseFloatCollection(s string) FloatCollection {
	fc, err := ParseFloatCollection(s)
	if err != nil {
		panic(err)
	}
	return fc
}

// Add appends a value to the collection.
func (fc *FloatCollection) Add(v float32) { *fc = append(*fc, v) }

// Clear empties the collection.
func (fc *FloatCollection) Clear() { *fc = (*fc)[:0] }

// Len returns the number of elements.
func (fc FloatCollection) Len() int { return len(fc) }

// Get returns the element at index i.
func (fc FloatCollection) Get(i int) float32 { return fc[i] }

// Insert inserts value at the given index, shifting later elements right.
// Mirrors C# FloatCollection.Insert (FloatCollection.cs line 51-53).
func (fc *FloatCollection) Insert(index int, value float32) {
	*fc = append(*fc, 0)
	copy((*fc)[index+1:], (*fc)[index:])
	(*fc)[index] = value
}

// IndexOf returns the zero-based index of the first element within 0.01 of
// value, or -1 if not found.
// Mirrors C# FloatCollection.IndexOf (FloatCollection.cs line 73-81).
func (fc FloatCollection) IndexOf(value float32) int {
	for i, v := range fc {
		if v-value < 0.01 && value-v < 0.01 {
			return i
		}
	}
	return -1
}

// Contains returns true when value is in the collection (within 0.01).
// Mirrors C# FloatCollection.Contains (FloatCollection.cs line 88-90).
func (fc FloatCollection) Contains(value float32) bool {
	return fc.IndexOf(value) != -1
}

// Remove removes the first occurrence of value (within 0.01) from the collection.
// Mirrors C# FloatCollection.Remove (FloatCollection.cs line 60-64).
func (fc *FloatCollection) Remove(value float32) {
	i := fc.IndexOf(value)
	if i >= 0 {
		*fc = append((*fc)[:i], (*fc)[i+1:]...)
	}
}

// AddRange appends all values from values to the collection.
// Mirrors C# FloatCollection.AddRange (FloatCollection.cs line 28-33).
func (fc *FloatCollection) AddRange(values []float32) {
	*fc = append(*fc, values...)
}

// Assign replaces the collection contents with a copy of src.
// Mirrors C# FloatCollection.Assign (FloatCollection.cs line 97-103).
func (fc *FloatCollection) Assign(src FloatCollection) {
	*fc = make(FloatCollection, len(src))
	copy(*fc, src)
}

// RemoveAt removes the element at the given index.
// Mirrors C# FloatCollection.RemoveAt (inherited from CollectionBase).
func (fc *FloatCollection) RemoveAt(index int) {
	*fc = append((*fc)[:index], (*fc)[index+1:]...)
}
