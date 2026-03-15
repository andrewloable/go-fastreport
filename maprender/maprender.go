// Package maprender provides a geographic-map placeholder renderer for the
// go-fastreport MapObject component.
//
// Full vector rendering requires ESRI Shapefile data that is not bundled with
// the library.  This package renders a visually recognisable map-style image:
//   - Ocean background
//   - Latitude/longitude graticule grid (every 30°)
//   - Equator and Prime-Meridian highlight lines
//   - A simplified world-outline silhouette drawn from a small embedded
//     polygon dataset (major continental outlines only)
//   - Layer count and palette label in the corner
//
// The output is an image.Image that callers encode to PNG and store in the
// report BlobStore.
package maprender

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Layer carries the minimal information extracted from a MapLayer that affects
// how the map is rendered.
type Layer struct {
	Shapefile string
	Palette   string
	Type      string // "Choropleth", "Bubble", etc.

	// GeoFeatures contains user-supplied geographic features for this layer.
	// When non-nil, the built-in continental outlines are replaced by these.
	GeoFeatures []GeoFeature

	// Bubbles is the list of bubble overlays for "Bubble" type layers.
	Bubbles []Bubble

	// ShowLegend enables a legend for this layer.
	ShowLegend bool
}

// Options control the rendering.
type Options struct {
	Width   int
	Height  int
	OffsetX float64 // pan offset in degrees longitude
	OffsetY float64 // pan offset in degrees latitude
	Zoom    float64 // 1.0 = full world (-180..180 × -90..90)
	Layers  []Layer
}

