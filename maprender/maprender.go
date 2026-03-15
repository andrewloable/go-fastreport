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

	// Draw simplified continental outlines (major blobs only).
	for _, poly := range continents {
		drawPoly(poly, land)
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
