// html_coverage4_test.go — external tests that improve coverage for:
//   - renderObject: barcode picture path (IsBarcode + background-size)
//   - renderObject: picture with opaque FillColor
//   - renderTextObject: RTL text direction, strikeout decoration, clip overflow,
//     paragraph offset, explicit line height, padding fields, strikeout+underline
//   - renderObjectLayered: text object (z-index injection without position:absolute prefix)
//   - ExportPageEnd: unlimited height (maxBottom > pageH), multi-page CSS emission,
//     non-first page break div, EmbedCSS disabled
//   - checkbox: unchecked symbols (minus, slash, backslash), checked plus/fill

package html_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// buildPageWithObjs creates a PreparedPages with a single page containing the given objects.
func buildPageWithObjs(objects []preview.PreparedObject) *preview.PreparedPages {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:    "Band",
		Top:     0,
		Height:  200,
		Objects: objects,
	})
	return pp
}

// exportHTMLCustom runs Export with a custom exporter and returns the HTML string.
func exportHTMLCustom(t *testing.T, exp *html.Exporter, pp *preview.PreparedPages) string {
	t.Helper()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	return buf.String()
}

// makePNG creates a small PNG image blob.
func makePNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// ── Barcode picture rendering (IsBarcode path) ────────────────────────────────

func TestRenderObject_Picture_Barcode(t *testing.T) {
	// Create a barcode picture object with blob data.
	pngData := makePNG(30, 30)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("barcode.png", pngData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 200,
		Objects: []preview.PreparedObject{
			{
				Name:    "Barcode1",
				Kind:    preview.ObjectTypePicture,
				Left:    10,
				Top:     10,
				Width:   100,
				Height:  50,
				BlobIdx: blobIdx,
				// Mark as barcode — C# does not use background-size for barcode images.
				IsBarcode: true,
			},
		},
	})
	exp := html.NewExporter()
	out := exportHTMLCustom(t, exp, pp)
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("Barcode: expected base64 PNG image in output, not found")
	}
}

func TestRenderObject_Picture_WithFillColor(t *testing.T) {
	// Picture with opaque FillColor → should use background-color:rgb(...) not transparent.
	pngData := makePNG(20, 20)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("pic.png", pngData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 200,
		Objects: []preview.PreparedObject{
			{
				Name:      "Pic1",
				Kind:      preview.ObjectTypePicture,
				Left:      10,
				Top:       10,
				Width:     100,
				Height:    50,
				BlobIdx:   blobIdx,
				FillColor: color.RGBA{R: 100, G: 200, B: 50, A: 255},
			},
		},
	})
	exp := html.NewExporter()
	out := exportHTMLCustom(t, exp, pp)
	if !strings.Contains(out, "rgb(100, 200, 50)") {
		t.Errorf("Picture FillColor: expected rgb(100, 200, 50), not found")
	}
}

// ── RTL text ──────────────────────────────────────────────────────────────────

func TestRenderObject_Text_RTL_DefaultAlign(t *testing.T) {
	// RTL with default alignment (HorzAlign=0) → text-align:right
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "RTL1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 30,
			Text: "Hebrew text",
			Font: style.Font{Name: "Arial", Size: 10},
			RTL:  true,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "text-align:right") {
		t.Errorf("RTL default: expected text-align:right, not found")
	}
	if !strings.Contains(out, "direction:rtl") {
		t.Errorf("RTL default: expected direction:rtl, not found")
	}
}

func TestRenderObject_Text_RTL_RightAlign(t *testing.T) {
	// RTL with HorzAlign=2 (Right) → in RTL mode, Right becomes text-align:left.
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "RTL2", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 30,
			Text:      "Hebrew right",
			Font:      style.Font{Name: "Arial", Size: 10},
			RTL:       true,
			HorzAlign: 2,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "text-align:left") {
		t.Errorf("RTL right-align: expected text-align:left, not found")
	}
}

// ── Strikeout text decoration ─────────────────────────────────────────────────

