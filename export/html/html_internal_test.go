// Internal tests for the html package.
// These tests live in package html so they can access unexported functions.
package html

import (
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

// ── borderCSS ─────────────────────────────────────────────────────────────────

func TestBorderCSS_Nil(t *testing.T) {
	got := borderCSS(nil, 1.0)
	if got != "" {
		t.Errorf("nil border: expected empty string, got %q", got)
	}
}

func TestBorderCSS_NoneVisible(t *testing.T) {
	b := &style.Border{VisibleLines: style.BorderLinesNone}
	got := borderCSS(b, 1.0)
	if got != "" {
		t.Errorf("BorderLinesNone: expected empty string, got %q", got)
	}
}

func TestBorderCSS_AllSides_Defaults(t *testing.T) {
	// All sides visible, nil Lines entries → uses code defaults (1px solid black).
	b := &style.Border{
		VisibleLines: style.BorderLinesAll,
		// Lines is zero-value: all nil
	}
	got := borderCSS(b, 1.0)
	// All four sides should be present.
	for _, side := range []string{"border-top:", "border-right:", "border-bottom:", "border-left:"} {
		if !strings.Contains(got, side) {
			t.Errorf("all-sides nil-lines: expected %q in %q", side, got)
		}
	}
	if !strings.Contains(got, "solid") {
		t.Errorf("nil line entry should default to solid style, got %q", got)
	}
}

func TestBorderCSS_TopOnly(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Style: style.LineStyleSolid,
		Width: 1,
	}
	b := &style.Border{
		VisibleLines: style.BorderLinesTop,
		Lines: [4]*style.BorderLine{
			nil, bl, nil, nil,
		},
	}
	got := borderCSS(b, 1.0)
	if !strings.Contains(got, "border-top:") {
		t.Errorf("top only: expected border-top, got %q", got)
	}
	// C# HTMLBorder: non-visible sides get explicit "none" declarations to prevent
	// border inheritance in browsers.
	if !strings.Contains(got, "border-bottom:none;") {
		t.Errorf("top only: expected border-bottom:none; for non-visible side, got %q", got)
	}
	if !strings.Contains(got, "border-left:none;") {
		t.Errorf("top only: expected border-left:none; for non-visible side, got %q", got)
	}
	if !strings.Contains(got, "border-right:none;") {
		t.Errorf("top only: expected border-right:none; for non-visible side, got %q", got)
	}
}

func TestBorderCSS_LeftOnly(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Style: style.LineStyleSolid,
		Width: 2,
	}
	b := &style.Border{
		VisibleLines: style.BorderLinesLeft,
		Lines: [4]*style.BorderLine{
			bl, nil, nil, nil,
		},
	}
	got := borderCSS(b, 1.0)
	if !strings.Contains(got, "border-left:") {
		t.Errorf("left only: expected border-left, got %q", got)
	}
	// C# HTMLBorder: non-visible sides get explicit "none" declarations.
	if !strings.Contains(got, "border-right:none;") {
		t.Errorf("left only: expected border-right:none; for non-visible side, got %q", got)
	}
	if !strings.Contains(got, "border-top:none;") {
		t.Errorf("left only: expected border-top:none; for non-visible side, got %q", got)
	}
	if !strings.Contains(got, "border-bottom:none;") {
		t.Errorf("left only: expected border-bottom:none; for non-visible side, got %q", got)
	}
}

func TestBorderCSS_RightOnly(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 255, A: 255},
		Style: style.LineStyleSolid,
		Width: 1,
	}
	b := &style.Border{
		VisibleLines: style.BorderLinesRight,
		Lines: [4]*style.BorderLine{
			nil, nil, bl, nil,
		},
	}
	got := borderCSS(b, 1.0)
	if !strings.Contains(got, "border-right:") {
		t.Errorf("right only: expected border-right, got %q", got)
	}
}

func TestBorderCSS_BottomOnly(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Style: style.LineStyleSolid,
		Width: 1,
	}
	b := &style.Border{
		VisibleLines: style.BorderLinesBottom,
		Lines: [4]*style.BorderLine{
			nil, nil, nil, bl,
		},
	}
	got := borderCSS(b, 1.0)
	if !strings.Contains(got, "border-bottom:") {
		t.Errorf("bottom only: expected border-bottom, got %q", got)
	}
}

func TestBorderCSS_LineStyleDash(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Style: style.LineStyleDash,
		Width: 1,
	}
	b := &style.Border{
		VisibleLines: style.BorderLinesTop,
		Lines:        [4]*style.BorderLine{nil, bl, nil, nil},
	}
	got := borderCSS(b, 1.0)
	if !strings.Contains(got, "dashed") {
		t.Errorf("LineStyleDash: expected 'dashed', got %q", got)
	}
}

