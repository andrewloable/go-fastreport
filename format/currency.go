package format

import (
	"fmt"
	"math"
)

// CurrencyFormat defines how currency values are formatted and displayed.
//
// PositivePattern controls the layout for positive amounts:
//
//	0  →  $n
//	1  →  n$
//	2  →  $ n
//	3  →  n $
//
// NegativePattern controls the layout for negative amounts (0–15, matching
// the .NET CurrencyNegativePattern values):
//
//	0  → ($n)   1  → -$n   2  → $-n   3  → $n-
//	4  → (n$)   5  → -n$   6  → n-$   7  → n$-
//	8  → -n $   9  → -$ n  10 → n $-  11 → $ n-
//	12 → $ -n  13 → n- $  14 → ($ n) 15 → (n $)
type CurrencyFormat struct {
	// UseLocaleSettings, when true, uses the invariant locale defaults.
	UseLocaleSettings bool
	// DecimalDigits is the number of digits after the decimal point.
	DecimalDigits int
	// DecimalSeparator is the character(s) used as the decimal point.
	DecimalSeparator string
	// GroupSeparator is the character(s) used to separate thousands.
	GroupSeparator string
	// CurrencySymbol is the currency symbol to prepend/append.
	CurrencySymbol string
	// PositivePattern selects the layout for positive amounts (0–3).
	PositivePattern int
	// NegativePattern selects the layout for negative amounts (0–15).
	NegativePattern int
}

// NewCurrencyFormat returns a CurrencyFormat with settings from the current
// system locale, mirroring C#'s constructor which reads from
// CultureInfo.CurrentCulture.NumberFormat.
func NewCurrencyFormat() *CurrencyFormat {
	loc := currentLocale()
	return &CurrencyFormat{
		UseLocaleSettings: true,
		DecimalDigits:     loc.CurrencyDecimalDigits,
		DecimalSeparator:  loc.CurrencyDecimalSeparator,
		GroupSeparator:    loc.CurrencyGroupSeparator,
		CurrencySymbol:    loc.CurrencySymbol,
		PositivePattern:   loc.CurrencyPositivePattern,
		NegativePattern:   loc.CurrencyNegativePattern,
	}
}

// FormatType implements Format.
func (f *CurrencyFormat) FormatType() string { return "Currency" }

// FormatValue implements Format.
func (f *CurrencyFormat) FormatValue(v any) string {
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

	var dec, grp, sym string
	var posPattern, negPattern int
	if f.UseLocaleSettings {
		// Mirror C# GetNumberFormatInfo(): when UseLocale=true, clone the
		// current culture's NumberFormat and only override DecimalDigits.
		loc := currentLocale()
		dec = loc.CurrencyDecimalSeparator
		grp = loc.CurrencyGroupSeparator
		sym = loc.CurrencySymbol
		posPattern = loc.CurrencyPositivePattern
		negPattern = loc.CurrencyNegativePattern
	} else {
		dec = f.DecimalSeparator
		grp = f.GroupSeparator
		sym = f.CurrencySymbol
		posPattern = f.PositivePattern
		negPattern = f.NegativePattern
	}

	neg := val < 0
	abs := math.Abs(val)
	n := formatNumber(abs, f.DecimalDigits, dec, grp)

	if !neg {
		return applyCurrencyPositivePattern(n, sym, posPattern)
	}
	return applyCurrencyNegativePattern(n, sym, negPattern)
}

func applyCurrencyPositivePattern(n, sym string, pattern int) string {
	switch pattern {
	case 0:
		return sym + n
	case 1:
		return n + sym
	case 2:
		return sym + " " + n
	case 3:
		return n + " " + sym
	default:
		return sym + n
	}
}

// Clone returns a deep copy of this CurrencyFormat.
// Mirrors C# CurrencyFormat.Clone().
func (f *CurrencyFormat) Clone() Format {
	return &CurrencyFormat{
		UseLocaleSettings: f.UseLocaleSettings,
		DecimalDigits:     f.DecimalDigits,
		DecimalSeparator:  f.DecimalSeparator,
		GroupSeparator:    f.GroupSeparator,
		CurrencySymbol:    f.CurrencySymbol,
		PositivePattern:   f.PositivePattern,
		NegativePattern:   f.NegativePattern,
	}
}

// Equals reports whether f and other represent the same format configuration.
// Mirrors C# CurrencyFormat.Equals().
func (f *CurrencyFormat) Equals(other Format) bool {
	o, ok := other.(*CurrencyFormat)
	return ok &&
		f.UseLocaleSettings == o.UseLocaleSettings &&
		f.DecimalDigits == o.DecimalDigits &&
		f.DecimalSeparator == o.DecimalSeparator &&
		f.GroupSeparator == o.GroupSeparator &&
		f.CurrencySymbol == o.CurrencySymbol &&
		f.PositivePattern == o.PositivePattern &&
		f.NegativePattern == o.NegativePattern
}

// GetSampleValue returns a representative formatted string for UI preview.
// Mirrors C# CurrencyFormat.GetSampleValue() which calls FormatValue(-12345).
func (f *CurrencyFormat) GetSampleValue() string {
	return f.FormatValue(-12345)
}

func applyCurrencyNegativePattern(n, sym string, pattern int) string {
	switch pattern {
	case 0:
		return "(" + sym + n + ")"
	case 1:
		return "-" + sym + n
	case 2:
		return sym + "-" + n
	case 3:
		return sym + n + "-"
	case 4:
		return "(" + n + sym + ")"
	case 5:
		return "-" + n + sym
	case 6:
		return n + "-" + sym
	case 7:
		return n + sym + "-"
	case 8:
		return "-" + n + " " + sym
	case 9:
		return "-" + sym + " " + n
	case 10:
		return n + " " + sym + "-"
	case 11:
		return sym + " " + n + "-"
	case 12:
		return sym + " -" + n
	case 13:
		return n + "- " + sym
	case 14:
		return "(" + sym + " " + n + ")"
	case 15:
		return "(" + n + " " + sym + ")"
	default:
		return "-" + sym + n
	}
}
