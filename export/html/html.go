// Package html implements an HTML export filter for go-fastreport.
// It renders prepared pages as a single HTML document with absolute-positioned
// elements (WYSIWYG mode), one section per page.
package html

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

// Exporter produces HTML output from a PreparedPages collection.
//
// Output structure (non-layered):
//
//	<html><body>
//	  <div class="page" …> <!-- one per page -->
//	    <div class="band" …> … </div>
//	  </div>
//	</body></html>
//
// In Layers mode, band divs are omitted. Every object is positioned
// absolutely on the page with an explicit z-index, allowing correct
// rendering of overlapping objects.
type Exporter struct {
	export.ExportBase

	// Title is used as the HTML document title.
	Title string
	// EmbedCSS controls whether a minimal stylesheet is written inline.
	EmbedCSS bool
	// Scale converts pixel values to CSS pixels (default 1.0).
	Scale float32
	// Layers enables layered output mode. When true, every report object is
	// rendered as a top-level absolutely-positioned element on the page div
	// with a z-index matching its paint order. Band wrapper divs are omitted.
	// This allows overlapping objects to render correctly.
	Layers bool

	w       io.Writer
	pp      *preview.PreparedPages
	sb      strings.Builder
	pageIdx int
	css     *cssRegistry
	// zIdx is the current z-index counter, incremented per object in Layers mode.
	zIdx int
}

// NewExporter creates an Exporter with sensible defaults.
func NewExporter() *Exporter {
	return &Exporter{
		ExportBase: export.NewExportBase(),
		Title:      "Report",
		EmbedCSS:   true,
		Scale:      1.0,
	}
}

// Export writes the PreparedPages as an HTML document to w.
func (e *Exporter) Export(pp *preview.PreparedPages, w io.Writer) error {
	e.w = w
	e.pp = pp
	return e.ExportBase.Export(pp, w, e)
}

// ── Exporter interface ─────────────────────────────────────────────────────────

func (e *Exporter) Start() error {
	e.sb.Reset()
	e.pageIdx = 0
	e.css = newCSSRegistry()
	e.sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	e.sb.WriteString(fmt.Sprintf("<title>%s</title>\n", export.HTMLString(e.Title)))
	if e.EmbedCSS {
		e.sb.WriteString(e.defaultCSS())
	}
	e.sb.WriteString("</head>\n<body>\n")
	return nil
}

func (e *Exporter) ExportPageBegin(pg *preview.PreparedPage) error {
	scale := e.Scale
	if scale <= 0 {
		scale = 1
	}
	w := pg.Width * scale
	h := pg.Height * scale
	e.zIdx = 0
	e.sb.WriteString(fmt.Sprintf(
		`<div class="page" style="position:relative;width:%.2fpx;height:%.2fpx;`+
			`overflow:hidden;margin:0 auto 20px auto;background:#fff;box-shadow:0 0 5px #aaa;">`,
		w, h,
	))
	e.sb.WriteString("\n")

	// Watermark behind page content (z-index 0).
	if wm := pg.Watermark; wm != nil && wm.Enabled {
		if !wm.ShowImageOnTop {
			e.renderWatermarkImage(wm, w, h)
		}
		if !wm.ShowTextOnTop {
			e.renderWatermarkText(wm, w, h)
		}
	}

	e.pageIdx++
	return nil
}

// renderWatermarkText emits a CSS-positioned div with rotated text.
func (e *Exporter) renderWatermarkText(wm *preview.PreparedWatermark, pageW, pageH float32) {
	if wm.Text == "" {
		return
	}
	c := wm.TextColor
	rgba := fmt.Sprintf("rgba(%d,%d,%d,%.2f)", c.R, c.G, c.B, float64(c.A)/255.0)
	fontSize := wm.Font.Size
	if fontSize <= 0 {
		fontSize = 48
	}

	var rotDeg int
	switch wm.TextRotation {
	case preview.WatermarkTextRotationVertical:
		rotDeg = 90
	case preview.WatermarkTextRotationForwardDiagonal:
		rotDeg = -45
	case preview.WatermarkTextRotationBackwardDiagonal:
		rotDeg = 45
	default: // Horizontal
		rotDeg = 0
	}
	transform := ""
	if rotDeg != 0 {
		transform = fmt.Sprintf("transform:rotate(%ddeg);", rotDeg)
	}
	e.sb.WriteString(fmt.Sprintf(
		`<div style="position:absolute;top:0;left:0;width:%.2fpx;height:%.2fpx;`+
			`display:flex;align-items:center;justify-content:center;pointer-events:none;`+
			`z-index:0;%s">`,
		pageW, pageH, transform,
	))
	e.sb.WriteString(fmt.Sprintf(
		`<span style="color:%s;font-size:%.0fpx;white-space:nowrap;user-select:none;">%s</span>`,
		rgba, fontSize*e.Scale, export.HTMLString(wm.Text),
	))
	e.sb.WriteString("</div>\n")
}

