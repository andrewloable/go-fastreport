package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"io"
	"math"
	"strings"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/export/pdf/core"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
)

// Exporter is a PDF export filter.
// It renders each PreparedPage as a PDF page with text, shapes, and images.
//
// Usage:
//
//	exp := pdf.NewExporter()
//	err := exp.Export(preparedPages, outputWriter)
type Exporter struct {
	export.ExportBase
	w       io.Writer
	writer  *Writer
	pages   *Pages
	catalog *Catalog
	curPage *Page
	pp      *preview.PreparedPages // access to blob store for images
	imgIdx  int                    // counter for unique XObject names
	fontMgr *pdfFontManager        // document-level TrueType font embedding

	// UseCMYK controls color space selection.
	// When true, colors are emitted using DeviceCMYK (k/K operators).
	// When false (default), DeviceRGB (rg/RG operators) is used.
	UseCMYK bool

	// sigFields accumulates /Widget annotation references for the /AcroForm.
	sigFields []*core.IndirectObject
}

// NewExporter creates an Exporter with default settings (all pages).
func NewExporter() *Exporter {
	return &Exporter{
		ExportBase: export.NewExportBase(),
	}
}

// Export writes the PreparedPages as a PDF document to w.
func (e *Exporter) Export(pp *preview.PreparedPages, w io.Writer) error {
	e.w = w
	e.pp = pp
	return e.ExportBase.Export(pp, w, e)
}

// ── Exporter interface implementation ─────────────────────────────────────────

// Start initialises the PDF writer and document structure.
func (e *Exporter) Start() error {
	e.writer = NewWriter()

	// Build catalog → pages tree.
	e.pages = NewPages(e.writer)
	e.catalog = NewCatalog(e.writer, e.pages)

	// Create document-level TrueType font manager.
	e.fontMgr = NewPDFFontManager(e.writer)

	return nil
}

// ExportPageBegin starts a new PDF page.
// Width and height are converted from pixels (96 dpi) to PDF points (72 dpi).
func (e *Exporter) ExportPageBegin(pg *preview.PreparedPage) error {
	widthPt := float64(export.PixelsToPoints(pg.Width))
	heightPt := float64(export.PixelsToPoints(pg.Height))
	if widthPt <= 0 {
		widthPt = 595 // A4 width in points
	}
	if heightPt <= 0 {
		heightPt = 842 // A4 height in points
	}
	e.curPage = NewPage(e.writer, e.pages, widthPt, heightPt)
	return nil
}

// ExportBand renders a single band onto the current PDF page.
// It iterates PreparedBand.Objects and renders text objects as PDF text streams.
func (e *Exporter) ExportBand(b *preview.PreparedBand) error {
	if e.curPage == nil {
		return nil
	}
	contents := e.curPage.Contents()

	for _, obj := range b.Objects {
		switch obj.Kind {
		case preview.ObjectTypeText, preview.ObjectTypeHtml, preview.ObjectTypeRTF:
			if obj.Text == "" {
				continue
			}
			if obj.Kind == preview.ObjectTypeRTF {
				// Strip RTF control words — PDF renders plain text only.
				plain := obj
				plain.Text = utils.StripRTF(obj.Text)
				e.renderTextObject(contents, b, plain)
			} else {
				e.renderTextObject(contents, b, obj)
			}
		// Line and Shape rendered as rectangle outlines.
		case preview.ObjectTypeLine, preview.ObjectTypeShape:
			e.renderRectObject(contents, b, obj)
		case preview.ObjectTypePicture:
			if obj.IsBarcode && obj.BarcodeModules != nil {
				// Render barcode as PDF vector paths (crisp at any zoom).
				e.renderBarcodeVector(contents, b, obj)
			} else if e.pp != nil && obj.BlobIdx >= 0 {
				e.renderPictureObject(contents, b, obj)
			}
		case preview.ObjectTypePolyLine:
			e.renderPolyPath(contents, b, obj, false)
		case preview.ObjectTypePolygon:
			e.renderPolyPath(contents, b, obj, true)
		case preview.ObjectTypeCheckBox:
			e.renderCheckBoxObject(contents, b, obj)
		case preview.ObjectTypeDigitalSignature:
			e.renderDigitalSignatureField(b, obj)
		}

		// Add hyperlink annotation for any object that carries a hyperlink.
		if obj.HyperlinkKind != 0 && obj.HyperlinkValue != "" {
			e.renderHyperlinkAnnotation(b, obj)
		}
	}
	return nil
}

// renderCheckBoxObject draws a checkbox square and, when checked, a tick mark.
// The checkbox is drawn using PDF path operators at the object's position.
func (e *Exporter) renderCheckBoxObject(c *Contents, b *preview.PreparedBand, obj preview.PreparedObject) {
	objTopPx := b.Top + obj.Top
	xPt := float64(export.PixelsToPoints(obj.Left))
	yPt := e.curPage.Height - float64(export.PixelsToPoints(objTopPx+obj.Height))
	wPt := float64(export.PixelsToPoints(obj.Width))
	hPt := float64(export.PixelsToPoints(obj.Height))

	// Default line width.
	lw := 0.75
	if obj.Border.Lines[0].Width > 0 {
		lw = float64(export.PixelsToPoints(obj.Border.Lines[0].Width))
	}

	// Draw checkbox border rectangle.
	c.WriteString(fmt.Sprintf(
		"q %.3f w %s %.4f %.4f %.4f %.4f re S Q\n",
		lw, e.pdfStrokeColorOp(color.RGBA{A: 255}), xPt, yPt, wPt, hPt,
	))

	if obj.Checked {
		// Draw a tick mark (√) as two line segments inside the box.
		// First leg: bottom-left to lower-middle; second leg: lower-middle to top-right.
		margin := wPt * 0.15
		tickLW := lw * 1.5
		x1 := xPt + margin
		y1 := yPt + hPt*0.45
		x2 := xPt + wPt*0.4
		y2 := yPt + margin
		x3 := xPt + wPt - margin
		y3 := yPt + hPt - margin
		c.WriteString(fmt.Sprintf(
			"q %.3f w %s %.4f %.4f m %.4f %.4f l %.4f %.4f l S Q\n",
			tickLW, e.pdfStrokeColorOp(color.RGBA{A: 255}), x1, y1, x2, y2, x3, y3,
		))
	}
}

