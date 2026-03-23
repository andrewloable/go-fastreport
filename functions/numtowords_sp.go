package functions

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// NumToWordsSp converts an integer to its Spanish word representation using
// the Sp (Spain) dialectal variant. The key difference from NumToWordsEs is
// the use of "millardo/millardos" for 10^9 instead of "mil millones".
func NumToWordsSp(n int64) string {
	if n == 0 {
		return "cero"
	}
	if n < 0 {
		return "menos " + NumToWordsSp(-n)
	}
	return strings.TrimSpace(spPositive(n))
}

// NumToWordsSpFloat converts a float64 to Spanish (Sp) words.
func NumToWordsSpFloat(v float64) string {
	whole := int64(math.Abs(v))
	cents := int(math.Round(math.Abs(v)*100)) % 100
	result := NumToWordsSp(whole)
	if v < 0 {
		result = "menos " + result
	}
	if cents > 0 {
		result += fmt.Sprintf(" y %d/100", cents)
	}
	return result
}

// spCurrencies maps ISO currency codes to their Spanish (Sp) word forms.
// Mirrors NumToWordSp.cs static constructor currencyList.
var spCurrencies = map[string]struct {
	s1, sm string
	j1, jm string
}{
	"EUR": {"euro", "euros", "céntimo", "céntimos"},
	"USD": {"dólar", "dólares", "céntimo", "céntimos"},
}

// spSimpleCase returns one if n==1, otherwise many.
func spSimpleCase(n int64, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// ConvertCurrencySp converts a float64 monetary value to Spanish (Sp) words for the given ISO currency code.
// If decimalPartToWord is true, the cents are also expressed in words; otherwise as a numeric "NN " prefix.
// Mirrors NumToWordSp.cs / NumToWordsBase.cs ConvertCurrency logic.
func ConvertCurrencySp(value float64, currencyName string, decimalPartToWord bool) (string, error) {
	cur, ok := spCurrencies[strings.ToUpper(currencyName)]
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
		wholeWords = "cero"
	} else {
		wholeWords = strings.TrimSpace(spPositive(n))
	}
	seniorWord := spSimpleCase(n, cur.s1, cur.sm)
	result := wholeWords + " " + seniorWord

	if negative {
		result = "menos " + result
	}

	if cur.j1 != "" {
		var juniorPart string
		if decimalPartToWord {
			var centsWords string
			if cents == 0 {
				centsWords = "cero"
			} else {
				centsWords = strings.TrimSpace(spPositive(int64(cents)))
			}
			juniorWord := spSimpleCase(int64(cents), cur.j1, cur.jm)
			juniorPart = centsWords + " " + juniorWord
		} else {
			juniorWord := spSimpleCase(int64(cents), cur.j1, cur.jm)
			juniorPart = fmt.Sprintf("%02d ", cents) + juniorWord
		}
		// C# appends GetDecimalSeparator() + decimalPart → "y " + juniorPart
		result = result + " y " + juniorPart
	}

	r := []rune(result)
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
		result = string(r)
	}
	return result, nil
}

// spPositive renders n > 0 using the Sp dialect scale words.
// Shares word tables with the Es implementation.
func spPositive(n int64) string {
	if n >= 1_000_000_000_000 {
		g := n / 1_000_000_000_000
		rem := n % 1_000_000_000_000
		suffix := "billiones"
		if g == 1 {
			suffix = "billón"
		}
		s := spMillionPrefix(g) + " " + suffix
		if rem > 0 {
			s += " " + spPositive(rem)
		}
		return s
	}
	if n >= 1_000_000_000 {
		// Sp uses "millardo/millardos" for 10^9 (short scale for this dialectal form).
		g := n / 1_000_000_000
		rem := n % 1_000_000_000
		suffix := "millardos"
		if g == 1 {
			suffix = "millardo"
		}
		s := spMillionPrefix(g) + " " + suffix
		if rem > 0 {
			s += " " + spPositive(rem)
		}
		return s
	}
	if n >= 1_000_000 {
		g := n / 1_000_000
		rem := n % 1_000_000
		suffix := "millones"
		if g == 1 {
			suffix = "millón"
		}
		s := spMillionPrefix(g) + " " + suffix
		if rem > 0 {
			s += " " + spPositive(rem)
		}
		return s
	}
	if n >= 1000 {
		g := n / 1000
		rem := n % 1000
		prefix := ""
		if g > 1 {
			prefix = spBelow1000(g) + " "
		}
		s := prefix + "mil"
		if rem > 0 {
			s += " " + spPositive(rem)
		}
		return s
	}
	return spBelow1000(n)
}

// spMillionPrefix returns the count word before a large scale word.
// When the count itself is exactly 1, returns "un" (not "uno").
func spMillionPrefix(n int64) string {
	if n == 1 {
		return "un"
	}
	return spBelow1000(n)
}

// spBelow1000 renders 1–999 in Spanish (Sp), reusing Es word tables.
func spBelow1000(n int64) string {
	if n >= 100 {
		h := n / 100
		rem := n % 100
		base := esHundreds[h]
		// "cien" → "ciento" when followed by more words.
		if h == 1 && rem > 0 {
			base = "ciento"
		}
		if rem == 0 {
			return base
		}
		return base + " " + spBelow100(rem)
	}
	return spBelow100(n)
}

// spBelow100 renders 1–99 using the Sp word lists.
func spBelow100(n int64) string {
	if n < 30 {
		return spFixed[n]
	}
	t := n / 10
	o := n % 10
	if o == 0 {
		return esTens[t]
	}
	return esTens[t] + " y " + spFixed[o]
}

// spFixed covers 1–29. Index 21–29 differ from Es (use "veintiuno" etc.).
var spFixed = []string{
	"", "uno", "dos", "tres", "cuatro", "cinco", "seis", "siete", "ocho", "nueve",
	"diez", "once", "doce", "trece", "catorce", "quince",
	"dieciséis", "diecisiete", "dieciocho", "diecinueve",
	"veinte", "veintiuno", "veintidós", "veintitrés", "veinticuatro",
	"veinticinco", "veintiséis", "veintisiete", "veintiocho", "veintinueve",
}
