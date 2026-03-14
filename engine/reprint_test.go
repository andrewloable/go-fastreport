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
