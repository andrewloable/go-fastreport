package preview

import (
	"testing"
)

func TestNewSourcePages(t *testing.T) {
	sp := NewSourcePages()
	if sp == nil {
		t.Fatal("NewSourcePages returned nil")
	}
	if sp.Count() != 0 {
		t.Errorf("initial Count = %d, want 0", sp.Count())
	}
}

func TestSourcePages_Record_And_Range(t *testing.T) {
	sp := NewSourcePages()
	sp.Record(0, 0, 2)
	sp.Record(1, 3, 5)

	first, last, ok := sp.Range(0)
	if !ok || first != 0 || last != 2 {
		t.Errorf("Range(0) = %d,%d,%v, want 0,2,true", first, last, ok)
	}
	first, last, ok = sp.Range(1)
	if !ok || first != 3 || last != 5 {
		t.Errorf("Range(1) = %d,%d,%v, want 3,5,true", first, last, ok)
	}
}

func TestSourcePages_Range_Missing(t *testing.T) {
	sp := NewSourcePages()
	_, _, ok := sp.Range(99)
	if ok {
		t.Error("Range(missing) should return ok=false")
	}
}

func TestSourcePages_Record_Replace(t *testing.T) {
	sp := NewSourcePages()
	sp.Record(0, 0, 2)
	sp.Record(0, 5, 8) // replace
	_, last, ok := sp.Range(0)
	if !ok || last != 8 {
		t.Errorf("after replace: last = %d, want 8", last)
	}
	if sp.Count() != 1 {
		t.Errorf("Count after replace = %d, want 1", sp.Count())
	}
}

func TestSourcePages_Clear(t *testing.T) {
	sp := NewSourcePages()
	sp.Record(0, 0, 1)
	sp.Record(1, 2, 3)
	sp.Clear()
	if sp.Count() != 0 {
		t.Errorf("after Clear: Count = %d, want 0", sp.Count())
	}
}

func TestSourcePages_SourceIndices(t *testing.T) {
	sp := NewSourcePages()
	sp.Record(2, 0, 1)
	sp.Record(5, 2, 3)
	sp.Record(1, 4, 5)

	indices := sp.SourceIndices()
	if len(indices) != 3 {
		t.Fatalf("SourceIndices len = %d, want 3", len(indices))
	}
	// Should be in insertion order.
	if indices[0] != 2 || indices[1] != 5 || indices[2] != 1 {
		t.Errorf("SourceIndices = %v, want [2 5 1]", indices)
	}
}

func TestSourcePages_RemoveLast(t *testing.T) {
	sp := NewSourcePages()
	sp.Record(0, 0, 1)
	sp.Record(1, 2, 3)
	sp.RemoveLast()
	if sp.Count() != 1 {
		t.Errorf("after RemoveLast: Count = %d, want 1", sp.Count())
	}
	// Remaining entry should be index 0.
	first, _, ok := sp.Range(0)
	if !ok || first != 0 {
		t.Error("remaining entry should be index 0")
	}
}

func TestSourcePages_RemoveLast_Empty(t *testing.T) {
	sp := NewSourcePages()
	sp.RemoveLast() // should not panic
	if sp.Count() != 0 {
		t.Errorf("Count = %d after RemoveLast on empty", sp.Count())
	}
}

func TestSourcePages_IndexOf_Found(t *testing.T) {
	sp := NewSourcePages()
	sp.Record(10, 0, 1)
	sp.Record(20, 2, 3)
	sp.Record(30, 4, 5)

	if got := sp.IndexOf(10); got != 0 {
		t.Errorf("IndexOf(10) = %d, want 0", got)
	}
	if got := sp.IndexOf(20); got != 1 {
		t.Errorf("IndexOf(20) = %d, want 1", got)
	}
	if got := sp.IndexOf(30); got != 2 {
		t.Errorf("IndexOf(30) = %d, want 2", got)
	}
}

func TestSourcePages_IndexOf_NotFound(t *testing.T) {
	sp := NewSourcePages()
	sp.Record(1, 0, 1)

	if got := sp.IndexOf(99); got != -1 {
		t.Errorf("IndexOf(missing) = %d, want -1", got)
	}
}

func TestSourcePages_IndexOf_Empty(t *testing.T) {
	sp := NewSourcePages()
	if got := sp.IndexOf(0); got != -1 {
		t.Errorf("IndexOf on empty SourcePages = %d, want -1", got)
	}
}

func TestSourcePages_Get_Valid(t *testing.T) {
	sp := NewSourcePages()
	sp.Record(42, 0, 1)
	sp.Record(7, 2, 3)

	if got := sp.Get(0); got != 42 {
		t.Errorf("Get(0) = %d, want 42", got)
	}
	if got := sp.Get(1); got != 7 {
		t.Errorf("Get(1) = %d, want 7", got)
	}
}

func TestSourcePages_Get_OutOfRange(t *testing.T) {
	sp := NewSourcePages()
	sp.Record(5, 0, 1)

	if got := sp.Get(-1); got != -1 {
		t.Errorf("Get(-1) = %d, want -1", got)
	}
	if got := sp.Get(1); got != -1 {
		t.Errorf("Get(1) = %d on single-entry list, want -1", got)
	}
}

func TestSourcePages_Get_Empty(t *testing.T) {
	sp := NewSourcePages()
	if got := sp.Get(0); got != -1 {
		t.Errorf("Get(0) on empty SourcePages = %d, want -1", got)
	}
}

func TestSourcePages_Dispose(t *testing.T) {
	sp := NewSourcePages()
	sp.Record(0, 0, 1)
	sp.Record(1, 2, 3)
	sp.Dispose()
	if sp.Count() != 0 {
		t.Errorf("after Dispose: Count = %d, want 0", sp.Count())
	}
}

func TestSourcePages_ApplyPageSize(t *testing.T) {
	sp := NewSourcePages()
	sp.Record(0, 0, 1)
	// ApplyPageSize is a no-op stub; verify it does not panic or modify state.
	sp.ApplyPageSize()
	if sp.Count() != 1 {
		t.Errorf("after ApplyPageSize: Count = %d, want 1", sp.Count())
	}
}
