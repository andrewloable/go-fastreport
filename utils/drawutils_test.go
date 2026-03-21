package utils_test

import (
	"math"
	"testing"

	"github.com/andrewloable/go-fastreport/utils"
)

// ── ScreenDpi / ScreenDpiFX ──────────────────────────────────────────────────

func TestScreenDpi(t *testing.T) {
	got := utils.ScreenDpi()
	if got != 96 {
		t.Errorf("ScreenDpi() = %d, want 96", got)
	}
}

func TestScreenDpiFX(t *testing.T) {
	// ScreenDpiFX = 96 / ScreenDpi = 96 / 96 = 1.0
	got := utils.ScreenDpiFX()
	if got != 1.0 {
		t.Errorf("ScreenDpiFX() = %v, want 1.0", got)
	}
}

func TestScreenDpiFX_Idempotent(t *testing.T) {
	// Multiple calls return the same value.
	a := utils.ScreenDpiFX()
	b := utils.ScreenDpiFX()
	if a != b {
		t.Errorf("ScreenDpiFX not idempotent: %v != %v", a, b)
	}
}

// ── UIScale ──────────────────────────────────────────────────────────────────

func TestUIScale_Default(t *testing.T) {
	utils.SetUIScale(1.0) // reset to default
	got := utils.UIScale()
	if got != 1.0 {
		t.Errorf("UIScale default = %v, want 1.0", got)
	}
}

func TestSetUIScale_Clamp_Below(t *testing.T) {
	utils.SetUIScale(0.5) // below minimum
	got := utils.UIScale()
	if got != 1.0 {
		t.Errorf("SetUIScale(0.5) clamped to %v, want 1.0", got)
	}
}

func TestSetUIScale_Clamp_Above(t *testing.T) {
	utils.SetUIScale(2.0) // above maximum
	got := utils.UIScale()
	if got != 1.5 {
		t.Errorf("SetUIScale(2.0) clamped to %v, want 1.5", got)
	}
}

func TestSetUIScale_ValidRange(t *testing.T) {
	utils.SetUIScale(1.25)
	got := utils.UIScale()
	if got != 1.25 {
		t.Errorf("SetUIScale(1.25) = %v, want 1.25", got)
	}
	utils.SetUIScale(1.0) // restore default
}

func TestSetUIScale_Boundary_Min(t *testing.T) {
	utils.SetUIScale(1.0)
	if utils.UIScale() != 1.0 {
		t.Errorf("SetUIScale(1.0): got %v, want 1.0", utils.UIScale())
	}
}

func TestSetUIScale_Boundary_Max(t *testing.T) {
	utils.SetUIScale(1.5)
	if utils.UIScale() != 1.5 {
		t.Errorf("SetUIScale(1.5): got %v, want 1.5", utils.UIScale())
	}
	utils.SetUIScale(1.0) // restore default
}

// ── Unit conversions ─────────────────────────────────────────────────────────

func almostEqualF32(a, b, tol float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= tol
}

func TestPixelsToMM(t *testing.T) {
	// 96px = 25.4mm (1 inch)
	got := utils.PixelsToMM(96)
	want := float32(25.4)
	if !almostEqualF32(got, want, 0.01) {
		t.Errorf("PixelsToMM(96) = %v, want %v", got, want)
	}
}

func TestMMToPixels(t *testing.T) {
	// 25.4mm = 96px (1 inch)
	got := utils.MMToPixels(25.4)
	want := float32(96)
	if !almostEqualF32(got, want, 0.01) {
		t.Errorf("MMToPixels(25.4) = %v, want %v", got, want)
	}
}

func TestMMPixelsRoundTrip(t *testing.T) {
	orig := float32(210) // A4 width
	px := utils.MMToPixels(orig)
	got := utils.PixelsToMM(px)
	if !almostEqualF32(got, orig, 0.001) {
		t.Errorf("MM→px→MM round-trip: got %v, want %v", got, orig)
	}
}

func TestPixelsToInches(t *testing.T) {
	got := utils.PixelsToInches(96)
	want := float32(1.0)
	if !almostEqualF32(got, want, 0.001) {
		t.Errorf("PixelsToInches(96) = %v, want %v", got, want)
	}
}

func TestInchesToPixels(t *testing.T) {
	got := utils.InchesToPixels(1.0)
	want := float32(96)
	if !almostEqualF32(got, want, 0.001) {
		t.Errorf("InchesToPixels(1.0) = %v, want %v", got, want)
	}
}

func TestInchesPixelsRoundTrip(t *testing.T) {
	orig := float32(8.5) // letter width
	px := utils.InchesToPixels(orig)
	got := utils.PixelsToInches(px)
	if !almostEqualF32(got, orig, 0.001) {
		t.Errorf("In→px→In round-trip: got %v, want %v", got, orig)
	}
}

// ── RectF ────────────────────────────────────────────────────────────────────

