package gauge

// gauge_internal_test.go — white-box tests for gauge.go Serialize/Deserialize.
// Uses package gauge (internal) to access unexported fields and verify attribute
// branch coverage.
//
// NOTE: The `return err` branches inside Serialize/Deserialize that propagate
// errors from the embedded ReportComponentBase are dead code: the entire
// base Serialize/Deserialize chain (BaseObject → ComponentBase →
// ReportComponentBase) always returns nil because none of the report.Writer
// write methods (WriteStr, WriteInt, WriteBool, WriteFloat) can fail.
// Those branches cannot be covered without modifying source code.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// serializeToXML serializes obj under the given tag and returns the XML string.
func serializeToXML(t *testing.T, tag string, obj interface {
	Serialize(w report.Writer) error
}) string {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject(tag); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := obj.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.String()
}

// deserializeFromXML deserializes the given XML into obj.
func deserializeFromXML(t *testing.T, xmlStr string, obj interface {
	Deserialize(r report.Reader) error
}) {
	t.Helper()
	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed; XML:\n%s", xmlStr)
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
}

// ── GaugeObject attribute branches ───────────────────────────────────────────

// TestGaugeObject_Serialize_PointerHeight_NonDefault verifies that a non-zero
// Pointer.Height is serialized.
func TestGaugeObject_Serialize_PointerHeight_NonDefault(t *testing.T) {
	g := NewGaugeObject()
	g.Pointer.Height = 12.5
	xml := serializeToXML(t, "GaugeObject", g)
	if !strings.Contains(xml, "Pointer.Height") {
		t.Errorf("Pointer.Height should appear when non-zero, got:\n%s", xml)
	}
}

// TestGaugeObject_Deserialize_PointerHeight verifies that Pointer.Height is
// read back from XML.
func TestGaugeObject_Deserialize_PointerHeight_RoundTrip(t *testing.T) {
	orig := NewGaugeObject()
	orig.Pointer.Height = 12.5
	xml := serializeToXML(t, "GaugeObject", orig)

	got := NewGaugeObject()
	deserializeFromXML(t, xml, got)

	if got.Pointer.Height != orig.Pointer.Height {
		t.Errorf("Pointer.Height: got %v, want %v", got.Pointer.Height, orig.Pointer.Height)
	}
}

// TestGaugeObject_Serialize_ScaleMinorTicksLength verifies serialization of
// Scale.MinorTicks.Length when non-zero.
func TestGaugeObject_Serialize_ScaleMinorTicksLength(t *testing.T) {
	g := NewGaugeObject()
	g.Scale.MinorTicks.Length = 3.5
	xml := serializeToXML(t, "GaugeObject", g)
	if !strings.Contains(xml, "Scale.MinorTicks.Length") {
		t.Errorf("Scale.MinorTicks.Length should appear, got:\n%s", xml)
	}
}

// TestGaugeObject_Serialize_NilScale verifies that nil Scale doesn't panic.
func TestGaugeObject_Serialize_NilScale(t *testing.T) {
	g := NewGaugeObject()
	g.Scale = nil
	// Should not panic.
	xml := serializeToXML(t, "GaugeObject", g)
	_ = xml
}

// TestGaugeObject_Serialize_NilPointer verifies that nil Pointer doesn't panic.
func TestGaugeObject_Serialize_NilPointer(t *testing.T) {
	g := NewGaugeObject()
	g.Pointer = nil
	// Should not panic.
	xml := serializeToXML(t, "GaugeObject", g)
	_ = xml
}

// TestGaugeObject_Serialize_LabelFont verifies Label.Font is serialized.
func TestGaugeObject_Serialize_LabelFont(t *testing.T) {
	g := NewGaugeObject()
	g.Label.Font = "Arial, 12pt"
	xml := serializeToXML(t, "GaugeObject", g)
	if !strings.Contains(xml, "Label.Font") {
		t.Errorf("Label.Font should appear, got:\n%s", xml)
	}
}

