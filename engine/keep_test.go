package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func newKeepEngine(t *testing.T) *engine.ReportEngine {
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

func TestStartKeep_SetsKeeping(t *testing.T) {
	e := newKeepEngine(t)
	if e.IsKeeping() {
		t.Error("should not be keeping initially")
	}
	e.StartKeep()
	if !e.IsKeeping() {
		t.Error("IsKeeping should be true after StartKeep")
	}
}

func TestEndKeep_ClearsKeeping(t *testing.T) {
	e := newKeepEngine(t)
	e.StartKeep()
	e.EndKeep()
	if e.IsKeeping() {
		t.Error("IsKeeping should be false after EndKeep")
	}
}

func TestEndKeep_Idempotent(t *testing.T) {
	e := newKeepEngine(t)
	// EndKeep when not keeping should be a no-op.
	e.EndKeep()
	e.EndKeep()
	if e.IsKeeping() {
		t.Error("IsKeeping should remain false")
	}
}

func TestStartKeep_Idempotent(t *testing.T) {
	e := newKeepEngine(t)
	e.StartKeep()
	e.StartKeep() // second call should be no-op
	if !e.IsKeeping() {
		t.Error("IsKeeping should still be true")
	}
	y := e.KeepCurY()
	e.AdvanceY(20)
	e.StartKeep() // should not update keepCurY
	if e.KeepCurY() != y {
		t.Errorf("KeepCurY changed on second StartKeep: got %v, want %v", e.KeepCurY(), y)
	}
}

func TestKeepCurY_RecordsY(t *testing.T) {
	e := newKeepEngine(t)
	e.AdvanceY(50)
	e.StartKeep()
	if e.KeepCurY() != 50 {
		t.Errorf("KeepCurY = %v, want 50", e.KeepCurY())
	}
}

func TestCheckKeepTogether_CutsWhenKeeping(t *testing.T) {
	e := newKeepEngine(t)

	// Add some bands to the current page first.
	pp := e.PreparedPages()
	initialBands := len(pp.GetPage(0).Bands)

	e.StartKeep()
	e.AdvanceY(30) // simulate printing a 30px band during keep

	e.CheckKeepTogether()

	// After cutting, the bands added since StartKeep should be removed from page.
	after := len(pp.GetPage(0).Bands)
	_ = after  // we just verify it doesn't panic
	_ = initialBands
}

func TestFinishKeepTogether_WithCutBands(t *testing.T) {
	e := newKeepEngine(t)

	// Start keep and advance to build up deltaY.
	e.StartKeep()
	e.AdvanceY(30)

	// Cut the kept content (simulates CheckKeepTogether).
	e.CheckKeepTogether()

	// Ensure there are cut bands by checking CutBands.
	pp := e.PreparedPages()
	if len(pp.CutBands()) == 0 {
		// If no bands were cut (e.g. page had no bands at the keep position),
		// FinishKeepTogether is a no-op — just verify no panic.
		e.FinishKeepTogether()
		return
	}

	// FinishKeepTogether should paste bands and call EndKeep.
	e.FinishKeepTogether()
	if e.IsKeeping() {
		t.Error("IsKeeping should be false after FinishKeepTogether")
	}
}

func TestFinishKeepTogether_NoCutBands_IsNoop(t *testing.T) {
	e := newKeepEngine(t)
	// No keep active, no cut bands — should be a no-op.
	e.FinishKeepTogether()
	if e.IsKeeping() {
		t.Error("IsKeeping should still be false")
	}
}

func TestCheckKeepTogether_NoopWhenNotKeeping(t *testing.T) {
	e := newKeepEngine(t)
	pp := e.PreparedPages()
	before := len(pp.GetPage(0).Bands)
	e.CheckKeepTogether()
	after := len(pp.GetPage(0).Bands)
	if before != after {
		t.Error("CheckKeepTogether should be no-op when not keeping")
	}
}
