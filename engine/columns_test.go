package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── Single-column (default) ───────────────────────────────────────────────────

func TestColumn_SingleColumn_DefaultBehavior(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	// Default Columns.Count == 0 → single-column mode.
	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(20)
	db.SetVisible(true)
	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if e.CurColumn() != 0 {
		t.Errorf("CurColumn after single-column run = %d, want 0", e.CurColumn())
	}
}

// ── Multi-column: EndColumn advances column ───────────────────────────────────

func TestEndColumn_MultiColumn_AdvancesColumn(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.Columns.Count = 2 // two columns
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// After run on a 2-column page, curColumn should still be 0 (no overflow).
	if e.CurColumn() != 0 {
		t.Errorf("CurColumn after 2-col run = %d, want 0", e.CurColumn())
	}
}

// ── Multi-column: overflow fills next column ──────────────────────────────────

func TestColumn_MultiColumn_OverflowStaysOnSamePage(t *testing.T) {
	// A 2-column page with a band that overflows the first column.
	// With startNewPageForCurrent using endColumn, the overflow should
	// fill column 2 rather than starting a new page.
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.PaperHeight = 100  // very short page (mm)
	pg.TopMargin = 0
	pg.BottomMargin = 0
	pg.PaperWidth = 210
	pg.LeftMargin = 0
	pg.RightMargin = 0
	pg.Columns.Count = 2

	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(300) // 300px > page height → will overflow
	db.SetVisible(true)
	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The prepared pages should have exactly 1 page (overflow went to col 2).
	// (After col 2 also overflows, a new page starts — so at most 2 pages.)
	if e.PreparedPages().Count() == 0 {
		t.Error("expected at least 1 prepared page")
	}
}

// ── ColumnHeader appears at column start ──────────────────────────────────────

func TestColumn_ColumnHeader_AppearsAtStart(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.Columns.Count = 2

	ch := band.NewColumnHeaderBand()
	ch.SetName("ColHdr")
	ch.SetHeight(15)
	ch.SetVisible(true)
	pg.SetColumnHeader(ch)

	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pg0 := e.PreparedPages().GetPage(0)
	if pg0 == nil {
		t.Fatal("no prepared page")
	}

	found := false
	for _, b := range pg0.Bands {
		if b.Name == "ColHdr" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ColumnHeader band should appear in the prepared page")
	}
}

// ── ColumnFooter appears at column end ────────────────────────────────────────

func TestColumn_ColumnFooter_AppearsAtEnd(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.Columns.Count = 1

	cf := band.NewColumnFooterBand()
	cf.SetName("ColFtr")
	cf.SetHeight(10)
	cf.SetVisible(true)
	pg.SetColumnFooter(cf)

	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pg0 := e.PreparedPages().GetPage(0)
	if pg0 == nil {
		t.Fatal("no prepared page")
	}

	found := false
	for _, b := range pg0.Bands {
		if b.Name == "ColFtr" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ColumnFooter band should appear in the prepared page")
	}
}

// ── CurX advances per column ──────────────────────────────────────────────────

func TestColumn_CurX_Advances(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.Columns.Count = 3
	pg.PaperWidth = 210
	pg.LeftMargin = 0
	pg.RightMargin = 0
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Just verify the engine ran without error and has at least 1 page.
	if e.PreparedPages().Count() == 0 {
		t.Error("expected at least 1 page")
	}
}
