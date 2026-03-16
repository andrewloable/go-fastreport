package object_test

// container_misc_test.go — coverage tests for container.go, html.go, svg.go,
// and digital_signature.go uncovered branches.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── ContainerObject: SetChildOrder ───────────────────────────────────────────

type minimalBase struct {
	name   string
	parent report.Parent
}

func (m *minimalBase) Name() string              { return m.name }
func (m *minimalBase) SetName(n string)          { m.name = n }
func (m *minimalBase) BaseName() string          { return "Base" }
func (m *minimalBase) Parent() report.Parent     { return m.parent }
func (m *minimalBase) SetParent(p report.Parent) { m.parent = p }
func (m *minimalBase) Serialize(w report.Writer) error { return nil }
func (m *minimalBase) Deserialize(r report.Reader) error { return nil }

func TestContainerObject_SetChildOrder_Reorder(t *testing.T) {
	c := object.NewContainerObject()

	child1 := &minimalBase{name: "C1"}
	child2 := &minimalBase{name: "C2"}
	child3 := &minimalBase{name: "C3"}

	c.AddChild(child1)
	c.AddChild(child2)
	c.AddChild(child3)

	// Move child1 from position 0 to position 2.
	c.SetChildOrder(child1, 2)

	objs := c.Objects()
	if objs.Len() != 3 {
		t.Fatalf("expected 3 children, got %d", objs.Len())
	}
	// After reorder, child1 should be at position 2.
	if c.GetChildOrder(child1) != 2 {
		t.Errorf("child1 order: got %d, want 2", c.GetChildOrder(child1))
	}
}

func TestContainerObject_SetChildOrder_OrderExceedsLen(t *testing.T) {
	c := object.NewContainerObject()
	child1 := &minimalBase{name: "A"}
	child2 := &minimalBase{name: "B"}
	c.AddChild(child1)
	c.AddChild(child2)

	// Order > len(children) - 1 should clamp to end.
	c.SetChildOrder(child1, 100)
	// Should not panic and child1 should end up at the end.
	if c.GetChildOrder(child1) < 0 {
		t.Error("child1 should still be in the container")
	}
}

func TestContainerObject_SetChildOrder_NotFound(t *testing.T) {
	c := object.NewContainerObject()
	child1 := &minimalBase{name: "A"}
	stranger := &minimalBase{name: "Stranger"}
	c.AddChild(child1)

	// stranger is not in c — SetChildOrder should be a no-op.
	c.SetChildOrder(stranger, 0) // should not panic
}

// ── ContainerObject: UpdateLayout ────────────────────────────────────────────

func TestContainerObject_UpdateLayout_NoOpCoverage(t *testing.T) {
	c := object.NewContainerObject()
	// UpdateLayout is documented as a no-op in the base implementation.
	c.UpdateLayout(10, 20) // should not panic
}

// ── ContainerObject: Serialize with events ───────────────────────────────────

func TestContainerObject_Serialize_WithEvents(t *testing.T) {
	c := object.NewContainerObject()
	c.SetBeforeLayoutEvent("OnBeforeLayout")
	c.SetAfterLayoutEvent("OnAfterLayout")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ContainerObject", c); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `BeforeLayoutEvent=`) {
		t.Errorf("expected BeforeLayoutEvent in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `AfterLayoutEvent=`) {
		t.Errorf("expected AfterLayoutEvent in XML:\n%s", xml)
	}
}

// ── SubreportObject: Serialize/Deserialize with ReportPage + PrintOnParent ───

func TestSubreportObject_SerializeDeserialize_NonDefaults(t *testing.T) {
	orig := object.NewSubreportObject()
	orig.SetReportPageName("Page2")
	orig.SetPrintOnParent(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SubreportObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `ReportPage=`) {
		t.Errorf("expected ReportPage in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `PrintOnParent="true"`) {
		t.Errorf("expected PrintOnParent in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSubreportObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ReportPageName() != "Page2" {
		t.Errorf("ReportPageName: got %q, want Page2", got.ReportPageName())
	}
	if !got.PrintOnParent() {
		t.Error("PrintOnParent should be true")
	}
}

// ── HtmlObject: Serialize/Deserialize with RightToLeft ───────────────────────

func TestHtmlObject_SerializeDeserialize_RightToLeft(t *testing.T) {
	orig := object.NewHtmlObject()
	orig.SetRightToLeft(true)
	orig.SetText("<b>Hello</b>")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("HtmlObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `RightToLeft="true"`) {
		t.Errorf("expected RightToLeft in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewHtmlObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !got.RightToLeft() {
		t.Error("RightToLeft should be true after round-trip")
	}
}

// ── SVGObject: Serialize/Deserialize with SvgData ────────────────────────────

func TestSVGObject_SerializeDeserialize_WithData(t *testing.T) {
	orig := object.NewSVGObject()
	orig.SvgData = "PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmci/>"

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SVGObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `SvgData=`) {
		t.Errorf("expected SvgData in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSVGObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.SvgData != orig.SvgData {
		t.Errorf("SvgData: got %q, want %q", got.SvgData, orig.SvgData)
	}
}

// ── DigitalSignatureObject: Serialize/Deserialize with Placeholder ────────────

func TestDigitalSignatureObject_SerializeDeserialize_WithPlaceholder(t *testing.T) {
	orig := object.NewDigitalSignatureObject()
	orig.SetPlaceholder("Sign here")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("DigitalSignatureObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Placeholder=`) {
		t.Errorf("expected Placeholder in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewDigitalSignatureObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Placeholder() != "Sign here" {
		t.Errorf("Placeholder: got %q, want 'Sign here'", got.Placeholder())
	}
}
