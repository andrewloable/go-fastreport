package pdf_test

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/export/pdf"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func pdfOK(t *testing.T, pp *preview.PreparedPages) string {
	t.Helper()
	exp := pdf.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	return buf.String()
}

func pdfOKWith(t *testing.T, pp *preview.PreparedPages, exp *pdf.Exporter) string {
	t.Helper()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	return buf.String()
}

func pageWithObjects(objects []preview.PreparedObject) *preview.PreparedPages {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:    "Band1",
		Top:     0,
		Height:  50,
		Objects: objects,
	})
	return pp
}

// ── renderCheckBoxObject ──────────────────────────────────────────────────────

func TestPDFExporter_CheckBox_Checked(t *testing.T) {
	bl := &style.BorderLine{Color: color.RGBA{A: 255}, Width: 1}
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:    "cbOn",
			Kind:    preview.ObjectTypeCheckBox,
			Left:    10, Top: 5, Width: 20, Height: 20,
			Checked: true,
			Border:  style.Border{Lines: [4]*style.BorderLine{bl, bl, bl, bl}},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF output")
	}
}

func TestPDFExporter_CheckBox_Unchecked(t *testing.T) {
	bl := &style.BorderLine{Color: color.RGBA{A: 255}, Width: 1}
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:    "cbOff",
			Kind:    preview.ObjectTypeCheckBox,
			Left:    10, Top: 5, Width: 20, Height: 20,
			Checked: false,
			Border:  style.Border{Lines: [4]*style.BorderLine{bl, bl, bl, bl}},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF output")
	}
}

// ── renderRectObject (Line/Shape) ─────────────────────────────────────────────

func TestPDFExporter_Shape_WithFill(t *testing.T) {
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:      "shape",
			Kind:      preview.ObjectTypeShape,
			Left:      10, Top: 5, Width: 80, Height: 30,
			FillColor: color.RGBA{R: 200, G: 200, B: 255, A: 255},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF output")
	}
}

func TestPDFExporter_Line_NoBorder(t *testing.T) {
	// Line with VisibleLines=None → renderRectObject draws a fallback outline.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name: "line",
			Kind: preview.ObjectTypeLine,
			Left: 10, Top: 5, Width: 80, Height: 5,
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF output")
	}
}

func TestPDFExporter_Shape_WithBorderAllSides(t *testing.T) {
	bl := &style.BorderLine{Color: color.RGBA{A: 255}, Width: 1}
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name: "bordered",
			Kind: preview.ObjectTypeShape,
			Left: 10, Top: 5, Width: 80, Height: 30,
			Border: style.Border{
				VisibleLines: style.BorderLinesAll,
				Lines:        [4]*style.BorderLine{bl, bl, bl, bl},
			},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF output")
	}
}

func TestPDFExporter_Shape_BorderWithShadow(t *testing.T) {
	bl := &style.BorderLine{Color: color.RGBA{A: 255}, Width: 1}
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name: "shadow",
			Kind: preview.ObjectTypeShape,
			Left: 10, Top: 5, Width: 80, Height: 30,
			Border: style.Border{
				VisibleLines: style.BorderLinesAll,
				Lines:        [4]*style.BorderLine{bl, bl, bl, bl},
				Shadow:       true,
				ShadowColor:  color.RGBA{A: 200},
				ShadowWidth:  4,
			},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF output")
	}
}

func TestPDFExporter_Shape_BorderDashStyles(t *testing.T) {
	// Test dash, dot, dash-dot line styles in pdfDrawBorder.
	blDash := &style.BorderLine{Color: color.RGBA{A: 255}, Width: 1, Style: style.LineStyleDash}
	blDot := &style.BorderLine{Color: color.RGBA{A: 255}, Width: 1, Style: style.LineStyleDot}
	blDashDot := &style.BorderLine{Color: color.RGBA{A: 255}, Width: 1, Style: style.LineStyleDashDot}
	blSolid := &style.BorderLine{Color: color.RGBA{A: 255}, Width: 1, Style: style.LineStyleSolid}
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name: "dashBorder",
			Kind: preview.ObjectTypeShape,
			Left: 10, Top: 5, Width: 80, Height: 30,
			Border: style.Border{
				VisibleLines: style.BorderLinesAll,
				Lines:        [4]*style.BorderLine{blDash, blDot, blDashDot, blSolid},
			},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF output")
	}
}

