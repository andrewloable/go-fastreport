package data_test

// targeted_coverage_test.go — closes remaining coverage gaps in:
//   - datacomponent.go:76 InitializeComponent (empty function body; covered by calling it)
//   - filtered.go:84 seekInner (inner.Next error during seek loop)
//   - view.go:60 First (rebuildIndex error path; structurally unreachable, covered by proxy)

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// ── DataComponentBase.InitializeComponent ────────────────────────────────────
//
// InitializeComponent has an empty body ({}); the Go coverage tool does not
// insert instrumentation into empty function bodies, so the metric will remain
// at 0.0% regardless of how many times it is called. The test below documents
// that the function is callable and does not panic, satisfying correctness
// requirements even though the coverage number cannot change.

func TestDataComponentBase_InitializeComponent_Called(t *testing.T) {
	d := data.NewDataComponentBase("comp")
	// Calling InitializeComponent must never panic.
	d.InitializeComponent()
	// Calling it multiple times is also safe.
	d.InitializeComponent()
	d.InitializeComponent()
}

// ── FilteredDataSource.seekInner — inner.Next error during loop ───────────────
//
// seekInner advances inner from row 0 to the target row by calling Next()
// repeatedly.  The error-return inside the loop (filtered.go:93-94) fires when
// inner.Next() returns an error while the cursor is still below the target row.
//
// To trigger this we need the first matching row to be at inner index > 0 so
// that seekInner calls Next() at least once, AND inner.Next() must return an
// error on that call.

// seekNextFailSource is a DataSource that lets us inject a Next() failure
// after a configurable number of successful Next() calls per seekInner round.
// Each call to First() resets the failure counter so we can allow the
// rebuildIndex scan to succeed while failing during the subsequent seekInner.
type seekNextFailSource struct {
	rows      []map[string]any
	cursor    int
	failNextN int // fail Next() on the Nth call in the current round
	nextCount int // how many Next() calls have happened since last First()
}

func newSeekNextFailSource(rows []map[string]any) *seekNextFailSource {
	return &seekNextFailSource{rows: rows, cursor: -1, failNextN: -1}
}

func (s *seekNextFailSource) Name() string  { return "seekFail" }
func (s *seekNextFailSource) Alias() string { return "seekFail" }
func (s *seekNextFailSource) Init() error {
	s.cursor = -1
	s.nextCount = 0
	return nil
}
func (s *seekNextFailSource) First() error {
	s.cursor = 0
	s.nextCount = 0 // reset per-round counter
	if len(s.rows) == 0 {
		return data.ErrEOF
	}
	return nil
}
func (s *seekNextFailSource) Next() error {
	s.nextCount++
	if s.failNextN >= 0 && s.nextCount >= s.failNextN {
		return errors.New("seekNextFailSource: injected Next error")
	}
	s.cursor++
	if s.cursor >= len(s.rows) {
		return data.ErrEOF
	}
	return nil
}
func (s *seekNextFailSource) EOF() bool         { return s.cursor >= len(s.rows) }
func (s *seekNextFailSource) RowCount() int     { return len(s.rows) }
func (s *seekNextFailSource) CurrentRowNo() int { return s.cursor }
func (s *seekNextFailSource) GetValue(col string) (any, error) {
	if s.cursor < 0 || s.cursor >= len(s.rows) {
		return nil, errors.New("out of range")
	}
	v, ok := s.rows[s.cursor][col]
	if !ok {
		return nil, nil
	}
	return v, nil
}
func (s *seekNextFailSource) Close() error { return nil }

// TestFilteredDataSource_SeekInner_NextError exercises the error-return path
// inside the seekInner for-loop (filtered.go lines 93-94).
//
// Setup:
//   - 3 inner rows; id="a","b","c"
//   - filter on id=="c" → only inner row 2 matches → rows=[2]
//   - rebuildIndex runs with failNextN=-1 (no failure) → builds index
//   - Then we set failNextN=1 so that the 1st Next() call in seekInner fails
//   - fds.First() → cursor=0 → seekInner(target=2) → First() ok →
//     loop: CurrentRowNo(0)<2 → Next() #1 → ERROR → return err
func TestFilteredDataSource_SeekInner_NextError(t *testing.T) {
	inner := newSeekNextFailSource([]map[string]any{
		{"id": "a"}, // inner row 0
		{"id": "b"}, // inner row 1
		{"id": "c"}, // inner row 2
	})

	// Build the filtered index with no failures so rebuildIndex succeeds.
	fds, err := data.NewFilteredDataSource(inner, []string{"id"}, []string{"c"})
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	if fds.RowCount() != 1 {
		t.Fatalf("expected 1 matching row, got %d", fds.RowCount())
	}

	// Inject a failure on the 1st Next() call so seekInner's loop fails.
	inner.failNextN = 1

	err = fds.First()
	if err == nil {
		t.Error("First() should propagate seekInner Next() error")
	}
}

