package preview_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
)

// ── BlobStore ─────────────────────────────────────────────────────────────────

func TestBlobStore_AddGet(t *testing.T) {
	bs := preview.NewBlobStore()
	idx := bs.Add("img1", []byte{1, 2, 3})
	if idx != 0 {
		t.Errorf("first Add idx = %d, want 0", idx)
	}
	got := bs.Get(0)
	if len(got) != 3 || got[0] != 1 {
		t.Error("Get did not return stored blob")
	}
}

func TestBlobStore_DuplicateName(t *testing.T) {
	bs := preview.NewBlobStore()
	idx1 := bs.Add("img", []byte{1})
	idx2 := bs.Add("img", []byte{2}) // same name
	if idx1 != idx2 {
		t.Errorf("duplicate name should return same idx: %d vs %d", idx1, idx2)
	}
	if bs.Count() != 1 {
		t.Errorf("Count = %d, want 1", bs.Count())
	}
}

func TestBlobStore_OutOfRange(t *testing.T) {
	bs := preview.NewBlobStore()
	if bs.Get(0) != nil {
		t.Error("Get on empty store should return nil")
	}
	if bs.Get(-1) != nil {
		t.Error("Get(-1) should return nil")
	}
}

func TestBlobStore_Count(t *testing.T) {
	bs := preview.NewBlobStore()
	bs.Add("a", []byte{1})
	bs.Add("b", []byte{2})
	if bs.Count() != 2 {
		t.Errorf("Count = %d, want 2", bs.Count())
	}
}

// ── Bookmarks ─────────────────────────────────────────────────────────────────

func TestBookmarks_AddFind(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "section1", PageIdx: 2, OffsetY: 100})
	found := bk.Find("section1")
	if found == nil {
		t.Fatal("Find returned nil")
	}
	if found.PageIdx != 2 || found.OffsetY != 100 {
		t.Errorf("bookmark = %+v", found)
	}
}

func TestBookmarks_FindNotFound(t *testing.T) {
	bk := preview.NewBookmarks()
	if bk.Find("x") != nil {
		t.Error("Find on missing name should return nil")
	}
}

func TestBookmarks_Overwrite(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "bm", PageIdx: 0})
	bk.Add(&preview.Bookmark{Name: "bm", PageIdx: 5})
	if bk.Count() != 1 {
		t.Errorf("Count = %d, want 1", bk.Count())
	}
	if bk.Find("bm").PageIdx != 5 {
		t.Error("second Add should overwrite")
	}
}

func TestBookmarks_All(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "a", PageIdx: 0})
	bk.Add(&preview.Bookmark{Name: "b", PageIdx: 1})
	all := bk.All()
	if len(all) != 2 {
		t.Errorf("All len = %d, want 2", len(all))
	}
}

// ── Outline ───────────────────────────────────────────────────────────────────

func TestOutline_New(t *testing.T) {
	o := preview.NewOutline()
	if o.Root == nil {
		t.Error("Root should not be nil")
	}
}

func TestOutline_AddChild(t *testing.T) {
	o := preview.NewOutline()
	child := &preview.OutlineItem{Text: "Chapter 1", PageIdx: 0}
	o.Root.AddChild(child)
	if len(o.Root.Children) != 1 {
		t.Errorf("Children len = %d, want 1", len(o.Root.Children))
	}
}

// ── PreparedPage ──────────────────────────────────────────────────────────────

func TestPreparedPage_AddBand(t *testing.T) {
	pg := &preview.PreparedPage{PageNo: 1, Width: 595, Height: 842}
	pg.AddBand(&preview.PreparedBand{Name: "header", Top: 0, Height: 50})
	if len(pg.Bands) != 1 {
		t.Errorf("Bands len = %d, want 1", len(pg.Bands))
	}
}

// ── PreparedPages ─────────────────────────────────────────────────────────────

