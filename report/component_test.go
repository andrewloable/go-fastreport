package report_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

func TestAnchorStyleConstants(t *testing.T) {
	if report.AnchorLeft != 1 {
		t.Errorf("AnchorLeft = %d, want 1", report.AnchorLeft)
	}
	if report.AnchorTop != 2 {
		t.Errorf("AnchorTop = %d, want 2", report.AnchorTop)
	}
	if report.AnchorRight != 4 {
		t.Errorf("AnchorRight = %d, want 4", report.AnchorRight)
	}
	if report.AnchorBottom != 8 {
		t.Errorf("AnchorBottom = %d, want 8", report.AnchorBottom)
	}
	if report.AnchorDefault != report.AnchorLeft|report.AnchorTop {
		t.Errorf("AnchorDefault = %d, want %d", report.AnchorDefault, report.AnchorLeft|report.AnchorTop)
	}
}

func TestDockStyleConstants(t *testing.T) {
	if report.DockNone != 0 {
		t.Errorf("DockNone = %d, want 0", report.DockNone)
	}
}

func TestRectMethods(t *testing.T) {
	r := report.Rect{Left: 10, Top: 20, Width: 100, Height: 50}
	if r.Right() != 110 {
		t.Errorf("Right() = %v, want 110", r.Right())
	}
	if r.Bottom() != 70 {
		t.Errorf("Bottom() = %v, want 70", r.Bottom())
	}
}

func TestNewComponentBase(t *testing.T) {
	c := report.NewComponentBase()
	if !c.Visible() {
		t.Error("default Visible should be true")
	}
	if !c.Printable() {
		t.Error("default Printable should be true")
	}
	if c.Anchor() != report.AnchorDefault {
		t.Errorf("default Anchor = %v, want AnchorDefault", c.Anchor())
	}
	if c.Dock() != report.DockNone {
		t.Errorf("default Dock = %v, want DockNone", c.Dock())
	}
}

func TestComponentSetGetLeftTop(t *testing.T) {
	c := report.NewComponentBase()
	c.SetLeft(10.5)
	c.SetTop(20.75)
	if c.Left() != 10.5 {
		t.Errorf("Left() = %v, want 10.5", c.Left())
	}
	if c.Top() != 20.75 {
		t.Errorf("Top() = %v, want 20.75", c.Top())
	}
}

func TestComponentSetGetWidthHeight(t *testing.T) {
	c := report.NewComponentBase()
	c.SetWidth(200)
	c.SetHeight(150)
	if c.Width() != 200 {
		t.Errorf("Width() = %v, want 200", c.Width())
	}
	if c.Height() != 150 {
		t.Errorf("Height() = %v, want 150", c.Height())
	}
}

func TestComponentRightBottom(t *testing.T) {
	c := report.NewComponentBase()
	c.SetLeft(10)
	c.SetTop(20)
	c.SetWidth(100)
	c.SetHeight(50)
	if c.Right() != 110 {
		t.Errorf("Right() = %v, want 110", c.Right())
	}
	if c.Bottom() != 70 {
		t.Errorf("Bottom() = %v, want 70", c.Bottom())
	}
}

func TestComponentBounds(t *testing.T) {
	c := report.NewComponentBase()
	c.SetBounds(report.Rect{Left: 5, Top: 10, Width: 200, Height: 100})
	b := c.Bounds()
	if b.Left != 5 || b.Top != 10 || b.Width != 200 || b.Height != 100 {
		t.Errorf("Bounds = %+v, want {5 10 200 100}", b)
	}
}

func TestComponentAbsCoordinates(t *testing.T) {
	c := report.NewComponentBase()
	c.SetLeft(50)
	c.SetTop(100)
	// Without parent, absolute == relative
	if c.AbsLeft() != 50 {
		t.Errorf("AbsLeft() = %v, want 50", c.AbsLeft())
	}
	if c.AbsTop() != 100 {
		t.Errorf("AbsTop() = %v, want 100", c.AbsTop())
	}
	if c.AbsRight() != 50 {
		t.Errorf("AbsRight() (no width) = %v, want 50", c.AbsRight())
	}
	if c.AbsBottom() != 100 {
		t.Errorf("AbsBottom() (no height) = %v, want 100", c.AbsBottom())
	}
}

// positionedParent is a ComponentBase that also implements report.Parent so
// it can be used as both a positioned ancestor and a parent container.
type positionedParent struct {
	*report.ComponentBase
	children []report.Base
}

func newPositionedParent() *positionedParent {
	return &positionedParent{ComponentBase: report.NewComponentBase()}
}
func (p *positionedParent) CanContain(report.Base) bool { return true }
func (p *positionedParent) GetChildObjects(list *[]report.Base) {
	*list = append(*list, p.children...)
}
func (p *positionedParent) AddChild(child report.Base) {
	p.children = append(p.children, child)
	child.SetParent(p)
}
func (p *positionedParent) RemoveChild(child report.Base) {}
func (p *positionedParent) GetChildOrder(child report.Base) int { return 0 }
func (p *positionedParent) SetChildOrder(child report.Base, order int) {}
func (p *positionedParent) UpdateLayout(dx, dy float32) {}

func TestComponentAbsCoordinatesWithParent(t *testing.T) {
	parent := newPositionedParent()
	parent.SetLeft(100)
	parent.SetTop(200)

	child := report.NewComponentBase()
	child.SetLeft(10)
	child.SetTop(20)
	child.SetParent(parent)

	if child.AbsLeft() != 110 {
		t.Errorf("child AbsLeft() = %v, want 110", child.AbsLeft())
	}
	if child.AbsTop() != 220 {
		t.Errorf("child AbsTop() = %v, want 220", child.AbsTop())
	}
}

