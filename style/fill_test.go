package style

import (
	"image/color"
	"testing"
)

// ---------------------------------------------------------------------------
// FillType constants
// ---------------------------------------------------------------------------

func TestFillType_Values(t *testing.T) {
	tests := []struct {
		name string
		ft   FillType
		want string
	}{
		{"Solid", FillTypeSolid, "Solid"},
		{"Linear", FillTypeLinear, "Linear"},
		{"Glass", FillTypeGlass, "Glass"},
		{"Hatch", FillTypeHatch, "Hatch"},
		{"None", FillTypeNone, "None"},
	}
	for _, tt := range tests {
		if string(tt.ft) != tt.want {
			t.Errorf("%s: want %q, got %q", tt.name, tt.want, tt.ft)
		}
	}
}

// ---------------------------------------------------------------------------
// SolidFill
// ---------------------------------------------------------------------------

func TestNewSolidFill(t *testing.T) {
	red := color.RGBA{R: 255, A: 255}
	f := NewSolidFill(red)

	if f.Color != red {
		t.Errorf("Color: want %v, got %v", red, f.Color)
	}
	if f.FillType() != FillTypeSolid {
		t.Errorf("FillType: want Solid, got %v", f.FillType())
	}
}

func TestSolidFill_IsTransparent(t *testing.T) {
	opaque := NewSolidFill(color.RGBA{R: 255, A: 255})
	if opaque.IsTransparent() {
		t.Error("opaque SolidFill should not be transparent")
	}

	transparent := NewSolidFill(color.RGBA{R: 255, A: 0})
	if !transparent.IsTransparent() {
		t.Error("zero-alpha SolidFill should be transparent")
	}
}

func TestSolidFill_Clone(t *testing.T) {
	orig := NewSolidFill(color.RGBA{R: 0, G: 128, B: 255, A: 255})
	clone := orig.Clone()

	sf, ok := clone.(*SolidFill)
	if !ok {
		t.Fatalf("Clone did not return *SolidFill")
	}
	if sf == orig {
		t.Error("Clone returned the same pointer")
	}
	if sf.Color != orig.Color {
		t.Errorf("Clone color differs: %v vs %v", sf.Color, orig.Color)
	}
	if sf.FillType() != FillTypeSolid {
		t.Errorf("Clone FillType: want Solid, got %v", sf.FillType())
	}

	// Mutation independence.
	sf.Color = color.RGBA{}
	if orig.Color == (color.RGBA{}) {
		t.Error("mutating clone affected original")
	}
}

// ---------------------------------------------------------------------------
// LinearGradientFill
// ---------------------------------------------------------------------------

func TestNewLinearGradientFill_Defaults(t *testing.T) {
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	f := NewLinearGradientFill(black, white)

	if f.StartColor != black {
		t.Errorf("StartColor: want %v, got %v", black, f.StartColor)
	}
	if f.EndColor != white {
		t.Errorf("EndColor: want %v, got %v", white, f.EndColor)
	}
	if f.Angle != 0 {
		t.Errorf("Angle: want 0, got %d", f.Angle)
	}
	if f.Focus != 0 {
		t.Errorf("Focus: want 0, got %v", f.Focus)
	}
	if f.Contrast != 1 {
		t.Errorf("Contrast: want 1, got %v", f.Contrast)
	}
	if f.FillType() != FillTypeLinear {
		t.Errorf("FillType: want Linear, got %v", f.FillType())
	}
}

func TestLinearGradientFill_IsTransparent(t *testing.T) {
	both := &LinearGradientFill{
		StartColor: color.RGBA{A: 0},
		EndColor:   color.RGBA{A: 0},
	}
	if !both.IsTransparent() {
		t.Error("both alpha=0 should be transparent")
	}

	oneOpaque := &LinearGradientFill{
		StartColor: color.RGBA{A: 255},
		EndColor:   color.RGBA{A: 0},
	}
	if oneOpaque.IsTransparent() {
		t.Error("one opaque colour should not be transparent")
	}
}

