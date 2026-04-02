package reportpkg_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/style"
)

// -----------------------------------------------------------------------
// ReportPage tests
// -----------------------------------------------------------------------

func TestNewReportPage_Defaults(t *testing.T) {
	p := reportpkg.NewReportPage()
	if p == nil {
		t.Fatal("NewReportPage returned nil")
	}
	if p.PaperWidth != 210 {
		t.Errorf("PaperWidth default = %v, want 210", p.PaperWidth)
	}
	if p.PaperHeight != 297 {
		t.Errorf("PaperHeight default = %v, want 297", p.PaperHeight)
	}
	if p.Landscape {
		t.Error("Landscape should default to false")
	}
	if p.LeftMargin != 10 || p.TopMargin != 10 || p.RightMargin != 10 || p.BottomMargin != 10 {
		t.Errorf("Margins default = L=%v T=%v R=%v B=%v, want all 10",
			p.LeftMargin, p.TopMargin, p.RightMargin, p.BottomMargin)
	}
	if p.MirrorMargins {
		t.Error("MirrorMargins should default to false")
	}
	if !p.TitleBeforeHeader {
		t.Error("TitleBeforeHeader should default to true (C# [DefaultValue(true)])")
	}
	if p.Fill() == nil {
		t.Error("Fill should default to non-nil (white)")
	}
}

func TestReportPage_Landscape(t *testing.T) {
	p := reportpkg.NewReportPage()
	p.Landscape = true
	if !p.Landscape {
		t.Error("Landscape should be true")
	}
}

func TestReportPage_Margins(t *testing.T) {
	p := reportpkg.NewReportPage()
	p.LeftMargin = 20
	p.TopMargin = 25
	p.RightMargin = 20
	p.BottomMargin = 25
	if p.LeftMargin != 20 || p.TopMargin != 25 {
		t.Error("Margins not set correctly")
	}
}

func TestReportPage_BandSlots(t *testing.T) {
	p := reportpkg.NewReportPage()

	ph := band.NewPageHeaderBand()
	rt := band.NewReportTitleBand()
	ch := band.NewColumnHeaderBand()
	rs := band.NewReportSummaryBand()
	cf := band.NewColumnFooterBand()
	pf := band.NewPageFooterBand()
	ov := band.NewOverlayBand()

	p.SetPageHeader(ph)
	p.SetReportTitle(rt)
	p.SetColumnHeader(ch)
	p.SetReportSummary(rs)
	p.SetColumnFooter(cf)
	p.SetPageFooter(pf)
	p.SetOverlay(ov)

	if p.PageHeader() != ph {
		t.Error("PageHeader should be set")
	}
	if p.ReportTitle() != rt {
		t.Error("ReportTitle should be set")
	}
	if p.ColumnHeader() != ch {
		t.Error("ColumnHeader should be set")
	}
	if p.ReportSummary() != rs {
		t.Error("ReportSummary should be set")
	}
	if p.ColumnFooter() != cf {
		t.Error("ColumnFooter should be set")
	}
	if p.PageFooter() != pf {
		t.Error("PageFooter should be set")
	}
	if p.Overlay() != ov {
		t.Error("Overlay should be set")
	}
}

func TestReportPage_AddBand(t *testing.T) {
	p := reportpkg.NewReportPage()
	db := band.NewDataBand()
	p.AddBand(db)
	if len(p.Bands()) != 1 {
		t.Errorf("Bands len = %d, want 1", len(p.Bands()))
	}
	if p.Bands()[0] != db {
		t.Error("Band at index 0 should be the data band")
	}
}

func TestReportPage_Fill(t *testing.T) {
	p := reportpkg.NewReportPage()
	f := &style.SolidFill{}
	p.SetFill(f)
	if p.Fill() != f {
		t.Error("Fill should be set")
	}
}

func TestReportPage_Border(t *testing.T) {
	p := reportpkg.NewReportPage()
	b := style.Border{Shadow: true}
	p.SetBorder(b)
	if !p.Border().Shadow {
		t.Error("Border.Shadow should be true")
	}
}