// renderTextObject writes PDF operators for a TextObject.
// It renders the fill background, border lines, and multi-line wrapped text.
func (e *Exporter) renderTextObject(c *Contents, b *preview.PreparedBand, obj preview.PreparedObject) {
	objTopPx := b.Top + obj.Top
	xPt := float64(export.PixelsToPoints(obj.Left))
	yPt := e.curPage.Height - float64(export.PixelsToPoints(objTopPx+obj.Height))
	wPt := float64(export.PixelsToPoints(obj.Width))
	hPt := float64(export.PixelsToPoints(obj.Height))

	// Fill background.
	if obj.FillColor.A > 0 {
		e.pdfFillRect(c, xPt, yPt, wPt, hPt, obj.FillColor)
	}

	// Border lines.
	e.pdfDrawBorder(c, xPt, yPt, wPt, hPt, &obj.Border)

	if obj.Text == "" {
		return
	}

	// Font selection.
	font := obj.Font
	if font.Name == "" {
		font = style.DefaultFont()
	}
	if font.Size <= 0 {
		font.Size = style.DefaultFont().Size
	}
	bold := font.Style&style.FontStyleBold != 0
	italic := font.Style&style.FontStyleItalic != 0
	var fontAlias string
	if e.fontMgr != nil {
		fontAlias = e.fontMgr.RegisterFont(font.Name, bold, italic)
	} else {
		fontAlias = pdfFontAlias(font.Name, bold, italic)
	}

	// Text color.
	tc := obj.TextColor

	// Line height: 1.2× font size (in points).
	fontPt := float64(font.Size)
	lineHeight := fontPt * 1.2

	// Padding: 2px on each side so text doesn't touch the border.
	const padPx = 2
	padPt := float64(export.PixelsToPoints(padPx))
	innerX := xPt + padPt
	innerW := wPt - 2*padPt
	if innerW <= 0 {
		innerW = wPt
		innerX = xPt
	}

	// Split text into lines, respecting word-wrap and hard newlines.
	measureFn := func(text string) float64 {
		return e.measureText(fontAlias, text, fontPt)
	}
	lines := pdfWrapTextFn(obj.Text, innerW, measureFn, obj.WordWrap)
	if len(lines) == 0 {
		return
	}

	totalTextH := float64(len(lines)) * lineHeight

	// Vertical starting position based on VertAlign.
	// PDF y is bottom-up; yPt is the bottom of the object box.
	var startY float64
	switch obj.VertAlign {
	case 1: // Center
		startY = yPt + (hPt+totalTextH)/2 - lineHeight
	case 2: // Bottom
		startY = yPt + totalTextH - lineHeight
	default: // Top (0)
		startY = yPt + hPt - lineHeight
	}
	// Clamp so first line baseline is inside the box.
	if startY > yPt+hPt-fontPt {
		startY = yPt + hPt - fontPt
	}

	// Write PDF text block.
	// Use Tm (text matrix) for absolute line positioning to avoid Td accumulation.
	c.WriteString(fmt.Sprintf("q %s\n", e.pdfFillColorOp(tc)))
	c.WriteString("BT\n")
	c.WriteString(fmt.Sprintf("/%s %.2f Tf\n", fontAlias, fontPt))

	isEmbedded := strings.HasPrefix(fontAlias, "EF")

	for i, line := range lines {
		lineY := startY - float64(i)*lineHeight
		if lineY < yPt-lineHeight { // below the object bottom — stop rendering
			break
		}

		// Horizontal alignment.
		lineX := innerX
		lineW := e.measureText(fontAlias, line, fontPt)
		var wordSpacing float64
		switch obj.HorzAlign {
		case 1: // Center
			lineX = xPt + (wPt-lineW)/2
		case 2: // Right
			lineX = xPt + wPt - lineW - padPt
		case 3: // Justify — spread words, but not on the last line
			lineX = innerX
			if obj.WordWrap && i < len(lines)-1 {
				wordCount := float64(strings.Count(line, " "))
				if wordCount > 0 {
					ws := (innerW - lineW) / wordCount
					if ws > 0 { // only expand, never compress
						wordSpacing = ws
					}
				}
			}
		}
		if lineX < xPt {
			lineX = xPt
		}

		// Tm sets text line matrix: [1 0 0 1 tx ty] = identity with translation.
		if wordSpacing != 0 {
			c.WriteString(fmt.Sprintf("%.4f Tw\n", wordSpacing))
		}
		if isEmbedded && e.fontMgr != nil {
			hexStr := e.fontMgr.EncodeText(fontAlias, line)
			c.WriteString(fmt.Sprintf("1 0 0 1 %.4f %.4f Tm %s Tj\n", lineX, lineY, hexStr))
		} else {
			c.WriteString(fmt.Sprintf("1 0 0 1 %.4f %.4f Tm (%s) Tj\n", lineX, lineY, pdfEscape(line)))
		}
		if wordSpacing != 0 {
			c.WriteString("0 Tw\n")
		}
	}
	c.WriteString("ET\nQ\n")
}

// pdfWrapText splits text into lines that fit within maxWidth (in points),
// using a rough character-width estimate (fontPt × 0.6 per character).
// Hard newlines (\n) are always honoured. If wordWrap is false, each paragraph
// is emitted as a single line without breaking.
func pdfWrapText(text string, maxWidth, fontPt float64, wordWrap bool) []string {
	return pdfWrapTextFn(text, maxWidth, func(s string) float64 {
		return pdfEstimateTextWidth(s, fontPt)
	}, wordWrap)
}

