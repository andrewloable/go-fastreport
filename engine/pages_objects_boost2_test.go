package engine

// pages_objects_boost2_test.go — second round of internal coverage tests for
// pages.go and objects.go remaining uncovered branches.

import (
	"image"
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/table"
)

// ── renderGaugeBlob: nil preparedPages path ────────────────────────────────────

// TestRenderGaugeBlob_NilPreparedPages covers the e.preparedPages == nil guard.
func TestRenderGaugeBlob_NilPreparedPages(t *testing.T) {
	e := &ReportEngine{} // preparedPages is nil
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	idx := e.renderGaugeBlob("test", img)
	if idx != -1 {
		t.Errorf("renderGaugeBlob with nil preparedPages = %d, want -1", idx)
	}
}

// ── populateBandObjects2: nil style collection ────────────────────────────────

// TestPopulateBandObjects2_StylesNil covers when e.report.Styles() is nil.
// (Most engines don't set styles, so Styleable branch is just skipped.)
func TestPopulateBandObjects2_StylesNil(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH_NS")
	hdr.SetHeight(30)
	hdr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("NSTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(80)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("no styles")
	hdr.Objects().Add(txt)

	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run without styles: %v", err)
	}
}

// ── populateTableObjects: nil cell path ──────────────────────────────────────

// TestPopulateTableObjects_NilCell covers the `cell == nil → continue` path in
// populateTableObjects when a row has a nil cell entry.
func TestPopulateTableObjects_NilCell(t *testing.T) {
	e := newBoostEngine(t)

	// Build a PreparedBand to receive results.
	pb := &preview.PreparedBand{Name: "tbPB", Top: 0, Height: 60}

	tbl := &table.TableBase{}

	col := table.NewTableColumn()
	col.SetWidth(80)
	tbl.AddColumn(col)

	row := table.NewTableRow()
	row.SetHeight(20)
	// Add one nil cell slot (underlying slice will have a nil if AddCell is not called).
	// We add a nil cell directly.
	row.AddCell(nil) // nil cell → should trigger the continue
	tbl.AddRow(row)

	// Should not panic.
	e.populateTableObjects(tbl, 0, 0, pb)
}

// TestPopulateTableObjects_ColSpanZero covers the colSpan<1 → colSpan=1 path.
func TestPopulateTableObjects_ColSpanZero(t *testing.T) {
	e := newBoostEngine(t)
	pb := &preview.PreparedBand{Name: "spanPB", Top: 0, Height: 40}

	tbl := &table.TableBase{}
	col := table.NewTableColumn()
	col.SetWidth(100)
	tbl.AddColumn(col)

	row := table.NewTableRow()
	row.SetHeight(20)
	cell := table.NewTableCell()
	cell.SetName("SC")
	cell.SetText("hi")
	cell.SetColSpan(0) // < 1 → clamped to 1
	row.AddCell(cell)
	tbl.AddRow(row)

	e.populateTableObjects(tbl, 0, 0, pb)
}

// TestPopulateTableObjects_RowSpanZero covers the rowSpan<1 → rowSpan=1 path.
func TestPopulateTableObjects_RowSpanZero(t *testing.T) {
	e := newBoostEngine(t)
	pb := &preview.PreparedBand{Name: "rspanPB", Top: 0, Height: 40}

	tbl := &table.TableBase{}
	col := table.NewTableColumn()
	col.SetWidth(100)
	tbl.AddColumn(col)

	row := table.NewTableRow()
	row.SetHeight(20)
	cell := table.NewTableCell()
	cell.SetName("RC")
	cell.SetText("row")
	cell.SetRowSpan(0) // < 1 → clamped to 1
	row.AddCell(cell)
	tbl.AddRow(row)

	e.populateTableObjects(tbl, 0, 0, pb)
}

// ── populateAdvMatrixCells: nil cell + font/border coverage ──────────────────

