package utils

import (
	"strings"
	"testing"

	"golang.org/x/image/font/basicfont"
)

// face is a convenience alias used throughout the tests.
var testFace = basicfont.Face7x13

// ---- FontHeight -------------------------------------------------------

func TestFontHeight(t *testing.T) {
	h := FontHeight(testFace)
	if h <= 0 {
		t.Errorf("FontHeight() = %v, want > 0", h)
	}
}

// ---- PixelsToPoints / PointsToPixels ----------------------------------

func TestPixelsToPoints(t *testing.T) {
	// 96 px → 72 pts (one inch at 96 DPI equals 72 points)
	got := PixelsToPoints(96)
	if got != 72 {
		t.Errorf("PixelsToPoints(96) = %v, want 72", got)
	}
}

func TestPointsToPixels(t *testing.T) {
	// 72 pts → 96 px
	got := PointsToPixels(72)
	if got != 96 {
		t.Errorf("PointsToPixels(72) = %v, want 96", got)
	}
}

func TestPixelsPointsRoundtrip(t *testing.T) {
	original := float32(48.0)
	roundtripped := PixelsToPoints(PointsToPixels(original))
	if roundtripped != original {
		t.Errorf("roundtrip: got %v, want %v", roundtripped, original)
	}
}

// ---- MeasureLine ------------------------------------------------------

func TestMeasureLine_Empty(t *testing.T) {
	w, h := MeasureLine("", testFace)
	if w != 0 {
		t.Errorf("MeasureLine(\"\") width = %v, want 0", w)
	}
	if h <= 0 {
		t.Errorf("MeasureLine(\"\") height = %v, want > 0", h)
	}
}

func TestMeasureLine_Short(t *testing.T) {
	w, h := MeasureLine("Hello", testFace)
	if w <= 0 {
		t.Errorf("MeasureLine(\"Hello\") width = %v, want > 0", w)
	}
	if h <= 0 {
		t.Errorf("MeasureLine(\"Hello\") height = %v, want > 0", h)
	}
}

func TestMeasureLine_LongerWider(t *testing.T) {
	wShort, _ := MeasureLine("Hi", testFace)
	wLong, _ := MeasureLine("Hello World", testFace)
	if wLong <= wShort {
		t.Errorf("longer string should be wider: short=%v long=%v", wShort, wLong)
	}
}

// ---- WrapText ---------------------------------------------------------

func TestWrapText_NoWrapNeeded(t *testing.T) {
	lines := WrapText("Hi", testFace, 500)
	if len(lines) != 1 || lines[0] != "Hi" {
		t.Errorf("WrapText short text: got %v", lines)
	}
}

func TestWrapText_HardNewline(t *testing.T) {
	lines := WrapText("Line1\nLine2", testFace, 1000)
	if len(lines) != 2 {
		t.Errorf("WrapText hard newline: got %d lines, want 2; lines=%v", len(lines), lines)
	}
	if lines[0] != "Line1" || lines[1] != "Line2" {
		t.Errorf("WrapText hard newline: got %v", lines)
	}
}

func TestWrapText_EmptyString(t *testing.T) {
	lines := WrapText("", testFace, 100)
	if len(lines) != 1 {
		t.Errorf("WrapText empty: got %d lines, want 1", len(lines))
	}
	if lines[0] != "" {
		t.Errorf("WrapText empty: got %q, want \"\"", lines[0])
	}
}

func TestWrapText_WordWrap(t *testing.T) {
	// Very narrow width forces wrapping.
	long := "one two three four five"
	lines := WrapText(long, testFace, 30)
	if len(lines) <= 1 {
		t.Errorf("WrapText narrow: expected multiple lines, got %v", lines)
	}
	// Reassembled text must contain all original words.
	joined := strings.Join(lines, " ")
	for _, w := range strings.Fields(long) {
		if !strings.Contains(joined, w) {
			t.Errorf("WrapText: word %q missing from wrapped lines", w)
		}
	}
}

