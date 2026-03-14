package format

// Format is the base interface for all value formatters.
type Format interface {
	// FormatValue formats v to a string.
	FormatValue(v any) string
	// FormatType returns the format type name (for serialization).
	FormatType() string
}
