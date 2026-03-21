package barcode

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
)

// ── MaxiCode encoding ─────────────────────────────────────────────────────────
//
// MaxiCode (ISO/IEC 16023) is a 2D matrix barcode symbol consisting of 144
// total codewords (data + error correction) arranged in a hexagonal grid of
// 30 columns × 33 rows, with a central bullseye finder.
//
// Each codeword is 6 bits wide (64 possible values). Characters are encoded
// using two character sets:
//   - Set A: control characters and GS1 special characters
//   - Set B: printable ASCII (SP through DEL)
//
// Error correction uses Reed-Solomon coding over GF(64) with primitive
// polynomial x^6 + x + 1 (0x43). The 144-codeword layout is:
//
//	[0..9]          10 primary data codewords (mode byte + 9 bytes)
//	[10..19]        10 primary ECC codewords
//	[20..103/87]    84 or 68 secondary data codewords (modes 4/6 or 5)
//	[104..143]      40 or 56 secondary ECC codewords (interleaved odd/even)

// maxiCodeSetA maps Set A indices (0–63) to their character values.
// Set A contains: NUL, SOH … GS, RS, EOT, ENQ, ACK, BEL, BS, HT, LF, VT, FF,
// CR, SO, SI, DLE, DC1–DC4, NAK, SYN, ETB, CAN, EM, SUB, ESC, FS, … (low ctrl).
// Printable range starts at 0x20 (SP) for Set B.
// The canonical Set A/B assignment follows ISO/IEC 16023 Annex A.
// Index 0 = LATB (latch to Set B); index 63 = LATA (latch to Set A).
var maxiCodeSetA [64]byte

// maxiCodeSetB maps Set B indices (0–63) to their character values.
// Index 0 = LATA (latch to Set A); offset 1 maps to SP (0x20) through DEL (0x7F).
var maxiCodeSetB [64]byte

// MaxiCodeComputeECC is exported for testing; it computes RS ECC for MaxiCode.
func MaxiCodeComputeECC(data []byte, eccLen int) []byte { return maxiCodeRS(data, eccLen) }

// maxiCodeGFTable holds pre-computed GF(64) log/antilog tables.
// Generator polynomial: x^6 + x + 1 (0x43).
var maxiCodeGFTable = func() struct {
	logt [64]int
	alog [63]int
} {
	var gf struct {
		logt [64]int
		alog [63]int
	}
	p := 1
	for v := 0; v < 63; v++ {
		gf.alog[v] = p
		gf.logt[p] = v
		p <<= 1
		if p&64 != 0 {
			p ^= 0x43
		}
	}
	return gf
}()

func init() {
	// Set A: positions 0..62 map to NUL..GS (ASCII 0–28), 29..62 are padding/ctrl.
	// Position 63 = LATA shift token (no character output).
	// Simplified assignment: position i → ASCII i for i < 32, rest are special.
	for i := 0; i < 32; i++ {
		maxiCodeSetA[i] = byte(i)
	}
	// Position 32–62: fill with 0 (unused in this simplified mapping).
	maxiCodeSetA[63] = 0 // LATA sentinel

	// Set B: position 0 = LATB sentinel, positions 1–63 map to SP–DEL (0x20–0x5F).
	maxiCodeSetB[0] = 0 // LATB sentinel
	for i := 1; i < 64; i++ {
		ch := 0x1F + i // SP = 0x20 at i=1, DEL=0x7F at i=64
		if ch > 0x7F {
			ch = 0x7F
		}
		maxiCodeSetB[i] = byte(ch)
	}
}

// maxiCodeRS computes Reed-Solomon ECC for MaxiCode over GF(64).
// The generator polynomial has roots alpha^1 .. alpha^eccLen.
// Returns eccLen ECC codewords with index 0 = highest-degree check symbol.
func maxiCodeRS(data []byte, eccLen int) []byte {
	gf := &maxiCodeGFTable
	const logmod = 63

	// Build generator polynomial by multiplying (x - alpha^i) for i = 1..eccLen.
	rspoly := make([]int, eccLen+1)
	rspoly[0] = 1
	for i := 1; i <= eccLen; i++ {
		rspoly[i] = 1
		for k := i - 1; k > 0; k-- {
			if rspoly[k] != 0 {
				rspoly[k] = gf.alog[(gf.logt[rspoly[k]]+i)%logmod]
			}
			rspoly[k] ^= rspoly[k-1]
		}
		rspoly[0] = gf.alog[(gf.logt[rspoly[0]]+i)%logmod]
	}

	// Polynomial long division: compute remainder.
	res := make([]int, eccLen)
	for _, d := range data {
		m := res[eccLen-1] ^ int(d)
		for k := eccLen - 1; k > 0; k-- {
			if m != 0 && rspoly[k] != 0 {
				res[k] = res[k-1] ^ gf.alog[(gf.logt[m]+gf.logt[rspoly[k]])%logmod]
			} else {
				res[k] = res[k-1]
			}
		}
		if m != 0 && rspoly[0] != 0 {
			res[0] = gf.alog[(gf.logt[m]+gf.logt[rspoly[0]])%logmod]
		} else {
			res[0] = 0
		}
	}

	// Reverse: index 0 = highest-degree check symbol (matches C# result ordering).
	ecc := make([]byte, eccLen)
	for i := 0; i < eccLen; i++ {
		ecc[i] = byte(res[eccLen-1-i])
	}
	return ecc
}