// ── renderPolyPath ────────────────────────────────────────────────────────────

func TestPDFExporter_PolyLine(t *testing.T) {
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:   "polyline",
			Kind:   preview.ObjectTypePolyLine,
			Left:   10, Top: 5, Width: 80, Height: 40,
			Points: [][2]float32{{0, 0}, {40, 20}, {80, 0}},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF output")
	}
}

func TestPDFExporter_PolyLine_TooFewPoints(t *testing.T) {
	// < 2 points → early return in renderPolyPath.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:   "polyFew",
			Kind:   preview.ObjectTypePolyLine,
			Left:   10, Top: 5, Width: 80, Height: 40,
			Points: [][2]float32{{0, 0}},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF output")
	}
}

func TestPDFExporter_Polygon(t *testing.T) {
	bl := &style.BorderLine{Color: color.RGBA{R: 200, A: 255}, Width: 1}
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:   "polygon",
			Kind:   preview.ObjectTypePolygon,
			Left:   10, Top: 5, Width: 60, Height: 50,
			Points: [][2]float32{{30, 0}, {60, 50}, {0, 50}},
			Border: style.Border{Lines: [4]*style.BorderLine{bl, nil, nil, nil}},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF output")
	}
}

// ── CMYK mode → rgbaToCMYK, pdfFillColorOp CMYK, pdfStrokeColorOp CMYK ──────

func TestPDFExporter_CMYK_Text(t *testing.T) {
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:  "cmykText",
			Kind:  preview.ObjectTypeText,
			Left:  10, Top: 5, Width: 80, Height: 20,
			Text:  "CMYK color",
			Font:  style.Font{Size: 10},
			FillColor: color.RGBA{R: 100, G: 200, B: 50, A: 255},
		},
	})
	exp := pdf.NewExporter()
	exp.UseCMYK = true
	out := pdfOKWith(t, pp, exp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF output with CMYK")
	}
	// CMYK operators use 'k' and 'K' instead of 'rg' and 'RG'.
	if !strings.Contains(out, " k") && !strings.Contains(out, " K") {
		t.Error("CMYK mode should use 'k'/'K' color operators")
	}
}

