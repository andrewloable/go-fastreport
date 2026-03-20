// html_coverage3_test.go — tests that improve coverage for:
//   - fontMeasureRatioFloat (html.go:1047) — all font family branches
//   - formatTextContent (html.go:1104) — \v replacement, \r\n normalization,
//     leading newline, subsequent newlines, HTML entity escaping, trailing space,
//     single-char space
//   - resizeImagePNG (html.go:1268) — non-PNG data, zero target, same-size shortcut
//   - imageMIMEForCSS (html.go:1319) — BMP, TIFF LE/BE, SVG detection, short data

package html

import (
	"bytes"
	"image"
	"image/png"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
)

// ── fontMeasureRatioFloat ──────────────────────────────────────────────────────

func TestFontMeasureRatioFloat_AllCases(t *testing.T) {
	tests := []struct {
		font string
		want float64
	}{
		{"Arial", 2355.0 / 2048.0},
		{"arial narrow", 2355.0 / 2048.0},
		{"Times New Roman", 2355.0 / 2048.0},
		{"times", 2355.0 / 2048.0},
		{"Tahoma", 2472.0 / 2048.0},
		{"Microsoft Sans Serif", 2472.0 / 2048.0},
		{"Verdana", 2489.0 / 2048.0},
		{"Arial Unicode MS", 2743.0 / 2048.0},
		{"Arial Black", 2899.0 / 2048.0},
		{"Georgia", 2327.0 / 2048.0},
		{"Courier New", 2320.0 / 2048.0},
		{"Courier", 2320.0 / 2048.0},
		{"Segoe UI", 2724.0 / 2048.0},
		{"UnknownFont", 2472.0 / 2048.0}, // default
		{"", 2472.0 / 2048.0},            // default for empty
	}
	for _, tt := range tests {
		got := fontMeasureRatioFloat(tt.font)
		if got != tt.want {
			t.Errorf("fontMeasureRatioFloat(%q) = %v, want %v", tt.font, got, tt.want)
		}
	}
}

// ── formatTextContent ──────────────────────────────────────────────────────────

func TestFormatTextContent_PlainText(t *testing.T) {
	// No special characters, no line breaks.
	got := formatTextContent("Hello world", 15)
	if got != "Hello world" {
		t.Errorf("plain text: got %q, want %q", got, "Hello world")
	}
}

func TestFormatTextContent_VerticalTab(t *testing.T) {
	// \v should be replaced with \n.
	got := formatTextContent("A\vB", 15)
	// A followed by \n → lineBreakCount=0, then B resets it.
	if !strings.Contains(got, `<p style="margin-top:0px;margin-bottom:0px;"></p>`) {
		t.Errorf("vertical tab: expected line break p tag, got %q", got)
	}
	if !strings.Contains(got, "B") {
		t.Errorf("vertical tab: expected 'B' after break, got %q", got)
	}
}

func TestFormatTextContent_CRLF(t *testing.T) {
	// \r\n should be normalized to \n (single line break, not two).
	got := formatTextContent("X\r\nY", 15)
	// X, then linebreak, then Y.
	count := strings.Count(got, "<p style=")
	if count != 1 {
		t.Errorf("CRLF: expected 1 line break, got %d p tags", count)
	}
}

func TestFormatTextContent_CROnly(t *testing.T) {
	// \r alone should be treated as \n.
	got := formatTextContent("X\rY", 15)
	if !strings.Contains(got, `<p style="margin-top:0px;margin-bottom:0px;"></p>`) {
		t.Errorf("CR only: expected line break, got %q", got)
	}
}

func TestFormatTextContent_LeadingNewline(t *testing.T) {
	// \n at position 0 → extra p tag with margin-top:fontSize.
	got := formatTextContent("\nHello", 15)
	if !strings.Contains(got, `margin-top:15px;`) {
		t.Errorf("leading newline: expected margin-top with fontSize, got %q", got)
	}
}

