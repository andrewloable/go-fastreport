package barcode

// intelligentmail.go — USPS Intelligent Mail Barcode (IMb) full encoder.
//
// Ported from FastReport.Base/Barcode/BarcodeIntelligentMail.cs.
// Reference: http://ribbs.usps.gov/onecodesolution/USPS-B-3200D001.pdf

import (
	"fmt"
	"image"
	"strconv"
	"strings"
)

// ── IMb bar-type constants ────────────────────────────────────────────────────

const (
	imbTracker   = 0 // E — middle third only
	imbAscender  = 1 // F — top two-thirds
	imbDescender = 2 // G — bottom two-thirds
	imbFull      = 3 // 6 — full height
)

// ── IMb lookup tables ─────────────────────────────────────────────────────────

var barTopCharIndexArray = [65]int{
	4, 0, 2, 6, 3, 5, 1, 9, 8, 7,
	1, 2, 0, 6, 4, 8, 2, 9, 5, 3,
	0, 1, 3, 7, 4, 6, 8, 9, 2, 0,
	5, 1, 9, 4, 3, 8, 6, 7, 1, 2,
	4, 3, 9, 5, 7, 8, 3, 0, 2, 1,
	4, 0, 9, 1, 7, 0, 2, 4, 6, 3,
	7, 1, 9, 5, 8,
}

var barBottomCharIndexArray = [65]int{
	7, 1, 9, 5, 8, 0, 2, 4, 6, 3,
	5, 8, 9, 7, 3, 0, 6, 1, 7, 4,
	6, 8, 9, 2, 5, 1, 7, 5, 4, 3,
	8, 7, 6, 0, 2, 5, 4, 9, 3, 0,
	1, 6, 8, 2, 0, 4, 5, 9, 6, 7,
	5, 2, 6, 3, 8, 5, 1, 9, 8, 7,
	4, 0, 2, 6, 3,
}

var barTopCharShiftArray = [65]int{
	3, 0, 8, 11, 1, 12, 8, 11, 10, 6,
	4, 12, 2, 7, 9, 6, 7, 9, 2, 8,
	4, 0, 12, 7, 10, 9, 0, 7, 10, 5,
	7, 9, 6, 8, 2, 12, 1, 4, 2, 0,
	1, 5, 4, 6, 12, 1, 0, 9, 4, 7,
	5, 10, 2, 6, 9, 11, 2, 12, 6, 7,
	5, 11, 0, 3, 2,
}

var barBottomCharShiftArray = [65]int{
	2, 10, 12, 5, 9, 1, 5, 4, 3, 9,
	11, 5, 10, 1, 6, 3, 4, 1, 10, 0,
	2, 11, 8, 6, 1, 12, 3, 8, 6, 4,
	4, 11, 0, 6, 1, 9, 11, 5, 3, 7,
	3, 10, 7, 11, 8, 2, 10, 3, 5, 8,
	0, 3, 12, 11, 8, 4, 5, 1, 3, 0,
	7, 12, 9, 8, 10,
}

const (
	imbTable2Of13Size = 78
	imbTable5Of13Size = 1287
)

// ── N-of-13 table builder ─────────────────────────────────────────────────────

// imbBuildNof13Table builds the N-of-13 lookup table as in OneCodeInitializeNof13Table.
// n is the number of bits set (2 or 5), size is the table size.
func imbBuildNof13Table(n, size int) []int {
	a := make([]int, size+1)
	i1 := 0
	j1 := size - 1
	for k := 0; k <= 8191; k++ {
		// count set bits
		k1 := 0
		for l1 := 0; l1 <= 12; l1++ {
			if (k & (1 << uint(l1))) != 0 {
				k1++
			}
		}
		if k1 == n {
			l := imbMathReverse(k) >> 3
			flag := k == l
			if l >= k {
				if flag {
					a[j1] = k
					j1--
				} else {
					a[i1] = k
					i1++
					a[i1] = l
					i1++
				}
			}
		}
	}
	return a
}

// imbMathReverse reverses the 16-bit representation of i.
func imbMathReverse(i int) int {
	j := 0
	for k := 0; k <= 15; k++ {
		j <<= 1
		j = j | (i & 1)
		i >>= 1
	}
	return j
}

// ── Big-number math (byte-array, 13 bytes big-endian, indices 0–12) ───────────