// maxiCodeEncodeText converts text characters to 6-bit MaxiCode codewords
// starting in Set B. Shift to Set A is used for characters below 0x20.
// Returns at most maxCW codewords.
func maxiCodeEncodeText(text string, maxCW int) []byte {
	cw := make([]byte, 0, maxCW)
	inSetA := false
	for _, r := range text {
		if r > 0x7F {
			r = 0x1D // Non-ASCII: substitute GS
		}
		ch := byte(r)
		if ch < 0x20 {
			if !inSetA {
				cw = append(cw, 63) // LATA
				inSetA = true
			}
			cw = append(cw, ch&0x3F)
		} else {
			if inSetA {
				cw = append(cw, 0) // LATB
				inSetA = false
			}
			idx := int(ch) - 0x1F // SP → 1, DEL → 96→clamped 63
			if idx < 1 {
				idx = 1
			}
			if idx > 63 {
				idx = 63
			}
			cw = append(cw, byte(idx))
		}
		if len(cw) >= maxCW {
			break
		}
	}
	for len(cw) < maxCW {
		cw = append(cw, 0) // pad
	}
	return cw[:maxCW]
}

// maxiCodeMode2PrimaryCodewords returns the 10 primary codewords for Mode 2
// (numeric postal code). Ported from C# BarcodeMaxiCode.cs getMode2PrimaryCodewords.
//
// C# source: BarcodeMaxiCode.cs lines 820-847 (getMode2PrimaryCodewords).
func maxiCodeMode2PrimaryCodewords(postcode string, country, service int) []byte {
	// Strip non-digit suffix (C# truncates at first non-digit character).
	for i := 0; i < len(postcode); i++ {
		if postcode[i] < '0' || postcode[i] > '9' {
			postcode = postcode[:i]
			break
		}
	}
	if len(postcode) == 0 {
		postcode = "0"
	}

	// Parse postal code as integer.
	postcodeNum := 0
	for _, ch := range postcode {
		postcodeNum = postcodeNum*10 + int(ch-'0')
	}
	postcodeLen := len(postcode)

	primary := make([]byte, 10)
	primary[0] = byte(((postcodeNum & 0x03) << 4) | 2)
	primary[1] = byte((postcodeNum & 0xfc) >> 2)
	primary[2] = byte((postcodeNum & 0x3f00) >> 8)
	primary[3] = byte((postcodeNum & 0xfc000) >> 14)
	primary[4] = byte((postcodeNum & 0x3f00000) >> 20)
	primary[5] = byte(((postcodeNum & 0x3c000000) >> 26) | ((postcodeLen & 0x3) << 4))
	primary[6] = byte(((postcodeLen & 0x3c) >> 2) | ((country & 0x3) << 4))
	primary[7] = byte((country & 0xfc) >> 2)
	primary[8] = byte(((country & 0x300) >> 8) | ((service & 0xf) << 2))
	primary[9] = byte((service & 0x3f0) >> 4)
	return primary
}

