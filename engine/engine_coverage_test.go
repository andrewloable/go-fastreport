package engine_test

// engine_coverage_test.go — targeted coverage for uncovered branches in
// engine.go and prepare_registration.go: Run with Context, initParameters
// with nested params, initializeData with registered data sources,
// runReportPages context cancellation.

import (
	"context"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── Run: Context option ───────────────────────────────────────────────────────

// TestRun_WithContext exercises the opts.Context != nil branch in Run.
func TestRun_WithContext(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)

	ctx := context.Background()
	opts := engine.DefaultRunOptions()
	opts.Context = ctx

	if err := e.Run(opts); err != nil {
		t.Fatalf("Run with context: %v", err)
	}
}

// TestRun_WithCancelledContext exercises the context-cancelled path in Run.
// A cancelled context causes Run to return an error after phase 1.
func TestRun_WithCancelledContext(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel before Run so context.Err() returns immediately.
	cancel()

	opts := engine.DefaultRunOptions()
	opts.Context = ctx

	// Run should return an error because the context is already cancelled.
	err := e.Run(opts)
	if err == nil {
		t.Error("expected error with cancelled context, got nil")
	}
}

// ── initializeData: with registered data sources ──────────────────────────────

// TestRun_WithRegisteredDataSource exercises initializeData when at least one
// data source is registered — this covers the loop body with ds.Init() calls.
func TestRun_WithRegisteredDataSource(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)

	ds := data.NewBaseDataSource("TestDS")
	ds.SetAlias("TestDS")
	ds.AddColumn(data.Column{Name: "Val"})
	ds.AddRow(map[string]any{"Val": 1})
	e.RegisterDataSource(ds)

	opts := engine.DefaultRunOptions()
	if err := e.Run(opts); err != nil {
		t.Fatalf("Run with registered data source: %v", err)
	}

	if len(e.DataSources()) != 1 {
		t.Errorf("DataSources len = %d, want 1", len(e.DataSources()))
	}
}

// TestRun_MultipleDataSources exercises initializeData with multiple
// registered data sources to cover iteration of all of them.
func TestRun_MultipleDataSources(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)

	for i := 0; i < 3; i++ {
		ds := data.NewBaseDataSource("DS")
		ds.SetAlias("DS")
		ds.AddColumn(data.Column{Name: "Val"})
		e.RegisterDataSource(ds)
	}

	opts := engine.DefaultRunOptions()
	if err := e.Run(opts); err != nil {
		t.Fatalf("Run with multiple data sources: %v", err)
	}
}

// ── initParameters: nested child parameters ───────────────────────────────────

// TestRun_WithNestedParameters exercises the recursive evalParams function in
// initParameters, specifically the child-parameter recursion path.
func TestRun_WithNestedParameters(t *testing.T) {
	r := reportpkg.NewReport()
	dict := data.NewDictionary()

	// Parent parameter with an expression.
	parent := &data.Parameter{
		Name:       "Parent",
		Expression: "1 + 1",
	}

	// Child parameter with its own expression — triggers recursive evalParams.
	child := &data.Parameter{
		Name:       "Child",
		Expression: "2 + 2",
	}
	parent.AddParameter(child)

	// Grandchild parameter — exercises deeper recursion.
	grandchild := &data.Parameter{
		Name:       "Grandchild",
		Expression: "",
	}
	child.AddParameter(grandchild)

	dict.AddParameter(parent)
	r.SetDictionary(dict)
	r.AddPage(reportpkg.NewReportPage())

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with nested parameters: %v", err)
	}
}

// TestRun_WithParameterExpression exercises the expression-evaluation branch in
// initParameters where p.Expression != "".
func TestRun_WithParameterExpression(t *testing.T) {
	r := reportpkg.NewReport()
	dict := data.NewDictionary()

	p := &data.Parameter{
		Name:       "Computed",
		Expression: "1 + 1",
	}
	dict.AddParameter(p)
	r.SetDictionary(dict)
	r.AddPage(reportpkg.NewReportPage())

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with parameter expression: %v", err)
	}
}

// ── runReportPages: context cancellation between pages ───────────────────────

