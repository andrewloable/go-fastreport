package engine_test

// engine_databands_extra_test.go — targeted coverage for remaining uncovered
// branches in:
//   databands.go: RunDataBandFull (alias resolution, aborted break)
//   databands.go: runDataBandHierarchical (ds.First() error on second call)
//   groups.go:    makeGroupTree (aborted break), showGroupTree (empty DS)

import (
	"errors"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── test helpers ──────────────────────────────────────────────────────────────

func newExtraEngine(t *testing.T) *engine.ReportEngine {
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

// dataBandFromXML deserializes a *band.DataBand from an XML snippet.
// This is the only way from an external package to set dataSourceAlias, which
// is populated from the FRX "DataSource" attribute during Deserialize.
func dataBandFromXML(t *testing.T, xmlStr string) *band.DataBand {
	t.Helper()
	sr := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := sr.ReadObjectHeader()
	if !ok {
		t.Fatal("dataBandFromXML: ReadObjectHeader returned false")
	}
	db := band.NewDataBand()
	if err := db.Deserialize(sr); err != nil {
		t.Fatalf("dataBandFromXML: Deserialize: %v", err)
	}
	return db
}

// errOnSecondFirstDS implements band.DataSource and data.DataSource.
// It succeeds on the first First() call and fails on the second, used to
// exercise the ds.First() error path inside runDataBandHierarchical.
type errOnSecondFirstDS struct {
	rows      []map[string]any
	pos       int
	firstCall int
}

func newErrOnSecondFirstDS(rows []map[string]any) *errOnSecondFirstDS {
	return &errOnSecondFirstDS{rows: rows, pos: -1}
}

func (d *errOnSecondFirstDS) RowCount() int { return len(d.rows) }
func (d *errOnSecondFirstDS) First() error {
	d.firstCall++
	if d.firstCall > 1 {
		return errors.New("ds.First() intentional failure on second call")
	}
	d.pos = 0
	return nil
}
func (d *errOnSecondFirstDS) Next() error { d.pos++; return nil }
func (d *errOnSecondFirstDS) EOF() bool   { return d.pos >= len(d.rows) }
func (d *errOnSecondFirstDS) GetValue(col string) (any, error) {
	if d.pos < 0 || d.pos >= len(d.rows) {
		return nil, nil
	}
	return d.rows[d.pos][col], nil
}

// data.DataSource methods (in addition to band.DataSource above).
func (d *errOnSecondFirstDS) Name() string      { return "ErrSecondFirst" }
func (d *errOnSecondFirstDS) Alias() string     { return "ErrSecondFirst" }
func (d *errOnSecondFirstDS) Init() error       { d.pos = -1; d.firstCall = 0; return nil }
func (d *errOnSecondFirstDS) Close() error      { return nil }
func (d *errOnSecondFirstDS) CurrentRowNo() int { return d.pos }
func (d *errOnSecondFirstDS) Columns() []data.Column {
	return []data.Column{{Name: "id"}, {Name: "parentId"}}
}

// ── RunDataBandFull: alias resolution path (lines 108-113 in databands.go) ───
//
// When db.DataSourceRef() is nil but db.DataSourceAlias() is non-empty and
// the engine's dictionary resolves it to a band.DataSource, the engine binds
// it and uses it.
//
// dataSourceAlias is only set via Deserialize (from the FRX "DataSource"
// attribute), so we create the band through XML deserialization.

func TestRunDataBandFull_AliasResolution(t *testing.T) {
	r := reportpkg.NewReport()

	ds := data.NewBaseDataSource("AliasDS")
	ds.SetAlias("AliasDS")
	ds.AddColumn(data.Column{Name: "Val"})
	ds.AddRow(map[string]any{"Val": 1})
	ds.AddRow(map[string]any{"Val": 2})
	if err := ds.Init(); err != nil {
		t.Fatalf("ds.Init: %v", err)
	}
	r.Dictionary().AddDataSource(ds)

	pg := reportpkg.NewReportPage()
	r.AddPage(pg)

	e := engine.New(r)
	e.RegisterDataSource(ds)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Band with alias "AliasDS" set via the "DataSource" XML attribute.
	// DataSourceRef() returns nil because no SetDataSource was called.
	db := dataBandFromXML(t, `<Data Name="AliasBand" Height="10" Visible="true" DataSource="AliasDS"/>`)

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull with alias: %v", err)
	}
	// 2 rows × 10px = 20.
	if e.CurY() != beforeY+20 {
		t.Errorf("alias resolution: CurY = %v, want %v", e.CurY(), beforeY+20)
	}
}

// ── RunDataBandFull: aborted break inside row loop (lines 217-218) ────────────
//
// e.aborted is set to true by e.Abort(). The loop checks it after ds.Next()
// and breaks. We call Abort() from a state handler after the first
// EngineStateBlockFinished event.

func TestRunDataBandFull_AbortedBreak(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	ds := data.NewBaseDataSource("AbortDS")
	ds.SetAlias("AbortDS")
	ds.AddColumn(data.Column{Name: "Val"})
	for i := 0; i < 5; i++ {
		ds.AddRow(map[string]any{"Val": i})
	}
	if err := ds.Init(); err != nil {
		t.Fatalf("ds.Init: %v", err)
	}

	db := band.NewDataBand()
	db.SetName("AbortBand")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(ds)

	abortFired := false
	e.AddStateHandler(func(_ any, state engine.EngineState) {
		if state == engine.EngineStateBlockFinished && !abortFired {
			abortFired = true
			e.Abort()
		}
	})

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull aborted: %v", err)
	}
	// At least 1 row must be shown.
	if e.CurY() <= beforeY {
		t.Error("aborted: expected at least 1 row shown")
	}
}