func TestRenderObject_Text_FontStrikeout(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "Strike1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 30,
			Text: "Struck out",
			Font: style.Font{Name: "Arial", Size: 10, Style: style.FontStyleStrikeout},
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "text-decoration:line-through") {
		t.Errorf("Strikeout: expected text-decoration:line-through, not found")
	}
}

func TestRenderObject_Text_FontUnderlineAndStrikeout(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "Both1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 30,
			Text: "Both decorations",
			Font: style.Font{Name: "Arial", Size: 10, Style: style.FontStyleUnderline | style.FontStyleStrikeout},
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "text-decoration:underline|line-through") {
		t.Errorf("Underline+Strikeout: expected text-decoration:underline|line-through, not found")
	}
}

// ── Clip (overflow:hidden) ────────────────────────────────────────────────────

func TestRenderObject_Text_Clip(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "Clip1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 30,
			Text: "Clipped text",
			Font: style.Font{Name: "Arial", Size: 10},
			Clip: true,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "overflow:hidden") {
		t.Errorf("Clip: expected overflow:hidden, not found")
	}
}

// ── ParagraphOffset / text-indent ─────────────────────────────────────────────

func TestRenderObject_Text_ParagraphOffset(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "Para1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 30,
			Text:            "Indented text",
			Font:            style.Font{Name: "Arial", Size: 10},
			ParagraphOffset: 20,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "text-indent:") {
		t.Errorf("ParagraphOffset: expected text-indent, not found")
	}
}

// ── Explicit line height ──────────────────────────────────────────────────────

func TestRenderObject_Text_ExplicitLineHeight(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "LH1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 30,
			Text:       "With line height",
			Font:       style.Font{Name: "Arial", Size: 10},
			LineHeight: 24,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "line-height: 24px") {
		t.Errorf("LineHeight: expected 'line-height: 24px', not found")
	}
}

// ── Padding fields ────────────────────────────────────────────────────────────

func TestRenderObject_Text_Padding(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "Pad1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 30,
			Text:          "Padded text",
			Font:          style.Font{Name: "Arial", Size: 10},
			PaddingLeft:   5,
			PaddingRight:  3,
			PaddingTop:    7,
			PaddingBottom: 2,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "padding-left:5px") {
		t.Errorf("PaddingLeft: expected padding-left:5px, not found")
	}
	if !strings.Contains(out, "padding-right:3px") {
		t.Errorf("PaddingRight: expected padding-right:3px, not found")
	}
	if !strings.Contains(out, "padding-top:7px") {
		t.Errorf("PaddingTop: expected padding-top:7px, not found")
	}
}

// ── Layered text objects (z-index without position:absolute prefix) ───────────

func TestRenderObject_Text_Layered_ZIndex(t *testing.T) {
	// In layers mode, text objects should get z-index injected via the style="" path
	// (since text uses renderTextObject which includes position:absolute in the CSS class,
	// the z-index should be injected).
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "LayerText", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 30,
			Text: "Layered text",
			Font: style.Font{Name: "Arial", Size: 10},
		},
	})
	exp := html.NewExporter()
	exp.Layers = true
	out := exportHTMLCustom(t, exp, pp)
	if !strings.Contains(out, "z-index:") {
		t.Errorf("Layered text: expected z-index in output, not found")
	}
}

// ── UnlimitedHeight (maxBottom > pageH) ───────────────────────────────────────

func TestExportPageEnd_UnlimitedHeight(t *testing.T) {
	// Create a page with a band that extends beyond the declared page height.
	// The page is 200px high but the band bottom is at 500px.
	pp := preview.New()
	pp.AddPage(794, 200, 1) // short page
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "TallBand",
		Top:    0,
		Height: 500, // extends beyond page height 200
		Width:  794,
	})
	exp := html.NewExporter()
	out := exportHTMLCustom(t, exp, pp)
	// The page div should have its height expanded from 200 to 500.
	if !strings.Contains(out, "height:500px") {
		t.Errorf("UnlimitedHeight: expected height:500px in output, not found")
	}
}

// ── Multi-page CSS emission ───────────────────────────────────────────────────

