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

// ── ApplyGrayscale ────────────────────────────────────────────────────────────

func TestApplyGrayscale_Nil(t *testing.T) {
	result := ApplyGrayscale(nil)
	if result != nil {
		t.Error("ApplyGrayscale(nil) should return nil")
	}
}

func TestApplyGrayscale_ReducesChroma(t *testing.T) {
	// Saturated red image: all output pixels must be neutral grey.
	src := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	dst := ApplyGrayscale(src)
	if dst == nil {
		t.Fatal("ApplyGrayscale returned nil")
	}
	nrgba, ok := dst.(*image.NRGBA)
	if !ok {
		t.Fatal("ApplyGrayscale did not return *image.NRGBA")
	}
	b := dst.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			px := nrgba.NRGBAAt(x, y)
			if px.R != px.G || px.G != px.B {
				t.Errorf("pixel (%d,%d) is not grey: R=%d G=%d B=%d", x, y, px.R, px.G, px.B)
			}
		}
	}
}

func TestApplyGrayscale_PreservesAlpha(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	src.SetNRGBA(0, 0, color.NRGBA{R: 100, G: 150, B: 200, A: 128})
	src.SetNRGBA(1, 0, color.NRGBA{R: 100, G: 150, B: 200, A: 0})
	src.SetNRGBA(0, 1, color.NRGBA{R: 100, G: 150, B: 200, A: 255})
	src.SetNRGBA(1, 1, color.NRGBA{R: 100, G: 150, B: 200, A: 64})
	dst := ApplyGrayscale(src)
	nrgba, ok := dst.(*image.NRGBA)
	if !ok {
		t.Fatal("ApplyGrayscale did not return *image.NRGBA")
	}
	cases := []struct {
		x, y  int
		wantA uint8
	}{
		{0, 0, 128},
		{1, 0, 0},
		{0, 1, 255},
		{1, 1, 64},
	}
	for _, tc := range cases {
		px := nrgba.NRGBAAt(tc.x, tc.y)
		if px.A != tc.wantA {
			t.Errorf("pixel (%d,%d) alpha: got %d, want %d", tc.x, tc.y, px.A, tc.wantA)
		}
	}
}

func TestApplyGrayscale_NTSCWeights(t *testing.T) {
	// White (255,255,255) must remain white after grayscale.
	// NTSC lum = 255*0.299 + 255*0.587 + 255*0.114 = 254.999…; with math.Round → 255.
	src := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	dst := ApplyGrayscale(src)
	nrgba, ok := dst.(*image.NRGBA)
	if !ok {
		t.Fatal("ApplyGrayscale did not return *image.NRGBA")
	}
	px := nrgba.NRGBAAt(0, 0)
	if px.R != 255 || px.G != 255 || px.B != 255 {
		t.Errorf("white pixel not preserved: R=%d G=%d B=%d", px.R, px.G, px.B)
	}
}

// ── ApplyTransparency ─────────────────────────────────────────────────────────

func TestApplyTransparency_Nil(t *testing.T) {
	result := ApplyTransparency(nil, 0.5)
	if result != nil {
		t.Error("ApplyTransparency(nil) should return nil")
	}
}

func TestApplyTransparency_ZeroIsNoOp(t *testing.T) {
	pngBytes := buildTestPNG(t)
	src, _ := BytesToImage(pngBytes)
	dst := ApplyTransparency(src, 0)
	if dst != src {
		t.Error("ApplyTransparency with 0.0 should return src unchanged")
	}
}

func TestApplyTransparency_NegativeIsNoOp(t *testing.T) {
	pngBytes := buildTestPNG(t)
	src, _ := BytesToImage(pngBytes)
	dst := ApplyTransparency(src, -0.1)
	if dst != src {
		t.Error("ApplyTransparency with negative should return src unchanged")
	}
}

func TestApplyTransparency_HalfReducesAlpha(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 200})
		}
	}
	// factor = 0.5 → new alpha = uint8(200 * 0.5) = 100
	dst := ApplyTransparency(src, 0.5)
	nrgba, ok := dst.(*image.NRGBA)
	if !ok {
		t.Fatal("ApplyTransparency did not return *image.NRGBA")
	}
	px := nrgba.NRGBAAt(0, 0)
	if px.A < 95 || px.A > 105 {
		t.Errorf("alpha after 50%% transparency: got %d, want ~100", px.A)
	}
}

func TestApplyTransparency_FullMakesInvisible(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	dst := ApplyTransparency(src, 1.0) // factor = 0 → alpha = 0
	nrgba, ok := dst.(*image.NRGBA)
	if !ok {
		t.Fatal("ApplyTransparency did not return *image.NRGBA")
	}
	px := nrgba.NRGBAAt(0, 0)
	if px.A != 0 {
		t.Errorf("alpha after 100%% transparency: got %d, want 0", px.A)
	}
}

func TestApplyTransparency_PreservesRGB(t *testing.T) {
	// ApplyTransparency must not alter R, G, B — only the alpha channel changes.
	// Use NRGBAAt to read non-premultiplied values directly.
	src := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	src.SetNRGBA(0, 0, color.NRGBA{R: 100, G: 150, B: 200, A: 255})
	dst := ApplyTransparency(src, 0.5)
	nrgba, ok := dst.(*image.NRGBA)
	if !ok {
		t.Fatal("ApplyTransparency did not return *image.NRGBA")
	}
	px := nrgba.NRGBAAt(0, 0)
	if px.R != 100 || px.G != 150 || px.B != 200 {
		t.Errorf("RGB channels altered: R=%d G=%d B=%d", px.R, px.G, px.B)
	}
}
