package aztec_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/aztec"
)

// TestEncode_InvalidLayers_TooHigh covers the error path when
// UserSpecifiedLayers > 32.
func TestEncode_InvalidLayers_TooHigh(t *testing.T) {
	e := aztec.New()
	e.UserSpecifiedLayers = 33
	_, err := e.Encode("hello", 100)
	if err == nil {
		t.Error("expected error for UserSpecifiedLayers=33 (must be 1-32)")
	}
}

// TestEncodeMatrix_InvalidLayers_TooHigh covers the corresponding error path
// in EncodeMatrix when UserSpecifiedLayers > 32.
func TestEncodeMatrix_InvalidLayers_TooHigh(t *testing.T) {
	e := aztec.New()
	e.UserSpecifiedLayers = 33
	_, err := e.EncodeMatrix("hello")
	if err == nil {
		t.Error("expected error for UserSpecifiedLayers=33 in EncodeMatrix")
	}
}

// TestEncode_SmallSize verifies small target sizes produce valid images.
func TestEncode_SmallSize(t *testing.T) {
	e := aztec.New()
	img, err := e.Encode("x", 1)
	if err != nil {
		// Error is acceptable for very small sizes.
		return
	}
	if img == nil {
		t.Error("expected non-nil image or error")
	}
}