// Render returns a map-style image for the given options.
// Width and Height default to 400×200 when zero.
func Render(opts Options) image.Image {
	w := opts.Width
	h := opts.Height
	if w <= 0 {
		w = 400
	}
	if h <= 0 {
		h = 200
	}
	if opts.Zoom <= 0 {
		opts.Zoom = 1.0
	}

	img := image.NewNRGBA(image.Rect(0, 0, w, h))

	// ── Background: ocean blue ────────────────────────────────────────────────
	ocean := color.NRGBA{R: 173, G: 216, B: 230, A: 255} // light-blue
	land := color.NRGBA{R: 200, G: 220, B: 160, A: 255}  // light-green
	draw.Draw(img, img.Bounds(), &image.Uniform{ocean}, image.Point{}, draw.Src)

	// Coordinate-to-pixel helpers.
	// Mercator projection clipped to ±85°.
	mercY := func(latDeg float64) float64 {
		lat := latDeg * math.Pi / 180.0
		return math.Log(math.Tan(math.Pi/4 + lat/2))
	}
	topMerc := mercY(85.0)
	botMerc := mercY(-85.0)

	lonToX := func(lon float64) float64 {
		return float64(w) * (lon + 180.0) / 360.0
	}
	latToY := func(lat float64) float64 {
		m := mercY(lat)
		return float64(h) * (topMerc - m) / (topMerc - botMerc)
	}

	// ── Draw continental outline polygons ────────────────────────────────────
	drawPoly := func(pts [][2]float64, fill color.NRGBA) {
		if len(pts) < 3 {
			return
		}
		// Find bounding box in pixel space.
		minX, minY := math.MaxFloat64, math.MaxFloat64
		maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
		px := make([]float64, len(pts))
		py := make([]float64, len(pts))
		for i, p := range pts {
			px[i] = lonToX(p[0])
			py[i] = latToY(p[1])
			if px[i] < minX {
				minX = px[i]
			}
			if px[i] > maxX {
				maxX = px[i]
			}
			if py[i] < minY {
				minY = py[i]
			}
			if py[i] > maxY {
				maxY = py[i]
			}
		}
		// Scanline fill.
		y0 := int(math.Max(0, minY))
		y1 := int(math.Min(float64(h-1), maxY))
		for iy := y0; iy <= y1; iy++ {
			fy := float64(iy)
			var xs []float64
			n := len(px)
			for i := 0; i < n; i++ {
				j := (i + 1) % n
				if (py[i] <= fy && py[j] > fy) || (py[j] <= fy && py[i] > fy) {
					t := (fy - py[i]) / (py[j] - py[i])
					xs = append(xs, px[i]+t*(px[j]-px[i]))
				}
			}
			// Sort xs (bubble for small slices).
			for a := 0; a < len(xs); a++ {
				for b := a + 1; b < len(xs); b++ {
					if xs[b] < xs[a] {
						xs[a], xs[b] = xs[b], xs[a]
					}
				}
			}
			for k := 0; k+1 < len(xs); k += 2 {
				x0 := int(math.Max(0, xs[k]))
				x1 := int(math.Min(float64(w-1), xs[k+1]))
				for ix := x0; ix <= x1; ix++ {
					img.SetNRGBA(ix, iy, fill)
				}
			}
		}
	}

	// Collect all user-supplied GeoFeatures across all layers.
	hasUserFeatures := false
	for _, l := range opts.Layers {
		if len(l.GeoFeatures) > 0 {
			hasUserFeatures = true
			break
		}
	}

	if hasUserFeatures {
		// Render user-supplied GeoFeatures with choropleth coloring.
		for _, layer := range opts.Layers {
			if len(layer.GeoFeatures) == 0 {
				continue
			}
			pal := NamedPalette(layer.Palette)

			// Find value range for normalization.
			minVal, maxVal := math.Inf(1), math.Inf(-1)
			for _, f := range layer.GeoFeatures {
				if f.Value < minVal {
					minVal = f.Value
				}
				if f.Value > maxVal {
					maxVal = f.Value
				}
			}
			valRange := maxVal - minVal
			if valRange == 0 {
				valRange = 1
			}

			for _, feat := range layer.GeoFeatures {
				t := (feat.Value - minVal) / valRange
				fc := pal.Color(t)
				nfc := color.NRGBA{fc.R, fc.G, fc.B, fc.A}
				for _, poly := range feat.Polygons {
					drawPoly(poly, nfc)
				}
			}
		}
	} else {
		// Draw simplified continental outlines (major blobs only).
		for _, poly := range continents {
			drawPoly(poly, land)
		}
	}

	// ── Graticule ─────────────────────────────────────────────────────────────
	gridColor := color.NRGBA{R: 100, G: 150, B: 200, A: 100}
	equatorColor := color.NRGBA{R: 60, G: 100, B: 180, A: 200}

	setPixel := func(x, y int, c color.NRGBA) {
		if x >= 0 && x < w && y >= 0 && y < h {
			img.SetNRGBA(x, y, c)
		}
	}

	// Longitude lines every 30°.
	for lon := -180.0; lon <= 180.0; lon += 30.0 {
		px := int(lonToX(lon))
		for iy := 0; iy < h; iy++ {
			setPixel(px, iy, gridColor)
		}
	}
	// Latitude lines every 30°.
	for lat := -90.0; lat <= 90.0; lat += 30.0 {
		py := int(latToY(lat))
		c := gridColor
		if lat == 0 {
			c = equatorColor
		}
		for ix := 0; ix < w; ix++ {
			setPixel(ix, py, c)
		}
	}
	// Prime meridian highlight.
	pPM := int(lonToX(0))
	for iy := 0; iy < h; iy++ {
		setPixel(pPM, iy, equatorColor)
	}

	// ── Bubble overlays ───────────────────────────────────────────────────────
	for _, layer := range opts.Layers {
		if layer.Type != "Bubble" || len(layer.Bubbles) == 0 {
			continue
		}
		// Find max value for radius scaling.
		maxVal := 0.0
		for _, b := range layer.Bubbles {
			if b.Value > maxVal {
				maxVal = b.Value
			}
		}
		if maxVal == 0 {
			maxVal = 1
		}
		maxRadius := math.Min(float64(w), float64(h)) * 0.12
		for _, bubble := range layer.Bubbles {
			bx := int(lonToX(bubble.Lon))
			by := int(latToY(bubble.Lat))
			r := int(math.Sqrt(bubble.Value/maxVal) * maxRadius)
			if r < 2 {
				r = 2
			}
			bc := bubble.Color
			if bc.A == 0 {
				bc = color.NRGBA{220, 80, 80, 180}
			}
			fillCircle(img, bx, by, r, bc)
			// Outline.
			outlineCircle(img, bx, by, r, color.NRGBA{bc.R / 2, bc.G / 2, bc.B / 2, 220})
		}
	}

	// ── Legend ────────────────────────────────────────────────────────────────
	for _, layer := range opts.Layers {
		if !layer.ShowLegend {
			continue
		}
		pal := NamedPalette(layer.Palette)
		drawLegend(img, w-60, 10, 50, 100, pal, layer.Palette)
		break // one legend
	}

	// ── Layer info label ─────────────────────────────────────────────────────
	// Draw a small text banner at the top-left showing layer count + palette.
	labelColor := color.NRGBA{R: 40, G: 40, B: 40, A: 220}
	bgColor := color.NRGBA{R: 255, G: 255, B: 255, A: 180}

	palette := ""
	for _, l := range opts.Layers {
		if l.Palette != "" {
			palette = l.Palette
			break
		}
	}
	label := fmt.Sprintf("Map (%d layer", len(opts.Layers))
	if len(opts.Layers) != 1 {
		label += "s"
	}
	label += ")"
	if palette != "" {
		label += " · " + palette
	}

	// Draw a simple pixel-font banner (3×5 pixel letters, rudimentary).
	drawBanner(img, 4, 4, label, labelColor, bgColor)

	return img
}

