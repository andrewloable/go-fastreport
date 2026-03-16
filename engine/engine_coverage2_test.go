package engine_test

// engine_coverage2_test.go — additional coverage for subreports.go,
// reprint.go (addReprintBand keeping+footer), and relations.go branches.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── addReprintBand: keeping=true, isHeader=false → keepReprintFooters ─────────

func TestAddReprintBand_KeepingFooter(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// StartKeep sets keeping=true; AddReprintDataFooter calls addReprintBand
	// with isHeader=false, exercising the keeping=true + isHeader=false branch.
	e.StartKeep()
	df := band.NewDataFooterBand()
	df.SetName("KF")
	df.SetHeight(10)
	df.SetVisible(true)
	e.AddReprintDataFooter(df)
	e.EndKeep()
	// No panic — the keepReprintFooters branch was exercised.
}

func TestAddReprintBand_KeepingGroupFooter(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	e.StartKeep()
	gf := band.NewGroupFooterBand()
	gf.SetName("GF2")
	gf.SetHeight(10)
	e.AddReprintGroupFooter(gf)
	e.EndKeep()
}

// ── RenderInnerSubreports: non-SubreportObject child ─────────────────────────
//
// Exercises the `!ok` continue branch by adding a non-SubreportObject to the
// band's object collection before a real inner subreport.

func TestRenderInnerSubreports_NonSubreportObject_Skipped(t *testing.T) {
	r := reportpkg.NewReport()
	pg1 := reportpkg.NewReportPage()
	pg1.SetName("Main2")
	r.AddPage(pg1)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	b := band.NewBandBase()

	// A TextObject satisfies object.ReportObject but not *SubreportObject.
	txt := object.NewTextObject()
	txt.SetName("Txt1")
	b.Objects().Add(txt)

	// An outer subreport (PrintOnParent=false) — should be skipped by
	// RenderInnerSubreports.
	sr := object.NewSubreportObject()
	sr.SetReportPageName("NonExistent2")
	sr.SetPrintOnParent(false)
	b.Objects().Add(sr)

	beforeY := e.CurY()
	e.RenderInnerSubreports(b)
	if e.CurY() != beforeY {
		t.Error("RenderInnerSubreports changed CurY when no inner subreports present")
	}
}

// ── RenderOuterSubreports: inner-only (PrintOnParent=true) object skipped ─────

func TestRenderOuterSubreports_SkipsInnerSubreport(t *testing.T) {
	r := reportpkg.NewReport()
	pg1 := reportpkg.NewReportPage()
	pg1.SetName("Main3")
	r.AddPage(pg1)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	b := band.NewBandBase()

	// PrintOnParent=true → should be skipped by RenderOuterSubreports.
	inner := object.NewSubreportObject()
	inner.SetReportPageName("NonExistent3")
	inner.SetPrintOnParent(true)
	b.Objects().Add(inner)

	beforeY := e.CurY()
	e.RenderOuterSubreports(b)
	// hasSubreports remains false → CurY should not change.
	if e.CurY() != beforeY {
		t.Errorf("RenderOuterSubreports with inner-only subreport changed CurY: got %v, want %v",
			e.CurY(), beforeY)
	}
}

// ── applyRelationFilters: parentDS does not satisfy data.DataSource ───────────
//
// When parentBand.DataSourceRef() returns a band.DataSource that does NOT
// implement data.DataSource (e.g. sliceDS), applyRelationFilters returns early.

