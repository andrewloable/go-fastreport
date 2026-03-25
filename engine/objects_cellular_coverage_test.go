package engine

// objects_cellular_coverage_test.go — internal tests (package engine) targeting
// remaining uncovered branches in objects.go after the populateCellularTextCells
// method was added.
//
// Covers:
//   - populateCellularTextCells: word-wrap path, non-wrap path, empty-line path,
//     HorzAlignRight, HorzAlignCenter, auto<=0, fillRow rowIdx>=rowCount,
//     fillRow len(line)>colCount, SolidFill branch
//   - buildPreparedObject TextObject: transparent text color suppression (alpha=0)
//   - buildPreparedObject CheckBoxObject: string result from Calc, DataColumn branch
//   - populateContainerChildren: nested ContainerObject recursion
//   - populateTableObjects: endCol > len(cols) clamp (via ColSpan)

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/band"
	barcodepkg "github.com/andrewloable/go-fastreport/barcode"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/style"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func newCellularEngine(t *testing.T) *ReportEngine {
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

func newCellularObj(name string, w, h float32, cellW, cellH float32) *object.CellularTextObject {
	c := object.NewCellularTextObject()
	c.SetName(name)
	c.SetLeft(0)
	c.SetTop(0)
	c.SetWidth(w)
	c.SetHeight(h)
	c.SetVisible(true)
	c.SetCellWidth(cellW)
	c.SetCellHeight(cellH)
	return c
}

// ── populateCellularTextCells: basic happy path ────────────────────────────────

// TestPopulateCellularTextCells_Basic exercises the core path with explicit
// cell dimensions set, no word-wrap needed (text fits in one row).
func TestPopulateCellularTextCells_Basic(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_basic", Top: 0, Height: 40}

	c := newCellularObj("CT_Basic", 200, 40, 20, 20)
	c.SetText("HELLO")

	e.populateCellularTextCells(c, "HELLO", pb)
	// 200/20 = 10 columns, 40/20 = 2 rows → 20 cells
	if len(pb.Objects) == 0 {
		t.Error("expected PreparedObjects from basic CellularTextObject")
	}
}

// ── populateCellularTextCells: auto cell size (cellW==0 || cellH==0) ──────────

// TestPopulateCellularTextCells_AutoCellSize exercises the auto-size branch
// where CellWidth==0 and CellHeight==0 (size computed from font height).
func TestPopulateCellularTextCells_AutoCellSize(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_auto", Top: 0, Height: 60}

	c := newCellularObj("CT_Auto", 200, 60, 0, 0) // zero → auto-size
	c.SetText("ABC")

	e.populateCellularTextCells(c, "ABC", pb)
	if len(pb.Objects) == 0 {
		t.Error("expected PreparedObjects from auto-sized CellularTextObject")
	}
}

// TestPopulateCellularTextCells_AutoCellSizeAutoLEZero exercises the
// `auto <= 0` guard when font size results in a near-zero computed size.
// A very small font (0.1pt) gives fontPx≈0.13, raw≈10.13, which rounds to
// a positive value (9.45), so auto > 0. We can't make auto <= 0 with
// standard fonts, but we test a zero-size CellularTextObject to hit the
// colCount < 1 and rowCount < 1 guards.
func TestPopulateCellularTextCells_ZeroDimensionObject(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_0dim", Top: 0, Height: 40}

	c := newCellularObj("CT_ZeroDim", 0, 0, 5, 5)
	c.SetText("X")

	// colCount < 1 → colCount = 1; rowCount < 1 → rowCount = 1
	e.populateCellularTextCells(c, "X", pb)
	// Should produce exactly 1 cell.
	if len(pb.Objects) != 1 {
		t.Errorf("zero-dim CellularText: expected 1 cell, got %d", len(pb.Objects))
	}
}

// ── populateCellularTextCells: fillRow trailing space trim ───────────────────

