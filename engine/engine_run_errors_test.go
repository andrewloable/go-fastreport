package engine

// engine_run_errors_test.go — internal tests (package engine) targeting the
// remaining uncovered branches in engine.go:
//
//   Run (73.3%):
//     lines 218-219  runPhase1 error → e.runFinished() + return "engine phase 1: …"
//     lines 227-228  runPhase2 error → e.runFinished() + return "engine phase 2: …"
//
//   runPhase1 (93.3%):
//     line 250  initializeData error return
//
//   runPhase2 (66.7%):
//     line 273  first runReportPages error return
//     line 284  double-pass second runReportPages error return
//
//   initParameters (84.6%):
//     the implicit-else branch of `if val, err := Calc(expr); err == nil`
//     (body skipped when Calc returns an error — p.Value left unchanged)
//
//   runReportPages (72.7%):
//     line 375-376  context-cancelled error return (between pages)
//     line 366-367  e.aborted break (mid-iteration, not just at RunReportPage level)
//
// Strategy for context-cancel tests: use AddStateHandler to cancel the context
// from inside the engine's own OnStateChanged call after the first page finishes.
// The handler fires synchronously during RunReportPage → endPage → OnStateChanged.
// When runReportPages then loops to the next page and checks e.ctx.Err(), it sees
// the cancellation and returns the error.
//
// Strategy for initParameters Calc-error: use an expression that the expr
// evaluator rejects (e.g. "!!!invalid!!!") — same technique used in
// sysvars_totals_coverage_test.go for accumulateTotals.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// newErrDataBand returns a *band.DataBand whose data source always fails First().
func newErrDataBand() *band.DataBand {
	db := band.NewDataBand()
	db.SetName("ErrFirstDB")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(&errFirstDS{})
	return db
}

// ── helpers ───────────────────────────────────────────────────────────────────

// newMultiPageReport returns a report with n empty ReportPages.
func newMultiPageReport(n int) *reportpkg.Report {
	r := reportpkg.NewReport()
	for i := 0; i < n; i++ {
		r.AddPage(reportpkg.NewReportPage())
	}
	return r
}

// cancelAfterFirstPage returns a cancel function and registers a state handler
// on e that calls cancel() after the first EngineStateReportPageFinished event.
func cancelAfterFirstPage(e *ReportEngine, cancel context.CancelFunc) {
	fired := false
	e.AddStateHandler(func(_ any, state EngineState) {
		if state == EngineStateReportPageFinished && !fired {
			fired = true
			cancel()
		}
	})
}

// ── initParameters: Calc failure — implicit-else branch ──────────────────────

// TestInitParameters_CalcFailure covers the branch inside initParameters where
// e.report.Calc(p.Expression) returns an error.  In that case the body of
// `if val, err := Calc(expr); err == nil { p.Value = val }` is NOT entered,
// leaving p.Value unchanged (nil).  Run must still succeed because initParameters
// never propagates the Calc error.
func TestInitParameters_CalcFailure(t *testing.T) {
	r := newMultiPageReport(1)
	dict := data.NewDictionary()

	bad := &data.Parameter{
		Name:       "BadExprParam",
		Expression: "!!!invalid!!!", // rejected by the expr evaluator
	}
	dict.AddParameter(bad)
	r.SetDictionary(dict)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with bad parameter expression should succeed: %v", err)
	}
	// p.Value should remain nil because Calc failed and the body was skipped.
	if bad.Value != nil {
		t.Errorf("bad.Value = %v, want nil (Calc error path)", bad.Value)
	}
}

// TestInitParameters_CalcFailure_NestedChild covers the same Calc-failure branch
// inside the recursive evalParams call for a child parameter.
func TestInitParameters_CalcFailure_NestedChild(t *testing.T) {
	r := newMultiPageReport(1)
	dict := data.NewDictionary()

	parent := &data.Parameter{
		Name:       "ValidParent",
		Expression: "1 + 1", // succeeds
	}
	child := &data.Parameter{
		Name:       "BadChild",
		Expression: "!!!bad!!!", // fails
	}
	parent.AddParameter(child)
	dict.AddParameter(parent)
	r.SetDictionary(dict)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with nested bad expression: %v", err)
	}
	if child.Value != nil {
		t.Errorf("child.Value = %v, want nil", child.Value)
	}
}

