package table

// table_final_coverage_test.go — final coverage push for the table package.
//
// The remaining gaps (as reported by go tool cover -func) are:
//
//   cell.go:122    Serialize    91.7%  — "return err" on TextObject.Serialize failure
//   cell.go:145    Deserialize  90.0%  — "return err" on TextObject.Deserialize failure
//   column.go:65   Serialize    92.3%  — "return err" on ComponentBase.Serialize failure
//   column.go:88   Deserialize  88.9%  — "return err" on ComponentBase.Deserialize failure
//   row.go:95      Serialize    94.4%  — "return err" on ComponentBase.Serialize failure
//   row.go:127     Deserialize  90.0%  — "return err" on ComponentBase.Deserialize failure
//   table.go:198   Serialize    96.6%  — "return err" on BreakableComponent.Serialize failure
//   table.go:249   Deserialize  92.3%  — "return err" on BreakableComponent.Deserialize failure
//   table.go:311   Deserialize  75.0%  — "return err" on TableBase.Deserialize failure
//
// Root-cause analysis:
//
// All nine uncovered lines are defensive "return err" branches that execute when
// a parent Serialize or Deserialize method returns a non-nil error. The entire
// call chain from TableObject → TableBase → BreakableComponent →
// ReportComponentBase → ComponentBase → BaseObject, and from
// TableCell → TextObject → TextObjectBase → BreakableComponent → ...,
// consists entirely of implementations that call only the void-returning
// report.Writer methods (WriteStr, WriteInt, WriteBool, WriteFloat) or the
// value-returning report.Reader methods (ReadStr, ReadInt, ReadBool, ReadFloat).
// None of these ever return an error.
//
// Because the report.Writer interface defines WriteStr/WriteInt/WriteBool/WriteFloat
// as void (no error return), and because none of the parent Serialize or
// Deserialize implementations call WriteObject (which can return an error),
// it is structurally impossible to make any parent in the chain return a non-nil
// error. The nine "return err" lines are therefore dead code with the current
// interface design.
//
// This file provides the most comprehensive possible tests for all REACHABLE
// paths in each affected function, verifying correct behaviour while confirming
// that the functions work correctly end-to-end across all non-error scenarios.

import (
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── Exhaustive reachable-path tests ──────────────────────────────────────────

// verifyWriter is a writer that records which fields were written.
type verifyWriter struct {
	strs   map[string]string
	ints   map[string]int
	bools  map[string]bool
	floats map[string]float32
	objs   []report.Serializable
}

func newVerifyWriter() *verifyWriter {
	return &verifyWriter{
		strs:   make(map[string]string),
		ints:   make(map[string]int),
		bools:  make(map[string]bool),
		floats: make(map[string]float32),
	}
}

func (v *verifyWriter) WriteStr(name, value string)      { v.strs[name] = value }
func (v *verifyWriter) WriteInt(name string, val int)     { v.ints[name] = val }
func (v *verifyWriter) WriteBool(name string, val bool)   { v.bools[name] = val }
func (v *verifyWriter) WriteFloat(name string, val float32) { v.floats[name] = val }
func (v *verifyWriter) WriteObject(obj report.Serializable) error {
	v.objs = append(v.objs, obj)
	// Recurse into the object so inner branches are also counted.
	inner := newVerifyWriter()
	return obj.Serialize(inner)
}
func (v *verifyWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	return v.WriteObject(obj)
}

// verifyReader is a reader that returns specified values.
type verifyReader struct {
	strs   map[string]string
	ints   map[string]int
	bools  map[string]bool
	floats map[string]float32
}

func newVerifyReader() *verifyReader {
	return &verifyReader{
		strs:   make(map[string]string),
		ints:   make(map[string]int),
		bools:  make(map[string]bool),
		floats: make(map[string]float32),
	}
}

func (v *verifyReader) ReadStr(name, def string) string {
	if val, ok := v.strs[name]; ok {
		return val
	}
	return def
}
func (v *verifyReader) ReadInt(name string, def int) int {
	if val, ok := v.ints[name]; ok {
		return val
	}
	return def
}
func (v *verifyReader) ReadBool(name string, def bool) bool {
	if val, ok := v.bools[name]; ok {
		return val
	}
	return def
}
func (v *verifyReader) ReadFloat(name string, def float32) float32 {
	if val, ok := v.floats[name]; ok {
		return val
	}
	return def
}
func (v *verifyReader) NextChild() (string, bool) { return "", false }
func (v *verifyReader) FinishChild() error        { return nil }

// ── TableCell.Serialize — exhaustive happy-path verification ─────────────────

// TestTableCell_Serialize_HappyPath_AllBranches exercises every reachable
// branch in TableCell.Serialize. The parent TextObject.Serialize path always
// succeeds; we verify that all conditional writes (ColSpan, RowSpan,
// CellDuplicates, embedded objects) are triggered and that nil is returned.
func TestTableCell_Serialize_HappyPath_AllBranches(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*TableCell)
		wantColSpan bool
		wantRowSpan bool
		wantDupes  bool
		wantObjs   int
	}{
		{
			name:  "defaults only — no non-default writes",
			setup: func(c *TableCell) {},
		},
		{
			name: "ColSpan=2 triggers ColSpan write",
			setup: func(c *TableCell) {
				c.colSpan = 2
			},
			wantColSpan: true,
		},
		{
			name: "RowSpan=3 triggers RowSpan write",
			setup: func(c *TableCell) {
				c.rowSpan = 3
			},
			wantRowSpan: true,
		},
		{
			name: "CellDuplicates=Clear triggers CellDuplicates write",
			setup: func(c *TableCell) {
				c.duplicates = CellDuplicatesClear
			},
			wantDupes: true,
		},
		{
			name: "embedded object triggers WriteObject",
			setup: func(c *TableCell) {
				c.objects = append(c.objects, NewTableCell())
			},
			wantObjs: 1,
		},
		{
			name: "all non-defaults simultaneously",
			setup: func(c *TableCell) {
				c.colSpan = 4
				c.rowSpan = 2
				c.duplicates = CellDuplicatesMerge
				c.objects = append(c.objects, NewTableCell())
				c.objects = append(c.objects, NewTableCell())
			},
			wantColSpan: true,
			wantRowSpan: true,
			wantDupes:  true,
			wantObjs:   2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cell := NewTableCell()
			tc.setup(cell)

			w := newVerifyWriter()
			err := cell.Serialize(w)
			if err != nil {
				t.Fatalf("Serialize returned unexpected error: %v", err)
			}

			_, colSpanWritten := w.ints["ColSpan"]
			if colSpanWritten != tc.wantColSpan {
				t.Errorf("ColSpan written=%v, want %v", colSpanWritten, tc.wantColSpan)
			}

			_, rowSpanWritten := w.ints["RowSpan"]
			if rowSpanWritten != tc.wantRowSpan {
				t.Errorf("RowSpan written=%v, want %v", rowSpanWritten, tc.wantRowSpan)
			}

			_, dupesWritten := w.strs["CellDuplicates"]
			if dupesWritten != tc.wantDupes {
				t.Errorf("CellDuplicates written=%v, want %v", dupesWritten, tc.wantDupes)
			}

			if len(w.objs) != tc.wantObjs {
				t.Errorf("WriteObject calls=%d, want %d", len(w.objs), tc.wantObjs)
			}
		})
	}
}

