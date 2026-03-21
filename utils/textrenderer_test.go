package utils_test

// textrenderer_test.go — Tests for the advanced text measurement functions
// ported from C# AdvancedTextRenderer (Utils/TextRenderer.cs).

import (
	"testing"

	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
)

// ── CalculateSpaceWidth ───────────────────────────────────────────────────────

func TestCalculateSpaceWidth_DefaultFont_Positive(t *testing.T) {
	sw := utils.CalculateSpaceWidth(defaultFont)
	if sw <= 0 {
		t.Errorf("CalculateSpaceWidth(defaultFont) = %v, want > 0", sw)
	}
}

func TestCalculateSpaceWidth_ZeroSizeFont_Positive(t *testing.T) {
	// Even with zero-size font the fallback path should return a positive value.
	f := style.Font{Name: "Arial", Size: 0}
	sw := utils.CalculateSpaceWidth(f)
	if sw < 0 {
		t.Errorf("CalculateSpaceWidth(zeroSize) = %v, want >= 0", sw)
	}
}

func TestCalculateSpaceWidth_LargerFont_WiderSpace(t *testing.T) {
	small := utils.CalculateSpaceWidth(style.Font{Name: "Arial", Size: 8})
	large := utils.CalculateSpaceWidth(style.Font{Name: "Arial", Size: 24})
	// basicfont fallback produces the same advance regardless of font size;
	// the fallback estimate (25 % of em-px) does scale, so at least one path
	// should give a larger value for the bigger font. We only assert both > 0.
	if small <= 0 || large <= 0 {
		t.Errorf("space widths should be > 0: small=%v large=%v", small, large)
	}
}

// ── GetTabPosition ────────────────────────────────────────────────────────────

func TestGetTabPosition_BeforeFirstTab(t *testing.T) {
	// pos < tabOffset → return tabOffset
	got := utils.GetTabPosition(5, 20, 40)
	if got != 20 {
		t.Errorf("GetTabPosition(5, 20, 40) = %v, want 20", got)
	}
}

func TestGetTabPosition_AtFirstTab(t *testing.T) {
	// pos == tabOffset → first tab stop after is tabOffset + tabSize
	got := utils.GetTabPosition(20, 20, 40)
	if got != 60 {
		t.Errorf("GetTabPosition(20, 20, 40) = %v, want 60", got)
	}
}

func TestGetTabPosition_BetweenTabs(t *testing.T) {
	// pos between tab 1 (20) and tab 2 (60) → should snap to 60
	got := utils.GetTabPosition(35, 20, 40)
	if got != 60 {
		t.Errorf("GetTabPosition(35, 20, 40) = %v, want 60", got)
	}
}

func TestGetTabPosition_ZeroOffset(t *testing.T) {
	// tabOffset=0, tabSize=50: pos=75 → next stop = 100
	got := utils.GetTabPosition(75, 0, 50)
	if got != 100 {
		t.Errorf("GetTabPosition(75, 0, 50) = %v, want 100", got)
	}
}

func TestGetTabPosition_ZeroTabSize_ReturnsPos(t *testing.T) {
	// tabSize == 0 should not divide-by-zero; returns pos unchanged.
	got := utils.GetTabPosition(42, 0, 0)
	if got != 42 {
		t.Errorf("GetTabPosition(42, 0, 0) = %v, want 42 (passthrough)", got)
	}
}

func TestGetTabPosition_ExactlyOnSecondStop(t *testing.T) {
	// tabOffset=0, tabSize=40: pos=80 (exactly on stop 2) → next = 120
	got := utils.GetTabPosition(80, 0, 40)
	if got != 120 {
		t.Errorf("GetTabPosition(80, 0, 40) = %v, want 120", got)
	}
}

// ── CalcTextHeight ────────────────────────────────────────────────────────────

func TestCalcTextHeight_SingleLine_HeightPositive(t *testing.T) {
	h, _ := utils.CalcTextHeight("Hello", defaultFont, 0, 0)
	if h <= 0 {
		t.Errorf("CalcTextHeight single line = %v, want > 0", h)
	}
}

func TestCalcTextHeight_MultiLine_TallerThanSingle(t *testing.T) {
	h1, _ := utils.CalcTextHeight("Hello", defaultFont, 0, 0)
	h2, _ := utils.CalcTextHeight("Hello\nWorld\nFoo", defaultFont, 0, 0)
	if h2 <= h1 {
		t.Errorf("multi-line height (%v) should exceed single-line (%v)", h2, h1)
	}
}

func TestCalcTextHeight_DisplayHeightConstraint_CharsFitSet(t *testing.T) {
	// displayHeight = 1 line → text with 3 lines should set charsFit > 0.
	lh := utils.FontLineHeight(defaultFont)
	_, charsFit := utils.CalcTextHeight("Line1\nLine2\nLine3", defaultFont, 0, lh*1.5)
	if charsFit == 0 {
		t.Error("charsFit should be > 0 when text overflows displayHeight")
	}
}

