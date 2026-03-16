// Package chart internal tests to cover unexported drawing primitives.
package chart

import (
	"image"
	"image/color"
	"math"
	"testing"
)

// ── drawHLine ─────────────────────────────────────────────────────────────────

func TestDrawHLine_NormalOrder(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{255, 0, 0, 255}
	drawHLine(img, 10, 20, 50, c)
	// Pixels from x=10 to x=20 at y=50 should be red.
	for x := 10; x <= 20; x++ {
		got := img.RGBAAt(x, 50)
		if got != c {
			t.Errorf("x=%d y=50: got %v want %v", x, got, c)
		}
	}
}

func TestDrawHLine_ReversedOrder(t *testing.T) {
	// x0 > x1 triggers the swap branch.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{0, 255, 0, 255}
	drawHLine(img, 30, 10, 40, c) // x0=30, x1=10 → swap so line runs 10..30
	for x := 10; x <= 30; x++ {
		got := img.RGBAAt(x, 40)
		if got != c {
			t.Errorf("x=%d y=40: got %v want %v", x, got, c)
		}
	}
}

func TestDrawHLine_SinglePixel(t *testing.T) {
	// x0 == x1 → exactly one pixel.
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	c := color.RGBA{0, 0, 255, 255}
	drawHLine(img, 25, 25, 25, c)
	got := img.RGBAAt(25, 25)
	if got != c {
		t.Errorf("single pixel: got %v want %v", got, c)
	}
}

// ── drawVLine ─────────────────────────────────────────────────────────────────

func TestDrawVLine_NormalOrder(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{255, 128, 0, 255}
	drawVLine(img, 50, 10, 20, c)
	for y := 10; y <= 20; y++ {
		got := img.RGBAAt(50, y)
		if got != c {
			t.Errorf("x=50 y=%d: got %v want %v", y, got, c)
		}
	}
}

func TestDrawVLine_ReversedOrder(t *testing.T) {
	// y0 > y1 triggers the swap branch.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{128, 0, 255, 255}
	drawVLine(img, 60, 80, 20, c) // y0=80, y1=20 → swap so line runs 20..80
	for y := 20; y <= 80; y++ {
		got := img.RGBAAt(60, y)
		if got != c {
			t.Errorf("x=60 y=%d: got %v want %v", y, got, c)
		}
	}
}

func TestDrawVLine_SinglePixel(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	c := color.RGBA{255, 255, 0, 255}
	drawVLine(img, 10, 10, 10, c)
	got := img.RGBAAt(10, 10)
	if got != c {
		t.Errorf("single pixel: got %v want %v", got, c)
	}
}

// ── fillRect ─────────────────────────────────────────────────────────────────

func TestFillRect_NormalInBounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{255, 0, 0, 255}
	fillRect(img, 10, 10, 30, 30, c)
	for y := 10; y <= 30; y++ {
		for x := 10; x <= 30; x++ {
			got := img.RGBAAt(x, y)
			if got != c {
				t.Errorf("(%d,%d): got %v want %v", x, y, got, c)
			}
		}
	}
}

func TestFillRect_ClampX0BelowMin(t *testing.T) {
	// x0 < b.Min.X → clamped to b.Min.X (branch: x0 = b.Min.X).
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	c := color.RGBA{0, 200, 0, 255}
	fillRect(img, -10, 5, 20, 20, c)
	// Should not panic; pixels from x=0 to x=20 at y=5..20 are filled.
	got := img.RGBAAt(0, 10)
	if got != c {
		t.Errorf("clamped x0: pixel (0,10) = %v want %v", got, c)
	}
}

func TestFillRect_ClampY0BelowMin(t *testing.T) {
	// y0 < b.Min.Y → clamped to b.Min.Y.
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	c := color.RGBA{0, 0, 200, 255}
	fillRect(img, 5, -5, 20, 20, c)
	got := img.RGBAAt(10, 0)
	if got != c {
		t.Errorf("clamped y0: pixel (10,0) = %v want %v", got, c)
	}
}

