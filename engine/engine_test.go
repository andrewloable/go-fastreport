package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func newEngine() (*engine.ReportEngine, *reportpkg.Report) {
	r := reportpkg.NewReport()
	return engine.New(r), r
}

func newEngineWithPage() (*engine.ReportEngine, *reportpkg.Report) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	return engine.New(r), r
}

// ── constructor ───────────────────────────────────────────────────────────────

func TestNew_NotNil(t *testing.T) {
	e, _ := newEngine()
	if e == nil {
		t.Fatal("engine.New returned nil")
	}
}

func TestNew_DefaultProperties(t *testing.T) {
	e, _ := newEngine()

	if e.CurX() != 0 {
		t.Errorf("CurX default = %v, want 0", e.CurX())
	}
	if e.CurY() != 0 {
		t.Errorf("CurY default = %v, want 0", e.CurY())
	}
	if e.CurColumn() != 0 {
		t.Errorf("CurColumn default = %d, want 0", e.CurColumn())
	}
	if e.PageNo() != 1 {
		t.Errorf("PageNo default = %d, want 1", e.PageNo())
	}
	if e.RowNo() != 1 {
		t.Errorf("RowNo default = %d, want 1", e.RowNo())
	}
	if e.AbsRowNo() != 1 {
		t.Errorf("AbsRowNo default = %d, want 1", e.AbsRowNo())
	}
	if e.FinalPass() {
		t.Error("FinalPass should default to false")
	}
	if !e.FirstPass() {
		t.Error("FirstPass should default to true")
	}
	if e.HierarchyLevel() != 0 {
		t.Errorf("HierarchyLevel default = %d, want 0", e.HierarchyLevel())
	}
	if e.HierarchyRowNo() != "" {
		t.Errorf("HierarchyRowNo default = %q, want empty", e.HierarchyRowNo())
	}
	if e.Aborted() {
		t.Error("Aborted should default to false")
	}
}

// ── SetCurX / SetCurY ────────────────────────────────────────────────────────

func TestSetCurX(t *testing.T) {
	e, _ := newEngine()
	e.SetCurX(42.5)
	if e.CurX() != 42.5 {
		t.Errorf("CurX = %v, want 42.5", e.CurX())
	}
}

func TestSetCurY(t *testing.T) {
	e, _ := newEngine()
	e.SetCurY(100)
	if e.CurY() != 100 {
		t.Errorf("CurY = %v, want 100", e.CurY())
	}
}

// ── Abort ─────────────────────────────────────────────────────────────────────

func TestAbort(t *testing.T) {
	e, _ := newEngine()
	e.Abort()
	if !e.Aborted() {
		t.Error("Aborted should be true after Abort()")
	}
}

// ── DefaultRunOptions ─────────────────────────────────────────────────────────

func TestDefaultRunOptions(t *testing.T) {
	opts := engine.DefaultRunOptions()
	if !opts.ResetDataState {
		t.Error("ResetDataState should be true in DefaultRunOptions")
	}
	if opts.Append {
		t.Error("Append should be false in DefaultRunOptions")
	}
	if opts.MaxPages != 0 {
		t.Errorf("MaxPages default = %d, want 0", opts.MaxPages)
	}
}

// ── Run – single pass ─────────────────────────────────────────────────────────

func TestRun_EmptyReport(t *testing.T) {
	e, _ := newEngine()
	opts := engine.DefaultRunOptions()
	if err := e.Run(opts); err != nil {
		t.Fatalf("Run on empty report failed: %v", err)
	}
}

func TestRun_SinglePage(t *testing.T) {
	e, _ := newEngineWithPage()
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	// processPage increments totalPages once per page.
	if e.TotalPages() != 1 {
		t.Errorf("TotalPages = %d, want 1", e.TotalPages())
	}
}

func TestRun_MultiplePages(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	r.AddPage(reportpkg.NewReportPage())
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if e.TotalPages() != 3 {
		t.Errorf("TotalPages = %d, want 3", e.TotalPages())
	}
}

func TestRun_InitialPageNumber(t *testing.T) {
	r := reportpkg.NewReport()
	r.InitialPageNumber = 5
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	// After running one page, PageNo should be InitialPageNumber+1.
	if e.PageNo() != 6 {
		t.Errorf("PageNo = %d, want 6", e.PageNo())
	}
}

// ── Run – double pass ─────────────────────────────────────────────────────────

func TestRun_DoublePass(t *testing.T) {
	r := reportpkg.NewReport()
	r.DoublePass = true
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run double pass error: %v", err)
	}
	// Two passes × 1 page each = 2 total pages counted.
	if e.TotalPages() != 2 {
		t.Errorf("TotalPages after double pass = %d, want 2", e.TotalPages())
	}
}

func TestRun_DoublePass_FinalPassTrue(t *testing.T) {
	r := reportpkg.NewReport()
	r.DoublePass = true
	e := engine.New(r)
	// After run the engine stays in FinalPass state.
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if !e.FinalPass() {
		t.Error("FinalPass should be true after double-pass run")
	}
	if e.FirstPass() {
		t.Error("FirstPass should be false after final pass")
	}
}

