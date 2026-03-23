package functions

import (
	"fmt"
	"math"
	"strings"
	"unicode"
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

// enCurrencies maps ISO currency codes to their English word forms.
var enCurrencies = map[string]struct {
	s1, sm string
	j1, jm string
}{
	"USD": {"dollar", "dollars", "cent", "cents"},
	"CAD": {"dollar", "dollars", "cent", "cents"},
	"EUR": {"euro", "euros", "cent", "cents"},
	"GBP": {"pound", "pounds", "penny", "pence"},
}

// enSimpleCase returns one if n==1, otherwise many.
func enSimpleCase(n int64, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// ConvertCurrencyEn converts a float64 monetary value to English words for the given ISO currency code.
// If decimalPartToWord is true, the cents are also expressed in words; otherwise as a numeric "NN" prefix.
func ConvertCurrencyEn(value float64, currencyName string, decimalPartToWord bool) (string, error) {
	cur, ok := enCurrencies[currencyName]
	if !ok {
		return "", fmt.Errorf("unknown currency: %s", currencyName)
	}
	n := int64(math.Abs(value))
	cents := int(math.Round((math.Abs(value) - float64(n)) * 100))
	if cents >= 100 {
		n++
		cents = 0
	}
	negative := value < 0

	var wholeWords string
	if n == 0 {
		wholeWords = "zero"
	} else {
		wholeWords = strings.TrimSpace(numToWordsPositive(n))
	}
	seniorWord := enSimpleCase(n, cur.s1, cur.sm)
	result := wholeWords + " " + seniorWord

	if negative {
		result = "negative " + result
	}

	if cur.j1 != "" {
		var juniorPart string
		if decimalPartToWord {
			var centsWords string
			if cents == 0 {
				centsWords = "zero"
			} else {
				centsWords = strings.TrimSpace(numToWordsPositive(int64(cents)))
			}
			juniorWord := enSimpleCase(int64(cents), cur.j1, cur.jm)
			juniorPart = centsWords + " " + juniorWord
		} else {
			juniorWord := enSimpleCase(int64(cents), cur.j1, cur.jm)
			juniorPart = fmt.Sprintf("%02d ", cents) + juniorWord
		}
		result = result + " " + juniorPart
	}

	r := []rune(result)
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
		result = string(r)
	}
	return result, nil
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
