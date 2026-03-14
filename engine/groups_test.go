package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// simpleDS is a DataSource backed by a slice of strings (one column "val").
type simpleDS struct {
	rows []string
	pos  int
}

func newSimpleDS(rows ...string) *simpleDS { return &simpleDS{rows: rows, pos: -1} }

func (d *simpleDS) RowCount() int { return len(d.rows) }
func (d *simpleDS) First() error  { d.pos = 0; return nil }
func (d *simpleDS) Next() error {
	d.pos++
	return nil
}
func (d *simpleDS) EOF() bool { return d.pos >= len(d.rows) }
func (d *simpleDS) GetValue(column string) (any, error) {
	if d.pos < 0 || d.pos >= len(d.rows) {
		return nil, nil
	}
	return d.rows[d.pos], nil
}

// buildGroupEngine returns a fresh engine with one page, ready to run.
func buildGroupEngine(t *testing.T) *engine.ReportEngine {
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

// ── RunGroup tests ────────────────────────────────────────────────────────────

func TestRunGroup_NilDataBand(t *testing.T) {
	e := buildGroupEngine(t)
	gh := band.NewGroupHeaderBand()
	// No Data band attached — RunGroup should be a no-op.
	e.RunGroup(gh)
}

func TestRunGroup_NilDataSource(t *testing.T) {
	e := buildGroupEngine(t)
	gh := band.NewGroupHeaderBand()
	db := band.NewDataBand()
	// No DataSource attached — RunGroup should be a no-op.
	gh.SetData(db)
	e.RunGroup(gh)
}

func TestRunGroup_EmptyDataSource(t *testing.T) {
	e := buildGroupEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GroupHeader")
	gh.SetVisible(true)
	gh.SetHeight(20)

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(20)
	db.SetDataSource(newSimpleDS()) // 0 rows
	gh.SetData(db)

	initialBands := e.PreparedPages().GetPage(0).Bands
	e.RunGroup(gh)

	// With 0 rows, nothing should be added.
	if len(e.PreparedPages().GetPage(0).Bands) != len(initialBands) {
		t.Error("RunGroup with empty data source should not add any bands")
	}
}

func TestRunGroup_SingleGroup(t *testing.T) {
	e := buildGroupEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GroupHeader")
	gh.SetVisible(true)
	gh.SetHeight(20)
	gh.SetCondition("val")

	gf := band.NewGroupFooterBand()
	gf.SetName("GroupFooter")
	gf.SetVisible(true)
	gf.SetHeight(10)
	gh.SetGroupFooter(gf)

	db := band.NewDataBand()
	db.SetName("DataBand")
	db.SetVisible(true)
	db.SetHeight(15)
	db.SetDataSource(newSimpleDS("A", "A", "B")) // 3 rows, 2 groups
	gh.SetData(db)

	initialCount := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)

	newCount := len(e.PreparedPages().GetPage(0).Bands)
	if newCount <= initialCount {
		t.Error("RunGroup should have added bands for groups")
	}
}

func TestRunGroup_WithFooter(t *testing.T) {
	e := buildGroupEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH")
	gh.SetVisible(true)
	gh.SetHeight(20)

	gf := band.NewGroupFooterBand()
	gf.SetName("GF")
	gf.SetVisible(true)
	gf.SetHeight(10)
	gh.SetGroupFooter(gf)

	db := band.NewDataBand()
	db.SetName("DB")
	db.SetVisible(true)
	db.SetHeight(15)
	db.SetDataSource(newSimpleDS("X")) // 1 row, 1 group
	gh.SetData(db)

	initialCount := len(e.PreparedPages().GetPage(0).Bands)
	e.RunGroup(gh)

	added := len(e.PreparedPages().GetPage(0).Bands) - initialCount
	// Expect: GroupHeader(1) + DataBand row(1) + GroupFooter(1) = 3 minimum
	if added < 3 {
		t.Errorf("expected at least 3 bands added (header+data+footer), got %d", added)
	}
}

func TestRunGroup_RowCounters(t *testing.T) {
	e := buildGroupEngine(t)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH")
	gh.SetVisible(true)
	gh.SetHeight(10)
	gh.SetCondition("val") // match the column name in simpleDS

	db := band.NewDataBand()
	db.SetName("DB")
	db.SetVisible(true)
	db.SetHeight(10)
	db.SetDataSource(newSimpleDS("A", "B", "C")) // 3 different groups
	gh.SetData(db)

	e.RunGroup(gh)

	// After running 3 groups (A, B, C all different), RowNo on GroupHeader should be 3.
	if gh.RowNo() != 3 {
		t.Errorf("GroupHeader.RowNo = %d, want 3", gh.RowNo())
	}
}
