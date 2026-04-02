// Package reportpkg contains Report and ReportPage — the top-level containers
// for a go-fastreport report definition.
package reportpkg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/units"
)

// parseFloatList parses a comma-separated list of floats (e.g. "0,90").
func parseFloatList(s string) []float32 {
	parts := strings.Split(s, ",")
	result := make([]float32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if v, err := strconv.ParseFloat(p, 32); err == nil {
			result = append(result, float32(v))
		}
	}
	return result
}

// PageColumns holds multi-column page-level layout settings.
type PageColumns struct {
	// Count is the number of columns (0 or 1 = single column).
	Count int
	// Width is the column width in mm (0 = auto).
	Width float32
	// Positions holds the left-edge X offset of each column in mm.
	// Parsed from "Columns.Positions" FRX attribute (comma-separated).
	Positions []float32
}

// SetCount sets the column count and auto-calculates Width and Positions
// from the page paper dimensions. Count must be >= 1.
// C# ref: PageColumns.Count setter (PageColumns.cs:28-41).
func (pc *PageColumns) SetCount(page *ReportPage, count int) error {
	if count <= 0 {
		return fmt.Errorf("PageColumns.Count: value must be greater than 0, got %d", count)
	}
	pc.Count = count
	if page != nil {
		pc.Width = (page.PaperWidth - page.LeftMargin - page.RightMargin) / float32(count)
	}
	pc.Positions = make([]float32, count)
	for i := range pc.Positions {
		pc.Positions[i] = float32(i) * pc.Width
	}
	return nil
}

// Assign copies all fields from src into this PageColumns.
// C# ref: PageColumns.Assign (PageColumns.cs:94-99).
// Note: C# Assign calls Count setter (triggering auto-calculation from page),
// then immediately overrides Width and Positions — net result is identical to
// a direct field copy, which is what this Go implementation does.
func (pc *PageColumns) Assign(src PageColumns) {
	pc.Count = src.Count
	pc.Width = src.Width
	if src.Positions != nil {
		pc.Positions = make([]float32, len(src.Positions))
		copy(pc.Positions, src.Positions)
	} else {
		pc.Positions = nil
	}
}

