package html_test

import (
	"bytes"
	"fmt"
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/preview"
)

func buildPages(n int, bands []string) *preview.PreparedPages {
	pp := preview.New()
	for i := 0; i < n; i++ {
		pp.AddPage(794, 1123, i+1) // A4 at 96dpi approx
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

func TestExporter_BasicHTML(t *testing.T) {
	pp := buildPages(1, []string{"Header", "DataBand"})
	exp := html.NewExporter()
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	out := buf.String()
	// New DOCTYPE is HTML 4.01.
	if !strings.Contains(out, "<!DOCTYPE HTML PUBLIC") {
		t.Error("output should contain HTML 4.01 DOCTYPE")
	}
	// Page div is now frpage0 (not class="page").
	if !strings.Contains(out, `class="frpage0"`) {
		t.Error("output should contain frpage0 div")
	}
}

func TestExporter_CustomTitle(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	exp := html.NewExporter()
	exp.Title = "My Report"
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	if !strings.Contains(buf.String(), "My Report") {
		t.Error("HTML should contain custom title")
	}
}

func TestExporter_NoCSSWhenDisabled(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	exp := html.NewExporter()
	exp.EmbedCSS = false
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	if strings.Contains(buf.String(), "<style") {
		t.Error("CSS should not be embedded when EmbedCSS=false")
	}
}

func TestExporter_MultiplePages(t *testing.T) {
	pp := buildPages(3, []string{"Band"})
	exp := html.NewExporter()
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	out := buf.String()
	// Count frpage divs (frpage0, frpage1, frpage2).
	count := 0
	for i := 0; i < 3; i++ {
		if strings.Contains(out, fmt.Sprintf(`class="frpage%d"`, i)) {
			count++
		}
	}
	if count != 3 {
		t.Errorf("expected 3 frpage divs, got %d", count)
	}
}

func TestExporter_PageRangeCurrent(t *testing.T) {
	pp := buildPages(5, []string{"Band"})
	exp := html.NewExporter()
	exp.PageRange = export.PageRangeCurrent
	exp.CurPage = 2
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	out := buf.String()
	// Only one page rendered → frpage0 present, frpage1 not.
	if !strings.Contains(out, `class="frpage0"`) {
		t.Error("PageRangeCurrent: expected frpage0 div")
	}
	if strings.Contains(out, `class="frpage1"`) {
		t.Error("PageRangeCurrent: should not have frpage1 div")
	}
}

func TestExporter_EmptyPages(t *testing.T) {
	pp := preview.New()
	exp := html.NewExporter()
	var buf bytes.Buffer

	// Empty pages → no pages to export → Export returns nil without writing.
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export on empty pages: %v", err)
	}
	// No pages selected — output may be empty, which is valid.
}

func TestExporter_NilPages(t *testing.T) {
	exp := html.NewExporter()
	var buf bytes.Buffer
	err := exp.Export(nil, &buf)
	if err == nil {
		t.Error("expected error for nil PreparedPages")
	}
}

func TestExporter_HTMLEscaping(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "<script>", Top: 0, Height: 30})

	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	// Band names no longer appear in the output (flat rendering, no band divs).
	// Just verify the output doesn't have raw <script> tags.
	if strings.Contains(buf.String(), "<script>") {
		t.Error("HTML should not contain raw <script> from band names")
	}
}

func buildLayeredPages() *preview.PreparedPages {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	// Band 1 at top=0, Band 2 at top=50 — objects overlap across bands.
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band1",
		Top:    0,
		Height: 50,
		Objects: []preview.PreparedObject{
			{Name: "Obj1", Kind: preview.ObjectTypeText, Left: 10, Top: 5, Width: 100, Height: 20, Text: "Layer1"},
		},
	})
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band2",
		Top:    30, // overlaps with Band1 vertically
		Height: 40,
		Objects: []preview.PreparedObject{
			{Name: "Obj2", Kind: preview.ObjectTypeText, Left: 20, Top: 5, Width: 100, Height: 20, Text: "Layer2"},
		},
	})
	return pp
}

func TestExporter_LayersMode_NosBandDivs(t *testing.T) {
	pp := buildLayeredPages()
	exp := html.NewExporter()
	exp.Layers = true
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export (layers): %v", err)
	}
	out := buf.String()

	// In layers mode, there should be no band wrapper divs.
	if strings.Contains(out, `class="band"`) {
		t.Error("layers mode should not emit band wrapper divs")
	}
}

