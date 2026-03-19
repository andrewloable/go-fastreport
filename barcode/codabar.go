// codabar.go implements Codabar barcode encoding.
// Ported from C# BarcodeCodabar.cs.
package barcode

import "strings"

type codabarEntry struct {
	c    string
	data string
}

var tabelleCB = []codabarEntry{
	{"1", "5050615"},
	{"2", "5051506"},
	{"3", "6150505"},
	{"4", "5060515"},
	{"5", "6050515"},
	{"6", "5150506"},
	{"7", "5150605"},
	{"8", "5160505"},
	{"9", "6051505"},
	{"0", "5050516"},
	{"-", "5051605"},
	{"$", "5061505"},
	{":", "6050606"},
	{"/", "6060506"},
	{".", "6060605"},
	{"+", "5060606"},
	{"A", "5061515"},
	{"B", "5151506"},
	{"C", "5051516"},
	{"D", "5051615"},
}

func codabarFindItem(c string) int {
	for i, e := range tabelleCB {
		if e.c == c {
			return i
		}
	}
	return -1
}

func (b *CodabarBarcode) GetPattern() (string, error) {
	text := strings.ToUpper(b.encodedText)
	if len(text) < 2 {
		// Ensure at least start and stop chars
		text = "A" + text + "B"
	}

	var sb strings.Builder

	// Start char (first char of encoded text, must be A/B/C/D)
	startChar := string(text[0])
	if idx := codabarFindItem(startChar); idx >= 0 {
		sb.WriteString(tabelleCB[idx].data)
		sb.WriteByte('0')
	}

	// Data chars (everything between start and stop)
	for i := 1; i < len(text)-1; i++ {
		idx := codabarFindItem(string(text[i]))
		if idx < 0 {
			continue
		}
		sb.WriteString(tabelleCB[idx].data)
		sb.WriteByte('0')
	}

	// Stop char (last char of encoded text, must be A/B/C/D)
	stopChar := string(text[len(text)-1])
	if idx := codabarFindItem(stopChar); idx >= 0 {
		sb.WriteString(tabelleCB[idx].data)
	}

	return sb.String(), nil
}

func (b *CodabarBarcode) GetWideBarRatio() float32 { return 2 }