// drawBanner writes a minimal banner rectangle with a text label at (x,y).
// Uses a very small 3×5 pixel-art font for digits, letters, and common chars.
func drawBanner(img *image.NRGBA, x, y int, text string, fg, bg color.NRGBA) {
	charW := 4
	charH := 6
	pad := 2
	tw := len(text)*charW + pad*2
	th := charH + pad*2

	// Background rectangle.
	for dy := 0; dy < th; dy++ {
		for dx := 0; dx < tw; dx++ {
			px := x + dx
			py := y + dy
			if px >= 0 && px < img.Bounds().Max.X && py >= 0 && py < img.Bounds().Max.Y {
				img.SetNRGBA(px, py, bg)
			}
		}
	}

	// Draw each character using a 3×5 pixel bitmap.
	cx := x + pad
	cy := y + pad
	for _, ch := range text {
		bm := charBitmap(ch)
		for row := 0; row < 5; row++ {
			for col := 0; col < 3; col++ {
				if bm[row]>>uint(2-col)&1 == 1 {
					px := cx + col
					py := cy + row
					if px >= 0 && px < img.Bounds().Max.X && py >= 0 && py < img.Bounds().Max.Y {
						img.SetNRGBA(px, py, fg)
					}
				}
			}
		}
		cx += charW
	}
}

// charBitmap returns a 5-row, 3-bit-wide bitmap for a character.
// Each element is a byte where bits 2..0 correspond to columns 0..2.
func charBitmap(ch rune) [5]byte {
	switch ch {
	case 'M':
		return [5]byte{0b101, 0b111, 0b101, 0b101, 0b101}
	case 'a':
		return [5]byte{0b010, 0b001, 0b011, 0b101, 0b011}
	case 'p':
		return [5]byte{0b110, 0b101, 0b110, 0b100, 0b100}
	case '(':
		return [5]byte{0b010, 0b100, 0b100, 0b100, 0b010}
	case ')':
		return [5]byte{0b100, 0b010, 0b010, 0b010, 0b100}
	case 'l', '1':
		return [5]byte{0b110, 0b010, 0b010, 0b010, 0b111}
	case 'e':
		return [5]byte{0b010, 0b101, 0b111, 0b100, 0b011}
	case 'r':
		return [5]byte{0b000, 0b110, 0b101, 0b100, 0b100}
	case 's':
		return [5]byte{0b011, 0b100, 0b010, 0b001, 0b110}
	case 'y':
		return [5]byte{0b101, 0b101, 0b011, 0b001, 0b110}
	case 'g':
		return [5]byte{0b011, 0b100, 0b111, 0b101, 0b011}
	case 'n':
		return [5]byte{0b000, 0b110, 0b101, 0b101, 0b101}
	case 'i':
		return [5]byte{0b111, 0b010, 0b010, 0b010, 0b111}
	case 'o', '0':
		return [5]byte{0b010, 0b101, 0b101, 0b101, 0b010}
	case '2':
		return [5]byte{0b110, 0b001, 0b010, 0b100, 0b111}
	case '3':
		return [5]byte{0b110, 0b001, 0b110, 0b001, 0b110}
	case '4':
		return [5]byte{0b101, 0b101, 0b111, 0b001, 0b001}
	case '5':
		return [5]byte{0b111, 0b100, 0b110, 0b001, 0b110}
	case '·', '.':
		return [5]byte{0b000, 0b000, 0b000, 0b000, 0b010}
	case ' ':
		return [5]byte{}
	default:
		return [5]byte{0b111, 0b101, 0b111, 0b101, 0b111}
	}
}

