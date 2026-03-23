// Package gauge implements gauge visualization objects for go-fastreport.
// It is the Go equivalent of FastReport.Gauge.
//
// Three gauge types are provided:
//   - LinearGauge: horizontal or vertical progress-bar style gauge
//   - RadialGauge: circular dial with a rotating pointer
//   - SimpleGauge: minimal single-value indicator
//   - SimpleProgressGauge: simplified progress indicator
package gauge

import (
	"fmt"

	"github.com/andrewloable/go-fastreport/report"
)

// ── Scale ─────────────────────────────────────────────────────────────────────

// GaugeTicks holds tick appearance properties for major or minor ticks.
type GaugeTicks struct {
	// Width is the tick line width in pixels.
	Width float32
	// Height / Length of the tick mark.
	Length float32
	// Color is the tick color as a string (e.g. "128, 128, 128").
	Color string
}

// Scale holds the visual and numeric settings for the gauge scale.
type Scale struct {
	// MinorStep is the minor tick interval (default: 1).
	MinorStep float64
	// MajorStep is the major tick interval (default: 10).
	MajorStep float64
	// ShowLabels controls whether labels are drawn on major ticks.
	ShowLabels bool
	// LabelFormat is a fmt.Sprintf format string for labels (default: "%g").
	LabelFormat string
	// Font is the scale font descriptor string (e.g. "Arial, 8pt").
	Font string
	// MajorTicks holds major tick appearance.
	MajorTicks GaugeTicks
	// MinorTicks holds minor tick appearance.
	MinorTicks GaugeTicks
}

// NewScale creates a Scale with sensible defaults.
func NewScale() *Scale {
	return &Scale{
		MinorStep:   1,
		MajorStep:   10,
		ShowLabels:  true,
		LabelFormat: "%g",
	}
}

// Assign copies all fields from src into this Scale.
// Mirrors C# GaugeScale.Assign(GaugeScale src) (GaugeScale.cs:102-107).
func (s *Scale) Assign(src *Scale) {
	if src == nil {
		return
	}
	s.MinorStep = src.MinorStep
	s.MajorStep = src.MajorStep
	s.ShowLabels = src.ShowLabels
	s.LabelFormat = src.LabelFormat
	s.Font = src.Font
	s.MajorTicks = src.MajorTicks
	s.MinorTicks = src.MinorTicks
}

// FormatLabel formats a value according to LabelFormat.
func (s *Scale) FormatLabel(v float64) string {
	if s.LabelFormat == "" {
		return fmt.Sprintf("%g", v)
	}
	return fmt.Sprintf(s.LabelFormat, v)
}

// ── Pointer ───────────────────────────────────────────────────────────────────

// Pointer holds settings for the gauge pointer (needle or bar fill).
type Pointer struct {
	// Width is the pointer width in pixels (default: 6).
	Width float32
	// Height is the pointer height in pixels.
	Height float32
	// Color is the pointer color string (default: "204, 0, 0").
	Color string
}

// NewPointer creates a Pointer with defaults.
func NewPointer() *Pointer {
	return &Pointer{Width: 6, Color: "#CC0000"}
}

// ── Label ─────────────────────────────────────────────────────────────────────

// GaugeLabel holds label display properties for a gauge.
type GaugeLabel struct {
	// Font is the label font descriptor string.
	Font string
	// Text is a static label text override.
	Text string
}

// ── SimpleSubScale ────────────────────────────────────────────────────────────

// SimpleSubScale controls one of the two subscales on a SimpleGauge/SimpleScale.
// The first subscale appears above (or left of) the pointer; the second below
// (or right of) it.
type SimpleSubScale struct {
	// Enabled controls whether this subscale is drawn (default: true).
	Enabled bool
	// ShowCaption controls whether tick labels are drawn (default: true).
	ShowCaption bool
}

// NewSimpleSubScale creates a SimpleSubScale with defaults (both enabled).
func NewSimpleSubScale() SimpleSubScale {
	return SimpleSubScale{Enabled: true, ShowCaption: true}
}

// ── GaugeObject (base) ────────────────────────────────────────────────────────

