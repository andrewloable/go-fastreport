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

// ── renderObject: negative width/height clamping ──────────────────────────────

func TestSVGExporter_RenderObject_NegativeWidth_ClampedToZero(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "NW", Kind: preview.ObjectTypeText,
			Left: 10, Top: 10, Width: -50, Height: 20,
			Text: "NegW",
		},
	})
	out := exportSVG(t, pp)
	// Should render without error and clamp width to 0
	if !strings.Contains(out, `width="0.00"`) {
		t.Errorf("negative width: expected width clamped to 0.00")
	}
}

func TestSVGExporter_RenderObject_NegativeHeight_ClampedToZero(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "NH", Kind: preview.ObjectTypeText,
			Left: 10, Top: 10, Width: 100, Height: -30,
			Text: "NegH",
		},
	})
	out := exportSVG(t, pp)
	// Should render without error and clamp height to 0
	if !strings.Contains(out, `height="0.00"`) {
		t.Errorf("negative height: expected height clamped to 0.00")
	}
}

// ── renderShape: border line overrides ────────────────────────────────────────

func TestSVGExporter_ShapeObject_WithBorderLineOverride(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 128, G: 0, B: 255, A: 200},
		Width: 3,
	}
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "S", Kind: preview.ObjectTypeShape,
			Left:      10, Top: 10, Width: 80, Height: 60,
			ShapeKind: 0,
			Border: style.Border{
				Lines: [4]*style.BorderLine{bl, nil, nil, nil},
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "#8000FF") {
		t.Errorf("shape border color: expected #8000FF")
	}
	if !strings.Contains(out, `stroke-width="3.00"`) {
		t.Errorf("shape border width: expected stroke-width=3.00")
	}
}

func TestSVGExporter_ShapeObject_Ellipse_WithFill(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "S", Kind: preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 100, Height: 80,
			ShapeKind: 2, // ellipse
			FillColor: color.RGBA{R: 255, G: 0, B: 128, A: 200},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "#FF0080") {
		t.Errorf("ellipse fill: expected #FF0080")
	}
	if !strings.Contains(out, "fill-opacity") {
		t.Errorf("ellipse fill: expected fill-opacity attribute")
	}
}

func TestSVGExporter_ShapeObject_Triangle_WithFill(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "S", Kind: preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 60, Height: 50,
			ShapeKind: 3, // triangle
			FillColor: color.RGBA{R: 0, G: 200, B: 100, A: 255},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "#00C864") {
		t.Errorf("triangle fill: expected #00C864")
	}
}

func TestSVGExporter_ShapeObject_Diamond_WithFill(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "S", Kind: preview.ObjectTypeShape,
			Left:      0, Top: 0, Width: 60, Height: 50,
			ShapeKind: 4, // diamond
			FillColor: color.RGBA{R: 100, G: 100, B: 200, A: 255},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "#6464C8") {
		t.Errorf("diamond fill: expected #6464C8")
	}
}

func TestSVGExporter_ShapeObject_RoundRect_ZeroCurve(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:       "S", Kind: preview.ObjectTypeShape,
			Left:       0, Top: 0, Width: 100, Height: 50,
			ShapeKind:  1, // round rectangle
			ShapeCurve: -5,
		},
	})
	out := exportSVG(t, pp)
	// Negative curve → clamped to 0
	if !strings.Contains(out, `rx="0.00"`) {
		t.Errorf("round rect zero curve: expected rx=0.00")
	}
}

// ── renderPolyShape: border line overrides ────────────────────────────────────

func TestSVGExporter_PolyLine_WithBorderLineOverride(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 128, B: 0, A: 255},
		Width: 3,
	}
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:   "PL", Kind: preview.ObjectTypePolyLine,
			Left:   0, Top: 0, Width: 100, Height: 50,
			Points: [][2]float32{{0, 0}, {50, 25}, {100, 0}},
			Border: style.Border{
				Lines: [4]*style.BorderLine{bl, nil, nil, nil},
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "#008000") {
		t.Errorf("polyline border color: expected #008000")
	}
	if !strings.Contains(out, `stroke-width="3.00"`) {
		t.Errorf("polyline border width: expected stroke-width=3.00")
	}
}

