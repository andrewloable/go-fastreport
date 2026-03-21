package preview_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
)

// ── Bookmarks.Clear ───────────────────────────────────────────────────────────

func TestBookmarks_Clear_RemovesItems(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "a", PageIdx: 0, OffsetY: 10})
	bk.Add(&preview.Bookmark{Name: "b", PageIdx: 1, OffsetY: 20})
	bk.Clear()
	if bk.Count() != 0 {
		t.Errorf("Count after Clear = %d, want 0", bk.Count())
	}
	if bk.Find("a") != nil {
		t.Error("Find('a') should be nil after Clear")
	}
}

func TestBookmarks_Clear_Empty(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Clear() // should not panic on empty collection
	if bk.Count() != 0 {
		t.Errorf("Count after Clear on empty = %d, want 0", bk.Count())
	}
}

// ── Bookmarks.ClearFirstPass ──────────────────────────────────────────────────

func TestBookmarks_ClearFirstPass_FallbackGetPageNo(t *testing.T) {
	// Simulate double-pass: first-pass adds some bookmarks, then ClearFirstPass
	// saves them. GetPageNo should find them via fallback.
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "section1", PageIdx: 2, OffsetY: 100})
	bk.Add(&preview.Bookmark{Name: "section2", PageIdx: 4, OffsetY: 200})

	// End of first pass: save and reset.
	bk.ClearFirstPass()

	// Active items are now empty.
	if bk.Count() != 0 {
		t.Errorf("Count after ClearFirstPass = %d, want 0", bk.Count())
	}
	// GetPageNo should fall back to saved first-pass items.
	if got := bk.GetPageNo("section1"); got != 3 { // PageIdx 2 → 1-based = 3
		t.Errorf("GetPageNo('section1') = %d, want 3", got)
	}
	if got := bk.GetPageNo("section2"); got != 5 {
		t.Errorf("GetPageNo('section2') = %d, want 5", got)
	}
}

func TestBookmarks_ClearFirstPass_ActiveOverridesFallback(t *testing.T) {
	// After ClearFirstPass, add new active bookmarks.
	// GetPageNo should return active (not first-pass) value.
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "bm", PageIdx: 1, OffsetY: 50})
	bk.ClearFirstPass()

	// Add an updated version in the new active list.
	bk.Add(&preview.Bookmark{Name: "bm", PageIdx: 3, OffsetY: 150})

	// GetPageNo should return the active value (page 3 → 1-based = 4).
	if got := bk.GetPageNo("bm"); got != 4 {
		t.Errorf("GetPageNo = %d, want 4 (active overrides fallback)", got)
	}
}

func TestBookmarks_ClearFirstPass_MissingReturnsZero(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "x", PageIdx: 0, OffsetY: 0})
	bk.ClearFirstPass()

	// "y" is neither in active nor in first-pass items.
	if got := bk.GetPageNo("y"); got != 0 {
		t.Errorf("GetPageNo missing = %d, want 0", got)
	}
}

func TestBookmarks_ClearFirstPass_EmptyCollection(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.ClearFirstPass() // should not panic

	if got := bk.GetPageNo("anything"); got != 0 {
		t.Errorf("GetPageNo on empty = %d, want 0", got)
	}
}

// ── Outline.Clear and Outline.IsEmpty ────────────────────────────────────────

func TestOutline_IsEmpty_Empty(t *testing.T) {
	o := preview.NewOutline()
	if !o.IsEmpty() {
		t.Error("IsEmpty on new outline should return true")
	}
}

func TestOutline_IsEmpty_NotEmpty(t *testing.T) {
	o := preview.NewOutline()
	o.Add("Chapter 1", 0, 0)
	o.LevelUp()
	if o.IsEmpty() {
		t.Error("IsEmpty should return false after adding an item")
	}
}

func TestOutline_Clear_RemovesItems(t *testing.T) {
	o := preview.NewOutline()
	o.Add("A", 0, 0)
	o.LevelUp()
	o.Add("B", 1, 50)
	o.LevelUp()
	o.Clear()
	if !o.IsEmpty() {
		t.Error("outline should be empty after Clear")
	}
	if len(o.Root.Children) != 0 {
		t.Errorf("Root.Children len = %d after Clear, want 0", len(o.Root.Children))
	}
}

func TestOutline_Clear_ResetsToRoot(t *testing.T) {
	o := preview.NewOutline()
	o.Add("Chapter 1", 0, 0) // cursor descends into Chapter 1
	o.Clear()                 // should reset cursor to root
	// After Clear, adding an item should attach it to root.
	o.Add("New Chapter", 0, 0)
	o.LevelUp()
	if len(o.Root.Children) != 1 {
		t.Errorf("root children after Clear+Add = %d, want 1", len(o.Root.Children))
	}
	if o.Root.Children[0].Text != "New Chapter" {
		t.Errorf("root.Children[0].Text = %q, want 'New Chapter'", o.Root.Children[0].Text)
	}
}

