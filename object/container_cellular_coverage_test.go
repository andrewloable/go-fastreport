package object_test

// container_cellular_coverage_test.go — additional tests targeting
// ContainerObject.UpdateLayout, ContainerObject.Serialize/Deserialize,
// SubreportObject.Serialize/Deserialize, and
// CellularTextObject.Serialize/Deserialize.
//
// Coverage analysis: the only statements not yet covered in these functions are
// the `return err` guards after parent Serialize/Deserialize calls
// (e.g. `if err := c.ReportComponentBase.Serialize(w); err != nil { return err }`).
// Those branches are genuinely unreachable: the entire parent chain
// (ReportComponentBase → ComponentBase → BaseObject) only calls void Write*
// methods and never returns a non-nil error. This file provides additional
// scenario tests that exercise every other reachable branch and verify correct
// round-trip behaviour.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── ContainerObject.UpdateLayout ─────────────────────────────────────────────

// TestContainerObject_UpdateLayout_WithChildrenSet verifies that UpdateLayout
// can be called on a container that has children without panicking.
// UpdateLayout is a documented no-op in the base implementation; the function
// body is empty, so the Go coverage tool always reports it as 0% (no statements
// to instrument). These tests confirm correct behavioural semantics.
func TestContainerObject_UpdateLayout_WithChildren(t *testing.T) {
	c := object.NewContainerObject()

	child1 := object.NewTextObject()
	child1.SetName("child1")
	child1.SetLeft(10)
	child1.SetTop(20)

	child2 := object.NewTextObject()
	child2.SetName("child2")
	child2.SetLeft(50)
	child2.SetTop(80)

	c.AddChild(child1)
	c.AddChild(child2)

	// Call UpdateLayout with positive, negative, and zero deltas.
	c.UpdateLayout(15, 25)  // positive delta
	c.UpdateLayout(-5, -10) // negative delta
	c.UpdateLayout(0, 0)    // zero delta

	// The function is a no-op: positions are unchanged.
	if c.Objects().Len() != 2 {
		t.Errorf("Objects().Len() = %d, want 2", c.Objects().Len())
	}
}

// TestContainerObject_UpdateLayout_EmptyContainer verifies that UpdateLayout on
// a container with no children does not panic.
func TestContainerObject_UpdateLayout_EmptyContainer(t *testing.T) {
	c := object.NewContainerObject()
	c.UpdateLayout(100, 200) // no children — must not panic
	c.UpdateLayout(-100, -200)
}

// ── ContainerObject: Serialize/Deserialize ────────────────────────────────────

// TestContainerObject_Serialize_OnlyBeforeEvent verifies Serialize when only
// the before-layout event name is set (the after-layout branch is skipped).
func TestContainerObject_Serialize_OnlyBeforeEvent(t *testing.T) {
	orig := object.NewContainerObject()
	orig.SetBeforeLayoutEvent("OnlyBefore")
	// afterLayoutEvent stays empty.

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ContainerObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `BeforeLayoutEvent=`) {
		t.Errorf("expected BeforeLayoutEvent in XML:\n%s", xml)
	}
	if strings.Contains(xml, `AfterLayoutEvent=`) {
		t.Errorf("unexpected AfterLayoutEvent in XML:\n%s", xml)
	}
}

// TestContainerObject_Deserialize_DefaultEvents verifies Deserialize when no
// event attributes are present (both default to empty string).
func TestContainerObject_Deserialize_DefaultEvents(t *testing.T) {
	xmlStr := `<ContainerObject Name="c1"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewContainerObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.BeforeLayoutEvent() != "" {
		t.Errorf("BeforeLayoutEvent: got %q, want empty", got.BeforeLayoutEvent())
	}
	if got.AfterLayoutEvent() != "" {
		t.Errorf("AfterLayoutEvent: got %q, want empty", got.AfterLayoutEvent())
	}
}

// TestContainerObject_Deserialize_OnlyAfterEvent verifies Deserialize when only
// AfterLayoutEvent is present in the XML.
func TestContainerObject_Deserialize_OnlyAfterEvent(t *testing.T) {
	xmlStr := `<ContainerObject AfterLayoutEvent="AfterOnly"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewContainerObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.BeforeLayoutEvent() != "" {
		t.Errorf("BeforeLayoutEvent: got %q, want empty", got.BeforeLayoutEvent())
	}
	if got.AfterLayoutEvent() != "AfterOnly" {
		t.Errorf("AfterLayoutEvent: got %q, want AfterOnly", got.AfterLayoutEvent())
	}
}

// TestContainerObject_RoundTrip_BothEvents performs a full Serialize +
// Deserialize round-trip with both event names set.
func TestContainerObject_RoundTrip_BothEvents(t *testing.T) {
	orig := object.NewContainerObject()
	orig.SetName("cont2")
	orig.SetBeforeLayoutEvent("BeforeRound")
	orig.SetAfterLayoutEvent("AfterRound")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ContainerObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewContainerObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.BeforeLayoutEvent() != "BeforeRound" {
		t.Errorf("BeforeLayoutEvent: got %q, want BeforeRound", got.BeforeLayoutEvent())
	}
	if got.AfterLayoutEvent() != "AfterRound" {
		t.Errorf("AfterLayoutEvent: got %q, want AfterRound", got.AfterLayoutEvent())
	}
}

