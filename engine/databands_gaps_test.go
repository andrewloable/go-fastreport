package engine_test

// databands_gaps_test.go — targeted tests for the remaining uncovered branches
// in databands.go after databands_test.go and databands_coverage_test.go.
//
// Uncovered blocks (by profile line range):
//   RunDataBandRows:
//     62.64–64.4   startNewPageForCurrent inside rows loop
//     76.58–78.4   return after runBands error inside rows loop
//
//   RunDataBandFull:
//     164.36–165.10  break after ds.Next() error in filter-skip path
//     208.46–211.4   restore()+return err when runBands errors on sub-band
//     217.16–218.9   break when e.aborted inside row loop
//
//   runDataBandHierarchical:
//     274.14–276.3   return nil when ds is not data.DataSource (!hasFull)
//     286.35–287.9   break in snapshot loop when ds.Next() errors
//     325.37–327.5   return err from ds.First() inside renderRows
//     329.37–330.11  break from ds.Next() error in seek loop inside renderRows
//     362.69–364.6   return err from nested renderRows call
//     370.49–372.3   return err from top-level renderRows call

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── helper: error data sources ────────────────────────────────────────────────

// bandOnlyDS satisfies band.DataSource but NOT data.DataSource.
// Used to trigger the !hasFull early-return in runDataBandHierarchical.
type bandOnlyDS struct {
	rows []map[string]any
	pos  int
}

func newBandOnlyDS(n int) *bandOnlyDS {
	rows := make([]map[string]any, n)
	for i := range rows {
		rows[i] = map[string]any{"id": i + 1, "parentId": 0}
	}
	return &bandOnlyDS{rows: rows, pos: 0}
}

func (d *bandOnlyDS) RowCount() int { return len(d.rows) }
func (d *bandOnlyDS) First() error  { d.pos = 0; return nil }
func (d *bandOnlyDS) Next() error   { d.pos++; return nil }
func (d *bandOnlyDS) EOF() bool     { return d.pos >= len(d.rows) }
func (d *bandOnlyDS) GetValue(col string) (any, error) {
	if d.pos < 0 || d.pos >= len(d.rows) {
		return nil, nil
	}
	return d.rows[d.pos][col], nil
}

// errNextAfterNDS errors on the Nth call to Next() (1-based).
type errNextAfterNDS struct {
	rows     []map[string]any
	pos      int
	nextCall int
	errOnN   int // error on this Next() call number
}

func newErrNextAfterNDS(rows []map[string]any, errOnN int) *errNextAfterNDS {
	return &errNextAfterNDS{rows: rows, pos: -1, errOnN: errOnN}
}

func (d *errNextAfterNDS) RowCount() int { return len(d.rows) }
func (d *errNextAfterNDS) First() error  { d.pos = 0; return nil }
func (d *errNextAfterNDS) Next() error {
	d.nextCall++
	if d.nextCall >= d.errOnN {
		return errors.New("intentional Next() error")
	}
	d.pos++
	return nil
}
func (d *errNextAfterNDS) EOF() bool { return d.pos >= len(d.rows) }
func (d *errNextAfterNDS) GetValue(col string) (any, error) {
	if d.pos < 0 || d.pos >= len(d.rows) {
		return nil, nil
	}
	v, ok := d.rows[d.pos][col]
	if !ok {
		return nil, nil
	}
	return v, nil
}

// errNextAfterNDSFull adds data.DataSource methods to errNextAfterNDS so it can
// be used in hierarchical rendering.
func (d *errNextAfterNDS) Name() string      { return "ErrNextDS" }
func (d *errNextAfterNDS) Alias() string     { return "ErrNextDS" }
func (d *errNextAfterNDS) Init() error       { d.pos = -1; d.nextCall = 0; return nil }
func (d *errNextAfterNDS) Close() error      { return nil }
func (d *errNextAfterNDS) CurrentRowNo() int { return d.pos }
func (d *errNextAfterNDS) Columns() []struct{ Name string } {
	return []struct{ Name string }{{Name: "id"}, {Name: "parentId"}}
}

