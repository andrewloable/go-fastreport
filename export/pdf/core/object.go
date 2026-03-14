package core

import "io"

// ObjectType is a string tag identifying the kind of PDF object.
type ObjectType string

// PDF object type constants used for debugging and type assertions.
const (
	TypeIndirect   ObjectType = "indirect"
	TypeDictionary ObjectType = "dictionary"
	TypeArray      ObjectType = "array"
	TypeStream     ObjectType = "stream"
	TypeString     ObjectType = "string"
	TypeName       ObjectType = "name"
	TypeNumeric    ObjectType = "numeric"
	TypeBoolean    ObjectType = "boolean"
	TypeNull       ObjectType = "null"
)

// Object is the base interface implemented by every PDF primitive.
// WriteTo writes the PDF textual representation of the object to w and
// returns the number of bytes written together with any write error.
type Object interface {
	io.WriterTo

	// Type returns a string tag for the concrete type, useful for debugging.
	Type() ObjectType
}
