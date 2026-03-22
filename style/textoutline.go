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
	// DrawBehind controls whether the outline is drawn behind the text (true)
	// or on top of it (false, default).
	// Mirrors C# TextOutline.DrawBehind (TextOutline.cs line 44-47).
	DrawBehind bool
}

// DefaultTextOutline returns a TextOutline with default values (disabled).
func DefaultTextOutline() TextOutline {
	return TextOutline{
		Color: color.RGBA{A: 255},
		Width: 1,
	}
}

// Assign copies all fields from src into this TextOutline.
// Mirrors C# TextOutline.Assign (TextOutline.cs line 123-129).
func (o *TextOutline) Assign(src TextOutline) {
	o.Enabled = src.Enabled
	o.Color = src.Color
	o.Width = src.Width
	o.DashStyle = src.DashStyle
	o.DrawBehind = src.DrawBehind
}

// Clone returns a copy of this TextOutline.
// Mirrors C# TextOutline.Clone (TextOutline.cs line 136-138).
func (o TextOutline) Clone() TextOutline {
	return o
}
