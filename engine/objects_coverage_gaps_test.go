package engine

// objects_coverage_gaps_test.go — internal tests (package engine) targeting
// remaining uncovered branches in objects.go.
//
// Uses package engine (not engine_test) so we can call unexported methods
// directly without going through the full engine run.

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	barcodepkg "github.com/andrewloable/go-fastreport/barcode"
	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/table"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func newGapsEngine(t *testing.T) *ReportEngine {
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

// ── renderGaugeBlob: nil image path ──────────────────────────────────────────

// TestRenderGaugeBlob_NilImagePath exercises the img == nil guard in renderGaugeBlob.
func TestRenderGaugeBlob_NilImagePath(t *testing.T) {
	e := newGapsEngine(t)
	idx := e.renderGaugeBlob("test", nil)
	if idx != -1 {
		t.Errorf("renderGaugeBlob(nil image) = %d, want -1", idx)
	}
}

// TestRenderGaugeBlob_ValidImagePath exercises the happy path (img != nil, pages set).
func TestRenderGaugeBlob_ValidImagePath(t *testing.T) {
	e := newGapsEngine(t)
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	idx := e.renderGaugeBlob("valid", img)
	if idx < 0 {
		t.Errorf("renderGaugeBlob(valid image) = %d, want >= 0", idx)
	}
}

// ── extractBarcodeModules: additional edge cases ──────────────────────────────

// TestExtractBarcodeModules_ZeroWidthBounds exercises the w<=0 guard.
func TestExtractBarcodeModules_ZeroWidthBounds(t *testing.T) {
	// image.Rect(0,0,0,5) produces zero-width bounds (Dx()==0).
	img := image.NewRGBA(image.Rect(0, 0, 0, 5))
	result := extractBarcodeModules(img)
	if result != nil {
		t.Errorf("extractBarcodeModules(zero-width) = %v, want nil", result)
	}
}

// TestExtractBarcodeModules_ZeroHeightBounds exercises the h<=0 guard.
func TestExtractBarcodeModules_ZeroHeightBounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 5, 0))
	result := extractBarcodeModules(img)
	if result != nil {
		t.Errorf("extractBarcodeModules(zero-height) = %v, want nil", result)
	}
}

