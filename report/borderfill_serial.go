package report

import (
	"encoding/base64"
	"image/color"
	"strings"

	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
)

// ── BorderLines ──────────────────────────────────────────────────────────────

// parseBorderLines converts a comma-separated FRX BorderLines string to
// style.BorderLines flags. E.g. "Left, Top, Bottom" → 13.
func parseBorderLines(s string) style.BorderLines {
	if s == "" {
		return style.BorderLinesNone
	}
	s = strings.TrimSpace(s)
	if strings.EqualFold(s, "None") {
		return style.BorderLinesNone
	}
	if strings.EqualFold(s, "All") {
		return style.BorderLinesAll
	}
	var result style.BorderLines
	for _, p := range strings.Split(s, ",") {
		switch strings.TrimSpace(p) {
		case "Left":
			result |= style.BorderLinesLeft
		case "Right":
			result |= style.BorderLinesRight
		case "Top":
			result |= style.BorderLinesTop
		case "Bottom":
			result |= style.BorderLinesBottom
		}
	}
	return result
}

// formatBorderLines converts style.BorderLines to a FRX comma-separated string.
func formatBorderLines(bl style.BorderLines) string {
	if bl == style.BorderLinesNone {
		return "None"
	}
	if bl == style.BorderLinesAll {
		return "All"
	}
	var parts []string
	if bl&style.BorderLinesLeft != 0 {
		parts = append(parts, "Left")
	}
	if bl&style.BorderLinesRight != 0 {
		parts = append(parts, "Right")
	}
	if bl&style.BorderLinesTop != 0 {
		parts = append(parts, "Top")
	}
	if bl&style.BorderLinesBottom != 0 {
		parts = append(parts, "Bottom")
	}
	return strings.Join(parts, ", ")
}

// ── LineStyle ─────────────────────────────────────────────────────────────────

func parseLineStyle(s string) style.LineStyle {
	switch s {
	case "Dash":
		return style.LineStyleDash
	case "Dot":
		return style.LineStyleDot
	case "DashDot":
		return style.LineStyleDashDot
	case "DashDotDot":
		return style.LineStyleDashDotDot
	case "Double":
		return style.LineStyleDouble
	case "Custom":
		// LineStyle.Custom (value 6 in C#) — rendered as custom dash by GDI+;
		// treat as Solid for serialization since no FRX DashPattern attribute exists.
		return style.LineStyleSolid
	default:
		return style.LineStyleSolid
	}
}

func formatLineStyle(ls style.LineStyle) string {
	switch ls {
	case style.LineStyleDash:
		return "Dash"
	case style.LineStyleDot:
		return "Dot"
	case style.LineStyleDashDot:
		return "DashDot"
	case style.LineStyleDashDotDot:
		return "DashDotDot"
	case style.LineStyleDouble:
		return "Double"
	default:
		return "Solid"
	}
}

// ── PathGradientStyle ────────────────────────────────────────────────────────

// formatPathGradientStyle converts a PathGradientStyle to its FRX string.
func formatPathGradientStyle(s style.PathGradientStyle) string {
	switch s {
	case style.PathGradientRectangular:
		return "Rectangular"
	default:
		return "Elliptic"
	}
}

// parsePathGradientStyle converts an FRX Fill.Style string to PathGradientStyle.
func parsePathGradientStyle(s string) style.PathGradientStyle {
	switch strings.TrimSpace(s) {
	case "Rectangular":
		return style.PathGradientRectangular
	default:
		return style.PathGradientElliptic
	}
}

// ── WrapMode ─────────────────────────────────────────────────────────────────

// formatWrapMode converts a WrapMode to its C# WrapMode enum name.
func formatWrapMode(w style.WrapMode) string {
	switch w {
	case style.WrapModeTileFlipX:
		return "TileFlipX"
	case style.WrapModeTileFlipY:
		return "TileFlipY"
	case style.WrapModeTileFlipXY:
		return "TileFlipXY"
	case style.WrapModeClamp:
		return "Clamp"
	default:
		return "Tile"
	}
}

// parseWrapMode converts a C# WrapMode enum name to WrapMode.
func parseWrapMode(s string) style.WrapMode {
	switch strings.TrimSpace(s) {
	case "TileFlipX":
		return style.WrapModeTileFlipX
	case "TileFlipY":
		return style.WrapModeTileFlipY
	case "TileFlipXY":
		return style.WrapModeTileFlipXY
	case "Clamp":
		return style.WrapModeClamp
	default:
		return style.WrapModeTile
	}
}

// ── HatchStyle ───────────────────────────────────────────────────────────────

