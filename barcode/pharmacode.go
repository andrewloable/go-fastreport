// pharmacode.go implements Pharmacode barcode encoding.
// Ported from C# BarcodePharmacode.cs.
package barcode

import (
	"fmt"
	"strconv"
	"strings"
)

// QuietZone controls whether a leading/trailing space is added.
// Default: true (matching C# default).

func (b *PharmacodeBarcode) GetPattern() (string, error) {
	val, err := strconv.ParseUint(strings.TrimSpace(b.encodedText), 10, 64)
	if err != nil {
		return "", fmt.Errorf("pharmacode: invalid input %q: %w", b.encodedText, err)
	}
	val++ // +1 as per spec
	// Convert to binary string, drop leading '1'
	bin := strconv.FormatUint(val, 2)
	if len(bin) > 0 && bin[0] == '1' {
		bin = bin[1:]
	}

	const space = "2"
	var sb strings.Builder
	sb.WriteString(space) // leading quiet zone

	for _, c := range bin {
		switch c {
		case '0':
			sb.WriteByte('5')
			sb.WriteString(space)
		case '1':
			sb.WriteByte('7')
			sb.WriteString(space)
		}
	}
	return sb.String(), nil
}

func (b *PharmacodeBarcode) GetWideBarRatio() float32 { return 2 }