func TestExport_MultiPage_PageBreakDiv(t *testing.T) {
	// Two pages → second page should have a page break div.
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B1", Top: 0, Height: 30, Width: 794})
	pp.AddPage(794, 1123, 2)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B2", Top: 0, Height: 30, Width: 794})
	exp := html.NewExporter()
	out := exportHTMLCustom(t, exp, pp)
	// Both page anchors should be present.
	if !strings.Contains(out, `name="PageN1"`) {
		t.Errorf("MultiPage: expected PageN1 anchor")
	}
	if !strings.Contains(out, `name="PageN2"`) {
		t.Errorf("MultiPage: expected PageN2 anchor")
	}
	// Page break div should be present before page 2.
	if !strings.Contains(out, "break-after:page") {
		t.Errorf("MultiPage: expected break-after:page div for page 2")
	}
	// Per-page print CSS should appear twice.
	count := strings.Count(out, `media="print"`)
	if count < 2 {
		t.Errorf("MultiPage: expected at least 2 print CSS blocks, got %d", count)
	}
}

func TestExport_EmbedCSS_Disabled(t *testing.T) {
	// When EmbedCSS is false, no <style> blocks should appear.
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B1", Top: 0, Height: 30, Width: 794})
	exp := html.NewExporter()
	exp.EmbedCSS = false
	out := exportHTMLCustom(t, exp, pp)
	if strings.Contains(out, `<style type="text/css"`) {
		t.Errorf("EmbedCSS disabled: expected no <style> blocks, but found one")
	}
}

// ── Checkbox: unchecked symbols ───────────────────────────────────────────────

// Checkboxes are now rendered as PNG images (matching C# LayerBack + LayerPicture).
// These tests verify that the PNG rendering produces valid base64 image output.

func TestRenderObject_CheckBox_UncheckedMinus(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{Name: "CB1", Kind: preview.ObjectTypeCheckBox, Left: 0, Top: 0, Width: 20, Height: 20, Checked: false, UncheckedSymbol: 2},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "base64,") {
		t.Errorf("Unchecked minus: expected base64 PNG image")
	}
}

func TestRenderObject_CheckBox_UncheckedSlash(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{Name: "CB2", Kind: preview.ObjectTypeCheckBox, Left: 0, Top: 0, Width: 20, Height: 20, Checked: false, UncheckedSymbol: 3},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "base64,") {
		t.Errorf("Unchecked slash: expected base64 PNG image")
	}
}

func TestRenderObject_CheckBox_UncheckedBackslash(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{Name: "CB3", Kind: preview.ObjectTypeCheckBox, Left: 0, Top: 0, Width: 20, Height: 20, Checked: false, UncheckedSymbol: 4},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "base64,") {
		t.Errorf("Unchecked backslash: expected base64 PNG image")
	}
}

func TestRenderObject_CheckBox_CheckedPlus(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{Name: "CB4", Kind: preview.ObjectTypeCheckBox, Left: 0, Top: 0, Width: 20, Height: 20, Checked: true, CheckedSymbol: 2},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "base64,") {
		t.Errorf("Checked plus: expected base64 PNG image")
	}
}

func TestRenderObject_CheckBox_CheckedFill(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{Name: "CB5", Kind: preview.ObjectTypeCheckBox, Left: 0, Top: 0, Width: 20, Height: 20, Checked: true, CheckedSymbol: 3},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "base64,") {
		t.Errorf("Checked fill: expected base64 PNG image")
	}
}

func TestRenderObject_CheckBox_CheckedCross(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{Name: "CB6", Kind: preview.ObjectTypeCheckBox, Left: 0, Top: 0, Width: 20, Height: 20, Checked: true, CheckedSymbol: 1},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "base64,") {
		t.Errorf("Checked cross: expected base64 PNG image")
	}
}

// ── VertAlign center/bottom ───────────────────────────────────────────────────

func TestRenderObject_Text_VertAlignCenter_WithPadding(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "VC1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 60,
			Text:          "Centered",
			Font:          style.Font{Name: "Arial", Size: 10},
			VertAlign:     1, // center
			PaddingTop:    2,
			PaddingBottom: 2,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "margin-top:") {
		t.Errorf("VertAlign center: expected margin-top in output")
	}
}