// formatHatchStyle converts a HatchStyle to the C# System.Drawing.Drawing2D.HatchStyle
// enum name used by FRWriter.WriteValue.
func formatHatchStyle(h style.HatchStyle) string {
	switch h {
	case style.HatchHorizontal:
		return "Horizontal"
	case style.HatchVertical:
		return "Vertical"
	case style.HatchDiagonal1:
		return "ForwardDiagonal"
	case style.HatchDiagonal2:
		return "BackwardDiagonal"
	case style.HatchCross:
		return "Cross"
	case style.HatchDiagonalCross:
		return "DiagonalCross"
	default:
		return "Horizontal"
	}
}

// parseHatchStyle converts a C# HatchStyle enum name (or int string) to HatchStyle.
func parseHatchStyle(s string) style.HatchStyle {
	switch strings.TrimSpace(s) {
	case "Horizontal", "0":
		return style.HatchHorizontal
	case "Vertical", "1":
		return style.HatchVertical
	case "ForwardDiagonal", "2":
		return style.HatchDiagonal1
	case "BackwardDiagonal", "3":
		return style.HatchDiagonal2
	case "Cross", "4":
		return style.HatchCross
	case "DiagonalCross", "5":
		return style.HatchDiagonalCross
	default:
		return style.HatchHorizontal
	}
}

// ── Border ────────────────────────────────────────────────────────────────────

// defaultLineColor is the FRX default border line colour (opaque black).
var defaultLineColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}

// defaultShadowColor is the FRX default shadow colour (opaque black).
var defaultShadowColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}

// serializeBorder writes Border properties that differ from their FRX defaults.
// Attributes are written with the "Border." prefix (e.g. "Border.Lines").
// This matches the FastReport C# Border.Serialize(writer, "Border", c.Border)
// delta-serialization pattern.
func serializeBorder(w Writer, b *style.Border) {
	if b == nil {
		return
	}

	// Shadow.
	if b.Shadow {
		w.WriteBool("Border.Shadow", true)
	}
	if b.ShadowWidth != 4 {
		w.WriteFloat("Border.ShadowWidth", b.ShadowWidth)
	}
	if b.ShadowColor != defaultShadowColor {
		w.WriteStr("Border.ShadowColor", utils.FormatColor(b.ShadowColor))
	}

	// Visible lines bitmask (skip when none).
	if b.VisibleLines != style.BorderLinesNone {
		w.WriteStr("Border.Lines", formatBorderLines(b.VisibleLines))
	}

	// Line-level properties — only if lines are initialized.
	if b.Lines[0] == nil {
		return
	}

	// Determine whether all 4 lines share the same properties.
	allEqual := true
	for i := 1; i < 4; i++ {
		if b.Lines[i] == nil || !b.Lines[i].Equals(b.Lines[0]) {
			allEqual = false
			break
		}
	}

	if allEqual {
		l := b.Lines[0]
		if l.Color != defaultLineColor {
			w.WriteStr("Border.Color", utils.FormatColor(l.Color))
		}
		if l.Style != style.LineStyleSolid {
			w.WriteStr("Border.Style", formatLineStyle(l.Style))
		}
		if l.Width != 1 {
			w.WriteFloat("Border.Width", l.Width)
		}
	} else {
		// Write per-line overrides.
		lineNames := [4]string{"LeftLine", "TopLine", "RightLine", "BottomLine"}
		for i, n := range lineNames {
			l := b.Lines[i]
			if l == nil {
				continue
			}
			pfx := "Border." + n
			if l.Color != defaultLineColor {
				w.WriteStr(pfx+".Color", utils.FormatColor(l.Color))
			}
			if l.Style != style.LineStyleSolid {
				w.WriteStr(pfx+".Style", formatLineStyle(l.Style))
			}
			if l.Width != 1 {
				w.WriteFloat(pfx+".Width", l.Width)
			}
		}
	}
}

