// Package qr provides QR Code barcode generation for go-fastreport.
// Uses the native Go QR encoder from the parent barcode package.
package qr

import (
	"fmt"
	"image"
	"image/color"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ErrorCorrectionLevel controls the QR error-correction capacity.
type ErrorCorrectionLevel string

const (
	// ECLevelL recovers up to 7% of data.
	ECLevelL ErrorCorrectionLevel = "L"
	// ECLevelM recovers up to 15% of data (default).
	ECLevelM ErrorCorrectionLevel = "M"
	// ECLevelQ recovers up to 25% of data.
	ECLevelQ ErrorCorrectionLevel = "Q"
	// ECLevelH recovers up to 30% of data.
	ECLevelH ErrorCorrectionLevel = "H"
)

// Encoder generates QR Code images from text content.
type Encoder struct {
	// ECLevel sets the error-correction level (default ECLevelM).
	ECLevel ErrorCorrectionLevel
	// QuietZone is the number of quiet-zone modules to add on each side (default 4).
	QuietZone int
	// ForegroundColor is the bar/module color (default black).
	ForegroundColor color.Color
	// BackgroundColor is the background color (default white).
	BackgroundColor color.Color
}

// NewEncoder creates a QR Encoder with sensible defaults.
func NewEncoder() *Encoder {
	return &Encoder{
		ECLevel:         ECLevelM,
		QuietZone:       4,
		ForegroundColor: color.Black,
		BackgroundColor: color.White,
	}
}

// Encode generates a QR Code image for text at the given size in pixels.
func (e *Encoder) Encode(text string, size int) (image.Image, error) {
	if text == "" {
		return nil, fmt.Errorf("qr: text must not be empty")
	}
	if size <= 0 {
		return nil, fmt.Errorf("qr: size must be positive, got %d", size)
	}

	bc := barcode.NewQRBarcode()
	bc.ErrorCorrection = string(e.ECLevel)
	if err := bc.Encode(text); err != nil {
		return nil, fmt.Errorf("qr encode %q: %w", text, err)
	}
	img, err := bc.Render(size, size)
	if err != nil {
		return nil, fmt.Errorf("qr render: %w", err)
	}

	fg := e.ForegroundColor
	bg := e.BackgroundColor
	if isDefaultColors(fg, bg) {
		return img, nil
	}
	return applyColors(img, fg, bg), nil
}

// EncodeMatrix generates a QR Code and returns the raw module matrix.
func (e *Encoder) EncodeMatrix(text string) ([][]bool, error) {
	if text == "" {
		return nil, fmt.Errorf("qr: text must not be empty")
	}
	bc := barcode.NewQRBarcode()
	bc.ErrorCorrection = string(e.ECLevel)
	if err := bc.Encode(text); err != nil {
		return nil, fmt.Errorf("qr encode %q: %w", text, err)
	}
	matrix, _, _ := bc.GetMatrix()
	return matrix, nil
}

func isDefaultColors(fg, bg color.Color) bool {
	if fg == nil || bg == nil {
		return true
	}
	fr, fg2, fb, _ := fg.RGBA()
	br, bg2, bb, _ := bg.RGBA()
	black := fr == 0 && fg2 == 0 && fb == 0
	white := br == 0xffff && bg2 == 0xffff && bb == 0xffff
	return black && white
}

func applyColors(src image.Image, fg, bg color.Color) image.Image {
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := src.At(x, y).RGBA()
			lum := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			if lum < 32767 {
				dst.Set(x, y, fg)
			} else {
				dst.Set(x, y, bg)
			}
		}
	}
	return dst
}
