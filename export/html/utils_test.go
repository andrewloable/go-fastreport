package html

// utils_test.go — internal tests for the HTML utility functions ported from
// C# FastReport.Export.Html.HTMLExportUtils.cs.

import (
	"bytes"
	"image/color"
	"strings"
	"testing"
	"time"

	"github.com/andrewloable/go-fastreport/preview"
)

// ── px ────────────────────────────────────────────────────────────────────────

func TestPx_WholeNumber(t *testing.T) {
	// Integer pixels should format without a decimal point.
	got := px(100)
	if got != "100px;" {
		t.Errorf("px(100) = %q, want %q", got, "100px;")
	}
}

func TestPx_FractionalValue(t *testing.T) {
	// Fractional pixels: trailing zeros stripped.
	got := px(28.35)
	// FloatToString(28.35, 2) = "28.35"
	if got != "28.35px;" {
		t.Errorf("px(28.35) = %q, want %q", got, "28.35px;")
	}
}

func TestPx_Zero(t *testing.T) {
	got := px(0)
	if got != "0px;" {
		t.Errorf("px(0) = %q, want %q", got, "0px;")
	}
}

func TestPx_TrailingZeroStripped(t *testing.T) {
	// 12.50 → "12.5px;"
	got := px(12.50)
	if got != "12.5px;" {
		t.Errorf("px(12.50) = %q, want %q", got, "12.5px;")
	}
}

// ── SizeValue ─────────────────────────────────────────────────────────────────

func TestSizeValue_Pixel(t *testing.T) {
	got := SizeValue(150, 0, HtmlSizeUnitsPixel)
	if got != "150px;" {
		t.Errorf("SizeValue pixel: got %q, want %q", got, "150px;")
	}
}

func TestSizeValue_Percent(t *testing.T) {
	// 50 / 200 * 100 = 25%
	got := SizeValue(50, 200, HtmlSizeUnitsPercent)
	if got != "25%" {
		t.Errorf("SizeValue percent: got %q, want %q", got, "25%")
	}
}

func TestSizeValue_Percent_Rounding(t *testing.T) {
	// 1 / 3 * 100 = 33.333... → rounds to 33%
	got := SizeValue(1, 3, HtmlSizeUnitsPercent)
	if got != "33%" {
		t.Errorf("SizeValue percent rounding: got %q, want %q", got, "33%")
	}
}

func TestSizeValue_Default(t *testing.T) {
	// Any unit other than Pixel/Percent falls through to plain float string.
	got := SizeValue(42.5, 100, HtmlSizeUnits(99))
	if got != "42.5" {
		t.Errorf("SizeValue default: got %q, want %q", got, "42.5")
	}
}

// ── WriteMHTHeader ────────────────────────────────────────────────────────────

func TestWriteMHTHeader_ContainsRequiredHeaders(t *testing.T) {
	var buf bytes.Buffer
	now := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)
	err := WriteMHTHeader(&buf, "MyReport.html", "boundary_abc123", now)
	if err != nil {
		t.Fatalf("WriteMHTHeader error: %v", err)
	}
	out := buf.String()

	// Must have From: header with encoded filename.
	if !strings.Contains(out, "From: =?utf-8?B?") {
		t.Errorf("WriteMHTHeader: expected encoded From: header, got:\n%s", out)
	}
	// Subject: same encoded form.
	if !strings.Contains(out, "Subject: =?utf-8?B?") {
		t.Errorf("WriteMHTHeader: expected encoded Subject: header, got:\n%s", out)
	}
	// MIME-Version.
	if !strings.Contains(out, "MIME-Version: 1.0") {
		t.Errorf("WriteMHTHeader: expected MIME-Version header, got:\n%s", out)
	}
	// Content-Type with boundary.
	if !strings.Contains(out, "boundary_abc123") {
		t.Errorf("WriteMHTHeader: expected boundary in Content-Type, got:\n%s", out)
	}
	// MIME preamble text.
	if !strings.Contains(out, "This is a multi-part message in MIME format.") {
		t.Errorf("WriteMHTHeader: expected preamble text, got:\n%s", out)
	}
}

func TestWriteMHTHeader_EncodesFilename(t *testing.T) {
	var buf bytes.Buffer
	now := time.Now()
	// Use a filename with non-ASCII to verify encoding.
	err := WriteMHTHeader(&buf, "Отчёт.html", "b", now)
	if err != nil {
		t.Fatalf("WriteMHTHeader error: %v", err)
	}
	out := buf.String()
	// Must contain RFC 2047 encoded-word marker.
	if !strings.Contains(out, "=?utf-8?B?") {
		t.Errorf("WriteMHTHeader: expected encoded-word syntax, got:\n%s", out)
	}
	// Must end with ?= to close encoded-word.
	if !strings.Contains(out, "?=") {
		t.Errorf("WriteMHTHeader: expected closing ?=, got:\n%s", out)
	}
}

