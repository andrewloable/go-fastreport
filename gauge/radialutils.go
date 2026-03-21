package gauge

// radialutils.go — Go port of FastReport.Gauge.Radial.RadialUtils.
// Provides math helpers for rotating 2-D vectors and position/type predicates
// for RadialGauge variants (Circle, Semicircle, Quadrant).
//
// C# source: original-dotnet/FastReport.Base/Gauge/Radial/RadialUtils.cs

import "math"

// Point2F is a 2-D float point, analogous to System.Drawing.PointF.
type Point2F struct {
	X, Y float64
}

// RotateVector rotates a pair of points around a centre by angle radians.
// Mirrors C# RadialUtils.RotateVector(PointF[], double, PointF).
//
// Rotation formulae (standard 2-D rotation):
//   rx = cx + (px - cx)*cos(a) + (cy - py)*sin(a)   -- note sign: C# uses (cy-py)*sin
//   ry = cy + (px - cx)*sin(a) + (py - cy)*cos(a)
func RotateVector(v [2]Point2F, angle float64, center Point2F) [2]Point2F {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	var out [2]Point2F
	for i := 0; i < 2; i++ {
		dx := v[i].X - center.X
		dy := v[i].Y - center.Y
		// C# formula: x' = cx + dx*cos + (cy - y)*sin  =  cx + dx*cos + (-dy)*sin... wait
		// From RadialUtils.cs lines 21-24:
		//   rotated[i].X = cx + (px-cx)*cos + (cy-py)*sin
		//   rotated[i].Y = cy + (px-cx)*sin + (py-cy)*cos
		// (cy - py) == -dy, (py - cy) == dy
		out[i].X = center.X + dx*cos + (-dy)*sin
		out[i].Y = center.Y + dx*sin + dy*cos
	}
	return out
}

// ── RadialGaugeType ───────────────────────────────────────────────────────────

// RadialGaugeType enumerates the shape of a radial gauge.
// Matches C# enum RadialGaugeType (Radial/RadialGauge.cs).
type RadialGaugeType int

const (
	// RadialGaugeTypeCircle is a full 360° dial (default).
	RadialGaugeTypeCircle RadialGaugeType = 1
	// RadialGaugeTypeSemicircle is a 180° half-dial.
	RadialGaugeTypeSemicircle RadialGaugeType = 2
	// RadialGaugeTypeQuadrant is a 90° quarter-dial.
	RadialGaugeTypeQuadrant RadialGaugeType = 4
)

// ── RadialGaugePosition ───────────────────────────────────────────────────────

// RadialGaugePosition is a flags enum for the orientation of non-full gauges.
// Matches C# enum RadialGaugePosition (Radial/RadialGauge.cs).
type RadialGaugePosition int

const (
	// RadialGaugePositionNone — not set (used for Circle type).
	RadialGaugePositionNone RadialGaugePosition = 0
	// RadialGaugePositionTop — flat edge faces up (semicircle) or top-left/top-right (quadrant).
	RadialGaugePositionTop RadialGaugePosition = 1
	// RadialGaugePositionBottom — flat edge faces down.
	RadialGaugePositionBottom RadialGaugePosition = 2
	// RadialGaugePositionLeft — flat edge faces left.
	RadialGaugePositionLeft RadialGaugePosition = 4
	// RadialGaugePositionRight — flat edge faces right.
	RadialGaugePositionRight RadialGaugePosition = 8
)

// IsTop reports whether the position includes the Top flag.
// Mirrors C# RadialUtils.IsTop.
func (p RadialGaugePosition) IsTop() bool { return p&RadialGaugePositionTop != 0 }

// IsBottom reports whether the position includes the Bottom flag.
// Mirrors C# RadialUtils.IsBottom.
func (p RadialGaugePosition) IsBottom() bool { return p&RadialGaugePositionBottom != 0 }

// IsLeft reports whether the position includes the Left flag.
// Mirrors C# RadialUtils.IsLeft.
func (p RadialGaugePosition) IsLeft() bool { return p&RadialGaugePositionLeft != 0 }

// IsRight reports whether the position includes the Right flag.
// Mirrors C# RadialUtils.IsRight.
func (p RadialGaugePosition) IsRight() bool { return p&RadialGaugePositionRight != 0 }

// IsSemicircle reports whether t == Semicircle.
// Mirrors C# RadialUtils.IsSemicircle.
func (t RadialGaugeType) IsSemicircle() bool { return t&RadialGaugeTypeSemicircle != 0 }

// IsQuadrant reports whether t == Quadrant.
// Mirrors C# RadialUtils.IsQuadrant.
func (t RadialGaugeType) IsQuadrant() bool { return t&RadialGaugeTypeQuadrant != 0 }

// radialStartAngleFor returns the default needle start angle in degrees for the
// given type/position combination, following the same logic as RadialPointer.DrawHorz.
// C# source: original-dotnet/FastReport.Base/Gauge/Radial/RadialPointer.cs lines 61-91.
func radialStartAngleFor(typ RadialGaugeType, pos RadialGaugePosition) float64 {
	if typ.IsSemicircle() {
		switch {
		case pos == RadialGaugePositionBottom || pos == RadialGaugePositionTop:
			return -90
		case pos == RadialGaugePositionLeft:
			return -180
		case pos == RadialGaugePositionRight:
			return -180
		}
	} else if typ.IsQuadrant() {
		switch {
		case pos.IsLeft() && pos.IsTop():
			return -90
		case pos.IsLeft() && pos.IsBottom():
			return -180
		case pos.IsRight() && pos.IsTop():
			return 90
		case pos.IsRight() && pos.IsBottom():
			return 180
		}
	}
	// Circle default.
	return -135
}
