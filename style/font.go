package style

import (
	"fmt"
	"strconv"
	"strings"
)

// FontStyle holds font decoration flags (mirrors System.Drawing.FontStyle).
type FontStyle int

const (
	FontStyleRegular   FontStyle = 0
	FontStyleBold      FontStyle = 1
	FontStyleItalic    FontStyle = 2
	FontStyleUnderline FontStyle = 4
	FontStyleStrikeout FontStyle = 8
)

// Font holds the font properties used by TextObject.
// It is the Go equivalent of System.Drawing.Font.
type Font struct {
	// Name is the font family name (e.g. "Arial").
	Name string
	// Size is the font size in points.
	Size float32
	// Style is the combination of FontStyle flags.
	Style FontStyle
}

// DefaultFont returns the default report font (Arial 10pt Regular).
func DefaultFont() Font {
	return Font{Name: "Arial", Size: 10, Style: FontStyleRegular}
}

// FontEqual reports whether two Font values are identical.
func FontEqual(a, b Font) bool {
	return a.Name == b.Name && a.Size == b.Size && a.Style == b.Style
}

// FontToStr serialises a Font as "Name, Size, Style" (similar to .NET Font.ToString()).
func FontToStr(f Font) string {
	return fmt.Sprintf("%s, %.4g, %d", f.Name, f.Size, f.Style)
}

// FontFromStr parses "Name, Size, Style" produced by FontToStr.
// Returns DefaultFont() on any parse error.
func FontFromStr(s string) Font {
	parts := strings.SplitN(s, ",", 3)
	if len(parts) < 2 {
		return DefaultFont()
	}
	name := strings.TrimSpace(parts[0])
	size, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 32)
	if err != nil {
		return DefaultFont()
	}
	st := FontStyleRegular
	if len(parts) == 3 {
		n, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err == nil {
			st = FontStyle(n)
		}
	}
	return Font{Name: name, Size: float32(size), Style: st}
}
