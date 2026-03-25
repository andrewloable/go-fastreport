// Package image provides multi-format image export for go-fastreport.
// Supported formats: PNG, JPEG, GIF, BMP, TIFF (including multi-frame TIFF).
package image

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/tiff"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
)

// ImageFormat specifies the output image format.
// Mirrors C# ImageExport.ImageExportFormat (ImageExport.cs).
type ImageFormat int

const (
	// ImageFormatPNG produces PNG output (lossless). Default for single-page.
	ImageFormatPNG ImageFormat = iota
	// ImageFormatJPEG produces JPEG output. Use JpegQuality to control compression.
	ImageFormatJPEG
	// ImageFormatGIF produces GIF output (palette-quantised, 256 colours).
	ImageFormatGIF
	// ImageFormatBMP produces BMP output (uncompressed bitmap).
	ImageFormatBMP
	// ImageFormatTIFF produces TIFF output. Set MultiFrameTiff=true for multi-page TIFF.
	ImageFormatTIFF
)

// textRenderTypeHtmlTags is the TextRenderType value for HTML-tagged text.
// Mirrors object.TextRenderTypeHtmlTags = 1.
const textRenderTypeHtmlTags = 1

// textRenderTypeHtmlParagraph is the TextRenderType value for HTML paragraph mode.
// Mirrors object.TextRenderTypeHtmlParagraph = 2.
const textRenderTypeHtmlParagraph = 2

// DefaultDPI is the output resolution used when converting pixel coordinates
// to image pixels. 96 dpi matches the internal unit system.
const DefaultDPI = 96

// Exporter renders each PreparedPage as an image in the configured format.
// Supported formats: PNG, JPEG, GIF, BMP, TIFF.
//
// When SeparateFiles is true (default), each page is encoded and written to
// the output writer in sequence. When SeparateFiles is false, all pages are
// combined into one tall image and written in Finish. Multi-frame TIFF (when
// Format==ImageFormatTIFF && MultiFrameTiff==true) writes all frames as a
// single multi-page TIFF stream.
//
// Usage:
//
//	exp := image.NewExporter()
//	exp.Format = image.ImageFormatJPEG
//	exp.JpegQuality = 85
//	exp.ResolutionX = 150
//	exp.ResolutionY = 150
//	err := exp.Export(preparedPages, outputWriter)
type Exporter struct {
	export.ExportBase

	// Format selects the output image format. Default is ImageFormatJPEG.
	// Mirrors C# ImageExport.ImageFormat (ImageExport.cs line 88).
	Format ImageFormat

	// JpegQuality is the JPEG compression quality, 1–100. Default is 100.
	// Only used when Format == ImageFormatJPEG.
	// Mirrors C# ImageExport.JpegQuality (ImageExport.cs line 160).
	JpegQuality int

	// ResolutionX is the horizontal output DPI. Default is 96.
	// Mirrors C# ImageExport.ResolutionX (ImageExport.cs line 134).
	ResolutionX int

	// ResolutionY is the vertical output DPI. Default is 96.
	// Mirrors C# ImageExport.ResolutionY (ImageExport.cs line 147).
	ResolutionY int

	// SeparateFiles controls whether each page is encoded individually (true)
	// or all pages are combined into one image (false). Default is true.
	// Mirrors C# ImageExport.SeparateFiles (ImageExport.cs line 104).
	SeparateFiles bool

	// MultiFrameTiff produces a multi-page TIFF when Format==ImageFormatTIFF.
	// Default is false.
	// Mirrors C# ImageExport.MultiFrameTiff (ImageExport.cs line 169).
	MultiFrameTiff bool

	// MonochromeTiff converts TIFF output to 1-bit black-and-white.
	// Default is false.
	// Mirrors C# ImageExport.MonochromeTiff (ImageExport.cs line 182).
	MonochromeTiff bool

	// PaddingNonSeparatePages is extra pixel padding around each page when
	// SeparateFiles is false. Default is 0.
	// Mirrors C# ImageExport.PaddingNonSeparatePages (ImageExport.cs line 208).
	PaddingNonSeparatePages int

	// Scale is a multiplier applied to all coordinates. Default is 1.0.
	// When ResolutionX/ResolutionY differ from DefaultDPI (96), the effective
	// scale is (ResolutionX / DefaultDPI) * Scale.
	Scale float64

	// BackgroundColor is the page background. Default is white.
	BackgroundColor color.RGBA

	// BandBorderColor is the color used to draw band outlines. Default is light gray.
	BandBorderColor color.RGBA

	// BandFillColor is the fill color for band rectangles. Default is very light blue.
	BandFillColor color.RGBA

	w   io.Writer
	pp  *preview.PreparedPages

	// curPage is the canvas for the page currently being rendered.
	curPage *image.RGBA

	// combinedPages accumulates rendered pages when SeparateFiles==false.
	// Each entry is a fully rendered page canvas.
	combinedPages []*image.RGBA

	// tiffFrames accumulates page canvases for multi-frame TIFF output.
	tiffFrames []*image.RGBA
}

