package functions

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// NumToWordsDe converts an integer to its German word representation.
// Examples:
//
//	0  → "null"
//	21 → "einundzwanzig"
//	100 → "einhundert"
//	1000 → "tausend"
func NumToWordsDe(n int64) string {
	if n == 0 {
		return "null"
	}
	if n < 0 {
		return "minus " + NumToWordsDe(-n)
	}
	return strings.TrimSpace(dePositive(n, false))
}

// NumToWordsDeFloat converts a float64 to German words with "und N/100" cents.
func NumToWordsDeFloat(v float64) string {
	whole := int64(math.Abs(v))
	cents := int(math.Round(math.Abs(v)*100)) % 100
	result := NumToWordsDe(whole)
	if v < 0 {
		result = "minus " + result
	}
	if cents > 0 {
		result += fmt.Sprintf(" und %d/100", cents)
	}
	return result
}

// deCurrencies maps ISO currency codes to their German word forms.
var deCurrencies = map[string]struct {
	s1, sm string
	j1, jm string
}{
	"USD": {"Dollar", "Dollar", "Cent", "Cent"},
	"CAD": {"Dollar", "Dollar", "Cent", "Cent"},
	"EUR": {"Euro", "Euro", "Cent", "Cent"},
	"GBP": {"Pfund", "Pfund", "Penny", "Penny"},
}

// deSimpleCase returns one if n==1, otherwise many.
func deSimpleCase(n int64, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// ConvertCurrencyDe converts a float64 monetary value to German words for the given ISO currency code.
// If decimalPartToWord is true, the cents are also expressed in words; otherwise as a numeric "NN" prefix.
func ConvertCurrencyDe(value float64, currencyName string, decimalPartToWord bool) (string, error) {
	cur, ok := deCurrencies[currencyName]
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
		wholeWords = "null"
	} else {
		wholeWords = strings.TrimSpace(dePositive(n, false))
	}
	seniorWord := deSimpleCase(n, cur.s1, cur.sm)
	result := wholeWords + " " + seniorWord

	if negative {
		result = "minus " + result
	}

	if cur.j1 != "" {
		var juniorPart string
		if decimalPartToWord {
			var centsWords string
			if cents == 0 {
				centsWords = "null"
			} else {
				centsWords = strings.TrimSpace(dePositive(int64(cents), false))
			}
			juniorWord := deSimpleCase(int64(cents), cur.j1, cur.jm)
			juniorPart = centsWords + " " + juniorWord
		} else {
			juniorWord := deSimpleCase(int64(cents), cur.j1, cur.jm)
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

var deOnes = []string{
	"", "ein", "zwei", "drei", "vier", "fünf", "sechs", "sieben", "acht", "neun",
	"zehn", "elf", "zwölf", "dreizehn", "vierzehn", "fünfzehn",
	"sechzehn", "siebzehn", "achtzehn", "neunzehn",
}

var deTens = []string{
	"", "zehn", "zwanzig", "dreißig", "vierzig", "fünfzig",
	"sechzig", "siebzig", "achtzig", "neunzig",
}

var deHundreds = []string{
	"", "einhundert", "zweihundert", "dreihundert", "vierhundert", "fünfhundert",
	"sechshundert", "siebenhundert", "achthundert", "neunhundert",
}

// dePositive converts n > 0 to German words.
// female=true uses "eine" instead of "ein" for the value 1.
func dePositive(n int64, female bool) string {
	if n >= 1_000_000_000_000 {
		g := n / 1_000_000_000_000
		rem := n % 1_000_000_000_000
		suffix := "Billionen"
		if g == 1 {
			suffix = "Billion"
		}
		s := dePositive(g, false) + " " + suffix
		if rem > 0 {
			s += " " + dePositive(rem, female)
		}
		return s
	}
	if n >= 1_000_000 {
		g := n / 1_000_000
		rem := n % 1_000_000
		suffix := "Millionen"
		if g == 1 {
			suffix = "Million"
		}
		s := dePositive(g, false) + " " + suffix
		if rem > 0 {
			s += " " + dePositive(rem, female)
		}
		return s
	}
	if n >= 1000 {
		g := n / 1000
		rem := n % 1000
		var prefix string
		if g == 1 {
			prefix = "ein"
		} else {
			prefix = dePositive(g, false)
		}
		s := prefix + "tausend"
		if rem > 0 {
			s += dePositive(rem, female)
		}
		return s
	}
	if n >= 100 {
		h := n / 100
		rem := n % 100
		s := deHundreds[h]
		if rem > 0 {
			s += dePositive(rem, female)
		}
		return s
	}
	if n < 20 {
		if n == 1 {
			if female {
				return "eine"
			}
			return "ein"
		}
		return deOnes[n]
	}
	// 20-99
	o := n % 10
	t := n / 10
	if o == 0 {
		return deTens[t]
	}
	oStr := deOnes[o]
	if o == 1 && female {
		oStr = "eine"
	}
	return oStr + "und" + deTens[t]
}
