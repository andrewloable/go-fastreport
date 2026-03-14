// integration_test.go verifies end-to-end report preparation scenarios.
// Each test builds a complete Report object, binds data, runs the engine,
// and asserts the resulting PreparedPages structure.
package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// sliceDS implements band.DataSource backed by a slice of strings.
// It is used throughout the integration tests for simple data iteration.
type sliceDS struct {
	rows []string
	pos  int
}

func newSliceDS(rows ...string) *sliceDS { return &sliceDS{rows: rows, pos: -1} }

func (d *sliceDS) RowCount() int { return len(d.rows) }
func (d *sliceDS) First() error  { d.pos = 0; return nil }
func (d *sliceDS) Next() error   { d.pos++; return nil }
func (d *sliceDS) EOF() bool     { return d.pos >= len(d.rows) }
func (d *sliceDS) GetValue(col string) (any, error) {
	if d.pos < 0 || d.pos >= len(d.rows) {
		return nil, nil
	}
	return d.rows[d.pos], nil
}

// preparedBandCount returns the total number of PreparedBands across all pages.
func preparedBandCount(t *testing.T, pp interface {
	Count() int
	GetPage(int) interface{ Bands() interface{ Len() int } }
}) int {
	t.Helper()
	return 0 // not used directly; see inline calls below
}

// ── Scenario 1: Simple list report ───────────────────────────────────────────

// TestIntegration_SimpleListReport verifies that a DataBand with 5 rows
// produces 5 PreparedBands on a single prepared page.
func TestIntegration_SimpleListReport(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(20)
	db.SetDataSource(newSliceDS("R1", "R2", "R3", "R4", "R5"))
	pg.AddBand(db)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	if pp.Count() == 0 {
		t.Fatal("expected at least 1 prepared page")
	}

	total := 0
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p != nil {
			total += len(p.Bands)
		}
	}
	if total != 5 {
		t.Errorf("total PreparedBands = %d, want 5", total)
	}
}

// ── Scenario 2: Report with PageHeader and PageFooter ─────────────────────────

// TestIntegration_PageHeaderAndFooter verifies that PageHeader and PageFooter
// bands appear on the prepared page alongside data bands.
func TestIntegration_PageHeaderAndFooter(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PageHeader")
	hdr.SetVisible(true)
	hdr.SetHeight(30)
	pg.SetPageHeader(hdr)

	ftr := band.NewPageFooterBand()
	ftr.SetName("PageFooter")
	ftr.SetVisible(true)
	ftr.SetHeight(20)
	pg.SetPageFooter(ftr)

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(15)
	db.SetDataSource(newSliceDS("A", "B", "C"))
	pg.AddBand(db)

	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	if pp.Count() == 0 {
		t.Fatal("expected at least 1 prepared page")
	}

	// Find PageHeader and PageFooter in band names.
	found := map[string]bool{}
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p == nil {
			continue
		}
		for _, b := range p.Bands {
			found[b.Name] = true
		}
	}

	if !found["PageHeader"] {
		t.Error("PageHeader should appear in PreparedBands")
	}
	if !found["PageFooter"] {
		t.Error("PageFooter should appear in PreparedBands")
	}
}

// ── Scenario 3: Report with ReportTitle ──────────────────────────────────────

// TestIntegration_ReportTitle verifies that the ReportTitle band appears exactly
// once (only on the first page).
func TestIntegration_ReportTitle(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	title := band.NewReportTitleBand()
	title.SetName("ReportTitle")
	title.SetVisible(true)
	title.SetHeight(40)
	pg.SetReportTitle(title)

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newSliceDS("X"))
	pg.AddBand(db)

	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	titleCount := 0
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p == nil {
			continue
		}
		for _, b := range p.Bands {
			if b.Name == "ReportTitle" {
				titleCount++
			}
		}
	}
	if titleCount != 1 {
		t.Errorf("ReportTitle count = %d, want 1", titleCount)
	}
}

// ── Scenario 4: Grouped report ───────────────────────────────────────────────

// TestIntegration_GroupedReport verifies that GroupHeader and GroupFooter bands
// are rendered once per group, with data rows inside each group.
func TestIntegration_GroupedReport(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	gh := band.NewGroupHeaderBand()
	gh.SetName("GroupHeader")
	gh.SetVisible(true)
	gh.SetHeight(15)
	gh.SetCondition("val") // column name in sliceDS

	gf := band.NewGroupFooterBand()
	gf.SetName("GroupFooter")
	gf.SetVisible(true)
	gf.SetHeight(10)
	gh.SetGroupFooter(gf)

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(12)
	// 2 groups: "A" (2 rows), "B" (1 row)
	db.SetDataSource(newSliceDS("A", "A", "B"))
	gh.SetData(db)

	pg.AddBand(gh)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	names := map[string]int{}
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p == nil {
			continue
		}
		for _, b := range p.Bands {
			names[b.Name]++
		}
	}

	if names["GroupHeader"] != 2 {
		t.Errorf("GroupHeader count = %d, want 2", names["GroupHeader"])
	}
	if names["GroupFooter"] != 2 {
		t.Errorf("GroupFooter count = %d, want 2", names["GroupFooter"])
	}
	if names["DataBand"] != 3 {
		t.Errorf("DataBand count = %d, want 3", names["DataBand"])
	}
}