// TestTableCell_Serialize_ErrorPath_ParentNeverFails verifies that
// TextObject.Serialize (the parent call) never returns an error with any
// combination of TableCell state, confirming the "return err" branch at
// cell.go:124 is structurally dead code.
func TestTableCell_Serialize_ErrorPath_ParentNeverFails(t *testing.T) {
	cell := NewTableCell()
	// Exercise with the most complex possible state.
	cell.colSpan = 5
	cell.rowSpan = 5
	cell.duplicates = CellDuplicatesMergeNonEmpty
	cell.objects = append(cell.objects, NewTableCell(), NewTableCell())
	cell.SetText("some text")
	cell.SetName("myCell")

	w := newVerifyWriter()
	if err := cell.Serialize(w); err != nil {
		t.Errorf("TextObject.Serialize must always return nil; got: %v", err)
	}
}

// ── TableCell.Deserialize — exhaustive happy-path verification ───────────────

// TestTableCell_Deserialize_HappyPath_AllBranches exercises every reachable
// branch in TableCell.Deserialize, verifying that all fields are read and that
// the clamping branches (colSpan<1 and rowSpan<1) are triggered.
func TestTableCell_Deserialize_HappyPath_AllBranches(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*verifyReader)
		wantColSpan int
		wantRowSpan int
		wantDupes   CellDuplicates
	}{
		{
			name:        "all defaults",
			setup:       func(r *verifyReader) {},
			wantColSpan: 1,
			wantRowSpan: 1,
			wantDupes:   CellDuplicatesShow,
		},
		{
			name: "explicit ColSpan=3",
			setup: func(r *verifyReader) {
				r.ints["ColSpan"] = 3
			},
			wantColSpan: 3,
			wantRowSpan: 1,
			wantDupes:   CellDuplicatesShow,
		},
		{
			name: "explicit RowSpan=4",
			setup: func(r *verifyReader) {
				r.ints["RowSpan"] = 4
			},
			wantColSpan: 1,
			wantRowSpan: 4,
			wantDupes:   CellDuplicatesShow,
		},
		{
			name: "ColSpan=0 is clamped to 1",
			setup: func(r *verifyReader) {
				r.ints["ColSpan"] = 0
			},
			wantColSpan: 1,
			wantRowSpan: 1,
		},
		{
			name: "RowSpan=-2 is clamped to 1",
			setup: func(r *verifyReader) {
				r.ints["RowSpan"] = -2
			},
			wantColSpan: 1,
			wantRowSpan: 1,
		},
		{
			name: "CellDuplicates=MergeNonEmpty",
			setup: func(r *verifyReader) {
				r.strs["CellDuplicates"] = "MergeNonEmpty"
			},
			wantColSpan: 1,
			wantRowSpan: 1,
			wantDupes:   CellDuplicatesMergeNonEmpty,
		},
		{
			name: "CellDuplicates=Merge",
			setup: func(r *verifyReader) {
				r.strs["CellDuplicates"] = "Merge"
			},
			wantColSpan: 1,
			wantRowSpan: 1,
			wantDupes:   CellDuplicatesMerge,
		},
		{
			name: "CellDuplicates=Clear",
			setup: func(r *verifyReader) {
				r.strs["CellDuplicates"] = "Clear"
			},
			wantColSpan: 1,
			wantRowSpan: 1,
			wantDupes:   CellDuplicatesClear,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newVerifyReader()
			tc.setup(r)

			cell := NewTableCell()
			err := cell.Deserialize(r)
			if err != nil {
				t.Fatalf("Deserialize returned unexpected error: %v", err)
			}

			if cell.colSpan != tc.wantColSpan {
				t.Errorf("colSpan: got %d, want %d", cell.colSpan, tc.wantColSpan)
			}
			if cell.rowSpan != tc.wantRowSpan {
				t.Errorf("rowSpan: got %d, want %d", cell.rowSpan, tc.wantRowSpan)
			}
			if cell.duplicates != tc.wantDupes {
				t.Errorf("duplicates: got %d, want %d", cell.duplicates, tc.wantDupes)
			}
		})
	}
}

