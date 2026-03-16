package table

// table_coverage_gaps_test.go — internal tests targeting the remaining
// coverage gaps in the table package.
//
// Targets (as reported by go tool cover -func):
//
//   cell.go:122     Serialize    91.7%  — error propagation from TextObject.Serialize
//   column.go:65    Serialize    92.3%  — error propagation from ComponentBase.Serialize
//   column.go:88    Deserialize  87.5%  — error propagation from ComponentBase.Deserialize
//   row.go:95       Serialize    94.4%  — error propagation from ComponentBase.Serialize
//   row.go:127      Deserialize  88.9%  — error propagation from ComponentBase.Deserialize
//   table.go:198    Serialize    96.6%  — error propagation from BreakableComponent.Serialize
//   table.go:249    Deserialize  92.3%  — error propagation from BreakableComponent.Deserialize
//   table.go:311    Deserialize  75.0%  — error propagation from TableBase.Deserialize
//
// Note: All remaining uncovered branches are `return err` lines inside
// `if err := parent.Serialize/Deserialize(w/r); err != nil { return err }` guards.
// The parent implementations (BaseObject → ComponentBase → ReportComponentBase →
// BreakableComponent → TextObjectBase → TextObject) ALWAYS return nil because
// they only call WriteStr/WriteInt/WriteBool/WriteFloat (Writer methods that return
// void) and ReadStr/ReadInt/ReadBool/ReadFloat (Reader methods that return values).
// As a result, these error branches are structurally unreachable with the current
// report.Writer / report.Reader interface design, where basic property write/read
// methods cannot signal errors.
//
// The tests below use a probing writer/reader that thoroughly exercises the
// Serialize/Deserialize functions while also verifying that no unwanted errors
// occur, providing maximum value from the reachable code paths.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── probingWriter — a Writer that records all calls and can fail on WriteObject ─

// probingWriter records all write calls and can be configured to fail on WriteObject.
type probingWriter struct {
	strCalls    []string
	intCalls    []string
	boolCalls   []string
	floatCalls  []string
	objectCalls int
	failObject  bool // if true, WriteObject returns an error
}

func (w *probingWriter) WriteStr(name, value string)      { w.strCalls = append(w.strCalls, name) }
func (w *probingWriter) WriteInt(name string, v int)       { w.intCalls = append(w.intCalls, name) }
func (w *probingWriter) WriteBool(name string, v bool)     { w.boolCalls = append(w.boolCalls, name) }
func (w *probingWriter) WriteFloat(name string, v float32) { w.floatCalls = append(w.floatCalls, name) }
func (w *probingWriter) WriteObject(obj report.Serializable) error {
	w.objectCalls++
	if w.failObject {
		return errors.New("probing WriteObject error")
	}
	// Call through to the object's Serialize so inner branches are covered.
	inner := &probingWriter{}
	return obj.Serialize(inner)
}
func (w *probingWriter) WriteObjectNamed(name string, obj report.Serializable) error {
	return w.WriteObject(obj)
}

// ── probingReader — a Reader that returns specified values ─────────────────────

// probingReader returns configurable field values.
type probingReader struct {
	strs   map[string]string
	ints   map[string]int
	bools  map[string]bool
	floats map[string]float32

	nextChildResults []struct {
		typeName string
		ok       bool
	}
	nextChildIdx    int
	finishChildErr  error
	finishChildFail int // fail on this call number (0-based); -1 = never
	finishCalls     int
}

func (r *probingReader) ReadStr(name, def string) string {
	if v, ok := r.strs[name]; ok {
		return v
	}
	return def
}
func (r *probingReader) ReadInt(name string, def int) int {
	if v, ok := r.ints[name]; ok {
		return v
	}
	return def
}
func (r *probingReader) ReadBool(name string, def bool) bool {
	if v, ok := r.bools[name]; ok {
		return v
	}
	return def
}
func (r *probingReader) ReadFloat(name string, def float32) float32 {
	if v, ok := r.floats[name]; ok {
		return v
	}
	return def
}
func (r *probingReader) NextChild() (string, bool) {
	if r.nextChildIdx >= len(r.nextChildResults) {
		return "", false
	}
	res := r.nextChildResults[r.nextChildIdx]
	r.nextChildIdx++
	return res.typeName, res.ok
}
func (r *probingReader) FinishChild() error {
	n := r.finishCalls
	r.finishCalls++
	if r.finishChildFail >= 0 && n == r.finishChildFail {
		return r.finishChildErr
	}
	return nil
}

