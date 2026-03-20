package engine

// band_layout_internal_test.go — internal (package engine) tests for:
//   - calcBandLayout (bands.go line 121): height computation with CanGrow/CanShrink TextObjects
//   - applyBandObjectShifts (bands.go line 724): shift propagation after grow/shrink
//   - applyBandObjectHeights (bands.go line 781): height adjustment for CanGrow/CanShrink objects

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// buildLayoutEngine creates a minimal engine with one page and runs it.
func buildLayoutEngine(t *testing.T) *ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("buildLayoutEngine: Run: %v", err)
	}
	return e
}

// ── calcBandLayout ────────────────────────────────────────────────────────────

// TestCalcBandLayout_EmptyBand verifies the empty-objects early-return path:
// when a band has no children, layout height equals the declared base height.
func TestCalcBandLayout_EmptyBand(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(50)

	layout := calcBandLayout(bb, 50, nil)

	if layout.height != 50 {
		t.Errorf("empty band: height = %v, want 50", layout.height)
	}
	if layout.shifts != nil {
		t.Error("empty band: shifts should be nil")
	}
	if layout.effectiveH != nil {
		t.Error("empty band: effectiveH should be nil")
	}
}

// TestCalcBandLayout_NonTextObject verifies that a non-TextObject child
// contributes to maxBottom via its Top+Height but does not trigger grow/shrink.
func TestCalcBandLayout_NonTextObject(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(10)

	// A ChildBand has Top()/Height() but is not a TextObject.
	child := band.NewChildBand()
	child.SetTop(0)
	child.SetHeight(40)
	bb.Objects().Add(child)

	layout := calcBandLayout(bb, 10, nil)

	// maxBottom = 0 + 40 = 40, which is the layout height.
	if layout.height != 40 {
		t.Errorf("non-text object: height = %v, want 40", layout.height)
	}
	// No grow/shrink occurred so shifts should all be zero.
	for i, s := range layout.shifts {
		if s != 0 {
			t.Errorf("non-text object: shifts[%d] = %v, want 0", i, s)
		}
	}
}

// TestCalcBandLayout_CanGrow_TextObject verifies that when a TextObject has
// CanGrow=true and its measured content is taller than its declared height,
// the effective height grows and the band height reflects this.
func TestCalcBandLayout_CanGrow_TextObject(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(20)
	bb.SetWidth(200)

	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(10) // declared shorter than measured content
	txt.SetWidth(200)
	txt.SetCanGrow(true)
	// Multi-line text that is taller than 10px when measured.
	txt.SetText("Line1\nLine2\nLine3\nLine4\nLine5")
	bb.Objects().Add(txt)

	layout := calcBandLayout(bb, 20, nil)

	// The effective height for the TextObject must be taller than declared 10.
	if layout.effectiveH == nil || len(layout.effectiveH) < 1 {
		t.Fatal("CanGrow: effectiveH should be non-nil")
	}
	if layout.effectiveH[0] <= 10 {
		t.Errorf("CanGrow: effectiveH[0] = %v, want > 10 (text grew)", layout.effectiveH[0])
	}
	// Band height must reflect the grown object.
	if layout.height <= 10 {
		t.Errorf("CanGrow: layout.height = %v, want > 10", layout.height)
	}
}

// TestCalcBandLayout_CanShrink_TextObject verifies that when a TextObject has
// CanShrink=true and its measured content is shorter than its declared height,
// the effective height shrinks and the band height reflects this.
func TestCalcBandLayout_CanShrink_TextObject(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(200)
	bb.SetWidth(300)

	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(200) // much taller than one line of text
	txt.SetWidth(300)
	txt.SetCanShrink(true)
	txt.SetText("X") // very short content
	bb.Objects().Add(txt)

	layout := calcBandLayout(bb, 200, nil)

	if layout.effectiveH == nil || len(layout.effectiveH) < 1 {
		t.Fatal("CanShrink: effectiveH should be non-nil")
	}
	// effectiveH[0] must be less than declared 200.
	if layout.effectiveH[0] >= 200 {
		t.Errorf("CanShrink: effectiveH[0] = %v, want < 200 (text shrunk)", layout.effectiveH[0])
	}
	// Band height reflects the shrunken object.
	if layout.height >= 200 {
		t.Errorf("CanShrink: layout.height = %v, want < 200", layout.height)
	}
}