// imbMathMultiply multiplies the 13-byte big-endian byteArray by j.
// Mirrors C# OneCodeMathMultiply(ref bytearray, 13, j).
// The C# uses i=13 so indices go from 12 down to 1 (stepping -2), then handles index 0.
func imbMathMultiply(byteArray []int, j int) {
	l := 0
	k := 12 // i-1 = 13-1 = 12
	for k >= 1 {
		x := (byteArray[k] | (byteArray[k-1] << 8)) * j + l
		byteArray[k] = x & 255
		byteArray[k-1] = (x >> 8) & 255
		l = x >> 16
		k -= 2
	}
	// k == 0 when i is odd (13 is odd → after pairs [12,11],[10,9],...,[2,1] k becomes 0)
	if k == 0 {
		byteArray[0] = (byteArray[0]*j + l) & 255
	}
}

// imbMathAdd adds j to the 13-byte big-endian byteArray.
// Mirrors C# OneCodeMathAdd(ref bytearray, 13, j).
// The C# implementation: x = (byteArray[i-1] | (byteArray[i-2] << 8)) + j
// i=13 so byteArray[12] and byteArray[11].
func imbMathAdd(byteArray []int, j int) {
	// i = 13, so i-1=12, i-2=11
	x := (byteArray[12] | (byteArray[11] << 8)) + j
	l := x | 65535
	k := 10 // i - 3 = 13 - 3 = 10
	byteArray[12] = x & 255
	byteArray[11] = (x >> 8) & 255
	for l == 1 && k > 0 {
		x = l + byteArray[k]
		byteArray[k] = x & 255
		l = x | 255
		k--
	}
}

// imbMathFcs computes the CRC-11 FCS over the 13-byte big-endian byteArray.
// Mirrors C# OneCodeMathFcs.
func imbMathFcs(byteArray []int) int {
	const c = 3893
	i := 2047
	j := byteArray[0] << 5
	for b := 2; b <= 7; b++ {
		if ((i ^ j) & 1024) != 0 {
			i = (i<<1)^c
		} else {
			i <<= 1
		}
		i = i & 2047
		j <<= 1
	}
	for l := 1; l <= 12; l++ {
		k := byteArray[l] << 3
		for b := 0; b <= 7; b++ {
			if ((i ^ k) & 1024) != 0 {
				i = (i<<1)^c
			} else {
				i <<= 1
			}
			i = i & 2047
			k <<= 1
		}
	}
	return i
}

// ── String-based big-number division ─────────────────────────────────────────

// imbDivide performs the OneCodeMathDivide logic on the decimal string ds.
// It populates codewordVals[1..9] with remainders and codewordVals[0] with the
// final quotient, exactly as the C# does.
// codewordSizes contains the divisor for each codeword (indices 0..9).
func imbDivide(ds string, codewordSizes []int) [10]int {
	var codewordVals [10]int
	n := ds
	// k from 9 down to 1 (j-1 down to 1, j=10)
	for k := 9; k >= 1; k-- {
		divider := codewordSizes[k]
		copy := n
		left := "0"
		l := len(copy)
		r := ""
		i := 1
		for i <= l {
			// find minimum prefix that is >= divider (or end of string)
			divident, _ := strconv.Atoi(copy[:i])
			for divident < divider && i < l-1 {
				r += "0"
				i++
				divident, _ = strconv.Atoi(copy[:i])
			}
			q := divident / divider
			rem := divident % divider
			r += strconv.Itoa(q)
			// format remainder padded to i digits
			left = fmt.Sprintf("%0*d", i, rem)
			copy = left + copy[i:]
			i++
		}
		n = strings.TrimLeft(r, "0")
		if n == "" {
			n = "0"
		}
		leftVal, _ := strconv.Atoi(left)
		codewordVals[k] = leftVal
		if k == 1 {
			rVal, _ := strconv.Atoi(r)
			codewordVals[0] = rVal
		}
	}
	return codewordVals
}

// ── Main encoder ──────────────────────────────────────────────────────────────

