package utils

import (
	"strings"
	"testing"
)

func TestStripRTF_PlainText(t *testing.T) {
	result := StripRTF("Hello World")
	if result != "Hello World" {
		t.Errorf("StripRTF plain text = %q, want %q", result, "Hello World")
	}
}

func TestStripRTF_Bold(t *testing.T) {
	rtf := `{\rtf1\ansi{\b bold text}}`
	result := StripRTF(rtf)
	if !strings.Contains(result, "bold text") {
		t.Errorf("StripRTF should preserve text content, got %q", result)
	}
}

func TestStripRTF_Par(t *testing.T) {
	rtf := `{\rtf1 line1\par line2}`
	result := StripRTF(rtf)
	if !strings.Contains(result, "\n") {
		t.Errorf("StripRTF should convert \\par to newline, got %q", result)
	}
}

func TestRTFToHTML_PlainText(t *testing.T) {
	result := RTFToHTML("Hello World")
	if !strings.Contains(result, "Hello World") {
		t.Errorf("RTFToHTML plain = %q, want to contain 'Hello World'", result)
	}
}

func TestRTFToHTML_Bold(t *testing.T) {
	rtf := `{\rtf1\ansi {\b Bold Text}}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "<b>") || !strings.Contains(result, "</b>") {
		t.Errorf("RTFToHTML should emit <b> tags for \\b, got: %q", result)
	}
	if !strings.Contains(result, "Bold Text") {
		t.Errorf("RTFToHTML should preserve text content, got: %q", result)
	}
}

func TestRTFToHTML_Italic(t *testing.T) {
	rtf := `{\rtf1\ansi {\i Italic Text}}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "<i>") || !strings.Contains(result, "</i>") {
		t.Errorf("RTFToHTML should emit <i> tags for \\i, got: %q", result)
	}
}

func TestRTFToHTML_Underline(t *testing.T) {
	rtf := `{\rtf1\ansi {\ul Underlined}}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "<u>") || !strings.Contains(result, "</u>") {
		t.Errorf("RTFToHTML should emit <u> tags for \\ul, got: %q", result)
	}
}

func TestRTFToHTML_ParagraphBreak(t *testing.T) {
	rtf := `{\rtf1 Para 1\par Para 2}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "<br>") {
		t.Errorf("RTFToHTML should emit <br> for \\par, got: %q", result)
	}
	if !strings.Contains(result, "Para 1") || !strings.Contains(result, "Para 2") {
		t.Errorf("RTFToHTML should preserve both paragraphs, got: %q", result)
	}
}

func TestRTFToHTML_HTMLEscaping(t *testing.T) {
	rtf := `{\rtf1 a < b & c > d}`
	result := RTFToHTML(rtf)
	if strings.Contains(result, "<b") && !strings.Contains(result, "&lt;") {
		t.Errorf("RTFToHTML should HTML-escape < in plain text, got: %q", result)
	}
}

func TestRTFToHTML_SpecialChars(t *testing.T) {
	rtf := `{\rtf1\ansi \endash \emdash}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "&ndash;") {
		t.Errorf("RTFToHTML should emit &ndash; for \\endash, got: %q", result)
	}
	if !strings.Contains(result, "&mdash;") {
		t.Errorf("RTFToHTML should emit &mdash; for \\emdash, got: %q", result)
	}
}

func TestRTFToHTML_MixedFormatting(t *testing.T) {
	rtf := `{\rtf1 Normal {\b Bold {\i Both}} Normal}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "Normal") {
		t.Errorf("RTFToHTML should preserve normal text, got: %q", result)
	}
	if !strings.Contains(result, "<b>") {
		t.Errorf("RTFToHTML should emit bold, got: %q", result)
	}
}

// ── Additional branch coverage ─────────────────────────────────────────────────