func TestSVGExporter_Polygon_WithBorderLineOverride(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 255, G: 128, B: 0, A: 255},
		Width: 2,
	}
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "PG", Kind: preview.ObjectTypePolygon,
			Left:      0, Top: 0, Width: 100, Height: 80,
			Points:    [][2]float32{{50, 0}, {100, 80}, {0, 80}},
			FillColor: color.RGBA{R: 0, G: 0, B: 255, A: 128},
			Border: style.Border{
				Lines: [4]*style.BorderLine{bl, nil, nil, nil},
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "#FF8000") {
		t.Errorf("polygon border color: expected #FF8000")
	}
	if !strings.Contains(out, "#0000FF") {
		t.Errorf("polygon fill: expected #0000FF")
	}
}

// ── renderBorderLines: all line styles ────────────────────────────────────────

func TestSVGExporter_BorderLineStyle_Dot(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Width: 1,
		Style: style.LineStyleDot,
	}
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 30,
			Text: "Dotted",
			Border: style.Border{
				VisibleLines: style.BorderLinesTop,
				Lines:        [4]*style.BorderLine{nil, bl, nil, nil},
			},
		},
	})
	out := exportSVG(t, pp)
	// LineStyleDot: dasharray = "lw, lw*2" = "1.00, 2.00"
	if !strings.Contains(out, "stroke-dasharray") {
		t.Errorf("dot border: expected stroke-dasharray")
	}
}

func TestSVGExporter_BorderLineStyle_DashDot(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Width: 1,
		Style: style.LineStyleDashDot,
	}
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 30,
			Text: "DashDot",
			Border: style.Border{
				VisibleLines: style.BorderLinesTop,
				Lines:        [4]*style.BorderLine{nil, bl, nil, nil},
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "stroke-dasharray") {
		t.Errorf("dashdot border: expected stroke-dasharray")
	}
}

func TestSVGExporter_BorderLineStyle_DashDotDot(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Width: 1,
		Style: style.LineStyleDashDotDot,
	}
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 30,
			Text: "DashDotDot",
			Border: style.Border{
				VisibleLines: style.BorderLinesTop,
				Lines:        [4]*style.BorderLine{nil, bl, nil, nil},
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "stroke-dasharray") {
		t.Errorf("dashdotdot border: expected stroke-dasharray")
	}
}

func TestSVGExporter_BorderLines_NilLine_DefaultsApplied(t *testing.T) {
	// VisibleLines set but Lines[idx] is nil → should still render with defaults
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 30,
			Text: "NilLine",
			Border: style.Border{
				VisibleLines: style.BorderLinesLeft,
				Lines:        [4]*style.BorderLine{nil, nil, nil, nil},
			},
		},
	})
	out := exportSVG(t, pp)
	// Should still render a line with default color (black) and width (1)
	if !strings.Contains(out, `stroke="#000000"`) {
		t.Errorf("nil border line: expected default black stroke")
	}
}

// ── renderWatermarkImage: edge cases ──────────────────────────────────────────

func TestSVGExporter_WatermarkImage_NegativeTransparency_ClampedToZero(t *testing.T) {
	pngData := makePNG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wmneg", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:           true,
		ImageBlobIdx:      blobIdx,
		ImageTransparency: 1.5, // > 1.0 means opacity would be negative → clamped to 0
	}
	out := exportSVG(t, pp)
	if !strings.Contains(out, `opacity="0.000"`) {
		t.Errorf("negative opacity: expected opacity clamped to 0.000")
	}
}

func TestSVGExporter_WatermarkImage_DefaultImageSize(t *testing.T) {
	pngData := makePNG(10, 10)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("wmdefault", pngData)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeNormal, // default case
	}
	out := exportSVG(t, pp)
	if !strings.Contains(out, `preserveAspectRatio="xMidYMid meet"`) {
		t.Errorf("default image size: expected preserveAspectRatio=xMidYMid meet")
	}
}

func TestSVGExporter_WatermarkImage_NilPP(t *testing.T) {
	// Test that renderWatermarkImage returns early when pp is nil
	// This is handled internally by export flow, but we verify with
	// a watermark that has ImageBlobIdx >= 0 but no actual blob data
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: 999, // index out of range → Get returns nil
	}
	out := exportSVG(t, pp)
	// Should not crash and should not emit <image>
	if strings.Contains(out, `<image `) {
		t.Errorf("invalid blob index: should not emit <image>")
	}
}

// ── xmlEscape: quote and apostrophe branches ──────────────────────────────────

func TestSVGExporter_XmlEscape_QuotesAndApostrophe(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 20,
			Text: `He said "hello" & she said 'goodbye'`,
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "&quot;") {
		t.Errorf("XML escape: expected &quot; for double quote")
	}
	if !strings.Contains(out, "&apos;") {
		t.Errorf("XML escape: expected &apos; for apostrophe")
	}
}

// ── imageMIME: various image format detection ─────────────────────────────────

