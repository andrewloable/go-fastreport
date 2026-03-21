// render2d.go implements 2D barcode rendering.
//
// Ported from C# Barcode2DBase.DrawBarcode and BarcodeQR.Draw2DBarcode.
// A 2D barcode is represented as a boolean module grid; each true cell is a
// dark module rendered in the chosen shape.
package barcode

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Matrix2DProvider is implemented by 2D barcode types that can supply a
// boolean module grid for rendering by DrawBarcode2D.
type Matrix2DProvider interface {
	// GetMatrix returns (matrix[row][col], rows, cols).
	// matrix[r][c] == true means dark module at row r, column c.
	GetMatrix() (matrix [][]bool, rows, cols int)
}

// DrawBarcode2D renders a boolean module matrix to an image.
// Ported from C# Barcode2DBase.DrawBarcode which iterates over the module
// grid and fills each dark cell with a filled rectangle.
func DrawBarcode2D(matrix [][]bool, rows, cols, width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	if rows <= 0 || cols <= 0 || width <= 0 || height <= 0 {
		return img
	}

	black := image.NewUniform(color.Black)
	for r := range rows {
		if r >= len(matrix) {
			break
		}
		for c := range cols {
			if c >= len(matrix[r]) || !matrix[r][c] {
				continue
			}
			x0 := c * width / cols
			x1 := (c + 1) * width / cols
			y0 := r * height / rows
			y1 := (r + 1) * height / rows
			if x1 <= x0 {
				x1 = x0 + 1
			}
			if y1 <= y0 {
				y1 = y0 + 1
			}
			draw.Draw(img, image.Rect(x0, y0, x1, y1), black, image.Point{}, draw.Src)
		}
	}
	return img
}

