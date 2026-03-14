package reportpkg_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	_ "github.com/andrewloable/go-fastreport/engine" // registers prepare func
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func TestReport_Prepare_BasicRun(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(20)
	db.SetVisible(true)
	pg.AddBand(db)
	r.AddPage(pg)

	if err := r.Prepare(); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	pp := r.PreparedPages()
	if pp == nil {
		t.Fatal("PreparedPages is nil after Prepare")
	}
	if pp.Count() == 0 {
		t.Error("expected at least 1 prepared page")
	}
}

func TestReport_Prepare_EmptyReport(t *testing.T) {
	r := reportpkg.NewReport()
	// Report with no pages should prepare without error.
	if err := r.Prepare(); err != nil {
		t.Fatalf("Prepare: %v", err)
	}
}
