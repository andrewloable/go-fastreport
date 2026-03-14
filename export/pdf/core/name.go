package core

import (
	"fmt"
	"io"
)

// Name is a PDF name object.  In PDF syntax a name is prefixed with a forward
// slash and non-alphanumeric characters are encoded as "#XX" where XX is the
// uppercase hexadecimal byte value.
//
// PDF representation:  /FlateDecode  or  /my#20name
type Name struct {
	// Value is the name without the leading slash.
	Value string
}

// NewName returns a Name for the given string (without the leading slash).
func NewName(v string) *Name { return &Name{Value: v} }

// Type implements Object.
func (n *Name) Type() ObjectType { return TypeName }

// WriteTo writes the PDF name representation to w.
// An empty name produces no output (matching the original C# behaviour).
func (n *Name) WriteTo(w io.Writer) (int64, error) {
	if n.Value == "" {
		return 0, nil
	}
	cw := &countWriter{w: w}
	if _, err := fmt.Fprint(cw, "/"); err != nil {
		return cw.n, err
	}
	for i := 0; i < len(n.Value); i++ {
		c := n.Value[i]
		if isNameRegular(c) {
			if _, err := fmt.Fprintf(cw, "%c", c); err != nil {
				return cw.n, err
			}
		} else {
			if _, err := fmt.Fprintf(cw, "#%02X", c); err != nil {
				return cw.n, err
			}
		}
	}
	return cw.n, nil
}

// isNameRegular returns true for bytes that do not need #-encoding in a PDF
// name: ASCII letters and digits only (matches the original C# implementation).
func isNameRegular(c byte) bool {
	return ('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		('0' <= c && c <= '9')
}
