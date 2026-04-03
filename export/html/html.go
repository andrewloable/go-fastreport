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
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"strings"

	"github.com/andrewloable/go-fastreport/export"
	imgexp "github.com/andrewloable/go-fastreport/export/image"
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

// renderWatermarkText rasterizes the watermark text onto a transparent image
// and emits it using the LayerBack + LayerPicture two-div pattern, matching
// C# HTMLExportLayers.Watermark() which creates a PictureObject with DrawText
// then calls LayerBack (fill/border container) and LayerPicture (image overlay).
func (e *Exporter) renderWatermarkText(wm *preview.PreparedWatermark, pageW, pageH float32) {
	if wm.Text == "" {
		return
	}

	w := int(pageW)
	h := int(pageH)
	if w < 1 || h < 1 {
		return
	}

	img := utils.RenderWatermarkText(wm.Text, wm.Font, wm.TextColor, int(wm.TextRotation), w, h)
	if img == nil {
		return
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return
	}
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Build the LayerBack CSS (same as C# GetStyle for a transparent PictureObject).
	picCSS := fmt.Sprintf(
		"text-align:center;position:absolute;color:rgb(255, 255, 255);background-color:transparent;border:none;width:%spx;height:%spx;",
		pxVal(pageW), pxVal(pageH))
	// Build the LayerPicture CSS (base64 background image).
	imgCSS := fmt.Sprintf(
		"background: url('data:image/Png;base64,%s') no-repeat !important;-webkit-print-color-adjust:exact;",
		b64)

	// Emit two divs: LayerBack (fill container) + LayerPicture (image overlay).
	if e.InlineStyles {
		e.sb.WriteString(fmt.Sprintf(
			"<div style=\"%sleft:0px;top:0px;width:%spx;height:%spx;border:none;\">&nbsp;</div>\n",
			picCSS, pxVal(pageW), pxVal(pageH)))
		e.sb.WriteString(fmt.Sprintf(
			"<div style=\"%s%sleft:0px;top:0px;width:%spx;height:%spx;\">&nbsp;</div>\n",
			picCSS, imgCSS, pxVal(pageW), pxVal(pageH)))
	} else {
		picClass := e.css.Register(picCSS)
		imgClass := e.css.Register(imgCSS)
		e.sb.WriteString(fmt.Sprintf(
			"<div class=\"%s\" style=\"left:0px;top:0px;width:%spx;height:%spx;border:none;\">&nbsp;</div>\n",
			picClass, pxVal(pageW), pxVal(pageH)))
		e.sb.WriteString(fmt.Sprintf(
			"<div class=\"%s %s\" style=\"left:0px;top:0px;width:%spx;height:%spx;\">&nbsp;</div>\n",
			picClass, imgClass, pxVal(pageW), pxVal(pageH)))
	}
}

// renderWatermarkImage emits a positioned div with a base64-encoded background image.
// C# always calls Watermark(page, false) which creates a PictureObject with a
// transparent bitmap even if no image data exists. To match, we render a
// transparent PNG when ImageBlobIdx < 0 or blob data is empty.
func (e *Exporter) renderWatermarkImage(wm *preview.PreparedWatermark, pageW, pageH float32) {
	var b64 string
	var hasImage bool
	if wm.ImageBlobIdx >= 0 && e.pp != nil {
		imgData := e.pp.BlobStore.Get(wm.ImageBlobIdx)
		if len(imgData) > 0 {
			b64 = base64.StdEncoding.EncodeToString(imgData)
			hasImage = true
		}
	}
	if !hasImage {
		// C# renders a transparent bitmap even with no image data.
		w, h := int(pageW), int(pageH)
		if w < 1 || h < 1 {
			return
		}
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return
		}
		b64 = base64.StdEncoding.EncodeToString(buf.Bytes())
	}
	// Build LayerBack CSS (same as watermark text — transparent PictureObject).
	picCSS := fmt.Sprintf(
		"text-align:center;position:absolute;color:rgb(255, 255, 255);background-color:transparent;border:none;width:%spx;height:%spx;",
		pxVal(pageW), pxVal(pageH))
	// Build LayerPicture CSS with background image.
	imgCSS := fmt.Sprintf(
		"background: url('data:image/Png;base64,%s') no-repeat !important;-webkit-print-color-adjust:exact;",
		b64)

	// Emit LayerBack + LayerPicture divs.
	if e.InlineStyles {
		e.sb.WriteString(fmt.Sprintf(
			"<div style=\"%sleft:0px;top:0px;width:%spx;height:%spx;border:none;\">&nbsp;</div>\n",
			picCSS, pxVal(pageW), pxVal(pageH)))
		e.sb.WriteString(fmt.Sprintf(
			"<div style=\"%s%sleft:0px;top:0px;width:%spx;height:%spx;\">&nbsp;</div>\n",
			picCSS, imgCSS, pxVal(pageW), pxVal(pageH)))
	} else {
		picClass := e.css.Register(picCSS)
		imgClass := e.css.Register(imgCSS)
		e.sb.WriteString(fmt.Sprintf(
			"<div class=\"%s\" style=\"left:0px;top:0px;width:%spx;height:%spx;border:none;\">&nbsp;</div>\n",
			picClass, pxVal(pageW), pxVal(pageH)))
		e.sb.WriteString(fmt.Sprintf(
			"<div class=\"%s %s\" style=\"left:0px;top:0px;width:%spx;height:%spx;\">&nbsp;</div>\n",
			picClass, imgClass, pxVal(pageW), pxVal(pageH)))
	}
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
	// NoBackground bands (e.g. horizontal-split matrix continuations) skip this div.
	if b.Width > 0 && b.Height > 0 && !b.NoBackground {
		e.renderBandBackground(b, scale)
	}

	if e.Layers {
		// In layers mode, render each object directly onto the page with
		// absolute coordinates (band.Top + obj.Top) and ascending z-index.
		bandTop := b.Top
		for _, obj := range b.Objects {
			// Skip non-exportable objects. Mirrors C# HTMLExportLayers.cs line 967:
			// "if (c is ReportComponentBase obj && obj.Exportable)".
			if obj.NotExportable {
				continue
			}
			layered := obj
			layered.Top += bandTop
			e.renderObjectLayered(layered, scale)
		}
		return nil
	}

	// Flat rendering: render each object with page-absolute top coordinate (no band wrappers).
	for _, obj := range b.Objects {
		// Skip non-exportable objects (same check as layers mode above).
		if obj.NotExportable {
			continue
		}
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
	// Use BackgroundHeight if set (table/matrix bands where the section
	// background handles the table area separately).
	bgH := b.Height
	if b.BackgroundHeight > 0 {
		bgH = b.BackgroundHeight
	}
	h := bgH * scale
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
		// Add background image class for non-solid fills (GlassFill, etc.).
		// C# ref: HTMLExportLayers.cs LayerPicture → GetStyleTag(index1, index2)
		// emits dual-class "s{i1} s{i2}" when an additional background CSS is present.
		if b.BackgroundCSS != "" {
			bgClass := e.css.Register(b.BackgroundCSS)
			className = className + " " + bgClass
		}
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
				// Apply border width adjustment (C# Layer method): position shifts by
				// -borderWidth/2 and size shrinks by borderWidth. Mirrors renderTextObject.
				var bLeft, bTop, bRight, bBottom float32
				htmlBorderWidthValues(&obj.Border, scale, &bLeft, &bTop, &bRight, &bBottom)
				adjLeft := left - bLeft/2
				adjTop := top - bTop/2
				adjW := w - bRight/2 - bLeft/2
				adjH := h - bBottom/2 - bTop/2

				// C# LayerBack flow: GetStyle(obj) registers the picture's non-text
				// style class FIRST, then LayerPicture registers the image data class.
				// C# GetStyle for non-text: color:white; background-color:...; actual border CSS; width/height (unadjusted)
				// The border:none override goes in the LayerBack div's inline style, not in the CSS class.
				// The adjusted position/size go in each div's inline style.
				var picCSS strings.Builder
				picCSS.WriteString("text-align:center;position:absolute;color:rgb(255, 255, 255);")
				if obj.FillColor.A == 0 {
					picCSS.WriteString("background-color:transparent;")
				} else {
					picCSS.WriteString(fmt.Sprintf("background-color:%s;", rgbColor(obj.FillColor)))
				}
				// Include actual border CSS in the shared class (matching C# GetStyle).
				// If no border, emit border:none explicitly (C# always includes border in GetStyle).
				// Width/height in the class use unadjusted values; inline styles override with adjusted values.
				if bs := borderCSS(&obj.Border, scale); bs != "" {
					picCSS.WriteString(bs)
				} else {
					picCSS.WriteString("border:none;")
				}
				picCSS.WriteString(fmt.Sprintf("width:%spx;height:%spx;", pxVal(w), pxVal(h)))
				// picClass and imgClass are assigned below depending on InlineStyles mode.

				// Resize image to display dimensions for correct rendering in HTML.
				// For rotated objects, apply rotation to match C# GetLayerPicture.
				// Use PictureAngle (not Angle) — Angle is for text objects.
				var imgData []byte
				targetW := int(math.Round(float64(adjW)))
				targetH := int(math.Round(float64(adjH)))
				if obj.PictureAngle != 0 {
					imgData = imgexp.RotateImagePNG(data, obj.PictureAngle, targetW, targetH)
				} else {
					imgData = resizeImagePNG(data, targetW, targetH, obj.PictureSizeMode)
				}
				mime := imageMIMEForCSS(imgData)
				encoded := base64.StdEncoding.EncodeToString(imgData)
				imgCSS := fmt.Sprintf(
					"background: url('data:%s;base64,%s') no-repeat !important;-webkit-print-color-adjust:exact;",
					mime, encoded,
				)
				// Two overlapping divs: background/border div + image overlay div (C# pattern).
			// In InlineStyles mode, merge all CSS into inline style= attributes.
			// C# InlineStyle(style) → style="...{shared}left:...;top:..." pattern.
			// C# Layer() uses AppendLine("</div>") which adds \n after each div.
			if e.InlineStyles {
				// Inline mode: no CSS classes; emit style attributes directly.
				e.sb.WriteString(fmt.Sprintf(
					"<div style=\"%sleft:%spx;top:%spx;width:%spx;height:%spx;border:none;\">&nbsp;</div>\n",
					picCSS.String(), pxVal(adjLeft), pxVal(adjTop), pxVal(adjW), pxVal(adjH)))
				e.sb.WriteString(fmt.Sprintf(
					"<div style=\"%s%sleft:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
					picCSS.String(), imgCSS, pxVal(adjLeft), pxVal(adjTop), pxVal(adjW), pxVal(adjH)))
			} else {
				picClass := e.css.Register(picCSS.String())
				imgClass := e.css.Register(imgCSS)
				e.sb.WriteString(fmt.Sprintf(
					"<div class=\"%s\" style=\"left:%spx;top:%spx;width:%spx;height:%spx;border:none;\">&nbsp;</div>\n",
					picClass, pxVal(adjLeft), pxVal(adjTop), pxVal(adjW), pxVal(adjH)))
				e.sb.WriteString(fmt.Sprintf(
					"<div class=\"%s %s\" style=\"left:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
					picClass, imgClass, pxVal(adjLeft), pxVal(adjTop), pxVal(adjW), pxVal(adjH)))
			}
			break
			}
		}
		// No image data — render as a positioned background div (section background).
		// C# renders these as LayerBack elements: text-align:center, color:white,
		// background-color:transparent, border:none, with absolute positioning.
		var bgCSS strings.Builder
		bgCSS.WriteString("text-align:center;position:absolute;color:rgb(255, 255, 255);")
		if obj.FillColor.A == 0 {
			bgCSS.WriteString("background-color:transparent;")
		} else {
			bgCSS.WriteString(fmt.Sprintf("background-color:%s;", rgbColor(obj.FillColor)))
		}
		bgCSS.WriteString("border:none;")
		bgCSS.WriteString(fmt.Sprintf("width:%spx;height:%spx;", pxVal(w), pxVal(h)))

		if e.InlineStyles {
			e.sb.WriteString(fmt.Sprintf(
				"<div style=\"%sleft:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
				bgCSS.String(), pxVal(left), pxVal(top), pxVal(w), pxVal(h)))
		} else {
			bgClass := e.css.Register(bgCSS.String())
			e.sb.WriteString(fmt.Sprintf(
				"<div class=\"%s\" style=\"left:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
				bgClass, pxVal(left), pxVal(top), pxVal(w), pxVal(h)))
		}

	case preview.ObjectTypeLine:
		// C# converts lines with caps to PictureObjects via GetConvertedObjects()
		// which expands the bounding box. For lines without caps, C# uses LayerPicture
		// which adjusts position for negative Width/Height.
		// RenderGenericPNG returns the actual canvas size (including cap extension).
		if w < 0 {
			left += w
		}
		if h < 0 {
			top += h
		}
		// Pre-render the line PNG to get the actual canvas dimensions (may be
		// larger than abs(w)×abs(h) due to cap decorations).
		linePNG, canvasW, canvasH, lineErr := imgexp.RenderGenericPNG(obj)
		if lineErr == nil && len(linePNG) > 0 {
			// Use canvas dimensions for the CSS div so caps are not clipped.
			cssW := float32(canvasW)
			cssH := float32(canvasH)
			// Adjust position: the cap padding shifts the line within the canvas,
			// so shift the div left/up by the same amount.
			absW := w
			if absW < 0 {
				absW = -absW
			}
			absH := h
			if absH < 0 {
				absH = -absH
			}
			padL := (cssW - absW) / 2
			padT := (cssH - absH) / 2
			// For lines without caps, the canvas may be one pixel taller/wider than
			// the FRX dimensions due to ceiling rounding (e.g. 160.65 → 161px canvas).
			// C# uses the exact FRX dimensions for the CSS div, not the canvas size.
			// background-size:100% 100% handles any sub-pixel PNG scaling.
			// C# ref: HTMLExportLayers.cs LayerPicture (no GetConvertedObjects path).
			hasCaps := obj.LineStartCap.Style != preview.LineCapStyleNone ||
				obj.LineEndCap.Style != preview.LineCapStyleNone
			if !hasCaps {
				// No caps: use FRX dimensions for div; no position adjustment needed.
				e.renderObjectAsImageWithPNG(obj, left, top, absW, absH, scale, linePNG)
			} else {
				left -= padL
				top -= padT
				e.renderObjectAsImageWithPNG(obj, left, top, cssW, cssH, scale, linePNG)
			}
		}

	case preview.ObjectTypeShape:
		// C# renders all ShapeObjects as inline-style divs with base64 PNG images.
		// No CSS classes are used for shapes.
		e.renderObjectAsInlineImage(obj, left, top, w, h)

	case preview.ObjectTypeDigitalSignature:
		// Render a dashed-border placeholder box for digital signature fields.
		label := obj.Text
		if label == "" {
			label = "Digital Signature"
		}
		sigExtra := "border:2px dashed #888;color:#888;font-size:10pt;display:flex;align-items:center;justify-content:center;"
		e.sb.WriteString(fmt.Sprintf(`<div %s><span>%s</span></div>`, styleAttr(sigExtra), export.HTMLString(label)))

	case preview.ObjectTypeCheckBox:
		// C# renders checkboxes as PNG images using the LayerBack + LayerPicture
		// div pair pattern (same as other picture objects), with border adjustments.
		pngBytes, _, _, err := imgexp.RenderGenericPNG(obj)
		if err == nil && len(pngBytes) > 0 {
			// Apply border width adjustment (mirrors ObjectTypePicture rendering).
			var bLeft, bTop, bRight, bBottom float32
			htmlBorderWidthValues(&obj.Border, scale, &bLeft, &bTop, &bRight, &bBottom)
			adjLeft := left - bLeft/2
			adjTop := top - bTop/2
			adjW := w - bRight/2 - bLeft/2
			adjH := h - bBottom/2 - bTop/2

			// Build CSS matching C# GetStyleFromObject for non-text objects:
			// GetStyle(null, Color.White, obj.FillColor, false, HorzAlign.Center, ...)
			// Reference: HTMLExportStyles.cs:31-32
			var picCSS strings.Builder
			picCSS.WriteString("text-align:center;position:absolute;color:rgb(255, 255, 255);")
			if obj.FillColor.A == 0 {
				picCSS.WriteString("background-color:transparent;")
			} else {
				picCSS.WriteString(fmt.Sprintf("background-color:%s;", rgbColor(obj.FillColor)))
			}
			if bs := borderCSS(&obj.Border, scale); bs != "" {
				picCSS.WriteString(bs)
			} else {
				picCSS.WriteString("border:none;")
			}
			picCSS.WriteString(fmt.Sprintf("width:%spx;height:%spx;", pxVal(w), pxVal(h)))

			encoded := base64.StdEncoding.EncodeToString(pngBytes)
			imgCSS := fmt.Sprintf(
				"background: url('data:image/Png;base64,%s') no-repeat !important;-webkit-print-color-adjust:exact;",
				encoded,
			)

			if e.InlineStyles {
				e.sb.WriteString(fmt.Sprintf(
					"<div style=\"%sleft:%spx;top:%spx;width:%spx;height:%spx;border:none;\">&nbsp;</div>\n",
					picCSS.String(), pxVal(adjLeft), pxVal(adjTop), pxVal(adjW), pxVal(adjH)))
				e.sb.WriteString(fmt.Sprintf(
					"<div style=\"%s%sleft:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
					picCSS.String(), imgCSS, pxVal(adjLeft), pxVal(adjTop), pxVal(adjW), pxVal(adjH)))
			} else {
				picClass := e.css.Register(picCSS.String())
				imgClass := e.css.Register(imgCSS)
				e.sb.WriteString(fmt.Sprintf(
					"<div class=\"%s\" style=\"left:%spx;top:%spx;width:%spx;height:%spx;border:none;\">&nbsp;</div>\n",
					picClass, pxVal(adjLeft), pxVal(adjTop), pxVal(adjW), pxVal(adjH)))
				e.sb.WriteString(fmt.Sprintf(
					"<div class=\"%s %s\" style=\"left:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
					picClass, imgClass, pxVal(adjLeft), pxVal(adjTop), pxVal(adjW), pxVal(adjH)))
			}
		}

	case preview.ObjectTypePolyLine, preview.ObjectTypePolygon:
		// Render as base64 PNG image matching C# LayerBack + LayerPicture pattern.
		e.renderObjectAsImage(obj, left, top, w, h, scale)

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

	case preview.ObjectTypeBandBackground:
		// Band background for subreport bands merged in PrintOnParent mode.
		// Renders with the same CSS pattern as renderBandBackground so that the
		// CSS class is identical to (and shares a style slot with) the band's own
		// background div in normal rendering.
		// C# ref: HTMLExportLayers.cs ExportBandLayers → LayerBack(htmlPage, band, null).
		var bgCSS strings.Builder
		bgCSS.WriteString("text-align:center;position:absolute;color:rgb(255, 255, 255);")
		if obj.FillColor.A == 0 {
			bgCSS.WriteString("background-color:transparent;")
		} else {
			bgCSS.WriteString(fmt.Sprintf("background-color:%s;", rgbColor(obj.FillColor)))
		}
		if bs := borderCSS(&obj.Border, scale); bs != "" {
			bgCSS.WriteString(bs)
		} else {
			bgCSS.WriteString("border:none;")
		}
		bgCSS.WriteString(fmt.Sprintf("width:%spx;height:%spx;", pxVal(w), pxVal(h)))
		if e.InlineStyles {
			e.sb.WriteString(fmt.Sprintf(
				"<div style=\"%sleft:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
				bgCSS.String(), pxVal(left), pxVal(top), pxVal(w), pxVal(h),
			))
		} else {
			className := e.css.Register(bgCSS.String())
			e.sb.WriteString(fmt.Sprintf(
				"<div class=\"%s\" style=\"left:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
				className, pxVal(left), pxVal(top), pxVal(w), pxVal(h),
			))
		}

	default:
		// Unknown/unhandled type — render an empty placeholder.
		e.sb.WriteString(fmt.Sprintf(`<div %s></div>`, styleAttr("")))
	}
}

