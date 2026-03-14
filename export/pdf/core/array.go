package core

import (
	"fmt"
	"io"
)

// Array is a PDF array object that holds an ordered list of PDF Objects.
//
// PDF representation:
//
//	[ val1 val2 val3 ]
type Array struct {
	items []Object
}

// NewArray returns an Array pre-populated with the supplied items.
func NewArray(items ...Object) *Array {
	a := &Array{}
	a.items = append(a.items, items...)
	return a
}

// Type implements Object.
func (a *Array) Type() ObjectType { return TypeArray }

// Add appends item to the array and returns the receiver for chaining.
func (a *Array) Add(item Object) *Array {
	a.items = append(a.items, item)
	return a
}

// Len returns the number of items in the array.
func (a *Array) Len() int { return len(a.items) }

// WriteTo writes the PDF array representation to w.
func (a *Array) WriteTo(w io.Writer) (int64, error) {
	cw := &countWriter{w: w}
	if _, err := fmt.Fprint(cw, "[ "); err != nil {
		return cw.n, err
	}
	for _, item := range a.items {
		if _, err := item.WriteTo(cw); err != nil {
			return cw.n, err
		}
		if _, err := fmt.Fprint(cw, " "); err != nil {
			return cw.n, err
		}
	}
	_, err := fmt.Fprint(cw, "]")
	return cw.n, err
}