func TestPDFExporter_CMYK_BlackColor(t *testing.T) {
	// Black in CMYK: r=g=b=0 → k=1, special case.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name: "cmykBlack",
			Kind: preview.ObjectTypeShape,
			Left: 10, Top: 5, Width: 80, Height: 20,
			FillColor: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		},
	})
	exp := pdf.NewExporter()
	exp.UseCMYK = true
	out := pdfOKWith(t, pp, exp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

func TestPDFExporter_CMYK_CheckBox(t *testing.T) {
	// CheckBox uses pdfStrokeColorOp → triggers CMYK stroke path.
	bl := &style.BorderLine{Color: color.RGBA{A: 255}, Width: 1}
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:    "cmykCB",
			Kind:    preview.ObjectTypeCheckBox,
			Left:    10, Top: 5, Width: 20, Height: 20,
			Checked: true,
			Border:  style.Border{Lines: [4]*style.BorderLine{bl, bl, bl, bl}},
		},
	})
	exp := pdf.NewExporter()
	exp.UseCMYK = true
	out := pdfOKWith(t, pp, exp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

// ── text object extras ────────────────────────────────────────────────────────

func TestPDFExporter_Text_WithFill(t *testing.T) {
	// FillColor.A > 0 → pdfFillRect is called.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:      "filled",
			Kind:      preview.ObjectTypeText,
			Left:      10, Top: 5, Width: 80, Height: 20,
			Text:      "Filled text",
			Font:      style.Font{Size: 10},
			FillColor: color.RGBA{R: 200, G: 200, B: 100, A: 255},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

func TestPDFExporter_Text_RTFObject(t *testing.T) {
	// ObjectTypeRTF → strips RTF then renders as text.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name: "rtf",
			Kind: preview.ObjectTypeRTF,
			Left: 10, Top: 5, Width: 80, Height: 20,
			Text: `{\rtf1\ansi Hello World}`,
			Font: style.Font{Size: 10},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

func TestPDFExporter_Text_WordWrap(t *testing.T) {
	// WordWrap=true → pdfWrapTextFn uses the word-wrap branch.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:     "wrap",
			Kind:     preview.ObjectTypeText,
			Left:     10, Top: 5, Width: 60, Height: 40,
			Text:     "This is a long text that should be wrapped across multiple lines",
			Font:     style.Font{Size: 10},
			WordWrap: true,
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

func TestPDFExporter_Text_JustifyAlign(t *testing.T) {
	// HorzAlign=3 with WordWrap → justify branch with word spacing.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:      "justify",
			Kind:      preview.ObjectTypeText,
			Left:      10, Top: 5, Width: 100, Height: 40,
			Text:      "Hello World again here for justify",
			Font:      style.Font{Size: 10},
			WordWrap:  true,
			HorzAlign: 3,
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

func TestPDFExporter_Text_CenterAndBottomAlign(t *testing.T) {
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:      "centerBottom",
			Kind:      preview.ObjectTypeText,
			Left:      10, Top: 5, Width: 80, Height: 30,
			Text:      "Centered bottom",
			Font:      style.Font{Size: 10},
			HorzAlign: 1, VertAlign: 2,
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

func TestPDFExporter_Text_RightAlign(t *testing.T) {
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:      "right",
			Kind:      preview.ObjectTypeText,
			Left:      10, Top: 5, Width: 80, Height: 20,
			Text:      "Right aligned",
			Font:      style.Font{Size: 10},
			HorzAlign: 2,
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

func TestPDFExporter_Text_EmptySkipped(t *testing.T) {
	// Empty text → ExportBand continues (skips this object).
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name: "empty",
			Kind: preview.ObjectTypeText,
			Left: 10, Top: 5, Width: 80, Height: 20,
			Text: "",
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

// ── renderBarcodeVector ───────────────────────────────────────────────────────

func TestPDFExporter_BarcodeVector(t *testing.T) {
	// PreparedObject with IsBarcode=true and BarcodeModules → renderBarcodeVector.
	modules := [][]bool{
		{true, false, true, true, false},
		{false, true, false, false, true},
		{true, true, true, false, false},
	}
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:           "bc",
			Kind:           preview.ObjectTypePicture,
			Left:           10, Top: 5, Width: 100, Height: 30,
			IsBarcode:      true,
			BarcodeModules: modules,
			TextColor:      color.RGBA{A: 255},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

func TestPDFExporter_BarcodeVector_DefaultColor(t *testing.T) {
	// TextColor.A == 0 → default to opaque black.
	modules := [][]bool{
		{true, false, true},
		{false, true, false},
	}
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:           "bcBlack",
			Kind:           preview.ObjectTypePicture,
			Left:           10, Top: 5, Width: 60, Height: 20,
			IsBarcode:      true,
			BarcodeModules: modules,
			TextColor:      color.RGBA{A: 0}, // zero alpha → default black
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

func TestPDFExporter_BarcodeVector_EmptyModules(t *testing.T) {
	// Empty modules → early return in renderBarcodeVector.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:           "bcEmpty",
			Kind:           preview.ObjectTypePicture,
			Left:           10, Top: 5, Width: 60, Height: 20,
			IsBarcode:      true,
			BarcodeModules: [][]bool{},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

// ── renderHyperlinkAnnotation ─────────────────────────────────────────────────

func TestPDFExporter_HyperlinkURL(t *testing.T) {
	// HyperlinkKind=1 (URL) → renderHyperlinkAnnotation creates URI action.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:           "urlLink",
			Kind:           preview.ObjectTypeText,
			Left:           10, Top: 5, Width: 100, Height: 20,
			Text:           "Click here",
			HyperlinkKind:  1,
			HyperlinkValue: "https://example.com",
		},
	})
	out := pdfOK(t, pp)
	if !strings.Contains(out, "URI") {
		t.Error("PDF should contain URI action for URL hyperlink")
	}
}

func TestPDFExporter_HyperlinkPageNumber(t *testing.T) {
	// HyperlinkKind=2 (PageNumber) → GoTo action.
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "B",
		Top:    0,
		Height: 30,
		Objects: []preview.PreparedObject{
			{
				Name:           "pageLink",
				Kind:           preview.ObjectTypeText,
				Left:           10, Top: 5, Width: 100, Height: 20,
				Text:           "Go to page 2",
				HyperlinkKind:  2,
				HyperlinkValue: "2",
			},
		},
	})
	out := pdfOK(t, pp)
	if !strings.Contains(out, "GoTo") {
		t.Error("PDF should contain GoTo action for page number hyperlink")
	}
}

func TestPDFExporter_HyperlinkBookmark(t *testing.T) {
	// HyperlinkKind=3 (Bookmark) → GoTo named destination.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:           "bmLink",
			Kind:           preview.ObjectTypeText,
			Left:           10, Top: 5, Width: 100, Height: 20,
			Text:           "Go to section",
			HyperlinkKind:  3,
			HyperlinkValue: "section1",
		},
	})
	out := pdfOK(t, pp)
	if !strings.Contains(out, "GoTo") {
		t.Error("PDF should contain GoTo action for bookmark hyperlink")
	}
}

// ── renderDigitalSignatureField + finalizeAcroForm ────────────────────────────

func TestPDFExporter_DigitalSignature(t *testing.T) {
	// ObjectTypeDigitalSignature → renderDigitalSignatureField + finalizeAcroForm.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name: "sig",
			Kind: preview.ObjectTypeDigitalSignature,
			Left: 10, Top: 5, Width: 150, Height: 40,
		},
	})
	out := pdfOK(t, pp)
	if !strings.Contains(out, "Widget") {
		t.Error("PDF should contain /Widget annotation for signature field")
	}
	if !strings.Contains(out, "AcroForm") {
		t.Error("PDF should contain /AcroForm for signature field")
	}
}

