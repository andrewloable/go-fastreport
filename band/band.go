// Package band implements the band hierarchy for go-fastreport.
// Bands are the horizontal strips that make up a report page (header, data, footer, etc.).
package band

import (
	"github.com/andrewloable/go-fastreport/report"
)

// BandBase is the base struct for all report bands.
// It extends BreakableComponent and implements report.Parent so it can
// contain child report objects.
// It is the Go equivalent of FastReport.BandBase.
type BandBase struct {
	report.BreakableComponent

	// child is the ChildBand printed immediately after this band.
	child *ChildBand

	// objects holds all direct child report objects of this band.
	objects *report.ObjectCollection

	// Layout/paging behaviour.
	startNewPage          bool
	firstRowStartsNewPage bool // default true (matches .NET DefaultValue(true))
	printOnBottom         bool
	keepChild             bool
	outlineExpression     string
	repeatBandNTimes      int // default 1

	// Runtime row tracking (set by engine, not serialized).
	repeated  bool
	rowNo     int
	absRowNo  int
	isFirstRow bool
	isLastRow  bool

	// Internal engine flags.
	FlagUseStartNewPage bool
	FlagCheckFreeSpace  bool

	// Ruler guides in designer (pixel offsets from band left edge).
	guides []float32

	// Layout event callbacks.
	OnBeforeLayout report.EventHandler
	OnAfterLayout  report.EventHandler

	// FRX script event names.
	beforeLayoutEvent string
	afterLayoutEvent  string

	// Engine: offset applied when a band is reprinted on a new page.
	reprintOffset float32
}

// NewBandBase creates a BandBase with defaults:
// firstRowStartsNewPage=true, repeatBandNTimes=1.
func NewBandBase() *BandBase {
	b := &BandBase{
		BreakableComponent:    *report.NewBreakableComponent(),
		firstRowStartsNewPage: true,
		repeatBandNTimes:      1,
		objects:               report.NewObjectCollection(),
	}
	return b
}

// --- Child band ---

// Child returns the ChildBand that prints immediately after this band.
func (b *BandBase) Child() *ChildBand { return b.child }

// SetChild sets the child band.
func (b *BandBase) SetChild(c *ChildBand) { b.child = c }

// --- Objects ---

// Objects returns the collection of direct child report objects.
func (b *BandBase) Objects() *report.ObjectCollection { return b.objects }

// --- Layout / paging ---

// StartNewPage returns whether the band forces a page break before printing.
func (b *BandBase) StartNewPage() bool { return b.startNewPage }

// SetStartNewPage sets the start-new-page flag.
func (b *BandBase) SetStartNewPage(v bool) { b.startNewPage = v }

// FirstRowStartsNewPage returns whether the first data row may start a new page.
func (b *BandBase) FirstRowStartsNewPage() bool { return b.firstRowStartsNewPage }

// SetFirstRowStartsNewPage controls whether the first row starts a new page.
func (b *BandBase) SetFirstRowStartsNewPage(v bool) { b.firstRowStartsNewPage = v }

// PrintOnBottom returns whether the band is pushed to the page bottom.
func (b *BandBase) PrintOnBottom() bool { return b.printOnBottom }

// SetPrintOnBottom sets the print-on-bottom flag.
func (b *BandBase) SetPrintOnBottom(v bool) { b.printOnBottom = v }

// KeepChild returns whether the band must stay on the same page as its child.
func (b *BandBase) KeepChild() bool { return b.keepChild }

// SetKeepChild sets the keep-child flag.
func (b *BandBase) SetKeepChild(v bool) { b.keepChild = v }

// OutlineExpression returns the expression used to build the preview outline.
func (b *BandBase) OutlineExpression() string { return b.outlineExpression }

// SetOutlineExpression sets the outline expression.
func (b *BandBase) SetOutlineExpression(expr string) { b.outlineExpression = expr }

// RepeatBandNTimes returns how many times this band is repeated. Default 1.
func (b *BandBase) RepeatBandNTimes() int { return b.repeatBandNTimes }

// SetRepeatBandNTimes sets the repeat count.
func (b *BandBase) SetRepeatBandNTimes(n int) { b.repeatBandNTimes = n }

// --- Row tracking ---

// Repeated returns true when this band is being reprinted on a new page.
func (b *BandBase) Repeated() bool { return b.repeated }

// SetRepeated propagates the repeated flag to child bands.
func (b *BandBase) SetRepeated(v bool) {
	b.repeated = v
	c := b.child
	for c != nil {
		c.repeated = v
		c = c.child
	}
}

// RowNo returns the current data row number (same as [Row#] system variable).
func (b *BandBase) RowNo() int { return b.rowNo }

// SetRowNo sets the row number and propagates it to child bands.
func (b *BandBase) SetRowNo(n int) {
	b.rowNo = n
	if b.child != nil {
		b.child.SetRowNo(n)
	}
}

