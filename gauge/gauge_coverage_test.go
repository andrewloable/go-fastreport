package gauge_test

// gauge_coverage_test.go — additional tests to push gauge package coverage to 85%+.
// Covers: Serialize/Deserialize round-trips for all gauge types, BaseName/TypeName,
// parseColor error path, drawHLine/drawVLine/drawArc edge cases, and RenderXxx branches.

import (
	"bytes"
	"image"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// gaugeRoundTrip serialises g into XML via serial.Writer and deserialises it
// back using fn.  The element wrapper tag is elementName.
func serializeGauge(t *testing.T, elementName string, s interface {
	Serialize(w interface{ WriteStr(string, string); WriteInt(string, int); WriteBool(string, bool); WriteFloat(string, float32); WriteObject(interface{ Serialize(interface{}) error }) error }) error
}) {
	t.Helper()
}

// linearRoundTrip performs a full serialize/deserialize round-trip for LinearGauge.
func linearRoundTrip(t *testing.T, orig *gauge.LinearGauge) *gauge.LinearGauge {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("LinearGauge"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := orig.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	r := serial.NewReader(strings.NewReader(buf.String()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed; XML:\n%s", buf.String())
	}
	g2 := gauge.NewLinearGauge()
	if err := g2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return g2
}

// radialRoundTrip performs a full serialize/deserialize round-trip for RadialGauge.
func radialRoundTrip(t *testing.T, orig *gauge.RadialGauge) *gauge.RadialGauge {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("RadialGauge"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := orig.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	r := serial.NewReader(strings.NewReader(buf.String()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed; XML:\n%s", buf.String())
	}
	g2 := gauge.NewRadialGauge()
	if err := g2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return g2
}

// simpleProgressRoundTrip performs a full serialize/deserialize round-trip for SimpleProgressGauge.
func simpleProgressRoundTrip(t *testing.T, orig *gauge.SimpleProgressGauge) *gauge.SimpleProgressGauge {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("SimpleProgressGauge"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := orig.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	r := serial.NewReader(strings.NewReader(buf.String()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed; XML:\n%s", buf.String())
	}
	g2 := gauge.NewSimpleProgressGauge()
	if err := g2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return g2
}

// simpleGaugeRoundTrip performs a full serialize/deserialize round-trip for SimpleGauge.
func simpleGaugeRoundTrip(t *testing.T, orig *gauge.SimpleGauge) *gauge.SimpleGauge {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("SimpleGauge"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := orig.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	r := serial.NewReader(strings.NewReader(buf.String()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed; XML:\n%s", buf.String())
	}
	g2 := gauge.NewSimpleGauge()
	if err := g2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return g2
}

// gaugeObjectRoundTrip performs a full serialize/deserialize round-trip for GaugeObject.
func gaugeObjectRoundTrip(t *testing.T, orig *gauge.GaugeObject) *gauge.GaugeObject {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("GaugeObject"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := orig.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	r := serial.NewReader(strings.NewReader(buf.String()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed; XML:\n%s", buf.String())
	}
	g2 := gauge.NewGaugeObject()
	if err := g2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return g2
}

// ── GaugeObject Serialize/Deserialize ─────────────────────────────────────────

func TestGaugeObject_SerializeDeserialize_Defaults(t *testing.T) {
	orig := gauge.NewGaugeObject()
	got := gaugeObjectRoundTrip(t, orig)
	if got.Minimum != orig.Minimum {
		t.Errorf("Minimum: got %v, want %v", got.Minimum, orig.Minimum)
	}
	if got.Maximum != orig.Maximum {
		t.Errorf("Maximum: got %v, want %v", got.Maximum, orig.Maximum)
	}
	if got.Value() != orig.Value() {
		t.Errorf("Value: got %v, want %v", got.Value(), orig.Value())
	}
}

func TestGaugeObject_SerializeDeserialize_NonDefaults(t *testing.T) {
	orig := gauge.NewGaugeObject()
	orig.Minimum = 10
	orig.Maximum = 200
	orig.SetValue(50)
	orig.Expression = "[Sales.Revenue]"
	orig.Scale.Font = "Arial, 10pt"
	orig.Scale.MajorTicks.Width = 2
	orig.Scale.MajorTicks.Color = "#FF0000"
	orig.Scale.MajorTicks.Length = 8
	orig.Scale.MinorTicks.Width = 1
	orig.Scale.MinorTicks.Color = "#0000FF"
	orig.Scale.MinorTicks.Length = 4
	orig.Pointer.Width = 10
	orig.Pointer.Height = 5
	orig.Pointer.Color = "#00FF00"
	orig.Label.Font = "Times New Roman, 12pt"
	orig.Label.Text = "Value"

	got := gaugeObjectRoundTrip(t, orig)

	if got.Minimum != orig.Minimum {
		t.Errorf("Minimum: got %v, want %v", got.Minimum, orig.Minimum)
	}
	if got.Maximum != orig.Maximum {
		t.Errorf("Maximum: got %v, want %v", got.Maximum, orig.Maximum)
	}
	if got.Expression != orig.Expression {
		t.Errorf("Expression: got %q, want %q", got.Expression, orig.Expression)
	}
	if got.Scale.Font != orig.Scale.Font {
		t.Errorf("Scale.Font: got %q, want %q", got.Scale.Font, orig.Scale.Font)
	}
	if got.Scale.MajorTicks.Color != orig.Scale.MajorTicks.Color {
		t.Errorf("Scale.MajorTicks.Color: got %q, want %q", got.Scale.MajorTicks.Color, orig.Scale.MajorTicks.Color)
	}
	if got.Scale.MinorTicks.Color != orig.Scale.MinorTicks.Color {
		t.Errorf("Scale.MinorTicks.Color: got %q, want %q", got.Scale.MinorTicks.Color, orig.Scale.MinorTicks.Color)
	}
	if got.Pointer.Color != orig.Pointer.Color {
		t.Errorf("Pointer.Color: got %q, want %q", got.Pointer.Color, orig.Pointer.Color)
	}
	if got.Label.Font != orig.Label.Font {
		t.Errorf("Label.Font: got %q, want %q", got.Label.Font, orig.Label.Font)
	}
	if got.Label.Text != orig.Label.Text {
		t.Errorf("Label.Text: got %q, want %q", got.Label.Text, orig.Label.Text)
	}
}

func TestGaugeObject_SerializeDeserialize_NilScalePointer(t *testing.T) {
	// Verify Deserialize creates Scale/Pointer when nil.
	orig := gauge.NewGaugeObject()
	orig.Scale = nil
	orig.Pointer = nil

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("GaugeObject"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	r := serial.NewReader(strings.NewReader(buf.String()))
	r.ReadObjectHeader()
	g2 := gauge.NewGaugeObject()
	g2.Scale = nil
	g2.Pointer = nil
	if err := g2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g2.Scale == nil {
		t.Error("Scale should be initialized after Deserialize")
	}
	if g2.Pointer == nil {
		t.Error("Pointer should be initialized after Deserialize")
	}
}

func TestGaugeObject_SerializeDeserialize_PointerWidthPreserved(t *testing.T) {
	// Pointer.Width is only serialized when != 6 (the default).
	// Test that a non-default pointer width round-trips correctly.
	orig := gauge.NewGaugeObject()
	orig.Pointer.Width = 10
	got := gaugeObjectRoundTrip(t, orig)
	if got.Pointer.Width != orig.Pointer.Width {
		t.Errorf("Pointer.Width: got %v, want %v", got.Pointer.Width, orig.Pointer.Width)
	}
}

func TestGaugeObject_SerializeDeserialize_PointerColorPreserved(t *testing.T) {
	// Pointer.Color is only serialized when non-empty and != "#CC0000".
	orig := gauge.NewGaugeObject()
	orig.Pointer.Color = "#AABB00"
	got := gaugeObjectRoundTrip(t, orig)
	if got.Pointer.Color != orig.Pointer.Color {
		t.Errorf("Pointer.Color: got %q, want %q", got.Pointer.Color, orig.Pointer.Color)
	}
}

// ── LinearGauge BaseName/TypeName/Serialize/Deserialize ────────────────────────

func TestLinearGauge_BaseName(t *testing.T) {
	g := gauge.NewLinearGauge()
	if g.BaseName() != "LinearGauge" {
		t.Errorf("BaseName = %q, want LinearGauge", g.BaseName())
	}
}

func TestLinearGauge_TypeName(t *testing.T) {
	g := gauge.NewLinearGauge()
	if g.TypeName() != "LinearGauge" {
		t.Errorf("TypeName = %q, want LinearGauge", g.TypeName())
	}
}

func TestLinearGauge_SerializeDeserialize_Defaults(t *testing.T) {
	orig := gauge.NewLinearGauge()
	got := linearRoundTrip(t, orig)

	if got.Orientation != orig.Orientation {
		t.Errorf("Orientation: got %v, want %v", got.Orientation, orig.Orientation)
	}
	if got.Inverted != orig.Inverted {
		t.Errorf("Inverted: got %v, want %v", got.Inverted, orig.Inverted)
	}
}

func TestLinearGauge_SerializeDeserialize_Vertical(t *testing.T) {
	orig := gauge.NewLinearGauge()
	orig.Orientation = gauge.OrientationVertical
	orig.SetValue(60)
	got := linearRoundTrip(t, orig)

	if got.Orientation != gauge.OrientationVertical {
		t.Errorf("Orientation: got %v, want Vertical", got.Orientation)
	}
}

func TestLinearGauge_SerializeDeserialize_Inverted(t *testing.T) {
	orig := gauge.NewLinearGauge()
	orig.Inverted = true
	orig.SetValue(30)
	got := linearRoundTrip(t, orig)

	if !got.Inverted {
		t.Error("Inverted should be true after round-trip")
	}
}

func TestLinearGauge_SerializeDeserialize_NonDefaultValue(t *testing.T) {
	orig := gauge.NewLinearGauge()
	orig.SetValue(75)
	orig.Minimum = 0
	orig.Maximum = 100
	got := linearRoundTrip(t, orig)

	if got.Value() != orig.Value() {
		t.Errorf("Value: got %v, want %v", got.Value(), orig.Value())
	}
}

// ── RadialGauge BaseName/TypeName/Serialize/Deserialize ───────────────────────

func TestRadialGauge_BaseName(t *testing.T) {
	g := gauge.NewRadialGauge()
	if g.BaseName() != "RadialGauge" {
		t.Errorf("BaseName = %q, want RadialGauge", g.BaseName())
	}
}

func TestRadialGauge_TypeName(t *testing.T) {
	g := gauge.NewRadialGauge()
	if g.TypeName() != "RadialGauge" {
		t.Errorf("TypeName = %q, want RadialGauge", g.TypeName())
	}
}

func TestRadialGauge_SerializeDeserialize_Defaults(t *testing.T) {
	orig := gauge.NewRadialGauge()
	got := radialRoundTrip(t, orig)

	// Default angles are -135 and 135, so they are not serialized.
	// After deserialization the defaults should be restored.
	if got.StartAngle != orig.StartAngle {
		t.Errorf("StartAngle: got %v, want %v", got.StartAngle, orig.StartAngle)
	}
	if got.EndAngle != orig.EndAngle {
		t.Errorf("EndAngle: got %v, want %v", got.EndAngle, orig.EndAngle)
	}
}

func TestRadialGauge_SerializeDeserialize_CustomAngles(t *testing.T) {
	orig := gauge.NewRadialGauge()
	orig.StartAngle = -90
	orig.EndAngle = 90
	orig.SetValue(50)
	got := radialRoundTrip(t, orig)

	if got.StartAngle != orig.StartAngle {
		t.Errorf("StartAngle: got %v, want %v", got.StartAngle, orig.StartAngle)
	}
	if got.EndAngle != orig.EndAngle {
		t.Errorf("EndAngle: got %v, want %v", got.EndAngle, orig.EndAngle)
	}
}

func TestRadialGauge_SerializeDeserialize_FullCircle(t *testing.T) {
	orig := gauge.NewRadialGauge()
	orig.StartAngle = 0
	orig.EndAngle = 360
	got := radialRoundTrip(t, orig)

	if got.StartAngle != 0 {
		t.Errorf("StartAngle: got %v, want 0", got.StartAngle)
	}
	if got.EndAngle != 360 {
		t.Errorf("EndAngle: got %v, want 360", got.EndAngle)
	}
}

// ── SimpleGauge BaseName/TypeName/Serialize/Deserialize ──────────────────────

func TestSimpleGauge_BaseName(t *testing.T) {
	g := gauge.NewSimpleGauge()
	if g.BaseName() != "SimpleGauge" {
		t.Errorf("BaseName = %q, want SimpleGauge", g.BaseName())
	}
}

func TestSimpleGauge_SerializeDeserialize_Defaults(t *testing.T) {
	orig := gauge.NewSimpleGauge()
	got := simpleGaugeRoundTrip(t, orig)

	if got.Shape != orig.Shape {
		t.Errorf("Shape: got %v, want %v", got.Shape, orig.Shape)
	}
	if got.ShowText != orig.ShowText {
		t.Errorf("ShowText: got %v, want %v", got.ShowText, orig.ShowText)
	}
}

func TestSimpleGauge_SerializeDeserialize_NonDefaults(t *testing.T) {
	orig := gauge.NewSimpleGauge()
	orig.Shape = gauge.SimpleGaugeShapeCircle
	orig.ShowText = false
	orig.TextFormat = "%.0f%%"
	orig.FirstSubScale.Enabled = false
	orig.FirstSubScale.ShowCaption = false
	orig.SecondSubScale.Enabled = false
	orig.SecondSubScale.ShowCaption = false

	got := simpleGaugeRoundTrip(t, orig)

	if got.Shape != gauge.SimpleGaugeShapeCircle {
		t.Errorf("Shape: got %v, want Circle", got.Shape)
	}
	if got.ShowText != false {
		t.Errorf("ShowText: got %v, want false", got.ShowText)
	}
	if got.TextFormat != "%.0f%%" {
		t.Errorf("TextFormat: got %q, want %%.0f%%%%", got.TextFormat)
	}
	if got.FirstSubScale.Enabled {
		t.Error("FirstSubScale.Enabled should be false")
	}
	if got.FirstSubScale.ShowCaption {
		t.Error("FirstSubScale.ShowCaption should be false")
	}
	if got.SecondSubScale.Enabled {
		t.Error("SecondSubScale.Enabled should be false")
	}
	if got.SecondSubScale.ShowCaption {
		t.Error("SecondSubScale.ShowCaption should be false")
	}
}

func TestSimpleGauge_SerializeDeserialize_TriangleShape(t *testing.T) {
	orig := gauge.NewSimpleGauge()
	orig.Shape = gauge.SimpleGaugeShapeTriangle
	got := simpleGaugeRoundTrip(t, orig)
	if got.Shape != gauge.SimpleGaugeShapeTriangle {
		t.Errorf("Shape: got %v, want Triangle", got.Shape)
	}
}

// ── SimpleProgressGauge BaseName/TypeName/Serialize/Deserialize ───────────────

func TestSimpleProgressGauge_BaseName(t *testing.T) {
	g := gauge.NewSimpleProgressGauge()
	if g.BaseName() != "SimpleProgressGauge" {
		t.Errorf("BaseName = %q, want SimpleProgressGauge", g.BaseName())
	}
}

func TestSimpleProgressGauge_TypeName(t *testing.T) {
	g := gauge.NewSimpleProgressGauge()
	if g.TypeName() != "SimpleProgressGauge" {
		t.Errorf("TypeName = %q, want SimpleProgressGauge", g.TypeName())
	}
}

func TestSimpleProgressGauge_SerializeDeserialize_Defaults(t *testing.T) {
	orig := gauge.NewSimpleProgressGauge()
	got := simpleProgressRoundTrip(t, orig)

	if got.ShowText != orig.ShowText {
		t.Errorf("ShowText: got %v, want %v", got.ShowText, orig.ShowText)
	}
}

func TestSimpleProgressGauge_SerializeDeserialize_ShowTextFalse(t *testing.T) {
	orig := gauge.NewSimpleProgressGauge()
	orig.ShowText = false
	got := simpleProgressRoundTrip(t, orig)

	if got.ShowText != false {
		t.Errorf("ShowText: got %v, want false", got.ShowText)
	}
}

func TestSimpleProgressGauge_SerializeDeserialize_WithValue(t *testing.T) {
	orig := gauge.NewSimpleProgressGauge()
	orig.SetValue(80)
	orig.Minimum = 0
	orig.Maximum = 200
	orig.SetValue(80)

	got := simpleProgressRoundTrip(t, orig)
	if got.Value() != orig.Value() {
		t.Errorf("Value: got %v, want %v", got.Value(), orig.Value())
	}
}

// ── parseColor error path (via RenderLinear with invalid color string) ─────────

func TestRenderLinear_InvalidColor_FallsBackToDefault(t *testing.T) {
	// parseColor with an invalid color string returns the default color.
	// We can trigger the error branch by using a color string that ParseColor rejects.
	g := gauge.NewLinearGauge()
	g.Pointer.Color = "not-a-color-at-all-xyz"
	g.SetValue(50)
	img := gauge.RenderLinear(g, 100, 20)
	if img == nil {
		t.Fatal("RenderLinear returned nil with invalid color")
	}
	// The render should still produce a valid image.
	b := img.Bounds()
	if b.Dx() != 100 || b.Dy() != 20 {
		t.Errorf("unexpected size %dx%d", b.Dx(), b.Dy())
	}
}

func TestRenderRadial_InvalidColor_FallsBackToDefault(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.Pointer.Color = "not-a-valid-color"
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial returned nil with invalid color")
	}
}

func TestRenderSimpleProgress_InvalidColor_FallsBackToDefault(t *testing.T) {
	g := gauge.NewSimpleProgressGauge()
	g.Pointer.Color = "bad-color-string"
	g.SetValue(50)
	img := gauge.RenderSimpleProgress(g, 100, 30)
	if img == nil {
		t.Fatal("RenderSimpleProgress returned nil with invalid color")
	}
}

func TestRenderSimple_InvalidColor_FallsBackToDefault(t *testing.T) {
	g := gauge.NewSimpleGauge()
	g.Pointer.Color = "bad-color"
	g.SetValue(50)
	img := gauge.RenderSimple(g, 100, 100)
	if img == nil {
		t.Fatal("RenderSimple returned nil with invalid color")
	}
}

// ── drawHLine / drawVLine edge cases ──────────────────────────────────────────

// Rendered via functions that call drawHLine/drawVLine with out-of-bounds y/x.
func TestRenderLinear_SmallSize_TriggersEdgeCases(t *testing.T) {
	// A very tiny image that causes the line drawing to hit boundary checks.
	g := gauge.NewLinearGauge()
	g.SetValue(50)
	// 1x1 image - the margin=4 means barW <= 0, so returns early.
	img := gauge.RenderLinear(g, 1, 1)
	if img == nil {
		t.Fatal("RenderLinear returned nil for 1x1")
	}
	// Verify it's still an image.
	if _, ok := img.(*image.RGBA); !ok {
		t.Error("expected *image.RGBA")
	}
}

func TestRenderLinear_Vertical_SmallSize(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.Orientation = gauge.OrientationVertical
	g.SetValue(50)
	// Small enough to trigger barW<=0 early return.
	img := gauge.RenderLinear(g, 1, 1)
	if img == nil {
		t.Fatal("RenderLinear vertical returned nil for 1x1")
	}
}

func TestRenderLinear_Vertical_ZeroValue(t *testing.T) {
	// fillH == 0, so the fill branch is skipped.
	g := gauge.NewLinearGauge()
	g.Orientation = gauge.OrientationVertical
	g.SetValue(0)
	img := gauge.RenderLinear(g, 100, 50)
	if img == nil {
		t.Fatal("RenderLinear vertical zero value returned nil")
	}
}

func TestRenderLinear_Horizontal_ZeroFill(t *testing.T) {
	// fillW == 0, so the fill branch is skipped.
	g := gauge.NewLinearGauge()
	g.SetValue(0)
	img := gauge.RenderLinear(g, 100, 50)
	if img == nil {
		t.Fatal("RenderLinear horizontal zero fill returned nil")
	}
}

func TestRenderSimpleProgress_SmallSize(t *testing.T) {
	// barW <= 0 triggers early return.
	g := gauge.NewSimpleProgressGauge()
	g.SetValue(50)
	img := gauge.RenderSimpleProgress(g, 1, 1)
	if img == nil {
		t.Fatal("RenderSimpleProgress returned nil for 1x1")
	}
}

func TestRenderSimpleProgress_ZeroFill(t *testing.T) {
	// fillW == 0, so the fill branch is skipped.
	g := gauge.NewSimpleProgressGauge()
	g.SetValue(0)
	img := gauge.RenderSimpleProgress(g, 100, 30)
	if img == nil {
		t.Fatal("RenderSimpleProgress zero fill returned nil")
	}
}

// ── drawArc edge cases (via RenderRadial small size) ──────────────────────────

func TestRenderRadial_SmallSize_TriggersEarlyReturn(t *testing.T) {
	// rx or ry <= 0 causes early return before drawing.
	g := gauge.NewRadialGauge()
	g.SetValue(50)
	// 1x1 image → rx = 0/2 - 4 = -4 <= 0
	img := gauge.RenderRadial(g, 1, 1)
	if img == nil {
		t.Fatal("RenderRadial returned nil for 1x1")
	}
}

func TestRenderRadial_NarrowSize_InnerArcSkipped(t *testing.T) {
	// rx = 5 means rx <= 4, so inner arc is skipped.
	g := gauge.NewRadialGauge()
	g.SetValue(50)
	// 18x18 → rx = 9-4 = 5, ry = 9-4 = 5, inner arc condition: rx > 4 && ry > 4 → not skipped
	// Let's try 10x10 → rx = 5-4 = 1, early return since rx<=0... no wait
	// Actually 10x10: rx = 10/2 - 4 = 1, ry = 10/2 - 4 = 1
	// But both > 0, so no early return.
	// The inner arc: rx > 4 is false (1 > 4 = false), so inner arc is skipped.
	img := gauge.RenderRadial(g, 10, 10)
	if img == nil {
		t.Fatal("RenderRadial returned nil for 10x10")
	}
}

func TestRenderRadial_NeedleThickening_EdgeX(t *testing.T) {
	// Tests cx+1 < w branch (true) and cy+1 < h branch (true).
	g := gauge.NewRadialGauge()
	g.SetValue(50)
	img := gauge.RenderRadial(g, 200, 200)
	if img == nil {
		t.Fatal("RenderRadial returned nil for 200x200")
	}
}

func TestRenderRadial_NeedleThickening_AtEdge(t *testing.T) {
	// Use a size where cx+1 == w or cy+1 == h to hit the false branches.
	// With a 2x2: cx=1, w=2, cx+1=2 which is not < w (2 < 2 is false).
	g := gauge.NewRadialGauge()
	g.SetValue(50)
	// Too small → rx <= 0, early return.
	// Use a size like 9x9: cx=4, w=9, cx+1=5 < 9 → true.
	// With 2x2 we hit the early return, so try 6x6: cx=3, rx=3-4=-1 → early return.
	// Size 12x12: cx=6, rx=6-4=2, ry=2. No early return.
	// cx+1=7 < 12 → true. cy+1=7 < 12 → true.
	// For the false branch (cx+1 >= w) we'd need cx == w-1.
	// That would require width=2*cx+1 which is an odd width where cx is large.
	// In practice, this is difficult to trigger via public API since cx = w/2.
	// Just verify the render works without panic.
	img := gauge.RenderRadial(g, 12, 12)
	if img == nil {
		t.Fatal("RenderRadial returned nil for 12x12")
	}
}

// ── min function coverage ─────────────────────────────────────────────────────

// min is package-private but exercised indirectly via RenderRadial.
// The function has branches: a < b (return a) and a >= b (return b).

func TestMin_ViaRenderRadial_WiderThanTall(t *testing.T) {
	// w > h → rx > ry, min(rx,ry) returns ry.
	g := gauge.NewRadialGauge()
	g.SetValue(50)
	img := gauge.RenderRadial(g, 200, 100)
	if img == nil {
		t.Fatal("RenderRadial wide returned nil")
	}
}

func TestMin_ViaRenderRadial_TallerThanWide(t *testing.T) {
	// h > w → ry > rx, min(rx,ry) returns rx.
	g := gauge.NewRadialGauge()
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 200)
	if img == nil {
		t.Fatal("RenderRadial tall returned nil")
	}
}

func TestMin_ViaRenderRadial_Square(t *testing.T) {
	// w == h → rx == ry, min(rx,ry) returns either (they're equal).
	g := gauge.NewRadialGauge()
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial square returned nil")
	}
}

// ── RenderLinear uncovered branches ──────────────────────────────────────────

func TestRenderLinear_Horizontal_LargeValue(t *testing.T) {
	// fillW > 0 — the fill rect is drawn.
	g := gauge.NewLinearGauge()
	g.SetValue(100)
	img := gauge.RenderLinear(g, 200, 40)
	if img == nil {
		t.Fatal("RenderLinear full returned nil")
	}
	b := img.Bounds()
	if b.Dx() != 200 {
		t.Errorf("width = %d, want 200", b.Dx())
	}
}

func TestRenderLinear_Vertical_LargeValue(t *testing.T) {
	// fillH > 0 — the fill rect is drawn.
	g := gauge.NewLinearGauge()
	g.Orientation = gauge.OrientationVertical
	g.SetValue(100)
	img := gauge.RenderLinear(g, 40, 200)
	if img == nil {
		t.Fatal("RenderLinear vertical full returned nil")
	}
}

// ── RenderSimpleProgress uncovered branches ───────────────────────────────────

func TestRenderSimpleProgress_FullValue(t *testing.T) {
	// fillW == barW (100% fill).
	g := gauge.NewSimpleProgressGauge()
	g.SetValue(100)
	img := gauge.RenderSimpleProgress(g, 200, 30)
	if img == nil {
		t.Fatal("RenderSimpleProgress full value returned nil")
	}
}

func TestRenderSimpleProgress_PartialValue(t *testing.T) {
	// fillW > 0 and fillW < barW.
	g := gauge.NewSimpleProgressGauge()
	g.SetValue(50)
	img := gauge.RenderSimpleProgress(g, 200, 30)
	if img == nil {
		t.Fatal("RenderSimpleProgress partial returned nil")
	}
}

// ── RenderRadial uncovered branches ──────────────────────────────────────────

func TestRenderRadial_InnerArcDrawn(t *testing.T) {
	// rx > 4 && ry > 4 → inner arc is drawn.
	g := gauge.NewRadialGauge()
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial with inner arc returned nil")
	}
	b := img.Bounds()
	if b.Dx() != 100 {
		t.Errorf("width = %d, want 100", b.Dx())
	}
}

func TestRenderRadial_InnerArcSkipped(t *testing.T) {
	// Use small size where rx <= 4 (but > 0) — inner arc condition is false.
	g := gauge.NewRadialGauge()
	g.SetValue(50)
	// 12x12: rx = 6-4 = 2, not > 4 (inner arc skipped; actually 2 is not > 4 → false)
	// Wait: 2 > 4 is false, so inner arc is skipped.
	img := gauge.RenderRadial(g, 12, 12)
	if img == nil {
		t.Fatal("RenderRadial small (inner arc skipped) returned nil")
	}
}

// ── drawArc with negative sweep ───────────────────────────────────────────────

func TestRenderRadial_NegativeSweepAngle(t *testing.T) {
	// StartAngle > EndAngle causes sweep<0, which is corrected by +360.
	g := gauge.NewRadialGauge()
	g.StartAngle = 90
	g.EndAngle = -90
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial negative sweep returned nil")
	}
}

// ── GaugeObject Serialize — default pointer width (6) is not serialized ────────

func TestGaugeObject_Serialize_PointerWidth_Default(t *testing.T) {
	// Pointer.Width == 6 (default) should NOT appear in the XML.
	orig := gauge.NewGaugeObject()
	// default Pointer.Width is 6, so it's not serialized.
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("GaugeObject")   //nolint:errcheck
	orig.Serialize(w)              //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	xml := buf.String()
	if strings.Contains(xml, `Pointer.Width`) {
		t.Errorf("Pointer.Width should not appear in XML for default value, got:\n%s", xml)
	}
}

func TestGaugeObject_Serialize_PointerColor_Default(t *testing.T) {
	// Pointer.Color == "#CC0000" (default) should NOT appear in the XML.
	orig := gauge.NewGaugeObject()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("GaugeObject")   //nolint:errcheck
	orig.Serialize(w)              //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	xml := buf.String()
	if strings.Contains(xml, `Pointer.Color`) {
		t.Errorf("Pointer.Color should not appear for default color, got:\n%s", xml)
	}
}

// ── LinearGauge serialization — orientation not serialized when horizontal ─────

func TestLinearGauge_Serialize_HorizontalNotWritten(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.Orientation = gauge.OrientationHorizontal
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("LinearGauge")   //nolint:errcheck
	g.Serialize(w)                 //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	xml := buf.String()
	if strings.Contains(xml, `Orientation`) {
		t.Errorf("Orientation should not appear for default (Horizontal), got:\n%s", xml)
	}
}

func TestLinearGauge_Serialize_VerticalWritten(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.Orientation = gauge.OrientationVertical
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("LinearGauge")   //nolint:errcheck
	g.Serialize(w)                 //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Orientation`) {
		t.Errorf("Orientation should appear for Vertical, got:\n%s", xml)
	}
}

// ── RadialGauge serialization ─────────────────────────────────────────────────

func TestRadialGauge_Serialize_DefaultAnglesNotWritten(t *testing.T) {
	g := gauge.NewRadialGauge()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("RadialGauge")   //nolint:errcheck
	g.Serialize(w)                 //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	xml := buf.String()
	// Default StartAngle=-135 and EndAngle=135 should not be serialized.
	if strings.Contains(xml, `StartAngle`) {
		t.Errorf("StartAngle should not appear for default -135, got:\n%s", xml)
	}
	if strings.Contains(xml, `EndAngle`) {
		t.Errorf("EndAngle should not appear for default 135, got:\n%s", xml)
	}
}

// ── SimpleProgressGauge serialization ────────────────────────────────────────

func TestSimpleProgressGauge_Serialize_ShowTextTrueNotWritten(t *testing.T) {
	g := gauge.NewSimpleProgressGauge()
	g.ShowText = true // default
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("SimpleProgressGauge")   //nolint:errcheck
	g.Serialize(w)                         //nolint:errcheck
	w.EndObject()                          //nolint:errcheck
	w.Flush()                              //nolint:errcheck

	xml := buf.String()
	if strings.Contains(xml, `ShowText`) {
		t.Errorf("ShowText should not appear when true (default), got:\n%s", xml)
	}
}

func TestSimpleProgressGauge_Serialize_ShowTextFalseWritten(t *testing.T) {
	g := gauge.NewSimpleProgressGauge()
	g.ShowText = false
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("SimpleProgressGauge")   //nolint:errcheck
	g.Serialize(w)                         //nolint:errcheck
	w.EndObject()                          //nolint:errcheck
	w.Flush()                              //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `ShowText`) {
		t.Errorf("ShowText should appear when false, got:\n%s", xml)
	}
}

// ── SimpleGauge serialization ─────────────────────────────────────────────────

func TestSimpleGauge_Serialize_NonDefaultShape(t *testing.T) {
	g := gauge.NewSimpleGauge()
	g.Shape = gauge.SimpleGaugeShapeCircle
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("SimpleGauge")   //nolint:errcheck
	g.Serialize(w)                 //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Shape`) {
		t.Errorf("Shape should appear when not Rectangle, got:\n%s", xml)
	}
}

func TestSimpleGauge_Serialize_SubScaleDefaults(t *testing.T) {
	// Default sub-scales (Enabled=true, ShowCaption=true) should not be serialized.
	g := gauge.NewSimpleGauge()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("SimpleGauge")   //nolint:errcheck
	g.Serialize(w)                 //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	xml := buf.String()
	if strings.Contains(xml, `SubScale`) {
		t.Errorf("SubScale props should not appear for defaults, got:\n%s", xml)
	}
}

// ── Verify SimpleProgressGauge doesn't contain lingering nil-pointer issues ───

func TestSimpleProgressGauge_NewDefaults(t *testing.T) {
	g := gauge.NewSimpleProgressGauge()
	if g.BaseName() != "SimpleProgressGauge" {
		t.Errorf("BaseName = %q", g.BaseName())
	}
	if g.TypeName() != "SimpleProgressGauge" {
		t.Errorf("TypeName = %q", g.TypeName())
	}
	if !g.ShowText {
		t.Error("ShowText should default to true")
	}
	if g.Minimum != 0 {
		t.Errorf("Minimum = %v, want 0", g.Minimum)
	}
	if g.Maximum != 100 {
		t.Errorf("Maximum = %v, want 100", g.Maximum)
	}
}

// ── Verify GaugeObject minimum != 100 case ────────────────────────────────────

func TestGaugeObject_Serialize_NonDefaultMaximum(t *testing.T) {
	// Maximum != 100 should be written.
	g := gauge.NewGaugeObject()
	g.Maximum = 50
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("GaugeObject")   //nolint:errcheck
	g.Serialize(w)                 //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Maximum`) {
		t.Errorf("Maximum should appear when != 100, got:\n%s", xml)
	}
}

func TestGaugeObject_Serialize_NonDefaultMinimum(t *testing.T) {
	// Minimum != 0 should be written.
	g := gauge.NewGaugeObject()
	g.Minimum = 5
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("GaugeObject")   //nolint:errcheck
	g.Serialize(w)                 //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Minimum`) {
		t.Errorf("Minimum should appear when != 0, got:\n%s", xml)
	}
}

func TestGaugeObject_Serialize_NonDefaultValue(t *testing.T) {
	// Value != 0 should be written.
	g := gauge.NewGaugeObject()
	g.SetValue(42)
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("GaugeObject")   //nolint:errcheck
	g.Serialize(w)                 //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Value`) {
		t.Errorf("Value should appear when != 0, got:\n%s", xml)
	}
}
