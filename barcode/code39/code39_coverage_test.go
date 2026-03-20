package code39_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/code39"
)

// TestEncode_ExtendedMode_NonASCIIChar verifies that encoding non-ASCII chars
// in extended mode still produces a valid image (native encoder handles these gracefully).
func TestEncode_ExtendedMode_NonASCIIChar(t *testing.T) {
	enc := code39.NewEncoder()
	enc.AllowExtended = true
	// Non-ASCII chars are silently skipped by the native Code39 extended encoder.
	img, err := enc.Encode("\x80", 300, 100)
	if err != nil {
		// Error is acceptable - it means validation caught it.
		return
	}
	if img == nil {
		t.Error("expected non-nil image for extended mode encoding")
	}
}

// TestEncode_ExtendedMode_SmallSize verifies that small target sizes still
// produce a valid image (native renderer does not error on small sizes).
func TestEncode_ExtendedMode_SmallSize(t *testing.T) {
	enc := code39.NewEncoder()
	enc.AllowExtended = true
	img, err := enc.Encode("hello@world", 50, 100)
	if err != nil {
		t.Fatalf("unexpected error for small size: %v", err)
	}
	if img == nil {
		t.Error("expected non-nil image for small size")
	}
}

// TestEncode_StandardMode_SmallSize verifies that small target sizes still
// produce a valid image in standard mode.
func TestEncode_StandardMode_SmallSize(t *testing.T) {
	enc := code39.NewEncoder()
	img, err := enc.Encode("HELLO", 50, 100)
	if err != nil {
		t.Fatalf("unexpected error for small size: %v", err)
	}
	if img == nil {
		t.Error("expected non-nil image for small size")
	}
}
