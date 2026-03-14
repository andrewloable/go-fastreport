package utils_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
)

var defaultFont = style.DefaultFont()

func TestMeasureText_EmptyString(t *testing.T) {
	w, h := utils.MeasureText("", defaultFont, 0)
	if w != 0 {
		t.Errorf("empty string width: got %v, want 0", w)
	}
	if h <= 0 {
		t.Errorf("empty string height should be > 0, got %v", h)
	}
}

func TestMeasureText_SingleLine(t *testing.T) {
	w, h := utils.MeasureText("Hello World", defaultFont, 0)
	if w <= 0 {
		t.Errorf("width should be > 0, got %v", w)
	}
	if h <= 0 {
		t.Errorf("height should be > 0, got %v", h)
	}
}

func TestMeasureText_WiderTextIsWider(t *testing.T) {
	w1, _ := utils.MeasureText("Hi", defaultFont, 0)
	w2, _ := utils.MeasureText("Hello World, this is a longer string", defaultFont, 0)
	if w2 <= w1 {
		t.Errorf("longer text should be wider: %v <= %v", w2, w1)
	}
}

func TestMeasureText_ExplicitNewlines_IncreasesHeight(t *testing.T) {
	_, h1 := utils.MeasureText("Line one", defaultFont, 0)
	_, h2 := utils.MeasureText("Line one\nLine two\nLine three", defaultFont, 0)
	if h2 <= h1 {
		t.Errorf("multiline height should be larger: %v <= %v", h2, h1)
	}
}

func TestMeasureText_WordWrap_IncreasesHeight(t *testing.T) {
	longText := "This is a fairly long sentence that should wrap when constrained to a narrow width."
	// Measure without wrap (full width available).
	_, hNoWrap := utils.MeasureText(longText, defaultFont, 0)
	// Measure with narrow wrap.
	_, hWrapped := utils.MeasureText(longText, defaultFont, 50)
	if hWrapped <= hNoWrap {
		t.Errorf("wrapped height (%v) should exceed single-line height (%v)", hWrapped, hNoWrap)
	}
}

func TestMeasureLines_SingleLine(t *testing.T) {
	n := utils.MeasureLines("Hello", defaultFont, 0)
	if n != 1 {
		t.Errorf("single word: got %d lines, want 1", n)
	}
}

func TestMeasureLines_ExplicitNewlines(t *testing.T) {
	n := utils.MeasureLines("A\nB\nC", defaultFont, 0)
	if n != 3 {
		t.Errorf("3 lines: got %d, want 3", n)
	}
}

func TestMeasureLines_WordWrap(t *testing.T) {
	// Very narrow max width should force wrapping.
	text := "word1 word2 word3 word4 word5"
	n := utils.MeasureLines(text, defaultFont, 10) // 10px is very narrow
	if n <= 1 {
		t.Errorf("narrow wrap should produce > 1 line, got %d", n)
	}
}

func TestMeasureLines_EmptyString(t *testing.T) {
	n := utils.MeasureLines("", defaultFont, 100)
	if n != 1 {
		t.Errorf("empty string: got %d lines, want 1", n)
	}
}

func TestEstimateTextWidth_Empty(t *testing.T) {
	w := utils.EstimateTextWidth("", 7)
	if w != 0 {
		t.Errorf("empty: got %v, want 0", w)
	}
}

func TestEstimateTextWidth_SingleLine(t *testing.T) {
	w := utils.EstimateTextWidth("Hello", 7)
	if w != 35 { // 5 chars × 7px
		t.Errorf("got %v, want 35", w)
	}
}

func TestEstimateTextWidth_MultiLine_UsesWidestLine(t *testing.T) {
	w := utils.EstimateTextWidth("Hi\nHello World\nBye", 7)
	// "Hello World" = 11 chars → 77px
	if w != 77 {
		t.Errorf("got %v, want 77", w)
	}
}

func TestEstimateTextWidth_DefaultCharWidth(t *testing.T) {
	// charWidth=0 should use default (7).
	w := utils.EstimateTextWidth("ABC", 0)
	if w != 21 {
		t.Errorf("got %v, want 21", w)
	}
}
