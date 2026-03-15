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
		case preview.ObjectTypePolyLine:
			e.renderPolyPath(contents, b, obj, false)
		case preview.ObjectTypePolygon:
			e.renderPolyPath(contents, b, obj, true)
		case preview.ObjectTypeCheckBox:
			e.renderCheckBoxObject(contents, b, obj)
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
		"q %.3f w 0 0 0 RG %.4f %.4f %.4f %.4f re S Q\n",
		lw, xPt, yPt, wPt, hPt,
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
			"q %.3f w 0 0 0 RG %.4f %.4f m %.4f %.4f l %.4f %.4f l S Q\n",
			tickLW, x1, y1, x2, y2, x3, y3,
		))
	}
}

// renderTextObject writes PDF operators for a TextObject.
// It renders the fill background, border lines, and text.
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
	fontAlias := pdfFontAlias(font.Name, bold, italic)

	// Text color.
	tc := obj.TextColor
	r := float64(tc.R) / 255.0
	g := float64(tc.G) / 255.0
	bl := float64(tc.B) / 255.0

	// Position text near the top of the object (simple single-line placement).
	textY := yPt + hPt - float64(export.PixelsToPoints(font.Size))*1.2
	if textY < yPt {
		textY = yPt
	}

	c.WriteString(fmt.Sprintf(
		"q %.4f %.4f %.4f rg BT /%s %.2f Tf %.2f %.2f Td (%s) Tj ET Q\n",
		r, g, bl, fontAlias, font.Size, xPt, textY, pdfEscape(obj.Text),
	))
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
	var strokeR, strokeG, strokeB float64
	if lc != nil && lc.Color.A > 0 {
		strokeR = float64(lc.Color.R) / 255.0
		strokeG = float64(lc.Color.G) / 255.0
		strokeB = float64(lc.Color.B) / 255.0
	}

	heightPt := float64(export.PixelsToPoints(obj.Height))

	// Build the path.  Points are pixel offsets from the object's top-left corner.
	// PDF uses bottom-left origin, so y must be flipped.
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("q %.4f %.4f %.4f RG ", strokeR, strokeG, strokeB))

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

// pdfFillRect draws a filled rectangle using the given color.
func (e *Exporter) pdfFillRect(c *Contents, x, y, w, h float64, col color.RGBA) {
	r := float64(col.R) / 255.0
	g := float64(col.G) / 255.0
	b := float64(col.B) / 255.0
	c.WriteString(fmt.Sprintf("q %.4f %.4f %.4f rg %.2f %.2f %.2f %.2f re f Q\n",
		r, g, b, x, y, w, h))
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
		r := float64(lc.R) / 255.0
		g := float64(lc.G) / 255.0
		bl := float64(lc.B) / 255.0

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

		c.WriteString(fmt.Sprintf("q %.4f %.4f %.4f RG %.2f w %s%.2f %.2f m %.2f %.2f l S Q\n",
			r, g, bl, lw, dashCmd, s.x1, s.y1, s.x2, s.y2))
	}

	// Drop shadow.
	if b.Shadow && b.ShadowWidth > 0 {
		sw := float64(export.PixelsToPoints(b.ShadowWidth))
		sc := b.ShadowColor
		sr := float64(sc.R) / 255.0
		sg := float64(sc.G) / 255.0
		sb := float64(sc.B) / 255.0
		c.WriteString(fmt.Sprintf("q %.4f %.4f %.4f rg %.2f %.2f %.2f %.2f re f Q\n",
			sr, sg, sb, x+sw, y-sw, w, h))
	}
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

// renderWatermark draws the watermark text (if any) as a centred, semi-transparent
// text string on the PDF page using the cm (current transformation matrix) operator
// to rotate the text diagonally.
func (e *Exporter) renderWatermark(c *Contents, pg *preview.PreparedPage) {
	wm := pg.Watermark
	if wm == nil || !wm.Enabled || wm.Text == "" {
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
	r := float64(col.R) / 255.0
	g := float64(col.G) / 255.0
	b := float64(col.B) / 255.0
	alpha := float64(255-col.A) / 255.0 // A=0 → opaque, A=255 → invisible
	if alpha < 0.1 {
		alpha = 0.3 // default light transparency
	}

	// Write a simple text rendering using PDF operators.
	// We push a graphics state, apply transparency, rotate+centre, draw text, pop.
	c.WriteString(fmt.Sprintf("q\n"))
	// Set fill colour.
	c.WriteString(fmt.Sprintf("%.4f %.4f %.4f rg\n", r, g, b))
	// Apply rotation matrix centred at (cx, cy).
	c.WriteString(fmt.Sprintf("%.6f %.6f %.6f %.6f %.4f %.4f cm\n",
		cos, sin, -sin, cos, cx, cy))
	// Text block.
	c.WriteString(fmt.Sprintf("BT\n"))
	c.WriteString(fmt.Sprintf("/F1 %.2f Tf\n", fontSize))
	c.WriteString(fmt.Sprintf("%.4f Tr\n", alpha*2)) // Tr 0=fill, 2=invisible-ish hack
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
