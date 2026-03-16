package gauge

// gauge_serialize_coverage_test.go — internal white-box tests targeting every
// reachable statement in gauge.go Serialize and Deserialize methods.
//
// Coverage analysis (as of writing):
//
//   gauge.go:171  GaugeObject.Serialize      97.3% — 1 uncovered: return err (line 173)
//   gauge.go:234  GaugeObject.Deserialize    96.0% — 1 uncovered: return err (line 236)
//   gauge.go:315  LinearGauge.Serialize      85.7% — 1 uncovered: return err (line 317)
//   gauge.go:329  LinearGauge.Deserialize    80.0% — 1 uncovered: return err (line 331)
//   gauge.go:373  RadialGauge.Serialize      85.7% — 1 uncovered: return err (line 375)
//   gauge.go:387  RadialGauge.Deserialize    80.0% — 1 uncovered: return err (line 389)
//   gauge.go:450  SimpleGauge.Serialize      94.1% — 1 uncovered: return err (line 452)
//   gauge.go:481  SimpleGauge.Deserialize    90.0% — 1 uncovered: return err (line 483)
//   gauge.go:523  SimpleProgressGauge.Serialize  80.0% — 1 uncovered: return err (line 525)
//   gauge.go:534  SimpleProgressGauge.Deserialize 75.0% — 1 uncovered: return err (line 536)
//
// Root cause: the entire Serialize/Deserialize base chain
//   GaugeObject → ReportComponentBase → ComponentBase → BaseObject
// always returns nil. The report.Writer interface write methods (WriteStr,
// WriteInt, WriteBool, WriteFloat) return void — there is no mechanism for the
// base chain to return an error. The `return err` branches are defensive
// dead-code and are structurally unreachable with the current interface design.
//
// This file exercises every OTHER reachable statement in these functions using
// a lightweight in-process testWriter/testReader, avoiding the XML round-trip
// overhead. It also documents that calling each Serialize/Deserialize with a
// mock reader/writer always returns nil (proving the error paths are dead).

