package utils

import (
	"testing"

	"golang.org/x/image/font/basicfont"

	"github.com/andrewloable/go-fastreport/style"
)

// ── lineHeight ────────────────────────────────────────────────────────────────

func TestLineHeight_NilFace_NonZeroSize(t *testing.T) {
	f := style.DefaultFont()
	h := lineHeight(nil, f)
	if h <= 0 {
		t.Errorf("lineHeight(nil, defaultFont) = %v, want > 0", h)
	}
}

func TestLineHeight_NilFace_ZeroSize(t *testing.T) {
	f := style.Font{Size: 0}
	h := lineHeight(nil, f)
	if h <= 0 {
		t.Errorf("lineHeight(nil, zeroSizeFont) = %v, want > 0 (uses default)", h)
	}
}

func TestLineHeight_WithFace(t *testing.T) {
	face := basicfont.Face7x13
	f := style.DefaultFont()
	h := lineHeight(face, f)
	if h <= 0 {
		t.Errorf("lineHeight(face, font) = %v, want > 0", h)
	}
}

// ── measureLine ───────────────────────────────────────────────────────────────

func TestMeasureLine_NilFace(t *testing.T) {
	w := measureLine("hello", nil)
	if w != 0 {
		t.Errorf("measureLine(nil face) = %v, want 0", w)
	}
}

func TestMeasureLine_EmptyLine(t *testing.T) {
	face := basicfont.Face7x13
	w := measureLine("", face)
	if w != 0 {
		t.Errorf("measureLine(empty, face) = %v, want 0", w)
	}
}

func TestMeasureLine_NonEmpty(t *testing.T) {
	face := basicfont.Face7x13
	w := measureLine("hello", face)
	if w <= 0 {
		t.Errorf("measureLine(hello, face) = %v, want > 0", w)
	}
}

// ── wordWrap ──────────────────────────────────────────────────────────────────

func TestWordWrap_Empty(t *testing.T) {
	face := basicfont.Face7x13
	lines := wordWrap("", face, 100)
	if len(lines) != 1 || lines[0] != "" {
		t.Errorf("wordWrap empty = %v, want [\"\"]", lines)
	}
}

func TestWordWrap_WhitespaceOnly(t *testing.T) {
	face := basicfont.Face7x13
	// All whitespace → Fields returns nil → len(words) == 0 branch
	lines := wordWrap("   ", face, 100)
	if len(lines) != 1 {
		t.Errorf("wordWrap whitespace = %v, want 1 empty line", lines)
	}
}

func TestWordWrap_SingleWord(t *testing.T) {
	face := basicfont.Face7x13
	lines := wordWrap("hello", face, 100)
	if len(lines) != 1 || lines[0] != "hello" {
		t.Errorf("wordWrap single = %v, want [hello]", lines)
	}
}

func TestWordWrap_NarrowWidth_Wraps(t *testing.T) {
	face := basicfont.Face7x13
	// Very narrow — each word should be on its own line.
	lines := wordWrap("one two three", face, 1)
	if len(lines) < 2 {
		t.Errorf("wordWrap narrow = %d lines, want >= 2", len(lines))
	}
}
