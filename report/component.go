package report

import "math"

// AnchorStyle specifies which edges of the parent an object is anchored to.
type AnchorStyle int

const (
	// AnchorNone means no anchoring.
	AnchorNone AnchorStyle = 0
	// AnchorLeft anchors to the left edge.
	AnchorLeft AnchorStyle = 1
	// AnchorTop anchors to the top edge.
	AnchorTop AnchorStyle = 2
	// AnchorRight anchors to the right edge.
	AnchorRight AnchorStyle = 4
	// AnchorBottom anchors to the bottom edge.
	AnchorBottom AnchorStyle = 8
	// AnchorDefault is the default anchor (Left+Top).
	AnchorDefault AnchorStyle = AnchorLeft | AnchorTop
)

// DockStyle specifies how an object is docked within its parent.
type DockStyle int

const (
	// DockNone means no docking.
	DockNone DockStyle = iota
	// DockLeft docks to the left edge.
	DockLeft
	// DockTop docks to the top edge.
	DockTop
	// DockRight docks to the right edge.
	DockRight
	// DockBottom docks to the bottom edge.
	DockBottom
	// DockFill fills the entire parent.
	DockFill
)

// Rect represents a bounding rectangle in pixel coordinates.
type Rect struct {
	Left, Top, Width, Height float32
}

// Right returns Left + Width.
func (r Rect) Right() float32 { return r.Left + r.Width }

// Bottom returns Top + Height.
func (r Rect) Bottom() float32 { return r.Top + r.Height }

// ComponentBase provides positioning, sizing, visibility, and docking
// properties for all report components. It is the Go equivalent of
// FastReport.ComponentBase.
//
// Coordinates are stored in screen pixels. Use the units package for
// conversion to/from other measurement systems.
type ComponentBase struct {
	BaseObject

	left   float32
	top    float32
	width  float32
	height float32

	anchor  AnchorStyle
	dock    DockStyle
	visible bool

	visibleExpression   string
	printable           bool
	printableExpression string
	groupIndex          int
}

// NewComponentBase creates a ComponentBase with default values:
// visible=true, printable=true, anchor=AnchorDefault.
func NewComponentBase() *ComponentBase {
	c := &ComponentBase{
		BaseObject: *NewBaseObject(),
		visible:    true,
		printable:  true,
		anchor:     AnchorDefault,
	}
	return c
}

// Left returns the x coordinate relative to the parent, in pixels.
func (c *ComponentBase) Left() float32 { return c.left }

// SetLeft sets the x coordinate, rounded to 2 decimal places.
func (c *ComponentBase) SetLeft(v float32) { c.left = roundF(v) }

// Top returns the y coordinate relative to the parent, in pixels.
func (c *ComponentBase) Top() float32 { return c.top }

// SetTop sets the y coordinate, rounded to 2 decimal places.
func (c *ComponentBase) SetTop(v float32) { c.top = roundF(v) }

// Width returns the width in pixels.
func (c *ComponentBase) Width() float32 { return c.width }

// SetWidth sets the width, rounded to 2 decimal places.
func (c *ComponentBase) SetWidth(v float32) { c.width = roundF(v) }

// Height returns the height in pixels.
func (c *ComponentBase) Height() float32 { return c.height }

// SetHeight sets the height, rounded to 2 decimal places.
func (c *ComponentBase) SetHeight(v float32) { c.height = roundF(v) }

// Right returns Left + Width.
func (c *ComponentBase) Right() float32 { return c.left + c.width }

// Bottom returns Top + Height.
func (c *ComponentBase) Bottom() float32 { return c.top + c.height }

// Bounds returns the bounding Rect.
func (c *ComponentBase) Bounds() Rect {
	return Rect{Left: c.left, Top: c.top, Width: c.width, Height: c.height}
}

// SetBounds sets all four layout properties at once.
func (c *ComponentBase) SetBounds(r Rect) {
	c.left = roundF(r.Left)
	c.top = roundF(r.Top)
	c.width = roundF(r.Width)
	c.height = roundF(r.Height)
}

// Positioned is satisfied by any object that reports its absolute screen
// coordinates. Used by AbsLeft/AbsTop to traverse the parent chain.
type Positioned interface {
	AbsLeft() float32
	AbsTop() float32
}

// AbsLeft returns the absolute x coordinate (accumulated from all parents).
func (c *ComponentBase) AbsLeft() float32 {
	if p, ok := c.parent.(Positioned); ok {
		return c.left + p.AbsLeft()
	}
	return c.left
}

