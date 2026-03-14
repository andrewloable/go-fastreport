package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func newOutlineEngine(t *testing.T) *engine.ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return e
}

func TestAddOutline_AddsChildToRoot(t *testing.T) {
	e := newOutlineEngine(t)
	pp := e.PreparedPages()
	e.AddOutline("Section 1")
	if len(pp.Outline.Root.Children) != 1 {
		t.Errorf("expected 1 child in root, got %d", len(pp.Outline.Root.Children))
	}
	if pp.Outline.Root.Children[0].Text != "Section 1" {
		t.Errorf("child text = %q, want 'Section 1'", pp.Outline.Root.Children[0].Text)
	}
}

func TestAddOutline_NestsAfterAdd(t *testing.T) {
	e := newOutlineEngine(t)
	pp := e.PreparedPages()
	e.AddOutline("Parent")
	e.AddOutline("Child")
	// "Child" should be a child of "Parent"
	parent := pp.Outline.Root.Children[0]
	if len(parent.Children) != 1 {
		t.Errorf("expected 1 child of Parent, got %d", len(parent.Children))
	}
}

func TestOutlineUp_MovesLevelUp(t *testing.T) {
	e := newOutlineEngine(t)
	pp := e.PreparedPages()
	e.AddOutline("Section 1")
	e.OutlineUp()
	e.AddOutline("Section 2")
	// Both sections should be at root level.
	if len(pp.Outline.Root.Children) != 2 {
		t.Errorf("expected 2 root children, got %d", len(pp.Outline.Root.Children))
	}
}

func TestOutlineRoot_ResetsToRoot(t *testing.T) {
	e := newOutlineEngine(t)
	pp := e.PreparedPages()
	e.AddOutline("Section 1")
	e.AddOutline("Deep")
	e.OutlineRoot()
	e.AddOutline("Section 2")
	// Both Section 1 and Section 2 should be at root level.
	if len(pp.Outline.Root.Children) != 2 {
		t.Errorf("expected 2 root children after OutlineRoot, got %d", len(pp.Outline.Root.Children))
	}
}

func TestAddBookmark_Registered(t *testing.T) {
	e := newOutlineEngine(t)
	e.AddBookmark("mymark")
	pp := e.PreparedPages()
	bk := pp.Bookmarks.Find("mymark")
	if bk == nil {
		t.Error("bookmark 'mymark' not found")
	}
}

func TestAddBookmark_Empty_NoOp(t *testing.T) {
	e := newOutlineEngine(t)
	e.AddBookmark("")
	// Should not panic or add an entry.
	pp := e.PreparedPages()
	if pp.Bookmarks.Count() != 0 {
		t.Error("empty bookmark name should not be registered")
	}
}

func TestGetBookmarkPage_ReturnsZeroForMissing(t *testing.T) {
	e := newOutlineEngine(t)
	if page := e.GetBookmarkPage("nope"); page != 0 {
		t.Errorf("GetBookmarkPage = %d, want 0", page)
	}
}

func TestGetBookmarkPage_ReturnsPage(t *testing.T) {
	e := newOutlineEngine(t)
	e.AddBookmark("here")
	page := e.GetBookmarkPage("here")
	if page <= 0 {
		t.Errorf("GetBookmarkPage = %d, expected > 0", page)
	}
}