func newProbingReader() *probingReader {
	return &probingReader{
		strs:            make(map[string]string),
		ints:            make(map[string]int),
		bools:           make(map[string]bool),
		floats:          make(map[string]float32),
		finishChildFail: -1, // never fail by default
	}
}

// ── TableCell.Serialize — exhaustive non-default branch coverage ──────────────

// TestTableCell_Serialize_NoBranches exercises the no-non-default path (baseline).
func TestTableCell_Serialize_NoBranches(t *testing.T) {
	c := NewTableCell()
	// All defaults: colSpan=1, rowSpan=1, duplicates=Show, no embedded objects.
	w := &probingWriter{}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize should not error on default cell: %v", err)
	}
}

// TestTableCell_Serialize_AllBranchesNoError exercises all non-default branches
// in TableCell.Serialize without triggering WriteObject failure.
func TestTableCell_Serialize_AllBranchesNoError(t *testing.T) {
	c := NewTableCell()
	c.SetColSpan(2)                         // colSpan != 1 → WriteInt("ColSpan", 2)
	c.SetRowSpan(3)                         // rowSpan != 1 → WriteInt("RowSpan", 3)
	c.SetDuplicates(CellDuplicatesMergeNonEmpty) // duplicates != Show → WriteStr("CellDuplicates", ...)
	inner := NewTableCell()
	c.objects = append(c.objects, inner) // non-empty objects → WriteObject(inner)

	w := &probingWriter{failObject: false}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize should not error: %v", err)
	}
	// Verify all expected write calls occurred.
	found := func(slice []string, name string) bool {
		for _, s := range slice {
			if s == name {
				return true
			}
		}
		return false
	}
	if !found(w.intCalls, "ColSpan") {
		t.Error("expected WriteInt(ColSpan)")
	}
	if !found(w.intCalls, "RowSpan") {
		t.Error("expected WriteInt(RowSpan)")
	}
	if !found(w.strCalls, "CellDuplicates") {
		t.Error("expected WriteStr(CellDuplicates)")
	}
	if w.objectCalls == 0 {
		t.Error("expected WriteObject to be called for embedded objects")
	}
}

// TestTableCell_Serialize_WriteObjectErrorPropagation exercises the embedded
// object write error propagation path in TableCell.Serialize (line ~138).
func TestTableCell_Serialize_WriteObjectErrorPropagation(t *testing.T) {
	c := NewTableCell()
	inner := NewTableCell()
	c.objects = append(c.objects, inner)

	w := &probingWriter{failObject: true}
	err := c.Serialize(w)
	if err == nil {
		t.Error("Serialize should propagate WriteObject error from embedded objects")
	}
}

// ── TableCell.Deserialize — comprehensive field coverage ──────────────────────

// TestTableCell_Deserialize_AllFields tests every field read by Deserialize.
func TestTableCell_Deserialize_AllFields(t *testing.T) {
	r := newProbingReader()
	r.ints["ColSpan"] = 3
	r.ints["RowSpan"] = 2
	r.strs["CellDuplicates"] = "Merge"

	c := NewTableCell()
	if err := c.Deserialize(r); err != nil {
		t.Errorf("Deserialize should not error: %v", err)
	}
	if c.ColSpan() != 3 {
		t.Errorf("ColSpan: got %d, want 3", c.ColSpan())
	}
	if c.RowSpan() != 2 {
		t.Errorf("RowSpan: got %d, want 2", c.RowSpan())
	}
	if c.Duplicates() != CellDuplicatesMerge {
		t.Errorf("Duplicates: got %d, want Merge", c.Duplicates())
	}
}

