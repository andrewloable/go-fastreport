package engine_test

// outline_porting_gaps_test.go — tests for outline porting-gap fixes:
//   1. !band.Repeated guard: repeated bands must NOT add duplicate outline entries.
//   2. DataBand OutlineUp: only called when DataBand has an OutlineExpression.
//   3. GroupHeaderBand OutlineUp: only called in showGroupFooter when GroupHeaderBand
//      has an OutlineExpression.
//   4. Non-DataBand/non-GroupHeaderBand bands still call OutlineUp immediately.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// outlineCountRootChildren returns the number of top-level outline children.
func outlineCountRootChildren(pp *preview.PreparedPages) int {
	return len(pp.Outline.Root.Children)
}

// newOutlineGapEngine builds an engine ready for post-run outline testing.
func newOutlineGapEngine(t *testing.T) *engine.ReportEngine {
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

// ── Gap 1: !band.Repeated guard ───────────────────────────────────────────────

// TestShowFullBand_Repeated_NoOutline verifies that when a band is marked
// Repeated (reprinted on a new page) it does NOT add a duplicate outline entry.
// C# AddBandOutline: if (band.Visible && !IsNullOrEmpty(band.OutlineExpression)
//                         && !band.Repeated) { AddOutline(...) }
func TestShowFullBand_Repeated_NoOutline(t *testing.T) {
	e := newOutlineGapEngine(t)
	pp := e.PreparedPages()

	b := band.NewReportTitleBand()
	b.SetName("Title")
	b.SetVisible(true)
	b.SetHeight(20)
	b.SetOutlineExpression("Title") // literal, no brackets needed for evalText

	// First show: not repeated — should add an outline entry and immediately
	// call OutlineUp (ReportTitleBand is not DataBand or GroupHeaderBand).
	e.ShowFullBand(&b.BandBase)
	afterFirst := outlineCountRootChildren(pp)
	if afterFirst != 1 {
		t.Fatalf("after first ShowFullBand: expected 1 root outline child, got %d", afterFirst)
	}

	// Mark as repeated and show again — must NOT add another outline entry.
	b.SetRepeated(true)
	e.ShowFullBand(&b.BandBase)
	afterSecond := outlineCountRootChildren(pp)
	if afterSecond != 1 {
		t.Errorf("after repeated ShowFullBand: expected still 1 root outline child, got %d (duplicate added)", afterSecond)
	}
}

// ── Gap 2: DataBand OutlineUp only when OutlineExpression set ─────────────────

// TestDataBand_WithOutlineExpression_OutlineUp verifies that when a DataBand
// has an OutlineExpression, each rendered row descends into the outline then
// levels up, resulting in sibling entries rather than a nested chain.
// C# OutlineUp(dataBand): if (!IsNullOrEmpty(dataBand.OutlineExpression)) OutlineUp()
func TestDataBand_WithOutlineExpression_OutlineUp(t *testing.T) {
	e := newOutlineGapEngine(t)
	pp := e.PreparedPages()

	db := band.NewDataBand()
	db.SetName("Data")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetOutlineExpression("Row") // evaluated as literal "Row" by evalText

	// Run 2 rows — each row should add an outline entry then level up.
	e.RunDataBandRows(db, 2)

	// Both entries should be at the root level (siblings), not nested.
	rootCount := outlineCountRootChildren(pp)
	if rootCount != 2 {
		t.Errorf("DataBand with OutlineExpression, 2 rows: expected 2 root outline children, got %d", rootCount)
	}
	// Neither child should itself have children (OutlineUp was called per row).
	for i, child := range pp.Outline.Root.Children {
		if len(child.Children) != 0 {
			t.Errorf("root outline child[%d] has %d sub-children, want 0 (OutlineUp not called)", i, len(child.Children))
		}
	}
}

// TestDataBand_NoOutlineExpression_NoOutlineUp verifies that when a DataBand
// has no OutlineExpression, no outline entries are added and the outline level
// is not modified (no spurious OutlineUp).
func TestDataBand_NoOutlineExpression_NoOutlineUp(t *testing.T) {
	e := newOutlineGapEngine(t)
	pp := e.PreparedPages()

	db := band.NewDataBand()
	db.SetName("DataNoOutline")
	db.SetVisible(true)
	db.SetHeight(10)
	// OutlineExpression intentionally not set.

	e.RunDataBandRows(db, 3)

	rootCount := outlineCountRootChildren(pp)
	if rootCount != 0 {
		t.Errorf("DataBand without OutlineExpression: expected 0 outline entries, got %d", rootCount)
	}
}

// ── Gap 3: GroupHeaderBand OutlineUp only when OutlineExpression set ──────────

// TestRunGroup_WithOutlineExpression_OutlineUp verifies that when a
// GroupHeaderBand has an OutlineExpression, showGroupFooter calls OutlineUp so
// the next group becomes a sibling rather than a child.
// C# OutlineUp(header): if (!IsNullOrEmpty(header.OutlineExpression)) OutlineUp()
func TestRunGroup_WithOutlineExpression_OutlineUp(t *testing.T) {
	e := newOutlineGapEngine(t)
	pp := e.PreparedPages()

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH")
	gh.SetVisible(true)
	gh.SetHeight(20)
	gh.SetCondition("val")
	gh.SetOutlineExpression("Group") // literal; each group adds one outline entry

	gf := band.NewGroupFooterBand()
	gf.SetName("GF")
	gf.SetVisible(true)
	gf.SetHeight(5)
	gh.SetGroupFooter(gf)

	db := band.NewDataBand()
	db.SetName("DB")
	db.SetVisible(true)
	db.SetHeight(10)
	// Two distinct group values so we get two separate group headers rendered.
	db.SetDataSource(newSimpleDS("A", "B"))
	gh.SetData(db)

	e.RunGroup(gh)

	// Each group header added an outline entry. showGroupFooter should have
	// called OutlineUp (because OutlineExpression is set), so the two entries
	// are siblings at the root level rather than parent/child.
	rootCount := outlineCountRootChildren(pp)
	if rootCount != 2 {
		t.Errorf("GroupHeaderBand with OutlineExpression, 2 groups: expected 2 root outline children, got %d", rootCount)
	}
	for i, child := range pp.Outline.Root.Children {
		if len(child.Children) != 0 {
			t.Errorf("root outline child[%d] has %d sub-children, want 0 (OutlineUp not called by showGroupFooter)", i, len(child.Children))
		}
	}
}

// TestRunGroup_NoOutlineExpression_NoOutlineUp verifies that when a
// GroupHeaderBand has no OutlineExpression, no outline entries are created and
// the outline level is unchanged (no spurious OutlineUp from showGroupFooter).
func TestRunGroup_NoOutlineExpression_NoOutlineUp(t *testing.T) {
	e := newOutlineGapEngine(t)
	pp := e.PreparedPages()

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH2")
	gh.SetVisible(true)
	gh.SetHeight(15)
	gh.SetCondition("val")
	// OutlineExpression intentionally not set.

	gf := band.NewGroupFooterBand()
	gf.SetName("GF2")
	gf.SetVisible(true)
	gf.SetHeight(5)
	gh.SetGroupFooter(gf)

	db := band.NewDataBand()
	db.SetName("DB2")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newSimpleDS("X", "Y"))
	gh.SetData(db)

	e.RunGroup(gh)

	rootCount := outlineCountRootChildren(pp)
	if rootCount != 0 {
		t.Errorf("GroupHeaderBand without OutlineExpression: expected 0 outline entries, got %d", rootCount)
	}
}

