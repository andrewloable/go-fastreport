package object_test

// text_font_coverage_test.go — coverage for TextObject Font serialization and
// deserialization branches:
//   - TextObject.Serialize: if !style.FontEqual(t.font, style.DefaultFont()) { WriteStr("Font",...) }
//   - TextObject.Deserialize: if s := r.ReadStr("Font", ""); s != "" { FontFromStr(s) }

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/style"
)

// TestTextObject_Serialize_NonDefaultFont verifies that when a TextObject has a
// font that differs from the default (Arial 10 Regular), the Font attribute is
// written in the serialized XML.
func TestTextObject_Serialize_NonDefaultFont(t *testing.T) {
	obj := object.NewTextObject()
	// Set a font different from the default Arial 10pt Regular.
	obj.SetFont(style.Font{Name: "Courier New", Size: 12, Style: style.FontStyleBold})

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", obj); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `Font=`) {
		t.Errorf("expected Font attribute in XML:\n%s", xml)
	}
	if !strings.Contains(xml, "Courier New") {
		t.Errorf("expected font name 'Courier New' in XML:\n%s", xml)
	}
}

// TestTextObject_Deserialize_WithFontAttribute verifies that when a TextObject
// is deserialized from XML containing a Font attribute, the font is parsed and
// stored correctly via FontFromStr.
func TestTextObject_Deserialize_WithFontAttribute(t *testing.T) {
	// Build raw XML with a Font attribute in FastReport .NET format.
	rawXML := `<TextObject Name="txt1" Font="Courier New, 12, 1"/>`

	r := serial.NewReader(strings.NewReader(rawXML))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TextObject" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	obj := object.NewTextObject()
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	got := obj.Font()
	if got.Name != "Courier New" {
		t.Errorf("Font.Name: got %q, want %q", got.Name, "Courier New")
	}
	if got.Size != 12 {
		t.Errorf("Font.Size: got %v, want 12", got.Size)
	}
}

// TestTextObject_FontRoundTrip verifies that a TextObject with a non-default font
// can be serialized and deserialized with full round-trip fidelity.
func TestTextObject_FontRoundTrip(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetFont(style.Font{Name: "Times New Roman", Size: 14, Style: style.FontStyleItalic})

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TextObject" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	gotFont := got.Font()
	if gotFont.Name != "Times New Roman" {
		t.Errorf("Font.Name: got %q, want %q", gotFont.Name, "Times New Roman")
	}
	if gotFont.Size != 14 {
		t.Errorf("Font.Size: got %v, want 14", gotFont.Size)
	}
}