// ── runReportPages: context-cancelled error return (line 375-376) ────────────

// TestRunReportPages_ContextCancelledBetweenPages covers the
// `return fmt.Errorf("engine cancelled: %w", err)` path inside runReportPages.
// The context is cancelled by a state handler after the first page finishes;
// when the loop checks e.ctx.Err() before the second page, it returns the error.
func TestRunReportPages_ContextCancelledBetweenPages(t *testing.T) {
	r := newMultiPageReport(3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e := New(r)
	cancelAfterFirstPage(e, cancel)

	opts := DefaultRunOptions()
	opts.Context = ctx

	err := e.Run(opts)
	if err == nil {
		t.Fatal("expected error from context-cancelled runReportPages, got nil")
	}
	if !strings.Contains(err.Error(), "cancelled") && !strings.Contains(err.Error(), "cancel") {
		t.Errorf("error should mention cancellation: %v", err)
	}
}

// ── runPhase2: first-pass runReportPages error return (line 273) ─────────────

// TestRunPhase2_FirstPassError covers `return err` at line 273 in runPhase2
// (the first runReportPages call failing).  We call runPhase1 + runPhase2
// directly as internal methods so we can inject the context after phase 1
// completes, bypassing Run's own between-phases context check.
func TestRunPhase2_FirstPassError(t *testing.T) {
	r := newMultiPageReport(3)
	e := New(r)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancelAfterFirstPage(e, cancel)

	// Run phase 1 first (context is still valid here).
	if err := e.runPhase1(true); err != nil {
		t.Fatalf("runPhase1: %v", err)
	}
	// Set the (still-valid) context on the engine before calling runPhase2.
	e.ctx = ctx

	// runPhase2 → prepareToFirstPass (ok) → runReportPages → processes page 1
	// → state handler cancels ctx → loop checks ctx before page 2 → returns error
	// → runPhase2 returns that error at line 273.
	err := e.runPhase2(false)
	if err == nil {
		t.Fatal("expected error from runPhase2 first-pass runReportPages, got nil")
	}
}

// ── runPhase2: double-pass second-pass runReportPages error (line 284) ────────

// TestRunPhase2_SecondPassError covers `return err` at line 284 in runPhase2
// (the second runReportPages call in double-pass mode failing).
// A 2-page, DoublePass report fires EngineStateReportPageFinished twice in the
// first pass and twice in the second.  We cancel after the 3rd event (= first
// page of the second pass) so the second-pass runReportPages loop returns an
// error on its second iteration.
func TestRunPhase2_SecondPassError(t *testing.T) {
	r := newMultiPageReport(2)
	r.DoublePass = true

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e := New(r)

	eventCount := 0
	e.AddStateHandler(func(_ any, state EngineState) {
		if state == EngineStateReportPageFinished {
			eventCount++
			if eventCount == 3 { // first page of the second pass
				cancel()
			}
		}
	})

	if err := e.runPhase1(true); err != nil {
		t.Fatalf("runPhase1: %v", err)
	}
	e.ctx = ctx

	err := e.runPhase2(false)
	if err == nil {
		t.Fatal("expected error from runPhase2 second-pass runReportPages, got nil")
	}
}

// ── Run: phase-2 error path (lines 227-228) ──────────────────────────────────

// TestRun_Phase2ErrorPath covers the `e.runFinished(); return "engine phase 2: …"`
// branch in Run (lines 227-228).  Same cancel-after-first-page strategy but
// called through the public Run API.
func TestRun_Phase2ErrorPath(t *testing.T) {
	r := newMultiPageReport(3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e := New(r)
	cancelAfterFirstPage(e, cancel)

	opts := DefaultRunOptions()
	opts.Context = ctx

	err := e.Run(opts)
	if err == nil {
		t.Fatal("expected Run to return a phase-2 error, got nil")
	}
	if !strings.Contains(err.Error(), "phase 2") {
		t.Errorf("error should contain 'phase 2': %v", err)
	}
}

// ── Run: phase-1 error path (lines 218-219) ──────────────────────────────────

// TestRun_Phase1ErrorPath covers the `e.runFinished(); return "engine phase 1: …"`
// branch in Run (lines 218-219).  runPhase1 can only fail if initializeData
// returns an error.  initializeData iterates e.dataSources ([]*data.BaseDataSource).
// Because BaseDataSource.Init() always returns nil, we cannot make it fail via the
// public RegisterDataSource API.
//
// We therefore exercise this branch by calling runPhase1 directly after injecting
// an entry via a nil-pointer trick: if dataSources contains a nil *BaseDataSource,
// ds.Init() panics — that is not the right test.
//
// Instead we test that Run correctly surfaces a phase-1 error by temporarily
// wrapping the engine.  The only feasible approach with the current API is to
// confirm Run's phase-1 error path exists and is the right shape, using a
// direct runPhase1 call from a sub-test and then verifying the error message
// format that Run would produce.
//
// Note: line 331 in initializeData (the error-return branch) and lines 218-219
// in Run (the phase-1-error branch) are structurally unreachable with the current
// *data.BaseDataSource concrete type because BaseDataSource.Init() never errors.
// These lines require either an interface-typed dataSources slice or an error
// hook — neither of which the current implementation provides.
// We document this constraint and skip those two specific lines.
func TestRun_Phase1ErrorPath_DocumentedConstraint(t *testing.T) {
	t.Skip("initializeData error path is unreachable: BaseDataSource.Init() never errors " +
		"and dataSources is []*data.BaseDataSource (no interface polymorphism). " +
		"Lines 218-219 and 331 in engine.go cannot be covered without a hook or interface.")
}

// ── runReportPages: RunReportPage error return (line 379) ────────────────────

// errFirstDS implements band.DataSource and always fails on First().
// Setting this on a DataBand causes RunDataBandFull → ds.First() → error,
// which propagates: runBands → RunReportPage → runReportPages → return err (line 379).
type errFirstDS struct{}

func (d *errFirstDS) RowCount() int                  { return 1 }
func (d *errFirstDS) First() error                   { return errors.New("first() intentional failure") }
func (d *errFirstDS) Next() error                    { return errors.New("next() intentional failure") }
func (d *errFirstDS) EOF() bool                      { return false }
func (d *errFirstDS) GetValue(_ string) (any, error) { return nil, nil }

// TestRunReportPages_RunReportPageError covers `return err` at line 379
// in runReportPages, reached when RunReportPage returns a non-nil error.
func TestRunReportPages_RunReportPageError(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	// Add a DataBand with a data source that always errors on First().
	db := newErrDataBand()
	pg.AddBand(db)
	r.AddPage(pg)

	e := New(r)
	err := e.Run(DefaultRunOptions())
	if err == nil {
		t.Fatal("expected Run to return a RunReportPage error, got nil")
	}
}

// ── initParameters: nil-report and nil-dict guard returns ────────────────────

// TestInitParameters_NilReport covers `if e.report == nil { return }` (line 306).
func TestInitParameters_NilReport(t *testing.T) {
	e := &ReportEngine{} // report is nil
	// Must not panic; returns immediately on the nil-report guard.
	e.initParameters()
}

// TestInitParameters_NilDict covers `if dict == nil { return }` (line 310).
func TestInitParameters_NilDict(t *testing.T) {
	r := reportpkg.NewReport()
	r.SetDictionary(nil)
	e := New(r)
	// Must not panic; returns immediately on the nil-dict guard.
	e.initParameters()
}

// ── runReportPages: e.aborted break mid-iteration ────────────────────────────

// TestRunReportPages_AbortMidIteration covers the `if e.aborted { break }` path
// in runReportPages when Abort() is called after the first page — distinct from
// the RunReportPage-level abort which is checked inside runBands.
//
// With 3 pages the handler calls e.Abort() after page 1 finishes.
// runReportPages then sees e.aborted=true at the top of the loop for page 2
// and breaks, returning nil.
func TestRunReportPages_AbortMidIteration(t *testing.T) {
	r := newMultiPageReport(3)
	e := New(r)

	fired := false
	e.AddStateHandler(func(_ any, state EngineState) {
		if state == EngineStateReportPageFinished && !fired {
			fired = true
			e.Abort()
		}
	})

	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with mid-iteration abort: %v", err)
	}
	// Only page 1 was processed before abort.
	if e.TotalPages() != 1 {
		t.Errorf("TotalPages = %d, want 1 (aborted after page 1)", e.TotalPages())
	}
	if !e.Aborted() {
		t.Error("engine should report Aborted() = true")
	}
}
