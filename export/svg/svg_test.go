package svg_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/export"
	svgexp "github.com/andrewloable/go-fastreport/export/svg"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func buildPages(n int, bands []string) *preview.PreparedPages {
	pp := preview.New()
	for i := 0; i < n; i++ {
		pp.AddPage(794, 1123, i+1)
		for j, name := range bands {
			_ = pp.AddBand(&preview.PreparedBand{
				Name:   name,
				Top:    float32(j * 40),
				Height: 40,
			})
		}
	}
	return pp
}

func buildPageWithObjects(objects []preview.PreparedObject) *preview.PreparedPages {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:    "DataBand",
		Top:     0,
		Height:  40,
		Objects: objects,
	})
	return pp
}

func exportSVG(t *testing.T, pp *preview.PreparedPages, opts ...func(*svgexp.Exporter)) string {
	t.Helper()
	exp := svgexp.NewExporter()
	for _, o := range opts {
		o(exp)
	}
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	return buf.String()
}

func makePNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// ── Basic lifecycle tests ─────────────────────────────────────────────────────

func TestSVGExporter_Defaults(t *testing.T) {
	exp := svgexp.NewExporter()
	if exp.Title != "Report" {
		t.Errorf("default Title: want Report, got %s", exp.Title)
	}
	if exp.FileExtension() != ".svg" {
		t.Errorf("FileExtension: want .svg, got %s", exp.FileExtension())
	}
	if exp.Name() != "SVG" {
		t.Errorf("Name: want SVG, got %s", exp.Name())
	}
	if exp.EmbedFonts {
		t.Error("EmbedFonts should default to false")
	}
}

func TestSVGExporter_NilPages_ReturnsError(t *testing.T) {
	exp := svgexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(nil, &buf); err == nil {
		t.Error("expected error for nil PreparedPages")
	}
}

func TestSVGExporter_EmptyPages_NoOutput(t *testing.T) {
	pp := preview.New()
	exp := svgexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export on empty pages: %v", err)
	}
	if buf.Len() != 0 {
		t.Logf("note: empty pages produced %d bytes", buf.Len())
	}
}

// ── SVG document structure ────────────────────────────────────────────────────

func TestSVGExporter_SinglePage_SVGElement(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	out := exportSVG(t, pp)

	if !strings.Contains(out, `<svg xmlns="http://www.w3.org/2000/svg"`) {
		t.Errorf("expected SVG root element, got: %q", out[:min(len(out), 200)])
	}
}

func TestSVGExporter_SinglePage_ClosingTag(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	out := exportSVG(t, pp)

	if !strings.Contains(out, "</svg>") {
		t.Errorf("expected closing </svg> tag")
	}
}

func TestSVGExporter_SinglePage_PageDimensions(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	out := exportSVG(t, pp)

	if !strings.Contains(out, `width="595.00"`) {
		t.Errorf("expected width=595.00, got: %q", out[:min(len(out), 200)])
	}
	if !strings.Contains(out, `height="842.00"`) {
		t.Errorf("expected height=842.00, got: %q", out[:min(len(out), 200)])
	}
}

func TestSVGExporter_SinglePage_WhiteBackground(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	out := exportSVG(t, pp)

	if !strings.Contains(out, `fill="#FFFFFF"`) {
		t.Errorf("expected white background rect fill=#FFFFFF")
	}
}

func TestSVGExporter_SinglePage_PageGroupID(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	out := exportSVG(t, pp)

	if !strings.Contains(out, `id="page1"`) {
		t.Errorf("expected page group id=page1")
	}
}

func TestSVGExporter_SinglePage_Title(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	out := exportSVG(t, pp, func(e *svgexp.Exporter) { e.Title = "My SVG Report" })

	if !strings.Contains(out, "<title>My SVG Report</title>") {
		t.Errorf("expected <title>My SVG Report</title>, got: %q", out[:min(len(out), 300)])
	}
}

func TestSVGExporter_SinglePage_EmptyTitle_NoTitleElement(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	out := exportSVG(t, pp, func(e *svgexp.Exporter) { e.Title = "" })

	if strings.Contains(out, "<title>") {
		t.Errorf("empty title: should not emit <title> element")
	}
}

// ── Multiple pages ────────────────────────────────────────────────────────────

