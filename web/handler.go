// Package web provides HTTP handler utilities for go-fastreport.
package web

import (
	"net/http"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/export/html"
	imgexport "github.com/andrewloable/go-fastreport/export/image"
	"github.com/andrewloable/go-fastreport/export/pdf"
	"github.com/andrewloable/go-fastreport/preview"
)

// HTMLHandler returns an http.Handler that exports pp as an HTML document.
// Options configure the HTML exporter (title, CSS embedding, scale, etc.).
//
// Query parameter "pages" (optional): page range string like "1,3-5" or "all".
// Content-Type is set to "text/html; charset=utf-8".
func HTMLHandler(pp *preview.PreparedPages, opts ...HTMLOption) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exp := html.NewExporter()
		for _, opt := range opts {
			opt(exp)
		}
		applyQueryPageRange(&exp.ExportBase, r)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := exp.Export(pp, w); err != nil {
			http.Error(w, "report generation failed: "+err.Error(), http.StatusInternalServerError)
		}
	})
}

// PDFHandler returns an http.Handler that exports pp as a PDF document.
//
// If filename is non-empty, a Content-Disposition: attachment header is set.
// Query parameter "pages" (optional): page range string.
func PDFHandler(pp *preview.PreparedPages, filename string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exp := pdf.NewExporter()
		applyQueryPageRange(&exp.ExportBase, r)

		w.Header().Set("Content-Type", "application/pdf")
		if filename != "" {
			w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
		}
		if err := exp.Export(pp, w); err != nil {
			http.Error(w, "report generation failed: "+err.Error(), http.StatusInternalServerError)
		}
	})
}

// ImageHandler returns an http.Handler that exports pages of pp as PNG images.
//
// Query parameter "pages" (optional): page range string.
// Content-Type is set to "image/png".
func ImageHandler(pp *preview.PreparedPages, opts ...ImageOption) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exp := imgexport.NewExporter()
		for _, opt := range opts {
			opt(exp)
		}
		applyQueryPageRange(&exp.ExportBase, r)

		w.Header().Set("Content-Type", "image/png")
		if err := exp.Export(pp, w); err != nil {
			http.Error(w, "report generation failed: "+err.Error(), http.StatusInternalServerError)
		}
	})
}

// applyQueryPageRange reads the "pages" query parameter and sets the export
// base's page range. Supported values:
//
//   - empty or "all" → PageRangeAll (default)
//   - a page range string like "1,3-5,7" → PageRangeCustom
func applyQueryPageRange(base *export.ExportBase, r *http.Request) {
	pages := r.URL.Query().Get("pages")
	if pages == "" || pages == "all" {
		base.PageRange = export.PageRangeAll
		return
	}
	base.PageRange = export.PageRangeCustom
	base.PageNumbers = pages
}

// ── HTML options ──────────────────────────────────────────────────────────────

// HTMLOption is a functional option for configuring the HTML exporter.
type HTMLOption func(*html.Exporter)

// WithTitle sets the HTML document title.
func WithTitle(title string) HTMLOption {
	return func(e *html.Exporter) { e.Title = title }
}

// WithEmbedCSS controls whether CSS is embedded inline (default true).
func WithEmbedCSS(embed bool) HTMLOption {
	return func(e *html.Exporter) { e.EmbedCSS = embed }
}

// WithHTMLScale sets the rendering scale factor.
func WithHTMLScale(scale float32) HTMLOption {
	return func(e *html.Exporter) { e.Scale = scale }
}

// ── Image options ─────────────────────────────────────────────────────────────

// ImageOption is a functional option for configuring the image exporter.
type ImageOption func(*imgexport.Exporter)

// WithImageScale sets the PNG rendering scale factor.
func WithImageScale(scale float64) ImageOption {
	return func(e *imgexport.Exporter) { e.Scale = scale }
}
