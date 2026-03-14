// Package html implements an HTML export filter for go-fastreport.
// It renders prepared pages as a single HTML document with absolute-positioned
// elements (WYSIWYG mode), one section per page.
package html

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// Exporter produces HTML output from a PreparedPages collection.
//
// Output structure:
//
//	<html><body>
//	  <div class="page" …> <!-- one per page -->
//	    <div class="band" …> … </div>
//	  </div>
//	</body></html>
type Exporter struct {
	export.ExportBase

	// Title is used as the HTML document title.
	Title string
	// EmbedCSS controls whether a minimal stylesheet is written inline.
	EmbedCSS bool
	// Scale converts pixel values to CSS pixels (default 1.0).
	Scale float32

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
	e.sb.WriteString(fmt.Sprintf(
		`<div class="page" style="position:relative;width:%.2fpx;height:%.2fpx;`+
			`overflow:hidden;margin:0 auto 20px auto;background:#fff;box-shadow:0 0 5px #aaa;">`,
		w, h,
	))
	e.sb.WriteString("\n")
	e.pageIdx++
	return nil
}

func (e *Exporter) ExportBand(b *preview.PreparedBand) error {
	scale := e.Scale
	if scale <= 0 {
		scale = 1
	}
	top := b.Top * scale
	h := b.Height * scale

	label := export.HTMLString(b.Name)
	e.sb.WriteString(fmt.Sprintf(
		`<div class="band" data-name="%s" style="position:absolute;top:%.2fpx;`+
			`left:0;right:0;height:%.2fpx;">`,
		label, top, h,
	))

	// Render each child object.
	for _, obj := range b.Objects {
		e.renderObject(obj, scale)
	}

	e.sb.WriteString("</div>\n")
	return nil
}

