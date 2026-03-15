package reportpkg_test

// reportpkg_extra_test.go — additional coverage for reportpkg package.
// Covers: PrepareWithContext, SetDictionary, SetStyles, FindPage,
// LoadWithPassword, LoadFromStringWithPassword, JSON/XML connection
// deserialization, styles_serial.go Serialize/Deserialize/formatBorderLinesLocal,
// watermark.Serialize, parseTotalType (all cases), CalcText (error branch),
// mergeFromBase, serializeBands, page.AllBands, page.Serialize/Deserialize,
// and report.Serialize (full content).

import (
	"bytes"
	"compress/gzip"
	"context"
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	_ "github.com/andrewloable/go-fastreport/engine" // registers prepare func
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/style"
)

// ── PrepareWithContext ─────────────────────────────────────────────────────

func TestPrepareWithContext_BasicRun(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("Page1")
	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(20)
	db.SetVisible(true)
	pg.AddBand(db)
	r.AddPage(pg)

	ctx := context.Background()
	if err := r.PrepareWithContext(ctx); err != nil {
		t.Fatalf("PrepareWithContext: %v", err)
	}
	pp := r.PreparedPages()
	if pp == nil {
		t.Fatal("PreparedPages is nil after PrepareWithContext")
	}
	if pp.Count() == 0 {
		t.Error("expected at least 1 prepared page")
	}
}

func TestPrepareWithContext_EmptyReport(t *testing.T) {
	r := reportpkg.NewReport()
	ctx := context.Background()
	if err := r.PrepareWithContext(ctx); err != nil {
		t.Fatalf("PrepareWithContext on empty report: %v", err)
	}
}

func TestPrepareWithContext_CancelledContext(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("Page1")
	r.AddPage(pg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	// A cancelled context should either succeed quickly (no engine work) or return
	// context.Canceled. Both are acceptable; we just confirm it doesn't panic.
	_ = r.PrepareWithContext(ctx)
}

// ── SetDictionary ──────────────────────────────────────────────────────────

func TestReport_SetDictionary(t *testing.T) {
	r := reportpkg.NewReport()
	d := data.NewDictionary()
	p := &data.Parameter{Name: "X", Value: 42}
	d.AddParameter(p)
	r.SetDictionary(d)

	if r.Dictionary() != d {
		t.Error("Dictionary() should return the newly set dictionary")
	}
	if len(r.Dictionary().Parameters()) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(r.Dictionary().Parameters()))
	}
}

func TestReport_SetDictionary_Nil(t *testing.T) {
	r := reportpkg.NewReport()
	r.SetDictionary(nil)
	if r.Dictionary() != nil {
		t.Error("SetDictionary(nil) should result in Dictionary() returning nil")
	}
}

// ── SetStyles ─────────────────────────────────────────────────────────────

func TestReport_SetStyles(t *testing.T) {
	r := reportpkg.NewReport()
	ss := style.NewStyleSheet()
	e := &style.StyleEntry{Name: "Heading"}
	ss.Add(e)
	r.SetStyles(ss)

	if r.Styles() != ss {
		t.Error("Styles() should return the newly set stylesheet")
	}
	if r.Styles().Len() != 1 {
		t.Errorf("expected 1 style, got %d", r.Styles().Len())
	}
}

// ── FindPage ──────────────────────────────────────────────────────────────

func TestReport_FindPage_Found(t *testing.T) {
	r := reportpkg.NewReport()
	p1 := reportpkg.NewReportPage()
	p1.SetName("Cover")
	p2 := reportpkg.NewReportPage()
	p2.SetName("Content")
	r.AddPage(p1)
	r.AddPage(p2)

	found := r.FindPage("Content")
	if found == nil {
		t.Fatal("FindPage(Content) returned nil")
	}
	if found != p2 {
		t.Error("FindPage returned wrong page")
	}
}

func TestReport_FindPage_NotFound(t *testing.T) {
	r := reportpkg.NewReport()
	p := reportpkg.NewReportPage()
	p.SetName("Cover")
	r.AddPage(p)

	found := r.FindPage("Missing")
	if found != nil {
		t.Error("FindPage should return nil for non-existent page")
	}
}

func TestReport_FindPage_EmptyReport(t *testing.T) {
	r := reportpkg.NewReport()
	found := r.FindPage("Anything")
	if found != nil {
		t.Error("FindPage on empty report should return nil")
	}
}

// ── LoadWithPassword ───────────────────────────────────────────────────────

func TestReport_LoadWithPassword_NonExistentFile(t *testing.T) {
	r := reportpkg.NewReport()
	err := r.LoadWithPassword("/nonexistent/path/report.frx", "secret")
	if err == nil {
		t.Error("LoadWithPassword on non-existent file should return an error")
	}
}

func TestReport_LoadWithPassword_PlainXML(t *testing.T) {
	// An unencrypted FRX file should work fine when a password is provided.
	frx := `<?xml version="1.0" encoding="utf-8"?><Report ReportName="PwdTest"><ReportPage Name="Page1"></ReportPage></Report>`
	r := reportpkg.NewReport()
	if err := r.LoadFromStringWithPassword(frx, "ignored"); err != nil {
		t.Fatalf("LoadFromStringWithPassword on plain XML: %v", err)
	}
	if r.Info.Name != "PwdTest" {
		t.Errorf("ReportName: got %q, want PwdTest", r.Info.Name)
	}
	if r.PageCount() != 1 {
		t.Errorf("PageCount: got %d, want 1", r.PageCount())
	}
}

