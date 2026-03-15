package functions

// NumToLetters converts a non-negative integer to a letter sequence using
// the Excel-style column labelling scheme:
//
//	0 → "A",  1 → "B", ..., 25 → "Z", 26 → "AA", 27 → "AB", ...
//
// Returns an empty string for negative values.
func NumToLetters(n int) string {
	return numToLettersWith(n, 'A')
}

// NumToLettersLower is the lowercase variant: 0 → "a", 1 → "b", ..., 26 → "aa".
func NumToLettersLower(n int) string {
	return numToLettersWith(n, 'a')
}

// numToLettersWith is the shared implementation ported from NumToLettersBase.Str.
// base is either 'A' (uppercase) or 'a' (lowercase).
func numToLettersWith(n int, base rune) string {
	if n < 0 {
		return ""
	}
	const radix = 26
	buf := make([]byte, 0, 8)
	for {
		letter := n % radix
		buf = append([]byte{byte(base) + byte(letter)}, buf...)
		n /= radix
		if n == 0 {
			break
		}
		n-- // adjust for the non-zero-based nature of the sequence
	}
	return string(buf)
}