// TestCalcBandLayout_NoGrowShrink_TextObject verifies that a TextObject without
// CanGrow or CanShrink keeps its declared height even when text would differ.
func TestCalcBandLayout_NoGrowShrink_TextObject(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(50)
	bb.SetWidth(200)

	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(50)
	txt.SetWidth(200)
	// Neither CanGrow nor CanShrink — defaults are false.
	txt.SetText("Short text")
	bb.Objects().Add(txt)

	layout := calcBandLayout(bb, 50, nil)

	// effectiveH[0] must remain 50 (declared height unchanged).
	if layout.effectiveH == nil || len(layout.effectiveH) < 1 {
		t.Fatal("no-grow-shrink: effectiveH should not be nil (has objects)")
	}
	if layout.effectiveH[0] != 50 {
		t.Errorf("no-grow-shrink: effectiveH[0] = %v, want 50", layout.effectiveH[0])
	}
	// No grow/shrink occurred, so hasGrowShrink=false → shifts are all zero.
	for i, s := range layout.shifts {
		if s != 0 {
			t.Errorf("no-grow-shrink: shifts[%d] = %v, want 0", i, s)
		}
	}
}

// TestCalcBandLayout_EvalFn verifies that the evalFn callback is used to
// evaluate text content before measuring, so the measured height is based
// on the resolved text.
func TestCalcBandLayout_EvalFn(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(200)
	bb.SetWidth(300)

	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(200)
	txt.SetWidth(300)
	txt.SetCanShrink(true)
	txt.SetText("[SomeExpr]") // template expression
	bb.Objects().Add(txt)

	// evalFn replaces the template with "Hi" (very short text).
	evalFn := func(s string) string { return "Hi" }
	layout := calcBandLayout(bb, 200, evalFn)

	if layout.effectiveH == nil || len(layout.effectiveH) < 1 {
		t.Fatal("evalFn: effectiveH should be non-nil")
	}
	// After shrinking to "Hi", height must be less than 200.
	if layout.effectiveH[0] >= 200 {
		t.Errorf("evalFn: effectiveH[0] = %v, want < 200 (shrunk to short text)", layout.effectiveH[0])
	}
}

// TestCalcBandLayout_InvisibleObject verifies that invisible objects are skipped
// when computing maxBottom (they do not contribute to the layout height).
func TestCalcBandLayout_InvisibleObject(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(50)
	bb.SetWidth(200)

	// Tall invisible object — should not expand band height.
	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(200)
	txt.SetWidth(200)
	txt.SetText("tall content")
	txt.SetVisible(false) // invisible
	bb.Objects().Add(txt)

	layout := calcBandLayout(bb, 50, nil)

	// maxBottom stays 0 because the only object is invisible.
	// The function returns baseHeight when maxBottom <= 0.
	if layout.height != 50 {
		t.Errorf("invisible object: height = %v, want 50 (baseHeight)", layout.height)
	}
}