// ── MaxPages limit ────────────────────────────────────────────────────────────

func TestRun_MaxPages(t *testing.T) {
	r := reportpkg.NewReport()
	for i := 0; i < 5; i++ {
		r.AddPage(reportpkg.NewReportPage())
	}
	e := engine.New(r)
	opts := engine.DefaultRunOptions()
	opts.MaxPages = 3
	if err := e.Run(opts); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if e.TotalPages() != 3 {
		t.Errorf("TotalPages = %d, want 3 (MaxPages limit)", e.TotalPages())
	}
}

// ── PageWidth / PageHeight / FreeSpace ───────────────────────────────────────

func TestRun_PageDimensions(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	// A4: 210×297 mm, margins 10 mm each side.
	// Usable: 190 mm wide, 277 mm tall at 3.78 px/mm (matching C# Units.Millimeters).
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	const mmPerPx = float32(3.78)
	wantW := (210 - 10 - 10) * mmPerPx // 190 * 3.78 = 718.2
	wantH := (297 - 10 - 10) * mmPerPx // 277 * 3.78 = 1047.06
	if e.PageWidth() != wantW {
		t.Errorf("PageWidth = %v, want %v", e.PageWidth(), wantW)
	}
	if e.PageHeight() != wantH {
		t.Errorf("PageHeight = %v, want %v", e.PageHeight(), wantH)
	}
}

// ── AdvanceY ──────────────────────────────────────────────────────────────────

func TestAdvanceY(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	initialFree := e.FreeSpace()
	e.AdvanceY(50)
	if e.CurY() != 50 {
		t.Errorf("CurY = %v, want 50", e.CurY())
	}
	if e.FreeSpace() != initialFree-50 {
		t.Errorf("FreeSpace = %v, want %v", e.FreeSpace(), initialFree-50)
	}
}

func TestAdvanceY_ClampsFreeSpaceAtZero(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())
	// Advance beyond page height.
	e.AdvanceY(e.PageHeight() + 1000)
	if e.FreeSpace() != 0 {
		t.Errorf("FreeSpace should be clamped to 0, got %v", e.FreeSpace())
	}
}

// ── NewPage ───────────────────────────────────────────────────────────────────

func TestNewPage(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	prev := e.TotalPages()
	e.SetCurY(100)
	e.NewPage()

	if e.CurY() != 0 {
		t.Errorf("CurY after NewPage = %v, want 0", e.CurY())
	}
	if e.CurX() != 0 {
		t.Errorf("CurX after NewPage = %v, want 0", e.CurX())
	}
	if e.TotalPages() != prev+1 {
		t.Errorf("TotalPages after NewPage = %d, want %d", e.TotalPages(), prev+1)
	}
}

// ── Data source registration ──────────────────────────────────────────────────

func TestRegisterDataSource(t *testing.T) {
	e, _ := newEngine()
	ds := data.NewBaseDataSource("Products")

	e.RegisterDataSource(ds)
	if len(e.DataSources()) != 1 {
		t.Errorf("DataSources len = %d, want 1", len(e.DataSources()))
	}
	if e.DataSources()[0].Name() != "Products" {
		t.Errorf("DataSources[0].Name = %q, want Products", e.DataSources()[0].Name())
	}
}

func TestRegisterMultipleDataSources(t *testing.T) {
	e, _ := newEngine()
	e.RegisterDataSource(data.NewBaseDataSource("A"))
	e.RegisterDataSource(data.NewBaseDataSource("B"))
	if len(e.DataSources()) != 2 {
		t.Errorf("DataSources len = %d, want 2", len(e.DataSources()))
	}
}

// ── Date ──────────────────────────────────────────────────────────────────────

func TestRun_SetsDate(t *testing.T) {
	e, _ := newEngine()
	before := e.Date()
	_ = e.Run(engine.DefaultRunOptions())
	after := e.Date()
	if !after.After(before) && after != before {
		t.Errorf("Date not updated by Run: before=%v after=%v", before, after)
	}
	if after.IsZero() {
		t.Error("Date should not be zero after Run")
	}
}

// ── Append mode ───────────────────────────────────────────────────────────────

func TestRun_AppendMode_DoesNotResetCounters(t *testing.T) {
	r := reportpkg.NewReport()
	r.AddPage(reportpkg.NewReportPage())
	e := engine.New(r)

	// First run.
	_ = e.Run(engine.DefaultRunOptions())
	afterFirst := e.TotalPages()

	// Second run in append mode.
	opts := engine.DefaultRunOptions()
	opts.Append = true
	_ = e.Run(opts)
	afterSecond := e.TotalPages()

	// In append mode the page counter continues from where it left off.
	if afterSecond <= afterFirst {
		t.Errorf("Append mode should accumulate pages: after1=%d after2=%d", afterFirst, afterSecond)
	}
}
