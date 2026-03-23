// Package svg implements an SVG export filter for go-fastreport.
// It renders prepared pages as SVG documents suitable for web embedding
// and infinite-zoom display.
//
// Multi-page reports are emitted as multiple top-level <svg> elements wrapped
// in an HTML-compatible fragment, or as individual SVG files when accessed
// one page at a time via ExportPage.
package svg

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"io"
	"strings"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
)

// Exporter produces SVG output from a PreparedPages collection.
//
// Each page is rendered as a separate top-level <svg> element. When a report
// contains multiple pages, the elements are concatenated in the output stream
// separated by a newline. Consumers that need a single-page SVG should set
// PageRange / PageNumbers on the embedded ExportBase before calling Export.
//
// Coordinates are pixel-native (96 dpi internal units). SVG user units default
// to pixels, so no coordinate conversion is required.
type Exporter struct {
	export.ExportBase

	// EmbedFonts controls whether a <style> block is included with @font-face
	// declarations. Requires base64-encoded font data to be available; currently
	// defaults to false (fonts are referenced by family name only).
	EmbedFonts bool

	// Title is used as the <title> element inside each <svg>.
	Title string

	w       io.Writer
	pp      *preview.PreparedPages
	sb      strings.Builder
	pageIdx int
}

// NewExporter creates an Exporter with sensible defaults.
func NewExporter() *Exporter {
	return &Exporter{
		ExportBase: export.NewExportBase(),
		Title:      "Report",
	}
}

// Export writes the PreparedPages as SVG output to w.
func (e *Exporter) Export(pp *preview.PreparedPages, w io.Writer) error {
	e.w = w
	e.pp = pp
	return e.ExportBase.Export(pp, w, e)
}

// FileExtension returns the recommended file extension for SVG output.
func (e *Exporter) FileExtension() string { return ".svg" }

// Name returns the human-readable format name.
func (e *Exporter) Name() string { return "SVG" }

// ── Exporter interface ─────────────────────────────────────────────────────────

// Start resets internal state before the export run begins.
func (e *Exporter) Start() error {
	e.sb.Reset()
	e.pageIdx = 0
	return nil
}

// ExportPageBegin opens a new <svg> element for the page.
func (e *Exporter) ExportPageBegin(pg *preview.PreparedPage) error {
	w := pg.Width
	h := pg.Height
	if w <= 0 {
		w = 794 // A4 at 96 dpi
	}
	if h <= 0 {
		h = 1123
	}

	e.pageIdx++
	if e.pageIdx > 1 {
		e.sb.WriteByte('\n')
	}

	e.sb.WriteString(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"`+
			` width="%.2f" height="%.2f" viewBox="0 0 %.2f %.2f">`+"\n",
		w, h, w, h,
	))

	// Optional document title.
	if e.Title != "" {
		e.sb.WriteString(fmt.Sprintf("  <title>%s</title>\n", xmlEscape(e.Title)))
	}

	// Emit SVG marker defs for any line caps used on this page.
	// Pre-scan bands to determine which cap styles are needed.
	// Mirrors C# LineObject.GetConvertedObjects() cap-to-bitmap logic (LineObject.OpenSource.cs).
	e.emitLineCapDefs(pg)

	// White page background.
	e.sb.WriteString(fmt.Sprintf(
		`  <rect x="0" y="0" width="%.2f" height="%.2f" fill="#FFFFFF"/>`, w, h,
	))
	e.sb.WriteByte('\n')

	// Watermark behind page content.
	if wm := pg.Watermark; wm != nil && wm.Enabled {
		if !wm.ShowImageOnTop {
			e.renderWatermarkImage(wm, w, h)
		}
		if !wm.ShowTextOnTop {
			e.renderWatermarkText(wm, w, h)
		}
	}

	// Open a group for the page content.
	e.sb.WriteString(fmt.Sprintf(`  <g id="page%d">`+"\n", e.pageIdx))
	return nil
}

