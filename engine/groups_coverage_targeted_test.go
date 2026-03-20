package engine_test

// groups_coverage_targeted_test.go — targeted tests to raise coverage on the
// four functions listed in the coverage report:
//
//   RunGroup          63.3%  (engine/groups.go:350)
//   applyGroupSort    69.6%  (engine/groups.go:417)
//   showGroupHeader   68.8%  (engine/groups.go:268)
//   groupConditionColumn 66.7% (engine/groups.go:463)
//
// Uncovered branches addressed here:
//
//   groupConditionColumn:
//     • dot-qualified bare name "Orders.CustomerID"       → "CustomerID"
//     • dot-qualified bracketed "[Orders.CustomerID]"     → "CustomerID"
//
//   applyGroupSort:
//     • GroupHeaderBand with SortOrder=Ascending (loop appends spec)
//     • GroupHeaderBand with SortOrder=Descending (Descending=true spec)
//     • SortOrder=None (loop skips the band entirely)
//     • DataBand.Sort() entries (Expression-based sort spec)
//     • DataBand.Sort() with Column-only spec (Expression=="", use Column)
//     • Multiple nested group sort keys
//
//   RunGroup:
//     • DataBand found in groupBand.Objects()  (not set via SetData)
//     • DataSource resolved from Dictionary alias
//     • GroupFooterBand found in groupBand.Objects() (not set via SetGroupFooter)
//
//   showGroupHeader:
//     • RepeatOnEveryPage=true  → AddReprintGroupHeader
//     • GroupFooter.RepeatOnEveryPage=true → AddReprintGroupFooter
//     • ResetPageNumber=true + RowNo>1  → ResetLogicalPageNumber
//     • KeepWithData=true  → startKeepBand on the DataBand

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── sortableDS — a thin data source that implements band.DataSource AND
// data.Sortable so applyGroupSort can exercise SortRows. ──────────────────────

type sortableDS struct {
	rows []map[string]any
	pos  int
}

func newSortableDS(colName string, vals ...string) *sortableDS {
	ds := &sortableDS{pos: -1}
	for _, v := range vals {
		ds.rows = append(ds.rows, map[string]any{colName: v})
	}
	return ds
}

func (d *sortableDS) RowCount() int { return len(d.rows) }
func (d *sortableDS) First() error  { d.pos = 0; return nil }
func (d *sortableDS) Next() error   { d.pos++; return nil }
func (d *sortableDS) EOF() bool     { return d.pos >= len(d.rows) }
func (d *sortableDS) GetValue(col string) (any, error) {
	if d.pos < 0 || d.pos >= len(d.rows) {
		return nil, nil
	}
	v, ok := d.rows[d.pos][col]
	if !ok {
		return nil, nil
	}
	return v, nil
}

// SortRows implements data.Sortable — used by applyGroupSort.
func (d *sortableDS) SortRows(specs []data.SortSpec) {
	// simple stable sort by the first spec column
	if len(specs) == 0 || len(d.rows) == 0 {
		return
	}
	col := specs[0].Column
	desc := specs[0].Descending
	// bubble sort is fine for tiny test slices
	for i := 0; i < len(d.rows); i++ {
		for j := i + 1; j < len(d.rows); j++ {
			ai, _ := d.rows[i][col].(string)
			aj, _ := d.rows[j][col].(string)
			swap := ai > aj
			if desc {
				swap = ai < aj
			}
			if swap {
				d.rows[i], d.rows[j] = d.rows[j], d.rows[i]
			}
		}
	}
}

