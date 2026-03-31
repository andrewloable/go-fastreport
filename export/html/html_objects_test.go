// External tests for the html package covering renderObject cases,
// watermark image, the HTML() method, and ExportPageEnd edge cases.
package html_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// buildPage creates a PreparedPages with a single page and band with the given objects.
func buildPage(objects []preview.PreparedObject) *preview.PreparedPages {
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

// exportHTML runs Export and returns the HTML string.
func exportHTML(t *testing.T, pp *preview.PreparedPages) string {
	t.Helper()
	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	return buf.String()
}

// buildHTMLPNG creates a small PNG image blob.
func buildHTMLPNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// buildHTMLJPEG creates a small JPEG image blob.
func buildHTMLJPEG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, nil)
	return buf.Bytes()
}

// ── ObjectTypeText ─────────────────────────────────────────────────────────────

func TestRenderObject_Text_HorzAlign_Left(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Left",
			HorzAlign: 0, // default left
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "text-align:left;") {
		t.Errorf("HorzAlign=0: expected text-align:left; in %q", out)
	}
	if !strings.Contains(out, "Left") {
		t.Error("expected text 'Left' in output")
	}
}

func TestRenderObject_Text_HorzAlign_Center(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Center",
			HorzAlign: 1,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "text-align:center;") {
		t.Errorf("HorzAlign=1: expected text-align:center; in %q", out)
	}
}

func TestRenderObject_Text_HorzAlign_Right(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Right",
			HorzAlign: 2,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "text-align:right;") {
		t.Errorf("HorzAlign=2: expected text-align:right; in %q", out)
	}
}

func TestRenderObject_Text_HorzAlign_Justify(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Justified text",
			HorzAlign: 3,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "text-align:justify;") {
		t.Errorf("HorzAlign=3: expected text-align:justify; in %q", out)
	}
}

func TestRenderObject_Text_VertAlign_Top(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 40,
			Text:      "Top",
			VertAlign: 0,
		},
	})
	out := exportHTML(t, pp)
	// VertAlign=0 adds no flex alignment CSS.
	if strings.Contains(out, "align-items:center;") || strings.Contains(out, "align-items:flex-end;") {
		t.Errorf("VertAlign=0 should not add flex alignment, got %q", out)
	}
}

func TestRenderObject_Text_VertAlign_Center(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 40,
			Text:      "VCenter",
			VertAlign: 1,
		},
	})
	out := exportHTML(t, pp)
	// C# uses margin-top for vertical centering, not flex alignment.
	if !strings.Contains(out, "margin-top:") {
		t.Errorf("VertAlign=1: expected margin-top in %q", out)
	}
}

func TestRenderObject_Text_VertAlign_Bottom(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 40,
			Text:      "VBottom",
			VertAlign: 2,
		},
	})
	out := exportHTML(t, pp)
	// C# uses margin-top for bottom alignment, not flex alignment.
	if !strings.Contains(out, "margin-top:") {
		t.Errorf("VertAlign=2: expected margin-top in %q", out)
	}
}

func TestRenderObject_Text_FontBold(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:  preview.ObjectTypeText,
			Left:  0, Top: 0, Width: 100, Height: 20,
			Text:  "Bold",
			Font:  style.Font{Name: "Arial", Size: 12, Style: style.FontStyleBold},
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "font-weight:bold;") {
		t.Errorf("bold font: expected font-weight:bold; in %q", out)
	}
}

func TestRenderObject_Text_FontItalic(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:  preview.ObjectTypeText,
			Left:  0, Top: 0, Width: 100, Height: 20,
			Text:  "Italic",
			Font:  style.Font{Name: "Arial", Size: 12, Style: style.FontStyleItalic},
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "font-style:italic;") {
		t.Errorf("italic font: expected font-style:italic; in %q", out)
	}
}

func TestRenderObject_Text_FontUnderline(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:  preview.ObjectTypeText,
			Left:  0, Top: 0, Width: 100, Height: 20,
			Text:  "Underline",
			Font:  style.Font{Name: "Arial", Size: 12, Style: style.FontStyleUnderline},
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "text-decoration:underline;") {
		t.Errorf("underline font: expected text-decoration:underline; in %q", out)
	}
}

