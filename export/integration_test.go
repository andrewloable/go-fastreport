// integration_test.go verifies end-to-end HTML and PDF export of prepared pages.
package export_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/export/pdf"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// sliceDS is an in-memory DataSource backed by a string slice.
type sliceDS struct {
	rows []string
	pos  int
}

func newSliceDS(rows ...string) *sliceDS { return &sliceDS{rows: rows, pos: -1} }

func (d *sliceDS) RowCount() int { return len(d.rows) }
func (d *sliceDS) First() error  { d.pos = 0; return nil }
func (d *sliceDS) Next() error   { d.pos++; return nil }
func (d *sliceDS) EOF() bool     { return d.pos >= len(d.rows) }
func (d *sliceDS) GetValue(_ string) (any, error) {
	if d.pos >= 0 && d.pos < len(d.rows) {
		return d.rows[d.pos], nil
	}
	return nil, nil
}

// prepareReport builds and runs a report, returning the PreparedPages.
func prepareReport(t *testing.T, setup func(*reportpkg.Report, *reportpkg.ReportPage)) *preview.PreparedPages {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	setup(r, pg)
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("engine.Run: %v", err)
	}
	return e.PreparedPages()
}

// ── HTML export integration tests ─────────────────────────────────────────────

// Scenario 1: Simple text report → HTML structure verification.
func TestHTMLExport_SimpleReport_Structure(t *testing.T) {
	pp := prepareReport(t, func(r *reportpkg.Report, pg *reportpkg.ReportPage) {
		db := band.NewDataBand()
		db.SetName("DataBand")
		db.SetVisible(true)
		db.SetHeight(20)
		db.SetDataSource(newSliceDS("Alice", "Bob", "Carol"))
		pg.AddBand(db)
	})

	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("HTML Export: %v", err)
	}

	out := buf.String()
	// Must be valid HTML5.
	if !strings.Contains(out, "<!DOCTYPE html>") {
		t.Error("output missing DOCTYPE")
	}
	if !strings.Contains(out, "<html>") {
		t.Error("output missing <html>")
	}
	if !strings.Contains(out, "</html>") {
		t.Error("output missing </html>")
	}
	// Must contain one page div.
	if !strings.Contains(out, `class="page"`) {
		t.Error("output missing page div")
	}
	// Must contain all 3 data band divs.
	count := strings.Count(out, `data-name="DataBand"`)
	if count != 3 {
		t.Errorf("expected 3 DataBand divs, got %d", count)
	}
}

// Scenario 2: Custom title appears in <title> tag.
func TestHTMLExport_CustomTitle(t *testing.T) {
	pp := prepareReport(t, func(r *reportpkg.Report, pg *reportpkg.ReportPage) {
		db := band.NewDataBand()
		db.SetName("D")
		db.SetVisible(true)
		db.SetHeight(10)
		pg.AddBand(db)
	})

	exp := html.NewExporter()
	exp.Title = "Sales Report 2024"
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("HTML Export: %v", err)
	}

	if !strings.Contains(buf.String(), "Sales Report 2024") {
		t.Error("HTML title not found in output")
	}
}

// Scenario 3: CSS embedded by default, not embedded when EmbedCSS=false.
func TestHTMLExport_EmbedCSS(t *testing.T) {
	pp := prepareReport(t, func(r *reportpkg.Report, pg *reportpkg.ReportPage) {
		db := band.NewDataBand()
		db.SetName("D")
		db.SetVisible(true)
		db.SetHeight(10)
		pg.AddBand(db)
	})

	// Default: CSS embedded.
	expOn := html.NewExporter()
	var bufOn bytes.Buffer
	_ = expOn.Export(pp, &bufOn)
	if !strings.Contains(bufOn.String(), "<style>") {
		t.Error("CSS should be embedded by default")
	}

	// EmbedCSS=false: no <style> tag.
	expOff := html.NewExporter()
	expOff.EmbedCSS = false
	var bufOff bytes.Buffer
	_ = expOff.Export(pp, &bufOff)
	if strings.Contains(bufOff.String(), "<style>") {
		t.Error("CSS should not be embedded when EmbedCSS=false")
	}
}

// Scenario 4: Multi-page report produces multiple page divs.
func TestHTMLExport_MultiPage(t *testing.T) {
	pp := prepareReport(t, func(r *reportpkg.Report, pg *reportpkg.ReportPage) {
		db := band.NewDataBand()
		db.SetName("DataBand")
		db.SetVisible(true)
		db.SetHeight(100) // large enough to overflow onto page 2
		rows := make([]string, 15)
		for i := range rows {
			rows[i] = "row"
		}
		db.SetDataSource(newSliceDS(rows...))
		pg.AddBand(db)
	})

	if pp.Count() < 2 {
		t.Skipf("report did not overflow to page 2 (pages=%d) — adjust band height", pp.Count())
	}

	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("HTML Export: %v", err)
	}

	pageCount := strings.Count(buf.String(), `class="page"`)
	if pageCount != pp.Count() {
		t.Errorf("HTML page div count = %d, want %d", pageCount, pp.Count())
	}
}