// renderTextObject renders a text object matching C# HTMLExportLayers output.
// Outer div: class with font+alignment+color+background+border+size, inline style with position.
// Inner div: class with display:block;border:0;width;padding;margin-top.
// renderObjectAsImage renders a PreparedObject (polygon, shape, etc.) as a
// single div with a base64 PNG background image, matching C# HTMLExportLayers
// pattern for non-text objects (one div, two CSS classes: styling + image).
func (e *Exporter) renderObjectAsImage(obj preview.PreparedObject, left, top, w, h, scale float32) {
	pngBytes, _, _, err := imgexp.RenderGenericPNG(obj)
	if err != nil || len(pngBytes) == 0 {
		return
	}

	var picCSS strings.Builder
	picCSS.WriteString("text-align:center;position:absolute;color:rgb(255, 255, 255);")
	if obj.FillColor.A == 0 {
		picCSS.WriteString("background-color:transparent;")
	} else {
		picCSS.WriteString(fmt.Sprintf("background-color:%s;", rgbColor(obj.FillColor)))
	}
	// C# HTMLExportStyles.GetStyle uses Math.Abs(Width) and Math.Abs(Height)
	// for the CSS class. The inline style keeps the raw (possibly negative) values.
	// When the browser encounters an invalid negative inline height, it falls back
	// to the positive CSS class value, making the div render at the correct size.
	absW := w
	if absW < 0 {
		absW = -absW
	}
	absH := h
	if absH < 0 {
		absH = -absH
	}
	picCSS.WriteString(fmt.Sprintf("border:none;width:%spx;height:%spx;", pxVal(absW), pxVal(absH)))

	encoded := base64.StdEncoding.EncodeToString(pngBytes)
	imgCSS := fmt.Sprintf(
		"background: url('data:image/Png;base64,%s') no-repeat !important;-webkit-print-color-adjust:exact;",
		encoded,
	)

	if e.InlineStyles {
		e.sb.WriteString(fmt.Sprintf(
			"<div style=\"%s%sleft:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
			picCSS.String(), imgCSS, pxVal(left), pxVal(top), pxVal(w), pxVal(h)))
	} else {
		picClass := e.css.Register(picCSS.String())
		imgClass := e.css.Register(imgCSS)
		e.sb.WriteString(fmt.Sprintf(
			"<div class=\"%s %s\" style=\"left:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
			picClass, imgClass, pxVal(left), pxVal(top), pxVal(w), pxVal(h)))
	}
}