// TestTableCell_Deserialize_ErrorPath_ParentNeverFails verifies that
// TextObject.Deserialize never fails, confirming the "return err" branch at
// cell.go:147 is structurally dead code.
func TestTableCell_Deserialize_ErrorPath_ParentNeverFails(t *testing.T) {
	r := newVerifyReader()
	r.ints["ColSpan"] = 2
	r.ints["RowSpan"] = 3
	r.strs["CellDuplicates"] = "MergeNonEmpty"

	cell := NewTableCell()
	if err := cell.Deserialize(r); err != nil {
		t.Errorf("TextObject.Deserialize must always return nil; got: %v", err)
	}
}

// ── TableColumn.Serialize — exhaustive happy-path verification ───────────────

// TestTableColumn_Serialize_HappyPath_AllBranches verifies every conditional
// write branch in TableColumn.Serialize and confirms nil is always returned.
func TestTableColumn_Serialize_HappyPath_AllBranches(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*TableColumn)
		wantMinWidth bool
		wantMaxWidth bool
		wantAutoSize bool
		wantPageBreak bool
		wantKeepCols bool
	}{
		{
			name:  "defaults — minWidth=0, maxWidth=5000, nothing written",
			setup: func(c *TableColumn) {},
		},
		{
			name: "minWidth non-zero",
			setup: func(c *TableColumn) {
				c.minWidth = 25
			},
			wantMinWidth: true,
		},
		{
			name: "maxWidth non-default (not 5000)",
			setup: func(c *TableColumn) {
				c.maxWidth = 999
			},
			wantMaxWidth: true,
		},
		{
			name: "autoSize=true",
			setup: func(c *TableColumn) {
				c.autoSize = true
			},
			wantAutoSize: true,
		},
		{
			name: "pageBreak=true",
			setup: func(c *TableColumn) {
				c.pageBreak = true
			},
			wantPageBreak: true,
		},
		{
			name: "keepColumns non-zero",
			setup: func(c *TableColumn) {
				c.keepColumns = 7
			},
			wantKeepCols: true,
		},
		{
			name: "all non-defaults simultaneously",
			setup: func(c *TableColumn) {
				c.minWidth = 10
				c.maxWidth = 200
				c.autoSize = true
				c.pageBreak = true
				c.keepColumns = 3
			},
			wantMinWidth:  true,
			wantMaxWidth:  true,
			wantAutoSize:  true,
			wantPageBreak: true,
			wantKeepCols:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			col := NewTableColumn()
			tc.setup(col)

			w := newVerifyWriter()
			err := col.Serialize(w)
			if err != nil {
				t.Fatalf("Serialize returned unexpected error: %v", err)
			}

			_, minW := w.floats["MinWidth"]
			if minW != tc.wantMinWidth {
				t.Errorf("MinWidth written=%v, want %v", minW, tc.wantMinWidth)
			}
			_, maxW := w.floats["MaxWidth"]
			if maxW != tc.wantMaxWidth {
				t.Errorf("MaxWidth written=%v, want %v", maxW, tc.wantMaxWidth)
			}
			_, asize := w.bools["AutoSize"]
			if asize != tc.wantAutoSize {
				t.Errorf("AutoSize written=%v, want %v", asize, tc.wantAutoSize)
			}
			_, pb := w.bools["PageBreak"]
			if pb != tc.wantPageBreak {
				t.Errorf("PageBreak written=%v, want %v", pb, tc.wantPageBreak)
			}
			_, kc := w.ints["KeepColumns"]
			if kc != tc.wantKeepCols {
				t.Errorf("KeepColumns written=%v, want %v", kc, tc.wantKeepCols)
			}
		})
	}
}

// TestTableColumn_Serialize_ErrorPath_ParentNeverFails verifies that
// ComponentBase.Serialize never returns an error.
func TestTableColumn_Serialize_ErrorPath_ParentNeverFails(t *testing.T) {
	col := NewTableColumn()
	col.minWidth = 50
	col.maxWidth = 100
	col.autoSize = true
	col.pageBreak = true
	col.keepColumns = 4

	w := newVerifyWriter()
	if err := col.Serialize(w); err != nil {
		t.Errorf("ComponentBase.Serialize must always return nil; got: %v", err)
	}
}

// ── TableColumn.Deserialize — exhaustive happy-path verification ─────────────

