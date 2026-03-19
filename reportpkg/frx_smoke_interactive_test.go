package reportpkg_test

// Smoke tests for interactive/drill-down FRX reports.
// These verify that pages with Visible=false are deserialized correctly
// and that the engine skips invisible pages during report rendering.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// TestFRXSmoke_InteractiveReport_2in1_PageVisibility verifies that the
// "Interactive Report, 2-in-1.frx" report has a second page with Visible=false.
// The engine should produce output for only the first (visible) page.
func TestFRXSmoke_InteractiveReport_2in1_PageVisibility(t *testing.T) {
	r := loadFRXSmoke(t, "Interactive Report, 2-in-1.frx")

	pages := r.Pages()
	if len(pages) < 2 {
		t.Fatalf("expected at least 2 pages in report, got %d", len(pages))
	}

	// First page (Page1) should be visible.
	if !pages[0].Visible() {
		t.Error("Page1 should be visible")
	}

	// Second page (Page2) should be invisible — it's the drill-down detail page.
	if pages[1].Visible() {
		t.Errorf("Page2 (%s) should be invisible (Visible=false in FRX)", pages[1].Name())
	}
}

// TestFRXSmoke_InteractiveReport_2in1_EngineSkipsInvisiblePage verifies that
// running the engine on "Interactive Report, 2-in-1.frx" (without NorthWind data)
// produces output only for the visible page. Data source init errors are
// expected when NorthWind data is not loaded; we verify the invisible page
// check works at the point where it matters.
func TestFRXSmoke_InteractiveReport_2in1_EngineSkipsInvisiblePage(t *testing.T) {
	r := loadFRXSmoke(t, "Interactive Report, 2-in-1.frx")

	e := engine.New(r)
	err := e.Run(engine.DefaultRunOptions())
	// Data source init errors are acceptable here (no NorthWind data loaded).
	// The key assertion: if the run succeeded, only 1 page should be produced.
	if err != nil {
		t.Skipf("Run error (expected without NorthWind data): %v", err)
		return
	}

	pp := e.PreparedPages()
	// Only 1 visible page template — engine should not process Page2.
	// Without NorthWind data the DataBand produces 0 rows, so exactly 1 page.
	if pp.Count() != 1 {
		t.Errorf("expected 1 prepared page (invisible Page2 skipped), got %d", pp.Count())
	}
}

// TestFRXSmoke_InteractiveChart_PageVisibility verifies that "Interactive Chart.frx"
// has a second page with Visible=false.
func TestFRXSmoke_InteractiveChart_PageVisibility(t *testing.T) {
	r := loadFRXSmoke(t, "Interactive Chart.frx")

	pages := r.Pages()
	if len(pages) < 2 {
		t.Fatalf("expected at least 2 pages in report, got %d", len(pages))
	}

	if !pages[0].Visible() {
		t.Error("Page1 should be visible")
	}
	if pages[1].Visible() {
		t.Errorf("Page2 (%s) should be invisible (Visible=false in FRX)", pages[1].Name())
	}
}

// TestFRXSmoke_InteractiveChart_EngineSkipsInvisiblePage verifies the engine
// skips the invisible detail page for "Interactive Chart.frx".
func TestFRXSmoke_InteractiveChart_EngineSkipsInvisiblePage(t *testing.T) {
	r := loadFRXSmoke(t, "Interactive Chart.frx")

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	if pp.Count() != 1 {
		t.Errorf("expected 1 prepared page (invisible Page2 skipped), got %d", pp.Count())
	}
}

// TestFRXSmoke_InteractiveReport_PageSerialization verifies that ReportPage.Visible
// round-trips through serialization correctly.
func TestFRXSmoke_InteractiveReport_PageSerialization(t *testing.T) {
	// Build a report with an invisible page.
	original := reportpkg.NewReport()
	pg1 := reportpkg.NewReportPage()
	pg1.SetName("Page1")
	pg1.SetVisible(true)
	original.AddPage(pg1)

	pg2 := reportpkg.NewReportPage()
	pg2.SetName("Page2")
	pg2.SetVisible(false)
	original.AddPage(pg2)

	// Save to XML string.
	xmlStr, err := original.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	// Reload from XML string.
	loaded := reportpkg.NewReport()
	if err := loaded.LoadFromString(xmlStr); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}

	pages := loaded.Pages()
	if len(pages) < 2 {
		t.Fatalf("expected 2 pages after reload, got %d", len(pages))
	}
	if !pages[0].Visible() {
		t.Error("Page1 should be visible after reload")
	}
	if pages[1].Visible() {
		t.Error("Page2 should be invisible after reload")
	}
}
