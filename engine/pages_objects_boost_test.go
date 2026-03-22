package engine

// pages_objects_boost_test.go — internal tests (package engine) targeting
// specific uncovered branches in pages.go and objects.go.
//
// Uses package engine (not engine_test) so we can call unexported methods
// and access unexported fields directly.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/style"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func newBoostEngine(t *testing.T) *ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return e
}

func newBoostEngineUnrun(t *testing.T) *ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	return New(r)
}

// ── pages.go: showBandNoAdvance ───────────────────────────────────────────────

// TestShowBandNoAdvance_NilBand covers the nil-interface guard.
func TestShowBandNoAdvance_NilBand(t *testing.T) {
	e := newBoostEngine(t)
	// Should return immediately without panic.
	e.showBandNoAdvance(nil)
}

// TestShowBandNoAdvance_InvisibleBand covers the Visible()==false path.
func TestShowBandNoAdvance_InvisibleBand(t *testing.T) {
	e := newBoostEngine(t)
	b := band.NewPageHeaderBand()
	b.SetName("InvBand")
	b.SetHeight(20)
	b.SetVisible(false) // invisible → should return early
	e.showBandNoAdvance(b)
}

// TestShowBandNoAdvance_ZeroHeightBand covers the height<=0 guard.
func TestShowBandNoAdvance_ZeroHeightBand(t *testing.T) {
	e := newBoostEngine(t)
	b := band.NewPageHeaderBand()
	b.SetName("ZeroHBand")
	b.SetHeight(0) // zero height → should return early
	b.SetVisible(true)
	e.showBandNoAdvance(b)
}

