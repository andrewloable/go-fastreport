package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"io"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/export/pdf/core"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
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
	curPage *Page
	pp      *preview.PreparedPages // access to blob store for images
	imgIdx  int                    // counter for unique XObject names
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
		case preview.ObjectTypePicture:
			if e.pp != nil && obj.BlobIdx >= 0 {
				e.renderPictureObject(contents, b, obj)
			}
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