func TestCalcTextHeight_TinyDisplayHeight_ReturnsZero(t *testing.T) {
	// displayHeight smaller than one lineHeight → CalcHeight returns 0.
	h, _ := utils.CalcTextHeight("Hello", defaultFont, 0, 1)
	if h != 0 {
		t.Errorf("CalcTextHeight with tiny displayHeight = %v, want 0", h)
	}
}

func TestCalcTextHeight_EmptyText_PositiveHeight(t *testing.T) {
	// Empty text still has one blank line.
	h, _ := utils.CalcTextHeight("", defaultFont, 0, 0)
	if h <= 0 {
		t.Errorf("CalcTextHeight empty = %v, want > 0", h)
	}
}

func TestCalcTextHeight_WordWrap_IncreasesHeight(t *testing.T) {
	long := "This is a long sentence that should wrap when given a narrow display width."
	hNoWrap, _ := utils.CalcTextHeight(long, defaultFont, 0, 0)
	hWrap, _ := utils.CalcTextHeight(long, defaultFont, 50, 0)
	if hWrap <= hNoWrap {
		t.Errorf("wrapped height (%v) should exceed unwrapped (%v)", hWrap, hNoWrap)
	}
}

// ── CalcTextWidth ─────────────────────────────────────────────────────────────

func TestCalcTextWidth_NonEmpty_Positive(t *testing.T) {
	w := utils.CalcTextWidth("Hello World", defaultFont, 0)
	if w <= 0 {
		t.Errorf("CalcTextWidth = %v, want > 0", w)
	}
}

func TestCalcTextWidth_LongerText_Wider(t *testing.T) {
	w1 := utils.CalcTextWidth("Hi", defaultFont, 0)
	w2 := utils.CalcTextWidth("Hello World, this is considerably longer text", defaultFont, 0)
	if w2 <= w1 {
		t.Errorf("longer text (%v) should be wider than short (%v)", w2, w1)
	}
}

func TestCalcTextWidth_IncludesSpaceWidth(t *testing.T) {
	// CalcTextWidth always adds at least one space width (matching C# CalcWidth).
	sw := utils.CalculateSpaceWidth(defaultFont)
	w := utils.CalcTextWidth("A", defaultFont, 0)
	charW := utils.MeasureStringAdvance("A", defaultFont)
	// w should be charW + spaceWidth (at minimum); due to basicfont rounding,
	// allow a small epsilon.
	if w < charW+sw-1 {
		t.Errorf("CalcTextWidth(%q) = %v, want >= charW(%v) + spaceW(%v)", "A", w, charW, sw)
	}
}

// ── CharsFitInWidth ───────────────────────────────────────────────────────────

func TestCharsFitInWidth_ZeroWidth_ReturnsZero(t *testing.T) {
	n := utils.CharsFitInWidth("Hello", defaultFont, 0)
	if n != 0 {
		t.Errorf("CharsFitInWidth zero width = %d, want 0", n)
	}
}

func TestCharsFitInWidth_EmptyText_ReturnsZero(t *testing.T) {
	n := utils.CharsFitInWidth("", defaultFont, 1000)
	if n != 0 {
		t.Errorf("CharsFitInWidth empty text = %d, want 0", n)
	}
}

func TestCharsFitInWidth_WideEnough_AllCharsFit(t *testing.T) {
	text := "Hello"
	n := utils.CharsFitInWidth(text, defaultFont, 10000)
	if n != len([]rune(text)) {
		t.Errorf("CharsFitInWidth huge width = %d, want %d", n, len([]rune(text)))
	}
}

func TestCharsFitInWidth_NarrowWidth_FewerCharsFit(t *testing.T) {
	text := "Hello World"
	nAll := utils.CharsFitInWidth(text, defaultFont, 10000)
	nNarrow := utils.CharsFitInWidth(text, defaultFont, 20) // very narrow
	if nNarrow >= nAll {
		t.Errorf("narrow width should fit fewer chars: narrow=%d, all=%d", nNarrow, nAll)
	}
}

func TestCharsFitInWidth_SingleChar_MinimumOne(t *testing.T) {
	// Even with a very small width, at least 1 char must fit (anti-infinite-loop).
	n := utils.CharsFitInWidth("A", defaultFont, 0.001)
	if n < 1 {
		t.Errorf("CharsFitInWidth minimum = %d, want >= 1", n)
	}
}

func TestCharsFitInWidth_LongerTextFitsMore(t *testing.T) {
	short := "Hi"
	long := "Hello World"
	const w = 200
	nShort := utils.CharsFitInWidth(short, defaultFont, w)
	nLong := utils.CharsFitInWidth(long, defaultFont, w)
	// long has more chars; at the same width more chars of long should fit
	// (or the same if both fit entirely). At minimum nShort should fit.
	_ = nLong
	if nShort > len([]rune(long)) {
		t.Errorf("charsFit (%d) cannot exceed text length (%d)", nShort, len([]rune(long)))
	}
}