// pdfWrapTextFn is like pdfWrapText but accepts a custom width measurement
// function, enabling accurate glyph-advance-based measurement for embedded fonts.
func pdfWrapTextFn(text string, maxWidth float64, measureFn func(string) float64, wordWrap bool) []string {
	// Split on hard newlines first.
	paragraphs := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	var lines []string
	for _, para := range paragraphs {
		if para == "" {
			lines = append(lines, "")
			continue
		}
		if !wordWrap {
			lines = append(lines, para)
			continue
		}
		// Word-wrap within this paragraph.
		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		current := words[0]
		for _, w := range words[1:] {
			candidate := current + " " + w
			if measureFn(candidate) <= maxWidth {
				current = candidate
			} else {
				lines = append(lines, current)
				current = w
			}
		}
		lines = append(lines, current)
	}
	return lines
}

// pdfEstimateTextWidth estimates the rendered width of text in points using the
// rough heuristic that the average glyph width is 0.6× the font point size.
func pdfEstimateTextWidth(text string, fontPt float64) float64 {
	return float64(len(text)) * fontPt * 0.6
}

// measureText returns the rendered width of text using either the embedded font
// manager (for EF-prefixed aliases) or the rough heuristic estimator.
func (e *Exporter) measureText(alias, text string, fontPt float64) float64 {
	if strings.HasPrefix(alias, "EF") && e.fontMgr != nil {
		return e.fontMgr.MeasureText(alias, text, fontPt)
	}
	return pdfEstimateTextWidth(text, fontPt)
}

// renderRectObject draws fill and border for Line/Shape objects.
func (e *Exporter) renderRectObject(c *Contents, b *preview.PreparedBand, obj preview.PreparedObject) {
	xPt := float64(export.PixelsToPoints(obj.Left))
	yPt := e.curPage.Height - float64(export.PixelsToPoints(b.Top+obj.Top+obj.Height))
	wPt := float64(export.PixelsToPoints(obj.Width))
	hPt := float64(export.PixelsToPoints(obj.Height))

	// Fill background.
	if obj.FillColor.A > 0 {
		e.pdfFillRect(c, xPt, yPt, wPt, hPt, obj.FillColor)
	}

	// Border.
	e.pdfDrawBorder(c, xPt, yPt, wPt, hPt, &obj.Border)

	// If no border was specified but it's a shape, draw a simple outline.
	if obj.Border.VisibleLines == style.BorderLinesNone {
		c.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f re S\n", xPt, yPt, wPt, hPt))
	}
}

// renderPolyPath draws a polyline or polygon path in PDF.
// If closed is true the path is closed and stroked; otherwise it is just stroked.
func (e *Exporter) renderPolyPath(c *Contents, b *preview.PreparedBand, obj preview.PreparedObject, closed bool) {
	if len(obj.Points) < 2 {
		return
	}

	// Origin of the object in PDF coordinates (bottom-left origin, points).
	originXPt := float64(export.PixelsToPoints(obj.Left))
	originYPt := e.curPage.Height - float64(export.PixelsToPoints(b.Top+obj.Top+obj.Height))

	// Line color from the first border line, defaulting to black.
	lc := obj.Border.Lines[0]
	strokeCol := color.RGBA{A: 255} // black
	if lc != nil && lc.Color.A > 0 {
		strokeCol = lc.Color
	}

	heightPt := float64(export.PixelsToPoints(obj.Height))

	// Build the path.  Points are pixel offsets from the object's top-left corner.
	// PDF uses bottom-left origin, so y must be flipped.
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("q %s ", e.pdfStrokeColorOp(strokeCol)))

	for i, pt := range obj.Points {
		pxPt := originXPt + float64(export.PixelsToPoints(pt[0]))
		// pt[1] is down from top of object; flip for PDF.
		pyPt := originYPt + heightPt - float64(export.PixelsToPoints(pt[1]))
		if i == 0 {
			sb.WriteString(fmt.Sprintf("%.4f %.4f m ", pxPt, pyPt))
		} else {
			sb.WriteString(fmt.Sprintf("%.4f %.4f l ", pxPt, pyPt))
		}
	}

	if closed {
		sb.WriteString("h S Q\n")
	} else {
		sb.WriteString("S Q\n")
	}

	c.WriteString(sb.String())
}

// rgbaToCMYK converts an RGBA color to CMYK components in [0,1].
func rgbaToCMYK(col color.RGBA) (c, m, y, k float64) {
	r := float64(col.R) / 255.0
	g := float64(col.G) / 255.0
	b := float64(col.B) / 255.0
	k = 1.0 - math.Max(r, math.Max(g, b))
	if k == 1.0 {
		return 0, 0, 0, 1
	}
	c = (1.0 - r - k) / (1.0 - k)
	m = (1.0 - g - k) / (1.0 - k)
	y = (1.0 - b - k) / (1.0 - k)
	return
}

// pdfFillColorOp returns a PDF fill color operator string ("rg" or "k") for col.
func (e *Exporter) pdfFillColorOp(col color.RGBA) string {
	if e.UseCMYK {
		c, m, y, k := rgbaToCMYK(col)
		return fmt.Sprintf("%.4f %.4f %.4f %.4f k", c, m, y, k)
	}
	r := float64(col.R) / 255.0
	g := float64(col.G) / 255.0
	b := float64(col.B) / 255.0
	return fmt.Sprintf("%.4f %.4f %.4f rg", r, g, b)
}

// pdfStrokeColorOp returns a PDF stroke color operator string ("RG" or "K") for col.
func (e *Exporter) pdfStrokeColorOp(col color.RGBA) string {
	if e.UseCMYK {
		c, m, y, k := rgbaToCMYK(col)
		return fmt.Sprintf("%.4f %.4f %.4f %.4f K", c, m, y, k)
	}
	r := float64(col.R) / 255.0
	g := float64(col.G) / 255.0
	b := float64(col.B) / 255.0
	return fmt.Sprintf("%.4f %.4f %.4f RG", r, g, b)
}

