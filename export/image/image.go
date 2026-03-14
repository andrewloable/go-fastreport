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

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
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

// ExportBand draws the band as a labeled rectangle on the current page canvas.
// Bands span the full page width; Top and Height come from the PreparedBand.
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

	// Fill band rectangle.
	draw.Draw(e.curPage, r, &image.Uniform{e.BandFillColor}, image.Point{}, draw.Src)

	// Draw border (top + bottom + left + right lines).
	e.drawRect(r, e.BandBorderColor)

	return nil
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
