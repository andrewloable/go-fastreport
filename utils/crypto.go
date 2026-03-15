// Package utils — crypto.go implements AES encryption/decryption compatible
// with the FastReport .NET Crypter class (Utils/Crypter.cs).
//
// FastReport uses AES-128-CBC with ISO10126 padding.
// The key and IV are derived via a .NET PasswordDeriveBytes-compatible PBKDF1:
//
//	base = SHA1^100(password || salt)        — 20 bytes (SHA-1 output)
//	key  = base[0:16]
//	iv   = base[16:20] || SHA1("1" || base || password || salt)[0:12]
//
// Encrypted FRX streams are prefixed with the 3-byte "rij" signature (114 105 106).
// Encrypted strings are prefixed with the literal string "rij" followed by the
// base64-encoded ciphertext.
package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// rijSignature is the 3-byte magic prefix on encrypted FRX streams.
var rijSignature = []byte{114, 105, 106} // "rij"

// rijPrefix is the literal string prefix on encrypted connection strings.
const rijPrefix = "rij"

// IsStreamEncrypted reports whether the stream begins with the "rij" signature.
// It reads 3 bytes from r; the caller must not rewind r afterwards — use
// DetachOrDecrypt instead if transparent decryption is needed.
func IsStreamEncrypted(r io.Reader) (bool, error) {
	sig := make([]byte, 3)
	_, err := io.ReadFull(r, sig)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			return false, nil
		}
		return false, err
	}
	return bytes.Equal(sig, rijSignature), nil
}

// IsStringEncrypted reports whether data begins with the "rij" prefix.
func IsStringEncrypted(data string) bool {
	return len(data) >= 3 && data[:3] == rijPrefix
}

// EncryptString encrypts plaintext using the given password and returns
// "rij" + base64(ciphertext), matching C# Crypter.EncryptString.
func EncryptString(plaintext, password string) (string, error) {
	if plaintext == "" || password == "" {
		return plaintext, nil
	}
	key, iv := deriveKeyIV(password)
	ciphertext, err := aesCBCEncrypt([]byte(plaintext), key, iv)
	if err != nil {
		return "", err
	}
	return rijPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts a "rij"-prefixed base64 string produced by
// EncryptString (or the C# Crypter.EncryptString equivalent).
// Returns plaintext unchanged if not encrypted or either argument is empty.
func DecryptString(data, password string) (string, error) {
	if data == "" || password == "" || !IsStringEncrypted(data) {
		return data, nil
	}
	raw, err := base64.StdEncoding.DecodeString(data[3:])
	if err != nil {
		return "", fmt.Errorf("crypto: base64 decode: %w", err)
	}
	key, iv := deriveKeyIV(password)
	plain, err := aesCBCDecrypt(raw, key, iv)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// EncryptStream wraps dest so that writes are AES-128-CBC encrypted.
// It writes the "rij" signature to dest before returning the writer.
// The caller must Close the returned WriteCloser to flush the final block.
func EncryptStream(dest io.Writer, password string) (io.WriteCloser, error) {
	if _, err := dest.Write(rijSignature); err != nil {
		return nil, err
	}
	key, iv := deriveKeyIV(password)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &cipherWriter{
		mode: cipher.NewCBCEncrypter(block, iv),
		dest: dest,
		buf:  bytes.NewBuffer(nil),
	}, nil
}

// DecryptStream reads from source assuming AES-128-CBC encryption and the
// "rij" signature has already been consumed. Use PeekAndDecrypt for
// transparent handling.
func DecryptStream(source io.Reader, password string) (io.Reader, error) {
	key, iv := deriveKeyIV(password)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	all, err := io.ReadAll(source)
	if err != nil {
		return nil, err
	}
	plain, err := aesCBCDecrypt(all, key, iv)
	if err != nil {
		return nil, err
	}
	_ = block
	return bytes.NewReader(plain), nil
}

// PeekAndDecrypt reads the first 3 bytes from r. If they are the "rij"
// signature, it decrypts the remainder using password and returns the
// plaintext reader. Otherwise it returns a reader that prepends the 3 bytes
// back and returns all original content unmodified.
func PeekAndDecrypt(r io.Reader, password string) (io.Reader, bool, error) {
	sig := make([]byte, 3)
	n, err := io.ReadFull(r, sig)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		if errors.Is(err, io.EOF) {
			// Empty stream.
			return bytes.NewReader(nil), false, nil
		}
		return nil, false, err
	}
	sig = sig[:n]
	if n == 3 && bytes.Equal(sig, rijSignature) {
		// Encrypted: decrypt the rest.
		plain, err := DecryptStream(r, password)
		if err != nil {
			return nil, false, err
		}
		return plain, true, nil
	}
	// Not encrypted: prepend the bytes back.
	return io.MultiReader(bytes.NewReader(sig), r), false, nil
}

// ── Key derivation ────────────────────────────────────────────────────────────

// deriveKeyIV returns the 16-byte AES key and 16-byte IV derived from password
// using the .NET PasswordDeriveBytes algorithm (PBKDF1-extended with SHA-1,
// 100 iterations, salt = UTF-8("Salt")).
func deriveKeyIV(password string) (key, iv []byte) {
	passwordBytes := []byte(password)
	saltBytes := []byte("Salt")
	const iterations = 100

	// Step 1: base = SHA1^100(password || salt)
	h := sha1.New()
	h.Write(passwordBytes)
	h.Write(saltBytes)
	base := h.Sum(nil) // 20 bytes
	for i := 1; i < iterations; i++ {
		h.Reset()
		h.Write(base)
		base = h.Sum(nil)
	}

	// Step 2: derived stream via counter blocks.
	// The state is shared across the two GetBytes(16) calls.
	derived := make([]byte, 0, 40)
	derived = append(derived, base...) // first 20 bytes = base
	counter := 1
	for len(derived) < 32 {
		// Next block: SHA1(counter_string || base || password || salt)
		h.Reset()
		h.Write([]byte(strconv.Itoa(counter)))
		h.Write(base)
		h.Write(passwordBytes)
		h.Write(saltBytes)
		derived = append(derived, h.Sum(nil)...)
		counter++
	}

	key = derived[0:16]
	iv = derived[16:32]
	return
}

// ── AES-CBC helpers ───────────────────────────────────────────────────────────

// aesCBCEncrypt encrypts plaintext using AES-128-CBC with PKCS7 padding.
// (ISO10126 is accepted by .NET when decrypting PKCS7 — the last byte is identical.)
func aesCBCEncrypt(plaintext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padded := pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)
	return ciphertext, nil
}

