package report

import (
	"image/color"

	"github.com/andrewloable/go-fastreport/style"
)

// ShiftMode controls how a component shifts when overlapping components grow.
type ShiftMode int

const (
	// ShiftNever means the component never shifts.
	ShiftNever ShiftMode = iota
	// ShiftAlways means the component always shifts down with the previous band.
	ShiftAlways
	// ShiftWhenOverlapped means the component shifts only when overlapped.
	ShiftWhenOverlapped
)

// PrintOn controls which pages a component is printed on.
type PrintOn int

const (
	// PrintOnAllPages prints on every page (default).
	PrintOnAllPages PrintOn = 0
	// PrintOnFirstPage prints only on the first page.
	PrintOnFirstPage PrintOn = 1
	// PrintOnLastPage prints only on the last page.
	PrintOnLastPage PrintOn = 2
	// PrintOnOddPages prints on odd-numbered pages.
	PrintOnOddPages PrintOn = 4
	// PrintOnEvenPages prints on even-numbered pages.
	PrintOnEvenPages PrintOn = 8
	// PrintOnRepeatedBand prints when the band is repeated (e.g. group header).
	PrintOnRepeatedBand PrintOn = 16
	// PrintOnSinglePage prints only when the report fits on a single page.
	PrintOnSinglePage PrintOn = 32
)

// Hyperlink holds hyperlink properties for a report component.
type Hyperlink struct {
	// Expression is the expression that evaluates to the URL.
	Expression string
	// Kind is the hyperlink kind (e.g. "URL", "Bookmark", "DetailReport").
	Kind string
	// Target is the hyperlink target (e.g. "_blank").
	Target string
}

// EventArgs holds context passed to report event callbacks.
type EventArgs struct{}

// EventHandler is the type for report lifecycle event callbacks.
type EventHandler func(sender Base, e *EventArgs)

// ReportComponentBase extends ComponentBase with visual styling, print control,
// and event callbacks. It is the Go equivalent of FastReport.ReportComponentBase.
type ReportComponentBase struct {
	ComponentBase

	// Visual styling.
	border style.Border
	fill   style.Fill

	// Style references.
	styleName      string
	evenStyleName  string
	hoverStyleName string

	// Export control.
	exportable           bool
	exportableExpression string

	// Growth / shrink control.
	canGrow      bool
	canShrink    bool
	growToBottom bool
	shiftMode    ShiftMode

	// Page print control.
	printOn   PrintOn
	pageBreak bool

	// Navigation.
	bookmark  string
	hyperlink *Hyperlink

	// Event callbacks.
	OnBeforePrint EventHandler
	OnAfterPrint  EventHandler
	OnAfterData   EventHandler
	OnClick       EventHandler
}

// NewReportComponentBase creates a ReportComponentBase with defaults:
// exportable=true, PrintOn=PrintOnAllPages, solid white fill.
func NewReportComponentBase() *ReportComponentBase {
	rc := &ReportComponentBase{
		ComponentBase: *NewComponentBase(),
		exportable:    true,
		printOn:       PrintOnAllPages,
		fill:          &style.SolidFill{Color: color.RGBA{R: 255, G: 255, B: 255, A: 255}},
	}
	return rc
}

// --- Border ---

// Border returns the component's border.
func (rc *ReportComponentBase) Border() style.Border { return rc.border }

// SetBorder sets the component's border.
func (rc *ReportComponentBase) SetBorder(b style.Border) { rc.border = b }

// --- Fill ---

// Fill returns the component's fill.
func (rc *ReportComponentBase) Fill() style.Fill { return rc.fill }

// SetFill sets the component's fill.
func (rc *ReportComponentBase) SetFill(f style.Fill) { rc.fill = f }

// --- Styles ---

// StyleName returns the style name applied to the component.
func (rc *ReportComponentBase) StyleName() string { return rc.styleName }

// SetStyleName sets the style name.
func (rc *ReportComponentBase) SetStyleName(s string) { rc.styleName = s }

// EvenStyleName returns the style applied to alternating (even) rows.
func (rc *ReportComponentBase) EvenStyleName() string { return rc.evenStyleName }

// SetEvenStyleName sets the even-row style name.
func (rc *ReportComponentBase) SetEvenStyleName(s string) { rc.evenStyleName = s }

// HoverStyleName returns the style applied on hover (web export).
func (rc *ReportComponentBase) HoverStyleName() string { return rc.hoverStyleName }

// SetHoverStyleName sets the hover style name.
func (rc *ReportComponentBase) SetHoverStyleName(s string) { rc.hoverStyleName = s }

// --- Export ---

// Exportable returns whether the component is included in exports.
func (rc *ReportComponentBase) Exportable() bool { return rc.exportable }

// SetExportable sets the exportable flag.
func (rc *ReportComponentBase) SetExportable(v bool) { rc.exportable = v }

// ExportableExpression returns the expression controlling exportability.
func (rc *ReportComponentBase) ExportableExpression() string { return rc.exportableExpression }

// SetExportableExpression sets the exportable expression.
func (rc *ReportComponentBase) SetExportableExpression(expr string) {
	rc.exportableExpression = expr
}

// --- Growth / Shrink ---

// CanGrow returns whether the component can grow to fit its content.
func (rc *ReportComponentBase) CanGrow() bool { return rc.canGrow }

// SetCanGrow sets the canGrow flag.
func (rc *ReportComponentBase) SetCanGrow(v bool) { rc.canGrow = v }