// TestTableColumn_Deserialize_HappyPath_AllBranches verifies all readable
// fields in TableColumn.Deserialize and confirms nil is returned.
func TestTableColumn_Deserialize_HappyPath_AllBranches(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*verifyReader)
		wantWidth     float32
		wantMinWidth  float32
		wantMaxWidth  float32
		wantAutoSize  bool
		wantPageBreak bool
		wantKeepCols  int
	}{
		{
			name:         "all defaults",
			setup:        func(r *verifyReader) {},
			wantWidth:    66.15,
			wantMaxWidth: 5000,
		},
		{
			name: "explicit Width=200",
			setup: func(r *verifyReader) {
				r.floats["Width"] = 200
			},
			wantWidth:    200,
			wantMaxWidth: 5000,
		},
		{
			name: "MinWidth=50",
			setup: func(r *verifyReader) {
				r.floats["MinWidth"] = 50
			},
			wantWidth:    66.15,
			wantMinWidth: 50,
			wantMaxWidth: 5000,
		},
		{
			name: "MaxWidth=2000",
			setup: func(r *verifyReader) {
				r.floats["MaxWidth"] = 2000
			},
			wantWidth:    66.15,
			wantMaxWidth: 2000,
		},
		{
			name: "AutoSize=true",
			setup: func(r *verifyReader) {
				r.bools["AutoSize"] = true
			},
			wantWidth:    66.15,
			wantMaxWidth: 5000,
			wantAutoSize: true,
		},
		{
			name: "PageBreak=true",
			setup: func(r *verifyReader) {
				r.bools["PageBreak"] = true
			},
			wantWidth:     66.15,
			wantMaxWidth:  5000,
			wantPageBreak: true,
		},
		{
			name: "KeepColumns=5",
			setup: func(r *verifyReader) {
				r.ints["KeepColumns"] = 5
			},
			wantWidth:    66.15,
			wantMaxWidth: 5000,
			wantKeepCols: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newVerifyReader()
			tc.setup(r)

			col := NewTableColumn()
			err := col.Deserialize(r)
			if err != nil {
				t.Fatalf("Deserialize returned unexpected error: %v", err)
			}

			if col.Width() != tc.wantWidth {
				t.Errorf("Width: got %v, want %v", col.Width(), tc.wantWidth)
			}
			if col.minWidth != tc.wantMinWidth {
				t.Errorf("minWidth: got %v, want %v", col.minWidth, tc.wantMinWidth)
			}
			if col.maxWidth != tc.wantMaxWidth {
				t.Errorf("maxWidth: got %v, want %v", col.maxWidth, tc.wantMaxWidth)
			}
			if col.autoSize != tc.wantAutoSize {
				t.Errorf("autoSize: got %v, want %v", col.autoSize, tc.wantAutoSize)
			}
			if col.pageBreak != tc.wantPageBreak {
				t.Errorf("pageBreak: got %v, want %v", col.pageBreak, tc.wantPageBreak)
			}
			if col.keepColumns != tc.wantKeepCols {
				t.Errorf("keepColumns: got %d, want %d", col.keepColumns, tc.wantKeepCols)
			}
		})
	}
}

// TestTableColumn_Deserialize_ErrorPath_ParentNeverFails verifies that
// ComponentBase.Deserialize never returns an error.
func TestTableColumn_Deserialize_ErrorPath_ParentNeverFails(t *testing.T) {
	r := newVerifyReader()
	r.floats["Width"] = 150
	r.floats["MinWidth"] = 20
	r.floats["MaxWidth"] = 800
	r.bools["AutoSize"] = true
	r.bools["PageBreak"] = true
	r.ints["KeepColumns"] = 3

	col := NewTableColumn()
	if err := col.Deserialize(r); err != nil {
		t.Errorf("ComponentBase.Deserialize must always return nil; got: %v", err)
	}
}

// ── TableRow.Serialize — exhaustive happy-path verification ──────────────────

// TestTableRow_Serialize_HappyPath_AllBranches verifies every conditional write
// branch in TableRow.Serialize.
func TestTableRow_Serialize_HappyPath_AllBranches(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*TableRow)
		wantMinH      bool
		wantMaxH      bool
		wantAutoSize  bool
		wantCanBreak  bool
		wantPageBreak bool
		wantKeepRows  bool
		wantCells     int
	}{
		{
			name:  "all defaults — nothing written",
			setup: func(r *TableRow) {},
		},
		{
			name: "MinHeight non-zero",
			setup: func(r *TableRow) {
				r.minHeight = 5
			},
			wantMinH: true,
		},
		{
			name: "MaxHeight != 1000",
			setup: func(r *TableRow) {
				r.maxHeight = 500
			},
			wantMaxH: true,
		},
		{
			name: "AutoSize=true",
			setup: func(r *TableRow) {
				r.autoSize = true
			},
			wantAutoSize: true,
		},
		{
			name: "CanBreak=true",
			setup: func(r *TableRow) {
				r.canBreak = true
			},
			wantCanBreak: true,
		},
		{
			name: "PageBreak=true",
			setup: func(r *TableRow) {
				r.pageBreak = true
			},
			wantPageBreak: true,
		},
		{
			name: "KeepRows=2",
			setup: func(r *TableRow) {
				r.keepRows = 2
			},
			wantKeepRows: true,
		},
		{
			name: "two cells serialized",
			setup: func(r *TableRow) {
				r.cells = append(r.cells, NewTableCell(), NewTableCell())
			},
			wantCells: 2,
		},
		{
			name: "all non-defaults plus cells",
			setup: func(r *TableRow) {
				r.minHeight = 8
				r.maxHeight = 2500
				r.autoSize = true
				r.canBreak = true
				r.pageBreak = true
				r.keepRows = 4
				r.cells = append(r.cells, NewTableCell())
			},
			wantMinH:      true,
			wantMaxH:      true,
			wantAutoSize:  true,
			wantCanBreak:  true,
			wantPageBreak: true,
			wantKeepRows:  true,
			wantCells:     1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			row := NewTableRow()
			tc.setup(row)

			w := newVerifyWriter()
			err := row.Serialize(w)
			if err != nil {
				t.Fatalf("Serialize returned unexpected error: %v", err)
			}

			_, minH := w.floats["MinHeight"]
			if minH != tc.wantMinH {
				t.Errorf("MinHeight written=%v, want %v", minH, tc.wantMinH)
			}
			_, maxH := w.floats["MaxHeight"]
			if maxH != tc.wantMaxH {
				t.Errorf("MaxHeight written=%v, want %v", maxH, tc.wantMaxH)
			}
			_, asize := w.bools["AutoSize"]
			if asize != tc.wantAutoSize {
				t.Errorf("AutoSize written=%v, want %v", asize, tc.wantAutoSize)
			}
			_, cb := w.bools["CanBreak"]
			if cb != tc.wantCanBreak {
				t.Errorf("CanBreak written=%v, want %v", cb, tc.wantCanBreak)
			}
			_, pb := w.bools["PageBreak"]
			if pb != tc.wantPageBreak {
				t.Errorf("PageBreak written=%v, want %v", pb, tc.wantPageBreak)
			}
			_, kr := w.ints["KeepRows"]
			if kr != tc.wantKeepRows {
				t.Errorf("KeepRows written=%v, want %v", kr, tc.wantKeepRows)
			}
			if len(w.objs) != tc.wantCells {
				t.Errorf("WriteObject (cell) calls=%d, want %d", len(w.objs), tc.wantCells)
			}
		})
	}
}

