package preview_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
)

// ── Outline.CurPosition ──────────────────────────────────────────────────────

func TestOutline_CurPosition_EmptyRoot(t *testing.T) {
	o := preview.NewOutline()
	if got := o.CurPosition(); got != nil {
		t.Errorf("CurPosition on empty root = %v, want nil", got)
	}
}

func TestOutline_CurPosition_ReturnsLastChild(t *testing.T) {
	o := preview.NewOutline()
	o.Add("A", 0, 10)
	o.LevelUp()
	o.Add("B", 1, 20)
	o.LevelUp()

	// Cursor is at root; last child is "B".
	pos := o.CurPosition()
	if pos == nil {
		t.Fatal("CurPosition returned nil, expected last child")
	}
	if pos.Text != "B" {
		t.Errorf("CurPosition.Text = %q, want 'B'", pos.Text)
	}
}

func TestOutline_CurPosition_NestedCursor(t *testing.T) {
	o := preview.NewOutline()
	o.Add("Parent", 0, 0)
	// Cursor is now inside "Parent".
	o.Add("Child1", 0, 10)
	o.LevelUp() // back to "Parent"
	o.Add("Child2", 0, 20)
	o.LevelUp() // back to "Parent"

	// CurPosition should return "Child2" (last child of "Parent").
	pos := o.CurPosition()
	if pos == nil || pos.Text != "Child2" {
		t.Errorf("CurPosition = %v, want Child2", pos)
	}
}

// ── Outline.Shift ────────────────────────────────────────────────────────────

func TestOutline_Shift_NilFrom(t *testing.T) {
	o := preview.NewOutline()
	o.Shift(nil, 100) // should not panic
}

func TestOutline_Shift_FromNotInTree(t *testing.T) {
	o := preview.NewOutline()
	o.Add("A", 0, 10)
	o.LevelUp()

	orphan := &preview.OutlineItem{Text: "orphan"}
	o.Shift(orphan, 100) // parent not found → no-op, no panic
}

func TestOutline_Shift_NoNextSibling(t *testing.T) {
	o := preview.NewOutline()
	o.Add("A", 0, 10)
	o.LevelUp()

	// "A" is the only child of root; no next sibling to shift.
	from := o.Root.Children[0]
	o.Shift(from, 50) // no-op

	// "A" should be unchanged.
	if from.PageIdx != 0 || from.OffsetY != 10 {
		t.Errorf("A should be unchanged: PageIdx=%d OffsetY=%v", from.PageIdx, from.OffsetY)
	}
}

func TestOutline_Shift_ShiftsNextSibling(t *testing.T) {
	o := preview.NewOutline()
	o.Add("A", 0, 10)
	o.LevelUp()
	o.Add("B", 0, 50)
	o.LevelUp()
	o.Add("C", 0, 80)
	o.LevelUp()

	// Shift from "A" with newY=100. Next sibling is "B" at OffsetY=50.
	// deltaY = 100 - 50 = 50
	from := o.Root.Children[0] // "A"
	o.Shift(from, 100)

	// "A" should be unchanged (only the next sibling and beyond are shifted).
	a := o.Root.Children[0]
	if a.PageIdx != 0 || a.OffsetY != 10 {
		t.Errorf("A should be unchanged: PageIdx=%d OffsetY=%v", a.PageIdx, a.OffsetY)
	}

	// "B" should be shifted: PageIdx=0+1=1, OffsetY=50+50=100.
	b := o.Root.Children[1]
	if b.PageIdx != 1 {
		t.Errorf("B.PageIdx = %d, want 1", b.PageIdx)
	}
	if b.OffsetY != 100 {
		t.Errorf("B.OffsetY = %v, want 100", b.OffsetY)
	}

	// "C" should NOT be shifted (only the immediate next sibling is shifted).
	c := o.Root.Children[2]
	if c.PageIdx != 0 || c.OffsetY != 80 {
		t.Errorf("C should be unchanged: PageIdx=%d OffsetY=%v", c.PageIdx, c.OffsetY)
	}
}

func TestOutline_Shift_RecursiveChildren(t *testing.T) {
	o := preview.NewOutline()
	o.Add("A", 0, 10)
	o.LevelUp()
	o.Add("B", 1, 200)
	// Cursor is on "B"; add children to "B".
	o.Add("B1", 1, 210)
	o.LevelUp() // back to B
	o.Add("B2", 1, 220)
	o.LevelUp() // back to B
	o.LevelUp() // back to root

	// Shift from "A", newY=300. Next sibling is "B" at OffsetY=200.
	// deltaY = 300 - 200 = 100
	from := o.Root.Children[0] // "A"
	o.Shift(from, 300)

	b := o.Root.Children[1]
	if b.PageIdx != 2 || b.OffsetY != 300 {
		t.Errorf("B: PageIdx=%d OffsetY=%v, want 2/300", b.PageIdx, b.OffsetY)
	}
	b1 := b.Children[0]
	if b1.PageIdx != 2 || b1.OffsetY != 310 {
		t.Errorf("B1: PageIdx=%d OffsetY=%v, want 2/310", b1.PageIdx, b1.OffsetY)
	}
	b2 := b.Children[1]
	if b2.PageIdx != 2 || b2.OffsetY != 320 {
		t.Errorf("B2: PageIdx=%d OffsetY=%v, want 2/320", b2.PageIdx, b2.OffsetY)
	}
}

