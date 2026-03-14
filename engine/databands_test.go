package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── RunDataBandRows ───────────────────────────────────────────────────────────

func TestRunDataBandRows_Empty(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(20)
	db.SetVisible(true)

	beforeY := e.CurY()
	e.RunDataBandRows(db, 0)
	if e.CurY() != beforeY {
		t.Errorf("empty rows: CurY changed from %v to %v", beforeY, e.CurY())
	}
}

func TestRunDataBandRows_SingleRow(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(25)
	db.SetVisible(true)

	beforeY := e.CurY()
	e.RunDataBandRows(db, 1)

	if db.RowNo() != 1 {
		t.Errorf("RowNo = %d, want 1", db.RowNo())
	}
	if db.IsFirstRow() != true {
		t.Error("IsFirstRow should be true for single row")
	}
	if db.IsLastRow() != true {
		t.Error("IsLastRow should be true for single row")
	}
	if e.CurY() != beforeY+25 {
		t.Errorf("CurY = %v, want %v", e.CurY(), beforeY+25)
	}
}

func TestRunDataBandRows_MultipleRows(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)

	beforeY := e.CurY()
	e.RunDataBandRows(db, 5)

	if db.RowNo() != 5 {
		t.Errorf("RowNo = %d, want 5 (last row)", db.RowNo())
	}
	if e.CurY() != beforeY+50 {
		t.Errorf("CurY = %v, want %v (5×10)", e.CurY(), beforeY+50)
	}
}

func TestRunDataBandRows_WithHeader(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	hdr := band.NewDataHeaderBand()
	hdr.SetHeight(15)
	hdr.SetVisible(true)

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetHeader(hdr)

	beforeY := e.CurY()
	e.RunDataBandRows(db, 2)
	// header (15) + 2×row (10) = 35
	if e.CurY() != beforeY+35 {
		t.Errorf("CurY = %v, want %v (15+20)", e.CurY(), beforeY+35)
	}
}

func TestRunDataBandRows_WithHeaderAndFooter(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	hdr := band.NewDataHeaderBand()
	hdr.SetHeight(10)
	hdr.SetVisible(true)

	ftr := band.NewDataFooterBand()
	ftr.SetHeight(10)
	ftr.SetVisible(true)

	db := band.NewDataBand()
	db.SetHeight(5)
	db.SetVisible(true)
	db.SetHeader(hdr)
	db.SetFooter(ftr)

	beforeY := e.CurY()
	e.RunDataBandRows(db, 3)
	// header (10) + 3×row (5) + footer (10) = 35
	if e.CurY() != beforeY+35 {
		t.Errorf("CurY = %v, want %v", e.CurY(), beforeY+35)
	}
}

func TestRunDataBandRows_IsFirstLastRow(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(5)
	db.SetVisible(true)

	e.RunDataBandRows(db, 3)

	// After RunDataBandRows the band reflects the last row state.
	if db.IsLastRow() != true {
		t.Error("last row: IsLastRow should be true")
	}
}

// ── RunDataBandFull ───────────────────────────────────────────────────────────

// mockDS is a simple in-memory data source for testing.
type mockDS struct {
	rows    []map[string]any
	current int
}

func newMockDS(n int) *mockDS {
	rows := make([]map[string]any, n)
	for i := range rows {
		rows[i] = map[string]any{"id": i + 1}
	}
	return &mockDS{rows: rows, current: -1}
}

func (m *mockDS) RowCount() int { return len(m.rows) }
func (m *mockDS) First() error  { m.current = 0; return nil }
func (m *mockDS) Next() error {
	m.current++
	if m.current >= len(m.rows) {
		return nil
	}
	return nil
}
func (m *mockDS) EOF() bool { return m.current >= len(m.rows) }
func (m *mockDS) GetValue(col string) (any, error) {
	if m.current < 0 || m.current >= len(m.rows) {
		return nil, nil
	}
	return m.rows[m.current][col], nil
}

func TestRunDataBandFull_WithDataSource(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newMockDS(4))

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull: %v", err)
	}

	// 4 rows × 10px = 40
	if e.CurY() != beforeY+40 {
		t.Errorf("CurY = %v, want %v", e.CurY(), beforeY+40)
	}
	if db.RowNo() != 4 {
		t.Errorf("RowNo = %d, want 4", db.RowNo())
	}
}

func TestRunDataBandFull_MaxRows(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newMockDS(10))
	db.SetMaxRows(3)

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull: %v", err)
	}

	// MaxRows=3 → 3 rows × 10 = 30
	if e.CurY() != beforeY+30 {
		t.Errorf("CurY = %v, want %v (MaxRows=3)", e.CurY(), beforeY+30)
	}
}

func TestRunDataBandFull_NoDataSource(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(20)
	db.SetVisible(true)
	// No data source.

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull: %v", err)
	}
	// No DataSource: band renders once as a static band, CurY advances by height.
	if e.CurY() != beforeY+20 {
		t.Errorf("no datasource: CurY = %v, want %v (band rendered once)", e.CurY(), beforeY+20)
	}
}

func TestRunDataBandFull_EmptyDataSource(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := engine.New(r)
	_ = e.Run(engine.DefaultRunOptions())

	db := band.NewDataBand()
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetDataSource(newMockDS(0)) // empty
	db.SetPrintIfDSEmpty(true)

	beforeY := e.CurY()
	if err := e.RunDataBandFull(db); err != nil {
		t.Fatalf("RunDataBandFull: %v", err)
	}
	// PrintIfDSEmpty: render 1 empty row.
	if e.CurY() != beforeY+10 {
		t.Errorf("PrintIfDSEmpty: CurY = %v, want %v", e.CurY(), beforeY+10)
	}
}
