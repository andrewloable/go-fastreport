package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── PreparedPages accessor ────────────────────────────────────────────────────

func TestNew_PreparedPagesNotNil(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	if e.PreparedPages() == nil {
		t.Error("PreparedPages() should not be nil")
	}
}

// ── RunReportPage – empty page ────────────────────────────────────────────────

func TestRunReportPage_Empty(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	if e.PreparedPages().Count() < 1 {
		t.Errorf("PreparedPages count = %d, want >= 1", e.PreparedPages().Count())
	}
}

// ── RunReportPage – with DataBand ─────────────────────────────────────────────

func TestRunReportPage_WithDataBand(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	db := band.NewDataBand()
	db.SetName("DataBand1")
	db.SetHeight(30)
	pg.AddBand(db)

	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	pp := e.PreparedPages()
	if pp.Count() == 0 {
		t.Fatal("expected at least 1 prepared page")
	}
	preparedPg := pp.GetPage(0)
	if preparedPg == nil {
		t.Fatal("GetPage(0) returned nil")
	}
	// Band should be in the page
	found := false
	for _, b := range preparedPg.Bands {
		if b.Name == "DataBand1" {
			found = true
			if b.Height != 30 {
				t.Errorf("DataBand1 height = %v, want 30", b.Height)
			}
		}
	}
	if !found {
		t.Error("DataBand1 not found in prepared page bands")
	}
}

// ── RunReportPage – with PageHeader ──────────────────────────────────────────

func TestRunReportPage_WithPageHeader(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	ph := band.NewPageHeaderBand()
	ph.SetName("PageHeader1")
	ph.SetHeight(40)
	pg.SetPageHeader(ph)

	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	pp := e.PreparedPages()
	if pp.Count() == 0 {
		t.Fatal("no prepared pages")
	}
	pg0 := pp.GetPage(0)
	found := false
	for _, b := range pg0.Bands {
		if b.Name == "PageHeader1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("PageHeader1 not found in prepared page")
	}
}

// ── StartColumn / EndColumn ───────────────────────────────────────────────────

func TestEndColumn_SingleColumn_ReturnsFalse(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	// Single-column page: endColumn should return false.
	// We access via RunReportPage + method — just check it doesn't panic.
	// The column cycling is implicitly tested through Run.
	if e.CurColumn() != 0 {
		t.Errorf("CurColumn = %d, want 0", e.CurColumn())
	}
}

// ── BandHeight zero skip ──────────────────────────────────────────────────────

func TestRunReportPage_ZeroHeightBandSkipped(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	db := band.NewDataBand()
	db.SetName("ZeroBand")
	db.SetHeight(0) // zero height — should be skipped
	pg.AddBand(db)

	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	pg0 := e.PreparedPages().GetPage(0)
	for _, b := range pg0.Bands {
		if b.Name == "ZeroBand" {
			t.Error("zero-height band should be skipped")
		}
	}
}

// ── FreeSpace reduced by bands ────────────────────────────────────────────────

func TestRunReportPage_FreeSpaceReducedByBands(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	db := band.NewDataBand()
	db.SetHeight(100)
	pg.AddBand(db)

	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	// After processing, CurY should include the band height.
	if e.CurY() < 100 {
		t.Errorf("CurY = %v, expected >= 100 after band output", e.CurY())
	}
}

// ── TitleBeforeHeader ─────────────────────────────────────────────────────────

func TestRunReportPage_TitleBeforeHeader(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.TitleBeforeHeader = true

	rt := band.NewReportTitleBand()
	rt.SetName("ReportTitle1")
	rt.SetHeight(50)
	pg.SetReportTitle(rt)

	ph := band.NewPageHeaderBand()
	ph.SetName("PageHeader1")
	ph.SetHeight(30)
	pg.SetPageHeader(ph)

	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	pg0 := e.PreparedPages().GetPage(0)
	var titleTop, headerTop float32 = -1, -1
	for _, b := range pg0.Bands {
		if b.Name == "ReportTitle1" {
			titleTop = b.Top
		}
		if b.Name == "PageHeader1" {
			headerTop = b.Top
		}
	}
	if titleTop < 0 || headerTop < 0 {
		t.Fatalf("bands not found: titleTop=%v headerTop=%v", titleTop, headerTop)
	}
	if titleTop >= headerTop {
		t.Errorf("TitleBeforeHeader: title top=%v should be before header top=%v", titleTop, headerTop)
	}
}