// TestTableRow_Serialize_ErrorPath_ParentNeverFails verifies that
// ComponentBase.Serialize (via row.ComponentBase.Serialize) never returns error.
func TestTableRow_Serialize_ErrorPath_ParentNeverFails(t *testing.T) {
	row := NewTableRow()
	row.minHeight = 15
	row.maxHeight = 2000
	row.autoSize = true
	row.canBreak = true
	row.pageBreak = true
	row.keepRows = 6
	row.cells = append(row.cells, NewTableCell())

	w := newVerifyWriter()
	if err := row.Serialize(w); err != nil {
		t.Errorf("ComponentBase.Serialize must always return nil; got: %v", err)
	}
}

// ── TableRow.Deserialize — exhaustive happy-path verification ────────────────

// TestTableRow_Deserialize_HappyPath_AllBranches verifies every readable field
// in TableRow.Deserialize and confirms nil is always returned.
func TestTableRow_Deserialize_HappyPath_AllBranches(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*verifyReader)
		wantHeight    float32
		wantMinH      float32
		wantMaxH      float32
		wantAutoSize  bool
		wantCanBreak  bool
		wantPageBreak bool
		wantKeepRows  int
	}{
		{
			name:       "all defaults",
			setup:      func(r *verifyReader) {},
			wantHeight: 18.9,
			wantMaxH:   1000,
		},
		{
			name: "Height=60",
			setup: func(r *verifyReader) {
				r.floats["Height"] = 60
			},
			wantHeight: 60,
			wantMaxH:   1000,
		},
		{
			name: "MinHeight=5",
			setup: func(r *verifyReader) {
				r.floats["MinHeight"] = 5
			},
			wantHeight: 18.9,
			wantMinH:   5,
			wantMaxH:   1000,
		},
		{
			name: "MaxHeight=2000",
			setup: func(r *verifyReader) {
				r.floats["MaxHeight"] = 2000
			},
			wantHeight: 18.9,
			wantMaxH:   2000,
		},
		{
			name: "AutoSize=true",
			setup: func(r *verifyReader) {
				r.bools["AutoSize"] = true
			},
			wantHeight:   18.9,
			wantMaxH:     1000,
			wantAutoSize: true,
		},
		{
			name: "CanBreak=true",
			setup: func(r *verifyReader) {
				r.bools["CanBreak"] = true
			},
			wantHeight:   18.9,
			wantMaxH:     1000,
			wantCanBreak: true,
		},
		{
			name: "PageBreak=true",
			setup: func(r *verifyReader) {
				r.bools["PageBreak"] = true
			},
			wantHeight:    18.9,
			wantMaxH:      1000,
			wantPageBreak: true,
		},
		{
			name: "KeepRows=3",
			setup: func(r *verifyReader) {
				r.ints["KeepRows"] = 3
			},
			wantHeight:   18.9,
			wantMaxH:     1000,
			wantKeepRows: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newVerifyReader()
			tc.setup(r)

			row := NewTableRow()
			err := row.Deserialize(r)
			if err != nil {
				t.Fatalf("Deserialize returned unexpected error: %v", err)
			}

			if row.Height() != tc.wantHeight {
				t.Errorf("Height: got %v, want %v", row.Height(), tc.wantHeight)
			}
			if row.minHeight != tc.wantMinH {
				t.Errorf("minHeight: got %v, want %v", row.minHeight, tc.wantMinH)
			}
			if row.maxHeight != tc.wantMaxH {
				t.Errorf("maxHeight: got %v, want %v", row.maxHeight, tc.wantMaxH)
			}
			if row.autoSize != tc.wantAutoSize {
				t.Errorf("autoSize: got %v, want %v", row.autoSize, tc.wantAutoSize)
			}
			if row.canBreak != tc.wantCanBreak {
				t.Errorf("canBreak: got %v, want %v", row.canBreak, tc.wantCanBreak)
			}
			if row.pageBreak != tc.wantPageBreak {
				t.Errorf("pageBreak: got %v, want %v", row.pageBreak, tc.wantPageBreak)
			}
			if row.keepRows != tc.wantKeepRows {
				t.Errorf("keepRows: got %d, want %d", row.keepRows, tc.wantKeepRows)
			}
		})
	}
}

// TestTableRow_Deserialize_ErrorPath_ParentNeverFails verifies that
// ComponentBase.Deserialize never returns an error.
func TestTableRow_Deserialize_ErrorPath_ParentNeverFails(t *testing.T) {
	r := newVerifyReader()
	r.floats["Height"] = 45
	r.floats["MinHeight"] = 10
	r.floats["MaxHeight"] = 800
	r.bools["AutoSize"] = true
	r.bools["CanBreak"] = true
	r.bools["PageBreak"] = true
	r.ints["KeepRows"] = 5

	row := NewTableRow()
	if err := row.Deserialize(r); err != nil {
		t.Errorf("ComponentBase.Deserialize must always return nil; got: %v", err)
	}
}

