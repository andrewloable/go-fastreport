package object

import (
	"image"
	"image/color"
	"math"

	"github.com/andrewloable/go-fastreport/units"
)

// ── digit definitions ────────────────────────────────────────────────────────

// zipDigits holds the polyline paths for digits 0-9. Coordinates are in a
// 5×10 grid. Reference: C# ZipCodeObject.cs static constructor (line 380-393).
var zipDigits = [10][][2]int{
	{{0, 0}, {5, 0}, {5, 10}, {0, 10}, {0, 0}},                // 0
	{{0, 5}, {5, 0}, {5, 10}},                                  // 1
	{{0, 0}, {5, 0}, {5, 5}, {0, 10}, {5, 10}},                 // 2
	{{0, 0}, {5, 0}, {0, 5}, {5, 5}, {0, 10}},                  // 3
	{{0, 0}, {0, 5}, {5, 5}, {5, 0}, {5, 10}},                  // 4
	{{5, 0}, {0, 0}, {0, 5}, {5, 5}, {5, 10}, {0, 10}},         // 5
	{{5, 0}, {0, 5}, {0, 10}, {5, 10}, {5, 5}, {0, 5}},         // 6
	{{0, 0}, {5, 0}, {0, 5}, {0, 10}},                          // 7
	{{0, 5}, {0, 0}, {5, 0}, {5, 10}, {0, 10}, {0, 5}, {5, 5}}, // 8
	{{5, 5}, {0, 5}, {0, 0}, {5, 0}, {5, 5}, {0, 10}},          // 9
}

// zipGridRows holds the dot-grid pattern (11 rows). Each value is a decimal
// number whose digits (read right-to-left by % 10 / 10) encode which of the 6
// column positions should receive a dot. Matches C# DrawSegmentGrid grid array.
var zipGridRows = [11]int{
	111111, 110001, 101001, 100101, 100011,
	111111, 110001, 101001, 100101, 100011,
	111111,
}

// ── public render function ───────────────────────────────────────────────────

