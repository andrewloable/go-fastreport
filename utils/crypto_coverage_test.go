package utils

// crypto_coverage_test.go — internal tests for uncovered error paths in crypto.go.
// Uses package utils (not utils_test) to access unexported helpers like
// pkcs7Unpad, aesCBCEncrypt, aesCBCDecrypt, and cipherWriter.

import (
	"bytes"
	"crypto/aes"
	"errors"
	"io"
	"testing"
)

// ── pkcs7Unpad error paths ────────────────────────────────────────────────────

func TestPkcs7Unpad_PadZero(t *testing.T) {
	// pad byte == 0 → invalid
	data := make([]byte, 16)
	data[15] = 0
	_, err := pkcs7Unpad(data)
	if err == nil {
		t.Error("pkcs7Unpad: pad=0 should return error")
	}
}

func TestPkcs7Unpad_PadExceedsBlockSize(t *testing.T) {
	// pad byte > aes.BlockSize (16) → invalid
	data := make([]byte, 16)
	data[15] = 17 // > 16
	_, err := pkcs7Unpad(data)
	if err == nil {
		t.Error("pkcs7Unpad: pad > blockSize should return error")
	}
}

func TestPkcs7Unpad_PadExceedsDataLen(t *testing.T) {
	// pad byte > len(data) → invalid
	data := []byte{5} // len=1, pad=5
	_, err := pkcs7Unpad(data)
	if err == nil {
		t.Error("pkcs7Unpad: pad > len(data) should return error")
	}
}

