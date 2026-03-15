package engine_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// buildReportWithBackPage creates a two-page report where the main page
// references a back page. The back page has a single PageHeader band
// named "BackHeader".
func buildReportWithBackPage() (*reportpkg.Report, *reportpkg.ReportPage) {
	r := reportpkg.NewReport()

	// Back-page template (never printed directly — only used as background).
	backPg := reportpkg.NewReportPage()
	backPg.SetName("BackPage")
	hdr := band.NewPageHeaderBand()
	hdr.SetName("BackHeader")
	hdr.SetHeight(50)
	backPg.SetPageHeader(hdr)
	r.AddPage(backPg)

	// Main page that references the back page.
	mainPg := reportpkg.NewReportPage()
	mainPg.SetName("Page1")
	mainPg.BackPage = "BackPage"
	r.AddPage(mainPg)

	return r, mainPg
}

func TestBackPage_BandsAppearBehindContent(t *testing.T) {
	r, _ := buildReportWithBackPage()
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	if pp.Count() == 0 {
		t.Fatal("no prepared pages")
	}

	// Find the page produced by Page1 (second page template, first produced page).
	// The back page is not directly run as a main page.
	found := false
	for i := 0; i < pp.Count(); i++ {
		pg := pp.GetPage(i)
		for _, b := range pg.Bands {
			if strings.HasSuffix(b.Name, "_back") && strings.Contains(b.Name, "BackHeader") {
				found = true
			}
		}
	}
	if !found {
		t.Error("back page bands (BackHeader_back) should appear in prepared pages")
	}
}

// TestBackPage_OddEven_OddOnly verifies that the back page appears when
// BackPageOddEven=1 (odd only) and the main page is odd-numbered.
// With InitialPageNumber=1 (default), the first startPage makes pageNo=2 (even)
// for the BackPage template, and pageNo=3 (odd) for the Main template.
func TestBackPage_OddEven_OddOnly(t *testing.T) {
	r, mainPg := buildReportWithBackPage()
	mainPg.BackPageOddEven = 1 // odd pages only

	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	pp := e.PreparedPages()
	if pp.Count() < 2 {
		t.Fatalf("expected >= 2 prepared pages, got %d", pp.Count())
	}

	// The main page produces a prepared page that is odd-numbered (pageNo=3).
	// OddOnly should therefore allow the back page to appear.
	pg := pp.GetPage(1) // prepared page index 1 = main template output
	found := false
	for _, b := range pg.Bands {
		if strings.Contains(b.Name, "_back") {
			found = true
		}
	}
	if !found {
		t.Error("back page should appear on odd page when BackPageOddEven=1 (odd only)")
	}
}

// TestBackPage_OddEven_EvenOnly_SkipsOdd verifies that when BackPageOddEven=2
// (even only), the back page does NOT appear on odd-numbered pages.
// The main page template produces an odd-numbered page (pageNo=3).
func TestBackPage_OddEven_EvenOnly_SkipsOdd(t *testing.T) {
	r, mainPg := buildReportWithBackPage()
	mainPg.BackPageOddEven = 2 // even pages only — main page is odd, so skip

	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	pp := e.PreparedPages()
	if pp.Count() < 2 {
		t.Fatalf("expected >= 2 prepared pages, got %d", pp.Count())
	}

	// Main page is odd-numbered — back page should NOT appear.
	pg := pp.GetPage(1)
	for _, b := range pg.Bands {
		if strings.Contains(b.Name, "_back") {
			t.Errorf("back page should NOT appear on odd page when BackPageOddEven=2 (even only)")
		}
	}
}

func TestBackPage_MissingReference_NoError(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("Main")
	pg.BackPage = "NonExistent" // references a page that doesn't exist
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with missing back page should not error, got: %v", err)
	}
}

func TestBackPage_Serialization(t *testing.T) {
	pg := reportpkg.NewReportPage()
	pg.BackPage = "TemplatePage"
	pg.BackPageOddEven = 1

	// Verify that the fields survive clone.
	cp := pg.Clone()
	if cp.BackPage != "TemplatePage" {
		t.Errorf("Clone.BackPage = %q, want TemplatePage", cp.BackPage)
	}
	if cp.BackPageOddEven != 1 {
		t.Errorf("Clone.BackPageOddEven = %d, want 1", cp.BackPageOddEven)
	}
}
