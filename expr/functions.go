package expr

import (
	"fmt"
	"math"
	"time"
)

// BuiltinFunctions returns a map of built-in function implementations for
// use in expr-lang/expr evaluation environments.
//
// Available functions:
//
//	IIF(condition, trueVal, falseVal) – conditional expression
//	Format(value, fmt)               – string formatting
//	DateDiff(date1, date2, unit)     – date difference (unit: "days","hours","minutes","seconds")
//	Str(value)                       – convert to string
//	Int(value)                       – convert to int
//	Float(value)                     – convert to float64
//	Len(s)                           – string length
//	Upper(s)                         – uppercase
//	Lower(s)                         – lowercase
//
// Aggregate functions (Sum, Count, Avg, Min, Max) are registered externally
// by the engine because they require access to data source rows.
func BuiltinFunctions() map[string]any {
	return map[string]any{
		"IIF":       iif,
		"Format":    formatValue,
		"DateDiff":  dateDiff,
		"Str":       str,
		"Int":       toInt,
		"Float":     toFloat,
		"Len":       strLen,
		"Upper":     upper,
		"Lower":     lower,
		"Substring": substring,
	}
}

// iif implements a conditional expression: IIF(condition, trueVal, falseVal).
func iif(condition bool, trueVal, falseVal any) any {
	if condition {
		return trueVal
	}
	return falseVal
}

// formatValue formats a value using a format string: Format(value, "0.00").
// The format string follows Go's fmt.Sprintf conventions (e.g. "%v", "%.2f").
// For convenience, if the format string does not start with "%", it is treated
// as a numeric format pattern and "%" is prepended automatically.
func formatValue(value any, format string) string {
	if len(format) == 0 {
		return fmt.Sprintf("%v", value)
	}
	if format[0] != '%' {
		format = "%" + format
	}
	return fmt.Sprintf(format, value)
}

// dateDiff returns the difference between two time.Time values in the given unit.
// Supported units: "days", "hours", "minutes", "seconds".
func dateDiff(date1, date2 time.Time, unit string) (float64, error) {
	diff := date2.Sub(date1)
	switch unit {
	case "days":
		return diff.Hours() / 24, nil
	case "hours":
		return diff.Hours(), nil
	case "minutes":
		return diff.Minutes(), nil
	case "seconds":
		return diff.Seconds(), nil
	default:
		return 0, fmt.Errorf("dateDiff: unknown unit %q (use days, hours, minutes, seconds)", unit)
	}
}

// str converts any value to its string representation.
func str(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// toInt converts a numeric or string value to int.
func toInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	case float32:
		return int(math.Round(float64(v))), nil
	case float64:
		return int(math.Round(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case string:
		var i int
		_, err := fmt.Sscanf(v, "%d", &i)
		if err != nil {
			return 0, fmt.Errorf("Int: cannot convert %q to int: %w", v, err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("Int: unsupported type %T", value)
	}
}

// toFloat converts a numeric or string value to float64.
func toFloat(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	case string:
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		if err != nil {
			return 0, fmt.Errorf("Float: cannot convert %q to float64: %w", v, err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("Float: unsupported type %T", value)
	}
}

// strLen returns the number of runes in the string.
func strLen(s string) int {
	return len([]rune(s))
}

// upper returns the uppercase version of s.
func upper(s string) string {
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
		}
		result = append(result, r)
	}
	return string(result)
}

// lower returns the lowercase version of s.
func lower(s string) string {
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		result = append(result, r)
	}
	return string(result)
}

// substring returns a substring of s starting at startIndex with the given
// length. Mirrors the C# String.Substring(startIndex, length) method.
// Both startIndex and length are rune-based (Unicode-aware).
// If startIndex is out of bounds it returns ""; if length exceeds the
// remaining runes it is clamped to the available length.
func substring(s string, startIndex, length int) string {
	runes := []rune(s)
	if startIndex < 0 || startIndex >= len(runes) {
		return ""
	}
	end := startIndex + length
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[startIndex:end])
}
