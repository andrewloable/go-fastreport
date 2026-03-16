package engine_test

// band_print_callbacks_test.go — verifies that the engine fires
// OnBeforePrint and OnAfterPrint on BandBase (via FireBeforePrint /
// FireAfterPrint) during ShowFullBand.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// TestShowFullBand_FiresBeforePrint verifies OnBeforePrint is called before
// the band is rendered (i.e. before CurY advances).
func TestShowFullBand_FiresBeforePrint(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetName("PrintCbBand")
	db.SetHeight(50)
	db.SetVisible(true)

	var beforePrintCalled bool
	var beforePrintY float32 = -1
	db.OnBeforePrint = func(_ report.Base, _ *report.EventArgs) {
		beforePrintCalled = true
		beforePrintY = e.CurY()
	}

	startY := e.CurY()
	e.ShowFullBand(&db.BandBase)

	if !beforePrintCalled {
		t.Fatal("OnBeforePrint was not called")
	}
	// BeforePrint must fire before CurY advances.
	if beforePrintY != startY {
		t.Errorf("OnBeforePrint fired with CurY=%v, want %v (should fire before rendering)", beforePrintY, startY)
	}
}

// TestShowFullBand_FiresAfterPrint verifies OnAfterPrint is called after the
// band has been rendered (i.e. after CurY has advanced).
func TestShowFullBand_FiresAfterPrint(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetName("AfterPrintCbBand")
	db.SetHeight(50)
	db.SetVisible(true)

	var afterPrintCalled bool
	var afterPrintY float32 = -1
	db.OnAfterPrint = func(_ report.Base, _ *report.EventArgs) {
		afterPrintCalled = true
		afterPrintY = e.CurY()
	}

	startY := e.CurY()
	e.ShowFullBand(&db.BandBase)

	if !afterPrintCalled {
		t.Fatal("OnAfterPrint was not called")
	}
	// AfterPrint must fire after CurY has advanced by the band height.
	if afterPrintY != startY+50 {
		t.Errorf("OnAfterPrint fired with CurY=%v, want %v (should fire after rendering)", afterPrintY, startY+50)
	}
}

// TestShowFullBand_BothCallbacks verifies that both callbacks fire in order
// and are called exactly once per ShowFullBand invocation.
func TestShowFullBand_BothCallbacks(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetName("BothCbBand")
	db.SetHeight(30)
	db.SetVisible(true)

	var order []string
	db.OnBeforePrint = func(_ report.Base, _ *report.EventArgs) {
		order = append(order, "before")
	}
	db.OnAfterPrint = func(_ report.Base, _ *report.EventArgs) {
		order = append(order, "after")
	}

	e.ShowFullBand(&db.BandBase)

	if len(order) != 2 {
		t.Fatalf("expected 2 callback invocations, got %d: %v", len(order), order)
	}
	if order[0] != "before" || order[1] != "after" {
		t.Errorf("callback order = %v, want [before after]", order)
	}
}

// TestShowFullBand_PrintCallbacks_Invisible verifies that neither callback is
// fired when the band is invisible (band is skipped entirely).
func TestShowFullBand_PrintCallbacks_Invisible(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(30)
	db.SetVisible(false)

	var called bool
	db.OnBeforePrint = func(_ report.Base, _ *report.EventArgs) { called = true }
	db.OnAfterPrint = func(_ report.Base, _ *report.EventArgs) { called = true }

	e.ShowFullBand(&db.BandBase)

	if called {
		t.Error("callbacks should not fire for invisible bands")
	}
}

// TestShowFullBand_PrintCallbacks_ZeroHeight verifies that both callbacks are
// still fired even when the band has zero height (nothing is added to pages
// but the callbacks must still execute for correctness).
func TestShowFullBand_PrintCallbacks_ZeroHeight(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(0)
	db.SetVisible(true)

	var beforeCalled, afterCalled bool
	db.OnBeforePrint = func(_ report.Base, _ *report.EventArgs) { beforeCalled = true }
	db.OnAfterPrint = func(_ report.Base, _ *report.EventArgs) { afterCalled = true }

	e.ShowFullBand(&db.BandBase)

	if !beforeCalled {
		t.Error("OnBeforePrint should fire even for zero-height band")
	}
	if !afterCalled {
		t.Error("OnAfterPrint should fire even for zero-height band")
	}
}