func TestReport_LoadFromStringWithPassword_EmptyPassword(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report ReportName="PWTest2"><ReportPage Name="P1"></ReportPage></Report>`
	r := reportpkg.NewReport()
	if err := r.LoadFromStringWithPassword(frx, ""); err != nil {
		t.Fatalf("LoadFromStringWithPassword with empty password: %v", err)
	}
	if r.PageCount() != 1 {
		t.Errorf("PageCount: got %d, want 1", r.PageCount())
	}
}

// ── page.AllBands ─────────────────────────────────────────────────────────

func TestReportPage_AllBands_Empty(t *testing.T) {
	p := reportpkg.NewReportPage()
	bands := p.AllBands()
	if len(bands) != 0 {
		t.Errorf("AllBands on empty page = %d, want 0", len(bands))
	}
}

func TestReportPage_AllBands_AllSlots(t *testing.T) {
	p := reportpkg.NewReportPage()

	rt := band.NewReportTitleBand()
	rt.SetName("ReportTitle")
	ph := band.NewPageHeaderBand()
	ph.SetName("PageHeader")
	ch := band.NewColumnHeaderBand()
	ch.SetName("ColumnHeader")
	db := band.NewDataBand()
	db.SetName("Data")
	rs := band.NewReportSummaryBand()
	rs.SetName("ReportSummary")
	cf := band.NewColumnFooterBand()
	cf.SetName("ColumnFooter")
	pf := band.NewPageFooterBand()
	pf.SetName("PageFooter")
	ov := band.NewOverlayBand()
	ov.SetName("Overlay")

	p.SetReportTitle(rt)
	p.SetPageHeader(ph)
	p.SetColumnHeader(ch)
	p.AddBand(db)
	p.SetReportSummary(rs)
	p.SetColumnFooter(cf)
	p.SetPageFooter(pf)
	p.SetOverlay(ov)

	all := p.AllBands()
	// Expected order: ReportTitle, PageHeader, ColumnHeader, Data, ColumnFooter, ReportSummary, PageFooter, Overlay
	if len(all) != 8 {
		t.Fatalf("AllBands len = %d, want 8", len(all))
	}
	names := make([]string, len(all))
	for i, b := range all {
		names[i] = b.Name()
	}
	// Verify expected order
	expected := []string{"ReportTitle", "PageHeader", "ColumnHeader", "Data", "ColumnFooter", "ReportSummary", "PageFooter", "Overlay"}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("AllBands[%d] = %q, want %q", i, names[i], name)
		}
	}
}

func TestReportPage_AllBands_OnlySomeSlots(t *testing.T) {
	p := reportpkg.NewReportPage()
	ph := band.NewPageHeaderBand()
	ph.SetName("PH")
	pf := band.NewPageFooterBand()
	pf.SetName("PF")
	p.SetPageHeader(ph)
	p.SetPageFooter(pf)

	all := p.AllBands()
	if len(all) != 2 {
		t.Errorf("AllBands len = %d, want 2", len(all))
	}
}

func TestReportPage_AllBands_WithDataBands(t *testing.T) {
	p := reportpkg.NewReportPage()
	db1 := band.NewDataBand()
	db1.SetName("DB1")
	db2 := band.NewDataBand()
	db2.SetName("DB2")
	p.AddBand(db1)
	p.AddBand(db2)

	all := p.AllBands()
	if len(all) != 2 {
		t.Fatalf("AllBands len = %d, want 2", len(all))
	}
	if all[0].Name() != "DB1" || all[1].Name() != "DB2" {
		t.Error("AllBands order wrong for data bands")
	}
}

// ── mergeFromBase ─────────────────────────────────────────────────────────

func TestReportPage_MergeFromBase_InheritsSingletonBands(t *testing.T) {
	// Test mergeFromBase via Report.ApplyBase which calls mergeFromBase.
	baseReport := reportpkg.NewReport()
	basePage := reportpkg.NewReportPage()
	basePage.SetName("SharedPage")

	ph := band.NewPageHeaderBand()
	ph.SetName("BasePageHeader")
	pf := band.NewPageFooterBand()
	pf.SetName("BasePageFooter")
	rt := band.NewReportTitleBand()
	rt.SetName("BaseReportTitle")
	rs := band.NewReportSummaryBand()
	rs.SetName("BaseReportSummary")
	ch := band.NewColumnHeaderBand()
	ch.SetName("BaseColumnHeader")
	cf := band.NewColumnFooterBand()
	cf.SetName("BaseColumnFooter")
	ov := band.NewOverlayBand()
	ov.SetName("BaseOverlay")

	basePage.SetPageHeader(ph)
	basePage.SetPageFooter(pf)
	basePage.SetReportTitle(rt)
	basePage.SetReportSummary(rs)
	basePage.SetColumnHeader(ch)
	basePage.SetColumnFooter(cf)
	basePage.SetOverlay(ov)
	baseReport.AddPage(basePage)

	childReport := reportpkg.NewReport()
	childPage := reportpkg.NewReportPage()
	childPage.SetName("SharedPage") // same name → will be merged
	childReport.AddPage(childPage)
	childReport.ApplyBase(baseReport)

	merged := childReport.Page(0)
	if merged.PageHeader() == nil {
		t.Error("mergeFromBase should inherit PageHeader")
	}
	if merged.PageFooter() == nil {
		t.Error("mergeFromBase should inherit PageFooter")
	}
	if merged.ReportTitle() == nil {
		t.Error("mergeFromBase should inherit ReportTitle")
	}
	if merged.ReportSummary() == nil {
		t.Error("mergeFromBase should inherit ReportSummary")
	}
	if merged.ColumnHeader() == nil {
		t.Error("mergeFromBase should inherit ColumnHeader")
	}
	if merged.ColumnFooter() == nil {
		t.Error("mergeFromBase should inherit ColumnFooter")
	}
	if merged.Overlay() == nil {
		t.Error("mergeFromBase should inherit Overlay")
	}
}

func TestReportPage_MergeFromBase_DataBands_Prepended(t *testing.T) {
	baseReport := reportpkg.NewReport()
	basePage := reportpkg.NewReportPage()
	basePage.SetName("SharedPage")
	baseDB := band.NewDataBand()
	baseDB.SetName("BaseDataBand")
	basePage.AddBand(baseDB)
	baseReport.AddPage(basePage)

	childReport := reportpkg.NewReport()
	childPage := reportpkg.NewReportPage()
	childPage.SetName("SharedPage")
	childDB := band.NewDataBand()
	childDB.SetName("ChildDataBand")
	childPage.AddBand(childDB)
	childReport.AddPage(childPage)

	childReport.ApplyBase(baseReport)

	merged := childReport.Page(0)
	if len(merged.Bands()) != 2 {
		t.Fatalf("expected 2 bands after merge, got %d", len(merged.Bands()))
	}
	// Base bands are prepended.
	if merged.Bands()[0].Name() != "BaseDataBand" {
		t.Errorf("expected BaseDataBand first, got %q", merged.Bands()[0].Name())
	}
	if merged.Bands()[1].Name() != "ChildDataBand" {
		t.Errorf("expected ChildDataBand second, got %q", merged.Bands()[1].Name())
	}
}

func TestReportPage_MergeFromBase_DoesNotDuplicateExisting(t *testing.T) {
	baseReport := reportpkg.NewReport()
	basePage := reportpkg.NewReportPage()
	basePage.SetName("SharedPage")
	ph := band.NewPageHeaderBand()
	ph.SetName("PH")
	basePage.SetPageHeader(ph)
	baseDB := band.NewDataBand()
	baseDB.SetName("DataBand1")
	basePage.AddBand(baseDB)
	baseReport.AddPage(basePage)

	childReport := reportpkg.NewReport()
	childPage := reportpkg.NewReportPage()
	childPage.SetName("SharedPage")
	// child already has same-named band
	childDB := band.NewDataBand()
	childDB.SetName("DataBand1")
	childPage.AddBand(childDB)
	// child already has page header
	childPH := band.NewPageHeaderBand()
	childPH.SetName("PH")
	childPage.SetPageHeader(childPH)
	childReport.AddPage(childPage)

	childReport.ApplyBase(baseReport)

	merged := childReport.Page(0)
	// DataBand1 should NOT be duplicated.
	if len(merged.Bands()) != 1 {
		t.Errorf("expected 1 band (no duplication), got %d", len(merged.Bands()))
	}
	// PageHeader should NOT be overwritten (child wins).
	if merged.PageHeader() != childPH {
		t.Error("child PageHeader should not be replaced by base PageHeader")
	}
}

// ── serializeBands / page.Serialize ──────────────────────────────────────

func TestReportPage_Serialize_TitleBeforeHeader(t *testing.T) {
	// When TitleBeforeHeader = true, ReportTitle comes before PageHeader in XML.
	frx := buildPageFRX(t, func(pg *reportpkg.ReportPage) {
		pg.TitleBeforeHeader = true
		rt := band.NewReportTitleBand()
		rt.SetName("RT")
		pg.SetReportTitle(rt)
		ph := band.NewPageHeaderBand()
		ph.SetName("PH")
		pg.SetPageHeader(ph)
	})

	// Verify TitleBeforeHeader attribute present, then confirm RT before PH.
	if !strings.Contains(frx, `TitleBeforeHeader="true"`) {
		t.Error("expected TitleBeforeHeader=true in XML")
	}
	rtPos := strings.Index(frx, "<ReportTitle")
	phPos := strings.Index(frx, "<PageHeader")
	if rtPos == -1 || phPos == -1 {
		t.Fatal("expected both ReportTitle and PageHeader in XML")
	}
	if rtPos > phPos {
		t.Error("expected ReportTitle before PageHeader when TitleBeforeHeader=true")
	}
}

func TestReportPage_Serialize_DefaultHeaderOrder(t *testing.T) {
	// When TitleBeforeHeader = false (default), PageHeader comes before ReportTitle.
	frx := buildPageFRX(t, func(pg *reportpkg.ReportPage) {
		rt := band.NewReportTitleBand()
		rt.SetName("RT")
		pg.SetReportTitle(rt)
		ph := band.NewPageHeaderBand()
		ph.SetName("PH")
		pg.SetPageHeader(ph)
	})

	rtPos := strings.Index(frx, "<ReportTitle")
	phPos := strings.Index(frx, "<PageHeader")
	if rtPos == -1 || phPos == -1 {
		t.Fatal("expected both ReportTitle and PageHeader in XML")
	}
	if phPos > rtPos {
		t.Error("expected PageHeader before ReportTitle when TitleBeforeHeader=false")
	}
}

func TestReportPage_Serialize_AllBandSlots(t *testing.T) {
	frx := buildPageFRX(t, func(pg *reportpkg.ReportPage) {
		rt := band.NewReportTitleBand()
		rt.SetName("RT")
		pg.SetReportTitle(rt)
		ph := band.NewPageHeaderBand()
		ph.SetName("PH")
		pg.SetPageHeader(ph)
		ch := band.NewColumnHeaderBand()
		ch.SetName("CH")
		pg.SetColumnHeader(ch)
		db := band.NewDataBand()
		db.SetName("DB")
		pg.AddBand(db)
		rs := band.NewReportSummaryBand()
		rs.SetName("RS")
		pg.SetReportSummary(rs)
		cf := band.NewColumnFooterBand()
		cf.SetName("CF")
		pg.SetColumnFooter(cf)
		pf := band.NewPageFooterBand()
		pf.SetName("PF")
		pg.SetPageFooter(pf)
		ov := band.NewOverlayBand()
		ov.SetName("OV")
		pg.SetOverlay(ov)
	})

	for _, tag := range []string{"<ReportTitle", "<PageHeader", "<ColumnHeader", "<Data", "<ReportSummary", "<ColumnFooter", "<PageFooter", "<Overlay"} {
		if !strings.Contains(frx, tag) {
			t.Errorf("expected %q in serialized XML", tag)
		}
	}
}

func TestReportPage_Serialize_AllFlags(t *testing.T) {
	frx := buildPageFRX(t, func(pg *reportpkg.ReportPage) {
		pg.PaperWidth = 420
		pg.PaperHeight = 594
		pg.Landscape = true
		pg.LeftMargin = 15
		pg.TopMargin = 20
		pg.RightMargin = 15
		pg.BottomMargin = 20
		pg.MirrorMargins = true
		pg.TitleBeforeHeader = true
		pg.PrintOnPreviousPage = true
		pg.ResetPageNumber = true
		pg.StartOnOddPage = true
		pg.OutlineExpression = "[PageNum]"
		pg.CreatePageEvent = "OnCreate"
		pg.StartPageEvent = "OnStart"
		pg.FinishPageEvent = "OnFinish"
		pg.ManualBuildEvent = "OnBuild"
		pg.BackPage = "BackPage1"
		pg.BackPageOddEven = 1
	})

	for _, want := range []string{
		`PaperWidth="420"`,
		`PaperHeight="594"`,
		`Landscape="true"`,
		`LeftMargin="15"`,
		`TopMargin="20"`,
		`RightMargin="15"`,
		`BottomMargin="20"`,
		`MirrorMargins="true"`,
		`TitleBeforeHeader="true"`,
		`PrintOnPreviousPage="true"`,
		`ResetPageNumber="true"`,
		`StartOnOddPage="true"`,
		`OutlineExpression="[PageNum]"`,
		`CreatePageEvent="OnCreate"`,
		`StartPageEvent="OnStart"`,
		`FinishPageEvent="OnFinish"`,
		`ManualBuildEvent="OnBuild"`,
		`BackPage="BackPage1"`,
		`BackPageOddEven="1"`,
	} {
		if !strings.Contains(frx, want) {
			t.Errorf("missing %q in serialized page", want)
		}
	}
}

func TestReportPage_Deserialize_AllFields(t *testing.T) {
	// Build a page with all fields set, save it, then reload it.
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("TestPage")
	pg.PaperWidth = 420
	pg.PaperHeight = 594
	pg.Landscape = true
	pg.LeftMargin = 15
	pg.TopMargin = 20
	pg.RightMargin = 15
	pg.BottomMargin = 20
	pg.MirrorMargins = true
	pg.PrintOnPreviousPage = true
	pg.ResetPageNumber = true
	pg.StartOnOddPage = true
	pg.OutlineExpression = "[PageNo]"
	pg.CreatePageEvent = "Create"
	pg.StartPageEvent = "Start"
	pg.FinishPageEvent = "Finish"
	pg.ManualBuildEvent = "Manual"
	pg.BackPage = "BP"
	pg.BackPageOddEven = 2
	pg.Columns.Count = 3
	pg.Columns.Width = 55.5
	r.AddPage(pg)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	r2 := reportpkg.NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r2.PageCount() == 0 {
		t.Fatal("no pages after round-trip")
	}
	pg2 := r2.Page(0)
	if pg2.PaperWidth != 420 {
		t.Errorf("PaperWidth: got %v", pg2.PaperWidth)
	}
	if pg2.PaperHeight != 594 {
		t.Errorf("PaperHeight: got %v", pg2.PaperHeight)
	}
	if !pg2.Landscape {
		t.Error("Landscape should be true")
	}
	if pg2.LeftMargin != 15 {
		t.Errorf("LeftMargin: got %v", pg2.LeftMargin)
	}
	if !pg2.MirrorMargins {
		t.Error("MirrorMargins should be true")
	}
	if !pg2.PrintOnPreviousPage {
		t.Error("PrintOnPreviousPage should be true")
	}
	if !pg2.ResetPageNumber {
		t.Error("ResetPageNumber should be true")
	}
	if !pg2.StartOnOddPage {
		t.Error("StartOnOddPage should be true")
	}
	if pg2.OutlineExpression != "[PageNo]" {
		t.Errorf("OutlineExpression: got %q", pg2.OutlineExpression)
	}
	if pg2.BackPage != "BP" {
		t.Errorf("BackPage: got %q", pg2.BackPage)
	}
	if pg2.BackPageOddEven != 2 {
		t.Errorf("BackPageOddEven: got %d", pg2.BackPageOddEven)
	}
}

// buildPageFRX is a helper that creates a report with one page, applies fn,
// saves it to a string, and returns the XML.
func buildPageFRX(t *testing.T, fn func(*reportpkg.ReportPage)) string {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("Page1")
	fn(pg)
	r.AddPage(pg)
	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	return xml
}

// ── report.Serialize (full coverage) ──────────────────────────────────────

func TestReport_Serialize_AllInfoFields(t *testing.T) {
	r := reportpkg.NewReport()
	r.Info.Name = "MyReport"
	r.Info.Author = "Alice"
	r.Info.Description = "Test report"
	r.Info.Version = "2.0"
	r.Info.Created = "2024-01-01"
	r.Info.Modified = "2024-06-15"
	r.Info.CreatorVersion = "2023.1"
	r.Info.SavePreviewPicture = true
	r.Compressed = true
	r.ConvertNulls = true
	r.DoublePass = true
	r.InitialPageNumber = 3
	r.MaxPages = 10
	r.StartReportEvent = "OnStart"
	r.FinishReportEvent = "OnFinish"

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	// We save with Compressed=true so the output is gzip'd. Re-set Compressed
	// to false for plain XML comparison, or reload.
	r.Compressed = false
	xml, err = r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString (uncompressed): %v", err)
	}

	for _, want := range []string{
		`ReportName="MyReport"`,
		`ReportAuthor="Alice"`,
		`ReportDescription="Test report"`,
		`ReportVersion="2.0"`,
		`Created="2024-01-01"`,
		`Modified="2024-06-15"`,
		`CreatorVersion="2023.1"`,
		`SavePreviewPicture="true"`,
		`ConvertNulls="true"`,
		`DoublePass="true"`,
		`InitialPageNumber="3"`,
		`MaxPages="10"`,
		`StartReportEvent="OnStart"`,
		`FinishReportEvent="OnFinish"`,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("missing %q in serialized report", want)
		}
	}
}

func TestReport_Serialize_WithStyles(t *testing.T) {
	r := reportpkg.NewReport()
	ss := r.Styles()
	e := &style.StyleEntry{
		Name:          "Header",
		ApplyBorder:   true,
		ApplyFill:     true,
		ApplyTextFill: true,
		ApplyFont:     true,
	}
	ss.Add(e)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, "<Styles") {
		t.Error("expected <Styles> element in XML")
	}
	if !strings.Contains(xml, `Name="Header"`) {
		t.Error("expected style Name=Header in XML")
	}
}

func TestReport_Serialize_Compressed_RoundTrip(t *testing.T) {
	r := reportpkg.NewReport()
	r.Info.Name = "Compressed"
	r.Compressed = true
	pg := reportpkg.NewReportPage()
	pg.SetName("Page1")
	r.AddPage(pg)

	var buf bytes.Buffer
	if err := r.SaveTo(&buf); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	// Verify gzip magic bytes.
	b := buf.Bytes()
	if len(b) < 2 || b[0] != 0x1f || b[1] != 0x8b {
		t.Error("expected gzip-compressed output (magic 0x1f 0x8b)")
	}

	// Reload.
	r2 := reportpkg.NewReport()
	if err := r2.LoadFrom(bytes.NewReader(b)); err != nil {
		t.Fatalf("LoadFrom (gzip): %v", err)
	}
	if r2.Info.Name != "Compressed" {
		t.Errorf("ReportName after gzip round-trip: %q", r2.Info.Name)
	}
	if r2.PageCount() != 1 {
		t.Errorf("PageCount: %d", r2.PageCount())
	}
}

func TestReport_Serialize_WithPages(t *testing.T) {
	r := reportpkg.NewReport()
	pg1 := reportpkg.NewReportPage()
	pg1.SetName("Cover")
	pg2 := reportpkg.NewReportPage()
	pg2.SetName("Content")
	r.AddPage(pg1)
	r.AddPage(pg2)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	// Both page names must appear.
	if !strings.Contains(xml, `Name="Cover"`) {
		t.Error("missing Cover page in XML")
	}
	if !strings.Contains(xml, `Name="Content"`) {
		t.Error("missing Content page in XML")
	}
}

// ── Watermark.Serialize ────────────────────────────────────────────────────

func TestWatermark_Serialize_Enabled(t *testing.T) {
	// Watermark is on ReportPage; test via SaveToString round-trip.
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("WMPage")
	wm := pg.Watermark
	wm.Enabled = true
	wm.Text = "DRAFT"
	wm.ShowTextOnTop = false
	wm.ShowImageOnTop = true
	wm.ImageTransparency = 0.5
	// Use non-default rotation and image size.
	wm.TextRotation = reportpkg.WatermarkTextRotationHorizontal
	wm.ImageSize = reportpkg.WatermarkImageSizeStretch
	r.AddPage(pg)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	for _, want := range []string{
		`Watermark.Enabled="true"`,
		`Watermark.Text="DRAFT"`,
		`Watermark.ShowTextOnTop="false"`,
		`Watermark.ShowImageOnTop="true"`,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("missing %q in serialized watermark XML", want)
		}
	}
}

func TestWatermark_Serialize_Disabled(t *testing.T) {
	// When Enabled=false, nothing should be written about watermark.
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("NWMPage")
	// Watermark.Enabled defaults to false.
	r.AddPage(pg)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	if strings.Contains(xml, "Watermark.Enabled") {
		t.Error("Watermark.Enabled should not appear when disabled")
	}
}

func TestWatermark_Serialize_NonDefaultFont(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("FontWM")
	wm := pg.Watermark
	wm.Enabled = true
	wm.Font = style.Font{Name: "Times New Roman", Size: 48}
	r.AddPage(pg)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, "Watermark.Font") {
		t.Error("expected Watermark.Font in XML when font differs from default")
	}
}

func TestWatermark_RoundTrip(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.SetName("WMRound")
	wm := pg.Watermark
	wm.Enabled = true
	wm.Text = "CONFIDENTIAL"
	wm.TextRotation = reportpkg.WatermarkTextRotationVertical
	wm.ImageSize = reportpkg.WatermarkImageSizeCenter
	wm.ImageTransparency = 0.3
	wm.ShowImageOnTop = true
	wm.ShowTextOnTop = false
	r.AddPage(pg)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	r2 := reportpkg.NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r2.PageCount() == 0 {
		t.Fatal("no pages after round-trip")
	}
	wm2 := r2.Page(0).Watermark
	if !wm2.Enabled {
		t.Error("Watermark.Enabled should be true after round-trip")
	}
	if wm2.Text != "CONFIDENTIAL" {
		t.Errorf("Watermark.Text: got %q, want CONFIDENTIAL", wm2.Text)
	}
	if wm2.TextRotation != reportpkg.WatermarkTextRotationVertical {
		t.Errorf("Watermark.TextRotation: got %d", wm2.TextRotation)
	}
	if !wm2.ShowImageOnTop {
		t.Error("Watermark.ShowImageOnTop should be true after round-trip")
	}
	if wm2.ShowTextOnTop {
		t.Error("Watermark.ShowTextOnTop should be false after round-trip")
	}
}

// ── styles_serial.go: Serialize and Deserialize ────────────────────────────

func TestStylesSerializer_Serialize_RoundTrip(t *testing.T) {
	// Build a report with a stylesheet, save, reload, and verify.
	r := reportpkg.NewReport()
	ss := r.Styles()

	e1 := &style.StyleEntry{
		Name:        "Bold",
		ApplyFont:   true,
		FontChanged: true,
		Font:        style.Font{Name: "Arial", Size: 12, Style: style.FontStyleBold},
	}
	e2 := &style.StyleEntry{
		Name:        "NoBorder",
		ApplyBorder: false,
	}
	ss.Add(e1)
	ss.Add(e2)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, "<Styles") {
		t.Error("expected <Styles> element")
	}
	if !strings.Contains(xml, `Name="Bold"`) {
		t.Error("expected Bold style")
	}
	if !strings.Contains(xml, `Name="NoBorder"`) {
		t.Error("expected NoBorder style")
	}

	// Reload and verify.
	r2 := reportpkg.NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r2.Styles().Len() != 2 {
		t.Errorf("style count after round-trip: got %d, want 2", r2.Styles().Len())
	}
	bold := r2.Styles().Find("Bold")
	if bold == nil {
		t.Fatal("Bold style not found after round-trip")
	}
	noBorder := r2.Styles().Find("NoBorder")
	if noBorder == nil {
		t.Fatal("NoBorder style not found after round-trip")
	}
	if noBorder.ApplyBorder {
		t.Error("NoBorder ApplyBorder should be false after round-trip")
	}
}

func TestStylesSerializer_Serialize_ApplyFlags(t *testing.T) {
	r := reportpkg.NewReport()
	ss := r.Styles()

	e := &style.StyleEntry{
		Name:          "NoFlags",
		ApplyBorder:   false,
		ApplyFill:     false,
		ApplyTextFill: false,
		ApplyFont:     false,
	}
	ss.Add(e)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	// Reloaded entry should have all Apply flags false.
	r2 := reportpkg.NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	e2 := r2.Styles().Find("NoFlags")
	if e2 == nil {
		t.Fatal("NoFlags style not found")
	}
	if e2.ApplyBorder || e2.ApplyFill || e2.ApplyTextFill || e2.ApplyFont {
		t.Error("all Apply flags should be false after round-trip")
	}
}

func TestStylesSerializer_Serialize_BorderLines(t *testing.T) {
	r := reportpkg.NewReport()
	ss := r.Styles()

	e := &style.StyleEntry{
		Name: "Bordered",
	}
	e.Border = *style.NewBorder()
	e.Border.VisibleLines = style.BorderLinesLeft | style.BorderLinesRight
	ss.Add(e)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, "Border.Lines") {
		t.Error("expected Border.Lines in serialized style")
	}
	// Verify "Left" and "Right" appear.
	if !strings.Contains(xml, "Left") || !strings.Contains(xml, "Right") {
		t.Error("expected Left and Right in Border.Lines value")
	}
}

func TestStylesSerializer_Serialize_BorderShadow(t *testing.T) {
	r := reportpkg.NewReport()
	ss := r.Styles()

	e := &style.StyleEntry{Name: "Shadow"}
	e.Border = *style.NewBorder()
	e.Border.Shadow = true
	ss.Add(e)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, `Border.Shadow="true"`) {
		t.Error("expected Border.Shadow=true in XML")
	}
}

func TestStylesSerializer_Deserialize_IsNoOp(t *testing.T) {
	// The stylesSerializer.Deserialize is documented as a no-op; confirm it doesn't error.
	// We test this indirectly: loading a report with a Styles child works fine.
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Styles>
			<Style Name="Test" ApplyBorder="false"/>
		</Styles>
		<ReportPage Name="Page1"></ReportPage>
	</Report>`
	r := reportpkg.NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	// stylesSerializer.Deserialize is a no-op, but deserializeStyles in loadsave.go handles the real work.
	e := r.Styles().Find("Test")
	if e == nil {
		t.Fatal("Test style not found")
	}
	if e.ApplyBorder {
		t.Error("ApplyBorder should be false")
	}
}

