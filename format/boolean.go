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
// FalseText. Otherwise returns fmt.Sprint(v). nil returns "".
func (f *BooleanFormat) FormatValue(v any) string {
	if v == nil {
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