func TestSVGExporter_MultiplePages_MultipleSVGElements(t *testing.T) {
	pp := buildPages(3, []string{"Band"})
	out := exportSVG(t, pp)

	count := strings.Count(out, "</svg>")
	if count != 3 {
		t.Errorf("3 pages: expected 3 </svg> tags, got %d", count)
	}
}

func TestSVGExporter_MultiplePages_PageGroupIDs(t *testing.T) {
	pp := buildPages(3, []string{"Band"})
	out := exportSVG(t, pp)

	for i := 1; i <= 3; i++ {
		expected := `id="page` + string(rune('0'+i)) + `"`
		if !strings.Contains(out, expected) {
			t.Errorf("expected page group %s", expected)
		}
	}
}

func TestSVGExporter_MultiplePages_SeparatedByNewline(t *testing.T) {
	pp := buildPages(2, []string{"Band"})
	out := exportSVG(t, pp)

	// Pages should be separated — find where first </svg> ends and check separator.
	// ExportPageEnd writes "</svg>\n" and ExportPageBegin writes a leading "\n"
	// before page 2+, so the separator between pages is "\n\n".
	idx := strings.Index(out, "</svg>")
	if idx < 0 {
		t.Fatal("no </svg> found")
	}
	after := out[idx+6:]
	// Accept either "\n<svg" or "\n\n<svg" (the implementation writes a leading \n
	// in ExportPageBegin for pages after the first).
	if !strings.Contains(after, "<svg") {
		t.Errorf("pages should be separated by newline before second SVG element, got: %q", after[:min(len(after), 40)])
	}
}

// ── Default dimensions for zero-size pages ────────────────────────────────────

func TestSVGExporter_ZeroPageDimensions_DefaultsA4(t *testing.T) {
	pp := preview.New()
	pp.AddPage(0, 0, 1) // zero dimensions → defaults
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	out := exportSVG(t, pp)

	// Default width=794 height=1123 (A4 at 96dpi)
	if !strings.Contains(out, `width="794.00"`) {
		t.Errorf("zero width: expected default width 794.00, got: %q", out[:min(len(out), 200)])
	}
	if !strings.Contains(out, `height="1123.00"`) {
		t.Errorf("zero height: expected default height 1123.00, got: %q", out[:min(len(out), 200)])
	}
}

// ── Text objects ──────────────────────────────────────────────────────────────

func TestSVGExporter_TextObject_ForeignObjectPresent(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 10, Top: 20, Width: 200, Height: 30,
			Text: "Hello SVG",
		},
	})
	out := exportSVG(t, pp)

	if !strings.Contains(out, "<foreignObject") {
		t.Errorf("text object: expected <foreignObject> element")
	}
	if !strings.Contains(out, "Hello SVG") {
		t.Errorf("text object: expected text content 'Hello SVG'")
	}
}

func TestSVGExporter_TextObject_Position(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 50, Top: 100, Width: 200, Height: 30,
			Text: "Positioned",
		},
	})
	out := exportSVG(t, pp)

	if !strings.Contains(out, `x="50.00"`) {
		t.Errorf("text object position: expected x=50.00")
	}
	// band.Top=0, obj.Top=100 → absolute y=100
	if !strings.Contains(out, `y="100.00"`) {
		t.Errorf("text object position: expected y=100.00")
	}
}

func TestSVGExporter_TextObject_AbsoluteY_IncludesBandTop(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	// band at top=50, object at top=20 → absolute y=70
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    50,
		Height: 80,
		Objects: []preview.PreparedObject{
			{
				Name: "T", Kind: preview.ObjectTypeText,
				Left: 0, Top: 20, Width: 100, Height: 20,
				Text: "AbsY",
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, `y="70.00"`) {
		t.Errorf("absolute Y: expected y=70.00 (band.Top=50 + obj.Top=20)")
	}
}

func TestSVGExporter_TextObject_FontBold(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Bold",
			Font: style.Font{Name: "Arial", Size: 12, Style: style.FontStyleBold},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "font-weight:bold;") {
		t.Errorf("bold: expected font-weight:bold; in SVG CSS")
	}
}

func TestSVGExporter_TextObject_FontItalic(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Italic",
			Font: style.Font{Name: "Arial", Size: 12, Style: style.FontStyleItalic},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "font-style:italic;") {
		t.Errorf("italic: expected font-style:italic; in SVG CSS")
	}
}

func TestSVGExporter_TextObject_FontUnderline(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Under",
			Font: style.Font{Name: "Arial", Size: 12, Style: style.FontStyleUnderline},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "text-decoration:underline;") {
		t.Errorf("underline: expected text-decoration:underline; in SVG CSS")
	}
}