// emitLineCapDefs pre-scans pg for line objects with non-None caps and emits
// a <defs> block containing the necessary SVG <marker> elements.
// Each marker ID is scoped to the page index to avoid cross-page conflicts.
// C# equivalent: LineObject.GetConvertedObjects() renders caps via GDI+;
// Go uses SVG markers instead (LineObject.OpenSource.cs).
func (e *Exporter) emitLineCapDefs(pg *preview.PreparedPage) {
	needed := make(map[preview.LineCapStyle]bool)
	for _, b := range pg.Bands {
		for _, obj := range b.Objects {
			if obj.Kind != preview.ObjectTypeLine {
				continue
			}
			if obj.LineStartCap.Style != preview.LineCapStyleNone {
				needed[obj.LineStartCap.Style] = true
			}
			if obj.LineEndCap.Style != preview.LineCapStyleNone {
				needed[obj.LineEndCap.Style] = true
			}
		}
	}
	if len(needed) == 0 {
		return
	}
	e.sb.WriteString("  <defs>\n")
	for style := range needed {
		e.sb.WriteString(e.svgCapMarkerDef(style))
	}
	e.sb.WriteString("  </defs>\n")
}

// svgCapMarkerDef returns the SVG <marker> definition for a given cap style.
// Each marker is sized relative to stroke width via markerUnits="strokeWidth".
func (e *Exporter) svgCapMarkerDef(s preview.LineCapStyle) string {
	id := e.capMarkerID(s)
	switch s {
	case preview.LineCapStyleArrow:
		// Arrow: right-pointing filled triangle, auto-oriented.
		return fmt.Sprintf(
			`    <marker id="%s" markerWidth="6" markerHeight="6" refX="6" refY="3" orient="auto" markerUnits="strokeWidth">`+"\n"+
				`      <path d="M0,0 L6,3 L0,6 Z" fill="context-stroke"/>`+"\n"+
				`    </marker>`+"\n", id)
	case preview.LineCapStyleCircle:
		return fmt.Sprintf(
			`    <marker id="%s" markerWidth="4" markerHeight="4" refX="2" refY="2" orient="auto" markerUnits="strokeWidth">`+"\n"+
				`      <circle cx="2" cy="2" r="2" fill="context-stroke"/>`+"\n"+
				`    </marker>`+"\n", id)
	case preview.LineCapStyleSquare:
		return fmt.Sprintf(
			`    <marker id="%s" markerWidth="4" markerHeight="4" refX="2" refY="2" orient="auto" markerUnits="strokeWidth">`+"\n"+
				`      <rect x="0" y="0" width="4" height="4" fill="context-stroke"/>`+"\n"+
				`    </marker>`+"\n", id)
	case preview.LineCapStyleDiamond:
		return fmt.Sprintf(
			`    <marker id="%s" markerWidth="4" markerHeight="4" refX="2" refY="2" orient="auto" markerUnits="strokeWidth">`+"\n"+
				`      <path d="M2,0 L4,2 L2,4 L0,2 Z" fill="context-stroke"/>`+"\n"+
				`    </marker>`+"\n", id)
	default:
		return ""
	}
}

// capMarkerID returns a stable SVG marker element ID for the given cap style.
// Scoped to the page index to avoid cross-page conflicts in multi-page SVG output.
func (e *Exporter) capMarkerID(s preview.LineCapStyle) string {
	var name string
	switch s {
	case preview.LineCapStyleArrow:
		name = "arrow"
	case preview.LineCapStyleCircle:
		name = "circle"
	case preview.LineCapStyleSquare:
		name = "square"
	case preview.LineCapStyleDiamond:
		name = "diamond"
	default:
		name = "none"
	}
	return fmt.Sprintf("cap-p%d-%s", e.pageIdx, name)
}

// ExportBand renders all objects in a band onto the current SVG page.
func (e *Exporter) ExportBand(b *preview.PreparedBand) error {
	bandTop := b.Top

	// Render each object, adjusting Top to page-absolute coordinates.
	for _, obj := range b.Objects {
		abs := obj
		abs.Top += bandTop
		e.renderObject(abs)
	}
	return nil
}

