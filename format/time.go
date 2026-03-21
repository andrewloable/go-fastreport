package format

import (
	"fmt"
	"time"
)

// TimeFormat defines how time-of-day values are formatted and displayed.
// The Format field uses Go time layout strings (e.g. "15:04:05").
// Default is "15:04" matching C#'s short-time "t" pattern.
type TimeFormat struct {
	// Format is a Go time layout string.
	// Default: "15:04"
	Format string
	// UseLocaleSettings is reserved for future locale-aware formatting.
	UseLocaleSettings bool
}

// NewTimeFormat returns a TimeFormat with default settings.
func NewTimeFormat() *TimeFormat {
	return &TimeFormat{
		Format:            "15:04",
		UseLocaleSettings: false,
	}
}

// FormatType implements Format.
func (f *TimeFormat) FormatType() string { return "Time" }

// FormatValue implements Format. Accepts time.Time, time.Duration (converted
// to a time on the zero date), string (parsed via common layouts), or
// anything that can be Sprint-ed.
func (f *TimeFormat) FormatValue(v any) string {
	if v == nil {
		return ""
	}
	if isNilPointer(v) {
		return ""
	}
	layout := f.Format
	if layout == "" {
		layout = "15:04"
	}

	// Handle time.Duration: treat as elapsed time from midnight.
	if d, ok := v.(time.Duration); ok {
		t := time.Time{}.Add(d)
		return t.Format(layout)
	}

	if t, ok := toTime(v); ok {
		return t.Format(layout)
	}
	return fmt.Sprint(v)
}
