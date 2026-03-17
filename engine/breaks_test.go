package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// newRunningEngine creates an engine that has completed a Run() so
// PreparedPages and page state are initialised.
func newRunningEngine(t *testing.T) *engine.ReportEngine {
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

// ── BandHasHardPageBreaks ─────────────────────────────────────────────────────

func TestBandHasHardPageBreaks_NoBreaks(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	b := band.NewBandBase()
	if e.BandHasHardPageBreaks(b) {
		t.Error("empty band should have no hard page breaks")
	}
}

func TestBandHasHardPageBreaks_WithBreak(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	b := band.NewBandBase()

	// Add a component with PageBreak=true.
	rc := report.NewReportComponentBase()
	rc.SetPageBreak(true)
	b.AddChild(rc)

	if !e.BandHasHardPageBreaks(b) {
		t.Error("band should have hard page break")
	}
}

func TestBandHasHardPageBreaks_WithoutPageBreakComponent(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	b := band.NewBandBase()

	rc := report.NewReportComponentBase()
	rc.SetPageBreak(false)
	b.AddChild(rc)

	if e.BandHasHardPageBreaks(b) {
		t.Error("component without PageBreak should not trigger hard break")
	}
}

// ── SplitHardPageBreaks ───────────────────────────────────────────────────────

func TestSplitHardPageBreaks_NoBreaks(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	b := band.NewBandBase()
	b.SetHeight(100)

	parts := e.SplitHardPageBreaks(b)
	if len(parts) != 1 {
		t.Errorf("no breaks: expected 1 part, got %d", len(parts))
	}
	if parts[0] != b {
		t.Error("no breaks: returned part should be original band")
	}
}

func TestSplitHardPageBreaks_OneBreak(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	b := band.NewBandBase()
	b.SetHeight(100)
	b.SetName("TestBand")

	rc := report.NewReportComponentBase()
	rc.SetTop(40)
	rc.SetPageBreak(true)
	b.AddChild(rc)

	parts := e.SplitHardPageBreaks(b)
	// C# behaviour: when the only object has PageBreak=true, a single part is
	// created starting at that object's position with StartNewPage=true.
	// There is no "before" part because there are no objects before the break.
	if len(parts) != 1 {
		t.Fatalf("1 break (only PageBreak obj): expected 1 part, got %d", len(parts))
	}
	if parts[0].Height() != 60 {
		t.Errorf("part[0] height = %v, want 60", parts[0].Height())
	}
	if !parts[0].StartNewPage() {
		t.Error("part[0] should have StartNewPage=true")
	}
}

func TestSplitHardPageBreaks_OneBreak_WithObjsBefore(t *testing.T) {
	e := engine.New(reportpkg.NewReport())
	b := band.NewBandBase()
	b.SetHeight(100)
	b.SetName("TestBand2")

	// Object before the break.
	obj1 := report.NewReportComponentBase()
	obj1.SetTop(10)
	obj1.SetHeight(15)
	b.AddChild(obj1)

	// PageBreak object.
	rc := report.NewReportComponentBase()
	rc.SetTop(40)
	rc.SetPageBreak(true)
	b.AddChild(rc)

	parts := e.SplitHardPageBreaks(b)
	// C# behaviour: 2 parts. First part has objects before the break,
	// second part starts at the break with StartNewPage=true.
	if len(parts) != 2 {
		t.Fatalf("1 break with objs before: expected 2 parts, got %d", len(parts))
	}
	if parts[0].Height() != 40 {
		t.Errorf("part[0] height = %v, want 40", parts[0].Height())
	}
	if parts[0].StartNewPage() {
		t.Error("part[0] should NOT have StartNewPage=true")
	}
	if parts[1].Height() != 60 {
		t.Errorf("part[1] height = %v, want 60", parts[1].Height())
	}
	if !parts[1].StartNewPage() {
		t.Error("part[1] should have StartNewPage=true")
	}
}

// ── BreakBand ─────────────────────────────────────────────────────────────────

func TestBreakBand_CanBreakSplitsBand(t *testing.T) {
	e := newRunningEngine(t)

	// Consume most of the page so only 50px remain.
	toConsume := e.FreeSpace() - 50
	if toConsume > 0 {
		e.AdvanceY(toConsume)
	}

	b := band.NewBandBase()
	b.SetName("BreakTest")
	b.SetHeight(200) // larger than remaining 50px
	b.SetVisible(true)
	b.SetCanBreak(true)

	pp := e.PreparedPages()
	pg0Count := len(pp.GetPage(0).Bands)

	e.BreakBand(b)

	// The top portion should have been added to the current page.
	pg0 := pp.GetPage(0)
	if len(pg0.Bands) <= pg0Count {
		t.Error("BreakBand should have added a portion of the band to the current page")
	}
}

func TestBreakBand_NoBreak_NewPage(t *testing.T) {
	e := newRunningEngine(t)

	// Consume all but 10px of free space.
	toConsume := e.FreeSpace() - 10
	if toConsume > 0 {
		e.AdvanceY(toConsume)
	}

	b := band.NewBandBase()
	b.SetName("NoBreakTest")
	b.SetHeight(200)
	b.SetVisible(true)
	b.SetCanBreak(false)

	initialPages := e.PreparedPages().Count()
	e.BreakBand(b)

	// A new page should have been started.
	if e.PreparedPages().Count() <= initialPages {
		t.Error("BreakBand (no-break) should have started a new page")
	}
}

// ── ShowBandWithPageBreaks ────────────────────────────────────────────────────

func TestShowBandWithPageBreaks_NoBreak(t *testing.T) {
	e := newRunningEngine(t)

	b := band.NewBandBase()
	b.SetName("Simple")
	b.SetHeight(30)
	b.SetVisible(true)

	beforeY := e.CurY()
	e.ShowBandWithPageBreaks(b)
	if e.CurY() != beforeY+30 {
		t.Errorf("CurY = %v, want %v", e.CurY(), beforeY+30)
	}
}