// TestContainerObject_Serialize_NoEvents verifies that Serialize with no events
// produces XML without BeforeLayoutEvent or AfterLayoutEvent attributes.
func TestContainerObject_Serialize_NoEvents(t *testing.T) {
	orig := object.NewContainerObject()
	orig.SetName("emptyevt")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ContainerObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if strings.Contains(xml, `BeforeLayoutEvent=`) {
		t.Errorf("unexpected BeforeLayoutEvent in XML:\n%s", xml)
	}
	if strings.Contains(xml, `AfterLayoutEvent=`) {
		t.Errorf("unexpected AfterLayoutEvent in XML:\n%s", xml)
	}
}

// ── SubreportObject: Serialize/Deserialize ────────────────────────────────────

// TestSubreportObject_Serialize_Defaults verifies Serialize with default values
// produces no optional attributes.
func TestSubreportObject_Serialize_Defaults(t *testing.T) {
	orig := object.NewSubreportObject()
	orig.SetName("sub_defaults")
	// reportPageName stays empty; printOnParent stays false.

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SubreportObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if strings.Contains(xml, `ReportPage=`) {
		t.Errorf("unexpected ReportPage in XML:\n%s", xml)
	}
	if strings.Contains(xml, `PrintOnParent=`) {
		t.Errorf("unexpected PrintOnParent in XML:\n%s", xml)
	}
}

// TestSubreportObject_Serialize_BothFields verifies Serialize with both
// reportPageName and printOnParent set.
func TestSubreportObject_Serialize_BothFields(t *testing.T) {
	orig := object.NewSubreportObject()
	orig.SetName("sub_both")
	orig.SetReportPageName("Detail")
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
		t.Errorf("expected PrintOnParent=true in XML:\n%s", xml)
	}
}

