package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── CanPrint ──────────────────────────────────────────────────────────────────

func TestCanPrint_AllPages(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	rc := report.NewReportComponentBase()
	rc.SetVisible(true)
	rc.SetPrintOn(report.PrintOnAllPages)
	if !e.CanPrint(rc, 0, 5) {
		t.Error("AllPages should always be true")
	}
	if !e.CanPrint(rc, 4, 5) {
		t.Error("AllPages should be true on last page too")
	}
}

func TestCanPrint_Invisible(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	rc := report.NewReportComponentBase()
	rc.SetVisible(false)
	if e.CanPrint(rc, 0, 5) {
		t.Error("invisible component should not print")
	}
}

func TestCanPrint_OddPages(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	rc := report.NewReportComponentBase()
	rc.SetVisible(true)
	rc.SetPrintOn(report.PrintOnOddPages)
	if !e.CanPrint(rc, 0, 5) { // page 1 is odd
		t.Error("page index 0 = page 1 (odd), should print")
	}
	if e.CanPrint(rc, 1, 5) { // page 2 is even
		t.Error("page index 1 = page 2 (even), should not print on odd-only")
	}
}

func TestCanPrint_EvenPages(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	rc := report.NewReportComponentBase()
	rc.SetVisible(true)
	rc.SetPrintOn(report.PrintOnEvenPages)
	if e.CanPrint(rc, 0, 5) { // page 1 is odd
		t.Error("page 1 is odd, should not print on even-only")
	}
	if !e.CanPrint(rc, 1, 5) { // page 2 is even
		t.Error("page 2 is even, should print")
	}
}

func TestCanPrint_FirstPage(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	rc := report.NewReportComponentBase()
	rc.SetVisible(true)
	rc.SetPrintOn(report.PrintOnFirstPage)
	if !e.CanPrint(rc, 0, 5) {
		t.Error("should print on first page")
	}
	if e.CanPrint(rc, 1, 5) {
		t.Error("should not print on non-first page")
	}
}

func TestCanPrint_LastPage(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	rc := report.NewReportComponentBase()
	rc.SetVisible(true)
	rc.SetPrintOn(report.PrintOnLastPage)
	if !e.CanPrint(rc, 4, 5) {
		t.Error("should print on last page")
	}
	if e.CanPrint(rc, 0, 5) {
		t.Error("should not print on non-last page")
	}
}

func TestCanPrint_SinglePage(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	rc := report.NewReportComponentBase()
	rc.SetVisible(true)
	rc.SetPrintOn(report.PrintOnSinglePage)
	if !e.CanPrint(rc, 0, 1) {
		t.Error("should print when report is a single page")
	}
	if e.CanPrint(rc, 0, 2) {
		t.Error("should not print on first page of multi-page report")
	}
}

// ── CalcBandHeight ────────────────────────────────────────────────────────────

func TestCalcBandHeight_Basic(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	db := band.NewDataBand()
	db.SetHeight(50)
	if h := e.CalcBandHeight(db); h != 50 {
		t.Errorf("CalcBandHeight = %v, want 50", h)
	}
}

func TestCalcBandHeight_Zero(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	db := band.NewDataBand()
	db.SetHeight(0)
	if h := e.CalcBandHeight(db); h != 0 {
		t.Errorf("CalcBandHeight = %v, want 0", h)
	}
}

// ── AddBandToPreparedPages ────────────────────────────────────────────────────

func TestAddBandToPreparedPages_Basic(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetName("TestBand")
	db.SetHeight(40)
	db.SetVisible(true)

	pp := e.PreparedPages()
	beforeCount := 0
	if pg0 := pp.GetPage(0); pg0 != nil {
		beforeCount = len(pg0.Bands)
	}

	ok := e.AddBandToPreparedPages(&db.BandBase)
	if !ok {
		t.Fatal("AddBandToPreparedPages returned false")
	}

	pg0 := pp.GetPage(0)
	if pg0 == nil {
		t.Fatal("no page 0")
	}
	if len(pg0.Bands) <= beforeCount {
		t.Errorf("band count unchanged: %d", len(pg0.Bands))
	}
}

func TestAddBandToPreparedPages_ZeroHeight(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(0)
	db.SetVisible(true)

	ok := e.AddBandToPreparedPages(&db.BandBase)
	if ok {
		t.Error("AddBandToPreparedPages should return false for zero-height band")
	}
}

// ── ShowFullBand ──────────────────────────────────────────────────────────────

func TestShowFullBand_Basic(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetName("FullBand1")
	db.SetHeight(60)
	db.SetVisible(true)

	beforeY := e.CurY()
	e.ShowFullBand(&db.BandBase)
	if e.CurY() != beforeY+60 {
		t.Errorf("CurY after ShowFullBand = %v, want %v", e.CurY(), beforeY+60)
	}
}

func TestShowFullBand_Repeat(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetName("RepBand")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetRepeatBandNTimes(3)

	beforeY := e.CurY()
	e.ShowFullBand(&db.BandBase)
	if e.CurY() != beforeY+30 {
		t.Errorf("CurY after 3 repeats = %v, want %v", e.CurY(), beforeY+30)
	}
}

func TestShowFullBand_Invisible(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(50)
	db.SetVisible(false)

	beforeY := e.CurY()
	e.ShowFullBand(&db.BandBase)
	if e.CurY() != beforeY {
		t.Errorf("invisible band should not advance CurY: got %v want %v", e.CurY(), beforeY)
	}
}

func TestShowFullBand_Nil(t *testing.T) {
	r := reportpkg.NewReport()
	e := engine.New(r)
	// Should not panic.
	e.ShowFullBand(nil)
}

func TestShowFullBand_WithChild(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	child := band.NewChildBand()
	child.SetName("Child1")
	child.SetHeight(20)
	child.SetVisible(true)

	db := band.NewDataBand()
	db.SetName("Parent1")
	db.SetHeight(30)
	db.SetVisible(true)
	db.SetChild(child)

	beforeY := e.CurY()
	e.ShowFullBand(&db.BandBase)
	// Parent (30) + Child (20) = 50
	if e.CurY() != beforeY+50 {
		t.Errorf("CurY after parent+child = %v, want %v", e.CurY(), beforeY+50)
	}
}

// ── ShowDataBandRow ───────────────────────────────────────────────────────────

func TestShowDataBandRow_SetsRowNo(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(25)
	db.SetVisible(true)

	beforeY := e.CurY()
	e.ShowDataBandRow(db, 3, 7)
	if db.RowNo() != 3 {
		t.Errorf("RowNo = %d, want 3", db.RowNo())
	}
	if db.AbsRowNo() != 7 {
		t.Errorf("AbsRowNo = %d, want 7", db.AbsRowNo())
	}
	if e.CurY() != beforeY+25 {
		t.Errorf("CurY = %v, want %v", e.CurY(), beforeY+25)
	}
}
