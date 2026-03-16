// Package datamatrix provides a DataMatrix 2D barcode encoder for go-fastreport.
package datamatrix

import (
	"fmt"
	"image"
	"image/color"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/datamatrix"
)

// scaleBarcode is the barcode.Scale function, replaceable in tests.
var scaleBarcode = barcode.Scale

// Encoder encodes DataMatrix barcodes.
type Encoder struct {
	// ForegroundColor is the module color (default: black).
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

// Encode encodes text as a DataMatrix barcode and scales it to width×height pixels.
func (e *Encoder) Encode(text string, width, height int) (image.Image, error) {
	if text == "" {
		return nil, fmt.Errorf("datamatrix: text must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("datamatrix: width and height must be > 0")
	}
	bc, err := datamatrix.Encode(text)
	if err != nil {
		return nil, fmt.Errorf("datamatrix: encode %w", err)
	}
	scaled, err := scaleBarcode(bc, width, height)
	if err != nil {
		return nil, fmt.Errorf("datamatrix: scale %w", err)
	}
	return e.applyColors(scaled), nil
}

// EncodeMatrix returns the raw boolean matrix for the DataMatrix barcode.
func (e *Encoder) EncodeMatrix(text string) ([][]bool, error) {
	if text == "" {
		return nil, fmt.Errorf("datamatrix: text must not be empty")
	}
	bc, err := datamatrix.Encode(text)
	if err != nil {
		return nil, fmt.Errorf("datamatrix: encode %w", err)
	}
	bounds := bc.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y
	matrix := make([][]bool, h)
	for y := 0; y < h; y++ {
		matrix[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := bc.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			matrix[y][x] = r == 0 && g == 0 && b == 0
		}
	}
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
