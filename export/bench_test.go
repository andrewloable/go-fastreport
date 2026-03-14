package export_test

import (
	"io"
	"testing"

	"github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/export/pdf"
	imgexport "github.com/andrewloable/go-fastreport/export/image"
	"github.com/andrewloable/go-fastreport/preview"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func buildBenchPages(pages, bandsPerPage int) *preview.PreparedPages {
	pp := preview.New()
	for p := 0; p < pages; p++ {
		pp.AddPage(794, 1123, p+1)
		for b := 0; b < bandsPerPage; b++ {
			_ = pp.AddBand(&preview.PreparedBand{
				Name:   "DataBand",
				Top:    float32(b * 40),
				Height: 40,
			})
		}
	}
	return pp
}

// ── HTML export benchmarks ─────────────────────────────────────────────────────

func BenchmarkHTMLExport_1Page_10Bands(b *testing.B) {
	pp := buildBenchPages(1, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exp := html.NewExporter()
		if err := exp.Export(pp, io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTMLExport_10Pages_10Bands(b *testing.B) {
	pp := buildBenchPages(10, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exp := html.NewExporter()
		if err := exp.Export(pp, io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTMLExport_100Pages_20Bands(b *testing.B) {
	pp := buildBenchPages(100, 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exp := html.NewExporter()
		if err := exp.Export(pp, io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

// ── PDF export benchmarks ─────────────────────────────────────────────────────

func BenchmarkPDFExport_1Page_10Bands(b *testing.B) {
	pp := buildBenchPages(1, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exp := pdf.NewExporter()
		if err := exp.Export(pp, io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPDFExport_10Pages_10Bands(b *testing.B) {
	pp := buildBenchPages(10, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exp := pdf.NewExporter()
		if err := exp.Export(pp, io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

// ── Image export benchmarks ───────────────────────────────────────────────────

func BenchmarkImageExport_1Page_10Bands(b *testing.B) {
	pp := buildBenchPages(1, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exp := imgexport.NewExporter()
		if err := exp.Export(pp, io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImageExport_5Pages_10Bands(b *testing.B) {
	pp := buildBenchPages(5, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exp := imgexport.NewExporter()
		if err := exp.Export(pp, io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}
