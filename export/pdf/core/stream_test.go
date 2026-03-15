package core

import (
	"bytes"
	"compress/zlib"
	"io"
	"strings"
	"testing"
)

func TestStream_Type(t *testing.T) {
	s := NewStream()
	if s.Type() != TypeStream {
		t.Fatalf("expected TypeStream, got %q", s.Type())
	}
}

func TestNewStream_Defaults(t *testing.T) {
	s := NewStream()
	if s.Dict == nil {
		t.Fatal("Dict should be non-nil")
	}
	if !s.Compressed {
		t.Fatal("Compressed should default to true")
	}
}

func TestStream_WriteTo_Uncompressed(t *testing.T) {
	s := NewStream()
	s.Compressed = false
	s.Data = []byte("Hello PDF")

	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()

	if !strings.Contains(got, "stream\n") {
		t.Fatalf("missing 'stream' keyword: %q", got)
	}
	if !strings.Contains(got, "\nendstream\n") {
		t.Fatalf("missing 'endstream' keyword: %q", got)
	}
	// Should not have /Filter when uncompressed
	if strings.Contains(got, "/Filter") {
		t.Fatalf("unexpected /Filter in uncompressed stream: %q", got)
	}
	// Should have /Length 9
	if !strings.Contains(got, "/Length 9") {
		t.Fatalf("expected /Length 9 in output: %q", got)
	}
	// Data should be present
	if !strings.Contains(got, "Hello PDF") {
		t.Fatalf("data not found in output: %q", got)
	}
}

func TestStream_WriteTo_Compressed(t *testing.T) {
	s := NewStream()
	s.Compressed = true
	s.Data = []byte("compress me please")

	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()

	if !strings.Contains(got, "/Filter /FlateDecode") {
		t.Fatalf("expected /Filter /FlateDecode: %q", got)
	}
	if !strings.Contains(got, "stream\n") {
		t.Fatalf("missing 'stream' keyword: %q", got)
	}
	if !strings.Contains(got, "\nendstream\n") {
		t.Fatalf("missing 'endstream': %q", got)
	}
}

func TestStream_WriteTo_CompressedDataRoundtrip(t *testing.T) {
	original := []byte("round trip test data 1234567890")
	s := NewStream()
	s.Compressed = true
	s.Data = original

	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.Bytes()

	// Extract the compressed bytes between "stream\n" and "\nendstream"
	streamMarker := []byte("stream\n")
	endMarker := []byte("\nendstream\n")
	start := bytes.Index(output, streamMarker)
	end := bytes.Index(output, endMarker)
	if start < 0 || end < 0 || end <= start {
		t.Fatalf("could not find stream boundaries in output")
	}
	compressed := output[start+len(streamMarker) : end]

	// Decompress and verify
	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("zlib.NewReader: %v", err)
	}
	decompressed, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("decompression error: %v", err)
	}
	if !bytes.Equal(decompressed, original) {
		t.Fatalf("decompressed data mismatch: got %q want %q", decompressed, original)
	}
}

func TestStream_WriteTo_NilData(t *testing.T) {
	s := NewStream()
	s.Compressed = false
	s.Data = nil

	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// Should have /Length 0
	if !strings.Contains(got, "/Length 0") {
		t.Fatalf("expected /Length 0 for nil data: %q", got)
	}
}

func TestStream_WriteTo_WithExtraDict(t *testing.T) {
	s := NewStream()
	s.Compressed = false
	s.Dict.Add("Subtype", NewName("Image"))
	s.Data = []byte("img")

	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "/Subtype /Image") {
		t.Fatalf("extra dict entry missing: %q", got)
	}
}

func TestStream_WriteTo_ByteCountMatches(t *testing.T) {
	s := NewStream()
	s.Compressed = false
	s.Data = []byte("abc")

	var buf bytes.Buffer
	n, _ := s.WriteTo(&buf)
	if n != int64(buf.Len()) {
		t.Fatalf("byte count mismatch: WriteTo=%d buf.Len=%d", n, buf.Len())
	}
}

func TestZlibCompress_RoundTrip(t *testing.T) {
	src := []byte("test compression round trip")
	compressed, err := zlibCompress(src)
	if err != nil {
		t.Fatal(err)
	}
	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, src) {
		t.Fatalf("round trip failed: got %q want %q", got, src)
	}
}

func TestZlibCompress_Empty(t *testing.T) {
	compressed, err := zlibCompress([]byte{})
	if err != nil {
		t.Fatal(err)
	}
	if len(compressed) == 0 {
		t.Fatal("expected non-empty compressed output for empty input")
	}
	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(r)
	if len(got) != 0 {
		t.Fatalf("expected empty decompressed output, got %q", got)
	}
}

