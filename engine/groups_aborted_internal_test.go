package engine

// groups_aborted_internal_test.go — internal tests (package engine) targeting
// the two uncovered branches in groups.go:
//
//   makeGroupTree line 130: `if e.aborted { break }` inside the row loop
//   showGroupTree line 202: `if e.aborted { break }` inside the items loop
//
// These branches require setting e.aborted = true, which is only accessible
// from within the package. External tests cannot reach these branches directly.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── helper: minimal string-backed DataSource ──────────────────────────────────

type groupsAbortStringDS struct {
	rows []string
	pos  int
}

func newGroupsAbortStringDS(rows ...string) *groupsAbortStringDS {
	return &groupsAbortStringDS{rows: rows, pos: -1}
}

func (d *groupsAbortStringDS) RowCount() int { return len(d.rows) }
func (d *groupsAbortStringDS) First() error  { d.pos = 0; return nil }
func (d *groupsAbortStringDS) Next() error {
	d.pos++
	return nil
}
func (d *groupsAbortStringDS) EOF() bool { return d.pos >= len(d.rows) }
func (d *groupsAbortStringDS) GetValue(column string) (any, error) {
	if d.pos < 0 || d.pos >= len(d.rows) {
		return nil, nil
	}
	return d.rows[d.pos], nil
}

func buildAbortedGroupEngine(t *testing.T) *ReportEngine {
	t.Helper()
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return e
}

// ── makeGroupTree: e.aborted break (groups.go:130) ───────────────────────────

func TestMakeGroupTree_AbortedPreset(t *testing.T) {
	e := buildAbortedGroupEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_Aborted")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val")

	db := band.NewDataBand()
	db.SetName("DB_Aborted")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newGroupsAbortStringDS("A", "B", "C", "D", "E"))
	gh.SetData(db)

	// Pre-set aborted so the row loop breaks immediately.
	e.aborted = true
	e.RunGroup(gh)

	if !e.aborted {
		t.Error("expected e.aborted to remain true after RunGroup with pre-set abort")
	}
}

// ── showGroupTree: e.aborted break (groups.go:202) ───────────────────────────

func TestShowGroupTree_AbortedPreset(t *testing.T) {
	e := buildAbortedGroupEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_ShowAbort")
	gh.SetVisible(true)
	gh.SetHeight(10)

	db := band.NewDataBand()
	db.SetName("DB_ShowAbort")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newGroupsAbortStringDS("A", "B", "C"))
	gh.SetData(db)

	child1 := &groupTreeItem{band: gh, rowNo: 0, rowCount: 1}
	child2 := &groupTreeItem{band: gh, rowNo: 1, rowCount: 1}
	root := &groupTreeItem{}
	root.addItem(child1)
	root.addItem(child2)

	e.aborted = true
	e.showGroupTree(root)

	if !e.aborted {
		t.Error("expected e.aborted to remain true after aborted showGroupTree")
	}
}