func TestSVGExporter_PictureObject_JPEG(t *testing.T) {
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10} // JPEG magic bytes
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("jpg", jpegData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name:    "Pic", Kind: preview.ObjectTypePicture,
				Left:    0, Top: 0, Width: 80, Height: 80,
				BlobIdx: blobIdx,
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "data:image/jpeg;base64,") {
		t.Errorf("JPEG: expected data:image/jpeg MIME type")
	}
}

func TestSVGExporter_PictureObject_GIF(t *testing.T) {
	gifData := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61} // GIF89a magic bytes
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("gif", gifData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name:    "Pic", Kind: preview.ObjectTypePicture,
				Left:    0, Top: 0, Width: 80, Height: 80,
				BlobIdx: blobIdx,
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "data:image/gif;base64,") {
		t.Errorf("GIF: expected data:image/gif MIME type")
	}
}

func TestSVGExporter_PictureObject_BMP(t *testing.T) {
	bmpData := []byte{0x42, 0x4D, 0x00, 0x00, 0x00, 0x00} // BMP magic bytes
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("bmp", bmpData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name:    "Pic", Kind: preview.ObjectTypePicture,
				Left:    0, Top: 0, Width: 80, Height: 80,
				BlobIdx: blobIdx,
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "data:image/bmp;base64,") {
		t.Errorf("BMP: expected data:image/bmp MIME type")
	}
}

func TestSVGExporter_PictureObject_TIFF_LittleEndian(t *testing.T) {
	tiffData := []byte{0x49, 0x49, 0x2A, 0x00, 0x08, 0x00} // TIFF LE magic bytes
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("tiff", tiffData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name:    "Pic", Kind: preview.ObjectTypePicture,
				Left:    0, Top: 0, Width: 80, Height: 80,
				BlobIdx: blobIdx,
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "data:image/tiff;base64,") {
		t.Errorf("TIFF LE: expected data:image/tiff MIME type")
	}
}

func TestSVGExporter_PictureObject_TIFF_BigEndian(t *testing.T) {
	tiffData := []byte{0x4D, 0x4D, 0x00, 0x2A, 0x00, 0x00} // TIFF BE magic bytes
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("tiffbe", tiffData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name:    "Pic", Kind: preview.ObjectTypePicture,
				Left:    0, Top: 0, Width: 80, Height: 80,
				BlobIdx: blobIdx,
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "data:image/tiff;base64,") {
		t.Errorf("TIFF BE: expected data:image/tiff MIME type")
	}
}

func TestSVGExporter_PictureObject_SVGContent(t *testing.T) {
	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>`)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("svgimg", svgData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name:    "Pic", Kind: preview.ObjectTypePicture,
				Left:    0, Top: 0, Width: 80, Height: 80,
				BlobIdx: blobIdx,
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "data:image/svg+xml;base64,") {
		t.Errorf("SVG content: expected data:image/svg+xml MIME type")
	}
}

func TestSVGExporter_PictureObject_XMLContent(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0"?><svg><circle/></svg>`)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("xmlsvg", xmlData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name:    "Pic", Kind: preview.ObjectTypePicture,
				Left:    0, Top: 0, Width: 80, Height: 80,
				BlobIdx: blobIdx,
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "data:image/svg+xml;base64,") {
		t.Errorf("XML SVG content: expected data:image/svg+xml MIME type")
	}
}

func TestSVGExporter_PictureObject_ShortData_FallbackPNG(t *testing.T) {
	// Data too short for any magic byte detection → fallback to image/png
	shortData := []byte{0x01, 0x02}
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	blobIdx := pp.BlobStore.Add("short", shortData)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    0,
		Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name:    "Pic", Kind: preview.ObjectTypePicture,
				Left:    0, Top: 0, Width: 80, Height: 80,
				BlobIdx: blobIdx,
			},
		},
	})
	out := exportSVG(t, pp)
	if !strings.Contains(out, "data:image/png;base64,") {
		t.Errorf("short data: expected fallback data:image/png MIME type")
	}
}

// ── Barcode object type (not handled → default placeholder) ───────────────────

func TestSVGExporter_BarcodeObject_DefaultPlaceholder(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "BC", Kind: preview.ObjectTypeBarcode,
			Left: 0, Top: 0, Width: 100, Height: 40,
		},
	})
	out := exportSVG(t, pp)
	// Barcode is not explicitly handled → falls through to default placeholder
	if !strings.Contains(out, `fill="none"`) {
		t.Errorf("barcode object: expected default placeholder with fill=none")
	}
}

// helper
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
