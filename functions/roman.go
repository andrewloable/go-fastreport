package functions

import (
	"fmt"
	"strings"
)

// ToRoman converts a positive integer to its Roman numeral representation.
// Returns an error for values outside the range 1–3999.
// Examples:
//
//	1    → "I"
//	4    → "IV"
//	14   → "XIV"
//	1994 → "MCMXCIV"
//	3999 → "MMMCMXCIX"
func ToRoman(n int) (string, error) {
	if n < 1 || n > 3999 {
		return "", fmt.Errorf("roman: value %d is out of range [1, 3999]", n)
	}

	type romanPair struct {
		value  int
		symbol string
	}
	pairs := []romanPair{
		{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"},
		{100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"},
		{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"},
	}

	var sb strings.Builder
	for _, p := range pairs {
		for n >= p.value {
			sb.WriteString(p.symbol)
			n -= p.value
		}
	}
	return sb.String(), nil
}

// MustToRoman converts n to a Roman numeral, panicking if n is out of range.
// Useful in tests and template contexts where error handling is inconvenient.
func MustToRoman(n int) string {
	r, err := ToRoman(n)
	if err != nil {
		panic(err)
	}
	return r
}

// FromRoman converts a Roman numeral string to its integer value.
// Returns an error for invalid input.
func FromRoman(s string) (int, error) {
	values := map[rune]int{
		'I': 1, 'V': 5, 'X': 10, 'L': 50,
		'C': 100, 'D': 500, 'M': 1000,
	}
	runes := []rune(strings.ToUpper(s))
	if len(runes) == 0 {
		return 0, fmt.Errorf("roman: empty string")
	}
	result := 0
	for i, r := range runes {
		val, ok := values[r]
		if !ok {
			return 0, fmt.Errorf("roman: invalid character %q", r)
		}
		// If this symbol is less than the next, subtract it (e.g. IV, IX).
		if i+1 < len(runes) && values[runes[i+1]] > val {
			result -= val
		} else {
			result += val
		}
	}
	return result, nil
}
