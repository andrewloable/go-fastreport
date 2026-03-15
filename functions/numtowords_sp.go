package functions

import (
	"fmt"
	"math"
	"strings"
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
