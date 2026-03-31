package gauge

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/andrewloable/go-fastreport/style"
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

// ── Text helpers ──────────────────────────────────────────────────────────────

// gaugeFont returns a font.Face for the given FRX font descriptor string.
// Falls back to basicfont.Face7x13 if the font cannot be loaded.
func gaugeFont(fontStr string) xfont.Face {
	if fontStr == "" {
		return basicfont.Face7x13
	}
	f := style.FontFromStr(fontStr)
	desc := utils.FontDescriptor{
		Family: f.Name,
		Size:   f.Size,
		Style:  utils.FontStyle(f.Style),
	}
	face := utils.DefaultFontManager.GetFace(desc)
	if face == nil {
		return basicfont.Face7x13
	}
	return face
}

// textWidthPx returns the pixel width of text rendered with face.
func textWidthPx(face xfont.Face, text string) int {
	w := xfont.MeasureString(face, text)
	return int(math.Ceil(float64(w) / 64.0))
}

// textHeightPx returns the pixel line height of face.
func textHeightPx(face xfont.Face) int {
	m := face.Metrics()
	return int(math.Ceil(float64(m.Height) / 64.0))
}

// textAscentPx returns the ascent in pixels of face.
func textAscentPx(face xfont.Face) int {
	m := face.Metrics()
	return int(math.Ceil(float64(m.Ascent) / 64.0))
}

// drawGaugeText draws text on img at (x, y) where y is the TOP of the text
// (matching C# DrawString which also takes the top-left corner).
func drawGaugeText(img *image.RGBA, face xfont.Face, text string, x, y int, c color.RGBA) {
	ascent := textAscentPx(face)
	dot := fixed.P(x, y+ascent)
	b := img.Bounds()
	for _, r := range text {
		dr, mask, maskp, advance, ok := face.Glyph(dot, r)
		if ok {
			clip := b.Intersect(dr)
			for py := clip.Min.Y; py < clip.Max.Y; py++ {
				for px := clip.Min.X; px < clip.Max.X; px++ {
					mx := px - dr.Min.X + maskp.X
					my := py - dr.Min.Y + maskp.Y
					_, _, _, ma := mask.At(mx, my).RGBA()
					if ma > 0 {
						alpha := uint8(uint32(c.A) * ma / 0xFFFF)
						img.SetRGBA(px, py, color.RGBA{R: c.R, G: c.G, B: c.B, A: alpha})
					}
				}
			}
		}
		dot.X += advance
	}
}

// formatGaugeLabel formats a gauge value as a scale label string.
// Integer values are formatted without a decimal point.
func formatGaugeLabel(v float64) string {
	if v == math.Trunc(v) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%g", v)
}

// ── Arrow pointer fill ────────────────────────────────────────────────────────

