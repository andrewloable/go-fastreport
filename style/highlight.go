package style

import "image/color"

// HighlightCondition holds one conditional-formatting rule for a TextObject.
// When the Expression evaluates to true, the associated visual overrides are
// applied to the rendered object. Only properties flagged with Apply* are used.
//
// It is the Go equivalent of FastReport.HighlightCondition / FastReport.StyleBase.
type HighlightCondition struct {
	// Expression is the boolean expression that enables this condition
	// (e.g. "[Row#] % 2 == 0"). Evaluated via Report.Calc().
	Expression string

	// Visible controls object visibility when the condition is true.
	// Default true (matching FastReport.HighlightCondition defaults).
	Visible bool

	// Apply flags mirror FastReport StyleBase.Apply* properties.
	ApplyBorder   bool
	ApplyFill     bool
	ApplyFont     bool
	ApplyTextFill bool

	// Visual overrides — applied only when the corresponding Apply flag is true.
	FillColor     color.RGBA // background fill override (solid fill colour)
	TextFillColor color.RGBA // text (foreground) colour override
	Font          Font       // font override
	// Border is the border override applied when ApplyBorder is true.
	// Mirrors C# StyleBase.Border (StyleBase.cs) used by HighlightCondition.
	Border Border
}

// NewHighlightCondition returns a HighlightCondition with the same defaults
// as FastReport.HighlightCondition: Visible=true, ApplyTextFill=true, text
// fill colour = red.
func NewHighlightCondition() HighlightCondition {
	return HighlightCondition{
		Visible:       true,
		ApplyTextFill: true,
		TextFillColor: color.RGBA{R: 255, A: 255},
		Border:        *NewBorder(),
	}
}

// Assign copies all fields from src into this HighlightCondition.
// Mirrors C# HighlightCondition.Assign (HighlightCondition.cs).
func (h *HighlightCondition) Assign(src HighlightCondition) {
	h.Expression = src.Expression
	h.Visible = src.Visible
	h.ApplyBorder = src.ApplyBorder
	h.ApplyFill = src.ApplyFill
	h.ApplyFont = src.ApplyFont
	h.ApplyTextFill = src.ApplyTextFill
	h.FillColor = src.FillColor
	h.TextFillColor = src.TextFillColor
	h.Font = src.Font
	// Deep-copy Border lines to avoid sharing pointers.
	h.Border = src.Border
	for i, l := range src.Border.Lines {
		if l != nil {
			cp := *l
			h.Border.Lines[i] = &cp
		}
	}
}

// Clone returns a copy of this HighlightCondition.
// Mirrors C# HighlightCondition.Clone (HighlightCondition.cs line 75-80).
func (h HighlightCondition) Clone() HighlightCondition {
	c := h
	// Deep-copy Border lines to avoid sharing pointers.
	for i, l := range h.Border.Lines {
		if l != nil {
			cp := *l
			c.Border.Lines[i] = &cp
		}
	}
	return c
}

// Equals reports whether h and other have identical field values.
// Mirrors C# HighlightCondition.Equals (HighlightCondition.cs line 83-90).
func (h HighlightCondition) Equals(other HighlightCondition) bool {
	return h.Expression == other.Expression &&
		h.Visible == other.Visible &&
		h.ApplyBorder == other.ApplyBorder &&
		h.ApplyFill == other.ApplyFill &&
		h.ApplyFont == other.ApplyFont &&
		h.ApplyTextFill == other.ApplyTextFill &&
		h.FillColor == other.FillColor &&
		h.TextFillColor == other.TextFillColor &&
		h.Font == other.Font
}