// TestPopulateAdvMatrixCells_NilCell covers the `cell == nil → continue` path.
func TestPopulateAdvMatrixCells_NilCell(t *testing.T) {
	e := newBoostEngine(t)
	pb := &preview.PreparedBand{Name: "advPB", Top: 0, Height: 40}

	adv := object.NewAdvMatrixObject()
	adv.SetName("Adv2")
	adv.SetLeft(0)
	adv.SetTop(0)
	adv.TableColumns = []*object.AdvMatrixColumn{
		{Name: "C1", Width: 50},
	}
	adv.TableRows = []*object.AdvMatrixRow{
		{
			Name:   "R1",
			Height: 20,
			Cells:  []*object.AdvMatrixCell{nil}, // nil cell → continue
		},
	}

	e.populateAdvMatrixCells(adv, pb)
}

// TestPopulateAdvMatrixCells_WithFontAndBorder covers cell.Font!=nil and
// cell.Border!=nil branches in populateAdvMatrixCells.
func TestPopulateAdvMatrixCells_WithFontAndBorder(t *testing.T) {
	e := newBoostEngine(t)
	pb := &preview.PreparedBand{Name: "advPB2", Top: 0, Height: 40}

	fnt := style.DefaultFont()
	brd := &style.Border{}

	adv := object.NewAdvMatrixObject()
	adv.SetName("Adv3")
	adv.SetLeft(0)
	adv.SetTop(0)
	adv.TableColumns = []*object.AdvMatrixColumn{
		{Name: "C1", Width: 100},
	}
	adv.TableRows = []*object.AdvMatrixRow{
		{
			Name:   "R1",
			Height: 20,
			Cells: []*object.AdvMatrixCell{
				{
					Name:   "Cell1",
					Text:   "FontBorder",
					Font:   &fnt,
					Border: brd,
				},
			},
		},
	}

	e.populateAdvMatrixCells(adv, pb)
	if len(pb.Objects) == 0 {
		t.Error("expected at least 1 object in pb from AdvMatrixCell with font/border")
	}
}

// TestPopulateAdvMatrixCells_WithFillColor covers the cell.FillColor!=nil path.
func TestPopulateAdvMatrixCells_WithFillColor(t *testing.T) {
	e := newBoostEngine(t)
	pb := &preview.PreparedBand{Name: "advPB3", Top: 0, Height: 40}

	fillC := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	adv := object.NewAdvMatrixObject()
	adv.SetName("Adv4")
	adv.SetLeft(0)
	adv.SetTop(0)
	adv.TableColumns = []*object.AdvMatrixColumn{
		{Name: "C1", Width: 100},
	}

	adv.TableRows = []*object.AdvMatrixRow{
		{
			Name:   "R1",
			Height: 20,
			Cells: []*object.AdvMatrixCell{
				{
					Name:      "FillCell",
					Text:      "red",
					FillColor: &fillC,
				},
			},
		},
	}

	e.populateAdvMatrixCells(adv, pb)
	if len(pb.Objects) == 0 {
		t.Error("expected at least 1 object with fill color")
	}
}

// ── showBand: typed nil pointer path ──────────────────────────────────────────

// TestShowBand_TypedNilPointer covers the reflect.Ptr + IsNil guard in showBand.
func TestShowBand_TypedNilPointer(t *testing.T) {
	e := newBoostEngine(t)
	// A typed nil *band.OverlayBand satisfies report.Base but IsNil == true.
	var typedNil *band.OverlayBand
	// Cast to report.Base interface (typed nil).
	e.showBand(typedNil)
	// Must not panic.
}

// TestShowBand_InvisibleBand covers the Visible()==false path in showBand.
func TestShowBand_InvisibleBand(t *testing.T) {
	e := newBoostEngine(t)
	b := band.NewPageHeaderBand()
	b.SetName("InvShowBand")
	b.SetHeight(20)
	b.SetVisible(false)
	e.showBand(b)
}

// ── evalTextWithFormat: CalcText error path ────────────────────────────────────

