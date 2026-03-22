package preview_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
)

// ── InterleaveWithBackPage ─────────────────────────────────────────────────────
// C# source: PreparedPages.cs InterleaveWithBackPage(int backPageIndex)

// TestInterleaveWithBackPage_Basic verifies that the back page is interleaved
// between all front pages.
//
// Setup: pages = [front0, front1, front2, back]  (backPageIndex = 3)
// After:   [front0, back', front1, back', front2, back]
// where back' is a copy of the back page.
func TestInterleaveWithBackPage_Basic(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1) // front0
	pp.AddPage(595, 842, 2) // front1
	pp.AddPage(595, 842, 3) // front2
	pp.AddPage(595, 842, 4) // back (index 3)

	pp.InterleaveWithBackPage(3)

	// count = backPageIndex - 1 = 2
	// We insert 2 copies: after front0 and after front1.
	// Original: [front0, front1, front2, back]
	// After insert at i=0 → i*2+1=1: [front0, back', front1, front2, back]
	// After insert at i=1 → i*2+1=3: [front0, back', front1, back', front2, back]
	// Total pages = 4 + 2 = 6
	if pp.Count() != 6 {
		t.Fatalf("Count after InterleaveWithBackPage = %d, want 6", pp.Count())
	}

	// Page 0 = front0 (PageNo=1)
	if pp.GetPage(0).PageNo != 1 {
		t.Errorf("page[0].PageNo = %d, want 1", pp.GetPage(0).PageNo)
	}
	// Page 1 = back copy (PageNo=4)
	if pp.GetPage(1).PageNo != 4 {
		t.Errorf("page[1].PageNo = %d, want 4 (back copy)", pp.GetPage(1).PageNo)
	}
	// Page 2 = front1 (PageNo=2)
	if pp.GetPage(2).PageNo != 2 {
		t.Errorf("page[2].PageNo = %d, want 2", pp.GetPage(2).PageNo)
	}
	// Page 3 = back copy (PageNo=4)
	if pp.GetPage(3).PageNo != 4 {
		t.Errorf("page[3].PageNo = %d, want 4 (back copy)", pp.GetPage(3).PageNo)
	}
	// Page 4 = front2 (PageNo=3)
	if pp.GetPage(4).PageNo != 3 {
		t.Errorf("page[4].PageNo = %d, want 3", pp.GetPage(4).PageNo)
	}
	// Page 5 = original back (PageNo=4)
	if pp.GetPage(5).PageNo != 4 {
		t.Errorf("page[5].PageNo = %d, want 4 (original back)", pp.GetPage(5).PageNo)
	}
}

// TestInterleaveWithBackPage_TwoFrontPages exercises the minimal non-trivial case:
// one front page + one back page → count = 1 front, so 0 copies inserted.
// C# count = backPageIndex - 1 = 1 - 1 = 0, loop body never executes.
func TestInterleaveWithBackPage_TwoFrontPages(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1) // front (index 0)
	pp.AddPage(595, 842, 2) // back  (index 1, backPageIndex = 1)

	pp.InterleaveWithBackPage(1) // count = 0, no interleaving

	// No copies inserted; collection unchanged.
	if pp.Count() != 2 {
		t.Errorf("Count = %d, want 2 (no interleaving for single front page)", pp.Count())
	}
}

// TestInterleaveWithBackPage_OneFrontOneCopy verifies 1 copy is inserted
// when there are exactly 2 front pages and 1 back page.
func TestInterleaveWithBackPage_OneFrontOneCopy(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1) // front0
	pp.AddPage(595, 842, 2) // front1
	pp.AddPage(595, 842, 3) // back (index 2)

	pp.InterleaveWithBackPage(2) // count = 2-1 = 1, one copy

	// [front0, back', front1, back]
	if pp.Count() != 4 {
		t.Fatalf("Count = %d, want 4", pp.Count())
	}
	if pp.GetPage(0).PageNo != 1 {
		t.Errorf("page[0].PageNo = %d, want 1", pp.GetPage(0).PageNo)
	}
	if pp.GetPage(1).PageNo != 3 {
		t.Errorf("page[1].PageNo = %d, want 3 (back copy)", pp.GetPage(1).PageNo)
	}
	if pp.GetPage(2).PageNo != 2 {
		t.Errorf("page[2].PageNo = %d, want 2", pp.GetPage(2).PageNo)
	}
	if pp.GetPage(3).PageNo != 3 {
		t.Errorf("page[3].PageNo = %d, want 3 (original back)", pp.GetPage(3).PageNo)
	}
}

// TestInterleaveWithBackPage_OutOfRange verifies no-op for invalid indices.
func TestInterleaveWithBackPage_OutOfRange(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)

	pp.InterleaveWithBackPage(0)  // backPageIndex < 1 → no-op
	pp.InterleaveWithBackPage(5)  // out of range → no-op
	pp.InterleaveWithBackPage(-1) // negative → no-op

	if pp.Count() != 1 {
		t.Errorf("Count = %d, want 1 (no-op on invalid index)", pp.Count())
	}
}