func TestPDFExporter_DigitalSignature_EmptyName(t *testing.T) {
	// Empty Name → uses default "Signature".
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name: "", // empty → fallback to "Signature"
			Kind: preview.ObjectTypeDigitalSignature,
			Left: 10, Top: 5, Width: 100, Height: 30,
		},
	})
	out := pdfOK(t, pp)
	if !strings.Contains(out, "Widget") {
		t.Error("PDF should contain /Widget")
	}
}

// ── writeOutlines ─────────────────────────────────────────────────────────────

func TestPDFExporter_WriteOutlines(t *testing.T) {
	// Add outline items to PreparedPages → triggers writeOutlines in Finish.
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B1", Top: 0, Height: 20})
	// Add nested outline items.
	pp.Outline.Add("Chapter 1", 0, 0)
	pp.Outline.Add("Section 1.1", 0, 50)
	pp.Outline.LevelUp()
	pp.Outline.LevelUp()
	pp.Outline.Add("Chapter 2", 1, 0)
	pp.Outline.LevelUp()

	out := pdfOK(t, pp)
	if !strings.Contains(out, "Outlines") {
		t.Error("PDF should contain /Outlines")
	}
}

// ── writeNamedDests ────────────────────────────────────────────────────────────

func TestPDFExporter_WriteNamedDests(t *testing.T) {
	// Add bookmarks → triggers writeNamedDests in Finish.
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	pp.AddBookmark("section1", 0)
	pp.AddBookmark("section2", 100)

	out := pdfOK(t, pp)
	// Named destinations dictionary should be embedded.
	if !strings.Contains(out, "Names") {
		t.Error("PDF should contain /Names for bookmarks")
	}
}

// ── watermark extras ──────────────────────────────────────────────────────────

func TestPDFExporter_WatermarkImage_Zoom(t *testing.T) {
	pngBytes := buildPNG(50, 50)
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	blobIdx := pp.BlobStore.Add("wm", pngBytes)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeZoom,
	}
	out := pdfOK(t, pp)
	if !strings.Contains(out, "WmIm") {
		t.Error("PDF should reference the watermark image XObject")
	}
}

func TestPDFExporter_WatermarkImage_Normal(t *testing.T) {
	pngBytes := buildPNG(30, 30)
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	blobIdx := pp.BlobStore.Add("wm", pngBytes)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeNormal,
	}
	out := pdfOK(t, pp)
	if !strings.Contains(out, "WmIm") {
		t.Error("PDF should reference the watermark image XObject")
	}
}

func TestPDFExporter_WatermarkImage_Tile(t *testing.T) {
	pngBytes := buildPNG(20, 20)
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	blobIdx := pp.BlobStore.Add("wm", pngBytes)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeTile,
	}
	out := pdfOK(t, pp)
	if !strings.Contains(out, "WmIm") {
		t.Error("PDF should reference the watermark image XObject")
	}
}

