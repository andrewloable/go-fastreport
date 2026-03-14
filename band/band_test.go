package band_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/report"
)

// --- helpers ---

// minimalBase is a minimal report.Base stub for testing child containment.
type minimalBase struct {
	report.BaseObject
}

func newMinimalBase(name string) *minimalBase {
	b := &minimalBase{BaseObject: *report.NewBaseObject()}
	b.SetName(name)
	return b
}

// Satisfy report.Serializable (noop).
func (m *minimalBase) Serialize(w report.Writer) error   { return nil }
func (m *minimalBase) Deserialize(r report.Reader) error { return nil }

// --- BandBase ---

func TestNewBandBase_Defaults(t *testing.T) {
	b := band.NewBandBase()
	if b == nil {
		t.Fatal("NewBandBase returned nil")
	}
	if b.StartNewPage() {
		t.Error("StartNewPage should default to false")
	}
	if !b.FirstRowStartsNewPage() {
		t.Error("FirstRowStartsNewPage should default to true")
	}
	if b.PrintOnBottom() {
		t.Error("PrintOnBottom should default to false")
	}
	if b.KeepChild() {
		t.Error("KeepChild should default to false")
	}
	if b.RepeatBandNTimes() != 1 {
		t.Errorf("RepeatBandNTimes default = %d, want 1", b.RepeatBandNTimes())
	}
	if b.Child() != nil {
		t.Error("Child should default to nil")
	}
	if b.Objects() == nil {
		t.Error("Objects should not be nil")
	}
	// Inherits BreakableComponent defaults.
	if !b.CanBreak() {
		t.Error("CanBreak should default to true (from BreakableComponent)")
	}
}

func TestBandBase_StartNewPage(t *testing.T) {
	b := band.NewBandBase()
	b.SetStartNewPage(true)
	if !b.StartNewPage() {
		t.Error("StartNewPage should be true after SetStartNewPage(true)")
	}
}

func TestBandBase_FirstRowStartsNewPage(t *testing.T) {
	b := band.NewBandBase()
	b.SetFirstRowStartsNewPage(false)
	if b.FirstRowStartsNewPage() {
		t.Error("FirstRowStartsNewPage should be false after set")
	}
}

func TestBandBase_PrintOnBottom(t *testing.T) {
	b := band.NewBandBase()
	b.SetPrintOnBottom(true)
	if !b.PrintOnBottom() {
		t.Error("PrintOnBottom should be true")
	}
}

func TestBandBase_KeepChild(t *testing.T) {
	b := band.NewBandBase()
	b.SetKeepChild(true)
	if !b.KeepChild() {
		t.Error("KeepChild should be true")
	}
}

func TestBandBase_OutlineExpression(t *testing.T) {
	b := band.NewBandBase()
	b.SetOutlineExpression("[Name]")
	if b.OutlineExpression() != "[Name]" {
		t.Errorf("OutlineExpression = %q, want [Name]", b.OutlineExpression())
	}
}

func TestBandBase_RepeatBandNTimes(t *testing.T) {
	b := band.NewBandBase()
	b.SetRepeatBandNTimes(3)
	if b.RepeatBandNTimes() != 3 {
		t.Errorf("RepeatBandNTimes = %d, want 3", b.RepeatBandNTimes())
	}
}

func TestBandBase_RowTracking(t *testing.T) {
	b := band.NewBandBase()
	b.SetRowNo(5)
	b.SetAbsRowNo(42)
	b.SetIsFirstRow(true)
	b.SetIsLastRow(true)

	if b.RowNo() != 5 {
		t.Errorf("RowNo = %d, want 5", b.RowNo())
	}
	if b.AbsRowNo() != 42 {
		t.Errorf("AbsRowNo = %d, want 42", b.AbsRowNo())
	}
	if !b.IsFirstRow() {
		t.Error("IsFirstRow should be true")
	}
	if !b.IsLastRow() {
		t.Error("IsLastRow should be true")
	}
}

