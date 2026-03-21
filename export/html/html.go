// Package html implements an HTML export filter for go-fastreport.
// It renders prepared pages as a single HTML document with absolute-positioned
// elements (WYSIWYG mode), one section per page, matching C# FastReport output.
package html

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"strings"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"

	xdraw "golang.org/x/image/draw"
)

// Exporter produces HTML output from a PreparedPages collection.
//
// Output structure (matching C# FastReport HTML export):
//
//	<html><body>
//	  <style media="print">...page break CSS...</style>
//	  <style>...content CSS classes...</style>
//	  <div class="frpage-container">
//	    <a name="PageN1" ...></a>
//	    <div class="frpage0" ...>  <!-- one per page, 0-indexed -->
//	      <div class="sN" style="left:...;top:...;...">  <!-- objects -->
//	        <div class="sM">text content</div>
//	      </div>
//	    </div>
//	  </div>
//	</body></html>
//
// When InlineStyles is true, all CSS properties are emitted as inline style
// attributes instead of CSS class references. This matches the C# InlineStyles
// option in HTMLExportStyles.cs: GetStyle(string) → InlineStyle(style) vs
// GetStyleTag(index). Use InlineStyles=true when embedding the HTML in email or
// other contexts where external CSS classes may be stripped.
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
	// InlineStyles controls whether CSS properties are emitted as inline
	// style="..." attributes (true) or as CSS class references (false, default).
	// When InlineStyles=true the <style> block is omitted and all styling is
	// inlined on each element. Matches C# HTMLExportStyles.InlineStyles.
	// C# reference: HTMLExportStyles.cs GetStyle(string) → InlineStyle(style).
	InlineStyles bool
	// Mode controls the output structure.
	// ExportModeSingleFile (default): one HTML file written to the io.Writer.
	// ExportModeMultiPage: one file per page written to OutputDir.
	// ExportModeNavigator: per-page files + index.html + nav.html in OutputDir.
	// C# reference: HTMLExport.cs ExportType enum, SinglePage/Navigator fields.
	Mode ExportMode
	// OutputDir is the directory where files are written in MultiPage and
	// Navigator modes. It is created if it does not exist.
	OutputDir string
	// BaseName is the base file name (without extension) used for per-page
	// files and the navigator prefix. Defaults to "page" when empty.
	BaseName string

	w       io.Writer
	pp      *preview.PreparedPages
	sb      strings.Builder
	pageIdx int
	css     *cssRegistry
	// zIdx is the current z-index counter, incremented per object in Layers mode.
	zIdx int
	// cssCountBeforePage tracks the CSS class count before each page starts,
	// so we can emit only new classes per page (matching C# per-page CSS emission).
	cssCountBeforePage int
	// pageBuf buffers each page's rendered objects. In ExportPageEnd, the per-page
	// CSS is prepended before the page content, matching C#'s per-page emission.
	pageBuf strings.Builder
	// pageH is the declared page height for the current page (set in ExportPageBegin).
	pageH float32
	// maxBottom tracks the maximum Y+Height across all bands on the current page.
	// Used for UnlimitedHeight pages where the page div must expand to fit content.
	maxBottom float32
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
	// Head and body preamble are written here; the CSS blocks and frpage-container
	// are assembled in Finish() to match C# order (CSS before content).
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
	e.pageH = h
	e.maxBottom = 0
	pageN := e.pageIdx + 1

	// Record CSS count before this page's objects are rendered.
	e.cssCountBeforePage = e.css.Count()

	// Swap to page buffer — all object rendering writes to pageBuf.
	e.pageBuf.Reset()
	e.sb, e.pageBuf = e.pageBuf, e.sb // pageBuf now holds prior output; sb is fresh for this page

	e.sb.WriteString(fmt.Sprintf(
		"<a name=\"PageN%d\" id=\"PageN%d\" style=\"padding:0;margin:0;font-size:1px;\"></a>",
		pageN, pageN,
	))

	// C# inserts a page-break div before pages 2+ (singlePage && pageBreaks).
	if e.pageIdx > 0 {
		e.sb.WriteString("<div style=\"break-after:page\"></div>")
	}

	e.sb.WriteString(fmt.Sprintf(
		"<div class=\"frpage%d\" style=\"position:relative; width:%spx; height:%spx; background-color:%s\">",
		e.pageIdx, pxVal(pageW), pxVal(h), "rgb(255, 255, 255)",
	))

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

	// Track maximum content bottom for UnlimitedHeight page expansion.
	bandBottom := (b.Top + b.Height) * scale
	if bandBottom > e.maxBottom {
		e.maxBottom = bandBottom
	}

	// Render band background div (C# LayerBack pattern).
	// C# renders each band as a positioned div with its fill color before child objects.
	if b.Width > 0 && b.Height > 0 {
		e.renderBandBackground(b, scale)
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

// renderBandBackground renders the band itself as a background div, matching
// C#'s ExportBandLayers → LayerBack(htmlPage, band, null) pattern.
// The style is: text-align:center; position:absolute; color:white; background-color:...; border:none;
func (e *Exporter) renderBandBackground(b *preview.PreparedBand, scale float32) {
	left := b.Left * scale
	w := b.Width * scale
	h := b.Height * scale
	top := b.Top * scale

	// Build CSS class matching C# GetStyle(null, White, FillColor, ...) for non-text objects.
	var css strings.Builder
	css.WriteString("text-align:center;position:absolute;color:rgb(255, 255, 255);")
	fc := b.FillColor
	if fc.A == 0 {
		css.WriteString("background-color:transparent;")
	} else {
		css.WriteString(fmt.Sprintf("background-color:%s;", rgbColor(fc)))
	}
	css.WriteString("border:none;")
	css.WriteString(fmt.Sprintf("width:%spx;height:%spx;", pxVal(w), pxVal(h)))

	if e.InlineStyles {
		// InlineStyles mode: merge all CSS (including position) into a single style attribute.
		// C# InlineStyle(style) → style="...{shared}left:...;top:...;..."
		e.sb.WriteString(fmt.Sprintf(
			"<div style=\"%sleft:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
			css.String(), pxVal(left), pxVal(top), pxVal(w), pxVal(h),
		))
	} else {
		className := e.css.Register(css.String())
		e.sb.WriteString(fmt.Sprintf(
			"<div class=\"%s\" style=\"left:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
			className, pxVal(left), pxVal(top), pxVal(w), pxVal(h),
		))
	}
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
	// Bookmark anchor — emit before the object element so that navigation lands
	// at the object's position. C# reference: HTMLExportLayers.cs ExportObject →
	// if (!String.IsNullOrEmpty(obj.Bookmark)) htmlPage.Append("<a name=\"...\">");
	if obj.Bookmark != "" {
		e.sb.WriteString(fmt.Sprintf(`<a name="%s"></a>`, export.HTMLString(obj.Bookmark)))
	}

	left := obj.Left * scale
	top := obj.Top * scale
	w := obj.Width * scale
	h := obj.Height * scale

	// positional is always unique per object — kept inline.
	// C# Layer() only puts left/top/width/height in inline style.
	// position:absolute and overflow go in the CSS class.
	positional := fmt.Sprintf(
		"left:%spx;top:%spx;width:%spx;height:%spx;",
		pxVal(left), pxVal(top), pxVal(w), pxVal(h),
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
	// In InlineStyles mode, all CSS is merged into a single style= attribute.
	// C# reference: HTMLExportStyles.cs GetStyle(obj) → InlineStyle vs GetStyleTag.
	styleAttr := func(extra string) string {
		combined := shared.String() + extra
		if e.InlineStyles {
			// Inline mode: merge positional + shared CSS into a single style= attribute.
			return fmt.Sprintf(`style="%s%s"`, positional, combined)
		}
		name := e.css.Register(combined)
		if name != "" {
			return fmt.Sprintf(`style="%s" class="%s"`, positional, name)
		}
		return fmt.Sprintf(`style="%s%s"`, positional, combined)
	}

	switch obj.Kind {
	case preview.ObjectTypeText:
		e.renderTextObject(obj, scale)

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
				// C# LayerBack flow: GetStyle(obj) registers the picture's non-text
				// style class FIRST, then LayerPicture registers the image data class.
				// GetStyle for non-text: text-align:center;position:absolute;color:white;background-color:...;border:none;width;height
				var picCSS strings.Builder
				picCSS.WriteString("text-align:center;position:absolute;color:rgb(255, 255, 255);")
				if obj.FillColor.A == 0 {
					picCSS.WriteString("background-color:transparent;")
				} else {
					picCSS.WriteString(fmt.Sprintf("background-color:%s;", rgbColor(obj.FillColor)))
				}
				picCSS.WriteString(fmt.Sprintf("border:none;width:%spx;height:%spx;", pxVal(w), pxVal(h)))
				// picClass and imgClass are assigned below depending on InlineStyles mode.

				// For barcodes rendered at higher resolution (3x), keep the
				// high-res image and let the browser downscale via background-size.
				// For regular pictures, resize to display dimensions matching
				// C#'s GetLayerPicture behavior.
				var imgData []byte
				bgSize := ""
				if obj.IsBarcode {
					imgData = data
					bgSize = "background-size:100% 100%;"
				} else {
					targetW := int(math.Round(float64(w)))
					targetH := int(math.Round(float64(h)))
					imgData = resizeImagePNG(data, targetW, targetH)
				}
				mime := imageMIMEForCSS(imgData)
				encoded := base64.StdEncoding.EncodeToString(imgData)
				imgCSS := fmt.Sprintf(
					"%sbackground: url('data:%s;base64,%s') no-repeat !important;-webkit-print-color-adjust:exact;",
					bgSize, mime, encoded,
				)
				// Two overlapping divs: background/border div + image overlay div (C# pattern).
			// In InlineStyles mode, merge all CSS into inline style= attributes.
			// C# InlineStyle(style) → style="...{shared}left:...;top:..." pattern.
			// C# Layer() uses AppendLine("</div>") which adds \n after each div.
			if e.InlineStyles {
				// Inline mode: no CSS classes; emit style attributes directly.
				e.sb.WriteString(fmt.Sprintf(
					"<div style=\"%sleft:%spx;top:%spx;width:%spx;height:%spx;border:none;\">&nbsp;</div>\n",
					picCSS.String(), pxVal(left), pxVal(top), pxVal(w), pxVal(h)))
				e.sb.WriteString(fmt.Sprintf(
					"<div style=\"%s%sleft:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
					picCSS.String(), imgCSS, pxVal(left), pxVal(top), pxVal(w), pxVal(h)))
			} else {
				picClass := e.css.Register(picCSS.String())
				imgClass := e.css.Register(imgCSS)
				e.sb.WriteString(fmt.Sprintf(
					"<div class=\"%s\" style=\"left:%spx;top:%spx;width:%spx;height:%spx;border:none;\">&nbsp;</div>\n",
					picClass, pxVal(left), pxVal(top), pxVal(w), pxVal(h)))
				e.sb.WriteString(fmt.Sprintf(
					"<div class=\"%s %s\" style=\"left:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
					picClass, imgClass, pxVal(left), pxVal(top), pxVal(w), pxVal(h)))
			}
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
		case 0: // checkmark
			e.sb.WriteString(fmt.Sprintf(
				`<polyline points="%.2f,%.2f %.2f,%.2f %.2f,%.2f" stroke="%s" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/>`,
				w*0.15, h*0.5,
				w*0.4, h*0.75,
				w*0.85, h*0.25,
				strokeColor,
			))
		case 1: // cross
			e.sb.WriteString(fmt.Sprintf(
				`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-width="2"/>`,
				w*0.2, h*0.2, w*0.8, h*0.8, strokeColor,
			))
			e.sb.WriteString(fmt.Sprintf(
				`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-width="2"/>`,
				w*0.8, h*0.2, w*0.2, h*0.8, strokeColor,
			))
		case 2: // plus
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
				e.sb.WriteString(fmt.Sprintf(`<div %s>`, styleAttr("")))
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

// renderTextObject renders a text object matching C# HTMLExportLayers output.
// Outer div: class with font+alignment+color+background+border+size, inline style with position.
// Inner div: class with display:block;border:0;width;padding;margin-top.
func (e *Exporter) renderTextObject(obj preview.PreparedObject, scale float32) {
	left := obj.Left * scale
	top := obj.Top * scale
	w := obj.Width * scale
	h := obj.Height * scale

	font := obj.Font
	// Convert font size from pt to px (96dpi / 72pt = 1.3333...) to match C# output.
	fontPx := font.Size * 96.0 / 72.0

	// ── Build outer CSS class (matches C# GetStyle + GetStyleFromObject) ──
	var outerCSS strings.Builder

	// Font style (C# HTMLFontStyle: bold, italic/normal, underline+strikeout, family, size, line-height)
	if font.Style&style.FontStyleBold != 0 {
		outerCSS.WriteString("font-weight:bold;")
	}
	if font.Style&style.FontStyleItalic != 0 {
		outerCSS.WriteString("font-style:italic;")
	} else {
		outerCSS.WriteString("font-style:normal;")
	}
	if font.Style&style.FontStyleUnderline != 0 && font.Style&style.FontStyleStrikeout != 0 {
		outerCSS.WriteString("text-decoration:underline|line-through;")
	} else if font.Style&style.FontStyleUnderline != 0 {
		outerCSS.WriteString("text-decoration:underline;")
	} else if font.Style&style.FontStyleStrikeout != 0 {
		outerCSS.WriteString("text-decoration:line-through;")
	}
	// C# uses unquoted font-family and "line-height: " (space after colon).
	outerCSS.WriteString(fmt.Sprintf("font-family:%s;font-size:%spx;", font.Name, pxVal(fontPx)))

	// Line height: C# computes font.FontFamily.GetLineSpacing(style) / GetEmHeight(style),
	// then rounds to 2 decimals. If the object has an explicit LineHeight > 0, use that.
	if obj.LineHeight > 0 {
		outerCSS.WriteString(fmt.Sprintf("line-height: %spx;", pxVal(obj.LineHeight)))
	} else {
		outerCSS.WriteString(fmt.Sprintf("line-height: %s;", fontLineHeightRatio(font.Name)))
	}

	// text-align (C# GetStyle: uses RTL-aware alignment).
	outerCSS.WriteString("text-align:")
	switch obj.HorzAlign {
	case 1:
		outerCSS.WriteString("center")
	case 2:
		if obj.RTL {
			outerCSS.WriteString("left")
		} else {
			outerCSS.WriteString("right")
		}
	case 3:
		outerCSS.WriteString("justify")
	default:
		if obj.RTL {
			outerCSS.WriteString("right")
		} else {
			outerCSS.WriteString("left")
		}
	}
	outerCSS.WriteString(";")

	// word-wrap and overflow (C# GetStyle).
	if obj.WordWrap {
		outerCSS.WriteString("word-wrap:break-word;")
	}
	if obj.Clip {
		outerCSS.WriteString("overflow:hidden;")
	}

	// position:absolute, color, background-color, RTL (C# GetStyle).
	outerCSS.WriteString("position:absolute;")
	tc := obj.TextColor
	outerCSS.WriteString(fmt.Sprintf("color:%s;", rgbColor(tc)))
	fc := obj.FillColor
	if fc.A > 0 {
		outerCSS.WriteString(fmt.Sprintf("background-color:%s;", rgbColor(fc)))
	} else {
		outerCSS.WriteString("background-color:transparent;")
	}
	if obj.RTL {
		outerCSS.WriteString("direction:rtl;")
	}

	// Border (C# HTMLBorder).
	borderStr := borderCSS(&obj.Border, scale)
	if borderStr != "" {
		outerCSS.WriteString(borderStr)
	} else {
		outerCSS.WriteString("border:none;")
	}

	// Width and height (C# GetStyle appends width/height).
	outerCSS.WriteString(fmt.Sprintf("width:%spx;height:%spx;", pxVal(w), pxVal(h)))

	// ── Build inner CSS class FIRST (matches C# flow: GetSpanText is called
	// before LayerBack→GetStyle, so inner class is registered first) ──
	var innerCSS strings.Builder

	// Compute inner width: object width minus horizontal padding (C# pattern).
	padH := (obj.PaddingLeft + obj.PaddingRight) * scale
	innerW := w - padH

	innerCSS.WriteString(fmt.Sprintf("display:block;border:0;width:%spx;", pxVal(innerW)))

	// Paragraph offset / text-indent.
	if obj.ParagraphOffset != 0 {
		innerCSS.WriteString(fmt.Sprintf("text-indent:%spx;", pxVal(obj.ParagraphOffset*scale)))
	}
	// Padding (C# GetSpanText: only emits non-zero values).
	if obj.PaddingLeft != 0 {
		innerCSS.WriteString(fmt.Sprintf("padding-left:%spx;", pxVal(obj.PaddingLeft*scale)))
	}
	if obj.PaddingRight != 0 {
		innerCSS.WriteString(fmt.Sprintf("padding-right:%spx;", pxVal(obj.PaddingRight*scale)))
	}
	if obj.PaddingTop != 0 {
		innerCSS.WriteString(fmt.Sprintf("padding-top:%spx;", pxVal(obj.PaddingTop*scale)))
	}

	// Vertical alignment: C# uses margin-top for non-top alignment.
	// For vert-center/bottom, C# computes top offset via text height measurement.
	// Use the MeasureString-calibrated ratio (from textmeasure.go), NOT the CSS
	// line-height ratio, because C# computes margin from rendered text height.
	lhRatio := fontMeasureRatioFloat(font.Name)
	switch obj.VertAlign {
	case 1: // center
		lineH := float64(fontPx) * lhRatio
		textH := lineH // single line estimate
		marginTop := (float64(h) - textH - float64(obj.PaddingBottom*scale) + float64(obj.PaddingTop*scale)) / 2
		if marginTop > 0 {
			innerCSS.WriteString(fmt.Sprintf("margin-top:%spx;", pxVal(float32(marginTop))))
		}
	case 2: // bottom
		lineH := float64(fontPx) * lhRatio
		textH := lineH
		marginTop := float64(h) - textH - float64(obj.PaddingBottom*scale)
		if marginTop > 0 {
			innerCSS.WriteString(fmt.Sprintf("margin-top:%spx;", pxVal(float32(marginTop))))
		}
	}

	// No-wrap for non-wrapping text objects (C# GetSpanText).
	if !obj.WordWrap {
		innerCSS.WriteString("overflow: hidden; text-wrap: nowrap;")
	}

	// In class mode, register CSS classes (C# flow: inner registered first via
	// GetSpanText, then outer via LayerBack→GetStyle).
	// In InlineStyles mode, classes are empty strings; CSS is emitted inline.
	innerClass := e.cssClass(innerCSS.String())
	outerClass := e.cssClass(outerCSS.String())

	// ── Format text content ──
	// C# HtmlString converts line breaks to <p> tags and double-spaces to &nbsp;&nbsp;.
	innerText := formatTextContent(obj.Text, fontPx)

	// ── Compute border width adjustments (C# Layer method) ──
	var bLeft, bTop, bRight, bBottom float32
	htmlBorderWidthValues(&obj.Border, scale, &bLeft, &bTop, &bRight, &bBottom)

	// Outer div inline style: position (C# Layer method with border adjustments).
	adjLeft := left - bLeft/2
	adjTop := top - bTop/2
	adjW := w - bRight/2 - bLeft/2
	adjH := h - bBottom/2 - bTop/2

	outerStyle := fmt.Sprintf("left:%spx;top:%spx;width:%spx;height:%spx;",
		pxVal(adjLeft), pxVal(adjTop), pxVal(adjW), pxVal(adjH))

	// Hyperlink wrapping (C# GetHref + Layer pattern).
	// Go HyperlinkKind: 0=None, 1=URL, 2=PageNumber, 3=Bookmark.
	// C# SinglePage mode: Bookmark → #bookmark, PageNumber → #PageN{n}.
	href := hyperlinkHref(obj.HyperlinkKind, obj.HyperlinkValue)
	if href != "" {
		// C# GetHref: <a style="color:{textColor}[;text-decoration:none]" href="{value}"[target="_blank"]>
		// target attribute is taken from HyperlinkTarget (C# Hyperlink.Target / OpenLinkInNewTab).
		linkColor := rgbColor(tc)
		hrefStyle := fmt.Sprintf("color:%s", linkColor)
		if font.Style&style.FontStyleUnderline == 0 {
			hrefStyle += ";text-decoration:none"
		}
		aTag := fmt.Sprintf(`<a style="%s" href="%s"`, hrefStyle, href)
		if obj.HyperlinkTarget != "" {
			aTag += fmt.Sprintf(` target="%s"`, export.HTMLString(obj.HyperlinkTarget))
		}
		aTag += ">"
		e.sb.WriteString(aTag)
		if e.InlineStyles {
			// InlineStyles: merge outerCSS + position/size into inline style; no class attr.
			e.sb.WriteString(fmt.Sprintf(`<div style="cursor:pointer;%s%s">`, outerCSS.String(), outerStyle))
			e.sb.WriteString(fmt.Sprintf(`<div style="%s">%s</div>`, innerCSS.String(), innerText))
		} else {
			e.sb.WriteString(fmt.Sprintf(`<div class="%s" style="cursor:pointer;%s">`, outerClass, outerStyle))
			e.sb.WriteString(fmt.Sprintf(`<div class="%s">%s</div>`, innerClass, innerText))
		}
		e.sb.WriteString("</div>\n</a>")
	} else {
		if e.InlineStyles {
			// InlineStyles: merge outerCSS + position/size into inline style; no class attr.
			e.sb.WriteString(fmt.Sprintf(`<div style="%s%s">`, outerCSS.String(), outerStyle))
			e.sb.WriteString(fmt.Sprintf(`<div style="%s">%s</div>`, innerCSS.String(), innerText))
		} else {
			e.sb.WriteString(fmt.Sprintf(`<div class="%s" style="%s">`, outerClass, outerStyle))
			e.sb.WriteString(fmt.Sprintf(`<div class="%s">%s</div>`, innerClass, innerText))
		}
		e.sb.WriteString("</div>\n")
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
	e.sb.WriteString("</div>") // close frpage{n}

	// For UnlimitedHeight pages, expand the page div height to fit all content.
	// C# uses page.UnlimitedHeightValue for this.
	if e.maxBottom > e.pageH {
		oldH := pxVal(e.pageH)
		newH := pxVal(e.maxBottom)
		// Replace the height in the page div that was already written to sb.
		pageStr := e.sb.String()
		pageStr = strings.Replace(pageStr, "height:"+oldH+"px;", "height:"+newH+"px;", 1)
		e.sb.Reset()
		e.sb.WriteString(pageStr)
	}

	// Capture the page content and swap buffers back.
	pageContent := e.sb.String()
	e.sb, e.pageBuf = e.pageBuf, e.sb // restore: sb = main output, pageBuf = page content

	// C# per-page CSS emission: print CSS + new content CSS BEFORE the page content.
	pageNum := e.pageIdx - 1 // 0-based (pageIdx was incremented in ExportPageBegin)

	// In InlineStyles mode, skip the <style> class block — all CSS is inline.
	// C# reference: HTMLExportStyles.cs InlineStyles branch skips cssStyles emission.
	if e.EmbedCSS && e.css != nil && !e.InlineStyles {
		// Print CSS (one block per page).
		e.sb.WriteString(fmt.Sprintf("<style type=\"text/css\" media=\"print\"><!--\ndiv.frpage%d { break-after: always; page-break-inside: avoid; }\n @page { size: portrait; margin: 0; }--></style>\n", pageNum))
		// New content CSS classes for this page.
		if block := e.css.StyleBlockSince(e.cssCountBeforePage); block != "" {
			e.sb.WriteString(block)
		}
	}
	// C# opens frpage-container div AFTER the first page's CSS blocks.
	if pageNum == 0 {
		e.sb.WriteString("<div class=\"frpage-container\"\n>\n")
	}
	// Page content (anchor + page div + objects).
	e.sb.WriteString(pageContent)

	return nil
}

func (e *Exporter) Finish() error {
	// Build the final HTML document.
	// C# structure: head + body preamble, then per-page CSS + content (already in e.sb),
	// then frpage-container wrapper around the per-page content, then close body/html.
	var out strings.Builder

	out.WriteString("<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\">\n")
	out.WriteString("<html><head>\n")
	out.WriteString("<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\">\n")
	out.WriteString("<meta name=Generator content=\"FastReport http://www.fast-report.com\">\n")
	out.WriteString(fmt.Sprintf("<title>%s</title>\n", export.HTMLString(e.Title)))
	out.WriteString("</head>\r\n")
	out.WriteString("<body bgcolor=\"#FFFFFF\" text=\"#000000\">\r\n")

	// Per-page CSS + content is already assembled in e.sb by ExportPageEnd.
	// The frpage-container div wraps the per-page blocks.
	// In C#, the frpage-container div is opened inside the first page's output
	// and is NOT closed (C# doesn't close it).
	out.WriteString(e.sb.String())

	// C# uses AppendLine(BODY_END) which adds \r\n after </body>.
	out.WriteString("</body>\r\n</html>\n")

	// Write final result.
	_, err := io.WriteString(e.w, out.String())

	// Update sb to contain the full output for HTML() method.
	e.sb.Reset()
	e.sb.WriteString(out.String())

	return err
}

// buildPrintCSS is no longer used — print CSS is emitted per page in ExportPageEnd.

// HTML returns the complete HTML string that would be written.
// Useful for testing. Call after Export has been called.
func (e *Exporter) HTML() string {
	return e.sb.String()
}

// ── InlineStyles helpers ─────────────────────────────────────────────────────────

// cssAttr returns either a class="sN" attribute (normal mode) or an inline
// style="..." attribute (InlineStyles mode) for the given CSS content string.
// This matches C# GetStyle(string): InlineStyle(style) vs GetStyleTag(index).
// C# reference: HTMLExportStyles.cs lines 98-109.
func (e *Exporter) cssAttr(css string) string {
	if e.InlineStyles {
		// C# InlineStyle: return $"style=\"{style}\""
		return fmt.Sprintf(`style="%s"`, css)
	}
	name := e.css.Register(css)
	if name == "" {
		return ""
	}
	return fmt.Sprintf(`class="%s"`, name)
}

// cssClass returns only the class name string (no attribute wrapping), or ""
// in InlineStyles mode. Used when the caller needs to compose multiple classes.
func (e *Exporter) cssClass(css string) string {
	if e.InlineStyles {
		return ""
	}
	return e.css.Register(css)
}

// ── CSS helpers ─────────────────────────────────────────────────────────────────

// hyperlinkHref returns the href attribute value for a hyperlink, matching C# GetHref
// in SinglePage mode (the Go exporter always produces a single HTML document).
//
// Go HyperlinkKind values: 0=None, 1=URL, 2=PageNumber, 3=Bookmark.
// Returns "" when no href should be emitted.
func hyperlinkHref(kind int, value string) string {
	if value == "" {
		return ""
	}
	switch kind {
	case 1: // URL — direct href to external URL
		return value
	case 2: // PageNumber — link to page anchor within the same document
		// C# SinglePage: <a href="#PageN{value}">
		return "#PageN" + value
	case 3: // Bookmark — link to named anchor within the same document
		// C# SinglePage: <a href="#{value}">
		return "#" + value
	}
	return ""
}

// pxVal formats a float as a CSS pixel value string, matching C# ExportUtils.FloatToString.
// It drops trailing zeros: 718.20 → "718.2", 28.35 → "28.35", 0.00 → "0".
// fontLineHeightRatio returns the CSS line-height ratio for a given font family,
// matching C#'s FontFamily.GetLineSpacing(style) / GetEmHeight(style) rounded to
// 2 decimals. Values extracted from C# FastReport .NET HTML output.
func fontLineHeightRatio(fontFamily string) string {
	switch strings.ToLower(fontFamily) {
	case "arial", "arial narrow", "times new roman", "times":
		return "1.15"
	case "tahoma", "microsoft sans serif":
		return "1.21"
	case "verdana":
		return "1.22"
	case "arial unicode ms":
		return "1.34"
	case "arial black":
		return "1.41"
	case "georgia":
		return "1.14"
	case "courier new", "courier":
		return "1.13"
	case "segoe ui":
		return "1.33"
	default:
		return "1.21" // Tahoma is the most common default in FastReport
	}
}

// fontLineHeightRatioFloat returns the line-height ratio as a float64 for
// margin-top calculations. Uses the same values as fontLineHeightRatio.
func fontLineHeightRatioFloat(fontFamily string) float64 {
	switch strings.ToLower(fontFamily) {
	case "arial", "arial narrow", "times new roman", "times":
		return 1.15
	case "tahoma", "microsoft sans serif":
		return 1.21
	case "verdana":
		return 1.22
	case "arial unicode ms":
		return 1.34
	case "arial black":
		return 1.41
	case "georgia":
		return 1.14
	case "courier new", "courier":
		return 1.13
	case "segoe ui":
		return 1.33
	default:
		return 1.21
	}
}

// fontMeasureRatioFloat returns the MeasureString-calibrated line height ratio.
// This matches C#'s Graphics.MeasureString output for text height measurement
// (used for margin-top calculations). Slightly different from the CSS line-height
// ratio because MeasureString uses actual font ascent+descent.
// fontMeasureRatioFloat returns the exact .NET font line-spacing ratio for
// margin-top calculations. Uses full precision from font OS/2 table metrics:
// (usWinAscent + usWinDescent + typoLineGap) / unitsPerEm.
func fontMeasureRatioFloat(fontFamily string) float64 {
	switch strings.ToLower(fontFamily) {
	case "arial", "arial narrow":
		return 2355.0 / 2048.0
	case "times new roman", "times":
		return 2355.0 / 2048.0
	case "tahoma", "microsoft sans serif":
		return 2472.0 / 2048.0
	case "verdana":
		return 2489.0 / 2048.0
	case "arial unicode ms":
		return 2743.0 / 2048.0
	case "arial black":
		return 2899.0 / 2048.0
	case "georgia":
		return 2327.0 / 2048.0
	case "courier new", "courier":
		return 2320.0 / 2048.0
	case "segoe ui":
		return 2724.0 / 2048.0
	default:
		return 2472.0 / 2048.0
	}
}

func pxVal(v float32) string {
	s := fmt.Sprintf("%.2f", v)
	// Strip trailing zeros after decimal point.
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

// rgbColor returns "rgb(R, G, B)" for fully opaque colors and "rgba(R, G, B, A)" for semi-transparent ones.
func rgbColor(c color.RGBA) string {
	if c.A == 255 {
		return fmt.Sprintf("rgb(%d, %d, %d)", c.R, c.G, c.B)
	}
	return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", c.R, c.G, c.B, float32(c.A)/255.0)
}

// formatTextContent converts plain text to HTML, matching C# ExportUtils.HtmlString.
// Line breaks become <p> tags. Double spaces become &nbsp;&nbsp;.
// The fontPx parameter provides the font size in CSS pixels for the p tag height.
// formatTextContent converts text to HTML matching C# ExportUtils.HtmlString exactly.
//
// C# logic for line breaks (crlf == CRLF.html):
//   - Normalize \r\n and \r to \n
//   - First \n at position 0: <p style="margin-top:{fontSize}margin-bottom:0px"></p>
//   - First line break (lineBreakCount==0): <p style="margin-top:0px;margin-bottom:0px;"></p>
//   - Subsequent breaks (lineBreakCount>0): <p style="margin-top:0px;height:{fontSize}margin-bottom:0px"></p>
//   - Consecutive spaces → &nbsp;
//   - Trailing space → &nbsp;
//
// fontSize is Px(Math.Round(font.Size * 96 / 72)) e.g. "15px;" for 11pt.
func formatTextContent(text string, fontPx float32) string {
	// C# uses Px(Math.Round(obj.Font.Size * 96 / 72)) — rounds fontPx to integer first.
	fontSize := pxVal(float32(math.Round(float64(fontPx)))) + "px;"

	// Replace \v with \n (C# text.Replace('\v', '\n')).
	text = strings.ReplaceAll(text, "\v", "\n")
	// Normalize \r\n to \n (C# crlf==html: replace \r\n with \n, then \r with \n).
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	var result strings.Builder
	lineBreakCount := 0
	runes := []rune(text)
	n := len(runes)

	for i := 0; i < n; i++ {
		ch := runes[i]

		// Consecutive/trailing spaces → &nbsp; (C# logic).
		if ch == ' ' && (n == 1 ||
			(i < n-1 && runes[i+1] == ' ') ||
			(i > 0 && runes[i-1] == ' ') ||
			i == n-1) {
			result.WriteString("&nbsp;")
			continue
		}

		if ch == '\n' {
			// C#: if first char is \n, prepend margin-top with fontSize.
			if i == 0 {
				result.WriteString(fmt.Sprintf(`<p style="margin-top:%smargin-bottom:0px"></p>`, fontSize))
			}
			if lineBreakCount == 0 {
				result.WriteString(`<p style="margin-top:0px;margin-bottom:0px;"></p>`)
			} else {
				result.WriteString(fmt.Sprintf(`<p style="margin-top:0px;height:%smargin-bottom:0px"></p>`, fontSize))
			}
			lineBreakCount++
			continue
		}

		// C# resets lineBreakCount on any non-break, non-space character.
		lineBreakCount = 0

		// HTML-escape special chars.
		switch ch {
		case '<':
			result.WriteString("&lt;")
		case '>':
			result.WriteString("&gt;")
		case '&':
			result.WriteString("&amp;")
		case '"':
			result.WriteString("&quot;")
		default:
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// htmlBorderWidthValues computes the border widths for each side, matching C# HTMLBorderWidth.
func htmlBorderWidthValues(b *style.Border, scale float32, left, top, right, bottom *float32) {
	if b == nil || b.VisibleLines == style.BorderLinesNone {
		return
	}
	type side struct {
		flag style.BorderLines
		idx  int
		out  *float32
	}
	sides := []side{
		{style.BorderLinesLeft, int(style.BorderLeft), left},
		{style.BorderLinesTop, int(style.BorderTop), top},
		{style.BorderLinesRight, int(style.BorderRight), right},
		{style.BorderLinesBottom, int(style.BorderBottom), bottom},
	}
	for _, s := range sides {
		if b.VisibleLines&s.flag == 0 {
			continue
		}
		line := b.Lines[s.idx]
		w := float32(1) * scale
		if line != nil {
			if line.Style == style.LineStyleDouble {
				w = line.Width * 3 * scale
			} else {
				w = line.Width * scale
			}
		}
		*s.out = w
	}
}

// borderCSS converts a style.Border into CSS border/box-shadow declarations.
// It handles per-side borders using border-top/right/bottom/left shorthand.
// The format matches C# HTMLBorder output: separate width/color/style properties per side.
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
			// C# HTMLBorder: emit explicit none for invisible sides so they are
			// not affected by inherited or default browser border styles.
			sb.WriteString(s.prop + ":none;")
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

// resizeImagePNG resizes a PNG image to the target dimensions, matching C#'s
// GetLayerPicture which re-renders images at Width*Zoom × Height*Zoom.
// If the image is already the target size or decoding fails, returns the original data.
// resizeImagePNG resizes a PNG image to the target dimensions, matching C#'s
// GetLayerPicture which creates a Bitmap at Width*Zoom × Height*Zoom and
// calls obj.Draw(). PictureObject default SizeMode is Zoom (preserve aspect
// ratio, center within bounds, transparent background).
func resizeImagePNG(data []byte, targetW, targetH int) []byte {
	if targetW <= 0 || targetH <= 0 {
		return data
	}
	// Check PNG magic.
	if len(data) < 8 || data[0] != 0x89 || data[1] != 'P' || data[2] != 'N' || data[3] != 'G' {
		return data // not PNG, return as-is
	}

	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return data
	}
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	if srcW == targetW && srcH == targetH {
		return data // already correct size
	}

	// Create target bitmap with transparent background (matching C# Bitmap + clear).
	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))

	// C# PictureObject.SizeMode defaults to Zoom: scale to fit while preserving
	// aspect ratio, then center within the target bounds.
	scaleX := float64(targetW) / float64(srcW)
	scaleY := float64(targetH) / float64(srcH)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}
	drawW := int(math.Round(float64(srcW) * scale))
	drawH := int(math.Round(float64(srcH) * scale))
	offsetX := (targetW - drawW) / 2
	offsetY := (targetH - drawH) / 2

	// Draw scaled image centered in the target bitmap using bicubic interpolation
	// (matching C#'s HighQualityBicubic / InterpolationMode).
	drawRect := image.Rect(offsetX, offsetY, offsetX+drawW, offsetY+drawH)
	xdraw.CatmullRom.Scale(dst, drawRect, src, srcBounds, xdraw.Over, nil)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return data
	}
	return buf.Bytes()
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

// Ensure math is used (for future use in vertical alignment calculations).
var _ = math.Round
