package engine

// pages_coverage_gaps_internal_test.go — internal (package engine) tests to
// cover remaining branches in pages.go that cannot be reached from external tests:
//
//  1. showBandNoAdvance: invisible band (Visible()==false) → early return
//  2. showBandNoAdvance: height <= 0 → early return
//  3. showBandNoAdvance: nil band → early return
//  4. showBandNoAdvance: valid visible band → advances curY
//  5. endColumn: curColumn >= cols → reset to 0, return false (all columns filled)
//  6. runBands: default branch (band that is not DataBand or GroupHeaderBand)
//  8. showBand: ContainsBand dedup — page bands not added twice to same page

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func buildPagesInternalEngine(t *testing.T) *ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return e
}

// ── 1: showBandNoAdvance — invisible band ─────────────────────────────────────

func TestShowBandNoAdvance_InvisibleBand_NoAdvance(t *testing.T) {
	e := buildPagesInternalEngine(t)
	startY := e.curY

	db := band.NewDataBand()
	db.SetName("InvisibleBand")
	db.SetHeight(20)
	db.SetVisible(false) // invisible → should be skipped

	e.showBandNoAdvance(db)

	if e.curY != startY {
		t.Errorf("showBandNoAdvance invisible: curY = %v, want %v (no advance)", e.curY, startY)
	}
}

// ── 2: showBandNoAdvance — zero height ────────────────────────────────────────

func TestShowBandNoAdvance_ZeroHeight_NoAdvance(t *testing.T) {
	e := buildPagesInternalEngine(t)
	startY := e.curY

	db := band.NewDataBand()
	db.SetName("ZeroHeightBand")
	db.SetHeight(0) // height <= 0 → should be skipped
	db.SetVisible(true)

	e.showBandNoAdvance(db)

	if e.curY != startY {
		t.Errorf("showBandNoAdvance zero-height: curY = %v, want %v (no advance)", e.curY, startY)
	}
}

// ── 3: showBandNoAdvance — nil band ──────────────────────────────────────────

func TestShowBandNoAdvance_NilBand_NoPanic(t *testing.T) {
	e := buildPagesInternalEngine(t)
	startY := e.curY

	// nil band → should return immediately without panic.
	e.showBandNoAdvance(nil)

	if e.curY != startY {
		t.Errorf("showBandNoAdvance nil: curY should not change, got %v", e.curY)
	}
}

// ── 4: showBandNoAdvance — valid visible band advances curY ──────────────────

func TestShowBandNoAdvance_ValidBand_Advances(t *testing.T) {
	e := buildPagesInternalEngine(t)
	startY := e.curY

	db := band.NewDataBand()
	db.SetName("ValidBackBand")
	db.SetHeight(25)
	db.SetVisible(true)

	e.showBandNoAdvance(db)

	// curY should advance by the band height.
	if e.curY != startY+25 {
		t.Errorf("showBandNoAdvance valid: curY = %v, want %v", e.curY, startY+25)
	}
}

// ── 5: endColumn — all columns filled (curColumn >= cols → reset, return false)

// TestEndColumn_WrapsToZero tests that when endColumn is called while the
// engine is on the last column (curColumn+1 == cols), it wraps curColumn back
// to 0 and returns false (caller should start a new page).
func TestEndColumn_WrapsToZero(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.Columns.Count = 2
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Position curColumn at the last slot (index 1 of a 2-column layout).
	e.curColumn = 1
	e.currentPage = pg

	// endColumn increments curColumn to 2, which >= cols(2), so wraps to 0
	// and returns false.
	result := e.endColumn(pg)
	if result {
		t.Error("endColumn: expected false when all columns are filled (wraps around)")
	}
	if e.curColumn != 0 {
		t.Errorf("endColumn: curColumn = %d, want 0 after wrap", e.curColumn)
	}
}

// ── 6: runBands default branch ────────────────────────────────────────────────
//
// The default case in runBands dispatches a band that is neither DataBand nor
// GroupHeaderBand to showBand. A PageHeaderBand exercises this path.

func TestRunBands_DefaultBranch_ShowsBand(t *testing.T) {
	e := buildPagesInternalEngine(t)

	ph := band.NewPageHeaderBand()
	ph.SetName("PHDefault")
	ph.SetHeight(15)
	ph.SetVisible(true)

	pg0 := e.preparedPages.GetPage(0)
	before := len(pg0.Bands)

	bands := []report.Base{ph}
	if err := e.runBands(bands); err != nil {
		t.Fatalf("runBands: %v", err)
	}

	after := len(pg0.Bands)
	if after <= before {
		t.Errorf("runBands default: expected band to be added; before=%d after=%d", before, after)
	}
}

