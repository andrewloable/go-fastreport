package functions

import (
	"fmt"
	"math"
	"strings"
)

// NumToWordsPl converts an integer to its Polish word representation.
func NumToWordsPl(n int64) string {
	if n == 0 {
		return "zero"
	}
	if n < 0 {
		return "minus " + NumToWordsPl(-n)
	}
	return strings.TrimSpace(plPositive(n, false))
}

// NumToWordsPlFloat converts a float64 to Polish words.
func NumToWordsPlFloat(v float64) string {
	whole := int64(math.Abs(v))
	cents := int(math.Round(math.Abs(v)*100)) % 100
	result := NumToWordsPl(whole)
	if v < 0 {
		result = "minus " + result
	}
	if cents > 0 {
		result += fmt.Sprintf(" i %d/100", cents)
	}
	return result
}

var plOnes = []string{
	"", "jeden", "dwa", "trzy", "cztery", "pięć", "sześć", "siedem", "osiem", "dziewięć",
	"dziesięć", "jedenaście", "dwanaście", "trzynaście", "czternaście", "piętnaście",
	"szesnaście", "siedemnaście", "osiemnaście", "dziewiętnaście",
}

// female forms for 1 and 2 (used for thousands which are feminine)
var plOnesFemale = [3]string{"", "jedna", "dwie"}

var plTens = []string{
	"", "dziesięć", "dwadzieścia", "trzydzieści", "czterdzieści", "pięćdziesiąt",
	"sześćdziesiąt", "siedemdziesiąt", "osiemdziesiąt", "dziewięćdziesiąt",
}

var plHundreds = []string{
	"", "sto", "dwieście", "trzysta", "czterysta", "pięćset",
	"sześćset", "siedemset", "osiemset", "dziewięćset",
}

// plScaleWord returns the correct declension for a scale word.
// Polish has three forms: one (1), few (2-4), many (5+, 0, teens).
func plScaleWord(n int64, one, few, many string) string {
	last2 := n % 100
	last1 := n % 10
	if last2 > 10 && last2 < 20 {
		return many
	}
	switch last1 {
	case 1:
		return one
	case 2, 3, 4:
		return few
	}
	return many
}

// plPositive converts n > 0 to Polish.
// female = true for thousands (tysiąc is feminine).
func plPositive(n int64, female bool) string {
	if n >= 1_000_000_000_000 {
		g := n / 1_000_000_000_000
		rem := n % 1_000_000_000_000
		sw := plScaleWord(g, "bilion", "biliony", "bilionów")
		s := plPositive(g, false) + " " + sw
		if rem > 0 {
			s += " " + plPositive(rem, female)
		}
		return s
	}
	if n >= 1_000_000_000 {
		g := n / 1_000_000_000
		rem := n % 1_000_000_000
		sw := plScaleWord(g, "miliard", "miliardy", "miliardów")
		s := plPositive(g, false) + " " + sw
		if rem > 0 {
			s += " " + plPositive(rem, female)
		}
		return s
	}
	if n >= 1_000_000 {
		g := n / 1_000_000
		rem := n % 1_000_000
		sw := plScaleWord(g, "milion", "miliony", "milionów")
		s := plPositive(g, false) + " " + sw
		if rem > 0 {
			s += " " + plPositive(rem, female)
		}
		return s
	}
	if n >= 1000 {
		g := n / 1000
		rem := n % 1000
		sw := plScaleWord(g, "tysiąc", "tysiące", "tysięcy")
		s := plPositive(g, true) + " " + sw
		if rem > 0 {
			s += " " + plPositive(rem, female)
		}
		return s
	}
	if n >= 100 {
		h := n / 100
		rem := n % 100
		s := plHundreds[h]
		if rem > 0 {
			s += " " + plPositive(rem, female)
		}
		return s
	}
	if n < 20 {
		if female {
			switch n {
			case 1:
				return plOnesFemale[1]
			case 2:
				return plOnesFemale[2]
			}
		}
		return plOnes[n]
	}
	t := n / 10
	o := n % 10
	if o == 0 {
		return plTens[t]
	}
	oStr := plOnes[o]
	if female {
		switch o {
		case 1:
			oStr = plOnesFemale[1]
		case 2:
			oStr = plOnesFemale[2]
		}
	}
	return plTens[t] + " " + oStr
}
