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
