// html_coverage2_test.go — additional tests that raise coverage for:
//   • renderBandBackground (html.go:280)  — was 0%
//   • HasClass, RegisterClass, StyleBlock (css_registry.go) — were 0%
//   • fontLineHeightRatioFloat (html.go:1008) — was 0%
//   • fontLineHeightRatio additional cases (html.go:983) — was 30%
//   • htmlBorderWidthValues (html.go:1157) — was 14.3%

package html

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// ── cssRegistry: HasClass ──────────────────────────────────────────────────────

func TestCSSRegistry_HasClass_NotRegistered(t *testing.T) {
	r := newCSSRegistry()
	if r.HasClass("myclass") {
		t.Error("HasClass: expected false for unregistered key")
	}
}

func TestCSSRegistry_HasClass_AfterRegisterClass(t *testing.T) {
	r := newCSSRegistry()
	r.RegisterClass("myclass", "color:red;")
	if !r.HasClass("myclass") {
		t.Error("HasClass: expected true after RegisterClass")
	}
}

func TestCSSRegistry_HasClass_DoesNotMatchRegularRegister(t *testing.T) {
	// Register via Register() (not RegisterClass) — HasClass should not find it.
	r := newCSSRegistry()
	r.Register("color:blue;")
	if r.HasClass("s0") {
		t.Error("HasClass: should not find auto-named classes registered via Register()")
	}
}

// ── cssRegistry: RegisterClass ────────────────────────────────────────────────

func TestCSSRegistry_RegisterClass_Basic(t *testing.T) {
	r := newCSSRegistry()
	r.RegisterClass("bold", "font-weight:bold;")
	if !r.HasClass("bold") {
		t.Error("RegisterClass: class not found after registration")
	}
	// The index should map the css to the key name.
	name := r.Register("font-weight:bold;")
	if name != "bold" {
		t.Errorf("RegisterClass: expected index to map css → key 'bold', got %q", name)
	}
}

func TestCSSRegistry_RegisterClass_NoOp_IfAlreadyRegistered(t *testing.T) {
	r := newCSSRegistry()
	r.RegisterClass("mykey", "color:red;")
	// Second registration with different css should be ignored.
	r.RegisterClass("mykey", "color:green;")
	// The stored css should still be the first one.
	block := r.StyleBlock()
	if !strings.Contains(block, "color:red;") {
		t.Errorf("RegisterClass no-op: expected first css 'color:red;' still present, got:\n%s", block)
	}
	if strings.Contains(block, "color:green;") {
		t.Errorf("RegisterClass no-op: second css 'color:green;' should not be present, got:\n%s", block)
	}
}

func TestCSSRegistry_RegisterClass_CountIncreases(t *testing.T) {
	r := newCSSRegistry()
	before := r.Count()
	r.RegisterClass("k1", "padding:0;")
	if r.Count() != before+1 {
		t.Errorf("RegisterClass: expected Count to increase by 1, got %d", r.Count())
	}
}

// ── cssRegistry: StyleBlock ────────────────────────────────────────────────────

func TestCSSRegistry_StyleBlock_EmptyWhenNoClasses(t *testing.T) {
	r := newCSSRegistry()
	block := r.StyleBlock()
	if block != "" {
		t.Errorf("StyleBlock: expected empty string when no classes registered, got %q", block)
	}
}

func TestCSSRegistry_StyleBlock_ContainsRegisteredCSS(t *testing.T) {
	r := newCSSRegistry()
	r.Register("color:red;")
	block := r.StyleBlock()
	if !strings.Contains(block, "color:red;") {
		t.Errorf("StyleBlock: expected 'color:red;' in output, got:\n%s", block)
	}
	if !strings.Contains(block, "<style") {
		t.Errorf("StyleBlock: expected <style tag, got:\n%s", block)
	}
	if !strings.Contains(block, "</style>") {
		t.Errorf("StyleBlock: expected </style>, got:\n%s", block)
	}
}

func TestCSSRegistry_StyleBlock_ContainsNamedClass(t *testing.T) {
	r := newCSSRegistry()
	r.RegisterClass("highlight", "background:yellow;")
	block := r.StyleBlock()
	if !strings.Contains(block, ".highlight") {
		t.Errorf("StyleBlock: expected '.highlight' selector, got:\n%s", block)
	}
	if !strings.Contains(block, "background:yellow;") {
		t.Errorf("StyleBlock: expected 'background:yellow;' in output, got:\n%s", block)
	}
}

func TestCSSRegistry_StyleBlock_ContainsParagraphReset(t *testing.T) {
	r := newCSSRegistry()
	r.Register("color:black;")
	block := r.StyleBlock()
	if !strings.Contains(block, "margin-block-start") {
		t.Errorf("StyleBlock: expected paragraph reset rule, got:\n%s", block)
	}
}

func TestCSSRegistry_StyleBlock_MixedRegisterAndRegisterClass(t *testing.T) {
	r := newCSSRegistry()
	r.Register("font-size:12px;")
	r.RegisterClass("named", "font-weight:bold;")
	block := r.StyleBlock()
	if !strings.Contains(block, ".s0") {
		t.Errorf("StyleBlock: expected .s0 auto class, got:\n%s", block)
	}
	if !strings.Contains(block, ".named") {
		t.Errorf("StyleBlock: expected .named class, got:\n%s", block)
	}
}

