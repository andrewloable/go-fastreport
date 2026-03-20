package code93

import (
	"fmt"
	"image"
	"testing"
)

// mockBarcode is a test double that can be configured to return errors.
type mockBarcode struct {
	encodeErr error
	renderErr error
}

func (m *mockBarcode) Encode(text string) error { return m.encodeErr }
func (m *mockBarcode) Render(width, height int) (image.Image, error) {
	if m.renderErr != nil {
		return nil, m.renderErr
	}
	return image.NewRGBA(image.Rect(0, 0, width, height)), nil
}

// TestEncode_StandardEncodeError exercises the error path when the standard
// Code93 barcode's Encode returns an error.
func TestEncode_StandardEncodeError(t *testing.T) {
	origFactory := newStandardBarcode
	defer func() { newStandardBarcode = origFactory }()

	newStandardBarcode = func(includeChecksum, fullASCII bool) barcodeEncoder {
		return &mockBarcode{encodeErr: fmt.Errorf("mock standard encode failure")}
	}

	enc := New()
	enc.FullASCIIMode = false
	_, err := enc.Encode("HELLO", 200, 100)
	if err == nil {
		t.Fatal("expected error from standard Encode, got nil")
	}
	if got := err.Error(); got != "code93: mock standard encode failure" {
		t.Errorf("unexpected error message: %s", got)
	}
}

// TestEncode_ExtendedEncodeError exercises the error path when the extended
// Code93 barcode's Encode returns an error.
func TestEncode_ExtendedEncodeError(t *testing.T) {
	origFactory := newExtendedBarcode
	defer func() { newExtendedBarcode = origFactory }()

	newExtendedBarcode = func() barcodeEncoder {
		return &mockBarcode{encodeErr: fmt.Errorf("mock extended encode failure")}
	}

	enc := New()
	enc.FullASCIIMode = true
	_, err := enc.Encode("Hello!", 600, 100)
	if err == nil {
		t.Fatal("expected error from extended Encode, got nil")
	}
	if got := err.Error(); got != "code93: mock extended encode failure" {
		t.Errorf("unexpected error message: %s", got)
	}
}

// TestEncode_RenderError exercises the error path when Render returns an error.
func TestEncode_RenderError(t *testing.T) {
	origFactory := newStandardBarcode
	defer func() { newStandardBarcode = origFactory }()

	newStandardBarcode = func(includeChecksum, fullASCII bool) barcodeEncoder {
		return &mockBarcode{renderErr: fmt.Errorf("mock render failure")}
	}

	enc := New()
	enc.FullASCIIMode = false
	_, err := enc.Encode("HELLO", 200, 100)
	if err == nil {
		t.Fatal("expected error from Render, got nil")
	}
	if got := err.Error(); got != "code93: render mock render failure" {
		t.Errorf("unexpected error message: %s", got)
	}
}
