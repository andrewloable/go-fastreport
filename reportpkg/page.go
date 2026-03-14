// Package reportpkg contains Report and ReportPage — the top-level containers
// for a go-fastreport report definition.
package reportpkg

import (
	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
)

// PageColumns holds multi-column page-level layout settings.
type PageColumns struct {
	// Count is the number of columns (0 or 1 = single column).
	Count int
	// Width is the column width in pixels (0 = auto).
	Width float32
}

// ReportPage represents a single page template in the report.
// It holds bands that define what prints on each physical page.
// It is the Go equivalent of FastReport.ReportPage.
type ReportPage struct {
	report.BaseObject

	// Paper dimensions in millimetres.
	PaperWidth  float32 // default 210 (A4)
	PaperHeight float32 // default 297 (A4)
	Landscape   bool

	// Margins in millimetres.
	LeftMargin   float32 // default 10
	TopMargin    float32 // default 10
	RightMargin  float32 // default 10
	BottomMargin float32 // default 10
	MirrorMargins bool

	// Page-level styling.
	border style.Border
	fill   style.Fill

	// Band slots.
	pageHeader    *band.PageHeaderBand
	reportTitle   *band.ReportTitleBand
	columnHeader  *band.ColumnHeaderBand
	reportSummary *band.ReportSummaryBand
	columnFooter  *band.ColumnFooterBand
	pageFooter    *band.PageFooterBand
	overlay       *band.OverlayBand

	// Data bands and group headers (ordered).
	bands []report.Base

	// Behaviour flags.
	TitleBeforeHeader  bool
	PrintOnPreviousPage bool
	ResetPageNumber    bool
	StartOnOddPage     bool

	// Script event names.
	CreatePageEvent  string
	StartPageEvent   string
	FinishPageEvent  string
	ManualBuildEvent string

	// Outline expression for preview navigator.
	OutlineExpression string

	// Page-level columns.
	Columns PageColumns
}

// NewReportPage creates a ReportPage with A4 defaults.
func NewReportPage() *ReportPage {
	p := &ReportPage{
		BaseObject:   *report.NewBaseObject(),
		PaperWidth:   210,
		PaperHeight:  297,
		LeftMargin:   10,
		TopMargin:    10,
		RightMargin:  10,
		BottomMargin: 10,
	}
	p.fill = &style.SolidFill{Color: style.ColorWhite}
	return p
}

// --- Band accessors ---

// PageHeader returns the page header band.
func (p *ReportPage) PageHeader() *band.PageHeaderBand { return p.pageHeader }

// SetPageHeader sets the page header band.
func (p *ReportPage) SetPageHeader(b *band.PageHeaderBand) { p.pageHeader = b }

// ReportTitle returns the report title band.
func (p *ReportPage) ReportTitle() *band.ReportTitleBand { return p.reportTitle }

// SetReportTitle sets the report title band.
func (p *ReportPage) SetReportTitle(b *band.ReportTitleBand) { p.reportTitle = b }

// ColumnHeader returns the column header band.
func (p *ReportPage) ColumnHeader() *band.ColumnHeaderBand { return p.columnHeader }

// SetColumnHeader sets the column header band.
func (p *ReportPage) SetColumnHeader(b *band.ColumnHeaderBand) { p.columnHeader = b }

// ReportSummary returns the report summary band.
func (p *ReportPage) ReportSummary() *band.ReportSummaryBand { return p.reportSummary }

// SetReportSummary sets the report summary band.
func (p *ReportPage) SetReportSummary(b *band.ReportSummaryBand) { p.reportSummary = b }

// ColumnFooter returns the column footer band.
func (p *ReportPage) ColumnFooter() *band.ColumnFooterBand { return p.columnFooter }

// SetColumnFooter sets the column footer band.
func (p *ReportPage) SetColumnFooter(b *band.ColumnFooterBand) { p.columnFooter = b }

// PageFooter returns the page footer band.
func (p *ReportPage) PageFooter() *band.PageFooterBand { return p.pageFooter }

// SetPageFooter sets the page footer band.
func (p *ReportPage) SetPageFooter(b *band.PageFooterBand) { p.pageFooter = b }

// Overlay returns the overlay band.
func (p *ReportPage) Overlay() *band.OverlayBand { return p.overlay }

// SetOverlay sets the overlay band.
func (p *ReportPage) SetOverlay(b *band.OverlayBand) { p.overlay = b }

