package format

import "fmt"

// BooleanFormat defines how boolean values are formatted and displayed.
// Non-bool values are passed through fmt.Sprint.
type BooleanFormat struct {
	// TrueText is the string returned when the value is true.
	TrueText string
	// FalseText is the string returned when the value is false.
	FalseText string
}

// NewBooleanFormat returns a BooleanFormat with default settings.
func NewBooleanFormat() *BooleanFormat {
	return &BooleanFormat{
		TrueText:  "True",
		FalseText: "False",
	}
}

// FormatType implements Format.
func (f *BooleanFormat) FormatType() string { return "Boolean" }

// FormatValue implements Format. If v is a bool, returns TrueText or
// FalseText. Otherwise returns fmt.Sprint(v). nil and typed-nil pointers
// return "" to match C# BooleanFormat.FormatValue which returns "" for null.
func (f *BooleanFormat) FormatValue(v any) string {
	if v == nil {
		return ""
	}
	// Typed nil pointers (e.g. (*string)(nil)) must return "" — same as C# null.
	if isNilPointer(v) {
		return ""
	}
	if b, ok := v.(bool); ok {
		if b {
			return f.TrueText
		}
		return f.FalseText
	}
	return fmt.Sprint(v)
}

// Clone returns a deep copy of this BooleanFormat.
// Mirrors C# BooleanFormat.Clone().
func (f *BooleanFormat) Clone() Format {
	return &BooleanFormat{
		TrueText:  f.TrueText,
		FalseText: f.FalseText,
	}
}

// Equals reports whether f and other represent the same format configuration.
// Mirrors C# BooleanFormat.Equals().
func (f *BooleanFormat) Equals(other Format) bool {
	o, ok := other.(*BooleanFormat)
	return ok && f.TrueText == o.TrueText && f.FalseText == o.FalseText
}

// GetSampleValue returns a representative formatted string for UI preview.
// Mirrors C# BooleanFormat.GetSampleValue() which calls FormatValue(false).
func (f *BooleanFormat) GetSampleValue() string {
	return f.FormatValue(false)
}
