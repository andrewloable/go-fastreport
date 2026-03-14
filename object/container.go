package object

import (
	"image/color"

	"github.com/andrewloable/go-fastreport/report"
)

// -----------------------------------------------------------------------
// CheckBoxObject
// -----------------------------------------------------------------------

// CheckedSymbol specifies the symbol drawn when a CheckBox is checked.
type CheckedSymbol int

const (
	CheckedSymbolCheck CheckedSymbol = iota
	CheckedSymbolCross
	CheckedSymbolPlus
	CheckedSymbolFill
)

// UncheckedSymbol specifies the symbol drawn when a CheckBox is unchecked.
type UncheckedSymbol int

const (
	UncheckedSymbolNone      UncheckedSymbol = iota
	UncheckedSymbolCross
	UncheckedSymbolMinus
	UncheckedSymbolSlash
	UncheckedSymbolBackSlash
)

// CheckBoxObject displays a check box that can be bound to a boolean expression.
// It is the Go equivalent of FastReport.CheckBoxObject.
type CheckBoxObject struct {
	report.ReportComponentBase

	isChecked        bool
	checkedSymbol    CheckedSymbol
	uncheckedSymbol  UncheckedSymbol
	checkColor       color.RGBA
	dataColumn       string
	expression       string
	checkWidthRatio  float32 // default 1.0
	hideIfUnchecked  bool
	editable         bool
}

// NewCheckBoxObject creates a CheckBoxObject with defaults.
func NewCheckBoxObject() *CheckBoxObject {
	return &CheckBoxObject{
		ReportComponentBase: *report.NewReportComponentBase(),
		checkColor:          color.RGBA{A: 255}, // black
		checkWidthRatio:     1.0,
	}
}

// Checked returns whether the check box is in the checked state.
func (c *CheckBoxObject) Checked() bool { return c.isChecked }

// SetChecked sets the checked state.
func (c *CheckBoxObject) SetChecked(v bool) { c.isChecked = v }

// CheckedSymbol returns the symbol used in the checked state.
func (c *CheckBoxObject) CheckedSymbol() CheckedSymbol { return c.checkedSymbol }

// SetCheckedSymbol sets the checked symbol.
func (c *CheckBoxObject) SetCheckedSymbol(s CheckedSymbol) { c.checkedSymbol = s }

// UncheckedSymbol returns the symbol used in the unchecked state.
func (c *CheckBoxObject) UncheckedSymbol() UncheckedSymbol { return c.uncheckedSymbol }

// SetUncheckedSymbol sets the unchecked symbol.
func (c *CheckBoxObject) SetUncheckedSymbol(s UncheckedSymbol) { c.uncheckedSymbol = s }

// CheckColor returns the colour used to draw the check symbol.
func (c *CheckBoxObject) CheckColor() color.RGBA { return c.checkColor }

// SetCheckColor sets the check colour.
func (c *CheckBoxObject) SetCheckColor(col color.RGBA) { c.checkColor = col }

// DataColumn returns the data source column that provides the checked value.
func (c *CheckBoxObject) DataColumn() string { return c.dataColumn }

// SetDataColumn sets the data column binding.
func (c *CheckBoxObject) SetDataColumn(s string) { c.dataColumn = s }

// Expression returns the boolean expression that determines the checked state.
func (c *CheckBoxObject) Expression() string { return c.expression }

// SetExpression sets the boolean expression.
func (c *CheckBoxObject) SetExpression(s string) { c.expression = s }

// CheckWidthRatio returns the width scaling factor for the check symbol.
func (c *CheckBoxObject) CheckWidthRatio() float32 { return c.checkWidthRatio }

// SetCheckWidthRatio sets the check width ratio.
func (c *CheckBoxObject) SetCheckWidthRatio(v float32) { c.checkWidthRatio = v }

// HideIfUnchecked returns whether the object is hidden when unchecked.
func (c *CheckBoxObject) HideIfUnchecked() bool { return c.hideIfUnchecked }

// SetHideIfUnchecked sets the hide-if-unchecked flag.
func (c *CheckBoxObject) SetHideIfUnchecked(v bool) { c.hideIfUnchecked = v }

