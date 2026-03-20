// Package code93 provides a Code 93 barcode encoder for go-fastreport.
// Uses the native Go Code93 encoder from the parent barcode package.
package code93

import (
	"fmt"
	"image"
	"image/color"

	"github.com/andrewloable/go-fastreport/barcode"
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

// barcodeEncoder abstracts the barcode encode+render operations for testing.
type barcodeEncoder interface {
	Encode(text string) error
	Render(width, height int) (image.Image, error)
}

// newStandardBarcode creates a standard Code93 encoder. Tests can override this.
var newStandardBarcode = func(includeChecksum, fullASCII bool) barcodeEncoder {
	b := barcode.NewCode93Barcode()
	b.IncludeChecksum = includeChecksum
	b.FullASCIIMode = fullASCII
	return b
}

// newExtendedBarcode creates an extended Code93 encoder. Tests can override this.
var newExtendedBarcode = func() barcodeEncoder {
	return barcode.NewCode93ExtendedBarcode()
}

// Encode encodes text as a Code 93 barcode and renders it to width*height pixels.
func (e *Encoder) Encode(text string, width, height int) (image.Image, error) {
	if text == "" {
		return nil, fmt.Errorf("code93: text must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("code93: width and height must be > 0")
	}

	var bc barcodeEncoder
	if e.FullASCIIMode {
		b := newExtendedBarcode()
		if err := b.Encode(text); err != nil {
			return nil, fmt.Errorf("code93: %w", err)
		}
		bc = b
	} else {
		b := newStandardBarcode(e.IncludeChecksum, e.FullASCIIMode)
		if err := b.Encode(text); err != nil {
			return nil, fmt.Errorf("code93: %w", err)
		}
		bc = b
	}

	img, err := bc.Render(width, height)
	if err != nil {
		return nil, fmt.Errorf("code93: render %w", err)
	}
	return e.applyColors(img), nil
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