// ── Gap 4: Non-DataBand/non-GroupHeader bands still call OutlineUp immediately ─

// TestShowFullBand_NonDataBand_OutlineUpImmediate verifies that for bands that
// are neither DataBand nor GroupHeaderBand (e.g. ReportTitleBand), ShowFullBand
// immediately calls OutlineUp after adding the outline entry so that the next
// band added at the same level is a sibling, not a child.
// C# AddBandOutline: if (!(band is DataBand) && !(band is GroupHeaderBand)) OutlineUp()
func TestShowFullBand_NonDataBand_OutlineUpImmediate(t *testing.T) {
	e := newOutlineGapEngine(t)
	pp := e.PreparedPages()

	b1 := band.NewReportTitleBand()
	b1.SetName("Title1")
	b1.SetVisible(true)
	b1.SetHeight(20)
	b1.SetOutlineExpression("Section1")

	b2 := band.NewReportTitleBand()
	b2.SetName("Title2")
	b2.SetVisible(true)
	b2.SetHeight(20)
	b2.SetOutlineExpression("Section2")

	e.ShowFullBand(&b1.BandBase)
	e.ShowFullBand(&b2.BandBase)

	// Both sections should be root-level siblings (OutlineUp was called after b1).
	rootCount := outlineCountRootChildren(pp)
	if rootCount != 2 {
		t.Errorf("two non-DataBand bands: expected 2 root outline siblings, got %d", rootCount)
	}
}
