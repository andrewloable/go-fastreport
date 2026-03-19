// upc.go implements UPC-A, UPC-E0, UPC-E1, and Supplement 2/5 barcode encoding.
// Ported from C# BarcodeUPC.cs.
package barcode

import "strings"

// tabUPCE0 is the UPC-E parity pattern table (for both number systems 0 and 1).
// 'E' = EAN-B encoding (= character-wise reversal of EAN-C), 'o' = EAN-A encoding.
// For UPC-E1, the sense is inverted: 'E'→EAN-A, 'o'→EAN-B.
var tabUPCE0 = [10][6]byte{
	{'E', 'E', 'E', 'o', 'o', 'o'}, // 0
	{'E', 'E', 'o', 'E', 'o', 'o'}, // 1
	{'E', 'E', 'o', 'o', 'E', 'o'}, // 2
	{'E', 'E', 'o', 'o', 'o', 'E'}, // 3
	{'E', 'o', 'E', 'E', 'o', 'o'}, // 4
	{'E', 'o', 'o', 'E', 'E', 'o'}, // 5
	{'E', 'o', 'o', 'o', 'E', 'E'}, // 6
	{'E', 'o', 'E', 'o', 'E', 'o'}, // 7
	{'E', 'o', 'E', 'o', 'o', 'E'}, // 8
	{'E', 'o', 'o', 'E', 'o', 'E'}, // 9
}

// upcMakeLong converts bar chars '5'-'8' to long-bar chars 'A'-'D'.
// Used for UPC-A guard bars that extend below the data bars.
func upcMakeLong(code string) string {
	b := []byte(code)
	for i, c := range b {
		if c >= '5' && c <= '8' {
			b[i] = 'A' + (c - '5')
		}
	}
	return string(b)
}

// suppChecksum computes the supplemental barcode check digit.
// Weights alternate 9/3 from the leftmost digit.
func suppChecksum(text string) byte {
	sum := 0
	fak := len(text)
	for i := 0; i < len(text); i++ {
		d := int(text[i] - '0')
		if fak%2 == 0 {
			sum += d * 9
		} else {
			sum += d * 3
		}
		fak--
	}
	return byte('0' + sum%10)
}

// ── UPC-A ─────────────────────────────────────────────────────────────────────

func (b *UPCABarcode) GetPattern() (string, error) {
	text := eanSetLen(b.encodedText, 11)
	text = CheckSumModulo10(text) // now 12 digits

	var sb strings.Builder
	sb.WriteString("A0A") // start guard
	for i := 0; i < 6; i++ {
		code := eanTableA[eanDigit(text[i])]
		if i == 0 {
			code = upcMakeLong(code) // first digit has long guard bar
		}
		sb.WriteString(code)
	}
	sb.WriteString("0A0A0") // centre guard
	for i := 6; i < 12; i++ {
		code := eanTableC[eanDigit(text[i])]
		if i == 11 {
			code = upcMakeLong(code) // last digit has long guard bar
		}
		sb.WriteString(code)
	}
	sb.WriteString("A0A") // stop guard
	return sb.String(), nil
}

func (b *UPCABarcode) GetWideBarRatio() float32 { return 2 }

// ── UPC-E0 ────────────────────────────────────────────────────────────────────

func (b *UPCE0Barcode) GetPattern() (string, error) {
	text := eanSetLen(b.encodedText, 6)
	text = CheckSumModulo10(text) // 7 digits; text[6] is check digit
	c := eanDigit(text[6])

	var sb strings.Builder
	sb.WriteString("A0A") // start guard
	for i := 0; i < 6; i++ {
		d := eanDigit(text[i])
		if tabUPCE0[c][i] == 'E' {
			sb.WriteString(eanTableB[d]) // eanTableB is char-wise reversal of eanTableC
		} else {
			sb.WriteString(eanTableA[d])
		}
	}
	sb.WriteString("0A0A0A") // stop guard
	return sb.String(), nil
}

func (b *UPCE0Barcode) GetWideBarRatio() float32 { return 2 }

// ── UPC-E1 ────────────────────────────────────────────────────────────────────

func (b *UPCE1Barcode) GetPattern() (string, error) {
	text := eanSetLen(b.encodedText, 6)
	text = CheckSumModulo10(text) // 7 digits; text[6] is check digit
	c := eanDigit(text[6])

	var sb strings.Builder
	sb.WriteString("A0A") // start guard
	for i := 0; i < 6; i++ {
		d := eanDigit(text[i])
		// UPC-E1 is inverted: 'E'→eanTableA, 'o'→eanTableB
		if tabUPCE0[c][i] == 'E' {
			sb.WriteString(eanTableA[d])
		} else {
			sb.WriteString(eanTableB[d])
		}
	}
	sb.WriteString("0A0A0A") // stop guard
	return sb.String(), nil
}

func (b *UPCE1Barcode) GetWideBarRatio() float32 { return 2 }

// ── Supplement 2 ─────────────────────────────────────────────────────────────

func (b *Supplement2Barcode) GetPattern() (string, error) {
	text := eanSetLen(b.encodedText, 2)
	value := int(text[0]-'0')*10 + int(text[1]-'0')

	var parity [2]byte
	switch value % 4 {
	case 0:
		parity = [2]byte{'o', 'o'}
	case 1:
		parity = [2]byte{'o', 'E'}
	case 2:
		parity = [2]byte{'E', 'o'}
	case 3:
		parity = [2]byte{'E', 'E'}
	}

	var sb strings.Builder
	sb.WriteString("506") // start
	for i := 0; i < 2; i++ {
		d := eanDigit(text[i])
		if parity[i] == 'E' {
			sb.WriteString(eanTableB[d])
		} else {
			sb.WriteString(eanTableA[d])
		}
		if i < 1 {
			sb.WriteString("05") // character delineator
		}
	}
	return sb.String(), nil
}

func (b *Supplement2Barcode) GetWideBarRatio() float32 { return 2 }

// ── Supplement 5 ─────────────────────────────────────────────────────────────

func (b *Supplement5Barcode) GetPattern() (string, error) {
	text := eanSetLen(b.encodedText, 5)
	checkDigit := suppChecksum(text)
	c := eanDigit(checkDigit)

	var sb strings.Builder
	sb.WriteString("506") // start
	for i := 0; i < 5; i++ {
		d := eanDigit(text[i])
		if tabUPCE0[c][1+i] == 'E' {
			sb.WriteString(eanTableB[d])
		} else {
			sb.WriteString(eanTableA[d])
		}
		if i < 4 {
			sb.WriteString("05") // character delineator
		}
	}
	return sb.String(), nil
}

func (b *Supplement5Barcode) GetWideBarRatio() float32 { return 2 }