// AbsTop returns the absolute y coordinate (accumulated from all parents).
func (c *ComponentBase) AbsTop() float32 {
	if p, ok := c.parent.(Positioned); ok {
		return c.top + p.AbsTop()
	}
	return c.top
}

// AbsRight returns AbsLeft + Width.
func (c *ComponentBase) AbsRight() float32 { return c.AbsLeft() + c.width }

// AbsBottom returns AbsTop + Height.
func (c *ComponentBase) AbsBottom() float32 { return c.AbsTop() + c.height }

// Visible returns whether the component is visible.
func (c *ComponentBase) Visible() bool { return c.visible }

// SetVisible sets the visibility.
func (c *ComponentBase) SetVisible(v bool) { c.visible = v }

// VisibleExpression returns the expression controlling visibility.
func (c *ComponentBase) VisibleExpression() string { return c.visibleExpression }

// SetVisibleExpression sets the visibility expression.
func (c *ComponentBase) SetVisibleExpression(expr string) { c.visibleExpression = expr }

// Printable returns whether the component is printed in the output.
func (c *ComponentBase) Printable() bool { return c.printable }

// SetPrintable sets the printable flag.
func (c *ComponentBase) SetPrintable(v bool) { c.printable = v }

// PrintableExpression returns the expression controlling printability.
func (c *ComponentBase) PrintableExpression() string { return c.printableExpression }

// SetPrintableExpression sets the printable expression.
func (c *ComponentBase) SetPrintableExpression(expr string) { c.printableExpression = expr }

// Anchor returns the anchor style.
func (c *ComponentBase) Anchor() AnchorStyle { return c.anchor }

// SetAnchor sets the anchor style.
func (c *ComponentBase) SetAnchor(a AnchorStyle) { c.anchor = a }

// Dock returns the dock style.
func (c *ComponentBase) Dock() DockStyle { return c.dock }

// SetDock sets the dock style.
func (c *ComponentBase) SetDock(d DockStyle) { c.dock = d }

// GroupIndex returns the group index (for designer grouping).
func (c *ComponentBase) GroupIndex() int { return c.groupIndex }

// SetGroupIndex sets the group index.
func (c *ComponentBase) SetGroupIndex(idx int) { c.groupIndex = idx }

// Serialize writes ComponentBase properties that differ from defaults.
func (c *ComponentBase) Serialize(w Writer) error {
	if err := c.BaseObject.Serialize(w); err != nil {
		return err
	}
	if c.left != 0 {
		w.WriteFloat("Left", c.left)
	}
	if c.top != 0 {
		w.WriteFloat("Top", c.top)
	}
	if c.width != 0 {
		w.WriteFloat("Width", c.width)
	}
	if c.height != 0 {
		w.WriteFloat("Height", c.height)
	}
	if !c.visible {
		w.WriteBool("Visible", false)
	}
	if !c.printable {
		w.WriteBool("Printable", false)
	}
	if c.visibleExpression != "" {
		w.WriteStr("VisibleExpression", c.visibleExpression)
	}
	if c.printableExpression != "" {
		w.WriteStr("PrintableExpression", c.printableExpression)
	}
	if c.anchor != AnchorDefault {
		w.WriteInt("Anchor", int(c.anchor))
	}
	if c.dock != DockNone {
		w.WriteInt("Dock", int(c.dock))
	}
	if c.groupIndex != 0 {
		w.WriteInt("GroupIndex", c.groupIndex)
	}
	return nil
}

// Deserialize reads ComponentBase properties.
func (c *ComponentBase) Deserialize(r Reader) error {
	if err := c.BaseObject.Deserialize(r); err != nil {
		return err
	}
	c.left = r.ReadFloat("Left", 0)
	c.top = r.ReadFloat("Top", 0)
	c.width = r.ReadFloat("Width", 0)
	c.height = r.ReadFloat("Height", 0)
	c.visible = r.ReadBool("Visible", true)
	c.printable = r.ReadBool("Printable", true)
	c.visibleExpression = r.ReadStr("VisibleExpression", "")
	c.printableExpression = r.ReadStr("PrintableExpression", "")
	c.anchor = AnchorStyle(r.ReadInt("Anchor", int(AnchorDefault)))
	c.dock = DockStyle(r.ReadInt("Dock", int(DockNone)))
	c.groupIndex = r.ReadInt("GroupIndex", 0)
	return nil
}

// roundF rounds a float32 to 2 decimal places.
func roundF(v float32) float32 {
	return float32(math.Round(float64(v)*100) / 100)
}
