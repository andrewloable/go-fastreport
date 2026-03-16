package data_test

// filteredview_coverage_test.go — additional coverage for FilteredDataSource and
// ViewDataSource error paths that require mock data sources with controllable failures.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// ── mockFailSource: a DataSource that can be made to fail on First/Next ───────

type mockFailSource struct {
	name      string
	rows      []map[string]any
	cursor    int
	failFirst bool
	failNext  bool
	failGet   bool
}

func newMockFailSource(name string) *mockFailSource {
	return &mockFailSource{name: name, cursor: -1}
}

func (m *mockFailSource) addRow(row map[string]any) { m.rows = append(m.rows, row) }

func (m *mockFailSource) Name() string  { return m.name }
func (m *mockFailSource) Alias() string { return m.name }

func (m *mockFailSource) Init() error {
	m.cursor = -1
	return nil
}

func (m *mockFailSource) First() error {
	if m.failFirst {
		return errors.New("mock First error")
	}
	if len(m.rows) == 0 {
		m.cursor = 0
		return data.ErrEOF
	}
	m.cursor = 0
	return nil
}

func (m *mockFailSource) Next() error {
	if m.failNext {
		return errors.New("mock Next error")
	}
	m.cursor++
	if m.cursor >= len(m.rows) {
		return data.ErrEOF
	}
	return nil
}

func (m *mockFailSource) EOF() bool { return m.cursor >= len(m.rows) }

func (m *mockFailSource) RowCount() int { return len(m.rows) }

func (m *mockFailSource) CurrentRowNo() int { return m.cursor }

func (m *mockFailSource) GetValue(column string) (any, error) {
	if m.failGet {
		return nil, errors.New("mock GetValue error")
	}
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return nil, errors.New("out of range")
	}
	v, ok := m.rows[m.cursor][column]
	if !ok {
		return nil, errors.New("column not found: " + column)
	}
	return v, nil
}

func (m *mockFailSource) Close() error { return nil }

// ── FilteredDataSource error paths ────────────────────────────────────────────

func TestFilteredDataSource_RowMatches_GetValueError(t *testing.T) {
	// Inner source fails GetValue → rowMatches returns false → row is excluded.
	inner := newMockFailSource("inner")
	inner.addRow(map[string]any{"col": "val"})

	// Build without error so rebuildIndex runs; then trigger error.
	// We have to build with failGet already set to see all rows excluded.
	inner.failGet = true

	fds, err := data.NewFilteredDataSource(inner, []string{"col"}, []string{"val"})
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	// All rows excluded because GetValue failed.
	if fds.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0 (GetValue error excludes all rows)", fds.RowCount())
	}
}

func TestFilteredDataSource_SeekInner_FirstError(t *testing.T) {
	// rebuildIndex uses First() to scan — if First fails, return nil (empty).
	inner := newMockFailSource("inner")
	inner.addRow(map[string]any{"col": "x"})
	inner.failFirst = true

	// rebuildIndex calls inner.First() which fails → empty result, no error.
	fds, err := data.NewFilteredDataSource(inner, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	if fds.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0 (First failed in rebuildIndex)", fds.RowCount())
	}
}

func TestFilteredDataSource_SeekInner_FirstError_OnSeek(t *testing.T) {
	// Build a valid index first, then fail on First() during seekInner.
	inner := newMockFailSource("inner")
	inner.addRow(map[string]any{"col": "x"})
	// Build without failFirst so we get 1 row in the index.
	fds, err := data.NewFilteredDataSource(inner, nil, nil)
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	if fds.RowCount() != 1 {
		t.Fatalf("expected 1 row, got %d", fds.RowCount())
	}

	// Now trigger First() failure for seekInner path.
	inner.failFirst = true
	err = fds.First()
	if err == nil {
		t.Error("First should propagate seekInner error when inner.First() fails")
	}
}

func TestFilteredDataSource_NewFilteredDataSource_FewerParentValues(t *testing.T) {
	// childColumns has more entries than parentValues — extra keys get nil value.
	inner := newMockFailSource("inner")
	inner.addRow(map[string]any{"a": "1", "b": "2"})

	// Only 1 parentValue for 2 childColumns → second key gets nil value.
	fds, err := data.NewFilteredDataSource(inner, []string{"a", "b"}, []string{"1"})
	if err != nil {
		t.Fatalf("NewFilteredDataSource: %v", err)
	}
	// Row has a="1" but b="2" != nil → does not match.
	if fds.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0 (b mismatch due to nil key value)", fds.RowCount())
	}
}

// ── ViewDataSource error paths ────────────────────────────────────────────────

