package core

import (
	"fmt"
	"io"
)

// Null is the PDF null object.
//
// PDF representation:  null
type Null struct{}

// Type implements Object.
func (n *Null) Type() ObjectType { return TypeNull }

// WriteTo writes the literal token "null" to w.
func (n *Null) WriteTo(w io.Writer) (int64, error) {
	cw := &countWriter{w: w}
	_, err := fmt.Fprint(cw, "null")
	return cw.n, err
}