// GaugeObject is the base type for all gauge variants.
// It embeds ReportComponentBase to satisfy report.Base, and holds the value
// range and the current value.
type GaugeObject struct {
	report.ReportComponentBase

	// Minimum is the lower bound of the gauge range (default: 0).
	Minimum float64
	// Maximum is the upper bound of the gauge range (default: 100).
	Maximum float64
	// value is the current gauge value (clamped to [Minimum, Maximum]).
	value float64

	// Scale holds scale appearance settings.
	Scale *Scale
	// Pointer holds pointer appearance settings.
	Pointer *Pointer
	// Label holds label display settings.
	Label GaugeLabel

	// Expression is a report expression whose value drives the gauge.
	Expression string
}

// NewGaugeObject creates a GaugeObject with defaults (0–100, value=0).
func NewGaugeObject() *GaugeObject {
	return &GaugeObject{
		ReportComponentBase: *report.NewReportComponentBase(),
		Minimum:             0,
		Maximum:             100,
		Scale:               NewScale(),
		Pointer:             NewPointer(),
	}
}

// Value returns the current gauge value.
func (g *GaugeObject) Value() float64 { return g.value }

// SetValue sets the gauge value, clamped to [Minimum, Maximum].
func (g *GaugeObject) SetValue(v float64) {
	if v < g.Minimum {
		v = g.Minimum
	}
	if v > g.Maximum {
		v = g.Maximum
	}
	g.value = v
}

// Percent returns the gauge value as a percentage of the range [0, 1].
// Returns 0 if the range is zero.
func (g *GaugeObject) Percent() float64 {
	span := g.Maximum - g.Minimum
	if span == 0 {
		return 0
	}
	return (g.value - g.Minimum) / span
}

// Vertical reports whether the gauge is taller than it is wide.
// Mirrors C# GaugeObject.Vertical property: Width < Height.
func (g *GaugeObject) Vertical() bool {
	return g.Width() < g.Height()
}

// Serialize writes GaugeObject properties that differ from defaults.
func (g *GaugeObject) Serialize(w report.Writer) error {
	if err := g.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if g.Minimum != 0 {
		w.WriteFloat("Minimum", float32(g.Minimum))
	}
	if g.Maximum != 100 {
		w.WriteFloat("Maximum", float32(g.Maximum))
	}
	if g.value != 0 {
		w.WriteFloat("Value", float32(g.value))
	}
	if g.Expression != "" {
		w.WriteStr("Expression", g.Expression)
	}
	// Scale dot-notation properties.
	if g.Scale != nil {
		if g.Scale.Font != "" {
			w.WriteStr("Scale.Font", g.Scale.Font)
		}
		if g.Scale.MajorTicks.Width != 0 {
			w.WriteFloat("Scale.MajorTicks.Width", g.Scale.MajorTicks.Width)
		}
		if g.Scale.MajorTicks.Color != "" {
			w.WriteStr("Scale.MajorTicks.Color", g.Scale.MajorTicks.Color)
		}
		if g.Scale.MajorTicks.Length != 0 {
			w.WriteFloat("Scale.MajorTicks.Length", g.Scale.MajorTicks.Length)
		}
		if g.Scale.MinorTicks.Width != 0 {
			w.WriteFloat("Scale.MinorTicks.Width", g.Scale.MinorTicks.Width)
		}
		if g.Scale.MinorTicks.Color != "" {
			w.WriteStr("Scale.MinorTicks.Color", g.Scale.MinorTicks.Color)
		}
		if g.Scale.MinorTicks.Length != 0 {
			w.WriteFloat("Scale.MinorTicks.Length", g.Scale.MinorTicks.Length)
		}
	}
	// Pointer dot-notation properties.
	if g.Pointer != nil {
		if g.Pointer.Width != 6 {
			w.WriteFloat("Pointer.Width", g.Pointer.Width)
		}
		if g.Pointer.Height != 0 {
			w.WriteFloat("Pointer.Height", g.Pointer.Height)
		}
		if g.Pointer.Color != "" && g.Pointer.Color != "#CC0000" {
			w.WriteStr("Pointer.Color", g.Pointer.Color)
		}
	}
	// Label dot-notation properties.
	if g.Label.Font != "" {
		w.WriteStr("Label.Font", g.Label.Font)
	}
	if g.Label.Text != "" {
		w.WriteStr("Label.Text", g.Label.Text)
	}
	return nil
}