// Editable returns whether the check box can be toggled in the viewer.
func (c *CheckBoxObject) Editable() bool { return c.editable }

// SetEditable sets the editable flag.
func (c *CheckBoxObject) SetEditable(v bool) { c.editable = v }

// Serialize writes CheckBoxObject properties that differ from defaults.
func (c *CheckBoxObject) Serialize(w report.Writer) error {
	if err := c.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if c.isChecked {
		w.WriteBool("Checked", true)
	}
	if c.checkedSymbol != CheckedSymbolCheck {
		w.WriteInt("CheckedSymbol", int(c.checkedSymbol))
	}
	if c.uncheckedSymbol != UncheckedSymbolNone {
		w.WriteInt("UncheckedSymbol", int(c.uncheckedSymbol))
	}
	if c.dataColumn != "" {
		w.WriteStr("DataColumn", c.dataColumn)
	}
	if c.expression != "" {
		w.WriteStr("Expression", c.expression)
	}
	if c.checkWidthRatio != 1.0 {
		w.WriteFloat("CheckWidthRatio", c.checkWidthRatio)
	}
	if c.hideIfUnchecked {
		w.WriteBool("HideIfUnchecked", true)
	}
	if c.editable {
		w.WriteBool("Editable", true)
	}
	return nil
}

// Deserialize reads CheckBoxObject properties.
func (c *CheckBoxObject) Deserialize(r report.Reader) error {
	if err := c.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	c.isChecked = r.ReadBool("Checked", false)
	c.checkedSymbol = CheckedSymbol(r.ReadInt("CheckedSymbol", 0))
	c.uncheckedSymbol = UncheckedSymbol(r.ReadInt("UncheckedSymbol", 0))
	c.dataColumn = r.ReadStr("DataColumn", "")
	c.expression = r.ReadStr("Expression", "")
	c.checkWidthRatio = r.ReadFloat("CheckWidthRatio", 1.0)
	c.hideIfUnchecked = r.ReadBool("HideIfUnchecked", false)
	c.editable = r.ReadBool("Editable", false)
	return nil
}

// -----------------------------------------------------------------------
// ContainerObject
// -----------------------------------------------------------------------

// ContainerObject is a layout container that holds child report objects.
// It is the Go equivalent of FastReport.ContainerObject.
type ContainerObject struct {
	report.ReportComponentBase

	objects             *report.ObjectCollection
	beforeLayoutEvent   string
	afterLayoutEvent    string
	OnBeforeLayout      report.EventHandler
	OnAfterLayout       report.EventHandler
}

// NewContainerObject creates a ContainerObject with defaults.
func NewContainerObject() *ContainerObject {
	return &ContainerObject{
		ReportComponentBase: *report.NewReportComponentBase(),
		objects:             report.NewObjectCollection(),
	}
}

// Objects returns the child object collection.
func (c *ContainerObject) Objects() *report.ObjectCollection { return c.objects }

// BeforeLayoutEvent returns the script event name fired before layout.
func (c *ContainerObject) BeforeLayoutEvent() string { return c.beforeLayoutEvent }

// SetBeforeLayoutEvent sets the before-layout event name.
func (c *ContainerObject) SetBeforeLayoutEvent(s string) { c.beforeLayoutEvent = s }

// AfterLayoutEvent returns the script event name fired after layout.
func (c *ContainerObject) AfterLayoutEvent() string { return c.afterLayoutEvent }

// SetAfterLayoutEvent sets the after-layout event name.
func (c *ContainerObject) SetAfterLayoutEvent(s string) { c.afterLayoutEvent = s }

// FireBeforeLayout fires the OnBeforeLayout event if set.
func (c *ContainerObject) FireBeforeLayout() {
	if c.OnBeforeLayout != nil {
		c.OnBeforeLayout(c, &report.EventArgs{})
	}
}

// FireAfterLayout fires the OnAfterLayout event if set.
func (c *ContainerObject) FireAfterLayout() {
	if c.OnAfterLayout != nil {
		c.OnAfterLayout(c, &report.EventArgs{})
	}
}