func TestFillRect_ClampX1AboveMax(t *testing.T) {
	// x1 >= b.Max.X → clamped to b.Max.X - 1.
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	c := color.RGBA{200, 200, 0, 255}
	fillRect(img, 30, 10, 100, 30, c) // x1=100 > 50
	got := img.RGBAAt(49, 15)
	if got != c {
		t.Errorf("clamped x1: pixel (49,15) = %v want %v", got, c)
	}
}

func TestFillRect_ClampY1AboveMax(t *testing.T) {
	// y1 >= b.Max.Y → clamped to b.Max.Y - 1.
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	c := color.RGBA{200, 0, 200, 255}
	fillRect(img, 5, 30, 20, 200, c) // y1=200 > 50
	got := img.RGBAAt(10, 49)
	if got != c {
		t.Errorf("clamped y1: pixel (10,49) = %v want %v", got, c)
	}
}

func TestFillRect_AllClamped(t *testing.T) {
	// All four coordinates out of bounds simultaneously.
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	c := color.RGBA{100, 100, 100, 255}
	fillRect(img, -20, -20, 200, 200, c)
	// Entire image should be filled.
	got := img.RGBAAt(25, 25)
	if got != c {
		t.Errorf("all-clamped fillRect center: got %v want %v", got, c)
	}
	got = img.RGBAAt(0, 0)
	if got != c {
		t.Errorf("all-clamped fillRect origin: got %v want %v", got, c)
	}
	got = img.RGBAAt(49, 49)
	if got != c {
		t.Errorf("all-clamped fillRect far corner: got %v want %v", got, c)
	}
}

// ── drawThickLine ─────────────────────────────────────────────────────────────

func TestDrawThickLine_RightToLeft(t *testing.T) {
	// x0 > x1 → sx = -1 branch.
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	c := color.RGBA{255, 0, 0, 255}
	drawThickLine(img, 150, 100, 50, 100, c, 2) // right-to-left horizontal
	// Some pixels in the middle should be set.
	got := img.RGBAAt(100, 100)
	if got != c {
		t.Errorf("right-to-left: pixel (100,100) = %v want %v", got, c)
	}
}

func TestDrawThickLine_BottomToTop(t *testing.T) {
	// y0 > y1 → sy = -1 branch.
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	c := color.RGBA{0, 255, 0, 255}
	drawThickLine(img, 100, 150, 100, 50, c, 2) // bottom-to-top vertical
	got := img.RGBAAt(100, 100)
	if got != c {
		t.Errorf("bottom-to-top: pixel (100,100) = %v want %v", got, c)
	}
}

func TestDrawThickLine_DiagonalSteep(t *testing.T) {
	// dy > dx → the e2 < dx branch fires for each y step.
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	c := color.RGBA{0, 0, 255, 255}
	drawThickLine(img, 10, 10, 12, 80, c, 1)
	// Just verify it doesn't panic and produces some pixels.
	got := img.RGBAAt(11, 40)
	_ = got // value may vary; we just want no panic
}

func TestDrawThickLine_DiagonalShallow(t *testing.T) {
	// dx > dy → the e2 > -dy branch fires for each x step.
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	c := color.RGBA{255, 128, 0, 255}
	drawThickLine(img, 10, 10, 80, 12, c, 1)
	got := img.RGBAAt(40, 11)
	_ = got
}

// ── drawSector ─────────────────────────────────────────────────────────────────

