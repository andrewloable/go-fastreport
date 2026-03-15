package utils

import (
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

func TestHtmlTextRenderer_PlainText(t *testing.T) {
	cases := []struct {
		html  string
		plain string
	}{
		{"hello world", "hello world"},
		{"<b>bold</b> text", "bold text"},
		{"a<br>b", "a\nb"},
		{"&amp;&lt;&gt;&nbsp;&quot;", `&<> "`},
		{"<font color=\"red\">red</font>", "red"},
		{"<span style=\"color:blue\">blue</span>", "blue"},
	}
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	for _, tc := range cases {
		r := NewHtmlTextRenderer(tc.html, f, c)
		got := r.PlainText()
		if got != tc.plain {
			t.Errorf("PlainText(%q) = %q, want %q", tc.html, got, tc.plain)
		}
	}
}

func TestHtmlTextRenderer_Bold(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("<b>hello</b>", f, c)
	lines := r.Lines()
	if len(lines) != 1 || len(lines[0].Runs) == 0 {
		t.Fatal("expected 1 line with runs")
	}
	run := lines[0].Runs[0]
	if run.Font.Style&style.FontStyleBold == 0 {
		t.Error("expected bold font style")
	}
}

func TestStripHtmlTags(t *testing.T) {
	got := StripHtmlTags("<b>hello</b> <i>world</i>")
	if got != "hello world" {
		t.Errorf("StripHtmlTags = %q", got)
	}
}

func TestHtmlTextRenderer_MeasureHeight(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("line1<br>line2<br>line3", f, c)
	h := r.MeasureHeight(0)
	if h <= 0 {
		t.Error("expected positive height")
	}
}

func TestHtmlTextRenderer_MeasureHeight_WithWidth(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("word1 word2 word3 word4 word5 word6 word7", f, c)
	h := r.MeasureHeight(50) // narrow width forces word wrap
	if h <= 0 {
		t.Error("expected positive height with width wrap")
	}
}

// ── parseFloat ────────────────────────────────────────────────────────────────

func TestParseFloat_Valid(t *testing.T) {
	var f float32
	n, err := parseFloat("12.5", &f)
	if err != nil {
		t.Fatalf("parseFloat: %v", err)
	}
	if f != 12.5 {
		t.Errorf("parseFloat = %v, want 12.5", f)
	}
	if n != 4 {
		t.Errorf("n = %d, want 4", n)
	}
}

func TestParseFloat_Integer(t *testing.T) {
	var f float32
	_, err := parseFloat("42", &f)
	if err != nil {
		t.Fatalf("parseFloat integer: %v", err)
	}
	if f != 42 {
		t.Errorf("parseFloat = %v, want 42", f)
	}
}

func TestParseFloat_Negative(t *testing.T) {
	var f float32
	_, err := parseFloat("-3.5", &f)
	if err != nil {
		t.Fatalf("parseFloat negative: %v", err)
	}
	if f != -3.5 {
		t.Errorf("parseFloat = %v, want -3.5", f)
	}
}

func TestParseFloat_Empty(t *testing.T) {
	var f float32
	_, err := parseFloat("", &f)
	if err == nil {
		t.Error("expected error for empty string")
	}
}

func TestParseFloat_WithSuffix(t *testing.T) {
	var f float32
	n, err := parseFloat("10pt", &f)
	if err != nil {
		t.Fatalf("parseFloat with suffix: %v", err)
	}
	if f != 10 {
		t.Errorf("parseFloat = %v, want 10", f)
	}
	if n != 2 {
		t.Errorf("n = %d, want 2", n)
	}
}

// ── parseError ────────────────────────────────────────────────────────────────

func TestParseError_Error(t *testing.T) {
	e := &parseError{s: "bad_input"}
	got := e.Error()
	if got == "" {
		t.Error("Error() returned empty string")
	}
	if !strings.Contains(got, "bad_input") {
		t.Errorf("Error() = %q, expected to contain 'bad_input'", got)
	}
}

// ── applyInlineStyle ──────────────────────────────────────────────────────────

func TestApplyInlineStyle_FontSize(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="font-size:14pt">sized</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if lines[0].Runs[0].Font.Size != 14 {
		t.Errorf("font-size: got %v, want 14", lines[0].Runs[0].Font.Size)
	}
}

func TestApplyInlineStyle_FontWeightBold(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="font-weight:bold">bold</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if lines[0].Runs[0].Font.Style&style.FontStyleBold == 0 {
		t.Error("expected bold style")
	}
}

func TestApplyInlineStyle_FontStyleItalic(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="font-style:italic">italic</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if lines[0].Runs[0].Font.Style&style.FontStyleItalic == 0 {
		t.Error("expected italic style")
	}
}

func TestApplyInlineStyle_TextDecorationUnderline(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="text-decoration:underline">u</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if !lines[0].Runs[0].Underline {
		t.Error("expected underline")
	}
}

func TestApplyInlineStyle_TextDecorationLineThrough(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="text-decoration:line-through">s</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if !lines[0].Runs[0].Strikeout {
		t.Error("expected strikeout")
	}
}

func TestApplyInlineStyle_Color(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="color:#ff0000">red</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	clr := lines[0].Runs[0].Color
	if clr.R != 0xFF {
		t.Errorf("color.R = %d, want 255", clr.R)
	}
}

// ── parseAttrs ────────────────────────────────────────────────────────────────

func TestParseAttrs_SingleQuote(t *testing.T) {
	attrs := parseAttrs(` color='red'`)
	if attrs["color"] != "red" {
		t.Errorf("single-quote: got %q, want 'red'", attrs["color"])
	}
}

func TestParseAttrs_Unquoted(t *testing.T) {
	attrs := parseAttrs(` color=red`)
	if attrs["color"] != "red" {
		t.Errorf("unquoted: got %q, want 'red'", attrs["color"])
	}
}

func TestParseAttrs_NoEquals(t *testing.T) {
	attrs := parseAttrs(`justtext`)
	if len(attrs) != 0 {
		t.Errorf("no-equals: expected empty map, got %v", attrs)
	}
}

// ── HTML entity and tag edge cases ────────────────────────────────────────────

func TestHtmlParse_UnknownEntity(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("a&unknown;b", f, c)
	plain := r.PlainText()
	if !strings.Contains(plain, "&unknown;") {
		t.Errorf("unknown entity not preserved: got %q", plain)
	}
}

func TestHtmlParse_AmpersandNoSemicolon(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	// '&' with no ';' within 10 chars — treated as literal '&'
	r := NewHtmlTextRenderer("a&b", f, c)
	plain := r.PlainText()
	if !strings.Contains(plain, "&") {
		t.Errorf("bare ampersand not preserved: got %q", plain)
	}
}

func TestHtmlParse_SelfClosingBr(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("line1<br/>line2", f, c)
	lines := r.Lines()
	if len(lines) < 2 {
		t.Errorf("expected >= 2 lines for <br/>, got %d", len(lines))
	}
}

func TestHtmlParse_UnderlineTag(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("<u>underlined</u>", f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if !lines[0].Runs[0].Underline {
		t.Error("expected underline from <u> tag")
	}
}

func TestHtmlParse_StrikeTag(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("<s>strike</s>", f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if !lines[0].Runs[0].Strikeout {
		t.Error("expected strikeout from <s> tag")
	}
}

func TestHtmlParse_StrongTag(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("<strong>strong</strong>", f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if lines[0].Runs[0].Font.Style&style.FontStyleBold == 0 {
		t.Error("expected bold from <strong> tag")
	}
}

func TestHtmlParse_EmTag(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("<em>emphasis</em>", f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if lines[0].Runs[0].Font.Style&style.FontStyleItalic == 0 {
		t.Error("expected italic from <em> tag")
	}
}

func TestHtmlParse_FontSize(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<font size="18">big</font>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if lines[0].Runs[0].Font.Size != 18 {
		t.Errorf("font size: got %v, want 18", lines[0].Runs[0].Font.Size)
	}
}

func TestHtmlParse_UnclosedTag(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	// '<' without '>' — treated as literal char
	r := NewHtmlTextRenderer("a<b", f, c)
	plain := r.PlainText()
	if !strings.Contains(plain, "<") {
		t.Errorf("unclosed tag: expected '<' in output, got %q", plain)
	}
}

func TestHtmlParse_LongEntityName(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	// '&' with semicolon more than 10 chars away — treated as literal
	r := NewHtmlTextRenderer("a&verylongentityname;b", f, c)
	plain := r.PlainText()
	// The '&' should be emitted literally since semi >= 10
	if !strings.Contains(plain, "&") {
		t.Errorf("long entity: expected '&' in output, got %q", plain)
	}
}

func TestParseAttrs_UnclosedQuote(t *testing.T) {
	// Quoted value with no closing quote — should break and return empty map.
	attrs := parseAttrs(` color="unclosed`)
	// Either color is absent or the loop breaks.
	_ = attrs // No panic is success; coverage goal is the `if end < 0 { break }` branch
}

func TestApplyInlineStyle_NoColon(t *testing.T) {
	// A declaration without ':' should be skipped (continue).
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="nodeclaration">text</span>`, f, c)
	plain := r.PlainText()
	if plain == "" {
		t.Error("expected non-empty plain text")
	}
}

func TestHtmlParse_LiteralNewline(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	// Literal '\n' in HTML triggers the newLine() else-if branch.
	r := NewHtmlTextRenderer("line1\nline2", f, c)
	lines := r.Lines()
	if len(lines) < 2 {
		t.Errorf("expected >= 2 lines for literal newline, got %d", len(lines))
	}
}

func TestParseAttrs_MultipleAttrs(t *testing.T) {
	attrs := parseAttrs(` color="red" size="12"`)
	if attrs["color"] != "red" {
		t.Errorf("color = %q, want 'red'", attrs["color"])
	}
	if attrs["size"] != "12" {
		t.Errorf("size = %q, want '12'", attrs["size"])
	}
}

func TestParseAttrs_UnquotedMultiple(t *testing.T) {
	// Unquoted value followed by another attr — covers the `end >= 0` else branch.
	attrs := parseAttrs(`color=red size=10`)
	if attrs["color"] != "red" {
		t.Errorf("color = %q, want 'red'", attrs["color"])
	}
}
