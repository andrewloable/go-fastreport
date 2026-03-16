package object

// sparkline_svg_internal_test.go — internal package tests to cover the
// unreachable `return err` branches in SVGObject and SparklineObject after
// their parent Serialize/Deserialize calls.
//
// The parent chain (ReportComponentBase → ComponentBase → BaseObject) never
// returns an error from Serialize or Deserialize. To exercise the guard
// `if err := s.ReportComponentBase.Serialize(w); err != nil { return err }`
// we use a mock writer whose WriteStr triggers a downstream error by wrapping
// the object in an erroring parent — however, since the parent Serialize uses
// only Write* void methods (no WriteObjectNamed), the error is truly unreachable
// via an external writer.
//
// Instead we confirm the functions are callable without error and verify
// the positive paths with default/non-default values, ensuring the tool
// correctly sees both branches of each conditional write/read exercised.
//
// The true unreachable `return err` lines (e.g. svg.go:31, sparkline.go:34)
// remain at ~80-85% because they are dead code. This file focuses on
// other achievable improvements.

import (
	"errors"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ── SVGObject internal error-path tests ──────────────────────────────────────

// svgErrWriterBase is a report.Writer whose WriteStr always succeeds but whose
// WriteObject/WriteObjectNamed return an error. Since SVGObject.Serialize only
// calls WriteStr (not WriteObjectNamed), we cannot trigger the dead `return err`
// guard via the Writer interface alone. However, we include this writer for
// completeness and to ensure the object handles writers without panicking.
type svgNoopWriter struct{}

func (w *svgNoopWriter) WriteStr(name, value string)                                     {}
func (w *svgNoopWriter) WriteInt(name string, value int)                                 {}
func (w *svgNoopWriter) WriteBool(name string, value bool)                               {}
func (w *svgNoopWriter) WriteFloat(name string, value float32)                           {}
func (w *svgNoopWriter) WriteObject(obj report.Serializable) error                       { return nil }
func (w *svgNoopWriter) WriteObjectNamed(name string, obj report.Serializable) error     { return nil }

// TestSVGObject_Serialize_WithNoopWriter exercises SVGObject.Serialize with a
// noop writer, confirming no panic and nil error for both empty and non-empty SvgData.
func TestSVGObject_Serialize_WithNoopWriter(t *testing.T) {
	w := &svgNoopWriter{}

	// Empty SvgData — WriteStr("SvgData", ...) branch skipped.
	empty := &SVGObject{}
	empty.ReportComponentBase = *report.NewReportComponentBase()
	if err := empty.Serialize(w); err != nil {
		t.Errorf("SVGObject.Serialize (empty): unexpected error: %v", err)
	}

	// Non-empty SvgData — WriteStr("SvgData", ...) branch taken.
	filled := &SVGObject{SvgData: "PHN2Zy8+"}
	filled.ReportComponentBase = *report.NewReportComponentBase()
	if err := filled.Serialize(w); err != nil {
		t.Errorf("SVGObject.Serialize (filled): unexpected error: %v", err)
	}
}

// ── errReader covers Deserialize error guard paths ────────────────────────────

// alwaysErrReader is a report.Reader that returns errors for ReadStr/etc. so
// we can confirm the guard `if err := base.Deserialize(r)` branches.
// In practice the real BaseObject/ComponentBase.Deserialize never errors —
// they only call ReadStr/ReadInt/ReadBool/ReadFloat which this reader can return
// defaults for. So Deserialize will NOT error; we confirm Deserialize returns nil.
type alwaysDefaultReader struct{}

func (r *alwaysDefaultReader) ReadStr(name, def string) string          { return def }
func (r *alwaysDefaultReader) ReadInt(name string, def int) int         { return def }
func (r *alwaysDefaultReader) ReadBool(name string, def bool) bool      { return def }
func (r *alwaysDefaultReader) ReadFloat(name string, def float32) float32 { return def }
func (r *alwaysDefaultReader) NextChild() (string, bool)                { return "", false }
func (r *alwaysDefaultReader) FinishChild() error                       { return nil }

// TestSVGObject_Deserialize_WithDefaultReader exercises SVGObject.Deserialize
// with a reader that returns all defaults, confirming nil error.
func TestSVGObject_Deserialize_WithDefaultReader(t *testing.T) {
	svg := &SVGObject{}
	svg.ReportComponentBase = *report.NewReportComponentBase()
	r := &alwaysDefaultReader{}
	if err := svg.Deserialize(r); err != nil {
		t.Errorf("SVGObject.Deserialize: unexpected error: %v", err)
	}
	if svg.SvgData != "" {
		t.Errorf("SvgData: got %q, want empty", svg.SvgData)
	}
}

// ── SparklineObject internal error-path tests ─────────────────────────────────

// TestSparklineObject_Serialize_WithNoopWriter exercises SparklineObject.Serialize
// with a noop writer for empty and non-empty ChartData/Dock.
func TestSparklineObject_Serialize_WithNoopWriter(t *testing.T) {
	w := &svgNoopWriter{} // reuse the same noop writer shape

	// Empty fields — both WriteStr branches skipped.
	empty := &SparklineObject{}
	empty.ReportComponentBase = *report.NewReportComponentBase()
	if err := empty.Serialize(w); err != nil {
		t.Errorf("SparklineObject.Serialize (empty): unexpected error: %v", err)
	}

	// Non-empty ChartData + Dock — both WriteStr branches taken.
	filled := &SparklineObject{ChartData: "abc", Dock: "Fill"}
	filled.ReportComponentBase = *report.NewReportComponentBase()
	if err := filled.Serialize(w); err != nil {
		t.Errorf("SparklineObject.Serialize (filled): unexpected error: %v", err)
	}
}

// TestSparklineObject_Deserialize_WithDefaultReader exercises
// SparklineObject.Deserialize with a reader that returns all defaults.
func TestSparklineObject_Deserialize_WithDefaultReader(t *testing.T) {
	sp := &SparklineObject{}
	sp.ReportComponentBase = *report.NewReportComponentBase()
	r := &alwaysDefaultReader{}
	if err := sp.Deserialize(r); err != nil {
		t.Errorf("SparklineObject.Deserialize: unexpected error: %v", err)
	}
	if sp.ChartData != "" {
		t.Errorf("ChartData: got %q, want empty", sp.ChartData)
	}
	if sp.Dock != "" {
		t.Errorf("Dock: got %q, want empty", sp.Dock)
	}
}

// ── Confirm the "dead" return-err paths require a parent that can error ───────
// These tests document that the only way to hit the `return err` guard would be
// if ComponentBase/BaseObject.Serialize returned a non-nil error, which they never
// do in the current implementation. We therefore use a sentinel to confirm
// the behavior stays as expected.

// errReportComponentWriter is a writer that counts WriteStr calls. Used to
// verify that Serialize does call WriteStr (and thus ReportComponentBase.Serialize
// ran) before our own code.
type countingWriter struct {
	svgNoopWriter
	strCount int
}

func (w *countingWriter) WriteStr(name, value string) {
	w.strCount++
}

func TestSVGObject_Serialize_ParentSerializeRuns(t *testing.T) {
	// Set a Name so BaseObject.Serialize writes at least one WriteStr("Name", ...).
	svg := &SVGObject{SvgData: "data"}
	svg.ReportComponentBase = *report.NewReportComponentBase()
	svg.SetName("mySVG")

	w := &countingWriter{}
	if err := svg.Serialize(w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// At minimum: Name (from BaseObject) + SvgData (from SVGObject) = 2 WriteStr calls.
	if w.strCount < 2 {
		t.Errorf("expected at least 2 WriteStr calls, got %d", w.strCount)
	}
}

func TestSparklineObject_Serialize_ParentSerializeRuns(t *testing.T) {
	sp := &SparklineObject{ChartData: "cd", Dock: "Fill"}
	sp.ReportComponentBase = *report.NewReportComponentBase()
	sp.SetName("mySpark")

	w := &countingWriter{}
	if err := sp.Serialize(w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// At minimum: Name + ChartData + Dock = 3 WriteStr calls.
	if w.strCount < 3 {
		t.Errorf("expected at least 3 WriteStr calls, got %d", w.strCount)
	}
}

// ── AdvMatrixObject.Serialize dead-code guard ─────────────────────────────────
// Same pattern: the `return err` after ReportComponentBase.Serialize in
// AdvMatrixObject.Serialize (line 113) is dead code. We confirm the path
// with a counting writer to show ReportComponentBase.Serialize ran.

func TestAdvMatrixObject_Serialize_ParentSerializeRuns(t *testing.T) {
	a := NewAdvMatrixObject()
	a.SetName("myMatrix")
	a.DataSource = "DS1"

	w := &countingWriter{}
	if err := a.Serialize(w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// At minimum: Name + DataSource = 2 WriteStr calls.
	if w.strCount < 2 {
		t.Errorf("expected at least 2 WriteStr calls, got %d", w.strCount)
	}
}

// ── AdvMatrixObject.Deserialize dead-code guard ───────────────────────────────
// The `return err` after ReportComponentBase.Deserialize (line 369) is dead
// code. Confirm that Deserialize with an all-defaults reader returns nil and
// reads DataSource correctly.

type dataSourceReader struct {
	alwaysDefaultReader
}

func (r *dataSourceReader) ReadStr(name, def string) string {
	if name == "DataSource" {
		return "TestDS"
	}
	return def
}

func TestAdvMatrixObject_Deserialize_WithDataSource_SVG(t *testing.T) {
	a := NewAdvMatrixObject()
	r := &dataSourceReader{}
	if err := a.Deserialize(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.DataSource != "TestDS" {
		t.Errorf("DataSource: got %q, want TestDS", a.DataSource)
	}
}

// sentinel to avoid "errors imported and not used" if tests above are adjusted
var _ = errors.New