// deserializeBorder reads Border properties from r into b.
// b must be non-nil; uninitialized Lines pointers are created on demand.
func deserializeBorder(r Reader, b *style.Border) {
	// Ensure all line slots are initialized.
	for i := range b.Lines {
		if b.Lines[i] == nil {
			b.Lines[i] = style.NewBorderLine()
		}
	}

	// Shadow.
	if r.ReadBool("Border.Shadow", false) {
		b.Shadow = true
	}
	b.ShadowWidth = r.ReadFloat("Border.ShadowWidth", 4)
	if s := r.ReadStr("Border.ShadowColor", ""); s != "" {
		if c, err := utils.ParseColor(s); err == nil {
			b.ShadowColor = c
		}
	}

	// Visible lines.
	if s := r.ReadStr("Border.Lines", ""); s != "" {
		b.VisibleLines = parseBorderLines(s)
	}

	// Common line properties (applied to all 4 lines).
	if s := r.ReadStr("Border.Color", ""); s != "" {
		if c, err := utils.ParseColor(s); err == nil {
			for i := range b.Lines {
				b.Lines[i].Color = c
			}
		}
	}
	if s := r.ReadStr("Border.Style", ""); s != "" {
		ls := parseLineStyle(s)
		for i := range b.Lines {
			b.Lines[i].Style = ls
		}
	}
	if bw := r.ReadFloat("Border.Width", 0); bw > 0 {
		for i := range b.Lines {
			b.Lines[i].Width = bw
		}
	}

	// Per-line overrides.
	type lineSpec struct {
		idx int
		pfx string
	}
	lineSpecs := []lineSpec{
		{0, "Border.LeftLine"},
		{1, "Border.TopLine"},
		{2, "Border.RightLine"},
		{3, "Border.BottomLine"},
	}
	for _, ls := range lineSpecs {
		l := b.Lines[ls.idx]
		if s := r.ReadStr(ls.pfx+".Color", ""); s != "" {
			if c, err := utils.ParseColor(s); err == nil {
				l.Color = c
			}
		}
		if s := r.ReadStr(ls.pfx+".Style", ""); s != "" {
			l.Style = parseLineStyle(s)
		}
		if w := r.ReadFloat(ls.pfx+".Width", 0); w > 0 {
			l.Width = w
		}
	}
}

// ── Fill ──────────────────────────────────────────────────────────────────────

// serializeFill writes Fill properties that differ from the FRX defaults.
// The default fill is SolidFill with transparent colour; only non-transparent
// solid fills and all non-solid fills are emitted.
func serializeFill(w Writer, f style.Fill) {
	if f == nil {
		return
	}
	transparent := color.RGBA{}

	switch ft := f.(type) {
	case *style.SolidFill:
		// No "Fill" type attribute needed (Solid is the implicit default).
		// Only write color when non-transparent.
		if ft.Color != transparent {
			w.WriteStr("Fill.Color", utils.FormatColor(ft.Color))
		}

	case *style.LinearGradientFill:
		w.WriteStr("Fill", "LinearGradient")
		if ft.StartColor != transparent {
			w.WriteStr("Fill.StartColor", utils.FormatColor(ft.StartColor))
		}
		if ft.EndColor != transparent {
			w.WriteStr("Fill.EndColor", utils.FormatColor(ft.EndColor))
		}
		if ft.Angle != 0 {
			w.WriteInt("Fill.Angle", ft.Angle)
		}
		if ft.Focus != 0 {
			w.WriteFloat("Fill.Focus", ft.Focus)
		}
		// Default Contrast in C# is 100 (mapped to 1.0 in Go).
		if ft.Contrast != 1 {
			w.WriteFloat("Fill.Contrast", ft.Contrast)
		}

	case *style.GlassFill:
		w.WriteStr("Fill", "Glass")
		if ft.Color != transparent {
			w.WriteStr("Fill.Color", utils.FormatColor(ft.Color))
		}
		// Default Blend is 0.2 (C# uses 20, but we map to 0–1).
		if ft.Blend != 0.2 {
			w.WriteFloat("Fill.Blend", ft.Blend)
		}
		// Default Hatch is true.
		if !ft.Hatch {
			w.WriteBool("Fill.Hatch", false)
		}

	case *style.HatchFill:
		w.WriteStr("Fill", "Hatch")
		if ft.ForeColor != transparent {
			w.WriteStr("Fill.ForeColor", utils.FormatColor(ft.ForeColor))
		}
		if ft.BackColor != transparent {
			w.WriteStr("Fill.BackColor", utils.FormatColor(ft.BackColor))
		}
		if ft.Style != 0 {
			w.WriteStr("Fill.Style", formatHatchStyle(ft.Style))
		}

	case *style.PathGradientFill:
		w.WriteStr("Fill", "PathGradient")
		if ft.CenterColor != transparent {
			w.WriteStr("Fill.CenterColor", utils.FormatColor(ft.CenterColor))
		}
		if ft.EdgeColor != transparent {
			w.WriteStr("Fill.EdgeColor", utils.FormatColor(ft.EdgeColor))
		}
		// Default style is Elliptic — only write when Rectangular.
		if ft.Style != style.PathGradientElliptic {
			w.WriteStr("Fill.Style", formatPathGradientStyle(ft.Style))
		}

	case *style.TextureFill:
		w.WriteStr("Fill", "Texture")
		if ft.ImageWidth != 0 {
			w.WriteInt("Fill.ImageWidth", ft.ImageWidth)
		}
		if ft.ImageHeight != 0 {
			w.WriteInt("Fill.ImageHeight", ft.ImageHeight)
		}
		if ft.PreserveAspectRatio {
			w.WriteBool("Fill.PreserveAspectRatio", true)
		}
		if ft.WrapMode != style.WrapModeTile {
			w.WriteStr("Fill.WrapMode", formatWrapMode(ft.WrapMode))
		}
		if ft.ImageOffsetX != 0 {
			w.WriteInt("Fill.ImageOffsetX", ft.ImageOffsetX)
		}
		if ft.ImageOffsetY != 0 {
			w.WriteInt("Fill.ImageOffsetY", ft.ImageOffsetY)
		}
		if len(ft.ImageData) > 0 {
			w.WriteStr("Fill.ImageData", base64.StdEncoding.EncodeToString(ft.ImageData))
		}

	// NoneFill and unknown types: no output.
	}
}

