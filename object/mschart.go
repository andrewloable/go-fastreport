package object

import (
	"encoding/base64"
	"encoding/xml"
	"image"
	"image/color"
	"strings"

	"github.com/andrewloable/go-fastreport/chart"
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

// RenderToImage renders the MSChartObject as a raster image of the given size.
// If ChartData is set, all embedded series are decoded and rendered.
// If no ChartData is available but Series have static values, those are used.
// Returns nil if there is no data to render.
func (m *MSChartObject) RenderToImage(w, h int) image.Image {
	c := &chart.Chart{
		Title:      m.Name(),
		Width:      w,
		Height:     h,
		ShowAxes:   true,
		ShowGrid:   true,
		ShowLegend: len(m.Series) > 1,
		Type:       msChartType(m.ChartType),
	}

	// Decode series from ChartData XML.
	if m.ChartData != "" {
		decoded := decodeAllSeries(m.ChartData)
		for i, ds := range decoded {
			sc := chart.Series{
				Name:   ds.name,
				Type:   msChartTypeStr(ds.chartType),
				Values: ds.values,
				Labels: ds.labels,
			}
			// Assign color from MSChartSeries if available.
			if i < len(m.Series) && m.Series[i].Color != (color.RGBA{}) {
				sc.Color = m.Series[i].Color
			}
			c.Series = append(c.Series, sc)
		}
	}

	// Fall back to MSChartSeries if no ChartData decoded.
	if len(c.Series) == 0 {
		for _, s := range m.Series {
			sc := chart.Series{
				Name:  s.LegendText,
				Type:  msChartTypeStr(s.ChartType),
				Color: s.Color,
			}
			c.Series = append(c.Series, sc)
		}
	}

	if len(c.Series) == 0 {
		return nil
	}

	return c.Render()
}

// msChartType converts a chart type string to chart.SeriesType.
func msChartType(t string) chart.SeriesType {
	return msChartTypeStr(t)
}

func msChartTypeStr(t string) chart.SeriesType {
	switch strings.ToLower(t) {
	case "bar", "column", "stackedbar", "stackedcolumn":
		return chart.SeriesTypeBar
	case "area", "stackedarea":
		return chart.SeriesTypeArea
	case "pie", "doughnut":
		return chart.SeriesTypePie
	default:
		return chart.SeriesTypeLine
	}
}

// ── ChartData XML decoding ────────────────────────────────────────────────────

type mscXML struct {
	Series mscSeriesCollection `xml:"Series"`
}
type mscSeriesCollection struct {
	Items []mscSeriesXML `xml:"Series"`
}
type mscSeriesXML struct {
	Name      string             `xml:"Name,attr"`
	ChartType string             `xml:"ChartType,attr"`
	Points    mscPointCollection `xml:"Points"`
}
type mscPointCollection struct {
	Items []mscDataPointXML `xml:"DataPoint"`
}
type mscDataPointXML struct {
	YValues string `xml:"YValues,attr"`
	AxisLabel string `xml:"AxisLabel,attr"`
}

type decodedSeries struct {
	name      string
	chartType string
	values    []float64
	labels    []string
}

// decodeAllSeries decodes all series from the base64 or raw XML ChartData.
func decodeAllSeries(chartData string) []decodedSeries {
	if chartData == "" {
		return nil
	}
	xmlBytes, err := base64.StdEncoding.DecodeString(chartData)
	if err != nil {
		xmlBytes = []byte(chartData)
	}
	var ch mscXML
	if err := xml.Unmarshal(xmlBytes, &ch); err != nil {
		return nil
	}
	result := make([]decodedSeries, 0, len(ch.Series.Items))
	for _, s := range ch.Series.Items {
		ds := decodedSeries{
			name:      s.Name,
			chartType: s.ChartType,
		}
		for _, pt := range s.Points.Items {
			raw := strings.SplitN(pt.YValues, ",", 2)[0]
			ds.values = append(ds.values, parseFloatMSC(raw))
			ds.labels = append(ds.labels, pt.AxisLabel)
		}
		result = append(result, ds)
	}
	return result
}

func parseFloatMSC(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}
	parts := strings.SplitN(s, ".", 2)
	var intPart, fracPart float64
	for _, c := range parts[0] {
		if c >= '0' && c <= '9' {
			intPart = intPart*10 + float64(c-'0')
		}
	}
	if len(parts) == 2 {
		mul := 0.1
		for _, c := range parts[1] {
			if c >= '0' && c <= '9' {
				fracPart += float64(c-'0') * mul
				mul *= 0.1
			}
		}
	}
	v := intPart + fracPart
	if neg {
		v = -v
	}
	return v
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
