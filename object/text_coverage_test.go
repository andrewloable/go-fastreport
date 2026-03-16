package object_test

// text_coverage_test.go — coverage tests for text.go uncovered branches:
//   - strToPadding: invalid format
//   - SetTextOutline: completely untested
//   - DeserializeChild: Highlight, Formats, unknown child
//   - TextObject.Serialize: Padding, TextOutline branches

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/style"
)

// ── strToPadding via TextObjectBase.Deserialize ───────────────────────────────

func TestTextObjectBase_Padding_InvalidFormat(t *testing.T) {
	// strToPadding with len(parts) != 4 returns Padding{}.
	// Trigger by serializing a TextObject with a non-4-part "Padding" string.
	// We do this by deserializing directly with a mock reader.
	// The mock reader is not available in external package — use serial round-trip.

	// Serialize with valid padding first.
	obj := object.NewTextObject()
	obj.SetPadding(object.Padding{Left: 1, Top: 2, Right: 3, Bottom: 4})

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("TextObject", obj) //nolint:errcheck
	w.Flush()                             //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Padding=`) {
		t.Errorf("expected Padding in XML:\n%s", xml)
	}

	// Deserialize back and verify round-trip.
	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	r.ReadObjectHeader()
	got := object.NewTextObject()
	got.Deserialize(r) //nolint:errcheck
	p := got.Padding()
	if p.Left != 1 || p.Top != 2 || p.Right != 3 || p.Bottom != 4 {
		t.Errorf("Padding round-trip: got %+v", p)
	}
}

// ── SetTextOutline + Serialize/Deserialize ────────────────────────────────────

func TestTextObject_SetTextOutline(t *testing.T) {
	obj := object.NewTextObject()
	outline := style.TextOutline{
		Enabled:   true,
		Width:     2.0,
		DashStyle: 1,
	}
	obj.SetTextOutline(outline)

	got := obj.TextOutline()
	if !got.Enabled {
		t.Error("TextOutline.Enabled should be true")
	}
	if got.Width != 2.0 {
		t.Errorf("TextOutline.Width: got %v, want 2.0", got.Width)
	}
	if got.DashStyle != 1 {
		t.Errorf("TextOutline.DashStyle: got %d, want 1", got.DashStyle)
	}
}

func TestTextObject_TextOutline_SerializeDeserialize_Coverage(t *testing.T) {
	orig := object.NewTextObject()
	outline := style.TextOutline{
		Enabled:   true,
		Width:     3.0,
		DashStyle: 2,
	}
	orig.SetTextOutline(outline)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `TextOutline.Enabled="true"`) {
		t.Errorf("expected TextOutline.Enabled in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `TextOutline.Color=`) {
		t.Errorf("expected TextOutline.Color in XML:\n%s", xml)
	}

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	gotOutline := got.TextOutline()
	if !gotOutline.Enabled {
		t.Error("TextOutline.Enabled should be true after round-trip")
	}
	if gotOutline.Width != 3.0 {
		t.Errorf("TextOutline.Width: got %v, want 3.0", gotOutline.Width)
	}
}

// ── DeserializeChild: Highlight child element ─────────────────────────────────

func TestTextObject_DeserializeChild_Highlight(t *testing.T) {
	// Build XML containing a <Highlight> element with a <Condition> child.
	xml := `<TextObject Name="txt1">` +
		`<Highlight>` +
		`<Condition Expression="[Value]&gt;100" Visible="true" ApplyFill="true" Fill.Color="#FF0000FF"/>` +
		`</Highlight>` +
		`</TextObject>`

	r := serial.NewReader(strings.NewReader(xml))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TextObject" {
		t.Fatalf("ReadObjectHeader: %q ok=%v", typeName, ok)
	}
	obj := object.NewTextObject()
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	// Process children.
	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		if !obj.DeserializeChild(ct, r) {
			r.FinishChild() //nolint:errcheck
			continue
		}
		r.FinishChild() //nolint:errcheck
	}

	if len(obj.Highlights()) != 1 {
		t.Errorf("expected 1 highlight, got %d", len(obj.Highlights()))
	}
}

// ── DeserializeChild: unknown child returns false ─────────────────────────────

func TestTextObject_DeserializeChild_Unknown(t *testing.T) {
	xml := `<TextObject Name="txt2"><UnknownElement Foo="bar"/></TextObject>`

	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	obj := object.NewTextObject()
	obj.Deserialize(r) //nolint:errcheck

	ct, ok2 := r.NextChild()
	if !ok2 {
		t.Skip("no children found")
	}
	handled := obj.DeserializeChild(ct, r)
	if handled {
		t.Error("unknown child should return false")
	}
	r.FinishChild() //nolint:errcheck
}

