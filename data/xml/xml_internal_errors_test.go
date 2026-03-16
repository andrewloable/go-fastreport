// Package xml internal tests covering error paths in unexported helpers.
// These tests use package xml (not xml_test) to access unexported functions.
package xml

// xml_internal_errors_test.go — covers the error-return paths inside:
//   - skipToChildElement: dec.Token() error, dec.Skip() error
//   - readChildren:       dec.Token() error, readElementText error
//   - readElementText:    dec.Token() error, dec.Skip() error
//
// All of these require injecting a decoder whose underlying reader fails
// mid-stream. We do this by composing a valid XML prefix (already consumed)
// with an io.Reader that returns an error on subsequent reads.

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"
)

// errReader is an io.Reader that serves initial bytes from buf and then
// returns errAfter on all subsequent reads.
type errReader struct {
	buf      *strings.Reader
	errAfter error
}

func (e *errReader) Read(p []byte) (int, error) {
	if e.buf.Len() > 0 {
		n, err := e.buf.Read(p)
		if err == io.EOF {
			// Don't expose EOF; next call will return errAfter.
			return n, nil
		}
		return n, err
	}
	return 0, e.errAfter
}

var errInjected = errors.New("injected reader error")

// newErrDecoder builds an xml.Decoder whose reader serves `prefix` and then
// returns errInjected. The prefix must be valid XML that can be tokenised
// without buffering issues, but incomplete enough that Token() will call Read
// again once the prefix is exhausted.
func newErrDecoder(prefix string) *xml.Decoder {
	r := &errReader{
		buf:      strings.NewReader(prefix),
		errAfter: errInjected,
	}
	return xml.NewDecoder(r)
}

// ── skipToChildElement ────────────────────────────────────────────────────────

// TestSkipToChildElement_TokenError covers the dec.Token() error path in
// skipToChildElement (the `if err != nil { return err }` branch at line 213).
// We position the decoder inside an open element and then let the reader fail.
func TestSkipToChildElement_TokenError(t *testing.T) {
	// Feed just enough to open <Root> (so skipToElement can consume it), then
	// break. skipToChildElement will call Token() looking for <Target> and get
	// the injected error instead.
	dec := newErrDecoder(`<Root>`)

	// Consume the <Root> start element first (simulating what parseXML does
	// after calling skipToElement).
	tok, err := dec.Token()
	if err != nil {
		t.Fatalf("unexpected error consuming <Root>: %v", err)
	}
	if _, ok := tok.(xml.StartElement); !ok {
		t.Fatalf("expected StartElement, got %T", tok)
	}

	// Now the reader is exhausted; the next Token() call will hit the injected
	// error inside skipToChildElement.
	gotErr := skipToChildElement(dec, "Target")
	if gotErr == nil {
		t.Fatal("skipToChildElement: expected error from broken reader, got nil")
	}
}

// TestSkipToChildElement_SkipError covers the dec.Skip() error path inside
// skipToChildElement when a non-matching subtree cannot be skipped.
// We feed <Root><Other> (an open sub-element that we never close), so Token()
// returns <Other> (StartElement, name != target), dec.Skip() tries to read the
// closing </Other> and hits the injected error.
func TestSkipToChildElement_SkipError(t *testing.T) {
	// "<Root><Other>" — starts <Other> but never closes it, so Skip() must
	// read more tokens and will hit the error.
	dec := newErrDecoder(`<Root><Other>`)

	// Consume the <Root> start element.
	if _, err := dec.Token(); err != nil {
		t.Fatalf("unexpected error consuming <Root>: %v", err)
	}

	// skipToChildElement will see <Other> (not "Target"), try to Skip() it,
	// and fail because the closing tag is missing and the reader is broken.
	gotErr := skipToChildElement(dec, "Target")
	if gotErr == nil {
		t.Fatal("skipToChildElement: expected Skip() error, got nil")
	}
}

