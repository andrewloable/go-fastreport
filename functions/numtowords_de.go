package functions

import (
	"fmt"
	"math"
	"strings"
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