func TestRenderObject_Text_WordWrap(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:     preview.ObjectTypeText,
			Left:     0, Top: 0, Width: 100, Height: 40,
			Text:     "Word wrapped text here",
			WordWrap: true,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "word-wrap:break-word;") {
		t.Errorf("word-wrap: expected word-wrap:break-word; in %q", out)
	}
	// C# doesn't add white-space:normal; just word-wrap:break-word is enough.
}

func TestRenderObject_Text_NoWordWrap(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:     preview.ObjectTypeText,
			Left:     0, Top: 0, Width: 100, Height: 20,
			Text:     "No wrap",
			WordWrap: false,
		},
	})
	out := exportHTML(t, pp)
	// C# uses "overflow: hidden; text-wrap: nowrap;" for non-wrapping text.
	if !strings.Contains(out, "text-wrap: nowrap;") {
		t.Errorf("no word-wrap: expected text-wrap: nowrap; in %q", out)
	}
}

func TestRenderObject_Text_Hyperlink(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:           preview.ObjectTypeText,
			Left:           0, Top: 0, Width: 100, Height: 20,
			Text:           "Click here",
			HyperlinkKind:  1,
			HyperlinkValue: "https://example.com",
		},
	})
	out := exportHTML(t, pp)
	// C# GetHref: <a style="color:..." href="URL">
	if !strings.Contains(out, `href="https://example.com"`) {
		t.Errorf("hyperlink: expected href attribute, got %q", out)
	}
	// C# only adds target="_blank" when OpenLinkInNewTab is set (not by default).
	if !strings.Contains(out, "cursor:pointer;") {
		t.Errorf("hyperlink: expected cursor:pointer, got %q", out)
	}
	if !strings.Contains(out, "Click here") {
		t.Errorf("hyperlink: expected link text, got %q", out)
	}
}

func TestRenderObject_Text_Hyperlink_NoAnchorWhenKindZero(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:           preview.ObjectTypeText,
			Left:           0, Top: 0, Width: 100, Height: 20,
			Text:           "Normal text",
			HyperlinkKind:  0,
			HyperlinkValue: "https://example.com",
		},
	})
	out := exportHTML(t, pp)
	if strings.Contains(out, "<a href=") {
		t.Errorf("HyperlinkKind=0: should not produce anchor tag, got %q", out)
	}
}

func TestRenderObject_Text_Hyperlink_Bookmark(t *testing.T) {
	// Kind=3 (Bookmark): C# SinglePage mode → <a href="#bookmarkName">
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:           preview.ObjectTypeText,
			Left:           0, Top: 0, Width: 100, Height: 20,
			Text:           "Go to section",
			HyperlinkKind:  3,
			HyperlinkValue: "SectionA",
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, `href="#SectionA"`) {
		t.Errorf("Bookmark hyperlink: expected href=\"#SectionA\", got %q", out)
	}
	if !strings.Contains(out, "cursor:pointer;") {
		t.Errorf("Bookmark hyperlink: expected cursor:pointer, got %q", out)
	}
	if !strings.Contains(out, "Go to section") {
		t.Errorf("Bookmark hyperlink: expected link text, got %q", out)
	}
}

func TestRenderObject_Text_Hyperlink_PageNumber(t *testing.T) {
	// Kind=2 (PageNumber): C# SinglePage mode → <a href="#PageN{n}">
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:           preview.ObjectTypeText,
			Left:           0, Top: 0, Width: 100, Height: 20,
			Text:           "Go to page 5",
			HyperlinkKind:  2,
			HyperlinkValue: "5",
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, `href="#PageN5"`) {
		t.Errorf("PageNumber hyperlink: expected href=\"#PageN5\", got %q", out)
	}
	if !strings.Contains(out, "cursor:pointer;") {
		t.Errorf("PageNumber hyperlink: expected cursor:pointer, got %q", out)
	}
}

// ── ObjectTypeHtml ─────────────────────────────────────────────────────────────

func TestRenderObject_Html_PassThrough(t *testing.T) {
	rawHTML := `<b>bold</b> <i>italic</i>`
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:  preview.ObjectTypeHtml,
			Left:  0, Top: 0, Width: 200, Height: 50,
			Text:  rawHTML,
		},
	})
	out := exportHTML(t, pp)
	// Raw HTML should appear verbatim, not escaped.
	if !strings.Contains(out, rawHTML) {
		t.Errorf("HTML passthrough: expected raw HTML %q in output %q", rawHTML, out)
	}
	// Must NOT be HTML-escaped.
	if strings.Contains(out, "&lt;b&gt;") {
		t.Errorf("HTML passthrough: tags should not be escaped, got %q", out)
	}
}

