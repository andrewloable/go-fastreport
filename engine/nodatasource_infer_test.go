package engine

// nodatasource_infer_test.go — internal tests (package engine) covering:
//   - inferFilterDataSource: with a matching [Alias.Column] filter expression
//   - inferFilterDataSource: with a filter expression that references an unknown alias
//   - inferFilterDataSource: with no bracket expression in the filter
//   - inferFilterDataSource: with an empty filter
//   - runDataBandNoDS: renders the band once when rowCount == 1 (virtual DS)
//   - runDataBandNoDS: renders the band even when no filter is present
//   - runDataBandNoDS: filter returns false, band is suppressed
//
// These are internal tests because both functions are unexported.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// newInferEngine builds a minimal engine with the report's dictionary
// pre-populated with a single data source named "Orders" / alias "Orders".
func newInferEngine(t *testing.T) (*ReportEngine, *data.BaseDataSource) {
	t.Helper()
	ds := data.NewBaseDataSource("Orders")
	ds.SetAlias("Orders")
	ds.AddColumn(data.Column{Name: "OrderID"})
	ds.AddRow(map[string]any{"OrderID": int64(10248)})
	ds.AddRow(map[string]any{"OrderID": int64(10249)})
	if err := ds.Init(); err != nil {
		t.Fatalf("ds.Init: %v", err)
	}

	r := reportpkg.NewReport()
	r.Dictionary().AddDataSource(ds)
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("engine.Run: %v", err)
	}
	return e, ds
}

// ── inferFilterDataSource ─────────────────────────────────────────────────────

// TestInferFilterDataSource_MatchingAlias exercises the happy path: the filter
// "[Orders.OrderID] > 5" contains the alias "Orders" which is registered in
// the dictionary, so inferFilterDataSource should return the data source.
func TestInferFilterDataSource_MatchingAlias(t *testing.T) {
	e, wantDS := newInferEngine(t)

	db := band.NewDataBand()
	db.SetName("InferMatch")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetFilter("[Orders.OrderID] > 5")

	got := e.inferFilterDataSource(db)
	if got == nil {
		t.Fatal("inferFilterDataSource: expected non-nil data source, got nil")
	}
	// The returned datasource should be the same object registered.
	if got != data.DataSource(wantDS) {
		t.Errorf("inferFilterDataSource: got %v, want Orders datasource", got)
	}
}

// TestInferFilterDataSource_UnknownAlias exercises the case where the bracket
// expression references an alias not present in the dictionary.
func TestInferFilterDataSource_UnknownAlias(t *testing.T) {
	e, _ := newInferEngine(t)

	db := band.NewDataBand()
	db.SetName("InferUnknown")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetFilter("[Customers.CustomerID] == 'ALFKI'")

	got := e.inferFilterDataSource(db)
	if got != nil {
		t.Errorf("inferFilterDataSource: expected nil for unknown alias, got %v", got)
	}
}

// TestInferFilterDataSource_NoBracket exercises the case where the filter
// expression has no bracketed name at all.
func TestInferFilterDataSource_NoBracket(t *testing.T) {
	e, _ := newInferEngine(t)

	db := band.NewDataBand()
	db.SetName("InferNoBracket")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetFilter("1 == 1") // no [brackets]

	got := e.inferFilterDataSource(db)
	if got != nil {
		t.Errorf("inferFilterDataSource: expected nil for filter with no bracket, got %v", got)
	}
}

// TestInferFilterDataSource_EmptyFilter exercises the early return when the
// filter string is empty.
func TestInferFilterDataSource_EmptyFilter(t *testing.T) {
	e, _ := newInferEngine(t)

	db := band.NewDataBand()
	db.SetName("InferEmpty")
	db.SetHeight(10)
	db.SetVisible(true)
	// No SetFilter call — filter is ""

	got := e.inferFilterDataSource(db)
	if got != nil {
		t.Errorf("inferFilterDataSource: expected nil for empty filter, got %v", got)
	}
}

// TestInferFilterDataSource_NilReport exercises the early return when the
// engine has no report set.
func TestInferFilterDataSource_NilReport(t *testing.T) {
	e := New(nil) // engine with nil report

	db := band.NewDataBand()
	db.SetName("InferNilReport")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetFilter("[Orders.OrderID] > 5")

	got := e.inferFilterDataSource(db)
	if got != nil {
		t.Errorf("inferFilterDataSource: expected nil when report is nil, got %v", got)
	}
}

