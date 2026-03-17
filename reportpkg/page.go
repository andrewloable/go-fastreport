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

	// BackPage references another ReportPage by name. When set, the referenced
	// page's content is rendered as a background layer behind this page's content.
	// This mirrors FastReport's ReportPage.BackPage property.
	BackPage string

	// BackPageOddEven controls when the back page is applied:
	//   0 = always (both odd and even pages)
	//   1 = odd pages only
	//   2 = even pages only
	BackPageOddEven int

	// Script event names.
	CreatePageEvent  string
	StartPageEvent   string
	FinishPageEvent  string
	ManualBuildEvent string

	// Outline expression for preview navigator.
	OutlineExpression string

	// Page-level columns.
	Columns PageColumns

	// UnlimitedHeight, when true, prevents page breaks so the page grows
	// to fit all content. Mirrors FastReport's ReportPage.UnlimitedHeight.
	UnlimitedHeight bool

	// Watermark is the optional page watermark (text or image).
	Watermark *Watermark

	// inherited marks this page as coming from a base (parent) report.
	inherited bool
}

// TypeName returns the FRX element name used during serialization.
func (*ReportPage) TypeName() string { return "ReportPage" }

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
	p.Watermark = NewWatermark()
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
// Singleton bands (PageHeader, PageFooter, ReportTitle, etc.) are NOT included;
// use AllBands to iterate every band on the page.
func (p *ReportPage) Bands() []report.Base { return p.bands }

// AllBands returns every band on the page — singleton bands (PageHeader,
// ReportTitle, etc.) followed by the ordered data/group bands.
// Nil singleton slots are omitted.
func (p *ReportPage) AllBands() []report.Base {
	var all []report.Base
	if p.reportTitle != nil {
		all = append(all, p.reportTitle)
	}
	if p.pageHeader != nil {
		all = append(all, p.pageHeader)
	}
	if p.columnHeader != nil {
		all = append(all, p.columnHeader)
	}
	all = append(all, p.bands...)
	if p.columnFooter != nil {
		all = append(all, p.columnFooter)
	}
	if p.reportSummary != nil {
		all = append(all, p.reportSummary)
	}
	if p.pageFooter != nil {
		all = append(all, p.pageFooter)
	}
	if p.overlay != nil {
		all = append(all, p.overlay)
	}
	return all
}

// AddBand appends a band to the page.
func (p *ReportPage) AddBand(b report.Base) { p.bands = append(p.bands, b) }

// AddBandByTypeName routes a deserialized band to the appropriate page slot
// based on its FRX type name. Singleton bands (PageHeader, PageFooter, etc.)
// are placed in their dedicated fields; all other bands go into the ordered list.
// Both short names ("PageHeader") and full FastReport names ("PageHeaderBand") are accepted.
func (p *ReportPage) AddBandByTypeName(typeName string, b report.Base) {
	switch typeName {
	case "PageHeader", "PageHeaderBand":
		if pb, ok := b.(*band.PageHeaderBand); ok {
			p.pageHeader = pb
		}
	case "PageFooter", "PageFooterBand":
		if pb, ok := b.(*band.PageFooterBand); ok {
			p.pageFooter = pb
		}
	case "ReportTitle", "ReportTitleBand":
		if pb, ok := b.(*band.ReportTitleBand); ok {
			p.reportTitle = pb
		}
	case "ReportSummary", "ReportSummaryBand":
		if pb, ok := b.(*band.ReportSummaryBand); ok {
			p.reportSummary = pb
		}
	case "ColumnHeader", "ColumnHeaderBand":
		if pb, ok := b.(*band.ColumnHeaderBand); ok {
			p.columnHeader = pb
		}
	case "ColumnFooter", "ColumnFooterBand":
		if pb, ok := b.(*band.ColumnFooterBand); ok {
			p.columnFooter = pb
		}
	case "Overlay", "OverlayBand":
		if pb, ok := b.(*band.OverlayBand); ok {
			p.overlay = pb
		}
	default:
		p.bands = append(p.bands, b)
	}
}

// --- Style ---

// Border returns the page border.
func (p *ReportPage) Border() style.Border { return p.border }

// SetBorder sets the page border.
func (p *ReportPage) SetBorder(b style.Border) { p.border = b }

// Fill returns the page background fill.
func (p *ReportPage) Fill() style.Fill { return p.fill }

// SetFill sets the page background fill.
func (p *ReportPage) SetFill(f style.Fill) { p.fill = f }

// Inherited returns true if this page originates from a base report.
func (p *ReportPage) Inherited() bool { return p.inherited }

// SetInherited marks the page as inherited.
func (p *ReportPage) SetInherited(v bool) { p.inherited = v }

// Clone returns a shallow copy of the page (pointer fields shared).
func (p *ReportPage) Clone() *ReportPage {
	cp := *p
	// Copy band slice so appends don't affect original.
	cp.bands = make([]report.Base, len(p.bands))
	copy(cp.bands, p.bands)
	return &cp
}