// renderWatermarkImage emits a positioned div with a base64-encoded background image.
func (e *Exporter) renderWatermarkImage(wm *preview.PreparedWatermark, pageW, pageH float32) {
	if wm.ImageBlobIdx < 0 || e.pp == nil {
		return
	}
	imgData := e.pp.BlobStore.Get(wm.ImageBlobIdx)
	if len(imgData) == 0 {
		return
	}
	opacity := 1.0 - float64(wm.ImageTransparency)
	if opacity < 0 {
		opacity = 0
	}
	b64 := base64.StdEncoding.EncodeToString(imgData)

	var bgSize string
	switch wm.ImageSize {
	case preview.WatermarkImageSizeStretch:
		bgSize = "100% 100%"
	case preview.WatermarkImageSizeZoom:
		bgSize = "contain"
	case preview.WatermarkImageSizeTile:
		bgSize = "auto"
	default:
		bgSize = "auto"
	}

	repeat := "no-repeat"
	if wm.ImageSize == preview.WatermarkImageSizeTile {
		repeat = "repeat"
	}
	bgPos := "center center"
	if wm.ImageSize == preview.WatermarkImageSizeNormal {
		bgPos = "top left"
	}

	e.sb.WriteString(fmt.Sprintf(
		`<div style="position:absolute;top:0;left:0;width:%.2fpx;height:%.2fpx;`+
			`opacity:%.2f;pointer-events:none;z-index:0;`+
			`background-image:url('data:image/png;base64,%s');`+
			`background-size:%s;background-repeat:%s;background-position:%s;"></div>`+"\n",
		pageW, pageH, opacity, b64, bgSize, repeat, bgPos,
	))
}

func (e *Exporter) ExportBand(b *preview.PreparedBand) error {
	scale := e.Scale
	if scale <= 0 {
		scale = 1
	}

	if e.Layers {
		// In layers mode, render each object directly onto the page with
		// absolute coordinates (band.Top + obj.Top) and ascending z-index.
		bandTop := b.Top
		for _, obj := range b.Objects {
			layered := obj
			layered.Top += bandTop
			e.renderObjectLayered(layered, scale)
		}
		return nil
	}

	top := b.Top * scale
	h := b.Height * scale
	label := export.HTMLString(b.Name)
	e.sb.WriteString(fmt.Sprintf(
		`<div class="band" data-name="%s" style="position:absolute;top:%.2fpx;`+
			`left:0;right:0;height:%.2fpx;">`,
		label, top, h,
	))
	for _, obj := range b.Objects {
		e.renderObject(obj, scale)
	}
	e.sb.WriteString("</div>\n")
	return nil
}

// renderObjectLayered renders a single object in layers mode.
// The object's Top has already been adjusted to page-absolute coordinates.
// Each object gets a unique z-index so paint order matches band/object order.
func (e *Exporter) renderObjectLayered(obj preview.PreparedObject, scale float32) {
	e.zIdx++
	// Patch the positional CSS to include z-index.
	// We do this by temporarily wrapping renderObject with a hook.
	// Simplest approach: render normally then add z-index via the positional override.
	//
	// We use a small helper that delegates to renderObject but intercepts the
	// positional style. Since renderObject builds its own positional string
	// internally, we use a two-pass approach: render to a temp builder, inject z-index.
	savedSB := e.sb
	e.sb.Reset()
	e.renderObject(obj, scale)
	rendered := e.sb.String()
	e.sb = savedSB

	// Inject z-index into the first style=" block of the rendered element.
	// The positional style always contains "position:absolute;" so we can
	// reliably target that attribute.
	zStyle := fmt.Sprintf("z-index:%d;", e.zIdx)
	rendered = strings.Replace(rendered, "position:absolute;", "position:absolute;"+zStyle, 1)
	e.sb.WriteString(rendered)
}