// pdfFillRect draws a filled rectangle using the given color.
func (e *Exporter) pdfFillRect(c *Contents, x, y, w, h float64, col color.RGBA) {
	c.WriteString(fmt.Sprintf("q %s %.2f %.2f %.2f %.2f re f Q\n",
		e.pdfFillColorOp(col), x, y, w, h))
}

// pdfDrawBorder draws the visible border lines of b around the given rectangle.
func (e *Exporter) pdfDrawBorder(c *Contents, x, y, w, h float64, b *style.Border) {
	if b == nil || b.VisibleLines == style.BorderLinesNone {
		return
	}

	type sideInfo struct {
		flag style.BorderLines
		idx  int
		x1, y1, x2, y2 float64
	}
	sides := []sideInfo{
		{style.BorderLinesLeft, int(style.BorderLeft), x, y, x, y + h},
		{style.BorderLinesRight, int(style.BorderRight), x + w, y, x + w, y + h},
		{style.BorderLinesBottom, int(style.BorderBottom), x, y, x + w, y},
		{style.BorderLinesTop, int(style.BorderTop), x, y + h, x + w, y + h},
	}

	for _, s := range sides {
		if b.VisibleLines&s.flag == 0 {
			continue
		}
		line := b.Lines[s.idx]
		if line == nil {
			line = style.NewBorderLine()
		}
		lc := line.Color
		lw := float64(export.PixelsToPoints(line.Width))
		if lw < 0.5 {
			lw = 0.5
		}

		var dashCmd string
		switch line.Style {
		case style.LineStyleDash:
			dashCmd = "[4 2] 0 d "
		case style.LineStyleDot:
			dashCmd = "[1 2] 0 d "
		case style.LineStyleDashDot:
			dashCmd = "[4 2 1 2] 0 d "
		default:
			dashCmd = "[] 0 d " // solid
		}

		c.WriteString(fmt.Sprintf("q %s %.2f w %s%.2f %.2f m %.2f %.2f l S Q\n",
			e.pdfStrokeColorOp(lc), lw, dashCmd, s.x1, s.y1, s.x2, s.y2))
	}

	// Drop shadow.
	if b.Shadow && b.ShadowWidth > 0 {
		sw := float64(export.PixelsToPoints(b.ShadowWidth))
		sc := b.ShadowColor
		c.WriteString(fmt.Sprintf("q %s %.2f %.2f %.2f %.2f re f Q\n",
			e.pdfFillColorOp(sc), x+sw, y-sw, w, h))
	}
}

// renderHyperlinkAnnotation creates a PDF /Annot dictionary with /Subtype /Link
// for the given object and appends it to the current page's /Annots array.
// Supports URL (external) and PageNumber (internal GoTo) hyperlinks.
func (e *Exporter) renderHyperlinkAnnotation(b *preview.PreparedBand, obj preview.PreparedObject) {
	if e.curPage == nil {
		return
	}
	objTopPx := b.Top + obj.Top
	xPt := float64(export.PixelsToPoints(obj.Left))
	yPt := e.curPage.Height - float64(export.PixelsToPoints(objTopPx+obj.Height))
	wPt := float64(export.PixelsToPoints(obj.Width))
	hPt := float64(export.PixelsToPoints(obj.Height))

	// /Rect [llx lly urx ury] in PDF coordinates (bottom-left origin).
	rect := core.NewArray(
		core.NewFloat(xPt),
		core.NewFloat(yPt),
		core.NewFloat(xPt+wPt),
		core.NewFloat(yPt+hPt),
	)

	annotDict := core.NewDictionary()
	annotDict.Add("Type", core.NewName("Annot"))
	annotDict.Add("Subtype", core.NewName("Link"))
	annotDict.Add("Rect", rect)
	annotDict.Add("Border", core.NewArray(core.NewInt(0), core.NewInt(0), core.NewInt(0))) // no border

	const (
		hyperlinkURL        = 1
		hyperlinkPageNumber = 2
		hyperlinkBookmark   = 3
	)

	switch obj.HyperlinkKind {
	case hyperlinkURL:
		// External URI action.
		actionDict := core.NewDictionary()
		actionDict.Add("Type", core.NewName("Action"))
		actionDict.Add("S", core.NewName("URI"))
		actionDict.Add("URI", core.NewString(obj.HyperlinkValue))
		annotDict.Add("A", actionDict)

	case hyperlinkPageNumber:
		// Internal GoTo action — navigate to the Nth page.
		pages := e.pages.PageList()
		pageNo := 0
		for _, c := range obj.HyperlinkValue {
			if c >= '0' && c <= '9' {
				pageNo = pageNo*10 + int(c-'0')
			}
		}
		pageNo-- // convert 1-based to 0-based
		if pageNo >= 0 && pageNo < len(pages) {
			dest := core.NewArray(
				core.NewRef(pages[pageNo].Obj()),
				core.NewName("XYZ"),
				core.NewInt(0),
				core.NewFloat(pages[pageNo].Height),
				core.NewInt(0),
			)
			actionDict := core.NewDictionary()
			actionDict.Add("Type", core.NewName("Action"))
			actionDict.Add("S", core.NewName("GoTo"))
			actionDict.Add("D", dest)
			annotDict.Add("A", actionDict)
		}

	case hyperlinkBookmark:
		// GoTo named destination.
		actionDict := core.NewDictionary()
		actionDict.Add("Type", core.NewName("Action"))
		actionDict.Add("S", core.NewName("GoTo"))
		actionDict.Add("D", core.NewHexString(obj.HyperlinkValue))
		annotDict.Add("A", actionDict)
	}

	e.curPage.AddAnnotation(annotDict)
}

