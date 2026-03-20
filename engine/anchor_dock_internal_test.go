package engine

// anchor_dock_internal_test.go — internal tests (package engine) for
// applyAnchorAdjustments (engine/bands.go) and applyDockLayout (engine/dock.go).
//
// Both functions are unexported and operate on band/object internals, so they
// must be tested from within the engine package.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
)

// ── applyAnchorAdjustments ────────────────────────────────────────────────────

// makeAnchorBand creates a BandBase with a single visible TextObject at the
// given position, sets its Anchor, adds a matching PreparedObject at startIdx
// inside pb, and returns (bb, pb).
func makeAnchorBand(left, top, w, h float32, anchor report.AnchorStyle) (*band.BandBase, *preview.PreparedBand) {
	bb := band.NewBandBase()
	bb.SetName("AnchorBand")
	bb.SetWidth(200)
	bb.SetHeight(100)

	txt := object.NewTextObject()
	txt.SetName("T1")
	txt.SetLeft(left)
	txt.SetTop(top)
	txt.SetWidth(w)
	txt.SetHeight(h)
	txt.SetVisible(true)
	txt.SetAnchor(anchor)
	bb.Objects().Add(txt)

	pb := &preview.PreparedBand{
		Name:   "AnchorBand",
		Width:  200,
		Height: 100,
		Objects: []preview.PreparedObject{
			{Name: "T1", Left: left, Top: top, Width: w, Height: h},
		},
	}
	return bb, pb
}

// TestApplyAnchorAdjustments_NoOp_ZeroDeltas verifies that when both deltas are
// zero, PreparedObject coordinates are unchanged.
func TestApplyAnchorAdjustments_NoOp_ZeroDeltas(t *testing.T) {
	bb, pb := makeAnchorBand(10, 20, 50, 30, report.AnchorBottom)
	origTop := pb.Objects[0].Top

	applyAnchorAdjustments(bb, pb, 0, 0, 0)

	if pb.Objects[0].Top != origTop {
		t.Errorf("zero delta: Top changed from %v to %v", origTop, pb.Objects[0].Top)
	}
}

// TestApplyAnchorAdjustments_NoOp_NilBand verifies that a nil BandBase causes
// an immediate return without panicking.
func TestApplyAnchorAdjustments_NoOp_NilBand(t *testing.T) {
	pb := &preview.PreparedBand{
		Objects: []preview.PreparedObject{{Name: "T1", Top: 20}},
	}
	// Should not panic.
	applyAnchorAdjustments(nil, pb, 0, 10, 0)
}

// TestApplyAnchorAdjustments_AnchorBottom_MovesDown verifies that an object
// anchored only to the bottom edge moves down by deltaH when the band grows.
func TestApplyAnchorAdjustments_AnchorBottom_MovesDown(t *testing.T) {
	// AnchorBottom only (no AnchorTop) → Top should increase by deltaH.
	const (
		origTop float32 = 60
		deltaH  float32 = 20
	)
	bb, pb := makeAnchorBand(10, origTop, 50, 30, report.AnchorBottom)

	applyAnchorAdjustments(bb, pb, 0, deltaH, 0)

	wantTop := origTop + deltaH
	if pb.Objects[0].Top != wantTop {
		t.Errorf("AnchorBottom: Top = %v, want %v", pb.Objects[0].Top, wantTop)
	}
	// Height must be unchanged.
	if pb.Objects[0].Height != 30 {
		t.Errorf("AnchorBottom: Height = %v, want 30 (unchanged)", pb.Objects[0].Height)
	}
}

// TestApplyAnchorAdjustments_AnchorTop_NoChange verifies that an object
// anchored only to the top edge is unaffected by a vertical delta.
func TestApplyAnchorAdjustments_AnchorTop_NoChange(t *testing.T) {
	const (
		origTop float32 = 10
		deltaH  float32 = 20
	)
	bb, pb := makeAnchorBand(0, origTop, 50, 30, report.AnchorTop)

	applyAnchorAdjustments(bb, pb, 0, deltaH, 0)

	if pb.Objects[0].Top != origTop {
		t.Errorf("AnchorTop: Top = %v, want %v (unchanged)", pb.Objects[0].Top, origTop)
	}
}

