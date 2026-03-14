package format

import (
	"fmt"
	"time"
)

// DateFormat defines how date values are formatted and displayed.
// The Format field uses Go time layout strings (e.g. "2006-01-02").
// The zero value Format defaults to "2006-01-02" (ISO 8601 short date),
// matching the intent of C#'s short-date "d" pattern.
type DateFormat struct {
	// Format is a Go time layout string.
	// Default: "2006-01-02"
	Format string
	// UseLocaleSettings is reserved for future locale-aware formatting.
	UseLocaleSettings bool
}

// NewDateFormat returns a DateFormat with default settings.
func NewDateFormat() *DateFormat {
	return &DateFormat{
		Format:            "2006-01-02",
		UseLocaleSettings: false,
	}
}

// FormatType implements Format.
func (f *DateFormat) FormatType() string { return "Date" }

// FormatValue implements Format. Accepts time.Time, string (parsed via
// time.Parse with common layouts), or any type that implements fmt.Stringer.
func (f *DateFormat) FormatValue(v any) string {
	if v == nil {
		return ""
	}
	layout := f.Format
	if layout == "" {
		layout = "2006-01-02"
	}

	if t, ok := toTime(v); ok {
		return t.Format(layout)
	}
	// A typed nil pointer produces no useful output.
	if isNilPointer(v) {
		return ""
	}
	return fmt.Sprint(v)
}

// toTime converts common date-carrying types to time.Time.
func toTime(v any) (time.Time, bool) {
	switch t := v.(type) {
	case time.Time:
		return t, true
	case *time.Time:
		if t == nil {
			return time.Time{}, false
		}
		return *t, true
	case string:
		layouts := []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
			"01/02/2006",
			"02-01-2006",
		}
		for _, l := range layouts {
			if parsed, err := time.Parse(l, t); err == nil {
				return parsed, true
			}
		}
	}
	return time.Time{}, false
}