// ── formatBorderLinesLocal (tested indirectly via Serialize) ──────────────

func TestFormatBorderLinesLocal_All(t *testing.T) {
	// BorderLinesAll should serialize as "All"
	r := reportpkg.NewReport()
	ss := r.Styles()
	e := &style.StyleEntry{Name: "AllBorders"}
	e.Border = *style.NewBorder()
	e.Border.VisibleLines = style.BorderLinesAll
	ss.Add(e)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, `Border.Lines="All"`) {
		t.Errorf("expected Border.Lines=All, got XML: %q", extractBorderLines(xml))
	}
}

func TestFormatBorderLinesLocal_None(t *testing.T) {
	// BorderLinesNone means nothing should be written about Border.Lines.
	r := reportpkg.NewReport()
	ss := r.Styles()
	e := &style.StyleEntry{Name: "NoBorders"}
	e.Border = *style.NewBorder()
	e.Border.VisibleLines = style.BorderLinesNone
	ss.Add(e)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	// VisibleLines=None → Border.Lines should NOT appear.
	if strings.Contains(xml, "Border.Lines") {
		t.Error("Border.Lines should not be written when None")
	}
}

func TestFormatBorderLinesLocal_TopBottom(t *testing.T) {
	r := reportpkg.NewReport()
	ss := r.Styles()
	e := &style.StyleEntry{Name: "TopBottom"}
	e.Border = *style.NewBorder()
	e.Border.VisibleLines = style.BorderLinesTop | style.BorderLinesBottom
	ss.Add(e)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, "Top") || !strings.Contains(xml, "Bottom") {
		t.Error("expected Top and Bottom in Border.Lines value")
	}
}