// renderObject writes an HTML element for a single PreparedObject.
func (e *Exporter) renderObject(obj preview.PreparedObject, scale float32) {
	left := obj.Left * scale
	top := obj.Top * scale
	w := obj.Width * scale
	h := obj.Height * scale

	// Build inline CSS.
	var css strings.Builder
	css.WriteString(fmt.Sprintf(
		"position:absolute;left:%.2fpx;top:%.2fpx;width:%.2fpx;height:%.2fpx;overflow:hidden;",
		left, top, w, h,
	))

	// Fill color.
	if obj.FillColor.A > 0 {
		css.WriteString(fmt.Sprintf(
			"background-color:rgba(%d,%d,%d,%.2f);",
			obj.FillColor.R, obj.FillColor.G, obj.FillColor.B,
			float32(obj.FillColor.A)/255.0,
		))
	}

	switch obj.Kind {
	case preview.ObjectTypeText:
		// Font styling.
		font := obj.Font
		css.WriteString(fmt.Sprintf("font-family:'%s';font-size:%.1fpt;", font.Name, font.Size))
		if font.Style&style.FontStyleBold != 0 {
			css.WriteString("font-weight:bold;")
		}
		if font.Style&style.FontStyleItalic != 0 {
			css.WriteString("font-style:italic;")
		}
		if font.Style&style.FontStyleUnderline != 0 {
			css.WriteString("text-decoration:underline;")
		}
		// Text color.
		tc := obj.TextColor
		css.WriteString(fmt.Sprintf("color:rgba(%d,%d,%d,%.2f);", tc.R, tc.G, tc.B, float32(tc.A)/255.0))
		// Horizontal alignment.
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
		// Vertical alignment via flex.
		switch obj.VertAlign {
		case 1:
			css.WriteString("display:flex;align-items:center;")
		case 2:
			css.WriteString("display:flex;align-items:flex-end;")
		}
		if obj.WordWrap {
			css.WriteString("word-wrap:break-word;white-space:normal;")
		} else {
			css.WriteString("white-space:nowrap;")
		}

		e.sb.WriteString(fmt.Sprintf(`<div style="%s">%s</div>`, css.String(), export.HTMLString(obj.Text)))

	case preview.ObjectTypePicture:
		if obj.BlobIdx >= 0 && e.pp != nil {
			if data := e.pp.BlobStore.Get(obj.BlobIdx); len(data) > 0 {
				mime := imageMIME(data)
				encoded := base64.StdEncoding.EncodeToString(data)
				e.sb.WriteString(fmt.Sprintf(
					`<div style="%s"><img src="data:%s;base64,%s" style="width:100%%;height:100%%;object-fit:contain;" alt=""></div>`,
					css.String(), mime, encoded,
				))
				break
			}
		}
		// No image data — empty placeholder.
		e.sb.WriteString(fmt.Sprintf(`<div style="%s"></div>`, css.String()))

	case preview.ObjectTypeLine:
		if obj.LineDiagonal {
			// Render diagonal line as an inline SVG.
			e.sb.WriteString(fmt.Sprintf(
				`<div style="%s"><svg width="%.2f" height="%.2f" style="position:absolute;top:0;left:0;">`,
				css.String(), w, h,
			))
			e.sb.WriteString(fmt.Sprintf(
				`<line x1="0" y1="0" x2="%.2f" y2="%.2f" stroke="#000" stroke-width="1"/>`,
				w, h,
			))
			e.sb.WriteString(`</svg></div>`)
		} else {
			// Horizontal or vertical: use border-bottom or border-left.
			if h <= w {
				css.WriteString("border-bottom:1px solid #000;height:1px;")
			} else {
				css.WriteString("border-left:1px solid #000;width:1px;")
			}
			e.sb.WriteString(fmt.Sprintf(`<div style="%s"></div>`, css.String()))
		}

	case preview.ObjectTypeShape:
		switch obj.ShapeKind {
		case 2: // Ellipse
			css.WriteString("border:1px solid #000;border-radius:50%;")
		case 1: // RoundRectangle
			css.WriteString(fmt.Sprintf("border:1px solid #000;border-radius:%.2fpx;", obj.ShapeCurve*scale))
		case 3: // Triangle — use inline SVG
			e.sb.WriteString(fmt.Sprintf(`<div style="%s">`, css.String()))
			e.sb.WriteString(fmt.Sprintf(
				`<svg width="%.2f" height="%.2f" style="position:absolute;top:0;left:0;">`,
				w, h,
			))
			e.sb.WriteString(fmt.Sprintf(
				`<polygon points="%.2f,0 0,%.2f %.2f,%.2f" stroke="#000" stroke-width="1" fill="none"/>`,
				w/2, h, w, h,
			))
			e.sb.WriteString(`</svg></div>`)
			break
		case 4: // Diamond — use inline SVG
			e.sb.WriteString(fmt.Sprintf(`<div style="%s">`, css.String()))
			e.sb.WriteString(fmt.Sprintf(
				`<svg width="%.2f" height="%.2f" style="position:absolute;top:0;left:0;">`,
				w, h,
			))
			e.sb.WriteString(fmt.Sprintf(
				`<polygon points="%.2f,0 %.2f,%.2f %.2f,%.2f 0,%.2f" stroke="#000" stroke-width="1" fill="none"/>`,
				w/2, w, h/2, w/2, h, h/2,
			))
			e.sb.WriteString(`</svg></div>`)
			break
		default: // Rectangle
			css.WriteString("border:1px solid #000;")
		}
		if obj.ShapeKind != 3 && obj.ShapeKind != 4 {
			e.sb.WriteString(fmt.Sprintf(`<div style="%s"></div>`, css.String()))
		}

	case preview.ObjectTypeCheckBox:
		checked := ""
		if obj.Text == "true" {
			checked = " checked"
		}
		e.sb.WriteString(fmt.Sprintf(
			`<div style="%s"><input type="checkbox"%s disabled style="margin:auto;"></div>`,
			css.String(), checked,
		))

	default:
		// Unknown/unhandled type — render an empty placeholder.
		e.sb.WriteString(fmt.Sprintf(`<div style="%s"></div>`, css.String()))
	}
}

func (e *Exporter) ExportPageEnd(_ *preview.PreparedPage) error {
	e.sb.WriteString("</div>\n") // close .page
	return nil
}

func (e *Exporter) Finish() error {
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

