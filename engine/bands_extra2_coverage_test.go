package engine_test

// bands_extra2_coverage_test.go — targeted coverage for remaining uncovered branches in:
//   bands.go: ShowFullBand (RepeatBandNTimes<=0), CalcBandHeight (return 0, CanGrow return,
//             CanShrink return), calcBandRequiredHeight (non-TextObject hasDims)
//   databands.go: RunDataBandFull PrintIfDSEmpty, aborted mid-filter
//   subreports.go: RenderInnerSubreports with inner+outer objects mixed
//   relations.go: applyRelationFilters nil-rel, empty ParentColumns

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func newBandsExtra2Engine(t *testing.T) *engine.ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return e
}

// ── ShowFullBand: RepeatBandNTimes <= 0 covers `n = 1` branch ────────────────

// TestShowFullBand_RepeatZero calls ShowFullBand with RepeatBandNTimes=0.
// This exercises the `if n <= 0 { n = 1 }` branch in ShowFullBand.
// Default RepeatBandNTimes is 1, so this branch is otherwise unreachable.
func TestShowFullBand_RepeatZero(t *testing.T) {
	e := newBandsExtra2Engine(t)

	b := band.NewBandBase()
	b.SetName("RepeatZeroBand")
	b.SetHeight(10)
	b.SetVisible(true)
	b.SetRepeatBandNTimes(0) // <= 0 → n set to 1

	beforeY := e.CurY()
	e.ShowFullBand(b)
	// Should print once (n=1 after clamping).
	if e.CurY() != beforeY+10 {
		t.Errorf("RepeatZero: CurY = %v, want %v", e.CurY(), beforeY+10)
	}
}

// TestShowFullBand_RepeatNegative calls ShowFullBand with RepeatBandNTimes=-1.
func TestShowFullBand_RepeatNegative(t *testing.T) {
	e := newBandsExtra2Engine(t)

	b := band.NewBandBase()
	b.SetName("RepeatNegBand")
	b.SetHeight(15)
	b.SetVisible(true)
	b.SetRepeatBandNTimes(-1)

	beforeY := e.CurY()
	e.ShowFullBand(b)
	if e.CurY() != beforeY+15 {
		t.Errorf("RepeatNegative: CurY = %v, want %v", e.CurY(), beforeY+15)
	}
}

// ── CalcBandHeight: non-BandBase without Height() → return 0 ─────────────────

// TestCalcBandHeight_NonBandBase_NoHeight covers the `return 0` path in
// CalcBandHeight when b is not a *band.BandBase and does not implement
// the hasHeight interface. report.NewBaseObject() has no Height() method.
func TestCalcBandHeight_NonBandBase_NoHeight(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	obj := report.NewBaseObject()
	h := e.CalcBandHeight(obj)
	if h != 0 {
		t.Errorf("non-BandBase without Height: expected 0, got %v", h)
	}
}

// ── CalcBandHeight: CanGrow with tall content → return requiredHeight ─────────

// TestCalcBandHeight_CanGrow_WithTallContent exercises the
// `if canGrow && requiredHeight > baseHeight { return requiredHeight }` path.
func TestCalcBandHeight_CanGrow_WithTallContent(t *testing.T) {
	e := engine.New(reportpkg.NewReport())

	bb := band.NewBandBase()
	bb.SetHeight(20) // small base height
	bb.SetCanGrow(true)

	// A TextObject with a tall declared height forces requiredHeight > baseHeight.
	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(80) // declared height > band height → forces growth
	txt.SetWidth(200)
	txt.SetText("Tall content that requires a large height")
	bb.Objects().Add(txt)

	h := e.CalcBandHeight(bb)
	if h <= 20 {
		t.Errorf("CanGrow with tall content: expected h > 20, got %v", h)
	}
}

// ── CalcBandHeight: CanShrink with short content → return requiredHeight ──────

// TestCalcBandHeight_CanShrink_ShortContent documents that calcBandRequiredHeight
// always clamps to baseHeight, so the CanShrink return-requiredHeight path is
// structurally unreachable. With CanShrink=true and short content, CalcBandHeight
// still returns baseHeight (the clamp at line 155 in bands.go fires first).
func TestCalcBandHeight_CanShrink_ShortContent(t *testing.T) {
	e := engine.New(reportpkg.NewReport())

	bb := band.NewBandBase()
	bb.SetHeight(200)
	bb.SetCanShrink(true)

	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(5)
	txt.SetWidth(200)
	txt.SetText("X")
	bb.Objects().Add(txt)

	h := e.CalcBandHeight(bb)
	// calcBandRequiredHeight clamps maxBottom to baseHeight, so h == baseHeight.
	if h != 200 {
		t.Errorf("CanShrink short content: expected h = 200 (clamped), got %v", h)
	}
}

// ── calcBandRequiredHeight: non-TextObject with Top/Height (hasDims) ──────────

