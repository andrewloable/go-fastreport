package engine_test

// engine_reprint_keep_rel_test.go — coverage tests for reprint.go, keepwithdata.go,
// and relations.go uncovered branches.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── AddReprint: keeping=true path (goes to keepReprintFooters) ────────────────

func TestAddReprint_WhileKeeping(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// StartKeep sets e.keeping=true so AddReprint goes to keepReprintFooters.
	// This exercises the keeping=true branch in AddReprint.
	e.StartKeep()
	b := band.NewBandBase()
	b.SetName("B1")
	b.SetHeight(10)
	b.SetVisible(true)
	e.AddReprint(b)
	// Note: EndKeep does NOT merge keepReprintFooters — it just clears keeping.
	// The keep-scoped lists are internal and cleared by endKeepReprint, which is
	// called internally. So ReprintFooterCount stays 0.
	e.EndKeep()
	// Just verify no panic; keeping branch was exercised.
}

// ── addReprintBand: keeping=true header path ──────────────────────────────────

func TestAddReprintBand_KeepingHeader(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// exercises keeping=true path in addReprintBand with isHeader=true.
	e.StartKeep()
	dh := band.NewDataHeaderBand()
	dh.SetName("DH")
	dh.SetHeight(10)
	dh.SetVisible(true)
	e.AddReprintDataHeader(dh)
	e.EndKeep()
	// No panic expected; keeping branch for header exercised.
}

// ── ShowReprintHeaders: with non-empty list ───────────────────────────────────

func TestShowReprintHeaders_WithEntries(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Register a data header for reprinting.
	dh := band.NewDataHeaderBand()
	dh.SetName("ReprintHdr")
	dh.SetHeight(15)
	dh.SetVisible(true)
	e.AddReprintDataHeader(dh)
	if e.ReprintHeaderCount() != 1 {
		t.Fatalf("ReprintHeaderCount = %d, want 1", e.ReprintHeaderCount())
	}

	// ShowReprintHeaders iterates the list — should not panic.
	e.ShowReprintHeaders()
}

// ── removeReprintEntry: removing non-existent entry (no-op) ──────────────────

func TestRemoveReprint_NonExistentEntry(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Add one band, then try to remove a different band.
	b1 := band.NewBandBase()
	b1.SetName("B1")
	b1.SetHeight(10)
	e.AddReprint(b1)

	b2 := band.NewBandBase()
	b2.SetName("B2")
	b2.SetHeight(10)
	// b2 was never added — RemoveReprint should be a no-op.
	e.RemoveReprint(b2)

	if e.ReprintFooterCount() != 1 {
		t.Errorf("ReprintFooterCount = %d, want 1 (non-existent removal should not remove b1)", e.ReprintFooterCount())
	}
}

// ── applyRelationFilters: full master-detail setup ────────────────────────────