// ReportPage represents a single page template in the report.
// It holds bands that define what prints on each physical page.
// It is the Go equivalent of FastReport.ReportPage, which inherits from
// FastReport.PageBase (see original-dotnet/FastReport.Base/PageBase.cs and
// original-dotnet/FastReport.Base/ReportPage.cs).
type ReportPage struct {
	report.BaseObject

	// visible controls whether this page template is processed during report run.
	// A page with Visible=false is skipped entirely (used for drill-down/detail pages
	// that are only shown when triggered interactively). Default is true.
	visible bool

	// --- PageBase fields (FastReport.Base/PageBase.cs) ---

	// pageName is an explicit override for the page name shown in the preview
	// navigator. When empty, Name() is used instead.
	// Mirrors PageBase.PageName (internal).
	pageName string

	// needRefresh is set by Refresh() and Modify() to signal the preview window
	// to repaint this page. Mirrors PageBase.NeedRefresh (internal).
	needRefresh bool

	// needModify is set by Modify() to signal that page content has changed and
	// must be saved. Mirrors PageBase.NeedModify (internal).
	needModify bool

	// --- ReportPage fields ---

	// Paper dimensions in millimetres.
	PaperWidth  float32 // default 210 (A4)
	PaperHeight float32 // default 297 (A4)
	Landscape   bool

	// RawPaperSize stores the RawKind value of the selected paper size.
	// Used to distinguish papers with identical dimensions (e.g. "A3" vs
	// "A3 with no margins"). Not required for rendering; FastReport uses
	// PaperWidth/PaperHeight when this is 0.
	// Mirrors FastReport.ReportPage.RawPaperSize.
	RawPaperSize int

	// ExportAlias is an optional page name override used during export.
	// When set, exporters use this value instead of Name().
	// Mirrors FastReport.ReportPage.ExportAlias.
	ExportAlias string

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

	// ExtraDesignWidth gives the page extra width in the report designer.
	// Useful when working with Matrix or Table objects that extend beyond
	// the normal page boundary. Not used at runtime.
	// Mirrors FastReport.ReportPage.ExtraDesignWidth.
	ExtraDesignWidth bool

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

	// PrintOnRollPaper controls whether an unlimited-height page is printed
	// on roll paper. Only meaningful when UnlimitedHeight is true.
	// Mirrors FastReport.ReportPage.PrintOnRollPaper.
	PrintOnRollPaper bool

	// UnlimitedWidth, when true, allows the page to grow horizontally to
	// fit all content. Mirrors FastReport.ReportPage.UnlimitedWidth.
	UnlimitedWidth bool

	// UnlimitedHeightValue is the current rendered height of an unlimited page
	// in report units (pixels at 96 dpi). Set by the engine during report run.
	// Mirrors FastReport.ReportPage.UnlimitedHeightValue.
	UnlimitedHeightValue float32

	// UnlimitedWidthValue is the current rendered width of an unlimited page
	// in report units. Set by the engine during report run.
	// Mirrors FastReport.ReportPage.UnlimitedWidthValue.
	UnlimitedWidthValue float32

	// Paper-source / printer-tray selectors.
	// Default value 7 matches System.Drawing.Printing.PaperSourceKind.AutomaticFeed.
	// These are only relevant when printing to a physical printer.
	// Mirrors FastReport.ReportPage.FirstPageSource / OtherPagesSource / LastPageSource.
	FirstPageSource  int // default 7
	OtherPagesSource int // default 7
	LastPageSource   int // default 7

	// Duplex specifies the printer duplex mode for this page.
	// Values mirror System.Drawing.Printing.Duplex: "Default", "Simplex",
	// "Vertical" (long-edge), "Horizontal" (short-edge).
	// Empty string is treated as "Default" at print time.
	// Mirrors FastReport.ReportPage.Duplex.
	Duplex string

	// Watermark is the optional page watermark (text or image).
	Watermark *Watermark

	// PrintableExpression is an expression that controls whether this page
	// template is included in the report output. Evaluated by the engine;
	// an empty string means always printable.
	// Mirrors C# PageBase.PrintableExpression (inherited from ComponentBase).
	PrintableExpression string

	// VisibleExpression is an expression that, when set, overrides the static
	// Visible flag for this page. Evaluated by the engine before running the page.
	// Mirrors C# ComponentBase.VisibleExpression (ReportEngine.Pages.cs lines 81-84).
	VisibleExpression string

	// inherited marks this page as coming from a base (parent) report.
	inherited bool
}

// TypeName returns the FRX element name used during serialization.
func (*ReportPage) TypeName() string { return "ReportPage" }

// NewReportPage creates a ReportPage with A4 defaults.
// Default paper-source values (7) mirror System.Drawing.Printing.PaperSourceKind.AutomaticFeed,
// matching FastReport's ReportPage constructor defaults.
func NewReportPage() *ReportPage {
	p := &ReportPage{
		BaseObject:        *report.NewBaseObject(),
		visible:           true,
		PaperWidth:        210,
		PaperHeight:       297,
		LeftMargin:        10,
		TopMargin:         10,
		RightMargin:       10,
		BottomMargin:      10,
		FirstPageSource:   7,
		OtherPagesSource:  7,
		LastPageSource:    7,
		TitleBeforeHeader: true, // C# default: [DefaultValue(true)]
	}
	p.fill = &style.SolidFill{Color: style.ColorWhite}
	p.Watermark = NewWatermark()
	return p
}

// GetPaperWidth returns the page paper width in millimetres.
// Satisfies export.ReportPageDims.
func (p *ReportPage) GetPaperWidth() float32 { return p.PaperWidth }

// GetPaperHeight returns the page paper height in millimetres.
// Satisfies export.ReportPageDims.
func (p *ReportPage) GetPaperHeight() float32 { return p.PaperHeight }

// IsUnlimitedHeight reports whether the page has unlimited (dynamic) height.
// Satisfies export.ReportPageDims.
func (p *ReportPage) IsUnlimitedHeight() bool { return p.UnlimitedHeight }

