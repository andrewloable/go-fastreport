// postnet.go implements PostNet and Japan Post 4-State barcode encoding.
// Ported from C# BarcodePostNet.cs.
package barcode

import (
	"fmt"
	"regexp"
	"strings"
)

// tabPostNet encodes digits 0-9 into 5-char patterns of '5' (full-height) and '9' (half-height).
// '5' = full bar, '9' = short bar (from C# table where '9'=narrow/half).
var tabPostNet = [10]string{
	"5151919191", // 0
	"9191915151", // 1
	"9191519151", // 2
	"9191515191", // 3
	"9151919151", // 4
	"9151915191", // 5
	"9151519191", // 6
	"5191919151", // 7
	"5191915191", // 8
	"5191519191", // 9
}

func (b *PostNetBarcode) GetPattern() (string, error) {
	result := "51"
	for _, c := range b.encodedText {
		if c < '0' || c > '9' {
			return "", fmt.Errorf("postnet: invalid character %q", c)
		}
		result += tabPostNet[c-'0']
	}
	result += "5"
	return result, nil
}

func (b *PostNetBarcode) GetWideBarRatio() float32 { return 2 }

// ── Japan Post 4-State Code ──────────────────────────────────────────────────

var japanEncodeTable = "1234567890-abcdefgh"
var japanCheckDigitSet = "0123456789-abcdefgh"
var japanTable = [19]string{
	"6161E", // 1
	"61G1F", // 2
	"G161F", // 3
	"61F1G", // 4
	"61E16", // 5
	"G1F16", // 6
	"F161G", // 7
	"F1G16", // 8
	"E1616", // 9
	"61E1E", // 0
	"E161E", // -
	"G1F1E", // a
	"G1E1F", // b
	"F1G1E", // c
	"E1G1F", // d
	"F1E1G", // e
	"E1F1G", // f
	"E1E16", // g
	"61616", // h
}

var japanValidFirstRe = regexp.MustCompile(`^[0-9\-]+$`)
var japanValidRestRe = regexp.MustCompile(`^[A-Z0-9\-]*$`)

func (b *JapanPost4StateBarcode) GetPattern() (string, error) {
	src := b.encodedText
	if len(src) < 7 {
		return "", fmt.Errorf("japan post 4-state: input too short (min 7 chars)")
	}
	if !japanValidFirstRe.MatchString(src[:7]) || !japanValidRestRe.MatchString(src[7:]) {
		return "", fmt.Errorf("japan post 4-state: invalid input characters")
	}

	var encoded strings.Builder
	weight := 0
	for _, c := range src {
		switch {
		case (c >= '0' && c <= '9') || c == '-':
			encoded.WriteRune(c)
			weight++
		case c >= 'A' && c <= 'J':
			encoded.WriteByte('a')
			encoded.WriteByte(byte(c-'A') + '0')
			weight += 2
		case c >= 'K' && c <= 'T':
			encoded.WriteByte('b')
			encoded.WriteByte(byte(c-'K') + '0')
			weight += 2
		case c >= 'U' && c <= 'Z':
			encoded.WriteByte('c')
			encoded.WriteByte(byte(c-'U') + '0')
			weight += 2
		}
	}

	enc := encoded.String()
	// Remove hyphens at positions 3 and 7
	if idx := strings.Index(enc, "-"); idx == 3 {
		enc = enc[:3] + enc[4:]
		weight--
	}
	if idx := strings.Index(enc[5:], "-"); idx == 2 { // position 7 in original
		pos := 5 + idx
		enc = enc[:pos] + enc[pos+1:]
		weight--
	}

	if weight > 20 {
		return "", fmt.Errorf("japan post 4-state: too many encoded characters")
	}

	// Pad to 20 chars with 'd'
	for len(enc) < 20 {
		enc += "d"
	}

	var result strings.Builder
	result.WriteString("61G1") // start bar
	sum := 0
	for i := 0; i < 20; i++ {
		tIdx := strings.IndexByte(japanEncodeTable, enc[i])
		if tIdx < 0 {
			return "", fmt.Errorf("japan post 4-state: unencodable character %q", enc[i])
		}
		result.WriteString(japanTable[tIdx])
		sum += strings.IndexByte(japanCheckDigitSet, enc[i])
		result.WriteByte('1')
	}

	// Check digit
	check := 19 - (sum % 19)
	if check == 19 {
		check = 0
	}
	var checkChar byte
	switch {
	case check <= 9:
		checkChar = byte('0' + check)
	case check == 10:
		checkChar = '-'
	default:
		checkChar = byte('a' + (check - 11))
	}
	cIdx := strings.IndexByte(japanEncodeTable, checkChar)
	result.WriteString(japanTable[cIdx])
	result.WriteString("1G16") // stop bar
	return result.String(), nil
}

func (b *JapanPost4StateBarcode) GetWideBarRatio() float32 { return 2 }
