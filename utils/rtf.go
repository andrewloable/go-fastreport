package utils

import (
	"fmt"
	"strings"
)

// RTFToHTML converts RTF-formatted text to an HTML fragment preserving basic
// formatting: bold (\b), italic (\i), underline (\ul / \ulnone), paragraph
// breaks (\par), line breaks (\line), tabs (\tab), font size (\fsN), and
// special characters (\'XX hex escapes, \uN Unicode, named dashes/quotes).
//
// The output is a minimal HTML fragment (no <html> or <body> wrapper).
// Unknown control words are silently discarded.
func RTFToHTML(rtf string) string {
	if !strings.HasPrefix(strings.TrimSpace(rtf), `{\rtf`) {
		// Not RTF — return HTML-escaped plain text.
		return htmlEscapeString(rtf)
	}

	var sb strings.Builder
	sb.Grow(len(rtf))

	// Formatting state stack (one entry per group level).
	type fmtState struct {
		bold      bool
		italic    bool
		underline bool
		fontSize  int // half-points (RTF \fsN unit)
	}

	var stack []fmtState
	cur := fmtState{}

	// Emit closing tags for the difference between prev and next state.
	closeTags := func(prev fmtState) {
		if prev.underline && !cur.underline {
			sb.WriteString("</u>")
		}
		if prev.italic && !cur.italic {
			sb.WriteString("</i>")
		}
		if prev.bold && !cur.bold {
			sb.WriteString("</b>")
		}
	}
	openTags := func(prev fmtState) {
		if !prev.bold && cur.bold {
			sb.WriteString("<b>")
		}
		if !prev.italic && cur.italic {
			sb.WriteString("<i>")
		}
		if !prev.underline && cur.underline {
			sb.WriteString("<u>")
		}
	}

	i := 0
	n := len(rtf)
	groupDepth := 0
	skipDepth := 0

	for i < n {
		ch := rtf[i]
		switch ch {
		case '{':
			groupDepth++
			stack = append(stack, cur) // push
			i++
			if i < n && rtf[i] == '\\' && i+1 < n && rtf[i+1] == '*' {
				skipDepth = groupDepth
				i += 2
			} else if i < n && rtf[i] == '\\' {
				// Skip plain header destination groups such as
				// {\fonttbl ...}, {\colortbl ...}, {\stylesheet ...}, etc.
				if isRTFHeaderDestination(peekControlWordName(rtf, i)) {
					skipDepth = groupDepth
				}
			}
		case '}':
			if skipDepth == groupDepth {
				skipDepth = 0
			}
			if len(stack) > 0 {
				prev := cur
				cur = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				// Emit closing/opening tags for state change.
				closeTags(prev)
				// Restore opens if state changed back.
				// (tags already open in parent context — don't re-open them)
				_ = openTags
			}
			groupDepth--
			i++
		case '\\':
			if skipDepth > 0 {
				i = skipControlWord(rtf, i)
				continue
			}
			// Parse control word.
			i++ // skip backslash
			if i >= n {
				break
			}
			bch := rtf[i]
			if !isAlpha(bch) {
				// Control symbol.
				switch bch {
				case '\'':
					if i+2 < n {
						hi := hexVal(rtf[i+1])
						lo := hexVal(rtf[i+2])
						if hi >= 0 && lo >= 0 {
							b := byte(hi<<4 | lo)
							if b < 0x80 {
								sb.WriteByte(b)
							} else {
								fmt.Fprintf(&sb, "&#%d;", int(b))
							}
						}
						i += 3
					} else {
						i++
					}
				case '-', '~':
					sb.WriteString("&#45;")
					i++
				case '_':
					sb.WriteString("&#8209;") // non-breaking hyphen
					i++
				default:
					i++
				}
				continue
			}
			// Control word.
			start := i
			for i < n && isAlpha(rtf[i]) {
				i++
			}
			word := rtf[start:i]
			// Optional numeric param.
			paramStart := i
			negative := false
			if i < n && rtf[i] == '-' {
				negative = true
				i++
			}
			for i < n && rtf[i] >= '0' && rtf[i] <= '9' {
				i++
			}
			paramStr := rtf[paramStart:i]
			if i < n && rtf[i] == ' ' {
				i++ // delimiter
			}
			paramVal := parseSignedInt(paramStr)
			if negative {
				paramVal = -paramVal
			}

			prev := cur
			switch word {
			case "b":
				cur.bold = paramVal != 0 || paramStr == ""
			case "i":
				cur.italic = paramVal != 0 || paramStr == ""
			case "ul":
				cur.underline = paramVal != 0 || paramStr == ""
			case "ulnone":
				cur.underline = false
			case "fs":
				cur.fontSize = paramVal
			case "pard":
				cur.bold, cur.italic, cur.underline = false, false, false
				closeTags(prev)
				continue
			case "par":
				closeTags(prev)
				sb.WriteString("<br>\n")
				openTags(prev)
				continue
			case "line":
				sb.WriteString("<br>\n")
				continue
			case "tab":
				sb.WriteString("&nbsp;&nbsp;&nbsp;&nbsp;")
				continue
			case "u":
				cp := parseSignedInt(paramStr)
				if cp < 0 {
					cp += 65536
				}
				if cp > 0 {
					fmt.Fprintf(&sb, "&#%d;", cp)
				}
				continue
			case "enspace", "emspace", "qmspace":
				sb.WriteString("&nbsp;")
				continue
			case "endash":
				sb.WriteString("&ndash;")
				continue
			case "emdash":
				sb.WriteString("&mdash;")
				continue
			case "lquote":
				sb.WriteString("&lsquo;")
				continue
			case "rquote":
				sb.WriteString("&rsquo;")
				continue
			case "ldblquote":
				sb.WriteString("&ldquo;")
				continue
			case "rdblquote":
				sb.WriteString("&rdquo;")
				continue
			case "bullet":
				sb.WriteString("&bull;")
				continue
			}
			// Apply formatting change.
			closeTags(prev)
			openTags(prev)
		default:
			if skipDepth == 0 {
				if ch != '\n' && ch != '\r' {
					switch ch {
					case '<':
						sb.WriteString("&lt;")
					case '>':
						sb.WriteString("&gt;")
					case '&':
						sb.WriteString("&amp;")
					default:
						sb.WriteByte(ch)
					}
				}
			}
			i++
		}
	}

	// Close any still-open formatting tags.
	if cur.underline {
		sb.WriteString("</u>")
	}
	if cur.italic {
		sb.WriteString("</i>")
	}
	if cur.bold {
		sb.WriteString("</b>")
	}

	return sb.String()
}