// HeightInPixels returns the current page height in report pixels (96 dpi).
// When UnlimitedHeight is true, returns UnlimitedHeightValue (set by the engine
// during Run); otherwise returns PaperHeight * Units.Millimeters.
// Mirrors FastReport.ReportPage.HeightInPixels (ReportPage.cs:374-379).
func (p *ReportPage) HeightInPixels() float32 {
	if p.UnlimitedHeight {
		return p.UnlimitedHeightValue
	}
	return p.PaperHeight * units.Millimeters
}

// WidthInPixels returns the current page width in report pixels (96 dpi).
// When UnlimitedWidth is true and the engine has set UnlimitedWidthValue,
// returns UnlimitedWidthValue; otherwise returns PaperWidth * Units.Millimeters.
// Mirrors FastReport.ReportPage.WidthInPixels (ReportPage.cs:385-398).
// Note: C# checks IsDesigning to gate UnlimitedWidthValue; Go always uses the
// runtime value when it is non-zero, which matches the headless engine use case.
func (p *ReportPage) WidthInPixels() float32 {
	if p.UnlimitedWidth && p.UnlimitedWidthValue != 0 {
		return p.UnlimitedWidthValue
	}
	return p.PaperWidth * units.Millimeters
}

// Visible returns whether this page template is included in report output.
// Pages with Visible=false are skipped by the engine (e.g. drill-down detail pages).
func (p *ReportPage) Visible() bool { return p.visible }

// SetVisible controls whether this page template is included in report output.
func (p *ReportPage) SetVisible(v bool) { p.visible = v }

// --- PageBase methods (FastReport.Base/PageBase.cs) ---

// PageName returns the display name used in the preview navigator.
// When an explicit override has been set via SetPageName, that value is returned;
// otherwise Name() is used as a fallback.
// Mirrors FastReport.PageBase.PageName (internal getter).
func (p *ReportPage) PageName() string {
	if p.pageName != "" {
		return p.pageName
	}
	return p.Name()
}

// SetPageName sets an explicit display-name override for the preview navigator.
// Pass an empty string to revert to using Name().
// Mirrors FastReport.PageBase.PageName (internal setter).
func (p *ReportPage) SetPageName(name string) { p.pageName = name }

// NeedRefresh reports whether Refresh() or Modify() has been called since the
// last repaint. The preview controller is responsible for clearing this flag.
// Mirrors FastReport.PageBase.NeedRefresh (internal).
func (p *ReportPage) NeedRefresh() bool { return p.needRefresh }

// NeedModify reports whether Modify() has been called, indicating that page
// content has changed and the change must be saved to the prepared report.
// The caller is responsible for clearing this flag after processing.
// Mirrors FastReport.PageBase.NeedModify (internal).
func (p *ReportPage) NeedModify() bool { return p.needModify }

// Refresh signals the preview window to repaint this page.
// It sets NeedRefresh to true without marking the content as changed.
// Call this from interactive-object event handlers (MouseMove, MouseEnter, etc.)
// when only a visual refresh is needed, not a content update.
// Mirrors FastReport.PageBase.Refresh().
func (p *ReportPage) Refresh() { p.needRefresh = true }

// Modify signals that page content has changed and refreshes the preview window.
// It sets both NeedModify and NeedRefresh to true.
// Call this from interactive-object event handlers (Click, MouseDown, MouseUp)
// when you want to change an object and have the preview reflect the change.
// Mirrors FastReport.PageBase.Modify().
func (p *ReportPage) Modify() {
	p.needModify = true
	p.needRefresh = true
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

// AddBand appends a band to the page and sets the band's parent to this page.
// Mirrors C# BandCollection behaviour: adding a band transfers ownership to the page.
func (p *ReportPage) AddBand(b report.Base) {
	p.bands = append(p.bands, b)
	b.SetParent(p)
}

// AddBandByTypeName routes a deserialized band to the appropriate page slot
// based on its FRX type name. Singleton bands (PageHeader, PageFooter, etc.)
// are placed in their dedicated fields; all other bands go into the ordered list.
// Both short names ("PageHeader") and full FastReport names ("PageHeaderBand") are accepted.
// The band's parent is always set to this page.
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
	b.SetParent(p)
}

// --- report.Parent implementation ---