// mergeFromBase adds bands from base page that are not already present
// in p (matched by Name).
func (p *ReportPage) mergeFromBase(base *ReportPage) {
	childNames := make(map[string]bool, len(p.bands))
	for _, b := range p.bands {
		childNames[b.Name()] = true
	}

	// Inherit slot-bands from base if child doesn't have them.
	if p.pageHeader == nil && base.pageHeader != nil {
		p.pageHeader = base.pageHeader
	}
	if p.reportTitle == nil && base.reportTitle != nil {
		p.reportTitle = base.reportTitle
	}
	if p.columnHeader == nil && base.columnHeader != nil {
		p.columnHeader = base.columnHeader
	}
	if p.reportSummary == nil && base.reportSummary != nil {
		p.reportSummary = base.reportSummary
	}
	if p.columnFooter == nil && base.columnFooter != nil {
		p.columnFooter = base.columnFooter
	}
	if p.pageFooter == nil && base.pageFooter != nil {
		p.pageFooter = base.pageFooter
	}
	if p.overlay == nil && base.overlay != nil {
		p.overlay = base.overlay
	}

	// Prepend ordered base bands that the child doesn't have.
	var inherited []report.Base
	for _, b := range base.bands {
		if !childNames[b.Name()] {
			inherited = append(inherited, b)
		}
	}
	if len(inherited) > 0 {
		p.bands = append(inherited, p.bands...)
	}
}

// --- Serialization ---

// Serialize writes ReportPage properties that differ from defaults,
// then writes all non-nil band slots as child elements.
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
	if p.UnlimitedHeight {
		w.WriteBool("UnlimitedHeight", true)
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
	if p.BackPage != "" {
		w.WriteStr("BackPage", p.BackPage)
	}
	if p.BackPageOddEven != 0 {
		w.WriteInt("BackPageOddEven", p.BackPageOddEven)
	}

	// Watermark properties are written as flat attributes on the page element.
	if p.Watermark != nil {
		p.Watermark.Serialize(w)
	}

	// Write bands in the same order as FastReport's GetChildObjects().
	// Singleton bands use their declared FRX element name via TypeNamer.
	if err := p.serializeBands(w); err != nil {
		return err
	}
	return nil
}

// serializeBands writes all non-nil band slots as child XML elements.
// Order mirrors FastReport's ReportPage.GetChildObjects().
func (p *ReportPage) serializeBands(w report.Writer) error {
	// Header bands.
	if p.TitleBeforeHeader {
		if p.reportTitle != nil {
			if err := w.WriteObject(p.reportTitle); err != nil {
				return err
			}
		}
		if p.pageHeader != nil {
			if err := w.WriteObject(p.pageHeader); err != nil {
				return err
			}
		}
	} else {
		if p.pageHeader != nil {
			if err := w.WriteObject(p.pageHeader); err != nil {
				return err
			}
		}
		if p.reportTitle != nil {
			if err := w.WriteObject(p.reportTitle); err != nil {
				return err
			}
		}
	}
	if p.columnHeader != nil {
		if err := w.WriteObject(p.columnHeader); err != nil {
			return err
		}
	}

	// Dynamic bands (DataBand, GroupHeaderBand, etc.).
	for _, b := range p.bands {
		if b == nil {
			continue
		}
		if err := w.WriteObject(b); err != nil {
			return err
		}
	}

	// Footer bands.
	if p.reportSummary != nil {
		if err := w.WriteObject(p.reportSummary); err != nil {
			return err
		}
	}
	if p.columnFooter != nil {
		if err := w.WriteObject(p.columnFooter); err != nil {
			return err
		}
	}
	if p.pageFooter != nil {
		if err := w.WriteObject(p.pageFooter); err != nil {
			return err
		}
	}
	if p.overlay != nil {
		if err := w.WriteObject(p.overlay); err != nil {
			return err
		}
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
	p.UnlimitedHeight = r.ReadBool("UnlimitedHeight", false)
	p.OutlineExpression = r.ReadStr("OutlineExpression", "")
	p.CreatePageEvent = r.ReadStr("CreatePageEvent", "")
	p.StartPageEvent = r.ReadStr("StartPageEvent", "")
	p.FinishPageEvent = r.ReadStr("FinishPageEvent", "")
	p.ManualBuildEvent = r.ReadStr("ManualBuildEvent", "")
	p.BackPage = r.ReadStr("BackPage", "")
	p.BackPageOddEven = r.ReadInt("BackPageOddEven", 0)
	p.Columns.Count = r.ReadInt("Columns.Count", 0)
	p.Columns.Width = r.ReadFloat("Columns.Width", 0)
	if p.Watermark == nil {
		p.Watermark = NewWatermark()
	}
	p.Watermark.Deserialize(r)
	return nil
}
