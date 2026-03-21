package preview_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
)

// TestMergeTextObjects_VerticalMerge verifies that two adjacent text objects
// with the same text and MergeModeVertical are merged into the first one.
// The first object's height should grow to cover both; the second should be
// hidden (Width=0, Height=0, Text="").
// C# source: PreparedPagePostprocessor.cs MergeTextObjects / Merge (vertical branch).
func TestMergeTextObjects_VerticalMerge(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []preview.PreparedObject{
			{
				Name:      "cell",
				Kind:      preview.ObjectTypeText,
				Left:      0,
				Top:       0,
				Width:     100,
				Height:    20,
				Text:      "hello",
				MergeMode: preview.MergeModeVertical,
			},
			{
				Name:      "cell2",
				Kind:      preview.ObjectTypeText,
				Left:      0,
				Top:       20,
				Width:     100,
				Height:    20,
				Text:      "hello",
				MergeMode: preview.MergeModeVertical,
			},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	// First object should be stretched to height 40.
	if objs[0].Height != 40 {
		t.Errorf("merged height = %v, want 40", objs[0].Height)
	}
	if objs[0].Text != "hello" {
		t.Errorf("merged text = %q, want hello", objs[0].Text)
	}
	// Second object should be hidden.
	if objs[1].Width != 0 || objs[1].Height != 0 || objs[1].Text != "" {
		t.Errorf("second obj should be hidden: width=%v height=%v text=%q", objs[1].Width, objs[1].Height, objs[1].Text)
	}
}

// TestMergeTextObjects_VerticalMerge_ThreeObjects tests merging of three
// vertically stacked same-text objects into one.
func TestMergeTextObjects_VerticalMerge_ThreeObjects(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 60,
		Objects: []preview.PreparedObject{
			{Name: "c1", Kind: preview.ObjectTypeText, Left: 0, Top: 0, Width: 100, Height: 20, Text: "ABC", MergeMode: preview.MergeModeVertical},
			{Name: "c2", Kind: preview.ObjectTypeText, Left: 0, Top: 20, Width: 100, Height: 20, Text: "ABC", MergeMode: preview.MergeModeVertical},
			{Name: "c3", Kind: preview.ObjectTypeText, Left: 0, Top: 40, Width: 100, Height: 20, Text: "ABC", MergeMode: preview.MergeModeVertical},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Height != 60 {
		t.Errorf("triple-merged height = %v, want 60", objs[0].Height)
	}
	for i := 1; i < len(objs); i++ {
		if objs[i].Height != 0 {
			t.Errorf("objs[%d].Height = %v, want 0 (hidden)", i, objs[i].Height)
		}
	}
}

// TestMergeTextObjects_HorizontalMerge verifies that two side-by-side text
// objects with the same text and MergeModeHorizontal are merged.
// C# source: PreparedPagePostprocessor.cs Merge (horizontal branch).
func TestMergeTextObjects_HorizontalMerge(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 20,
		Objects: []preview.PreparedObject{
			{Name: "a1", Kind: preview.ObjectTypeText, Left: 0, Top: 0, Width: 50, Height: 20, Text: "X", MergeMode: preview.MergeModeHorizontal},
			{Name: "a2", Kind: preview.ObjectTypeText, Left: 50, Top: 0, Width: 50, Height: 20, Text: "X", MergeMode: preview.MergeModeHorizontal},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	// First object should be stretched to width 100.
	if objs[0].Width != 100 {
		t.Errorf("merged width = %v, want 100", objs[0].Width)
	}
	// Second object should be hidden.
	if objs[1].Width != 0 || objs[1].Height != 0 || objs[1].Text != "" {
		t.Errorf("second obj should be hidden: width=%v height=%v text=%q", objs[1].Width, objs[1].Height, objs[1].Text)
	}
}

// TestMergeTextObjects_DifferentText_NoMerge verifies that objects with
// different text content are not merged even when MergeMode is set.
func TestMergeTextObjects_DifferentText_NoMerge(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []preview.PreparedObject{
			{Name: "r1", Kind: preview.ObjectTypeText, Left: 0, Top: 0, Width: 100, Height: 20, Text: "A", MergeMode: preview.MergeModeVertical},
			{Name: "r2", Kind: preview.ObjectTypeText, Left: 0, Top: 20, Width: 100, Height: 20, Text: "B", MergeMode: preview.MergeModeVertical},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Height != 20 || objs[1].Height != 20 {
		t.Error("different-text objects should not be merged")
	}
}

// TestMergeTextObjects_MergeModeNone_NoMerge verifies that objects with
// MergeModeNone are not collected for merging.
func TestMergeTextObjects_MergeModeNone_NoMerge(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []preview.PreparedObject{
			{Name: "n1", Kind: preview.ObjectTypeText, Left: 0, Top: 0, Width: 100, Height: 20, Text: "same", MergeMode: preview.MergeModeNone},
			{Name: "n2", Kind: preview.ObjectTypeText, Left: 0, Top: 20, Width: 100, Height: 20, Text: "same", MergeMode: preview.MergeModeNone},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Height != 20 || objs[1].Height != 20 {
		t.Error("MergeModeNone objects should not be merged")
	}
}

// TestMergeTextObjects_NonAdjacent_NoMerge verifies that non-adjacent objects
// (with a gap between them) are not merged even when same text and MergeMode.
func TestMergeTextObjects_NonAdjacent_NoMerge(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 50,
		Objects: []preview.PreparedObject{
			// bottom of first is 20; top of second is 30 → gap = 10 > 0.01
			{Name: "g1", Kind: preview.ObjectTypeText, Left: 0, Top: 0, Width: 100, Height: 20, Text: "Z", MergeMode: preview.MergeModeVertical},
			{Name: "g2", Kind: preview.ObjectTypeText, Left: 0, Top: 30, Width: 100, Height: 20, Text: "Z", MergeMode: preview.MergeModeVertical},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Height != 20 || objs[1].Height != 20 {
		t.Error("non-adjacent objects should not be merged")
	}
}

// TestMergeTextObjects_DifferentWidth_NoVerticalMerge verifies that vertical
// merge does not apply when the objects have different widths.
func TestMergeTextObjects_DifferentWidth_NoVerticalMerge(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []preview.PreparedObject{
			{Name: "w1", Kind: preview.ObjectTypeText, Left: 0, Top: 0, Width: 100, Height: 20, Text: "same", MergeMode: preview.MergeModeVertical},
			{Name: "w2", Kind: preview.ObjectTypeText, Left: 0, Top: 20, Width: 80, Height: 20, Text: "same", MergeMode: preview.MergeModeVertical},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Height != 20 {
		t.Errorf("different-width objects should not be vertically merged, got height=%v", objs[0].Height)
	}
}

// TestMergeTextObjects_HorizontalMergeAbove verifies that when obj is directly
// to the LEFT of obj2, obj2 is extended leftward (Left decremented, Width grows).
// C# source: PreparedPagePostprocessor.cs Merge horizontal branch line 205-210.
func TestMergeTextObjects_HorizontalMergeAbove(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	// obj2 (index 0) is at x=50..100; obj (index 1) is at x=0..50.
	// After sorting by absLeft, obj (left=0) comes first, obj2 (left=50) second.
	// The postprocessor tries merge(obj[1], obj[0]): obj[1].absLeft=50 == obj[0].absRight=50,
	// which hits the "obj2.Left == obj.Right" branch in C#.
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 20,
		Objects: []preview.PreparedObject{
			{Name: "h1", Kind: preview.ObjectTypeText, Left: 50, Top: 0, Width: 50, Height: 20, Text: "X", MergeMode: preview.MergeModeHorizontal},
			{Name: "h2", Kind: preview.ObjectTypeText, Left: 0, Top: 0, Width: 50, Height: 20, Text: "X", MergeMode: preview.MergeModeHorizontal},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	// After merging, one object should cover the full 100px width; the other should be hidden.
	totalCovered := float32(0)
	for _, o := range objs {
		totalCovered += o.Width
	}
	if totalCovered != 100 {
		t.Errorf("total width after horizontal merge = %v, want 100", totalCovered)
	}
}

// TestMergeTextObjects_EmptyPage verifies that Process() does not panic on
// empty prepared pages.
func TestMergeTextObjects_EmptyPage(t *testing.T) {
	pp := preview.New()
	preview.NewPostprocessor(pp).Process() // must not panic
}

// TestMergeTextObjects_NoBand verifies that Process() does not panic on a
// page with no bands.
func TestMergeTextObjects_NoBand(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	preview.NewPostprocessor(pp).Process() // must not panic
}

// TestMergeTextObjects_BandOffset verifies that band.Left and band.Top are
// correctly used when computing absolute positions for merge adjacency checks.
func TestMergeTextObjects_BandOffset(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	// Band starts at Top=100; objects at Top=0 and Top=20 within band.
	// Absolute tops: 100 and 120 respectively — they are adjacent.
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Left:   10,
		Top:    100,
		Height: 40,
		Objects: []preview.PreparedObject{
			{Name: "bo1", Kind: preview.ObjectTypeText, Left: 0, Top: 0, Width: 100, Height: 20, Text: "V", MergeMode: preview.MergeModeVertical},
			{Name: "bo2", Kind: preview.ObjectTypeText, Left: 0, Top: 20, Width: 100, Height: 20, Text: "V", MergeMode: preview.MergeModeVertical},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Height != 40 {
		t.Errorf("merged height (with band offset) = %v, want 40", objs[0].Height)
	}
}

// TestMergeMode_TypeValues verifies the MergeMode type constants have expected values,
// matching the C# MergeMode enum bit flags.
func TestMergeMode_TypeValues(t *testing.T) {
	if preview.MergeModeNone != 0 {
		t.Errorf("MergeModeNone = %d, want 0", preview.MergeModeNone)
	}
	if preview.MergeModeHorizontal != 1 {
		t.Errorf("MergeModeHorizontal = %d, want 1", preview.MergeModeHorizontal)
	}
	if preview.MergeModeVertical != 2 {
		t.Errorf("MergeModeVertical = %d, want 2", preview.MergeModeVertical)
	}
}