// TestApplyAnchorAdjustments_AnchorTopAndBottom_StretchesHeight verifies that
// an object anchored to both top and bottom has its height increased by deltaH.
func TestApplyAnchorAdjustments_AnchorTopAndBottom_StretchesHeight(t *testing.T) {
	const (
		origTop    float32 = 10
		origHeight float32 = 30
		deltaH     float32 = 20
	)
	anchor := report.AnchorTop | report.AnchorBottom
	bb, pb := makeAnchorBand(0, origTop, 50, origHeight, anchor)

	applyAnchorAdjustments(bb, pb, 0, deltaH, 0)

	// Top should be unchanged.
	if pb.Objects[0].Top != origTop {
		t.Errorf("AnchorTop|Bottom: Top = %v, want %v (unchanged)", pb.Objects[0].Top, origTop)
	}
	// Height should grow by deltaH.
	wantHeight := origHeight + deltaH
	if pb.Objects[0].Height != wantHeight {
		t.Errorf("AnchorTop|Bottom: Height = %v, want %v", pb.Objects[0].Height, wantHeight)
	}
}

// TestApplyAnchorAdjustments_AnchorRight_MovesRight verifies that an object
// anchored only to the right edge moves right by deltaW when the band grows.
func TestApplyAnchorAdjustments_AnchorRight_MovesRight(t *testing.T) {
	const (
		origLeft float32 = 140
		deltaW   float32 = 30
	)
	// Use AnchorRight only (no AnchorLeft) so Left shifts right.
	bb, pb := makeAnchorBand(origLeft, 10, 50, 30, report.AnchorRight)

	applyAnchorAdjustments(bb, pb, 0, 0, deltaW)

	wantLeft := origLeft + deltaW
	if pb.Objects[0].Left != wantLeft {
		t.Errorf("AnchorRight: Left = %v, want %v", pb.Objects[0].Left, wantLeft)
	}
	// Width must be unchanged.
	if pb.Objects[0].Width != 50 {
		t.Errorf("AnchorRight: Width = %v, want 50 (unchanged)", pb.Objects[0].Width)
	}
}

// TestApplyAnchorAdjustments_AnchorLeft_NoChange verifies that an object
// anchored only to the left edge is unaffected by a horizontal delta.
func TestApplyAnchorAdjustments_AnchorLeft_NoChange(t *testing.T) {
	const (
		origLeft float32 = 10
		deltaW   float32 = 30
	)
	bb, pb := makeAnchorBand(origLeft, 10, 50, 30, report.AnchorLeft)

	applyAnchorAdjustments(bb, pb, 0, 0, deltaW)

	if pb.Objects[0].Left != origLeft {
		t.Errorf("AnchorLeft: Left = %v, want %v (unchanged)", pb.Objects[0].Left, origLeft)
	}
}

// TestApplyAnchorAdjustments_AnchorLeftAndRight_StretchesWidth verifies that
// an object anchored to both left and right has its width increased by deltaW.
func TestApplyAnchorAdjustments_AnchorLeftAndRight_StretchesWidth(t *testing.T) {
	const (
		origLeft  float32 = 10
		origWidth float32 = 80
		deltaW    float32 = 30
	)
	anchor := report.AnchorLeft | report.AnchorRight
	bb, pb := makeAnchorBand(origLeft, 10, origWidth, 30, anchor)

	applyAnchorAdjustments(bb, pb, 0, 0, deltaW)

	// Left should be unchanged.
	if pb.Objects[0].Left != origLeft {
		t.Errorf("AnchorLeft|Right: Left = %v, want %v (unchanged)", pb.Objects[0].Left, origLeft)
	}
	// Width should grow by deltaW.
	wantWidth := origWidth + deltaW
	if pb.Objects[0].Width != wantWidth {
		t.Errorf("AnchorLeft|Right: Width = %v, want %v", pb.Objects[0].Width, wantWidth)
	}
}

// TestApplyAnchorAdjustments_AnchorNone_NoChange verifies that AnchorNone
// objects are unaffected by either delta.
func TestApplyAnchorAdjustments_AnchorNone_NoChange(t *testing.T) {
	const (
		origLeft float32 = 10
		origTop  float32 = 20
	)
	bb, pb := makeAnchorBand(origLeft, origTop, 50, 30, report.AnchorNone)

	applyAnchorAdjustments(bb, pb, 0, 15, 25)

	if pb.Objects[0].Left != origLeft {
		t.Errorf("AnchorNone: Left = %v, want %v", pb.Objects[0].Left, origLeft)
	}
	if pb.Objects[0].Top != origTop {
		t.Errorf("AnchorNone: Top = %v, want %v", pb.Objects[0].Top, origTop)
	}
}