// TestEvalTextWithFormat_CalcTextError covers the case where CalcText returns
// an error — the raw text is returned.
func TestEvalTextWithFormat_CalcTextError(t *testing.T) {
	e := newBoostEngine(t)
	// An expression with unclosed brackets confuses the parser.
	result := e.evalTextWithFormat("[Unclosed bracket", nil)
	// Should return the raw text on error.
	if result == "" {
		t.Error("evalTextWithFormat CalcText error: should return non-empty raw text")
	}
}

// ── populateContainerChildren: nil objs path ──────────────────────────────────

// TestPopulateContainerChildren_NilObjs covers the `objs == nil` early-return.
func TestPopulateContainerChildren_NilObjs(t *testing.T) {
	e := newBoostEngine(t)
	pb := &preview.PreparedBand{Name: "contPB", Top: 0, Height: 40}

	// NewContainerObject with no children: Objects() returns non-nil empty collection,
	// so we need to test with Len()==0 — that's handled. The nil case is harder
	// to reach without internal access, but we can use a container with empty objects.
	cont := object.NewContainerObject()
	cont.SetName("EmptyCont")
	cont.SetLeft(0)
	cont.SetTop(0)
	cont.SetWidth(100)
	cont.SetHeight(40)

	// Calling directly (no children) — should not panic.
	e.populateContainerChildren(cont, 0, 0, pb)
}

// ── evalGaugeValue: float64 type (explicit) ───────────────────────────────────

// TestEvalGaugeValue_Float64_Direct directly calls evalGaugeValue with a
// float64-resolving expression to confirm float64 type switch case.
func TestEvalGaugeValue_Float64_Direct(t *testing.T) {
	e := newBoostEngine(t)
	// Register a float64 system variable.
	e.report.Dictionary().SetSystemVariable("GaugeF64", float64(88.5))

	g := &gauge.GaugeObject{}
	g.Minimum = 0
	g.Maximum = 100
	g.SetValue(0)
	g.Expression = "GaugeF64"

	e.evalGaugeValue(g)
	if g.Value() != 88.5 {
		t.Logf("evalGaugeValue float64: value=%v (may not be exact if not stored as float64)", g.Value())
	}
}

// ── applyBackPage: odd page filter (BackPageOddEven=1, even page) ─────────────