// CanContain returns true when this page can accept child as a direct child.
// A ReportPage can contain any band (report.Base that is a band type).
// Mirrors C# ReportPage.CanContain (ReportPage.cs).
func (p *ReportPage) CanContain(child report.Base) bool {
	// Pages accept any band as a child.
	switch child.(type) {
	case *band.DataBand, *band.GroupHeaderBand, *band.GroupFooterBand,
		*band.PageHeaderBand, *band.PageFooterBand,
		*band.ReportTitleBand, *band.ReportSummaryBand,
		*band.ColumnHeaderBand, *band.ColumnFooterBand,
		*band.OverlayBand, *band.ChildBand:
		return true
	}
	return false
}

// GetChildObjects fills list with all bands on the page (singleton + ordered).
func (p *ReportPage) GetChildObjects(list *[]report.Base) {
	*list = append(*list, p.AllBands()...)
}

// AddChild adds child to the page, routing to the appropriate slot.
// Equivalent to AddBandByTypeName using the child's type name.
// Mirrors C# BandCollection.Add (BandCollection.cs).
func (p *ReportPage) AddChild(child report.Base) {
	p.AddBandByTypeName(child.BaseName(), child)
}

// RemoveChild removes child from the page, clearing the parent reference.
// Mirrors C# BandCollection.Remove (BandCollection.cs).
func (p *ReportPage) RemoveChild(child report.Base) {
	// Clear singleton slots.
	if p.pageHeader == child {
		p.pageHeader = nil
		child.SetParent(nil)
		return
	}
	if p.pageFooter == child {
		p.pageFooter = nil
		child.SetParent(nil)
		return
	}
	if p.reportTitle == child {
		p.reportTitle = nil
		child.SetParent(nil)
		return
	}
	if p.reportSummary == child {
		p.reportSummary = nil
		child.SetParent(nil)
		return
	}
	if p.columnHeader == child {
		p.columnHeader = nil
		child.SetParent(nil)
		return
	}
	if p.columnFooter == child {
		p.columnFooter = nil
		child.SetParent(nil)
		return
	}
	if p.overlay == child {
		p.overlay = nil
		child.SetParent(nil)
		return
	}
	// Remove from ordered bands slice.
	for i, b := range p.bands {
		if b == child {
			p.bands = append(p.bands[:i], p.bands[i+1:]...)
			child.SetParent(nil)
			return
		}
	}
}

// GetChildOrder returns the index of child in the ordered bands slice, or -1.
func (p *ReportPage) GetChildOrder(child report.Base) int {
	for i, b := range p.bands {
		if b == child {
			return i
		}
	}
	return -1
}

// SetChildOrder moves child to the specified position in the ordered bands slice.
func (p *ReportPage) SetChildOrder(child report.Base, order int) {
	idx := p.GetChildOrder(child)
	if idx < 0 {
		return
	}
	// Remove from current position.
	p.bands = append(p.bands[:idx], p.bands[idx+1:]...)
	// Clamp target.
	if order < 0 {
		order = 0
	}
	if order > len(p.bands) {
		order = len(p.bands)
	}
	// Insert at target position.
	p.bands = append(p.bands, nil)
	copy(p.bands[order+1:], p.bands[order:])
	p.bands[order] = child
}

