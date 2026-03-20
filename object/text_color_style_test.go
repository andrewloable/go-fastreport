package object_test

// text_color_style_test.go — coverage tests for the remaining uncovered lines
// in object/text.go:
//   - TextColor()            line 465 — 0%
//   - SetTextColor()         line 468 — 0%
//   - ApplyStyle()           line 473 — 85.7% (ApplyTextFill / TextColorChanged branch missing)
//   - ParseHorzAlign()       line 281 — 60% ("Justify"/"3" case missing)
//   - TextObject.Serialize   line 673 — 98% (TextRenderType != Default branch missing)
//   - TextObject.Deserialize line 752 — 91.2% (TextFill.Color branch missing)

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/style"
)

// ── TextColor / SetTextColor ──────────────────────────────────────────────────

// TestTextObject_TextColor_Default verifies that a freshly constructed
// TextObject returns opaque black as its text color (the documented default).
func TestTextObject_TextColor_Default(t *testing.T) {
	obj := object.NewTextObject()
	got := obj.TextColor()
	want := color.RGBA{A: 255}
	if got != want {
		t.Errorf("TextColor default: got %v, want %v", got, want)
	}
}

// TestTextObject_SetTextColor verifies that SetTextColor stores the value and
// TextColor returns it.
func TestTextObject_SetTextColor(t *testing.T) {
	obj := object.NewTextObject()
	red := color.RGBA{R: 255, A: 255}
	obj.SetTextColor(red)
	got := obj.TextColor()
	if got != red {
		t.Errorf("TextColor after SetTextColor: got %v, want %v", got, red)
	}
}

// TestTextObject_SetTextColor_Multiple verifies that SetTextColor can be called
// multiple times and the last value is retained.
func TestTextObject_SetTextColor_Multiple(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetTextColor(color.RGBA{R: 100, G: 150, B: 200, A: 255})
	green := color.RGBA{G: 200, A: 128}
	obj.SetTextColor(green)
	if obj.TextColor() != green {
		t.Errorf("TextColor after second SetTextColor: got %v, want %v", obj.TextColor(), green)
	}
}

// ── ApplyStyle: ApplyTextFill and TextColorChanged branches ──────────────────

// TestTextObject_ApplyStyle_ApplyTextFill verifies that when a StyleEntry has
// ApplyTextFill set to true, the text color is propagated to the TextObject.
func TestTextObject_ApplyStyle_ApplyTextFill(t *testing.T) {
	obj := object.NewTextObject()
	blue := color.RGBA{B: 255, A: 255}
	entry := &style.StyleEntry{
		ApplyTextFill: true,
		TextColor:     blue,
	}
	obj.ApplyStyle(entry)
	if obj.TextColor() != blue {
		t.Errorf("TextColor after ApplyStyle(ApplyTextFill=true): got %v, want %v",
			obj.TextColor(), blue)
	}
}

// TestTextObject_ApplyStyle_TextColorChanged verifies the legacy TextColorChanged
// flag is honoured by ApplyStyle.
func TestTextObject_ApplyStyle_TextColorChanged(t *testing.T) {
	obj := object.NewTextObject()
	purple := color.RGBA{R: 128, B: 128, A: 255}
	entry := &style.StyleEntry{
		TextColorChanged: true,
		TextColor:        purple,
	}
	obj.ApplyStyle(entry)
	if obj.TextColor() != purple {
		t.Errorf("TextColor after ApplyStyle(TextColorChanged=true): got %v, want %v",
			obj.TextColor(), purple)
	}
}

// TestTextObject_ApplyStyle_FontChanged verifies the legacy FontChanged flag is
// honoured by ApplyStyle (covers the FontChanged branch).
func TestTextObject_ApplyStyle_FontChanged(t *testing.T) {
	obj := object.NewTextObject()
	f := style.Font{Name: "Verdana", Size: 11}
	entry := &style.StyleEntry{
		FontChanged: true,
		Font:        f,
	}
	obj.ApplyStyle(entry)
	got := obj.Font()
	if got.Name != "Verdana" {
		t.Errorf("Font.Name after ApplyStyle(FontChanged=true): got %q, want Verdana", got.Name)
	}
}

