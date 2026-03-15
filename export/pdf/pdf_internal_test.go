package pdf

// Internal tests (package pdf) cover unexported functions that are not
// reachable through the exported Export() API.

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// ── pdfWrapText / pdfEstimateTextWidth ────────────────────────────────────────

func TestPdfWrapText_BasicWordWrap(t *testing.T) {
	// Words that exceed maxWidth are split onto separate lines.
	lines := pdfWrapText("hello world foo bar", 30, 10, true) // 30pt ≈ 5 chars
	if len(lines) < 2 {
		t.Errorf("expected multiple lines for narrow width, got %d: %v", len(lines), lines)
	}
}

func TestPdfWrapText_NoWordWrap(t *testing.T) {
	// wordWrap=false: paragraph is returned as-is.
	lines := pdfWrapText("hello world foo bar", 5, 10, false)
	if len(lines) != 1 {
		t.Errorf("expected 1 line for no-wrap, got %d: %v", len(lines), lines)
	}
	if lines[0] != "hello world foo bar" {
		t.Errorf("unexpected content: %q", lines[0])
	}
}

func TestPdfWrapText_HardNewlines(t *testing.T) {
	// Hard newlines always produce line breaks.
	lines := pdfWrapText("line1\nline2\nline3", 1000, 10, true)
	if len(lines) != 3 {
		t.Errorf("expected 3 lines for \\n-separated text, got %d: %v", len(lines), lines)
	}
}

func TestPdfWrapText_EmptyParagraph(t *testing.T) {
	// Empty para (consecutive \n) produces an empty string entry.
	lines := pdfWrapText("a\n\nb", 1000, 10, true)
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	if lines[1] != "" {
		t.Errorf("middle line should be empty, got %q", lines[1])
	}
}

func TestPdfWrapText_SpacesOnlyParagraph(t *testing.T) {
	// Para containing only whitespace: Fields() returns [], so words==0.
	lines := pdfWrapText("a\n   \nb", 1000, 10, true)
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	// Middle entry should be empty (Fields of "   " is []).
	if lines[1] != "" {
		t.Errorf("whitespace-only para should produce empty line, got %q", lines[1])
	}
}

func TestPdfWrapText_CRLF(t *testing.T) {
	// \r\n should be normalised to \n.
	lines := pdfWrapText("a\r\nb", 1000, 10, true)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines for CRLF, got %d: %v", len(lines), lines)
	}
}

func TestPdfWrapText_Empty(t *testing.T) {
	lines := pdfWrapText("", 100, 10, true)
	// Single empty paragraph.
	if len(lines) != 1 {
		t.Errorf("expected 1 line for empty text, got %d", len(lines))
	}
}

// ── pdfEstimateTextWidth ──────────────────────────────────────────────────────

func TestPdfEstimateTextWidth_Basic(t *testing.T) {
	w := pdfEstimateTextWidth("hello", 10)
	// 5 chars × 10pt × 0.6 = 30
	if w != 30 {
		t.Errorf("expected 30, got %.2f", w)
	}
}

func TestPdfEstimateTextWidth_Empty(t *testing.T) {
	w := pdfEstimateTextWidth("", 12)
	if w != 0 {
		t.Errorf("expected 0 for empty string, got %.2f", w)
	}
}

// ── pdfFontAlias ──────────────────────────────────────────────────────────────

func TestPdfFontAlias_Helvetica(t *testing.T) {
	// Default (Arial / sans) variants.
	tests := []struct {
		bold, italic bool
		want         string
	}{
		{false, false, "F1"},
		{true, false, "F2"},
		{false, true, "F3"},
		{true, true, "F4"},
	}
	for _, tt := range tests {
		got := pdfFontAlias("Arial", tt.bold, tt.italic)
		if got != tt.want {
			t.Errorf("pdfFontAlias(Arial, %v, %v) = %q, want %q", tt.bold, tt.italic, got, tt.want)
		}
	}
}

func TestPdfFontAlias_Times(t *testing.T) {
	tests := []struct {
		bold, italic bool
		want         string
	}{
		{false, false, "F5"},
		{true, false, "F6"},
		{false, true, "F7"},
		{true, true, "F8"},
	}
	for _, tt := range tests {
		got := pdfFontAlias("Times New Roman", tt.bold, tt.italic)
		if got != tt.want {
			t.Errorf("pdfFontAlias(Times, %v, %v) = %q, want %q", tt.bold, tt.italic, got, tt.want)
		}
	}
}