// errFirstAfterNDS errors on the Nth First() call (1-based).
// Satisfies both band.DataSource and data.DataSource.
type errFirstAfterNDS struct {
	rows      []map[string]any
	pos       int
	firstCall int
	errOnN    int
}

func newErrFirstAfterNDS(rows []map[string]any, errOnN int) *errFirstAfterNDS {
	return &errFirstAfterNDS{rows: rows, pos: -1, errOnN: errOnN}
}

func (d *errFirstAfterNDS) RowCount() int { return len(d.rows) }
func (d *errFirstAfterNDS) First() error {
	d.firstCall++
	if d.firstCall >= d.errOnN {
		return errors.New("intentional First() error")
	}
	d.pos = 0
	return nil
}
func (d *errFirstAfterNDS) Next() error  { d.pos++; return nil }
func (d *errFirstAfterNDS) EOF() bool    { return d.pos >= len(d.rows) }
func (d *errFirstAfterNDS) GetValue(col string) (any, error) {
	if d.pos < 0 || d.pos >= len(d.rows) {
		return nil, nil
	}
	v, ok := d.rows[d.pos][col]
	if !ok {
		return nil, nil
	}
	return v, nil
}
func (d *errFirstAfterNDS) Name() string      { return "ErrFirstDS" }
func (d *errFirstAfterNDS) Alias() string     { return "ErrFirstDS" }
func (d *errFirstAfterNDS) Init() error       { d.pos = -1; d.firstCall = 0; return nil }
func (d *errFirstAfterNDS) Close() error      { return nil }
func (d *errFirstAfterNDS) CurrentRowNo() int { return d.pos }
func (d *errFirstAfterNDS) Columns() []struct{ Name string } {
	return []struct{ Name string }{{Name: "id"}, {Name: "parentId"}}
}

// errFirstDS always errors on First().
type errFirstDS struct{}

func (d *errFirstDS) RowCount() int                      { return 1 }
func (d *errFirstDS) First() error                       { return errors.New("First() always fails") }
func (d *errFirstDS) Next() error                        { return nil }
func (d *errFirstDS) EOF() bool                          { return true }
func (d *errFirstDS) GetValue(string) (any, error)       { return nil, nil }

// ── helper: build a minimal engine ───────────────────────────────────────────

func newGapsEngine(t *testing.T) *engine.ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("engine.Run: %v", err)
	}
	return e
}

// ── RunDataBandRows: startNewPageForCurrent (lines 62-64) ────────────────────
//
// Requires db.StartNewPage()==true, db.FlagUseStartNewPage==true, and rowIdx>0.
// PreparedPages must be non-nil (ensured by Run()).

func TestRunDataBandRows_StartNewPage(t *testing.T) {
	e := newGapsEngine(t)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetStartNewPage(true)
	db.FlagUseStartNewPage = true

	initialPages := e.PreparedPages().Count()
	e.RunDataBandRows(db, 3) // rowIdx 1 and 2 both trigger startNewPageForCurrent
	if e.PreparedPages().Count() <= initialPages {
		t.Errorf("StartNewPage in RunDataBandRows: expected more pages, got %d → %d",
			initialPages, e.PreparedPages().Count())
	}
}

// ── RunDataBandRows: runBands error return (lines 76-78) ─────────────────────
//
// A sub-DataBand added to db.Objects() errors when its data source's First()
// fails, causing runBands to return an error and RunDataBandRows to return early.

func TestRunDataBandRows_SubBandRunBandsError(t *testing.T) {
	e := newGapsEngine(t)

	// Sub-band: a DataBand whose data source errors on First().
	subDB := band.NewDataBand()
	subDB.SetHeight(5)
	subDB.SetVisible(true)
	subDB.SetDataSource(&errFirstDS{})

	// Parent DataBand: add subDB to its Objects collection.
	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.Objects().Add(report.Base(subDB))

	// RunDataBandRows should return early after runBands errors on row 0.
	// The test must not panic.
	e.RunDataBandRows(db, 2)
}

