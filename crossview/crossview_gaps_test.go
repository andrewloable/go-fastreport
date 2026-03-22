package crossview_test

// crossview_gaps_test.go — tests for the porting gaps addressed in go-fastreport-vt567:
//
//   1. Descriptor.Assign()                   — CrossViewDescriptor.cs Assign()
//   2. CrossViewHeader.IndexOf()             — CrossViewHeader.cs IndexOf()
//   3. CrossViewHeader.Contains()            — CrossViewHeader.cs Contains()
//   4. CrossViewHeader.Insert()              — CrossViewHeader.cs Insert()
//   5. CrossViewHeader.Remove()              — CrossViewHeader.cs Remove()
//   6. CrossViewHeader.ToArray()             — CrossViewHeader.cs ToArray()
//   7. CrossViewHeader.AddRange()            — CrossViewHeader.cs AddRange()

import (
	"testing"

	"github.com/andrewloable/go-fastreport/crossview"
)

// ── Descriptor.Assign ─────────────────────────────────────────────────────────

// TestDescriptor_Assign verifies that Assign copies Expression.
// Mirrors CrossViewDescriptor.Assign() in CrossViewDescriptor.cs (line 90-94).
func TestDescriptor_Assign(t *testing.T) {
	// HeaderDescriptor embeds Descriptor, so we can test Assign via it.
	src := &crossview.HeaderDescriptor{}
	src.Expression = "[Region.Name]"

	var dst crossview.HeaderDescriptor
	dst.Assign(src)

	if dst.Expression != "[Region.Name]" {
		t.Errorf("Expression: got %q, want %q", dst.Expression, "[Region.Name]")
	}
}

// TestDescriptor_Assign_Nil verifies that Assign with nil is a no-op.
func TestDescriptor_Assign_Nil(t *testing.T) {
	var dst crossview.HeaderDescriptor
	dst.Expression = "original"
	dst.Assign(nil) // should not panic or modify dst
	if dst.Expression != "original" {
		t.Errorf("Assign(nil) changed Expression to %q", dst.Expression)
	}
}

// TestDescriptor_Assign_Empty verifies that Assign with empty Expression clears dst.
func TestDescriptor_Assign_Empty(t *testing.T) {
	var dst crossview.HeaderDescriptor
	dst.Expression = "old"

	src := &crossview.HeaderDescriptor{}
	src.Expression = ""
	dst.Assign(src)

	if dst.Expression != "" {
		t.Errorf("Expression: got %q, want empty string", dst.Expression)
	}
}

// ── CrossViewHeader.IndexOf ───────────────────────────────────────────────────

// TestCrossViewHeader_IndexOf_Found verifies IndexOf returns the correct index.
// Mirrors CrossViewHeader.IndexOf() in CrossViewHeader.cs (line 82-85).
func TestCrossViewHeader_IndexOf_Found(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d0 := &crossview.HeaderDescriptor{FieldName: "A"}
	d1 := &crossview.HeaderDescriptor{FieldName: "B"}
	d2 := &crossview.HeaderDescriptor{FieldName: "C"}
	h.Add(d0)
	h.Add(d1)
	h.Add(d2)

	if idx := h.IndexOf(d1); idx != 1 {
		t.Errorf("IndexOf(d1) = %d, want 1", idx)
	}
	if idx := h.IndexOf(d2); idx != 2 {
		t.Errorf("IndexOf(d2) = %d, want 2", idx)
	}
	if idx := h.IndexOf(d0); idx != 0 {
		t.Errorf("IndexOf(d0) = %d, want 0", idx)
	}
}

// TestCrossViewHeader_IndexOf_NotFound verifies IndexOf returns -1 when not present.
func TestCrossViewHeader_IndexOf_NotFound(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	h.Add(&crossview.HeaderDescriptor{FieldName: "A"})

	other := &crossview.HeaderDescriptor{FieldName: "NotAdded"}
	if idx := h.IndexOf(other); idx != -1 {
		t.Errorf("IndexOf(other) = %d, want -1", idx)
	}
}

// TestCrossViewHeader_IndexOf_Empty verifies IndexOf returns -1 on empty collection.
func TestCrossViewHeader_IndexOf_Empty(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d := &crossview.HeaderDescriptor{FieldName: "X"}
	if idx := h.IndexOf(d); idx != -1 {
		t.Errorf("IndexOf on empty collection = %d, want -1", idx)
	}
}

// ── CrossViewHeader.Contains ──────────────────────────────────────────────────

// TestCrossViewHeader_Contains_True verifies Contains returns true for members.
// Mirrors CrossViewHeader.Contains() in CrossViewHeader.cs (line 92-95).
func TestCrossViewHeader_Contains_True(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d := &crossview.HeaderDescriptor{FieldName: "Region"}
	h.Add(d)

	if !h.Contains(d) {
		t.Error("Contains should return true for a member descriptor")
	}
}

