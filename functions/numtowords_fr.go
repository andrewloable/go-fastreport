package functions

import (
	"fmt"
	"math"
	"strings"
)

// NumToWordsFr converts an integer to its French word representation.
func NumToWordsFr(n int64) string {
	if n == 0 {
		return "zéro"
	}
	if n < 0 {
		return "moins " + NumToWordsFr(-n)
	}
	return strings.TrimSpace(frPositive(n, false, true))
}

// NumToWordsFrFloat converts a float64 to French words.
func NumToWordsFrFloat(v float64) string {
	whole := int64(math.Abs(v))
	cents := int(math.Round(math.Abs(v)*100)) % 100
	result := NumToWordsFr(whole)
	if v < 0 {
		result = "moins " + result
	}
	if cents > 0 {
		result += fmt.Sprintf(" et %d/100", cents)
	}
	return result
}

var frOnes = []string{
	"", "un", "deux", "trois", "quatre", "cinq", "six", "sept", "huit", "neuf",
	"dix", "onze", "douze", "treize", "quatorze", "quinze", "seize",
	"dix-sept", "dix-huit", "dix-neuf",
}

var frTens = []string{
	"", "dix", "vingt", "trente", "quarante", "cinquante",
	"soixante", "soixante-dix", "quatre-vingt", "quatre-vingt-dix",
}

// 71-79 special forms
var frSeventies = []string{
	"soixante et onze", "soixante-douze", "soixante-treize", "soixante-quatorze",
	"soixante-quinze", "soixante-seize", "soixante-dix-sept", "soixante-dix-huit", "soixante-dix-neuf",
}

// 91-99 special forms
var frNineties = []string{
	"quatre-vingt-onze", "quatre-vingt-douze", "quatre-vingt-treize", "quatre-vingt-quatorze",
	"quatre-vingt-quinze", "quatre-vingt-seize", "quatre-vingt-dix-sept", "quatre-vingt-dix-huit", "quatre-vingt-dix-neuf",
}

// frPositive converts n>0 to French.
// female: use "une" for 1 (not "un")
// final: true when this group is the last (matters for "cents" / "vingts" plurals)
func frPositive(n int64, female, final bool) string {
	if n >= 1_000_000_000_000 {
		g := n / 1_000_000_000_000
		rem := n % 1_000_000_000_000
		suffix := "billions"
		if g == 1 {
			suffix = "billion"
		}
		s := frPositive(g, false, false) + " " + suffix
		if rem > 0 {
			s += " " + frPositive(rem, female, final)
		}
		return s
	}
	if n >= 1_000_000_000 {
		g := n / 1_000_000_000
		rem := n % 1_000_000_000
		suffix := "milliards"
		if g == 1 {
			suffix = "milliard"
		}
		s := frPositive(g, false, false) + " " + suffix
		if rem > 0 {
			s += " " + frPositive(rem, female, final)
		}
		return s
	}
	if n >= 1_000_000 {
		g := n / 1_000_000
		rem := n % 1_000_000
		suffix := "millions"
		if g == 1 {
			suffix = "million"
		}
		s := frPositive(g, false, false) + " " + suffix
		if rem > 0 {
			s += " " + frPositive(rem, female, final)
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
			prefix = frPositive(g, false, false) + " "
		}
		s := prefix + "mille"
		if rem > 0 {
			s += " " + frPositive(rem, female, final)
		}
		return s
	}
	if n >= 100 {
		h := n / 100
		rem := n % 100
		var prefix string
		if h == 1 {
			prefix = "cent"
		} else {
			prefix = frOnes[h] + " cent"
		}
		if rem == 0 && h > 1 && final {
			return prefix + "s"
		}
		if rem == 0 {
			return prefix
		}
		return prefix + " " + fr2Digits(rem, female, final)
	}
	return fr2Digits(n, female, final)
}

func fr2Digits(n int64, female, final bool) string {
	if n < 20 {
		if n == 1 && female {
			return "une"
		}
		return frOnes[n]
	}
	t := n / 10
	o := n % 10
	switch {
	case n == 80:
		if final {
			return "quatre-vingts"
		}
		return "quatre-vingt"
	case t == 7: // 70-79
		if o == 0 {
			return "soixante-dix"
		}
		return frSeventies[o-1]
	case t == 9: // 90-99
		if o == 0 {
			return "quatre-vingt-dix"
		}
		return frNineties[o-1]
	}
	tens := frTens[t]
	if o == 0 {
		return tens
	}
	oStr := frOnes[o]
	if o == 1 {
		if female {
			oStr = "une"
		}
		return tens + " et " + oStr
	}
	return tens + "-" + oStr
}
