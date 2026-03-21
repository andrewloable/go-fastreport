package utils

import (
	"image/color"
	"strings"

	"github.com/andrewloable/go-fastreport/style"
)

// BaselineType indicates character vertical placement relative to the line baseline.
// Mirrors HtmlTextRenderer.BaseLine in C# (HtmlTextRenderer.cs, line 1304).
type BaselineType int

const (
	// BaselineNormal is the default baseline (no vertical shift).
	BaselineNormal BaselineType = iota
	// BaselineSubscript places the run below the normal baseline (via <sub>).
	BaselineSubscript
	// BaselineSuperscript places the run above the normal baseline (via <sup>).
	BaselineSuperscript
)

// HtmlRun is a styled run of text produced by the HTML text renderer.
// It represents a contiguous span of text with uniform styling.
type HtmlRun struct {
	// Text is the plain-text content of this run.
	Text string
	// Font is the font to use for this run.
	Font style.Font
	// Color is the text foreground color.
	Color color.RGBA
	// BackgroundColor is the text background highlight color (from CSS background-color).
	// An alpha of 0 means no background is applied.
	BackgroundColor color.RGBA
	// Underline indicates underlined text.
	Underline bool
	// Strikeout indicates struck-out text.
	Strikeout bool
	// LineBreak is true when this run ends with an explicit line break.
	LineBreak bool
	// Baseline indicates subscript/superscript placement.
	// Mirrors HtmlTextRenderer.BaseLine (HtmlTextRenderer.cs, line 1304).
	Baseline BaselineType
}

// HtmlLine is one visual line produced by the HTML text renderer.
type HtmlLine struct {
	Runs []HtmlRun
}

// HtmlTextRenderer parses inline HTML markup in text objects and produces
// styled runs for measurement and rendering. It is the Go equivalent of
// FastReport.Utils.HtmlTextRenderer.
//
// Supported tags: <b>, <i>, <u>, <s>, <strike>, <br>, <font>, <span>,
// <sub>, <sup>, plus style="color:…;font-size:…" on span/font.
// Entities: &amp; &lt; &gt; &nbsp; &quot;
type HtmlTextRenderer struct {
	baseFont  style.Font
	baseColor color.RGBA
	lines     []HtmlLine
}

// NewHtmlTextRenderer creates a renderer for the given HTML text, base font,
// and base foreground colour. Call Lines() to access the parsed output.
func NewHtmlTextRenderer(htmlText string, baseFont style.Font, baseColor color.RGBA) *HtmlTextRenderer {
	r := &HtmlTextRenderer{
		baseFont:  baseFont,
		baseColor: baseColor,
	}
	r.parse(htmlText)
	return r
}

// Lines returns the parsed visual lines.
func (r *HtmlTextRenderer) Lines() []HtmlLine { return r.lines }

// PlainText returns the plain-text content with all HTML tags removed.
func (r *HtmlTextRenderer) PlainText() string {
	var sb strings.Builder
	for i, line := range r.lines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		for _, run := range line.Runs {
			sb.WriteString(run.Text)
		}
	}
	return sb.String()
}

// MeasureHeight returns the total height in pixels needed to render the HTML
// text in a box of the given width using the base font.
func (r *HtmlTextRenderer) MeasureHeight(width float32) float32 {
	if len(r.lines) == 0 {
		return 0
	}
	face := faceForStyle(r.baseFont)
	lh := lineHeight(face, r.baseFont)
	totalLines := 0
	for _, line := range r.lines {
		// Collect the plain text of this logical line.
		var sb strings.Builder
		for _, run := range line.Runs {
			sb.WriteString(run.Text)
		}
		text := sb.String()
		if width > 0 && face != nil {
			totalLines += len(wrapLines(text, face, width))
		} else {
			totalLines++
		}
	}
	return lh * float32(totalLines)
}

// ── Parser ────────────────────────────────────────────────────────────────────

