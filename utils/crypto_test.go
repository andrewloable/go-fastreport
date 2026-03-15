package utils

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// ── IsStreamEncrypted ─────────────────────────────────────────────────────────

func TestIsStreamEncrypted_Encrypted(t *testing.T) {
	r := bytes.NewReader(rijSignature)
	ok, err := IsStreamEncrypted(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected true for rij-prefixed stream")
	}
}

func TestIsStreamEncrypted_NotEncrypted(t *testing.T) {
	r := bytes.NewReader([]byte("<?xml"))
	ok, err := IsStreamEncrypted(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false for non-encrypted stream")
	}
}

func TestIsStreamEncrypted_TooShort(t *testing.T) {
	r := bytes.NewReader([]byte("ri")) // only 2 bytes
	ok, err := IsStreamEncrypted(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false for stream shorter than 3 bytes")
	}
}

func TestIsStreamEncrypted_Empty(t *testing.T) {
	r := bytes.NewReader([]byte{})
	ok, err := IsStreamEncrypted(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false for empty stream")
	}
}

func TestIsStreamEncrypted_ReaderError(t *testing.T) {
	r := &errorReader{err: errors.New("read fail")}
	_, err := IsStreamEncrypted(r)
	if err == nil {
		t.Error("expected error from failing reader")
	}
}

// ── IsStringEncrypted ─────────────────────────────────────────────────────────

func TestIsStringEncrypted_Encrypted(t *testing.T) {
	if !IsStringEncrypted("rijABCDEF") {
		t.Error("expected true for rij-prefixed string")
	}
}

func TestIsStringEncrypted_NotEncrypted(t *testing.T) {
	if IsStringEncrypted("hello") {
		t.Error("expected false for non-rij string")
	}
}

func TestIsStringEncrypted_TooShort(t *testing.T) {
	if IsStringEncrypted("ri") {
		t.Error("expected false for string shorter than 3 chars")
	}
}

func TestIsStringEncrypted_ExactlyThree(t *testing.T) {
	if !IsStringEncrypted("rij") {
		t.Error("expected true for exactly 'rij'")
	}
}

// ── EncryptString / DecryptString ─────────────────────────────────────────────

func TestEncryptDecryptString_RoundTrip(t *testing.T) {
	cases := []struct {
		plaintext string
		password  string
	}{
		{"hello world", "password123"},
		{"FastReport data", "secret"},
		{"single", "pw"},
		{strings.Repeat("a", 100), "longpassword"},
	}
	for _, tc := range cases {
		enc, err := EncryptString(tc.plaintext, tc.password)
		if err != nil {
			t.Fatalf("EncryptString(%q): %v", tc.plaintext, err)
		}
		if !IsStringEncrypted(enc) {
			t.Errorf("encrypted string should start with 'rij', got %q", enc[:3])
		}
		dec, err := DecryptString(enc, tc.password)
		if err != nil {
			t.Fatalf("DecryptString: %v", err)
		}
		if dec != tc.plaintext {
			t.Errorf("round-trip mismatch: got %q, want %q", dec, tc.plaintext)
		}
	}
}

func TestEncryptString_EmptyPlaintext(t *testing.T) {
	result, err := EncryptString("", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("empty plaintext should return empty, got %q", result)
	}
}

func TestEncryptString_EmptyPassword(t *testing.T) {
	result, err := EncryptString("hello", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("empty password should return plaintext unchanged, got %q", result)
	}
}

