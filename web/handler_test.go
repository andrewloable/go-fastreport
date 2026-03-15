package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/web"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func buildPages(n int) *preview.PreparedPages {
	pp := preview.New()
	for i := 0; i < n; i++ {
		pp.AddPage(794, 1123, i+1)
		_ = pp.AddBand(&preview.PreparedBand{
			Name:   "DataBand",
			Top:    0,
			Height: 40,
		})
	}
	return pp
}

func get(handler http.Handler, url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// ── HTMLHandler ───────────────────────────────────────────────────────────────

func TestHTMLHandler_ContentType(t *testing.T) {
	h := web.HTMLHandler(buildPages(1))
	rec := get(h, "/report.html")

	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

func TestHTMLHandler_StatusOK(t *testing.T) {
	h := web.HTMLHandler(buildPages(1))
	rec := get(h, "/report.html")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestHTMLHandler_BodyContainsDoctype(t *testing.T) {
	h := web.HTMLHandler(buildPages(1))
	rec := get(h, "/report.html")
	if !strings.Contains(rec.Body.String(), "<!DOCTYPE html>") {
		t.Error("HTML body should contain DOCTYPE")
	}
}

func TestHTMLHandler_WithTitle(t *testing.T) {
	h := web.HTMLHandler(buildPages(1), web.WithTitle("My Report"))
	rec := get(h, "/")
	if !strings.Contains(rec.Body.String(), "My Report") {
		t.Error("HTML body should contain title 'My Report'")
	}
}

func TestHTMLHandler_WithEmbedCSS(t *testing.T) {
	h := web.HTMLHandler(buildPages(1), web.WithEmbedCSS(false))
	rec := get(h, "/")
	body := rec.Body.String()
	// Without embedded CSS there should be no <style> block.
	if strings.Contains(body, "<style>") {
		t.Error("should not embed CSS when WithEmbedCSS(false)")
	}
}

func TestHTMLHandler_PageRangeQuery(t *testing.T) {
	h := web.HTMLHandler(buildPages(3))
	// Export only page 2 via query parameter.
	rec := get(h, "/report.html?pages=2")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestHTMLHandler_NilPages(t *testing.T) {
	h := web.HTMLHandler(nil)
	rec := get(h, "/")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("nil pages: status = %d, want 500", rec.Code)
	}
}

// ── PDFHandler ────────────────────────────────────────────────────────────────

func TestPDFHandler_ContentType(t *testing.T) {
	h := web.PDFHandler(buildPages(1), "")
	rec := get(h, "/report.pdf")
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Content-Type = %q, want application/pdf", ct)
	}
}

func TestPDFHandler_StatusOK(t *testing.T) {
	h := web.PDFHandler(buildPages(1), "")
	rec := get(h, "/report.pdf")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestPDFHandler_ContentDisposition(t *testing.T) {
	h := web.PDFHandler(buildPages(1), "report.pdf")
	rec := get(h, "/report.pdf")
	cd := rec.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "report.pdf") {
		t.Errorf("Content-Disposition = %q, want 'attachment; filename=\"report.pdf\"'", cd)
	}
}

func TestPDFHandler_NoContentDisposition_WhenFilenameEmpty(t *testing.T) {
	h := web.PDFHandler(buildPages(1), "")
	rec := get(h, "/report.pdf")
	if cd := rec.Header().Get("Content-Disposition"); cd != "" {
		t.Errorf("expected no Content-Disposition, got %q", cd)
	}
}

func TestPDFHandler_NilPages(t *testing.T) {
	h := web.PDFHandler(nil, "")
	rec := get(h, "/")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("nil pages: status = %d, want 500", rec.Code)
	}
}

// ── ImageHandler ──────────────────────────────────────────────────────────────

func TestImageHandler_ContentType(t *testing.T) {
	h := web.ImageHandler(buildPages(1))
	rec := get(h, "/report.png")
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", ct)
	}
}

