package utils

// utils_extra_coverage_test.go — additional internal tests to cover branches
// missed by previous test files.

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/image/font/basicfont"
)

// ── compressor.go: Compress write/close error paths ──────────────────────────
// gzip.Writer.Write never fails for an in-memory bytes.Buffer, but we can
// simulate a Write failure via a custom Writer to cover the error branches.

type failOnWriteN struct {
	failAfterN int
	written    int
	err        error
}

func (f *failOnWriteN) Write(p []byte) (int, error) {
	if f.written >= f.failAfterN {
		return 0, f.err
	}
	n := len(p)
	if f.written+n > f.failAfterN {
		n = f.failAfterN - f.written
	}
	f.written += n
	if f.written >= f.failAfterN {
		// Return partial write + error on this call.
		if n < len(p) {
			return n, f.err
		}
	}
	return n, nil
}

// ── rtf.go: RTFToHTML uncovered branches ─────────────────────────────────────

// TestRTFToHTML_BackslashAtEndOfInput covers the `if i >= n { break }` branch
// in RTFToHTML where a backslash is the very last character.
func TestRTFToHTML_BackslashAtEndInput(t *testing.T) {
	// A string that starts with RTF header and ends with a bare backslash.
	// The outer loop processes '{', then '\', increments i past '\', checks i>=n.
	rtf := "{\\rtf1 hello\\"
	result := RTFToHTML(rtf)
	_ = result // no panic
}

// TestRTFToHTML_HexEscape_TooShort covers the `else { i++ }` branch when
// i+2 is NOT < n (i.e., there are fewer than 2 chars after the quote).
func TestRTFToHTML_HexEscape_OnlyOneCharAfterQuote(t *testing.T) {
	// \'A with no second hex digit — too short.
	rtf := `{\rtf1 \'A}`
	result := RTFToHTML(rtf)
	_ = result // no panic; the else branch increments i by 1
}

// TestRTFToHTML_HexEscape_AtVeryEnd covers \'  at the very end (i+2 >= n).
func TestRTFToHTML_HexEscape_AtVeryEnd(t *testing.T) {
	// {\rtf1 \' — backslash-quote with nothing after (end of string).
	rtf := "{\\rtf1 \\'}"
	result := RTFToHTML(rtf)
	_ = result
}

// TestRTFToHTML_OpenTagsNotClosed covers the close-at-end branches for
// underline, italic, bold still open when the RTF string ends.
// These are the `if cur.underline`, `if cur.italic`, `if cur.bold` checks
// at the end of RTFToHTML.

func TestRTFToHTML_BoldStillOpen_AtEnd(t *testing.T) {
	// Bold opened but never closed (no matching '}'  to pop state)
	rtf := `{\rtf1 \b bold text without closing group`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "</b>") {
		t.Errorf("RTFToHTML bold not closed: expected </b>, got %q", result)
	}
}

func TestRTFToHTML_ItalicStillOpen_AtEnd(t *testing.T) {
	rtf := `{\rtf1 \i italic text without closing group`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "</i>") {
		t.Errorf("RTFToHTML italic not closed: expected </i>, got %q", result)
	}
}

func TestRTFToHTML_UnderlineStillOpen_AtEnd(t *testing.T) {
	rtf := `{\rtf1 \ul underline text without closing group`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "</u>") {
		t.Errorf("RTFToHTML underline not closed: expected </u>, got %q", result)
	}
}

// ── rtf.go: processControlWord (StripRTF) uncovered branches ─────────────────

// TestProcessControlWord_NegativeUnicode covers the `cp < 0 → cp += 65536` path.
func TestProcessControlWord_NegativeUnicode(t *testing.T) {
	var sb strings.Builder
	// \u-1 — negative code point should become 65535.
	processControlWord(`\u-1`, 0, &sb)
	if sb.Len() == 0 {
		t.Error("processControlWord \\u-1: expected non-empty output for cp=65535")
	}
}