// AbsRowNo returns the absolute row number across all data groups.
func (b *BandBase) AbsRowNo() int { return b.absRowNo }

// SetAbsRowNo sets the absolute row number and propagates it to child bands.
func (b *BandBase) SetAbsRowNo(n int) {
	b.absRowNo = n
	if b.child != nil {
		b.child.SetAbsRowNo(n)
	}
}

// IsFirstRow returns true when this is the first data row.
func (b *BandBase) IsFirstRow() bool { return b.isFirstRow }

// SetIsFirstRow sets the first-row flag.
func (b *BandBase) SetIsFirstRow(v bool) { b.isFirstRow = v }

// IsLastRow returns true when this is the last data row.
func (b *BandBase) IsLastRow() bool { return b.isLastRow }

// SetIsLastRow sets the last-row flag.
func (b *BandBase) SetIsLastRow(v bool) { b.isLastRow = v }

// --- Guides ---

// Guides returns the ruler guide positions in pixels.
func (b *BandBase) Guides() []float32 { return b.guides }

// SetGuides sets the guide positions.
func (b *BandBase) SetGuides(g []float32) { b.guides = g }

// AddGuide appends a guide at the given pixel offset.
func (b *BandBase) AddGuide(pos float32) { b.guides = append(b.guides, pos) }

// --- Events ---

// FireBeforeLayout calls OnBeforeLayout if set.
func (b *BandBase) FireBeforeLayout() {
	if b.OnBeforeLayout != nil {
		b.OnBeforeLayout(b, &report.EventArgs{})
	}
}

// FireAfterLayout calls OnAfterLayout if set.
func (b *BandBase) FireAfterLayout() {
	if b.OnAfterLayout != nil {
		b.OnAfterLayout(b, &report.EventArgs{})
	}
}

// BeforeLayoutEvent returns the FRX script event name for before-layout.
func (b *BandBase) BeforeLayoutEvent() string { return b.beforeLayoutEvent }

// SetBeforeLayoutEvent sets the before-layout script event name.
func (b *BandBase) SetBeforeLayoutEvent(s string) { b.beforeLayoutEvent = s }

// AfterLayoutEvent returns the FRX script event name for after-layout.
func (b *BandBase) AfterLayoutEvent() string { return b.afterLayoutEvent }

// SetAfterLayoutEvent sets the after-layout script event name.
func (b *BandBase) SetAfterLayoutEvent(s string) { b.afterLayoutEvent = s }

// ReprintOffset is the vertical offset applied when reprinting on a new page.
func (b *BandBase) ReprintOffset() float32 { return b.reprintOffset }

// SetReprintOffset sets the reprint offset.
func (b *BandBase) SetReprintOffset(v float32) { b.reprintOffset = v }

// --- report.Parent implementation ---

// CanContain returns true when this band can accept child as a direct child.
// Accepts any report.Base that is not another BandBase (except ChildBand).
func (b *BandBase) CanContain(child report.Base) bool {
	if _, ok := child.(*ChildBand); ok {
		return true
	}
	if _, ok := child.(*BandBase); ok {
		return false
	}
	return true
}

// GetChildObjects fills list with all direct children (objects + child band).
func (b *BandBase) GetChildObjects(list *[]report.Base) {
	for i := 0; i < b.objects.Len(); i++ {
		*list = append(*list, b.objects.Get(i))
	}
	if b.child != nil {
		*list = append(*list, b.child)
	}
}

// AddChild adds child to this band's object collection or sets it as the child band.
func (b *BandBase) AddChild(child report.Base) {
	if cb, ok := child.(*ChildBand); ok {
		b.child = cb
		cb.SetParent(b)
		return
	}
	b.objects.Add(child)
	child.SetParent(b)
}

// RemoveChild removes child from the objects collection or clears the child band.
func (b *BandBase) RemoveChild(child report.Base) {
	if cb, ok := child.(*ChildBand); ok && b.child == cb {
		b.child = nil
		cb.SetParent(nil)
		return
	}
	if b.objects.Remove(child) {
		child.SetParent(nil)
	}
}

// GetChildOrder returns the z-order index of child in the objects collection,
// or -1 when not found.
func (b *BandBase) GetChildOrder(child report.Base) int {
	return b.objects.IndexOf(child)
}

// SetChildOrder moves child to the given z-order position.
func (b *BandBase) SetChildOrder(child report.Base, order int) {
	idx := b.objects.IndexOf(child)
	if idx < 0 {
		return
	}
	b.objects.RemoveAt(idx)
	if order > b.objects.Len() {
		order = b.objects.Len()
	}
	b.objects.Insert(order, child)
}

// UpdateLayout adjusts child positions when the parent band resizes by dx, dy.
func (b *BandBase) UpdateLayout(dx, dy float32) {
	// Default: no-op. Engine handles layout during prepare.
}