// TestExtractBarcodeModules_LightPixels exercises lum >= 0x7FFF → false module.
func TestExtractBarcodeModules_LightPixels(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	// Set all pixels to white.
	img.Set(0, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	img.Set(1, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	img.Set(0, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	result := extractBarcodeModules(img)
	if result == nil {
		t.Fatal("extractBarcodeModules returned nil for valid image")
	}
	for y, row := range result {
		for x, dark := range row {
			if dark {
				t.Errorf("pixel (%d,%d) should be light (false), got dark (true)", x, y)
			}
		}
	}
}

// TestExtractBarcodeModules_DarkPixels exercises lum < 0x7FFF → true module.
func TestExtractBarcodeModules_DarkPixels(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	// Set all pixels to black.
	img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	img.Set(1, 1, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	result := extractBarcodeModules(img)
	if result == nil {
		t.Fatal("extractBarcodeModules returned nil for black image")
	}
	for y, row := range result {
		for x, dark := range row {
			if !dark {
				t.Errorf("pixel (%d,%d) should be dark (true), got light (false)", x, y)
			}
		}
	}
}

// ── evalTextWithFormat: format != nil, Calc fails ────────────────────────────

// TestEvalTextWithFormat_FormatSetCalcFails exercises the path where f != nil,
// text is a single bracket expression, but Calc returns an error (undefined var).
func TestEvalTextWithFormat_FormatSetCalcFails(t *testing.T) {
	e := newGapsEngine(t)
	f := format.NewNumberFormat()
	// An undefined variable causes Calc to return an error,
	// so it falls through to CalcText.
	result := e.evalTextWithFormat("[NonExistentVar99X]", f)
	// Should not panic; result may be the raw text or empty string.
	_ = result
}

// TestEvalTextWithFormat_FormatSetCalcSucceeds exercises the f != nil path
// where Calc succeeds and FormatValue is applied.
func TestEvalTextWithFormat_FormatSetCalcSucceeds(t *testing.T) {
	e := newGapsEngine(t)
	e.report.Dictionary().SetSystemVariable("NumericVal", float64(42.5))
	f := format.NewNumberFormat()
	result := e.evalTextWithFormat("[NumericVal]", f)
	if result == "" {
		t.Error("evalTextWithFormat with valid numeric var should return non-empty string")
	}
}

// TestEvalTextWithFormat_FormatSetMultiBracket exercises the branch where
// f != nil but text has multiple bracket expressions (strings.Count > 1),
// so it falls through to CalcText.
func TestEvalTextWithFormat_FormatSetMultiBracket(t *testing.T) {
	e := newGapsEngine(t)
	f := format.NewNumberFormat()
	result := e.evalTextWithFormat("[PageNumber] of [TotalPages]", f)
	_ = result
}

// TestEvalTextWithFormat_FormatSetNotBracketed exercises the branch where
// f != nil but text doesn't start with '[', so falls through to CalcText.
func TestEvalTextWithFormat_FormatSetNotBracketed(t *testing.T) {
	e := newGapsEngine(t)
	f := format.NewNumberFormat()
	result := e.evalTextWithFormat("plain text", f)
	if result != "plain text" {
		t.Errorf("evalTextWithFormat non-bracketed = %q, want %q", result, "plain text")
	}
}

// ── populateBandObjects2: ManualBuild path ────────────────────────────────────

// TestPopulateBandObjects2_ManualBuildWithCallback exercises the IsManualBuild() == true
// branch where InvokeManualBuild returns non-nil.
func TestPopulateBandObjects2_ManualBuildWithCallback(t *testing.T) {
	e := newGapsEngine(t)
	pb := &preview.PreparedBand{Name: "mb_cb_pb", Top: 0, Height: 60}

	bb := band.NewBandBase()
	bb.SetName("MB_CB_Band")
	bb.SetHeight(60)
	bb.SetVisible(true)

	tbl := table.NewTableObject()
	tbl.SetName("MB_CB_Tbl")
	tbl.SetLeft(0)
	tbl.SetTop(0)
	tbl.SetWidth(200)
	tbl.SetHeight(40)
	tbl.SetVisible(true)

	// Add a template column/row so the ManualBuild result has data.
	col := table.NewTableColumn()
	col.SetWidth(100)
	tbl.AddColumn(col)

	row := table.NewTableRow()
	row.SetHeight(20)
	cell := table.NewTableCell()
	cell.SetName("MB_C1")
	cell.SetText("MB Cell")
	row.AddCell(cell)
	tbl.AddRow(row)

	// ManualBuild callback uses PrintRow/PrintColumns to build result.
	tbl.ManualBuild = func(h *table.TableHelper) {
		h.PrintRow(0)
		h.PrintColumns()
	}

	bb.Objects().Add(tbl)
	e.populateBandObjects(bb, pb)
	// Should not panic; exercises IsManualBuild=true, InvokeManualBuild!=nil path.
}

// TestPopulateBandObjects2_ManualBuildEventOnly exercises the
// InvokeManualBuild() returning nil path (event name only, no callback).
func TestPopulateBandObjects2_ManualBuildEventOnly(t *testing.T) {
	e := newGapsEngine(t)
	pb := &preview.PreparedBand{Name: "mb_ev_pb", Top: 0, Height: 60}

	bb := band.NewBandBase()
	bb.SetName("MB_Ev_Band")
	bb.SetHeight(60)
	bb.SetVisible(true)

	tbl := table.NewTableObject()
	tbl.SetName("MB_Ev_Tbl")
	tbl.SetLeft(0)
	tbl.SetTop(0)
	tbl.SetWidth(200)
	tbl.SetHeight(40)
	tbl.SetVisible(true)
	// ManualBuildEvent set but no callback → IsManualBuild() true, InvokeManualBuild returns nil.
	tbl.ManualBuildEvent = "SomeEventName"

	col := table.NewTableColumn()
	col.SetWidth(100)
	tbl.AddColumn(col)

	bb.Objects().Add(tbl)
	e.populateBandObjects(bb, pb)
}

// ── populateContainerChildren: child returns nil from buildPreparedObject ─────

// TestPopulateContainerChildren_InvisibleChild2 exercises the path where a child
// object returns nil from buildPreparedObject (invisible child is skipped).
func TestPopulateContainerChildren_InvisibleChild2(t *testing.T) {
	e := newGapsEngine(t)
	pb := &preview.PreparedBand{Name: "cont_inv2_pb", Top: 0, Height: 40}

	cont := object.NewContainerObject()
	cont.SetName("Cont_InvChild2")
	cont.SetLeft(0)
	cont.SetTop(0)
	cont.SetWidth(100)
	cont.SetHeight(40)
	cont.SetVisible(true)

	// Add an invisible child — buildPreparedObject returns nil.
	child := object.NewTextObject()
	child.SetName("InvisChild2")
	child.SetLeft(0)
	child.SetTop(0)
	child.SetWidth(80)
	child.SetHeight(15)
	child.SetVisible(false) // invisible → buildPreparedObject returns nil
	child.SetText("Hidden")
	cont.AddChild(child)

	e.populateContainerChildren(cont, 0, 0, pb)
	if len(pb.Objects) != 0 {
		t.Errorf("expected 0 objects (invisible child skipped), got %d", len(pb.Objects))
	}
}

// ── populateTableObjects: SolidFill on cell ──────────────────────────────────

// TestPopulateTableObjects_CellWithSolidFill2 exercises the cell.Fill().(SolidFill)
// branch in populateTableObjects.
func TestPopulateTableObjects_CellWithSolidFill2(t *testing.T) {
	e := newGapsEngine(t)
	pb := &preview.PreparedBand{Name: "fill2_pb", Top: 0, Height: 40}

	tbl := &table.TableBase{}
	col := table.NewTableColumn()
	col.SetWidth(100)
	tbl.AddColumn(col)

	row := table.NewTableRow()
	row.SetHeight(20)
	cell := table.NewTableCell()
	cell.SetName("FillCell2")
	cell.SetText("Colored")
	cell.SetFill(style.NewSolidFill(color.RGBA{R: 200, G: 100, B: 50, A: 255}))
	row.AddCell(cell)
	tbl.AddRow(row)

	e.populateTableObjects(tbl, 0, 0, pb)
	if len(pb.Objects) == 0 {
		t.Error("expected at least 1 PreparedObject from cell with SolidFill")
	}
}

// TestPopulateTableObjects_EndColClamped2 exercises endCol > len(cols) clamp.
func TestPopulateTableObjects_EndColClamped2(t *testing.T) {
	e := newGapsEngine(t)
	pb := &preview.PreparedBand{Name: "clamp2_pb", Top: 0, Height: 40}

	tbl := &table.TableBase{}
	col := table.NewTableColumn()
	col.SetWidth(100)
	tbl.AddColumn(col)

	row := table.NewTableRow()
	row.SetHeight(20)
	cell := table.NewTableCell()
	cell.SetName("SpanCell2")
	cell.SetText("Span")
	cell.SetColSpan(5) // ColSpan > number of columns → triggers endCol > len(cols) clamp
	row.AddCell(cell)
	tbl.AddRow(row)

	e.populateTableObjects(tbl, 0, 0, pb)
	if len(pb.Objects) == 0 {
		t.Error("expected at least 1 PreparedObject from clamped ColSpan cell")
	}
}

// ── populateAdvMatrixCells: overflow clamps ───────────────────────────────────

// TestPopulateAdvMatrixCells_EndColOverflow exercises endCol > len(colX)-1 clamp.
func TestPopulateAdvMatrixCells_EndColOverflow(t *testing.T) {
	e := newGapsEngine(t)
	pb := &preview.PreparedBand{Name: "adv_col_ov_pb", Top: 0, Height: 40}

	adv := object.NewAdvMatrixObject()
	adv.SetName("AdvColOv")
	adv.SetLeft(0)
	adv.SetTop(0)
	adv.TableColumns = []*object.AdvMatrixColumn{
		{Name: "C1", Width: 50},
	}
	adv.TableRows = []*object.AdvMatrixRow{
		{
			Name:   "R1",
			Height: 20,
			Cells: []*object.AdvMatrixCell{
				{Name: "Cell1", Text: "A", ColSpan: 10}, // exceeds col count
			},
		},
	}
	e.populateAdvMatrixCells(adv, pb)
	if len(pb.Objects) == 0 {
		t.Error("expected at least 1 PreparedObject with clamped ColSpan")
	}
}

// TestPopulateAdvMatrixCells_EndRowOverflow exercises endRow > len(rowYOff)-1 clamp.
func TestPopulateAdvMatrixCells_EndRowOverflow(t *testing.T) {
	e := newGapsEngine(t)
	pb := &preview.PreparedBand{Name: "adv_row_ov_pb", Top: 0, Height: 40}

	adv := object.NewAdvMatrixObject()
	adv.SetName("AdvRowOv")
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
				{Name: "Cell1", Text: "B", RowSpan: 10}, // exceeds row count
			},
		},
	}
	e.populateAdvMatrixCells(adv, pb)
	if len(pb.Objects) == 0 {
		t.Error("expected at least 1 PreparedObject with clamped RowSpan")
	}
}

// TestPopulateAdvMatrixCells_RowHeightDefault exercises h<=0 → h=20 default.
func TestPopulateAdvMatrixCells_RowHeightDefault(t *testing.T) {
	e := newGapsEngine(t)
	pb := &preview.PreparedBand{Name: "adv_h0_def_pb", Top: 0, Height: 40}

	adv := object.NewAdvMatrixObject()
	adv.SetName("AdvH0Def")
	adv.SetLeft(0)
	adv.SetTop(0)
	adv.TableColumns = []*object.AdvMatrixColumn{
		{Name: "C1", Width: 100},
	}
	adv.TableRows = []*object.AdvMatrixRow{
		{
			Name:   "R1",
			Height: 0, // <= 0 → uses default 20
			Cells: []*object.AdvMatrixCell{
				{Name: "Cell1", Text: "ZeroH"},
			},
		},
	}
	e.populateAdvMatrixCells(adv, pb)
	if len(pb.Objects) == 0 {
		t.Error("expected at least 1 PreparedObject from zero-height row")
	}
}

// ── buildPreparedObject: HtmlObject with SolidFill (color.A > 0) ─────────────

// TestBuildPreparedObject_HtmlObjectSolidFill exercises the SolidFill branch
// where f.Color.A > 0 in the HtmlObject case.
func TestBuildPreparedObject_HtmlObjectSolidFill(t *testing.T) {
	e := newGapsEngine(t)
	obj := object.NewHtmlObject()
	obj.SetName("Html_SFill")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(40)
	obj.SetVisible(true)
	obj.SetText("<b>Filled</b>")
	obj.SetFill(style.NewSolidFill(color.RGBA{R: 255, G: 200, B: 100, A: 255}))

	po := e.buildPreparedObject(obj)
	if po == nil {
		t.Fatal("buildPreparedObject(HtmlObject with fill) returned nil")
	}
	if po.FillColor.A == 0 {
		t.Error("HtmlObject with SolidFill: FillColor.A should be non-zero")
	}
}

// ── buildPreparedObject: TextObject highlight branches ────────────────────────

// TestBuildPreparedObject_TextHighlightCalcError exercises the highlight loop
// where report.Calc returns an error (continue path).
func TestBuildPreparedObject_TextHighlightCalcError(t *testing.T) {
	e := newGapsEngine(t)
	txt := object.NewTextObject()
	txt.SetName("HLCalcErr")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("HLError")

	// Highlight with expression that causes a Calc error.
	hl := style.HighlightCondition{
		Expression: "%%invalid_expr%%",
		Visible:    true,
		ApplyFill:  true,
		FillColor:  color.RGBA{R: 255, A: 255},
	}
	txt.AddHighlight(hl)

	po := e.buildPreparedObject(txt)
	if po == nil {
		t.Error("buildPreparedObject should return non-nil even with highlight calc error")
	}
}

// TestBuildPreparedObject_TextHighlightApplyFont exercises the ApplyFont path.
func TestBuildPreparedObject_TextHighlightApplyFont(t *testing.T) {
	e := newGapsEngine(t)
	txt := object.NewTextObject()
	txt.SetName("HLApplyFont")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("FontHighlight")

	boldFont := style.Font{Name: "Arial", Size: 14, Style: style.FontStyleBold}
	hl := style.HighlightCondition{
		Expression: "true",
		Visible:    true,
		ApplyFont:  true,
		Font:       boldFont,
	}
	txt.AddHighlight(hl)

	po := e.buildPreparedObject(txt)
	if po == nil {
		t.Fatal("buildPreparedObject returned nil")
	}
	if po.Font.Style != style.FontStyleBold {
		t.Errorf("highlight ApplyFont: font style = %v, want %v", po.Font.Style, style.FontStyleBold)
	}
}

// TestBuildPreparedObject_TextHighlightApplyTextFill exercises ApplyTextFill path.
func TestBuildPreparedObject_TextHighlightApplyTextFill(t *testing.T) {
	e := newGapsEngine(t)
	txt := object.NewTextObject()
	txt.SetName("HLApplyTxtFill")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("TxtFillHighlight")

	red := color.RGBA{R: 255, A: 255}
	hl := style.HighlightCondition{
		Expression:    "true",
		Visible:       true,
		ApplyTextFill: true,
		TextFillColor: red,
	}
	txt.AddHighlight(hl)

	po := e.buildPreparedObject(txt)
	if po == nil {
		t.Fatal("buildPreparedObject returned nil")
	}
	if po.TextColor.R != 255 {
		t.Errorf("highlight ApplyTextFill: TextColor.R = %d, want 255", po.TextColor.R)
	}
}

// ── buildPreparedObject: PictureObject with image data ───────────────────────

// TestBuildPreparedObject_PicObjWithData exercises the len(data) > 0 branch.
func TestBuildPreparedObject_PicObjWithData(t *testing.T) {
	e := newGapsEngine(t)
	pic := object.NewPictureObject()
	pic.SetName("Pic_WithData2")
	pic.SetLeft(0)
	pic.SetTop(0)
	pic.SetWidth(100)
	pic.SetHeight(60)
	pic.SetVisible(true)
	pic.SetImageData([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})

	po := e.buildPreparedObject(pic)
	if po == nil {
		t.Fatal("buildPreparedObject(PictureObject with data) returned nil")
	}
	if po.BlobIdx < 0 {
		t.Errorf("PictureObject with image data: BlobIdx = %d, want >= 0", po.BlobIdx)
	}
}

// ── buildPreparedObject: SparklineObject with valid ChartData ────────────────

// buildGapsSparklineChartData creates a base64-encoded sparkline chart XML.
func buildGapsSparklineChartData(values []float64) string {
	pts := ""
	for _, v := range values {
		pts += fmt.Sprintf(`<DataPoint YValues="%.6g"/>`, v)
	}
	xmlStr := fmt.Sprintf(`<Chart><Series><Series Name="S1" ChartType="Line"><Points>%s</Points></Series></Series></Chart>`, pts)
	return base64.StdEncoding.EncodeToString([]byte(xmlStr))
}

// TestBuildPreparedObject_SparklineObjWithData exercises the SparklineObject
// case where ChartData is valid (DecodeChartData returns non-nil, Render returns img).
func TestBuildPreparedObject_SparklineObjWithData(t *testing.T) {
	e := newGapsEngine(t)
	obj := object.NewSparklineObject()
	obj.SetName("Spark_WithData")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(40)
	obj.SetVisible(true)
	obj.ChartData = buildGapsSparklineChartData([]float64{10, 20, 15, 25, 5})

	po := e.buildPreparedObject(obj)
	if po == nil {
		t.Fatal("buildPreparedObject(SparklineObject with data) returned nil")
	}
	if po.BlobIdx < 0 {
		t.Errorf("SparklineObject with data: BlobIdx = %d, want >= 0", po.BlobIdx)
	}
}

// ── buildPreparedObject: SVGObject with data and nil preparedPages ────────────

// TestBuildPreparedObject_SVGObjNilPreparedPages exercises the SVGObject branch
// where preparedPages is nil (BlobStore.Add is not called).
func TestBuildPreparedObject_SVGObjNilPreparedPages(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := &ReportEngine{report: r}

	svg := object.NewSVGObject()
	svg.SetName("SVG_NilPP")
	svg.SetLeft(0)
	svg.SetTop(0)
	svg.SetWidth(100)
	svg.SetHeight(50)
	svg.SetVisible(true)
	svg.SvgData = `<svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>`

	po := e.buildPreparedObject(svg)
	if po == nil {
		t.Fatal("buildPreparedObject(SVGObject) returned nil even with valid SVG")
	}
	if po.BlobIdx != -1 {
		t.Errorf("SVGObject with nil preparedPages: BlobIdx = %d, want -1", po.BlobIdx)
	}
}

// ── buildPreparedObject: unrecognised type with geometry (default case) ───────

// geomOnlyObj has geometry but is not a known report object type.
type geomOnlyObj struct {
	report.BaseObject
}

func (m *geomOnlyObj) Left() float32   { return 0 }
func (m *geomOnlyObj) Top() float32    { return 0 }
func (m *geomOnlyObj) Width() float32  { return 100 }
func (m *geomOnlyObj) Height() float32 { return 20 }

// TestBuildPreparedObject_UnknownTypeWithGeom exercises the default-case return nil
// for a type with geometry but not matching any known type.
func TestBuildPreparedObject_UnknownTypeWithGeom(t *testing.T) {
	e := newGapsEngine(t)
	obj := &geomOnlyObj{}
	po := e.buildPreparedObject(obj)
	if po != nil {
		t.Errorf("unknown geom type: expected nil, got %+v", po)
	}
}

// ── buildPreparedObject: MapObject with zero dimensions ──────────────────────

// TestBuildPreparedObject_MapObjZeroDim exercises the Width/Height <= 0 → use
// 400/200 defaults path in MapObject case.
func TestBuildPreparedObject_MapObjZeroDim(t *testing.T) {
	e := newGapsEngine(t)
	obj := object.NewMapObject()
	obj.SetName("MapZeroDim2")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(0)  // triggers opts.Width = 400 default
	obj.SetHeight(0) // triggers opts.Height = 200 default
	obj.SetVisible(true)

	po := e.buildPreparedObject(obj)
	if po == nil {
		t.Fatal("buildPreparedObject(MapObject zero dim) returned nil")
	}
}

// ── buildPreparedObject: BarcodeObject zero dimensions ───────────────────────

// TestBuildPreparedObject_BarcodeObjZeroDim exercises the w <= 0 → w=200 and
// h <= 0 → h=60 default dimension branches in the barcode case.
func TestBuildPreparedObject_BarcodeObjZeroDim(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH_BC0D")
	hdr.SetHeight(80)
	hdr.SetVisible(true)

	bc := barcodepkg.NewBarcodeObject()
	bc.SetName("BC_ZD")
	bc.SetLeft(0)
	bc.SetTop(0)
	bc.SetWidth(0)  // triggers w = 200 default
	bc.SetHeight(0) // triggers h = 60 default
	bc.SetVisible(true)
	bc.SetText("CODE128TEST")
	bc.Barcode = barcodepkg.NewCode128Barcode()

	hdr.Objects().Add(bc)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with zero-dim BarcodeObject: %v", err)
	}
}

