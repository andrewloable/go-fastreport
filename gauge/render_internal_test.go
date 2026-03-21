package gauge

// render_internal_test.go — white-box tests for unexported render helpers.
// Uses package gauge (internal) to access drawHLine, drawVLine, drawArc directly.

import (
	"image"
	"image/color"
	"testing"
)

// TestDrawHLine_YOutOfBounds tests the early-return branch when y is outside
// the image bounds (y < b.Min.Y or y >= b.Max.Y).
func TestDrawHLine_YOutOfBounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	red := color.RGBA{R: 255, A: 255}

	// y < 0 — below the minimum Y bound (b.Min.Y == 0).
	drawHLine(img, 0, -1, 10, red)
	// y >= b.Max.Y — at or above the maximum Y bound.
	drawHLine(img, 0, 10, 10, red)
	drawHLine(img, 0, 100, 10, red)

	// Verify no pixels were set (the image should remain transparent).
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if img.RGBAAt(x, y) != (color.RGBA{}) {
				t.Errorf("pixel (%d,%d) was set; expected transparent", x, y)
			}
		}
	}
}

// TestDrawHLine_YInBounds verifies that drawing within bounds works correctly.
func TestDrawHLine_YInBounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	red := color.RGBA{R: 255, A: 255}

	// y = 5 is within [0, 10).
	drawHLine(img, 2, 5, 7, red)

	// Check that pixels (2,5), (3,5), (4,5), (5,5), (6,5) are set.
	for x := 2; x < 7; x++ {
		if img.RGBAAt(x, 5) != red {
			t.Errorf("pixel (%d,5) not set after drawHLine", x)
		}
	}
}

// TestDrawVLine_XOutOfBounds tests the early-return branch when x is outside
// the image bounds (x < b.Min.X or x >= b.Max.X).
func TestDrawVLine_XOutOfBounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	blue := color.RGBA{B: 255, A: 255}

	// x < 0 — below the minimum X bound (b.Min.X == 0).
	drawVLine(img, -1, 0, 10, blue)
	// x >= b.Max.X — at or above the maximum X bound.
	drawVLine(img, 10, 0, 10, blue)
	drawVLine(img, 100, 0, 10, blue)

	// Verify no pixels were set.
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if img.RGBAAt(x, y) != (color.RGBA{}) {
				t.Errorf("pixel (%d,%d) was set; expected transparent", x, y)
			}
		}
	}
}

// TestDrawVLine_XInBounds verifies that drawing within bounds works correctly.
func TestDrawVLine_XInBounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	blue := color.RGBA{B: 255, A: 255}

	// x = 5 is within [0, 10).
	drawVLine(img, 5, 2, 7, blue)

	// Check that pixels (5,2), (5,3), (5,4), (5,5), (5,6) are set.
	for y := 2; y < 7; y++ {
		if img.RGBAAt(5, y) != blue {
			t.Errorf("pixel (5,%d) not set after drawVLine", y)
		}
	}
}

// TestDrawArc_ZeroOrNegativeRadius tests the early-return branch when rx or ry
// is zero or negative (rx <= 0 || ry <= 0).
func TestDrawArc_ZeroOrNegativeRadius(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	green := color.RGBA{G: 255, A: 255}

	// rx == 0 — should return immediately without drawing.
	drawArc(img, 50, 50, 0, 10, 0, 360, green)
	// ry == 0 — should return immediately without drawing.
	drawArc(img, 50, 50, 10, 0, 0, 360, green)
	// rx < 0 — should return immediately.
	drawArc(img, 50, 50, -1, 10, 0, 360, green)
	// ry < 0 — should return immediately.
	drawArc(img, 50, 50, 10, -5, 0, 360, green)

	// Verify no pixels were set.
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if img.RGBAAt(x, y) != (color.RGBA{}) {
				t.Errorf("pixel (%d,%d) was set; expected transparent after zero/negative radius arc", x, y)
			}
		}
	}
}

