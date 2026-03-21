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

// isCodabarStartStop reports whether ch is a valid Codabar start/stop character (A-D).
func isCodabarStartStop(ch byte) bool {
	return ch == 'A' || ch == 'B' || ch == 'C' || ch == 'D'
}

func (b *CodabarBarcode) GetPattern() (string, error) {
	text := strings.ToUpper(b.encodedText)

	// C# BarcodeCodabar.GetPattern() always prepends StartChar and appends
	// StopChar around the data text (BarcodeCodabar.cs:89-113).
	// Use the struct's StartChar/StopChar properties.
	startCh := b.StartChar
	if !isCodabarStartStop(startCh) {
		startCh = 'A'
	}
	stopCh := b.StopChar
	if !isCodabarStartStop(stopCh) {
		stopCh = 'B'
	}

	// Strip any existing start/stop characters from the data text.
	if len(text) >= 2 && isCodabarStartStop(text[0]) && isCodabarStartStop(text[len(text)-1]) {
		text = text[1 : len(text)-1]
	}

	var sb strings.Builder

	// Start character
	if idx := codabarFindItem(string(startCh)); idx >= 0 {
		sb.WriteString(tabelleCB[idx].data)
		sb.WriteByte('0')
	}

	// Data characters
	for i := 0; i < len(text); i++ {
		idx := codabarFindItem(string(text[i]))
		if idx < 0 {
			continue
		}
		sb.WriteString(tabelleCB[idx].data)
		sb.WriteByte('0')
	}

	// Stop character
	if idx := codabarFindItem(string(stopCh)); idx >= 0 {
		sb.WriteString(tabelleCB[idx].data)
	}

	return sb.String(), nil
}

// GetWideBarRatio returns the effective wide bar ratio, clamped to [2, 3].
// C# BarcodeCodabar constructor: ratioMin=2, ratioMax=3 (BarcodeCodabar.cs:141-142).
func (b *CodabarBarcode) GetWideBarRatio() float32 { return b.clampedWBR(2) }