// TestCalcBandLayout_TwoObjects_GrowShiftsPropagation verifies that when the top
// TextObject grows, the bottom TextObject receives a positive Y shift.
func TestCalcBandLayout_TwoObjects_GrowShiftsPropagation(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(100)
	bb.SetWidth(200)

	// Top object: CanGrow, with lots of multi-line text.
	top := object.NewTextObject()
	top.SetTop(0)
	top.SetHeight(20) // small declared height
	top.SetWidth(200)
	top.SetCanGrow(true)
	top.SetText("L1\nL2\nL3\nL4\nL5\nL6\nL7\nL8\nL9\nL10")
	bb.Objects().Add(top)

	// Bottom object: fixed, sits below top object's declared bottom (top=20).
	bottom := object.NewTextObject()
	bottom.SetTop(25) // clearly below top's original bottom (0+20=20)
	bottom.SetHeight(15)
	bottom.SetWidth(200)
	bottom.SetText("fixed")
	bb.Objects().Add(bottom)

	layout := calcBandLayout(bb, 100, nil)

	if layout.shifts == nil || len(layout.shifts) < 2 {
		t.Fatal("two-objects grow: shifts should have 2 entries")
	}
	// Top object grew → bottom object should have a positive shift.
	if layout.shifts[1] <= 0 {
		t.Errorf("two-objects grow: shifts[1] = %v, want > 0 (bottom shifted down)", layout.shifts[1])
	}
}

// TestCalcBandLayout_TwoObjects_ShrinkShiftsPropagation verifies that when the
// top TextObject shrinks, the bottom TextObject receives a negative Y shift.
func TestCalcBandLayout_TwoObjects_ShrinkShiftsPropagation(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(200)
	bb.SetWidth(300)

	// Top object: large declared height, CanShrink, short text.
	top := object.NewTextObject()
	top.SetTop(0)
	top.SetHeight(100) // declared tall
	top.SetWidth(300)
	top.SetCanShrink(true)
	top.SetText("X") // short content → shrinks
	bb.Objects().Add(top)

	// Bottom object: fixed, sits at top's declared bottom edge (top=100).
	bottom := object.NewTextObject()
	bottom.SetTop(105) // at or below top's original bottom
	bottom.SetHeight(20)
	bottom.SetWidth(300)
	bottom.SetText("below")
	bb.Objects().Add(bottom)

	layout := calcBandLayout(bb, 200, nil)

	if layout.shifts == nil || len(layout.shifts) < 2 {
		t.Fatal("two-objects shrink: shifts should have 2 entries")
	}
	// Top shrunk → bottom object should shift upward (negative shift).
	if layout.shifts[1] >= 0 {
		t.Errorf("two-objects shrink: shifts[1] = %v, want < 0 (bottom shifted up)", layout.shifts[1])
	}
}

// TestCalcBandLayout_ZeroWidthObject verifies that when a TextObject has zero
// width, the band width is used for measurement instead.
func TestCalcBandLayout_ZeroWidthObject(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(200)
	bb.SetWidth(300)

	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(200)
	txt.SetWidth(0) // zero width → falls back to band width
	txt.SetCanShrink(true)
	txt.SetText("short")
	bb.Objects().Add(txt)

	// Should not panic; the fallback to bb.Width() (300) is exercised.
	layout := calcBandLayout(bb, 200, nil)

	if layout.effectiveH == nil {
		t.Fatal("zero-width: effectiveH should not be nil")
	}
	// Text is short, so it should shrink.
	if layout.effectiveH[0] >= 200 {
		t.Errorf("zero-width: effectiveH[0] = %v, want < 200", layout.effectiveH[0])
	}
}

// ── applyBandObjectShifts ─────────────────────────────────────────────────────

// TestApplyBandObjectShifts_EmptyShifts verifies the early return when no
// non-zero shift is present — no PreparedObject positions are changed.
func TestApplyBandObjectShifts_EmptyShifts(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(50)
	bb.SetWidth(200)

	txt := object.NewTextObject()
	txt.SetTop(5)
	txt.SetHeight(20)
	txt.SetWidth(200)
	txt.SetText("hello")
	txt.SetVisible(true)
	bb.Objects().Add(txt)

	pb := &preview.PreparedBand{Name: "test", Height: 50}
	e.populateBandObjects(bb, pb)
	originalTop := pb.Objects[0].Top

	// All-zero shifts → no change expected.
	shifts := []float32{0}
	applyBandObjectShifts(bb, pb, shifts)

	if pb.Objects[0].Top != originalTop {
		t.Errorf("all-zero shifts: Top = %v, want %v (unchanged)", pb.Objects[0].Top, originalTop)
	}
}

