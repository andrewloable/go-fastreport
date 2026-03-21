// code128.go implements Code128 (A, B, C, Auto) barcode encoding.
// Ported from C# Barcode128.cs.
package barcode

import (
	"fmt"
	"strings"
)

// code128Entry holds the three code-page characters and the bar pattern for one symbol.
type code128Entry struct {
	a, b, c string
	data    string
}

// tabelle128 is the Code128 character table.
// Indices 103–105 are Start A/B/C; indices 96–102 are specials (FNC/SHIFT/CODE).
// Index 106 is the stop code (no entry needed here — appended directly).
var tabelle128 = [...]code128Entry{
	{" ", " ", "00", "212222"},   // 0
	{"!", "!", "01", "222122"},   // 1
	{"\"", "\"", "02", "222221"}, // 2
	{"#", "#", "03", "121223"},   // 3
	{"$", "$", "04", "121322"},   // 4
	{"%", "%", "05", "131222"},   // 5
	{"&", "&", "06", "122213"},   // 6
	{"'", "'", "07", "122312"},   // 7
	{"(", "(", "08", "132212"},   // 8
	{")", ")", "09", "221213"},   // 9
	{"*", "*", "10", "221312"},   // 10
	{"+", "+", "11", "231212"},   // 11
	{",", ",", "12", "112232"},   // 12
	{"-", "-", "13", "122132"},   // 13
	{".", ".", "14", "122231"},   // 14
	{"/", "/", "15", "113222"},   // 15
	{"0", "0", "16", "123122"},   // 16
	{"1", "1", "17", "123221"},   // 17
	{"2", "2", "18", "223211"},   // 18
	{"3", "3", "19", "221132"},   // 19
	{"4", "4", "20", "221231"},   // 20
	{"5", "5", "21", "213212"},   // 21
	{"6", "6", "22", "223112"},   // 22
	{"7", "7", "23", "312131"},   // 23
	{"8", "8", "24", "311222"},   // 24
	{"9", "9", "25", "321122"},   // 25
	{":", ":", "26", "321221"},   // 26
	{";", ";", "27", "312212"},   // 27
	{"<", "<", "28", "322112"},   // 28
	{"=", "=", "29", "322211"},   // 29
	{">", ">", "30", "212123"},   // 30
	{"?", "?", "31", "212321"},   // 31
	{"@", "@", "32", "232121"},   // 32
	{"A", "A", "33", "111323"},   // 33
	{"B", "B", "34", "131123"},   // 34
	{"C", "C", "35", "131321"},   // 35
	{"D", "D", "36", "112313"},   // 36
	{"E", "E", "37", "132113"},   // 37
	{"F", "F", "38", "132311"},   // 38
	{"G", "G", "39", "211313"},   // 39
	{"H", "H", "40", "231113"},   // 40
	{"I", "I", "41", "231311"},   // 41
	{"J", "J", "42", "112133"},   // 42
	{"K", "K", "43", "112331"},   // 43
	{"L", "L", "44", "132131"},   // 44
	{"M", "M", "45", "113123"},   // 45
	{"N", "N", "46", "113321"},   // 46
	{"O", "O", "47", "133121"},   // 47
	{"P", "P", "48", "313121"},   // 48
	{"Q", "Q", "49", "211331"},   // 49
	{"R", "R", "50", "231131"},   // 50
	{"S", "S", "51", "213113"},   // 51
	{"T", "T", "52", "213311"},   // 52
	{"U", "U", "53", "213131"},   // 53
	{"V", "V", "54", "311123"},   // 54
	{"W", "W", "55", "311321"},   // 55
	{"X", "X", "56", "331121"},   // 56
	{"Y", "Y", "57", "312113"},   // 57
	{"Z", "Z", "58", "312311"},   // 58
	{"[", "[", "59", "332111"},   // 59
	{"\\", "\\", "60", "314111"}, // 60
	{"]", "]", "61", "221411"},   // 61
	{"^", "^", "62", "431111"},   // 62
	{"_", "_", "63", "111224"},   // 63
	{"\x00", "`", "64", "111422"},  // 64
	{"\x01", "a", "65", "121124"},  // 65
	{"\x02", "b", "66", "121421"},  // 66
	{"\x03", "c", "67", "141122"},  // 67
	{"\x04", "d", "68", "141221"},  // 68
	{"\x05", "e", "69", "112214"},  // 69
	{"\x06", "f", "70", "112412"},  // 70
	{"\x07", "g", "71", "122114"},  // 71
	{"\x08", "h", "72", "122411"},  // 72
	{"\x09", "i", "73", "142112"},  // 73
	{"\x0A", "j", "74", "142211"},  // 74
	{"\x0B", "k", "75", "241211"},  // 75
	{"\x0C", "l", "76", "221114"},  // 76
	{"\x0D", "m", "77", "413111"},  // 77
	{"\x0E", "n", "78", "241112"},  // 78
	{"\x0F", "o", "79", "134111"},  // 79
	{"\x10", "p", "80", "111242"},  // 80
	{"\x11", "q", "81", "121142"},  // 81
	{"\x12", "r", "82", "121241"},  // 82
	{"\x13", "s", "83", "114212"},  // 83
	{"\x14", "t", "84", "124112"},  // 84
	{"\x15", "u", "85", "124211"},  // 85
	{"\x16", "v", "86", "411212"},  // 86
	{"\x17", "w", "87", "421112"},  // 87
	{"\x18", "x", "88", "421211"},  // 88
	{"\x19", "y", "89", "212141"},  // 89
	{"\x1A", "z", "90", "214121"},  // 90
	{"\x1B", "{", "91", "412121"},  // 91
	{"\x1C", "|", "92", "111143"},  // 92
	{"\x1D", "}", "93", "111341"},  // 93
	{"\x1E", "~", "94", "131141"},  // 94
	{"\x1F", "\x7F", "95", "114113"}, // 95
	{" ", " ", "96", "114311"},   // 96 FNC3
	{" ", " ", "97", "411113"},   // 97 FNC2
	{" ", " ", "98", "411311"},   // 98 SHIFT
	{" ", " ", "99", "113141"},   // 99 CODE C
	{" ", " ", "  ", "114131"},   // 100 FNC4 / CODE B
	{" ", " ", "  ", "311141"},   // 101 FNC4 / CODE A
	{" ", " ", "  ", "411131"},   // 102 FNC1
	{" ", " ", "  ", "211412"},   // 103 START A
	{" ", " ", "  ", "211214"},   // 104 START B
	{" ", " ", "  ", "211232"},   // 105 START C
}

