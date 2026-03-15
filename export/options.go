package export

import "github.com/andrewloable/go-fastreport/report"

// ExportFormat identifies a known output format by name (case-insensitive).
// Examples: "PDF", "HTML", "Image".
type ExportFormat string

const (
	ExportFormatPDF   ExportFormat = "PDF"
	ExportFormatHTML  ExportFormat = "HTML"
	ExportFormatImage ExportFormat = "Image"
)

// ExportsOptions holds report-level export configuration.
// It is the Go equivalent of FastReport.Utils.ExportsOptions.
//
// AllowedExports and HideExports control which formats appear in the UI;
// DefaultFormat selects the pre-selected format in a dialog.
// These fields are advisory for host applications that present an export UI —
// the programmatic export API does not enforce them.
type ExportsOptions struct {
	// AllowedExports lists the export formats the user is allowed to choose.
	// An empty slice means all formats are allowed.
	AllowedExports []ExportFormat

	// HideExports lists export formats that should be hidden from the UI even
	// if they are technically available.
	HideExports []ExportFormat

	// DefaultFormat is the pre-selected format when the user opens the export dialog.
	DefaultFormat ExportFormat

	// ShowProgress controls whether a progress indicator is shown during export.
	ShowProgress bool

	// OpenAfterExport controls whether the exported file is opened automatically.
	OpenAfterExport bool
}

// NewExportsOptions returns ExportsOptions with sensible defaults.
func NewExportsOptions() *ExportsOptions {
	return &ExportsOptions{
		DefaultFormat:   ExportFormatPDF,
		ShowProgress:    true,
		OpenAfterExport: false,
	}
}

// IsAllowed returns true if format is in the AllowedExports list (or the list is empty).
func (o *ExportsOptions) IsAllowed(format ExportFormat) bool {
	if len(o.AllowedExports) == 0 {
		return true
	}
	for _, f := range o.AllowedExports {
		if f == format {
			return true
		}
	}
	return false
}

// IsHidden returns true if format is in the HideExports list.
func (o *ExportsOptions) IsHidden(format ExportFormat) bool {
	for _, f := range o.HideExports {
		if f == format {
			return true
		}
	}
	return false
}

// Serialize writes ExportsOptions properties that differ from defaults.
func (o *ExportsOptions) Serialize(w report.Writer) {
	if o.DefaultFormat != "" && o.DefaultFormat != ExportFormatPDF {
		w.WriteStr("ExportsOptions.DefaultFormat", string(o.DefaultFormat))
	}
	if !o.ShowProgress {
		w.WriteBool("ExportsOptions.ShowProgress", false)
	}
	if o.OpenAfterExport {
		w.WriteBool("ExportsOptions.OpenAfterExport", true)
	}
}

// Deserialize reads ExportsOptions properties.
func (o *ExportsOptions) Deserialize(r report.Reader) {
	if s := r.ReadStr("ExportsOptions.DefaultFormat", ""); s != "" {
		o.DefaultFormat = ExportFormat(s)
	}
	o.ShowProgress = r.ReadBool("ExportsOptions.ShowProgress", true)
	o.OpenAfterExport = r.ReadBool("ExportsOptions.OpenAfterExport", false)
}
