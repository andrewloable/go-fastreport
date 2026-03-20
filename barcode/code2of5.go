// code2of5.go implements 2/5 family barcode encoding:
// Interleaved, Industrial, Matrix, ITF-14, Deutsche Identcode, Deutsche Leitcode.
// Ported from C# Barcode2of5.cs.
package barcode

import (
	"fmt"
	"strings"
)

// tabelle25 is the 2-of-5 encoding table: 10 digits × 5 bits.
// 0 = narrow, 1 = wide.
var tabelle25 = [10][5]int{
	{0, 0, 1, 1, 0}, // 0
	{1, 0, 0, 0, 1}, // 1
	{0, 1, 0, 0, 1}, // 2
	{1, 1, 0, 0, 0}, // 3
	{0, 0, 1, 0, 1}, // 4
	{1, 0, 1, 0, 0}, // 5
	{0, 1, 1, 0, 0}, // 6
	{0, 0, 0, 1, 1}, // 7
	{1, 0, 0, 1, 0}, // 8
	{0, 1, 0, 1, 0}, // 9
}

// code2of5setLen pads text with leading zeros to n digits, then truncates to n.
func code2of5setLen(text string, n int) string {
	for len(text) < n {
		text = "0" + text
	}
	if len(text) > n {
		return text[:n]
	}
	return text
}

// deutscheChecksum appends a check digit using Deutsche Post 9/4 weights.
func deutscheChecksum(data string) string {
	sum := 0
	fak := len(data)
	for i := 0; i < len(data); i++ {
		d := int(data[i] - '0')
		if fak%2 == 0 {
			sum += d * 9
		} else {
			sum += d * 4
		}
		fak--
	}
	if sum%10 == 0 {
		return data + "0"
	}
	return data + string(rune('0'+10-sum%10))
}

// interleavedEncode encodes text (must be even-length digits) as interleaved 2/5.
func interleavedEncode(text string) string {
	var sb strings.Builder
	sb.WriteString("5050") // start
	for i := 0; i < len(text)/2; i++ {
		d1 := int(text[i*2] - '0')
		d2 := int(text[i*2+1] - '0')
		if d1 < 0 || d1 > 9 || d2 < 0 || d2 > 9 {
			continue // skip invalid digits
		}
		for j := 0; j < 5; j++ {
			if tabelle25[d1][j] == 1 {
				sb.WriteByte('6') // wide bar
			} else {
				sb.WriteByte('5') // narrow bar
			}
			if tabelle25[d2][j] == 1 {
				sb.WriteByte('1') // wide space
			} else {
				sb.WriteByte('0') // narrow space
			}
		}
	}
	sb.WriteString("605") // stop
	return sb.String()
}

// ── 2/5 Interleaved ──────────────────────────────────────────────────────────

func (b *Code2of5Barcode) GetPattern() (string, error) {
	text := strings.TrimSpace(b.encodedText)
	if text == "" {
		return "", fmt.Errorf("code2of5: empty input")
	}
	// Pad to even length
	if len(text)%2 != 0 {
		text = "0" + text
	}
	return interleavedEncode(text), nil
}

func (b *Code2of5Barcode) GetWideBarRatio() float32 { return 2 }

// ── 2/5 Industrial ───────────────────────────────────────────────────────────

func (b *Code2of5IndustrialBarcode) GetPattern() (string, error) {
	text := strings.TrimSpace(b.encodedText)
	if text == "" {
		return "", fmt.Errorf("code2of5industrial: empty input")
	}
	var sb strings.Builder
	sb.WriteString("606050") // start
	for i := 0; i < len(text); i++ {
		d := int(text[i] - '0')
		for j := 0; j < 5; j++ {
			if tabelle25[d][j] == 1 {
				sb.WriteString("60")
			} else {
				sb.WriteString("50")
			}
		}
	}
	sb.WriteString("605060") // stop
	return sb.String(), nil
}

