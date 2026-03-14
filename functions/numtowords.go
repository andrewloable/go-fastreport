package functions

import (
	"fmt"
	"math"
	"strings"
)

// NumToWords converts an integer to its English word representation.
// Examples:
//
//	0  → "zero"
//	1  → "one"
//	-5 → "negative five"
//	123 → "one hundred twenty-three"
//	1000 → "one thousand"
func NumToWords(n int64) string {
	if n == 0 {
		return "zero"
	}
	if n < 0 {
		return "negative " + NumToWords(-n)
	}
	return strings.TrimSpace(numToWordsPositive(n))
}

// NumToWordsFloat converts a float64 to its English word representation,
// with cents expressed as "and N/100" for the fractional part.
func NumToWordsFloat(v float64) string {
	whole := int64(math.Abs(v))
	cents := int(math.Round(math.Abs(v)*100)) % 100
	result := NumToWords(whole)
	if v < 0 {
		result = "negative " + result
	}
	if cents > 0 {
		result += fmt.Sprintf(" and %d/100", cents)
	}
	return result
}

// ── internal helpers ──────────────────────────────────────────────────────────

var ones = []string{
	"", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine",
	"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen",
	"sixteen", "seventeen", "eighteen", "nineteen",
}

var tens = []string{
	"", "", "twenty", "thirty", "forty", "fifty",
	"sixty", "seventy", "eighty", "ninety",
}

type scaleWord struct {
	value int64
	word  string
}

var scales = []scaleWord{
	{1_000_000_000_000, "trillion"},
	{1_000_000_000, "billion"},
	{1_000_000, "million"},
	{1_000, "thousand"},
	{100, "hundred"},
}

// numToWordsPositive converts a positive integer to English words.
func numToWordsPositive(n int64) string {
	if n == 0 {
		return ""
	}
	if n < 20 {
		return ones[n]
	}
	if n < 100 {
		t := tens[n/10]
		o := ones[n%10]
		if o == "" {
			return t
		}
		return t + "-" + o
	}

	for _, sc := range scales {
		if n >= sc.value {
			high := numToWordsPositive(n / sc.value)
			rem := n % sc.value
			if rem == 0 {
				return high + " " + sc.word
			}
			return high + " " + sc.word + " " + numToWordsPositive(rem)
		}
	}
	return ""
}
