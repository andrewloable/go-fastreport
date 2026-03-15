package image_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"testing"

	imgexport "github.com/andrewloable/go-fastreport/export/image"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func singlePageReport(width, height float32, bandHeight float32) *preview.PreparedPages {
	pp := preview.New()
	pp.AddPage(width, height, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "DataBand1",
		Top:    0,
		Height: bandHeight,
	})
	return pp
}

func twoPageReport() *preview.PreparedPages {
	pp := preview.New()
	for i := 0; i < 2; i++ {
		pp.AddPage(400, 300, i+1)
		_ = pp.AddBand(&preview.PreparedBand{
			Name:   "DataBand",
			Top:    float32(i * 10),
			Height: 20,
		})
	}
	return pp
}

// ── basic output ──────────────────────────────────────────────────────────────

func TestImageExport_ProducesPNG(t *testing.T) {
	pp := singlePageReport(400, 300, 50)
	exp := imgexport.NewExporter()

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("output is empty")
	}

	// Verify output is a valid PNG.
	_, err := png.Decode(&buf)
	if err != nil {
		t.Errorf("output is not valid PNG: %v", err)
	}
}

func TestImageExport_ImageDimensions(t *testing.T) {
	pp := singlePageReport(500, 400, 100)
	exp := imgexport.NewExporter()

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("PNG decode: %v", err)
	}

	if img.Bounds().Dx() != 500 {
		t.Errorf("image width = %d, want 500", img.Bounds().Dx())
	}
	if img.Bounds().Dy() != 400 {
		t.Errorf("image height = %d, want 400", img.Bounds().Dy())
	}
}

func TestImageExport_BackgroundIsWhite(t *testing.T) {
	pp := singlePageReport(100, 100, 10)
	exp := imgexport.NewExporter()

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	img, _ := png.Decode(bytes.NewReader(buf.Bytes()))
	// Bottom-right pixel should be white (not covered by the 10px band).
	r, g, b, _ := img.At(99, 99).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("background pixel is not white: r=%d g=%d b=%d", r>>8, g>>8, b>>8)
	}
}

// ── scale ─────────────────────────────────────────────────────────────────────

func TestImageExport_Scale(t *testing.T) {
	pp := singlePageReport(200, 150, 30)
	exp := imgexport.NewExporter()
	exp.Scale = 2.0

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("PNG decode: %v", err)
	}

	if img.Bounds().Dx() != 400 {
		t.Errorf("scaled width = %d, want 400", img.Bounds().Dx())
	}
	if img.Bounds().Dy() != 300 {
		t.Errorf("scaled height = %d, want 300", img.Bounds().Dy())
	}
}

// ── multi-page (sequential PNG output) ───────────────────────────────────────

func TestImageExport_MultiPage_NonEmpty(t *testing.T) {
	pp := twoPageReport()
	exp := imgexport.NewExporter()

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("multi-page output should not be empty")
	}
}

// ── nil/empty pages ───────────────────────────────────────────────────────────

func TestImageExport_NilPages_Error(t *testing.T) {
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(nil, &buf); err == nil {
		t.Error("expected error for nil PreparedPages")
	}
}

func TestImageExport_EmptyPages_NoError(t *testing.T) {
	pp := preview.New()
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Errorf("unexpected error for empty pages: %v", err)
	}
}

// ── custom colors ─────────────────────────────────────────────────────────────

func TestImageExport_CustomBackgroundColor(t *testing.T) {
	pp := singlePageReport(100, 100, 10)
	exp := imgexport.NewExporter()
	exp.BackgroundColor = color.RGBA{R: 0, G: 128, B: 0, A: 255} // green

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	img, _ := png.Decode(bytes.NewReader(buf.Bytes()))
	// Bottom-right pixel should be green (not covered by 10px band).
	r, g, b, _ := img.At(99, 99).RGBA()
	if r != 0 || g>>8 != 128 || b != 0 {
		t.Errorf("background pixel not green: r=%d g=%d b=%d", r>>8, g>>8, b>>8)
	}
}

// ── page range ────────────────────────────────────────────────────────────────

