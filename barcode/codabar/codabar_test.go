package codabar_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/codabar"
)

func TestNew_Defaults(t *testing.T) {
	enc := codabar.New()
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

func TestEncode_ValidContent(t *testing.T) {
	enc := codabar.New()
	// Codabar content must start and end with A, B, C, or D.
	img, err := enc.Encode("A1234B", 200, 100)
	if err != nil {
		t.Fatalf("Encode A1234B error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_ImageSize(t *testing.T) {
	enc := codabar.New()
	img, err := enc.Encode("A1234B", 200, 100)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 200 || b.Dy() != 100 {
		t.Errorf("image size: got %dx%d, want 200x100", b.Dx(), b.Dy())
	}
}

func TestEncode_EmptyContent_Error(t *testing.T) {
	enc := codabar.New()
	_, err := enc.Encode("", 200, 100)
	if err == nil {
		t.Error("expected error for empty content")
	}
}

func TestEncode_ZeroWidth_Error(t *testing.T) {
	enc := codabar.New()
	_, err := enc.Encode("A1234B", 0, 100)
	if err == nil {
		t.Error("expected error for zero width")
	}
}

func TestEncode_ZeroHeight_Error(t *testing.T) {
	enc := codabar.New()
	_, err := enc.Encode("A1234B", 100, 0)
	if err == nil {
		t.Error("expected error for zero height")
	}
}

func TestEncode_AllStartStopChars(t *testing.T) {
	enc := codabar.New()
	cases := []string{
		"A123A",
		"B456B",
		"C789C",
		"D012D",
		"A0B",
		"C0D",
	}
	for _, c := range cases {
		img, err := enc.Encode(c, 200, 100)
		if err != nil {
			t.Errorf("Encode(%q) error: %v", c, err)
			continue
		}
		if img == nil {
			t.Errorf("Encode(%q) returned nil image", c)
		}
	}
}

func TestEncode_CustomColors(t *testing.T) {
	enc := codabar.New()
	enc.ForegroundColor = color.RGBA{R: 255, G: 0, B: 0, A: 255} // red
	enc.BackgroundColor = color.RGBA{R: 0, G: 0, B: 255, A: 255} // blue
	img, err := enc.Encode("A1234B", 200, 100)
	if err != nil {
		t.Fatalf("Encode with custom colors error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode with custom colors returned nil image")
	}
}

func TestEncode_InvalidStartStop_Error(t *testing.T) {
	// Content "12345" has no valid Codabar start/stop characters → boombuler error.
	enc := codabar.New()
	_, err := enc.Encode("12345", 200, 100)
	if err == nil {
		t.Error("expected error for content without valid Codabar start/stop chars")
	}
}
