package object

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/utils"
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

// Assign copies all LineObject properties from src.
// Mirrors C# LineObject.Assign (LineObject.cs:81-89).
func (l *LineObject) Assign(src *LineObject) {
	l.ReportComponentBase.Assign(&src.ReportComponentBase)
	l.diagonal = src.diagonal
	l.StartCap.Assign(src.StartCap)
	l.EndCap.Assign(src.EndCap)
	if len(src.dashPattern) > 0 {
		l.dashPattern = make([]float32, len(src.dashPattern))
		copy(l.dashPattern, src.dashPattern)
	} else {
		l.dashPattern = nil
	}
}

// Serialize writes LineObject properties that differ from defaults.
// Follows LineObject.cs Serialize(): writes Diagonal, then delegates to
// CapSettings.Serialize("StartCap", …) and CapSettings.Serialize("EndCap", …)
// which produce dot-qualified attributes (e.g. StartCap.Style="Arrow").
// Also writes DashPattern when non-empty (LineObject.cs line 274-275).
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
	// DashPattern — only written when non-empty (LineObject.cs:274).
	if len(l.dashPattern) > 0 {
		w.WriteStr("DashPattern", utils.FloatCollection(l.dashPattern).String())
	}
	return nil
}

// Deserialize reads LineObject properties.
// Cap attributes are read as dot-qualified names matching the FRX format
// produced by C# CapSettings.Serialize() (e.g. StartCap.Style="Arrow").
// DashPattern is read when present (LineObject.cs line 274).
func (l *LineObject) Deserialize(r report.Reader) error {
	if err := l.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	l.diagonal = r.ReadBool("Diagonal", false)
	def := DefaultCapSettings()
	l.StartCap = DeserializeCap("StartCap", r, def)
	l.EndCap = DeserializeCap("EndCap", r, def)
	if s := r.ReadStr("DashPattern", ""); s != "" {
		fc, err := utils.ParseFloatCollection(s)
		if err == nil {
			l.dashPattern = []float32(fc)
		}
	}
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

// Assign copies all ShapeObject properties from src.
// Mirrors C# ShapeObject.Assign (ShapeObject.cs lines 115-123).
func (s *ShapeObject) Assign(src *ShapeObject) {
	if src == nil {
		return
	}
	s.ReportComponentBase = src.ReportComponentBase
	s.shape = src.shape
	s.curve = src.curve
	if src.dashPattern != nil {
		s.dashPattern = make([]float32, len(src.dashPattern))
		copy(s.dashPattern, src.dashPattern)
	} else {
		s.dashPattern = nil
	}
}

// Serialize writes ShapeObject properties that differ from defaults.
// Mirrors C# ShapeObject.Serialize (ShapeObject.cs lines 204-215).
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
	if len(s.dashPattern) > 0 {
		w.WriteStr("DashPattern", utils.FloatCollection(s.dashPattern).String())
	}
	return nil
}