// ── buildPreparedObject: BarcodeObject HideIfNoData=true, text empty ──────────

// TestBuildPreparedObject_BarcodeObjHideIfNoData exercises the HideIfNoData=true
// branch where text is empty so the barcode is not rendered.
func TestBuildPreparedObject_BarcodeObjHideIfNoData(t *testing.T) {
	e := newGapsEngine(t)
	bc := barcodepkg.NewBarcodeObject()
	bc.SetName("BC_Hide2")
	bc.SetLeft(0)
	bc.SetTop(0)
	bc.SetWidth(200)
	bc.SetHeight(60)
	bc.SetVisible(true)
	bc.SetText("") // empty text
	bc.SetHideIfNoData(true)
	bc.SetNoDataText("NODATA")
	bc.Barcode = barcodepkg.NewCode128Barcode()

	po := e.buildPreparedObject(bc)
	if po == nil {
		t.Fatal("buildPreparedObject(BarcodeObject HideIfNoData) returned nil")
	}
	// BlobIdx should be -1 since text is empty and HideIfNoData is true.
	if po.BlobIdx != -1 {
		t.Errorf("BarcodeObject HideIfNoData: BlobIdx = %d, want -1", po.BlobIdx)
	}
}

// TestBuildPreparedObject_BarcodeObjNoDataText exercises the branch where
// text is empty and HideIfNoData is false — NoDataText is used as fallback.
func TestBuildPreparedObject_BarcodeObjNoDataText(t *testing.T) {
	e := newGapsEngine(t)
	bc := barcodepkg.NewBarcodeObject()
	bc.SetName("BC_NoData2")
	bc.SetLeft(0)
	bc.SetTop(0)
	bc.SetWidth(200)
	bc.SetHeight(60)
	bc.SetVisible(true)
	bc.SetText("")          // empty text
	bc.SetHideIfNoData(false)
	bc.SetNoDataText("CODE128") // fallback
	bc.Barcode = barcodepkg.NewCode128Barcode()

	po := e.buildPreparedObject(bc)
	if po == nil {
		t.Fatal("buildPreparedObject(BarcodeObject NoDataText) returned nil")
	}
}