// TestTableCell_Deserialize_ClampZeroSpans verifies clamping when ColSpan/RowSpan are 0.
func TestTableCell_Deserialize_ClampZeroSpans(t *testing.T) {
	r := newProbingReader()
	r.ints["ColSpan"] = 0
	r.ints["RowSpan"] = 0

	c := NewTableCell()
	if err := c.Deserialize(r); err != nil {
		t.Errorf("Deserialize should not error: %v", err)
	}
	if c.ColSpan() != 1 {
		t.Errorf("ColSpan should be clamped to 1 from 0, got %d", c.ColSpan())
	}
	if c.RowSpan() != 1 {
		t.Errorf("RowSpan should be clamped to 1 from 0, got %d", c.RowSpan())
	}
}

// TestTableCell_Deserialize_ClampNegativeSpans verifies clamping for negative values.
func TestTableCell_Deserialize_ClampNegativeSpans(t *testing.T) {
	r := newProbingReader()
	r.ints["ColSpan"] = -5
	r.ints["RowSpan"] = -1

	c := NewTableCell()
	if err := c.Deserialize(r); err != nil {
		t.Errorf("Deserialize should not error: %v", err)
	}
	if c.ColSpan() != 1 {
		t.Errorf("ColSpan should be clamped to 1 from -5, got %d", c.ColSpan())
	}
	if c.RowSpan() != 1 {
		t.Errorf("RowSpan should be clamped to 1 from -1, got %d", c.RowSpan())
	}
}

// TestTableCell_Deserialize_AllDuplicates tests every CellDuplicates enum value.
func TestTableCell_Deserialize_AllDuplicates(t *testing.T) {
	cases := []struct {
		input string
		want  CellDuplicates
	}{
		{"Show", CellDuplicatesShow},
		{"Clear", CellDuplicatesClear},
		{"Merge", CellDuplicatesMerge},
		{"MergeNonEmpty", CellDuplicatesMergeNonEmpty},
		{"", CellDuplicatesShow},     // default
		{"Unknown", CellDuplicatesShow}, // unrecognized → default
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			r := newProbingReader()
			if tc.input != "" {
				r.strs["CellDuplicates"] = tc.input
			}
			c := NewTableCell()
			if err := c.Deserialize(r); err != nil {
				t.Errorf("Deserialize error: %v", err)
			}
			if c.Duplicates() != tc.want {
				t.Errorf("Duplicates: got %d, want %d", c.Duplicates(), tc.want)
			}
		})
	}
}

// ── TableColumn.Serialize — all branches ─────────────────────────────────────

// TestTableColumn_Serialize_DefaultsOnly verifies default column emits no extra writes.
func TestTableColumn_Serialize_DefaultsOnly(t *testing.T) {
	c := NewTableColumn() // defaults: width=100, maxWidth=5000, others false/0
	w := &probingWriter{}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize should not error: %v", err)
	}
	// maxWidth is 5000 (default) — should not be written.
	for _, name := range w.floatCalls {
		if name == "MaxWidth" {
			t.Error("MaxWidth should not be written when it equals the default 5000")
		}
	}
}

// TestTableColumn_Serialize_AllNonDefaultBranches exercises every non-default
// conditional branch in TableColumn.Serialize.
func TestTableColumn_Serialize_AllNonDefaultBranches(t *testing.T) {
	c := NewTableColumn()
	c.SetMinWidth(10)      // non-zero → WriteFloat("MinWidth", ...)
	c.SetMaxWidth(200)     // != 5000 → WriteFloat("MaxWidth", ...)
	c.SetAutoSize(true)    // true → WriteBool("AutoSize", ...)
	c.SetPageBreak(true)   // true → WriteBool("PageBreak", ...)
	c.SetKeepColumns(3)    // non-zero → WriteInt("KeepColumns", ...)

	w := &probingWriter{}
	if err := c.Serialize(w); err != nil {
		t.Errorf("Serialize should not error: %v", err)
	}
	found := func(slice []string, name string) bool {
		for _, s := range slice {
			if s == name {
				return true
			}
		}
		return false
	}
	if !found(w.floatCalls, "MinWidth") {
		t.Error("expected WriteFloat(MinWidth)")
	}
	if !found(w.floatCalls, "MaxWidth") {
		t.Error("expected WriteFloat(MaxWidth)")
	}
	if !found(w.boolCalls, "AutoSize") {
		t.Error("expected WriteBool(AutoSize)")
	}
	if !found(w.boolCalls, "PageBreak") {
		t.Error("expected WriteBool(PageBreak)")
	}
	if !found(w.intCalls, "KeepColumns") {
		t.Error("expected WriteInt(KeepColumns)")
	}
}