func TestImageHandler_StatusOK(t *testing.T) {
	h := web.ImageHandler(buildPages(1))
	rec := get(h, "/report.png")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestImageHandler_BodyNonEmpty(t *testing.T) {
	h := web.ImageHandler(buildPages(1))
	rec := get(h, "/report.png")
	if rec.Body.Len() == 0 {
		t.Error("PNG body should not be empty")
	}
}

func TestImageHandler_WithScale(t *testing.T) {
	h := web.ImageHandler(buildPages(1), web.WithImageScale(2.0))
	rec := get(h, "/report.png")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestImageHandler_NilPages(t *testing.T) {
	h := web.ImageHandler(nil)
	rec := get(h, "/")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("nil pages: status = %d, want 500", rec.Code)
	}
}

// ── NewHandler / Middleware ────────────────────────────────────────────────────

func TestNewHandler_HTML(t *testing.T) {
	h := web.NewHandler(buildPages(1), web.FormatHTML, web.HandlerOptions{})
	rec := get(h, "/")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

func TestNewHandler_PDF(t *testing.T) {
	h := web.NewHandler(buildPages(1), web.FormatPDF, web.HandlerOptions{})
	rec := get(h, "/")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Content-Type = %q, want application/pdf", ct)
	}
}

func TestNewHandler_Image(t *testing.T) {
	h := web.NewHandler(buildPages(1), web.FormatImage, web.HandlerOptions{})
	rec := get(h, "/")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", ct)
	}
}

func TestNewHandler_CustomHeaders(t *testing.T) {
	h := web.NewHandler(buildPages(1), web.FormatHTML, web.HandlerOptions{
		CustomHeaders: map[string]string{"X-Report-Id": "abc123"},
	})
	rec := get(h, "/")
	if v := rec.Header().Get("X-Report-Id"); v != "abc123" {
		t.Errorf("X-Report-Id = %q, want abc123", v)
	}
}

func TestNewHandler_Attachment(t *testing.T) {
	h := web.NewHandler(buildPages(1), web.FormatPDF, web.HandlerOptions{
		Disposition: "attachment",
		Filename:    "my-report.pdf",
	})
	rec := get(h, "/")
	cd := rec.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "my-report.pdf") {
		t.Errorf("Content-Disposition = %q, want to contain my-report.pdf", cd)
	}
}

func TestNewHandler_ErrorHandler(t *testing.T) {
	called := false
	h := web.NewHandler(nil, web.FormatHTML, web.HandlerOptions{
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			called = true
			w.WriteHeader(http.StatusTeapot)
		},
	})
	rec := get(h, "/")
	if !called {
		t.Error("custom error handler should have been called")
	}
	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want 418", rec.Code)
	}
}

func TestNewHandler_Middleware(t *testing.T) {
	var order []string
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m1-before")
			next.ServeHTTP(w, r)
			order = append(order, "m1-after")
		})
	}
	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m2-before")
			next.ServeHTTP(w, r)
			order = append(order, "m2-after")
		})
	}
	h := web.NewHandler(buildPages(1), web.FormatHTML, web.HandlerOptions{
		Middleware: []web.Middleware{m1, m2},
	})
	get(h, "/")
	if len(order) != 4 {
		t.Fatalf("middleware call order len = %d, want 4: %v", len(order), order)
	}
}

func TestChain_Order(t *testing.T) {
	var calls []string
	m1 := web.Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "m1")
			next.ServeHTTP(w, r)
		})
	})
	m2 := web.Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "m2")
			next.ServeHTTP(w, r)
		})
	})
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, "handler")
	})
	web.Chain(m1, m2)(final).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if len(calls) != 3 || calls[0] != "m1" || calls[1] != "m2" || calls[2] != "handler" {
		t.Errorf("chain order = %v, want [m1 m2 handler]", calls)
	}
}

func TestWithHeader_Middleware(t *testing.T) {
	h := web.WithHeader("X-Custom", "value")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if v := rec.Header().Get("X-Custom"); v != "value" {
		t.Errorf("X-Custom = %q, want value", v)
	}
}

func TestNewHandler_UnknownFormat(t *testing.T) {
	h := web.NewHandler(buildPages(1), "csv", web.HandlerOptions{})
	rec := get(h, "/")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("unknown format: status = %d, want 400", rec.Code)
	}
}