// TestApplyBandObjectShifts_NilOrEmpty verifies the nil/empty early-return paths.
func TestApplyBandObjectShifts_NilOrEmpty(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(30)
	bb.SetWidth(100)

	txt := object.NewTextObject()
	txt.SetTop(10)
	txt.SetHeight(15)
	txt.SetWidth(100)
	txt.SetText("T")
	txt.SetVisible(true)
	bb.Objects().Add(txt)

	pb := &preview.PreparedBand{Name: "test", Height: 30}
	e.populateBandObjects(bb, pb)

	originalTop := pb.Objects[0].Top

	// nil shifts → early return
	applyBandObjectShifts(bb, pb, nil)
	if pb.Objects[0].Top != originalTop {
		t.Errorf("nil shifts: Top changed to %v, want %v", pb.Objects[0].Top, originalTop)
	}

	// empty shifts → early return
	applyBandObjectShifts(bb, pb, []float32{})
	if pb.Objects[0].Top != originalTop {
		t.Errorf("empty shifts: Top changed to %v, want %v", pb.Objects[0].Top, originalTop)
	}
}

// TestApplyBandObjectShifts_PositiveShift verifies that a positive shift value
// increases the PreparedObject's Top position.
func TestApplyBandObjectShifts_PositiveShift(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(100)
	bb.SetWidth(200)

	txt := object.NewTextObject()
	txt.SetTop(30)
	txt.SetHeight(20)
	txt.SetWidth(200)
	txt.SetText("shifted down")
	txt.SetVisible(true)
	bb.Objects().Add(txt)

	pb := &preview.PreparedBand{Name: "test", Height: 100}
	e.populateBandObjects(bb, pb)
	originalTop := pb.Objects[0].Top

	// Apply a positive shift of 15 to the first object.
	shifts := []float32{15}
	applyBandObjectShifts(bb, pb, shifts)

	want := originalTop + 15
	if pb.Objects[0].Top != want {
		t.Errorf("positive shift: Top = %v, want %v", pb.Objects[0].Top, want)
	}
}

// TestApplyBandObjectShifts_NegativeShift verifies that a negative shift value
// decreases the PreparedObject's Top position.
func TestApplyBandObjectShifts_NegativeShift(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(100)
	bb.SetWidth(200)

	txt := object.NewTextObject()
	txt.SetTop(50)
	txt.SetHeight(20)
	txt.SetWidth(200)
	txt.SetText("shifted up")
	txt.SetVisible(true)
	bb.Objects().Add(txt)

	pb := &preview.PreparedBand{Name: "test", Height: 100}
	e.populateBandObjects(bb, pb)
	originalTop := pb.Objects[0].Top

	// Apply a negative shift of -10.
	shifts := []float32{-10}
	applyBandObjectShifts(bb, pb, shifts)

	want := originalTop - 10
	if pb.Objects[0].Top != want {
		t.Errorf("negative shift: Top = %v, want %v", pb.Objects[0].Top, want)
	}
}

