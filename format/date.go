package format

import (
	"fmt"
	"time"
)


// DateFormat defines how date values are formatted and displayed.
// The Format field accepts either Go time layout strings (e.g. "2006-01-02")
// or C# standard date/time format specifiers (e.g. "d" for short date).
type DateFormat struct {
	// Format is a Go time layout string or a C# standard format specifier.
	// Default: "d" (C# short date → M/d/yyyy, e.g. "8/12/2013")
	Format string
	// UseLocaleSettings is reserved for future locale-aware formatting.
	UseLocaleSettings bool
}

// csharpDateLayouts maps C# standard date/time format specifiers to Go layouts.
// These are the en-US defaults; locale-aware formatting is not yet supported.
var csharpDateLayouts = map[string]string{
	"d": "1/2/2006",                   // short date   → M/d/yyyy
	"D": "Monday, January 2, 2006",    // long date
	"f": "Monday, January 2, 2006 3:04 PM",  // full date/time (short time)
	"F": "Monday, January 2, 2006 3:04:05 PM", // full date/time (long time)
	"g": "1/2/2006 3:04 PM",           // general date/time (short time)
	"G": "1/2/2006 3:04:05 PM",        // general date/time (long time)
	"t": "3:04 PM",                    // short time
	"T": "3:04:05 PM",                 // long time
	"M": "January 2",                  // month/day
	"m": "January 2",
	"Y": "January 2006",               // year/month
	"y": "January 2006",
	"s": "2006-01-02T15:04:05",        // sortable
	"u": "2006-01-02 15:04:05Z",       // universal sortable
	"R": "Mon, 02 Jan 2006 15:04:05 GMT", // RFC1123
	"r": "Mon, 02 Jan 2006 15:04:05 GMT",
	"o": time.RFC3339Nano,             // round-trip
	"O": time.RFC3339Nano,
}

// NewDateFormat returns a DateFormat with default settings.
func NewDateFormat() *DateFormat {
	return &DateFormat{
		Format:            "d",
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
		layout = "d"
	}
	// Translate C# standard format specifiers to Go layout strings.
	if goLayout, ok := csharpDateLayouts[layout]; ok {
		layout = goLayout
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

// Clone returns a deep copy of this DateFormat.
// Mirrors C# DateFormat.Clone().
func (f *DateFormat) Clone() Format {
	return &DateFormat{
		Format:            f.Format,
		UseLocaleSettings: f.UseLocaleSettings,
	}
}

// Equals reports whether f and other represent the same format configuration.
// Mirrors C# DateFormat.Equals().
func (f *DateFormat) Equals(other Format) bool {
	o, ok := other.(*DateFormat)
	return ok && f.Format == o.Format && f.UseLocaleSettings == o.UseLocaleSettings
}

// GetSampleValue returns a representative formatted string for UI preview.
// Mirrors C# DateFormat.GetSampleValue() which uses 2007-11-30 13:30:00.
func (f *DateFormat) GetSampleValue() string {
	sample := time.Date(2007, 11, 30, 13, 30, 0, 0, time.UTC)
	return f.FormatValue(sample)
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
				// Convert to local time to match C# DateTime.Parse() behaviour.
				// C# parses RFC3339 strings into DateTimeKind.Local, converting
				// the value to the system's local timezone. Without this, dates
				// near midnight with a different UTC offset (e.g. "23:00+03:00"
				// on a UTC+8 system) would format as the previous calendar day.
				return parsed.Local(), true
			}
		}
	}
	return time.Time{}, false
}
