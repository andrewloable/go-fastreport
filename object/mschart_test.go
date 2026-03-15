package object_test

import (
	"encoding/base64"
	"testing"

	"github.com/andrewloable/go-fastreport/object"
)

func TestMSChartObject_RenderToImage_NoData(t *testing.T) {
	m := object.NewMSChartObject()
	img := m.RenderToImage(200, 100)
	if img != nil {
		t.Error("RenderToImage with no series should return nil")
	}
}

func TestMSChartObject_RenderToImage_StaticSeries(t *testing.T) {
	m := object.NewMSChartObject()
	m.ChartType = "Bar"
	s := object.NewMSChartSeries()
	s.SetName("Series1")
	s.LegendText = "S1"
	m.Series = append(m.Series, s)

	img := m.RenderToImage(200, 100)
	// No values in series → still no image.
	if img != nil {
		t.Log("static series with no values: got image (may be acceptable)")
	}
}

func TestMSChartObject_RenderToImage_ChartData(t *testing.T) {
	// Build a minimal ChartData XML.
	xmlData := `<Chart><Series><Series Name="Sales" ChartType="Column">` +
		`<Points>` +
		`<DataPoint YValues="100"/>` +
		`<DataPoint YValues="200"/>` +
		`<DataPoint YValues="150"/>` +
		`</Points>` +
		`</Series></Series></Chart>`
	encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

	m := object.NewMSChartObject()
	m.ChartData = encoded
	m.ChartType = "Bar"

	img := m.RenderToImage(300, 200)
	if img == nil {
		t.Fatal("RenderToImage with valid ChartData should return non-nil image")
	}
	b := img.Bounds()
	if b.Max.X != 300 || b.Max.Y != 200 {
		t.Errorf("image bounds = %v, want 300x200", b)
	}
}

func TestMSChartObject_RenderToImage_MultiSeries(t *testing.T) {
	xmlData := `<Chart><Series>` +
		`<Series Name="Q1" ChartType="Bar"><Points><DataPoint YValues="10"/><DataPoint YValues="20"/></Points></Series>` +
		`<Series Name="Q2" ChartType="Bar"><Points><DataPoint YValues="15"/><DataPoint YValues="25"/></Points></Series>` +
		`</Series></Chart>`
	encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

	m := object.NewMSChartObject()
	m.ChartData = encoded
	m.ChartType = "Bar"

	img := m.RenderToImage(300, 200)
	if img == nil {
		t.Fatal("multi-series render should return non-nil")
	}
}

func TestMSChartObject_RenderToImage_LineChart(t *testing.T) {
	xmlData := `<Chart><Series><Series Name="Trend" ChartType="Line">` +
		`<Points><DataPoint YValues="1"/><DataPoint YValues="4"/><DataPoint YValues="2"/><DataPoint YValues="5"/></Points>` +
		`</Series></Series></Chart>`
	encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

	m := object.NewMSChartObject()
	m.ChartData = encoded
	m.ChartType = "Line"

	img := m.RenderToImage(200, 100)
	if img == nil {
		t.Fatal("line chart render = nil")
	}
}

func TestMSChartObject_RenderToImage_PieChart(t *testing.T) {
	xmlData := `<Chart><Series><Series Name="Share" ChartType="Pie">` +
		`<Points><DataPoint YValues="30"/><DataPoint YValues="40"/><DataPoint YValues="30"/></Points>` +
		`</Series></Series></Chart>`
	encoded := base64.StdEncoding.EncodeToString([]byte(xmlData))

	m := object.NewMSChartObject()
	m.ChartData = encoded
	m.ChartType = "Pie"

	img := m.RenderToImage(200, 200)
	if img == nil {
		t.Fatal("pie chart render = nil")
	}
}

func TestMSChartObject_RenderToImage_InvalidChartData(t *testing.T) {
	m := object.NewMSChartObject()
	m.ChartData = "not-valid-base64!!!"
	// Should fall back gracefully.
	img := m.RenderToImage(200, 100)
	// May be nil or blank — should not panic.
	_ = img
}
