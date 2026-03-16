// Package preview internal tests — covers branches in processDuplicates and
// processGroup that cannot be reached through the public postprocessor API.
package preview

import "testing"

// TestProcessGroup_EmptySlice calls processGroup with an empty dupEntry slice.
// This exercises the loop header of processGroup when there are zero entries,
// confirming no panic and no side effects.
func TestProcessGroup_EmptySlice(t *testing.T) {
	pp := New()
	pp.AddPage(595, 842, 1)
	proc := &Postprocessor{pp: pp}
	// processGroup with zero entries: the for loop condition i<=0 is false
	// immediately so the body is never entered — no crash, no side effects.
	proc.processGroup([]dupEntry{}) // must not panic
}

// TestProcessGroup_SingleEntry calls processGroup with exactly one entry.
// A run of length 1 is never a duplicate, so applyDuplicateMode is never
// called and the object is left unchanged.
func TestProcessGroup_SingleEntry(t *testing.T) {
	pp := New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 20,
		Objects: []PreparedObject{
			{Name: "solo", Kind: ObjectTypeText, Top: 0, Height: 20, Text: "hello",
				Duplicates: DuplicatesClear},
		},
	})
	proc := &Postprocessor{pp: pp}
	entry := dupEntry{pageIdx: 0, bandIdx: 0, objIdx: 0, absTop: 0, absBottom: 20}
	proc.processGroup([]dupEntry{entry}) // single entry → run length 1 → no change
	if got := pp.GetPage(0).Bands[0].Objects[0].Text; got != "hello" {
		t.Errorf("single-entry processGroup changed text to %q, want hello", got)
	}
}

// TestProcessDuplicates_ViaInternalAccess directly calls processDuplicates via
// the unexported method so that the groups loop body (lines 98-103) is reached
// with at least one group entry, exercising the len(entries)==0 guard and the
// p.processGroup call path without going through the public Process() API.
func TestProcessDuplicates_ViaInternalAccess(t *testing.T) {
	pp := New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []PreparedObject{
			{Name: "dup", Kind: ObjectTypeText, Top: 0, Height: 20, Text: "X",
				Duplicates: DuplicatesClear},
			{Name: "dup", Kind: ObjectTypeText, Top: 20, Height: 20, Text: "X",
				Duplicates: DuplicatesClear},
		},
	})
	proc := &Postprocessor{pp: pp}
	proc.processDuplicates() // directly invoke the unexported method

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Text != "X" {
		t.Errorf("first obj text = %q, want X", objs[0].Text)
	}
	if objs[1].Text != "" {
		t.Errorf("second obj text = %q after processDuplicates, want empty", objs[1].Text)
	}
}

// TestProcessDuplicates_MultiPage_CrossPage exercises the groups loop with
// entries spanning multiple pages, ensuring all page indices are correctly
// resolved and that the adjacent-vertical check (which considers per-page
// absTop values) does not produce false positives across pages.
func TestProcessDuplicates_MultiPage_CrossPage(t *testing.T) {
	pp := New()
	// Page 0: object at top=0, height=20 → absTop=0, absBottom=20.
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 20,
		Objects: []PreparedObject{
			{Name: "fieldX", Kind: ObjectTypeText, Top: 0, Height: 20, Text: "same",
				Duplicates: DuplicatesClear},
		},
	})
	// Page 1: object at top=100, height=20 → absTop=100, absBottom=120.
	// Gap between pages: curr.absTop(100) - prev.absBottom(20) = 80 > 0.5 → NOT adjacent.
	pp.AddPage(595, 842, 2)
	_ = pp.AddBand(&PreparedBand{
		Name:   "b2",
		Top:    100,
		Height: 20,
		Objects: []PreparedObject{
			{Name: "fieldX", Kind: ObjectTypeText, Top: 0, Height: 20, Text: "same",
				Duplicates: DuplicatesClear},
		},
	})
	proc := &Postprocessor{pp: pp}
	proc.processDuplicates()

	// Because the two identical-text objects are NOT vertically adjacent
	// (they're on different pages with a large gap), neither should be cleared.
	pg0Txt := pp.GetPage(0).Bands[0].Objects[0].Text
	pg1Txt := pp.GetPage(1).Bands[0].Objects[0].Text
	if pg0Txt != "same" {
		t.Errorf("pg0 text = %q, want same", pg0Txt)
	}
	if pg1Txt != "same" {
		t.Errorf("pg1 text = %q, want same (non-adjacent cross-page)", pg1Txt)
	}
}

// TestProcessDuplicates_EmptyPages exercises the groups loop when pp.pages has
// no qualifying objects, so the groups map remains empty and the for-range body
// is never entered.  This is distinct from TestProcess_EmptyPages (which uses
// the public API) because here we call the unexported method directly.
func TestProcessDuplicates_EmptyPages_Internal(t *testing.T) {
	pp := New()
	proc := &Postprocessor{pp: pp}
	proc.processDuplicates() // must not panic with zero pages
}

// TestProcessDuplicates_AllShow exercises the early-continue for DuplicatesShow,
// ensuring that when all objects have DuplicatesShow the groups map stays empty
// and the for-range body is never entered.
func TestProcessDuplicates_AllShow_Internal(t *testing.T) {
	pp := New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []PreparedObject{
			{Name: "f", Kind: ObjectTypeText, Top: 0, Height: 20, Text: "v",
				Duplicates: DuplicatesShow},
			{Name: "f", Kind: ObjectTypeText, Top: 20, Height: 20, Text: "v",
				Duplicates: DuplicatesShow},
		},
	})
	proc := &Postprocessor{pp: pp}
	proc.processDuplicates()

	// All DuplicatesShow → nothing should be changed.
	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Text != "v" || objs[1].Text != "v" {
		t.Error("DuplicatesShow objects should not be modified by processDuplicates")
	}
}