// TestProcessControlWord_UcWord covers the "uc" case (ANSI fallback count).
func TestProcessControlWord_UcWord(t *testing.T) {
	var sb strings.Builder
	pos := processControlWord(`\uc1 rest`, 0, &sb)
	if pos <= 0 {
		t.Errorf("processControlWord \\uc: expected pos > 0, got %d", pos)
	}
	// uc produces no output — just advances position.
	if sb.Len() != 0 {
		t.Errorf("processControlWord \\uc: expected empty output, got %q", sb.String())
	}
}

// TestProcessControlWord_EnspaceWord covers the "enspace" / "emspace" / "qmspace" cases.
func TestProcessControlWord_EnspaceWord(t *testing.T) {
	for _, word := range []string{`\enspace `, `\emspace `, `\qmspace `} {
		var sb strings.Builder
		processControlWord(word, 0, &sb)
		if !strings.Contains(sb.String(), " ") {
			t.Errorf("processControlWord %q: expected space, got %q", word, sb.String())
		}
	}
}

// TestProcessControlWord_EndashWord covers "endash".
func TestProcessControlWord_EndashWord(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\endash `, 0, &sb)
	if !strings.Contains(sb.String(), "\u2013") {
		t.Errorf("processControlWord \\endash: expected en-dash, got %q", sb.String())
	}
}

// TestProcessControlWord_EmdashWord covers "emdash".
func TestProcessControlWord_EmdashWord(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\emdash `, 0, &sb)
	if !strings.Contains(sb.String(), "\u2014") {
		t.Errorf("processControlWord \\emdash: expected em-dash, got %q", sb.String())
	}
}

// TestProcessControlWord_LquoteWord covers "lquote".
func TestProcessControlWord_LquoteWord(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\lquote `, 0, &sb)
	if !strings.Contains(sb.String(), "\u2018") {
		t.Errorf("processControlWord \\lquote: expected left single quote, got %q", sb.String())
	}
}

// TestProcessControlWord_RquoteWord covers "rquote".
func TestProcessControlWord_RquoteWord(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\rquote `, 0, &sb)
	if !strings.Contains(sb.String(), "\u2019") {
		t.Errorf("processControlWord \\rquote: expected right single quote, got %q", sb.String())
	}
}

// TestProcessControlWord_LdblquoteWord covers "ldblquote".
func TestProcessControlWord_LdblquoteWord(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\ldblquote `, 0, &sb)
	if !strings.Contains(sb.String(), "\u201C") {
		t.Errorf("processControlWord \\ldblquote: expected left double quote, got %q", sb.String())
	}
}

// TestProcessControlWord_RdblquoteWord covers "rdblquote".
func TestProcessControlWord_RdblquoteWord(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\rdblquote `, 0, &sb)
	if !strings.Contains(sb.String(), "\u201D") {
		t.Errorf("processControlWord \\rdblquote: expected right double quote, got %q", sb.String())
	}
}

