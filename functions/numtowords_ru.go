package functions

import (
	"fmt"
	"math"
	"strings"
)

// NumToWordsRu converts an integer to its Russian word representation.
// Russian uses grammatical gender for 1 and 2 (тысяча is feminine),
// and three declension forms: одна (1), две/три/четыре (2–4), много (5+, 11–19).
func NumToWordsRu(n int64) string {
	if n == 0 {
		return "ноль"
	}
	if n < 0 {
		return "минус " + NumToWordsRu(-n)
	}
	return strings.TrimSpace(ruPositive(n, false))
}

// NumToWordsRuFloat converts a float64 to Russian words.
func NumToWordsRuFloat(v float64) string {
	whole := int64(math.Abs(v))
	cents := int(math.Round(math.Abs(v)*100)) % 100
	result := NumToWordsRu(whole)
	if v < 0 {
		result = "минус " + result
	}
	if cents > 0 {
		result += fmt.Sprintf(" и %d/100", cents)
	}
	return result
}

// Russian ones (masculine forms by default).
var ruOnes = []string{
	"", "один", "два", "три", "четыре", "пять",
	"шесть", "семь", "восемь", "девять",
	"десять", "одиннадцать", "двенадцать", "тринадцать", "четырнадцать", "пятнадцать",
	"шестнадцать", "семнадцать", "восемнадцать", "девятнадцать",
}

// Female forms for 1 and 2 (тысяча is feminine).
var ruOnesFemale = [3]string{"", "одна", "две"}

var ruTens = []string{
	"", "десять", "двадцать", "тридцать", "сорок", "пятьдесят",
	"шестьдесят", "семьдесят", "восемьдесят", "девяносто",
}

var ruHundreds = []string{
	"", "сто", "двести", "триста", "четыреста", "пятьсот",
	"шестьсот", "семьсот", "восемьсот", "девятьсот",
}

// ruScaleWord returns the correct Russian declension for a scale word.
// Russian has three forms: one (1), few (2–4), many (5+, 0, 11–19).
func ruScaleWord(n int64, one, few, many string) string {
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

// ruPositive converts n > 0 to Russian.
// female = true when expressing thousands (тысяча is a feminine noun).
func ruPositive(n int64, female bool) string {
	if n >= 1_000_000_000_000 {
		g := n / 1_000_000_000_000
		rem := n % 1_000_000_000_000
		sw := ruScaleWord(g, "триллион", "триллиона", "триллионов")
		s := ruPositive(g, false) + " " + sw
		if rem > 0 {
			s += " " + ruPositive(rem, female)
		}
		return s
	}
	if n >= 1_000_000_000 {
		g := n / 1_000_000_000
		rem := n % 1_000_000_000
		sw := ruScaleWord(g, "миллиард", "миллиарда", "миллиардов")
		s := ruPositive(g, false) + " " + sw
		if rem > 0 {
			s += " " + ruPositive(rem, female)
		}
		return s
	}
	if n >= 1_000_000 {
		g := n / 1_000_000
		rem := n % 1_000_000
		sw := ruScaleWord(g, "миллион", "миллиона", "миллионов")
		s := ruPositive(g, false) + " " + sw
		if rem > 0 {
			s += " " + ruPositive(rem, female)
		}
		return s
	}
	if n >= 1000 {
		g := n / 1000
		rem := n % 1000
		sw := ruScaleWord(g, "тысяча", "тысячи", "тысяч")
		s := ruPositive(g, true) + " " + sw // тысяча is feminine
		if rem > 0 {
			s += " " + ruPositive(rem, female)
		}
		return s
	}
	if n >= 100 {
		h := n / 100
		rem := n % 100
		s := ruHundreds[h]
		if rem > 0 {
			s += " " + ruPositive(rem, female)
		}
		return s
	}
	if n < 20 {
		if female {
			switch n {
			case 1:
				return ruOnesFemale[1]
			case 2:
				return ruOnesFemale[2]
			}
		}
		return ruOnes[n]
	}
	t := n / 10
	o := n % 10
	if o == 0 {
		return ruTens[t]
	}
	oStr := ruOnes[o]
	if female {
		switch o {
		case 1:
			oStr = ruOnesFemale[1]
		case 2:
			oStr = ruOnesFemale[2]
		}
	}
	return ruTens[t] + " " + oStr
}