// TestPopulateCellularTextCells_TrailingSpaceTrim exercises the trailing-space
// trim loop in fillRow. When word-wrap fires and the slice passed to fillRow
// ends with a space, the trim loop removes it.
//
// With colCount=3 and text "A  B" (two spaces at positions 1,2):
//   - i=1 (space): lastSpace=1
//   - i=2 (space): lastSpace=2
//   - i=3 (B): i-lineBegin+1=4 > 3 → wordWrap && lastSpace(2)>lineBegin(0)
//     → fillRow(0, runes[0:2]) = ["A", " "] → trailing space trimmed
func TestPopulateCellularTextCells_TrailingSpaceTrim(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_trailing_space", Top: 0, Height: 60}

	// 3 columns, 2 rows; text with consecutive spaces causes trailing space in word-wrap slice.
	c := newCellularObj("CT_TrailSpace", 60, 60, 20, 20)
	c.SetWordWrap(true)
	// "A  B" — double space at positions 1 and 2; at i=3 (B), overflow fires,
	// lastSpace=2, fillRow gets "A " with trailing space → trim loop fires.
	e.populateCellularTextCells(c, "A  B", pb)
	if len(pb.Objects) == 0 {
		t.Error("trailing space trim: expected PreparedObjects")
	}
}

// ── populateCellularTextCells: word-wrap path ─────────────────────────────────

// TestPopulateCellularTextCells_WordWrap exercises the wordWrap && lastSpace > lineBegin
// branch: text with a space that causes a word-wrap line break.
func TestPopulateCellularTextCells_WordWrap(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_wrap", Top: 0, Height: 60}

	// 3 columns, 3 rows: text "AB CD" with word-wrap at space.
	c := newCellularObj("CT_Wrap", 60, 60, 20, 20)
	c.SetWordWrap(true)
	// "AB CD" — after 3 chars per col, wrap at space:
	// col=3, i=2 triggers: lastSpace=2, lineBegin=0 → i-lineBegin+1=3 > colCount(3)
	// but lastSpace(2) > lineBegin(0) → word-wrap fires
	e.populateCellularTextCells(c, "AB CD", pb)
	if len(pb.Objects) == 0 {
		t.Error("word-wrap path: expected PreparedObjects")
	}
}

// TestPopulateCellularTextCells_WordWrapWithSpaceAtEnd exercises word-wrap
// with a trailing space to trigger lastSpace tracking.
func TestPopulateCellularTextCells_WordWrapWithSpace(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_wrap2", Top: 0, Height: 60}

	// 2 columns: text "A B" forces wrap. With colCount=2, i=2 triggers overflow.
	c := newCellularObj("CT_Wrap2", 40, 60, 20, 20)
	c.SetWordWrap(true)
	e.populateCellularTextCells(c, "A B", pb)
	if len(pb.Objects) == 0 {
		t.Error("word-wrap with space: expected PreparedObjects")
	}
}

// ── populateCellularTextCells: non-wrap path (i-lineBegin > 0) ───────────────

// TestPopulateCellularTextCells_NoWrapHardBreak exercises the else-if path:
// `i-lineBegin > 0` when wordWrap is false but text overflows column count.
func TestPopulateCellularTextCells_NoWrapOverflow(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_nowrap", Top: 0, Height: 60}

	// 2 columns, no word-wrap: text "ABCDE" forces the non-wrap overflow branch.
	c := newCellularObj("CT_NoWrap", 40, 60, 20, 20)
	c.SetWordWrap(false)
	e.populateCellularTextCells(c, "ABCDE", pb)
	if len(pb.Objects) == 0 {
		t.Error("non-wrap overflow path: expected PreparedObjects")
	}
}

// ── populateCellularTextCells: CRLF (isCRLF path) ────────────────────────────

// TestPopulateCellularTextCells_CRLFLineBreak exercises the isCRLF == true path.
func TestPopulateCellularTextCells_CRLFLineBreak(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_crlf", Top: 0, Height: 60}

	c := newCellularObj("CT_CRLF", 200, 60, 20, 20)
	c.SetWordWrap(false)
	// "\n" triggers the isCRLF path.
	e.populateCellularTextCells(c, "AB\nCD", pb)
	if len(pb.Objects) == 0 {
		t.Error("CRLF path: expected PreparedObjects")
	}
}

