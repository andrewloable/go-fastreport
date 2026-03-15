package style

import "image/color"

// TextOutline defines a stroke drawn around each character in a text object.
// It is the Go equivalent of FastReport.TextOutline.
type TextOutline struct {
	// Enabled controls whether the outline is drawn.
	Enabled bool
	// Color is the stroke color.
	Color color.RGBA
	// Width is the stroke width in pixels.
	Width float32
	// DashStyle is the line style (0=Solid, 1=Dash, 2=Dot, 3=DashDot, 4=DashDotDot).
	DashStyle int
}

// DefaultTextOutline returns a TextOutline with default values (disabled).
func DefaultTextOutline() TextOutline {
	return TextOutline{
		Color: color.RGBA{A: 255},
		Width: 1,
	}
}
