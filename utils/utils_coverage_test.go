package utils

// utils_coverage_test.go — internal package tests for uncovered branches in
// image.go, rtf.go, and compressor.go.
// Uses package utils (not utils_test) to access unexported helpers.

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── image.go: loadFromFile error paths ───────────────────────────────────────

func TestLoadFromFile_NotExist(t *testing.T) {
	_, err := loadFromFile("/no/such/file/ever.png")
	if err == nil {
		t.Error("loadFromFile for non-existent path should return error")
	}
}

func TestLoadFromFile_InvalidContent(t *testing.T) {
	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.png")
	if err := os.WriteFile(badFile, []byte("not a png"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, err := loadFromFile(badFile)
	if err == nil {
		t.Error("loadFromFile with invalid image content should return error")
	}
}

// ── image.go: scaleDraw empty rects ──────────────────────────────────────────

func TestScaleDraw_EmptyRects(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	dst := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	// empty srcRect → early return (no panic)
	scaleDraw(img, image.Rect(0, 0, 0, 0), dst, image.Rect(0, 0, 5, 5))
	// empty dstRect → early return (no panic)
	scaleDraw(img, image.Rect(0, 0, 5, 5), dst, image.Rect(0, 0, 0, 0))
}

// ── image.go: ImageToBytes with unknown format (defaults to PNG) ──────────────

func TestImageToBytes_UnknownFormat(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	out, err := ImageToBytes(img, ImageFormat(99))
	if err != nil {
		t.Fatalf("ImageToBytes unknown format: %v", err)
	}
	if len(out) == 0 {
		t.Error("expected non-empty PNG output for unknown format")
	}
}

// ── image.go: loadFromDataURI plain body ─────────────────────────────────────

func TestLoadFromDataURI_PlainBody(t *testing.T) {
	// data: URI without ;base64 — body treated as raw bytes → image decode fails.
	uri := "data:text/plain,hello world"
	_, err := loadFromDataURI(uri)
	// No panic expected; image decode of "hello world" will fail.
	_ = err
}

func TestLoadFromDataURI_Malformed_NoComma(t *testing.T) {
	_, err := loadFromDataURI("data:image/png;base64")
	if err == nil {
		t.Error("expected error for malformed data URI (no comma)")
	}
}

// ── rtf.go: processControlWord uncovered control symbols ─────────────────────

func TestProcessControlWord_Hyphen(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\-rest`, 0, &sb)
	if !strings.Contains(sb.String(), "-") {
		t.Errorf("processControlWord \\-: expected '-', got %q", sb.String())
	}
}

func TestProcessControlWord_Tilde(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\~rest`, 0, &sb)
	if !strings.Contains(sb.String(), "-") {
		t.Errorf("processControlWord \\~: expected '-', got %q", sb.String())
	}
}

func TestProcessControlWord_Underscore(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\_rest`, 0, &sb)
	if !strings.Contains(sb.String(), "-") {
		t.Errorf("processControlWord \\_: expected '-', got %q", sb.String())
	}
}

func TestProcessControlWord_Pipe(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\|rest`, 0, &sb)
	if sb.Len() != 0 {
		t.Errorf("processControlWord \\|: expected empty output, got %q", sb.String())
	}
}

func TestProcessControlWord_Colon(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\:rest`, 0, &sb)
	if sb.Len() != 0 {
		t.Errorf("processControlWord \\:: expected empty, got %q", sb.String())
	}
}

func TestProcessControlWord_Star(t *testing.T) {
	var sb strings.Builder
	processControlWord(`\*rest`, 0, &sb)
	if sb.Len() != 0 {
		t.Errorf("processControlWord \\*: expected empty, got %q", sb.String())
	}
}

func TestProcessControlWord_OtherSymbol(t *testing.T) {
	// e.g., \! is unknown control symbol — discarded
	var sb strings.Builder
	processControlWord(`\!rest`, 0, &sb)
	if sb.Len() != 0 {
		t.Errorf("processControlWord \\!: expected empty, got %q", sb.String())
	}
}

func TestProcessControlWord_EOFAfterBackslash(t *testing.T) {
	var sb strings.Builder
	pos := processControlWord(`\`, 0, &sb)
	_ = pos // verify no panic
}

func TestProcessControlWord_HexEscapeShort(t *testing.T) {
	// \'X where there is only 1 char after — too short for 2-digit hex.
	var sb strings.Builder
	processControlWord(`\'A`, 0, &sb)
	_ = sb.String() // no panic
}