// TestCrossViewHeader_Contains_False verifies Contains returns false for non-members.
func TestCrossViewHeader_Contains_False(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	h.Add(&crossview.HeaderDescriptor{FieldName: "Region"})

	other := &crossview.HeaderDescriptor{FieldName: "Region"} // same value, different pointer
	if h.Contains(other) {
		t.Error("Contains should return false for a non-member (different pointer)")
	}
}

// ── CrossViewHeader.Insert ────────────────────────────────────────────────────

// TestCrossViewHeader_Insert_AtFront verifies Insert at index 0.
// Mirrors CrossViewHeader.Insert() in CrossViewHeader.cs (line 60-63).
func TestCrossViewHeader_Insert_AtFront(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d0 := &crossview.HeaderDescriptor{FieldName: "A"}
	d1 := &crossview.HeaderDescriptor{FieldName: "B"}
	h.Add(d0)

	h.Insert(0, d1)

	if h.Count() != 2 {
		t.Fatalf("Count = %d, want 2", h.Count())
	}
	if h.Get(0) != d1 {
		t.Errorf("Get(0) should be d1 after Insert(0, d1)")
	}
	if h.Get(1) != d0 {
		t.Errorf("Get(1) should be d0 after Insert(0, d1)")
	}
}

// TestCrossViewHeader_Insert_AtEnd verifies Insert at last position.
func TestCrossViewHeader_Insert_AtEnd(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d0 := &crossview.HeaderDescriptor{FieldName: "A"}
	d1 := &crossview.HeaderDescriptor{FieldName: "B"}
	h.Add(d0)

	h.Insert(1, d1) // insert after last

	if h.Count() != 2 {
		t.Fatalf("Count = %d, want 2", h.Count())
	}
	if h.Get(0) != d0 {
		t.Errorf("Get(0) should be d0")
	}
	if h.Get(1) != d1 {
		t.Errorf("Get(1) should be d1")
	}
}

// TestCrossViewHeader_Insert_Middle verifies Insert in the middle of the collection.
func TestCrossViewHeader_Insert_Middle(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d0 := &crossview.HeaderDescriptor{FieldName: "A"}
	d1 := &crossview.HeaderDescriptor{FieldName: "B"}
	d2 := &crossview.HeaderDescriptor{FieldName: "C"}
	h.Add(d0)
	h.Add(d2)

	h.Insert(1, d1) // insert between A and C

	if h.Count() != 3 {
		t.Fatalf("Count = %d, want 3", h.Count())
	}
	if h.Get(0) != d0 {
		t.Errorf("Get(0) should be d0")
	}
	if h.Get(1) != d1 {
		t.Errorf("Get(1) should be d1 (inserted)")
	}
	if h.Get(2) != d2 {
		t.Errorf("Get(2) should be d2")
	}
}

// ── CrossViewHeader.Remove ────────────────────────────────────────────────────

// TestCrossViewHeader_Remove_Existing verifies Remove deletes the correct item.
// Mirrors CrossViewHeader.Remove() in CrossViewHeader.cs (line 69-74).
func TestCrossViewHeader_Remove_Existing(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d0 := &crossview.HeaderDescriptor{FieldName: "A"}
	d1 := &crossview.HeaderDescriptor{FieldName: "B"}
	d2 := &crossview.HeaderDescriptor{FieldName: "C"}
	h.Add(d0)
	h.Add(d1)
	h.Add(d2)

	h.Remove(d1)

	if h.Count() != 2 {
		t.Fatalf("Count = %d, want 2", h.Count())
	}
	if h.Contains(d1) {
		t.Error("d1 should no longer be in the collection after Remove")
	}
	if h.Get(0) != d0 {
		t.Errorf("Get(0) should be d0 after removing d1")
	}
	if h.Get(1) != d2 {
		t.Errorf("Get(1) should be d2 after removing d1")
	}
}

// TestCrossViewHeader_Remove_NotExisting verifies Remove is a no-op for non-members.
func TestCrossViewHeader_Remove_NotExisting(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d0 := &crossview.HeaderDescriptor{FieldName: "A"}
	h.Add(d0)

	other := &crossview.HeaderDescriptor{FieldName: "NotAdded"}
	h.Remove(other) // should be a no-op

	if h.Count() != 1 {
		t.Errorf("Count = %d after removing non-member, want 1", h.Count())
	}
}

// TestCrossViewHeader_Remove_First verifies Remove of the first element.
func TestCrossViewHeader_Remove_First(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d0 := &crossview.HeaderDescriptor{FieldName: "First"}
	d1 := &crossview.HeaderDescriptor{FieldName: "Second"}
	h.Add(d0)
	h.Add(d1)

	h.Remove(d0)

	if h.Count() != 1 {
		t.Fatalf("Count = %d, want 1", h.Count())
	}
	if h.Get(0) != d1 {
		t.Errorf("Get(0) should be d1 after removing d0")
	}
}