// imb_encode encodes a 20/25/29/31-digit IMb string into 65 bar-type values.
// Bar types: 0=Tracker, 1=Ascender, 2=Descender, 3=Full.
func imb_encode(text string) ([]byte, error) {
	// Step 1: strip spaces, dashes, dots.
	text = strings.NewReplacer(" ", "", "-", "", ".", "").Replace(text)

	// Validate: first char 0-9, second char 0-4, then 18 more digits = 20 total,
	// plus optional routing: 5 (zip), 9 (zip+4), 11 (zip+4+dp) digits.
	if len(text) < 20 {
		return nil, fmt.Errorf("intelligentmail: need at least 20 digits, got %d", len(text))
	}
	// Basic validation: all digits
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return nil, fmt.Errorf("intelligentmail: non-digit character %q", ch)
		}
	}
	// Second digit must be 0-4
	if text[1] > '4' {
		return nil, fmt.Errorf("intelligentmail: second digit must be 0-4, got %c", text[1])
	}
	// Length must be 20, 25, 29, or 31
	switch len(text) {
	case 20, 25, 29, 31:
	default:
		return nil, fmt.Errorf("intelligentmail: invalid length %d (must be 20/25/29/31)", len(text))
	}

	// Step 2: Compute routing code l from zip digits (positions 20+).
	var l int64
	zip := text[20:]
	if len(zip) > 0 {
		zipVal, err := strconv.ParseInt(zip, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("intelligentmail: bad zip: %w", err)
		}
		switch len(zip) {
		case 5:
			l = zipVal + 1
		case 9:
			l = zipVal + 100001
		case 11:
			l = zipVal + 1000100001
		}
	}

	// Step 3: Build ds (the decimal string for division) and byteArray.
	// C#: v = l*10 + source[0]; v = v*5 + source[1]; ds = v.ToString() + source[2..19]
	// We replicate this as a string operation for the division step.
	d0 := int64(text[0] - '0')
	d1 := int64(text[1] - '0')
	v := l*10 + d0
	v = v*5 + d1
	ds := strconv.FormatInt(v, 10) + text[2:20]

	// Build byteArray (13 bytes, indices 0-12) from l, then process the 20 digits.
	// C# uses a 14-element array but only indices 0-12 are used by the math fns
	// (i=13 in all calls). Index 13 is never read after initialisation.
	byteArray := make([]int, 13)
	byteArray[12] = int(l & 255)
	byteArray[11] = int((l >> 8) & 255)
	byteArray[10] = int((l >> 16) & 255)
	byteArray[9] = int((l >> 24) & 255)
	byteArray[8] = int((l >> 32) & 255)

	// Multiply/add for first two digits (barcode identifier digits).
	imbMathMultiply(byteArray, 10)
	imbMathAdd(byteArray, int(text[0]-'0'))
	imbMathMultiply(byteArray, 5)
	imbMathAdd(byteArray, int(text[1]-'0'))

	// Process remaining 18 message digits (indices 2-19).
	for i := 2; i <= 19; i++ {
		imbMathMultiply(byteArray, 10)
		imbMathAdd(byteArray, int(text[i]-'0'))
	}

	// Step 4: Compute FCS (CRC-11).
	fcs := imbMathFcs(byteArray)

	// Step 5: Build N-of-13 tables.
	// C# case 1: entries5Of13 = table2Of13Size = 78
	// C# case 2: entries2Of13 = table5Of13Size = 1287
	table2Of13 := imbBuildNof13Table(2, imbTable2Of13Size)
	table5Of13 := imbBuildNof13Table(5, imbTable5Of13Size)
	entries5Of13 := imbTable2Of13Size // = 78
	entries2Of13 := imbTable5Of13Size // = 1287

	// Step 6: Set up codeword sizes.
	// C#: for i=0..9: codewordArray[i][0] = entries2Of13 + entries5Of13
	//     codewordArray[0][0] = 659
	//     codewordArray[9][0] = 636
	totalEntries := entries2Of13 + entries5Of13 // 1287 + 78 = 1365
	codewordSizes := [10]int{}
	for i := 0; i < 10; i++ {
		codewordSizes[i] = totalEntries // 1365
	}
	codewordSizes[0] = 659
	codewordSizes[9] = 636

	// Step 7: Divide ds into 10 codewords using string-based big number division.
	codewordVals := imbDivide(ds, codewordSizes[:])

	// C#: codewordArray[9][1] *= 2
	codewordVals[9] *= 2

	// C#: if (fcs >> 10 != 0) codewordArray[0][1] += 659
	if fcs>>10 != 0 {
		codewordVals[0] += 659
	}

	// Step 8: Map codeword values to 13-bit bar patterns using N-of-13 tables.
	// C#: ad[i][1] = (codewordArray[i][1] >= entries2Of13) ?
	//               table2Of13[codewordArray[i][1] - entries2Of13] :
	//               table5Of13[codewordArray[i][1]]
	barPatterns := [10]int{}
	for i := 0; i < 10; i++ {
		cv := codewordVals[i]
		if cv >= totalEntries {
			return nil, fmt.Errorf("intelligentmail: codeword[%d]=%d out of range", i, cv)
		}
		if cv >= entries2Of13 {
			barPatterns[i] = table2Of13[cv-entries2Of13]
		} else {
			barPatterns[i] = table5Of13[cv]
		}
	}

	// Step 9: Apply FCS bits to flip bar patterns (XOR the lower 13 bits).
	// C#: for i=0..9: if ((fcs & 1<<i) != 0) ad[i][1] = ~(int)ad[i][1] & 8191
	for i := 0; i < 10; i++ {
		if (fcs & (1 << uint(i))) != 0 {
			barPatterns[i] = (^barPatterns[i]) & 8191
		}
	}

	// Step 10: Extract top/bottom bits and combine into 65 bar types.
	// ai[i]  = bit from barPatterns[barTopCharIndexArray[i]]    at shift barTopCharShiftArray[i]
	// ai1[i] = bit from barPatterns[barBottomCharIndexArray[i]] at shift barBottomCharShiftArray[i]
	bars := make([]byte, 65)
	for i := 0; i <= 64; i++ {
		ai := (barPatterns[barTopCharIndexArray[i]] >> uint(barTopCharShiftArray[i])) & 1
		ai1 := (barPatterns[barBottomCharIndexArray[i]] >> uint(barBottomCharShiftArray[i])) & 1
		// C# mapping:
		// ai=0, ai1=0 → 'E' = Tracker   (0)
		// ai=0, ai1=1 → 'G' = Descender (2)
		// ai=1, ai1=0 → 'F' = Ascender  (1)
		// ai=1, ai1=1 → '6' = Full      (3)
		if ai == 0 {
			if ai1 == 0 {
				bars[i] = imbTracker
			} else {
				bars[i] = imbDescender
			}
		} else {
			if ai1 == 0 {
				bars[i] = imbAscender
			} else {
				bars[i] = imbFull
			}
		}
	}
	return bars, nil
}