// ExportPageEnd closes the page group and the <svg> element.
func (e *Exporter) ExportPageEnd(pg *preview.PreparedPage) error {
	// Watermark on top of page content.
	if wm := pg.Watermark; wm != nil && wm.Enabled {
		if wm.ShowImageOnTop {
			e.renderWatermarkImage(wm, pg.Width, pg.Height)
		}
		if wm.ShowTextOnTop {
			e.renderWatermarkText(wm, pg.Width, pg.Height)
		}
	}

	e.sb.WriteString("  </g>\n")
	e.sb.WriteString("</svg>\n")
	return nil
}

// Finish writes the accumulated SVG string to the output writer.
func (e *Exporter) Finish() error {
	_, err := io.WriteString(e.w, e.sb.String())
	return err
}

// ── Object rendering ───────────────────────────────────────────────────────────

// renderObject emits an SVG element for a single PreparedObject.
// obj.Top must already be adjusted to page-absolute coordinates.
func (e *Exporter) renderObject(obj preview.PreparedObject) {
	x := obj.Left
	y := obj.Top
	w := obj.Width
	h := obj.Height
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}

	// Handle RTF: strip control words to render as plain text.
	if obj.Kind == preview.ObjectTypeRTF {
		plain := obj
		plain.Text = utils.StripRTF(obj.Text)
		plain.Kind = preview.ObjectTypeText
		obj = plain
	}

	switch obj.Kind {
	case preview.ObjectTypeText, preview.ObjectTypeHtml:
		e.renderText(obj, x, y, w, h)

	case preview.ObjectTypeLine:
		e.renderLine(obj, x, y, w, h)

	case preview.ObjectTypeShape:
		e.renderShape(obj, x, y, w, h)

	case preview.ObjectTypePicture:
		e.renderPicture(obj, x, y, w, h)

	case preview.ObjectTypeCheckBox:
		e.renderCheckBox(obj, x, y, w, h)

	case preview.ObjectTypePolyLine, preview.ObjectTypePolygon:
		e.renderPolyShape(obj, x, y, w, h)

	case preview.ObjectTypeSVG:
		e.renderSVGObject(obj, x, y, w, h)

	case preview.ObjectTypeDigitalSignature:
		e.renderDigitalSignature(obj, x, y, w, h)

	default:
		// Unknown type: emit an empty transparent rectangle as placeholder.
		e.sb.WriteString(fmt.Sprintf(
			`    <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="none"/>`+"\n",
			x, y, w, h,
		))
	}
}

// renderText emits a <foreignObject> containing an HTML div for text objects.
// This supports word-wrap, vertical/horizontal alignment and basic font styling.
func (e *Exporter) renderText(obj preview.PreparedObject, x, y, w, h float32) {
	// Background fill.
	if obj.FillColor.A > 0 {
		e.sb.WriteString(fmt.Sprintf(
			`    <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s" fill-opacity="%.3f"/>`,
			x, y, w, h,
			rgbHex(obj.FillColor),
			alphaF(obj.FillColor.A),
		))
		e.sb.WriteByte('\n')
	}

	// Border lines.
	e.renderBorderLines(&obj.Border, x, y, w, h)

	text := obj.Text
	if obj.Kind == preview.ObjectTypeHtml {
		// Strip HTML tags for plain SVG text output.
		text = stripHTMLTags(text)
	}
	if text == "" {
		return
	}

	// Font styling.
	font := obj.Font
	fontName := font.Name
	if fontName == "" {
		fontName = "Arial"
	}
	fontSize := font.Size
	if fontSize <= 0 {
		fontSize = 10
	}
	// Convert pt → px for SVG (96dpi internal, font size in points → px = pt * 96/72).
	fontPx := fontSize * 96.0 / 72.0

	tc := obj.TextColor
	if tc.A == 0 {
		tc = color.RGBA{A: 255}
	}

	// Build CSS style for the foreignObject div.
	var css strings.Builder
	css.WriteString(fmt.Sprintf("font-family:'%s';font-size:%.2fpx;", fontName, fontPx))
	css.WriteString(fmt.Sprintf("color:%s;", rgbaCSS(tc)))
	if font.Style&style.FontStyleBold != 0 {
		css.WriteString("font-weight:bold;")
	}
	if font.Style&style.FontStyleItalic != 0 {
		css.WriteString("font-style:italic;")
	}
	if font.Style&style.FontStyleUnderline != 0 {
		css.WriteString("text-decoration:underline;")
	}
	switch obj.HorzAlign {
	case 1:
		css.WriteString("text-align:center;")
	case 2:
		css.WriteString("text-align:right;")
	case 3:
		css.WriteString("text-align:justify;")
	default:
		css.WriteString("text-align:left;")
	}
	if obj.WordWrap {
		css.WriteString("word-wrap:break-word;white-space:normal;")
	} else {
		css.WriteString("white-space:nowrap;overflow:hidden;")
	}
	// Vertical alignment via flexbox.
	switch obj.VertAlign {
	case 1:
		css.WriteString("display:flex;align-items:center;")
	case 2:
		css.WriteString("display:flex;align-items:flex-end;")
	default:
		css.WriteString("display:block;")
	}
	css.WriteString("width:100%;height:100%;overflow:hidden;box-sizing:border-box;")

	innerText := xmlEscape(text)
	if obj.HyperlinkKind == 1 && obj.HyperlinkValue != "" {
		innerText = fmt.Sprintf(`<a xmlns="http://www.w3.org/1999/xhtml" href="%s" style="color:inherit;text-decoration:inherit;">%s</a>`,
			xmlEscape(obj.HyperlinkValue), innerText)
	}

	e.sb.WriteString(fmt.Sprintf(
		`    <foreignObject x="%.2f" y="%.2f" width="%.2f" height="%.2f">`+"\n",
		x, y, w, h,
	))
	e.sb.WriteString(fmt.Sprintf(
		`      <div xmlns="http://www.w3.org/1999/xhtml" style="%s">%s</div>`+"\n",
		css.String(), innerText,
	))
	e.sb.WriteString("    </foreignObject>\n")
}