// renderDigitalSignatureField creates a PDF /Widget annotation for a digital
// signature field and registers it with the page /Annots array and the AcroForm
// /Fields array.  The field has /Subtype /Sig and an empty /V (unsigned).
// A minimal appearance stream draws a dashed border around the field area.
func (e *Exporter) renderDigitalSignatureField(b *preview.PreparedBand, obj preview.PreparedObject) {
	if e.curPage == nil || e.writer == nil {
		return
	}

	// Convert field rectangle to PDF coordinates (points, bottom-left origin).
	xPt := float64(export.PixelsToPoints(obj.Left))
	yPt := e.curPage.Height - float64(export.PixelsToPoints(b.Top+obj.Top+obj.Height))
	wPt := float64(export.PixelsToPoints(obj.Width))
	hPt := float64(export.PixelsToPoints(obj.Height))

	// Build the /Widget annotation dictionary.
	rect := core.NewArray(
		core.NewFloat(xPt), core.NewFloat(yPt),
		core.NewFloat(xPt+wPt), core.NewFloat(yPt+hPt),
	)

	// Build a minimal appearance stream that draws a dashed border.
	apContent := fmt.Sprintf(
		"q [2 2] 0 d 0.5 w %.4f %.4f %.4f %.4f re S Q\n",
		0.0, 0.0, wPt, hPt,
	)
	apStream := core.NewStream()
	apStream.Data = []byte(apContent)
	apStreamObj := e.writer.NewObject(apStream)

	// Build appearance /AP dict.
	apDict := core.NewDictionary()
	apDict.Add("N", core.NewRef(apStreamObj))

	// Signature value dict (empty = unsigned field).
	sigValDict := core.NewDictionary()
	sigValDict.Add("Type", core.NewName("Sig"))
	sigValDict.Add("Filter", core.NewName("Adobe.PPKLite"))
	sigValDict.Add("SubFilter", core.NewName("adbe.pkcs7.detached"))

	// The field name: use object name or default.
	fieldName := obj.Name
	if fieldName == "" {
		fieldName = "Signature"
	}

	widgetDict := core.NewDictionary()
	widgetDict.Add("Type", core.NewName("Annot"))
	widgetDict.Add("Subtype", core.NewName("Widget"))
	widgetDict.Add("FT", core.NewName("Sig"))
	widgetDict.Add("Rect", rect)
	widgetDict.Add("T", core.NewHexString(fieldName))
	widgetDict.Add("F", core.NewInt(4)) // Print flag
	widgetDict.Add("AP", apDict)
	widgetDict.Add("V", sigValDict)

	widgetObj := e.writer.NewObject(widgetDict)
	e.curPage.AddAnnotation(widgetDict)
	e.sigFields = append(e.sigFields, widgetObj)
}

// finalizeAcroForm builds the /AcroForm dictionary and registers it in the
// catalog when there are signature fields.  Called from Finish().
func (e *Exporter) finalizeAcroForm() {
	if len(e.sigFields) == 0 || e.catalog == nil {
		return
	}
	fields := core.NewArray()
	for _, f := range e.sigFields {
		fields.Add(core.NewRef(f))
	}
	acroForm := core.NewDictionary()
	acroForm.Add("Fields", fields)
	acroForm.Add("SigFlags", core.NewInt(3)) // 3 = AppendOnly | SignaturesExist
	e.catalog.SetAcroForm(acroForm)
}

// pdfFontAlias selects the pre-registered PDF font alias based on the font
// family name and bold/italic flags. Maps common Windows names to PDF Type1 fonts.
func pdfFontAlias(name string, bold, italic bool) string {
	lower := strings.ToLower(name)
	var family int // 0=Helvetica, 1=Times, 2=Courier
	if strings.Contains(lower, "times") || strings.Contains(lower, "garamond") ||
		strings.Contains(lower, "palatino") || strings.Contains(lower, "georgia") {
		family = 1
	} else if strings.Contains(lower, "courier") || strings.Contains(lower, "consolas") ||
		strings.Contains(lower, "monaco") || strings.Contains(lower, "monospace") {
		family = 2
	}
	switch family {
	case 1: // Times
		switch {
		case bold && italic:
			return "F8"
		case bold:
			return "F6"
		case italic:
			return "F7"
		default:
			return "F5"
		}
	case 2: // Courier
		switch {
		case bold && italic:
			return "F12"
		case bold:
			return "F10"
		case italic:
			return "F11"
		default:
			return "F9"
		}
	default: // Helvetica (Arial and others)
		switch {
		case bold && italic:
			return "F4"
		case bold:
			return "F2"
		case italic:
			return "F3"
		default:
			return "F1"
		}
	}
}

// ExportPageEnd finalises the current PDF page, rendering any watermark.
func (e *Exporter) ExportPageEnd(pg *preview.PreparedPage) error {
	if e.curPage != nil {
		if pg != nil && pg.Watermark != nil && pg.Watermark.Enabled {
			e.renderWatermark(e.curPage.Contents(), pg)
		}
		e.curPage.Contents().Finalize()
		e.curPage = nil
	}
	return nil
}

// renderWatermark draws the watermark (text and/or image) on the PDF page.
// It respects ShowTextOnTop / ShowImageOnTop ordering: items not "on top"
// are emitted first (behind content), items "on top" after (this function is
// called twice from ExportPageEnd — once before content via ExportPageBegin
// and once after via ExportPageEnd using the onTop flag).
func (e *Exporter) renderWatermark(c *Contents, pg *preview.PreparedPage) {
	wm := pg.Watermark
	if wm == nil || !wm.Enabled {
		return
	}
	// Render image watermark (behind text in Z order).
	if wm.ImageBlobIdx >= 0 && e.pp != nil {
		e.renderWatermarkImage(c, pg, wm)
	}
	// Render text watermark on top of image watermark.
	if wm.Text != "" {
		e.renderWatermarkText(c, pg, wm)
	}
}

