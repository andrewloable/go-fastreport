package code2of5_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/code2of5"
)

func TestNew_Defaults(t *testing.T) {
	enc := code2of5.New()
	if enc == nil {
		t.Fatal("New returned nil")
	}
	if !enc.Interleaved {
		t.Error("Interleaved should default to true")
	}
	if enc.ForegroundColor != color.Black {
		t.Error("ForegroundColor should default to black")
	}
	if enc.BackgroundColor != color.White {
		t.Error("BackgroundColor should default to white")
	}
}

func TestEncode_Interleaved_FourDigits(t *testing.T) {
	enc := code2of5.New() // Interleaved=true
	img, err := enc.Encode("1234", 200, 100)
	if err != nil {
		t.Fatalf("Encode interleaved 4 digits error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_Interleaved_TwoDigits(t *testing.T) {
	enc := code2of5.New()
	img, err := enc.Encode("12", 200, 100)
	if err != nil {
		t.Fatalf("Encode interleaved 2 digits error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_EmptyContent_Error(t *testing.T) {
	enc := code2of5.New()
	_, err := enc.Encode("", 200, 100)
	if err == nil {
		t.Error("expected error for empty content")
	}
}

func TestEncode_ZeroWidth_Error(t *testing.T) {
	enc := code2of5.New()
	_, err := enc.Encode("1234", 0, 100)
	if err == nil {
		t.Error("expected error for zero width")
	}
}

func TestEncode_ZeroHeight_Error(t *testing.T) {
	enc := code2of5.New()
	_, err := enc.Encode("1234", 100, 0)
	if err == nil {
		t.Error("expected error for zero height")
	}
}

func TestEncode_Standard_OddDigits(t *testing.T) {
	enc := code2of5.New()
	enc.Interleaved = false
	// Standard 2-of-5 accepts odd number of digits.
	img, err := enc.Encode("123", 200, 100)
	if err != nil {
		t.Fatalf("Encode standard 3 digits error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_ImageSize(t *testing.T) {
	enc := code2of5.New()
	img, err := enc.Encode("1234", 200, 100)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 200 || b.Dy() != 100 {
		t.Errorf("image size: got %dx%d, want 200x100", b.Dx(), b.Dy())
	}
}

func TestValidate_EvenDigits(t *testing.T) {
	enc := code2of5.New() // Interleaved=true
	if err := enc.Validate("1234"); err != nil {
		t.Errorf("Validate even digits unexpected error: %v", err)
	}
}

func TestValidate_InvalidChars_Error(t *testing.T) {
	enc := code2of5.New()
	if err := enc.Validate("ABCD"); err == nil {
		t.Error("expected error for non-digit characters")
	}
}

func TestValidate_Interleaved_OddDigits_Error(t *testing.T) {
	enc := code2of5.New() // Interleaved=true
	if err := enc.Validate("123"); err == nil {
		t.Error("expected error for odd digit count in interleaved mode")
	}
}

func TestValidate_Standard_OddDigits(t *testing.T) {
	enc := code2of5.New()
	enc.Interleaved = false
	// Standard 2-of-5 should accept odd number of digits.
	if err := enc.Validate("123"); err != nil {
		t.Errorf("Validate standard odd digits unexpected error: %v", err)
	}
}

func TestEncode_CustomColors(t *testing.T) {
	enc := code2of5.New()
	enc.ForegroundColor = color.RGBA{R: 255, G: 0, B: 0, A: 255} // red
	enc.BackgroundColor = color.RGBA{R: 0, G: 0, B: 255, A: 255} // blue
	img, err := enc.Encode("1234", 200, 100)
	if err != nil {
		t.Fatalf("Encode with custom colors error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode with custom colors returned nil image")
	}
}

func TestEncode_NonDigit_Error(t *testing.T) {
	// Non-digit input → boomtwooffive.Encode returns error.
	enc := code2of5.New()
	_, err := enc.Encode("ABCD", 200, 100)
	if err == nil {
		t.Error("expected error for non-digit content in code2of5")
	}
}
