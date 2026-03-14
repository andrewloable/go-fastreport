package code128_test

import (
	"image"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/code128"
)

func TestNewEncoder(t *testing.T) {
	enc := code128.NewEncoder()
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
}

func TestEncode_Basic(t *testing.T) {
	enc := code128.NewEncoder()
	img, err := enc.Encode("HELLO123", 300, 100)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
	b := img.Bounds()
	if b.Dx() != 300 || b.Dy() != 100 {
		t.Errorf("image size: got %dx%d, want 300x100", b.Dx(), b.Dy())
	}
}

func TestEncode_NumericOnly(t *testing.T) {
	enc := code128.NewEncoder()
	img, err := enc.Encode("1234567890", 400, 120)
	if err != nil {
		t.Fatalf("Encode numeric error: %v", err)
	}
	if img == nil {
		t.Fatal("expected image, got nil")
	}
}

func TestEncode_LowerCase(t *testing.T) {
	enc := code128.NewEncoder()
	img, err := enc.Encode("hello world", 300, 100)
	if err != nil {
		t.Fatalf("Encode lowercase error: %v", err)
	}
	if img == nil {
		t.Fatal("expected image, got nil")
	}
}

func TestEncode_SpecialChars(t *testing.T) {
	enc := code128.NewEncoder()
	img, err := enc.Encode("ABC-123 xyz", 300, 100)
	if err != nil {
		t.Fatalf("Encode special chars error: %v", err)
	}
	if img == nil {
		t.Fatal("expected image, got nil")
	}
}

func TestEncode_EmptyText_Error(t *testing.T) {
	enc := code128.NewEncoder()
	_, err := enc.Encode("", 300, 100)
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestEncode_ZeroWidth_Error(t *testing.T) {
	enc := code128.NewEncoder()
	_, err := enc.Encode("test", 0, 100)
	if err == nil {
		t.Error("expected error for zero width")
	}
}

func TestEncode_ZeroHeight_Error(t *testing.T) {
	enc := code128.NewEncoder()
	_, err := enc.Encode("test", 300, 0)
	if err == nil {
		t.Error("expected error for zero height")
	}
}

func TestEncode_NegativeSize_Error(t *testing.T) {
	enc := code128.NewEncoder()
	_, err := enc.Encode("test", -100, -50)
	if err == nil {
		t.Error("expected error for negative size")
	}
}

func TestEncode_ReturnsImageInterface(t *testing.T) {
	enc := code128.NewEncoder()
	img, err := enc.Encode("test", 200, 80)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	var _ image.Image = img
}

func TestEncode_LargeBarcode(t *testing.T) {
	enc := code128.NewEncoder()
	text := strings.Repeat("ABC123", 10)
	img, err := enc.Encode(text, 800, 200)
	if err != nil {
		t.Fatalf("Encode large error: %v", err)
	}
	if img == nil {
		t.Fatal("expected image, got nil")
	}
}

func TestValidate_ValidText(t *testing.T) {
	valid := []string{
		"HELLO",
		"12345",
		"ABC 123",
		"hello world",
		"!@#$%",
	}
	for _, v := range valid {
		if err := code128.Validate(v); err != nil {
			t.Errorf("Validate(%q): unexpected error: %v", v, err)
		}
	}
}

func TestValidate_EmptyText_Error(t *testing.T) {
	if err := code128.Validate(""); err == nil {
		t.Error("expected error for empty text")
	}
}

func TestValidate_InvalidHighUnicode(t *testing.T) {
	// Characters above 0x7E are outside Code 128 range.
	if err := code128.Validate("hello\u00E9"); err == nil {
		t.Error("expected error for non-ASCII character")
	}
}

func TestValidate_DEL_Invalid(t *testing.T) {
	// DEL (0x7F) is outside printable range.
	if err := code128.Validate("test\x7F"); err == nil {
		t.Error("expected error for DEL character (0x7F)")
	}
}
