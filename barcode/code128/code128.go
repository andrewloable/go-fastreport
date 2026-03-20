// Package code128 provides Code 128 barcode generation for go-fastreport.
// Uses the native Go Code128 encoder from the parent barcode package.
package code128

import (
	"fmt"
	"image"

	"github.com/andrewloable/go-fastreport/barcode"
)

// Encoder generates Code 128 barcode images from text content.
type Encoder struct{}

// NewEncoder creates a Code128 Encoder.
func NewEncoder() *Encoder { return &Encoder{} }

// barcodeEncoder abstracts the barcode encode+render operations for testing.
type barcodeEncoder interface {
	Encode(text string) error
	Render(width, height int) (image.Image, error)
}

// newBarcode is the factory function used to create the underlying barcode.
// Tests can override this to inject errors.
var newBarcode = func() barcodeEncoder {
	return barcode.NewCode128Barcode()
}

// Encode generates a Code 128 barcode image for text at the given width and
// height in pixels.
func (e *Encoder) Encode(text string, width, height int) (image.Image, error) {
	if text == "" {
		return nil, fmt.Errorf("code128: text must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("code128: width and height must be positive, got %dx%d", width, height)
	}

	bc := newBarcode()
	if err := bc.Encode(text); err != nil {
		return nil, fmt.Errorf("code128 encode %q: %w", text, err)
	}
	return bc.Render(width, height)
}

// Validate returns an error if text cannot be encoded as Code 128.
func Validate(text string) error {
	if text == "" {
		return fmt.Errorf("code128: text must not be empty")
	}
	for i, r := range text {
		if r > 0x7E {
			return fmt.Errorf("code128: character %q at position %d is outside Code 128 range", r, i)
		}
	}
	return nil
}
