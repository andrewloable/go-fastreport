// gs1.go implements GS1-128 barcode encoding (formerly EAN-128 / UCC-128).
// GS1 DataBar variants remain as stubs (require DrawLineBars renderer).
// Ported from C# BarcodeEAN128.cs + GS1Helper.cs.
package barcode

import "strings"

// gs1AICode describes one GS1 Application Identifier.
type gs1AICode struct {
	numAI         string
	numAILength   int
	minDataLength int
	maxDataLength int
	useFNC1       bool
}

// gs1AICodes is the full table of GS1 Application Identifiers.
var gs1AICodes = []gs1AICode{
	{"00", 2, 18, 18, false},
	{"01", 2, 14, 14, false},
	{"02", 2, 14, 14, false},
	{"10", 2, 1, 20, true},
	{"11", 2, 6, 6, false},
	{"12", 2, 6, 6, false},
	{"13", 2, 6, 6, false},
	{"15", 2, 6, 6, false},
	{"16", 2, 6, 6, false},
	{"17", 2, 6, 6, false},
	{"20", 2, 2, 2, false},
	{"21", 2, 1, 20, true},
	{"22", 2, 1, 20, true},
	{"240", 3, 1, 30, true},
	{"241", 3, 1, 30, true},
	{"242", 3, 1, 6, true},
	{"243", 3, 1, 20, true},
	{"250", 3, 1, 30, true},
	{"251", 3, 1, 30, true},
	{"253", 3, 13, 30, true},
	{"254", 3, 1, 20, true},
	{"255", 3, 13, 25, true},
	{"30", 2, 1, 8, true},
	{"31XX", 4, 6, 6, false},
	{"32XX", 4, 6, 6, false},
	{"33XX", 4, 6, 6, false},
	{"34XX", 4, 6, 6, false},
	{"35XX", 4, 6, 6, false},
	{"36XX", 4, 6, 6, false},
	{"37", 2, 1, 8, true},
	{"390X", 4, 1, 15, true},
	{"391X", 4, 3, 18, true},
	{"392X", 4, 1, 15, true},
	{"393X", 4, 3, 18, true},
	{"394X", 4, 4, 4, true},
	{"400", 3, 1, 30, true},
	{"401", 3, 1, 30, true},
	{"402", 3, 17, 17, true},
	{"403", 3, 1, 30, true},
	{"41X", 3, 13, 13, false},
	{"420", 3, 1, 20, true},
	{"421", 3, 3, 12, true},
	{"422", 3, 3, 3, true},
	{"423", 3, 3, 15, true},
	{"424", 3, 3, 3, true},
	{"425", 3, 3, 15, true},
	{"426", 3, 3, 3, true},
	{"7001", 4, 13, 13, true},
	{"7002", 4, 1, 30, true},
	{"7003", 4, 10, 10, true},
	{"7004", 4, 1, 4, true},
	{"7005", 4, 1, 12, true},
	{"7006", 4, 6, 6, true},
	{"7007", 4, 6, 12, true},
	{"7008", 4, 1, 3, true},
	{"7009", 4, 1, 10, true},
	{"7010", 4, 1, 2, true},
	{"7020", 4, 1, 20, true},
	{"7021", 4, 1, 20, true},
	{"7022", 4, 1, 20, true},
	{"7023", 4, 1, 30, true},
	{"703X", 4, 3, 30, true},
	{"710", 3, 1, 20, true},
	{"711", 3, 1, 20, true},
	{"712", 3, 1, 20, true},
	{"713", 3, 1, 20, true},
	{"71X", 3, 1, 20, true},
	{"8001", 4, 14, 14, true},
	{"8002", 4, 1, 20, true},
	{"8003", 4, 14, 30, true},
	{"8004", 4, 1, 30, true},
	{"8005", 4, 6, 6, true},
	{"8006", 4, 18, 18, true},
	{"8007", 4, 1, 34, true},
	{"8008", 4, 8, 12, true},
	{"8010", 4, 1, 30, true},
	{"8011", 4, 1, 12, true},
	{"8012", 4, 1, 20, true},
	{"8013", 4, 1, 30, true},
	{"8017", 4, 18, 18, true},
	{"8018", 4, 18, 18, true},
	{"8019", 4, 1, 10, true},
	{"8020", 4, 1, 25, true},
	{"8110", 4, 1, 70, true},
	{"8111", 4, 4, 4, true},
	{"8112", 4, 1, 70, true},
	{"8200", 4, 1, 70, true},
	{"90", 2, 1, 30, true},
	{"9X", 2, 1, 90, true},
}