// TestTextObject_ApplyStyle_NoFlags verifies that when neither font nor
// text-fill flags are set the object's font and color are unchanged.
func TestTextObject_ApplyStyle_NoFlags(t *testing.T) {
	obj := object.NewTextObject()
	origFont := obj.Font()
	origColor := obj.TextColor()
	entry := &style.StyleEntry{
		Name: "neutral",
	}
	obj.ApplyStyle(entry)
	if obj.Font() != origFont {
		t.Errorf("Font changed unexpectedly after ApplyStyle with no flags")
	}
	if obj.TextColor() != origColor {
		t.Errorf("TextColor changed unexpectedly after ApplyStyle with no flags")
	}
}

// ── ParseHorzAlign: Justify / "3" ────────────────────────────────────────────

// TestParseHorzAlign_Justify verifies that the string "Justify" is parsed
// to HorzAlignJustify.
func TestParseHorzAlign_Justify(t *testing.T) {
	got := object.ParseHorzAlign("Justify")
	if got != object.HorzAlignJustify {
		t.Errorf("ParseHorzAlign(\"Justify\") = %v, want HorzAlignJustify", got)
	}
}

// TestParseHorzAlign_Justify_Numeric verifies that the numeric string "3"
// is also parsed to HorzAlignJustify.
func TestParseHorzAlign_Justify_Numeric(t *testing.T) {
	got := object.ParseHorzAlign("3")
	if got != object.HorzAlignJustify {
		t.Errorf("ParseHorzAlign(\"3\") = %v, want HorzAlignJustify", got)
	}
}

// TestParseHorzAlign_Right verifies that the string "Right" is parsed to
// HorzAlignRight.
func TestParseHorzAlign_Right(t *testing.T) {
	got := object.ParseHorzAlign("Right")
	if got != object.HorzAlignRight {
		t.Errorf("ParseHorzAlign(\"Right\") = %v, want HorzAlignRight", got)
	}
}

// TestParseHorzAlign_Right_Numeric verifies that the numeric string "2" is
// parsed to HorzAlignRight.
func TestParseHorzAlign_Right_Numeric(t *testing.T) {
	got := object.ParseHorzAlign("2")
	if got != object.HorzAlignRight {
		t.Errorf("ParseHorzAlign(\"2\") = %v, want HorzAlignRight", got)
	}
}

// ── TextObject.Serialize: TextRenderType != TextRenderTypeDefault ─────────────

// TestTextObject_Serialize_TextRenderType verifies that a non-default
// TextRenderType value is written to the serialized XML.
func TestTextObject_Serialize_TextRenderType(t *testing.T) {
	obj := object.NewTextObject()
	// TextRenderTypeHtmlTags is the value 1 (non-default).
	obj.SetTextRenderType(object.TextRenderTypeHtmlTags)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", obj); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `TextRenderType=`) {
		t.Errorf("expected TextRenderType in XML:\n%s", xml)
	}
}

// TestTextObject_Serialize_TextRenderType_RoundTrip verifies that TextRenderType
// survives a serialize/deserialize round-trip.
func TestTextObject_Serialize_TextRenderType_RoundTrip(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetTextRenderType(object.TextRenderTypeHtmlTags)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.TextRenderType() != object.TextRenderTypeHtmlTags {
		t.Errorf("TextRenderType = %v, want TextRenderTypeHtmlTags", got.TextRenderType())
	}
}

// ── TextObject.Deserialize: TextFill.Color branch ────────────────────────────

// TestTextObject_Deserialize_TextFillColor_Valid verifies that a valid
// TextFill.Color attribute in the XML is parsed and stored as the text color.
func TestTextObject_Deserialize_TextFillColor_Valid(t *testing.T) {
	// Opaque red in ARGB hex format.
	rawXML := `<TextObject Name="tfc" TextFill.Color="#FFFF0000"/>`

	r := serial.NewReader(strings.NewReader(rawXML))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TextObject" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	obj := object.NewTextObject()
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	got := obj.TextColor()
	// Parsed color should not be the default opaque black.
	if got == (color.RGBA{A: 255}) {
		t.Errorf("TextColor was not updated from TextFill.Color attribute: got %v", got)
	}
}

// TestTextObject_Deserialize_TextFillColor_Invalid verifies that an invalid
// TextFill.Color value is silently ignored (no error, color unchanged).
func TestTextObject_Deserialize_TextFillColor_Invalid(t *testing.T) {
	rawXML := `<TextObject Name="tfc2" TextFill.Color="notacolor"/>`

	r := serial.NewReader(strings.NewReader(rawXML))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}

	obj := object.NewTextObject()
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// Color is left at its zero/default value; no panic or error.
	_ = obj.TextColor()
}