// UpdateLayout is a no-op for ReportPage; the engine manages band layout.
func (p *ReportPage) UpdateLayout(dx, dy float32) {}

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
	if !p.visible {
		w.WriteBool("Visible", false)
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
	if !p.TitleBeforeHeader {
		w.WriteBool("TitleBeforeHeader", false) // only write when non-default (C# default is true)
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
	if p.RawPaperSize != 0 {
		w.WriteInt("RawPaperSize", p.RawPaperSize)
	}
	if p.ExportAlias != "" {
		w.WriteStr("ExportAlias", p.ExportAlias)
	}
	if p.ExtraDesignWidth {
		w.WriteBool("ExtraDesignWidth", true)
	}
	if p.UnlimitedHeight {
		w.WriteBool("UnlimitedHeight", true)
	}
	if p.PrintOnRollPaper {
		w.WriteBool("PrintOnRollPaper", true)
	}
	if p.UnlimitedWidth {
		w.WriteBool("UnlimitedWidth", true)
	}
	if p.UnlimitedHeightValue != 0 {
		w.WriteFloat("UnlimitedHeightValue", p.UnlimitedHeightValue)
	}
	if p.UnlimitedWidthValue != 0 {
		w.WriteFloat("UnlimitedWidthValue", p.UnlimitedWidthValue)
	}
	if p.FirstPageSource != 7 {
		w.WriteFloat("FirstPageSource", float32(p.FirstPageSource))
	}
	// C# attribute name is "OtherPageSource" (singular) — FastReport.Base/ReportPage.cs:1198
	if p.OtherPagesSource != 7 {
		w.WriteFloat("OtherPageSource", float32(p.OtherPagesSource))
	}
	if p.LastPageSource != 7 {
		w.WriteFloat("LastPageSource", float32(p.LastPageSource))
	}
	if p.Duplex != "" && p.Duplex != "Default" {
		w.WriteStr("Duplex", p.Duplex)
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
	if p.PrintableExpression != "" {
		w.WriteStr("PrintableExpression", p.PrintableExpression)
	}
	if p.VisibleExpression != "" {
		w.WriteStr("VisibleExpression", p.VisibleExpression)
	}

	// Page columns — mirror PageColumns.Serialize (FastReport.Base/PageColumns.cs:101-111).
	// Count != 1 (C# default) triggers column serialization; Go default is 0 (single column).
	if p.Columns.Count > 1 {
		w.WriteInt("Columns.Count", p.Columns.Count)
		w.WriteFloat("Columns.Width", p.Columns.Width)
		if len(p.Columns.Positions) > 0 {
			parts := make([]string, len(p.Columns.Positions))
			for i, v := range p.Columns.Positions {
				parts[i] = fmt.Sprintf("%g", v)
			}
			w.WriteStr("Columns.Positions", strings.Join(parts, ","))
		}
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
	p.visible = r.ReadBool("Visible", true)
	p.PaperWidth = r.ReadFloat("PaperWidth", 210)
	p.PaperHeight = r.ReadFloat("PaperHeight", 297)
	p.Landscape = r.ReadBool("Landscape", false)
	p.RawPaperSize = r.ReadInt("RawPaperSize", 0)
	p.ExportAlias = r.ReadStr("ExportAlias", "")
	p.LeftMargin = r.ReadFloat("LeftMargin", 10)
	p.TopMargin = r.ReadFloat("TopMargin", 10)
	p.RightMargin = r.ReadFloat("RightMargin", 10)
	p.BottomMargin = r.ReadFloat("BottomMargin", 10)
	p.MirrorMargins = r.ReadBool("MirrorMargins", false)
	p.TitleBeforeHeader = r.ReadBool("TitleBeforeHeader", true) // C# default is true
	p.PrintOnPreviousPage = r.ReadBool("PrintOnPreviousPage", false)
	p.ResetPageNumber = r.ReadBool("ResetPageNumber", false)
	p.StartOnOddPage = r.ReadBool("StartOnOddPage", false)
	p.ExtraDesignWidth = r.ReadBool("ExtraDesignWidth", false)
	p.UnlimitedHeight = r.ReadBool("UnlimitedHeight", false)
	p.PrintOnRollPaper = r.ReadBool("PrintOnRollPaper", false)
	p.UnlimitedWidth = r.ReadBool("UnlimitedWidth", false)
	p.UnlimitedHeightValue = r.ReadFloat("UnlimitedHeightValue", 0)
	p.UnlimitedWidthValue = r.ReadFloat("UnlimitedWidthValue", 0)
	p.FirstPageSource = r.ReadInt("FirstPageSource", 7)
	// C# uses "OtherPageSource" (singular); fall back to "OtherPagesSource" for older Go-generated files.
	p.OtherPagesSource = r.ReadInt("OtherPageSource", -1)
	if p.OtherPagesSource == -1 {
		p.OtherPagesSource = r.ReadInt("OtherPagesSource", 7)
	}
	p.LastPageSource = r.ReadInt("LastPageSource", 7)
	p.Duplex = r.ReadStr("Duplex", "")
	p.OutlineExpression = r.ReadStr("OutlineExpression", "")
	p.CreatePageEvent = r.ReadStr("CreatePageEvent", "")
	p.StartPageEvent = r.ReadStr("StartPageEvent", "")
	p.FinishPageEvent = r.ReadStr("FinishPageEvent", "")
	p.ManualBuildEvent = r.ReadStr("ManualBuildEvent", "")
	p.BackPage = r.ReadStr("BackPage", "")
	p.BackPageOddEven = r.ReadInt("BackPageOddEven", 0)
	p.PrintableExpression = r.ReadStr("PrintableExpression", "")
	p.VisibleExpression = r.ReadStr("VisibleExpression", "")
	p.Columns.Count = r.ReadInt("Columns.Count", 0)
	p.Columns.Width = r.ReadFloat("Columns.Width", 0)
	if posStr := r.ReadStr("Columns.Positions", ""); posStr != "" {
		p.Columns.Positions = parseFloatList(posStr)
	}
	if p.Watermark == nil {
		p.Watermark = NewWatermark()
	}
	p.Watermark.Deserialize(r)
	return nil
}

// GetExpressions returns all expressions used by this page.
// Includes the OutlineExpression when set.
// Mirrors C# ReportPage.GetExpressions (ReportPage.cs line 1409–1418).
func (p *ReportPage) GetExpressions() []string {
	var exprs []string
	if p.OutlineExpression != "" {
		exprs = append(exprs, p.OutlineExpression)
	}
	return exprs
}

// Assign copies all ReportPage properties from src (excluding bands and children).
// Mirrors C# ReportPage.Assign (ReportPage.cs line 1080–1099).
func (p *ReportPage) Assign(src *ReportPage) {
	if src == nil {
		return
	}
	p.BaseObject.SetName(src.Name())
	p.ExportAlias = src.ExportAlias
	p.Landscape = src.Landscape
	p.PaperWidth = src.PaperWidth
	p.PaperHeight = src.PaperHeight
	p.RawPaperSize = src.RawPaperSize
	p.LeftMargin = src.LeftMargin
	p.TopMargin = src.TopMargin
	p.RightMargin = src.RightMargin
	p.BottomMargin = src.BottomMargin
	p.MirrorMargins = src.MirrorMargins
	p.FirstPageSource = src.FirstPageSource
	p.OtherPagesSource = src.OtherPagesSource
	p.LastPageSource = src.LastPageSource
	p.Duplex = src.Duplex
	p.Columns.Assign(src.Columns)
	p.TitleBeforeHeader = src.TitleBeforeHeader
	p.OutlineExpression = src.OutlineExpression
	p.PrintOnPreviousPage = src.PrintOnPreviousPage
	p.ResetPageNumber = src.ResetPageNumber
	p.ExtraDesignWidth = src.ExtraDesignWidth
	p.StartOnOddPage = src.StartOnOddPage
	p.BackPage = src.BackPage
	p.BackPageOddEven = src.BackPageOddEven
	p.UnlimitedHeight = src.UnlimitedHeight
	p.PrintOnRollPaper = src.PrintOnRollPaper
	p.UnlimitedWidth = src.UnlimitedWidth
	p.border = src.border
	p.fill = src.fill
	if src.Watermark != nil {
		if p.Watermark == nil {
			p.Watermark = NewWatermark()
		}
		*p.Watermark = *src.Watermark
	}
	p.CreatePageEvent = src.CreatePageEvent
	p.StartPageEvent = src.StartPageEvent
	p.FinishPageEvent = src.FinishPageEvent
	p.ManualBuildEvent = src.ManualBuildEvent
	p.PrintableExpression = src.PrintableExpression
	p.VisibleExpression = src.VisibleExpression
}

// SetColumnsCount sets the number of columns and auto-calculates Width and Positions
// from the current page margins. Count must be >= 1.
// Mirrors C# PageColumns.Count setter (PageColumns.cs line 27–42).
func (p *ReportPage) SetColumnsCount(count int) {
	if count <= 0 {
		count = 1
	}
	p.Columns.Count = count
	usableWidth := p.PaperWidth - p.LeftMargin - p.RightMargin
	if count > 0 {
		p.Columns.Width = usableWidth / float32(count)
	}
	p.Columns.Positions = make([]float32, count)
	for i := 0; i < count; i++ {
		p.Columns.Positions[i] = float32(i) * p.Columns.Width
	}
}
