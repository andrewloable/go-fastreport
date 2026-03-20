// Package code39 provides Code 39 barcode generation for go-fastreport.
// Uses the native Go Code39 encoder from the parent barcode package.
package code39

import (
	"fmt"
	"image"
	"strings"

	"github.com/andrewloable/go-fastreport/barcode"
)

// validChars is the set of characters allowed in standard Code 39.
const validChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-. $/+%*"

// Encoder generates Code 39 barcode images from text content.
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

	bc := barcode.NewCode39Barcode()
	bc.CalcChecksum = e.CalcChecksum
	bc.AllowExtended = e.AllowExtended
	if err := bc.Encode(strings.ToUpper(text)); err != nil {
		return nil, fmt.Errorf("code39 encode %q: %w", text, err)
	}
	return bc.Render(width, height)
}

// Validate returns an error if text contains characters outside the standard
// Code 39 character set.
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
