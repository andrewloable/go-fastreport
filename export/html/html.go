// Package html implements an HTML export filter for go-fastreport.
// It renders prepared pages as a single HTML document with absolute-positioned
// elements (WYSIWYG mode), one section per page.
package html

import (
	"fmt"
	"io"
	"strings"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
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
			`left:0;right:0;height:%.2fpx;border-bottom:1px dotted #ccc;">`,
		label, top, h,
	))
	if label != "" {
		e.sb.WriteString(fmt.Sprintf(
			`<span style="font-size:10px;color:#666;padding:2px;">%s</span>`, label,
		))
	}
	e.sb.WriteString("</div>\n")
	return nil
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