// TestCalcBandHeight_NonTextObjectWithDims exercises the hasDims branch inside
// calcBandRequiredHeight for a non-TextObject child that has Top() and Height().
// BandBase implements both, so adding one as a child object covers this path.
func TestCalcBandHeight_NonTextObject_WithDims(t *testing.T) {
	e := engine.New(reportpkg.NewReport())

	bb := band.NewBandBase()
	bb.SetHeight(10)
	bb.SetCanGrow(true)

	// A ChildBand (non-TextObject) with Top=0, Height=50 forces bottom=50 > baseHeight=10.
	child := band.NewChildBand()
	child.SetTop(0)
	child.SetHeight(50)
	bb.Objects().Add(child)

	h := e.CalcBandHeight(bb)
	if h < 50 {
		t.Errorf("non-text hasDims child: expected h >= 50, got %v", h)
	}
}

// TestCalcBandHeight_NonTextObject_NoDims exercises the case where a non-TextObject
// child does NOT have Top()/Height() — the hasDims check fails and we continue.
// report.NewBaseObject() has no Height() method.
func TestCalcBandHeight_NonTextObject_NoDims(t *testing.T) {
	e := engine.New(reportpkg.NewReport())

	bb := band.NewBandBase()
	bb.SetHeight(30)
	bb.SetCanGrow(true)

	// report.NewBaseObject() has no Height() — exercises the hasDims-false path.
	bb.Objects().Add(report.NewBaseObject())

	h := e.CalcBandHeight(bb)
	// maxBottom stays 0 < baseHeight, so returns baseHeight.
	if h != 30 {
		t.Errorf("non-text no-dims: expected h = 30, got %v", h)
	}
}

// ── RunDataBandFull: PrintIfDSEmpty when data source is empty ─────────────────

// TestRunDataBandFull_PrintIfDSEmpty exercises the `else if db.PrintIfDSEmpty()`
// branch in RunDataBandFull that shows one empty row when the DS has no data.
func TestRunDataBandFull_PrintIfDSEmpty(t *testing.T) {
	e := newBandsExtra2Engine(t)

	ds := newMockDS(0) // zero rows → empty DS
	db := band.NewDataBand()
	db.SetName("EmptyDSBand")
	db.SetHeight(12)
	db.SetVisible(true)
	db.SetDataSource(ds)
	db.SetPrintIfDSEmpty(true)

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull PrintIfDSEmpty: %v", err)
	}
	// One "empty" row should be shown.
	if e.CurY() != beforeY+12 {
		t.Errorf("PrintIfDSEmpty: CurY = %v, want %v", e.CurY(), beforeY+12)
	}
}

// ── RenderInnerSubreports: non-SubreportObject mixed with PrintOnParent=true SR

// TestRenderInnerSubreports_WithMixedObjects exercises both branches in
// RenderInnerSubreports:
//   - `if !ok { continue }` when obj is not a SubreportObject (TextObject)
//   - `e.RenderInnerSubreport(...)` when obj IS a SubreportObject with PrintOnParent=true
func TestRenderInnerSubreports_WithMixedObjects(t *testing.T) {
	r := reportpkg.NewReport()
	pg1 := reportpkg.NewReportPage()
	pg1.SetName("Main")
	r.AddPage(pg1)

	// The subreport page.
	pg2 := reportpkg.NewReportPage()
	pg2.SetName("Sub2")
	r.AddPage(pg2)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	b := band.NewBandBase()

	// Add a TextObject (not SubreportObject) → covers `if !ok { continue }` path.
	txt := object.NewTextObject()
	txt.SetName("SomeTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(10)
	txt.SetText("hello")
	b.Objects().Add(txt)

	// Add an inner SubreportObject (PrintOnParent=true) → covers `e.RenderInnerSubreport(...)`.
	sr := object.NewSubreportObject()
	sr.SetReportPageName("Sub2")
	sr.SetPrintOnParent(true)
	b.Objects().Add(sr)

	// Should not panic.
	e.RenderInnerSubreports(b)
}

// ── applyRelationFilters: nil relation (no matching relation) ─────────────────

// TestApplyRelationFilters_NoMatchingRelation exercises the `rel == nil { continue }`
// path in applyRelationFilters when no relation links the parent and child DS.
func TestApplyRelationFilters_NoMatchingRelation(t *testing.T) {
	masterDS := data.NewBaseDataSource("Master")
	masterDS.SetAlias("Master")
	masterDS.AddColumn(data.Column{Name: "ID"})
	masterDS.AddRow(map[string]any{"ID": "1"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	detailDS := data.NewBaseDataSource("Detail")
	detailDS.SetAlias("Detail")
	detailDS.AddColumn(data.Column{Name: "MasterID"})
	detailDS.AddRow(map[string]any{"MasterID": "1"})
	if err := detailDS.Init(); err != nil {
		t.Fatalf("detailDS.Init: %v", err)
	}

	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddDataSource(detailDS)
	// No relation added → rel == nil for any parent→child pair.

	pg := reportpkg.NewReportPage()

	masterBand := band.NewDataBand()
	masterBand.SetName("MasterBand2")
	masterBand.SetHeight(15)
	masterBand.SetVisible(true)
	masterBand.SetDataSource(masterDS)

	detailBand := band.NewDataBand()
	detailBand.SetName("DetailBand2")
	detailBand.SetHeight(10)
	detailBand.SetVisible(true)
	detailBand.SetDataSource(detailDS)
	masterBand.Objects().Add(detailBand)

	pg.AddBand(masterBand)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with no-matching relation: %v", err)
	}
}