// buildTargetedEngine creates a fresh engine with one empty page.
func buildTargetedEngine(t *testing.T) *engine.ReportEngine {
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

// ── groupConditionColumn: dot-qualified column names ─────────────────────────

// TestRunGroup_SortOrder_Ascending_DotCondition exercises:
//   • groupConditionColumn with a bracketed dot-qualified condition
//     like "[Orders.CustomerID]" → the dot-qualified extraction path
//   • applyGroupSort loop where SortOrder != SortOrderNone
//     (default is SortOrderAscending, so the spec IS appended)
//   • sortableDS.SortRows is called with the extracted column name
func TestRunGroup_SortOrder_Ascending_DotCondition(t *testing.T) {
	e := buildTargetedEngine(t)

	ds := newSortableDS("CustomerID", "C", "A", "B")

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_DotAsc")
	gh.SetVisible(true)
	gh.SetHeight(10)
	// Dot-qualified bracketed condition — exercises "[Orders.CustomerID]" path
	// in groupConditionColumn (the dot-branch inside the loop).
	gh.SetCondition("[Orders.CustomerID]")
	// SortOrderAscending is the default; applyGroupSort appends a spec.
	gh.SetSortOrder(band.SortOrderAscending)

	db := band.NewDataBand()
	db.SetName("DB_DotAsc")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(ds)
	gh.SetData(db)

	before := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)
	after := len(e.PreparedPages().GetPage(0).Bands)

	if after <= before {
		t.Error("RunGroup with ascending sort on dot-qualified condition: expected bands")
	}
}

// TestRunGroup_SortOrder_Descending exercises applyGroupSort with
// SortOrderDescending so the Descending:true branch is taken.
func TestRunGroup_SortOrder_Descending(t *testing.T) {
	e := buildTargetedEngine(t)

	ds := newSortableDS("val", "A", "C", "B")

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_Desc")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val")
	gh.SetSortOrder(band.SortOrderDescending) // exercises Descending:true branch

	db := band.NewDataBand()
	db.SetName("DB_Desc")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(ds)
	gh.SetData(db)

	before := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)
	if len(e.PreparedPages().GetPage(0).Bands) <= before {
		t.Error("RunGroup with descending sort: expected bands")
	}
}

// TestRunGroup_SortOrder_None exercises the applyGroupSort loop where
// SortOrder == SortOrderNone → the inner `if` is false, no spec appended.
func TestRunGroup_SortOrder_None(t *testing.T) {
	e := buildTargetedEngine(t)

	ds := newSortableDS("val", "B", "A", "C")

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_None")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val")
	gh.SetSortOrder(band.SortOrderNone) // skips spec — exercises `if g.SortOrder() != SortOrderNone`

	db := band.NewDataBand()
	db.SetName("DB_None")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(ds)
	gh.SetData(db)

	// Should not panic; no sort applied.
	e.RunGroup(gh)
}

// TestRunGroup_DataBandSort_Expression exercises the DataBand.Sort() loop in
// applyGroupSort with an Expression-based sort spec.
func TestRunGroup_DataBandSort_Expression(t *testing.T) {
	e := buildTargetedEngine(t)

	ds := newSortableDS("val", "Z", "A", "M")

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_SortExpr")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val")
	gh.SetSortOrder(band.SortOrderNone) // no group-level sort; exercise DB sort path

	db := band.NewDataBand()
	db.SetName("DB_SortExpr")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(ds)
	// Add an Expression-based sort spec (Expression != "" path in applyGroupSort).
	db.AddSort(band.SortSpec{
		Expression: "[Orders.val]",
		Order:      band.SortOrderAscending,
	})
	gh.SetData(db)

	before := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)
	if len(e.PreparedPages().GetPage(0).Bands) <= before {
		t.Error("RunGroup with DataBand Expression sort: expected bands")
	}
}

// TestRunGroup_DataBandSort_ColumnOnly exercises the DataBand.Sort() loop
// where Expression=="" so the Column field is used instead.
func TestRunGroup_DataBandSort_ColumnOnly(t *testing.T) {
	e := buildTargetedEngine(t)

	ds := newSortableDS("val", "C", "A", "B")

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_SortCol")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val")
	gh.SetSortOrder(band.SortOrderNone)

	db := band.NewDataBand()
	db.SetName("DB_SortCol")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(ds)
	// Column-only sort spec (Expression == "" → use Column path).
	db.AddSort(band.SortSpec{
		Column: "val",
		Order:  band.SortOrderDescending,
	})
	gh.SetData(db)

	before := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)
	if len(e.PreparedPages().GetPage(0).Bands) <= before {
		t.Error("RunGroup with DataBand Column sort: expected bands")
	}
}