// extractBorderLines is a simple helper to find the Border.Lines value in XML.
func extractBorderLines(xml string) string {
	const key = `Border.Lines="`
	idx := strings.Index(xml, key)
	if idx == -1 {
		return "(not found)"
	}
	start := idx + len(key)
	end := strings.Index(xml[start:], `"`)
	if end == -1 {
		return "(malformed)"
	}
	return xml[start : start+end]
}

// ── parseTotalType — all cases ─────────────────────────────────────────────

func TestParseTotalType_AllCases(t *testing.T) {
	// parseTotalType is exercised via LoadFromString with a Dictionary/Total element.
	cases := []struct {
		totalTypeStr string
		wantType     data.TotalType
	}{
		{"Sum", data.TotalTypeSum},
		{"Min", data.TotalTypeMin},
		{"Max", data.TotalTypeMax},
		{"Avg", data.TotalTypeAvg},
		{"Count", data.TotalTypeCount},
		{"CountDistinct", data.TotalTypeCountDistinct},
		{"Unknown", data.TotalTypeSum}, // default
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.totalTypeStr, func(t *testing.T) {
			frx := `<?xml version="1.0" encoding="utf-8"?><Report>
				<Dictionary>
					<Total Name="T1" Expression="[Val]" TotalType="` + tc.totalTypeStr + `"/>
				</Dictionary>
				<ReportPage Name="Page1"/>
			</Report>`

			r := reportpkg.NewReport()
			if err := r.LoadFromString(frx); err != nil {
				t.Fatalf("LoadFromString: %v", err)
			}
			totals := r.Dictionary().Totals()
			if len(totals) != 1 {
				t.Fatalf("expected 1 total, got %d", len(totals))
			}
			if totals[0].TotalType != tc.wantType {
				t.Errorf("TotalType for %q: got %v, want %v", tc.totalTypeStr, totals[0].TotalType, tc.wantType)
			}
		})
	}
}

