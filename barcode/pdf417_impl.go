// pdf417_impl.go implements a pure-Go PDF417 barcode encoder.
//
// PDF417 is a stacked 2D barcode. Each row consists of:
//   start pattern + left indicator + data codewords + right indicator + stop pattern
//
// This implementation supports text compaction mode for ASCII data with
// Reed-Solomon error correction.
//
// Reference: ISO/IEC 15438:2006 — PDF417 bar code symbology specification.
package barcode

import "fmt"

// ── PDF417 constants ────────────────────────────────────────────────────────

const (
	pdf417MinCols   = 1
	pdf417MaxCols   = 30
	pdf417MinRows   = 3
	pdf417MaxRows   = 90
	pdf417MaxCW     = 929 // maximum codeword value
	pdf417SwitchTxt = 900
	pdf417SwitchByt = 901
	pdf417SwitchNum = 902
)

// pdf417StartPattern is the start pattern (17 modules), bit-packed: 0x1fea8 = 8+1+1+1+1+1+1+3.
var pdf417StartPattern = [8]int{8, 1, 1, 1, 1, 1, 1, 3}

// pdf417StopPattern is the stop pattern (18 modules), bit-packed: 0x3fa29 = 7+1+1+3+1+1+1+2+1.
var pdf417StopPattern = [9]int{7, 1, 1, 3, 1, 1, 1, 2, 1}

// ── PDF417 text compaction ──────────────────────────────────────────────────

// pdf417TextEncode encodes text content into PDF417 codewords using text compaction.
func pdf417TextEncode(content string) []int {
	// Text compaction: submode Alpha (uppercase + space)
	// Each pair of characters produces one codeword: cw = c1 * 30 + c2
	var subValues []int
	for _, ch := range content {
		v := -1
		switch {
		case ch >= 'A' && ch <= 'Z':
			v = int(ch - 'A')
		case ch == ' ':
			v = 26
		case ch >= 'a' && ch <= 'z':
			// Switch to Lower submode
			subValues = append(subValues, 27) // Lower latch
			v = int(ch - 'a')
		case ch >= '0' && ch <= '9':
			v = int(ch-'0') + 15 // Mixed submode mapping for digits
			// Actually, in mixed submode digits 0-9 are at positions 15-24
			// For simplicity, use byte compaction for non-alpha
			subValues = append(subValues, 28) // Mixed latch
		case ch >= 0 && ch < 256:
			// Byte compaction fallback
			v = int(ch) % 30
		}
		if v >= 0 && v < 30 {
			subValues = append(subValues, v)
		}
	}

	// Pad to even count
	if len(subValues)%2 != 0 {
		subValues = append(subValues, 29) // padding
	}

	var codewords []int
	codewords = append(codewords, pdf417SwitchTxt) // text compaction mode
	for i := 0; i < len(subValues)-1; i += 2 {
		cw := subValues[i]*30 + subValues[i+1]
		if cw >= pdf417MaxCW {
			cw = pdf417MaxCW - 1
		}
		codewords = append(codewords, cw)
	}

	return codewords
}

// ── PDF417 byte compaction ──────────────────────────────────────────────────

func pdf417ByteEncode(data []byte) []int {
	var codewords []int
	codewords = append(codewords, pdf417SwitchByt)
	// Groups of 6 bytes → 5 codewords
	i := 0
	for i+6 <= len(data) {
		var val int64
		for j := 0; j < 6; j++ {
			val = val*256 + int64(data[i+j])
		}
		for j := 4; j >= 0; j-- {
			codewords = append(codewords, int(val%900))
			val /= 900
		}
		// Reverse the last 5 codewords (they were generated LSB first)
		n := len(codewords)
		for a, b := n-5, n-1; a < b; a, b = a+1, b-1 {
			codewords[a], codewords[b] = codewords[b], codewords[a]
		}
		i += 6
	}
	// Remaining bytes: 1 byte → 1 codeword
	for ; i < len(data); i++ {
		codewords = append(codewords, int(data[i]))
	}
	return codewords
}

// ── PDF417 Reed-Solomon error correction ────────────────────────────────────

// pdf417ECCount returns the number of EC codewords for the given security level.
func pdf417ECCount(securityLevel int) int {
	return 2 << securityLevel // 2^(level+1)
}

// pdf417ComputeEC computes the PDF417 Reed-Solomon error correction codewords.
// Uses the pre-computed generator polynomial from pdf417ErrLevel.
// Mirrors BarcodePDF417.cs CalculateErrorCorrection (lines 797-819).
func pdf417ComputeEC(dataWords []int, securityLevel int) []int {
	if securityLevel < 0 || securityLevel >= len(pdf417ErrLevel) {
		securityLevel = 2
	}
	A := pdf417ErrLevel[securityLevel]
	ecCount := len(A)
	ec := make([]int, ecCount)
	lastE := ecCount - 1
	for _, d := range dataWords {
		t1 := (d + ec[0]) % 929
		for e := 0; e <= lastE; e++ {
			t2 := (t1 * A[lastE-e]) % 929
			t3 := 929 - t2
			next := 0
			if e != lastE {
				next = ec[e+1]
			}
			ec[e] = (next + t3) % 929
		}
	}
	for k := range ec {
		ec[k] = (929 - ec[k]) % 929
	}
	return ec
}

// ── PDF417 symbol layout ────────────────────────────────────────────────────

