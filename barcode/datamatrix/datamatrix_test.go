package datamatrix_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/datamatrix"
)

func TestNew_Defaults(t *testing.T) {
	e := datamatrix.New()
	if e.ForegroundColor != color.Black {
		t.Error("ForegroundColor should default to black")
	}
	if e.BackgroundColor != color.White {
		t.Error("BackgroundColor should default to white")
	}
}

func TestEncode_Basic(t *testing.T) {
	e := datamatrix.New()
	img, err := e.Encode("Hello DataMatrix", 200, 200)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 200 || b.Dy() != 200 {
		t.Errorf("image size = %dx%d, want 200x200", b.Dx(), b.Dy())
	}
}

func TestEncode_EmptyText(t *testing.T) {
	e := datamatrix.New()
	_, err := e.Encode("", 200, 200)
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestEncode_ZeroWidth(t *testing.T) {
	e := datamatrix.New()
	_, err := e.Encode("test", 0, 100)
	if err == nil {
		t.Error("expected error for zero width")
	}
}

func TestEncode_ZeroHeight(t *testing.T) {
	e := datamatrix.New()
	_, err := e.Encode("test", 100, 0)
	if err == nil {
		t.Error("expected error for zero height")
	}
}

func TestEncodeMatrix_Basic(t *testing.T) {
	e := datamatrix.New()
	matrix, err := e.EncodeMatrix("Test123")
	if err != nil {
		t.Fatalf("EncodeMatrix: %v", err)
	}
	if len(matrix) == 0 {
		t.Error("matrix should not be empty")
	}
}

func TestEncodeMatrix_EmptyText(t *testing.T) {
	e := datamatrix.New()
	_, err := e.EncodeMatrix("")
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestEncode_CustomColors(t *testing.T) {
	e := datamatrix.New()
	e.ForegroundColor = color.RGBA{G: 255, A: 255} // green
	e.BackgroundColor = color.RGBA{R: 255, A: 255} // red
	img, err := e.Encode("Colors", 100, 100)
	if err != nil {
		t.Fatalf("Encode with custom colors: %v", err)
	}
	if img == nil {
		t.Error("image should not be nil")
	}
}

func TestEncode_NonSquare(t *testing.T) {
	e := datamatrix.New()
	img, err := e.Encode("Hello", 300, 150)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 300 || b.Dy() != 150 {
		t.Errorf("image size = %dx%d, want 300x150", b.Dx(), b.Dy())
	}
}

func TestEncode_LongText(t *testing.T) {
	e := datamatrix.New()
	long := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	img, err := e.Encode(long, 200, 200)
	if err != nil {
		t.Fatalf("Encode long text: %v", err)
	}
	if img.Bounds().Dx() != 200 {
		t.Error("image width should be 200")
	}
}