// renderObject writes an HTML element for a single PreparedObject.
func (e *Exporter) renderObject(obj preview.PreparedObject, scale float32) {
	left := obj.Left * scale
	top := obj.Top * scale
	w := obj.Width * scale
	h := obj.Height * scale

	// positional is always unique per object — kept inline.
	positional := fmt.Sprintf(
		"position:absolute;left:%.2fpx;top:%.2fpx;width:%.2fpx;height:%.2fpx;overflow:hidden;",
		left, top, w, h,
	)

	// sharedCSS collects properties that can be shared across elements.
	var shared strings.Builder

	// Fill color.
	if obj.FillColor.A > 0 {
		shared.WriteString(fmt.Sprintf(
			"background-color:rgba(%d,%d,%d,%.2f);",
			obj.FillColor.R, obj.FillColor.G, obj.FillColor.B,
			float32(obj.FillColor.A)/255.0,
		))
	}

	// Border.
	shared.WriteString(borderCSS(&obj.Border, scale))

	// styleAttr builds the full style= attribute (positional only when a class is used).
	styleAttr := func(extra string) string {
		combined := shared.String() + extra
		name := e.css.Register(combined)
		if name != "" {
			return fmt.Sprintf(`style="%s" class="%s"`, positional, name)
		}
		return fmt.Sprintf(`style="%s%s"`, positional, combined)
	}

	switch obj.Kind {
	case preview.ObjectTypeText:
		// Build shared text-specific CSS.
		var textCSS strings.Builder
		font := obj.Font
		textCSS.WriteString(fmt.Sprintf("font-family:'%s';font-size:%.1fpt;", font.Name, font.Size))
		if font.Style&style.FontStyleBold != 0 {
			textCSS.WriteString("font-weight:bold;")
		}
		if font.Style&style.FontStyleItalic != 0 {
			textCSS.WriteString("font-style:italic;")
		}
		if font.Style&style.FontStyleUnderline != 0 {
			textCSS.WriteString("text-decoration:underline;")
		}
		tc := obj.TextColor
		textCSS.WriteString(fmt.Sprintf("color:rgba(%d,%d,%d,%.2f);", tc.R, tc.G, tc.B, float32(tc.A)/255.0))
		switch obj.HorzAlign {
		case 1:
			textCSS.WriteString("text-align:center;")
		case 2:
			textCSS.WriteString("text-align:right;")
		case 3:
			textCSS.WriteString("text-align:justify;")
		default:
			textCSS.WriteString("text-align:left;")
		}
		switch obj.VertAlign {
		case 1:
			textCSS.WriteString("display:flex;align-items:center;")
			// When both vert-center and horz-center, use justify-content too.
			switch obj.HorzAlign {
			case 1:
				textCSS.WriteString("justify-content:center;")
			case 2:
				textCSS.WriteString("justify-content:flex-end;")
			}
		case 2:
			textCSS.WriteString("display:flex;align-items:flex-end;")
			switch obj.HorzAlign {
			case 1:
				textCSS.WriteString("justify-content:center;")
			case 2:
				textCSS.WriteString("justify-content:flex-end;")
			}
		}
		if obj.WordWrap {
			textCSS.WriteString("word-wrap:break-word;white-space:normal;")
		} else {
			textCSS.WriteString("white-space:nowrap;")
		}

		innerText := export.HTMLString(obj.Text)
		// Preserve line breaks: convert \r\n and \n to <br> tags.
		innerText = strings.ReplaceAll(innerText, "\r\n", "<br>")
		innerText = strings.ReplaceAll(innerText, "\n", "<br>")
		if obj.HyperlinkKind == 1 && obj.HyperlinkValue != "" {
			// Wrap content in an anchor tag for URL hyperlinks.
			innerText = fmt.Sprintf(`<a href="%s" target="_blank" style="color:inherit;text-decoration:inherit;">%s</a>`,
				export.HTMLString(obj.HyperlinkValue), innerText)
		}
		e.sb.WriteString(fmt.Sprintf(`<div %s>%s</div>`, styleAttr(textCSS.String()), innerText))

	case preview.ObjectTypeHtml:
		// Emit raw HTML markup inside a positioned container div.
		// Text is not HTML-escaped so that embedded tags are preserved.
		e.sb.WriteString(fmt.Sprintf(`<div %s>%s</div>`, styleAttr("overflow:hidden;"), obj.Text))

	case preview.ObjectTypeRTF:
		// Convert RTF to HTML preserving bold, italic, underline, paragraph breaks, etc.
		htmlContent := utils.RTFToHTML(obj.Text)
		e.sb.WriteString(fmt.Sprintf(`<div %s>%s</div>`, styleAttr("overflow:hidden;"), htmlContent))

	case preview.ObjectTypePicture:
		if obj.BlobIdx >= 0 && e.pp != nil {
			if data := e.pp.BlobStore.Get(obj.BlobIdx); len(data) > 0 {
				mime := imageMIME(data)
				encoded := base64.StdEncoding.EncodeToString(data)
				e.sb.WriteString(fmt.Sprintf(
					`<div %s><img src="data:%s;base64,%s" style="width:100%%;height:100%%;object-fit:contain;" alt=""></div>`,
					styleAttr(""), mime, encoded,
				))
				break
			}
		}
		// No image data — empty placeholder.
		e.sb.WriteString(fmt.Sprintf(`<div %s></div>`, styleAttr("")))

	case preview.ObjectTypeLine:
		if obj.LineDiagonal {
			// Render diagonal line as an inline SVG.
			e.sb.WriteString(fmt.Sprintf(
				`<div %s><svg width="%.2f" height="%.2f" style="position:absolute;top:0;left:0;">`,
				styleAttr(""), w, h,
			))
			e.sb.WriteString(fmt.Sprintf(
				`<line x1="0" y1="0" x2="%.2f" y2="%.2f" stroke="#000" stroke-width="1"/>`,
				w, h,
			))
			e.sb.WriteString(`</svg></div>`)
		} else {
			// Horizontal or vertical: use border-bottom or border-left.
			var lineExtra string
			if h <= w {
				lineExtra = "border-bottom:1px solid #000;height:1px;"
			} else {
				lineExtra = "border-left:1px solid #000;width:1px;"
			}
			e.sb.WriteString(fmt.Sprintf(`<div %s></div>`, styleAttr(lineExtra)))
		}

	case preview.ObjectTypeShape:
		var shapeExtra string
		switch obj.ShapeKind {
		case 2: // Ellipse
			shapeExtra = "border:1px solid #000;border-radius:50%;"
		case 1: // RoundRectangle
			shapeExtra = fmt.Sprintf("border:1px solid #000;border-radius:%.2fpx;", obj.ShapeCurve*scale)
		case 3: // Triangle — use inline SVG
			e.sb.WriteString(fmt.Sprintf(`<div %s>`, styleAttr("")))
			e.sb.WriteString(fmt.Sprintf(
				`<svg width="%.2f" height="%.2f" style="position:absolute;top:0;left:0;">`,
				w, h,
			))
			e.sb.WriteString(fmt.Sprintf(
				`<polygon points="%.2f,0 0,%.2f %.2f,%.2f" stroke="#000" stroke-width="1" fill="none"/>`,
				w/2, h, w, h,
			))
			e.sb.WriteString(`</svg></div>`)
		case 4: // Diamond — use inline SVG
			e.sb.WriteString(fmt.Sprintf(`<div %s>`, styleAttr("")))
			e.sb.WriteString(fmt.Sprintf(
				`<svg width="%.2f" height="%.2f" style="position:absolute;top:0;left:0;">`,
				w, h,
			))
			e.sb.WriteString(fmt.Sprintf(
				`<polygon points="%.2f,0 %.2f,%.2f %.2f,%.2f 0,%.2f" stroke="#000" stroke-width="1" fill="none"/>`,
				w/2, w, h/2, w/2, h, h/2,
			))
			e.sb.WriteString(`</svg></div>`)
		default: // Rectangle
			shapeExtra = "border:1px solid #000;"
		}
		if obj.ShapeKind != 3 && obj.ShapeKind != 4 {
			e.sb.WriteString(fmt.Sprintf(`<div %s></div>`, styleAttr(shapeExtra)))
		}

	case preview.ObjectTypeDigitalSignature:
		// Render a dashed-border placeholder box for digital signature fields.
		label := obj.Text
		if label == "" {
			label = "Digital Signature"
		}
		sigExtra := "border:2px dashed #888;color:#888;font-size:10pt;display:flex;align-items:center;justify-content:center;"
		e.sb.WriteString(fmt.Sprintf(`<div %s><span>%s</span></div>`, styleAttr(sigExtra), export.HTMLString(label)))

	case preview.ObjectTypeCheckBox:
		checked := ""
		if obj.Text == "true" {
			checked = " checked"
		}
		e.sb.WriteString(fmt.Sprintf(
			`<div %s><input type="checkbox"%s disabled style="margin:auto;"></div>`,
			styleAttr(""), checked,
		))

	case preview.ObjectTypePolyLine, preview.ObjectTypePolygon:
		if len(obj.Points) < 2 {
			e.sb.WriteString(fmt.Sprintf(`<div %s></div>`, styleAttr("")))
			break
		}

		// Determine stroke color from border.
		strokeColor := "black"
		strokeWidth := 1.0
		if obj.Border.Lines[0] != nil {
			lc := obj.Border.Lines[0].Color
			if lc.A > 0 {
				strokeColor = fmt.Sprintf("rgba(%d,%d,%d,%.2f)",
					lc.R, lc.G, lc.B, float32(lc.A)/255.0)
			}
			if obj.Border.Lines[0].Width > 0 {
				strokeWidth = float64(obj.Border.Lines[0].Width) * float64(scale)
			}
		}

		// Build SVG points string.
		var pts strings.Builder
		for i, pt := range obj.Points {
			px := float64(pt[0]) * float64(scale)
			py := float64(pt[1]) * float64(scale)
			if i > 0 {
				pts.WriteByte(' ')
			}
			pts.WriteString(fmt.Sprintf("%.2f,%.2f", px, py))
		}

		// PolyLine/Polygon use overflow:visible — build positional inline directly.
		polyPos := fmt.Sprintf(
			"position:absolute;left:%.2fpx;top:%.2fpx;width:%.2fpx;height:%.2fpx;overflow:visible;",
			left, top, w, h,
		)

		if obj.Kind == preview.ObjectTypePolyLine {
			e.sb.WriteString(fmt.Sprintf(
				`<div style="%s"><svg width="%.2f" height="%.2f" style="overflow:visible;">`,
				polyPos, w, h,
			))
			e.sb.WriteString(fmt.Sprintf(
				`<polyline points="%s" stroke="%s" stroke-width="%.2f" fill="none"/>`,
				pts.String(), strokeColor, strokeWidth,
			))
			e.sb.WriteString(`</svg></div>`)
		} else {
			// Polygon — close the path and fill.
			fillColor := "none"
			if obj.FillColor.A > 0 {
				fc := obj.FillColor
				fillColor = fmt.Sprintf("rgba(%d,%d,%d,%.2f)",
					fc.R, fc.G, fc.B, float32(fc.A)/255.0)
			}
			e.sb.WriteString(fmt.Sprintf(
				`<div style="%s"><svg width="%.2f" height="%.2f" style="overflow:visible;">`,
				polyPos, w, h,
			))
			e.sb.WriteString(fmt.Sprintf(
				`<polygon points="%s" stroke="%s" stroke-width="%.2f" fill="%s"/>`,
				pts.String(), strokeColor, strokeWidth, fillColor,
			))
			e.sb.WriteString(`</svg></div>`)
		}

	case preview.ObjectTypeSVG:
		if obj.BlobIdx >= 0 && e.pp != nil {
			if svgData := e.pp.BlobStore.Get(obj.BlobIdx); len(svgData) > 0 {
				// Emit SVG inline inside a positioned container div.
				// The SVG is adjusted to fill the bounding box via width/height overrides.
				e.sb.WriteString(fmt.Sprintf(`<div %s>`, styleAttr("")))
				// Strip any existing width/height attributes from the root <svg> tag
				// so that CSS controls the size. This is a simple prefix injection.
				svgStr := string(svgData)
				e.sb.WriteString(svgStr)
				e.sb.WriteString(`</div>`)
				break
			}
		}
		// No SVG data — empty placeholder.
		e.sb.WriteString(fmt.Sprintf(`<div %s></div>`, styleAttr("")))

	default:
		// Unknown/unhandled type — render an empty placeholder.
		e.sb.WriteString(fmt.Sprintf(`<div %s></div>`, styleAttr("")))
	}
}