// Deserialize reads ShapeObject properties.
// Mirrors C# ShapeObject.Deserialize (reads Shape, Curve, DashPattern).
func (s *ShapeObject) Deserialize(r report.Reader) error {
	if err := s.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	s.shape = parseShapeKind(r.ReadStr("Shape", "Rectangle"))
	s.curve = r.ReadFloat("Curve", 0)
	if dp := r.ReadStr("DashPattern", ""); dp != "" {
		if fc, err := utils.ParseFloatCollection(dp); err == nil {
			s.dashPattern = []float32(fc)
		}
	}
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

// clonePolyPoint returns a deep copy of pp including its Left/Right control points.
func clonePolyPoint(pp *PolyPoint) *PolyPoint {
	if pp == nil {
		return nil
	}
	cp := &PolyPoint{X: pp.X, Y: pp.Y}
	cp.Left = clonePolyPoint(pp.Left)
	cp.Right = clonePolyPoint(pp.Right)
	return cp
}

// Remove removes the point at the given index.
// No-op when index is out of range.
// Mirrors C# PolyPointCollection.Remove (PolyLineObject.cs).
func (c *PolyPointCollection) Remove(index int) {
	if index < 0 || index >= len(c.points) {
		return
	}
	c.points = append(c.points[:index], c.points[index+1:]...)
}

// Insert inserts point p at the given index, pushing existing points right.
// If index >= Len() the point is appended.
// Mirrors C# PolyPointCollection.Insert (PolyLineObject.cs).
func (c *PolyPointCollection) Insert(index int, p *PolyPoint) {
	if index >= len(c.points) {
		c.points = append(c.points, p)
		return
	}
	if index < 0 {
		index = 0
	}
	c.points = append(c.points, nil)
	copy(c.points[index+1:], c.points[index:])
	c.points[index] = p
}

// Clone returns a deep copy of the collection.
// Mirrors C# pointsCollection.Clone() (PolyLineObject.cs line 144).
func (c *PolyPointCollection) Clone() *PolyPointCollection {
	clone := &PolyPointCollection{
		points: make([]*PolyPoint, len(c.points)),
	}
	for i, p := range c.points {
		clone.points[i] = clonePolyPoint(p)
	}
	return clone
}

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

// Assign copies all PolyLineObject properties from src.
// Points are deep-cloned so both objects remain independent.
// Mirrors C# PolyLineObject.Assign (PolyLineObject.cs lines 138-148).
func (p *PolyLineObject) Assign(src *PolyLineObject) {
	if src == nil {
		return
	}
	p.ReportComponentBase = src.ReportComponentBase
	p.centerX = src.centerX
	p.centerY = src.centerY
	p.points = src.points.Clone()
	if src.dashPattern != nil {
		p.dashPattern = make([]float32, len(src.dashPattern))
		copy(p.dashPattern, src.dashPattern)
	} else {
		p.dashPattern = nil
	}
}

// serializePolyPoint converts a PolyPoint to the PolyPoints_v2 token format.
// Format: "X/Y" with optional "/L/LX/LY" and/or "/R/RX/RY" suffixes.
// Matches C# PolyPoint.Serialize(StringBuilder) in PolyLineObject.cs.
func serializePolyPoint(pp *PolyPoint) string {
	round := func(v float32) string {
		s := strconv.FormatFloat(float64(v), 'f', 4, 32)
		// Trim trailing zeros after decimal point to match C# Math.Round(v, 4).
		if strings.Contains(s, ".") {
			s = strings.TrimRight(s, "0")
			s = strings.TrimRight(s, ".")
		}
		return s
	}
	var sb strings.Builder
	sb.WriteString(round(pp.X))
	sb.WriteByte('/')
	sb.WriteString(round(pp.Y))
	if pp.Left != nil {
		sb.WriteString("/L/")
		sb.WriteString(round(pp.Left.X))
		sb.WriteByte('/')
		sb.WriteString(round(pp.Left.Y))
	}
	if pp.Right != nil {
		sb.WriteString("/R/")
		sb.WriteString(round(pp.Right.X))
		sb.WriteByte('/')
		sb.WriteString(round(pp.Right.Y))
	}
	return sb.String()
}

// deserializePolyPointV2 parses a single PolyPoints_v2 token ("X/Y[/L/lx/ly][/R/rx/ry]").
// Matches C# PolyPoint.Deserialize(string) in PolyLineObject.cs.
func deserializePolyPointV2(s string) (*PolyPoint, error) {
	parts := strings.Split(s, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid PolyPoint token: %q", s)
	}
	parseF := func(tok string) (float32, error) {
		v, err := strconv.ParseFloat(strings.TrimSpace(tok), 32)
		return float32(v), err
	}
	x, err := parseF(parts[0])
	if err != nil {
		return nil, err
	}
	y, err := parseF(parts[1])
	if err != nil {
		return nil, err
	}
	pp := &PolyPoint{X: x, Y: y}
	for i := 2; i < len(parts); {
		marker := parts[i]
		i++
		if (marker == "L" || marker == "R") && i+1 < len(parts) {
			cx, err := parseF(parts[i])
			if err != nil {
				break
			}
			cy, err := parseF(parts[i+1])
			if err != nil {
				break
			}
			cp := &PolyPoint{X: cx, Y: cy}
			if marker == "L" {
				pp.Left = cp
			} else {
				pp.Right = cp
			}
			i += 2
		}
	}
	return pp, nil
}

// Serialize writes PolyLineObject properties.
// Matches C# PolyLineObject.Serialize(): writes PolyPoints_v2, CenterX, CenterY,
// and DashPattern when non-empty (PolyLineObject.cs lines 496-517).
func (p *PolyLineObject) Serialize(w report.Writer) error {
	if err := p.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	// Build PolyPoints_v2 string: "X/Y[/L/…][/R/…]" per point, separated by "|".
	var parts []string
	for i := 0; i < p.points.Len(); i++ {
		parts = append(parts, serializePolyPoint(p.points.Get(i)))
	}
	w.WriteStr("PolyPoints_v2", strings.Join(parts, "|"))
	w.WriteFloat("CenterX", p.centerX)
	w.WriteFloat("CenterY", p.centerY)
	if len(p.dashPattern) > 0 {
		w.WriteStr("DashPattern", utils.FloatCollection(p.dashPattern).String())
	}
	return nil
}

// Deserialize reads PolyLineObject properties.
// Handles both legacy PolyPoints format ("X\Y\type" separated by "|") and
// the current PolyPoints_v2 format ("X/Y[/L/…][/R/…]" separated by "|").
// Also reads CenterX, CenterY, and DashPattern.
// Matches C# PolyLineObject.Deserialize() (PolyLineObject.cs lines 151-184).
func (p *PolyLineObject) Deserialize(r report.Reader) error {
	if err := p.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	p.points.Clear()
	if s := r.ReadStr("PolyPoints_v2", ""); s != "" {
		// PolyPoints_v2: each token is "X/Y[/L/lx/ly][/R/rx/ry]".
		for _, tok := range strings.Split(s, "|") {
			if tok == "" {
				continue
			}
			pp, err := deserializePolyPointV2(tok)
			if err == nil {
				p.points.Add(pp)
			}
		}
	} else if s := r.ReadStr("PolyPoints", ""); s != "" {
		// Legacy PolyPoints (v1): each token is "X\Y\type" (type is ignored).
		for _, tok := range strings.Split(s, "|") {
			parts := strings.Split(tok, `\`)
			if len(parts) < 2 {
				continue
			}
			x, err1 := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(parts[0], ",", ".")), 32)
			y, err2 := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(parts[1], ",", ".")), 32)
			if err1 == nil && err2 == nil {
				p.points.Add(&PolyPoint{X: float32(x), Y: float32(y)})
			}
		}
	}
	p.centerX = r.ReadFloat("CenterX", 0)
	p.centerY = r.ReadFloat("CenterY", 0)
	if s := r.ReadStr("DashPattern", ""); s != "" {
		fc, err := utils.ParseFloatCollection(s)
		if err == nil {
			p.dashPattern = []float32(fc)
		}
	}
	return nil
}

// PathPointType identifies how a path segment point is connected.
// Mirrors GDI+ PathPointType enum used in C# GraphicsPath.
type PathPointType byte

const (
	// PathPointTypeStart marks the beginning of a subpath.
	PathPointTypeStart PathPointType = 0
	// PathPointTypeLine marks a line-to endpoint.
	PathPointTypeLine PathPointType = 1
	// PathPointTypeBezier marks a cubic bezier control point or endpoint.
	PathPointTypeBezier PathPointType = 3
)

// PathPoint is one point in a vector path together with its segment type.
// Equivalent to GDI+ GraphicsPath element (PointF + PathPointType).
type PathPoint struct {
	X, Y float32
	Type PathPointType
}

// getPseudoPoint returns a pseudo bezier control point 1/3 of the way from
// start to end. Used when one side of a bezier segment lacks a real control
// point (PolyPoint.Left / Right == nil).
// Mirrors C# PolyLineObject.GetPseudoPoint (PolyLineObject.cs line 649).
func getPseudoPoint(start, end *PolyPoint) *PolyPoint {
	return &PolyPoint{
		X: start.X + (end.X-start.X)/3,
		Y: start.Y + (end.Y-start.Y)/3,
	}
}

// recalculateBoundsDelta computes the discriminant for a cubic bezier extremum.
// Mirrors C# PolyLineObject.RecalculateBounds_Delta (PolyLineObject.cs line 474).
func recalculateBoundsDelta(p0, p1, p2, p3 float32) float32 {
	return p1*p1 - p0*p2 - p1*p2 + p2*p2 + p0*p3 - p1*p3
}

// recalculateBoundsSolve solves for the bezier parameter t at an extremum.
// Mirrors C# PolyLineObject.RecalculateBounds_Solve (PolyLineObject.cs line 479).
func recalculateBoundsSolve(p0, p1, p2, p3, deltaSqrt float32) float32 {
	a := p0 - 3*p1 + 3*p2 - p3
	if a > -1e-4 && a < 1e-4 {
		denom := 2*p0 - 4*p1 + 2*p2
		if denom == 0 {
			return 0
		}
		return (p0 - p1) / denom
	}
	return (p0 - 2*p1 + p2 + deltaSqrt) / a
}

// recalculateBoundsValue evaluates a cubic bezier at parameter t.
// Mirrors C# PolyLineObject.RecalculateBounds_Value (PolyLineObject.cs line 488).
func recalculateBoundsValue(p0, p1, p2, p3, t float32) float32 {
	t1 := 1 - t
	return p0*t1*t1*t1 + 3*p1*t1*t1*t + 3*p2*t1*t*t + p3*t*t*t
}

// getPath generates the bezier path for the polyline/polygon.
// When closed is true (PolygonObject), an extra segment closes the path from
// the last point back to the first.
// Mirrors C# PolyLineObject.GetPath (PolyLineObject.cs lines 223-297).
func (p *PolyLineObject) getPath(closed bool, left, top, right, bottom, scaleX, scaleY float32) []PathPoint {
	n := p.points.Len()
	if n == 0 {
		return []PathPoint{
			{X: left * scaleX, Y: top * scaleX, Type: PathPointTypeStart},
			{X: (right + 1) * scaleX, Y: (bottom + 1) * scaleX, Type: PathPointTypeLine},
		}
	}
	if n == 1 {
		x := left + p.centerX + p.points.Get(0).X
		y := top + p.centerY + p.points.Get(0).Y
		return []PathPoint{
			{X: x * scaleX, Y: y * scaleX, Type: PathPointTypeStart},
			{X: (x + 1) * scaleX, Y: (y + 1) * scaleX, Type: PathPointTypeLine},
		}
	}

	cx, cy := p.centerX, p.centerY
	result := make([]PathPoint, 0, n*3)
	first := p.points.Get(0)
	result = append(result, PathPoint{
		X:    (first.X + left + cx) * scaleX,
		Y:    (first.Y + top + cy) * scaleY,
		Type: PathPointTypeStart,
	})

	count := n
	if closed {
		count++ // extra iteration closes polygon back to first point
	}

	prev := first
	for i := 1; i < count; i++ {
		point := p.points.Get(i % n) // % n handles the polygon-closing wrap-around

		if prev.Right != nil || point.Left != nil {
			// Bezier segment.
			var cp1 *PolyPoint
			if prev.Right != nil {
				cp1 = &PolyPoint{X: prev.X + prev.Right.X, Y: prev.Y + prev.Right.Y}
			} else {
				cp1 = getPseudoPoint(prev, point)
			}
			result = append(result, PathPoint{
				X:    (cp1.X + left + cx) * scaleX,
				Y:    (cp1.Y + top + cy) * scaleY,
				Type: PathPointTypeBezier,
			})

			var cp2 *PolyPoint
			if point.Left != nil {
				cp2 = &PolyPoint{X: point.X + point.Left.X, Y: point.Y + point.Left.Y}
			} else {
				cp2 = getPseudoPoint(point, prev)
			}
			result = append(result, PathPoint{
				X:    (cp2.X + left + cx) * scaleX,
				Y:    (cp2.Y + top + cy) * scaleY,
				Type: PathPointTypeBezier,
			})

			result = append(result, PathPoint{
				X:    (point.X + left + cx) * scaleX,
				Y:    (point.Y + top + cy) * scaleY,
				Type: PathPointTypeBezier,
			})
		} else {
			// Straight line segment.
			result = append(result, PathPoint{
				X:    (point.X + left + cx) * scaleX,
				Y:    (point.Y + top + cy) * scaleY,
				Type: PathPointTypeLine,
			})
		}
		prev = point
	}
	return result
}

// GetPath generates the bezier path for the polyline.
// Returns a slice of PathPoints (Type 0=start, 1=line, 3=bezier control/endpoint).
// Mirrors C# PolyLineObject.GetPath (PolyLineObject.cs lines 223-297).
func (p *PolyLineObject) GetPath(left, top, right, bottom, scaleX, scaleY float32) []PathPoint {
	return p.getPath(false, left, top, right, bottom, scaleX, scaleY)
}

// recalculateBounds updates Left/Top/Width/Height/CenterX/CenterY to tightly
// contain all anchor points and bezier curve extrema.
// When closed is true (PolygonObject), the closing segment is also included.
// Mirrors C# PolyLineObject.RecalculateBounds (PolyLineObject.cs lines 302-472).
func (p *PolyLineObject) recalculateBounds(closed bool) {
	n := p.points.Len()
	if n == 0 {
		p.centerX = 0
		p.centerY = 0
		p.SetWidth(5)
		p.SetHeight(5)
		return
	}

	first := p.points.Get(0)
	left := first.X
	top := first.Y
	right := first.X
	bottom := first.Y

	// Stage 1: bounding box of all anchor points.
	for i := 1; i < n; i++ {
		pt := p.points.Get(i)
		if pt.X < left {
			left = pt.X
		} else if pt.X > right {
			right = pt.X
		}
		if pt.Y < top {
			top = pt.Y
		} else if pt.Y > bottom {
			bottom = pt.Y
		}
	}

	// Stage 2: expand for bezier control point extrema.
	count := n
	if closed {
		count++
	}

	prev := first
	for i := 1; i < count; i++ {
		point := p.points.Get(i % n)

		needsCalculate := false
		var p1, p2 *PolyPoint

		if prev.Right != nil {
			p1 = &PolyPoint{X: prev.X + prev.Right.X, Y: prev.Y + prev.Right.Y}
			if p1.X < left || p1.X > right || p1.Y < top || p1.Y > bottom {
				needsCalculate = true
			}
		}
		if point.Left != nil {
			p2 = &PolyPoint{X: point.X + point.Left.X, Y: point.Y + point.Left.Y}
			if p2.X < left || p2.X > right || p2.Y < top || p2.Y > bottom {
				needsCalculate = true
			}
		}

		if needsCalculate {
			if p1 == nil {
				p1 = getPseudoPoint(prev, point)
			}
			if p2 == nil {
				p2 = getPseudoPoint(point, prev)
			}

			// X extrema.
			delta := recalculateBoundsDelta(prev.X, p1.X, p2.X, point.X)
			if delta > 0 {
				dSqrt := float32(math.Sqrt(float64(delta)))
				t1 := recalculateBoundsSolve(prev.X, p1.X, p2.X, point.X, -dSqrt)
				if t1 > 0 && t1 < 1 {
					x := recalculateBoundsValue(prev.X, p1.X, p2.X, point.X, t1)
					if x < left {
						left = x
					} else if x > right {
						right = x
					}
				}
				t2 := recalculateBoundsSolve(prev.X, p1.X, p2.X, point.X, dSqrt)
				if t2 > 0 && t2 < 1 {
					x := recalculateBoundsValue(prev.X, p1.X, p2.X, point.X, t2)
					if x < left {
						left = x
					} else if x > right {
						right = x
					}
				}
			}

			// Y extrema.
			delta = recalculateBoundsDelta(prev.Y, p1.Y, p2.Y, point.Y)
			if delta > 0 {
				dSqrt := float32(math.Sqrt(float64(delta)))
				t1 := recalculateBoundsSolve(prev.Y, p1.Y, p2.Y, point.Y, -dSqrt)
				if t1 > 0 && t1 < 1 {
					y := recalculateBoundsValue(prev.Y, p1.Y, p2.Y, point.Y, t1)
					if y < top {
						top = y
					} else if y > bottom {
						bottom = y
					}
				}
				t2 := recalculateBoundsSolve(prev.Y, p1.Y, p2.Y, point.Y, dSqrt)
				if t2 > 0 && t2 < 1 {
					y := recalculateBoundsValue(prev.Y, p1.Y, p2.Y, point.Y, t2)
					if y < top {
						top = y
					} else if y > bottom {
						bottom = y
					}
				}
			}
		}
		prev = point
	}

	// Update geometry.
	prevCX := p.centerX
	prevCY := p.centerY
	p.centerX = -left
	p.centerY = -top
	p.SetLeft(p.Left() + left + prevCX)
	p.SetTop(p.Top() + top + prevCY)
	p.SetHeight(bottom - top)
	p.SetWidth(right - left)
}

// RecalculateBounds updates the object's bounds to tightly contain all points.
// Mirrors C# PolyLineObject.RecalculateBounds (PolyLineObject.cs line 302).
func (p *PolyLineObject) RecalculateBounds() {
	p.recalculateBounds(false)
}

// SetPolyLine replaces all points with plain X/Y pairs (no bezier control
// points) and recalculates bounds.
// Mirrors C# PolyLineObject.SetPolyLine(PointF[]) (PolyLineObject.cs line 519).
func (p *PolyLineObject) SetPolyLine(newPoints [][2]float32) {
	p.points.Clear()
	if newPoints != nil {
		p.centerX = 0
		p.centerY = 0
		for _, pt := range newPoints {
			p.points.Add(&PolyPoint{X: pt[0], Y: pt[1]})
		}
	}
	l := p.Left()
	t := p.Top()
	p.RecalculateBounds()
	p.SetLeft(l)
	p.SetTop(t)
}

// addPoint appends a point at (localX, localY) and recalculates bounds.
// Mirrors C# PolyLineObject.addPoint (PolyLineObject.cs line 574).
func (p *PolyLineObject) addPoint(localX, localY float32) *PolyPoint {
	pt := &PolyPoint{X: localX, Y: localY}
	p.points.Add(pt)
	p.RecalculateBounds()
	return pt
}

// deletePoint removes the point at index and recalculates bounds.
// Mirrors C# PolyLineObject.deletePoint (PolyLineObject.cs line 588).
func (p *PolyLineObject) deletePoint(index int) {
	p.points.Remove(index)
	p.RecalculateBounds()
}

// insertPoint inserts a point at (localX, localY) at the given index and
// recalculates bounds.
// Mirrors C# PolyLineObject.insertPoint (PolyLineObject.cs line 620).
func (p *PolyLineObject) insertPoint(index int, localX, localY float32) *PolyPoint {
	pt := &PolyPoint{X: localX, Y: localY}
	p.points.Insert(index, pt)
	p.RecalculateBounds()
	return pt
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
// C# PolygonObject() sets FlagUseFill = true (PolygonObject.cs:88), meaning
// it participates in fill serialization. In Go, fill is always available via
// ReportComponentBase.Fill — no separate flag is needed.
func NewPolygonObject() *PolygonObject {
	return &PolygonObject{PolyLineObject: *NewPolyLineObject()}
}

// Serialize writes PolygonObject properties.
// C# PolygonObject.Serialize() (PolygonObject.cs:73-78) only sets
// Border.SimpleBorder = true and calls base.Serialize(), which handles
// PolyPoints_v2, CenterX, CenterY, and DashPattern.
func (pg *PolygonObject) Serialize(w report.Writer) error {
	return pg.PolyLineObject.Serialize(w)
}

// Deserialize reads PolygonObject properties.
// Delegates entirely to PolyLineObject.Deserialize; PolygonObject has no
// additional serialized fields.
func (pg *PolygonObject) Deserialize(r report.Reader) error {
	return pg.PolyLineObject.Deserialize(r)
}

// Assign copies all PolygonObject properties from src.
func (pg *PolygonObject) Assign(src *PolygonObject) {
	if src == nil {
		return
	}
	pg.PolyLineObject.Assign(&src.PolyLineObject)
}

// GetPath generates the closed bezier path for the polygon, connecting the
// last point back to the first.
// Mirrors C# PolyLineObject.GetPath with `this is PolygonObject` active
// (PolyLineObject.cs lines 249-253).
func (pg *PolygonObject) GetPath(left, top, right, bottom, scaleX, scaleY float32) []PathPoint {
	return pg.PolyLineObject.getPath(true, left, top, right, bottom, scaleX, scaleY)
}

// RecalculateBounds updates the polygon's bounds including the closing segment.
// Mirrors C# PolyLineObject.RecalculateBounds with `this is PolygonObject` active
// (PolyLineObject.cs lines 328-330).
func (pg *PolygonObject) RecalculateBounds() {
	pg.PolyLineObject.recalculateBounds(true)
}

// SetPolyLine replaces all points and recalculates closed polygon bounds.
func (pg *PolygonObject) SetPolyLine(newPoints [][2]float32) {
	pg.PolyLineObject.points.Clear()
	if newPoints != nil {
		pg.PolyLineObject.centerX = 0
		pg.PolyLineObject.centerY = 0
		for _, pt := range newPoints {
			pg.PolyLineObject.points.Add(&PolyPoint{X: pt[0], Y: pt[1]})
		}
	}
	l := pg.Left()
	t := pg.Top()
	pg.RecalculateBounds()
	pg.SetLeft(l)
	pg.SetTop(t)
}
