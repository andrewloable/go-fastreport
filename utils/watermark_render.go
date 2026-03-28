package utils

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/andrewloable/go-fastreport/style"
)

// Watermark text rotation constants matching preview.WatermarkTextRotation.
const (
	WatermarkRotationHorizontal      = 0
	WatermarkRotationVertical        = 1
	WatermarkRotationForwardDiagonal = 2
	WatermarkRotationBackward        = 3
)

// RenderWatermarkText rasterizes watermark text onto a transparent RGBA image,
// matching C# Watermark.DrawText() which renders via TextObject.DrawText onto
// a Graphics surface (Watermark.cs line 253-276). The C# HTML exporter then
// embeds this rasterised image as a PictureObject.
//
// Since Go uses basicfont.Face7x13 as fallback (13px tall), the text is first
// drawn at basicfont scale onto a small image, then scaled up to the target
// font size and composited onto the output image with rotation.
func RenderWatermarkText(text string, f style.Font, textColor color.RGBA, rotation int, w, h int) image.Image {
	if text == "" || w < 1 || h < 1 {
		return nil
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	face := faceForStyle(f)
	if face == nil {
		return img
	}

	// Target font height in pixels.
	fontPx := float64(f.Size) * 96.0 / 72.0
	if fontPx < 1 {
		fontPx = 48 * 96.0 / 72.0
	}

	// basicfont.Face7x13 metrics.
	metrics := face.Metrics()
	basicHeight := float64(metrics.Height) / 64.0 // ~13px
	if basicHeight < 1 {
		basicHeight = 13
	}
	basicAscent := float64(metrics.Ascent) / 64.0

	// Scale factor from basicfont to target font.
	scale := fontPx / basicHeight

	// Measure text at basicfont scale.
	textWidth := font.MeasureString(face, text)
	textWidthPx := float64(textWidth) / 64.0

	// Render text onto a small image at basicfont scale with some padding.
	pad := 2
	smallW := int(math.Ceil(textWidthPx)) + pad*2
	smallH := int(math.Ceil(basicHeight)) + pad*2
	if smallW < 1 || smallH < 1 {
		return img
	}
	small := image.NewRGBA(image.Rect(0, 0, smallW, smallH))
	// Draw at full opacity first so bilinear scaling produces good anti-aliased
	// edges. The target alpha is applied after rotation.
	opaqueColor := color.RGBA{R: textColor.R, G: textColor.G, B: textColor.B, A: 255}
	drawString(small, face, text, float64(pad), float64(pad)+basicAscent, opaqueColor)

	// Scale up the text to the target font size.
	scaledW := int(math.Round(float64(smallW) * scale))
	scaledH := int(math.Round(float64(smallH) * scale))
	if scaledW < 1 || scaledH < 1 {
		return img
	}
	scaled := image.NewRGBA(image.Rect(0, 0, scaledW, scaledH))
	draw.BiLinear.Scale(scaled, scaled.Bounds(), small, small.Bounds(), draw.Over, nil)

	// Compute the angle matching C# Watermark.DrawText (Watermark.cs line 257-271).
	var angleDeg float64
	switch rotation {
	case WatermarkRotationVertical:
		angleDeg = 270
	case WatermarkRotationForwardDiagonal:
		angleDeg = 360 - math.Atan(float64(h)/float64(w))*180/math.Pi
	case WatermarkRotationBackward:
		angleDeg = math.Atan(float64(h)/float64(w)) * 180 / math.Pi
	default:
		angleDeg = 0
	}

	// Composite the scaled text onto the output image, centered and rotated.
	// Use inverse mapping (iterate dest pixels, sample source) to avoid gaps.
	cx := float64(w) / 2
	cy := float64(h) / 2
	rad := angleDeg * math.Pi / 180
	// Inverse rotation: rotate dest coords back to source coords.
	cosA := math.Cos(-rad)
	sinA := math.Sin(-rad)

	halfSW := float64(scaledW) / 2
	halfSH := float64(scaledH) / 2
	for oy := 0; oy < h; oy++ {
		for ox := 0; ox < w; ox++ {
			// Position relative to output center.
			dx := float64(ox) - cx
			dy := float64(oy) - cy
			// Inverse-rotate to find source coordinates.
			sx := dx*cosA - dy*sinA + halfSW
			sy := dx*sinA + dy*cosA + halfSH
			// Bilinear sample from the scaled text image.
			ix := int(math.Floor(sx))
			iy := int(math.Floor(sy))
			if ix < 0 || ix >= scaledW-1 || iy < 0 || iy >= scaledH-1 {
				continue
			}
			fx := sx - float64(ix)
			fy := sy - float64(iy)
			c00 := scaled.RGBAAt(ix, iy)
			c10 := scaled.RGBAAt(ix+1, iy)
			c01 := scaled.RGBAAt(ix, iy+1)
			c11 := scaled.RGBAAt(ix+1, iy+1)
			a := bilinear(float64(c00.A), float64(c10.A), float64(c01.A), float64(c11.A), fx, fy)
			if a < 0.5 {
				continue
			}
			r := bilinear(float64(c00.R), float64(c10.R), float64(c01.R), float64(c11.R), fx, fy)
			g := bilinear(float64(c00.G), float64(c10.G), float64(c01.G), float64(c11.G), fx, fy)
			b := bilinear(float64(c00.B), float64(c10.B), float64(c01.B), float64(c11.B), fx, fy)
			// Apply the target text alpha (scaled text was drawn at full opacity
			// for better anti-aliasing during bilinear scale/rotation).
			finalA := a * float64(textColor.A) / 255.0
			img.SetRGBA(ox, oy, color.RGBA{
				R: uint8(math.Round(r)),
				G: uint8(math.Round(g)),
				B: uint8(math.Round(b)),
				A: uint8(math.Round(finalA)),
			})
		}
	}

	return img
}

func bilinear(c00, c10, c01, c11, fx, fy float64) float64 {
	return c00*(1-fx)*(1-fy) + c10*fx*(1-fy) + c01*(1-fx)*fy + c11*fx*fy
}

// drawString draws text at (x, y baseline) on img.
func drawString(img *image.RGBA, face font.Face, text string, x, y float64, c color.RGBA) {
	dot := fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6(y * 64),
	}
	for _, r := range text {
		dr, mask, maskp, advance, ok := face.Glyph(dot, r)
		if ok {
			drawGlyph(img, dr, mask, maskp, c)
		}
		dot.X += advance
	}
}

// drawGlyph composites a single glyph onto img with the given color.
func drawGlyph(img *image.RGBA, dr image.Rectangle, mask image.Image, maskp image.Point, c color.RGBA) {
	b := img.Bounds().Intersect(dr)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			mx := x - dr.Min.X + maskp.X
			my := y - dr.Min.Y + maskp.Y
			_, _, _, ma := mask.At(mx, my).RGBA()
			if ma == 0 {
				continue
			}
			// Apply text color with mask alpha.
			alpha := uint8(uint32(c.A) * ma / 0xFFFF)
			img.SetRGBA(x, y, color.RGBA{R: c.R, G: c.G, B: c.B, A: alpha})
		}
	}
}
