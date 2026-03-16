package utils

// crypto_zip_error_hooks_test.go — internal tests that cover error branches in
// EncryptString, EncryptStream, DecryptStream (crypto.go) and ZipStream
// (zip.go) that are unreachable via the public API because deriveKeyIV always
// produces a valid 16-byte AES key and flate.DefaultCompression is always valid.
// Hook-based injection follows the same pattern as compressor_image_zip_errors_test.go.

import (
	"bytes"
	"compress/flate"
	"crypto/cipher"
	"errors"
	"io"
	"testing"
)

var errInjectedCipher = errors.New("injected aes.NewCipher failure")

// ── EncryptString: aesCBCEncrypt cipher error path ────────────────────────────

func TestEncryptString_AesNewCipherError_ViaHook(t *testing.T) {
	orig := cryptoAesNewCipher
	defer func() { cryptoAesNewCipher = orig }()

	cryptoAesNewCipher = func(_ []byte) (cipher.Block, error) {
		return nil, errInjectedCipher
	}

	_, err := EncryptString("hello", "password")
	if err == nil {
		t.Error("EncryptString: expected error from injected aes.NewCipher failure, got nil")
	}
}

// ── EncryptStream: aes.NewCipher error path ───────────────────────────────────

func TestEncryptStream_AesNewCipherError_ViaHook(t *testing.T) {
	orig := cryptoAesNewCipher
	defer func() { cryptoAesNewCipher = orig }()

	cryptoAesNewCipher = func(_ []byte) (cipher.Block, error) {
		return nil, errInjectedCipher
	}

	var buf bytes.Buffer
	_, err := EncryptStream(&buf, "password")
	if err == nil {
		t.Error("EncryptStream: expected error from injected aes.NewCipher failure, got nil")
	}
}

// ── DecryptStream: aes.NewCipher error path ───────────────────────────────────

func TestDecryptStream_AesNewCipherError_ViaHook(t *testing.T) {
	orig := cryptoAesNewCipher
	defer func() { cryptoAesNewCipher = orig }()

	cryptoAesNewCipher = func(_ []byte) (cipher.Block, error) {
		return nil, errInjectedCipher
	}

	_, err := DecryptStream(bytes.NewReader([]byte("somedata")), "password")
	if err == nil {
		t.Error("DecryptStream: expected error from injected aes.NewCipher failure, got nil")
	}
}

// ── ZipStream: flate.NewWriter error path ─────────────────────────────────────

func TestZipStream_FlateNewWriterError_ViaHook(t *testing.T) {
	orig := zipFlateNewWriter
	defer func() { zipFlateNewWriter = orig }()

	zipFlateNewWriter = func(_ io.Writer, _ int) (*flate.Writer, error) {
		return nil, errors.New("injected flate.NewWriter failure")
	}

	var buf bytes.Buffer
	err := ZipStream(&buf, bytes.NewReader([]byte("test data")))
	if err == nil {
		t.Error("ZipStream: expected error from injected flate.NewWriter failure, got nil")
	}
}
