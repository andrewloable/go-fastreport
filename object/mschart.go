package object

import "github.com/andrewloable/go-fastreport/report"

// MSChartSeries represents one data series inside an MSChartObject.
type MSChartSeries struct {
	report.ReportComponentBase
}

// NewMSChartSeries creates an MSChartSeries.
func NewMSChartSeries() *MSChartSeries {
	return &MSChartSeries{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (s *MSChartSeries) BaseName() string { return "Series" }

// TypeName returns "MSChartSeries".
func (s *MSChartSeries) TypeName() string { return "MSChartSeries" }

// Serialize writes MSChartSeries properties.
func (s *MSChartSeries) Serialize(w report.Writer) error {
	return s.ReportComponentBase.Serialize(w)
}

// Deserialize reads MSChartSeries properties.
func (s *MSChartSeries) Deserialize(r report.Reader) error {
	return s.ReportComponentBase.Deserialize(r)
}

// MSChartObject renders a Microsoft Chart (MSChart) visualisation.
// The chart data and series definitions are stored in ChartData (base64).
//
// It is the Go equivalent of FastReport.MSChart.MSChartObject.
// This stub supports FRX loading; rendering is not yet implemented.
type MSChartObject struct {
	report.ReportComponentBase

	// ChartData holds the base64-encoded chart configuration.
	ChartData string
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
	return nil
}

// Deserialize reads MSChartObject properties from an FRX reader.
func (m *MSChartObject) Deserialize(r report.Reader) error {
	if err := m.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	m.ChartData = r.ReadStr("ChartData", "")
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
			_ = r.FinishChild()
		}
		m.Series = append(m.Series, s)
		return true
	}
	return false
}