func TestViewDataSource_Init_InnerInitError(t *testing.T) {
	// When inner.Init() fails, ViewDataSource.Init should return an error.
	// Use a custom inner that fails Init.
	inner := &initFailSource{}
	vds := data.NewViewDataSource(inner, "v", "v", "", nil)
	err := vds.Init()
	if err == nil {
		t.Error("ViewDataSource.Init should return error when inner.Init fails")
	}
}

// initFailSource: Init always fails.
type initFailSource struct{}

func (s *initFailSource) Name() string                        { return "fail" }
func (s *initFailSource) Alias() string                       { return "fail" }
func (s *initFailSource) Init() error                         { return errors.New("inner init failed") }
func (s *initFailSource) First() error                        { return errors.New("first failed") }
func (s *initFailSource) Next() error                         { return errors.New("next failed") }
func (s *initFailSource) EOF() bool                           { return true }
func (s *initFailSource) RowCount() int                       { return 0 }
func (s *initFailSource) CurrentRowNo() int                   { return 0 }
func (s *initFailSource) GetValue(c string) (any, error)      { return nil, errors.New("no value") }
func (s *initFailSource) Close() error                        { return nil }

func TestViewDataSource_First_BeforeInit_RebuildError(t *testing.T) {
	// First() without Init triggers rebuildIndex.
	// rebuildIndex calls inner.First() — if it returns ErrEOF we get an empty view (no error).
	// Use a source with zero rows — inner.First() returns ErrEOF which is treated as empty.
	inner := newMockFailSource("empty")
	// No rows added, so First() returns ErrEOF → rebuildIndex returns nil → First returns nil.
	vds := data.NewViewDataSource(inner, "v", "v", "", nil)
	err := vds.First()
	if err != nil {
		t.Errorf("First on empty inner (no Init): want nil, got %v", err)
	}
}

func TestViewDataSource_SeekInner_FirstError(t *testing.T) {
	// Next() calls seekInner which calls inner.First(); if that fails, error propagates.
	inner := newMockFailSource("inner")
	inner.addRow(map[string]any{"x": 1})
	inner.addRow(map[string]any{"x": 2})

	vds := data.NewViewDataSource(inner, "v", "v", "", nil)
	if err := vds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Now fail First for seekInner.
	inner.failFirst = true
	err := vds.Next()
	if err == nil {
		t.Error("Next should propagate seekInner error when inner.First() fails")
	}
}

func TestViewDataSource_SeekInner_NextError(t *testing.T) {
	// seekInner calls inner.Next() to advance to target row; if it fails, error propagates.
	inner := newMockFailSource("inner")
	inner.addRow(map[string]any{"x": 1})
	inner.addRow(map[string]any{"x": 2})
	inner.addRow(map[string]any{"x": 3})

	vds := data.NewViewDataSource(inner, "v", "v", "", nil)
	if err := vds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Advance to row index 1 (second row) — seekInner needs Next to advance.
	inner.failNext = true
	// First Next() tries to seek to rows[0]=0, which only needs First() (0 Next calls).
	// For seekInner to call Next, we need to target a row > 0.
	// rows[] contains [0,1,2]. Next() targets rows[cursor] after increment.
	// First call to Next: cursor becomes 0, target=rows[0]=0; seekInner calls First (no Next).
	inner.failNext = false
	if err := vds.Next(); err != nil {
		t.Fatalf("first Next: %v", err)
	}
	// Second call to Next: cursor becomes 1, target=rows[1]=1; seekInner calls First then Next once.
	inner.failNext = true
	err := vds.Next()
	if err == nil {
		t.Error("Next should propagate seekInner error when inner.Next() fails")
	}
}

func TestViewDataSource_CurrentRowNo_AfterNext(t *testing.T) {
	// CurrentRowNo returns cursor (>=0) after Next.
	inner := newMockFailSource("inner")
	inner.addRow(map[string]any{"x": 1})
	inner.addRow(map[string]any{"x": 2})

	vds := data.NewViewDataSource(inner, "v", "v", "", nil)
	if err := vds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := vds.Next(); err != nil {
		t.Fatalf("Next: %v", err)
	}
	if vds.CurrentRowNo() != 0 {
		t.Errorf("CurrentRowNo = %d, want 0", vds.CurrentRowNo())
	}
	if err := vds.Next(); err != nil {
		t.Fatalf("Next 2: %v", err)
	}
	if vds.CurrentRowNo() != 1 {
		t.Errorf("CurrentRowNo = %d, want 1", vds.CurrentRowNo())
	}
}