func TestRTFToHTML_Tab(t *testing.T) {
	rtf := `{\rtf1 col1\tab col2}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "col1") || !strings.Contains(result, "col2") {
		t.Errorf("RTFToHTML tab: missing text content, got: %q", result)
	}
	// \tab should emit &nbsp; sequences
	if !strings.Contains(result, "&nbsp;") {
		t.Errorf("RTFToHTML tab: expected &nbsp; in output, got: %q", result)
	}
}

func TestRTFToHTML_Line(t *testing.T) {
	rtf := `{\rtf1 line1\line line2}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "<br>") {
		t.Errorf("RTFToHTML \\line: expected <br>, got: %q", result)
	}
}

func TestRTFToHTML_HexEscape_ASCII(t *testing.T) {
	// \'41 = 'A'
	rtf := `{\rtf1 \'41}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "A") {
		t.Errorf("RTFToHTML hex escape \\' ASCII: expected 'A', got: %q", result)
	}
}

func TestRTFToHTML_HexEscape_HighByte(t *testing.T) {
	// \'e9 = 0xe9 (high byte) → &#233;
	rtf := `{\rtf1 \'e9}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "&#") {
		t.Errorf("RTFToHTML hex high-byte: expected &#NNN; entity, got: %q", result)
	}
}

func TestRTFToHTML_UnicodeEscape(t *testing.T) {
	// \u233 = é
	rtf := `{\rtf1 \u233?}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "&#233;") {
		t.Errorf("RTFToHTML unicode: expected &#233;, got: %q", result)
	}
}

func TestRTFToHTML_UnicodeEscape_Negative(t *testing.T) {
	// Negative code point (high surrogate) should be corrected via +65536
	rtf := `{\rtf1 \u-1?}`
	result := RTFToHTML(rtf)
	// Should not panic; result may be an entity or empty
	_ = result
}

func TestRTFToHTML_Quotes(t *testing.T) {
	rtf := `{\rtf1 \lquote hello\rquote}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "&lsquo;") || !strings.Contains(result, "&rsquo;") {
		t.Errorf("RTFToHTML quotes: got %q", result)
	}
}

func TestRTFToHTML_DoubleQuotes(t *testing.T) {
	rtf := `{\rtf1 \ldblquote hello\rdblquote}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "&ldquo;") || !strings.Contains(result, "&rdquo;") {
		t.Errorf("RTFToHTML double quotes: got %q", result)
	}
}

func TestRTFToHTML_Bullet(t *testing.T) {
	rtf := `{\rtf1 \bullet item}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "&bull;") {
		t.Errorf("RTFToHTML bullet: expected &bull;, got: %q", result)
	}
}

func TestRTFToHTML_Spaces(t *testing.T) {
	rtf := `{\rtf1 \enspace\emspace\qmspace}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "&nbsp;") {
		t.Errorf("RTFToHTML spaces: expected &nbsp;, got: %q", result)
	}
}

func TestRTFToHTML_UlNone(t *testing.T) {
	rtf := `{\rtf1 {\ul underline\ulnone plain}}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "<u>") {
		t.Errorf("RTFToHTML ulnone: expected <u> before ulnone, got: %q", result)
	}
}

func TestRTFToHTML_FontSize(t *testing.T) {
	rtf := `{\rtf1 {\fs24 sized text}}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "sized text") {
		t.Errorf("RTFToHTML font size: expected text, got: %q", result)
	}
}

func TestRTFToHTML_SkipDestination(t *testing.T) {
	// {\*\keyword content} should be skipped entirely
	rtf := `{\rtf1 visible{\*\fonttbl hidden}text}`
	result := RTFToHTML(rtf)
	if strings.Contains(result, "hidden") {
		t.Errorf("RTFToHTML skip destination: 'hidden' should not appear, got: %q", result)
	}
	if !strings.Contains(result, "text") {
		t.Errorf("RTFToHTML skip destination: 'text' should appear, got: %q", result)
	}
}

func TestRTFToHTML_ControlSymbol_Hyphen(t *testing.T) {
	// \- is an optional hyphen
	rtf := `{\rtf1 soft\-hyphen}`
	result := RTFToHTML(rtf)
	_ = result // just verify no panic
}

func TestRTFToHTML_ControlSymbol_Tilde(t *testing.T) {
	rtf := `{\rtf1 non\~breaking}`
	result := RTFToHTML(rtf)
	_ = result
}

func TestRTFToHTML_ControlSymbol_Pipe(t *testing.T) {
	rtf := `{\rtf1 index\|entry}`
	result := RTFToHTML(rtf)
	_ = result
}

func TestRTFToHTML_UnknownControlWord(t *testing.T) {
	// Unknown control words should be silently discarded
	rtf := `{\rtf1 \unknownword123 hello}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "hello") {
		t.Errorf("RTFToHTML unknown word: expected 'hello' in output, got: %q", result)
	}
}