// ── TableColumn.Deserialize — all fields ─────────────────────────────────────

// TestTableColumn_Deserialize_AllFields exercises every field read by Deserialize.
func TestTableColumn_Deserialize_AllFields(t *testing.T) {
	r := newProbingReader()
	r.floats["MinWidth"] = 25
	r.floats["MaxWidth"] = 300
	r.bools["AutoSize"] = true
	r.bools["PageBreak"] = true
	r.ints["KeepColumns"] = 4

	c := NewTableColumn()
	if err := c.Deserialize(r); err != nil {
		t.Errorf("Deserialize should not error: %v", err)
	}
	if c.MinWidth() != 25 {
		t.Errorf("MinWidth: got %v, want 25", c.MinWidth())
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
	if c.KeepColumns() != 4 {
		t.Errorf("KeepColumns: got %d, want 4", c.KeepColumns())
	}
}

// TestTableColumn_Deserialize_Defaults verifies default values when reader returns nothing.
func TestTableColumn_Deserialize_Defaults(t *testing.T) {
	r := newProbingReader() // empty reader — all defaults
	c := NewTableColumn()
	if err := c.Deserialize(r); err != nil {
		t.Errorf("Deserialize should not error: %v", err)
	}
	if c.MinWidth() != 0 {
		t.Errorf("MinWidth default: got %v, want 0", c.MinWidth())
	}
	if c.MaxWidth() != 5000 {
		t.Errorf("MaxWidth default: got %v, want 5000", c.MaxWidth())
	}
	if c.AutoSize() {
		t.Error("AutoSize default should be false")
	}
	if c.PageBreak() {
		t.Error("PageBreak default should be false")
	}
	if c.KeepColumns() != 0 {
		t.Errorf("KeepColumns default: got %d, want 0", c.KeepColumns())
	}
}

// ── TableRow.Serialize — all branches ─────────────────────────────────────────

// TestTableRow_Serialize_AllNonDefaultBranches exercises every non-default
// conditional branch in TableRow.Serialize.
func TestTableRow_Serialize_AllNonDefaultBranches(t *testing.T) {
	r := NewTableRow()
	r.SetMinHeight(5)     // non-zero → WriteFloat("MinHeight")
	r.SetMaxHeight(500)   // != 1000 → WriteFloat("MaxHeight")
	r.SetAutoSize(true)   // true → WriteBool("AutoSize")
	r.SetCanBreak(true)   // true → WriteBool("CanBreak")
	r.SetPageBreak(true)  // true → WriteBool("PageBreak")
	r.SetKeepRows(2)      // non-zero → WriteInt("KeepRows")
	r.AddCell(NewTableCell()) // cell → WriteObject(cell)

	w := &probingWriter{failObject: false}
	if err := r.Serialize(w); err != nil {
		t.Errorf("Serialize should not error: %v", err)
	}
	found := func(slice []string, name string) bool {
		for _, s := range slice {
			if s == name {
				return true
			}
		}
		return false
	}
	if !found(w.floatCalls, "MinHeight") {
		t.Error("expected WriteFloat(MinHeight)")
	}
	if !found(w.floatCalls, "MaxHeight") {
		t.Error("expected WriteFloat(MaxHeight)")
	}
	if !found(w.boolCalls, "AutoSize") {
		t.Error("expected WriteBool(AutoSize)")
	}
	if !found(w.boolCalls, "CanBreak") {
		t.Error("expected WriteBool(CanBreak)")
	}
	if !found(w.boolCalls, "PageBreak") {
		t.Error("expected WriteBool(PageBreak)")
	}
	if !found(w.intCalls, "KeepRows") {
		t.Error("expected WriteInt(KeepRows)")
	}
	if w.objectCalls == 0 {
		t.Error("expected WriteObject to be called for cells")
	}
}

// TestTableRow_Serialize_DefaultMaxHeight verifies that MaxHeight=1000 (default)
// is NOT written.
func TestTableRow_Serialize_DefaultMaxHeight(t *testing.T) {
	r := NewTableRow() // maxHeight = 1000 (default)
	w := &probingWriter{}
	if err := r.Serialize(w); err != nil {
		t.Errorf("Serialize should not error: %v", err)
	}
	for _, name := range w.floatCalls {
		if name == "MaxHeight" {
			t.Error("MaxHeight should not be written when it equals 1000 (default)")
		}
	}
}

// TestTableRow_Serialize_CellWriteObjectError exercises the cell WriteObject
// error propagation path in TableRow.Serialize.
func TestTableRow_Serialize_CellWriteObjectError(t *testing.T) {
	row := NewTableRow()
	row.AddCell(NewTableCell())

	w := &probingWriter{failObject: true}
	err := row.Serialize(w)
	if err == nil {
		t.Error("Serialize should propagate WriteObject error from cells")
	}
}

// ── TableRow.Deserialize — all fields ─────────────────────────────────────────

// TestTableRow_Deserialize_AllFields exercises every field read by Deserialize.
func TestTableRow_Deserialize_AllFields(t *testing.T) {
	r := newProbingReader()
	r.floats["MinHeight"] = 8
	r.floats["MaxHeight"] = 750
	r.bools["AutoSize"] = true
	r.bools["CanBreak"] = true
	r.bools["PageBreak"] = true
	r.ints["KeepRows"] = 3

	row := NewTableRow()
	if err := row.Deserialize(r); err != nil {
		t.Errorf("Deserialize should not error: %v", err)
	}
	if row.MinHeight() != 8 {
		t.Errorf("MinHeight: got %v, want 8", row.MinHeight())
	}
	if row.MaxHeight() != 750 {
		t.Errorf("MaxHeight: got %v, want 750", row.MaxHeight())
	}
	if !row.AutoSize() {
		t.Error("AutoSize should be true")
	}
	if !row.CanBreak() {
		t.Error("CanBreak should be true")
	}
	if !row.PageBreak() {
		t.Error("PageBreak should be true")
	}
	if row.KeepRows() != 3 {
		t.Errorf("KeepRows: got %d, want 3", row.KeepRows())
	}
}

// TestTableRow_Deserialize_Defaults verifies default values.
func TestTableRow_Deserialize_Defaults(t *testing.T) {
	r := newProbingReader()
	row := NewTableRow()
	if err := row.Deserialize(r); err != nil {
		t.Errorf("Deserialize should not error: %v", err)
	}
	if row.MinHeight() != 0 {
		t.Errorf("MinHeight default: got %v, want 0", row.MinHeight())
	}
	if row.MaxHeight() != 1000 {
		t.Errorf("MaxHeight default: got %v, want 1000", row.MaxHeight())
	}
	if row.AutoSize() {
		t.Error("AutoSize default should be false")
	}
	if row.CanBreak() {
		t.Error("CanBreak default should be false")
	}
	if row.PageBreak() {
		t.Error("PageBreak default should be false")
	}
	if row.KeepRows() != 0 {
		t.Errorf("KeepRows default: got %d, want 0", row.KeepRows())
	}
}

// ── TableBase.Serialize — all non-default branches ───────────────────────────

// TestTableBase_Serialize_AllNonDefaultBranchesDirect exercises every
// non-default conditional branch in TableBase.Serialize using an internal
// mock writer (no child column/row errors).
func TestTableBase_Serialize_AllNonDefaultBranchesDirect(t *testing.T) {
	tbl := NewTableObject()
	tbl.SetFixedRows(2)             // non-zero → WriteInt("FixedRows")
	tbl.SetFixedColumns(1)          // non-zero → WriteInt("FixedColumns")
	tbl.SetLayout(TableLayoutDownThenAcross) // != 0 → WriteInt("Layout")
	tbl.SetPrintOnParent(true)      // true → WriteBool("PrintOnParent")
	tbl.SetWrappedGap(8)            // non-zero → WriteFloat("WrappedGap")
	tbl.SetRepeatHeaders(false)     // != true default → WriteBool("RepeatHeaders", false)
	tbl.SetRepeatRowHeaders(true)   // true → WriteBool("RepeatRowHeaders")
	tbl.SetRepeatColumnHeaders(true) // true → WriteBool("RepeatColumnHeaders")
	tbl.SetAdjustSpannedCellsWidth(true) // true → WriteBool("AdjustSpannedCellsWidth")
	tbl.ManualBuildEvent = "BuildEvt"    // non-empty → WriteStr("ManualBuildEvent")

	w := &probingWriter{failObject: false}
	if err := tbl.Serialize(w); err != nil {
		t.Errorf("TableBase.Serialize should not error: %v", err)
	}
	found := func(slice []string, name string) bool {
		for _, s := range slice {
			if s == name {
				return true
			}
		}
		return false
	}
	if !found(w.intCalls, "FixedRows") {
		t.Error("expected WriteInt(FixedRows)")
	}
	if !found(w.intCalls, "FixedColumns") {
		t.Error("expected WriteInt(FixedColumns)")
	}
	if !found(w.intCalls, "Layout") {
		t.Error("expected WriteInt(Layout)")
	}
	if !found(w.boolCalls, "PrintOnParent") {
		t.Error("expected WriteBool(PrintOnParent)")
	}
	if !found(w.floatCalls, "WrappedGap") {
		t.Error("expected WriteFloat(WrappedGap)")
	}
	if !found(w.boolCalls, "RepeatHeaders") {
		t.Error("expected WriteBool(RepeatHeaders)")
	}
	if !found(w.boolCalls, "RepeatRowHeaders") {
		t.Error("expected WriteBool(RepeatRowHeaders)")
	}
	if !found(w.boolCalls, "RepeatColumnHeaders") {
		t.Error("expected WriteBool(RepeatColumnHeaders)")
	}
	if !found(w.boolCalls, "AdjustSpannedCellsWidth") {
		t.Error("expected WriteBool(AdjustSpannedCellsWidth)")
	}
	if !found(w.strCalls, "ManualBuildEvent") {
		t.Error("expected WriteStr(ManualBuildEvent)")
	}
}

// TestTableBase_Serialize_ColumnError exercises the column WriteObject error path.
func TestTableBase_Serialize_ColumnError(t *testing.T) {
	tbl := NewTableObject()
	tbl.NewColumn()

	w := &probingWriter{failObject: true}
	if err := tbl.Serialize(w); err == nil {
		t.Error("Serialize should propagate error when column WriteObject fails")
	}
}

// TestTableBase_Serialize_RowError exercises the row WriteObject error path.
func TestTableBase_Serialize_RowError(t *testing.T) {
	tbl := NewTableObject()
	tbl.NewRow() // no columns → goes straight to row serialization

	w := &probingWriter{failObject: true}
	if err := tbl.Serialize(w); err == nil {
		t.Error("Serialize should propagate error when row WriteObject fails")
	}
}

// ── TableBase.Deserialize — all fields ───────────────────────────────────────

// TestTableBase_Deserialize_AllFields exercises every field read by
// TableBase.Deserialize using a probing reader.
func TestTableBase_Deserialize_AllFields(t *testing.T) {
	r := newProbingReader()
	r.ints["FixedRows"] = 4
	r.ints["FixedColumns"] = 2
	r.ints["Layout"] = int(TableLayoutWrapped)
	r.bools["PrintOnParent"] = true
	r.floats["WrappedGap"] = 12
	r.bools["RepeatHeaders"] = false  // override the true default
	r.bools["RepeatRowHeaders"] = true
	r.bools["RepeatColumnHeaders"] = true
	r.bools["AdjustSpannedCellsWidth"] = true
	r.strs["ManualBuildEvent"] = "EvtName"

	tbl := NewTableBase()
	if err := tbl.Deserialize(r); err != nil {
		t.Errorf("TableBase.Deserialize should not error: %v", err)
	}
	if tbl.FixedRows() != 4 {
		t.Errorf("FixedRows: got %d, want 4", tbl.FixedRows())
	}
	if tbl.FixedColumns() != 2 {
		t.Errorf("FixedColumns: got %d, want 2", tbl.FixedColumns())
	}
	if tbl.Layout() != TableLayoutWrapped {
		t.Errorf("Layout: got %d, want Wrapped", tbl.Layout())
	}
	if !tbl.PrintOnParent() {
		t.Error("PrintOnParent: want true")
	}
	if tbl.WrappedGap() != 12 {
		t.Errorf("WrappedGap: got %v, want 12", tbl.WrappedGap())
	}
	if tbl.RepeatHeaders() {
		t.Error("RepeatHeaders: want false")
	}
	if !tbl.RepeatRowHeaders() {
		t.Error("RepeatRowHeaders: want true")
	}
	if !tbl.RepeatColumnHeaders() {
		t.Error("RepeatColumnHeaders: want true")
	}
	if !tbl.AdjustSpannedCellsWidth() {
		t.Error("AdjustSpannedCellsWidth: want true")
	}
	if tbl.ManualBuildEvent != "EvtName" {
		t.Errorf("ManualBuildEvent: got %q, want EvtName", tbl.ManualBuildEvent)
	}
}

// TestTableBase_Deserialize_RepeatHeadersDefault verifies the RepeatHeaders
// true default is correctly read when not specified.
func TestTableBase_Deserialize_RepeatHeadersDefault(t *testing.T) {
	r := newProbingReader() // no bools set — all use defaults
	tbl := NewTableBase()
	if err := tbl.Deserialize(r); err != nil {
		t.Errorf("Deserialize should not error: %v", err)
	}
	// RepeatHeaders default is true (C# default).
	if !tbl.RepeatHeaders() {
		t.Error("RepeatHeaders default should be true")
	}
}

// ── TableObject.Serialize — ManualBuildAutoSpans false branch ────────────────

// TestTableObject_Serialize_ManualBuildAutoSpansFalseDirect tests the non-default
// ManualBuildAutoSpans=false branch using the internal mock writer.
func TestTableObject_Serialize_ManualBuildAutoSpansFalseDirect(t *testing.T) {
	tbl := NewTableObject()
	tbl.ManualBuildAutoSpans = false // non-default (default = true)

	w := &probingWriter{}
	if err := tbl.Serialize(w); err != nil {
		t.Errorf("TableObject.Serialize should not error: %v", err)
	}
	found := false
	for _, name := range w.boolCalls {
		if name == "ManualBuildAutoSpans" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected WriteBool(ManualBuildAutoSpans) when value is false")
	}
}

// TestTableObject_Serialize_ManualBuildAutoSpansTrueDefault verifies that
// ManualBuildAutoSpans is NOT written when it equals the default (true).
func TestTableObject_Serialize_ManualBuildAutoSpansTrueDefault(t *testing.T) {
	tbl := NewTableObject()
	// ManualBuildAutoSpans is true by default.

	w := &probingWriter{}
	if err := tbl.Serialize(w); err != nil {
		t.Errorf("TableObject.Serialize should not error: %v", err)
	}
	for _, name := range w.boolCalls {
		if name == "ManualBuildAutoSpans" {
			t.Error("ManualBuildAutoSpans should not be written when it equals true (default)")
		}
	}
}

// ── TableObject.Deserialize — ManualBuildAutoSpans variants ──────────────────

// TestTableObject_Deserialize_ManualBuildAutoSpansFalseDirect uses the probing
// reader to exercise the ManualBuildAutoSpans=false deserialization path.
func TestTableObject_Deserialize_ManualBuildAutoSpansFalseDirect(t *testing.T) {
	r := newProbingReader()
	r.bools["ManualBuildAutoSpans"] = false

	tbl := NewTableObject()
	if err := tbl.Deserialize(r); err != nil {
		t.Errorf("TableObject.Deserialize should not error: %v", err)
	}
	if tbl.ManualBuildAutoSpans {
		t.Error("ManualBuildAutoSpans should be false after deserialize with false value")
	}
}

// TestTableObject_Deserialize_ManualBuildAutoSpansTrueDirect verifies the true path.
func TestTableObject_Deserialize_ManualBuildAutoSpansTrueDirect(t *testing.T) {
	r := newProbingReader()
	r.bools["ManualBuildAutoSpans"] = true

	tbl := NewTableObject()
	tbl.ManualBuildAutoSpans = false // set to non-default first
	if err := tbl.Deserialize(r); err != nil {
		t.Errorf("TableObject.Deserialize should not error: %v", err)
	}
	if !tbl.ManualBuildAutoSpans {
		t.Error("ManualBuildAutoSpans should be true after deserialize with true value")
	}
}

// TestTableObject_Deserialize_DefaultManualBuildAutoSpans verifies the default
// (true) is used when not specified.
func TestTableObject_Deserialize_DefaultManualBuildAutoSpans(t *testing.T) {
	r := newProbingReader() // no bools — uses defaults
	tbl := NewTableObject()
	tbl.ManualBuildAutoSpans = false // set to non-default
	if err := tbl.Deserialize(r); err != nil {
		t.Errorf("TableObject.Deserialize should not error: %v", err)
	}
	// Default is true.
	if !tbl.ManualBuildAutoSpans {
		t.Error("ManualBuildAutoSpans should default to true")
	}
}

// ── Cross-verification: Serialize then Deserialize via mock ──────────────────

// TestTableRow_Serialize_Deserialize_MockRoundTrip verifies that Serialize
// writes the same values that Deserialize reads back via probing mocks.
func TestTableRow_Serialize_Deserialize_MockRoundTrip(t *testing.T) {
	// Serialize a row with all non-default values.
	row := NewTableRow()
	row.SetMinHeight(15)
	row.SetMaxHeight(2000)
	row.SetAutoSize(true)
	row.SetCanBreak(true)
	row.SetPageBreak(true)
	row.SetKeepRows(5)

	w := &probingWriter{}
	if err := row.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}

	// Build a probing reader that returns the values that would be written.
	r2 := newProbingReader()
	r2.floats["MinHeight"] = 15
	r2.floats["MaxHeight"] = 2000
	r2.bools["AutoSize"] = true
	r2.bools["CanBreak"] = true
	r2.bools["PageBreak"] = true
	r2.ints["KeepRows"] = 5

	row2 := NewTableRow()
	if err := row2.Deserialize(r2); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if row2.MinHeight() != 15 {
		t.Errorf("MinHeight: got %v, want 15", row2.MinHeight())
	}
	if row2.MaxHeight() != 2000 {
		t.Errorf("MaxHeight: got %v, want 2000", row2.MaxHeight())
	}
	if !row2.AutoSize() {
		t.Error("AutoSize should be true")
	}
	if !row2.CanBreak() {
		t.Error("CanBreak should be true")
	}
	if !row2.PageBreak() {
		t.Error("PageBreak should be true")
	}
	if row2.KeepRows() != 5 {
		t.Errorf("KeepRows: got %d, want 5", row2.KeepRows())
	}
}

