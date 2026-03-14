package style

import "image/color"

// LineStyle specifies the style of a border line.
// It is the Go equivalent of FastReport.LineStyle.
type LineStyle int

const (
	// LineStyleSolid is a solid line.
	LineStyleSolid LineStyle = iota
	// LineStyleDash is a dashed line.
	LineStyleDash
	// LineStyleDot is a dotted line.
	LineStyleDot
	// LineStyleDashDot is a dash-dot repeating pattern.
	LineStyleDashDot
	// LineStyleDashDotDot is a dash-dot-dot repeating pattern.
	LineStyleDashDotDot
	// LineStyleDouble is a double line.
	LineStyleDouble
)

// BorderSide identifies one of the four sides of a Border.
type BorderSide int

const (
	// BorderLeft is the left side (index 0).
	BorderLeft BorderSide = iota
	// BorderTop is the top side (index 1).
	BorderTop
	// BorderRight is the right side (index 2).
	BorderRight
	// BorderBottom is the bottom side (index 3).
	BorderBottom
)

// BorderLines is a flag set that controls which sides of a Border are visible.
// It mirrors the FastReport.BorderLines flags enum.
type BorderLines int

const (
	// BorderLinesNone hides all sides.
	BorderLinesNone BorderLines = 0
	// BorderLinesLeft shows the left side.
	BorderLinesLeft BorderLines = 1
	// BorderLinesRight shows the right side.
	BorderLinesRight BorderLines = 2
	// BorderLinesTop shows the top side.
	BorderLinesTop BorderLines = 4
	// BorderLinesBottom shows the bottom side.
	BorderLinesBottom BorderLines = 8
	// BorderLinesAll shows all four sides.
	BorderLinesAll BorderLines = 15
)

// BorderLine represents a single side of a border.
// It is the Go equivalent of FastReport.BorderLine.
type BorderLine struct {
	// Color is the line colour. Defaults to opaque black.
	Color color.RGBA
	// Style is the dash/dot pattern. Defaults to LineStyleSolid.
	Style LineStyle
	// Width is the line width in pixels. Defaults to 1.
	Width float32
}

// NewBorderLine returns a BorderLine with the FastReport defaults
// (black, solid, 1 px wide).
func NewBorderLine() *BorderLine {
	return &BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Style: LineStyleSolid,
		Width: 1,
	}
}

// Clone returns a deep copy of the BorderLine.
func (l *BorderLine) Clone() *BorderLine {
	c := *l
	return &c
}

// Equals reports whether l and other have the same Color, Style and Width.
func (l *BorderLine) Equals(other *BorderLine) bool {
	if other == nil {
		return false
	}
	return l.Color == other.Color && l.Style == other.Style && l.Width == other.Width
}

// Border holds the four sides of a report object's border plus an optional
// drop shadow.  It is the Go equivalent of FastReport.Border.
type Border struct {
	// Lines holds the four BorderLine values indexed by BorderSide.
	// Use Left(), Top(), Right(), Bottom() for named access.
	Lines [4]*BorderLine

	// VisibleLines controls which sides are actually drawn.
	VisibleLines BorderLines

	// Shadow enables a drop shadow on the bottom-right of the object.
	Shadow bool
	// ShadowColor is the colour of the drop shadow. Defaults to opaque black.
	ShadowColor color.RGBA
	// ShadowWidth is the width of the shadow in pixels. Defaults to 4.
	ShadowWidth float32
}

// NewBorder returns a Border initialised with FastReport defaults:
// four black, solid, 1 px wide border lines; no shadow (shadow width 4,
// shadow colour black, matching the .NET constructor).
func NewBorder() *Border {
	return &Border{
		Lines: [4]*BorderLine{
			NewBorderLine(),
			NewBorderLine(),
			NewBorderLine(),
			NewBorderLine(),
		},
		VisibleLines: BorderLinesNone,
		Shadow:       false,
		ShadowColor:  color.RGBA{R: 0, G: 0, B: 0, A: 255},
		ShadowWidth:  4,
	}
}

// Left returns the left BorderLine (index 0).
func (b *Border) Left() *BorderLine { return b.Lines[BorderLeft] }

// Top returns the top BorderLine (index 1).
func (b *Border) Top() *BorderLine { return b.Lines[BorderTop] }

// Right returns the right BorderLine (index 2).
func (b *Border) Right() *BorderLine { return b.Lines[BorderRight] }

// Bottom returns the bottom BorderLine (index 3).
func (b *Border) Bottom() *BorderLine { return b.Lines[BorderBottom] }

// Color returns the colour of the left line (representative of all lines when
// they share the same settings, matching the .NET Border.Color getter).
func (b *Border) Color() color.RGBA { return b.Lines[BorderLeft].Color }

// SetColor sets the same colour on all four border lines.
func (b *Border) SetColor(c color.RGBA) {
	for i := range b.Lines {
		b.Lines[i].Color = c
	}
}

// LineStyle returns the style of the left line.
func (b *Border) LineStyle() LineStyle { return b.Lines[BorderLeft].Style }

// SetLineStyle sets the same line style on all four border lines.
func (b *Border) SetLineStyle(s LineStyle) {
	for i := range b.Lines {
		b.Lines[i].Style = s
	}
}

// Width returns the width of the left line in pixels.
func (b *Border) Width() float32 { return b.Lines[BorderLeft].Width }

// SetWidth sets the same width on all four border lines.
func (b *Border) SetWidth(w float32) {
	for i := range b.Lines {
		b.Lines[i].Width = w
	}
}

// Clone returns a deep copy of the Border.
func (b *Border) Clone() *Border {
	nb := &Border{
		VisibleLines: b.VisibleLines,
		Shadow:       b.Shadow,
		ShadowColor:  b.ShadowColor,
		ShadowWidth:  b.ShadowWidth,
	}
	for i, l := range b.Lines {
		nb.Lines[i] = l.Clone()
	}
	return nb
}

// Equals reports whether b and other are identical in all fields.
func (b *Border) Equals(other *Border) bool {
	if other == nil {
		return false
	}
	if b.VisibleLines != other.VisibleLines ||
		b.Shadow != other.Shadow ||
		b.ShadowColor != other.ShadowColor ||
		b.ShadowWidth != other.ShadowWidth {
		return false
	}
	for i := range b.Lines {
		if !b.Lines[i].Equals(other.Lines[i]) {
			return false
		}
	}
	return true
}