func TestDecryptString_NotEncrypted(t *testing.T) {
	result, err := DecryptString("plain text", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "plain text" {
		t.Errorf("unencrypted string should be returned as-is, got %q", result)
	}
}

func TestDecryptString_EmptyData(t *testing.T) {
	result, err := DecryptString("", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("empty data should return empty, got %q", result)
	}
}

func TestDecryptString_EmptyPassword(t *testing.T) {
	result, err := DecryptString("rijABCD", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "rijABCD" {
		t.Errorf("empty password should return data unchanged, got %q", result)
	}
}

func TestDecryptString_BadBase64(t *testing.T) {
	_, err := DecryptString("rij!!!not-base64!!!", "password")
	if err == nil {
		t.Error("expected error for bad base64 payload")
	}
}

// ── EncryptStream / DecryptStream ─────────────────────────────────────────────

func TestEncryptDecryptStream_RoundTrip(t *testing.T) {
	plaintext := []byte("stream data to encrypt and decrypt")

	var encrypted bytes.Buffer
	wc, err := EncryptStream(&encrypted, "streampassword")
	if err != nil {
		t.Fatalf("EncryptStream: %v", err)
	}
	if _, err := wc.Write(plaintext); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := wc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Verify rij prefix
	if !bytes.HasPrefix(encrypted.Bytes(), rijSignature) {
		t.Error("encrypted stream should start with rij signature")
	}

	// Strip the signature before decrypting
	cipherReader := bytes.NewReader(encrypted.Bytes()[3:])
	plain, err := DecryptStream(cipherReader, "streampassword")
	if err != nil {
		t.Fatalf("DecryptStream: %v", err)
	}
	result, err := io.ReadAll(plain)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if !bytes.Equal(result, plaintext) {
		t.Errorf("round-trip mismatch: got %q, want %q", result, plaintext)
	}
}

func TestEncryptStream_WritesToDest(t *testing.T) {
	var dest bytes.Buffer
	wc, err := EncryptStream(&dest, "pw")
	if err != nil {
		t.Fatalf("EncryptStream: %v", err)
	}
	wc.Write([]byte("hello world encryption test")) //nolint:errcheck
	wc.Close()                                       //nolint:errcheck
	if dest.Len() == 0 {
		t.Error("encrypted stream should have written bytes")
	}
}

func TestDecryptStream_Empty(t *testing.T) {
	plain, err := DecryptStream(bytes.NewReader(nil), "pw")
	if err != nil {
		t.Fatalf("DecryptStream(empty): %v", err)
	}
	data, _ := io.ReadAll(plain)
	if len(data) != 0 {
		t.Errorf("empty ciphertext should yield empty plaintext, got %d bytes", len(data))
	}
}

// ── PeekAndDecrypt ────────────────────────────────────────────────────────────

func TestPeekAndDecrypt_Encrypted(t *testing.T) {
	plaintext := []byte("secret message")
	var buf bytes.Buffer
	wc, _ := EncryptStream(&buf, "pw")
	wc.Write(plaintext) //nolint:errcheck
	wc.Close()          //nolint:errcheck

	reader, encrypted, err := PeekAndDecrypt(bytes.NewReader(buf.Bytes()), "pw")
	if err != nil {
		t.Fatalf("PeekAndDecrypt: %v", err)
	}
	if !encrypted {
		t.Error("expected encrypted=true")
	}
	got, _ := io.ReadAll(reader)
	if !bytes.Equal(got, plaintext) {
		t.Errorf("decrypted = %q, want %q", got, plaintext)
	}
}

func TestPeekAndDecrypt_NotEncrypted(t *testing.T) {
	data := []byte("plain xml content here")
	reader, encrypted, err := PeekAndDecrypt(bytes.NewReader(data), "pw")
	if err != nil {
		t.Fatalf("PeekAndDecrypt: %v", err)
	}
	if encrypted {
		t.Error("expected encrypted=false")
	}
	got, _ := io.ReadAll(reader)
	if !bytes.Equal(got, data) {
		t.Errorf("non-encrypted: got %q, want %q", got, data)
	}
}

func TestPeekAndDecrypt_Empty(t *testing.T) {
	reader, encrypted, err := PeekAndDecrypt(bytes.NewReader(nil), "pw")
	if err != nil {
		t.Fatalf("PeekAndDecrypt(empty): %v", err)
	}
	if encrypted {
		t.Error("empty stream should not be encrypted")
	}
	got, _ := io.ReadAll(reader)
	if len(got) != 0 {
		t.Errorf("expected empty output, got %d bytes", len(got))
	}
}

func TestPeekAndDecrypt_Short(t *testing.T) {
	// Less than 3 bytes — not encrypted
	reader, encrypted, err := PeekAndDecrypt(bytes.NewReader([]byte("hi")), "pw")
	if err != nil {
		t.Fatalf("PeekAndDecrypt(short): %v", err)
	}
	if encrypted {
		t.Error("short stream should not be encrypted")
	}
	got, _ := io.ReadAll(reader)
	if string(got) != "hi" {
		t.Errorf("short non-encrypted: got %q, want 'hi'", got)
	}
}

// ── cipherWriter multi-block write ────────────────────────────────────────────

func TestCipherWriter_MultiBlockWrite(t *testing.T) {
	// Write more than one AES block (16 bytes) to ensure the flush-per-block path runs
	plaintext := bytes.Repeat([]byte("ABCDEFGH"), 10) // 80 bytes = 5 blocks
	var dest bytes.Buffer
	wc, err := EncryptStream(&dest, "blocktest")
	if err != nil {
		t.Fatalf("EncryptStream: %v", err)
	}
	if _, err := wc.Write(plaintext); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := wc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Decrypt and verify
	cipherReader := bytes.NewReader(dest.Bytes()[3:])
	plain, err := DecryptStream(cipherReader, "blocktest")
	if err != nil {
		t.Fatalf("DecryptStream: %v", err)
	}
	got, _ := io.ReadAll(plain)
	if !bytes.Equal(got, plaintext) {
		t.Errorf("multi-block round-trip failed")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// errorReader always returns an error on Read.
type errorReader struct{ err error }

func (e *errorReader) Read(p []byte) (int, error) { return 0, e.err }