func TestLinearGradientFill_Clone(t *testing.T) {
	orig := &LinearGradientFill{
		StartColor: color.RGBA{R: 10, A: 255},
		EndColor:   color.RGBA{B: 200, A: 255},
		Angle:      45,
		Focus:      0.5,
		Contrast:   0.8,
	}
	clone := orig.Clone()

	lf, ok := clone.(*LinearGradientFill)
	if !ok {
		t.Fatalf("Clone did not return *LinearGradientFill")
	}
	if lf == orig {
		t.Error("Clone returned the same pointer")
	}
	if *lf != *orig {
		t.Errorf("Clone content differs: %+v vs %+v", *lf, *orig)
	}

	lf.Angle = 180
	if orig.Angle == 180 {
		t.Error("mutating clone affected original")
	}
}

// ---------------------------------------------------------------------------
// GlassFill
// ---------------------------------------------------------------------------

func TestNewGlassFill_Defaults(t *testing.T) {
	c := color.RGBA{R: 100, G: 149, B: 237, A: 255}
	f := NewGlassFill(c)

	if f.Color != c {
		t.Errorf("Color: want %v, got %v", c, f.Color)
	}
	if f.Blend != 0.2 {
		t.Errorf("Blend: want 0.2, got %v", f.Blend)
	}
	if !f.Hatch {
		t.Error("Hatch should default to true")
	}
	if f.FillType() != FillTypeGlass {
		t.Errorf("FillType: want Glass, got %v", f.FillType())
	}
}

func TestGlassFill_IsTransparent(t *testing.T) {
	opaque := NewGlassFill(color.RGBA{R: 255, A: 255})
	if opaque.IsTransparent() {
		t.Error("opaque GlassFill should not be transparent")
	}

	transparent := NewGlassFill(color.RGBA{A: 0})
	if !transparent.IsTransparent() {
		t.Error("zero-alpha GlassFill should be transparent")
	}
}

func TestGlassFill_Clone(t *testing.T) {
	orig := &GlassFill{
		Color: color.RGBA{R: 200, G: 200, B: 200, A: 255},
		Blend: 0.5,
		Hatch: false,
	}
	clone := orig.Clone()

	gf, ok := clone.(*GlassFill)
	if !ok {
		t.Fatalf("Clone did not return *GlassFill")
	}
	if gf == orig {
		t.Error("Clone returned the same pointer")
	}
	if *gf != *orig {
		t.Errorf("Clone content differs: %+v vs %+v", *gf, *orig)
	}

	gf.Blend = 0.9
	if orig.Blend == 0.9 {
		t.Error("mutating clone affected original")
	}
}

// ---------------------------------------------------------------------------
// HatchStyle constants
// ---------------------------------------------------------------------------

func TestHatchStyle_Values(t *testing.T) {
	if HatchHorizontal != 0 {
		t.Errorf("HatchHorizontal want 0, got %d", HatchHorizontal)
	}
	if HatchVertical != 1 {
		t.Errorf("HatchVertical want 1, got %d", HatchVertical)
	}
	if HatchDiagonal1 != 2 {
		t.Errorf("HatchDiagonal1 want 2, got %d", HatchDiagonal1)
	}
	if HatchDiagonal2 != 3 {
		t.Errorf("HatchDiagonal2 want 3, got %d", HatchDiagonal2)
	}
	if HatchCross != 4 {
		t.Errorf("HatchCross want 4, got %d", HatchCross)
	}
	if HatchDiagonalCross != 5 {
		t.Errorf("HatchDiagonalCross want 5, got %d", HatchDiagonalCross)
	}
}

// ---------------------------------------------------------------------------
// HatchFill
// ---------------------------------------------------------------------------

func TestNewHatchFill(t *testing.T) {
	fore := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	back := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	f := NewHatchFill(fore, back, HatchCross)

	if f.ForeColor != fore {
		t.Errorf("ForeColor: want %v, got %v", fore, f.ForeColor)
	}
	if f.BackColor != back {
		t.Errorf("BackColor: want %v, got %v", back, f.BackColor)
	}
	if f.Style != HatchCross {
		t.Errorf("Style: want HatchCross, got %v", f.Style)
	}
	if f.FillType() != FillTypeHatch {
		t.Errorf("FillType: want Hatch, got %v", f.FillType())
	}
}

func TestHatchFill_IsTransparent(t *testing.T) {
	bothTransparent := &HatchFill{
		ForeColor: color.RGBA{A: 0},
		BackColor: color.RGBA{A: 0},
	}
	if !bothTransparent.IsTransparent() {
		t.Error("both alpha=0 should be transparent")
	}

	oneOpaque := &HatchFill{
		ForeColor: color.RGBA{A: 255},
		BackColor: color.RGBA{A: 0},
	}
	if oneOpaque.IsTransparent() {
		t.Error("one opaque colour should not be transparent")
	}
}

