package functions

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// NumToWordsEs converts an integer to its Spanish word representation.
func NumToWordsEs(n int64) string {
	if n == 0 {
		return "cero"
	}
	if n < 0 {
		return "minus " + NumToWordsEs(-n)
	}
	return strings.TrimSpace(esPositive(n))
}

// NumToWordsEsFloat converts a float64 to Spanish words.
func NumToWordsEsFloat(v float64) string {
	whole := int64(math.Abs(v))
	cents := int(math.Round(math.Abs(v)*100)) % 100
	result := NumToWordsEs(whole)
	if v < 0 {
		result = "minus " + result
	}
	if cents > 0 {
		result += fmt.Sprintf(" con %d/100", cents)
	}
	return result
}

// esCurrencies maps ISO currency codes to their Spanish word forms.
var esCurrencies = map[string]struct {
	s1, sm string
	j1, jm string
}{
	"USD": {"dolar", "dolares", "centavo", "centavos"},
	"CAD": {"dolar", "dolares", "centavo", "centavos"},
	"EUR": {"euro", "euros", "centavo", "centavos"},
	"MXN": {"peso", "pesos", "centavo", "centavos"},
}

// esSimpleCase returns one if n==1, otherwise many.
func esSimpleCase(n int64, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// ConvertCurrencyEs converts a float64 monetary value to Spanish words for the given ISO currency code.
// If decimalPartToWord is true, the cents are also expressed in words; otherwise as a numeric "NN" prefix.
func ConvertCurrencyEs(value float64, currencyName string, decimalPartToWord bool) (string, error) {
	cur, ok := esCurrencies[currencyName]
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
		wholeWords = strings.TrimSpace(esPositive(n))
	}
	seniorWord := esSimpleCase(n, cur.s1, cur.sm)
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
				centsWords = strings.TrimSpace(esPositive(int64(cents)))
			}
			juniorWord := esSimpleCase(int64(cents), cur.j1, cur.jm)
			juniorPart = centsWords + " " + juniorWord
		} else {
			juniorWord := esSimpleCase(int64(cents), cur.j1, cur.jm)
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

// esFixed covers 0-29 (indices 1-29 are used, index 0 is unused).
var esFixed = []string{
	"", "un", "dos", "tres", "cuatro", "cinco", "seis", "siete", "ocho", "nueve",
	"diez", "once", "doce", "trece", "catorce", "quince",
	"dieciséis", "diecisiete", "dieciocho", "diecinueve",
	"veinte", "veintiún", "veintidós", "veintitrés", "veinticuatro",
	"veinticinco", "veintiséis", "veintisiete", "veintiocho", "veintinueve",
}

var esTens = []string{
	"", "diez", "veinte", "treinta", "cuarenta", "cincuenta",
	"sesenta", "setenta", "ochenta", "noventa",
}

var esHundreds = []string{
	"", "cien", "doscientos", "trescientos", "cuatrocientos", "quinientos",
	"seiscientos", "setecientos", "ochocientos", "novecientos",
}

func esPositive(n int64) string {
	if n >= 1_000_000_000_000 {
		g := n / 1_000_000_000_000
		rem := n % 1_000_000_000_000
		suffix := "billones"
		if g == 1 {
			suffix = "billón"
		}
		s := esPositive(g) + " " + suffix
		if rem > 0 {
			s += " " + esPositive(rem)
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
		s := esPositive(g) + " " + suffix
		if rem > 0 {
			s += " " + esPositive(rem)
		}
		return s
	}
	if n >= 1000 {
		g := n / 1000
		rem := n % 1000
		var prefix string
		if g == 1 {
			prefix = ""
		} else {
			prefix = esPositive(g) + " "
		}
		s := prefix + "mil"
		if rem > 0 {
			s += " " + esPositive(rem)
		}
		return s
	}
	if n >= 100 {
		h := n / 100
		rem := n % 100
		hundredWord := esHundreds[h]
		if h == 1 && rem > 0 {
			hundredWord = "ciento"
		}
		if rem == 0 {
			return hundredWord
		}
		return hundredWord + " " + esPositive(rem)
	}
	// 1-99
	if n < 30 {
		return esFixed[n]
	}
	t := n / 10
	o := n % 10
	if o == 0 {
		return esTens[t]
	}
	return esTens[t] + " y " + esFixed[o]
}