// CanShrink returns whether the component can shrink when content is small.
func (rc *ReportComponentBase) CanShrink() bool { return rc.canShrink }

// SetCanShrink sets the canShrink flag.
func (rc *ReportComponentBase) SetCanShrink(v bool) { rc.canShrink = v }

// GrowToBottom returns whether the component grows to the bottom of the band.
func (rc *ReportComponentBase) GrowToBottom() bool { return rc.growToBottom }

// SetGrowToBottom sets the growToBottom flag.
func (rc *ReportComponentBase) SetGrowToBottom(v bool) { rc.growToBottom = v }

// ShiftMode returns the shift mode for overlapping components.
func (rc *ReportComponentBase) ShiftMode() ShiftMode { return rc.shiftMode }

// SetShiftMode sets the shift mode.
func (rc *ReportComponentBase) SetShiftMode(m ShiftMode) { rc.shiftMode = m }

// --- Page print control ---

// PrintOn returns the pages this component is printed on.
func (rc *ReportComponentBase) PrintOn() PrintOn { return rc.printOn }

// SetPrintOn sets the print page mask.
func (rc *ReportComponentBase) SetPrintOn(p PrintOn) { rc.printOn = p }

// PageBreak returns whether a page break occurs before this component.
func (rc *ReportComponentBase) PageBreak() bool { return rc.pageBreak }

// SetPageBreak sets the page break flag.
func (rc *ReportComponentBase) SetPageBreak(v bool) { rc.pageBreak = v }

// --- Navigation ---

// Bookmark returns the bookmark name for this component.
func (rc *ReportComponentBase) Bookmark() string { return rc.bookmark }

// SetBookmark sets the bookmark name.
func (rc *ReportComponentBase) SetBookmark(s string) { rc.bookmark = s }

// Hyperlink returns the hyperlink, or nil if not set.
func (rc *ReportComponentBase) Hyperlink() *Hyperlink { return rc.hyperlink }

// SetHyperlink sets the hyperlink.
func (rc *ReportComponentBase) SetHyperlink(h *Hyperlink) { rc.hyperlink = h }

// --- Events ---

// FireBeforePrint invokes OnBeforePrint if set.
func (rc *ReportComponentBase) FireBeforePrint() {
	if rc.OnBeforePrint != nil {
		rc.OnBeforePrint(rc, &EventArgs{})
	}
}

// FireAfterPrint invokes OnAfterPrint if set.
func (rc *ReportComponentBase) FireAfterPrint() {
	if rc.OnAfterPrint != nil {
		rc.OnAfterPrint(rc, &EventArgs{})
	}
}

// FireAfterData invokes OnAfterData if set.
func (rc *ReportComponentBase) FireAfterData() {
	if rc.OnAfterData != nil {
		rc.OnAfterData(rc, &EventArgs{})
	}
}

// FireClick invokes OnClick if set.
func (rc *ReportComponentBase) FireClick() {
	if rc.OnClick != nil {
		rc.OnClick(rc, &EventArgs{})
	}
}

// --- Serialization ---

// Serialize writes ReportComponentBase properties that differ from defaults.
func (rc *ReportComponentBase) Serialize(w Writer) error {
	if err := rc.ComponentBase.Serialize(w); err != nil {
		return err
	}
	if !rc.exportable {
		w.WriteBool("Exportable", false)
	}
	if rc.exportableExpression != "" {
		w.WriteStr("ExportableExpression", rc.exportableExpression)
	}
	if rc.canGrow {
		w.WriteBool("CanGrow", true)
	}
	if rc.canShrink {
		w.WriteBool("CanShrink", true)
	}
	if rc.growToBottom {
		w.WriteBool("GrowToBottom", true)
	}
	if rc.shiftMode != ShiftNever {
		w.WriteInt("ShiftMode", int(rc.shiftMode))
	}
	if rc.printOn != PrintOnAllPages {
		w.WriteInt("PrintOn", int(rc.printOn))
	}
	if rc.pageBreak {
		w.WriteBool("PageBreak", true)
	}
	if rc.styleName != "" {
		w.WriteStr("Style", rc.styleName)
	}
	if rc.evenStyleName != "" {
		w.WriteStr("EvenStyle", rc.evenStyleName)
	}
	if rc.hoverStyleName != "" {
		w.WriteStr("HoverStyle", rc.hoverStyleName)
	}
	if rc.bookmark != "" {
		w.WriteStr("Bookmark", rc.bookmark)
	}
	return nil
}

// Deserialize reads ReportComponentBase properties.
func (rc *ReportComponentBase) Deserialize(r Reader) error {
	if err := rc.ComponentBase.Deserialize(r); err != nil {
		return err
	}
	rc.exportable = r.ReadBool("Exportable", true)
	rc.exportableExpression = r.ReadStr("ExportableExpression", "")
	rc.canGrow = r.ReadBool("CanGrow", false)
	rc.canShrink = r.ReadBool("CanShrink", false)
	rc.growToBottom = r.ReadBool("GrowToBottom", false)
	rc.shiftMode = ShiftMode(r.ReadInt("ShiftMode", int(ShiftNever)))
	rc.printOn = PrintOn(r.ReadInt("PrintOn", int(PrintOnAllPages)))
	rc.pageBreak = r.ReadBool("PageBreak", false)
	rc.styleName = r.ReadStr("Style", "")
	rc.evenStyleName = r.ReadStr("EvenStyle", "")
	rc.hoverStyleName = r.ReadStr("HoverStyle", "")
	rc.bookmark = r.ReadStr("Bookmark", "")
	return nil
}
