package code93_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/code93"
)

// TestEncode_LowercaseStandardMode_LibraryError covers the error path from
// the underlying code93.Encode library call (lines 43–45).
// Code 93 standard mode only supports uppercase A–Z, digits, and a few specials.
// Lowercase letters are rejected with "invalid data!". Our wrapper has no
// Validate guard, so lowercase passes straight through to the library.
func TestEncode_LowercaseStandardMode_LibraryError(t *testing.T) {
	enc := code93.New()
	enc.FullASCIIMode = false
	_, err := enc.Encode("hello", 200, 100)
	if err == nil {
		t.Error("expected error for lowercase text in Code 93 standard mode")
	}
}

// TestEncode_InvalidCharStandard_LibraryError covers the library error for
// characters not in the Code 93 standard character set (e.g., '@').
func TestEncode_InvalidCharStandard_LibraryError(t *testing.T) {
	enc := code93.New()
	enc.FullASCIIMode = false
	_, err := enc.Encode("A@B", 200, 100)
	if err == nil {
		t.Error("expected error for '@' in Code 93 standard mode")
	}
}

// TestEncode_NonASCIIFullASCIIMode_LibraryError covers the library error
// for non-ASCII characters even in full-ASCII mode (lines 43–45).
// The boombuler code93 encoder rejects chars > 0x7F in full-ASCII mode.
func TestEncode_NonASCIIFullASCIIMode_LibraryError(t *testing.T) {
	enc := code93.New()
	enc.FullASCIIMode = true
	_, err := enc.Encode("\xff", 200, 100)
	if err == nil {
		t.Error("expected error for non-ASCII char \\xff in Code 93 full-ASCII mode")
	}
}

// TestEncode_ScaleTooSmall_Standard covers the barcode.Scale error path
// (lines 47–49) when the requested width is smaller than the barcode's native
// pixel width. "HELLO" in Code 93 is 73 pixels wide; requesting width=50
// passes our positive-width guard but fails in barcode.Scale.
func TestEncode_ScaleTooSmall_Standard(t *testing.T) {
	enc := code93.New()
	enc.FullASCIIMode = false
	_, err := enc.Encode("HELLO", 50, 100)
	if err == nil {
		t.Error("expected error when scale width < native barcode width")
	}
}

// TestEncode_ScaleTooSmall_FullASCII covers the Scale error path in full-ASCII
// mode. "Hello" in full-ASCII mode produces a wider barcode (~118px); width=50
// fails in Scale.
func TestEncode_ScaleTooSmall_FullASCII(t *testing.T) {
	enc := code93.New()
	enc.FullASCIIMode = true
	_, err := enc.Encode("Hello", 50, 100)
	if err == nil {
		t.Error("expected error when scale width < native barcode width (full-ASCII mode)")
	}
}
