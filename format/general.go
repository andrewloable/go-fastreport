package format

import (
	"fmt"
	"reflect"
)

// GeneralFormat is a no-op formatter that returns fmt.Sprint(v).
// It matches the C# GeneralFormat which calls value.ToString().
type GeneralFormat struct{}

// NewGeneralFormat returns a GeneralFormat.
func NewGeneralFormat() *GeneralFormat { return &GeneralFormat{} }

// FormatType implements Format.
func (f *GeneralFormat) FormatType() string { return "General" }

// FormatValue implements Format. Returns fmt.Sprint(v), or "" for nil.
// Matches C# GeneralFormat.FormatValue which returns "" for null.
func (f *GeneralFormat) FormatValue(v any) string {
	if v == nil {
		return ""
	}
	// Typed nil pointers (e.g. (*string)(nil)) produce "<nil>" via fmt.Sprint,
	// but C# returns "" for null — so treat them the same way.
	if isNilPointer(v) {
		return ""
	}
	return fmt.Sprint(v)
}

// Clone returns a new GeneralFormat.
// Mirrors C# GeneralFormat.Clone().
func (f *GeneralFormat) Clone() Format {
	return &GeneralFormat{}
}

// Equals reports whether other is also a GeneralFormat.
// Mirrors C# GeneralFormat.Equals().
func (f *GeneralFormat) Equals(other Format) bool {
	_, ok := other.(*GeneralFormat)
	return ok
}

// GetSampleValue returns a representative formatted string for UI preview.
// Mirrors C# GeneralFormat.GetSampleValue() which returns "".
func (f *GeneralFormat) GetSampleValue() string {
	return ""
}

// isNilPointer reports whether v is a typed nil pointer.
func isNilPointer(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Pointer && rv.IsNil()
}