// ── readChildren ──────────────────────────────────────────────────────────────

// TestReadChildren_TokenError covers the dec.Token() error path inside
// readChildren (the `if err != nil { return err }` branch at line 236).
func TestReadChildren_TokenError(t *testing.T) {
	// Feed just enough to open <Row> (the parent), then break. readChildren
	// will call Token() looking for child elements and get the error.
	dec := newErrDecoder(`<Root><Row>`)

	// Consume <Root> and <Row>.
	for range 2 {
		if _, err := dec.Token(); err != nil {
			t.Fatalf("unexpected error consuming tokens: %v", err)
		}
	}

	row := make(map[string]any)
	colSet := make(map[string]bool)
	var colOrder []string
	gotErr := readChildren(dec, "Row", row, colSet, &colOrder)
	if gotErr == nil {
		t.Fatal("readChildren: expected Token() error, got nil")
	}
}

// TestReadChildren_ReadElementTextError covers the path where readElementText
// returns an error inside readChildren. We feed <Row><Col> with the reader
// then failing, so readElementText's Token() call breaks.
func TestReadChildren_ReadElementTextError(t *testing.T) {
	// <Root><Row><Col> — opens Col but never provides content or closing tag.
	dec := newErrDecoder(`<Root><Row><Col>`)

	// Consume <Root> and <Row>.
	for range 2 {
		if _, err := dec.Token(); err != nil {
			t.Fatalf("unexpected error consuming tokens: %v", err)
		}
	}

	row := make(map[string]any)
	colSet := make(map[string]bool)
	var colOrder []string
	gotErr := readChildren(dec, "Row", row, colSet, &colOrder)
	if gotErr == nil {
		t.Fatal("readChildren: expected readElementText error, got nil")
	}
}

// ── readElementText ───────────────────────────────────────────────────────────

// TestReadElementText_TokenError covers the dec.Token() error path inside
// readElementText (the `if err != nil { return "", err }` branch at line 265).
func TestReadElementText_TokenError(t *testing.T) {
	// Feed just enough to open <Name> (already consumed by caller), then fail.
	// readElementText is called after the StartElement has been consumed, so
	// we start after <Name>.
	dec := newErrDecoder(`<Root><Row><Name>`)

	// Consume <Root>, <Row>, <Name>.
	for range 3 {
		if _, err := dec.Token(); err != nil {
			t.Fatalf("unexpected error consuming tokens: %v", err)
		}
	}

	// Reader is now exhausted; readElementText will call Token() and hit error.
	_, gotErr := readElementText(dec, "Name")
	if gotErr == nil {
		t.Fatal("readElementText: expected Token() error, got nil")
	}
}

// TestReadElementText_SkipError covers the dec.Skip() error path inside
// readElementText when a nested element cannot be fully skipped.
// We feed <Root><Row><Name><Nested> — the nested element's closing tag is
// missing and the reader fails, so Skip() returns an error.
func TestReadElementText_SkipError(t *testing.T) {
	// <Root><Row><Name><Nested> — outer Name element containing Nested that
	// is never closed. After consuming <Name>, readElementText sees <Nested>
	// (StartElement), calls dec.Skip(), which fails due to missing close tag.
	dec := newErrDecoder(`<Root><Row><Name><Nested>`)

	// Consume <Root>, <Row>, <Name>.
	for range 3 {
		if _, err := dec.Token(); err != nil {
			t.Fatalf("unexpected error consuming tokens: %v", err)
		}
	}

	// readElementText will see <Nested> and call Skip(); the reader fails.
	_, gotErr := readElementText(dec, "Name")
	if gotErr == nil {
		t.Fatal("readElementText: expected Skip() error, got nil")
	}
}

// ── parseXML error paths ──────────────────────────────────────────────────────

