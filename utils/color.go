package utils

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// Predefined colors for common use.
var (
	// ColorTransparent is a fully transparent black.
	ColorTransparent = color.RGBA{R: 0, G: 0, B: 0, A: 0}
	// ColorBlack is fully opaque black.
	ColorBlack = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	// ColorWhite is fully opaque white.
	ColorWhite = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	// ColorRed is fully opaque red.
	ColorRed = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	// ColorGreen is fully opaque green (HTML green, not lime).
	ColorGreen = color.RGBA{R: 0, G: 128, B: 0, A: 255}
	// ColorBlue is fully opaque blue.
	ColorBlue = color.RGBA{R: 0, G: 0, B: 255, A: 255}
)

// ParseColor parses a color from various string formats:
//
//   - "#RGB"      — 3-digit shorthand; each digit is doubled; alpha = 0xFF.
//   - "#RRGGBB"   — 6-digit RGB; alpha = 0xFF.
//   - "#AARRGGBB" — 8-digit ARGB; the first two hex digits are alpha.
//   - A decimal integer string representing a signed 32-bit ARGB value
//     (compatible with .NET's Color.ToArgb()).
//
// Returns an error when the string cannot be recognised as any of those formats.
func ParseColor(s string) (color.RGBA, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return color.RGBA{}, fmt.Errorf("utils.ParseColor: empty string")
	}

	if strings.HasPrefix(s, "#") {
		hex := s[1:]
		switch len(hex) {
		case 3:
			// "#RGB" → expand each nibble to a byte, alpha = 0xFF
			rr := string([]byte{hex[0], hex[0]})
			gg := string([]byte{hex[1], hex[1]})
			bb := string([]byte{hex[2], hex[2]})
			hex = "FF" + rr + gg + bb
		case 6:
			// "#RRGGBB" → prepend full alpha
			hex = "FF" + hex
		case 8:
			// "#AARRGGBB" — already complete
		default:
			return color.RGBA{}, fmt.Errorf("utils.ParseColor: invalid hex length in %q", s)
		}

		v, err := strconv.ParseUint(hex, 16, 32)
		if err != nil {
			return color.RGBA{}, fmt.Errorf("utils.ParseColor: invalid hex value %q: %w", s, err)
		}
		// hex is AARRGGBB
		return color.RGBA{
			A: uint8(v >> 24),
			R: uint8(v >> 16),
			G: uint8(v >> 8),
			B: uint8(v),
		}, nil
	}

	// Try decimal ARGB integer (possibly negative, as .NET Color.ToArgb() returns int32).
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		argb := uint32(n) // reinterpret sign bit as alpha bit
		return color.RGBA{
			A: uint8(argb >> 24),
			R: uint8(argb >> 16),
			G: uint8(argb >> 8),
			B: uint8(argb),
		}, nil
	}

	return color.RGBA{}, fmt.Errorf("utils.ParseColor: unrecognised color format %q", s)
}

// FormatColor formats c as an "#AARRGGBB" uppercase hex string, which is the
// canonical FRX serialisation format.
func FormatColor(c color.RGBA) string {
	return fmt.Sprintf("#%02X%02X%02X%02X", c.A, c.R, c.G, c.B)
}

// ColorEqual reports whether a and b represent the same color.
func ColorEqual(a, b color.RGBA) bool {
	return a == b
}
