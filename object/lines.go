package object

import (
	"github.com/andrewloable/go-fastreport/report"
)

// -----------------------------------------------------------------------
// CapStyle and CapSettings
// -----------------------------------------------------------------------

// CapStyle specifies the style of a line end cap.
type CapStyle int

const (
	// CapStyleNone draws no cap.
	CapStyleNone CapStyle = iota
	// CapStyleCircle draws a circle cap.
	CapStyleCircle
	// CapStyleSquare draws a square cap.
	CapStyleSquare
	// CapStyleDiamond draws a diamond cap.
	CapStyleDiamond
	// CapStyleArrow draws an arrow cap.
	CapStyleArrow
)

// CapSettings defines the visual cap at one end of a line.
// It is the Go equivalent of FastReport.CapSettings.
type CapSettings struct {
	// Width of the cap in pixels (default 8).
	Width float32
	// Height of the cap in pixels (default 8).
	Height float32
	// Style of the cap (default None).
	Style CapStyle
}

// DefaultCapSettings returns a CapSettings with default values.
func DefaultCapSettings() CapSettings {
	return CapSettings{Width: 8, Height: 8, Style: CapStyleNone}
}

// -----------------------------------------------------------------------
// LineObject
// -----------------------------------------------------------------------

// LineObject represents a line that can be diagonal or axis-aligned.
// It is the Go equivalent of FastReport.LineObject.
type LineObject struct {
	report.ReportComponentBase

	// diagonal indicates a diagonal (instead of horizontal/vertical) line.
	diagonal bool
	// StartCap is the cap drawn at the start of the line.
	StartCap CapSettings
	// EndCap is the cap drawn at the end of the line.
	EndCap CapSettings
	// dashPattern holds custom dash lengths in pixels (nil = solid line).
	dashPattern []float32
}

// NewLineObject creates a LineObject with defaults.
func NewLineObject() *LineObject {
	return &LineObject{
		ReportComponentBase: *report.NewReportComponentBase(),
		StartCap:            DefaultCapSettings(),
		EndCap:              DefaultCapSettings(),
	}
}

// Diagonal returns whether the line is drawn diagonally.
func (l *LineObject) Diagonal() bool { return l.diagonal }

// SetDiagonal sets the diagonal flag.
func (l *LineObject) SetDiagonal(v bool) { l.diagonal = v }

// DashPattern returns the custom dash pattern (nil = solid).
func (l *LineObject) DashPattern() []float32 { return l.dashPattern }

// SetDashPattern sets the dash pattern.
func (l *LineObject) SetDashPattern(dp []float32) { l.dashPattern = dp }

// Serialize writes LineObject properties that differ from defaults.
func (l *LineObject) Serialize(w report.Writer) error {
	if err := l.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if l.diagonal {
		w.WriteBool("Diagonal", true)
	}
	def := DefaultCapSettings()
	if l.StartCap != def {
		w.WriteStr("StartCap", capToStr(l.StartCap))
	}
	if l.EndCap != def {
		w.WriteStr("EndCap", capToStr(l.EndCap))
	}
	return nil
}

// Deserialize reads LineObject properties.
func (l *LineObject) Deserialize(r report.Reader) error {
	if err := l.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	l.diagonal = r.ReadBool("Diagonal", false)
	if s := r.ReadStr("StartCap", ""); s != "" {
		l.StartCap = capFromStr(s)
	}
	if s := r.ReadStr("EndCap", ""); s != "" {
		l.EndCap = capFromStr(s)
	}
	return nil
}

// capToStr serialises a CapSettings as "W,H,Style".
func capToStr(c CapSettings) string {
	return report.FormatFloat(c.Width) + "," + report.FormatFloat(c.Height) + "," +
		report.FormatFloat(float32(c.Style))
}

// capFromStr parses "W,H,Style".
func capFromStr(s string) CapSettings {
	parts := report.SplitComma(s)
	c := DefaultCapSettings()
	if len(parts) >= 1 {
		c.Width = report.ParseFloat(parts[0])
	}
	if len(parts) >= 2 {
		c.Height = report.ParseFloat(parts[1])
	}
	if len(parts) >= 3 {
		c.Style = CapStyle(int(report.ParseFloat(parts[2])))
	}
	return c
}

// -----------------------------------------------------------------------
// ShapeKind
// -----------------------------------------------------------------------

// ShapeKind identifies the geometric shape drawn by a ShapeObject.
type ShapeKind int

const (
	// ShapeKindRectangle draws a rectangle.
	ShapeKindRectangle ShapeKind = iota
	// ShapeKindRoundRectangle draws a round-cornered rectangle.
	ShapeKindRoundRectangle
	// ShapeKindEllipse draws an ellipse.
	ShapeKindEllipse
	// ShapeKindTriangle draws a triangle.
	ShapeKindTriangle
	// ShapeKindDiamond draws a diamond.
	ShapeKindDiamond
)

// -----------------------------------------------------------------------
// ShapeObject
// -----------------------------------------------------------------------

