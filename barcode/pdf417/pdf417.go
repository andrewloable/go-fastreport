// Package pdf417 provides a PDF417 stacked 2D barcode encoder for go-fastreport.
// Uses the native Go PDF417 encoder from the parent barcode package.
package pdf417

import (
	"fmt"
	"image"
	"image/color"

	"github.com/andrewloable/go-fastreport/barcode"
)

// Encoder encodes PDF417 barcodes.
type Encoder struct {
	// SecurityLevel controls error correction (0-8, default: 2).
	SecurityLevel byte
	// ForegroundColor is the module color (default: black).
	ForegroundColor color.Color
	// BackgroundColor is the background color (default: white).
	BackgroundColor color.Color
}

// New creates an Encoder with defaults.
func New() *Encoder {
	return &Encoder{
		SecurityLevel:   2,
		ForegroundColor: color.Black,
		BackgroundColor: color.White,
	}
}

// Encode encodes text as a PDF417 barcode and renders it to width*height pixels.
func (e *Encoder) Encode(text string, width, height int) (image.Image, error) {
	if text == "" {
		return nil, fmt.Errorf("pdf417: text must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("pdf417: width and height must be > 0")
	}

	bc := barcode.NewPDF417Barcode()
	bc.SecurityLevel = int(e.SecurityLevel)
	if err := bc.Encode(text); err != nil {
		return nil, fmt.Errorf("pdf417: encode %w", err)
	}

	img, err := bc.Render(width, height)
	if err != nil {
		return nil, fmt.Errorf("pdf417: render %w", err)
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