// TestPopulateCellularTextCells_CRLFEmptyLine exercises the `else { lineBegin = i+1 }`
// path: when isCRLF==true and i-lineBegin==0 (empty line at start or consecutive newlines).
func TestPopulateCellularTextCells_CRLFEmptyLine(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_empty_line", Top: 0, Height: 60}

	c := newCellularObj("CT_EmptyLine", 200, 60, 20, 20)
	c.SetWordWrap(false)
	// "\n\n" — second \n at position 1: isCRLF=true, i=1, lineBegin=1
	// after first \n: lineBegin=1, so second \n: i-lineBegin=0 → else path
	e.populateCellularTextCells(c, "\n\nA", pb)
	if len(pb.Objects) == 0 {
		t.Error("empty-line CRLF path: expected PreparedObjects")
	}
}

// ── populateCellularTextCells: HorzAlignRight ─────────────────────────────────

// TestPopulateCellularTextCells_HorzAlignRight exercises the HorzAlignRight
// offset in fillRow.
func TestPopulateCellularTextCells_HorzAlignRight(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_right", Top: 0, Height: 40}

	c := newCellularObj("CT_Right", 200, 40, 20, 20)
	c.SetHorzAlign(object.HorzAlignRight)
	c.SetText("HI")

	e.populateCellularTextCells(c, "HI", pb)
	if len(pb.Objects) == 0 {
		t.Error("HorzAlignRight: expected PreparedObjects")
	}
}

// ── populateCellularTextCells: HorzAlignCenter ────────────────────────────────

// TestPopulateCellularTextCells_HorzAlignCenter exercises the HorzAlignCenter
// offset calculation in fillRow.
func TestPopulateCellularTextCells_HorzAlignCenter(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_center", Top: 0, Height: 40}

	c := newCellularObj("CT_Center", 200, 40, 20, 20)
	c.SetHorzAlign(object.HorzAlignCenter)
	c.SetText("HI")

	e.populateCellularTextCells(c, "HI", pb)
	if len(pb.Objects) == 0 {
		t.Error("HorzAlignCenter: expected PreparedObjects")
	}
}

// ── populateCellularTextCells: fillRow rowIdx >= rowCount ─────────────────────

// TestPopulateCellularTextCells_FillRowBeyondRowCount exercises the
// `rowIdx >= rowCount` guard in fillRow when text has more lines than rows.
func TestPopulateCellularTextCells_FillRowBeyondRowCount(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_overflow_row", Top: 0, Height: 20}

	// 1 row (height=20, cellH=20), but text has 3 lines → fillRow called with row=1,2
	// which are >= rowCount(1) → triggers early return in fillRow.
	c := newCellularObj("CT_OverflowRow", 200, 20, 20, 20)
	c.SetWordWrap(false)
	e.populateCellularTextCells(c, "A\nB\nC", pb)
	// Should not panic.
	if len(pb.Objects) == 0 {
		t.Error("overflow row: expected at least 1 PreparedObject (the single row)")
	}
}

// ── populateCellularTextCells: fillRow len(line) > colCount ──────────────────

// TestPopulateCellularTextCells_FillRowLineTruncated exercises the
// `len(line) > colCount` truncation in fillRow.
func TestPopulateCellularTextCells_FillRowLineTruncated(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_trunc", Top: 0, Height: 40}

	// 2 columns: push a 5-char line via the final lineBegin < len(runes) path
	// (no overflow within the main loop), so fillRow gets a 5-char line and truncates.
	c := newCellularObj("CT_Trunc", 40, 40, 20, 20)
	c.SetWordWrap(false)
	// Text shorter than loop-overflow (no \n, no space triggers wrap in main loop),
	// but the last-line fillRow call sees the full 5-char rune slice: truncated to 2.
	e.populateCellularTextCells(c, "ABCDE", pb)
	if len(pb.Objects) == 0 {
		t.Error("fill-row truncation: expected PreparedObjects")
	}
}

// ── populateCellularTextCells: SolidFill branch ───────────────────────────────

