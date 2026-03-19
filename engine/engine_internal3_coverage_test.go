package engine

// engine_internal3_coverage_test.go — internal tests (package engine) targeting
// unexported functions that cannot be reached from external test packages:
//
//   keep.go startKeepBand:  `if b != nil && b.AbsRowNo() == 1 && !b.StartNewPage() { return }`

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── startKeepBand: b != nil, AbsRowNo==1, !FirstRowStartsNewPage → early return ─

// TestStartKeepBand_NonNilBand_FirstRow exercises the guard:
//   `if b != nil && b.AbsRowNo() == 1 && !b.FirstRowStartsNewPage() { return }`
// startKeepBand is only called via StartKeep (which passes nil), so this
// branch is unreachable through the public API.  We call it directly.
func TestStartKeepBand_NonNilBand_FirstRow(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	db := band.NewDataBand()
	db.SetAbsRowNo(1)                  // first absolute row
	db.SetFirstRowStartsNewPage(false) // !FirstRowStartsNewPage → guard fires

	// Keeping must be false before the call.
	if e.keeping {
		t.Fatal("expected keeping=false before test")
	}

	// Call the unexported function directly.
	// Because AbsRowNo()==1 and !FirstRowStartsNewPage(), keeping should remain false.
	e.startKeepBand(&db.BandBase)

	if e.keeping {
		t.Error("startKeepBand with AbsRowNo==1 and !FirstRowStartsNewPage should NOT set keeping=true")
	}
}

// ── AddBandToPreparedPages: nil preparedPages guard (bands.go:169) ─────────────

// TestAddBandToPreparedPages_NilPreparedPages exercises the `if e.preparedPages == nil { return false }`
// guard. Since New() always initialises preparedPages, this requires direct field access.
func TestAddBandToPreparedPages_NilPreparedPages(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	e.preparedPages = nil // force the guard condition

	db := band.NewDataBand()
	db.SetName("NilPP")
	db.SetHeight(10)
	db.SetVisible(true)

	if e.AddBandToPreparedPages(&db.BandBase) {
		t.Error("AddBandToPreparedPages with nil preparedPages should return false")
	}
}

// ── calcBandRequiredHeight: HTML render paths (bands.go:135-138) ────────────

// TestCalcBandRequiredHeight_HtmlTagsRenderType exercises the HtmlTags branch.
func TestCalcBandRequiredHeight_HtmlTagsRenderType(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	db := band.NewDataBand()
	db.SetName("HtmlTagsBand")
	db.SetHeight(50)
	db.SetVisible(true)
	db.SetCanGrow(true)
	db.SetWidth(200)

	txt := object.NewTextObject()
	txt.SetName("HtmlTxt")
	txt.SetText("<b>Hello</b> <i>World</i>")
	txt.SetWidth(180)
	txt.SetHeight(20)
	txt.SetTop(5)
	txt.SetTextRenderType(object.TextRenderTypeHtmlTags)
	db.Objects().Add(txt)

	_ = e.CalcBandHeight(&db.BandBase)
}

// TestCalcBandRequiredHeight_HtmlParagraphRenderType exercises the HtmlParagraph branch.
func TestCalcBandRequiredHeight_HtmlParagraphRenderType(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	db := band.NewDataBand()
	db.SetName("HtmlParaBand")
	db.SetHeight(50)
	db.SetVisible(true)
	db.SetCanGrow(true)
	db.SetWidth(200)

	txt := object.NewTextObject()
	txt.SetName("HtmlPara")
	txt.SetText("<p>Paragraph one</p><p>Paragraph two</p>")
	txt.SetWidth(180)
	txt.SetHeight(20)
	txt.SetTop(5)
	txt.SetTextRenderType(object.TextRenderTypeHtmlParagraph)
	db.Objects().Add(txt)

	_ = e.CalcBandHeight(&db.BandBase)
}

// TestCalcBandRequiredHeight_ZeroWidthTextObject exercises the `if objWidth <= 0 { objWidth = bb.Width() }` fallback.
func TestCalcBandRequiredHeight_ZeroWidthTextObject(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	db := band.NewDataBand()
	db.SetName("ZeroWidthBand")
	db.SetHeight(50)
	db.SetVisible(true)
	db.SetCanGrow(true)
	db.SetWidth(200)

	txt := object.NewTextObject()
	txt.SetName("ZeroWidthTxt")
	txt.SetText("some text")
	txt.SetWidth(0) // zero width → uses band width
	txt.SetHeight(20)
	txt.SetTop(5)
	db.Objects().Add(txt)

	_ = e.CalcBandHeight(&db.BandBase)
}

// TestStartKeepBand_AlreadyKeeping exercises the `if e.keeping { return }` guard.
// This is covered by TestStartKeep_Idempotent but we add an explicit internal
// test for clarity and to ensure direct startKeepBand coverage.
func TestStartKeepBand_AlreadyKeeping(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Set keeping=true directly.
	e.keeping = true
	savedY := e.keepCurY

	// Call again — should return immediately without updating keepCurY.
	e.AdvanceY(50)
	e.startKeepBand(nil)

	if e.keepCurY != savedY {
		t.Errorf("startKeepBand while keeping: keepCurY changed from %v to %v", savedY, e.keepCurY)
	}
}

