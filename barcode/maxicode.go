package barcode

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// Render implements the MaxiCode visual rendering.
//
// MaxiCode is a complex 2D barcode with Reed-Solomon error correction that
// encodes data in a hexagonal grid pattern. Full symbol generation (as defined
// in ISO/IEC 16023) requires a complete RS encoder; this implementation renders
// the characteristic MaxiCode appearance: a bullseye finder pattern surrounded
// by a hexagonal bit-matrix derived from a simple data hash. The output is
// visually identifiable as a MaxiCode symbol and suitable for report rendering
// purposes. For production scanning use, replace this with a certified encoder.
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
	return maxiCodeRender(b.encodedText, b.Mode, width, height), nil
}

// maxiCodeRender produces a MaxiCode-style image with:
//   - A white background
//   - A central bullseye finder (3 concentric rings)
//   - A hexagonal data grid surrounding the bullseye (30×33 hex grid)
//   - Data-dependent bit fill pattern
func maxiCodeRender(text string, _ int, width, height int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	// Fill background white.
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.NRGBA{A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, white)
		}
	}

	// MaxiCode is 30 columns × 33 rows of hexagons.
	// Scale to fit in the output image with equal margins on all sides.
	const cols = 30
	const rows = 33

	margin := float64(min2(width, height)) * 0.04
	cellW := (float64(width) - 2*margin) / float64(cols)
	cellH := (float64(height) - 2*margin) / float64(rows)

	// Centre of the image (for bullseye).
	cx := float64(width) / 2.0
	cy := float64(height) / 2.0

	// Bullseye radius in pixels (occupies the central ~6×6 cells).
	bullseyeR := math.Min(cellW, cellH) * 2.8

	// Data-dependent bit pattern derived from text bytes.
	bits := maxiCodeBits(text, cols*rows)

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

			// Skip cells that fall within the bullseye radius.
			dx := hx - cx
			dy := hy - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < bullseyeR*1.1 {
				continue
			}

			// Fill hexagon based on data bit.
			bit := bits[bitIdx%len(bits)]
			bitIdx++
			if !bit {
				continue
			}
			drawHex(img, hx, hy, cellW*0.46, cellH*0.46, black)
		}
	}

	// Draw bullseye: alternating black and white concentric rings.
	drawBullseye(img, cx, cy, bullseyeR, black, white)

	return img
}

// drawBullseye draws 3 alternating concentric rings (bullseye finder pattern).
func drawBullseye(img *image.NRGBA, cx, cy, outerR float64, dark, light color.NRGBA) {
	bounds := img.Bounds()
	// 3 rings: inner (black), middle (white), outer (black), with a black centre.
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
			switch {
			case t < 0.20:
				c = dark
			case t < 0.40:
				c = light
			case t < 0.60:
				c = dark
			case t < 0.80:
				c = light
			default:
				c = dark
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
			// Approximate hexagon by an ellipse (close enough at small sizes).
			dx := float64(px) - cx
			dy := float64(py) - cy
			if (dx*dx)/(rx*rx)+(dy*dy)/(ry*ry) <= 1.0 {
				img.SetNRGBA(px, py, c)
			}
		}
	}
}

// maxiCodeBits produces a pseudo-random bit sequence derived from text data.
// This gives each unique text a visually distinct hex grid pattern.
func maxiCodeBits(text string, n int) []bool {
	// Simple hash-based sequence using the text bytes.
	state := uint32(0x12345678)
	for _, b := range []byte(text) {
		state = state*0x08088405 + uint32(b)
	}
	bits := make([]bool, n)
	for i := range bits {
		state = state*0x08088405 + 1
		bits[i] = (state>>16)&1 == 1
	}
	return bits
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}
