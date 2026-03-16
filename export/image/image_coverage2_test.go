package image

// image_coverage2_test.go — additional internal tests for export/image to
// target specific branches not reachable through the public API.

import (
	"bytes"
	goimage "image"
	"image/color"
	"image/draw"
	"image/png"
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// ── drawPictureObject: nil pp guard ──────────────────────────────────────────

func TestDrawPictureObject_NilPP(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))
	e.pp = nil // nil PreparedPages
	// BlobIdx >= 0 but pp is nil → early return.
	e.drawPictureObject(0, 0, 100, 100, preview.PreparedObject{BlobIdx: 0})
	// Must not panic.
}

// ── drawPictureObject: invalid image data (image.Decode error) ───────────────

func TestDrawPictureObject_InvalidImageData(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))

	pp := preview.New()
	pp.AddPage(200, 200, 1)
	blobIdx := pp.BlobStore.Add("bad_img", []byte("not an image"))
	e.pp = pp

	// image.Decode fails → early return.
	e.drawPictureObject(0, 0, 100, 100, preview.PreparedObject{BlobIdx: blobIdx})
}

// ── drawPictureObject: srcY >= Max clamping (scale-up path) ──────────────────
// When a small source image is scaled up into a large destination, the srcY
// calculation (srcY = Min.Y + int(float64(py-y)*scaleY)) can produce values
// >= srcBounds.Max.Y for the last pixel row, triggering the Max clamp.

func TestDrawPictureObject_ScaleUpClamping(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))

	pp := preview.New()
	pp.AddPage(200, 200, 1)
	// 1×1 source image → scale up to 50×50 destination.
	// scaleY = 1.0/50.0 ≈ 0.02. For py at the last rows, srcY can reach or
	// exceed 1 (srcBounds.Max.Y), triggering the >= Max clamp at line 404-406.
	img := goimage.NewRGBA(goimage.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	blobIdx := pp.BlobStore.Add("1x1", buf.Bytes())
	e.pp = pp

	// Scale 1×1 → 50×50. The last destination pixel py=49 gives
	// srcY = int(float64(49)*scaleY) = int(0.98) = 0 (within bounds).
	// Actually for scaleY=1/50: srcY = int(49 * 0.02) = int(0.98) = 0.
	// To trigger srcY >= Max we need scaleY such that srcY reaches 1.
	// Use h=1 destination so scaleY = 1.0/1.0 = 1.0. Then py=y=0 → srcY=0,
	// and at the boundary (scaleY=1), no clamping is needed.
	// Better: h > srcBounds.Dy() so scaleY < 1. But srcY = int(py * scaleY) < 1.
	// The Max clamp requires scaleY * (py-y) >= srcBounds.Max.Y. That means
	// py-y >= srcBounds.Dy()/scaleY = srcBounds.Dy() * h/srcBounds.Dy() = h.
	// Since py < y+h, py-y < h, so py-y never reaches h. The Max clamp can only
	// fire due to floating-point rounding, not in the clean case.
	//
	// Therefore this test exercises the pixel loop without triggering the Max clamp.
	// It still provides additional coverage for the normal pixel-copy path.
	e.drawPictureObject(0, 0, 50, 50, preview.PreparedObject{BlobIdx: blobIdx})
}

// ── drawPictureObject: srcX Max clamping via scale-up ────────────────────────

func TestDrawPictureObject_ScaleUpXClamping(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))

	pp := preview.New()
	pp.AddPage(200, 200, 1)
	// 2×1 source image, scale to 100×1 destination.
	img := goimage.NewRGBA(goimage.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{G: 200, A: 255})
	img.SetRGBA(1, 0, color.RGBA{B: 200, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	blobIdx := pp.BlobStore.Add("2x1", buf.Bytes())
	e.pp = pp

	// Scale 2×1 → 100×1. scaleX = 2/100 = 0.02. For px=99:
	// srcX = int(99 * 0.02) = int(1.98) = 1 (within bounds, Max.X=2).
	// The Max clamp fires when srcX >= srcBounds.Max.X = 2.
	// That requires px * 0.02 >= 2, i.e. px >= 100, but px < 100. So no clamp.
	// However we still exercise the pixel copy path.
	e.drawPictureObject(0, 0, 100, 1, preview.PreparedObject{BlobIdx: blobIdx})
}

// ── renderWatermarkImageOnPage: nil pp guard ─────────────────────────────────

func TestRenderWatermarkImageOnPage_NilPP(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))
	e.pp = nil

	wm := &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: 0,
	}
	// pp is nil → early return.
	e.renderWatermarkImageOnPage(wm)
}

// ── renderWatermarkImageOnPage: nil curPage guard ────────────────────────────

func TestRenderWatermarkImageOnPage_NilCurPage(t *testing.T) {
	e := NewExporter()
	e.curPage = nil

	pp := preview.New()
	pp.AddPage(200, 200, 1)
	e.pp = pp

	wm := &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: 0,
	}
	// curPage is nil → early return.
	e.renderWatermarkImageOnPage(wm)
}

// ── renderWatermarkImageOnPage: negative ImageBlobIdx ────────────────────────

func TestRenderWatermarkImageOnPage_NegativeBlobIdx(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))

	pp := preview.New()
	pp.AddPage(200, 200, 1)
	e.pp = pp

	wm := &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: -1, // negative → early return
	}
	e.renderWatermarkImageOnPage(wm)
}