// ── runDataBandHierarchical: ds.First() error on second call (line 310-311) ──

func TestRunDataBandFull_Hierarchical_SecondFirstError(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	rows := []map[string]any{
		{"id": "1", "parentId": "0"},
		{"id": "2", "parentId": "0"},
	}
	ds := newErrOnSecondFirstDS(rows)

	db := band.NewDataBand()
	db.SetName("HierErr")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetIDColumn("id")
	db.SetParentIDColumn("parentId")
	db.SetDataSource(ds)

	err := e.RunDataBandFull(db)
	if err == nil {
		t.Fatal("expected error from ds.First() on second call, got nil")
	}
}

// ── makeGroupTree: aborted break (line 130 in groups.go) ─────────────────────

func TestMakeGroupTree_AbortedBreak(t *testing.T) {
	ds := data.NewBaseDataSource("GroupAbortDS")
	ds.SetAlias("GroupAbortDS")
	ds.AddColumn(data.Column{Name: "Group"})
	for i := 0; i < 5; i++ {
		ds.AddRow(map[string]any{"Group": "A"})
	}
	if err := ds.Init(); err != nil {
		t.Fatalf("ds.Init: %v", err)
	}

	r := reportpkg.NewReport()

	db := band.NewDataBand()
	db.SetName("GroupData")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(ds)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_Abort")
	gh.SetHeight(12)
	gh.SetVisible(true)
	gh.SetCondition("Group")
	gh.SetData(db)

	pg := reportpkg.NewReportPage()
	pg.AddBand(gh)
	r.AddPage(pg)

	e := engine.New(r)

	aborted := false
	e.AddStateHandler(func(_ any, state engine.EngineState) {
		if state == engine.EngineStateGroupFinished && !aborted {
			aborted = true
			e.Abort()
		}
	})

	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with group abort: %v", err)
	}
	if !e.Aborted() {
		t.Error("engine should report Aborted() = true after group abort")
	}
}

// ── showGroupTree: empty data source ─────────────────────────────────────────
//
// When the data source has no rows, makeGroupTree produces a root with no child
// items, and showGroupTree encounters an empty root (len(root.items) == 0,
// root.band == nil), which is the synthetic-root leaf path.

func TestShowGroupTree_EmptyDataSource(t *testing.T) {
	ds := data.NewBaseDataSource("EmptyGroupDS")
	ds.SetAlias("EmptyGroupDS")
	ds.AddColumn(data.Column{Name: "Group"})
	if err := ds.Init(); err != nil {
		t.Fatalf("ds.Init: %v", err)
	}

	r := reportpkg.NewReport()

	db := band.NewDataBand()
	db.SetName("EmptyGroupData")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(ds)

	gh := band.NewGroupHeaderBand()
	gh.SetName("GH_Empty")
	gh.SetHeight(12)
	gh.SetVisible(true)
	gh.SetCondition("Group")
	gh.SetData(db)

	pg := reportpkg.NewReportPage()
	pg.AddBand(gh)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run with empty group data source: %v", err)
	}
}

// ── RunDataBandFull: report.SetCalcContext (data.DataSource path) ────────────
//
// Exercises the `if e.report != nil { if fullDS, ok := ds.(data.DataSource); ok { ... } }`
// path inside the row loop, using a BaseDataSource which satisfies data.DataSource.

func TestRunDataBandFull_ReportCalcContext(t *testing.T) {
	e := newExtraEngine(t)

	ds := data.NewBaseDataSource("CalcCtxDS")
	ds.SetAlias("CalcCtxDS")
	ds.AddColumn(data.Column{Name: "Val"})
	ds.AddRow(map[string]any{"Val": 42})
	if err := ds.Init(); err != nil {
		t.Fatalf("ds.Init: %v", err)
	}

	db := band.NewDataBand()
	db.SetName("CalcCtxBand")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(ds)

	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull CalcContext: %v", err)
	}
}
