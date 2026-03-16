package engine_test

// engine_objects_coverage_test.go — coverage tests for objects.go and subreports.go
// uncovered branches: buildPreparedObject for many object types, evalGaugeValue
// with float64/string returns, RenderInnerSubreports with PrintOnParent subreport.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// newCoverageEngine builds a simple engine with a page-header band populated
// from obj (which must be a report.Base). Panics if Run fails.
func newCoveragePageHeader(t *testing.T) (*reportpkg.Report, *band.PageHeaderBand, *reportpkg.ReportPage) {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH")
	hdr.SetHeight(80)
	hdr.SetVisible(true)
	return r, hdr, pg
}

func runReport(t *testing.T, r *reportpkg.Report) *engine.ReportEngine {
	t.Helper()
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return e
}

// ── HtmlObject in band ────────────────────────────────────────────────────────

func TestBuildPreparedObject_HtmlObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewHtmlObject()
	obj.SetName("Html1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(40)
	obj.SetVisible(true)
	obj.SetText("<b>Hello</b>")
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── RichObject in band ────────────────────────────────────────────────────────

func TestBuildPreparedObject_RichObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewRichObject()
	obj.SetName("Rich1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(40)
	obj.SetVisible(true)
	obj.SetText("{\\rtf1 Hello}")
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── LineObject in band ────────────────────────────────────────────────────────

func TestBuildPreparedObject_LineObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewLineObject()
	obj.SetName("Line1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(2)
	obj.SetVisible(true)
	obj.SetDiagonal(true)
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── ShapeObject in band ───────────────────────────────────────────────────────

func TestBuildPreparedObject_ShapeObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewShapeObject()
	obj.SetName("Shape1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(80)
	obj.SetHeight(40)
	obj.SetVisible(true)
	obj.SetShape(object.ShapeKindEllipse)
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── CellularTextObject in band ───────────────────────────────────────────────

func TestBuildPreparedObject_CellularTextObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewCellularTextObject()
	obj.SetName("CT1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(40)
	obj.SetVisible(true)
	obj.SetText("ABC")
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── ZipCodeObject in band ────────────────────────────────────────────────────

func TestBuildPreparedObject_ZipCodeObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewZipCodeObject()
	obj.SetName("Zip1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(30)
	obj.SetVisible(true)
	obj.SetText("12345")
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── CheckBoxObject in band ───────────────────────────────────────────────────

func TestBuildPreparedObject_CheckBoxObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewCheckBoxObject()
	obj.SetName("CB1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(20)
	obj.SetHeight(20)
	obj.SetVisible(true)
	obj.SetChecked(true)
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── PictureObject in band ────────────────────────────────────────────────────

func TestBuildPreparedObject_PictureObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewPictureObject()
	obj.SetName("Pic1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(60)
	obj.SetVisible(true)
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── MapObject in band ────────────────────────────────────────────────────────

func TestBuildPreparedObject_MapObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewMapObject()
	obj.SetName("Map1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(200)
	obj.SetHeight(100)
	obj.SetVisible(true)
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── MSChartObject in band ────────────────────────────────────────────────────

func TestBuildPreparedObject_MSChartObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewMSChartObject()
	obj.SetName("Chart1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(200)
	obj.SetHeight(100)
	obj.SetVisible(true)
	obj.ChartType = "Bar"
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── SparklineObject in band ──────────────────────────────────────────────────

func TestBuildPreparedObject_SparklineObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewSparklineObject()
	obj.SetName("Spark1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(40)
	obj.SetVisible(true)
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── RFIDLabel in band ────────────────────────────────────────────────────────

func TestBuildPreparedObject_RFIDLabel(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewRFIDLabel()
	obj.SetName("RFID1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(40)
	obj.SetVisible(true)
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── DigitalSignatureObject in band ───────────────────────────────────────────

func TestBuildPreparedObject_DigitalSignatureObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewDigitalSignatureObject()
	obj.SetName("DS1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(40)
	obj.SetVisible(true)
	obj.SetPlaceholder("Sign here")
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── PolyLineObject in band ───────────────────────────────────────────────────

func TestBuildPreparedObject_PolyLineObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewPolyLineObject()
	obj.SetName("PL1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(50)
	obj.SetVisible(true)
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── PolygonObject in band ────────────────────────────────────────────────────

func TestBuildPreparedObject_PolygonObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	obj := object.NewPolygonObject()
	obj.SetName("Poly1")
	obj.SetLeft(0)
	obj.SetTop(0)
	obj.SetWidth(100)
	obj.SetHeight(50)
	obj.SetVisible(true)
	hdr.Objects().Add(obj)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── RadialGauge in band ──────────────────────────────────────────────────────

func TestBuildPreparedObject_RadialGauge(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	g := gauge.NewRadialGauge()
	g.SetName("RadialG")
	g.SetLeft(0)
	g.SetTop(0)
	g.SetWidth(100)
	g.SetHeight(80)
	g.SetVisible(true)
	g.GaugeObject.Minimum = 0
	g.GaugeObject.Maximum = 100
	g.GaugeObject.SetValue(75)
	hdr.Objects().Add(g)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── SimpleGauge in band ───────────────────────────────────────────────────────

func TestBuildPreparedObject_SimpleGauge(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	g := gauge.NewSimpleGauge()
	g.SetName("SimpleG")
	g.SetLeft(0)
	g.SetTop(0)
	g.SetWidth(100)
	g.SetHeight(80)
	g.SetVisible(true)
	g.GaugeObject.Minimum = 0
	g.GaugeObject.Maximum = 100
	g.GaugeObject.SetValue(50)
	hdr.Objects().Add(g)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── SimpleProgressGauge in band ───────────────────────────────────────────────

func TestBuildPreparedObject_SimpleProgressGauge(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	g := gauge.NewSimpleProgressGauge()
	g.SetName("SPG")
	g.SetLeft(0)
	g.SetTop(0)
	g.SetWidth(100)
	g.SetHeight(40)
	g.SetVisible(true)
	g.GaugeObject.Minimum = 0
	g.GaugeObject.Maximum = 100
	g.GaugeObject.SetValue(30)
	hdr.Objects().Add(g)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── evalGaugeValue: expression that fails to resolve ─────────────────────────

func TestEvalGaugeValue_ExpressionError(t *testing.T) {
	// Use an unresolvable expression to trigger the err != nil return path.
	r, hdr, pg := newCoveragePageHeader(t)
	g := gauge.NewLinearGauge()
	g.SetName("LG2")
	g.SetLeft(0)
	g.SetTop(0)
	g.SetWidth(120)
	g.SetHeight(40)
	g.SetVisible(true)
	g.GaugeObject.Minimum = 0
	g.GaugeObject.Maximum = 100
	g.GaugeObject.SetValue(0)
	g.GaugeObject.Expression = "NonExistentVariable9999"
	hdr.Objects().Add(g)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}

// ── RenderInnerSubreports: subreport with PrintOnParent=true ──────────────────

func TestRenderInnerSubreports_PrintOnParent(t *testing.T) {
	r := reportpkg.NewReport()

	// Main page.
	pg := reportpkg.NewReportPage()
	pg.SetName("Page1")

	// A subreport page with a band.
	subPg := reportpkg.NewReportPage()
	subPg.SetName("SubPage1")
	subHdr := band.NewPageHeaderBand()
	subHdr.SetName("SubPH")
	subHdr.SetHeight(30)
	subHdr.SetVisible(true)
	subTxt := object.NewTextObject()
	subTxt.SetName("SubTxt")
	subTxt.SetLeft(0)
	subTxt.SetTop(0)
	subTxt.SetWidth(100)
	subTxt.SetHeight(20)
	subTxt.SetVisible(true)
	subTxt.SetText("Sub Content")
	subHdr.Objects().Add(subTxt)
	subPg.SetPageHeader(subHdr)

	// Main page header with PrintOnParent subreport.
	mainHdr := band.NewPageHeaderBand()
	mainHdr.SetName("MainPH")
	mainHdr.SetHeight(60)
	mainHdr.SetVisible(true)

	sr := object.NewSubreportObject()
	sr.SetName("SR1")
	sr.SetLeft(0)
	sr.SetTop(0)
	sr.SetWidth(200)
	sr.SetHeight(40)
	sr.SetVisible(true)
	sr.SetReportPageName("SubPage1")
	sr.SetPrintOnParent(true)

	mainHdr.Objects().Add(sr)
	pg.SetPageHeader(mainHdr)

	r.AddPage(pg)
	r.AddPage(subPg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with PrintOnParent subreport: %v", err)
	}
	if e.PreparedPages().Count() == 0 {
		t.Error("expected at least 1 prepared page")
	}
}

// ── populateBandObjects: invisible object path ────────────────────────────────

func TestPopulateBandObjects_InvisibleObject(t *testing.T) {
	r, hdr, pg := newCoveragePageHeader(t)
	txt := object.NewTextObject()
	txt.SetName("HiddenTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(false)
	txt.SetText("Hidden")
	hdr.Objects().Add(txt)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)
	runReport(t, r)
}
