package upc

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

// TestEncode_LibraryEncodeError exercises the error path when the underlying
// UPC-A library's Encode returns an error.
func TestEncode_LibraryEncodeError(t *testing.T) {
	origFactory := newBarcode
	defer func() { newBarcode = origFactory }()

	newBarcode = func() barcodeEncoder {
		return &mockBarcode{encodeErr: fmt.Errorf("mock encode failure")}
	}

	enc := New()
	_, err := enc.Encode("01234567890", 200, 100)
	if err == nil {
		t.Fatal("expected error from library Encode, got nil")
	}
	if got := err.Error(); got != "upc: mock encode failure" {
		t.Errorf("unexpected error message: %s", got)
	}
}

// TestEncode_LibraryRenderError exercises the error path when the underlying
// UPC-A library's Render returns an error.
func TestEncode_LibraryRenderError(t *testing.T) {
	origFactory := newBarcode
	defer func() { newBarcode = origFactory }()

	newBarcode = func() barcodeEncoder {
		return &mockBarcode{renderErr: fmt.Errorf("mock render failure")}
	}

	enc := New()
	_, err := enc.Encode("01234567890", 200, 100)
	if err == nil {
		t.Fatal("expected error from library Render, got nil")
	}
	if got := err.Error(); got != "upc: render mock render failure" {
		t.Errorf("unexpected error message: %s", got)
	}
}