func TestPDFExporter_WatermarkText_Horizontal(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "HORIZONTAL",
		TextRotation: preview.WatermarkTextRotationHorizontal,
		ImageBlobIdx: -1,
	}
	out := pdfOK(t, pp)
	if !strings.Contains(out, "HORIZONTAL") {
		t.Error("PDF should contain watermark text")
	}
}

func TestPDFExporter_WatermarkText_Vertical(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "VERTICAL",
		TextRotation: preview.WatermarkTextRotationVertical,
		ImageBlobIdx: -1,
	}
	out := pdfOK(t, pp)
	if !strings.Contains(out, "VERTICAL") {
		t.Error("PDF should contain watermark text")
	}
}

func TestPDFExporter_WatermarkText_BackwardDiagonal(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "BACKWARD",
		TextRotation: preview.WatermarkTextRotationBackwardDiagonal,
		ImageBlobIdx: -1,
	}
	out := pdfOK(t, pp)
	if !strings.Contains(out, "BACKWARD") {
		t.Error("PDF should contain watermark text")
	}
}

func TestPDFExporter_WatermarkText_ZeroFontSize(t *testing.T) {
	// Font.Size == 0 → defaults to 60pt.
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "ZEROSIZE",
		Font:         style.Font{Size: 0},
		ImageBlobIdx: -1,
	}
	out := pdfOK(t, pp)
	if !strings.Contains(out, "ZEROSIZE") {
		t.Error("PDF should contain watermark text")
	}
}

// ── pdfWrapText (free function coverage via exposed path) ─────────────────────

func TestPDFExporter_Text_WordWrapDisabled(t *testing.T) {
	// WordWrap=false → pdfWrapTextFn appends paragraph as single line.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:     "noWrap",
			Kind:     preview.ObjectTypeText,
			Left:     10, Top: 5, Width: 200, Height: 20,
			Text:     "This is a long text line\nwith a second paragraph",
			Font:     style.Font{Size: 10},
			WordWrap: false,
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF")
	}
}

// ── ExportPageBegin zero dimensions ──────────────────────────────────────────

func TestPDFExporter_ZeroPageDimensions(t *testing.T) {
	// Width=0, Height=0 → A4 defaults (595×842 pt).
	pp := preview.New()
	pp.AddPage(0, 0, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF even with zero page dimensions")
	}
}

// ── ExportBand nil curPage ─────────────────────────────────────────────────────

func TestPDFExporter_ExportBand_NilCurPage(t *testing.T) {
	exp := pdf.NewExporter()
	if err := exp.ExportBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10}); err != nil {
		t.Fatalf("ExportBand with nil curPage: %v", err)
	}
}

// ── renderPictureObject JPEG ──────────────────────────────────────────────────

func TestPDFExporter_PictureObject_JPEG(t *testing.T) {
	// Build a real JPEG so the magic bytes 0xFF 0xD8 are present.
	jpegData := buildJPEG(10, 10)
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	blobIdx := pp.BlobStore.Add("pic", jpegData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "B",
		Top:    0,
		Height: 50,
		Objects: []preview.PreparedObject{
			{
				Name:    "jpgPic",
				Kind:    preview.ObjectTypePicture,
				Left:    10, Top: 5, Width: 30, Height: 30,
				BlobIdx: blobIdx,
			},
		},
	})
	out := pdfOK(t, pp)
	if !strings.Contains(out, "DCTDecode") {
		t.Error("PDF should contain DCTDecode filter for JPEG image")
	}
}

// ── watermark image Stretch / Center ─────────────────────────────────────────

func TestPDFExporter_WatermarkImage_Stretch(t *testing.T) {
	pngBytes := buildPNG(40, 40)
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	blobIdx := pp.BlobStore.Add("wm", pngBytes)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeStretch,
	}
	out := pdfOK(t, pp)
	if !strings.Contains(out, "WmIm") {
		t.Error("PDF should reference the watermark image XObject")
	}
}

