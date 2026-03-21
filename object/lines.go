package object

import (
	"github.com/andrewloable/go-fastreport/report"
)

// -----------------------------------------------------------------------
// CapStyle and CapSettings
// -----------------------------------------------------------------------

// CapStyle specifies the style of a line end cap.
// It is the Go equivalent of FastReport.CapStyle.
// CapSettings.cs, FastReport.Base — enum order matches C# definition.
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

// formatCapStyle converts a CapStyle to its FRX string name (e.g. "Arrow").
// Matches the C# enum name used by FRWriter.WriteValue.
func formatCapStyle(s CapStyle) string {
	switch s {
	case CapStyleCircle:
		return "Circle"
	case CapStyleSquare:
		return "Square"
	case CapStyleDiamond:
		return "Diamond"
	case CapStyleArrow:
		return "Arrow"
	default:
		return "None"
	}
}

// parseCapStyle converts an FRX string name to a CapStyle.
func parseCapStyle(s string) CapStyle {
	switch s {
	case "Circle":
		return CapStyleCircle
	case "Square":
		return CapStyleSquare
	case "Diamond":
		return CapStyleDiamond
	case "Arrow":
		return CapStyleArrow
	default:
		return CapStyleNone
	}
}

// CapSettings defines the visual cap at one end of a line.
// It is the Go equivalent of FastReport.CapSettings (CapSettings.cs).
type CapSettings struct {
	// Width of the cap in pixels (default 8).
	Width float32
	// Height of the cap in pixels (default 8).
	Height float32
	// Style of the cap (default None).
	Style CapStyle
}

// DefaultCapSettings returns a CapSettings with default values matching
// the C# constructor: width=8, height=8, style=None.
func DefaultCapSettings() CapSettings {
	return CapSettings{Width: 8, Height: 8, Style: CapStyleNone}
}

// Assign copies all fields from src into c.
// Equivalent to C# CapSettings.Assign(CapSettings source).
func (c *CapSettings) Assign(src CapSettings) {
	c.Width = src.Width
	c.Height = src.Height
	c.Style = src.Style
}

// Clone returns an independent copy of c.
// Equivalent to C# CapSettings.Clone().
func (c CapSettings) Clone() CapSettings {
	var result CapSettings
	result.Assign(c)
	return result
}

// Equals reports whether c and other have identical field values.
// Equivalent to C# CapSettings.Equals(object obj).
func (c CapSettings) Equals(other CapSettings) bool {
	return c.Width == other.Width && c.Height == other.Height && c.Style == other.Style
}

// SerializeCap writes the three dot-qualified attributes for a cap property
// using the FRX format that C# CapSettings.Serialize(prefix, writer, diff) produces:
//
//	prefix.Width, prefix.Height, prefix.Style
//
// Only attributes that differ from def are written (diff-encoding, same as C#).
func SerializeCap(prefix string, w report.Writer, c, def CapSettings) {
	if c.Width != def.Width {
		w.WriteFloat(prefix+".Width", c.Width)
	}
	if c.Height != def.Height {
		w.WriteFloat(prefix+".Height", c.Height)
	}
	if c.Style != def.Style {
		w.WriteStr(prefix+".Style", formatCapStyle(c.Style))
	}
}

// DeserializeCap reads the dot-qualified cap attributes written by SerializeCap
// and returns the resulting CapSettings starting from def.
func DeserializeCap(prefix string, r report.Reader, def CapSettings) CapSettings {
	c := def
	c.Width = r.ReadFloat(prefix+".Width", def.Width)
	c.Height = r.ReadFloat(prefix+".Height", def.Height)
	if s := r.ReadStr(prefix+".Style", ""); s != "" {
		c.Style = parseCapStyle(s)
	}
	return c
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
// Follows LineObject.cs Serialize(): writes Diagonal, then delegates to
// CapSettings.Serialize("StartCap", …) and CapSettings.Serialize("EndCap", …)
// which produce dot-qualified attributes (e.g. StartCap.Style="Arrow").
func (l *LineObject) Serialize(w report.Writer) error {
	if err := l.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if l.diagonal {
		w.WriteBool("Diagonal", true)
	}
	def := DefaultCapSettings()
	SerializeCap("StartCap", w, l.StartCap, def)
	SerializeCap("EndCap", w, l.EndCap, def)
	return nil
}

// Deserialize reads LineObject properties.
// Cap attributes are read as dot-qualified names matching the FRX format
// produced by C# CapSettings.Serialize() (e.g. StartCap.Style="Arrow").
func (l *LineObject) Deserialize(r report.Reader) error {
	if err := l.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	l.diagonal = r.ReadBool("Diagonal", false)
	def := DefaultCapSettings()
	l.StartCap = DeserializeCap("StartCap", r, def)
	l.EndCap = DeserializeCap("EndCap", r, def)
	return nil
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

// formatShapeKind converts ShapeKind to its FRX string name.
func formatShapeKind(k ShapeKind) string {
	switch k {
	case ShapeKindRoundRectangle:
		return "RoundRectangle"
	case ShapeKindEllipse:
		return "Ellipse"
	case ShapeKindTriangle:
		return "Triangle"
	case ShapeKindDiamond:
		return "Diamond"
	default:
		return "Rectangle"
	}
}

// parseShapeKind converts an FRX string to ShapeKind (handles both names and ints).
func parseShapeKind(s string) ShapeKind {
	switch s {
	case "RoundRectangle", "1":
		return ShapeKindRoundRectangle
	case "Ellipse", "2":
		return ShapeKindEllipse
	case "Triangle", "3":
		return ShapeKindTriangle
	case "Diamond", "4":
		return ShapeKindDiamond
	default:
		return ShapeKindRectangle
	}
}

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
		w.WriteStr("Shape", formatShapeKind(s.shape))
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
	s.shape = parseShapeKind(r.ReadStr("Shape", "Rectangle"))
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
