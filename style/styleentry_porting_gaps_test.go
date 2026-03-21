package style_test

// styleentry_porting_gaps_test.go — tests for StyleEntry.Assign, Clone,
// EffectiveFill, and EffectiveTextFill, implemented while reviewing porting
// gaps in:
//   - FastReport.Base/StyleBase.cs  (Assign method)
//   - FastReport.Base/Style.cs      (Clone method)
//
// These mirror the C# StyleBase.Assign(StyleBase source) and Style.Clone().

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

// ── StyleEntry.Assign ─────────────────────────────────────────────────────────

func TestStyleEntry_Assign_NilSource_NoOp(t *testing.T) {
	e := &style.StyleEntry{Name: "orig"}
	e.Assign(nil) // must not panic
	if e.Name != "orig" {
		t.Error("Assign(nil) should not change the target")
	}
}

func TestStyleEntry_Assign_CopiesAllFields(t *testing.T) {
	src := &style.StyleEntry{
		Name:               "Src",
		ApplyBorder:        true,
		ApplyFill:          true,
		ApplyTextFill:      true,
		ApplyFont:          true,
		FillColor:          color.RGBA{R: 255, A: 255},
		TextColor:          color.RGBA{G: 255, A: 255},
		FontChanged:        true,
		TextColorChanged:   true,
		FillColorChanged:   true,
		BorderColorChanged: true,
		BorderColor:        color.RGBA{B: 255, A: 255},
		Font:               style.Font{Name: "Arial", Size: 12},
	}
	src.Fill = style.NewLinearGradientFill(color.RGBA{R: 255, A: 255}, color.RGBA{B: 255, A: 255})
	src.TextFill = style.NewSolidFill(color.RGBA{G: 128, A: 255})

	dst := &style.StyleEntry{}
	dst.Assign(src)

	if dst.Name != "Src" {
		t.Errorf("Name = %q, want Src", dst.Name)
	}
	if !dst.ApplyBorder || !dst.ApplyFill || !dst.ApplyTextFill || !dst.ApplyFont {
		t.Error("Apply* flags not copied")
	}
	if dst.FillColor != src.FillColor {
		t.Errorf("FillColor = %v, want %v", dst.FillColor, src.FillColor)
	}
	if dst.TextColor != src.TextColor {
		t.Errorf("TextColor = %v, want %v", dst.TextColor, src.TextColor)
	}
	if !dst.FontChanged || dst.Font.Name != "Arial" || dst.Font.Size != 12 {
		t.Errorf("Font not copied correctly: %+v", dst.Font)
	}
	if !dst.FillColorChanged || !dst.TextColorChanged || !dst.BorderColorChanged {
		t.Error("legacy Changed flags not copied")
	}
	if dst.BorderColor != src.BorderColor {
		t.Errorf("BorderColor = %v, want %v", dst.BorderColor, src.BorderColor)
	}
}

// TestStyleEntry_Assign_FillIsDeepCopy verifies that mutating the source Fill
// after Assign does not affect the destination.
func TestStyleEntry_Assign_FillIsDeepCopy(t *testing.T) {
	origColor := color.RGBA{R: 100, G: 0, B: 0, A: 255}
	src := &style.StyleEntry{
		ApplyFill: true,
		Fill:      style.NewSolidFill(origColor),
	}
	dst := &style.StyleEntry{}
	dst.Assign(src)

	// Mutate the source fill.
	src.Fill.(*style.SolidFill).Color = color.RGBA{R: 255, A: 255}

	// Destination fill should still have the original colour.
	if dst.Fill == nil {
		t.Fatal("dst.Fill should not be nil after Assign")
	}
	sf, ok := dst.Fill.(*style.SolidFill)
	if !ok {
		t.Fatalf("dst.Fill type = %T, want *style.SolidFill", dst.Fill)
	}
	if sf.Color != origColor {
		t.Errorf("dst.Fill.Color = %v, want original %v", sf.Color, origColor)
	}
}

