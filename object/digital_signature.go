package object

import "github.com/andrewloable/go-fastreport/report"

// DigitalSignatureObject displays a digital signature placeholder.
// At print time it renders a signature field that can be signed digitally.
//
// It is the Go equivalent of FastReport.DigitalSignature.DigitalSignatureObject.
// This stub supports FRX loading; actual signature rendering is not yet implemented.
type DigitalSignatureObject struct {
	report.ReportComponentBase

	// placeholder is the display text shown before a signature is applied.
	placeholder string
}

// NewDigitalSignatureObject creates a DigitalSignatureObject with defaults.
func NewDigitalSignatureObject() *DigitalSignatureObject {
	return &DigitalSignatureObject{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (d *DigitalSignatureObject) BaseName() string { return "DigitalSignature" }

// TypeName returns "DigitalSignatureObject".
func (d *DigitalSignatureObject) TypeName() string { return "DigitalSignatureObject" }

// Placeholder returns the placeholder text.
func (d *DigitalSignatureObject) Placeholder() string { return d.placeholder }

// SetPlaceholder sets the placeholder text.
func (d *DigitalSignatureObject) SetPlaceholder(v string) { d.placeholder = v }

// Serialize writes DigitalSignatureObject properties that differ from defaults.
func (d *DigitalSignatureObject) Serialize(w report.Writer) error {
	if err := d.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if d.placeholder != "" {
		w.WriteStr("Placeholder", d.placeholder)
	}
	return nil
}

// Deserialize reads DigitalSignatureObject properties from an FRX reader.
func (d *DigitalSignatureObject) Deserialize(r report.Reader) error {
	if err := d.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	d.placeholder = r.ReadStr("Placeholder", "")
	return nil
}
