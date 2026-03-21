package reportpkg

// DefaultPaperSize specifies the default paper size used when creating new report pages.
// C# FastReport.ReportSettings.DefaultPaperSize (ReportSettings.cs).
type DefaultPaperSize int

const (
	PaperA4     DefaultPaperSize = iota // A4 paper (210 × 297 mm)
	PaperLetter                         // US Letter (8.5 × 11 inches)
)

// String returns a human-readable name for the paper size.
func (d DefaultPaperSize) String() string {
	switch d {
	case PaperA4:
		return "A4"
	case PaperLetter:
		return "Letter"
	default:
		return "A4"
	}
}

// ReportSettings contains global runtime settings for a Report.
// Ported from C# FastReport.ReportSettings (ReportSettings.cs).
// Designer-only features (ShowProgress, OnProgress, DefaultLanguage for
// CodeDom compilation) are omitted.
type ReportSettings struct {
	// DefaultPaperSize is the paper format applied to newly created pages.
	DefaultPaperSize DefaultPaperSize

	// ImageLocationRoot is the base path prepended to PictureObject.ImageLocation
	// when resolving relative image paths at runtime.
	ImageLocationRoot string

	// UsePropValuesToDiscoverBO controls whether the engine inspects property
	// values (not just types) when auto-discovering business-object columns.
	// Default: true, matching C# ReportSettings.UsePropValuesToDiscoverBO.
	UsePropValuesToDiscoverBO bool
}

// NewReportSettings returns settings with default values matching C#.
func NewReportSettings() *ReportSettings {
	return &ReportSettings{
		DefaultPaperSize:          PaperA4,
		UsePropValuesToDiscoverBO: true,
	}
}