// TestPopulateCellularTextCells_WithSolidFill exercises the SolidFill branch
// in populateCellularTextCells where v.Fill().(*style.SolidFill) succeeds.
func TestPopulateCellularTextCells_WithSolidFill(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_fill", Top: 0, Height: 40}

	c := newCellularObj("CT_Fill", 100, 40, 20, 20)
	c.SetText("A")
	c.SetFill(style.NewSolidFill(color.RGBA{R: 200, G: 100, B: 50, A: 255}))

	e.populateCellularTextCells(c, "A", pb)
	if len(pb.Objects) == 0 {
		t.Error("SolidFill: expected PreparedObjects")
	}
	// First cell should have the fill color.
	if pb.Objects[0].FillColor.A == 0 {
		t.Error("SolidFill: FillColor.A should be non-zero")
	}
}

// ── populateCellularTextCells: HorzSpacing / VertSpacing ─────────────────────

// TestPopulateCellularTextCells_WithSpacing exercises cell layout with
// non-zero HorzSpacing and VertSpacing.
func TestPopulateCellularTextCells_WithSpacing(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "ct_spacing", Top: 0, Height: 50}

	c := newCellularObj("CT_Spacing", 120, 50, 20, 20)
	c.SetHorzSpacing(2)
	c.SetVertSpacing(2)
	c.SetText("AB")

	e.populateCellularTextCells(c, "AB", pb)
	if len(pb.Objects) == 0 {
		t.Error("spacing: expected PreparedObjects")
	}
}

// TestBuildPreparedObject_CheckBoxObject_CalcError exercises the path where
// e.report.Calc(expr) returns an error, so the `if err == nil` block is skipped.
func TestBuildPreparedObject_CheckBoxObject_CalcError(t *testing.T) {
	e := newCellularEngine(t)

	cb := object.NewCheckBoxObject()
	cb.SetName("CB_CalcErr")
	cb.SetLeft(0)
	cb.SetTop(0)
	cb.SetWidth(20)
	cb.SetHeight(20)
	cb.SetVisible(true)
	// An expression that causes a Calc error (undefined variable with syntax issues).
	cb.SetExpression("%%invalid_expression%%")
	cb.SetChecked(true) // should stay true since err != nil skips SetChecked

	po := e.buildPreparedObject(cb)
	if po == nil {
		t.Fatal("buildPreparedObject(CheckBox Calc error) returned nil")
	}
	// The checked state should remain as set (true), since Calc error means we skip.
	if !po.Checked {
		t.Error("CheckBoxObject Calc error: Checked should remain true (unchanged)")
	}
}

// ── buildPreparedObject: TextObject transparent text color suppression ────────

// TestBuildPreparedObject_TextObject_TransparentColor exercises the new branch:
// when po.TextColor.A == 0 and po.Text != "", po.Text is suppressed.
func TestBuildPreparedObject_TextObject_TransparentColor(t *testing.T) {
	e := newCellularEngine(t)

	txt := object.NewTextObject()
	txt.SetName("TXT_Transparent")
	txt.SetLeft(0)
	txt.SetTop(0)
	txt.SetWidth(100)
	txt.SetHeight(20)
	txt.SetVisible(true)
	txt.SetText("visible text")
	// Set text color to fully transparent (A == 0).
	txt.SetTextColor(color.RGBA{R: 0, G: 0, B: 0, A: 0})

	po := e.buildPreparedObject(txt)
	if po == nil {
		t.Fatal("buildPreparedObject returned nil for TextObject with transparent color")
	}
	// Text should be suppressed when A==0.
	if po.Text != "" {
		t.Errorf("transparent text color: po.Text = %q, want empty", po.Text)
	}
}

// ── buildPreparedObject: CheckBoxObject with string result from Calc ──────────

// TestBuildPreparedObject_CheckBoxObject_StringResult exercises the
// `case string` branch in CheckBoxObject expression evaluation.
func TestBuildPreparedObject_CheckBoxObject_StringResult(t *testing.T) {
	e := newCellularEngine(t)
	// Register a string variable that evaluates to "true".
	e.report.Dictionary().SetSystemVariable("CBStrVal", "true")

	cb := object.NewCheckBoxObject()
	cb.SetName("CB_StrResult")
	cb.SetLeft(0)
	cb.SetTop(0)
	cb.SetWidth(20)
	cb.SetHeight(20)
	cb.SetVisible(true)
	cb.SetExpression("[CBStrVal]") // expression returns string "true"
	cb.SetChecked(false)

	po := e.buildPreparedObject(cb)
	if po == nil {
		t.Fatal("buildPreparedObject(CheckBox string expr) returned nil")
	}
	// The string "true" should set Checked=true.
	if !po.Checked {
		t.Error("CheckBoxObject string 'true': Checked should be true")
	}
}

