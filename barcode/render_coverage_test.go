package barcode_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ---------------------------------------------------------------------------
// Code39Barcode.Render (barcode.go:177) — 0% coverage
// ---------------------------------------------------------------------------

func TestCode39Barcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewCode39Barcode()
	if err := b.Encode("CODE39"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestCode39Barcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewCode39Barcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// QRBarcode.Render (barcode.go:213) — 0% coverage
// ---------------------------------------------------------------------------

func TestQRBarcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewQRBarcode()
	if err := b.Encode("https://example.com"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestQRBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewQRBarcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// EAN13Barcode.Render (barcode.go:468) — 0% coverage
// ---------------------------------------------------------------------------

func TestEAN13Barcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewEAN13Barcode()
	if err := b.Encode("590123412345"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestEAN13Barcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewEAN13Barcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// AztecBarcode.Render (barcode.go:515) — 0% coverage
// ---------------------------------------------------------------------------

func TestAztecBarcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewAztecBarcode()
	if err := b.Encode("Aztec"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestAztecBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewAztecBarcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// PDF417Barcode.Render (barcode.go:558) — 0% coverage
// ---------------------------------------------------------------------------

func TestPDF417Barcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewPDF417Barcode()
	if err := b.Encode("PDF417"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestPDF417Barcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewPDF417Barcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// EAN8Barcode.Render (missing_types.go:56) — 0% coverage
// ---------------------------------------------------------------------------

func TestEAN8Barcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewEAN8Barcode()
	if err := b.Encode("1234567"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestEAN8Barcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewEAN8Barcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// UPCEBarcode.Render (missing_types.go:147) — 0% coverage
// ---------------------------------------------------------------------------

func TestUPCEBarcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewUPCEBarcode()
	if err := b.Encode("1234567"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestUPCEBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewUPCEBarcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// Code128ABarcode.Render (missing_types.go:217) — 0% coverage
// ---------------------------------------------------------------------------

func TestCode128ABarcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewCode128ABarcode()
	if err := b.Encode("CODE128A"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestCode128ABarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewCode128ABarcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// Code128BBarcode.Render (missing_types.go:251) — 0% coverage
// ---------------------------------------------------------------------------

func TestCode128BBarcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewCode128BBarcode()
	if err := b.Encode("Code128B"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestCode128BBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewCode128BBarcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// Code128CBarcode.Render (missing_types.go:283) — 0% coverage
// ---------------------------------------------------------------------------

func TestCode128CBarcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewCode128CBarcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestCode128CBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewCode128CBarcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// Supplement2Barcode.Render (missing_types.go:527) — 0% coverage
// ---------------------------------------------------------------------------

func TestSupplement2Barcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewSupplement2Barcode()
	if err := b.Encode("53"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestSupplement2Barcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewSupplement2Barcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// Supplement5Barcode.Render (missing_types.go:567) — 0% coverage
// ---------------------------------------------------------------------------

func TestSupplement5Barcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewSupplement5Barcode()
	if err := b.Encode("52495"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestSupplement5Barcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewSupplement5Barcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// Code39ExtendedBarcode.Render (missing_types.go:606) — 0% coverage
// ---------------------------------------------------------------------------

func TestCode39ExtendedBarcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewCode39ExtendedBarcode()
	if err := b.Encode("abc-1234"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestCode39ExtendedBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewCode39ExtendedBarcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// UPCE0Barcode.Render (missing_types.go:635) — 0% coverage
// ---------------------------------------------------------------------------

func TestUPCE0Barcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewUPCE0Barcode()
	if err := b.Encode("01234565"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestUPCE0Barcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewUPCE0Barcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// UPCE1Barcode.Render (missing_types.go:664) — 0% coverage
// ---------------------------------------------------------------------------

func TestUPCE1Barcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewUPCE1Barcode()
	if err := b.Encode("11234565"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestUPCE1Barcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewUPCE1Barcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// GS1_128Barcode.Render (missing_types.go:693) — 0% coverage
// ---------------------------------------------------------------------------

func TestGS1_128Barcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewGS1_128Barcode()
	if err := b.Encode("(01)12345678901231"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestGS1_128Barcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewGS1_128Barcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// GS1DatamatrixBarcode.Render (missing_types.go:722) — 0% coverage
// ---------------------------------------------------------------------------

func TestGS1DatamatrixBarcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewGS1DatamatrixBarcode()
	if err := b.Encode("(01)12345678901231"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestGS1DatamatrixBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewGS1DatamatrixBarcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ---------------------------------------------------------------------------
// JapanPost4StateBarcode.Render (missing_types.go:749) — 0% coverage
// ---------------------------------------------------------------------------

func TestJapanPost4StateBarcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewJapanPost4StateBarcode()
	if err := b.Encode("597-8615-5-7-6"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}
}

func TestJapanPost4StateBarcode_Render_NotEncoded(t *testing.T) {
	b := barcode.NewJapanPost4StateBarcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}
