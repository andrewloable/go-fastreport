// Package chart renders multi-series charts as Go images.
// It supports Bar, Line, Area, and Pie chart types with axes,
// gridlines, and a legend. No external dependencies are required —
// all rendering uses the standard image/draw/color packages.
package chart

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// ── Public types ──────────────────────────────────────────────────────────────

// SeriesType identifies how a data series is rendered.
type SeriesType int

const (
	SeriesTypeLine     SeriesType = iota // connected polyline
	SeriesTypeBar                        // vertical bars (grouped)
	SeriesTypeArea                       // filled area under the line
	SeriesTypePie                        // pie/donut slice (first series only for full pie)
	SeriesTypeSpline                     // smooth curve (cubic Bezier approximation)
	SeriesTypeDoughnut                   // doughnut (ring) chart
)

// Series is a single data series to be plotted.
type Series struct {
	// Name is displayed in the legend.
	Name string
	// Type overrides the chart-level type for this series.
	// If SeriesTypeLine (0), the chart-level Type is used.
	Type SeriesType
	// Values is the Y-axis data.
	Values []float64
	// Labels is optional per-point category labels (X-axis).
	Labels []string
	// Color is the series color. Zero value → auto-assigned from palette.
	Color color.RGBA
}

// Chart holds the full chart definition.
type Chart struct {
	// Title is drawn above the chart area.
	Title string
	// Type is the default series type (used when a series has SeriesTypeLine == 0 and no override).
	Type SeriesType
	// Series is the list of data series.
	Series []Series
	// Background fill color (default white).
	Background color.RGBA
	// Width / Height are output dimensions in pixels.
	Width, Height int
	// ShowLegend enables the legend panel on the right.
	ShowLegend bool
	// ShowGrid enables horizontal grid lines.
	ShowGrid bool
	// ShowAxes enables axis lines and tick marks.
	ShowAxes bool
}

// defaultPalette is the series color palette used when Color is zero.
var defaultPalette = []color.RGBA{
	{R: 0x00, G: 0x70, B: 0xC0, A: 255}, // blue
	{R: 0xC0, G: 0x00, B: 0x00, A: 255}, // red
	{R: 0x00, G: 0x80, B: 0x00, A: 255}, // green
	{R: 0xFF, G: 0x80, B: 0x00, A: 255}, // orange
	{R: 0x70, G: 0x00, B: 0x80, A: 255}, // purple
	{R: 0x00, G: 0x80, B: 0x80, A: 255}, // teal
	{R: 0x80, G: 0x40, B: 0x00, A: 255}, // brown
	{R: 0xFF, G: 0xC0, B: 0x00, A: 255}, // gold
}

// piePalette used when rendering pie charts.
var piePalette = []color.RGBA{
	{R: 0x00, G: 0x70, B: 0xC0, A: 255},
	{R: 0xC0, G: 0x00, B: 0x00, A: 255},
	{R: 0x00, G: 0x80, B: 0x00, A: 255},
	{R: 0xFF, G: 0x80, B: 0x00, A: 255},
	{R: 0x70, G: 0x00, B: 0x80, A: 255},
	{R: 0x00, G: 0x80, B: 0x80, A: 255},
	{R: 0x80, G: 0x40, B: 0x00, A: 255},
	{R: 0xFF, G: 0xC0, B: 0x00, A: 255},
}

