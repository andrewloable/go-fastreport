package object

import "github.com/andrewloable/go-fastreport/report"

// RichObject renders Rich Text Format (RTF) content. The RTF text may
// contain bracket-expressions like [DataSource.Field] that are evaluated
// at report generation time.
//
// It is the Go equivalent of FastReport.RichObject.
type RichObject struct {
	report.ReportComponentBase

	// text holds the RTF content (may include [bracket] expressions).
	text string
	// canGrow allows the object to expand vertically to fit its content.
	canGrow bool
}

// NewRichObject creates a RichObject with default settings.
func NewRichObject() *RichObject {
	return &RichObject{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (r *RichObject) BaseName() string { return "Rich" }

// TypeName returns "RichObject".
func (r *RichObject) TypeName() string { return "RichObject" }

// Text returns the RTF content.
func (r *RichObject) Text() string { return r.text }

// SetText sets the RTF content.
func (r *RichObject) SetText(v string) { r.text = v }

// CanGrow reports whether the object may grow vertically.
func (r *RichObject) CanGrow() bool { return r.canGrow }

// SetCanGrow sets whether the object may grow vertically.
func (r *RichObject) SetCanGrow(v bool) { r.canGrow = v }

// Serialize writes RichObject properties that differ from defaults.
func (r *RichObject) Serialize(w report.Writer) error {
	if err := r.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if r.text != "" {
		w.WriteStr("Text", r.text)
	}
	if r.canGrow {
		w.WriteBool("CanGrow", true)
	}
	return nil
}

// Deserialize reads RichObject properties from an FRX reader.
func (r *RichObject) Deserialize(rd report.Reader) error {
	if err := r.ReportComponentBase.Deserialize(rd); err != nil {
		return err
	}
	r.text = rd.ReadStr("Text", "")
	r.canGrow = rd.ReadBool("CanGrow", false)
	return nil
}