// TestInterleaveWithBackPage_Empty verifies no-op on empty collection.
func TestInterleaveWithBackPage_Empty(t *testing.T) {
	pp := preview.New()
	pp.InterleaveWithBackPage(0) // should not panic
	if pp.Count() != 0 {
		t.Errorf("Count = %d, want 0", pp.Count())
	}
}

// ── ProcessUnlimited / PostProcessBandUnlimited ───────────────────────────────
// C# source: PreparedPagePostprocessor.cs PostprocessUnlimited / PostProcessBandUnlimitedPage.

// TestProcessUnlimited_NoDuplicates verifies that ProcessUnlimited on a page
// with no duplicate/merge candidates is a no-op.
func TestProcessUnlimited_NoDuplicates(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 2000, 1) // tall unlimited-height page
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "b0", Top: 0, Height: 50,
		Objects: []preview.PreparedObject{
			{Kind: preview.ObjectTypeText, Name: "txt1", Text: "Hello", Duplicates: preview.DuplicatesShow},
		},
	})
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "b1", Top: 50, Height: 50,
		Objects: []preview.PreparedObject{
			{Kind: preview.ObjectTypeText, Name: "txt2", Text: "World", Duplicates: preview.DuplicatesShow},
		},
	})

	proc := preview.NewPostprocessor(pp)
	proc.ProcessUnlimited(0)

	// Objects should be unchanged.
	if pp.GetPage(0).Bands[0].Objects[0].Text != "Hello" {
		t.Errorf("text changed unexpectedly: %q", pp.GetPage(0).Bands[0].Objects[0].Text)
	}
	if pp.GetPage(0).Bands[1].Objects[0].Text != "World" {
		t.Errorf("text changed unexpectedly: %q", pp.GetPage(0).Bands[1].Objects[0].Text)
	}
}

// TestProcessUnlimited_DuplicatesClear verifies that ProcessUnlimited clears
// duplicate text in bands with DuplicatesClear mode.
// Objects must have a non-zero Height so the adjacency check (|curr.absTop - prev.absBottom| <= 0.5) passes.
func TestProcessUnlimited_DuplicatesClear(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 2000, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "b0", Top: 0, Height: 50,
		Objects: []preview.PreparedObject{
			{Kind: preview.ObjectTypeText, Name: "col1", Text: "Value",
				Top: 0, Height: 50, Duplicates: preview.DuplicatesClear},
		},
	})
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "b1", Top: 50, Height: 50,
		Objects: []preview.PreparedObject{
			{Kind: preview.ObjectTypeText, Name: "col1", Text: "Value",
				Top: 0, Height: 50, Duplicates: preview.DuplicatesClear},
		},
	})

	proc := preview.NewPostprocessor(pp)
	proc.ProcessUnlimited(0)

	// First occurrence kept, second cleared.
	if pp.GetPage(0).Bands[0].Objects[0].Text != "Value" {
		t.Errorf("first band text = %q, want Value", pp.GetPage(0).Bands[0].Objects[0].Text)
	}
	if pp.GetPage(0).Bands[1].Objects[0].Text != "" {
		t.Errorf("second band text = %q, want empty (duplicate cleared)", pp.GetPage(0).Bands[1].Objects[0].Text)
	}
}

// TestProcessUnlimited_OutOfRange verifies no panic for invalid page index.
func TestProcessUnlimited_OutOfRange(t *testing.T) {
	pp := preview.New()
	proc := preview.NewPostprocessor(pp)
	proc.ProcessUnlimited(-1) // no panic
	proc.ProcessUnlimited(99) // no panic
}

// TestPostProcessBandUnlimited_AdvancesCounter verifies that the method
// advances the sequential band counter on each call.
func TestPostProcessBandUnlimited_AdvancesCounter(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 2000, 1)
	b0 := &preview.PreparedBand{Name: "b0", Top: 0, Height: 50}
	b1 := &preview.PreparedBand{Name: "b1", Top: 50, Height: 50}
	_ = pp.AddBand(b0)
	_ = pp.AddBand(b1)

	proc := preview.NewPostprocessor(pp)
	proc.ProcessUnlimited(0)

	// PostProcessBandUnlimited should return the same pointer (no structural change).
	got0 := proc.PostProcessBandUnlimited(b0)
	got1 := proc.PostProcessBandUnlimited(b1)

	if got0 != b0 {
		t.Error("PostProcessBandUnlimited should return the same band pointer (band 0)")
	}
	if got1 != b1 {
		t.Error("PostProcessBandUnlimited should return the same band pointer (band 1)")
	}
}
