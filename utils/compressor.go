package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
)

// gzipMagic holds the two-byte GZip magic number (0x1F 0x8B).
var gzipMagic = [2]byte{0x1F, 0x8B}

// compressNewWriter is a package-level hook so tests can inject a custom
// io.Writer destination into Compress to exercise its error paths.
// In production it always supplies a fresh *bytes.Buffer.
var compressNewWriter func() (io.Writer, *bytes.Buffer) = func() (io.Writer, *bytes.Buffer) {
	b := &bytes.Buffer{}
	return b, b
}

// Compress compresses data using GZip (the same format as the FastReport .NET
// Compressor class, which uses GZipStream) and returns the result as a
// standard base64-encoded string.
func Compress(data []byte) (string, error) {
	dest, buf := compressNewWriter()
	w := gzip.NewWriter(dest)
	if _, err := w.Write(data); err != nil {
		w.Close()
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// Decompress decodes a base64-encoded string produced by Compress and returns
// the original uncompressed bytes.  If the decoded bytes do not begin with
// the GZip magic number the raw decoded bytes are returned unchanged,
// matching the behaviour of the .NET Compressor.Decompress(byte[]) method.
func Decompress(s string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	// If not gzip-compressed, return raw bytes (mirrors .NET behaviour).
	if len(decoded) < 2 || decoded[0] != gzipMagic[0] || decoded[1] != gzipMagic[1] {
		return decoded, nil
	}

	r, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return out, nil
}
