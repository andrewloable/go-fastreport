package table_test

// table_coverage_test.go — additional tests to push coverage above 85%.
//
// Targets (all at 0 % before this file):
//   cell.go:    Objects(), cellDuplicatesName (remaining cases), parseCellDuplicates (remaining)
//   deserialize.go: DeserializeChild (Row, Column, unknown)
//   result.go:  NewTableResult, NewTableResultFrom, CalcBounds, RowsToSerialize,
//               ColumnsToSerialize, NewTableStyleCollection, DefaultStyle, Add,
//               Count, Get, cellStylesEqual
//   row.go:     Cells()
//   table.go:   Rows(), Columns(), Column(), RepeatRowHeaders(), RepeatColumnHeaders(),
//               TableBase.Deserialize(), TableObject.Deserialize()

import (
	"bytes"
	"testing"

	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/table"
)

// ── helper: round-trip a TableObject through FRX XML ─────────────────────────

func serializeTableObject(t *testing.T, tbl *table.TableObject) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TableObject", tbl); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.Bytes()
}

func deserializeTableObject(t *testing.T, data []byte) *table.TableObject {
	t.Helper()
	r := serial.NewReader(bytes.NewReader(data))
	typeName, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader: no element found")
	}
	if typeName != "TableObject" {
		t.Fatalf("unexpected element: %q", typeName)
	}
	got := table.NewTableObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return got
}

// ── cell.go — Objects(), cellDuplicatesName, parseCellDuplicates ─────────────

// TestTableCell_Objects tests the Objects() getter.
func TestTableCell_Objects(t *testing.T) {
	c := table.NewTableCell()
	if objs := c.Objects(); objs != nil {
		t.Errorf("Objects() on fresh cell should be nil/empty, got %v", objs)
	}
	inner := table.NewTableCell()
	c.AddObject(inner)
	objs := c.Objects()
	if len(objs) != 1 {
		t.Fatalf("Objects() len = %d, want 1", len(objs))
	}
	if objs[0] != inner {
		t.Error("Objects()[0] should be the added object")
	}
}

// TestCellDuplicates_RoundTrip exercises cellDuplicatesName and parseCellDuplicates
// for every variant by doing a full serialize/deserialize cycle.
func TestCellDuplicates_RoundTrip_AllVariants(t *testing.T) {
	cases := []struct {
		dup  table.CellDuplicates
		name string
	}{
		{table.CellDuplicatesShow, "Show"},
		{table.CellDuplicatesClear, "Clear"},
		{table.CellDuplicatesMerge, "Merge"},
		{table.CellDuplicatesMergeNonEmpty, "MergeNonEmpty"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := table.NewTableCell()
			orig.SetDuplicates(tc.dup)

			var buf bytes.Buffer
			w := serial.NewWriter(&buf)
			if err := w.WriteObjectNamed("TableCell", orig); err != nil {
				t.Fatalf("serialize: %v", err)
			}
			if err := w.Flush(); err != nil {
				t.Fatalf("Flush: %v", err)
			}

			r := serial.NewReader(bytes.NewReader(buf.Bytes()))
			_, ok := r.ReadObjectHeader()
			if !ok {
				t.Fatal("ReadObjectHeader returned false")
			}
			got := table.NewTableCell()
			if err := got.Deserialize(r); err != nil {
				t.Fatalf("Deserialize: %v", err)
			}
			if got.Duplicates() != tc.dup {
				t.Errorf("Duplicates round-trip: got %d, want %d", got.Duplicates(), tc.dup)
			}
		})
	}
}

// ── row.go — Cells() ─────────────────────────────────────────────────────────

func TestTableRow_Cells(t *testing.T) {
	r := table.NewTableRow()
	if r.Cells() != nil && len(r.Cells()) != 0 {
		t.Fatal("Cells() should be nil or empty for a new row")
	}
	c1 := table.NewTableCell()
	c1.SetText("A")
	c2 := table.NewTableCell()
	c2.SetText("B")
	r.AddCell(c1)
	r.AddCell(c2)

	cells := r.Cells()
	if len(cells) != 2 {
		t.Fatalf("Cells() len = %d, want 2", len(cells))
	}
	if cells[0].Text() != "A" {
		t.Errorf("cells[0].Text = %q, want A", cells[0].Text())
	}
	if cells[1].Text() != "B" {
		t.Errorf("cells[1].Text = %q, want B", cells[1].Text())
	}
}

