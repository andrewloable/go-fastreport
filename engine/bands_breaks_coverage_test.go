package engine_test

// bands_breaks_coverage_test.go — targeted coverage for uncovered branches in
// bands.go and breaks.go: splitPopulateTop, splitPopulateBottom, BreakBand
// with objects, AddBandToPreparedPages overflow, ShowFullBand outline, and
// ShowDataBandRow StartNewPage.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── BreakBand with objects (exercises splitPopulateTop / splitPopulateBottom) ─

// newEngineForBreaks creates a running engine and consumes all but `remaining`
// pixels of free space, forcing BreakBand to split at that line.
func newEngineForBreaks(t *testing.T, remaining float32) *engine.ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	toConsume := e.FreeSpace() - remaining
	if toConsume > 0 {
		e.AdvanceY(toConsume)
	}
	return e
}

// TestBreakBand_WithObjectsAboveBreakLine exercises splitPopulateTop for an
// object that is entirely above the break line.
func TestBreakBand_WithObjectsAboveBreakLine(t *testing.T) {
	// Leave 50px free. We'll add a band of 200px height with an object at top=10.
	e := newEngineForBreaks(t, 50)

	b := band.NewBandBase()
	b.SetName("BrkObjAbove")
	b.SetHeight(200)
	b.SetVisible(true)
	b.SetCanBreak(true)

	// Object entirely above the break line (top=5, height=20 → bottom=25 < 50).
	txt := object.NewTextObject()
	txt.SetName("TxtAbove")
	txt.SetLeft(0)
	txt.SetTop(5)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetText("above")
	txt.SetVisible(true)
	b.Objects().Add(txt)

	pp := e.PreparedPages()
	pgCount := pp.Count()

	e.BreakBand(b)

	// A new page should have been started.
	if pp.Count() <= pgCount {
		t.Error("BreakBand with object above: expected new page")
	}
}

// TestBreakBand_WithObjectsBelowBreakLine exercises splitPopulateBottom for an
// object that is entirely below the break line.
func TestBreakBand_WithObjectsBelowBreakLine(t *testing.T) {
	e := newEngineForBreaks(t, 50)

	b := band.NewBandBase()
	b.SetName("BrkObjBelow")
	b.SetHeight(200)
	b.SetVisible(true)
	b.SetCanBreak(true)

	// Object entirely below the break line (top=80, height=20 → bottom=100 > 50).
	txt := object.NewTextObject()
	txt.SetName("TxtBelow")
	txt.SetLeft(0)
	txt.SetTop(80)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetText("below")
	txt.SetVisible(true)
	b.Objects().Add(txt)

	e.BreakBand(b)
	// Should not panic; bottom object ends up in the remainder portion.
}

// TestBreakBand_WithObjectStraddlingBreakLine exercises both splitPopulateTop
// and splitPopulateBottom for an object that crosses the break line.
func TestBreakBand_WithObjectStraddlingBreakLine(t *testing.T) {
	e := newEngineForBreaks(t, 50)

	b := band.NewBandBase()
	b.SetName("BrkObjStraddle")
	b.SetHeight(200)
	b.SetVisible(true)
	b.SetCanBreak(true)

	// Object that straddles the break line (top=30, height=40 → bottom=70; breakLine≈50).
	txt := object.NewTextObject()
	txt.SetName("TxtStraddle")
	txt.SetLeft(0)
	txt.SetTop(30)
	txt.SetWidth(100)
	txt.SetHeight(40)
	txt.SetText("straddles")
	txt.SetVisible(true)
	b.Objects().Add(txt)

	e.BreakBand(b)
	// Should not panic; object split between top/bottom portions.
}

// TestBreakBand_NonBreakableObjectPullsBreakLine exercises the loop in BreakBand
// that pulls breakLine down when a non-breakable object crosses it.
func TestBreakBand_NonBreakableObjectPullsBreakLine(t *testing.T) {
	e := newEngineForBreaks(t, 60)

	b := band.NewBandBase()
	b.SetName("BrkPullLine")
	b.SetHeight(200)
	b.SetVisible(true)
	b.SetCanBreak(true)

	// Non-breakable sub-band that crosses the break line (top=40, height=40 → bottom=80).
	// BandBase.CanBreak defaults to false, so this will pull the break line down.
	inner := band.NewBandBase()
	inner.SetTop(40)
	inner.SetHeight(40)
	inner.SetCanBreak(false)
	b.Objects().Add(inner)

	pp := e.PreparedPages()
	pgBefore := pp.Count()

	e.BreakBand(b)

	if pp.Count() <= pgBefore {
		t.Error("non-breakable object: expected new page after BreakBand")
	}
}