func TestPDFExporter_WatermarkImage_Center(t *testing.T) {
	pngBytes := buildPNG(40, 40)
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	blobIdx := pp.BlobStore.Add("wm", pngBytes)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeCenter,
	}
	out := pdfOK(t, pp)
	if !strings.Contains(out, "WmIm") {
		t.Error("PDF should reference the watermark image XObject")
	}
}

// ── watermark text ForwardDiagonal ───────────────────────────────────────────

func TestPDFExporter_WatermarkText_ForwardDiagonal(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "FORWARD",
		TextRotation: preview.WatermarkTextRotationForwardDiagonal,
		ImageBlobIdx: -1,
	}
	out := pdfOK(t, pp)
	if !strings.Contains(out, "FORWARD") {
		t.Error("PDF should contain watermark text")
	}
}

// ── text object with mono font (Courier) ─────────────────────────────────────

func TestPDFExporter_Text_MonoFont(t *testing.T) {
	// Using a font name that maps to the "mono" family triggers the
	// mono variant of ttfDataFor and familyKeywordPDF.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:  "monoText",
			Kind:  preview.ObjectTypeText,
			Left:  10, Top: 5, Width: 200, Height: 20,
			Text:  "mono font text",
			Font:  style.Font{Name: "Courier New", Size: 10, Style: 0},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF with mono font")
	}
}

func TestPDFExporter_Text_MonoBoldItalicFont(t *testing.T) {
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:  "monoBoldItalic",
			Kind:  preview.ObjectTypeText,
			Left:  10, Top: 5, Width: 200, Height: 20,
			Text:  "bold italic mono",
			Font:  style.Font{Name: "Consolas", Size: 12, Style: style.FontStyleBold | style.FontStyleItalic},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF with mono bold italic font")
	}
}

func TestPDFExporter_Text_SansBold(t *testing.T) {
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:  "sansBold",
			Kind:  preview.ObjectTypeText,
			Left:  10, Top: 5, Width: 200, Height: 20,
			Text:  "bold sans text",
			Font:  style.Font{Name: "Arial", Size: 10, Style: style.FontStyleBold},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF with bold font")
	}
}

func TestPDFExporter_Text_SansItalic(t *testing.T) {
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:  "sansItalic",
			Kind:  preview.ObjectTypeText,
			Left:  10, Top: 5, Width: 200, Height: 20,
			Text:  "italic sans text",
			Font:  style.Font{Name: "Arial", Size: 10, Style: style.FontStyleItalic},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF with italic font")
	}
}

func TestPDFExporter_Text_MonoBold(t *testing.T) {
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:  "monoBold",
			Kind:  preview.ObjectTypeText,
			Left:  10, Top: 5, Width: 200, Height: 20,
			Text:  "bold mono text",
			Font:  style.Font{Name: "Courier", Size: 10, Style: style.FontStyleBold},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF with mono bold font")
	}
}

func TestPDFExporter_Text_MonoItalic(t *testing.T) {
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:  "monoItalic",
			Kind:  preview.ObjectTypeText,
			Left:  10, Top: 5, Width: 200, Height: 20,
			Text:  "italic mono text",
			Font:  style.Font{Name: "Monaco", Size: 10, Style: style.FontStyleItalic},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF with mono italic font")
	}
}

// ── renderTextObject edge cases ───────────────────────────────────────────────

func TestPDFExporter_Text_VertAlignCenter(t *testing.T) {
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:      "centerAlign",
			Kind:      preview.ObjectTypeText,
			Left:      10, Top: 5, Width: 200, Height: 40,
			Text:      "centered text",
			Font:      style.Font{Name: "Arial", Size: 10},
			VertAlign: 1, // Center
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF with center-aligned text")
	}
}

func TestPDFExporter_Text_ZeroFontSize(t *testing.T) {
	// Font.Size == 0 → defaults to style.DefaultFont().Size inside renderTextObject.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:  "zeroFont",
			Kind:  preview.ObjectTypeText,
			Left:  10, Top: 5, Width: 200, Height: 30,
			Text:  "zero size font",
			Font:  style.Font{Name: "Arial", Size: 0}, // triggers size defaulting
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF with zero font size defaulted")
	}
}

