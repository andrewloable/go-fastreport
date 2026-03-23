package functions

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// NumToWordsUk converts an integer to its Ukrainian word representation.
func NumToWordsUk(n int64) string {
	if n == 0 {
		return "\u043D\u0443\u043B\u044C" // нуль
	}
	if n < 0 {
		return "\u043C\u0456\u043D\u0443\u0441 " + NumToWordsUk(-n) // мінус
	}
	return strings.TrimSpace(ukPositive(n, false))
}

// NumToWordsUkFloat converts a float64 to Ukrainian words.
func NumToWordsUkFloat(v float64) string {
	whole := int64(math.Abs(v))
	cents := int(math.Round(math.Abs(v)*100)) % 100
	result := NumToWordsUk(whole)
	if v < 0 {
		result = "\u043C\u0456\u043D\u0443\u0441 " + result
	}
	if cents > 0 {
		result += fmt.Sprintf(" \u0456 %d/100", cents)
	}
	return result
}

var ukOnes = []string{
	"", "\u043E\u0434\u0438\u043D", "\u0434\u0432\u0430", "\u0442\u0440\u0438",
	"\u0447\u043E\u0442\u0438\u0440\u0438", "\u043F'\u044F\u0442\u044C",
	"\u0448\u0456\u0441\u0442\u044C", "\u0441\u0456\u043C", "\u0432\u0456\u0441\u0456\u043C",
	"\u0434\u0435\u0432'\u044F\u0442\u044C",
	"\u0434\u0435\u0441\u044F\u0442\u044C", "\u043E\u0434\u0438\u043D\u0430\u0434\u0446\u044F\u0442\u044C",
	"\u0434\u0432\u0430\u043D\u0430\u0434\u0446\u044F\u0442\u044C",
	"\u0442\u0440\u0438\u043D\u0430\u0434\u0446\u044F\u0442\u044C",
	"\u0447\u043E\u0442\u0438\u0440\u043D\u0430\u0434\u0446\u044F\u0442\u044C",
	"\u043F'\u044F\u0442\u043D\u0430\u0434\u0446\u044F\u0442\u044C",
	"\u0448\u0456\u0441\u0442\u043D\u0430\u0434\u0446\u044F\u0442\u044C",
	"\u0441\u0456\u043C\u043D\u0430\u0434\u0446\u044F\u0442\u044C",
	"\u0432\u0456\u0441\u0456\u043C\u043D\u0430\u0434\u0446\u044F\u0442\u044C",
	"\u0434\u0435\u0432'\u044F\u0442\u043D\u0430\u0434\u0446\u044F\u0442\u044C",
}

// female forms: одна, дві
var ukOnesFemale = [3]string{"", "\u043E\u0434\u043D\u0430", "\u0434\u0432\u0456"}

var ukTens = []string{
	"", "\u0434\u0435\u0441\u044F\u0442\u044C",
	"\u0434\u0432\u0430\u0434\u0446\u044F\u0442\u044C",
	"\u0442\u0440\u0438\u0434\u0446\u044F\u0442\u044C",
	"\u0441\u043E\u0440\u043E\u043A",
	"\u043F'\u044F\u0442\u0434\u0435\u0441\u044F\u0442",
	"\u0448\u0456\u0441\u0442\u0434\u0435\u0441\u044F\u0442",
	"\u0441\u0456\u043C\u0434\u0435\u0441\u044F\u0442",
	"\u0432\u0456\u0441\u0456\u043C\u0434\u0435\u0441\u044F\u0442",
	"\u0434\u0435\u0432'\u044F\u043D\u043E\u0441\u0442\u043E",
}

var ukHundreds = []string{
	"", "\u0441\u0442\u043E", "\u0434\u0432\u0456\u0441\u0442\u0456",
	"\u0442\u0440\u0438\u0441\u0442\u0430", "\u0447\u043E\u0442\u0438\u0440\u0438\u0441\u0442\u0430",
	"\u043F'\u044F\u0442\u0441\u043E\u0442", "\u0448\u0456\u0441\u0442\u0441\u043E\u0442",
	"\u0441\u0456\u043C\u0441\u043E\u0442", "\u0432\u0456\u0441\u0456\u043C\u0441\u043E\u0442",
	"\u0434\u0435\u0432'\u044F\u0442\u0441\u043E\u0442",
}