type c128Encoding int

const (
	c128None  c128Encoding = iota
	c128A
	c128B
	c128C
	c128AorB
)

func c128FindA(ch byte) int {
	for i, e := range tabelle128 {
		if len(e.a) > 0 && e.a[0] == ch {
			return i
		}
	}
	return -1
}

func c128FindB(ch byte) int {
	for i, e := range tabelle128 {
		if len(e.b) > 0 && e.b[0] == ch {
			return i
		}
	}
	return -1
}

func c128FindC(s string) int {
	for i, e := range tabelle128 {
		if e.c == s {
			return i
		}
	}
	return -1
}

func c128IsDigit(b byte) bool { return b >= '0' && b <= '9' }

func c128CountDigits(code string, idx int) int {
	n := 0
	for idx+n < len(code) && c128IsDigit(code[idx+n]) {
		n++
	}
	return n
}

// c128StripControlCodes removes &A;, &B;, &C;, &S; tags (and optionally FNC codes).
func c128StripControlCodes(code string, stripFN bool) string {
	var b strings.Builder
	i := 0
	for i < len(code) {
		tok, adv := c128GetNextChar(code, i, c128None)
		i += adv
		switch tok {
		case "&A;", "&B;", "&C;", "&S;":
			// always strip
		case "&1;", "&2;", "&3;", "&4;":
			if !stripFN {
				b.WriteString(tok)
			}
		default:
			b.WriteString(tok)
		}
	}
	return b.String()
}