func TestPDFExporter_Text_NarrowBox(t *testing.T) {
	// Width so narrow that innerW <= 0 → triggers the innerW=wPt fallback.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:  "narrowBox",
			Kind:  preview.ObjectTypeText,
			Left:  10, Top: 5, Width: 3, Height: 30, // 3px wide → innerW = 3 - 4 ≤ 0
			Text:  "narrow",
			Font:  style.Font{Name: "Arial", Size: 10},
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF with narrow text box")
	}
}

func TestPDFExporter_Text_ManyLinesOverflow(t *testing.T) {
	// Many lines in a small box: some lines go below the object — break loop.
	pp := pageWithObjects([]preview.PreparedObject{
		{
			Name:     "overflow",
			Kind:     preview.ObjectTypeText,
			Left:     10, Top: 5, Width: 100, Height: 15, // only fits ~1 line
			Text:     "line1\nline2\nline3\nline4\nline5",
			Font:     style.Font{Name: "Arial", Size: 10},
			WordWrap: false,
		},
	})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF with overflow text")
	}
}

// ── watermark image JPEG ──────────────────────────────────────────────────────

func TestPDFExporter_WatermarkImage_JPEG(t *testing.T) {
	// JPEG watermark triggers the DCTDecode path in renderWatermarkImage.
	jpegData := buildJPEG(40, 40)
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	blobIdx := pp.BlobStore.Add("wm", jpegData)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeStretch,
	}
	out := pdfOK(t, pp)
	if !strings.Contains(out, "WmIm") {
		t.Error("PDF should reference the watermark image XObject")
	}
}

func TestPDFExporter_WatermarkImage_HighTransparency(t *testing.T) {
	// ImageTransparency > 1.0 → opacity < 0 → clamped to 0.
	pngBytes := buildPNG(20, 20)
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	blobIdx := pp.BlobStore.Add("wm", pngBytes)
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:           true,
		ImageBlobIdx:      blobIdx,
		ImageSize:         preview.WatermarkImageSizeStretch,
		ImageTransparency: 1.5, // > 1.0 → opacity = 1-1.5 = -0.5 → clamped to 0
	}
	out := pdfOK(t, pp)
	if !strings.Contains(out, "WmIm") {
		t.Error("PDF should reference the watermark image XObject")
	}
}

// ── writeOutlines with Prev/Next sibling linking ──────────────────────────────

func TestPDFExporter_WriteOutlines_MultipleSiblings(t *testing.T) {
	// Three root-level outline items trigger the top-level Prev/Next sibling linking.
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)
	pp.AddPage(595, 842, 3)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	pp.Outline.Add("Chapter 1", 0, 0)
	pp.Outline.LevelUp()
	pp.Outline.Add("Chapter 2", 1, 0)
	pp.Outline.LevelUp()
	pp.Outline.Add("Chapter 3", 2, 0)
	pp.Outline.LevelUp()
	out := pdfOK(t, pp)
	if !strings.Contains(out, "Outlines") {
		t.Error("PDF should contain /Outlines for multiple siblings")
	}
}

func TestPDFExporter_WriteOutlines_ChildSiblings(t *testing.T) {
	// A parent item with 2 children triggers the child-level Prev/Next sibling
	// linking code inside buildItem.
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	// Root → Parent → Child1, Child2 (two siblings at child level)
	pp.Outline.Add("Parent", 0, 0)   // cur = Parent
	pp.Outline.Add("Child1", 0, 10)  // cur = Child1
	pp.Outline.LevelUp()              // cur = Parent
	pp.Outline.Add("Child2", 0, 20)  // cur = Child2
	pp.Outline.LevelUp()              // cur = Parent
	pp.Outline.LevelUp()              // cur = Root
	out := pdfOK(t, pp)
	if !strings.Contains(out, "Outlines") {
		t.Error("PDF should contain /Outlines with child siblings")
	}
}

// ── writeNamedDests: page index out of range ──────────────────────────────────

func TestPDFExporter_WriteNamedDests_OutOfRange(t *testing.T) {
	// Bookmark with pageIdx beyond the page count — should skip via continue.
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	// Add a valid bookmark first so writeNamedDests is triggered at all.
	pp.AddBookmark("valid", 0)
	// Add a bookmark with an out-of-range page index directly.
	pp.Bookmarks.Add(&preview.Bookmark{Name: "far", PageIdx: 999, OffsetY: 0})
	out := pdfOK(t, pp)
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("expected valid PDF even with out-of-range bookmark page")
	}
}