// Deserialize reads GaugeObject properties.
func (g *GaugeObject) Deserialize(r report.Reader) error {
	if err := g.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	g.Minimum = float64(r.ReadFloat("Minimum", 0))
	g.Maximum = float64(r.ReadFloat("Maximum", 100))
	g.value = float64(r.ReadFloat("Value", 0))
	g.Expression = r.ReadStr("Expression", "")
	// Scale dot-notation properties.
	if g.Scale == nil {
		g.Scale = NewScale()
	}
	g.Scale.Font = r.ReadStr("Scale.Font", "")
	g.Scale.MajorTicks.Width = r.ReadFloat("Scale.MajorTicks.Width", 0)
	g.Scale.MajorTicks.Color = r.ReadStr("Scale.MajorTicks.Color", "")
	g.Scale.MajorTicks.Length = r.ReadFloat("Scale.MajorTicks.Length", 0)
	g.Scale.MinorTicks.Width = r.ReadFloat("Scale.MinorTicks.Width", 0)
	g.Scale.MinorTicks.Color = r.ReadStr("Scale.MinorTicks.Color", "")
	g.Scale.MinorTicks.Length = r.ReadFloat("Scale.MinorTicks.Length", 0)
	// Pointer dot-notation properties.
	if g.Pointer == nil {
		g.Pointer = NewPointer()
	}
	if v := r.ReadFloat("Pointer.Width", -1); v >= 0 {
		g.Pointer.Width = v
	}
	g.Pointer.Height = r.ReadFloat("Pointer.Height", 0)
	if c := r.ReadStr("Pointer.Color", ""); c != "" {
		g.Pointer.Color = c
	}
	// Label dot-notation properties.
	g.Label.Font = r.ReadStr("Label.Font", "")
	g.Label.Text = r.ReadStr("Label.Text", "")
	return nil
}

// GetExpressions returns the list of expressions used by this GaugeObject:
// the base component expressions plus Expression when non-empty.
// Mirrors C# GaugeObject.GetExpressions (GaugeObject.cs:208-217).
func (g *GaugeObject) GetExpressions() []string {
	exprs := g.ReportComponentBase.GetExpressions()
	if g.Expression != "" {
		exprs = append(exprs, g.Expression)
	}
	return exprs
}

// Clone returns a deep copy of this GaugeObject.
// Mirrors C# GaugeObject.Clone() via Base.Clone() which calls Assign().
func (g *GaugeObject) Clone() *GaugeObject {
	dst := &GaugeObject{
		ReportComponentBase: *report.NewReportComponentBase(),
		Scale:               NewScale(),
		Pointer:             NewPointer(),
	}
	dst.Assign(g)
	return dst
}

// Assign copies all GaugeObject fields from src into this GaugeObject.
// Mirrors C# GaugeObject.Assign(Base source) (GaugeObject.cs:251-262).
func (g *GaugeObject) Assign(src *GaugeObject) {
	if src == nil {
		return
	}
	g.ReportComponentBase = src.ReportComponentBase
	g.Minimum = src.Minimum
	g.Maximum = src.Maximum
	g.value = src.value
	g.Expression = src.Expression
	if src.Scale != nil {
		if g.Scale == nil {
			g.Scale = NewScale()
		}
		g.Scale.Assign(src.Scale)
	}
	if src.Pointer != nil {
		if g.Pointer == nil {
			g.Pointer = NewPointer()
		}
		*g.Pointer = *src.Pointer
	}
	g.Label = src.Label
}

// ── LinearGauge ───────────────────────────────────────────────────────────────

// Orientation controls the direction of a linear gauge.
type Orientation int

const (
	OrientationHorizontal Orientation = iota
	OrientationVertical
)

// LinearGauge is a progress-bar style gauge.
type LinearGauge struct {
	GaugeObject

	// Orientation controls horizontal or vertical layout (default: Horizontal).
	Orientation Orientation
	// Inverted reverses the fill direction (right-to-left or top-to-bottom).
	Inverted bool
}

