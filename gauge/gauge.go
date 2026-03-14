// Package gauge implements gauge visualization objects for go-fastreport.
// It is the Go equivalent of FastReport.Gauge.
//
// Three gauge types are provided:
//   - LinearGauge: horizontal or vertical progress-bar style gauge
//   - RadialGauge: circular dial with a rotating pointer
//   - SimpleGauge: minimal single-value indicator
package gauge

import "fmt"

// ── Scale ─────────────────────────────────────────────────────────────────────

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
	Width float64
	// Color is the pointer color hex string (default: "#CC0000").
	Color string
}

// NewPointer creates a Pointer with defaults.
func NewPointer() *Pointer {
	return &Pointer{Width: 6, Color: "#CC0000"}
}

// ── GaugeObject (base) ────────────────────────────────────────────────────────

// GaugeObject is the base type for all gauge variants.
// It holds the value range and the current value.
type GaugeObject struct {
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

	// Expression is a report expression whose value drives the gauge.
	Expression string
	// Style is the named style (e.g. "Red", "Green").
	Style string
}

// NewGaugeObject creates a GaugeObject with defaults (0–100, value=0).
func NewGaugeObject() *GaugeObject {
	return &GaugeObject{
		Minimum: 0,
		Maximum: 100,
		Scale:   NewScale(),
		Pointer: NewPointer(),
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

// ── RadialGauge ───────────────────────────────────────────────────────────────

// RadialGauge is a circular dial gauge.
type RadialGauge struct {
	GaugeObject

	// StartAngle is the angle (degrees) at which the scale begins (default: -135).
	StartAngle float64
	// EndAngle is the angle (degrees) at which the scale ends (default: 135).
	EndAngle float64
}

// NewRadialGauge creates a RadialGauge with a 270-degree sweep (–135° to +135°).
func NewRadialGauge() *RadialGauge {
	return &RadialGauge{
		GaugeObject: *NewGaugeObject(),
		StartAngle:  -135,
		EndAngle:    135,
	}
}

// TypeName returns "RadialGauge".
func (g *RadialGauge) TypeName() string { return "RadialGauge" }

// NeedleAngle returns the current needle angle in degrees.
// It interpolates between StartAngle and EndAngle based on Percent().
func (g *RadialGauge) NeedleAngle() float64 {
	sweep := g.EndAngle - g.StartAngle
	return g.StartAngle + sweep*g.Percent()
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
}

// NewSimpleGauge creates a SimpleGauge with defaults.
func NewSimpleGauge() *SimpleGauge {
	return &SimpleGauge{
		GaugeObject: *NewGaugeObject(),
		Shape:       SimpleGaugeShapeRectangle,
		ShowText:    true,
		TextFormat:  "%g%%",
	}
}

// TypeName returns "SimpleGauge".
func (g *SimpleGauge) TypeName() string { return "SimpleGauge" }

// Text returns the formatted value text.
func (g *SimpleGauge) Text() string {
	if g.TextFormat == "" {
		return fmt.Sprintf("%g%%", g.Percent()*100)
	}
	return fmt.Sprintf(g.TextFormat, g.Percent()*100)
}
