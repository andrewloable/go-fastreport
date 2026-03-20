// Package code2of5 provides 2-of-5 barcode encoders for go-fastreport.
// Uses the native Go 2-of-5 encoder from the parent barcode package.
package code2of5

import (
	"fmt"
	"image"
	"image/color"

	"github.com/andrewloable/go-fastreport/barcode"
)

// Encoder encodes 2-of-5 barcodes.
type Encoder struct {
	// Interleaved selects Interleaved 2-of-5 (ITF) when true (default: true).
	Interleaved bool
	// ForegroundColor is the bar color (default: black).
	ForegroundColor color.Color
	// BackgroundColor is the background color (default: white).
	BackgroundColor color.Color
}

// New creates an Encoder with Interleaved 2-of-5 as default.
func New() *Encoder {
	return &Encoder{
		Interleaved:     true,
		ForegroundColor: color.Black,
		BackgroundColor: color.White,
	}
}

// Encode encodes numeric content as a 2-of-5 barcode and renders it to width*height pixels.
func (e *Encoder) Encode(content string, width, height int) (image.Image, error) {
	if content == "" {
		return nil, fmt.Errorf("code2of5: content must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("code2of5: width and height must be > 0")
	}

	var bc interface {
		Encode(text string) error
		Render(width, height int) (image.Image, error)
	}

	if e.Interleaved {
		b := barcode.NewCode2of5Barcode()
		b.Interleaved = true
		bc = b
	} else {
		bc = barcode.NewCode2of5IndustrialBarcode()
	}

	if err := bc.Encode(content); err != nil {
		return nil, fmt.Errorf("code2of5: %w", err)
	}
	img, err := bc.Render(width, height)
	if err != nil {
		return nil, fmt.Errorf("code2of5: render %w", err)
	}
	return e.applyColors(img), nil
}

// Validate checks whether content is valid for 2-of-5 encoding (digits only).
func (e *Encoder) Validate(content string) error {
	for _, r := range content {
		if r < '0' || r > '9' {
			return fmt.Errorf("code2of5: invalid character %q (only digits allowed)", r)
		}
	}
	if e.Interleaved && len(content)%2 != 0 {
		return fmt.Errorf("code2of5: interleaved mode requires even number of digits, got %d", len(content))
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
