package utils

import (
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	// standardDPI is the reference DPI used for point/pixel conversions.
	standardDPI = 96.0
	// pointsPerInch is the typographic definition.
	pointsPerInch = 72.0
)

// TextMeasurement holds the results of text measurement.
type TextMeasurement struct {
	Width  float32 // in pixels
	Height float32 // in pixels
	Lines  int     // number of lines
}

// MeasureString measures the dimensions of text rendered with the given face.
// If maxWidth > 0, text is wrapped at word boundaries.
func MeasureString(text string, face font.Face, maxWidth float32) TextMeasurement {
	if text == "" {
		h := FontHeight(face)
		return TextMeasurement{Width: 0, Height: h, Lines: 1}
	}

	var lines []string
	if maxWidth > 0 {
		lines = WrapText(text, face, maxWidth)
	} else {
		lines = strings.Split(text, "\n")
	}

	lineHeight := FontHeight(face)
	var maxW float32
	for _, l := range lines {
		w, _ := MeasureLine(l, face)
		if w > maxW {
			maxW = w
		}
	}

	return TextMeasurement{
		Width:  maxW,
		Height: lineHeight * float32(len(lines)),
		Lines:  len(lines),
	}
}

// WrapText wraps text to fit within maxWidth pixels using the given font face.
// Returns a slice of lines. Hard newlines in the source are always honoured.
func WrapText(text string, face font.Face, maxWidth float32) []string {
	hardLines := strings.Split(text, "\n")
	result := make([]string, 0, len(hardLines))

	for _, hard := range hardLines {
		wrapped := wrapLine(hard, face, maxWidth)
		result = append(result, wrapped...)
	}

	return result
}

// wrapLine wraps a single (newline-free) line of text into one or more lines
// that each fit within maxWidth pixels.
func wrapLine(line string, face font.Face, maxWidth float32) []string {
	if line == "" {
		return []string{""}
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{""}
	}

	var result []string
	current := ""

	for _, word := range words {
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}

		w, _ := MeasureLine(candidate, face)
		if w <= maxWidth || current == "" {
			// Either it fits, or this is the very first word on the line (cannot
			// break a single word that is wider than maxWidth).
			current = candidate
		} else {
			result = append(result, current)
			current = word
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

// MeasureLine measures a single line of text (no wrapping).
// Returns width and height in pixels.
func MeasureLine(line string, face font.Face) (width, height float32) {
	height = FontHeight(face)
	if line == "" {
		return 0, height
	}

	bounds, advance := font.BoundString(face, line)
	_ = bounds // we use the advance for width; bounds can be used for precise bbox

	// advance is in fixed.Int26_6; convert to float32 pixels.
	width = float32(advance) / float32(fixed.Int26_6(64))
	return width, height
}

// FontHeight returns the height of the font face in pixels.
func FontHeight(face font.Face) float32 {
	metrics := face.Metrics()
	// Ascent + Descent gives the total cell height.
	total := metrics.Ascent + metrics.Descent
	return float32(total) / float32(fixed.Int26_6(64))
}

// PixelsToPoints converts pixel measurements to points (at 96 DPI).
func PixelsToPoints(pixels float32) float32 {
	return pixels * pointsPerInch / standardDPI
}

// PointsToPixels converts points to pixels (at 96 DPI).
func PointsToPixels(points float32) float32 {
	return points * standardDPI / pointsPerInch
}