func TestRTFToHTML_NegativeParam(t *testing.T) {
	// \fi-360 is a common indentation control word with negative parameter
	rtf := `{\rtf1 \fi-360 hello}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "hello") {
		t.Errorf("RTFToHTML negative param: expected 'hello', got: %q", result)
	}
}

func TestRTFToHTML_Pard(t *testing.T) {
	// \pard resets paragraph formatting
	rtf := `{\rtf1 {\b bold\pard normal}}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "bold") || !strings.Contains(result, "normal") {
		t.Errorf("RTFToHTML pard: got %q", result)
	}
}

func TestRTFToHTML_BackslashAtEnd(t *testing.T) {
	// backslash at end of string should not panic
	rtf := "{\rtf1 hello\\}"
	result := RTFToHTML(rtf)
	_ = result
}

// ── StripRTF additional branches ──────────────────────────────────────────────

func TestStripRTF_HexEscape(t *testing.T) {
	// \'41 = 'A'
	rtf := `{\rtf1 \'41}`
	result := StripRTF(rtf)
	if !strings.Contains(result, "A") {
		t.Errorf("StripRTF hex \\': expected 'A', got: %q", result)
	}
}

func TestStripRTF_Tab(t *testing.T) {
	rtf := `{\rtf1 col1\tab col2}`
	result := StripRTF(rtf)
	if !strings.Contains(result, "col1") || !strings.Contains(result, "col2") {
		t.Errorf("StripRTF tab: got %q", result)
	}
}

func TestStripRTF_UnknownWord(t *testing.T) {
	rtf := `{\rtf1 \unknownword hello}`
	result := StripRTF(rtf)
	if !strings.Contains(result, "hello") {
		t.Errorf("StripRTF unknown word: expected 'hello', got: %q", result)
	}
}

// ── skipControlWord direct coverage ───────────────────────────────────────────

func TestSkipControlWord_ControlSymbol(t *testing.T) {
	// backslash followed by non-alpha (e.g., \')
	rtf := `\' rest`
	end := skipControlWord(rtf, 0)
	if end != 2 { // consumed '\' and '\''
		t.Errorf("skipControlWord symbol: end = %d, want 2", end)
	}
}

func TestSkipControlWord_WordOnly(t *testing.T) {
	rtf := `\par rest`
	end := skipControlWord(rtf, 0)
	// skip '\' + 'par' + ' ' = 5
	if end != 5 {
		t.Errorf("skipControlWord word: end = %d, want 5", end)
	}
}

func TestSkipControlWord_WordWithParam(t *testing.T) {
	rtf := `\fs24 rest`
	end := skipControlWord(rtf, 0)
	// '\' + 'fs' + '24' + ' ' = 6
	if end != 6 {
		t.Errorf("skipControlWord word+param: end = %d, want 6", end)
	}
}

func TestSkipControlWord_WordWithNegativeParam(t *testing.T) {
	rtf := `\fi-360 text`
	end := skipControlWord(rtf, 0)
	// '\' + 'fi' + '-360' + ' ' = 8
	if end != 8 {
		t.Errorf("skipControlWord negative param: end = %d, want 8", end)
	}
}