// ── rtf.go: StripRTF uncovered branches ──────────────────────────────────────

func TestStripRTF_ControlSymbolUnderscore(t *testing.T) {
	rtf := `{\rtf1 non\_breaking}`
	result := StripRTF(rtf)
	if !strings.Contains(result, "-") {
		t.Errorf("StripRTF \\_: expected '-', got %q", result)
	}
}

func TestStripRTF_ControlSymbolColon(t *testing.T) {
	rtf := `{\rtf1 index\:entry}`
	result := StripRTF(rtf)
	_ = result // no panic
}

func TestStripRTF_DestinationWithControlSymbol(t *testing.T) {
	// Inside {\*\...}, a control symbol like \' should be skipped via skipControlWord.
	rtf := `{\rtf1 before{\*\fonttbl \'AB xyz}after}`
	result := StripRTF(rtf)
	if strings.Contains(result, "xyz") {
		t.Errorf("StripRTF: destination content should be skipped, got %q", result)
	}
	if !strings.Contains(result, "before") || !strings.Contains(result, "after") {
		t.Errorf("StripRTF: outer content missing in %q", result)
	}
}

func TestStripRTF_BareNewlines(t *testing.T) {
	// Bare \n/\r in RTF body should be ignored (not emitted as text output).
	rtf := "{\n\r\\rtf1\n text\r}"
	result := StripRTF(rtf)
	_ = result // no panic
}

func TestStripRTF_SkipGroupWithNumericParam(t *testing.T) {
	// A skip group with a numeric param control word.
	rtf := `{\rtf1 before{\*\keyword123 body}after}`
	result := StripRTF(rtf)
	if strings.Contains(result, "body") {
		t.Errorf("skip group content should not appear: %q", result)
	}
	if !strings.Contains(result, "before") {
		t.Errorf("outer text should appear: %q", result)
	}
}

// ── compressor.go: Decompress with less than 2 bytes ─────────────────────────

func TestDecompress_OneByte(t *testing.T) {
	// 1 decoded byte: len < 2, so not gzip → returned as-is.
	oneByte := []byte{0x42}
	encoded := base64.StdEncoding.EncodeToString(oneByte)
	out, err := Decompress(encoded)
	if err != nil {
		t.Fatalf("Decompress 1-byte: %v", err)
	}
	if len(out) != 1 || out[0] != 0x42 {
		t.Errorf("Decompress 1-byte: got %v, want [0x42]", out)
	}
}

// ── BytesToImage: empty input ─────────────────────────────────────────────────

func TestBytesToImage_Empty(t *testing.T) {
	_, err := BytesToImage([]byte{})
	if err == nil {
		t.Error("BytesToImage with empty bytes should return error")
	}
}

// ── LoadImage: URL branch exercised with bad host ────────────────────────────

func TestLoadImage_URLBadHost(t *testing.T) {
	// Exercises the loadFromURL error path.
	_, err := LoadImage("http://no.such.host.invalid/image.png")
	if err == nil {
		t.Error("LoadImage bad URL should return error")
	}
}

// ── loadFromURL: success and invalid-image-body paths ────────────────────────

// buildSmallPNGBytes creates a minimal 2×2 PNG image as raw bytes.
func buildSmallPNGBytes() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.NRGBA{R: 128, G: 64, B: 32, A: 255})
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestLoadFromURL_SuccessWithValidPNG(t *testing.T) {
	// Serve a valid PNG from a test HTTP server to cover the success path of
	// loadFromURL (stmts: defer body.Close, io.ReadAll, return BytesToImage).
	pngData := buildSmallPNGBytes()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(pngData)
	}))
	defer srv.Close()

	img, err := loadFromURL(srv.URL)
	if err != nil {
		t.Fatalf("loadFromURL success: unexpected error: %v", err)
	}
	if img == nil {
		t.Fatal("loadFromURL success: returned nil image")
	}
}

