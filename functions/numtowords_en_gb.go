package functions

import (
	"fmt"
	"math"
	"strings"
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
