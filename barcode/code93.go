// code93.go implements Code93 and Code93 Extended barcode encoding.
// Ported from C# Barcode93.cs.
package barcode

import (
	"fmt"
	"strings"
)

type code93Entry struct {
	c    string
	data string
}

var tabelle93 = []code93Entry{
	{"0", "131112"},
	{"1", "111213"},
	{"2", "111312"},
	{"3", "111411"},
	{"4", "121113"},
	{"5", "121212"},
	{"6", "121311"},
	{"7", "111114"},
	{"8", "131211"},
	{"9", "141111"},
	{"A", "211113"},
	{"B", "211212"},
	{"C", "211311"},
	{"D", "221112"},
	{"E", "221211"},
	{"F", "231111"},
	{"G", "112113"},
	{"H", "112212"},
	{"I", "112311"},
	{"J", "122112"},
	{"K", "132111"},
	{"L", "111123"},
	{"M", "111222"},
	{"N", "111321"},
	{"O", "121122"},
	{"P", "131121"},
	{"Q", "212112"},
	{"R", "212211"},
	{"S", "211122"},
	{"T", "211221"},
	{"U", "221121"},
	{"V", "222111"},
	{"W", "112122"},
	{"X", "112221"},
	{"Y", "122121"},
	{"Z", "123111"},
	{"-", "121131"},
	{".", "311112"},
	{" ", "311211"},
	{"$", "321111"},
	{"/", "112131"},
	{"+", "113121"},
	{"%", "211131"},
	{"[", "121221"}, // ($) — extended shift
	{"]", "312111"}, // (%) — extended shift
	{"{", "311121"}, // (/) — extended shift
	{"}", "122211"}, // (+) — extended shift
}

// code93x maps ASCII 0-127 to Code93 Extended encoding pairs.
var code93x = [128]string{
	"]U", "[A", "[B", "[C", "[D", "[E", "[F", "[G",
	"[H", "[I", "[J", "[K", "[L", "[M", "[N", "[O",
	"[P", "[Q", "[R", "[S", "[T", "[U", "[V", "[W",
	"[X", "[Y", "[Z", "]A", "]B", "]C", "]D", "]E",
	" ", "{A", "{B", "{C", "$", "%", "{F", "{G",
	"{H", "{I", "{J", "+", "{L", "-", ".", "/",
	"0", "1", "2", "3", "4", "5", "6", "7",
	"8", "9", "{Z", "]F", "]G", "]H", "]I", "]J",
	"]V", "A", "B", "C", "D", "E", "F", "G",
	"H", "I", "J", "K", "L", "M", "N", "O",
	"P", "Q", "R", "S", "T", "U", "V", "W",
	"X", "Y", "Z", "]K", "]L", "]M", "]N", "]O",
	"]W", "}A", "}B", "}C", "}D", "}E", "}F", "}G",
	"}H", "}I", "}J", "}K", "}L", "}M", "}N", "}O",
	"}P", "}Q", "}R", "}S", "}T", "}U", "}V", "}W",
	"}X", "}Y", "}Z", "]P", "]Q", "]R", "]S", "]T",
}

func code93FindItem(c string) int {
	for i, e := range tabelle93 {
		if e.c == c {
			return i
		}
	}
	return -1
}

// code93GetPattern builds the Code93 pattern for the given (possibly expanded) text.
// Always computes the two check digits C and K per Code93 spec.
func code93GetPattern(text string) (string, error) {
	var sb strings.Builder
	sb.WriteString("111141") // start

	indices := make([]int, 0, len(text))
	for _, c := range text {
		idx := code93FindItem(string(c))
		if idx < 0 {
			return "", fmt.Errorf("code93: invalid character %q", c)
		}
		sb.WriteString(tabelle93[idx].data)
		indices = append(indices, idx)
	}

	// Check digit C (mod-47, weight 1-20 cycling, right-to-left)
	checkC := 0
	weightC := 1
	for i := len(indices) - 1; i >= 0; i-- {
		checkC += indices[i] * weightC
		weightC++
		if weightC > 20 {
			weightC = 1
		}
	}
	checkC = checkC % 47

	// Check digit K (mod-47, weight 2-15 cycling, right-to-left, includes C)
	checkK := 0
	weightK := 2
	for i := len(indices) - 1; i >= 0; i-- {
		checkK += indices[i] * weightK
		weightK++
		if weightK > 15 {
			weightK = 1
		}
	}
	checkK += checkC
	checkK = checkK % 47

	sb.WriteString(tabelle93[checkC].data)
	sb.WriteString(tabelle93[checkK].data)

	// Stop
	sb.WriteString("1111411")

	return doConvert(sb.String()), nil
}

func (b *Code93Barcode) GetPattern() (string, error) {
	return code93GetPattern(b.encodedText)
}

func (b *Code93Barcode) GetWideBarRatio() float32 { return 2 }

func (b *Code93ExtendedBarcode) GetPattern() (string, error) {
	var expanded strings.Builder
	for _, c := range b.encodedText {
		if c <= 127 {
			expanded.WriteString(code93x[c])
		}
	}
	return code93GetPattern(expanded.String())
}

func (b *Code93ExtendedBarcode) GetWideBarRatio() float32 { return 2 }
