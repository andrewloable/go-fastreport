// gs1_databar.go implements the GS1 DataBar family of barcode encoders.
//
// Ported from C# BarcodeGS1.cs (FastReport.Base/Barcode/BarcodeGS1.cs).
// Supports four symbologies:
//   - GS1 DataBar Omnidirectional  (BarcodeGS1Omnidirectional)
//   - GS1 DataBar Stacked          (BarcodeGS1Stacked)
//   - GS1 DataBar Stacked Omni     (BarcodeGS1StackedOmnidirectional)
//   - GS1 DataBar Limited          (BarcodeGS1Limited)
package barcode

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"
	"strings"
)

// ── BarcodeType constants ────────────────────────────────────────────────────

const (
	BarcodeTypeGS1DataBarOmni       BarcodeType = "GS1DataBarOmnidirectional"
	BarcodeTypeGS1DataBarStacked    BarcodeType = "GS1DataBarStacked"
	BarcodeTypeGS1DataBarStackedOmni BarcodeType = "GS1DataBarStackedOmnidirectional"
	BarcodeTypeGS1DataBarLimited    BarcodeType = "GS1DataBarLimited"
)

// ── Helper: combinatorics ────────────────────────────────────────────────────

// gs1Combins returns the number of combinations of r selected from n.
// Exact port of C# BarcodeGS1Base.Combins (BarcodeGS1.cs:80-115).
func gs1Combins(n, r int) int {
	var minDenom, maxDenom int
	if n-r > r {
		minDenom = r
		maxDenom = n - r
	} else {
		minDenom = n - r
		maxDenom = r
	}
	val := 1
	j := 1
	for i := n; i > maxDenom; i-- {
		val *= i
		if j <= minDenom {
			val /= j
			j++
		}
	}
	for ; j <= minDenom; j++ {
		val /= j
	}
	return val
}

// gs1GetWidths generates element widths for GS1 DataBar encoding.
// Exact port of C# BarcodeGS1Base.GetGS1Widths (BarcodeGS1.cs:26-72).
//
//   - val:      required value
//   - n:        number of modules
//   - elements: elements in a set (4 for omni/stacked, 7 for limited)
//   - maxWidth: maximum module width of an element
//   - noNarrow: 0 will skip patterns without a 1-module-wide element
func gs1GetWidths(val, n, elements, maxWidth, noNarrow int) []int {
	narrowMask := 0
	widths := make([]int, 0, elements)

	for bar := 0; bar < elements-1; bar++ {
		var elmWidth, subVal int
		// C# for-loop init: narrowMask |= (1 << bar) runs ONCE before the inner loop.
		// C# for-loop post: narrowMask &= ~(1 << bar) runs AFTER each non-breaking iteration.
		narrowMask |= (1 << bar)
		for elmWidth = 1; ; elmWidth++ {
			subVal = gs1Combins(n-elmWidth-1, elements-bar-2)

			// Less combinations with no single-module element.
			if noNarrow == 0 && narrowMask == 0 &&
				n-elmWidth-(elements-bar-1) >= elements-bar-1 {
				subVal -= gs1Combins(n-elmWidth-(elements-bar), elements-bar-2)
			}

			// Less combinations with elements > maxVal.
			if elements-bar-1 > 1 {
				lessVal := 0
				for mxwElement := n - elmWidth - (elements - bar - 2); mxwElement > maxWidth; mxwElement-- {
					lessVal += gs1Combins(n-elmWidth-mxwElement-1, elements-bar-3)
				}
				subVal -= lessVal * (elements - 1 - bar)
			} else if n-elmWidth > maxWidth {
				subVal--
			}

			val -= subVal
			if val < 0 {
				break
			}
			narrowMask &^= (1 << bar)
		}
		val += subVal // restore overshoot
		n -= elmWidth
		widths = append(widths, elmWidth)
	}
	widths = append(widths, n)
	return widths
}

// ── GS1 DataBar Omnidirectional ──────────────────────────────────────────────

// Omnidirectional encoding tables.
// Ported from C# BarcodeGS1Omnidirectional (BarcodeGS1.cs:184-378).
var omniChecksumWeight = [32]int{
	1, 3, 9, 27, 2, 6, 18, 54,
	4, 12, 36, 29, 8, 24, 72, 58,
	16, 48, 65, 37, 32, 17, 51, 74,
	64, 34, 23, 69, 49, 68, 46, 59,
}

var omniFinderPattern = [45]int{
	3, 8, 2, 1, 1,
	3, 5, 5, 1, 1,
	3, 3, 7, 1, 1,
	3, 1, 9, 1, 1,
	2, 7, 4, 1, 1,
	2, 5, 6, 1, 1,
	2, 3, 8, 1, 1,
	1, 5, 7, 1, 1,
	1, 3, 9, 1, 1,
}

var omniModulesOdd = [9]int{12, 10, 8, 6, 4, 5, 7, 9, 11}
var omniModulesEven = [9]int{4, 6, 8, 10, 12, 10, 8, 6, 4}
var omniWidthsOdd = [9]int{8, 6, 4, 3, 1, 2, 4, 6, 8}
var omniWidthsEven = [9]int{1, 3, 5, 6, 8, 7, 5, 3, 1}
var omniGSums = [9]int{0, 161, 961, 2015, 2715, 0, 336, 1036, 1516}
var omniTList = [9]int{1, 10, 34, 70, 126, 4, 20, 48, 81}

// GS1DataBarOmniBarcode implements GS1 DataBar Omnidirectional.
// Ported from C# BarcodeGS1Omnidirectional (BarcodeGS1.cs:183-378).
type GS1DataBarOmniBarcode struct {
	BaseBarcodeImpl
	encodedData  []string
	wideBarRatio float32
}

