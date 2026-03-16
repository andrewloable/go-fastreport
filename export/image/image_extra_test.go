package image_test

// image_extra_test.go — additional tests to increase coverage in
// export/image for selectFace, drawPictureObject, drawEllipse, and ExportBand.

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"testing"

	imgexport "github.com/andrewloable/go-fastreport/export/image"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// ── selectFace: error branches ────────────────────────────────────────────────
// The opentype.Parse error path is not reachable with valid Go fonts, but the
// opentype.NewFace error path can be triggered by a zero DPI (which maps to a
// face options error on some platforms) or by a zero-size font. We can also
// exercise the fallback by using a very large or very small point size.
// Since the font bytes are always valid, we focus on the cache-miss paths
// that call opentype.Parse + opentype.NewFace for every (family, bold, italic,
// size, dpi) combination that wasn't seen before.

// TestSelectFace_AllVariants exercises all four bold/italic combinations for
// both "sans" and "mono" families to maximise the uncovered branches in selectFace.
func TestSelectFace_AllVariants(t *testing.T) {
	// Use unique font sizes so each call is a cache miss.
	sizes := []float32{7, 9, 11, 13, 15, 17, 19, 21, 23, 25}
	styles := []style.FontStyle{
		0,
		style.FontStyleBold,
		style.FontStyleItalic,
		style.FontStyleBold | style.FontStyleItalic,
	}
	fontNames := []string{
		"Arial",        // → sans
		"Courier New",  // → mono
		"Consolas",     // → mono
		"Monaco",       // → mono
		"Unknown Font", // → falls through to sans
	}

	for _, sz := range sizes {
		for _, st := range styles {
			for _, nm := range fontNames {
				obj := preview.PreparedObject{
					Name:  "vary",
					Kind:  preview.ObjectTypeText,
					Left:  5,
					Top:   5,
					Width: 80,
					Height: 20,
					Text:  "variant",
					Font:  style.Font{Size: sz, Name: nm, Style: st},
				}
				pp := pageWithObjectBand([]preview.PreparedObject{obj})
				exp := imgexport.NewExporter()
				var buf bytes.Buffer
				if err := exp.Export(pp, &buf); err != nil {
					t.Errorf("Export (name=%q size=%v style=%v): %v", nm, sz, st, err)
				}
			}
		}
	}
}

// TestSelectFace_ZeroSizeFont exercises the fontPt = 10 fallback for font size 0.
func TestSelectFace_ZeroSizeFont(t *testing.T) {
	obj := preview.PreparedObject{
		Name:  "zfsv",
		Kind:  preview.ObjectTypeText,
		Left:  5,
		Top:   5,
		Width: 80,
		Height: 20,
		Text:  "zero",
		Font:  style.Font{Size: 0, Name: ""},
	}
	pp := pageWithObjectBand([]preview.PreparedObject{obj})
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Errorf("Export zero font: %v", err)
	}
}

// ── ExportBand: x1 <= x0 branch ──────────────────────────────────────────────
// ExportBand has a branch: if x1 <= x0 { x1 = x0 + 1 }. This is guarded by
// pageW (the canvas width) being <= 0.  We can reach it by forcing Export to
// create a zero-width canvas (Scale = 0 forces scaled(w) = w unchanged, but
// a very tiny page w=0 is already tested).  The other way is to call ExportBand
// directly against a 0-width canvas.  We'll use a variant of the existing
// TestExportBand_ZeroHeight but with a wider page to cover the y1 <= y0 path
// more explicitly.

func TestExportBand_NarrowPage(t *testing.T) {
	// Page width == 1 so x1 = 1 and x0 = 0 → no clamp. Just ensure it doesn't panic.
	pp := preview.New()
	pp.AddPage(1, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Narrow", Top: 0, Height: 50})
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Errorf("Export narrow page: %v", err)
	}
}

func TestExportBand_WithObjectsOnNarrowPage(t *testing.T) {
	// A picture on a narrow page exercises the clamp branches in drawPictureObject.
	pp := preview.New()
	pp.AddPage(10, 100, 1)
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 200, A: 255}}, image.Point{}, draw.Src)
	var raw bytes.Buffer
	png.Encode(&raw, img) //nolint:errcheck
	blobIdx := pp.BlobStore.Add("small", raw.Bytes())
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "B",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name:    "pic",
				Kind:    preview.ObjectTypePicture,
				Left:    0,
				Top:     0,
				Width:   5,
				Height:  5,
				BlobIdx: blobIdx,
			},
		},
	})
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Errorf("Export narrow page with picture: %v", err)
	}
}

// ── drawPictureObject: zero-dimension source image ────────────────────────────
// Covers the `srcBounds.Dx() == 0 || srcBounds.Dy() == 0` branch.
// We cannot easily build a 0×0 PNG, but we can trigger the w==0 || h==0 path
// by using an object with Width=0 (which the renderObject code normalises to 1
// before calling drawPictureObject — so this path is actually unreachable from
// the public API). Instead, test the boundary where src is valid but dst is tiny.

func TestDrawPictureObject_TinyDst(t *testing.T) {
	// 1×1 destination.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{G: 200, A: 255}}, image.Point{}, draw.Src)
	var raw bytes.Buffer
	png.Encode(&raw, img) //nolint:errcheck
	blobIdx := pp.BlobStore.Add("p2", raw.Bytes())
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "B",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name:    "pic1x1",
				Kind:    preview.ObjectTypePicture,
				Left:    0,
				Top:     0,
				Width:   1,
				Height:  1,
				BlobIdx: blobIdx,
			},
		},
	})
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Errorf("Export 1×1 picture: %v", err)
	}
}