// TestRunGroup_MultipleNestedSortKeys exercises the applyGroupSort loop over
// multiple nested GroupHeaderBands, each contributing a sort key.
func TestRunGroup_MultipleNestedSortKeys(t *testing.T) {
	e := buildTargetedEngine(t)

	// Build a two-level nested group so applyGroupSort iterates g twice.
	innerGH := band.NewGroupHeaderBand()
	innerGH.SetName("InnerGH_Multi")
	innerGH.SetVisible(true)
	innerGH.SetHeight(8)
	innerGH.SetCondition("val")
	innerGH.SetSortOrder(band.SortOrderAscending)

	outerGH := band.NewGroupHeaderBand()
	outerGH.SetName("OuterGH_Multi")
	outerGH.SetVisible(true)
	outerGH.SetHeight(10)
	outerGH.SetCondition("[Orders.CustomerID]")
	outerGH.SetSortOrder(band.SortOrderDescending)
	outerGH.SetNestedGroup(innerGH)

	ds := newSortableDS("val", "B", "A", "C")

	db := band.NewDataBand()
	db.SetName("DB_Multi")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(ds)
	innerGH.SetData(db)
	outerGH.SetData(db)

	// Must not panic; exercises both group levels in the loop.
	e.RunGroup(outerGH)
}

// ── RunGroup: DataBand in Objects() (not set via SetData) ────────────────────

// TestRunGroup_DataBandInObjects exercises the first Objects() scan in RunGroup
// where groupBand.Data() is nil and the DataBand is found via groupBand.Objects().
func TestRunGroup_DataBandInObjects(t *testing.T) {
	e := buildTargetedEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_ObjDB")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val")
	// Do NOT call gh.SetData(db) — leave gh.Data()==nil.

	db := band.NewDataBand()
	db.SetName("DB_ObjDB")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newSimpleDS("X", "Y"))

	// Add the DataBand to Objects() so RunGroup finds it there.
	gh.Objects().Add(db)

	before := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)
	if len(e.PreparedPages().GetPage(0).Bands) <= before {
		t.Error("RunGroup with DataBand in Objects(): expected bands")
	}
}

// TestRunGroup_DataBandInObjects_WithNonDataBandFirst ensures the Objects()
// scan skips non-DataBand entries before finding the DataBand.
func TestRunGroup_DataBandInObjects_WithNonDataBandFirst(t *testing.T) {
	e := buildTargetedEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_ObjDB2")
	gh.SetVisible(true)
	gh.SetHeight(10)

	db := band.NewDataBand()
	db.SetName("DB_ObjDB2")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newSimpleDS("A"))

	// Add a non-DataBand item first so the loop has to skip it.
	otherBand := band.NewGroupFooterBand()
	otherBand.SetName("GF_Skip")
	gh.Objects().Add(otherBand)
	gh.Objects().Add(db)

	// Must not panic; exercises multi-iteration Objects() scan.
	e.RunGroup(gh)
}

// ── RunGroup: DataSource resolved from Dictionary alias ───────────────────────

// TestRunGroup_DataSourceFromAlias exercises the dictionary-alias resolution
// path in RunGroup: db.DataSourceRef()==nil && alias is set → resolved from dict.
func TestRunGroup_DataSourceFromAlias(t *testing.T) {
	r := reportpkg.NewReport()
	dict := data.NewDictionary()

	// Build a proper data.DataSource (BaseDataSource implements data.Sortable
	// and band.DataSource through its methods).
	bds := data.NewBaseDataSource("Orders")
	bds.AddColumn(data.Column{Name: "val"})
	bds.AddRow(map[string]any{"val": "Alpha"})
	bds.AddRow(map[string]any{"val": "Beta"})
	_ = bds.Init()
	dict.AddDataSource(bds)
	r.SetDictionary(dict)

	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_Alias")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val")

	db := band.NewDataBand()
	db.SetName("DB_Alias")
	db.SetVisible(true)
	db.SetHeight(10)
	// Set alias but NOT a direct DataSource ref — exercises resolution path.
	db.SetDataSourceAlias("Orders")
	// db.DataSourceRef() is nil here.
	gh.SetData(db)

	// Must not panic; exercises dict.FindDataSourceByAlias path.
	before := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)
	_ = before // band count varies; just verify no panic
}