// TestCrossViewHeader_Remove_Last verifies Remove of the last element.
func TestCrossViewHeader_Remove_Last(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d0 := &crossview.HeaderDescriptor{FieldName: "First"}
	d1 := &crossview.HeaderDescriptor{FieldName: "Last"}
	h.Add(d0)
	h.Add(d1)

	h.Remove(d1)

	if h.Count() != 1 {
		t.Fatalf("Count = %d, want 1", h.Count())
	}
	if h.Get(0) != d0 {
		t.Errorf("Get(0) should be d0 after removing d1")
	}
}

// ── CrossViewHeader.ToArray ───────────────────────────────────────────────────

// TestCrossViewHeader_ToArray verifies ToArray returns all items in order.
// Mirrors CrossViewHeader.ToArray() in CrossViewHeader.cs (line 101-109).
func TestCrossViewHeader_ToArray(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d0 := &crossview.HeaderDescriptor{FieldName: "A"}
	d1 := &crossview.HeaderDescriptor{FieldName: "B"}
	d2 := &crossview.HeaderDescriptor{FieldName: "C"}
	h.Add(d0)
	h.Add(d1)
	h.Add(d2)

	arr := h.ToArray()

	if len(arr) != 3 {
		t.Fatalf("ToArray len = %d, want 3", len(arr))
	}
	if arr[0] != d0 {
		t.Errorf("arr[0] should be d0")
	}
	if arr[1] != d1 {
		t.Errorf("arr[1] should be d1")
	}
	if arr[2] != d2 {
		t.Errorf("arr[2] should be d2")
	}
}

// TestCrossViewHeader_ToArray_Empty verifies ToArray on empty collection returns empty slice.
func TestCrossViewHeader_ToArray_Empty(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	arr := h.ToArray()
	if len(arr) != 0 {
		t.Errorf("ToArray on empty header = len %d, want 0", len(arr))
	}
}

// TestCrossViewHeader_ToArray_IsACopy verifies that modifying the returned array
// does not affect the original collection.
func TestCrossViewHeader_ToArray_IsACopy(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d0 := &crossview.HeaderDescriptor{FieldName: "A"}
	d1 := &crossview.HeaderDescriptor{FieldName: "B"}
	h.Add(d0)
	h.Add(d1)

	arr := h.ToArray()
	arr[0] = nil // mutate the returned slice

	// The original collection should be unaffected.
	if h.Get(0) != d0 {
		t.Error("ToArray modification should not affect the original collection")
	}
}

// ── CrossViewHeader.AddRange ──────────────────────────────────────────────────

// TestCrossViewHeader_AddRange verifies AddRange appends all given descriptors.
// Mirrors CrossViewHeader.AddRange() in CrossViewHeader.cs (line 37-43).
func TestCrossViewHeader_AddRange_AppendsAll(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	d0 := &crossview.HeaderDescriptor{FieldName: "A"}

	newItems := []*crossview.HeaderDescriptor{
		{FieldName: "B"},
		{FieldName: "C"},
	}
	h.Add(d0)
	h.AddRange(newItems)

	if h.Count() != 3 {
		t.Fatalf("Count = %d, want 3", h.Count())
	}
	if h.Get(0) != d0 {
		t.Errorf("Get(0) should be d0")
	}
	if h.Get(1) != newItems[0] {
		t.Errorf("Get(1) should be newItems[0]")
	}
	if h.Get(2) != newItems[1] {
		t.Errorf("Get(2) should be newItems[1]")
	}
}

// TestCrossViewHeader_AddRange_Empty verifies AddRange with empty slice is a no-op.
func TestCrossViewHeader_AddRange_Empty(t *testing.T) {
	h := crossview.NewCrossViewHeader("Columns")
	h.Add(&crossview.HeaderDescriptor{FieldName: "X"})

	h.AddRange(nil)
	if h.Count() != 1 {
		t.Errorf("Count = %d after AddRange(nil), want 1", h.Count())
	}

	h.AddRange([]*crossview.HeaderDescriptor{})
	if h.Count() != 1 {
		t.Errorf("Count = %d after AddRange([]), want 1", h.Count())
	}
}

// TestCrossViewHeader_AddRange_RoundTripWithToArray verifies AddRange + ToArray round-trip.
func TestCrossViewHeader_AddRange_RoundTripWithToArray(t *testing.T) {
	// Populate source collection.
	src := crossview.NewCrossViewHeader("Source")
	src.Add(&crossview.HeaderDescriptor{FieldName: "Year"})
	src.Add(&crossview.HeaderDescriptor{FieldName: "Quarter"})

	// Copy via ToArray → AddRange pattern (used in C# helper code).
	dst := crossview.NewCrossViewHeader("Dest")
	dst.AddRange(src.ToArray())

	if dst.Count() != 2 {
		t.Fatalf("Count = %d, want 2", dst.Count())
	}
	if dst.Get(0).FieldName != "Year" {
		t.Errorf("Get(0).FieldName = %q, want Year", dst.Get(0).FieldName)
	}
	if dst.Get(1).FieldName != "Quarter" {
		t.Errorf("Get(1).FieldName = %q, want Quarter", dst.Get(1).FieldName)
	}
}