// styleState holds the current formatting context during parsing.
type styleState struct {
	font            style.Font
	clr             color.RGBA
	backgroundColor color.RGBA // CSS background-color; zero alpha = none
	bold            bool
	italic          bool
	underline       bool
	strike          bool
	// baseline is the subscript/superscript placement for <sub>/<sup> tags.
	// Mirrors HtmlTextRenderer.BaseLine (HtmlTextRenderer.cs, line 1304).
	baseline BaselineType
}

func (r *HtmlTextRenderer) parse(src string) {
	base := styleState{
		font:   r.baseFont,
		clr:    r.baseColor,
		bold:   r.baseFont.Style&style.FontStyleBold != 0,
		italic: r.baseFont.Style&style.FontStyleItalic != 0,
	}
	var stack []styleState
	cur := base

	var currentLine HtmlLine
	var currentRun strings.Builder

	flush := func() {
		text := currentRun.String()
		if text == "" {
			return
		}
		f := cur.font
		f.Style &^= style.FontStyleBold | style.FontStyleItalic | style.FontStyleUnderline | style.FontStyleStrikeout
		if cur.bold {
			f.Style |= style.FontStyleBold
		}
		if cur.italic {
			f.Style |= style.FontStyleItalic
		}
		currentLine.Runs = append(currentLine.Runs, HtmlRun{
			Text:            text,
			Font:            f,
			Color:           cur.clr,
			BackgroundColor: cur.backgroundColor,
			Underline:       cur.underline,
			Strikeout:       cur.strike,
			Baseline:        cur.baseline,
		})
		currentRun.Reset()
	}

	newLine := func() {
		flush()
		r.lines = append(r.lines, currentLine)
		currentLine = HtmlLine{}
	}

	i := 0
	for i < len(src) {
		if src[i] == '<' {
			// Parse tag.
			end := strings.IndexByte(src[i:], '>')
			if end < 0 {
				currentRun.WriteByte(src[i])
				i++
				continue
			}
			rawTag := strings.TrimSpace(src[i+1 : i+end])
			i += end + 1
			lower := strings.ToLower(rawTag)
			// Pass rawTag so attribute values (e.g. font names) preserve their case.
			r.applyTag(lower, rawTag, &stack, &cur, flush, newLine)
		} else if src[i] == '&' {
			// Parse entity.
			semi := strings.IndexByte(src[i:], ';')
			if semi > 0 && semi < 10 {
				entity := src[i+1 : i+semi]
				i += semi + 1
				switch entity {
				case "amp":
					currentRun.WriteByte('&')
				case "lt":
					currentRun.WriteByte('<')
				case "gt":
					currentRun.WriteByte('>')
				case "nbsp":
					currentRun.WriteByte(' ')
				case "quot":
					currentRun.WriteByte('"')
				default:
					currentRun.WriteString("&" + entity + ";")
				}
			} else {
				currentRun.WriteByte(src[i])
				i++
			}
		} else if src[i] == '\n' {
			newLine()
			i++
		} else {
			currentRun.WriteByte(src[i])
			i++
		}
	}
	flush()
	r.lines = append(r.lines, currentLine)
}

