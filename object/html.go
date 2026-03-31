package object

import (
	"strings"

	"github.com/andrewloable/go-fastreport/expr"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
)

// HtmlObject is a report component that renders its text content as HTML.
// It extends TextObjectBase and is the Go equivalent of FastReport.HtmlObject.
// During HTML export the text is emitted verbatim as raw HTML; other exporters
// strip HTML tags and render the underlying plain text.
type HtmlObject struct {
	TextObjectBase

	// rightToLeft indicates right-to-left text direction.
	rightToLeft bool

	// savedText holds the pre-engine-pass text so RestoreState can undo
	// expression evaluation. Mirrors C# HtmlObject.savedText field
	// (HtmlObject.cs line 29).
	savedText string
}

// NewHtmlObject creates an HtmlObject with defaults.
func NewHtmlObject() *HtmlObject {
	return &HtmlObject{
		TextObjectBase: *NewTextObjectBase(),
	}
}

// RightToLeft returns whether text is rendered right-to-left.
func (h *HtmlObject) RightToLeft() bool { return h.rightToLeft }

// SetRightToLeft sets the right-to-left flag.
func (h *HtmlObject) SetRightToLeft(v bool) { h.rightToLeft = v }

// Assign copies all HtmlObject fields from src.
// Mirrors C# HtmlObject.Assign(Base source) (HtmlObject.cs lines 80-86).
func (h *HtmlObject) Assign(src *HtmlObject) {
	if src == nil {
		return
	}
	h.BreakableComponent.Assign(&src.BreakableComponent)
	// Copy TextObjectBase value fields (format, padding, text, etc.).
	h.TextObjectBase = src.TextObjectBase
	h.rightToLeft = src.rightToLeft
}

// GetExpressions returns the list of expression strings embedded in the Text
// field (delimited by Brackets), merged with the base component expressions
// (VisibleExpression, PrintableExpression).
// Mirrors C# HtmlObject.GetExpressions() (HtmlObject.cs lines 161-172).
func (h *HtmlObject) GetExpressions() []string {
	// Start with base expressions (VisibleExpression, PrintableExpression, etc.).
	expressions := h.BreakableComponent.GetExpressions()

	if h.allowExpressions && h.brackets != "" {
		parts := strings.SplitN(h.brackets, ",", 2)
		if len(parts) == 2 {
			open, close := parts[0], parts[1]
			for _, tok := range expr.ParseWithBrackets(h.text, open, close) {
				if tok.IsExpr {
					expressions = append(expressions, tok.Value)
				}
			}
		}
	}
	return expressions
}

// SaveState saves the current component state (via the embedded base) and also
// captures the current Text value. Mirrors C# HtmlObject.SaveState()
// (HtmlObject.cs lines 177-180).
func (h *HtmlObject) SaveState() {
	h.BreakableComponent.SaveState()
	h.savedText = h.text
}

// RestoreState restores the component state and the saved Text value.
// Mirrors C# HtmlObject.RestoreState() (HtmlObject.cs lines 183-187).
func (h *HtmlObject) RestoreState() {
	h.BreakableComponent.RestoreState()
	h.text = h.savedText
}

// CalcWidth returns the component's current width, matching the C# stub
// InternalCalcWidth() which simply returns this.Width.
// Mirrors C# HtmlObject.CalcWidth() (HtmlObject.cs lines 193-196).
func (h *HtmlObject) CalcWidth() float32 {
	return h.Width()
}

// ApplyCondition applies the visual overrides from a HighlightCondition.
// Fill (background colour), Border, and Visible flags are supported.
// Mirrors C# HtmlObject.ApplyCondition() (HtmlObject.cs lines 147-155).
func (h *HtmlObject) ApplyCondition(c style.HighlightCondition) {
	if c.ApplyBorder {
		cloned := c.Border.Clone()
		h.SetBorder(*cloned)
	}
	if c.ApplyFill && c.Fill != nil {
		h.SetFill(c.Fill.Clone())
	}
	if !c.Visible {
		h.SetVisible(false)
	}
}

// Serialize writes HtmlObject properties that differ from defaults.
func (h *HtmlObject) Serialize(w report.Writer) error {
	if err := h.TextObjectBase.Serialize(w); err != nil {
		return err
	}
	if h.rightToLeft {
		w.WriteBool("RightToLeft", true)
	}
	return nil
}

// Deserialize reads HtmlObject properties.
func (h *HtmlObject) Deserialize(r report.Reader) error {
	if err := h.TextObjectBase.Deserialize(r); err != nil {
		return err
	}
	h.rightToLeft = r.ReadBool("RightToLeft", false)
	return nil
}