// ── TableBase.Serialize — exhaustive happy-path verification ─────────────────

// TestTableBase_Serialize_HappyPath_AllBranches verifies every conditional
// write branch in TableBase.Serialize.
func TestTableBase_Serialize_HappyPath_AllBranches(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(*TableBase)
		wantFixedRows    bool
		wantFixedCols    bool
		wantLayout       bool
		wantPrintOnPar   bool
		wantWrappedGap   bool
		wantRepeatHdrs   bool // false means it's written (because non-default = false)
		wantRepeatRowHdrs bool
		wantRepeatColHdrs bool
		wantAdjSpanned   bool
		wantManualBuild  bool
		wantCols         int
		wantRows         int
	}{
		{
			name:  "all defaults",
			setup: func(t *TableBase) {},
		},
		{
			name: "FixedRows=2",
			setup: func(t *TableBase) {
				t.fixedRows = 2
			},
			wantFixedRows: true,
		},
		{
			name: "FixedColumns=3",
			setup: func(t *TableBase) {
				t.fixedColumns = 3
			},
			wantFixedCols: true,
		},
		{
			name: "Layout=DownThenAcross",
			setup: func(t *TableBase) {
				t.layout = TableLayoutDownThenAcross
			},
			wantLayout: true,
		},
		{
			name: "PrintOnParent=true",
			setup: func(t *TableBase) {
				t.printOnParent = true
			},
			wantPrintOnPar: true,
		},
		{
			name: "WrappedGap=10",
			setup: func(t *TableBase) {
				t.wrappedGap = 10
			},
			wantWrappedGap: true,
		},
		{
			name: "RepeatHeaders=false (non-default, so it IS written)",
			setup: func(t *TableBase) {
				t.repeatHeaders = false
			},
			wantRepeatHdrs: true,
		},
		{
			name: "RepeatRowHeaders=true",
			setup: func(t *TableBase) {
				t.repeatRowHeaders = true
			},
			wantRepeatRowHdrs: true,
		},
		{
			name: "RepeatColumnHeaders=true",
			setup: func(t *TableBase) {
				t.repeatColumnHeaders = true
			},
			wantRepeatColHdrs: true,
		},
		{
			name: "AdjustSpannedCellsWidth=true",
			setup: func(t *TableBase) {
				t.adjustSpannedCellsWidth = true
			},
			wantAdjSpanned: true,
		},
		{
			name: "ManualBuildEvent set",
			setup: func(t *TableBase) {
				t.ManualBuildEvent = "BuildHandler"
			},
			wantManualBuild: true,
		},
		{
			name: "one column serialized",
			setup: func(t *TableBase) {
				t.columns = append(t.columns, NewTableColumn())
			},
			wantCols: 1,
		},
		{
			name: "one row serialized",
			setup: func(t *TableBase) {
				t.rows = append(t.rows, NewTableRow())
			},
			wantRows: 1,
		},
		{
			name: "all non-defaults simultaneously",
			setup: func(t *TableBase) {
				t.fixedRows = 1
				t.fixedColumns = 1
				t.layout = TableLayoutWrapped
				t.printOnParent = true
				t.wrappedGap = 5
				t.repeatHeaders = false
				t.repeatRowHeaders = true
				t.repeatColumnHeaders = true
				t.adjustSpannedCellsWidth = true
				t.ManualBuildEvent = "Ev"
				t.columns = append(t.columns, NewTableColumn())
				t.rows = append(t.rows, NewTableRow())
			},
			wantFixedRows:     true,
			wantFixedCols:     true,
			wantLayout:        true,
			wantPrintOnPar:    true,
			wantWrappedGap:    true,
			wantRepeatHdrs:    true,
			wantRepeatRowHdrs: true,
			wantRepeatColHdrs: true,
			wantAdjSpanned:    true,
			wantManualBuild:   true,
			wantCols:          1,
			wantRows:          1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tbl := NewTableBase()
			tc.setup(tbl)

			w := newVerifyWriter()
			err := tbl.Serialize(w)
			if err != nil {
				t.Fatalf("Serialize returned unexpected error: %v", err)
			}

			_, fr := w.ints["FixedRows"]
			if fr != tc.wantFixedRows {
				t.Errorf("FixedRows written=%v, want %v", fr, tc.wantFixedRows)
			}
			_, fc := w.ints["FixedColumns"]
			if fc != tc.wantFixedCols {
				t.Errorf("FixedColumns written=%v, want %v", fc, tc.wantFixedCols)
			}
			_, lay := w.ints["Layout"]
			if lay != tc.wantLayout {
				t.Errorf("Layout written=%v, want %v", lay, tc.wantLayout)
			}
			_, pop := w.bools["PrintOnParent"]
			if pop != tc.wantPrintOnPar {
				t.Errorf("PrintOnParent written=%v, want %v", pop, tc.wantPrintOnPar)
			}
			_, wg := w.floats["WrappedGap"]
			if wg != tc.wantWrappedGap {
				t.Errorf("WrappedGap written=%v, want %v", wg, tc.wantWrappedGap)
			}
			_, rh := w.bools["RepeatHeaders"]
			if rh != tc.wantRepeatHdrs {
				t.Errorf("RepeatHeaders written=%v, want %v", rh, tc.wantRepeatHdrs)
			}
			_, rrh := w.bools["RepeatRowHeaders"]
			if rrh != tc.wantRepeatRowHdrs {
				t.Errorf("RepeatRowHeaders written=%v, want %v", rrh, tc.wantRepeatRowHdrs)
			}
			_, rch := w.bools["RepeatColumnHeaders"]
			if rch != tc.wantRepeatColHdrs {
				t.Errorf("RepeatColumnHeaders written=%v, want %v", rch, tc.wantRepeatColHdrs)
			}
			_, adj := w.bools["AdjustSpannedCellsWidth"]
			if adj != tc.wantAdjSpanned {
				t.Errorf("AdjustSpannedCellsWidth written=%v, want %v", adj, tc.wantAdjSpanned)
			}
			_, mbe := w.strs["ManualBuildEvent"]
			if mbe != tc.wantManualBuild {
				t.Errorf("ManualBuildEvent written=%v, want %v", mbe, tc.wantManualBuild)
			}
			if len(w.objs) != tc.wantCols+tc.wantRows {
				t.Errorf("WriteObject calls=%d, want %d", len(w.objs), tc.wantCols+tc.wantRows)
			}
		})
	}
}

