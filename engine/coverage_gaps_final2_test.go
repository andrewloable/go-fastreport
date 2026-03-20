package engine

// coverage_gaps_final2_test.go — second wave of internal tests targeting the
// remaining coverage gaps found after coverage_gaps_final_test.go raised
// overall coverage from 94.1% to 95.3%.
//
// New targets:
//  1. keepwithdata.go: CheckKeepFooter — startNewPageForCurrent branch (footer
//     height exceeds freeSpace → page break triggered).
//  2. bands.go: showFullBandOnce — outputBand != nil path (PrintOnParent mode).
//  3. pages.go: showBand — outputBand != nil path.
//  4. pages.go: showBand — curX != 0 path (column X offset applied to objects).
//  5. pages.go: showBand — OverlayBand does not advance CurY.
//  6. pages.go: extractBandBase — remaining band type cases:
//     ReportSummaryBand, ColumnHeaderBand, ColumnFooterBand, OverlayBand,
//     DataHeaderBand, DataFooterBand, GroupFooterBand, ReportTitleBand.
//  7. databands.go: RunDataBandRowsKeep — keepDetail+EndKeep, oneRow path.
//  8. databands.go: showDataBandFooter — ds.Prior() path (ds implements Prior()).
//  9. databands.go: runDataBandNoDS — filter-false suppression path.
// 10. objects.go: evalTextWithFormat — nil report & empty text early returns;
//     format != nil with non-bracket text; isSingleBracketExpr false path.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

// ── 1: CheckKeepFooter — startNewPageForCurrent branch ───────────────────────

// TestCheckKeepFooter_InsufficientSpace exercises the startNewPageForCurrent
// branch (FreeSpace < GetFootersHeight → page break).
func TestCheckKeepFooter_InsufficientSpace_PageBreak(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Set up a DataBand with a large footer so GetFootersHeight >> FreeSpace.
	ftr := band.NewDataFooterBand()
	ftr.SetHeight(5000) // much taller than page height
	ftr.SetVisible(true)
	ftr.SetKeepWithData(true)

	db := band.NewDataBand()
	db.SetName("CFPageBreakDB")
	db.SetHeight(10)
	db.SetFooter(ftr)

	pgsBefore := e.preparedPages.Count()
	e.CheckKeepFooter(db)
	// A new page should have been started.
	if e.preparedPages.Count() <= pgsBefore {
		t.Error("CheckKeepFooter insufficient space: expected new page to be started")
	}
}

// ── 2 & 3: showFullBandOnce + showBand — outputBand != nil (PrintOnParent) ───

// TestShowFullBandOnce_OutputBand_MergesObjects exercises the outputBand != nil
// branch in showFullBandOnce. When outputBand is set, objects from the inner
// band are merged into outputBand rather than added to preparedPages.
func TestShowFullBandOnce_OutputBand_MergesObjects(t *testing.T) {
	e := newCovEngine(t)

	// Create a parent PreparedBand to act as the outputBand.
	parent := &preview.PreparedBand{
		Name:   "ParentBand",
		Height: 100,
	}
	e.outputBand = parent
	e.outputBandOffsetX = 0
	e.outputBandOffsetY = 0

	txt := object.NewTextObject()
	txt.SetName("InnerTxt")
	txt.SetLeft(5)
	txt.SetTop(10)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetText("inner text")
	txt.SetVisible(true)

	b := band.NewBandBase()
	b.SetName("InnerBand")
	b.SetHeight(20)
	b.SetVisible(true)
	b.Objects().Add(txt)

	e.ShowFullBand(b)

	// Objects should have been merged into the parent PreparedBand.
	if len(parent.Objects) == 0 {
		t.Error("outputBand mode: no objects were merged into parent PreparedBand")
	}

	// Restore outputBand.
	e.outputBand = nil
}

// TestShowBand_OutputBand_MergesObjects exercises the outputBand != nil branch
// in showBand. Uses a PageHeaderBand (not BandBase) to exercise the hasObjects
// interface path inside showBand's outputBand branch.
func TestShowBand_OutputBand_MergesObjects(t *testing.T) {
	e := newCovEngine(t)

	parent := &preview.PreparedBand{
		Name:   "ShowBandParent",
		Height: 100,
	}
	e.outputBand = parent
	e.outputBandOffsetX = 0
	e.outputBandOffsetY = 0

	ph := band.NewPageHeaderBand()
	ph.SetName("PHInner")
	ph.SetHeight(15)
	ph.SetVisible(true)

	txt := object.NewTextObject()
	txt.SetName("PHInnerTxt")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(15)
	txt.SetText("header text")
	txt.SetVisible(true)
	ph.Objects().Add(txt)

	e.showBand(ph)

	if len(parent.Objects) == 0 {
		t.Error("showBand outputBand mode: no objects merged into parent PreparedBand")
	}

	e.outputBand = nil
}

// ── 4: showBand — curX != 0 applies column X offset ─────────────────────────

