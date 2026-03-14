package object_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/object"
)

// -----------------------------------------------------------------------
// CapSettings
// -----------------------------------------------------------------------

func TestDefaultCapSettings(t *testing.T) {
	c := object.DefaultCapSettings()
	if c.Width != 8 {
		t.Errorf("Width default = %v, want 8", c.Width)
	}
	if c.Height != 8 {
		t.Errorf("Height default = %v, want 8", c.Height)
	}
	if c.Style != object.CapStyleNone {
		t.Errorf("Style default = %d, want None", c.Style)
	}
}

// -----------------------------------------------------------------------
// LineObject
// -----------------------------------------------------------------------

func TestNewLineObject_Defaults(t *testing.T) {
	l := object.NewLineObject()
	if l == nil {
		t.Fatal("NewLineObject returned nil")
	}
	if l.Diagonal() {
		t.Error("Diagonal should default to false")
	}
	if l.StartCap != object.DefaultCapSettings() {
		t.Errorf("StartCap default = %+v", l.StartCap)
	}
	if l.EndCap != object.DefaultCapSettings() {
		t.Errorf("EndCap default = %+v", l.EndCap)
	}
	if l.DashPattern() != nil {
		t.Error("DashPattern should default to nil")
	}
}

func TestLineObject_Diagonal(t *testing.T) {
	l := object.NewLineObject()
	l.SetDiagonal(true)
	if !l.Diagonal() {
		t.Error("Diagonal should be true")
	}
}

func TestLineObject_StartCap(t *testing.T) {
	l := object.NewLineObject()
	l.StartCap = object.CapSettings{Width: 12, Height: 10, Style: object.CapStyleArrow}
	if l.StartCap.Style != object.CapStyleArrow {
		t.Errorf("StartCap.Style = %d, want Arrow", l.StartCap.Style)
	}
}

func TestLineObject_EndCap(t *testing.T) {
	l := object.NewLineObject()
	l.EndCap = object.CapSettings{Width: 6, Height: 6, Style: object.CapStyleCircle}
	if l.EndCap.Style != object.CapStyleCircle {
		t.Error("EndCap.Style should be Circle")
	}
}

func TestLineObject_DashPattern(t *testing.T) {
	l := object.NewLineObject()
	l.SetDashPattern([]float32{5, 3, 2, 3})
	if len(l.DashPattern()) != 4 {
		t.Errorf("DashPattern len = %d, want 4", len(l.DashPattern()))
	}
}

func TestLineObject_InheritsVisible(t *testing.T) {
	l := object.NewLineObject()
	if !l.Visible() {
		t.Error("LineObject should inherit Visible=true from ReportComponentBase")
	}
}

// -----------------------------------------------------------------------
// ShapeObject
// -----------------------------------------------------------------------

func TestNewShapeObject_Defaults(t *testing.T) {
	s := object.NewShapeObject()
	if s == nil {
		t.Fatal("NewShapeObject returned nil")
	}
	if s.Shape() != object.ShapeKindRectangle {
		t.Errorf("Shape default = %d, want Rectangle", s.Shape())
	}
	if s.Curve() != 0 {
		t.Errorf("Curve default = %v, want 0", s.Curve())
	}
}

func TestShapeObject_Shape(t *testing.T) {
	s := object.NewShapeObject()
	s.SetShape(object.ShapeKindEllipse)
	if s.Shape() != object.ShapeKindEllipse {
		t.Error("Shape should be Ellipse")
	}
}

func TestShapeObject_Curve(t *testing.T) {
	s := object.NewShapeObject()
	s.SetCurve(10)
	if s.Curve() != 10 {
		t.Errorf("Curve = %v, want 10", s.Curve())
	}
}

func TestShapeObject_DashPattern(t *testing.T) {
	s := object.NewShapeObject()
	s.SetDashPattern([]float32{4, 2})
	if len(s.DashPattern()) != 2 {
		t.Errorf("DashPattern len = %d, want 2", len(s.DashPattern()))
	}
}

func TestShapeObject_ShapeKinds(t *testing.T) {
	kinds := []struct {
		k    object.ShapeKind
		name string
	}{
		{object.ShapeKindRectangle, "Rectangle"},
		{object.ShapeKindRoundRectangle, "RoundRectangle"},
		{object.ShapeKindEllipse, "Ellipse"},
		{object.ShapeKindTriangle, "Triangle"},
		{object.ShapeKindDiamond, "Diamond"},
	}
	for _, tc := range kinds {
		s := object.NewShapeObject()
		s.SetShape(tc.k)
		if s.Shape() != tc.k {
			t.Errorf("Shape %s: got %d, want %d", tc.name, s.Shape(), tc.k)
		}
	}
}

// -----------------------------------------------------------------------
// PolyLineObject / PolygonObject
// -----------------------------------------------------------------------

func TestNewPolyLineObject_Defaults(t *testing.T) {
	p := object.NewPolyLineObject()
	if p == nil {
		t.Fatal("NewPolyLineObject returned nil")
	}
	if p.Points() == nil {
		t.Error("Points should not be nil")
	}
	if p.Points().Len() != 0 {
		t.Errorf("Points.Len = %d, want 0", p.Points().Len())
	}
}

func TestPolyLineObject_AddPoint(t *testing.T) {
	p := object.NewPolyLineObject()
	p.Points().Add(&object.PolyPoint{X: 10, Y: 20})
	p.Points().Add(&object.PolyPoint{X: 30, Y: 40})
	if p.Points().Len() != 2 {
		t.Errorf("Points.Len = %d, want 2", p.Points().Len())
	}
	if p.Points().Get(0).X != 10 {
		t.Errorf("Point[0].X = %v, want 10", p.Points().Get(0).X)
	}
}

func TestPolyLineObject_Center(t *testing.T) {
	p := object.NewPolyLineObject()
	p.SetCenterX(50)
	p.SetCenterY(75)
	if p.CenterX() != 50 || p.CenterY() != 75 {
		t.Errorf("Center = (%v,%v), want (50,75)", p.CenterX(), p.CenterY())
	}
}

func TestPolyLineObject_DashPattern(t *testing.T) {
	p := object.NewPolyLineObject()
	p.SetDashPattern([]float32{6, 2})
	if len(p.DashPattern()) != 2 {
		t.Errorf("DashPattern len = %d, want 2", len(p.DashPattern()))
	}
}

func TestPolyPointCollection_Clear(t *testing.T) {
	p := object.NewPolyLineObject()
	p.Points().Add(&object.PolyPoint{X: 1, Y: 2})
	p.Points().Clear()
	if p.Points().Len() != 0 {
		t.Errorf("after Clear, Points.Len = %d, want 0", p.Points().Len())
	}
}

func TestNewPolygonObject(t *testing.T) {
	pg := object.NewPolygonObject()
	if pg == nil {
		t.Fatal("NewPolygonObject returned nil")
	}
	pg.Points().Add(&object.PolyPoint{X: 0, Y: 0})
	pg.Points().Add(&object.PolyPoint{X: 50, Y: 0})
	pg.Points().Add(&object.PolyPoint{X: 25, Y: 50})
	if pg.Points().Len() != 3 {
		t.Errorf("PolygonObject Points.Len = %d, want 3", pg.Points().Len())
	}
}
