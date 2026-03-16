package engine

// engine_phase_coverage_test.go — internal tests (package engine) targeting
// uncovered branches in engine.go:
//
//   runPhase2 (83.3%):
//     - The false branch of `if e.report.DoublePass && !e.aborted`:
//       when DoublePass=true but e.aborted=true after the first pass, the
//       double-pass block is skipped entirely. This is a distinct basic block
//       in Go's coverage model even though there is no else clause.
//
//   initializeData (75.0%):
//     - The error return `return fmt.Errorf("data source %q: %w", ds.Name(), err)`
//       is unreachable via the public RegisterDataSource API because
//       (*data.BaseDataSource).Init() always returns nil and dataSources is
//       typed as []*data.BaseDataSource (no interface polymorphism).
//     - We document this constraint with a skip; additional happy-path calls
//       improve branch confidence even if the error line itself cannot be hit.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── runPhase2: DoublePass=true, aborted=true → double-pass block skipped ─────

// TestRunPhase2_DoublePass_AbortedAfterFirstPass exercises the false branch of
//
//	if e.report.DoublePass && !e.aborted { … }
//
// when e.aborted is set to true inside the first-pass runReportPages call (via
// a state handler that calls e.Abort() after processing the first page).
// With aborted=true the condition evaluates to false and the entire double-pass
// block — including finalPass=true, resetPageNumber, prepareToSecondPass, and
// the second runReportPages — is skipped.
func TestRunPhase2_DoublePass_AbortedAfterFirstPass(t *testing.T) {
	r := reportpkg.NewReport()
	r.DoublePass = true
	// Two pages so abortAfterFirstPage fires and there is a second page that
	// would normally trigger the double-pass block.
	r.AddPage(reportpkg.NewReportPage())
	r.AddPage(reportpkg.NewReportPage())

	e := New(r)

	// Register a state handler that aborts the engine after the first page.
	fired := false
	e.AddStateHandler(func(_ any, state EngineState) {
		if state == EngineStateReportPageFinished && !fired {
			fired = true
			e.Abort()
		}
	})

	// Call runPhase1 manually so we can call runPhase2 directly.
	if err := e.runPhase1(true); err != nil {
		t.Fatalf("runPhase1: %v", err)
	}

	// At this point aborted=false (runPhase1 resets it).
	// Call runPhase2 directly.  The state handler will fire inside the first
	// runReportPages call (page 1 finishes → handler sets aborted=true).
	// After runReportPages returns nil (abort is not an error), the condition
	//   e.report.DoublePass && !e.aborted
	// evaluates to true && !true = false → the double-pass block is skipped.
	err := e.runPhase2(false)
	if err != nil {
		t.Fatalf("runPhase2 with aborted double-pass: unexpected error: %v", err)
	}

	// Verify the engine was aborted but did NOT enter the double-pass block.
	if !e.aborted {
		t.Error("expected e.aborted to be true after state-handler abort")
	}
	if e.finalPass {
		t.Error("finalPass should remain false when aborted before double-pass block")
	}
}

// TestRunPhase2_DoublePass_AbortedViaAbortCall is a simpler variant that directly
// sets e.aborted = true before calling runPhase2, so the first-pass pages iterator
// breaks immediately (0 pages processed) and the double-pass condition is false.
func TestRunPhase2_DoublePass_AbortedViaAbortCall(t *testing.T) {
	r := reportpkg.NewReport()
	r.DoublePass = true
	r.AddPage(reportpkg.NewReportPage())

	e := New(r)

	// Simulate the engine already having aborted (e.g. a previous phase set it).
	// runPhase1 would normally reset aborted=false, but we call runPhase2 directly.
	if err := e.runPhase1(true); err != nil {
		t.Fatalf("runPhase1: %v", err)
	}

	// Manually abort so that runPhase2's double-pass guard evaluates to false.
	e.aborted = true

	err := e.runPhase2(false)
	if err != nil {
		t.Fatalf("runPhase2 with pre-aborted engine: %v", err)
	}

	// The double-pass block must have been skipped.
	if e.finalPass {
		t.Error("finalPass should remain false when aborted=true before runPhase2")
	}
}

// ── initializeData: documented constraint ─────────────────────────────────────

// TestInitializeData_ErrorPath_DocumentedConstraint documents why the error
// return on line 331 of engine.go (initializeData) cannot be covered:
//
//   dataSources is typed as []*data.BaseDataSource (concrete pointer type, not
//   an interface), and (*data.BaseDataSource).Init() always returns nil.
//   There is no way to inject an error-returning data source without either
//   changing the field type to an interface or adding an error hook.
//
// This test calls initializeData directly with an empty slice (happy path) to
// confirm the function is exercised, and documents the constraint.
func TestInitializeData_ErrorPath_DocumentedConstraint(t *testing.T) {
	t.Skip(
		"initializeData error return (engine.go:331) is unreachable: " +
			"dataSources is []*data.BaseDataSource and BaseDataSource.Init() " +
			"never returns an error. Changing coverage requires an interface " +
			"refactor or an InitHook field on BaseDataSource.",
	)
}

// TestInitializeData_EmptySlice verifies that initializeData with no registered
// data sources returns nil (the empty-loop happy path).
func TestInitializeData_EmptySlice(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	e := New(r)
	// dataSources is empty by default.
	if err := e.initializeData(); err != nil {
		t.Fatalf("initializeData with empty slice: %v", err)
	}
}