// ── fontLineHeightRatio ────────────────────────────────────────────────────────

func TestFontLineHeightRatio_AllCases(t *testing.T) {
	tests := []struct {
		font string
		want string
	}{
		{"Arial", "1.15"},
		{"arial narrow", "1.15"},
		{"Times New Roman", "1.15"},
		{"times", "1.15"},
		{"Tahoma", "1.21"},
		{"Microsoft Sans Serif", "1.21"},
		{"Verdana", "1.22"},
		{"Arial Unicode MS", "1.34"},
		{"Arial Black", "1.41"},
		{"Georgia", "1.14"},
		{"Courier New", "1.13"},
		{"Courier", "1.13"},
		{"Segoe UI", "1.33"},
		{"UnknownFont", "1.21"}, // default
		{"", "1.21"},            // default for empty string
	}
	for _, tt := range tests {
		got := fontLineHeightRatio(tt.font)
		if got != tt.want {
			t.Errorf("fontLineHeightRatio(%q) = %q, want %q", tt.font, got, tt.want)
		}
	}
}

// ── fontLineHeightRatioFloat ───────────────────────────────────────────────────

func TestFontLineHeightRatioFloat_AllCases(t *testing.T) {
	tests := []struct {
		font string
		want float64
	}{
		{"Arial", 1.15},
		{"arial narrow", 1.15},
		{"Times New Roman", 1.15},
		{"times", 1.15},
		{"Tahoma", 1.21},
		{"Microsoft Sans Serif", 1.21},
		{"Verdana", 1.22},
		{"Arial Unicode MS", 1.34},
		{"Arial Black", 1.41},
		{"Georgia", 1.14},
		{"Courier New", 1.13},
		{"Courier", 1.13},
		{"Segoe UI", 1.33},
		{"UnknownFont", 1.21}, // default
		{"", 1.21},            // default for empty string
	}
	for _, tt := range tests {
		got := fontLineHeightRatioFloat(tt.font)
		if got != tt.want {
			t.Errorf("fontLineHeightRatioFloat(%q) = %v, want %v", tt.font, got, tt.want)
		}
	}
}

// ── htmlBorderWidthValues ─────────────────────────────────────────────────────

func TestHTMLBorderWidthValues_Nil_NoChange(t *testing.T) {
	var left, top, right, bottom float32
	htmlBorderWidthValues(nil, 1.0, &left, &top, &right, &bottom)
	if left != 0 || top != 0 || right != 0 || bottom != 0 {
		t.Error("nil border: expected no change to output values")
	}
}

func TestHTMLBorderWidthValues_None_NoChange(t *testing.T) {
	b := &style.Border{VisibleLines: style.BorderLinesNone}
	var left, top, right, bottom float32
	htmlBorderWidthValues(b, 1.0, &left, &top, &right, &bottom)
	if left != 0 || top != 0 || right != 0 || bottom != 0 {
		t.Error("BorderLinesNone: expected no change to output values")
	}
}

func TestHTMLBorderWidthValues_AllSides_NilLines_DefaultWidth(t *testing.T) {
	b := &style.Border{VisibleLines: style.BorderLinesAll}
	var left, top, right, bottom float32
	htmlBorderWidthValues(b, 1.0, &left, &top, &right, &bottom)
	// nil line entry → default width 1 * scale=1 = 1
	if left != 1 {
		t.Errorf("AllSides nil lines: expected left=1, got %v", left)
	}
	if top != 1 {
		t.Errorf("AllSides nil lines: expected top=1, got %v", top)
	}
	if right != 1 {
		t.Errorf("AllSides nil lines: expected right=1, got %v", right)
	}
	if bottom != 1 {
		t.Errorf("AllSides nil lines: expected bottom=1, got %v", bottom)
	}
}

func TestHTMLBorderWidthValues_AllSides_WithScale(t *testing.T) {
	bl := &style.BorderLine{Width: 2, Style: style.LineStyleSolid}
	b := &style.Border{
		VisibleLines: style.BorderLinesAll,
		Lines:        [4]*style.BorderLine{bl, bl, bl, bl},
	}
	var left, top, right, bottom float32
	htmlBorderWidthValues(b, 3.0, &left, &top, &right, &bottom)
	// width=2 * scale=3 = 6
	if left != 6 {
		t.Errorf("scale: expected left=6, got %v", left)
	}
	if top != 6 {
		t.Errorf("scale: expected top=6, got %v", top)
	}
	if right != 6 {
		t.Errorf("scale: expected right=6, got %v", right)
	}
	if bottom != 6 {
		t.Errorf("scale: expected bottom=6, got %v", bottom)
	}
}