// ukCurrencies maps ISO currency codes to their Ukrainian word forms.
var ukCurrencies = map[string]struct {
	male        bool
	s1, s2, sm  string
	j1, j2, jm  string
	jmale       bool
}{
	"RUB": {true, "рубль", "рубля", "рублів", "копійка", "копійки", "копійок", false},
	"UAH": {false, "гривня", "гривні", "гривень", "копійка", "копійки", "копійок", false},
	"EUR": {true, "євро", "євро", "євро", "євроцент", "євроцента", "євроценту", true},
	"USD": {true, "долар", "долара", "доларів", "цент", "цента", "центів", true},
}

// ConvertCurrencyUk converts a float64 monetary value to Ukrainian words for the given ISO currency code.
// If decimalPartToWord is true, the cents are also expressed in words; otherwise as a numeric "NN" prefix.
func ConvertCurrencyUk(value float64, currencyName string, decimalPartToWord bool) (string, error) {
	cur, ok := ukCurrencies[currencyName]
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
		wholeWords = "нуль"
	} else {
		wholeWords = strings.TrimSpace(ukPositive(n, !cur.male))
	}
	seniorWord := ukScaleWord(n, cur.s1, cur.s2, cur.sm)
	result := wholeWords + " " + seniorWord

	if negative {
		result = "мінус " + result
	}

	if cur.j1 != "" {
		var juniorPart string
		if decimalPartToWord {
			var centsWords string
			if cents == 0 {
				centsWords = "нуль"
			} else {
				centsWords = strings.TrimSpace(ukPositive(int64(cents), !cur.jmale))
			}
			juniorWord := ukScaleWord(int64(cents), cur.j1, cur.j2, cur.jm)
			juniorPart = centsWords + " " + juniorWord
		} else {
			juniorWord := ukScaleWord(int64(cents), cur.j1, cur.j2, cur.jm)
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

func ukScaleWord(n int64, one, few, many string) string {
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

func ukPositive(n int64, female bool) string {
	if n >= 1_000_000_000_000 {
		g := n / 1_000_000_000_000
		rem := n % 1_000_000_000_000
		sw := ukScaleWord(g,
			"\u0442\u0440\u0438\u043B\u044C\u0439\u043E\u043D",
			"\u0442\u0440\u0438\u043B\u044C\u0439\u043E\u043D\u0430",
			"\u0442\u0440\u0438\u043B\u044C\u0439\u043E\u043D\u0456\u0432")
		s := ukPositive(g, false) + " " + sw
		if rem > 0 {
			s += " " + ukPositive(rem, female)
		}
		return s
	}
	if n >= 1_000_000_000 {
		g := n / 1_000_000_000
		rem := n % 1_000_000_000
		sw := ukScaleWord(g,
			"\u043C\u0456\u043B\u044C\u044F\u0440\u0434",
			"\u043C\u0456\u043B\u044C\u044F\u0440\u0434\u0430",
			"\u043C\u0456\u043B\u044C\u044F\u0440\u0434\u0456\u0432")
		s := ukPositive(g, false) + " " + sw
		if rem > 0 {
			s += " " + ukPositive(rem, female)
		}
		return s
	}
	if n >= 1_000_000 {
		g := n / 1_000_000
		rem := n % 1_000_000
		sw := ukScaleWord(g,
			"\u043C\u0456\u043B\u044C\u0439\u043E\u043D",
			"\u043C\u0456\u043B\u044C\u0439\u043E\u043D\u0430",
			"\u043C\u0456\u043B\u044C\u0439\u043E\u043D\u0456\u0432")
		s := ukPositive(g, false) + " " + sw
		if rem > 0 {
			s += " " + ukPositive(rem, female)
		}
		return s
	}
	if n >= 1000 {
		g := n / 1000
		rem := n % 1000
		sw := ukScaleWord(g,
			"\u0442\u0438\u0441\u044F\u0447\u0430",
			"\u0442\u0438\u0441\u044F\u0447\u0456",
			"\u0442\u0438\u0441\u044F\u0447")
		s := ukPositive(g, true) + " " + sw
		if rem > 0 {
			s += " " + ukPositive(rem, female)
		}
		return s
	}
	if n >= 100 {
		h := n / 100
		rem := n % 100
		s := ukHundreds[h]
		if rem > 0 {
			s += " " + ukPositive(rem, female)
		}
		return s
	}
	if n < 20 {
		if female {
			switch n {
			case 1:
				return ukOnesFemale[1]
			case 2:
				return ukOnesFemale[2]
			}
		}
		return ukOnes[n]
	}
	t := n / 10
	o := n % 10
	if o == 0 {
		return ukTens[t]
	}
	oStr := ukOnes[o]
	if female {
		switch o {
		case 1:
			oStr = ukOnesFemale[1]
		case 2:
			oStr = ukOnesFemale[2]
		}
	}
	return ukTens[t] + " " + oStr
}
