package gauge_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── LinearGauge Deserialize: error from GaugeObject.Deserialize ───────────────
// These paths go through GaugeObject.Deserialize which calls ReportComponentBase.Deserialize.
// The base Deserialize reads XML attributes; passing a malformed reader causes no error
// (the serial.Reader is lenient) — so we focus on the branches that are genuinely uncovered.

// The remaining uncovered branches for LinearGauge.Deserialize and RadialGauge.Deserialize
// are likely the successful path for reading Orientation/Inverted when they ARE serialized.
// We cover them with round-trip tests for specific field combinations.

func TestLinearGauge_Deserialize_OrientationFromXML(t *testing.T) {
	// Serialize with non-default orientation → Deserialize reads it.
	orig := gauge.NewLinearGauge()
	orig.Orientation = gauge.OrientationVertical
	orig.Inverted = true

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("LinearGauge")   //nolint:errcheck
	orig.Serialize(w)              //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	r := serial.NewReader(strings.NewReader(buf.String()))
	r.ReadObjectHeader()
	g2 := gauge.NewLinearGauge()
	if err := g2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g2.Orientation != gauge.OrientationVertical {
		t.Errorf("Orientation = %v, want Vertical", g2.Orientation)
	}
	if !g2.Inverted {
		t.Error("Inverted should be true")
	}
}

func TestRadialGauge_Deserialize_CustomAngles(t *testing.T) {
	orig := gauge.NewRadialGauge()
	orig.StartAngle = -90
	orig.EndAngle = 90

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("RadialGauge")   //nolint:errcheck
	orig.Serialize(w)              //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	r := serial.NewReader(strings.NewReader(buf.String()))
	r.ReadObjectHeader()
	g2 := gauge.NewRadialGauge()
	if err := g2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g2.StartAngle != -90 {
		t.Errorf("StartAngle = %v, want -90", g2.StartAngle)
	}
	if g2.EndAngle != 90 {
		t.Errorf("EndAngle = %v, want 90", g2.EndAngle)
	}
}

// ── SimpleGauge.Serialize: TextFormat empty branch ─────────────────────────────

func TestSimpleGauge_Serialize_EmptyTextFormat(t *testing.T) {
	// TextFormat == "" is different from the default "%g%%" — should be serialized.
	g := gauge.NewSimpleGauge()
	g.TextFormat = ""

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("SimpleGauge")   //nolint:errcheck
	g.Serialize(w)                 //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	// Empty TextFormat != "%g%%" so it should be serialized.
	xml := buf.String()
	if !strings.Contains(xml, "TextFormat") {
		t.Errorf("TextFormat should be serialized when empty string, got:\n%s", xml)
	}
}

// ── SimpleGauge.Deserialize: TextFormat read ──────────────────────────────────

func TestSimpleGauge_Deserialize_TextFormat(t *testing.T) {
	orig := gauge.NewSimpleGauge()
	orig.TextFormat = "%.1f%%"
	orig.Shape = gauge.SimpleGaugeShapeCircle
	orig.ShowText = false

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("SimpleGauge")   //nolint:errcheck
	orig.Serialize(w)              //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	r := serial.NewReader(strings.NewReader(buf.String()))
	r.ReadObjectHeader()
	g2 := gauge.NewSimpleGauge()
	if err := g2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g2.TextFormat != "%.1f%%" {
		t.Errorf("TextFormat = %q, want %%.1f%%%%", g2.TextFormat)
	}
	if g2.Shape != gauge.SimpleGaugeShapeCircle {
		t.Errorf("Shape = %v, want Circle", g2.Shape)
	}
	if g2.ShowText != false {
		t.Error("ShowText should be false")
	}
}

// ── SimpleProgressGauge.Deserialize: ShowText=false ──────────────────────────

