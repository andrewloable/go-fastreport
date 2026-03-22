package report

import (
	"fmt"
	"strings"

	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
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

// StylePriority controls which properties of an even-row style are applied.
// It is the Go equivalent of FastReport.StylePriority.
type StylePriority int

const (
	// StylePriorityUseFill applies only the fill from the even style.
	// This is the default, matching FastReport C# default.
	StylePriorityUseFill StylePriority = iota
	// StylePriorityUseAll applies all style properties (border, fill, font,
	// text fill) from the even style.
	StylePriorityUseAll
)

// StyleLookup is a minimal interface that lets ReportComponentBase look up a
// named StyleEntry without importing the reportpkg package (which would create
// a cycle). The *reportpkg.Report struct satisfies this interface by delegating
// to its embedded *style.StyleSheet.
type StyleLookup interface {
	// FindStyle returns the StyleEntry registered under the given name, or nil.
	FindStyle(name string) *style.StyleEntry
}

// Hyperlink holds hyperlink properties for a report component.
type Hyperlink struct {
	// Kind is the hyperlink kind (e.g. "URL", "Bookmark", "DetailReport", "DetailPage").
	Kind string
	// Expression is the expression that evaluates to the URL or anchor value.
	Expression string
	// Value is a static URL or anchor value (used when Expression is empty).
	Value string
	// Target is the hyperlink target frame (e.g. "_blank", "_self").
	Target string
	// DetailPageName is the name of the detail report page (for DetailPage kind).
	DetailPageName string
	// DetailReportName is the name of the detail report (for DetailReport kind).
	DetailReportName string
	// ReportParameter is the parameter name to pass to the detail report.
	ReportParameter string
	// OpenLinkInNewTab controls whether the link opens in a new browser tab.
	// Used by HTML export only. Mirrors C# Hyperlink.OpenLinkInNewTab (Hyperlink.cs).
	OpenLinkInNewTab bool
	// ValuesSeparator is the separator string for multi-value parameters.
	// Default is ";" to match C# Hyperlink constructor default (Hyperlink.cs line 352).
	ValuesSeparator string
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
	styleName         string
	evenStyleName     string
	hoverStyleName    string
	evenStylePriority StylePriority

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

	// savedState holds the pre-print snapshot used by SaveState/RestoreState.
	// The engine calls SaveState before processing an object and RestoreState
	// after, so that style application and expression evaluation do not
	// permanently mutate design-time properties.
	// Mirrors C# ReportComponentBase private saved* fields (savedBounds,
	// savedVisible, savedBookmark, savedBorder, savedFill).
	savedState *savedComponentState
}

// savedComponentState holds a snapshot of the mutable properties saved by
// SaveState. This mirrors the C# ReportComponentBase private saved* fields.
type savedComponentState struct {
	bounds   Rect
	visible  bool
	bookmark string
	border   style.Border
	fill     style.Fill
}

// NewReportComponentBase creates a ReportComponentBase with defaults:
// exportable=true, PrintOn=PrintOnAllPages, solid transparent fill,
// evenStylePriority=StylePriorityUseFill.
// Matches C# ReportComponentBase constructor defaults.
func NewReportComponentBase() *ReportComponentBase {
	rc := &ReportComponentBase{
		ComponentBase:     *NewComponentBase(),
		border:            *style.NewBorder(),
		exportable:        true,
		printOn:           PrintOnAllPages,
		fill:              style.NewSolidFill(style.ColorTransparent), // C# default: Color.Transparent
		evenStylePriority: StylePriorityUseFill,                       // C# default: StylePriority.UseFill
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

// ApplyStyle applies the visual overrides from a style.StyleEntry to the
// component. Both the modern Apply* flags and the legacy *Changed flags are
// honoured; a property is applied when either the new flag or its legacy
// equivalent is true.
//
// The fill is applied via entry.EffectiveFill() so that gradient and hatch fills
// stored in entry.Fill are used when present, falling back to entry.FillColor
// for the common solid-fill case. This matches C# StyleBase.Fill as FillBase.
//
// Subclasses (e.g. TextObject) should call this method first and then apply
// their own font/text-color overrides, since ReportComponentBase does not hold
// a font field directly.
//
// This is called by style.StyleSheet.ApplyToObject before rendering.
func (rc *ReportComponentBase) ApplyStyle(entry *style.StyleEntry) {
	if entry == nil {
		return
	}
	// Fill: use EffectiveFill to support gradient/hatch fills in addition to solid.
	if f := entry.EffectiveFill(); f != nil {
		rc.fill = f
	}
	// Border: apply when ApplyBorder or the legacy BorderColorChanged is true.
	// When the entry carries a full Border (lines[0] non-nil), replace the
	// component's border entirely. Otherwise only update the colour on existing
	// lines, matching C# ApplyStyle which calls Border = style.Border.Clone().
	if entry.ApplyBorder || entry.BorderColorChanged {
		if entry.Border.Lines[0] != nil {
			// Full Border override from FRX-deserialized style.
			cloned := *style.NewBorder()
			for i := range entry.Border.Lines {
				if entry.Border.Lines[i] != nil {
					*cloned.Lines[i] = *entry.Border.Lines[i]
				}
			}
			cloned.VisibleLines = entry.Border.VisibleLines
			cloned.Shadow = entry.Border.Shadow
			rc.border = cloned
		} else {
			// Legacy colour-only override (border colour without line config).
			b := rc.border
			for i := range b.Lines {
				if b.Lines[i] != nil {
					b.Lines[i].Color = entry.BorderColor
				}
			}
			rc.border = b
		}
	}
}

// EvenStyleName returns the style applied to alternating (even) rows.
func (rc *ReportComponentBase) EvenStyleName() string { return rc.evenStyleName }

// SetEvenStyleName sets the even-row style name.
func (rc *ReportComponentBase) SetEvenStyleName(s string) { rc.evenStyleName = s }

// EvenStylePriority returns which properties of the even style are applied.
// Defaults to StylePriorityUseFill, matching C# EvenStylePriority default.
func (rc *ReportComponentBase) EvenStylePriority() StylePriority { return rc.evenStylePriority }

// SetEvenStylePriority sets the even-style priority.
func (rc *ReportComponentBase) SetEvenStylePriority(p StylePriority) { rc.evenStylePriority = p }

// ApplyEvenStyle looks up EvenStyleName in ss and applies it to this component.
// When EvenStylePriority is StylePriorityUseFill only the fill is applied;
// when StylePriorityUseAll the full style is applied.
//
// This mirrors C# ReportComponentBase.ApplyEvenStyle()
// (FastReport.Base/ReportComponentBase.cs line 734–748).
// The engine should call this on every other data row for banded DataBands.
func (rc *ReportComponentBase) ApplyEvenStyle(ss StyleLookup) {
	if rc.evenStyleName == "" || ss == nil {
		return
	}
	entry := ss.FindStyle(rc.evenStyleName)
	if entry == nil {
		return
	}
	if rc.evenStylePriority == StylePriorityUseFill {
		// Apply only fill — mirrors: Fill = style.Fill.Clone()
		if f := entry.EffectiveFill(); f != nil {
			rc.fill = f.Clone()
		}
	} else {
		rc.ApplyStyle(entry)
	}
}

// HoverStyleName returns the style applied on hover (web export).
func (rc *ReportComponentBase) HoverStyleName() string { return rc.hoverStyleName }

// SetHoverStyleName sets the hover style name.
func (rc *ReportComponentBase) SetHoverStyleName(s string) { rc.hoverStyleName = s }

// --- Save / Restore State ---

// SaveState saves the current Bounds, Visible, Bookmark, Border, and Fill so
// they can be restored after the engine has applied dynamic overrides.
// This mirrors C# ReportComponentBase.SaveState()
// (FastReport.Base/ReportComponentBase.cs line 957–965).
func (rc *ReportComponentBase) SaveState() {
	var savedFill style.Fill
	if rc.fill != nil {
		savedFill = rc.fill.Clone()
	}
	rc.savedState = &savedComponentState{
		bounds:   rc.ComponentBase.Bounds(),
		visible:  rc.ComponentBase.Visible(),
		bookmark: rc.bookmark,
		border:   rc.border,
		fill:     savedFill,
	}
}

// RestoreState restores Bounds, Visible, Bookmark, Border, and Fill to the
// values saved by the last SaveState call. It is a no-op when SaveState has not
// been called.
// This mirrors C# ReportComponentBase.RestoreState()
// (FastReport.Base/ReportComponentBase.cs line 975–983).
func (rc *ReportComponentBase) RestoreState() {
	if rc.savedState == nil {
		return
	}
	rc.ComponentBase.SetBounds(rc.savedState.bounds)
	rc.ComponentBase.SetVisible(rc.savedState.visible)
	rc.bookmark = rc.savedState.bookmark
	rc.border = rc.savedState.border
	rc.fill = rc.savedState.fill
	rc.savedState = nil
}

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

// Assign copies all ReportComponentBase fields from src into this component.
// Mirrors C# ReportComponentBase.Assign (ReportComponentBase.cs line 678-709).
// Note: event handler fields (OnBeforePrint, OnAfterPrint, etc.) are not copied
// because they are runtime callbacks, not design-time properties.
func (rc *ReportComponentBase) Assign(src *ReportComponentBase) {
	if src == nil {
		return
	}
	rc.ComponentBase = src.ComponentBase
	rc.border = src.border
	if src.fill != nil {
		rc.fill = src.fill.Clone()
	} else {
		rc.fill = nil
	}
	rc.styleName = src.styleName
	rc.evenStyleName = src.evenStyleName
	rc.hoverStyleName = src.hoverStyleName
	rc.evenStylePriority = src.evenStylePriority
	rc.exportable = src.exportable
	rc.exportableExpression = src.exportableExpression
	rc.canGrow = src.canGrow
	rc.canShrink = src.canShrink
	rc.growToBottom = src.growToBottom
	rc.shiftMode = src.shiftMode
	rc.printOn = src.printOn
	rc.pageBreak = src.pageBreak
	rc.bookmark = src.bookmark
	if src.hyperlink != nil {
		h := *src.hyperlink
		rc.hyperlink = &h
	} else {
		rc.hyperlink = nil
	}
}

// --- Serialization ---

// Serialize writes ReportComponentBase properties that differ from defaults.
func (rc *ReportComponentBase) Serialize(w Writer) error {
	if err := rc.ComponentBase.Serialize(w); err != nil {
		return err
	}
	// Border and Fill — delta against FRX defaults.
	serializeBorder(w, &rc.border)
	serializeFill(w, rc.fill)
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
	if rc.evenStylePriority != StylePriorityUseFill {
		w.WriteInt("EvenStylePriority", int(rc.evenStylePriority))
	}
	if rc.hoverStyleName != "" {
		w.WriteStr("HoverStyle", rc.hoverStyleName)
	}
	if rc.bookmark != "" {
		w.WriteStr("Bookmark", rc.bookmark)
	}
	if h := rc.hyperlink; h != nil {
		if h.Kind != "" {
			w.WriteStr("Hyperlink.Kind", h.Kind)
		}
		if h.Expression != "" {
			w.WriteStr("Hyperlink.Expression", h.Expression)
		}
		if h.Value != "" {
			w.WriteStr("Hyperlink.Value", h.Value)
		}
		if h.Target != "" {
			w.WriteStr("Hyperlink.Target", h.Target)
		}
		if h.DetailPageName != "" {
			w.WriteStr("Hyperlink.DetailPageName", h.DetailPageName)
		}
		if h.DetailReportName != "" {
			w.WriteStr("Hyperlink.DetailReportName", h.DetailReportName)
		}
		if h.ReportParameter != "" {
			w.WriteStr("Hyperlink.ReportParameter", h.ReportParameter)
		}
		// ValuesSeparator default is ";" — only write when different.
		// C# ref: Hyperlink.ShouldSerializeValuesSeparator / Hyperlink.cs line 218.
		if h.ValuesSeparator != "" && h.ValuesSeparator != ";" {
			w.WriteStr("Hyperlink.ValuesSeparator", h.ValuesSeparator)
		}
		// OpenLinkInNewTab — only write when true (default false).
		// C# ref: Hyperlink.Serialize (Hyperlink.cs line 318-319).
		if h.OpenLinkInNewTab {
			w.WriteBool("Hyperlink.OpenLinkInNewTab", true)
		}
	}
	return nil
}

// Deserialize reads ReportComponentBase properties.
func (rc *ReportComponentBase) Deserialize(r Reader) error {
	if err := rc.ComponentBase.Deserialize(r); err != nil {
		return err
	}
	// Border and Fill.
	deserializeBorder(r, &rc.border)
	rc.fill = deserializeFill(r, rc.fill)
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
	rc.evenStylePriority = StylePriority(r.ReadInt("EvenStylePriority", int(StylePriorityUseFill)))
	rc.hoverStyleName = r.ReadStr("HoverStyle", "")
	rc.bookmark = r.ReadStr("Bookmark", "")
	// Hyperlink dot-notation attributes.
	hlKind := r.ReadStr("Hyperlink.Kind", "")
	hlExpr := r.ReadStr("Hyperlink.Expression", "")
	hlValue := r.ReadStr("Hyperlink.Value", "")
	hlTarget := r.ReadStr("Hyperlink.Target", "")
	hlDetailPage := r.ReadStr("Hyperlink.DetailPageName", "")
	hlDetailReport := r.ReadStr("Hyperlink.DetailReportName", "")
	hlParam := r.ReadStr("Hyperlink.ReportParameter", "")
	hlSep := r.ReadStr("Hyperlink.ValuesSeparator", "")
	hlNewTab := r.ReadBool("Hyperlink.OpenLinkInNewTab", false)
	if hlKind != "" || hlExpr != "" || hlValue != "" || hlTarget != "" ||
		hlDetailPage != "" || hlDetailReport != "" || hlParam != "" ||
		hlSep != "" || hlNewTab {
		hl := &Hyperlink{
			Kind:             hlKind,
			Expression:       hlExpr,
			Value:            hlValue,
			Target:           hlTarget,
			DetailPageName:   hlDetailPage,
			DetailReportName: hlDetailReport,
			ReportParameter:  hlParam,
			OpenLinkInNewTab: hlNewTab,
		}
		// Apply ValuesSeparator; default in C# is ";" (Hyperlink.cs line 352).
		if hlSep != "" {
			hl.ValuesSeparator = hlSep
		} else {
			hl.ValuesSeparator = ";"
		}
		rc.hyperlink = hl
	}
	return nil
}

// GetExpressions returns the list of expressions used by this
// ReportComponentBase: the base component expressions (VisibleExpression,
// PrintableExpression) plus the hyperlink expression, bookmark, and
// exportable expression.
//
// Mirrors C# ReportComponentBase.GetExpressions()
// (FastReport.Base/ReportComponentBase.cs lines 1018–1044).
func (rc *ReportComponentBase) GetExpressions() []string {
	exprs := rc.ComponentBase.GetExpressions()
	if rc.hyperlink != nil && rc.hyperlink.Expression != "" {
		exprs = append(exprs, rc.hyperlink.Expression)
	}
	if rc.bookmark != "" {
		exprs = append(exprs, rc.bookmark)
	}
	if rc.exportableExpression != "" {
		expr := rc.exportableExpression
		if len(expr) > 2 && expr[0] == '[' && expr[len(expr)-1] == ']' {
			expr = expr[1 : len(expr)-1]
		}
		lower := strings.ToLower(expr)
		if lower == "true" || lower == "false" {
			expr = lower
		}
		exprs = append(exprs, expr)
	}
	return exprs
}

// GetData is called by the report engine to evaluate dynamic properties
// such as Bookmark and Hyperlink before the object is printed.
//
// Mirrors C# ReportComponentBase.GetData()
// (FastReport.Base/ReportComponentBase.cs lines 1006–1015).
// The calc function receives a bracket expression and returns its value.
func (rc *ReportComponentBase) GetData(calc func(string) (any, error)) {
	// Evaluate Hyperlink expression.
	if rc.hyperlink != nil && rc.hyperlink.Expression != "" {
		if val, err := calc(rc.hyperlink.Expression); err == nil && val != nil {
			rc.hyperlink.Value = fmt.Sprintf("%v", val)
		}
	}
	// Evaluate Bookmark expression.
	if rc.bookmark != "" {
		if val, err := calc(rc.bookmark); err == nil {
			if val == nil {
				rc.bookmark = ""
			} else {
				rc.bookmark = fmt.Sprintf("%v", val)
			}
		}
	}
}

// ResetData resets any data-bound state from the previous report run.
// Mirrors C# ReportComponentBase.ResetData() which is virtual and empty by default
// (FastReport.Base/ReportComponentBase.cs line 920).
func (rc *ReportComponentBase) ResetData() {}

// Validate checks this component for common structural problems and returns a
// slice of validation issues. An empty slice means the component is valid.
//
// The checks mirror C# ReportComponentBase.Validate()
// (FastReport.Base/ReportComponentBase.cs lines 802–816):
//   1. Width or Height is zero or negative → Error "incorrect size"
//   2. Name is empty → Error "unnamed object"
//   3. Component's AbsBounds is not contained within its parent's AbsBounds
//      (when the parent exposes geometry) → Error "out of bounds"
//
// This method satisfies the utils.Validatable interface.
func (rc *ReportComponentBase) Validate() []utils.ValidationIssue {
	var issues []utils.ValidationIssue
	name := rc.Name()

	// Check 1: size must be positive.
	// C# reference: ReportComponentBase.cs line 806.
	if rc.Height() <= 0 || rc.Width() <= 0 {
		issues = append(issues, utils.ValidationIssue{
			Severity:   utils.ValidationError,
			ObjectName: name,
			Message:    "incorrect size: width and height must be positive",
		})
	}

	// Check 2: name must not be empty.
	// C# reference: ReportComponentBase.cs line 809.
	if name == "" {
		issues = append(issues, utils.ValidationIssue{
			Severity: utils.ValidationError,
			Message:  "unnamed object: report component has no name",
		})
	}

	// Check 3: component must be contained within its parent's bounds.
	// Only checked when the parent exposes AbsLeft/AbsTop/Width/Height, matching
	// the C# check "(Parent is ReportComponentBase)" guard.
	// C# reference: ReportComponentBase.cs line 812 — Validator.RectContainInOtherRect.
	if p, ok := rc.Parent().(interface {
		AbsLeft() float32
		AbsTop() float32
		Width() float32
		Height() float32
	}); ok {
		abs := rc.AbsBounds()
		if !utils.RectContainInOtherF(p.AbsLeft(), p.AbsTop(), p.Width(), p.Height(),
			abs.Left, abs.Top, abs.Width, abs.Height) {
			issues = append(issues, utils.ValidationIssue{
				Severity:   utils.ValidationError,
				ObjectName: name,
				Message:    "object is out of bounds relative to its parent",
			})
		}
	}

	return issues
}
