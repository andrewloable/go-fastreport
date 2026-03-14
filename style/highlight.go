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
	FillColor     color.RGBA // background fill override
	TextFillColor color.RGBA // text (foreground) colour override
	Font          Font       // font override
	// Border override is intentionally omitted; it is rarely used in highlight
	// conditions and its serialisation is complex.
}

// NewHighlightCondition returns a HighlightCondition with the same defaults
// as FastReport.HighlightCondition: Visible=true, ApplyTextFill=true, text
// fill colour = red.
func NewHighlightCondition() HighlightCondition {
	return HighlightCondition{
		Visible:       true,
		ApplyTextFill: true,
		TextFillColor: color.RGBA{R: 255, A: 255},
	}
}
