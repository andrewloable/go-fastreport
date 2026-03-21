package utils

// textrenderer.go — Advanced text measurement helpers ported from
// FastReport.Base/Utils/TextRenderer.cs (AdvancedTextRenderer class).
//
// The C# AdvancedTextRenderer performs rich layout (paragraphs, lines, words,
// HTML tags, justification) backed by System.Drawing.Graphics.MeasureString.
// In pure-Go we cannot call GDI+, so this file provides the measurement
// primitives that the engine and PDF exporter need:
//
//   - CalculateSpaceWidth   — pixel width of one space character
//   - GetTabPosition        — next tab stop position for a given X offset
//   - CalcTextHeight        — total pixel height of text in a bounding rect
//   - CalcTextWidth         — maximum pixel width of any wrapped line
//   - CharsFitInWidth       — how many leading characters of a string fit
//                             within a given pixel width
//
// All measurements use the same font metrics as textmeasure.go (faceForStyle /
// lineHeight) so results are consistent across the package.

import (
	"github.com/andrewloable/go-fastreport/style"
	"golang.org/x/image/font"
)

// CalculateSpaceWidth returns the pixel width of a single space character for
// the given font. This matches C# AdvancedTextRenderer.CalculateSpaceSize:
//
//	w_ab  = MeasureString("abcdefabcdef",      font)
//	w_a40b= MeasureString("abcdef{40 spaces}abcdef", font)
//	space = (w_a40b - w_ab) / 40
//
// Ref: TextRenderer.cs CalculateSpaceSize (line ~241)
func CalculateSpaceWidth(f style.Font) float32 {
	face := faceForStyle(f)
	if face == nil {
		// Fall back to a proportional estimate: ~25 % of em height.
		fontPx := f.Size * 96.0 / 72.0
		return fontPx * 0.25
	}
	adv, ok := face.GlyphAdvance(' ')
	if !ok {
		fontPx := f.Size * 96.0 / 72.0
		return fontPx * 0.25
	}
	return float32(adv) / 64.0
}

// GetTabPosition returns the X position of the next tab stop after pos.
// tabOffset is the position of the first tab stop; tabSize is the distance
// between consecutive tab stops (both in pixels).
//
// Ref: TextRenderer.cs AdvancedTextRenderer.GetTabPosition (line ~391)
//
//	tabPosition = int((pos - tabOffset) / tabSize)
//	if pos < tabOffset → return tabOffset
//	return (tabPosition + 1) * tabSize + tabOffset
func GetTabPosition(pos, tabOffset, tabSize float32) float32 {
	if tabSize <= 0 {
		return pos
	}
	if pos < tabOffset {
		return tabOffset
	}
	tabPosition := int((pos - tabOffset) / tabSize)
	return float32(tabPosition+1)*tabSize + tabOffset
}

// CalcTextHeight returns the total pixel height occupied by text when rendered
// inside a box of the given displayWidth. It sums line heights for all wrapped
// lines and stops accumulating once height exceeds displayHeight (returns the
// full content height regardless, matching C# CalcHeight behaviour).
//
// charsFit is set to the character index at which the text first overflows
// displayHeight (0 means the entire text fits).
//
// Ref: TextRenderer.cs AdvancedTextRenderer.CalcHeight (line ~340)
func CalcTextHeight(text string, f style.Font, displayWidth, displayHeight float32) (height float32, charsFit int) {
	face := faceForStyle(f)
	lh := lineHeight(face, f)

	// If even a single line does not fit, return 0 per C# logic.
	if lh > displayHeight && displayHeight > 0 {
		return 0, 0
	}

	lines := wrapLines(text, face, scaleMaxWidth(displayWidth, f))
	charsFit = 0
	charIdx := 0

	for _, line := range lines {
		height += lh
		if charsFit == 0 && displayHeight > 0 && height > displayHeight {
			charsFit = charIdx
		}
		charIdx += len([]rune(line)) + 1 // +1 for the newline separator
	}

	if charsFit == 0 {
		charsFit = len([]rune(text))
	}
	return height, charsFit
}

// CalcTextWidth returns the maximum pixel width of any wrapped line of text
// plus a trailing space width (matching C# CalcWidth which adds spaceWidth).
//
// Ref: TextRenderer.cs AdvancedTextRenderer.CalcWidth (line ~375)
func CalcTextWidth(text string, f style.Font, maxWidth float32) float32 {
	face := faceForStyle(f)
	lines := wrapLines(text, face, scaleMaxWidth(maxWidth, f))
	var maxW float32
	for _, line := range lines {
		w := measureLine(line, face)
		if w > maxW {
			maxW = w
		}
	}
	return maxW + CalculateSpaceWidth(f)
}

// CharsFitInWidth returns how many leading characters (runes) of text fit
// within maxWidth pixels when rendered with font f.
//
// This matches the inner MeasureString loop in C# Paragraph.WrapLines:
// it progressively adds characters until the measured width exceeds maxWidth,
// then returns the count that fit. The minimum return value is 1 (a single
// character always "fits" even if it overflows, to prevent infinite loops).
//
// Ref: TextRenderer.cs Paragraph.WrapLines + MeasureString (line ~550)
func CharsFitInWidth(text string, f style.Font, maxWidth float32) int {
	if maxWidth <= 0 || text == "" {
		return 0
	}
	face := faceForStyle(f)
	if face == nil {
		return 0
	}

	runes := []rune(text)
	scaledMax := scaleMaxWidth(maxWidth, f)
	scaledFixed := float32(0)

	for i, r := range runes {
		adv, ok := face.GlyphAdvance(r)
		if !ok {
			continue
		}
		scaledFixed += float32(adv) / 64.0
		if scaledFixed > scaledMax {
			if i == 0 {
				return 1 // always fit at least one character
			}
			return i
		}
	}
	return len(runes)
}

// MeasureStringAdvance returns the pixel width of text using the font.Face
// advance metrics. It is the low-level primitive underlying CharsFitInWidth and
// CalcTextWidth. Tab characters are treated as zero-width.
func MeasureStringAdvance(text string, f style.Font) float32 {
	face := faceForStyle(f)
	return measureLine(text, face)
}

// TabStopPositions returns the positions (in pixels) of n evenly spaced tab
// stops given a tabOffset (position of the first stop) and tabSize (spacing).
//
// This is a convenience helper — the C# equivalent is StringFormat.GetTabStops.
func TabStopPositions(tabOffset, tabSize float32, n int) []float32 {
	if n <= 0 || tabSize <= 0 {
		return nil
	}
	stops := make([]float32, n)
	for i := range stops {
		stops[i] = tabOffset + float32(i)*tabSize
	}
	return stops
}

// FontLineHeight returns the pixel line height for the given font.
// Exposed so callers outside the utils package can obtain consistent values
// without importing internal helpers.
func FontLineHeight(f style.Font) float32 {
	face := faceForStyle(f)
	return lineHeight(face, f)
}

// MeasureStringSize returns (width, height) for a single line of text
// (no word-wrap). Height is exactly one line height.
// This is the simplest form matching GDI+ Graphics.MeasureString with no
// bounding rectangle.
func MeasureStringSize(text string, f style.Font) (width, height float32) {
	face := faceForStyle(f)
	lh := lineHeight(face, f)
	w := font.MeasureString(face, text)
	return float32(w) / 64.0, lh
}