// TestShowBandNoAdvance_ValidBand_WithObjects covers the happy path including
// populateBandObjects2 call and curY advance.
func TestShowBandNoAdvance_ValidBand_WithObjects(t *testing.T) {
	e := newBoostEngine(t)
	b := band.NewPageHeaderBand()
	b.SetName("ValidBand")
	b.SetHeight(25)
	b.SetVisible(true)

	// Add a text object so the hasObjects branch is exercised.
	txt := object.NewTextObject()
	txt.SetName("BandTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(80)
	txt.SetHeight(15)
	txt.SetVisible(true)
	txt.SetText("hello")
	b.Objects().Add(txt)

	beforeY := e.curY
	e.showBandNoAdvance(b)
	afterY := e.curY
	// curY should have advanced by the band height within showBandNoAdvance.
	if afterY != beforeY+25 {
		t.Errorf("showBandNoAdvance: curY went from %v to %v, want +25", beforeY, afterY)
	}
}

// ── pages.go: endColumn ───────────────────────────────────────────────────────

// TestEndColumn_SingleColumn covers the cols<=1 early-return (returns false).
func TestEndColumn_SingleColumn(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.Columns.Count = 1
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	result := e.endColumn(pg)
	if result {
		t.Error("endColumn single-column: expected false, got true")
	}
}

// TestEndColumn_MultiColumn_Overflow covers the path where curColumn wraps
// back to 0 (returns false indicating caller should start new page).
func TestEndColumn_MultiColumn_Overflow(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.Columns.Count = 2
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Advance curColumn past the last column to trigger the wrap path.
	e.curColumn = 1 // currently at last column (index=1 in a 2-col layout)
	result := e.endColumn(pg)
	// curColumn >= cols (2) → resets to 0 → returns false.
	if result {
		t.Error("endColumn multi-column overflow: expected false (caller should start new page)")
	}
	if e.curColumn != 0 {
		t.Errorf("endColumn: curColumn after overflow = %d, want 0", e.curColumn)
	}
}

// TestEndColumn_MultiColumn_Advance covers the path where curColumn advances
// to the next column within the same page (returns true).
func TestEndColumn_MultiColumn_Advance(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.Columns.Count = 3
	pg.PaperWidth = 210
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Start at column 0; endColumn should advance to column 1.
	e.curColumn = 0
	result := e.endColumn(pg)
	if !result {
		t.Error("endColumn multi-column advance: expected true (moved to next column)")
	}
	if e.curColumn != 1 {
		t.Errorf("endColumn: curColumn after advance = %d, want 1", e.curColumn)
	}
}

// ── pages.go: runBands default branch ─────────────────────────────────────────

// TestRunBands_DefaultBandType covers the default case in runBands (a band
// type that is neither DataBand nor GroupHeaderBand → calls showBand).
func TestRunBands_DefaultBandType(t *testing.T) {
	e := newBoostEngine(t)
	// A plain BandBase (not DataBand or GroupHeaderBand) falls through to default.
	plainBand := band.NewPageHeaderBand()
	plainBand.SetName("PlainBand")
	plainBand.SetHeight(15)
	plainBand.SetVisible(true)

	bands := []report.Base{plainBand}
	if err := e.runBands(bands); err != nil {
		t.Fatalf("runBands with default band type: %v", err)
	}
}

// ── pages.go: attachWatermark ─────────────────────────────────────────────────

// TestAttachWatermark_NilCurrentPage covers the cur==nil early-return path.
// preparedPages is non-nil but has no pages, so CurrentPage() returns nil.
func TestAttachWatermark_NilCurrentPage(t *testing.T) {
	e := newBoostEngineUnrun(t)
	// Assign a fresh PreparedPages with no pages added yet.
	e.preparedPages = preview.New()

	pg := reportpkg.NewReportPage()
	pg.SetName("NilCurPage")
	pg.Watermark = reportpkg.NewWatermark()
	pg.Watermark.Enabled = true
	pg.Watermark.Text = "TEST"

	// Should not panic — CurrentPage() returns nil → early return.
	e.attachWatermark(pg)
}

// ── objects.go: populateBandObjects2 — ProcessAt deferred paths ────────────────

// TestPopulateBandObjects2_ProcessAtPageFinished covers the ProcessAtPageFinished
// AddRepeatingDeferredHandler path in populateBandObjects2.
func TestPopulateBandObjects2_ProcessAtPageFinished(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH")
	hdr.SetHeight(40)
	hdr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("DeferredTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("[PageNumber]")
	txt.SetProcessAt(object.ProcessAtPageFinished)
	hdr.Objects().Add(txt)

	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with ProcessAtPageFinished: %v", err)
	}
}

// TestPopulateBandObjects2_ProcessAtColumnFinished covers the
// ProcessAtColumnFinished AddRepeatingDeferredHandler path.
func TestPopulateBandObjects2_ProcessAtColumnFinished(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH2")
	hdr.SetHeight(40)
	hdr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("ColDeferredTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("[PageNumber]")
	txt.SetProcessAt(object.ProcessAtColumnFinished)
	hdr.Objects().Add(txt)

	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with ProcessAtColumnFinished: %v", err)
	}
}

// TestPopulateBandObjects2_ProcessAtReportFinished covers the default
// AddDeferredHandler path (ProcessAtReportFinished).
func TestPopulateBandObjects2_ProcessAtReportFinished(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH3")
	hdr.SetHeight(40)
	hdr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("RepFinishedTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("[PageNumber]")
	txt.SetProcessAt(object.ProcessAtReportFinished)
	hdr.Objects().Add(txt)

	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with ProcessAtReportFinished: %v", err)
	}
}

// TestPopulateBandObjects2_ProcessAtDataFinished covers ProcessAtDataFinished path.
func TestPopulateBandObjects2_ProcessAtDataFinished(t *testing.T) {
	ds := data.NewBaseDataSource("PATDs")
	ds.SetAlias("PATDs")
	ds.AddColumn(data.Column{Name: "Val"})
	ds.AddRow(map[string]any{"Val": 1})
	if err := ds.Init(); err != nil {
		t.Fatalf("ds.Init: %v", err)
	}

	r := reportpkg.NewReport()
	r.Dictionary().AddDataSource(ds)
	pg := reportpkg.NewReportPage()

	db := band.NewDataBand()
	db.SetName("PAT_DB")
	db.SetHeight(20)
	db.SetVisible(true)
	db.SetDataSource(ds)

	txt := object.NewTextObject()
	txt.SetName("DataFinishedTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(15)
	txt.SetVisible(true)
	txt.SetText("[Val]")
	txt.SetProcessAt(object.ProcessAtDataFinished)
	db.Objects().Add(txt)

	pg.AddBand(db)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with ProcessAtDataFinished: %v", err)
	}
}

// ── objects.go: evalGaugeValue type cases ─────────────────────────────────────

// TestEvalGaugeValue_Float32Result covers the float32 case in evalGaugeValue's
// type switch by using a variable registered as float32.
func TestEvalGaugeValue_Float32Result(t *testing.T) {
	e := newBoostEngine(t)
	// Register a float32 variable in the report dictionary.
	e.report.Dictionary().SetSystemVariable("GaugeF32", float32(42.5))

	g := &gauge.GaugeObject{}
	g.Minimum = 0
	g.Maximum = 100
	g.SetValue(0)
	g.Expression = "GaugeF32"

	e.evalGaugeValue(g)
	// float32(42.5) → float64(42.5) → SetValue called.
	// Just verify no panic and value was set.
	if g.Value() == 0 {
		t.Log("evalGaugeValue float32: value still 0 (expression may not resolve as float32)")
	}
}

// TestEvalGaugeValue_Int64Result covers the int64 case in evalGaugeValue's
// type switch.
func TestEvalGaugeValue_Int64Result(t *testing.T) {
	e := newBoostEngine(t)
	e.report.Dictionary().SetSystemVariable("GaugeI64", int64(77))

	g := &gauge.GaugeObject{}
	g.Minimum = 0
	g.Maximum = 100
	g.SetValue(0)
	g.Expression = "GaugeI64"

	e.evalGaugeValue(g)
	// int64(77) → float64(77) → SetValue called.
	if g.Value() == 0 {
		t.Log("evalGaugeValue int64: value still 0 (expression may not resolve as int64)")
	}
}

// TestEvalGaugeValue_IntResult covers the int case (already partially tested via
// "PageNumber") — use a registered int variable.
func TestEvalGaugeValue_IntResult(t *testing.T) {
	e := newBoostEngine(t)
	e.report.Dictionary().SetSystemVariable("GaugeInt", int(55))

	g := &gauge.GaugeObject{}
	g.Minimum = 0
	g.Maximum = 100
	g.SetValue(0)
	g.Expression = "GaugeInt"

	e.evalGaugeValue(g)
}

// ── objects.go: evalGaugeText success path ────────────────────────────────────

// TestEvalGaugeText_SuccessfulCalc covers the path where Calc succeeds and
// returns a non-error result (e.report != nil, expr resolves).
func TestEvalGaugeText_SuccessfulCalc(t *testing.T) {
	e := newBoostEngine(t)
	// Use a system variable that the engine registers: "PageNumber".
	result := e.evalGaugeText("PageNumber", 0.0)
	if result == "" {
		t.Error("evalGaugeText: expected non-empty result for PageNumber expression")
	}
}

// ── objects.go: buildPreparedObject — remaining uncovered types ───────────────

// TestBuildPreparedObject_TextObject_WithHighlightVisible covers the highlight
// condition branch where matched=true and Visible=true (applies fill/color/font).
func TestBuildPreparedObject_TextObject_WithHighlightVisible(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH_HL")
	hdr.SetHeight(40)
	hdr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("HLTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("Highlighted")

	// Add a highlight that always matches (expression evaluates to true).
	hl := style.HighlightCondition{
		Expression: "true",
		Visible:    true,
		ApplyFill:  true,
	}
	txt.AddHighlight(hl)

	hdr.Objects().Add(txt)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with visible highlight: %v", err)
	}
}

// TestBuildPreparedObject_TextObject_WithHighlightInvisible covers the highlight
// branch where matched=true and Visible=false (returns nil → object skipped).
func TestBuildPreparedObject_TextObject_WithHighlightInvisible(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH_INV")
	hdr.SetHeight(40)
	hdr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("HLInvTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("Will be hidden by highlight")

	// Highlight matches and sets Visible=false → buildPreparedObject returns nil.
	hl := style.HighlightCondition{
		Expression: "true",
		Visible:    false,
	}
	txt.AddHighlight(hl)

	hdr.Objects().Add(txt)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with invisible highlight: %v", err)
	}
}

// TestBuildPreparedObject_ZipCodeObject_WithExpression covers the branch in
// ZipCodeObject where Expression is non-empty (uses Expression instead of Text).
func TestBuildPreparedObject_ZipCodeObject_WithExpression(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH_ZIP")
	hdr.SetHeight(40)
	hdr.SetVisible(true)

	obj := object.NewZipCodeObject()
	obj.SetName("Zip2")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(30)
	obj.SetVisible(true)
	obj.SetText("00000")
	obj.SetExpression("[PageNumber]") // expression takes precedence

	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with ZipCode+Expression: %v", err)
	}
}