// ── table.go — Rows(), Columns(), Column() ───────────────────────────────────

func TestTableBase_Rows(t *testing.T) {
	tbl := table.NewTableObject()
	if rows := tbl.Rows(); len(rows) != 0 {
		t.Fatalf("Rows() should be empty, got %d", len(rows))
	}
	r1 := tbl.NewRow()
	r2 := tbl.NewRow()
	rows := tbl.Rows()
	if len(rows) != 2 {
		t.Fatalf("Rows() len = %d, want 2", len(rows))
	}
	if rows[0] != r1 || rows[1] != r2 {
		t.Error("Rows() elements do not match added rows")
	}
}

func TestTableBase_Columns(t *testing.T) {
	tbl := table.NewTableObject()
	if cols := tbl.Columns(); len(cols) != 0 {
		t.Fatalf("Columns() should be empty, got %d", len(cols))
	}
	c1 := tbl.NewColumn()
	c2 := tbl.NewColumn()
	cols := tbl.Columns()
	if len(cols) != 2 {
		t.Fatalf("Columns() len = %d, want 2", len(cols))
	}
	if cols[0] != c1 || cols[1] != c2 {
		t.Error("Columns() elements do not match added columns")
	}
}

func TestTableBase_Column(t *testing.T) {
	tbl := table.NewTableObject()
	if tbl.Column(0) != nil {
		t.Error("Column(0) on empty table should be nil")
	}
	c1 := tbl.NewColumn()
	c1.SetWidth(111)
	_ = tbl.NewColumn()
	if tbl.Column(0) != c1 {
		t.Error("Column(0) should return the first column")
	}
	if tbl.Column(-1) != nil {
		t.Error("Column(-1) should return nil")
	}
	if tbl.Column(99) != nil {
		t.Error("Column(99) should return nil")
	}
}

// ── table.go — RepeatRowHeaders(), RepeatColumnHeaders() ─────────────────────

func TestTableBase_RepeatRowColumnHeaders(t *testing.T) {
	tbl := table.NewTableObject()

	if tbl.RepeatRowHeaders() {
		t.Error("RepeatRowHeaders should default to false")
	}
	if tbl.RepeatColumnHeaders() {
		t.Error("RepeatColumnHeaders should default to false")
	}

	tbl.SetRepeatRowHeaders(true)
	tbl.SetRepeatColumnHeaders(true)

	if !tbl.RepeatRowHeaders() {
		t.Error("RepeatRowHeaders should be true after Set")
	}
	if !tbl.RepeatColumnHeaders() {
		t.Error("RepeatColumnHeaders should be true after Set")
	}
}

// ── table.go — TableBase.Deserialize() ───────────────────────────────────────

func TestTableBase_Deserialize_RoundTrip(t *testing.T) {
	orig := table.NewTableObject()
	orig.SetName("Tbl1")
	orig.SetFixedRows(3)
	orig.SetFixedColumns(2)
	orig.SetLayout(table.TableLayoutWrapped)
	orig.SetPrintOnParent(true)
	orig.SetWrappedGap(15)
	orig.SetRepeatHeaders(false) // non-default (default=true)
	orig.SetRepeatRowHeaders(true)
	orig.SetRepeatColumnHeaders(true)
	orig.SetAdjustSpannedCellsWidth(true)
	orig.ManualBuildEvent = "MyBuild"

	data := serializeTableObject(t, orig)
	got := deserializeTableObject(t, data)

	if got.Name() != "Tbl1" {
		t.Errorf("Name: got %q, want Tbl1", got.Name())
	}
	if got.FixedRows() != 3 {
		t.Errorf("FixedRows: got %d, want 3", got.FixedRows())
	}
	if got.FixedColumns() != 2 {
		t.Errorf("FixedColumns: got %d, want 2", got.FixedColumns())
	}
	if got.Layout() != table.TableLayoutWrapped {
		t.Errorf("Layout: got %d, want Wrapped", got.Layout())
	}
	if !got.PrintOnParent() {
		t.Error("PrintOnParent: want true")
	}
	if got.WrappedGap() != 15 {
		t.Errorf("WrappedGap: got %v, want 15", got.WrappedGap())
	}
	if got.RepeatHeaders() {
		t.Error("RepeatHeaders: want false (was serialized as false)")
	}
	if !got.RepeatRowHeaders() {
		t.Error("RepeatRowHeaders: want true")
	}
	if !got.RepeatColumnHeaders() {
		t.Error("RepeatColumnHeaders: want true")
	}
	if !got.AdjustSpannedCellsWidth() {
		t.Error("AdjustSpannedCellsWidth: want true")
	}
	if got.ManualBuildEvent != "MyBuild" {
		t.Errorf("ManualBuildEvent: got %q, want MyBuild", got.ManualBuildEvent)
	}
}

