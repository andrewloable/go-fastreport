package aztec_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/aztec"
)

func TestNew_Defaults(t *testing.T) {
	e := aztec.New()
	if e.MinECCPercent != 23 {
		t.Errorf("MinECCPercent = %d, want 23", e.MinECCPercent)
	}
	if e.UserSpecifiedLayers != 0 {
		t.Errorf("UserSpecifiedLayers = %d, want 0", e.UserSpecifiedLayers)
	}
	if e.ForegroundColor != color.Black {
		t.Error("ForegroundColor should default to black")
	}
	if e.BackgroundColor != color.White {
		t.Error("BackgroundColor should default to white")
	}
}

func TestEncode_Basic(t *testing.T) {
	e := aztec.New()
	img, err := e.Encode("Hello Aztec", 200)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 200 || b.Dy() != 200 {
		t.Errorf("image size = %dx%d, want 200x200", b.Dx(), b.Dy())
	}
}

func TestEncode_EmptyText(t *testing.T) {
	e := aztec.New()
	_, err := e.Encode("", 200)
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestEncode_ZeroSize(t *testing.T) {
	e := aztec.New()
	_, err := e.Encode("hello", 0)
	if err == nil {
		t.Error("expected error for size 0")
	}
}

func TestEncode_NegativeSize(t *testing.T) {
	e := aztec.New()
	_, err := e.Encode("hello", -1)
	if err == nil {
		t.Error("expected error for negative size")
	}
}

func TestEncodeMatrix_Basic(t *testing.T) {
	e := aztec.New()
	matrix, err := e.EncodeMatrix("Test")
	if err != nil {
		t.Fatalf("EncodeMatrix: %v", err)
	}
	if len(matrix) == 0 {
		t.Error("matrix should not be empty")
	}
	// Check matrix is square-ish (Aztec is square).
	for _, row := range matrix {
		if len(row) != len(matrix[0]) {
			t.Error("matrix rows have inconsistent width")
		}
	}
}

func TestEncodeMatrix_EmptyText(t *testing.T) {
	e := aztec.New()
	_, err := e.EncodeMatrix("")
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestEncode_CustomECC(t *testing.T) {
	e := aztec.New()
	e.MinECCPercent = 50
	img, err := e.Encode("Higher ECC", 100)
	if err != nil {
		t.Fatalf("Encode with custom ECC: %v", err)
	}
	if img.Bounds().Dx() != 100 {
		t.Error("image width should be 100")
	}
}

func TestEncode_CustomColors(t *testing.T) {
	e := aztec.New()
	e.ForegroundColor = color.RGBA{R: 255, A: 255} // red
	e.BackgroundColor = color.RGBA{B: 255, A: 255} // blue
	img, err := e.Encode("Colors", 100)
	if err != nil {
		t.Fatalf("Encode with custom colors: %v", err)
	}
	if img == nil {
		t.Error("image should not be nil")
	}
}

func TestEncode_LongText(t *testing.T) {
	e := aztec.New()
	long := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	img, err := e.Encode(long, 300)
	if err != nil {
		t.Fatalf("Encode long text: %v", err)
	}
	if img.Bounds().Dx() != 300 {
		t.Error("image width should be 300")
	}
}
