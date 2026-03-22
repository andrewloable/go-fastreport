// Package utils — crypto_murmur3_test.go tests the Murmur3-based ComputeHash
// functions and the default-password helpers ported from Crypter.cs.
package utils

import (
	"bytes"
	"strings"
	"testing"
)

// TestComputeHashString_empty verifies that an empty string produces the
// all-zero 32-character hex fingerprint (Murmur3 with seed=0, length=0).
func TestComputeHashString_empty(t *testing.T) {
	got := ComputeHashString("")
	const want = "00000000000000000000000000000000"
	if got != want {
		t.Errorf("ComputeHashString(\"\") = %q, want %q", got, want)
	}
}

// TestComputeHashBytes_known verifies the hash of "hello" is 32 hex chars,
// uppercase, no dashes, and matches the known C#-compatible value.
func TestComputeHashBytes_known(t *testing.T) {
	got := ComputeHashBytes([]byte("hello"))
	const want = "029BBD41B3A7D8CB191DAE486A901E5B"

	if len(got) != 32 {
		t.Errorf("ComputeHashBytes(\"hello\") length = %d, want 32", len(got))
	}
	if strings.Contains(got, "-") {
		t.Errorf("ComputeHashBytes(\"hello\") contains dash: %q", got)
	}
	if got != strings.ToUpper(got) {
		t.Errorf("ComputeHashBytes(\"hello\") is not uppercase: %q", got)
	}
	if got != want {
		t.Errorf("ComputeHashBytes(\"hello\") = %q, want %q", got, want)
	}
}

// TestComputeHashReader verifies that ComputeHashReader produces the same
// result as ComputeHashBytes for the same input bytes ("test").
func TestComputeHashReader(t *testing.T) {
	input := []byte("test")
	want := ComputeHashBytes(input)

	got, err := ComputeHashReader(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("ComputeHashReader: unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("ComputeHashReader(\"test\") = %q, want %q", got, want)
	}
}

// TestDefaultPassword verifies the initial default password matches the C#
// typeof(Crypter).FullName value, and that SetDefaultPassword + the Default
// encrypt/decrypt helpers function correctly after changing it.
func TestDefaultPassword(t *testing.T) {
	// Save and restore the original value so other tests are not affected.
	original := defaultCrypterPassword
	defer func() { defaultCrypterPassword = original }()

	// Initial value must match C# typeof(Crypter).FullName.
	if defaultCrypterPassword != "FastReport.Utils.Crypter" {
		t.Errorf("initial defaultCrypterPassword = %q, want %q",
			defaultCrypterPassword, "FastReport.Utils.Crypter")
	}

	// Change the default password and verify encrypt/decrypt round-trip.
	const newPwd = "MyCustomPassword"
	SetDefaultPassword(newPwd)

	const plaintext = "Hello, World!"
	encrypted, err := EncryptStringDefault(plaintext)
	if err != nil {
		t.Fatalf("EncryptStringDefault: %v", err)
	}
	if !IsStringEncrypted(encrypted) {
		t.Errorf("EncryptStringDefault result not marked as encrypted: %q", encrypted)
	}

	decrypted, err := DecryptStringDefault(encrypted)
	if err != nil {
		t.Fatalf("DecryptStringDefault: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("DecryptStringDefault = %q, want %q", decrypted, plaintext)
	}
}
