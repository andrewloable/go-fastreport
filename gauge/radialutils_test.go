package gauge_test

// radialutils_test.go — tests for radialutils.go helpers:
//   RotateVector, position/type predicates, radialStartAngleFor,
//   EffectiveStartAngle, Vertical, and new RadialGauge fields.
//
// C# reference: original-dotnet/FastReport.Base/Gauge/Radial/RadialUtils.cs

import (
	"bytes"
	"math"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── RotateVector ──────────────────────────────────────────────────────────────

func TestRotateVector_ZeroAngle(t *testing.T) {
	v := [2]gauge.Point2F{{X: 1, Y: 0}, {X: 2, Y: 0}}
	center := gauge.Point2F{X: 0, Y: 0}
	got := gauge.RotateVector(v, 0, center)
	if math.Abs(got[0].X-1) > 1e-9 || math.Abs(got[0].Y-0) > 1e-9 {
		t.Errorf("zero rotation: got (%g,%g), want (1,0)", got[0].X, got[0].Y)
	}
}

func TestRotateVector_90Degrees(t *testing.T) {
	// Rotating (1,0) around (0,0) by 90° should give approximately (0,1).
	// C# formula: x' = cx + dx*cos + (-dy)*sin, y' = cy + dx*sin + dy*cos
	// dx=1, dy=0, angle=pi/2: x'=0+1*0+0=0, y'=0+1*1+0=1
	v := [2]gauge.Point2F{{X: 1, Y: 0}, {X: 0, Y: 1}}
	center := gauge.Point2F{X: 0, Y: 0}
	angle := math.Pi / 2
	got := gauge.RotateVector(v, angle, center)
	if math.Abs(got[0].X-0) > 1e-9 || math.Abs(got[0].Y-1) > 1e-9 {
		t.Errorf("90° rotation of (1,0): got (%g,%g), want (0,1)", got[0].X, got[0].Y)
	}
}

func TestRotateVector_180Degrees(t *testing.T) {
	// Rotating (1,0) around (0,0) by 180° → (-1,0).
	v := [2]gauge.Point2F{{X: 1, Y: 0}, {X: 0, Y: 0}}
	center := gauge.Point2F{X: 0, Y: 0}
	got := gauge.RotateVector(v, math.Pi, center)
	if math.Abs(got[0].X-(-1)) > 1e-9 || math.Abs(got[0].Y-0) > 1e-9 {
		t.Errorf("180° rotation of (1,0): got (%g,%g), want (-1,0)", got[0].X, got[0].Y)
	}
}

func TestRotateVector_AroundNonOrigin(t *testing.T) {
	// Rotating (3,1) around (1,1) by 90°.
	// dx=2, dy=0 → x' = 1 + 2*0 + 0 = 1, y' = 1 + 2*1 + 0 = 3
	v := [2]gauge.Point2F{{X: 3, Y: 1}, {X: 0, Y: 0}}
	center := gauge.Point2F{X: 1, Y: 1}
	got := gauge.RotateVector(v, math.Pi/2, center)
	if math.Abs(got[0].X-1) > 1e-9 || math.Abs(got[0].Y-3) > 1e-9 {
		t.Errorf("90° rotation around (1,1): got (%g,%g), want (1,3)", got[0].X, got[0].Y)
	}
}

func TestRotateVector_BothPoints(t *testing.T) {
	// Ensure both elements of the result slice are computed independently.
	v := [2]gauge.Point2F{{X: 2, Y: 0}, {X: 0, Y: 2}}
	center := gauge.Point2F{X: 0, Y: 0}
	got := gauge.RotateVector(v, math.Pi/2, center)
	// (2,0) → (0,2)
	if math.Abs(got[0].X-0) > 1e-9 || math.Abs(got[0].Y-2) > 1e-9 {
		t.Errorf("point[0]: got (%g,%g)", got[0].X, got[0].Y)
	}
	// (0,2) → (−2,0)   dx=0,dy=2 → x'=0+0-(-2)*1... wait
	// dx=0, dy=2, cos=0, sin=1: x'= 0 + 0*0 + (-2)*1 = -2,  y'= 0 + 0*1 + 2*0 = 0
	if math.Abs(got[1].X-(-2)) > 1e-9 || math.Abs(got[1].Y-0) > 1e-9 {
		t.Errorf("point[1]: got (%g,%g), want (-2,0)", got[1].X, got[1].Y)
	}
}

// ── RadialGaugePosition predicates ───────────────────────────────────────────

func TestRadialGaugePosition_IsTop(t *testing.T) {
	if !gauge.RadialGaugePositionTop.IsTop() {
		t.Error("Top.IsTop() should be true")
	}
	if gauge.RadialGaugePositionBottom.IsTop() {
		t.Error("Bottom.IsTop() should be false")
	}
	if gauge.RadialGaugePositionNone.IsTop() {
		t.Error("None.IsTop() should be false")
	}
}

func TestRadialGaugePosition_IsBottom(t *testing.T) {
	if !gauge.RadialGaugePositionBottom.IsBottom() {
		t.Error("Bottom.IsBottom() should be true")
	}
	if gauge.RadialGaugePositionTop.IsBottom() {
		t.Error("Top.IsBottom() should be false")
	}
}

func TestRadialGaugePosition_IsLeft(t *testing.T) {
	if !gauge.RadialGaugePositionLeft.IsLeft() {
		t.Error("Left.IsLeft() should be true")
	}
	if gauge.RadialGaugePositionRight.IsLeft() {
		t.Error("Right.IsLeft() should be false")
	}
}

func TestRadialGaugePosition_IsRight(t *testing.T) {
	if !gauge.RadialGaugePositionRight.IsRight() {
		t.Error("Right.IsRight() should be true")
	}
	if gauge.RadialGaugePositionLeft.IsRight() {
		t.Error("Left.IsRight() should be false")
	}
}

func TestRadialGaugePosition_Combined(t *testing.T) {
	// Quadrant top-left uses both Top and Left flags.
	p := gauge.RadialGaugePositionTop | gauge.RadialGaugePositionLeft
	if !p.IsTop() || !p.IsLeft() {
		t.Error("Top|Left should satisfy both IsTop and IsLeft")
	}
	if p.IsBottom() || p.IsRight() {
		t.Error("Top|Left should NOT satisfy IsBottom or IsRight")
	}
}

// ── RadialGaugeType predicates ────────────────────────────────────────────────

func TestRadialGaugeType_IsSemicircle(t *testing.T) {
	if !gauge.RadialGaugeTypeSemicircle.IsSemicircle() {
		t.Error("Semicircle.IsSemicircle() should be true")
	}
	if gauge.RadialGaugeTypeCircle.IsSemicircle() {
		t.Error("Circle.IsSemicircle() should be false")
	}
	if gauge.RadialGaugeTypeQuadrant.IsSemicircle() {
		t.Error("Quadrant.IsSemicircle() should be false")
	}
}

func TestRadialGaugeType_IsQuadrant(t *testing.T) {
	if !gauge.RadialGaugeTypeQuadrant.IsQuadrant() {
		t.Error("Quadrant.IsQuadrant() should be true")
	}
	if gauge.RadialGaugeTypeCircle.IsQuadrant() {
		t.Error("Circle.IsQuadrant() should be false")
	}
}

// ── EffectiveStartAngle / radialStartAngleFor ─────────────────────────────────

func TestEffectiveStartAngle_Circle(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.StartAngle = -120
	if got := g.EffectiveStartAngle(); got != -120 {
		t.Errorf("Circle: got %g, want -120", got)
	}
}

func TestEffectiveStartAngle_SemicircleTop(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeSemicircle
	g.Position = gauge.RadialGaugePositionTop
	// C# startAngle for Semi+Top = -90°
	if got := g.EffectiveStartAngle(); got != -90 {
		t.Errorf("Semi+Top: got %g, want -90", got)
	}
}

func TestEffectiveStartAngle_SemicircleBottom(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeSemicircle
	g.Position = gauge.RadialGaugePositionBottom
	if got := g.EffectiveStartAngle(); got != -90 {
		t.Errorf("Semi+Bottom: got %g, want -90", got)
	}
}

func TestEffectiveStartAngle_SemicircleLeft(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeSemicircle
	g.Position = gauge.RadialGaugePositionLeft
	if got := g.EffectiveStartAngle(); got != -180 {
		t.Errorf("Semi+Left: got %g, want -180", got)
	}
}

func TestEffectiveStartAngle_SemicircleRight(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeSemicircle
	g.Position = gauge.RadialGaugePositionRight
	if got := g.EffectiveStartAngle(); got != -180 {
		t.Errorf("Semi+Right: got %g, want -180", got)
	}
}

func TestEffectiveStartAngle_QuadrantTopLeft(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeQuadrant
	g.Position = gauge.RadialGaugePositionTop | gauge.RadialGaugePositionLeft
	if got := g.EffectiveStartAngle(); got != -90 {
		t.Errorf("Quad+TopLeft: got %g, want -90", got)
	}
}

func TestEffectiveStartAngle_QuadrantTopRight(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeQuadrant
	g.Position = gauge.RadialGaugePositionTop | gauge.RadialGaugePositionRight
	if got := g.EffectiveStartAngle(); got != 90 {
		t.Errorf("Quad+TopRight: got %g, want 90", got)
	}
}

func TestEffectiveStartAngle_QuadrantBottomLeft(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeQuadrant
	g.Position = gauge.RadialGaugePositionBottom | gauge.RadialGaugePositionLeft
	if got := g.EffectiveStartAngle(); got != -180 {
		t.Errorf("Quad+BottomLeft: got %g, want -180", got)
	}
}

func TestEffectiveStartAngle_QuadrantBottomRight(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeQuadrant
	g.Position = gauge.RadialGaugePositionBottom | gauge.RadialGaugePositionRight
	if got := g.EffectiveStartAngle(); got != 180 {
		t.Errorf("Quad+BottomRight: got %g, want 180", got)
	}
}

// ── GaugeObject.Vertical ─────────────────────────────────────────────────────

func TestGaugeObject_Vertical_Horizontal(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.SetWidth(200)
	g.SetHeight(40)
	if g.Vertical() {
		t.Error("200×40 gauge should NOT be vertical")
	}
}

func TestGaugeObject_Vertical_Portrait(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.SetWidth(40)
	g.SetHeight(200)
	if !g.Vertical() {
		t.Error("40×200 gauge should be vertical")
	}
}

func TestGaugeObject_Vertical_Square(t *testing.T) {
	g := gauge.NewLinearGauge()
	g.SetWidth(100)
	g.SetHeight(100)
	// Width == Height → not strictly less than, so Vertical() should be false.
	if g.Vertical() {
		t.Error("100×100 gauge should NOT be vertical (Width == Height)")
	}
}

// ── RadialGauge new fields Serialize/Deserialize round-trips ─────────────────

func radialRoundTripFull(t *testing.T, orig *gauge.RadialGauge) *gauge.RadialGauge {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("RadialGauge") //nolint:errcheck
	orig.Serialize(w)            //nolint:errcheck
	w.EndObject()                //nolint:errcheck
	w.Flush()                    //nolint:errcheck

	r := serial.NewReader(strings.NewReader(buf.String()))
	r.ReadObjectHeader()
	g2 := gauge.NewRadialGauge()
	if err := g2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	return g2
}

func TestRadialGauge_Serialize_GaugeTypeSemicircle(t *testing.T) {
	orig := gauge.NewRadialGauge()
	orig.GaugeType = gauge.RadialGaugeTypeSemicircle
	orig.Position = gauge.RadialGaugePositionTop

	g2 := radialRoundTripFull(t, orig)
	if g2.GaugeType != gauge.RadialGaugeTypeSemicircle {
		t.Errorf("GaugeType = %v, want Semicircle", g2.GaugeType)
	}
	if g2.Position != gauge.RadialGaugePositionTop {
		t.Errorf("Position = %v, want Top", g2.Position)
	}
}

func TestRadialGauge_Serialize_GaugeTypeQuadrant(t *testing.T) {
	orig := gauge.NewRadialGauge()
	orig.GaugeType = gauge.RadialGaugeTypeQuadrant
	orig.Position = gauge.RadialGaugePositionTop | gauge.RadialGaugePositionLeft

	g2 := radialRoundTripFull(t, orig)
	if g2.GaugeType != gauge.RadialGaugeTypeQuadrant {
		t.Errorf("GaugeType = %v, want Quadrant", g2.GaugeType)
	}
	if !g2.Position.IsTop() || !g2.Position.IsLeft() {
		t.Errorf("Position = %v, want Top|Left", g2.Position)
	}
}

func TestRadialGauge_Serialize_SemicircleOffsetRatio(t *testing.T) {
	orig := gauge.NewRadialGauge()
	orig.SemicircleOffsetRatio = 1.5

	g2 := radialRoundTripFull(t, orig)
	if math.Abs(g2.SemicircleOffsetRatio-1.5) > 1e-4 {
		t.Errorf("SemicircleOffsetRatio = %g, want 1.5", g2.SemicircleOffsetRatio)
	}
}

func TestRadialGauge_Serialize_GradientAutoRotateFalse(t *testing.T) {
	orig := gauge.NewRadialGauge()
	orig.GradientAutoRotate = false

	g2 := radialRoundTripFull(t, orig)
	if g2.GradientAutoRotate {
		t.Error("GradientAutoRotate should be false after round-trip")
	}
}

func TestRadialGauge_DefaultsNotSerialized(t *testing.T) {
	// When all fields are at their defaults, Serialize should NOT write them.
	orig := gauge.NewRadialGauge() // GaugeType=Circle, Position=None, Offset=1, AutoRotate=true
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.BeginObject("RadialGauge") //nolint:errcheck
	orig.Serialize(w)            //nolint:errcheck
	w.EndObject()                //nolint:errcheck
	w.Flush()                    //nolint:errcheck
	xml := buf.String()
	if strings.Contains(xml, "GaugeType") {
		t.Error("default GaugeType should not be serialized")
	}
	if strings.Contains(xml, "Position") {
		t.Error("default Position should not be serialized")
	}
	if strings.Contains(xml, "SemicircleOffsetRatio") {
		t.Error("default SemicircleOffsetRatio should not be serialized")
	}
	if strings.Contains(xml, "GradientAutoRotate") {
		t.Error("default GradientAutoRotate should not be serialized")
	}
}

// ── RenderRadial: Semicircle and Quadrant branches ───────────────────────────

func TestRenderRadial_SemicircleTop(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeSemicircle
	g.Position = gauge.RadialGaugePositionTop
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial Semicircle+Top returned nil")
	}
}