// TestStyleEntry_Assign_NilFillCopied verifies that a nil Fill in the source
// results in a nil Fill in the destination.
func TestStyleEntry_Assign_NilFillCopied(t *testing.T) {
	src := &style.StyleEntry{ApplyFill: true, Fill: nil}
	dst := &style.StyleEntry{Fill: style.NewSolidFill(color.RGBA{R: 1, A: 1})}
	dst.Assign(src)
	if dst.Fill != nil {
		t.Errorf("dst.Fill should be nil after Assign from nil-fill source, got %T", dst.Fill)
	}
}

// ── StyleEntry.Clone ──────────────────────────────────────────────────────────

func TestStyleEntry_Clone_ProducesIndependentCopy(t *testing.T) {
	src := &style.StyleEntry{
		Name:      "Original",
		ApplyFill: true,
		FillColor: color.RGBA{R: 50, A: 255},
		Font:      style.Font{Name: "Times", Size: 14},
	}
	clone := src.Clone()

	if clone == src {
		t.Fatal("Clone must return a different pointer")
	}
	if clone.Name != src.Name {
		t.Errorf("clone.Name = %q, want %q", clone.Name, src.Name)
	}
	if clone.FillColor != src.FillColor {
		t.Errorf("clone.FillColor = %v, want %v", clone.FillColor, src.FillColor)
	}
	if clone.Font.Name != src.Font.Name {
		t.Errorf("clone.Font.Name = %q, want %q", clone.Font.Name, src.Font.Name)
	}

	// Mutating the clone must not affect the source.
	clone.Name = "Mutated"
	if src.Name != "Original" {
		t.Error("mutating clone.Name affected source.Name")
	}
}

func TestStyleEntry_Clone_WithGradientFill(t *testing.T) {
	grad := style.NewLinearGradientFill(
		color.RGBA{R: 255, A: 255},
		color.RGBA{B: 255, A: 255},
	)
	src := &style.StyleEntry{
		Name:      "GradStyle",
		ApplyFill: true,
		Fill:      grad,
	}
	clone := src.Clone()

	lf, ok := clone.Fill.(*style.LinearGradientFill)
	if !ok {
		t.Fatalf("clone.Fill type = %T, want *style.LinearGradientFill", clone.Fill)
	}
	if lf.StartColor != grad.StartColor {
		t.Errorf("clone gradient StartColor = %v, want %v", lf.StartColor, grad.StartColor)
	}
	// Deep-copy check: mutate source gradient.
	grad.StartColor = color.RGBA{G: 255, A: 255}
	if lf.StartColor == grad.StartColor {
		t.Error("clone gradient should be independent of source after Clone")
	}
}

// ── StyleEntry.EffectiveFill ──────────────────────────────────────────────────

func TestEffectiveFill_NeitherFlagSet_ReturnsNil(t *testing.T) {
	e := &style.StyleEntry{
		// ApplyFill=false (zero), FillColorChanged=false (zero)
		FillColor: color.RGBA{R: 255, A: 255},
	}
	if f := e.EffectiveFill(); f != nil {
		t.Errorf("EffectiveFill with no flags = %v, want nil", f)
	}
}

func TestEffectiveFill_ApplyFill_FillNil_ReturnsSolidFill(t *testing.T) {
	want := color.RGBA{R: 128, G: 64, B: 32, A: 255}
	e := &style.StyleEntry{
		ApplyFill: true,
		FillColor: want,
	}
	f := e.EffectiveFill()
	if f == nil {
		t.Fatal("EffectiveFill should not be nil when ApplyFill is true")
	}
	sf, ok := f.(*style.SolidFill)
	if !ok {
		t.Fatalf("EffectiveFill type = %T, want *style.SolidFill", f)
	}
	if sf.Color != want {
		t.Errorf("SolidFill.Color = %v, want %v", sf.Color, want)
	}
}