// gs1FindAIIndex finds the AI table entry index for code starting at index.
// Returns -1 if not found.
func gs1FindAIIndex(code string, index int) int {
	if index < 0 || index >= len(code) || code[index] != '(' {
		return -1
	}
	codeLen := len(code) - index
	if codeLen < 3 {
		return -1
	}
	for i, ai := range gs1AICodes {
		maxLen := ai.numAILength
		if maxLen > codeLen {
			continue
		}
		matched := true
		for j := 0; j < maxLen; j++ {
			if ai.numAI[j] != 'X' && ai.numAI[j] != code[index+j+1] {
				matched = false
				break
			}
		}
		if matched {
			return i
		}
	}
	return -1
}

// gs1GetCode parses one AI+data segment from code starting at *index.
// Returns the encoded segment and advances *index.
func gs1GetCode(code string, index int) (string, int) {
	foundAI := gs1FindAIIndex(code, index)
	if foundAI < 0 {
		return "", index
	}
	ai := gs1AICodes[foundAI]
	index += ai.numAILength + 1 // skip '(' + AI digits
	if index >= len(code) || code[index] != ')' {
		return "", index
	}
	index++ // skip ')'
	codeLen := len(code) - index

	aiDigits := code[index-ai.numAILength-1 : index-1] // the AI number string

	if !ai.useFNC1 && codeLen >= ai.maxDataLength {
		result := aiDigits + code[index:index+ai.maxDataLength]
		return result, index + ai.maxDataLength
	} else if ai.useFNC1 {
		maxLen := codeLen
		nextParen := strings.IndexByte(code[index:], '(')
		if nextParen >= 0 {
			maxLen = nextParen
		}
		if maxLen < 0 {
			maxLen = codeLen
		}
		if maxLen >= ai.minDataLength && maxLen <= ai.maxDataLength {
			result := aiDigits + code[index:index+maxLen]
			if maxLen < codeLen {
				result += "&1;" // FNC1 separator between AIs
			}
			return result, index + maxLen
		}
	}
	return "", index
}

// gs1ParseGS1 converts GS1 AI-formatted input "(AI)data(AI)data..." into
// a Code128 control string with FNC1 (&1;) separators.
// Returns "" on parse failure.
func gs1ParseGS1(code string) string {
	if len(code) < 3 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("&1;")
	i := 0
	for i < len(code) {
		seg, newI := gs1GetCode(code, i)
		if seg == "" {
			return "" // parse failed
		}
		sb.WriteString(seg)
		i = newI
	}
	return sb.String()
}

// ── GS1-128 ──────────────────────────────────────────────────────────────────

func (b *GS1_128Barcode) GetPattern() (string, error) {
	text := b.encodedText
	var msg string
	if strings.HasPrefix(text, "(") {
		parsed := gs1ParseGS1(text)
		if parsed != "" {
			msg = "&C;" + parsed
		} else {
			// Fall back: strip parentheses, prepend FNC1
			stripped := strings.ReplaceAll(strings.ReplaceAll(text, ")", ""), " ", "")
			if len(stripped) > 3 {
				stripped = stripped[3:] // remove first AI prefix "(XX" equivalent
			}
			msg = "&C;&1;" + strings.ReplaceAll(stripped, "(", "&A;")
		}
	} else {
		clean := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(text, "(", "&A;"), ")", ""), " ", "")
		msg = "&C;&1;" + clean
	}
	return code128GetPattern(msg)
}

func (b *GS1_128Barcode) GetWideBarRatio() float32 { return 1 }
