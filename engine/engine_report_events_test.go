package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// TestOnStartReport_CalledOnRun verifies that OnStartReport fires exactly once
// when engine.Run() executes.
func TestOnStartReport_CalledOnRun(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())

	called := 0
	r.OnStartReport = func() { called++ }

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	if called != 1 {
		t.Errorf("OnStartReport called %d times, want 1", called)
	}
}

// TestOnFinishReport_CalledOnRun verifies that OnFinishReport fires exactly once
// when engine.Run() executes.
func TestOnFinishReport_CalledOnRun(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())

	called := 0
	r.OnFinishReport = func() { called++ }

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	if called != 1 {
		t.Errorf("OnFinishReport called %d times, want 1", called)
	}
}

// TestOnStartReport_NilCallbackDoesNotPanic verifies that a nil OnStartReport
// does not panic.
func TestOnStartReport_NilCallbackDoesNotPanic(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	// OnStartReport intentionally left nil

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}
}

// TestOnFinishReport_NilCallbackDoesNotPanic verifies that a nil OnFinishReport
// does not panic.
func TestOnFinishReport_NilCallbackDoesNotPanic(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	// OnFinishReport intentionally left nil

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}
}

// TestOnStartReport_OrderBeforeFinish verifies that OnStartReport is called
// before OnFinishReport within a single Run().
func TestOnStartReport_OrderBeforeFinish(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())

	var order []string
	r.OnStartReport = func() { order = append(order, "start") }
	r.OnFinishReport = func() { order = append(order, "finish") }

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	if len(order) != 2 || order[0] != "start" || order[1] != "finish" {
		t.Errorf("event order = %v, want [start finish]", order)
	}
}
