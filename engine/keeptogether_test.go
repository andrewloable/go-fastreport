package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ktDS is a simple in-memory DataSource for KeepTogether tests.
type ktDS struct {
	rows []string
	pos  int
}

func newKtDS(rows ...string) *ktDS { return &ktDS{rows: rows, pos: -1} }

func (d *ktDS) RowCount() int { return len(d.rows) }
func (d *ktDS) First() error  { d.pos = 0; return nil }
func (d *ktDS) Next() error   { d.pos++; return nil }
func (d *ktDS) EOF() bool     { return d.pos >= len(d.rows) }
func (d *ktDS) GetValue(_ string) (any, error) {
	if d.pos >= 0 && d.pos < len(d.rows) {
		return d.rows[d.pos], nil
	}
	return nil, nil
}

// TestKeepTogether_GeneratesMultiplePages verifies that a DataBand with
// KeepTogether=true produces multiple pages when rows don't fit on one page.
// This is the regression test for go-fastreport-i96z.
func TestKeepTogether_GeneratesMultiplePages(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	// Small page: 53mm usable ≈ 200px at 3.78px/mm
	pg.PaperWidth = 210
	pg.PaperHeight = 73 // 73-10-10 = 53mm ≈ 200px usable
	pg.TopMargin = 10
	pg.BottomMargin = 10

	db := band.NewDataBand()
	db.SetName("DB")
	db.SetVisible(true)
	db.SetHeight(60) // 60px per row → 6 rows = 360px > 200px usable
	db.SetKeepTogether(true)
	db.SetDataSource(newKtDS("A", "B", "C", "D", "E", "F"))
	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	total := e.PreparedPages().Count()
	if total < 2 {
		t.Errorf("KeepTogether with 6×60px rows on 200px page: expected ≥2 pages, got %d", total)
	}
}
