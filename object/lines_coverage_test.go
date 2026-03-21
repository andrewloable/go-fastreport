package object_test

// lines_coverage_test.go — additional coverage for lines.go:
// LineObject Serialize/Deserialize with non-default StartCap/EndCap,
// ShapeObject Serialize/Deserialize with non-Rectangle shape and Curve,
// CapSettings Assign/Clone/Equals methods, and SerializeCap/DeserializeCap helpers.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── CapSettings: Assign ──────────────────────────────────────────────────────

func TestCapSettings_Assign(t *testing.T) {
	src := object.CapSettings{Width: 16, Height: 12, Style: object.CapStyleArrow}
	var dst object.CapSettings
	dst.Assign(src)
	if dst.Width != 16 || dst.Height != 12 || dst.Style != object.CapStyleArrow {
		t.Errorf("Assign failed: got %+v", dst)
	}
}

// ── CapSettings: Clone ───────────────────────────────────────────────────────

func TestCapSettings_Clone(t *testing.T) {
	orig := object.CapSettings{Width: 10, Height: 20, Style: object.CapStyleDiamond}
	cloned := orig.Clone()
	if cloned.Width != 10 || cloned.Height != 20 || cloned.Style != object.CapStyleDiamond {
		t.Errorf("Clone values wrong: got %+v", cloned)
	}
	// Mutation of clone must not affect original.
	cloned.Width = 99
	if orig.Width != 10 {
		t.Error("Clone is not independent — orig.Width was mutated")
	}
}

// ── CapSettings: Equals ──────────────────────────────────────────────────────

func TestCapSettings_Equals_True(t *testing.T) {
	a := object.CapSettings{Width: 8, Height: 8, Style: object.CapStyleNone}
	b := object.CapSettings{Width: 8, Height: 8, Style: object.CapStyleNone}
	if !a.Equals(b) {
		t.Error("Equals should return true for identical values")
	}
}

func TestCapSettings_Equals_False_Width(t *testing.T) {
	a := object.CapSettings{Width: 8, Height: 8, Style: object.CapStyleNone}
	b := object.CapSettings{Width: 9, Height: 8, Style: object.CapStyleNone}
	if a.Equals(b) {
		t.Error("Equals should return false when Width differs")
	}
}

func TestCapSettings_Equals_False_Height(t *testing.T) {
	a := object.CapSettings{Width: 8, Height: 8, Style: object.CapStyleNone}
	b := object.CapSettings{Width: 8, Height: 16, Style: object.CapStyleNone}
	if a.Equals(b) {
		t.Error("Equals should return false when Height differs")
	}
}

func TestCapSettings_Equals_False_Style(t *testing.T) {
	a := object.CapSettings{Width: 8, Height: 8, Style: object.CapStyleNone}
	b := object.CapSettings{Width: 8, Height: 8, Style: object.CapStyleArrow}
	if a.Equals(b) {
		t.Error("Equals should return false when Style differs")
	}
}

// ── CapStyle enum: all values ────────────────────────────────────────────────

func TestCapStyleEnumValues(t *testing.T) {
	// C# enum: None=0, Circle=1, Square=2, Diamond=3, Arrow=4
	cases := []struct {
		style object.CapStyle
		want  int
	}{
		{object.CapStyleNone, 0},
		{object.CapStyleCircle, 1},
		{object.CapStyleSquare, 2},
		{object.CapStyleDiamond, 3},
		{object.CapStyleArrow, 4},
	}
	for _, tc := range cases {
		if int(tc.style) != tc.want {
			t.Errorf("CapStyle %v = %d, want %d", tc.style, int(tc.style), tc.want)
		}
	}
}

// ── LineObject: Serialize with non-default StartCap ──────────────────────────
// Verifies that dot-qualified attributes (StartCap.Style, StartCap.Width, etc.)
// are written, matching the FRX format produced by C# CapSettings.Serialize().