// DrawQRCode2D renders QR-specific module styles (Rectangle, Circle, Diamond,
// RoundedSquare, PillHorizontal, PillVertical, Plus, Hexagon, Star, Snowflake).
//
// Ported from C# BarcodeQR.Draw2DBarcode (BarcodeQR.cs:302–508).
// The shape parameter matches the C# QrModuleShape enum names.
// angle is used only by rotational shapes (Hexagon, Star, Snowflake).
func DrawQRCode2D(matrix [][]bool, rows, cols, width, height int, shape string, useThinModules, quietZone bool, angle int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	if rows <= 0 || cols <= 0 || width <= 0 || height <= 0 {
		return img
	}

	quiet := 0
	if quietZone {
		quiet = 4
	}
	black := color.RGBA{A: 255}
	// C# BarcodeQR.Draw2DBarcode: scale = 1.25 for Diamond/Hexagon/Star/Plus,
	// scale = 1.05 for Circle.  See individual shape cases below.
	const defaultScale = 1.25
	const circleScale = 1.05
	const paddingRatio = 0.1 // 10% inset for UseThinModules

	for r := range rows {
		if r >= len(matrix) {
			break
		}
		for c := range cols {
			if c >= len(matrix[r]) || !matrix[r][c] {
				continue
			}

			// Cell bounds in pixels.
			x0 := float64(c) * float64(width) / float64(cols)
			x1 := float64(c+1) * float64(width) / float64(cols)
			y0 := float64(r) * float64(height) / float64(rows)
			y1 := float64(r+1) * float64(height) / float64(rows)

			isFinder := isQRFinderCell(r, c, rows, cols, quiet)

			// Render coordinates: finder pattern always uses full cell;
			// other modules may be inset when UseThinModules is set.
			// C# BarcodeQR.Draw2DBarcode:330–357.
			renderX, renderY, renderW, renderH := x0, y0, x1-x0, y1-y0
			if useThinModules && !isFinder {
				insetX := renderW * paddingRatio
				insetY := renderH * paddingRatio
				renderX += insetX
				renderY += insetY
				renderW -= 2 * insetX
				renderH -= 2 * insetY
			}

			switch {
			case stringsEqualFold(shape, "Circle"):
				// C# BarcodeQR.Draw2DBarcode:365–379 — scale 1.05, centered ellipse.
				scale := circleScale
				scaledW := renderW * scale
				scaledH := renderH * scale
				offsetX := (scaledW - renderW) / 2
				offsetY := (scaledH - renderH) / 2
				drawFilledEllipse(img, renderX-offsetX, renderY-offsetY, renderX-offsetX+scaledW, renderY-offsetY+scaledH, black)

			case stringsEqualFold(shape, "Diamond"):
				// C# BarcodeQR.Draw2DBarcode:381–383 — diamond polygon scaled 1.25.
				pts := qrCreateDiamondPoints(renderX, renderY, renderW, renderH, defaultScale)
				drawFilledPolygon(img, pts, black)

			case stringsEqualFold(shape, "RoundedSquare"):
				// C# BarcodeQR.Draw2DBarcode:385–419 — adaptive corner rounding.
				var tl, tr, br, bl bool
				if useThinModules && !isFinder {
					// All corners rounded when thin modules are active.
					tl, tr, br, bl = true, true, true, true
				} else {
					// Adaptive: round a corner when the corner neighbour is absent or
					// either of its two orthogonal neighbours is absent.
					// C# BarcodeQR.Draw2DBarcode:408–411.
					tl = (!qrNeighborDark(matrix, rows, cols, r, c, 0, -1) || !qrNeighborDark(matrix, rows, cols, r, c, -1, 0)) || !qrNeighborDark(matrix, rows, cols, r, c, -1, -1)
					tr = (!qrNeighborDark(matrix, rows, cols, r, c, 0, -1) || !qrNeighborDark(matrix, rows, cols, r, c, 1, 0)) || !qrNeighborDark(matrix, rows, cols, r, c, 1, -1)
					br = (!qrNeighborDark(matrix, rows, cols, r, c, 0, 1) || !qrNeighborDark(matrix, rows, cols, r, c, 1, 0)) || !qrNeighborDark(matrix, rows, cols, r, c, 1, 1)
					bl = (!qrNeighborDark(matrix, rows, cols, r, c, 0, 1) || !qrNeighborDark(matrix, rows, cols, r, c, -1, 0)) || !qrNeighborDark(matrix, rows, cols, r, c, -1, 1)
				}
				pts := qrCreateRoundedRectPoints(renderX, renderY, renderW, renderH, tl, tr, br, bl)
				drawFilledPolygon(img, pts, black)

			case stringsEqualFold(shape, "PillHorizontal"):
				// C# BarcodeQR.Draw2DBarcode:422–432.
				hasLeft := qrNeighborDark(matrix, rows, cols, r, c, -1, 0)
				hasRight := qrNeighborDark(matrix, rows, cols, r, c, 1, 0)
				pts := qrCreateHorizontalPillPoints(renderX, renderY, renderW, renderH, hasLeft, hasRight)
				drawFilledPolygon(img, pts, black)

			case stringsEqualFold(shape, "PillVertical"):
				// C# BarcodeQR.Draw2DBarcode:434–443.
				hasTop := qrNeighborDark(matrix, rows, cols, r, c, 0, -1)
				hasBottom := qrNeighborDark(matrix, rows, cols, r, c, 0, 1)
				pts := qrCreateVerticalPillPoints(renderX, renderY, renderW, renderH, hasTop, hasBottom)
				drawFilledPolygon(img, pts, black)

			case stringsEqualFold(shape, "Plus"):
				// C# BarcodeQR.Draw2DBarcode:445–466 — two overlapping rectangles.
				cx := renderX + renderW/2
				cy := renderY + renderH/2
				horizW := renderW / 2 * defaultScale
				horizH := renderH * 0.2
				vertW := renderW * 0.2
				vertH := renderH / 2 * defaultScale
				// Horizontal bar
				drawFilledRect(img, cx-horizW, cy-horizH, cx+horizW, cy+horizH, black)
				// Vertical bar
				drawFilledRect(img, cx-vertW, cy-vertH, cx+vertW, cy+vertH, black)

			case stringsEqualFold(shape, "Hexagon"):
				// C# BarcodeQR.Draw2DBarcode:469–478.
				rx := renderW / 2 * defaultScale
				ry := renderH / 2 * defaultScale
				pts := qrCreateHexagonPoints(renderX+renderW/2, renderY+renderH/2, rx, ry, angle)
				drawFilledPolygon(img, pts, black)

			case stringsEqualFold(shape, "Star"):
				// C# BarcodeQR.Draw2DBarcode:481–492.
				outerRx := renderW / 2 * defaultScale
				outerRy := renderH / 2 * defaultScale
				pts := qrCreateStarPoints(renderX+renderW/2, renderY+renderH/2, outerRx, outerRy, angle)
				drawFilledPolygon(img, pts, black)

			case stringsEqualFold(shape, "Snowflake"):
				// C# BarcodeQR.Draw2DBarcode:494–503.
				pts := qrCreateSnowflakePoints(renderX+renderW/2, renderY+renderH/2, renderW/2, renderH/2, angle)
				drawFilledPolygon(img, pts, black)

			default:
				// Rectangle (default) and any unrecognised shape.
				drawFilledRect(img, renderX, renderY, renderX+renderW, renderY+renderH, black)
			}
		}
	}
	return img
}

