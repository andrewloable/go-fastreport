package pdf_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/export/pdf"
	"github.com/andrewloable/go-fastreport/preview"
)

// buildTestPages creates a PreparedPages with nPages pages, each containing
// the given band names. Each band has a single text PreparedObject whose
// Text is the band name, so it appears in the PDF content stream.
func buildTestPages(nPages int, bandNames []string) *preview.PreparedPages {
	pp := preview.New()
	for i := 0; i < nPages; i++ {
		pp.AddPage(595, 842, i+1)
		for j, name := range bandNames {
			b := &preview.PreparedBand{
				Name:   name,
				Top:    float32(j * 30),
				Height: 30,
			}
			b.Objects = []preview.PreparedObject{
				{
					Name:   name + "Obj",
					Kind:   preview.ObjectTypeText,
					Left:   0,
					Top:    0,
					Width:  200,
					Height: 20,
					Text:   name,
				},
			}
			_ = pp.AddBand(b)
		}
	}
	return pp
}

func TestExporter_BasicPDF(t *testing.T) {
	pp := buildTestPages(1, []string{"PageHeader", "DataBand"})
	exp := pdf.NewExporter()
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	out := buf.String()
	if !strings.HasPrefix(out, "%PDF-") {
		prefix := out
		if len(prefix) > 20 {
			prefix = prefix[:20]
		}
		t.Errorf("output does not start with %%PDF-: %q", prefix)
	}
	if !strings.Contains(out, "%EOF") {
		t.Error("output does not contain EOF marker")
	}
}

func TestExporter_EmptyPages(t *testing.T) {
	pp := preview.New()
	exp := pdf.NewExporter()
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export on empty pages: %v", err)
	}
	// No pages → empty or minimal PDF output (just header+EOF).
}

func TestExporter_MultiPage(t *testing.T) {
	pp := buildTestPages(3, []string{"Band1", "Band2"})
	exp := pdf.NewExporter()
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	out := buf.String()
	// Should contain all band names in content streams.
	if !strings.Contains(out, "Band1") {
		t.Error("PDF should contain Band1")
	}
}

func TestExporter_PageRangeCustom(t *testing.T) {
	pp := buildTestPages(5, []string{"Band"})
	exp := pdf.NewExporter()
	exp.PageRange = export.PageRangeCustom
	exp.PageNumbers = "2-3"
	var buf bytes.Buffer

	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	out := buf.String()
	// With only 2 pages exported, the PDF should be smaller.
	if len(out) == 0 {
		t.Error("PDF output should not be empty")
	}
}

func TestExporter_NilPages(t *testing.T) {
	exp := pdf.NewExporter()
	var buf bytes.Buffer
	err := exp.Export(nil, &buf)
	if err == nil {
		t.Error("expected error for nil PreparedPages")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
