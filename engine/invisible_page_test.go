package engine_test

// Tests for invisible ReportPage and invisible DataBand behaviour.
// Invisible pages (Visible=false) should be skipped by the engine, matching
// C# FastReport behaviour where drill-down / detail pages are excluded from
// static report output.
//
// Drill-down groups (e.g. "Drill-Down Groups.frx") use a different pattern:
// the GroupHeaderBand is visible but its nested DataBand and GroupFooterBand
// are set Visible=false by default. The C# report script shows/hides them
// interactively. In static export, the group headers appear but data rows
// and footers are suppressed by the engine's existing Visible() checks.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// TestInvisiblePage_IsSkipped verifies that a ReportPage with Visible=false
// produces no prepared pages.
func TestInvisiblePage_IsSkipped(t *testing.T) {
	r := reportpkg.NewReport()

	// Visible page — should produce 1 prepared page.
	pg1 := reportpkg.NewReportPage()
	pg1.SetName("Page1")
	ph := band.NewPageHeaderBand()
	ph.SetName("Header1")
	ph.SetHeight(30)
	pg1.SetPageHeader(ph)
	r.AddPage(pg1)

	// Invisible page — should be skipped entirely.
	pg2 := reportpkg.NewReportPage()
	pg2.SetName("Page2")
	pg2.SetVisible(false)
	db := band.NewDataBand()
	db.SetName("DetailBand")
	db.SetHeight(20)
	pg2.AddBand(db)
	r.AddPage(pg2)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	// Only pg1 should produce output.
	if pp.Count() != 1 {
		t.Errorf("expected 1 prepared page (invisible page skipped), got %d", pp.Count())
	}

	// Verify no band from the invisible page appears in the output.
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p == nil {
			continue
		}
		for _, b := range p.Bands {
			if b.Name == "DetailBand" {
				t.Error("DetailBand from invisible page should not appear in output")
			}
		}
	}
}

// TestInvisiblePage_AllInvisible verifies that a report with only invisible pages
// produces zero prepared pages (engine runs without error).
func TestInvisiblePage_AllInvisible(t *testing.T) {
	r := reportpkg.NewReport()

	pg1 := reportpkg.NewReportPage()
	pg1.SetName("Page1")
	pg1.SetVisible(false)
	r.AddPage(pg1)

	pg2 := reportpkg.NewReportPage()
	pg2.SetName("Page2")
	pg2.SetVisible(false)
	r.AddPage(pg2)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if n := e.PreparedPages().Count(); n != 0 {
		t.Errorf("all invisible pages: expected 0 prepared pages, got %d", n)
	}
}

// TestInvisiblePage_VisibleDefault verifies that a newly created ReportPage is
// visible by default (Visible=true).
func TestInvisiblePage_VisibleDefault(t *testing.T) {
	pg := reportpkg.NewReportPage()
	if !pg.Visible() {
		t.Error("new ReportPage should be visible by default")
	}
}

// TestInvisiblePage_SetVisible verifies SetVisible/Visible round-trip.
func TestInvisiblePage_SetVisible(t *testing.T) {
	pg := reportpkg.NewReportPage()
	pg.SetVisible(false)
	if pg.Visible() {
		t.Error("Visible should be false after SetVisible(false)")
	}
	pg.SetVisible(true)
	if !pg.Visible() {
		t.Error("Visible should be true after SetVisible(true)")
	}
}

// TestInvisiblePage_MultiPage verifies that with 3 pages where the middle one
// is invisible, only 2 prepared pages are produced.
func TestInvisiblePage_MultiPage(t *testing.T) {
	r := reportpkg.NewReport()

	for i, vis := range []bool{true, false, true} {
		pg := reportpkg.NewReportPage()
		pg.SetName("Page" + string(rune('1'+i)))
		pg.SetVisible(vis)
		ph := band.NewPageHeaderBand()
		ph.SetName("Header" + string(rune('1'+i)))
		ph.SetHeight(20)
		pg.SetPageHeader(ph)
		r.AddPage(pg)
	}

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	if pp.Count() != 2 {
		t.Errorf("expected 2 prepared pages (middle invisible), got %d", pp.Count())
	}
}

// TestInvisibleDataBand_PageLevel verifies that a page-level DataBand with
// Visible=false produces no rows in the output.
func TestInvisibleDataBand_PageLevel(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	// Visible band.
	visDB := band.NewDataBand()
	visDB.SetName("VisibleData")
	visDB.SetHeight(15)
	visDB.SetDataSource(newSliceDS("A", "B"))
	pg.AddBand(visDB)

	// Invisible band.
	hidDB := band.NewDataBand()
	hidDB.SetName("HiddenData")
	hidDB.SetHeight(15)
	hidDB.SetVisible(false)
	hidDB.SetDataSource(newSliceDS("X", "Y", "Z"))
	pg.AddBand(hidDB)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p == nil {
			continue
		}
		for _, b := range p.Bands {
			if b.Name == "HiddenData" {
				t.Error("invisible DataBand 'HiddenData' should not appear in output")
			}
		}
	}

	// Visible band rows should still appear.
	visCount := 0
	for i := 0; i < pp.Count(); i++ {
		if p := pp.GetPage(i); p != nil {
			for _, b := range p.Bands {
				if b.Name == "VisibleData" {
					visCount++
				}
			}
		}
	}
	if visCount != 2 {
		t.Errorf("visible DataBand row count = %d, want 2", visCount)
	}
}

// TestInvisibleDataBand_InsideGroupHeader verifies that an invisible DataBand
// nested inside a GroupHeaderBand (drill-down pattern) produces no row output
// while the visible GroupHeader is still rendered.
// This matches the "Drill-Down Groups.frx" pattern where Data1 and GroupFooter1
// are Visible=false by default, collapsed until user interaction.
func TestInvisibleDataBand_InsideGroupHeader(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GroupHeader1")
	gh.SetHeight(20)
	gh.SetCondition("val")

	// GroupFooter is invisible — should be suppressed.
	gf := band.NewGroupFooterBand()
	gf.SetName("GroupFooter1")
	gf.SetHeight(15)
	gf.SetVisible(false)
	gh.SetGroupFooter(gf)

	// DataBand is invisible — rows should be suppressed.
	db := band.NewDataBand()
	db.SetName("Data1")
	db.SetHeight(12)
	db.SetVisible(false)
	db.SetDataSource(newSliceDS("A", "A", "B"))
	gh.SetData(db)

	pg.AddBand(gh)

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

	// GroupHeader should appear (2 groups: "A" and "B").
	if names["GroupHeader1"] != 2 {
		t.Errorf("GroupHeader1 count = %d, want 2", names["GroupHeader1"])
	}
	// Invisible DataBand rows should NOT appear.
	if names["Data1"] != 0 {
		t.Errorf("Data1 (invisible DataBand) count = %d, want 0", names["Data1"])
	}
	// Invisible GroupFooter should NOT appear.
	if names["GroupFooter1"] != 0 {
		t.Errorf("GroupFooter1 (invisible) count = %d, want 0", names["GroupFooter1"])
	}
}