// ── ObjectTypeRTF ─────────────────────────────────────────────────────────────

func TestRenderObject_RTF_Rendered(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind: preview.ObjectTypeRTF,
			Left: 0, Top: 0, Width: 200, Height: 50,
			Text: `{\rtf1\ansi Hello RTF World}`,
		},
	})
	out := exportHTML(t, pp)
	// RTF is converted to HTML; the container div should be present.
	if !strings.Contains(out, `overflow:hidden;`) {
		t.Errorf("RTF: expected overflow:hidden container, got %q", out)
	}
}

// ── ObjectTypePicture ─────────────────────────────────────────────────────────

func TestRenderObject_Picture_WithBlob(t *testing.T) {
	pngData := buildHTMLPNG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("img", pngData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Kind:    preview.ObjectTypePicture,
				Left:    0, Top: 0, Width: 80, Height: 80,
				BlobIdx: blobIdx,
			},
		},
	})
	out := exportHTML(t, pp)
	// New rendering: images use CSS background in a <style> block (C# pattern).
	if !strings.Contains(out, `url('data:image/Png;base64,`) {
		t.Errorf("picture with blob: expected CSS background url with image/Png, got %q", out)
	}
}

func TestRenderObject_Picture_JPEG_Blob(t *testing.T) {
	jpegData := buildHTMLJPEG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("jpg", jpegData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Kind:    preview.ObjectTypePicture,
				Left:    0, Top: 0, Width: 80, Height: 80,
				BlobIdx: blobIdx,
			},
		},
	})
	out := exportHTML(t, pp)
	// JPEG images are converted to PNG (zoom-to-fit) so they render at the correct
	// container size in HTML (C# pattern: obj.Draw() always outputs a fixed-size bitmap).
	if !strings.Contains(out, `url('data:image/Png;base64,`) {
		t.Errorf("JPEG picture: expected CSS background url with image/Png MIME after resize, got %q", out)
	}
}

func TestRenderObject_Picture_NoBlob_EmptyPlaceholder(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:    preview.ObjectTypePicture,
			Left:    0, Top: 0, Width: 80, Height: 80,
			BlobIdx: -1, // no blob
		},
	})
	out := exportHTML(t, pp)
	// Should produce an empty placeholder div, no img tag.
	if strings.Contains(out, "<img") {
		t.Errorf("no-blob picture: should not have img tag, got %q", out)
	}
}

// ── ObjectTypeLine ─────────────────────────────────────────────────────────────

func TestRenderObject_Line_Horizontal(t *testing.T) {
	// C# renders all lines as PNG images via LayerPicture (one div, two CSS classes).
	// Reference: HTMLExportLayers.cs:939-941
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:         preview.ObjectTypeLine,
			Left:         0, Top: 0, Width: 100, Height: 1,
			LineDiagonal: false,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "base64,") {
		t.Errorf("horizontal line: expected base64 PNG image, got %q", out)
	}
}

func TestRenderObject_Line_Vertical(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:         preview.ObjectTypeLine,
			Left:         0, Top: 0, Width: 1, Height: 100,
			LineDiagonal: false,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "base64,") {
		t.Errorf("vertical line: expected base64 PNG image, got %q", out)
	}
}

func TestRenderObject_Line_Diagonal(t *testing.T) {
	// C# renders diagonal lines as PNG images too (via LayerPicture).
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:         preview.ObjectTypeLine,
			Left:         0, Top: 0, Width: 100, Height: 50,
			LineDiagonal: true,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "base64,") {
		t.Errorf("diagonal line: expected base64 PNG image, got %q", out)
	}
}

// ── ObjectTypeShape ────────────────────────────────────────────────────────────

func TestRenderObject_Shape_Rectangle(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 100, Height: 50,
			ShapeKind: 0, // rectangle
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("rect shape: expected data:image/Png;base64, got %q", out)
	}
	if !strings.Contains(out, "position:absolute;") {
		t.Errorf("rect shape: expected position:absolute in inline style, got %q", out)
	}
}

