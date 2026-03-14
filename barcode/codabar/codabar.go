// Package codabar provides a Codabar barcode encoder for go-fastreport.
package codabar

import (
	"fmt"
	"image"
	"image/color"

	"github.com/boombuler/barcode"
	boomcodabar "github.com/boombuler/barcode/codabar"
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

// Encode encodes content as a Codabar barcode and scales it to width×height pixels.
// Content must start and end with a valid Codabar start/stop character (A, B, C, or D).
func (e *Encoder) Encode(content string, width, height int) (image.Image, error) {
	if content == "" {
		return nil, fmt.Errorf("codabar: content must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("codabar: width and height must be > 0")
	}
	bc, err := boomcodabar.Encode(content)
	if err != nil {
		return nil, fmt.Errorf("codabar: %w", err)
	}
	scaled, err := barcode.Scale(bc, width, height)
	if err != nil {
		return nil, fmt.Errorf("codabar: scale %w", err)
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
