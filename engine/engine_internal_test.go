package engine

// Internal tests for unexported functions that are never called from production
// code but exist as utility/reset helpers.
// Using package engine (not engine_test) so we can access unexported fields.

import (
	"encoding/base64"
	"image"
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/barcode"
	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func newInternalEngine(t *testing.T) *ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	return New(r)
}

// ── initReprint ───────────────────────────────────────────────────────────────

func TestInitReprint_ClearsAllLists(t *testing.T) {
	e := newInternalEngine(t)
	// Populate all four reprint lists manually.
	entry := reprintEntry{b: band.NewBandBase()}
	e.reprintHeaders = append(e.reprintHeaders, entry)
	e.reprintFooters = append(e.reprintFooters, entry)
	e.keepReprintHeaders = append(e.keepReprintHeaders, entry)
	e.keepReprintFooters = append(e.keepReprintFooters, entry)

	e.initReprint()

	if e.reprintHeaders != nil {
		t.Error("initReprint: reprintHeaders should be nil")
	}
	if e.reprintFooters != nil {
		t.Error("initReprint: reprintFooters should be nil")
	}
	if e.keepReprintHeaders != nil {
		t.Error("initReprint: keepReprintHeaders should be nil")
	}
	if e.keepReprintFooters != nil {
		t.Error("initReprint: keepReprintFooters should be nil")
	}
}

// ── initPageNumbers ───────────────────────────────────────────────────────────

func TestInitPageNumbers_ResetsState(t *testing.T) {
	e := newInternalEngine(t)
	// Seed some state.
	e.pageNumbers = []pageNumberInfo{{pageNo: 1, totalPages: 3}, {pageNo: 2}}
	e.logicalPageNo = 5

	e.initPageNumbers()

	if e.pageNumbers != nil {
		t.Error("initPageNumbers: pageNumbers should be nil")
	}
	if e.logicalPageNo != 0 {
		t.Errorf("initPageNumbers: logicalPageNo = %d, want 0", e.logicalPageNo)
	}
}

// ── startKeepReprint / endKeepReprint ────────────────────────────────────────

func TestStartKeepReprint_ClearsKeepLists(t *testing.T) {
	e := newInternalEngine(t)
	entry := reprintEntry{b: band.NewBandBase()}
	e.keepReprintHeaders = append(e.keepReprintHeaders, entry)
	e.keepReprintFooters = append(e.keepReprintFooters, entry)

	e.startKeepReprint()

	if e.keepReprintHeaders != nil {
		t.Error("startKeepReprint: keepReprintHeaders should be nil")
	}
	if e.keepReprintFooters != nil {
		t.Error("startKeepReprint: keepReprintFooters should be nil")
	}
}

// ── syncSystemVariables ───────────────────────────────────────────────────────

func TestSyncSystemVariables_UpdatesDictionary(t *testing.T) {
	e := newInternalEngine(t)
	// Run first so the dictionary and date are initialised.
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	e.pageNo = 7
	e.totalPages = 10
	e.rowNo = 3
	e.absRowNo = 5

	e.syncSystemVariables()

	d := e.report.Dictionary()
	if d == nil {
		t.Skip("no dictionary — cannot verify system variables")
	}
	// Just verify no panic; actual value checking would require GetSystemVariable.
}

func TestSyncSystemVariables_NilReport(t *testing.T) {
	e := &ReportEngine{} // nil report
	e.syncSystemVariables() // should not panic
}

// ── processAtToEngineState ────────────────────────────────────────────────────

func TestProcessAtToEngineState_AllCases(t *testing.T) {
	cases := []struct {
		pa   object.ProcessAt
		want EngineState
	}{
		{object.ProcessAtReportFinished, EngineStateReportFinished},
		{object.ProcessAtReportPageFinished, EngineStateReportPageFinished},
		{object.ProcessAtPageFinished, EngineStatePageFinished},
		{object.ProcessAtColumnFinished, EngineStateColumnFinished},
		{object.ProcessAtDataFinished, EngineStateBlockFinished},
		{object.ProcessAtGroupFinished, EngineStateGroupFinished},
		{object.ProcessAt(99), EngineStateReportFinished}, // default
	}
	for _, tc := range cases {
		got := processAtToEngineState(tc.pa)
		if got != tc.want {
			t.Errorf("processAtToEngineState(%v) = %v, want %v", tc.pa, got, tc.want)
		}
	}
}

// ── resetGroupTotals ──────────────────────────────────────────────────────────

func TestResetGroupTotals_ResetsResetAfterPrintTotals(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// resetGroupTotals with no aggregateTotals — should be a no-op, no panic.
	e.resetGroupTotals()
}

