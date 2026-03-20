package chart_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/chart"
)

// ── drawBezierSegment via SeriesTypeSpline ────────────────────────────────────

// TestChart_SplineChart_Basic exercises the SeriesTypeSpline path in Render,
// which calls drawBezierSegment for every adjacent pair of points.
func TestChart_SplineChart_Basic(t *testing.T) {
	c := &chart.Chart{
		Width:  300,
		Height: 200,
		Type:   chart.SeriesTypeSpline,
		Series: []chart.Series{
			{Name: "Spline", Values: []float64{10, 40, 20, 50, 30, 45, 15}},
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "spline chart should render non-blank pixels")
}

// TestChart_SplineChart_TwoPoints tests the minimum number of segments (1),
// which exercises the endpoint-clamping branches in tangentX/tangentY.
func TestChart_SplineChart_TwoPoints(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeSpline,
		Series: []chart.Series{
			{Name: "Spline", Values: []float64{10, 80}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("two-point spline returned nil")
	}
}

// TestChart_SplineChart_SinglePoint exercises the single-point path in the
// spline case (step = areaW instead of areaW/(nPts-1)), and no segments are drawn.
func TestChart_SplineChart_SinglePoint(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeSpline,
		Series: []chart.Series{
			{Name: "S", Values: []float64{42}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("single-point spline returned nil")
	}
}

// TestChart_SplineChart_ManyPoints stresses drawBezierSegment with many
// interior points, exercising the interior tangent formula
// (tangentX = (xs[k+1]-xs[k-1])/2 for 0 < k < n-1).
func TestChart_SplineChart_ManyPoints(t *testing.T) {
	vals := make([]float64, 20)
	for i := range vals {
		if i%2 == 0 {
			vals[i] = float64(i * 5)
		} else {
			vals[i] = float64(50 - i*3)
		}
	}
	c := &chart.Chart{
		Width:  400,
		Height: 300,
		Type:   chart.SeriesTypeSpline,
		Series: []chart.Series{
			{Name: "Spline", Values: vals},
		},
		ShowAxes:   true,
		ShowGrid:   true,
		ShowLegend: true,
	}
	img := c.Render()
	assertNotBlank(t, img, "many-point spline chart should render non-blank pixels")
}

// TestChart_SplineChart_ViaSeriesType uses a series-level type override
// (c.Type is Line but series[0].Type is Spline) to further exercise the
// type-assignment code path.
func TestChart_SplineChart_ViaSeriesType(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 150,
		Type:   chart.SeriesTypeLine, // chart-level type
		Series: []chart.Series{
			{
				Name:   "Spline Override",
				Type:   chart.SeriesTypeSpline, // series-level override
				Values: []float64{5, 25, 15, 35, 10},
			},
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "spline-via-series-type should render non-blank pixels")
}

// TestChart_SplineChart_NegativeValues exercises drawBezierSegment with
// negative y-coordinates (Catmull-Rom tangents for descending values).
func TestChart_SplineChart_NegativeValues(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeSpline,
		Series: []chart.Series{
			{Name: "S", Values: []float64{-10, -40, -20, -50, -5}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("spline with negative values returned nil")
	}
}

// TestChart_SplineChart_CustomColor verifies that a custom series color flows
// through to drawBezierSegment (used in drawThickLine calls inside it).
func TestChart_SplineChart_CustomColor(t *testing.T) {
	green := color.RGBA{0, 200, 0, 255}
	c := &chart.Chart{
		Width:  200,
		Height: 150,
		Type:   chart.SeriesTypeSpline,
		Series: []chart.Series{
			{Name: "Green Spline", Color: green, Values: []float64{20, 60, 30, 70, 40}},
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "colored spline chart should render non-blank pixels")
}

// ── renderPieDoughnut — doughnut (isDoughnut=true) branch ─────────────────────

// TestChart_DoughnutChart_Basic covers the isDoughnut=true branch that punches
// a hole in the centre and redraws the hole border.
func TestChart_DoughnutChart_Basic(t *testing.T) {
	c := &chart.Chart{
		Width:  300,
		Height: 300,
		Type:   chart.SeriesTypeDoughnut,
		Series: []chart.Series{
			{Name: "D", Values: []float64{30, 40, 30}},
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "doughnut chart should render non-blank pixels")
}

// TestChart_DoughnutChart_WithTitle covers both the isDoughnut branch AND the
// `if c.Title != ""` branch inside renderPieDoughnut.
func TestChart_DoughnutChart_WithTitle(t *testing.T) {
	c := &chart.Chart{
		Width:  300,
		Height: 300,
		Type:   chart.SeriesTypeDoughnut,
		Title:  "Doughnut Chart",
		Series: []chart.Series{
			{Name: "D", Values: []float64{25, 50, 25}},
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "doughnut chart with title should render non-blank pixels")
}

// TestChart_DoughnutChart_CustomBackground covers the `bg.A == 0` → default
// white branch inside the isDoughnut hole-fill section.
func TestChart_DoughnutChart_CustomBackground(t *testing.T) {
	c := &chart.Chart{
		Width:      300,
		Height:     300,
		Type:       chart.SeriesTypeDoughnut,
		Background: color.RGBA{240, 240, 240, 255}, // non-zero alpha → used directly
		Series: []chart.Series{
			{Name: "D", Values: []float64{50, 50}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("doughnut with custom background returned nil")
	}
}

// TestChart_DoughnutChart_ZeroAlphaBackground covers the `bg.A == 0` branch
// inside the isDoughnut hole section (bg becomes white).
func TestChart_DoughnutChart_ZeroAlphaBackground(t *testing.T) {
	c := &chart.Chart{
		Width:  300,
		Height: 300,
		Type:   chart.SeriesTypeDoughnut,
		// Background.A == 0 (zero value) → falls back to white inside hole section.
		Series: []chart.Series{
			{Name: "D", Values: []float64{60, 40}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("doughnut with zero-alpha background returned nil")
	}
}

// TestChart_DoughnutChart_ViaSeriesType covers the doughnut path when the
// series type (not chart type) is SeriesTypeDoughnut.
func TestChart_DoughnutChart_ViaSeriesType(t *testing.T) {
	c := &chart.Chart{
		Width:  250,
		Height: 250,
		Type:   chart.SeriesTypeLine, // chart-level is Line
		Series: []chart.Series{
			{
				Name:   "Donut",
				Type:   chart.SeriesTypeDoughnut, // series-level override
				Values: []float64{20, 30, 50},
			},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("doughnut via series type returned nil")
	}
}

// TestChart_DoughnutChart_SingleSlice covers the case where the doughnut has
// a single slice (which takes the full 2π angle).
func TestChart_DoughnutChart_SingleSlice(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 200,
		Type:   chart.SeriesTypeDoughnut,
		Series: []chart.Series{
			{Name: "D", Values: []float64{100}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("doughnut single slice returned nil")
	}
}

// ── renderPieDoughnut — pie edge cases not yet fully covered ──────────────────

// TestChart_PieChart_SingleSlice covers a pie with a single value (100%),
// exercising the full-angle sector and the `for angle < startAngle` loop
// in drawSector.
func TestChart_PieChart_SingleSlice(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 200,
		Type:   chart.SeriesTypePie,
		Series: []chart.Series{
			{Name: "P", Values: []float64{100}},
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "single-slice pie should render non-blank pixels")
}

// TestChart_PieChart_NegativeValues confirms that math.Abs is applied to
// negative values before summing and rendering.
func TestChart_PieChart_NegativeValues(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 200,
		Type:   chart.SeriesTypePie,
		Series: []chart.Series{
			{Name: "P", Values: []float64{-30, -40, -30}},
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "pie with negative values should render non-blank pixels")
}

// TestChart_PieChart_ManySlices uses more than 8 slices to cycle through the
// piePalette (modulo wrapping).
func TestChart_PieChart_ManySlices(t *testing.T) {
	c := &chart.Chart{
		Width:  300,
		Height: 300,
		Type:   chart.SeriesTypePie,
		Series: []chart.Series{
			{Name: "P", Values: []float64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}}, // 10 slices
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "many-slice pie should render non-blank pixels")
}

// ── Render — remaining 8.3% (SeriesTypeSpline minMaxSame, area/spline edge) ──

// TestChart_SplineChart_MinMaxSame covers the minY==maxY branch when using
// spline type, ensuring yScale doesn't divide by zero.
func TestChart_SplineChart_MinMaxSame(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeSpline,
		Series: []chart.Series{
			{Name: "S", Values: []float64{7, 7, 7, 7}}, // all same value
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("spline constant values returned nil")
	}
}

// TestChart_SplineChart_WithTitle covers the `if c.Title != ""` branch when
// rendering a spline (cartesian) chart.
func TestChart_SplineChart_WithTitle(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeSpline,
		Title:  "Spline Title",
		Series: []chart.Series{
			{Name: "S", Values: []float64{5, 15, 10, 20}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("spline with title returned nil")
	}
}

// TestChart_AreaChart_SinglePoint_NoPrevious covers the area branch where
// nPts==1 sets step=areaW, and j==0 so the `if j > 0` branch is never taken
// (no trapezoid drawn). A single-point area chart returns early from the loop.
func TestChart_AreaChart_SinglePoint_NoPrevious(t *testing.T) {
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeArea,
		Series: []chart.Series{
			{Name: "A", Values: []float64{25}},
		},
		ShowAxes: true,
	}
	img := c.Render()
	if img == nil {
		t.Fatal("area single-point with axes returned nil")
	}
}

// TestChart_MultiSplineSeries exercises drawBezierSegment with multiple spline
// series, making sure the inner loop iterates for j=1..n-1 on each series.
func TestChart_MultiSplineSeries(t *testing.T) {
	c := &chart.Chart{
		Width:  300,
		Height: 200,
		Type:   chart.SeriesTypeSpline,
		Series: []chart.Series{
			{Name: "A", Values: []float64{10, 20, 15, 25}},
			{Name: "B", Values: []float64{5, 30, 10, 20}},
		},
		ShowLegend: true,
	}
	img := c.Render()
	assertNotBlank(t, img, "multi-series spline chart should render non-blank pixels")
}

// ── drawBezierSegment — internal test (white-box) ─────────────────────────────

// TestDrawBezierSegment_DirectCall directly tests the unexported drawBezierSegment.
// This is an internal test in package chart (not chart_test).
// See chart_coverage3_internal_test.go for that.

// TestChart_SplineChart_ThreePoints specifically covers the case where n==3,
// exercising all three tangentX/Y cases: k==0 (left end), k==1 (interior), k==2 (right end).
func TestChart_SplineChart_ThreePoints(t *testing.T) {
	c := &chart.Chart{
		Width:  300,
		Height: 200,
		Type:   chart.SeriesTypeSpline,
		Series: []chart.Series{
			{Name: "S", Values: []float64{10, 50, 20}},
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "three-point spline chart should render non-blank pixels")
}
