package reportpkg

import (
	"github.com/andrewloable/go-fastreport/band"
	"github.com/andrewloable/go-fastreport/gauge"
	"github.com/andrewloable/go-fastreport/matrix"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/table"
)

func init() {
	regs := []struct {
		name    string
		factory serial.Factory
	}{
		// Report-level containers
		{"Report", func() report.Base { return NewReport() }},
		{"ReportPage", func() report.Base { return NewReportPage() }},
		{"DialogPage", func() report.Base { return NewDialogPage() }},

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
		{"BarcodeObject", func() report.Base { return object.NewBarcodeObject() }},
		{"ZipCodeObject", func() report.Base { return object.NewZipCodeObject() }},
		{"HtmlObject", func() report.Base { return object.NewHtmlObject() }},
		{"CellularTextObject", func() report.Base { return object.NewCellularTextObject() }},
		{"SVGObject", func() report.Base { return object.NewSVGObject() }},
		{"RichObject", func() report.Base { return object.NewRichObject() }},
		{"SparklineObject", func() report.Base { return object.NewSparklineObject() }},
		{"AdvMatrixObject", func() report.Base { return object.NewAdvMatrixObject() }},
		{"MSChartObject", func() report.Base { return object.NewMSChartObject() }},
		{"MSChartSeries", func() report.Base { return object.NewMSChartSeries() }},
		{"DigitalSignatureObject", func() report.Base { return object.NewDigitalSignatureObject() }},
		{"MapObject", func() report.Base { return object.NewMapObject() }},
		{"MapLayer", func() report.Base { return object.NewMapLayer() }},
		{"RFIDLabel", func() report.Base { return object.NewRFIDLabel() }},

		// Table object and its children
		{"TableObject", func() report.Base { return table.NewTableObject() }},
		{"TableColumn", func() report.Base { return table.NewTableColumn() }},
		{"TableRow", func() report.Base { return table.NewTableRow() }},
		{"TableCell", func() report.Base { return table.NewTableCell() }},

		// Matrix object
		{"MatrixObject", func() report.Base { return matrix.New() }},

		// Gauge objects
		{"LinearGauge", func() report.Base { return gauge.NewLinearGauge() }},
		{"RadialGauge", func() report.Base { return gauge.NewRadialGauge() }},
		{"SimpleGauge", func() report.Base { return gauge.NewSimpleGauge() }},
		{"SimpleProgressGauge", func() report.Base { return gauge.NewSimpleProgressGauge() }},
	}

	for _, reg := range regs {
		// Ignore double-registration errors (e.g. if init runs multiple times in tests).
		_ = serial.DefaultRegistry.Register(reg.name, reg.factory)
	}
}