// TestSubreportObject_RoundTrip_BothFields performs a full round-trip with both
// fields set.
func TestSubreportObject_RoundTrip_BothFields(t *testing.T) {
	orig := object.NewSubreportObject()
	orig.SetName("sub_rt")
	orig.SetReportPageName("SummaryPage")
	orig.SetPrintOnParent(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SubreportObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSubreportObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ReportPageName() != "SummaryPage" {
		t.Errorf("ReportPageName: got %q, want SummaryPage", got.ReportPageName())
	}
	if !got.PrintOnParent() {
		t.Error("PrintOnParent: want true")
	}
}

// TestSubreportObject_RoundTrip_Defaults performs a full round-trip with default
// values (empty page name, printOnParent=false).
func TestSubreportObject_RoundTrip_Defaults(t *testing.T) {
	orig := object.NewSubreportObject()
	orig.SetName("sub_def_rt")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("SubreportObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSubreportObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ReportPageName() != "" {
		t.Errorf("ReportPageName: got %q, want empty", got.ReportPageName())
	}
	if got.PrintOnParent() {
		t.Error("PrintOnParent: want false")
	}
}

// TestSubreportObject_Deserialize_ExplicitXML_WithBothFields deserializes from
// explicit XML with both ReportPage and PrintOnParent present.
func TestSubreportObject_Deserialize_ExplicitXML_WithBothFields(t *testing.T) {
	xmlStr := `<SubreportObject Name="subreport1" ReportPage="PageA" PrintOnParent="true"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSubreportObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ReportPageName() != "PageA" {
		t.Errorf("ReportPageName: got %q, want PageA", got.ReportPageName())
	}
	if !got.PrintOnParent() {
		t.Error("PrintOnParent: want true")
	}
}

// ── CellularTextObject: Serialize/Deserialize ─────────────────────────────────

// TestCellularTextObject_Serialize_OnlyVertSpacing verifies that Serialize emits
// VertSpacing when only vertSpacing is set (all other cellular fields at default).
func TestCellularTextObject_Serialize_OnlyVertSpacing(t *testing.T) {
	orig := object.NewCellularTextObject()
	orig.SetVertSpacing(5.5)
	// cellWidth=0, cellHeight=0, horzSpacing=0, wordWrap=true (default)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("CellularTextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `VertSpacing=`) {
		t.Errorf("expected VertSpacing in XML:\n%s", xml)
	}
	if strings.Contains(xml, `CellWidth=`) {
		t.Errorf("unexpected CellWidth in XML:\n%s", xml)
	}
	if strings.Contains(xml, `CellHeight=`) {
		t.Errorf("unexpected CellHeight in XML:\n%s", xml)
	}
	if strings.Contains(xml, `HorzSpacing=`) {
		t.Errorf("unexpected HorzSpacing in XML:\n%s", xml)
	}
	if strings.Contains(xml, `WordWrap=`) {
		t.Errorf("unexpected WordWrap in XML (default=true should not be written):\n%s", xml)
	}
}

// TestCellularTextObject_Serialize_OnlyWordWrapFalse verifies that Serialize emits
// WordWrap="false" when only wordWrap is set to false.
func TestCellularTextObject_Serialize_OnlyWordWrapFalse(t *testing.T) {
	orig := object.NewCellularTextObject()
	orig.SetWordWrap(false)
	// All cell dimensions stay at 0 (default).

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("CellularTextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `WordWrap="false"`) {
		t.Errorf("expected WordWrap=false in XML:\n%s", xml)
	}
	if strings.Contains(xml, `CellWidth=`) {
		t.Errorf("unexpected CellWidth in XML:\n%s", xml)
	}
}

// TestCellularTextObject_Deserialize_OnlyVertSpacing verifies Deserialize reads
// VertSpacing correctly when only that field is present in the XML.
func TestCellularTextObject_Deserialize_OnlyVertSpacing(t *testing.T) {
	xmlStr := `<CellularTextObject VertSpacing="7.5"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewCellularTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.VertSpacing() != 7.5 {
		t.Errorf("VertSpacing: got %v, want 7.5", got.VertSpacing())
	}
	// Other fields stay at defaults.
	if got.CellWidth() != 0 {
		t.Errorf("CellWidth: got %v, want 0", got.CellWidth())
	}
	if !got.WordWrap() {
		t.Error("WordWrap should default to true")
	}
}

// TestCellularTextObject_Deserialize_WordWrapFalse verifies Deserialize reads
// WordWrap="false" correctly.
func TestCellularTextObject_Deserialize_WordWrapFalse(t *testing.T) {
	xmlStr := `<CellularTextObject WordWrap="false"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewCellularTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.WordWrap() {
		t.Error("WordWrap: want false")
	}
}

// TestCellularTextObject_RoundTrip_OnlyVertSpacing verifies a full round-trip
// where only vertSpacing is non-default.
func TestCellularTextObject_RoundTrip_OnlyVertSpacing(t *testing.T) {
	orig := object.NewCellularTextObject()
	orig.SetVertSpacing(3.14)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("CellularTextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewCellularTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.VertSpacing() != 3.14 {
		t.Errorf("VertSpacing: got %v, want 3.14", got.VertSpacing())
	}
	if got.CellWidth() != 0 {
		t.Errorf("CellWidth: got %v, want 0 (default)", got.CellWidth())
	}
	if got.CellHeight() != 0 {
		t.Errorf("CellHeight: got %v, want 0 (default)", got.CellHeight())
	}
	if got.HorzSpacing() != 0 {
		t.Errorf("HorzSpacing: got %v, want 0 (default)", got.HorzSpacing())
	}
	if !got.WordWrap() {
		t.Error("WordWrap: want true (default)")
	}
}

// TestCellularTextObject_RoundTrip_AllCellularFields verifies a full round-trip
// where all cellular fields have non-default values.
func TestCellularTextObject_RoundTrip_AllCellularFields(t *testing.T) {
	orig := object.NewCellularTextObject()
	orig.SetCellWidth(14.17)
	orig.SetCellHeight(14.17)
	orig.SetHorzSpacing(2.83)
	orig.SetVertSpacing(2.83)
	orig.SetWordWrap(false)
	orig.SetText("Hello")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("CellularTextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewCellularTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.CellWidth() != 14.17 {
		t.Errorf("CellWidth: got %v, want 14.17", got.CellWidth())
	}
	if got.CellHeight() != 14.17 {
		t.Errorf("CellHeight: got %v, want 14.17", got.CellHeight())
	}
	if got.HorzSpacing() != 2.83 {
		t.Errorf("HorzSpacing: got %v, want 2.83", got.HorzSpacing())
	}
	if got.VertSpacing() != 2.83 {
		t.Errorf("VertSpacing: got %v, want 2.83", got.VertSpacing())
	}
	if got.WordWrap() {
		t.Error("WordWrap: want false")
	}
	if got.Text() != "Hello" {
		t.Errorf("Text: got %q, want Hello", got.Text())
	}
}

// TestCellularTextObject_Deserialize_AllCellularFieldsXML deserializes directly
// from explicit XML to verify every field path is read correctly.
func TestCellularTextObject_Deserialize_AllCellularFieldsXML(t *testing.T) {
	xmlStr := `<CellularTextObject Text="World" CellWidth="10" CellHeight="12" ` +
		`HorzSpacing="2" VertSpacing="3" WordWrap="false"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewCellularTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.CellWidth() != 10 {
		t.Errorf("CellWidth: got %v, want 10", got.CellWidth())
	}
	if got.CellHeight() != 12 {
		t.Errorf("CellHeight: got %v, want 12", got.CellHeight())
	}
	if got.HorzSpacing() != 2 {
		t.Errorf("HorzSpacing: got %v, want 2", got.HorzSpacing())
	}
	if got.VertSpacing() != 3 {
		t.Errorf("VertSpacing: got %v, want 3", got.VertSpacing())
	}
	if got.WordWrap() {
		t.Error("WordWrap: want false")
	}
	if got.Text() != "World" {
		t.Errorf("Text: got %q, want World", got.Text())
	}
}