// ── Shape helper functions ────────────────────────────────────────────────────

// qrNeighborDark reports whether the neighbour at (col+dc, row+dr) is a dark module.
// Ported from C# BarcodeQR.IsNeighborDark (BarcodeQR.cs:886).
// Note: matrix is indexed as matrix[row][col]; dx maps to column delta, dy to row delta.
func qrNeighborDark(matrix [][]bool, rows, cols, row, col, dc, dr int) bool {
	nc := col + dc
	nr := row + dr
	if nr < 0 || nr >= rows || nc < 0 || nc >= cols {
		return false
	}
	if nr >= len(matrix) || nc >= len(matrix[nr]) {
		return false
	}
	return matrix[nr][nc]
}

// qrCreateDiamondPoints returns the 5 vertices of a diamond (rotated square)
// scaled uniformly from its center.
// Ported from C# BarcodeQR.CreateDiamondPoints (BarcodeQR.cs:520–542).
func qrCreateDiamondPoints(x, y, w, h, scale float64) [][2]float64 {
	cx := x + w/2
	cy := y + h/2
	pts := [][2]float64{
		{cx, y},     // top
		{x + w, cy}, // right
		{cx, y + h}, // bottom
		{x, cy},     // left
		{cx, y},     // close
	}
	for i := range pts {
		dx := pts[i][0] - cx
		dy := pts[i][1] - cy
		pts[i][0] = cx + dx*scale
		pts[i][1] = cy + dy*scale
	}
	return pts
}

// qrCreateRoundedRectPoints approximates a rounded rectangle as a polygon
// by sampling arc segments at each corner.
// Ported from C# BarcodeQR.CreateRoundedRectanglePath (BarcodeQR.cs:556–622).
func qrCreateRoundedRectPoints(x, y, w, h float64, topLeft, topRight, bottomRight, bottomLeft bool) [][2]float64 {
	radius := math.Min(w, h) / 4
	if radius < 1 {
		radius = 1
	}
	const arcSteps = 8 // number of line segments per 90° arc
	var pts [][2]float64
	addArc := func(cx, cy, rx, ry, startDeg, endDeg float64) {
		steps := arcSteps
		for i := 0; i <= steps; i++ {
			t := startDeg + (endDeg-startDeg)*float64(i)/float64(steps)
			rad := t * math.Pi / 180
			pts = append(pts, [2]float64{cx + rx*math.Cos(rad), cy + ry*math.Sin(rad)})
		}
	}
	// Top-left corner
	if topLeft {
		addArc(x+radius, y+radius, radius, radius, 180, 270)
	} else {
		pts = append(pts, [2]float64{x, y})
	}
	// Top edge → top-right corner
	if topRight {
		addArc(x+w-radius, y+radius, radius, radius, 270, 360)
	} else {
		pts = append(pts, [2]float64{x + w, y})
	}
	// Right edge → bottom-right corner
	if bottomRight {
		addArc(x+w-radius, y+h-radius, radius, radius, 0, 90)
	} else {
		pts = append(pts, [2]float64{x + w, y + h})
	}
	// Bottom edge → bottom-left corner
	if bottomLeft {
		addArc(x+radius, y+h-radius, radius, radius, 90, 180)
	} else {
		pts = append(pts, [2]float64{x, y + h})
	}
	return pts
}

