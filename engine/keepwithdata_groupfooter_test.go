package engine

// Internal tests for the getAllFooters fix that walks up the parent chain to
// include GroupFooterBands from ancestor GroupHeaderBands.
// Uses package engine (not engine_test) to access the unexported getAllFooters.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// newFooterEngine creates a minimal engine that has completed Run() so that
// internal state (preparedPages, freeSpace, etc.) is properly initialized.
func newFooterEngine(t *testing.T) *ReportEngine {
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

// TestGetAllFooters_NoFooter verifies that getAllFooters returns an empty slice
// when the DataBand has no DataFooterBand and no parent GroupHeaderBand.
func TestGetAllFooters_NoFooter(t *testing.T) {
	e := newFooterEngine(t)
	db := band.NewDataBand()
	db.SetName("DB")

	footers := e.getAllFooters(db)
	if len(footers) != 0 {
		t.Errorf("expected 0 footers, got %d", len(footers))
	}
}

// TestGetAllFooters_DataFooterOnly_KeepWithData verifies that getAllFooters
// returns the DataFooterBand when KeepWithData=true and there is no parent
// GroupHeaderBand.
func TestGetAllFooters_DataFooterOnly_KeepWithData(t *testing.T) {
	e := newFooterEngine(t)
	db := band.NewDataBand()
	db.SetName("DB")

	ftr := band.NewDataFooterBand()
	ftr.SetName("DBFooter")
	ftr.SetHeight(20)
	ftr.SetKeepWithData(true)
	db.SetFooter(ftr)

	footers := e.getAllFooters(db)
	if len(footers) != 1 {
		t.Fatalf("expected 1 footer, got %d", len(footers))
	}
	if !footers[0].keepWithData {
		t.Error("footer[0].keepWithData should be true")
	}
}

// TestGetAllFooters_DataFooterOnly_NoKeepWithData verifies that getAllFooters
// strips trailing footers without KeepWithData, so returns empty when the only
// footer has KeepWithData=false.
func TestGetAllFooters_DataFooterOnly_NoKeepWithData(t *testing.T) {
	e := newFooterEngine(t)
	db := band.NewDataBand()
	db.SetName("DB")

	ftr := band.NewDataFooterBand()
	ftr.SetName("DBFooter")
	ftr.SetHeight(20)
	ftr.SetKeepWithData(false)
	db.SetFooter(ftr)

	footers := e.getAllFooters(db)
	if len(footers) != 0 {
		t.Errorf("expected 0 footers (stripped trailing non-keepwithdata), got %d", len(footers))
	}
}

// TestGetAllFooters_GroupFooter_LastRow verifies that getAllFooters includes the
// GroupFooterBand from the parent GroupHeaderBand when IsLastRow=true.
func TestGetAllFooters_GroupFooter_LastRow(t *testing.T) {
	e := newFooterEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH")
	gh.SetIsLastRow(true) // last group instance

	gftr := band.NewGroupFooterBand()
	gftr.SetName("GFooter")
	gftr.SetHeight(15)
	gftr.SetKeepWithData(true)
	gh.SetGroupFooter(gftr)

	db := band.NewDataBand()
	db.SetName("DB")

	// Simulate engine state: groupStack holds the active GroupHeaderBand.
	e.groupStack = []*band.GroupHeaderBand{gh}

	footers := e.getAllFooters(db)
	if len(footers) != 1 {
		t.Fatalf("expected 1 footer (GroupFooter), got %d", len(footers))
	}
	if !footers[0].keepWithData {
		t.Error("footer[0].keepWithData should be true")
	}
}

// TestGetAllFooters_GroupFooter_NotLastRow verifies that getAllFooters does NOT
// include the GroupFooterBand from the parent GroupHeaderBand when IsLastRow=false.
func TestGetAllFooters_GroupFooter_NotLastRow(t *testing.T) {
	e := newFooterEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH")
	gh.SetIsLastRow(false) // NOT the last row of this group

	gftr := band.NewGroupFooterBand()
	gftr.SetName("GFooter")
	gftr.SetHeight(15)
	gftr.SetKeepWithData(true)
	gh.SetGroupFooter(gftr)

	db := band.NewDataBand()
	db.SetName("DB")
	e.groupStack = []*band.GroupHeaderBand{gh}

	footers := e.getAllFooters(db)
	if len(footers) != 0 {
		t.Errorf("expected 0 footers when group IsLastRow=false, got %d", len(footers))
	}
}

// TestGetAllFooters_DataFooterAndGroupFooter_BothLastRow verifies that when
// both DataFooterBand.KeepWithData=true and the parent GroupFooterBand.KeepWithData=true
// and the group IsLastRow=true, both footers are returned.
func TestGetAllFooters_DataFooterAndGroupFooter_BothLastRow(t *testing.T) {
	e := newFooterEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH")
	gh.SetIsLastRow(true)

	gftr := band.NewGroupFooterBand()
	gftr.SetName("GFooter")
	gftr.SetHeight(15)
	gftr.SetKeepWithData(true)
	gh.SetGroupFooter(gftr)

	db := band.NewDataBand()
	db.SetName("DB")

	dftr := band.NewDataFooterBand()
	dftr.SetName("DBFooter")
	dftr.SetHeight(10)
	dftr.SetKeepWithData(true)
	db.SetFooter(dftr)

	e.groupStack = []*band.GroupHeaderBand{gh}

	footers := e.getAllFooters(db)
	if len(footers) != 2 {
		t.Fatalf("expected 2 footers (DataFooter + GroupFooter), got %d", len(footers))
	}
	if !footers[0].keepWithData {
		t.Error("footer[0] (DataFooter) keepWithData should be true")
	}
	if !footers[1].keepWithData {
		t.Error("footer[1] (GroupFooter) keepWithData should be true")
	}
}

// TestGetAllFooters_NestedGroups_BothLastRow verifies that getAllFooters walks
// through two nested GroupHeaderBands when both are on their last row.
func TestGetAllFooters_NestedGroups_BothLastRow(t *testing.T) {
	e := newFooterEngine(t)

	// Outer group
	outerGH := band.NewGroupHeaderBand()
	outerGH.SetName("OuterGH")
	outerGH.SetIsLastRow(true)

	outerGFtr := band.NewGroupFooterBand()
	outerGFtr.SetName("OuterGFooter")
	outerGFtr.SetHeight(20)
	outerGFtr.SetKeepWithData(true)
	outerGH.SetGroupFooter(outerGFtr)

	// Inner group
	innerGH := band.NewGroupHeaderBand()
	innerGH.SetName("InnerGH")
	innerGH.SetIsLastRow(true)

	innerGFtr := band.NewGroupFooterBand()
	innerGFtr.SetName("InnerGFooter")
	innerGFtr.SetHeight(15)
	innerGFtr.SetKeepWithData(true)
	innerGH.SetGroupFooter(innerGFtr)

	db := band.NewDataBand()
	db.SetName("DB")

	// groupStack: innermost first, then outer
	e.groupStack = []*band.GroupHeaderBand{innerGH, outerGH}

	footers := e.getAllFooters(db)
	// Should have: InnerGroupFooter, OuterGroupFooter
	if len(footers) != 2 {
		t.Fatalf("expected 2 footers (inner+outer GroupFooter), got %d", len(footers))
	}
}

// TestGetAllFooters_NestedGroups_OuterNotLastRow verifies that walking stops at
// an outer GroupHeaderBand that is not on its last row.
func TestGetAllFooters_NestedGroups_OuterNotLastRow(t *testing.T) {
	e := newFooterEngine(t)

	// Outer group is NOT on its last row — should stop here.
	outerGH := band.NewGroupHeaderBand()
	outerGH.SetName("OuterGH")
	outerGH.SetIsLastRow(false)

	outerGFtr := band.NewGroupFooterBand()
	outerGFtr.SetName("OuterGFooter")
	outerGFtr.SetHeight(20)
	outerGFtr.SetKeepWithData(true)
	outerGH.SetGroupFooter(outerGFtr)

	// Inner group is on its last row
	innerGH := band.NewGroupHeaderBand()
	innerGH.SetName("InnerGH")
	innerGH.SetIsLastRow(true)

	innerGFtr := band.NewGroupFooterBand()
	innerGFtr.SetName("InnerGFooter")
	innerGFtr.SetHeight(15)
	innerGFtr.SetKeepWithData(true)
	innerGH.SetGroupFooter(innerGFtr)

	db := band.NewDataBand()
	db.SetName("DB")

	// groupStack: innermost first, then outer (not last row)
	e.groupStack = []*band.GroupHeaderBand{innerGH, outerGH}

	footers := e.getAllFooters(db)
	// Only the inner group footer (outer is not last row, so we stop before it).
	if len(footers) != 1 {
		t.Fatalf("expected 1 footer (inner GroupFooter only), got %d", len(footers))
	}
}

// TestGetAllFooters_GroupFooterNilNoGroupFooter verifies that a GroupHeaderBand
// with no GroupFooterBand set contributes nothing to the footer list.
func TestGetAllFooters_GroupFooterNilNoGroupFooter(t *testing.T) {
	e := newFooterEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH")
	gh.SetIsLastRow(true)
	// No GroupFooter set — gh.GroupFooter() is nil.

	db := band.NewDataBand()
	db.SetName("DB")
	e.groupStack = []*band.GroupHeaderBand{gh}

	footers := e.getAllFooters(db)
	if len(footers) != 0 {
		t.Errorf("expected 0 footers when GroupFooter is nil, got %d", len(footers))
	}
}

// TestGetAllFooters_StripGroupFooter_NoKeepWithData verifies that a GroupFooterBand
// with KeepWithData=false is stripped from the trailing footers list.
func TestGetAllFooters_StripGroupFooter_NoKeepWithData(t *testing.T) {
	e := newFooterEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH")
	gh.SetIsLastRow(true)

	gftr := band.NewGroupFooterBand()
	gftr.SetName("GFooter")
	gftr.SetHeight(15)
	gftr.SetKeepWithData(false) // no KeepWithData — should be stripped
	gh.SetGroupFooter(gftr)

	db := band.NewDataBand()
	db.SetName("DB")
	e.groupStack = []*band.GroupHeaderBand{gh}

	footers := e.getAllFooters(db)
	if len(footers) != 0 {
		t.Errorf("expected 0 footers (GroupFooter stripped, no KeepWithData), got %d", len(footers))
	}
}

// TestGetAllFooters_ParentIsNotGroupHeader verifies that walking stops immediately
// when the DataBand's parent is not a GroupHeaderBand (e.g. it's a page or nil).
func TestGetAllFooters_ParentIsNotGroupHeader(t *testing.T) {
	e := newFooterEngine(t)

	// DataBand with no parent (parent is nil — Parent() returns nil).
	db := band.NewDataBand()
	db.SetName("DB")

	dftr := band.NewDataFooterBand()
	dftr.SetName("DBFooter")
	dftr.SetHeight(10)
	dftr.SetKeepWithData(true)
	db.SetFooter(dftr)

	footers := e.getAllFooters(db)
	// Only the DataFooterBand; no group footer walk happens.
	if len(footers) != 1 {
		t.Fatalf("expected 1 footer (DataFooter only), got %d", len(footers))
	}
}

// TestNeedKeepLastRow_WithGroupFooter verifies that NeedKeepLastRow returns true
// when the parent GroupHeaderBand has a GroupFooterBand with KeepWithData=true
// and the group is on its last row.
func TestNeedKeepLastRow_WithGroupFooter(t *testing.T) {
	e := newFooterEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH")
	gh.SetIsLastRow(true)

	gftr := band.NewGroupFooterBand()
	gftr.SetName("GFooter")
	gftr.SetHeight(15)
	gftr.SetKeepWithData(true)
	gh.SetGroupFooter(gftr)

	db := band.NewDataBand()
	db.SetName("DB")
	e.groupStack = []*band.GroupHeaderBand{gh}

	if !e.NeedKeepLastRow(db) {
		t.Error("NeedKeepLastRow should be true when parent GroupFooter has KeepWithData=true and IsLastRow=true")
	}
}

// TestGetFootersHeight_IncludesGroupFooter verifies that GetFootersHeight includes
// the height of the GroupFooterBand when it qualifies.
func TestGetFootersHeight_IncludesGroupFooter(t *testing.T) {
	e := newFooterEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH")
	gh.SetIsLastRow(true)

	gftr := band.NewGroupFooterBand()
	gftr.SetName("GFooter")
	gftr.SetHeight(15)
	gftr.SetKeepWithData(true)
	gh.SetGroupFooter(gftr)

	db := band.NewDataBand()
	db.SetName("DB")
	e.groupStack = []*band.GroupHeaderBand{gh}

	h := e.GetFootersHeight(db)
	// Height should be at least 15 (the GroupFooter height).
	if h < 15 {
		t.Errorf("GetFootersHeight = %v, expected >= 15 (GroupFooter height)", h)
	}
}
