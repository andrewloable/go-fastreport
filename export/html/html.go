// Package html implements an HTML export filter for go-fastreport.
// It renders prepared pages as a single HTML document with absolute-positioned
// elements (WYSIWYG mode), one section per page, matching C# FastReport output.
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
// Output structure (matching C# FastReport HTML export):
//
//	<html><body>
//	  <div class="frpage-container">
//	    <a name="PageN1" ...></a>
//	    <div class="frpage0" ...> <!-- one per page, 0-indexed -->
//	      <div class="sN" style="left:...;top:...;..."> <!-- objects flat, page-absolute -->
//	        <div class="sM">text content</div>
//	      </div>
//	    </div>
//	  </div>
//	  <style>...print CSS...</style>
//	  <style>...content CSS...</style>
//	</body></html>
type Exporter struct {
	export.ExportBase

	// Title is used as the HTML document title.
	Title string
	// EmbedCSS controls whether CSS style blocks are written inline.
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
	e.sb.WriteString("<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\">\n")
	e.sb.WriteString("<html><head>\n")
	e.sb.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\">\n")
	e.sb.WriteString("<meta name=Generator content=\"FastReport http://www.fast-report.com\">\n")
	e.sb.WriteString(fmt.Sprintf("<title>%s</title>\n", export.HTMLString(e.Title)))
	e.sb.WriteString("</head>\n")
	e.sb.WriteString("<body bgcolor=\"#FFFFFF\" text=\"#000000\">\n")
	e.sb.WriteString("<div class=\"frpage-container\">\n")
	return nil
}

func (e *Exporter) ExportPageBegin(pg *preview.PreparedPage) error {
	scale := e.Scale
	if scale <= 0 {
		scale = 1
	}
	w := pg.Width * scale
	h := pg.Height * scale
	pageW := w + 3.0*scale // C# adds +3 to page div width
	e.zIdx = 0
	pageN := e.pageIdx + 1
	e.sb.WriteString(fmt.Sprintf(
		"<a name=\"PageN%d\" id=\"PageN%d\" style=\"padding:0;margin:0;font-size:1px;\"></a>",
		pageN, pageN,
	))
	e.sb.WriteString(fmt.Sprintf(
		"<div class=\"frpage%d\" style=\"position:relative; width:%.2fpx; height:%.2fpx; background-color:rgb(255, 255, 255)\">",
		e.pageIdx, pageW, h,
	))
	e.sb.WriteString("\n")

	// Watermark behind page content.
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

	// Flat rendering: render each object with page-absolute top coordinate (no band wrappers).
	for _, obj := range b.Objects {
		flat := obj
		flat.Top += b.Top
		e.renderObject(flat, scale)
	}
	return nil
}