// TestDrawArc_PositiveRadius verifies that a valid arc draws pixels.
func TestDrawArc_PositiveRadius(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	green := color.RGBA{G: 255, A: 255}

	// Draw a quarter arc — should produce some pixels.
	drawArc(img, 50, 50, 30, 30, 0, 90, green)

	// At least one pixel should be set.
	found := false
	for y := 0; y < 100 && !found; y++ {
		for x := 0; x < 100 && !found; x++ {
			if img.RGBAAt(x, y) == green {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected at least one pixel to be set by drawArc with positive radius")
	}
}

// TestDrawArc_NegativeSweep tests the sweep < 0 branch (sweep += 360).
func TestDrawArc_NegativeSweep(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	green := color.RGBA{G: 255, A: 255}

	// startDeg=90, endDeg=-90: sweep = -90-90 = -180 < 0 → sweep becomes 180.
	drawArc(img, 50, 50, 30, 30, 90, -90, green)

	// At least one pixel should be drawn.
	found := false
	for y := 0; y < 100 && !found; y++ {
		for x := 0; x < 100 && !found; x++ {
			if img.RGBAAt(x, y) == green {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected pixels to be drawn by drawArc with negative sweep")
	}
}

// ── drawRadialScaleTicks guard conditions ─────────────────────────────────────

// TestDrawRadialScaleTicks_ZeroRadius tests the early-return guard (radius <= 0).
func TestDrawRadialScaleTicks_ZeroRadius(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	tick := color.RGBA{R: 255, A: 255}
	// radius = 0 → should return immediately without drawing.
	drawRadialScaleTicks(img, 25, 25, 0, -135, 135, 11, 4, tick)
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			if img.RGBAAt(x, y) != (color.RGBA{}) {
				t.Errorf("pixel (%d,%d) set with zero radius", x, y)
			}
		}
	}
}

// TestDrawRadialScaleTicks_MajorCountLessThan2 tests early-return for majorCount < 2.
func TestDrawRadialScaleTicks_MajorCountLessThan2(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	tick := color.RGBA{R: 255, A: 255}
	drawRadialScaleTicks(img, 25, 25, 20, -135, 135, 1, 4, tick)
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			if img.RGBAAt(x, y) != (color.RGBA{}) {
				t.Errorf("pixel (%d,%d) set with majorCount=1", x, y)
			}
		}
	}
}

// TestDrawRadialScaleTicks_NegativeSweep tests the sweep < 0 correction branch.
func TestDrawRadialScaleTicks_NegativeSweep(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	tick := color.RGBA{B: 255, A: 255}
	// startAngle > endAngle forces sweep < 0 → sweep += 360.
	drawRadialScaleTicks(img, 50, 50, 40, 90, -90, 5, 3, tick)
	found := false
	for y := 0; y < 100 && !found; y++ {
		for x := 0; x < 100 && !found; x++ {
			if img.RGBAAt(x, y) == tick {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected tick pixels with negative sweep")
	}
}

// ── drawRadialLabelMarkers guard conditions ───────────────────────────────────

// TestDrawRadialLabelMarkers_ZeroRadius tests early-return guard.
func TestDrawRadialLabelMarkers_ZeroRadius(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	mark := color.RGBA{G: 255, A: 255}
	drawRadialLabelMarkers(img, 25, 25, 0, -135, 135, 11, mark)
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			if img.RGBAAt(x, y) != (color.RGBA{}) {
				t.Errorf("pixel (%d,%d) set with zero labelRadius", x, y)
			}
		}
	}
}

// TestDrawRadialLabelMarkers_NegativeSweep tests the sweep < 0 correction branch.
func TestDrawRadialLabelMarkers_NegativeSweep(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	mark := color.RGBA{G: 200, A: 255}
	drawRadialLabelMarkers(img, 50, 50, 40, 90, -90, 5, mark)
	found := false
	for y := 0; y < 100 && !found; y++ {
		for x := 0; x < 100 && !found; x++ {
			if img.RGBAAt(x, y) == mark {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected marker pixels with negative sweep")
	}
}

// TestDrawRadialLabelMarkers_MajorCountLessThan2 tests early return for majorCount < 2.
func TestDrawRadialLabelMarkers_MajorCountLessThan2(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	mark := color.RGBA{G: 255, A: 255}
	drawRadialLabelMarkers(img, 25, 25, 20, -135, 135, 1, mark)
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			if img.RGBAAt(x, y) != (color.RGBA{}) {
				t.Errorf("pixel (%d,%d) set with majorCount=1", x, y)
			}
		}
	}
}

// ── drawRadialPointerNeedle edge case ─────────────────────────────────────────

// TestDrawRadialPointerNeedle_TinyRadius ensures the needleLen >= 1 guard fires.
// When radius is very small, needleLen would be 0; the guard clamps it to 1.
func TestDrawRadialPointerNeedle_TinyRadius(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	ptr := color.RGBA{R: 200, A: 255}
	// radius = 1, so 0.82*1 ≈ 0 (rounds to 0) → guard sets needleLen = 1.
	drawRadialPointerNeedle(img, 5, 5, 0, 1, ptr)
	// Should not panic; just verify at least the center pixel may be set.
}