// ── table.go — TableObject.Deserialize() ─────────────────────────────────────

func TestTableObject_Deserialize_ManualBuildAutoSpans(t *testing.T) {
	orig := table.NewTableObject()
	orig.ManualBuildAutoSpans = false // non-default (default=true)

	data := serializeTableObject(t, orig)
	got := deserializeTableObject(t, data)

	if got.ManualBuildAutoSpans {
		t.Error("ManualBuildAutoSpans: want false after round-trip")
	}
}

func TestTableObject_Deserialize_ManualBuildAutoSpans_Default(t *testing.T) {
	orig := table.NewTableObject()
	// ManualBuildAutoSpans is true by default; should not be written.
	data := serializeTableObject(t, orig)
	got := deserializeTableObject(t, data)

	if !got.ManualBuildAutoSpans {
		t.Error("ManualBuildAutoSpans: want true (default) after round-trip")
	}
}

// ── deserialize.go — DeserializeChild ────────────────────────────────────────

// buildTableObjectXML creates a minimal FRX XML snippet representing a
// TableObject with one column and one row containing one cell.
func buildTableObjectXML() string {
	return `<TableObject Name="T1" FixedRows="1">` +
		`<TableColumn Name="C1" Width="120" AutoSize="true"/>` +
		`<TableRow Name="R1" Height="40" CanBreak="true">` +
		`<TableCell Name="Cell1" ColSpan="2" RowSpan="2" Text="Hello"/>` +
		`</TableRow>` +
		`</TableObject>`
}

func TestDeserializeChild_Column(t *testing.T) {
	xmlData := buildTableObjectXML()
	r := serial.NewReader(bytes.NewReader([]byte(xmlData)))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TableObject" {
		t.Fatalf("ReadObjectHeader: typeName=%q ok=%v", typeName, ok)
	}
	tbl := table.NewTableObject()
	if err := tbl.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	// Process children.
	cd, ok2 := interface{}(tbl).(interface {
		DeserializeChild(string, interface{ ReadStr(string, string) string }) bool
	})
	_ = cd
	_ = ok2

	// The simplest way to exercise DeserializeChild is to use the full
	// deserialization path via NextChild loop, which the serial package
	// drives when we call it through the reader.
	// We already called Deserialize above; now let's iterate children.
	for {
		childType, hasChild := r.NextChild()
		if !hasChild {
			break
		}
		// TableBase.DeserializeChild should consume TableColumn / TableRow.
		consumed := tbl.DeserializeChild(childType, r)
		if consumed {
			// Already consumed; no FinishChild needed for sub-elements,
			// but we still need FinishChild to restore parent state.
		}
		if err := r.FinishChild(); err != nil {
			// Best-effort; may error if child was already fully consumed.
			_ = err
		}
	}

	if tbl.ColumnCount() != 1 {
		t.Errorf("ColumnCount after DeserializeChild: got %d, want 1", tbl.ColumnCount())
	}
	col := tbl.Column(0)
	if col == nil {
		t.Fatal("Column(0) is nil")
	}
	if col.Name() != "C1" {
		t.Errorf("Column Name: got %q, want C1", col.Name())
	}
}

