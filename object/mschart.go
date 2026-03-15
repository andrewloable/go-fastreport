package object

import (
	"image/color"

	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/utils"
)

// MSChartSeries represents one data series inside an MSChartObject.
// It stores data binding and visual properties preserved for round-trip FRX fidelity.
type MSChartSeries struct {
	report.ReportComponentBase

	// ChartType is the series-level chart type override (e.g. "Line", "Bar", "Pie").
	// Empty string means inherit from the parent MSChartObject.
	ChartType string

	// Color is the series fill/line color.
	Color color.RGBA

	// ValuesSource is the data field expression bound to the Y-axis values.
	ValuesSource string

	// ArgumentSource is the data field expression bound to the X-axis / category.
	ArgumentSource string

	// LegendText is the label shown for this series in the chart legend.
	LegendText string
}

// NewMSChartSeries creates an MSChartSeries with defaults.
func NewMSChartSeries() *MSChartSeries {
	return &MSChartSeries{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (s *MSChartSeries) BaseName() string { return "Series" }

// TypeName returns "MSChartSeries".
func (s *MSChartSeries) TypeName() string { return "MSChartSeries" }

// Serialize writes MSChartSeries properties that differ from defaults.
func (s *MSChartSeries) Serialize(w report.Writer) error {
	if err := s.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if s.ChartType != "" {
		w.WriteStr("ChartType", s.ChartType)
	}
	if s.Color != (color.RGBA{}) {
		w.WriteStr("Color", utils.FormatColor(s.Color))
	}
	if s.ValuesSource != "" {
		w.WriteStr("ValuesSource", s.ValuesSource)
	}
	if s.ArgumentSource != "" {
		w.WriteStr("ArgumentSource", s.ArgumentSource)
	}
	if s.LegendText != "" {
		w.WriteStr("LegendText", s.LegendText)
	}
	return nil
}

// Deserialize reads MSChartSeries properties.
func (s *MSChartSeries) Deserialize(r report.Reader) error {
	if err := s.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	s.ChartType = r.ReadStr("ChartType", "")
	if cs := r.ReadStr("Color", ""); cs != "" {
		if c, err := utils.ParseColor(cs); err == nil {
			s.Color = c
		}
	}
	s.ValuesSource = r.ReadStr("ValuesSource", "")
	s.ArgumentSource = r.ReadStr("ArgumentSource", "")
	s.LegendText = r.ReadStr("LegendText", "")
	return nil
}

// MSChartObject renders a Microsoft Chart (MSChart) visualisation.
// The chart data and series definitions are stored in ChartData (base64).
//
// It is the Go equivalent of FastReport.MSChart.MSChartObject.
// This implementation supports FRX loading (deserialization) and serialization
// for round-trip fidelity; rendering is not yet implemented.
type MSChartObject struct {
	report.ReportComponentBase

	// ChartData holds the base64-encoded chart configuration / image data.
	ChartData string

	// ChartType is the global chart type (e.g. "Bar", "Line", "Pie").
	ChartType string

	// DataSource is the name of the bound data source.
	DataSource string

	// Series is the ordered list of data series.
	Series []*MSChartSeries
}

// NewMSChartObject creates an MSChartObject with default settings.
func NewMSChartObject() *MSChartObject {
	return &MSChartObject{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (m *MSChartObject) BaseName() string { return "Chart" }

// TypeName returns "MSChartObject".
func (m *MSChartObject) TypeName() string { return "MSChartObject" }

// Serialize writes MSChartObject properties that differ from defaults.
func (m *MSChartObject) Serialize(w report.Writer) error {
	if err := m.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if m.ChartData != "" {
		w.WriteStr("ChartData", m.ChartData)
	}
	if m.ChartType != "" {
		w.WriteStr("ChartType", m.ChartType)
	}
	if m.DataSource != "" {
		w.WriteStr("DataSource", m.DataSource)
	}
	for _, s := range m.Series {
		if err := w.WriteObjectNamed("MSChartSeries", s); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads MSChartObject properties from an FRX reader.
func (m *MSChartObject) Deserialize(r report.Reader) error {
	if err := m.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	m.ChartData = r.ReadStr("ChartData", "")
	m.ChartType = r.ReadStr("ChartType", "")
	m.DataSource = r.ReadStr("DataSource", "")
	return nil
}

// DeserializeChild handles MSChartSeries child elements.
func (m *MSChartObject) DeserializeChild(childType string, r report.Reader) bool {
	if childType == "MSChartSeries" {
		s := NewMSChartSeries()
		_ = s.Deserialize(r)
		// Drain any grandchildren.
		for {
			_, ok := r.NextChild()
			if !ok {
				break
			}
			if r.FinishChild() != nil { break }
		}
		m.Series = append(m.Series, s)
		return true
	}
	return false
}