func TestHatchFill_Clone(t *testing.T) {
	fore := color.RGBA{R: 50, A: 255}
	back := color.RGBA{B: 200, A: 255}
	orig := NewHatchFill(fore, back, HatchDiagonal1)
	clone := orig.Clone()

	hf, ok := clone.(*HatchFill)
	if !ok {
		t.Fatalf("Clone did not return *HatchFill")
	}
	if hf == orig {
		t.Error("Clone returned the same pointer")
	}
	if *hf != *orig {
		t.Errorf("Clone content differs: %+v vs %+v", *hf, *orig)
	}

	hf.Style = HatchVertical
	if orig.Style == HatchVertical {
		t.Error("mutating clone affected original")
	}
}

// ---------------------------------------------------------------------------
// NoneFill
// ---------------------------------------------------------------------------

func TestNoneFill(t *testing.T) {
	f := &NoneFill{}

	if f.FillType() != FillTypeNone {
		t.Errorf("FillType: want None, got %v", f.FillType())
	}
	if !f.IsTransparent() {
		t.Error("NoneFill should always be transparent")
	}

	clone := f.Clone()
	if _, ok := clone.(*NoneFill); !ok {
		t.Error("Clone did not return *NoneFill")
	}
	if clone == Fill(f) {
		t.Error("Clone returned the same pointer")
	}
}

// ---------------------------------------------------------------------------
// PathGradientStyle constants
// ---------------------------------------------------------------------------

func TestPathGradientStyle_Values(t *testing.T) {
	if PathGradientElliptic != 0 {
		t.Errorf("PathGradientElliptic want 0, got %d", PathGradientElliptic)
	}
	if PathGradientRectangular != 1 {
		t.Errorf("PathGradientRectangular want 1, got %d", PathGradientRectangular)
	}
}

// ---------------------------------------------------------------------------
// PathGradientFill
// ---------------------------------------------------------------------------

func TestNewPathGradientFill(t *testing.T) {
	center := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	edge := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	f := NewPathGradientFill(center, edge, PathGradientElliptic)

	if f.CenterColor != center {
		t.Errorf("CenterColor: want %v, got %v", center, f.CenterColor)
	}
	if f.EdgeColor != edge {
		t.Errorf("EdgeColor: want %v, got %v", edge, f.EdgeColor)
	}
	if f.Style != PathGradientElliptic {
		t.Errorf("Style: want Elliptic, got %v", f.Style)
	}
	if f.FillType() != FillTypePathGradient {
		t.Errorf("FillType: want PathGradient, got %v", f.FillType())
	}
}

func TestPathGradientFill_IsTransparent(t *testing.T) {
	both := &PathGradientFill{
		CenterColor: color.RGBA{A: 0},
		EdgeColor:   color.RGBA{A: 0},
	}
	if !both.IsTransparent() {
		t.Error("both alpha=0 should be transparent")
	}

	oneOpaque := &PathGradientFill{
		CenterColor: color.RGBA{A: 255},
		EdgeColor:   color.RGBA{A: 0},
	}
	if oneOpaque.IsTransparent() {
		t.Error("one opaque colour should not be transparent")
	}
}

func TestPathGradientFill_Clone(t *testing.T) {
	orig := NewPathGradientFill(
		color.RGBA{R: 50, A: 255},
		color.RGBA{B: 200, A: 255},
		PathGradientRectangular,
	)
	clone := orig.Clone()

	pg, ok := clone.(*PathGradientFill)
	if !ok {
		t.Fatalf("Clone did not return *PathGradientFill")
	}
	if pg == orig {
		t.Error("Clone returned the same pointer")
	}
	if *pg != *orig {
		t.Errorf("Clone content differs: %+v vs %+v", *pg, *orig)
	}
	pg.Style = PathGradientElliptic
	if orig.Style == PathGradientElliptic {
		t.Error("mutating clone affected original")
	}
}

// ---------------------------------------------------------------------------
// WrapMode constants
// ---------------------------------------------------------------------------