func TestOutline_Shift_NestedFrom(t *testing.T) {
	// Test shift where 'from' is a nested child, not a root child.
	o := preview.NewOutline()
	o.Add("Parent", 0, 0)
	// Cursor is "Parent"
	o.Add("Child1", 0, 10)
	o.LevelUp() // back to Parent
	o.Add("Child2", 0, 50)
	o.LevelUp() // back to Parent
	o.LevelUp() // back to root

	child1 := o.Root.Children[0].Children[0] // "Child1"
	o.Shift(child1, 100)

	// Child2 is the next sibling of Child1. deltaY = 100 - 50 = 50
	child2 := o.Root.Children[0].Children[1]
	if child2.PageIdx != 1 {
		t.Errorf("Child2.PageIdx = %d, want 1", child2.PageIdx)
	}
	if child2.OffsetY != 100 {
		t.Errorf("Child2.OffsetY = %v, want 100", child2.OffsetY)
	}

	// Child1 should be unchanged.
	if child1.PageIdx != 0 || child1.OffsetY != 10 {
		t.Errorf("Child1 changed unexpectedly: PageIdx=%d OffsetY=%v", child1.PageIdx, child1.OffsetY)
	}
}

// ── Outline.PrepareToFirstPass / ClearFirstPass ──────────────────────────────

func TestOutline_PrepareToFirstPass_EmptyRoot(t *testing.T) {
	o := preview.NewOutline()
	o.PrepareToFirstPass()
	// With no children, firstPassPos should be -1.
	// ClearFirstPass should clear all children (nil).
	o.Add("Added after prepare", 0, 0)
	o.LevelUp()
	if len(o.Root.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(o.Root.Children))
	}
	o.ClearFirstPass()
	if len(o.Root.Children) != 0 {
		t.Errorf("after ClearFirstPass: children = %d, want 0", len(o.Root.Children))
	}
}

func TestOutline_PrepareToFirstPass_WithChildren(t *testing.T) {
	o := preview.NewOutline()
	o.Add("Ch1", 0, 0)
	o.LevelUp()
	o.Add("Ch2", 0, 50)
	o.LevelUp()

	o.PrepareToFirstPass() // saves pos=2

	// Add more children (simulating second pass additions).
	o.Add("Ch3", 1, 0)
	o.LevelUp()
	o.Add("Ch4", 1, 50)
	o.LevelUp()

	if len(o.Root.Children) != 4 {
		t.Fatalf("before clear: children = %d, want 4", len(o.Root.Children))
	}

	o.ClearFirstPass() // trims back to 2

	if len(o.Root.Children) != 2 {
		t.Errorf("after ClearFirstPass: children = %d, want 2", len(o.Root.Children))
	}
	if o.Root.Children[0].Text != "Ch1" || o.Root.Children[1].Text != "Ch2" {
		t.Error("wrong children after trim")
	}
}

func TestOutline_PrepareToFirstPass_ResetsToRoot(t *testing.T) {
	o := preview.NewOutline()
	o.Add("A", 0, 0)
	// Cursor is on "A"

	o.PrepareToFirstPass()

	// After PrepareToFirstPass, cursor should be at root.
	o.Add("B", 1, 0)
	o.LevelUp()
	// "B" should be at root level.
	if len(o.Root.Children) != 2 {
		t.Errorf("root children = %d, want 2", len(o.Root.Children))
	}
}

func TestOutline_ClearFirstPass_NoTrimNeeded(t *testing.T) {
	o := preview.NewOutline()
	o.Add("A", 0, 0)
	o.LevelUp()
	o.Add("B", 0, 10)
	o.LevelUp()

	o.PrepareToFirstPass() // saves pos=2

	// Don't add anything new; ClearFirstPass should be a no-op on children.
	o.ClearFirstPass()
	if len(o.Root.Children) != 2 {
		t.Errorf("children = %d, want 2 (no trim needed)", len(o.Root.Children))
	}
}