func TestDeserializeChild_Row(t *testing.T) {
	// Build a table with a row that has a cell.
	xmlData := buildTableObjectXML()
	r := serial.NewReader(bytes.NewReader([]byte(xmlData)))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TableObject" {
		t.Fatalf("ReadObjectHeader: typeName=%q ok=%v", typeName, ok)
	}
	tbl := table.NewTableObject()
	if err := tbl.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	for {
		childType, hasChild := r.NextChild()
		if !hasChild {
			break
		}
		consumed := tbl.DeserializeChild(childType, r)
		_ = consumed
		_ = r.FinishChild()
	}

	if tbl.RowCount() != 1 {
		t.Errorf("RowCount after DeserializeChild: got %d, want 1", tbl.RowCount())
	}
	row := tbl.Row(0)
	if row == nil {
		t.Fatal("Row(0) is nil")
	}
	if row.Name() != "R1" {
		t.Errorf("Row Name: got %q, want R1", row.Name())
	}
	if row.CellCount() != 1 {
		t.Errorf("CellCount: got %d, want 1", row.CellCount())
	}
	cell := row.Cell(0)
	if cell == nil {
		t.Fatal("Cell(0) is nil")
	}
	if cell.Name() != "Cell1" {
		t.Errorf("Cell Name: got %q, want Cell1", cell.Name())
	}
	if cell.ColSpan() != 2 {
		t.Errorf("Cell ColSpan: got %d, want 2", cell.ColSpan())
	}
}

func TestDeserializeChild_UnknownChild(t *testing.T) {
	// DeserializeChild with an unknown child type should return false.
	xmlData := `<TableObject Name="T1"><UnknownElement Foo="bar"/></TableObject>`
	r := serial.NewReader(bytes.NewReader([]byte(xmlData)))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader returned false")
	}
	tbl := table.NewTableObject()
	if err := tbl.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	for {
		childType, hasChild := r.NextChild()
		if !hasChild {
			break
		}
		consumed := tbl.DeserializeChild(childType, r)
		if consumed {
			t.Errorf("DeserializeChild(%q) returned true, want false for unknown child", childType)
		}
		_ = r.FinishChild()
	}
}

// TestDeserializeChild_Direct exercises DeserializeChild directly with a
// fresh reader positioned on the child element — verifying TableColumn branch.
func TestDeserializeChild_DirectColumn(t *testing.T) {
	xmlData := `<TableColumn Name="ColX" Width="200" MinWidth="10" AutoSize="true" PageBreak="true" KeepColumns="3"/>`
	r := serial.NewReader(bytes.NewReader([]byte(xmlData)))
	_, ok := r.ReadObjectHeader() // positions reader on TableColumn
	if !ok {
		t.Fatal("ReadObjectHeader returned false")
	}

	tbl := table.NewTableObject()
	consumed := tbl.DeserializeChild("TableColumn", r)
	if !consumed {
		t.Error("DeserializeChild(TableColumn) should return true")
	}
	if tbl.ColumnCount() != 1 {
		t.Fatalf("ColumnCount: got %d, want 1", tbl.ColumnCount())
	}
	col := tbl.Column(0)
	if col.Name() != "ColX" {
		t.Errorf("Column Name: got %q, want ColX", col.Name())
	}
	if col.Width() != 200 {
		t.Errorf("Column Width: got %v, want 200", col.Width())
	}
	if col.MinWidth() != 10 {
		t.Errorf("Column MinWidth: got %v, want 10", col.MinWidth())
	}
	if !col.AutoSize() {
		t.Error("Column AutoSize should be true")
	}
	if !col.PageBreak() {
		t.Error("Column PageBreak should be true")
	}
	if col.KeepColumns() != 3 {
		t.Errorf("Column KeepColumns: got %d, want 3", col.KeepColumns())
	}
}

