package functions

import (
	"fmt"
	"math"
	"strings"
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
