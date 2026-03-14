// Package functions implements built-in report functions for go-fastreport.
// These correspond to FastReport.Functions.StdFunctions in the .NET library.
package functions

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"
)

// ── Math / comparison ─────────────────────────────────────────────────────────

// MaxInt returns the larger of two ints.
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt returns the smaller of two ints.
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxFloat returns the larger of two float64 values.
func MaxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// MinFloat returns the smaller of two float64 values.
func MinFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Abs returns the absolute value of v.
func Abs(v float64) float64 { return math.Abs(v) }

// Round rounds v to the nearest integer (half away from zero).
func Round(v float64) float64 { return math.Round(v) }

// RoundTo rounds v to decimalPlaces decimal places.
func RoundTo(v float64, decimalPlaces int) float64 {
	factor := math.Pow(10, float64(decimalPlaces))
	return math.Round(v*factor) / factor
}

// Ceiling returns the smallest integer ≥ v.
func Ceiling(v float64) float64 { return math.Ceil(v) }

// Floor returns the largest integer ≤ v.
func Floor(v float64) float64 { return math.Floor(v) }

// ── String functions ──────────────────────────────────────────────────────────

// Length returns the number of runes in s (equivalent to C# String.Length).
func Length(s string) int { return len([]rune(s)) }

// LowerCase returns s in all lowercase letters.
func LowerCase(s string) string { return strings.ToLower(s) }

// UpperCase returns s in all uppercase letters.
func UpperCase(s string) string { return strings.ToUpper(s) }

// TitleCase returns s with the first letter of each word capitalised.
func TitleCase(s string) string {
	runes := []rune(s)
	inWord := false
	for i, r := range runes {
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			runes[i] = unicode.ToUpper(r)
			inWord = true
		}
	}
	return string(runes)
}

// Trim removes leading and trailing whitespace from s.
func Trim(s string) string { return strings.TrimSpace(s) }

// PadLeft pads s on the left with spaces to reach totalWidth runes.
// If s is already at least totalWidth runes long, s is returned unchanged.
func PadLeft(s string, totalWidth int) string {
	return PadLeftChar(s, totalWidth, ' ')
}

// PadLeftChar pads s on the left with paddingChar to reach totalWidth runes.
func PadLeftChar(s string, totalWidth int, paddingChar rune) string {
	n := len([]rune(s))
	if n >= totalWidth {
		return s
	}
	return strings.Repeat(string(paddingChar), totalWidth-n) + s
}

// PadRight pads s on the right with spaces to reach totalWidth runes.
func PadRight(s string, totalWidth int) string {
	return PadRightChar(s, totalWidth, ' ')
}

// PadRightChar pads s on the right with paddingChar to reach totalWidth runes.
func PadRightChar(s string, totalWidth int, paddingChar rune) string {
	n := len([]rune(s))
	if n >= totalWidth {
		return s
	}
	return s + strings.Repeat(string(paddingChar), totalWidth-n)
}

// Insert inserts value into s at startIndex (rune-based, 0-indexed).
func Insert(s string, startIndex int, value string) string {
	runes := []rune(s)
	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex > len(runes) {
		startIndex = len(runes)
	}
	result := make([]rune, 0, len(runes)+len([]rune(value)))
	result = append(result, runes[:startIndex]...)
	result = append(result, []rune(value)...)
	result = append(result, runes[startIndex:]...)
	return string(result)
}

// Remove removes characters from s starting at startIndex through end of string.
func Remove(s string, startIndex int) string {
	runes := []rune(s)
	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex >= len(runes) {
		return s
	}
	return string(runes[:startIndex])
}

// RemoveCount removes count characters from s starting at startIndex.
func RemoveCount(s string, startIndex, count int) string {
	runes := []rune(s)
	n := len(runes)
	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex >= n {
		return s
	}
	end := startIndex + count
	if end > n {
		end = n
	}
	return string(append(runes[:startIndex], runes[end:]...))
}

// Replace replaces all occurrences of oldValue in s with newValue.
func Replace(s, oldValue, newValue string) string {
	return strings.ReplaceAll(s, oldValue, newValue)
}

// Substring returns the substring of s from startIndex to the end.
func Substring(s string, startIndex int) string {
	runes := []rune(s)
	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex >= len(runes) {
		return ""
	}
	return string(runes[startIndex:])
}

// SubstringLen returns the substring of s starting at startIndex with given length.
func SubstringLen(s string, startIndex, length int) string {
	runes := []rune(s)
	n := len(runes)
	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex >= n {
		return ""
	}
	end := startIndex + length
	if end > n {
		end = n
	}
	return string(runes[startIndex:end])
}

// Contains returns true if s contains value.
func Contains(s, value string) bool { return strings.Contains(s, value) }

// IndexOf returns the rune index of the first occurrence of value in s, or -1.
func IndexOf(s, value string) int {
	idx := strings.Index(s, value)
	if idx < 0 {
		return -1
	}
	return len([]rune(s[:idx]))
}

// Asc returns the ASCII / Unicode code point of the first rune in s.
func Asc(s string) int {
	runes := []rune(s)
	if len(runes) == 0 {
		return 0
	}
	return int(runes[0])
}

// Chr returns the string containing the rune for Unicode code point i.
func Chr(i int) string { return string(rune(i)) }

// ── Date / time functions ─────────────────────────────────────────────────────

// AddDays adds value days to date.
func AddDays(date time.Time, value float64) time.Time {
	return date.Add(time.Duration(value * float64(24*time.Hour)))
}

// AddHours adds value hours to date.
func AddHours(date time.Time, value float64) time.Time {
	return date.Add(time.Duration(value * float64(time.Hour)))
}

// AddMinutes adds value minutes to date.
func AddMinutes(date time.Time, value float64) time.Time {
	return date.Add(time.Duration(value * float64(time.Minute)))
}