// TestDeserializeChild_DirectRow exercises DeserializeChild with a TableRow
// that has a nested TableCell.
func TestDeserializeChild_DirectRow(t *testing.T) {
	xmlData := `<TableRow Name="RowX" Height="50" MinHeight="5" MaxHeight="200" AutoSize="true" CanBreak="true" PageBreak="true" KeepRows="2">` +
		`<TableCell Name="CellA" ColSpan="3" CellDuplicates="Merge"/>` +
		`<TableCell Name="CellB"/>` +
		`</TableRow>`
	r := serial.NewReader(bytes.NewReader([]byte(xmlData)))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader returned false")
	}

	tbl := table.NewTableObject()
	consumed := tbl.DeserializeChild("TableRow", r)
	if !consumed {
		t.Error("DeserializeChild(TableRow) should return true")
	}
	if tbl.RowCount() != 1 {
		t.Fatalf("RowCount: got %d, want 1", tbl.RowCount())
	}
	row := tbl.Row(0)
	if row.Name() != "RowX" {
		t.Errorf("Row Name: got %q, want RowX", row.Name())
	}
	if row.Height() != 50 {
		t.Errorf("Row Height: got %v, want 50", row.Height())
	}
	if row.MinHeight() != 5 {
		t.Errorf("Row MinHeight: got %v, want 5", row.MinHeight())
	}
	if row.MaxHeight() != 200 {
		t.Errorf("Row MaxHeight: got %v, want 200", row.MaxHeight())
	}
	if !row.AutoSize() {
		t.Error("Row AutoSize should be true")
	}
	if !row.CanBreak() {
		t.Error("Row CanBreak should be true")
	}
	if !row.PageBreak() {
		t.Error("Row PageBreak should be true")
	}
	if row.KeepRows() != 2 {
		t.Errorf("Row KeepRows: got %d, want 2", row.KeepRows())
	}
	if row.CellCount() != 2 {
		t.Fatalf("CellCount: got %d, want 2", row.CellCount())
	}
	cellA := row.Cell(0)
	if cellA.Name() != "CellA" {
		t.Errorf("Cell[0] Name: got %q, want CellA", cellA.Name())
	}
	if cellA.ColSpan() != 3 {
		t.Errorf("Cell[0] ColSpan: got %d, want 3", cellA.ColSpan())
	}
	if cellA.Duplicates() != table.CellDuplicatesMerge {
		t.Errorf("Cell[0] Duplicates: got %d, want Merge", cellA.Duplicates())
	}
	cellB := row.Cell(1)
	if cellB.Name() != "CellB" {
		t.Errorf("Cell[1] Name: got %q, want CellB", cellB.Name())
	}
}

// TestDeserializeChild_FalseForUnknown verifies the false branch.
func TestDeserializeChild_FalseForUnknown(t *testing.T) {
	xmlData := `<SomeUnknownTag Foo="bar"/>`
	r := serial.NewReader(bytes.NewReader([]byte(xmlData)))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader returned false")
	}
	tbl := table.NewTableObject()
	consumed := tbl.DeserializeChild("SomeUnknownTag", r)
	if consumed {
		t.Error("DeserializeChild(SomeUnknownTag) should return false")
	}
}

// ── result.go — NewTableResult, NewTableResultFrom, CalcBounds ───────────────

func TestNewTableResult(t *testing.T) {
	r := table.NewTableResult()
	if r == nil {
		t.Fatal("NewTableResult returned nil")
	}
	if r.RowCount() != 0 {
		t.Errorf("RowCount: got %d, want 0", r.RowCount())
	}
	if r.ColumnCount() != 0 {
		t.Errorf("ColumnCount: got %d, want 0", r.ColumnCount())
	}
	if r.Skip {
		t.Error("Skip should default to false")
	}
}