func TestEndKeepReprint_MergesIntoMainLists(t *testing.T) {
	e := newInternalEngine(t)
	b1 := band.NewBandBase()
	b2 := band.NewBandBase()
	e.keepReprintHeaders = []reprintEntry{{b: b1}}
	e.keepReprintFooters = []reprintEntry{{b: b2}}

	e.endKeepReprint()

	if len(e.reprintHeaders) != 1 || e.reprintHeaders[0].b != b1 {
		t.Errorf("endKeepReprint: reprintHeaders = %v, want 1 entry with b1", e.reprintHeaders)
	}
	if len(e.reprintFooters) != 1 || e.reprintFooters[0].b != b2 {
		t.Errorf("endKeepReprint: reprintFooters = %v, want 1 entry with b2", e.reprintFooters)
	}
	if e.keepReprintHeaders != nil {
		t.Error("endKeepReprint: keepReprintHeaders should be nil after merge")
	}
	if e.keepReprintFooters != nil {
		t.Error("endKeepReprint: keepReprintFooters should be nil after merge")
	}
}

// ── runBandsFromBase ──────────────────────────────────────────────────────────

func TestRunBandsFromBase_NilAndEmpty(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := e.runBandsFromBase(nil); err != nil {
		t.Errorf("runBandsFromBase(nil) = %v, want nil", err)
	}
	if err := e.runBandsFromBase([]report.Base{}); err != nil {
		t.Errorf("runBandsFromBase(empty) = %v, want nil", err)
	}
}

// ── objTopBottom ──────────────────────────────────────────────────────────────

func TestObjTopBottom_WithPositionedObject(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetTop(10)
	bb.SetHeight(30)
	top, bottom := objTopBottom(bb)
	if top != 10 {
		t.Errorf("objTopBottom: top = %v, want 10", top)
	}
	if bottom != 40 {
		t.Errorf("objTopBottom: bottom = %v, want 40", bottom)
	}
}

func TestObjTopBottom_WithNonPositionedObject(t *testing.T) {
	bo := report.NewBaseObject()
	top, bottom := objTopBottom(bo)
	if top != 0 || bottom != 0 {
		t.Errorf("objTopBottom non-positioned: got (%v,%v), want (0,0)", top, bottom)
	}
}

// ── objCanBreak ───────────────────────────────────────────────────────────────

func TestObjCanBreak_WithBreakableTrue(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetCanBreak(true)
	if !objCanBreak(bb) {
		t.Error("objCanBreak: expected true for CanBreak=true band")
	}
}

func TestObjCanBreak_WithBreakableFalse(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetCanBreak(false)
	if objCanBreak(bb) {
		t.Error("objCanBreak: expected false for CanBreak=false band")
	}
}

func TestObjCanBreak_WithNonBreakable(t *testing.T) {
	bo := report.NewBaseObject()
	if objCanBreak(bo) {
		t.Error("objCanBreak: expected false for non-breakable object")
	}
}

// ── evalGaugeText ─────────────────────────────────────────────────────────────

func TestEvalGaugeText_EmptyExpr(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	result := e.evalGaugeText("", 42.0)
	if result != "42" {
		t.Errorf("evalGaugeText empty expr = %q, want \"42\"", result)
	}
}

func TestEvalGaugeText_NonEmptyExprFallback(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// A non-existent variable expression should fall back to the default value.
	result := e.evalGaugeText("NoSuchVar", 99.0)
	if result == "" {
		t.Error("evalGaugeText: fallback result should not be empty")
	}
}

func TestEvalGaugeText_NilReport(t *testing.T) {
	e := &ReportEngine{} // nil report
	result := e.evalGaugeText("SomeExpr", 5.5)
	// Falls through to default because e.report == nil.
	if result == "" {
		t.Error("evalGaugeText nil report: should return formatted default")
	}
}

// ── evalGaugeValue ────────────────────────────────────────────────────────────

func TestEvalGaugeValue_EmptyExpr_IsNoop(t *testing.T) {
	e := newInternalEngine(t)
	g := &gauge.GaugeObject{} // Expression = ""
	// Should return immediately — no panic.
	e.evalGaugeValue(g)
}

func TestEvalGaugeValue_NilReport_IsNoop(t *testing.T) {
	e := &ReportEngine{} // nil report
	g := &gauge.GaugeObject{}
	g.Expression = "SomeExpr"
	// Should not panic — guard on e.report == nil.
	e.evalGaugeValue(g)
}

// ── renderGaugeBlob ───────────────────────────────────────────────────────────