// NewGS1DataBarOmniBarcode creates a GS1DataBarOmniBarcode.
func NewGS1DataBarOmniBarcode() *GS1DataBarOmniBarcode {
	return &GS1DataBarOmniBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeGS1DataBarOmni),
		wideBarRatio:    2.0,
	}
}

// DefaultValue returns the GS1 DataBar default sample value.
// Ported from C# BarcodeGS1Base.GetDefaultValue (BarcodeGS1.cs:174-177).
func (b *GS1DataBarOmniBarcode) DefaultValue() string { return "(01)0000123456789" }

// Encode validates the text and pre-computes encodedData.
func (b *GS1DataBarOmniBarcode) Encode(text string) error {
	b.encodedText = text
	_, err := b.getPattern()
	return err
}

// CalcBounds returns the natural (width, height) of the encoded symbol.
// Ported from C# BarcodeGS1Omnidirectional.GetWidth + caller padding * 1.25
// (BarcodeGS1.cs:365-371).
func (b *GS1DataBarOmniBarcode) CalcBounds() (float32, float32) {
	if len(b.encodedData) == 0 {
		return 0, 0
	}
	return gs1OmniWidth(b.encodedData[0], b.wideBarRatio) * 1.25, 0
}

// gs1OmniWidth sums the digit values in the 46-char barWeights string.
// Ported from C# GetWidth (BarcodeGS1.cs:365-371).
func gs1OmniWidth(data string, wideBarRatio float32) float32 {
	var w float32
	for i := range len(data) {
		w += float32(data[i] - '0')
	}
	return w * wideBarRatio
}

// gs1OmniGetValue parses the input text "(01)NNNNNNNNNNNNN" or "NNNNNNNNNNNNN"
// into an int64 and stores the canonical text (with checksum) in b.encodedText.
// Ported from C# BarcodeGS1Omnidirectional.GetValue (BarcodeGS1.cs:190-214).
func (b *GS1DataBarOmniBarcode) gs1OmniGetValue() (int64, error) {
	data := b.encodedText
	prefix := ""
	startParen := strings.Index(data, "(")
	endParen := strings.Index(data, ")")
	if startParen >= 0 && endParen > 0 {
		prefix = data[startParen : endParen+1]
		data = strings.Replace(data, prefix, "", 1)
	}
	if len(data) > 13 {
		data = data[:13] + data[14:] // remove position 13 (the checksum)
	}
	result, err := strconv.ParseInt(data, 10, 64)
	if err != nil || len(data) != 13 || result < 0 {
		return 0, fmt.Errorf("gs1databar: invalid barcode value %q", b.encodedText)
	}
	if prefix == "" {
		prefix = "(01)"
	}
	// Rewrite stored text with checksum.
	b.encodedText = prefix + CheckSumModulo10(data)
	return result + 10000000000000, nil
}