// Render produces an image for the chart. Returns nil if width/height ≤ 0.
func (c *Chart) Render() image.Image {
	w, h := c.Width, c.Height
	if w <= 0 {
		w = 400
	}
	if h <= 0 {
		h = 300
	}

	// Fill background.
	bg := c.Background
	if bg.A == 0 {
		bg = color.RGBA{255, 255, 255, 255}
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)

	if len(c.Series) == 0 {
		return img
	}

	// Assign default colors to series without colors.
	for i := range c.Series {
		if c.Series[i].Color == (color.RGBA{}) {
			c.Series[i].Color = defaultPalette[i%len(defaultPalette)]
		}
		if c.Series[i].Type == 0 && c.Type != 0 {
			c.Series[i].Type = c.Type
		}
	}

	// Pie / Doughnut chart: special layout.
	if c.Type == SeriesTypePie || c.Type == SeriesTypeDoughnut ||
		(len(c.Series) > 0 && (c.Series[0].Type == SeriesTypePie || c.Series[0].Type == SeriesTypeDoughnut)) {
		isDoughnut := c.Type == SeriesTypeDoughnut ||
			(len(c.Series) > 0 && c.Series[0].Type == SeriesTypeDoughnut)
		c.renderPieDoughnut(img, w, h, isDoughnut)
		return img
	}

	// ── Cartesian chart layout ────────────────────────────────────────────────
	const (
		padLeft   = 45 // space for Y-axis labels
		padBottom = 25 // space for X-axis labels
		padTop    = 30 // space for title
		padRight  = 10
		legendW   = 90 // width of legend panel (when ShowLegend)
	)

	chartRight := w - padRight
	if c.ShowLegend {
		chartRight = w - legendW - padRight
	}
	chartLeft := padLeft
	chartTop := padTop
	chartBottom := h - padBottom

	areaW := chartRight - chartLeft
	areaH := chartBottom - chartTop
	if areaW <= 0 || areaH <= 0 {
		return img
	}

	// Find global Y range across all series.
	minY, maxY := globalRange(c.Series)
	if minY == maxY {
		minY = 0
		if maxY == 0 {
			maxY = 1
		}
	}
	// Extend range to include 0 for bar/area charts.
	if c.Type != SeriesTypeLine && c.Type != SeriesTypeSpline {
		if minY > 0 {
			minY = 0
		}
	}

	// Scale helpers.
	yScale := func(v float64) int {
		t := (v - minY) / (maxY - minY)
		return chartBottom - int(t*float64(areaH))
	}

	// Draw title.
	if c.Title != "" {
		drawLabel(img, c.Title, w/2, padTop/2, color.RGBA{0, 0, 0, 255})
	}

	// Draw axes.
	axisColor := color.RGBA{180, 180, 180, 255}
	gridColor := color.RGBA{230, 230, 230, 255}
	black := color.RGBA{0, 0, 0, 255}

	if c.ShowAxes {
		// Y-axis.
		drawVLine(img, chartLeft, chartTop, chartBottom, black)
		// X-axis (at y=0 or bottom).
		zeroY := yScale(0)
		if zeroY < chartTop {
			zeroY = chartBottom
		}
		drawHLine(img, chartLeft, chartRight, zeroY, black)
	}

	// Draw grid lines + Y-axis tick labels.
	const nTicks = 5
	for i := 0; i <= nTicks; i++ {
		v := minY + (maxY-minY)*float64(i)/float64(nTicks)
		gy := yScale(v)
		if c.ShowGrid && gy > chartTop && gy < chartBottom {
			drawHLine(img, chartLeft+1, chartRight, gy, gridColor)
		}
		if c.ShowAxes {
			drawHLine(img, chartLeft-3, chartLeft, gy, black)
		}
		_ = axisColor
	}

	// Find max point count across series.
	nPts := 0
	for _, s := range c.Series {
		if len(s.Values) > nPts {
			nPts = len(s.Values)
		}
	}
	if nPts == 0 {
		return img
	}

	// Render each series.
	nSeries := len(c.Series)
	for si, s := range c.Series {
		n := len(s.Values)
		if n == 0 {
			continue
		}

		sType := s.Type
		if sType == 0 {
			sType = c.Type
		}

		switch sType {
		case SeriesTypeBar:
			// Grouped bars: each category gets a slot; series bars share that slot.
			slotW := float64(areaW) / float64(nPts)
			barW := int(slotW / float64(nSeries) * 0.8)
			if barW < 1 {
				barW = 1
			}
			for j, v := range s.Values {
				slotX := chartLeft + int(float64(j)*slotW)
				barX := slotX + int(slotW/float64(nSeries)*float64(si)) + int(slotW/float64(nSeries)*0.1)
				top := yScale(v)
				bot := yScale(0)
				if top > bot {
					top, bot = bot, top
				}
				fillRect(img, barX, top, barX+barW, bot, s.Color)
			}

		case SeriesTypeArea:
			step := float64(areaW) / float64(nPts-1)
			if nPts == 1 {
				step = float64(areaW)
			}
			var prevX, prevY int
			zeroY := yScale(0)
			for j, v := range s.Values {
				x := chartLeft + int(float64(j)*step)
				y := yScale(v)
				if j > 0 {
					// Fill trapezoid from line to zero axis.
					fillTrapezoid(img, prevX, prevY, x, y, zeroY, s.Color, 180)
					drawThickLine(img, prevX, prevY, x, y, s.Color, 2)
				}
				prevX, prevY = x, y
			}

		case SeriesTypeSpline:
			// Spline: smooth curve using cubic Bezier approximation.
			// Control points are chosen using the Catmull-Rom convention:
			// for each interior point the tangent is parallel to the chord
			// between its neighbours, scaled by 1/3 of the segment length.
			step := float64(areaW) / float64(nPts-1)
			if nPts == 1 {
				step = float64(areaW)
			}
			// Build pixel coordinates for all points.
			xs := make([]int, n)
			ys := make([]int, n)
			for j, v := range s.Values {
				xs[j] = chartLeft + int(float64(j)*step)
				ys[j] = yScale(v)
				// Draw point marker.
				fillRect(img, xs[j]-2, ys[j]-2, xs[j]+2, ys[j]+2, s.Color)
			}
			// Draw smooth segments.
			for j := 1; j < n; j++ {
				drawBezierSegment(img, xs, ys, j-1, j, s.Color)
			}

		default: // Line
			step := float64(areaW) / float64(nPts-1)
			if nPts == 1 {
				step = float64(areaW)
			}
			var prevX, prevY int
			for j, v := range s.Values {
				x := chartLeft + int(float64(j)*step)
				y := yScale(v)
				if j > 0 {
					drawThickLine(img, prevX, prevY, x, y, s.Color, 2)
				}
				// Draw point marker.
				fillRect(img, x-2, y-2, x+2, y+2, s.Color)
				prevX, prevY = x, y
			}
		}
	}

	// Draw legend.
	if c.ShowLegend {
		lx := chartRight + 8
		ly := chartTop
		for _, s := range c.Series {
			fillRect(img, lx, ly+2, lx+12, ly+12, s.Color)
			drawLabel(img, s.Name, lx+15, ly+7, black)
			ly += 16
		}
	}

	return img
}