func TestRenderRadial_SemicircleBottom(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeSemicircle
	g.Position = gauge.RadialGaugePositionBottom
	g.SetValue(75)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial Semicircle+Bottom returned nil")
	}
}

func TestRenderRadial_SemicircleLeft(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeSemicircle
	g.Position = gauge.RadialGaugePositionLeft
	g.SetValue(25)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial Semicircle+Left returned nil")
	}
}

func TestRenderRadial_SemicircleRight(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeSemicircle
	g.Position = gauge.RadialGaugePositionRight
	g.SetValue(100)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial Semicircle+Right returned nil")
	}
}

func TestRenderRadial_SemicircleDefault(t *testing.T) {
	// Position=None for semicircle uses default branch.
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeSemicircle
	g.Position = gauge.RadialGaugePositionNone
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial Semicircle+None returned nil")
	}
}

func TestRenderRadial_QuadrantTopLeft(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeQuadrant
	g.Position = gauge.RadialGaugePositionTop | gauge.RadialGaugePositionLeft
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial Quadrant+TopLeft returned nil")
	}
}

func TestRenderRadial_QuadrantTopRight(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeQuadrant
	g.Position = gauge.RadialGaugePositionTop | gauge.RadialGaugePositionRight
	g.SetValue(30)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial Quadrant+TopRight returned nil")
	}
}

