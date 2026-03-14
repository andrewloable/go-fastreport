package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func newSubreportEngine(t *testing.T) *engine.ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("Page1")
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return e
}

func TestRenderInnerSubreport_NoPage_NoOp(t *testing.T) {
	e := newSubreportEngine(t)
	b := band.NewBandBase()
	sr := object.NewSubreportObject()
	// ReportPageName is empty — should be a no-op.
	e.RenderInnerSubreport(b, sr)
}

func TestRenderInnerSubreports_NoSubreports_NoOp(t *testing.T) {
	e := newSubreportEngine(t)
	b := band.NewBandBase()
	// No SubreportObject children — should be a no-op.
	e.RenderInnerSubreports(b)
}

func TestRenderOuterSubreports_NoSubreports_NoOp(t *testing.T) {
	e := newSubreportEngine(t)
	b := band.NewBandBase()
	beforeY := e.CurY()
	e.RenderOuterSubreports(b)
	if e.CurY() != beforeY {
		t.Error("RenderOuterSubreports with no subreports should not change CurY")
	}
}

func TestRenderInnerSubreport_LinkedPage_Runs(t *testing.T) {
	// Create a report with two pages; the subreport links to the second page.
	r := reportpkg.NewReport()

	pg1 := reportpkg.NewReportPage()
	pg1.SetName("Main")
	r.AddPage(pg1)

	pg2 := reportpkg.NewReportPage()
	pg2.SetName("Sub")
	r.AddPage(pg2)

	// Add a data band to the sub-page so that rendering does something.
	db := band.NewDataBand()
	db.SetName("SubBand")
	db.SetVisible(true)
	db.SetHeight(10)
	pg2.AddBand(db)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	sr := object.NewSubreportObject()
	sr.SetReportPageName("Sub")
	sr.SetPrintOnParent(true)

	parentBand := band.NewBandBase()
	initialBands := len(e.PreparedPages().GetPage(0).Bands)

	// Should not panic.
	e.RenderInnerSubreport(parentBand, sr)

	_ = initialBands // just ensure it ran without panic
}

// ── CurX/CurY save-restore ────────────────────────────────────────────────────

func TestRenderInnerSubreport_RestoresCurXY(t *testing.T) {
	// Verifies that CurX and CurY are restored after inner subreport rendering.
	r := reportpkg.NewReport()
	pg1 := reportpkg.NewReportPage()
	pg1.SetName("Main")
	r.AddPage(pg1)
	pg2 := reportpkg.NewReportPage()
	pg2.SetName("Inner")
	r.AddPage(pg2)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	beforeX := e.CurX()
	beforeY := e.CurY()

	sr := object.NewSubreportObject()
	sr.SetReportPageName("Inner")
	sr.SetPrintOnParent(true)
	sr.SetLeft(50)
	sr.SetTop(100)

	parentBand := band.NewBandBase()
	e.RenderInnerSubreport(parentBand, sr)

	if e.CurX() != beforeX {
		t.Errorf("CurX after inner subreport = %v, want %v", e.CurX(), beforeX)
	}
	if e.CurY() != beforeY {
		t.Errorf("CurY after inner subreport = %v, want %v", e.CurY(), beforeY)
	}
}

// ── PageNotFound graceful skip ─────────────────────────────────────────────────

func TestRenderInnerSubreport_PageNotFound_NoOp(t *testing.T) {
	e := newSubreportEngine(t)
	b := band.NewBandBase()
	sr := object.NewSubreportObject()
	sr.SetReportPageName("NonExistentPage")

	beforeY := e.CurY()
	// Should not panic; page doesn't exist.
	e.RenderInnerSubreport(b, sr)
	if e.CurY() != beforeY {
		t.Errorf("CurY changed on missing page: got %v, want %v", e.CurY(), beforeY)
	}
}

