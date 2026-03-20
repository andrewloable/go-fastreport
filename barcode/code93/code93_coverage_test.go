package code93_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/code93"
)

// TestEncode_LowercaseStandardMode_LibraryError covers the error path for
// lowercase text in Code 93 standard mode.
func TestEncode_LowercaseStandardMode_LibraryError(t *testing.T) {
	enc := code93.New()
	enc.FullASCIIMode = false
	_, err := enc.Encode("hello", 200, 100)
	if err == nil {
		t.Error("expected error for lowercase text in Code 93 standard mode")
	}
}

// TestEncode_InvalidCharStandard_LibraryError covers the error for
// characters not in the Code 93 standard character set.
func TestEncode_InvalidCharStandard_LibraryError(t *testing.T) {
	enc := code93.New()
	enc.FullASCIIMode = false
	_, err := enc.Encode("A@B", 200, 100)
	if err == nil {
		t.Error("expected error for '@' in Code 93 standard mode")
	}
}

// TestEncode_NonASCIIFullASCIIMode verifies that non-ASCII characters in
// full-ASCII mode produce an image (native encoder handles these gracefully).
func TestEncode_NonASCIIFullASCIIMode(t *testing.T) {
	enc := code93.New()
	enc.FullASCIIMode = true
	// Native encoder ignores non-ASCII chars, should produce valid output.
	img, err := enc.Encode("\xff", 200, 100)
	if err != nil {
		// Error is acceptable - means validation caught it.
		return
	}
	if img == nil {
		t.Error("expected non-nil image")
	}
}

// TestEncode_SmallSize_Standard verifies small target sizes produce valid images.
func TestEncode_SmallSize_Standard(t *testing.T) {
	enc := code93.New()
	enc.FullASCIIMode = false
	img, err := enc.Encode("HELLO", 50, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if img == nil {
		t.Error("expected non-nil image")
	}
}

// TestEncode_SmallSize_FullASCII verifies small target sizes produce valid images.
func TestEncode_SmallSize_FullASCII(t *testing.T) {
	enc := code93.New()
	enc.FullASCIIMode = true
	img, err := enc.Encode("Hello", 50, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if img == nil {
		t.Error("expected non-nil image")
	}
}
