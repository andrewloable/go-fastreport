package gauge_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/gauge"
)

// ── Scale ─────────────────────────────────────────────────────────────────────

func TestNewScale_Defaults(t *testing.T) {
	s := gauge.NewScale()
	if s.MinorStep != 1 {
		t.Errorf("MinorStep = %v, want 1", s.MinorStep)
	}
	if s.MajorStep != 10 {
		t.Errorf("MajorStep = %v, want 10", s.MajorStep)
	}
	if !s.ShowLabels {
		t.Error("ShowLabels should default to true")
	}
	if s.LabelFormat != "%g" {
		t.Errorf("LabelFormat = %q, want %%g", s.LabelFormat)
	}
}

func TestScale_FormatLabel(t *testing.T) {
	s := gauge.NewScale()
	if got := s.FormatLabel(42); got != "42" {
		t.Errorf("FormatLabel(42) = %q, want 42", got)
	}
}

func TestScale_FormatLabel_Custom(t *testing.T) {
	s := gauge.NewScale()
	s.LabelFormat = "%.1f"
	if got := s.FormatLabel(42); got != "42.0" {
		t.Errorf("FormatLabel(42) = %q, want 42.0", got)
	}
}

func TestScale_FormatLabel_EmptyFormat(t *testing.T) {
	s := gauge.NewScale()
	s.LabelFormat = ""
	if got := s.FormatLabel(42); got != "42" {
		t.Errorf("FormatLabel(42) = %q, want 42", got)
	}
}

// ── Pointer ───────────────────────────────────────────────────────────────────

func TestNewPointer_Defaults(t *testing.T) {
	p := gauge.NewPointer()
	if p.Width != 6 {
		t.Errorf("Width = %v, want 6", p.Width)
	}
	if p.Color != "#CC0000" {
		t.Errorf("Color = %q, want #CC0000", p.Color)
	}
}

// ── GaugeObject ───────────────────────────────────────────────────────────────

func TestNewGaugeObject_Defaults(t *testing.T) {
	g := gauge.NewGaugeObject()
	if g.Minimum != 0 {
		t.Errorf("Minimum = %v, want 0", g.Minimum)
	}
	if g.Maximum != 100 {
		t.Errorf("Maximum = %v, want 100", g.Maximum)
	}
	if g.Value() != 0 {
		t.Errorf("Value = %v, want 0", g.Value())
	}
	if g.Scale == nil {
		t.Error("Scale should not be nil")
	}
	if g.Pointer == nil {
		t.Error("Pointer should not be nil")
	}
}

func TestGaugeObject_SetValue(t *testing.T) {
	g := gauge.NewGaugeObject()
	g.SetValue(50)
	if g.Value() != 50 {
		t.Errorf("Value = %v, want 50", g.Value())
	}
}

func TestGaugeObject_SetValue_ClampMin(t *testing.T) {
	g := gauge.NewGaugeObject()
	g.SetValue(-10)
	if g.Value() != 0 {
		t.Errorf("Value below min should clamp to %v, got %v", g.Minimum, g.Value())
	}
}

func TestGaugeObject_SetValue_ClampMax(t *testing.T) {
	g := gauge.NewGaugeObject()
	g.SetValue(200)
	if g.Value() != 100 {
		t.Errorf("Value above max should clamp to %v, got %v", g.Maximum, g.Value())
	}
}

func TestGaugeObject_Percent(t *testing.T) {
	g := gauge.NewGaugeObject()
	g.SetValue(75)
	if got := g.Percent(); math.Abs(got-0.75) > 1e-9 {
		t.Errorf("Percent = %v, want 0.75", got)
	}
}

func TestGaugeObject_Percent_Zero(t *testing.T) {
	g := gauge.NewGaugeObject()
	g.SetValue(0)
	if g.Percent() != 0 {
		t.Errorf("Percent at min = %v, want 0", g.Percent())
	}
}

func TestGaugeObject_Percent_Full(t *testing.T) {
	g := gauge.NewGaugeObject()
	g.SetValue(100)
	if g.Percent() != 1 {
		t.Errorf("Percent at max = %v, want 1", g.Percent())
	}
}

func TestGaugeObject_Percent_ZeroRange(t *testing.T) {
	g := gauge.NewGaugeObject()
	g.Minimum = 5
	g.Maximum = 5
	g.SetValue(5)
	if g.Percent() != 0 {
		t.Errorf("Percent with zero range = %v, want 0", g.Percent())
	}
}

func TestGaugeObject_CustomRange(t *testing.T) {
	g := gauge.NewGaugeObject()
	g.Minimum = 10
	g.Maximum = 20
	g.SetValue(15)
	if math.Abs(g.Percent()-0.5) > 1e-9 {
		t.Errorf("Percent = %v, want 0.5", g.Percent())
	}
}

// ── LinearGauge ───────────────────────────────────────────────────────────────