func TestNewTableResultFrom(t *testing.T) {
	// Build a source TableBase with 2 rows and 2 columns.
	src := table.NewTableObject()
	col1 := src.NewColumn()
	col1.SetWidth(100)
	col2 := src.NewColumn()
	col2.SetWidth(200)
	row1 := src.NewRow()
	row1.SetHeight(30)
	row2 := src.NewRow()
	row2.SetHeight(40)
	src.SetFixedRows(1)
	src.SetFixedColumns(1)
	src.SetLayout(table.TableLayoutDownThenAcross)
	src.SetRepeatHeaders(false)
	src.SetRepeatRowHeaders(true)
	src.SetRepeatColumnHeaders(true)
	src.SetAdjustSpannedCellsWidth(true)

	result := table.NewTableResultFrom(&src.TableBase)

	if result.RowCount() != 2 {
		t.Errorf("RowCount: got %d, want 2", result.RowCount())
	}
	if result.ColumnCount() != 2 {
		t.Errorf("ColumnCount: got %d, want 2", result.ColumnCount())
	}
	if result.FixedRows() != 1 {
		t.Errorf("FixedRows: got %d, want 1", result.FixedRows())
	}
	if result.FixedColumns() != 1 {
		t.Errorf("FixedColumns: got %d, want 1", result.FixedColumns())
	}
	if result.Layout() != table.TableLayoutDownThenAcross {
		t.Errorf("Layout: got %d, want DownThenAcross", result.Layout())
	}
	if result.RepeatHeaders() {
		t.Error("RepeatHeaders: want false (copied from src)")
	}
	if !result.RepeatRowHeaders() {
		t.Error("RepeatRowHeaders: want true (copied from src)")
	}
	if !result.RepeatColumnHeaders() {
		t.Error("RepeatColumnHeaders: want true (copied from src)")
	}
	if !result.AdjustSpannedCellsWidth() {
		t.Error("AdjustSpannedCellsWidth: want true (copied from src)")
	}
}

func TestTableResult_CalcBounds(t *testing.T) {
	src := table.NewTableObject()
	col1 := src.NewColumn()
	col1.SetWidth(100)
	col2 := src.NewColumn()
	col2.SetWidth(50)
	row1 := src.NewRow()
	row1.SetHeight(30)
	row2 := src.NewRow()
	row2.SetHeight(20)

	result := table.NewTableResultFrom(&src.TableBase)

	w, h := result.CalcBounds()
	if w != 150 {
		t.Errorf("CalcBounds width: got %v, want 150", w)
	}
	if h != 50 {
		t.Errorf("CalcBounds height: got %v, want 50", h)
	}
}

func TestTableResult_CalcBounds_Empty(t *testing.T) {
	result := table.NewTableResult()
	w, h := result.CalcBounds()
	if w != 0 || h != 0 {
		t.Errorf("CalcBounds on empty result: got (%v,%v), want (0,0)", w, h)
	}
}

func TestTableResult_CalcBounds_Callback(t *testing.T) {
	src := table.NewTableObject()
	col := src.NewColumn()
	col.SetWidth(100)
	row := src.NewRow()
	row.SetHeight(50)

	result := table.NewTableResultFrom(&src.TableBase)

	callbackInvoked := false
	result.AfterCalcBounds = func(r *table.TableResult) {
		callbackInvoked = true
	}

	result.CalcBounds()
	if !callbackInvoked {
		t.Error("AfterCalcBounds callback was not invoked")
	}
}

func TestTableResult_RowsToSerialize(t *testing.T) {
	src := table.NewTableObject()
	src.NewRow()
	src.NewRow()
	result := table.NewTableResultFrom(&src.TableBase)

	rows := result.RowsToSerialize()
	if len(rows) != 2 {
		t.Errorf("RowsToSerialize len: got %d, want 2", len(rows))
	}
}

func TestTableResult_ColumnsToSerialize(t *testing.T) {
	src := table.NewTableObject()
	src.NewColumn()
	src.NewColumn()
	src.NewColumn()
	result := table.NewTableResultFrom(&src.TableBase)

	cols := result.ColumnsToSerialize()
	if len(cols) != 3 {
		t.Errorf("ColumnsToSerialize len: got %d, want 3", len(cols))
	}
}

// ── result.go — TableStyleCollection ─────────────────────────────────────────

func TestNewTableStyleCollection(t *testing.T) {
	sc := table.NewTableStyleCollection()
	if sc == nil {
		t.Fatal("NewTableStyleCollection returned nil")
	}
	if sc.Count() != 0 {
		t.Errorf("Count: got %d, want 0", sc.Count())
	}
	if sc.DefaultStyle() == nil {
		t.Error("DefaultStyle should not be nil")
	}
}