// fillDownArrow fills a downward-pointing arrow/chevron shape.
// The tip is at (px, py) and the body widens downward to (px±hw, py+h).
// The triangular tip section covers the top 30% of the height.
//
// C# source: original-dotnet/FastReport.Base/Gauge/Linear/LinearPointer.cs
// DrawHorz method — pentagon points p[0]..p[4] (non-inverted).
func fillDownArrow(img *image.RGBA, px, py, hw, h int, c color.RGBA) {
	if h <= 0 || hw <= 0 {
		return
	}
	tipH := int(math.Round(float64(h) * 0.3))
	b := img.Bounds()
	for y := py; y < py+h; y++ {
		if y < b.Min.Y || y >= b.Max.Y {
			continue
		}
		var x0, x1 int
		if y < py+tipH && tipH > 0 {
			frac := float64(y-py) / float64(tipH)
			off := int(math.Round(float64(hw) * frac))
			x0 = px - off
			x1 = px + off
		} else {
			x0 = px - hw
			x1 = px + hw
		}
		for x := x0; x <= x1; x++ {
			if x >= b.Min.X && x < b.Max.X {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

// fillUpArrow fills an upward-pointing arrow/chevron shape.
// The tip is at (px, py) and the body widens upward to (px±hw, py-h).
// Used for Inverted LinearGauge.
//
// C# source: original-dotnet/FastReport.Base/Gauge/Linear/LinearPointer.cs
// DrawHorz method — pentagon points with Inverted=true.
func fillUpArrow(img *image.RGBA, px, py, hw, h int, c color.RGBA) {
	if h <= 0 || hw <= 0 {
		return
	}
	tipH := int(math.Round(float64(h) * 0.3))
	b := img.Bounds()
	for y := py - h; y <= py; y++ {
		if y < b.Min.Y || y >= b.Max.Y {
			continue
		}
		var x0, x1 int
		if y >= py-tipH && tipH > 0 {
			frac := float64(py-y) / float64(tipH)
			off := int(math.Round(float64(hw) * frac))
			x0 = px - off
			x1 = px + off
		} else {
			x0 = px - hw
			x1 = px + hw
		}
		for x := x0; x <= x1; x++ {
			if x >= b.Min.X && x < b.Max.X {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

// ── Constants ─────────────────────────────────────────────────────────────────

const cmPx = 37.8 // pixels per cm at 96 DPI (units.Centimeters)

var (
	colorWhite    = color.RGBA{R: 255, G: 255, B: 255, A: 255} // default gauge background (C# GDI+ white canvas)
	colorDarkGray = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	colorBlack    = color.RGBA{A: 255}
	colorOrange   = color.RGBA{R: 255, G: 165, A: 255} // Color.Orange default
)

// gaugeBgColor returns the background colour for a gauge image.
// If the gauge has a non-transparent solid fill, that colour is used.
// Otherwise returns white (matching C# rendering on a white GDI+ canvas).
func gaugeBgColor(fill style.Fill) color.RGBA {
	if sf, ok := fill.(*style.SolidFill); ok && sf.Color.A > 0 {
		return sf.Color
	}
	return colorWhite
}

// fillEllipse fills an axis-aligned ellipse centred at (cx,cy) with radii (rx,ry).
func fillEllipse(img *image.RGBA, cx, cy, rx, ry int, c color.RGBA) {
	if rx <= 0 || ry <= 0 {
		return
	}
	b := img.Bounds()
	ry2 := float64(ry * ry)
	rx2 := float64(rx * rx)
	for dy := -ry; dy <= ry; dy++ {
		y := cy + dy
		if y < b.Min.Y || y >= b.Max.Y {
			continue
		}
		frac := float64(dy*dy) / ry2
		if frac > 1 {
			continue
		}
		xSpan := int(math.Round(math.Sqrt(rx2 * (1.0 - frac))))
		for x := cx - xSpan; x <= cx+xSpan; x++ {
			if x >= b.Min.X && x < b.Max.X {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

// fillConvexPoly fills a convex polygon given as a slice of (x,y) pairs using
// per-scanline min/max tracking.
func fillConvexPoly(img *image.RGBA, pts [][2]int, c color.RGBA) {
	if len(pts) < 3 {
		return
	}
	b := img.Bounds()
	// Find y extents.
	minY, maxY := pts[0][1], pts[0][1]
	for _, p := range pts {
		if p[1] < minY {
			minY = p[1]
		}
		if p[1] > maxY {
			maxY = p[1]
		}
	}
	n := len(pts)
	for y := minY; y <= maxY; y++ {
		if y < b.Min.Y || y >= b.Max.Y {
			continue
		}
		xMin, xMax := int(1e9), int(-1e9)
		for i := range pts {
			j := (i + 1) % n
			y0, y1 := pts[i][1], pts[j][1]
			x0, x1 := pts[i][0], pts[j][0]
			if (y0 <= y && y < y1) || (y1 <= y && y < y0) {
				// Linear interpolation.
				t := float64(y-y0) / float64(y1-y0)
				x := x0 + int(math.Round(t*float64(x1-x0)))
				if x < xMin {
					xMin = x
				}
				if x > xMax {
					xMax = x
				}
			}
		}
		for x := xMin; x <= xMax; x++ {
			if x >= b.Min.X && x < b.Max.X {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

// ── LinearGauge ───────────────────────────────────────────────────────────────

// RenderLinear renders a LinearGauge into an RGBA image of size (w,h).
// The result is a scale ruler with major/minor tick marks, value labels, and
// an arrow pointer at the current value position.
//
// C# source: original-dotnet/FastReport.Base/Gauge/Linear/LinearGauge.cs Draw()
// original-dotnet/FastReport.Base/Gauge/Linear/LinearScale.cs
// original-dotnet/FastReport.Base/Gauge/Linear/LinearPointer.cs
func RenderLinear(g *LinearGauge, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fillRect(img, 0, 0, w, h, gaugeBgColor(g.Fill()))

	majorTickColor := parseColor(g.Scale.MajorTicks.Color, colorDarkGray)
	minorTickColor := parseColor(g.Scale.MinorTicks.Color, colorDarkGray)
	pointerColor := parseColor(g.Pointer.Color, colorOrange)
	face := gaugeFont(g.Scale.Font)

	// C# LinearScale: majorTicksNum = 6 (hardcoded in constructor)
	// original-dotnet/FastReport.Base/Gauge/Linear/LinearScale.cs line 43.
	const majorCount = 6

	if g.Orientation == OrientationHorizontal {
		// C# LinearScale.Draw (horizontal):
		// left = (AbsLeft + 0.5cm) * scaleX  (AbsLeft=0 in our renderer)
		// top = (AbsTop + 0.6cm) * scaleY
		// width = (Width - 1.0cm) * scaleX
		// height = (Height - 1.2cm) * scaleY
		// original-dotnet/FastReport.Base/Gauge/Linear/LinearScale.cs lines 177-184.
		scaleLeft := int(math.Round(0.5 * cmPx))
		scaleTop := int(math.Round(0.6 * cmPx))
		scaleW := w - int(math.Round(1.0*cmPx))
		scaleH := h - int(math.Round(1.2*cmPx))
		if scaleW <= 0 || scaleH <= 0 {
			return img
		}

		step := float64(scaleW) / float64(majorCount-1)
		valueStep := (g.Maximum - g.Minimum) / float64(majorCount-1)

		// Minor ticks: 3 between each major pair, 20% inset from top/bottom.
		// C# DrawMinorTicksHorz: y1=top+height*0.2, y2=top+height-height*0.2
		// original-dotnet/FastReport.Base/Gauge/Linear/LinearScale.cs lines 79-95.
		y1minor := scaleTop + int(math.Round(float64(scaleH)*0.2))
		y2minor := scaleTop + scaleH - int(math.Round(float64(scaleH)*0.2))
		minorStep := step / 4.0
		for i := 0; i < majorCount-1; i++ {
			for j := 1; j <= 3; j++ {
				mx := scaleLeft + int(math.Round(float64(i)*step+float64(j)*minorStep))
				drawVLine(img, mx, y1minor, y2minor, minorTickColor)
			}
		}

		// Major ticks and labels.
		// C# DrawMajorTicksHorz: ticks from y1 to y2; labels at y = y1 - 0.4cm.
		// original-dotnet/FastReport.Base/Gauge/Linear/LinearScale.cs lines 50-76.
		lh := textHeightPx(face)
		for i := 0; i < majorCount; i++ {
			tx := scaleLeft + int(math.Round(float64(i)*step))
			drawVLine(img, tx, scaleTop, scaleTop+scaleH, majorTickColor)

			val := g.Minimum + valueStep*float64(i)
			lbl := formatGaugeLabel(val)
			lw := textWidthPx(face, lbl)

			if g.Inverted {
				// Inverted: labels below the scale
				// y3 = y2 - fontHeight + 0.4cm (C# LinearScale.cs line 67)
				labelY := scaleTop + scaleH - lh + int(math.Round(0.4*cmPx))
				drawGaugeText(img, face, lbl, tx-lw/2, labelY, colorBlack)
			} else {
				// Normal: labels above the scale at top-of-text = y1 - 0.4cm
				// C# DrawString takes top-left; drawGaugeText also takes top-left.
				// original-dotnet/FastReport.Base/Gauge/Linear/LinearScale.cs line 63.
				labelY := scaleTop - int(math.Round(0.4*cmPx))
				drawGaugeText(img, face, lbl, tx-lw/2, labelY, colorBlack)
			}
		}

		// Pointer: arrow at current value position.
		// C# LinearPointer.DrawHorz:
		// left = AbsLeft + 0.5cm + (Width - 1.0cm) * percent  → scaleLeft + scaleW*percent
		// top = AbsTop + Height/2  → h/2
		// height = Height * 0.4
		// width = Width * 0.036
		// original-dotnet/FastReport.Base/Gauge/Linear/LinearPointer.cs lines 67-101.
		pct := g.FillPercent()
		px := scaleLeft + int(math.Round(float64(scaleW)*pct))
		py := h / 2
		pHeight := int(math.Round(float64(h) * 0.4))
		pHalfW := int(math.Round(float64(w) * 0.036 / 2))
		if pHalfW < 2 {
			pHalfW = 2
		}
		if g.Inverted {
			fillUpArrow(img, px, py, pHalfW, pHeight, pointerColor)
		} else {
			fillDownArrow(img, px, py, pHalfW, pHeight, pointerColor)
		}
	} else {
		// Vertical orientation.
		// C# LinearScale.Draw (vertical):
		// left = (AbsLeft + 0.7cm) * scaleX
		// top = (AbsTop + 0.5cm) * scaleY
		// height = (Height - 1.0cm) * scaleY
		// width = (Width - 1.4cm) * scaleX
		// original-dotnet/FastReport.Base/Gauge/Linear/LinearScale.cs lines 166-184.
		scaleLeft := int(math.Round(0.7 * cmPx))
		scaleTop := int(math.Round(0.5 * cmPx))
		scaleW := w - int(math.Round(1.4*cmPx))
		scaleH := h - int(math.Round(1.0*cmPx))
		if scaleW <= 0 || scaleH <= 0 {
			return img
		}

		step := float64(scaleH) / float64(majorCount-1)
		valueStep := (g.Maximum - g.Minimum) / float64(majorCount-1)

		// Minor ticks.
		x1minor := scaleLeft + int(math.Round(float64(scaleW)*0.2))
		x2minor := scaleLeft + scaleW - int(math.Round(float64(scaleW)*0.2))
		minorStep := step / 4.0
		for i := 0; i < majorCount-1; i++ {
			for j := 1; j <= 3; j++ {
				my := scaleTop + scaleH - int(math.Round(float64(i)*step+float64(j)*minorStep))
				drawHLine(img, x1minor, my, x2minor, minorTickColor)
			}
		}

		// Major ticks and labels.
		lh := textHeightPx(face)
		for i := 0; i < majorCount; i++ {
			ty := scaleTop + scaleH - int(math.Round(float64(i)*step))
			drawHLine(img, scaleLeft, ty, scaleLeft+scaleW, majorTickColor)

			val := g.Minimum + valueStep*float64(i)
			lbl := formatGaugeLabel(val)
			lw := textWidthPx(face, lbl)

			if g.Inverted {
				labelX := scaleLeft + scaleW + int(math.Round(0.04*cmPx))
				drawGaugeText(img, face, lbl, labelX, ty-lh/2, colorBlack)
			} else {
				labelX := scaleLeft - lw - int(math.Round(0.04*cmPx))
				drawGaugeText(img, face, lbl, labelX, ty-lh/2, colorBlack)
			}
		}

		// Pointer: horizontal arrow at current value position.
		pct := g.FillPercent()
		py := scaleTop + scaleH - int(math.Round(float64(scaleH)*pct))
		px := w / 2
		pHeight := int(math.Round(float64(w) * 0.4))
		pHalfH := int(math.Round(float64(h) * 0.036 / 2))
		if pHalfH < 2 {
			pHalfH = 2
		}
		// Vertical gauge pointer: draw as horizontal line of height pHalfH*2
		fillRect(img, px-pHeight, py-pHalfH, pHeight, pHalfH*2, pointerColor)
	}

	return img
}

// ── RadialGauge ───────────────────────────────────────────────────────────────

// drawRadialScaleTicks draws major and minor tick marks along the scale arc of
// a radial gauge.
//
// C# source: original-dotnet/FastReport.Base/Gauge/Radial/RadialScale.cs
// DrawMajorTicks / DrawMinorTicks methods.
func drawRadialScaleTicks(img *image.RGBA, cx, cy, radius int, startAngle, endAngle float64, majorCount, minorCount int, tickColor color.RGBA) {
	if majorCount < 2 || radius <= 0 {
		return
	}
	sweep := endAngle - startAngle
	if sweep < 0 {
		sweep += 360
	}

	outerR := float64(radius)
	majorInnerR := outerR * 0.85
	minorInnerR := outerR * 0.90

	// Major ticks.
	for i := 0; i < majorCount; i++ {
		frac := float64(i) / float64(majorCount-1)
		angleDeg := startAngle + sweep*frac
		angleRad := angleDeg * math.Pi / 180
		cos, sin := math.Cos(angleRad), math.Sin(angleRad)
		x0 := cx + int(math.Round(majorInnerR*cos))
		y0 := cy + int(math.Round(majorInnerR*sin))
		x1 := cx + int(math.Round(outerR*cos))
		y1 := cy + int(math.Round(outerR*sin))
		drawLine(img, x0, y0, x1, y1, tickColor)
	}

	// Minor ticks between each pair of major ticks.
	for i := 0; i < majorCount-1; i++ {
		for j := 1; j <= minorCount; j++ {
			frac := (float64(i) + float64(j)/float64(minorCount+1)) / float64(majorCount-1)
			angleDeg := startAngle + sweep*frac
			angleRad := angleDeg * math.Pi / 180
			cos, sin := math.Cos(angleRad), math.Sin(angleRad)
			x0 := cx + int(math.Round(minorInnerR*cos))
			y0 := cy + int(math.Round(minorInnerR*sin))
			x1 := cx + int(math.Round(outerR*cos))
			y1 := cy + int(math.Round(outerR*sin))
			drawLine(img, x0, y0, x1, y1, tickColor)
		}
	}
}

// drawRadialLabels draws numeric value labels at major tick positions around
// the arc. Text is centered at each label position (just inside the tick marks).
//
// C# source: original-dotnet/FastReport.Base/Gauge/Radial/RadialScale.cs
// DrawMajorTicks — DrawText call for each tick.
func drawRadialLabels(img *image.RGBA, face xfont.Face, cx, cy, labelRadius int, startAngle, endAngle float64, majorCount int, minimum, maximum float64, textColor color.RGBA) {
	if majorCount < 2 || labelRadius <= 0 {
		return
	}
	sweep := endAngle - startAngle
	if sweep < 0 {
		sweep += 360
	}
	valueStep := (maximum - minimum) / float64(majorCount-1)
	lh := textHeightPx(face)
	for i := 0; i < majorCount; i++ {
		frac := float64(i) / float64(majorCount-1)
		angleDeg := startAngle + sweep*frac
		angleRad := angleDeg * math.Pi / 180
		px := cx + int(math.Round(float64(labelRadius)*math.Cos(angleRad)))
		py := cy + int(math.Round(float64(labelRadius)*math.Sin(angleRad)))

		val := minimum + valueStep*float64(i)
		lbl := formatGaugeLabel(val)
		lw := textWidthPx(face, lbl)
		// Center text on (px, py)
		drawGaugeText(img, face, lbl, px-lw/2, py-lh/2, textColor)
	}
}

// drawRadialLabelMarkers places 2×2 dot markers at label positions along the arc.
// Retained for internal test compatibility; the live renderer uses drawRadialLabels
// which renders actual text strings.
//
// C# source: original-dotnet/FastReport.Base/Gauge/Radial/RadialScale.cs
// DrawMajorTicks – DrawText call for each tick.
func drawRadialLabelMarkers(img *image.RGBA, cx, cy, labelRadius int, startAngle, endAngle float64, majorCount int, markerColor color.RGBA) {
	if majorCount < 2 || labelRadius <= 0 {
		return
	}
	sweep := endAngle - startAngle
	if sweep < 0 {
		sweep += 360
	}
	b := img.Bounds()
	for i := 0; i < majorCount; i++ {
		frac := float64(i) / float64(majorCount-1)
		angleDeg := startAngle + sweep*frac
		angleRad := angleDeg * math.Pi / 180
		px := cx + int(math.Round(float64(labelRadius)*math.Cos(angleRad)))
		py := cy + int(math.Round(float64(labelRadius)*math.Sin(angleRad)))
		for dy := 0; dy < 2; dy++ {
			for dx := 0; dx < 2; dx++ {
				lx, ly := px+dx, py+dy
				if lx >= b.Min.X && lx < b.Max.X && ly >= b.Min.Y && ly < b.Max.Y {
					img.SetRGBA(lx, ly, markerColor)
				}
			}
		}
	}
}

// radialCircleScaleParams computes shared geometry parameters for a full-circle
// radial gauge matching C# RadialScale.DrawMajorTicks/DrawMinorTicks.
// Returns: majorTicksOffset (px gap from arc edge to outer tick end),
//          majorLen (px length of major ticks), minorLen (px length of minor ticks).
//
// C# source: original-dotnet/FastReport.Base/Gauge/Radial/RadialScale.cs lines 209-215, 517.
func radialCircleScaleParams(face xfont.Face, gaugeW int, minimum, maximum float64) (majorTicksOffset, majorLen, minorLen int) {
	lh := textHeightPx(face)
	txtSz := func(v float64) int {
		lbl := formatGaugeLabel(v)
		tw := textWidthPx(face, lbl)
		if lh > tw {
			return lh
		}
		return tw
	}
	avgVal := minimum + (maximum-minimum)/2
	mto := txtSz(maximum)
	if t := txtSz(minimum); t > mto {
		mto = t
	}
	if t := txtSz(avgVal); t > mto {
		mto = t
	}
	majorTicksOffset = mto + 2 // 2px gap between label and arc
	majorLen = gaugeW / 12
	if majorLen < 3 {
		majorLen = 3
	}
	minorLen = gaugeW / 24
	if minorLen < 2 {
		minorLen = 2
	}
	return
}

// drawRadialCircleScale draws tick marks and value labels for a full-circle radial gauge
// exactly matching C# RadialScale geometry for Circle type.
//
// Key geometry (C# RadialScale.cs DrawMajorTicks/DrawMinorTicks):
//   - Center tick at screen angle -90° (UP), representing the midpoint value
//   - 5 ticks clockwise (+27° each) and 5 counter-clockwise (-27° each)
//   - majorTicksOffset = max(textWidth, textHeight) for min/max/avg labels
//   - MajorTicks.Length = gaugeW/12 (overrides FRX value at draw time)
//   - MinorTicks.Length = gaugeW/24
//   - minorTicksOffset = majorTicksOffset + majorLen/2 - minorLen/2 (minor centered with major)
//   - Tick outer end at radius (r - majorTicksOffset), inner end minus majorLen
//   - Labels at outer tick end, aligned radially outward from center
//
// C# source: original-dotnet/FastReport.Base/Gauge/Radial/RadialScale.cs
// DrawMajorTicks (lines 191-509) and DrawMinorTicks (lines 512-579).
func drawRadialCircleScale(img *image.RGBA, face xfont.Face, cx, cy, r, gaugeW int,
	minimum, maximum float64, tickColor, textColor color.RGBA, showLabels bool) {
	lh := textHeightPx(face)

	majorTicksOffset, majorLen, minorLen := radialCircleScaleParams(face, gaugeW, minimum, maximum)

	// minorTicksOffset centers minor ticks at the same radius as major tick centers.
	// C# DrawMinorTicks line 518: minorTicksOffset = majorTicksOffset + majorLen/2 - minorLen/2.
	minorTicksOffset := float64(majorTicksOffset) + float64(majorLen)/2.0 - float64(minorLen)/2.0

	majorOuterR := float64(r - majorTicksOffset)
	majorInnerR := majorOuterR - float64(majorLen)
	minorOuterR := float64(r) - minorTicksOffset
	minorInnerR := minorOuterR - float64(minorLen)

	// Minor ticks: 4 ticks between each adjacent major pair, 5.4° apart.
	// C#: 10 gaps × 4 ticks = 40 total, symmetric about the center tick at -90°.
	for i := -5; i < 5; i++ {
		for j := 1; j <= 4; j++ {
			angleDeg := -90.0 + float64(i)*27.0 + float64(j)*5.4
			rad := angleDeg * math.Pi / 180
			cosA, sinA := math.Cos(rad), math.Sin(rad)
			x0 := cx + int(math.Round(minorOuterR*cosA))
			y0 := cy + int(math.Round(minorOuterR*sinA))
			x1 := cx + int(math.Round(minorInnerR*cosA))
			y1 := cy + int(math.Round(minorInnerR*sinA))
			drawLine(img, x0, y0, x1, y1, tickColor)
		}
	}

	// Major ticks and labels: 11 ticks, i = -5..5.
	// i=-5 → minimum at 135° (lower-left), i=0 → avg at -90° (top), i=5 → maximum at 45° (lower-right).
	for i := -5; i <= 5; i++ {
		angleDeg := -90.0 + float64(i)*27.0
		rad := angleDeg * math.Pi / 180
		cosA, sinA := math.Cos(rad), math.Sin(rad)

		ox := cx + int(math.Round(majorOuterR*cosA))
		oy := cy + int(math.Round(majorOuterR*sinA))
		ix := cx + int(math.Round(majorInnerR*cosA))
		iy := cy + int(math.Round(majorInnerR*sinA))
		drawLine(img, ox, oy, ix, iy, tickColor)

		if !showLabels {
			continue
		}

		val := minimum + float64(i+5)/10.0*(maximum-minimum)
		lbl := formatGaugeLabel(val)
		lw := textWidthPx(face, lbl)

		// Place label at outer tick end, extending radially outward from center.
		// Matches C# DrawText HorAlign/VertAlign logic: text is placed in the quadrant
		// that is outward (away from center) relative to the tick direction.
		const hThresh = 0.26 // ~15°: threshold for horizontal vs vertical tick direction
		const vThresh = 0.26
		var labelX, labelY int
		if cosA > hThresh {
			labelX = ox // text extends right (tick points right)
		} else if cosA < -hThresh {
			labelX = ox - lw // text extends left (tick points left)
		} else {
			labelX = ox - lw/2 // centered (tick points up or down)
		}
		if sinA > vThresh {
			labelY = oy // text extends below (tick points down)
		} else if sinA < -vThresh {
			labelY = oy - lh // text extends above (tick points up)
		} else {
			labelY = oy - lh/2 // centered (tick points left or right)
		}
		drawGaugeText(img, face, lbl, labelX, labelY, textColor)
	}
}

// drawRadialPointerNeedle draws a tapered trapezoid needle matching the C#
// RadialPointer shape: a wide quadrilateral with the broad base near the center
// and a narrow tip near the scale arc, then rotated to the needle angle.
//
// C# source: original-dotnet/FastReport.Base/Gauge/Radial/RadialPointer.cs
// DrawHorz method, lines 94–169.
// Geometry (before rotation, pointing right = angle 0°):
//   base at +baseDist from centre, half-width ±baseHalfW
//   tip  at +tipDist  from centre, half-width ±tipHalfW
// All points rotated by needleAngleDeg.
func drawRadialPointerNeedle(img *image.RGBA, cx, cy int, needleAngleDeg float64, radius int, pointerColor color.RGBA) {
	fr := float64(radius)
	baseDist := fr * 0.08
	tipDist := fr * 0.80
	baseHalfW := math.Max(2, fr/12)
	tipHalfW := math.Max(1, baseHalfW/3)

	rad := needleAngleDeg * math.Pi / 180
	cosA, sinA := math.Cos(rad), math.Sin(rad)

	rotPt := func(dx, dy float64) [2]int {
		x := cx + int(math.Round(dx*cosA-dy*sinA))
		y := cy + int(math.Round(dx*sinA+dy*cosA))
		return [2]int{x, y}
	}

	p0 := rotPt(baseDist, -baseHalfW) // base upper
	p1 := rotPt(baseDist, +baseHalfW) // base lower
	p2 := rotPt(tipDist, +tipHalfW)   // tip lower
	p3 := rotPt(tipDist, -tipHalfW)   // tip upper

	fillConvexPoly(img, [][2]int{p0, p1, p2, p3}, pointerColor)
}

// RenderRadial renders a RadialGauge into an RGBA image of size (w,h).
// It draws a circular (or semi/quadrant) arc, scale tick marks, value labels
// and a needle pointer.
//
// C# source: original-dotnet/FastReport.Base/Gauge/Radial/RadialGauge.cs Draw()
func RenderRadial(g *RadialGauge, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgColor := gaugeBgColor(g.Fill())
	fillRect(img, 0, 0, w, h, bgColor)

	cx := w / 2
	cy := h / 2
	rx := w/2 - 4
	ry := h/2 - 4
	if rx <= 0 || ry <= 0 {
		return img
	}

	// Fill the elliptical dial background.
	// C# RadialGauge.Draw: FillAndDrawEllipse with gauge's Fill brush.
	// original-dotnet/FastReport.Base/Gauge/Radial/RadialGauge.cs lines 273-282.
	fillEllipse(img, cx, cy, rx, ry, bgColor)

	// Determine arc angles based on gauge type/position.
	var arcStart, arcEnd float64
	switch g.GaugeType {
	case RadialGaugeTypeSemicircle:
		switch g.Position {
		case RadialGaugePositionTop:
			arcStart, arcEnd = -180, 0
		case RadialGaugePositionBottom:
			arcStart, arcEnd = 0, 180
		case RadialGaugePositionLeft:
			arcStart, arcEnd = 90, 270
		case RadialGaugePositionRight:
			arcStart, arcEnd = -90, 90
		default:
			arcStart, arcEnd = -180, 0
		}
	case RadialGaugeTypeQuadrant:
		switch {
		case g.Position.IsTop() && g.Position.IsLeft():
			arcStart, arcEnd = -180, -90
		case g.Position.IsBottom() && g.Position.IsLeft():
			arcStart, arcEnd = -270, -180
		case g.Position.IsTop() && g.Position.IsRight():
			arcStart, arcEnd = -90, 0
		case g.Position.IsBottom() && g.Position.IsRight():
			arcStart, arcEnd = 0, 90
		default:
			arcStart, arcEnd = -180, -90
		}
	default: // Circle: full perimeter arc
		arcStart, arcEnd = 0, 360
	}

	// Draw the scale arc (outer rim border).
	// C# RadialGauge.Draw: DrawArc/DrawEllipse with Border pen.
	borderColor := colorDarkGray
	drawArc(img, cx, cy, rx, ry, arcStart, arcEnd, borderColor)
	if rx > 4 && ry > 4 {
		drawArc(img, cx, cy, rx-1, ry-1, arcStart, arcEnd, borderColor)
	}

	// Draw scale ticks and labels.
	tickColor := colorDarkGray
	if c := parseColor(g.Scale.MajorTicks.Color, color.RGBA{}); c != (color.RGBA{}) {
		tickColor = c
	}
	r := min(rx, ry)
	face := gaugeFont(g.Scale.Font)

	if g.GaugeType == RadialGaugeTypeCircle {
		// C# RadialScale geometry: ticks go inward from arc, labels fit in the gap.
		// original-dotnet/FastReport.Base/Gauge/Radial/RadialScale.cs DrawMajorTicks/DrawMinorTicks.
		drawRadialCircleScale(img, face, cx, cy, r, w, g.Minimum, g.Maximum,
			tickColor, colorBlack, g.Scale.ShowLabels)
	} else {
		// Semicircle / Quadrant: use generic evenly-spaced tick distribution.
		majorCount, minorCount := 5, 3
		tickStart, tickEnd := arcStart, arcEnd
		drawRadialScaleTicks(img, cx, cy, r, tickStart, tickEnd, majorCount, minorCount, tickColor)
		if g.Scale.ShowLabels {
			lh := textHeightPx(face)
			labelR := r - 6 - lh/2
			if labelR > 0 {
				drawRadialLabels(img, face, cx, cy, labelR, tickStart, tickEnd, majorCount, g.Minimum, g.Maximum, colorBlack)
			}
		}
	}

	// Draw the needle pointer.
	// For Circle type, NeedleAngle() computes the angle from StartAngle + sweep*Percent().
	effectiveStart := g.EffectiveStartAngle()
	var needleAngle float64
	if g.GaugeType == RadialGaugeTypeCircle {
		needleAngle = g.NeedleAngle()
	} else {
		var needleSweep float64
		if g.GaugeType.IsQuadrant() {
			needleSweep = 90
		} else {
			needleSweep = 180
		}
		dir := 1.0
		if g.Position == RadialGaugePositionBottom ||
			(g.GaugeType.IsQuadrant() && g.Position.IsBottom() && g.Position.IsRight()) {
			dir = -1
		}
		needleAngle = effectiveStart + dir*needleSweep*g.Percent()
	}

	pointerColor := parseColor(g.Pointer.Color, colorOrange)

	if g.GaugeType == RadialGaugeTypeCircle {
		// C# RadialPointer.DrawHorz exact needle geometry:
		// circleWidth = gaugeWidth/16, hub radius = circleWidth/2 * scaleX = w/32
		// ptrLineY  = center.Y - (circleWidth*scaleX/2 + circleWidth*scaleX/5) = cy - 7w/160
		//   → baseDist (from center to needle base) = 7w/160
		// ptrLineY1 = avrTick.Y + minorLen * 1.7 (from top)
		//   → tipDist (from center to tip) = h/2 - majorTicksOffset - minorLen * 1.7
		// ptrLineWidth = (circleWidth/3) * scaleX = w/48
		//   → baseHalfW = w/48, tipHalfW = w/144
		// original-dotnet/FastReport.Base/Gauge/Radial/RadialPointer.cs lines 55-169.
		mto, _, minLen := radialCircleScaleParams(face, w, g.Minimum, g.Maximum)

		baseDist := 7.0 * float64(w) / 160.0
		tipDist := float64(h)/2.0 - float64(mto) - float64(minLen)*1.7
		if tipDist <= baseDist+2 {
			tipDist = baseDist + 2
		}

		baseHalfW := math.Max(1, float64(w)/48.0)
		tipHalfW := math.Max(0.5, baseHalfW/3.0)

		nrad := needleAngle * math.Pi / 180
		ncos, nsin := math.Cos(nrad), math.Sin(nrad)
		rotPt := func(dx, dy float64) [2]int {
			return [2]int{
				cx + int(math.Round(dx*ncos-dy*nsin)),
				cy + int(math.Round(dx*nsin+dy*ncos)),
			}
		}
		fillConvexPoly(img, [][2]int{
			rotPt(baseDist, -baseHalfW),
			rotPt(baseDist, +baseHalfW),
			rotPt(tipDist, +tipHalfW),
			rotPt(tipDist, -tipHalfW),
		}, pointerColor)

		// Hub: C# pointerCircle = Width/16 × Height/16, radius = w/32.
		hubR := w / 32
		if hubR < 2 {
			hubR = 2
		}
		fillEllipse(img, cx, cy, hubR, hubR, pointerColor)

		// Label text at bottom center of dial.
		// C# RadialLabel.Draw: lblPt = (cx, h-1-avrTick.Y), text centered.
		// original-dotnet/FastReport.Base/Gauge/Radial/RadialLabel.cs lines 27-31.
		if g.Label.Text != "" {
			lface := gaugeFont(g.Label.Font)
			lw := textWidthPx(lface, g.Label.Text)
			lh2 := textHeightPx(lface)
			labelY := h - 1 - mto - lh2/2
			drawGaugeText(img, lface, g.Label.Text, cx-lw/2, labelY, colorBlack)
		}
	} else {
		drawRadialPointerNeedle(img, cx, cy, needleAngle, r, pointerColor)

		// Center hub circle.
		// C# RadialPointer.DrawHorz: FillAndDrawEllipse(pen, brush, pointerCircle)
		// where pointerCircle = Width/16 × Height/16, filled with pointer brush (orange).
		// original-dotnet/FastReport.Base/Gauge/Radial/RadialPointer.cs line 167.
		hubR := int(math.Round(float64(r) / 8))
		if hubR < 2 {
			hubR = 2
		}
		fillEllipse(img, cx, cy, hubR, hubR, pointerColor)
	}

	return img
}

// ── SimpleGauge ───────────────────────────────────────────────────────────────

// RenderSimple renders a SimpleGauge into an RGBA image of size (w,h).
// It draws two subscales (above/below center) with major/minor tick marks and
// a thin rectangular pointer bar at the current value position.
//
// C# source: original-dotnet/FastReport.Base/Gauge/Simple/SimpleGauge.cs Draw()
// original-dotnet/FastReport.Base/Gauge/Simple/SimpleScale.cs
// original-dotnet/FastReport.Base/Gauge/Simple/SimplePointer.cs
func RenderSimple(g *SimpleGauge, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fillRect(img, 0, 0, w, h, gaugeBgColor(g.Fill()))

	majorTickColor := parseColor(g.Scale.MajorTicks.Color, colorDarkGray)
	minorTickColor := parseColor(g.Scale.MinorTicks.Color, colorDarkGray)
	pointerColor := parseColor(g.Pointer.Color, colorOrange)
	face := gaugeFont(g.Scale.Font)

	// C# SimpleScale: majorTicksNum = 6 (same as LinearScale)
	// original-dotnet/FastReport.Base/Gauge/Simple/SimpleScale.cs line 65.
	const majorCount = 6

	if !g.Vertical() {
		// Horizontal: same geometry as LinearScale.
		// left = (AbsLeft + 0.5cm) * scaleX
		// top = (AbsTop + 0.6cm) * scaleY
		// width = (Width - 1.0cm) * scaleX
		// height = (Height - 1.2cm) * scaleY
		// original-dotnet/FastReport.Base/Gauge/Simple/SimpleScale.cs lines 281-289.
		scaleLeft := int(math.Round(0.5 * cmPx))
		scaleTop := int(math.Round(0.6 * cmPx))
		scaleW := w - int(math.Round(1.0*cmPx))
		scaleH := h - int(math.Round(1.2*cmPx))
		if scaleW <= 0 || scaleH <= 0 {
			return img
		}

		step := float64(scaleW) / float64(majorCount-1)
		valueStep := (g.Maximum - g.Minimum) / float64(majorCount-1)

		// Pointer height: ptrRatio = 0.08 (C# SimplePointer default).
		// original-dotnet/FastReport.Base/Gauge/Simple/SimplePointer.cs line 23.
		const ptrRatio = 0.08
		pointerH := int(math.Round(float64(h) * ptrRatio))
		if pointerH < 1 {
			pointerH = 1
		}
		ptrHalfH := pointerH / 2

		scaleCenter := scaleTop + scaleH/2

		// y1=top, y2=center-ptrHalf, y3=center+ptrHalf, y4=bottom
		// C# SimpleScale.DrawMajorTicksHorz lines 80-84.
		y1 := scaleTop
		y2 := scaleCenter - ptrHalfH - 1
		y3 := scaleCenter + ptrHalfH + 1
		y4 := scaleTop + scaleH

		// Minor ticks (3 between each major pair, 15% inset).
		minorStep := step / 4.0
		y1minor := y1 + int(math.Round(float64(scaleH)*0.15))
		y2minor := y2
		y3minor := y3
		y4minor := y4 - int(math.Round(float64(scaleH)*0.15))

		for i := 0; i < majorCount-1; i++ {
			for j := 1; j <= 3; j++ {
				mx := scaleLeft + int(math.Round(float64(i)*step+float64(j)*minorStep))
				if g.FirstSubScale.Enabled {
					drawVLine(img, mx, y1minor, y2minor, minorTickColor)
				}
				if g.SecondSubScale.Enabled {
					drawVLine(img, mx, y3minor, y4minor, minorTickColor)
				}
			}
		}

		// Major ticks and labels.
		for i := 0; i < majorCount; i++ {
			tx := scaleLeft + int(math.Round(float64(i)*step))
			val := g.Minimum + valueStep*float64(i)
			lbl := formatGaugeLabel(val)
			lw := textWidthPx(face, lbl)

			if g.FirstSubScale.Enabled {
				drawVLine(img, tx, y1, y2, majorTickColor)
				if g.FirstSubScale.ShowCaption {
					// Labels above the first subscale: top-of-text = y1 - 0.4cm
					// C# DrawString takes top-left; drawGaugeText also takes top-left.
					// original-dotnet/FastReport.Base/Gauge/Simple/SimpleScale.cs line 97.
					labelY := y1 - int(math.Round(0.4*cmPx))
					drawGaugeText(img, face, lbl, tx-lw/2, labelY, colorBlack)
				}
			}
			if g.SecondSubScale.Enabled {
				drawVLine(img, tx, y3, y4, majorTickColor)
				if g.SecondSubScale.ShowCaption {
					// Labels below the second subscale: y = y4 + 0.08cm
					// C# SimpleScale.DrawMajorTicksHorz line 114.
					labelY := y4 + int(math.Round(0.08*cmPx))
					drawGaugeText(img, face, lbl, tx-lw/2, labelY, colorBlack)
				}
			}
		}

		// Pointer: thin horizontal bar from left edge to current value.
		// C# SimplePointer.DrawHorz:
		// left = AbsLeft + border/2 + horizontalOffset (=0.5cm)
		// width = (Width - border - hOff*2) * percent = scaleW * percent
		// height = (Height - border) * ptrRatio = h * 0.08
		// top = AbsTop + (Height-border)/2 - height/2  (vertically centered)
		// original-dotnet/FastReport.Base/Gauge/Simple/SimplePointer.cs lines 119-125.
		pct := g.Percent()
		ptrW := int(math.Round(float64(scaleW) * pct))
		ptrTop := scaleCenter - ptrHalfH
		if ptrW > 0 {
			fillRect(img, scaleLeft, ptrTop, ptrW, pointerH, pointerColor)
		}
	} else {
		// Vertical: similar layout rotated 90°.
		// left = (AbsLeft + 0.7cm) * scaleX
		// top = (AbsTop + 0.5cm) * scaleY
		// height = (Height - 1.0cm) * scaleY
		// width = (Width - 1.4cm) * scaleX
		scaleLeft := int(math.Round(0.7 * cmPx))
		scaleTop := int(math.Round(0.5 * cmPx))
		scaleW := w - int(math.Round(1.4*cmPx))
		scaleH := h - int(math.Round(1.0*cmPx))
		if scaleW <= 0 || scaleH <= 0 {
			return img
		}

		step := float64(scaleH) / float64(majorCount-1)
		valueStep := (g.Maximum - g.Minimum) / float64(majorCount-1)

		const ptrRatio = 0.08
		pointerW := int(math.Round(float64(w) * ptrRatio))
		if pointerW < 1 {
			pointerW = 1
		}
		ptrHalfW := pointerW / 2
		scaleCenter := scaleLeft + scaleW/2

		x1 := scaleLeft
		x2 := scaleCenter - ptrHalfW - 1
		x3 := scaleCenter + ptrHalfW + 1
		x4 := scaleLeft + scaleW

		lh := textHeightPx(face)
		minorStep := step / 4.0
		x1minor := x1 + int(math.Round(float64(scaleW)*0.15))
		x4minor := x4 - int(math.Round(float64(scaleW)*0.15))

		for i := 0; i < majorCount-1; i++ {
			for j := 1; j <= 3; j++ {
				my := scaleTop + scaleH - int(math.Round(float64(i)*step+float64(j)*minorStep))
				if g.FirstSubScale.Enabled {
					drawHLine(img, x1minor, my, x2, minorTickColor)
				}
				if g.SecondSubScale.Enabled {
					drawHLine(img, x3, my, x4minor, minorTickColor)
				}
			}
		}

		for i := 0; i < majorCount; i++ {
			ty := scaleTop + scaleH - int(math.Round(float64(i)*step))
			val := g.Minimum + valueStep*float64(i)
			lbl := formatGaugeLabel(val)
			lw := textWidthPx(face, lbl)

			if g.FirstSubScale.Enabled {
				drawHLine(img, x1, ty, x2, majorTickColor)
				if g.FirstSubScale.ShowCaption {
					labelX := x1 - lw - int(math.Round(0.04*cmPx))
					drawGaugeText(img, face, lbl, labelX, ty-lh/2, colorBlack)
				}
			}
			if g.SecondSubScale.Enabled {
				drawHLine(img, x3, ty, x4, majorTickColor)
				if g.SecondSubScale.ShowCaption {
					labelX := x4 + int(math.Round(0.04*cmPx))
					drawGaugeText(img, face, lbl, labelX, ty-lh/2, colorBlack)
				}
			}
		}

		pct := g.Percent()
		ptrH := int(math.Round(float64(scaleH) * pct))
		ptrLeft := scaleCenter - ptrHalfW
		if ptrH > 0 {
			fillRect(img, ptrLeft, scaleTop+scaleH-ptrH, pointerW, ptrH, pointerColor)
		}
	}

	return img
}

// ── SimpleProgressGauge ───────────────────────────────────────────────────────

// RenderSimpleProgress renders a SimpleProgressGauge as a horizontal or vertical
// progress bar with a centered percentage label.
//
// C# source: original-dotnet/FastReport.Base/Gauge/Simple/Progress/SimpleProgressPointer.cs
// DrawHorz/DrawVert, and SimpleProgressLabel.Draw()
func RenderSimpleProgress(g *SimpleProgressGauge, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// White background to show the fill color clearly.
	fillRect(img, 0, 0, w, h, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	pct := g.Percent()
	pointerColor := parseColor(g.Pointer.Color, colorOrange)

	// C# SimpleProgressPointer.DrawHorz with PointerRatio=1, HorizontalOffset=0:
	// Left = border/2 ≈ 0 (border default is thin)
	// Height = (Parent.Height - border) * PointerRatio = full height
	// Width = (Parent.Width - border) * percent
	// original-dotnet/FastReport.Base/Gauge/Simple/Progress/SimpleProgressPointer.cs lines 66-89.
	if g.Vertical() {
		// Vertical fill from bottom upward.
		fillH := int(math.Round(float64(h) * pct))
		if fillH > 0 {
			fillRect(img, 0, h-fillH, w, fillH, pointerColor)
		}
	} else {
		// Horizontal fill from left.
		fillW := int(math.Round(float64(w) * pct))
		if fillW > 0 {
			fillRect(img, 0, 0, fillW, h, pointerColor)
		}
	}

	// Percentage label centered on the gauge.
	// C# SimpleProgressLabel.Draw: text = percent%, drawn centered.
	// original-dotnet/FastReport.Base/Gauge/Simple/Progress/SimpleProgressLabel.cs lines 52-65.
	if g.ShowText {
		face := gaugeFont(g.Label.Font)
		pctVal := (g.Value() - g.Minimum) / (g.Maximum - g.Minimum) * 100
		text := fmt.Sprintf("%d%%", int(math.Round(pctVal)))
		lw := textWidthPx(face, text)
		lh := textHeightPx(face)
		tx := (w - lw) / 2
		ty := (h - lh) / 2
		drawGaugeText(img, face, text, tx, ty, colorBlack)
	}

	return img
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