// ── 7: runBands aborted path ──────────────────────────────────────────────────

func TestRunBands_Aborted_StopsEarly(t *testing.T) {
	e := buildPagesInternalEngine(t)

	ph := band.NewPageHeaderBand()
	ph.SetName("AbortedBand")
	ph.SetHeight(10)
	ph.SetVisible(true)

	pg0 := e.preparedPages.GetPage(0)
	before := len(pg0.Bands)

	e.aborted = true
	bands := []report.Base{ph}
	if err := e.runBands(bands); err != nil {
		t.Fatalf("runBands aborted: %v", err)
	}

	after := len(pg0.Bands)
	if after != before {
		t.Errorf("runBands aborted: expected no bands added; before=%d after=%d", before, after)
	}
}

// ── 8: showBand — ContainsBand deduplication for page bands ──────────────────
// Mirrors C# AddToPreparedPages lines 484-499 (ReportEngine.Bands.cs):
// page bands (PageHeader, PageFooter, Overlay) must not be added to the same
// prepared page twice (as can happen when rendering subreports).

func TestShowBand_PageHeaderDedup_NotAddedTwice(t *testing.T) {
	e := buildPagesInternalEngine(t)

	ph := band.NewPageHeaderBand()
	ph.SetName("PageHeader1")
	ph.SetHeight(30)
	ph.SetVisible(true)

	pg0 := e.preparedPages.GetPage(0)
	countBefore := len(pg0.Bands)

	// Call showBand twice with the same PageHeaderBand type.
	e.showBand(ph)
	countAfter1 := len(pg0.Bands)
	e.showBand(ph)
	countAfter2 := len(pg0.Bands)

	if countAfter1 != countBefore+1 {
		t.Errorf("first showBand: want %d bands, got %d", countBefore+1, countAfter1)
	}
	if countAfter2 != countAfter1 {
		t.Errorf("second showBand (duplicate): want %d bands (no change), got %d", countAfter1, countAfter2)
	}
}

func TestShowBand_PageFooterDedup_NotAddedTwice(t *testing.T) {
	e := buildPagesInternalEngine(t)

	pf := band.NewPageFooterBand()
	pf.SetName("PageFooter1")
	pf.SetHeight(25)
	pf.SetVisible(true)

	pg0 := e.preparedPages.GetPage(0)
	countBefore := len(pg0.Bands)

	e.showBand(pf)
	countAfter1 := len(pg0.Bands)
	e.showBand(pf)
	countAfter2 := len(pg0.Bands)

	if countAfter1 != countBefore+1 {
		t.Errorf("first showBand PageFooter: want %d bands, got %d", countBefore+1, countAfter1)
	}
	if countAfter2 != countAfter1 {
		t.Errorf("second showBand PageFooter (duplicate): want %d bands (no change), got %d", countAfter1, countAfter2)
	}
}

func TestShowBand_OverlayDedup_NotAddedTwice(t *testing.T) {
	e := buildPagesInternalEngine(t)

	ov := band.NewOverlayBand()
	ov.SetName("Overlay1")
	ov.SetHeight(20)
	ov.SetVisible(true)

	pg0 := e.preparedPages.GetPage(0)
	countBefore := len(pg0.Bands)

	e.showBand(ov)
	countAfter1 := len(pg0.Bands)
	e.showBand(ov)
	countAfter2 := len(pg0.Bands)

	if countAfter1 != countBefore+1 {
		t.Errorf("first showBand Overlay: want %d bands, got %d", countBefore+1, countAfter1)
	}
	if countAfter2 != countAfter1 {
		t.Errorf("second showBand Overlay (duplicate): want %d bands (no change), got %d", countAfter1, countAfter2)
	}
}

func TestShowBand_DataBand_AllowedTwice(t *testing.T) {
	e := buildPagesInternalEngine(t)

	db := band.NewDataBand()
	db.SetName("DataBand1")
	db.SetHeight(20)
	db.SetVisible(true)

	pg0 := e.preparedPages.GetPage(0)
	countBefore := len(pg0.Bands)

	// Non-page bands (DataBand) should always be added — no deduplication.
	e.showBand(db)
	countAfter1 := len(pg0.Bands)
	e.showBand(db)
	countAfter2 := len(pg0.Bands)

	if countAfter1 != countBefore+1 {
		t.Errorf("first showBand DataBand: want %d bands, got %d", countBefore+1, countAfter1)
	}
	if countAfter2 != countAfter1+1 {
		t.Errorf("second showBand DataBand: want %d bands (added again), got %d", countAfter1+1, countAfter2)
	}
}