// fillCircle fills a circle at (cx,cy) with radius r.
func fillCircle(img *image.NRGBA, cx, cy, r int, col color.NRGBA) {
	b := img.Bounds()
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				px, py := cx+dx, cy+dy
				if px >= b.Min.X && py >= b.Min.Y && px < b.Max.X && py < b.Max.Y {
					img.SetNRGBA(px, py, col)
				}
			}
		}
	}
}

// outlineCircle draws a 1-pixel circle outline using midpoint algorithm.
func outlineCircle(img *image.NRGBA, cx, cy, r int, col color.NRGBA) {
	b := img.Bounds()
	set := func(x, y int) {
		if x >= b.Min.X && y >= b.Min.Y && x < b.Max.X && y < b.Max.Y {
			img.SetNRGBA(x, y, col)
		}
	}
	x, y := r, 0
	for x >= y {
		set(cx+x, cy+y); set(cx-x, cy+y)
		set(cx+x, cy-y); set(cx-x, cy-y)
		set(cx+y, cy+x); set(cx-y, cy+x)
		set(cx+y, cy-x); set(cx-y, cy-x)
		y++
		if x*x+y*y > r*r {
			x--
		}
	}
}

// drawLegend renders a vertical color gradient legend bar at (x,y) with size (w,h).
func drawLegend(img *image.NRGBA, x, y, w, h int, pal ChoroplethPalette, title string) {
	bounds := img.Bounds()
	// Gradient bar.
	for dy := 0; dy < h; dy++ {
		t := 1.0 - float64(dy)/float64(h)
		col := pal.Color(t)
		for dx := 0; dx < w; dx++ {
			px, py := x+dx, y+dy
			if px >= bounds.Min.X && py >= bounds.Min.Y && px < bounds.Max.X && py < bounds.Max.Y {
				img.SetNRGBA(px, py, col)
			}
		}
	}
	// Border.
	borderCol := color.NRGBA{80, 80, 80, 255}
	for dy := 0; dy <= h; dy++ {
		img.SetNRGBA(clampInt(x, bounds.Max.X-1), clampInt(y+dy, bounds.Max.Y-1), borderCol)
		img.SetNRGBA(clampInt(x+w, bounds.Max.X-1), clampInt(y+dy, bounds.Max.Y-1), borderCol)
	}
	for dx := 0; dx <= w; dx++ {
		img.SetNRGBA(clampInt(x+dx, bounds.Max.X-1), clampInt(y, bounds.Max.Y-1), borderCol)
		img.SetNRGBA(clampInt(x+dx, bounds.Max.X-1), clampInt(y+h, bounds.Max.Y-1), borderCol)
	}
}

func clampInt(v, max int) int {
	if v < 0 {
		return 0
	}
	if v >= max {
		return max - 1
	}
	return v
}

// continents holds simplified continental outline polygons as [lon,lat] pairs.
// These are greatly simplified blobs for visual identification only.
var continents = [][][2]float64{
	// North America (simplified)
	{{-140, 70}, {-60, 70}, {-55, 45}, {-70, 25}, {-90, 15}, {-120, 20}, {-140, 50}, {-140, 70}},
	// South America (simplified)
	{{-80, 10}, {-50, 10}, {-35, -5}, {-40, -55}, {-70, -55}, {-80, -35}, {-80, 10}},
	// Europe (simplified)
	{{-10, 35}, {40, 35}, {40, 70}, {-10, 70}, {-10, 35}},
	// Africa (simplified)
	{{-20, 35}, {55, 35}, {50, -35}, {15, -40}, {-20, -30}, {-20, 35}},
	// Asia (simplified)
	{{25, 70}, {180, 70}, {145, 30}, {100, 5}, {60, 10}, {25, 35}, {25, 70}},
	// Australia (simplified)
	{{114, -20}, {154, -20}, {154, -40}, {130, -40}, {114, -35}, {114, -20}},
}