// ── Pie / Doughnut chart ──────────────────────────────────────────────────────

// renderPieDoughnut renders a pie chart, or if isDoughnut is true, punches a
// hole in the centre to produce a doughnut (ring) chart.
func (c *Chart) renderPieDoughnut(img *image.RGBA, w, h int, isDoughnut bool) {
	if len(c.Series) == 0 || len(c.Series[0].Values) == 0 {
		return
	}
	vals := c.Series[0].Values
	total := 0.0
	for _, v := range vals {
		total += math.Abs(v)
	}
	if total == 0 {
		return
	}

	cx, cy := w/2, h/2
	r := min(cx, cy) - 20
	if r <= 0 {
		return
	}

	// Draw pie sectors.
	startAngle := -math.Pi / 2 // start at top
	for i, v := range vals {
		sweepAngle := 2 * math.Pi * math.Abs(v) / total
		col := piePalette[i%len(piePalette)]
		drawSector(img, cx, cy, r, startAngle, startAngle+sweepAngle, col)
		startAngle += sweepAngle
	}

	// For doughnut: overwrite the inner hole with the background colour.
	if isDoughnut {
		holeR := r / 2
		bg := c.Background
		if bg.A == 0 {
			bg = color.RGBA{255, 255, 255, 255}
		}
		for y := cy - holeR; y <= cy+holeR; y++ {
			for x := cx - holeR; x <= cx+holeR; x++ {
				dx := float64(x - cx)
				dy := float64(y - cy)
				if dx*dx+dy*dy <= float64(holeR*holeR) {
					setPixel(img, x, y, bg)
				}
			}
		}
		// Redraw hole border.
		for i := 0; i < 360; i++ {
			a := float64(i) * math.Pi / 180
			x := cx + int(math.Cos(a)*float64(holeR))
			y := cy + int(math.Sin(a)*float64(holeR))
			setPixel(img, x, y, color.RGBA{0, 0, 0, 255})
		}
	}

	// Draw title.
	if c.Title != "" {
		drawLabel(img, c.Title, w/2, 12, color.RGBA{0, 0, 0, 255})
	}
}

