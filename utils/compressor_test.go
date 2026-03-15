package utils

import (
	"encoding/base64"
	"testing"
)

func TestCompressDecompressRoundTrip(t *testing.T) {
	original := []byte("Hello, FastReport! Compress and decompress me.")

	encoded, err := Compress(original)
	if err != nil {
		t.Fatalf("Compress error: %v", err)
	}
	if encoded == "" {
		t.Fatal("Compress returned empty string")
	}

	decoded, err := Decompress(encoded)
	if err != nil {
		t.Fatalf("Decompress error: %v", err)
	}

	if string(decoded) != string(original) {
		t.Fatalf("round-trip mismatch: got %q, want %q", decoded, original)
	}
}

func TestCompressDecompressEmpty(t *testing.T) {
	encoded, err := Compress([]byte{})
	if err != nil {
		t.Fatalf("Compress(empty) error: %v", err)
	}

	decoded, err := Decompress(encoded)
	if err != nil {
		t.Fatalf("Decompress(empty compressed) error: %v", err)
	}
	if len(decoded) != 0 {
		t.Fatalf("expected empty output, got %d bytes", len(decoded))
	}
}

func TestCompressDecompressLargeData(t *testing.T) {
	// 64 KB of repeated pattern
	original := make([]byte, 64*1024)
	for i := range original {
		original[i] = byte(i % 251)
	}

	encoded, err := Compress(original)
	if err != nil {
		t.Fatalf("Compress error: %v", err)
	}

	decoded, err := Decompress(encoded)
	if err != nil {
		t.Fatalf("Decompress error: %v", err)
	}

	if string(decoded) != string(original) {
		t.Fatal("large data round-trip mismatch")
	}
}

func TestDecompressOutputIsGzipFormat(t *testing.T) {
	// Verify that Compress produces GZip-framed data (magic bytes 0x1F 0x8B).
	encoded, err := Compress([]byte("test"))
	if err != nil {
		t.Fatalf("Compress error: %v", err)
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("base64 decode error: %v", err)
	}
	if len(raw) < 2 || raw[0] != 0x1F || raw[1] != 0x8B {
		t.Fatalf("expected GZip magic bytes, got 0x%02X 0x%02X", raw[0], raw[1])
	}
}

func TestDecompressNonGzipPassthrough(t *testing.T) {
	// If the decoded bytes do not start with GZip magic, they are returned as-is.
	raw := []byte("plain text, not compressed")
	encoded := base64.StdEncoding.EncodeToString(raw)

	out, err := Decompress(encoded)
	if err != nil {
		t.Fatalf("Decompress error: %v", err)
	}
	if string(out) != string(raw) {
		t.Fatalf("non-gzip passthrough mismatch: got %q, want %q", out, raw)
	}
}

func TestDecompressInvalidBase64(t *testing.T) {
	_, err := Decompress("this is not valid base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64, got nil")
	}
}

func TestDecompressInvalidGzip(t *testing.T) {
	// Craft bytes starting with GZip magic but otherwise invalid (header too short).
	bad := []byte{0x1F, 0x8B, 0x00, 0x01, 0x02, 0x03}
	encoded := base64.StdEncoding.EncodeToString(bad)
	_, err := Decompress(encoded)
	if err == nil {
		t.Fatal("expected error for invalid gzip body, got nil")
	}
}

func TestDecompressCorruptGzipBody(t *testing.T) {
	// Valid 10-byte gzip header + invalid deflate body → io.ReadAll error.
	// Header: magic(2) + method(1) + flags(1) + mtime(4) + xfl(1) + os(1)
	header := []byte{0x1F, 0x8B, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF}
	// Garbage deflate bytes that will fail decompression.
	body := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	bad := append(header, body...)
	encoded := base64.StdEncoding.EncodeToString(bad)
	_, err := Decompress(encoded)
	if err == nil {
		t.Fatal("expected error for corrupt gzip body, got nil")
	}
}