// TestBreakBand_WithMultipleObjects exercises splitPopulateTop/Bottom with
// multiple objects in various positions.
func TestBreakBand_WithMultipleObjects(t *testing.T) {
	e := newEngineForBreaks(t, 80)

	b := band.NewBandBase()
	b.SetName("BrkMulti")
	b.SetHeight(300)
	b.SetVisible(true)
	b.SetCanBreak(true)

	// Object 1: entirely above break line (top=10, bottom=30).
	txt1 := object.NewTextObject()
	txt1.SetTop(10)
	txt1.SetHeight(20)
	txt1.SetWidth(100)
	txt1.SetText("obj1")
	b.Objects().Add(txt1)

	// Object 2: straddles break line (top=60, bottom=100; breakLine≈80).
	txt2 := object.NewTextObject()
	txt2.SetTop(60)
	txt2.SetHeight(40)
	txt2.SetWidth(100)
	txt2.SetText("obj2")
	b.Objects().Add(txt2)

	// Object 3: entirely below break line (top=120, bottom=150).
	txt3 := object.NewTextObject()
	txt3.SetTop(120)
	txt3.SetHeight(30)
	txt3.SetWidth(100)
	txt3.SetText("obj3")
	b.Objects().Add(txt3)

	e.BreakBand(b)
	// Should not panic.
}

// ── AddBandToPreparedPages: FlagCheckFreeSpace + overflow triggers new page ──

// TestAddBandToPreparedPages_CheckFreeSpaceOverflow exercises the branch where
// FlagCheckFreeSpace is true, the band doesn't fit, and CanBreak is false →
// a new page is started (recursive call with FlagMustBreak=true).
func TestAddBandToPreparedPages_CheckFreeSpaceOverflow(t *testing.T) {
	e := newEngineForBreaks(t, 10)

	db := band.NewDataBand()
	db.SetName("OverflowBand")
	db.SetHeight(100) // larger than remaining 10px
	db.SetVisible(true)
	db.BandBase.FlagCheckFreeSpace = true
	db.BandBase.SetCanBreak(false)

	pgBefore := e.PreparedPages().Count()
	ok := e.AddBandToPreparedPages(&db.BandBase)
	if !ok {
		t.Error("AddBandToPreparedPages overflow: expected true (band added on new page)")
	}
	if e.PreparedPages().Count() <= pgBefore {
		t.Error("AddBandToPreparedPages overflow: expected new page to be created")
	}
}

// TestAddBandToPreparedPages_CheckFreeSpaceCanBreak exercises the CanBreak=true
// branch where the band doesn't fit but can break.
func TestAddBandToPreparedPages_CheckFreeSpaceCanBreak(t *testing.T) {
	e := newEngineForBreaks(t, 10)

	db := band.NewDataBand()
	db.SetName("CanBreakBand")
	db.SetHeight(100)
	db.SetVisible(true)
	db.BandBase.FlagCheckFreeSpace = true
	db.BandBase.SetCanBreak(true)

	ok := e.AddBandToPreparedPages(&db.BandBase)
	if !ok {
		t.Error("AddBandToPreparedPages CanBreak: expected true")
	}
}

// ── ShowFullBand: outline expression ──────────────────────────────────────────

// TestShowFullBand_WithOutlineExpression exercises the OutlineExpression branch
// in showFullBandOnce that adds an outline entry then calls OutlineUp.
func TestShowFullBand_WithOutlineExpression(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	b := band.NewBandBase()
	b.SetName("OutlineBand")
	b.SetHeight(20)
	b.SetVisible(true)
	b.SetOutlineExpression("Section 1")

	e.ShowFullBand(b)
	// Verify outline was added.
	root := e.PreparedPages().Outline.Root
	if len(root.Children) == 0 {
		t.Error("ShowFullBand with OutlineExpression: expected outline entry")
	}
}

// TestShowFullBand_ZeroHeightWithOutline exercises the zero-height path where
// the outline entry is added but then OutlineUp is called without adding to pages.
func TestShowFullBand_ZeroHeightWithOutline(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	b := band.NewBandBase()
	b.SetName("ZeroOutlineBand")
	b.SetHeight(0) // zero height → early return after outline
	b.SetVisible(true)
	b.SetOutlineExpression("Zero Height Section")

	beforeY := e.CurY()
	e.ShowFullBand(b)
	if e.CurY() != beforeY {
		t.Errorf("zero height band should not advance CurY: got %v, want %v", e.CurY(), beforeY)
	}
}