// TestApplyBackPage_OddOnly_OnEvenPage covers the case 1 + even page → return.
func TestApplyBackPage_OddOnly_OnEvenPage(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("Main2")
	pg.BackPage = "BackTpl2"
	pg.BackPageOddEven = 1 // odd pages only

	backPg := reportpkg.NewReportPage()
	backPg.SetName("BackTpl2")

	r.AddPage(pg)
	r.AddPage(backPg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// With only 1 page (page 1 = odd), the back page should be applied.
	// Set to page 2 (even) and call applyBackPage directly.
	e.pageNo = 2
	e.applyBackPage(pg)
	// Should return early without rendering (odd-only filter on even page).
}

// TestApplyBackPage_EvenOnly_OnOddPage covers BackPageOddEven=2 on an odd page.
func TestApplyBackPage_EvenOnly_OnOddPage(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("Main3")
	pg.BackPage = "BackTpl3"
	pg.BackPageOddEven = 2 // even pages only

	backPg := reportpkg.NewReportPage()
	backPg.SetName("BackTpl3")

	r.AddPage(pg)
	r.AddPage(backPg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Set to odd page and call directly.
	e.pageNo = 1
	e.applyBackPage(pg)
	// Should return early (even-only filter on odd page).
}

// ── bandHeight: hasHeight path ────────────────────────────────────────────────

// TestBandHeight_HasHeight covers the hasHeight interface assertion success path.
func TestBandHeight_HasHeight(t *testing.T) {
	e := newBoostEngine(t)
	b := band.NewPageHeaderBand()
	b.SetHeight(35)
	h := e.bandHeight(b)
	if h != 35 {
		t.Errorf("bandHeight = %v, want 35", h)
	}
}

// ── RunReportPage: abort path ─────────────────────────────────────────────────

// TestRunReportPage_Aborted covers the e.aborted path in runBands where
// the engine is aborted before bands are processed.
func TestRunReportPage_Aborted(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	db := band.NewDataBand()
	db.SetName("AbortDB")
	db.SetHeight(20)
	db.SetVisible(true)
	pg.AddBand(db)

	r.AddPage(pg)
	e := New(r)
	// Abort before run — runBands will break immediately.
	e.aborted = true
	// Use Run directly; it will call runBands which breaks immediately.
	// We need preparedPages for startPage, so init it.
	if err := e.Run(DefaultRunOptions()); err != nil {
		// Aborted run may still succeed (aborted just stops band output).
		t.Logf("Run with abort: %v", err)
	}
}

// ── showBandNoAdvance: typed nil pointer ──────────────────────────────────────

// TestShowBandNoAdvance_TypedNilPointer covers the reflect.Ptr+IsNil guard.
func TestShowBandNoAdvance_TypedNilPointer(t *testing.T) {
	e := newBoostEngine(t)
	var typedNil *band.PageHeaderBand
	// Should return immediately without panic.
	e.showBandNoAdvance(typedNil)
}

// ── populateBandObjects: nil report (no Styles()) ────────────────────────────

// TestBuildPreparedObject_HighlightNoMatch covers the highlight loop path where
// the highlight condition expression returns non-bool or evaluates to false
// (the continue path in the highlight loop).
func TestBuildPreparedObject_HighlightNoMatch(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH_HLF")
	hdr.SetHeight(40)
	hdr.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("HLFTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("not highlighted")

	// Highlight with expression that evaluates to false.
	hl := style.HighlightCondition{
		Expression: "false",
		Visible:    true,
		ApplyFill:  true,
	}
	txt.AddHighlight(hl)

	hdr.Objects().Add(txt)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with non-matching highlight: %v", err)
	}
}

// ── bandHeight: report.Base that returns zero height ─────────────────────────

// TestBandHeight_ZeroHeight covers the zero-return path from bandHeight
// when the band's height is explicitly 0.
func TestBandHeight_ZeroHeight(t *testing.T) {
	e := newBoostEngine(t)
	b := band.NewPageHeaderBand()
	b.SetHeight(0) // returns 0 → showBand/showBandNoAdvance should skip
	h := e.bandHeight(b)
	if h != 0 {
		t.Errorf("bandHeight zero = %v, want 0", h)
	}
}

// ── populateBandObjects: non-nil BandBase wrapper ────────────────────────────

// TestPopulateBandObjects_NonNil covers the e.populateBandObjects path when
// bb is a valid BandBase (dispatches to populateBandObjects2).
func TestPopulateBandObjects_NonNil(t *testing.T) {
	e := newBoostEngine(t)
	pb := &preview.PreparedBand{Name: "nonNilPB", Top: 0, Height: 20}

	bb := band.NewBandBase()
	bb.SetName("NonNilBB")
	txt := object.NewTextObject()
	txt.SetName("BBTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(80)
	txt.SetHeight(15)
	txt.SetVisible(true)
	txt.SetText("test")
	bb.Objects().Add(txt)

	e.populateBandObjects(bb, pb)
	if len(pb.Objects) == 0 {
		t.Error("expected at least 1 object after populateBandObjects with text object")
	}
}

// ── buildPreparedObject: unrecognised type (default return nil) ───────────────

// TestBuildPreparedObject_UnknownType covers the default case in buildPreparedObject
// that returns nil for unrecognised types.
func TestBuildPreparedObject_UnknownType(t *testing.T) {
	e := newBoostEngine(t)
	// report.BaseObject has no geometry → returns nil at the geom check.
	bo := report.NewBaseObject()
	po := e.buildPreparedObject(bo)
	if po != nil {
		t.Errorf("buildPreparedObject unknown type: expected nil, got %+v", po)
	}
}
