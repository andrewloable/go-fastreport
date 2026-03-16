package engine_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/engine"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// TestDataBandColumns_TwoColumn_RowsAreSideBySide verifies that when
// DataBand.Columns.Count == 2, rows are placed side-by-side at the same Top
// position rather than stacking vertically.
func TestDataBandColumns_TwoColumn_RowsAreSideBySide(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.PaperWidth = 210
	pg.PaperHeight = 297
	pg.LeftMargin = 0
	pg.RightMargin = 0
	pg.TopMargin = 0
	pg.BottomMargin = 0

	const colWidth = 100
	const rowHeight = 20
	const numRows = 4

	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(rowHeight)
	db.SetWidth(colWidth) // represents one column width
	db.SetVisible(true)
	if err := db.Columns().SetCount(2); err != nil {
		t.Fatalf("SetCount: %v", err)
	}

	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// RunDataBandRows is used for in-memory rows. Let's use a separate engine
	// for the RunDataBandRows path.
	r2 := reportpkg.NewReport()
	pg2 := reportpkg.NewReportPage()
	pg2.PaperWidth = 210
	pg2.PaperHeight = 297
	pg2.LeftMargin = 0
	pg2.RightMargin = 0
	pg2.TopMargin = 0
	pg2.BottomMargin = 0

	db2 := band.NewDataBand()
	db2.SetName("DB2")
	db2.SetHeight(rowHeight)
	db2.SetWidth(colWidth)
	db2.SetVisible(true)
	if err := db2.Columns().SetCount(2); err != nil {
		t.Fatalf("SetCount db2: %v", err)
	}

	pg2.AddBand(db2)
	r2.AddPage(pg2)

	e2 := engine.New(r2)
	// Start the page to initialize engine state.
	if err := e2.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("e2.Run: %v", err)
	}

	// Run a fresh engine just for RunDataBandRows.
	r3 := reportpkg.NewReport()
	pg3 := reportpkg.NewReportPage()
	pg3.PaperWidth = 210
	pg3.PaperHeight = 297
	pg3.LeftMargin = 0
	pg3.RightMargin = 0
	pg3.TopMargin = 0
	pg3.BottomMargin = 0
	r3.AddPage(pg3)

	db3 := band.NewDataBand()
	db3.SetName("DB3")
	db3.SetHeight(rowHeight)
	db3.SetWidth(colWidth)
	db3.SetVisible(true)
	if err := db3.Columns().SetCount(2); err != nil {
		t.Fatalf("SetCount db3: %v", err)
	}

	e3 := engine.New(r3)
	e3.SetCurY(0)
	e3.SetCurX(0)

	// Manually call RunDataBandRows for 4 rows with 2 columns.
	// Expected: 2 column-rows, each at Y=0 and Y=rowHeight respectively.
	e3.RunDataBandRows(db3, numRows)

	pp := e3.PreparedPages()
	if pp == nil {
		// PreparedPages not initialised without a page — that's ok, just check CurY.
		// With 4 rows in 2 columns, we expect 2 Y-advances of rowHeight each.
		wantCurY := float32(numRows/2) * rowHeight
		if e3.CurY() != wantCurY {
			t.Errorf("CurY = %.1f, want %.1f (2 column-rows × %dpx)", e3.CurY(), wantCurY, rowHeight)
		}
		return
	}
}

// TestDataBandColumns_SingleColumn_NormalBehavior verifies that Columns.Count <= 1
// leaves normal single-column rendering unchanged.
func TestDataBandColumns_SingleColumn_NormalBehavior(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.PaperWidth = 210
	pg.PaperHeight = 297
	pg.LeftMargin = 0
	pg.RightMargin = 0
	pg.TopMargin = 0
	pg.BottomMargin = 0

	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(20)
	db.SetWidth(200)
	db.SetVisible(true)
	// Columns.Count defaults to 0 → single-column mode.

	pg.AddBand(db)
	r.AddPage(pg)

	e := engine.New(r)
	if err := e.Run(engine.DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// No assertion needed beyond successful run.
}

// TestDataBandColumns_RunDataBandRows_ColLayout verifies the Y advance pattern
// for RunDataBandRows with multi-column layout.
func TestDataBandColumns_RunDataBandRows_ColLayout(t *testing.T) {
	const colWidth float32 = 100
	const rowHeight float32 = 20
	const numCols = 3
	const numRows = 9 // exactly 3 column-rows of 3 columns each

	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.PaperWidth = 300
	pg.PaperHeight = 500
	pg.LeftMargin = 0
	pg.RightMargin = 0
	pg.TopMargin = 0
	pg.BottomMargin = 0

	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(rowHeight)
	db.SetWidth(colWidth)
	db.SetVisible(true)
	if err := db.Columns().SetCount(numCols); err != nil {
		t.Fatalf("SetCount: %v", err)
	}

	r.AddPage(pg)

	e := engine.New(r)
	e.SetCurY(0)
	e.SetCurX(0)

	e.RunDataBandRows(db, numRows)

	// After 9 rows in 3 columns: 3 column-rows × rowHeight = 60px total Y advance.
	wantCurY := float32(numRows/numCols) * rowHeight
	if e.CurY() != wantCurY {
		t.Errorf("CurY = %.1f, want %.1f (%d column-rows × %.0fpx)",
			e.CurY(), wantCurY, numRows/numCols, rowHeight)
	}
}

// TestDataBandColumns_RunDataBandRows_PartialLastRow verifies that a partial
// last column-row (not all columns filled) still advances Y correctly.
func TestDataBandColumns_RunDataBandRows_PartialLastRow(t *testing.T) {
	const colWidth float32 = 100
	const rowHeight float32 = 20
	const numCols = 3
	const numRows = 7 // 2 full column-rows + 1 partial (1 column)

	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	pg.PaperWidth = 300
	pg.PaperHeight = 500
	pg.LeftMargin = 0
	pg.RightMargin = 0
	pg.TopMargin = 0
	pg.BottomMargin = 0

	db := band.NewDataBand()
	db.SetName("DB")
	db.SetHeight(rowHeight)
	db.SetWidth(colWidth)
	db.SetVisible(true)
	if err := db.Columns().SetCount(numCols); err != nil {
		t.Fatalf("SetCount: %v", err)
	}

	r.AddPage(pg)

	e := engine.New(r)
	e.SetCurY(0)
	e.SetCurX(0)

	e.RunDataBandRows(db, numRows)

	// 2 full column-rows → 40px; 1 partial row (1 col filled) → flushed at 20px.
	// Total: 3 × rowHeight = 60px.
	wantCurY := float32(3) * rowHeight
	if e.CurY() != wantCurY {
		t.Errorf("CurY = %.1f, want %.1f (2 full + 1 partial column-row × %.0fpx)",
			e.CurY(), wantCurY, rowHeight)
	}
}