// ── deserializeJsonConnection ──────────────────────────────────────────────

func TestDeserializeJsonConnection(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<JsonDataConnection Name="MyJsonConn" ConnectionString="/data/products.json" Enabled="true">
				<TableDataSource Name="Products" Alias="Products" TableName="$.products"/>
			</JsonDataConnection>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`

	r := reportpkg.NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}

	// Verify connection was registered.
	conns := r.Dictionary().Connections()
	if len(conns) == 0 {
		t.Fatal("expected at least one connection")
	}
	found := false
	for _, c := range conns {
		if c.Name() == "MyJsonConn" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected connection named MyJsonConn")
	}

	// Verify data source was registered.
	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Fatal("expected at least one data source")
	}
	foundDS := false
	for _, ds := range sources {
		if ds.Name() == "Products" {
			foundDS = true
			break
		}
	}
	if !foundDS {
		t.Error("expected data source named Products")
	}
}

func TestDeserializeJsonTableDataSource_NoChildren(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<JsonDataConnection Name="JConn" ConnectionString="/f.json">
				<TableDataSource Name="Items" TableName="$.items"/>
			</JsonDataConnection>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`

	r := reportpkg.NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Fatal("expected Items data source")
	}
}

// ── deserializeXmlConnection ──────────────────────────────────────────────

