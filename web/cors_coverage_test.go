package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrewloable/go-fastreport/web"
)

// ── WithCORS remaining branch coverage ───────────────────────────────────────
// The missing 2% in WithCORS is the len(cfg.AllowedHeaders) == 0 false branch
// (when AllowedHeaders is empty, the header-setting block is skipped).
// All other CORS tests use non-empty AllowedHeaders.

func TestWithCORS_NoAllowedHeaders_SkipsHeaderSetting(t *testing.T) {
	// AllowedHeaders is nil/empty → the `if len(cfg.AllowedHeaders) > 0` branch is false.
	cfg := web.CORSConfig{
		AllowedOrigins:  []string{"*"},
		AllowedMethods:  []string{"GET", "POST"},
		AllowedHeaders:  nil, // explicitly nil → no header set
		AllowCredentials: false,
		MaxAge:           0,
	}
	h := web.WithCORS(corsBase(), cfg)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	// ACAO should be set (wildcard match).
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao == "" {
		t.Error("Access-Control-Allow-Origin should be set for * origin")
	}
	// ACAH should NOT be set when AllowedHeaders is empty.
	if acah := rec.Header().Get("Access-Control-Allow-Headers"); acah != "" {
		t.Errorf("Access-Control-Allow-Headers should not be set when empty, got %q", acah)
	}
}

func TestWithCORS_EmptyAllowedHeaders_Slice_SkipsHeaderSetting(t *testing.T) {
	// AllowedHeaders is an empty (not nil) slice.
	cfg := web.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{}, // empty slice
	}
	h := web.WithCORS(corsBase(), cfg)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if acah := rec.Header().Get("Access-Control-Allow-Headers"); acah != "" {
		t.Errorf("Access-Control-Allow-Headers should be empty when AllowedHeaders is [], got %q", acah)
	}
}

func TestWithCORS_MaxAge_WithCredentials_NoHeaders(t *testing.T) {
	// MaxAge > 0 + AllowCredentials + no AllowedHeaders — covers multiple branches.
	cfg := web.CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   nil,
		AllowCredentials: true,
		MaxAge:           600,
	}
	h := web.WithCORS(corsBase(), cfg)
	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("OPTIONS status = %d, want 204", rec.Code)
	}
	if v := rec.Header().Get("Access-Control-Allow-Credentials"); v != "true" {
		t.Errorf("Allow-Credentials = %q, want true", v)
	}
	if v := rec.Header().Get("Access-Control-Max-Age"); v == "" {
		t.Error("Max-Age should be set")
	}
	if acah := rec.Header().Get("Access-Control-Allow-Headers"); acah != "" {
		t.Errorf("Allow-Headers should not be set, got %q", acah)
	}
}

func TestWithCORS_SpecificOriginMatch(t *testing.T) {
	// o == origin (not wildcard) → returns origin.
	cfg := web.CORSConfig{
		AllowedOrigins: []string{"http://trusted.com", "http://example.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	}
	h := web.WithCORS(corsBase(), cfg)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "http://example.com" {
		t.Errorf("ACAO = %q, want http://example.com", acao)
	}
}

func TestWithCORS_JoinStr_MultipleHeaders(t *testing.T) {
	// joinStr with multiple methods/headers — covers the `i > 0` branch.
	cfg := web.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Authorization"},
	}
	h := web.WithCORS(corsBase(), cfg)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	acam := rec.Header().Get("Access-Control-Allow-Methods")
	if acam == "" {
		t.Error("ACAM should be set")
	}
	acah := rec.Header().Get("Access-Control-Allow-Headers")
	if acah == "" {
		t.Error("ACAH should be set")
	}
}