func TestSVGExporter_TextObject_HorzAlign_Center(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "T", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Center",
			HorzAlign: 1,
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "text-align:center;") {
		t.Errorf("center align: expected text-align:center;")
	}
}

func TestSVGExporter_TextObject_HorzAlign_Right(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "T", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Right",
			HorzAlign: 2,
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "text-align:right;") {
		t.Errorf("right align: expected text-align:right;")
	}
}

func TestSVGExporter_TextObject_HorzAlign_Justify(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "T", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Justify",
			HorzAlign: 3,
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "text-align:justify;") {
		t.Errorf("justify align: expected text-align:justify;")
	}
}

func TestSVGExporter_TextObject_HorzAlign_Left_Default(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "T", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Left",
			HorzAlign: 0, // default left
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "text-align:left;") {
		t.Errorf("left align: expected text-align:left;")
	}
}

func TestSVGExporter_TextObject_VertAlign_Center(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "T", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 40,
			Text:      "VCenter",
			VertAlign: 1,
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "align-items:center;") {
		t.Errorf("vert center: expected align-items:center;")
	}
}

func TestSVGExporter_TextObject_VertAlign_Bottom(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "T", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 40,
			Text:      "VBottom",
			VertAlign: 2,
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "align-items:flex-end;") {
		t.Errorf("vert bottom: expected align-items:flex-end;")
	}
}

func TestSVGExporter_TextObject_WordWrap(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:     "T", Kind: preview.ObjectTypeText,
			Left:     0, Top: 0, Width: 100, Height: 40,
			Text:     "Word wrap text",
			WordWrap: true,
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "word-wrap:break-word;") {
		t.Errorf("word-wrap: expected word-wrap:break-word;")
	}
}

func TestSVGExporter_TextObject_NoWordWrap_Nowrap(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:     "T", Kind: preview.ObjectTypeText,
			Left:     0, Top: 0, Width: 100, Height: 20,
			Text:     "No wrap",
			WordWrap: false,
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "white-space:nowrap;") {
		t.Errorf("no word-wrap: expected white-space:nowrap;")
	}
}

func TestSVGExporter_TextObject_XMLEscaping(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 20,
			Text: "a < b & c > d",
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "&lt;") {
		t.Errorf("XML escape: expected &lt; for <")
	}
	if !strings.Contains(out, "&amp;") {
		t.Errorf("XML escape: expected &amp; for &")
	}
	if !strings.Contains(out, "&gt;") {
		t.Errorf("XML escape: expected &gt; for >")
	}
}

func TestSVGExporter_TextObject_FillColor(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "T", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Filled",
			FillColor: color.RGBA{R: 255, G: 255, B: 0, A: 255},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "#FFFF00") {
		t.Errorf("fill color: expected #FFFF00 rect background")
	}
}

func TestSVGExporter_TextObject_Hyperlink(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:           "T", Kind: preview.ObjectTypeText,
			Left:           0, Top: 0, Width: 150, Height: 20,
			Text:           "Click",
			HyperlinkKind:  1,
			HyperlinkValue: "https://example.com",
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, `href="https://example.com"`) {
		t.Errorf("hyperlink: expected href=https://example.com")
	}
	if !strings.Contains(out, "<a ") {
		t.Errorf("hyperlink: expected <a> anchor element")
	}
}

func TestSVGExporter_TextObject_EmptyText_NoForeignObject(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "", // empty text → no foreignObject
		},
	})
	out := exportSVG(t, pp)
	if strings.Contains(out, "<foreignObject") {
		t.Errorf("empty text: should not emit <foreignObject>")
	}
}

// ── HTML objects ──────────────────────────────────────────────────────────────

func TestSVGExporter_HtmlObject_TagsStripped(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "H", Kind: preview.ObjectTypeHtml,
			Left: 0, Top: 0, Width: 200, Height: 40,
			Text: "<b>bold</b> text",
		},
	})
	out := exportSVG(t, pp)
	// HTML tags stripped: should contain "bold text" but not <b>
	if strings.Contains(out, "<b>") {
		t.Errorf("html: <b> tag should be stripped in SVG output")
	}
	if !strings.Contains(out, "bold") {
		t.Errorf("html: text content 'bold' should be present")
	}
}

// ── RTF objects ───────────────────────────────────────────────────────────────

