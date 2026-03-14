package format

import (
	"fmt"
	"math"
)

// PercentFormat defines how percent values are formatted and displayed.
// Following .NET convention the input value is a fraction (e.g. 0.25 → 25%).
//
// PositivePattern (0–3):
//
//	0 → n %   1 → n%   2 → %n   3 → % n
//
// NegativePattern (0–11):
//
//	0 → -n %   1 → -n%   2 → -%n   3 → %-n
//	4 → %n-    5 → n-%   6 → n%-   7 → -%n
//	8 → n %-   9 → % n-  10 → % -n  11 → n- %
type PercentFormat struct {
	// UseLocaleSettings, when true, uses the invariant locale defaults.
	UseLocaleSettings bool
	// DecimalDigits is the number of digits after the decimal point.
	DecimalDigits int
	// DecimalSeparator is the character(s) used as the decimal point.
	DecimalSeparator string
	// GroupSeparator is the character(s) used to separate thousands.
	GroupSeparator string
	// PercentSymbol is the symbol to use (default "%").
	PercentSymbol string
	// PositivePattern selects the layout for positive values (0–3).
	PositivePattern int
	// NegativePattern selects the layout for negative values (0–11).
	NegativePattern int
}

// NewPercentFormat returns a PercentFormat with default settings.
func NewPercentFormat() *PercentFormat {
	return &PercentFormat{
		UseLocaleSettings: true,
		DecimalDigits:     2,
		DecimalSeparator:  ".",
		GroupSeparator:    ",",
		PercentSymbol:     "%",
		PositivePattern:   0,
		NegativePattern:   0,
	}
}

// FormatType implements Format.
func (f *PercentFormat) FormatType() string { return "Percent" }

// FormatValue implements Format. The input value is expected to be a fraction
// in [0,1] (e.g. 0.25) which is multiplied by 100 before formatting.
func (f *PercentFormat) FormatValue(v any) string {
	if v == nil {
		return ""
	}
	val, ok := toFloat64(v)
	if !ok {
		return fmt.Sprint(v)
	}

	dec := "."
	grp := ","
	sym := "%"
	if !f.UseLocaleSettings {
		dec = f.DecimalSeparator
		grp = f.GroupSeparator
		sym = f.PercentSymbol
	}

	// Multiply by 100, as per .NET percent-format convention.
	pct := val * 100.0

	neg := pct < 0
	abs := math.Abs(pct)
	n := formatNumber(abs, f.DecimalDigits, dec, grp)

	if !neg {
		return applyPercentPositivePattern(n, sym, f.PositivePattern)
	}
	return applyPercentNegativePattern(n, sym, f.NegativePattern)
}

func applyPercentPositivePattern(n, sym string, pattern int) string {
	switch pattern {
	case 0:
		return n + " " + sym
	case 1:
		return n + sym
	case 2:
		return sym + n
	case 3:
		return sym + " " + n
	default:
		return n + " " + sym
	}
}

func applyPercentNegativePattern(n, sym string, pattern int) string {
	switch pattern {
	case 0:
		return "-" + n + " " + sym
	case 1:
		return "-" + n + sym
	case 2:
		return "-" + sym + n
	case 3:
		return sym + "-" + n
	case 4:
		return sym + n + "-"
	case 5:
		return n + "-" + sym
	case 6:
		return n + sym + "-"
	case 7:
		return "-" + sym + n
	case 8:
		return n + " " + sym + "-"
	case 9:
		return sym + " " + n + "-"
	case 10:
		return sym + " -" + n
	case 11:
		return n + "- " + sym
	default:
		return "-" + n + " " + sym
	}
}