// TestGaugeObject_Serialize_LabelText verifies Label.Text is serialized.
func TestGaugeObject_Serialize_LabelText(t *testing.T) {
	g := NewGaugeObject()
	g.Label.Text = "Speed"
	xml := serializeToXML(t, "GaugeObject", g)
	if !strings.Contains(xml, "Label.Text") {
		t.Errorf("Label.Text should appear, got:\n%s", xml)
	}
}

// TestGaugeObject_Deserialize_LabelFont_RoundTrip verifies Label.Font round-trip.
func TestGaugeObject_Deserialize_LabelFont_RoundTrip(t *testing.T) {
	orig := NewGaugeObject()
	orig.Label.Font = "Times New Roman, 10pt"
	xml := serializeToXML(t, "GaugeObject", orig)

	got := NewGaugeObject()
	deserializeFromXML(t, xml, got)

	if got.Label.Font != orig.Label.Font {
		t.Errorf("Label.Font: got %q, want %q", got.Label.Font, orig.Label.Font)
	}
}

// TestGaugeObject_Deserialize_AllScaleFields verifies all Scale tick fields round-trip.
func TestGaugeObject_Deserialize_AllScaleFields(t *testing.T) {
	orig := NewGaugeObject()
	orig.Scale.Font = "Courier, 8pt"
	orig.Scale.MajorTicks.Width = 3
	orig.Scale.MajorTicks.Color = "#FF0000"
	orig.Scale.MajorTicks.Length = 10
	orig.Scale.MinorTicks.Width = 1
	orig.Scale.MinorTicks.Color = "#0000FF"
	orig.Scale.MinorTicks.Length = 5
	xml := serializeToXML(t, "GaugeObject", orig)

	got := NewGaugeObject()
	deserializeFromXML(t, xml, got)

	checks := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Scale.Font", got.Scale.Font, orig.Scale.Font},
		{"Scale.MajorTicks.Width", got.Scale.MajorTicks.Width, orig.Scale.MajorTicks.Width},
		{"Scale.MajorTicks.Color", got.Scale.MajorTicks.Color, orig.Scale.MajorTicks.Color},
		{"Scale.MajorTicks.Length", got.Scale.MajorTicks.Length, orig.Scale.MajorTicks.Length},
		{"Scale.MinorTicks.Width", got.Scale.MinorTicks.Width, orig.Scale.MinorTicks.Width},
		{"Scale.MinorTicks.Color", got.Scale.MinorTicks.Color, orig.Scale.MinorTicks.Color},
		{"Scale.MinorTicks.Length", got.Scale.MinorTicks.Length, orig.Scale.MinorTicks.Length},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
}

// ── LinearGauge attribute branches ───────────────────────────────────────────

// TestLinearGauge_Serialize_AllNonDefaults exercises all non-default attribute branches.
func TestLinearGauge_Serialize_AllNonDefaults(t *testing.T) {
	g := NewLinearGauge()
	g.Orientation = OrientationVertical
	g.Inverted = true
	g.Minimum = 5
	g.Maximum = 200
	g.SetValue(100)
	g.Expression = "[Field.Value]"

	xml := serializeToXML(t, "LinearGauge", g)

	for _, want := range []string{"Orientation", "Inverted", "Minimum", "Maximum", "Value", "Expression"} {
		if !strings.Contains(xml, want) {
			t.Errorf("expected %q in XML, got:\n%s", want, xml)
		}
	}
}