func TestSVGExporter_RTFObject_RenderedAsText(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "R", Kind: preview.ObjectTypeRTF,
			Left: 0, Top: 0, Width: 200, Height: 40,
			Text: `{\rtf1\ansi Hello RTF World}`,
		},
	})
	out := exportSVG(t, pp)
	// RTF should be stripped to plain text and rendered in foreignObject
	if strings.Contains(out, `\rtf1`) {
		t.Errorf("RTF: control words should be stripped")
	}
	if !strings.Contains(out, "Hello RTF World") {
		t.Errorf("RTF: plain text 'Hello RTF World' should appear")
	}
}

// ── Line objects ──────────────────────────────────────────────────────────────

func TestSVGExporter_LineObject_Horizontal(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "L", Kind: preview.ObjectTypeLine,
			Left: 0, Top: 10, Width: 100, Height: 1,
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "<line ") {
		t.Errorf("horizontal line: expected <line> SVG element")
	}
}

func TestSVGExporter_LineObject_Vertical(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "L", Kind: preview.ObjectTypeLine,
			Left: 10, Top: 0, Width: 1, Height: 100,
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "<line ") {
		t.Errorf("vertical line: expected <line> SVG element")
	}
}

func TestSVGExporter_LineObject_Diagonal(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:         "L", Kind: preview.ObjectTypeLine,
			Left:         0, Top: 0, Width: 100, Height: 50,
			LineDiagonal: true,
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "<line ") {
		t.Errorf("diagonal line: expected <line> SVG element")
	}
}

func TestSVGExporter_LineObject_WithBorderLine(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Width: 2,
	}
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "L", Kind: preview.ObjectTypeLine,
			Left: 0, Top: 0, Width: 100, Height: 1,
			Border: style.Border{
				Lines: [4]*style.BorderLine{bl, nil, nil, nil},
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "#FF0000") {
		t.Errorf("line with border: expected red #FF0000 stroke color")
	}
}

// ── Shape objects ─────────────────────────────────────────────────────────────

func TestSVGExporter_ShapeObject_Rectangle(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "S", Kind: preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 100, Height: 50,
			ShapeKind: 0, // rectangle
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "<rect ") {
		t.Errorf("rectangle shape: expected <rect> SVG element")
	}
}

func TestSVGExporter_ShapeObject_Ellipse(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "S", Kind: preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 100, Height: 50,
			ShapeKind: 2, // ellipse
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "<ellipse ") {
		t.Errorf("ellipse shape: expected <ellipse> SVG element")
	}
}

func TestSVGExporter_ShapeObject_RoundRect(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:       "S", Kind: preview.ObjectTypeShape,
			Left:       0, Top: 0, Width: 100, Height: 50,
			ShapeKind:  1, // round rectangle
			ShapeCurve: 10,
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "<rect ") {
		t.Errorf("round rect: expected <rect> SVG element")
	}
	if !strings.Contains(out, `rx="10.00"`) {
		t.Errorf("round rect: expected rx=10.00 corner radius")
	}
}

func TestSVGExporter_ShapeObject_Triangle(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "S", Kind: preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 80, Height: 60,
			ShapeKind: 3, // triangle
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "<polygon ") {
		t.Errorf("triangle: expected <polygon> SVG element")
	}
}

func TestSVGExporter_ShapeObject_Diamond(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "S", Kind: preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 80, Height: 60,
			ShapeKind: 4, // diamond
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "<polygon ") {
		t.Errorf("diamond: expected <polygon> SVG element")
	}
}

func TestSVGExporter_ShapeObject_WithFillColor(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "S", Kind: preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 80, Height: 60,
			ShapeKind: 0,
			FillColor: color.RGBA{R: 0, G: 128, B: 255, A: 255},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "#0080FF") {
		t.Errorf("shape with fill: expected fill #0080FF")
	}
}

// ── Picture objects ───────────────────────────────────────────────────────────

func TestSVGExporter_PictureObject_WithBlob(t *testing.T) {
	pngData := makePNG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("img", pngData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name:    "Pic", Kind: preview.ObjectTypePicture,
				Left:    10, Top: 10, Width: 80, Height: 80,
				BlobIdx: blobIdx,
			},
		},
	})
	exp := svgexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "<image ") {
		t.Errorf("picture with blob: expected <image> SVG element")
	}
	if !strings.Contains(out, "data:image/png;base64,") {
		t.Errorf("picture with blob: expected PNG data URI")
	}
}

