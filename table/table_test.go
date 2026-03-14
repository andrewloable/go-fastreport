package table_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/table"
)

// ── TableColumn ───────────────────────────────────────────────────────────────

func TestNewTableColumn_Defaults(t *testing.T) {
	c := table.NewTableColumn()
	if c == nil {
		t.Fatal("NewTableColumn returned nil")
	}
	if c.Width() != 100 {
		t.Errorf("Width default = %v, want 100", c.Width())
	}
	if c.MinWidth() != 0 {
		t.Errorf("MinWidth default = %v, want 0", c.MinWidth())
	}
	if c.MaxWidth() != 500 {
		t.Errorf("MaxWidth default = %v, want 500", c.MaxWidth())
	}
	if c.AutoSize() {
		t.Error("AutoSize should default to false")
	}
	if c.PageBreak() {
		t.Error("PageBreak should default to false")
	}
	if c.KeepColumns() != 0 {
		t.Errorf("KeepColumns default = %d, want 0", c.KeepColumns())
	}
}

func TestTableColumn_SetFields(t *testing.T) {
	c := table.NewTableColumn()
	c.SetWidth(150)
	c.SetMinWidth(50)
	c.SetMaxWidth(300)
	c.SetAutoSize(true)
	c.SetPageBreak(true)
	c.SetKeepColumns(3)

	if c.Width() != 150 {
		t.Errorf("Width: got %v, want 150", c.Width())
	}
	if c.MinWidth() != 50 {
		t.Errorf("MinWidth: got %v, want 50", c.MinWidth())
	}
	if c.MaxWidth() != 300 {
		t.Errorf("MaxWidth: got %v, want 300", c.MaxWidth())
	}
	if !c.AutoSize() {
		t.Error("AutoSize should be true")
	}
	if !c.PageBreak() {
		t.Error("PageBreak should be true")
	}
	if c.KeepColumns() != 3 {
		t.Errorf("KeepColumns: got %d, want 3", c.KeepColumns())
	}
}

// ── TableRow ──────────────────────────────────────────────────────────────────

func TestNewTableRow_Defaults(t *testing.T) {
	r := table.NewTableRow()
	if r == nil {
		t.Fatal("NewTableRow returned nil")
	}
	if r.Height() != 30 {
		t.Errorf("Height default = %v, want 30", r.Height())
	}
	if r.MinHeight() != 0 {
		t.Errorf("MinHeight default = %v, want 0", r.MinHeight())
	}
	if r.AutoSize() {
		t.Error("AutoSize should default to false")
	}
	if r.CanBreak() {
		t.Error("CanBreak should default to false")
	}
	if r.CellCount() != 0 {
		t.Errorf("CellCount default = %d, want 0", r.CellCount())
	}
}

func TestTableRow_SetFields(t *testing.T) {
	r := table.NewTableRow()
	r.SetHeight(45)
	r.SetMinHeight(10)
	r.SetMaxHeight(200)
	r.SetAutoSize(true)
	r.SetCanBreak(true)
	r.SetPageBreak(true)
	r.SetKeepRows(2)

	if r.Height() != 45 {
		t.Errorf("Height: got %v, want 45", r.Height())
	}
	if r.MinHeight() != 10 {
		t.Errorf("MinHeight: got %v, want 10", r.MinHeight())
	}
	if r.MaxHeight() != 200 {
		t.Errorf("MaxHeight: got %v, want 200", r.MaxHeight())
	}
	if !r.AutoSize() {
		t.Error("AutoSize should be true")
	}
	if !r.CanBreak() {
		t.Error("CanBreak should be true")
	}
	if !r.PageBreak() {
		t.Error("PageBreak should be true")
	}
	if r.KeepRows() != 2 {
		t.Errorf("KeepRows: got %d, want 2", r.KeepRows())
	}
}

func TestTableRow_AddCell(t *testing.T) {
	r := table.NewTableRow()
	c1 := table.NewTableCell()
	c1.SetName("Cell1")
	c2 := table.NewTableCell()
	c2.SetName("Cell2")
	r.AddCell(c1)
	r.AddCell(c2)

	if r.CellCount() != 2 {
		t.Fatalf("CellCount: got %d, want 2", r.CellCount())
	}
	if r.Cell(0).Name() != "Cell1" {
		t.Errorf("Cell(0).Name: got %q, want Cell1", r.Cell(0).Name())
	}
	if r.Cell(1).Name() != "Cell2" {
		t.Errorf("Cell(1).Name: got %q, want Cell2", r.Cell(1).Name())
	}
	if r.Cell(99) != nil {
		t.Error("Cell(99) should be nil for out-of-range index")
	}
}

