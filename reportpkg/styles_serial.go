package reportpkg

import (
	"image/color"

	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
)

// zeroRGBA is the zero value for color.RGBA (fully transparent black).
var zeroRGBA = color.RGBA{}

// stylesSerializer wraps a StyleSheet so it can be written as a <Styles>
// child element via report.Writer.WriteObject.
type stylesSerializer struct {
	ss *style.StyleSheet
}

// TypeName implements serial.TypeNamer; returns "Styles" as the XML element name.
func (s *stylesSerializer) TypeName() string { return "Styles" }

// Serialize writes each StyleEntry as a <Style> child element.
func (s *stylesSerializer) Serialize(w report.Writer) error {
	for _, e := range s.ss.All() {
		if err := w.WriteObject(&styleEntrySerializer{e}); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize is a no-op (loading handled separately in loadsave.go).
func (s *stylesSerializer) Deserialize(_ report.Reader) error { return nil }

// styleEntrySerializer wraps a single StyleEntry for serialization.
type styleEntrySerializer struct {
	e *style.StyleEntry
}

// TypeName returns "Style" as the XML element name.
func (s *styleEntrySerializer) TypeName() string { return "Style" }

// Serialize writes StyleEntry attributes that differ from defaults.
func (s *styleEntrySerializer) Serialize(w report.Writer) error {
	e := s.e
	if e.Name != "" {
		w.WriteStr("Name", e.Name)
	}
	// Apply flags — only write when false (default is true).
	if !e.ApplyBorder {
		w.WriteBool("ApplyBorder", false)
	}
	if !e.ApplyFill {
		w.WriteBool("ApplyFill", false)
	}
	if !e.ApplyTextFill {
		w.WriteBool("ApplyTextFill", false)
	}
	if !e.ApplyFont {
		w.WriteBool("ApplyFont", false)
	}
	// Fill: use SerializeFill so gradient/hatch fills in StyleEntry are
	// emitted correctly, matching C# StyleBase.Serialize (StyleBase.cs line 158).
	// C# always writes the fill regardless of the ApplyFill flag; the flag
	// controls application, not serialization. Prefer the rich Fill interface
	// field when set; fall back to the legacy FillColor scalar.
	var fillToWrite style.Fill
	if e.Fill != nil {
		fillToWrite = e.Fill
	} else if e.FillColor != (zeroRGBA) {
		fillToWrite = &style.SolidFill{Color: e.FillColor}
	}
	report.SerializeFill(w, "Fill", fillToWrite)
	// TextFill: same pattern for text colour.
	var textFillToWrite style.Fill
	if e.TextFill != nil {
		textFillToWrite = e.TextFill
	} else if e.TextColor != (zeroRGBA) {
		textFillToWrite = &style.SolidFill{Color: e.TextColor}
	}
	report.SerializeFill(w, "TextFill", textFillToWrite)
	// Font.
	if e.FontChanged && e.Font.Name != "" {
		w.WriteStr("Font", style.FontToStr(e.Font))
	}
	// Border color (shared for all lines).
	var zeroColor [4]byte
	if [4]byte{e.BorderColor.R, e.BorderColor.G, e.BorderColor.B, e.BorderColor.A} != zeroColor {
		w.WriteStr("Border.Color", utils.FormatColor(e.BorderColor))
	}
	// Border visible lines.
	if e.Border.VisibleLines != style.BorderLinesNone {
		w.WriteStr("Border.Lines", formatBorderLinesLocal(e.Border.VisibleLines))
	}
	// Border shadow.
	if e.Border.Shadow {
		w.WriteBool("Border.Shadow", true)
	}
	return nil
}

// Deserialize is a no-op.
func (s *styleEntrySerializer) Deserialize(_ report.Reader) error { return nil }

// formatBorderLinesLocal is a local copy of the border-lines formatter
// (avoids importing the report package from reportpkg, which would cycle).
func formatBorderLinesLocal(bl style.BorderLines) string {
	if bl == style.BorderLinesNone {
		return "None"
	}
	if bl == style.BorderLinesAll {
		return "All"
	}
	parts := make([]string, 0, 4)
	if bl&style.BorderLinesLeft != 0 {
		parts = append(parts, "Left")
	}
	if bl&style.BorderLinesRight != 0 {
		parts = append(parts, "Right")
	}
	if bl&style.BorderLinesTop != 0 {
		parts = append(parts, "Top")
	}
	if bl&style.BorderLinesBottom != 0 {
		parts = append(parts, "Bottom")
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
