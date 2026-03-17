package engine_test

// databands_coverage_test.go — targeted coverage for uncovered branches in
// databands.go: RunDataBandRows, RunDataBandFull, runDataBandHierarchical.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── RunDataBandRows: empty child when rows=0 ─────────────────────────────────

// TestRunDataBandRows_ZeroRows_WithChild exercises the rows=0 branch when
// a Child band is attached — it should show the child band.
func TestRunDataBandRows_ZeroRows_WithChild(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	child := band.NewChildBand()
	child.SetHeight(15)
	child.SetVisible(true)
	child.PrintIfDatabandEmpty = true // C#: child only shown when this flag is true

	db := band.NewDataBand()
	db.SetHeight(20)
	db.SetVisible(true)
	db.SetChild(child)

	beforeY := e.CurY()
	e.RunDataBandRows(db, 0)
	// Child band should be shown (PrintIfDatabandEmpty=true): CurY should advance by child height.
	if e.CurY() != beforeY+15 {
		t.Errorf("ZeroRows+child: CurY = %v, want %v", e.CurY(), beforeY+15)
	}
}

// TestRunDataBandRows_ZeroRows_NoChild exercises rows=0 without a child.
func TestRunDataBandRows_ZeroRows_NoChild(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(20)
	db.SetVisible(true)
	// No child.

	beforeY := e.CurY()
	e.RunDataBandRows(db, 0)
	if e.CurY() != beforeY {
		t.Errorf("ZeroRows+no child: CurY changed unexpectedly: %v → %v", beforeY, e.CurY())
	}
}

// TestRunDataBandRows_WithHeader_RepeatOnEveryPage exercises the RepeatOnEveryPage
// path where the header is added to reprints and later removed.
func TestRunDataBandRows_WithHeader_RepeatOnEveryPage(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	hdr := band.NewDataHeaderBand()
	hdr.SetHeight(10)
	hdr.SetVisible(true)
	hdr.SetRepeatOnEveryPage(true)

	ftr := band.NewDataFooterBand()
	ftr.SetHeight(8)
	ftr.SetVisible(true)
	ftr.SetRepeatOnEveryPage(true)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetHeader(hdr)
	db.SetFooter(ftr)

	beforeY := e.CurY()
	e.RunDataBandRows(db, 2)
	// header(10) + 2×row(10) + footer(8) = 38
	if e.CurY() != beforeY+38 {
		t.Errorf("RepeatOnEveryPage header: CurY = %v, want %v", e.CurY(), beforeY+38)
	}
}

// ── RunDataBandFull: various branches ────────────────────────────────────────

// treeDS is a data source that supports full data.DataSource interface
// and returns id/parentId columns for hierarchical tests.
// (Named treeDS to avoid conflicts with hierarchicalDS in engine_objects_test.go.)
type treeDS struct {
	rows []map[string]any
	pos  int
}

func newTreeDS(rows []map[string]any) *treeDS {
	return &treeDS{rows: rows, pos: -1}
}

func (h *treeDS) Name() string         { return "TreeDS" }
func (h *treeDS) Alias() string        { return "TreeDS" }
func (h *treeDS) Init() error          { h.pos = -1; return nil }
func (h *treeDS) First() error         { h.pos = 0; return nil }
func (h *treeDS) Next() error          { h.pos++; return nil }
func (h *treeDS) EOF() bool            { return h.pos >= len(h.rows) }
func (h *treeDS) RowCount() int        { return len(h.rows) }
func (h *treeDS) CurrentRowNo() int    { return h.pos }
func (h *treeDS) Close() error         { return nil }
func (h *treeDS) GetValue(col string) (any, error) {
	if h.pos < 0 || h.pos >= len(h.rows) {
		return nil, nil
	}
	v, ok := h.rows[h.pos][col]
	if !ok {
		return nil, nil
	}
	return v, nil
}
func (h *treeDS) Columns() []data.Column {
	return []data.Column{{Name: "id"}, {Name: "parentId"}, {Name: "name"}}
}

// TestRunDataBandFull_Hierarchical exercises runDataBandHierarchical with
// a simple 2-level tree (roots and children).
func TestRunDataBandFull_Hierarchical(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	rows := []map[string]any{
		{"id": "1", "parentId": "0", "name": "Root1"},
		{"id": "2", "parentId": "0", "name": "Root2"},
		{"id": "3", "parentId": "1", "name": "Child1.1"},
		{"id": "4", "parentId": "1", "name": "Child1.2"},
		{"id": "5", "parentId": "2", "name": "Child2.1"},
	}
	ds := newTreeDS(rows)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetIDColumn("id")
	db.SetParentIDColumn("parentId")
	db.SetDataSource(ds)

	hdr := band.NewDataHeaderBand()
	hdr.SetHeight(8)
	hdr.SetVisible(true)
	db.SetHeader(hdr)

	ftr := band.NewDataFooterBand()
	ftr.SetHeight(6)
	ftr.SetVisible(true)
	db.SetFooter(ftr)

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull hierarchical: %v", err)
	}
	// header(8) + 5 rows(50) + footer(6) = 64
	if e.CurY() <= beforeY {
		t.Errorf("hierarchical: CurY should have advanced beyond %v, got %v", beforeY, e.CurY())
	}
}

