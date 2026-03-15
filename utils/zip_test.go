package utils

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestZipDataRoundTrip(t *testing.T) {
	original := []byte("Hello, FastReport! This is a test string for compression.")

	compressed, err := ZipData(original)
	if err != nil {
		t.Fatalf("ZipData error: %v", err)
	}
	if len(compressed) == 0 {
		t.Fatal("ZipData returned empty bytes")
	}

	decompressed, err := UnzipData(compressed)
	if err != nil {
		t.Fatalf("UnzipData error: %v", err)
	}
	if !bytes.Equal(original, decompressed) {
		t.Fatalf("round-trip mismatch: got %q, want %q", decompressed, original)
	}
}

func TestZipDataEmpty(t *testing.T) {
	compressed, err := ZipData([]byte{})
	if err != nil {
		t.Fatalf("ZipData(empty) error: %v", err)
	}

	decompressed, err := UnzipData(compressed)
	if err != nil {
		t.Fatalf("UnzipData(empty compressed) error: %v", err)
	}
	if len(decompressed) != 0 {
		t.Fatalf("expected empty output, got %d bytes", len(decompressed))
	}
}

func TestZipDataBinary(t *testing.T) {
	original := make([]byte, 256)
	for i := range original {
		original[i] = byte(i)
	}

	compressed, err := ZipData(original)
	if err != nil {
		t.Fatalf("ZipData error: %v", err)
	}

	decompressed, err := UnzipData(compressed)
	if err != nil {
		t.Fatalf("UnzipData error: %v", err)
	}
	if !bytes.Equal(original, decompressed) {
		t.Fatal("binary round-trip mismatch")
	}
}

func TestZipStreamRoundTrip(t *testing.T) {
	original := []byte(strings.Repeat("abcdefghij", 100))

	var compressed bytes.Buffer
	if err := ZipStream(&compressed, bytes.NewReader(original)); err != nil {
		t.Fatalf("ZipStream error: %v", err)
	}

	var decompressed bytes.Buffer
	if err := UnzipStream(&decompressed, &compressed); err != nil {
		t.Fatalf("UnzipStream error: %v", err)
	}

	if !bytes.Equal(original, decompressed.Bytes()) {
		t.Fatal("stream round-trip mismatch")
	}
}

func TestUnzipDataInvalidInput(t *testing.T) {
	_, err := UnzipData([]byte("not valid deflate data!!!"))
	if err == nil {
		t.Fatal("expected error for invalid deflate input, got nil")
	}
}

func TestUnzipStreamInvalidInput(t *testing.T) {
	var out bytes.Buffer
	err := UnzipStream(&out, strings.NewReader("garbage"))
	if err == nil {
		t.Fatal("expected error for invalid deflate stream, got nil")
	}
}

// errReader returns an error after the specified number of bytes.
type errReader struct {
	data []byte
	pos  int
	fail int // fail after this many bytes read
}

func (r *errReader) Read(p []byte) (int, error) {
	if r.pos >= r.fail {
		return 0, fmt.Errorf("simulated read error")
	}
	n := copy(p, r.data[r.pos:])
	if r.pos+n > r.fail {
		n = r.fail - r.pos
	}
	r.pos += n
	if r.pos >= r.fail {
		return n, fmt.Errorf("simulated read error")
	}
	return n, nil
}

func TestZipStreamReaderError(t *testing.T) {
	// errReader will error after 0 bytes, so io.Copy from it will fail.
	var out bytes.Buffer
	er := &errReader{data: []byte("some data"), fail: 0}
	err := ZipStream(&out, er)
	if err == nil {
		t.Fatal("expected error from ZipStream with failing reader, got nil")
	}
}

func TestZipDataReaderError(t *testing.T) {
	// ZipData uses ZipStream internally; this tests the error propagation.
	// We can't inject an errReader into ZipData directly, so test via ZipStream.
	var out bytes.Buffer
	er := &errReader{data: []byte("hello world"), fail: 3}
	err := ZipStream(&out, er)
	if err == nil {
		t.Fatal("expected error from ZipStream with partially-failing reader")
	}
}