// renderLine emits a <line> SVG element, optionally with marker-start/marker-end
// for line-cap decorations (Arrow, Circle, Square, Diamond).
// Mirrors C# LineObject rendering with cap support (LineObject.OpenSource.cs).
func (e *Exporter) renderLine(obj preview.PreparedObject, x, y, w, h float32) {
	lineColor := color.RGBA{A: 255}
	lineWidth := float32(1)
	if obj.Border.Lines[0] != nil {
		if obj.Border.Lines[0].Color.A > 0 {
			lineColor = obj.Border.Lines[0].Color
		}
		if obj.Border.Lines[0].Width > 0 {
			lineWidth = obj.Border.Lines[0].Width
		}
	}

	stroke := rgbHex(lineColor)
	strokeOpacity := alphaF(lineColor.A)

	// Build optional marker attributes for start/end caps.
	var markerAttrs string
	if obj.LineStartCap.Style != preview.LineCapStyleNone {
		markerAttrs += fmt.Sprintf(` marker-start="url(#%s)"`, e.capMarkerID(obj.LineStartCap.Style))
	}
	if obj.LineEndCap.Style != preview.LineCapStyleNone {
		markerAttrs += fmt.Sprintf(` marker-end="url(#%s)"`, e.capMarkerID(obj.LineEndCap.Style))
	}

	if obj.LineDiagonal {
		e.sb.WriteString(fmt.Sprintf(
			`    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-opacity="%.3f" stroke-width="%.2f"%s/>`,
			x, y, x+w, y+h, stroke, strokeOpacity, lineWidth, markerAttrs,
		))
	} else {
		// Horizontal or vertical based on dominant dimension.
		if w >= h {
			// Horizontal line.
			midY := y + h/2
			e.sb.WriteString(fmt.Sprintf(
				`    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-opacity="%.3f" stroke-width="%.2f"%s/>`,
				x, midY, x+w, midY, stroke, strokeOpacity, lineWidth, markerAttrs,
			))
		} else {
			// Vertical line.
			midX := x + w/2
			e.sb.WriteString(fmt.Sprintf(
				`    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-opacity="%.3f" stroke-width="%.2f"%s/>`,
				midX, y, midX, y+h, stroke, strokeOpacity, lineWidth, markerAttrs,
			))
		}
	}
	e.sb.WriteByte('\n')
}