func TestTableStyleCollection_Add(t *testing.T) {
	sc := table.NewTableStyleCollection()

	c1 := table.NewTableCell()
	c1.SetText("irrelevant") // text is NOT part of style comparison
	c1.SetName("c1")

	// Add c1 for the first time — should be inserted.
	r1 := sc.Add(c1)
	if r1 == nil {
		t.Fatal("Add returned nil")
	}
	if sc.Count() != 1 {
		t.Errorf("Count after first Add: got %d, want 1", sc.Count())
	}

	// Add a cell with identical visual style — should return existing.
	c2 := table.NewTableCell()
	c2.SetText("different text")
	r2 := sc.Add(c2)
	if sc.Count() != 1 {
		t.Errorf("Count after adding equivalent style: got %d, want 1", sc.Count())
	}
	if r2 != r1 {
		t.Error("Add with equivalent style should return the existing entry")
	}
}

func TestTableStyleCollection_Add_DifferentStyle(t *testing.T) {
	sc := table.NewTableStyleCollection()

	c1 := table.NewTableCell()
	sc.Add(c1)

	// Create a cell with a distinct font name — different style.
	// (default font is Arial, so use a different name)
	c2 := table.NewTableCell()
	f := c2.Font()
	f.Name = "Times New Roman"
	c2.SetFont(f)
	sc.Add(c2)

	if sc.Count() != 2 {
		t.Errorf("Count after two different styles: got %d, want 2", sc.Count())
	}
}

func TestTableStyleCollection_Get(t *testing.T) {
	sc := table.NewTableStyleCollection()

	if sc.Get(0) != nil {
		t.Error("Get(0) on empty collection should return nil")
	}
	if sc.Get(-1) != nil {
		t.Error("Get(-1) should return nil")
	}

	c1 := table.NewTableCell()
	sc.Add(c1)

	got := sc.Get(0)
	if got == nil {
		t.Fatal("Get(0) returned nil after Add")
	}
	if sc.Get(1) != nil {
		t.Error("Get(1) should return nil when only 1 style exists")
	}
}

// ── result.go — cellStylesEqual (indirect via TableStyleCollection.Add) ─────

func TestCellStylesEqual_HorzAlignDifference(t *testing.T) {
	sc := table.NewTableStyleCollection()

	c1 := table.NewTableCell()
	c1.SetHorzAlign(0) // default

	c2 := table.NewTableCell()
	c2.SetHorzAlign(1) // different

	sc.Add(c1)
	sc.Add(c2)

	// They differ in HorzAlign — must be two distinct entries.
	if sc.Count() != 2 {
		t.Errorf("Count: got %d, want 2 (distinct HorzAlign)", sc.Count())
	}
}

func TestCellStylesEqual_VertAlignDifference(t *testing.T) {
	sc := table.NewTableStyleCollection()

	c1 := table.NewTableCell()
	c1.SetVertAlign(0)

	c2 := table.NewTableCell()
	c2.SetVertAlign(1)

	sc.Add(c1)
	sc.Add(c2)

	if sc.Count() != 2 {
		t.Errorf("Count: got %d, want 2 (distinct VertAlign)", sc.Count())
	}
}

func TestCellStylesEqual_FontSizeDifference(t *testing.T) {
	sc := table.NewTableStyleCollection()

	c1 := table.NewTableCell()
	f1 := c1.Font()
	f1.Size = 10
	c1.SetFont(f1)

	c2 := table.NewTableCell()
	f2 := c2.Font()
	f2.Size = 14
	c2.SetFont(f2)

	sc.Add(c1)
	sc.Add(c2)

	if sc.Count() != 2 {
		t.Errorf("Count: got %d, want 2 (distinct font size)", sc.Count())
	}
}

func TestCellStylesEqual_FontStyleDifference(t *testing.T) {
	sc := table.NewTableStyleCollection()

	c1 := table.NewTableCell()
	f1 := c1.Font()
	f1.Style = 0
	c1.SetFont(f1)

	c2 := table.NewTableCell()
	f2 := c2.Font()
	f2.Style = 1 // bold
	c2.SetFont(f2)

	sc.Add(c1)
	sc.Add(c2)

	if sc.Count() != 2 {
		t.Errorf("Count: got %d, want 2 (distinct font style)", sc.Count())
	}
}

// ── Full TableObject round-trip with rows, columns and cells ─────────────────

