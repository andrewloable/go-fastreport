package object_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
)

// -----------------------------------------------------------------------
// CheckBoxObject
// -----------------------------------------------------------------------

func TestNewCheckBoxObject_Defaults(t *testing.T) {
	c := object.NewCheckBoxObject()
	if c == nil {
		t.Fatal("NewCheckBoxObject returned nil")
	}
	if c.Checked() {
		t.Error("Checked should default to false")
	}
	if c.CheckedSymbol() != object.CheckedSymbolCheck {
		t.Errorf("CheckedSymbol default = %d, want Check", c.CheckedSymbol())
	}
	if c.UncheckedSymbol() != object.UncheckedSymbolNone {
		t.Errorf("UncheckedSymbol default = %d, want None", c.UncheckedSymbol())
	}
	if c.CheckWidthRatio() != 1.0 {
		t.Errorf("CheckWidthRatio default = %v, want 1.0", c.CheckWidthRatio())
	}
	if c.DataColumn() != "" {
		t.Errorf("DataColumn default = %q, want empty", c.DataColumn())
	}
	if c.HideIfUnchecked() {
		t.Error("HideIfUnchecked should default to false")
	}
	if c.Editable() {
		t.Error("Editable should default to false")
	}
}

func TestCheckBoxObject_Checked(t *testing.T) {
	c := object.NewCheckBoxObject()
	c.SetChecked(true)
	if !c.Checked() {
		t.Error("Checked should be true")
	}
}

func TestCheckBoxObject_CheckedSymbol(t *testing.T) {
	syms := []object.CheckedSymbol{
		object.CheckedSymbolCheck,
		object.CheckedSymbolCross,
		object.CheckedSymbolPlus,
		object.CheckedSymbolFill,
	}
	for _, s := range syms {
		c := object.NewCheckBoxObject()
		c.SetCheckedSymbol(s)
		if c.CheckedSymbol() != s {
			t.Errorf("CheckedSymbol = %d, want %d", c.CheckedSymbol(), s)
		}
	}
}

func TestCheckBoxObject_UncheckedSymbol(t *testing.T) {
	c := object.NewCheckBoxObject()
	c.SetUncheckedSymbol(object.UncheckedSymbolMinus)
	if c.UncheckedSymbol() != object.UncheckedSymbolMinus {
		t.Error("UncheckedSymbol should be Minus")
	}
}

func TestCheckBoxObject_CheckColor(t *testing.T) {
	c := object.NewCheckBoxObject()
	red := color.RGBA{R: 255, A: 255}
	c.SetCheckColor(red)
	if c.CheckColor() != red {
		t.Errorf("CheckColor = %v, want red", c.CheckColor())
	}
}

func TestCheckBoxObject_DataColumn(t *testing.T) {
	c := object.NewCheckBoxObject()
	c.SetDataColumn("IsActive")
	if c.DataColumn() != "IsActive" {
		t.Errorf("DataColumn = %q, want IsActive", c.DataColumn())
	}
}

func TestCheckBoxObject_Expression(t *testing.T) {
	c := object.NewCheckBoxObject()
	c.SetExpression("[Amount] > 0")
	if c.Expression() != "[Amount] > 0" {
		t.Errorf("Expression = %q", c.Expression())
	}
}

func TestCheckBoxObject_CheckWidthRatio(t *testing.T) {
	c := object.NewCheckBoxObject()
	c.SetCheckWidthRatio(0.8)
	if c.CheckWidthRatio() != 0.8 {
		t.Errorf("CheckWidthRatio = %v, want 0.8", c.CheckWidthRatio())
	}
}

func TestCheckBoxObject_HideIfUnchecked(t *testing.T) {
	c := object.NewCheckBoxObject()
	c.SetHideIfUnchecked(true)
	if !c.HideIfUnchecked() {
		t.Error("HideIfUnchecked should be true")
	}
}

func TestCheckBoxObject_Editable(t *testing.T) {
	c := object.NewCheckBoxObject()
	c.SetEditable(true)
	if !c.Editable() {
		t.Error("Editable should be true")
	}
}

func TestCheckBoxObject_InheritsVisible(t *testing.T) {
	c := object.NewCheckBoxObject()
	if !c.Visible() {
		t.Error("CheckBoxObject should inherit Visible=true")
	}
}

// -----------------------------------------------------------------------
// ContainerObject
// -----------------------------------------------------------------------

// minimalObj is a simple Base stub for containment tests.
type minimalObj struct {
	report.BaseObject
}

func newMinimalObj(name string) *minimalObj {
	o := &minimalObj{BaseObject: *report.NewBaseObject()}
	o.SetName(name)
	return o
}
func (m *minimalObj) Serialize(w report.Writer) error   { return nil }
func (m *minimalObj) Deserialize(r report.Reader) error { return nil }

func TestNewContainerObject_Defaults(t *testing.T) {
	c := object.NewContainerObject()
	if c == nil {
		t.Fatal("NewContainerObject returned nil")
	}
	if c.Objects() == nil {
		t.Error("Objects should not be nil")
	}
	if c.Objects().Len() != 0 {
		t.Errorf("Objects.Len = %d, want 0", c.Objects().Len())
	}
	if c.BeforeLayoutEvent() != "" {
		t.Errorf("BeforeLayoutEvent default = %q, want empty", c.BeforeLayoutEvent())
	}
}

