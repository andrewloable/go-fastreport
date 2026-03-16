package object_test

// container_coverage_test.go — additional coverage tests for container.go
// Targets remaining uncovered branches in CheckBoxObject, ContainerObject,
// and SubreportObject Serialize/Deserialize methods.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── CheckBoxObject: Serialize with all non-default values ────────────────────

func TestCheckBoxObject_Serialize_AllNonDefaults(t *testing.T) {
	orig := object.NewCheckBoxObject()
	orig.SetChecked(true)
	orig.SetCheckedSymbol(object.CheckedSymbolCross)
	orig.SetUncheckedSymbol(object.UncheckedSymbolMinus)
	orig.SetDataColumn("Active")
	orig.SetExpression("[Field] > 0")
	orig.SetCheckWidthRatio(0.75)
	orig.SetHideIfUnchecked(true)
	orig.SetEditable(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("CheckBoxObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	for _, attr := range []string{"Checked=", "CheckedSymbol=", "UncheckedSymbol=",
		"DataColumn=", "Expression=", "CheckWidthRatio=", "HideIfUnchecked=", "Editable="} {
		if !strings.Contains(xml, attr) {
			t.Errorf("expected %q in XML:\n%s", attr, xml)
		}
	}
}

// ── CheckBoxObject: round-trip ────────────────────────────────────────────────

func TestCheckBoxObject_RoundTrip_AllFields(t *testing.T) {
	orig := object.NewCheckBoxObject()
	orig.SetChecked(true)
	orig.SetCheckedSymbol(object.CheckedSymbolPlus)
	orig.SetUncheckedSymbol(object.UncheckedSymbolSlash)
	orig.SetDataColumn("IsActive")
	orig.SetExpression("[Amount] > 0")
	orig.SetCheckWidthRatio(0.6)
	orig.SetHideIfUnchecked(true)
	orig.SetEditable(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("CheckBoxObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewCheckBoxObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if !got.Checked() {
		t.Error("Checked: want true")
	}
	if got.CheckedSymbol() != object.CheckedSymbolPlus {
		t.Errorf("CheckedSymbol: got %d, want CheckedSymbolPlus", got.CheckedSymbol())
	}
	if got.UncheckedSymbol() != object.UncheckedSymbolSlash {
		t.Errorf("UncheckedSymbol: got %d, want UncheckedSymbolSlash", got.UncheckedSymbol())
	}
	if got.DataColumn() != "IsActive" {
		t.Errorf("DataColumn: got %q, want IsActive", got.DataColumn())
	}
	if got.Expression() != "[Amount] > 0" {
		t.Errorf("Expression: got %q", got.Expression())
	}
	if got.CheckWidthRatio() != 0.6 {
		t.Errorf("CheckWidthRatio: got %v, want 0.6", got.CheckWidthRatio())
	}
	if !got.HideIfUnchecked() {
		t.Error("HideIfUnchecked: want true")
	}
	if !got.Editable() {
		t.Error("Editable: want true")
	}
}

// ── CheckBoxObject: Deserialize with explicit XML ────────────────────────────

func TestCheckBoxObject_Deserialize_ExplicitXML(t *testing.T) {
	xmlStr := `<CheckBoxObject Checked="true" CheckedSymbol="2" UncheckedSymbol="3" ` +
		`DataColumn="Col1" Expression="[X]>0" CheckWidthRatio="0.5" HideIfUnchecked="true" Editable="true"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewCheckBoxObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !got.Checked() {
		t.Error("Checked: want true")
	}
	if got.CheckedSymbol() != object.CheckedSymbolPlus {
		t.Errorf("CheckedSymbol: got %d, want 2 (Plus)", got.CheckedSymbol())
	}
	if got.UncheckedSymbol() != object.UncheckedSymbolSlash {
		t.Errorf("UncheckedSymbol: got %d, want 3 (Slash)", got.UncheckedSymbol())
	}
	if got.DataColumn() != "Col1" {
		t.Errorf("DataColumn: got %q, want Col1", got.DataColumn())
	}
	if got.Expression() != "[X]>0" {
		t.Errorf("Expression: got %q, want [X]>0", got.Expression())
	}
	if got.CheckWidthRatio() != 0.5 {
		t.Errorf("CheckWidthRatio: got %v, want 0.5", got.CheckWidthRatio())
	}
	if !got.HideIfUnchecked() {
		t.Error("HideIfUnchecked: want true")
	}
	if !got.Editable() {
		t.Error("Editable: want true")
	}
}

// ── ContainerObject: Serialize with child objects and both event names ────────

func TestContainerObject_Serialize_WithChildrenAndBothEvents(t *testing.T) {
	c := object.NewContainerObject()
	c.SetBeforeLayoutEvent("BeforeEvt")
	c.SetAfterLayoutEvent("AfterEvt")

	// Add a child object so the children loop executes.
	child := object.NewBarcodeObject()
	child.SetText("child1")
	c.AddChild(child)

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
	if !strings.Contains(xml, `BarcodeObject`) {
		t.Errorf("expected child BarcodeObject in XML:\n%s", xml)
	}
}

// ── ContainerObject: Deserialize with both event names ───────────────────────

func TestContainerObject_Deserialize_BothEvents(t *testing.T) {
	xmlStr := `<ContainerObject BeforeLayoutEvent="BeforeEvt" AfterLayoutEvent="AfterEvt"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewContainerObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.BeforeLayoutEvent() != "BeforeEvt" {
		t.Errorf("BeforeLayoutEvent: got %q, want BeforeEvt", got.BeforeLayoutEvent())
	}
	if got.AfterLayoutEvent() != "AfterEvt" {
		t.Errorf("AfterLayoutEvent: got %q, want AfterEvt", got.AfterLayoutEvent())
	}
}

// ── ContainerObject: Serialize with only AfterLayoutEvent set ─────────────────

func TestContainerObject_Serialize_OnlyAfterEvent(t *testing.T) {
	c := object.NewContainerObject()
	c.SetAfterLayoutEvent("AfterOnly")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("ContainerObject", c); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `AfterLayoutEvent=`) {
		t.Errorf("expected AfterLayoutEvent in XML:\n%s", xml)
	}
	if strings.Contains(xml, `BeforeLayoutEvent=`) {
		t.Errorf("unexpected BeforeLayoutEvent in XML:\n%s", xml)
	}
}

// ── SubreportObject: Serialize with only reportPageName (no PrintOnParent) ────

func TestSubreportObject_Serialize_OnlyReportPage(t *testing.T) {
	orig := object.NewSubreportObject()
	orig.SetReportPageName("Page1")
	// printOnParent stays false (default)

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
	if strings.Contains(xml, `PrintOnParent=`) {
		t.Errorf("unexpected PrintOnParent in XML:\n%s", xml)
	}
}

// ── SubreportObject: Deserialize with ReportPage and PrintOnParent=false ──────

func TestSubreportObject_Deserialize_ExplicitFalse(t *testing.T) {
	xmlStr := `<SubreportObject ReportPage="MyPage" PrintOnParent="false"/>`

	r := serial.NewReader(strings.NewReader(xmlStr))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewSubreportObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ReportPageName() != "MyPage" {
		t.Errorf("ReportPageName: got %q, want MyPage", got.ReportPageName())
	}
	if got.PrintOnParent() {
		t.Error("PrintOnParent: want false")
	}
}

// ── ContainerObject: UpdateLayout coverage ────────────────────────────────────

func TestContainerObject_UpdateLayout_Called(t *testing.T) {
	c := object.NewContainerObject()
	// Call UpdateLayout with various values to hit the empty-body function.
	c.UpdateLayout(0, 0)
	c.UpdateLayout(100, 200)
	c.UpdateLayout(-1, -1)
}
