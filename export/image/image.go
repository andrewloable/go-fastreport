// Package image provides PNG image export for go-fastreport.
package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// DefaultDPI is the output resolution used when converting pixel coordinates
// to image pixels. 96 dpi matches the internal unit system.
const DefaultDPI = 96

// Exporter renders each PreparedPage as a PNG image.
// When a report contains multiple pages, each page is written sequentially
// to the output writer as a separate PNG. For single-page reports this
// behaves identically to a standard PNG encoder.
//
// Usage:
//
//	exp := image.NewExporter()
//	err := exp.Export(preparedPages, outputWriter)
type Exporter struct {
	export.ExportBase

	// Scale is a multiplier applied to all coordinates. Default is 1.0.
	Scale float64

	// BackgroundColor is the page background. Default is white.
	BackgroundColor color.RGBA

	// BandBorderColor is the color used to draw band outlines. Default is light gray.
	BandBorderColor color.RGBA

	// BandFillColor is the fill color for band rectangles. Default is very light blue.
	BandFillColor color.RGBA

	w       io.Writer
	pp      *preview.PreparedPages
	curPage *image.RGBA
}

// NewExporter creates an Exporter with default settings.
func NewExporter() *Exporter {
	return &Exporter{
		ExportBase:      export.NewExportBase(),
		Scale:           1.0,
		BackgroundColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		BandBorderColor: color.RGBA{R: 180, G: 180, B: 180, A: 255},
		BandFillColor:   color.RGBA{R: 230, G: 240, B: 255, A: 255},
	}
}

// Export writes each page of pp as a PNG image to w.
func (e *Exporter) Export(pp *preview.PreparedPages, w io.Writer) error {
	e.w = w
	e.pp = pp
	return e.ExportBase.Export(pp, w, e)
}

// ── Exporter interface ────────────────────────────────────────────────────────

// Start is a no-op for the image exporter.
func (e *Exporter) Start() error { return nil }

// ExportPageBegin creates a new blank RGBA canvas for the page.
func (e *Exporter) ExportPageBegin(pg *preview.PreparedPage) error {
	w := e.scaled(int(pg.Width))
	h := e.scaled(int(pg.Height))
	if w <= 0 {
		w = 794 // A4 width at 96 dpi
	}
	if h <= 0 {
		h = 1123 // A4 height at 96 dpi
	}
	e.curPage = image.NewRGBA(image.Rect(0, 0, w, h))
	// Fill with background color.
	draw.Draw(e.curPage, e.curPage.Bounds(), &image.Uniform{e.BackgroundColor}, image.Point{}, draw.Src)
	return nil
}

// ExportBand draws the band background and renders child objects on the canvas.
func (e *Exporter) ExportBand(b *preview.PreparedBand) error {
	if e.curPage == nil {
		return nil
	}

	pageW := e.curPage.Bounds().Dx()
	x0 := 0
	y0 := e.scaled(int(b.Top))
	x1 := pageW
	y1 := y0 + e.scaled(int(b.Height))

	if x1 <= x0 {
		x1 = x0 + 1
	}
	if y1 <= y0 {
		y1 = y0 + 1
	}

	r := image.Rect(x0, y0, x1, y1)

	// Fill band background.
	draw.Draw(e.curPage, r, &image.Uniform{e.BandFillColor}, image.Point{}, draw.Src)
	// Draw band border.
	e.drawRect(r, e.BandBorderColor)

	// Render child objects.
	for _, obj := range b.Objects {
		e.renderObject(obj, b.Top)
	}

	return nil
}

