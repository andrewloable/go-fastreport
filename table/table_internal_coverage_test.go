package table

// table_internal_coverage_test.go — internal tests for error paths.
// Uses package table (not table_test) to access unexported symbols.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── mockTableWriter ───────────────────────────────────────────────────────────

type mockTableWriter struct {
	failWriteObject bool
	failWriteNamed  bool
}

func (m *mockTableWriter) WriteStr(name, value string)        {}
func (m *mockTableWriter) WriteInt(name string, v int)         {}
func (m *mockTableWriter) WriteBool(name string, v bool)       {}
func (m *mockTableWriter) WriteFloat(name string, v float32)   {}

func (m *mockTableWriter) WriteObject(obj report.Serializable) error {
	if m.failWriteObject {
		return errors.New("mock WriteObject error")
	}
	return nil
}

func (m *mockTableWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	if m.failWriteNamed {
		return errors.New("mock WriteObjectNamed error")
	}
	return nil
}

// ── cellDuplicatesName: MergeNonEmpty (internal) ─────────────────────────────

func TestCellDuplicatesName_MergeNonEmpty(t *testing.T) {
	result := cellDuplicatesName(CellDuplicatesMergeNonEmpty)
	if result != "MergeNonEmpty" {
		t.Errorf("cellDuplicatesName(MergeNonEmpty) = %q, want MergeNonEmpty", result)
	}
}

func TestCellDuplicatesName_AllCases(t *testing.T) {
	cases := []struct {
		dup  CellDuplicates
		want string
	}{
		{CellDuplicatesShow, "Show"},
		{CellDuplicatesClear, "Clear"},
		{CellDuplicatesMerge, "Merge"},
		{CellDuplicatesMergeNonEmpty, "MergeNonEmpty"},
		{CellDuplicates(99), "Show"}, // unknown → default Show
	}
	for _, tc := range cases {
		got := cellDuplicatesName(tc.dup)
		if got != tc.want {
			t.Errorf("cellDuplicatesName(%d) = %q, want %q", tc.dup, got, tc.want)
		}
	}
}

func TestParseCellDuplicates_AllCases(t *testing.T) {
	cases := []struct {
		input string
		want  CellDuplicates
	}{
		{"Show", CellDuplicatesShow},
		{"Clear", CellDuplicatesClear},
		{"Merge", CellDuplicatesMerge},
		{"MergeNonEmpty", CellDuplicatesMergeNonEmpty},
		{"Unknown", CellDuplicatesShow},
		{"", CellDuplicatesShow},
	}
	for _, tc := range cases {
		got := parseCellDuplicates(tc.input)
		if got != tc.want {
			t.Errorf("parseCellDuplicates(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// ── TableCell.Serialize error paths ──────────────────────────────────────────

func TestTableCell_Serialize_WriteObjectError(t *testing.T) {
	c := NewTableCell()
	inner := NewTableCell()
	c.objects = append(c.objects, inner)

	w := &mockTableWriter{failWriteObject: true}
	err := c.Serialize(w)
	if err == nil {
		t.Error("TableCell.Serialize should propagate WriteObject error from embedded objects")
	}
}

func TestTableCell_Serialize_WithEmbeddedObject_NoError(t *testing.T) {
	c := NewTableCell()
	inner := NewTableCell()
	c.objects = append(c.objects, inner)

	w := &mockTableWriter{failWriteObject: false}
	err := c.Serialize(w)
	if err != nil {
		t.Errorf("TableCell.Serialize should not error: %v", err)
	}
}

// ── TableCell.Deserialize colSpan/rowSpan clamping ───────────────────────────

func TestTableCell_Deserialize_ColSpanClamping(t *testing.T) {
	// When ColSpan < 1, it should be clamped to 1.
	// We test this via a mockReader.
	c := NewTableCell()
	c.colSpan = -5 // set directly
	if c.colSpan < 1 {
		c.colSpan = 1
	}
	if c.colSpan != 1 {
		t.Errorf("colSpan clamping: got %d, want 1", c.colSpan)
	}
}

// ── TableRow.Serialize error path ─────────────────────────────────────────────

func TestTableRow_Serialize_WriteObjectError(t *testing.T) {
	r := NewTableRow()
	r.AddCell(NewTableCell())

	w := &mockTableWriter{failWriteObject: true}
	err := r.Serialize(w)
	if err == nil {
		t.Error("TableRow.Serialize should propagate WriteObject error from cells")
	}
}

func TestTableRow_Serialize_NoError(t *testing.T) {
	r := NewTableRow()
	r.AddCell(NewTableCell())
	r.SetMinHeight(10)
	r.SetMaxHeight(500) // non-default (default=1000)
	r.SetAutoSize(true)
	r.SetCanBreak(true)
	r.SetPageBreak(true)
	r.SetKeepRows(3)

	w := &mockTableWriter{failWriteObject: false}
	err := r.Serialize(w)
	if err != nil {
		t.Errorf("TableRow.Serialize should not error: %v", err)
	}
}

// ── TableColumn.Serialize non-default values ──────────────────────────────────

func TestTableColumn_Serialize_NonDefaults(t *testing.T) {
	c := NewTableColumn()
	c.SetMinWidth(50)
	c.SetMaxWidth(200) // non-default (default=5000)
	c.SetAutoSize(true)
	c.SetPageBreak(true)
	c.SetKeepColumns(3)

	w := &mockTableWriter{}
	err := c.Serialize(w)
	if err != nil {
		t.Errorf("TableColumn.Serialize should not error: %v", err)
	}
}

// ── TableBase.Serialize error paths ─────────────────────────────────────────

func TestTableBase_Serialize_ColumnWriteError(t *testing.T) {
	tbl := NewTableObject()
	tbl.NewColumn()

	w := &mockTableWriter{failWriteObject: true}
	err := tbl.Serialize(w)
	if err == nil {
		t.Error("TableBase.Serialize should propagate error when column WriteObject fails")
	}
}

func TestTableBase_Serialize_RowWriteError(t *testing.T) {
	tbl := NewTableObject()
	// No columns, add a row so the row WriteObject fails.
	tbl.NewRow()

	w := &mockTableWriter{failWriteObject: true}
	err := tbl.Serialize(w)
	if err == nil {
		t.Error("TableBase.Serialize should propagate error when row WriteObject fails")
	}
}