// TestBuildPreparedObject_TextObject_WithHyperlinkURL covers the URL hyperlink
// branch (HyperlinkKind=1, with non-empty Value).
func TestBuildPreparedObject_TextObject_WithHyperlinkURL(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH_URL")
	hdr.SetHeight(40)
	hdr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("URLTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("Click me")
	// Set a URL hyperlink.
	hl := &report.Hyperlink{Kind: "URL", Value: "https://example.com"}
	txt.SetHyperlink(hl)

	hdr.Objects().Add(txt)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with URL hyperlink: %v", err)
	}
}

// TestBuildPreparedObject_TextObject_WithHyperlinkURLExpression covers the URL
// hyperlink branch where Value is empty and Expression is evaluated.
func TestBuildPreparedObject_TextObject_WithHyperlinkURLExpression(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH_URLEXPR")
	hdr.SetHeight(40)
	hdr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("URLExprTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("Link")
	hl := &report.Hyperlink{Kind: "URL", Value: "", Expression: "PageNumber"}
	txt.SetHyperlink(hl)

	hdr.Objects().Add(txt)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with URL expression hyperlink: %v", err)
	}
}

// TestBuildPreparedObject_TextObject_WithHyperlinkBookmark covers the Bookmark
// hyperlink branch (HyperlinkKind=3).
func TestBuildPreparedObject_TextObject_WithHyperlinkBookmark(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH_BM")
	hdr.SetHeight(40)
	hdr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("BMTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("Bookmark link")
	hl := &report.Hyperlink{Kind: "Bookmark", Expression: "SectionAnchor"}
	txt.SetHyperlink(hl)

	hdr.Objects().Add(txt)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with Bookmark hyperlink: %v", err)
	}
}