// TestApplyBandObjectShifts_TwoObjects verifies independent shift values are
// applied to separate objects in the correct order.
func TestApplyBandObjectShifts_TwoObjects(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(100)
	bb.SetWidth(200)

	txt1 := object.NewTextObject()
	txt1.SetTop(10)
	txt1.SetHeight(20)
	txt1.SetWidth(200)
	txt1.SetText("obj1")
	txt1.SetVisible(true)
	bb.Objects().Add(txt1)

	txt2 := object.NewTextObject()
	txt2.SetTop(40)
	txt2.SetHeight(20)
	txt2.SetWidth(200)
	txt2.SetText("obj2")
	txt2.SetVisible(true)
	bb.Objects().Add(txt2)

	pb := &preview.PreparedBand{Name: "test", Height: 100}
	e.populateBandObjects(bb, pb)

	top1Before := pb.Objects[0].Top
	top2Before := pb.Objects[1].Top

	// obj1 gets shift=0, obj2 gets shift=20.
	shifts := []float32{0, 20}
	applyBandObjectShifts(bb, pb, shifts)

	if pb.Objects[0].Top != top1Before {
		t.Errorf("two-objects: obj1 Top = %v, want %v (no shift)", pb.Objects[0].Top, top1Before)
	}
	if pb.Objects[1].Top != top2Before+20 {
		t.Errorf("two-objects: obj2 Top = %v, want %v (+20 shift)", pb.Objects[1].Top, top2Before+20)
	}
}

// TestApplyBandObjectShifts_InvisibleObjectSkipped verifies that an invisible
// object is not counted when matching source objects to PreparedObjects.
func TestApplyBandObjectShifts_InvisibleObjectSkipped(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(100)
	bb.SetWidth(200)

	// Invisible object — buildPreparedObject returns nil, no PreparedObject added.
	inv := object.NewTextObject()
	inv.SetTop(0)
	inv.SetHeight(20)
	inv.SetWidth(200)
	inv.SetText("invisible")
	inv.SetVisible(false)
	bb.Objects().Add(inv)

	// Visible object — gets a PreparedObject.
	vis := object.NewTextObject()
	vis.SetTop(30)
	vis.SetHeight(20)
	vis.SetWidth(200)
	vis.SetText("visible")
	vis.SetVisible(true)
	bb.Objects().Add(vis)

	pb := &preview.PreparedBand{Name: "test", Height: 100}
	e.populateBandObjects(bb, pb)

	// Only one PreparedObject should exist (the visible one).
	if len(pb.Objects) != 1 {
		t.Fatalf("invisible skipped: expected 1 PreparedObject, got %d", len(pb.Objects))
	}

	originalTop := pb.Objects[0].Top

	// shifts[0]=5 (for invisible), shifts[1]=10 (for visible).
	// The invisible object is skipped, so only shifts[1]=10 is applied.
	shifts := []float32{5, 10}
	applyBandObjectShifts(bb, pb, shifts)

	want := originalTop + 10
	if pb.Objects[0].Top != want {
		t.Errorf("invisible skipped: visible Top = %v, want %v", pb.Objects[0].Top, want)
	}
}

// ── applyBandObjectHeights ────────────────────────────────────────────────────

// TestApplyBandObjectHeights_EmptyEffectiveH verifies the empty-slice early return.
func TestApplyBandObjectHeights_EmptyEffectiveH(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(50)
	bb.SetWidth(200)

	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(50)
	txt.SetWidth(200)
	txt.SetText("no change")
	txt.SetVisible(true)
	bb.Objects().Add(txt)

	pb := &preview.PreparedBand{Name: "test", Height: 50}
	e.populateBandObjects(bb, pb)
	originalH := pb.Objects[0].Height

	// nil effectiveH → early return, no change.
	applyBandObjectHeights(bb, pb, nil)
	if pb.Objects[0].Height != originalH {
		t.Errorf("nil effectiveH: Height = %v, want %v", pb.Objects[0].Height, originalH)
	}

	// empty slice → early return
	applyBandObjectHeights(bb, pb, []float32{})
	if pb.Objects[0].Height != originalH {
		t.Errorf("empty effectiveH: Height = %v, want %v", pb.Objects[0].Height, originalH)
	}
}