// ── TableCell ─────────────────────────────────────────────────────────────────

func TestNewTableCell_Defaults(t *testing.T) {
	c := table.NewTableCell()
	if c == nil {
		t.Fatal("NewTableCell returned nil")
	}
	if c.ColSpan() != 1 {
		t.Errorf("ColSpan default = %d, want 1", c.ColSpan())
	}
	if c.RowSpan() != 1 {
		t.Errorf("RowSpan default = %d, want 1", c.RowSpan())
	}
	if c.Duplicates() != table.CellDuplicatesShow {
		t.Errorf("Duplicates default = %d, want Show", c.Duplicates())
	}
	if c.ObjectCount() != 0 {
		t.Errorf("ObjectCount default = %d, want 0", c.ObjectCount())
	}
	if c.TypeName() != "TableCell" {
		t.Errorf("TypeName: got %q, want TableCell", c.TypeName())
	}
}

func TestTableCell_SetSpans(t *testing.T) {
	c := table.NewTableCell()
	c.SetColSpan(3)
	c.SetRowSpan(2)

	if c.ColSpan() != 3 {
		t.Errorf("ColSpan: got %d, want 3", c.ColSpan())
	}
	if c.RowSpan() != 2 {
		t.Errorf("RowSpan: got %d, want 2", c.RowSpan())
	}
}

func TestTableCell_SpanClampedToOne(t *testing.T) {
	c := table.NewTableCell()
	c.SetColSpan(0)
	c.SetRowSpan(-5)

	if c.ColSpan() != 1 {
		t.Errorf("ColSpan should clamp to 1, got %d", c.ColSpan())
	}
	if c.RowSpan() != 1 {
		t.Errorf("RowSpan should clamp to 1, got %d", c.RowSpan())
	}
}

func TestTableCell_Duplicates(t *testing.T) {
	c := table.NewTableCell()
	c.SetDuplicates(table.CellDuplicatesMerge)
	if c.Duplicates() != table.CellDuplicatesMerge {
		t.Errorf("Duplicates: got %d, want Merge", c.Duplicates())
	}
}

func TestTableCell_TextAndName(t *testing.T) {
	c := table.NewTableCell()
	c.SetName("Cell1")
	c.SetText("[Product.Name]")

	if c.Name() != "Cell1" {
		t.Errorf("Name: got %q, want Cell1", c.Name())
	}
	if c.Text() != "[Product.Name]" {
		t.Errorf("Text: got %q", c.Text())
	}
}

func TestTableCell_AddObject(t *testing.T) {
	c := table.NewTableCell()
	inner := table.NewTableCell() // use another cell as embedded object
	c.AddObject(inner)
	if c.ObjectCount() != 1 {
		t.Errorf("ObjectCount: got %d, want 1", c.ObjectCount())
	}
}

// ── TableBase / TableObject ───────────────────────────────────────────────────

func TestNewTableObject_Defaults(t *testing.T) {
	tbl := table.NewTableObject()
	if tbl == nil {
		t.Fatal("NewTableObject returned nil")
	}
	if tbl.RowCount() != 0 {
		t.Errorf("RowCount default = %d, want 0", tbl.RowCount())
	}
	if tbl.ColumnCount() != 0 {
		t.Errorf("ColumnCount default = %d, want 0", tbl.ColumnCount())
	}
	if tbl.Layout() != table.TableLayoutAcrossThenDown {
		t.Errorf("Layout default = %d, want AcrossThenDown", tbl.Layout())
	}
	if tbl.FixedRows() != 0 {
		t.Errorf("FixedRows default = %d, want 0", tbl.FixedRows())
	}
	if tbl.TypeName() != "TableObject" {
		t.Errorf("TypeName: got %q, want TableObject", tbl.TypeName())
	}
}

