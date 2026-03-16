// Package rtf implements an RTF export filter for go-fastreport.
// It renders prepared pages as an RTF document with absolutely-positioned
// text frames (WYSIWYG mode), using twip units (1 inch = 1440 twips).
package rtf

import (
	"fmt"
	"io"
	"strings"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

const (
	// twipsPerInch is the number of twips per inch.
	twipsPerInch = 1440.0
	// pixelsPerInch is the internal report DPI.
	pixelsPerInch = 96.0
)

// pixelsToTwips converts pixels (96 dpi) to RTF twips.
func pixelsToTwips(px float32) int {
	return int(px / pixelsPerInch * twipsPerInch)
}

// halfPoints converts font size in points to RTF half-points (used by \fsN).
func halfPoints(pt float32) int {
	if pt <= 0 {
		pt = 12
	}
	return int(pt * 2)
}

// Exporter produces RTF output from a PreparedPages collection.
//
// Each page is emitted as a section. Text objects become absolutely-positioned
// text frames (\absxN\absyN\abswN\abshN). Non-text objects (images, shapes,
// lines) are silently skipped — RTF is a text-oriented format.
type Exporter struct {
	export.ExportBase

	// Title is stored in the RTF document info block.
	Title string

	w      io.Writer
	pp     *preview.PreparedPages
	sb     strings.Builder
	colors []colorEntry
	fonts  []string
}

type colorEntry struct {
	r, g, b uint8
}

// NewExporter creates an Exporter with sensible defaults.
func NewExporter() *Exporter {
	return &Exporter{
		ExportBase: export.NewExportBase(),
		Title:      "Report",
	}
}

// Export writes the PreparedPages as an RTF document to w.
func (e *Exporter) Export(pp *preview.PreparedPages, w io.Writer) error {
	e.w = w
	e.pp = pp
	return e.ExportBase.Export(pp, w, e)
}

// FileExtension returns the file extension for RTF files.
func (e *Exporter) FileExtension() string { return ".rtf" }

// Name returns the human-readable name of this exporter.
func (e *Exporter) Name() string { return "RTF" }

// ── Exporter interface ─────────────────────────────────────────────────────────

func (e *Exporter) Start() error {
	e.sb.Reset()
	e.colors = nil
	e.fonts = nil

	// Register default fonts (index 0 = Arial, index 1 = Times New Roman).
	e.fonts = append(e.fonts, "Arial", "Times New Roman")

	// Register a default black color (index 1; RTF color table is 1-based).
	e.colors = append(e.colors, colorEntry{0, 0, 0})

	// RTF header — font table, color table, and document info are written in Finish
	// after we have collected all fonts/colors from the page content.
	// We use a two-pass approach: collect in Start/ExportBand, then write everything
	// in Finish. However, since RTF headers must appear first, we buffer all content
	// and prepend the header in Finish.

	return nil
}

func (e *Exporter) ExportPageBegin(pg *preview.PreparedPage) error {
	// Page dimensions in twips.
	pw := pixelsToTwips(pg.Width)
	ph := pixelsToTwips(pg.Height)

	// Use minimal margins (720 twips = 0.5 inch) so the full page is available.
	margin := 720
	e.sb.WriteString(fmt.Sprintf(
		`\paperw%d\paperh%d\margl%d\margr%d\margt%d\margb%d`+"\n",
		pw, ph, margin, margin, margin, margin,
	))
	return nil
}

func (e *Exporter) ExportBand(b *preview.PreparedBand) error {
	bandTop := b.Top
	for _, obj := range b.Objects {
		absTop := obj.Top + bandTop
		e.renderObject(obj, absTop)
	}
	return nil
}

func (e *Exporter) ExportPageEnd(pg *preview.PreparedPage) error {
	// Emit a page break after each page (except we omit on last — handled in Finish).
	e.sb.WriteString(`\page` + "\n")
	return nil
}

func (e *Exporter) Finish() error {
	// Build the final RTF document by prepending the header to the buffered body.
	var doc strings.Builder

	// RTF signature and character set.
	doc.WriteString(`{\rtf1\ansi\ansicpg1252\deff0` + "\n")

	// Font table.
	doc.WriteString(`{\fonttbl`)
	for i, name := range e.fonts {
		doc.WriteString(fmt.Sprintf(`{\f%d %s;}`, i, name))
	}
	doc.WriteString("}\n")

	// Color table (entry 0 is the implicit auto color; entries start at 1).
	doc.WriteString(`{\colortbl ;`)
	for _, c := range e.colors {
		doc.WriteString(fmt.Sprintf(`\red%d\green%d\blue%d;`, c.r, c.g, c.b))
	}
	doc.WriteString("}\n")

	// Document info.
	if e.Title != "" {
		doc.WriteString(fmt.Sprintf(`{\info{\title %s}}`, rtfEscape(e.Title)) + "\n")
	}

	// Body content collected during Export.
	body := e.sb.String()
	// Remove trailing \page (last page does not need a page break).
	body = strings.TrimSuffix(strings.TrimRight(body, "\n"), `\page`)
	body = strings.TrimRight(body, "\n")

	doc.WriteString(body)
	doc.WriteString("\n}")

	_, err := io.WriteString(e.w, doc.String())
	return err
}

// renderObject emits an RTF text frame for a single PreparedObject.
// absTop is the object's absolute Y coordinate on the page (band.Top + obj.Top).
func (e *Exporter) renderObject(obj preview.PreparedObject, absTop float32) {
	switch obj.Kind {
	case preview.ObjectTypeText, preview.ObjectTypeHtml, preview.ObjectTypeRTF:
		e.renderTextObject(obj, absTop)
	default:
		// Non-text objects (pictures, shapes, lines, barcodes, etc.) are skipped.
		// RTF is primarily a text format; image embedding requires complex WMF/EMF
		// blobs that are out of scope for this implementation.
	}
}

// renderTextObject writes a positioned RTF text frame paragraph.
func (e *Exporter) renderTextObject(obj preview.PreparedObject, absTop float32) {
	// Convert geometry to twips.
	x := pixelsToTwips(obj.Left)
	y := pixelsToTwips(absTop)
	w := pixelsToTwips(obj.Width)
	h := pixelsToTwips(obj.Height)

	// Resolve font index.
	fontIdx := e.resolveFontIndex(obj.Font.Name)
	fs := halfPoints(obj.Font.Size)

	// Resolve text color index (1-based in RTF).
	colorIdx := e.resolveColorIndex(obj.TextColor.R, obj.TextColor.G, obj.TextColor.B)

	// Horizontal alignment.
	var align string
	switch obj.HorzAlign {
	case 1:
		align = `\qc`
	case 2:
		align = `\qr`
	case 3:
		align = `\qj`
	default:
		align = `\ql`
	}

	// Font style flags.
	var fontFlags strings.Builder
	if obj.Font.Style&style.FontStyleBold != 0 {
		fontFlags.WriteString(`\b`)
	}
	if obj.Font.Style&style.FontStyleItalic != 0 {
		fontFlags.WriteString(`\i`)
	}
	if obj.Font.Style&style.FontStyleUnderline != 0 {
		fontFlags.WriteString(`\ul`)
	}
	if obj.Font.Style&style.FontStyleStrikeout != 0 {
		fontFlags.WriteString(`\strike`)
	}

	// Text content — strip RTF control words if the source is RTF, otherwise escape.
	var text string
	switch obj.Kind {
	case preview.ObjectTypeRTF:
		// Strip RTF markup and use plain text.
		text = rtfStripControlWords(obj.Text)
	default:
		text = rtfEscape(obj.Text)
	}

	// Emit the text frame as an absolutely-positioned paragraph.
	// \absxN\absyN position the frame; \abswN\abshN set the frame size.
	// \dxfrtext sets internal padding (0 = none).
	e.sb.WriteString(fmt.Sprintf(
		`{\pard\plain%s\f%d\fs%d\cf%d%s`+
			`\absw%d\absh%d\absx%d\absy%d\dxfrtext0 %s\par}`+"\n",
		align, fontIdx, fs, colorIdx, fontFlags.String(),
		w, h, x, y,
		text,
	))
}

// resolveFontIndex returns the index of fontName in the font table, adding it
// if not already present. Returns 0 (Arial) if fontName is empty.
func (e *Exporter) resolveFontIndex(fontName string) int {
	if fontName == "" {
		return 0
	}
	for i, f := range e.fonts {
		if f == fontName {
			return i
		}
	}
	idx := len(e.fonts)
	e.fonts = append(e.fonts, fontName)
	return idx
}

// resolveColorIndex returns the 1-based RTF color table index for the given
// RGB values, adding a new entry if not already present.
func (e *Exporter) resolveColorIndex(r, g, b uint8) int {
	for i, c := range e.colors {
		if c.r == r && c.g == g && c.b == b {
			return i + 1
		}
	}
	e.colors = append(e.colors, colorEntry{r, g, b})
	return len(e.colors) // 1-based
}

// rtfEscape escapes a plain text string for inclusion in RTF output.
// RTF requires that \, {, and } are escaped, and that non-ASCII characters
// are encoded as Unicode escapes (\uN?).
func rtfEscape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\\':
			b.WriteString(`\\`)
		case r == '{':
			b.WriteString(`\{`)
		case r == '}':
			b.WriteString(`\}`)
		case r == '\n':
			b.WriteString(`\line `)
		case r == '\r':
			// skip bare CR
		case r > 127:
			// Encode as RTF Unicode escape: \uN? where N is the signed UTF-16 code unit.
			// The ? is the fallback ASCII character displayed by RTF readers that do not
			// support Unicode (we use '?' as a safe placeholder).
			n := int16(r) //nolint:gosec // intentional signed cast for RTF \uN format
			b.WriteString(fmt.Sprintf(`\u%d?`, n))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// rtfStripControlWords removes RTF control sequences from an RTF string,
// returning the best-effort plain text content.
func rtfStripControlWords(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		ch := s[i]
		switch {
		case ch == '{' || ch == '}':
			i++
		case ch == '\\':
			i++
			if i >= len(s) {
				break
			}
			next := s[i]
			if next == '\\' || next == '{' || next == '}' {
				b.WriteByte(next)
				i++
			} else if next == '\'' {
				// Hex-encoded byte: \'XX
				i++
				if i+2 <= len(s) {
					var v byte
					fmt.Sscanf(s[i:i+2], "%02x", &v)
					if v > 0 {
						b.WriteByte(v)
					}
					i += 2
				}
			} else if next == '\n' || next == '\r' {
				i++
			} else {
				// Skip control word / control symbol.
				for i < len(s) && (isAlpha(s[i]) || s[i] == '-') {
					i++
				}
				// Optional numeric parameter.
				for i < len(s) && (s[i] >= '0' && s[i] <= '9') {
					i++
				}
				// Optional trailing space (delimiter).
				if i < len(s) && s[i] == ' ' {
					i++
				}
			}
		default:
			b.WriteByte(ch)
			i++
		}
	}
	return rtfEscape(b.String())
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
