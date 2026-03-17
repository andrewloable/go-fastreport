package data_test

// data_filtered_coverage_test.go — closes remaining coverage gaps in:
//   - filtered.go:84 seekInner — cursor < 0 early-return branch (structurally
//     unreachable through the public API; seekInner is only called from First/Next
//     which always set cursor >= 0 before calling seekInner)
//   - filtered.go:29 NewFilteredDataSource — rebuildIndex error branch
//     (rebuildIndex always returns nil; dead code guard)
//
// Since both branches are architecturally dead through the public API, this
// file instead adds supplementary tests that maximise coverage of every
// REACHABLE code path inside these functions to document their correctness.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// TestFilteredDataSource_NewFilteredDataSource_EmptyChildColumns verifies that
// NewFilteredDataSource works correctly when childColumns is empty (zero-length
// key list). In this case the loop body never executes, so every row of inner
// matches (no keys to fail) and all rows are included in the index.
func TestFilteredDataSource_NewFilteredDataSource_EmptyChildColumns(t *testing.T) {
	inner := data.NewBaseDataSource("items")
	inner.AddColumn(data.Column{Name: "id"})
	inner.AddRow(map[string]any{"id": "a"})
	inner.AddRow(map[string]any{"id": "b"})
	_ = inner.Init()

	fds, err := data.NewFilteredDataSource(inner, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource with empty keys: %v", err)
	}
	// No filter keys → all rows pass rowMatches.
	if fds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2 (no filter keys, all rows pass)", fds.RowCount())
	}
}

// TestFilteredDataSource_NewFilteredDataSource_ParentValuesLonger exercises the
// branch where i < len(parentValues) for every index (parentValues has exactly
// as many elements as childColumns). Each key gets the corresponding parent value.
func TestFilteredDataSource_NewFilteredDataSource_ParentValuesLonger(t *testing.T) {
	inner := data.NewBaseDataSource("items")
	inner.AddColumn(data.Column{Name: "type"})
	inner.AddRow(map[string]any{"type": "A"})
	inner.AddRow(map[string]any{"type": "B"})
	inner.AddRow(map[string]any{"type": "A"})
	_ = inner.Init()

	// One child column, one parent value — the `i < len(parentValues)` branch is taken.
	fds, err := data.NewFilteredDataSource(inner, []string{"type"}, []string{"A"})
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	if fds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2 (two rows with type=A)", fds.RowCount())
	}
	if err := fds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	v, err := fds.GetValue("type")
	if err != nil {
		t.Fatalf("GetValue: %v", err)
	}
	if v != "A" {
		t.Errorf("GetValue(type) = %v, want A", v)
	}
}

// TestFilteredDataSource_SeekInner_CursorAtBounds documents the fact that
// seekInner's cursor-bounds guard (line 85-87) is architecturally dead through
// the public API. seekInner is unexported; First() always sets cursor=0 before
// calling it, and Next() only calls seekInner when cursor < len(rows).
//
// This test exercises the Next()-past-EOF path to ensure the EOF check
// (cursor >= len rows) is properly handled in the caller and seekInner
// is never reached in that state.
func TestFilteredDataSource_SeekInner_NotCalledWhenCursorOutOfBounds(t *testing.T) {
	inner := data.NewBaseDataSource("items")
	inner.AddColumn(data.Column{Name: "v"})
	inner.AddRow(map[string]any{"v": 1})
	_ = inner.Init()

	fds, err := data.NewFilteredDataSource(inner, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	if fds.RowCount() != 1 {
		t.Fatalf("RowCount = %d, want 1", fds.RowCount())
	}

	// First() sets cursor=0, calls seekInner(cursor=0 which is in-bounds).
	if err := fds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}

	// Next() increments cursor to 1 (>= len(rows)=1) → returns ErrEOF BEFORE
	// calling seekInner, so seekInner is never called with cursor out of bounds.
	err = fds.Next()
	if err != data.ErrEOF {
		t.Errorf("Next past last row should return ErrEOF, got %v", err)
	}
	if !fds.EOF() {
		t.Error("EOF() should be true after Next past last row")
	}
}

// TestFilteredDataSource_NewFilteredDataSource_RebuildIndexEmptyInner exercises
// the path where inner.First() returns ErrEOF (empty inner source) inside
// rebuildIndex, which causes rebuildIndex to return nil and the FDS has 0 rows.
// This documents that the `if err := f.rebuildIndex(); err != nil` guard in
// NewFilteredDataSource (lines 41-43) is dead code — rebuildIndex only returns
// nil even when inner.First() fails.
func TestFilteredDataSource_NewFilteredDataSource_RebuildIndexEmptyInner(t *testing.T) {
	inner := data.NewBaseDataSource("empty")
	// No rows added; Init succeeds with zero rows.
	_ = inner.Init()

	// rebuildIndex calls inner.First() which returns ErrEOF → treats as empty → returns nil.
	// The NewFilteredDataSource error branch (lines 41-43) is never taken.
	fds, err := data.NewFilteredDataSource(inner, []string{"col"}, []string{"val"})
	if err != nil {
		t.Fatalf("NewFilteredDataSource with empty inner: %v", err)
	}
	if fds.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0 for empty inner", fds.RowCount())
	}
	// First() on empty FDS returns ErrEOF.
	if err := fds.First(); err != data.ErrEOF {
		t.Errorf("First on empty FDS: want ErrEOF, got %v", err)
	}
}

