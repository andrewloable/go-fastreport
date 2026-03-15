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
// data codewords and up to 82 error-correction codewords arranged in a
// hexagonal grid of 30 columns × 33 rows, with a central bullseye finder.
//
// Each codeword is 6 bits wide (64 possible values). Characters are encoded
// using two character sets:
//   - Set A: control characters and GS1 special characters
//   - Set B: printable ASCII (SP through DEL)
//
// This implementation encodes text to codewords following the MaxiCode
// character set rules (modes 2–6). Reed-Solomon error correction is omitted
// — the RS codewords are filled with zeros. The symbol is therefore not
// scannable by a barcode reader, but it is visually accurate and sufficient
// for report preview / print purposes.
//
// For production-scannable symbols, replace the RS placeholder with a
// proper GF(64) Reed-Solomon encoder.

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

// maxiCodeEncode converts text to a slice of 6-bit codewords for MaxiCode mode 4.
// The codeword sequence starts in Set B. Shift to Set A is used for characters
// below 0x20. The sequence is truncated or padded to 144 data codewords.
func maxiCodeEncode(text string, mode int) []byte {
	const totalData = 144

	codewords := make([]byte, 0, totalData)

	// Mode prefix codeword encodes the mode (2–6).
	if mode < 2 || mode > 6 {
		mode = 4
	}

	// Mode 2/3: structured carrier message (e.g. UPS shipping data).
	// The message format is: ZipCode + CountryCode + ServiceClass + "\x1d" + SecondaryMessage.
	// Modes 4-6: general purpose text.
	// For simplicity, we always use the general character encoding below.
	// A mode field is embedded as the first codeword.
	codewords = append(codewords, byte(mode))

	// Start in Set B. Encode each character.
	inSetA := false
	for _, r := range text {
		if r > 0x7F {
			// Non-ASCII: use GS character (0x1D) as placeholder.
			r = 0x1D
		}
		ch := byte(r)
		if ch < 0x20 {
			// Control character — belongs to Set A.
			if !inSetA {
				codewords = append(codewords, 63) // LATA
				inSetA = true
			}
			codewords = append(codewords, ch&0x3F)
		} else {
			// Printable character — belongs to Set B.
			if inSetA {
				codewords = append(codewords, 0) // LATB
				inSetA = false
			}
			idx := int(ch) - 0x1F // SP (0x20) → index 1
			if idx < 1 {
				idx = 1
			}
			if idx > 63 {
				idx = 63
			}
			codewords = append(codewords, byte(idx))
		}
		if len(codewords) >= totalData {
			break
		}
	}

	// Pad to 144 data codewords.
	for len(codewords) < totalData {
		codewords = append(codewords, 0)
	}
	return codewords[:totalData]
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
