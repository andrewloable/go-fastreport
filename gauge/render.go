package gauge

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/andrewloable/go-fastreport/utils"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func parseColor(s string, def color.RGBA) color.RGBA {
	if s == "" {
		return def
	}
	c, err := utils.ParseColor(s)
	if err != nil {
		return def
	}
	return c
}

// fillRect fills a rectangle on img with the given color.
func fillRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	r := image.Rect(x, y, x+w, y+h).Intersect(img.Bounds())
	draw.Draw(img, r, &image.Uniform{c}, image.Point{}, draw.Src)
}

// drawHLine draws a horizontal line.
func drawHLine(img *image.RGBA, x0, y, x1 int, c color.RGBA) {
	b := img.Bounds()
	if y < b.Min.Y || y >= b.Max.Y {
		return
	}
	for x := x0; x < x1; x++ {
		if x >= b.Min.X && x < b.Max.X {
			img.SetRGBA(x, y, c)
		}
	}
}

// drawVLine draws a vertical line.
func drawVLine(img *image.RGBA, x, y0, y1 int, c color.RGBA) {
	b := img.Bounds()
	if x < b.Min.X || x >= b.Max.X {
		return
	}
	for y := y0; y < y1; y++ {
		if y >= b.Min.Y && y < b.Max.Y {
			img.SetRGBA(x, y, c)
		}
	}
}

// drawRectBorder draws a 1-pixel border around the given area.
func drawRectBorder(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	drawHLine(img, x, y, x+w, c)
	drawHLine(img, x, y+h-1, x+w, c)
	drawVLine(img, x, y, y+h, c)
	drawVLine(img, x+w-1, y, y+h, c)
}