// TestLinearGauge_Deserialize_AllNonDefaults round-trips all non-default fields.
func TestLinearGauge_Deserialize_AllNonDefaults(t *testing.T) {
	orig := NewLinearGauge()
	orig.Orientation = OrientationVertical
	orig.Inverted = true
	orig.Minimum = 5
	orig.Maximum = 200
	orig.SetValue(100)
	orig.Expression = "[Field.Value]"
	xml := serializeToXML(t, "LinearGauge", orig)

	got := NewLinearGauge()
	deserializeFromXML(t, xml, got)

	if got.Orientation != OrientationVertical {
		t.Errorf("Orientation: got %v, want Vertical", got.Orientation)
	}
	if !got.Inverted {
		t.Error("Inverted: got false, want true")
	}
	if got.Minimum != orig.Minimum {
		t.Errorf("Minimum: got %v, want %v", got.Minimum, orig.Minimum)
	}
	if got.Maximum != orig.Maximum {
		t.Errorf("Maximum: got %v, want %v", got.Maximum, orig.Maximum)
	}
	if got.Expression != orig.Expression {
		t.Errorf("Expression: got %q, want %q", got.Expression, orig.Expression)
	}
}

// ── RadialGauge attribute branches ───────────────────────────────────────────

// TestRadialGauge_Serialize_NonDefaultAngles exercises StartAngle/EndAngle branches.
func TestRadialGauge_Serialize_NonDefaultAngles(t *testing.T) {
	g := NewRadialGauge()
	g.StartAngle = -45 // != default 135
	g.EndAngle = 90    // != default 45
	xml := serializeToXML(t, "RadialGauge", g)

	if !strings.Contains(xml, "StartAngle") {
		t.Errorf("StartAngle should appear for non-default, got:\n%s", xml)
	}
	if !strings.Contains(xml, "EndAngle") {
		t.Errorf("EndAngle should appear for non-default, got:\n%s", xml)
	}
}

// TestRadialGauge_Deserialize_DefaultAngles verifies defaults are preserved when
// no angles are serialized.
func TestRadialGauge_Deserialize_DefaultAngles(t *testing.T) {
	orig := NewRadialGauge()
	// Default angles (135, 45) should NOT be serialized.
	xml := serializeToXML(t, "RadialGauge", orig)

	got := NewRadialGauge()
	deserializeFromXML(t, xml, got)

	if got.StartAngle != 135 {
		t.Errorf("StartAngle: got %v, want 135", got.StartAngle)
	}
	if got.EndAngle != 45 {
		t.Errorf("EndAngle: got %v, want 45", got.EndAngle)
	}
}

// ── SimpleGauge attribute branches ───────────────────────────────────────────

// TestSimpleGauge_Serialize_AllNonDefaults exercises every non-default serialization
// branch in SimpleGauge.Serialize.
func TestSimpleGauge_Serialize_AllNonDefaults(t *testing.T) {
	g := NewSimpleGauge()
	g.Shape = SimpleGaugeShapeTriangle
	g.ShowText = false
	g.TextFormat = "%.1f%%"
	g.FirstSubScale.Enabled = false
	g.FirstSubScale.ShowCaption = false
	g.SecondSubScale.Enabled = false
	g.SecondSubScale.ShowCaption = false

	xml := serializeToXML(t, "SimpleGauge", g)

	for _, want := range []string{
		"Shape", "ShowText", "TextFormat",
		"Scale.FirstSubScale.Enabled", "Scale.FirstSubScale.ShowCaption",
		"Scale.SecondSubScale.Enabled", "Scale.SecondSubScale.ShowCaption",
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("expected %q in XML, got:\n%s", want, xml)
		}
	}
}

