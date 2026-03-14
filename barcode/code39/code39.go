// Package code39 provides Code 39 barcode generation for go-fastreport.
// It is backed by the pure-Go github.com/boombuler/barcode library.
package code39

import (
	"fmt"
	"image"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code39"
)

// validChars is the set of characters allowed in standard Code 39.
const validChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-. $/+%*"

// Encoder generates Code 39 barcode images from text content.
// Code 39 supports uppercase A–Z, digits 0–9, and the special characters
// - . SPACE $ / + %.  Extended Code 39 (AllowExtended) supports full ASCII.
type Encoder struct {
	// AllowExtended enables Code 39 Extended for full-ASCII encoding.
	AllowExtended bool
	// CalcChecksum adds a modulo-43 checksum character.
	CalcChecksum bool
}

// NewEncoder creates a Code39 Encoder with standard settings.
func NewEncoder() *Encoder { return &Encoder{} }

// Encode generates a Code 39 barcode image for text at the given width and
// height in pixels.
func (e *Encoder) Encode(text string, width, height int) (image.Image, error) {
	if text == "" {
		return nil, fmt.Errorf("code39: text must not be empty")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("code39: width and height must be positive, got %dx%d", width, height)
	}
	if !e.AllowExtended {
		if err := Validate(text); err != nil {
			return nil, err
		}
	}

	bc, err := code39.Encode(strings.ToUpper(text), e.CalcChecksum, e.AllowExtended)
	if err != nil {
		return nil, fmt.Errorf("code39 encode %q: %w", text, err)
	}

	scaled, err := barcode.Scale(bc, width, height)
	if err != nil {
		return nil, fmt.Errorf("code39 scale: %w", err)
	}
	return scaled, nil
}

// Validate returns an error if text contains characters outside the standard
// Code 39 character set (uppercase A–Z, 0–9, - . SPACE $ / + %).
// Call with AllowExtended=true to skip this check.
func Validate(text string) error {
	if text == "" {
		return fmt.Errorf("code39: text must not be empty")
	}
	upper := strings.ToUpper(text)
	for i, r := range upper {
		if !strings.ContainsRune(validChars, r) {
			return fmt.Errorf("code39: character %q at position %d is not valid for standard Code 39", r, i)
		}
	}
	return nil
}