// TestApplyAnchorAdjustments_InvisibleObjectSkipped verifies that an invisible
// object (which buildPreparedObject returns nil for) is not counted against
// the PreparedObject index.
func TestApplyAnchorAdjustments_InvisibleObjectSkipped(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetName("MixedBand")
	bb.SetWidth(200)
	bb.SetHeight(100)

	// First object: invisible — skipped by applyAnchorAdjustments.
	invisible := object.NewTextObject()
	invisible.SetName("Inv")
	invisible.SetLeft(0)
	invisible.SetTop(10)
	invisible.SetWidth(50)
	invisible.SetHeight(20)
	invisible.SetVisible(false)
	invisible.SetAnchor(report.AnchorBottom) // would shift if counted

	// Second object: visible, AnchorBottom — should match pb.Objects[0].
	visible := object.NewTextObject()
	visible.SetName("Vis")
	visible.SetLeft(0)
	visible.SetTop(60)
	visible.SetWidth(50)
	visible.SetHeight(20)
	visible.SetVisible(true)
	visible.SetAnchor(report.AnchorBottom)

	bb.Objects().Add(invisible)
	bb.Objects().Add(visible)

	// pb has only one PreparedObject (for the visible object).
	pb := &preview.PreparedBand{
		Objects: []preview.PreparedObject{
			{Name: "Vis", Left: 0, Top: 60, Width: 50, Height: 20},
		},
	}

	const deltaH float32 = 10
	applyAnchorAdjustments(bb, pb, 0, deltaH, 0)

	// The visible object (AnchorBottom) must shift down.
	wantTop := float32(60) + deltaH
	if pb.Objects[0].Top != wantTop {
		t.Errorf("visible AnchorBottom after skip: Top = %v, want %v", pb.Objects[0].Top, wantTop)
	}
}

// TestApplyAnchorAdjustments_EmptyObjects verifies that a band with no objects
// returns immediately without panicking.
func TestApplyAnchorAdjustments_EmptyObjects(t *testing.T) {
	bb := band.NewBandBase()
	pb := &preview.PreparedBand{}
	applyAnchorAdjustments(bb, pb, 0, 10, 10) // should not panic
}

// ── applyDockLayout ───────────────────────────────────────────────────────────

// makeDockObject creates a TextObject with the given dock style and initial geometry.
func makeDockObject(dock report.DockStyle, left, top, w, h float32) *object.TextObject {
	txt := object.NewTextObject()
	txt.SetLeft(left)
	txt.SetTop(top)
	txt.SetWidth(w)
	txt.SetHeight(h)
	txt.SetDock(dock)
	return txt
}

// TestApplyDockLayout_NilCollection verifies that a nil collection returns
// immediately without panicking.
func TestApplyDockLayout_NilCollection(t *testing.T) {
	applyDockLayout(nil, 200, 100) // should not panic
}

// TestApplyDockLayout_EmptyCollection verifies that an empty collection is a
// no-op.
func TestApplyDockLayout_EmptyCollection(t *testing.T) {
	coll := report.NewObjectCollection()
	applyDockLayout(coll, 200, 100) // should not panic
}

// TestApplyDockLayout_DockNone_Unchanged verifies that DockNone objects are
// left untouched.
func TestApplyDockLayout_DockNone_Unchanged(t *testing.T) {
	coll := report.NewObjectCollection()
	obj := makeDockObject(report.DockNone, 20, 30, 60, 40)
	coll.Add(obj)

	applyDockLayout(coll, 200, 100)

	if obj.Left() != 20 || obj.Top() != 30 || obj.Width() != 60 || obj.Height() != 40 {
		t.Errorf("DockNone: got (%v,%v,%v,%v), want (20,30,60,40)",
			obj.Left(), obj.Top(), obj.Width(), obj.Height())
	}
}