func TestLineObject_Serialize_NonDefaultStartCap(t *testing.T) {
	orig := object.NewLineObject()
	orig.StartCap = object.CapSettings{Width: 12, Height: 10, Style: object.CapStyleArrow}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("LineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	// C# serializes as StartCap.Style="Arrow", StartCap.Width="12", etc.
	if !strings.Contains(xml, `StartCap.Style="Arrow"`) {
		t.Errorf("expected StartCap.Style=\"Arrow\" in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `StartCap.Width="12"`) {
		t.Errorf("expected StartCap.Width=\"12\" in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `StartCap.Height="10"`) {
		t.Errorf("expected StartCap.Height=\"10\" in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.StartCap.Style != object.CapStyleArrow {
		t.Errorf("StartCap.Style: got %d, want CapStyleArrow", got.StartCap.Style)
	}
	if got.StartCap.Width != 12 {
		t.Errorf("StartCap.Width: got %v, want 12", got.StartCap.Width)
	}
	if got.StartCap.Height != 10 {
		t.Errorf("StartCap.Height: got %v, want 10", got.StartCap.Height)
	}
}

// ── LineObject: Serialize with non-default EndCap ────────────────────────────

func TestLineObject_Serialize_NonDefaultEndCap(t *testing.T) {
	orig := object.NewLineObject()
	orig.EndCap = object.CapSettings{Width: 6, Height: 6, Style: object.CapStyleCircle}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("LineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `EndCap.Style="Circle"`) {
		t.Errorf("expected EndCap.Style=\"Circle\" in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.EndCap.Style != object.CapStyleCircle {
		t.Errorf("EndCap.Style: got %d, want CapStyleCircle", got.EndCap.Style)
	}
}

// ── LineObject: Serialize with both non-default caps + diagonal ──────────────

func TestLineObject_Serialize_BothCapsAndDiagonal(t *testing.T) {
	orig := object.NewLineObject()
	orig.SetDiagonal(true)
	orig.StartCap = object.CapSettings{Width: 10, Height: 10, Style: object.CapStyleSquare}
	orig.EndCap = object.CapSettings{Width: 8, Height: 8, Style: object.CapStyleDiamond}

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("LineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Diagonal="true"`) {
		t.Errorf("expected Diagonal in XML:\n%s", xml)
	}
	// Only Style attributes differ here (Width and Height match defaults of 8);
	// StartCap has Width=10/Height=10 which both differ from default 8.
	if !strings.Contains(xml, `StartCap.Style="Square"`) {
		t.Errorf("expected StartCap.Style=\"Square\" in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `EndCap.Style="Diamond"`) {
		t.Errorf("expected EndCap.Style=\"Diamond\" in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !got.Diagonal() {
		t.Error("Diagonal should be true after round-trip")
	}
	if got.StartCap.Style != object.CapStyleSquare {
		t.Errorf("StartCap.Style: got %d, want CapStyleSquare", got.StartCap.Style)
	}
	if got.EndCap.Style != object.CapStyleDiamond {
		t.Errorf("EndCap.Style: got %d, want CapStyleDiamond", got.EndCap.Style)
	}
}

// ── ShapeObject: Serialize with non-Rectangle shape + Curve ──────────────────

func TestShapeObject_Serialize_RoundRect_WithCurve(t *testing.T) {
	orig := object.NewShapeObject()
	orig.SetShape(object.ShapeKindRoundRectangle)
	orig.SetCurve(15)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ShapeObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, "Shape=") {
		t.Errorf("expected Shape attribute in XML:\n%s", xml)
	}
	if !strings.Contains(xml, "Curve=") {
		t.Errorf("expected Curve attribute in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewShapeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Shape() != object.ShapeKindRoundRectangle {
		t.Errorf("Shape: got %d, want RoundRectangle", got.Shape())
	}
	if got.Curve() != 15 {
		t.Errorf("Curve: got %v, want 15", got.Curve())
	}
}

func TestShapeObject_Serialize_Triangle(t *testing.T) {
	orig := object.NewShapeObject()
	orig.SetShape(object.ShapeKindTriangle)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ShapeObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewShapeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Shape() != object.ShapeKindTriangle {
		t.Errorf("Shape: got %d, want Triangle", got.Shape())
	}
}

func TestShapeObject_Serialize_Diamond(t *testing.T) {
	orig := object.NewShapeObject()
	orig.SetShape(object.ShapeKindDiamond)
	orig.SetCurve(5)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ShapeObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewShapeObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Shape() != object.ShapeKindDiamond {
		t.Errorf("Shape: got %d, want Diamond", got.Shape())
	}
	if got.Curve() != 5 {
		t.Errorf("Curve: got %v, want 5", got.Curve())
	}
}

// ── LineObject: Deserialize with FRX-format dot-qualified cap attributes ──────
// This tests reading real FRX files like "Lines and Shapes.frx" which contain
// attributes like StartCap.Style="Arrow", EndCap.Height="16", EndCap.Style="Diamond".

func TestLineObject_Deserialize_ExplicitCaps(t *testing.T) {
	// Use the FRX attribute format that C# CapSettings.Serialize() produces.
	// See test-reports/Lines and Shapes.frx lines 23-26 for real examples.
	xmlStr := `<LineObject StartCap.Width="10" StartCap.Height="10" StartCap.Style="Arrow" EndCap.Height="16" EndCap.Style="Circle"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.StartCap.Style != object.CapStyleArrow {
		t.Errorf("StartCap.Style: got %d, want CapStyleArrow", got.StartCap.Style)
	}
	if got.StartCap.Width != 10 {
		t.Errorf("StartCap.Width: got %v, want 10", got.StartCap.Width)
	}
	if got.EndCap.Style != object.CapStyleCircle {
		t.Errorf("EndCap.Style: got %d, want CapStyleCircle", got.EndCap.Style)
	}
	if got.EndCap.Height != 16 {
		t.Errorf("EndCap.Height: got %v, want 16", got.EndCap.Height)
	}
}

// ── LineObject: Deserialize FRX real-world example (only Style attribute) ─────
// FRX files often only serialize non-default fields. Line5 in "Lines and Shapes.frx"
// has only StartCap.Style="Arrow" with Width/Height remaining at their defaults (8).

func TestLineObject_Deserialize_StyleOnly(t *testing.T) {
	xmlStr := `<LineObject StartCap.Style="Arrow"/>`
	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.StartCap.Style != object.CapStyleArrow {
		t.Errorf("StartCap.Style: got %d, want CapStyleArrow", got.StartCap.Style)
	}
	// Width and Height should remain at defaults (8) since not specified.
	if got.StartCap.Width != 8 {
		t.Errorf("StartCap.Width: got %v, want 8 (default)", got.StartCap.Width)
	}
	if got.StartCap.Height != 8 {
		t.Errorf("StartCap.Height: got %v, want 8 (default)", got.StartCap.Height)
	}
}

// ── LineObject: Serialize default caps produces no cap attributes ─────────────

func TestLineObject_Serialize_DefaultCaps_NoAttributes(t *testing.T) {
	// When caps are at their default values, no cap attributes should be written
	// (diff-encoding matches C# CapSettings.Serialize behavior).
	orig := object.NewLineObject()

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("LineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if strings.Contains(xml, "StartCap") {
		t.Errorf("should NOT write StartCap attributes for default caps:\n%s", xml)
	}
	if strings.Contains(xml, "EndCap") {
		t.Errorf("should NOT write EndCap attributes for default caps:\n%s", xml)
	}
}

// ── LineObject: DashPattern serialize / deserialize round-trip ────────────────
// C# LineObject.Serialize() (LineObject.cs:274-275): writes DashPattern only
// when Count > 0.

func TestLineObject_DashPattern_RoundTrip(t *testing.T) {
	orig := object.NewLineObject()
	orig.SetDashPattern([]float32{4, 2, 1, 2})

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("LineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `DashPattern=`) {
		t.Errorf("expected DashPattern attribute in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	dp := got.DashPattern()
	if len(dp) != 4 {
		t.Fatalf("DashPattern: got %d elements, want 4", len(dp))
	}
	for i, want := range []float32{4, 2, 1, 2} {
		if dp[i] != want {
			t.Errorf("DashPattern[%d] = %v, want %v", i, dp[i], want)
		}
	}
}

func TestLineObject_DashPattern_Empty_NotWritten(t *testing.T) {
	// When DashPattern is empty, the attribute must NOT be written.
	orig := object.NewLineObject()
	// dashPattern is nil by default — confirm nothing written.

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("LineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	if strings.Contains(buf.String(), "DashPattern") {
		t.Errorf("DashPattern should NOT be written when empty:\n%s", buf.String())
	}
}

// ── PolyLineObject: Serialize / Deserialize (PolyPoints_v2) ──────────────────
// C# PolyLineObject.Serialize() writes PolyPoints_v2, CenterX, CenterY, and
// DashPattern (PolyLineObject.cs lines 496-517).

func TestPolyLineObject_Serialize_TwoPoints(t *testing.T) {
	orig := object.NewPolyLineObject()
	orig.Points().Add(&object.PolyPoint{X: 0, Y: 0})
	orig.Points().Add(&object.PolyPoint{X: 100, Y: 50})
	orig.SetCenterX(10)
	orig.SetCenterY(20)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PolyLineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, "PolyPoints_v2=") {
		t.Errorf("expected PolyPoints_v2 in XML:\n%s", xml)
	}
	if !strings.Contains(xml, "CenterX=") {
		t.Errorf("expected CenterX in XML:\n%s", xml)
	}
	if !strings.Contains(xml, "CenterY=") {
		t.Errorf("expected CenterY in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPolyLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Points().Len() != 2 {
		t.Fatalf("Points().Len() = %d, want 2", got.Points().Len())
	}
	if got.Points().Get(0).X != 0 || got.Points().Get(0).Y != 0 {
		t.Errorf("Point[0]: got (%v,%v), want (0,0)", got.Points().Get(0).X, got.Points().Get(0).Y)
	}
	if got.Points().Get(1).X != 100 || got.Points().Get(1).Y != 50 {
		t.Errorf("Point[1]: got (%v,%v), want (100,50)", got.Points().Get(1).X, got.Points().Get(1).Y)
	}
	if got.CenterX() != 10 {
		t.Errorf("CenterX: got %v, want 10", got.CenterX())
	}
	if got.CenterY() != 20 {
		t.Errorf("CenterY: got %v, want 20", got.CenterY())
	}
}

func TestPolyLineObject_Serialize_DashPattern(t *testing.T) {
	orig := object.NewPolyLineObject()
	orig.SetDashPattern([]float32{6, 3})
	orig.Points().Add(&object.PolyPoint{X: 0, Y: 0})

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PolyLineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	if !strings.Contains(buf.String(), "DashPattern=") {
		t.Errorf("expected DashPattern in XML:\n%s", buf.String())
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPolyLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	dp := got.DashPattern()
	if len(dp) != 2 || dp[0] != 6 || dp[1] != 3 {
		t.Errorf("DashPattern: got %v, want [6 3]", dp)
	}
}

// ── PolyLineObject: Deserialize legacy PolyPoints v1 format ──────────────────
// FRX files older than PolyPoints_v2 (e.g. Box.frx) use the format
// "X\Y\type|X\Y\type|…" where type is ignored.
// C# PolyLineObject.Deserialize() handles this as the "PolyPoints" branch
// (PolyLineObject.cs lines 155-168).

func TestPolyLineObject_Deserialize_LegacyPolyPointsV1(t *testing.T) {
	// Real example from test-reports/Box.frx (Polygon11).
	xmlStr := `<PolyLineObject PolyPoints="0\0\0|378\0\1|378\491.4\1|0\491.4\1" CenterX="0" CenterY="0"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPolyLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Points().Len() != 4 {
		t.Fatalf("Points().Len() = %d, want 4", got.Points().Len())
	}
	// First point: (0, 0)
	p0 := got.Points().Get(0)
	if p0.X != 0 || p0.Y != 0 {
		t.Errorf("Point[0]: got (%v,%v), want (0,0)", p0.X, p0.Y)
	}
	// Second point: (378, 0)
	p1 := got.Points().Get(1)
	if p1.X != 378 || p1.Y != 0 {
		t.Errorf("Point[1]: got (%v,%v), want (378,0)", p1.X, p1.Y)
	}
	// Fourth point: (0, 491.4)
	p3 := got.Points().Get(3)
	if p3.X != 0 || p3.Y != 491.4 {
		t.Errorf("Point[3]: got (%v,%v), want (0,491.4)", p3.X, p3.Y)
	}
	if got.CenterX() != 0 || got.CenterY() != 0 {
		t.Errorf("Center: got (%v,%v), want (0,0)", got.CenterX(), got.CenterY())
	}
}

// ── PolyLineObject: Deserialize PolyPoints_v2 with bezier L/R curves ─────────
// C# PolyPoint.Serialize(StringBuilder) and Deserialize(string) use
// "X/Y[/L/lx/ly][/R/rx/ry]" format.

func TestPolyLineObject_Deserialize_V2_WithBezierCurves(t *testing.T) {
	// Construct a PolyPoints_v2 string with a point that has both L and R curves.
	xmlStr := `<PolyLineObject PolyPoints_v2="0/0|50/100/L/-10/-5/R/10/5|100/0" CenterX="5" CenterY="3"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPolyLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Points().Len() != 3 {
		t.Fatalf("Points().Len() = %d, want 3", got.Points().Len())
	}
	// Middle point should have Left and Right bezier control points.
	mid := got.Points().Get(1)
	if mid.X != 50 || mid.Y != 100 {
		t.Errorf("mid: got (%v,%v), want (50,100)", mid.X, mid.Y)
	}
	if mid.Left == nil {
		t.Fatal("mid.Left should not be nil")
	}
	if mid.Left.X != -10 || mid.Left.Y != -5 {
		t.Errorf("mid.Left: got (%v,%v), want (-10,-5)", mid.Left.X, mid.Left.Y)
	}
	if mid.Right == nil {
		t.Fatal("mid.Right should not be nil")
	}
	if mid.Right.X != 10 || mid.Right.Y != 5 {
		t.Errorf("mid.Right: got (%v,%v), want (10,5)", mid.Right.X, mid.Right.Y)
	}
	if got.CenterX() != 5 || got.CenterY() != 3 {
		t.Errorf("Center: got (%v,%v), want (5,3)", got.CenterX(), got.CenterY())
	}
}

// ── PolyLineObject: bezier L/R round-trip via Serialize/Deserialize ──────────

func TestPolyLineObject_BezierRoundTrip(t *testing.T) {
	orig := object.NewPolyLineObject()
	p0 := &object.PolyPoint{X: 0, Y: 0}
	p1 := &object.PolyPoint{
		X:     50,
		Y:     100,
		Left:  &object.PolyPoint{X: -10, Y: -5},
		Right: &object.PolyPoint{X: 10, Y: 5},
	}
	p2 := &object.PolyPoint{X: 100, Y: 0}
	orig.Points().Add(p0)
	orig.Points().Add(p1)
	orig.Points().Add(p2)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PolyLineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPolyLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Points().Len() != 3 {
		t.Fatalf("Points().Len() = %d, want 3", got.Points().Len())
	}
	mid := got.Points().Get(1)
	if mid.Left == nil || mid.Right == nil {
		t.Fatal("bezier control points lost after round-trip")
	}
	if mid.Left.X != -10 || mid.Left.Y != -5 {
		t.Errorf("mid.Left after round-trip: got (%v,%v), want (-10,-5)", mid.Left.X, mid.Left.Y)
	}
	if mid.Right.X != 10 || mid.Right.Y != 5 {
		t.Errorf("mid.Right after round-trip: got (%v,%v), want (10,5)", mid.Right.X, mid.Right.Y)
	}
}

// ── PolygonObject: Serialize / Deserialize ────────────────────────────────────
// C# PolygonObject.Serialize() calls base.Serialize() (PolygonObject.cs:76-77),
// so PolyPoints_v2 + CenterX/Y are written by PolyLineObject.Serialize().

func TestPolygonObject_Serialize_Deserialize_RoundTrip(t *testing.T) {
	orig := object.NewPolygonObject()
	orig.Points().Add(&object.PolyPoint{X: 0, Y: 0})
	orig.Points().Add(&object.PolyPoint{X: 200, Y: 0})
	orig.Points().Add(&object.PolyPoint{X: 100, Y: 150})
	orig.SetCenterX(0)
	orig.SetCenterY(0)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PolygonObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, "PolyPoints_v2=") {
		t.Errorf("expected PolyPoints_v2 in PolygonObject XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPolygonObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Points().Len() != 3 {
		t.Fatalf("Points().Len() = %d, want 3", got.Points().Len())
	}
	if got.Points().Get(1).X != 200 || got.Points().Get(1).Y != 0 {
		t.Errorf("Point[1]: got (%v,%v), want (200,0)",
			got.Points().Get(1).X, got.Points().Get(1).Y)
	}
}

// ── PolygonObject: Deserialize legacy PolyPoints v1 from Box.frx ─────────────

func TestPolygonObject_Deserialize_LegacyPolyPointsV1(t *testing.T) {
	// Taken verbatim from test-reports/Box.frx (Polygon8).
	xmlStr := `<PolygonObject Name="Polygon8" Left="66.15" Top="633.15" Width="378" Height="75.6" Fill.Color="Orange" PolyPoints="0\0\0|0\-75.6\1|378\-75.6\1|378\0\1" CenterX="0" CenterY="75.6"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPolygonObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Points().Len() != 4 {
		t.Fatalf("Points().Len() = %d, want 4", got.Points().Len())
	}
	// Point[1] should be (0, -75.6)
	p1 := got.Points().Get(1)
	if p1.X != 0 {
		t.Errorf("Point[1].X: got %v, want 0", p1.X)
	}
	if p1.Y != -75.6 {
		t.Errorf("Point[1].Y: got %v, want -75.6", p1.Y)
	}
	if got.CenterY() != 75.6 {
		t.Errorf("CenterY: got %v, want 75.6", got.CenterY())
	}
}

// ── PolyLineObject: empty points list serializes/deserializes cleanly ─────────

func TestPolyLineObject_Serialize_EmptyPoints(t *testing.T) {
	orig := object.NewPolyLineObject()
	// No points added — PolyPoints_v2 should be an empty string.

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("PolyLineObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewPolyLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Points().Len() != 0 {
		t.Errorf("Points().Len() = %d, want 0", got.Points().Len())
	}
}