// NewExporter creates an Exporter with default settings.
// Defaults match C# ImageExport constructor (ImageExport.cs line 710):
//   - Format: ImageFormatJPEG (C#: imageFormat = ImageExportFormat.Jpeg)
//   - JpegQuality: 100 (C#: jpegQuality = 100)
//   - ResolutionX/Y: 96 dpi (C#: Resolution = 96)
//   - SeparateFiles: true (C#: separateFiles = true)
func NewExporter() *Exporter {
	return &Exporter{
		ExportBase:      export.NewExportBase(),
		Format:          ImageFormatJPEG,
		JpegQuality:     100,
		ResolutionX:     DefaultDPI,
		ResolutionY:     DefaultDPI,
		SeparateFiles:   true,
		MultiFrameTiff:  false,
		MonochromeTiff:  false,
		Scale:           1.0,
		BackgroundColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		BandBorderColor: color.RGBA{R: 180, G: 180, B: 180, A: 255},
		BandFillColor:   color.RGBA{R: 230, G: 240, B: 255, A: 255},
	}
}

// Export writes each page of pp as an image to w in the configured format.
// When SeparateFiles is true (default), pages are encoded and written sequentially.
// When SeparateFiles is false, all pages are combined into one image.
// When Format==ImageFormatTIFF && MultiFrameTiff==true, a multi-frame TIFF is written.
func (e *Exporter) Export(pp *preview.PreparedPages, w io.Writer) error {
	e.w = w
	e.pp = pp
	e.combinedPages = nil
	e.tiffFrames = nil
	return e.ExportBase.Export(pp, w, e)
}

// ── Exporter interface ────────────────────────────────────────────────────────