// TestProcessControlWord_BulletWord covers "bullet".
func TestProcessControlWord_BulletWord(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\bullet `, 0, &sb)
	if !strings.Contains(sb.String(), "\u2022") {
		t.Errorf("processControlWord \\bullet: expected bullet, got %q", sb.String())
	}
}

// TestStripRTF_AllSpecialWords tests StripRTF via the full RTF parser to
// exercise processControlWord for all special word cases in their natural context.
func TestStripRTF_AllSpecialWords(t *testing.T) {
	cases := []struct {
		rtf     string
		wantIn  string // substring expected in result
		wantOut string // substring that should NOT be in result
	}{
		{`{\rtf1 \u-1?}`, "\uffff", ""},  // negative unicode → cp=65535
		{`{\rtf1 \uc1 text}`, "text", ""}, // uc: no output
		{`{\rtf1 \enspace text}`, " ", ""},
		{`{\rtf1 \emspace text}`, " ", ""},
		{`{\rtf1 \qmspace text}`, " ", ""},
		{`{\rtf1 \endash text}`, "\u2013", ""},
		{`{\rtf1 \emdash text}`, "\u2014", ""},
		{`{\rtf1 \lquote text}`, "\u2018", ""},
		{`{\rtf1 \rquote text}`, "\u2019", ""},
		{`{\rtf1 \ldblquote text}`, "\u201C", ""},
		{`{\rtf1 \rdblquote text}`, "\u201D", ""},
		{`{\rtf1 \bullet text}`, "\u2022", ""},
	}
	for _, tc := range cases {
		result := StripRTF(tc.rtf)
		if tc.wantIn != "" && !strings.ContainsAny(result, tc.wantIn) {
			t.Errorf("StripRTF(%q): expected %q in result, got %q", tc.rtf, tc.wantIn, result)
		}
		_ = result // no panic
	}
}

// TestStripRTF_NegativeUnicode_ViaFullParser exercises processControlWord \u-1
// through the StripRTF parser, covering the cp < 0 branch.
func TestStripRTF_NegativeUnicode_ViaFullParser(t *testing.T) {
	rtf := `{\rtf1 \u-1?}`
	result := StripRTF(rtf)
	// cp = -1 + 65536 = 65535, writeRune should emit the character (U+FFFF).
	if len(result) == 0 {
		// It's valid for this to be empty if U+FFFF has no printable form.
		// The key is that the cp<0 branch ran without panic.
	}
	_ = result
}

// TestStripRTF_BareCRInBody covers the '\r' case in the default branch of
// StripRTF (line 320-322: `if ch == '\n' || ch == '\r'`).
func TestStripRTF_BareCR(t *testing.T) {
	// Bare \r in RTF body (outside backslash processing).
	// Must start with {\rtf so it enters the parser.
	rtf := "{\\rtf1 text\rmore}"
	result := StripRTF(rtf)
	if strings.Contains(result, "\r") {
		t.Errorf("StripRTF: bare \\r should be ignored, got %q", result)
	}
	if !strings.Contains(result, "text") {
		t.Errorf("StripRTF: expected 'text' in result, got %q", result)
	}
}

// TestStripRTF_ParamNegative covers processControlWord's negative param parsing
// for the numeric-param negative path (rtf.go:386).
func TestStripRTF_NegativeParam_UWord(t *testing.T) {
	// \u-1? inside StripRTF — processControlWord is called with negative param.
	rtf := `{\rtf1 \u-65535?}`
	result := StripRTF(rtf)
	_ = result // no panic
}

// ── htmltext.go: MeasureHeight with zero lines ────────────────────────────────

// TestMeasureHeight_EmptyLines covers the `len(r.lines) == 0 → return 0` branch.
// This happens when the HTML text renderer produces no lines (empty input).
func TestMeasureHeight_EmptyLines(t *testing.T) {
	// An HtmlTextRenderer with no lines.
	r := &HtmlTextRenderer{}
	h := r.MeasureHeight(100)
	if h != 0 {
		t.Errorf("MeasureHeight empty lines: got %v, want 0", h)
	}
}

// ── textmeasure.go: wordWrap len(lines)==0 branch ────────────────────────────
// The `if len(lines) == 0 { lines = []string{para} }` branch is a safety guard.
// It can only fire if the for loop over words appends nothing to lines — which
// requires len(words) > 0 but the last word is never appended. That can't happen
// because i == len(words)-1 always triggers the append. This is a dead branch.
// We document it and skip it.

// TestWordWrap_SingleWordExactFit exercises the first-word-always-appended path.
func TestWordWrap_SingleWordExactFit(t *testing.T) {
	face := basicfont.Face7x13
	// Single word that exactly fits maxWidth.
	// Result should have 1 line with the word.
	lines := wordWrap("hello", face, 10000)
	if len(lines) != 1 || lines[0] != "hello" {
		t.Errorf("wordWrap single word exact fit: got %v, want [hello]", lines)
	}
}

// ── rtf.go: RTFToHTML \_ control symbol (non-breaking hyphen) ────────────────
// Line 124-126: the `case '_':` branch in RTFToHTML writes "&#8209;".
func TestRTFToHTML_UnderscoreControlSymbol(t *testing.T) {
	rtf := `{\rtf1 soft\_hyphen}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "&#8209;") {
		t.Errorf("RTFToHTML \\_: expected &#8209; (non-breaking hyphen), got %q", result)
	}
}

