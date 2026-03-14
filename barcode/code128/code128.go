// Package code128 provides Code 128 barcode generation for go-fastreport.
// It is backed by the pure-Go github.com/boombuler/barcode library.
package code128

import (
	"fmt"
	"image"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
)

// Encoder generates Code 128 barcode images from text content.
// Code 128 supports the full printable ASCII character set and uses automatic
// code-set switching (A/B/C) to minimise symbol length.
type Encoder struct{}

// NewEncoder creates a Code128 Encoder.
func NewEncoder() *Encoder { return &Encoder{} }

// Encode generates a Code 128 barcode image for text at the given width and
// height in pixels.
// Returns an error when text contains characters not supported by Code 128.
func (e *Encoder) Encode(text string, width, height int) (image.Image, error) {
	if text == "" {
		return nil, fmt.Errorf("code128: text must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("code128: width and height must be positive, got %dx%d", width, height)
	}

	bc, err := code128.Encode(text)
	if err != nil {
		return nil, fmt.Errorf("code128 encode %q: %w", text, err)
	}

	scaled, err := barcode.Scale(bc, width, height)
	if err != nil {
		return nil, fmt.Errorf("code128 scale: %w", err)
	}
	return scaled, nil
}

// Validate returns an error if text cannot be encoded as Code 128.
// Code 128 supports printable ASCII (0x20–0x7E) plus some control characters.
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