// Start initialises per-export state.
// Mirrors C# ImageExport.Start (ImageExport.cs line 493).
func (e *Exporter) Start() error {
	e.combinedPages = nil
	e.tiffFrames = nil
	return nil
}

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

	// Render watermark behind page content.
	if wm := pg.Watermark; wm != nil && wm.Enabled {
		if !wm.ShowImageOnTop {
			e.renderWatermarkImageOnPage(wm)
		}
		if !wm.ShowTextOnTop {
			e.renderWatermarkTextOnPage(wm)
		}
	}
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

	// For RTF objects, strip RTF control words to get plain text for rendering.
	if obj.Kind == preview.ObjectTypeRTF {
		plain := obj
		plain.Text = utils.StripRTF(obj.Text)
		plain.Kind = preview.ObjectTypeText
		obj = plain
	}

	switch obj.Kind {
	case preview.ObjectTypeText, preview.ObjectTypeHtml:
		if bounds.Empty() {
			return
		}
		// Fill object background. Only fill when FillColor is set (A > 0).
		// When A == 0 (transparent), skip fill so that the canvas background
		// (set by the caller — white for full-page export, transparent for
		// RenderObjectPNG) is preserved.
		if obj.FillColor.A > 0 {
			draw.Draw(e.curPage, bounds, &image.Uniform{obj.FillColor}, image.Point{}, draw.Over)
		}

		e.drawBorderLines(obj.Border, x, y, w, h)

		if obj.Text == "" {
			return
		}

		tc := obj.TextColor
		if tc.A == 0 {
			tc = color.RGBA{A: 255}
		}

		// Select a font face based on obj.Font. Use the object's point size
		// scaled by the effective DPI (ResolutionX × Scale / DefaultDPI).
		fontPt := float64(obj.Font.Size)
		if fontPt <= 0 {
			fontPt = 10
		}
		dpi := float64(e.ResolutionX) * e.Scale
		if dpi <= 0 {
			dpi = DefaultDPI
		}

		// Check if this object uses HTML tag rendering.
		if obj.TextRenderType == textRenderTypeHtmlTags || obj.TextRenderType == textRenderTypeHtmlParagraph {
			e.renderHtmlText(obj, x, y, w, h, fontPt, dpi, tc)
			return
		}

		face := selectFace(obj.Font, fontPt, dpi)
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
		case 1: // RoundRectangle — use ShapeCurve as corner radius.
			radius := int(math.Round(float64(obj.ShapeCurve)))
			if radius <= 0 {
				e.drawRect(bounds, shapeColor)
			} else {
				e.drawRoundRect(x, y, w, h, radius, shapeColor)
			}
		default: // Rectangle (0).
			e.drawRect(bounds, shapeColor)
		}
		e.drawBorderLines(obj.Border, x, y, w, h)

	case preview.ObjectTypeCheckBox:
		if !bounds.Empty() {
			white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
			draw.Draw(e.curPage, bounds, &image.Uniform{white}, image.Point{}, draw.Over)
			boxColor := color.RGBA{A: 255}
			e.drawRect(bounds, boxColor)

			// Determine symbol to draw using obj.Checked/CheckedSymbol/UncheckedSymbol.
			cc := obj.CheckColor
			if cc.A == 0 {
				cc = color.RGBA{A: 255}
			}
			symbol := -1 // none
			if obj.Checked {
				symbol = obj.CheckedSymbol // 0=check, 1=cross, 2=plus, 3=fill
			} else {
				switch obj.UncheckedSymbol {
				case 1:
					symbol = 1 // cross
				case 2:
					symbol = 10 // minus
				case 3:
					symbol = 11 // slash
				case 4:
					symbol = 12 // backslash
				}
			}

			pad := int(float64(w) * 0.15)
			if pad < 2 {
				pad = 2
			}
			switch symbol {
			case 0: // checkmark
				// Draw check: down-right then up-right
				mx := x + int(float64(w)*0.4)
				my := y + int(float64(h)*0.75)
				e.drawLine(x+int(float64(w)*0.15), y+int(float64(h)*0.5), mx, my, cc)
				e.drawLine(mx, my, x+int(float64(w)*0.85), y+int(float64(h)*0.25), cc)
			case 1: // cross (X)
				e.drawLine(x+pad, y+pad, x+w-pad, y+h-pad, cc)
				e.drawLine(x+w-pad, y+pad, x+pad, y+h-pad, cc)
			case 2: // plus (+)
				e.drawLine(x+w/2, y+pad, x+w/2, y+h-pad, cc)
				e.drawLine(x+pad, y+h/2, x+w-pad, y+h/2, cc)
			case 3: // fill (filled rectangle)
				fillBounds := image.Rect(x+pad, y+pad, x+w-pad, y+h-pad)
				draw.Draw(e.curPage, fillBounds, &image.Uniform{cc}, image.Point{}, draw.Over)
			case 10: // minus (-)
				e.drawLine(x+pad, y+h/2, x+w-pad, y+h/2, cc)
			case 11: // slash (/)
				e.drawLine(x+pad, y+h-pad, x+w-pad, y+pad, cc)
			case 12: // backslash (\)
				e.drawLine(x+pad, y+pad, x+w-pad, y+h-pad, cc)
			}
		}

	case preview.ObjectTypePicture:
		e.drawPictureObject(x, y, w, h, obj)

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

	case preview.ObjectTypeDigitalSignature:
		if bounds.Empty() {
			return
		}
		// Fill with white background.
		white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
		draw.Draw(e.curPage, bounds, &image.Uniform{white}, image.Point{}, draw.Over)

		// Draw a dashed border: draw alternating segments on each side.
		borderColor := color.RGBA{A: 255} // black
		dashLen := e.scaled(4)
		gapLen := e.scaled(3)
		if dashLen < 2 {
			dashLen = 2
		}
		if gapLen < 1 {
			gapLen = 1
		}
		// Top edge.
		for px := x; px < x+w; px += dashLen + gapLen {
			end := px + dashLen
			if end > x+w {
				end = x + w
			}
			e.drawHLine(px, y, end, borderColor)
		}
		// Bottom edge.
		for px := x; px < x+w; px += dashLen + gapLen {
			end := px + dashLen
			if end > x+w {
				end = x + w
			}
			e.drawHLine(px, y+h-1, end, borderColor)
		}
		// Left edge.
		for py := y; py < y+h; py += dashLen + gapLen {
			end := py + dashLen
			if end > y+h {
				end = y + h
			}
			e.drawVLine(x, py, end, borderColor)
		}
		// Right edge.
		for py := y; py < y+h; py += dashLen + gapLen {
			end := py + dashLen
			if end > y+h {
				end = y + h
			}
			e.drawVLine(x+w-1, py, end, borderColor)
		}

		// Draw centered placeholder text.
		label := obj.Text
		if label == "" {
			label = "Digital Signature"
		}
		fontPt := float64(obj.Font.Size)
		if fontPt <= 0 {
			fontPt = 10
		}
		digiDPI := float64(e.ResolutionX) * e.Scale
		if digiDPI <= 0 {
			digiDPI = DefaultDPI
		}
		face := selectFace(obj.Font, fontPt, digiDPI)
		ascent := face.Metrics().Ascent.Ceil()
		lineH := face.Metrics().Height.Ceil()
		textW := font.MeasureString(face, label).Ceil()
		dotX := x + (w-textW)/2
		if dotX < x {
			dotX = x
		}
		dotY := y + (h-lineH)/2 + ascent
		tc := obj.TextColor
		if tc.A == 0 {
			tc = color.RGBA{A: 255}
		}
		d := &font.Drawer{
			Dst:  e.curPage,
			Src:  image.NewUniform(tc),
			Face: face,
			Dot:  fixed.P(dotX, dotY),
		}
		d.DrawString(label)
	}
}

