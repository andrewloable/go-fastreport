package engine

// groups_coverage_gaps_internal_test.go — internal (package engine) tests to
// cover the remaining branches in groups.go that cannot be reached from
// external tests:
//
//  1. groupTreeItem.firstItem() → nil  (empty items slice)
//  2. groupTreeItem.lastItem()  → nil  (empty items slice)
//  3. makeGroupTree orphan path: checkGroupItem where last == nil
//     (curItem has no children when NestedGroup is encountered mid-scan)
//  4. showDataHeader: db == nil early-return (GroupHeaderBand with no DataBand)
//  5. showDataFooter: db == nil early-return (GroupHeaderBand with no DataBand)

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── 1 & 2: firstItem / lastItem on empty tree ─────────────────────────────────

func TestGroupTreeItem_FirstItem_EmptyItems(t *testing.T) {
	g := &groupTreeItem{}
	if g.firstItem() != nil {
		t.Error("firstItem on empty tree should return nil")
	}
}

func TestGroupTreeItem_LastItem_EmptyItems(t *testing.T) {
	g := &groupTreeItem{}
	if g.lastItem() != nil {
		t.Error("lastItem on empty tree should return nil")
	}
}

// ── 3: makeGroupTree orphan — checkGroupItem last==nil ───────────────────────
//
// The "orphan" branch fires when checkGroupItem walks to a nested group level
// but curItem has no children yet (last == nil).  This happens when the outer
// group condition never changes (so we always stay in checkGroupItem) but the
// inner group condition changes — at that point curItem is the root and has
// only just received its first child, so last starts as nil on the first
// checkGroupItem call for the inner group.
//
// Concretely: outer has no condition (always same value ""→""), inner has
// condition "val".  On the second row the outer value is still the same so
// checkGroupItem is called.  At the outer level last = root.lastItem() which
// is the first child created during initGroupItem for row 0.  At the inner
// level last = child.lastItem() which is nil because initGroupItem only added
// one child to root, not to root's child.  So `if last != nil` is false.

func TestMakeGroupTree_CheckGroupItem_LastNil(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Inner group band — condition on "val".
	innerGH := band.NewGroupHeaderBand()
	innerGH.SetName("InnerGH_OrphanTest")
	innerGH.SetVisible(true)
	innerGH.SetHeight(8)
	innerGH.SetCondition("val")

	// Outer group band — no condition (always empty string = same value).
	// This means on every row after the first, checkGroupItem is called and
	// the outer group value never changes, so we walk to the inner level.
	outerGH := band.NewGroupHeaderBand()
	outerGH.SetName("OuterGH_OrphanTest")
	outerGH.SetVisible(true)
	outerGH.SetHeight(10)
	// No condition set — groupConditionValue returns "" always.
	outerGH.SetNestedGroup(innerGH)

	db := band.NewDataBand()
	db.SetName("DB_OrphanTest")
	db.SetVisible(true)
	db.SetHeight(10)
	// Two distinct inner-group values: row0="X", row1="Y".
	// On row1 the outer value is still "" (same), so checkGroupItem is called.
	// At the inner level, curItem (the outer's first child from row0) has no
	// children yet → last == nil → exercises the "last == nil" branch.
	ds := &groupsAbortStringDS{rows: []string{"X", "Y"}}
	db.SetDataSource(ds)
	innerGH.SetData(db)
	outerGH.SetData(db)

	// Must not panic.
	tree := e.makeGroupTree(outerGH)
	if tree == nil {
		t.Error("makeGroupTree should return a non-nil root")
	}
}

// ── 4: showDataHeader db==nil early return ───────────────────────────────────
//
// showDataHeader is called with the first child's GroupHeaderBand.  If that
// band has no DataBand set (Data()==nil), the function must return without
// panicking.  We exercise this by building a tree node whose band.Data() is
// nil and calling showDataHeader directly.

func TestShowDataHeader_NilDataBand(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_NilData")
	gh.SetVisible(true)
	gh.SetHeight(10)
	// No Data set → gh.Data() returns nil → showDataHeader returns early.

	// Must not panic.
	e.showDataHeader(gh)
}

// ── 5: showDataFooter db==nil early return ───────────────────────────────────

func TestShowDataFooter_NilDataBand(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_NilDataFtr")
	gh.SetVisible(true)
	gh.SetHeight(10)
	// No Data set → gh.Data() returns nil → showDataFooter returns early.

	// Must not panic.
	e.showDataFooter(gh)
}

// ── showGroupTree: leaf node with band != nil but rowCount == 0 ──────────────
//
// The leaf path in showGroupTree reads:
//   if root.band != nil && root.rowCount > 0 { ... }
// When rowCount == 0 the RunDataBandRows call is skipped.  This exercises
// the false branch of that guard.

func TestShowGroupTree_LeafBandNoRows(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_LeafNoRows")
	gh.SetVisible(true)
	gh.SetHeight(10)

	db := band.NewDataBand()
	db.SetName("DB_LeafNoRows")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(&groupsAbortStringDS{rows: []string{"A"}})
	gh.SetData(db)

	// Leaf node with rowCount=0 — exercises the `root.rowCount > 0` guard.
	leaf := &groupTreeItem{band: gh, rowNo: 0, rowCount: 0}
	root := &groupTreeItem{band: nil}
	root.addItem(leaf)
	// NOTE: showGroupTree on root (band==nil, items has leaf) goes into the
	// non-leaf branch: showDataHeader(leaf.band), loop over items, showDataFooter.
	// leaf.band != nil so showDataHeader is called with gh.
	// Inside the items loop: leaf.band.SetIsFirstRow(0==0) is called.
	// showGroupTree(leaf) → leaf has no items → leaf node path → rowCount==0 → skip.
	// Must not panic.
	e.showGroupTree(root)
}