func TestPreparedPages_New(t *testing.T) {
	pp := preview.New()
	if pp.Count() != 0 {
		t.Errorf("Count = %d, want 0", pp.Count())
	}
	if pp.CurPage() != -1 {
		t.Errorf("CurPage = %d, want -1", pp.CurPage())
	}
}

func TestPreparedPages_AddPage(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	if pp.Count() != 1 {
		t.Errorf("Count = %d, want 1", pp.Count())
	}
	if pp.CurPage() != 0 {
		t.Errorf("CurPage = %d, want 0", pp.CurPage())
	}
}

func TestPreparedPages_AddPage_Multiple(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)
	pp.AddPage(595, 842, 3)
	if pp.Count() != 3 {
		t.Errorf("Count = %d, want 3", pp.Count())
	}
	if pp.CurPage() != 2 {
		t.Errorf("CurPage = %d, want 2", pp.CurPage())
	}
}

func TestPreparedPages_AddPage_WriteOver(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1) // creates page
	pp.SetAddPageAction(preview.AddPageActionWriteOver)
	pp.AddPage(595, 842, 2) // should rewrite curPage
	if pp.Count() != 1 {
		t.Errorf("WriteOver Count = %d, want 1", pp.Count())
	}
	if pp.GetPage(0).PageNo != 2 {
		t.Errorf("WriteOver PageNo = %d, want 2", pp.GetPage(0).PageNo)
	}
}

func TestPreparedPages_GetPage(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pg := pp.GetPage(0)
	if pg == nil {
		t.Fatal("GetPage(0) returned nil")
	}
	if pg.PageNo != 1 {
		t.Errorf("PageNo = %d, want 1", pg.PageNo)
	}
}

func TestPreparedPages_GetPage_OutOfRange(t *testing.T) {
	pp := preview.New()
	if pp.GetPage(0) != nil {
		t.Error("GetPage(0) on empty should be nil")
	}
	if pp.GetPage(-1) != nil {
		t.Error("GetPage(-1) should be nil")
	}
}

func TestPreparedPages_AddBand(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	err := pp.AddBand(&preview.PreparedBand{Name: "DataBand1", Top: 0, Height: 30})
	if err != nil {
		t.Fatalf("AddBand: %v", err)
	}
	pg := pp.CurrentPage()
	if len(pg.Bands) != 1 {
		t.Errorf("Bands len = %d, want 1", len(pg.Bands))
	}
}

func TestPreparedPages_AddBand_NoPage(t *testing.T) {
	pp := preview.New()
	err := pp.AddBand(&preview.PreparedBand{Name: "b"})
	if err == nil {
		t.Error("expected error when no page started")
	}
}

func TestPreparedPages_RemoveLast(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)
	pp.RemoveLast()
	if pp.Count() != 1 {
		t.Errorf("Count after RemoveLast = %d, want 1", pp.Count())
	}
}

func TestPreparedPages_RemoveLast_Empty(t *testing.T) {
	pp := preview.New()
	pp.RemoveLast() // should not panic
}

func TestPreparedPages_Clear(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)
	pp.Clear()
	if pp.Count() != 0 {
		t.Errorf("Count after Clear = %d, want 0", pp.Count())
	}
	if pp.CurPage() != -1 {
		t.Errorf("CurPage after Clear = %d, want -1", pp.CurPage())
	}
}

func TestPreparedPages_AddBookmark(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddBookmark("intro", 50)
	bm := pp.Bookmarks.Find("intro")
	if bm == nil {
		t.Fatal("bookmark not found")
	}
	if bm.PageIdx != 0 || bm.OffsetY != 50 {
		t.Errorf("bookmark = %+v", bm)
	}
}

func TestPreparedPages_CurrentPage_None(t *testing.T) {
	pp := preview.New()
	if pp.CurrentPage() != nil {
		t.Error("CurrentPage should be nil when no pages added")
	}
}

func TestPreparedPages_NextPage(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)
	pp.NextPage()
	if pp.CurPage() != 2 {
		t.Errorf("CurPage after NextPage = %d, want 2", pp.CurPage())
	}
}