// qrCreateHorizontalPillPoints returns a polygon approximating a horizontal pill
// (capsule) shape. Ends are rounded unless a dark neighbor is present on that side.
// Ported from C# BarcodeQR.CreateHorizontalPillPath (BarcodeQR.cs:636–669).
func qrCreateHorizontalPillPoints(x, y, w, h float64, hasLeft, hasRight bool) [][2]float64 {
	const arcSteps = 12
	var pts [][2]float64
	addSemicircle := func(cx, cy, rx, ry, startDeg float64) {
		for i := 0; i <= arcSteps; i++ {
			deg := startDeg + 180*float64(i)/float64(arcSteps)
			rad := deg * math.Pi / 180
			pts = append(pts, [2]float64{cx + rx*math.Cos(rad), cy + ry*math.Sin(rad)})
		}
	}
	rx := w / 2
	ry := h / 2
	cy := y + ry
	if hasLeft && hasRight {
		// Full rectangle.
		pts = append(pts,
			[2]float64{x, y}, [2]float64{x + w, y},
			[2]float64{x + w, y + h}, [2]float64{x, y + h},
		)
	} else if hasLeft && !hasRight {
		// Left flat, right rounded.
		pts = append(pts, [2]float64{x, y}, [2]float64{x + w - rx, y})
		addSemicircle(x+w-rx, cy, rx, ry, -90)
		pts = append(pts, [2]float64{x, y + h})
	} else if !hasLeft && hasRight {
		// Left rounded, right flat.
		addSemicircle(x+rx, cy, rx, ry, 90)
		pts = append(pts, [2]float64{x + w, y + h}, [2]float64{x + w, y})
	} else {
		// Both ends rounded.
		addSemicircle(x+rx, cy, rx, ry, 90)    // left semicircle
		addSemicircle(x+w-rx, cy, rx, ry, -90) // right semicircle
	}
	return pts
}

// qrCreateVerticalPillPoints returns a polygon approximating a vertical pill
// (capsule) shape. Ends are rounded unless a dark neighbor is present on that side.
// Ported from C# BarcodeQR.CreateVerticalPillPath (BarcodeQR.cs:685–719).
func qrCreateVerticalPillPoints(x, y, w, h float64, hasTop, hasBottom bool) [][2]float64 {
	const arcSteps = 12
	var pts [][2]float64
	addSemicircle := func(cx, cy, rx, ry, startDeg float64) {
		for i := 0; i <= arcSteps; i++ {
			deg := startDeg + 180*float64(i)/float64(arcSteps)
			rad := deg * math.Pi / 180
			pts = append(pts, [2]float64{cx + rx*math.Cos(rad), cy + ry*math.Sin(rad)})
		}
	}
	rx := w / 2
	ry := h / 2
	cx := x + rx
	if hasTop && hasBottom {
		// Full rectangle.
		pts = append(pts,
			[2]float64{x, y}, [2]float64{x + w, y},
			[2]float64{x + w, y + h}, [2]float64{x, y + h},
		)
	} else if hasTop && !hasBottom {
		// Top flat, bottom rounded.
		pts = append(pts, [2]float64{x, y}, [2]float64{x + w, y})
		addSemicircle(cx, y+h-ry, rx, ry, 0)
		pts = append(pts, [2]float64{x, y + h})
	} else if !hasTop && hasBottom {
		// Top rounded, bottom flat.
		addSemicircle(cx, y+ry, rx, ry, 180)
		pts = append(pts, [2]float64{x + w, y + h}, [2]float64{x, y + h})
	} else {
		// Both ends rounded.
		addSemicircle(cx, y+ry, rx, ry, 180)        // top semicircle
		addSemicircle(cx, y+h-ry, rx, ry, 0)        // bottom semicircle
	}
	return pts
}