func TestOutline_Clear_EmptyOutline(t *testing.T) {
	o := preview.NewOutline()
	o.Clear() // should not panic on empty outline
	if !o.IsEmpty() {
		t.Error("outline should be empty after Clear of empty outline")
	}
}

// ── PreparedBandKind ─────────────────────────────────────────────────────────

func TestPreparedBandKind_DefaultIsNormal(t *testing.T) {
	b := &preview.PreparedBand{Name: "data", Top: 0, Height: 20}
	if b.Kind != preview.PreparedBandKindNormal {
		t.Errorf("default Kind = %v, want PreparedBandKindNormal", b.Kind)
	}
}

func TestPreparedBandKind_PageFooter(t *testing.T) {
	b := &preview.PreparedBand{
		Name:   "PageFooter1",
		Kind:   preview.PreparedBandKindPageFooter,
		Top:    800,
		Height: 42,
	}
	if b.Kind != preview.PreparedBandKindPageFooter {
		t.Errorf("Kind = %v, want PreparedBandKindPageFooter", b.Kind)
	}
}

func TestPreparedBandKind_Overlay(t *testing.T) {
	b := &preview.PreparedBand{
		Name:   "Overlay1",
		Kind:   preview.PreparedBandKindOverlay,
		Top:    0,
		Height: 842,
	}
	if b.Kind != preview.PreparedBandKindOverlay {
		t.Errorf("Kind = %v, want PreparedBandKindOverlay", b.Kind)
	}
}

// ── PreparedPages.Clear (extended) ───────────────────────────────────────────

func TestPreparedPages_Clear_ClearsBookmarks(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddBookmark("section1", 100)

	pp.Clear()

	if pp.Bookmarks.Count() != 0 {
		t.Errorf("Bookmarks.Count after Clear = %d, want 0", pp.Bookmarks.Count())
	}
	if pp.Bookmarks.Find("section1") != nil {
		t.Error("Bookmark 'section1' should not exist after Clear")
	}
}

func TestPreparedPages_Clear_ClearsOutline(t *testing.T) {
	pp := preview.New()
	pp.Outline.Add("Chapter 1", 0, 0)
	pp.Outline.LevelUp()

	pp.Clear()

	if !pp.Outline.IsEmpty() {
		t.Error("Outline should be empty after Clear")
	}
}

func TestPreparedPages_Clear_ClearsBlobStore(t *testing.T) {
	pp := preview.New()
	pp.BlobStore.Add("img1", []byte{1, 2, 3})
	pp.BlobStore.Add("img2", []byte{4, 5, 6})

	pp.Clear()

	if pp.BlobStore.Count() != 0 {
		t.Errorf("BlobStore.Count after Clear = %d, want 0", pp.BlobStore.Count())
	}
}

// ── PreparedPages.PrepareToFirstPass ─────────────────────────────────────────

func TestPreparedPages_PrepareToFirstPass_SavesCheckpoint(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "b0"})
	_ = pp.AddBand(&preview.PreparedBand{Name: "b1"})

	pp.PrepareToFirstPass() // saves page=0, position=2

	pp.AddPage(595, 842, 2)
	pp.AddPage(595, 842, 3)
	_ = pp.AddBand(&preview.PreparedBand{Name: "b2"})

	// Before ClearFirstPass: 3 pages
	if pp.Count() != 3 {
		t.Errorf("Count before ClearFirstPass = %d, want 3", pp.Count())
	}
}

func TestPreparedPages_PrepareToFirstPass_SetsOutlineFirstPassPos(t *testing.T) {
	pp := preview.New()
	pp.Outline.Add("Before", 0, 0)
	pp.Outline.LevelUp()

	pp.PrepareToFirstPass()

	// PrepareToFirstPass should reset outline cursor to root.
	pp.Outline.Add("After", 1, 0)
	pp.Outline.LevelUp()

	// Both children should be at root level.
	if len(pp.Outline.Root.Children) != 2 {
		t.Errorf("root children = %d, want 2", len(pp.Outline.Root.Children))
	}
}

// ── PreparedPages.ClearFirstPass ─────────────────────────────────────────────

