// sparkline_coverage_test.go — targeted tests for 100% coverage on:
//   - parseFloat: empty-after-trim branch
//   - drawLine: dx==0&&dy==0 branch, dy-dominant branch
//   - drawVLine: y0>y1 swap branch
//   - scaleY: vmax==vmin branch (white-box via internal package)
//   - scaleX: n<=1 branch (white-box via internal package)

package sparkline

import (
	"image"
	"image/color"
	"testing"
)

// ── parseFloat: whitespace-only string → s=="" after TrimSpace ────────────────

func TestParseFloat_WhitespaceOnly(t *testing.T) {
	// TrimSpace("   ") == "" → early return 0.
	got := parseFloat("   ")
	if got != 0 {
		t.Errorf("parseFloat(whitespace) = %v, want 0", got)
	}
}

func TestParseFloat_EmptyString(t *testing.T) {
	got := parseFloat("")
	if got != 0 {
		t.Errorf("parseFloat(\"\") = %v, want 0", got)
	}
}

// ── drawLine: zero-length segment (dx==0 && dy==0) ───────────────────────────

func TestDrawLine_ZeroLength(t *testing.T) {
	img := newImg(10, 10)
	c := color.RGBA{R: 255, A: 255}
	// Call drawLine with same start and end point.
	drawLine(img, 5, 5, 5, 5, c)
	// Pixel (5,5) should be set.
	got := img.RGBAAt(5, 5)
	if got.R != 255 {
		t.Errorf("zero-length drawLine: pixel (5,5) R=%d, want 255", got.R)
	}
}

// ── drawLine: dy-dominant segment (|dy| > |dx|) ───────────────────────────────

func TestDrawLine_DyDominant(t *testing.T) {
	// dy=10, dx=1 → |dy| > |dx| → steps = |dy|.
	img := newImg(20, 20)
	c := color.RGBA{G: 255, A: 255}
	drawLine(img, 5, 0, 6, 10, c)
	// Several pixels along the nearly-vertical line should be set.
	found := false
	for y := 0; y <= 10; y++ {
		p := img.RGBAAt(5, y)
		q := img.RGBAAt(6, y)
		if p.G == 255 || q.G == 255 {
			found = true
			break
		}
	}
	if !found {
		t.Error("dy-dominant drawLine: no green pixels found")
	}
}

// ── drawVLine: y0 > y1 swap ───────────────────────────────────────────────────

func TestDrawVLine_SwapNeeded(t *testing.T) {
	img := newImg(10, 10)
	c := color.RGBA{B: 255, A: 255}
	// y0=8, y1=3 → y0 > y1 → swap.
	drawVLine(img, 5, 8, 3, c)
	// Pixels (5,3) through (5,8) should be set.
	for y := 3; y <= 8; y++ {
		p := img.RGBAAt(5, y)
		if p.B != 255 {
			t.Errorf("drawVLine(swap): pixel (5,%d) B=%d, want 255", y, p.B)
		}
	}
}

// ── scaleY: vmax == vmin → return h/2 ────────────────────────────────────────

func TestScaleY_ZeroRange(t *testing.T) {
	// vmax == vmin → returns h/2.
	h := 40
	got := scaleY(5.0, 5.0, 5.0, h)
	want := h / 2
	if got != want {
		t.Errorf("scaleY zero range = %d, want %d", got, want)
	}
}

func TestScaleY_ZeroRange_OddHeight(t *testing.T) {
	h := 41
	got := scaleY(7.0, 7.0, 7.0, h)
	want := h / 2 // integer division
	if got != want {
		t.Errorf("scaleY zero range odd height = %d, want %d", got, want)
	}
}

// ── scaleX: n <= 1 → return w/2 ──────────────────────────────────────────────

func TestScaleX_SinglePoint(t *testing.T) {
	w := 100
	got := scaleX(0, 1, w)
	want := w / 2
	if got != want {
		t.Errorf("scaleX n=1 = %d, want %d", got, want)
	}
}

func TestScaleX_ZeroPoints(t *testing.T) {
	// n <= 1 (n==0 edge) → returns w/2.
	w := 80
	got := scaleX(0, 0, w)
	want := w / 2
	if got != want {
		t.Errorf("scaleX n=0 = %d, want %d", got, want)
	}
}

// ── Render-level test: ensure drawLine dx==0&&dy==0 is hit via Render ─────────

func TestRenderLine_TinyCanvas_ZeroLengthSegment(t *testing.T) {
	// With w=5, n=3 points, scaleX(0,3,5)=scaleX(1,3,5)=2 (same x pixel).
	// With constant first two values, dy=0 too → dx==0 && dy==0 branch.
	s := &Series{Type: ChartTypeLine, Values: []float64{5, 5, 8}}
	img := Render(s, 5, 10)
	if img == nil {
		t.Fatal("Render tiny canvas returned nil")
	}
	b := img.Bounds()
	if b.Dx() != 5 || b.Dy() != 10 {
		t.Errorf("size = %dx%d, want 5x10", b.Dx(), b.Dy())
	}
}

// ── Render-level test: dy-dominant line via tall narrow canvas ─────────────────

func TestRenderLine_DyDominant_TallNarrow(t *testing.T) {
	// With w=3 and values [0,100], dx is small but dy is large.
	// scaleX(0,2,3)=0, scaleX(1,2,3)=3 → dx=3
	// scaleY(0,0,100,100)=98, scaleY(100,0,100,100)=2 → dy=96
	// |dy|=96 > |dx|=3 → dy-dominant.
	s := &Series{Type: ChartTypeLine, Values: []float64{0, 100}}
	img := Render(s, 3, 100)
	if img == nil {
		t.Fatal("Render dy-dominant returned nil")
	}
	// At least some non-white pixels expected.
	found := false
	rgba := img.(*image.RGBA)
	b := rgba.Bounds()
	for y := b.Min.Y; y < b.Max.Y && !found; y++ {
		for x := b.Min.X; x < b.Max.X && !found; x++ {
			p := rgba.RGBAAt(x, y)
			if p.B == 0xC0 { // lineColor blue component
				found = true
			}
		}
	}
	if !found {
		t.Error("dy-dominant render: no line pixels found")
	}
}
