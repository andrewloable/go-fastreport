package code39_test

import (
	"image"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/code39"
)

func TestNewEncoder(t *testing.T) {
	enc := code39.NewEncoder()
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	if enc.AllowExtended {
		t.Error("AllowExtended should default to false")
	}
	if enc.CalcChecksum {
		t.Error("CalcChecksum should default to false")
	}
}

func TestEncode_UppercaseASCII(t *testing.T) {
	enc := code39.NewEncoder()
	img, err := enc.Encode("HELLO", 300, 100)
	if err != nil {
		t.Fatalf("Encode HELLO error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
	b := img.Bounds()
	if b.Dx() != 300 || b.Dy() != 100 {
		t.Errorf("image size: got %dx%d, want 300x100", b.Dx(), b.Dy())
	}
}

func TestEncode_LowercaseConverted(t *testing.T) {
	// Code 39 standard mode converts lowercase to uppercase.
	enc := code39.NewEncoder()
	img, err := enc.Encode("hello", 300, 100)
	if err != nil {
		t.Fatalf("Encode lowercase error: %v", err)
	}
	if img == nil {
		t.Fatal("expected image, got nil")
	}
}

func TestEncode_Numeric(t *testing.T) {
	enc := code39.NewEncoder()
	img, err := enc.Encode("0123456789", 400, 120)
	if err != nil {
		t.Fatalf("Encode numeric error: %v", err)
	}
	if img == nil {
		t.Fatal("expected image, got nil")
	}
}

func TestEncode_SpecialChars(t *testing.T) {
	enc := code39.NewEncoder()
	img, err := enc.Encode("AB-12.CD", 300, 100)
	if err != nil {
		t.Fatalf("Encode special chars error: %v", err)
	}
	if img == nil {
		t.Fatal("expected image, got nil")
	}
}

func TestEncode_WithChecksum(t *testing.T) {
	enc := code39.NewEncoder()
	enc.CalcChecksum = true
	img, err := enc.Encode("TEST", 300, 100)
	if err != nil {
		t.Fatalf("Encode with checksum error: %v", err)
	}
	if img == nil {
		t.Fatal("expected image, got nil")
	}
}

func TestEncode_EmptyText_Error(t *testing.T) {
	enc := code39.NewEncoder()
	_, err := enc.Encode("", 300, 100)
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestEncode_ZeroWidth_Error(t *testing.T) {
	enc := code39.NewEncoder()
	_, err := enc.Encode("TEST", 0, 100)
	if err == nil {
		t.Error("expected error for zero width")
	}
}

func TestEncode_ZeroHeight_Error(t *testing.T) {
	enc := code39.NewEncoder()
	_, err := enc.Encode("TEST", 300, 0)
	if err == nil {
		t.Error("expected error for zero height")
	}
}

func TestEncode_NegativeSize_Error(t *testing.T) {
	enc := code39.NewEncoder()
	_, err := enc.Encode("TEST", -100, -50)
	if err == nil {
		t.Error("expected error for negative size")
	}
}

func TestEncode_InvalidCharStandard_Error(t *testing.T) {
	enc := code39.NewEncoder()
	// Standard Code 39 does not support '@' character.
	_, err := enc.Encode("HELLO@WORLD", 300, 100)
	if err == nil {
		t.Error("expected error for invalid char in standard Code 39")
	}
}

func TestEncode_ExtendedAllowsSpecialChars(t *testing.T) {
	enc := code39.NewEncoder()
	enc.AllowExtended = true
	// Extended Code 39 supports full ASCII including '@'.
	img, err := enc.Encode("hello@world", 300, 100)
	if err != nil {
		t.Fatalf("Extended Code39 encode error: %v", err)
	}
	if img == nil {
		t.Fatal("expected image, got nil")
	}
}

func TestEncode_ReturnsImageInterface(t *testing.T) {
	enc := code39.NewEncoder()
	img, err := enc.Encode("TEST123", 200, 80)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	var _ image.Image = img
}

// ── Validate ──────────────────────────────────────────────────────────────────

func TestValidate_ValidChars(t *testing.T) {
	valid := []string{
		"HELLO",
		"12345",
		"ABC-123",
		"TEST.VALUE",
		"A B",
	}
	for _, v := range valid {
		if err := code39.Validate(v); err != nil {
			t.Errorf("Validate(%q): unexpected error: %v", v, err)
		}
	}
}

func TestValidate_EmptyText_Error(t *testing.T) {
	if err := code39.Validate(""); err == nil {
		t.Error("expected error for empty text")
	}
}

func TestValidate_InvalidAt_Error(t *testing.T) {
	if err := code39.Validate("HELLO@WORLD"); err == nil {
		t.Error("expected error for '@' in standard Code 39")
	}
}

func TestValidate_InvalidExclamation_Error(t *testing.T) {
	if err := code39.Validate("HELLO!"); err == nil {
		t.Error("expected error for '!' in standard Code 39")
	}
}

func TestValidate_LowercaseIsValid(t *testing.T) {
	// Validate normalises to uppercase, so lowercase is OK.
	if err := code39.Validate("hello"); err != nil {
		t.Errorf("expected lowercase to be valid in Validate, got: %v", err)
	}
}