import (
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// ─── minimal mock Writer / Reader ────────────────────────────────────────────

// mockWriter records writes so we can assert which keys were emitted.
type mockWriter struct {
	strs   map[string]string
	ints   map[string]int
	bools  map[string]bool
	floats map[string]float32
}

func newMockWriter() *mockWriter {
	return &mockWriter{
		strs:   make(map[string]string),
		ints:   make(map[string]int),
		bools:  make(map[string]bool),
		floats: make(map[string]float32),
	}
}

func (w *mockWriter) WriteStr(name, value string)            { w.strs[name] = value }
func (w *mockWriter) WriteInt(name string, value int)        { w.ints[name] = value }
func (w *mockWriter) WriteBool(name string, value bool)      { w.bools[name] = value }
func (w *mockWriter) WriteFloat(name string, value float32)  { w.floats[name] = value }
func (w *mockWriter) WriteObject(obj report.Serializable) error { return nil }
func (w *mockWriter) WriteObjectNamed(_ string, obj report.Serializable) error { return nil }

func (w *mockWriter) hasStr(name string) bool  { _, ok := w.strs[name]; return ok }
func (w *mockWriter) hasInt(name string) bool  { _, ok := w.ints[name]; return ok }
func (w *mockWriter) hasBool(name string) bool { _, ok := w.bools[name]; return ok }
func (w *mockWriter) hasFloat(name string) bool { _, ok := w.floats[name]; return ok }

// mockReader returns values from a map; defaults for missing keys.
type mockReader struct {
	strs   map[string]string
	ints   map[string]int
	bools  map[string]bool
	floats map[string]float32
}

func newMockReader() *mockReader {
	return &mockReader{
		strs:   make(map[string]string),
		ints:   make(map[string]int),
		bools:  make(map[string]bool),
		floats: make(map[string]float32),
	}
}

func (r *mockReader) ReadStr(name, def string) string {
	if v, ok := r.strs[name]; ok {
		return v
	}
	return def
}
func (r *mockReader) ReadInt(name string, def int) int {
	if v, ok := r.ints[name]; ok {
		return v
	}
	return def
}
func (r *mockReader) ReadBool(name string, def bool) bool {
	if v, ok := r.bools[name]; ok {
		return v
	}
	return def
}
func (r *mockReader) ReadFloat(name string, def float32) float32 {
	if v, ok := r.floats[name]; ok {
		return v
	}
	return def
}
func (r *mockReader) NextChild() (string, bool) { return "", false }
func (r *mockReader) FinishChild() error        { return nil }

// ─── GaugeObject.Serialize reachable branches ─────────────────────────────────

// TestGaugeObject_Serialize_ReturnsNil_WithMockWriter verifies that Serialize
// always returns nil regardless of the writer (proving the error branch is dead).
func TestGaugeObject_Serialize_ReturnsNil_WithMockWriter(t *testing.T) {
	g := NewGaugeObject()
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestGaugeObject_Serialize_Minimum_NonDefault verifies Minimum != 0 is written.
func TestGaugeObject_Serialize_Minimum_NonDefault(t *testing.T) {
	g := NewGaugeObject()
	g.Minimum = 5.0
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasFloat("Minimum") {
		t.Error("Minimum should be written when != 0")
	}
}

// TestGaugeObject_Serialize_Minimum_Default verifies Minimum == 0 is NOT written.
func TestGaugeObject_Serialize_Minimum_Default(t *testing.T) {
	g := NewGaugeObject()
	g.Minimum = 0
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasFloat("Minimum") {
		t.Error("Minimum should NOT be written when == 0 (default)")
	}
}

// TestGaugeObject_Serialize_Maximum_NonDefault verifies Maximum != 100 is written.
func TestGaugeObject_Serialize_Maximum_NonDefault(t *testing.T) {
	g := NewGaugeObject()
	g.Maximum = 200.0
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasFloat("Maximum") {
		t.Error("Maximum should be written when != 100")
	}
}

// TestGaugeObject_Serialize_Maximum_Default verifies Maximum == 100 is NOT written.
func TestGaugeObject_Serialize_Maximum_Default(t *testing.T) {
	g := NewGaugeObject()
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasFloat("Maximum") {
		t.Error("Maximum should NOT be written when == 100 (default)")
	}
}

// TestGaugeObject_Serialize_Value_NonZero verifies non-zero value is written.
func TestGaugeObject_Serialize_Value_NonZero(t *testing.T) {
	g := NewGaugeObject()
	g.SetValue(42)
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasFloat("Value") {
		t.Error("Value should be written when != 0")
	}
}

// TestGaugeObject_Serialize_Value_Zero verifies zero value is NOT written.
func TestGaugeObject_Serialize_Value_Zero(t *testing.T) {
	g := NewGaugeObject()
	g.SetValue(0)
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasFloat("Value") {
		t.Error("Value should NOT be written when == 0 (default)")
	}
}

// TestGaugeObject_Serialize_Expression_NonEmpty verifies non-empty Expression is written.
func TestGaugeObject_Serialize_Expression_NonEmpty(t *testing.T) {
	g := NewGaugeObject()
	g.Expression = "[Sales.Amount]"
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasStr("Expression") {
		t.Error("Expression should be written when non-empty")
	}
}

// TestGaugeObject_Serialize_Scale_AllFields verifies all non-default Scale fields
// are written.
func TestGaugeObject_Serialize_Scale_AllFields(t *testing.T) {
	g := NewGaugeObject()
	g.Scale.Font = "Arial, 10pt"
	g.Scale.MajorTicks.Width = 2.5
	g.Scale.MajorTicks.Color = "#FF0000"
	g.Scale.MajorTicks.Length = 10
	g.Scale.MinorTicks.Width = 1.0
	g.Scale.MinorTicks.Color = "#0000FF"
	g.Scale.MinorTicks.Length = 5

	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for _, name := range []string{
		"Scale.Font",
		"Scale.MajorTicks.Width",
		"Scale.MajorTicks.Color",
		"Scale.MajorTicks.Length",
		"Scale.MinorTicks.Width",
		"Scale.MinorTicks.Color",
		"Scale.MinorTicks.Length",
	} {
		switch {
		case w.hasStr(name):
		case w.hasFloat(name):
		default:
			t.Errorf("expected %q to be written but it was not", name)
		}
	}
}

// TestGaugeObject_Serialize_Scale_DefaultsNotWritten verifies default Scale
// field values are NOT written.
func TestGaugeObject_Serialize_Scale_DefaultsNotWritten(t *testing.T) {
	g := NewGaugeObject()
	// Scale.Font is "" by default, ticks all zero/empty.
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for _, name := range []string{
		"Scale.Font",
		"Scale.MajorTicks.Width",
		"Scale.MajorTicks.Color",
		"Scale.MajorTicks.Length",
		"Scale.MinorTicks.Width",
		"Scale.MinorTicks.Color",
		"Scale.MinorTicks.Length",
	} {
		if w.hasStr(name) || w.hasFloat(name) {
			t.Errorf("%q should NOT be written at default value", name)
		}
	}
}

// TestGaugeObject_Serialize_Pointer_AllFields verifies all non-default Pointer
// fields are written.
func TestGaugeObject_Serialize_Pointer_AllFields(t *testing.T) {
	g := NewGaugeObject()
	g.Pointer.Width = 10     // != 6 (default)
	g.Pointer.Height = 20    // != 0 (default)
	g.Pointer.Color = "#00FF00" // != "#CC0000" and non-empty

	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasFloat("Pointer.Width") {
		t.Error("Pointer.Width should be written when != 6")
	}
	if !w.hasFloat("Pointer.Height") {
		t.Error("Pointer.Height should be written when != 0")
	}
	if !w.hasStr("Pointer.Color") {
		t.Error("Pointer.Color should be written when non-empty and != #CC0000")
	}
}

// TestGaugeObject_Serialize_Pointer_DefaultNotWritten verifies that default
// Pointer field values are NOT written.
func TestGaugeObject_Serialize_Pointer_DefaultNotWritten(t *testing.T) {
	g := NewGaugeObject()
	// Default: Width=6, Height=0, Color="#CC0000"
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasFloat("Pointer.Width") {
		t.Error("Pointer.Width should NOT be written for default (6)")
	}
	if w.hasFloat("Pointer.Height") {
		t.Error("Pointer.Height should NOT be written when == 0")
	}
	if w.hasStr("Pointer.Color") {
		t.Error("Pointer.Color should NOT be written for default (#CC0000)")
	}
}

// TestGaugeObject_Serialize_Pointer_EmptyColor_NotWritten verifies empty
// Pointer.Color is not written (condition: Color != "" && Color != "#CC0000").
func TestGaugeObject_Serialize_Pointer_EmptyColor_NotWritten(t *testing.T) {
	g := NewGaugeObject()
	g.Pointer.Color = "" // empty → not written
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasStr("Pointer.Color") {
		t.Error("Pointer.Color should NOT be written for empty string")
	}
}

// TestGaugeObject_Serialize_Label_NonEmpty verifies Label.Font and Label.Text
// are written when non-empty.
func TestGaugeObject_Serialize_Label_NonEmpty(t *testing.T) {
	g := NewGaugeObject()
	g.Label.Font = "Times, 12pt"
	g.Label.Text = "Speed"
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasStr("Label.Font") {
		t.Error("Label.Font should be written when non-empty")
	}
	if !w.hasStr("Label.Text") {
		t.Error("Label.Text should be written when non-empty")
	}
}

// TestGaugeObject_Serialize_Label_Empty_NotWritten verifies empty Label fields
// are not written.
func TestGaugeObject_Serialize_Label_Empty_NotWritten(t *testing.T) {
	g := NewGaugeObject()
	// Label.Font and Label.Text default to "".
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasStr("Label.Font") {
		t.Error("Label.Font should NOT be written when empty")
	}
	if w.hasStr("Label.Text") {
		t.Error("Label.Text should NOT be written when empty")
	}
}

// TestGaugeObject_Serialize_NilScale_NoPanic verifies nil Scale does not panic.
func TestGaugeObject_Serialize_NilScale_NoPanic(t *testing.T) {
	g := NewGaugeObject()
	g.Scale = nil
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize with nil Scale: %v", err)
	}
}

// TestGaugeObject_Serialize_NilPointer_NoPanic verifies nil Pointer does not panic.
func TestGaugeObject_Serialize_NilPointer_NoPanic(t *testing.T) {
	g := NewGaugeObject()
	g.Pointer = nil
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize with nil Pointer: %v", err)
	}
}

