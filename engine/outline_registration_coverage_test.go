package engine

// outline_registration_coverage_test.go — internal tests for uncovered branches:
//   • outline.go: nil preparedPages guard paths in AddOutline, OutlineRoot,
//     OutlineUp, GetBookmarkPage  (require direct field access → package engine)
//   • prepare_registration.go: runEngine with a non-BaseDataSource dictionary
//     entry (the if-ok=false branch of the type-assertion loop)

import (
	"context"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── outline.go: nil preparedPages guard paths ─────────────────────────────────

// newEngineNilPages returns an engine whose preparedPages is forcibly nil.
// engine.New() initialises preparedPages to preview.New(), so we must zero it
// afterwards using direct field access (only possible from package engine).
func newEngineNilPages(t *testing.T) *ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	e := New(r)
	e.preparedPages = nil // set the unexported field directly
	return e
}

func TestAddOutline_NilPreparedPages_NoOp(t *testing.T) {
	e := newEngineNilPages(t)
	// Must not panic; the nil guard should cause an early return.
	e.AddOutline("any text")
}

func TestOutlineRoot_NilPreparedPages_NoOp(t *testing.T) {
	e := newEngineNilPages(t)
	e.OutlineRoot()
}

func TestOutlineUp_NilPreparedPages_NoOp(t *testing.T) {
	e := newEngineNilPages(t)
	e.OutlineUp()
}

func TestGetBookmarkPage_NilPreparedPages_ReturnsZero(t *testing.T) {
	e := newEngineNilPages(t)
	got := e.GetBookmarkPage("anything")
	if got != 0 {
		t.Errorf("GetBookmarkPage with nil preparedPages = %d, want 0", got)
	}
}

// ── prepare_registration.go: runEngine error return path ────────────────────

// TestRunEngine_CancelledContext exercises the `return nil, err` branch in
// runEngine by providing a pre-cancelled context via PrepareWithContext.
// eng.Run() returns an error because context.Err() != nil, which causes
// runEngine to propagate the error instead of returning PreparedPages.
func TestRunEngine_CancelledContext(t *testing.T) {
	r := reportpkg.NewReport()
	// Add multiple pages so cancellation is more likely to be detected.
	r.AddPage(reportpkg.NewReportPage())
	r.AddPage(reportpkg.NewReportPage())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel so context.Err() returns immediately inside Run

	err := r.PrepareWithContext(ctx)
	if err == nil {
		t.Error("expected error from PrepareWithContext with cancelled context, got nil")
	}
}

// ── prepare_registration.go: non-BaseDataSource in dictionary ────────────────

// TestRunEngine_NonBaseDataSourceInDictionary exercises the false branch of the
// type-assertion `if bds, ok := ds.(*data.BaseDataSource); ok` inside runEngine.
// A VirtualDataSource satisfies the DataSource interface but is not a
// *BaseDataSource, so the assertion fails and RegisterDataSource is not called.
func TestRunEngine_NonBaseDataSourceInDictionary(t *testing.T) {
	r := reportpkg.NewReport()

	// VirtualDataSource is a DataSource but NOT a *BaseDataSource.
	vds := data.NewVirtualDataSource("VirtDS", 3)
	r.Dictionary().AddDataSource(vds)

	r.AddPage(reportpkg.NewReportPage())

	// r.Prepare() → runEngine → loop skips vds (type assertion ok=false)
	if err := r.Prepare(); err != nil {
		t.Fatalf("r.Prepare with non-BaseDataSource: %v", err)
	}
	if r.PreparedPages() == nil {
		t.Error("PreparedPages should not be nil after Prepare")
	}
}

// TestRunEngine_NonBaseDataSourceWithContext exercises the same branch via the
// context-aware path (PrepareWithContext → runEngine).
func TestRunEngine_NonBaseDataSourceWithContext(t *testing.T) {
	r := reportpkg.NewReport()

	vds := data.NewVirtualDataSource("VirtCtxDS", 2)
	r.Dictionary().AddDataSource(vds)

	r.AddPage(reportpkg.NewReportPage())

	ctx := context.Background()
	if err := r.PrepareWithContext(ctx); err != nil {
		t.Fatalf("r.PrepareWithContext with non-BaseDataSource: %v", err)
	}
}

// TestRunEngine_MixedDataSourcesInDictionary exercises both the ok=true and
// ok=false branches in the same runEngine call.
func TestRunEngine_MixedDataSourcesInDictionary(t *testing.T) {
	r := reportpkg.NewReport()

	// BaseDataSource → ok=true, RegisterDataSource is called.
	bds := data.NewBaseDataSource("RealDS")
	bds.SetAlias("RealDS")
	bds.AddColumn(data.Column{Name: "Val"})
	r.Dictionary().AddDataSource(bds)

	// VirtualDataSource → ok=false, RegisterDataSource is NOT called.
	vds := data.NewVirtualDataSource("VirtDS2", 1)
	r.Dictionary().AddDataSource(vds)

	r.AddPage(reportpkg.NewReportPage())

	if err := r.Prepare(); err != nil {
		t.Fatalf("r.Prepare with mixed data sources: %v", err)
	}
}