// TestApplyDockLayout_DockTop positions the object at the top of the container,
// stretching its width to the full container width while keeping height.
func TestApplyDockLayout_DockTop(t *testing.T) {
	coll := report.NewObjectCollection()
	obj := makeDockObject(report.DockTop, 999, 999, 50, 20)
	coll.Add(obj)

	applyDockLayout(coll, 200, 100)

	if obj.Left() != 0 {
		t.Errorf("DockTop: Left = %v, want 0", obj.Left())
	}
	if obj.Top() != 0 {
		t.Errorf("DockTop: Top = %v, want 0", obj.Top())
	}
	if obj.Width() != 200 {
		t.Errorf("DockTop: Width = %v, want 200", obj.Width())
	}
	// Height is kept as-is.
	if obj.Height() != 20 {
		t.Errorf("DockTop: Height = %v, want 20 (kept)", obj.Height())
	}
}

// TestApplyDockLayout_DockBottom positions the object at the bottom of the
// container, stretching its width while keeping height.
func TestApplyDockLayout_DockBottom(t *testing.T) {
	coll := report.NewObjectCollection()
	obj := makeDockObject(report.DockBottom, 999, 999, 50, 25)
	coll.Add(obj)

	applyDockLayout(coll, 200, 100)

	if obj.Left() != 0 {
		t.Errorf("DockBottom: Left = %v, want 0", obj.Left())
	}
	// Top = containerH - height = 100 - 25 = 75.
	if obj.Top() != 75 {
		t.Errorf("DockBottom: Top = %v, want 75", obj.Top())
	}
	if obj.Width() != 200 {
		t.Errorf("DockBottom: Width = %v, want 200", obj.Width())
	}
	if obj.Height() != 25 {
		t.Errorf("DockBottom: Height = %v, want 25 (kept)", obj.Height())
	}
}

// TestApplyDockLayout_DockLeft positions the object on the left edge, stretching
// its height to the full container height while keeping width.
func TestApplyDockLayout_DockLeft(t *testing.T) {
	coll := report.NewObjectCollection()
	obj := makeDockObject(report.DockLeft, 999, 999, 40, 999)
	coll.Add(obj)

	applyDockLayout(coll, 200, 100)

	if obj.Left() != 0 {
		t.Errorf("DockLeft: Left = %v, want 0", obj.Left())
	}
	if obj.Top() != 0 {
		t.Errorf("DockLeft: Top = %v, want 0", obj.Top())
	}
	// Width is kept as-is.
	if obj.Width() != 40 {
		t.Errorf("DockLeft: Width = %v, want 40 (kept)", obj.Width())
	}
	if obj.Height() != 100 {
		t.Errorf("DockLeft: Height = %v, want 100", obj.Height())
	}
}

// TestApplyDockLayout_DockRight positions the object on the right edge,
// stretching its height while keeping width.
func TestApplyDockLayout_DockRight(t *testing.T) {
	coll := report.NewObjectCollection()
	obj := makeDockObject(report.DockRight, 999, 999, 40, 999)
	coll.Add(obj)

	applyDockLayout(coll, 200, 100)

	// Left = containerW - width = 200 - 40 = 160.
	if obj.Left() != 160 {
		t.Errorf("DockRight: Left = %v, want 160", obj.Left())
	}
	if obj.Top() != 0 {
		t.Errorf("DockRight: Top = %v, want 0", obj.Top())
	}
	if obj.Width() != 40 {
		t.Errorf("DockRight: Width = %v, want 40 (kept)", obj.Width())
	}
	if obj.Height() != 100 {
		t.Errorf("DockRight: Height = %v, want 100", obj.Height())
	}
}

// TestApplyDockLayout_DockFill fills the entire container rect.
func TestApplyDockLayout_DockFill(t *testing.T) {
	coll := report.NewObjectCollection()
	obj := makeDockObject(report.DockFill, 999, 999, 1, 1)
	coll.Add(obj)

	applyDockLayout(coll, 200, 100)

	if obj.Left() != 0 {
		t.Errorf("DockFill: Left = %v, want 0", obj.Left())
	}
	if obj.Top() != 0 {
		t.Errorf("DockFill: Top = %v, want 0", obj.Top())
	}
	if obj.Width() != 200 {
		t.Errorf("DockFill: Width = %v, want 200", obj.Width())
	}
	if obj.Height() != 100 {
		t.Errorf("DockFill: Height = %v, want 100", obj.Height())
	}
}