// ── Scenario 5: Multi-page overflow ──────────────────────────────────────────

// TestIntegration_MultiPageOverflow verifies that when data rows exceed the
// page height, the engine starts a new page automatically.
func TestIntegration_MultiPageOverflow(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	// A4 page at 96dpi: usable height = (297-20) * 96/25.4 ≈ 1047 px
	// Use 100px bands; 11 bands = ~1100px → overflow onto page 2

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

	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	if pp.Count() < 2 {
		t.Errorf("expected at least 2 pages for overflow, got %d", pp.Count())
	}
}

// ── Scenario 6: Double-pass report ───────────────────────────────────────────

// TestIntegration_DoublePass verifies that DoublePass=true causes the engine
// to run two passes and FinalPass to be true after completion.
func TestIntegration_DoublePass(t *testing.T) {
	r := reportpkg.NewReport()
	r.DoublePass = true

	pg := reportpkg.NewReportPage()
	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(20)
	db.SetDataSource(newSliceDS("R1", "R2"))
	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !e.FinalPass() {
		t.Error("FinalPass should be true after double-pass run")
	}
	if e.FirstPass() {
		t.Error("FirstPass should be false after final pass")
	}
}

// ── Scenario 7: DataHeader and DataFooter ────────────────────────────────────

// TestIntegration_DataHeaderFooter verifies that DataHeader appears before the
// first data row and DataFooter appears after the last.
func TestIntegration_DataHeaderFooter(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewDataHeaderBand()
	hdr.SetName("DataHeader")
	hdr.SetVisible(true)
	hdr.SetHeight(15)

	ftr := band.NewDataFooterBand()
	ftr.SetName("DataFooter")
	ftr.SetVisible(true)
	ftr.SetHeight(15)

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(12)
	db.SetDataSource(newSliceDS("A", "B"))
	db.SetHeader(hdr)
	db.SetFooter(ftr)
	pg.AddBand(db)

	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	names := []string{}
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p == nil {
			continue
		}
		for _, b := range p.Bands {
			names = append(names, b.Name)
		}
	}

	// Expected order: DataHeader, DataBand, DataBand, DataFooter
	if len(names) < 4 {
		t.Fatalf("expected at least 4 bands, got %d: %v", len(names), names)
	}
	if names[0] != "DataHeader" {
		t.Errorf("first band = %q, want DataHeader", names[0])
	}
	if names[len(names)-1] != "DataFooter" {
		t.Errorf("last band = %q, want DataFooter", names[len(names)-1])
	}
}

// ── Scenario 8: Keep-together ─────────────────────────────────────────────────

// TestIntegration_KeepTogether verifies that StartKeep/EndKeep can be used
// without panicking and that the keeping state is managed correctly.
func TestIntegration_KeepTogether(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(20)
	db.SetDataSource(newSliceDS("X", "Y"))
	pg.AddBand(db)

	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// StartKeep / EndKeep lifecycle works after a run.
	e.StartKeep()
	if !e.IsKeeping() {
		t.Error("IsKeeping should be true after StartKeep")
	}
	e.EndKeep()
	if e.IsKeeping() {
		t.Error("IsKeeping should be false after EndKeep")
	}
}

// ── Scenario 9: Empty data source with PrintIfDSEmpty ────────────────────────

// TestIntegration_PrintIfDSEmpty verifies that a DataBand with an empty data
// source and PrintIfDSEmpty=true still renders one row.
func TestIntegration_PrintIfDSEmpty(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(20)
	db.SetDataSource(newSliceDS()) // 0 rows
	db.SetPrintIfDSEmpty(true)
	pg.AddBand(db)

	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	total := 0
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p != nil {
			for _, b := range p.Bands {
				if b.Name == "DataBand" {
					total++
				}
			}
		}
	}
	if total != 1 {
		t.Errorf("PrintIfDSEmpty: DataBand count = %d, want 1", total)
	}
}

// ── Scenario 10: Report summary band ─────────────────────────────────────────

// TestIntegration_ReportSummary verifies that the ReportSummary band appears
// after all data bands.
func TestIntegration_ReportSummary(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	summary := band.NewReportSummaryBand()
	summary.SetName("ReportSummary")
	summary.SetVisible(true)
	summary.SetHeight(25)
	pg.SetReportSummary(summary)

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newSliceDS("A"))
	pg.AddBand(db)

	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	names := []string{}
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p != nil {
			for _, b := range p.Bands {
				names = append(names, b.Name)
			}
		}
	}

	// ReportSummary should be the last band.
	if len(names) == 0 {
		t.Fatal("no bands rendered")
	}
	if names[len(names)-1] != "ReportSummary" {
		t.Errorf("last band = %q, want ReportSummary", names[len(names)-1])
	}
}