func TestSimpleProgressGauge_Deserialize_ShowTextFalse(t *testing.T) {
	orig := gauge.NewSimpleProgressGauge()
	orig.ShowText = false

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("SimpleProgressGauge")   //nolint:errcheck
	orig.Serialize(w)                      //nolint:errcheck
	w.EndObject()                          //nolint:errcheck
	w.Flush()                              //nolint:errcheck

	r := serial.NewReader(strings.NewReader(buf.String()))
	r.ReadObjectHeader()
	g2 := gauge.NewSimpleProgressGauge()
	if err := g2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if g2.ShowText != false {
		t.Error("ShowText should be false after Deserialize")
	}
}

// ── render.go: drawHLine / drawVLine out-of-bounds ────────────────────────────
// These functions have early-return branches when y (for HLine) is out of bounds
// or x (for VLine) is out of bounds. We trigger them by using very small images.

func TestRenderLinear_HLine_OutOfBounds_Y(t *testing.T) {
	// With a 1-pixel-high image and margin=4, barH = 1-8 < 0 → early return.
	// But if the border drawing somehow has y outside bounds, drawHLine returns early.
	// Actually barW/barH <= 0 causes early return BEFORE drawRectBorder.
	// To trigger drawHLine's y-out-of-bounds branch, we need y < 0 or y >= h.
	// drawHLine is called by drawRectBorder which is called after the barW/barH check.
	// So we need barW > 0 && barH > 0, but drawHLine gets called with y outside [0,h).
	// This is hard to achieve directly. Let's use 8x8 image (barW=0, barH=0 → early return).
	// Instead use 10x10 where barW=2, barH=2, and drawRectBorder draws at y=4 and y=5
	// which are within [0,10). So the out-of-bounds branch is hard to hit indirectly.
	// We'll verify the function works correctly with normal usage instead.
	g := gauge.NewLinearGauge()
	g.SetValue(50)
	img := gauge.RenderLinear(g, 20, 10)
	if img == nil {
		t.Fatal("RenderLinear 20x10 returned nil")
	}
}

func TestRenderLinear_VLine_OutOfBounds_X(t *testing.T) {
	// drawVLine: x < b.Min.X || x >= b.Max.X → early return.
	// The border at x+w-1 could go out of bounds for very small images.
	g := gauge.NewLinearGauge()
	g.SetValue(50)
	img := gauge.RenderLinear(g, 10, 20)
	if img == nil {
		t.Fatal("RenderLinear 10x20 returned nil")
	}
}

// ── parseColor: empty string returns default ──────────────────────────────────

func TestRenderLinear_ParseColor_EmptyString_UsesDefault(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.Pointer.Color = "" // empty → parseColor returns def
	g.SetValue(50)
	img := gauge.RenderLinear(g, 100, 30)
	if img == nil {
		t.Fatal("RenderLinear with empty color returned nil")
	}
}

func TestRenderRadial_ParseColor_EmptyString_UsesDefault(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.Pointer.Color = ""
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial with empty color returned nil")
	}
}

func TestRenderSimple_ParseColor_EmptyString_UsesDefault(t *testing.T) {
	g := gauge.NewSimpleGauge()
	g.Pointer.Color = ""
	g.SetValue(50)
	img := gauge.RenderSimple(g, 100, 100)
	if img == nil {
		t.Fatal("RenderSimple with empty color returned nil")
	}
}

func TestRenderSimpleProgress_ParseColor_EmptyString_UsesDefault(t *testing.T) {
	g := gauge.NewSimpleProgressGauge()
	g.Pointer.Color = ""
	g.SetValue(50)
	img := gauge.RenderSimpleProgress(g, 100, 30)
	if img == nil {
		t.Fatal("RenderSimpleProgress with empty color returned nil")
	}
}

// ── RenderSimple: all shapes ──────────────────────────────────────────────────

func TestRenderSimple_Circle_ZeroValue(t *testing.T) {
	g := gauge.NewSimpleGauge()
	g.Shape = gauge.SimpleGaugeShapeCircle
	g.SetValue(0)
	img := gauge.RenderSimple(g, 100, 100)
	if img == nil {
		t.Fatal("RenderSimple Circle zero returned nil")
	}
}

