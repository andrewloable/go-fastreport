// code39.go implements Code39 and Code39 Extended barcode encoding.
// Ported from C# Barcode39.cs.
package barcode

import "strings"

type code39Entry struct {
	c   string
	data string
	chk int
}

var tabelle39 = []code39Entry{
	{"0", "505160605", 0},
	{"1", "605150506", 1},
	{"2", "506150506", 2},
	{"3", "606150505", 3},
	{"4", "505160506", 4},
	{"5", "605160505", 5},
	{"6", "506160505", 6},
	{"7", "505150606", 7},
	{"8", "605150605", 8},
	{"9", "506150605", 9},
	{"A", "605051506", 10},
	{"B", "506051506", 11},
	{"C", "606051505", 12},
	{"D", "505061506", 13},
	{"E", "605061505", 14},
	{"F", "506061505", 15},
	{"G", "505051606", 16},
	{"H", "605051605", 17},
	{"I", "506051605", 18},
	{"J", "505061605", 19},
	{"K", "605050516", 20},
	{"L", "506050516", 21},
	{"M", "606050515", 22},
	{"N", "505060516", 23},
	{"O", "605060515", 24},
	{"P", "506060515", 25},
	{"Q", "505050616", 26},
	{"R", "605050615", 27},
	{"S", "506050615", 28},
	{"T", "505060615", 29},
	{"U", "615050506", 30},
	{"V", "516050506", 31},
	{"W", "616050505", 32},
	{"X", "515060506", 33},
	{"Y", "615060505", 34},
	{"Z", "516060505", 35},
	{"-", "515050606", 36},
	{".", "615050605", 37},
	{" ", "516050605", 38},
	{"*", "515060605", 0},
	{"$", "515151505", 39},
	{"/", "515150515", 40},
	{"+", "515051515", 41},
	{"%", "505151515", 42},
}

// code39x maps ASCII 0-127 to Code39 Extended two-char pairs.
var code39x = [128]string{
	"%U", "$A", "$B", "$C", "$D", "$E", "$F", "$G",
	"$H", "$I", "$J", "$K", "$L", "$M", "$N", "$O",
	"$P", "$Q", "$R", "$S", "$T", "$U", "$V", "$W",
	"$X", "$Y", "$Z", "%A", "%B", "%C", "%D", "%E",
	" ", "/A", "/B", "/C", "/D", "/E", "/F", "/G",
	"/H", "/I", "/J", "/K", "/L", "/M", "/N", "/O",
	"0", "1", "2", "3", "4", "5", "6", "7",
	"8", "9", "/Z", "%F", "%G", "%H", "%I", "%J",
	"%V", "A", "B", "C", "D", "E", "F", "G",
	"H", "I", "J", "K", "L", "M", "N", "O",
	"P", "Q", "R", "S", "T", "U", "V", "W",
	"X", "Y", "Z", "%K", "%L", "%M", "%N", "%O",
	"%W", "+A", "+B", "+C", "+D", "+E", "+F", "+G",
	"+H", "+I", "+J", "+K", "+L", "+M", "+N", "+O",
	"+P", "+Q", "+R", "+S", "+T", "+U", "+V", "+W",
	"+X", "+Y", "+Z", "%P", "%Q", "%R", "%S", "%T",
}

func code39FindItem(c string) int {
	for i, e := range tabelle39 {
		if e.c == c {
			return i
		}
	}
	return -1
}

// code39GetPattern builds the Code39 pattern string for the given text.
// calcChecksum: true → append mod-43 check character (C# default).
func code39GetPattern(text string, calcChecksum bool) string {
	var sb strings.Builder
	checksum := 0
	starIdx := code39FindItem("*")

	// Start '*'
	sb.WriteString(tabelle39[starIdx].data)
	sb.WriteByte('0')

	for _, c := range strings.ToUpper(text) {
		idx := code39FindItem(string(c))
		if idx < 0 {
			continue
		}
		sb.WriteString(tabelle39[idx].data)
		sb.WriteByte('0')
		checksum += tabelle39[idx].chk
	}

	if calcChecksum {
		checksum = checksum % 43
		for _, e := range tabelle39 {
			if checksum == e.chk {
				sb.WriteString(e.data)
				sb.WriteByte('0')
				break
			}
		}
	}

	// Stop '*'
	sb.WriteString(tabelle39[starIdx].data)
	return sb.String()
}

func (b *Code39Barcode) GetPattern() (string, error) {
	return code39GetPattern(b.encodedText, b.CalcChecksum), nil
}

// GetWideBarRatio returns the effective wide bar ratio, clamped to [2, 3].
// C# Barcode39 constructor: ratioMin=2, ratioMax=3; default WideBarRatio=2 (Barcode39.cs:137-138).
func (b *Code39Barcode) GetWideBarRatio() float32 { return b.clampedWBR(2) }

func (b *Code39ExtendedBarcode) GetPattern() (string, error) {
	var expanded strings.Builder
	for _, c := range b.encodedText {
		if c <= 127 {
			expanded.WriteString(code39x[c])
		}
	}
	return code39GetPattern(expanded.String(), b.CalcChecksum), nil
}

// GetWideBarRatio returns the effective wide bar ratio, clamped to [2, 3].
// C# Barcode39 (base) constructor: ratioMin=2, ratioMax=3 (Barcode39.cs:137-138).
func (b *Code39ExtendedBarcode) GetWideBarRatio() float32 { return b.clampedWBR(2) }