// ── renderWatermarkImageOnPage: empty blob ───────────────────────────────────

func TestRenderWatermarkImageOnPage_EmptyBlob(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))

	pp := preview.New()
	pp.AddPage(200, 200, 1)
	blobIdx := pp.BlobStore.Add("empty", []byte{})
	e.pp = pp

	wm := &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
	}
	// Empty blob → early return.
	e.renderWatermarkImageOnPage(wm)
}

// ── renderWatermarkImageOnPage: invalid image data (Decode error) ─────────────

func TestRenderWatermarkImageOnPage_DecodeError(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))

	pp := preview.New()
	pp.AddPage(200, 200, 1)
	blobIdx := pp.BlobStore.Add("bad", []byte("not an image at all"))
	e.pp = pp

	wm := &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
	}
	// image.Decode fails → early return.
	e.renderWatermarkImageOnPage(wm)
}

// ── renderWatermarkImageOnPage: all ImageSize variants ───────────────────────

func makeSmallPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := goimage.NewRGBA(goimage.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &goimage.Uniform{C: color.RGBA{R: 128, A: 200}}, goimage.Point{}, draw.Src)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("makeSmallPNG: %v", err)
	}
	return buf.Bytes()
}

func TestRenderWatermarkImageOnPage_Stretch(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 100, 100))

	pp := preview.New()
	pp.AddPage(100, 100, 1)
	blobIdx := pp.BlobStore.Add("stretch", makeSmallPNG(t, 20, 20))
	e.pp = pp

	wm := &preview.PreparedWatermark{
		Enabled:            true,
		ImageBlobIdx:       blobIdx,
		ImageSize:          preview.WatermarkImageSizeStretch,
		ImageTransparency: 0.5,
	}
	e.renderWatermarkImageOnPage(wm)
}

func TestRenderWatermarkImageOnPage_Center(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 100, 100))

	pp := preview.New()
	pp.AddPage(100, 100, 1)
	blobIdx := pp.BlobStore.Add("center", makeSmallPNG(t, 30, 30))
	e.pp = pp

	wm := &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeCenter,
	}
	e.renderWatermarkImageOnPage(wm)
}

func TestRenderWatermarkImageOnPage_Zoom(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 100, 100))

	pp := preview.New()
	pp.AddPage(100, 100, 1)
	blobIdx := pp.BlobStore.Add("zoom", makeSmallPNG(t, 40, 40))
	e.pp = pp

	wm := &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeZoom,
	}
	e.renderWatermarkImageOnPage(wm)
}

func TestRenderWatermarkImageOnPage_Normal(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 100, 100))

	pp := preview.New()
	pp.AddPage(100, 100, 1)
	blobIdx := pp.BlobStore.Add("normal", makeSmallPNG(t, 50, 50))
	e.pp = pp

	wm := &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeNormal, // default branch
	}
	e.renderWatermarkImageOnPage(wm)
}

func TestRenderWatermarkImageOnPage_Tile_Small(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 100, 100))

	pp := preview.New()
	pp.AddPage(100, 100, 1)
	// 10×10 tile on 100×100 canvas → tile loop runs 10×10 times.
	blobIdx := pp.BlobStore.Add("tile_small", makeSmallPNG(t, 10, 10))
	e.pp = pp

	wm := &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeTile,
	}
	e.renderWatermarkImageOnPage(wm)
}

// ── selectFace: cache hit path ────────────────────────────────────────────────

func TestSelectFace_CacheHit(t *testing.T) {
	f := style.Font{Name: "Arial", Style: 0}
	// First call: cache miss → parse and create face.
	face1 := selectFace(f, 11.0, 96.0)
	// Second call: cache hit → return stored face.
	face2 := selectFace(f, 11.0, 96.0)
	if face1 != face2 {
		t.Error("selectFace cache hit should return the same face object")
	}
}

// ── selectFace: mono font bold+italic path ────────────────────────────────────

func TestSelectFace_Mono_BoldItalic(t *testing.T) {
	f := style.Font{Name: "Courier New", Style: style.FontStyleBold | style.FontStyleItalic}
	face := selectFace(f, 13.5, 96.0)
	if face == nil {
		t.Error("selectFace mono bold+italic returned nil")
	}
}

func TestSelectFace_Mono_Bold(t *testing.T) {
	f := style.Font{Name: "Courier New", Style: style.FontStyleBold}
	face := selectFace(f, 14.5, 96.0)
	if face == nil {
		t.Error("selectFace mono bold returned nil")
	}
}

func TestSelectFace_Mono_Italic(t *testing.T) {
	f := style.Font{Name: "Courier New", Style: style.FontStyleItalic}
	face := selectFace(f, 15.5, 96.0)
	if face == nil {
		t.Error("selectFace mono italic returned nil")
	}
}

func TestSelectFace_Mono_Regular(t *testing.T) {
	f := style.Font{Name: "Courier New", Style: 0}
	face := selectFace(f, 16.5, 96.0)
	if face == nil {
		t.Error("selectFace mono regular returned nil")
	}
}

func TestSelectFace_Sans_BoldItalic(t *testing.T) {
	f := style.Font{Name: "Arial", Style: style.FontStyleBold | style.FontStyleItalic}
	face := selectFace(f, 17.5, 96.0)
	if face == nil {
		t.Error("selectFace sans bold+italic returned nil")
	}
}