func TestRenderRadial_QuadrantBottomLeft(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeQuadrant
	g.Position = gauge.RadialGaugePositionBottom | gauge.RadialGaugePositionLeft
	g.SetValue(70)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial Quadrant+BottomLeft returned nil")
	}
}

func TestRenderRadial_QuadrantBottomRight(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeQuadrant
	g.Position = gauge.RadialGaugePositionBottom | gauge.RadialGaugePositionRight
	g.SetValue(90)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial Quadrant+BottomRight returned nil")
	}
}

func TestRenderRadial_QuadrantDefault(t *testing.T) {
	// Position=None for quadrant uses default branch.
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeQuadrant
	g.Position = gauge.RadialGaugePositionNone
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial Quadrant+None returned nil")
	}
}

func TestRenderRadial_ScaleLabelsDisabled(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.Scale.ShowLabels = false
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial with ShowLabels=false returned nil")
	}
}

func TestRenderRadial_ScaleTickCustomColor(t *testing.T) {
	g := gauge.NewRadialGauge()
	g.Scale.MajorTicks.Color = "#FF0000"
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("RenderRadial with custom tick color returned nil")
	}
}

func TestRenderRadial_TinySize(t *testing.T) {
	// Tiny size should return early without panicking.
	g := gauge.NewRadialGauge()
	g.SetValue(50)
	img := gauge.RenderRadial(g, 4, 4)
	if img == nil {
		t.Fatal("RenderRadial tiny returned nil")
	}
}