func (r *HtmlTextRenderer) applyTag(
	tag string, // lowercased tag for name matching
	rawTag string, // original tag preserving attribute value case
	stack *[]styleState,
	cur *styleState,
	flush func(),
	newLine func(),
) {
	flush()
	closing := strings.HasPrefix(tag, "/")
	if closing {
		tag = strings.TrimPrefix(tag, "/")
		tag = strings.Fields(tag)[0]
		// Pop the stack.
		if len(*stack) > 0 {
			*cur = (*stack)[len(*stack)-1]
			*stack = (*stack)[:len(*stack)-1]
		}
		return
	}

	// Self-closing tags.
	if tag == "br" || strings.HasPrefix(tag, "br ") || strings.HasSuffix(tag, "/") {
		newLine()
		return
	}

	// Push current state and apply tag styles.
	*stack = append(*stack, *cur)
	tagName := strings.Fields(tag)[0]
	switch tagName {
	case "b", "strong":
		cur.bold = true
	case "i", "em":
		cur.italic = true
	case "u":
		cur.underline = true
	case "s", "strike", "del":
		cur.strike = true
	case "sub":
		// Subscript: mirrors HtmlTextRenderer.cs case "sub" (line 1012).
		cur.baseline = BaselineSubscript
	case "sup":
		// Superscript: mirrors HtmlTextRenderer.cs case "sup" (line 1016).
		cur.baseline = BaselineSuperscript
	case "font", "span":
		// Parse attributes from the original (case-preserving) raw tag so that
		// font names and colors are not forced to lower-case.
		rawTagName := strings.Fields(rawTag)[0]
		attrStr := rawTag[len(rawTagName):]
		attrs := parseAttrs(attrStr)
		if c, ok := attrs["color"]; ok {
			if parsed, err := ParseColor(c); err == nil {
				cur.clr = parsed
			}
		}
		if sz, ok := attrs["size"]; ok {
			var s float32
			if _, err := parseFloat(sz, &s); err == nil && s > 0 {
				cur.font.Size = s
			}
		}
		if face, ok := attrs["face"]; ok && face != "" {
			// <font face="Arial"> sets the font family name.
			// Mirrors HtmlTextRenderer.cs font tag handling.
			cur.font.Name = strings.ReplaceAll(face, "'", "")
		}
		if styleAttr, ok := attrs["style"]; ok {
			applyInlineStyle(styleAttr, cur)
		}
	}
}

// parseAttrs extracts key="value" attribute pairs from a tag attribute string.
func parseAttrs(attrStr string) map[string]string {
	attrs := make(map[string]string)
	attrStr = strings.TrimSpace(attrStr)
	for len(attrStr) > 0 {
		// Find next key.
		eqIdx := strings.IndexByte(attrStr, '=')
		if eqIdx < 0 {
			break
		}
		key := strings.ToLower(strings.TrimSpace(attrStr[:eqIdx]))
		attrStr = strings.TrimSpace(attrStr[eqIdx+1:])
		// Find value (quoted or unquoted).
		var val string
		if len(attrStr) > 0 && (attrStr[0] == '"' || attrStr[0] == '\'') {
			quote := attrStr[0]
			end := strings.IndexByte(attrStr[1:], quote)
			if end < 0 {
				break
			}
			val = attrStr[1 : end+1]
			attrStr = strings.TrimSpace(attrStr[end+2:])
		} else {
			end := strings.IndexAny(attrStr, " \t\r\n")
			if end < 0 {
				val = attrStr
				attrStr = ""
			} else {
				val = attrStr[:end]
				attrStr = strings.TrimSpace(attrStr[end:])
			}
		}
		attrs[key] = val
	}
	return attrs
}

// applyInlineStyle applies CSS inline style properties to the current state.
// Mirrors HtmlTextRenderer.cs CssStyle() (line 574).
func applyInlineStyle(styleStr string, cur *styleState) {
	for _, decl := range strings.Split(styleStr, ";") {
		parts := strings.SplitN(decl, ":", 2)
		if len(parts) != 2 {
			continue
		}
		prop := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		switch prop {
		case "color":
			if parsed, ok := parseCSSColor(val); ok {
				cur.clr = parsed
			}
		case "background-color":
			// CSS background-color: mirrors CssStyle() "background-color" (line 670).
			if parsed, ok := parseCSSColor(val); ok {
				cur.backgroundColor = parsed
			}
		case "font-size":
			// Supports pt, px, em units. Mirrors CssStyle() "font-size" (line 609).
			var s float32
			switch {
			case strings.HasSuffix(val, "px"):
				if _, err := parseFloat(strings.TrimSuffix(val, "px"), &s); err == nil && s > 0 {
					// 0.75 = 96dpi px → pt conversion (same as C# "px" branch, line 612)
					cur.font.Size = s * 0.75
				}
			case strings.HasSuffix(val, "em"):
				if _, err := parseFloat(strings.TrimSuffix(val, "em"), &s); err == nil && s > 0 {
					// em is relative: scale the current font size (C# line 616)
					cur.font.Size = cur.font.Size * s
				}
			default:
				if _, err := parseFloat(strings.TrimSuffix(val, "pt"), &s); err == nil && s > 0 {
					cur.font.Size = s
				}
			}
		case "font-family":
			// Mirrors CssStyle() "font-family" (line 618).
			if val != "" {
				cur.font.Name = strings.ReplaceAll(val, "'", "")
			}
		case "font-weight":
			cur.bold = strings.Contains(val, "bold") || val == "700" || val == "800" || val == "900"
		case "font-style":
			cur.italic = strings.Contains(val, "italic") || strings.Contains(val, "oblique")
		case "text-decoration":
			cur.underline = strings.Contains(val, "underline")
			cur.strike = strings.Contains(val, "line-through")
		}
	}
}

