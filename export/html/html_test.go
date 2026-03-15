package html_test

import (
	"bytes"
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
	if !strings.Contains(out, "<!DOCTYPE html>") {
		t.Error("output should contain DOCTYPE")
	}
	if !strings.Contains(out, "Header") {
		t.Error("output should contain band name 'Header'")
	}
	if !strings.Contains(out, "DataBand") {
		t.Error("output should contain band name 'DataBand'")
	}
	if !strings.Contains(out, `class="page"`) {
		t.Error("output should contain page div")
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

	if strings.Contains(buf.String(), "<style>") {
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

	count := strings.Count(buf.String(), `class="page"`)
	if count != 3 {
		t.Errorf("expected 3 page divs, got %d", count)
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

	count := strings.Count(buf.String(), `class="page"`)
	if count != 1 {
		t.Errorf("PageRangeCurrent: expected 1 page, got %d", count)
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

	if strings.Contains(buf.String(), "<script>") {
		t.Error("HTML should escape < and > in band names")
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

func TestExporter_NonLayered_HasBandDivs(t *testing.T) {
	pp := buildLayeredPages()
	exp := html.NewExporter()
	// Layers=false by default.
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, `class="band"`) {
		t.Error("non-layered mode should emit band wrapper divs")
	}
	if strings.Contains(out, "z-index:") {
		t.Error("non-layered mode should not have z-index")
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

	// At 0.5 scale, the 794px page becomes 397px.
	if !strings.Contains(buf.String(), "397.00px") {
		t.Error("scaled output should contain 397px page width")
	}
}