// renderShape emits SVG elements for ShapeObject types.
func (e *Exporter) renderShape(obj preview.PreparedObject, x, y, w, h float32) {
	shapeColor := color.RGBA{A: 255}
	shapeWidth := float32(1)
	if obj.Border.Lines[0] != nil {
		if obj.Border.Lines[0].Color.A > 0 {
			shapeColor = obj.Border.Lines[0].Color
		}
		if obj.Border.Lines[0].Width > 0 {
			shapeWidth = obj.Border.Lines[0].Width
		}
	}

	stroke := rgbHex(shapeColor)
	strokeOpacity := alphaF(shapeColor.A)

	fillAttr := `fill="none"`
	if obj.FillColor.A > 0 {
		fillAttr = fmt.Sprintf(`fill="%s" fill-opacity="%.3f"`, rgbHex(obj.FillColor), alphaF(obj.FillColor.A))
	}

	switch obj.ShapeKind {
	case 2: // Ellipse
		rx := w / 2
		ry := h / 2
		cx := x + rx
		cy := y + ry
		e.sb.WriteString(fmt.Sprintf(
			`    <ellipse cx="%.2f" cy="%.2f" rx="%.2f" ry="%.2f" %s stroke="%s" stroke-opacity="%.3f" stroke-width="%.2f"/>`,
			cx, cy, rx, ry, fillAttr, stroke, strokeOpacity, shapeWidth,
		))

	case 1: // RoundRectangle
		rx := obj.ShapeCurve
		if rx <= 0 {
			rx = 0
		}
		e.sb.WriteString(fmt.Sprintf(
			`    <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="%.2f" ry="%.2f" %s stroke="%s" stroke-opacity="%.3f" stroke-width="%.2f"/>`,
			x, y, w, h, rx, rx, fillAttr, stroke, strokeOpacity, shapeWidth,
		))

	case 3: // Triangle
		pts := fmt.Sprintf("%.2f,%.2f %.2f,%.2f %.2f,%.2f",
			x+w/2, y,
			x, y+h,
			x+w, y+h,
		)
		e.sb.WriteString(fmt.Sprintf(
			`    <polygon points="%s" %s stroke="%s" stroke-opacity="%.3f" stroke-width="%.2f"/>`,
			pts, fillAttr, stroke, strokeOpacity, shapeWidth,
		))

	case 4: // Diamond
		pts := fmt.Sprintf("%.2f,%.2f %.2f,%.2f %.2f,%.2f %.2f,%.2f",
			x+w/2, y,
			x+w, y+h/2,
			x+w/2, y+h,
			x, y+h/2,
		)
		e.sb.WriteString(fmt.Sprintf(
			`    <polygon points="%s" %s stroke="%s" stroke-opacity="%.3f" stroke-width="%.2f"/>`,
			pts, fillAttr, stroke, strokeOpacity, shapeWidth,
		))

	default: // Rectangle (0)
		e.sb.WriteString(fmt.Sprintf(
			`    <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" %s stroke="%s" stroke-opacity="%.3f" stroke-width="%.2f"/>`,
			x, y, w, h, fillAttr, stroke, strokeOpacity, shapeWidth,
		))
	}
	e.sb.WriteByte('\n')
}

// renderPicture emits an <image> element with a base64 data URI.
func (e *Exporter) renderPicture(obj preview.PreparedObject, x, y, w, h float32) {
	if obj.BlobIdx >= 0 && e.pp != nil {
		if data := e.pp.BlobStore.Get(obj.BlobIdx); len(data) > 0 {
			mime := imageMIME(data)
			b64 := base64.StdEncoding.EncodeToString(data)
			e.sb.WriteString(fmt.Sprintf(
				`    <image x="%.2f" y="%.2f" width="%.2f" height="%.2f"`+
					` href="data:%s;base64,%s" preserveAspectRatio="xMidYMid meet"/>`,
				x, y, w, h, mime, b64,
			))
			e.sb.WriteByte('\n')
			return
		}
	}
	// No image data: emit an empty placeholder rect.
	e.sb.WriteString(fmt.Sprintf(
		`    <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="#F0F0F0" stroke="#CCCCCC" stroke-width="1"/>`,
		x, y, w, h,
	))
	e.sb.WriteByte('\n')
}