// renderWatermarkImage embeds a blob-stored image as a full-page (or sized) watermark.
func (e *Exporter) renderWatermarkImage(c *Contents, pg *preview.PreparedPage, wm *preview.PreparedWatermark) {
	data := e.pp.BlobStore.Get(wm.ImageBlobIdx)
	if len(data) == 0 {
		return
	}

	widthPt := float64(export.PixelsToPoints(pg.Width))
	heightPt := float64(export.PixelsToPoints(pg.Height))

	// Decode image to obtain dimensions.
	imgCfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return
	}
	imgW := float64(imgCfg.Width)
	imgH := float64(imgCfg.Height)

	// Build PDF image stream (same as renderPictureObject).
	imgStream := core.NewStream()
	imgStream.Dict.Add("Type", core.NewName("XObject"))
	imgStream.Dict.Add("Subtype", core.NewName("Image"))

	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		imgStream.Compressed = false
		imgStream.Dict.Add("Filter", core.NewName("DCTDecode"))
		imgStream.Data = data
	} else {
		img, _, decErr := image.Decode(bytes.NewReader(data))
		if decErr != nil {
			return
		}
		bounds := img.Bounds()
		rgba := image.NewNRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
		rgb := make([]byte, 0, bounds.Dx()*bounds.Dy()*3)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				px := rgba.NRGBAAt(x, y)
				rgb = append(rgb, px.R, px.G, px.B)
			}
		}
		imgStream.Compressed = true
		imgStream.Data = rgb
	}
	imgStream.Dict.Add("Width", core.NewInt(int(imgW)))
	imgStream.Dict.Add("Height", core.NewInt(int(imgH)))
	imgStream.Dict.Add("ColorSpace", core.NewName("DeviceRGB"))
	imgStream.Dict.Add("BitsPerComponent", core.NewInt(8))

	xObjRef := e.writer.NewObject(imgStream)
	imgName := fmt.Sprintf("WmIm%d", e.imgIdx)
	e.imgIdx++
	e.curPage.AddXObject(imgName, xObjRef)

	// Transparency.
	opacity := 1.0 - float64(wm.ImageTransparency)
	if opacity < 0 {
		opacity = 0
	}
	gsDict := core.NewDictionary()
	gsDict.Add("Type", core.NewName("ExtGState"))
	gsDict.Add("ca", core.NewFloat(opacity))
	gsDict.Add("CA", core.NewFloat(opacity))
	gsName := fmt.Sprintf("WmGs%d", e.imgIdx)
	e.curPage.AddExtGState(gsName, gsDict)

	// Compute destination rectangle in PDF points (origin bottom-left).
	var dstX, dstY, dstW, dstH float64
	imgPtW := imgW * 72.0 / 96.0 // pixels → points (assuming 96 dpi)
	imgPtH := imgH * 72.0 / 96.0

	switch wm.ImageSize {
	case preview.WatermarkImageSizeStretch:
		dstX, dstY, dstW, dstH = 0, 0, widthPt, heightPt
	case preview.WatermarkImageSizeCenter:
		dstW, dstH = imgPtW, imgPtH
		dstX = (widthPt - dstW) / 2
		dstY = (heightPt - dstH) / 2
	case preview.WatermarkImageSizeZoom:
		scaleX := widthPt / imgPtW
		scaleY := heightPt / imgPtH
		scale := math.Min(scaleX, scaleY)
		dstW, dstH = imgPtW*scale, imgPtH*scale
		dstX = (widthPt - dstW) / 2
		dstY = (heightPt - dstH) / 2
	default: // Normal
		dstW, dstH = imgPtW, imgPtH
		dstX, dstY = 0, heightPt-dstH // top-left
	}

	c.WriteString(fmt.Sprintf("q /%s gs\n", gsName))
	if wm.ImageSize == preview.WatermarkImageSizeTile {
		// Tile: repeat across the page.
		for y := heightPt - imgPtH; y > -imgPtH; y -= imgPtH {
			for x := 0.0; x < widthPt; x += imgPtW {
				c.WriteString(fmt.Sprintf("q %.4f 0 0 %.4f %.4f %.4f cm /%s Do Q\n",
					imgPtW, imgPtH, x, y, imgName))
			}
		}
	} else {
		c.WriteString(fmt.Sprintf("%.4f 0 0 %.4f %.4f %.4f cm /%s Do\n",
			dstW, dstH, dstX, dstY, imgName))
	}
	c.WriteString("Q\n")
}

// renderWatermarkText draws the watermark text (if any) as a centred, semi-transparent
// text string on the PDF page using the cm (current transformation matrix) operator
// to rotate the text diagonally.
func (e *Exporter) renderWatermarkText(c *Contents, pg *preview.PreparedPage, wm *preview.PreparedWatermark) {
	if wm.Text == "" {
		return
	}
	widthPt := float64(export.PixelsToPoints(pg.Width))
	heightPt := float64(export.PixelsToPoints(pg.Height))

	// Rotation angle in radians.
	var angleDeg float64
	switch wm.TextRotation {
	case preview.WatermarkTextRotationVertical:
		angleDeg = 90
	case preview.WatermarkTextRotationForwardDiagonal:
		angleDeg = 45
	case preview.WatermarkTextRotationBackwardDiagonal:
		angleDeg = -45
	default: // Horizontal
		angleDeg = 0
	}
	angleRad := angleDeg * 3.14159265358979 / 180.0
	cos := math.Cos(angleRad)
	sin := math.Sin(angleRad)

	// Font size in points — use the watermark font size directly (already in pt).
	fontSize := float64(wm.Font.Size)
	if fontSize <= 0 {
		fontSize = 60
	}

	// Centre of the page.
	cx := widthPt / 2
	cy := heightPt / 2

	// Encode the watermark text as a PDF string literal (basic ASCII escape).
	escaped := pdfEscape(wm.Text)

	col := wm.TextColor
	// alpha: col.A = 0 → fully opaque; col.A = 255 → invisible.
	// Convert to PDF /ca fill opacity (0.0 = transparent, 1.0 = opaque).
	fillOpacity := 1.0 - float64(col.A)/255.0
	if fillOpacity > 0.9 {
		fillOpacity = 0.3 // default light transparency when no alpha specified
	}

	// Register an ExtGState with the fill opacity (/ca) for this watermark.
	gsDict := core.NewDictionary()
	gsDict.Add("Type", core.NewName("ExtGState"))
	gsDict.Add("ca", core.NewFloat(fillOpacity)) // fill opacity
	gsDict.Add("CA", core.NewFloat(fillOpacity)) // stroke opacity
	e.curPage.AddExtGState("WM1", gsDict)

	// Write PDF operators: rotate+centre transformation, apply ExtGState, draw text.
	c.WriteString("q\n")
	// Set fill colour.
	c.WriteString(fmt.Sprintf("%s\n", e.pdfFillColorOp(col)))
	// Apply rotation matrix centred at (cx, cy).
	c.WriteString(fmt.Sprintf("%.6f %.6f %.6f %.6f %.4f %.4f cm\n",
		cos, sin, -sin, cos, cx, cy))
	// Apply graphics state with proper transparency.
	c.WriteString("/WM1 gs\n")
	// Text block.
	c.WriteString("BT\n")
	c.WriteString(fmt.Sprintf("/F1 %.2f Tf\n", fontSize))
	c.WriteString("0 Tr\n") // Tr 0 = fill (normal rendering mode)
	// Centre the text horizontally (approximate: half the font size * num chars).
	approxW := fontSize * float64(len(wm.Text)) * 0.5
	c.WriteString(fmt.Sprintf("%.4f %.4f Td\n", -approxW/2, -fontSize/2))
	c.WriteString(fmt.Sprintf("(%s) Tj\n", escaped))
	c.WriteString("ET\n")
	c.WriteString("Q\n")
}

