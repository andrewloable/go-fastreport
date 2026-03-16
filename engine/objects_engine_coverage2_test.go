package engine_test

// objects_engine_coverage2_test.go — additional coverage tests for engine/objects.go,
// engine/subreports.go, engine/bands.go, engine/pages.go, and engine/relations.go.
//
// Uses package engine_test (external). Internal-only paths (dead code branches
// gated by defensive nil checks that can never be triggered through the public
// API) are documented below but are not tested, as they are structurally
// unreachable:
//
//   - populateContainerChildren: objs == nil (ContainerObject.Objects() never returns nil)
//   - populateTableObjects: colSpan < 1 (TableCell.SetColSpan clamps to 1)
//   - populateTableObjects: rowSpan < 1 (TableCell.SetRowSpan clamps to 1)
//   - renderGaugeBlob: png.Encode error (png.Encode on RGBA images never fails)
//   - evalTextWithFormat: CalcText returns an error (CalcText always returns nil error)
//   - applyRelationFilters: NewFilteredDataSource error (rebuildIndex never errors)
//   - CalcBandHeight: canShrink path (calcBandRequiredHeight always returns >= baseHeight)

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// newCoverage2Engine creates a minimal running engine for coverage tests.
func newCoverage2Engine(t *testing.T) *engine.ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("Cov2Page")
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return e
}

// ── subreports.go: RenderOuterSubreports — !ok (non-SubreportObject) branch ────

// TestRenderOuterSubreports_WithMixedObjectsNotOk covers the `if !ok { continue }`
// branch in RenderOuterSubreports (subreports.go line 77-78). When the band
// contains a non-SubreportObject (e.g. a TextObject), the type assertion to
// *object.SubreportObject fails and the object is skipped.
func TestRenderOuterSubreports_WithMixedObjectsNotOk(t *testing.T) {
	r := reportpkg.NewReport()
	pg1 := reportpkg.NewReportPage()
	pg1.SetName("MainOuter2")
	r.AddPage(pg1)

	pg2 := reportpkg.NewReportPage()
	pg2.SetName("OuterSub2")
	r.AddPage(pg2)

	// Add a band on the outer subreport page so CurY advances.
	ob := band.NewDataBand()
	ob.SetName("OuterSubBand2")
	ob.SetHeight(30)
	ob.SetVisible(true)
	pg2.AddBand(ob)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	parentBand := band.NewBandBase()

	// Add a TextObject first — NOT a SubreportObject.
	// When RenderOuterSubreports iterates, this triggers `if !ok { continue }`.
	txt := object.NewTextObject()
	txt.SetName("NonSRObj")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(80)
	txt.SetHeight(10)
	txt.SetText("skip me")
	txt.SetVisible(true)
	parentBand.Objects().Add(txt)

	// Add an outer subreport (PrintOnParent=false) that links to pg2.
	sr := object.NewSubreportObject()
	sr.SetReportPageName("OuterSub2")
	sr.SetPrintOnParent(false)
	sr.SetLeft(0)
	parentBand.Objects().Add(sr)

	startY := e.CurY()
	e.RenderOuterSubreports(parentBand)

	// CurY should have advanced (subreport rendered the 30px band).
	if e.CurY() <= startY {
		t.Errorf("RenderOuterSubreports: CurY should advance; start=%v after=%v", startY, e.CurY())
	}
}

// TestRenderOuterSubreports_PrintOnParentTrueSkipped covers the
// `if sr.PrintOnParent() { continue }` branch in RenderOuterSubreports
// (subreports.go line 80-82). When a SubreportObject is inner (PrintOnParent=true),
// it must be skipped by RenderOuterSubreports.
func TestRenderOuterSubreports_PrintOnParentTrueSkipped(t *testing.T) {
	e := newCoverage2Engine(t)

	parentBand := band.NewBandBase()

	// Add an inner subreport (PrintOnParent=true) — should be skipped.
	innerSR := object.NewSubreportObject()
	innerSR.SetReportPageName("NonExistent")
	innerSR.SetPrintOnParent(true) // inner → must be skipped by RenderOuterSubreports
	parentBand.Objects().Add(innerSR)

	beforeY := e.CurY()
	e.RenderOuterSubreports(parentBand)

	// No outer subreports → CurY unchanged and hasSubreports=false.
	if e.CurY() != beforeY {
		t.Errorf("RenderOuterSubreports inner-only: CurY should not change; before=%v after=%v",
			beforeY, e.CurY())
	}
}
