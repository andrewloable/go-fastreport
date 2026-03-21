// Package preview internal merge tests — covers branches in mergeTextObjects
// and mergeObjects that require direct access to unexported state.
package preview

import "testing"

// TestMergeTextObjects_Internal_VerticalBelowObj2 exercises the vertical merge
// branch where obj is ABOVE obj2 (obj2.Top == obj.Bottom), so obj2 is extended
// upward (Top decremented). This is the second vertical branch in Merge().
// C# source: PreparedPagePostprocessor.cs Merge line 191-196.
func TestMergeTextObjects_Internal_VerticalBelowObj2(t *testing.T) {
	pp := New()
	pp.AddPage(595, 842, 1)
	// After sorting by (absLeft, absTop): obj at Top=0 has absTop=0, comes first.
	// obj at Top=20 has absTop=20, comes second.
	// mergeObjectsInBand tries merge(entries[1], entries[0]).
	// obj = entries[1] (Top=20), obj2 = entries[0] (Top=0).
	// obj2.Bottom = 0+20 = 20 == obj.Top = 20 → obj2 is extended downward.
	_ = pp.AddBand(&PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []PreparedObject{
			{Name: "m1", Kind: ObjectTypeText, Left: 0, Top: 0, Width: 100, Height: 20, Text: "T", MergeMode: MergeModeVertical},
			{Name: "m2", Kind: ObjectTypeText, Left: 0, Top: 20, Width: 100, Height: 20, Text: "T", MergeMode: MergeModeVertical},
		},
	})
	proc := &Postprocessor{pp: pp}
	proc.mergeTextObjects()

	objs := pp.GetPage(0).Bands[0].Objects
	// One of the two should have grown; the other should be hidden.
	if objs[0].Height+objs[1].Height != 40 {
		t.Errorf("total height after vertical merge = %v, want 40", objs[0].Height+objs[1].Height)
	}
}

// TestMergeTextObjects_Internal_EmptyBand verifies that mergeTextObjects does
// not panic or modify objects when a band contains no text objects.
func TestMergeTextObjects_Internal_EmptyBand(t *testing.T) {
	pp := New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&PreparedBand{
		Name:    "empty",
		Top:     0,
		Height:  20,
		Objects: []PreparedObject{},
	})
	proc := &Postprocessor{pp: pp}
	proc.mergeTextObjects() // must not panic
}

// TestMergeTextObjects_Internal_MixedModes exercises a band where some objects
// have MergeModeNone (excluded from collection) and others have MergeModeVertical
// (included). Only the MergeModeVertical pair should be merged.
func TestMergeTextObjects_Internal_MixedModes(t *testing.T) {
	pp := New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&PreparedBand{
		Name:   "mix",
		Top:    0,
		Height: 60,
		Objects: []PreparedObject{
			{Name: "n1", Kind: ObjectTypeText, Left: 0, Top: 0, Width: 100, Height: 20, Text: "same", MergeMode: MergeModeNone},
			{Name: "n2", Kind: ObjectTypeText, Left: 0, Top: 20, Width: 100, Height: 20, Text: "same", MergeMode: MergeModeNone},
			{Name: "v1", Kind: ObjectTypeText, Left: 100, Top: 0, Width: 100, Height: 20, Text: "yes", MergeMode: MergeModeVertical},
			{Name: "v2", Kind: ObjectTypeText, Left: 100, Top: 20, Width: 100, Height: 20, Text: "yes", MergeMode: MergeModeVertical},
		},
	})

	proc := &Postprocessor{pp: pp}
	proc.mergeTextObjects()

	objs := pp.GetPage(0).Bands[0].Objects
	// MergeModeNone objects unchanged.
	if objs[0].Height != 20 || objs[1].Height != 20 {
		t.Error("MergeModeNone objects should not be merged")
	}
	// MergeModeVertical objects should be merged.
	if objs[2].Height+objs[3].Height != 40 {
		t.Errorf("MergeModeVertical pair total height = %v, want 40", objs[2].Height+objs[3].Height)
	}
}

// TestMergeTextObjects_Internal_HorizontalDifferentHeight verifies that
// horizontal merge does not apply when the two objects have different heights.
func TestMergeTextObjects_Internal_HorizontalDifferentHeight(t *testing.T) {
	pp := New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&PreparedBand{
		Name:   "dh",
		Top:    0,
		Height: 30,
		Objects: []PreparedObject{
			{Name: "dh1", Kind: ObjectTypeText, Left: 0, Top: 0, Width: 50, Height: 20, Text: "Z", MergeMode: MergeModeHorizontal},
			{Name: "dh2", Kind: ObjectTypeText, Left: 50, Top: 0, Width: 50, Height: 30, Text: "Z", MergeMode: MergeModeHorizontal},
		},
	})

	proc := &Postprocessor{pp: pp}
	proc.mergeTextObjects()

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Width != 50 {
		t.Errorf("different-height horizontal objects should not be merged, got width=%v", objs[0].Width)
	}
}

// TestMergeTextObjects_Internal_VerticalDifferentLeft verifies that vertical
// merge does not apply when the two objects have different left positions.
func TestMergeTextObjects_Internal_VerticalDifferentLeft(t *testing.T) {
	pp := New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&PreparedBand{
		Name:   "dl",
		Top:    0,
		Height: 40,
		Objects: []PreparedObject{
			{Name: "dl1", Kind: ObjectTypeText, Left: 0, Top: 0, Width: 100, Height: 20, Text: "Q", MergeMode: MergeModeVertical},
			{Name: "dl2", Kind: ObjectTypeText, Left: 10, Top: 20, Width: 100, Height: 20, Text: "Q", MergeMode: MergeModeVertical},
		},
	})

	proc := &Postprocessor{pp: pp}
	proc.mergeTextObjects()

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Height != 20 {
		t.Errorf("different-left vertical objects should not be merged, got height=%v", objs[0].Height)
	}
}

// TestIsEqualWithInaccuracy exercises the isEqualWithInaccuracy helper with
// values within and outside the 0.01 threshold.
// Note: float32 precision means 0.01 as float32 rounds to ~0.009999999776...,
// so isEqualWithInaccuracy(0, 0.01f32) returns true (values differ by < 0.01).
// C# source: PreparedPagePostprocessor.cs IsEqualWithInaccuracy (uses float, not double).
func TestIsEqualWithInaccuracy(t *testing.T) {
	// Exact match should always be true.
	if !isEqualWithInaccuracy(0, 0) {
		t.Error("isEqualWithInaccuracy(0, 0) should be true")
	}
	// Small difference within threshold.
	if !isEqualWithInaccuracy(0, 0.005) {
		t.Error("isEqualWithInaccuracy(0, 0.005) should be true")
	}
	// Large difference clearly outside threshold.
	if isEqualWithInaccuracy(0, 1.0) {
		t.Error("isEqualWithInaccuracy(0, 1.0) should be false")
	}
	if isEqualWithInaccuracy(10, 10.1) {
		t.Error("isEqualWithInaccuracy(10, 10.1) should be false")
	}
	// Symmetry.
	if isEqualWithInaccuracy(0, 0.005) != isEqualWithInaccuracy(0.005, 0) {
		t.Error("isEqualWithInaccuracy should be symmetric")
	}
}