func TestNewLinearGauge_Defaults(t *testing.T) {
	g := gauge.NewLinearGauge()
	if g.TypeName() != "LinearGauge" {
		t.Errorf("TypeName = %q", g.TypeName())
	}
	if g.Orientation != gauge.OrientationHorizontal {
		t.Error("default Orientation should be Horizontal")
	}
	if g.Inverted {
		t.Error("Inverted should default to false")
	}
}

func TestLinearGauge_FillPercent(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.SetValue(30)
	if math.Abs(g.FillPercent()-0.3) > 1e-9 {
		t.Errorf("FillPercent = %v, want 0.3", g.FillPercent())
	}
}

func TestLinearGauge_FillPercent_Inverted(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.Inverted = true
	g.SetValue(30)
	if math.Abs(g.FillPercent()-0.7) > 1e-9 {
		t.Errorf("FillPercent (inverted) = %v, want 0.7", g.FillPercent())
	}
}

func TestLinearGauge_Vertical(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.Orientation = gauge.OrientationVertical
	if g.Orientation != gauge.OrientationVertical {
		t.Error("Orientation should be Vertical")
	}
}

// ── RadialGauge ───────────────────────────────────────────────────────────────

func TestNewRadialGauge_Defaults(t *testing.T) {
	g := gauge.NewRadialGauge()
	if g.TypeName() != "RadialGauge" {
		t.Errorf("TypeName = %q", g.TypeName())
	}
	if g.StartAngle != -135 {
		t.Errorf("StartAngle = %v, want -135", g.StartAngle)
	}
	if g.EndAngle != 135 {
		t.Errorf("EndAngle = %v, want 135", g.EndAngle)
	}
}

func TestRadialGauge_NeedleAngle_AtMin(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.SetValue(0)
	if g.NeedleAngle() != -135 {
		t.Errorf("NeedleAngle at min = %v, want -135", g.NeedleAngle())
	}
}

func TestRadialGauge_NeedleAngle_AtMax(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.SetValue(100)
	if g.NeedleAngle() != 135 {
		t.Errorf("NeedleAngle at max = %v, want 135", g.NeedleAngle())
	}
}

func TestRadialGauge_NeedleAngle_AtMid(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.SetValue(50)
	// -135 + 270*0.5 = 0
	if math.Abs(g.NeedleAngle()-0) > 1e-9 {
		t.Errorf("NeedleAngle at mid = %v, want 0", g.NeedleAngle())
	}
}

func TestRadialGauge_CustomAngles(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.StartAngle = 0
	g.EndAngle = 360
	g.SetValue(50)
	if math.Abs(g.NeedleAngle()-180) > 1e-9 {
		t.Errorf("NeedleAngle = %v, want 180", g.NeedleAngle())
	}
}

// ── SimpleGauge ───────────────────────────────────────────────────────────────

func TestNewSimpleGauge_Defaults(t *testing.T) {
	g := gauge.NewSimpleGauge()
	if g.TypeName() != "SimpleGauge" {
		t.Errorf("TypeName = %q", g.TypeName())
	}
	if g.Shape != gauge.SimpleGaugeShapeRectangle {
		t.Error("default Shape should be Rectangle")
	}
	if !g.ShowText {
		t.Error("ShowText should default to true")
	}
}

func TestSimpleGauge_Text(t *testing.T) {
	g := gauge.NewSimpleGauge()
	g.SetValue(75)
	// Percent() = 0.75, TextFormat = "%g%%" → "75%"
	text := g.Text()
	if text != "75%" {
		t.Errorf("Text = %q, want 75%%", text)
	}
}

func TestSimpleGauge_Text_Empty(t *testing.T) {
	g := gauge.NewSimpleGauge()
	g.TextFormat = ""
	g.SetValue(50)
	text := g.Text()
	if text != "50%" {
		t.Errorf("Text = %q, want 50%%", text)
	}
}

func TestSimpleGauge_Text_Custom(t *testing.T) {
	g := gauge.NewSimpleGauge()
	g.TextFormat = "%.0f%%"
	g.SetValue(33)
	text := g.Text()
	if text != "33%" {
		t.Errorf("Text = %q, want 33%%", text)
	}
}

func TestSimpleGauge_Shapes(t *testing.T) {
	shapes := []gauge.SimpleGaugeShape{
		gauge.SimpleGaugeShapeRectangle,
		gauge.SimpleGaugeShapeCircle,
		gauge.SimpleGaugeShapeTriangle,
	}
	for _, sh := range shapes {
		g := gauge.NewSimpleGauge()
		g.Shape = sh
		if g.Shape != sh {
			t.Errorf("Shape = %v, want %v", g.Shape, sh)
		}
	}
}

// ── Style / Expression ────────────────────────────────────────────────────────

func TestGaugeObject_Style(t *testing.T) {
	g := gauge.NewGaugeObject()
	g.SetStyleName("Red")
	if g.StyleName() != "Red" {
		t.Errorf("StyleName = %q, want Red", g.StyleName())
	}
}

func TestGaugeObject_Expression(t *testing.T) {
	g := gauge.NewGaugeObject()
	g.Expression = "[DataSource.Value]"
	if g.Expression != "[DataSource.Value]" {
		t.Errorf("Expression = %q", g.Expression)
	}
}