func TestExporter_LayersMode_ZIndex(t *testing.T) {
	pp := buildLayeredPages()
	exp := html.NewExporter()
	exp.Layers = true
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export (layers): %v", err)
	}
	out := buf.String()

	// Objects should have z-index CSS.
	if !strings.Contains(out, "z-index:") {
		t.Error("layers mode should include z-index on objects")
	}
	// z-index:1 for first object, z-index:2 for second.
	if !strings.Contains(out, "z-index:1;") || !strings.Contains(out, "z-index:2;") {
		t.Errorf("expected z-index:1 and z-index:2, got: %s", out)
	}
}

func TestExporter_LayersMode_AbsolutePageCoords(t *testing.T) {
	pp := buildLayeredPages()
	exp := html.NewExporter()
	exp.Layers = true
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export (layers): %v", err)
	}
	out := buf.String()

	// Obj1 is at band.Top(0) + obj.Top(5) = 5px from page top.
	if !strings.Contains(out, "top:5.00px") {
		t.Errorf("Obj1 should be at top:5.00px (page-absolute), got: %s", out)
	}
	// Obj2 is at band.Top(30) + obj.Top(5) = 35px from page top.
	if !strings.Contains(out, "top:35.00px") {
		t.Errorf("Obj2 should be at top:35.00px (page-absolute), got: %s", out)
	}
}

func TestExporter_LayersMode_BothObjectsPresent(t *testing.T) {
	pp := buildLayeredPages()
	exp := html.NewExporter()
	exp.Layers = true
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export (layers): %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "Layer1") || !strings.Contains(out, "Layer2") {
		t.Error("both layer objects should appear in output")
	}
}

func TestExporter_NonLayered_FlatLayout(t *testing.T) {
	// New behavior: flat layout — no band wrapper divs, objects at page-absolute coords.
	pp := buildLayeredPages()
	exp := html.NewExporter()
	// Layers=false by default.
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()

	// Flat mode should NOT emit band wrapper divs.
	if strings.Contains(out, `class="band"`) {
		t.Error("flat mode should not emit band wrapper divs")
	}
	// No z-index in flat mode.
	if strings.Contains(out, "z-index:") {
		t.Error("flat mode should not have z-index")
	}
	// Objects should still appear.
	if !strings.Contains(out, "Layer1") || !strings.Contains(out, "Layer2") {
		t.Error("both objects should appear in flat output")
	}
}

func TestExporter_Scale(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	exp := html.NewExporter()
	exp.Scale = 0.5
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	// At 0.5 scale, the 794px page becomes 794*0.5 + 3*0.5 = 397 + 1.5 = 398.50px.
	if !strings.Contains(buf.String(), "398.50px") {
		t.Error("scaled output should contain 398.50px page width")
	}
}

// ── Watermark ──────────────────────────────────────────────────────────────────

func buildPagesWithWatermark(text string, onTop bool) *preview.PreparedPages {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "DataBand", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:       true,
		Text:          text,
		TextColor:     color.RGBA{R: 128, G: 0, B: 0, A: 128},
		TextRotation:  preview.WatermarkTextRotationForwardDiagonal,
		ShowTextOnTop: onTop,
		ImageBlobIdx:  -1,
	}
	return pp
}

func TestExporter_WatermarkText_Behind(t *testing.T) {
	pp := buildPagesWithWatermark("CONFIDENTIAL", false)
	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "CONFIDENTIAL") {
		t.Error("watermark text 'CONFIDENTIAL' should appear in HTML output")
	}
	if !strings.Contains(out, "rotate(") {
		t.Error("watermark should include CSS rotation transform")
	}
}

func TestExporter_WatermarkText_OnTop(t *testing.T) {
	pp := buildPagesWithWatermark("DRAFT", true)
	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "DRAFT") {
		t.Error("on-top watermark text 'DRAFT' should appear in HTML output")
	}
}

func TestExporter_WatermarkText_Horizontal(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 40})
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "HORIZONTAL",
		TextRotation: preview.WatermarkTextRotationHorizontal,
		ImageBlobIdx: -1,
	}
	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "HORIZONTAL") {
		t.Error("watermark text should appear")
	}
	// Horizontal has no CSS rotation transform.
	if strings.Contains(out, "rotate(") {
		t.Error("horizontal watermark should not have rotate transform")
	}
}