// renderCheckBox emits a <rect> with an optional cross for a checkbox.
func (e *Exporter) renderCheckBox(obj preview.PreparedObject, x, y, w, h float32) {
	e.sb.WriteString(fmt.Sprintf(
		`    <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="#FFFFFF" stroke="#000000" stroke-width="1"/>`,
		x, y, w, h,
	))
	e.sb.WriteByte('\n')
	if obj.Text == "true" || obj.Checked {
		pad := float32(2)
		e.sb.WriteString(fmt.Sprintf(
			`    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#000000" stroke-width="1"/>`,
			x+pad, y+pad, x+w-pad, y+h-pad,
		))
		e.sb.WriteByte('\n')
		e.sb.WriteString(fmt.Sprintf(
			`    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#000000" stroke-width="1"/>`,
			x+w-pad, y+pad, x+pad, y+h-pad,
		))
		e.sb.WriteByte('\n')
	}
}

// renderPolyShape emits a <polyline> or <polygon> SVG element.
func (e *Exporter) renderPolyShape(obj preview.PreparedObject, x, y, w, h float32) {
	if len(obj.Points) < 2 {
		return
	}

	lineColor := color.RGBA{A: 255}
	lineWidth := float32(1)
	if obj.Border.Lines[0] != nil {
		if obj.Border.Lines[0].Color.A > 0 {
			lineColor = obj.Border.Lines[0].Color
		}
		if obj.Border.Lines[0].Width > 0 {
			lineWidth = obj.Border.Lines[0].Width
		}
	}

	var pts strings.Builder
	for i, pt := range obj.Points {
		if i > 0 {
			pts.WriteByte(' ')
		}
		pts.WriteString(fmt.Sprintf("%.2f,%.2f", x+pt[0], y+pt[1]))
	}

	stroke := rgbHex(lineColor)
	strokeOpacity := alphaF(lineColor.A)

	if obj.Kind == preview.ObjectTypePolyLine {
		e.sb.WriteString(fmt.Sprintf(
			`    <polyline points="%s" fill="none" stroke="%s" stroke-opacity="%.3f" stroke-width="%.2f"/>`,
			pts.String(), stroke, strokeOpacity, lineWidth,
		))
	} else {
		// Polygon — may have fill.
		fillAttr := `fill="none"`
		if obj.FillColor.A > 0 {
			fillAttr = fmt.Sprintf(`fill="%s" fill-opacity="%.3f"`, rgbHex(obj.FillColor), alphaF(obj.FillColor.A))
		}
		e.sb.WriteString(fmt.Sprintf(
			`    <polygon points="%s" %s stroke="%s" stroke-opacity="%.3f" stroke-width="%.2f"/>`,
			pts.String(), fillAttr, stroke, strokeOpacity, lineWidth,
		))
	}
	e.sb.WriteByte('\n')

	// Suppress unused parameter warnings — w and h are for context only.
	_ = w
	_ = h
}

// renderSVGObject embeds raw SVG content inside a <g> element.
func (e *Exporter) renderSVGObject(obj preview.PreparedObject, x, y, w, h float32) {
	if obj.BlobIdx >= 0 && e.pp != nil {
		if svgData := e.pp.BlobStore.Get(obj.BlobIdx); len(svgData) > 0 {
			e.sb.WriteString(fmt.Sprintf(
				`    <g transform="translate(%.2f,%.2f)">`, x, y,
			))
			e.sb.WriteByte('\n')
			e.sb.WriteString("    ")
			e.sb.Write(svgData)
			e.sb.WriteByte('\n')
			e.sb.WriteString("    </g>\n")
			return
		}
	}
	// No SVG data: empty placeholder.
	e.sb.WriteString(fmt.Sprintf(
		`    <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="#F0F0F0" stroke="#CCCCCC" stroke-dasharray="4,2"/>`,
		x, y, w, h,
	))
	e.sb.WriteByte('\n')
}