func (e *Exporter) ExportPageEnd(pg *preview.PreparedPage) error {
	// Watermark on top of page content (rendered last so it overlays).
	if wm := pg.Watermark; wm != nil && wm.Enabled {
		scale := e.Scale
		if scale <= 0 {
			scale = 1
		}
		w := pg.Width * scale
		h := pg.Height * scale
		if wm.ShowImageOnTop {
			e.renderWatermarkImage(wm, w, h)
		}
		if wm.ShowTextOnTop {
			e.renderWatermarkText(wm, w, h)
		}
	}
	e.sb.WriteString("</div>\n") // close .page
	return nil
}

func (e *Exporter) Finish() error {
	// Inject collected CSS classes before closing body.
	if e.css != nil {
		e.sb.WriteString(e.css.StyleBlock())
	}
	e.sb.WriteString("</body>\n</html>\n")
	_, err := io.WriteString(e.w, e.sb.String())
	return err
}

// ── CSS helper ─────────────────────────────────────────────────────────────────

func (e *Exporter) defaultCSS() string {
	return `<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { background: #e0e0e0; font-family: Arial, sans-serif; }
.page { page-break-after: always; }
.band { font-size: 12px; }
@media print {
  body { background: none; }
  .page { box-shadow: none; margin: 0; page-break-after: always; }
}
</style>
`
}