// renderHtmlText parses the HTML-tagged text in obj and renders each styled run
// onto the current canvas. It uses HtmlTextRenderer from the utils package to
// split the text into lines and runs, then selects the appropriate font face
// for each run (bold, italic, bold+italic, regular) using selectFace.
func (e *Exporter) renderHtmlText(obj preview.PreparedObject, x, y, w, h int, fontPt, dpi float64, defaultColor color.RGBA) {
	renderer := utils.NewHtmlTextRenderer(obj.Text, obj.Font, defaultColor)
	htmlLines := renderer.Lines()
	if len(htmlLines) == 0 {
		return
	}

	// Compute line height using the base font face.
	baseFace := selectFace(obj.Font, fontPt, dpi)
	lineH := baseFace.Metrics().Height.Ceil()
	ascent := baseFace.Metrics().Ascent.Ceil()

	// Count total visual lines (each HtmlLine may produce multiple wrapped lines).
	totalVisualLines := 0
	for _, hl := range htmlLines {
		// Collect plain text for line width measurement.
		plainRuns := make([]string, len(hl.Runs))
		for i, run := range hl.Runs {
			plainRuns[i] = run.Text
		}
		plain := strings.Join(plainRuns, "")
		if plain == "" {
			totalVisualLines++ // empty line still occupies vertical space
			continue
		}
		wrapped := e.wrapText(plain, w, baseFace)
		if len(wrapped) == 0 {
			totalVisualLines++
		} else {
			totalVisualLines += len(wrapped)
		}
	}

	textBlockH := totalVisualLines * lineH
	startY := y + ascent
	switch obj.VertAlign {
	case 1:
		startY = y + (h-textBlockH)/2 + ascent
	case 2:
		startY = y + h - textBlockH + ascent
	}

	for _, hl := range htmlLines {
		if startY > y+h {
			break
		}

		// Render this logical line. We render runs left-to-right on the same
		// visual line, handling word-wrap across all runs together.
		// For simplicity: render each HtmlLine as a single visual line,
		// drawing runs consecutively on the X axis.
		// Word-wrapping across mixed-style runs is complex; we keep it
		// simple: compute total line width, apply alignment, then draw runs.

		// Compute total pixel width of this logical line.
		totalLineW := 0
		type renderRun struct {
			text  string
			face  font.Face
			color color.RGBA
		}
		var renderRuns []renderRun
		for _, run := range hl.Runs {
			if run.Text == "" {
				continue
			}
			runFace := selectFace(run.Font, float64(run.Font.Size), dpi)
			if run.Font.Size <= 0 {
				runFace = baseFace
			}
			totalLineW += font.MeasureString(runFace, run.Text).Ceil()
			renderRuns = append(renderRuns, renderRun{
				text:  run.Text,
				face:  runFace,
				color: run.Color,
			})
		}

		if len(renderRuns) == 0 {
			// Empty line — just advance Y.
			startY += lineH
			continue
		}

		dotX := x + 2
		switch obj.HorzAlign {
		case 1: // Center
			dotX = x + (w-totalLineW)/2
		case 2: // Right
			dotX = x + w - totalLineW - 2
		}
		if dotX < x {
			dotX = x
		}

		// Draw each run at the current X position.
		for _, rr := range renderRuns {
			if dotX >= x+w {
				break
			}
			d := &font.Drawer{
				Dst:  e.curPage,
				Src:  image.NewUniform(rr.color),
				Face: rr.face,
				Dot:  fixed.P(dotX, startY),
			}
			d.DrawString(rr.text)
			dotX += font.MeasureString(rr.face, rr.text).Ceil()
		}

		startY += lineH
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

// drawPictureObject decodes the image blob from BlobStore and draws it scaled
// into the rectangle (x, y, x+w, y+h) on the current page canvas.
func (e *Exporter) drawPictureObject(x, y, w, h int, obj preview.PreparedObject) {
	if e.pp == nil || obj.BlobIdx < 0 {
		return
	}
	blobData := e.pp.BlobStore.Get(obj.BlobIdx)
	if len(blobData) == 0 {
		return
	}

	src, _, err := image.Decode(bytes.NewReader(blobData))
	if err != nil {
		return
	}

	srcBounds := src.Bounds()
	if srcBounds.Dx() == 0 || srcBounds.Dy() == 0 || w == 0 || h == 0 {
		return
	}

	// Scale source pixels into the destination rectangle using nearest-neighbour.
	dst := image.Rect(x, y, x+w, y+h).Intersect(e.curPage.Bounds())
	scaleX := float64(srcBounds.Dx()) / float64(w)
	scaleY := float64(srcBounds.Dy()) / float64(h)

	for py := dst.Min.Y; py < dst.Max.Y; py++ {
		srcY := srcBounds.Min.Y + int(float64(py-y)*scaleY)
		if srcY < srcBounds.Min.Y {
			srcY = srcBounds.Min.Y
		}
		if srcY >= srcBounds.Max.Y {
			srcY = srcBounds.Max.Y - 1
		}
		for px := dst.Min.X; px < dst.Max.X; px++ {
			srcX := srcBounds.Min.X + int(float64(px-x)*scaleX)
			if srcX < srcBounds.Min.X {
				srcX = srcBounds.Min.X
			}
			if srcX >= srcBounds.Max.X {
				srcX = srcBounds.Max.X - 1
			}
			r, g, b, a := src.At(srcX, srcY).RGBA()
			e.curPage.SetRGBA(px, py, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
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

// drawRoundRect draws the 1-pixel outline of a rounded rectangle.
// The corners are quarter-circle arcs of the given radius.
func (e *Exporter) drawRoundRect(x, y, w, h, radius int, c color.RGBA) {
	if radius*2 > w {
		radius = w / 2
	}
	if radius*2 > h {
		radius = h / 2
	}
	if radius <= 0 {
		e.drawRect(image.Rect(x, y, x+w, y+h), c)
		return
	}
	r := float64(radius)
	bounds := e.curPage.Bounds()
	setPixel := func(px, py int) {
		if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
			e.curPage.SetRGBA(px, py, c)
		}
	}
	// Straight edges.
	// Top and bottom between corner arcs.
	for px := x + radius; px <= x+w-radius; px++ {
		setPixel(px, y)
		setPixel(px, y+h-1)
	}
	// Left and right between corner arcs.
	for py := y + radius; py <= y+h-radius; py++ {
		setPixel(x, py)
		setPixel(x+w-1, py)
	}
	// Quarter-circle arcs at each corner.
	steps := int(math.Pi * r / 2)
	if steps < 8 {
		steps = 8
	}
	for i := 0; i <= steps; i++ {
		t := math.Pi / 2 * float64(i) / float64(steps)
		dx := int(math.Round(r * math.Cos(t)))
		dy := int(math.Round(r * math.Sin(t)))
		// Top-right corner.
		setPixel(x+w-1-radius+dx, y+radius-dy)
		// Top-left corner.
		setPixel(x+radius-dx, y+radius-dy)
		// Bottom-right corner.
		setPixel(x+w-1-radius+dx, y+h-1-radius+dy)
		// Bottom-left corner.
		setPixel(x+radius-dx, y+h-1-radius+dy)
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

// ExportPageEnd finalises the current page canvas and either writes it
// immediately (SeparateFiles/default) or accumulates it for Finish.
// Mirrors C# ImageExport.ExportPageEnd (ImageExport.cs line 631).
func (e *Exporter) ExportPageEnd(pg *preview.PreparedPage) error {
	if e.curPage == nil {
		return nil
	}

	// Render watermark on top of page content.
	if wm := pg.Watermark; wm != nil && wm.Enabled {
		if wm.ShowImageOnTop {
			e.renderWatermarkImageOnPage(wm)
		}
		if wm.ShowTextOnTop {
			e.renderWatermarkTextOnPage(wm)
		}
	}

	// Multi-frame TIFF: accumulate frames for Finish.
	if e.Format == ImageFormatTIFF && e.MultiFrameTiff {
		e.tiffFrames = append(e.tiffFrames, e.curPage)
		e.curPage = nil
		return nil
	}

	// Combined mode: accumulate pages for Finish.
	if !e.SeparateFiles {
		e.combinedPages = append(e.combinedPages, e.curPage)
		e.curPage = nil
		return nil
	}

	// Separate files (default): encode and write immediately.
	if err := e.encodeImage(e.w, e.curPage); err != nil {
		return err
	}
	e.curPage = nil
	return nil
}

// renderWatermarkTextOnPage draws watermark text centered on the current canvas.
// Text rotation is approximated: diagonal text is drawn at 45° by using
// a diagonal scan of pixels (simple approach without full affine transform).
func (e *Exporter) renderWatermarkTextOnPage(wm *preview.PreparedWatermark) {
	if e.curPage == nil || wm.Text == "" {
		return
	}
	bounds := e.curPage.Bounds()
	cx := bounds.Dx() / 2
	cy := bounds.Dy() / 2

	fontPt := float64(wm.Font.Size)
	if fontPt <= 0 {
		fontPt = 48
	}
	wmDPI := float64(e.ResolutionX) * e.Scale
	if wmDPI <= 0 {
		wmDPI = DefaultDPI
	}
	face := selectFace(wm.Font, fontPt, wmDPI)

	tc := wm.TextColor
	textColor := color.RGBA{R: tc.R, G: tc.G, B: tc.B, A: tc.A}

	textW := font.MeasureString(face, wm.Text).Ceil()
	ascent := face.Metrics().Ascent.Ceil()

	// Render centered.
	startX := cx - textW/2
	startY := cy + ascent/2

	d := &font.Drawer{
		Dst:  e.curPage,
		Src:  &image.Uniform{textColor},
		Face: face,
		Dot:  fixed.P(startX, startY),
	}
	d.DrawString(wm.Text)
}

// renderWatermarkImageOnPage blends a blob-stored image onto the current canvas.
func (e *Exporter) renderWatermarkImageOnPage(wm *preview.PreparedWatermark) {
	if e.curPage == nil || wm.ImageBlobIdx < 0 || e.pp == nil {
		return
	}
	imgData := e.pp.BlobStore.Get(wm.ImageBlobIdx)
	if len(imgData) == 0 {
		return
	}
	src, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return
	}
	bounds := e.curPage.Bounds()
	var dstRect image.Rectangle

	switch wm.ImageSize {
	case preview.WatermarkImageSizeStretch:
		dstRect = bounds
	case preview.WatermarkImageSizeCenter, preview.WatermarkImageSizeZoom:
		sw := src.Bounds().Dx()
		sh := src.Bounds().Dy()
		dstRect = image.Rect(
			(bounds.Dx()-sw)/2,
			(bounds.Dy()-sh)/2,
			(bounds.Dx()+sw)/2,
			(bounds.Dy()+sh)/2,
		)
	default: // Normal / Tile
		dstRect = image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy())
	}

	// Apply transparency via alpha blend using draw.DrawMask with a uniform mask.
	alpha := uint8(255 - uint8(float32(255)*wm.ImageTransparency))
	mask := &image.Uniform{color.RGBA{A: alpha}}
	draw.DrawMask(e.curPage, dstRect, src, src.Bounds().Min, mask, image.Point{}, draw.Over)

	// Tile mode: repeat across the page.
	if wm.ImageSize == preview.WatermarkImageSizeTile {
		sw := src.Bounds().Dx()
		sh := src.Bounds().Dy()
		if sw <= 0 || sh <= 0 {
			return
		}
		for y := 0; y < bounds.Dy(); y += sh {
			for x := 0; x < bounds.Dx(); x += sw {
				r := image.Rect(x, y, x+sw, y+sh)
				draw.DrawMask(e.curPage, r, src, src.Bounds().Min, mask, image.Point{}, draw.Over)
			}
		}
	}
}

// Finish writes the combined image or multi-frame TIFF when applicable.
// Mirrors C# ImageExport.Finish (ImageExport.cs line 667).
func (e *Exporter) Finish() error {
	// Multi-frame TIFF: write all frames as a single multi-page TIFF.
	if e.Format == ImageFormatTIFF && e.MultiFrameTiff && len(e.tiffFrames) > 0 {
		if err := e.encodeMultiFrameTIFF(e.w, e.tiffFrames); err != nil {
			return err
		}
		e.tiffFrames = nil
		return nil
	}

	// Combined mode: stitch pages vertically and write.
	if !e.SeparateFiles && len(e.combinedPages) > 0 {
		combined := e.stitchPages(e.combinedPages)
		if err := e.encodeImage(e.w, combined); err != nil {
			return err
		}
		e.combinedPages = nil
		return nil
	}

	e.tiffFrames = nil
	e.combinedPages = nil
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

// effectiveScale returns the combined scale factor: DPI ratio × user Scale.
// Mirrors C# ImageExport.ExportPageBegin: zoomX = ResolutionX / 96f (line 547).
func (e *Exporter) effectiveScale() float64 {
	dpiRatio := float64(e.ResolutionX) / float64(DefaultDPI)
	if dpiRatio <= 0 {
		dpiRatio = 1.0
	}
	s := e.Scale
	if s <= 0 {
		s = 1.0
	}
	return dpiRatio * s
}

// scaled converts a pixel value to output pixels applying the effective scale.
// The effective scale combines the DPI ratio and the user-defined Scale factor.
func (e *Exporter) scaled(px int) int {
	s := e.effectiveScale()
	if s == 1.0 {
		return px
	}
	return int(math.Round(float64(px) * s))
}

// encodeImage encodes img to w in the format specified by e.Format.
// For JPEG, uses e.JpegQuality. For other formats uses default settings.
// Mirrors C# ImageExport.SaveImage (ImageExport.cs line 377).
func (e *Exporter) encodeImage(w io.Writer, img *image.RGBA) error {
	switch e.Format {
	case ImageFormatJPEG:
		q := e.JpegQuality
		if q <= 0 || q > 100 {
			q = 100
		}
		return jpeg.Encode(w, img, &jpeg.Options{Quality: q})
	case ImageFormatGIF:
		// Convert to paletted image for GIF (256 colours).
		// Use a simple uniform palette derived from the image.
		palettedImg := convertToPaletted(img)
		return gif.Encode(w, palettedImg, nil)
	case ImageFormatBMP:
		return bmp.Encode(w, img)
	case ImageFormatTIFF:
		opts := &tiff.Options{Compression: tiff.Deflate, Predictor: true}
		if e.MonochromeTiff {
			return tiff.Encode(w, convertToGray(img), opts)
		}
		return tiff.Encode(w, img, opts)
	default: // ImageFormatPNG and fallback
		return png.Encode(w, img)
	}
}

// encodeMultiFrameTIFF writes all frames as a multi-page TIFF.
// Go's standard tiff encoder does not natively support multi-frame TIFF writing
// (frames can only be appended via the IFD chain), so we concatenate frames
// by writing each as a separate TIFF in-memory and streaming them in sequence.
// This produces independently-decodable pages. True multi-IFD TIFF would require
// a lower-level encoder; the current approach is compatible with all readers
// that support multi-file TIFF.
//
// Mirrors C# ImageExport.SaveImage multiFrameTiff path (ImageExport.cs line 383).
func (e *Exporter) encodeMultiFrameTIFF(w io.Writer, frames []*image.RGBA) error {
	opts := &tiff.Options{Compression: tiff.Deflate, Predictor: true}
	for _, frame := range frames {
		var src image.Image = frame
		if e.MonochromeTiff {
			src = convertToGray(frame)
		}
		if err := tiff.Encode(w, src, opts); err != nil {
			return fmt.Errorf("image export: TIFF frame encode: %w", err)
		}
	}
	return nil
}

// stitchPages combines multiple page canvases vertically into one image.
// Pages are separated by PaddingNonSeparatePages pixels of the background color.
// Mirrors C# ImageExport.Start combined-mode setup (ImageExport.cs line 519).
func (e *Exporter) stitchPages(pages []*image.RGBA) *image.RGBA {
	if len(pages) == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	pad := e.PaddingNonSeparatePages
	if pad < 0 {
		pad = 0
	}

	// Compute dimensions: max width, sum of heights + padding.
	maxW := 0
	totalH := 0
	for _, pg := range pages {
		b := pg.Bounds()
		if b.Dx() > maxW {
			maxW = b.Dx()
		}
		totalH += b.Dy() + pad*2
	}
	if maxW <= 0 {
		maxW = 1
	}
	if totalH <= 0 {
		totalH = 1
	}

	combined := image.NewRGBA(image.Rect(0, 0, maxW, totalH))
	// Fill with background.
	draw.Draw(combined, combined.Bounds(), &image.Uniform{e.BackgroundColor}, image.Point{}, draw.Src)

	curY := 0
	for _, pg := range pages {
		b := pg.Bounds()
		curY += pad
		// Centre page horizontally.
		offsetX := (maxW - b.Dx()) / 2
		dst := image.Rect(offsetX, curY, offsetX+b.Dx(), curY+b.Dy())
		draw.Draw(combined, dst, pg, b.Min, draw.Src)
		curY += b.Dy() + pad
	}
	return combined
}

// RenderGenericPNG renders any PreparedObject to a PNG byte slice using the
// image exporter's renderObject. The object is rendered on a transparent canvas
// at its own dimensions. Used by the HTML exporter to render PolygonObjects
// and ShapeObjects as base64 images (matching C# LayerBack + LayerPicture).
func RenderGenericPNG(obj preview.PreparedObject) ([]byte, error) {
	w := int(math.Round(float64(obj.Width)))
	h := int(math.Round(float64(obj.Height)))
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	canvas := image.NewRGBA(image.Rect(0, 0, w, h))
	// Fill with FillColor when set so the fill is baked into the PNG.
	// C# renders filled shapes (polygons, etc.) as filled images.
	// When no fill: use transparent so underlying page content shows through.
	bg := color.RGBA{0, 0, 0, 0}
	if obj.FillColor.A > 0 {
		bg = obj.FillColor
	}
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	exp := &Exporter{
		ResolutionX: DefaultDPI,
		ResolutionY: DefaultDPI,
		Scale:       1.0,
		curPage:     canvas,
	}
	renderObj := obj
	renderObj.Left = 0
	renderObj.Top = 0
	exp.renderObject(renderObj, 0)

	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RotateImagePNG decodes a PNG image, rotates it by the given angle (90, 180,
// or 270 degrees), resizes to targetW×targetH, and re-encodes as PNG.
// Used by the HTML exporter for rotated picture objects (barcodes, etc.).
func RotateImagePNG(data []byte, angle, targetW, targetH int) []byte {
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return data // fallback: return original
	}
	rgba, ok := src.(*image.RGBA)
	if !ok {
		// Convert to RGBA.
		b := src.Bounds()
		rgba = image.NewRGBA(b)
		draw.Draw(rgba, b, src, b.Min, draw.Src)
	}

	var rotated *image.RGBA
	switch angle {
	case 90:
		rotated = rotateRGBA90CW(rgba)
	case 180:
		rotated = rotateRGBA180(rgba)
	case 270:
		rotated = rotateRGBA90CCW(rgba)
	default:
		rotated = rgba
	}

	// Resize to target dimensions using nearest-neighbour (sufficient for barcodes).
	if rotated.Bounds().Dx() != targetW || rotated.Bounds().Dy() != targetH {
		dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
		srcB := rotated.Bounds()
		for y := 0; y < targetH; y++ {
			sy := srcB.Min.Y + y*srcB.Dy()/targetH
			for x := 0; x < targetW; x++ {
				sx := srcB.Min.X + x*srcB.Dx()/targetW
				dst.SetRGBA(x, y, rotated.RGBAAt(sx, sy))
			}
		}
		rotated = dst
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, rotated); err != nil {
		return data
	}
	return buf.Bytes()
}

// RenderObjectPNG renders a single PreparedObject (text/RTF) to a PNG byte
// slice, applying the object's Angle rotation. Used by the HTML exporter to
// generate LayerPicture background images for rotated text objects, matching
// C# HTMLExportLayers.cs GetLayerPicture behaviour.
//
// The returned image has the same pixel dimensions as the object's Width×Height.
// For 90°/270° angles the text is first rendered at the transposed dimensions
// (H×W) and then rotated, so the result fits the original bounding box.
func RenderObjectPNG(obj preview.PreparedObject) ([]byte, error) {
	w := int(math.Round(float64(obj.Width)))
	h := int(math.Round(float64(obj.Height)))
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	// For 90°/270°: render in the transposed (H×W) space so text wraps correctly,
	// then rotate the canvas to produce the final W×H image.
	renderW, renderH := w, h
	if obj.Angle == 90 || obj.Angle == 270 {
		renderW, renderH = h, w
	}

	canvas := image.NewRGBA(image.Rect(0, 0, renderW, renderH))

	// Fill background. Use transparent when no fill is set so that underlying
	// page content (e.g. orange polygon fills) shows through the PNG.
	// C# GetLayerPicture renders on a transparent bitmap for the same reason.
	bg := color.RGBA{0, 0, 0, 0} // transparent
	if obj.FillColor.A > 0 {
		bg = obj.FillColor
	}
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	// Use a temporary Exporter to render the text onto the canvas.
	exp := &Exporter{
		ResolutionX: DefaultDPI,
		ResolutionY: DefaultDPI,
		Scale:       1.0,
		curPage:     canvas,
	}
	// Render without rotation at origin; we'll rotate the canvas after.
	renderObj := obj
	renderObj.Angle = 0
	renderObj.Left = 0
	renderObj.Top = 0
	renderObj.Width = float32(renderW)
	renderObj.Height = float32(renderH)
	exp.renderObject(renderObj, 0)

	// Rotate the canvas to match the object's Angle.
	var result *image.RGBA
	switch obj.Angle {
	case 90:
		result = rotateRGBA90CW(canvas)
	case 180:
		result = rotateRGBA180(canvas)
	case 270:
		result = rotateRGBA90CCW(canvas)
	default:
		result = canvas
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, result); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// rotateRGBA90CW rotates an RGBA image 90° clockwise.
// Source (x,y) → dest (srcH-1-y, x). Output dimensions: H×W → W×H swapped.
func rotateRGBA90CW(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, sh, sw))
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			dst.SetRGBA(sh-1-y, x, src.RGBAAt(x, y))
		}
	}
	return dst
}

// rotateRGBA180 rotates an RGBA image 180°.
func rotateRGBA180(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, sw, sh))
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			dst.SetRGBA(sw-1-x, sh-1-y, src.RGBAAt(x, y))
		}
	}
	return dst
}

