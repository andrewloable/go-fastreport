package export_test

// exportbase_gaps_test.go — tests for features added to close ExportBase porting gaps:
//   - HasMultipleFiles field
//   - ShiftNonExportable field
//   - ExportBase.Serialize / Deserialize
//
// C# ref: FastReport.Base/Export/ExportBase.cs

import (
	"bytes"
	"testing"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/report"
)

// ── intCapableWriter/Reader — test doubles that also track int values ──────────

type intWriter struct {
	strs  map[string]string
	ints  map[string]int
	bools map[string]bool
}

func newIntWriter() *intWriter {
	return &intWriter{
		strs:  make(map[string]string),
		ints:  make(map[string]int),
		bools: make(map[string]bool),
	}
}

func (w *intWriter) WriteStr(name, value string)              { w.strs[name] = value }
func (w *intWriter) WriteInt(name string, value int)           { w.ints[name] = value }
func (w *intWriter) WriteBool(name string, value bool)         { w.bools[name] = value }
func (w *intWriter) WriteFloat(name string, value float32)     {}
func (w *intWriter) WriteObject(obj report.Serializable) error { return nil }
func (w *intWriter) WriteObjectNamed(_ string, _ report.Serializable) error { return nil }

type intReader struct {
	strs  map[string]string
	ints  map[string]int
	bools map[string]bool
}

func newIntReader(strs map[string]string, ints map[string]int, bools map[string]bool) *intReader {
	if strs == nil {
		strs = map[string]string{}
	}
	if ints == nil {
		ints = map[string]int{}
	}
	if bools == nil {
		bools = map[string]bool{}
	}
	return &intReader{strs: strs, ints: ints, bools: bools}
}

func (r *intReader) ReadStr(name, def string) string {
	if v, ok := r.strs[name]; ok {
		return v
	}
	return def
}
func (r *intReader) ReadInt(name string, def int) int {
	if v, ok := r.ints[name]; ok {
		return v
	}
	return def
}
func (r *intReader) ReadBool(name string, def bool) bool {
	if v, ok := r.bools[name]; ok {
		return v
	}
	return def
}
func (r *intReader) ReadFloat(name string, def float32) float32 { return def }
func (r *intReader) NextChild() (string, bool)                  { return "", false }
func (r *intReader) FinishChild() error                         { return nil }

// ── HasMultipleFiles ───────────────────────────────────────────────────────────

func TestExportBase_HasMultipleFiles_DefaultFalse(t *testing.T) {
	base := export.NewExportBase()
	if base.HasMultipleFiles {
		t.Error("HasMultipleFiles should default to false")
	}
}

func TestExportBase_HasMultipleFiles_CanBeSet(t *testing.T) {
	base := export.NewExportBase()
	base.HasMultipleFiles = true
	if !base.HasMultipleFiles {
		t.Error("HasMultipleFiles should be true after setting")
	}
}

// ── ShiftNonExportable ────────────────────────────────────────────────────────

func TestExportBase_ShiftNonExportable_DefaultFalse(t *testing.T) {
	base := export.NewExportBase()
	if base.ShiftNonExportable {
		t.Error("ShiftNonExportable should default to false")
	}
}

func TestExportBase_ShiftNonExportable_CanBeSet(t *testing.T) {
	base := export.NewExportBase()
	base.ShiftNonExportable = true
	if !base.ShiftNonExportable {
		t.Error("ShiftNonExportable should be true after setting")
	}
}

// ShiftNonExportable does not change Export behaviour in the Go port
// (band exportability is handled by the engine before PreparedPages are
// created), so we only test that the field is settable and that Export
// still succeeds.
func TestExportBase_ShiftNonExportable_ExportSucceeds(t *testing.T) {
	pp := buildPreparedPages(2, []string{"Header", "Data"})
	base := export.NewExportBase()
	base.ShiftNonExportable = true
	rec := newRecorder(new(bytes.Buffer))

	if err := base.Export(pp, rec.w, rec); err != nil {
		t.Fatalf("Export with ShiftNonExportable=true: %v", err)
	}
	if len(rec.pageBegin) != 2 {
		t.Errorf("expected 2 pages, got %d", len(rec.pageBegin))
	}
}