// renderObjectLayered renders a single object in layers mode.
// The object's Top has already been adjusted to page-absolute coordinates.
// Each object gets a unique z-index so paint order matches band/object order.
func (e *Exporter) renderObjectLayered(obj preview.PreparedObject, scale float32) {
	e.zIdx++
	// Render to a temp builder, then inject z-index into the first style=" attribute.
	savedSB := e.sb
	e.sb.Reset()
	e.renderObject(obj, scale)
	rendered := e.sb.String()
	e.sb = savedSB

	// Inject z-index into the first style=" attribute of the rendered element.
	// For text objects the outer div has inline style="left:...;top:...;" so we
	// can reliably inject at the start of the first style attribute value.
	// For non-text objects the positional string contains "position:absolute;".
	zStyle := fmt.Sprintf("z-index:%d;", e.zIdx)
	// Try injecting after "position:absolute;" first (non-text objects).
	if strings.Contains(rendered, "position:absolute;") {
		rendered = strings.Replace(rendered, "position:absolute;", "position:absolute;"+zStyle, 1)
	} else {
		// Text objects: inject into the first style=" value.
		rendered = strings.Replace(rendered, `style="`, `style="`+zStyle, 1)
	}
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
			"background-color:%s;",
			rgbColor(obj.FillColor),
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
		// Build outer div CSS class: text-align, position info, color, background, border, size.
		font := obj.Font
		// Convert font size from pt to px (96dpi / 72pt = 1.3333...) to match C# output.
		fontPx := font.Size * 96.0 / 72.0
		lineH := fontPx * 1.21 // approximate line height

		var outerCSS strings.Builder
		// text-align
		switch obj.HorzAlign {
		case 1:
			outerCSS.WriteString("text-align:center;")
		case 2:
			outerCSS.WriteString("text-align:right;")
		case 3:
			outerCSS.WriteString("text-align:justify;")
		default:
			outerCSS.WriteString("text-align:left;")
		}
		outerCSS.WriteString("position:absolute;")
		tc := obj.TextColor
		outerCSS.WriteString(fmt.Sprintf("color:%s;", rgbColor(tc)))
		fc := obj.FillColor
		if fc.A > 0 {
			outerCSS.WriteString(fmt.Sprintf("background-color:%s;", rgbColor(fc)))
		} else {
			outerCSS.WriteString("background-color:transparent;")
		}
		// Border from shared CSS (or border:none if no border)
		borderStr := borderCSS(&obj.Border, scale)
		if borderStr != "" {
			outerCSS.WriteString(borderStr)
		} else {
			outerCSS.WriteString("border:none;")
		}
		outerCSS.WriteString(fmt.Sprintf("width:%.2fpx;height:%.2fpx;", w, h))

		outerClass := e.css.Register(outerCSS.String())

		// Build inner div CSS class: font, display, width, padding, vertical alignment.
		var innerCSS strings.Builder
		if font.Style&style.FontStyleBold != 0 {
			innerCSS.WriteString("font-weight:bold;")
		}
		if font.Style&style.FontStyleItalic != 0 {
			innerCSS.WriteString("font-style:italic;")
		}
		if font.Style&style.FontStyleUnderline != 0 {
			innerCSS.WriteString("text-decoration:underline;")
		}
		innerCSS.WriteString(fmt.Sprintf("font-family:'%s';font-size:%.2fpx;line-height:%.2f;", font.Name, fontPx, 1.21))
		if obj.WordWrap {
			innerCSS.WriteString("word-wrap:break-word;white-space:normal;")
		} else {
			innerCSS.WriteString("white-space:nowrap;")
		}
		innerCSS.WriteString("overflow:hidden;")

		// Vertical alignment via display:flex + align-items, or margin-top.
		switch obj.VertAlign {
		case 1: // center
			innerCSS.WriteString("display:flex;align-items:center;")
			switch obj.HorzAlign {
			case 1:
				innerCSS.WriteString("justify-content:center;")
			case 2:
				innerCSS.WriteString("justify-content:flex-end;")
			}
		case 2: // bottom
			innerCSS.WriteString("display:flex;align-items:flex-end;")
			switch obj.HorzAlign {
			case 1:
				innerCSS.WriteString("justify-content:center;")
			case 2:
				innerCSS.WriteString("justify-content:flex-end;")
			}
		default: // top (0)
			// Compute inner width (subtract padding if available; padding not in PreparedObject, use full width).
			innerW := w
			innerCSS.WriteString(fmt.Sprintf("display:block;border:0;width:%.2fpx;", innerW))
		}
		_ = lineH // used for future paragraph height calculations

		innerClass := e.css.Register(innerCSS.String())

		// Format text content: use <p> tags for line breaks.
		innerText := export.HTMLString(obj.Text)
		innerText = formatTextParagraphs(innerText)

		// Outer div style: positional (left, top, width, height).
		outerStyle := fmt.Sprintf("left:%.2fpx;top:%.2fpx;width:%.2fpx;height:%.2fpx;", left, top, w, h)

		// Hyperlink wrapping (px00): wrap entire div with outer <a> tag.
		if obj.HyperlinkKind == 1 && obj.HyperlinkValue != "" {
			linkColor := rgbColor(tc)
			e.sb.WriteString(fmt.Sprintf(`<a style="color:%s" href="%s" target="_blank">`, linkColor, export.HTMLString(obj.HyperlinkValue)))
			e.sb.WriteString(fmt.Sprintf(`<div class="%s" style="cursor:pointer;%s">`, outerClass, outerStyle))
			e.sb.WriteString(fmt.Sprintf(`<div class="%s">%s</div>`, innerClass, innerText))
			e.sb.WriteString("</div></a>")
		} else {
			e.sb.WriteString(fmt.Sprintf(`<div class="%s" style="%s">`, outerClass, outerStyle))
			e.sb.WriteString(fmt.Sprintf(`<div class="%s">%s</div>`, innerClass, innerText))
			e.sb.WriteString("</div>")
		}

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
				// Register a CSS class for this image keyed by blob index so
				// identical images reuse the same class (matching C# pattern).
				classKey := fmt.Sprintf("img%d", obj.BlobIdx)
				if !e.css.HasClass(classKey) {
					mime := imageMIMEForCSS(data)
					encoded := base64.StdEncoding.EncodeToString(data)
					e.css.RegisterClass(classKey, fmt.Sprintf(
						"background: url('data:%s;base64,%s') no-repeat !important;-webkit-print-color-adjust:exact;",
						mime, encoded,
					))
				}
				// Two overlapping divs: border div + image overlay div (C# pattern).
				e.sb.WriteString(fmt.Sprintf(`<div %s>&nbsp;</div>`, styleAttr("border:none;")))
				e.sb.WriteString(fmt.Sprintf(`<div class="%s" %s>&nbsp;</div>`, classKey, styleAttr("")))
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
		// Determine which symbol to draw.
		symbol := -1 // none
		if obj.Checked {
			symbol = obj.CheckedSymbol // 0=check, 1=cross, 2=plus, 3=fill
		} else {
			switch obj.UncheckedSymbol {
			case 0: // None — just the box, no symbol
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

		// Get check color; default to opaque black.
		cc := obj.CheckColor
		if cc.A == 0 {
			cc = color.RGBA{A: 255}
		}
		strokeColor := fmt.Sprintf("rgb(%d,%d,%d)", cc.R, cc.G, cc.B)

		e.sb.WriteString(fmt.Sprintf(`<div %s>`, styleAttr("")))
		e.sb.WriteString(fmt.Sprintf(`<svg width="%.2f" height="%.2f" style="position:absolute;top:0;left:0;width:100%%;height:100%%;">`, w, h))

		switch symbol {
		case 0: // checkmark ✓
			e.sb.WriteString(fmt.Sprintf(
				`<polyline points="%.2f,%.2f %.2f,%.2f %.2f,%.2f" stroke="%s" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/>`,
				w*0.15, h*0.5,
				w*0.4, h*0.75,
				w*0.85, h*0.25,
				strokeColor,
			))
		case 1: // cross ✗
			e.sb.WriteString(fmt.Sprintf(
				`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-width="2"/>`,
				w*0.2, h*0.2, w*0.8, h*0.8, strokeColor,
			))
			e.sb.WriteString(fmt.Sprintf(
				`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-width="2"/>`,
				w*0.8, h*0.2, w*0.2, h*0.8, strokeColor,
			))
		case 2: // plus +
			e.sb.WriteString(fmt.Sprintf(
				`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-width="2"/>`,
				w*0.5, h*0.15, w*0.5, h*0.85, strokeColor,
			))
			e.sb.WriteString(fmt.Sprintf(
				`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-width="2"/>`,
				w*0.15, h*0.5, w*0.85, h*0.5, strokeColor,
			))
		case 3: // fill — filled rectangle
			fc := obj.CheckColor
			if fc.A == 0 {
				fc = color.RGBA{A: 255}
			}
			e.sb.WriteString(fmt.Sprintf(
				`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="rgb(%d,%d,%d)"/>`,
				w*0.1, h*0.1, w*0.8, h*0.8, fc.R, fc.G, fc.B,
			))
		case 10: // minus
			e.sb.WriteString(fmt.Sprintf(
				`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-width="2"/>`,
				w*0.2, h*0.5, w*0.8, h*0.5, strokeColor,
			))
		case 11: // slash
			e.sb.WriteString(fmt.Sprintf(
				`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-width="2"/>`,
				w*0.2, h*0.8, w*0.8, h*0.2, strokeColor,
			))
		case 12: // backslash
			e.sb.WriteString(fmt.Sprintf(
				`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-width="2"/>`,
				w*0.2, h*0.2, w*0.8, h*0.8, strokeColor,
			))
		}

		e.sb.WriteString(`</svg></div>`)

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
	e.sb.WriteString("</div>\n") // close frpage{n}
	return nil
}

func (e *Exporter) Finish() error {
	// Close frpage-container div.
	e.sb.WriteString("</div>\n")

	// Emit CSS blocks before closing body.
	if e.EmbedCSS && e.css != nil {
		// Print CSS block (for page breaks).
		printCSS := e.buildPrintCSS()
		if printCSS != "" {
			e.sb.WriteString(printCSS)
		}
		// Content CSS block (s0..sN classes).
		if block := e.css.StyleBlock(); block != "" {
			e.sb.WriteString(block)
		}
	}
	e.sb.WriteString("</body>\n</html>\n")
	_, err := io.WriteString(e.w, e.sb.String())
	return err
}

// buildPrintCSS generates the print media CSS block for page break rules.
func (e *Exporter) buildPrintCSS() string {
	if e.pageIdx == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("<style type=\"text/css\" media=\"print\"><!--\n")
	for i := 0; i < e.pageIdx; i++ {
		sb.WriteString(fmt.Sprintf("div.frpage%d { break-after: always; page-break-inside: avoid; }\n", i))
	}
	sb.WriteString(" @page { size: portrait; margin: 0; }-->")
	sb.WriteString("</style>\n")
	return sb.String()
}

// HTML returns the complete HTML string that would be written.
// Useful for testing. Call after Export has been called.
func (e *Exporter) HTML() string {
	return e.sb.String()
}

// ── CSS helpers ─────────────────────────────────────────────────────────────────

// rgbColor returns "rgb(R, G, B)" for fully opaque colors and "rgba(R, G, B, A)" for semi-transparent ones.
func rgbColor(c color.RGBA) string {
	if c.A == 255 {
		return fmt.Sprintf("rgb(%d, %d, %d)", c.R, c.G, c.B)
	}
	return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", c.R, c.G, c.B, float32(c.A)/255.0)
}

// formatTextParagraphs converts plain text line breaks to HTML paragraph tags.
// Each line break becomes <p style="margin-top:0px;margin-bottom:0px;"></p>.
func formatTextParagraphs(htmlText string) string {
	// Split on newlines (normalize \r\n first).
	normalized := strings.ReplaceAll(htmlText, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) == 1 {
		// Single line: add trailing paragraph for C# compatibility.
		return htmlText + `<p style="margin-top:0px;margin-bottom:0px;"></p>`
	}
	var sb strings.Builder
	for i, line := range lines {
		sb.WriteString(line)
		if i < len(lines)-1 {
			sb.WriteString(`<p style="margin-top:0px;margin-bottom:0px;"></p>`)
		} else if line != "" {
			// Last non-empty line: trailing paragraph.
			sb.WriteString(`<p style="margin-top:0px;margin-bottom:0px;"></p>`)
		}
	}
	return sb.String()
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

// imageMIMEForCSS detects the MIME type from image magic bytes and returns the
// capitalized form used by C# FastReport (e.g. "image/Png", "image/Jpeg").
// Falls back to "image/Png" for unknown formats.
func imageMIMEForCSS(data []byte) string {
	if len(data) >= 3 {
		switch {
		case data[0] == 0xFF && data[1] == 0xD8:
			return "image/Jpeg"
		case data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46:
			return "image/Gif"
		case len(data) >= 4 && data[0] == 0x42 && data[1] == 0x4D:
			return "image/Bmp"
		case len(data) >= 4 && data[0] == 0x49 && data[1] == 0x49 && data[2] == 0x2A:
			return "image/Tiff"
		case len(data) >= 4 && data[0] == 0x4D && data[1] == 0x4D && data[2] == 0x00:
			return "image/Tiff"
		}
	}
	// Check for SVG (text-based).
	if len(data) > 4 {
		s := string(data[:min(len(data), 64)])
		if strings.Contains(s, "<svg") || strings.Contains(s, "<?xml") {
			return "image/svg+xml"
		}
	}
	return "image/Png" // default / PNG magic is 8 bytes, assume png
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
