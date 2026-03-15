package utils

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// buildTestPNG creates a minimal valid PNG image (8×8 red square) as bytes.
func buildTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("buildTestPNG: %v", err)
	}
	return buf.Bytes()
}

// ── BytesToImage ──────────────────────────────────────────────────────────────

func TestBytesToImage_Valid(t *testing.T) {
	pngBytes := buildTestPNG(t)
	img, err := BytesToImage(pngBytes)
	if err != nil {
		t.Fatalf("BytesToImage: %v", err)
	}
	if img == nil {
		t.Fatal("BytesToImage returned nil image")
	}
	if img.Bounds().Dx() != 8 || img.Bounds().Dy() != 8 {
		t.Errorf("unexpected bounds: %v", img.Bounds())
	}
}

func TestBytesToImage_Invalid(t *testing.T) {
	_, err := BytesToImage([]byte("not an image"))
	if err == nil {
		t.Error("expected error for invalid image bytes")
	}
}

// ── ImageToBytes ──────────────────────────────────────────────────────────────

func TestImageToBytes_PNG(t *testing.T) {
	pngBytes := buildTestPNG(t)
	img, _ := BytesToImage(pngBytes)
	out, err := ImageToBytes(img, ImageFormatPNG)
	if err != nil {
		t.Fatalf("ImageToBytes PNG: %v", err)
	}
	if len(out) == 0 {
		t.Error("PNG output is empty")
	}
	// Re-decode to verify it's valid PNG
	img2, err := BytesToImage(out)
	if err != nil {
		t.Fatalf("re-decode PNG: %v", err)
	}
	if img2.Bounds() != img.Bounds() {
		t.Errorf("bounds mismatch: got %v, want %v", img2.Bounds(), img.Bounds())
	}
}

func TestImageToBytes_JPEG(t *testing.T) {
	pngBytes := buildTestPNG(t)
	img, _ := BytesToImage(pngBytes)
	out, err := ImageToBytes(img, ImageFormatJPEG)
	if err != nil {
		t.Fatalf("ImageToBytes JPEG: %v", err)
	}
	if len(out) == 0 {
		t.Error("JPEG output is empty")
	}
}

// ── ResizeImage ───────────────────────────────────────────────────────────────

func TestResizeImage_Nil(t *testing.T) {
	result := ResizeImage(nil, 100, 100, SizeModeNormal)
	if result != nil {
		t.Error("nil src should return nil")
	}
}

func TestResizeImage_ZeroDimension(t *testing.T) {
	pngBytes := buildTestPNG(t)
	src, _ := BytesToImage(pngBytes)
	// zero width
	result := ResizeImage(src, 0, 100, SizeModeNormal)
	if result != src {
		t.Error("zero width should return src unchanged")
	}
	// zero height
	result = ResizeImage(src, 100, 0, SizeModeNormal)
	if result != src {
		t.Error("zero height should return src unchanged")
	}
}

func TestResizeImage_Stretch(t *testing.T) {
	pngBytes := buildTestPNG(t)
	src, _ := BytesToImage(pngBytes)
	dst := ResizeImage(src, 32, 16, SizeModeStretchImage)
	if dst == nil {
		t.Fatal("ResizeImage returned nil")
	}
	if dst.Bounds().Dx() != 32 || dst.Bounds().Dy() != 16 {
		t.Errorf("stretch: got bounds %v, want 32×16", dst.Bounds())
	}
}

func TestResizeImage_Center_SmallerSrc(t *testing.T) {
	pngBytes := buildTestPNG(t) // 8×8
	src, _ := BytesToImage(pngBytes)
	dst := ResizeImage(src, 64, 64, SizeModeCenterImage)
	if dst == nil {
		t.Fatal("ResizeImage returned nil")
	}
	if dst.Bounds().Dx() != 64 || dst.Bounds().Dy() != 64 {
		t.Errorf("center small: got bounds %v, want 64×64", dst.Bounds())
	}
}

func TestResizeImage_Center_LargerSrc(t *testing.T) {
	// Build a 64×64 image and center into 8×8
	img := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	var buf bytes.Buffer
	png.Encode(&buf, img) //nolint:errcheck
	src, _ := BytesToImage(buf.Bytes())
	dst := ResizeImage(src, 8, 8, SizeModeCenterImage)
	if dst == nil {
		t.Fatal("ResizeImage returned nil")
	}
	if dst.Bounds().Dx() != 8 || dst.Bounds().Dy() != 8 {
		t.Errorf("center large: got bounds %v, want 8×8", dst.Bounds())
	}
}

