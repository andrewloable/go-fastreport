package functions

import (
	"fmt"
	"math"
	"strings"
	"unicode"
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

// ruCurrencies maps ISO currency codes to their Russian word forms.
var ruCurrencies = map[string]struct {
	male        bool
	s1, s2, sm  string
	j1, j2, jm  string
	jmale       bool
}{
	"RUR":  {true, "рубль", "рубля", "рублей", "копейка", "копейки", "копеек", false},
	"RUB":  {true, "рубль", "рубля", "рублей", "копейка", "копейки", "копеек", false},
	"UAH":  {false, "гривна", "гривны", "гривен", "копейка", "копейки", "копеек", false},
	"EUR":  {true, "евро", "евро", "евро", "евроцент", "евроцента", "евроцентов", true},
	"USD":  {true, "доллар", "доллара", "долларов", "цент", "цента", "центов", true},
	"BYN":  {true, "рубль", "рубля", "рублей", "копейка", "копейки", "копеек", false},
	"BBYN": {true, "белорусский рубль", "белорусских рубля", "белорусских рублей", "белорусская копейка", "белорусских копейки", "белорусских копеек", false},
}

// ConvertCurrencyRu converts a float64 monetary value to Russian words for the given ISO currency code.
// If decimalPartToWord is true, the cents are also expressed in words; otherwise as a numeric "NN" prefix.
func ConvertCurrencyRu(value float64, currencyName string, decimalPartToWord bool) (string, error) {
	cur, ok := ruCurrencies[currencyName]
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
		wholeWords = "ноль"
	} else {
		wholeWords = strings.TrimSpace(ruPositive(n, !cur.male))
	}
	seniorWord := ruScaleWord(n, cur.s1, cur.s2, cur.sm)
	result := wholeWords + " " + seniorWord

	if negative {
		result = "минус " + result
	}

	if cur.j1 != "" {
		var juniorPart string
		if decimalPartToWord {
			var centsWords string
			if cents == 0 {
				centsWords = "ноль"
			} else {
				centsWords = strings.TrimSpace(ruPositive(int64(cents), !cur.jmale))
			}
			juniorWord := ruScaleWord(int64(cents), cur.j1, cur.j2, cur.jm)
			juniorPart = centsWords + " " + juniorWord
		} else {
			juniorWord := ruScaleWord(int64(cents), cur.j1, cur.j2, cur.jm)
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