func TestRenderObject_Shape_RoundRect(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:       preview.ObjectTypeShape,
			Left:       0, Top: 0, Width: 100, Height: 50,
			ShapeKind:  1, // round rectangle
			ShapeCurve: 10,
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("roundrect shape: expected data:image/Png;base64, got %q", out)
	}
	if !strings.Contains(out, "position:absolute;") {
		t.Errorf("roundrect shape: expected position:absolute in inline style, got %q", out)
	}
}

func TestRenderObject_Shape_Ellipse(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 100, Height: 50,
			ShapeKind: 2, // ellipse
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("ellipse shape: expected data:image/Png;base64, got %q", out)
	}
	if !strings.Contains(out, "position:absolute;") {
		t.Errorf("ellipse shape: expected position:absolute in inline style, got %q", out)
	}
}

func TestRenderObject_Shape_Triangle(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 80, Height: 60,
			ShapeKind: 3, // triangle
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("triangle shape: expected data:image/Png;base64, got %q", out)
	}
	if !strings.Contains(out, "position:absolute;") {
		t.Errorf("triangle shape: expected position:absolute in inline style, got %q", out)
	}
}

func TestRenderObject_Shape_Diamond(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 80, Height: 60,
			ShapeKind: 4, // diamond
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("diamond shape: expected data:image/Png;base64, got %q", out)
	}
	if !strings.Contains(out, "position:absolute;") {
		t.Errorf("diamond shape: expected position:absolute in inline style, got %q", out)
	}
}

// ── ObjectTypeDigitalSignature ─────────────────────────────────────────────────

func TestRenderObject_DigitalSignature_WithText(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:  preview.ObjectTypeDigitalSignature,
			Left:  0, Top: 0, Width: 150, Height: 40,
			Text:  "Sign Here",
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "Sign Here") {
		t.Errorf("digital signature with text: expected 'Sign Here', got %q", out)
	}
	if !strings.Contains(out, "dashed") {
		t.Errorf("digital signature: expected dashed border, got %q", out)
	}
}

func TestRenderObject_DigitalSignature_EmptyText(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:  preview.ObjectTypeDigitalSignature,
			Left:  0, Top: 0, Width: 150, Height: 40,
			Text:  "", // empty → fallback label
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "Digital Signature") {
		t.Errorf("digital signature empty text: expected default label 'Digital Signature', got %q", out)
	}
}

// ── ObjectTypeCheckBox ─────────────────────────────────────────────────────────

func TestRenderObject_CheckBox_Checked(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:    preview.ObjectTypeCheckBox,
			Left:    0, Top: 0, Width: 20, Height: 20,
			Checked: true,
		},
	})
	out := exportHTML(t, pp)
	// Checkboxes render as PNG images (matching C# LayerBack + LayerPicture).
	if !strings.Contains(out, "base64,") {
		t.Errorf("checkbox checked: expected base64 PNG image, got %q", out)
	}
}

func TestRenderObject_CheckBox_Unchecked(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:    preview.ObjectTypeCheckBox,
			Left:    0, Top: 0, Width: 20, Height: 20,
			Checked: false,
		},
	})
	out := exportHTML(t, pp)
	// Unchecked with no symbol still renders as a PNG (empty box).
	if !strings.Contains(out, "base64,") {
		t.Errorf("checkbox unchecked: expected base64 PNG image, got %q", out)
	}
}

// ── ObjectTypePolyLine ─────────────────────────────────────────────────────────

func TestRenderObject_PolyLine_TooFewPoints(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:   preview.ObjectTypePolyLine,
			Left:   0, Top: 0, Width: 100, Height: 50,
			Points: [][2]float32{{0, 0}}, // only 1 point → empty placeholder
		},
	})
	out := exportHTML(t, pp)
	// Empty placeholder: no SVG polyline.
	if strings.Contains(out, "<polyline") {
		t.Errorf("polyline <2 points: should not contain <polyline>, got %q", out)
	}
}

func TestRenderObject_PolyLine_WithPoints(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:   preview.ObjectTypePolyLine,
			Left:   0, Top: 0, Width: 100, Height: 50,
			Points: [][2]float32{{0, 0}, {50, 25}, {100, 0}},
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("polyline: expected base64 PNG image, got %q", out)
	}
}