// AddSeconds adds value seconds to date.
func AddSeconds(date time.Time, value float64) time.Time {
	return date.Add(time.Duration(value * float64(time.Second)))
}

// AddMonths adds value months to date.
func AddMonths(date time.Time, value int) time.Time { return date.AddDate(0, value, 0) }

// AddYears adds value years to date.
func AddYears(date time.Time, value int) time.Time { return date.AddDate(value, 0, 0) }

// DateSerial returns a time.Time for year/month/day.
func DateSerial(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// Day returns the day-of-month component (1–31).
func Day(date time.Time) int { return date.Day() }

// Month returns the month component (1–12).
func Month(date time.Time) int { return int(date.Month()) }

// Year returns the year component.
func Year(date time.Time) int { return date.Year() }

// Hour returns the hour component (0–23).
func Hour(date time.Time) int { return date.Hour() }

// Minute returns the minute component (0–59).
func Minute(date time.Time) int { return date.Minute() }

// Second returns the second component (0–59).
func Second(date time.Time) int { return date.Second() }

// DayOfWeek returns the name of the weekday (e.g. "Monday").
func DayOfWeek(date time.Time) string { return date.Weekday().String() }

// DayOfYear returns the ordinal day of the year (1–366).
func DayOfYear(date time.Time) int { return date.YearDay() }

// DaysInMonth returns the number of days in the given month/year.
func DaysInMonth(year, month int) int {
	// First day of next month minus one day.
	t := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
	return t.Day()
}

// MonthName returns the full English name of the given month (1–12).
func MonthName(month int) string {
	if month < 1 || month > 12 {
		return ""
	}
	return time.Month(month).String()
}

// WeekOfYear returns the ISO week number (1–53) for the given date.
func WeekOfYear(date time.Time) int {
	_, week := date.ISOWeek()
	return week
}

// ── Formatting ────────────────────────────────────────────────────────────────

// FormatNumber formats a numeric value with the given number of decimal places.
func FormatNumber(value float64, decimalPlaces int) string {
	return fmt.Sprintf("%.*f", decimalPlaces, value)
}

// FormatCurrency formats a numeric value as currency with 2 decimal places.
func FormatCurrency(value float64) string {
	return fmt.Sprintf("$%.2f", value)
}

// FormatPercent formats a fractional value as a percentage (×100, with %).
func FormatPercent(value float64) string {
	return fmt.Sprintf("%.2f%%", value*100)
}

// FormatDateTime formats a time.Time using a Go layout string.
// If layout is empty, RFC3339 is used.
func FormatDateTime(date time.Time, layout string) string {
	if layout == "" {
		layout = time.RFC3339
	}
	return date.Format(layout)
}

// ── Control flow ──────────────────────────────────────────────────────────────

// IIF returns trueVal when condition is true, falseVal otherwise.
func IIF(condition bool, trueVal, falseVal any) any {
	if condition {
		return trueVal
	}
	return falseVal
}

// Choose returns the value from vals at the given 1-based index.
// Returns nil if index is out of range.
func Choose(index int, vals ...any) any {
	if index < 1 || index > len(vals) {
		return nil
	}
	return vals[index-1]
}

// Switch evaluates pairs of (condition, value) and returns the first value
// whose condition is true.  If no condition matches, returns nil.
// The varargs must contain an even number of elements: cond1, val1, cond2, val2, …
func Switch(pairs ...any) any {
	for i := 0; i+1 < len(pairs); i += 2 {
		cond, ok := pairs[i].(bool)
		if ok && cond {
			return pairs[i+1]
		}
	}
	return nil
}

// ── All returns a map suitable for registering with the expression evaluator ──

// All returns all built-in functions as a name→implementation map.
func All() map[string]any {
	return map[string]any{
		// Math
		"MaxInt":    MaxInt,
		"MinInt":    MinInt,
		"MaxFloat":  MaxFloat,
		"MinFloat":  MinFloat,
		"Abs":       Abs,
		"Round":     Round,
		"RoundTo":   RoundTo,
		"Ceiling":   Ceiling,
		"Floor":     Floor,
		// String
		"Length":     Length,
		"LowerCase":  LowerCase,
		"UpperCase":  UpperCase,
		"TitleCase":  TitleCase,
		"Trim":       Trim,
		"PadLeft":    PadLeft,
		"PadRight":   PadRight,
		"Insert":     Insert,
		"Remove":     Remove,
		"Replace":    Replace,
		"Substring":  Substring,
		"Contains":   Contains,
		"IndexOf":    IndexOf,
		"Asc":        Asc,
		"Chr":        Chr,
		// Date/time
		"AddDays":      AddDays,
		"AddHours":     AddHours,
		"AddMinutes":   AddMinutes,
		"AddSeconds":   AddSeconds,
		"AddMonths":    AddMonths,
		"AddYears":     AddYears,
		"DateSerial":   DateSerial,
		"Day":          Day,
		"Month":        Month,
		"Year":         Year,
		"Hour":         Hour,
		"Minute":       Minute,
		"Second":       Second,
		"DayOfWeek":    DayOfWeek,
		"DayOfYear":    DayOfYear,
		"DaysInMonth":  DaysInMonth,
		"MonthName":    MonthName,
		"WeekOfYear":   WeekOfYear,
		// Formatting
		"FormatNumber":   FormatNumber,
		"FormatCurrency": FormatCurrency,
		"FormatPercent":  FormatPercent,
		"FormatDateTime": FormatDateTime,
		// Control flow
		"IIF":    IIF,
		"Choose": Choose,
		"Switch": Switch,
		// Barcode / special
		"NumToWords": NumToWords,
		"Roman":      ToRoman,
	}
}