// TestApplyBandObjectHeights_GrowsObject verifies that when effectiveH is larger
// than the PreparedObject's current height, the height is updated.
func TestApplyBandObjectHeights_GrowsObject(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(50)
	bb.SetWidth(200)

	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(50)
	txt.SetWidth(200)
	txt.SetText("short")
	txt.SetVisible(true)
	bb.Objects().Add(txt)

	pb := &preview.PreparedBand{Name: "test", Height: 50}
	e.populateBandObjects(bb, pb)

	// effectiveH is larger than declared height (grow scenario).
	effectiveH := []float32{80}
	applyBandObjectHeights(bb, pb, effectiveH)

	if pb.Objects[0].Height != 80 {
		t.Errorf("grows object: Height = %v, want 80", pb.Objects[0].Height)
	}
}

// TestApplyBandObjectHeights_ShrinksObject verifies that when effectiveH is smaller
// than the PreparedObject's current height, the height is updated (shrink scenario).
func TestApplyBandObjectHeights_ShrinksObject(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(100)
	bb.SetWidth(200)

	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(100)
	txt.SetWidth(200)
	txt.SetText("short")
	txt.SetVisible(true)
	bb.Objects().Add(txt)

	pb := &preview.PreparedBand{Name: "test", Height: 100}
	e.populateBandObjects(bb, pb)

	// effectiveH is smaller (shrink scenario).
	effectiveH := []float32{25}
	applyBandObjectHeights(bb, pb, effectiveH)

	if pb.Objects[0].Height != 25 {
		t.Errorf("shrinks object: Height = %v, want 25", pb.Objects[0].Height)
	}
}

// TestApplyBandObjectHeights_MatchingHeight verifies that when effectiveH equals
// the PreparedObject's current height, no change is applied.
func TestApplyBandObjectHeights_MatchingHeight(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(50)
	bb.SetWidth(200)

	txt := object.NewTextObject()
	txt.SetTop(0)
	txt.SetHeight(50)
	txt.SetWidth(200)
	txt.SetText("same")
	txt.SetVisible(true)
	bb.Objects().Add(txt)

	pb := &preview.PreparedBand{Name: "test", Height: 50}
	e.populateBandObjects(bb, pb)
	originalH := pb.Objects[0].Height

	// effectiveH exactly matches current height → no change.
	effectiveH := []float32{originalH}
	applyBandObjectHeights(bb, pb, effectiveH)

	if pb.Objects[0].Height != originalH {
		t.Errorf("matching height: Height = %v, want %v", pb.Objects[0].Height, originalH)
	}
}

// TestApplyBandObjectHeights_InvisibleObjectSkipped verifies that invisible
// objects are not matched to PreparedObjects, so only visible objects get
// their heights updated.
func TestApplyBandObjectHeights_InvisibleObjectSkipped(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(100)
	bb.SetWidth(200)

	// Invisible: no PreparedObject will be added.
	inv := object.NewTextObject()
	inv.SetTop(0)
	inv.SetHeight(30)
	inv.SetWidth(200)
	inv.SetText("invisible")
	inv.SetVisible(false)
	bb.Objects().Add(inv)

	// Visible: one PreparedObject added.
	vis := object.NewTextObject()
	vis.SetTop(40)
	vis.SetHeight(30)
	vis.SetWidth(200)
	vis.SetText("visible")
	vis.SetVisible(true)
	bb.Objects().Add(vis)

	pb := &preview.PreparedBand{Name: "test", Height: 100}
	e.populateBandObjects(bb, pb)

	if len(pb.Objects) != 1 {
		t.Fatalf("invisible skipped: expected 1 PreparedObject, got %d", len(pb.Objects))
	}

	// effectiveH[0]=99 (for inv, invisible, skipped), effectiveH[1]=55 (for vis).
	effectiveH := []float32{99, 55}
	applyBandObjectHeights(bb, pb, effectiveH)

	// Only the visible object's PreparedObject exists; effectiveH[1]=55 should be applied.
	if pb.Objects[0].Height != 55 {
		t.Errorf("invisible skipped: visible Height = %v, want 55", pb.Objects[0].Height)
	}
}