// htmlEscapeString replaces <, >, & with HTML entities.
func htmlEscapeString(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// StripRTF converts RTF-formatted text to plain text by removing all RTF
// control words, control symbols, and group delimiters.  It preserves the
// visible text content and converts \par / \line to newlines.
//
// It handles the basic subset used by FastReport RichObject:
//   - Nested groups { … }
//   - Control words: \rtf1, \ansi, \b, \i, \u, \par, \pard, \line, \tab, etc.
//   - Unicode escapes: \uN? (N = UTF-16 code point)
//   - Hex escapes: \'XX
//   - Destination groups: {\*\keyword …} are skipped entirely
//
// The caller should evaluate bracket expressions AFTER stripping RTF, so that
// the expression evaluator only sees plain text.
func StripRTF(rtf string) string {
	if !strings.HasPrefix(strings.TrimSpace(rtf), `{\rtf`) {
		// Not RTF — return as-is so plain text passes through unchanged.
		return rtf
	}

	var sb strings.Builder
	sb.Grow(len(rtf))

	i := 0
	n := len(rtf)
	// groupDepth tracks brace nesting. skipDepth > 0 means we are inside a
	// destination group ({\*\keyword …}) that should be discarded entirely.
	groupDepth := 0
	skipDepth := 0

	for i < n {
		ch := rtf[i]

		switch ch {
		case '{':
			groupDepth++
			i++
			// Check for destination group: {\* ...} (starred destination)
			if i < n && rtf[i] == '\\' && i+1 < n && rtf[i+1] == '*' {
				skipDepth = groupDepth
				i += 2 // consume \*
			} else if i < n && rtf[i] == '\\' {
				// Check for plain (non-starred) header destination groups such as
				// {\fonttbl ...}, {\colortbl ...}, {\stylesheet ...}, etc.
				if isRTFHeaderDestination(peekControlWordName(rtf, i)) {
					skipDepth = groupDepth
				}
			}

		case '}':
			if skipDepth == groupDepth {
				skipDepth = 0
			}
			groupDepth--
			i++

		case '\\':
			if skipDepth > 0 {
				i = skipControlWord(rtf, i)
				continue
			}
			i = processControlWord(rtf, i, &sb)

		default:
			if skipDepth == 0 {
				if ch == '\n' || ch == '\r' {
					// Bare newlines in RTF are ignored (paragraph breaks come from \par).
				} else {
					sb.WriteByte(ch)
				}
			}
			i++
		}
	}

	return sb.String()
}

// processControlWord handles a backslash sequence starting at position i.
// It writes any visible output to sb and returns the new position.
func processControlWord(rtf string, i int, sb *strings.Builder) int {
	n := len(rtf)
	i++ // skip leading backslash

	if i >= n {
		return i
	}

	ch := rtf[i]

	// Control symbol (single non-alpha character after backslash)
	if !isAlpha(ch) {
		switch ch {
		case '\'':
			// Hex escape \'XX
			if i+2 < n {
				hi := hexVal(rtf[i+1])
				lo := hexVal(rtf[i+2])
				if hi >= 0 && lo >= 0 {
					sb.WriteByte(byte(hi<<4 | lo))
				}
				i += 3
			} else {
				i++
			}
		case '-', '~':
			sb.WriteByte('-')
			i++
		case '_':
			sb.WriteByte('-') // non-breaking hyphen
			i++
		case '|', ':':
			i++ // index entry symbols — discard
		case '*':
			i++ // destination marker (already handled at '{' level)
		default:
			i++ // discard other control symbols
		}
		return i
	}

	// Control word: read the alpha word name.
	start := i
	for i < n && isAlpha(rtf[i]) {
		i++
	}
	word := rtf[start:i]

	// Optional numeric parameter (may be negative).
	paramStart := i
	if i < n && (rtf[i] == '-' || (rtf[i] >= '0' && rtf[i] <= '9')) {
		if rtf[i] == '-' {
			i++
		}
		for i < n && rtf[i] >= '0' && rtf[i] <= '9' {
			i++
		}
	}
	param := rtf[paramStart:i]

	// Skip optional trailing space (delimiter).
	if i < n && rtf[i] == ' ' {
		i++
	}

	// Handle known control words.
	switch word {
	case "par", "pard":
		sb.WriteByte('\n')
	case "line":
		sb.WriteByte('\n')
	case "tab":
		sb.WriteByte('\t')
	case "u":
		// Unicode: \uN? where N is a signed decimal UTF-16 code point.
		// The '?' after the parameter is the ANSI fallback character and is
		// already consumed as part of the stream — skip it.
		cp := parseSignedInt(param)
		if cp < 0 {
			cp += 65536
		}
		if cp > 0 {
			writeRune(sb, rune(cp))
		}
	case "uc":
		// Number of ANSI chars following a \u escape — we ignore the fallback.
	case "enspace", "emspace", "qmspace":
		sb.WriteByte(' ')
	case "endash":
		sb.WriteString("\u2013")
	case "emdash":
		sb.WriteString("\u2014")
	case "lquote":
		sb.WriteString("\u2018")
	case "rquote":
		sb.WriteString("\u2019")
	case "ldblquote":
		sb.WriteString("\u201C")
	case "rdblquote":
		sb.WriteString("\u201D")
	case "bullet":
		sb.WriteString("\u2022")
	default:
		_ = param // suppress unused
	}

	return i
}

// skipControlWord advances past a backslash sequence without producing output.
func skipControlWord(rtf string, i int) int {
	n := len(rtf)
	i++ // skip backslash
	if i >= n {
		return i
	}
	if !isAlpha(rtf[i]) {
		i++ // control symbol
		return i
	}
	// control word name
	for i < n && isAlpha(rtf[i]) {
		i++
	}
	// optional numeric param
	if i < n && (rtf[i] == '-' || (rtf[i] >= '0' && rtf[i] <= '9')) {
		if rtf[i] == '-' {
			i++
		}
		for i < n && rtf[i] >= '0' && rtf[i] <= '9' {
			i++
		}
	}
	// optional trailing space
	if i < n && rtf[i] == ' ' {
		i++
	}
	return i
}

// isRTFHeaderDestination returns true for RTF destination group keywords whose
// entire content should be skipped (not emitted as visible text).  These groups
// carry metadata such as font tables, colour tables, style sheets, etc.
func isRTFHeaderDestination(name string) bool {
	switch name {
	case "fonttbl", "colortbl", "stylesheet", "info",
		"header", "footer", "headerl", "headerr", "headerf",
		"footerl", "footerr", "footerf",
		"pict", "object", "fldinst",
		"private", "revtbl",
		"listtable", "listoverridetable", "rsidtbl",
		"generator", "pgdsctbl", "ftnsep", "ftnsepc",
		"aftnsep", "aftnsepc", "latentstyles":
		return true
	}
	return false
}

// peekControlWordName returns the control-word name that immediately follows
// position i in rtf (i must point at the backslash) without advancing i.
// Returns "" if there is no control word at i.
func peekControlWordName(rtf string, i int) string {
	n := len(rtf)
	if i >= n || rtf[i] != '\\' {
		return ""
	}
	i++ // skip backslash
	if i >= n || !isAlpha(rtf[i]) {
		return ""
	}
	start := i
	for i < n && isAlpha(rtf[i]) {
		i++
	}
	return rtf[start:i]
}

// ── helpers ───────────────────────────────────────────────────────────────────

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func hexVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

func parseSignedInt(s string) int {
	if s == "" {
		return 0
	}
	neg := false
	if s[0] == '-' {
		neg = true
		s = s[1:]
	}
	v := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		v = v*10 + int(c-'0')
	}
	if neg {
		v = -v
	}
	return v
}

func writeRune(sb *strings.Builder, r rune) {
	buf := make([]byte, 4)
	n := encodeRuneUTF8(buf, r)
	sb.Write(buf[:n])
}

// encodeRuneUTF8 encodes r into buf (UTF-8) and returns the number of bytes written.
func encodeRuneUTF8(p []byte, r rune) int {
	switch {
	case r < 0x80:
		p[0] = byte(r)
		return 1
	case r < 0x800:
		p[0] = byte(0xC0 | (r >> 6))
		p[1] = byte(0x80 | (r & 0x3F))
		return 2
	case r < 0x10000:
		p[0] = byte(0xE0 | (r >> 12))
		p[1] = byte(0x80 | ((r >> 6) & 0x3F))
		p[2] = byte(0x80 | (r & 0x3F))
		return 3
	default:
		p[0] = byte(0xF0 | (r >> 18))
		p[1] = byte(0x80 | ((r >> 12) & 0x3F))
		p[2] = byte(0x80 | ((r >> 6) & 0x3F))
		p[3] = byte(0x80 | (r & 0x3F))
		return 4
	}
}