// getPattern encodes the barcode and populates b.encodedData.
// Ported from C# BarcodeGS1Omnidirectional.GetPattern (BarcodeGS1.cs:244-363).
func (b *GS1DataBarOmniBarcode) getPattern() (string, error) {
	b.encodedData = nil

	value, err := b.gs1OmniGetValue()
	if err != nil {
		return "", err
	}

	left := value / 4537077
	right := value % 4537077
	data1 := int(left / 1597)
	data2 := int(left % 1597)
	data3 := int(right / 1597)
	data4 := int(right % 1597)

	dataGroup := [4]int{}
	vOdd := [4]int{}
	vEven := [4]int{}

	// Assign data groups based on range thresholds (C# lines 259-276).
	switch {
	case data1 <= 160:
		dataGroup[0] = 0
	case data1 <= 960:
		dataGroup[0] = 1
	case data1 <= 2014:
		dataGroup[0] = 2
	case data1 <= 2714:
		dataGroup[0] = 3
	default:
		dataGroup[0] = 4
	}
	switch {
	case data2 <= 335:
		dataGroup[1] = 5
	case data2 <= 1035:
		dataGroup[1] = 6
	case data2 <= 1515:
		dataGroup[1] = 7
	default:
		dataGroup[1] = 8
	}
	switch {
	case data3 <= 160:
		dataGroup[2] = 0
	case data3 <= 960:
		dataGroup[2] = 1
	case data3 <= 2014:
		dataGroup[2] = 2
	case data3 <= 2714:
		dataGroup[2] = 3
	default:
		dataGroup[2] = 4
	}
	switch {
	case data4 <= 335:
		dataGroup[3] = 5
	case data4 <= 1035:
		dataGroup[3] = 6
	case data4 <= 1515:
		dataGroup[3] = 7
	default:
		dataGroup[3] = 8
	}

	// v_odd / v_even for each of the 4 data segments.
	// C# lines 278-285.
	vOdd[0] = (data1 - omniGSums[dataGroup[0]]) / omniTList[dataGroup[0]]
	vEven[0] = (data1 - omniGSums[dataGroup[0]]) % omniTList[dataGroup[0]]
	vOdd[1] = (data2 - omniGSums[dataGroup[1]]) % omniTList[dataGroup[1]]
	vEven[1] = (data2 - omniGSums[dataGroup[1]]) / omniTList[dataGroup[1]]
	vOdd[3] = (data4 - omniGSums[dataGroup[3]]) % omniTList[dataGroup[3]]
	vEven[3] = (data4 - omniGSums[dataGroup[3]]) / omniTList[dataGroup[3]]
	vOdd[2] = (data3 - omniGSums[dataGroup[2]]) / omniTList[dataGroup[2]]
	vEven[2] = (data3 - omniGSums[dataGroup[2]]) % omniTList[dataGroup[2]]

	// data_widths[8][4] — C# lines 287-319.
	var dataWidths [8][4]int
	for i := 0; i < 4; i++ {
		if i == 0 || i == 2 {
			w := gs1GetWidths(vOdd[i], omniModulesOdd[dataGroup[i]], 4, omniWidthsOdd[dataGroup[i]], 1)
			dataWidths[0][i] = w[0]
			dataWidths[2][i] = w[1]
			dataWidths[4][i] = w[2]
			dataWidths[6][i] = w[3]
			w = gs1GetWidths(vEven[i], omniModulesEven[dataGroup[i]], 4, omniWidthsEven[dataGroup[i]], 0)
			dataWidths[1][i] = w[0]
			dataWidths[3][i] = w[1]
			dataWidths[5][i] = w[2]
			dataWidths[7][i] = w[3]
		} else {
			w := gs1GetWidths(vOdd[i], omniModulesOdd[dataGroup[i]], 4, omniWidthsOdd[dataGroup[i]], 0)
			dataWidths[0][i] = w[0]
			dataWidths[2][i] = w[1]
			dataWidths[4][i] = w[2]
			dataWidths[6][i] = w[3]
			w = gs1GetWidths(vEven[i], omniModulesEven[dataGroup[i]], 4, omniWidthsEven[dataGroup[i]], 1)
			dataWidths[1][i] = w[0]
			dataWidths[3][i] = w[1]
			dataWidths[5][i] = w[2]
			dataWidths[7][i] = w[3]
		}
	}

	// Calculate checksum (C# lines 323-337).
	checksum := 0
	for i := 0; i < 8; i++ {
		checksum += omniChecksumWeight[i] * dataWidths[i][0]
		checksum += omniChecksumWeight[i+8] * dataWidths[i][1]
		checksum += omniChecksumWeight[i+16] * dataWidths[i][2]
		checksum += omniChecksumWeight[i+24] * dataWidths[i][3]
	}
	checksum %= 79
	if checksum >= 8 {
		checksum++
	}
	if checksum >= 72 {
		checksum++
	}
	cLeft := checksum / 9
	cRight := checksum % 9

	// Assemble barWeights[46] (C# lines 338-356).
	barWeights := [46]int{}
	barWeights[0] = 1
	barWeights[1] = 1
	barWeights[44] = 1
	barWeights[45] = 1
	for i := 0; i < 8; i++ {
		barWeights[i+2] = dataWidths[i][0]
		barWeights[i+15] = dataWidths[7-i][1]
		barWeights[i+23] = dataWidths[i][3]
		barWeights[i+36] = dataWidths[7-i][2]
	}
	for i := 0; i < 5; i++ {
		barWeights[i+10] = omniFinderPattern[i+(5*cLeft)]
		barWeights[i+31] = omniFinderPattern[(4-i)+(5*cRight)]
	}

	// Build encodedData[0] string from digit values (C# lines 358-362).
	var sb strings.Builder
	for _, v := range barWeights {
		sb.WriteByte(byte('0' + v))
	}
	b.encodedData = []string{sb.String()}
	return b.encodedData[0], nil
}

