package reportpkg

import (
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
)

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
	// Fill color.
	transparent := [4]byte{}
	if [4]byte{e.FillColor.R, e.FillColor.G, e.FillColor.B, e.FillColor.A} != transparent {
		w.WriteStr("Fill.Color", utils.FormatColor(e.FillColor))
	}
	// Text fill color.
	if [4]byte{e.TextColor.R, e.TextColor.G, e.TextColor.B, e.TextColor.A} != transparent {
		w.WriteStr("TextFill.Color", utils.FormatColor(e.TextColor))
	}
	// Font.
	if e.FontChanged && e.Font.Name != "" {
		w.WriteStr("Font", style.FontToStr(e.Font))
	}
	// Border color (shared for all lines).
	if [4]byte{e.BorderColor.R, e.BorderColor.G, e.BorderColor.B, e.BorderColor.A} != transparent {
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
