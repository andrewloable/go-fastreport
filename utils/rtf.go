package utils

import "strings"

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
			// Check for destination group: {\* ...}
			if i < n && rtf[i] == '\\' && i+1 < n && rtf[i+1] == '*' {
				skipDepth = groupDepth
				i += 2 // consume \*
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