func TestRectF_Right(t *testing.T) {
	r := utils.RectF{X: 10, Y: 20, Width: 100, Height: 50}
	if r.Right() != 110 {
		t.Errorf("RectF.Right() = %v, want 110", r.Right())
	}
}

func TestRectF_Bottom(t *testing.T) {
	r := utils.RectF{X: 10, Y: 20, Width: 100, Height: 50}
	if r.Bottom() != 70 {
		t.Errorf("RectF.Bottom() = %v, want 70", r.Bottom())
	}
}

func TestRectF_IsEmpty_Zero(t *testing.T) {
	r := utils.RectF{}
	if !r.IsEmpty() {
		t.Error("zero RectF should be empty")
	}
}

func TestRectF_IsEmpty_ZeroWidth(t *testing.T) {
	r := utils.RectF{X: 5, Y: 5, Width: 0, Height: 10}
	if !r.IsEmpty() {
		t.Error("RectF with Width=0 should be empty")
	}
}

func TestRectF_IsEmpty_ZeroHeight(t *testing.T) {
	r := utils.RectF{X: 5, Y: 5, Width: 10, Height: 0}
	if !r.IsEmpty() {
		t.Error("RectF with Height=0 should be empty")
	}
}

func TestRectF_IsEmpty_NonEmpty(t *testing.T) {
	r := utils.RectF{X: 5, Y: 5, Width: 10, Height: 10}
	if r.IsEmpty() {
		t.Error("non-zero RectF should not be empty")
	}
}

// ── SizeF ────────────────────────────────────────────────────────────────────

func TestSizeF_IsEmpty_Zero(t *testing.T) {
	s := utils.SizeF{}
	if !s.IsEmpty() {
		t.Error("zero SizeF should be empty")
	}
}

func TestSizeF_IsEmpty_NonZero(t *testing.T) {
	s := utils.SizeF{Width: 10, Height: 5}
	if s.IsEmpty() {
		t.Error("non-zero SizeF should not be empty")
	}
}

func TestSizeF_IsEmpty_OnlyWidth(t *testing.T) {
	s := utils.SizeF{Width: 10, Height: 0}
	// IsEmpty returns true only when both are zero.
	if s.IsEmpty() {
		t.Error("SizeF with non-zero Width should not be empty")
	}
}

// ── PointF ───────────────────────────────────────────────────────────────────

func TestPointF_ZeroValue(t *testing.T) {
	p := utils.PointF{}
	if p.X != 0 || p.Y != 0 {
		t.Errorf("PointF zero value: got (%v, %v), want (0, 0)", p.X, p.Y)
	}
}

func TestPointF_Assignment(t *testing.T) {
	p := utils.PointF{X: 3.14, Y: 2.72}
	if !almostEqualF32(p.X, 3.14, 0.001) || !almostEqualF32(p.Y, 2.72, 0.001) {
		t.Errorf("PointF assignment: got (%v, %v)", p.X, p.Y)
	}
}

// ── RoundFloat64 ─────────────────────────────────────────────────────────────

func TestRoundFloat64_TwoDecimal(t *testing.T) {
	got := utils.RoundFloat64(3.14159, 2)
	want := 3.14
	if math.Abs(got-want) > 1e-10 {
		t.Errorf("RoundFloat64(3.14159, 2) = %v, want %v", got, want)
	}
}

func TestRoundFloat64_Zero(t *testing.T) {
	got := utils.RoundFloat64(0, 3)
	if got != 0 {
		t.Errorf("RoundFloat64(0, 3) = %v, want 0", got)
	}
}

func TestRoundFloat64_RoundUp(t *testing.T) {
	got := utils.RoundFloat64(2.555, 2)
	// math.Round rounds 2.555*100=255.5 → 256, so result = 2.56
	want := 2.56
	if math.Abs(got-want) > 1e-10 {
		t.Errorf("RoundFloat64(2.555, 2) = %v, want %v", got, want)
	}
}

func TestRoundFloat64_NoDecimal(t *testing.T) {
	got := utils.RoundFloat64(4.7, 0)
	want := 5.0
	if math.Abs(got-want) > 1e-10 {
		t.Errorf("RoundFloat64(4.7, 0) = %v, want %v", got, want)
	}
}

func TestRoundFloat64_Negative(t *testing.T) {
	got := utils.RoundFloat64(-1.235, 2)
	want := -1.24
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("RoundFloat64(-1.235, 2) = %v, want %v", got, want)
	}
}

// ── Zero-value unit conversions ──────────────────────────────────────────────

func TestConversions_Zero(t *testing.T) {
	if utils.PixelsToMM(0) != 0 {
		t.Error("PixelsToMM(0) should be 0")
	}
	if utils.MMToPixels(0) != 0 {
		t.Error("MMToPixels(0) should be 0")
	}
	if utils.PixelsToInches(0) != 0 {
		t.Error("PixelsToInches(0) should be 0")
	}
	if utils.InchesToPixels(0) != 0 {
		t.Error("InchesToPixels(0) should be 0")
	}
}