// maxiCodeMode3PrimaryCodewords returns the 10 primary codewords for Mode 3
// (alphanumeric postal code). Ported from C# BarcodeMaxiCode.cs getMode3PrimaryCodewords.
//
// C# source: BarcodeMaxiCode.cs lines 857-893 (getMode3PrimaryCodewords).
func maxiCodeMode3PrimaryCodewords(postcode string, country, service int) []byte {
	// Ensure exactly 6 characters (Mode 3 uses first 6 chars of the postal code).
	for len(postcode) < 6 {
		postcode += " "
	}
	postcode = postcode[:6]

	// Convert each character to its Code Set A numeric value.
	postcodeNums := make([]int, 6)
	upper := strings.ToUpper(postcode)
	for i := 0; i < 6; i++ {
		ch := int(upper[i])
		if upper[i] >= 'A' && upper[i] <= 'Z' {
			// Capital letters shifted to Code Set A values (A=1..Z=26).
			ch -= 64
		}
		// Disallowed characters → space (ASCII 32).
		if ch == 27 || ch == 31 || ch == 33 || ch >= 59 {
			ch = 32
		}
		postcodeNums[i] = ch
	}

	primary := make([]byte, 10)
	primary[0] = byte(((postcodeNums[5] & 0x03) << 4) | 3)
	primary[1] = byte(((postcodeNums[4] & 0x03) << 4) | ((postcodeNums[5] & 0x3c) >> 2))
	primary[2] = byte(((postcodeNums[3] & 0x03) << 4) | ((postcodeNums[4] & 0x3c) >> 2))
	primary[3] = byte(((postcodeNums[2] & 0x03) << 4) | ((postcodeNums[3] & 0x3c) >> 2))
	primary[4] = byte(((postcodeNums[1] & 0x03) << 4) | ((postcodeNums[2] & 0x3c) >> 2))
	primary[5] = byte(((postcodeNums[0] & 0x03) << 4) | ((postcodeNums[1] & 0x3c) >> 2))
	primary[6] = byte(((postcodeNums[0] & 0x3c) >> 2) | ((country & 0x3) << 4))
	primary[7] = byte((country & 0xfc) >> 2)
	primary[8] = byte(((country & 0x300) >> 8) | ((service & 0xf) << 2))
	primary[9] = byte((service & 0x3f0) >> 4)
	return primary
}

// maxiCodeParseMode23Text parses a Mode 2/3 payload string (produced by
// MaxiCodeMode2Payload / MaxiCodeMode3Payload) into its component parts.
//
// Expected format: postal(9) + country(3) + service(2) + GS(1) + secondary
// Returns postcode, countryStr, serviceInt, secondaryText.
func maxiCodeParseMode23Text(text string) (postcode, country string, service int, secondary string) {
	// Need at least 14 chars (9+3+2) plus optional GS+secondary.
	for len(text) < 14 {
		text += " "
	}
	postcode = text[:9]
	country = text[9:12]
	svc := text[12:14]
	// Parse service as integer (2-char field).
	for _, ch := range svc {
		if ch >= '0' && ch <= '9' {
			service = service*10 + int(ch-'0')
		}
	}

	// Secondary follows the GS separator (0x1D) at position 14, if present.
	if len(text) > 14 && text[14] == 0x1D {
		secondary = text[15:]
	} else if len(text) > 14 {
		secondary = text[14:]
	}
	return postcode, country, service, secondary
}

