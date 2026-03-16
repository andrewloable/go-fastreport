package code2of5

// Internal package tests that inject a scale failure to cover the
// barcode.Scale error path in Encode (line 49 of code2of5.go).

import (
	"errors"
	"testing"

	"github.com/boombuler/barcode"
)

func TestEncode_ScaleError_InternalPath(t *testing.T) {
	// Replace scaleBarcode with a function that always returns an error.
	orig := scaleBarcode
	scaleBarcode = func(_ barcode.Barcode, _, _ int) (barcode.Barcode, error) {
		return nil, errors.New("injected scale failure")
	}
	defer func() { scaleBarcode = orig }()

	enc := New()
	_, err := enc.Encode("1234", 200, 100)
	if err == nil {
		t.Error("expected error from injected scale failure, got nil")
	}
}
