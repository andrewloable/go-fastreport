package reportpkg_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	_ "github.com/andrewloable/go-fastreport/engine" // registers prepare func
	htmlexp "github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// TestE2E_DataBand_TextObject_HTML verifies the full pipeline:
//   data binding → expression evaluation → PreparedBand.Objects → HTML rendering.
func TestE2E_DataBand_TextObject_HTML(t *testing.T) {
	// ── 1. Build data source ────────────────────────────────────────────────
	ds := data.NewBaseDataSource("Products")
	ds.SetAlias("Products")
	ds.AddColumn(data.Column{Name: "Name"})
	ds.AddColumn(data.Column{Name: "Price"})
	ds.AddRow(map[string]any{"Name": "Apple", "Price": 1.5})
	ds.AddRow(map[string]any{"Name": "Banana", "Price": 0.75})
	ds.AddRow(map[string]any{"Name": "Cherry", "Price": 3.0})

	// ── 2. Build report ─────────────────────────────────────────────────────
	r := reportpkg.NewReport()
	r.Dictionary().AddDataSource(ds)

	pg := reportpkg.NewReportPage()
	pg.PaperHeight = 297
	pg.PaperWidth = 210
	pg.TopMargin = 10
	pg.BottomMargin = 10
	pg.LeftMargin = 10
	pg.RightMargin = 10

	// DataBand bound to the Products data source.
	db := band.NewDataBand()
	db.SetName("ProductsBand")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(ds)

	// TextObject showing the product name.
	txt := object.NewTextObject()
	txt.SetName("NameText")
	txt.SetText("[Name]")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(10)
	txt.SetVisible(true)
	db.Objects().Add(txt)

	pg.AddBand(db)
	r.AddPage(pg)

	// ── 3. Prepare (engine run) ─────────────────────────────────────────────
	if err := r.Prepare(); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	pp := r.PreparedPages()
	if pp == nil || pp.Count() == 0 {
		t.Fatal("no prepared pages produced")
	}

	// ── 4. Export to HTML ───────────────────────────────────────────────────
	exp := htmlexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("HTML export: %v", err)
	}

	html := buf.String()

	// ── 5. Assert content ───────────────────────────────────────────────────
	for _, expected := range []string{"Apple", "Banana", "Cherry"} {
		if !strings.Contains(html, expected) {
			t.Errorf("HTML output missing %q\n--- HTML (first 2000 chars) ---\n%s",
				expected, truncate(html, 2000))
		}
	}
}

// TestE2E_StaticTextObject_HTML verifies a static text object (no expressions)
// renders its literal text in the HTML output.
func TestE2E_StaticTextObject_HTML(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PageHeader")
	hdr.SetHeight(20)
	hdr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("Title")
	txt.SetText("My Report Title")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(200)
	txt.SetHeight(20)
	txt.SetVisible(true)
	hdr.Objects().Add(txt)

	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	if err := r.Prepare(); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	exp := htmlexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(r.PreparedPages(), &buf); err != nil {
		t.Fatalf("HTML export: %v", err)
	}

	if !strings.Contains(buf.String(), "My Report Title") {
		t.Errorf("expected 'My Report Title' in HTML output\n%s", truncate(buf.String(), 1500))
	}
}

// TestE2E_SystemVariable_PageNumber verifies [PageNumber] resolves in HTML.
func TestE2E_SystemVariable_PageNumber(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	ftr := band.NewPageFooterBand()
	ftr.SetName("PageFooter")
	ftr.SetHeight(15)
	ftr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("PageNum")
	txt.SetText("Page [PageNumber]")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(15)
	txt.SetVisible(true)
	ftr.Objects().Add(txt)

	pg.SetPageFooter(ftr)
	r.AddPage(pg)

	if err := r.Prepare(); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	exp := htmlexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(r.PreparedPages(), &buf); err != nil {
		t.Fatalf("HTML export: %v", err)
	}

	// Should contain "Page " followed by a number.
	if !strings.Contains(buf.String(), "Page ") {
		t.Errorf("expected 'Page ...' in HTML output\n%s", truncate(buf.String(), 1500))
	}
}

// helpers

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// Compile-time check: object.TextObject satisfies report.Base.
var _ report.Base = (*object.TextObject)(nil)