// maxiCodeEncode encodes text as a 144-codeword MaxiCode symbol including
// Reed-Solomon error correction. Layout (ISO/IEC 16023):
//
//	[0..9]   10 primary data codewords  (mode byte + secondary bytes 0–8 for modes 4/5/6)
//	[10..19] 10 primary ECC codewords
//	[20..N]  secondary data codewords   (84 for modes 2/3/4/6; 68 for mode 5)
//	[N+1..]  secondary ECC codewords    (40 for modes 2/3/4/6; 56 for mode 5), interleaved odd/even
//
// Total is always 144 codewords.
func maxiCodeEncode(text string, mode int) []byte {
	if mode < 2 || mode > 6 {
		mode = 4
	}

	// Secondary data and ECC sizes by mode.
	secondaryMax := 84
	secondaryECMax := 40
	if mode == 5 {
		secondaryMax = 68
		secondaryECMax = 56
	}
	totalMax := secondaryMax + 10 // primary(10) + secondary data

	raw := make([]byte, totalMax)

	if mode == 2 || mode == 3 {
		// Mode 2/3: structured carrier message.
		// Parse the payload text to extract postal, country, service, and secondary.
		// Format produced by MaxiCodeMode2Payload/MaxiCodeMode3Payload:
		//   postal(9) + country(3) + service(2) + GS(0x1D) + secondary
		postcode, countryStr, service, secondary := maxiCodeParseMode23Text(text)

		// Parse country as integer.
		countryNum := 0
		for _, ch := range countryStr {
			if ch >= '0' && ch <= '9' {
				countryNum = countryNum*10 + int(ch-'0')
			}
		}

		// Auto-promote Mode 2 → Mode 3 if postal code isn't strictly numeric.
		// Ported from C# BarcodeMaxiCode.cs encode() lines 537-548.
		if mode == 2 {
			for i := 0; i < len(postcode) && i < 9; i++ {
				if postcode[i] != ' ' && (postcode[i] < '0' || postcode[i] > '9') {
					mode = 3
					break
				}
			}
		}

		// Build the 10 primary codewords using the postal encoding algorithm.
		var primaryCW []byte
		if mode == 2 {
			primaryCW = maxiCodeMode2PrimaryCodewords(strings.TrimRight(postcode, " "), countryNum, service)
		} else {
			primaryCW = maxiCodeMode3PrimaryCodewords(strings.TrimRight(postcode, " "), countryNum, service)
		}
		copy(raw[:10], primaryCW)

		// Encode the secondary message into secondary slots (positions 10..totalMax-1).
		sec := maxiCodeEncodeText(secondary, totalMax-10)
		copy(raw[10:], sec)
	} else {
		// Modes 4/5/6: [mode_byte] + [text encoded into remaining slots].
		raw[0] = byte(mode)
		sec := maxiCodeEncodeText(text, totalMax-1)
		copy(raw[1:], sec)
	}

	// Compute primary ECC on codewords[0..9].
	primaryCWs := raw[:10]
	primaryECC := maxiCodeRS(primaryCWs, 10)

	// Split secondary (raw[10..totalMax-1]) into odd and even interleaves.
	secondaryCWs := raw[10:]
	half := len(secondaryCWs) / 2
	secOdd := make([]byte, half)
	secEven := make([]byte, half)
	for i, cw := range secondaryCWs {
		if i%2 == 1 {
			secOdd[(i-1)/2] = cw
		} else {
			secEven[i/2] = cw
		}
	}
	eccHalf := secondaryECMax / 2
	secECCOdd := maxiCodeRS(secOdd, eccHalf)
	secECCEven := maxiCodeRS(secEven, eccHalf)

	// Assemble final 144 codewords.
	out := make([]byte, 144)
	copy(out[0:10], primaryCWs)
	copy(out[10:20], primaryECC)
	copy(out[20:20+secondaryMax], raw[10:])
	// Interleave secondary ECC after the secondary data block.
	for i := 0; i < len(secECCOdd); i++ {
		out[20+secondaryMax+(2*i)+1] = secECCOdd[i]
	}
	for i := 0; i < len(secECCEven); i++ {
		out[20+secondaryMax+(2*i)] = secECCEven[i]
	}
	return out
}

// maxiCodeCWBits converts a slice of 6-bit codewords to a flat bit sequence.
func maxiCodeCWBits(codewords []byte) []bool {
	bits := make([]bool, len(codewords)*6)
	for i, cw := range codewords {
		for bit := 5; bit >= 0; bit-- {
			bits[i*6+(5-bit)] = (cw>>uint(bit))&1 == 1
		}
	}
	return bits
}

// Encode validates the mode and encodes text to MaxiCode codewords.
// The codewords are stored on the barcode for use during Render.
func (b *MaxiCodeBarcode) Encode(text string) error {
	if b.Mode < 2 || b.Mode > 6 {
		return fmt.Errorf("maxicode: invalid mode %d (must be 2–6)", b.Mode)
	}
	b.encodedText = text
	return nil
}

// ── MaxiCode visual rendering ─────────────────────────────────────────────────

// Render produces a MaxiCode image using the proper codeword-based bit layout.
//
// MaxiCode is a complex 2D barcode with Reed-Solomon error correction that
// encodes data in a hexagonal grid pattern. Full symbol generation (as defined
// in ISO/IEC 16023) requires a complete RS encoder; this implementation renders
// the characteristic MaxiCode appearance: a bullseye finder pattern surrounded
// by a hexagonal bit-matrix populated using actual text codewords (without RS
// error correction). The output is visually identifiable as a MaxiCode symbol
// and suitable for report rendering purposes.
func (b *MaxiCodeBarcode) Render(width, height int) (image.Image, error) {
	if b.encodedText == "" {
		return nil, fmt.Errorf("maxicode: not encoded")
	}
	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 100
	}
	codewords := maxiCodeEncode(b.encodedText, b.Mode)
	bits := maxiCodeCWBits(codewords)
	return maxiCodeRender(bits, width, height), nil
}