// TestInferFilterDataSource_BareNameNoAlias exercises the path where the
// bracket expression has no dot (no alias component), so the whole name is
// used as the alias for the dictionary lookup.
func TestInferFilterDataSource_BareNameNoAlias(t *testing.T) {
	e, _ := newInferEngine(t)

	// "[Orders]" — name with no dot; alias = "Orders" → should match.
	db := band.NewDataBand()
	db.SetName("InferBareName")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetFilter("[Orders] > 0")

	got := e.inferFilterDataSource(db)
	// "Orders" alias exists so this should resolve.
	if got == nil {
		t.Fatal("inferFilterDataSource: expected non-nil for bare [Orders] expression, got nil")
	}
}

// ── runDataBandNoDS ───────────────────────────────────────────────────────────

// TestRunDataBandNoDS_RendersOnce exercises the primary code path of
// runDataBandNoDS: a DataBand with no data source is rendered exactly once
// (VirtualDataSource RowCount=1), advancing CurY by the band height.
func TestRunDataBandNoDS_RendersOnce(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	db := band.NewDataBand()
	db.SetName("NoDSBand")
	db.SetHeight(30)
	db.SetVisible(true)
	// No data source, no filter.

	beforeY := e.curY
	if err := e.runDataBandNoDS(db); err != nil {
		t.Fatalf("runDataBandNoDS: %v", err)
	}
	// Band should have been rendered once: curY advances by height.
	if e.curY != beforeY+30 {
		t.Errorf("runDataBandNoDS: curY = %v, want %v (one render)", e.curY, beforeY+30)
	}
	// RowNo should be set to 1 (first and last row of the virtual DS).
	if db.RowNo() != 1 {
		t.Errorf("runDataBandNoDS: RowNo = %d, want 1", db.RowNo())
	}
	if !db.IsFirstRow() {
		t.Error("runDataBandNoDS: IsFirstRow should be true")
	}
	if !db.IsLastRow() {
		t.Error("runDataBandNoDS: IsLastRow should be true")
	}
}

// TestRunDataBandNoDS_WithInferredDS exercises runDataBandNoDS when the filter
// expression references a known data source alias. The engine should call
// inferFilterDataSource, find the "Orders" DS, call First() on it to set the
// calc context, then render the band once.
func TestRunDataBandNoDS_WithInferredDS(t *testing.T) {
	e, _ := newInferEngine(t)

	db := band.NewDataBand()
	db.SetName("NoDSWithFilter")
	db.SetHeight(15)
	db.SetVisible(true)
	// Filter references the "Orders" alias — inferFilterDataSource should
	// find it and set it as the calc context.
	db.SetFilter("[Orders.OrderID] > 0")

	beforeY := e.curY
	if err := e.runDataBandNoDS(db); err != nil {
		t.Fatalf("runDataBandNoDS with inferred DS: %v", err)
	}
	// Filter evaluates to true (OrderID=10248 > 0), so the band renders once.
	if e.curY != beforeY+15 {
		t.Errorf("runDataBandNoDS with inferred DS: curY = %v, want %v", e.curY, beforeY+15)
	}
}

// TestRunDataBandNoDS_FilterFalse exercises the filter suppression path: when
// the filter expression evaluates to false the band must not render (curY
// must not advance).
func TestRunDataBandNoDS_FilterFalse(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	db := band.NewDataBand()
	db.SetName("NoDSFilterFalse")
	db.SetHeight(20)
	db.SetVisible(true)
	// Expression that always resolves to false via Calc (literal bool).
	// Use a parameter expression that the engine can evaluate.
	db.SetFilter("false")

	beforeY := e.curY
	if err := e.runDataBandNoDS(db); err != nil {
		t.Fatalf("runDataBandNoDS filter-false: %v", err)
	}
	// Filter returned false — band should be suppressed, curY unchanged.
	if e.curY != beforeY {
		t.Errorf("runDataBandNoDS filter-false: curY = %v, want %v (band suppressed)", e.curY, beforeY)
	}
}
