package aztec_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/aztec"
)

// TestEncode_InvalidLayers_TooHigh covers the error path in Encode when
// UserSpecifiedLayers > 32, which causes the underlying aztec.Encode to
// return "Illegal value N for layers".
func TestEncode_InvalidLayers_TooHigh(t *testing.T) {
	e := aztec.New()
	e.UserSpecifiedLayers = 33
	_, err := e.Encode("hello", 100)
	if err == nil {
		t.Error("expected error for UserSpecifiedLayers=33 (must be 1–32)")
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

// TestEncode_ScaleTooSmall covers the barcode.Scale error path by requesting
// a canvas smaller than the barcode's native minimum pixel size.
// An Aztec barcode for short text is ≥15×15 pixels. Requesting size=1
// will pass our size > 0 guard but fail in barcode.Scale.
func TestEncode_ScaleTooSmall(t *testing.T) {
	e := aztec.New()
	_, err := e.Encode("x", 1)
	if err == nil {
		t.Error("expected error when scale target is smaller than barcode minimum size")
	}
}