// ── RunDataBandFull: break in filter-skip after Next() error (lines 164-165) ─
//
// Use a DS that errors on Next() immediately. Set a filter that always evaluates
// to false so the skip path is taken. The DS's Next() errors, triggering break.

func TestRunDataBandFull_FilterSkipNextError(t *testing.T) {
	e := newGapsEngine(t)

	// A DS with 2 rows where Next() errors on the 1st call.
	rows := []map[string]any{
		{"Val": -1},
		{"Val": -2},
	}
	ds := newErrNextAfterNDS(rows, 1)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(ds)
	// Filter always false → filter-skip path taken; Next() errors → break.
	db.SetFilter("[Val] > 0")

	// Should not panic or hang; error from Next() causes break.
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── RunDataBandFull: restore+return err from sub-band runBands (lines 208-211) ─
//
// A DataBand whose Objects collection contains a sub-DataBand that errors.
// The error propagates through runBands → applyRelationFilters restore() →
// return err at lines 208-211.

func TestRunDataBandFull_SubBandRunBandsError(t *testing.T) {
	e := newGapsEngine(t)

	// Sub-band that errors on First().
	subDB := band.NewDataBand()
	subDB.SetHeight(5)
	subDB.SetVisible(true)
	subDB.SetDataSource(&errFirstDS{})

	// Parent DataBand with 2 data rows and the sub-band in its Objects.
	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newMockDS(2))
	db.Objects().Add(report.Base(subDB))

	err := e.RunDataBandFull(db)
	if err == nil {
		t.Fatal("expected error from sub-band runBands, got nil")
	}
}

// ── RunDataBandFull: aborted break inside row loop (lines 217-218) ───────────
//
// Pre-set e.aborted = true by calling e.Abort() before RunDataBandFull.
// The loop processes the first row, then ds.Next() succeeds, then
// "if e.aborted { break }" fires and exits the loop early.

func TestRunDataBandFull_AbortedBreakMidLoop(t *testing.T) {
	e := newGapsEngine(t)
	e.Abort() // pre-set aborted before the loop starts

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newMockDS(5)) // 5 rows, but loop should break after row 1

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// At least 1 row shown (the loop ran at least once before abort break).
	if e.CurY() <= beforeY {
		t.Error("expected at least 1 row to be shown before abort break")
	}
}

// ── runDataBandHierarchical: !hasFull early return (lines 274-276) ────────────
//
// Use bandOnlyDS which satisfies band.DataSource but not data.DataSource.
// The hasFull type assertion fails → function returns nil immediately.

func TestRunDataBandFull_Hierarchical_NonDataDS(t *testing.T) {
	e := newGapsEngine(t)

	ds := newBandOnlyDS(3)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetIDColumn("id")
	db.SetParentIDColumn("parentId")
	db.SetDataSource(ds)

	// Should not error; returns nil immediately because ds is not data.DataSource.
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("expected nil error for non-data.DataSource hierarchical, got: %v", err)
	}
}

// ── runDataBandHierarchical: break in snapshot loop on Next() error (lines 286-287) ─
//
// Provide a DS (implementing data.DataSource) where Next() errors on the 1st call.
// The snapshot loop iterates and calls ds.Next() to advance; on error it breaks.

