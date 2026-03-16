package sparkline_test

import (
	"encoding/base64"
	"fmt"
	"image"
	"testing"

	"github.com/andrewloable/go-fastreport/sparkline"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// buildChartXML builds a minimal ChartData XML string for testing.
func buildChartXML(chartType string, values []float64) string {
	pts := ""
	for _, v := range values {
		pts += fmt.Sprintf(`<DataPoint YValues="%.6g"/>`, v)
	}
	return fmt.Sprintf(`<Chart><Series><Series Name="S1" ChartType="%s"><Points>%s</Points></Series></Series></Chart>`,
		chartType, pts)
}

// buildChartDataBase64 returns base64-encoded ChartData XML.
func buildChartDataBase64(chartType string, values []float64) string {
	return base64.StdEncoding.EncodeToString([]byte(buildChartXML(chartType, values)))
}

// ── DecodeChartData ───────────────────────────────────────────────────────────

func TestDecodeChartData_Empty(t *testing.T) {
	s := sparkline.DecodeChartData("")
	if s != nil {
		t.Errorf("empty chartData: got %v, want nil", s)
	}
}

func TestDecodeChartData_InvalidBase64_FallsBackToRawXML(t *testing.T) {
	// Not valid base64, but valid XML.
	raw := buildChartXML("FastLine", []float64{1, 2, 3})
	s := sparkline.DecodeChartData(raw)
	if s == nil {
		t.Fatal("raw XML should parse successfully")
	}
	if len(s.Values) != 3 {
		t.Errorf("values count = %d, want 3", len(s.Values))
	}
}

func TestDecodeChartData_InvalidXML(t *testing.T) {
	// Invalid XML → nil.
	s := sparkline.DecodeChartData("not xml at all!")
	if s != nil {
		t.Errorf("invalid XML should return nil, got %v", s)
	}
}

func TestDecodeChartData_ValidBase64_Line(t *testing.T) {
	cd := buildChartDataBase64("Line", []float64{10, 20, 15, 25})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData returned nil for Line")
	}
	if s.Type != sparkline.ChartTypeLine {
		t.Errorf("Type = %v, want ChartTypeLine", s.Type)
	}
	if len(s.Values) != 4 {
		t.Errorf("values count = %d, want 4", len(s.Values))
	}
}

func TestDecodeChartData_ValidBase64_Area(t *testing.T) {
	cd := buildChartDataBase64("Area", []float64{5, 10, 8})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData returned nil for Area")
	}
	if s.Type != sparkline.ChartTypeArea {
		t.Errorf("Type = %v, want ChartTypeArea", s.Type)
	}
}

func TestDecodeChartData_ValidBase64_Column(t *testing.T) {
	cd := buildChartDataBase64("Column", []float64{3, 5, 2})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData returned nil for Column")
	}
	if s.Type != sparkline.ChartTypeColumn {
		t.Errorf("Type = %v, want ChartTypeColumn", s.Type)
	}
}

func TestDecodeChartData_ValidBase64_Bar(t *testing.T) {
	cd := buildChartDataBase64("Bar", []float64{1, 2, 3})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData returned nil for Bar")
	}
	if s.Type != sparkline.ChartTypeColumn {
		t.Errorf("Type = %v, want ChartTypeColumn (Bar maps to Column)", s.Type)
	}
}

func TestDecodeChartData_ValidBase64_StackedColumn(t *testing.T) {
	cd := buildChartDataBase64("StackedColumn", []float64{10, 20})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData returned nil for StackedColumn")
	}
	if s.Type != sparkline.ChartTypeColumn {
		t.Errorf("Type = %v, want ChartTypeColumn", s.Type)
	}
}

func TestDecodeChartData_ValidBase64_StackedBar(t *testing.T) {
	cd := buildChartDataBase64("StackedBar", []float64{1})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData returned nil for StackedBar")
	}
	if s.Type != sparkline.ChartTypeColumn {
		t.Errorf("Type = %v, want ChartTypeColumn", s.Type)
	}
}

func TestDecodeChartData_ValidBase64_WinLoss(t *testing.T) {
	cd := buildChartDataBase64("WinLoss", []float64{1, -1, 1, -1})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData returned nil for WinLoss")
	}
	if s.Type != sparkline.ChartTypeWinLoss {
		t.Errorf("Type = %v, want ChartTypeWinLoss", s.Type)
	}
}

func TestDecodeChartData_ValidBase64_FastLine(t *testing.T) {
	cd := buildChartDataBase64("FastLine", []float64{1, 5, 3})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData returned nil for FastLine")
	}
	if s.Type != sparkline.ChartTypeLine {
		t.Errorf("Type = %v, want ChartTypeLine", s.Type)
	}
}