func TestRenderRadial_SemicircleBottomNeedleInversion(t *testing.T) {
	// For Semicircle+Bottom the needle direction is inverted (dir = -1).
	// This exercises the dir=-1 branch in RenderRadial.
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeSemicircle
	g.Position = gauge.RadialGaugePositionBottom
	g.SetValue(0)
	img0 := gauge.RenderRadial(g, 100, 100)
	g.SetValue(100)
	img100 := gauge.RenderRadial(g, 100, 100)
	if img0 == nil || img100 == nil {
		t.Fatal("RenderRadial Semicircle+Bottom returned nil")
	}
}

func TestRenderRadial_QuadrantBottomRightNeedleInversion(t *testing.T) {
	// Quadrant+BottomRight triggers dir=-1 branch.
	g := gauge.NewRadialGauge()
	g.GaugeType = gauge.RadialGaugeTypeQuadrant
	g.Position = gauge.RadialGaugePositionBottom | gauge.RadialGaugePositionRight
	g.SetValue(50)
	img := gauge.RenderRadial(g, 100, 100)
	if img == nil {
		t.Fatal("nil image")
	}
}

// ── drawRadialScaleTicks edge cases ──────────────────────────────────────────

func TestRenderRadial_ScaleTicks_Smoke(t *testing.T) {
	// Smoke test covering drawRadialScaleTicks with a non-default arc.
	g := gauge.NewRadialGauge()
	g.StartAngle = -90
	g.EndAngle = 90
	g.SetValue(50)
	img := gauge.RenderRadial(g, 120, 120)
	if img == nil {
		t.Fatal("nil image")
	}
}
