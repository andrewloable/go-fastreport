package utils

// htmltext_renderer_test.go covers the advanced layout features ported from
// FastReport's HtmlTextRenderer.cs: paragraph model, subscript/superscript,
// font-family, background-color, rgb()/rgba() color parsing, and CSS font-size
// units (px, em).

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

// ── BaselineType constants ────────────────────────────────────────────────────

func TestBaselineType_Constants(t *testing.T) {
	if BaselineNormal != 0 {
		t.Errorf("BaselineNormal = %d, want 0", BaselineNormal)
	}
	if BaselineSubscript != 1 {
		t.Errorf("BaselineSubscript = %d, want 1", BaselineSubscript)
	}
	if BaselineSuperscript != 2 {
		t.Errorf("BaselineSuperscript = %d, want 2", BaselineSuperscript)
	}
}

// ── Subscript <sub> tag ───────────────────────────────────────────────────────

func TestHtmlParse_SubTag_Baseline(t *testing.T) {
	// <sub> must set Baseline = BaselineSubscript on the enclosed run.
	// Mirrors HtmlTextRenderer.cs case "sub" (line 1012).
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("H<sub>2</sub>O", f, c)
	lines := r.Lines()
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	// Find the subscript run (text "2").
	var found bool
	for _, run := range lines[0].Runs {
		if run.Text == "2" {
			found = true
			if run.Baseline != BaselineSubscript {
				t.Errorf("run Baseline = %v, want BaselineSubscript", run.Baseline)
			}
		}
	}
	if !found {
		t.Error("subscript run 'text=2' not found")
	}
}

func TestHtmlParse_SubTag_NormalRun_Unchanged(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("H<sub>2</sub>O", f, c)
	lines := r.Lines()
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	for _, run := range lines[0].Runs {
		if run.Text == "H" || run.Text == "O" {
			if run.Baseline != BaselineNormal {
				t.Errorf("run %q: Baseline = %v, want BaselineNormal", run.Text, run.Baseline)
			}
		}
	}
}

// ── Superscript <sup> tag ─────────────────────────────────────────────────────

func TestHtmlParse_SupTag_Baseline(t *testing.T) {
	// <sup> must set Baseline = BaselineSuperscript on the enclosed run.
	// Mirrors HtmlTextRenderer.cs case "sup" (line 1016).
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("x<sup>2</sup>", f, c)
	lines := r.Lines()
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	var found bool
	for _, run := range lines[0].Runs {
		if run.Text == "2" {
			found = true
			if run.Baseline != BaselineSuperscript {
				t.Errorf("run Baseline = %v, want BaselineSuperscript", run.Baseline)
			}
		}
	}
	if !found {
		t.Error("superscript run 'text=2' not found")
	}
}

func TestHtmlParse_SupTag_PlainText(t *testing.T) {
	// PlainText must include the superscripted text content.
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("x<sup>2</sup>", f, c)
	plain := r.PlainText()
	if plain != "x2" {
		t.Errorf("PlainText = %q, want %q", plain, "x2")
	}
}

// ── Sub/Sup baseline reset after closing tag ──────────────────────────────────

func TestHtmlParse_SubTag_BaselineResetAfterClose(t *testing.T) {
	// Text after </sub> should revert to BaselineNormal.
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("a<sub>b</sub>c", f, c)
	lines := r.Lines()
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	for _, run := range lines[0].Runs {
		if run.Text == "c" {
			if run.Baseline != BaselineNormal {
				t.Errorf("after </sub>: run 'c' Baseline = %v, want BaselineNormal", run.Baseline)
			}
		}
	}
}

// ── background-color CSS property ─────────────────────────────────────────────

func TestApplyInlineStyle_BackgroundColor_Hex(t *testing.T) {
	// CSS background-color:#rrggbb should set BackgroundColor on the run.
	// Mirrors HtmlTextRenderer.cs CssStyle() "background-color" (line 670).
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="background-color:#00ff00">green bg</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	bg := lines[0].Runs[0].BackgroundColor
	if bg.G != 0xFF || bg.R != 0 || bg.B != 0 {
		t.Errorf("BackgroundColor = %v, want G=255 R=0 B=0", bg)
	}
	if bg.A == 0 {
		t.Error("BackgroundColor.A should be non-zero for visible background")
	}
}

