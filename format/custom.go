package format

import "fmt"

// CustomFormat applies a Go fmt.Sprintf format string to the value.
// The Format field is used directly as the format verb, e.g. "%.2f" or "%d".
// Default Format is "%v" (equivalent to C#'s "G" general format).
type CustomFormat struct {
	// Format is a Go fmt.Sprintf format string (e.g. "%.2f", "%d", "%v").
	Format string
}

// NewCustomFormat returns a CustomFormat with default settings.
func NewCustomFormat() *CustomFormat {
	return &CustomFormat{Format: "%v"}
}

// FormatType implements Format.
func (f *CustomFormat) FormatType() string { return "Custom" }

// FormatValue implements Format. Passes v through fmt.Sprintf using Format.
// If Format is empty, falls back to "%v". nil v returns "".
func (f *CustomFormat) FormatValue(v any) string {
	if v == nil {
		return ""
	}
	layout := f.Format
	if layout == "" {
		layout = "%v"
	}
	return fmt.Sprintf(layout, v)
}

// Clone returns a deep copy of this CustomFormat.
// Mirrors C# CustomFormat.Clone().
func (f *CustomFormat) Clone() Format {
	return &CustomFormat{Format: f.Format}
}

// Equals reports whether f and other represent the same format configuration.
// Mirrors C# CustomFormat.Equals().
func (f *CustomFormat) Equals(other Format) bool {
	o, ok := other.(*CustomFormat)
	return ok && f.Format == o.Format
}

// GetSampleValue returns a representative formatted string for UI preview.
// Mirrors C# CustomFormat.GetSampleValue() which returns "".
func (f *CustomFormat) GetSampleValue() string {
	return ""
}