func TestWrapText_SingleWordWiderThanMax(t *testing.T) {
	// A single very long word cannot be split; it must appear on its own line.
	lines := WrapText("superlongwordthatcannotfit", testFace, 1)
	if len(lines) != 1 {
		t.Errorf("WrapText single long word: got %d lines, want 1", len(lines))
	}
}

func TestWrapText_MultipleHardNewlines(t *testing.T) {
	lines := WrapText("a\nb\nc", testFace, 1000)
	if len(lines) != 3 {
		t.Errorf("WrapText multiple hard newlines: got %d, want 3; %v", len(lines), lines)
	}
}

func TestWrapText_EmptyLine(t *testing.T) {
	// A hard newline with nothing on the second line.
	lines := WrapText("hello\n", testFace, 1000)
	if len(lines) != 2 {
		t.Errorf("WrapText trailing newline: got %d lines, want 2; %v", len(lines), lines)
	}
}

func TestWrapText_WhitespaceOnlyLine(t *testing.T) {
	// A line containing only spaces has no words after Fields(); it should
	// produce a single empty-string line (the len(words)==0 branch).
	lines := WrapText("   ", testFace, 1000)
	if len(lines) != 1 {
		t.Errorf("WrapText whitespace-only: got %d lines, want 1; %v", len(lines), lines)
	}
	if lines[0] != "" {
		t.Errorf("WrapText whitespace-only: got %q, want \"\"", lines[0])
	}
}

// ---- MeasureString ----------------------------------------------------

func TestMeasureString_Empty(t *testing.T) {
	m := MeasureString("", testFace, 0)
	if m.Lines != 1 {
		t.Errorf("MeasureString empty: Lines=%d, want 1", m.Lines)
	}
	if m.Width != 0 {
		t.Errorf("MeasureString empty: Width=%v, want 0", m.Width)
	}
	if m.Height <= 0 {
		t.Errorf("MeasureString empty: Height=%v, want > 0", m.Height)
	}
}

func TestMeasureString_SingleLine(t *testing.T) {
	m := MeasureString("Hello", testFace, 0)
	if m.Lines != 1 {
		t.Errorf("MeasureString single: Lines=%d, want 1", m.Lines)
	}
	if m.Width <= 0 {
		t.Errorf("MeasureString single: Width=%v, want > 0", m.Width)
	}
	if m.Height <= 0 {
		t.Errorf("MeasureString single: Height=%v, want > 0", m.Height)
	}
}

func TestMeasureString_HardNewlines(t *testing.T) {
	m := MeasureString("Line1\nLine2\nLine3", testFace, 0)
	if m.Lines != 3 {
		t.Errorf("MeasureString multiline: Lines=%d, want 3", m.Lines)
	}
	if m.Height <= 0 {
		t.Errorf("MeasureString multiline: Height=%v, want > 0", m.Height)
	}
}

func TestMeasureString_WithMaxWidth(t *testing.T) {
	text := "one two three four five six seven eight"
	mNarrow := MeasureString(text, testFace, 40)
	mWide := MeasureString(text, testFace, 0)

	if mNarrow.Lines <= mWide.Lines {
		t.Errorf("narrow wrap should produce more lines: narrow=%d wide=%d",
			mNarrow.Lines, mWide.Lines)
	}
}

func TestMeasureString_HeightScalesWithLines(t *testing.T) {
	h1 := MeasureString("one", testFace, 0).Height
	h2 := MeasureString("one\ntwo", testFace, 0).Height
	if h2 <= h1 {
		t.Errorf("two lines should be taller: h1=%v h2=%v", h1, h2)
	}
}

func TestMeasureString_MaxWidthZeroNoWrap(t *testing.T) {
	text := "word1 word2 word3"
	m := MeasureString(text, testFace, 0)
	if m.Lines != 1 {
		t.Errorf("maxWidth=0 should not wrap: got %d lines", m.Lines)
	}
}