// TestRun_ContextCancelledBetweenPages exercises the context check in
// runReportPages that runs between each page iteration. We use a context that
// is cancelled via the OnStateChanged mechanism indirectly — but the simplest
// way is to pre-cancel the context and add multiple pages.
func TestRun_ContextCancelledBetweenPages(t *testing.T) {
	r := reportpkg.NewReport()
	// Add 5 pages — with a cancelled context, it should stop early.
	for i := 0; i < 5; i++ {
		r.AddPage(reportpkg.NewReportPage())
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Don't cancel yet — we'll cancel after first page by using a very short timeout.
	cancel() // Pre-cancel: context.Err() will return immediately during runReportPages.

	e := engine.New(r)
	opts := engine.DefaultRunOptions()
	opts.Context = ctx

	// Should return error because context is cancelled.
	err := e.Run(opts)
	if err == nil {
		t.Error("expected error with pre-cancelled context in multi-page run, got nil")
	}
}

// TestRun_WithDeadlineExceeded exercises the context deadline path.
func TestRun_WithDeadlineExceeded(t *testing.T) {
	r := reportpkg.NewReport()
	for i := 0; i < 3; i++ {
		r.AddPage(reportpkg.NewReportPage())
	}

	// Create an already-expired deadline context.
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	e := engine.New(r)
	opts := engine.DefaultRunOptions()
	opts.Context = ctx

	err := e.Run(opts)
	if err == nil {
		t.Error("expected error with expired deadline context, got nil")
	}
}

// ── runEngine (prepare_registration.go): with dictionary data sources ─────────

// TestRunEngine_WithDictionaryDataSources exercises the data-source registration
// loop in runEngine that iterates r.Dictionary().DataSources() for BaseDataSource.
func TestRunEngine_WithDictionaryDataSources(t *testing.T) {
	r := reportpkg.NewReport()

	ds := data.NewBaseDataSource("DictDS")
	ds.SetAlias("DictDS")
	ds.AddColumn(data.Column{Name: "Val"})
	ds.AddRow(map[string]any{"Val": 99})
	r.Dictionary().AddDataSource(ds)

	r.AddPage(reportpkg.NewReportPage())

	// r.Prepare() uses runEngine which registers BaseDataSource from dictionary.
	if err := r.Prepare(); err != nil {
		t.Fatalf("r.Prepare with dictionary data source: %v", err)
	}
	if r.PreparedPages() == nil {
		t.Error("PreparedPages should not be nil after Prepare")
	}
}

// TestRunEngine_WithContext exercises the PrepareWithContext path in runEngine.
func TestRunEngine_WithContext(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())

	ds := data.NewBaseDataSource("CtxDS")
	ds.SetAlias("CtxDS")
	r.Dictionary().AddDataSource(ds)

	ctx := context.Background()
	if err := r.PrepareWithContext(ctx); err != nil {
		t.Fatalf("r.PrepareWithContext with dict data source: %v", err)
	}
}

// ── runPhase2: double-pass + context combination ──────────────────────────────

// TestRun_DoublePass_WithContext exercises runPhase2 with both double-pass and
// context to ensure the finalPass/resetPageNumber path is hit.
func TestRun_DoublePass_WithContext(t *testing.T) {
	r := reportpkg.NewReport()
	r.DoublePass = true
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)

	ctx := context.Background()
	opts := engine.DefaultRunOptions()
	opts.Context = ctx

	if err := e.Run(opts); err != nil {
		t.Fatalf("double-pass with context: %v", err)
	}
	if !e.FinalPass() {
		t.Error("FinalPass should be true after double-pass run")
	}
}

// TestRun_ResetDataState_False exercises the resetDataState=false branch in
// runPhase1, skipping initializeData.
func TestRun_ResetDataState_False(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)

	opts := engine.RunOptions{
		ResetDataState: false,
		MaxPages:       0,
	}
	if err := e.Run(opts); err != nil {
		t.Fatalf("Run with ResetDataState=false: %v", err)
	}
}

// TestRun_WithStartReportEvent exercises the non-empty StartReportEvent path.
func TestRun_WithStartReportEvent(t *testing.T) {
	r := reportpkg.NewReport()
	r.StartReportEvent = "OnStartReport" // non-empty, but no-op in current impl
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)

	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with StartReportEvent: %v", err)
	}
}

// TestRun_WithFinishReportEvent exercises the non-empty FinishReportEvent path.
func TestRun_WithFinishReportEvent(t *testing.T) {
	r := reportpkg.NewReport()
	r.FinishReportEvent = "OnFinishReport" // non-empty, but no-op in current impl
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)

	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with FinishReportEvent: %v", err)
	}
}