func TestResizeImage_Zoom(t *testing.T) {
	pngBytes := buildTestPNG(t)
	src, _ := BytesToImage(pngBytes)
	dst := ResizeImage(src, 100, 50, SizeModeZoom)
	if dst == nil {
		t.Fatal("ResizeImage returned nil")
	}
	if dst.Bounds().Dx() != 100 || dst.Bounds().Dy() != 50 {
		t.Errorf("zoom: got bounds %v, want 100×50", dst.Bounds())
	}
}

func TestResizeImage_Normal(t *testing.T) {
	pngBytes := buildTestPNG(t)
	src, _ := BytesToImage(pngBytes)
	dst := ResizeImage(src, 32, 32, SizeModeNormal)
	if dst == nil {
		t.Fatal("ResizeImage returned nil")
	}
	if dst.Bounds().Dx() != 32 || dst.Bounds().Dy() != 32 {
		t.Errorf("normal: got bounds %v, want 32×32", dst.Bounds())
	}
}

func TestResizeImage_AutoSize(t *testing.T) {
	pngBytes := buildTestPNG(t)
	src, _ := BytesToImage(pngBytes)
	// SizeModeAutoSize falls through to normal (default)
	dst := ResizeImage(src, 20, 20, SizeModeAutoSize)
	if dst == nil {
		t.Fatal("ResizeImage returned nil")
	}
}

// ── LoadImage — file path ─────────────────────────────────────────────────────

func TestLoadImage_FilePath(t *testing.T) {
	dir := t.TempDir()
	pngBytes := buildTestPNG(t)
	path := filepath.Join(dir, "test.png")
	if err := os.WriteFile(path, pngBytes, 0o644); err != nil {
		t.Fatalf("write temp PNG: %v", err)
	}
	img, err := LoadImage(path)
	if err != nil {
		t.Fatalf("LoadImage(file): %v", err)
	}
	if img == nil {
		t.Fatal("LoadImage returned nil")
	}
}

// ── LoadImage — empty source ──────────────────────────────────────────────────

func TestLoadImage_EmptySource(t *testing.T) {
	_, err := LoadImage("")
	if err == nil {
		t.Error("expected error for empty source")
	}
}

// ── LoadImage — data URI ──────────────────────────────────────────────────────

func TestLoadImage_DataURI(t *testing.T) {
	pngBytes := buildTestPNG(t)
	b64 := base64.StdEncoding.EncodeToString(pngBytes)
	uri := "data:image/png;base64," + b64
	img, err := LoadImage(uri)
	if err != nil {
		t.Fatalf("LoadImage(data URI): %v", err)
	}
	if img == nil {
		t.Fatal("LoadImage returned nil for data URI")
	}
}

func TestLoadImage_DataURI_URLSafe(t *testing.T) {
	pngBytes := buildTestPNG(t)
	b64 := base64.URLEncoding.EncodeToString(pngBytes)
	uri := "data:image/png;base64," + b64
	img, err := LoadImage(uri)
	if err != nil {
		t.Fatalf("LoadImage(URL-safe data URI): %v", err)
	}
	if img == nil {
		t.Fatal("LoadImage returned nil")
	}
}

func TestLoadImage_DataURI_Malformed(t *testing.T) {
	_, err := LoadImage("data:image/png;base64") // no comma
	if err == nil {
		t.Error("expected error for malformed data URI")
	}
}

func TestLoadImage_DataURI_BadBase64(t *testing.T) {
	_, err := LoadImage("data:image/png;base64,!!!not-base64!!!")
	if err == nil {
		t.Error("expected error for bad base64 in data URI")
	}
}

// ── LoadImage — raw base64 fallback ──────────────────────────────────────────

func TestLoadImage_Base64Fallback_Invalid(t *testing.T) {
	_, err := LoadImage("not-a-valid-path-or-base64!!")
	if err == nil {
		t.Error("expected error for invalid path/base64")
	}
}

func TestLoadImage_Base64Fallback_Valid(t *testing.T) {
	pngBytes := buildTestPNG(t)
	b64 := base64.StdEncoding.EncodeToString(pngBytes)
	img, err := LoadImage(b64)
	if err != nil {
		t.Fatalf("LoadImage(base64 fallback): %v", err)
	}
	if img == nil {
		t.Fatal("LoadImage returned nil for base64")
	}
}
