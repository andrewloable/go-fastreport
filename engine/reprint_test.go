package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func newReprintEngine(t *testing.T) *engine.ReportEngine {
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

func TestAddReprint_AddsToFooters(t *testing.T) {
	e := newReprintEngine(t)
	b := band.NewBandBase()
	b.SetName("Footer")
	b.SetVisible(true)
	b.SetHeight(10)
	e.AddReprint(b)
	if e.ReprintFooterCount() != 1 {
		t.Errorf("ReprintFooterCount = %d, want 1", e.ReprintFooterCount())
	}
}

func TestRemoveReprint_RemovesFromFooters(t *testing.T) {
	e := newReprintEngine(t)
	b := band.NewBandBase()
	b.SetName("Footer")
	b.SetHeight(10)
	e.AddReprint(b)
	e.RemoveReprint(b)
	if e.ReprintFooterCount() != 0 {
		t.Errorf("ReprintFooterCount after remove = %d, want 0", e.ReprintFooterCount())
	}
}

func TestRemoveReprint_NilSafe(t *testing.T) {
	e := newReprintEngine(t)
	// Should not panic.
	e.RemoveReprint(nil)
}

func TestShowReprintFooters_NoPanic(t *testing.T) {
	e := newReprintEngine(t)
	b := band.NewBandBase()
	b.SetName("Footer")
	b.SetVisible(true)
	b.SetHeight(10)
	e.AddReprint(b)
	// Just verify it doesn't panic; the band is rendered via ShowFullBand.
	e.ShowReprintFooters()
}

func TestShowReprintHeaders_NoPanic(t *testing.T) {
	e := newReprintEngine(t)
	// Empty list — should be a no-op.
	e.ShowReprintHeaders()
}

func TestAddReprintDataHeader_NilSafe(t *testing.T) {
	e := newReprintEngine(t)
	e.AddReprintDataHeader(nil) // should not panic
	if e.ReprintHeaderCount() != 0 {
		t.Errorf("nil AddReprintDataHeader should not add entry")
	}
}

func TestAddReprintDataHeader_Registers(t *testing.T) {
	e := newReprintEngine(t)
	dh := band.NewDataHeaderBand()
	dh.SetName("DH")
	dh.SetHeight(10)
	dh.SetVisible(true)
	e.AddReprintDataHeader(dh)
	if e.ReprintHeaderCount() != 1 {
		t.Errorf("ReprintHeaderCount = %d, want 1", e.ReprintHeaderCount())
	}
}

func TestAddReprintGroupHeader_NilSafe(t *testing.T) {
	e := newReprintEngine(t)
	e.AddReprintGroupHeader(nil) // should not panic
}

func TestAddReprintGroupHeader_Registers(t *testing.T) {
	e := newReprintEngine(t)
	gh := band.NewGroupHeaderBand()
	gh.SetName("GH")
	gh.SetHeight(10)
	e.AddReprintGroupHeader(gh)
	if e.ReprintHeaderCount() != 1 {
		t.Errorf("ReprintHeaderCount = %d, want 1 after AddReprintGroupHeader", e.ReprintHeaderCount())
	}
}

func TestAddReprintDataFooter_NilSafe(t *testing.T) {
	e := newReprintEngine(t)
	e.AddReprintDataFooter(nil) // should not panic
}

func TestAddReprintDataFooter_Registers(t *testing.T) {
	e := newReprintEngine(t)
	df := band.NewDataFooterBand()
	df.SetName("DF")
	df.SetHeight(10)
	e.AddReprintDataFooter(df)
	if e.ReprintFooterCount() != 1 {
		t.Errorf("ReprintFooterCount = %d, want 1 after AddReprintDataFooter", e.ReprintFooterCount())
	}
}

func TestAddReprintGroupFooter_NilSafe(t *testing.T) {
	e := newReprintEngine(t)
	e.AddReprintGroupFooter(nil) // should not panic
}

func TestAddReprintGroupFooter_Registers(t *testing.T) {
	e := newReprintEngine(t)
	gf := band.NewGroupFooterBand()
	gf.SetName("GF")
	gf.SetHeight(10)
	e.AddReprintGroupFooter(gf)
	if e.ReprintFooterCount() != 1 {
		t.Errorf("ReprintFooterCount = %d, want 1 after AddReprintGroupFooter", e.ReprintFooterCount())
	}
}

func TestReprintHeaderCount(t *testing.T) {
	e := newReprintEngine(t)
	if e.ReprintHeaderCount() != 0 {
		t.Errorf("initial ReprintHeaderCount = %d, want 0", e.ReprintHeaderCount())
	}
	dh := band.NewDataHeaderBand()
	dh.SetName("DH")
	dh.SetHeight(10)
	e.AddReprintDataHeader(dh)
	if e.ReprintHeaderCount() != 1 {
		t.Errorf("ReprintHeaderCount after add = %d, want 1", e.ReprintHeaderCount())
	}
}