// ── WriteMimePart ─────────────────────────────────────────────────────────────

func TestWriteMimePart_HTMLPartUsesQuotedPrintable(t *testing.T) {
	var buf bytes.Buffer
	data := []byte("<html><body>Hello</body></html>")
	err := WriteMimePart(&buf, data, "text/html", "utf-8", "index.html", "bndry")
	if err != nil {
		t.Fatalf("WriteMimePart error: %v", err)
	}
	out := buf.String()
	// Must use quoted-printable for text/html.
	if !strings.Contains(out, "Content-Transfer-Encoding: quoted-printable") {
		t.Errorf("WriteMimePart text/html: expected quoted-printable encoding, got:\n%s", out)
	}
	// Boundary marker present.
	if !strings.Contains(out, "--bndry") {
		t.Errorf("WriteMimePart: expected boundary marker, got:\n%s", out)
	}
	// Content-Location present.
	if !strings.Contains(out, "Content-Location: index.html") {
		t.Errorf("WriteMimePart: expected Content-Location header, got:\n%s", out)
	}
}

func TestWriteMimePart_ImagePartUsesBase64(t *testing.T) {
	var buf bytes.Buffer
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG magic
	err := WriteMimePart(&buf, data, "image/png", "", "img.png", "bndry")
	if err != nil {
		t.Fatalf("WriteMimePart error: %v", err)
	}
	out := buf.String()
	// Must use base64 for non-text/html content.
	if !strings.Contains(out, "Content-Transfer-Encoding: base64") {
		t.Errorf("WriteMimePart image: expected base64 encoding, got:\n%s", out)
	}
	// No charset header when charset is empty.
	if strings.Contains(out, "charset") {
		t.Errorf("WriteMimePart: unexpected charset header for empty charset, got:\n%s", out)
	}
}

