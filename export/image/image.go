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
	"golang.org/x/image/vector"

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
	// For diagonal lines, w or h can be negative (encoding direction).
	// Skip the min-1 clamping so the direction is preserved for drawLine.
	if obj.Kind != preview.ObjectTypeLine || !obj.LineDiagonal {
		if w < 1 {
			w = 1
		}
		if h < 1 {
			h = 1
		}
	}
	// Use absolute values for the bounding rectangle (needed for diagonal lines
	// with negative w or h).
	bx0, bx1 := x, x+w
	if bx0 > bx1 {
		bx0, bx1 = bx1, bx0
	}
	by0, by1 := y, y+h
	if by0 > by1 {
		by0, by1 = by1, by0
	}
	bounds := image.Rect(bx0, by0, bx1, by1).Intersect(e.curPage.Bounds())

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
		lineWidth := 1
		if obj.Border.Lines[0] != nil && obj.Border.Lines[0].Width > 0 {
			lineWidth = int(math.Round(float64(obj.Border.Lines[0].Width)))
		}
		if obj.LineDiagonal {
			x1, y1 := x, y
			x2, y2 := x+w, y+h

			// C# shortens the line by cap insets so the line doesn't overlap caps.
			// C# ref: LineObject.cs:144-158 — insets from GetCustomCapPath.
			// Insets: Arrow=0, Circle=H/2, Square=H/2, Diamond=H/1.4.
			// The inset is applied along the line direction (in line-local Y axis).
			lx1, ly1 := float64(x1), float64(y1)
			lx2, ly2 := float64(x2), float64(y2)
			ldx := lx2 - lx1
			ldy := ly2 - ly1
			lineLen := math.Sqrt(ldx*ldx + ldy*ldy)
			if lineLen > 0 {
				capInset := func(cap preview.LineCap) float64 {
					ch := float64(cap.Height)
					if ch <= 0 {
						ch = 8
					}
					switch cap.Style {
					case preview.LineCapStyleCircle, preview.LineCapStyleSquare:
						return ch / 2 * float64(lineWidth)
					case preview.LineCapStyleDiamond:
						return ch / 1.4 * float64(lineWidth)
					}
					return 0 // Arrow has inset=0
				}
				startInset := capInset(obj.LineStartCap)
				endInset := capInset(obj.LineEndCap)
				// Move start point forward along line by startInset
				ux, uy := ldx/lineLen, ldy/lineLen
				lx1 += ux * startInset
				ly1 += uy * startInset
				// Move end point backward along line by endInset
				lx2 -= ux * endInset
				ly2 -= uy * endInset
			}
			ix1, iy1 := int(math.Round(lx1)), int(math.Round(ly1))
			ix2, iy2 := int(math.Round(lx2)), int(math.Round(ly2))

			// Draw thick diagonal line (shortened by cap insets).
			e.drawLine(ix1, iy1, ix2, iy2, lineColor)
			dx := ix2 - ix1
			if dx < 0 {
				dx = -dx
			}
			dy := ix2 - ix1 // intentionally using same for perpendicular check
			dy = iy2 - iy1
			if dy < 0 {
				dy = -dy
			}
			for i := 1; i < lineWidth; i++ {
				if dx >= dy {
					e.drawLine(ix1, iy1+i, ix2, iy2+i, lineColor)
				} else {
					e.drawLine(ix1+i, iy1, ix2+i, iy2, lineColor)
				}
			}

			// Draw caps on top of the line (at original endpoints, not inset ones).
			e.drawLineCap(obj.LineStartCap, x1, y1, x2, y2, lineWidth, lineColor, true)
			e.drawLineCap(obj.LineEndCap, x1, y1, x2, y2, lineWidth, lineColor, false)
		} else {
			// Non-diagonal: horizontal or vertical filled rectangle.
			if w >= h {
				// Horizontal line: fill lineWidth pixel rows centered at y+h/2.
				cy := y + h/2
				for i := 0; i < lineWidth; i++ {
					e.drawHLine(x, cy+i, x+w, lineColor)
				}
			} else {
				// Vertical line: fill lineWidth pixel columns centered at x+w/2.
				cx := x + w/2
				for i := 0; i < lineWidth; i++ {
					e.drawVLine(cx+i, y, y+h, lineColor)
				}
			}
		}

	case preview.ObjectTypeShape:
		// C# ShapeObject.Draw fills then outlines the polygon/shape.
		// Ref: ShapeObject.cs:156-189 — g.FillAndDrawPolygon for Diamond/Triangle,
		// g.FillAndDrawRectangle for Rectangle, g.FillAndDrawEllipse for Ellipse.
		// Each shape fills only its geometric area (not the full bounding rect).
		fc := obj.FillColor
		shapeColor := color.RGBA{A: 255}
		if obj.Border.Lines[0] != nil && obj.Border.Lines[0].Color.A > 0 {
			shapeColor = obj.Border.Lines[0].Color
		}
		shapePenW := 1
		if obj.Border.Lines[0] != nil && obj.Border.Lines[0].Width > 0 {
			shapePenW = int(math.Round(float64(obj.Border.Lines[0].Width)))
		}
		// drawThickShapeLine draws a line with the shape's border width.
		drawThickLine := func(x0, y0, x1, y1 int) {
			e.drawLine(x0, y0, x1, y1, shapeColor)
			adx, ady := x1-x0, y1-y0
			if adx < 0 {
				adx = -adx
			}
			if ady < 0 {
				ady = -ady
			}
			for i := 1; i < shapePenW; i++ {
				if adx >= ady {
					e.drawLine(x0, y0+i, x1, y1+i, shapeColor)
				} else {
					e.drawLine(x0+i, y0, x1+i, y1, shapeColor)
				}
			}
		}
		switch obj.ShapeKind {
		case 2: // Ellipse — fill ellipse, then draw outline.
			if fc.A > 0 {
				e.fillEllipse(x, y, w, h, fc)
			}
			// Draw thick ellipse by drawing multiple concentric outlines.
			for i := 0; i < shapePenW; i++ {
				e.drawEllipse(x+i, y+i, w-i*2, h-i*2, shapeColor)
			}
		case 3: // Triangle — fill polygon, then draw outline.
			// Inset bottom edge by 1px so thick lines don't clip at canvas boundary.
			bInset := 1
			triPts := []image.Point{
				{X: x + w/2, Y: y},
				{X: x + w - bInset, Y: y + h - bInset},
				{X: x, Y: y + h - bInset},
			}
			if fc.A > 0 {
				e.fillPolygon(triPts, fc)
			}
			drawThickLine(x+w/2, y, x+w-bInset, y+h-bInset)
			// Bottom edge: use drawHLine for crisp horizontal rendering.
			for i := 0; i < shapePenW; i++ {
				e.drawHLine(x, y+h-bInset-i, x+w-bInset, shapeColor)
			}
			drawThickLine(x, y+h-bInset, x+w/2, y)
		case 4: // Diamond — fill polygon, then draw outline.
			diaPts := []image.Point{
				{X: x + w/2, Y: y},
				{X: x + w, Y: y + h/2},
				{X: x + w/2, Y: y + h},
				{X: x, Y: y + h/2},
			}
			if fc.A > 0 {
				e.fillPolygon(diaPts, fc)
			}
			drawThickLine(x+w/2, y, x+w, y+h/2)
			drawThickLine(x+w, y+h/2, x+w/2, y+h)
			drawThickLine(x+w/2, y+h, x, y+h/2)
			drawThickLine(x, y+h/2, x+w/2, y)
		case 1: // RoundRectangle — fill then draw outline.
			if !bounds.Empty() && fc.A > 0 {
				draw.Draw(e.curPage, bounds, &image.Uniform{fc}, image.Point{}, draw.Over)
			}
			radius := int(math.Round(float64(obj.ShapeCurve)))
			for i := 0; i < shapePenW; i++ {
				if radius <= 0 {
					e.drawRect(image.Rect(x+i, y+i, x+w-i, y+h-i), shapeColor)
				} else {
					e.drawRoundRect(x+i, y+i, w-i*2, h-i*2, radius, shapeColor)
				}
			}
		default: // Rectangle (0) — fill bounding rect, then draw outline.
			if !bounds.Empty() && fc.A > 0 {
				draw.Draw(e.curPage, bounds, &image.Uniform{fc}, image.Point{}, draw.Over)
			}
			for i := 0; i < shapePenW; i++ {
				e.drawRect(image.Rect(x+i, y+i, x+w-i, y+h-i), shapeColor)
			}
		}
		e.drawBorderLines(obj.Border, x, y, w, h)

	case preview.ObjectTypeCheckBox:
		if !bounds.Empty() {
			// C# CheckBoxObject.Draw: base.Draw() → DrawBackground → Fill.Draw().
			// Default fill is SolidFill(Color.Transparent) so the background is
			// transparent unless the object has an explicit FillColor (e.g. from
			// EvenStyle). No box outline is drawn; Border.Draw handles borders
			// separately. In GetLayerPicture, borders are set to None.
			// Reference: CheckBoxObject.cs:290-296, ReportComponentBase.cs:773-778,
			//            Fills.cs:173, HTMLExportLayers.cs:516-519.
			if obj.FillColor.A > 0 {
				draw.Draw(e.curPage, bounds, &image.Uniform{obj.FillColor}, image.Point{}, draw.Over)
			}

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

			// Match C# DrawCheck padding: ratio = Width / (5mm * 3.78), pad = 4 * ratio.
			// Pen width: 1.6 * ratio * CheckWidthRatio (C# CheckBoxObject.cs:210-212).
			ratio := float64(w) / 18.9
			padF := 4.0 * ratio
			pad := int(math.Round(padF))
			if pad < 1 {
				pad = 1
			}
			checkWidthRatio := float64(obj.CheckWidthRatio)
			if checkWidthRatio <= 0 {
				checkWidthRatio = 1.0
			}
			thickness := int(math.Round(1.6 * ratio * checkWidthRatio))
			if thickness < 1 {
				thickness = 1
			}
			switch symbol {
			case 0: // checkmark
				// Match C# DrawCheck: lines from (Left, Top+H*5/10) to (Left+W*4/10, Bottom-H/10)
				// to (Right, Top+H/10). Reference: CheckBoxObject.cs:222-225.
				dx := int(padF)
				dy := int(padF)
				dw := w - 2*dx
				dh := h - 2*dy
				mx := x + dx + int(float64(dw)*0.4)
				my := y + dy + dh - int(float64(dh)*0.1)
				e.drawThickLine(x+dx, y+dy+int(float64(dh)*0.5), mx, my, thickness, cc)
				e.drawThickLine(mx, my, x+dx+dw, y+dy+int(float64(dh)*0.1), thickness, cc)
			case 1: // cross (X)
				e.drawThickLine(x+pad, y+pad, x+w-pad, y+h-pad, thickness, cc)
				e.drawThickLine(x+w-pad, y+pad, x+pad, y+h-pad, thickness, cc)
			case 2: // plus (+)
				e.drawThickLine(x+w/2, y+pad, x+w/2, y+h-pad, thickness, cc)
				e.drawThickLine(x+pad, y+h/2, x+w-pad, y+h/2, thickness, cc)
			case 3: // fill (filled rectangle)
				fillBounds := image.Rect(x+pad, y+pad, x+w-pad, y+h-pad)
				draw.Draw(e.curPage, fillBounds, &image.Uniform{cc}, image.Point{}, draw.Over)
			case 10: // minus (-)
				e.drawThickLine(x+pad, y+h/2, x+w-pad, y+h/2, thickness, cc)
			case 11: // slash (/)
				e.drawThickLine(x+pad, y+h-pad, x+w-pad, y+pad, thickness, cc)
			case 12: // backslash (\)
				e.drawThickLine(x+pad, y+pad, x+w-pad, y+h-pad, thickness, cc)
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
		// Build pixel points (engine stores them as bounding-box-relative after
		// CenterX/CenterY transform — see engine/objects.go PolygonObject case).
		n := len(obj.Points)
		imgPts := make([]image.Point, n)
		for i, p := range obj.Points {
			imgPts[i] = image.Point{X: x + e.scaled(int(p[0])), Y: y + e.scaled(int(p[1]))}
		}
		// Fill polygon interior before drawing outline (mirrors C# FillPath + DrawPath).
		// Ref: PolyLineObject.cs DoDrawPoly / PolygonObject.cs drawPoly.
		if obj.FillColor.A > 0 {
			e.fillPolygon(imgPts, obj.FillColor)
		}
		lineColor := color.RGBA{A: 255}
		if obj.Border.Lines[0] != nil && obj.Border.Lines[0].Color.A > 0 {
			lineColor = obj.Border.Lines[0].Color
		}
		for i := 0; i < n; i++ {
			next := (i + 1) % n
			e.drawLine(imgPts[i].X, imgPts[i].Y, imgPts[next].X, imgPts[next].Y, lineColor)
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
// drawThickLine draws a line from (x0,y0) to (x1,y1) with the given thickness
// by filling a circular disk of radius r = thickness/2 at each Bresenham pixel.
// Mirrors the effect of C# Graphics.DrawLine with a Pen of given width.
func (e *Exporter) drawThickLine(x0, y0, x1, y1, thickness int, c color.RGBA) {
	if thickness <= 1 {
		e.drawLine(x0, y0, x1, y1, c)
		return
	}
	r := thickness / 2
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
	xerr := dx - dy
	bounds := e.curPage.Bounds()
	for {
		// Fill a disk of radius r centred at (x0, y0).
		for dy2 := -r; dy2 <= r; dy2++ {
			for dx2 := -r; dx2 <= r; dx2++ {
				if dx2*dx2+dy2*dy2 <= r*r {
					px, py := x0+dx2, y0+dy2
					if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
						e.curPage.SetRGBA(px, py, c)
					}
				}
			}
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * xerr
		if e2 > -dy {
			xerr -= dy
			x0 += sx
		}
		if e2 < dx {
			xerr += dx
			y0 += sy
		}
	}
}

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

// drawLineCap draws a cap decoration (arrow, circle, square, diamond) at a
// line endpoint using the golang.org/x/image/vector rasterizer for anti-aliased
// rendering. Mirrors C# LineObject.Draw cap rendering (LineObject.cs:130-179)
// with SmoothingMode.AntiAlias enabled.
func (e *Exporter) drawLineCap(cap preview.LineCap, x1, y1, x2, y2, lineWidth int, c color.RGBA, isStart bool) {
	if cap.Style == preview.LineCapStyleNone {
		return
	}

	capW := float64(cap.Width)
	capH := float64(cap.Height)
	if capW <= 0 {
		capW = 8
	}
	if capH <= 0 {
		capH = 8
	}

	// C# ref: scale = Border.Width * e.ScaleX (applied via ScaleTransform).
	scale := float64(lineWidth)

	// Compute line angle. C# uses atan2(dx, dy) — note reversed args.
	angleDeg := math.Atan2(float64(x2-x1), float64(y2-y1)) * 180.0 / math.Pi

	var cx, cy float64
	var rotDeg float64
	if isStart {
		cx, cy = float64(x1), float64(y1)
		rotDeg = 180 - angleDeg
	} else {
		cx, cy = float64(x2), float64(y2)
		rotDeg = -angleDeg
	}

	rotRad := rotDeg * math.Pi / 180.0
	sinR := math.Sin(rotRad)
	cosR := math.Cos(rotRad)

	// GDI+ counterclockwise rotation: x'=x*cos-y*sin, y'=x*sin+y*cos.
	// Scale applied before rotation.
	tf := func(px, py float64) (float32, float32) {
		spx, spy := px*scale, py*scale
		rx := spx*cosR - spy*sinR + cx
		ry := spx*sinR + spy*cosR + cy
		return float32(rx), float32(ry)
	}

	// Use the vector rasterizer for anti-aliased path rendering.
	bounds := e.curPage.Bounds()
	r := &vector.Rasterizer{}
	r.Reset(bounds.Dx(), bounds.Dy())

	// strokePath draws a path by tracing it forward and backward with an offset,
	// creating a filled polygon that approximates a stroked line with the given width.
	// C# draws caps with pen width=1 (the scale is baked into ScaleTransform).
	penW := float32(1.0) // C# cap pen width = 1

	switch cap.Style {
	case preview.LineCapStyleArrow:
		// C# Arrow: AddLine(0,0 → -W,-H), AddLine(0,0 → W,-H).
		tx0, ty0 := tf(0, 0)
		tx1, ty1 := tf(-capW, -capH)
		tx2, ty2 := tf(capW, -capH)
		// Left arm
		strokeLine(r, tx0, ty0, tx1, ty1, penW)
		// Right arm
		strokeLine(r, tx0, ty0, tx2, ty2, penW)

	case preview.LineCapStyleCircle:
		// C# Circle: AddEllipse(-W/2, -H/2, W, H).
		tcx, tcy := tf(0, 0)
		rw := float32(capW * scale / 2)
		rh := float32(capH * scale / 2)
		strokeEllipse(r, tcx, tcy, rw, rh, penW)

	case preview.LineCapStyleSquare:
		// C# Square: AddRectangle(-W/2, -H/2, W, H).
		hw, hh := capW/2, capH/2
		ax, ay := tf(-hw, -hh)
		bx, by := tf(hw, -hh)
		cx2, cy2 := tf(hw, hh)
		dx, dy := tf(-hw, hh)
		strokeLine(r, ax, ay, bx, by, penW)
		strokeLine(r, bx, by, cx2, cy2, penW)
		strokeLine(r, cx2, cy2, dx, dy, penW)
		strokeLine(r, dx, dy, ax, ay, penW)

	case preview.LineCapStyleDiamond:
		// C# Diamond: 4 lines at (0,-H/1.4), (-W/1.4,0), (0,H/1.4), (W/1.4,0).
		dw, dh := capW/1.4, capH/1.4
		tx0, ty0 := tf(0, -dh)
		tx1, ty1 := tf(-dw, 0)
		tx2, ty2 := tf(0, dh)
		tx3, ty3 := tf(dw, 0)
		strokeLine(r, tx0, ty0, tx1, ty1, penW)
		strokeLine(r, tx1, ty1, tx2, ty2, penW)
		strokeLine(r, tx2, ty2, tx3, ty3, penW)
		strokeLine(r, tx3, ty3, tx0, ty0, penW)
	}

	// Rasterize the path and composite onto the canvas with anti-aliasing.
	mask := image.NewAlpha(bounds)
	r.Draw(mask, mask.Bounds(), image.Opaque, image.Point{})
	src := &image.Uniform{c}
	draw.DrawMask(e.curPage, bounds, src, image.Point{}, mask, bounds.Min, draw.Over)
}

// strokeLine adds a thick line segment to the vector rasterizer as a filled
// rectangle along the line direction with the given width.
func strokeLine(r *vector.Rasterizer, x0, y0, x1, y1, width float32) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	ln := math.Sqrt(dx*dx + dy*dy)
	if ln < 0.001 {
		return
	}
	// Perpendicular unit vector × half-width
	hw := float64(width) / 2.0
	nx := float32(-dy / ln * hw)
	ny := float32(dx / ln * hw)

	// Four corners of the thick line rectangle
	r.MoveTo(x0+nx, y0+ny)
	r.LineTo(x1+nx, y1+ny)
	r.LineTo(x1-nx, y1-ny)
	r.LineTo(x0-nx, y0-ny)
	r.ClosePath()
}

// strokeEllipse adds an ellipse outline to the vector rasterizer as a filled
// annular ring (outer - inner ellipse).
func strokeEllipse(r *vector.Rasterizer, cx, cy, rx, ry, width float32) {
	hw := width / 2
	const n = 36 // segments for circle approximation
	// Outer ellipse (clockwise)
	for i := 0; i <= n; i++ {
		a := float64(i) * 2 * math.Pi / float64(n)
		px := cx + (rx+hw)*float32(math.Cos(a))
		py := cy + (ry+hw)*float32(math.Sin(a))
		if i == 0 {
			r.MoveTo(px, py)
		} else {
			r.LineTo(px, py)
		}
	}
	r.ClosePath()
	// Inner ellipse (counter-clockwise to cut out)
	for i := n; i >= 0; i-- {
		a := float64(i) * 2 * math.Pi / float64(n)
		px := cx + (rx-hw)*float32(math.Cos(a))
		py := cy + (ry-hw)*float32(math.Sin(a))
		if i == n {
			r.MoveTo(px, py)
		} else {
			r.LineTo(px, py)
		}
	}
	r.ClosePath()
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
// RenderGenericPNG renders any PreparedObject to a PNG byte slice.
// Returns the PNG data and the actual canvas width/height (which may be larger
// than abs(obj.Width/Height) due to line cap expansion).
func RenderGenericPNG(obj preview.PreparedObject) (pngData []byte, canvasW, canvasH int, err error) {
	// C# HTMLExportLayers.GetLayerPicture uses Math.Abs(Width) and Math.Abs(Height)
	// for the bitmap dimensions. Lines can have negative Width/Height encoding
	// direction (e.g., top-right to bottom-left has negative Height).
	w := int(math.Round(math.Abs(float64(obj.Width))))
	h := int(math.Round(math.Abs(float64(obj.Height))))
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	// For lines with caps, compute the exact bounding rect including cap shapes
	// and expand the canvas. This matches C#'s GetConvertedObjects which creates
	// a PictureObject at the extended bounds.
	capPadL, capPadT, capPadR, capPadB := 0, 0, 0, 0
	if obj.Kind == preview.ObjectTypeLine && obj.LineDiagonal &&
		(obj.LineStartCap.Style != preview.LineCapStyleNone || obj.LineEndCap.Style != preview.LineCapStyleNone) {

		bw := 1.0
		if obj.Border.Lines[0] != nil && obj.Border.Lines[0].Width > 0 {
			bw = float64(obj.Border.Lines[0].Width)
		}
		dx, dy := float64(obj.Width), float64(obj.Height)
		angleDeg := math.Abs(math.Atan2(dx, dy) * 180.0 / math.Pi)

		// Line endpoints in canvas coords (before cap expansion).
		var x1c, y1c, x2c, y2c float64
		if obj.Width >= 0 {
			x1c, x2c = 0, float64(w)
		} else {
			x1c, x2c = float64(w), 0
		}
		if obj.Height >= 0 {
			y1c, y2c = 0, float64(h)
		} else {
			y1c, y2c = float64(h), 0
		}

		minX, maxX := 0.0, float64(w)
		minY, maxY := 0.0, float64(h)

		expandCap := func(cap preview.LineCap, cx, cy float64, rotDeg float64) {
			if cap.Style == preview.LineCapStyleNone {
				return
			}
			cw, ch := float64(cap.Width), float64(cap.Height)
			if cw <= 0 {
				cw = 8
			}
			if ch <= 0 {
				ch = 8
			}
			var pts [][2]float64
			switch cap.Style {
			case preview.LineCapStyleArrow:
				pts = [][2]float64{{0, 0}, {-cw, -ch}, {cw, -ch}}
			case preview.LineCapStyleCircle:
				// Circle is rotation-invariant; use 4 extremes of the axis-aligned extent.
				pts = [][2]float64{{-cw / 2, 0}, {cw / 2, 0}, {0, -ch / 2}, {0, ch / 2}}
			case preview.LineCapStyleSquare:
				pts = [][2]float64{{-cw / 2, -ch / 2}, {cw / 2, -ch / 2}, {cw / 2, ch / 2}, {-cw / 2, ch / 2}}
			case preview.LineCapStyleDiamond:
				d := 1.4
				pts = [][2]float64{{0, -ch / d}, {-cw / d, 0}, {0, ch / d}, {cw / d, 0}}
			}
			r := rotDeg * math.Pi / 180.0
			sinR, cosR := math.Sin(r), math.Cos(r)
			for _, p := range pts {
				spx, spy := p[0]*bw, p[1]*bw
				rx := spx*cosR - spy*sinR + cx
				ry := spx*sinR + spy*cosR + cy
				if rx < minX {
					minX = rx
				}
				if rx > maxX {
					maxX = rx
				}
				if ry < minY {
					minY = ry
				}
				if ry > maxY {
					maxY = ry
				}
			}
		}
		expandCap(obj.LineStartCap, x1c, y1c, 180-angleDeg)
		expandCap(obj.LineEndCap, x2c, y2c, -angleDeg)

		// Add border width padding.
		minX -= bw / 2
		minY -= bw / 2
		maxX += bw / 2
		maxY += bw / 2

		capPadL = int(math.Ceil(-minX))
		capPadT = int(math.Ceil(-minY))
		capPadR = int(math.Ceil(maxX - float64(w)))
		capPadB = int(math.Ceil(maxY - float64(h)))
		if capPadL < 0 {
			capPadL = 0
		}
		if capPadT < 0 {
			capPadT = 0
		}
		if capPadR < 0 {
			capPadR = 0
		}
		if capPadB < 0 {
			capPadB = 0
		}
		w += capPadL + capPadR
		h += capPadT + capPadB
	}

	canvas := image.NewRGBA(image.Rect(0, 0, w, h))

	exp := &Exporter{
		ResolutionX: DefaultDPI,
		ResolutionY: DefaultDPI,
		Scale:       1.0,
		curPage:     canvas,
	}
	renderObj := obj
	// Adjust Left/Top for direction + cap padding offset.
	if obj.Width < 0 {
		renderObj.Left = float32(int(math.Round(math.Abs(float64(obj.Width)))) + capPadL)
	} else {
		renderObj.Left = float32(capPadL)
	}
	if obj.Height < 0 {
		renderObj.Top = float32(int(math.Round(math.Abs(float64(obj.Height)))) + capPadT)
	} else {
		renderObj.Top = float32(capPadT)
	}
	exp.renderObject(renderObj, 0)

	var buf bytes.Buffer
	if err2 := png.Encode(&buf, canvas); err2 != nil {
		return nil, 0, 0, err2
	}
	return buf.Bytes(), w, h, nil
}

// RotateImagePNG decodes a PNG image, rotates it by the given angle (90, 180,
// or 270 degrees), resizes to targetW×targetH, and re-encodes as PNG.
// Used by the HTML exporter for rotated picture objects (barcodes, etc.).
func RotateImagePNG(data []byte, angle, targetW, targetH int) []byte {
	// Use image.Decode (not png.Decode) to handle JPEG, GIF, etc. in addition to PNG.
	// C# ref: ImageHelper.RotateImage supports any bitmap format.
	src, _, err := image.Decode(bytes.NewReader(data))
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

// fillPolygon fills the interior of a polygon defined by pts using a scanline
// fill algorithm. This matches C# Graphics.FillPolygon / FillAndDrawPolygon
// used for Diamond and Triangle ShapeObjects (ShapeObject.cs:183-188).
func (e *Exporter) fillPolygon(pts []image.Point, c color.RGBA) {
	if len(pts) < 3 || c.A == 0 {
		return
	}
	bounds := e.curPage.Bounds()

	// Find y extent.
	yMin, yMax := pts[0].Y, pts[0].Y
	for _, p := range pts[1:] {
		if p.Y < yMin {
			yMin = p.Y
		}
		if p.Y > yMax {
			yMax = p.Y
		}
	}
	if yMin < bounds.Min.Y {
		yMin = bounds.Min.Y
	}
	if yMax >= bounds.Max.Y {
		yMax = bounds.Max.Y - 1
	}

	n := len(pts)
	for y := yMin; y <= yMax; y++ {
		// Find x intersections with all polygon edges at scanline y.
		var xs []int
		for i := 0; i < n; i++ {
			j := (i + 1) % n
			y0, y1 := pts[i].Y, pts[j].Y
			x0, x1 := pts[i].X, pts[j].X
			if (y0 <= y && y < y1) || (y1 <= y && y < y0) {
				// Compute x intersection using linear interpolation.
				x := x0 + (y-y0)*(x1-x0)/(y1-y0)
				xs = append(xs, x)
			}
		}
		// Sort intersections.
		for i := 1; i < len(xs); i++ {
			for j := i; j > 0 && xs[j] < xs[j-1]; j-- {
				xs[j], xs[j-1] = xs[j-1], xs[j]
			}
		}
		// Fill between pairs of intersections.
		for i := 0; i+1 < len(xs); i += 2 {
			for x := xs[i]; x <= xs[i+1]; x++ {
				if x >= bounds.Min.X && x < bounds.Max.X {
					e.curPage.SetRGBA(x, y, c)
				}
			}
		}
	}
}

// fillEllipse fills the interior of an ellipse inscribed in the given rect.
// Matches C# Graphics.FillEllipse used for Ellipse ShapeObjects (ShapeObject.cs:173).
func (e *Exporter) fillEllipse(x, y, w, h int, c color.RGBA) {
	if w <= 0 || h <= 0 || c.A == 0 {
		return
	}
	a := float64(w) / 2.0
	b := float64(h) / 2.0
	cx := float64(x) + a
	cy := float64(y) + b
	bounds := e.curPage.Bounds()
	for row := y; row < y+h; row++ {
		if row < bounds.Min.Y || row >= bounds.Max.Y {
			continue
		}
		dy := float64(row) + 0.5 - cy
		if math.Abs(dy) > b {
			continue
		}
		dx := a * math.Sqrt(1-dy*dy/(b*b))
		x0 := int(math.Ceil(cx - dx))
		x1 := int(math.Floor(cx + dx))
		if x0 < bounds.Min.X {
			x0 = bounds.Min.X
		}
		if x1 >= bounds.Max.X {
			x1 = bounds.Max.X - 1
		}
		for px := x0; px <= x1; px++ {
			e.curPage.SetRGBA(px, row, c)
		}
	}
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