// drawSector fills a pie sector by scanning pixels within the arc.
func drawSector(img *image.RGBA, cx, cy, r int, startAngle, endAngle float64, col color.RGBA) {
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > float64(r) {
				continue
			}
			angle := math.Atan2(dy, dx)
			// Normalize angle to [startAngle, startAngle + 2π).
			for angle < startAngle {
				angle += 2 * math.Pi
			}
			if angle <= endAngle || (endAngle-startAngle >= 2*math.Pi-1e-9) {
				setPixel(img, x, y, col)
			}
		}
	}
	// Draw sector border.
	n := int((endAngle - startAngle) / (2 * math.Pi) * float64(2*math.Pi*float64(r)) / 1.0)
	if n < 4 {
		n = 4
	}
	for i := 0; i <= n; i++ {
		a := startAngle + (endAngle-startAngle)*float64(i)/float64(n)
		x := cx + int(math.Cos(a)*float64(r))
		y := cy + int(math.Sin(a)*float64(r))
		setPixel(img, x, y, color.RGBA{0, 0, 0, 255})
	}
	// Radial lines.
	for dist := 0; dist <= r; dist++ {
		x := cx + int(math.Cos(startAngle)*float64(dist))
		y := cy + int(math.Sin(startAngle)*float64(dist))
		setPixel(img, x, y, color.RGBA{0, 0, 0, 255})
	}
}

// ── Spline drawing ────────────────────────────────────────────────────────────

// drawBezierSegment draws a smooth cubic Bezier curve between points[i] and
// points[i+1] using Catmull-Rom tangents. xs and ys hold the full array of
// pixel coordinates; i0 and i1 are the indices of the two endpoints.
func drawBezierSegment(img *image.RGBA, xs, ys []int, i0, i1 int, col color.RGBA) {
	n := len(xs)
	// Catmull-Rom: tangent at point k = (P[k+1] - P[k-1]) / 2.
	// For endpoints, we clamp to a one-sided difference.
	tangentX := func(k int) float64 {
		if k <= 0 {
			return float64(xs[1] - xs[0])
		}
		if k >= n-1 {
			return float64(xs[n-1] - xs[n-2])
		}
		return float64(xs[k+1]-xs[k-1]) / 2
	}
	tangentY := func(k int) float64 {
		if k <= 0 {
			return float64(ys[1] - ys[0])
		}
		if k >= n-1 {
			return float64(ys[n-1] - ys[n-2])
		}
		return float64(ys[k+1]-ys[k-1]) / 2
	}

	x0, y0 := float64(xs[i0]), float64(ys[i0])
	x3, y3 := float64(xs[i1]), float64(ys[i1])

	// Control points (1/3 of tangent length).
	tx0, ty0 := tangentX(i0)/3, tangentY(i0)/3
	tx1, ty1 := tangentX(i1)/3, tangentY(i1)/3

	x1, y1 := x0+tx0, y0+ty0
	x2, y2 := x3-tx1, y3-ty1

	// Evaluate the Bezier at enough steps for a smooth curve.
	dx := math.Abs(x3 - x0)
	dy := math.Abs(y3 - y0)
	steps := int(math.Sqrt(dx*dx+dy*dy)) * 2
	if steps < 8 {
		steps = 8
	}

	prevX, prevY := int(x0), int(y0)
	for s := 1; s <= steps; s++ {
		t := float64(s) / float64(steps)
		u := 1 - t
		// Cubic Bezier formula.
		bx := u*u*u*x0 + 3*u*u*t*x1 + 3*u*t*t*x2 + t*t*t*x3
		by := u*u*u*y0 + 3*u*u*t*y1 + 3*u*t*t*y2 + t*t*t*y3
		cx, cy := int(bx), int(by)
		drawThickLine(img, prevX, prevY, cx, cy, col, 2)
		prevX, prevY = cx, cy
	}
}

// ── Drawing primitives ────────────────────────────────────────────────────────

func setPixel(img *image.RGBA, x, y int, c color.RGBA) {
	b := img.Bounds()
	if x >= b.Min.X && y >= b.Min.Y && x < b.Max.X && y < b.Max.Y {
		img.SetRGBA(x, y, c)
	}
}

