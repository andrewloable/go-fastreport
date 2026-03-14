package utils

import (
	"bytes"
	"compress/flate"
	"io"
)

// ZipData compresses data using raw DEFLATE encoding, the same algorithm used
// by the FastReport .NET ZipArchive (which wraps streams in DeflateStream).
// The returned bytes contain raw DEFLATE-compressed data with no zlib or gzip
// framing header.
func ZipData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := ZipStream(&buf, bytes.NewReader(data)); err != nil {
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
	fw, err := flate.NewWriter(w, flate.DefaultCompression)
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
