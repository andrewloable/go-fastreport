// Package datamatrix provides a DataMatrix 2D barcode encoder for go-fastreport.
// Uses the native Go DataMatrix encoder from the parent barcode package.
package datamatrix

import (
	"fmt"
	"image"
	"image/color"

	"github.com/andrewloable/go-fastreport/barcode"
)

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

// Encode encodes text as a DataMatrix barcode and renders it to width*height pixels.
func (e *Encoder) Encode(text string, width, height int) (image.Image, error) {
	if text == "" {
		return nil, fmt.Errorf("datamatrix: text must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("datamatrix: width and height must be > 0")
	}

	bc := barcode.NewDataMatrixBarcode()
	if err := bc.Encode(text); err != nil {
		return nil, fmt.Errorf("datamatrix: encode %w", err)
	}
	img, err := bc.Render(width, height)
	if err != nil {
		return nil, fmt.Errorf("datamatrix: render %w", err)
	}
	return e.applyColors(img), nil
}

// EncodeMatrix returns the raw boolean matrix for the DataMatrix barcode.
func (e *Encoder) EncodeMatrix(text string) ([][]bool, error) {
	if text == "" {
		return nil, fmt.Errorf("datamatrix: text must not be empty")
	}
	bc := barcode.NewDataMatrixBarcode()
	if err := bc.Encode(text); err != nil {
		return nil, fmt.Errorf("datamatrix: encode %w", err)
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
