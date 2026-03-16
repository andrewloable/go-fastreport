package object_test

// lines_coverage_test.go — additional coverage for lines.go:
// LineObject Serialize/Deserialize with non-default StartCap/EndCap,
// ShapeObject Serialize/Deserialize with non-Rectangle shape and Curve.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── LineObject: Serialize with non-default StartCap ──────────────────────────

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
	if !strings.Contains(xml, "StartCap=") {
		t.Errorf("expected StartCap attribute in XML:\n%s", xml)
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
	if !strings.Contains(xml, "EndCap=") {
		t.Errorf("expected EndCap attribute in XML:\n%s", xml)
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
	if !strings.Contains(xml, "StartCap=") {
		t.Errorf("expected StartCap in XML:\n%s", xml)
	}
	if !strings.Contains(xml, "EndCap=") {
		t.Errorf("expected EndCap in XML:\n%s", xml)
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

// ── LineObject: Deserialize with explicit StartCap/EndCap strings ─────────────

func TestLineObject_Deserialize_ExplicitCaps(t *testing.T) {
	// Direct XML with StartCap and EndCap attributes.
	xmlStr := `<LineObject StartCap="10,10,4" EndCap="6,6,1"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewLineObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// Style=4 → CapStyleArrow
	if got.StartCap.Style != object.CapStyleArrow {
		t.Errorf("StartCap.Style: got %d, want CapStyleArrow(4)", got.StartCap.Style)
	}
	// Style=1 → CapStyleCircle
	if got.EndCap.Style != object.CapStyleCircle {
		t.Errorf("EndCap.Style: got %d, want CapStyleCircle(1)", got.EndCap.Style)
	}
}
