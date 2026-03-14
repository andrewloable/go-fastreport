package reportpkg

import (
	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
)

func init() {
	regs := []struct {
		name    string
		factory serial.Factory
	}{
		// Report-level containers
		{"Report", func() report.Base { return NewReport() }},
		{"ReportPage", func() report.Base { return NewReportPage() }},

		// Band types — short names (used by our custom FRX files and serializer)
		{"ReportTitle", func() report.Base { return band.NewReportTitleBand() }},
		{"ReportSummary", func() report.Base { return band.NewReportSummaryBand() }},
		{"PageHeader", func() report.Base { return band.NewPageHeaderBand() }},
		{"PageFooter", func() report.Base { return band.NewPageFooterBand() }},
		{"ColumnHeader", func() report.Base { return band.NewColumnHeaderBand() }},
		{"ColumnFooter", func() report.Base { return band.NewColumnFooterBand() }},
		{"DataHeader", func() report.Base { return band.NewDataHeaderBand() }},
		{"DataFooter", func() report.Base { return band.NewDataFooterBand() }},
		{"Data", func() report.Base { return band.NewDataBand() }},
		{"GroupHeader", func() report.Base { return band.NewGroupHeaderBand() }},
		{"GroupFooter", func() report.Base { return band.NewGroupFooterBand() }},
		{"Child", func() report.Base { return band.NewChildBand() }},
		{"Overlay", func() report.Base { return band.NewOverlayBand() }},

		// Band types — full names used by FastReport .NET FRX files (with "Band" suffix)
		{"ReportTitleBand", func() report.Base { return band.NewReportTitleBand() }},
		{"ReportSummaryBand", func() report.Base { return band.NewReportSummaryBand() }},
		{"PageHeaderBand", func() report.Base { return band.NewPageHeaderBand() }},
		{"PageFooterBand", func() report.Base { return band.NewPageFooterBand() }},
		{"ColumnHeaderBand", func() report.Base { return band.NewColumnHeaderBand() }},
		{"ColumnFooterBand", func() report.Base { return band.NewColumnFooterBand() }},
		{"DataHeaderBand", func() report.Base { return band.NewDataHeaderBand() }},
		{"DataFooterBand", func() report.Base { return band.NewDataFooterBand() }},
		{"DataBand", func() report.Base { return band.NewDataBand() }},
		{"GroupHeaderBand", func() report.Base { return band.NewGroupHeaderBand() }},
		{"GroupFooterBand", func() report.Base { return band.NewGroupFooterBand() }},
		{"ChildBand", func() report.Base { return band.NewChildBand() }},
		{"OverlayBand", func() report.Base { return band.NewOverlayBand() }},

		// Object types
		{"TextObject", func() report.Base { return object.NewTextObject() }},
		{"PictureObject", func() report.Base { return object.NewPictureObject() }},
		{"LineObject", func() report.Base { return object.NewLineObject() }},
		{"ShapeObject", func() report.Base { return object.NewShapeObject() }},
		{"PolyLineObject", func() report.Base { return object.NewPolyLineObject() }},
		{"PolygonObject", func() report.Base { return object.NewPolygonObject() }},
		{"CheckBoxObject", func() report.Base { return object.NewCheckBoxObject() }},
		{"ContainerObject", func() report.Base { return object.NewContainerObject() }},
		{"SubreportObject", func() report.Base { return object.NewSubreportObject() }},
	}

	for _, reg := range regs {
		// Ignore double-registration errors (e.g. if init runs multiple times in tests).
		_ = serial.DefaultRegistry.Register(reg.name, reg.factory)
	}
}
