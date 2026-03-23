package functions

import (
	"fmt"
	"math"
	"strings"
	"unicode"
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

// inCurrencies maps ISO currency codes to their Indian English word forms.
// Mirrors NumToWordsIn.cs static constructor currencyList.
var inCurrencies = map[string]struct {
	s1, sm string
	j1, jm string
}{
	"USD": {"dollar", "dollars", "cent", "cents"},
	"EUR": {"euro", "euros", "cent", "cents"},
	"INR": {"rupee", "rupees", "paise", "paise"},
}

// inSimpleCase returns one if n==1, otherwise many.
func inSimpleCase(n int64, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// ConvertCurrencyIn converts a float64 monetary value to Indian English words for the given ISO currency code.
// If decimalPartToWord is true, the cents/paise are also expressed in words; otherwise as a numeric "NN " prefix.
// Mirrors NumToWordsIn.cs / NumToWordsBase.cs ConvertCurrency logic.
func ConvertCurrencyIn(value float64, currencyName string, decimalPartToWord bool) (string, error) {
	cur, ok := inCurrencies[strings.ToUpper(currencyName)]
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
		wholeWords = strings.TrimSpace(inPositive(n))
	}
	seniorWord := inSimpleCase(n, cur.s1, cur.sm)
	result := wholeWords + " " + seniorWord

	if negative {
		result = "minus " + result
	}

	if cur.j1 != "" {
		var juniorPart string
		if decimalPartToWord {
			var centsWords string
			if cents == 0 {
				centsWords = "zero"
			} else {
				centsWords = strings.TrimSpace(inPositive(int64(cents)))
			}
			juniorWord := inSimpleCase(int64(cents), cur.j1, cur.jm)
			juniorPart = centsWords + " " + juniorWord
		} else {
			juniorWord := inSimpleCase(int64(cents), cur.j1, cur.jm)
			juniorPart = fmt.Sprintf("%02d ", cents) + juniorWord
		}
		// C# appends GetDecimalSeparator() + decimalPart → "and " + juniorPart
		result = result + " and " + juniorPart
	}

	r := []rune(result)
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
		result = string(r)
	}
	return result, nil
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