func TestReportPage_EventNames(t *testing.T) {
	p := reportpkg.NewReportPage()
	p.CreatePageEvent = "Page1_Create"
	p.StartPageEvent = "Page1_Start"
	p.FinishPageEvent = "Page1_Finish"
	p.ManualBuildEvent = "Page1_Build"
	if p.CreatePageEvent != "Page1_Create" {
		t.Errorf("CreatePageEvent = %q", p.CreatePageEvent)
	}
}

func TestReportPage_OutlineExpression(t *testing.T) {
	p := reportpkg.NewReportPage()
	p.OutlineExpression = "[PageNo]"
	if p.OutlineExpression != "[PageNo]" {
		t.Errorf("OutlineExpression = %q", p.OutlineExpression)
	}
}

// -----------------------------------------------------------------------
// Report tests
// -----------------------------------------------------------------------

func TestNewReport_Defaults(t *testing.T) {
	r := reportpkg.NewReport()
	if r == nil {
		t.Fatal("NewReport returned nil")
	}
	if r.PageCount() != 0 {
		t.Errorf("PageCount default = %d, want 0", r.PageCount())
	}
	if r.InitialPageNumber != 1 {
		t.Errorf("InitialPageNumber default = %d, want 1", r.InitialPageNumber)
	}
	if r.MaxPages != 0 {
		t.Errorf("MaxPages default = %d, want 0", r.MaxPages)
	}
	if r.Compressed {
		t.Error("Compressed should default to false")
	}
	if r.DoublePass {
		t.Error("DoublePass should default to false")
	}
}

func TestReport_AddPage(t *testing.T) {
	r := reportpkg.NewReport()
	p1 := reportpkg.NewReportPage()
	p2 := reportpkg.NewReportPage()
	r.AddPage(p1)
	r.AddPage(p2)
	if r.PageCount() != 2 {
		t.Errorf("PageCount = %d, want 2", r.PageCount())
	}
	if r.Page(0) != p1 {
		t.Error("Page(0) should be p1")
	}
	if r.Page(1) != p2 {
		t.Error("Page(1) should be p2")
	}
}

func TestReport_RemovePage(t *testing.T) {
	r := reportpkg.NewReport()
	p1 := reportpkg.NewReportPage()
	p2 := reportpkg.NewReportPage()
	r.AddPage(p1)
	r.AddPage(p2)
	r.RemovePage(p1)
	if r.PageCount() != 1 {
		t.Errorf("PageCount after remove = %d, want 1", r.PageCount())
	}
	if r.Page(0) != p2 {
		t.Error("Page(0) after remove should be p2")
	}
}

func TestReport_RemovePage_NotFound(t *testing.T) {
	r := reportpkg.NewReport()
	p := reportpkg.NewReportPage()
	r.RemovePage(p) // should not panic
}

func TestReport_Info(t *testing.T) {
	r := reportpkg.NewReport()
	r.Info.Name = "Sales Report"
	r.Info.Author = "Alice"
	r.Info.Description = "Monthly sales"
	if r.Info.Name != "Sales Report" {
		t.Errorf("Info.Name = %q", r.Info.Name)
	}
	if r.Info.Author != "Alice" {
		t.Errorf("Info.Author = %q", r.Info.Author)
	}
}

func TestReport_ScriptFlags(t *testing.T) {
	r := reportpkg.NewReport()
	r.ConvertNulls = true
	r.DoublePass = true
	r.MaxPages = 100
	if !r.ConvertNulls || !r.DoublePass || r.MaxPages != 100 {
		t.Error("script flags not set correctly")
	}
}

func TestReport_EventNames(t *testing.T) {
	r := reportpkg.NewReport()
	r.StartReportEvent = "Report_Start"
	r.FinishReportEvent = "Report_Finish"
	if r.StartReportEvent != "Report_Start" {
		t.Errorf("StartReportEvent = %q", r.StartReportEvent)
	}
	if r.FinishReportEvent != "Report_Finish" {
		t.Errorf("FinishReportEvent = %q", r.FinishReportEvent)
	}
}

func TestReport_Pages_Nil(t *testing.T) {
	r := reportpkg.NewReport()
	pages := r.Pages()
	if pages == nil {
		// nil is acceptable for an empty slice
		return
	}
	if len(pages) != 0 {
		t.Errorf("Pages() len = %d, want 0", len(pages))
	}
}
