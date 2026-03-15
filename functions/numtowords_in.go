package functions

import (
	"fmt"
	"math"
	"strings"
)

// NumToWordsIn converts an integer to Indian English words using the
// lakh/crore numbering system.
//
// Indian grouping: ones < thousand < lakh (10^5) < crore (10^7) < arab (10^9)
// < kharab (10^11) < nil (10^13).
func NumToWordsIn(n int64) string {
	if n == 0 {
		return "zero"
	}
	if n < 0 {
		return "minus " + NumToWordsIn(-n)
	}
	return strings.TrimSpace(inPositive(n))
}

// NumToWordsInFloat converts a float64 to Indian English words.
func NumToWordsInFloat(v float64) string {
	whole := int64(math.Abs(v))
	cents := int(math.Round(math.Abs(v)*100)) % 100
	result := NumToWordsIn(whole)
	if v < 0 {
		result = "minus " + result
	}
	if cents > 0 {
		result += fmt.Sprintf(" and %d/100", cents)
	}
	return result
}

// inPositive converts n > 0 to Indian English.
func inPositive(n int64) string {
	const (
		nil_    int64 = 10_000_000_000_000 // 10^13
		kharab  int64 = 100_000_000_000    // 10^11
		arab    int64 = 1_000_000_000      // 10^9
		crore   int64 = 10_000_000         // 10^7
		lakh    int64 = 100_000            // 10^5
		thousand int64 = 1_000
		hundred int64 = 100
	)

	scales := []struct {
		value int64
		name  string
	}{
		{nil_, "nil"},
		{kharab, "kharab"},
		{arab, "arab"},
		{crore, "crore"},
		{lakh, "lakh"},
		{thousand, "thousand"},
		{hundred, "hundred"},
	}

	var parts []string
	rem := n
	for _, sc := range scales {
		if rem >= sc.value {
			g := rem / sc.value
			rem = rem % sc.value
			parts = append(parts, in3digits(g)+" "+sc.name)
		}
	}
	if rem > 0 {
		parts = append(parts, in3digits(rem))
	}
	return strings.Join(parts, " ")
}

var inOnes = []string{
	"", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine",
	"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen",
	"sixteen", "seventeen", "eighteen", "nineteen",
}

var inTens = []string{
	"", "ten", "twenty", "thirty", "forty", "fifty",
	"sixty", "seventy", "eighty", "ninety",
}

// in3digits handles numbers 1-999.
func in3digits(n int64) string {
	if n >= 100 {
		h := n / 100
		rem := n % 100
		s := inOnes[h] + " hundred"
		if rem > 0 {
			s += " " + in2digits(rem)
		}
		return s
	}
	return in2digits(n)
}

func in2digits(n int64) string {
	if n < 20 {
		return inOnes[n]
	}
	t := n / 10
	o := n % 10
	if o == 0 {
		return inTens[t]
	}
	return inTens[t] + "-" + inOnes[o]
}