// RenderZipCode renders a ZipCodeObject as a PNG-ready image.Image.
// The image includes fill, border, markers, grid, and digit strokes.
// Parameters borderColor/borderWidth come from the object's Border; fillColor
// from the object's Fill.  drawBorder controls whether the rectangle outline is
// drawn (true only when Border.Lines != None).  w×h are the image dimensions.
//
// This mirrors C# ZipCodeObject.Draw() (ZipCodeObject.cs line 266-292).
func RenderZipCode(z *ZipCodeObject, borderColor color.RGBA, borderWidth float32, drawBorder bool, fillColor color.RGBA, w, h int) image.Image {
	if w < 1 || h < 1 {
		return nil
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// 1. Fill background (C# base.Draw → DrawBackground).
	if fillColor.A > 0 {
		zipFillRect(img, 0, 0, w, h, fillColor)
	}

	// 2. Draw simple border only when Border.Lines != None (C# base.Draw handles
	//    the outline; it checks Border.Lines before drawing).
	if drawBorder && borderWidth > 0 && borderColor.A > 0 {
		bw := int(math.Round(float64(borderWidth)))
		bw = max(bw, 1)
		zipDrawRectOutline(img, 0, 0, w, h, bw, borderColor)
	}

	// 3. Draw segments (C# ZipCodeObject.Draw loop, ZipCodeObject.cs line 273-291).
	offsetX := float32(0)
	if z.showMarkers {
		// Draw starting marker.
		zipDrawSegment(img, z, -1, 0, borderColor, borderWidth)
		offsetX += z.spacing
	}

	text := z.text
	for len(text) < z.segmentCount {
		text = "0" + text
	}
	if len(text) > z.segmentCount {
		text = text[:z.segmentCount]
	}

	for _, ch := range text {
		symbol := -1
		if ch >= '0' && ch <= '9' {
			symbol = int(ch - '0')
		}
		zipDrawSegment(img, z, symbol, offsetX, borderColor, borderWidth)
		offsetX += z.spacing
	}

	return img
}

// ZipCodeDimensions computes the Width and Height for a ZipCodeObject,
// matching C# ZipCodeObject.Draw() (ZipCodeObject.cs line 268-269).
func ZipCodeDimensions(z *ZipCodeObject, borderWidth float32) (width, height float32) {
	markerCount := 0
	if z.showMarkers {
		markerCount = 1
	}
	width = float32(markerCount+z.segmentCount) * z.spacing

	height = z.segmentHeight
	if z.showMarkers {
		height += units.Millimeters * 4 // C# Units.Millimeters * 4
	}
	height += borderWidth
	return
}

// ── segment drawing ──────────────────────────────────────────────────────────

// zipDrawSegment draws one segment (marker or digit) at the given X offset.
// Mirrors C# ZipCodeObject.DrawSegment (ZipCodeObject.cs line 195-242).
func zipDrawSegment(img *image.RGBA, z *ZipCodeObject, symbol int, offsetX float32, borderColor color.RGBA, borderWidth float32) {
	offsetY := float32(0)

	if z.showMarkers {
		zipDrawReferenceLine(img, offsetX, borderColor)
		if offsetX == 0 {
			return // first marker segment – only draws the reference line
		}
		offsetX += units.Millimeters * 1
		offsetY = units.Millimeters * 4
	} else {
		offsetX += borderWidth / 2
		offsetY += borderWidth / 2
	}

	if z.showGrid {
		zipDrawSegmentGrid(img, z, offsetX, offsetY, borderColor)
	}

	if symbol >= 0 && symbol <= 9 {
		zipDrawDigit(img, z, symbol, offsetX, offsetY, borderColor, borderWidth)
	}
}

// zipDrawReferenceLine draws the top reference bar (and start line for the
// first segment). Mirrors C# DrawReferenceLine (ZipCodeObject.cs line 177-193).
func zipDrawReferenceLine(img *image.RGBA, offsetX float32, c color.RGBA) {
	mm := units.Millimeters

	// Main bar: 7mm wide × 2mm tall at (offsetX, 0).
	x0 := int(math.Round(float64(offsetX)))
	w := int(math.Round(float64(mm * 7)))
	h := int(math.Round(float64(mm * 2)))
	zipFillRect(img, x0, 0, w, h, c)

	// Start line (only for the first segment, offsetX == 0):
	// 7mm × 1mm at (offsetX, 3mm).
	if offsetX == 0 {
		y := int(math.Round(float64(mm * 3)))
		h2 := int(math.Round(float64(mm * 1)))
		zipFillRect(img, x0, y, w, h2, c)
	}
}

// zipDrawSegmentGrid draws the dot grid for one segment.
// Mirrors C# DrawSegmentGrid (ZipCodeObject.cs line 142-175).
func zipDrawSegmentGrid(img *image.RGBA, z *ZipCodeObject, offsetX, offsetY float32, c color.RGBA) {
	mm := units.Millimeters
	ratioX := z.segmentWidth / (units.Centimeters * 0.5)
	ratioY := z.segmentHeight / (units.Centimeters * 1)
	pointSize := mm * 0.25 // C# Units.Millimeters * 0.25f = 0.945 px

	y := float32(0) // AbsTop = 0
	for _, gridRow := range zipGridRows {
		row := gridRow
		x := float32(0) // AbsLeft = 0
		for row > 0 {
			if row%10 == 1 {
				// Fill a small circle at (x + offsetX, y + offsetY).
				cx := x + offsetX
				cy := y + offsetY
				zipFillCircle(img, cx, cy, pointSize/2, c)
			}
			row /= 10
			x += mm * 1 * ratioX
		}
		y += mm * 1 * ratioY
	}
}

// zipDrawDigit draws one digit (0-9) as a polyline.
// Mirrors C# DrawSegment digit drawing (ZipCodeObject.cs line 221-241).
func zipDrawDigit(img *image.RGBA, z *ZipCodeObject, symbol int, offsetX, offsetY float32, c color.RGBA, penWidth float32) {
	mm := units.Millimeters
	ratioX := z.segmentWidth / (units.Centimeters * 0.5)
	ratioY := z.segmentHeight / (units.Centimeters * 1)

	digit := zipDigits[symbol]
	if len(digit) < 2 {
		return
	}

	// Build path in pixel coordinates.
	// C# path[i] = (AbsLeft + digit[i].X * mm * ratioX + offsetX,
	//               AbsTop  + digit[i].Y * mm * ratioY + offsetY)
	// AbsLeft = 0, AbsTop = 0 for image rendering.
	path := make([][2]float32, len(digit))
	for i, pt := range digit {
		path[i][0] = float32(pt[0])*mm*ratioX + offsetX
		path[i][1] = float32(pt[1])*mm*ratioY + offsetY
	}

	// Draw polyline with rounded caps and joins (C# LineCap.Round, LineJoin.Round).
	radius := penWidth / 2
	for i := 0; i < len(path)-1; i++ {
		zipDrawThickLine(img, path[i][0], path[i][1], path[i+1][0], path[i+1][1], radius, c)
	}
}

// ── drawing primitives ───────────────────────────────────────────────────────

// zipFillRect fills a rectangle on img. Coordinates are clipped to bounds.
func zipFillRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	b := img.Bounds()
	x0 := max(x, b.Min.X)
	y0 := max(y, b.Min.Y)
	x1 := min(x+w, b.Max.X)
	y1 := min(y+h, b.Max.Y)
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			img.SetRGBA(px, py, c)
		}
	}
}