// c128GetNextChar returns the next logical token and how many bytes to advance.
func c128GetNextChar(code string, idx int, enc c128Encoding) (tok string, adv int) {
	if idx >= len(code) {
		return "", 0
	}
	// Check for &X; control codes
	if code[idx] == '&' && idx+2 < len(code) && code[idx+2] == ';' {
		c := code[idx+1]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		switch c {
		case 'A', 'B', 'C', 'S', '1', '2', '3', '4':
			return "&" + string(c) + ";", 3
		}
	}
	// Code C: grab two digits
	if enc == c128C && idx+1 < len(code) {
		return code[idx : idx+2], 2
	}
	return code[idx : idx+1], 1
}

// c128AutoEncode inserts &A;/&B;/&C; control codes into a plain string.
// Mirrors C# Barcode128.Encode().
func c128AutoEncode(code string) string {
	// First strip any existing control codes (keep FNC)
	code = c128StripControlCodes(code, false)
	var sb strings.Builder
	idx := 0
	enc := c128None
	for idx < len(code) {
		portion, newIdx, newEnc := c128NextPortion(code, idx, enc)
		sb.WriteString(portion)
		idx = newIdx
		enc = newEnc
	}
	return sb.String()
}

// c128NextPortion returns (portionText, newIdx, newEncoding)
func c128NextPortion(code string, idx int, enc c128Encoding) (string, int, c128Encoding) {
	if idx >= len(code) {
		return "", idx, enc
	}

	aIdx := c128FindA(code[idx])
	bIdx := c128FindB(code[idx])
	firstEnc := c128A
	if aIdx == -1 && bIdx != -1 {
		firstEnc = c128B
	} else if aIdx != -1 && bIdx != -1 {
		firstEnc = c128AorB
	}

	// 4+ digits → use C encoding
	numDigits := c128CountDigits(code, idx)
	if numDigits >= 4 {
		numDigits = (numDigits / 2) * 2
		result := "&C;" + code[idx:idx+numDigits]
		return result, idx + numDigits, c128C
	}

	// Scan same-encoding characters
	numChars := 1
	for idx+numChars < len(code) {
		ai := c128FindA(code[idx+numChars])
		bi := c128FindB(code[idx+numChars])
		nextEnc := c128A
		if ai == -1 && bi != -1 {
			nextEnc = c128B
		} else if ai != -1 && bi != -1 {
			nextEnc = c128AorB
		}
		if c128CountDigits(code, idx+numChars) >= 4 {
			nextEnc = c128C
		}
		if nextEnc != c128C && nextEnc != firstEnc {
			if firstEnc == c128AorB {
				firstEnc = nextEnc
			} else if nextEnc == c128AorB {
				nextEnc = firstEnc
			}
		}
		if firstEnc != nextEnc {
			break
		}
		numChars++
	}

	if firstEnc == c128AorB {
		firstEnc = c128B
	}

	prefix := "&A;"
	newEnc := enc
	if firstEnc == c128B {
		prefix = "&B;"
	}
	// Use SHIFT for a single char swap between A and B
	if enc != firstEnc && numChars == 1 &&
		(enc == c128A || enc == c128B) &&
		(firstEnc == c128A || firstEnc == c128B) {
		prefix = "&S;"
	} else {
		newEnc = firstEnc
	}

	return prefix + code[idx:idx+numChars], idx + numChars, newEnc
}

// doConvert converts a bar-width string (digit chars) into the pattern char encoding
// used by DrawLinearBarcode.  C# DoConvert: even positions add 5, all positions subtract 1.
// '1'→'5', '2'→'6', ... '4'→'8'  (black bars) for even positions
// '1'→'0', '2'→'1', ...           (white spaces) for odd positions
func doConvert(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		v := int(s[i]) - 1
		if i%2 == 0 {
			v += 5
		}
		b[i] = byte(v)
	}
	return string(b)
}

