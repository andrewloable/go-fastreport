package utils

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"io"
)

// zipDataWriter is a package-level hook so tests can inject a custom io.Writer
// as the destination for ZipData to exercise its error path. In production it
// is always nil, causing ZipData to use its own internal bytes.Buffer.
var zipDataWriter func() io.Writer

// zipFlateNewWriter is a package-level hook so tests can inject a failing
// flate writer factory to exercise the ZipStream error branch that is
// unreachable via the public API (flate.DefaultCompression is always valid).
var zipFlateNewWriter = flate.NewWriter

// ZipData compresses data using raw DEFLATE encoding, the same algorithm used
// by the FastReport .NET ZipArchive (which wraps streams in DeflateStream).
// The returned bytes contain raw DEFLATE-compressed data with no zlib or gzip
// framing header.
func ZipData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	var dest io.Writer = &buf
	if zipDataWriter != nil {
		dest = zipDataWriter()
	}
	if err := ZipStream(dest, bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnzipData decompresses raw DEFLATE-compressed data produced by ZipData or
// any compatible DeflateStream writer.
func UnzipData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := UnzipStream(&buf, bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ZipStream reads all data from r, compresses it using raw DEFLATE, and writes
// the compressed bytes to w.
func ZipStream(w io.Writer, r io.Reader) error {
	fw, err := zipFlateNewWriter(w, flate.DefaultCompression)
	if err != nil {
		return err
	}
	if _, err = io.Copy(fw, r); err != nil {
		fw.Close()
		return err
	}
	return fw.Close()
}

// UnzipStream reads raw DEFLATE-compressed data from r and writes the
// decompressed bytes to w.
func UnzipStream(w io.Writer, r io.Reader) error {
	fr := flate.NewReader(r)
	defer fr.Close()
	_, err := io.Copy(w, fr)
	return err
}

// ── ZipArchive ────────────────────────────────────────────────────────────────

// ZipArchive is a simple in-memory ZIP archive builder.
// Mirrors the FastReport .NET ZipArchive helper used to construct XLSX/ODS files.
type ZipArchive struct {
	buf bytes.Buffer
	w   *zip.Writer
}

// NewZipArchive creates an empty ZipArchive.
func NewZipArchive() *ZipArchive {
	za := &ZipArchive{}
	za.w = zip.NewWriter(&za.buf)
	return za
}

// AddEntry adds a named entry with the given byte content.
func (za *ZipArchive) AddEntry(name string, data []byte) error {
	return za.AddEntryFromStream(name, bytes.NewReader(data))
}

// AddEntryFromStream adds a named entry by reading from r.
func (za *ZipArchive) AddEntryFromStream(name string, r io.Reader) error {
	fw, err := za.w.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, r)
	return err
}

// Bytes closes the archive (if not already closed) and returns the ZIP bytes.
func (za *ZipArchive) Bytes() ([]byte, error) {
	if za.w != nil {
		if err := za.w.Close(); err != nil {
			return nil, err
		}
		za.w = nil
	}
	return za.buf.Bytes(), nil
}
