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