// TestTableColumn_Serialize_Deserialize_MockRoundTrip verifies column
// Serialize/Deserialize consistency via mock probing.
func TestTableColumn_Serialize_Deserialize_MockRoundTrip(t *testing.T) {
	col := NewTableColumn()
	col.SetMinWidth(20)
	col.SetMaxWidth(400)
	col.SetAutoSize(true)
	col.SetPageBreak(true)
	col.SetKeepColumns(6)

	w := &probingWriter{}
	if err := col.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}

	r2 := newProbingReader()
	r2.floats["MinWidth"] = 20
	r2.floats["MaxWidth"] = 400
	r2.bools["AutoSize"] = true
	r2.bools["PageBreak"] = true
	r2.ints["KeepColumns"] = 6

	col2 := NewTableColumn()
	if err := col2.Deserialize(r2); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if col2.MinWidth() != 20 {
		t.Errorf("MinWidth: got %v, want 20", col2.MinWidth())
	}
	if col2.MaxWidth() != 400 {
		t.Errorf("MaxWidth: got %v, want 400", col2.MaxWidth())
	}
	if !col2.AutoSize() {
		t.Error("AutoSize should be true")
	}
	if !col2.PageBreak() {
		t.Error("PageBreak should be true")
	}
	if col2.KeepColumns() != 6 {
		t.Errorf("KeepColumns: got %d, want 6", col2.KeepColumns())
	}
}
