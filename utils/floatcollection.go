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
