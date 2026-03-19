// ean.go implements EAN-8 and EAN-13 barcode encoding.
// Ported from C# BarcodeEAN.cs.
package barcode

// EAN character sets A (left odd), B (left even), C (right).
// Each entry is 4-char string: two bar-width digits followed by two space-width digits.
// Actually the pattern encoding uses 4 chars: L1 S1 L2 S2 for A/B and S1 L1 S2 L2 for C.
var eanTableA = [10]string{
	"2605", "1615", "1516", "0805", "0526", "0625", "0508", "0706", "0607", "2506",
}

var eanTableB = [10]string{
	"0517", "0616", "1606", "0535", "1705", "0715", "3505", "1525", "2515", "1507",
}

var eanTableC = [10]string{
	"7150", "6160", "6061", "5350", "5071", "5170", "5053", "5251", "5152", "7051",
}

// ean13Parity selects A/B encoding for the 6 left-half digits based on the first digit.
var ean13Parity = [10][6]byte{
	{'A', 'A', 'A', 'A', 'A', 'A'}, // 0
	{'A', 'A', 'B', 'A', 'B', 'B'}, // 1
	{'A', 'A', 'B', 'B', 'A', 'B'}, // 2
	{'A', 'A', 'B', 'B', 'B', 'A'}, // 3
	{'A', 'B', 'A', 'A', 'B', 'B'}, // 4
	{'A', 'B', 'B', 'A', 'A', 'B'}, // 5
	{'A', 'B', 'B', 'B', 'A', 'A'}, // 6
	{'A', 'B', 'A', 'B', 'A', 'B'}, // 7
	{'A', 'B', 'A', 'B', 'B', 'A'}, // 8
	{'A', 'B', 'B', 'A', 'B', 'A'}, // 9
}

// eanDigit converts a character '0'-'9' to int.
func eanDigit(c byte) int { return int(c - '0') }

// eanSetLen pads or truncates to exactly n chars, zero-filled on the left.
func eanSetLen(text string, n int) string {
	for len(text) < n {
		text = "0" + text
	}
	return text[:n]
}

// eanChecksum computes the mod-10 check digit (same as CheckSumModulo10).
func eanChecksum(data string) string {
	return CheckSumModulo10(data)
}

// ── EAN-8 ─────────────────────────────────────────────────────────────────────

func (b *EAN8Barcode) GetPattern() (string, error) {
	text := eanSetLen(b.encodedText, 7)
	text = eanChecksum(text) // now 8 chars

	result := "A0A" // start guard
	for i := 0; i < 4; i++ {
		result += eanTableA[eanDigit(text[i])]
	}
	result += "0A0A0" // centre guard
	for i := 4; i < 8; i++ {
		result += eanTableC[eanDigit(text[i])]
	}
	result += "A0A" // stop guard
	return result, nil
}

func (b *EAN8Barcode) GetWideBarRatio() float32 { return 2 }

// ── EAN-13 ────────────────────────────────────────────────────────────────────

func (b *EAN13Barcode) GetPattern() (string, error) {
	text := eanSetLen(b.encodedText, 12)
	text = eanChecksum(text) // now 13 chars

	lk := eanDigit(text[0])
	body := text[1:] // 12 digits

	result := "A0A" // start guard
	for i := 0; i < 6; i++ {
		d := eanDigit(body[i])
		switch ean13Parity[lk][i] {
		case 'A':
			result += eanTableA[d]
		case 'B':
			result += eanTableB[d]
		}
	}
	result += "0A0A0" // centre guard
	for i := 6; i < 12; i++ {
		result += eanTableC[eanDigit(body[i])]
	}
	result += "A0A" // stop guard
	return result, nil
}

func (b *EAN13Barcode) GetWideBarRatio() float32 { return 2 }