// ── mux integration ──────────────────────────────────────────────────────────

func TestHandlers_RegisteredOnMux(t *testing.T) {
	pp := buildPages(2)
	mux := http.NewServeMux()
	mux.Handle("/report.html", web.HTMLHandler(pp, web.WithTitle("Test")))
	mux.Handle("/report.pdf", web.PDFHandler(pp, "output.pdf"))
	mux.Handle("/report.png", web.ImageHandler(pp))

	for _, path := range []string{"/report.html", "/report.pdf", "/report.png"} {
		rec := get(mux, path)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s: status = %d, want 200", path, rec.Code)
		}
	}
}

// ── DefaultCORSConfig ─────────────────────────────────────────────────────────

func TestDefaultCORSConfig_Fields(t *testing.T) {
	cfg := web.DefaultCORSConfig()
	if len(cfg.AllowedOrigins) == 0 || cfg.AllowedOrigins[0] != "*" {
		t.Errorf("AllowedOrigins = %v, want [*]", cfg.AllowedOrigins)
	}
	if len(cfg.AllowedMethods) == 0 {
		t.Error("AllowedMethods should not be empty")
	}
	if len(cfg.AllowedHeaders) == 0 {
		t.Error("AllowedHeaders should not be empty")
	}
}

// ── WithCORS ──────────────────────────────────────────────────────────────────

func corsBase() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestWithCORS_AllowedOrigin_SetsHeader(t *testing.T) {
	h := web.WithCORS(corsBase(), web.DefaultCORSConfig())
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao == "" {
		t.Error("Access-Control-Allow-Origin should be set for allowed origin")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestWithCORS_OptionsPreflight_Returns204(t *testing.T) {
	h := web.WithCORS(corsBase(), web.DefaultCORSConfig())
	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Errorf("OPTIONS status = %d, want 204", rec.Code)
	}
}

func TestWithCORS_NonAllowedOrigin_NoCORSHeaders(t *testing.T) {
	cfg := web.CORSConfig{
		AllowedOrigins: []string{"http://trusted.com"},
		AllowedMethods: []string{"GET"},
	}
	h := web.WithCORS(corsBase(), cfg)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "" {
		t.Errorf("non-allowed origin should not get CORS headers, got %q", acao)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestWithCORS_NoOriginHeader_PassesThrough(t *testing.T) {
	h := web.WithCORS(corsBase(), web.DefaultCORSConfig())
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestWithCORS_EmptyAllowedMethods_DefaultsApplied(t *testing.T) {
	cfg := web.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{}, // empty — should default to GET, OPTIONS
	}
	h := web.WithCORS(corsBase(), cfg)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if acam := rec.Header().Get("Access-Control-Allow-Methods"); acam == "" {
		t.Error("Access-Control-Allow-Methods should be set with default methods")
	}
}

func TestWithCORS_AllowCredentials_SetsHeader(t *testing.T) {
	cfg := web.CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET"},
		AllowCredentials: true,
	}
	h := web.WithCORS(corsBase(), cfg)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if v := rec.Header().Get("Access-Control-Allow-Credentials"); v != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want true", v)
	}
}

func TestWithCORS_MaxAge_SetsHeader(t *testing.T) {
	cfg := web.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
		MaxAge:         300,
	}
	h := web.WithCORS(corsBase(), cfg)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if v := rec.Header().Get("Access-Control-Max-Age"); v == "" {
		t.Error("Access-Control-Max-Age should be set when MaxAge > 0")
	}
}

func TestWithCORS_EmptyAllowedOrigins_NoCORSHeaders(t *testing.T) {
	cfg := web.CORSConfig{
		AllowedOrigins: []string{},
		AllowedMethods: []string{"GET"},
	}
	h := web.WithCORS(corsBase(), cfg)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "" {
		t.Errorf("empty AllowedOrigins should produce no CORS header, got %q", acao)
	}
}

// ── WithHTMLScale ─────────────────────────────────────────────────────────────

