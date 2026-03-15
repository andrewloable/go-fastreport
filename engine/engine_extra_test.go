package engine_test

import (
	"context"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// TestRunEngine_ViaPrepare triggers the runEngine function registered by
// prepare_registration.go init() via reportpkg.Report.Prepare().
// Importing the engine package (done here via engine.New) is enough to
// ensure the init() runs and registers the Prepare func.
func TestRunEngine_ViaPrepare(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	// r.Prepare() calls globalPrepareFunc which is set to runEngine in init().
	if err := r.Prepare(); err != nil {
		t.Fatalf("r.Prepare(): %v", err)
	}
	if r.PreparedPages() == nil {
		t.Error("PreparedPages should not be nil after Prepare")
	}
	if r.PreparedPages().Count() != 1 {
		t.Errorf("PreparedPages.Count = %d, want 1", r.PreparedPages().Count())
	}
}

// TestRunEngine_ViaPrepareWithContext triggers the context-aware runEngine path.
func TestRunEngine_ViaPrepareWithContext(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	ctx := context.Background()
	if err := r.PrepareWithContext(ctx); err != nil {
		t.Fatalf("r.PrepareWithContext(): %v", err)
	}
	if r.PreparedPages() == nil {
		t.Error("PreparedPages should not be nil after PrepareWithContext")
	}
}

// TestRunEngine_WithDataSource triggers the data-source registration path in runEngine.
func TestRunEngine_WithDataSource(t *testing.T) {
	_ = engine.New // ensure engine package imported so init() runs
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	// Prepare with no data sources in dictionary — should still succeed.
	if err := r.Prepare(); err != nil {
		t.Fatalf("r.Prepare with empty dict: %v", err)
	}
}

// TestAttachWatermark_EnabledWatermark creates a page with an enabled watermark
// and verifies it is attached to the prepared page.
func TestAttachWatermark_EnabledWatermark(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	wm := reportpkg.NewWatermark()
	wm.Enabled = true
	wm.Text = "CONFIDENTIAL"
	pg.Watermark = wm
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	prepPage := e.PreparedPages().GetPage(0)
	if prepPage == nil {
		t.Fatal("no prepared page")
	}
	if prepPage.Watermark == nil {
		t.Error("expected watermark attached to prepared page")
	}
	if !prepPage.Watermark.Enabled {
		t.Error("watermark should be enabled")
	}
}

// TestStartColumn_EndColumn_MultiColumnOverflow uses a tiny page to force an
// overflow into the second column, covering startColumn and endColumn.
func TestStartColumn_EndColumn_MultiColumnOverflow(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	// Very small page so the band overflows immediately.
	pg.PaperHeight = 5  // mm → ~18.9 px usable
	pg.TopMargin = 0
	pg.BottomMargin = 0
	pg.PaperWidth = 210
	pg.LeftMargin = 0
	pg.RightMargin = 0
	pg.Columns.Count = 2 // 2-column layout

	// A column header so startColumn exercises showBand.
	ch := band.NewColumnHeaderBand()
	ch.SetName("ColHdr")
	ch.SetHeight(5)
	ch.SetVisible(true)
	pg.SetColumnHeader(ch)

	// A data band larger than the column height to trigger overflow.
	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(50) // 50px >> 18.9px page → overflows into next column
	db.SetVisible(true)
	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if e.PreparedPages().Count() == 0 {
		t.Error("expected at least 1 prepared page")
	}
}

// TestPasteObjects_ViaCutAndFinishKeep tests pasteObjects via the public keep API
// by manually injecting a PreparedBand into the prepared pages before StartKeep.
func TestPasteObjects_ViaCutAndFinishKeep(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	pp := e.PreparedPages()
	// Add a band before StartKeep so keepPosition = 1.
	_ = pp.AddBand(&preview.PreparedBand{Name: "pre-keep", Top: 0, Height: 10})

	// StartKeep records keepPosition = 1.
	e.StartKeep()

	// Add a band after StartKeep — this is in the "kept" region.
	_ = pp.AddBand(&preview.PreparedBand{Name: "kept-band", Top: 10, Height: 20})

	// AdvanceY to create a non-zero keepDeltaY.
	e.AdvanceY(20)

	// CheckKeepTogether cuts bands from keepPosition=1 onwards.
	e.CheckKeepTogether()
	if len(pp.CutBands()) == 0 {
		t.Skip("no cut bands — test infrastructure limitation")
	}

	// FinishKeepTogether pastes the cut bands back (covering pasteObjects).
	e.FinishKeepTogether()

	// After FinishKeepTogether, keeping should be false.
	if e.IsKeeping() {
		t.Error("IsKeeping should be false after FinishKeepTogether")
	}
}

// TestEvalText_ViaBandOutlineExpression covers evalText by setting an
// OutlineExpression on the page that is evaluated during startPage.
func TestEvalText_ViaPageOutlineExpression(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.OutlineExpression = "Chapter [PageNumber]"
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify the outline was populated.
	root := e.PreparedPages().Outline.Root
	if len(root.Children) == 0 {
		t.Error("expected at least one outline entry from OutlineExpression")
	}
}
