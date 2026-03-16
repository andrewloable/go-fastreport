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
	"github.com/andrewloable/go-fastreport/style"
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

// ── drawPictureObject: srcY/srcX clamping via existing internal calls ─────────
// Note: The original SrcYClampMax/SrcXClampMax tests place the picture at (-1,-1)
// with w=1,h=1.  The destination rect Rect(-1,-1,0,0).Intersect(canvas) is empty
// so the pixel loop never runs.  These tests exercise the pre-loop setup code but
// not the clamping arms.  The clamping branches (srcY<Min, srcY>=Max, srcX<Min,
// srcX>=Max) require srcBounds.Min to be non-zero, which is impossible via PNG
// decoding.  The tests below cover what IS reachable.

func TestDrawPictureObject_SrcYClampMax(t *testing.T) {
	// Source: 1×100 image placed at y=-1 in a 200×200 canvas.
	// dst = Rect(-1,-1,0,0).Intersect(canvas) = empty → loop does not run.
	// This exercises the setup path but not the clamp arms.
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
	// dst = Rect(-1,-1,0,0).Intersect(canvas) = empty → loop does not run.
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))

	pp := preview.New()
	pp.AddPage(200, 200, 1)
	blobIdx := pp.BlobStore.Add("sxclamp", makePNG(t, 100, 1, color.RGBA{G: 200, A: 255}))
	e.pp = pp

	e.drawPictureObject(-1, -1, 1, 1, preview.PreparedObject{BlobIdx: blobIdx})
}

// ── drawPictureObject: w==0 and h==0 destination (zero-dimension dst guard) ───
// The condition `srcBounds.Dx()==0 || srcBounds.Dy()==0 || w==0 || h==0` at
// line 390 fires for w==0 or h==0. renderObject normalises object dimensions to
// >= 1 before calling drawPictureObject, so these branches are only reachable
// by calling the method directly.

func TestDrawPictureObject_ZeroWidth(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))
	pp := preview.New()
	pp.AddPage(200, 200, 1)
	blobIdx := pp.BlobStore.Add("zpicw", makePNG(t, 10, 10, color.RGBA{R: 200, A: 255}))
	e.pp = pp
	// w == 0 → hits the `w == 0` arm of the zero-dimension guard → early return
	e.drawPictureObject(10, 10, 0, 20, preview.PreparedObject{BlobIdx: blobIdx})
}

func TestDrawPictureObject_ZeroHeight(t *testing.T) {
	e := NewExporter()
	e.curPage = goimage.NewRGBA(goimage.Rect(0, 0, 200, 200))
	pp := preview.New()
	pp.AddPage(200, 200, 1)
	blobIdx := pp.BlobStore.Add("zpich", makePNG(t, 10, 10, color.RGBA{G: 200, A: 255}))
	e.pp = pp
	// h == 0 → hits the `h == 0` arm of the zero-dimension guard → early return
	e.drawPictureObject(10, 10, 20, 0, preview.PreparedObject{BlobIdx: blobIdx})
}

// ── selectFace: opentype.Parse error branch ───────────────────────────────────
// Replace the font family data with corrupt bytes so opentype.Parse fails,
// triggering the error fallback to basicfont.Face7x13.  We use unique (sizePt,
// dpi) pairs to avoid hitting the cache from previous test runs.

func TestSelectFace_ParseError_Regular(t *testing.T) {
	orig := fontFamilies["sans"]
	fontFamilies["sans"] = goFontData{
		regular:    []byte("not a valid ttf"),
		bold:       []byte("not a valid ttf"),
		italic:     []byte("not a valid ttf"),
		boldItalic: []byte("not a valid ttf"),
	}
	defer func() { fontFamilies["sans"] = orig }()

	f := style.Font{Name: "Arial", Style: 0}
	face := selectFace(f, 991.1, 991.1)
	if face == nil {
		t.Error("selectFace parse error (regular): expected basicfont fallback, got nil")
	}
}