func TestHTMLBorderWidthValues_TopOnly(t *testing.T) {
	bl := &style.BorderLine{Width: 1, Style: style.LineStyleSolid}
	b := &style.Border{
		VisibleLines: style.BorderLinesTop,
		Lines:        [4]*style.BorderLine{nil, bl, nil, nil},
	}
	var left, top, right, bottom float32
	htmlBorderWidthValues(b, 1.0, &left, &top, &right, &bottom)
	if top != 1 {
		t.Errorf("TopOnly: expected top=1, got %v", top)
	}
	if left != 0 || right != 0 || bottom != 0 {
		t.Errorf("TopOnly: expected left/right/bottom=0, got l=%v r=%v b=%v", left, right, bottom)
	}
}

func TestHTMLBorderWidthValues_DoubleLineStyle_TripleWidth(t *testing.T) {
	bl := &style.BorderLine{Width: 2, Style: style.LineStyleDouble}
	b := &style.Border{
		VisibleLines: style.BorderLinesLeft,
		Lines:        [4]*style.BorderLine{bl, nil, nil, nil},
	}
	var left, top, right, bottom float32
	htmlBorderWidthValues(b, 1.0, &left, &top, &right, &bottom)
	// Double style: width * 3 * scale = 2 * 3 * 1 = 6
	if left != 6 {
		t.Errorf("DoubleLineStyle: expected left=6, got %v", left)
	}
}

func TestHTMLBorderWidthValues_DoubleLineStyle_WithScale(t *testing.T) {
	bl := &style.BorderLine{Width: 1, Style: style.LineStyleDouble}
	b := &style.Border{
		VisibleLines: style.BorderLinesRight,
		Lines:        [4]*style.BorderLine{nil, nil, bl, nil},
	}
	var left, top, right, bottom float32
	htmlBorderWidthValues(b, 2.0, &left, &top, &right, &bottom)
	// Double style: 1 * 3 * 2 = 6
	if right != 6 {
		t.Errorf("DoubleLineStyle+scale: expected right=6, got %v", right)
	}
}

func TestHTMLBorderWidthValues_LeftRightOnly(t *testing.T) {
	blLeft := &style.BorderLine{Width: 3, Style: style.LineStyleSolid}
	blRight := &style.BorderLine{Width: 5, Style: style.LineStyleSolid}
	b := &style.Border{
		VisibleLines: style.BorderLinesLeft | style.BorderLinesRight,
		Lines:        [4]*style.BorderLine{blLeft, nil, blRight, nil},
	}
	var left, top, right, bottom float32
	htmlBorderWidthValues(b, 1.0, &left, &top, &right, &bottom)
	if left != 3 {
		t.Errorf("LeftRight: expected left=3, got %v", left)
	}
	if right != 5 {
		t.Errorf("LeftRight: expected right=5, got %v", right)
	}
	if top != 0 || bottom != 0 {
		t.Errorf("LeftRight: expected top/bottom=0, got t=%v b=%v", top, bottom)
	}
}

func TestHTMLBorderWidthValues_BottomOnly_NilLine(t *testing.T) {
	// nil line entry → default width 1
	b := &style.Border{
		VisibleLines: style.BorderLinesBottom,
		// Lines[3] (Bottom) = nil → default
	}
	var left, top, right, bottom float32
	htmlBorderWidthValues(b, 2.0, &left, &top, &right, &bottom)
	// default 1 * scale 2 = 2
	if bottom != 2 {
		t.Errorf("BottomOnly nil line: expected bottom=2, got %v", bottom)
	}
}

// ── renderBandBackground (via Export) ─────────────────────────────────────────

// bandWithFill creates a PreparedPages with a single band that has a fill color.
func bandWithFill(fillColor color.RGBA) *preview.PreparedPages {
	pp := preview.New()
	pp.AddPage(594, 841, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:      "TestBand",
		Left:      0,
		Top:       0,
		Width:     594,
		Height:    40,
		FillColor: fillColor,
	})
	return pp
}

// exportBandBackground is a helper that runs Export and returns the HTML string.
func exportBandBackground(t *testing.T, pp *preview.PreparedPages) string {
	t.Helper()
	exp := NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("exportBandBackground Export: %v", err)
	}
	return buf.String()
}

func TestRenderBandBackground_OpaqueColor(t *testing.T) {
	// FillColor with A=255 → background-color:rgb(...)
	pp := bandWithFill(color.RGBA{R: 200, G: 100, B: 50, A: 255})
	out := exportBandBackground(t, pp)
	if !strings.Contains(out, "rgb(200, 100, 50)") {
		t.Errorf("OpaqueColor: expected 'rgb(200, 100, 50)', got:\n%s", out)
	}
}

func TestRenderBandBackground_TransparentColor(t *testing.T) {
	// FillColor with A=0 → renderBandBackground emits background-color:transparent
	// (not an opaque rgb(...) color).
	pp := bandWithFill(color.RGBA{R: 0, G: 0, B: 0, A: 0})
	out := exportBandBackground(t, pp)
	// The band should use "transparent", not an rgb(0,0,0) color.
	if strings.Contains(out, "background-color:rgb(0, 0, 0)") {
		t.Errorf("Transparent: expected no black background-color, got:\n%s", out)
	}
}