// TestDrawPictureObject_ClampedSrcCoords tests the srcX/srcY clamping branches
// inside drawPictureObject (the min/max guard on source coordinates).
func TestDrawPictureObject_ClampedSrcCoords(t *testing.T) {
	// A picture object larger than the source image so scaling reverses the
	// clamp: srcY/srcX computation overflows the source bounds → clamp to max-1.
	pp := preview.New()
	pp.AddPage(200, 200, 1)
	// 2×2 source image scaled into 100×100 destination.
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{R: 100, A: 255})
	img.SetRGBA(1, 0, color.RGBA{G: 100, A: 255})
	img.SetRGBA(0, 1, color.RGBA{B: 100, A: 255})
	img.SetRGBA(1, 1, color.RGBA{R: 50, G: 50, A: 255})
	var raw bytes.Buffer
	png.Encode(&raw, img) //nolint:errcheck
	blobIdx := pp.BlobStore.Add("p3", raw.Bytes())
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "B",
		Top:    0,
		Height: 200,
		Objects: []preview.PreparedObject{
			{
				Name:    "bigpic",
				Kind:    preview.ObjectTypePicture,
				Left:    0,
				Top:     0,
				Width:   100,
				Height:  100,
				BlobIdx: blobIdx,
			},
		},
	})
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Errorf("Export big picture: %v", err)
	}
}

// ── drawEllipse: w<=0 or h<=0 guard ──────────────────────────────────────────
// drawEllipse returns early when w <= 0 || h <= 0.
// The public API normalises object dims to >= 1, so we must test via a custom
// shape object with explicit zero dimensions. Since renderObject ensures w >= 1
// and h >= 1, the drawEllipse guard is only reachable from a renderObject call
// with Width=0/Height=0 after normalisation to 1. The guard is for safety so
// let's exercise it via the exported renderObject path with ShapeKind==2
// and Width/Height = 0 (which normalise to 1, which is still > 0).
// To hit w<=0 directly we call Export on a report with a zero-dim shape.

func TestDrawEllipse_ViaNormalisation(t *testing.T) {
	// Shape with Width=0, Height=0 → w/h normalised to 1 inside renderObject
	// → drawEllipse(x, y, 1, 1, …) → w > 0 so guard not hit, but the path runs.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name:      "ellipseZero",
			Kind:      preview.ObjectTypeShape,
			Left:      5,
			Top:       5,
			Width:     0,
			Height:    0,
			ShapeKind: 2,
		},
	})
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Errorf("Export zero ellipse: %v", err)
	}
}

// TestDrawEllipse_LargeEllipse tests a large ellipse with many steps.
func TestDrawEllipse_LargeEllipse(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name:      "bigEllipse",
			Kind:      preview.ObjectTypeShape,
			Left:      5,
			Top:       5,
			Width:     180,
			Height:    80,
			ShapeKind: 2,
			FillColor: color.RGBA{R: 100, G: 200, B: 100, A: 255},
		},
	})
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Errorf("Export large ellipse: %v", err)
	}
}

// ── renderWatermarkImageOnPage: tile with zero-size image ─────────────────────
// The branch `if sw <= 0 || sh <= 0 { return }` inside renderWatermarkImageOnPage
// when ImageSize == Tile. A zero-dimension image cannot be decoded by the
// standard PNG decoder, so we hit this through a very small tile image instead.

func TestWatermark_Tile_LargeImage(t *testing.T) {
	// Large tile image (larger than page) → tile loop runs once and exits.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	img := image.NewRGBA(image.Rect(0, 0, 300, 200))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 50, A: 100}}, image.Point{}, draw.Src)
	var raw bytes.Buffer
	png.Encode(&raw, img) //nolint:errcheck
	blobIdx := pp.BlobStore.Add("bigTile", raw.Bytes())
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ImageBlobIdx:   blobIdx,
		ImageSize:      preview.WatermarkImageSizeTile,
		ShowImageOnTop: false,
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Errorf("Export large tile watermark: %v", err)
	}
}

// ── ExportBand x1 <= x0 and x0 > pageW ───────────────────────────────────────
// When the page width is 0, ExportBand uses the fallback A4 width so x0/x1 are
// always 0/794. There's no practical way to make x1 < x0 via the public API.
// Instead we exercise the BandFillColor path on a multi-band page.

func TestExportBand_MultipleBands(t *testing.T) {
	pp := preview.New()
	pp.AddPage(200, 200, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B1", Top: 0, Height: 50})
	_ = pp.AddBand(&preview.PreparedBand{Name: "B2", Top: 50, Height: 50})
	_ = pp.AddBand(&preview.PreparedBand{Name: "B3", Top: 100, Height: 50})
	exp := imgexport.NewExporter()
	exp.BandFillColor = color.RGBA{R: 200, G: 220, B: 255, A: 255}
	exp.BandBorderColor = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Errorf("Export multi-band: %v", err)
	}
}

// ── renderObject: RTF with empty stripped text ────────────────────────────────
func TestRenderObject_RTF_EmptyAfterStrip(t *testing.T) {
	// RTF that strips to empty string → empty-text early return.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name:  "rtfEmpty",
			Kind:  preview.ObjectTypeRTF,
			Left:  5,
			Top:   5,
			Width: 80,
			Height: 20,
			Text:  `{\rtf1\ansi}`, // no visible text
			Font:  style.Font{Size: 10},
		},
	})
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Errorf("Export RTF empty: %v", err)
	}
}