func TestOutline_ClearFirstPass_ResetsToRoot(t *testing.T) {
	o := preview.NewOutline()
	o.Add("A", 0, 0)
	// Cursor is on "A"
	o.PrepareToFirstPass()

	// Add nested item.
	o.Add("B", 1, 0)
	// Cursor is on "B" (nested in root)

	o.ClearFirstPass()

	// After ClearFirstPass, cursor should be at root.
	o.Add("C", 2, 0)
	o.LevelUp()
	// "C" should be at root level, not nested under something else.
	found := false
	for _, ch := range o.Root.Children {
		if ch.Text == "C" {
			found = true
		}
	}
	if !found {
		t.Error("C should be a root-level child after ClearFirstPass resets cursor")
	}
}

// ── Bookmarks.CurPosition ───────────────────────────────────────────────────

func TestBookmarks_CurPosition_Empty(t *testing.T) {
	bk := preview.NewBookmarks()
	if got := bk.CurPosition(); got != 0 {
		t.Errorf("CurPosition empty = %d, want 0", got)
	}
}

func TestBookmarks_CurPosition_AfterAdd(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "a", PageIdx: 0, OffsetY: 10})
	bk.Add(&preview.Bookmark{Name: "b", PageIdx: 0, OffsetY: 20})
	if got := bk.CurPosition(); got != 2 {
		t.Errorf("CurPosition = %d, want 2", got)
	}
}

// ── Bookmarks.Shift ──────────────────────────────────────────────────────────

func TestBookmarks_Shift_NegativeIndex(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "a", PageIdx: 0, OffsetY: 10})
	bk.Shift(-1, 100) // no-op
	if bk.Find("a").PageIdx != 0 {
		t.Error("Shift(-1, ...) should be a no-op")
	}
}

func TestBookmarks_Shift_IndexOutOfRange(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "a", PageIdx: 0, OffsetY: 10})
	bk.Shift(5, 100) // no-op
	if bk.Find("a").PageIdx != 0 {
		t.Error("Shift with out-of-range index should be a no-op")
	}
}

func TestBookmarks_Shift_EmptyBookmarks(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Shift(0, 100) // no-op, should not panic
}

func TestBookmarks_Shift_SingleItem(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "a", PageIdx: 2, OffsetY: 50})

	bk.Shift(0, 100)

	a := bk.Find("a")
	if a.PageIdx != 3 {
		t.Errorf("PageIdx = %d, want 3", a.PageIdx)
	}
	// shift = 100 - 50 = 50; new OffsetY = 50 + 50 = 100
	if a.OffsetY != 100 {
		t.Errorf("OffsetY = %v, want 100", a.OffsetY)
	}
}

func TestBookmarks_Shift_MultipleItems(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "a", PageIdx: 0, OffsetY: 10})
	bk.Add(&preview.Bookmark{Name: "b", PageIdx: 0, OffsetY: 50})
	bk.Add(&preview.Bookmark{Name: "c", PageIdx: 0, OffsetY: 80})

	// Shift from index 1 (bookmark "b") with newY=200.
	// topY = 50, shift = 200 - 50 = 150
	bk.Shift(1, 200)

	// "a" (index 0) should be unchanged.
	a := bk.Find("a")
	if a.PageIdx != 0 || a.OffsetY != 10 {
		t.Errorf("a should be unchanged: PageIdx=%d OffsetY=%v", a.PageIdx, a.OffsetY)
	}

	// "b" (index 1): PageIdx=0+1=1, OffsetY=50+150=200
	b := bk.Find("b")
	if b.PageIdx != 1 {
		t.Errorf("b.PageIdx = %d, want 1", b.PageIdx)
	}
	if b.OffsetY != 200 {
		t.Errorf("b.OffsetY = %v, want 200", b.OffsetY)
	}

	// "c" (index 2): PageIdx=0+1=1, OffsetY=80+150=230
	c := bk.Find("c")
	if c.PageIdx != 1 {
		t.Errorf("c.PageIdx = %d, want 1", c.PageIdx)
	}
	if c.OffsetY != 230 {
		t.Errorf("c.OffsetY = %v, want 230", c.OffsetY)
	}
}

func TestBookmarks_Shift_FromFirstItem(t *testing.T) {
	bk := preview.NewBookmarks()
	bk.Add(&preview.Bookmark{Name: "a", PageIdx: 0, OffsetY: 100})
	bk.Add(&preview.Bookmark{Name: "b", PageIdx: 0, OffsetY: 200})

	// Shift from index 0 with newY=150.
	// topY = 100, shift = 150 - 100 = 50
	bk.Shift(0, 150)

	a := bk.Find("a")
	if a.PageIdx != 1 || a.OffsetY != 150 {
		t.Errorf("a: PageIdx=%d OffsetY=%v, want 1/150", a.PageIdx, a.OffsetY)
	}
	b := bk.Find("b")
	if b.PageIdx != 1 || b.OffsetY != 250 {
		t.Errorf("b: PageIdx=%d OffsetY=%v, want 1/250", b.PageIdx, b.OffsetY)
	}
}
