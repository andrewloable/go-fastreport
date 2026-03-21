package utils

import (
	"math"
	"sync"
)

// screenDPI is the fixed DPI assumed by the Go port.
// The report engine uses 96 DPI internally for all pixel measurements.
// Ported from C# DrawUtils.GetDpi() → Graphics.DpiX, which returns 96 on
// most screen configurations.
const screenDPI = 96

// uiScaleMu guards uiScaleVal.
var uiScaleMu sync.RWMutex

// uiScaleVal holds the current UI scale; valid range 1.0–1.5.
// Defaults to 1.0 to match C# DrawUtils._uiScale initialiser (DrawUtils.cs:53).
var uiScaleVal float32 = 1.0

// ScreenDpi returns the screen DPI used by the report engine.
// In the Go port this is always 96 (the report internal unit is 96-dpi pixels).
// Matches C# DrawUtils.ScreenDpi (DrawUtils.cs:24-31).
func ScreenDpi() int {
	return screenDPI
}

// ScreenDpiFX returns the DPI scaling factor relative to 96 dpi.
// Formula: 96 / ScreenDpi.  Because ScreenDpi is always 96 in the Go port,
// this returns 1.0.  The function is provided so that ported C# code that
// references DrawUtils.ScreenDpiFX compiles without change.
// Matches C# DrawUtils.ScreenDpiFX (DrawUtils.cs:34-42).
func ScreenDpiFX() float32 {
	return 96.0 / float32(screenDPI)
}

// UIScale returns the additional UI scale factor applied to report forms.
// Valid range is 1.0 to 1.5.  Defaults to 1.0.
// Matches C# DrawUtils.UIScale (DrawUtils.cs:53-69).
func UIScale() float32 {
	uiScaleMu.RLock()
	defer uiScaleMu.RUnlock()
	return uiScaleVal
}

// SetUIScale sets the additional UI scale factor.
// The value is clamped to [1.0, 1.5].
// Matches C# DrawUtils.UIScale setter (DrawUtils.cs:62-68).
func SetUIScale(v float32) {
	if v < 1.0 {
		v = 1.0
	}
	if v > 1.5 {
		v = 1.5
	}
	uiScaleMu.Lock()
	uiScaleVal = v
	uiScaleMu.Unlock()
}

// ── Unit conversion helpers ─────────────────────────────────────────────────
// These mirror the C# unit-conversion functions used across FastReport.
// The report engine stores all measurements as screen pixels at 96 DPI.

// drawMMPerInch is millimetres per inch, used for MM/pixel conversions.
const drawMMPerInch = 25.4

// PixelsToMM converts pixels (96 dpi) to millimetres.
// Uses the same 96 DPI base as standardDPI in text.go.
func PixelsToMM(px float32) float32 {
	return px / standardDPI * drawMMPerInch
}

// MMToPixels converts millimetres to pixels (96 dpi).
// Uses the same 96 DPI base as standardDPI in text.go.
func MMToPixels(mm float32) float32 {
	return mm / drawMMPerInch * standardDPI
}

// PixelsToInches converts pixels (96 dpi) to inches.
func PixelsToInches(px float32) float32 {
	return px / standardDPI
}

// InchesToPixels converts inches to pixels (96 dpi).
func InchesToPixels(in float32) float32 {
	return in * standardDPI
}

// ── Coordinate / bounds helpers ─────────────────────────────────────────────

// RectF represents a floating-point rectangle, matching C# System.Drawing.RectangleF.
type RectF struct {
	X, Y, Width, Height float32
}

// Right returns X + Width.
func (r RectF) Right() float32 { return r.X + r.Width }

// Bottom returns Y + Height.
func (r RectF) Bottom() float32 { return r.Y + r.Height }

// IsEmpty reports whether the rectangle has zero area.
func (r RectF) IsEmpty() bool { return r.Width <= 0 || r.Height <= 0 }

// SizeF represents a floating-point size, matching C# System.Drawing.SizeF.
type SizeF struct {
	Width, Height float32
}

// IsEmpty reports whether either dimension is zero.
func (s SizeF) IsEmpty() bool { return s.Width == 0 && s.Height == 0 }

// PointF represents a floating-point point, matching C# System.Drawing.PointF.
type PointF struct {
	X, Y float32
}

// ── Numeric helpers ─────────────────────────────────────────────────────────

// RoundFloat64 rounds v to the given number of decimal places.
// This is a convenience wrapper used by exporters and matches the
// rounding logic in C# ExportUtils.FloatToString.
func RoundFloat64(v float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	return math.Round(v*pow) / pow
}
