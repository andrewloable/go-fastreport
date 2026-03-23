package functions

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// NumToWordsFa converts an integer to its Persian (Farsi) word representation.
func NumToWordsFa(n int64) string {
	if n == 0 {
		return "\u0635\u0641\u0631" // صفر
	}
	if n < 0 {
		return "\u0645\u0646\u0641\u06CC " + NumToWordsFa(-n) // منفی
	}
	return strings.TrimSpace(faPositive(n))
}

// NumToWordsFaFloat converts a float64 to Persian words.
func NumToWordsFaFloat(v float64) string {
	whole := int64(math.Abs(v))
	cents := int(math.Round(math.Abs(v)*100)) % 100
	result := NumToWordsFa(whole)
	if v < 0 {
		result = "\u0645\u0646\u0641\u06CC " + result
	}
	if cents > 0 {
		result += fmt.Sprintf(" \u0648 %d/100", cents)
	}
	return result
}

var faOnes = []string{
	"", "\u06CC\u06A9", "\u062F\u0648", "\u0633\u0647", "\u0686\u0647\u0627\u0631",
	"\u067E\u0646\u062C", "\u0634\u0634", "\u0647\u0641\u062A", "\u0647\u0634\u062A", "\u0646\u0647",
	"\u062F\u0647", "\u06CC\u0627\u0632\u062F\u0647", "\u062F\u0648\u0627\u0632\u062F\u0647",
	"\u0633\u06CC\u0632\u062F\u0647", "\u0686\u0647\u0627\u0631\u062F\u0647",
	"\u067E\u0627\u0646\u0632\u062F\u0647", "\u0634\u0627\u0646\u0632\u062F\u0647",
	"\u0647\u0641\u062F\u0647", "\u0647\u062C\u062F\u0647", "\u0646\u0648\u0632\u062F\u0647",
}

var faTens = []string{
	"", "\u062F\u0647", "\u0628\u06CC\u0633\u062A", "\u0633\u06CC",
	"\u0686\u0647\u0644", "\u067E\u0646\u062C\u0627\u0647",
	"\u0634\u0635\u062A", "\u0647\u0641\u062A\u0627\u062F",
	"\u0647\u0634\u062A\u0627\u062F", "\u0646\u0648\u062F",
}

var faHundreds = []string{
	"", "\u0635\u062F", "\u062F\u0648\u06CC\u0633\u062A",
	"\u0633\u06CC\u0635\u062F", "\u0686\u0647\u0627\u0631\u0635\u062F",
	"\u067E\u0627\u0646\u0635\u062F", "\u0634\u0634\u0635\u062F",
	"\u0647\u0641\u062A\u0635\u062F", "\u0647\u0634\u062A\u0635\u062F",
	"\u0646\u0647\u0635\u062F",
}

// faCurrencies maps ISO currency codes to their Persian word forms.
// Mirrors NumToWordsPersian.cs static constructor currencyList.
// WordInfo with a single word uses it for both one and many.
var faCurrencies = map[string]struct {
	s1, sm string
	j1, jm string
}{
	"EUR": {"یورو", "یورو", "یورو سنت", "یورو سنت"},
	"USD": {"دلار", "دلار", "سنت", "سنت"},
	"IRR": {"ریال\u200E", "ریال\u200E", "دینار", "دینار"},
}

// faSimpleCase returns one if n==1, otherwise many.
// Persian WordInfo typically has the same form for one and many.
func faSimpleCase(n int64, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// ConvertCurrencyFa converts a float64 monetary value to Persian words for the given ISO currency code.
// If decimalPartToWord is true, the cents are also expressed in words; otherwise as a numeric "NN " prefix.
// Mirrors NumToWordsPersian.cs / NumToWordsBase.cs ConvertCurrency logic.
func ConvertCurrencyFa(value float64, currencyName string, decimalPartToWord bool) (string, error) {
	cur, ok := faCurrencies[strings.ToUpper(currencyName)]
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
		wholeWords = "صفر"
	} else {
		wholeWords = strings.TrimSpace(faPositive(n))
	}
	seniorWord := faSimpleCase(n, cur.s1, cur.sm)
	result := wholeWords + " " + seniorWord

	if negative {
		result = "منفی " + result
	}

	if cur.j1 != "" {
		var juniorPart string
		if decimalPartToWord {
			var centsWords string
			if cents == 0 {
				centsWords = "صفر"
			} else {
				centsWords = strings.TrimSpace(faPositive(int64(cents)))
			}
			juniorWord := faSimpleCase(int64(cents), cur.j1, cur.jm)
			juniorPart = centsWords + " " + juniorWord
		} else {
			juniorWord := faSimpleCase(int64(cents), cur.j1, cur.jm)
			juniorPart = fmt.Sprintf("%02d ", cents) + juniorWord
		}
		// C# appends GetDecimalSeparator() + decimalPart → "و " + juniorPart
		result = result + " و " + juniorPart
	}

	r := []rune(result)
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
		result = string(r)
	}
	return result, nil
}

const faSep = " \u0648 " // " و "

func faPositive(n int64) string {
	if n >= 1_000_000_000_000 {
		g := n / 1_000_000_000_000
		rem := n % 1_000_000_000_000
		s := faPositive(g) + " \u062A\u0631\u06CC\u0644\u06CC\u0648\u0646"
		if rem > 0 {
			s += faSep + faPositive(rem)
		}
		return s
	}
	if n >= 1_000_000_000 {
		g := n / 1_000_000_000
		rem := n % 1_000_000_000
		s := faPositive(g) + " \u0645\u06CC\u0644\u06CC\u0627\u0631\u062F"
		if rem > 0 {
			s += faSep + faPositive(rem)
		}
		return s
	}
	if n >= 1_000_000 {
		g := n / 1_000_000
		rem := n % 1_000_000
		s := faPositive(g) + " \u0645\u06CC\u0644\u06CC\u0648\u0646"
		if rem > 0 {
			s += faSep + faPositive(rem)
		}
		return s
	}
	if n >= 1000 {
		g := n / 1000
		rem := n % 1000
		s := faPositive(g) + " \u0647\u0632\u0627\u0631"
		if rem > 0 {
			s += faSep + faPositive(rem)
		}
		return s
	}
	if n >= 100 {
		h := n / 100
		rem := n % 100
		s := faHundreds[h]
		if rem > 0 {
			s += faSep + faPositive(rem)
		}
		return s
	}
	if n < 20 {
		return faOnes[n]
	}
	t := n / 10
	o := n % 10
	if o == 0 {
		return faTens[t]
	}
	return faTens[t] + faSep + faOnes[o]
}