func TestImageExport_PageRangeCurrent(t *testing.T) {
	pp := twoPageReport()
	exp := imgexport.NewExporter()
	exp.PageRange = 1 // PageRangeCurrent
	exp.CurPage = 1   // first page

	var buf1 bytes.Buffer
	if err := exp.Export(pp, &buf1); err != nil {
		t.Fatalf("Export page 1: %v", err)
	}
	if buf1.Len() == 0 {
		t.Error("expected non-empty output for single page")
	}
}

// ── accessor defaults ─────────────────────────────────────────────────────────

func TestImageExport_DefaultScale(t *testing.T) {
	exp := imgexport.NewExporter()
	if exp.Scale != 1.0 {
		t.Errorf("Scale default = %v, want 1.0", exp.Scale)
	}
}

func TestImageExport_DefaultBackground(t *testing.T) {
	exp := imgexport.NewExporter()
	want := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if exp.BackgroundColor != want {
		t.Errorf("BackgroundColor = %v, want %v", exp.BackgroundColor, want)
	}
}

// ── object rendering helpers ──────────────────────────────────────────────────

// makePNGBlob creates a small single-color PNG and returns its encoded bytes.
func makePNGBlob(t *testing.T, width, height int, c color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("makePNGBlob: %v", err)
	}
	return buf.Bytes()
}

// pageWithObjectBand creates a PreparedPages with one page (200×100) and one
// band holding the supplied objects.
func pageWithObjectBand(objects []preview.PreparedObject) *preview.PreparedPages {
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:    "Band1",
		Top:     0,
		Height:  100,
		Objects: objects,
	})
	return pp
}

func exportOK(t *testing.T, pp *preview.PreparedPages) {
	t.Helper()
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
}

// ── renderObject: text / html / rtf ──────────────────────────────────────────