func TestContainerObject_AddChild(t *testing.T) {
	c := object.NewContainerObject()
	obj := newMinimalObj("child1")
	c.AddChild(obj)
	if c.Objects().Len() != 1 {
		t.Errorf("Objects.Len = %d, want 1", c.Objects().Len())
	}
	if obj.Parent() != c {
		t.Error("child.Parent should be the container")
	}
}

func TestContainerObject_RemoveChild(t *testing.T) {
	c := object.NewContainerObject()
	obj := newMinimalObj("child1")
	c.AddChild(obj)
	c.RemoveChild(obj)
	if c.Objects().Len() != 0 {
		t.Errorf("Objects.Len = %d after remove, want 0", c.Objects().Len())
	}
	if obj.Parent() != nil {
		t.Error("child.Parent should be nil after remove")
	}
}

func TestContainerObject_CanContain_NonContainer(t *testing.T) {
	c := object.NewContainerObject()
	obj := newMinimalObj("x")
	if !c.CanContain(obj) {
		t.Error("ContainerObject should accept non-container children")
	}
}

func TestContainerObject_CanContain_Container(t *testing.T) {
	outer := object.NewContainerObject()
	inner := object.NewContainerObject()
	if outer.CanContain(inner) {
		t.Error("ContainerObject should NOT accept nested ContainerObject")
	}
}

func TestContainerObject_GetChildObjects(t *testing.T) {
	c := object.NewContainerObject()
	c.AddChild(newMinimalObj("a"))
	c.AddChild(newMinimalObj("b"))
	var list []report.Base
	c.GetChildObjects(&list)
	if len(list) != 2 {
		t.Errorf("GetChildObjects len = %d, want 2", len(list))
	}
}

func TestContainerObject_GetChildOrder(t *testing.T) {
	c := object.NewContainerObject()
	obj1 := newMinimalObj("first")
	obj2 := newMinimalObj("second")
	c.AddChild(obj1)
	c.AddChild(obj2)
	if c.GetChildOrder(obj1) != 0 {
		t.Errorf("GetChildOrder(obj1) = %d, want 0", c.GetChildOrder(obj1))
	}
	if c.GetChildOrder(obj2) != 1 {
		t.Errorf("GetChildOrder(obj2) = %d, want 1", c.GetChildOrder(obj2))
	}
}

func TestContainerObject_SetChildOrder(t *testing.T) {
	c := object.NewContainerObject()
	obj1 := newMinimalObj("first")
	obj2 := newMinimalObj("second")
	c.AddChild(obj1)
	c.AddChild(obj2)
	c.SetChildOrder(obj1, 1)
	if c.Objects().Get(0) != obj2 {
		t.Error("after reorder, obj2 should be at index 0")
	}
	if c.Objects().Get(1) != obj1 {
		t.Error("after reorder, obj1 should be at index 1")
	}
}

func TestContainerObject_LayoutEvents(t *testing.T) {
	c := object.NewContainerObject()
	var log []string
	c.OnBeforeLayout = func(sender report.Base, e *report.EventArgs) { log = append(log, "before") }
	c.OnAfterLayout = func(sender report.Base, e *report.EventArgs) { log = append(log, "after") }
	c.FireBeforeLayout()
	c.FireAfterLayout()
	if len(log) != 2 || log[0] != "before" || log[1] != "after" {
		t.Errorf("layout events: got %v, want [before after]", log)
	}
}

func TestContainerObject_LayoutEventNames(t *testing.T) {
	c := object.NewContainerObject()
	c.SetBeforeLayoutEvent("C1_Before")
	c.SetAfterLayoutEvent("C1_After")
	if c.BeforeLayoutEvent() != "C1_Before" {
		t.Errorf("BeforeLayoutEvent = %q", c.BeforeLayoutEvent())
	}
	if c.AfterLayoutEvent() != "C1_After" {
		t.Errorf("AfterLayoutEvent = %q", c.AfterLayoutEvent())
	}
}

func TestContainerObject_UpdateLayout_NoOp(t *testing.T) {
	c := object.NewContainerObject()
	c.UpdateLayout(10, 20) // must not panic
}

// -----------------------------------------------------------------------
// SubreportObject
// -----------------------------------------------------------------------

func TestNewSubreportObject_Defaults(t *testing.T) {
	s := object.NewSubreportObject()
	if s == nil {
		t.Fatal("NewSubreportObject returned nil")
	}
	if s.ReportPageName() != "" {
		t.Errorf("ReportPageName default = %q, want empty", s.ReportPageName())
	}
	if s.PrintOnParent() {
		t.Error("PrintOnParent should default to false")
	}
}

func TestSubreportObject_ReportPageName(t *testing.T) {
	s := object.NewSubreportObject()
	s.SetReportPageName("DetailPage")
	if s.ReportPageName() != "DetailPage" {
		t.Errorf("ReportPageName = %q, want DetailPage", s.ReportPageName())
	}
}

func TestSubreportObject_PrintOnParent(t *testing.T) {
	s := object.NewSubreportObject()
	s.SetPrintOnParent(true)
	if !s.PrintOnParent() {
		t.Error("PrintOnParent should be true")
	}
}

func TestSubreportObject_InheritsVisible(t *testing.T) {
	s := object.NewSubreportObject()
	if !s.Visible() {
		t.Error("SubreportObject should inherit Visible=true")
	}
}