// qrCreateHexagonPoints returns the 6 vertices of a regular hexagon.
// Ported from C# BarcodeQR.CreateHexagonPoints (BarcodeQR.cs:730–743).
func qrCreateHexagonPoints(cx, cy, rx, ry float64, angleDeg int) [][2]float64 {
	rotRad := float64(angleDeg) * math.Pi / 180
	pts := make([][2]float64, 6)
	for i := range 6 {
		a := math.Pi/3*float64(i) - math.Pi/6 + rotRad // flat side on top
		pts[i] = [2]float64{cx + rx*math.Cos(a), cy + ry*math.Sin(a)}
	}
	return pts
}

// qrCreateStarPoints returns the 10 vertices of a 5-pointed star.
// Ported from C# BarcodeQR.CreateStarPoints (BarcodeQR.cs:758–778).
func qrCreateStarPoints(cx, cy, outerRx, outerRy float64, angleDeg int) [][2]float64 {
	rotRad := float64(angleDeg) * math.Pi / 180
	pts := make([][2]float64, 10)
	for i := range 10 {
		a := math.Pi/5*float64(i) - math.Pi/2 + rotRad
		ratio := 1.0
		if i%2 != 0 {
			ratio = 0.6
		}
		rx := outerRx * ratio
		ry := outerRy * ratio
		pts[i] = [2]float64{cx + rx*math.Cos(a), cy + ry*math.Sin(a)}
	}
	return pts
}

// qrCreateSnowflakePoints returns a 120-point polygon approximating a snowflake.
// Ported from C# BarcodeQR.CreateSnowflakePoints (BarcodeQR.cs:790–831).
func qrCreateSnowflakePoints(cx, cy, sx, sy float64, angleDeg int) [][2]float64 {
	const pointCount = 120
	pts := make([][2]float64, pointCount)
	useRotation := angleDeg != 0
	var cosA, sinA float64 = 1, 0
	if useRotation {
		rad := float64(angleDeg) * math.Pi / 180
		cosA = math.Cos(rad)
		sinA = math.Sin(rad)
	}
	step := math.Pi * 2 / pointCount
	for i := range pointCount {
		a := float64(i) * step
		dist := 1.0 + 0.4*math.Cos(6*a+math.Pi) // 6 lobes
		dx := dist * math.Cos(a) * sx
		dy := dist * math.Sin(a) * sy
		if useRotation {
			dx, dy = dx*cosA-dy*sinA, dx*sinA+dy*cosA
		}
		pts[i] = [2]float64{cx + dx, cy + dy}
	}
	return pts
}

// DrawSwissCross overlays a Swiss cross in the centre of img.
// Ported from C# Barcode2DBase.DrawBarcode (Barcode2DBase.cs:22–29).
//
// The overlay consists of four rectangles (percentages of image dimensions):
//  1. White outer square  (±7% from centre)
//  2. Black filled square (±6%)
//  3. White horizontal bar (±4% wide, ±1.5% tall)
//  4. White vertical bar  (±1.5% wide, ±4% tall)
//
// width and height are the image pixel dimensions. The C# showText adjustment
// (subtracting 21px when showText is true) is not applied here; callers should
// pass the effective drawing height when showText is active.
func DrawSwissCross(img *image.RGBA, width, height int) {
	w := float64(width)
	h := float64(height)
	cx := w / 2
	cy := h / 2

	// 1. White outer square.
	drawFilledRect(img, cx-w*0.07, cy-h*0.07, cx+w*0.07, cy+h*0.07, color.White)
	// 2. Black filled square.
	drawFilledRect(img, cx-w*0.06, cy-h*0.06, cx+w*0.06, cy+h*0.06, color.Black)
	// 3. White horizontal bar.
	drawFilledRect(img, cx-w*0.04, cy-h*0.015, cx+w*0.04, cy+h*0.015, color.White)
	// 4. White vertical bar.
	drawFilledRect(img, cx-w*0.015, cy-h*0.04, cx+w*0.015, cy+h*0.04, color.White)
}