// TestApplyDockLayout_DockFill_ConsumesRemainingRect verifies that a second
// DockFill object after the first receives zero remaining space.
func TestApplyDockLayout_DockFill_ConsumesRemainingRect(t *testing.T) {
	coll := report.NewObjectCollection()
	first := makeDockObject(report.DockFill, 0, 0, 10, 10)
	second := makeDockObject(report.DockFill, 0, 0, 50, 50)
	coll.Add(first)
	coll.Add(second)

	applyDockLayout(coll, 200, 100)

	// First fills everything.
	if first.Width() != 200 || first.Height() != 100 {
		t.Errorf("DockFill first: got (%v,%v), want (200,100)", first.Width(), first.Height())
	}
	// Second gets the remaining rect which is zero-sized (remLeft==remRight, remTop==remBottom).
	if second.Width() != 0 {
		t.Errorf("DockFill second: Width = %v, want 0", second.Width())
	}
	if second.Height() != 0 {
		t.Errorf("DockFill second: Height = %v, want 0", second.Height())
	}
}

// TestApplyDockLayout_TopThenBottom_ReducesRemainingRect verifies that DockTop
// followed by DockBottom each consume from the remaining rect.
func TestApplyDockLayout_TopThenBottom_ReducesRemainingRect(t *testing.T) {
	coll := report.NewObjectCollection()
	top := makeDockObject(report.DockTop, 0, 0, 0, 20)    // 20px from top
	bottom := makeDockObject(report.DockBottom, 0, 0, 0, 15) // 15px from bottom
	coll.Add(top)
	coll.Add(bottom)

	const (
		containerW float32 = 200
		containerH float32 = 100
	)
	applyDockLayout(coll, containerW, containerH)

	// DockTop: placed at top, width stretched.
	if top.Left() != 0 || top.Top() != 0 || top.Width() != containerW {
		t.Errorf("DockTop: got (%v,%v,%v), want (0,0,200)", top.Left(), top.Top(), top.Width())
	}

	// DockBottom: remaining top is 20 (after DockTop consumed 20px).
	// remBottom is still 100 (DockTop only advances remTop).
	// So DockBottom Top = remBottom - height = 100 - 15 = 85.
	if bottom.Top() != 85 {
		t.Errorf("DockBottom after DockTop: Top = %v, want 85", bottom.Top())
	}
	if bottom.Width() != containerW {
		t.Errorf("DockBottom: Width = %v, want 200", bottom.Width())
	}
}

// TestApplyDockLayout_LeftThenRight_ReducesRemainingRect verifies that DockLeft
// followed by DockRight each consume from the remaining rect.
func TestApplyDockLayout_LeftThenRight_ReducesRemainingRect(t *testing.T) {
	coll := report.NewObjectCollection()
	left := makeDockObject(report.DockLeft, 0, 0, 30, 0)  // 30px from left
	right := makeDockObject(report.DockRight, 0, 0, 25, 0) // 25px from right
	coll.Add(left)
	coll.Add(right)

	const (
		containerW float32 = 200
		containerH float32 = 100
	)
	applyDockLayout(coll, containerW, containerH)

	// DockLeft: at x=0, height stretched.
	if left.Left() != 0 || left.Top() != 0 || left.Height() != containerH {
		t.Errorf("DockLeft: got (%v,%v,h=%v), want (0,0,h=100)", left.Left(), left.Top(), left.Height())
	}

	// DockRight: remLeft is now 30, remRight is 200.
	// Right object Left = remRight - width = 200 - 25 = 175.
	if right.Left() != 175 {
		t.Errorf("DockRight after DockLeft: Left = %v, want 175", right.Left())
	}
	if right.Height() != containerH {
		t.Errorf("DockRight: Height = %v, want 100", right.Height())
	}
}

// TestApplyDockLayout_NonDockableObject_Skipped verifies that objects which do
// not implement the dockable interface are skipped without panicking.
// report.NewBaseObject() does not embed ComponentBase so it is not dockable.
func TestApplyDockLayout_NonDockableObject_Skipped(t *testing.T) {
	coll := report.NewObjectCollection()
	// BaseObject does not satisfy the dockable interface (no Left/Top/Width/Height/SetLeft etc.).
	coll.Add(report.NewBaseObject())

	// Should not panic.
	applyDockLayout(coll, 200, 100)
}
