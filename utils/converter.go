package utils

import (
	"image/color"
	"strconv"
	"strings"
)

// BoolFromString parses a boolean from a string.
// It returns true when s (trimmed, case-insensitive) equals "true" and false
// for any other value, including empty strings.
func BoolFromString(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), "true")
}

// BoolToString converts a bool to its canonical lowercase string representation
// ("true" or "false").
func BoolToString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// IntFromString parses a decimal integer from s.
// On parse failure it returns 0.
func IntFromString(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}

// IntToString converts an int to its decimal string representation.
func IntToString(v int) string {
	return strconv.Itoa(v)
}

// Float32FromString parses a 32-bit floating-point number from s using the
// invariant (dot) decimal separator. On parse failure it returns 0.
func Float32FromString(s string) float32 {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 32)
	if err != nil {
		return 0
	}
	return float32(f)
}

// Float32ToString converts a float32 to a string using the shortest
// representation that round-trips back to the same value (invariant locale).
func Float32ToString(v float32) string {
	return strconv.FormatFloat(float64(v), 'f', -1, 32)
}

// RGBAFromString parses a color.RGBA from s.
// Accepted formats are the same as [ParseColor]: "#RGB", "#RRGGBB",
// "#AARRGGBB", or a decimal ARGB integer string.
// On parse failure it returns a zero color.RGBA (transparent black).
func RGBAFromString(s string) color.RGBA {
	c, err := ParseColor(s)
	if err != nil {
		return color.RGBA{}
	}
	return c
}

// RGBAToString converts c to its "#AARRGGBB" canonical hex string, which is
// the FRX serialisation format understood by [RGBAFromString].
func RGBAToString(c color.RGBA) string {
	return FormatColor(c)
}