func TestEffectiveFill_FillSet_ReturnsFillInterface(t *testing.T) {
	grad := style.NewLinearGradientFill(
		color.RGBA{R: 255, A: 255},
		color.RGBA{B: 255, A: 255},
	)
	e := &style.StyleEntry{
		ApplyFill: true,
		Fill:      grad,
		FillColor: color.RGBA{R: 200, A: 255}, // should be ignored when Fill is set
	}
	f := e.EffectiveFill()
	if _, ok := f.(*style.LinearGradientFill); !ok {
		t.Fatalf("EffectiveFill type = %T, want *style.LinearGradientFill", f)
	}
}

func TestEffectiveFill_LegacyFillColorChanged(t *testing.T) {
	want := color.RGBA{G: 200, A: 255}
	e := &style.StyleEntry{
		FillColorChanged: true,
		FillColor:        want,
	}
	f := e.EffectiveFill()
	if f == nil {
		t.Fatal("EffectiveFill should not be nil when FillColorChanged is true")
	}
	sf, ok := f.(*style.SolidFill)
	if !ok {
		t.Fatalf("EffectiveFill type = %T, want *style.SolidFill", f)
	}
	if sf.Color != want {
		t.Errorf("SolidFill.Color = %v, want %v", sf.Color, want)
	}
}

// ── StyleEntry.EffectiveTextFill ──────────────────────────────────────────────

func TestEffectiveTextFill_NeitherFlagSet_ReturnsNil(t *testing.T) {
	e := &style.StyleEntry{
		TextColor: color.RGBA{R: 255, A: 255},
	}
	if f := e.EffectiveTextFill(); f != nil {
		t.Errorf("EffectiveTextFill with no flags = %v, want nil", f)
	}
}

func TestEffectiveTextFill_ApplyTextFill_TextFillNil_ReturnsSolidFill(t *testing.T) {
	want := color.RGBA{R: 0, G: 0, B: 0, A: 255} // black text
	e := &style.StyleEntry{
		ApplyTextFill: true,
		TextColor:     want,
	}
	f := e.EffectiveTextFill()
	if f == nil {
		t.Fatal("EffectiveTextFill should not be nil when ApplyTextFill is true")
	}
	sf, ok := f.(*style.SolidFill)
	if !ok {
		t.Fatalf("EffectiveTextFill type = %T, want *style.SolidFill", f)
	}
	if sf.Color != want {
		t.Errorf("SolidFill.Color = %v, want %v", sf.Color, want)
	}
}

func TestEffectiveTextFill_TextFillSet_ReturnsInterface(t *testing.T) {
	grad := style.NewLinearGradientFill(
		color.RGBA{R: 255, A: 255},
		color.RGBA{B: 255, A: 255},
	)
	e := &style.StyleEntry{
		ApplyTextFill: true,
		TextFill:      grad,
		TextColor:     color.RGBA{A: 255}, // should be ignored
	}
	f := e.EffectiveTextFill()
	if _, ok := f.(*style.LinearGradientFill); !ok {
		t.Fatalf("EffectiveTextFill type = %T, want *style.LinearGradientFill", f)
	}
}

func TestEffectiveTextFill_LegacyTextColorChanged(t *testing.T) {
	want := color.RGBA{B: 200, A: 255}
	e := &style.StyleEntry{
		TextColorChanged: true,
		TextColor:        want,
	}
	f := e.EffectiveTextFill()
	if f == nil {
		t.Fatal("EffectiveTextFill should not be nil when TextColorChanged is true")
	}
	sf, ok := f.(*style.SolidFill)
	if !ok {
		t.Fatalf("EffectiveTextFill type = %T, want *style.SolidFill", f)
	}
	if sf.Color != want {
		t.Errorf("SolidFill.Color = %v, want %v", sf.Color, want)
	}
}