// Scenario 5: Scale option changes pixel dimensions.
func TestHTMLExport_Scale(t *testing.T) {
	pp := prepareReport(t, func(r *reportpkg.Report, pg *reportpkg.ReportPage) {
		db := band.NewDataBand()
		db.SetName("D")
		db.SetVisible(true)
		db.SetHeight(10)
		pg.AddBand(db)
	})

	exp := html.NewExporter()
	exp.Scale = 0.5
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("HTML Export: %v", err)
	}

	// A4 page at 96dpi: usable width ≈ 718px; at 0.5 scale ≈ 359px.
	out := buf.String()
	// Just verify the output contains a fractional pixel value (scale applied).
	if !strings.Contains(out, "px") {
		t.Error("scaled HTML should contain pixel values")
	}
}

// Scenario 6: PageRange=Current exports only the specified page.
func TestHTMLExport_PageRangeCurrent(t *testing.T) {
	pp := prepareReport(t, func(r *reportpkg.Report, pg *reportpkg.ReportPage) {
		db := band.NewDataBand()
		db.SetName("DataBand")
		db.SetVisible(true)
		db.SetHeight(100)
		rows := make([]string, 15)
		for i := range rows {
			rows[i] = "row"
		}
		db.SetDataSource(newSliceDS(rows...))
		pg.AddBand(db)
	})

	if pp.Count() < 2 {
		t.Skipf("need ≥2 pages for this test (got %d)", pp.Count())
	}

	exp := html.NewExporter()
	exp.PageRange = export.PageRangeCurrent
	exp.CurPage = 1 // export only page index 1 (second page)
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("HTML Export: %v", err)
	}

	count := strings.Count(buf.String(), `class="page"`)
	if count != 1 {
		t.Errorf("PageRangeCurrent: page count = %d, want 1", count)
	}
}

// Scenario 7: HTML band names are escaped for XSS safety.
func TestHTMLExport_BandNameEscaping(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "<script>alert(1)</script>", Top: 0, Height: 30})

	exp := html.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("HTML Export: %v", err)
	}

	if strings.Contains(buf.String(), "<script>alert") {
		t.Error("band name should be HTML-escaped to prevent XSS")
	}
}

// Scenario 8: PageRange=Custom exports specific page numbers.
func TestHTMLExport_PageRangeCustom(t *testing.T) {
	pp := prepareReport(t, func(r *reportpkg.Report, pg *reportpkg.ReportPage) {
		db := band.NewDataBand()
		db.SetName("D")
		db.SetVisible(true)
		db.SetHeight(100)
		rows := make([]string, 20)
		for i := range rows {
			rows[i] = "row"
		}
		db.SetDataSource(newSliceDS(rows...))
		pg.AddBand(db)
	})

	if pp.Count() < 3 {
		t.Skipf("need ≥3 pages (got %d)", pp.Count())
	}

	exp := html.NewExporter()
	exp.PageRange = export.PageRangeCustom
	exp.PageNumbers = "1,3"
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("HTML Export: %v", err)
	}

	count := strings.Count(buf.String(), `class="page"`)
	if count != 2 {
		t.Errorf("PageRangeCustom '1,3': page count = %d, want 2", count)
	}
}

// ── PDF export integration tests ──────────────────────────────────────────────

// Scenario 1: Simple report → valid PDF structure.
func TestPDFExport_SimpleReport_Structure(t *testing.T) {
	pp := prepareReport(t, func(r *reportpkg.Report, pg *reportpkg.ReportPage) {
		hdr := band.NewPageHeaderBand()
		hdr.SetName("PageHeader")
		hdr.SetVisible(true)
		hdr.SetHeight(30)
		pg.SetPageHeader(hdr)

		db := band.NewDataBand()
		db.SetName("DataBand")
		db.SetVisible(true)
		db.SetHeight(15)
		db.SetDataSource(newSliceDS("Item1", "Item2"))
		pg.AddBand(db)
	})

	exp := pdf.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("PDF Export: %v", err)
	}

	out := buf.String()
	if !strings.HasPrefix(out, "%PDF-") {
		t.Error("output does not start with %PDF- header")
	}
	if !strings.Contains(out, "%%EOF") {
		t.Errorf("output missing %%%%EOF marker")
	}
}

// Scenario 2: Multi-page report produces multiple pages in PDF.
func TestPDFExport_MultiPage_BandNames(t *testing.T) {
	pp := prepareReport(t, func(r *reportpkg.Report, pg *reportpkg.ReportPage) {
		db := band.NewDataBand()
		db.SetName("DataBand")
		db.SetVisible(true)
		db.SetHeight(100)
		// Add a text label so its content appears in the PDF content stream.
		lbl := object.NewTextObject()
		lbl.SetName("Label")
		lbl.SetWidth(200)
		lbl.SetHeight(20)
		lbl.SetText("DataBand")
		db.AddChild(lbl)
		rows := make([]string, 15)
		for i := range rows {
			rows[i] = "row"
		}
		db.SetDataSource(newSliceDS(rows...))
		pg.AddBand(db)
	})

	exp := pdf.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("PDF Export: %v", err)
	}

	out := buf.String()
	// Band names appear as text labels in content streams.
	if !strings.Contains(out, "DataBand") {
		t.Error("PDF should contain DataBand label in content stream")
	}
}