// ── buildPreparedObject: BarcodeObject with Expression ───────────────────────

// TestBuildPreparedObject_BarcodeObjWithExpression exercises the branch where
// Expression() != "" is used instead of Text().
func TestBuildPreparedObject_BarcodeObjWithExpression(t *testing.T) {
	e := newGapsEngine(t)
	e.report.Dictionary().SetSystemVariable("BarcodeText", "HELLO123")

	bc := barcodepkg.NewBarcodeObject()
	bc.SetName("BC_Expr")
	bc.SetLeft(0)
	bc.SetTop(0)
	bc.SetWidth(200)
	bc.SetHeight(60)
	bc.SetVisible(true)
	bc.SetExpression("[BarcodeText]")
	bc.SetText("FALLBACK")
	bc.Barcode = barcodepkg.NewCode128Barcode()

	po := e.buildPreparedObject(bc)
	if po == nil {
		t.Fatal("buildPreparedObject(BarcodeObject with expression) returned nil")
	}
}

// ── populateTableObjects: rowSpan < 1 clamp + multi-row RowSpan ──────────────

// TestPopulateTableObjects_RowSpanZeroClamp exercises the rowSpan < 1 → rowSpan=1 clamp.
func TestPopulateTableObjects_RowSpanZeroClamp(t *testing.T) {
	e := newGapsEngine(t)
	pb := &preview.PreparedBand{Name: "rspan0_pb", Top: 0, Height: 40}

	tbl := &table.TableBase{}
	col := table.NewTableColumn()
	col.SetWidth(100)
	tbl.AddColumn(col)

	row := table.NewTableRow()
	row.SetHeight(20)
	cell := table.NewTableCell()
	cell.SetName("RS0Cell")
	cell.SetText("RSZero")
	cell.SetRowSpan(0) // < 1 → clamped to 1
	row.AddCell(cell)
	tbl.AddRow(row)

	e.populateTableObjects(tbl, 0, 0, pb)
	if len(pb.Objects) == 0 {
		t.Error("expected at least 1 PreparedObject from rowSpan=0 cell")
	}
}