func TestComponentVisible(t *testing.T) {
	c := report.NewComponentBase()
	c.SetVisible(false)
	if c.Visible() {
		t.Error("Visible should be false")
	}
	c.SetVisible(true)
	if !c.Visible() {
		t.Error("Visible should be true")
	}
}

func TestComponentPrintable(t *testing.T) {
	c := report.NewComponentBase()
	c.SetPrintable(false)
	if c.Printable() {
		t.Error("Printable should be false")
	}
}

func TestComponentExpressions(t *testing.T) {
	c := report.NewComponentBase()
	c.SetVisibleExpression("[Field] > 0")
	if c.VisibleExpression() != "[Field] > 0" {
		t.Errorf("VisibleExpression = %q", c.VisibleExpression())
	}
	c.SetPrintableExpression("[PageNumber] > 1")
	if c.PrintableExpression() != "[PageNumber] > 1" {
		t.Errorf("PrintableExpression = %q", c.PrintableExpression())
	}
}

func TestComponentAnchorDock(t *testing.T) {
	c := report.NewComponentBase()
	c.SetAnchor(report.AnchorLeft | report.AnchorRight)
	if c.Anchor() != report.AnchorLeft|report.AnchorRight {
		t.Errorf("Anchor = %v", c.Anchor())
	}
	c.SetDock(report.DockFill)
	if c.Dock() != report.DockFill {
		t.Errorf("Dock = %v", c.Dock())
	}
}

func TestComponentGroupIndex(t *testing.T) {
	c := report.NewComponentBase()
	c.SetGroupIndex(5)
	if c.GroupIndex() != 5 {
		t.Errorf("GroupIndex = %d, want 5", c.GroupIndex())
	}
}

func TestComponentRounding(t *testing.T) {
	c := report.NewComponentBase()
	// Values should be rounded to 2 decimal places
	c.SetLeft(10.123)
	if c.Left() != 10.12 {
		t.Errorf("Left rounding: got %v, want 10.12", c.Left())
	}
	c.SetTop(5.999)
	if c.Top() != 6.0 {
		t.Errorf("Top rounding: got %v, want 6.0", c.Top())
	}
}

func TestComponentSerializeDeserialize(t *testing.T) {
	c := report.NewComponentBase()
	c.SetLeft(10)
	c.SetTop(20)
	c.SetWidth(200)
	c.SetHeight(100)
	c.SetVisible(false)
	c.SetPrintable(false)
	c.SetVisibleExpression("[x] > 0")
	c.SetPrintableExpression("[y] > 0")
	c.SetAnchor(report.AnchorRight)
	c.SetDock(report.DockTop)
	c.SetGroupIndex(3)

	w := newMockWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}

	c2 := report.NewComponentBase()
	r := newMockReader()
	for k, v := range w.strings {
		r.strings[k] = v
	}
	for k, v := range w.ints {
		r.ints[k] = v
	}
	for k, v := range w.bools {
		r.bools[k] = v
	}
	for k, v := range w.floats {
		r.floats[k] = v
	}
	if err := c2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}

	if c2.Left() != 10 {
		t.Errorf("Left = %v, want 10", c2.Left())
	}
	if c2.Top() != 20 {
		t.Errorf("Top = %v, want 20", c2.Top())
	}
	if c2.Width() != 200 {
		t.Errorf("Width = %v, want 200", c2.Width())
	}
	if c2.Height() != 100 {
		t.Errorf("Height = %v, want 100", c2.Height())
	}
	if c2.Visible() {
		t.Error("Visible should be false")
	}
	if c2.Printable() {
		t.Error("Printable should be false")
	}
	if c2.VisibleExpression() != "[x] > 0" {
		t.Errorf("VisibleExpression = %q", c2.VisibleExpression())
	}
	if c2.Anchor() != report.AnchorRight {
		t.Errorf("Anchor = %v, want AnchorRight", c2.Anchor())
	}
	if c2.Dock() != report.DockTop {
		t.Errorf("Dock = %v, want DockTop", c2.Dock())
	}
	if c2.GroupIndex() != 3 {
		t.Errorf("GroupIndex = %d, want 3", c2.GroupIndex())
	}
}

func TestComponentDeserializeDefaults(t *testing.T) {
	// When reader returns defaults, component should have default values
	c := report.NewComponentBase()
	r := newMockReader()
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if !c.Visible() {
		t.Error("default Visible from deserialize should be true")
	}
	if !c.Printable() {
		t.Error("default Printable from deserialize should be true")
	}
	if c.Anchor() != report.AnchorDefault {
		t.Errorf("default Anchor = %v, want AnchorDefault", c.Anchor())
	}
}

func TestComponentSerializePrintableExpression(t *testing.T) {
	c := report.NewComponentBase()
	c.SetPrintableExpression("[PageNo] > 1")
	w := newMockWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if w.strings["PrintableExpression"] != "[PageNo] > 1" {
		t.Errorf("PrintableExpression not serialized: %v", w.strings)
	}
}

func TestComponentSerializeDefaults(t *testing.T) {
	// When values are default, they should not be serialized
	c := report.NewComponentBase()
	w := newMockWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	// Default visible=true should not be written (to keep delta serialization)
	if _, ok := w.bools["Visible"]; ok {
		t.Error("Visible=true should not be serialized (is default)")
	}
	if _, ok := w.floats["Left"]; ok {
		t.Error("Left=0 should not be serialized (is default)")
	}
}
