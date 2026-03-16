package upc

// Internal package tests that inject failures to cover the unreachable error
// paths in Encode:
//   - boomean.Encode error path (line 50-51 of upc.go)
//   - barcode.Scale error path (line 54-55 of upc.go)

import (
	"errors"
	"testing"

	"github.com/boombuler/barcode"
)

func TestEncode_EANEncodeError_InternalPath(t *testing.T) {
	// Replace encodeEAN with a function that always returns an error.
	origEncode := encodeEAN
	encodeEAN = func(_ string) (barcode.BarcodeIntCS, error) {
		return nil, errors.New("injected ean encode failure")
	}
	defer func() { encodeEAN = origEncode }()

	enc := New()
	_, err := enc.Encode("01234567890", 200, 100)
	if err == nil {
		t.Error("expected error from injected ean encode failure, got nil")
	}
}

func TestEncode_ScaleError_InternalPath(t *testing.T) {
	// Replace scaleBarcode with a function that always returns an error.
	orig := scaleBarcode
	scaleBarcode = func(_ barcode.Barcode, _, _ int) (barcode.Barcode, error) {
		return nil, errors.New("injected scale failure")
	}
	defer func() { scaleBarcode = orig }()

	enc := New()
	_, err := enc.Encode("01234567890", 200, 100)
	if err == nil {
		t.Error("expected error from injected scale failure, got nil")
	}
}
