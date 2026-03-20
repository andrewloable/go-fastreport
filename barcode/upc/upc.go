// Package upc provides UPC-A barcode encoding for go-fastreport.
// Uses the native Go UPC-A encoder from the parent barcode package.
package upc

import (
	"fmt"
	"image"
	"image/color"

	"github.com/andrewloable/go-fastreport/barcode"
)

// Encoder encodes UPC-A barcodes.
type Encoder struct {
	// ForegroundColor is the bar color (default: black).
	ForegroundColor color.Color
	// BackgroundColor is the background color (default: white).
	BackgroundColor color.Color
}

// New creates an Encoder with defaults.
func New() *Encoder {
	return &Encoder{
		ForegroundColor: color.Black,
		BackgroundColor: color.White,
	}
}

// barcodeEncoder abstracts the barcode encode+render operations for testing.
type barcodeEncoder interface {
	Encode(text string) error
	Render(width, height int) (image.Image, error)
}

// newBarcode is the factory function used to create the underlying barcode.
// Tests can override this to inject errors.
var newBarcode = func() barcodeEncoder {
	return barcode.NewUPCABarcode()
}

// Encode encodes an 11- or 12-digit UPC-A code and renders it to width*height pixels.
func (e *Encoder) Encode(code string, width, height int) (image.Image, error) {
	if code == "" {
		return nil, fmt.Errorf("upc: code must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("upc: width and height must be > 0")
	}
	if err := e.Validate(code); err != nil {
		return nil, err
	}

	bc := newBarcode()
	if err := bc.Encode(code); err != nil {
		return nil, fmt.Errorf("upc: %w", err)
	}
	img, err := bc.Render(width, height)
	if err != nil {
		return nil, fmt.Errorf("upc: render %w", err)
	}
	return e.applyColors(img), nil
}

// Validate checks whether code is a valid UPC-A digit string (11 or 12 digits).
func (e *Encoder) Validate(code string) error {
	if len(code) != 11 && len(code) != 12 {
		return fmt.Errorf("upc: code must be 11 or 12 digits, got %d", len(code))
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return fmt.Errorf("upc: invalid character %q (only digits allowed)", r)
		}
	}
	return nil
}

func (e *Encoder) applyColors(img image.Image) image.Image {
	if e.ForegroundColor == color.Black && e.BackgroundColor == color.White {
		return img
	}
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 {
				out.Set(x, y, e.ForegroundColor)
			} else {
				out.Set(x, y, e.BackgroundColor)
			}
		}
	}
	return out
}
