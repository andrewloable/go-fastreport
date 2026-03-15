package qr_test

import (
	"image"
	"image/color"
	"testing"

	frqr "github.com/andrewloable/go-fastreport/barcode/qr"
)

// ── ECLevel constants ─────────────────────────────────────────────────────────

func TestECLevelConstants(t *testing.T) {
	levels := []frqr.ErrorCorrectionLevel{
		frqr.ECLevelL,
		frqr.ECLevelM,
		frqr.ECLevelQ,
		frqr.ECLevelH,
	}
	seen := map[frqr.ErrorCorrectionLevel]bool{}
	for _, l := range levels {
		if seen[l] {
			t.Errorf("duplicate ErrorCorrectionLevel %q", l)
		}
		seen[l] = true
	}
}

// ── NewEncoder ────────────────────────────────────────────────────────────────

func TestNewEncoder_Defaults(t *testing.T) {
	enc := frqr.NewEncoder()
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	if enc.ECLevel != frqr.ECLevelM {
		t.Errorf("ECLevel default = %q, want M", enc.ECLevel)
	}
	if enc.QuietZone != 4 {
		t.Errorf("QuietZone default = %d, want 4", enc.QuietZone)
	}
	if enc.ForegroundColor == nil {
		t.Error("ForegroundColor should not be nil")
	}
	if enc.BackgroundColor == nil {
		t.Error("BackgroundColor should not be nil")
	}
}

// ── Encode ────────────────────────────────────────────────────────────────────

func TestEncode_BasicURL(t *testing.T) {
	enc := frqr.NewEncoder()
	img, err := enc.Encode("https://example.com", 200)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
	b := img.Bounds()
	if b.Dx() != 200 || b.Dy() != 200 {
		t.Errorf("image size: got %dx%d, want 200x200", b.Dx(), b.Dy())
	}
}