func TestDecodeChartData_EmptySeries(t *testing.T) {
	// XML with no series → returns nil.
	raw := `<Chart><Series></Series></Chart>`
	s := sparkline.DecodeChartData(raw)
	if s != nil {
		t.Errorf("empty series should return nil, got %v", s)
	}
}

func TestDecodeChartData_NegativeValues(t *testing.T) {
	cd := buildChartDataBase64("Line", []float64{-5, -10, 3})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData returned nil for negative values")
	}
	if len(s.Values) != 3 {
		t.Errorf("count = %d, want 3", len(s.Values))
	}
	if s.Values[0] != -5 {
		t.Errorf("Values[0] = %v, want -5", s.Values[0])
	}
}

func TestDecodeChartData_FloatValues(t *testing.T) {
	raw := `<Chart><Series><Series Name="S" ChartType="Line"><Points>
		<DataPoint YValues="3.14"/>
		<DataPoint YValues="2.72"/>
	</Points></Series></Series></Chart>`
	s := sparkline.DecodeChartData(raw)
	if s == nil {
		t.Fatal("float values returned nil")
	}
	if len(s.Values) != 2 {
		t.Errorf("count = %d, want 2", len(s.Values))
	}
}

func TestDecodeChartData_CommaSeparatedYValues(t *testing.T) {
	// YValues with comma-separated values — only first is used.
	raw := `<Chart><Series><Series Name="S" ChartType="Line"><Points>
		<DataPoint YValues="10,20,30"/>
	</Points></Series></Series></Chart>`
	s := sparkline.DecodeChartData(raw)
	if s == nil {
		t.Fatal("comma-separated YValues returned nil")
	}
	if s.Values[0] != 10 {
		t.Errorf("first YValue = %v, want 10", s.Values[0])
	}
}

func TestDecodeChartData_EmptyYValues(t *testing.T) {
	// YValues="" → value is 0.
	raw := `<Chart><Series><Series Name="S" ChartType="Line"><Points>
		<DataPoint YValues=""/>
	</Points></Series></Series></Chart>`
	s := sparkline.DecodeChartData(raw)
	if s == nil {
		t.Fatal("empty YValues returned nil")
	}
	if s.Values[0] != 0 {
		t.Errorf("empty YValue = %v, want 0", s.Values[0])
	}
}

// ── parseFloat ────────────────────────────────────────────────────────────────

func TestDecodeChartData_ParseFloat_WithFractional(t *testing.T) {
	raw := `<Chart><Series><Series Name="S" ChartType="Line"><Points>
		<DataPoint YValues="1.5"/>
		<DataPoint YValues="2.75"/>
	</Points></Series></Series></Chart>`
	s := sparkline.DecodeChartData(raw)
	if s == nil {
		t.Fatal("fractional values returned nil")
	}
	if s.Values[0] != 1.5 {
		t.Errorf("Values[0] = %v, want 1.5", s.Values[0])
	}
	if s.Values[1] != 2.75 {
		t.Errorf("Values[1] = %v, want 2.75", s.Values[1])
	}
}

func TestDecodeChartData_ParseFloat_NegativeWithFractional(t *testing.T) {
	raw := `<Chart><Series><Series Name="S" ChartType="Line"><Points>
		<DataPoint YValues="-3.5"/>
	</Points></Series></Series></Chart>`
	s := sparkline.DecodeChartData(raw)
	if s == nil {
		t.Fatal("negative fractional returned nil")
	}
	if s.Values[0] != -3.5 {
		t.Errorf("Values[0] = %v, want -3.5", s.Values[0])
	}
}

func TestDecodeChartData_ParseFloat_WhitespaceStripped(t *testing.T) {
	raw := `<Chart><Series><Series Name="S" ChartType="Line"><Points>
		<DataPoint YValues="  42  "/>
	</Points></Series></Series></Chart>`
	s := sparkline.DecodeChartData(raw)
	if s == nil {
		t.Fatal("whitespace value returned nil")
	}
	if s.Values[0] != 42 {
		t.Errorf("Values[0] = %v, want 42", s.Values[0])
	}
}

// ── Render ─────────────────────────────────────────────────────────────────────

func TestRender_NilSeries(t *testing.T) {
	img := sparkline.Render(nil, 100, 40)
	if img != nil {
		t.Error("nil series should return nil image")
	}
}

func TestRender_EmptyValues(t *testing.T) {
	s := &sparkline.Series{Type: sparkline.ChartTypeLine, Values: []float64{}}
	img := sparkline.Render(s, 100, 40)
	if img != nil {
		t.Error("empty values should return nil image")
	}
}