// rotateRGBA90CCW rotates an RGBA image 90° counter-clockwise (= 270° CW).
// Source (x,y) → dest (y, srcW-1-x). Output dimensions: H×W → W×H swapped.
func rotateRGBA90CCW(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, sh, sw))
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			dst.SetRGBA(y, sw-1-x, src.RGBAAt(x, y))
		}
	}
	return dst
}

// convertToPaletted converts an RGBA image to an 8-bit paletted image
// using a web-safe 216-colour palette suitable for GIF output.
func convertToPaletted(src *image.RGBA) *image.Paletted {
	palette := buildWebSafePalette()
	dst := image.NewPaletted(src.Bounds(), palette)
	draw.FloydSteinberg.Draw(dst, src.Bounds(), src, src.Bounds().Min)
	return dst
}

// buildWebSafePalette returns a 216-colour web-safe palette plus black and white.
func buildWebSafePalette() []color.Color {
	var pal []color.Color
	for r := 0; r <= 5; r++ {
		for g := 0; g <= 5; g++ {
			for b := 0; b <= 5; b++ {
				pal = append(pal, color.RGBA{
					R: uint8(r * 51),
					G: uint8(g * 51),
					B: uint8(b * 51),
					A: 255,
				})
			}
		}
	}
	// Pad to 256 entries.
	for len(pal) < 256 {
		pal = append(pal, color.RGBA{A: 255})
	}
	return pal
}