// TestPopulateTableObjects_MultiRowRowSpan exercises the inner rowSpan loop that
// iterates over multiple rows (RowSpan > 1 with multiple rows available).
func TestPopulateTableObjects_MultiRowRowSpan(t *testing.T) {
	e := newGapsEngine(t)
	pb := &preview.PreparedBand{Name: "mrowspan_pb", Top: 0, Height: 60}

	tbl := &table.TableBase{}
	col := table.NewTableColumn()
	col.SetWidth(100)
	tbl.AddColumn(col)

	// First row: cell spans 2 rows (so inner loop iterates twice).
	row1 := table.NewTableRow()
	row1.SetHeight(20)
	cell1 := table.NewTableCell()
	cell1.SetName("Span2Cell")
	cell1.SetText("Span2")
	cell1.SetRowSpan(2) // spans row 0 and row 1
	row1.AddCell(cell1)
	tbl.AddRow(row1)

	// Second row: needed so RowSpan can span into it.
	row2 := table.NewTableRow()
	row2.SetHeight(20)
	cell2 := table.NewTableCell()
	cell2.SetName("R2Cell")
	cell2.SetText("Row2")
	row2.AddCell(cell2)
	tbl.AddRow(row2)

	e.populateTableObjects(tbl, 0, 0, pb)
	// Should produce 2 PreparedObjects (one for each cell).
	if len(pb.Objects) == 0 {
		t.Error("expected PreparedObjects from multi-row RowSpan table")
	}
}