func TestBandBase_Repeated_PropagatestoChild(t *testing.T) {
	parent := band.NewBandBase()
	child := band.NewChildBand()
	parent.SetChild(child)

	parent.SetRepeated(true)
	if !parent.Repeated() {
		t.Error("parent.Repeated should be true")
	}
	if !child.Repeated() {
		t.Error("child.Repeated should be true after propagation")
	}
}

func TestBandBase_RowNo_PropagatestoChild(t *testing.T) {
	parent := band.NewBandBase()
	child := band.NewChildBand()
	parent.SetChild(child)

	parent.SetRowNo(7)
	if child.RowNo() != 7 {
		t.Errorf("child.RowNo = %d, want 7 after propagation", child.RowNo())
	}

	parent.SetAbsRowNo(99)
	if child.AbsRowNo() != 99 {
		t.Errorf("child.AbsRowNo = %d, want 99", child.AbsRowNo())
	}
}

func TestBandBase_Guides(t *testing.T) {
	b := band.NewBandBase()
	b.AddGuide(100)
	b.AddGuide(200)
	if len(b.Guides()) != 2 {
		t.Fatalf("Guides len = %d, want 2", len(b.Guides()))
	}
	if b.Guides()[0] != 100 || b.Guides()[1] != 200 {
		t.Errorf("Guides = %v, want [100 200]", b.Guides())
	}
	b.SetGuides([]float32{50})
	if len(b.Guides()) != 1 || b.Guides()[0] != 50 {
		t.Errorf("SetGuides: Guides = %v, want [50]", b.Guides())
	}
}

func TestBandBase_LayoutEvents(t *testing.T) {
	b := band.NewBandBase()
	var log []string
	b.OnBeforeLayout = func(sender report.Base, e *report.EventArgs) { log = append(log, "before") }
	b.OnAfterLayout = func(sender report.Base, e *report.EventArgs) { log = append(log, "after") }

	b.FireBeforeLayout()
	b.FireAfterLayout()

	if len(log) != 2 || log[0] != "before" || log[1] != "after" {
		t.Errorf("layout events: got %v, want [before after]", log)
	}
}

func TestBandBase_LayoutEvents_NilSafe(t *testing.T) {
	b := band.NewBandBase()
	b.FireBeforeLayout()
	b.FireAfterLayout()
}

func TestBandBase_LayoutEventNames(t *testing.T) {
	b := band.NewBandBase()
	b.SetBeforeLayoutEvent("Band1_BeforeLayout")
	b.SetAfterLayoutEvent("Band1_AfterLayout")
	if b.BeforeLayoutEvent() != "Band1_BeforeLayout" {
		t.Errorf("BeforeLayoutEvent = %q", b.BeforeLayoutEvent())
	}
	if b.AfterLayoutEvent() != "Band1_AfterLayout" {
		t.Errorf("AfterLayoutEvent = %q", b.AfterLayoutEvent())
	}
}

func TestBandBase_FlagFields(t *testing.T) {
	b := band.NewBandBase()
	b.FlagUseStartNewPage = true
	b.FlagCheckFreeSpace = true
	b.FlagMustBreak = true

	if !b.FlagUseStartNewPage || !b.FlagCheckFreeSpace || !b.FlagMustBreak {
		t.Error("engine flag fields should be settable")
	}
}

func TestBandBase_ReprintOffset(t *testing.T) {
	b := band.NewBandBase()
	b.SetReprintOffset(12.5)
	if b.ReprintOffset() != 12.5 {
		t.Errorf("ReprintOffset = %v, want 12.5", b.ReprintOffset())
	}
}

func TestBandBase_ChildBand(t *testing.T) {
	parent := band.NewBandBase()
	child := band.NewChildBand()

	parent.SetChild(child)
	if parent.Child() != child {
		t.Error("Child should be set")
	}

	parent.SetChild(nil)
	if parent.Child() != nil {
		t.Error("Child should be nil after SetChild(nil)")
	}
}

// --- report.Parent implementation ---