// TestFilteredDataSource_Name_Alias2 verifies the Name and Alias delegation.
func TestFilteredDataSource_Name_Alias2(t *testing.T) {
	inner := data.NewBaseDataSource("mySource")
	inner.SetAlias("myAlias")
	_ = inner.Init()

	fds, err := data.NewFilteredDataSource(inner, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	if fds.Name() != "mySource" {
		t.Errorf("Name = %q, want mySource", fds.Name())
	}
	if fds.Alias() != "myAlias" {
		t.Errorf("Alias = %q, want myAlias", fds.Alias())
	}
}

// TestFilteredDataSource_Init_IsNoOp verifies that Init() is a no-op and
// does not reset state or cause errors.
func TestFilteredDataSource_Init_IsNoOp(t *testing.T) {
	inner := data.NewBaseDataSource("src")
	inner.AddRow(map[string]any{"x": 1})
	_ = inner.Init()

	fds, err := data.NewFilteredDataSource(inner, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	// Init should be a no-op.
	if err := fds.Init(); err != nil {
		t.Errorf("Init: %v", err)
	}
	if fds.RowCount() != 1 {
		t.Errorf("RowCount after Init = %d, want 1", fds.RowCount())
	}
}

// TestFilteredDataSource_Columns_NoColumnsInterface verifies that Columns()
// returns nil when the inner source does not implement the columns interface.
func TestFilteredDataSource_Columns_NoColumnsInterface(t *testing.T) {
	// mockFailSource (from filteredview_coverage_test.go) does not implement Columns().
	inner := newMockFailSource("plain")
	inner.addRow(map[string]any{"a": 1})

	fds, err := data.NewFilteredDataSource(inner, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	// inner does not implement Columns() → returns nil.
	if cols := fds.Columns(); cols != nil {
		t.Errorf("Columns on non-column source = %v, want nil", cols)
	}
}

// TestFilteredDataSource_Columns_WithColumnsInterface verifies that Columns()
// delegates to inner when it implements the Columns() interface.
func TestFilteredDataSource_Columns_WithColumnsInterface(t *testing.T) {
	inner := data.NewBaseDataSource("colSrc")
	inner.AddColumn(data.Column{Name: "id", Alias: "ID"})
	inner.AddColumn(data.Column{Name: "name"})
	_ = inner.Init()

	fds, err := data.NewFilteredDataSource(inner, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	cols := fds.Columns()
	if len(cols) != 2 {
		t.Errorf("Columns len = %d, want 2", len(cols))
	}
}

// TestFilteredDataSource_Close_Delegates verifies that Close() delegates to inner.
func TestFilteredDataSource_Close_Delegates(t *testing.T) {
	inner := data.NewBaseDataSource("src")
	_ = inner.Init()

	fds, err := data.NewFilteredDataSource(inner, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	if err := fds.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

// TestFilteredDataSource_CurrentRowNo_BeforeFirst verifies CurrentRowNo before First.
func TestFilteredDataSource_CurrentRowNo_BeforeFirst(t *testing.T) {
	inner := data.NewBaseDataSource("src")
	inner.AddRow(map[string]any{"x": 1})
	_ = inner.Init()

	fds, err := data.NewFilteredDataSource(inner, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	// cursor starts at -1 after construction.
	if fds.CurrentRowNo() != -1 {
		t.Errorf("CurrentRowNo before First = %d, want -1", fds.CurrentRowNo())
	}
}

// TestFilteredDataSource_MultiKey_AllMatch exercises rowMatches with multiple
// filter keys where all keys match → row is included.
func TestFilteredDataSource_MultiKey_AllMatch(t *testing.T) {
	inner := data.NewBaseDataSource("items")
	inner.AddColumn(data.Column{Name: "cat"})
	inner.AddColumn(data.Column{Name: "status"})
	inner.AddRow(map[string]any{"cat": "fruit", "status": "fresh"})
	inner.AddRow(map[string]any{"cat": "fruit", "status": "stale"})
	inner.AddRow(map[string]any{"cat": "veggie", "status": "fresh"})
	_ = inner.Init()

	fds, err := data.NewFilteredDataSource(inner,
		[]string{"cat", "status"}, []string{"fruit", "fresh"})
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	if fds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1 (one row matches both keys)", fds.RowCount())
	}
	_ = fds.First()
	v, _ := fds.GetValue("cat")
	if v != "fruit" {
		t.Errorf("GetValue(cat) = %v, want fruit", v)
	}
	v2, _ := fds.GetValue("status")
	if v2 != "fresh" {
		t.Errorf("GetValue(status) = %v, want fresh", v2)
	}
}

// TestFilteredDataSource_SeekInner_NextError_MockFail exercises seekInner's inner.Next()
// error path (filtered.go:93-95). This requires target row index > 0 so seekInner
// must call Next to advance; if Next fails the error propagates.
func TestFilteredDataSource_SeekInner_NextError_MockFail(t *testing.T) {
	inner := newMockFailSource("inner")
	inner.addRow(map[string]any{"col": "x"})
	inner.addRow(map[string]any{"col": "y"})
	// Build with failNext=false so rebuildIndex captures both rows (indices 0 and 1).
	fds, err := data.NewFilteredDataSource(inner, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	if fds.RowCount() != 2 {
		t.Fatalf("RowCount = %d, want 2", fds.RowCount())
	}
	// First(): cursor=0, target=0, seekInner calls First only (CurrentRowNo=0 not < 0) → ok.
	if err := fds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}
	// Next(): cursor=1, target=1. seekInner calls First() (ok), then needs Next() to advance
	// from CurrentRowNo=0 to target=1. With failNext=true, Next() returns error.
	inner.failNext = true
	if err := fds.Next(); err == nil {
		t.Error("Next should propagate seekInner error when inner.Next() fails")
	}
}