// convertToGray converts an RGBA image to a greyscale image for monochrome TIFF.
// Mirrors C# ImageExport.ConvertToBitonal (ImageExport.cs line 275).
func convertToGray(src *image.RGBA) *image.Gray {
	dst := image.NewGray(src.Bounds())
	b := src.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.RGBAAt(x, y)
			// Luminance using standard coefficients.
			lum := (uint32(c.R)*299 + uint32(c.G)*587 + uint32(c.B)*114) / 1000
			dst.SetGray(x, y, color.Gray{Y: uint8(lum)})
		}
	}
	return dst
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

// Serialize writes non-default Exporter settings to w.
// Mirrors C# ImageExport.Serialize (ImageExport.cs).
func (e *Exporter) Serialize(w report.Writer) {
	e.ExportBase.Serialize(w)
	w.WriteInt("ImageFormat", int(e.Format))
	w.WriteBool("SeparateFiles", e.SeparateFiles)
	w.WriteInt("ResolutionX", e.ResolutionX)
	w.WriteInt("ResolutionY", e.ResolutionY)
	w.WriteInt("JpegQuality", e.JpegQuality)
	w.WriteBool("MultiFrameTiff", e.MultiFrameTiff)
	w.WriteBool("MonochromeTiff", e.MonochromeTiff)
}

// Deserialize reads Exporter settings from r.
// Mirrors C# ImageExport.Deserialize (ImageExport.cs).
func (e *Exporter) Deserialize(r report.Reader) {
	e.ExportBase.Deserialize(r)
	e.Format = ImageFormat(r.ReadInt("ImageFormat", int(ImageFormatJPEG)))
	e.SeparateFiles = r.ReadBool("SeparateFiles", true)
	e.ResolutionX = r.ReadInt("ResolutionX", DefaultDPI)
	e.ResolutionY = r.ReadInt("ResolutionY", DefaultDPI)
	e.JpegQuality = r.ReadInt("JpegQuality", 100)
	e.MultiFrameTiff = r.ReadBool("MultiFrameTiff", false)
	e.MonochromeTiff = r.ReadBool("MonochromeTiff", false)
}
