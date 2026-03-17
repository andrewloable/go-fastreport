package engine

import (
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func newKWDEngine(t *testing.T) *ReportEngine {
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

func TestCheckKeepFooterWithData_NilFooter(t *testing.T) {
	e := newKWDEngine(t)
	e.checkKeepFooterWithData(nil, 0)
}

func TestCheckKeepFooterWithData_KeepWithDataFalse(t *testing.T) {
	e := newKWDEngine(t)
	ftr := band.NewDataFooterBand()
	ftr.SetKeepWithData(false)
	ftr.SetHeight(10)
	e.checkKeepFooterWithData(ftr, 0)
}

func TestCheckKeepFooterWithData_FooterFits(t *testing.T) {
	e := newKWDEngine(t)
	ftr := band.NewDataFooterBand()
	ftr.SetKeepWithData(true)
	ftr.SetHeight(5)
	before := e.curY
	e.checkKeepFooterWithData(ftr, 0)
	if e.curY != before {
		t.Errorf("curY changed from %v to %v", before, e.curY)
	}
}

func TestCheckKeepFooterWithData_NilPreparedPages(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	e.freeSpace = 0
	ftr := band.NewDataFooterBand()
	ftr.SetKeepWithData(true)
	ftr.SetHeight(20)
	e.checkKeepFooterWithData(ftr, 0)
}

func TestCheckKeepFooterWithData_DoesNotFit_CutsAndPastes(t *testing.T) {
	e := newKWDEngine(t)
	e.freeSpace = 2
	ftr := band.NewDataFooterBand()
	ftr.SetKeepWithData(true)
	ftr.SetHeight(50)
	e.keepCurY = e.curY
	e.checkKeepFooterWithData(ftr, 0)
}

// TestCheckKeepFooterWithData_FreeSpaceExhausted covers the else branch
// (e.freeSpace = 0) when keepDeltaY >= freeSpace on the new page.
// keepDeltaY = curY - keepCurY. A4 page height ≈ 1047 px; setting keepCurY=0
// and curY=2000 gives keepDeltaY=2000 which exceeds the new page's freeSpace.
func TestCheckKeepFooterWithData_FreeSpaceExhausted(t *testing.T) {
	e := newKWDEngine(t)
	ftr := band.NewDataFooterBand()
	ftr.SetKeepWithData(true)
	ftr.SetHeight(50)
	// keepCurY=0, curY=2000 → keepDeltaY=2000 > pageHeight ≈ 1047
	e.keepCurY = 0
	e.curY = 2000
	e.freeSpace = 1 // footer doesn't fit on current page
	e.checkKeepFooterWithData(ftr, 0)
	if e.freeSpace != 0 {
		t.Errorf("expected freeSpace=0 after keepDeltaY exhausts free space, got %v", e.freeSpace)
	}
}

// TestCheckKeepFooterWithData_NilPreparedPages_Direct covers the pp==nil branch
// by explicitly zeroing preparedPages after construction.
func TestCheckKeepFooterWithData_NilPreparedPages_Direct(t *testing.T) {
	e := newKWDEngine(t)
	e.preparedPages = nil // force the nil branch
	e.freeSpace = 0       // ensure we pass the footerH <= freeSpace guard
	ftr := band.NewDataFooterBand()
	ftr.SetKeepWithData(true)
	ftr.SetHeight(20) // 20 > 0 so we enter the cut path
	e.checkKeepFooterWithData(ftr, 0)
}

func TestCheckKeepHeaderWithData_NilHdr(t *testing.T) {
	e := newKWDEngine(t)
	db := band.NewDataBand()
	e.checkKeepHeaderWithData(nil, db, 0)
}

func TestCheckKeepHeaderWithData_NilDB(t *testing.T) {
	e := newKWDEngine(t)
	hdr := band.NewDataHeaderBand()
	hdr.SetKeepWithData(true)
	e.checkKeepHeaderWithData(hdr, nil, 0)
}

func TestCheckKeepHeaderWithData_KeepWithDataFalse(t *testing.T) {
	e := newKWDEngine(t)
	hdr := band.NewDataHeaderBand()
	hdr.SetKeepWithData(false)
	db := band.NewDataBand()
	db.SetHeight(10)
	e.checkKeepHeaderWithData(hdr, db, 0)
}

func TestCheckKeepHeaderWithData_Fits(t *testing.T) {
	e := newKWDEngine(t)
	hdr := band.NewDataHeaderBand()
	hdr.SetKeepWithData(true)
	hdr.SetHeight(5)
	db := band.NewDataBand()
	db.SetHeight(5)
	before := e.curY
	e.checkKeepHeaderWithData(hdr, db, 0)
	if e.curY != before {
		t.Errorf("curY changed from %v to %v", before, e.curY)
	}
}

func TestCheckKeepHeaderWithData_NilPreparedPages(t *testing.T) {
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()
	r.AddPage(pg)
	e := New(r)
	e.freeSpace = 0
	hdr := band.NewDataHeaderBand()
	hdr.SetKeepWithData(true)
	hdr.SetHeight(30)
	db := band.NewDataBand()
	db.SetHeight(30)
	e.checkKeepHeaderWithData(hdr, db, 0)
}

func TestCheckKeepHeaderWithData_DoesNotFit_CutsAndPastes(t *testing.T) {
	e := newKWDEngine(t)
	e.freeSpace = 2
	hdr := band.NewDataHeaderBand()
	hdr.SetKeepWithData(true)
	hdr.SetHeight(30)
	db := band.NewDataBand()
	db.SetHeight(30)
	e.keepCurY = e.curY
	e.checkKeepHeaderWithData(hdr, db, 0)
}

// TestCheckKeepHeaderWithData_FreeSpaceExhausted covers the else branch
// (e.freeSpace = 0) when keepDeltaY >= freeSpace on the new page.
func TestCheckKeepHeaderWithData_FreeSpaceExhausted(t *testing.T) {
	e := newKWDEngine(t)
	hdr := band.NewDataHeaderBand()
	hdr.SetKeepWithData(true)
	hdr.SetHeight(30)
	db := band.NewDataBand()
	db.SetHeight(30)
	// keepCurY=0, curY=2000 → keepDeltaY=2000 > pageHeight ≈ 1047
	e.keepCurY = 0
	e.curY = 2000
	e.freeSpace = 1 // header+row doesn't fit on current page
	e.checkKeepHeaderWithData(hdr, db, 0)
	if e.freeSpace != 0 {
		t.Errorf("expected freeSpace=0 after keepDeltaY exhausts free space, got %v", e.freeSpace)
	}
}

// TestCheckKeepHeaderWithData_NilPreparedPages_Direct covers the pp==nil branch
// by explicitly zeroing preparedPages after construction.
func TestCheckKeepHeaderWithData_NilPreparedPages_Direct(t *testing.T) {
	e := newKWDEngine(t)
	e.preparedPages = nil // force the nil branch
	e.freeSpace = 0       // ensure we pass the headerH+rowH <= freeSpace guard
	hdr := band.NewDataHeaderBand()
	hdr.SetKeepWithData(true)
	hdr.SetHeight(20)
	db := band.NewDataBand()
	db.SetHeight(20) // 40 > 0 so we enter the cut path
	e.checkKeepHeaderWithData(hdr, db, 0)
}

func TestExtractBracketedNames_NoOpenBracket(t *testing.T) {
	result := extractBracketedNames("val > 3")
	if len(result) != 0 {
		t.Errorf("got %v, want empty", result)
	}
}

func TestExtractBracketedNames_EmptyBrackets(t *testing.T) {
	result := extractBracketedNames("[] + [val]")
	if len(result) != 1 || result[0] != "val" {
		t.Errorf("got %v, want [val]", result)
	}
}

func TestExtractBracketedNames_UnclosedBracket(t *testing.T) {
	result := extractBracketedNames("[unclosed")
	if len(result) != 0 {
		t.Errorf("got %v, want empty", result)
	}
}

func TestConvertBracketExpr_NoOpenBracket(t *testing.T) {
	got := convertBracketExpr("val > 3")
	if got != "val > 3" {
		t.Errorf("got %q, want %q", got, "val > 3")
	}
}

func TestConvertBracketExpr_UnclosedBracket(t *testing.T) {
	got := convertBracketExpr("prefix [unclosed")
	if got != "prefix [unclosed" {
		t.Errorf("got %q, want %q", got, "prefix [unclosed")
	}
}

func TestConvertBracketExpr_MultipleBrackets(t *testing.T) {
	got := convertBracketExpr("[A] + [B]")
	if got != "A + B" {
		t.Errorf("got %q, want %q", got, "A + B")
	}
}

func TestGroupTreeItem_FirstItem_NonEmpty(t *testing.T) {
	child := &groupTreeItem{rowNo: 1}
	g := &groupTreeItem{items: []*groupTreeItem{child}}
	if g.firstItem() != child {
		t.Error("firstItem should return items[0]")
	}
}

func TestGroupTreeItem_LastItem_NonEmpty(t *testing.T) {
	c1 := &groupTreeItem{rowNo: 1}
	c2 := &groupTreeItem{rowNo: 2}
	g := &groupTreeItem{items: []*groupTreeItem{c1, c2}}
	if g.lastItem() != c2 {
		t.Error("lastItem should return the last item")
	}
}

func TestMakeGroupTree_NilDataBand(t *testing.T) {
	e := newKWDEngine(t)
	gh := band.NewGroupHeaderBand()
	root := e.makeGroupTree(gh)
	if root == nil {
		t.Fatal("should return non-nil root")
	}
	if len(root.items) != 0 {
		t.Errorf("expected 0 items, got %d", len(root.items))
	}
}

func TestMakeGroupTree_NilDataSource(t *testing.T) {
	e := newKWDEngine(t)
	gh := band.NewGroupHeaderBand()
	db := band.NewDataBand()
	gh.SetData(db)
	root := e.makeGroupTree(gh)
	if root == nil {
		t.Fatal("should return non-nil root")
	}
	if len(root.items) != 0 {
		t.Errorf("expected 0 items, got %d", len(root.items))
	}
}

func TestShowDataHeader_NilHeader(t *testing.T) {
	e := newKWDEngine(t)
	gh := band.NewGroupHeaderBand()
	db := band.NewDataBand()
	db.SetName("DBNoHdr")
	db.SetHeight(10)
	gh.SetData(db)
	e.showDataHeader(gh)
}

func TestShowDataFooter_NilFooter(t *testing.T) {
	e := newKWDEngine(t)
	gh := band.NewGroupHeaderBand()
	db := band.NewDataBand()
	db.SetName("DBNoFtr2")
	db.SetHeight(10)
	gh.SetData(db)
	e.showDataFooter(gh)
}

func TestShowGroupTree_LeafWithZeroRows(t *testing.T) {
	e := newKWDEngine(t)
	gh := band.NewGroupHeaderBand()
	gh.SetName("LeafGH")
	gh.SetVisible(true)
	gh.SetHeight(10)
	db := band.NewDataBand()
	db.SetName("LeafDB")
	db.SetHeight(10)
	db.SetVisible(true)
	gh.SetData(db)
	root := &groupTreeItem{band: gh, rowCount: 0, items: nil}
	e.showGroupTree(root)
}
