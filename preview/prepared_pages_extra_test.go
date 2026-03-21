package preview_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
)

// ── Bookmarks.GetPageNo ───────────────────────────────────────────────────────

func TestBookmarks_GetPageNo(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "intro", PageIdx: 2, OffsetY: 100})
	if got := bk.GetPageNo("intro"); got != 3 {
		t.Errorf("GetPageNo = %d, want 3 (PageIdx 2 → 1-based)", got)
	}
}

func TestBookmarks_GetPageNo_Missing(t *testing.T) {
	bk := preview.NewBookmarks()
	if got := bk.GetPageNo("missing"); got != 0 {
		t.Errorf("GetPageNo missing = %d, want 0", got)
	}
}

// ── Outline ───────────────────────────────────────────────────────────────────

func TestOutline_Add_And_LevelUp(t *testing.T) {
	o := preview.NewOutline()
	o.Add("Chapter 1", 0, 0)
	o.Add("Section 1.1", 0, 100)
	o.LevelUp() // back to Chapter 1
	o.Add("Section 1.2", 1, 0) // sibling of Section 1.1

	if len(o.Root.Children) != 1 {
		t.Fatalf("root children = %d, want 1", len(o.Root.Children))
	}
	ch1 := o.Root.Children[0]
	if ch1.Text != "Chapter 1" {
		t.Errorf("ch1.Text = %q", ch1.Text)
	}
	if len(ch1.Children) != 2 {
		t.Errorf("Chapter 1 children = %d, want 2", len(ch1.Children))
	}
}

func TestOutline_LevelUp_EmptyStack(t *testing.T) {
	o := preview.NewOutline()
	o.LevelUp() // should not panic on empty stack
}

func TestOutline_LevelRoot(t *testing.T) {
	o := preview.NewOutline()
	o.Add("Chapter 1", 0, 0)
	o.Add("Section 1.1", 0, 100) // now cursor is at Section 1.1
	o.LevelRoot()                 // reset to root
	o.Add("Chapter 2", 1, 0)     // sibling of Chapter 1

	if len(o.Root.Children) != 2 {
		t.Errorf("root children = %d, want 2 (Chapter 1 and Chapter 2)", len(o.Root.Children))
	}
}

// ── PreparedPages – AddPageAction ─────────────────────────────────────────────

func TestPreparedPages_AddPageAction_Getter(t *testing.T) {
	pp := preview.New()
	if pp.AddPageAction() != preview.AddPageActionAdd {
		t.Error("default AddPageAction should be AddPageActionAdd")
	}
	pp.SetAddPageAction(preview.AddPageActionWriteOver)
	if pp.AddPageAction() != preview.AddPageActionWriteOver {
		t.Error("AddPageAction should be AddPageActionWriteOver after set")
	}
}

// ── CurPosition ───────────────────────────────────────────────────────────────

func TestPreparedPages_CurPosition_NoPage(t *testing.T) {
	pp := preview.New()
	if got := pp.CurPosition(); got != 0 {
		t.Errorf("CurPosition no page = %d, want 0", got)
	}
}

func TestPreparedPages_CurPosition(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	if pp.CurPosition() != 0 {
		t.Error("CurPosition with no bands should be 0")
	}
	_ = pp.AddBand(&preview.PreparedBand{Name: "b1"})
	_ = pp.AddBand(&preview.PreparedBand{Name: "b2"})
	if got := pp.CurPosition(); got != 2 {
		t.Errorf("CurPosition = %d, want 2", got)
	}
}

// ── CutObjects / PasteObjects / CutBands ─────────────────────────────────────

func TestPreparedPages_CutObjects_PasteObjects(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "b0", Top: 0, Height: 20})
	_ = pp.AddBand(&preview.PreparedBand{Name: "b1", Top: 20, Height: 20})
	_ = pp.AddBand(&preview.PreparedBand{Name: "b2", Top: 40, Height: 20})

	pp.CutObjects(1) // cut bands[1:] = b1, b2

	if pp.CurPosition() != 1 {
		t.Errorf("after CutObjects: CurPosition = %d, want 1", pp.CurPosition())
	}
	if len(pp.CutBands()) != 2 {
		t.Errorf("CutBands len = %d, want 2", len(pp.CutBands()))
	}

	// Paste onto a new page with dy=10.
	pp.AddPage(595, 842, 2)
	pp.PasteObjects(0, 10)

	pg2 := pp.GetPage(1)
	if len(pg2.Bands) != 2 {
		t.Fatalf("pg2 Bands = %d, want 2", len(pg2.Bands))
	}
	if pg2.Bands[0].Top != 30 { // 20 + 10
		t.Errorf("pasted band[0].Top = %v, want 30", pg2.Bands[0].Top)
	}
	if pg2.Bands[1].Top != 50 { // 40 + 10
		t.Errorf("pasted band[1].Top = %v, want 50", pg2.Bands[1].Top)
	}
	if len(pp.CutBands()) != 0 {
		t.Error("CutBands should be empty after paste")
	}
}

func TestPreparedPages_CutObjects_NoPage(t *testing.T) {
	pp := preview.New()
	pp.CutObjects(0) // no page → cutBands = nil, no panic
	if pp.CutBands() != nil {
		t.Error("CutBands should be nil when no page")
	}
}