func TestFormatTextContent_MultipleNewlines(t *testing.T) {
	// Multiple consecutive newlines: first is lineBreakCount=0, subsequent >0.
	got := formatTextContent("A\n\n\nB", 15)
	// First \n: lineBreakCount==0 → margin-top:0px
	// Second \n: lineBreakCount==1 → height:15px
	// Third \n: lineBreakCount==2 → height:15px
	if !strings.Contains(got, `height:15px;`) {
		t.Errorf("multiple newlines: expected height:fontSize, got %q", got)
	}
}

func TestFormatTextContent_HTMLEntities(t *testing.T) {
	got := formatTextContent(`<b>&"test"</b>`, 15)
	if !strings.Contains(got, "&lt;") {
		t.Errorf("expected &lt;, got %q", got)
	}
	if !strings.Contains(got, "&gt;") {
		t.Errorf("expected &gt;, got %q", got)
	}
	if !strings.Contains(got, "&amp;") {
		t.Errorf("expected &amp;, got %q", got)
	}
	if !strings.Contains(got, "&quot;") {
		t.Errorf("expected &quot;, got %q", got)
	}
}

func TestFormatTextContent_TrailingSpace(t *testing.T) {
	// Trailing space → &nbsp;
	got := formatTextContent("Hello ", 15)
	if !strings.Contains(got, "&nbsp;") {
		t.Errorf("trailing space: expected &nbsp;, got %q", got)
	}
}

func TestFormatTextContent_ConsecutiveSpaces(t *testing.T) {
	// Two consecutive spaces → both become &nbsp;
	got := formatTextContent("A  B", 15)
	if !strings.Contains(got, "&nbsp;&nbsp;") {
		t.Errorf("consecutive spaces: expected &nbsp;&nbsp;, got %q", got)
	}
}

func TestFormatTextContent_SingleCharSpace(t *testing.T) {
	// A single space as the entire text → &nbsp; (n==1 case).
	got := formatTextContent(" ", 15)
	if got != "&nbsp;" {
		t.Errorf("single space: got %q, want %q", got, "&nbsp;")
	}
}

// ── resizeImagePNG ─────────────────────────────────────────────────────────────

func TestResizeImagePNG_NonPNGData(t *testing.T) {
	// Non-PNG data should be returned as-is.
	data := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46} // JPEG header
	got := resizeImagePNG(data, 100, 100)
	if !bytes.Equal(got, data) {
		t.Error("non-PNG data: expected original data returned unchanged")
	}
}

func TestResizeImagePNG_ZeroTargetWidth(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	data := buf.Bytes()

	got := resizeImagePNG(data, 0, 100)
	if !bytes.Equal(got, data) {
		t.Error("zero target width: expected original data returned")
	}
}

func TestResizeImagePNG_ZeroTargetHeight(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	data := buf.Bytes()

	got := resizeImagePNG(data, 100, 0)
	if !bytes.Equal(got, data) {
		t.Error("zero target height: expected original data returned")
	}
}

func TestResizeImagePNG_SameSize(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	data := buf.Bytes()

	got := resizeImagePNG(data, 50, 50)
	if !bytes.Equal(got, data) {
		t.Error("same size: expected original data returned")
	}
}

func TestResizeImagePNG_ShortData(t *testing.T) {
	// Data shorter than PNG header (< 8 bytes).
	data := []byte{0x89, 0x50}
	got := resizeImagePNG(data, 100, 100)
	if !bytes.Equal(got, data) {
		t.Error("short data: expected original data returned")
	}
}

func TestResizeImagePNG_Resize(t *testing.T) {
	// Create a 100x50 PNG, resize to 50x25.
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	data := buf.Bytes()

	got := resizeImagePNG(data, 50, 25)
	// Should not be the same data (resized).
	if bytes.Equal(got, data) {
		t.Error("resize: expected different data after resize")
	}
	// Should be valid PNG.
	_, err := png.Decode(bytes.NewReader(got))
	if err != nil {
		t.Errorf("resize: result is not valid PNG: %v", err)
	}
}

// ── imageMIMEForCSS ────────────────────────────────────────────────────────────