// renderObject draws a single PreparedObject on the current page canvas.
func (e *Exporter) renderObject(obj preview.PreparedObject, bandTop float32) {
	if e.curPage == nil {
		return
	}

	x := e.scaled(int(obj.Left))
	y := e.scaled(int(bandTop + obj.Top))
	w := e.scaled(int(obj.Width))
	h := e.scaled(int(obj.Height))
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	bounds := image.Rect(x, y, x+w, y+h).Intersect(e.curPage.Bounds())

	switch obj.Kind {
	case preview.ObjectTypeText:
		if bounds.Empty() {
			return
		}
		// Fill object background.
		if obj.FillColor.A > 0 {
			draw.Draw(e.curPage, bounds, &image.Uniform{obj.FillColor}, image.Point{}, draw.Over)
		} else {
			white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
			draw.Draw(e.curPage, bounds, &image.Uniform{white}, image.Point{}, draw.Over)
		}

		e.drawBorderLines(obj.Border, x, y, w, h)

		if obj.Text == "" {
			return
		}

		tc := obj.TextColor
		if tc.A == 0 {
			tc = color.RGBA{A: 255}
		}

		face := basicfont.Face7x13
		lineH := face.Metrics().Height.Ceil()
		ascent := face.Metrics().Ascent.Ceil()
		lines := e.wrapText(obj.Text, w, face)

		textBlockH := len(lines) * lineH
		startY := y + ascent
		switch obj.VertAlign {
		case 1:
			startY = y + (h-textBlockH)/2 + ascent
		case 2:
			startY = y + h - textBlockH + ascent
		}
		d := &font.Drawer{
			Dst:  e.curPage,
			Src:  image.NewUniform(tc),
			Face: face,
		}
		for _, line := range lines {
			if startY > y+h {
				break
			}
			lineW := font.MeasureString(face, line).Ceil()
			dotX := x + 2
			switch obj.HorzAlign {
			case 1:
				dotX = x + (w-lineW)/2
			case 2:
				dotX = x + w - lineW - 2
			}
			if dotX < x {
				dotX = x
			}
			d.Dot = fixed.P(dotX, startY)
			d.DrawString(line)
			startY += lineH
		}

	case preview.ObjectTypeLine:
		lineColor := color.RGBA{A: 255}
		if obj.Border.Lines[0] != nil && obj.Border.Lines[0].Color.A > 0 {
			lineColor = obj.Border.Lines[0].Color
		}
		if obj.LineDiagonal {
			e.drawLine(x, y, x+w, y+h, lineColor)
		} else {
			// Non-diagonal: horizontal or vertical based on dominant dimension.
			if w >= h {
				e.drawHLine(x, y+h/2, x+w, lineColor)
			} else {
				e.drawVLine(x+w/2, y, y+h, lineColor)
			}
		}

	case preview.ObjectTypeShape:
		if !bounds.Empty() && obj.FillColor.A > 0 {
			draw.Draw(e.curPage, bounds, &image.Uniform{obj.FillColor}, image.Point{}, draw.Over)
		}
		shapeColor := color.RGBA{A: 255}
		if obj.Border.Lines[0] != nil && obj.Border.Lines[0].Color.A > 0 {
			shapeColor = obj.Border.Lines[0].Color
		}
		switch obj.ShapeKind {
		case 2: // Ellipse
			e.drawEllipse(x, y, w, h, shapeColor)
		case 3: // Triangle — draw as simple polygon outline
			e.drawLine(x+w/2, y, x+w, y+h, shapeColor)
			e.drawLine(x+w, y+h, x, y+h, shapeColor)
			e.drawLine(x, y+h, x+w/2, y, shapeColor)
		case 4: // Diamond
			e.drawLine(x+w/2, y, x+w, y+h/2, shapeColor)
			e.drawLine(x+w, y+h/2, x+w/2, y+h, shapeColor)
			e.drawLine(x+w/2, y+h, x, y+h/2, shapeColor)
			e.drawLine(x, y+h/2, x+w/2, y, shapeColor)
		default: // Rectangle (0) and RoundRectangle (1) — simplified to plain rect
			e.drawRect(bounds, shapeColor)
		}
		e.drawBorderLines(obj.Border, x, y, w, h)

	case preview.ObjectTypeCheckBox:
		if !bounds.Empty() {
			white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
			draw.Draw(e.curPage, bounds, &image.Uniform{white}, image.Point{}, draw.Over)
			boxColor := color.RGBA{A: 255}
			e.drawRect(bounds, boxColor)
			// Draw check-mark (X) when checked.
			if obj.Text == "true" {
				pad := 2
				e.drawLine(x+pad, y+pad, x+w-pad, y+h-pad, boxColor)
				e.drawLine(x+w-pad, y+pad, x+pad, y+h-pad, boxColor)
			}
		}

	case preview.ObjectTypePicture:
		// Picture blobs are embedded by the engine; image export skips them here
		// as decoding is handled by the PDF/HTML exporters.

	case preview.ObjectTypePolyLine:
		if len(obj.Points) < 2 {
			return
		}
		lineColor := color.RGBA{A: 255}
		if obj.Border.Lines[0] != nil && obj.Border.Lines[0].Color.A > 0 {
			lineColor = obj.Border.Lines[0].Color
		}
		for i := 1; i < len(obj.Points); i++ {
			x0p := x + e.scaled(int(obj.Points[i-1][0]))
			y0p := y + e.scaled(int(obj.Points[i-1][1]))
			x1p := x + e.scaled(int(obj.Points[i][0]))
			y1p := y + e.scaled(int(obj.Points[i][1]))
			e.drawLine(x0p, y0p, x1p, y1p, lineColor)
		}

	case preview.ObjectTypePolygon:
		if len(obj.Points) < 2 {
			return
		}
		lineColor := color.RGBA{A: 255}
		if obj.Border.Lines[0] != nil && obj.Border.Lines[0].Color.A > 0 {
			lineColor = obj.Border.Lines[0].Color
		}
		n := len(obj.Points)
		for i := 0; i < n; i++ {
			next := (i + 1) % n
			x0p := x + e.scaled(int(obj.Points[i][0]))
			y0p := y + e.scaled(int(obj.Points[i][1]))
			x1p := x + e.scaled(int(obj.Points[next][0]))
			y1p := y + e.scaled(int(obj.Points[next][1]))
			e.drawLine(x0p, y0p, x1p, y1p, lineColor)
		}
	}
}