// ── populateAdvMatrixCells: empty TableRows ───────────────────────────────────

// TestPopulateAdvMatrixCells_EmptyRows exercises the len(adv.TableRows)==0 → return path.
func TestPopulateAdvMatrixCells_EmptyRows(t *testing.T) {
	e := newGapsEngine(t)
	pb := &preview.PreparedBand{Name: "adv_empty_pb", Top: 0, Height: 40}

	adv := object.NewAdvMatrixObject()
	adv.SetName("AdvEmpty")
	adv.SetLeft(0)
	adv.SetTop(0)
	// TableRows is empty (nil) → should return immediately.
	adv.TableColumns = []*object.AdvMatrixColumn{{Name: "C1", Width: 100}}
	adv.TableRows = nil

	e.populateAdvMatrixCells(adv, pb)
	if len(pb.Objects) != 0 {
		t.Errorf("expected 0 objects from empty AdvMatrix, got %d", len(pb.Objects))
	}
}

// ── buildPreparedObject: PolyLineObject with points ───────────────────────────

// TestBuildPreparedObject_PolyLineObjWithPoints exercises the inner loop in
// the PolyLineObject case where pts.Len() > 0.
func TestBuildPreparedObject_PolyLineObjWithPoints(t *testing.T) {
	e := newGapsEngine(t)
	obj := object.NewPolyLineObject()
	obj.SetName("PL_Points")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(50)
	obj.SetVisible(true)
	// Add points so the loop body executes.
	obj.Points().Add(&object.PolyPoint{X: 10, Y: 5})
	obj.Points().Add(&object.PolyPoint{X: 50, Y: 25})
	obj.Points().Add(&object.PolyPoint{X: 90, Y: 45})

	po := e.buildPreparedObject(obj)
	if po == nil {
		t.Fatal("buildPreparedObject(PolyLineObject with points) returned nil")
	}
	if len(po.Points) != 3 {
		t.Errorf("PolyLineObject points: got %d, want 3", len(po.Points))
	}
}

