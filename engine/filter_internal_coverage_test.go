package engine

// filter_internal_coverage_test.go — internal test (package engine) covering
// the ds == nil guard in evalBandFilter (filter.go:24-26).
//
// This branch is unreachable via the public API because RunDataBandFull checks
// ds == nil itself before ever calling evalBandFilter. Direct internal access
// is required to call evalBandFilter with a DataBand that has no DataSource.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// TestEvalBandFilter_NilDataSource exercises the ds == nil guard (filter.go:24-26).
// A filter expression is set but no data source is bound, so DataSourceRef()
// returns nil and evalBandFilter must return true (pass-through).
func TestEvalBandFilter_NilDataSource(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	db := band.NewDataBand()
	db.SetName("FilterNilDS")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetFilter("[val] > 0") // filter set, but no SetDataSource call → DataSourceRef() == nil

	result := e.evalBandFilter(db)
	if !result {
		t.Error("evalBandFilter with nil DataSource should return true, got false")
	}
}