func (b *Code2of5IndustrialBarcode) GetWideBarRatio() float32 { return 2 }

// ── 2/5 Matrix ───────────────────────────────────────────────────────────────

func (b *Code2of5MatrixBarcode) GetPattern() (string, error) {
	text := strings.TrimSpace(b.encodedText)
	if text == "" {
		return "", fmt.Errorf("code2of5matrix: empty input")
	}
	var sb strings.Builder
	sb.WriteString("705050") // start
	for i := 0; i < len(text); i++ {
		d := int(text[i] - '0')
		for j := 0; j < 5; j++ {
			var c byte
			if tabelle25[d][j] == 1 {
				c = '1'
			} else {
				c = '0'
			}
			if j%2 == 0 {
				c += 5 // even positions → bar (add 5: '0'→'5', '1'→'6')
			}
			sb.WriteByte(c)
		}
		sb.WriteByte('0')
	}
	sb.WriteString("70505") // stop
	return sb.String(), nil
}

// GetWideBarRatio returns 2.25 per C# Barcode2of5.cs:487 (WideBarRatio=2.25F for Matrix 2/5).
func (b *Code2of5MatrixBarcode) GetWideBarRatio() float32 { return 2.25 }

// ── ITF-14 ───────────────────────────────────────────────────────────────────

func (b *ITF14Barcode) GetPattern() (string, error) {
	// Pad to 13, append checksum → 14 digits
	text := CheckSumModulo10(code2of5setLen(strings.TrimSpace(b.encodedText), 13))

	var sb strings.Builder
	// 14 leading quiet zone / bearer bar markers
	for range 14 {
		sb.WriteByte('0')
	}
	sb.WriteString("5050") // interleaved start
	for i := 0; i < len(text)/2; i++ {
		d1 := int(text[i*2] - '0')
		d2 := int(text[i*2+1] - '0')
		for j := 0; j < 5; j++ {
			if tabelle25[d1][j] == 1 {
				sb.WriteByte('6')
			} else {
				sb.WriteByte('5')
			}
			if tabelle25[d2][j] == 1 {
				sb.WriteByte('1')
			} else {
				sb.WriteByte('0')
			}
		}
	}
	sb.WriteString("605") // stop
	// 14 trailing quiet zone / bearer bar markers
	for range 14 {
		sb.WriteByte('0')
	}
	return sb.String(), nil
}

// GetWideBarRatio returns 2.25 per C# Barcode2of5.cs:566 (WideBarRatio=2.25F for ITF-14).
func (b *ITF14Barcode) GetWideBarRatio() float32 { return 2.25 }

// ── Deutsche Identcode ───────────────────────────────────────────────────────

func (b *DeutscheIdentcodeBarcode) GetPattern() (string, error) {
	text := strings.ReplaceAll(strings.ReplaceAll(b.encodedText, ".", ""), " ", "")
	switch len(text) {
	case 11:
		text = deutscheChecksum(text)
	case 12:
		// already includes check digit
	default:
		return "", fmt.Errorf("deutsche identcode: expected 11 or 12 digits, got %d", len(text))
	}
	return interleavedEncode(text), nil
}

func (b *DeutscheIdentcodeBarcode) GetWideBarRatio() float32 { return 3 }

// ── Deutsche Leitcode ────────────────────────────────────────────────────────

func (b *DeutscheLeitcodeBarcode) GetPattern() (string, error) {
	text := strings.ReplaceAll(strings.ReplaceAll(b.encodedText, ".", ""), " ", "")
	switch len(text) {
	case 13:
		text = deutscheChecksum(text)
	case 14:
		// already includes check digit
	default:
		return "", fmt.Errorf("deutsche leitcode: expected 13 or 14 digits, got %d", len(text))
	}
	return interleavedEncode(text), nil
}

func (b *DeutscheLeitcodeBarcode) GetWideBarRatio() float32 { return 3 }