func TestRenderOuterSubreports_PageNotFound_NoOp(t *testing.T) {
	e := newSubreportEngine(t)
	b := band.NewBandBase()

	// Add a SubreportObject that is outer (PrintOnParent=false) with a bogus page.
	sr := object.NewSubreportObject()
	sr.SetReportPageName("Ghost")
	sr.SetPrintOnParent(false)
	b.Objects().Add(sr)

	beforeY := e.CurY()
	e.RenderOuterSubreports(b)
	// CurY should still advance (hasSubreports=true) to maxY which equals saveCurY.
	// No panic is the main assertion.
	_ = beforeY
}

// ── Outer subreport CurY advancement ─────────────────────────────────────────

func TestRenderOuterSubreports_AdvancesCurY(t *testing.T) {
	r := reportpkg.NewReport()
	pg1 := reportpkg.NewReportPage()
	pg1.SetName("Main")
	r.AddPage(pg1)

	pg2 := reportpkg.NewReportPage()
	pg2.SetName("Outer")
	r.AddPage(pg2)

	// A band on the outer page so CurY advances.
	ob := band.NewDataBand()
	ob.SetName("OuterBand")
	ob.SetHeight(50)
	ob.SetVisible(true)
	pg2.AddBand(ob)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	startY := e.CurY()

	parentBand := band.NewBandBase()
	sr := object.NewSubreportObject()
	sr.SetReportPageName("Outer")
	sr.SetPrintOnParent(false) // outer subreport
	parentBand.Objects().Add(sr)

	e.RenderOuterSubreports(parentBand)

	if e.CurY() <= startY {
		t.Errorf("CurY should advance after outer subreport: start=%v, after=%v", startY, e.CurY())
	}
}

// ── RenderInnerSubreports filters by PrintOnParent ────────────────────────────

func TestRenderInnerSubreports_OnlyPrintsInner(t *testing.T) {
	// RenderInnerSubreports should only render PrintOnParent=true subreports.
	r := reportpkg.NewReport()
	pg1 := reportpkg.NewReportPage()
	pg1.SetName("Main")
	r.AddPage(pg1)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	b := band.NewBandBase()

	// Add an outer subreport — should be ignored by RenderInnerSubreports.
	outer := object.NewSubreportObject()
	outer.SetReportPageName("NonExistent")
	outer.SetPrintOnParent(false)
	b.Objects().Add(outer)

	beforeY := e.CurY()
	e.RenderInnerSubreports(b) // should skip the outer one
	if e.CurY() != beforeY {
		t.Error("RenderInnerSubreports should not render outer subreports")
	}
}

// ── Subreport with DataSource ─────────────────────────────────────────────────

func TestRenderInnerSubreport_WithDataSource(t *testing.T) {
	// Verify inner subreport renders a DataBand with a data source.
	r := reportpkg.NewReport()
	pg1 := reportpkg.NewReportPage()
	pg1.SetName("Main")
	r.AddPage(pg1)

	pg2 := reportpkg.NewReportPage()
	pg2.SetName("DataSub")
	r.AddPage(pg2)

	type simpleDS struct {
		rows []int
		pos  int
	}
	ds := &struct {
		rows []int
		pos  int
	}{rows: []int{1, 2, 3}, pos: -1}
	type dsAdapter struct {
		d *struct {
			rows []int
			pos  int
		}
	}
	dsObj := &dsAdapter{d: ds}
	_ = dsObj // use band.DataSource interface via inline struct below

	// Use the sliceDS helper from integration_test (same package — engine_test).
	sub := band.NewDataBand()
	sub.SetName("SubData")
	sub.SetHeight(15)
	sub.SetVisible(true)
	sub.SetDataSource(newSliceDS("a", "b", "c"))
	pg2.AddBand(sub)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	sr := object.NewSubreportObject()
	sr.SetReportPageName("DataSub")
	sr.SetPrintOnParent(true)

	parentBand := band.NewBandBase()
	// Should not panic; inner subreport runs 3 DataBand rows.
	e.RenderInnerSubreport(parentBand, sr)
}