// TestParseXML_TokenError covers the dec.Token() non-EOF error path inside
// parseXML's main loop (line 150: `if err != nil { return nil, nil, err }`).
// The decoder must be inside the container element (after rootPath navigation)
// and then hit a reader error on the next Token() call.
func TestParseXML_TokenError(t *testing.T) {
	// Feed <Root> as the container element (no rootPath needed). After
	// parseXML's skipToElement consumes <Root>, the main loop calls Token()
	// and hits the injected reader error.
	raw := []byte(`<Root>`)
	// Use a reader that serves <Root> then fails.
	_, _, gotErr := parseXML(raw, "", "")
	// The reader error should propagate from Token() in the main loop.
	// (The standard library decoder may return io.EOF here since the input is
	//  just "<Root>" with nothing after it — EOF is handled by the break, not
	//  the error path. So we need a broken reader, not just a short input.)
	// We can't inject a broken reader through parseXML's public signature, so
	// use a short-but-complete XML to hit the EOF branch instead and verify
	// the function works normally. The actual error injection test is below.
	_ = gotErr // not an error path test — just shape-testing

	// Real error injection: use the errReader approach via a thin wrapper.
	// parseXML constructs its own decoder internally; we cannot inject a broken
	// reader via its API. Instead, test the Token-error path via a malformed
	// XML body that yields a non-EOF decode error mid-stream.
	//
	// XML with a valid root open tag followed by invalid content:
	// After skipToElement consumes <Root>, the main loop calls Token() on
	// `< invalid` — the decoder returns a syntax error (not io.EOF).
	brokenRaw := []byte("<Root>< invalid")
	_, _, err2 := parseXML(brokenRaw, "", "")
	if err2 == nil {
		t.Fatal("parseXML: expected Token() decode error for malformed XML, got nil")
	}
}

// TestParseXML_SkipSiblingError covers the dec.Skip() error return path inside
// parseXML's main loop (line 165: `if err := dec.Skip(); err != nil { return }`).
// This fires when rowElem is set, Token() returns a StartElement for a sibling
// with a different name (entering the skip branch), and then dec.Skip() fails
// because the sibling's subtree is malformed.
//
// Key constraint: Token() must succeed (returning the <Other> StartElement) so
// we enter the `if start.Name.Local != rowElem` branch. Then Skip() must fail
// because it cannot find the matching </Other> end tag.
//
// XML: after <Row Name="A"/>, we open <Other> successfully, then insert
// invalid XML that causes Skip()'s internal Token() calls to error out.
func TestParseXML_SkipSiblingError(t *testing.T) {
	// "<Root><Row Name="A"/><Other>< invalid" —
	//   Token() → <Root>  (consumed by skipToElement)
	//   Token() → <Row Name="A"/>  (processed as row)
	//   Token() → <Other> StartElement (name != "Row", enter skip branch)
	//   dec.Skip() calls Token() → encounters "< invalid" → syntax error
	raw := []byte(`<Root><Row Name="A"/><Other>< invalid`)
	_, _, err := parseXML(raw, "", "Row")
	if err == nil {
		t.Fatal("parseXML: expected dec.Skip() error for malformed sibling subtree, got nil")
	}
}

// TestParseXML_ReadChildrenError covers the readChildren() error path inside
// parseXML's main loop (line 185: `if err := readChildren(...); err != nil`).
// This fires when a matching row element is found but its children cannot be
// read because the decoder fails mid-stream.
//
// We feed XML where <Row> starts (so parseXML enters the row-building code)
// but the element content is malformed, causing readChildren's Token() to
// return a decode error.
func TestParseXML_ReadChildrenError(t *testing.T) {
	// <Root><Row followed immediately by invalid XML — the decoder matches
	// <Row> as a StartElement (rowElem auto-detected), then readChildren
	// calls Token() and gets a syntax error from the malformed `<` content.
	raw := []byte("<Root><Row>< bad")
	_, _, err := parseXML(raw, "", "")
	if err == nil {
		t.Fatal("parseXML: expected readChildren error for malformed row content, got nil")
	}
}