// TestApplyBandObjectHeights_MultipleObjects verifies that height updates are
// applied to each object independently and in order.
func TestApplyBandObjectHeights_MultipleObjects(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(100)
	bb.SetWidth(200)

	txt1 := object.NewTextObject()
	txt1.SetTop(0)
	txt1.SetHeight(30)
	txt1.SetWidth(200)
	txt1.SetText("obj1")
	txt1.SetVisible(true)
	bb.Objects().Add(txt1)

	txt2 := object.NewTextObject()
	txt2.SetTop(35)
	txt2.SetHeight(25)
	txt2.SetWidth(200)
	txt2.SetText("obj2")
	txt2.SetVisible(true)
	bb.Objects().Add(txt2)

	pb := &preview.PreparedBand{Name: "test", Height: 100}
	e.populateBandObjects(bb, pb)

	effectiveH := []float32{45, 10}
	applyBandObjectHeights(bb, pb, effectiveH)

	if pb.Objects[0].Height != 45 {
		t.Errorf("multi-object: obj1 Height = %v, want 45", pb.Objects[0].Height)
	}
	if pb.Objects[1].Height != 10 {
		t.Errorf("multi-object: obj2 Height = %v, want 10", pb.Objects[1].Height)
	}
}

// ── Integration: calcBandLayout + apply functions ─────────────────────────────

// TestBandLayout_CanGrow_EndToEnd verifies the full pipeline: calcBandLayout
// computes shifts and effectiveH for a growing TextObject, then both apply
// functions update the PreparedBand correctly.
func TestBandLayout_CanGrow_EndToEnd(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(50)
	bb.SetWidth(200)

	// Top object: CanGrow, tall text.
	topTxt := object.NewTextObject()
	topTxt.SetTop(0)
	topTxt.SetHeight(20)
	topTxt.SetWidth(200)
	topTxt.SetCanGrow(true)
	topTxt.SetText("A\nB\nC\nD\nE\nF\nG\nH\nI\nJ") // 10 lines
	topTxt.SetVisible(true)
	bb.Objects().Add(topTxt)

	// Bottom object: fixed, below the top.
	botTxt := object.NewTextObject()
	botTxt.SetTop(25)
	botTxt.SetHeight(15)
	botTxt.SetWidth(200)
	botTxt.SetText("bottom")
	botTxt.SetVisible(true)
	bb.Objects().Add(botTxt)

	pb := &preview.PreparedBand{Name: "endtoend", Height: 50}
	e.populateBandObjects(bb, pb)

	if len(pb.Objects) != 2 {
		t.Fatalf("end-to-end: expected 2 PreparedObjects, got %d", len(pb.Objects))
	}

	topH0 := pb.Objects[0].Height
	botTop0 := pb.Objects[1].Top

	layout := calcBandLayout(bb, 50, nil)

	// effectiveH[0] should be taller than declared 20 (grew).
	if layout.effectiveH == nil || layout.effectiveH[0] <= 20 {
		t.Errorf("end-to-end: top effectiveH = %v, want > 20", layout.effectiveH)
	}

	if layout.shifts != nil {
		applyBandObjectShifts(bb, pb, layout.shifts)
	}
	if layout.effectiveH != nil {
		applyBandObjectHeights(bb, pb, layout.effectiveH)
	}

	// Top object's height in PreparedBand must have increased.
	if pb.Objects[0].Height <= topH0 {
		t.Errorf("end-to-end: top PreparedObject height = %v, want > %v", pb.Objects[0].Height, topH0)
	}
	// Bottom object's Top must have shifted down.
	if pb.Objects[1].Top <= botTop0 {
		t.Errorf("end-to-end: bottom PreparedObject Top = %v, want > %v", pb.Objects[1].Top, botTop0)
	}
}