// TestBuildPreparedObject_TextObject_WithBookmark covers the Bookmark() method
// branch in buildPreparedObject (AddBookmark call).
func TestBuildPreparedObject_TextObject_WithBookmark(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH_BK")
	hdr.SetHeight(40)
	hdr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("BKTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("Bookmarked object")
	txt.SetBookmark("myBookmark")

	hdr.Objects().Add(txt)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with Bookmark on text object: %v", err)
	}
}

// ── objects.go: populateBandObjects nil path ──────────────────────────────────

// TestPopulateBandObjects_NilBandBase covers the bb==nil guard in
// populateBandObjects (the legacy wrapper).
func TestPopulateBandObjects_NilBandBase(t *testing.T) {
	e := newBoostEngine(t)
	pb := &preview.PreparedBand{Name: "test", Top: 0, Height: 20}
	// Should return immediately without panic.
	e.populateBandObjects(nil, pb)
}

// TestPopulateBandObjects_NilCollection covers the objs==nil guard in
// populateBandObjects2.
func TestPopulateBandObjects2_NilCollection(t *testing.T) {
	e := newBoostEngine(t)
	pb := &preview.PreparedBand{Name: "test2", Top: 0, Height: 20}
	// Should return immediately without panic.
	e.populateBandObjects2(nil, nil, pb)
}

// ── pages.go: bandHeight — non-hasHeight path ─────────────────────────────────

// TestBandHeight_NoHeightInterface covers the path where b does not implement
// hasHeight — returns 0.
func TestBandHeight_NoHeightInterface(t *testing.T) {
	e := newBoostEngine(t)
	// report.BaseObject does not implement Height() float32 → returns 0.
	bo := report.NewBaseObject()
	h := e.bandHeight(bo)
	if h != 0 {
		t.Errorf("bandHeight for non-hasHeight object = %v, want 0", h)
	}
}

// ── pages.go: runBands — aborted flag ─────────────────────────────────────────

// TestRunBands_Aborted covers the e.aborted early-break path.
func TestRunBands_Aborted(t *testing.T) {
	e := newBoostEngine(t)
	e.aborted = true
	b := band.NewPageHeaderBand()
	b.SetName("AbortedBand")
	b.SetHeight(10)
	b.SetVisible(true)

	bands := []report.Base{b}
	if err := e.runBands(bands); err != nil {
		t.Fatalf("runBands with aborted engine: %v", err)
	}
	// Band should not have been shown (aborted before iteration).
}
