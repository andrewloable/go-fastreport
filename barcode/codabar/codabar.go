// Package codabar provides a Codabar barcode encoder for go-fastreport.
// Uses the native Go Codabar encoder from the parent barcode package.
package codabar

import (
	"fmt"
	"image"
	"image/color"

	"github.com/andrewloable/go-fastreport/barcode"
)

// Encoder encodes Codabar barcodes.
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
	return barcode.NewCodabarBarcode()
}

// Encode encodes content as a Codabar barcode and renders it to width*height pixels.
func (e *Encoder) Encode(content string, width, height int) (image.Image, error) {
	if content == "" {
		return nil, fmt.Errorf("codabar: content must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("codabar: width and height must be > 0")
	}

	bc := newBarcode()
	if err := bc.Encode(content); err != nil {
		return nil, fmt.Errorf("codabar: %w", err)
	}
	img, err := bc.Render(width, height)
	if err != nil {
		return nil, fmt.Errorf("codabar: render %w", err)
	}
	return e.applyColors(img), nil
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
