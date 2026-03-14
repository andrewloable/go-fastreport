package core

import (
	"fmt"
	"io"
	"strconv"
)

// Numeric is a PDF numeric object.  It can represent either an integer or a
// real (floating-point) number.
//
// PDF representation:
//
//	42        (integer)
//	3.1400    (real, 4 decimal places)
type Numeric struct {
	// Value holds the numeric value for both integer and real cases.
	Value float64
	// IsInt selects integer output (no decimal point) when true.
	IsInt bool
}

// NewInt returns a Numeric that renders as a PDF integer.
func NewInt(v int) *Numeric { return &Numeric{Value: float64(v), IsInt: true} }

// NewFloat returns a Numeric that renders as a PDF real number with 4 decimal
// places of precision.
func NewFloat(v float64) *Numeric { return &Numeric{Value: v} }

// Type implements Object.
func (n *Numeric) Type() ObjectType { return TypeNumeric }

// WriteTo writes the PDF numeric representation to w.
func (n *Numeric) WriteTo(w io.Writer) (int64, error) {
	cw := &countWriter{w: w}
	var s string
	if n.IsInt {
		s = strconv.Itoa(int(n.Value))
	} else {
		s = strconv.FormatFloat(n.Value, 'f', 4, 64)
	}
	_, err := fmt.Fprint(cw, s)
	return cw.n, err
}