// renderObjectAsImageWithPNG is like renderObjectAsImage but uses pre-rendered PNG data
// and explicit CSS dimensions (for line cap expansion where canvas > line dims).
func (e *Exporter) renderObjectAsImageWithPNG(obj preview.PreparedObject, left, top, w, h, scale float32, pngBytes []byte) {
	var picCSS strings.Builder
	picCSS.WriteString("text-align:center;position:absolute;color:rgb(255, 255, 255);")
	if obj.FillColor.A == 0 {
		picCSS.WriteString("background-color:transparent;")
	} else {
		picCSS.WriteString(fmt.Sprintf("background-color:%s;", rgbColor(obj.FillColor)))
	}
	absW := w
	if absW < 0 {
		absW = -absW
	}
	absH := h
	if absH < 0 {
		absH = -absH
	}
	picCSS.WriteString(fmt.Sprintf("border:none;width:%spx;height:%spx;", pxVal(absW), pxVal(absH)))

	encoded := base64.StdEncoding.EncodeToString(pngBytes)
	imgCSS := fmt.Sprintf(
		"background: url('data:image/Png;base64,%s') no-repeat !important;background-size:100%% 100%%;-webkit-print-color-adjust:exact;",
		encoded,
	)

	if e.InlineStyles {
		e.sb.WriteString(fmt.Sprintf(
			"<div style=\"%s%sleft:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
			picCSS.String(), imgCSS, pxVal(left), pxVal(top), pxVal(absW), pxVal(absH)))
	} else {
		picClass := e.css.Register(picCSS.String())
		imgClass := e.css.Register(imgCSS)
		e.sb.WriteString(fmt.Sprintf(
			"<div class=\"%s %s\" style=\"left:%spx;top:%spx;width:%spx;height:%spx;\">&nbsp;</div>\n",
			picClass, imgClass, pxVal(left), pxVal(top), pxVal(absW), pxVal(absH)))
	}
}

