// Package sparkline renders small inline charts (sparklines) from chart
// configuration data stored in FastReport FRX files.
//
// The ChartData field of a SparklineObject is a base64-encoded XML document
// whose structure mirrors the MSChart / DevExpress chart format used by the
// .NET engine.  This package decodes that XML, extracts the data points, and
// renders a miniature chart as a Go image.Image.
//
// Supported chart types:
//   - Line / FastLine  — connected line plot
//   - Area             — filled area under the line
//   - Column / Bar     — vertical bars
//   - WinLoss          — positive/negative bars (±1)
package sparkline

import (
	"encoding/base64"
	"encoding/xml"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"
)

// ── XML types for ChartData ────────────────────────────────────────────────

type chartXML struct {
	Series chartSeriesCollection `xml:"Series"`
}

type chartSeriesCollection struct {
	Items []seriesXML `xml:"Series"`
}

type seriesXML struct {
	Name      string           `xml:"Name,attr"`
	ChartType string           `xml:"ChartType,attr"`
	Points    pointsCollection `xml:"Points"`
}

type pointsCollection struct {
	Items []dataPointXML `xml:"DataPoint"`
}

type dataPointXML struct {
	YValues string `xml:"YValues,attr"`
}

// ── Public types ──────────────────────────────────────────────────────────

// ChartType identifies the style of sparkline chart.
type ChartType int

const (
	ChartTypeLine    ChartType = iota // connected line
	ChartTypeArea                     // filled area
	ChartTypeColumn                   // vertical bars
	ChartTypeWinLoss                  // positive/negative indicator bars
)

// Series is a decoded data series ready for rendering.
type Series struct {
	Type   ChartType
	Values []float64
}

// ── Decoding ─────────────────────────────────────────────────────────────

// DecodeChartData decodes the base64-encoded ChartData string from an FRX
// SparklineObject and returns the first data series found, or nil if the data
// is empty or cannot be parsed.
func DecodeChartData(chartData string) *Series {
	if chartData == "" {
		return nil
	}
	xmlBytes, err := base64.StdEncoding.DecodeString(chartData)
	if err != nil {
		// Try raw XML (not base64).
		xmlBytes = []byte(chartData)
	}
	var ch chartXML
	if err := xml.Unmarshal(xmlBytes, &ch); err != nil {
		return nil
	}
	if len(ch.Series.Items) == 0 {
		return nil
	}
	s := ch.Series.Items[0]
	ct := parseChartType(s.ChartType)
	values := make([]float64, 0, len(s.Points.Items))
	for _, pt := range s.Points.Items {
		// YValues may be comma-separated; take the first.
		raw := strings.SplitN(pt.YValues, ",", 2)[0]
		var v float64
		if raw != "" {
			for _, c := range raw {
				if (c >= '0' && c <= '9') || c == '.' || c == '-' {
					v = v*10 + float64(c-'0')
				}
			}
			// Simple parse: use strconv-style float parsing via a helper.
			v = parseFloat(raw)
		}
		values = append(values, v)
	}
	return &Series{Type: ct, Values: values}
}

func parseChartType(s string) ChartType {
	switch strings.ToLower(s) {
	case "area":
		return ChartTypeArea
	case "column", "bar", "stackedcolumn", "stackedbar":
		return ChartTypeColumn
	case "winloss":
		return ChartTypeWinLoss
	default:
		return ChartTypeLine
	}
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	neg := false
	if len(s) > 0 && s[0] == '-' {
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

// ── Rendering ────────────────────────────────────────────────────────────

var (
	lineColor = color.RGBA{R: 0x00, G: 0x70, B: 0xC0, A: 255} // blue
	barPos    = color.RGBA{R: 0x00, G: 0x70, B: 0xC0, A: 255}
	barNeg    = color.RGBA{R: 0xC0, G: 0x00, B: 0x00, A: 255}
	bgColor   = color.RGBA{R: 255, G: 255, B: 255, A: 255}
)

// Render draws the series as a sparkline chart into an image of the given size.
// Returns nil if the series has no data points.
func Render(s *Series, w, h int) image.Image {
	if s == nil || len(s.Values) == 0 || w <= 0 || h <= 0 {
		return nil
	}
	switch s.Type {
	case ChartTypeColumn:
		return renderBars(s.Values, w, h, false)
	case ChartTypeWinLoss:
		return renderWinLoss(s.Values, w, h)
	case ChartTypeArea:
		return renderArea(s.Values, w, h)
	default:
		return renderLine(s.Values, w, h)
	}
}

// padding around the chart area.
const pad = 2

func newImg(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), image.NewUniform(bgColor), image.Point{}, draw.Src)
	return img
}