func TestTableObject_AddRowsAndColumns(t *testing.T) {
	tbl := table.NewTableObject()

	col1 := tbl.NewColumn()
	col1.SetWidth(120)
	col2 := tbl.NewColumn()
	col2.SetWidth(80)

	row1 := tbl.NewRow()
	row1.SetHeight(40)

	if tbl.ColumnCount() != 2 {
		t.Fatalf("ColumnCount: got %d, want 2", tbl.ColumnCount())
	}
	if tbl.RowCount() != 1 {
		t.Fatalf("RowCount: got %d, want 1", tbl.RowCount())
	}

	// NewRow should create cells for each column.
	if row1.CellCount() != 2 {
		t.Errorf("Row1.CellCount: got %d, want 2", row1.CellCount())
	}
}

func TestTableObject_CellAccess(t *testing.T) {
	tbl := table.NewTableObject()
	tbl.NewColumn()
	tbl.NewColumn()
	row := tbl.NewRow()

	row.Cell(0).SetText("R0C0")
	row.Cell(1).SetText("R0C1")

	if tbl.Cell(0, 0).Text() != "R0C0" {
		t.Errorf("Cell(0,0).Text: got %q, want R0C0", tbl.Cell(0, 0).Text())
	}
	if tbl.Cell(0, 1).Text() != "R0C1" {
		t.Errorf("Cell(0,1).Text: got %q, want R0C1", tbl.Cell(0, 1).Text())
	}
	if tbl.Cell(99, 0) != nil {
		t.Error("Cell(99,0) should be nil for out-of-range row")
	}
	if tbl.Cell(0, 99) != nil {
		t.Error("Cell(0,99) should be nil for out-of-range col")
	}
}

func TestTableObject_Properties(t *testing.T) {
	tbl := table.NewTableObject()
	tbl.SetFixedRows(2)
	tbl.SetFixedColumns(1)
	tbl.SetLayout(table.TableLayoutDownThenAcross)
	tbl.SetPrintOnParent(true)
	tbl.SetWrappedGap(10)
	tbl.SetRepeatHeaders(true)
	tbl.SetRepeatRowHeaders(true)
	tbl.SetRepeatColumnHeaders(true)
	tbl.SetAdjustSpannedCellsWidth(true)
	tbl.ManualBuildEvent = "OnManualBuild"

	if tbl.FixedRows() != 2 {
		t.Errorf("FixedRows: got %d, want 2", tbl.FixedRows())
	}
	if tbl.FixedColumns() != 1 {
		t.Errorf("FixedColumns: got %d, want 1", tbl.FixedColumns())
	}
	if tbl.Layout() != table.TableLayoutDownThenAcross {
		t.Errorf("Layout: got %d, want DownThenAcross", tbl.Layout())
	}
	if !tbl.PrintOnParent() {
		t.Error("PrintOnParent should be true")
	}
	if tbl.WrappedGap() != 10 {
		t.Errorf("WrappedGap: got %v, want 10", tbl.WrappedGap())
	}
	if !tbl.RepeatHeaders() {
		t.Error("RepeatHeaders should be true")
	}
	if !tbl.AdjustSpannedCellsWidth() {
		t.Error("AdjustSpannedCellsWidth should be true")
	}
	if tbl.ManualBuildEvent != "OnManualBuild" {
		t.Errorf("ManualBuildEvent: got %q", tbl.ManualBuildEvent)
	}
}

func TestTableObject_AddColumn_AutoCells(t *testing.T) {
	// Adding a column AFTER rows exist should create cells in existing rows.
	tbl := table.NewTableObject()
	row := tbl.NewRow()
	if row.CellCount() != 0 {
		t.Errorf("expected 0 cells before any columns, got %d", row.CellCount())
	}
	tbl.AddColumn(table.NewTableColumn())
	if row.CellCount() != 1 {
		t.Errorf("expected 1 cell after AddColumn, got %d", row.CellCount())
	}
}

// ── Serialization round-trip ──────────────────────────────────────────────────

func serializeTable(t *testing.T, tbl *table.TableObject) []byte {
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

func TestTableObject_Serialize_ContainsKeyAttributes(t *testing.T) {
	tbl := table.NewTableObject()
	tbl.SetName("Table1")
	tbl.SetFixedRows(2)
	tbl.SetLayout(table.TableLayoutWrapped)
	tbl.SetRepeatHeaders(true)
	tbl.ManualBuildEvent = "BuildTable"

	col := tbl.NewColumn()
	col.SetWidth(200)
	row := tbl.NewRow()
	row.SetHeight(50)
	row.Cell(0).SetText("Header")

	data := serializeTable(t, tbl)
	xml := string(data)

	for _, want := range []string{
		`TableObject`,
		`Name="Table1"`,
		`FixedRows="2"`,
		`Layout="2"`, // TableLayoutWrapped = 2
		`RepeatHeaders="true"`,
		`ManualBuildEvent="BuildTable"`,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("XML missing %q in:\n%s", want, xml)
		}
	}
}