func TestPdfFontAlias_Courier(t *testing.T) {
	tests := []struct {
		bold, italic bool
		want         string
	}{
		{false, false, "F9"},
		{true, false, "F10"},
		{false, true, "F11"},
		{true, true, "F12"},
	}
	for _, tt := range tests {
		got := pdfFontAlias("Courier New", tt.bold, tt.italic)
		if got != tt.want {
			t.Errorf("pdfFontAlias(Courier, %v, %v) = %q, want %q", tt.bold, tt.italic, got, tt.want)
		}
	}
}

// ── familyKeywordPDF ──────────────────────────────────────────────────────────

func TestFamilyKeywordPDF_MonoVariants(t *testing.T) {
	monoNames := []string{
		"Courier New", "Consolas", "Monospace", "Monaco",
		"Lucida Console", "Cascadia Code",
	}
	for _, name := range monoNames {
		got := familyKeywordPDF(name)
		if got != "mono" {
			t.Errorf("familyKeywordPDF(%q) = %q, want %q", name, got, "mono")
		}
	}
}

func TestFamilyKeywordPDF_Sans(t *testing.T) {
	got := familyKeywordPDF("Arial")
	if got != "sans" {
		t.Errorf("familyKeywordPDF(Arial) = %q, want sans", got)
	}
}

// ── ttfDataFor ────────────────────────────────────────────────────────────────

func TestTtfDataFor_AllVariants(t *testing.T) {
	tests := []struct {
		family     string
		bold, ital bool
		wantName   string
	}{
		{"sans", false, false, "GoRegular"},
		{"sans", true, false, "GoBold"},
		{"sans", false, true, "GoItalic"},
		{"sans", true, true, "GoBoldItalic"},
		{"mono", false, false, "GoMono"},
		{"mono", true, false, "GoMonoBold"},
		{"mono", false, true, "GoMonoItalic"},
		{"mono", true, true, "GoMonoBoldItalic"},
	}
	for _, tt := range tests {
		data, name := ttfDataFor(tt.family, tt.bold, tt.ital)
		if len(data) == 0 {
			t.Errorf("ttfDataFor(%q, %v, %v): empty data", tt.family, tt.bold, tt.ital)
		}
		if name != tt.wantName {
			t.Errorf("ttfDataFor(%q, %v, %v) name = %q, want %q", tt.family, tt.bold, tt.ital, name, tt.wantName)
		}
	}
}

// ── measureText ──────────────────────────────────────────────────────────────

func TestMeasureText_NonEFPrefix(t *testing.T) {
	// A non-EF alias (e.g. pdfFontAlias returns "F1") uses pdfEstimateTextWidth.
	w := NewWriter()
	fm := NewPDFFontManager(w)
	exp := &Exporter{fontMgr: fm}

	// Use alias "F1" which doesn't start with "EF".
	width := exp.measureText("F1", "hello", 10)
	expected := pdfEstimateTextWidth("hello", 10)
	if width != expected {
		t.Errorf("measureText(F1,...) = %.2f, want %.2f", width, expected)
	}
}

func TestMeasureText_EFPrefix_NilFontMgr(t *testing.T) {
	// If fontMgr is nil and alias has EF prefix, falls back to estimate.
	exp := &Exporter{fontMgr: nil}
	width := exp.measureText("EF0", "hello", 10)
	// With nil fontMgr, the EF branch is skipped and pdfEstimateTextWidth is used.
	expected := pdfEstimateTextWidth("hello", 10)
	if width != expected {
		t.Errorf("measureText(EF0, nil fontMgr) = %.2f, want %.2f", width, expected)
	}
}

// ── lookupFont ────────────────────────────────────────────────────────────────

func TestLookupFont_UnknownAlias(t *testing.T) {
	w := NewWriter()
	fm := NewPDFFontManager(w)
	ef := fm.lookupFont("NONEXISTENT")
	if ef != nil {
		t.Errorf("lookupFont(unknown) should return nil, got %v", ef)
	}
}

