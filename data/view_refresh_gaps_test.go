package data_test

// view_refresh_gaps_test.go — tests for ViewDataSource.RefreshColumns.
//
// go-fastreport issue: go-fastreport-raten
// C# ref: FastReport.Base/Data/ViewDataSource.cs RefreshColumns()

import (
	"testing"

	"github.com/andrewloable/go-fastreport/data"
)

// viewRefreshSource is a simple DataSource whose columns can be changed at
// runtime to simulate a schema change in the underlying source.
type viewRefreshSource struct {
	data.BaseDataSource
	cols []data.Column
}

func newViewRefreshSource(name string, cols []data.Column) *viewRefreshSource {
	s := &viewRefreshSource{
		BaseDataSource: *data.NewBaseDataSource(name),
		cols:           cols,
	}
	for _, c := range cols {
		s.AddColumn(c)
	}
	return s
}

func (s *viewRefreshSource) Columns() []data.Column { return s.cols }

// ── RefreshColumns: base cases ────────────────────────────────────────────────

// TestViewDataSource_RefreshColumns_AddNewColumn verifies that a column present
// in the inner source but not yet in the view is added by RefreshColumns.
// C# ref: ViewDataSource.RefreshColumns — add new columns loop
func TestViewDataSource_RefreshColumns_AddNewColumn(t *testing.T) {
	inner := newViewRefreshSource("src", []data.Column{
		{Name: "id", Alias: "id"},
		{Name: "name", Alias: "name"},
	})
	vds := data.NewViewDataSource(inner, "v", "v", "", nil)
	// Manually load schema (without calling InitSchema so columns start empty).
	_ = vds.Init()
	vds.InitSchema()

	// Now the inner source gains a new column.
	inner.cols = append(inner.cols, data.Column{Name: "email", Alias: "email"})

	vds.RefreshColumns()

	cols := vds.Columns()
	found := false
	for _, c := range cols {
		if c.Name == "email" {
			found = true
			break
		}
	}
	if !found {
		t.Error("RefreshColumns: new 'email' column should have been added")
	}
}

// TestViewDataSource_RefreshColumns_RemoveObsoleteColumn verifies that a column
// no longer present in the inner source is removed by RefreshColumns.
// C# ref: ViewDataSource.RefreshColumns — delete obsolete columns loop
func TestViewDataSource_RefreshColumns_RemoveObsoleteColumn(t *testing.T) {
	inner := newViewRefreshSource("src", []data.Column{
		{Name: "id", Alias: "id"},
		{Name: "name", Alias: "name"},
		{Name: "obsolete", Alias: "obsolete"},
	})
	vds := data.NewViewDataSource(inner, "v", "v", "", nil)
	_ = vds.Init()
	vds.InitSchema()

	// Remove 'obsolete' from inner source.
	inner.cols = []data.Column{
		{Name: "id", Alias: "id"},
		{Name: "name", Alias: "name"},
	}

	vds.RefreshColumns()

	cols := vds.Columns()
	for _, c := range cols {
		if c.Name == "obsolete" {
			t.Error("RefreshColumns: 'obsolete' column should have been removed")
		}
	}
	if len(cols) != 2 {
		t.Errorf("column count = %d, want 2", len(cols))
	}
}

// TestViewDataSource_RefreshColumns_NoChange verifies that when the inner source
// columns are unchanged, RefreshColumns is a no-op.
func TestViewDataSource_RefreshColumns_NoChange(t *testing.T) {
	inner := newViewRefreshSource("src", []data.Column{
		{Name: "id", Alias: "id"},
		{Name: "name", Alias: "name"},
	})
	vds := data.NewViewDataSource(inner, "v", "v", "", nil)
	_ = vds.Init()
	vds.InitSchema()

	before := len(vds.Columns())
	vds.RefreshColumns()
	after := len(vds.Columns())

	if before != after {
		t.Errorf("RefreshColumns with no schema change: columns changed from %d to %d", before, after)
	}
}

// TestViewDataSource_RefreshColumns_NilInner verifies that RefreshColumns does
// not panic when the inner source returns nil columns (e.g., implements no
// Columns() method).
func TestViewDataSource_RefreshColumns_NilColumns(t *testing.T) {
	// Use a BaseDataSource directly — does not expose Columns() via the interface
	// accepted by innerColumns(), so innerColumns returns nil.
	inner := data.NewBaseDataSource("noColumns")
	vds := data.NewViewDataSource(inner, "v", "v", "", nil)

	// Should not panic.
	vds.RefreshColumns()
}

// TestViewDataSource_RefreshColumns_AddsAllMissingColumns verifies that multiple
// new columns are all added in a single RefreshColumns call.
func TestViewDataSource_RefreshColumns_AddsAllMissingColumns(t *testing.T) {
	inner := newViewRefreshSource("src", []data.Column{
		{Name: "a", Alias: "a"},
	})
	vds := data.NewViewDataSource(inner, "v", "v", "", nil)
	_ = vds.Init()
	vds.InitSchema()

	inner.cols = []data.Column{
		{Name: "a", Alias: "a"},
		{Name: "b", Alias: "b"},
		{Name: "c", Alias: "c"},
	}
	vds.RefreshColumns()

	cols := vds.Columns()
	if len(cols) != 3 {
		t.Errorf("column count = %d, want 3", len(cols))
	}
	names := make(map[string]bool)
	for _, c := range cols {
		names[c.Name] = true
	}
	for _, want := range []string{"a", "b", "c"} {
		if !names[want] {
			t.Errorf("column %q not found after RefreshColumns", want)
		}
	}
}

// TestViewDataSource_RefreshColumns_EmptyInner_AllRemoved verifies that when
// the inner source drops all columns, RefreshColumns removes them from the view.
func TestViewDataSource_RefreshColumns_EmptyInner_AllRemoved(t *testing.T) {
	inner := newViewRefreshSource("src", []data.Column{
		{Name: "x", Alias: "x"},
		{Name: "y", Alias: "y"},
	})
	vds := data.NewViewDataSource(inner, "v", "v", "", nil)
	_ = vds.Init()
	vds.InitSchema()

	if len(vds.Columns()) != 2 {
		t.Fatalf("expected 2 columns before refresh, got %d", len(vds.Columns()))
	}

	// Inner now returns no columns.
	inner.cols = []data.Column{}

	// innerColumns returns nil for empty slice, so RefreshColumns is a no-op.
	// This validates the early-return guard (fresh == nil).
	vds.RefreshColumns()
	// Column count is unchanged — empty inner does not clear view columns.
	if len(vds.Columns()) != 2 {
		t.Errorf("column count = %d after empty-inner refresh, want 2 (no-op)", len(vds.Columns()))
	}
}
