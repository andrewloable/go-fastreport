package xlsx

import (
	"testing"
)

// ── imageExtension (white-box tests for unexported function) ─────────────────

func TestImageExtension_JPEG(t *testing.T) {
	// JPEG magic bytes: FF D8 ...
	data := []byte{0xFF, 0xD8, 0x00, 0x00}
	if got := imageExtension(data); got != ".jpg" {
		t.Errorf("JPEG: want .jpg, got %s", got)
	}
}

func TestImageExtension_PNG(t *testing.T) {
	// PNG magic bytes: 89 50 4E 47 0D 0A 1A 0A
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if got := imageExtension(data); got != ".png" {
		t.Errorf("PNG: want .png, got %s", got)
	}
}

func TestImageExtension_GIF(t *testing.T) {
	// GIF magic bytes: 47 49 46 (GIF)
	data := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}
	if got := imageExtension(data); got != ".gif" {
		t.Errorf("GIF: want .gif, got %s", got)
	}
}

func TestImageExtension_BMP(t *testing.T) {
	// BMP magic bytes: 42 4D (BM)
	data := []byte{0x42, 0x4D, 0x00, 0x00}
	if got := imageExtension(data); got != ".bmp" {
		t.Errorf("BMP: want .bmp, got %s", got)
	}
}

func TestImageExtension_Unknown_DefaultsPNG(t *testing.T) {
	// Unknown magic bytes → default to .png
	data := []byte{0x00, 0x01, 0x02, 0x03}
	if got := imageExtension(data); got != ".png" {
		t.Errorf("unknown: want .png (default), got %s", got)
	}
}

func TestImageExtension_TooShort_DefaultsPNG(t *testing.T) {
	// Single byte → len < 2 → default
	data := []byte{0xFF}
	if got := imageExtension(data); got != ".png" {
		t.Errorf("too short: want .png (default), got %s", got)
	}
}

func TestImageExtension_Empty_DefaultsPNG(t *testing.T) {
	// Empty slice
	data := []byte{}
	if got := imageExtension(data); got != ".png" {
		t.Errorf("empty: want .png (default), got %s", got)
	}
}

func TestImageExtension_PNGNeedsEightBytes(t *testing.T) {
	// Only first 4 bytes provided with PNG-like first byte but not 8 bytes
	// 0x89 0x50 matches PNG check (len >= 8 required), but only 4 bytes here.
	// len(data) < 8 so PNG won't match. BMP/GIF/JPEG also don't match.
	// Falls through to default.
	data := []byte{0x89, 0x50, 0x00, 0x00}
	if got := imageExtension(data); got != ".png" {
		t.Errorf("short PNG-like: want .png (default), got %s", got)
	}
}

func TestImageExtension_GIFNeedsThreeBytes(t *testing.T) {
	// GIF check needs len >= 3 (G=0x47, I=0x49, F=0x46).
	// If only 2 bytes matching GIF prefix, it falls through.
	data := []byte{0x47, 0x49}
	if got := imageExtension(data); got != ".png" {
		t.Errorf("short GIF-like: want .png (default), got %s", got)
	}
}