func TestLookupFont_KnownAlias(t *testing.T) {
	w := NewWriter()
	fm := NewPDFFontManager(w)
	alias := fm.RegisterFont("Arial", false, false)
	ef := fm.lookupFont(alias)
	if ef == nil {
		t.Errorf("lookupFont(%q) should return non-nil", alias)
	}
}

// ── pdfEscape non-printable characters ───────────────────────────────────────

func TestPdfEscape_NonPrintable(t *testing.T) {
	// Control character \x01 and byte > 0x7E (\x80) are dropped.
	got := pdfEscape("a\x01b\x80c")
	if got != "abc" {
		t.Errorf("pdfEscape with non-printable = %q, want %q", got, "abc")
	}
}

func TestPdfEscape_ParensAndBackslash(t *testing.T) {
	got := pdfEscape(`(hello\world)`)
	want := `\(hello\\world\)`
	if got != want {
		t.Errorf("pdfEscape parens/backslash = %q, want %q", got, want)
	}
}

// ── EncodeText with unknown alias ─────────────────────────────────────────────

func TestEncodeText_UnknownAlias(t *testing.T) {
	w := NewWriter()
	fm := NewPDFFontManager(w)
	// "UNKNOWN" alias does not exist → returns literal PDF string.
	result := fm.EncodeText("UNKNOWN", "hello")
	if !strings.HasPrefix(result, "(") {
		t.Errorf("EncodeText(unknown) = %q, want literal (…) form", result)
	}
}

// ── MeasureText with glyph not in font (CJK) ─────────────────────────────────

func TestMeasureText_GlyphNotInFont(t *testing.T) {
	// A CJK character is unlikely to be in Go's embedded sans font cmap.
	// This triggers the gi==0 branch in MeasureText.
	w := NewWriter()
	fm := NewPDFFontManager(w)
	alias := fm.RegisterFont("Arial", false, false)
	// 中 (U+4E2D) — Chinese character, not expected in GoRegular's Latin cmap.
	width := fm.MeasureText(alias, "中", 10)
	// Should fall back to 10 * 0.6 = 6 estimate.
	if width <= 0 {
		t.Errorf("MeasureText with CJK = %.2f, want > 0", width)
	}
}

// ── pdfDrawBorder with nil Lines entries ─────────────────────────────────────

func TestPdfDrawBorder_NilLineEntry(t *testing.T) {
	// VisibleLines has bits set but Lines array entries are nil.
	// pdfDrawBorder should use NewBorderLine() defaults without panicking.
	w := NewWriter()
	c := NewContents(w)
	border := &style.Border{
		VisibleLines: style.BorderLinesLeft | style.BorderLinesTop,
		// Lines stays nil (zero value of [4]*BorderLine array)
	}
	exp := &Exporter{}
	// Should not panic — nil Lines entries are handled gracefully.
	exp.pdfDrawBorder(c, 0, 0, 100, 50, border)
}

// ── Finish with nil writer ────────────────────────────────────────────────────

func TestFinish_NilWriter(t *testing.T) {
	// Finish on a zero-value Exporter (nil writer) should be a no-op.
	exp := &Exporter{}
	if err := exp.Finish(); err != nil {
		t.Fatalf("Finish with nil writer: %v", err)
	}
}

// ── renderTextObject with nil fontMgr (non-embedded alias path) ──────────────

func TestRenderTextObject_NilFontMgr(t *testing.T) {
	// With nil fontMgr, renderTextObject falls back to pdfFontAlias (non-EF alias).
	// This covers the fontAlias = pdfFontAlias(...) and !isEmbedded text path.
	w := NewWriter()
	c := NewContents(w)
	exp := &Exporter{fontMgr: nil}
	// Set up a fake curPage so renderTextObject can access Height.
	// We need to construct a Page in a way that sets Height.
	pages := NewPages(w)
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 50}
	obj := preview.PreparedObject{
		Name:  "obj",
		Kind:  preview.ObjectTypeText,
		Left:  10, Top: 5, Width: 200, Height: 20,
		Text:  "hello world",
		Font:  style.Font{Name: "Arial", Size: 10},
	}
	exp.renderTextObject(c, b, obj)
	// The content buffer should contain text operators.
	if c.buf.Len() == 0 {
		t.Error("expected non-empty content from renderTextObject with nil fontMgr")
	}
	if !strings.Contains(c.buf.String(), "hello world") {
		t.Error("expected text content in output")
	}
}
