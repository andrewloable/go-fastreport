package export

import (
	"fmt"
	"math"
	"strings"
)

// ── Page dimension helpers ─────────────────────────────────────────────────────

const (
	// mmPerInch is millimetres per inch.
	mmPerInch = 25.4
	// pointsPerInch is PostScript points per inch.
	pointsPerInch = 72.0
	// pixelsPerInch is the screen/report internal DPI.
	pixelsPerInch = 96.0
)

// PixelsToMM converts pixels (at 96 dpi) to millimetres.
func PixelsToMM(px float32) float32 {
	return px / pixelsPerInch * mmPerInch
}

// MMToPixels converts millimetres to pixels (at 96 dpi).
func MMToPixels(mm float32) float32 {
	return mm / mmPerInch * pixelsPerInch
}

// PixelsToPoints converts pixels (96 dpi) to PostScript points (72 dpi).
func PixelsToPoints(px float32) float32 {
	return px / pixelsPerInch * pointsPerInch
}

// PointsToPixels converts PostScript points to pixels (96 dpi).
func PointsToPixels(pt float32) float32 {
	return pt / pointsPerInch * pixelsPerInch
}

// PixelsToInches converts pixels (96 dpi) to inches.
func PixelsToInches(px float32) float32 {
	return px / pixelsPerInch
}

// InchesToPixels converts inches to pixels (96 dpi).
func InchesToPixels(in float32) float32 {
	return in * pixelsPerInch
}

// ── Numeric helpers ────────────────────────────────────────────────────────────

// Round rounds v to the given number of decimal places.
func Round(v float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	return math.Round(v*pow) / pow
}

// ── HTML/XML encoding helpers ──────────────────────────────────────────────────

// HTMLString escapes s for safe inclusion in HTML text content.
// It replaces &, <, >, " and non-breaking space (0xa0).
func HTMLString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\u00a0':
			b.WriteString("&nbsp;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// XMLString escapes s for safe inclusion in XML text content.
// It escapes &, <, > and the control characters CR/LF/TAB.
func XMLString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '\r':
			b.WriteString("&#xD;")
		case '\n':
			b.WriteString("&#xA;")
		case '\t':
			b.WriteString("&#x9;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ── Colour helpers ─────────────────────────────────────────────────────────────

// RGBToHTMLColor returns a CSS hex colour string like "#RRGGBB".
func RGBToHTMLColor(r, g, b uint8) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// HTMLColorToRGB parses a CSS hex colour string (#RGB or #RRGGBB) and returns
// the r, g, b components. Returns (0,0,0) and false on parse error.
func HTMLColorToRGB(s string) (r, g, b uint8, ok bool) {
	s = strings.TrimPrefix(s, "#")
	switch len(s) {
	case 3:
		var v uint32
		_, err := fmt.Sscanf(s, "%1x%1x%1x", &v, &v, &v)
		if err != nil {
			return
		}
		// Expand short form.
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
		fallthrough
	case 6:
		var rv, gv, bv uint32
		_, err := fmt.Sscanf(s, "%02x%02x%02x", &rv, &gv, &bv)
		if err != nil {
			return
		}
		return uint8(rv), uint8(gv), uint8(bv), true
	}
	return
}

// ── Excel cell reference helpers ───────────────────────────────────────────────

// ExcelColName converts a zero-based column index to an Excel column letter(s).
// 0→"A", 25→"Z", 26→"AA", etc.
func ExcelColName(col int) string {
	result := ""
	col++ // make 1-based
	for col > 0 {
		col--
		result = string(rune('A'+col%26)) + result
		col /= 26
	}
	return result
}

// ExcelCellRef returns an Excel cell reference like "A1" for the given
// zero-based column and 1-based row.
func ExcelCellRef(col, row int) string {
	return fmt.Sprintf("%s%d", ExcelColName(col), row)
}

// ── Format helpers ─────────────────────────────────────────────────────────────

// FormatFloat formats v with the given number of decimal places, stripping
// trailing zeros when stripZeros is true.
func FormatFloat(v float64, decimals int, stripZeros bool) string {
	s := fmt.Sprintf("%.*f", decimals, v)
	if stripZeros && strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}