// --- Serialization ---

// serializeAttrs writes BandBase XML attributes only (no child elements).
// This is called by derived band types that need to add their own attributes
// before the child elements are written, because XML attributes must precede
// nested child elements in a streaming encoder.
func (b *BandBase) serializeAttrs(w report.Writer) error {
	if err := b.BreakableComponent.Serialize(w); err != nil {
		return err
	}
	if b.startNewPage {
		w.WriteBool("StartNewPage", true)
	}
	if !b.firstRowStartsNewPage {
		w.WriteBool("FirstRowStartsNewPage", false)
	}
	if b.printOnBottom {
		w.WriteBool("PrintOnBottom", true)
	}
	if b.keepChild {
		w.WriteBool("KeepChild", true)
	}
	if b.outlineExpression != "" {
		w.WriteStr("OutlineExpression", b.outlineExpression)
	}
	if b.repeatBandNTimes != 1 {
		w.WriteInt("RepeatBandNTimes", b.repeatBandNTimes)
	}
	if b.beforeLayoutEvent != "" {
		w.WriteStr("BeforeLayoutEvent", b.beforeLayoutEvent)
	}
	if b.afterLayoutEvent != "" {
		w.WriteStr("AfterLayoutEvent", b.afterLayoutEvent)
	}
	return nil
}

// serializeChildren writes the BandBase child object elements.
// Must be called after all attributes have been written.
func (b *BandBase) serializeChildren(w report.Writer) error {
	for i := 0; i < b.objects.Len(); i++ {
		if err := w.WriteObject(b.objects.Get(i)); err != nil {
			return err
		}
	}
	return nil
}

// Serialize writes BandBase properties that differ from defaults,
// followed by child object elements.
func (b *BandBase) Serialize(w report.Writer) error {
	if err := b.serializeAttrs(w); err != nil {
		return err
	}
	return b.serializeChildren(w)
}

// Deserialize reads BandBase properties.
func (b *BandBase) Deserialize(r report.Reader) error {
	if err := b.BreakableComponent.Deserialize(r); err != nil {
		return err
	}
	b.startNewPage = r.ReadBool("StartNewPage", false)
	b.firstRowStartsNewPage = r.ReadBool("FirstRowStartsNewPage", true)
	b.printOnBottom = r.ReadBool("PrintOnBottom", false)
	b.keepChild = r.ReadBool("KeepChild", false)
	b.outlineExpression = r.ReadStr("OutlineExpression", "")
	b.repeatBandNTimes = r.ReadInt("RepeatBandNTimes", 1)
	b.beforeLayoutEvent = r.ReadStr("BeforeLayoutEvent", "")
	b.afterLayoutEvent = r.ReadStr("AfterLayoutEvent", "")
	return nil
}

// ChildBand is a band that prints immediately after its parent band.
// It is the Go equivalent of FastReport.ChildBand.
type ChildBand struct {
	BandBase

	// FillUnusedSpace causes the band to be printed repeatedly to fill any
	// remaining unused space on the page after all data rows are printed.
	FillUnusedSpace bool

	// CompleteToNRows repeats the band to make the data area reach a total
	// of N rows. If the parent DataBand has fewer than N rows, this band
	// fills the remaining rows with blank content (default 0 = disabled).
	CompleteToNRows int

	// PrintIfDatabandEmpty causes the band to be printed when its parent
	// DataBand has no rows (e.g. to show a "No data" message).
	PrintIfDatabandEmpty bool
}

// NewChildBand creates a new ChildBand.
func NewChildBand() *ChildBand {
	return &ChildBand{BandBase: *NewBandBase()}
}

// TypeName returns the FRX element name for this band.
func (*ChildBand) TypeName() string { return "Child" }

// Serialize writes ChildBand properties that differ from defaults.
func (c *ChildBand) Serialize(w report.Writer) error {
	if err := c.BandBase.Serialize(w); err != nil {
		return err
	}
	if c.FillUnusedSpace {
		w.WriteBool("FillUnusedSpace", true)
	}
	if c.CompleteToNRows != 0 {
		w.WriteInt("CompleteToNRows", c.CompleteToNRows)
	}
	if c.PrintIfDatabandEmpty {
		w.WriteBool("PrintIfDatabandEmpty", true)
	}
	return nil
}

// Deserialize reads ChildBand properties.
func (c *ChildBand) Deserialize(r report.Reader) error {
	if err := c.BandBase.Deserialize(r); err != nil {
		return err
	}
	c.FillUnusedSpace = r.ReadBool("FillUnusedSpace", false)
	c.CompleteToNRows = r.ReadInt("CompleteToNRows", 0)
	c.PrintIfDatabandEmpty = r.ReadBool("PrintIfDatabandEmpty", false)
	return nil
}
