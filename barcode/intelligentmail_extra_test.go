package barcode_test

// Extra coverage tests for intelligentmail.go uncovered branches.
// These complement the tests in barcode_ext_coverage_test.go and barcode_test.go.
// NOTE: barcode_test.go already covers 25-digit and 29-digit Render paths.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ── imb_encode zip-length branches (Encode only) ─────────────────────────────

// 25-digit IMb: 20 tracking + 5 zip digits → zip length=5 → l = zipVal+1.
// barcode_test.go covers Render for 25-digit; here we test Encode separately.
func TestIntelligentMailBarcode_Encode_25Digits_Extra(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	// 20 tracking digits (second digit ≤ 4) + 5-digit ZIP = 25 total.
	const input = "0123456709498765432190210"
	if err := b.Encode(input); err != nil {
		t.Fatalf("Encode 25-digit IMb: %v", err)
	}
	if b.EncodedText() != input {
		t.Errorf("EncodedText = %q, want %q", b.EncodedText(), input)
	}
}

// 29-digit IMb: 20 tracking + 9 zip+4 digits → zip length=9 → l = zipVal+100001.
func TestIntelligentMailBarcode_Encode_29Digits_Extra(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	const input = "01234567094987654321902101234"
	if err := b.Encode(input); err != nil {
		t.Fatalf("Encode 29-digit IMb: %v", err)
	}
	if b.EncodedText() != input {
		t.Errorf("EncodedText = %q, want %q", b.EncodedText(), input)
	}
}

// ── Render: various dimension edge cases ──────────────────────────────────────

func TestIntelligentMailBarcode_Render_SmallWidth(t *testing.T) {
	// Very small width to exercise bar x1 clamping logic.
	b := barcode.NewIntelligentMailBarcode()
	if err := b.Encode("01234567094987654321"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(10, 60)
	if err != nil {
		t.Fatalf("Render small width: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil for small width")
	}
	if img.Bounds().Dx() != 10 {
		t.Errorf("width = %d, want 10", img.Bounds().Dx())
	}
}

func TestIntelligentMailBarcode_Render_VariousLengths(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"20-digit-extra", "01234567094987654321"},
		{"25-digit-extra", "0123456709498765432190210"},
		{"29-digit-extra", "01234567094987654321902101234"},
		{"31-digit-extra", "0123456709498765432112345678901"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := barcode.NewIntelligentMailBarcode()
			if err := b.Encode(tt.input); err != nil {
				t.Fatalf("Encode(%s): %v", tt.name, err)
			}
			img, err := b.Render(200, 60)
			if err != nil {
				t.Fatalf("Render(%s): %v", tt.name, err)
			}
			if img == nil {
				t.Fatalf("Render(%s) returned nil", tt.name)
			}
		})
	}
}

// ── imb_encode: invalid input lengths (rejected by Encode) ───────────────────

func TestIntelligentMailBarcode_Encode_InvalidLength_21Digits(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	// 21 digits — not a valid IMb length (must be 20/25/29/31).
	if err := b.Encode("012345670949876543211"); err == nil {
		t.Error("expected error for 21-digit input, got nil")
	}
}

func TestIntelligentMailBarcode_Encode_InvalidLength_24Digits(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	// 24 digits — not a valid IMb length.
	if err := b.Encode("012345670949876543219999"); err == nil {
		t.Error("expected error for 24-digit input, got nil")
	}
}

func TestIntelligentMailBarcode_Encode_InvalidLength_26Digits(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	// 26 digits — not a valid IMb length.
	if err := b.Encode("01234567094987654321902101"); err == nil {
		t.Error("expected error for 26-digit input, got nil")
	}
}

// ── Render: placeholderImage fallback ─────────────────────────────────────────
//
// Encode only validates digit count (20/25/29/31), NOT the second-digit range
// (must be 0-4). If second digit is 5-9, Encode succeeds but imb_encode fails
// inside Render, triggering the placeholderImage fallback path.

func TestIntelligentMailBarcode_Render_InvalidSecondDigit_Placeholder(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	// Second digit '5' > '4': passes Encode digit-count check (20 digits),
	// but imb_encode rejects it → triggers placeholderImage fallback in Render.
	if err := b.Encode("05234567094987654321"); err != nil {
		t.Fatalf("Encode unexpectedly rejected: %v", err)
	}
	// Render should NOT return an error (placeholder is returned).
	img, err := b.Render(200, 60)
	if err != nil {
		t.Fatalf("Render with bad second digit returned error: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestIntelligentMailBarcode_Render_InvalidSecondDigit_ZeroSize_Placeholder(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	// Same bad second digit with zero dimensions.
	if err := b.Encode("05234567094987654321"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(0, 0)
	if err != nil {
		t.Fatalf("Render(0,0) with bad digit: %v", err)
	}
	if img == nil {
		t.Fatal("Render(0,0) returned nil image")
	}
}
