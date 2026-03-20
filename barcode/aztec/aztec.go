// Package aztec provides an Aztec 2D barcode encoder for go-fastreport.
// Uses the native Go Aztec encoder from the parent barcode package.
package aztec

import (
	"fmt"
	"image"
	"image/color"

	"github.com/andrewloable/go-fastreport/barcode"
)

// Encoder encodes Aztec barcodes.
type Encoder struct {
	// MinECCPercent is the minimum error correction percentage (default: 23).
	MinECCPercent int
	// UserSpecifiedLayers controls the number of layers (0 = auto).
	UserSpecifiedLayers int
	// ForegroundColor is the module color (default: black).
	ForegroundColor color.Color
	// BackgroundColor is the background color (default: white).
	BackgroundColor color.Color
}

// New creates an Encoder with defaults.
func New() *Encoder {
	return &Encoder{
		MinECCPercent:   23,
		ForegroundColor: color.Black,
		BackgroundColor: color.White,
	}
}

// Encode encodes text as an Aztec barcode and scales it to size*size pixels.
func (e *Encoder) Encode(text string, size int) (image.Image, error) {
	if text == "" {
		return nil, fmt.Errorf("aztec: text must not be empty")
	}
	if size <= 0 {
		return nil, fmt.Errorf("aztec: size must be > 0")
	}

	bc := barcode.NewAztecBarcode()
	bc.MinECCPercent = e.MinECCPercent
	bc.UserSpecifiedLayers = e.UserSpecifiedLayers
	if err := bc.Encode(text); err != nil {
		return nil, fmt.Errorf("aztec: encode %w", err)
	}

	img, err := bc.Render(size, size)
	if err != nil {
		return nil, fmt.Errorf("aztec: render %w", err)
	}
	return e.applyColors(img), nil
}

// EncodeMatrix returns the raw boolean matrix for the Aztec barcode.
func (e *Encoder) EncodeMatrix(text string) ([][]bool, error) {
	if text == "" {
		return nil, fmt.Errorf("aztec: text must not be empty")
	}
	bc := barcode.NewAztecBarcode()
	bc.MinECCPercent = e.MinECCPercent
	bc.UserSpecifiedLayers = e.UserSpecifiedLayers
	if err := bc.Encode(text); err != nil {
		return nil, fmt.Errorf("aztec: encode %w", err)
	}
	matrix, _, _ := bc.GetMatrix()
	return matrix, nil
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