func TestBorderCSS_LineStyleDot(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Style: style.LineStyleDot,
		Width: 1,
	}
	b := &style.Border{
		VisibleLines: style.BorderLinesTop,
		Lines:        [4]*style.BorderLine{nil, bl, nil, nil},
	}
	got := borderCSS(b, 1.0)
	if !strings.Contains(got, "dotted") {
		t.Errorf("LineStyleDot: expected 'dotted', got %q", got)
	}
}

func TestBorderCSS_LineStyleDashDot(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Style: style.LineStyleDashDot,
		Width: 1,
	}
	b := &style.Border{
		VisibleLines: style.BorderLinesTop,
		Lines:        [4]*style.BorderLine{nil, bl, nil, nil},
	}
	got := borderCSS(b, 1.0)
	if !strings.Contains(got, "dashed") {
		t.Errorf("LineStyleDashDot: expected 'dashed', got %q", got)
	}
}

func TestBorderCSS_LineStyleDashDotDot(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Style: style.LineStyleDashDotDot,
		Width: 1,
	}
	b := &style.Border{
		VisibleLines: style.BorderLinesTop,
		Lines:        [4]*style.BorderLine{nil, bl, nil, nil},
	}
	got := borderCSS(b, 1.0)
	if !strings.Contains(got, "dashed") {
		t.Errorf("LineStyleDashDotDot: expected 'dashed', got %q", got)
	}
}

func TestBorderCSS_LineStyleDouble(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Style: style.LineStyleDouble,
		Width: 1,
	}
	b := &style.Border{
		VisibleLines: style.BorderLinesTop,
		Lines:        [4]*style.BorderLine{nil, bl, nil, nil},
	}
	got := borderCSS(b, 1.0)
	if !strings.Contains(got, "double") {
		t.Errorf("LineStyleDouble: expected 'double', got %q", got)
	}
}

func TestBorderCSS_LineStyleSolid_Default(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Style: style.LineStyleSolid,
		Width: 1,
	}
	b := &style.Border{
		VisibleLines: style.BorderLinesTop,
		Lines:        [4]*style.BorderLine{nil, bl, nil, nil},
	}
	got := borderCSS(b, 1.0)
	if !strings.Contains(got, "solid") {
		t.Errorf("LineStyleSolid: expected 'solid', got %q", got)
	}
}

func TestBorderCSS_Shadow(t *testing.T) {
	b := &style.Border{
		VisibleLines: style.BorderLinesNone,
		Shadow:       true,
		ShadowWidth:  4,
		ShadowColor:  color.RGBA{R: 0, G: 0, B: 0, A: 128},
	}
	// Shadow is rendered even when VisibleLines==None only if VisibleLines != None.
	// Per the source: returns "" when VisibleLines == None, so we need at least one side.
	b.VisibleLines = style.BorderLinesTop
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Style: style.LineStyleSolid,
		Width: 1,
	}
	b.Lines[int(style.BorderTop)] = bl
	got := borderCSS(b, 1.0)
	if !strings.Contains(got, "box-shadow:") {
		t.Errorf("shadow: expected 'box-shadow:', got %q", got)
	}
}

func TestBorderCSS_Shadow_WithShadowWidth(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Style: style.LineStyleSolid,
		Width: 1,
	}
	b := &style.Border{
		VisibleLines: style.BorderLinesAll,
		Lines:        [4]*style.BorderLine{bl, bl, bl, bl},
		Shadow:       true,
		ShadowWidth:  8,
		ShadowColor:  color.RGBA{R: 50, G: 50, B: 50, A: 200},
	}
	got := borderCSS(b, 2.0)
	// ShadowWidth=8 * scale=2 = 16px
	if !strings.Contains(got, "box-shadow:16.00px 16.00px") {
		t.Errorf("shadow with scale: expected '16.00px 16.00px', got %q", got)
	}
}

func TestBorderCSS_Scale(t *testing.T) {
	// Line width should be multiplied by scale.
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Style: style.LineStyleSolid,
		Width: 2,
	}
	b := &style.Border{
		VisibleLines: style.BorderLinesTop,
		Lines:        [4]*style.BorderLine{nil, bl, nil, nil},
	}
	got := borderCSS(b, 3.0)
	// Width = 2 * 3 = 6px
	if !strings.Contains(got, "6.00px") {
		t.Errorf("scale: expected '6.00px', got %q", got)
	}
}

func TestBorderCSS_NilLineEntry_UsesDefaults(t *testing.T) {
	// When a Lines[idx] is nil, the code uses defaults: 1px solid black.
	b := &style.Border{
		VisibleLines: style.BorderLinesLeft | style.BorderLinesBottom,
		// Lines[0] (Left) = nil → defaults
		// Lines[3] (Bottom) = nil → defaults
	}
	got := borderCSS(b, 1.0)
	if !strings.Contains(got, "border-left:") {
		t.Errorf("nil left line: expected border-left, got %q", got)
	}
	if !strings.Contains(got, "border-bottom:") {
		t.Errorf("nil bottom line: expected border-bottom, got %q", got)
	}
	// Width defaults to 1px, style defaults to solid.
	if !strings.Contains(got, "1.00px") {
		t.Errorf("nil line: expected default '1.00px', got %q", got)
	}
	if !strings.Contains(got, "solid") {
		t.Errorf("nil line: expected default 'solid', got %q", got)
	}
}

