package core

import (
	"fmt"
	"io"
)

// TypeRef is the ObjectType tag for an indirect-reference object.
const TypeRef ObjectType = "ref"

// Ref is a PDF indirect-reference token, e.g. "5 0 R".
// Use NewRef(obj) to create one from an IndirectObject.
type Ref struct {
	// Value is the complete reference string, e.g. "5 0 R".
	Value string
}

// NewRef creates a Ref pointing at the given indirect object.
func NewRef(obj *IndirectObject) *Ref {
	return &Ref{Value: obj.Reference()}
}

// Type implements Object.
func (r *Ref) Type() ObjectType { return TypeRef }

// WriteTo writes the raw reference string (e.g. "5 0 R") to w.
func (r *Ref) WriteTo(w io.Writer) (int64, error) {
	n, err := fmt.Fprint(w, r.Value)
	return int64(n), err
}
