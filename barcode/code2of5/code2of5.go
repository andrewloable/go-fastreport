// Package code2of5 provides 2-of-5 barcode encoders for go-fastreport.
// Supports both Interleaved 2-of-5 (ITF) and Standard 2-of-5.
package code2of5

import (
	"fmt"
	"image"
	"image/color"

	"github.com/boombuler/barcode"
	boomtwooffive "github.com/boombuler/barcode/twooffive"
)

// scaleBarcode is the barcode.Scale function, replaceable in tests.
var scaleBarcode = barcode.Scale

// Encoder encodes 2-of-5 barcodes.
type Encoder struct {
	// Interleaved selects Interleaved 2-of-5 (ITF) when true (default: true).
	// When false, Standard (Industrial) 2-of-5 is used.
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

// Encode encodes numeric content as a 2-of-5 barcode and scales it to width×height pixels.
// Interleaved 2-of-5 requires an even number of digits.
func (e *Encoder) Encode(content string, width, height int) (image.Image, error) {
	if content == "" {
		return nil, fmt.Errorf("code2of5: content must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("code2of5: width and height must be > 0")
	}
	bc, err := boomtwooffive.Encode(content, e.Interleaved)
	if err != nil {
		return nil, fmt.Errorf("code2of5: %w", err)
	}
	scaled, err := scaleBarcode(bc, width, height)
	if err != nil {
		return nil, fmt.Errorf("code2of5: scale %w", err)
	}
	return e.applyColors(scaled), nil
}

// Validate checks whether content is valid for 2-of-5 encoding (digits only).
// For interleaved mode, digit count must be even.
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
