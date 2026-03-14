// Package ean provides EAN-8 and EAN-13 barcode encoders for go-fastreport.
package ean

import (
	"fmt"
	"image"
	"image/color"

	"github.com/boombuler/barcode"
	boomean "github.com/boombuler/barcode/ean"
)

// Encoder encodes EAN-8 and EAN-13 barcodes.
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

// Encode encodes a numeric EAN code (7 or 8 digits for EAN-8; 12 or 13 for EAN-13)
// and scales it to width×height pixels.
// The checksum digit is automatically calculated if omitted.
func (e *Encoder) Encode(code string, width, height int) (image.Image, error) {
	if code == "" {
		return nil, fmt.Errorf("ean: code must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("ean: width and height must be > 0")
	}
	bc, err := boomean.Encode(code)
	if err != nil {
		return nil, fmt.Errorf("ean: %w", err)
	}
	scaled, err := barcode.Scale(bc, width, height)
	if err != nil {
		return nil, fmt.Errorf("ean: scale %w", err)
	}
	return e.applyColors(scaled), nil
}

// Validate checks whether code is a valid EAN-8 or EAN-13 digit string.
func (e *Encoder) Validate(code string) error {
	if len(code) != 7 && len(code) != 8 && len(code) != 12 && len(code) != 13 {
		return fmt.Errorf("ean: code must be 7/8 (EAN-8) or 12/13 (EAN-13) digits, got %d", len(code))
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return fmt.Errorf("ean: invalid character %q (only digits allowed)", r)
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