// renderObjectAsInlineImage renders a PreparedObject as a single div with
// inline style containing a base64 PNG background image. No CSS classes.
// Matches C# rendering of ShapeObjects (HTMLExportLayers.cs).
func (e *Exporter) renderObjectAsInlineImage(obj preview.PreparedObject, left, top, w, h float32) {
	pngBytes, _, _, err := imgexp.RenderGenericPNG(obj)
	if err != nil || len(pngBytes) == 0 {
		return
	}
	encoded := base64.StdEncoding.EncodeToString(pngBytes)
	e.sb.WriteString(fmt.Sprintf(
		"<div  style=\"left:%spx;top:%spx;width:%spx;height:%spx;position:absolute;background: url('data:image/Png;base64,%s') no-repeat !important;-webkit-print-color-adjust:exact;\">&nbsp;</div>\n",
		pxVal(left), pxVal(top), pxVal(w), pxVal(h), encoded))
}

func (e *Exporter) renderTextObject(obj preview.PreparedObject, scale float32) {
	left := obj.Left * scale
	top := obj.Top * scale
	w := obj.Width * scale
	h := obj.Height * scale

	font := obj.Font
	// Convert font size from pt to px (96dpi / 72pt = 1.3333...) to match C# output.
	fontPx := font.Size * 96.0 / 72.0

	// Wingdings/Webdings symbol fonts: remap characters to the Unicode Private Use Area
	// so browsers can display them correctly.
	// Mirrors C# HTMLExportLayers.LayerText lines 230-233.
	text := obj.Text
	if strings.EqualFold(font.Name, "Wingdings") || strings.EqualFold(font.Name, "Webdings") {
		text = utils.WingdingsToUnicode(text)
	}

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

	// Line height: ParagraphFormat.LineSpacingType takes precedence over LineHeight.
	// Mirrors C# AdvancedTextRenderer line-spacing computation (HtmlTextRenderer.cs lines 1670-1684).
	if obj.ParagraphLineSpacingType != 0 && obj.ParagraphLineSpacing > 0 {
		switch obj.ParagraphLineSpacingType {
		case 2: // Exactly — total line height is exactly LineSpacing px
			outerCSS.WriteString(fmt.Sprintf("line-height: %spx;", pxVal(obj.ParagraphLineSpacing*scale)))
		case 3: // Multiple — LineSpacing is a multiplier of the natural line height
			outerCSS.WriteString(fmt.Sprintf("line-height: %.2f;", obj.ParagraphLineSpacing))
		default: // AtLeast (1) — use as minimum; CSS line-height is already an "at least" spec
			outerCSS.WriteString(fmt.Sprintf("line-height: %spx;", pxVal(obj.ParagraphLineSpacing*scale)))
		}
	} else if obj.LineHeight > 0 {
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
		// ForceJustify: also justify the last line via CSS text-align-last.
		// Mirrors C# ForceJustify → AdvancedTextRenderer parameter (TextObject.cs line 1212).
		if obj.ForceJustify {
			outerCSS.WriteString(";text-align-last:justify")
		}
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
	if tc.A == 0 {
		tc = color.RGBA{A: 255} // default opaque black, matching C#
	}
	outerCSS.WriteString(fmt.Sprintf("color:%s;", rgbColor(tc)))
	fc := obj.FillColor
	// C# ref: when a glass fill background image is present, the outer div uses
	// background-color:transparent since the fill is entirely in the PNG image.
	if obj.BackgroundCSS != "" {
		outerCSS.WriteString("background-color:transparent;")
	} else if fc.A > 0 {
		outerCSS.WriteString(fmt.Sprintf("background-color:%s;", rgbColor(fc)))
	} else {
		outerCSS.WriteString("background-color:transparent;")
	}
	if obj.RTL {
		outerCSS.WriteString("direction:rtl;")
	}

	// Border (C# HTMLBorder).
	// Shadow is rendered as separate overlay divs (C# LayerBack lines 631-648), not CSS box-shadow.
	// Use a shadow-free copy so the CSS class does not include box-shadow.
	// C# HTMLBorder (HTMLExportDraw.cs) never emits box-shadow; shadow is handled by LayerBack.
	borderForCSS := obj.Border
	borderForCSS.Shadow = false
	borderStr := borderCSS(&borderForCSS, scale)
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
	// ParagraphFirstLineIndent (from ParagraphFormat) takes precedence over ParagraphOffset
	// when non-zero. Mirrors C# GetStartPosition() (HtmlTextRenderer.cs lines 1278-1283).
	if obj.ParagraphFirstLineIndent != 0 {
		innerCSS.WriteString(fmt.Sprintf("text-indent:%spx;", pxVal(obj.ParagraphFirstLineIndent*scale)))
	} else if obj.ParagraphOffset != 0 {
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
	// C# LayerText (HTMLExportLayers.cs:259-320) uses AdvancedTextRenderer.CalcHeight() for
	// the actual multi-line text height. We approximate using utils.MeasureLines which counts
	// word-wrap line breaks as the C# renderer does.
	// C# formula (GetSpanText line 201): margin-top = (top - paddingTop) * Zoom
	//   where top = (Height - textHeight - paddingBottom + paddingTop) / 2 for center.
	// Substituting: margin-top = (Height - textHeight - paddingBottom - paddingTop) / 2.
	lhRatio := fontMeasureRatioFloat(font.Name)
	switch obj.VertAlign {
	case 1: // center
		lineH := float64(fontPx) * lhRatio
		availW := float32(w) - obj.PaddingLeft*scale - obj.PaddingRight*scale
		lineCount := utils.MeasureLines(text, font, availW)
		textH := float64(lineCount) * lineH
		marginTop := (float64(h) - textH - float64(obj.PaddingBottom*scale) - float64(obj.PaddingTop*scale)) / 2
		if marginTop > 0 {
			innerCSS.WriteString(fmt.Sprintf("margin-top:%spx;", pxVal(float32(marginTop))))
		}
	case 2: // bottom
		lineH := float64(fontPx) * lhRatio
		availW := float32(w) - obj.PaddingLeft*scale - obj.PaddingRight*scale
		lineCount := utils.MeasureLines(text, font, availW)
		textH := float64(lineCount) * lineH
		marginTop := float64(h) - textH - float64(obj.PaddingBottom*scale) - float64(obj.PaddingTop*scale)
		if marginTop > 0 {
			innerCSS.WriteString(fmt.Sprintf("margin-top:%spx;", pxVal(float32(marginTop))))
		}
	}

	// No-wrap / ellipsis trimming for non-wrapping text objects (C# GetSpanText).
	// Trimming 3=EllipsisCharacter, 4=EllipsisWord, 5=EllipsisPath → CSS text-overflow:ellipsis.
	// Mirrors C# StringTrimming.EllipsisCharacter/EllipsisWord/EllipsisPath handling
	// (TextRenderer.cs AdvancedTextRenderer.WrapLines).
	if !obj.WordWrap {
		if obj.Trimming >= 3 {
			innerCSS.WriteString("overflow:hidden;white-space:nowrap;text-overflow:ellipsis;")
		} else {
			innerCSS.WriteString("overflow: hidden; text-wrap: nowrap;")
		}
	}

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

	// Shadow rendering: C# LayerBack emits two extra overlay divs for the drop shadow.
	// One horizontal strip below the object and one vertical strip to the right.
	// C# ref: HTMLExportLayers.cs LayerBack lines 631-648.
	if obj.Border.Shadow {
		sw := obj.Border.ShadowWidth * scale
		leftW := borderSideWidth(&obj.Border, style.BorderLeft, scale)
		topW := borderSideWidth(&obj.Border, style.BorderTop, scale)
		rightW := borderSideWidth(&obj.Border, style.BorderRight, scale)
		bottomW := borderSideWidth(&obj.Border, style.BorderBottom, scale)
		sc := obj.Border.ShadowColor
		// position:absolute required so the shadow div is placed by left/top inline style.
		// C# GetStyle always generates position:absolute for its shadow TextObject.
		shadowFill := fmt.Sprintf("position:absolute;background-color:%s;", rgbColor(sc))
		// Bottom horizontal shadow strip.
		s1Left := left + sw + leftW
		s1Top := top + h + bottomW
		s1W := w + rightW
		s1H := sw + bottomW
		shadow1Style := fmt.Sprintf("left:%spx;top:%spx;width:%spx;height:%spx;border:none;",
			pxVal(s1Left), pxVal(s1Top), pxVal(s1W), pxVal(s1H))
		// Right vertical shadow strip.
		s2Left := left + w + rightW
		s2Top := top + sw + topW
		s2W := sw + rightW
		s2H := h
		shadow2Style := fmt.Sprintf("left:%spx;top:%spx;width:%spx;height:%spx;border:none;",
			pxVal(s2Left), pxVal(s2Top), pxVal(s2W), pxVal(s2H))
		if e.InlineStyles {
			e.sb.WriteString(fmt.Sprintf("<div style=\"%s%s\">&nbsp;</div>\n", shadowFill, shadow1Style))
			e.sb.WriteString(fmt.Sprintf("<div style=\"%s%s\">&nbsp;</div>\n", shadowFill, shadow2Style))
		} else {
			shadowClass := e.cssClass(shadowFill)
			e.sb.WriteString(fmt.Sprintf("<div class=\"%s\" style=\"%s\">&nbsp;</div>\n", shadowClass, shadow1Style))
			e.sb.WriteString(fmt.Sprintf("<div class=\"%s\" style=\"%s\">&nbsp;</div>\n", shadowClass, shadow2Style))
		}
	}

	// For non-memo text objects (Angle != 0 or Underlines == true), emit LayerBack + LayerPicture and return.
	// C# IsMemo() returns false when Angle != 0 or Underlines == true, so the object is rendered as:
	//   LayerBack:    div with full text styling CSS + border:none in inline style
	//   LayerPicture: div with same CSS class + base64-PNG background-image
	// C# uses GetStyle(obj) for LayerBack which includes the full text styling.
	// C# ref: HTMLExportLayers.cs IsMemo() line 719, LayerBack/LayerPicture lines 937-955.
	if obj.Angle != 0 || obj.Underlines {
		posStyle := fmt.Sprintf("left:%spx;top:%spx;width:%spx;height:%spx;", pxVal(adjLeft), pxVal(adjTop), pxVal(adjW), pxVal(adjH))

		// Render the text object as a PNG and embed as base64 data URL (LayerPicture).
		picCSS := ""
		if pngBytes, err := imgexp.RenderObjectPNG(obj); err == nil {
			b64 := base64.StdEncoding.EncodeToString(pngBytes)
			picCSS = fmt.Sprintf("background: url('data:image/Png;base64,%s') no-repeat !important;-webkit-print-color-adjust:exact;", b64)
		}

		if e.InlineStyles {
			// LayerBack: full text styling + border:none
			e.sb.WriteString(fmt.Sprintf("<div style=\"%s%sborder:none;\">&nbsp;</div>\n", outerCSS.String(), posStyle))
			// LayerPicture: full text styling + background-image
			e.sb.WriteString(fmt.Sprintf("<div style=\"%s%s%s\">&nbsp;</div>\n", outerCSS.String(), posStyle, picCSS))
		} else {
			// Use outerCSS for the LayerBack class (matches C# GetStyle(obj))
			layerClass := e.cssClass(outerCSS.String())
			// LayerBack
			e.sb.WriteString(fmt.Sprintf("<div class=\"%s\" style=\"%sborder:none;\">&nbsp;</div>\n", layerClass, posStyle))
			// LayerPicture: same class + image class
			if picCSS != "" {
				imgClass := e.css.Register(picCSS)
				e.sb.WriteString(fmt.Sprintf("<div class=\"%s %s\" style=\"%s\">&nbsp;</div>\n", layerClass, imgClass, posStyle))
			} else {
				e.sb.WriteString(fmt.Sprintf("<div class=\"%s\" style=\"%s\">&nbsp;</div>\n", layerClass, posStyle))
			}
		}
		return
	}

	// Register CSS classes only for non-rotated text objects (rotated ones returned above).
	// C# flow: inner registered first via GetSpanText, then outer via LayerBack→GetStyle.
	innerClass := e.cssClass(innerCSS.String())
	outerClass := e.cssClass(outerCSS.String())
	// Add background image class for non-solid fills (GlassFill, etc.).
	// C# ref: HTMLExportLayers.cs LayerPicture → GetStyleTag creates dual-class CSS.
	if obj.BackgroundCSS != "" && !e.InlineStyles {
		bgClass := e.css.Register(obj.BackgroundCSS)
		outerClass = outerClass + " " + bgClass
	}

	// ── Format text content ──
	var innerText string
	if obj.TextRenderType == 1 || obj.TextRenderType == 2 {
		innerText = text
	} else {
		innerText = formatTextContent(text, fontPx)
	}

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
		// For landscape pages, C# adds width + 90° rotation to the frpage div so the
		// landscape content displays correctly in portrait HTML layout.
		// C# ref: HTMLExportLayers.cs Layer() landscape rotation, HTMLExportStyles.cs.
		landscapeCSS := ""
		if pg.Landscape {
			// C# uses the actual content height (max bottom of all bands) for
			// the rotated page width, not the page dimension minus margins.
			var contentH float32
			for _, b := range pg.Bands {
				if bottom := b.Top + b.Height; bottom > contentH {
					contentH = bottom
				}
			}
			landscapeCSS = fmt.Sprintf("width:%.2fpx !important;transform: rotate(90deg); -webkit-transform: rotate(90deg)", contentH)
		}
		e.sb.WriteString(fmt.Sprintf("<style type=\"text/css\" media=\"print\"><!--\ndiv.frpage%d { break-after: always; page-break-inside: avoid; %s}\n @page { size: portrait; margin: 0; }--></style>\n", pageNum, landscapeCSS))
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

// borderSideWidth returns the rendered width of a single border side in pixels (scaled).
// Mirrors C# HTMLBorderWidth: Double lines use Width*3, others use Width.
// Returns 1*scale as default when the border or its line is nil.
// Used for shadow offset calculations where we need the raw line width regardless of
// whether VisibleLines includes that side.
// C# ref: HTMLExportDraw.cs HTMLBorderWidth lines 63-68.
func borderSideWidth(b *style.Border, side style.BorderSide, scale float32) float32 {
	if b == nil || b.Lines[side] == nil {
		return 1 * scale
	}
	line := b.Lines[side]
	if line.Style == style.LineStyleDouble {
		return line.Width * 3 * scale
	}
	return line.Width * scale
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
		// C# HTMLBorder uses longhand properties: border-X-width, border-X-color,
		// border-X-style (matching BandBase.cs Layer method).
		sb.WriteString(fmt.Sprintf("%s-width:%spx;%s-color:rgb(%d, %d, %d);%s-style:%s;",
			s.prop, pxVal(width), s.prop, c.R, c.G, c.B, s.prop, lineStyle))
	}

	if b.Shadow {
		sw := b.ShadowWidth * scale
		sc := b.ShadowColor
		sb.WriteString(fmt.Sprintf("box-shadow:%.2fpx %.2fpx 0 rgba(%d,%d,%d,%.2f);",
			sw, sw, sc.R, sc.G, sc.B, float32(sc.A)/255.0))
	}

	return sb.String()
}

// resizeImagePNG decodes any image format (PNG, JPEG, GIF) and scales/positions
// it according to sizeMode (matches C# PictureObject SizeMode enum values):
//   0 = Normal     — show at original size, clip to bounds
//   1 = Stretch    — stretch to fill entire bounds (no aspect ratio preservation)
//   2 = AutoSize   — like Zoom (object auto-sizes to image; in export, treat as Zoom)
//   3 = CenterImage— center at original size, no scaling
//   4 = Zoom       — proportional fit, centered (default)
//
// The result is PNG-encoded. Matches C# GetLayerPicture / PictureObject.Draw().
func resizeImagePNG(data []byte, targetW, targetH, sizeMode int) []byte {
	if targetW <= 0 || targetH <= 0 {
		return data
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data
	}
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	// Create target bitmap with transparent background (matching C# Bitmap + clear).
	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))

	switch sizeMode {
	case 1: // StretchImage — fill entire target, ignoring aspect ratio.
		// C# ref: PictureBoxSizeMode.StretchImage
		xdraw.CatmullRom.Scale(dst, image.Rect(0, 0, targetW, targetH), src, srcBounds, xdraw.Over, nil)

	case 3: // CenterImage — center at original size, no scaling.
		// C# ref: PictureBoxSizeMode.CenterImage
		offsetX := (targetW - srcW) / 2
		offsetY := (targetH - srcH) / 2
		// Clip source to target bounds if image is larger.
		srcClipX := 0
		srcClipY := 0
		drawW := srcW
		drawH := srcH
		if offsetX < 0 {
			srcClipX = -offsetX
			drawW = targetW
			offsetX = 0
		}
		if offsetY < 0 {
			srcClipY = -offsetY
			drawH = targetH
			offsetY = 0
		}
		if drawW > targetW-offsetX {
			drawW = targetW - offsetX
		}
		if drawH > targetH-offsetY {
			drawH = targetH - offsetY
		}
		srcRect := image.Rect(srcBounds.Min.X+srcClipX, srcBounds.Min.Y+srcClipY,
			srcBounds.Min.X+srcClipX+drawW, srcBounds.Min.Y+srcClipY+drawH)
		dstRect := image.Rect(offsetX, offsetY, offsetX+drawW, offsetY+drawH)
		xdraw.NearestNeighbor.Scale(dst, dstRect, src, srcRect, xdraw.Over, nil)

	case 0: // Normal — show at original size, clip to bounds (top-left origin).
		// C# ref: PictureBoxSizeMode.Normal
		drawW := srcW
		drawH := srcH
		if drawW > targetW {
			drawW = targetW
		}
		if drawH > targetH {
			drawH = targetH
		}
		srcRect := image.Rect(srcBounds.Min.X, srcBounds.Min.Y, srcBounds.Min.X+drawW, srcBounds.Min.Y+drawH)
		dstRect := image.Rect(0, 0, drawW, drawH)
		xdraw.NearestNeighbor.Scale(dst, dstRect, src, srcRect, xdraw.Over, nil)

	default: // 2 (AutoSize) and 4 (Zoom) — proportional fit, centered.
		// C# ref: PictureBoxSizeMode.Zoom / AutoSize
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
		drawRect := image.Rect(offsetX, offsetY, offsetX+drawW, offsetY+drawH)
		xdraw.CatmullRom.Scale(dst, drawRect, src, srcBounds, xdraw.Over, nil)
	}

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

