// Package qr provides QR Code barcode generation for go-fastreport.
// It is backed by the pure-Go github.com/boombuler/barcode library.
package qr

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
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

// toECLevel converts the string level to the boombuler ErrorCorrectionLevel.
func toECLevel(level ErrorCorrectionLevel) qr.ErrorCorrectionLevel {
	switch level {
	case ECLevelL:
		return qr.L
	case ECLevelQ:
		return qr.Q
	case ECLevelH:
		return qr.H
	default: // ECLevelM and anything unrecognised
		return qr.M
	}
}

// scaleBarcode is the barcode.Scale function, replaceable in tests.
var scaleBarcode = barcode.Scale

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
// The returned image is an NRGBA image of size×size pixels.
func (e *Encoder) Encode(text string, size int) (image.Image, error) {
	if text == "" {
		return nil, fmt.Errorf("qr: text must not be empty")
	}
	if size <= 0 {
		return nil, fmt.Errorf("qr: size must be positive, got %d", size)
	}

	// Generate the QR code using boombuler/barcode.
	bc, err := qr.Encode(text, toECLevel(e.ECLevel), qr.Auto)
	if err != nil {
		return nil, fmt.Errorf("qr encode %q: %w", text, err)
	}

	// Scale to the requested size.
	scaled, err := scaleBarcode(bc, size, size)
	if err != nil {
		return nil, fmt.Errorf("qr scale: %w", err)
	}

	// Apply foreground/background colours if they differ from the defaults.
	fg := e.ForegroundColor
	bg := e.BackgroundColor
	if isDefaultColors(fg, bg) {
		return scaled, nil
	}
	return applyColors(scaled, fg, bg), nil
}

// EncodeMatrix generates a QR Code and returns the raw module matrix as a
// [][]bool (true = dark module).  Useful for tests and custom rendering.
func (e *Encoder) EncodeMatrix(text string) ([][]bool, error) {
	if text == "" {
		return nil, fmt.Errorf("qr: text must not be empty")
	}
	bc, err := qr.Encode(text, toECLevel(e.ECLevel), qr.Auto)
	if err != nil {
		return nil, fmt.Errorf("qr encode %q: %w", text, err)
	}

	bounds := bc.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y

	matrix := make([][]bool, h)
	for y := 0; y < h; y++ {
		matrix[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := bc.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
			// Dark module = low luminance (black-ish).
			lum := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			matrix[y][x] = lum < 32767 // mid-point of uint16 range
		}
	}
	return matrix, nil
}

// isDefaultColors returns true when fg/bg are the standard black/white.
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

// applyColors replaces black pixels with fg and white pixels with bg.
func applyColors(src image.Image, fg, bg color.Color) image.Image {
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)
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