func TestRender_ZeroWidth(t *testing.T) {
	s := &sparkline.Series{Type: sparkline.ChartTypeLine, Values: []float64{1, 2, 3}}
	img := sparkline.Render(s, 0, 40)
	if img != nil {
		t.Error("zero width should return nil image")
	}
}

func TestRender_ZeroHeight(t *testing.T) {
	s := &sparkline.Series{Type: sparkline.ChartTypeLine, Values: []float64{1, 2, 3}}
	img := sparkline.Render(s, 100, 0)
	if img != nil {
		t.Error("zero height should return nil image")
	}
}

func TestRender_Line_Basic(t *testing.T) {
	s := &sparkline.Series{Type: sparkline.ChartTypeLine, Values: []float64{10, 20, 15, 25, 5}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("Render Line returned nil")
	}
	b := img.Bounds()
	if b.Dx() != 100 || b.Dy() != 40 {
		t.Errorf("size = %dx%d, want 100x40", b.Dx(), b.Dy())
	}
}

func TestRender_Line_SinglePoint(t *testing.T) {
	s := &sparkline.Series{Type: sparkline.ChartTypeLine, Values: []float64{42}}
	img := sparkline.Render(s, 80, 30)
	if img == nil {
		t.Fatal("Render Line single point returned nil")
	}
}

func TestRender_Line_ConstantValues(t *testing.T) {
	// vmax == vmin → vmax = vmin + 1.
	s := &sparkline.Series{Type: sparkline.ChartTypeLine, Values: []float64{5, 5, 5}}
	img := sparkline.Render(s, 80, 30)
	if img == nil {
		t.Fatal("Render constant values returned nil")
	}
}

func TestRender_Area_Basic(t *testing.T) {
	s := &sparkline.Series{Type: sparkline.ChartTypeArea, Values: []float64{5, 10, 8, 12, 3}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("Render Area returned nil")
	}
	assertImgNotBlank(t, img, "area chart should have non-white pixels")
}

func TestRender_Area_NegativeValues(t *testing.T) {
	// vmin > 0 → vmin = 0 (not triggered for negative); vmin <= 0 already.
	s := &sparkline.Series{Type: sparkline.ChartTypeArea, Values: []float64{-5, -10, 3}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("Render Area negative returned nil")
	}
}

func TestRender_Area_AllPositive(t *testing.T) {
	// vmin > 0 → set vmin = 0.
	s := &sparkline.Series{Type: sparkline.ChartTypeArea, Values: []float64{5, 10, 8}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("Render Area all positive returned nil")
	}
}

func TestRender_Area_SinglePoint(t *testing.T) {
	// n == 1 → loop doesn't execute, returns blank.
	s := &sparkline.Series{Type: sparkline.ChartTypeArea, Values: []float64{7}}
	img := sparkline.Render(s, 80, 30)
	if img == nil {
		t.Fatal("Render Area single point returned nil")
	}
}

func TestRender_Area_LineBelowBaseline(t *testing.T) {
	// y > baseline (negative values where y is below baseline in image coords).
	s := &sparkline.Series{Type: sparkline.ChartTypeArea, Values: []float64{10, -5, 8}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("Render Area y<baseline returned nil")
	}
}

func TestRender_Column_Basic(t *testing.T) {
	s := &sparkline.Series{Type: sparkline.ChartTypeColumn, Values: []float64{3, 5, 2, 8, 1}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("Render Column returned nil")
	}
	assertImgNotBlank(t, img, "column chart should have non-white pixels")
}

func TestRender_Column_NegativeValues(t *testing.T) {
	s := &sparkline.Series{Type: sparkline.ChartTypeColumn, Values: []float64{-3, 5, -2}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("Render Column negative returned nil")
	}
}

func TestRender_Column_AllNegative(t *testing.T) {
	// vmax < 0 → set vmax = 0.
	s := &sparkline.Series{Type: sparkline.ChartTypeColumn, Values: []float64{-5, -10, -3}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("Render Column all negative returned nil")
	}
}

func TestRender_Column_BarWidthMin1(t *testing.T) {
	// barW < 1 → barW = 1 (many points in small image).
	vals := make([]float64, 200)
	for i := range vals {
		vals[i] = float64(i % 10)
	}
	s := &sparkline.Series{Type: sparkline.ChartTypeColumn, Values: vals}
	img := sparkline.Render(s, 50, 30)
	if img == nil {
		t.Fatal("Render Column many points returned nil")
	}
}

func TestRender_WinLoss_Basic(t *testing.T) {
	s := &sparkline.Series{Type: sparkline.ChartTypeWinLoss, Values: []float64{1, -1, 1, 0, -1}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("Render WinLoss returned nil")
	}
	assertImgNotBlank(t, img, "winloss chart should have non-white pixels")
}