func TestLoadFromURL_SuccessWithInvalidBody(t *testing.T) {
	// Serve non-image bytes to cover the BytesToImage-failure path inside
	// loadFromURL (the body is read successfully but decoding fails).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("this is not an image"))
	}))
	defer srv.Close()

	_, err := loadFromURL(srv.URL)
	if err == nil {
		t.Error("loadFromURL with non-image body: expected error from BytesToImage")
	}
}

func TestLoadFromURL_HTTPErrorStatus(t *testing.T) {
	// 404 response: body is read and passed to BytesToImage, which fails because
	// the body is an error page, not an image. This covers the same code paths
	// as the invalid-body test but via an HTTP error status code.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := loadFromURL(srv.URL)
	if err == nil {
		t.Error("loadFromURL 404: expected error from BytesToImage")
	}
}

func TestLoadFromURL_BodyReadError(t *testing.T) {
	// Serve a response where the connection is hijacked and dropped after headers
	// are sent, causing io.ReadAll to return an error (covering the
	// "read body" error branch in loadFromURL).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hijack the connection so we can write a partial HTTP response and
		// then close it abruptly, causing the client's io.ReadAll to fail.
		hj, ok := w.(http.Hijacker)
		if !ok {
			// If hijacking is not supported, skip: write a large Content-Length
			// with no body so the client gets an unexpected EOF.
			w.Header().Set("Content-Length", "1000000")
			w.WriteHeader(http.StatusOK)
			// Write partial body then return — connection closed by server.
			_, _ = w.Write([]byte("partial"))
			return
		}
		conn, bufrw, err := hj.Hijack()
		if err != nil {
			return
		}
		// Write a minimal HTTP response header claiming a body of 1 MB, then
		// immediately close the connection — the client will get an unexpected EOF.
		_, _ = bufrw.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 1048576\r\n\r\npartial")
		_ = bufrw.Flush()
		_ = conn.Close()
	}))
	defer srv.Close()

	_, err := loadFromURL(srv.URL)
	if err == nil {
		t.Error("loadFromURL body read error: expected error from io.ReadAll or BytesToImage")
	}
}

// ── ImageToBytes: JPEG and PNG error paths ───────────────────────────────────

// badEncoderImage is a minimal image.Image that returns a non-empty Bounds() but
// panics if its pixels are accessed. jpeg.Encode and png.Encode iterate over
// pixels, so encoding will fail in a recoverable way only if the encoder itself
// surfaces an error. Since standard encoders don't return errors for valid images,
// these error paths cannot be exercised without a custom writer.
// Instead we verify the happy paths for JPEG (covered below) are hit correctly.
// The actual error branches (`return nil, err`) inside ImageToBytes are dead code
// because both jpeg.Encode and png.Encode only error on I/O failures to the writer,
// and the internal bytes.Buffer never fails. They are noted here for completeness.

func TestImageToBytes_JPEG_RoundTrip(t *testing.T) {
	// Encode a PNG image as JPEG and verify the output is non-empty and decodable.
	pngData := buildSmallPNGBytes()
	img, err := BytesToImage(pngData)
	if err != nil {
		t.Fatalf("BytesToImage: %v", err)
	}
	out, err := ImageToBytes(img, ImageFormatJPEG)
	if err != nil {
		t.Fatalf("ImageToBytes JPEG: %v", err)
	}
	if len(out) == 0 {
		t.Error("ImageToBytes JPEG: output is empty")
	}
	// Re-decode to verify it is a valid JPEG.
	decoded, err := BytesToImage(out)
	if err != nil {
		t.Fatalf("BytesToImage re-decode JPEG: %v", err)
	}
	if decoded == nil {
		t.Error("decoded JPEG image is nil")
	}
}

func TestImageToBytes_PNG_RoundTrip(t *testing.T) {
	// Encode a small image as PNG and verify the output is non-empty and decodable.
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	out, err := ImageToBytes(img, ImageFormatPNG)
	if err != nil {
		t.Fatalf("ImageToBytes PNG: %v", err)
	}
	if len(out) == 0 {
		t.Error("ImageToBytes PNG: output is empty")
	}
}
