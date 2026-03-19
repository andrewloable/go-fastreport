// msi.go implements MSI (Modified Plessey) barcode encoding.
// Ported from C# BarcodeMSI.cs.
package barcode

import "strings"

var tabelleMSI = [10]string{
	"51515151", // 0
	"51515160", // 1
	"51516051", // 2
	"51516060", // 3
	"51605151", // 4
	"51605160", // 5
	"51606051", // 6
	"51606060", // 7
	"60515151", // 8
	"60515160", // 9
}

// msiDigitSum returns the sum of all digits of x.
func msiDigitSum(x int) int {
	s := 0
	for x > 0 {
		s += x % 10
		x /= 10
	}
	return s
}

func (b *MSIBarcode) GetPattern() (string, error) {
	text := strings.TrimSpace(b.encodedText)
	if text == "" {
		text = "0"
	}

	var sb strings.Builder
	sb.WriteString("60") // start

	checkEven := 0
	checkOdd := 0
	for i, c := range text {
		d := int(c - '0')
		if i%2 != 0 {
			checkOdd = checkOdd*10 + d
		} else {
			checkEven += d
		}
		sb.WriteString(tabelleMSI[d])
	}

	checksum := msiDigitSum(checkOdd*2) + checkEven
	checksum %= 10
	if checksum > 0 {
		checksum = 10 - checksum
	}
	sb.WriteString(tabelleMSI[checksum])

	sb.WriteString("515") // stop
	return sb.String(), nil
}

func (b *MSIBarcode) GetWideBarRatio() float32 { return 2 }
