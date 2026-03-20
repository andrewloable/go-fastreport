package ean

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

// TestEncode_EAN8_LibraryEncodeError exercises the error path when the EAN-8
// library's Encode returns an error.
func TestEncode_EAN8_LibraryEncodeError(t *testing.T) {
	origFactory := newEAN8Barcode
	defer func() { newEAN8Barcode = origFactory }()

	newEAN8Barcode = func() barcodeEncoder {
		return &mockBarcode{encodeErr: fmt.Errorf("mock ean8 encode failure")}
	}

	enc := New()
	_, err := enc.Encode("1234567", 200, 100)
	if err == nil {
		t.Fatal("expected error from EAN-8 library Encode, got nil")
	}
	if got := err.Error(); got != "ean: mock ean8 encode failure" {
		t.Errorf("unexpected error message: %s", got)
	}
}

// TestEncode_EAN8_LibraryRenderError exercises the error path when the EAN-8
// library's Render returns an error.
func TestEncode_EAN8_LibraryRenderError(t *testing.T) {
	origFactory := newEAN8Barcode
	defer func() { newEAN8Barcode = origFactory }()

	newEAN8Barcode = func() barcodeEncoder {
		return &mockBarcode{renderErr: fmt.Errorf("mock ean8 render failure")}
	}

	enc := New()
	_, err := enc.Encode("1234567", 200, 100)
	if err == nil {
		t.Fatal("expected error from EAN-8 library Render, got nil")
	}
	if got := err.Error(); got != "ean: render mock ean8 render failure" {
		t.Errorf("unexpected error message: %s", got)
	}
}

// TestEncode_EAN13_LibraryEncodeError exercises the error path when the EAN-13
// library's Encode returns an error.
func TestEncode_EAN13_LibraryEncodeError(t *testing.T) {
	origFactory := newEAN13Barcode
	defer func() { newEAN13Barcode = origFactory }()

	newEAN13Barcode = func() barcodeEncoder {
		return &mockBarcode{encodeErr: fmt.Errorf("mock ean13 encode failure")}
	}

	enc := New()
	_, err := enc.Encode("123456789012", 200, 100)
	if err == nil {
		t.Fatal("expected error from EAN-13 library Encode, got nil")
	}
	if got := err.Error(); got != "ean: mock ean13 encode failure" {
		t.Errorf("unexpected error message: %s", got)
	}
}

// TestEncode_EAN13_LibraryRenderError exercises the error path when the EAN-13
// library's Render returns an error.
func TestEncode_EAN13_LibraryRenderError(t *testing.T) {
	origFactory := newEAN13Barcode
	defer func() { newEAN13Barcode = origFactory }()

	newEAN13Barcode = func() barcodeEncoder {
		return &mockBarcode{renderErr: fmt.Errorf("mock ean13 render failure")}
	}

	enc := New()
	_, err := enc.Encode("123456789012", 200, 100)
	if err == nil {
		t.Fatal("expected error from EAN-13 library Render, got nil")
	}
	if got := err.Error(); got != "ean: render mock ean13 render failure" {
		t.Errorf("unexpected error message: %s", got)
	}
}