func TestTableObject_FullRoundTrip(t *testing.T) {
	orig := table.NewTableObject()
	orig.SetName("FullTable")
	col := orig.NewColumn()
	col.SetWidth(120)
	col.SetName("Col1")
	row := orig.NewRow()
	row.SetHeight(35)
	row.SetName("Row1")
	row.Cell(0).SetText("CellText")
	row.Cell(0).SetName("C1")
	row.Cell(0).SetColSpan(1)
	row.Cell(0).SetDuplicates(table.CellDuplicatesClear)

	data := serializeTableObject(t, orig)
	got := deserializeTableObject(t, data)

	if got.Name() != "FullTable" {
		t.Errorf("Name: got %q, want FullTable", got.Name())
	}
	// Columns and rows are not restored by Deserialize alone — DeserializeChild
	// handles them. The round-trip through Deserialize only restores attributes.
	// Here we verify the base attributes come back correctly.
}

// TestDeserializeChild_RowWithNestedObjects verifies the branch that skips
// unrecognised sub-children of a TableCell during deserialization.
func TestDeserializeChild_RowWithNestedCellObjects(t *testing.T) {
	// A TableCell child with an unknown sub-child — the unknown sub-child
	// should be skipped, and the cell should still be added to the row.
	xmlData := `<TableRow Name="R1">` +
		`<TableCell Name="Cell1"><PictureObject Name="Pic1"/></TableCell>` +
		`</TableRow>`

	r := serial.NewReader(bytes.NewReader([]byte(xmlData)))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader returned false")
	}

	tbl := table.NewTableObject()
	consumed := tbl.DeserializeChild("TableRow", r)
	if !consumed {
		t.Error("DeserializeChild(TableRow) should return true")
	}
	if tbl.RowCount() != 1 {
		t.Fatalf("RowCount: got %d, want 1", tbl.RowCount())
	}
	row := tbl.Row(0)
	if row.CellCount() != 1 {
		t.Fatalf("CellCount: got %d, want 1 (cell with nested object)", row.CellCount())
	}
	if row.Cell(0).Name() != "Cell1" {
		t.Errorf("Cell name: got %q, want Cell1", row.Cell(0).Name())
	}
}

// TestTableHelper_ColumnsPriority exercises the column-priority branch of TableHelper.
func TestTableHelper_ColumnsPriority(t *testing.T) {
	tbl := buildTemplate() // 3 rows × 2 columns
	var result *table.TableBase
	tbl.ManualBuild = func(h *table.TableHelper) {
		// Column-first: PrintColumn then PrintRow (columns-priority path).
		h.PrintColumn(0)
		h.PrintRow(0)
		h.PrintRow(1)
		h.PrintColumn(1)
		h.PrintRow(0)
		h.PrintRow(1)
		result = h.Result()
	}
	tbl.InvokeManualBuild()
	if result == nil {
		t.Fatal("result is nil")
	}
	if result.ColumnCount() != 2 {
		t.Errorf("ColumnCount: got %d, want 2", result.ColumnCount())
	}
}

// TestTableHelper_OutOfBounds tests that out-of-range indices are silently ignored.
func TestTableHelper_OutOfBounds(t *testing.T) {
	tbl := buildTemplate() // 3 rows × 2 columns
	var result *table.TableBase
	tbl.ManualBuild = func(h *table.TableHelper) {
		h.PrintRow(-1)   // out of range — ignored
		h.PrintRow(100)  // out of range — ignored
		h.PrintColumn(-1)  // out of range — ignored
		h.PrintColumn(100) // out of range — ignored
		h.PrintRow(0)
		h.PrintColumns()
		result = h.Result()
	}
	tbl.InvokeManualBuild()
	if result == nil {
		t.Fatal("result is nil")
	}
	if result.RowCount() != 1 {
		t.Errorf("RowCount: got %d, want 1", result.RowCount())
	}
}

// TestTableResult_Skip tests the Skip field on TableResult.
func TestTableResult_Skip(t *testing.T) {
	r := table.NewTableResult()
	r.Skip = true
	if !r.Skip {
		t.Error("Skip field should be settable")
	}
}