// drawBorderLines draws the visible border sides of a style.Border onto the canvas.
func (e *Exporter) drawBorderLines(b style.Border, x, y, w, h int) {
	black := color.RGBA{A: 255}
	// Top side (bit 4 = BorderLinesTop).
	if b.VisibleLines&style.BorderLinesTop != 0 && b.Lines[style.BorderTop] != nil {
		c := b.Lines[style.BorderTop].Color
		if c.A == 0 {
			c = black
		}
		e.drawHLine(x, y, x+w, c)
	}
	// Bottom side (bit 8 = BorderLinesBottom).
	if b.VisibleLines&style.BorderLinesBottom != 0 && b.Lines[style.BorderBottom] != nil {
		c := b.Lines[style.BorderBottom].Color
		if c.A == 0 {
			c = black
		}
		e.drawHLine(x, y+h-1, x+w, c)
	}
	// Left side (bit 1 = BorderLinesLeft).
	if b.VisibleLines&style.BorderLinesLeft != 0 && b.Lines[style.BorderLeft] != nil {
		c := b.Lines[style.BorderLeft].Color
		if c.A == 0 {
			c = black
		}
		e.drawVLine(x, y, y+h, c)
	}
	// Right side (bit 2 = BorderLinesRight).
	if b.VisibleLines&style.BorderLinesRight != 0 && b.Lines[style.BorderRight] != nil {
		c := b.Lines[style.BorderRight].Color
		if c.A == 0 {
			c = black
		}
		e.drawVLine(x+w-1, y, y+h, c)
	}
}