func TestDeserializeXmlConnection(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<XmlDataConnection Name="MyXmlConn" ConnectionString="/data/nwind.xml" Enabled="true">
				<TableDataSource Name="Categories" Alias="Categories" TableName="/Northwind/Categories/Category"/>
			</XmlDataConnection>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`

	r := reportpkg.NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}

	conns := r.Dictionary().Connections()
	if len(conns) == 0 {
		t.Fatal("expected at least one connection")
	}
	found := false
	for _, c := range conns {
		if c.Name() == "MyXmlConn" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected connection named MyXmlConn")
	}

	sources := r.Dictionary().DataSources()
	if len(sources) == 0 {
		t.Fatal("expected at least one data source")
	}
	foundDS := false
	for _, ds := range sources {
		if ds.Name() == "Categories" {
			foundDS = true
			break
		}
	}
	if !foundDS {
		t.Error("expected data source named Categories")
	}
}

func TestDeserializeXmlConnection_MultipleTableSources(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<XmlDataConnection Name="XConn" ConnectionString="/data.xml">
				<TableDataSource Name="Orders" TableName="/Root/Orders/Order"/>
				<TableDataSource Name="Customers" TableName="/Root/Customers/Customer"/>
			</XmlDataConnection>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`

	r := reportpkg.NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	sources := r.Dictionary().DataSources()
	if len(sources) < 2 {
		t.Errorf("expected 2 data sources, got %d", len(sources))
	}
}