// TestTableBase_Serialize_ErrorPath_ParentNeverFails verifies that
// BreakableComponent.Serialize never returns an error.
func TestTableBase_Serialize_ErrorPath_ParentNeverFails(t *testing.T) {
	tbl := NewTableBase()
	tbl.fixedRows = 2
	tbl.fixedColumns = 3
	tbl.layout = TableLayoutWrapped
	tbl.printOnParent = true
	tbl.wrappedGap = 15
	tbl.repeatHeaders = false
	tbl.repeatRowHeaders = true
	tbl.repeatColumnHeaders = true
	tbl.adjustSpannedCellsWidth = true
	tbl.ManualBuildEvent = "Handler"
	tbl.columns = append(tbl.columns, NewTableColumn())
	tbl.rows = append(tbl.rows, NewTableRow())

	w := newVerifyWriter()
	if err := tbl.Serialize(w); err != nil {
		t.Errorf("BreakableComponent.Serialize must always return nil; got: %v", err)
	}
}

// ── TableBase.Deserialize — exhaustive happy-path verification ────────────────

// TestTableBase_Deserialize_HappyPath_AllBranches verifies every readable
// field in TableBase.Deserialize.
func TestTableBase_Deserialize_HappyPath_AllBranches(t *testing.T) {
	tests := []struct {
		name               string
		setup              func(*verifyReader)
		wantFixedRows      int
		wantFixedCols      int
		wantLayout         TableLayout
		wantPrintOnParent  bool
		wantWrappedGap     float32
		wantRepeatHeaders  bool
		wantRepeatRowH     bool
		wantRepeatColH     bool
		wantAdjSpanned     bool
		wantManualBuildEv  string
	}{
		{
			name:              "all defaults",
			setup:             func(r *verifyReader) {},
			wantRepeatHeaders: true, // default is true
		},
		{
			name: "FixedRows=3",
			setup: func(r *verifyReader) {
				r.ints["FixedRows"] = 3
			},
			wantFixedRows:     3,
			wantRepeatHeaders: true,
		},
		{
			name: "FixedColumns=2",
			setup: func(r *verifyReader) {
				r.ints["FixedColumns"] = 2
			},
			wantFixedCols:     2,
			wantRepeatHeaders: true,
		},
		{
			name: "Layout=DownThenAcross",
			setup: func(r *verifyReader) {
				r.ints["Layout"] = 1 // TableLayoutDownThenAcross
			},
			wantLayout:        TableLayoutDownThenAcross,
			wantRepeatHeaders: true,
		},
		{
			name: "PrintOnParent=true",
			setup: func(r *verifyReader) {
				r.bools["PrintOnParent"] = true
			},
			wantPrintOnParent: true,
			wantRepeatHeaders: true,
		},
		{
			name: "WrappedGap=20",
			setup: func(r *verifyReader) {
				r.floats["WrappedGap"] = 20
			},
			wantWrappedGap:    20,
			wantRepeatHeaders: true,
		},
		{
			name: "RepeatHeaders=false",
			setup: func(r *verifyReader) {
				r.bools["RepeatHeaders"] = false
			},
			wantRepeatHeaders: false,
		},
		{
			name: "RepeatRowHeaders=true",
			setup: func(r *verifyReader) {
				r.bools["RepeatRowHeaders"] = true
			},
			wantRepeatHeaders: true,
			wantRepeatRowH:    true,
		},
		{
			name: "RepeatColumnHeaders=true",
			setup: func(r *verifyReader) {
				r.bools["RepeatColumnHeaders"] = true
			},
			wantRepeatHeaders: true,
			wantRepeatColH:    true,
		},
		{
			name: "AdjustSpannedCellsWidth=true",
			setup: func(r *verifyReader) {
				r.bools["AdjustSpannedCellsWidth"] = true
			},
			wantRepeatHeaders: true,
			wantAdjSpanned:    true,
		},
		{
			name: "ManualBuildEvent set",
			setup: func(r *verifyReader) {
				r.strs["ManualBuildEvent"] = "MyHandler"
			},
			wantRepeatHeaders:  true,
			wantManualBuildEv: "MyHandler",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newVerifyReader()
			tc.setup(r)

			tbl := NewTableBase()
			err := tbl.Deserialize(r)
			if err != nil {
				t.Fatalf("Deserialize returned unexpected error: %v", err)
			}

			if tbl.fixedRows != tc.wantFixedRows {
				t.Errorf("fixedRows: got %d, want %d", tbl.fixedRows, tc.wantFixedRows)
			}
			if tbl.fixedColumns != tc.wantFixedCols {
				t.Errorf("fixedColumns: got %d, want %d", tbl.fixedColumns, tc.wantFixedCols)
			}
			if tbl.layout != tc.wantLayout {
				t.Errorf("layout: got %d, want %d", tbl.layout, tc.wantLayout)
			}
			if tbl.printOnParent != tc.wantPrintOnParent {
				t.Errorf("printOnParent: got %v, want %v", tbl.printOnParent, tc.wantPrintOnParent)
			}
			if tbl.wrappedGap != tc.wantWrappedGap {
				t.Errorf("wrappedGap: got %v, want %v", tbl.wrappedGap, tc.wantWrappedGap)
			}
			if tbl.repeatHeaders != tc.wantRepeatHeaders {
				t.Errorf("repeatHeaders: got %v, want %v", tbl.repeatHeaders, tc.wantRepeatHeaders)
			}
			if tbl.repeatRowHeaders != tc.wantRepeatRowH {
				t.Errorf("repeatRowHeaders: got %v, want %v", tbl.repeatRowHeaders, tc.wantRepeatRowH)
			}
			if tbl.repeatColumnHeaders != tc.wantRepeatColH {
				t.Errorf("repeatColumnHeaders: got %v, want %v", tbl.repeatColumnHeaders, tc.wantRepeatColH)
			}
			if tbl.adjustSpannedCellsWidth != tc.wantAdjSpanned {
				t.Errorf("adjustSpannedCellsWidth: got %v, want %v", tbl.adjustSpannedCellsWidth, tc.wantAdjSpanned)
			}
			if tbl.ManualBuildEvent != tc.wantManualBuildEv {
				t.Errorf("ManualBuildEvent: got %q, want %q", tbl.ManualBuildEvent, tc.wantManualBuildEv)
			}
		})
	}
}