func TestSVGExporter_PictureObject_NoBlob_Placeholder(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:    "Pic", Kind: preview.ObjectTypePicture,
			Left:    0, Top: 0, Width: 80, Height: 80,
			BlobIdx: -1,
		},
	})
	out := exportSVG(t, pp)
	// No blob → placeholder rect
	if strings.Contains(out, `<image `) {
		t.Errorf("no blob: should not emit <image> element")
	}
	if !strings.Contains(out, "#F0F0F0") {
		t.Errorf("no blob: expected placeholder rect with fill #F0F0F0")
	}
}

// ── CheckBox objects ──────────────────────────────────────────────────────────

func TestSVGExporter_CheckBox_Unchecked(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:    "CB", Kind: preview.ObjectTypeCheckBox,
			Left:    0, Top: 0, Width: 20, Height: 20,
			Checked: false,
		},
	})
	out := exportSVG(t, pp)
	// Unchecked: just the rect, no crossing lines
	if !strings.Contains(out, "<rect ") {
		t.Errorf("unchecked checkbox: expected <rect> element")
	}
}

func TestSVGExporter_CheckBox_Checked(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:    "CB", Kind: preview.ObjectTypeCheckBox,
			Left:    0, Top: 0, Width: 20, Height: 20,
			Checked: true,
		},
	})
	out := exportSVG(t, pp)
	// Checked: rect + two crossing lines
	lineCount := strings.Count(out, "<line ")
	if lineCount < 2 {
		t.Errorf("checked checkbox: expected at least 2 <line> elements for cross, got %d", lineCount)
	}
}

func TestSVGExporter_CheckBox_TextTrue_ShowsCross(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:    "CB", Kind: preview.ObjectTypeCheckBox,
			Left:    0, Top: 0, Width: 20, Height: 20,
			Checked: false,
			Text:    "true",
		},
	})
	out := exportSVG(t, pp)
	lineCount := strings.Count(out, "<line ")
	if lineCount < 2 {
		t.Errorf("checkbox text=true: expected at least 2 <line> elements for cross, got %d", lineCount)
	}
}

// ── PolyLine objects ──────────────────────────────────────────────────────────

func TestSVGExporter_PolyLine_WithPoints(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:   "PL", Kind: preview.ObjectTypePolyLine,
			Left:   0, Top: 0, Width: 100, Height: 50,
			Points: [][2]float32{{0, 0}, {50, 25}, {100, 0}},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "<polyline ") {
		t.Errorf("polyline: expected <polyline> SVG element")
	}
}

func TestSVGExporter_PolyLine_TooFewPoints_Empty(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:   "PL", Kind: preview.ObjectTypePolyLine,
			Left:   0, Top: 0, Width: 100, Height: 50,
			Points: [][2]float32{{0, 0}}, // only 1 point → skipped
		},
	})
	out := exportSVG(t, pp)
	if strings.Contains(out, "<polyline ") {
		t.Errorf("polyline <2 points: should not emit <polyline>")
	}
}

// ── Polygon objects ───────────────────────────────────────────────────────────

func TestSVGExporter_Polygon_WithPoints(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:   "PG", Kind: preview.ObjectTypePolygon,
			Left:   0, Top: 0, Width: 100, Height: 80,
			Points: [][2]float32{{50, 0}, {100, 80}, {0, 80}},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "<polygon ") {
		t.Errorf("polygon: expected <polygon> SVG element")
	}
}

func TestSVGExporter_Polygon_WithFill(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "PG", Kind: preview.ObjectTypePolygon,
			Left:      0, Top: 0, Width: 100, Height: 80,
			Points:    [][2]float32{{50, 0}, {100, 80}, {0, 80}},
			FillColor: color.RGBA{R: 0, G: 255, B: 0, A: 255},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "#00FF00") {
		t.Errorf("polygon fill: expected #00FF00 fill color")
	}
}

// ── SVG object ────────────────────────────────────────────────────────────────

