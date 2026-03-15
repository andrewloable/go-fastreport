package gauge_test

import (
	"image"
	"testing"

	"github.com/andrewloable/go-fastreport/gauge"
)

func TestRenderLinear_Horizontal(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.SetValue(50)
	img := gauge.RenderLinear(g, 200, 40)
	if img == nil {
		t.Fatal("RenderLinear returned nil")
	}
	b := img.Bounds()
	if b.Dx() != 200 || b.Dy() != 40 {
		t.Errorf("unexpected size %dx%d", b.Dx(), b.Dy())
	}
}

func TestRenderLinear_Vertical(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.Orientation = gauge.OrientationVertical
	g.SetValue(75)
	img := gauge.RenderLinear(g, 40, 200)
	if img == nil {
		t.Fatal("RenderLinear vertical returned nil")
	}
}

func TestRenderLinear_ZeroValue(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.SetValue(0)
	img := gauge.RenderLinear(g, 100, 20)
	if img == nil {
		t.Fatal("RenderLinear zero value returned nil")
	}
}

func TestRenderLinear_FullValue(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.SetValue(100)
	img := gauge.RenderLinear(g, 100, 20)
	if img == nil {
		t.Fatal("RenderLinear full value returned nil")
	}
}

func TestRenderLinear_Inverted(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.Inverted = true
	g.SetValue(25)
	img := gauge.RenderLinear(g, 100, 20)
	if img == nil {
		t.Fatal("RenderLinear inverted returned nil")
	}
}

func TestRenderRadial(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.SetValue(50)
	img := gauge.RenderRadial(g, 200, 200)
	if img == nil {
		t.Fatal("RenderRadial returned nil")
	}
	b := img.Bounds()
	if b.Dx() != 200 || b.Dy() != 200 {
		t.Errorf("unexpected size %dx%d", b.Dx(), b.Dy())
	}
}

func TestRenderRadial_MinValue(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.SetValue(0)
	img := gauge.RenderRadial(g, 100, 100)
	if _, ok := img.(*image.RGBA); !ok {
		t.Error("expected *image.RGBA")
	}
}

func TestRenderRadial_MaxValue(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.SetValue(100)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial max value returned nil")
	}
}

func TestRenderSimple_Rectangle(t *testing.T) {
	g := gauge.NewSimpleGauge()
	g.Shape = gauge.SimpleGaugeShapeRectangle
	g.SetValue(60)
	img := gauge.RenderSimple(g, 100, 100)
	if img == nil {
		t.Fatal("RenderSimple rectangle returned nil")
	}
}

func TestRenderSimple_Circle(t *testing.T) {
	g := gauge.NewSimpleGauge()
	g.Shape = gauge.SimpleGaugeShapeCircle
	g.SetValue(30)
	img := gauge.RenderSimple(g, 100, 100)
	if img == nil {
		t.Fatal("RenderSimple circle returned nil")
	}
}

func TestRenderSimple_Triangle(t *testing.T) {
	g := gauge.NewSimpleGauge()
	g.Shape = gauge.SimpleGaugeShapeTriangle
	g.SetValue(80)
	img := gauge.RenderSimple(g, 100, 100)
	if img == nil {
		t.Fatal("RenderSimple triangle returned nil")
	}
}

func TestRenderSimpleProgress(t *testing.T) {
	g := gauge.NewSimpleProgressGauge()
	g.SetValue(40)
	img := gauge.RenderSimpleProgress(g, 200, 30)
	if img == nil {
		t.Fatal("RenderSimpleProgress returned nil")
	}
	b := img.Bounds()
	if b.Dx() != 200 || b.Dy() != 30 {
		t.Errorf("unexpected size %dx%d", b.Dx(), b.Dy())
	}
}

func TestRenderLinear_CustomColor(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.Pointer.Color = "#0066CC"
	g.SetValue(50)
	img := gauge.RenderLinear(g, 100, 20)
	if img == nil {
		t.Fatal("RenderLinear custom color returned nil")
	}
}

func TestRenderRadial_CustomAngles(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.StartAngle = -90
	g.EndAngle = 90
	g.SetValue(50)
	img := gauge.RenderRadial(g, 150, 150)
	if img == nil {
		t.Fatal("RenderRadial custom angles returned nil")
	}
}