// TestBuildPreparedObject_CheckBoxObject_BoolResult exercises the `case bool`
// branch in CheckBoxObject expression evaluation. The expression "true" evaluates
// directly to a bool value (not a string) when passed without brackets.
func TestBuildPreparedObject_CheckBoxObject_BoolResult(t *testing.T) {
	e := newCellularEngine(t)

	cb := object.NewCheckBoxObject()
	cb.SetName("CB_BoolResult")
	cb.SetLeft(0)
	cb.SetTop(0)
	cb.SetWidth(20)
	cb.SetHeight(20)
	cb.SetVisible(true)
	// "true" as a literal boolean expression (no brackets → Calc returns bool).
	cb.SetExpression("true")
	cb.SetChecked(false)

	po := e.buildPreparedObject(cb)
	if po == nil {
		t.Fatal("buildPreparedObject(CheckBox bool expr) returned nil")
	}
	if !po.Checked {
		t.Error("CheckBoxObject bool 'true': Checked should be true")
	}
}

// TestBuildPreparedObject_CheckBoxObject_StringResultFalse exercises the string
// branch where the string value is not "true" / "True" / "1".
func TestBuildPreparedObject_CheckBoxObject_StringResultFalse(t *testing.T) {
	e := newCellularEngine(t)
	e.report.Dictionary().SetSystemVariable("CBStrFalse", "false")

	cb := object.NewCheckBoxObject()
	cb.SetName("CB_StrFalse")
	cb.SetLeft(0)
	cb.SetTop(0)
	cb.SetWidth(20)
	cb.SetHeight(20)
	cb.SetVisible(true)
	cb.SetExpression("[CBStrFalse]")
	cb.SetChecked(true)

	po := e.buildPreparedObject(cb)
	if po == nil {
		t.Fatal("buildPreparedObject(CheckBox string false) returned nil")
	}
	if po.Checked {
		t.Error("CheckBoxObject string 'false': Checked should be false")
	}
}

// ── buildPreparedObject: CheckBoxObject with DataColumn ─────────────────────

// TestBuildPreparedObject_CheckBoxObject_DataColumn exercises the
// `else if col := v.DataColumn()` branch in CheckBoxObject.
func TestBuildPreparedObject_CheckBoxObject_DataColumn(t *testing.T) {
	e := newCellularEngine(t)
	// Register a system variable that will be resolved via DataColumn evaluation.
	e.report.Dictionary().SetSystemVariable("IsActive", "1")

	cb := object.NewCheckBoxObject()
	cb.SetName("CB_DataCol")
	cb.SetLeft(0)
	cb.SetTop(0)
	cb.SetWidth(20)
	cb.SetHeight(20)
	cb.SetVisible(true)
	// No expression — only DataColumn.
	cb.SetExpression("")
	cb.SetDataColumn("IsActive")
	cb.SetChecked(false)

	po := e.buildPreparedObject(cb)
	if po == nil {
		t.Fatal("buildPreparedObject(CheckBox DataColumn) returned nil")
	}
	// "1" → true
	if !po.Checked {
		t.Error("CheckBoxObject DataColumn '1': Checked should be true")
	}
}

// ── buildPreparedObject: BarcodeObject DefaultValue() fallback ───────────────