// deserializeFill reads Fill properties from r and returns the appropriate Fill
// implementation. current is used as a base for SolidFill (to preserve any
// existing colour when the type attribute is absent).
func deserializeFill(r Reader, current style.Fill) style.Fill {
	fillType := r.ReadStr("Fill", "")

	switch fillType {
	case "LinearGradient":
		f := &style.LinearGradientFill{Contrast: 1}
		if s := r.ReadStr("Fill.StartColor", ""); s != "" {
			if c, err := utils.ParseColor(s); err == nil {
				f.StartColor = c
			}
		}
		if s := r.ReadStr("Fill.EndColor", ""); s != "" {
			if c, err := utils.ParseColor(s); err == nil {
				f.EndColor = c
			}
		}
		f.Angle = r.ReadInt("Fill.Angle", 0)
		f.Focus = r.ReadFloat("Fill.Focus", 0)
		f.Contrast = r.ReadFloat("Fill.Contrast", 1)
		return f

	case "Glass":
		f := &style.GlassFill{Blend: 0.2, Hatch: true}
		if s := r.ReadStr("Fill.Color", ""); s != "" {
			if c, err := utils.ParseColor(s); err == nil {
				f.Color = c
			}
		}
		f.Blend = r.ReadFloat("Fill.Blend", 0.2)
		f.Hatch = r.ReadBool("Fill.Hatch", true)
		return f

	case "Hatch":
		f := &style.HatchFill{}
		if s := r.ReadStr("Fill.ForeColor", ""); s != "" {
			if c, err := utils.ParseColor(s); err == nil {
				f.ForeColor = c
			}
		}
		if s := r.ReadStr("Fill.BackColor", ""); s != "" {
			if c, err := utils.ParseColor(s); err == nil {
				f.BackColor = c
			}
		}
		f.Style = parseHatchStyle(r.ReadStr("Fill.Style", "Horizontal"))
		return f

	case "PathGradient":
		// Mirrors FastReport.PathGradientFill — CenterColor, EdgeColor, Style.
		f := &style.PathGradientFill{}
		if s := r.ReadStr("Fill.CenterColor", ""); s != "" {
			if c, err := utils.ParseColor(s); err == nil {
				f.CenterColor = c
			}
		}
		if s := r.ReadStr("Fill.EdgeColor", ""); s != "" {
			if c, err := utils.ParseColor(s); err == nil {
				f.EdgeColor = c
			}
		}
		f.Style = parsePathGradientStyle(r.ReadStr("Fill.Style", "Elliptic"))
		return f

	case "Texture":
		// Mirrors FastReport.TextureFill — inline ImageData path only.
		// BlobStore image-index mechanism is not implemented in the Go port.
		f := &style.TextureFill{
			WrapMode: style.WrapModeTile,
		}
		f.ImageWidth = r.ReadInt("Fill.ImageWidth", 0)
		f.ImageHeight = r.ReadInt("Fill.ImageHeight", 0)
		f.PreserveAspectRatio = r.ReadBool("Fill.PreserveAspectRatio", false)
		f.WrapMode = parseWrapMode(r.ReadStr("Fill.WrapMode", "Tile"))
		f.ImageOffsetX = r.ReadInt("Fill.ImageOffsetX", 0)
		f.ImageOffsetY = r.ReadInt("Fill.ImageOffsetY", 0)
		if s := r.ReadStr("Fill.ImageData", ""); s != "" {
			if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
				f.ImageData = decoded
			}
		}
		return f

	default:
		// "Solid" or "" — keep existing SolidFill and update colour if present.
		sf, ok := current.(*style.SolidFill)
		if !ok {
			sf = &style.SolidFill{}
		}
		if s := r.ReadStr("Fill.Color", ""); s != "" {
			if c, err := utils.ParseColor(s); err == nil {
				sf = &style.SolidFill{Color: c}
			}
		}
		return sf
	}
}