func TestRender_WinLoss_AllPositive(t *testing.T) {
	s := &sparkline.Series{Type: sparkline.ChartTypeWinLoss, Values: []float64{1, 2, 3}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("Render WinLoss all positive returned nil")
	}
}

func TestRender_WinLoss_AllNegative(t *testing.T) {
	s := &sparkline.Series{Type: sparkline.ChartTypeWinLoss, Values: []float64{-1, -2, -3}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("Render WinLoss all negative returned nil")
	}
}

func TestRender_WinLoss_AllZero(t *testing.T) {
	// All zero → normalized[i] = 0 for all.
	s := &sparkline.Series{Type: sparkline.ChartTypeWinLoss, Values: []float64{0, 0, 0}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("Render WinLoss all zero returned nil")
	}
}

// ── scaleX and scaleY edge cases ──────────────────────────────────────────────

func TestRender_ScaleX_SinglePoint(t *testing.T) {
	// n <= 1 → scaleX returns w/2.
	s := &sparkline.Series{Type: sparkline.ChartTypeLine, Values: []float64{10}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("single point returned nil")
	}
}

func TestRender_ScaleY_ConstantValue(t *testing.T) {
	// vmax == vmin → return h/2.
	s := &sparkline.Series{Type: sparkline.ChartTypeArea, Values: []float64{7, 7}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("constant area returned nil")
	}
}

// ── drawLine edge cases ───────────────────────────────────────────────────────

func TestRender_Line_VerticalSegment(t *testing.T) {
	// drawLine: |dy| > |dx| (vertical-dominant segment).
	s := &sparkline.Series{Type: sparkline.ChartTypeLine, Values: []float64{0, 100}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("vertical segment returned nil")
	}
}

func TestRender_Line_ZeroLengthSegment(t *testing.T) {
	// drawLine: dx == 0 && dy == 0 (same point) → setPixel once.
	// This happens when two consecutive values map to the same pixel.
	s := &sparkline.Series{Type: sparkline.ChartTypeLine, Values: []float64{50, 50, 50}}
	img := sparkline.Render(s, 10, 10) // tiny image → points overlap
	if img == nil {
		t.Fatal("zero-length segment returned nil")
	}
}

// ── drawVLine in renderBars ───────────────────────────────────────────────────

func TestRender_Column_BarAboveBaseline(t *testing.T) {
	// y <= baseline → drawVLine(x, y, baseline).
	s := &sparkline.Series{Type: sparkline.ChartTypeColumn, Values: []float64{10, 5, 8}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("column above baseline returned nil")
	}
}

func TestRender_Column_BarBelowBaseline(t *testing.T) {
	// y > baseline → drawVLine(x, baseline, y) (negative value).
	s := &sparkline.Series{Type: sparkline.ChartTypeColumn, Values: []float64{-5, 3, -8}}
	img := sparkline.Render(s, 100, 40)
	if img == nil {
		t.Fatal("column below baseline returned nil")
	}
}

// ── Full round-trip via DecodeChartData + Render ──────────────────────────────

func TestRoundTrip_Line(t *testing.T) {
	cd := buildChartDataBase64("Line", []float64{10, 20, 15, 25, 5})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData returned nil")
	}
	img := sparkline.Render(s, 120, 40)
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestRoundTrip_Area(t *testing.T) {
	cd := buildChartDataBase64("Area", []float64{5, 10, 8, 3})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData Area returned nil")
	}
	img := sparkline.Render(s, 120, 40)
	if img == nil {
		t.Fatal("Render Area returned nil")
	}
}

func TestRoundTrip_Column(t *testing.T) {
	cd := buildChartDataBase64("Column", []float64{3, -2, 5, -1})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData Column returned nil")
	}
	img := sparkline.Render(s, 120, 40)
	if img == nil {
		t.Fatal("Render Column returned nil")
	}
}

func TestRoundTrip_WinLoss(t *testing.T) {
	cd := buildChartDataBase64("WinLoss", []float64{1, -1, 1, 1, -1})
	s := sparkline.DecodeChartData(cd)
	if s == nil {
		t.Fatal("DecodeChartData WinLoss returned nil")
	}
	img := sparkline.Render(s, 120, 40)
	if img == nil {
		t.Fatal("Render WinLoss returned nil")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func assertImgNotBlank(t *testing.T, img image.Image, msg string) {
	t.Helper()
	if img == nil {
		t.Fatal(msg + " (got nil)")
	}
	b := img.Bounds()
	white := 0
	total := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bv, a := img.At(x, y).RGBA()
			total++
			if a > 0 && r > 0xfe00 && g > 0xfe00 && bv > 0xfe00 {
				white++
			}
		}
	}
	if total > 0 && white == total {
		t.Error(msg + " (all pixels are white)")
	}
}