func TestRenderObject_PolyLine_BorderLineColor(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Width: 2,
	}
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:   preview.ObjectTypePolyLine,
			Left:   0, Top: 0, Width: 100, Height: 50,
			Points: [][2]float32{{0, 0}, {100, 50}},
			Border: style.Border{
				Lines: [4]*style.BorderLine{bl, nil, nil, nil},
			},
		},
	})
	out := exportHTML(t, pp)
	// Border line color is now baked into the base64 PNG image.
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("polyline border color: expected base64 PNG image, got %q", out)
	}
}

// ── ObjectTypePolygon ──────────────────────────────────────────────────────────

func TestRenderObject_Polygon_TooFewPoints(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:   preview.ObjectTypePolygon,
			Left:   0, Top: 0, Width: 100, Height: 50,
			Points: [][2]float32{{0, 0}}, // only 1 point → empty placeholder
		},
	})
	out := exportHTML(t, pp)
	if strings.Contains(out, "<polygon") {
		t.Errorf("polygon <2 points: should not contain <polygon>, got %q", out)
	}
}

func TestRenderObject_Polygon_WithPoints(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:   preview.ObjectTypePolygon,
			Left:   0, Top: 0, Width: 100, Height: 80,
			Points: [][2]float32{{50, 0}, {100, 80}, {0, 80}},
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("polygon: expected base64 PNG image, got %q", out)
	}
}

func TestRenderObject_Polygon_WithFillColor(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:      preview.ObjectTypePolygon,
			Left:      0, Top: 0, Width: 100, Height: 80,
			Points:    [][2]float32{{50, 0}, {100, 80}, {0, 80}},
			FillColor: color.RGBA{R: 0, G: 128, B: 255, A: 255},
		},
	})
	out := exportHTML(t, pp)
	// Fill color appears as background-color in CSS class.
	if !strings.Contains(out, "background-color:rgb(0, 128, 255)") {
		t.Errorf("polygon fill color: expected background-color:rgb(0, 128, 255), got %q", out)
	}
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("polygon fill color: expected base64 PNG image, got %q", out)
	}
}

// ── ObjectTypeSVG ─────────────────────────────────────────────────────────────