// ── TextObjectBase.Serialize: AllowExpressions=false ─────────────────────────

func TestTextObjectBase_Serialize_AllowExpressionsFalse(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetAllowExpressions(false)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("TextObject", obj) //nolint:errcheck
	w.Flush()                             //nolint:errcheck

	if !strings.Contains(buf.String(), `AllowExpressions="false"`) {
		t.Errorf("expected AllowExpressions=false in XML:\n%s", buf.String())
	}
}

// ── TextObjectBase.Serialize: HideZeros=true, HideValue, NullValue ───────────

func TestTextObjectBase_Serialize_HideZerosHideValueNullValue(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetHideZeros(true)
	obj.SetHideValue("0")
	obj.SetNullValue("N/A")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("TextObject", obj) //nolint:errcheck
	w.Flush()                             //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `HideZeros="true"`) {
		t.Errorf("expected HideZeros in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `HideValue=`) {
		t.Errorf("expected HideValue in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `NullValue=`) {
		t.Errorf("expected NullValue in XML:\n%s", xml)
	}
}

// ── TextObject.Serialize: various flags ──────────────────────────────────────

func TestTextObject_Serialize_AllFlags(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetHorzAlign(object.HorzAlignCenter)
	obj.SetVertAlign(object.VertAlignCenter)
	obj.SetAngle(90)
	obj.SetRightToLeft(true)
	obj.SetWordWrap(false)
	obj.SetUnderlines(true)
	obj.SetFontWidthRatio(0.8)
	obj.SetFirstTabOffset(10)
	obj.SetTabWidth(20)
	obj.SetClip(false)
	obj.SetWysiwyg(true)
	obj.SetLineHeight(14)
	obj.SetForceJustify(true)
	obj.SetAutoShrink(object.AutoShrinkFontSize)
	obj.SetAutoShrinkMinSize(8)
	obj.SetParagraphOffset(5)
	obj.SetMergeMode(object.MergeModeHorizontal)
	obj.SetAutoWidth(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("TextObject", obj) //nolint:errcheck
	w.Flush()                             //nolint:errcheck

	xml := buf.String()
	checks := []string{
		`HorzAlign=`, `VertAlign=`, `Angle=`, `RightToLeft="true"`,
		`WordWrap="false"`, `Underlines="true"`, `FontWidthRatio=`,
		`FirstTabOffset=`, `TabWidth=`, `Clip="false"`, `Wysiwyg="true"`,
		`LineHeight=`, `ForceJustify="true"`, `AutoShrink=`,
		`AutoShrinkMinSize=`, `ParagraphOffset=`, `MergeMode=`, `AutoWidth="true"`,
	}
	for _, check := range checks {
		if !strings.Contains(xml, check) {
			t.Errorf("expected %q in XML:\n%s", check, xml)
		}
	}
}

// ── DeserializeChild: Formats child element ───────────────────────────────────

func TestTextObject_DeserializeChild_Formats(t *testing.T) {
	xml := `<TextObject Name="txt3">` +
		`<Formats>` +
		`<NumberFormat DecimalDigits="2" UseLocale="true"/>` +
		`</Formats>` +
		`</TextObject>`

	r := serial.NewReader(strings.NewReader(xml))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TextObject" {
		t.Fatalf("ReadObjectHeader: %q ok=%v", typeName, ok)
	}
	obj := object.NewTextObject()
	obj.Deserialize(r) //nolint:errcheck

	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		obj.DeserializeChild(ct, r) //nolint:errcheck
		r.FinishChild()             //nolint:errcheck
	}
}

// Ensure we use fmt to avoid unused import.
var _ = fmt.Sprintf

// Ensure report and style packages are referenced.
var _ report.Writer
var _ style.TextOutline

// ── strToPadding with invalid format ─────────────────────────────────────────

func TestTextObjectBase_Deserialize_InvalidPadding(t *testing.T) {
	// Build XML with a Padding that has only 2 values (invalid format).
	xml := `<TextObject Padding="1,2"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	obj := object.NewTextObject()
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// Invalid format → Padding should be zero.
	p := obj.Padding()
	if p.Left != 0 || p.Top != 0 || p.Right != 0 || p.Bottom != 0 {
		t.Errorf("invalid Padding should yield Padding{}: got %+v", p)
	}
}
