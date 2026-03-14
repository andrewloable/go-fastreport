package upc_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/upc"
)

func TestNew_Defaults(t *testing.T) {
	enc := upc.New()
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

func TestEncode_ElevenDigits(t *testing.T) {
	enc := upc.New()
	// 11-digit UPC-A (checksum appended automatically)
	img, err := enc.Encode("01234567890", 200, 100)
	if err != nil {
		t.Fatalf("Encode 11-digit UPC-A error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_TwelveDigits(t *testing.T) {
	enc := upc.New()
	// 12-digit UPC-A with checksum included
	img, err := enc.Encode("012345678905", 200, 100)
	if err != nil {
		t.Fatalf("Encode 12-digit UPC-A error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_EmptyCode_Error(t *testing.T) {
	enc := upc.New()
	_, err := enc.Encode("", 200, 100)
	if err == nil {
		t.Error("expected error for empty code")
	}
}

func TestEncode_ZeroWidth_Error(t *testing.T) {
	enc := upc.New()
	_, err := enc.Encode("01234567890", 0, 100)
	if err == nil {
		t.Error("expected error for zero width")
	}
}

func TestEncode_ZeroHeight_Error(t *testing.T) {
	enc := upc.New()
	_, err := enc.Encode("01234567890", 100, 0)
	if err == nil {
		t.Error("expected error for zero height")
	}
}

func TestEncode_InvalidChars_Error(t *testing.T) {
	enc := upc.New()
	_, err := enc.Encode("ABCDE123456", 200, 100)
	if err == nil {
		t.Error("expected error for non-digit characters")
	}
}

func TestEncode_TooShort_Error(t *testing.T) {
	enc := upc.New()
	_, err := enc.Encode("123", 200, 100)
	if err == nil {
		t.Error("expected error for too-short code")
	}
}

func TestValidate_ElevenDigits(t *testing.T) {
	enc := upc.New()
	if err := enc.Validate("01234567890"); err != nil {
		t.Errorf("Validate 11-digit UPC-A unexpected error: %v", err)
	}
}

func TestValidate_TwelveDigits(t *testing.T) {
	enc := upc.New()
	if err := enc.Validate("012345678905"); err != nil {
		t.Errorf("Validate 12-digit UPC-A unexpected error: %v", err)
	}
}

func TestValidate_TooShort_Error(t *testing.T) {
	enc := upc.New()
	if err := enc.Validate("123"); err == nil {
		t.Error("expected error for too-short code")
	}
}

func TestValidate_NonDigits_Error(t *testing.T) {
	enc := upc.New()
	if err := enc.Validate("ABCDE12345A"); err == nil {
		t.Error("expected error for non-digit characters")
	}
}

func TestEncode_CustomColors(t *testing.T) {
	enc := upc.New()
	enc.ForegroundColor = color.RGBA{R: 255, G: 0, B: 0, A: 255} // red
	enc.BackgroundColor = color.RGBA{R: 0, G: 0, B: 255, A: 255} // blue
	img, err := enc.Encode("01234567890", 200, 100)
	if err != nil {
		t.Fatalf("Encode with custom colors error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode with custom colors returned nil image")
	}
}