func TestPreparedPages_ClearFirstPass_TrimsPagesBack(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "b0", Top: 0, Height: 20})
	_ = pp.AddBand(&preview.PreparedBand{Name: "b1", Top: 20, Height: 20})

	pp.PrepareToFirstPass() // checkpoint: page=0, position=2

	pp.AddPage(595, 842, 2)
	pp.AddPage(595, 842, 3)

	pp.ClearFirstPass()

	// Should be back to 1 page (the first-pass page at index 0).
	if pp.Count() != 1 {
		t.Errorf("Count after ClearFirstPass = %d, want 1", pp.Count())
	}
	if pp.CurPage() != 0 {
		t.Errorf("CurPage after ClearFirstPass = %d, want 0", pp.CurPage())
	}
}

func TestPreparedPages_ClearFirstPass_TrimsBandsOnFirstPassPage(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "b0", Top: 0, Height: 20})
	_ = pp.AddBand(&preview.PreparedBand{Name: "b1", Top: 20, Height: 20})

	pp.PrepareToFirstPass() // checkpoint: page=0, position=2

	// Add 2 more bands on the same page.
	_ = pp.AddBand(&preview.PreparedBand{Name: "b2", Top: 40, Height: 20})
	_ = pp.AddBand(&preview.PreparedBand{Name: "b3", Top: 60, Height: 20})

	pp.ClearFirstPass()

	pg := pp.GetPage(0)
	if pg == nil {
		t.Fatal("GetPage(0) = nil after ClearFirstPass")
	}
	if len(pg.Bands) != 2 {
		t.Errorf("Bands after ClearFirstPass = %d, want 2", len(pg.Bands))
	}
}

func TestPreparedPages_ClearFirstPass_AtBeginning(t *testing.T) {
	// When firstPassPage=0 and firstPassPosition=0, the page itself is removed.
	pp := preview.New()
	pp.AddPage(595, 842, 1)

	pp.PrepareToFirstPass() // checkpoint: page=0, position=0 (no bands yet)

	_ = pp.AddBand(&preview.PreparedBand{Name: "b0"})

	pp.ClearFirstPass()

	if pp.Count() != 0 {
		t.Errorf("Count after ClearFirstPass at beginning = %d, want 0", pp.Count())
	}
}

func TestPreparedPages_ClearFirstPass_ClearsOutlineAndBookmarks(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddBookmark("bm1", 100)
	pp.Outline.Add("ch1", 0, 0)
	pp.Outline.LevelUp()

	pp.PrepareToFirstPass()

	pp.AddBookmark("bm2", 200)
	pp.Outline.Add("ch2", 1, 0)
	pp.Outline.LevelUp()

	pp.ClearFirstPass()

	// Bookmarks: ClearFirstPass saves first-pass bookmarks for fallback.
	// The new active list is empty (bm2 was added after PrepareToFirstPass).
	// But bm1 was added BEFORE PrepareToFirstPass, so bm1 is the first-pass item.
	// After ClearFirstPass, active is empty, fallback has bm1.
	if pp.Bookmarks.Count() != 0 {
		t.Errorf("active Bookmarks.Count = %d, want 0", pp.Bookmarks.Count())
	}
	// bm1 should be accessible via fallback.
	if got := pp.Bookmarks.GetPageNo("bm1"); got != 1 {
		t.Errorf("GetPageNo('bm1') via fallback = %d, want 1", got)
	}
}

// ── PreparedPages.GetLastY ────────────────────────────────────────────────────

func TestPreparedPages_GetLastY_NormalBands(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "h", Top: 0, Height: 30})
	_ = pp.AddBand(&preview.PreparedBand{Name: "d", Top: 30, Height: 50})

	lastY := pp.GetLastY()
	if lastY != 80 {
		t.Errorf("GetLastY = %v, want 80", lastY)
	}
}

func TestPreparedPages_GetLastY_ExcludesPageFooter(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "data", Top: 0, Height: 100})
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "footer",
		Kind:   preview.PreparedBandKindPageFooter,
		Top:    800,
		Height: 42,
	})

	lastY := pp.GetLastY()
	// PageFooter should be excluded; max is data band: 0+100=100.
	if lastY != 100 {
		t.Errorf("GetLastY = %v, want 100 (PageFooter excluded)", lastY)
	}
}

func TestPreparedPages_GetLastY_ExcludesOverlay(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "data", Top: 0, Height: 60})
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "overlay",
		Kind:   preview.PreparedBandKindOverlay,
		Top:    0,
		Height: 842,
	})

	lastY := pp.GetLastY()
	if lastY != 60 {
		t.Errorf("GetLastY = %v, want 60 (Overlay excluded)", lastY)
	}
}

func TestPreparedPages_GetLastY_NilPage(t *testing.T) {
	pp := preview.New()
	if got := pp.GetLastY(); got != 0 {
		t.Errorf("GetLastY with no page = %v, want 0", got)
	}
}

