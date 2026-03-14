package style

import (
	"image/color"
	"testing"
)

// ---------------------------------------------------------------------------
// BorderLine tests
// ---------------------------------------------------------------------------

func TestNewBorderLine_Defaults(t *testing.T) {
	l := NewBorderLine()
	if l.Color != (color.RGBA{R: 0, G: 0, B: 0, A: 255}) {
		t.Errorf("expected black, got %v", l.Color)
	}
	if l.Style != LineStyleSolid {
		t.Errorf("expected LineStyleSolid, got %v", l.Style)
	}
	if l.Width != 1 {
		t.Errorf("expected width 1, got %v", l.Width)
	}
}

func TestBorderLine_Clone(t *testing.T) {
	orig := &BorderLine{
		Color: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Style: LineStyleDash,
		Width: 2,
	}
	clone := orig.Clone()

	if clone == orig {
		t.Error("Clone returned the same pointer")
	}
	if *clone != *orig {
		t.Errorf("Clone content differs: %+v vs %+v", *clone, *orig)
	}

	// Mutation of clone must not affect original.
	clone.Width = 99
	if orig.Width == 99 {
		t.Error("mutating clone affected original")
	}
}

func TestBorderLine_Equals(t *testing.T) {
	a := NewBorderLine()
	b := NewBorderLine()

	if !a.Equals(b) {
		t.Error("two default BorderLines should be equal")
	}

	b.Width = 3
	if a.Equals(b) {
		t.Error("BorderLines with different widths should not be equal")
	}

	b.Width = a.Width
	b.Style = LineStyleDot
	if a.Equals(b) {
		t.Error("BorderLines with different styles should not be equal")
	}

	b.Style = a.Style
	b.Color = color.RGBA{R: 255, A: 255}
	if a.Equals(b) {
		t.Error("BorderLines with different colours should not be equal")
	}
}

func TestBorderLine_Equals_NilOther(t *testing.T) {
	l := NewBorderLine()
	if l.Equals(nil) {
		t.Error("Equals(nil) should return false")
	}
}

// ---------------------------------------------------------------------------
// LineStyle constant tests
// ---------------------------------------------------------------------------

func TestLineStyle_Values(t *testing.T) {
	if LineStyleSolid != 0 {
		t.Errorf("LineStyleSolid want 0, got %d", LineStyleSolid)
	}
	if LineStyleDash != 1 {
		t.Errorf("LineStyleDash want 1, got %d", LineStyleDash)
	}
	if LineStyleDot != 2 {
		t.Errorf("LineStyleDot want 2, got %d", LineStyleDot)
	}
	if LineStyleDashDot != 3 {
		t.Errorf("LineStyleDashDot want 3, got %d", LineStyleDashDot)
	}
	if LineStyleDashDotDot != 4 {
		t.Errorf("LineStyleDashDotDot want 4, got %d", LineStyleDashDotDot)
	}
	if LineStyleDouble != 5 {
		t.Errorf("LineStyleDouble want 5, got %d", LineStyleDouble)
	}
}

// ---------------------------------------------------------------------------
// BorderSide constant tests
// ---------------------------------------------------------------------------

func TestBorderSide_Values(t *testing.T) {
	if BorderLeft != 0 {
		t.Errorf("BorderLeft want 0, got %d", BorderLeft)
	}
	if BorderTop != 1 {
		t.Errorf("BorderTop want 1, got %d", BorderTop)
	}
	if BorderRight != 2 {
		t.Errorf("BorderRight want 2, got %d", BorderRight)
	}
	if BorderBottom != 3 {
		t.Errorf("BorderBottom want 3, got %d", BorderBottom)
	}
}

// ---------------------------------------------------------------------------
// BorderLines flag tests
// ---------------------------------------------------------------------------

func TestBorderLines_Values(t *testing.T) {
	if BorderLinesNone != 0 {
		t.Errorf("BorderLinesNone want 0, got %d", BorderLinesNone)
	}
	if BorderLinesLeft != 1 {
		t.Errorf("BorderLinesLeft want 1, got %d", BorderLinesLeft)
	}
	if BorderLinesRight != 2 {
		t.Errorf("BorderLinesRight want 2, got %d", BorderLinesRight)
	}
	if BorderLinesTop != 4 {
		t.Errorf("BorderLinesTop want 4, got %d", BorderLinesTop)
	}
	if BorderLinesBottom != 8 {
		t.Errorf("BorderLinesBottom want 8, got %d", BorderLinesBottom)
	}
	if BorderLinesAll != 15 {
		t.Errorf("BorderLinesAll want 15, got %d", BorderLinesAll)
	}
}

func TestBorderLines_FlagComposition(t *testing.T) {
	combo := BorderLinesLeft | BorderLinesTop
	if combo&BorderLinesLeft == 0 {
		t.Error("expected Left flag to be set")
	}
	if combo&BorderLinesTop == 0 {
		t.Error("expected Top flag to be set")
	}
	if combo&BorderLinesRight != 0 {
		t.Error("Right flag should not be set")
	}
}