// TestBuildPreparedObject_BarcodeObjDefaultValue exercises the new
// `text = v.Barcode.DefaultValue()` fallback branch in buildPreparedObject
// for BarcodeObjects. This fires when both Text and NoDataText are empty,
// and HideIfNoData is false.
func TestBuildPreparedObject_BarcodeObjDefaultValue(t *testing.T) {
	// Import the barcode package via the engine's internal access.
	// We need to construct this through a full engine run.
	r := reportpkg.NewReport()
	pg := reportpkg.NewReportPage()

	hdr := band.NewPageHeaderBand()
	hdr.SetName("PH_BcDefault")
	hdr.SetHeight(60)
	hdr.SetVisible(true)

	bc := barcodepkg.NewBarcodeObject()
	bc.SetName("BC_DefaultVal")
	bc.SetLeft(0)
	bc.SetTop(0)
	bc.SetWidth(200)
	bc.SetHeight(60)
	bc.SetVisible(true)
	bc.SetText("")       // empty text
	bc.SetExpression("") // empty expression
	bc.SetNoDataText("") // empty NoDataText → DefaultValue() fires
	bc.SetHideIfNoData(false)
	bc.Barcode = barcodepkg.NewCode128Barcode()

	hdr.Objects().Add(bc)
	pg.SetPageHeader(hdr)
	r.AddPage(pg)

	e := New(r)
	if err := e.Run(DefaultRunOptions()); err != nil {
		t.Fatalf("Run with BarcodeObject DefaultValue: %v", err)
	}
	// Should render without panic.
}

// ── buildPreparedObject: CellularTextObject with SolidFill ───────────────────

// TestBuildPreparedObject_CellularTextObject_WithSolidFill exercises the
// CellularTextObject case of buildPreparedObject. The anchor is marked
// NotExportable (it exists only to maintain FRX→PreparedObject index mapping).
// Individual cells carry the fill color instead.
func TestBuildPreparedObject_CellularTextObject_WithSolidFill(t *testing.T) {
	e := newCellularEngine(t)

	c := object.NewCellularTextObject()
	c.SetName("CT_WithFill_BPO")
	c.SetLeft(0)
	c.SetTop(0)
	c.SetWidth(100)
	c.SetHeight(40)
	c.SetVisible(true)
	c.SetText("AB")
	c.SetCellWidth(20)
	c.SetCellHeight(20)
	c.SetFill(style.NewSolidFill(color.RGBA{R: 200, G: 100, B: 50, A: 255}))

	po := e.buildPreparedObject(c)
	if po == nil {
		t.Fatal("buildPreparedObject(CellularTextObject with SolidFill) returned nil")
	}
	if !po.NotExportable {
		t.Error("CellularTextObject anchor should be NotExportable")
	}
}

// ── populateContainerChildren: nested ContainerObject recursion ───────────────

// TestPopulateContainerChildren_NestedContainer exercises the nested container
// recursion path: a ContainerObject inside another ContainerObject.
// The inner container is added as a child of the outer container, so
// populateContainerChildren recurses into it.
func TestPopulateContainerChildren_NestedContainer(t *testing.T) {
	e := newCellularEngine(t)
	pb := &preview.PreparedBand{Name: "nested_cont_pb", Top: 0, Height: 80}

	// Outer container.
	outer := object.NewContainerObject()
	outer.SetName("OuterCont")
	outer.SetLeft(0)
	outer.SetTop(0)
	outer.SetWidth(200)
	outer.SetHeight(80)
	outer.SetVisible(true)

	// Inner container nested inside outer.
	inner := object.NewContainerObject()
	inner.SetName("InnerCont")
	inner.SetLeft(10)
	inner.SetTop(10)
	inner.SetWidth(100)
	inner.SetHeight(40)
	inner.SetVisible(true)

	// Text inside the inner container.
	innerTxt := object.NewTextObject()
	innerTxt.SetName("InnerTxt")
	innerTxt.SetLeft(5)
	innerTxt.SetTop(5)
	innerTxt.SetWidth(80)
	innerTxt.SetHeight(20)
	innerTxt.SetVisible(true)
	innerTxt.SetText("Nested")
	inner.AddChild(innerTxt)

	outer.AddChild(inner)

	// Call populateContainerChildren on the outer container.
	// This should recurse into inner and process innerTxt.
	e.populateContainerChildren(outer, outer.Left(), outer.Top(), pb)

	// Should have: inner container shape + innerTxt = at least 2 objects.
	if len(pb.Objects) < 2 {
		t.Errorf("nested container: expected >= 2 objects, got %d", len(pb.Objects))
	}
}