func TestPreparedPages_GetLastY_EmptyPage(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	if got := pp.GetLastY(); got != 0 {
		t.Errorf("GetLastY empty page = %v, want 0", got)
	}
}

func TestPreparedPages_GetLastY_AllExcluded(t *testing.T) {
	// Page with only PageFooter and Overlay bands; GetLastY should return 0.
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "footer", Kind: preview.PreparedBandKindPageFooter,
		Top: 800, Height: 42,
	})
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "overlay", Kind: preview.PreparedBandKindOverlay,
		Top: 0, Height: 842,
	})
	if got := pp.GetLastY(); got != 0 {
		t.Errorf("GetLastY all-excluded = %v, want 0", got)
	}
}

// ── PreparedPages.ContainsBand ────────────────────────────────────────────────

func TestPreparedPages_ContainsBand_Found(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "DataBand1"})

	if !pp.ContainsBand("DataBand1") {
		t.Error("ContainsBand('DataBand1') = false, want true")
	}
}

func TestPreparedPages_ContainsBand_NotFound(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "DataBand1"})

	if pp.ContainsBand("PageHeader1") {
		t.Error("ContainsBand('PageHeader1') = true, want false")
	}
}

func TestPreparedPages_ContainsBand_NilPage(t *testing.T) {
	pp := preview.New()
	if pp.ContainsBand("anything") {
		t.Error("ContainsBand with no page should return false")
	}
}

func TestPreparedPages_ContainsBand_EmptyPage(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	if pp.ContainsBand("anything") {
		t.Error("ContainsBand on empty page should return false")
	}
}

func TestPreparedPages_ContainsBand_MultipleBands(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "PageHeader1"})
	_ = pp.AddBand(&preview.PreparedBand{Name: "DataBand1"})
	_ = pp.AddBand(&preview.PreparedBand{Name: "PageFooter1", Kind: preview.PreparedBandKindPageFooter})

	if !pp.ContainsBand("PageHeader1") {
		t.Error("ContainsBand('PageHeader1') should be true")
	}
	if !pp.ContainsBand("DataBand1") {
		t.Error("ContainsBand('DataBand1') should be true")
	}
	if !pp.ContainsBand("PageFooter1") {
		t.Error("ContainsBand('PageFooter1') should be true")
	}
	if pp.ContainsBand("Nonexistent") {
		t.Error("ContainsBand('Nonexistent') should be false")
	}
}

// ── FPX round-trip with PreparedBandKind ─────────────────────────────────────

func TestFPX_PreparedBandKind_RoundTrip(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "data", Kind: preview.PreparedBandKindNormal, Top: 0, Height: 20})
	_ = pp.AddBand(&preview.PreparedBand{Name: "footer", Kind: preview.PreparedBandKindPageFooter, Top: 800, Height: 42})
	_ = pp.AddBand(&preview.PreparedBand{Name: "overlay", Kind: preview.PreparedBandKindOverlay, Top: 0, Height: 842})

	var buf []byte
	bufWriter := &bytesWriter{&buf}
	if err := pp.Save(bufWriter); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := preview.Load(&bytesReader{buf, 0})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	pg := loaded.GetPage(0)
	if pg == nil {
		t.Fatal("GetPage(0) = nil")
	}
	if len(pg.Bands) != 3 {
		t.Fatalf("Bands len = %d, want 3", len(pg.Bands))
	}
	if pg.Bands[0].Kind != preview.PreparedBandKindNormal {
		t.Errorf("Bands[0].Kind = %v, want Normal", pg.Bands[0].Kind)
	}
	if pg.Bands[1].Kind != preview.PreparedBandKindPageFooter {
		t.Errorf("Bands[1].Kind = %v, want PageFooter", pg.Bands[1].Kind)
	}
	if pg.Bands[2].Kind != preview.PreparedBandKindOverlay {
		t.Errorf("Bands[2].Kind = %v, want Overlay", pg.Bands[2].Kind)
	}

	// GetLastY on the loaded pages should exclude PageFooter and Overlay.
	lastY := loaded.GetLastY()
	if lastY != 20 {
		t.Errorf("GetLastY after round-trip = %v, want 20", lastY)
	}
}

// ── helpers for the FPX round-trip test ──────────────────────────────────────

type bytesWriter struct {
	buf *[]byte
}

func (w *bytesWriter) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

type bytesReader struct {
	buf []byte
	pos int
}

func (r *bytesReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.buf) {
		return 0, nil
	}
	n := copy(p, r.buf[r.pos:])
	r.pos += n
	return n, nil
}