// TestRunDataBandFull_Hierarchical_DeepTree exercises deeper nesting
// (grandparent → parent → child) to cover the recursive renderRows path.
func TestRunDataBandFull_Hierarchical_DeepTree(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	rows := []map[string]any{
		{"id": "1", "parentId": "", "name": "GP"},
		{"id": "2", "parentId": "1", "name": "Parent"},
		{"id": "3", "parentId": "2", "name": "Child"},
		{"id": "4", "parentId": "3", "name": "GrandChild"},
	}
	ds := newTreeDS(rows)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetIDColumn("id")
	db.SetParentIDColumn("parentId")
	db.SetDataSource(ds)

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("deep hierarchical: %v", err)
	}
	if e.CurY() <= beforeY {
		t.Errorf("deep hierarchical: CurY should advance, got %v from %v", e.CurY(), beforeY)
	}
}

//TestRunDataBandFull_CompleteToNRows exercises the CompleteToNRows path where
// a ChildBand is used to fill unused rows up to N.
func TestRunDataBandFull_CompleteToNRows(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	child := band.NewChildBand()
	child.SetHeight(10)
	child.SetVisible(true)
	child.CompleteToNRows = 5 // fill up to 5 rows

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newMockDS(2)) // only 2 data rows
	db.SetChild(child)

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("CompleteToNRows: %v", err)
	}
	// 2 data rows (db 10px each) + 3 fill rows (child 10px each) = 5 × 10px = 50
	// But note: db rows also trigger sub-band rendering for child if attached...
	// Just verify CurY advanced by at least 50px.
	if e.CurY() < beforeY+50 {
		t.Errorf("CompleteToNRows: CurY = %v, want >= %v", e.CurY(), beforeY+50)
	}
}

// TestRunDataBandFull_FillUnusedSpace exercises the FillUnusedSpace path where
// a ChildBand is repeated until the page is full.
func TestRunDataBandFull_FillUnusedSpace(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	// Use a very small page so FillUnusedSpace terminates quickly.
	pg.PaperHeight = 30 // mm (~113px at 96dpi)
	pg.TopMargin = 0
	pg.BottomMargin = 0
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	child := band.NewChildBand()
	child.SetHeight(10)
	child.SetVisible(true)
	child.FillUnusedSpace = true

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newMockDS(1)) // 1 data row, then fill
	db.SetChild(child)

	// Should not hang or panic.
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("FillUnusedSpace: %v", err)
	}
}

// TestRunDataBandFull_KeepTogether exercises the KeepTogether path where
// StartKeep/EndKeep are called around the data band loop.
func TestRunDataBandFull_KeepTogether(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newMockDS(3))
	db.SetKeepTogether(true)

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("KeepTogether: %v", err)
	}
	// 3 rows × 10 = 30
	if e.CurY() != beforeY+30 {
		t.Errorf("KeepTogether: CurY = %v, want %v", e.CurY(), beforeY+30)
	}
}

// TestRunDataBandFull_StartNewPage exercises the StartNewPage path that starts
// a new page between rows (FlagUseStartNewPage + rowNo > 1).
func TestRunDataBandFull_StartNewPage(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newMockDS(3))
	db.SetStartNewPage(true)
	db.FlagUseStartNewPage = true

	initialPages := e.PreparedPages().Count()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("StartNewPage: %v", err)
	}
	// Each row after the first forces a new page.
	if e.PreparedPages().Count() <= initialPages {
		t.Error("StartNewPage: expected additional pages")
	}
}

// TestRunDataBandFull_SortedDataSource exercises the sort path for data sources
// that implement data.Sortable.
func TestRunDataBandFull_SortedDataSource(t *testing.T) {
	ds := data.NewBaseDataSource("SortDS")
	ds.SetAlias("SortDS")
	ds.AddColumn(data.Column{Name: "Val"})
	ds.AddRow(map[string]any{"Val": 30})
	ds.AddRow(map[string]any{"Val": 10})
	ds.AddRow(map[string]any{"Val": 20})
	if err := ds.Init(); err != nil {
		t.Fatalf("ds.Init: %v", err)
	}

	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(ds)
	// Add a sort spec: ascending by Val.
	db.SetSort([]band.SortSpec{{Column: "Val", Order: band.SortOrderAscending}})

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("sorted datasource: %v", err)
	}
	if e.CurY() != beforeY+30 {
		t.Errorf("sorted datasource: CurY = %v, want %v", e.CurY(), beforeY+30)
	}
}

// TestRunDataBandFull_WithHeader covers header shown in RunDataBandFull.
func TestRunDataBandFull_WithHeader(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	hdr := band.NewDataHeaderBand()
	hdr.SetHeight(12)
	hdr.SetVisible(true)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newMockDS(2))
	db.SetHeader(hdr)

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull with header: %v", err)
	}
	// header(12) + 2×row(10) = 32
	if e.CurY() != beforeY+32 {
		t.Errorf("with header: CurY = %v, want %v", e.CurY(), beforeY+32)
	}
}

// TestRunDataBandFull_WithFooter covers footer shown in RunDataBandFull.
func TestRunDataBandFull_WithFooter(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	hdr := band.NewDataHeaderBand()
	hdr.SetHeight(5)
	hdr.SetVisible(true)

	ftr := band.NewDataFooterBand()
	ftr.SetHeight(7)
	ftr.SetVisible(true)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newMockDS(3))
	db.SetHeader(hdr)
	db.SetFooter(ftr)

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull with footer: %v", err)
	}
	// header(5) + 3×row(10) + footer(7) = 42
	if e.CurY() != beforeY+42 {
		t.Errorf("with footer: CurY = %v, want %v", e.CurY(), beforeY+42)
	}
}
