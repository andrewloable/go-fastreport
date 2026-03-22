package style

import "image/color"

// FillType identifies which concrete fill implementation is in use.
type FillType string

const (
	// FillTypeSolid is a plain single-colour fill.
	FillTypeSolid FillType = "Solid"
	// FillTypeLinear is a linear gradient fill.
	FillTypeLinear FillType = "Linear"
	// FillTypeGlass is a glass-effect fill.
	FillTypeGlass FillType = "Glass"
	// FillTypeHatch is a hatch-pattern fill.
	FillTypeHatch FillType = "Hatch"
	// FillTypePathGradient is a radial/path gradient fill.
	FillTypePathGradient FillType = "PathGradient"
	// FillTypeTexture is an image-tiled texture fill.
	FillTypeTexture FillType = "Texture"
	// FillTypeNone represents no fill (fully transparent).
	FillTypeNone FillType = "None"
)

// Fill is the base interface for all fill types.
// It is the Go equivalent of FastReport.FillBase.
type Fill interface {
	// FillType returns the concrete fill variant.
	FillType() FillType
	// Clone returns a deep copy of the fill.
	Clone() Fill
	// IsTransparent reports whether the fill produces no visible colour.
	IsTransparent() bool
}

// ---------------------------------------------------------------------------
// SolidFill
// ---------------------------------------------------------------------------

// SolidFill fills a region with a single, uniform colour.
// It is the Go equivalent of FastReport.SolidFill.
type SolidFill struct {
	// Color is the fill colour.
	Color color.RGBA
}

// NewSolidFill returns a SolidFill with the given colour.
func NewSolidFill(c color.RGBA) *SolidFill {
	return &SolidFill{Color: c}
}

// FillType implements Fill.
func (f *SolidFill) FillType() FillType { return FillTypeSolid }

// Clone implements Fill.
func (f *SolidFill) Clone() Fill { return &SolidFill{Color: f.Color} }

// IsTransparent implements Fill; a SolidFill is transparent when alpha == 0.
func (f *SolidFill) IsTransparent() bool { return f.Color.A == 0 }

// ---------------------------------------------------------------------------
// LinearGradientFill
// ---------------------------------------------------------------------------

// LinearGradientFill fills a region with a linear colour gradient.
// It is the Go equivalent of FastReport.LinearGradientFill.
type LinearGradientFill struct {
	// StartColor is the gradient's start colour.
	StartColor color.RGBA
	// EndColor is the gradient's end colour.
	EndColor color.RGBA
	// Angle is the gradient direction in degrees (0 = left to right).
	Angle int
	// Focus is the normalised position of the gradient centre (0..1).
	Focus float32
	// Contrast controls how sharply the gradient transitions (0..1).
	Contrast float32
}

// NewLinearGradientFill returns a LinearGradientFill from start to end colour
// with all other settings at their FastReport defaults (angle 0, focus 0,
// contrast 100 mapped to 1.0 here since we use a 0..1 scale).
func NewLinearGradientFill(start, end color.RGBA) *LinearGradientFill {
	return &LinearGradientFill{
		StartColor: start,
		EndColor:   end,
		Angle:      0,
		Focus:      0,
		Contrast:   1,
	}
}

// FillType implements Fill.
func (f *LinearGradientFill) FillType() FillType { return FillTypeLinear }

// Clone implements Fill.
func (f *LinearGradientFill) Clone() Fill {
	c := *f
	return &c
}

// IsTransparent implements Fill; transparent only when both colours have
// alpha == 0, matching FastReport.LinearGradientFill.IsTransparent.
func (f *LinearGradientFill) IsTransparent() bool {
	return f.StartColor.A == 0 && f.EndColor.A == 0
}

// ---------------------------------------------------------------------------
// GlassFill
// ---------------------------------------------------------------------------

// GlassFill produces a glass-like sheen effect over a base colour.
// It is the Go equivalent of FastReport.GlassFill.
type GlassFill struct {
	// Color is the base fill colour.
	Color color.RGBA
	// Blend controls the white highlight strength (0..1). Default 0.2.
	Blend float32
	// Hatch enables a diagonal hatch overlay. Default true.
	Hatch bool
}

