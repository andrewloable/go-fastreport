package functions

// ── English ───────────────────────────────────────────────────────────────────

var enLetters = []rune{
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
}

var enLowerLetters = []rune{
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
}

// ToLetters converts an integer to an English alphabet string (Excel-style).
// 0 → "A", 25 → "Z", 26 → "AA".
func ToLetters(value any) string {
	return ToLettersEn(value, true)
}

// ToLettersEn converts an integer to an English alphabet string.
func ToLettersEn(value any, isUpper bool) string {
	letters := enLetters
	if !isUpper {
		letters = enLowerLetters
	}
	return str(ToInt(value), letters)
}

// ── Russian ───────────────────────────────────────────────────────────────────

var ruLetters = []rune{
	'а', 'б', 'в', 'г', 'д', 'е', 'ё', 'ж', 'з', 'и', 'й', 'к', 'л', 'м', 'н', 'о', 'п',
	'р', 'с', 'т', 'у', 'ф', 'х', 'ц', 'ч', 'ш', 'щ', 'ъ', 'ы', 'ь', 'э', 'ю', 'я',
}

var ruUpperLetters = []rune{
	'А', 'Б', 'В', 'Г', 'Д', 'Е', 'Ё', 'Ж', 'З', 'И', 'Й', 'К', 'Л', 'М', 'Н', 'О', 'П',
	'Р', 'С', 'Т', 'У', 'Ф', 'Х', 'Ц', 'Ч', 'Ш', 'Щ', 'Ъ', 'Ы', 'Ь', 'Э', 'Ю', 'Я',
}

// ToLettersRu converts an integer to a Russian alphabet string.
func ToLettersRu(value any, isUpper bool) string {
	letters := ruLetters
	if isUpper {
		letters = ruUpperLetters
	}
	return str(ToInt(value), letters)
}

// ── Common implementation ─────────────────────────────────────────────────────

// str is the shared implementation for converting a number to a letter sequence.
// It follows the Excel-style column labelling scheme where:
// 0 → letters[0], 1 → letters[1], ..., radix-1 → letters[radix-1],
// radix → letters[0]+letters[0], etc.
func str(n int, letters []rune) string {
	if n < 0 {
		return ""
	}

	radix := len(letters)
	res := make([]rune, 0, 8)

	for {
		letterIdx := n % radix
		res = append([]rune{letters[letterIdx]}, res...)
		n /= radix
		if n == 0 {
			break
		}
		n--
	}

	return string(res)
}

// ── Legacy / Compatibility ────────────────────────────────────────────────────

// NumToLetters converts a non-negative integer to a letter sequence using
// the Excel-style column labelling scheme.
//
// Deprecated: use ToLetters instead for C# parity.
func NumToLetters(n int) string {
	return ToLettersEn(n, true)
}

// NumToLettersLower is the lowercase variant.
//
// Deprecated: use ToLettersEn(n, false) instead.
func NumToLettersLower(n int) string {
	return ToLettersEn(n, false)
}