func TestApplyRelationFilters_MasterDetail(t *testing.T) {
	// Build a master (Orders) and detail (Items) data source.
	masterDS := data.NewBaseDataSource("Orders")
	masterDS.SetAlias("Orders")
	masterDS.AddColumn(data.Column{Name: "OrderID"})
	masterDS.AddColumn(data.Column{Name: "Customer"})
	masterDS.AddRow(map[string]any{"OrderID": "1", "Customer": "Alice"})
	masterDS.AddRow(map[string]any{"OrderID": "2", "Customer": "Bob"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	detailDS := data.NewBaseDataSource("Items")
	detailDS.SetAlias("Items")
	detailDS.AddColumn(data.Column{Name: "OrderID"})
	detailDS.AddColumn(data.Column{Name: "Item"})
	detailDS.AddRow(map[string]any{"OrderID": "1", "Item": "Widget"})
	detailDS.AddRow(map[string]any{"OrderID": "1", "Item": "Gadget"})
	detailDS.AddRow(map[string]any{"OrderID": "2", "Item": "Gizmo"})
	if err := detailDS.Init(); err != nil {
		t.Fatalf("detailDS.Init: %v", err)
	}

	// Create a relation linking Orders.OrderID → Items.OrderID.
	rel := &data.Relation{
		Name:             "OrdersToItems",
		ParentDataSource: masterDS,
		ChildDataSource:  detailDS,
		ParentColumns:    []string{"OrderID"},
		ChildColumns:     []string{"OrderID"},
	}

	// Build report.
	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddDataSource(detailDS)
	dict.AddRelation(rel)

	pg := reportpkg.NewReportPage()

	// Master DataBand.
	masterBand := band.NewDataBand()
	masterBand.SetName("MasterBand")
	masterBand.SetHeight(15)
	masterBand.SetVisible(true)
	masterBand.SetDataSource(masterDS)

	// Detail DataBand nested inside master (as sub-band via Objects).
	detailBand := band.NewDataBand()
	detailBand.SetName("DetailBand")
	detailBand.SetHeight(10)
	detailBand.SetVisible(true)
	detailBand.SetDataSource(detailDS)

	// Add a text object to detail band.
	detailTxt := object.NewTextObject()
	detailTxt.SetName("ItemText")
	detailTxt.SetLeft(0)
	detailTxt.SetTop(0)
	detailTxt.SetWidth(100)
	detailTxt.SetHeight(10)
	detailTxt.SetVisible(true)
	detailTxt.SetText("[Item]")
	detailBand.Objects().Add(detailTxt)

	// Add detail as sub-band of master.
	masterBand.Objects().Add(detailBand)

	pg.AddBand(masterBand)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with master-detail relation: %v", err)
	}
	if e.PreparedPages().Count() == 0 {
		t.Error("expected at least 1 prepared page")
	}
}

// ── checkKeepFooterWithData: DataFooterBand with KeepWithData=true ───────────

func TestCheckKeepFooterWithData_ViaEngine(t *testing.T) {
	// Create a data source with enough rows to fill the page.
	ds := data.NewBaseDataSource("DS")
	ds.SetAlias("DS")
	ds.AddColumn(data.Column{Name: "Val"})
	for i := 0; i < 30; i++ {
		ds.AddRow(map[string]any{"Val": i})
	}
	if err := ds.Init(); err != nil {
		t.Fatalf("ds.Init: %v", err)
	}

	r := reportpkg.NewReport()
	r.Dictionary().AddDataSource(ds)

	pg := reportpkg.NewReportPage()
	pg.PaperHeight = 100 // Very short page to force overflow.
	pg.PaperWidth = 210
	pg.TopMargin = 5
	pg.BottomMargin = 5

	// DataBand.
	db := band.NewDataBand()
	db.SetName("DataBand1")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(ds)

	// DataFooterBand with KeepWithData=true.
	ftr := band.NewDataFooterBand()
	ftr.SetName("Footer1")
	ftr.SetHeight(20) // Large footer to force overflow check.
	ftr.SetVisible(true)
	ftr.SetKeepWithData(true)
	db.SetFooter(ftr)

	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	// Just ensure it runs without panic.
	_ = e.Run(engine.DefaultRunOptions())
}

// ── checkKeepHeaderWithData: DataHeaderBand with KeepWithData=true ───────────

func TestCheckKeepHeaderWithData_ViaEngine(t *testing.T) {
	ds := data.NewBaseDataSource("DS2")
	ds.SetAlias("DS2")
	ds.AddColumn(data.Column{Name: "Val"})
	for i := 0; i < 20; i++ {
		ds.AddRow(map[string]any{"Val": i})
	}
	if err := ds.Init(); err != nil {
		t.Fatalf("ds.Init: %v", err)
	}

	r := reportpkg.NewReport()
	r.Dictionary().AddDataSource(ds)

	pg := reportpkg.NewReportPage()
	pg.PaperHeight = 80 // Short page to stress KeepWithData header.
	pg.PaperWidth = 210
	pg.TopMargin = 5
	pg.BottomMargin = 5

	// DataBand.
	db := band.NewDataBand()
	db.SetName("DataBand2")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(ds)

	// DataHeaderBand with KeepWithData=true.
	hdr := band.NewDataHeaderBand()
	hdr.SetName("Header1")
	hdr.SetHeight(30) // Large header to trigger relocation.
	hdr.SetVisible(true)
	hdr.SetKeepWithData(true)
	db.SetHeader(hdr)

	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())
}
