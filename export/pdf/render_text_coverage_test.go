package pdf

// render_text_coverage_test.go — extra tests for renderTextObject branches
// not covered by existing tests. These are internal tests (package pdf) so
// they have access to unexported functions and types.

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// TestRenderTextObject_InnerWZero covers the `if innerW <= 0` branch (lines
// 233-235) where the object width is so small that inner padding makes the
// usable width non-positive.  This is only reachable via direct internal call;
// the public API clamps Width to >= 1 px before calling renderTextObject.
func TestRenderTextObject_InnerWZero(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{fontMgr: nil}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 50}
	// Width = 2 px → wPt ≈ 1.5pt.  padPt = PixelsToPoints(2) ≈ 1.5pt.
	// innerW = wPt - 2*padPt ≈ 1.5 - 3.0 = -1.5 ≤ 0 → branch fires.
	obj := preview.PreparedObject{
		Name:   "narrow",
		Kind:   preview.ObjectTypeText,
		Left:   10, Top: 5, Width: 2, Height: 20,
		Text:   "hello narrow",
		Font:   style.Font{Name: "Arial", Size: 10},
	}
	exp.renderTextObject(c, b, obj)
	out := c.buf.String()
	if !strings.Contains(out, "BT") {
		t.Error("expected BT operator in output for narrow-width text object")
	}
}

// TestRenderTextObject_JustifyWordSpacing covers the HorzAlign=3 (Justify)
// branch with WordWrap=true and multiple words, triggering the
// `if ws > 0 { wordSpacing = ws }` path (line 294-296) and the
// `if wordSpacing != 0 { WriteString Tw / 0 Tw }` paths (lines 305-307, 314-316).
func TestRenderTextObject_JustifyWordSpacing(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{fontMgr: nil}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 100}
	// Wide object (200px) with two-line text, HorzAlign=3 (Justify), WordWrap=true.
	// The first line has multiple words and is not the last line → word spacing fires.
	obj := preview.PreparedObject{
		Name:      "justify",
		Kind:      preview.ObjectTypeText,
		Left:      10, Top: 5, Width: 200, Height: 100,
		Text:      "hello world this is justified text\nsecond line here",
		Font:      style.Font{Name: "Arial", Size: 10},
		HorzAlign: 3,   // Justify
		WordWrap:  true,
	}
	exp.renderTextObject(c, b, obj)
	out := c.buf.String()
	if !strings.Contains(out, "BT") {
		t.Error("expected BT in output")
	}
	// The Tw operator should appear for the non-last justified line.
	if !strings.Contains(out, " Tw") {
		t.Error("expected Tw word-spacing operator in justified output")
	}
}

// TestRenderTextObject_JustifyWordSpacing_NegativeWs covers the case where
// ws <= 0 (text is wider than innerW, so no positive word spacing is applied).
// This exercises the `if ws > 0` false branch.
func TestRenderTextObject_JustifyWordSpacing_NegativeWs(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{fontMgr: nil}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 100}
	// Narrow object (5px) so innerW is tiny; text is wider than innerW → ws <= 0.
	obj := preview.PreparedObject{
		Name:      "justify-neg",
		Kind:      preview.ObjectTypeText,
		Left:      10, Top: 5, Width: 5, Height: 100,
		Text:      "hello world\nsecond line",
		Font:      style.Font{Name: "Arial", Size: 10},
		HorzAlign: 3,   // Justify
		WordWrap:  true,
	}
	exp.renderTextObject(c, b, obj)
	out := c.buf.String()
	if !strings.Contains(out, "BT") {
		t.Error("expected BT in output for narrow justify text object")
	}
}

// TestRenderTextObject_LineXClamp covers the `if lineX < xPt { lineX = xPt }`
// clamp at lines 300-302. This fires when the computed lineX (from Center or
// Right alignment) would place the text left of the object's x coordinate —
// e.g. when the text is wider than the box in Right alignment.
func TestRenderTextObject_LineXClamp(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{fontMgr: nil}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 100}
	// Very narrow box (3px wide) with Center alignment; lineW >> wPt so
	// lineX = xPt + (wPt - lineW) / 2 is very negative → clamp to xPt fires.
	obj := preview.PreparedObject{
		Name:      "linex-clamp",
		Kind:      preview.ObjectTypeText,
		Left:      10, Top: 5, Width: 3, Height: 100,
		Text:      "this text is much wider than the box",
		Font:      style.Font{Name: "Arial", Size: 12},
		HorzAlign: 1, // Center
	}
	exp.renderTextObject(c, b, obj)
	out := c.buf.String()
	if !strings.Contains(out, "BT") {
		t.Error("expected BT in output for lineX clamp test")
	}
}

// TestRenderTextObject_BottomAlign covers the VertAlign=2 (Bottom) branch
// (line 255-256) which was not exercised by existing tests.
func TestRenderTextObject_BottomAlign(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{fontMgr: nil}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 100}
	obj := preview.PreparedObject{
		Name:      "bottom-align",
		Kind:      preview.ObjectTypeText,
		Left:      10, Top: 5, Width: 200, Height: 100,
		Text:      "bottom aligned text",
		Font:      style.Font{Name: "Arial", Size: 10},
		VertAlign: 2, // Bottom
	}
	exp.renderTextObject(c, b, obj)
	out := c.buf.String()
	if !strings.Contains(out, "BT") {
		t.Error("expected BT in output for bottom-aligned text")
	}
}