func TestRenderObject_Text_LeftAlign(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "t1", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text: "Hello",
			Font: style.Font{Size: 10, Name: "Arial"},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_CenterAlign(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "t2", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text:      "Center",
			Font:      style.Font{Size: 10},
			HorzAlign: 1, VertAlign: 1,
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_RightBottomAlign(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "t3", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text:      "Right",
			Font:      style.Font{Size: 10},
			HorzAlign: 2, VertAlign: 2,
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_WithFillColor(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "t4", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text:      "Filled",
			Font:      style.Font{Size: 10},
			FillColor: color.RGBA{R: 200, G: 200, B: 255, A: 255},
			TextColor: color.RGBA{A: 255},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_EmptyText(t *testing.T) {
	// Empty text returns early after background fill.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "t5", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text: "",
			Font: style.Font{Size: 10},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_MultilineAndWrap(t *testing.T) {
	// Text with newlines triggers wrapText's newline-split branch.
	// Narrow width triggers word-wrapping branch.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "t6", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 30, Height: 80,
			Text: "Line one\n\nThird line with many words that should wrap",
			Font: style.Font{Size: 8},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_WithBorderAllSides(t *testing.T) {
	// Setting all border lines visible triggers drawBorderLines for all 4 sides.
	bl := &style.BorderLine{Color: color.RGBA{A: 255}, Width: 1}
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "t7", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text: "Bordered",
			Font: style.Font{Size: 10},
			Border: style.Border{
				VisibleLines: style.BorderLinesAll,
				Lines:        [4]*style.BorderLine{bl, bl, bl, bl},
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_BorderZeroAlpha_FallsToBlack(t *testing.T) {
	// When a border line's Color.A == 0 drawBorderLines uses black.
	bl := &style.BorderLine{Color: color.RGBA{R: 0, G: 0, B: 0, A: 0}, Width: 1}
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "t8", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text: "BlackBorder",
			Font: style.Font{Size: 10},
			Border: style.Border{
				VisibleLines: style.BorderLinesAll,
				Lines:        [4]*style.BorderLine{bl, bl, bl, bl},
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Html(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "h1", Kind: preview.ObjectTypeHtml,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text: "<b>Bold</b>",
			Font: style.Font{Size: 10},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_RTF(t *testing.T) {
	// RTF objects: Kind gets rewritten to Text after stripping RTF.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "r1", Kind: preview.ObjectTypeRTF,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text: `{\rtf1\ansi Hello}`,
			Font: style.Font{Size: 10},
		},
	})
	exportOK(t, pp)
}

// ── font selection ────────────────────────────────────────────────────────────

func TestRenderObject_Text_MonoFont(t *testing.T) {
	// Font name "Courier New" → familyKeyword returns "mono".
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "tm", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text: "Mono",
			Font: style.Font{Size: 10, Name: "Courier New"},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_BoldFont(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "tb", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text: "Bold",
			Font: style.Font{Size: 10, Style: style.FontStyleBold},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_ItalicFont(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "ti", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text: "Italic",
			Font: style.Font{Size: 10, Style: style.FontStyleItalic},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_BoldItalicFont(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "tbi", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text: "BoldItalic",
			Font: style.Font{Size: 10, Style: style.FontStyleBold | style.FontStyleItalic},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_MonoBoldItalicFont(t *testing.T) {
	// Exercises mono bold-italic path in selectFace.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "tmbi", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text: "MonoBoldItalic",
			Font: style.Font{Size: 10, Name: "Consolas", Style: style.FontStyleBold | style.FontStyleItalic},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_CachedFont(t *testing.T) {
	// Second export with same font → hits cache in selectFace.
	obj := preview.PreparedObject{
		Name: "tc", Kind: preview.ObjectTypeText,
		Left: 5, Top: 5, Width: 80, Height: 20,
		Text: "Cached",
		Font: style.Font{Size: 12, Name: "Arial"},
	}
	exportOK(t, pageWithObjectBand([]preview.PreparedObject{obj}))
	exportOK(t, pageWithObjectBand([]preview.PreparedObject{obj})) // second call → cache hit
}

// ── renderObject: line ────────────────────────────────────────────────────────

func TestRenderObject_Line_Diagonal(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "ld", Kind: preview.ObjectTypeLine,
			Left: 5, Top: 5, Width: 60, Height: 40,
			LineDiagonal: true,
			Border: style.Border{
				Lines: [4]*style.BorderLine{{Color: color.RGBA{A: 255}, Width: 1}},
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Line_Horizontal(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "lh", Kind: preview.ObjectTypeLine,
			Left: 5, Top: 5, Width: 60, Height: 5,
			LineDiagonal: false,
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Line_Vertical(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "lv", Kind: preview.ObjectTypeLine,
			Left: 5, Top: 5, Width: 5, Height: 60,
			LineDiagonal: false,
		},
	})
	exportOK(t, pp)
}

// ── renderObject: shape ───────────────────────────────────────────────────────

func TestRenderObject_Shape_Rectangle(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "sr", Kind: preview.ObjectTypeShape,
			Left: 5, Top: 5, Width: 60, Height: 30,
			ShapeKind: 0,
			FillColor: color.RGBA{R: 200, G: 200, B: 255, A: 255},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Shape_RoundRect_WithCurve(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "srr", Kind: preview.ObjectTypeShape,
			Left: 5, Top: 5, Width: 60, Height: 30,
			ShapeKind: 1, ShapeCurve: 8,
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Shape_RoundRect_NoCurve(t *testing.T) {
	// ShapeCurve == 0 → falls back to drawRect.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "srr0", Kind: preview.ObjectTypeShape,
			Left: 5, Top: 5, Width: 60, Height: 30,
			ShapeKind: 1, ShapeCurve: 0,
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Shape_Ellipse(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "se", Kind: preview.ObjectTypeShape,
			Left: 5, Top: 5, Width: 60, Height: 30,
			ShapeKind: 2,
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Shape_Triangle(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "st", Kind: preview.ObjectTypeShape,
			Left: 5, Top: 5, Width: 60, Height: 40,
			ShapeKind: 3,
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Shape_Diamond(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "sd", Kind: preview.ObjectTypeShape,
			Left: 5, Top: 5, Width: 60, Height: 40,
			ShapeKind: 4,
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Shape_WithBorder(t *testing.T) {
	bl := &style.BorderLine{Color: color.RGBA{A: 255}, Width: 1}
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "sb", Kind: preview.ObjectTypeShape,
			Left: 5, Top: 5, Width: 60, Height: 30,
			ShapeKind: 0,
			Border: style.Border{
				VisibleLines: style.BorderLinesAll,
				Lines:        [4]*style.BorderLine{bl, bl, bl, bl},
			},
		},
	})
	exportOK(t, pp)
}

// ── renderObject: checkbox ────────────────────────────────────────────────────

func TestRenderObject_CheckBox_Checked(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "cbOn", Kind: preview.ObjectTypeCheckBox,
			Left: 5, Top: 5, Width: 20, Height: 20,
			Text: "true",
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_CheckBox_Unchecked(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "cbOff", Kind: preview.ObjectTypeCheckBox,
			Left: 5, Top: 5, Width: 20, Height: 20,
			Text: "false",
		},
	})
	exportOK(t, pp)
}

// ── renderObject: picture ─────────────────────────────────────────────────────

func TestRenderObject_Picture_ValidBlob(t *testing.T) {
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	blobIdx := pp.BlobStore.Add("pic1", makePNGBlob(t, 20, 15, color.RGBA{R: 255, A: 255}))
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band1",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name: "pic", Kind: preview.ObjectTypePicture,
				Left: 5, Top: 5, Width: 60, Height: 40,
				BlobIdx: blobIdx,
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Picture_NegativeBlobIdx(t *testing.T) {
	// BlobIdx < 0 → early return.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "picNeg", Kind: preview.ObjectTypePicture,
			Left: 5, Top: 5, Width: 60, Height: 40,
			BlobIdx: -1,
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Picture_EmptyBlob(t *testing.T) {
	// BlobIdx valid but no data stored → Get returns nil → early return.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band1",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name: "picEmpty", Kind: preview.ObjectTypePicture,
				Left: 5, Top: 5, Width: 60, Height: 40,
				BlobIdx: 999, // out-of-range → BlobStore.Get returns nil
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Picture_InvalidBlob(t *testing.T) {
	// Blob data exists but is not a valid image → Decode returns error → early return.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	blobIdx := pp.BlobStore.Add("badpic", []byte{0x00, 0x01, 0x02, 0x03})
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band1",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name: "picBad", Kind: preview.ObjectTypePicture,
				Left: 5, Top: 5, Width: 60, Height: 40,
				BlobIdx: blobIdx,
			},
		},
	})
	exportOK(t, pp)
}

// ── renderObject: polyline / polygon ─────────────────────────────────────────

func TestRenderObject_PolyLine_TwoPoints(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "pl", Kind: preview.ObjectTypePolyLine,
			Left: 5, Top: 5, Width: 60, Height: 40,
			Points: [][2]float32{{0, 0}, {50, 30}},
			Border: style.Border{
				Lines: [4]*style.BorderLine{{Color: color.RGBA{A: 255}, Width: 1}},
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_PolyLine_TooFewPoints(t *testing.T) {
	// < 2 points → early return (no-op).
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "plFew", Kind: preview.ObjectTypePolyLine,
			Left: 5, Top: 5, Width: 60, Height: 40,
			Points: [][2]float32{{0, 0}},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Polygon_ThreePoints(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "pg", Kind: preview.ObjectTypePolygon,
			Left: 5, Top: 5, Width: 60, Height: 40,
			Points: [][2]float32{{30, 0}, {60, 40}, {0, 40}},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Polygon_TooFewPoints(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "pgFew", Kind: preview.ObjectTypePolygon,
			Left: 5, Top: 5, Width: 60, Height: 40,
			Points: [][2]float32{{0, 0}},
		},
	})
	exportOK(t, pp)
}

// ── watermark ─────────────────────────────────────────────────────────────────

func TestWatermark_TextBehind(t *testing.T) {
	// ShowTextOnTop = false → text is rendered in ExportPageBegin.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "DRAFT",
		Font:         style.Font{Size: 24},
		TextColor:    color.RGBA{R: 128, G: 128, B: 128, A: 80},
		ShowTextOnTop: false,
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

func TestWatermark_TextOnTop(t *testing.T) {
	// ShowTextOnTop = true → text is rendered in ExportPageEnd.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "CONFIDENTIAL",
		Font:         style.Font{Size: 20},
		TextColor:    color.RGBA{A: 255},
		ShowTextOnTop: true,
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

func TestWatermark_EmptyText_Noop(t *testing.T) {
	// Empty watermark text → renderWatermarkTextOnPage returns early.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "",
		ShowTextOnTop: true,
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

func TestWatermark_Image_Stretch(t *testing.T) {
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	blobIdx := pp.BlobStore.Add("wm", makePNGBlob(t, 10, 10, color.RGBA{R: 255, A: 128}))
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ImageBlobIdx:   blobIdx,
		ImageSize:      preview.WatermarkImageSizeStretch,
		ShowImageOnTop: false,
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

func TestWatermark_Image_Center(t *testing.T) {
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	blobIdx := pp.BlobStore.Add("wmc", makePNGBlob(t, 10, 10, color.RGBA{G: 255, A: 128}))
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ImageBlobIdx:   blobIdx,
		ImageSize:      preview.WatermarkImageSizeCenter,
		ShowImageOnTop: true, // rendered in ExportPageEnd
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

func TestWatermark_Image_Tile(t *testing.T) {
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	blobIdx := pp.BlobStore.Add("wmt", makePNGBlob(t, 20, 20, color.RGBA{B: 255, A: 100}))
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ImageBlobIdx:   blobIdx,
		ImageSize:      preview.WatermarkImageSizeTile,
		ShowImageOnTop: false,
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

func TestWatermark_Image_NegativeBlobIdx(t *testing.T) {
	// ImageBlobIdx < 0 → renderWatermarkImageOnPage returns early.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ImageBlobIdx:   -1,
		ShowImageOnTop: true,
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

func TestWatermark_Image_EmptyBlobData(t *testing.T) {
	// BlobIdx points to index with no matching data → Get returns nil → early return.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ImageBlobIdx:   999, // out-of-range → empty
		ShowImageOnTop: false,
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

func TestWatermark_Image_InvalidBlobData(t *testing.T) {
	// Blob data is not a valid image → Decode fails → early return.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	blobIdx := pp.BlobStore.Add("wmBad", []byte{0x00, 0x01, 0x02})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ImageBlobIdx:   blobIdx,
		ShowImageOnTop: false,
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

func TestWatermark_Image_ZoomSize(t *testing.T) {
	// WatermarkImageSizeZoom falls into same branch as Center.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	blobIdx := pp.BlobStore.Add("wmz", makePNGBlob(t, 10, 10, color.RGBA{R: 100, G: 100, A: 255}))
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ImageBlobIdx:   blobIdx,
		ImageSize:      preview.WatermarkImageSizeZoom,
		ShowImageOnTop: false,
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

// ── wrapText edge cases ───────────────────────────────────────────────────────

func TestRenderObject_Text_SpacesOnlyLine(t *testing.T) {
	// A line of only spaces → fields is empty → appends "" branch in wrapText.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "spaces", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 40,
			Text: "line1\n   \nline3",
			Font: style.Font{Size: 10},
		},
	})
	exportOK(t, pp)
}

// ── additional branch coverage ────────────────────────────────────────────────

func TestExportPageBegin_ZeroDimensions(t *testing.T) {
	// When page width/height is 0, ExportPageBegin falls back to A4 dimensions.
	pp := preview.New()
	pp.AddPage(0, 0, 1) // both zero → both fallback branches
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

func TestExportBand_ZeroHeight(t *testing.T) {
	// Band height = 0 → y1 == y0 → y1 = y0+1 branch.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Zero", Top: 10, Height: 0})
	exportOK(t, pp)
}

func TestRenderObject_ZeroDimension(t *testing.T) {
	// Object with Width=0 and Height=0 → w<1 and h<1 → w=1, h=1 branches.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "zero", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 0, Height: 0,
			Text: "tiny",
			Font: style.Font{Size: 10},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_ZeroFontSize(t *testing.T) {
	// Font.Size == 0 → fontPt = 10 branch.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "zeroFont", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 80, Height: 20,
			Text: "NoFontSize",
			Font: style.Font{Size: 0},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_WordWrapFits(t *testing.T) {
	// Wide enough band → "Hello World" fits on one line → current = candidate branch.
	pp := preview.New()
	pp.AddPage(400, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "Band1", Top: 0, Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name: "ww", Kind: preview.ObjectTypeText,
				Left: 5, Top: 5, Width: 300, Height: 40,
				Text: "Hello World from the test suite",
				Font: style.Font{Size: 10},
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_TallTextOverflowsHeight(t *testing.T) {
	// Many lines overflow the object height → startY > y+h → break branch.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "tall", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 40, Height: 15, // very short height
			Text: "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10",
			Font: style.Font{Size: 10},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_OutOfBoundsObject(t *testing.T) {
	// Object entirely outside page → bounds.Empty() → early return.
	pp := preview.New()
	pp.AddPage(100, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "Band1", Top: 0, Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name: "oob", Kind: preview.ObjectTypeText,
				Left: 500, Top: 500, Width: 50, Height: 20,
				Text: "OutOfBounds",
				Font: style.Font{Size: 10},
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Shape_OutOfBounds(t *testing.T) {
	// Shape entirely outside page → bounds.Empty() → fill skip.
	pp := preview.New()
	pp.AddPage(100, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "Band1", Top: 0, Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name: "oobShape", Kind: preview.ObjectTypeShape,
				Left: 500, Top: 500, Width: 50, Height: 20,
				ShapeKind: 2,
				FillColor: color.RGBA{R: 255, A: 255},
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_CheckBox_OutOfBounds(t *testing.T) {
	// CheckBox entirely outside page → bounds.Empty() → early return.
	pp := preview.New()
	pp.AddPage(100, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "Band1", Top: 0, Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name: "oobCB", Kind: preview.ObjectTypeCheckBox,
				Left: 500, Top: 500, Width: 20, Height: 20,
				Text: "true",
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Shape_SmallEllipse_StepsMin(t *testing.T) {
	// Very small ellipse → steps < 8 → steps = 8 branch.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "smallE", Kind: preview.ObjectTypeShape,
			Left: 5, Top: 5, Width: 2, Height: 2,
			ShapeKind: 2,
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Shape_RoundRect_ClampedRadiusToZero(t *testing.T) {
	// Width=1, ShapeCurve=1: switch sends radius=1 to drawRoundRect;
	// inside: radius*2=2 > w=1 → radius=0 → e.drawRect fallback in drawRoundRect.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "rrSmall", Kind: preview.ObjectTypeShape,
			Left: 5, Top: 5, Width: 1, Height: 60,
			ShapeKind: 1, ShapeCurve: 1,
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Shape_RoundRect_StepsMin(t *testing.T) {
	// Small-radius roundrect → steps < 8 → steps = 8 branch in drawRoundRect.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "rrSteps", Kind: preview.ObjectTypeShape,
			Left: 5, Top: 5, Width: 30, Height: 30,
			ShapeKind: 1, ShapeCurve: 1, // radius=1 → steps = pi*1/2 ≈ 1 < 8 → steps = 8
		},
	})
	exportOK(t, pp)
}

func TestWatermark_TextZeroFontSize(t *testing.T) {
	// Watermark with Font.Size == 0 → fontPt = 48 branch.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "ZERO SIZE FONT",
		Font:         style.Font{Size: 0},
		ShowTextOnTop: true,
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

func TestExportBand_NilCurPage(t *testing.T) {
	// Calling ExportBand directly without ExportPageBegin → curPage == nil → returns nil.
	exp := imgexport.NewExporter()
	if err := exp.ExportBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10}); err != nil {
		t.Fatalf("ExportBand with nil curPage: %v", err)
	}
}

func TestExportPageEnd_NilCurPage(t *testing.T) {
	// Calling ExportPageEnd directly without ExportPageBegin → curPage == nil → returns nil.
	exp := imgexport.NewExporter()
	if err := exp.ExportPageEnd(nil); err != nil {
		t.Fatalf("ExportPageEnd with nil curPage: %v", err)
	}
}

func TestRenderObject_Line_WithBorderColor(t *testing.T) {
	// Line with border Lines[0] color set and A > 0 → lineColor = Lines[0].Color branch.
	bl := &style.BorderLine{Color: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Width: 2}
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "lc", Kind: preview.ObjectTypeLine,
			Left: 5, Top: 5, Width: 60, Height: 60,
			LineDiagonal: true,
			Border: style.Border{
				Lines: [4]*style.BorderLine{bl, nil, nil, nil},
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_PolyLine_WithBorderColor(t *testing.T) {
	// PolyLine with border Lines[0] color A > 0 → lineColor branch.
	bl := &style.BorderLine{Color: color.RGBA{G: 200, A: 255}, Width: 1}
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "plc", Kind: preview.ObjectTypePolyLine,
			Left: 5, Top: 5, Width: 80, Height: 60,
			Points: [][2]float32{{0, 0}, {40, 30}, {80, 0}},
			Border: style.Border{
				Lines: [4]*style.BorderLine{bl, nil, nil, nil},
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Polygon_WithBorderColor(t *testing.T) {
	// Polygon with border Lines[0] color A > 0 → lineColor branch.
	bl := &style.BorderLine{Color: color.RGBA{B: 200, A: 255}, Width: 1}
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "pgc", Kind: preview.ObjectTypePolygon,
			Left: 5, Top: 5, Width: 60, Height: 50,
			Points: [][2]float32{{30, 0}, {60, 50}, {0, 50}},
			Border: style.Border{
				Lines: [4]*style.BorderLine{bl, nil, nil, nil},
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Shape_RoundRect_ClampedByWidth(t *testing.T) {
	// radius*2 > w → radius = w/2 branch in drawRoundRect.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "rrCW", Kind: preview.ObjectTypeShape,
			Left: 5, Top: 5, Width: 10, Height: 60,
			ShapeKind: 1, ShapeCurve: 15, // radius=15, 15*2=30 > 10 → radius=5
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Shape_RoundRect_ClampedByHeight(t *testing.T) {
	// radius*2 > h → radius = h/2 branch in drawRoundRect.
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "rrCH", Kind: preview.ObjectTypeShape,
			Left: 5, Top: 5, Width: 80, Height: 6,
			ShapeKind: 1, ShapeCurve: 8, // radius=8, 8*2=16 > 6 → radius=3
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_NegativeTop_BorderOutOfBounds(t *testing.T) {
	// Object whose Top is negative → top border y < 0 → drawHLine/drawVLine out-of-bounds path.
	bl := &style.BorderLine{Color: color.RGBA{A: 255}, Width: 1}
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "Band1", Top: 0, Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name: "negTop", Kind: preview.ObjectTypeText,
				Left: -20, Top: -10, Width: 80, Height: 30,
				Text: "Partial",
				Font: style.Font{Size: 10},
				Border: style.Border{
					VisibleLines: style.BorderLinesAll,
					Lines:        [4]*style.BorderLine{bl, bl, bl, bl},
				},
			},
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Text_RightAlign_TextOverflows(t *testing.T) {
	// Very long text with right alignment → dotX < x → dotX = x branch.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "Band1", Top: 0, Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name: "overflow", Kind: preview.ObjectTypeText,
				Left: 5, Top: 5, Width: 20, Height: 20,
				Text:      "VeryVeryVeryLongTextThatOverflows",
				Font:      style.Font{Size: 12},
				HorzAlign: 2, // right align
			},
		},
	})
	exportOK(t, pp)
}

func TestWatermark_Image_Normal(t *testing.T) {
	// WatermarkImageSizeNormal → default case in switch → dstRect = natural size.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	blobIdx := pp.BlobStore.Add("wmn", makePNGBlob(t, 20, 20, color.RGBA{R: 100, A: 200}))
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ImageBlobIdx:   blobIdx,
		ImageSize:      preview.WatermarkImageSizeNormal,
		ShowImageOnTop: false,
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exportOK(t, pp)
}

// alwaysFailWriter is an io.Writer that always returns an error.
type alwaysFailWriter struct{}

func (w alwaysFailWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("intentional write failure")
}

func TestExportPageEnd_WriteError(t *testing.T) {
	// A failing io.Writer causes png.Encode to fail, covering the encode-error branch.
	pp := preview.New()
	pp.AddPage(200, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	exp := imgexport.NewExporter()
	err := exp.Export(pp, alwaysFailWriter{})
	if err == nil {
		t.Error("expected error from failing writer, got nil")
	}
}