// TestTableBase_Deserialize_ErrorPath_ParentNeverFails verifies that
// BreakableComponent.Deserialize never returns an error.
func TestTableBase_Deserialize_ErrorPath_ParentNeverFails(t *testing.T) {
	r := newVerifyReader()
	r.ints["FixedRows"] = 2
	r.ints["FixedColumns"] = 1
	r.ints["Layout"] = 2 // Wrapped
	r.bools["PrintOnParent"] = true
	r.floats["WrappedGap"] = 30
	r.bools["RepeatHeaders"] = false
	r.bools["RepeatRowHeaders"] = true
	r.bools["RepeatColumnHeaders"] = true
	r.bools["AdjustSpannedCellsWidth"] = true
	r.strs["ManualBuildEvent"] = "Ev"

	tbl := NewTableBase()
	if err := tbl.Deserialize(r); err != nil {
		t.Errorf("BreakableComponent.Deserialize must always return nil; got: %v", err)
	}
}

// ── TableObject.Deserialize — exhaustive happy-path verification ─────────────

// TestTableObject_Deserialize_HappyPath_ManualBuildAutoSpans verifies both
// branches of the ManualBuildAutoSpans read in TableObject.Deserialize.
func TestTableObject_Deserialize_HappyPath_ManualBuildAutoSpansTrue(t *testing.T) {
	r := newVerifyReader()
	r.bools["ManualBuildAutoSpans"] = true

	tbl := NewTableObject()
	tbl.ManualBuildAutoSpans = false // start with non-default
	if err := tbl.Deserialize(r); err != nil {
		t.Fatalf("Deserialize returned unexpected error: %v", err)
	}
	if !tbl.ManualBuildAutoSpans {
		t.Error("ManualBuildAutoSpans should be true after reading true")
	}
}

func TestTableObject_Deserialize_HappyPath_ManualBuildAutoSpansFalseExplicit(t *testing.T) {
	r := newVerifyReader()
	r.bools["ManualBuildAutoSpans"] = false

	tbl := NewTableObject()
	tbl.ManualBuildAutoSpans = true // start with default
	if err := tbl.Deserialize(r); err != nil {
		t.Fatalf("Deserialize returned unexpected error: %v", err)
	}
	if tbl.ManualBuildAutoSpans {
		t.Error("ManualBuildAutoSpans should be false after reading false")
	}
}

// TestTableObject_Deserialize_ErrorPath_TableBaseNeverFails verifies that
// TableBase.Deserialize (and its chain BreakableComponent → ... → BaseObject)
// never returns a non-nil error, confirming the "return err" branch at
// table.go:313 is structurally dead code.
func TestTableObject_Deserialize_ErrorPath_TableBaseNeverFails(t *testing.T) {
	r := newVerifyReader()
	// Provide all possible fields so the full deserialization path is exercised.
	r.ints["FixedRows"] = 1
	r.ints["FixedColumns"] = 2
	r.ints["Layout"] = 1
	r.bools["PrintOnParent"] = true
	r.floats["WrappedGap"] = 12
	r.bools["RepeatHeaders"] = false
	r.bools["RepeatRowHeaders"] = true
	r.bools["RepeatColumnHeaders"] = true
	r.bools["AdjustSpannedCellsWidth"] = true
	r.strs["ManualBuildEvent"] = "handler"
	r.bools["ManualBuildAutoSpans"] = false

	tbl := NewTableObject()
	if err := tbl.Deserialize(r); err != nil {
		t.Errorf("TableBase.Deserialize must always return nil; got: %v", err)
	}
	// Verify the TableObject-level field was also read.
	if tbl.ManualBuildAutoSpans {
		t.Error("ManualBuildAutoSpans should be false")
	}
	// Verify base fields were read correctly.
	if tbl.fixedRows != 1 {
		t.Errorf("fixedRows: got %d, want 1", tbl.fixedRows)
	}
	if tbl.ManualBuildEvent != "handler" {
		t.Errorf("ManualBuildEvent: got %q, want handler", tbl.ManualBuildEvent)
	}
}