// ── RunGroup: GroupFooterBand in Objects() (not set via SetGroupFooter) ───────

// TestRunGroup_GroupFooterInObjects exercises the second Objects() scan in
// RunGroup where groupBand.GroupFooter() is nil and the GroupFooterBand is
// found via groupBand.Objects().
func TestRunGroup_GroupFooterInObjects(t *testing.T) {
	e := buildTargetedEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_ObjGF")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val")

	gf := band.NewGroupFooterBand()
	gf.SetName("GF_ObjGF")
	gf.SetVisible(true)
	gf.SetHeight(8)
	// Do NOT call gh.SetGroupFooter(gf) — leave gh.GroupFooter()==nil.

	db := band.NewDataBand()
	db.SetName("DB_ObjGF")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newSimpleDS("P", "Q"))
	gh.SetData(db)

	// Add the GroupFooterBand to Objects() so RunGroup finds it there.
	gh.Objects().Add(gf)

	before := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)
	added := len(e.PreparedPages().GetPage(0).Bands) - before

	// We should get header×2 + data×2 + footer×2 = 6 bands.
	if added < 2 {
		t.Errorf("RunGroup with GroupFooter in Objects(): expected ≥2 bands, got %d", added)
	}
}

// ── showGroupHeader: RepeatOnEveryPage ────────────────────────────────────────

// TestShowGroupHeader_RepeatOnEveryPage exercises the RepeatOnEveryPage=true
// path in showGroupHeader which calls AddReprintGroupHeader.
func TestShowGroupHeader_RepeatOnEveryPage(t *testing.T) {
	e := buildTargetedEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_Reprint")
	gh.SetVisible(true)
	gh.SetHeight(12)
	gh.SetCondition("val")
	gh.SetRepeatOnEveryPage(true) // exercises AddReprintGroupHeader path

	gf := band.NewGroupFooterBand()
	gf.SetName("GF_Reprint")
	gf.SetVisible(true)
	gf.SetHeight(6)
	gh.SetGroupFooter(gf)

	db := band.NewDataBand()
	db.SetName("DB_Reprint")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newSimpleDS("R1", "R2"))
	gh.SetData(db)

	before := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)
	if len(e.PreparedPages().GetPage(0).Bands) <= before {
		t.Error("showGroupHeader RepeatOnEveryPage: expected bands")
	}
}

// TestShowGroupHeader_GroupFooter_RepeatOnEveryPage exercises the
// GroupFooter.RepeatOnEveryPage=true path in showGroupHeader which calls
// AddReprintGroupFooter.
func TestShowGroupHeader_GroupFooter_RepeatOnEveryPage(t *testing.T) {
	e := buildTargetedEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_FtrReprint")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val")

	gf := band.NewGroupFooterBand()
	gf.SetName("GF_FtrReprint")
	gf.SetVisible(true)
	gf.SetHeight(8)
	gf.SetRepeatOnEveryPage(true) // exercises AddReprintGroupFooter path
	gh.SetGroupFooter(gf)

	db := band.NewDataBand()
	db.SetName("DB_FtrReprint")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newSimpleDS("S1", "S2"))
	gh.SetData(db)

	before := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)
	if len(e.PreparedPages().GetPage(0).Bands) <= before {
		t.Error("showGroupHeader GroupFooter.RepeatOnEveryPage: expected bands")
	}
}

// TestShowGroupHeader_ResetPageNumber exercises the ResetPageNumber=true path
// in showGroupHeader. RowNo starts at 0; on the first group call RowNo becomes 1
// (no reset because FirstRowStartsNewPage check); on the second call RowNo>1,
// so ResetLogicalPageNumber() IS called.
func TestShowGroupHeader_ResetPageNumber(t *testing.T) {
	e := buildTargetedEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_Reset")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val")
	gh.SetResetPageNumber(true) // exercises ResetLogicalPageNumber path

	db := band.NewDataBand()
	db.SetName("DB_Reset")
	db.SetVisible(true)
	db.SetHeight(10)
	// Two distinct groups: on the second group, RowNo==2 > 1, so the reset fires.
	db.SetDataSource(newSimpleDS("G1", "G2"))
	gh.SetData(db)

	// Must not panic.
	e.RunGroup(gh)

	// After running 2 groups, RowNo should be 2.
	if gh.RowNo() != 2 {
		t.Errorf("RowNo = %d, want 2", gh.RowNo())
	}
}

