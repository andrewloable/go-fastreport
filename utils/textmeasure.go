package utils

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/andrewloable/go-fastreport/style"
)

// MeasureText measures the rendered width and height of text when drawn with
// the given font, optionally wrapping at maxWidth pixels (0 = no wrap).
// The returned width is the maximum line width; height is lineHeight * lineCount.
func MeasureText(text string, f style.Font, maxWidth float32) (width, height float32) {
	face := faceForStyle(f)
	lineH := lineHeight(face, f)
	if text == "" {
		return 0, lineH
	}

	lines := wrapLines(text, face, maxWidth)
	var maxW float32
	for _, line := range lines {
		w := measureLine(line, face)
		if w > maxW {
			maxW = w
		}
	}
	return maxW, lineH * float32(len(lines))
}

// MeasureLines returns the number of visual lines produced by word-wrapping
// text at maxWidth pixels using the given font.
// If maxWidth <= 0, each \n-delimited paragraph is one line.
func MeasureLines(text string, f style.Font, maxWidth float32) int {
	if text == "" {
		return 1
	}
	face := faceForStyle(f)
	return len(wrapLines(text, face, maxWidth))
}

// ── internal helpers ──────────────────────────────────────────────────────────

// faceForStyle returns a font.Face for the given style.Font.
// Uses DefaultFontManager if a matching face is registered; otherwise falls
// back to basicfont.Face7x13 scaled by the font size ratio.
func faceForStyle(f style.Font) font.Face {
	desc := FontDescriptor{
		Family: f.Name,
		Size:   f.Size,
	}
	return DefaultFontManager.GetFace(desc)
}

// lineHeight estimates the line height in pixels for a given face and style.Font.
// If the face provides metrics, use them; otherwise fall back to font.Size.
func lineHeight(face font.Face, f style.Font) float32 {
	if face != nil {
		m := face.Metrics()
		h := m.Ascent + m.Descent
		if h > 0 {
			// Convert fixed.Int26_6 to float32 pixels.
			return float32(h) / 64.0
		}
	}
	// Fallback: approximate line height as 1.2× the point size.
	sz := f.Size
	if sz <= 0 {
		sz = style.DefaultFont().Size
	}
	return sz * 1.2
}

// measureLine returns the pixel width of a single line of text.
func measureLine(line string, face font.Face) float32 {
	if face == nil || line == "" {
		return 0
	}
	advance := font.MeasureString(face, line)
	return float32(advance) / 64.0
}

// wrapLines splits text into display lines, respecting explicit \n breaks and
// word-wrapping at maxWidth pixels (if maxWidth > 0).
func wrapLines(text string, face font.Face, maxWidth float32) []string {
	// Split on explicit newlines first.
	paragraphs := strings.Split(text, "\n")
	if maxWidth <= 0 || face == nil {
		return paragraphs
	}

	var result []string
	for _, para := range paragraphs {
		result = append(result, wordWrap(para, face, maxWidth)...)
	}
	return result
}

// wordWrap wraps a single paragraph into lines that fit within maxWidth.
func wordWrap(para string, face font.Face, maxWidth float32) []string {
	if para == "" {
		return []string{""}
	}
	maxFixed := fixed.Int26_6(maxWidth * 64)

	words := strings.Fields(para)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	currentLine := ""
	currentWidth := fixed.Int26_6(0)
	spaceWidth, _ := face.GlyphAdvance(' ')

	for i, word := range words {
		wordWidth := font.MeasureString(face, word)
		if currentLine == "" {
			// First word on this line — always include it, even if it overflows.
			currentLine = word
			currentWidth = wordWidth
		} else {
			// Check if word fits on current line.
			needed := currentWidth + spaceWidth + wordWidth
			if needed <= maxFixed {
				currentLine += " " + word
				currentWidth = needed
			} else {
				lines = append(lines, currentLine)
				currentLine = word
				currentWidth = wordWidth
			}
		}
		if i == len(words)-1 {
			lines = append(lines, currentLine)
		}
	}
	return lines
}

// EstimateTextWidth returns a quick character-count-based width estimate
// (monospace approximation) when no font metrics are available.
// charWidth is the assumed width per character in pixels.
func EstimateTextWidth(text string, charWidth float32) float32 {
	if charWidth <= 0 {
		charWidth = 7 // pixels per char at ~10pt monospace
	}
	maxLen := 0
	for _, line := range strings.Split(text, "\n") {
		n := utf8.RuneCountInString(line)
		if n > maxLen {
			maxLen = n
		}
	}
	return float32(maxLen) * charWidth
}