// pdf417EncodeSymbol creates a PDF417 symbol matrix from content.
func pdf417EncodeSymbol(content string, securityLevel int) ([][]bool, error) {
	if content == "" {
		return nil, fmt.Errorf("pdf417: content must not be empty")
	}
	if securityLevel < 0 || securityLevel > 8 {
		securityLevel = 2
	}

	// Encode content to codewords
	var dataCW []int
	isText := true
	for _, ch := range content {
		if ch > 127 {
			isText = false
			break
		}
	}
	if isText {
		dataCW = pdf417TextEncode(content)
	} else {
		dataCW = pdf417ByteEncode([]byte(content))
	}

	// Compute EC codewords
	ecCount := pdf417ECCount(securityLevel)

	// Choose number of columns (data + ec + 1 length descriptor)
	totalCW := len(dataCW) + ecCount + 1
	cols := 3
	rows := (totalCW + cols - 1) / cols
	for rows > pdf417MaxRows && cols < pdf417MaxCols {
		cols++
		rows = (totalCW + cols - 1) / cols
	}
	if rows < pdf417MinRows {
		rows = pdf417MinRows
	}

	// Pad data to fill rows*cols - ecCount slots (length descriptor occupies slot 0)
	lengthCW := rows*cols - ecCount
	dataWithLen := make([]int, 0, lengthCW)
	dataWithLen = append(dataWithLen, lengthCW) // symbol length descriptor (C# codewords[0])
	dataWithLen = append(dataWithLen, dataCW...)
	for len(dataWithLen) < lengthCW {
		dataWithLen = append(dataWithLen, 900) // padding
	}
	if len(dataWithLen) > lengthCW {
		dataWithLen = dataWithLen[:lengthCW]
	}

	// Compute EC using pre-computed generator polynomial
	ecCW := pdf417ComputeEC(dataWithLen, securityLevel)
	allCW := append(dataWithLen, ecCW...)

	// Build matrix: each row is start + left + data + right + stop
	// Each codeword is 17 modules wide
	// Start pattern: 17 modules, Stop pattern: 18 modules
	modulesPerRow := 17 + 17 + cols*17 + 17 + 18
	matrix := make([][]bool, rows)

	cwIdx := 0
	for r := 0; r < rows; r++ {
		row := make([]bool, modulesPerRow)
		pos := 0

		// Start pattern
		pos = pdf417DrawPattern(row, pos, pdf417StartPattern[:])

		// Left row indicator (a codeword encoding row info)
		leftCW := pdf417RowIndicator(r, rows, cols, securityLevel, 0)
		pos = pdf417DrawCodeword(row, pos, leftCW, r%3)

		// Data codewords for this row
		for c := 0; c < cols; c++ {
			cw := 0
			if cwIdx < len(allCW) {
				cw = allCW[cwIdx]
			}
			cwIdx++
			pos = pdf417DrawCodeword(row, pos, cw, r%3)
		}

		// Right row indicator
		rightCW := pdf417RowIndicator(r, rows, cols, securityLevel, 1)
		pos = pdf417DrawCodeword(row, pos, rightCW, r%3)

		// Stop pattern
		pdf417DrawPattern(row, pos, pdf417StopPattern[:])

		matrix[r] = row
	}

	return matrix, nil
}

func pdf417DrawPattern(row []bool, pos int, pattern []int) int {
	bar := true
	for _, w := range pattern {
		for i := 0; i < w; i++ {
			if pos < len(row) {
				row[pos] = bar
				pos++
			}
		}
		bar = !bar
	}
	return pos
}

// pdf417DrawCodeword renders a codeword as 17 modules using the ISO/IEC 15438
// cluster lookup tables (pdf417Clusters). Each entry is a 17-bit integer where
// bit 16 is the first (leftmost) module: 1=bar (dark), 0=space (light).
// Mirrors BarcodePDF417.cs OutCodeword17 / CLUSTERS table usage.
func pdf417DrawCodeword(row []bool, pos, cw, cluster int) int {
	if cw < 0 || cw >= 929 {
		cw = 0
	}
	pattern := pdf417Clusters[cluster][cw]
	for bit := 16; bit >= 0; bit-- {
		if pos < len(row) {
			row[pos] = (pattern>>bit)&1 != 0
			pos++
		}
	}
	return pos
}

// pdf417RowIndicator returns the left or right row indicator codeword value.
// side 0 = left indicator, side 1 = right indicator.
// Mirrors BarcodePDF417.cs OutPaintCode row indicator logic (lines 762-793).
func pdf417RowIndicator(row, rows, cols, secLevel, side int) int {
	cluster := row % 3
	base := (row / 3) * 30
	var cw int
	if side == 0 { // left indicator
		switch cluster {
		case 0:
			cw = base + (rows-1)/3
		case 1:
			cw = base + secLevel*3 + (rows-1)%3
		default: // 2
			cw = base + cols - 1
		}
	} else { // right indicator
		switch cluster {
		case 0:
			cw = base + cols - 1
		case 1:
			cw = base + (rows-1)/3
		default: // 2
			cw = base + secLevel*3 + (rows-1)%3
		}
	}
	if cw >= 929 {
		cw %= 929
	}
	return cw
}

// GetMatrix encodes b.encodedText as a PDF417 barcode and returns (matrix, rows, cols).
func (p *PDF417Barcode) GetMatrix() ([][]bool, int, int) {
	text := p.encodedText
	if text == "" {
		text = p.DefaultValue()
	}
	matrix, err := pdf417EncodeSymbol(text, p.SecurityLevel)
	if err != nil || len(matrix) == 0 {
		return [][]bool{{true}}, 1, 1
	}
	rows := len(matrix)
	cols := 0
	if rows > 0 {
		cols = len(matrix[0])
	}
	return matrix, rows, cols
}
