package image

// image_internal_test.go — internal package tests for uncovered branches that
// require access to unexported methods/fields of the image exporter.

import (
	"bytes"
	goimage "image"
	"image/color"
	"image/draw"
	"image/png"
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
)

// makePNG returns a w×h RGBA PNG encoded as bytes.
func makePNG(t *testing.T, w, h int, c color.RGBA) []byte {
	t.Helper()
	img := goimage.NewRGBA(goimage.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &goimage.Uniform{C: c}, goimage.Point{}, draw.Src)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("makePNG: %v", err)
	}
	return buf.Bytes()
}

// ── drawEllipse: w <= 0 || h <= 0 guard ──────────────────────────────────────

func TestDrawEllipse_ZeroWidth(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 100, 100))
	e.drawEllipse(10, 10, 0, 50, color.RGBA{A: 255}) // w == 0 → early return
}

func TestDrawEllipse_ZeroHeight(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 100, 100))
	e.drawEllipse(10, 10, 50, 0, color.RGBA{A: 255}) // h == 0 → early return
}

func TestDrawEllipse_NegativeWidth(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 100, 100))
	e.drawEllipse(10, 10, -5, 30, color.RGBA{A: 255}) // w < 0 → early return
}

// ── ExportBand: x1 <= x0 branch ──────────────────────────────────────────────
// When the canvas width is 0, x1 = pageW = 0 = x0 → the clamp sets x1 = 1.

func TestExportBand_ZeroWidthCanvas(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 0, 100)) // width == 0
	err := e.ExportBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	if err != nil {
		t.Errorf("ExportBand zero-width canvas: %v", err)
	}
}

// ── renderObject: nil curPage guard ──────────────────────────────────────────
// ExportBand checks curPage == nil and returns before calling renderObject, so
// the nil guard inside renderObject itself is never exercised from ExportBand.
// Call renderObject directly with curPage == nil to cover that branch.

func TestRenderObject_NilCurPage_Direct(t *testing.T) {
	e := NewExporter()
	e.curPage = nil
	obj := preview.PreparedObject{
		Name: "niltest", Kind: preview.ObjectTypeText,
		Left: 5, Top: 5, Width: 80, Height: 20,
		Text: "hello",
	}
	e.renderObject(obj, 0) // must not panic; returns immediately
}

// ── drawPictureObject: srcY/srcX clamping ────────────────────────────────────
// When the picture is placed at a negative offset, the dst rect starts at
// coordinates before x/y, so (px-x) or (py-y) can overflow the source bounds.

func TestDrawPictureObject_SrcYClampMax(t *testing.T) {
	// Source: 1×100 image placed at y=-1 in a 200×200 canvas.
	// dst starts at py=0, y=-1 → (py-y) = 1, scaleY = 100/1 = 100 → srcY=100 >= Max.Y → clamp.
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))

	pp := preview.New()
	pp.AddPage(200, 200, 1)
	blobIdx := pp.BlobStore.Add("syclamp", makePNG(t, 1, 100, color.RGBA{R: 200, A: 255}))
	e.pp = pp

	e.drawPictureObject(-1, -1, 1, 1, preview.PreparedObject{BlobIdx: blobIdx})
}

func TestDrawPictureObject_SrcXClampMax(t *testing.T) {
	// Source: 100×1 image placed at x=-1 in a 200×200 canvas.
	// dst starts at px=0, x=-1 → (px-x) = 1, scaleX = 100/1 = 100 → srcX=100 >= Max.X → clamp.
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))

	pp := preview.New()
	pp.AddPage(200, 200, 1)
	blobIdx := pp.BlobStore.Add("sxclamp", makePNG(t, 100, 1, color.RGBA{G: 200, A: 255}))
	e.pp = pp

	e.drawPictureObject(-1, -1, 1, 1, preview.PreparedObject{BlobIdx: blobIdx})
}