// drawLine draws a line from (x0,y0) to (x1,y1) using Bresenham's algorithm.
func (e *Exporter) drawLine(x0, y0, x1, y1 int, c color.RGBA) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	bounds := e.curPage.Bounds()
	for {
		if x0 >= bounds.Min.X && x0 < bounds.Max.X && y0 >= bounds.Min.Y && y0 < bounds.Max.Y {
			e.curPage.SetRGBA(x0, y0, c)
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

// drawEllipse draws the outline of an ellipse inscribed in the given rect.
func (e *Exporter) drawEllipse(x, y, w, h int, c color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}
	a := float64(w) / 2.0
	b := float64(h) / 2.0
	cx := float64(x) + a
	cy := float64(y) + b
	bounds := e.curPage.Bounds()
	steps := int(math.Pi * (a + b))
	if steps < 8 {
		steps = 8
	}
	for i := 0; i < steps; i++ {
		t := 2 * math.Pi * float64(i) / float64(steps)
		px := int(math.Round(cx + a*math.Cos(t)))
		py := int(math.Round(cy + b*math.Sin(t)))
		if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
			e.curPage.SetRGBA(px, py, c)
		}
	}
}

// wrapText splits text into lines that fit within maxWidth pixels using face.
func (e *Exporter) wrapText(text string, maxWidth int, face font.Face) []string {
	// First split by explicit newlines.
	rawLines := strings.Split(text, "\n")
	var result []string
	for _, raw := range rawLines {
		if raw == "" {
			result = append(result, "")
			continue
		}
		words := strings.Fields(raw)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		current := words[0]
		for _, w := range words[1:] {
			candidate := current + " " + w
			if font.MeasureString(face, candidate).Ceil() <= maxWidth {
				current = candidate
			} else {
				result = append(result, current)
				current = w
			}
		}
		result = append(result, current)
	}
	return result
}

// ExportPageEnd encodes the current canvas as PNG and writes it to the output.
func (e *Exporter) ExportPageEnd(_ *preview.PreparedPage) error {
	if e.curPage == nil {
		return nil
	}
	if err := png.Encode(e.w, e.curPage); err != nil {
		return fmt.Errorf("image export: PNG encode: %w", err)
	}
	e.curPage = nil
	return nil
}

// Finish is a no-op for the image exporter.
func (e *Exporter) Finish() error { return nil }

// ── helpers ───────────────────────────────────────────────────────────────────

// scaled converts a pixel value to output pixels applying the Scale factor.
func (e *Exporter) scaled(px int) int {
	if e.Scale == 1.0 || e.Scale <= 0 {
		return px
	}
	return int(math.Round(float64(px) * e.Scale))
}

// drawRect draws a 1-pixel border around rect r in the given color.
func (e *Exporter) drawRect(r image.Rectangle, c color.RGBA) {
	// Top line.
	e.drawHLine(r.Min.X, r.Min.Y, r.Max.X, c)
	// Bottom line.
	e.drawHLine(r.Min.X, r.Max.Y-1, r.Max.X, c)
	// Left line.
	e.drawVLine(r.Min.X, r.Min.Y, r.Max.Y, c)
	// Right line.
	e.drawVLine(r.Max.X-1, r.Min.Y, r.Max.Y, c)
}

func (e *Exporter) drawHLine(x0, y, x1 int, c color.RGBA) {
	bounds := e.curPage.Bounds()
	if y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}
	for x := x0; x < x1; x++ {
		if x >= bounds.Min.X && x < bounds.Max.X {
			e.curPage.SetRGBA(x, y, c)
		}
	}
}

func (e *Exporter) drawVLine(x, y0, y1 int, c color.RGBA) {
	bounds := e.curPage.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X {
		return
	}
	for y := y0; y < y1; y++ {
		if y >= bounds.Min.Y && y < bounds.Max.Y {
			e.curPage.SetRGBA(x, y, c)
		}
	}
}