func TestDeserializeXmlConnection_Disabled(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<XmlDataConnection Name="Disabled" ConnectionString="/d.xml" Enabled="false">
				<TableDataSource Name="DS1" TableName="/root"/>
			</XmlDataConnection>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`

	r := reportpkg.NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	conns := r.Dictionary().Connections()
	for _, c := range conns {
		if c.Name() == "Disabled" && c.Enabled() {
			t.Error("disabled connection should have Enabled=false")
		}
	}
}

// ── CalcText edge cases ────────────────────────────────────────────────────

func TestReport_CalcText_ErrorBranch(t *testing.T) {
	// When Calc returns an error, CalcText emits the raw [bracket] expression.
	r := reportpkg.NewReport()
	// "UndefinedVar" is not in the dictionary → will fail expression evaluation.
	result, err := r.CalcText("Value=[UndefinedVar]")
	if err != nil {
		t.Fatalf("CalcText should not return an error even when Calc fails: %v", err)
	}
	// On error the raw bracket expression is emitted.
	if !strings.Contains(result, "UndefinedVar") {
		t.Errorf("CalcText error branch should emit raw expression, got %q", result)
	}
}

func TestReport_CalcText_MultipleExpressions(t *testing.T) {
	r := reportpkg.NewReport()
	r.Dictionary().AddParameter(&data.Parameter{Name: "First", Value: "Hello"})
	r.Dictionary().AddParameter(&data.Parameter{Name: "Last", Value: "World"})

	result, err := r.CalcText("[First] [Last]!")
	if err != nil {
		t.Fatalf("CalcText: %v", err)
	}
	if result != "Hello World!" {
		t.Errorf("got %q, want 'Hello World!'", result)
	}
}

func TestReport_CalcText_EmptyTemplate(t *testing.T) {
	r := reportpkg.NewReport()
	result, err := r.CalcText("")
	if err != nil {
		t.Fatalf("CalcText('') returned error: %v", err)
	}
	if result != "" {
		t.Errorf("CalcText('') should return empty string, got %q", result)
	}
}

// ── LoadFrom — gzip detection ──────────────────────────────────────────────

func TestReport_LoadFrom_GzipStream(t *testing.T) {
	// Build an uncompressed report, then gzip it manually and reload.
	r := reportpkg.NewReport()
	r.Info.Name = "GzipTest"
	pg := reportpkg.NewReportPage()
	pg.SetName("Page1")
	r.AddPage(pg)

	// Save uncompressed.
	r.Compressed = false
	plain, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}

	// Gzip it.
	var gz bytes.Buffer
	w := gzip.NewWriter(&gz)
	if _, err := w.Write([]byte(plain)); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	w.Close()

	r2 := reportpkg.NewReport()
	if err := r2.LoadFrom(bytes.NewReader(gz.Bytes())); err != nil {
		t.Fatalf("LoadFrom (manually gzipped): %v", err)
	}
	if r2.Info.Name != "GzipTest" {
		t.Errorf("ReportName: got %q", r2.Info.Name)
	}
}

// ── AddBandByTypeName — full coverage ─────────────────────────────────────

func TestPage_AddBandByTypeName_AllTypes(t *testing.T) {
	// Verify that all FRX type names route to the right slot.
	tests := []struct {
		typeName  string
		checkSlot func(*reportpkg.ReportPage) bool
	}{
		{"PageHeader", func(p *reportpkg.ReportPage) bool { return p.PageHeader() != nil }},
		{"PageHeaderBand", func(p *reportpkg.ReportPage) bool { return p.PageHeader() != nil }},
		{"PageFooter", func(p *reportpkg.ReportPage) bool { return p.PageFooter() != nil }},
		{"PageFooterBand", func(p *reportpkg.ReportPage) bool { return p.PageFooter() != nil }},
		{"ReportTitle", func(p *reportpkg.ReportPage) bool { return p.ReportTitle() != nil }},
		{"ReportTitleBand", func(p *reportpkg.ReportPage) bool { return p.ReportTitle() != nil }},
		{"ReportSummary", func(p *reportpkg.ReportPage) bool { return p.ReportSummary() != nil }},
		{"ReportSummaryBand", func(p *reportpkg.ReportPage) bool { return p.ReportSummary() != nil }},
		{"ColumnHeader", func(p *reportpkg.ReportPage) bool { return p.ColumnHeader() != nil }},
		{"ColumnHeaderBand", func(p *reportpkg.ReportPage) bool { return p.ColumnHeader() != nil }},
		{"ColumnFooter", func(p *reportpkg.ReportPage) bool { return p.ColumnFooter() != nil }},
		{"ColumnFooterBand", func(p *reportpkg.ReportPage) bool { return p.ColumnFooter() != nil }},
		{"Overlay", func(p *reportpkg.ReportPage) bool { return p.Overlay() != nil }},
		{"OverlayBand", func(p *reportpkg.ReportPage) bool { return p.Overlay() != nil }},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.typeName, func(t *testing.T) {
			// Load the band from registry via FRX round-trip.
			frx := `<?xml version="1.0" encoding="utf-8"?><Report>
				<ReportPage Name="Page1">
					<` + tc.typeName + ` Name="B1" Height="20"/>
				</ReportPage>
			</Report>`
			r := reportpkg.NewReport()
			if err := r.LoadFromString(frx); err != nil {
				t.Fatalf("LoadFromString(%s): %v", tc.typeName, err)
			}
			if r.PageCount() == 0 {
				t.Fatal("no pages")
			}
			if !tc.checkSlot(r.Page(0)) {
				t.Errorf("band type %q did not populate expected slot", tc.typeName)
			}
		})
	}
}

// ── Dictionary deserialization — edge cases ────────────────────────────────

func TestDeserializeDictionary_MultipleParameters(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<Parameter Name="Year" DataType="System.Int32" Expression="2024"/>
			<Parameter Name="Month" DataType="System.Int32" Expression="6"/>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`

	r := reportpkg.NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	params := r.Dictionary().Parameters()
	if len(params) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(params))
	}
}

func TestDeserializeDictionary_Relation(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<Relation Name="OrderCustomer" ParentDataSource="Customers" ChildDataSource="Orders" ParentColumns="ID" ChildColumns="CustomerID"/>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`

	r := reportpkg.NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	rels := r.Dictionary().Relations()
	if len(rels) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(rels))
	}
	rel := rels[0]
	if rel.Name != "OrderCustomer" {
		t.Errorf("Relation.Name: %q", rel.Name)
	}
	if rel.ParentSourceName != "Customers" {
		t.Errorf("ParentDataSource: %q", rel.ParentSourceName)
	}
	if rel.ChildSourceName != "Orders" {
		t.Errorf("ChildDataSource: %q", rel.ChildSourceName)
	}
}

func TestDeserializeDictionary_Total_AllTypes(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<Total Name="T_Avg" Expression="[Price]" TotalType="Avg" Evaluator="DataBand1" PrintOn="PageFooter"/>
			<Total Name="T_Max" Expression="[Qty]" TotalType="Max"/>
			<Total Name="T_Count" Expression="[ID]" TotalType="Count"/>
			<Total Name="T_CountDistinct" Expression="[Cat]" TotalType="CountDistinct"/>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`

	r := reportpkg.NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	totals := r.Dictionary().Totals()
	if len(totals) != 4 {
		t.Fatalf("expected 4 totals, got %d", len(totals))
	}
}