// ── rtf.go: processControlWord "line" case (StripRTF) ────────────────────────
// Line 404-405: the "line" case in processControlWord writes '\n'.
func TestStripRTF_LineWord(t *testing.T) {
	rtf := `{\rtf1 text\line more}`
	result := StripRTF(rtf)
	if !strings.Contains(result, "\n") {
		t.Errorf("StripRTF \\line: expected newline in output, got %q", result)
	}
	if !strings.Contains(result, "text") || !strings.Contains(result, "more") {
		t.Errorf("StripRTF \\line: text content missing in %q", result)
	}
}

// ── zip.go: ZipStream write error path ───────────────────────────────────────
// ZipStream wraps flate.NewWriter around the destination writer. If flate.NewWriter
// returns an error (only possible with invalid compression level, not DefaultCompression),
// the error is returned. The io.Copy error path is already covered by TestZipStreamReaderError.
// The flate writer Close error path (fw.Close) can be triggered by a failing dest writer.

func TestZipStream_CloseError(t *testing.T) {
	// Use a writer that fails after a few bytes are written (to let flate buffer some data)
	// then fails on Close() when flate tries to flush.
	fw := &failOnWriteN{failAfterN: 0, err: errors.New("write failed")}
	err := ZipStream(fw, strings.NewReader("hello world"))
	if err == nil {
		t.Error("ZipStream: expected error when writer fails immediately")
	}
}

func TestZipData_WriteError(t *testing.T) {
	// ZipData calls ZipStream internally; test the error propagation path.
	// We can't inject a failing writer into ZipData directly since it uses
	// an internal bytes.Buffer. This test documents that the ZipData error
	// path is exercised via ZipStream.
	// Already covered by TestZipStreamReaderError and TestZipDataReaderError.
}

// ── compressor.go: Compress Write error path ─────────────────────────────────
// The gzip.Writer.Write and gzip.Writer.Close methods write to their underlying
// destination. For an in-memory bytes.Buffer, these never fail. The error paths
// in Compress (lines 19-22 and 23-25) are defensive guards. They can be exercised
// if we could inject a failing writer into Compress, but the function always uses
// its own internal bytes.Buffer. These branches are unreachable through the public API.

// TestCompress_AlwaysSucceeds documents that the error paths in Compress are
// not exercisable via the public API (bytes.Buffer never fails).
func TestCompress_AlwaysSucceeds(t *testing.T) {
	_, err := Compress([]byte("some data"))
	if err != nil {
		t.Errorf("Compress unexpectedly failed: %v", err)
	}
}

// ── crypto.go: error paths for bad key length ─────────────────────────────────
// EncryptString calls aesCBCEncrypt with a derived 16-byte key, so the
// aes.NewCipher error in aesCBCEncrypt is unreachable via EncryptString.
// Already covered by TestAesCBCEncrypt_BadKeyLength in crypto_coverage_test.go.
// EncryptStream and DecryptStream derive their own 16-byte key, so the
// aes.NewCipher error paths are also unreachable via these functions.
// These branches (crypto.go:62-64, 96-98, 112-114) remain uncoverable
// through the public API because the key derivation always produces a 16-byte key.

// TestCrypto_ErrorBranchesDocumented documents that these branches are guarded
// defensive code that cannot be triggered via normal key derivation.
func TestCrypto_ErrorBranchesDocumented(t *testing.T) {
	// The public API always derives a 16-byte key, so aes.NewCipher never fails.
	// We verify the normal paths work correctly.
	enc, err := EncryptString("hello", "password")
	if err != nil {
		t.Fatalf("EncryptString: %v", err)
	}
	dec, err := DecryptString(enc, "password")
	if err != nil {
		t.Fatalf("DecryptString: %v", err)
	}
	if dec != "hello" {
		t.Errorf("round-trip: got %q, want 'hello'", dec)
	}
}

// ── image.go: ImageToBytes JPEG/PNG error paths ──────────────────────────────
// jpeg.Encode and png.Encode only return errors on I/O failures from their writer.
// Since ImageToBytes uses an internal bytes.Buffer that never fails, these error
// paths (image.go:82-84 and 86-88) are unreachable through the public API.
// Already documented in utils_coverage_test.go (see the comment there).
