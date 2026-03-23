package functions

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// NumToWordsNl converts an integer to its Dutch word representation.
func NumToWordsNl(n int64) string {
	if n == 0 {
		return "nul"
	}
	if n < 0 {
		return "min " + NumToWordsNl(-n)
	}
	return strings.TrimSpace(nlPositive(n))
}

// NumToWordsNlFloat converts a float64 to Dutch words.
func NumToWordsNlFloat(v float64) string {
	whole := int64(math.Abs(v))
	cents := int(math.Round(math.Abs(v)*100)) % 100
	result := NumToWordsNl(whole)
	if v < 0 {
		result = "min " + result
	}
	if cents > 0 {
		result += fmt.Sprintf(" en %d/100", cents)
	}
	return result
}

var nlOnes = []string{
	"", "een", "twee", "drie", "vier", "vijf", "zes", "zeven", "acht", "negen",
	"tien", "elf", "twaalf", "dertien", "veertien", "vijftien",
	"zestien", "zeventien", "achttien", "negentien",
}

var nlTens = []string{
	"", "tien", "twintig", "dertig", "veertig", "vijftig",
	"zestig", "zeventig", "tachtig", "negentig",
}

var nlHundreds = []string{
	"", "honderd", "tweehonderd", "driehonderd", "vierhonderd", "vijfhonderd",
	"zeshonderd", "zevenhonderd", "achthonderd", "negenhonderd",
}

// nlCurrencies maps ISO currency codes to their Dutch word forms.
// CAD is aliased to USD as in the C# source (NumToWordsNl.cs GetCurrency).
var nlCurrencies = map[string]struct {
	s1, sm string
	j1, jm string
}{
	"USD": {"dollar", "dollar", "cent", "cent"},
	"CAD": {"dollar", "dollar", "cent", "cent"},
	"EUR": {"euro", "euro", "cent", "cent"},
	"GBP": {"pound", "pound", "penny", "penny"},
}

// nlSimpleCase returns one if n==1, otherwise many.
func nlSimpleCase(n int64, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// ConvertCurrencyNl converts a float64 monetary value to Dutch words for the given ISO currency code.
// If decimalPartToWord is true, the cents are also expressed in words; otherwise as a numeric "NN " prefix.
// Mirrors NumToWordsNl.cs / NumToWordsBase.cs ConvertCurrency logic.
func ConvertCurrencyNl(value float64, currencyName string, decimalPartToWord bool) (string, error) {
	cur, ok := nlCurrencies[strings.ToUpper(currencyName)]
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
		wholeWords = "nul"
	} else {
		wholeWords = strings.TrimSpace(nlPositive(n))
	}
	seniorWord := nlSimpleCase(n, cur.s1, cur.sm)
	result := wholeWords + " " + seniorWord

	if negative {
		result = "min " + result
	}

	if cur.j1 != "" {
		var juniorPart string
		if decimalPartToWord {
			var centsWords string
			if cents == 0 {
				centsWords = "nul"
			} else {
				centsWords = strings.TrimSpace(nlPositive(int64(cents)))
			}
			juniorWord := nlSimpleCase(int64(cents), cur.j1, cur.jm)
			juniorPart = centsWords + " " + juniorWord
		} else {
			juniorWord := nlSimpleCase(int64(cents), cur.j1, cur.jm)
			juniorPart = fmt.Sprintf("%02d ", cents) + juniorWord
		}
		// C# appends GetDecimalSeparator() + decimalPart → "en " + juniorPart
		result = result + " en " + juniorPart
	}

	r := []rune(result)
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
		result = string(r)
	}
	return result, nil
}

func nlPositive(n int64) string {
	if n >= 1_000_000_000_000 {
		g := n / 1_000_000_000_000
		rem := n % 1_000_000_000_000
		s := nlPositive(g) + " trillion"
		if rem > 0 {
			s += " " + nlPositive(rem)
		}
		return s
	}
	if n >= 1_000_000_000 {
		g := n / 1_000_000_000
		rem := n % 1_000_000_000
		s := nlPositive(g) + " miljard"
		if rem > 0 {
			s += " " + nlPositive(rem)
		}
		return s
	}
	if n >= 1_000_000 {
		g := n / 1_000_000
		rem := n % 1_000_000
		s := nlPositive(g) + " miljoen"
		if rem > 0 {
			s += " " + nlPositive(rem)
		}
		return s
	}
	if n >= 1000 {
		g := n / 1000
		rem := n % 1000
		var prefix string
		if g == 1 {
			prefix = "één"
		} else {
			prefix = nlPositive(g)
		}
		s := prefix + "duizend"
		if rem > 0 {
			s += " " + nlPositive(rem)
		}
		return s
	}
	if n >= 100 {
		h := n / 100
		rem := n % 100
		s := nlHundreds[h]
		if rem > 0 {
			s += " " + nlPositive(rem)
		}
		return s
	}
	if n < 20 {
		return nlOnes[n]
	}
	// 20-99: ones + separator + tens
	o := n % 10
	t := n / 10
	if o == 0 {
		return nlTens[t]
	}
	oStr := nlOnes[o]
	sep := "en"
	if o == 2 || o == 3 {
		sep = "\u00EBn" // ën
	}
	return oStr + sep + nlTens[t]
}