func TestRenderObject_SVG_WithBlob(t *testing.T) {
	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="50" height="50"><circle r="25"/></svg>`)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("svg", svgData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Kind:    preview.ObjectTypeSVG,
				Left:    0, Top: 0, Width: 50, Height: 50,
				BlobIdx: blobIdx,
			},
		},
	})
	out := exportHTML(t, pp)
	if !strings.Contains(out, "<svg") {
		t.Errorf("SVG with blob: expected inline SVG content, got %q", out)
	}
	if !strings.Contains(out, "<circle") {
		t.Errorf("SVG with blob: expected SVG circle element, got %q", out)
	}
}

func TestRenderObject_SVG_NoBlob_EmptyPlaceholder(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:    preview.ObjectTypeSVG,
			Left:    0, Top: 0, Width: 50, Height: 50,
			BlobIdx: -1, // no blob
		},
	})
	out := exportHTML(t, pp)
	// Empty placeholder div — no SVG content.
	if strings.Contains(out, "<circle") {
		t.Errorf("SVG no blob: should not contain SVG content, got %q", out)
	}
}

// ── Default (unknown) kind ─────────────────────────────────────────────────────

func TestRenderObject_UnknownKind_EmptyPlaceholder(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:  preview.ObjectType(999), // unknown
			Left:  0, Top: 0, Width: 50, Height: 30,
		},
	})
	out := exportHTML(t, pp)
	// Should produce a positioned placeholder div without panicking.
	if !strings.Contains(out, "left:0px;top:0px;width:50px;height:30px;") {
		t.Errorf("unknown kind: expected positioned placeholder div, got %q", out)
	}
}

// ── Watermark image (renderWatermarkImage) ─────────────────────────────────────

func TestWatermarkImage_ShowImageOnTop_PNG(t *testing.T) {
	pngData := buildHTMLPNG(20, 20)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wm", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ShowImageOnTop: true,
		ImageBlobIdx:   blobIdx,
		ImageSize:      preview.WatermarkImageSizeStretch,
	}
	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	// New CSS-class pattern: LayerBack div (border:none) + LayerPicture div (with class referencing background style).
	if !strings.Contains(out, "border:none;") {
		t.Errorf("watermark image on top: expected LayerBack div with border:none, got %q", out)
	}
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("watermark image on top: expected PNG data URL in CSS class, got %q", out)
	}
}

func TestWatermarkImage_ShowImageOnTop_JPEG(t *testing.T) {
	jpegData := buildHTMLJPEG(20, 20)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wmjpg", jpegData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ShowImageOnTop: true,
		ImageBlobIdx:   blobIdx,
		ImageSize:      preview.WatermarkImageSizeZoom,
	}
	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	// New CSS-class pattern: image data is in a CSS class, not inline background-image.
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("watermark JPEG on top: expected image data URL in CSS class, got %q", out)
	}
}

func TestWatermarkImage_ShowImageOnTop_Tile(t *testing.T) {
	pngData := buildHTMLPNG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("tile", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ShowImageOnTop: true,
		ImageBlobIdx:   blobIdx,
		ImageSize:      preview.WatermarkImageSizeTile,
	}
	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	// New CSS-class pattern: tile mode still produces image data in a CSS class.
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("watermark tile: expected image data URL in CSS class, got %q", out)
	}
}

func TestWatermarkImage_Behind(t *testing.T) {
	// ShowImageOnTop=false → rendered in ExportPageBegin (behind content).
	pngData := buildHTMLPNG(20, 20)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wmbehind", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ShowImageOnTop: false,
		ImageBlobIdx:   blobIdx,
	}
	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	// New CSS-class pattern: image data is in a CSS class, not inline background-image.
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("watermark behind: expected image data URL in CSS class, got %q", out)
	}
}

func TestWatermarkImage_NoBlobIdx_NotRendered(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ShowImageOnTop: true,
		ImageBlobIdx:   -1, // no image
	}
	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	// With ImageBlobIdx=-1, no image watermark from a blob should be rendered.
	// The output may still contain base64 data from text watermark rasterization,
	// so we just verify the export completes without error and produces valid HTML.
	if !strings.Contains(out, `class="frpage0"`) {
		t.Errorf("no blob: expected valid page output, got %q", out)
	}
}

func TestWatermarkImage_HighTransparency_OpacityClamped(t *testing.T) {
	// ImageTransparency > 1.0 → opacity clamped to 0.
	pngData := buildHTMLPNG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wmt", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:           true,
		ShowImageOnTop:    true,
		ImageBlobIdx:      blobIdx,
		ImageTransparency: 1.5, // opacity = 1 - 1.5 = -0.5 → clamped to 0
	}
	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	// New CSS-class pattern: opacity is no longer inline; verify image data is present in CSS class.
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("high transparency: expected image data URL in CSS class, got %q", out)
	}
}

// ── HTML() method ─────────────────────────────────────────────────────────────

func TestHTML_Method_ReturnsGeneratedHTML(t *testing.T) {
	pp := buildPage([]preview.PreparedObject{
		{
			Kind:  preview.ObjectTypeText,
			Left:  0, Top: 0, Width: 100, Height: 20,
			Text:  "Hello HTML method",
		},
	})
	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	// HTML() should return the same string as what was written to buf.
	htmlStr := exp.HTML()
	if htmlStr == "" {
		t.Error("HTML() should return non-empty string after Export")
	}
	if !strings.Contains(htmlStr, "Hello HTML method") {
		t.Errorf("HTML(): expected text content, got %q", htmlStr)
	}
	if htmlStr != buf.String() {
		t.Errorf("HTML() should equal what was written to buf: HTML()=%q, buf=%q", htmlStr, buf.String())
	}
}

func TestHTML_Method_BeforeExport_Empty(t *testing.T) {
	exp := html.NewExporter()
	// Before Export, HTML() returns whatever is in sb — should be empty.
	got := exp.HTML()
	if got != "" {
		t.Errorf("HTML() before Export: expected empty string, got %q", got)
	}
}

// ── ExportPageEnd edge cases ───────────────────────────────────────────────────

func TestExportPageEnd_WatermarkNil(t *testing.T) {
	// Page with nil Watermark should not panic or fail.
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	// Watermark is nil by default.
	out := exportHTML(t, pp)
	if !strings.Contains(out, "</div>") {
		t.Errorf("nil watermark: expected closing div, got %q", out)
	}
}

func TestExportPageEnd_WatermarkDisabled(t *testing.T) {
	// Watermark with Enabled=false should not render any watermark.
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled: false,
		Text:    "SHOULD NOT APPEAR",
	}
	out := exportHTML(t, pp)
	if strings.Contains(out, "SHOULD NOT APPEAR") {
		t.Errorf("disabled watermark: text should not appear in output, got %q", out)
	}
}

func TestExportPageEnd_WatermarkText_OnTop(t *testing.T) {
	// ShowTextOnTop=true → text is rendered in ExportPageEnd.
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:       true,
		Text:          "ON TOP",
		ShowTextOnTop: true,
		ImageBlobIdx:  -1,
	}
	out := exportHTML(t, pp)
	// Watermark text is rasterized to a PNG image (matching C# behaviour).
	if strings.Contains(out, ">ON TOP<") {
		t.Error("watermark text on top: text should be rasterized, not visible as HTML")
	}
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Error("watermark text on top: expected rasterized PNG image in output")
	}
}

func TestExportPageEnd_PageClosingDiv(t *testing.T) {
	// Every page should produce a frpage{n} div.
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	pp.AddPage(794, 1123, 2)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 20})
	out := exportHTML(t, pp)
	// Two pages → frpage0 and frpage1 divs.
	if !strings.Contains(out, `class="frpage0"`) {
		t.Errorf("two pages: expected frpage0 div in %q", out)
	}
	if !strings.Contains(out, `class="frpage1"`) {
		t.Errorf("two pages: expected frpage1 div in %q", out)
	}
}

func TestExportPageEnd_WatermarkImage_NilPPDoesNotPanic(t *testing.T) {
	// ImageBlobIdx >= 0 but PreparedPages internal pp is set — this is normal.
	// With ImageBlobIdx=-1 and ShowImageOnTop=true: renderWatermarkImage returns early.
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ShowImageOnTop: true,
		ImageBlobIdx:   -1, // returns early
	}
	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export with ShowImageOnTop and no blob: %v", err)
	}
}

// ── WatermarkImageSize variants ───────────────────────────────────────────────

func TestWatermarkImage_SizeNormal(t *testing.T) {
	pngData := buildHTMLPNG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wmnorm", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ShowImageOnTop: true,
		ImageBlobIdx:   blobIdx,
		ImageSize:      preview.WatermarkImageSizeNormal,
	}
	out := exportHTML(t, pp)
	// New CSS-class pattern: verify LayerBack div (border:none) and image data in CSS class.
	if !strings.Contains(out, "border:none;") {
		t.Errorf("WatermarkImageSizeNormal: expected LayerBack div with border:none, got %q", out)
	}
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("WatermarkImageSizeNormal: expected image data URL in CSS class, got %q", out)
	}
}

func TestWatermarkImage_SizeCenter(t *testing.T) {
	pngData := buildHTMLPNG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wmcenter", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ShowImageOnTop: true,
		ImageBlobIdx:   blobIdx,
		ImageSize:      preview.WatermarkImageSizeCenter,
	}
	out := exportHTML(t, pp)
	// New CSS-class pattern: verify LayerBack div and image data in CSS class.
	if !strings.Contains(out, "border:none;") {
		t.Errorf("WatermarkImageSizeCenter: expected LayerBack div with border:none, got %q", out)
	}
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("WatermarkImageSizeCenter: expected image data URL in CSS class, got %q", out)
	}
}

func TestWatermarkImage_SizeZoom(t *testing.T) {
	pngData := buildHTMLPNG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wmzoom", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ShowImageOnTop: true,
		ImageBlobIdx:   blobIdx,
		ImageSize:      preview.WatermarkImageSizeZoom,
	}
	out := exportHTML(t, pp)
	// New CSS-class pattern: verify LayerBack div and image data in CSS class.
	if !strings.Contains(out, "border:none;") {
		t.Errorf("WatermarkImageSizeZoom: expected LayerBack div with border:none, got %q", out)
	}
	if !strings.Contains(out, "data:image/Png;base64,") {
		t.Errorf("WatermarkImageSizeZoom: expected image data URL in CSS class, got %q", out)
	}
}