// ShapeObject draws a geometric shape (rectangle, ellipse, etc.).
// It is the Go equivalent of FastReport.ShapeObject.
type ShapeObject struct {
	report.ReportComponentBase

	// shape is the geometric shape to draw.
	shape ShapeKind
	// curve is the corner radius for RoundRectangle (pixels).
	curve float32
	// dashPattern holds custom dash lengths (nil = solid border).
	dashPattern []float32
}

// NewShapeObject creates a ShapeObject with defaults.
func NewShapeObject() *ShapeObject {
	return &ShapeObject{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// Shape returns the geometric shape kind.
func (s *ShapeObject) Shape() ShapeKind { return s.shape }

// SetShape sets the geometric shape.
func (s *ShapeObject) SetShape(k ShapeKind) { s.shape = k }

// Curve returns the corner radius for RoundRectangle shapes.
func (s *ShapeObject) Curve() float32 { return s.curve }

// SetCurve sets the corner radius.
func (s *ShapeObject) SetCurve(v float32) { s.curve = v }

// DashPattern returns the dash pattern (nil = solid).
func (s *ShapeObject) DashPattern() []float32 { return s.dashPattern }

// SetDashPattern sets the dash pattern.
func (s *ShapeObject) SetDashPattern(dp []float32) { s.dashPattern = dp }

// Serialize writes ShapeObject properties that differ from defaults.
func (s *ShapeObject) Serialize(w report.Writer) error {
	if err := s.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if s.shape != ShapeKindRectangle {
		w.WriteInt("Shape", int(s.shape))
	}
	if s.curve != 0 {
		w.WriteFloat("Curve", s.curve)
	}
	return nil
}

// Deserialize reads ShapeObject properties.
func (s *ShapeObject) Deserialize(r report.Reader) error {
	if err := s.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	s.shape = ShapeKind(r.ReadInt("Shape", 0))
	s.curve = r.ReadFloat("Curve", 0)
	return nil
}

// -----------------------------------------------------------------------
// PolyPoint and PolyPointCollection
// -----------------------------------------------------------------------

// PolyPoint is a point on a poly-line or polygon with optional bezier control points.
type PolyPoint struct {
	// X, Y are the point coordinates in pixels.
	X, Y float32
	// Left and Right are optional bezier control points (nil = straight segment).
	Left, Right *PolyPoint
}

// PolyPointCollection holds an ordered list of PolyPoints.
type PolyPointCollection struct {
	points []*PolyPoint
}

// Add appends a point.
func (c *PolyPointCollection) Add(p *PolyPoint) { c.points = append(c.points, p) }

// Len returns the number of points.
func (c *PolyPointCollection) Len() int { return len(c.points) }

// Get returns the point at index i.
func (c *PolyPointCollection) Get(i int) *PolyPoint { return c.points[i] }

// Clear removes all points.
func (c *PolyPointCollection) Clear() { c.points = nil }

// -----------------------------------------------------------------------
// PolyLineObject
// -----------------------------------------------------------------------

// PolyLineObject draws a polyline (open polygon) that may use bezier curves.
// It is the Go equivalent of FastReport.PolyLineObject.
type PolyLineObject struct {
	report.ReportComponentBase

	// center is the local origin used for point coordinates.
	centerX, centerY float32
	// points is the collection of vertices.
	points      *PolyPointCollection
	// dashPattern holds custom dash lengths (nil = solid).
	dashPattern []float32
}

// NewPolyLineObject creates a PolyLineObject with defaults.
func NewPolyLineObject() *PolyLineObject {
	return &PolyLineObject{
		ReportComponentBase: *report.NewReportComponentBase(),
		points:              &PolyPointCollection{},
	}
}

// CenterX returns the local x origin.
func (p *PolyLineObject) CenterX() float32 { return p.centerX }

// SetCenterX sets the local x origin.
func (p *PolyLineObject) SetCenterX(v float32) { p.centerX = v }

// CenterY returns the local y origin.
func (p *PolyLineObject) CenterY() float32 { return p.centerY }

// SetCenterY sets the local y origin.
func (p *PolyLineObject) SetCenterY(v float32) { p.centerY = v }

// Points returns the vertex collection.
func (p *PolyLineObject) Points() *PolyPointCollection { return p.points }

// DashPattern returns the dash pattern (nil = solid).
func (p *PolyLineObject) DashPattern() []float32 { return p.dashPattern }

// SetDashPattern sets the dash pattern.
func (p *PolyLineObject) SetDashPattern(dp []float32) { p.dashPattern = dp }

// Serialize writes PolyLineObject properties.
func (p *PolyLineObject) Serialize(w report.Writer) error {
	return p.ReportComponentBase.Serialize(w)
}

// Deserialize reads PolyLineObject properties.
func (p *PolyLineObject) Deserialize(r report.Reader) error {
	return p.ReportComponentBase.Deserialize(r)
}

// -----------------------------------------------------------------------
// PolygonObject
// -----------------------------------------------------------------------

// PolygonObject is a closed polyline (polygon) with an optional fill.
// It is the Go equivalent of FastReport.PolygonObject.
type PolygonObject struct {
	PolyLineObject
}

// NewPolygonObject creates a PolygonObject with defaults.
func NewPolygonObject() *PolygonObject {
	return &PolygonObject{PolyLineObject: *NewPolyLineObject()}
}