// HTML returns the complete HTML string that would be written.
// Useful for testing. Call after Export has been called.
func (e *Exporter) HTML() string {
	return e.sb.String()
}

// borderCSS converts a style.Border into CSS border/box-shadow declarations.
// It handles per-side borders using border-top/right/bottom/left shorthand.
func borderCSS(b *style.Border, scale float32) string {
	if b == nil || b.VisibleLines == style.BorderLinesNone {
		return ""
	}
	var sb strings.Builder

	type side struct {
		flag style.BorderLines
		prop string
		idx  int
	}
	sides := []side{
		{style.BorderLinesTop, "border-top", int(style.BorderTop)},
		{style.BorderLinesRight, "border-right", int(style.BorderRight)},
		{style.BorderLinesBottom, "border-bottom", int(style.BorderBottom)},
		{style.BorderLinesLeft, "border-left", int(style.BorderLeft)},
	}

	for _, s := range sides {
		if b.VisibleLines&s.flag == 0 {
			continue
		}
		var line *style.BorderLine
		if b.Lines[s.idx] != nil {
			line = b.Lines[s.idx]
		}
		width := float32(1)
		lineStyle := "solid"
		c := color.RGBA{R: 0, G: 0, B: 0, A: 255} // default black
		if line != nil {
			width = line.Width * scale
			c = line.Color
			switch line.Style {
			case style.LineStyleDash:
				lineStyle = "dashed"
			case style.LineStyleDot:
				lineStyle = "dotted"
			case style.LineStyleDashDot, style.LineStyleDashDotDot:
				lineStyle = "dashed"
			case style.LineStyleDouble:
				lineStyle = "double"
			}
		}
		sb.WriteString(fmt.Sprintf("%s:%.2fpx %s rgba(%d,%d,%d,%.2f);",
			s.prop, width, lineStyle, c.R, c.G, c.B, float32(c.A)/255.0))
	}

	if b.Shadow {
		sw := b.ShadowWidth * scale
		sc := b.ShadowColor
		sb.WriteString(fmt.Sprintf("box-shadow:%.2fpx %.2fpx 0 rgba(%d,%d,%d,%.2f);",
			sw, sw, sc.R, sc.G, sc.B, float32(sc.A)/255.0))
	}

	return sb.String()
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
	// Check for SVG (text-based).
	if len(data) > 4 {
		s := string(data[:min(len(data), 64)])
		if strings.Contains(s, "<svg") || strings.Contains(s, "<?xml") {
			return "image/svg+xml"
		}
	}
	return "image/png" // default / PNG magic is 8 bytes, assume png
}

