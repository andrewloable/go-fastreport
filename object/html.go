package object

import (
	"github.com/andrewloable/go-fastreport/report"
)

// HtmlObject is a report component that renders its text content as HTML.
// It extends TextObjectBase and is the Go equivalent of FastReport.HtmlObject.
// During HTML export the text is emitted verbatim as raw HTML; other exporters
// strip HTML tags and render the underlying plain text.
type HtmlObject struct {
	TextObjectBase

	// rightToLeft indicates right-to-left text direction.
	rightToLeft bool
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