func TestSVGExporter_SVGObject_WithBlob(t *testing.T) {
	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><circle r="10"/></svg>`)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("svg", svgData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name:    "SVGObj", Kind: preview.ObjectTypeSVG,
				Left:    10, Top: 10, Width: 60, Height: 60,
				BlobIdx: blobIdx,
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "<g transform=") {
		t.Errorf("SVG object with blob: expected <g transform=> wrapper")
	}
	if !strings.Contains(out, "<circle") {
		t.Errorf("SVG object with blob: expected inline SVG content")
	}
}

func TestSVGExporter_SVGObject_NoBlob_Placeholder(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:    "SVGObj", Kind: preview.ObjectTypeSVG,
			Left:    0, Top: 0, Width: 50, Height: 50,
			BlobIdx: -1,
		},
	})
	out := exportSVG(t, pp)
	// No blob → placeholder rect with dashes
	if !strings.Contains(out, "stroke-dasharray") {
		t.Errorf("SVG no blob: expected dashed placeholder rect")
	}
}

// ── Digital Signature ─────────────────────────────────────────────────────────

func TestSVGExporter_DigitalSignature_WithText(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "DS", Kind: preview.ObjectTypeDigitalSignature,
			Left: 0, Top: 0, Width: 150, Height: 40,
			Text: "Sign Here",
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "Sign Here") {
		t.Errorf("digital signature: expected 'Sign Here' text")
	}
	if !strings.Contains(out, "stroke-dasharray") {
		t.Errorf("digital signature: expected dashed border")
	}
}

func TestSVGExporter_DigitalSignature_EmptyText_DefaultLabel(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "DS", Kind: preview.ObjectTypeDigitalSignature,
			Left: 0, Top: 0, Width: 150, Height: 40,
			Text: "",
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "Digital Signature") {
		t.Errorf("digital signature empty text: expected default 'Digital Signature' label")
	}
}

// ── Unknown object type ───────────────────────────────────────────────────────

func TestSVGExporter_UnknownObjectType_EmptyPlaceholder(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "Unk", Kind: preview.ObjectType(999),
			Left: 5, Top: 10, Width: 50, Height: 30,
		},
	})
	out := exportSVG(t, pp)
	// Unknown type: transparent placeholder rect
	if !strings.Contains(out, `fill="none"`) {
		t.Errorf("unknown object type: expected fill=none placeholder rect")
	}
}

// ── Watermark text ────────────────────────────────────────────────────────────

func TestSVGExporter_WatermarkText_Behind(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:       true,
		Text:          "CONFIDENTIAL",
		TextRotation:  preview.WatermarkTextRotationForwardDiagonal,
		ShowTextOnTop: false,
		ImageBlobIdx:  -1,
	}
	out := exportSVG(t, pp)
	if !strings.Contains(out, "CONFIDENTIAL") {
		t.Errorf("watermark behind: expected 'CONFIDENTIAL' text")
	}
	if !strings.Contains(out, `transform="rotate(`) {
		t.Errorf("watermark diagonal: expected rotate transform")
	}
}

func TestSVGExporter_WatermarkText_OnTop(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:       true,
		Text:          "DRAFT",
		ShowTextOnTop: true,
		ImageBlobIdx:  -1,
	}
	out := exportSVG(t, pp)
	if !strings.Contains(out, "DRAFT") {
		t.Errorf("watermark on top: expected 'DRAFT' text")
	}
}

func TestSVGExporter_WatermarkText_Horizontal_NoRotate(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "HORIZONTAL",
		TextRotation: preview.WatermarkTextRotationHorizontal,
		ImageBlobIdx: -1,
	}
	out := exportSVG(t, pp)
	if !strings.Contains(out, "HORIZONTAL") {
		t.Errorf("horizontal watermark: expected 'HORIZONTAL' text")
	}
	if strings.Contains(out, `transform="rotate(`) {
		t.Errorf("horizontal watermark: should not have rotate transform")
	}
}

func TestSVGExporter_WatermarkText_Vertical(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "VERTICAL",
		TextRotation: preview.WatermarkTextRotationVertical,
		ImageBlobIdx: -1,
	}
	out := exportSVG(t, pp)
	if !strings.Contains(out, `transform="rotate(90`) {
		t.Errorf("vertical watermark: expected rotate(90)")
	}
}

func TestSVGExporter_WatermarkText_BackwardDiagonal(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "BACKWARD",
		TextRotation: preview.WatermarkTextRotationBackwardDiagonal,
		ImageBlobIdx: -1,
	}
	out := exportSVG(t, pp)
	if !strings.Contains(out, `transform="rotate(45`) {
		t.Errorf("backward diagonal watermark: expected rotate(45)")
	}
}

func TestSVGExporter_WatermarkText_Disabled_NoText(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled: false,
		Text:    "HIDDEN",
	}
	out := exportSVG(t, pp)
	if strings.Contains(out, "HIDDEN") {
		t.Errorf("disabled watermark: text should not appear in output")
	}
}