func TestSkipControlWord_AtEnd(t *testing.T) {
	rtf := `\`
	end := skipControlWord(rtf, 0)
	if end != 1 {
		t.Errorf("skipControlWord at end: end = %d, want 1", end)
	}
}

// ── hexVal direct coverage ────────────────────────────────────────────────────

func TestHexVal(t *testing.T) {
	cases := []struct {
		c    byte
		want int
	}{
		{'0', 0}, {'9', 9},
		{'a', 10}, {'f', 15},
		{'A', 10}, {'F', 15},
		{'g', -1}, {'z', -1}, {'!', -1},
	}
	for _, tc := range cases {
		got := hexVal(tc.c)
		if got != tc.want {
			t.Errorf("hexVal(%q) = %d, want %d", tc.c, got, tc.want)
		}
	}
}

// ── writeRune / encodeRuneUTF8 direct coverage ────────────────────────────────

func TestEncodeRuneUTF8(t *testing.T) {
	cases := []struct {
		r    rune
		want string
	}{
		{'A', "A"},           // 1-byte (< 0x80)
		{0xE9, "é"},          // 2-byte (< 0x800)
		{0x4E2D, "中"},       // 3-byte (< 0x10000)
		{0x1F600, "😀"},     // 4-byte (≥ 0x10000)
	}
	for _, tc := range cases {
		buf := make([]byte, 4)
		n := encodeRuneUTF8(buf, tc.r)
		got := string(buf[:n])
		if got != tc.want {
			t.Errorf("encodeRuneUTF8(%U) = %q, want %q", tc.r, got, tc.want)
		}
	}
}

func TestWriteRune(t *testing.T) {
	var sb strings.Builder
	writeRune(&sb, '€') // U+20AC, 3-byte UTF-8
	if sb.String() != "€" {
		t.Errorf("writeRune = %q, want '€'", sb.String())
	}
}

// ── Header destination group skipping ─────────────────────────────────────────

func TestStripRTF_FonttblSkipped(t *testing.T) {
	// Font names inside \fonttbl must not appear in the plain-text output.
	rtf := `{\rtf1{\fonttbl{\f0 Arial;}}Hello}`
	result := StripRTF(rtf)
	if strings.Contains(result, "Arial") {
		t.Errorf("StripRTF fonttbl: font name 'Arial' leaked into output: %q", result)
	}
	if !strings.Contains(result, "Hello") {
		t.Errorf("StripRTF fonttbl: expected 'Hello' in output, got: %q", result)
	}
}

func TestStripRTF_ColortblSkipped(t *testing.T) {
	// \colortbl entries (semicolons and colour control words) must not leak.
	rtf := `{\rtf1{\colortbl ;\red255\green0\blue0;}Hello}`
	result := StripRTF(rtf)
	// "red", "green", "blue" come from control words and are consumed, but the
	// semicolons are literal text characters that would formerly leak through.
	if strings.Contains(result, ";") {
		t.Errorf("StripRTF colortbl: semicolons leaked into output: %q", result)
	}
	if !strings.Contains(result, "Hello") {
		t.Errorf("StripRTF colortbl: expected 'Hello' in output, got: %q", result)
	}
}

func TestStripRTF_StylesheetSkipped(t *testing.T) {
	// Style names inside \stylesheet must not appear in output.
	rtf := `{\rtf1{\stylesheet{\s0 Normal;}}Hello}`
	result := StripRTF(rtf)
	if strings.Contains(result, "Normal") {
		t.Errorf("StripRTF stylesheet: style name 'Normal' leaked into output: %q", result)
	}
	if !strings.Contains(result, "Hello") {
		t.Errorf("StripRTF stylesheet: expected 'Hello' in output, got: %q", result)
	}
}

func TestStripRTF_StarredDestinationStillSkipped(t *testing.T) {
	// Pre-existing behaviour: {\*\keyword ...} must remain skipped.
	rtf := `{\rtf1 visible{\*\somekw hidden}text}`
	result := StripRTF(rtf)
	if strings.Contains(result, "hidden") {
		t.Errorf("StripRTF starred dest: 'hidden' should not appear, got: %q", result)
	}
	if !strings.Contains(result, "visible") || !strings.Contains(result, "text") {
		t.Errorf("StripRTF starred dest: expected visible text, got: %q", result)
	}
}

func TestRTFToHTML_FonttblSkipped(t *testing.T) {
	// Font names inside \fonttbl must not appear in the HTML output.
	rtf := `{\rtf1{\fonttbl{\f0 Arial;}}Hello}`
	result := RTFToHTML(rtf)
	if strings.Contains(result, "Arial") {
		t.Errorf("RTFToHTML fonttbl: font name 'Arial' leaked into output: %q", result)
	}
	if !strings.Contains(result, "Hello") {
		t.Errorf("RTFToHTML fonttbl: expected 'Hello' in output, got: %q", result)
	}
}

func TestRTFToHTML_ColortblSkipped(t *testing.T) {
	// Colour table semicolons must not appear in HTML output.
	rtf := `{\rtf1{\colortbl ;\red255\green0\blue0;}Hello}`
	result := RTFToHTML(rtf)
	if strings.Contains(result, ";") {
		t.Errorf("RTFToHTML colortbl: semicolons leaked into output: %q", result)
	}
	if !strings.Contains(result, "Hello") {
		t.Errorf("RTFToHTML colortbl: expected 'Hello' in output, got: %q", result)
	}
}

func TestRTFToHTML_StylesheetSkipped(t *testing.T) {
	// Style names inside \stylesheet must not appear in HTML output.
	rtf := `{\rtf1{\stylesheet{\s0 Normal;}}Hello}`
	result := RTFToHTML(rtf)
	if strings.Contains(result, "Normal") {
		t.Errorf("RTFToHTML stylesheet: style name 'Normal' leaked into output: %q", result)
	}
	if !strings.Contains(result, "Hello") {
		t.Errorf("RTFToHTML stylesheet: expected 'Hello' in output, got: %q", result)
	}
}

func TestIsRTFHeaderDestination(t *testing.T) {
	yes := []string{
		"fonttbl", "colortbl", "stylesheet", "info",
		"header", "footer", "pict", "object", "fldinst",
		"private", "revtbl", "listtable", "listoverridetable",
		"rsidtbl", "generator", "latentstyles",
	}
	for _, name := range yes {
		if !isRTFHeaderDestination(name) {
			t.Errorf("isRTFHeaderDestination(%q) = false, want true", name)
		}
	}
	no := []string{"b", "i", "ul", "par", "fs", "rtf", "ansi", "unknown"}
	for _, name := range no {
		if isRTFHeaderDestination(name) {
			t.Errorf("isRTFHeaderDestination(%q) = true, want false", name)
		}
	}
}

func TestPeekControlWordName(t *testing.T) {
	cases := []struct {
		rtf  string
		i    int
		want string
	}{
		{`\fonttbl rest`, 0, "fonttbl"},
		{`\par `, 0, "par"},
		{`\b`, 0, "b"},
		{`\' hex`, 0, ""},  // control symbol, not a word
		{`text`, 0, ""},    // no backslash
		{`\`, 0, ""},       // backslash only
	}
	for _, tc := range cases {
		got := peekControlWordName(tc.rtf, tc.i)
		if got != tc.want {
			t.Errorf("peekControlWordName(%q, %d) = %q, want %q", tc.rtf, tc.i, got, tc.want)
		}
	}
}

// ── parseSignedInt ────────────────────────────────────────────────────────────

func TestParseSignedInt(t *testing.T) {
	cases := []struct {
		s    string
		want int
	}{
		{"", 0},
		{"42", 42},
		{"-360", -360},
		{"0", 0},
		{"1234abc", 1234}, // stops at non-digit
	}
	for _, tc := range cases {
		got := parseSignedInt(tc.s)
		if got != tc.want {
			t.Errorf("parseSignedInt(%q) = %d, want %d", tc.s, got, tc.want)
		}
	}
}
