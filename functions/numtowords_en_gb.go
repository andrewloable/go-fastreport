package functions

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// NumToWordsEnGb converts an integer to British English words.
// Uses the long scale: billion = 10^12, milliard = 10^9.
func NumToWordsEnGb(n int64) string {
	if n == 0 {
		return "zero"
	}
	if n < 0 {
		return "negative " + NumToWordsEnGb(-n)
	}
	return strings.TrimSpace(enGbPositive(n))
}

// NumToWordsEnGbFloat converts a float64 to British English words.
func NumToWordsEnGbFloat(v float64) string {
	whole := int64(math.Abs(v))
	cents := int(math.Round(math.Abs(v)*100)) % 100
	result := NumToWordsEnGb(whole)
	if v < 0 {
		result = "negative " + result
	}
	if cents > 0 {
		result += fmt.Sprintf(" and %d/100", cents)
	}
	return result
}

// ConvertCurrencyEnGb converts a float64 monetary value to British English words for the given ISO currency code.
// If decimalPartToWord is true, the cents are also expressed in words; otherwise as a numeric "NN" prefix.
// It uses the same enCurrencies table and enSimpleCase helper as ConvertCurrencyEn.
func ConvertCurrencyEnGb(value float64, currencyName string, decimalPartToWord bool) (string, error) {
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
		wholeWords = strings.TrimSpace(enGbPositive(n))
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
				centsWords = strings.TrimSpace(enGbPositive(int64(cents)))
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

// enGbPositive uses the same word tables as the US English implementation
// but with British long-scale names: milliard (10^9), billion (10^12).
func enGbPositive(n int64) string {
	if n == 0 {
		return ""
	}
	if n >= 1_000_000_000_000 {
		high := enGbPositive(n / 1_000_000_000_000)
		rem := n % 1_000_000_000_000
		if rem == 0 {
			return high + " billion"
		}
		return high + " billion " + enGbPositive(rem)
	}
	if n >= 1_000_000_000 {
		high := enGbPositive(n / 1_000_000_000)
		rem := n % 1_000_000_000
		if rem == 0 {
			return high + " milliard"
		}
		return high + " milliard " + enGbPositive(rem)
	}
	// Reuse the US English helpers for millions and below.
	return numToWordsPositive(n)
}