func TestPkcs7Unpad_Empty(t *testing.T) {
	result, err := pkcs7Unpad([]byte{})
	if err != nil {
		t.Fatalf("pkcs7Unpad empty: unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("pkcs7Unpad empty: expected empty, got %v", result)
	}
}

func TestPkcs7Unpad_ValidPad(t *testing.T) {
	// Correctly padded block: "hello" + 11 bytes of 0x0b
	padded := make([]byte, 16)
	copy(padded, []byte("hello"))
	for i := 5; i < 16; i++ {
		padded[i] = 11
	}
	result, err := pkcs7Unpad(padded)
	if err != nil {
		t.Fatalf("pkcs7Unpad valid: %v", err)
	}
	if string(result) != "hello" {
		t.Errorf("pkcs7Unpad valid: got %q, want 'hello'", result)
	}
}

// ── aesCBCEncrypt / aesCBCDecrypt error paths ─────────────────────────────────

func TestAesCBCDecrypt_NotMultipleOfBlock(t *testing.T) {
	key := make([]byte, 16)
	iv := make([]byte, 16)
	// Ciphertext not multiple of block size
	_, err := aesCBCDecrypt([]byte{1, 2, 3}, key, iv)
	if err == nil {
		t.Error("aesCBCDecrypt: non-block-aligned ciphertext should error")
	}
}

func TestAesCBCDecrypt_Empty(t *testing.T) {
	key := make([]byte, 16)
	iv := make([]byte, 16)
	result, err := aesCBCDecrypt([]byte{}, key, iv)
	if err != nil {
		t.Fatalf("aesCBCDecrypt empty: %v", err)
	}
	if result != nil {
		t.Errorf("aesCBCDecrypt empty: expected nil, got %v", result)
	}
}

func TestAesCBCDecrypt_InvalidPadding(t *testing.T) {
	// Encrypt something then corrupt the last byte (pad byte) to be 0
	key, iv := deriveKeyIV("testpassword")
	// We'll encrypt one block then flip its last byte to corrupt the padding.
	plain, _ := aesCBCEncrypt([]byte("hello world!!!!"), key, iv)
	// Flip last byte of ciphertext to corrupt padding
	plain[len(plain)-1] ^= 0xFF
	_, err := aesCBCDecrypt(plain, key, iv)
	if err == nil {
		t.Error("aesCBCDecrypt: corrupted ciphertext should produce invalid padding error")
	}
}

func TestAesCBCEncrypt_EmptyData(t *testing.T) {
	key := make([]byte, 16)
	iv := make([]byte, 16)
	// Empty input should produce one block of padding
	result, err := aesCBCEncrypt([]byte{}, key, iv)
	if err != nil {
		t.Fatalf("aesCBCEncrypt empty: %v", err)
	}
	if len(result) != aes.BlockSize {
		t.Errorf("aesCBCEncrypt empty: expected %d bytes, got %d", aes.BlockSize, len(result))
	}
}

func TestAesCBCEncrypt_ExactBlock(t *testing.T) {
	key := make([]byte, 16)
	iv := make([]byte, 16)
	// Exactly 16 bytes — should produce 32 bytes (16 data + 16 padding block)
	result, err := aesCBCEncrypt(make([]byte, 16), key, iv)
	if err != nil {
		t.Fatalf("aesCBCEncrypt exact block: %v", err)
	}
	if len(result) != 32 {
		t.Errorf("aesCBCEncrypt exact block: expected 32 bytes, got %d", len(result))
	}
}

// ── EncryptStream error paths ─────────────────────────────────────────────────

// failWriter fails after N bytes have been written.
type failAfterWriter struct {
	limit   int
	written int
}

func (f *failAfterWriter) Write(p []byte) (int, error) {
	if f.written >= f.limit {
		return 0, errors.New("write failed")
	}
	remaining := f.limit - f.written
	if len(p) > remaining {
		f.written += remaining
		return remaining, errors.New("write failed")
	}
	f.written += len(p)
	return len(p), nil
}

func TestEncryptStream_SignatureWriteError(t *testing.T) {
	// Fail on the very first write (the signature)
	w := &failAfterWriter{limit: 0}
	_, err := EncryptStream(w, "pw")
	if err == nil {
		t.Error("EncryptStream: signature write failure should return error")
	}
}

func TestCipherWriter_WriteError(t *testing.T) {
	// Allow signature (3 bytes) to succeed, then fail on encrypted data write.
	// Need to write at least 16 bytes of plaintext to trigger a block write.
	w := &failAfterWriter{limit: 3} // allow signature, fail on first block
	wc, err := EncryptStream(w, "pw")
	if err != nil {
		t.Fatalf("EncryptStream: %v", err)
	}
	// Write 16 bytes to trigger a block encryption + write
	_, err = wc.Write(make([]byte, 16))
	if err == nil {
		t.Error("cipherWriter.Write: dest write failure should return error")
	}
}

// ── DecryptStream error paths ─────────────────────────────────────────────────

func TestDecryptStream_CorruptedData(t *testing.T) {
	// Data that is not a multiple of 16 — should error
	_, err := DecryptStream(bytes.NewReader([]byte("not16byteslong")), "pw")
	if err == nil {
		t.Error("DecryptStream: non-block-aligned data should return error")
	}
}

func TestDecryptStream_WrongKey(t *testing.T) {
	// Encrypt with one password, decrypt with another — padding will be invalid
	plaintext := []byte("hello world!")
	var buf bytes.Buffer
	wc, _ := EncryptStream(&buf, "correctpassword")
	wc.Write(plaintext) //nolint:errcheck
	wc.Close()          //nolint:errcheck

	// Strip signature
	cipherReader := bytes.NewReader(buf.Bytes()[3:])
	_, err := DecryptStream(cipherReader, "wrongpassword")
	if err == nil {
		t.Error("DecryptStream: wrong key should produce padding error")
	}
}

// ── PeekAndDecrypt error paths ────────────────────────────────────────────────

func TestPeekAndDecrypt_ReaderError(t *testing.T) {
	// Reader that fails immediately (not EOF)
	r := &errAfterReader{failAfter: 0, err: errors.New("read failed")}
	_, _, err := PeekAndDecrypt(r, "pw")
	if err == nil {
		t.Error("PeekAndDecrypt: reader error should propagate")
	}
}

func TestPeekAndDecrypt_CorruptedEncryptedData(t *testing.T) {
	// Stream that starts with "rij" but has corrupted encrypted content
	corrupted := append([]byte("rij"), []byte("notvalidciphertext")...)
	_, _, err := PeekAndDecrypt(bytes.NewReader(corrupted), "pw")
	if err == nil {
		t.Error("PeekAndDecrypt: corrupted encrypted stream should return error")
	}
}

// errAfterReader fails with error after failAfter bytes.
type errAfterReader struct {
	failAfter int
	read      int
	err       error
}

func (r *errAfterReader) Read(p []byte) (int, error) {
	if r.read >= r.failAfter {
		return 0, r.err
	}
	n := r.failAfter - r.read
	if n > len(p) {
		n = len(p)
	}
	r.read += n
	return n, nil
}

// ── EncryptString error paths ─────────────────────────────────────────────────

func TestEncryptString_BothEmpty(t *testing.T) {
	// Both empty → returns plaintext (empty)
	result, err := EncryptString("", "")
	if err != nil {
		t.Fatalf("EncryptString both empty: %v", err)
	}
	if result != "" {
		t.Errorf("EncryptString both empty: expected '', got %q", result)
	}
}

// ── DecryptString error paths ─────────────────────────────────────────────────

func TestDecryptString_CorruptedCiphertext(t *testing.T) {
	// "rij" + valid base64 but not valid AES ciphertext (wrong length)
	// 3 bytes of garbage: base64("abc") = "YWJj"
	_, err := DecryptString("rijYWJj", "password")
	if err == nil {
		t.Error("DecryptString: corrupted ciphertext should return error")
	}
}

// ── cipherWriter.Close write error ───────────────────────────────────────────

func TestCipherWriter_CloseWriteError(t *testing.T) {
	// Write exactly 0 bytes of plaintext then close — Close must flush padding.
	// The dest writer fails on the close flush.
	var countBuf bytes.Buffer
	wc, err := EncryptStream(&countBuf, "pw")
	if err != nil {
		t.Fatalf("EncryptStream: %v", err)
	}
	// Wrap the inner cipherWriter to use a failing dest for Close
	cw := wc.(*cipherWriter)
	cw.dest = &failAfterWriter{limit: 0} // fail immediately
	err = cw.Close()
	if err == nil {
		t.Error("cipherWriter.Close: dest write failure should return error")
	}
}

// ── io.ReadAll error path in DecryptStream ────────────────────────────────────

func TestDecryptStream_ReadError(t *testing.T) {
	r := &errAfterReader{failAfter: 0, err: errors.New("read failed")}
	_, err := DecryptStream(r, "pw")
	if err == nil {
		t.Error("DecryptStream: reader error should propagate")
	}
}

// ── EncryptString / DecryptString round-trip with exact block size plaintext ──

func TestEncryptDecryptString_ExactBlockSize(t *testing.T) {
	// Exactly 16 bytes → 32 byte ciphertext
	plaintext := "1234567890123456"
	enc, err := EncryptString(plaintext, "pw")
	if err != nil {
		t.Fatalf("EncryptString: %v", err)
	}
	dec, err := DecryptString(enc, "pw")
	if err != nil {
		t.Fatalf("DecryptString: %v", err)
	}
	if dec != plaintext {
		t.Errorf("round-trip: got %q, want %q", dec, plaintext)
	}
}

// ── aesCBCEncrypt / aesCBCDecrypt bad-key error paths ────────────────────────
// aes.NewCipher rejects keys that are not 16, 24, or 32 bytes.
// These tests exercise the `return nil, err` branches in aesCBCEncrypt and
// aesCBCDecrypt that are unreachable via the public API (which always derives
// a 16-byte key) but must be covered for 100% statement coverage.

func TestAesCBCEncrypt_BadKeyLength(t *testing.T) {
	// A 5-byte key is invalid for AES; aes.NewCipher must return an error.
	badKey := []byte("short")
	iv := make([]byte, 16)
	_, err := aesCBCEncrypt([]byte("plaintext"), badKey, iv)
	if err == nil {
		t.Error("aesCBCEncrypt: bad key length should return error")
	}
}

func TestAesCBCDecrypt_BadKeyLength(t *testing.T) {
	// Same: a 5-byte key is invalid.
	badKey := []byte("short")
	iv := make([]byte, 16)
	// Use a 16-byte ciphertext (multiple of block size) so we pass the length check.
	ciphertext := make([]byte, 16)
	_, err := aesCBCDecrypt(ciphertext, badKey, iv)
	if err == nil {
		t.Error("aesCBCDecrypt: bad key length should return error")
	}
}

// Ensure io import is used.
var _ = io.EOF