func TestSelectFace_ParseError_Bold(t *testing.T) {
	orig := fontFamilies["sans"]
	fontFamilies["sans"] = goFontData{
		regular:    []byte("corrupt"),
		bold:       []byte("corrupt"),
		italic:     []byte("corrupt"),
		boldItalic: []byte("corrupt"),
	}
	defer func() { fontFamilies["sans"] = orig }()

	f := style.Font{Name: "Arial", Style: style.FontStyleBold}
	face := selectFace(f, 992.1, 992.1)
	if face == nil {
		t.Error("selectFace parse error (bold): expected basicfont fallback, got nil")
	}
}

func TestSelectFace_ParseError_Italic(t *testing.T) {
	orig := fontFamilies["sans"]
	fontFamilies["sans"] = goFontData{
		regular:    []byte("corrupt"),
		bold:       []byte("corrupt"),
		italic:     []byte("corrupt"),
		boldItalic: []byte("corrupt"),
	}
	defer func() { fontFamilies["sans"] = orig }()

	f := style.Font{Name: "Arial", Style: style.FontStyleItalic}
	face := selectFace(f, 993.1, 993.1)
	if face == nil {
		t.Error("selectFace parse error (italic): expected basicfont fallback, got nil")
	}
}

func TestSelectFace_ParseError_BoldItalic(t *testing.T) {
	orig := fontFamilies["sans"]
	fontFamilies["sans"] = goFontData{
		regular:    []byte("corrupt"),
		bold:       []byte("corrupt"),
		italic:     []byte("corrupt"),
		boldItalic: []byte("corrupt"),
	}
	defer func() { fontFamilies["sans"] = orig }()

	f := style.Font{Name: "Arial", Style: style.FontStyleBold | style.FontStyleItalic}
	face := selectFace(f, 994.1, 994.1)
	if face == nil {
		t.Error("selectFace parse error (bold+italic): expected basicfont fallback, got nil")
	}
}

func TestSelectFace_ParseError_Mono(t *testing.T) {
	orig := fontFamilies["mono"]
	fontFamilies["mono"] = goFontData{
		regular:    []byte("corrupt mono"),
		bold:       []byte("corrupt mono"),
		italic:     []byte("corrupt mono"),
		boldItalic: []byte("corrupt mono"),
	}
	defer func() { fontFamilies["mono"] = orig }()

	f := style.Font{Name: "Courier New", Style: 0}
	face := selectFace(f, 990.1, 990.1)
	if face == nil {
		t.Error("selectFace parse error (mono regular): expected basicfont fallback, got nil")
	}
}

// ── selectFace: opentype.NewFace error branch ─────────────────────────────────
// After a successful opentype.Parse, NewFace can fail if the FaceOptions are
// invalid.  We call selectFace with a negative sizePt to provoke a NewFace error.
// A unique (sizePt, dpi) pair ensures a cache miss.

func TestSelectFace_NewFaceError_NegativeSizePt(t *testing.T) {
	// Use a highly negative sizePt with a unique DPI to force a cache miss.
	// opentype.NewFace should return an error for a non-positive size.
	f := style.Font{Name: "Arial", Style: 0}
	face := selectFace(f, -42.0, 7777.0)
	// Whether it errors or not, the function must return a non-nil face.
	if face == nil {
		t.Error("selectFace negative sizePt: got nil face")
	}
}

func TestSelectFace_NewFaceError_NegativeDPI(t *testing.T) {
	f := style.Font{Name: "Arial", Style: 0}
	face := selectFace(f, 12.0, -1.0)
	if face == nil {
		t.Error("selectFace negative dpi: got nil face")
	}
}

func TestSelectFace_NewFaceError_ZeroDPI(t *testing.T) {
	f := style.Font{Name: "Arial", Style: 0}
	face := selectFace(f, 12.0, 0.0)
	if face == nil {
		t.Error("selectFace zero dpi: got nil face")
	}
}