// TestBandLayout_CanShrink_EndToEnd verifies the full pipeline for a shrinking
// TextObject: the bottom object shifts upward, top object height decreases.
func TestBandLayout_CanShrink_EndToEnd(t *testing.T) {
	e := buildLayoutEngine(t)

	bb := band.NewBandBase()
	bb.SetHeight(200)
	bb.SetWidth(300)

	// Top object: CanShrink, very short text.
	topTxt := object.NewTextObject()
	topTxt.SetTop(0)
	topTxt.SetHeight(100) // tall declared, will shrink
	topTxt.SetWidth(300)
	topTxt.SetCanShrink(true)
	topTxt.SetText("X")
	topTxt.SetVisible(true)
	bb.Objects().Add(topTxt)

	// Bottom object: fixed, at or below the top object.
	botTxt := object.NewTextObject()
	botTxt.SetTop(105)
	botTxt.SetHeight(20)
	botTxt.SetWidth(300)
	botTxt.SetText("below")
	botTxt.SetVisible(true)
	bb.Objects().Add(botTxt)

	pb := &preview.PreparedBand{Name: "shrink-e2e", Height: 200}
	e.populateBandObjects(bb, pb)

	if len(pb.Objects) != 2 {
		t.Fatalf("shrink-e2e: expected 2 PreparedObjects, got %d", len(pb.Objects))
	}

	topH0 := pb.Objects[0].Height
	botTop0 := pb.Objects[1].Top

	layout := calcBandLayout(bb, 200, nil)

	if layout.effectiveH == nil || layout.effectiveH[0] >= 100 {
		t.Errorf("shrink-e2e: top effectiveH = %v, want < 100", layout.effectiveH)
	}

	if layout.shifts != nil {
		applyBandObjectShifts(bb, pb, layout.shifts)
	}
	if layout.effectiveH != nil {
		applyBandObjectHeights(bb, pb, layout.effectiveH)
	}

	// Top object's height in PreparedBand must have decreased.
	if pb.Objects[0].Height >= topH0 {
		t.Errorf("shrink-e2e: top PreparedObject height = %v, want < %v", pb.Objects[0].Height, topH0)
	}
	// Bottom object's Top must have shifted up (decreased).
	if pb.Objects[1].Top >= botTop0 {
		t.Errorf("shrink-e2e: bottom PreparedObject Top = %v, want < %v", pb.Objects[1].Top, botTop0)
	}
}

// TestApplyBandObjectHeights_NoObjectsInBand verifies the early return when
// the band has no objects — no panic and the function exits cleanly.
func TestApplyBandObjectHeights_NoObjectsInBand(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(50)

	pb := &preview.PreparedBand{Name: "empty", Height: 50}

	// Should not panic; objs.Len() == 0 → early return.
	applyBandObjectHeights(bb, pb, []float32{10, 20})
}

// TestApplyBandObjectShifts_NoObjectsInBand verifies the early return when
// the band has no objects — no panic and the function exits cleanly.
func TestApplyBandObjectShifts_NoObjectsInBand(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(50)

	pb := &preview.PreparedBand{Name: "empty", Height: 50}

	// Should not panic; objs.Len() == 0 → early return.
	applyBandObjectShifts(bb, pb, []float32{10, 20})
}

// TestCalcBandLayout_ObjectWithNoDims verifies that a non-TextObject child that
// does not implement the hasDims interface is skipped gracefully.
func TestCalcBandLayout_ObjectWithNoDims(t *testing.T) {
	bb := band.NewBandBase()
	bb.SetHeight(40)

	// BaseObject has no Height() — exercises the hasDim-false path.
	bb.Objects().Add(report.NewBaseObject())

	// Should not panic.
	layout := calcBandLayout(bb, 40, nil)

	// maxBottom stays 0 → returns baseHeight.
	if layout.height != 40 {
		t.Errorf("no-dims object: height = %v, want 40 (baseHeight)", layout.height)
	}
}