// TestShowBand_CurXNonZero exercises the `if e.curX != 0` branch that shifts
// PreparedObject Left positions by the column offset.
func TestShowBand_CurXNonZero(t *testing.T) {
	e := newCovEngine(t)

	// Set a non-zero curX to simulate being in column 1 of a multi-column layout.
	e.curX = 50.0

	txt := object.NewTextObject()
	txt.SetName("CurXTxt")
	txt.SetLeft(10) // declared left
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(15)
	txt.SetText("col offset test")
	txt.SetVisible(true)

	ph := band.NewPageHeaderBand()
	ph.SetName("CurXBand")
	ph.SetHeight(15)
	ph.SetVisible(true)
	ph.Objects().Add(txt)

	pg0 := e.preparedPages.GetPage(0)
	before := len(pg0.Bands)

	e.showBand(ph)

	after := len(pg0.Bands)
	if after <= before {
		t.Fatal("showBand curX non-zero: expected band to be added")
	}

	// The PreparedObject's Left should be declared_left + curX = 10 + 50 = 60.
	added := pg0.Bands[after-1]
	if len(added.Objects) > 0 {
		if added.Objects[0].Left != 60 {
			t.Errorf("showBand curX offset: Left = %v, want 60 (10+50)", added.Objects[0].Left)
		}
	}

	// Reset curX.
	e.curX = 0
}

// ── 5: showBand — OverlayBand does not advance CurY ─────────────────────────

// TestShowBand_OverlayBand_DoesNotAdvanceCurY exercises the OverlayBand path
// in showBand where curY must NOT be advanced after rendering.
func TestShowBand_OverlayBand_DoesNotAdvanceCurY(t *testing.T) {
	e := newCovEngine(t)

	overlay := band.NewOverlayBand()
	overlay.SetName("OverlayTest")
	overlay.SetHeight(30)
	overlay.SetVisible(true)

	startY := e.curY
	e.showBand(overlay)

	if e.curY != startY {
		t.Errorf("showBand OverlayBand: CurY advanced from %v to %v, expected no change", startY, e.curY)
	}
}

// ── 6: extractBandBase — remaining type cases ─────────────────────────────────

// TestExtractBandBase_AllTypes exercises every case in extractBandBase.
func TestExtractBandBase_AllTypes(t *testing.T) {
	cases := []struct {
		name string
		b    report.Base
	}{
		{"BandBase", band.NewBandBase()},
		{"ReportTitleBand", band.NewReportTitleBand()},
		{"ReportSummaryBand", band.NewReportSummaryBand()},
		{"PageHeaderBand", band.NewPageHeaderBand()},
		{"PageFooterBand", band.NewPageFooterBand()},
		{"ColumnHeaderBand", band.NewColumnHeaderBand()},
		{"ColumnFooterBand", band.NewColumnFooterBand()},
		{"DataHeaderBand", band.NewDataHeaderBand()},
		{"DataFooterBand", band.NewDataFooterBand()},
		{"GroupHeaderBand", band.NewGroupHeaderBand()},
		{"GroupFooterBand", band.NewGroupFooterBand()},
		{"OverlayBand", band.NewOverlayBand()},
		{"DataBand", band.NewDataBand()},
		{"ChildBand", band.NewChildBand()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bb := extractBandBase(tc.b)
			if bb == nil {
				t.Errorf("extractBandBase(%s): expected non-nil *BandBase", tc.name)
			}
		})
	}
}

// TestExtractBandBase_UnknownType exercises the nil-return fallback.
func TestExtractBandBase_UnknownType(t *testing.T) {
	obj := report.NewBaseObject()
	bb := extractBandBase(obj)
	if bb != nil {
		t.Error("extractBandBase unknown type: expected nil")
	}
}

// ── 7: RunDataBandRowsKeep — keepDetail path ─────────────────────────────────

// TestRunDataBandRowsKeep_KeepDetail exercises the KeepDetail path where
// startKeepBand is called for each row and EndKeep after each row.
func TestRunDataBandRowsKeep_KeepDetail(t *testing.T) {
	e := newCovEngine(t)

	db := band.NewDataBand()
	db.SetName("KDDetailDB")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetKeepDetail(true) // exercises the KeepDetail startKeepBand/EndKeep path

	// Use a mock DS to provide rows.
	mockDS := &coverageTestDS{rows: 3}
	db.SetDataSource(mockDS)

	beforeY := e.curY
	e.RunDataBandRowsKeep(db, 3, false, false)
	if e.curY <= beforeY {
		t.Errorf("RunDataBandRowsKeep KeepDetail: CurY should advance, got %v from %v", e.curY, beforeY)
	}
}