func TestWrapMode_Values(t *testing.T) {
	if WrapModeTile != 0 {
		t.Errorf("WrapModeTile want 0, got %d", WrapModeTile)
	}
	if WrapModeTileFlipX != 1 {
		t.Errorf("WrapModeTileFlipX want 1, got %d", WrapModeTileFlipX)
	}
	if WrapModeTileFlipY != 2 {
		t.Errorf("WrapModeTileFlipY want 2, got %d", WrapModeTileFlipY)
	}
	if WrapModeTileFlipXY != 3 {
		t.Errorf("WrapModeTileFlipXY want 3, got %d", WrapModeTileFlipXY)
	}
	if WrapModeClamp != 4 {
		t.Errorf("WrapModeClamp want 4, got %d", WrapModeClamp)
	}
}

// ---------------------------------------------------------------------------
// TextureFill
// ---------------------------------------------------------------------------

func TestTextureFill_FillType(t *testing.T) {
	f := &TextureFill{}
	if f.FillType() != FillTypeTexture {
		t.Errorf("FillType: want Texture, got %v", f.FillType())
	}
}

func TestTextureFill_IsTransparent(t *testing.T) {
	// TextureFill.IsTransparent always returns false (matches C# implementation).
	f := &TextureFill{}
	if f.IsTransparent() {
		t.Error("TextureFill.IsTransparent should always be false")
	}
}

func TestTextureFill_Clone(t *testing.T) {
	orig := &TextureFill{
		ImageData:           []byte{1, 2, 3, 4},
		ImageWidth:          100,
		ImageHeight:         80,
		PreserveAspectRatio: true,
		WrapMode:            WrapModeClamp,
		ImageOffsetX:        5,
		ImageOffsetY:        10,
	}
	clone := orig.Clone()

	tf, ok := clone.(*TextureFill)
	if !ok {
		t.Fatalf("Clone did not return *TextureFill")
	}
	if tf == orig {
		t.Error("Clone returned the same pointer")
	}
	// Fields should match.
	if tf.ImageWidth != orig.ImageWidth {
		t.Errorf("ImageWidth: want %d, got %d", orig.ImageWidth, tf.ImageWidth)
	}
	if tf.WrapMode != orig.WrapMode {
		t.Errorf("WrapMode: want %v, got %v", orig.WrapMode, tf.WrapMode)
	}
	// ImageData should be a deep copy.
	if len(tf.ImageData) != len(orig.ImageData) {
		t.Errorf("ImageData length: want %d, got %d", len(orig.ImageData), len(tf.ImageData))
	}
	if &tf.ImageData[0] == &orig.ImageData[0] {
		t.Error("ImageData slice should be a deep copy, not a reference")
	}
	// Mutating clone should not affect original.
	tf.ImageData[0] = 99
	if orig.ImageData[0] == 99 {
		t.Error("mutating clone ImageData affected original")
	}
}

func TestTextureFill_Clone_NilImageData(t *testing.T) {
	orig := &TextureFill{ImageData: nil}
	clone := orig.Clone()
	tf, ok := clone.(*TextureFill)
	if !ok {
		t.Fatalf("Clone did not return *TextureFill")
	}
	if tf.ImageData != nil {
		t.Error("Clone of nil ImageData should be nil")
	}
}

// ---------------------------------------------------------------------------
// Fill interface satisfaction (extended with new types)
// ---------------------------------------------------------------------------

func TestFill_InterfaceSatisfaction(t *testing.T) {
	// Verify all concrete types satisfy the Fill interface at compile time via
	// runtime assertions.
	fills := []Fill{
		NewSolidFill(color.RGBA{}),
		NewLinearGradientFill(color.RGBA{}, color.RGBA{}),
		NewGlassFill(color.RGBA{}),
		NewHatchFill(color.RGBA{}, color.RGBA{}, HatchHorizontal),
		NewPathGradientFill(color.RGBA{}, color.RGBA{}, PathGradientElliptic),
		&TextureFill{},
		&NoneFill{},
	}
	for _, f := range fills {
		if f == nil {
			t.Error("unexpected nil Fill")
		}
		_ = f.FillType()
		_ = f.Clone()
		_ = f.IsTransparent()
	}
}

// ---------------------------------------------------------------------------
// FillType constants (extended)
// ---------------------------------------------------------------------------

func TestFillType_PathGradientAndTexture(t *testing.T) {
	if FillTypePathGradient != "PathGradient" {
		t.Errorf("FillTypePathGradient want PathGradient, got %q", FillTypePathGradient)
	}
	if FillTypeTexture != "Texture" {
		t.Errorf("FillTypeTexture want Texture, got %q", FillTypeTexture)
	}
}