// TestBuildPreparedObject_PolygonObjWithPoints exercises the inner loop in
// the PolygonObject case where pts.Len() > 0.
func TestBuildPreparedObject_PolygonObjWithPoints(t *testing.T) {
	e := newGapsEngine(t)
	obj := object.NewPolygonObject()
	obj.SetName("Poly_Points")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(50)
	obj.SetVisible(true)
	obj.Points().Add(&object.PolyPoint{X: 10, Y: 0})
	obj.Points().Add(&object.PolyPoint{X: 90, Y: 0})
	obj.Points().Add(&object.PolyPoint{X: 50, Y: 50})

	po := e.buildPreparedObject(obj)
	if po == nil {
		t.Fatal("buildPreparedObject(PolygonObject with points) returned nil")
	}
	if len(po.Points) != 3 {
		t.Errorf("PolygonObject points: got %d, want 3", len(po.Points))
	}
}

// ── buildPreparedObject: MapObject with Layers ───────────────────────────────

// TestBuildPreparedObject_MapObjWithLayers exercises the `for _, layer := range v.Layers`
// loop in the MapObject case.
func TestBuildPreparedObject_MapObjWithLayers(t *testing.T) {
	e := newGapsEngine(t)
	obj := object.NewMapObject()
	obj.SetName("MapWithLayers")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(200)
	obj.SetHeight(100)
	obj.SetVisible(true)
	// Add a layer to trigger the loop body.
	layer := object.NewMapLayer()
	layer.Shapefile = "world"
	layer.Palette = "Blue"
	layer.Type = "Choropleth"
	obj.Layers = append(obj.Layers, layer)

	po := e.buildPreparedObject(obj)
	if po == nil {
		t.Fatal("buildPreparedObject(MapObject with layers) returned nil")
	}
}