func TestImageMIMEForCSS_JPEG(t *testing.T) {
	data := []byte{0xFF, 0xD8, 0xFF}
	if got := imageMIMEForCSS(data); got != "image/Jpeg" {
		t.Errorf("JPEG: got %q, want image/Jpeg", got)
	}
}

func TestImageMIMEForCSS_GIF(t *testing.T) {
	data := []byte{0x47, 0x49, 0x46, 0x38}
	if got := imageMIMEForCSS(data); got != "image/Gif" {
		t.Errorf("GIF: got %q, want image/Gif", got)
	}
}

func TestImageMIMEForCSS_BMP(t *testing.T) {
	data := []byte{0x42, 0x4D, 0x00, 0x00}
	if got := imageMIMEForCSS(data); got != "image/Bmp" {
		t.Errorf("BMP: got %q, want image/Bmp", got)
	}
}

func TestImageMIMEForCSS_TIFF_LE(t *testing.T) {
	data := []byte{0x49, 0x49, 0x2A, 0x00}
	if got := imageMIMEForCSS(data); got != "image/Tiff" {
		t.Errorf("TIFF LE: got %q, want image/Tiff", got)
	}
}

func TestImageMIMEForCSS_TIFF_BE(t *testing.T) {
	data := []byte{0x4D, 0x4D, 0x00, 0x2A}
	if got := imageMIMEForCSS(data); got != "image/Tiff" {
		t.Errorf("TIFF BE: got %q, want image/Tiff", got)
	}
}

func TestImageMIMEForCSS_SVG_Tag(t *testing.T) {
	data := []byte("<svg xmlns='http://www.w3.org/2000/svg'>...</svg>")
	if got := imageMIMEForCSS(data); got != "image/svg+xml" {
		t.Errorf("SVG tag: got %q, want image/svg+xml", got)
	}
}

func TestImageMIMEForCSS_SVG_XML(t *testing.T) {
	data := []byte("<?xml version='1.0'?><svg>...</svg>")
	if got := imageMIMEForCSS(data); got != "image/svg+xml" {
		t.Errorf("SVG XML: got %q, want image/svg+xml", got)
	}
}

func TestImageMIMEForCSS_PNG_Default(t *testing.T) {
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if got := imageMIMEForCSS(data); got != "image/Png" {
		t.Errorf("PNG: got %q, want image/Png", got)
	}
}

func TestImageMIMEForCSS_Unknown(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	if got := imageMIMEForCSS(data); got != "image/Png" {
		t.Errorf("Unknown: got %q, want image/Png (default)", got)
	}
}

func TestImageMIMEForCSS_TooShort(t *testing.T) {
	data := []byte{0xFF, 0xD8}
	if got := imageMIMEForCSS(data); got != "image/Png" {
		t.Errorf("Too short: got %q, want image/Png (default)", got)
	}
}

// ── renderObjectLayered: z-index injection via style= fallback ────────────────

func TestRenderObjectLayered_NonTextObject_ZIndexInStyle(t *testing.T) {
	// Non-text objects rendered via renderObject don't include "position:absolute;"
	// in the inline style. renderObjectLayered should inject z-index via the style=
	// fallback path (the else branch at line 322-324).
	e := NewExporter()
	e.css = newCSSRegistry()
	e.pp = nil

	// Use a shape with no fill, no border — simplest possible non-text object.
	obj := preview.PreparedObject{
		Name:      "TestShape",
		Kind:      preview.ObjectTypeShape,
		Left:      10,
		Top:       20,
		Width:     50,
		Height:    50,
		ShapeKind: 0, // rectangle
	}
	e.renderObjectLayered(obj, 1.0)
	rendered := e.sb.String()

	if !strings.Contains(rendered, "z-index:1;") {
		t.Errorf("expected z-index:1; in rendered output, got %q", rendered)
	}
	// Verify the z-index was injected into the style= attribute (not after position:absolute).
	// The rendered div should look like: <div style="z-index:1;left:...;...
	if strings.Contains(rendered, "position:absolute;z-index:") {
		t.Errorf("z-index should NOT follow position:absolute; for non-text objects, got %q", rendered)
	}
}
