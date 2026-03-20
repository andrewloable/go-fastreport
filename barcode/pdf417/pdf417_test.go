package pdf417_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/pdf417"
)

func TestNew_Defaults(t *testing.T) {
	e := pdf417.New()
	if e.SecurityLevel != 2 {
		t.Errorf("SecurityLevel = %d, want 2", e.SecurityLevel)
	}
	if e.ForegroundColor != color.Black {
		t.Error("ForegroundColor should default to black")
	}
	if e.BackgroundColor != color.White {
		t.Error("BackgroundColor should default to white")
	}
}

func TestEncode_Basic(t *testing.T) {
	e := pdf417.New()
	img, err := e.Encode("Hello PDF417", 300, 150)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 300 || b.Dy() != 150 {
		t.Errorf("image size = %dx%d, want 300x150", b.Dx(), b.Dy())
	}
}

func TestEncode_EmptyText(t *testing.T) {
	e := pdf417.New()
	_, err := e.Encode("", 300, 150)
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestEncode_ZeroWidth(t *testing.T) {
	e := pdf417.New()
	_, err := e.Encode("test", 0, 100)
	if err == nil {
		t.Error("expected error for zero width")
	}
}

func TestEncode_ZeroHeight(t *testing.T) {
	e := pdf417.New()
	_, err := e.Encode("test", 100, 0)
	if err == nil {
		t.Error("expected error for zero height")
	}
}

func TestEncode_SecurityLevels(t *testing.T) {
	// Use a large image to accommodate the higher-ECC levels (7 and 8 are very dense).
	for sl := byte(1); sl <= 8; sl++ {
		e := pdf417.New()
		e.SecurityLevel = sl
		img, err := e.Encode("Security test", 600, 300)
		if err != nil {
			t.Errorf("SecurityLevel %d: Encode error: %v", sl, err)
			continue
		}
		if img == nil {
			t.Errorf("SecurityLevel %d: nil image", sl)
		}
	}
}

func TestEncode_CustomColors(t *testing.T) {
	e := pdf417.New()
	e.ForegroundColor = color.RGBA{B: 255, A: 255} // blue
	e.BackgroundColor = color.RGBA{G: 255, A: 255} // green
	img, err := e.Encode("Colors", 200, 100)
	if err != nil {
		t.Fatalf("Encode with custom colors: %v", err)
	}
	if img == nil {
		t.Error("image should not be nil")
	}
}

func TestEncode_LongText(t *testing.T) {
	e := pdf417.New()
	long := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%"
	img, err := e.Encode(long, 400, 150)
	if err != nil {
		t.Fatalf("Encode long text: %v", err)
	}
	if img.Bounds().Dx() != 400 {
		t.Error("image width should be 400")
	}
}

func TestEncode_ZeroSecurityLevel_UsesDefault(t *testing.T) {
	// SecurityLevel=0 → covered by `if sl == 0 { sl = 2 }` branch.
	e := pdf417.New()
	e.SecurityLevel = 0
	img, err := e.Encode("zero security level", 300, 100)
	if err != nil {
		t.Fatalf("Encode with SecurityLevel=0: %v", err)
	}
	if img == nil {
		t.Error("image should not be nil")
	}
}

func TestEncode_SecurityLevelTooHigh_ClampedGracefully(t *testing.T) {
	// SecurityLevel=9 is out of range (0-8 valid). The native encoder
	// clamps it to the valid range and produces output.
	e := pdf417.New()
	e.SecurityLevel = 9
	img, err := e.Encode("overflow", 300, 100)
	if err != nil {
		// Error is acceptable if the encoder validates the range.
		return
	}
	if img == nil {
		t.Error("expected non-nil image or error for SecurityLevel=9")
	}
}