func drawHLine(img *image.RGBA, x0, x1, y int, c color.RGBA) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	for x := x0; x <= x1; x++ {
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

func fillRect(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	b := img.Bounds()
	if x0 < b.Min.X {
		x0 = b.Min.X
	}
	if y0 < b.Min.Y {
		y0 = b.Min.Y
	}
	if x1 >= b.Max.X {
		x1 = b.Max.X - 1
	}
	if y1 >= b.Max.Y {
		y1 = b.Max.Y - 1
	}
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

// drawThickLine draws a 1-px-wide Bresenham line, then inflates by radius.
func drawThickLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA, thickness int) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	for {
		fillRect(img, x0-thickness/2, y0-thickness/2, x0+thickness/2, y0+thickness/2, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// fillTrapezoid fills the region between a line segment and a horizontal base.
func fillTrapezoid(img *image.RGBA, x0, y0, x1, y1, baseY int, col color.RGBA, alpha uint8) {
	c := color.RGBA{col.R, col.G, col.B, alpha}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}
	for x := x0; x <= x1; x++ {
		t := float64(x-x0) / float64(max(x1-x0, 1))
		lineY := y0 + int(t*float64(y1-y0))
		if lineY > baseY {
			lineY, baseY = baseY, lineY
		}
		for y := lineY; y <= baseY; y++ {
			setPixel(img, x, y, c)
		}
		if lineY < baseY {
			baseY = lineY // restore original baseY for next iteration
		}
	}
}

// drawLabel draws a simple text label using a 5x7 mini bitmap font.
func drawLabel(img *image.RGBA, text string, cx, cy int, col color.RGBA) {
	// Render left-aligned from (cx - len*3, cy-3).
	charW := 4
	totalW := len(text) * charW
	x := cx - totalW/2
	y := cy - 3
	for _, ch := range text {
		drawChar(img, x, y, ch, col)
		x += charW
	}
}

// drawChar renders a single ASCII character from a 3x5 bitmap font.
func drawChar(img *image.RGBA, x, y int, ch rune, col color.RGBA) {
	bm, ok := miniFont[ch]
	if !ok {
		bm = miniFont[' ']
	}
	for row := 0; row < 5; row++ {
		bits := bm[row]
		for col2 := 0; col2 < 3; col2++ {
			if bits&(1<<uint(2-col2)) != 0 {
				setPixel(img, x+col2, y+row, col)
			}
		}
	}
}

// globalRange returns the min and max Y values across all series.
func globalRange(series []Series) (float64, float64) {
	if len(series) == 0 {
		return 0, 1
	}
	minY, maxY := math.Inf(1), math.Inf(-1)
	for _, s := range series {
		for _, v := range s.Values {
			if v < minY {
				minY = v
			}
			if v > maxY {
				maxY = v
			}
		}
	}
	if math.IsInf(minY, 1) {
		minY = 0
	}
	if math.IsInf(maxY, -1) {
		maxY = 1
	}
	return minY, maxY
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ── Mini bitmap font (3×5 pixels per glyph) ──────────────────────────────────

// Each entry is 5 rows of 3 bits (MSB = left column).
var miniFont = map[rune][5]byte{
	' ':  {0b000, 0b000, 0b000, 0b000, 0b000},
	'0':  {0b111, 0b101, 0b101, 0b101, 0b111},
	'1':  {0b010, 0b110, 0b010, 0b010, 0b111},
	'2':  {0b111, 0b001, 0b111, 0b100, 0b111},
	'3':  {0b111, 0b001, 0b111, 0b001, 0b111},
	'4':  {0b101, 0b101, 0b111, 0b001, 0b001},
	'5':  {0b111, 0b100, 0b111, 0b001, 0b111},
	'6':  {0b111, 0b100, 0b111, 0b101, 0b111},
	'7':  {0b111, 0b001, 0b001, 0b001, 0b001},
	'8':  {0b111, 0b101, 0b111, 0b101, 0b111},
	'9':  {0b111, 0b101, 0b111, 0b001, 0b111},
	'A':  {0b010, 0b101, 0b111, 0b101, 0b101},
	'B':  {0b110, 0b101, 0b110, 0b101, 0b110},
	'C':  {0b011, 0b100, 0b100, 0b100, 0b011},
	'D':  {0b110, 0b101, 0b101, 0b101, 0b110},
	'E':  {0b111, 0b100, 0b111, 0b100, 0b111},
	'F':  {0b111, 0b100, 0b111, 0b100, 0b100},
	'G':  {0b011, 0b100, 0b101, 0b101, 0b111},
	'H':  {0b101, 0b101, 0b111, 0b101, 0b101},
	'I':  {0b111, 0b010, 0b010, 0b010, 0b111},
	'J':  {0b001, 0b001, 0b001, 0b101, 0b111},
	'K':  {0b101, 0b101, 0b110, 0b101, 0b101},
	'L':  {0b100, 0b100, 0b100, 0b100, 0b111},
	'M':  {0b101, 0b111, 0b101, 0b101, 0b101},
	'N':  {0b101, 0b111, 0b111, 0b101, 0b101},
	'O':  {0b010, 0b101, 0b101, 0b101, 0b010},
	'P':  {0b110, 0b101, 0b110, 0b100, 0b100},
	'Q':  {0b010, 0b101, 0b101, 0b111, 0b011},
	'R':  {0b110, 0b101, 0b110, 0b101, 0b101},
	'S':  {0b011, 0b100, 0b010, 0b001, 0b110},
	'T':  {0b111, 0b010, 0b010, 0b010, 0b010},
	'U':  {0b101, 0b101, 0b101, 0b101, 0b111},
	'V':  {0b101, 0b101, 0b101, 0b010, 0b010},
	'W':  {0b101, 0b101, 0b101, 0b111, 0b101},
	'X':  {0b101, 0b101, 0b010, 0b101, 0b101},
	'Y':  {0b101, 0b101, 0b010, 0b010, 0b010},
	'Z':  {0b111, 0b001, 0b010, 0b100, 0b111},
	'a':  {0b000, 0b111, 0b101, 0b101, 0b111},
	'b':  {0b100, 0b110, 0b101, 0b101, 0b110},
	'c':  {0b000, 0b011, 0b100, 0b100, 0b011},
	'd':  {0b001, 0b011, 0b101, 0b101, 0b011},
	'e':  {0b000, 0b111, 0b101, 0b110, 0b011},
	'f':  {0b001, 0b010, 0b111, 0b010, 0b010},
	'g':  {0b000, 0b111, 0b101, 0b111, 0b001},
	'h':  {0b100, 0b110, 0b101, 0b101, 0b101},
	'i':  {0b010, 0b000, 0b010, 0b010, 0b010},
	'j':  {0b001, 0b000, 0b001, 0b101, 0b010},
	'k':  {0b100, 0b101, 0b110, 0b101, 0b101},
	'l':  {0b110, 0b010, 0b010, 0b010, 0b111},
	'm':  {0b000, 0b101, 0b111, 0b101, 0b101},
	'n':  {0b000, 0b110, 0b101, 0b101, 0b101},
	'o':  {0b000, 0b010, 0b101, 0b101, 0b010},
	'p':  {0b000, 0b110, 0b101, 0b110, 0b100},
	'q':  {0b000, 0b011, 0b101, 0b011, 0b001},
	'r':  {0b000, 0b011, 0b100, 0b100, 0b100},
	's':  {0b000, 0b011, 0b110, 0b001, 0b110},
	't':  {0b010, 0b111, 0b010, 0b010, 0b001},
	'u':  {0b000, 0b101, 0b101, 0b101, 0b011},
	'v':  {0b000, 0b101, 0b101, 0b101, 0b010},
	'w':  {0b000, 0b101, 0b101, 0b111, 0b101},
	'x':  {0b000, 0b101, 0b010, 0b101, 0b101},
	'y':  {0b000, 0b101, 0b111, 0b001, 0b110},
	'z':  {0b000, 0b111, 0b010, 0b100, 0b111},
	'.':  {0b000, 0b000, 0b000, 0b000, 0b010},
	',':  {0b000, 0b000, 0b000, 0b010, 0b100},
	'-':  {0b000, 0b000, 0b111, 0b000, 0b000},
	'%':  {0b101, 0b001, 0b010, 0b100, 0b101},
	':':  {0b000, 0b010, 0b000, 0b010, 0b000},
}