func TestPreparedPages_CutObjects_OutOfRange(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "b"})

	pp.CutObjects(99) // position >= len(bands) → no-op
	if pp.CutBands() != nil {
		t.Error("CutBands should be nil for out-of-range position")
	}
	if pp.CurPosition() != 1 {
		t.Errorf("CurPosition = %d, want 1 (unchanged)", pp.CurPosition())
	}
}

func TestPreparedPages_CutObjects_Position0(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "b0", Top: 0, Height: 20})
	_ = pp.AddBand(&preview.PreparedBand{Name: "b1", Top: 20, Height: 20})

	pp.CutObjects(0) // cut from beginning
	if pp.CurPosition() != 0 {
		t.Errorf("CurPosition = %d, want 0", pp.CurPosition())
	}
	if len(pp.CutBands()) != 2 {
		t.Errorf("CutBands len = %d, want 2", len(pp.CutBands()))
	}
}

func TestPreparedPages_PasteObjects_NoPage(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "b", Top: 0, Height: 20})
	pp.CutObjects(0)
	pp.RemoveLast() // remove the page; now no current page
	pp.PasteObjects(0, 0) // should not panic
}

func TestPreparedPages_PasteObjects_NoCutBands(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.PasteObjects(0, 0) // no cut bands → no-op
	if pp.CurPosition() != 0 {
		t.Errorf("CurPosition = %d, want 0", pp.CurPosition())
	}
}

// ── TrimTo ────────────────────────────────────────────────────────────────────

func TestPreparedPages_TrimTo(t *testing.T) {
	pp := preview.New()
	for i := 1; i <= 5; i++ {
		pp.AddPage(595, 842, i)
	}
	pp.TrimTo(3)
	if pp.Count() != 3 {
		t.Errorf("Count after TrimTo(3) = %d, want 3", pp.Count())
	}
	if pp.CurPage() != 2 {
		t.Errorf("CurPage after TrimTo(3) = %d, want 2", pp.CurPage())
	}
}

func TestPreparedPages_TrimTo_ExceedCount(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.TrimTo(10) // n >= len → no-op
	if pp.Count() != 1 {
		t.Errorf("Count = %d, want 1", pp.Count())
	}
}

func TestPreparedPages_TrimTo_Zero(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)
	pp.TrimTo(0)
	if pp.Count() != 0 {
		t.Errorf("Count after TrimTo(0) = %d, want 0", pp.Count())
	}
}

func TestPreparedPages_TrimTo_Negative(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.TrimTo(-5) // negative → treated as 0
	if pp.Count() != 0 {
		t.Errorf("Count after TrimTo(-5) = %d, want 0", pp.Count())
	}
}

// ── GetCachedPage / ClearPageCache / RemovePageCache ─────────────────────────
// These methods delegate to the embedded PageCache (LRU) and are the Go
// equivalents of C# PreparedPages.GetCachedPage / ClearPageCache / RemovePageCache.

func TestPreparedPages_GetCachedPage_Valid(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)

	pg := pp.GetCachedPage(0)
	if pg == nil {
		t.Fatal("GetCachedPage(0) returned nil")
	}
	if pg.PageNo != 1 {
		t.Errorf("PageNo = %d, want 1", pg.PageNo)
	}
}

func TestPreparedPages_GetCachedPage_OutOfRange(t *testing.T) {
	pp := preview.New()
	if pg := pp.GetCachedPage(0); pg != nil {
		t.Error("GetCachedPage on empty collection should return nil")
	}
}

func TestPreparedPages_GetCachedPage_CachesResult(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)
	pp.AddPage(595, 842, 3)

	// Access pages to populate the cache.
	pg0a := pp.GetCachedPage(0)
	pg0b := pp.GetCachedPage(0) // second call → cache hit, same pointer
	if pg0a != pg0b {
		t.Error("GetCachedPage(0) twice should return the same pointer (cache hit)")
	}
}

func TestPreparedPages_ClearPageCache(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)

	// Warm the cache.
	pp.GetCachedPage(0)
	pp.GetCachedPage(1)

	// ClearPageCache should not panic and the next call should still succeed.
	pp.ClearPageCache()
	pg := pp.GetCachedPage(0)
	if pg == nil {
		t.Fatal("GetCachedPage after ClearPageCache returned nil")
	}
}

func TestPreparedPages_RemovePageCache(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)
	pp.AddPage(595, 842, 3)

	// Warm the cache for pages 0, 1, 2.
	pp.GetCachedPage(0)
	pp.GetCachedPage(1)
	pp.GetCachedPage(2)

	// Remove entry for index 1.
	pp.RemovePageCache(1)

	// Page 1 should still be accessible via GetCachedPage (re-fetched from backing store).
	pg := pp.GetCachedPage(1)
	if pg == nil {
		t.Fatal("GetCachedPage(1) after RemovePageCache returned nil")
	}
}

func TestPreparedPages_RemovePageCache_Missing(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	// RemovePageCache of an index not in the cache should be a no-op (not panic).
	pp.RemovePageCache(99)
}

func TestPreparedPages_Clear_InvalidatesPageCache(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)

	// Warm the cache.
	pp.GetCachedPage(0)
	pp.GetCachedPage(1)

	// Clear removes all pages and should also clear the cache.
	pp.Clear()
	if pp.Count() != 0 {
		t.Errorf("Count after Clear = %d, want 0", pp.Count())
	}
	// GetCachedPage on empty collection must return nil (not a stale cache entry).
	if pg := pp.GetCachedPage(0); pg != nil {
		t.Error("GetCachedPage after Clear should return nil (cache invalidated)")
	}
}