// ---------------------------------------------------------------------------
// Border tests
// ---------------------------------------------------------------------------

func TestNewBorder_Defaults(t *testing.T) {
	b := NewBorder()

	if b.VisibleLines != BorderLinesNone {
		t.Errorf("expected VisibleLines none, got %v", b.VisibleLines)
	}
	if b.Shadow {
		t.Error("expected Shadow false")
	}
	if b.ShadowWidth != 4 {
		t.Errorf("expected ShadowWidth 4, got %v", b.ShadowWidth)
	}
	if b.ShadowColor != (color.RGBA{R: 0, G: 0, B: 0, A: 255}) {
		t.Errorf("expected black shadow colour, got %v", b.ShadowColor)
	}
	for i, l := range b.Lines {
		if l == nil {
			t.Errorf("Lines[%d] is nil", i)
		}
	}
}

func TestBorder_SideAccessors(t *testing.T) {
	b := NewBorder()

	if b.Left() != b.Lines[0] {
		t.Error("Left() should return Lines[0]")
	}
	if b.Top() != b.Lines[1] {
		t.Error("Top() should return Lines[1]")
	}
	if b.Right() != b.Lines[2] {
		t.Error("Right() should return Lines[2]")
	}
	if b.Bottom() != b.Lines[3] {
		t.Error("Bottom() should return Lines[3]")
	}
}

func TestBorder_Color_SetColor(t *testing.T) {
	b := NewBorder()
	red := color.RGBA{R: 255, A: 255}

	b.SetColor(red)

	if b.Color() != red {
		t.Errorf("Color() want %v, got %v", red, b.Color())
	}
	for i, l := range b.Lines {
		if l.Color != red {
			t.Errorf("Lines[%d].Color not updated, got %v", i, l.Color)
		}
	}
}

func TestBorder_LineStyle_SetLineStyle(t *testing.T) {
	b := NewBorder()
	b.SetLineStyle(LineStyleDot)

	if b.LineStyle() != LineStyleDot {
		t.Errorf("LineStyle() want Dot, got %v", b.LineStyle())
	}
	for i, l := range b.Lines {
		if l.Style != LineStyleDot {
			t.Errorf("Lines[%d].Style not updated", i)
		}
	}
}

func TestBorder_Width_SetWidth(t *testing.T) {
	b := NewBorder()
	b.SetWidth(3.5)

	if b.Width() != 3.5 {
		t.Errorf("Width() want 3.5, got %v", b.Width())
	}
	for i, l := range b.Lines {
		if l.Width != 3.5 {
			t.Errorf("Lines[%d].Width not updated", i)
		}
	}
}

func TestBorder_Clone(t *testing.T) {
	orig := NewBorder()
	orig.Shadow = true
	orig.ShadowWidth = 8
	orig.ShadowColor = color.RGBA{R: 128, G: 128, B: 128, A: 255}
	orig.VisibleLines = BorderLinesAll
	orig.SetColor(color.RGBA{R: 0, G: 0, B: 255, A: 255})

	clone := orig.Clone()

	if clone == orig {
		t.Error("Clone returned the same pointer")
	}
	if !orig.Equals(clone) {
		t.Error("Clone is not equal to original")
	}

	// Mutating the clone must not affect the original.
	clone.Shadow = false
	if !orig.Shadow {
		t.Error("mutating clone.Shadow affected original")
	}

	clone.Lines[0].Width = 99
	if orig.Lines[0].Width == 99 {
		t.Error("mutating clone.Lines[0] affected original")
	}
}

func TestBorder_Equals(t *testing.T) {
	a := NewBorder()
	b := NewBorder()

	if !a.Equals(b) {
		t.Error("two default Borders should be equal")
	}

	b.Shadow = true
	if a.Equals(b) {
		t.Error("Borders with different Shadow should not be equal")
	}
	b.Shadow = false

	b.ShadowWidth = 10
	if a.Equals(b) {
		t.Error("Borders with different ShadowWidth should not be equal")
	}
	b.ShadowWidth = a.ShadowWidth

	b.ShadowColor = color.RGBA{R: 200, A: 255}
	if a.Equals(b) {
		t.Error("Borders with different ShadowColor should not be equal")
	}
	b.ShadowColor = a.ShadowColor

	b.VisibleLines = BorderLinesAll
	if a.Equals(b) {
		t.Error("Borders with different VisibleLines should not be equal")
	}
	b.VisibleLines = a.VisibleLines

	b.Lines[2].Width = 5
	if a.Equals(b) {
		t.Error("Borders with different line widths should not be equal")
	}
}

func TestBorder_Equals_NilOther(t *testing.T) {
	b := NewBorder()
	if b.Equals(nil) {
		t.Error("Equals(nil) should return false")
	}
}