// NewGlassFill returns a GlassFill with FastReport defaults
// (blend 0.2, hatch enabled).
func NewGlassFill(c color.RGBA) *GlassFill {
	return &GlassFill{
		Color: c,
		Blend: 0.2,
		Hatch: true,
	}
}

// FillType implements Fill.
func (f *GlassFill) FillType() FillType { return FillTypeGlass }

// Clone implements Fill.
func (f *GlassFill) Clone() Fill {
	c := *f
	return &c
}

// IsTransparent implements Fill.
func (f *GlassFill) IsTransparent() bool { return f.Color.A == 0 }

// ---------------------------------------------------------------------------
// HatchStyle
// ---------------------------------------------------------------------------

// HatchStyle enumerates the available hatch patterns.
// The values mirror the most common System.Drawing.Drawing2D.HatchStyle entries
// used by FastReport.
type HatchStyle int

const (
	// HatchHorizontal draws horizontal lines.
	HatchHorizontal HatchStyle = iota
	// HatchVertical draws vertical lines.
	HatchVertical
	// HatchDiagonal1 draws forward-diagonal lines (/).
	HatchDiagonal1
	// HatchDiagonal2 draws backward-diagonal lines (\).
	HatchDiagonal2
	// HatchCross draws a cross (horizontal + vertical) pattern.
	HatchCross
	// HatchDiagonalCross draws a diagonal cross (X) pattern.
	HatchDiagonalCross
)

// ---------------------------------------------------------------------------
// HatchFill
// ---------------------------------------------------------------------------

// HatchFill fills a region with a repeating hatch pattern.
// It is the Go equivalent of FastReport.HatchFill.
type HatchFill struct {
	// ForeColor is the foreground (line) colour.
	ForeColor color.RGBA
	// BackColor is the background colour between the lines.
	BackColor color.RGBA
	// Style selects the hatch pattern.
	Style HatchStyle
}

// NewHatchFill returns a HatchFill with the given colours and pattern.
func NewHatchFill(fore, back color.RGBA, style HatchStyle) *HatchFill {
	return &HatchFill{ForeColor: fore, BackColor: back, Style: style}
}

// FillType implements Fill.
func (f *HatchFill) FillType() FillType { return FillTypeHatch }

// Clone implements Fill.
func (f *HatchFill) Clone() Fill {
	c := *f
	return &c
}

// IsTransparent implements Fill; transparent when both colours have alpha == 0.
func (f *HatchFill) IsTransparent() bool {
	return f.ForeColor.A == 0 && f.BackColor.A == 0
}

// ---------------------------------------------------------------------------
// PathGradientStyle
// ---------------------------------------------------------------------------

// PathGradientStyle selects the shape of the path gradient.
// It mirrors FastReport.PathGradientStyle.
type PathGradientStyle int

const (
	// PathGradientElliptic uses an ellipse centred on the bounding box.
	PathGradientElliptic PathGradientStyle = iota
	// PathGradientRectangular uses the bounding rectangle directly.
	PathGradientRectangular
)

// ---------------------------------------------------------------------------
// PathGradientFill
// ---------------------------------------------------------------------------

// PathGradientFill fills a region with a radial (path) gradient that blends
// from a center colour outward to an edge colour.
// It is the Go equivalent of FastReport.PathGradientFill.
// Draw() is handled by exporters — only serialization data is stored here.
type PathGradientFill struct {
	// CenterColor is the colour at the centre of the gradient.
	CenterColor color.RGBA
	// EdgeColor is the colour at the edges of the gradient.
	EdgeColor color.RGBA
	// Style selects the gradient shape (Elliptic or Rectangular).
	Style PathGradientStyle
}

// NewPathGradientFill returns a PathGradientFill with the given colours and
// style, matching FastReport.PathGradientFill(CenterColor, EdgeColor, Style).
func NewPathGradientFill(center, edge color.RGBA, style PathGradientStyle) *PathGradientFill {
	return &PathGradientFill{CenterColor: center, EdgeColor: edge, Style: style}
}

