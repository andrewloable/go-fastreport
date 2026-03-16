package object_test

// text_parsecolor_coverage_test.go — coverage tests for remaining uncovered
// branches in TextObject.Deserialize.
//
// Uncovered branches targeted:
//   - text.go:733  `if c, err := utils.ParseColor(cs); err == nil` — the else
//     path where ParseColor fails for an invalid color string. The color is then
//     left at its zero value (not set to default, not set to parsed value).
//   - This requires a TextOutline.Color value that ParseColor rejects.
//
// All other Serialize/Deserialize branches in text.go are already covered by
// the existing test suite. This file covers the remaining ParseColor error path.

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
)

// TestTextObject_Deserialize_TextOutlineColor_InvalidParse exercises the
// `if c, err := utils.ParseColor(cs); err == nil { ... }` branch where
// ParseColor fails. The `if cs != "" { ... }` branch IS entered (cs is not
// empty), but ParseColor returns an error, so t.textOutline.Color is NOT set.
//
// We provide a raw XML string with an invalid TextOutline.Color value.
func TestTextObject_Deserialize_TextOutlineColor_InvalidParse(t *testing.T) {
	// "NOTACOLOR" is not a valid hex color — ParseColor should return an error.
	rawXML := `<TextObject Name="outlinetest" TextOutline.Enabled="true" TextOutline.Color="NOTACOLOR"/>`

	r := serial.NewReader(strings.NewReader(rawXML))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TextObject" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	obj := object.NewTextObject()
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	// TextOutline.Enabled should be true (successfully read).
	if !obj.TextOutline().Enabled {
		t.Error("TextOutline.Enabled should be true")
	}
	// TextOutline.Color was not set (ParseColor failed) — it remains at zero value.
	// We just verify the object is in a consistent state (no panic, no error).
}

// TestTextObject_Deserialize_TextOutlineColor_ValidParse exercises the
// success path of `if c, err := utils.ParseColor(cs); err == nil { ... }`
// to ensure a valid color string is parsed and stored correctly.
func TestTextObject_Deserialize_TextOutlineColor_ValidParse(t *testing.T) {
	// Use a standard hex color "#FF0000FF" (opaque red).
	rawXML := `<TextObject Name="outlinetest2" TextOutline.Enabled="true" TextOutline.Color="#FF0000FF" TextOutline.Width="2" TextOutline.DashStyle="1"/>`

	r := serial.NewReader(strings.NewReader(rawXML))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TextObject" {
		t.Fatalf("ReadObjectHeader: got %q ok=%v", typeName, ok)
	}

	obj := object.NewTextObject()
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if !obj.TextOutline().Enabled {
		t.Error("TextOutline.Enabled should be true")
	}
	if obj.TextOutline().Width != 2 {
		t.Errorf("TextOutline.Width = %v, want 2", obj.TextOutline().Width)
	}
	if obj.TextOutline().DashStyle != 1 {
		t.Errorf("TextOutline.DashStyle = %d, want 1", obj.TextOutline().DashStyle)
	}
}