// ── ShowDataBandRow: StartNewPage branch ─────────────────────────────────────

// TestShowDataBandRow_StartNewPage exercises the StartNewPage branch in
// ShowDataBandRow when rowNo > 1.
func TestShowDataBandRow_StartNewPage(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	db := band.NewDataBand()
	db.SetName("SNPBand")
	db.SetHeight(20)
	db.SetVisible(true)
	db.SetStartNewPage(true)
	db.FlagUseStartNewPage = true

	pgBefore := e.PreparedPages().Count()
	// rowNo=1: no new page.
	e.ShowDataBandRow(db, 1, 1)
	if e.PreparedPages().Count() != pgBefore {
		t.Error("ShowDataBandRow rowNo=1 should not start new page")
	}
	// rowNo=2: triggers StartNewPage.
	e.ShowDataBandRow(db, 2, 2)
	if e.PreparedPages().Count() <= pgBefore {
		t.Error("ShowDataBandRow rowNo=2 should start new page")
	}
}

// ── SplitHardPageBreaks: multiple breaks ─────────────────────────────────────

// TestSplitHardPageBreaks_MultipleBreaks exercises SplitHardPageBreaks with 2
// hard page breaks, verifying 3 parts are returned.
func TestSplitHardPageBreaks_MultipleBreaks(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	b := band.NewBandBase()
	b.SetHeight(150)
	b.SetName("MultiBreak")

	rc1 := report.NewReportComponentBase()
	rc1.SetTop(40)
	rc1.SetPageBreak(true)
	b.AddChild(rc1)

	rc2 := report.NewReportComponentBase()
	rc2.SetTop(90)
	rc2.SetPageBreak(true)
	b.AddChild(rc2)

	parts := e.SplitHardPageBreaks(b)
	if len(parts) != 3 {
		t.Fatalf("2 breaks: expected 3 parts, got %d", len(parts))
	}
	if parts[0].Height() != 40 {
		t.Errorf("part[0] height = %v, want 40", parts[0].Height())
	}
	if parts[1].Height() != 50 {
		t.Errorf("part[1] height = %v, want 50 (90-40)", parts[1].Height())
	}
	if parts[2].Height() != 60 {
		t.Errorf("part[2] height = %v, want 60 (150-90)", parts[2].Height())
	}
	if !parts[1].StartNewPage() {
		t.Error("part[1] should have StartNewPage=true")
	}
	if !parts[2].StartNewPage() {
		t.Error("part[2] should have StartNewPage=true")
	}
}

// ── CalcBandHeight: non-BandBase type with Height() ──────────────────────────

// TestCalcBandHeight_NonBandBase exercises the fallback path in CalcBandHeight
// for objects that are not *band.BandBase but implement Height().
func TestCalcBandHeight_NonBandBase(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	// DataBand is NOT *band.BandBase directly (it embeds it).
	// GroupHeaderBand embeds HeaderFooterBandBase which embeds BandBase.
	// Use an object.TextObject which has Height() but is not *band.BandBase.
	txt := object.NewTextObject()
	txt.SetHeight(25)
	h := e.CalcBandHeight(txt)
	if h != 25 {
		t.Errorf("CalcBandHeight non-BandBase: got %v, want 25", h)
	}
}

// ── showFullBandOnce: FlagCheckFreeSpace + not enough space ──────────────────

// TestShowFullBandOnce_FlagCheckFreeSpace_NewPage exercises the branch in
// showFullBandOnce where FlagCheckFreeSpace is true and freeSpace < height.
func TestShowFullBandOnce_FlagCheckFreeSpace_NewPage(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Consume almost all free space leaving only 5px.
	e.AdvanceY(e.FreeSpace() - 5)

	b := band.NewBandBase()
	b.SetName("CheckFreeBand")
	b.SetHeight(50) // larger than remaining 5px
	b.SetVisible(true)
	b.FlagCheckFreeSpace = true // triggers page break in showFullBandOnce

	pgBefore := e.PreparedPages().Count()
	e.ShowFullBand(b)
	if e.PreparedPages().Count() <= pgBefore {
		t.Error("ShowFullBand with FlagCheckFreeSpace: expected new page")
	}
}
