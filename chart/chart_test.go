package chart_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/chart"
)

func TestChart_EmptySeriesReturnsBlankImage(t *testing.T) {
	c := &chart.Chart{Width: 200, Height: 100}
	img := c.Render()
	if img == nil {
		t.Fatal("Render should return non-nil even for empty series")
	}
	b := img.Bounds()
	if b.Max.X != 200 || b.Max.Y != 100 {
		t.Errorf("bounds = %v, want 200x100", b)
	}
}

func TestChart_NilDimensionsGetDefaults(t *testing.T) {
	c := &chart.Chart{
		Series: []chart.Series{{Values: []float64{1, 2, 3}}},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("nil dimensions should use defaults")
	}
	// Default 400x300.
	b := img.Bounds()
	if b.Max.X != 400 || b.Max.Y != 300 {
		t.Errorf("default bounds = %v, want 400x300", b)
	}
}

func TestChart_LineChart(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "S1", Values: []float64{10, 20, 30, 20, 10}},
		},
		ShowAxes: true,
		ShowGrid: true,
	}
	img := c.Render()
	assertNotBlank(t, img, "line chart should have non-blank pixels")
}

func TestChart_BarChart(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeBar,
		Series: []chart.Series{
			{Name: "Sales", Values: []float64{100, 200, 150}},
		},
		ShowAxes: true,
	}
	img := c.Render()
	assertNotBlank(t, img, "bar chart should have non-blank pixels")
}

func TestChart_AreaChart(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeArea,
		Series: []chart.Series{
			{Name: "Revenue", Values: []float64{5, 10, 8, 12}},
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "area chart should have non-blank pixels")
}

func TestChart_PieChart(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 200,
		Type:   chart.SeriesTypePie,
		Series: []chart.Series{
			{Name: "Pie", Values: []float64{30, 40, 20, 10}},
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "pie chart should have non-blank pixels")
}

func TestChart_MultiSeries(t *testing.T) {
	c := &chart.Chart{
		Width:      300,
		Height:     200,
		Type:       chart.SeriesTypeBar,
		ShowLegend: true,
		Series: []chart.Series{
			{Name: "Q1", Values: []float64{10, 20, 30}},
			{Name: "Q2", Values: []float64{15, 25, 35}},
			{Name: "Q3", Values: []float64{12, 22, 32}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("multi-series render = nil")
	}
}

func TestChart_CustomColor(t *testing.T) {
	red := color.RGBA{200, 0, 0, 255}
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "Red Series", Values: []float64{1, 5, 3}, Color: red},
		},
	}
	img := c.Render()
	// Check that the red color appears somewhere in the rendered image.
	found := false
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y && !found; y++ {
		for x := b.Min.X; x < b.Max.X && !found; x++ {
			r, g, b2, a := img.At(x, y).RGBA()
			if r > 0x8000 && g < 0x4000 && b2 < 0x4000 && a > 0 {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected red pixels in custom-color line chart")
	}
}

func TestChart_NegativeValues(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeBar,
		Series: []chart.Series{
			{Name: "S", Values: []float64{-10, 5, -5, 10}},
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "chart with negative values should render")
}

func assertNotBlank(t *testing.T, img image.Image, msg string) {
	t.Helper()
	if img == nil {
		t.Fatal(msg + " (got nil)")
	}
	b := img.Bounds()
	white := color.RGBA{255, 255, 255, 255}
	nonWhite := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bv, a := img.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			pix := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bv >> 8), uint8(a >> 8)}
			if pix != white {
				nonWhite++
			}
		}
	}
	if nonWhite == 0 {
		t.Error(msg + " (all pixels are white)")
	}
}
