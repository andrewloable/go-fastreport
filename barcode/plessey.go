// plessey.go implements Plessey barcode encoding.
// Ported from C# BarcodePlessey.cs.
package barcode

import (
	"fmt"
	"strings"
)

// tabellePlessey maps hex-digit index (0–15) to its DrawLinearBarcode pattern string.
var tabellePlessey = [16]string{
	"500500500500", // 0
	"60500500500",  // 1
	"50060500500",  // 2
	"6060500500",   // 3
	"50050060500",  // 4
	"6050060500",   // 5
	"5006060500",   // 6
	"606060500",    // 7
	"50050050060",  // 8
	"6050050060",   // 9
	"5006050060",   // A / 10
	"606050060",    // B / 11
	"5005006060",   // C / 12
	"605006060",    // D / 13
	"500606060",    // E / 14
	"60606060",     // F / 15
}

const plesseyAlphabet = "0123456789ABCDEF"

var plesseyCRCGrid = [9]byte{1, 1, 1, 1, 0, 1, 0, 0, 1}

func (b *PlesseyBarcode) GetPattern() (string, error) {
	text := strings.ToUpper(b.encodedText)

	// Validate and collect nibble indices
	indices := make([]int, len(text))
	for i, c := range text {
		idx := strings.IndexRune(plesseyAlphabet, c)
		if idx < 0 {
			return "", fmt.Errorf("plessey: invalid character %q", c)
		}
		indices[i] = idx
	}

	// Build CRC buffer: 4 bits per character (LSB first) + 8 extra for CRC result
	crcBuffer := make([]byte, 4*len(text)+8)
	crcPos := 0
	for _, idx := range indices {
		crcBuffer[crcPos] = byte(idx & 1)
		crcBuffer[crcPos+1] = byte((idx >> 1) & 1)
		crcBuffer[crcPos+2] = byte((idx >> 2) & 1)
		crcBuffer[crcPos+3] = byte((idx >> 3) & 1)
		crcPos += 4
	}

	// ZXing CRC computation: XOR with polynomial for each set bit
	for i := 0; i < 4*len(text); i++ {
		if crcBuffer[i] != 0 {
			for j := 0; j < 9; j++ {
				crcBuffer[i+j] ^= plesseyCRCGrid[j]
			}
		}
	}

	// Build pattern: start + data + 8 CRC bits + end
	var sb strings.Builder
	sb.WriteString("606050060") // start
	for _, idx := range indices {
		sb.WriteString(tabellePlessey[idx])
	}
	// Append 8 CRC bits: 0→"500", 1→"60"
	base := len(text) * 4
	for i := 0; i < 8; i++ {
		if crcBuffer[base+i] == 0 {
			sb.WriteString("500")
		} else {
			sb.WriteString("60")
		}
	}
	sb.WriteString("70050050606") // end
	return sb.String(), nil
}

func (b *PlesseyBarcode) GetWideBarRatio() float32 { return 2 }