func TestRunDataBandFull_Hierarchical_SnapshotNextError(t *testing.T) {
	e := newGapsEngine(t)

	rows := []map[string]any{
		{"id": "1", "parentId": "0"},
		{"id": "2", "parentId": "0"},
	}
	// Error on first Next() call (so snapshot gets row 0 then breaks).
	ds := newErrNextAfterNDS(rows, 1)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetIDColumn("id")
	db.SetParentIDColumn("parentId")
	db.SetDataSource(ds)

	// Should not error; snapshot loop breaks on Next() error.
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── runDataBandHierarchical: ds.First() error in renderRows (lines 325-327) ──
// ── + nested renderRows error propagation (lines 362-364, 370-372) ──────────
//
// Use a DS that errors on the 3rd First() call:
//   call 1: line 139 in RunDataBandFull (ds.First() before hierarchical call)
//   call 2: line 310 in runDataBandHierarchical (reset before rendering)
//   call 3: line 325 in renderRows (seek to root row) → error returned
//
// The error propagates through renderRows return → line 362/364 (if child exists)
// → line 370/372 (top-level renderRows).

func TestRunDataBandFull_Hierarchical_RenderRowsFirstError(t *testing.T) {
	e := newGapsEngine(t)

	rows := []map[string]any{
		{"id": "1", "parentId": ""},
		{"id": "2", "parentId": ""},
	}
	// Error on 3rd First() call (covers lines 325-327, 370-372).
	ds := newErrFirstAfterNDS(rows, 3)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetIDColumn("id")
	db.SetParentIDColumn("parentId")
	db.SetDataSource(ds)

	err := e.RunDataBandFull(db)
	if err == nil {
		t.Fatal("expected error from ds.First() in renderRows, got nil")
	}
}

// TestRunDataBandFull_Hierarchical_RenderRowsFirstErrorWithChildren exercises
// lines 362-364 (nested renderRows error propagation) by having a parent row
// with children. The error occurs in the child's renderRows call:
//   - 1st First(): line 139 (RunDataBandFull)
//   - 2nd First(): line 310 (runDataBandHierarchical reset)
//   - 3rd First(): line 325 in renderRows for root row 1 (succeeds if errOnN==4)
//   - 4th First(): line 325 in renderRows for child row → error
//   → propagates through line 362 (child renderRows return) → line 364 return err

func TestRunDataBandFull_Hierarchical_NestedRenderRowsError(t *testing.T) {
	e := newGapsEngine(t)

	rows := []map[string]any{
		{"id": "1", "parentId": ""},   // root (idx=0)
		{"id": "2", "parentId": "1"},  // child of root (idx=1)
	}
	// Error on 4th First() call (covers lines 362-364 via child renderRows).
	ds := newErrFirstAfterNDS(rows, 4)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetIDColumn("id")
	db.SetParentIDColumn("parentId")
	db.SetDataSource(ds)

	err := e.RunDataBandFull(db)
	if err == nil {
		t.Fatal("expected error from nested renderRows, got nil")
	}
}

// ── runDataBandHierarchical: break from Next() error in seek loop (lines 329-330) ─
//
// The seek loop in renderRows does: for k := 0; k < row.idx; k++ { ds.Next() }
// For a root at row.idx==1, the seek runs Next() once. If that Next() errors,
// lines 329-330 fire.
//
// Strategy: need a root row with idx==1. That means the first snapshot row is
// NOT a root (has a parent in the id set), and the second snapshot row IS a root.
// Next() errors on the seek call (after ds.First() in renderRows at step 3).

func TestRunDataBandFull_Hierarchical_SeekNextError(t *testing.T) {
	e := newGapsEngine(t)

	rows := []map[string]any{
		{"id": "1", "parentId": "2"},  // idx=0, NOT a root (parentId "2" exists)
		{"id": "2", "parentId": ""},   // idx=1, root
	}
	// Error on 2nd Next() call.
	// Call sequence:
	//   snapshot loop: Next() call 1 (advance from row 0 to row 1) — succeeds
	//   seek in renderRows for root at idx=1: Next() call 2 — errors → break
	ds := newErrNextAfterNDS(rows, 2)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetIDColumn("id")
	db.SetParentIDColumn("parentId")
	db.SetDataSource(ds)

	// Should not panic; seek Next() error causes break (not return).
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
