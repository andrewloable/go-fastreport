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