// ── GetPattern / GetWideBarRatio / Render ────────────────────────────────────

// GetWideBarRatio returns the wide-bar ratio for IMb (default 2, matching C#).
// C# LinearBarcodeBase default WideBarRatio = 2f. IMb bars use modules[1]=2 and
// spaces use modules[2]=3, giving a 2:3 bar-to-space ratio.
func (b *IntelligentMailBarcode) GetWideBarRatio() float32 { return 2 }

// GetPattern returns the DrawLinearBarcode pattern string for the IMb barcode.
// Mirrors C# BarcodeIntelligentMail.Bars(): each bar character ('E'/'F'/'G'/'6')
// is followed by a space character '2', with the trailing space removed.
// QuietZone prepends/appends an extra '2'.
//
// Pattern characters (from C# LinearBarcodeBase.OneBarProps):
//   - 'E' = tracker   (modules[1] wide, middle-third height)
//   - 'F' = ascender  (modules[1] wide, top two-thirds height)
//   - 'G' = descender (modules[1] wide, bottom two-thirds height)
//   - '6' = full      (modules[1] wide, full height)
//   - '2' = space     (modules[2] wide, no bar drawn)
func (b *IntelligentMailBarcode) GetPattern() (string, error) {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, b.encodedText)

	bars, err := imb_encode(digits)
	if err != nil {
		return "", err
	}

	// Build pattern: barChar + '2' for each bar, strip trailing '2'.
	var buf strings.Builder
	for i, barType := range bars {
		switch barType {
		case imbTracker:
			buf.WriteByte('E')
		case imbAscender:
			buf.WriteByte('F')
		case imbDescender:
			buf.WriteByte('G')
		default: // imbFull
			buf.WriteByte('6')
		}
		if i < len(bars)-1 {
			buf.WriteByte('2') // inter-bar space (modules[2] wide)
		}
	}

	s := buf.String()
	if b.QuietZone {
		s = "2" + s + "2"
	}
	return s, nil
}

// Render renders the IMb as a 4-state bar image using DrawLinearBarcode.
// Mirrors C# BarcodeIntelligentMail which inherits LinearBarcodeBase.DrawBarcode.
func (b *IntelligentMailBarcode) Render(width, height int) (image.Image, error) {
	if b.encodedText == "" {
		return nil, fmt.Errorf("intelligentmail: not encoded")
	}
	if width <= 0 {
		width = 130 // 65 bars × 2px each
	}
	if height <= 0 {
		height = 60
	}
	pattern, err := b.GetPattern()
	if err != nil {
		return placeholderImage(width, height), nil
	}
	return DrawLinearBarcode(pattern, b.encodedText, width, height, b.showText, b.GetWideBarRatio()), nil
}