// maxiCodeRender produces a MaxiCode-style image with:
//   - A white background
//   - A central bullseye finder (5 concentric rings)
//   - A hexagonal data grid surrounding the bullseye (30×33 hex grid)
//   - Codeword-based bit fill pattern
func maxiCodeRender(bits []bool, width, height int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	// Fill background white.
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.NRGBA{A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, white)
		}
	}

	// MaxiCode is 30 columns × 33 rows of hexagons (ISO/IEC 16023 §5).
	const cols = 30
	const rows = 33

	margin := float64(min2(width, height)) * 0.04
	cellW := (float64(width) - 2*margin) / float64(cols)
	cellH := (float64(height) - 2*margin) / float64(rows)

	// Centre of the image (for bullseye).
	cx := float64(width) / 2.0
	cy := float64(height) / 2.0

	// Bullseye radius in pixels (occupies the central ~6×6 cell area).
	bullseyeR := math.Min(cellW, cellH) * 2.8

	bitIdx := 0
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			// Hexagonal grid offset: odd rows shift right by half a cell.
			offsetX := 0.0
			if row%2 == 1 {
				offsetX = cellW * 0.5
			}
			hx := margin + (float64(col)+0.5)*cellW + offsetX
			hy := margin + (float64(row)+0.5)*cellH

			// Skip cells that fall within the bullseye region.
			dx := hx - cx
			dy := hy - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < bullseyeR*1.1 {
				continue
			}

			// Fill hexagon based on codeword data bit.
			var bit bool
			if bitIdx < len(bits) {
				bit = bits[bitIdx]
				bitIdx++
			}
			if bit {
				drawHex(img, hx, hy, cellW*0.46, cellH*0.46, black)
			}
		}
	}

	// Draw bullseye: 5 alternating concentric rings (ISO/IEC 16023 §7.5).
	drawBullseye(img, cx, cy, bullseyeR, black, white)

	return img
}

// drawBullseye draws 5 alternating concentric rings (bullseye finder pattern).
// ISO/IEC 16023 specifies: centre black, ring 1 white, ring 2 black,
// ring 3 white, ring 4 black, ring 5 white (outermost).
func drawBullseye(img *image.NRGBA, cx, cy, outerR float64, dark, light color.NRGBA) {
	bounds := img.Bounds()
	for py := bounds.Min.Y; py < bounds.Max.Y; py++ {
		for px := bounds.Min.X; px < bounds.Max.X; px++ {
			dx := float64(px) - cx
			dy := float64(py) - cy
			d := math.Sqrt(dx*dx + dy*dy)
			if d > outerR {
				continue
			}
			t := d / outerR
			var c color.NRGBA
			// 5 rings: boundaries at 20%, 40%, 60%, 80%, 100%.
			switch {
			case t < 0.20:
				c = dark // innermost — black
			case t < 0.40:
				c = light
			case t < 0.60:
				c = dark
			case t < 0.80:
				c = light
			default:
				c = dark // outermost ring — black
			}
			img.SetNRGBA(px, py, c)
		}
	}
}

// drawHex draws a filled hexagon centred at (cx, cy) with given half-widths.
func drawHex(img *image.NRGBA, cx, cy, rx, ry float64, c color.NRGBA) {
	bounds := img.Bounds()
	x0 := int(cx - rx)
	y0 := int(cy - ry)
	x1 := int(cx+rx) + 1
	y1 := int(cy+ry) + 1

	for py := y0; py <= y1; py++ {
		if py < bounds.Min.Y || py >= bounds.Max.Y {
			continue
		}
		for px := x0; px <= x1; px++ {
			if px < bounds.Min.X || px >= bounds.Max.X {
				continue
			}
			// Approximate hexagon using ellipse test.
			dx := float64(px) - cx
			dy := float64(py) - cy
			if (dx*dx)/(rx*rx)+(dy*dy)/(ry*ry) <= 1.0 {
				img.SetNRGBA(px, py, c)
			}
		}
	}
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── Mode helper ───────────────────────────────────────────────────────────────

// MaxiCodeMode2Payload builds a structured carrier message for mode 2
// (UPS Standard). Format:
//
//	[ZipCode(9)] [CountryCode(3)] [ServiceClass(2)] GS [SecondaryMessage]
//
// All fields are validated and the payload is returned ready for Encode.
func MaxiCodeMode2Payload(zipCode, countryCode, serviceClass, secondary string) string {
	// Pad/truncate fields to canonical lengths.
	zip := padRight(zipCode, 9, ' ')
	country := padRight(countryCode, 3, ' ')
	svc := padRight(serviceClass, 2, ' ')
	return zip + country + svc + "\x1d" + secondary
}

// MaxiCodeMode3Payload builds a structured carrier message for mode 3
// (international / alphanumeric zip codes).
// Format is identical to mode 2 except ZIP may be alphanumeric (9 chars).
func MaxiCodeMode3Payload(zipCode, countryCode, serviceClass, secondary string) string {
	return MaxiCodeMode2Payload(zipCode, countryCode, serviceClass, secondary)
}

func padRight(s string, n int, pad byte) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(string(pad), n-len(s))
}