// Bands returns the ordered slice of data/group bands on this page.
func (p *ReportPage) Bands() []report.Base { return p.bands }

// AddBand appends a band to the page.
func (p *ReportPage) AddBand(b report.Base) { p.bands = append(p.bands, b) }

// --- Style ---

// Border returns the page border.
func (p *ReportPage) Border() style.Border { return p.border }

// SetBorder sets the page border.
func (p *ReportPage) SetBorder(b style.Border) { p.border = b }

// Fill returns the page background fill.
func (p *ReportPage) Fill() style.Fill { return p.fill }

// SetFill sets the page background fill.
func (p *ReportPage) SetFill(f style.Fill) { p.fill = f }

// --- Serialization ---

// Serialize writes ReportPage properties that differ from defaults.
func (p *ReportPage) Serialize(w report.Writer) error {
	if err := p.BaseObject.Serialize(w); err != nil {
		return err
	}
	if p.PaperWidth != 210 {
		w.WriteFloat("PaperWidth", p.PaperWidth)
	}
	if p.PaperHeight != 297 {
		w.WriteFloat("PaperHeight", p.PaperHeight)
	}
	if p.Landscape {
		w.WriteBool("Landscape", true)
	}
	if p.LeftMargin != 10 {
		w.WriteFloat("LeftMargin", p.LeftMargin)
	}
	if p.TopMargin != 10 {
		w.WriteFloat("TopMargin", p.TopMargin)
	}
	if p.RightMargin != 10 {
		w.WriteFloat("RightMargin", p.RightMargin)
	}
	if p.BottomMargin != 10 {
		w.WriteFloat("BottomMargin", p.BottomMargin)
	}
	if p.MirrorMargins {
		w.WriteBool("MirrorMargins", true)
	}
	if p.TitleBeforeHeader {
		w.WriteBool("TitleBeforeHeader", true)
	}
	if p.PrintOnPreviousPage {
		w.WriteBool("PrintOnPreviousPage", true)
	}
	if p.ResetPageNumber {
		w.WriteBool("ResetPageNumber", true)
	}
	if p.StartOnOddPage {
		w.WriteBool("StartOnOddPage", true)
	}
	if p.OutlineExpression != "" {
		w.WriteStr("OutlineExpression", p.OutlineExpression)
	}
	if p.CreatePageEvent != "" {
		w.WriteStr("CreatePageEvent", p.CreatePageEvent)
	}
	if p.StartPageEvent != "" {
		w.WriteStr("StartPageEvent", p.StartPageEvent)
	}
	if p.FinishPageEvent != "" {
		w.WriteStr("FinishPageEvent", p.FinishPageEvent)
	}
	if p.ManualBuildEvent != "" {
		w.WriteStr("ManualBuildEvent", p.ManualBuildEvent)
	}
	return nil
}

// Deserialize reads ReportPage properties.
func (p *ReportPage) Deserialize(r report.Reader) error {
	if err := p.BaseObject.Deserialize(r); err != nil {
		return err
	}
	p.PaperWidth = r.ReadFloat("PaperWidth", 210)
	p.PaperHeight = r.ReadFloat("PaperHeight", 297)
	p.Landscape = r.ReadBool("Landscape", false)
	p.LeftMargin = r.ReadFloat("LeftMargin", 10)
	p.TopMargin = r.ReadFloat("TopMargin", 10)
	p.RightMargin = r.ReadFloat("RightMargin", 10)
	p.BottomMargin = r.ReadFloat("BottomMargin", 10)
	p.MirrorMargins = r.ReadBool("MirrorMargins", false)
	p.TitleBeforeHeader = r.ReadBool("TitleBeforeHeader", false)
	p.PrintOnPreviousPage = r.ReadBool("PrintOnPreviousPage", false)
	p.ResetPageNumber = r.ReadBool("ResetPageNumber", false)
	p.StartOnOddPage = r.ReadBool("StartOnOddPage", false)
	p.OutlineExpression = r.ReadStr("OutlineExpression", "")
	p.CreatePageEvent = r.ReadStr("CreatePageEvent", "")
	p.StartPageEvent = r.ReadStr("StartPageEvent", "")
	p.FinishPageEvent = r.ReadStr("FinishPageEvent", "")
	p.ManualBuildEvent = r.ReadStr("ManualBuildEvent", "")
	return nil
}