// TestSimpleGauge_Deserialize_AllNonDefaults round-trips all non-default SimpleGauge fields.
func TestSimpleGauge_Deserialize_AllNonDefaults(t *testing.T) {
	orig := NewSimpleGauge()
	orig.Shape = SimpleGaugeShapeCircle
	orig.ShowText = false
	orig.TextFormat = "%.0f%%"
	orig.FirstSubScale.Enabled = false
	orig.FirstSubScale.ShowCaption = false
	orig.SecondSubScale.Enabled = false
	orig.SecondSubScale.ShowCaption = false
	xml := serializeToXML(t, "SimpleGauge", orig)

	got := NewSimpleGauge()
	deserializeFromXML(t, xml, got)

	if got.Shape != SimpleGaugeShapeCircle {
		t.Errorf("Shape: got %v, want Circle", got.Shape)
	}
	if got.ShowText != false {
		t.Errorf("ShowText: got %v, want false", got.ShowText)
	}
	if got.TextFormat != orig.TextFormat {
		t.Errorf("TextFormat: got %q, want %q", got.TextFormat, orig.TextFormat)
	}
	if got.FirstSubScale.Enabled {
		t.Error("FirstSubScale.Enabled: got true, want false")
	}
	if got.FirstSubScale.ShowCaption {
		t.Error("FirstSubScale.ShowCaption: got true, want false")
	}
	if got.SecondSubScale.Enabled {
		t.Error("SecondSubScale.Enabled: got true, want false")
	}
	if got.SecondSubScale.ShowCaption {
		t.Error("SecondSubScale.ShowCaption: got true, want false")
	}
}

// TestSimpleGauge_Serialize_DefaultTextFormat verifies that default TextFormat (%g%%)
// is NOT serialized.
func TestSimpleGauge_Serialize_DefaultTextFormat(t *testing.T) {
	g := NewSimpleGauge()
	// TextFormat defaults to "%g%%".
	xml := serializeToXML(t, "SimpleGauge", g)
	if strings.Contains(xml, "TextFormat") {
		t.Errorf("TextFormat should not appear for default value, got:\n%s", xml)
	}
}

// ── SimpleProgressGauge attribute branches ────────────────────────────────────

// TestSimpleProgressGauge_Serialize_ShowTextFalse exercises the non-default ShowText branch.
func TestSimpleProgressGauge_Serialize_ShowTextFalse(t *testing.T) {
	g := NewSimpleProgressGauge()
	g.ShowText = false
	xml := serializeToXML(t, "SimpleProgressGauge", g)
	if !strings.Contains(xml, "ShowText") {
		t.Errorf("ShowText should appear when false, got:\n%s", xml)
	}
}

// TestSimpleProgressGauge_Deserialize_ShowTextFalse_Internal round-trips ShowText=false.
func TestSimpleProgressGauge_Deserialize_ShowTextFalse_Internal(t *testing.T) {
	orig := NewSimpleProgressGauge()
	orig.ShowText = false
	xml := serializeToXML(t, "SimpleProgressGauge", orig)

	got := NewSimpleProgressGauge()
	deserializeFromXML(t, xml, got)

	if got.ShowText {
		t.Error("ShowText: got true, want false")
	}
}

// TestSimpleProgressGauge_Serialize_DefaultShowText verifies default (true) is NOT serialized.
func TestSimpleProgressGauge_Serialize_DefaultShowText(t *testing.T) {
	g := NewSimpleProgressGauge()
	g.ShowText = true
	xml := serializeToXML(t, "SimpleProgressGauge", g)
	if strings.Contains(xml, "ShowText") {
		t.Errorf("ShowText should not appear for default (true), got:\n%s", xml)
	}
}

// TestSimpleProgressGauge_Deserialize_DefaultShowText verifies default ShowText=true
// when not in XML.
func TestSimpleProgressGauge_Deserialize_DefaultShowText(t *testing.T) {
	orig := NewSimpleProgressGauge()
	// Default ShowText=true → not serialized.
	xml := serializeToXML(t, "SimpleProgressGauge", orig)

	got := NewSimpleProgressGauge()
	deserializeFromXML(t, xml, got)

	if !got.ShowText {
		t.Error("ShowText: got false, want true (default)")
	}
}

// ── SimpleSubScale exhaustive attribute tests ────────────────────────────────