// aesCBCDecrypt decrypts AES-128-CBC ciphertext with PKCS7/ISO10126 un-padding.
func aesCBCDecrypt(ciphertext, key, iv []byte) ([]byte, error) {
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("crypto: ciphertext length %d is not a multiple of AES block size", len(ciphertext))
	}
	if len(ciphertext) == 0 {
		return nil, nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, ciphertext)
	return pkcs7Unpad(plain)
}

// pkcs7Pad pads data to a multiple of blockSize using PKCS7.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padded := make([]byte, len(data)+padding)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}
	return padded
}

// pkcs7Unpad removes PKCS7 / ISO10126 padding (both are validated the same way).
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > aes.BlockSize || pad > len(data) {
		return nil, fmt.Errorf("crypto: invalid padding byte %d", pad)
	}
	return data[:len(data)-pad], nil
}

// ── cipherWriter ─────────────────────────────────────────────────────────────

// cipherWriter accumulates plaintext and encrypts complete blocks.
type cipherWriter struct {
	mode cipher.BlockMode
	dest io.Writer
	buf  *bytes.Buffer
}

func (w *cipherWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	blockSize := w.mode.BlockSize()
	for w.buf.Len() >= blockSize {
		block := make([]byte, blockSize)
		w.buf.Read(block) //nolint:errcheck
		encrypted := make([]byte, blockSize)
		w.mode.CryptBlocks(encrypted, block)
		if _, err := w.dest.Write(encrypted); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// Close flushes any remaining buffered plaintext with PKCS7 padding.
func (w *cipherWriter) Close() error {
	remaining := w.buf.Bytes()
	padded := pkcs7Pad(remaining, w.mode.BlockSize())
	encrypted := make([]byte, len(padded))
	w.mode.CryptBlocks(encrypted, padded)
	_, err := w.dest.Write(encrypted)
	return err
}