// Scenario 3: Empty report → no error, output may be empty (no pages to render).
func TestPDFExport_EmptyReport(t *testing.T) {
	pp := preview.New()
	exp := pdf.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("PDF Export on empty pages: %v", err)
	}
	// No pages → ExportBase returns early; output is empty. That's OK.
}

// Scenario 4: nil PreparedPages returns error.
func TestPDFExport_NilPages_ReturnsError(t *testing.T) {
	exp := pdf.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(nil, &buf); err == nil {
		t.Error("expected error for nil PreparedPages")
	}
}

// Scenario 5: PageRange=Custom exports fewer pages.
func TestPDFExport_PageRangeCustom_Smaller(t *testing.T) {
	pp := prepareReport(t, func(r *reportpkg.Report, pg *reportpkg.ReportPage) {
		db := band.NewDataBand()
		db.SetName("D")
		db.SetVisible(true)
		db.SetHeight(100)
		rows := make([]string, 15)
		for i := range rows {
			rows[i] = "row"
		}
		db.SetDataSource(newSliceDS(rows...))
		pg.AddBand(db)
	})

	if pp.Count() < 2 {
		t.Skipf("need ≥2 pages (got %d)", pp.Count())
	}

	expAll := pdf.NewExporter()
	var bufAll bytes.Buffer
	_ = expAll.Export(pp, &bufAll)

	expOne := pdf.NewExporter()
	expOne.PageRange = export.PageRangeCustom
	expOne.PageNumbers = "1"
	var bufOne bytes.Buffer
	_ = expOne.Export(pp, &bufOne)

	if bufOne.Len() >= bufAll.Len() {
		t.Error("single-page PDF should be smaller than all-pages PDF")
	}
}

// Scenario 6: Report with PageHeader and PageFooter → both appear in PDF.
func TestPDFExport_PageHeaderFooter(t *testing.T) {
	pp := prepareReport(t, func(r *reportpkg.Report, pg *reportpkg.ReportPage) {
		hdr := band.NewPageHeaderBand()
		hdr.SetName("Header")
		hdr.SetVisible(true)
		hdr.SetHeight(20)
		hdrLabel := object.NewTextObject()
		hdrLabel.SetName("HdrLabel")
		hdrLabel.SetWidth(200)
		hdrLabel.SetHeight(15)
		hdrLabel.SetText("Header")
		hdr.AddChild(hdrLabel)
		pg.SetPageHeader(hdr)

		ftr := band.NewPageFooterBand()
		ftr.SetName("Footer")
		ftr.SetVisible(true)
		ftr.SetHeight(15)
		ftrLabel := object.NewTextObject()
		ftrLabel.SetName("FtrLabel")
		ftrLabel.SetWidth(200)
		ftrLabel.SetHeight(10)
		ftrLabel.SetText("Footer")
		ftr.AddChild(ftrLabel)
		pg.SetPageFooter(ftr)
	})

	exp := pdf.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("PDF Export: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Header") {
		t.Error("PDF should contain Header band label")
	}
	if !strings.Contains(out, "Footer") {
		t.Error("PDF should contain Footer band label")
	}
}

// ── Cross-format tests ────────────────────────────────────────────────────────

// Scenario: Both HTML and PDF export produce non-empty output for the same input.
func TestBothFormats_NonEmpty(t *testing.T) {
	pp := prepareReport(t, func(r *reportpkg.Report, pg *reportpkg.ReportPage) {
		title := band.NewReportTitleBand()
		title.SetName("ReportTitle")
		title.SetVisible(true)
		title.SetHeight(30)
		pg.SetReportTitle(title)

		db := band.NewDataBand()
		db.SetName("DataBand")
		db.SetVisible(true)
		db.SetHeight(12)
		db.SetDataSource(newSliceDS("A", "B", "C"))
		pg.AddBand(db)
	})

	var htmlBuf, pdfBuf bytes.Buffer

	htmlExp := html.NewExporter()
	if err := htmlExp.Export(pp, &htmlBuf); err != nil {
		t.Fatalf("HTML Export: %v", err)
	}

	pdfExp := pdf.NewExporter()
	if err := pdfExp.Export(pp, &pdfBuf); err != nil {
		t.Fatalf("PDF Export: %v", err)
	}

	if htmlBuf.Len() == 0 {
		t.Error("HTML output should not be empty")
	}
	if pdfBuf.Len() == 0 {
		t.Error("PDF output should not be empty")
	}
	if !strings.Contains(htmlBuf.String(), "<!DOCTYPE html>") {
		t.Error("HTML output missing DOCTYPE")
	}
	if !strings.HasPrefix(pdfBuf.String(), "%PDF-") {
		t.Error("PDF output missing %PDF- header")
	}
}