// zipDrawRectOutline draws a rectangle outline with the given line width (inset).
func zipDrawRectOutline(img *image.RGBA, x, y, w, h, lineWidth int, c color.RGBA) {
	// Top edge.
	zipFillRect(img, x, y, w, lineWidth, c)
	// Bottom edge.
	zipFillRect(img, x, y+h-lineWidth, w, lineWidth, c)
	// Left edge.
	zipFillRect(img, x, y+lineWidth, lineWidth, h-2*lineWidth, c)
	// Right edge.
	zipFillRect(img, x+w-lineWidth, y+lineWidth, lineWidth, h-2*lineWidth, c)
}

// zipFillCircle fills a circle at center (cx, cy) with the given radius.
func zipFillCircle(img *image.RGBA, cx, cy, radius float32, c color.RGBA) {
	b := img.Bounds()
	r := radius
	if r < 0.5 {
		r = 0.5 // minimum: fill at least the center pixel
	}
	ix := int(math.Round(float64(cx)))
	iy := int(math.Round(float64(cy)))
	ir := int(math.Ceil(float64(r)))

	rSq := float64(r * r)
	for dy := -ir; dy <= ir; dy++ {
		for dx := -ir; dx <= ir; dx++ {
			if float64(dx*dx+dy*dy) <= rSq {
				px, py := ix+dx, iy+dy
				if px >= b.Min.X && px < b.Max.X && py >= b.Min.Y && py < b.Max.Y {
					img.SetRGBA(px, py, c)
				}
			}
		}
	}
}

// zipDrawThickLine draws a line segment from (x0, y0) to (x1, y1) with the
// given radius (half-width), using filled circles at each step for rounded
// caps (matching C# LineCap.Round, LineJoin.Round).
func zipDrawThickLine(img *image.RGBA, x0, y0, x1, y1, radius float32, c color.RGBA) {
	// Bresenham-style walk along the line, stamping circles.
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 0.5 {
		zipFillCircle(img, x0, y0, radius, c)
		return
	}

	steps := int(math.Ceil(dist))
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		cx := float32(float64(x0) + dx*t)
		cy := float32(float64(y0) + dy*t)
		zipFillCircle(img, cx, cy, radius, c)
	}
}
