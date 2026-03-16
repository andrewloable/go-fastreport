package engine_test

// relation_e2e_test.go — end-to-end tests for master-detail relation filtering.
// Verifies that when a detail DataBand's data source has a relation from a master,
// the engine creates a FilteredDataSource for the detail source with the current
// master row's join-key values as filter criteria.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// TestRelation_MasterDetail_BasicFiltering verifies that detail rows are
// filtered to match the current master row in a simple 2-level master-detail
// report.
//
// Setup:
//   - Master: Customers with CustomerID = C1, C2, C3
//   - Detail: Orders with (CustomerID, OrderID):
//     C1->O1, C1->O2, C3->O3  (C2 has no orders)
//   - Relation: Customers.CustomerID -> Orders.CustomerID
//
// Expected rendered bands:
//   - MasterBand x 3 (one per customer)
//   - DetailBand x 3 (O1+O2 for C1, nothing for C2, O3 for C3)
func TestRelation_MasterDetail_BasicFiltering(t *testing.T) {
	// Master: Customers (3 rows)
	masterDS := data.NewBaseDataSource("Customers")
	masterDS.SetAlias("Customers")
	masterDS.AddColumn(data.Column{Name: "CustomerID"})
	masterDS.AddRow(map[string]any{"CustomerID": "C1"})
	masterDS.AddRow(map[string]any{"CustomerID": "C2"})
	masterDS.AddRow(map[string]any{"CustomerID": "C3"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	// Detail: Orders (only C1 and C3 have orders)
	detailDS := data.NewBaseDataSource("Orders")
	detailDS.SetAlias("Orders")
	detailDS.AddColumn(data.Column{Name: "CustomerID"})
	detailDS.AddColumn(data.Column{Name: "OrderID"})
	detailDS.AddRow(map[string]any{"CustomerID": "C1", "OrderID": "O1"})
	detailDS.AddRow(map[string]any{"CustomerID": "C1", "OrderID": "O2"})
	detailDS.AddRow(map[string]any{"CustomerID": "C3", "OrderID": "O3"})
	if err := detailDS.Init(); err != nil {
		t.Fatalf("detailDS.Init: %v", err)
	}

	// Relation: Customers.CustomerID -> Orders.CustomerID
	rel := &data.Relation{
		Name:             "CustomersOrders",
		ParentDataSource: masterDS,
		ChildDataSource:  detailDS,
		ParentColumns:    []string{"CustomerID"},
		ChildColumns:     []string{"CustomerID"},
	}

	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddDataSource(detailDS)
	dict.AddRelation(rel)

	pg := reportpkg.NewReportPage()

	masterBand := band.NewDataBand()
	masterBand.SetName("MasterBand")
	masterBand.SetHeight(15)
	masterBand.SetVisible(true)
	masterBand.SetDataSource(masterDS)

	detailBand := band.NewDataBand()
	detailBand.SetName("DetailBand")
	detailBand.SetHeight(10)
	detailBand.SetVisible(true)
	detailBand.SetDataSource(detailDS)

	// Nest detail band inside master band's objects so dataBandSubBands picks it up.
	masterBand.Objects().Add(detailBand)
	pg.AddBand(masterBand)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	bandCounts := map[string]int{}
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p != nil {
			for _, b := range p.Bands {
				bandCounts[b.Name]++
			}
		}
	}

	// 3 master rows rendered
	if bandCounts["MasterBand"] != 3 {
		t.Errorf("MasterBand count = %d, want 3", bandCounts["MasterBand"])
	}
	// 3 detail rows rendered (2 for C1 + 0 for C2 + 1 for C3)
	if bandCounts["DetailBand"] != 3 {
		t.Errorf("DetailBand count = %d, want 3 (2 for C1 + 0 for C2 + 1 for C3)", bandCounts["DetailBand"])
	}
}

// TestRelation_MasterDetail_AllMastersHaveDetails verifies filtering when every
// master row has detail rows.
func TestRelation_MasterDetail_AllMastersHaveDetails(t *testing.T) {
	masterDS := data.NewBaseDataSource("Customers")
	masterDS.SetAlias("Customers")
	masterDS.AddColumn(data.Column{Name: "CID"})
	masterDS.AddRow(map[string]any{"CID": "A"})
	masterDS.AddRow(map[string]any{"CID": "B"})
	if err := masterDS.Init(); err != nil {
		t.Fatalf("masterDS.Init: %v", err)
	}

	detailDS := data.NewBaseDataSource("Orders")
	detailDS.SetAlias("Orders")
	detailDS.AddColumn(data.Column{Name: "CID"})
	detailDS.AddColumn(data.Column{Name: "OID"})
	detailDS.AddRow(map[string]any{"CID": "A", "OID": "A1"})
	detailDS.AddRow(map[string]any{"CID": "B", "OID": "B1"})
	detailDS.AddRow(map[string]any{"CID": "B", "OID": "B2"})
	if err := detailDS.Init(); err != nil {
		t.Fatalf("detailDS.Init: %v", err)
	}

	rel := &data.Relation{
		Name:             "Rel",
		ParentDataSource: masterDS,
		ChildDataSource:  detailDS,
		ParentColumns:    []string{"CID"},
		ChildColumns:     []string{"CID"},
	}

	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(masterDS)
	dict.AddDataSource(detailDS)
	dict.AddRelation(rel)

	pg := reportpkg.NewReportPage()

	masterBand := band.NewDataBand()
	masterBand.SetName("Master")
	masterBand.SetHeight(10)
	masterBand.SetVisible(true)
	masterBand.SetDataSource(masterDS)

	detailBand := band.NewDataBand()
	detailBand.SetName("Detail")
	detailBand.SetHeight(8)
	detailBand.SetVisible(true)
	detailBand.SetDataSource(detailDS)

	masterBand.Objects().Add(detailBand)
	pg.AddBand(masterBand)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	bandCounts := map[string]int{}
	for i := 0; i < pp.Count(); i++ {
		p := pp.GetPage(i)
		if p != nil {
			for _, b := range p.Bands {
				bandCounts[b.Name]++
			}
		}
	}

	// 2 master rows
	if bandCounts["Master"] != 2 {
		t.Errorf("Master count = %d, want 2", bandCounts["Master"])
	}
	// 3 detail rows (1 for A + 2 for B)
	if bandCounts["Detail"] != 3 {
		t.Errorf("Detail count = %d, want 3 (1 for A + 2 for B)", bandCounts["Detail"])
	}
}

// TestRelation_MasterDetail_EmptyBaseDataSource verifies that a DataBand with
// an empty BaseDataSource (which returns ErrEOF from First()) is handled
// gracefully rather than propagating the error.
func TestRelation_MasterDetail_EmptyBaseDataSource(t *testing.T) {
	ds := data.NewBaseDataSource("Empty")
	ds.SetAlias("Empty")
	ds.AddColumn(data.Column{Name: "ID"})
	// No rows added
	if err := ds.Init(); err != nil {
		t.Fatalf("ds.Init: %v", err)
	}

	r := reportpkg.NewReport()
	dict := r.Dictionary()
	dict.AddDataSource(ds)

	pg := reportpkg.NewReportPage()

	db := band.NewDataBand()
	db.SetName("EmptyBand")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(ds)

	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	// Should not error — empty data source is valid.
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with empty BaseDataSource: %v", err)
	}

	pp := e.PreparedPages()
	if pp.Count() == 0 {
		t.Fatal("expected at least 1 prepared page")
	}
}