// CanContain returns true for any non-container child.
func (c *ContainerObject) CanContain(child report.Base) bool {
	_, isCont := child.(*ContainerObject)
	return !isCont
}

// AddChild adds a child object to the container.
func (c *ContainerObject) AddChild(child report.Base) {
	c.objects.Add(child)
	child.SetParent(c)
}

// RemoveChild removes a child object from the container.
func (c *ContainerObject) RemoveChild(child report.Base) {
	if c.objects.Remove(child) {
		child.SetParent(nil)
	}
}

// GetChildObjects fills list with all child objects.
func (c *ContainerObject) GetChildObjects(list *[]report.Base) {
	for i := 0; i < c.objects.Len(); i++ {
		*list = append(*list, c.objects.Get(i))
	}
}

// GetChildOrder returns the z-order index of child.
func (c *ContainerObject) GetChildOrder(child report.Base) int {
	return c.objects.IndexOf(child)
}

// SetChildOrder moves child to the given z-order position.
func (c *ContainerObject) SetChildOrder(child report.Base, order int) {
	idx := c.objects.IndexOf(child)
	if idx < 0 {
		return
	}
	c.objects.RemoveAt(idx)
	if order > c.objects.Len() {
		order = c.objects.Len()
	}
	c.objects.Insert(order, child)
}

// UpdateLayout is a no-op in the base implementation.
func (c *ContainerObject) UpdateLayout(dx, dy float32) {}

// Serialize writes ContainerObject properties that differ from defaults.
func (c *ContainerObject) Serialize(w report.Writer) error {
	if err := c.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if c.beforeLayoutEvent != "" {
		w.WriteStr("BeforeLayoutEvent", c.beforeLayoutEvent)
	}
	if c.afterLayoutEvent != "" {
		w.WriteStr("AfterLayoutEvent", c.afterLayoutEvent)
	}
	for i := 0; i < c.objects.Len(); i++ {
		if err := w.WriteObject(c.objects.Get(i)); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads ContainerObject properties.
func (c *ContainerObject) Deserialize(r report.Reader) error {
	if err := c.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	c.beforeLayoutEvent = r.ReadStr("BeforeLayoutEvent", "")
	c.afterLayoutEvent = r.ReadStr("AfterLayoutEvent", "")
	return nil
}

// -----------------------------------------------------------------------
// SubreportObject
// -----------------------------------------------------------------------

// SubreportObject embeds a reference to another report page, allowing
// nested report execution.
// It is the Go equivalent of FastReport.SubreportObject.
type SubreportObject struct {
	report.ReportComponentBase

	// reportPageName is the name of the ReportPage this subreport points to.
	reportPageName string
	// printOnParent causes the subreport to print on the parent page.
	printOnParent bool
}

// NewSubreportObject creates a SubreportObject with defaults.
func NewSubreportObject() *SubreportObject {
	return &SubreportObject{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// ReportPageName returns the name of the linked ReportPage.
func (s *SubreportObject) ReportPageName() string { return s.reportPageName }

// SetReportPageName sets the linked page name.
func (s *SubreportObject) SetReportPageName(name string) { s.reportPageName = name }

// PrintOnParent returns whether the subreport prints on the parent page.
func (s *SubreportObject) PrintOnParent() bool { return s.printOnParent }

// SetPrintOnParent sets the print-on-parent flag.
func (s *SubreportObject) SetPrintOnParent(v bool) { s.printOnParent = v }

// Serialize writes SubreportObject properties that differ from defaults.
func (s *SubreportObject) Serialize(w report.Writer) error {
	if err := s.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if s.reportPageName != "" {
		w.WriteStr("ReportPage", s.reportPageName)
	}
	if s.printOnParent {
		w.WriteBool("PrintOnParent", true)
	}
	return nil
}

// Deserialize reads SubreportObject properties.
func (s *SubreportObject) Deserialize(r report.Reader) error {
	if err := s.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	s.reportPageName = r.ReadStr("ReportPage", "")
	s.printOnParent = r.ReadBool("PrintOnParent", false)
	return nil
}