func TestApplyRelationFilters_ParentNotDataDS(t *testing.T) {
	masterDS := data.NewBaseDataSource("Customers")
	masterDS.SetAlias("Customers")
	masterDS.AddColumn(data.Column{Name: "ID"})
	masterDS.AddRow(map[string]any{"ID": "1"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	detailDS := data.NewBaseDataSource("Orders2")
	detailDS.SetAlias("Orders2")
	detailDS.AddColumn(data.Column{Name: "CustID"})
	detailDS.AddRow(map[string]any{"CustID": "1"})
	if err := detailDS.Init(); err != nil {
		t.Fatalf("detailDS.Init: %v", err)
	}

	rel := &data.Relation{
		Name:             "CustToOrders",
		ParentDataSource: masterDS,
		ChildDataSource:  detailDS,
		ParentColumns:    []string{"ID"},
		ChildColumns:     []string{"CustID"},
	}

	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddDataSource(detailDS)
	dict.AddRelation(rel)

	pg := reportpkg.NewReportPage()

	// Master band uses a sliceDS which does NOT implement data.DataSource
	// (missing Init, Close, Name, Alias, CurrentRowNo, RowCount methods).
	masterBand := band.NewDataBand()
	masterBand.SetName("MB2")
	masterBand.SetHeight(10)
	masterBand.SetVisible(true)
	masterBand.SetDataSource(newSliceDS("1", "2")) // sliceDS ≠ data.DataSource

	detailBand := band.NewDataBand()
	detailBand.SetName("DB2")
	detailBand.SetHeight(10)
	detailBand.SetVisible(true)
	detailBand.SetDataSource(detailDS)

	masterBand.Objects().Add(detailBand)

	pg.AddBand(masterBand)
	r.AddPage(pg)

	e := engine.New(r)
	// Should not panic — applyRelationFilters returns early because sliceDS
	// does not satisfy data.DataSource.
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

// ── applyRelationFilters: child datasource not found / alias branch ───────────

func TestApplyRelationFilters_ChildDSNil_AliasResolution(t *testing.T) {
	masterDS := data.NewBaseDataSource("Parents")
	masterDS.SetAlias("Parents")
	masterDS.AddColumn(data.Column{Name: "PID"})
	masterDS.AddRow(map[string]any{"PID": "10"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	childDS := data.NewBaseDataSource("Children")
	childDS.SetAlias("Children")
	childDS.AddColumn(data.Column{Name: "PID"})
	childDS.AddRow(map[string]any{"PID": "10"})
	if err := childDS.Init(); err != nil {
		t.Fatalf("childDS.Init: %v", err)
	}

	rel := &data.Relation{
		Name:             "ParentChild",
		ParentDataSource: masterDS,
		ChildDataSource:  childDS,
		ParentColumns:    []string{"PID"},
		ChildColumns:     []string{"PID"},
	}

	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddDataSource(childDS)
	dict.AddRelation(rel)

	pg := reportpkg.NewReportPage()

	masterBand := band.NewDataBand()
	masterBand.SetName("ParentBand")
	masterBand.SetHeight(10)
	masterBand.SetVisible(true)
	masterBand.SetDataSource(masterDS)

	// Child band has NO direct data source set; the engine will try to resolve
	// by DataSourceAlias. We leave the data source unset to exercise the
	// childDS==nil branch. The alias is also not set here so the resolution
	// path returns nil → continue.
	childBand := band.NewDataBand()
	childBand.SetName("ChildBand")
	childBand.SetHeight(10)
	childBand.SetVisible(true)
	// No SetDataSource → DataSourceRef() returns nil; alias also empty → continue.

	masterBand.Objects().Add(childBand)
	pg.AddBand(masterBand)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

// ── applyRelationFilters: no matching relation between parent and child ────────

func TestApplyRelationFilters_NoRelation(t *testing.T) {
	masterDS := data.NewBaseDataSource("MasterNoRel")
	masterDS.SetAlias("MasterNoRel")
	masterDS.AddColumn(data.Column{Name: "ID"})
	masterDS.AddRow(map[string]any{"ID": "1"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	childDS := data.NewBaseDataSource("ChildNoRel")
	childDS.SetAlias("ChildNoRel")
	childDS.AddColumn(data.Column{Name: "ParentID"})
	childDS.AddRow(map[string]any{"ParentID": "1"})
	if err := childDS.Init(); err != nil {
		t.Fatalf("childDS.Init: %v", err)
	}

	// No relation added to dictionary — FindRelation returns nil → continue.
	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddDataSource(childDS)
	// intentionally no AddRelation

	pg := reportpkg.NewReportPage()

	masterBand := band.NewDataBand()
	masterBand.SetName("MB_NoRel")
	masterBand.SetHeight(10)
	masterBand.SetVisible(true)
	masterBand.SetDataSource(masterDS)

	childBand := band.NewDataBand()
	childBand.SetName("CB_NoRel")
	childBand.SetHeight(10)
	childBand.SetVisible(true)
	childBand.SetDataSource(childDS)

	masterBand.Objects().Add(childBand)
	pg.AddBand(masterBand)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

// ── applyRelationFilters: relation with empty ParentColumns ───────────────────

func TestApplyRelationFilters_EmptyParentColumns(t *testing.T) {
	masterDS := data.NewBaseDataSource("MasterEmpty")
	masterDS.SetAlias("MasterEmpty")
	masterDS.AddColumn(data.Column{Name: "ID"})
	masterDS.AddRow(map[string]any{"ID": "1"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	childDS := data.NewBaseDataSource("ChildEmpty")
	childDS.SetAlias("ChildEmpty")
	childDS.AddColumn(data.Column{Name: "PID"})
	childDS.AddRow(map[string]any{"PID": "1"})
	if err := childDS.Init(); err != nil {
		t.Fatalf("childDS.Init: %v", err)
	}

	rel := &data.Relation{
		Name:             "EmptyColsRel",
		ParentDataSource: masterDS,
		ChildDataSource:  childDS,
		ParentColumns:    []string{}, // empty → the len==0 guard fires
		ChildColumns:     []string{},
	}

	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddDataSource(childDS)
	dict.AddRelation(rel)

	pg := reportpkg.NewReportPage()

	masterBand := band.NewDataBand()
	masterBand.SetName("MB_Empty")
	masterBand.SetHeight(10)
	masterBand.SetVisible(true)
	masterBand.SetDataSource(masterDS)

	childBand := band.NewDataBand()
	childBand.SetName("CB_Empty")
	childBand.SetHeight(10)
	childBand.SetVisible(true)
	childBand.SetDataSource(childDS)

	masterBand.Objects().Add(childBand)
	pg.AddBand(masterBand)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

// ── applyRelationFilters: sub-band is not a DataBand (non-ok continue) ────────

func TestApplyRelationFilters_SubBandNotDataBand(t *testing.T) {
	masterDS := data.NewBaseDataSource("MasterNonDB")
	masterDS.SetAlias("MasterNonDB")
	masterDS.AddColumn(data.Column{Name: "ID"})
	masterDS.AddRow(map[string]any{"ID": "1"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	rel := &data.Relation{
		Name:             "NonDBRel",
		ParentDataSource: masterDS,
		ChildDataSource:  masterDS, // same, doesn't matter
		ParentColumns:    []string{"ID"},
		ChildColumns:     []string{"ID"},
	}

	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddRelation(rel)

	pg := reportpkg.NewReportPage()

	masterBand := band.NewDataBand()
	masterBand.SetName("MB_NonDB")
	masterBand.SetHeight(10)
	masterBand.SetVisible(true)
	masterBand.SetDataSource(masterDS)

	// Add a TextObject (not a DataBand) as a sub-object — exercises the !ok
	// continue branch in the subBands range.
	txt := object.NewTextObject()
	txt.SetName("TxtInBand")
	masterBand.Objects().Add(txt)

	pg.AddBand(masterBand)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

// ── applyRelationFilters: child DS satisfies band.DataSource but not data.DataSource

func TestApplyRelationFilters_ChildNotDataDS(t *testing.T) {
	masterDS := data.NewBaseDataSource("MasterChildNotDS")
	masterDS.SetAlias("MasterChildNotDS")
	masterDS.AddColumn(data.Column{Name: "ID"})
	masterDS.AddRow(map[string]any{"ID": "1"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	rel := &data.Relation{
		Name:             "ChildNotDSRel",
		ParentDataSource: masterDS,
		ChildDataSource:  masterDS,
		ParentColumns:    []string{"ID"},
		ChildColumns:     []string{"ID"},
	}

	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddRelation(rel)

	pg := reportpkg.NewReportPage()

	masterBand := band.NewDataBand()
	masterBand.SetName("MB_ChildNotDS")
	masterBand.SetHeight(10)
	masterBand.SetVisible(true)
	masterBand.SetDataSource(masterDS)

	// Child band uses sliceDS which satisfies band.DataSource but NOT data.DataSource.
	childBand := band.NewDataBand()
	childBand.SetName("CB_ChildNotDS")
	childBand.SetHeight(10)
	childBand.SetVisible(true)
	childBand.SetDataSource(newSliceDS("1")) // not data.DataSource

	masterBand.Objects().Add(childBand)
	pg.AddBand(masterBand)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}
