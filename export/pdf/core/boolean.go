package core

import (
	"fmt"
	"io"
)

// Boolean is a PDF boolean object.
//
// PDF representation:  true  or  false
type Boolean struct {
	// Value is the Go boolean value.
	Value bool
}

// NewBoolean returns a Boolean with the given value.
func NewBoolean(v bool) *Boolean { return &Boolean{Value: v} }

// Type implements Object.
func (b *Boolean) Type() ObjectType { return TypeBoolean }

// WriteTo writes "true" or "false" to w.
func (b *Boolean) WriteTo(w io.Writer) (int64, error) {
	cw := &countWriter{w: w}
	s := "false"
	if b.Value {
		s = "true"
	}
	_, err := fmt.Fprint(cw, s)
	return cw.n, err
}