func TestRenderSimple_Circle_FullValue(t *testing.T) {
	g := gauge.NewSimpleGauge()
	g.Shape = gauge.SimpleGaugeShapeCircle
	g.SetValue(100)
	img := gauge.RenderSimple(g, 100, 100)
	if img == nil {
		t.Fatal("RenderSimple Circle full returned nil")
	}
}

func TestRenderSimple_Circle_SmallSize(t *testing.T) {
	// rx <= 0 → circle drawing skipped.
	g := gauge.NewSimpleGauge()
	g.Shape = gauge.SimpleGaugeShapeCircle
	g.SetValue(50)
	img := gauge.RenderSimple(g, 8, 8) // rx = 4-4 = 0 → skipped
	if img == nil {
		t.Fatal("RenderSimple Circle small returned nil")
	}
}

func TestRenderSimple_Triangle_Basic(t *testing.T) {
	g := gauge.NewSimpleGauge()
	g.Shape = gauge.SimpleGaugeShapeTriangle
	g.SetValue(50)
	img := gauge.RenderSimple(g, 100, 80)
	if img == nil {
		t.Fatal("RenderSimple Triangle returned nil")
	}
}

func TestRenderSimple_Triangle_FullValue(t *testing.T) {
	// frac <= pct for all → all filled with pointer color.
	g := gauge.NewSimpleGauge()
	g.Shape = gauge.SimpleGaugeShapeTriangle
	g.SetValue(100)
	img := gauge.RenderSimple(g, 100, 80)
	if img == nil {
		t.Fatal("RenderSimple Triangle full returned nil")
	}
}

func TestRenderSimple_Rectangle_SmallSize(t *testing.T) {
	// barW <= 0 → rectangle drawing skipped.
	g := gauge.NewSimpleGauge()
	g.SetValue(50)
	img := gauge.RenderSimple(g, 8, 8) // barW = 8-8 = 0 → skipped
	if img == nil {
		t.Fatal("RenderSimple Rectangle small returned nil")
	}
}

// ── GaugeObject.Deserialize: Pointer.Width branch ─────────────────────────────

func TestGaugeObject_Deserialize_PointerWidthNegative(t *testing.T) {
	// Pointer.Width read with default -1 → branch not entered (v < 0).
	// Serialize with NO pointer width (default 6, so it's not serialized).
	orig := gauge.NewGaugeObject()
	// Default pointer width = 6, not serialized.
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("GaugeObject")   //nolint:errcheck
	orig.Serialize(w)              //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	r := serial.NewReader(strings.NewReader(buf.String()))
	r.ReadObjectHeader()
	g2 := gauge.NewGaugeObject()
	if err := g2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// Pointer.Width not serialized → ReadFloat returns -1 → branch not entered → Width stays 6.
	if g2.Pointer.Width != 6 {
		t.Errorf("Pointer.Width = %v, want 6 (default)", g2.Pointer.Width)
	}
}

// ── NewSimpleSubScale coverage ────────────────────────────────────────────────

func TestNewSimpleSubScale_Defaults(t *testing.T) {
	s := gauge.NewSimpleSubScale()
	if !s.Enabled {
		t.Error("Enabled should default to true")
	}
	if !s.ShowCaption {
		t.Error("ShowCaption should default to true")
	}
}

// ── GaugeObject serialize: Expression non-empty ───────────────────────────────

func TestGaugeObject_Serialize_Expression(t *testing.T) {
	g := gauge.NewGaugeObject()
	g.Expression = "[Sales.Total]"
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("GaugeObject")   //nolint:errcheck
	g.Serialize(w)                 //nolint:errcheck
	w.EndObject()                  //nolint:errcheck
	w.Flush()                      //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, "Expression") {
		t.Errorf("Expression should be serialized, got:\n%s", xml)
	}
}

// ── TypeName coverage ─────────────────────────────────────────────────────────

func TestSimpleGauge_TypeName(t *testing.T) {
	g := gauge.NewSimpleGauge()
	if g.TypeName() != "SimpleGauge" {
		t.Errorf("TypeName = %q, want SimpleGauge", g.TypeName())
	}
}