// NewLinearGauge creates a LinearGauge with defaults.
func NewLinearGauge() *LinearGauge {
	return &LinearGauge{
		GaugeObject: *NewGaugeObject(),
		Orientation: OrientationHorizontal,
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (g *LinearGauge) BaseName() string { return "LinearGauge" }

// TypeName returns "LinearGauge".
func (g *LinearGauge) TypeName() string { return "LinearGauge" }

// FillPercent returns the fill percentage for rendering.
// When Inverted, it is 1 - Percent().
func (g *LinearGauge) FillPercent() float64 {
	p := g.Percent()
	if g.Inverted {
		return 1 - p
	}
	return p
}

// Serialize writes LinearGauge properties.
func (g *LinearGauge) Serialize(w report.Writer) error {
	if err := g.GaugeObject.Serialize(w); err != nil {
		return err
	}
	if g.Orientation != OrientationHorizontal {
		w.WriteInt("Orientation", int(g.Orientation))
	}
	if g.Inverted {
		w.WriteBool("Inverted", true)
	}
	return nil
}

// Deserialize reads LinearGauge properties.
func (g *LinearGauge) Deserialize(r report.Reader) error {
	if err := g.GaugeObject.Deserialize(r); err != nil {
		return err
	}
	g.Orientation = Orientation(r.ReadInt("Orientation", 0))
	g.Inverted = r.ReadBool("Inverted", false)
	return nil
}

// Assign copies all LinearGauge fields from src into this LinearGauge.
// Mirrors C# LinearGauge.Assign(Base source) (LinearGauge.cs:63-67).
func (g *LinearGauge) Assign(src *LinearGauge) {
	if src == nil {
		return
	}
	g.GaugeObject.Assign(&src.GaugeObject)
	g.Orientation = src.Orientation
	g.Inverted = src.Inverted
}

// ── RadialGauge ───────────────────────────────────────────────────────────────

// RadialGauge is a circular dial gauge.
// It supports three shape types (Circle, Semicircle, Quadrant) and four position
// orientations (Top, Bottom, Left, Right) that mirror C# RadialGauge.
//
// C# source: original-dotnet/FastReport.Base/Gauge/Radial/RadialGauge.cs
type RadialGauge struct {
	GaugeObject

	// StartAngle is the angle (degrees) at which the scale begins (default: -135).
	StartAngle float64
	// EndAngle is the angle (degrees) at which the scale ends (default: 135).
	EndAngle float64

	// GaugeType controls the shape of the dial (Circle / Semicircle / Quadrant).
	// Mirrors C# RadialGauge.Type.  Default: RadialGaugeTypeCircle.
	GaugeType RadialGaugeType

	// Position controls the flat-edge orientation for Semicircle and Quadrant types.
	// Mirrors C# RadialGauge.Position.  Default: RadialGaugePositionNone.
	Position RadialGaugePosition

	// SemicircleOffsetRatio is a multiplier for the straight-edge fill offset used
	// when rendering Semicircle Left/Right orientations.
	// Mirrors C# RadialGauge.SemicircleOffsetRatio.  Default: 1.
	SemicircleOffsetRatio float64

	// GradientAutoRotate, when true, rotates the pointer fill gradient automatically
	// to match the needle angle.
	// Mirrors C# RadialPointer.GradientAutoRotate.  Default: true.
	GradientAutoRotate bool
}

// NewRadialGauge creates a RadialGauge with a 270-degree sweep (–135° to +135°).
func NewRadialGauge() *RadialGauge {
	return &RadialGauge{
		GaugeObject:           *NewGaugeObject(),
		StartAngle:            -135,
		EndAngle:              135,
		GaugeType:             RadialGaugeTypeCircle,
		Position:              RadialGaugePositionNone,
		SemicircleOffsetRatio: 1,
		GradientAutoRotate:    true,
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (g *RadialGauge) BaseName() string { return "RadialGauge" }

// TypeName returns "RadialGauge".
func (g *RadialGauge) TypeName() string { return "RadialGauge" }

// NeedleAngle returns the current needle angle in degrees.
// It interpolates between StartAngle and EndAngle based on Percent().
func (g *RadialGauge) NeedleAngle() float64 {
	sweep := g.EndAngle - g.StartAngle
	return g.StartAngle + sweep*g.Percent()
}

// EffectiveStartAngle returns the needle start angle that matches the C# pointer
// rendering logic for the current GaugeType / Position combination.
// When the type is Circle, it falls back to StartAngle.
func (g *RadialGauge) EffectiveStartAngle() float64 {
	if g.GaugeType == RadialGaugeTypeCircle {
		return g.StartAngle
	}
	return radialStartAngleFor(g.GaugeType, g.Position)
}

// Serialize writes RadialGauge properties.
func (g *RadialGauge) Serialize(w report.Writer) error {
	if err := g.GaugeObject.Serialize(w); err != nil {
		return err
	}
	if g.StartAngle != -135 {
		w.WriteFloat("StartAngle", float32(g.StartAngle))
	}
	if g.EndAngle != 135 {
		w.WriteFloat("EndAngle", float32(g.EndAngle))
	}
	if g.GaugeType != RadialGaugeTypeCircle {
		w.WriteInt("GaugeType", int(g.GaugeType))
	}
	if g.Position != RadialGaugePositionNone {
		w.WriteInt("Position", int(g.Position))
	}
	if g.SemicircleOffsetRatio != 1 {
		w.WriteFloat("SemicircleOffsetRatio", float32(g.SemicircleOffsetRatio))
	}
	if !g.GradientAutoRotate {
		w.WriteBool("GradientAutoRotate", false)
	}
	return nil
}

// Deserialize reads RadialGauge properties.
func (g *RadialGauge) Deserialize(r report.Reader) error {
	if err := g.GaugeObject.Deserialize(r); err != nil {
		return err
	}
	g.StartAngle = float64(r.ReadFloat("StartAngle", -135))
	g.EndAngle = float64(r.ReadFloat("EndAngle", 135))
	g.GaugeType = RadialGaugeType(r.ReadInt("GaugeType", int(RadialGaugeTypeCircle)))
	g.Position = RadialGaugePosition(r.ReadInt("Position", int(RadialGaugePositionNone)))
	g.SemicircleOffsetRatio = float64(r.ReadFloat("SemicircleOffsetRatio", 1))
	g.GradientAutoRotate = r.ReadBool("GradientAutoRotate", true)
	return nil
}

// Assign copies all RadialGauge fields from src into this RadialGauge.
// Mirrors C# RadialGauge.Assign(Base source) (RadialGauge.cs:245-250).
func (g *RadialGauge) Assign(src *RadialGauge) {
	if src == nil {
		return
	}
	g.GaugeObject.Assign(&src.GaugeObject)
	g.StartAngle = src.StartAngle
	g.EndAngle = src.EndAngle
	g.GaugeType = src.GaugeType
	g.Position = src.Position
	g.SemicircleOffsetRatio = src.SemicircleOffsetRatio
	g.GradientAutoRotate = src.GradientAutoRotate
}

// ── SimpleGauge ───────────────────────────────────────────────────────────────

// SimpleGaugeShape controls the appearance of a SimpleGauge.
type SimpleGaugeShape int

const (
	SimpleGaugeShapeRectangle SimpleGaugeShape = iota
	SimpleGaugeShapeCircle
	SimpleGaugeShapeTriangle
)

// SimpleGauge is a minimal single-value indicator.
type SimpleGauge struct {
	GaugeObject

	// Shape controls the visual shape (default: Rectangle).
	Shape SimpleGaugeShape
	// ShowText controls whether the value is displayed as text.
	ShowText bool
	// TextFormat is a fmt.Sprintf format string for the value text.
	TextFormat string
	// FirstSubScale controls the top/left subscale on the scale track.
	FirstSubScale SimpleSubScale
	// SecondSubScale controls the bottom/right subscale on the scale track.
	SecondSubScale SimpleSubScale
}

// NewSimpleGauge creates a SimpleGauge with defaults.
func NewSimpleGauge() *SimpleGauge {
	return &SimpleGauge{
		GaugeObject:    *NewGaugeObject(),
		Shape:          SimpleGaugeShapeRectangle,
		ShowText:       true,
		TextFormat:     "%g%%",
		FirstSubScale:  NewSimpleSubScale(),
		SecondSubScale: NewSimpleSubScale(),
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (g *SimpleGauge) BaseName() string { return "SimpleGauge" }

// TypeName returns "SimpleGauge".
func (g *SimpleGauge) TypeName() string { return "SimpleGauge" }

// Text returns the formatted value text.
func (g *SimpleGauge) Text() string {
	if g.TextFormat == "" {
		return fmt.Sprintf("%g%%", g.Percent()*100)
	}
	return fmt.Sprintf(g.TextFormat, g.Percent()*100)
}

// Serialize writes SimpleGauge properties.
func (g *SimpleGauge) Serialize(w report.Writer) error {
	if err := g.GaugeObject.Serialize(w); err != nil {
		return err
	}
	if g.Shape != SimpleGaugeShapeRectangle {
		w.WriteInt("Shape", int(g.Shape))
	}
	if !g.ShowText {
		w.WriteBool("ShowText", false)
	}
	if g.TextFormat != "%g%%" {
		w.WriteStr("TextFormat", g.TextFormat)
	}
	// FirstSubScale (defaults: Enabled=true, ShowCaption=true).
	if !g.FirstSubScale.Enabled {
		w.WriteBool("Scale.FirstSubScale.Enabled", false)
	}
	if !g.FirstSubScale.ShowCaption {
		w.WriteBool("Scale.FirstSubScale.ShowCaption", false)
	}
	// SecondSubScale.
	if !g.SecondSubScale.Enabled {
		w.WriteBool("Scale.SecondSubScale.Enabled", false)
	}
	if !g.SecondSubScale.ShowCaption {
		w.WriteBool("Scale.SecondSubScale.ShowCaption", false)
	}
	return nil
}

// Deserialize reads SimpleGauge properties.
func (g *SimpleGauge) Deserialize(r report.Reader) error {
	if err := g.GaugeObject.Deserialize(r); err != nil {
		return err
	}
	g.Shape = SimpleGaugeShape(r.ReadInt("Shape", 0))
	g.ShowText = r.ReadBool("ShowText", true)
	g.TextFormat = r.ReadStr("TextFormat", "%g%%")
	// FirstSubScale.
	g.FirstSubScale.Enabled = r.ReadBool("Scale.FirstSubScale.Enabled", true)
	g.FirstSubScale.ShowCaption = r.ReadBool("Scale.FirstSubScale.ShowCaption", true)
	// SecondSubScale.
	g.SecondSubScale.Enabled = r.ReadBool("Scale.SecondSubScale.Enabled", true)
	g.SecondSubScale.ShowCaption = r.ReadBool("Scale.SecondSubScale.ShowCaption", true)
	return nil
}

// Assign copies all SimpleGauge fields from src into this SimpleGauge.
// C# SimpleGauge has no override of Assign; base GaugeObject.Assign is used.
// Go adds type-specific fields (Shape, ShowText, TextFormat, subscales).
func (g *SimpleGauge) Assign(src *SimpleGauge) {
	if src == nil {
		return
	}
	g.GaugeObject.Assign(&src.GaugeObject)
	g.Shape = src.Shape
	g.ShowText = src.ShowText
	g.TextFormat = src.TextFormat
	g.FirstSubScale = src.FirstSubScale
	g.SecondSubScale = src.SecondSubScale
}

// ── SimpleProgressGauge ───────────────────────────────────────────────────────

// SimpleProgressGauge is a simplified horizontal progress bar gauge.
// It is the Go equivalent of FastReport.Gauge.SimpleProgressGauge.
type SimpleProgressGauge struct {
	GaugeObject

	// ShowText controls whether the percentage text is displayed.
	ShowText bool
}

// NewSimpleProgressGauge creates a SimpleProgressGauge with defaults.
func NewSimpleProgressGauge() *SimpleProgressGauge {
	return &SimpleProgressGauge{
		GaugeObject: *NewGaugeObject(),
		ShowText:    true,
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (g *SimpleProgressGauge) BaseName() string { return "SimpleProgressGauge" }

// TypeName returns "SimpleProgressGauge".
func (g *SimpleProgressGauge) TypeName() string { return "SimpleProgressGauge" }

// Serialize writes SimpleProgressGauge properties.
func (g *SimpleProgressGauge) Serialize(w report.Writer) error {
	if err := g.GaugeObject.Serialize(w); err != nil {
		return err
	}
	if !g.ShowText {
		w.WriteBool("ShowText", false)
	}
	return nil
}

// Deserialize reads SimpleProgressGauge properties.
func (g *SimpleProgressGauge) Deserialize(r report.Reader) error {
	if err := g.GaugeObject.Deserialize(r); err != nil {
		return err
	}
	g.ShowText = r.ReadBool("ShowText", true)
	return nil
}
