package code128_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/code128"
)

// TestEncode_NonASCIIChar verifies that non-ASCII characters are handled.
// The native Code128 encoder may produce output or return an error.
func TestEncode_NonASCIIChar(t *testing.T) {
	enc := code128.NewEncoder()
	img, err := enc.Encode("\x80test", 300, 100)
	if err != nil {
		// Error is acceptable - native encoder may reject non-ASCII.
		return
	}
	if img == nil {
		t.Error("expected non-nil image or error")
	}
}

// TestEncode_SmallWidth verifies that small target widths produce valid images.
func TestEncode_SmallWidth(t *testing.T) {
	enc := code128.NewEncoder()
	img, err := enc.Encode("hello", 50, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if img == nil {
		t.Error("expected non-nil image")
	}
}