func TestTableCell_RoundTrip(t *testing.T) {
	orig := table.NewTableCell()
	orig.SetName("C1")
	orig.SetText("Sales Total")
	orig.SetColSpan(2)
	orig.SetRowSpan(3)
	orig.SetDuplicates(table.CellDuplicatesMergeNonEmpty)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TableCell", orig); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TableCell" {
		t.Fatalf("got typeName=%q ok=%v", typeName, ok)
	}
	got := table.NewTableCell()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if got.Name() != "C1" {
		t.Errorf("Name: got %q, want C1", got.Name())
	}
	if got.Text() != "Sales Total" {
		t.Errorf("Text: got %q, want 'Sales Total'", got.Text())
	}
	if got.ColSpan() != 2 {
		t.Errorf("ColSpan: got %d, want 2", got.ColSpan())
	}
	if got.RowSpan() != 3 {
		t.Errorf("RowSpan: got %d, want 3", got.RowSpan())
	}
	if got.Duplicates() != table.CellDuplicatesMergeNonEmpty {
		t.Errorf("Duplicates: got %d, want MergeNonEmpty", got.Duplicates())
	}
}

func TestTableColumn_RoundTrip(t *testing.T) {
	orig := table.NewTableColumn()
	orig.SetName("Col1")
	orig.SetWidth(150)
	orig.SetMinWidth(50)
	orig.SetAutoSize(true)
	orig.SetKeepColumns(2)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TableColumn", orig); err != nil {
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
	got := table.NewTableColumn()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if got.Name() != "Col1" {
		t.Errorf("Name: got %q, want Col1", got.Name())
	}
	if got.Width() != 150 {
		t.Errorf("Width: got %v, want 150", got.Width())
	}
	if got.MinWidth() != 50 {
		t.Errorf("MinWidth: got %v, want 50", got.MinWidth())
	}
	if !got.AutoSize() {
		t.Error("AutoSize should be true")
	}
	if got.KeepColumns() != 2 {
		t.Errorf("KeepColumns: got %d, want 2", got.KeepColumns())
	}
}

func TestTableRow_RoundTrip(t *testing.T) {
	orig := table.NewTableRow()
	orig.SetName("Row1")
	orig.SetHeight(60)
	orig.SetMinHeight(20)
	orig.SetCanBreak(true)
	orig.SetKeepRows(3)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TableRow", orig); err != nil {
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
	got := table.NewTableRow()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if got.Name() != "Row1" {
		t.Errorf("Name: got %q, want Row1", got.Name())
	}
	if got.Height() != 60 {
		t.Errorf("Height: got %v, want 60", got.Height())
	}
	if got.MinHeight() != 20 {
		t.Errorf("MinHeight: got %v, want 20", got.MinHeight())
	}
	if !got.CanBreak() {
		t.Error("CanBreak should be true")
	}
	if got.KeepRows() != 3 {
		t.Errorf("KeepRows: got %d, want 3", got.KeepRows())
	}
}

func TestTableLayout_Constants(t *testing.T) {
	layouts := []table.TableLayout{
		table.TableLayoutAcrossThenDown,
		table.TableLayoutDownThenAcross,
		table.TableLayoutWrapped,
	}
	seen := map[table.TableLayout]bool{}
	for _, l := range layouts {
		if seen[l] {
			t.Errorf("duplicate TableLayout %d", l)
		}
		seen[l] = true
	}
}

func TestCellDuplicates_Constants(t *testing.T) {
	dups := []table.CellDuplicates{
		table.CellDuplicatesShow,
		table.CellDuplicatesClear,
		table.CellDuplicatesMerge,
		table.CellDuplicatesMergeNonEmpty,
	}
	seen := map[table.CellDuplicates]bool{}
	for _, d := range dups {
		if seen[d] {
			t.Errorf("duplicate CellDuplicates %d", d)
		}
		seen[d] = true
	}
}