// TestSimpleSubScale_AllCombinations exercises all combinations of Enabled/ShowCaption
// for both FirstSubScale and SecondSubScale.
func TestSimpleSubScale_AllCombinations(t *testing.T) {
	cases := []struct {
		name           string
		firstEnabled   bool
		firstCaption   bool
		secondEnabled  bool
		secondCaption  bool
	}{
		{"all_true", true, true, true, true},
		{"first_disabled", false, true, true, true},
		{"first_no_caption", true, false, true, true},
		{"second_disabled", true, true, false, true},
		{"second_no_caption", true, true, true, false},
		{"all_false", false, false, false, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			orig := NewSimpleGauge()
			orig.FirstSubScale.Enabled = tc.firstEnabled
			orig.FirstSubScale.ShowCaption = tc.firstCaption
			orig.SecondSubScale.Enabled = tc.secondEnabled
			orig.SecondSubScale.ShowCaption = tc.secondCaption

			xml := serializeToXML(t, "SimpleGauge", orig)

			got := NewSimpleGauge()
			deserializeFromXML(t, xml, got)

			if got.FirstSubScale.Enabled != tc.firstEnabled {
				t.Errorf("FirstSubScale.Enabled: got %v, want %v", got.FirstSubScale.Enabled, tc.firstEnabled)
			}
			if got.FirstSubScale.ShowCaption != tc.firstCaption {
				t.Errorf("FirstSubScale.ShowCaption: got %v, want %v", got.FirstSubScale.ShowCaption, tc.firstCaption)
			}
			if got.SecondSubScale.Enabled != tc.secondEnabled {
				t.Errorf("SecondSubScale.Enabled: got %v, want %v", got.SecondSubScale.Enabled, tc.secondEnabled)
			}
			if got.SecondSubScale.ShowCaption != tc.secondCaption {
				t.Errorf("SecondSubScale.ShowCaption: got %v, want %v", got.SecondSubScale.ShowCaption, tc.secondCaption)
			}
		})
	}
}

// ── GaugeObject Pointer.Width branch at exactly 6 (default, not serialized) ──

// TestGaugeObject_Pointer_Width_ExactlyDefault verifies Pointer.Width=6 is not
// serialized (it's the default).
func TestGaugeObject_Pointer_Width_ExactlyDefault(t *testing.T) {
	g := NewGaugeObject()
	g.Pointer.Width = 6 // exactly the default
	xml := serializeToXML(t, "GaugeObject", g)
	if strings.Contains(xml, "Pointer.Width") {
		t.Errorf("Pointer.Width should not appear for default (6), got:\n%s", xml)
	}
}

// TestGaugeObject_Pointer_Width_JustBelowDefault verifies Pointer.Width=5 IS serialized.
func TestGaugeObject_Pointer_Width_JustBelowDefault(t *testing.T) {
	g := NewGaugeObject()
	g.Pointer.Width = 5 // not 6
	xml := serializeToXML(t, "GaugeObject", g)
	if !strings.Contains(xml, "Pointer.Width") {
		t.Errorf("Pointer.Width should appear for non-default (5), got:\n%s", xml)
	}
}

// TestGaugeObject_Pointer_Width_Zero verifies Pointer.Width=0 IS serialized.
func TestGaugeObject_Pointer_Width_Zero(t *testing.T) {
	g := NewGaugeObject()
	g.Pointer.Width = 0
	xml := serializeToXML(t, "GaugeObject", g)
	if !strings.Contains(xml, "Pointer.Width") {
		t.Errorf("Pointer.Width should appear for 0, got:\n%s", xml)
	}
}

// ── GaugeObject Pointer.Color branch for default and empty string ─────────────

// TestGaugeObject_Pointer_Color_EmptyString verifies empty Pointer.Color is
// not serialized (the condition is Color != "" && Color != "Orange").
func TestGaugeObject_Pointer_Color_EmptyString(t *testing.T) {
	g := NewGaugeObject()
	g.Pointer.Color = ""
	xml := serializeToXML(t, "GaugeObject", g)
	if strings.Contains(xml, "Pointer.Color") {
		t.Errorf("Pointer.Color should not appear for empty string, got:\n%s", xml)
	}
}