// TestFilteredDataSource_SeekInner_NextError_ViaNext exercises the same
// seekInner error path via fds.Next() instead of fds.First().
func TestFilteredDataSource_SeekInner_NextError_ViaNext(t *testing.T) {
	inner := newSeekNextFailSource([]map[string]any{
		{"id": "a"},      // inner row 0
		{"id": "b"},      // inner row 1
		{"id": "target"}, // inner row 2
		{"id": "c"},      // inner row 3
		{"id": "target"}, // inner row 4
	})

	// Filter for "target" rows → rows = [2, 4].
	fds, err := data.NewFilteredDataSource(inner, []string{"id"}, []string{"target"})
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	if fds.RowCount() != 2 {
		t.Fatalf("expected 2 matching rows, got %d", fds.RowCount())
	}

	// First call succeeds normally (no injected failure yet).
	if err := fds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}

	// Now inject a failure so the next seekInner (targeting inner row 4) fails.
	inner.failNextN = 1

	err = fds.Next()
	if err == nil {
		t.Error("Next() should propagate seekInner Next() error when advancing to row 4")
	}
}

// ── ViewDataSource.First — rebuildIndex error path ───────────────────────────
//
// view.go First() (line 60-68):
//
//	func (v *ViewDataSource) First() error {
//	    if !v.initDone {
//	        if err := v.rebuildIndex(); err != nil {
//	            return err            // ← line 63: only reached if rebuildIndex errors
//	        }
//	    }
//	    v.cursor = -1
//	    return nil
//	}
//
// rebuildIndex() always returns nil in the current implementation (it never
// propagates errors, only uses inner.First() error to decide if the source is
// empty).  Therefore line 63 is structurally unreachable through the public API.
//
// The test below documents this: we call First() before Init() on a fresh
// ViewDataSource (so initDone=false) and verify it succeeds, demonstrating
// that the rebuildIndex call path inside First() IS exercised (covering
// statements 1-2 of the branch), while confirming the error arm (stmt 3)
// cannot be triggered.

func TestViewDataSource_First_InitDoneFalse_RebuildSucceeds(t *testing.T) {
	inner := data.NewBaseDataSource("inner")
	inner.AddRow(map[string]any{"v": 1})
	inner.AddRow(map[string]any{"v": 2})
	// Initialize the inner source so rebuildIndex can call inner.First().
	if err := inner.Init(); err != nil {
		t.Fatalf("inner.Init: %v", err)
	}

	vds := data.NewViewDataSource(inner, "v", "V", "", nil)

	// initDone is false → First() enters the rebuildIndex branch.
	if err := vds.First(); err != nil {
		t.Fatalf("First (initDone=false): %v", err)
	}
	// After the rebuild, rows should be indexed.
	if vds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", vds.RowCount())
	}
	// Calling First() again: initDone=true now → skips rebuildIndex.
	if err := vds.First(); err != nil {
		t.Fatalf("Second First (initDone=true): %v", err)
	}
}

// TestViewDataSource_First_RepeatedCallsAfterSetFilter verifies that SetFilter
// resets initDone so the next First() call re-enters the rebuildIndex branch.
func TestViewDataSource_First_RepeatedCallsAfterSetFilter(t *testing.T) {
	inner := data.NewBaseDataSource("inner")
	inner.AddRow(map[string]any{"active": true})
	inner.AddRow(map[string]any{"active": false})
	inner.AddRow(map[string]any{"active": true})

	eval := func(expr string, src data.DataSource) (bool, error) {
		v, _ := src.GetValue("active")
		b, _ := v.(bool)
		return b, nil
	}
	vds := data.NewViewDataSource(inner, "v", "V", "active", eval)
	if err := vds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if vds.RowCount() != 2 {
		t.Fatalf("RowCount after Init = %d, want 2", vds.RowCount())
	}

	// SetFilter clears initDone; First() will call rebuildIndex again.
	vds.SetFilter("")
	if err := vds.First(); err != nil {
		t.Fatalf("First after SetFilter: %v", err)
	}
	// With empty filter, all 3 rows pass.
	if vds.RowCount() != 3 {
		t.Errorf("RowCount after SetFilter+First = %d, want 3", vds.RowCount())
	}
}