// parseCSSColor parses a CSS color value string into color.RGBA.
// Supports: #hex, rgb(), rgba(), named colors, and decimal ARGB.
// Mirrors HtmlTextRenderer.cs CssStyle() "color" / "background-color" blocks (line 626).
func parseCSSColor(val string) (color.RGBA, bool) {
	val = strings.TrimSpace(val)
	// Try rgb(r, g, b) format (C# line 649).
	if strings.HasPrefix(val, "rgba(") || strings.HasPrefix(val, "rgba (") {
		inner := val[strings.Index(val, "(")+1:]
		inner = strings.TrimSuffix(inner, ")")
		parts := strings.Split(inner, ",")
		if len(parts) == 4 {
			var r, g, b, a float32
			if _, err := parseFloat(strings.TrimSpace(parts[0]), &r); err == nil {
				if _, err := parseFloat(strings.TrimSpace(parts[1]), &g); err == nil {
					if _, err := parseFloat(strings.TrimSpace(parts[2]), &b); err == nil {
						if _, err := parseFloat(strings.TrimSpace(parts[3]), &a); err == nil {
							return color.RGBA{
								R: uint8(r),
								G: uint8(g),
								B: uint8(b),
								A: uint8(a * 255),
							}, true
						}
					}
				}
			}
		}
	} else if strings.HasPrefix(val, "rgb(") || strings.HasPrefix(val, "rgb (") {
		inner := val[strings.Index(val, "(")+1:]
		inner = strings.TrimSuffix(inner, ")")
		parts := strings.Split(inner, ",")
		if len(parts) == 3 {
			var r, g, b float32
			if _, err := parseFloat(strings.TrimSpace(parts[0]), &r); err == nil {
				if _, err := parseFloat(strings.TrimSpace(parts[1]), &g); err == nil {
					if _, err := parseFloat(strings.TrimSpace(parts[2]), &b); err == nil {
						return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, true
					}
				}
			}
		}
	}
	// Fall back to ParseColor for hex, named, and decimal ARGB.
	if parsed, err := ParseColor(val); err == nil {
		return parsed, true
	}
	return color.RGBA{}, false
}

// parseFloat parses a float32 from s, writing to *out. Returns (n, nil) on success.
func parseFloat(s string, out *float32) (int, error) {
	var v float64
	var n int
	_, err := strings.NewReader(s), error(nil)
	n, err = 0, nil
	// Use fmt.Sscanf-style approach without fmt import:
	// Parse manually.
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, &parseError{s}
	}
	neg := false
	if s[0] == '-' {
		neg = true
		s = s[1:]
	}
	integer := float64(0)
	frac := float64(0)
	fracDiv := float64(1)
	dot := false
	for n < len(s) {
		c := s[n]
		if c >= '0' && c <= '9' {
			if dot {
				frac = frac*10 + float64(c-'0')
				fracDiv *= 10
			} else {
				integer = integer*10 + float64(c-'0')
			}
			n++
		} else if c == '.' && !dot {
			dot = true
			n++
		} else {
			break
		}
	}
	v = integer + frac/fracDiv
	if neg {
		v = -v
	}
	*out = float32(v)
	return n, err
}

type parseError struct{ s string }

func (e *parseError) Error() string { return "parse error: " + e.s }

// StripHtmlTags removes all HTML tags from s and decodes basic entities.
// Use this when you need plain text from an HTML-markup string.
func StripHtmlTags(s string) string {
	r := NewHtmlTextRenderer(s, style.DefaultFont(), color.RGBA{A: 255})
	return r.PlainText()
}