// TestRenderTextObject_RightAlign covers the HorzAlign=2 (Right) branch.
func TestRenderTextObject_RightAlign(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{fontMgr: nil}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 100}
	obj := preview.PreparedObject{
		Name:      "right-align",
		Kind:      preview.ObjectTypeText,
		Left:      10, Top: 5, Width: 200, Height: 50,
		Text:      "right aligned",
		Font:      style.Font{Name: "Arial", Size: 10},
		HorzAlign: 2, // Right
	}
	exp.renderTextObject(c, b, obj)
	out := c.buf.String()
	if !strings.Contains(out, "BT") {
		t.Error("expected BT in output for right-aligned text")
	}
}

// TestRenderTextObject_EmbeddedFont_WithFontMgr covers the
// `if isEmbedded && e.fontMgr != nil { fm.EncodeText(...) }` branch (line 308-310)
// by supplying a real fontMgr so that fontAlias starts with "EF".
func TestRenderTextObject_EmbeddedFont_WithFontMgr(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	fm := NewPDFFontManager(w)
	exp := &Exporter{fontMgr: fm}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 100}
	obj := preview.PreparedObject{
		Name:      "embedded-font",
		Kind:      preview.ObjectTypeText,
		Left:      10, Top: 5, Width: 200, Height: 50,
		Text:      "embedded font text",
		Font:      style.Font{Name: "Arial", Size: 10},
		HorzAlign: 0, // Left (default)
	}
	exp.renderTextObject(c, b, obj)
	out := c.buf.String()
	if !strings.Contains(out, "BT") {
		t.Error("expected BT in output for embedded-font text object")
	}
	// The encoded text should be a hex string <...> not a literal (...)
	if !strings.Contains(out, "<") {
		t.Error("expected hex-encoded text for embedded font")
	}
}

// TestRenderTextObject_LinesLimitedByBox covers the `break` at line 276 where
// lineY drops below yPt-lineHeight, stopping the rendering loop.
// A very small box with many lines ensures the loop stops early.
func TestRenderTextObject_LinesLimitedByBox(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{fontMgr: nil}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 10}
	// 10 lines of text in a box only 10px tall (≈7.5pt).
	// Font=20pt, lineHeight=24pt. After 1-2 lines, lineY < yPt-lineHeight → break.
	obj := preview.PreparedObject{
		Name:      "many-lines",
		Kind:      preview.ObjectTypeText,
		Left:      10, Top: 0, Width: 200, Height: 10,
		Text:      "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10",
		Font:      style.Font{Name: "Arial", Size: 20},
		VertAlign: 0, // Top — so startY = yPt + hPt - lineHeight, and some lines fall below
	}
	exp.renderTextObject(c, b, obj)
	out := c.buf.String()
	// The break fires, but at least "BT" should be present since some line(s) render.
	if !strings.Contains(out, "ET") {
		t.Error("expected ET operator in output (some lines should render before break)")
	}
}

// TestMeasureText_GlyphAdvanceError attempts to trigger the GlyphAdvance error
// path (line 209-211 in font.go). In practice GlyphAdvance on a well-formed
// sfnt font rarely fails; we use a very large ppem value via an absurdly large
// font size to try to provoke an error, or alternatively confirm the normal path.
func TestMeasureText_GlyphAdvanceError_LargeFontSize(t *testing.T) {
	w := NewWriter()
	fm := NewPDFFontManager(w)
	alias := fm.RegisterFont("Arial", false, false)

	// A very large font size (1e9 pt). This may or may not trigger the
	// GlyphAdvance error path depending on the sfnt implementation.
	// Either way the function must return a non-negative width.
	width := fm.MeasureText(alias, "hello", 1e9)
	if width < 0 {
		t.Errorf("MeasureText with giant font size returned negative width: %v", width)
	}
}

// TestRegisterFont_Mono_AllVariants exercises all four bold/italic combinations
// for the mono family to ensure maximum coverage of RegisterFont.
func TestRegisterFont_Mono_AllVariants(t *testing.T) {
	tests := []struct{ bold, italic bool }{
		{false, false},
		{true, false},
		{false, true},
		{true, true},
	}
	for _, tt := range tests {
		w := NewWriter()
		fm := NewPDFFontManager(w)
		alias := fm.RegisterFont("Courier New", tt.bold, tt.italic)
		if alias == "" {
			t.Errorf("RegisterFont(Courier, bold=%v, italic=%v) returned empty alias", tt.bold, tt.italic)
		}
		// Second call should return cached alias.
		alias2 := fm.RegisterFont("Courier New", tt.bold, tt.italic)
		if alias != alias2 {
			t.Errorf("RegisterFont second call returned different alias: %q vs %q", alias, alias2)
		}
	}
}