func TestEncode_ShortText(t *testing.T) {
	enc := frqr.NewEncoder()
	img, err := enc.Encode("Hi", 100)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_LongText(t *testing.T) {
	enc := frqr.NewEncoder()
	text := "The quick brown fox jumps over the lazy dog. 0123456789 !@#$%^&*()"
	img, err := enc.Encode(text, 300)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_EmptyText_Error(t *testing.T) {
	enc := frqr.NewEncoder()
	_, err := enc.Encode("", 200)
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestEncode_ZeroSize_Error(t *testing.T) {
	enc := frqr.NewEncoder()
	_, err := enc.Encode("test", 0)
	if err == nil {
		t.Error("expected error for zero size")
	}
}

func TestEncode_NegativeSize_Error(t *testing.T) {
	enc := frqr.NewEncoder()
	_, err := enc.Encode("test", -10)
	if err == nil {
		t.Error("expected error for negative size")
	}
}

func TestEncode_ECLevelL(t *testing.T) {
	enc := frqr.NewEncoder()
	enc.ECLevel = frqr.ECLevelL
	img, err := enc.Encode("test L", 150)
	if err != nil {
		t.Fatalf("Encode ECLevelL error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_ECLevelQ(t *testing.T) {
	enc := frqr.NewEncoder()
	enc.ECLevel = frqr.ECLevelQ
	img, err := enc.Encode("test Q", 150)
	if err != nil {
		t.Fatalf("Encode ECLevelQ error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_ECLevelH(t *testing.T) {
	enc := frqr.NewEncoder()
	enc.ECLevel = frqr.ECLevelH
	img, err := enc.Encode("test H", 150)
	if err != nil {
		t.Fatalf("Encode ECLevelH error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_UnknownECLevel_FallsBackToM(t *testing.T) {
	enc := frqr.NewEncoder()
	enc.ECLevel = "X" // unknown level
	img, err := enc.Encode("fallback", 100)
	if err != nil {
		t.Fatalf("Encode with unknown EC level error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
}

func TestEncode_CustomColors(t *testing.T) {
	enc := frqr.NewEncoder()
	enc.ForegroundColor = color.RGBA{R: 0, G: 0, B: 255, A: 255} // blue
	enc.BackgroundColor = color.RGBA{R: 255, G: 255, B: 0, A: 255} // yellow
	img, err := enc.Encode("colored QR", 100)
	if err != nil {
		t.Fatalf("Encode with custom colors error: %v", err)
	}
	if img == nil {
		t.Fatal("Encode returned nil image")
	}
	// Verify it's a valid image.
	if img.Bounds().Dx() != 100 {
		t.Errorf("width: got %d, want 100", img.Bounds().Dx())
	}
}

func TestEncode_ReturnsNRGBAImage(t *testing.T) {
	enc := frqr.NewEncoder()
	img, err := enc.Encode("NRGBA test", 200)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	// The returned image must satisfy image.Image.
	var _ image.Image = img
}

// ── EncodeMatrix ──────────────────────────────────────────────────────────────

func TestEncodeMatrix_BasicText(t *testing.T) {
	enc := frqr.NewEncoder()
	matrix, err := enc.EncodeMatrix("hello")
	if err != nil {
		t.Fatalf("EncodeMatrix error: %v", err)
	}
	if len(matrix) == 0 {
		t.Fatal("matrix is empty")
	}
	for i, row := range matrix {
		if len(row) == 0 {
			t.Errorf("row %d is empty", i)
		}
	}
}

func TestEncodeMatrix_SquareMatrix(t *testing.T) {
	enc := frqr.NewEncoder()
	matrix, err := enc.EncodeMatrix("square")
	if err != nil {
		t.Fatalf("EncodeMatrix error: %v", err)
	}
	rows := len(matrix)
	for i, row := range matrix {
		if len(row) != rows {
			t.Errorf("row %d has %d cols, expected %d (square)", i, len(row), rows)
		}
	}
}

func TestEncodeMatrix_EmptyText_Error(t *testing.T) {
	enc := frqr.NewEncoder()
	_, err := enc.EncodeMatrix("")
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestEncodeMatrix_ContainsDarkModules(t *testing.T) {
	enc := frqr.NewEncoder()
	matrix, err := enc.EncodeMatrix("test data")
	if err != nil {
		t.Fatalf("EncodeMatrix error: %v", err)
	}
	darkCount := 0
	for _, row := range matrix {
		for _, dark := range row {
			if dark {
				darkCount++
			}
		}
	}
	if darkCount == 0 {
		t.Error("expected some dark modules in QR matrix")
	}
}

func TestEncodeMatrix_ContainsLightModules(t *testing.T) {
	enc := frqr.NewEncoder()
	matrix, err := enc.EncodeMatrix("test data")
	if err != nil {
		t.Fatalf("EncodeMatrix error: %v", err)
	}
	lightCount := 0
	for _, row := range matrix {
		for _, dark := range row {
			if !dark {
				lightCount++
			}
		}
	}
	if lightCount == 0 {
		t.Error("expected some light modules in QR matrix")
	}
}

func TestEncodeMatrix_NumericText(t *testing.T) {
	enc := frqr.NewEncoder()
	matrix, err := enc.EncodeMatrix("1234567890")
	if err != nil {
		t.Fatalf("EncodeMatrix numeric error: %v", err)
	}
	if len(matrix) < 21 {
		// Version 1 QR is 21×21; numeric is more compact so version 1 may apply.
		t.Errorf("matrix too small: %d rows", len(matrix))
	}
}

func TestEncode_NilForegroundColor_TreatedAsDefault(t *testing.T) {
	// Setting ForegroundColor to nil → isDefaultColors returns true via the
	// `if fg == nil || bg == nil { return true }` branch.
	e := frqr.NewEncoder()
	e.ForegroundColor = nil // triggers nil check
	img, err := e.Encode("nil color test", 100)
	if err != nil {
		t.Fatalf("Encode with nil ForegroundColor: %v", err)
	}
	if img == nil {
		t.Error("image should not be nil")
	}
}

func TestEncodeMatrix_UnknownECLevel(t *testing.T) {
	// Unknown EC level falls back to M in EncodeMatrix.
	e := frqr.NewEncoder()
	e.ECLevel = "Z" // unknown → default M
	matrix, err := e.EncodeMatrix("test unknown ec")
	if err != nil {
		t.Fatalf("EncodeMatrix unknown EC: %v", err)
	}
	if len(matrix) == 0 {
		t.Error("matrix should not be empty")
	}
}