func TestBandBase_CanContain(t *testing.T) {
	b := band.NewBandBase()
	obj := newMinimalBase("obj1")
	child := band.NewChildBand()

	if !b.CanContain(obj) {
		t.Error("BandBase should accept non-band children")
	}
	if !b.CanContain(child) {
		t.Error("BandBase should accept ChildBand")
	}
	nested := band.NewBandBase()
	if b.CanContain(nested) {
		t.Error("BandBase should NOT accept another BandBase as child")
	}
}

func TestBandBase_AddChild_Object(t *testing.T) {
	b := band.NewBandBase()
	obj := newMinimalBase("obj1")

	b.AddChild(obj)
	if b.Objects().Len() != 1 {
		t.Errorf("Objects().Len = %d, want 1", b.Objects().Len())
	}
	if obj.Parent() != b {
		t.Error("obj.Parent should be the band after AddChild")
	}
}

func TestBandBase_AddChild_ChildBand(t *testing.T) {
	b := band.NewBandBase()
	child := band.NewChildBand()

	b.AddChild(child)
	if b.Child() != child {
		t.Error("AddChild with ChildBand should set Child")
	}
	if child.Parent() != b {
		t.Error("ChildBand.Parent should be the band after AddChild")
	}
}

func TestBandBase_RemoveChild_Object(t *testing.T) {
	b := band.NewBandBase()
	obj := newMinimalBase("obj1")
	b.AddChild(obj)
	b.RemoveChild(obj)

	if b.Objects().Len() != 0 {
		t.Errorf("Objects().Len = %d, want 0 after RemoveChild", b.Objects().Len())
	}
	if obj.Parent() != nil {
		t.Error("obj.Parent should be nil after RemoveChild")
	}
}

func TestBandBase_RemoveChild_ChildBand(t *testing.T) {
	b := band.NewBandBase()
	child := band.NewChildBand()
	b.AddChild(child)
	b.RemoveChild(child)

	if b.Child() != nil {
		t.Error("Child should be nil after RemoveChild")
	}
}

func TestBandBase_GetChildObjects(t *testing.T) {
	b := band.NewBandBase()
	obj1 := newMinimalBase("a")
	obj2 := newMinimalBase("b")
	child := band.NewChildBand()

	b.AddChild(obj1)
	b.AddChild(obj2)
	b.AddChild(child)

	var list []report.Base
	b.GetChildObjects(&list)
	if len(list) != 3 {
		t.Errorf("GetChildObjects len = %d, want 3", len(list))
	}
}

func TestBandBase_GetChildOrder(t *testing.T) {
	b := band.NewBandBase()
	obj1 := newMinimalBase("a")
	obj2 := newMinimalBase("b")
	b.AddChild(obj1)
	b.AddChild(obj2)

	if b.GetChildOrder(obj1) != 0 {
		t.Errorf("GetChildOrder(obj1) = %d, want 0", b.GetChildOrder(obj1))
	}
	if b.GetChildOrder(obj2) != 1 {
		t.Errorf("GetChildOrder(obj2) = %d, want 1", b.GetChildOrder(obj2))
	}
}

func TestBandBase_SetChildOrder(t *testing.T) {
	b := band.NewBandBase()
	obj1 := newMinimalBase("a")
	obj2 := newMinimalBase("b")
	b.AddChild(obj1)
	b.AddChild(obj2)

	b.SetChildOrder(obj1, 1)
	if b.Objects().Get(0) != obj2 {
		t.Error("after SetChildOrder, obj2 should be at index 0")
	}
	if b.Objects().Get(1) != obj1 {
		t.Error("after SetChildOrder, obj1 should be at index 1")
	}
}

func TestBandBase_UpdateLayout_NoOp(t *testing.T) {
	b := band.NewBandBase()
	// Should not panic.
	b.UpdateLayout(10, 20)
}

// --- ChildBand ---

func TestNewChildBand(t *testing.T) {
	c := band.NewChildBand()
	if c == nil {
		t.Fatal("NewChildBand returned nil")
	}
	// Inherits BandBase defaults.
	if !c.FirstRowStartsNewPage() {
		t.Error("ChildBand.FirstRowStartsNewPage should default to true")
	}
}
