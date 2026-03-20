// Internal (white-box) tests for drawBezierSegment. These live in package
// chart (not chart_test) so they can call the unexported function directly.
package chart

import (
	"image"
	"image/color"
	"testing"
)

// TestDrawBezierSegment_TwoPoints calls drawBezierSegment with n=2 (xs/ys
// have exactly two entries). This exercises:
//   - tangentX/Y for k==0  (left-endpoint clamp: return xs[1]-xs[0])
//   - tangentX/Y for k==1  (right-endpoint clamp: return xs[n-1]-xs[n-2])
//
// and the full cubic Bezier loop.
func TestDrawBezierSegment_TwoPoints(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	col := color.RGBA{255, 0, 0, 255}
	xs := []int{20, 180}
	ys := []int{100, 100}
	// Should not panic; draws a horizontal line via Bezier.
	drawBezierSegment(img, xs, ys, 0, 1, col)
	// Verify some pixels are drawn (the midpoint should be red).
	got := img.RGBAAt(100, 100)
	if got != col {
		t.Errorf("midpoint pixel = %v, want %v", got, col)
	}
}

// TestDrawBezierSegment_ThreePoints_Interior calls with n=3, i0=0,i1=1 and
// then i0=1,i1=2, exercising:
//   - k==0 left-endpoint clamp
//   - k==1 interior formula ((xs[2]-xs[0])/2)
//   - k==2 right-endpoint clamp
func TestDrawBezierSegment_ThreePoints_Interior(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 300, 200))
	col := color.RGBA{0, 200, 0, 255}
	xs := []int{10, 150, 290}
	ys := []int{100, 50, 100}
	// Segment 0→1.
	drawBezierSegment(img, xs, ys, 0, 1, col)
	// Segment 1→2 (i0==n-2, i1==n-1 exercises right-end clamp via tangentX(2) and tangentX(1) interior).
	drawBezierSegment(img, xs, ys, 1, 2, col)
	// Just verify no panic and some non-background pixels exist.
	found := false
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y && !found; y++ {
		for x := b.Min.X; x < b.Max.X && !found; x++ {
			if img.RGBAAt(x, y) == col {
				found = true
			}
		}
	}
	if !found {
		t.Error("drawBezierSegment three points: no pixels drawn")
	}
}

// TestDrawBezierSegment_Diagonal tests a diagonal segment to exercise
// the `steps` calculation (based on Euclidean distance).
func TestDrawBezierSegment_Diagonal(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	col := color.RGBA{0, 0, 255, 255}
	xs := []int{10, 100, 190}
	ys := []int{10, 100, 10}
	// Segment 0→1 (diagonal, large dx and dy → many steps).
	drawBezierSegment(img, xs, ys, 0, 1, col)
	// Some pixel near the midpoint should be blue.
	found := false
	for y := 40; y <= 80; y++ {
		for x := 40; x <= 80; x++ {
			if img.RGBAAt(x, y) == col {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("drawBezierSegment diagonal: no pixels drawn near midpoint")
	}
}

// TestDrawBezierSegment_SamePoint tests i0==i1 (degenerate: start==end).
// dx==0, dy==0 → steps clamped to 8; the loop runs 1..8 and all Bezier
// evaluations produce the same point. Should not panic.
func TestDrawBezierSegment_SamePoint(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	col := color.RGBA{128, 128, 128, 255}
	xs := []int{50, 50}
	ys := []int{50, 50}
	drawBezierSegment(img, xs, ys, 0, 1, col)
	// Pixel at (50,50) should be set.
	got := img.RGBAAt(50, 50)
	if got != col {
		t.Errorf("same-point segment: pixel (50,50) = %v, want %v", got, col)
	}
}

// TestDrawBezierSegment_ManyPoints exercises the interior tangent formula for
// a slice of n=5 points, covering all three branches in tangentX/Y:
//   - k==0 (left end)
//   - 0 < k < n-1 (interior: (xs[k+1]-xs[k-1])/2)
//   - k==n-1 (right end)
func TestDrawBezierSegment_ManyPoints(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 200))
	col := color.RGBA{200, 100, 50, 255}
	xs := []int{10, 100, 200, 300, 390}
	ys := []int{100, 50, 150, 50, 100}
	// Draw all 4 segments.
	for i := 1; i < len(xs); i++ {
		drawBezierSegment(img, xs, ys, i-1, i, col)
	}
	// Verify at least one pixel was drawn.
	found := false
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y && !found; y++ {
		for x := b.Min.X; x < b.Max.X && !found; x++ {
			if img.RGBAAt(x, y) == col {
				found = true
			}
		}
	}
	if !found {
		t.Error("drawBezierSegment many points: no pixels drawn")
	}
}
