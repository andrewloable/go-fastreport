package export

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/andrewloable/go-fastreport/style"
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

// ── Page dimension helpers (report-aware) ──────────────────────────────────────

// ReportPageDims is a minimal interface that both *reportpkg.ReportPage and
// any test stub must satisfy. It avoids an import cycle between export and reportpkg.
type ReportPageDims interface {
	GetPaperWidth() float32
	GetPaperHeight() float32
	IsUnlimitedHeight() bool
}

// GetPageWidth returns the printable width of page in millimetres.
// Matches C# ExportUtils.GetPageWidth (UnlimitedWidth is not exposed in the Go port).
func GetPageWidth(page ReportPageDims) float32 {
	return page.GetPaperWidth()
}

// GetPageHeight returns the printable height of page in millimetres.
// When IsUnlimitedHeight() is true the page has no fixed height; PaperHeight
// holds the accumulated content height set by the engine.
// Matches C# ExportUtils.GetPageHeight.
func GetPageHeight(page ReportPageDims) float32 {
	return page.GetPaperHeight()
}

// ── FloatToString ───────────────────────────────────────────────────────────────

// FloatToString rounds value to digits decimal places and formats it with a
// dot decimal separator (invariant culture), matching C# ExportUtils.FloatToString.
func FloatToString(value float64, digits int) string {
	rounded := Round(value, digits)
	s := fmt.Sprintf("%.*f", digits, rounded)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

// ── HTMLColor ───────────────────────────────────────────────────────────────────

// HTMLColor returns a CSS colour string for c.
// When alpha < 255 it returns an rgba() value; otherwise rgb().
// Matches C# ExportUtils.HTMLColor.
func HTMLColor(c color.RGBA) string {
	if c.A < 255 {
		alpha := fmt.Sprintf("%.2f", float64(c.A)/255.0)
		return fmt.Sprintf("rgba(%d, %d, %d, %s)", c.R, c.G, c.B, alpha)
	}
	return fmt.Sprintf("rgb(%d, %d, %d)", c.R, c.G, c.B)
}

// ── HtmlURL ─────────────────────────────────────────────────────────────────────

// HtmlURL percent-encodes characters in value that are not safe in HTML URLs.
// Matches C# ExportUtils.HtmlURL (non-DOTNET_4 branch).
func HtmlURL(value string) string {
	var b strings.Builder
	for i := 0; i < len(value); i++ {
		c := value[i]
		switch c {
		case '\\':
			b.WriteByte('/')
		case '&', '<', '>', '{', '}', ';', '?', ' ', '\'', '"':
			b.WriteString("%" + ByteToHex(c))
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// ── GetColorFromFill ────────────────────────────────────────────────────────────

// GetColorFromFill extracts a representative colour from fill.
// Matches C# ExportUtils.GetColorFromFill.
func GetColorFromFill(fill style.Fill) color.RGBA {
	if fill == nil {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	switch f := fill.(type) {
	case *style.SolidFill:
		return f.Color
	case *style.GlassFill:
		return f.Color
	case *style.HatchFill:
		return f.BackColor
	case *style.LinearGradientFill:
		return middleColor(f.StartColor, f.EndColor)
	default:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
}

func middleColor(c1, c2 color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8((int(c1.R) + int(c2.R)) / 2),
		G: uint8((int(c1.G) + int(c2.G)) / 2),
		B: uint8((int(c1.B) + int(c2.B)) / 2),
		A: 255,
	}
}

// ── GetRFCDate ──────────────────────────────────────────────────────────────────

// GetRFCDate formats datetime as an RFC 2822 / HTTP date string, replacing the
// "GMT" suffix with the local UTC offset when non-zero.
// Matches C# ExportUtils.GetRFCDate.
func GetRFCDate(datetime time.Time) string {
	s := datetime.UTC().Format(time.RFC1123)
	_, offset := datetime.Zone()
	if offset == 0 {
		return s
	}
	hours := offset / 3600
	minutes := (offset % 3600) / 60
	if minutes < 0 {
		minutes = -minutes
	}
	sign := "+"
	if hours < 0 {
		sign = "-"
		hours = -hours
	}
	offsetStr := fmt.Sprintf("%s%02d%02d", sign, hours, minutes)
	return strings.Replace(s, "UTC", offsetStr, 1)
}

// ── ByteToHex ───────────────────────────────────────────────────────────────────

const xconv = "0123456789ABCDEF"

// ByteToHex converts a single byte to a two-character uppercase hex string.
// Matches C# ExportUtils.ByteToHex.
func ByteToHex(b byte) string {
	return string([]byte{xconv[b>>4], xconv[b&0xF]})
}

// ── ReverseString ───────────────────────────────────────────────────────────────

// ReverseString returns str with its characters in reverse order.
// Matches C# ExportUtils.ReverseString.
func ReverseString(str string) string {
	runes := []rune(str)
	n := len(runes)
	result := make([]rune, n)
	for i, r := range runes {
		result[n-1-i] = r
	}
	return string(result)
}

// ── QuotedPrintable ─────────────────────────────────────────────────────────────

// QuotedPrintable encodes values using Quoted-Printable (RFC 2045).
// Lines are soft-wrapped at 73 characters with "=\r\n".
// Matches C# ExportUtils.QuotedPrintable.
func QuotedPrintable(values []byte) string {
	var b strings.Builder
	length := 0
	for _, c := range values {
		if length > 73 {
			length = 0
			b.WriteString("=\r\n")
		}
		if c < 9 || c == 61 || c > 126 {
			b.WriteByte('=')
			b.WriteByte(xconv[c>>4])
			b.WriteByte(xconv[c&0xF])
			length += 3
		} else {
			b.WriteByte(c)
			length++
		}
	}
	return b.String()
}