// ── buildPreparedObject: MSChartObject with valid data ───────────────────────

// TestBuildPreparedObject_MSChartObjWithData exercises the `img != nil` branch
// in the MSChartObject case.
func TestBuildPreparedObject_MSChartObjWithData(t *testing.T) {
	e := newGapsEngine(t)
	obj := object.NewMSChartObject()
	obj.SetName("Chart_Data")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(200)
	obj.SetHeight(100)
	obj.SetVisible(true)
	obj.ChartType = "Bar"

	// Add a series with static values so RenderToImage returns non-nil.
	s := object.NewMSChartSeries()
	s.SetName("S1")
	s.ChartType = "Bar"
	obj.Series = append(obj.Series, s)

	// Try to set ChartData with a valid bar chart so rendering succeeds.
	// Use the same format as MSChart chart data XML.
	chartXML := `<Chart><Series><Series Name="S1" ChartType="Bar"><Points><DataPoint YValues="10"/><DataPoint YValues="20"/></Points></Series></Series></Chart>`
	obj.ChartData = base64.StdEncoding.EncodeToString([]byte(chartXML))

	po := e.buildPreparedObject(obj)
	if po == nil {
		t.Fatal("buildPreparedObject(MSChartObject with data) returned nil")
	}
}

// ── evalTextWithFormat: empty text path ──────────────────────────────────────

// TestEvalTextWithFormat_EmptyText exercises the `text == ""` early return path.
func TestEvalTextWithFormat_EmptyText(t *testing.T) {
	e := newGapsEngine(t)
	result := e.evalTextWithFormat("", nil)
	if result != "" {
		t.Errorf("evalTextWithFormat empty text: got %q, want empty string", result)
	}
}

// TestEvalTextWithFormat_NilReport exercises the `e.report == nil` early return path.
func TestEvalTextWithFormat_NilReport(t *testing.T) {
	// Create an engine with nil report.
	e := &ReportEngine{}
	result := e.evalTextWithFormat("some text", nil)
	if result != "some text" {
		t.Errorf("evalTextWithFormat nil report: got %q, want %q", result, "some text")
	}
}
