package pdf

import (
	"fmt"
	"io"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// Exporter is a simple PDF export filter.
// It renders each PreparedPage as a PDF page, writing band labels as
// placeholder text elements.  Full text/image/shape rendering is
// intentionally deferred to keep this implementation dependency-free.
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
	curPage *Page
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
	return e.ExportBase.Export(pp, w, e)
}

// ── Exporter interface implementation ─────────────────────────────────────────

// Start initialises the PDF writer and document structure.
func (e *Exporter) Start() error {
	e.writer = NewWriter()

	// Build catalog → pages tree.
	e.pages = NewPages(e.writer)
	_ = NewCatalog(e.writer, e.pages) // registers catalog with writer

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
		case preview.ObjectTypeText:
			if obj.Text == "" {
				continue
			}
			e.renderTextObject(contents, b, obj)
		// Line and Shape rendered as rectangle outlines.
		case preview.ObjectTypeLine, preview.ObjectTypeShape:
			e.renderRectObject(contents, b, obj)
		}
	}
	return nil
}

// renderTextObject writes PDF operators for a TextObject.
func (e *Exporter) renderTextObject(c *Contents, b *preview.PreparedBand, obj preview.PreparedObject) {
	// PDF coordinate system: Y increases upward, origin at bottom-left.
	// Convert pixel positions to PDF points.
	objTopPx := b.Top + obj.Top
	objHeightPx := obj.Height

	xPt := float64(export.PixelsToPoints(obj.Left))
	// Align text to the top of the object box.
	yPt := e.curPage.Height - float64(export.PixelsToPoints(objTopPx+objHeightPx))

	// Choose font name and size.
	fontName := obj.Font.Name
	if fontName == "" {
		fontName = style.DefaultFont().Name
	}
	fontSize := obj.Font.Size
	if fontSize <= 0 {
		fontSize = style.DefaultFont().Size
	}

	// PDF uses /F1 as the standard font reference (Helvetica in the writer).
	// A full implementation would embed the exact font; here we use the built-in.
	c.WriteString(fmt.Sprintf(
		"BT /F1 %.2f Tf %.2f %.2f Td (%s) Tj ET\n",
		fontSize, xPt, yPt, pdfEscape(obj.Text),
	))
}

// renderRectObject draws a filled or stroked rectangle for Line/Shape objects.
func (e *Exporter) renderRectObject(c *Contents, b *preview.PreparedBand, obj preview.PreparedObject) {
	xPt := float64(export.PixelsToPoints(obj.Left))
	yPt := e.curPage.Height - float64(export.PixelsToPoints(b.Top+obj.Top+obj.Height))
	wPt := float64(export.PixelsToPoints(obj.Width))
	hPt := float64(export.PixelsToPoints(obj.Height))

	c.WriteString(fmt.Sprintf(
		"%.2f %.2f %.2f %.2f re S\n", xPt, yPt, wPt, hPt,
	))
}

// ExportPageEnd finalises the current PDF page.
func (e *Exporter) ExportPageEnd(_ *preview.PreparedPage) error {
	if e.curPage != nil {
		e.curPage.Contents().Finalize()
		e.curPage = nil
	}
	return nil
}

// Finish writes the complete PDF document to the output stream.
func (e *Exporter) Finish() error {
	if e.writer == nil {
		return nil
	}
	return e.writer.Write(e.w)
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
