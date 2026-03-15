package ean_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/ean"
)

func TestNew_Defaults(t *testing.T) {
	enc := ean.New()
	if enc == nil {
		t.Fatal("New returned nil")
	}
	if enc.ForegroundColor != color.Black {
		t.Error("ForegroundColor should default to black")
	}
	if enc.BackgroundColor != color.White {
		t.Error("BackgroundColor should default to white")
	}
}

func TestEncode_EAN8_SevenDigits(t *testing.T) {
	enc := ean.New()
	img, err := enc.Encode("1234567", 200, 100)
	if err != nil {
		t.Fatalf("Encode EAN-8 (7 digits) error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
	b := img.Bounds()
	if b.Dx() != 200 || b.Dy() != 100 {
		t.Errorf("image size: got %dx%d, want 200x100", b.Dx(), b.Dy())
	}
}

func TestEncode_EAN13_TwelveDigits(t *testing.T) {
	enc := ean.New()
	img, err := enc.Encode("123456789012", 200, 100)
	if err != nil {
		t.Fatalf("Encode EAN-13 (12 digits) error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_EmptyCode_Error(t *testing.T) {
	enc := ean.New()
	_, err := enc.Encode("", 200, 100)
	if err == nil {
		t.Error("expected error for empty code")
	}
}

func TestEncode_ZeroWidth_Error(t *testing.T) {
	enc := ean.New()
	_, err := enc.Encode("1234567", 0, 100)
	if err == nil {
		t.Error("expected error for zero width")
	}
}

func TestEncode_ZeroHeight_Error(t *testing.T) {
	enc := ean.New()
	_, err := enc.Encode("1234567", 100, 0)
	if err == nil {
		t.Error("expected error for zero height")
	}
}

func TestEncode_InvalidChars_Error(t *testing.T) {
	enc := ean.New()
	_, err := enc.Encode("ABCD", 200, 100)
	if err == nil {
		t.Error("expected error for non-digit characters")
	}
}

func TestValidate_EAN8_SevenDigits(t *testing.T) {
	enc := ean.New()
	if err := enc.Validate("1234567"); err != nil {
		t.Errorf("Validate 7-digit EAN-8 unexpected error: %v", err)
	}
}

func TestValidate_EAN13_TwelveDigits(t *testing.T) {
	enc := ean.New()
	if err := enc.Validate("123456789012"); err != nil {
		t.Errorf("Validate 12-digit EAN-13 unexpected error: %v", err)
	}
}

func TestValidate_TooShort_Error(t *testing.T) {
	enc := ean.New()
	if err := enc.Validate("123"); err == nil {
		t.Error("expected error for too-short code")
	}
}

func TestValidate_NonDigits_Error(t *testing.T) {
	enc := ean.New()
	if err := enc.Validate("ABC"); err == nil {
		t.Error("expected error for non-digit characters")
	}
}

func TestEncode_CustomColors(t *testing.T) {
	enc := ean.New()
	enc.ForegroundColor = color.RGBA{R: 255, G: 0, B: 0, A: 255} // red
	enc.BackgroundColor = color.RGBA{R: 0, G: 0, B: 255, A: 255} // blue
	img, err := enc.Encode("1234567", 200, 100)
	if err != nil {
		t.Fatalf("Encode with custom colors error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode with custom colors returned nil image")
	}
}

func TestEncode_WrongLength_Error(t *testing.T) {
	// 5-digit code: not empty, dimensions ok, but boomean.Encode rejects wrong length.
	enc := ean.New()
	_, err := enc.Encode("12345", 200, 100)
	if err == nil {
		t.Error("expected error from boomean for 5-digit code (wrong length)")
	}
}

func TestValidate_NonDigits_CorrectLength_Error(t *testing.T) {
	// 7-char string with non-digit → passes length check, fails digit check.
	enc := ean.New()
	if err := enc.Validate("123456A"); err == nil {
		t.Error("expected error for non-digit char in 7-char code")
	}
}