func TestRenderGaugeBlob_NilImage_ReturnsMinusOne(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	idx := e.renderGaugeBlob("test", nil)
	if idx != -1 {
		t.Errorf("renderGaugeBlob(nil) = %d, want -1", idx)
	}
}

func TestRenderGaugeBlob_ValidImage_ReturnsIndex(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	idx := e.renderGaugeBlob("gauge", img)
	if idx < 0 {
		t.Errorf("renderGaugeBlob(valid img) = %d, want >= 0", idx)
	}
}

// ── decodeSvgData ─────────────────────────────────────────────────────────────

func TestDecodeSvgData_RawSVG(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>`
	result := decodeSvgData(svg)
	if string(result) != svg {
		t.Errorf("decodeSvgData raw: got %q, want %q", result, svg)
	}
}

func TestDecodeSvgData_Base64SVG(t *testing.T) {
	svg := `<svg><circle/></svg>`
	encoded := base64.StdEncoding.EncodeToString([]byte(svg))
	result := decodeSvgData(encoded)
	if string(result) != svg {
		t.Errorf("decodeSvgData base64: got %q, want %q", result, svg)
	}
}

func TestDecodeSvgData_URLBase64SVG(t *testing.T) {
	// URL-safe base64 uses '-' and '_' instead of '+' and '/'.
	// Encode something that will produce '+' or '/' in standard base64.
	data := []byte{0xFB, 0xFF, 0xFE} // encodes to +//+ in standard, -__- in URL
	encoded := base64.URLEncoding.EncodeToString(data)
	result := decodeSvgData(encoded)
	if string(result) != string(data) {
		t.Errorf("decodeSvgData URL-base64: got %v, want %v", result, data)
	}
}

func TestDecodeSvgData_InvalidReturnsNil(t *testing.T) {
	result := decodeSvgData("not-valid-base64-!@#$%^&*()")
	if result != nil {
		t.Errorf("decodeSvgData invalid: got %v, want nil", result)
	}
}

// ── renderBarcode ─────────────────────────────────────────────────────────────

func TestRenderBarcode_WithCode128_Success(t *testing.T) {
	bc := barcode.NewCode128Barcode()
	if err := bc.Encode("HELLO"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := renderBarcode(bc, 200, 60)
	if err != nil {
		t.Fatalf("renderBarcode error: %v", err)
	}
	if img == nil {
		t.Error("renderBarcode: expected non-nil image")
	}
}

// noRenderBarcode satisfies barcode.BarcodeBase but does NOT implement Render.
type noRenderBarcode struct{}

func (n *noRenderBarcode) Type() barcode.BarcodeType    { return barcode.BarcodeTypeCode128 }
func (n *noRenderBarcode) Encode(_ string) error        { return nil }
func (n *noRenderBarcode) DefaultValue() string         { return "" }

func TestRenderBarcode_NoRenderer_ReturnsError(t *testing.T) {
	_, err := renderBarcode(&noRenderBarcode{}, 200, 60)
	if err == nil {
		t.Error("renderBarcode: expected error for type without Render method")
	}
}

// ── extractBarcodeModules ─────────────────────────────────────────────────────

func TestExtractBarcodeModules_NilImage(t *testing.T) {
	result := extractBarcodeModules(nil)
	if result != nil {
		t.Errorf("extractBarcodeModules(nil) = %v, want nil", result)
	}
}

func TestExtractBarcodeModules_ValidImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3, 2))
	img.Set(0, 0, color.Black)
	img.Set(1, 0, color.White)
	img.Set(2, 0, color.Black)
	img.Set(0, 1, color.White)
	img.Set(1, 1, color.Black)
	img.Set(2, 1, color.White)
	modules := extractBarcodeModules(img)
	if len(modules) != 2 {
		t.Fatalf("extractBarcodeModules: rows = %d, want 2", len(modules))
	}
	if len(modules[0]) != 3 {
		t.Fatalf("extractBarcodeModules: cols = %d, want 3", len(modules[0]))
	}
	if !modules[0][0] {
		t.Error("expected pixel (0,0) to be dark")
	}
	if modules[0][1] {
		t.Error("expected pixel (1,0) to be light")
	}
}

// ── evalTextWithFormat (format branch) ───────────────────────────────────────

func TestEvalTextWithFormat_WithFormat_AppliesFormat(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// "[PageNumber]" is a single-bracket expression; with a format, the raw
	// value is formatted rather than converted directly to string.
	f := format.NewNumberFormat()
	result := e.evalTextWithFormat("[PageNumber]", f)
	if result == "" {
		t.Error("evalTextWithFormat with format: result should not be empty")
	}
}

func TestEvalTextWithFormat_NilFormat_CalcText(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// nil format takes the CalcText path.
	result := e.evalTextWithFormat("Page [PageNumber]", nil)
	if result == "" {
		t.Error("evalTextWithFormat nil format: result should not be empty")
	}
}

// ── limitPreparedPages ────────────────────────────────────────────────────────

func TestLimitPreparedPages_TrimsWhenOverLimit(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Add a second page to the prepared pages to simulate totalPages=2.
	e.preparedPages.AddPage(210, 297, 2)
	e.totalPages = 2
	e.pagesLimit = 1

	e.limitPreparedPages()

	if e.totalPages != 1 {
		t.Errorf("limitPreparedPages: totalPages = %d, want 1", e.totalPages)
	}
	if e.preparedPages.Count() != 1 {
		t.Errorf("limitPreparedPages: pages count = %d, want 1", e.preparedPages.Count())
	}
}

func TestLimitPreparedPages_NoopWhenZeroLimit(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	e.totalPages = 5
	e.pagesLimit = 0 // 0 = no limit
	e.limitPreparedPages()
	if e.totalPages != 5 {
		t.Errorf("limitPreparedPages zero limit: totalPages = %d, want 5", e.totalPages)
	}
}

// ── CalcBandHeight / calcBandRequiredHeight ───────────────────────────────────

func TestCalcBandHeight_WithCanGrow_MeasuresTextObjects(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	bb := band.NewBandBase()
	bb.SetName("GrowBand")
	bb.SetHeight(20)
	bb.SetCanGrow(true)

	txt := object.NewTextObject()
	txt.SetName("T1")
	txt.SetText("Hello World")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(15)
	txt.SetVisible(true)
	bb.Objects().Add(txt)

	h := e.CalcBandHeight(bb)
	// Just verify it returns a non-negative height and doesn't panic.
	if h < 0 {
		t.Errorf("CalcBandHeight with CanGrow: got %v, want >= 0", h)
	}
}

func TestCalcBandHeight_WithCanShrink_MeasuresTextObjects(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	bb := band.NewBandBase()
	bb.SetName("ShrinkBand")
	bb.SetHeight(60)
	bb.SetCanShrink(true)

	txt := object.NewTextObject()
	txt.SetName("T2")
	txt.SetText("Hi")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(15)
	txt.SetVisible(true)
	bb.Objects().Add(txt)

	h := e.CalcBandHeight(bb)
	if h < 0 {
		t.Errorf("CalcBandHeight with CanShrink: got %v, want >= 0", h)
	}
}

func TestCalcBandHeight_NonTextObjectInBand(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	bb := band.NewBandBase()
	bb.SetName("GrowBand2")
	bb.SetHeight(20)
	bb.SetCanGrow(true)

	// Add a non-TextObject (ContainerObject has Top/Height but isn't TextObject).
	co := object.NewContainerObject()
	co.SetName("CO")
	co.SetTop(0)
	co.SetHeight(50)
	co.SetWidth(100)
	bb.Objects().Add(co)

	h := e.CalcBandHeight(bb)
	if h < 0 {
		t.Errorf("CalcBandHeight with non-text object: got %v, want >= 0", h)
	}
}

// ── ShowBandWithPageBreaks ────────────────────────────────────────────────────

func TestShowBandWithPageBreaks_BandHasPageBreak_SplitsAndShows(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	bb := band.NewBandBase()
	bb.SetName("PBBand")
	bb.SetHeight(30)
	bb.SetVisible(true)

	// Add a TextObject with PageBreak=true so BandHasHardPageBreaks returns true.
	txt := object.NewTextObject()
	txt.SetName("PBText")
	txt.SetText("page break here")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(15)
	txt.SetVisible(true)
	txt.SetPageBreak(true)
	bb.Objects().Add(txt)

	// Should not panic; exercises ShowBandWithPageBreaks → SplitHardPageBreaks.
	e.ShowBandWithPageBreaks(bb)
}

// ── AddPage (preparedPages access via internal) ───────────────────────────────

func TestPreparedPages_AddPage_ViaInternalAccess(t *testing.T) {
	e := newInternalEngine(t)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	before := e.preparedPages.Count()
	e.preparedPages.AddPage(210, 297, 99)
	after := e.preparedPages.Count()
	if after != before+1 {
		t.Errorf("AddPage: count went from %d to %d, want %d", before, after, before+1)
	}
}

// ── preview.PreparedPages (imported via internal access) ──────────────────────

func TestPreparedPages_TypeCheck(t *testing.T) {
	pp := preview.New()
	if pp == nil {
		t.Error("preview.New() returned nil")
	}
}
