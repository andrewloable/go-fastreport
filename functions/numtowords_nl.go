package functions

import (
	"fmt"
	"math"
	"strings"
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