// Render renders the GS1 DataBar Omnidirectional barcode.
func (b *GS1DataBarOmniBarcode) Render(width, height int) (image.Image, error) {
	if len(b.encodedData) == 0 {
		return nil, fmt.Errorf("gs1databar omni: Encode must be called before Render")
	}
	if width <= 0 || height <= 0 {
		return image.NewRGBA(image.Rect(0, 0, max(width, 1), max(height, 1))), nil
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	totalWidth := gs1OmniWidth(b.encodedData[0], b.wideBarRatio)
	zoom := float32(width) / (totalWidth * 1.25)

	barH := height
	if b.showText && b.encodedText != "" {
		textH := int(math.Round(float64(14 * zoom)))
		if textH < 1 {
			textH = 1
		}
		if textH < height {
			barH = height - textH
			drawLinearText(img, b.encodedText, 0, barH, width, textH)
		}
	}
	drawGS1Bars(b.encodedData[0], img, 0, barH, zoom, b.wideBarRatio, false, false)
	return img, nil
}

// ── GS1 DataBar Stacked ──────────────────────────────────────────────────────

// GS1DataBarStackedBarcode implements GS1 DataBar Stacked.
// Ported from C# BarcodeGS1Stacked (BarcodeGS1.cs:383-445).
type GS1DataBarStackedBarcode struct {
	GS1DataBarOmniBarcode
}

// NewGS1DataBarStackedBarcode creates a GS1DataBarStackedBarcode.
func NewGS1DataBarStackedBarcode() *GS1DataBarStackedBarcode {
	b := &GS1DataBarStackedBarcode{}
	b.BaseBarcodeImpl = newBaseBarcodeImpl(BarcodeTypeGS1DataBarStacked)
	b.wideBarRatio = 2.0
	return b
}

// DefaultValue returns the GS1 DataBar Stacked default sample value.
func (b *GS1DataBarStackedBarcode) DefaultValue() string { return "(01)0000123456789" }

// Encode validates and pre-computes encodedData for the stacked variant.
func (b *GS1DataBarStackedBarcode) Encode(text string) error {
	b.encodedText = text
	_, err := b.stackedGetPattern()
	return err
}

// CalcBounds returns the natural width of the first row * 1.25.
func (b *GS1DataBarStackedBarcode) CalcBounds() (float32, float32) {
	if len(b.encodedData) == 0 {
		return 0, 0
	}
	return gs1OmniWidth(b.encodedData[0], b.wideBarRatio) * 1.25, 0
}

// stackedGetPattern calls the parent omni GetPattern then builds the 3 stacked rows.
// Ported from C# BarcodeGS1Stacked.GetPattern (BarcodeGS1.cs:385-428).
func (b *GS1DataBarStackedBarcode) stackedGetPattern() (string, error) {
	// Get the full 46-char omni pattern.
	data, err := b.GS1DataBarOmniBarcode.getPattern()
	if err != nil {
		return "", err
	}

	// Split: top = first 23 chars + "11", bottom = "11" + next 23 chars.
	b.encodedData = []string{
		data[:23] + "11",
		"0000", // left padding of separator line
		"11" + data[23:46],
	}

	// Convert run-length encoded rows to module-by-module bit strings.
	// C# lines 395-413.
	bars := [2]strings.Builder{}
	row0 := b.encodedData[0]
	row2 := b.encodedData[2]
	for i := range len(row0) {
		if i%2 == 0 {
			// even index: row0 contributes '0' (white), row2 contributes '1' (black)
			for range int(row0[i] - '0') {
				bars[0].WriteByte('0')
			}
			for range int(row2[i] - '0') {
				bars[1].WriteByte('1')
			}
		} else {
			// odd index: row0 contributes '1' (black), row2 contributes '0' (white)
			for range int(row0[i] - '0') {
				bars[0].WriteByte('1')
			}
			for range int(row2[i] - '0') {
				bars[1].WriteByte('0')
			}
		}
	}

	// Build separator line from module strings (C# lines 415-425).
	// The separator is evaluated from position 4 to len-4 (skipping padding).
	bars0 := bars[0].String()
	bars1 := bars[1].String()
	end := len(bars0) - 4
	for i := 4; i < end; i++ {
		b0 := bars0[i]
		b1 := bars1[i]
		sep := b.encodedData[1]
		switch {
		case b0 == '1' && b1 == '1':
			b.encodedData[1] += "0"
		case b0 == '0' && b1 == '0':
			b.encodedData[1] += "1"
		default:
			// Toggle from last separator char.
			if sep[len(sep)-1] == '0' {
				b.encodedData[1] += "1"
			} else {
				b.encodedData[1] += "0"
			}
		}
	}

	return "", nil
}

// Render renders the 3-row stacked barcode.
// Ported from C# BarcodeGS1Stacked.DoLines (BarcodeGS1.cs:431-436).
func (b *GS1DataBarStackedBarcode) Render(width, height int) (image.Image, error) {
	if len(b.encodedData) < 3 {
		return nil, fmt.Errorf("gs1databar stacked: Encode must be called before Render")
	}
	if width <= 0 || height <= 0 {
		return image.NewRGBA(image.Rect(0, 0, max(width, 1), max(height, 1))), nil
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	totalWidth := gs1OmniWidth(b.encodedData[0], b.wideBarRatio)
	zoom := float32(width) / (totalWidth * 1.25)

	barH := height
	if b.showText && b.encodedText != "" {
		textH := int(math.Round(float64(14 * zoom)))
		if textH < 1 {
			textH = 1
		}
		if textH < height {
			barH = height - textH
			drawLinearText(img, b.encodedText, 0, barH, width, textH)
		}
	}

	// C# heights: row0 = 0..5/13, sep = 5/13..6/13, row2 = 6/13..13/13
	h := float32(barH)
	y0top := 0
	y0bot := int(math.Round(float64(h * 5 / 13)))
	y1top := y0bot
	y1bot := int(math.Round(float64(h * 6 / 13)))
	y2top := y1bot
	y2bot := barH

	drawGS1Bars(b.encodedData[0], img, y0top, y0bot, zoom, b.wideBarRatio, false, false)
	drawGS1Bars(b.encodedData[1], img, y1top, y1bot, zoom, b.wideBarRatio, false, true)
	drawGS1Bars(b.encodedData[2], img, y2top, y2bot, zoom, b.wideBarRatio, true, false)
	return img, nil
}

// ── GS1 DataBar Stacked Omnidirectional ──────────────────────────────────────

// GS1DataBarStackedOmniBarcode implements GS1 DataBar Stacked Omnidirectional.
// Ported from C# BarcodeGS1StackedOmnidirectional (BarcodeGS1.cs:450-543).
type GS1DataBarStackedOmniBarcode struct {
	GS1DataBarOmniBarcode
}

// NewGS1DataBarStackedOmniBarcode creates a GS1DataBarStackedOmniBarcode.
func NewGS1DataBarStackedOmniBarcode() *GS1DataBarStackedOmniBarcode {
	b := &GS1DataBarStackedOmniBarcode{}
	b.BaseBarcodeImpl = newBaseBarcodeImpl(BarcodeTypeGS1DataBarStackedOmni)
	b.wideBarRatio = 2.0
	return b
}

// DefaultValue returns the GS1 DataBar Stacked Omnidirectional default sample value.
func (b *GS1DataBarStackedOmniBarcode) DefaultValue() string { return "(01)0000123456789" }

// Encode validates and pre-computes encodedData.
func (b *GS1DataBarStackedOmniBarcode) Encode(text string) error {
	b.encodedText = text
	_, err := b.stackedOmniGetPattern()
	return err
}

// CalcBounds returns the natural width of the first row * 1.25.
func (b *GS1DataBarStackedOmniBarcode) CalcBounds() (float32, float32) {
	if len(b.encodedData) == 0 {
		return 0, 0
	}
	return gs1OmniWidth(b.encodedData[0], b.wideBarRatio) * 1.25, 0
}

// stackedOmniGetPattern builds 5 encoded segments for the stacked omnidirectional barcode.
// Ported from C# BarcodeGS1StackedOmnidirectional.GetPattern (BarcodeGS1.cs:453-524).
func (b *GS1DataBarStackedOmniBarcode) stackedOmniGetPattern() (string, error) {
	data, err := b.GS1DataBarOmniBarcode.getPattern()
	if err != nil {
		return "", err
	}

	// C# lines 460-464: initialise 5 encoded segments.
	b.encodedData = []string{
		data[:23] + "11",
		"0000",
		"0000010101010101010101010101010101010101010101",
		"0000",
		"11" + data[23:46],
	}

	// Build top and bottom bit-strings from EncodedData[0] and [4].
	// C# lines 467-519: the separator encoding rules for 5.3.2.2.
	bars := [2]strings.Builder{}
	nextBarBlack := true
	nextBarWhite := true
	row0 := b.encodedData[0]
	row4 := b.encodedData[4]
	for i := range len(row0) {
		if i%2 == 0 {
			// Even index: row0 normally black, row4 normally white.
			for range int(row0[i] - '0') {
				if i > 5 && i < 9 {
					if nextBarBlack {
						bars[0].WriteByte('1')
						nextBarBlack = false
					} else {
						bars[0].WriteByte('0')
						nextBarBlack = true
					}
				} else {
					bars[0].WriteByte('1')
				}
			}
			for range int(row4[i] - '0') {
				if i > 15 && i < 19 {
					if nextBarWhite {
						bars[1].WriteByte('0')
						nextBarWhite = false
					} else {
						bars[1].WriteByte('1')
						nextBarWhite = true
					}
				} else {
					bars[1].WriteByte('0')
				}
			}
		} else {
			// Odd index: row0 always white (0), row4 always black (1).
			for range int(row0[i] - '0') {
				bars[0].WriteByte('0')
			}
			for range int(row4[i] - '0') {
				bars[1].WriteByte('1')
			}
		}
	}

	// Trim 4 chars from each end, append to separators[1] and [3].
	// C# lines 521-522.
	bars0str := bars[0].String()
	bars1str := bars[1].String()
	if len(bars0str) >= 8 {
		b.encodedData[1] += bars0str[4 : len(bars0str)-4]
	}
	if len(bars1str) >= 8 {
		b.encodedData[3] += bars1str[4 : len(bars1str)-4]
	}

	return "", nil
}

// Render renders the 5-segment stacked omnidirectional barcode.
// Ported from C# BarcodeGS1StackedOmnidirectional.DoLines (BarcodeGS1.cs:536-543).
func (b *GS1DataBarStackedOmniBarcode) Render(width, height int) (image.Image, error) {
	if len(b.encodedData) < 5 {
		return nil, fmt.Errorf("gs1databar stacked omni: Encode must be called before Render")
	}
	if width <= 0 || height <= 0 {
		return image.NewRGBA(image.Rect(0, 0, max(width, 1), max(height, 1))), nil
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	totalWidth := gs1OmniWidth(b.encodedData[0], b.wideBarRatio)
	zoom := float32(width) / (totalWidth * 1.25)

	barH := height
	if b.showText && b.encodedText != "" {
		textH := int(math.Round(float64(14 * zoom)))
		if textH < 1 {
			textH = 1
		}
		if textH < height {
			barH = height - textH
			drawLinearText(img, b.encodedText, 0, barH, width, textH)
		}
	}

	// C# heights (fractions of barArea.Height):
	// row0:  0       .. 33/69
	// sep1:  33/69   .. 34/69
	// sep2:  34/69   .. 35/69
	// sep3:  35/69   .. 36/69
	// row4:  36/69   .. 69/69
	h := float32(barH)
	y := [6]int{
		0,
		int(math.Round(float64(h * 33 / 69))),
		int(math.Round(float64(h * 34 / 69))),
		int(math.Round(float64(h * 35 / 69))),
		int(math.Round(float64(h * 36 / 69))),
		barH,
	}

	drawGS1Bars(b.encodedData[0], img, y[0], y[1], zoom, b.wideBarRatio, false, false)
	drawGS1Bars(b.encodedData[1], img, y[1], y[2], zoom, b.wideBarRatio, false, true)
	drawGS1Bars(b.encodedData[2], img, y[2], y[3], zoom, b.wideBarRatio, false, true)
	drawGS1Bars(b.encodedData[3], img, y[3], y[4], zoom, b.wideBarRatio, false, true)
	drawGS1Bars(b.encodedData[4], img, y[4], y[5], zoom, b.wideBarRatio, true, false)
	return img, nil
}

// ── GS1 DataBar Limited ──────────────────────────────────────────────────────

// GS1 DataBar Limited encoding tables.
// Ported from C# BarcodeGS1Limited (BarcodeGS1.cs:549-838).
var limitedChecksumWeight = [28]int{
	1, 3, 9, 27, 81, 65, 17, 51, 64, 14, 42, 37, 22, 66,
	20, 60, 2, 6, 18, 54, 73, 41, 34, 13, 39, 28, 84, 74,
}

// limitedFinderPattern is the full 14×89 table flattened row-major.
// Ported from C# BarcodeGS1Limited.FinderPattern (BarcodeGS1.cs:587-677).
var limitedFinderPattern = [...]int{
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 3, 2, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 3, 2, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 2, 1, 2, 3, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 3, 1, 1, 3, 1, 1, 1,
	1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 3, 2, 1, 1,
	1, 1, 1, 1, 1, 2, 1, 1, 1, 2, 3, 1, 1, 1,
	1, 1, 1, 1, 1, 2, 1, 2, 1, 1, 3, 1, 1, 1,
	1, 1, 1, 1, 1, 3, 1, 1, 1, 1, 3, 1, 1, 1,
	1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 3, 2, 1, 1,
	1, 1, 1, 2, 1, 1, 1, 1, 1, 2, 3, 1, 1, 1,
	1, 1, 1, 2, 1, 1, 1, 2, 1, 1, 3, 1, 1, 1,
	1, 1, 1, 2, 1, 2, 1, 1, 1, 1, 3, 1, 1, 1,
	1, 1, 1, 3, 1, 1, 1, 1, 1, 1, 3, 1, 1, 1,
	1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 3, 2, 1, 1,
	1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 3, 1, 1, 1,
	1, 2, 1, 1, 1, 1, 1, 2, 1, 1, 3, 1, 1, 1,
	1, 2, 1, 1, 1, 2, 1, 1, 1, 1, 3, 1, 1, 1,
	1, 2, 1, 2, 1, 1, 1, 1, 1, 1, 3, 1, 1, 1,
	1, 3, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 2, 3, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 2, 3, 2, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 2, 2, 1, 2, 2, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 3, 2, 1, 2, 1, 1, 1,
	1, 1, 1, 1, 1, 2, 1, 1, 2, 1, 2, 2, 1, 1,
	1, 1, 1, 1, 1, 2, 1, 1, 2, 2, 2, 1, 1, 1,
	1, 1, 1, 1, 1, 2, 1, 2, 2, 1, 2, 1, 1, 1,
	1, 1, 1, 1, 1, 3, 1, 1, 2, 1, 2, 1, 1, 1,
	1, 1, 1, 2, 1, 1, 1, 1, 2, 1, 2, 2, 1, 1,
	1, 1, 1, 2, 1, 1, 1, 1, 2, 2, 2, 1, 1, 1,
	1, 1, 1, 2, 1, 1, 1, 2, 2, 1, 2, 1, 1, 1,
	1, 1, 1, 2, 1, 2, 1, 1, 2, 1, 2, 1, 1, 1,
	1, 1, 1, 3, 1, 1, 1, 1, 2, 1, 2, 1, 1, 1,
	1, 2, 1, 1, 1, 1, 1, 1, 2, 1, 2, 2, 1, 1,
	1, 2, 1, 1, 1, 1, 1, 1, 2, 2, 2, 1, 1, 1,
	1, 2, 1, 1, 1, 1, 1, 2, 2, 1, 2, 1, 1, 1,
	1, 2, 1, 1, 1, 2, 1, 1, 2, 1, 2, 1, 1, 1,
	1, 2, 1, 2, 1, 1, 1, 1, 2, 1, 2, 1, 1, 1,
	1, 3, 1, 1, 1, 1, 1, 1, 2, 1, 2, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 1, 3, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 3, 2, 1, 2, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 2, 3, 1, 1, 2, 1, 1,
	1, 1, 1, 2, 1, 1, 1, 1, 3, 1, 1, 2, 1, 1,
	1, 2, 1, 1, 1, 1, 1, 1, 3, 1, 1, 2, 1, 1,
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 2, 3, 1, 1,
	1, 1, 1, 1, 1, 1, 2, 1, 1, 2, 2, 2, 1, 1,
	1, 1, 1, 1, 1, 1, 2, 1, 1, 3, 2, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 2, 2, 1, 1, 2, 2, 1, 1,
	1, 1, 1, 2, 1, 1, 2, 1, 1, 1, 2, 2, 1, 1,
	1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 2, 1, 1, 1,
	1, 1, 1, 2, 1, 1, 2, 2, 1, 1, 2, 1, 1, 1,
	1, 1, 1, 2, 1, 2, 2, 1, 1, 1, 2, 1, 1, 1,
	1, 1, 1, 3, 1, 1, 2, 1, 1, 1, 2, 1, 1, 1,
	1, 2, 1, 1, 1, 1, 2, 1, 1, 1, 2, 2, 1, 1,
	1, 2, 1, 1, 1, 1, 2, 1, 1, 2, 2, 1, 1, 1,
	1, 2, 1, 2, 1, 1, 2, 1, 1, 1, 2, 1, 1, 1,
	1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 2, 3, 1, 1,
	1, 1, 1, 1, 2, 1, 1, 1, 1, 2, 2, 2, 1, 1,
	1, 1, 1, 1, 2, 1, 1, 1, 1, 3, 2, 1, 1, 1,
	1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 2, 1, 1,
	1, 1, 1, 1, 2, 1, 1, 2, 1, 2, 2, 1, 1, 1,
	1, 1, 1, 1, 2, 2, 1, 1, 1, 1, 2, 2, 1, 1,
	1, 2, 1, 1, 2, 1, 1, 1, 1, 1, 2, 2, 1, 1,
	1, 2, 1, 1, 2, 1, 1, 1, 1, 2, 2, 1, 1, 1,
	1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 1,
	1, 2, 1, 1, 2, 2, 1, 1, 1, 1, 2, 1, 1, 1,
	1, 2, 1, 2, 2, 1, 1, 1, 1, 1, 2, 1, 1, 1,
	1, 3, 1, 1, 2, 1, 1, 1, 1, 1, 2, 1, 1, 1,
	1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 3, 1, 1,
	1, 1, 2, 1, 1, 1, 1, 1, 1, 2, 2, 2, 1, 1,
	1, 1, 2, 1, 1, 1, 1, 1, 1, 3, 2, 1, 1, 1,
	1, 1, 2, 1, 1, 1, 1, 2, 1, 1, 2, 2, 1, 1,
	1, 1, 2, 1, 1, 1, 1, 2, 1, 2, 2, 1, 1, 1,
	1, 1, 2, 1, 1, 1, 1, 3, 1, 1, 2, 1, 1, 1,
	1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 2, 2, 1, 1,
	1, 1, 2, 1, 1, 2, 1, 1, 1, 2, 2, 1, 1, 1,
	1, 1, 2, 2, 1, 1, 1, 1, 1, 1, 2, 2, 1, 1,
	2, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 1, 1,
	2, 1, 1, 1, 1, 1, 1, 1, 1, 3, 2, 1, 1, 1,
	2, 1, 1, 1, 1, 1, 1, 2, 1, 1, 2, 2, 1, 1,
	2, 1, 1, 1, 1, 1, 1, 2, 1, 2, 2, 1, 1, 1,
	2, 1, 1, 1, 1, 1, 1, 3, 1, 1, 2, 1, 1, 1,
	2, 1, 1, 1, 1, 2, 1, 1, 1, 2, 2, 1, 1, 1,
	2, 1, 1, 1, 1, 2, 1, 2, 1, 1, 2, 1, 1, 1,
	2, 1, 1, 2, 1, 1, 1, 1, 1, 2, 2, 1, 1, 1,
}

var limitedModulesOdd = [7]int{17, 13, 9, 15, 11, 19, 7}
var limitedModulesEven = [7]int{9, 13, 17, 11, 15, 7, 19}
var limitedWidthsOdd = [7]int{6, 5, 3, 5, 4, 8, 1}
var limitedWidthsEven = [7]int{3, 4, 6, 4, 5, 1, 8}
var limitedTEven = [7]int{28, 728, 6454, 203, 2408, 1, 16632}

// GS1DataBarLimitedBarcode implements GS1 DataBar Limited.
// Ported from C# BarcodeGS1Limited (BarcodeGS1.cs:549-839).
type GS1DataBarLimitedBarcode struct {
	BaseBarcodeImpl
	encodedData  []string
	wideBarRatio float32
}

// NewGS1DataBarLimitedBarcode creates a GS1DataBarLimitedBarcode.
func NewGS1DataBarLimitedBarcode() *GS1DataBarLimitedBarcode {
	return &GS1DataBarLimitedBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeGS1DataBarLimited),
		wideBarRatio:    2.0,
	}
}

// DefaultValue returns the GS1 DataBar Limited default sample value.
func (b *GS1DataBarLimitedBarcode) DefaultValue() string { return "(01)0000123456789" }

// Encode validates and pre-computes encodedData.
func (b *GS1DataBarLimitedBarcode) Encode(text string) error {
	b.encodedText = text
	_, err := b.getPattern()
	return err
}

// CalcBounds returns the natural width of the encoded symbol.
func (b *GS1DataBarLimitedBarcode) CalcBounds() (float32, float32) {
	if len(b.encodedData) == 0 {
		return 0, 0
	}
	return gs1OmniWidth(b.encodedData[0], b.wideBarRatio) * 1.25, 0
}

// gs1LimitedGetValue parses input text and returns the raw value (without +10^13 offset).
// Ported from C# BarcodeGS1Limited.GetValue (BarcodeGS1.cs:556-578).
func (b *GS1DataBarLimitedBarcode) gs1LimitedGetValue() (int64, error) {
	data := b.encodedText
	prefix := ""
	startParen := strings.Index(data, "(")
	endParen := strings.Index(data, ")")
	if startParen >= 0 && endParen > 0 {
		prefix = data[startParen : endParen+1]
		data = strings.Replace(data, prefix, "", 1)
	}
	if len(data) > 13 {
		data = data[:13] + data[14:]
	}
	result, err := strconv.ParseInt(data, 10, 64)
	if err != nil || len(data) != 13 || result > 1999999999999 || result < 0 {
		return 0, fmt.Errorf("gs1databar limited: invalid barcode value %q (must be 0-1999999999999)", b.encodedText)
	}
	if prefix == "" {
		prefix = "(01)"
	}
	b.encodedText = prefix + CheckSumModulo10(data)
	return result, nil
}

// getPattern encodes the Limited barcode.
// Ported from C# BarcodeGS1Limited.GetPattern (BarcodeGS1.cs:685-824).
func (b *GS1DataBarLimitedBarcode) getPattern() (string, error) {
	b.encodedData = nil

	value, err := b.gs1LimitedGetValue()
	if err != nil {
		return "", err
	}

	left := value / 2013571
	right := value % 2013571

	// Assign left group (C# lines 693-698).
	leftGroup := 0
	if left > 183063 {
		leftGroup = 1
	}
	if left > 820063 {
		leftGroup = 2
	}
	if left > 1000775 {
		leftGroup = 3
	}
	if left > 1491020 {
		leftGroup = 4
	}
	if left > 1979844 {
		leftGroup = 5
	}
	if left > 1996938 {
		leftGroup = 6
	}

	// Assign right group (C# lines 700-706).
	rightGroup := 0
	if right > 183063 {
		rightGroup = 1
	}
	if right > 820063 {
		rightGroup = 2
	}
	if right > 1000775 {
		rightGroup = 3
	}
	if right > 1491020 {
		rightGroup = 4
	}
	if right > 1979844 {
		rightGroup = 5
	}
	if right > 1996938 {
		rightGroup = 6
	}

	// Subtract base offset for each group (C# lines 707-749).
	leftOffsets := [7]int64{0, 183064, 820064, 1000776, 1491021, 1979845, 1996939}
	rightOffsets := [7]int64{0, 183064, 820064, 1000776, 1491021, 1979845, 1996939}
	left -= leftOffsets[leftGroup]
	right -= rightOffsets[rightGroup]

	leftOdd := int(left / int64(limitedTEven[leftGroup]))
	leftEven := int(left % int64(limitedTEven[leftGroup]))
	rightOdd := int(right / int64(limitedTEven[rightGroup]))
	rightEven := int(right % int64(limitedTEven[rightGroup]))

	// Build left and right 14-element width arrays (C# lines 756-790).
	leftWidths := [14]int{}
	rightWidths := [14]int{}

	w := gs1GetWidths(leftOdd, limitedModulesOdd[leftGroup], 7, limitedWidthsOdd[leftGroup], 1)
	for j := 0; j < 7; j++ {
		leftWidths[j*2] = w[j]
	}
	w = gs1GetWidths(leftEven, limitedModulesEven[leftGroup], 7, limitedWidthsEven[leftGroup], 0)
	for j := 0; j < 7; j++ {
		leftWidths[j*2+1] = w[j]
	}
	w = gs1GetWidths(rightOdd, limitedModulesOdd[rightGroup], 7, limitedWidthsOdd[rightGroup], 1)
	for j := 0; j < 7; j++ {
		rightWidths[j*2] = w[j]
	}
	w = gs1GetWidths(rightEven, limitedModulesEven[rightGroup], 7, limitedWidthsEven[rightGroup], 0)
	for j := 0; j < 7; j++ {
		rightWidths[j*2+1] = w[j]
	}

	// Calculate checksum (C# lines 792-799).
	checksum := 0
	for i := 0; i < 14; i++ {
		checksum += limitedChecksumWeight[i] * leftWidths[i]
		checksum += limitedChecksumWeight[i+14] * rightWidths[i]
	}
	checksum %= 89

	// Look up check elements from finder pattern table (C# lines 801-805).
	checkElements := [14]int{}
	for i := 0; i < 14; i++ {
		checkElements[i] = limitedFinderPattern[i+(checksum*14)]
	}

	// Assemble totalWidths[46] (C# lines 807-817).
	totalWidths := [46]int{}
	totalWidths[0] = 1
	totalWidths[1] = 1
	totalWidths[44] = 1
	totalWidths[45] = 1
	for i := 0; i < 14; i++ {
		totalWidths[i+2] = leftWidths[i]
		totalWidths[i+16] = checkElements[i]
		totalWidths[i+30] = rightWidths[i]
	}

	var sb strings.Builder
	for _, v := range totalWidths {
		sb.WriteByte(byte('0' + v))
	}
	b.encodedData = []string{sb.String()}
	return b.encodedData[0], nil
}

// Render renders the GS1 DataBar Limited barcode.
// Ported from C# BarcodeGS1Limited.DoLines (BarcodeGS1.cs:827-830).
func (b *GS1DataBarLimitedBarcode) Render(width, height int) (image.Image, error) {
	if len(b.encodedData) == 0 {
		return nil, fmt.Errorf("gs1databar limited: Encode must be called before Render")
	}
	if width <= 0 || height <= 0 {
		return image.NewRGBA(image.Rect(0, 0, max(width, 1), max(height, 1))), nil
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	totalWidth := gs1OmniWidth(b.encodedData[0], b.wideBarRatio)
	zoom := float32(width) / (totalWidth * 1.25)

	barH := height
	if b.showText && b.encodedText != "" {
		textH := int(math.Round(float64(14 * zoom)))
		if textH < 1 {
			textH = 1
		}
		if textH < height {
			barH = height - textH
			drawLinearText(img, b.encodedText, 0, barH, width, textH)
		}
	}
	drawGS1Bars(b.encodedData[0], img, 0, barH, zoom, b.wideBarRatio, false, false)
	return img, nil
}

// ── drawGS1Bars — low-level rendering ────────────────────────────────────────

// drawGS1Bars renders a single GS1 DataBar strip onto img.
// Ported from C# BarcodeGS1Base.DrawLineBars (BarcodeGS1.cs:126-171).
//
//   - data:          digit string where each digit is the bar width in modules
//   - img:           destination RGBA image
//   - y0, y1:        vertical pixel range
//   - zoom:          pixels-per-module scale factor
//   - wideBarRatio:  modules-per-data-unit (2.0 default, matching C# LinearBarcodeBase)
//   - reversColor:   if true, bar drawing starts with black (first element black)
//   - separatorLine: if true, each char is treated as a single-module b/w cell
//
// In normal mode (reversColor=false), bars alternate white/black starting with
// white (even index = white, odd index = black).
// In reversColor mode, the parity is flipped (even = black, odd = white).
// In separatorLine mode, '0' = white, non-'0' = black; each cell is 1 module wide.
func drawGS1Bars(data string, img *image.RGBA, y0, y1 int, zoom, wideBarRatio float32, reversColor, separatorLine bool) {
	if y0 >= y1 || len(data) == 0 {
		return
	}
	imgW := img.Bounds().Max.X
	black := color.RGBA{A: 255}

	var curX float32
	for x := range len(data) {
		var barW float32
		if separatorLine {
			barW = wideBarRatio * zoom
		} else {
			barW = float32(data[x]-'0') * wideBarRatio * zoom
		}

		// Determine fill colour.
		drawBlack := false
		if separatorLine {
			drawBlack = data[x] != '0'
		} else {
			if reversColor {
				// even = black, odd = white
				drawBlack = (x % 2) == 0
			} else {
				// even = white, odd = black
				drawBlack = (x % 2) != 0
			}
		}

		if drawBlack {
			px0 := int(math.Round(float64(curX)))
			px1 := int(math.Round(float64(curX + barW)))
			if px1 <= px0 {
				px1 = px0 + 1
			}
			if px0 < 0 {
				px0 = 0
			}
			if px1 > imgW {
				px1 = imgW
			}
			if px0 < px1 {
				r := image.Rect(px0, y0, px1, y1)
				draw.Draw(img, r, image.NewUniform(black), image.Point{}, draw.Src)
			}
		}

		curX += barW
	}
}