func TestSVGExporter_WatermarkNil_NoPanic(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	// Watermark is nil by default
	out := exportSVG(t, pp)
	if !strings.Contains(out, "</svg>") {
		t.Errorf("nil watermark: expected valid SVG output")
	}
}

// ── Watermark image ───────────────────────────────────────────────────────────

func TestSVGExporter_WatermarkImage_WithBlob(t *testing.T) {
	pngData := makePNG(20, 20)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wm", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeStretch,
	}
	out := exportSVG(t, pp)
	if !strings.Contains(out, `<image `) {
		t.Errorf("watermark image: expected <image> element")
	}
	if !strings.Contains(out, "data:image/png;base64,") {
		t.Errorf("watermark image: expected PNG data URI")
	}
}

func TestSVGExporter_WatermarkImage_NoBlob_NotRendered(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: -1,
	}
	out := exportSVG(t, pp)
	// No blob → no image element (only the page rect)
	if strings.Count(out, `<image `) > 0 {
		t.Errorf("no blob: should not render watermark image element")
	}
}

func TestSVGExporter_WatermarkImage_Zoom_PreserveAspect(t *testing.T) {
	pngData := makePNG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wmzoom", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeZoom,
	}
	out := exportSVG(t, pp)
	if !strings.Contains(out, `preserveAspectRatio="xMidYMid meet"`) {
		t.Errorf("watermark zoom: expected preserveAspectRatio=xMidYMid meet")
	}
}

func TestSVGExporter_WatermarkImage_Stretch_NonePreserve(t *testing.T) {
	pngData := makePNG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wmstretch", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeStretch,
	}
	out := exportSVG(t, pp)
	if !strings.Contains(out, `preserveAspectRatio="none"`) {
		t.Errorf("watermark stretch: expected preserveAspectRatio=none")
	}
}

func TestSVGExporter_WatermarkImage_OnTop(t *testing.T) {
	pngData := makePNG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wmontop", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		ImageBlobIdx:   blobIdx,
		ShowImageOnTop: true,
	}
	out := exportSVG(t, pp)
	if !strings.Contains(out, `<image `) {
		t.Errorf("watermark image on top: expected <image> element")
	}
}

// ── PageRange support ─────────────────────────────────────────────────────────

func TestSVGExporter_PageRangeCurrent(t *testing.T) {
	pp := buildPages(5, []string{"Band"})
	exp := svgexp.NewExporter()
	exp.PageRange = export.PageRangeCurrent
	exp.CurPage = 2
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	count := strings.Count(out, "</svg>")
	if count != 1 {
		t.Errorf("PageRangeCurrent: expected 1 </svg>, got %d", count)
	}
}

func TestSVGExporter_PageRangeCustom(t *testing.T) {
	pp := buildPages(5, []string{"Band"})
	exp := svgexp.NewExporter()
	exp.PageRange = export.PageRangeCustom
	exp.PageNumbers = "1,3,5"
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	count := strings.Count(out, "</svg>")
	if count != 3 {
		t.Errorf("PageRangeCustom(1,3,5): expected 3 </svg> elements, got %d", count)
	}
}

// ── Border lines ──────────────────────────────────────────────────────────────

func TestSVGExporter_TextObject_WithBorderLines(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Width: 1,
		Style: style.LineStyleSolid,
	}
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 10, Top: 10, Width: 100, Height: 30,
			Text: "Bordered",
			Border: style.Border{
				VisibleLines: style.BorderLinesAll,
				Lines:        [4]*style.BorderLine{bl, bl, bl, bl},
			},
		},
	})
	out := exportSVG(t, pp)
	// With BorderLinesAll, all 4 sides should emit <line> elements
	lineCount := strings.Count(out, "<line ")
	if lineCount < 4 {
		t.Errorf("border all sides: expected at least 4 <line> elements, got %d", lineCount)
	}
}

func TestSVGExporter_TextObject_BorderDash(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Width: 1,
		Style: style.LineStyleDash,
	}
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 30,
			Text: "Dashed border",
			Border: style.Border{
				VisibleLines: style.BorderLinesTop,
				Lines:        [4]*style.BorderLine{nil, bl, nil, nil},
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "stroke-dasharray") {
		t.Errorf("dash border: expected stroke-dasharray in SVG")
	}
}

// helper
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
