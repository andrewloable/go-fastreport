package utils

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
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

	// Determine if we're using a real font or the basicfont fallback.
	isRealFont := face != basicfont.Face7x13

	// GDI+ MeasureString (GenericDefault StringFormat) adds extra horizontal
	// padding of approximately fontSize/6 on each side (total ≈ fontSize/3).
	// C# ref: MSDN "StringFormat.GenericDefault includes trailing space".
	// When using a real system font, Go's font.MeasureString returns advance
	// widths that closely match C#'s MeasureString (which already includes the
	// GDI trailing space). Only add gdiPadW for the basicfont fallback.
	gdiPadW := float32(0)
	if !isRealFont {
		// No GDI padding for basicfont; ScaleWidth handles the conversion.
	}

	// For wrapping: use the real maxWidth with real fonts, scale for basicfont.
	wrapW := maxWidth
	if !isRealFont {
		wrapW = scaleMaxWidth(maxWidth, f)
	}

	lines := wrapLines(text, face, wrapW)
	var maxW float32
	for _, line := range lines {
		w := measureLine(line, face)
		if w > maxW {
			maxW = w
		}
	}
	// Add GDI+ MeasureString horizontal padding (basicfont fallback only).
	maxW += gdiPadW
	return maxW, lineH * float32(len(lines))
}

// GDIMeasureStringPadding returns the extra height that .NET's
// Graphics.MeasureString adds with the default StringFormat (GenericDefault).
// This is fontSize_pt / 6 pixels (≡ fontPx / 8). It should be added to the
// measured height only when the C# simple MeasureString path is used (i.e.
// IsAdvancedRendererNeeded is false).
func GDIMeasureStringPadding(f style.Font) float32 {
	return f.Size / 6.0
}

// MeasureLines returns the number of visual lines produced by word-wrapping
// text at maxWidth pixels using the given font.
// If maxWidth <= 0, each \n-delimited paragraph is one line.
func MeasureLines(text string, f style.Font, maxWidth float32) int {
	if text == "" {
		return 1
	}
	if maxWidth <= 0 {
		// No wrapping — count explicit line breaks only.
		face := faceForStyle(f)
		return len(wrapLines(text, face, 0))
	}
	// Split on explicit line breaks first, then calculate word-wrap lines
	// for each paragraph. This ensures \r\n and \n are always counted.
	paragraphs := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	total := 0
	for _, para := range paragraphs {
		if para == "" {
			total++
			continue
		}
		fullW, _ := MeasureText(para, f, 0)
		if fullW <= maxWidth {
			total++
		} else {
			total += int(fullW/maxWidth) + 1
		}
	}
	if total < 1 {
		total = 1
	}
	return total
}

// scaleMaxWidth adjusts maxWidth for the basicfont fallback.
func scaleMaxWidth(maxWidth float32, f style.Font) float32 {
	if maxWidth <= 0 {
		return maxWidth
	}
	fontPx := f.Size * 96.0 / 72.0
	if fontPx <= 0 {
		return maxWidth
	}
	actualAvgWidth := fontPx * 0.48
	basicAvgWidth := float32(7.0)
	return maxWidth * (basicAvgWidth / actualAvgWidth)
}

// ScaleWidth converts a width measured by MeasureText (using the basicfont
// fallback) to the approximate width at the target font's proportions.
// This compensates for the monospace basicfont being wider than proportional
// fonts like Tahoma or Arial.
func ScaleWidth(measuredWidth float32, f style.Font) float32 {
	// When MeasureText used a real system font (not basicfont), the width is
	// already correct and no scaling is needed. Only scale when basicfont was
	// used as a fallback (monospace approximation).
	face := faceForStyle(f)
	if face != basicfont.Face7x13 {
		return measuredWidth
	}
	fontPx := f.Size * 96.0 / 72.0
	if fontPx <= 0 {
		return measuredWidth
	}
	actualAvgWidth := fontPx * 0.48
	basicAvgWidth := float32(7.0)
	return measuredWidth * (actualAvgWidth / basicAvgWidth)
}

// ── internal helpers ──────────────────────────────────────────────────────────

// faceForStyle returns a font.Face for the given style.Font.
// Uses DefaultFontManager if a matching face is registered; otherwise falls
// back to basicfont.Face7x13 scaled by the font size ratio.
func faceForStyle(f style.Font) font.Face {
	desc := FontDescriptor{
		Family: f.Name,
		Size:   f.Size,
		Style:  FontStyle(f.Style),
	}
	return DefaultFontManager.GetFace(desc)
}

// lineHeight computes the line height in pixels for a given font, matching
// C#'s font.FontFamily.GetLineSpacing(style) / GetEmHeight(style) ratio.
// This uses the font size in px (pt * 96/72) multiplied by the font-specific
// line-spacing ratio, producing the same value as .NET's MeasureString.
func lineHeight(face font.Face, f style.Font) float32 {
	// Convert point size to pixels: fontPx = pt * 96 / 72.
	fontPx := f.Size * 96.0 / 72.0
	if fontPx <= 0 {
		fontPx = style.DefaultFont().Size * 96.0 / 72.0
	}
	return fontPx * fontLineSpacingRatio(f.Name)
}

// fontLineSpacingRatio returns the exact .NET font line-spacing ratio.
// Computed from: (usWinAscent + usWinDescent + typoLineGap) / unitsPerEm
// These are the actual values from the font OS/2 and hhea tables, matching
// .NET's FontFamily.GetLineSpacing(style) / GetEmHeight(style).
// Using full precision ensures sub-pixel accuracy in shift computations.
func fontLineSpacingRatio(fontFamily string) float32 {
	switch strings.ToLower(fontFamily) {
	case "arial", "arial narrow":
		return 2355.0 / 2048.0 // 1.14990234375
	case "times new roman", "times":
		return 2355.0 / 2048.0 // 1.14990234375
	case "tahoma", "microsoft sans serif":
		return 2472.0 / 2048.0 // 1.20703125
	case "verdana":
		return 2489.0 / 2048.0 // 1.21533203125
	case "arial unicode ms":
		return 2743.0 / 2048.0 // 1.33935546875
	case "arial black":
		return 2899.0 / 2048.0 // 1.41552734375
	case "georgia":
		return 2327.0 / 2048.0 // 1.13623046875
	case "courier new", "courier":
		return 2320.0 / 2048.0 // 1.1328125
	case "segoe ui":
		return 2724.0 / 2048.0 // 1.33007812500
	default:
		return 2472.0 / 2048.0 // Tahoma default
	}
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
	// Normalize line endings before splitting.
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
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