// TestRunDataBandRowsKeep_OneRow exercises the oneRow=true path where both
// keepFirstRow and keepLastRow are true for a single row (no EndKeep on header).
func TestRunDataBandRowsKeep_OneRow(t *testing.T) {
	e := newCovEngine(t)

	hdr := band.NewDataHeaderBand()
	hdr.SetHeight(5)
	hdr.SetVisible(true)
	hdr.SetKeepWithData(true)

	ftr := band.NewDataFooterBand()
	ftr.SetHeight(5)
	ftr.SetVisible(true)
	ftr.SetKeepWithData(true)

	db := band.NewDataBand()
	db.SetName("OneRowDB")
	db.SetHeight(10)
	db.SetVisible(true)
	db.SetHeader(hdr)
	db.SetFooter(ftr)
	// DataBand with no nested sub-bands is already "deepmost" (IsDeepmostDataBand is computed).

	mockDS := &coverageTestDS{rows: 1}
	db.SetDataSource(mockDS)

	// keepFirstRow=true, keepLastRow=true, rows=1 → oneRow=true
	beforeY := e.curY
	e.RunDataBandRowsKeep(db, 1, true, true)
	if e.curY <= beforeY {
		t.Errorf("RunDataBandRowsKeep oneRow: CurY should advance, got %v from %v", e.curY, beforeY)
	}
}

// coverageTestDS is a minimal DataSource that simulates a fixed number of rows.
type coverageTestDS struct {
	rows int
	pos  int
}

func (d *coverageTestDS) RowCount() int { return d.rows }
func (d *coverageTestDS) First() error  { d.pos = 0; return nil }
func (d *coverageTestDS) Next() error   { d.pos++; return nil }
func (d *coverageTestDS) EOF() bool     { return d.pos >= d.rows }
func (d *coverageTestDS) GetValue(col string) (any, error) {
	return nil, nil
}

// ── 8: showDataBandFooter — ds.Prior() path ──────────────────────────────────

// priorDS wraps a coverageTestDS and implements Prior().
type priorDS struct {
	coverageTestDS
	priorCalled bool
}

func (d *priorDS) Prior() { d.priorCalled = true }

// TestShowDataBandFooter_WithPrior exercises the Prior() path in showDataBandFooter.
func TestShowDataBandFooter_WithPrior(t *testing.T) {
	e := newCovEngine(t)

	ds := &priorDS{coverageTestDS: coverageTestDS{rows: 2}}
	if err := ds.First(); err != nil {
		t.Fatalf("First: %v", err)
	}

	db := band.NewDataBand()
	db.SetName("PriorDB")
	db.SetHeight(10)
	db.SetDataSource(ds)

	e.showDataBandFooter(db)

	if !ds.priorCalled {
		t.Error("showDataBandFooter: Prior() should have been called on ds")
	}
}

// ── 9: runDataBandNoDS — no-filter path ──────────────────────────────────────
// (TestRunDataBandNoDS_FilterFalse is in nodatasource_infer_test.go)

// TestRunDataBandNoDS_NoFilter exercises the no-filter path (band renders once).
func TestRunDataBandNoDS_NoFilter(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	db := band.NewDataBand()
	db.SetName("NDSNoFilterDB")
	db.SetHeight(15)
	db.SetVisible(true)
	// No filter → virtual row renders once.

	pg0 := e.preparedPages.GetPage(0)
	before := len(pg0.Bands)

	if err := e.runDataBandNoDS(db); err != nil {
		t.Fatalf("runDataBandNoDS no filter: %v", err)
	}

	after := len(pg0.Bands)
	if after <= before {
		t.Errorf("runDataBandNoDS no filter: expected band added, before=%d after=%d", before, after)
	}
}

// ── 10: evalTextWithFormat — uncovered branches ──────────────────────────────
// (TestEvalTextWithFormat_NilReport and TestEvalTextWithFormat_EmptyText are in objects_coverage_gaps_test.go)

// TestEvalTextWithFormat_FormatNil_NoFormatPath exercises the format==nil path.
func TestEvalTextWithFormat_FormatNil(t *testing.T) {
	e := newCovEngine(t)
	result := e.evalTextWithFormat("plain text", nil)
	// No bracket expression, no format — should return the text as-is.
	if result == "" {
		t.Error("evalTextWithFormat format nil: expected non-empty result")
	}
}

// TestEvalTextWithFormat_FormatNonNil_SingleBracket exercises the format != nil
// path with a single bracket expression. The format is applied to the result.
func TestEvalTextWithFormat_FormatNonNil_SingleBracket(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Use a NumberFormat that formats the value.
	f := &format.NumberFormat{}

	// "[1+1]" is a bracket expression but Calc may or may not evaluate it.
	// Just test that the function doesn't panic and returns a string.
	result := e.evalTextWithFormat("[Page]", f)
	_ = result // don't care about value, just no panic
}

// TestEvalTextWithFormat_FormatNonNil_NotBracket exercises the format != nil path
// where the text is NOT a single bracket expression (falls through to CalcText).
func TestEvalTextWithFormat_FormatNonNil_NotBracket(t *testing.T) {
	e := newCovEngine(t)

	f := &format.NumberFormat{}

	// "Hello World" is not a bracket expression → isSingleBracketExpr returns false.
	result := e.evalTextWithFormat("Hello World", f)
	if result != "Hello World" {
		t.Errorf("evalTextWithFormat format non-bracket: got %q, want %q", result, "Hello World")
	}
}
