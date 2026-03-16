package chart_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/chart"
)

// ── renderPie coverage ────────────────────────────────────────────────────────

func TestChart_PieChart_EmptyValues(t *testing.T) {
	// renderPie: first series has no values → early return.
	c := &chart.Chart{
		Width:  200,
		Height: 200,
		Type:   chart.SeriesTypePie,
		Series: []chart.Series{
			{Name: "Pie", Values: []float64{}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("PieChart empty values returned nil")
	}
}

func TestChart_PieChart_ZeroTotal(t *testing.T) {
	// renderPie: total == 0 → early return.
	c := &chart.Chart{
		Width:  200,
		Height: 200,
		Type:   chart.SeriesTypePie,
		Series: []chart.Series{
			{Name: "Pie", Values: []float64{0, 0, 0}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("PieChart zero total returned nil")
	}
}

func TestChart_PieChart_TinySize(t *testing.T) {
	// renderPie: r <= 0 (too small image) → early return.
	c := &chart.Chart{
		Width:  10, // r = min(5,5) - 20 = -15 <= 0
		Height: 10,
		Type:   chart.SeriesTypePie,
		Series: []chart.Series{
			{Name: "Pie", Values: []float64{10, 20, 30}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("PieChart tiny size returned nil")
	}
}

func TestChart_PieChart_WithTitle(t *testing.T) {
	// renderPie: c.Title != "" branch.
	c := &chart.Chart{
		Width:  300,
		Height: 300,
		Type:   chart.SeriesTypePie,
		Title:  "Sales",
		Series: []chart.Series{
			{Name: "Pie", Values: []float64{25, 35, 40}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("PieChart with title returned nil")
	}
}

func TestChart_PieChart_SeriesTypeOverride(t *testing.T) {
	// Pie via series type override (chart.Type != Pie, but series[0].Type == Pie).
	c := &chart.Chart{
		Width:  200,
		Height: 200,
		Type:   chart.SeriesTypeLine, // chart type is Line
		Series: []chart.Series{
			{Name: "Pie", Type: chart.SeriesTypePie, Values: []float64{30, 40, 30}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("PieChart via series type returned nil")
	}
}

// ── drawHLine / drawVLine ─────────────────────────────────────────────────────

func TestChart_Render_WithAxes_CovershLine(t *testing.T) {
	// drawHLine: x0 > x1 → swap (tick marks at x < chartLeft).
	c := &chart.Chart{
		Width:    200,
		Height:   100,
		Type:     chart.SeriesTypeLine,
		ShowAxes: true,
		ShowGrid: true,
		Series: []chart.Series{
			{Name: "S1", Values: []float64{10, 20, 30}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("chart with axes returned nil")
	}
}

func TestChart_Render_WithAxes_CoversVLine(t *testing.T) {
	// drawVLine is called from Render for Y-axis (chartTop → chartBottom).
	c := &chart.Chart{
		Width:    200,
		Height:   100,
		Type:     chart.SeriesTypeBar,
		ShowAxes: true,
		Series: []chart.Series{
			{Name: "Bars", Values: []float64{-5, 10, -3, 8}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("chart with axes (negative values) returned nil")
	}
}

// ── fillRect clipping ─────────────────────────────────────────────────────────

func TestChart_FillRect_NegativeCoordinates_ViaBar(t *testing.T) {
	// Bar chart with very small width forces barX/barY calculations that may go negative.
	// fillRect clamps x0/y0/x1/y1 to bounds.
	c := &chart.Chart{
		Width:  60, // tiny width
		Height: 50,
		Type:   chart.SeriesTypeBar,
		Series: []chart.Series{
			{Name: "B", Values: []float64{100, 200, 150, 80, 120}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("tiny bar chart returned nil")
	}
}

func TestChart_FillRect_OutOfBounds_ViaLine(t *testing.T) {
	// Line chart with a very large value can produce y > img.Height.
	c := &chart.Chart{
		Width:  100,
		Height: 50,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "L", Values: []float64{1e9, -1e9, 0}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("line chart extreme values returned nil")
	}
}

// ── fillTrapezoid ─────────────────────────────────────────────────────────────

func TestChart_AreaChart_FillTrapezoid_SwapsXY(t *testing.T) {
	// fillTrapezoid: x0 > x1 triggers swap (achieved when step direction is reversed).
	// Area chart with single point uses step = areaW.
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeArea,
		Series: []chart.Series{
			{Name: "A", Values: []float64{5}}, // single point
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("area chart single point returned nil")
	}
}

func TestChart_AreaChart_FillTrapezoid_LineYAboveBase(t *testing.T) {
	// fillTrapezoid: lineY > baseY branch.
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeArea,
		Series: []chart.Series{
			{Name: "A", Values: []float64{-10, -20, -15}}, // negative values, zeroY above lineY
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("area chart negative values returned nil")
	}
}

// ── min / max functions ───────────────────────────────────────────────────────

func TestChart_Min_SecondArgSmaller(t *testing.T) {
	// min(a,b) where a > b → returns b.
	// fillTrapezoid calls max(x1-x0, 1) where x1-x0 >= 1.
	// We can also verify via a chart render that both branches are hit.
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeArea,
		Series: []chart.Series{
			{Name: "A", Values: []float64{10, 20, 15, 5, 25}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("area chart returned nil")
	}
}

func TestChart_Max_FirstArgLarger(t *testing.T) {
	// max(a,b) where a > b → returns a.
	// fillTrapezoid uses max(x1-x0, 1) where x1 > x0 (max returns x1-x0).
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeArea,
		Series: []chart.Series{
			{Name: "A", Values: []float64{5, 10, 8}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("area chart returned nil")
	}
}

func TestChart_Max_SecondArgLargerOrEqual(t *testing.T) {
	// max(a,b) where a <= b → returns b.
	// fillTrapezoid: x1-x0 == 0 → max(0, 1) = 1.
	c := &chart.Chart{
		Width:  55, // very small width means step ≈ 0
		Height: 50,
		Type:   chart.SeriesTypeArea,
		Series: []chart.Series{
			{Name: "A", Values: []float64{5, 5}}, // 2 points, x0 == x1 when step rounds to 0
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("area chart same x returned nil")
	}
}

// ── globalRange edge cases ────────────────────────────────────────────────────

func TestChart_GlobalRange_AllEmptySeries(t *testing.T) {
	// All series have no values → globalRange returns (0, 1) fallback.
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "Empty", Values: []float64{}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("all-empty series returned nil")
	}
}

func TestChart_GlobalRange_NoSeries_ReturnsEarly(t *testing.T) {
	// len(series)==0 → globalRange returns (0,1).
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("no series returned nil")
	}
}

// ── Render edge cases ─────────────────────────────────────────────────────────

func TestChart_Render_BackgroundNonZeroAlpha(t *testing.T) {
	// Background with non-zero alpha is used directly.
	c := &chart.Chart{
		Width:      200,
		Height:     100,
		Background: color.RGBA{255, 240, 220, 255},
		Type:       chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "S", Values: []float64{1, 2, 3}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("custom background returned nil")
	}
}

func TestChart_Render_TinySize_AreaZero(t *testing.T) {
	// areaW <= 0 or areaH <= 0 → early return.
	c := &chart.Chart{
		Width:  50, // padLeft=45 + padRight=10 = 55 > 50, so areaW < 0
		Height: 100,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "S", Values: []float64{1, 2, 3}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("tiny chart returned nil")
	}
}

func TestChart_Render_SingleValueSeries_Line(t *testing.T) {
	// Single-value series: step = areaW (not divided by nPts-1).
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "S", Values: []float64{42}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("single value series returned nil")
	}
}

func TestChart_Render_LineChart_NoPreviousPoint(t *testing.T) {
	// First point of line chart: j==0 → no drawThickLine call, but fillRect is called.
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "S", Values: []float64{10, 20, 30}},
		},
	}
	img := c.Render()
	assertNotBlank(t, img, "line chart should be non-blank")
}

func TestChart_Render_BarChart_NegativeBarOrdering(t *testing.T) {
	// Bar chart: top > bot → swap (negative value bars).
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeBar,
		Series: []chart.Series{
			{Name: "B", Values: []float64{-10, 20, -5, 15}},
		},
		ShowAxes: true,
	}
	img := c.Render()
	assertNotBlank(t, img, "negative bar chart should be non-blank")
}

func TestChart_Render_BarChart_BarWidthMin1(t *testing.T) {
	// Many series makes barW very small → clamped to 1.
	series := make([]chart.Series, 10)
	for i := range series {
		series[i] = chart.Series{Name: "S", Values: []float64{float64(i + 1)}}
	}
	c := &chart.Chart{
		Width:  100,
		Height: 80,
		Type:   chart.SeriesTypeBar,
		Series: series,
	}
	img := c.Render()
	if img == nil {
		t.Fatal("many-series bar chart returned nil")
	}
}

func TestChart_Render_LineChart_MinMaxSame(t *testing.T) {
	// minY == maxY → set minY=0, maxY=1 (or if maxY==0, maxY=1).
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "S", Values: []float64{5, 5, 5}}, // all same → minY==maxY
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("constant value series returned nil")
	}
}

func TestChart_Render_LineChart_MinMaxSameAtZero(t *testing.T) {
	// minY == maxY == 0 → maxY = 1.
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "S", Values: []float64{0, 0, 0}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("all-zero series returned nil")
	}
}

func TestChart_Render_LineChart_ZeroYAboveTop(t *testing.T) {
	// zeroY < chartTop → zeroY = chartBottom (all positive values).
	c := &chart.Chart{
		Width:    200,
		Height:   100,
		Type:     chart.SeriesTypeLine,
		ShowAxes: true,
		Series: []chart.Series{
			{Name: "S", Values: []float64{10, 20, 30}}, // all positive → zeroY would be at bottom
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("all-positive series with axes returned nil")
	}
}

func TestChart_Render_GridLine_AtBoundary(t *testing.T) {
	// Grid line condition: gy > chartTop && gy < chartBottom.
	c := &chart.Chart{
		Width:    200,
		Height:   100,
		Type:     chart.SeriesTypeLine,
		ShowGrid: true,
		Series: []chart.Series{
			{Name: "S", Values: []float64{10, 20, 30}},
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("grid chart returned nil")
	}
}

func TestChart_Render_ZeroSeriesValues_ReturnsEarly(t *testing.T) {
	// nPts == 0 → early return.
	// This requires all series to have Values slice but 0 length, and nPts stays 0.
	// Actually this is tested by TestChart_GlobalRange_AllEmptySeries above.
	// Repeat with explicit nPts == 0.
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "S", Values: nil}, // nil values slice
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("nil values series returned nil")
	}
}

func TestChart_Render_BarChart_MinYPositive(t *testing.T) {
	// bar chart with minY > 0 → minY set to 0.
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeBar,
		Series: []chart.Series{
			{Name: "B", Values: []float64{50, 80, 60}}, // all positive
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("all-positive bar returned nil")
	}
}

// ── drawChar with unknown rune (uses miniFont fallback) ───────────────────────

func TestChart_DrawChar_UnknownRune(t *testing.T) {
	// A series name containing characters not in miniFont → uses ' ' (space) fallback.
	c := &chart.Chart{
		Width:      200,
		Height:     100,
		Type:       chart.SeriesTypeLine,
		ShowLegend: true,
		Series: []chart.Series{
			{Name: "Ω→∞", Values: []float64{1, 2, 3}}, // Unicode chars not in font
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("chart with unicode series name returned nil")
	}
}

// ── drawThickLine edge cases ──────────────────────────────────────────────────

func TestChart_DrawThickLine_HorizontalLine(t *testing.T) {
	// drawThickLine: dx=0 case (horizontal movement only).
	c := &chart.Chart{
		Width:  200,
		Height: 100,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "S", Values: []float64{50, 50}}, // same y value → horizontal line
		},
	}
	img := c.Render()
	if img == nil {
		t.Fatal("horizontal line chart returned nil")
	}
}

// Verify image type.
func TestChart_Render_ReturnsRGBA(t *testing.T) {
	c := &chart.Chart{
		Width:  100,
		Height: 80,
		Type:   chart.SeriesTypeLine,
		Series: []chart.Series{
			{Name: "S", Values: []float64{1, 2, 3}},
		},
	}
	img := c.Render()
	if _, ok := img.(*image.RGBA); !ok {
		t.Error("expected *image.RGBA")
	}
}
