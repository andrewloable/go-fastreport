package code93_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/code93"
)

func TestNew_Defaults(t *testing.T) {
	enc := code93.New()
	if enc == nil {
		t.Fatal("New returned nil")
	}
	if !enc.IncludeChecksum {
		t.Error("IncludeChecksum should default to true")
	}
	if enc.FullASCIIMode {
		t.Error("FullASCIIMode should default to false")
	}
	if enc.ForegroundColor != color.Black {
		t.Error("ForegroundColor should default to black")
	}
	if enc.BackgroundColor != color.White {
		t.Error("BackgroundColor should default to white")
	}
}

func TestEncode_BasicText(t *testing.T) {
	enc := code93.New()
	img, err := enc.Encode("HELLO", 200, 100)
	if err != nil {
		t.Fatalf("Encode HELLO error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_EmptyText_Error(t *testing.T) {
	enc := code93.New()
	_, err := enc.Encode("", 200, 100)
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestEncode_ZeroWidth_Error(t *testing.T) {
	enc := code93.New()
	_, err := enc.Encode("HELLO", 0, 100)
	if err == nil {
		t.Error("expected error for zero width")
	}
}

func TestEncode_ZeroHeight_Error(t *testing.T) {
	enc := code93.New()
	_, err := enc.Encode("HELLO", 100, 0)
	if err == nil {
		t.Error("expected error for zero height")
	}
}

func TestEncode_NoChecksum(t *testing.T) {
	enc := code93.New()
	enc.IncludeChecksum = false
	img, err := enc.Encode("HELLO", 200, 100)
	if err != nil {
		t.Fatalf("Encode with IncludeChecksum=false error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_FullASCIIMode(t *testing.T) {
	enc := code93.New()
	enc.FullASCIIMode = true
	// Use a wider canvas to accommodate the longer barcode generated in full ASCII mode.
	img, err := enc.Encode("Hello!", 600, 100)
	if err != nil {
		t.Fatalf("Encode with FullASCIIMode=true error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_CustomColors(t *testing.T) {
	enc := code93.New()
	enc.ForegroundColor = color.RGBA{R: 255, G: 0, B: 0, A: 255} // red
	enc.BackgroundColor = color.RGBA{R: 0, G: 0, B: 255, A: 255} // blue
	img, err := enc.Encode("HELLO", 200, 100)
	if err != nil {
		t.Fatalf("Encode with custom colors error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode with custom colors returned nil image")
	}
}

func TestEncode_ImageSize(t *testing.T) {
	enc := code93.New()
	img, err := enc.Encode("TEST", 300, 150)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 300 || b.Dy() != 150 {
		t.Errorf("image size: got %dx%d, want 300x150", b.Dx(), b.Dy())
	}
}
