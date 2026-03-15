package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func newProcessAtEngine(t *testing.T) *engine.ReportEngine {
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

func TestAddDeferredHandler_FiresOnMatchingState(t *testing.T) {
	e := newProcessAtEngine(t)
	fired := false
	e.AddDeferredHandler(engine.EngineStateReportFinished, func() { fired = true })
	e.OnStateChanged(nil, engine.EngineStateReportFinished)
	if !fired {
		t.Error("deferred handler should have fired on ReportFinished")
	}
}

func TestAddDeferredHandler_DoesNotFireOnOtherState(t *testing.T) {
	e := newProcessAtEngine(t)
	fired := false
	e.AddDeferredHandler(engine.EngineStateReportFinished, func() { fired = true })
	e.OnStateChanged(nil, engine.EngineStatePageFinished)
	if fired {
		t.Error("deferred handler should not fire on a different state")
	}
}

func TestAddDeferredHandler_RemovedAfterFiring(t *testing.T) {
	e := newProcessAtEngine(t)
	count := 0
	e.AddDeferredHandler(engine.EngineStatePageStarted, func() { count++ })
	e.OnStateChanged(nil, engine.EngineStatePageStarted)
	e.OnStateChanged(nil, engine.EngineStatePageStarted)
	if count != 1 {
		t.Errorf("deferred handler fired %d times, want 1", count)
	}
}

func TestAddStateHandler_PersistentCallback(t *testing.T) {
	e := newProcessAtEngine(t)
	count := 0
	e.AddStateHandler(func(_ any, _ engine.EngineState) { count++ })
	e.OnStateChanged(nil, engine.EngineStatePageStarted)
	e.OnStateChanged(nil, engine.EngineStatePageFinished)
	if count != 2 {
		t.Errorf("state handler called %d times, want 2", count)
	}
}

func TestAddRepeatingDeferredHandler_FiresOnEveryOccurrence(t *testing.T) {
	e := newProcessAtEngine(t)
	count := 0
	e.AddRepeatingDeferredHandler(engine.EngineStatePageFinished, func() { count++ })
	e.OnStateChanged(nil, engine.EngineStatePageFinished)
	e.OnStateChanged(nil, engine.EngineStatePageFinished)
	e.OnStateChanged(nil, engine.EngineStatePageFinished)
	if count != 3 {
		t.Errorf("repeating handler fired %d times, want 3", count)
	}
}

func TestAddRepeatingDeferredHandler_DoesNotFireOnOtherState(t *testing.T) {
	e := newProcessAtEngine(t)
	count := 0
	e.AddRepeatingDeferredHandler(engine.EngineStatePageFinished, func() { count++ })
	e.OnStateChanged(nil, engine.EngineStateReportFinished)
	if count != 0 {
		t.Errorf("repeating handler should not fire on other state, count = %d", count)
	}
}

func TestClearDeferredHandlers_RemovesAll(t *testing.T) {
	e := newProcessAtEngine(t)
	fired := false
	e.AddDeferredHandler(engine.EngineStateReportFinished, func() { fired = true })
	e.ClearDeferredHandlers()
	e.OnStateChanged(nil, engine.EngineStateReportFinished)
	if fired {
		t.Error("cleared handler should not fire")
	}
}