func TestWriteMimePart_WithCharset(t *testing.T) {
	var buf bytes.Buffer
	data := []byte("<html/>")
	err := WriteMimePart(&buf, data, "text/html", "utf-8", "page.html", "bndry")
	if err != nil {
		t.Fatalf("WriteMimePart error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `charset="utf-8"`) {
		t.Errorf("WriteMimePart: expected charset header, got:\n%s", out)
	}
}

func TestWriteMimePart_WithoutCharset(t *testing.T) {
	var buf bytes.Buffer
	data := []byte{0xFF, 0xD8} // JPEG magic
	err := WriteMimePart(&buf, data, "image/jpeg", "", "photo.jpg", "bndry")
	if err != nil {
		t.Fatalf("WriteMimePart error: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "charset") {
		t.Errorf("WriteMimePart: unexpected charset for empty string, got:\n%s", out)
	}
}

// ── insertBase64LineBreaks ────────────────────────────────────────────────────

func TestInsertBase64LineBreaks_ShortString(t *testing.T) {
	// String shorter than 76 chars → no line breaks inserted.
	s := "ABCDE"
	got := insertBase64LineBreaks(s)
	if got != s {
		t.Errorf("insertBase64LineBreaks short: got %q, want %q", got, s)
	}
}

func TestInsertBase64LineBreaks_ExactlyOneLine(t *testing.T) {
	// Exactly 76 chars → no break needed.
	s := strings.Repeat("A", 76)
	got := insertBase64LineBreaks(s)
	if got != s {
		t.Errorf("insertBase64LineBreaks 76 chars: got unexpected break, want no break")
	}
}

func TestInsertBase64LineBreaks_TwoLines(t *testing.T) {
	// 77 chars → first 76 + "\r\n" + last char.
	s := strings.Repeat("A", 77)
	got := insertBase64LineBreaks(s)
	want := strings.Repeat("A", 76) + "\r\n" + "A"
	if got != want {
		t.Errorf("insertBase64LineBreaks 77 chars:\ngot  %q\nwant %q", got, want)
	}
}

func TestInsertBase64LineBreaks_MultipleLines(t *testing.T) {
	// 200 chars → three lines: 76 + 76 + 48.
	s := strings.Repeat("X", 200)
	got := insertBase64LineBreaks(s)
	parts := strings.Split(got, "\r\n")
	if len(parts) != 3 {
		t.Errorf("insertBase64LineBreaks 200 chars: expected 3 lines, got %d: %v", len(parts), parts)
	}
	if len(parts[0]) != 76 {
		t.Errorf("insertBase64LineBreaks: first line length %d, want 76", len(parts[0]))
	}
	if len(parts[1]) != 76 {
		t.Errorf("insertBase64LineBreaks: second line length %d, want 76", len(parts[1]))
	}
	if len(parts[2]) != 48 {
		t.Errorf("insertBase64LineBreaks: third line length %d, want 48", len(parts[2]))
	}
}

// ── HTMLPageData ──────────────────────────────────────────────────────────────

func TestNewHTMLPageData_PageNumber(t *testing.T) {
	d := NewHTMLPageData(3)
	if d.PageNumber != 3 {
		t.Errorf("NewHTMLPageData: PageNumber = %d, want 3", d.PageNumber)
	}
}

func TestNewHTMLPageData_EmptyFields(t *testing.T) {
	d := NewHTMLPageData(1)
	if d.Width != 0 || d.Height != 0 {
		t.Errorf("NewHTMLPageData: expected zero Width/Height, got %v/%v", d.Width, d.Height)
	}
	if len(d.Pictures) != 0 || len(d.Guids) != 0 {
		t.Errorf("NewHTMLPageData: expected empty Pictures/Guids slices")
	}
}

func TestHTMLPageData_AddPicture(t *testing.T) {
	d := NewHTMLPageData(1)
	data1 := []byte{0x01, 0x02}
	data2 := []byte{0x03, 0x04}
	d.AddPicture(data1, "hash1")
	d.AddPicture(data2, "hash2")

	if len(d.Pictures) != 2 {
		t.Errorf("AddPicture: expected 2 pictures, got %d", len(d.Pictures))
	}
	if len(d.Guids) != 2 {
		t.Errorf("AddPicture: expected 2 guids, got %d", len(d.Guids))
	}
	if d.Guids[0] != "hash1" || d.Guids[1] != "hash2" {
		t.Errorf("AddPicture: wrong guids: %v", d.Guids)
	}
}

func TestHTMLPageData_CSSTextAndPageText(t *testing.T) {
	d := NewHTMLPageData(1)
	d.CSSText.WriteString(".s0 { color:red; }")
	d.PageText.WriteString("<div>content</div>")

	if d.CSSText.String() != ".s0 { color:red; }" {
		t.Errorf("CSSText: got %q", d.CSSText.String())
	}
	if d.PageText.String() != "<div>content</div>" {
		t.Errorf("PageText: got %q", d.PageText.String())
	}
}

// ── Enums ─────────────────────────────────────────────────────────────────────

func TestHTMLExportFormat_Values(t *testing.T) {
	if HTMLExportFormatMessageHTML != 0 {
		t.Errorf("HTMLExportFormatMessageHTML: expected 0, got %d", HTMLExportFormatMessageHTML)
	}
	if HTMLExportFormatHTML != 1 {
		t.Errorf("HTMLExportFormatHTML: expected 1, got %d", HTMLExportFormatHTML)
	}
}

func TestHTMLImageFormat_Values(t *testing.T) {
	if HTMLImageFormatBmp != 0 {
		t.Errorf("HTMLImageFormatBmp: expected 0, got %d", HTMLImageFormatBmp)
	}
	if HTMLImageFormatPng != 1 {
		t.Errorf("HTMLImageFormatPng: expected 1, got %d", HTMLImageFormatPng)
	}
	if HTMLImageFormatJpeg != 2 {
		t.Errorf("HTMLImageFormatJpeg: expected 2, got %d", HTMLImageFormatJpeg)
	}
	if HTMLImageFormatGif != 3 {
		t.Errorf("HTMLImageFormatGif: expected 3, got %d", HTMLImageFormatGif)
	}
}

func TestHtmlSizeUnits_Values(t *testing.T) {
	if HtmlSizeUnitsPixel != 0 {
		t.Errorf("HtmlSizeUnitsPixel: expected 0, got %d", HtmlSizeUnitsPixel)
	}
	if HtmlSizeUnitsPercent != 1 {
		t.Errorf("HtmlSizeUnitsPercent: expected 1, got %d", HtmlSizeUnitsPercent)
	}
}

// ── cssAttr / cssClass (internal Exporter helpers) ────────────────────────────

// buildMinimalExporter returns an Exporter initialized enough to call cssAttr/cssClass.
func buildMinimalExporter() *Exporter {
	e := NewExporter()
	e.css = newCSSRegistry()
	return e
}

func TestCSSAttr_NormalMode_RegistersClass(t *testing.T) {
	e := buildMinimalExporter()
	attr := e.cssAttr("color:red;")
	// Should return class="sN" attribute.
	if !strings.HasPrefix(attr, `class="`) {
		t.Errorf("cssAttr normal: expected class= attribute, got %q", attr)
	}
	// The class name should be registered in the CSS registry.
	if e.css.Count() != 1 {
		t.Errorf("cssAttr normal: expected 1 registered class, got %d", e.css.Count())
	}
}

func TestCSSAttr_InlineStyles_ReturnsStyleAttr(t *testing.T) {
	e := buildMinimalExporter()
	e.InlineStyles = true
	attr := e.cssAttr("color:blue;")
	// Should return style="..." attribute.
	want := `style="color:blue;"`
	if attr != want {
		t.Errorf("cssAttr inline: got %q, want %q", attr, want)
	}
	// Should NOT register in the CSS registry.
	if e.css.Count() != 0 {
		t.Errorf("cssAttr inline: should not register class, got %d registered", e.css.Count())
	}
}

func TestCSSAttr_EmptyCSS_ReturnsEmpty(t *testing.T) {
	e := buildMinimalExporter()
	attr := e.cssAttr("")
	if attr != "" {
		t.Errorf("cssAttr empty: expected empty, got %q", attr)
	}
}

func TestCSSClass_NormalMode_ReturnsClassName(t *testing.T) {
	e := buildMinimalExporter()
	name := e.cssClass("font-weight:bold;")
	if name == "" {
		t.Errorf("cssClass normal: expected class name, got empty")
	}
	// Should start with "s".
	if name[0] != 's' {
		t.Errorf("cssClass normal: expected name starting with 's', got %q", name)
	}
}

func TestCSSClass_InlineStyles_ReturnsEmpty(t *testing.T) {
	e := buildMinimalExporter()
	e.InlineStyles = true
	name := e.cssClass("font-weight:bold;")
	if name != "" {
		t.Errorf("cssClass inline: expected empty string, got %q", name)
	}
}

// ── InlineStyles integration ───────────────────────────────────────────────────

// TestExporter_InlineStyles_BandBackground verifies that InlineStyles=true causes
// renderBandBackground to emit a style= attribute rather than a class= attribute.
func TestExporter_InlineStyles_BandBackground(t *testing.T) {
	pp := preview.New()
	pp.AddPage(200, 300, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:      "Band",
		Top:       0,
		Height:    50,
		Width:     200,
		FillColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
	})

	exp := NewExporter()
	exp.InlineStyles = true

	var buf strings.Builder
	type writerFunc func(p []byte) (n int, err error)
	w := &byteWriter{buf: &buf}
	if err := exp.Export(pp, w); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	// In InlineStyles mode, band backgrounds emit style= not class=.
	if !strings.Contains(out, `style="`) {
		t.Errorf("InlineStyles band background: expected style= attribute in output, got:\n%s", out)
	}
}

// TestExporter_InlineStyles_TextObject verifies that InlineStyles=true causes
// renderTextObject to emit style= attributes directly rather than class= references.
func TestExporter_InlineStyles_TextObject(t *testing.T) {
	pp := preview.New()
	pp.AddPage(200, 300, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 50,
		Width:  200,
		Objects: []preview.PreparedObject{
			{
				Kind:   preview.ObjectTypeText,
				Text:   "Hello",
				Left:   0,
				Top:    0,
				Width:  100,
				Height: 20,
			},
		},
	})

	exp := NewExporter()
	exp.InlineStyles = true

	var buf strings.Builder
	w := &byteWriter{buf: &buf}
	if err := exp.Export(pp, w); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	// In InlineStyles mode, text objects must not emit class= references.
	if strings.Contains(out, `class="s`) {
		t.Errorf("InlineStyles text object: unexpected class=sN in output:\n%s", out)
	}
	// Must contain the text content.
	if !strings.Contains(out, "Hello") {
		t.Errorf("InlineStyles text object: expected 'Hello' in output:\n%s", out)
	}
}

