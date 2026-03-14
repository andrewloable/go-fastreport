package serial

import (
	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func init() {
	regs := []struct {
		name    string
		factory Factory
	}{
		// Report-level containers
		{"Report", func() report.Base { return reportpkg.NewReport() }},
		{"ReportPage", func() report.Base { return reportpkg.NewReportPage() }},

		// Band types
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
		DefaultRegistry.MustRegister(reg.name, reg.factory)
	}
}