// ── Pixel-level drawing primitives ───────────────────────────────────────────

func drawFilledRect(img *image.RGBA, x0, y0, x1, y1 float64, fill color.Color) {
	left := int(math.Floor(x0))
	right := int(math.Ceil(x1))
	top := int(math.Floor(y0))
	bottom := int(math.Ceil(y1))
	if right <= left {
		right = left + 1
	}
	if bottom <= top {
		bottom = top + 1
	}
	draw.Draw(img, image.Rect(left, top, right, bottom), image.NewUniform(fill), image.Point{}, draw.Src)
}

func drawFilledEllipse(img *image.RGBA, x0, y0, x1, y1 float64, fill color.Color) {
	left := int(math.Floor(x0))
	right := int(math.Ceil(x1))
	top := int(math.Floor(y0))
	bottom := int(math.Ceil(y1))
	if right <= left {
		right = left + 1
	}
	if bottom <= top {
		bottom = top + 1
	}

	cx := (x0 + x1) / 2
	cy := (y0 + y1) / 2
	rx := math.Max((x1-x0)/2, 0.5)
	ry := math.Max((y1-y0)/2, 0.5)

	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			dx := (float64(x) + 0.5 - cx) / rx
			dy := (float64(y) + 0.5 - cy) / ry
			if dx*dx+dy*dy <= 1 {
				img.Set(x, y, fill)
			}
		}
	}
}

// drawFilledPolygon rasterises a polygon given as a slice of (x,y) pairs using
// a scanline fill algorithm (even-odd rule).
func drawFilledPolygon(img *image.RGBA, pts [][2]float64, fill color.Color) {
	if len(pts) < 3 {
		return
	}
	bounds := img.Bounds()

	// Find bounding box.
	minY, maxY := pts[0][1], pts[0][1]
	for _, p := range pts {
		if p[1] < minY {
			minY = p[1]
		}
		if p[1] > maxY {
			maxY = p[1]
		}
	}
	top := int(math.Floor(minY))
	bottom := int(math.Ceil(maxY))
	if top < bounds.Min.Y {
		top = bounds.Min.Y
	}
	if bottom > bounds.Max.Y {
		bottom = bounds.Max.Y
	}

	n := len(pts)
	for y := top; y < bottom; y++ {
		scanY := float64(y) + 0.5
		var xs []float64
		for i := range n {
			j := (i + 1) % n
			yi, yj := pts[i][1], pts[j][1]
			if (yi <= scanY && yj > scanY) || (yj <= scanY && yi > scanY) {
				t := (scanY - yi) / (yj - yi)
				xs = append(xs, pts[i][0]+t*(pts[j][0]-pts[i][0]))
			}
		}
		// Sort intersection X values (insertion sort — usually 2–4 elements).
		for a := 1; a < len(xs); a++ {
			for b := a; b > 0 && xs[b] < xs[b-1]; b-- {
				xs[b], xs[b-1] = xs[b-1], xs[b]
			}
		}
		for i := 0; i+1 < len(xs); i += 2 {
			x0 := int(math.Floor(xs[i]))
			x1 := int(math.Ceil(xs[i+1]))
			if x0 < bounds.Min.X {
				x0 = bounds.Min.X
			}
			if x1 > bounds.Max.X {
				x1 = bounds.Max.X
			}
			for x := x0; x < x1; x++ {
				img.Set(x, y, fill)
			}
		}
	}
}

// ── QR utility helpers ────────────────────────────────────────────────────────

func isQRFinderCell(row, col, rows, cols, quiet int) bool {
	if rows < 7 || cols < 7 {
		return false
	}
	top := quiet
	left := quiet
	right := cols - quiet - 7
	bottom := rows - quiet - 7

	inRect := func(r, c, top, left int) bool {
		return r >= top && r < top+7 && c >= left && c < left+7
	}
	return inRect(row, col, top, left) ||
		inRect(row, col, top, right) ||
		inRect(row, col, bottom, left)
}

func stringsEqualFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca := a[i]
		cb := b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
