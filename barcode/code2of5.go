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
	// C# Barcode2of5Interleaved.GetPattern (Barcode2of5.cs:39-49):
	// CalcCheckSum=true: strip first digit if even-length, then append mod-10 checksum.
	// CalcCheckSum=false: pad to even length with leading zero.
	if b.CalcChecksum {
		if len(text)%2 == 0 {
			text = text[1:] // strip first digit to make odd
		}
		text = CheckSumModulo10(text) // appends check digit → even length
	} else {
		if len(text)%2 != 0 {
			text = "0" + text
		}
	}
	return interleavedEncode(text), nil
}

// GetWideBarRatio returns the effective wide bar ratio, clamped to [2, 3].
// C# Barcode2of5Interleaved: ratioMin=2, ratioMax=3; default WideBarRatio=2 (LinearBarcodeBase.cs:636).
func (b *Code2of5Barcode) GetWideBarRatio() float32 { return b.clampedWBR(2) }

// ── 2/5 Industrial ───────────────────────────────────────────────────────────

func (b *Code2of5IndustrialBarcode) GetPattern() (string, error) {
	text := strings.TrimSpace(b.encodedText)
	if text == "" {
		return "", fmt.Errorf("code2of5industrial: empty input")
	}
	// C# Barcode2of5Industrial.GetPattern (Barcode2of5.cs:501-504):
	// appends Modulo-10 checksum only when CalcCheckSum is true.
	if b.CalcChecksum {
		text = CheckSumModulo10(text)
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

// GetWideBarRatio returns the effective wide bar ratio, clamped to [2, 3].
// C# Barcode2of5Industrial inherits Barcode2of5Interleaved: ratioMin=2, ratioMax=3 (Barcode2of5.cs:78-79).
func (b *Code2of5IndustrialBarcode) GetWideBarRatio() float32 { return b.clampedWBR(2) }

// ── 2/5 Matrix ───────────────────────────────────────────────────────────────

func (b *Code2of5MatrixBarcode) GetPattern() (string, error) {
	text := strings.TrimSpace(b.encodedText)
	if text == "" {
		return "", fmt.Errorf("code2of5matrix: empty input")
	}
	// C# Barcode2of5Matrix.GetPattern (Barcode2of5.cs:533-536):
	// appends Modulo-10 checksum only when CalcCheckSum is true.
	if b.CalcChecksum {
		text = CheckSumModulo10(text)
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

// GetWideBarRatio returns the effective wide bar ratio, clamped to [2.25, 3.0].
// C# Barcode2of5Matrix constructor: ratioMin=2.25, ratioMax=3.0, WideBarRatio=2.25 (Barcode2of5.cs:564-566).
func (b *Code2of5MatrixBarcode) GetWideBarRatio() float32 { return b.clampedWBR(2.25) }

// ── ITF-14 ───────────────────────────────────────────────────────────────────

func (b *ITF14Barcode) GetPattern() (string, error) {
	// C# BarcodeITF14.GetPattern (Barcode2of5.cs:367-371):
	// when CalcCheckSum=true: DoCheckSumming(text, 14) (pad to 13 + checksum)
	// when CalcCheckSum=false: SetLen(14) (pad/truncate to 14, no checksum)
	var text string
	if b.CalcChecksum {
		text = CheckSumModulo10(code2of5setLen(strings.TrimSpace(b.encodedText), 13))
	} else {
		text = code2of5setLen(strings.TrimSpace(b.encodedText), 14)
	}

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

// GetWideBarRatio returns the effective wide bar ratio, clamped to [2.25, 3.0].
// C# BarcodeITF14 constructor: ratioMin=2.25, ratioMax=3.0, WideBarRatio=2.25 (Barcode2of5.cs:485-487).
func (b *ITF14Barcode) GetWideBarRatio() float32 { return b.clampedWBR(2.25) }

// ── Deutsche Identcode ───────────────────────────────────────────────────────

// insertAt inserts s into text at byte position pos.
// Mirrors C# string.Insert(pos, s).
func insertAt(text, s string, pos int) string {
	if pos < 0 {
		pos = 0
	}
	if pos > len(text) {
		pos = len(text)
	}
	return text[:pos] + s + text[pos:]
}

func (b *DeutscheIdentcodeBarcode) GetPattern() (string, error) {
	text := strings.ReplaceAll(strings.ReplaceAll(b.encodedText, ".", ""), " ", "")
	// C# BarcodeDeutscheIdentcode.GetPattern (Barcode2of5.cs:122-134):
	// CalcCheckSum=true + 11 digits → compute checksum → 12 digits
	// CalcCheckSum=false → require exactly 12 digits (with pre-computed check digit)
	if b.CalcChecksum {
		switch len(text) {
		case 11:
			text = deutscheChecksum(text)
		case 12:
			// already includes check digit
		default:
			return "", fmt.Errorf("deutsche identcode: expected 11 or 12 digits, got %d", len(text))
		}
	} else {
		if len(text) != 12 {
			return "", fmt.Errorf("deutsche identcode: expected 12 digits (CalcCheckSum=false), got %d", len(text))
		}
	}
	pattern := interleavedEncode(text)

	// Format display text: insert dots and spaces as per Deutsche Post spec.
	// C# BarcodeDeutscheIdentcode.GetPattern (Barcode2of5.cs:156-163):
	//   base.text = text.Insert(2,".").Insert(6," ").Insert(10,".")
	//   if !PrintCheckSum { strip last char } else { insert space at pos 14 }
	display := insertAt(text, ".", 2) // "XX.XXXXXXXXX"
	display = insertAt(display, " ", 6) // "XX.XXX XXXXXX"
	display = insertAt(display, ".", 10) // "XX.XXX XXX.XX"
	if !b.PrintCheckSum {
		// strip the last character (the check digit)
		display = display[:len(display)-1]
	} else {
		display = insertAt(display, " ", 14) // "XX.XXX XXX.XX X"
	}
	b.encodedText = display

	return pattern, nil
}

// GetWideBarRatio returns the effective wide bar ratio, clamped to [2.25, 3.5].
// C# BarcodeDeutscheIdentcode constructor: ratioMin=2.25, ratioMax=3.5, WideBarRatio=3.0 (Barcode2of5.cs:191-193).
func (b *DeutscheIdentcodeBarcode) GetWideBarRatio() float32 { return b.clampedWBR(3) }

// ── Deutsche Leitcode ────────────────────────────────────────────────────────

func (b *DeutscheLeitcodeBarcode) GetPattern() (string, error) {
	text := strings.ReplaceAll(strings.ReplaceAll(b.encodedText, ".", ""), " ", "")
	// C# BarcodeDeutscheLeitcode.GetPattern (Barcode2of5.cs:267-279):
	// CalcCheckSum=true + 13 digits → compute checksum → 14 digits
	// CalcCheckSum=false → require exactly 14 digits
	if b.CalcChecksum {
		switch len(text) {
		case 13:
			text = deutscheChecksum(text)
		case 14:
			// already includes check digit
		default:
			return "", fmt.Errorf("deutsche leitcode: expected 13 or 14 digits, got %d", len(text))
		}
	} else {
		if len(text) != 14 {
			return "", fmt.Errorf("deutsche leitcode: expected 14 digits (CalcCheckSum=false), got %d", len(text))
		}
	}
	pattern := interleavedEncode(text)

	// Format display text: insert dots and spaces as per Deutsche Post spec.
	// C# BarcodeDeutscheLeitcode.GetPattern (Barcode2of5.cs:301-308):
	//   base.text = text.Insert(5,".").Insert(6," ").Insert(10,".").Insert(11," ")
	//               .Insert(15,".").Insert(16," ").Insert(19," ")
	// Note: unlike Identcode, Leitcode does not apply PrintCheckSum trimming here.
	display := insertAt(text, ".", 5)   // after 5th digit
	display = insertAt(display, " ", 6) // space after dot
	display = insertAt(display, ".", 10)
	display = insertAt(display, " ", 11)
	display = insertAt(display, ".", 15)
	display = insertAt(display, " ", 16)
	display = insertAt(display, " ", 19)
	b.encodedText = display

	return pattern, nil
}

// GetWideBarRatio returns the effective wide bar ratio, clamped to [2.25, 3.5].
// C# BarcodeDeutscheLeitcode constructor: ratioMin=2.25, ratioMax=3.5, WideBarRatio=3.0 (Barcode2of5.cs:318-320).
func (b *DeutscheLeitcodeBarcode) GetWideBarRatio() float32 { return b.clampedWBR(3) }