// Finish writes the complete PDF document to the output stream.
func (e *Exporter) Finish() error {
	if e.writer == nil {
		return nil
	}
	if e.catalog != nil && e.pp != nil {
		e.writeOutlines()
		e.writeNamedDests()
	}
	e.finalizeAcroForm()

	// Finalize embedded font objects and register them in every page's /Font dict.
	if e.fontMgr != nil {
		e.fontMgr.Finalize()
		for _, page := range e.pages.PageList() {
			e.fontMgr.AddToPage(page.FontDict())
		}
	}

	return e.writer.Write(e.w)
}

// writeOutlines builds the PDF outline tree from PreparedPages.Outline and
// registers it in the catalog under /Outlines.
func (e *Exporter) writeOutlines() {
	outline := e.pp.Outline
	if outline == nil || len(outline.Root.Children) == 0 {
		return
	}
	pages := e.pages.PageList()

	// Recursive builder — returns the outline item indirect object.
	// We forward-declare buildItem so it can recurse.
	type outlineRef struct {
		obj *core.IndirectObject
	}
	refs := make([]outlineRef, 0)

	var buildItem func(item *preview.OutlineItem, parentRef *core.IndirectObject) *core.IndirectObject
	buildItem = func(item *preview.OutlineItem, parentRef *core.IndirectObject) *core.IndirectObject {
		dict := core.NewDictionary()
		dict.Add("Title", core.NewHexString(item.Text))
		dict.Add("Parent", core.NewRef(parentRef))

		// /Dest [pageRef /XYZ left top zoom]
		if item.PageIdx >= 0 && item.PageIdx < len(pages) {
			pg := pages[item.PageIdx]
			// Y in PDF is bottom-up; convert OffsetY (top-down pixels) to points.
			yPt := pg.Height - float64(export.PixelsToPoints(item.OffsetY))
			dest := core.NewArray(
				core.NewRef(pg.Obj()),
				core.NewName("XYZ"),
				core.NewInt(0),
				core.NewFloat(yPt),
				core.NewInt(0),
			)
			dict.Add("Dest", dest)
		}

		obj := e.writer.NewObject(dict)
		refs = append(refs, outlineRef{obj: obj})

		if len(item.Children) > 0 {
			childObjs := make([]*core.IndirectObject, len(item.Children))
			for i, child := range item.Children {
				childObjs[i] = buildItem(child, obj)
			}
			dict.Add("First", core.NewRef(childObjs[0]))
			dict.Add("Last", core.NewRef(childObjs[len(childObjs)-1]))
			dict.Add("Count", core.NewInt(-len(childObjs))) // negative = collapsed
			// Link siblings.
			for i := range childObjs {
				if i+1 < len(childObjs) {
					// We need to add Next to child dict — but child dicts were
					// already registered. We can add at any time before Write().
					if d, ok := childObjs[i].Value.(*core.Dictionary); ok {
						d.Add("Next", core.NewRef(childObjs[i+1]))
					}
				}
				if i > 0 {
					if d, ok := childObjs[i].Value.(*core.Dictionary); ok {
						d.Add("Prev", core.NewRef(childObjs[i-1]))
					}
				}
			}
		}
		return obj
	}

	// Build root outline dictionary.
	rootDict := core.NewDictionary()
	rootDict.Add("Type", core.NewName("Outlines"))
	rootObj := e.writer.NewObject(rootDict)

	topChildren := outline.Root.Children
	topObjs := make([]*core.IndirectObject, len(topChildren))
	for i, child := range topChildren {
		topObjs[i] = buildItem(child, rootObj)
	}
	if len(topObjs) > 0 {
		rootDict.Add("First", core.NewRef(topObjs[0]))
		rootDict.Add("Last", core.NewRef(topObjs[len(topObjs)-1]))
		rootDict.Add("Count", core.NewInt(len(refs)))
		// Link top-level siblings.
		for i := range topObjs {
			if i+1 < len(topObjs) {
				if d, ok := topObjs[i].Value.(*core.Dictionary); ok {
					d.Add("Next", core.NewRef(topObjs[i+1]))
				}
			}
			if i > 0 {
				if d, ok := topObjs[i].Value.(*core.Dictionary); ok {
					d.Add("Prev", core.NewRef(topObjs[i-1]))
				}
			}
		}
	}

	e.catalog.SetOutlines(rootObj)
}

