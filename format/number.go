package format

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// NumberFormat defines how numeric values are formatted and displayed.
//
// NegativePattern controls how negative numbers are rendered:
//
//	0  →  (n)
//	1  →  -n
//	2  →  - n
//	3  →  n-
//	4  →  n -
type NumberFormat struct {
	// UseLocaleSettings, when true, uses the host locale's decimal/group
	// separators (always "." / "," on Go's invariant locale).
	UseLocaleSettings bool
	// DecimalDigits is the number of digits after the decimal point.
	DecimalDigits int
	// DecimalSeparator is the character(s) used as the decimal point.
	DecimalSeparator string
	// GroupSeparator is the character(s) used to separate thousands.
	GroupSeparator string
	// NegativePattern selects the layout for negative numbers (0–4).
	NegativePattern int
}

// NewNumberFormat returns a NumberFormat with default settings.
func NewNumberFormat() *NumberFormat {
	return &NumberFormat{
		UseLocaleSettings: true,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		NegativePattern:   1, // -n
	}
}

// FormatType implements Format.
func (f *NumberFormat) FormatType() string { return "Number" }

// FormatValue implements Format.
func (f *NumberFormat) FormatValue(v any) string {
	if v == nil {
		return ""
	}
	if isNilPointer(v) {
		return ""
	}
	val, ok := toFloat64(v)
	if !ok {
		return fmt.Sprint(v)
	}

	dec := "."
	grp := ","
	if !f.UseLocaleSettings {
		dec = f.DecimalSeparator
		grp = f.GroupSeparator
	}

	neg := val < 0
	abs := math.Abs(val)
	formatted := formatNumber(abs, f.DecimalDigits, dec, grp)

	if !neg {
		return formatted
	}
	return applyNegativePattern(formatted, f.NegativePattern)
}

// formatNumber formats an absolute float64 with the given decimal digits,
// decimal separator, and group separator.
func formatNumber(abs float64, decDigits int, dec, grp string) string {
	// Format with the required decimal places using Go's strconv.
	s := strconv.FormatFloat(abs, 'f', decDigits, 64)

	// Split into integer and fractional parts.
	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	var fracPart string
	if len(parts) == 2 {
		fracPart = parts[1]
	}

	// Insert group separators every 3 digits from the right.
	intPart = insertGroupSeparator(intPart, grp)

	if decDigits > 0 {
		return intPart + dec + fracPart
	}
	return intPart
}

// insertGroupSeparator inserts grp every 3 characters from the right.
func insertGroupSeparator(s, grp string) string {
	if grp == "" {
		return s
	}
	n := len(s)
	if n <= 3 {
		return s
	}
	var b strings.Builder
	mod := n % 3
	for i, ch := range s {
		if i > 0 && (i-mod)%3 == 0 && mod != 0 {
			b.WriteString(grp)
		} else if i > 0 && mod == 0 && i%3 == 0 {
			b.WriteString(grp)
		}
		b.WriteRune(ch)
	}
	return b.String()
}

// applyNegativePattern wraps a formatted absolute number according to the
// C#-compatible NegativePattern values 0–4.
func applyNegativePattern(n string, pattern int) string {
	switch pattern {
	case 0:
		return "(" + n + ")"
	case 1:
		return "-" + n
	case 2:
		return "- " + n
	case 3:
		return n + "-"
	case 4:
		return n + " -"
	default:
		return "-" + n
	}
}

// Clone returns a deep copy of this NumberFormat.
// Mirrors C# NumberFormat.Clone().
func (f *NumberFormat) Clone() Format {
	return &NumberFormat{
		UseLocaleSettings: f.UseLocaleSettings,
		DecimalDigits:     f.DecimalDigits,
		DecimalSeparator:  f.DecimalSeparator,
		GroupSeparator:    f.GroupSeparator,
		NegativePattern:   f.NegativePattern,
	}
}

// Equals reports whether f and other represent the same format configuration.
// Mirrors C# NumberFormat.Equals().
func (f *NumberFormat) Equals(other Format) bool {
	o, ok := other.(*NumberFormat)
	return ok &&
		f.UseLocaleSettings == o.UseLocaleSettings &&
		f.DecimalDigits == o.DecimalDigits &&
		f.DecimalSeparator == o.DecimalSeparator &&
		f.GroupSeparator == o.GroupSeparator &&
		f.NegativePattern == o.NegativePattern
}

// GetSampleValue returns a representative formatted string for UI preview.
// Mirrors C# NumberFormat.GetSampleValue() which calls FormatValue(-12345.678).
func (f *NumberFormat) GetSampleValue() string {
	return f.FormatValue(-12345.678)
}

// toFloat64 attempts to convert common numeric types to float64.
func toFloat64(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err == nil {
			return f, true
		}
	}
	return 0, false
}
