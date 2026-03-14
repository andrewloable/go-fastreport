package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// filterDS is a DataSource returning rows where each row has a "val" column.
type filterDS struct {
	rows []int
	pos  int
}

func newFilterDS(rows ...int) *filterDS { return &filterDS{rows: rows, pos: -1} }

func (d *filterDS) RowCount() int { return len(d.rows) }
func (d *filterDS) First() error  { d.pos = 0; return nil }
func (d *filterDS) Next() error   { d.pos++; return nil }
func (d *filterDS) EOF() bool     { return d.pos >= len(d.rows) }
func (d *filterDS) GetValue(col string) (any, error) {
	if d.pos < 0 || d.pos >= len(d.rows) {
		return nil, nil
	}
	if col == "val" {
		return d.rows[d.pos], nil
	}
	return nil, nil
}

// bandCount returns the number of PreparedBands on page 0 after running.
func bandCount(e *engine.ReportEngine) int {
	pg := e.PreparedPages().GetPage(0)
	if pg == nil {
		return 0
	}
	return len(pg.Bands)
}

func buildFilterEngine(t *testing.T) *engine.ReportEngine {
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

// ── No filter: all rows rendered ──────────────────────────────────────────────

func TestFilter_NoFilter_AllRowsRendered(t *testing.T) {
	e := buildFilterEngine(t)
	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newFilterDS(1, 2, 3, 4, 5))
	// No filter set.

	before := bandCount(e)
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull error: %v", err)
	}
	added := bandCount(e) - before
	if added != 5 {
		t.Errorf("expected 5 bands (no filter), got %d", added)
	}
}

// ── Filter with bracket expression ────────────────────────────────────────────

func TestFilter_BracketExpr_FiltersRows(t *testing.T) {
	e := buildFilterEngine(t)
	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(10)
	db.SetVisible(true)
	ds := newFilterDS(1, 2, 3, 4, 5)
	db.SetDataSource(ds)
	// Only render rows where val > 3 (rows 4 and 5).
	db.SetFilter("[val] > 3")

	before := bandCount(e)
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull error: %v", err)
	}
	added := bandCount(e) - before
	if added != 2 {
		t.Errorf("expected 2 bands (val>3), got %d", added)
	}
}

// ── Filter with bare expression ───────────────────────────────────────────────

func TestFilter_BareExpr_FiltersRows(t *testing.T) {
	e := buildFilterEngine(t)
	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(10)
	db.SetVisible(true)
	ds := newFilterDS(10, 20, 5, 15, 3)
	db.SetDataSource(ds)
	// Only rows where val >= 10.
	db.SetFilter("[val] >= 10")

	before := bandCount(e)
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull error: %v", err)
	}
	added := bandCount(e) - before
	// 10 ✓, 20 ✓, 5 ✗, 15 ✓, 3 ✗ → 3 rows pass
	if added != 3 {
		t.Errorf("expected 3 bands (val>=10), got %d", added)
	}
}

// ── Filter that excludes all rows ─────────────────────────────────────────────

func TestFilter_ExcludesAllRows(t *testing.T) {
	e := buildFilterEngine(t)
	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newFilterDS(1, 2, 3))
	db.SetFilter("[val] > 100") // nothing passes

	before := bandCount(e)
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull error: %v", err)
	}
	added := bandCount(e) - before
	if added != 0 {
		t.Errorf("expected 0 bands (all excluded), got %d", added)
	}
}

// ── extractBracketedNames helper ─────────────────────────────────────────────

func TestFilter_InvalidExpr_PassesThrough(t *testing.T) {
	// An expression that fails to compile should pass rows through (not crash).
	e := buildFilterEngine(t)
	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newFilterDS(1, 2))
	db.SetFilter("this is not valid +++")

	before := bandCount(e)
	// Should not panic.
	_ = e.RunDataBandFull(db)
	added := bandCount(e) - before
	// Pass-through behaviour: both rows rendered.
	if added != 2 {
		t.Errorf("expected 2 bands (invalid filter pass-through), got %d", added)
	}
}