// ── imageMIME ─────────────────────────────────────────────────────────────────

func TestImageMIME_JPEG(t *testing.T) {
	data := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	if got := imageMIME(data); got != "image/jpeg" {
		t.Errorf("JPEG magic: got %q, want image/jpeg", got)
	}
}

func TestImageMIME_GIF(t *testing.T) {
	// GIF magic: 'G','I','F'
	data := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}
	if got := imageMIME(data); got != "image/gif" {
		t.Errorf("GIF magic: got %q, want image/gif", got)
	}
}

func TestImageMIME_BMP(t *testing.T) {
	// BMP magic: 'B','M'
	data := []byte{0x42, 0x4D, 0x00, 0x00, 0x00, 0x00}
	if got := imageMIME(data); got != "image/bmp" {
		t.Errorf("BMP magic: got %q, want image/bmp", got)
	}
}

func TestImageMIME_TIFF_LE(t *testing.T) {
	// TIFF little-endian magic: 0x49,0x49,0x2A
	data := []byte{0x49, 0x49, 0x2A, 0x00}
	if got := imageMIME(data); got != "image/tiff" {
		t.Errorf("TIFF LE magic: got %q, want image/tiff", got)
	}
}

func TestImageMIME_TIFF_BE(t *testing.T) {
	// TIFF big-endian magic: 0x4D,0x4D,0x00
	data := []byte{0x4D, 0x4D, 0x00, 0x2A}
	if got := imageMIME(data); got != "image/tiff" {
		t.Errorf("TIFF BE magic: got %q, want image/tiff", got)
	}
}

func TestImageMIME_SVG_Tag(t *testing.T) {
	data := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"/>`)
	if got := imageMIME(data); got != "image/svg+xml" {
		t.Errorf("SVG <svg tag: got %q, want image/svg+xml", got)
	}
}

func TestImageMIME_SVG_XML(t *testing.T) {
	data := []byte(`<?xml version="1.0"?><svg/>`)
	if got := imageMIME(data); got != "image/svg+xml" {
		t.Errorf("SVG <?xml: got %q, want image/svg+xml", got)
	}
}

func TestImageMIME_PNG_Default(t *testing.T) {
	// PNG magic starts with 0x89, 0x50, 0x4E, 0x47 — no special case, falls through to default.
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if got := imageMIME(data); got != "image/png" {
		t.Errorf("PNG (default): got %q, want image/png", got)
	}
}

func TestImageMIME_Unknown_DefaultsPNG(t *testing.T) {
	// Completely unknown bytes → default to image/png.
	data := []byte{0x01, 0x02, 0x03, 0x04}
	if got := imageMIME(data); got != "image/png" {
		t.Errorf("unknown bytes: got %q, want image/png", got)
	}
}

func TestImageMIME_TooShort_DefaultsPNG(t *testing.T) {
	// len(data) < 3 → skips all magic checks → default.
	data := []byte{0xFF, 0xD8}
	if got := imageMIME(data); got != "image/png" {
		t.Errorf("short data: got %q, want image/png", got)
	}
}

// ── hyperlinkHref ──────────────────────────────────────────────────────────────

func TestHyperlinkHref_URL(t *testing.T) {
	// Kind=1 (URL): direct href to external URL.
	got := hyperlinkHref(1, "https://example.com")
	if got != "https://example.com" {
		t.Errorf("URL kind: expected 'https://example.com', got %q", got)
	}
}

func TestHyperlinkHref_PageNumber(t *testing.T) {
	// Kind=2 (PageNumber): link to #PageN{n} within the document.
	got := hyperlinkHref(2, "3")
	if got != "#PageN3" {
		t.Errorf("PageNumber kind: expected '#PageN3', got %q", got)
	}
}

func TestHyperlinkHref_Bookmark(t *testing.T) {
	// Kind=3 (Bookmark): link to #{bookmark} within the document.
	got := hyperlinkHref(3, "MySection")
	if got != "#MySection" {
		t.Errorf("Bookmark kind: expected '#MySection', got %q", got)
	}
}

func TestHyperlinkHref_None(t *testing.T) {
	// Kind=0 (None): no href should be emitted.
	got := hyperlinkHref(0, "something")
	if got != "" {
		t.Errorf("None kind: expected empty string, got %q", got)
	}
}

func TestHyperlinkHref_EmptyValue(t *testing.T) {
	// Any kind with empty value: no href emitted.
	for _, kind := range []int{1, 2, 3} {
		got := hyperlinkHref(kind, "")
		if got != "" {
			t.Errorf("kind=%d empty value: expected empty string, got %q", kind, got)
		}
	}
}