// renderDigitalSignature emits a dashed-border placeholder box.
func (e *Exporter) renderDigitalSignature(obj preview.PreparedObject, x, y, w, h float32) {
	e.sb.WriteString(fmt.Sprintf(
		`    <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f"`+
			` fill="#FFFFFF" stroke="#888888" stroke-width="2" stroke-dasharray="6,3"/>`,
		x, y, w, h,
	))
	e.sb.WriteByte('\n')

	label := obj.Text
	if label == "" {
		label = "Digital Signature"
	}
	fontSize := obj.Font.Size
	if fontSize <= 0 {
		fontSize = 10
	}
	fontPx := fontSize * 96.0 / 72.0
	fontName := obj.Font.Name
	if fontName == "" {
		fontName = "Arial"
	}
	tc := obj.TextColor
	if tc.A == 0 {
		tc = color.RGBA{R: 136, G: 136, B: 136, A: 255}
	}
	e.sb.WriteString(fmt.Sprintf(
		`    <text x="%.2f" y="%.2f" text-anchor="middle" dominant-baseline="middle"`+
			` font-family="%s" font-size="%.2f" fill="%s">%s</text>`,
		x+w/2, y+h/2, xmlEscape(fontName), fontPx, rgbHex(tc), xmlEscape(label),
	))
	e.sb.WriteByte('\n')
}

// renderBorderLines emits individual SVG <line> elements for each visible border side.
func (e *Exporter) renderBorderLines(b *style.Border, x, y, w, h float32) {
	if b == nil || b.VisibleLines == style.BorderLinesNone {
		return
	}

	type side struct {
		flag style.BorderLines
		idx  style.BorderSide
		x1   float32
		y1   float32
		x2   float32
		y2   float32
	}
	sides := []side{
		{style.BorderLinesTop, style.BorderTop, x, y, x + w, y},
		{style.BorderLinesBottom, style.BorderBottom, x, y + h, x + w, y + h},
		{style.BorderLinesLeft, style.BorderLeft, x, y, x, y + h},
		{style.BorderLinesRight, style.BorderRight, x + w, y, x + w, y + h},
	}

	for _, s := range sides {
		if b.VisibleLines&s.flag == 0 {
			continue
		}
		var line *style.BorderLine
		if b.Lines[s.idx] != nil {
			line = b.Lines[s.idx]
		}
		lc := color.RGBA{A: 255}
		lw := float32(1)
		dashArray := ""
		if line != nil {
			if line.Color.A > 0 {
				lc = line.Color
			}
			if line.Width > 0 {
				lw = line.Width
			}
			switch line.Style {
			case style.LineStyleDash:
				dashArray = fmt.Sprintf(` stroke-dasharray="%.2f,%.2f"`, lw*4, lw*2)
			case style.LineStyleDot:
				dashArray = fmt.Sprintf(` stroke-dasharray="%.2f,%.2f"`, lw, lw*2)
			case style.LineStyleDashDot:
				dashArray = fmt.Sprintf(` stroke-dasharray="%.2f,%.2f,%.2f,%.2f"`, lw*4, lw*2, lw, lw*2)
			case style.LineStyleDashDotDot:
				dashArray = fmt.Sprintf(` stroke-dasharray="%.2f,%.2f,%.2f,%.2f,%.2f,%.2f"`, lw*4, lw*2, lw, lw*2, lw, lw*2)
			}
		}
		e.sb.WriteString(fmt.Sprintf(
			`    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-opacity="%.3f" stroke-width="%.2f"%s/>`,
			s.x1, s.y1, s.x2, s.y2, rgbHex(lc), alphaF(lc.A), lw, dashArray,
		))
		e.sb.WriteByte('\n')
	}
}

// ── Watermark helpers ──────────────────────────────────────────────────────────