func TestRenderObject_Text_VertAlignBottom_WithPadding(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "VB1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 60,
			Text:          "Bottom aligned",
			Font:          style.Font{Name: "Arial", Size: 10},
			VertAlign:     2, // bottom
			PaddingBottom: 2,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "margin-top:") {
		t.Errorf("VertAlign bottom: expected margin-top in output")
	}
}

// ── Text with border ──────────────────────────────────────────────────────────

func TestRenderObject_Text_WithBorder(t *testing.T) {
	// Text object with visible border → should affect border CSS and border width adjustments.
	bl := &style.BorderLine{Width: 2, Style: style.LineStyleSolid, Color: color.RGBA{A: 255}}
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "TB1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 30,
			Text: "Bordered text",
			Font: style.Font{Name: "Arial", Size: 10},
			Border: style.Border{
				VisibleLines: style.BorderLinesAll,
				Lines:        [4]*style.BorderLine{bl, bl, bl, bl},
			},
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "border-top-width:") {
		t.Errorf("Text border: expected border-top-width, not found")
	}
}

// ── Text fill color transparent vs opaque ─────────────────────────────────────

func TestRenderObject_Text_TransparentFillColor(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "TF1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 30,
			Text:      "Transparent bg",
			Font:      style.Font{Name: "Arial", Size: 10},
			FillColor: color.RGBA{A: 0},
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "background-color:transparent") {
		t.Errorf("Text transparent fill: expected background-color:transparent, not found")
	}
}

func TestRenderObject_Text_OpaqueFillColor(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "TF2", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 30,
			Text:      "Red bg",
			Font:      style.Font{Name: "Arial", Size: 10},
			FillColor: color.RGBA{R: 255, A: 255},
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "background-color:rgb(255, 0, 0)") {
		t.Errorf("Text opaque fill: expected background-color:rgb(255, 0, 0), not found")
	}
}

// ── Layered non-text object (z-index injection via style= fallback) ───────────

func TestRenderObject_Shape_Layered_ZIndex(t *testing.T) {
	// Non-text objects in layers mode don't have position:absolute in their rendered
	// output, so z-index should be injected via the style=" fallback path.
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "LayerShape", Kind: preview.ObjectTypeShape,
			Left: 10, Top: 10, Width: 50, Height: 50,
			ShapeKind: 0, // rectangle
		},
	})
	exp := html.NewExporter()
	exp.Layers = true
	out := exportHTMLCustom(t, exp, pp)
	if !strings.Contains(out, "z-index:") {
		t.Errorf("Layered shape: expected z-index in output, not found")
	}
}

func TestRenderObject_Line_Layered_ZIndex(t *testing.T) {
	// A non-diagonal line in layers mode: the rendered div has no position:absolute;
	// inline, so z-index should be injected via the style=" fallback path.
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "LayerLine", Kind: preview.ObjectTypeLine,
			Left: 10, Top: 10, Width: 100, Height: 1,
			LineDiagonal: false,
		},
	})
	exp := html.NewExporter()
	exp.Layers = true
	out := exportHTMLCustom(t, exp, pp)
	if !strings.Contains(out, "z-index:") {
		t.Errorf("Layered line: expected z-index in output, not found")
	}
}

// ── renderObject: non-text with FillColor (in the generic path) ───────────────

func TestRenderObject_Shape_WithFillColor(t *testing.T) {
	pp := buildPageWithObjs([]preview.PreparedObject{
		{
			Name: "S1", Kind: preview.ObjectTypeShape,
			Left: 0, Top: 0, Width: 50, Height: 50,
			FillColor: color.RGBA{R: 0, G: 128, B: 255, A: 255},
			ShapeKind: 0, // rectangle
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("Shape FillColor: expected data:image/Png;base64, not found")
	}
	if !strings.Contains(out, "position:absolute;") {
		t.Errorf("Shape FillColor: expected position:absolute in inline style, not found")
	}
}
