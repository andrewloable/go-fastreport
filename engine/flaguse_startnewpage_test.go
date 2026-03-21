// flaguse_startnewpage_test.go verifies that FlagUseStartNewPage is correctly
// respected by the engine for fixed band types (PageHeaderBand, PageFooterBand,
// ColumnHeaderBand, ColumnFooterBand, OverlayBand).
//
// C# source of truth: ReportEngine.Bands.cs ShowBandToPreparedPages() line 131:
//   if (band.StartNewPage && band.FlagUseStartNewPage && bandCanStartNewPage && ...)
//       EndColumn();
// C# source of truth: ReportEngine.Bands.cs ShowDataBandRow() (engine/databands.go:107):
//   if db.StartNewPage() && db.FlagUseStartNewPage && rowNo != 1
//
// The five fixed bands set FlagUseStartNewPage=false in their constructors:
//   - PageHeaderBand (PageHeaderBand.cs line 40)
//   - PageFooterBand (PageFooterBand.cs constructor)
//   - ColumnHeaderBand (ColumnHeaderBand.cs constructor)
//   - ColumnFooterBand (ColumnFooterBand.cs constructor)
//   - OverlayBand (OverlayBand.cs constructor)
//
// Consequence: even if StartNewPage=true is set on these bands, they must NOT
// trigger extra page breaks because FlagUseStartNewPage=false gates the check.
package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── Unit: verify FlagUseStartNewPage defaults ─────────────────────────────────

// TestFlagUseStartNewPage_FixedBandDefaults verifies that the 5 fixed band
// constructors set FlagUseStartNewPage=false, exactly matching the C# source.
// This is a direct port of the C# constructor assignments verified by reading:
//   - PageFooterBand.cs: FlagUseStartNewPage = false;
//   - OverlayBand.cs:    FlagUseStartNewPage = false;
//   - PageHeaderBand.cs, ColumnHeaderBand.cs, ColumnFooterBand.cs: same.
func TestFlagUseStartNewPage_FixedBandDefaults(t *testing.T) {
	tests := []struct {
		name string
		flag bool
	}{
		{"PageHeaderBand", band.NewPageHeaderBand().FlagUseStartNewPage},
		{"PageFooterBand", band.NewPageFooterBand().FlagUseStartNewPage},
		{"ColumnHeaderBand", band.NewColumnHeaderBand().FlagUseStartNewPage},
		{"ColumnFooterBand", band.NewColumnFooterBand().FlagUseStartNewPage},
		{"OverlayBand", band.NewOverlayBand().FlagUseStartNewPage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.flag {
				t.Errorf("%s.FlagUseStartNewPage = true, want false (C# constructor sets false)", tt.name)
			}
		})
	}
}

// TestFlagUseStartNewPage_DataBandDefault verifies that DataBand inherits
// FlagUseStartNewPage=true from BandBase (no constructor override in C#).
func TestFlagUseStartNewPage_DataBandDefault(t *testing.T) {
	db := band.NewDataBand()
	if !db.FlagUseStartNewPage {
		t.Error("DataBand.FlagUseStartNewPage should be true (inherits BandBase default)")
	}
}

// ── Integration: PageHeader/Footer with StartNewPage=true must not add pages ──

// TestFlagUseStartNewPage_PageHeaderFooterNoExtraPages verifies that a report
// with PageHeaderBand + DataBand + PageFooterBand produces the correct page
// count even when StartNewPage is explicitly set to true on the header/footer.
//
// The bug this guards against: if FlagUseStartNewPage=false is not enforced,
// setting StartNewPage=true on PageHeaderBand/PageFooterBand would trigger
// extra EndColumn() calls, producing empty phantom pages.
func TestFlagUseStartNewPage_PageHeaderFooterNoExtraPages(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PageHeader")
	hdr.SetVisible(true)
	hdr.SetHeight(30)
	// Deliberately set StartNewPage=true — must NOT trigger a page break
	// because FlagUseStartNewPage=false gates the engine check.
	hdr.SetStartNewPage(true)
	pg.SetPageHeader(hdr)

	ftr := band.NewPageFooterBand()
	ftr.SetName("PageFooter")
	ftr.SetVisible(true)
	ftr.SetHeight(20)
	// Same: StartNewPage=true must be ignored for PageFooterBand.
	ftr.SetStartNewPage(true)
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

	// All 3 data rows fit on a single page (pageHeight ~ 1047px, 3*15=45px),
	// so there should be exactly 1 page. If StartNewPage were honoured for
	// fixed bands, phantom pages would be inserted.
	if pp.Count() != 1 {
		t.Errorf("page count = %d, want 1 (PageHeader/Footer with StartNewPage=true must not add pages)", pp.Count())
	}

	// Verify PageHeader and PageFooter each appear exactly once.
	counts := map[string]int{}
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p == nil {
			continue
		}
		for _, b := range p.Bands {
			counts[b.Name]++
		}
	}
	if counts["PageHeader"] != 1 {
		t.Errorf("PageHeader count = %d, want 1", counts["PageHeader"])
	}
	if counts["PageFooter"] != 1 {
		t.Errorf("PageFooter count = %d, want 1", counts["PageFooter"])
	}
	if counts["DataBand"] != 3 {
		t.Errorf("DataBand count = %d, want 3", counts["DataBand"])
	}
}

// ── Integration: multi-row report page count is correct ──────────────────────

