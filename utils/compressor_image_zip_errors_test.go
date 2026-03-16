package utils

// compressor_image_zip_errors_test.go — internal tests that cover the error
// branches in Compress (compressor.go:19-22, 23-25), ImageToBytes
// (image.go:82-84, 86-88), and ZipData (zip.go:15-17) by temporarily
// injecting a failing writer via the package-level hooks added for testing.
//
// These branches are "defensive guards" that can never be reached via the
// public API because all three functions normally use an internal bytes.Buffer
// that never returns a write error.

import (
	"bytes"
	"errors"
	"image"
	"io"
	"testing"
)

// ── shared helpers ────────────────────────────────────────────────────────────

// immediateFailWriter returns an error on every Write call.
type immediateFailWriter struct{ err error }

func (w *immediateFailWriter) Write(_ []byte) (int, error) { return 0, w.err }

// ── Compress: gzip write-error path (compressor.go:19-22) ───────────────────

// TestCompress_WriteError_ViaHook exercises the `w.Write(data)` error branch
// in Compress by injecting a writer that fails immediately.
func TestCompress_WriteError_ViaHook(t *testing.T) {
	orig := compressNewWriter
	defer func() { compressNewWriter = orig }()

	compressNewWriter = func() (io.Writer, *bytes.Buffer) {
		fw := &immediateFailWriter{err: errors.New("injected write failure")}
		// Return a dummy buffer for the base64 encoding (never reached on error).
		return fw, &bytes.Buffer{}
	}

	_, err := Compress([]byte("hello"))
	if err == nil {
		t.Error("Compress: expected error from injected failing writer, got nil")
	}
}

// ── Compress: gzip close-error path (compressor.go:31-33) ───────────────────

// countingFailWriter accepts exactly successWrites writes then returns an error.
// It is used to let w.Write succeed but make w.Close fail.
type countingFailWriter struct {
	buf          bytes.Buffer
	successWrites int
	calls        int
	err          error
}

func (c *countingFailWriter) Write(p []byte) (int, error) {
	c.calls++
	if c.calls > c.successWrites {
		return 0, c.err
	}
	return c.buf.Write(p)
}

// TestCompress_CloseError_ViaHook exercises the `w.Close()` error branch in
// Compress (compressor.go:31-33) by injecting a writer that accepts the first
// Write call (gzip header) but fails on the Close flush.
func TestCompress_CloseError_ViaHook(t *testing.T) {
	orig := compressNewWriter
	defer func() { compressNewWriter = orig }()

	// Allow the first write (gzip header) to succeed, then fail on Close flush.
	cfw := &countingFailWriter{successWrites: 1, err: errors.New("injected close failure")}
	compressNewWriter = func() (io.Writer, *bytes.Buffer) {
		return cfw, &bytes.Buffer{}
	}

	_, err := Compress([]byte("data"))
	// The error should come from w.Close() since w.Write succeeded.
	if err == nil {
		t.Error("Compress: expected close error from injected failing writer, got nil")
	}
}

// ── ImageToBytes: JPEG encode-error path (image.go:82-84) ───────────────────

// TestImageToBytes_JPEG_EncodeError_ViaHook exercises the jpeg.Encode error
// branch in ImageToBytes by injecting a failing writer via imageToBytesWriter.
func TestImageToBytes_JPEG_EncodeError_ViaHook(t *testing.T) {
	orig := imageToBytesWriter
	defer func() { imageToBytesWriter = orig }()

	imageToBytesWriter = func() io.Writer {
		return &immediateFailWriter{err: errors.New("injected jpeg write failure")}
	}

	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	_, err := ImageToBytes(img, ImageFormatJPEG)
	if err == nil {
		t.Error("ImageToBytes JPEG: expected error from injected failing writer, got nil")
	}
}

// ── ImageToBytes: PNG encode-error path (image.go:86-88) ────────────────────

// TestImageToBytes_PNG_EncodeError_ViaHook exercises the png.Encode error
// branch in ImageToBytes by injecting a failing writer via imageToBytesWriter.
func TestImageToBytes_PNG_EncodeError_ViaHook(t *testing.T) {
	orig := imageToBytesWriter
	defer func() { imageToBytesWriter = orig }()

	imageToBytesWriter = func() io.Writer {
		return &immediateFailWriter{err: errors.New("injected png write failure")}
	}

	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	_, err := ImageToBytes(img, ImageFormatPNG)
	if err == nil {
		t.Error("ImageToBytes PNG: expected error from injected failing writer, got nil")
	}
}

// ── ZipData: ZipStream error path (zip.go:15-17) ─────────────────────────────

// TestZipData_ErrorPath_ViaHook exercises the `if err := ZipStream(...)`
// error branch in ZipData by injecting a failing writer via zipDataWriter.
func TestZipData_ErrorPath_ViaHook(t *testing.T) {
	orig := zipDataWriter
	defer func() { zipDataWriter = orig }()

	zipDataWriter = func() io.Writer {
		return &immediateFailWriter{err: errors.New("injected deflate write failure")}
	}

	_, err := ZipData([]byte("test data"))
	if err == nil {
		t.Error("ZipData: expected error from injected failing writer, got nil")
	}
}