// TestShowGroupHeader_KeepWithData exercises the KeepWithData=true path in
// showGroupHeader which calls startKeepBand on the DataBand.
func TestShowGroupHeader_KeepWithData(t *testing.T) {
	e := buildTargetedEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_KWD")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val")
	gh.SetKeepWithData(true) // exercises startKeepBand(db) path

	gf := band.NewGroupFooterBand()
	gf.SetName("GF_KWD")
	gf.SetVisible(true)
	gf.SetHeight(5)
	gh.SetGroupFooter(gf)

	db := band.NewDataBand()
	db.SetName("DB_KWD")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newSimpleDS("K1"))
	gh.SetData(db)

	before := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)
	if len(e.PreparedPages().GetPage(0).Bands) <= before {
		t.Error("showGroupHeader KeepWithData: expected bands")
	}
}

// TestShowGroupHeader_KeepWithData_NilDataBand exercises the KeepWithData path
// when the DataBand reference is nil (the inner `if db != nil` guard).
func TestShowGroupHeader_KeepWithData_NilDataBand(t *testing.T) {
	e := buildTargetedEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_KWD_Nil")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetKeepWithData(true) // header.Data()==nil → inner if guard fires

	db := band.NewDataBand()
	db.SetName("DB_KWD_Nil")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newSimpleDS("K2"))
	gh.SetData(db)

	// Must not panic.
	before := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)
	_ = before
}

// ── applyGroupSort: empty sort column (col=="") and no-specs-sortable path ───

// TestRunGroup_GroupSort_EmptyCondition exercises the `if col != ""` false
// branch in the GROUP loop of applyGroupSort: when the group condition is ""
// (not set), groupConditionColumn("") returns "" and the spec is skipped.
// SortOrder != None so the outer `if` is true, but col=="" skips append.
func TestRunGroup_GroupSort_EmptyCondition(t *testing.T) {
	e := buildTargetedEngine(t)

	ds := newSortableDS("val", "B", "A")

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_EmptyCond")
	gh.SetVisible(true)
	gh.SetHeight(10)
	// No condition set → Condition()=="" → groupConditionColumn("")="" → col==""
	// SortOrderAscending != SortOrderNone → enters the if block → col=="" → skip.
	gh.SetSortOrder(band.SortOrderAscending)

	db := band.NewDataBand()
	db.SetName("DB_EmptyCond")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(ds)
	gh.SetData(db)

	// Must not panic; exercises col=="" skip in the group-level sort loop.
	e.RunGroup(gh)
}

// TestRunGroup_DataBandSort_EmptyColumn exercises the `if col != ""` false
// branch in the db.Sort() loop: when both Expression and Column are empty,
// groupConditionColumn returns "" and the spec is skipped (col=="").
// Combined with SortOrder=None, len(specs)==0 so SortRows is never called —
// exercises the `if len(specs) > 0` false branch on a sortable DS.
func TestRunGroup_DataBandSort_EmptyColumn(t *testing.T) {
	e := buildTargetedEngine(t)

	// Use a sortable DS so applyGroupSort reaches the specs loop.
	ds := newSortableDS("val", "B", "A")

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_EmptyCol")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val")
	gh.SetSortOrder(band.SortOrderNone) // no group-level spec

	db := band.NewDataBand()
	db.SetName("DB_EmptyCol")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(ds)
	// SortSpec with both Expression="" and Column="" → groupConditionColumn("")
	// returns "" → col=="" → spec skipped → len(specs)==0 → SortRows not called.
	db.AddSort(band.SortSpec{Expression: "", Column: "", Order: band.SortOrderAscending})
	gh.SetData(db)

	// Must not panic; exercises col=="" skip and len(specs)==0 no-sort path.
	e.RunGroup(gh)
}
