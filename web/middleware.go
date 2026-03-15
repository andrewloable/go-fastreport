package web

import (
	"context"
	"net/http"
	"time"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/export/html"
	imgexport "github.com/andrewloable/go-fastreport/export/image"
	"github.com/andrewloable/go-fastreport/export/pdf"
	"github.com/andrewloable/go-fastreport/preview"
)

// Middleware is a function that wraps an http.Handler.
// Multiple middleware can be composed via Chain.
type Middleware func(http.Handler) http.Handler

// Chain applies middleware in order: first middleware is outermost.
// Chain(m1, m2, m3)(h) is equivalent to m1(m2(m3(h))).
func Chain(middleware ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middleware) - 1; i >= 0; i-- {
			next = middleware[i](next)
		}
		return next
	}
}

// HandlerOptions configures the behaviour of handlers created by NewHandler.
type HandlerOptions struct {
	// Timeout limits how long report generation may take.
	// Zero means no timeout.
	Timeout time.Duration

	// CustomHeaders are added to every response before writing the body.
	CustomHeaders map[string]string

	// ErrorHandler is called instead of http.Error when export fails.
	// If nil, the default behaviour (500 + error message) is used.
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

	// Disposition controls the Content-Disposition header for file downloads.
	// "inline" serves the file inline in the browser (default for HTML/image).
	// "attachment" triggers a download prompt. Filename sets the suggested name.
	Disposition string
	Filename    string

	// Middleware is applied after the built-in headers but before export.
	Middleware []Middleware

	// HTMLOpts configure the HTML exporter. Ignored for other formats.
	HTMLOpts []HTMLOption

	// ImageOpts configure the image exporter. Ignored for other formats.
	ImageOpts []ImageOption
}

// ExportFormat selects the output format for NewHandler.
type ExportFormat string

const (
	FormatHTML  ExportFormat = "html"
	FormatPDF   ExportFormat = "pdf"
	FormatImage ExportFormat = "image"
)

// NewHandler returns a configured http.Handler that exports pp in the given format.
// Options in opts control timeouts, custom headers, middleware, and more.
func NewHandler(pp *preview.PreparedPages, format ExportFormat, opts HandlerOptions) http.Handler {
	h := buildCoreHandler(pp, format, &opts)

	// Wrap with user-supplied middleware (innermost first).
	if len(opts.Middleware) > 0 {
		h = Chain(opts.Middleware...)(h)
	}

	// Wrap with built-in timeout + header injection.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Inject custom headers.
		for k, v := range opts.CustomHeaders {
			w.Header().Set(k, v)
		}

		// Apply timeout via context.
		if opts.Timeout > 0 {
			ctx, cancel := context.WithTimeout(r.Context(), opts.Timeout)
			defer cancel()
			r = r.WithContext(ctx)
		}

		h.ServeHTTP(w, r)
	})
}

// buildCoreHandler constructs the format-specific inner handler.
func buildCoreHandler(pp *preview.PreparedPages, format ExportFormat, opts *HandlerOptions) http.Handler {
	errHandler := opts.ErrorHandler
	if errHandler == nil {
		errHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "report generation failed: "+err.Error(), http.StatusInternalServerError)
		}
	}

	switch format {
	case FormatHTML:
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			exp := html.NewExporter()
			for _, o := range opts.HTMLOpts {
				o(exp)
			}
			applyQueryPageRange(&exp.ExportBase, r)
			applyDisposition(w, opts.Disposition, opts.Filename, "report.html")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if err := exp.Export(pp, w); err != nil {
				errHandler(w, r, err)
			}
		})

	case FormatPDF:
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			exp := pdf.NewExporter()
			applyQueryPageRange(&exp.ExportBase, r)
			w.Header().Set("Content-Type", "application/pdf")
			applyDisposition(w, opts.Disposition, opts.Filename, "report.pdf")
			if err := exp.Export(pp, w); err != nil {
				errHandler(w, r, err)
			}
		})

	case FormatImage:
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			exp := imgexport.NewExporter()
			for _, o := range opts.ImageOpts {
				o(exp)
			}
			applyQueryPageRange(&exp.ExportBase, r)
			w.Header().Set("Content-Type", "image/png")
			applyDisposition(w, opts.Disposition, opts.Filename, "report.png")
			if err := exp.Export(pp, w); err != nil {
				errHandler(w, r, err)
			}
		})

	default:
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unsupported export format: "+string(format), http.StatusBadRequest)
		})
	}
}

// applyDisposition sets a Content-Disposition header when the disposition is
// "attachment", using filename if set or defaultName otherwise.
// "inline" and empty disposition emit no header (browser default).
func applyDisposition(w http.ResponseWriter, disposition, filename, defaultName string) {
	if disposition != "attachment" {
		return
	}
	name := filename
	if name == "" {
		name = defaultName
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
}

// WithTimeout returns a Middleware that limits the request context lifetime.
func WithTimeout(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithHeader returns a Middleware that sets a fixed response header.
func WithHeader(key, value string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(key, value)
			next.ServeHTTP(w, r)
		})
	}
}

// WithPageRange returns a Middleware that pre-applies a fixed page range,
// overriding any "pages" query parameter supplied by the client.
func WithPageRange(pageRange export.PageRange, pageNumbers string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Store the override in context so core handlers can read it.
			ctx := context.WithValue(r.Context(), ctxKeyPageRange{}, pageRangeOverride{pr: pageRange, pageNumbers: pageNumbers})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type ctxKeyPageRange struct{}

type pageRangeOverride struct {
	pr          export.PageRange
	pageNumbers string
}
