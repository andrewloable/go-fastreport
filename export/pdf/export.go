package pdf

import (
	"fmt"
	"io"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
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
// In this skeleton implementation it writes a simple text marker; a full
// implementation would render text objects, images, borders, etc.
func (e *Exporter) ExportBand(b *preview.PreparedBand) error {
	if e.curPage == nil {
		return nil
	}
	contents := e.curPage.Contents()
	// Approximate Y position in PDF space (PDF Y increases upward).
	pdfY := e.curPage.Height - float64(export.PixelsToPoints(b.Top+b.Height))

	// Write a simple text label for the band.
	// BT = Begin Text; ET = End Text; Tf = set font; Td = move to position; Tj = show text.
	if b.Name != "" {
		contents.WriteString(fmt.Sprintf(
			"BT /F1 8 Tf %.2f %.2f Td (%s) Tj ET\n",
			float64(export.PixelsToPoints(b.Top)), pdfY, pdfEscape(b.Name),
		))
	}
	return nil
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