// FillType implements Fill.
func (f *PathGradientFill) FillType() FillType { return FillTypePathGradient }

// Clone implements Fill.
func (f *PathGradientFill) Clone() Fill {
	c := *f
	return &c
}

// IsTransparent implements Fill; transparent only when both colours have alpha == 0,
// matching FastReport.PathGradientFill.IsTransparent.
func (f *PathGradientFill) IsTransparent() bool {
	return f.CenterColor.A == 0 && f.EdgeColor.A == 0
}

// ---------------------------------------------------------------------------
// WrapMode (texture fill tiling)
// ---------------------------------------------------------------------------

// WrapMode controls how a TextureFill tiles its image.
// It mirrors System.Drawing.Drawing2D.WrapMode.
type WrapMode int

const (
	// WrapModeTile tiles the image.
	WrapModeTile WrapMode = iota
	// WrapModeTileFlipX tiles the image, flipping horizontally on alternate columns.
	WrapModeTileFlipX
	// WrapModeTileFlipY tiles the image, flipping vertically on alternate rows.
	WrapModeTileFlipY
	// WrapModeTileFlipXY tiles the image, flipping both axes.
	WrapModeTileFlipXY
	// WrapModeClamp clamps the image without tiling.
	WrapModeClamp
)

// ---------------------------------------------------------------------------
// TextureFill
// ---------------------------------------------------------------------------

// TextureFill fills a region by tiling an embedded image.
// It is the Go equivalent of FastReport.TextureFill.
// Draw() is handled by exporters — only the serialization fields are stored.
// The image data is stored as a raw base-64-decoded byte slice matching the
// FRX Fill.ImageData attribute.
type TextureFill struct {
	// ImageData holds the raw image bytes (PNG/JPEG etc.) as decoded from
	// the FRX Fill.ImageData base-64 attribute.
	ImageData []byte
	// ImageWidth is the display width in pixels (0 = natural width).
	ImageWidth int
	// ImageHeight is the display height in pixels (0 = natural height).
	ImageHeight int
	// PreserveAspectRatio keeps the aspect ratio when one dimension is set.
	PreserveAspectRatio bool
	// WrapMode controls tiling behaviour. Default is WrapModeTile.
	WrapMode WrapMode
	// ImageOffsetX is the horizontal offset of the tile origin in pixels.
	ImageOffsetX int
	// ImageOffsetY is the vertical offset of the tile origin in pixels.
	ImageOffsetY int
	// ImageIndex is the BlobStore index used when the FRX is stored with a
	// BlobStore (designer/preview format). -1 means "not set" (inline path).
	// Mirrors C# TextureFill.ImageIndex (Fills.cs).
	ImageIndex int
}

// FillType implements Fill.
func (f *TextureFill) FillType() FillType { return FillTypeTexture }

// NewTextureFill returns a TextureFill with ImageIndex set to -1 (not set),
// matching the C# TextureFill constructor which calls ResetImageIndex().
func NewTextureFill() *TextureFill {
	return &TextureFill{
		WrapMode:   WrapModeTile,
		ImageIndex: -1,
	}
}

// Clone implements Fill.
func (f *TextureFill) Clone() Fill {
	c := *f
	if f.ImageData != nil {
		c.ImageData = make([]byte, len(f.ImageData))
		copy(c.ImageData, f.ImageData)
	}
	return &c
}

// IsTransparent implements Fill; always false (matches FastReport.TextureFill.IsTransparent).
func (f *TextureFill) IsTransparent() bool { return false }

// ---------------------------------------------------------------------------
// NoneFill
// ---------------------------------------------------------------------------

// NoneFill represents the absence of any fill (fully transparent).
// It is the Go equivalent of using no fill in FastReport.
type NoneFill struct{}

// FillType implements Fill.
func (f *NoneFill) FillType() FillType { return FillTypeNone }

// Clone implements Fill.
func (f *NoneFill) Clone() Fill { return &NoneFill{} }

// IsTransparent implements Fill; always true.
func (f *NoneFill) IsTransparent() bool { return true }