func TestHTMLHandler_WithHTMLScale_StatusOK(t *testing.T) {
	h := web.HTMLHandler(buildPages(1), web.WithHTMLScale(2.0))
	rec := get(h, "/report.html")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestNewHandler_HTMLOpts_WithHTMLScale(t *testing.T) {
	h := web.NewHandler(buildPages(1), web.FormatHTML, web.HandlerOptions{
		HTMLOpts: []web.HTMLOption{web.WithHTMLScale(1.5)},
	})
	rec := get(h, "/")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

// ── WithTimeout ───────────────────────────────────────────────────────────────

func TestWithTimeout_ContextDeadlineSet(t *testing.T) {
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Deadline(); !ok {
			t.Error("expected context deadline to be set by WithTimeout")
		}
		w.WriteHeader(http.StatusOK)
	})
	h := web.WithTimeout(5 * time.Second)(final)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestNewHandler_WithTimeoutOption(t *testing.T) {
	h := web.NewHandler(buildPages(1), web.FormatHTML, web.HandlerOptions{
		Timeout: 5 * time.Second,
	})
	rec := get(h, "/")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

// ── WithPageRange ─────────────────────────────────────────────────────────────

func TestWithPageRange_Middleware_PassesThrough(t *testing.T) {
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := web.WithPageRange(export.PageRangeCustom, "1-3")(final)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

// ── applyDisposition branches ─────────────────────────────────────────────────

func TestNewHandler_AttachmentDefaultName(t *testing.T) {
	// Disposition=attachment, Filename="" → uses format default name "report.html"
	h := web.NewHandler(buildPages(1), web.FormatHTML, web.HandlerOptions{
		Disposition: "attachment",
		Filename:    "",
	})
	rec := get(h, "/")
	cd := rec.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "report.html") {
		t.Errorf("Content-Disposition = %q, want to contain report.html", cd)
	}
}

func TestNewHandler_InlineDisposition_NoHeader(t *testing.T) {
	h := web.NewHandler(buildPages(1), web.FormatPDF, web.HandlerOptions{
		Disposition: "inline",
	})
	rec := get(h, "/")
	if cd := rec.Header().Get("Content-Disposition"); cd != "" {
		t.Errorf("inline disposition should not set Content-Disposition, got %q", cd)
	}
}

func TestNewHandler_ImageOpts(t *testing.T) {
	h := web.NewHandler(buildPages(1), web.FormatImage, web.HandlerOptions{
		ImageOpts: []web.ImageOption{web.WithImageScale(2.0)},
	})
	rec := get(h, "/")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestWithCORS_NoOriginHeader_SpecificOrigin_PassesThrough(t *testing.T) {
	// No Origin header + non-wildcard config → allowOriginFor("") returns ""
	// → allowed == "" && origin == "" → hits the "No Origin header" branch (line 137).
	cfg := web.CORSConfig{
		AllowedOrigins: []string{"http://example.com"},
		AllowedMethods: []string{"GET"},
	}
	h := web.WithCORS(corsBase(), cfg)
	req := httptest.NewRequest("GET", "/", nil) // no Origin header
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

// ── default errHandler and error paths in buildCoreHandler ────────────────────

func TestNewHandler_HTML_NilPages_DefaultError(t *testing.T) {
	// No custom ErrorHandler → triggers default errHandler body.
	h := web.NewHandler(nil, web.FormatHTML, web.HandlerOptions{})
	rec := get(h, "/")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("nil pages HTML: status = %d, want 500", rec.Code)
	}
}

func TestNewHandler_PDF_NilPages_DefaultError(t *testing.T) {
	// nil pages with PDF format → triggers PDF error path in buildCoreHandler.
	h := web.NewHandler(nil, web.FormatPDF, web.HandlerOptions{})
	rec := get(h, "/")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("nil pages PDF: status = %d, want 500", rec.Code)
	}
}

func TestNewHandler_Image_NilPages_DefaultError(t *testing.T) {
	// nil pages with Image format → triggers Image error path in buildCoreHandler.
	h := web.NewHandler(nil, web.FormatImage, web.HandlerOptions{})
	rec := get(h, "/")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("nil pages Image: status = %d, want 500", rec.Code)
	}
}
