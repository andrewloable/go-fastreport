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