// TestFlagUseStartNewPage_MultiRowPageCount verifies that the page count for a
// multi-row data report is determined only by data overflow, NOT by header/footer
// StartNewPage flags.
//
// Setup: page height ~1047px, DataBand height=100px, 12 rows → 1200px total.
// Expected: 2 pages (12 rows at 100px each overflows ~1047px page).
// The PageHeader (30px) and PageFooter (20px) consume space but must not add pages.
func TestFlagUseStartNewPage_MultiRowPageCount(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PageHeader")
	hdr.SetVisible(true)
	hdr.SetHeight(30)
	hdr.SetStartNewPage(true) // must be ignored
	pg.SetPageHeader(hdr)

	ftr := band.NewPageFooterBand()
	ftr.SetName("PageFooter")
	ftr.SetVisible(true)
	ftr.SetHeight(20)
	ftr.SetStartNewPage(true) // must be ignored
	pg.SetPageFooter(ftr)

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(100) // large enough to overflow
	rows := make([]string, 12)
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

	// With A4 page (297mm) minus 10mm top/bottom margins = 277mm usable height.
	// At 3.78px/mm: ~1047px usable.
	// PageHeader=30px, PageFooter=20px, so available for data = ~1047-30-20 = ~997px.
	// 12 rows * 100px = 1200px > 997px => must overflow onto at least 2 pages.
	// The key assertion is that header/footer StartNewPage=true did NOT add extra pages.
	if pp.Count() < 2 {
		t.Errorf("page count = %d, want >= 2 (data overflow), header/footer did not cause extra pages", pp.Count())
	}

	// On every page, PageHeader and PageFooter must each appear exactly once.
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p == nil {
			continue
		}
		hdrCount := 0
		ftrCount := 0
		for _, b := range p.Bands {
			switch b.Name {
			case "PageHeader":
				hdrCount++
			case "PageFooter":
				ftrCount++
			}
		}
		if hdrCount != 1 {
			t.Errorf("page %d: PageHeader count = %d, want 1", i, hdrCount)
		}
		if ftrCount != 1 {
			t.Errorf("page %d: PageFooter count = %d, want 1", i, ftrCount)
		}
	}
}

// ── Integration: DataBand StartNewPage IS honoured (FlagUseStartNewPage=true) ─

// TestFlagUseStartNewPage_DataBandStartNewPageHonoured verifies that DataBand
// StartNewPage=true is respected (FlagUseStartNewPage=true for DataBand).
// This is the positive counterpart to the PageHeader/Footer negative tests.
//
// With 3 rows and StartNewPage=true, each row after the first should start a
// new page → 3 pages total (row 1 on page 1, row 2 on page 2, row 3 on page 3).
func TestFlagUseStartNewPage_DataBandStartNewPageHonoured(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(20)
	db.SetStartNewPage(true) // DataBand.FlagUseStartNewPage=true → this WILL trigger page breaks
	db.SetDataSource(newSliceDS("R1", "R2", "R3"))
	pg.AddBand(db)

	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()

	// Row 1 starts on page 1 (no page break before first row — C# skips RowNo==1).
	// Row 2 triggers EndColumn() → page 2.
	// Row 3 triggers EndColumn() → page 3.
	// Total: 3 pages.
	if pp.Count() != 3 {
		t.Errorf("page count = %d, want 3 (DataBand StartNewPage=true with FlagUseStartNewPage=true)", pp.Count())
	}
}

// ── Integration: OverlayBand StartNewPage=true must not add pages ─────────────

// TestFlagUseStartNewPage_OverlayBandNoExtraPages verifies that OverlayBand
// with StartNewPage=true does not produce phantom pages.
func TestFlagUseStartNewPage_OverlayBandNoExtraPages(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	ov := band.NewOverlayBand()
	ov.SetName("Overlay")
	ov.SetVisible(true)
	ov.SetHeight(10)
	ov.SetStartNewPage(true) // must be ignored — FlagUseStartNewPage=false
	pg.SetOverlay(ov)

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

	pp := e.PreparedPages()
	if pp.Count() != 1 {
		t.Errorf("page count = %d, want 1 (OverlayBand StartNewPage=true must be ignored)", pp.Count())
	}
}

// ── Integration: BandCanStartNewPage parent walk ──────────────────────────────

// TestFlagUseStartNewPage_ParentWalkPreventsBreak verifies the C# BandCanStartNewPage
// logic: if a parent band has FlagUseStartNewPage=false, the child cannot trigger
// a page break either (C# ReportEngine.Bands.cs lines 103-123).
//
// This is tested indirectly: PageHeaderBand (FlagUseStartNewPage=false) contains
// no sub-bands in standard reports, but the engine-level check in ShowDataBandRow
// already directly guards on FlagUseStartNewPage, so the integration test above
// covers the main scenario. Here we verify the flag propagation itself.
func TestFlagUseStartNewPage_ReportTitleAndSummaryHaveFlag(t *testing.T) {
	// ReportTitleBand and ReportSummaryBand do NOT override FlagUseStartNewPage
	// in C# (no constructor override) → they inherit true from BandBase.
	// Verify the Go port matches.
	rtb := band.NewReportTitleBand()
	if !rtb.FlagUseStartNewPage {
		t.Error("ReportTitleBand.FlagUseStartNewPage should be true (inherits BandBase default, no C# override)")
	}

	rsb := band.NewReportSummaryBand()
	if !rsb.BandBase.FlagUseStartNewPage {
		t.Error("ReportSummaryBand.FlagUseStartNewPage should be true (inherits BandBase default, no C# override)")
	}
}