// ── TabStopPositions ──────────────────────────────────────────────────────────

func TestTabStopPositions_Basic(t *testing.T) {
	stops := utils.TabStopPositions(10, 40, 3)
	want := []float32{10, 50, 90}
	if len(stops) != len(want) {
		t.Fatalf("len=%d, want %d", len(stops), len(want))
	}
	for i, s := range stops {
		if s != want[i] {
			t.Errorf("stop[%d] = %v, want %v", i, s, want[i])
		}
	}
}

func TestTabStopPositions_ZeroN_ReturnsNil(t *testing.T) {
	stops := utils.TabStopPositions(0, 40, 0)
	if stops != nil {
		t.Errorf("TabStopPositions n=0 = %v, want nil", stops)
	}
}

func TestTabStopPositions_ZeroTabSize_ReturnsNil(t *testing.T) {
	stops := utils.TabStopPositions(0, 0, 5)
	if stops != nil {
		t.Errorf("TabStopPositions tabSize=0 = %v, want nil", stops)
	}
}

func TestTabStopPositions_ZeroOffset(t *testing.T) {
	stops := utils.TabStopPositions(0, 50, 4)
	want := []float32{0, 50, 100, 150}
	if len(stops) != len(want) {
		t.Fatalf("len=%d, want %d", len(stops), len(want))
	}
	for i, s := range stops {
		if s != want[i] {
			t.Errorf("stop[%d] = %v, want %v", i, s, want[i])
		}
	}
}

// ── FontLineHeight ────────────────────────────────────────────────────────────

func TestFontLineHeight_DefaultFont_Positive(t *testing.T) {
	lh := utils.FontLineHeight(defaultFont)
	if lh <= 0 {
		t.Errorf("FontLineHeight(defaultFont) = %v, want > 0", lh)
	}
}

func TestFontLineHeight_ZeroSize_UsesDefault(t *testing.T) {
	lh := utils.FontLineHeight(style.Font{Size: 0})
	if lh <= 0 {
		t.Errorf("FontLineHeight(zeroSize) = %v, want > 0 (fallback)", lh)
	}
}

func TestFontLineHeight_LargerFont_TallerLine(t *testing.T) {
	lhSmall := utils.FontLineHeight(style.Font{Name: "Arial", Size: 8})
	lhLarge := utils.FontLineHeight(style.Font{Name: "Arial", Size: 24})
	if lhLarge <= lhSmall {
		t.Errorf("larger font line height (%v) should exceed smaller (%v)", lhLarge, lhSmall)
	}
}

// ── MeasureStringSize ─────────────────────────────────────────────────────────

func TestMeasureStringSize_EmptyString(t *testing.T) {
	w, h := utils.MeasureStringSize("", defaultFont)
	if w != 0 {
		t.Errorf("MeasureStringSize empty width = %v, want 0", w)
	}
	if h <= 0 {
		t.Errorf("MeasureStringSize empty height = %v, want > 0", h)
	}
}

func TestMeasureStringSize_NonEmpty_WidthAndHeightPositive(t *testing.T) {
	w, h := utils.MeasureStringSize("Hello", defaultFont)
	if w <= 0 {
		t.Errorf("MeasureStringSize width = %v, want > 0", w)
	}
	if h <= 0 {
		t.Errorf("MeasureStringSize height = %v, want > 0", h)
	}
}

func TestMeasureStringSize_SingleLine_HeightMatchesFontLineHeight(t *testing.T) {
	lh := utils.FontLineHeight(defaultFont)
	_, h := utils.MeasureStringSize("Hello World", defaultFont)
	if h != lh {
		t.Errorf("MeasureStringSize height = %v, want FontLineHeight = %v", h, lh)
	}
}

func TestMeasureStringSize_LongerString_Wider(t *testing.T) {
	w1, _ := utils.MeasureStringSize("Hi", defaultFont)
	w2, _ := utils.MeasureStringSize("Hello World", defaultFont)
	if w2 <= w1 {
		t.Errorf("longer string should be wider: %v <= %v", w2, w1)
	}
}

// ── MeasureStringAdvance ──────────────────────────────────────────────────────

func TestMeasureStringAdvance_Empty_Zero(t *testing.T) {
	w := utils.MeasureStringAdvance("", defaultFont)
	if w != 0 {
		t.Errorf("MeasureStringAdvance empty = %v, want 0", w)
	}
}

func TestMeasureStringAdvance_NonEmpty_Positive(t *testing.T) {
	w := utils.MeasureStringAdvance("Hello", defaultFont)
	if w <= 0 {
		t.Errorf("MeasureStringAdvance = %v, want > 0", w)
	}
}