// drawLine draws a Bresenham line between two points.
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	b := img.Bounds()
	for {
		if x0 >= b.Min.X && x0 < b.Max.X && y0 >= b.Min.Y && y0 < b.Max.Y {
			img.SetRGBA(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// drawArc draws an arc of the ellipse (cx,cy,rx,ry) from startDeg to endDeg.
func drawArc(img *image.RGBA, cx, cy, rx, ry int, startDeg, endDeg float64, c color.RGBA) {
	if rx <= 0 || ry <= 0 {
		return
	}
	sweep := endDeg - startDeg
	if sweep < 0 {
		sweep += 360
	}
	circumference := math.Pi * (float64(rx) + float64(ry))
	steps := int(circumference * sweep / 180)
	if steps < 16 {
		steps = 16
	}
	b := img.Bounds()
	for i := 0; i <= steps; i++ {
		t := startDeg + sweep*float64(i)/float64(steps)
		rad := t * math.Pi / 180
		px := cx + int(math.Round(float64(rx)*math.Cos(rad)))
		py := cy + int(math.Round(float64(ry)*math.Sin(rad)))
		if px >= b.Min.X && px < b.Max.X && py >= b.Min.Y && py < b.Max.Y {
			img.SetRGBA(px, py, c)
		}
	}
}

// ── LinearGauge ───────────────────────────────────────────────────────────────

var (
	colorLightGray = color.RGBA{R: 220, G: 220, B: 220, A: 255}
	colorDarkGray  = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	colorBlack     = color.RGBA{A: 255}
)

// RenderLinear renders a LinearGauge into an RGBA image of size (w,h).
// The result is a horizontal or vertical progress bar.
func RenderLinear(g *LinearGauge, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// Background fill.
	fillRect(img, 0, 0, w, h, colorLightGray)

	pct := g.FillPercent()
	pointerColor := parseColor(g.Pointer.Color, color.RGBA{R: 204, A: 255})

	const margin = 4

	if g.Orientation == OrientationHorizontal {
		barX := margin
		barY := margin
		barW := w - 2*margin
		barH := h - 2*margin
		if barW <= 0 || barH <= 0 {
			return img
		}
		// Fill bar.
		fillW := int(math.Round(float64(barW) * pct))
		if fillW > 0 {
			fillRect(img, barX, barY, fillW, barH, pointerColor)
		}
		drawRectBorder(img, barX, barY, barW, barH, colorDarkGray)
	} else {
		// Vertical.
		barX := margin
		barY := margin
		barW := w - 2*margin
		barH := h - 2*margin
		if barW <= 0 || barH <= 0 {
			return img
		}
		fillH := int(math.Round(float64(barH) * pct))
		// Vertical bar fills from the bottom.
		if fillH > 0 {
			fillRect(img, barX, barY+barH-fillH, barW, fillH, pointerColor)
		}
		drawRectBorder(img, barX, barY, barW, barH, colorDarkGray)
	}
	return img
}

// ── RadialGauge ───────────────────────────────────────────────────────────────

// RenderRadial renders a RadialGauge into an RGBA image of size (w,h).
// It draws a circular arc and a needle pointing to the current value.
func RenderRadial(g *RadialGauge, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fillRect(img, 0, 0, w, h, colorLightGray)

	cx := w / 2
	cy := h / 2
	rx := w/2 - 4
	ry := h/2 - 4
	if rx <= 0 || ry <= 0 {
		return img
	}

	// Draw the scale arc (start to end angles).
	drawArc(img, cx, cy, rx, ry, g.StartAngle, g.EndAngle, colorDarkGray)
	// Draw a thicker inner arc for visual weight.
	if rx > 4 && ry > 4 {
		drawArc(img, cx, cy, rx-1, ry-1, g.StartAngle, g.EndAngle, colorDarkGray)
	}

	// Draw needle from center toward the arc.
	needleAngle := g.NeedleAngle()
	rad := needleAngle * math.Pi / 180
	needleLen := int(math.Round(float64(min(rx, ry)) * 0.85))
	nx := cx + int(math.Round(float64(needleLen)*math.Cos(rad)))
	ny := cy + int(math.Round(float64(needleLen)*math.Sin(rad)))

	pointerColor := parseColor(g.Pointer.Color, color.RGBA{R: 204, A: 255})
	drawLine(img, cx, cy, nx, ny, pointerColor)
	// Draw the needle slightly thicker.
	if cx+1 < w {
		drawLine(img, cx+1, cy, nx, ny, pointerColor)
	}
	if cy+1 < h {
		drawLine(img, cx, cy+1, nx, ny, pointerColor)
	}

	// Center dot.
	fillRect(img, cx-2, cy-2, 4, 4, colorBlack)

	return img
}

// ── SimpleGauge ───────────────────────────────────────────────────────────────

// RenderSimple renders a SimpleGauge into an RGBA image of size (w,h).
// The fill percentage colors the shape background.
func RenderSimple(g *SimpleGauge, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fillRect(img, 0, 0, w, h, colorLightGray)

	pointerColor := parseColor(g.Pointer.Color, color.RGBA{R: 204, A: 255})
	pct := g.Percent()

	const margin = 4

	switch g.Shape {
	case SimpleGaugeShapeCircle:
		// Draw partial-fill circle: fill a rectangle clipped to the ellipse arc.
		cx := w / 2
		cy := h / 2
		rx := w/2 - margin
		ry := h/2 - margin
		if rx > 0 && ry > 0 {
			// Fill ellipse outline with pointer color at the fill percentage.
			steps := 360
			b := img.Bounds()
			fillAngle := pct * 360 // degrees to fill clockwise from top
			for i := 0; i < steps; i++ {
				t := float64(i)
				rad := (t - 90) * math.Pi / 180
				px := cx + int(math.Round(float64(rx)*math.Cos(rad)))
				py := cy + int(math.Round(float64(ry)*math.Sin(rad)))
				if px >= b.Min.X && px < b.Max.X && py >= b.Min.Y && py < b.Max.Y {
					if t <= fillAngle {
						img.SetRGBA(px, py, pointerColor)
					} else {
						img.SetRGBA(px, py, colorDarkGray)
					}
				}
			}
		}
	case SimpleGaugeShapeTriangle:
		// Filled triangle pointing up.
		tx := w / 2
		ty := margin
		bl := margin
		br := w - margin
		bm := h - margin
		// Simple scan-fill approach: draw lines between edges.
		for i := 0; i < 20; i++ {
			frac := float64(i) / 20.0
			x0 := tx + int(math.Round(float64(bl-tx)*frac))
			x1 := tx + int(math.Round(float64(br-tx)*frac))
			y := ty + int(math.Round(float64(bm-ty)*frac))
			c := colorDarkGray
			if frac <= pct {
				c = pointerColor
			}
			drawHLine(img, x0, y, x1, c)
		}
	default: // SimpleGaugeShapeRectangle
		barX := margin
		barY := margin
		barW := w - 2*margin
		barH := h - 2*margin
		if barW > 0 && barH > 0 {
			fillH := int(math.Round(float64(barH) * pct))
			if fillH > 0 {
				fillRect(img, barX, barY+barH-fillH, barW, fillH, pointerColor)
			}
			drawRectBorder(img, barX, barY, barW, barH, colorDarkGray)
		}
	}
	return img
}

// ── SimpleProgressGauge ───────────────────────────────────────────────────────

// RenderSimpleProgress renders a SimpleProgressGauge as a horizontal progress bar.
func RenderSimpleProgress(g *SimpleProgressGauge, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fillRect(img, 0, 0, w, h, colorLightGray)

	pct := g.Percent()
	pointerColor := parseColor(g.Pointer.Color, color.RGBA{R: 204, A: 255})

	const margin = 4
	barX := margin
	barY := margin
	barW := w - 2*margin
	barH := h - 2*margin
	if barW <= 0 || barH <= 0 {
		return img
	}
	fillW := int(math.Round(float64(barW) * pct))
	if fillW > 0 {
		fillRect(img, barX, barY, fillW, barH, pointerColor)
	}
	drawRectBorder(img, barX, barY, barW, barH, colorDarkGray)
	return img
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