func TestApplyInlineStyle_BackgroundColor_Named(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="background-color:yellow">text</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	bg := lines[0].Runs[0].BackgroundColor
	if bg.A == 0 {
		t.Error("BackgroundColor.A should be non-zero for named color")
	}
}

func TestHtmlRun_NoBackgroundByDefault(t *testing.T) {
	// Runs without a background-color declaration should have A=0.
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("plain text", f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	bg := lines[0].Runs[0].BackgroundColor
	if bg.A != 0 {
		t.Errorf("expected no background (A=0), got %v", bg)
	}
}

// ── font-family CSS property ──────────────────────────────────────────────────

func TestApplyInlineStyle_FontFamily(t *testing.T) {
	// CSS font-family should update the run's font name.
	// Mirrors HtmlTextRenderer.cs CssStyle() "font-family" (line 618).
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="font-family:Arial">text</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if lines[0].Runs[0].Font.Name != "Arial" {
		t.Errorf("Font.Name = %q, want 'Arial'", lines[0].Runs[0].Font.Name)
	}
}

func TestApplyInlineStyle_FontFamily_SingleQuotes(t *testing.T) {
	// font-family value may include single quotes (e.g. font-family:'Times New Roman').
	// The C# code strips them (line 621).
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="font-family:'Times New Roman'">text</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	name := lines[0].Runs[0].Font.Name
	if name != "Times New Roman" {
		t.Errorf("Font.Name = %q, want 'Times New Roman'", name)
	}
}

// ── <font face="..."> attribute ───────────────────────────────────────────────

func TestHtmlParse_FontFaceAttr(t *testing.T) {
	// <font face="Courier New"> should set the font name.
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<font face="Courier New">text</font>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if lines[0].Runs[0].Font.Name != "Courier New" {
		t.Errorf("Font.Name = %q, want 'Courier New'", lines[0].Runs[0].Font.Name)
	}
}

// ── CSS font-size units: px ───────────────────────────────────────────────────

func TestApplyInlineStyle_FontSizePx(t *testing.T) {
	// font-size in px: 16px → 12pt (16 * 0.75).
	// Mirrors HtmlTextRenderer.cs CssStyle() "font-size" px branch (line 612).
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="font-size:16px">text</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	want := float32(16 * 0.75)
	if lines[0].Runs[0].Font.Size != want {
		t.Errorf("font-size 16px: Font.Size = %v, want %v", lines[0].Runs[0].Font.Size, want)
	}
}

// ── CSS font-size units: em ───────────────────────────────────────────────────

func TestApplyInlineStyle_FontSizeEm(t *testing.T) {
	// font-size:2em should double the current font size.
	// Mirrors HtmlTextRenderer.cs CssStyle() "font-size" em branch (line 616).
	base := style.DefaultFont()
	base.Size = 10
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="font-size:2em">text</span>`, base, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	want := float32(20)
	if lines[0].Runs[0].Font.Size != want {
		t.Errorf("font-size 2em: Font.Size = %v, want %v", lines[0].Runs[0].Font.Size, want)
	}
}

// ── parseCSSColor — rgb() / rgba() ────────────────────────────────────────────

func TestParseCSSColor_Rgb(t *testing.T) {
	// rgb(255, 0, 0) → red
	clr, ok := parseCSSColor("rgb(255, 0, 0)")
	if !ok {
		t.Fatal("parseCSSColor rgb() returned ok=false")
	}
	if clr.R != 255 || clr.G != 0 || clr.B != 0 || clr.A != 255 {
		t.Errorf("rgb(255,0,0) = %v, want R=255 G=0 B=0 A=255", clr)
	}
}

func TestParseCSSColor_Rgba(t *testing.T) {
	// rgba(0, 128, 255, 0.5) → A ≈ 127 (0.5 * 255)
	clr, ok := parseCSSColor("rgba(0, 128, 255, 0.5)")
	if !ok {
		t.Fatal("parseCSSColor rgba() returned ok=false")
	}
	if clr.R != 0 || clr.G != 128 || clr.B != 255 {
		t.Errorf("rgba rgb values wrong: %v", clr)
	}
	if clr.A < 120 || clr.A > 135 {
		t.Errorf("rgba alpha = %d, want ~127 (0.5*255)", clr.A)
	}
}

func TestParseCSSColor_Hex(t *testing.T) {
	clr, ok := parseCSSColor("#ff0000")
	if !ok {
		t.Fatal("parseCSSColor hex returned ok=false")
	}
	if clr.R != 255 || clr.G != 0 || clr.B != 0 {
		t.Errorf("hex #ff0000 = %v, want R=255", clr)
	}
}

func TestParseCSSColor_Named(t *testing.T) {
	clr, ok := parseCSSColor("blue")
	if !ok {
		t.Fatal("parseCSSColor named returned ok=false")
	}
	if clr.B != 255 || clr.R != 0 || clr.G != 0 {
		t.Errorf("named blue = %v, want B=255", clr)
	}
}

func TestParseCSSColor_Unknown(t *testing.T) {
	_, ok := parseCSSColor("notacolor_xyz")
	if ok {
		t.Error("parseCSSColor unknown should return ok=false")
	}
}

func TestParseCSSColor_Empty(t *testing.T) {
	_, ok := parseCSSColor("")
	if ok {
		t.Error("parseCSSColor empty should return ok=false")
	}
}

func TestParseCSSColor_RgbaFullOpaque(t *testing.T) {
	// rgba(255, 255, 255, 1) → fully opaque white
	clr, ok := parseCSSColor("rgba(255, 255, 255, 1)")
	if !ok {
		t.Fatal("parseCSSColor rgba fully opaque returned ok=false")
	}
	if clr.A != 255 {
		t.Errorf("rgba(...,1) A = %d, want 255", clr.A)
	}
}

// ── CSS color in color property via parseCSSColor ─────────────────────────────

func TestApplyInlineStyle_ColorRgb(t *testing.T) {
	// CSS color:rgb(0, 0, 255) should set the run color to blue.
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="color:rgb(0, 0, 255)">text</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	clr := lines[0].Runs[0].Color
	if clr.B != 255 || clr.R != 0 || clr.G != 0 {
		t.Errorf("color rgb(0,0,255) = %v, want B=255", clr)
	}
}

func TestApplyInlineStyle_ColorRgba(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer(`<span style="color:rgba(255, 0, 0, 1)">text</span>`, f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	clr := lines[0].Runs[0].Color
	if clr.R != 255 || clr.G != 0 || clr.B != 0 {
		t.Errorf("color rgba(255,0,0,1) = %v, want R=255", clr)
	}
}

// ── HtmlRun.Baseline zero value ───────────────────────────────────────────────

func TestHtmlRun_DefaultBaseline(t *testing.T) {
	// Runs without sub/sup should have Baseline == BaselineNormal.
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("plain", f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	if lines[0].Runs[0].Baseline != BaselineNormal {
		t.Errorf("default Baseline = %v, want BaselineNormal", lines[0].Runs[0].Baseline)
	}
}

// ── Nested sub/sup with other styles ─────────────────────────────────────────

func TestHtmlParse_BoldSub(t *testing.T) {
	// <b><sub>...</sub></b> — the run should be both bold and subscript.
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("<b><sub>x</sub></b>", f, c)
	lines := r.Lines()
	if len(lines) == 0 || len(lines[0].Runs) == 0 {
		t.Fatal("no runs")
	}
	run := lines[0].Runs[0]
	if run.Font.Style&style.FontStyleBold == 0 {
		t.Error("expected bold")
	}
	if run.Baseline != BaselineSubscript {
		t.Errorf("Baseline = %v, want BaselineSubscript", run.Baseline)
	}
}

// ── Multiple paragraphs (newline creates new paragraph) ───────────────────────

func TestHtmlParse_MultilineParagraphs(t *testing.T) {
	// Literal '\n' in the source creates separate lines.
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("para1\npara2\npara3", f, c)
	lines := r.Lines()
	if len(lines) < 3 {
		t.Errorf("expected >= 3 lines, got %d", len(lines))
	}
}

// ── PlainText preserves sub/sup text content ──────────────────────────────────

func TestHtmlParse_SubPlainText(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("H<sub>2</sub>O", f, c)
	got := r.PlainText()
	if got != "H2O" {
		t.Errorf("PlainText = %q, want %q", got, "H2O")
	}
}