// renderWatermarkText emits a centered, optionally rotated <text> element for watermarks.
func (e *Exporter) renderWatermarkText(wm *preview.PreparedWatermark, pageW, pageH float32) {
	if wm.Text == "" {
		return
	}
	fontSize := wm.Font.Size
	if fontSize <= 0 {
		fontSize = 48
	}
	fontPx := fontSize * 96.0 / 72.0
	fontName := wm.Font.Name
	if fontName == "" {
		fontName = "Arial"
	}
	tc := wm.TextColor
	fill := rgbHex(tc)
	fillOpacity := alphaF(tc.A)

	var rotateDeg float32
	switch wm.TextRotation {
	case preview.WatermarkTextRotationVertical:
		rotateDeg = 90
	case preview.WatermarkTextRotationForwardDiagonal:
		rotateDeg = -45
	case preview.WatermarkTextRotationBackwardDiagonal:
		rotateDeg = 45
	default:
		rotateDeg = 0
	}

	cx := pageW / 2
	cy := pageH / 2

	transformAttr := ""
	if rotateDeg != 0 {
		transformAttr = fmt.Sprintf(` transform="rotate(%.0f %.2f %.2f)"`, rotateDeg, cx, cy)
	}

	e.sb.WriteString(fmt.Sprintf(
		`  <text x="%.2f" y="%.2f" text-anchor="middle" dominant-baseline="middle"`+
			` font-family="%s" font-size="%.2f" fill="%s" fill-opacity="%.3f"`+
			` pointer-events="none" user-select="none"%s>%s</text>`,
		cx, cy, xmlEscape(fontName), fontPx, fill, fillOpacity, transformAttr, xmlEscape(wm.Text),
	))
	e.sb.WriteByte('\n')
}

// renderWatermarkImage emits an <image> element for a watermark image.
func (e *Exporter) renderWatermarkImage(wm *preview.PreparedWatermark, pageW, pageH float32) {
	if wm.ImageBlobIdx < 0 || e.pp == nil {
		return
	}
	imgData := e.pp.BlobStore.Get(wm.ImageBlobIdx)
	if len(imgData) == 0 {
		return
	}
	opacity := float32(1.0) - wm.ImageTransparency
	if opacity < 0 {
		opacity = 0
	}
	b64 := base64.StdEncoding.EncodeToString(imgData)

	var preserveAR string
	switch wm.ImageSize {
	case preview.WatermarkImageSizeStretch:
		preserveAR = "none"
	case preview.WatermarkImageSizeZoom:
		preserveAR = "xMidYMid meet"
	default:
		preserveAR = "xMidYMid meet"
	}

	e.sb.WriteString(fmt.Sprintf(
		`  <image x="0" y="0" width="%.2f" height="%.2f" href="data:image/png;base64,%s"`+
			` preserveAspectRatio="%s" opacity="%.3f" pointer-events="none"/>`,
		pageW, pageH, b64, preserveAR, opacity,
	))
	e.sb.WriteByte('\n')
}

// ── Helper utilities ───────────────────────────────────────────────────────────

// rgbHex returns a CSS hex colour string like "#RRGGBB" for the given RGBA.
func rgbHex(c color.RGBA) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// alphaF returns the alpha channel as a float in [0,1].
func alphaF(a uint8) float64 {
	return float64(a) / 255.0
}

// rgbaCSS returns a CSS rgba() color string.
func rgbaCSS(c color.RGBA) string {
	return fmt.Sprintf("rgba(%d,%d,%d,%.3f)", c.R, c.G, c.B, float64(c.A)/255.0)
}

// xmlEscape escapes a string for safe inclusion in SVG text content and attributes.
func xmlEscape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&apos;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// imageMIME detects the MIME type from image magic bytes.
// Falls back to "image/png" for unknown formats.
func imageMIME(data []byte) string {
	if len(data) >= 3 {
		switch {
		case data[0] == 0xFF && data[1] == 0xD8:
			return "image/jpeg"
		case data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46:
			return "image/gif"
		case len(data) >= 4 && data[0] == 0x42 && data[1] == 0x4D:
			return "image/bmp"
		case len(data) >= 4 && data[0] == 0x49 && data[1] == 0x49 && data[2] == 0x2A:
			return "image/tiff"
		case len(data) >= 4 && data[0] == 0x4D && data[1] == 0x4D && data[2] == 0x00:
			return "image/tiff"
		}
	}
	if len(data) > 4 {
		s := string(data[:min(len(data), 64)])
		if strings.Contains(s, "<svg") || strings.Contains(s, "<?xml") {
			return "image/svg+xml"
		}
	}
	return "image/png"
}

// stripHTMLTags removes HTML tags from a string for plain-text SVG rendering.
func stripHTMLTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// min returns the smaller of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