// ── ExportBase.Serialize / Deserialize ────────────────────────────────────────

func TestExportBase_Serialize_Defaults(t *testing.T) {
	// With default values nothing should be written (all values equal defaults).
	base := export.NewExportBase()
	w := newIntWriter()
	base.Serialize(w)

	// PageRange==0 (PageRangeAll) — Serialize skips it.
	if _, ok := w.ints["PageRange"]; ok {
		t.Error("PageRange should not be written for default PageRangeAll")
	}
	// PageNumbers=="" — Serialize skips it.
	if _, ok := w.strs["PageNumbers"]; ok {
		t.Error("PageNumbers should not be written when empty")
	}
	// ShiftNonExportable==false — Serialize skips it.
	if _, ok := w.bools["ShiftNonExportable"]; ok {
		t.Error("ShiftNonExportable should not be written when false")
	}
}

func TestExportBase_Serialize_NonDefaults(t *testing.T) {
	base := export.NewExportBase()
	base.PageRange = export.PageRangeCustom
	base.PageNumbers = "1,3-5"
	base.ShiftNonExportable = true
	base.HasMultipleFiles = true

	w := newIntWriter()
	base.Serialize(w)

	if w.ints["PageRange"] != int(export.PageRangeCustom) {
		t.Errorf("PageRange: got %d, want %d", w.ints["PageRange"], int(export.PageRangeCustom))
	}
	if w.strs["PageNumbers"] != "1,3-5" {
		t.Errorf("PageNumbers: got %q, want %q", w.strs["PageNumbers"], "1,3-5")
	}
	if !w.bools["ShiftNonExportable"] {
		t.Error("ShiftNonExportable should be written as true")
	}
	if !w.bools["HasMultipleFiles"] {
		t.Error("HasMultipleFiles should be written as true")
	}
}

func TestExportBase_Deserialize(t *testing.T) {
	base := export.NewExportBase()

	r := newIntReader(
		map[string]string{"PageNumbers": "2,4"},
		map[string]int{"PageRange": int(export.PageRangeCustom)},
		map[string]bool{
			"ShiftNonExportable": true,
			"HasMultipleFiles":   true,
		},
	)
	base.Deserialize(r)

	if base.PageRange != export.PageRangeCustom {
		t.Errorf("PageRange: got %d, want PageRangeCustom", base.PageRange)
	}
	if base.PageNumbers != "2,4" {
		t.Errorf("PageNumbers: got %q, want %q", base.PageNumbers, "2,4")
	}
	if !base.ShiftNonExportable {
		t.Error("ShiftNonExportable should be true after Deserialize")
	}
	if !base.HasMultipleFiles {
		t.Error("HasMultipleFiles should be true after Deserialize")
	}
}

func TestExportBase_Deserialize_Defaults(t *testing.T) {
	// Empty reader should restore all fields to their defaults.
	base := export.NewExportBase()
	base.PageRange = export.PageRangeCurrent // pre-set to non-default
	base.ShiftNonExportable = true            // pre-set to non-default

	r := newIntReader(nil, nil, nil)
	base.Deserialize(r)

	if base.PageRange != export.PageRangeAll {
		t.Errorf("PageRange: got %d, want PageRangeAll (default)", base.PageRange)
	}
	if base.ShiftNonExportable {
		t.Error("ShiftNonExportable should be false (default) after Deserialize with no data")
	}
}

// TestExportBase_Serialize_PageRangeCurrent verifies PageRangeCurrent is serialized.
func TestExportBase_Serialize_PageRangeCurrent(t *testing.T) {
	base := export.NewExportBase()
	base.PageRange = export.PageRangeCurrent

	w := newIntWriter()
	base.Serialize(w)

	if w.ints["PageRange"] != int(export.PageRangeCurrent) {
		t.Errorf("PageRange: got %d, want %d", w.ints["PageRange"], int(export.PageRangeCurrent))
	}
}