// TestExporter_InlineStyles_TextObject_Hyperlink verifies that hyperlink text
// objects with InlineStyles=true also emit style= attributes and no class= refs.
func TestExporter_InlineStyles_TextObject_Hyperlink(t *testing.T) {
	pp := preview.New()
	pp.AddPage(200, 300, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 50,
		Width:  200,
		Objects: []preview.PreparedObject{
			{
				Kind:           preview.ObjectTypeText,
				Text:           "Click",
				Left:           0,
				Top:            0,
				Width:          100,
				Height:         20,
				HyperlinkKind:  1,
				HyperlinkValue: "https://example.com",
			},
		},
	})

	exp := NewExporter()
	exp.InlineStyles = true

	var buf strings.Builder
	w := &byteWriter{buf: &buf}
	if err := exp.Export(pp, w); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	// Must emit an anchor tag.
	if !strings.Contains(out, `<a `) {
		t.Errorf("InlineStyles hyperlink: expected <a> tag in output:\n%s", out)
	}
	// Must not emit class=sN references.
	if strings.Contains(out, `class="s`) {
		t.Errorf("InlineStyles hyperlink: unexpected class=sN in output:\n%s", out)
	}
}

// byteWriter adapts a strings.Builder to io.Writer.
type byteWriter struct {
	buf *strings.Builder
}

func (w *byteWriter) Write(p []byte) (int, error) {
	n, err := w.buf.Write(p)
	return n, err
}