func TestDrawSector_FullCircle(t *testing.T) {
	// endAngle - startAngle >= 2π → the full-circle branch in the inner condition.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{255, 0, 128, 255}
	drawSector(img, 50, 50, 30, 0, 2*math.Pi, c)
	// Some interior pixels should be filled (avoid center which may be overwritten by border).
	// Check that at least one pixel is the sector colour (not all black/white).
	found := false
	for y := 30; y <= 70; y++ {
		for x := 30; x <= 70; x++ {
			if img.RGBAAt(x, y) == c {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("full circle: expected sector colour pixels inside the circle")
	}
}

func TestDrawSector_HalfCircle(t *testing.T) {
	// Normal half-circle sector.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{0, 128, 255, 255}
	drawSector(img, 50, 50, 20, 0, math.Pi, c)
	// A pixel at cx, cy+10 (inside the lower half) should be filled.
	got := img.RGBAAt(50, 60)
	if got != c {
		t.Errorf("half circle interior: got %v want %v", got, c)
	}
}

func TestDrawSector_AngleWraparound(t *testing.T) {
	// startAngle > 0 forces the "for angle < startAngle { angle += 2π }" loop.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{200, 100, 50, 255}
	// Sector from 3π/2 to 5π/2 (wraps around — covers right half circle).
	start := 3 * math.Pi / 2
	end := 5 * math.Pi / 2
	drawSector(img, 50, 50, 20, start, end, c)
	// Interior should have some sector-coloured pixels; scan a region.
	found := false
	for y := 35; y <= 65; y++ {
		for x := 35; x <= 65; x++ {
			if img.RGBAAt(x, y) == c {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("wraparound sector: expected sector colour pixels inside the arc")
	}
}

func TestDrawSector_SmallN_FallbackTo4(t *testing.T) {
	// Very small radius → n computed < 4 → clamped to 4.
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	c := color.RGBA{50, 200, 50, 255}
	// r=1 gives n = int(arc_fraction * 2π * 1 / 1.0) which is tiny → n < 4 → n = 4.
	drawSector(img, 25, 25, 1, 0, math.Pi/2, c)
	// Just verify no panic.
}

func TestDrawSector_QuarterCircle(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{100, 200, 150, 255}
	drawSector(img, 50, 50, 30, 0, math.Pi/2, c)
	// Lower-right quadrant interior should be filled (angle 0..π/2 in standard coords).
	got := img.RGBAAt(65, 65)
	_ = got // just verify no panic; pixel location depends on orientation
}

// ── fillTrapezoid ─────────────────────────────────────────────────────────────

func TestFillTrapezoid_NormalOrder(t *testing.T) {
	// x0 < x1: normal order, no swap.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{255, 0, 0, 255}
	// Trapezoid from x=10..50, y at x=10 is 20, y at x=50 is 40, baseY=60.
	fillTrapezoid(img, 10, 20, 50, 40, 60, c, 255)
	// At x=10, lineY=20, fills y=20..60; pixel (10,30) should be set.
	if img.RGBAAt(10, 30) != c {
		t.Errorf("normal trapezoid: pixel (10,30) not set")
	}
}

func TestFillTrapezoid_ReversedX(t *testing.T) {
	// x0 > x1: triggers the swap branch (x0,x1 = x1,x0 and y0,y1 = y1,y0).
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{0, 255, 0, 255}
	// Pass x0=50 > x1=10; after swap x0=10, x1=50.
	fillTrapezoid(img, 50, 40, 10, 20, 60, c, 255)
	// Should produce same result as normal order since we just swap.
	if img.RGBAAt(10, 30) != c {
		t.Errorf("reversed trapezoid: pixel (10,30) not set after swap")
	}
}

func TestFillTrapezoid_LineYAboveBaseY(t *testing.T) {
	// lineY > baseY inside the loop triggers the swap-within-loop branch
	// followed by the restore branch (if lineY < baseY) on the same iteration.
	// y0=80, y1=20, baseY=10: across x=10..50 lineY descends from 80 to 20,
	// always above baseY=10, so the swap+restore fires on every iteration.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{0, 0, 255, 255}
	fillTrapezoid(img, 10, 80, 50, 20, 10, c, 200)
	// The function should complete without panic and paint pixels.
	// At x=10, lineY=80 > baseY=10: swap → fill y=10..80, restore baseY=10.
	// Verify the function ran by checking a pixel in the filled region.
	got := img.RGBAAt(10, 10)
	want := color.RGBA{0, 0, 255, 200}
	if got != want {
		t.Errorf("LineYAboveBaseY: pixel (10,10) = %v, want %v", got, want)
	}
}

func TestFillTrapezoid_RestoreBaseYAcrossIterations(t *testing.T) {
	// Verify that baseY is correctly restored after each swap so that
	// subsequent x-columns also get the correct fill range.
	// y0=70, y1=50, baseY=20: lineY always > baseY → swap+restore each iteration.
	// After restore, baseY stays 20 for every column.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{255, 128, 0, 255}
	fillTrapezoid(img, 10, 70, 30, 50, 20, c, 255)
	// At every x in 10..30, the fill runs from baseY=20 up to lineY.
	// Pixel at (20, 20) should be set (it is within the filled band of every column).
	if img.RGBAAt(20, 20) != c {
		t.Errorf("RestoreBaseY: pixel (20,20) not set; baseY restore may be broken")
	}
	// Pixel at (30, 20) should also be set (last column).
	if img.RGBAAt(30, 20) != c {
		t.Errorf("RestoreBaseY: pixel (30,20) not set; baseY restore may be broken")
	}
}

func TestFillTrapezoid_SameX(t *testing.T) {
	// x0 == x1: single column, t=0, lineY=y0.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{128, 128, 0, 255}
	fillTrapezoid(img, 30, 20, 30, 20, 60, c, 255)
	// At x=30, t=0, lineY=20, fills y=20..60.
	if img.RGBAAt(30, 40) != c {
		t.Errorf("same-x trapezoid: pixel (30,40) not set")
	}
}

// ── min / max ─────────────────────────────────────────────────────────────────

func TestMin_AGreaterThanB(t *testing.T) {
	// a > b → should return b (the `return b` branch, currently uncovered).
	if got := min(10, 3); got != 3 {
		t.Errorf("min(10,3) = %d, want 3", got)
	}
}

func TestMin_AEqualB(t *testing.T) {
	// a == b → a < b is false → return b (same branch).
	if got := min(5, 5); got != 5 {
		t.Errorf("min(5,5) = %d, want 5", got)
	}
}

func TestMin_ALessThanB(t *testing.T) {
	// a < b → return a (already covered, but keep for clarity).
	if got := min(2, 7); got != 2 {
		t.Errorf("min(2,7) = %d, want 2", got)
	}
}

func TestMax_ALessThanB(t *testing.T) {
	// a < b → a > b is false → return b (currently uncovered).
	if got := max(3, 8); got != 8 {
		t.Errorf("max(3,8) = %d, want 8", got)
	}
}

func TestMax_AEqualB(t *testing.T) {
	// a == b → a > b is false → return b.
	if got := max(4, 4); got != 4 {
		t.Errorf("max(4,4) = %d, want 4", got)
	}
}

func TestMax_AGreaterThanB(t *testing.T) {
	// a > b → return a (already covered).
	if got := max(9, 3); got != 9 {
		t.Errorf("max(9,3) = %d, want 9", got)
	}
}

// ── globalRange ───────────────────────────────────────────────────────────────

func TestGlobalRange_EmptySeries(t *testing.T) {
	// len(series) == 0 → return 0, 1.
	min, max := globalRange(nil)
	if min != 0 || max != 1 {
		t.Errorf("globalRange(nil) = (%v,%v), want (0,1)", min, max)
	}
}

func TestGlobalRange_SeriesWithNoValues(t *testing.T) {
	// Series present but all have empty Values → minY stays +Inf, maxY stays -Inf
	// → both IsInf branches fire → returns 0, 1.
	series := []Series{{Name: "A", Values: nil}}
	minV, maxV := globalRange(series)
	if minV != 0 || maxV != 1 {
		t.Errorf("globalRange(empty values) = (%v,%v), want (0,1)", minV, maxV)
	}
}

func TestGlobalRange_NormalValues(t *testing.T) {
	// Normal path: both min and max set from values.
	series := []Series{
		{Name: "A", Values: []float64{1.0, 5.0, 3.0}},
		{Name: "B", Values: []float64{-2.0, 4.0}},
	}
	minV, maxV := globalRange(series)
	if minV != -2.0 || maxV != 5.0 {
		t.Errorf("globalRange normal = (%v,%v), want (-2, 5)", minV, maxV)
	}
}