// writeNamedDests builds the /Names /Dests name tree from PreparedPages.Bookmarks
// and registers it in the catalog.
func (e *Exporter) writeNamedDests() {
	bk := e.pp.Bookmarks
	if bk == nil || bk.Count() == 0 {
		return
	}
	pages := e.pages.PageList()

	// Build a flat /Names array: [(name1) dest1 (name2) dest2 ...]
	namesArr := core.NewArray()
	for _, bookmark := range bk.All() {
		if bookmark.PageIdx < 0 || bookmark.PageIdx >= len(pages) {
			continue
		}
		pg := pages[bookmark.PageIdx]
		yPt := pg.Height - float64(export.PixelsToPoints(bookmark.OffsetY))
		dest := core.NewArray(
			core.NewRef(pg.Obj()),
			core.NewName("XYZ"),
			core.NewInt(0),
			core.NewFloat(yPt),
			core.NewInt(0),
		)
		namesArr.Add(core.NewHexString(bookmark.Name))
		namesArr.Add(dest)
	}

	destsDict := core.NewDictionary()
	destsDict.Add("Names", namesArr)
	destsObj := e.writer.NewObject(destsDict)
	e.catalog.SetNamedDests(destsObj)
}

// renderPictureObject embeds an image from the blob store as a PDF XObject.
// It supports JPEG (embedded directly with /DCTDecode) and any format
// decodable by the Go standard library (converted to raw RGB + /FlateDecode).
func (e *Exporter) renderPictureObject(c *Contents, b *preview.PreparedBand, obj preview.PreparedObject) {
	data := e.pp.BlobStore.Get(obj.BlobIdx)
	if len(data) == 0 {
		return
	}

	// Build the PDF image stream.
	imgStream := core.NewStream()
	imgStream.Dict.Add("Type", core.NewName("XObject"))
	imgStream.Dict.Add("Subtype", core.NewName("Image"))

	var imgW, imgH int

	// Detect JPEG by magic bytes (0xFF 0xD8).
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		// JPEG: embed raw bytes with DCTDecode filter (no re-compression).
		imgStream.Compressed = false
		imgStream.Dict.Add("Filter", core.NewName("DCTDecode"))
		imgStream.Data = data

		// Decode just to get dimensions.
		cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return
		}
		imgW, imgH = cfg.Width, cfg.Height
	} else {
		// Generic: decode via Go's image package, emit raw 8-bit RGB + FlateDecode.
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return
		}
		bounds := img.Bounds()
		imgW, imgH = bounds.Dx(), bounds.Dy()

		// Convert to NRGBA then extract RGB bytes.
		rgba := image.NewNRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
		rgb := make([]byte, 0, imgW*imgH*3)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c4 := rgba.NRGBAAt(x, y)
				rgb = append(rgb, c4.R, c4.G, c4.B)
			}
		}
		imgStream.Compressed = true // FlateDecode via Stream.WriteTo
		imgStream.Data = rgb
	}

	imgStream.Dict.Add("Width", core.NewInt(imgW))
	imgStream.Dict.Add("Height", core.NewInt(imgH))
	imgStream.Dict.Add("ColorSpace", core.NewName("DeviceRGB"))
	imgStream.Dict.Add("BitsPerComponent", core.NewInt(8))

	// Register as an indirect object and add to page resources.
	xObjRef := e.writer.NewObject(imgStream)
	imgName := fmt.Sprintf("Im%d", e.imgIdx)
	e.imgIdx++
	e.curPage.AddXObject(imgName, xObjRef)

	// Emit content stream operators to place the image.
	// PDF coordinate origin is bottom-left; Y increases upward.
	xPt := float64(export.PixelsToPoints(obj.Left))
	yPt := e.curPage.Height - float64(export.PixelsToPoints(b.Top+obj.Top+obj.Height))
	wPt := float64(export.PixelsToPoints(obj.Width))
	hPt := float64(export.PixelsToPoints(obj.Height))

	c.WriteString(fmt.Sprintf(
		"q %.4f 0 0 %.4f %.4f %.4f cm /%s Do Q\n",
		wPt, hPt, xPt, yPt, imgName,
	))
}

// ── helpers ───────────────────────────────────────────────────────────────────

// pdfEscape replaces characters that are special inside a PDF literal string.
func pdfEscape(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '(', ')', '\\':
			out = append(out, '\\', c)
		default:
			if c >= 0x20 && c <= 0x7E {
				out = append(out, c)
			}
		}
	}
	return string(out)
}

// renderBarcodeVector renders a barcode as PDF vector paths (filled rectangles),
// one rectangle per dark module. This produces crisp output at any zoom level,
// unlike raster image embedding which can appear blurry when zoomed.
//
// The module matrix from PreparedObject.BarcodeModules is scaled to fit the
// object's rendered bounds.
func (e *Exporter) renderBarcodeVector(c *Contents, b *preview.PreparedBand, obj preview.PreparedObject) {
	modules := obj.BarcodeModules
	if len(modules) == 0 || len(modules[0]) == 0 {
		return
	}
	rows := len(modules)
	cols := len(modules[0])

	// Convert object bounds from pixels to PDF points (bottom-up coordinates).
	objTopPx := b.Top + obj.Top
	xPt := float64(export.PixelsToPoints(obj.Left))
	yPt := e.curPage.Height - float64(export.PixelsToPoints(objTopPx+obj.Height))
	wPt := float64(export.PixelsToPoints(obj.Width))
	hPt := float64(export.PixelsToPoints(obj.Height))
	if wPt <= 0 || hPt <= 0 {
		return
	}

	// Module size in points.
	modW := wPt / float64(cols)
	modH := hPt / float64(rows)

	// Use the barcode color (default black).
	bc := obj.TextColor
	if bc.A == 0 {
		bc.A = 255 // default opaque black
	}

	c.WriteString(fmt.Sprintf("q %s\n", e.pdfFillColorOp(bc)))
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			if !modules[row][col] {
				continue
			}
			// PDF y is bottom-up: row 0 is at the top of the barcode bounding box.
			mx := xPt + float64(col)*modW
			my := yPt + float64(rows-1-row)*modH
			// Draw filled rectangle for this dark module.
			c.WriteString(fmt.Sprintf("%.4f %.4f %.4f %.4f re f\n", mx, my, modW, modH))
		}
	}
	c.WriteString("Q\n")

}