// code128GetPattern produces the DrawLinearBarcode pattern for a Code128 message.
// msg must already contain &A;/&B;/&C; prefix tokens.
func code128GetPattern(msg string) (string, error) {
	enc := c128None
	idx := 0

	// Determine start code from first token
	tok, adv := c128GetNextChar(msg, idx, enc)
	idx += adv

	var checksum int
	var result strings.Builder

	switch tok {
	case "&A;":
		enc = c128A
		checksum = 103
		result.WriteString(tabelle128[103].data)
	case "&B;":
		enc = c128B
		checksum = 104
		result.WriteString(tabelle128[104].data)
	case "&C;":
		enc = c128C
		checksum = 105
		result.WriteString(tabelle128[105].data)
	default:
		return "", fmt.Errorf("code128: message must start with &A;, &B; or &C; (got %q)", tok)
	}

	codewordPos := 1
	for idx < len(msg) {
		tok, adv = c128GetNextChar(msg, idx, enc)
		idx += adv

		var symIdx int
		switch tok {
		case "&A;":
			enc = c128A
			symIdx = 101
		case "&B;":
			enc = c128B
			symIdx = 100
		case "&C;":
			enc = c128C
			symIdx = 99
		case "&S;":
			if enc == c128A {
				enc = c128B
			} else {
				enc = c128A
			}
			symIdx = 98
		case "&1;":
			symIdx = 102
		case "&2;":
			symIdx = 97
		case "&3;":
			symIdx = 96
		case "&4;":
			if enc == c128A {
				symIdx = 101
			} else {
				symIdx = 100
			}
		default:
			switch enc {
			case c128A:
				symIdx = c128FindA(tok[0])
			case c128B:
				symIdx = c128FindB(tok[0])
			case c128C:
				symIdx = c128FindC(tok)
			}
		}

		if symIdx < 0 {
			return "", fmt.Errorf("code128: invalid character %q for encoding", tok)
		}

		result.WriteString(tabelle128[symIdx].data)
		checksum += symIdx * codewordPos
		codewordPos++

		// Switch back after SHIFT
		if tok == "&S;" {
			if enc == c128A {
				enc = c128B
			} else {
				enc = c128A
			}
		}
	}

	checksum %= 103
	result.WriteString(tabelle128[checksum].data)
	result.WriteString("2331112") // stop code

	return doConvert(result.String()), nil
}

// ── Code128Barcode (Auto) ────────────────────────────────────────────────────

func (b *Code128Barcode) GetPattern() (string, error) {
	var msg string
	if b.AutoEncode {
		msg = c128AutoEncode(b.encodedText)
	} else {
		// Manual mode: use text as-is (caller must embed &A;/&B;/&C; codes).
		msg = b.encodedText
	}
	return code128GetPattern(msg)
}

// GetWideBarRatio returns 2 per C# LinearBarcodeBase.cs:636 (WideBarRatio = 2).
func (b *Code128Barcode) GetWideBarRatio() float32 { return 2 }

// ── Code128ABarcode ──────────────────────────────────────────────────────────

func (b *Code128ABarcode) GetPattern() (string, error) {
	msg := "&A;" + c128StripControlCodes(b.encodedText, false)
	return code128GetPattern(msg)
}

// GetWideBarRatio returns 2 per C# LinearBarcodeBase.cs:636.
func (b *Code128ABarcode) GetWideBarRatio() float32 { return 2 }

// ── Code128BBarcode ──────────────────────────────────────────────────────────

func (b *Code128BBarcode) GetPattern() (string, error) {
	msg := "&B;" + c128StripControlCodes(b.encodedText, false)
	return code128GetPattern(msg)
}

// GetWideBarRatio returns 2 per C# LinearBarcodeBase.cs:636.
func (b *Code128BBarcode) GetWideBarRatio() float32 { return 2 }

// ── Code128CBarcode ──────────────────────────────────────────────────────────

func (b *Code128CBarcode) GetPattern() (string, error) {
	msg := "&C;" + b.encodedText
	return code128GetPattern(msg)
}

// GetWideBarRatio returns 2 per C# LinearBarcodeBase.cs:636.
func (b *Code128CBarcode) GetWideBarRatio() float32 { return 2 }