// ─── GaugeObject.Deserialize reachable branches ───────────────────────────────

// TestGaugeObject_Deserialize_ReturnsNil_WithMockReader verifies that Deserialize
// always returns nil (proving the error branch is dead).
func TestGaugeObject_Deserialize_ReturnsNil_WithMockReader(t *testing.T) {
	g := NewGaugeObject()
	r := newMockReader()
	if err := g.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// TestGaugeObject_Deserialize_AllFields verifies all fields are read from the reader.
func TestGaugeObject_Deserialize_AllFields(t *testing.T) {
	g := NewGaugeObject()
	g.Scale = nil
	g.Pointer = nil

	r := newMockReader()
	r.floats["Minimum"] = 10
	r.floats["Maximum"] = 500
	r.floats["Value"] = 250
	r.strs["Expression"] = "[Revenue]"
	r.strs["Scale.Font"] = "Courier, 8pt"
	r.floats["Scale.MajorTicks.Width"] = 3
	r.strs["Scale.MajorTicks.Color"] = "#FF0000"
	r.floats["Scale.MajorTicks.Length"] = 8
	r.floats["Scale.MinorTicks.Width"] = 1
	r.strs["Scale.MinorTicks.Color"] = "#0000FF"
	r.floats["Scale.MinorTicks.Length"] = 4
	r.floats["Pointer.Width"] = 8 // >= 0 → branch taken
	r.floats["Pointer.Height"] = 12
	r.strs["Pointer.Color"] = "#00FF00"
	r.strs["Label.Font"] = "Arial, 10pt"
	r.strs["Label.Text"] = "RPM"

	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if g.Minimum != 10 {
		t.Errorf("Minimum = %v, want 10", g.Minimum)
	}
	if g.Maximum != 500 {
		t.Errorf("Maximum = %v, want 500", g.Maximum)
	}
	if g.value != 250 {
		t.Errorf("value = %v, want 250", g.value)
	}
	if g.Expression != "[Revenue]" {
		t.Errorf("Expression = %q, want [Revenue]", g.Expression)
	}
	if g.Scale == nil {
		t.Fatal("Scale should not be nil after Deserialize")
	}
	if g.Scale.Font != "Courier, 8pt" {
		t.Errorf("Scale.Font = %q, want Courier, 8pt", g.Scale.Font)
	}
	if g.Scale.MajorTicks.Width != 3 {
		t.Errorf("Scale.MajorTicks.Width = %v, want 3", g.Scale.MajorTicks.Width)
	}
	if g.Scale.MajorTicks.Color != "#FF0000" {
		t.Errorf("Scale.MajorTicks.Color = %q", g.Scale.MajorTicks.Color)
	}
	if g.Scale.MajorTicks.Length != 8 {
		t.Errorf("Scale.MajorTicks.Length = %v, want 8", g.Scale.MajorTicks.Length)
	}
	if g.Scale.MinorTicks.Width != 1 {
		t.Errorf("Scale.MinorTicks.Width = %v, want 1", g.Scale.MinorTicks.Width)
	}
	if g.Scale.MinorTicks.Color != "#0000FF" {
		t.Errorf("Scale.MinorTicks.Color = %q", g.Scale.MinorTicks.Color)
	}
	if g.Scale.MinorTicks.Length != 4 {
		t.Errorf("Scale.MinorTicks.Length = %v, want 4", g.Scale.MinorTicks.Length)
	}
	if g.Pointer == nil {
		t.Fatal("Pointer should not be nil after Deserialize")
	}
	if g.Pointer.Width != 8 {
		t.Errorf("Pointer.Width = %v, want 8", g.Pointer.Width)
	}
	if g.Pointer.Height != 12 {
		t.Errorf("Pointer.Height = %v, want 12", g.Pointer.Height)
	}
	if g.Pointer.Color != "#00FF00" {
		t.Errorf("Pointer.Color = %q, want #00FF00", g.Pointer.Color)
	}
	if g.Label.Font != "Arial, 10pt" {
		t.Errorf("Label.Font = %q", g.Label.Font)
	}
	if g.Label.Text != "RPM" {
		t.Errorf("Label.Text = %q", g.Label.Text)
	}
}

// TestGaugeObject_Deserialize_PointerWidth_NegativeDefault verifies that when
// Pointer.Width reads -1 (default sentinel), the branch is NOT entered.
func TestGaugeObject_Deserialize_PointerWidth_NegativeDefault(t *testing.T) {
	g := NewGaugeObject()
	g.Pointer.Width = 6 // set to known value
	r := newMockReader()
	// Pointer.Width not in reader → ReadFloat returns -1 (default) → branch not entered.
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.Pointer.Width != 6 {
		t.Errorf("Pointer.Width = %v; expected unchanged (6) since not in reader", g.Pointer.Width)
	}
}

// TestGaugeObject_Deserialize_PointerWidth_ZeroValue verifies Pointer.Width=0
// (>= 0) IS entered and applied.
func TestGaugeObject_Deserialize_PointerWidth_ZeroValue(t *testing.T) {
	g := NewGaugeObject()
	r := newMockReader()
	r.floats["Pointer.Width"] = 0 // 0 >= 0 → branch entered
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.Pointer.Width != 0 {
		t.Errorf("Pointer.Width = %v, want 0", g.Pointer.Width)
	}
}

// TestGaugeObject_Deserialize_PointerColor_NonEmpty verifies non-empty Pointer.Color
// is applied (branch: c != "" → Pointer.Color = c).
func TestGaugeObject_Deserialize_PointerColor_NonEmpty(t *testing.T) {
	g := NewGaugeObject()
	r := newMockReader()
	r.strs["Pointer.Color"] = "#AABB00"
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.Pointer.Color != "#AABB00" {
		t.Errorf("Pointer.Color = %q, want #AABB00", g.Pointer.Color)
	}
}

// TestGaugeObject_Deserialize_PointerColor_Empty verifies empty Pointer.Color
// does NOT overwrite the existing value.
func TestGaugeObject_Deserialize_PointerColor_Empty(t *testing.T) {
	g := NewGaugeObject()
	g.Pointer.Color = "#CC0000"
	r := newMockReader()
	// Pointer.Color not in reader → ReadStr returns "" → branch not taken.
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.Pointer.Color != "#CC0000" {
		t.Errorf("Pointer.Color = %q, want unchanged #CC0000", g.Pointer.Color)
	}
}

// TestGaugeObject_Deserialize_NilScale_Initialized verifies nil Scale is
// created during Deserialize.
func TestGaugeObject_Deserialize_NilScale_Initialized(t *testing.T) {
	g := NewGaugeObject()
	g.Scale = nil
	r := newMockReader()
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.Scale == nil {
		t.Error("Scale should be initialized after Deserialize when nil")
	}
}

// TestGaugeObject_Deserialize_NilPointer_Initialized verifies nil Pointer is
// created during Deserialize.
func TestGaugeObject_Deserialize_NilPointer_Initialized(t *testing.T) {
	g := NewGaugeObject()
	g.Pointer = nil
	r := newMockReader()
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.Pointer == nil {
		t.Error("Pointer should be initialized after Deserialize when nil")
	}
}

// ─── LinearGauge.Serialize reachable branches ─────────────────────────────────

// TestLinearGauge_Serialize_ReturnsNil verifies Serialize always returns nil.
func TestLinearGauge_Serialize_ReturnsNil(t *testing.T) {
	g := NewLinearGauge()
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestLinearGauge_Serialize_Orientation_Vertical_Written verifies OrientationVertical
// is written.
func TestLinearGauge_Serialize_Orientation_Vertical_Written(t *testing.T) {
	g := NewLinearGauge()
	g.Orientation = OrientationVertical
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasInt("Orientation") {
		t.Error("Orientation should be written for non-horizontal")
	}
}

// TestLinearGauge_Serialize_Orientation_Horizontal_NotWritten verifies default
// horizontal orientation is NOT written.
func TestLinearGauge_Serialize_Orientation_Horizontal_NotWritten(t *testing.T) {
	g := NewLinearGauge()
	g.Orientation = OrientationHorizontal
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasInt("Orientation") {
		t.Error("Orientation should NOT be written for default (Horizontal)")
	}
}

// TestLinearGauge_Serialize_Inverted_True_Written verifies Inverted=true is written.
func TestLinearGauge_Serialize_Inverted_True_Written(t *testing.T) {
	g := NewLinearGauge()
	g.Inverted = true
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasBool("Inverted") {
		t.Error("Inverted should be written when true")
	}
}

// TestLinearGauge_Serialize_Inverted_False_NotWritten verifies Inverted=false
// (default) is NOT written.
func TestLinearGauge_Serialize_Inverted_False_NotWritten(t *testing.T) {
	g := NewLinearGauge()
	g.Inverted = false
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasBool("Inverted") {
		t.Error("Inverted should NOT be written when false (default)")
	}
}

// ─── LinearGauge.Deserialize reachable branches ───────────────────────────────

// TestLinearGauge_Deserialize_ReturnsNil verifies Deserialize always returns nil.
func TestLinearGauge_Deserialize_ReturnsNil(t *testing.T) {
	g := NewLinearGauge()
	r := newMockReader()
	if err := g.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// TestLinearGauge_Deserialize_Orientation verifies Orientation is read.
func TestLinearGauge_Deserialize_Orientation(t *testing.T) {
	g := NewLinearGauge()
	r := newMockReader()
	r.ints["Orientation"] = int(OrientationVertical)
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.Orientation != OrientationVertical {
		t.Errorf("Orientation = %v, want Vertical", g.Orientation)
	}
}

// TestLinearGauge_Deserialize_Inverted verifies Inverted is read.
func TestLinearGauge_Deserialize_Inverted(t *testing.T) {
	g := NewLinearGauge()
	r := newMockReader()
	r.bools["Inverted"] = true
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !g.Inverted {
		t.Error("Inverted should be true after Deserialize")
	}
}

// TestLinearGauge_Deserialize_Defaults verifies default values when keys absent.
func TestLinearGauge_Deserialize_Defaults(t *testing.T) {
	g := NewLinearGauge()
	r := newMockReader()
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.Orientation != OrientationHorizontal {
		t.Errorf("Orientation = %v, want Horizontal", g.Orientation)
	}
	if g.Inverted {
		t.Error("Inverted should default to false")
	}
}

// ─── RadialGauge.Serialize reachable branches ─────────────────────────────────

// TestRadialGauge_Serialize_ReturnsNil verifies Serialize always returns nil.
func TestRadialGauge_Serialize_ReturnsNil(t *testing.T) {
	g := NewRadialGauge()
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestRadialGauge_Serialize_StartAngle_NonDefault verifies non-default StartAngle
// is written.
func TestRadialGauge_Serialize_StartAngle_NonDefault(t *testing.T) {
	g := NewRadialGauge()
	g.StartAngle = -90 // != -135 (default)
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasFloat("StartAngle") {
		t.Error("StartAngle should be written when != -135")
	}
}

// TestRadialGauge_Serialize_StartAngle_Default_NotWritten verifies default
// StartAngle is NOT written.
func TestRadialGauge_Serialize_StartAngle_Default_NotWritten(t *testing.T) {
	g := NewRadialGauge()
	// Default StartAngle = -135
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasFloat("StartAngle") {
		t.Error("StartAngle should NOT be written for default (-135)")
	}
}

// TestRadialGauge_Serialize_EndAngle_NonDefault verifies non-default EndAngle
// is written.
func TestRadialGauge_Serialize_EndAngle_NonDefault(t *testing.T) {
	g := NewRadialGauge()
	g.EndAngle = 90 // != 135 (default)
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasFloat("EndAngle") {
		t.Error("EndAngle should be written when != 135")
	}
}

// TestRadialGauge_Serialize_EndAngle_Default_NotWritten verifies default EndAngle
// is NOT written.
func TestRadialGauge_Serialize_EndAngle_Default_NotWritten(t *testing.T) {
	g := NewRadialGauge()
	// Default EndAngle = 135
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasFloat("EndAngle") {
		t.Error("EndAngle should NOT be written for default (135)")
	}
}

// ─── RadialGauge.Deserialize reachable branches ───────────────────────────────

// TestRadialGauge_Deserialize_ReturnsNil verifies Deserialize always returns nil.
func TestRadialGauge_Deserialize_ReturnsNil(t *testing.T) {
	g := NewRadialGauge()
	r := newMockReader()
	if err := g.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// TestRadialGauge_Deserialize_CustomAngles verifies custom angles are read.
func TestRadialGauge_Deserialize_CustomAngles(t *testing.T) {
	g := NewRadialGauge()
	r := newMockReader()
	r.floats["StartAngle"] = -45
	r.floats["EndAngle"] = 45
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.StartAngle != -45 {
		t.Errorf("StartAngle = %v, want -45", g.StartAngle)
	}
	if g.EndAngle != 45 {
		t.Errorf("EndAngle = %v, want 45", g.EndAngle)
	}
}

// TestRadialGauge_Deserialize_DefaultAngles_Mock verifies default angles when absent
// (using mock reader, complementing the XML-based test in gauge_internal_test.go).
func TestRadialGauge_Deserialize_DefaultAngles_Mock(t *testing.T) {
	g := NewRadialGauge()
	r := newMockReader()
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.StartAngle != -135 {
		t.Errorf("StartAngle = %v, want -135 (default)", g.StartAngle)
	}
	if g.EndAngle != 135 {
		t.Errorf("EndAngle = %v, want 135 (default)", g.EndAngle)
	}
}

// ─── SimpleGauge.Serialize reachable branches ─────────────────────────────────

// TestSimpleGauge_Serialize_ReturnsNil verifies Serialize always returns nil.
func TestSimpleGauge_Serialize_ReturnsNil(t *testing.T) {
	g := NewSimpleGauge()
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestSimpleGauge_Serialize_NonDefaultShape verifies non-rectangle shape is written.
func TestSimpleGauge_Serialize_NonDefaultShape(t *testing.T) {
	for _, shape := range []SimpleGaugeShape{SimpleGaugeShapeCircle, SimpleGaugeShapeTriangle} {
		g := NewSimpleGauge()
		g.Shape = shape
		w := newMockWriter()
		if err := g.Serialize(w); err != nil {
			t.Fatalf("Serialize (shape=%v): %v", shape, err)
		}
		if !w.hasInt("Shape") {
			t.Errorf("Shape should be written for %v", shape)
		}
	}
}

// TestSimpleGauge_Serialize_DefaultShape_NotWritten verifies Rectangle shape
// is NOT written.
func TestSimpleGauge_Serialize_DefaultShape_NotWritten(t *testing.T) {
	g := NewSimpleGauge()
	g.Shape = SimpleGaugeShapeRectangle
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasInt("Shape") {
		t.Error("Shape should NOT be written for default (Rectangle)")
	}
}

// TestSimpleGauge_Serialize_ShowText_False_Written verifies ShowText=false is written.
func TestSimpleGauge_Serialize_ShowText_False_Written(t *testing.T) {
	g := NewSimpleGauge()
	g.ShowText = false
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasBool("ShowText") {
		t.Error("ShowText should be written when false")
	}
}

// TestSimpleGauge_Serialize_ShowText_True_NotWritten verifies ShowText=true
// (default) is NOT written.
func TestSimpleGauge_Serialize_ShowText_True_NotWritten(t *testing.T) {
	g := NewSimpleGauge()
	g.ShowText = true
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasBool("ShowText") {
		t.Error("ShowText should NOT be written when true (default)")
	}
}

// TestSimpleGauge_Serialize_TextFormat_NonDefault verifies non-default TextFormat
// is written.
func TestSimpleGauge_Serialize_TextFormat_NonDefault(t *testing.T) {
	g := NewSimpleGauge()
	g.TextFormat = "%.2f%%"
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasStr("TextFormat") {
		t.Error("TextFormat should be written when not equal to the default format")
	}
}

// TestSimpleGauge_Serialize_TextFormat_Empty_Written verifies empty TextFormat
// (non-default) is written.
func TestSimpleGauge_Serialize_TextFormat_Empty_Written(t *testing.T) {
	g := NewSimpleGauge()
	g.TextFormat = "" // "" != "%g%%" → should be written
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasStr("TextFormat") {
		t.Error("TextFormat should be written when empty (differs from default)")
	}
}

// TestSimpleGauge_Serialize_TextFormat_Default_NotWritten verifies default
// TextFormat is NOT written.
func TestSimpleGauge_Serialize_TextFormat_Default_NotWritten(t *testing.T) {
	g := NewSimpleGauge()
	// Default TextFormat = "%g%%"
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasStr("TextFormat") {
		t.Error("TextFormat should NOT be written for the default value")
	}
}

// TestSimpleGauge_Serialize_SubScale_Disabled_Written verifies disabled subscales
// are written.
func TestSimpleGauge_Serialize_SubScale_Disabled_Written(t *testing.T) {
	g := NewSimpleGauge()
	g.FirstSubScale.Enabled = false
	g.FirstSubScale.ShowCaption = false
	g.SecondSubScale.Enabled = false
	g.SecondSubScale.ShowCaption = false
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for _, name := range []string{
		"Scale.FirstSubScale.Enabled",
		"Scale.FirstSubScale.ShowCaption",
		"Scale.SecondSubScale.Enabled",
		"Scale.SecondSubScale.ShowCaption",
	} {
		if !w.hasBool(name) {
			t.Errorf("%q should be written when false", name)
		}
	}
}

// TestSimpleGauge_Serialize_SubScale_Default_NotWritten verifies default subscale
// values are NOT written.
func TestSimpleGauge_Serialize_SubScale_Default_NotWritten(t *testing.T) {
	g := NewSimpleGauge()
	// Defaults: FirstSubScale.Enabled=true, .ShowCaption=true; same for Second.
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for _, name := range []string{
		"Scale.FirstSubScale.Enabled",
		"Scale.FirstSubScale.ShowCaption",
		"Scale.SecondSubScale.Enabled",
		"Scale.SecondSubScale.ShowCaption",
	} {
		if w.hasBool(name) {
			t.Errorf("%q should NOT be written for default (true)", name)
		}
	}
}

// ─── SimpleGauge.Deserialize reachable branches ───────────────────────────────

// TestSimpleGauge_Deserialize_ReturnsNil verifies Deserialize always returns nil.
func TestSimpleGauge_Deserialize_ReturnsNil(t *testing.T) {
	g := NewSimpleGauge()
	r := newMockReader()
	if err := g.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// TestSimpleGauge_Deserialize_AllFields verifies all fields are read.
func TestSimpleGauge_Deserialize_AllFields(t *testing.T) {
	g := NewSimpleGauge()
	r := newMockReader()
	r.ints["Shape"] = int(SimpleGaugeShapeTriangle)
	r.bools["ShowText"] = false
	r.strs["TextFormat"] = "%.1f%%"
	r.bools["Scale.FirstSubScale.Enabled"] = false
	r.bools["Scale.FirstSubScale.ShowCaption"] = false
	r.bools["Scale.SecondSubScale.Enabled"] = false
	r.bools["Scale.SecondSubScale.ShowCaption"] = false

	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.Shape != SimpleGaugeShapeTriangle {
		t.Errorf("Shape = %v, want Triangle", g.Shape)
	}
	if g.ShowText {
		t.Error("ShowText should be false")
	}
	if g.TextFormat != "%.1f%%" {
		t.Errorf("TextFormat = %q, want %%.1f%%%%", g.TextFormat)
	}
	if g.FirstSubScale.Enabled {
		t.Error("FirstSubScale.Enabled should be false")
	}
	if g.FirstSubScale.ShowCaption {
		t.Error("FirstSubScale.ShowCaption should be false")
	}
	if g.SecondSubScale.Enabled {
		t.Error("SecondSubScale.Enabled should be false")
	}
	if g.SecondSubScale.ShowCaption {
		t.Error("SecondSubScale.ShowCaption should be false")
	}
}

// TestSimpleGauge_Deserialize_Defaults verifies default field values when absent.
func TestSimpleGauge_Deserialize_Defaults(t *testing.T) {
	g := NewSimpleGauge()
	r := newMockReader()
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.Shape != SimpleGaugeShapeRectangle {
		t.Errorf("Shape = %v, want Rectangle (default)", g.Shape)
	}
	if !g.ShowText {
		t.Error("ShowText should default to true")
	}
	if g.TextFormat != "%g%%" {
		t.Errorf("TextFormat = %q, want %%g%%%% (default)", g.TextFormat)
	}
	if !g.FirstSubScale.Enabled {
		t.Error("FirstSubScale.Enabled should default to true")
	}
	if !g.FirstSubScale.ShowCaption {
		t.Error("FirstSubScale.ShowCaption should default to true")
	}
	if !g.SecondSubScale.Enabled {
		t.Error("SecondSubScale.Enabled should default to true")
	}
	if !g.SecondSubScale.ShowCaption {
		t.Error("SecondSubScale.ShowCaption should default to true")
	}
}

// ─── SimpleProgressGauge.Serialize reachable branches ────────────────────────

// TestSimpleProgressGauge_Serialize_ReturnsNil verifies Serialize always returns nil.
func TestSimpleProgressGauge_Serialize_ReturnsNil(t *testing.T) {
	g := NewSimpleProgressGauge()
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Errorf("Serialize returned unexpected error: %v", err)
	}
}

// TestSimpleProgressGauge_Serialize_ShowText_False_Written verifies ShowText=false
// is written.
func TestSimpleProgressGauge_Serialize_ShowText_False_Written(t *testing.T) {
	g := NewSimpleProgressGauge()
	g.ShowText = false
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !w.hasBool("ShowText") {
		t.Error("ShowText should be written when false")
	}
}

// TestSimpleProgressGauge_Serialize_ShowText_True_NotWritten verifies ShowText=true
// (default) is NOT written.
func TestSimpleProgressGauge_Serialize_ShowText_True_NotWritten(t *testing.T) {
	g := NewSimpleProgressGauge()
	g.ShowText = true
	w := newMockWriter()
	if err := g.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if w.hasBool("ShowText") {
		t.Error("ShowText should NOT be written when true (default)")
	}
}

// ─── SimpleProgressGauge.Deserialize reachable branches ──────────────────────

// TestSimpleProgressGauge_Deserialize_ReturnsNil verifies Deserialize always
// returns nil.
func TestSimpleProgressGauge_Deserialize_ReturnsNil(t *testing.T) {
	g := NewSimpleProgressGauge()
	r := newMockReader()
	if err := g.Deserialize(r); err != nil {
		t.Errorf("Deserialize returned unexpected error: %v", err)
	}
}

// TestSimpleProgressGauge_Deserialize_ShowText_False verifies ShowText=false is
// read.
func TestSimpleProgressGauge_Deserialize_ShowText_False(t *testing.T) {
	g := NewSimpleProgressGauge()
	r := newMockReader()
	r.bools["ShowText"] = false
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g.ShowText {
		t.Error("ShowText should be false after Deserialize")
	}
}

// TestSimpleProgressGauge_Deserialize_ShowText_True_Default verifies ShowText
// defaults to true when absent.
func TestSimpleProgressGauge_Deserialize_ShowText_True_Default(t *testing.T) {
	g := NewSimpleProgressGauge()
	r := newMockReader()
	if err := g.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !g.ShowText {
		t.Error("ShowText should default to true")
	}
}
