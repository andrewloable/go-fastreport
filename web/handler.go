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

// ── CORS middleware ───────────────────────────────────────────────────────────

// CORSConfig holds CORS configuration for the web handlers.
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins. Use ["*"] to allow any origin.
	AllowedOrigins []string
	// AllowedMethods is the list of allowed HTTP methods (default: GET, OPTIONS).
	AllowedMethods []string
	// AllowedHeaders is the list of allowed request headers.
	AllowedHeaders []string
	// AllowCredentials sets the Access-Control-Allow-Credentials header.
	AllowCredentials bool
	// MaxAge is the preflight cache duration in seconds (0 = no header).
	MaxAge int
}

// DefaultCORSConfig returns a permissive CORS configuration suitable for
// development. Production deployments should restrict AllowedOrigins.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}
}

// WithCORS wraps next with CORS headers based on cfg.
// It handles OPTIONS preflight requests and sets the appropriate
// Access-Control-* headers on all responses.
func WithCORS(next http.Handler, cfg CORSConfig) http.Handler {
	if len(cfg.AllowedMethods) == 0 {
		cfg.AllowedMethods = []string{"GET", "OPTIONS"}
	}

	allowOriginFor := func(origin string) string {
		if len(cfg.AllowedOrigins) == 0 {
			return ""
		}
		for _, o := range cfg.AllowedOrigins {
			if o == "*" || o == origin {
				return o
			}
		}
		return ""
	}

	joinStr := func(parts []string, sep string) string {
		result := ""
		for i, p := range parts {
			if i > 0 {
				result += sep
			}
			result += p
		}
		return result
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := allowOriginFor(origin)
		if allowed == "" && origin != "" {
			// Origin not allowed — serve request without CORS headers.
			next.ServeHTTP(w, r)
			return
		}
		if allowed == "" {
			// No Origin header (non-browser request).
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", allowed)
		w.Header().Set("Access-Control-Allow-Methods", joinStr(cfg.AllowedMethods, ", "))
		if len(cfg.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", joinStr(cfg.AllowedHeaders, ", "))
		}
		if cfg.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if cfg.MaxAge > 0 {
			// Use simple integer formatting without fmt import.
			age := cfg.MaxAge
			tmp := [20]byte{}
			idx := len(tmp)
			for age > 0 {
				idx--
				tmp[idx] = byte('0' + age%10)
				age /= 10
			}
			w.Header().Set("Access-Control-Max-Age", string(tmp[idx:]))
		}

		// Handle OPTIONS preflight.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
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