func setPixel(img *image.RGBA, x, y int, c color.RGBA) {
	if x >= 0 && y >= 0 && x < img.Bounds().Max.X && y < img.Bounds().Max.Y {
		img.SetRGBA(x, y, c)
	}
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := x1 - x0
	dy := y1 - y0
	if dx == 0 && dy == 0 {
		setPixel(img, x0, y0, c)
		return
	}
	steps := int(math.Abs(float64(dx)))
	if int(math.Abs(float64(dy))) > steps {
		steps = int(math.Abs(float64(dy)))
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(math.Round(float64(x0) + t*float64(dx)))
		y := int(math.Round(float64(y0) + t*float64(dy)))
		setPixel(img, x, y, c)
	}
}

func drawVLine(img *image.RGBA, x, y0, y1 int, c color.RGBA) {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for y := y0; y <= y1; y++ {
		setPixel(img, x, y, c)
	}
}

func minMax(vals []float64) (min, max float64) {
	min, max = vals[0], vals[0]
	for _, v := range vals[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return
}

// scaleY maps a data value into the image Y coordinate (top=0).
func scaleY(v, vmin, vmax float64, h int) int {
	ch := float64(h - 2*pad)
	if vmax == vmin {
		return h/2
	}
	frac := (v - vmin) / (vmax - vmin)
	return h - pad - int(math.Round(frac*ch))
}

// scaleX maps the i-th point (out of n total) to an X coordinate.
func scaleX(i, n, w int) int {
	if n <= 1 {
		return w / 2
	}
	cw := float64(w - 2*pad)
	return pad + int(math.Round(float64(i)*cw/float64(n-1)))
}

func renderLine(vals []float64, w, h int) image.Image {
	img := newImg(w, h)
	vmin, vmax := minMax(vals)
	if vmax == vmin {
		vmax = vmin + 1
	}
	n := len(vals)
	for i := 1; i < n; i++ {
		x0 := scaleX(i-1, n, w)
		y0 := scaleY(vals[i-1], vmin, vmax, h)
		x1 := scaleX(i, n, w)
		y1 := scaleY(vals[i], vmin, vmax, h)
		drawLine(img, x0, y0, x1, y1, lineColor)
	}
	return img
}

func renderArea(vals []float64, w, h int) image.Image {
	img := newImg(w, h)
	vmin, vmax := minMax(vals)
	if vmax == vmin {
		vmax = vmin + 1
	}
	if vmin > 0 {
		vmin = 0
	}
	n := len(vals)
	baseline := scaleY(0, vmin, vmax, h)
	for i := 1; i < n; i++ {
		x0 := scaleX(i-1, n, w)
		y0 := scaleY(vals[i-1], vmin, vmax, h)
		x1 := scaleX(i, n, w)
		y1 := scaleY(vals[i], vmin, vmax, h)
		// Draw the line segment.
		drawLine(img, x0, y0, x1, y1, lineColor)
		// Fill vertically from each x between x0 and x1 to the baseline.
		for x := x0; x <= x1; x++ {
			// Interpolate y at this x.
			t := 0.0
			if x1 != x0 {
				t = float64(x-x0) / float64(x1-x0)
			}
			y := y0 + int(math.Round(t*float64(y1-y0)))
			fill := color.RGBA{R: lineColor.R, G: lineColor.G, B: lineColor.B, A: 80}
			if y <= baseline {
				drawVLine(img, x, y, baseline, fill)
			} else {
				drawVLine(img, x, baseline, y, fill)
			}
		}
	}
	return img
}

func renderBars(vals []float64, w, h int, winloss bool) image.Image {
	img := newImg(w, h)
	vmin, vmax := minMax(vals)
	if vmax == vmin {
		vmax = vmin + 1
	}
	if !winloss {
		if vmin > 0 {
			vmin = 0
		}
		if vmax < 0 {
			vmax = 0
		}
	}
	n := len(vals)
	barW := (w - 2*pad) / n
	if barW < 1 {
		barW = 1
	}
	baseline := scaleY(0, vmin, vmax, h)
	for i, v := range vals {
		x0 := pad + i*(w-2*pad)/n
		x1 := x0 + barW - 1
		y := scaleY(v, vmin, vmax, h)
		c := barPos
		if v < 0 {
			c = barNeg
		}
		for x := x0; x <= x1; x++ {
			if y <= baseline {
				drawVLine(img, x, y, baseline, c)
			} else {
				drawVLine(img, x, baseline, y, c)
			}
		}
	}
	return img
}

func renderWinLoss(vals []float64, w, h int) image.Image {
	// Normalize to +1 / 0 / -1
	normalized := make([]float64, len(vals))
	for i, v := range vals {
		if v > 0 {
			normalized[i] = 1
		} else if v < 0 {
			normalized[i] = -1
		}
	}
	return renderBars(normalized, w, h, true)
}