func TestDeserializeDictionary_UnknownChild_Skipped(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report>
		<Dictionary>
			<SomeUnknownElement Name="Unknown"/>
			<Parameter Name="Known"/>
		</Dictionary>
		<ReportPage Name="Page1"/>
	</Report>`

	r := reportpkg.NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	params := r.Dictionary().Parameters()
	if len(params) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(params))
	}
}

// ── report.Deserialize — full field coverage ──────────────────────────────

func TestReport_Deserialize_AllFields(t *testing.T) {
	frx := `<?xml version="1.0" encoding="utf-8"?><Report
		ReportName="FullReport"
		ReportAuthor="Bob"
		ReportDescription="A full test"
		ReportVersion="3.0"
		Created="2023-01-01"
		Modified="2023-12-31"
		CreatorVersion="2023.4.0"
		SavePreviewPicture="true"
		Compressed="false"
		ConvertNulls="true"
		DoublePass="true"
		InitialPageNumber="5"
		MaxPages="20"
		StartReportEvent="OnStart"
		FinishReportEvent="OnFinish">
		<ReportPage Name="P1"/>
	</Report>`

	r := reportpkg.NewReport()
	if err := r.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if r.Info.Name != "FullReport" {
		t.Errorf("Name: %q", r.Info.Name)
	}
	if r.Info.Author != "Bob" {
		t.Errorf("Author: %q", r.Info.Author)
	}
	if r.Info.Description != "A full test" {
		t.Errorf("Description: %q", r.Info.Description)
	}
	if r.Info.Version != "3.0" {
		t.Errorf("Version: %q", r.Info.Version)
	}
	if r.Info.Created != "2023-01-01" {
		t.Errorf("Created: %q", r.Info.Created)
	}
	if r.Info.Modified != "2023-12-31" {
		t.Errorf("Modified: %q", r.Info.Modified)
	}
	if r.Info.CreatorVersion != "2023.4.0" {
		t.Errorf("CreatorVersion: %q", r.Info.CreatorVersion)
	}
	if !r.Info.SavePreviewPicture {
		t.Error("SavePreviewPicture should be true")
	}
	if !r.ConvertNulls {
		t.Error("ConvertNulls should be true")
	}
	if !r.DoublePass {
		t.Error("DoublePass should be true")
	}
	if r.InitialPageNumber != 5 {
		t.Errorf("InitialPageNumber: %d", r.InitialPageNumber)
	}
	if r.MaxPages != 20 {
		t.Errorf("MaxPages: %d", r.MaxPages)
	}
	if r.StartReportEvent != "OnStart" {
		t.Errorf("StartReportEvent: %q", r.StartReportEvent)
	}
	if r.FinishReportEvent != "OnFinish" {
		t.Errorf("FinishReportEvent: %q", r.FinishReportEvent)
	}
}

// ── style.FillColor and TextColor serialization ─────────────────────────

func TestStylesSerializer_Serialize_FillAndTextColor(t *testing.T) {
	r := reportpkg.NewReport()
	ss := r.Styles()
	e := &style.StyleEntry{
		Name:      "ColorTest",
		FillColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},   // red
		TextColor: color.RGBA{R: 0, G: 0, B: 255, A: 255},   // blue
	}
	ss.Add(e)

	xml, err := r.SaveToString()
	if err != nil {
		t.Fatalf("SaveToString: %v", err)
	}
	if !strings.Contains(xml, "Fill.Color") {
		t.Error("expected Fill.Color in XML")
	}
	if !strings.Contains(xml, "TextFill.Color") {
		t.Error("expected TextFill.Color in XML")
	}

	// Round-trip.
	r2 := reportpkg.NewReport()
	if err := r2.LoadFromString(xml); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	e2 := r2.Styles().Find("ColorTest")
	if e2 == nil {
		t.Fatal("ColorTest style not found after round-trip")
	}
}

// ── LoadFrom — invalid stream ──────────────────────────────────────────────

func TestReport_LoadFrom_InvalidXML(t *testing.T) {
	r := reportpkg.NewReport()
	err := r.LoadFromString("not xml at all")
	if err == nil {
		t.Error("expected error loading invalid XML")
	}
}

func TestReport_LoadFrom_EmptyStream(t *testing.T) {
	r := reportpkg.NewReport()
	err := r.LoadFrom(bytes.NewReader([]byte{}))
	if err == nil {
		t.Error("expected error loading empty stream")
	}
}

func TestReport_LoadFrom_WrongRootElement(t *testing.T) {
	r := reportpkg.NewReport()
	err := r.LoadFromString(`<?xml version="1.0"?><NotAReport Name="X"/>`)
	if err == nil {
		t.Error("expected error for wrong root element")
	}
}

// ── parseBorderLines — full coverage ─────────────────────────────────────

func TestParseBorderLines_ViaStyle(t *testing.T) {
	// parseBorderLines is called during deserializeStyleEntry.
	cases := []struct {
		frxLines string
		wantMask style.BorderLines
	}{
		{"Left, Right", style.BorderLinesLeft | style.BorderLinesRight},
		{"Top, Bottom", style.BorderLinesTop | style.BorderLinesBottom},
		{"All", style.BorderLinesAll},
		{"None", style.BorderLinesNone},
		{"Left", style.BorderLinesLeft},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.frxLines, func(t *testing.T) {
			frx := `<?xml version="1.0" encoding="utf-8"?><Report>
				<Styles>
					<Style Name="S1" Border.Lines="` + tc.frxLines + `"/>
				</Styles>
				<ReportPage Name="Page1"/>
			</Report>`
			r := reportpkg.NewReport()
			if err := r.LoadFromString(frx); err != nil {
				t.Fatalf("LoadFromString: %v", err)
			}
			e := r.Styles().Find("S1")
			if e == nil {
				t.Fatal("S1 style not found")
			}
			if e.Border.VisibleLines != tc.wantMask {
				t.Errorf("Border.Lines for %q: got %v, want %v", tc.frxLines, e.Border.VisibleLines, tc.wantMask)
			}
		})
	}
}
