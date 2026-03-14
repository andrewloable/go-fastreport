// Package code93 provides a Code 93 barcode encoder for go-fastreport.
package code93

import (
	"fmt"
	"image"
	"image/color"

	"github.com/boombuler/barcode"
	boomcode93 "github.com/boombuler/barcode/code93"
)

// Encoder encodes Code 93 barcodes.
type Encoder struct {
	// IncludeChecksum adds checksum characters (recommended, default: true).
	IncludeChecksum bool
	// FullASCIIMode enables encoding the full ASCII character set (default: false).
	FullASCIIMode bool
	// ForegroundColor is the bar color (default: black).
	ForegroundColor color.Color
	// BackgroundColor is the background color (default: white).
	BackgroundColor color.Color
}

// New creates an Encoder with defaults.
func New() *Encoder {
	return &Encoder{
		IncludeChecksum: true,
		ForegroundColor: color.Black,
		BackgroundColor: color.White,
	}
}

// Encode encodes text as a Code 93 barcode and scales it to width×height pixels.
func (e *Encoder) Encode(text string, width, height int) (image.Image, error) {
	if text == "" {
		return nil, fmt.Errorf("code93: text must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("code93: width and height must be > 0")
	}
	bc, err := boomcode93.Encode(text, e.IncludeChecksum, e.FullASCIIMode)
	if err != nil {
		return nil, fmt.Errorf("code93: %w", err)
	}
	scaled, err := barcode.Scale(bc, width, height)
	if err != nil {
		return nil, fmt.Errorf("code93: scale %w", err)
	}
	return e.applyColors(scaled), nil
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
