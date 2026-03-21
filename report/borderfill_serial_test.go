package report

import (
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

// ── parseBorderLines ─────────────────────────────────────────────────────────

func TestParseBorderLines(t *testing.T) {
	tests := []struct {
		input string
		want  style.BorderLines
	}{
		{"", style.BorderLinesNone},
		{"None", style.BorderLinesNone},
		{"All", style.BorderLinesAll},
		{"Left", style.BorderLinesLeft},
		{"Right", style.BorderLinesRight},
		{"Top", style.BorderLinesTop},
		{"Bottom", style.BorderLinesBottom},
		{"Left, Top", style.BorderLinesLeft | style.BorderLinesTop},
		{"Left, Right, Bottom", style.BorderLinesLeft | style.BorderLinesRight | style.BorderLinesBottom},
		{"Top, Bottom", style.BorderLinesTop | style.BorderLinesBottom},
	}
	for _, tc := range tests {
		got := parseBorderLines(tc.input)
		if got != tc.want {
			t.Errorf("parseBorderLines(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestFormatBorderLines(t *testing.T) {
	tests := []struct {
		input style.BorderLines
		want  string
	}{
		{style.BorderLinesNone, "None"},
		{style.BorderLinesAll, "All"},
		{style.BorderLinesLeft, "Left"},
		{style.BorderLinesLeft | style.BorderLinesTop, "Left, Top"},
		{style.BorderLinesRight | style.BorderLinesBottom, "Right, Bottom"},
	}
	for _, tc := range tests {
		got := formatBorderLines(tc.input)
		if got != tc.want {
			t.Errorf("formatBorderLines(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ── parseLineStyle ───────────────────────────────────────────────────────────

func TestParseLineStyle(t *testing.T) {
	tests := []struct {
		input string
		want  style.LineStyle
	}{
		{"Solid", style.LineStyleSolid},
		{"Dash", style.LineStyleDash},
		{"Dot", style.LineStyleDot},
		{"DashDot", style.LineStyleDashDot},
		{"DashDotDot", style.LineStyleDashDotDot},
		{"Double", style.LineStyleDouble},
		{"", style.LineStyleSolid},   // unknown → Solid
		{"unknown", style.LineStyleSolid},
	}
	for _, tc := range tests {
		got := parseLineStyle(tc.input)
		if got != tc.want {
			t.Errorf("parseLineStyle(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// ── serializeBorder / deserializeBorder round-trip ───────────────────────────

func TestSerializeBorder_Defaults(t *testing.T) {
	// Default border (no visible lines, black, 1px) should produce no output.
	b := style.NewBorder()
	w := newTestWriter()
	serializeBorder(w, b)
	for k := range w.data {
		t.Errorf("unexpected key %q serialized for default border", k)
	}
}

func TestSerializeBorder_VisibleLines(t *testing.T) {
	b := style.NewBorder()
	b.VisibleLines = style.BorderLinesAll
	w := newTestWriter()
	serializeBorder(w, b)
	if v, ok := w.data["Border.Lines"]; !ok || v != "All" {
		t.Errorf("Border.Lines = %v, want All", v)
	}
}

func TestSerializeBorder_Shadow(t *testing.T) {
	b := style.NewBorder()
	b.Shadow = true
	b.ShadowWidth = 6
	b.ShadowColor = color.RGBA{R: 64, G: 64, B: 64, A: 255}
	w := newTestWriter()
	serializeBorder(w, b)
	if v, ok := w.data["Border.Shadow"]; !ok || v != true {
		t.Errorf("Border.Shadow = %v, want true", v)
	}
	if v, ok := w.data["Border.ShadowWidth"]; !ok || v != float32(6) {
		t.Errorf("Border.ShadowWidth = %v, want 6", v)
	}
	if _, ok := w.data["Border.ShadowColor"]; !ok {
		t.Error("Border.ShadowColor should be serialized when non-default")
	}
}

func TestSerializeBorder_CommonColor(t *testing.T) {
	b := style.NewBorder()
	b.SetColor(color.RGBA{R: 255, G: 0, B: 0, A: 255}) // red
	w := newTestWriter()
	serializeBorder(w, b)
	if v, ok := w.data["Border.Color"]; !ok {
		t.Error("Border.Color should be serialized when all lines share a non-default color")
	} else if v != "#FFFF0000" {
		t.Errorf("Border.Color = %v, want #FFFF0000", v)
	}
}

func TestDeserializeBorder_RoundTrip(t *testing.T) {
	// Build a border with visible lines and a custom color.
	orig := style.NewBorder()
	orig.VisibleLines = style.BorderLinesLeft | style.BorderLinesBottom
	orig.SetColor(color.RGBA{R: 0, G: 128, B: 255, A: 255})
	orig.SetWidth(2)
	orig.Shadow = true

	// Serialize.
	w := newTestWriter()
	serializeBorder(w, orig)

	// Build reader from serialized data (convert map to string map for testReader).
	strData := make(map[string]any)
	for k, v := range w.data {
		strData[k] = v
	}
	r := newTestReader(strData)

	// Deserialize into a fresh border.
	got := style.NewBorder()
	deserializeBorder(r, got)

	if got.VisibleLines != orig.VisibleLines {
		t.Errorf("VisibleLines = %d, want %d", got.VisibleLines, orig.VisibleLines)
	}
	if !got.Shadow {
		t.Error("Shadow should be true")
	}
	for i, l := range got.Lines {
		if l.Color != orig.Lines[i].Color {
			t.Errorf("Lines[%d].Color = %v, want %v", i, l.Color, orig.Lines[i].Color)
		}
		if l.Width != orig.Lines[i].Width {
			t.Errorf("Lines[%d].Width = %v, want %v", i, l.Width, orig.Lines[i].Width)
		}
	}
}

// ── serializeFill / deserializeFill round-trip ───────────────────────────────

func TestSerializeFill_Transparent(t *testing.T) {
	// Transparent SolidFill → no output (matches FRX default).
	f := &style.SolidFill{Color: color.RGBA{}}
	w := newTestWriter()
	serializeFill(w, f)
	if len(w.data) != 0 {
		t.Errorf("transparent SolidFill should produce no output, got %v", w.data)
	}
}

func TestSerializeFill_White(t *testing.T) {
	f := &style.SolidFill{Color: color.RGBA{R: 255, G: 255, B: 255, A: 255}}
	w := newTestWriter()
	serializeFill(w, f)
	if v, ok := w.data["Fill.Color"]; !ok || v != "#FFFFFFFF" {
		t.Errorf("Fill.Color = %v, want #FFFFFFFF", v)
	}
	if _, ok := w.data["Fill"]; ok {
		t.Error("Fill type attribute should not be written for SolidFill")
	}
}

func TestSerializeFill_LinearGradient(t *testing.T) {
	f := &style.LinearGradientFill{
		StartColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		EndColor:   color.RGBA{R: 0, G: 0, B: 255, A: 255},
		Angle:      90,
		Contrast:   1,
	}
	w := newTestWriter()
	serializeFill(w, f)
	if v, ok := w.data["Fill"]; !ok || v != "LinearGradient" {
		t.Errorf("Fill = %v, want LinearGradient", v)
	}
	if _, ok := w.data["Fill.StartColor"]; !ok {
		t.Error("Fill.StartColor should be serialized")
	}
	if _, ok := w.data["Fill.EndColor"]; !ok {
		t.Error("Fill.EndColor should be serialized")
	}
	if v, ok := w.data["Fill.Angle"]; !ok || v != 90 {
		t.Errorf("Fill.Angle = %v, want 90", v)
	}
}

func TestDeserializeFill_SolidRoundTrip(t *testing.T) {
	orig := &style.SolidFill{Color: color.RGBA{R: 200, G: 100, B: 50, A: 255}}
	w := newTestWriter()
	serializeFill(w, orig)

	r := newTestReader(w.data)
	got := deserializeFill(r, &style.SolidFill{})

	sf, ok := got.(*style.SolidFill)
	if !ok {
		t.Fatalf("got fill type %T, want *style.SolidFill", got)
	}
	if sf.Color != orig.Color {
		t.Errorf("Color = %v, want %v", sf.Color, orig.Color)
	}
}

func TestDeserializeFill_LinearGradientRoundTrip(t *testing.T) {
	orig := &style.LinearGradientFill{
		StartColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		EndColor:   color.RGBA{R: 0, G: 0, B: 255, A: 255},
		Angle:      270,
		Focus:      0.5,
		Contrast:   0.8,
	}
	w := newTestWriter()
	serializeFill(w, orig)

	r := newTestReader(w.data)
	got := deserializeFill(r, &style.SolidFill{})

	lf, ok := got.(*style.LinearGradientFill)
	if !ok {
		t.Fatalf("got fill type %T, want *style.LinearGradientFill", got)
	}
	if lf.StartColor != orig.StartColor {
		t.Errorf("StartColor = %v, want %v", lf.StartColor, orig.StartColor)
	}
	if lf.EndColor != orig.EndColor {
		t.Errorf("EndColor = %v, want %v", lf.EndColor, orig.EndColor)
	}
	if lf.Angle != orig.Angle {
		t.Errorf("Angle = %d, want %d", lf.Angle, orig.Angle)
	}
}

func TestDeserializeFill_NamedColor(t *testing.T) {
	// FRX uses named colors like "WhiteSmoke" for Fill.Color.
	r := newTestReader(map[string]any{
		"Fill.Color": "WhiteSmoke",
	})
	got := deserializeFill(r, &style.SolidFill{})
	sf, ok := got.(*style.SolidFill)
	if !ok {
		t.Fatalf("got fill type %T, want *style.SolidFill", got)
	}
	want := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	if sf.Color != want {
		t.Errorf("Color = %v, want %v (WhiteSmoke)", sf.Color, want)
	}
}

func TestDeserializeBorder_NamedColor(t *testing.T) {
	// FRX uses named colors like "LightGray" for Border.Color.
	r := newTestReader(map[string]any{
		"Border.Lines": "All",
		"Border.Color": "LightGray",
	})
	b := style.NewBorder()
	deserializeBorder(r, b)
	if b.VisibleLines != style.BorderLinesAll {
		t.Errorf("VisibleLines = %d, want All", b.VisibleLines)
	}
	want := color.RGBA{R: 211, G: 211, B: 211, A: 255}
	for i, l := range b.Lines {
		if l.Color != want {
			t.Errorf("Lines[%d].Color = %v, want LightGray %v", i, l.Color, want)
		}
	}
}

// ── formatLineStyle ───────────────────────────────────────────────────────────

func TestFormatLineStyle_NonSolidStyles(t *testing.T) {
	cases := []struct {
		ls   style.LineStyle
		want string
	}{
		{style.LineStyleDash, "Dash"},
		{style.LineStyleDot, "Dot"},
		{style.LineStyleDashDot, "DashDot"},
		{style.LineStyleDashDotDot, "DashDotDot"},
		{style.LineStyleDouble, "Double"},
		{style.LineStyleSolid, "Solid"},
	}
	for _, tc := range cases {
		// Create a border where all lines share the given non-default style.
		b := style.NewBorder()
		b.SetLineStyle(tc.ls)
		w := newTestWriter()
		serializeBorder(w, b)
		if tc.ls == style.LineStyleSolid {
			if _, ok := w.data["Border.Style"]; ok {
				t.Errorf("style Solid should not write Border.Style key")
			}
		} else {
			got, ok := w.data["Border.Style"]
			if !ok {
				t.Errorf("style %v: Border.Style not written", tc.ls)
				continue
			}
			if got != tc.want {
				t.Errorf("style %v: Border.Style = %v, want %q", tc.ls, got, tc.want)
			}
		}
	}
}

// ── serializeBorder nil ───────────────────────────────────────────────────────

func TestSerializeBorder_NilBorder(t *testing.T) {
	w := newTestWriter()
	serializeBorder(w, nil) // should return immediately without panic
	if len(w.data) != 0 {
		t.Errorf("nil border should produce no output, got %v", w.data)
	}
}

// ── serializeBorder with Lines[0] == nil ──────────────────────────────────────

func TestSerializeBorder_NilLines(t *testing.T) {
	b := style.NewBorder()
	b.VisibleLines = style.BorderLinesAll
	// Clear all line pointers so b.Lines[0] == nil.
	for i := range b.Lines {
		b.Lines[i] = nil
	}
	w := newTestWriter()
	serializeBorder(w, b)
	// Border.Lines should be written, but no line-level attributes.
	if _, ok := w.data["Border.Lines"]; !ok {
		t.Error("Border.Lines should be written even when Lines are nil")
	}
	if _, ok := w.data["Border.Color"]; ok {
		t.Error("Border.Color should not be written when Lines are nil")
	}
}

// ── serializeBorder per-line overrides (non-equal lines) ─────────────────────

func TestSerializeBorder_PerLineOverrides(t *testing.T) {
	b := style.NewBorder()
	// Set different colors on different lines → allEqual = false → per-line path.
	b.Lines[0].Color = color.RGBA{R: 255, G: 0, B: 0, A: 255} // Left: red
	b.Lines[1].Color = color.RGBA{R: 0, G: 255, B: 0, A: 255} // Top: green
	b.Lines[2].Color = color.RGBA{R: 0, G: 0, B: 255, A: 255} // Right: blue
	b.Lines[3].Color = color.RGBA{R: 255, G: 255, B: 0, A: 255} // Bottom: yellow
	w := newTestWriter()
	serializeBorder(w, b)
	// Per-line attributes should be written.
	if _, ok := w.data["Border.LeftLine.Color"]; !ok {
		t.Error("Border.LeftLine.Color should be written in per-line mode")
	}
	if _, ok := w.data["Border.TopLine.Color"]; !ok {
		t.Error("Border.TopLine.Color should be written in per-line mode")
	}
	if _, ok := w.data["Border.RightLine.Color"]; !ok {
		t.Error("Border.RightLine.Color should be written in per-line mode")
	}
	if _, ok := w.data["Border.BottomLine.Color"]; !ok {
		t.Error("Border.BottomLine.Color should be written in per-line mode")
	}
	// Common Color should NOT be written.
	if _, ok := w.data["Border.Color"]; ok {
		t.Error("Border.Color should not be written in per-line mode")
	}
}

func TestSerializeBorder_PerLine_NonSolidStyle(t *testing.T) {
	b := style.NewBorder()
	b.Lines[0].Style = style.LineStyleDash
	b.Lines[1].Style = style.LineStyleDot
	b.Lines[2].Style = style.LineStyleSolid
	b.Lines[3].Style = style.LineStyleSolid
	w := newTestWriter()
	serializeBorder(w, b)
	if _, ok := w.data["Border.LeftLine.Style"]; !ok {
		t.Error("Border.LeftLine.Style should be written for Dash")
	}
	if _, ok := w.data["Border.TopLine.Style"]; !ok {
		t.Error("Border.TopLine.Style should be written for Dot")
	}
}

func TestSerializeBorder_PerLine_NonDefaultWidth(t *testing.T) {
	b := style.NewBorder()
	b.Lines[0].Width = 3
	b.Lines[1].Width = 1 // default — same as others? No, 0 and 2 differ
	b.Lines[2].Width = 2
	b.Lines[3].Width = 1
	w := newTestWriter()
	serializeBorder(w, b)
	if _, ok := w.data["Border.LeftLine.Width"]; !ok {
		t.Error("Border.LeftLine.Width should be written for width 3")
	}
}

func TestSerializeBorder_PerLine_NilLine(t *testing.T) {
	// In per-line mode, nil lines should be skipped (continue).
	b := style.NewBorder()
	b.Lines[0].Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	b.Lines[1] = nil // nil line — should not panic
	b.Lines[2].Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	b.Lines[3].Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	w := newTestWriter()
	serializeBorder(w, b)
	// Should not panic; LeftLine and RightLine+BottomLine written; TopLine skipped.
	if _, ok := w.data["Border.LeftLine.Color"]; !ok {
		t.Error("Border.LeftLine.Color should be written")
	}
}

// ── deserializeBorder — ShadowColor and per-line attributes ───────────────────

func TestDeserializeBorder_ShadowColor(t *testing.T) {
	r := newTestReader(map[string]any{
		"Border.Shadow":      true,
		"Border.ShadowColor": "#FFFF0000", // red
	})
	b := style.NewBorder()
	deserializeBorder(r, b)
	if !b.Shadow {
		t.Error("Shadow should be true")
	}
	want := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if b.ShadowColor != want {
		t.Errorf("ShadowColor = %v, want %v", b.ShadowColor, want)
	}
}

func TestDeserializeBorder_PerLineAttributes(t *testing.T) {
	r := newTestReader(map[string]any{
		"Border.LeftLine.Color": "#FFFF0000",
		"Border.LeftLine.Style": "Dash",
		"Border.LeftLine.Width": float32(3),
	})
	b := style.NewBorder()
	deserializeBorder(r, b)
	want := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if b.Lines[0].Color != want {
		t.Errorf("LeftLine.Color = %v, want %v", b.Lines[0].Color, want)
	}
	if b.Lines[0].Style != style.LineStyleDash {
		t.Errorf("LeftLine.Style = %v, want Dash", b.Lines[0].Style)
	}
	if b.Lines[0].Width != 3 {
		t.Errorf("LeftLine.Width = %v, want 3", b.Lines[0].Width)
	}
}

func TestDeserializeBorder_InitializesNilLines(t *testing.T) {
	// If border has nil Lines, deserializeBorder should init them.
	b := &style.Border{}
	for i := range b.Lines {
		b.Lines[i] = nil
	}
	r := newTestReader(map[string]any{})
	deserializeBorder(r, b)
	for i, l := range b.Lines {
		if l == nil {
			t.Errorf("Lines[%d] should not be nil after deserializeBorder", i)
		}
	}
}

func TestDeserializeBorder_CommonStyle(t *testing.T) {
	// Border.Style sets style on all lines (line 216-220).
	r := newTestReader(map[string]any{
		"Border.Style": "Dash",
		"Border.Width": float32(2),
	})
	b := style.NewBorder()
	deserializeBorder(r, b)
	for i, l := range b.Lines {
		if l.Style != style.LineStyleDash {
			t.Errorf("Lines[%d].Style = %v, want Dash", i, l.Style)
		}
		if l.Width != 2 {
			t.Errorf("Lines[%d].Width = %v, want 2", i, l.Width)
		}
	}
}

func TestDeserializeFill_Default_NonSolidCurrentFill(t *testing.T) {
	// default branch with current != *SolidFill → creates new SolidFill (line 376-378).
	r := newTestReader(map[string]any{}) // no Fill type attr → default branch
	got := deserializeFill(r, &style.NoneFill{}) // not a SolidFill
	if _, ok := got.(*style.SolidFill); !ok {
		t.Errorf("expected *style.SolidFill, got %T", got)
	}
}

// ── serializeFill GlassFill and HatchFill ─────────────────────────────────────

func TestSerializeFill_Nil(t *testing.T) {
	w := newTestWriter()
	serializeFill(w, nil)
	if len(w.data) != 0 {
		t.Errorf("nil fill should produce no output, got %v", w.data)
	}
}

func TestSerializeFill_GlassFill(t *testing.T) {
	f := &style.GlassFill{
		Color: color.RGBA{R: 100, G: 150, B: 200, A: 255},
		Blend: 0.5,
		Hatch: false,
	}
	w := newTestWriter()
	serializeFill(w, f)
	if v, ok := w.data["Fill"]; !ok || v != "Glass" {
		t.Errorf("Fill = %v, want Glass", v)
	}
	if _, ok := w.data["Fill.Color"]; !ok {
		t.Error("Fill.Color should be written for non-transparent GlassFill")
	}
	if _, ok := w.data["Fill.Blend"]; !ok {
		t.Error("Fill.Blend should be written when not 0.2")
	}
	if _, ok := w.data["Fill.Hatch"]; !ok {
		t.Error("Fill.Hatch should be written when false (not default true)")
	}
}

func TestSerializeFill_HatchFill(t *testing.T) {
	f := &style.HatchFill{
		ForeColor: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		BackColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Style:     style.HatchStyle(5),
	}
	w := newTestWriter()
	serializeFill(w, f)
	if v, ok := w.data["Fill"]; !ok || v != "Hatch" {
		t.Errorf("Fill = %v, want Hatch", v)
	}
	if _, ok := w.data["Fill.ForeColor"]; !ok {
		t.Error("Fill.ForeColor should be written")
	}
	if _, ok := w.data["Fill.BackColor"]; !ok {
		t.Error("Fill.BackColor should be written")
	}
	if v, ok := w.data["Fill.Style"]; !ok || v != "DiagonalCross" {
		t.Errorf("Fill.Style = %v, want DiagonalCross", v)
	}
}

// ── deserializeFill GlassFill and HatchFill ───────────────────────────────────

func TestDeserializeFill_GlassRoundTrip(t *testing.T) {
	orig := &style.GlassFill{
		Color: color.RGBA{R: 100, G: 150, B: 200, A: 255},
		Blend: 0.5,
		Hatch: false,
	}
	w := newTestWriter()
	serializeFill(w, orig)
	r := newTestReader(w.data)
	got := deserializeFill(r, &style.SolidFill{})
	gf, ok := got.(*style.GlassFill)
	if !ok {
		t.Fatalf("got fill type %T, want *style.GlassFill", got)
	}
	if gf.Color != orig.Color {
		t.Errorf("Color = %v, want %v", gf.Color, orig.Color)
	}
	if gf.Blend != orig.Blend {
		t.Errorf("Blend = %v, want %v", gf.Blend, orig.Blend)
	}
}

func TestDeserializeFill_HatchRoundTrip(t *testing.T) {
	orig := &style.HatchFill{
		ForeColor: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		BackColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Style:     style.HatchStyle(3),
	}
	w := newTestWriter()
	serializeFill(w, orig)
	r := newTestReader(w.data)
	got := deserializeFill(r, &style.SolidFill{})
	hf, ok := got.(*style.HatchFill)
	if !ok {
		t.Fatalf("got fill type %T, want *style.HatchFill", got)
	}
	if hf.ForeColor != orig.ForeColor {
		t.Errorf("ForeColor = %v, want %v", hf.ForeColor, orig.ForeColor)
	}
	if hf.Style != orig.Style {
		t.Errorf("Style = %v, want %v", hf.Style, orig.Style)
	}
}

func TestDeserializeFill_LinearGradient_WithColors(t *testing.T) {
	r := newTestReader(map[string]any{
		"Fill":            "LinearGradient",
		"Fill.StartColor": "#FFFF0000",
		"Fill.EndColor":   "#FF0000FF",
	})
	got := deserializeFill(r, &style.SolidFill{})
	lf, ok := got.(*style.LinearGradientFill)
	if !ok {
		t.Fatalf("got %T, want *LinearGradientFill", got)
	}
	wantStart := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if lf.StartColor != wantStart {
		t.Errorf("StartColor = %v, want %v", lf.StartColor, wantStart)
	}
	wantEnd := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	if lf.EndColor != wantEnd {
		t.Errorf("EndColor = %v, want %v", lf.EndColor, wantEnd)
	}
}

func TestDeserializeFill_GlassFill_WithColor(t *testing.T) {
	r := newTestReader(map[string]any{
		"Fill":       "Glass",
		"Fill.Color": "#FF6496C8",
	})
	got := deserializeFill(r, &style.SolidFill{})
	if _, ok := got.(*style.GlassFill); !ok {
		t.Fatalf("got %T, want *GlassFill", got)
	}
}

func TestDeserializeFill_HatchFill_WithColors(t *testing.T) {
	r := newTestReader(map[string]any{
		"Fill":           "Hatch",
		"Fill.ForeColor": "#FF000000",
		"Fill.BackColor": "#FFFFFFFF",
	})
	got := deserializeFill(r, &style.SolidFill{})
	if _, ok := got.(*style.HatchFill); !ok {
		t.Fatalf("got %T, want *HatchFill", got)
	}
}

// ── serializeFill LinearGradient Focus/Contrast non-defaults ─────────────────

func TestSerializeFill_LinearGradient_FocusContrast(t *testing.T) {
	f := &style.LinearGradientFill{
		StartColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Focus:      0.5,
		Contrast:   0.8,
	}
	w := newTestWriter()
	serializeFill(w, f)
	if _, ok := w.data["Fill.Focus"]; !ok {
		t.Error("Fill.Focus should be written when non-zero")
	}
	if _, ok := w.data["Fill.Contrast"]; !ok {
		t.Error("Fill.Contrast should be written when not 1")
	}
	_ = strings.Contains // suppress import warning
}

// ── PathGradientFill serialize / deserialize ──────────────────────────────────

func TestSerializeFill_PathGradient(t *testing.T) {
	f := style.NewPathGradientFill(
		color.RGBA{R: 0, G: 0, B: 0, A: 255},   // CenterColor: black
		color.RGBA{R: 255, G: 255, B: 255, A: 255}, // EdgeColor: white
		style.PathGradientElliptic,
	)
	w := newTestWriter()
	serializeFill(w, f)
	if v, ok := w.data["Fill"]; !ok || v != "PathGradient" {
		t.Errorf("Fill = %v, want PathGradient", v)
	}
	if _, ok := w.data["Fill.CenterColor"]; !ok {
		t.Error("Fill.CenterColor should be written")
	}
	if _, ok := w.data["Fill.EdgeColor"]; !ok {
		t.Error("Fill.EdgeColor should be written")
	}
	// Elliptic is the default — should not be written.
	if _, ok := w.data["Fill.Style"]; ok {
		t.Error("Fill.Style should not be written for default Elliptic")
	}
}

func TestSerializeFill_PathGradient_Rectangular(t *testing.T) {
	f := style.NewPathGradientFill(
		color.RGBA{R: 100, G: 100, B: 100, A: 255},
		color.RGBA{R: 200, G: 200, B: 200, A: 255},
		style.PathGradientRectangular,
	)
	w := newTestWriter()
	serializeFill(w, f)
	if v, ok := w.data["Fill.Style"]; !ok || v != "Rectangular" {
		t.Errorf("Fill.Style = %v, want Rectangular", v)
	}
}

func TestDeserializeFill_PathGradientRoundTrip(t *testing.T) {
	orig := style.NewPathGradientFill(
		color.RGBA{R: 10, G: 20, B: 30, A: 255},
		color.RGBA{R: 200, G: 210, B: 220, A: 255},
		style.PathGradientRectangular,
	)
	w := newTestWriter()
	serializeFill(w, orig)
	r := newTestReader(w.data)
	got := deserializeFill(r, &style.SolidFill{})
	pg, ok := got.(*style.PathGradientFill)
	if !ok {
		t.Fatalf("got fill type %T, want *style.PathGradientFill", got)
	}
	if pg.CenterColor != orig.CenterColor {
		t.Errorf("CenterColor = %v, want %v", pg.CenterColor, orig.CenterColor)
	}
	if pg.EdgeColor != orig.EdgeColor {
		t.Errorf("EdgeColor = %v, want %v", pg.EdgeColor, orig.EdgeColor)
	}
	if pg.Style != orig.Style {
		t.Errorf("Style = %v, want %v", pg.Style, orig.Style)
	}
}

func TestDeserializeFill_PathGradient_NamedColors(t *testing.T) {
	// Matches real FRX: Fill="PathGradient" Fill.CenterColor="Black" Fill.EdgeColor="White" Fill.Style="Elliptic"
	r := newTestReader(map[string]any{
		"Fill":            "PathGradient",
		"Fill.CenterColor": "Black",
		"Fill.EdgeColor":   "White",
		"Fill.Style":       "Elliptic",
	})
	got := deserializeFill(r, &style.SolidFill{})
	pg, ok := got.(*style.PathGradientFill)
	if !ok {
		t.Fatalf("got fill type %T, want *style.PathGradientFill", got)
	}
	wantCenter := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	if pg.CenterColor != wantCenter {
		t.Errorf("CenterColor = %v, want Black %v", pg.CenterColor, wantCenter)
	}
	wantEdge := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if pg.EdgeColor != wantEdge {
		t.Errorf("EdgeColor = %v, want White %v", pg.EdgeColor, wantEdge)
	}
	if pg.Style != style.PathGradientElliptic {
		t.Errorf("Style = %v, want Elliptic", pg.Style)
	}
}

// ── TextureFill serialize / deserialize ───────────────────────────────────────

func TestSerializeFill_Texture(t *testing.T) {
	f := &style.TextureFill{
		ImageData:           []byte{0x89, 0x50, 0x4e, 0x47}, // PNG magic bytes
		ImageWidth:          140,
		ImageHeight:         170,
		PreserveAspectRatio: true,
		WrapMode:            style.WrapModeClamp,
		ImageOffsetX:        20,
		ImageOffsetY:        0,
	}
	w := newTestWriter()
	serializeFill(w, f)
	if v, ok := w.data["Fill"]; !ok || v != "Texture" {
		t.Errorf("Fill = %v, want Texture", v)
	}
	if v, ok := w.data["Fill.ImageWidth"]; !ok || v != 140 {
		t.Errorf("Fill.ImageWidth = %v, want 140", v)
	}
	if v, ok := w.data["Fill.ImageHeight"]; !ok || v != 170 {
		t.Errorf("Fill.ImageHeight = %v, want 170", v)
	}
	if v, ok := w.data["Fill.PreserveAspectRatio"]; !ok || v != true {
		t.Errorf("Fill.PreserveAspectRatio = %v, want true", v)
	}
	if v, ok := w.data["Fill.WrapMode"]; !ok || v != "Clamp" {
		t.Errorf("Fill.WrapMode = %v, want Clamp", v)
	}
	if v, ok := w.data["Fill.ImageOffsetX"]; !ok || v != 20 {
		t.Errorf("Fill.ImageOffsetX = %v, want 20", v)
	}
	// ImageOffsetY is 0 — should not be written.
	if _, ok := w.data["Fill.ImageOffsetY"]; ok {
		t.Error("Fill.ImageOffsetY should not be written when 0")
	}
	if _, ok := w.data["Fill.ImageData"]; !ok {
		t.Error("Fill.ImageData should be written")
	}
}

func TestDeserializeFill_TextureRoundTrip(t *testing.T) {
	orig := &style.TextureFill{
		ImageData:           []byte{1, 2, 3, 4, 5},
		ImageWidth:          100,
		ImageHeight:         80,
		PreserveAspectRatio: true,
		WrapMode:            style.WrapModeTileFlipX,
		ImageOffsetX:        5,
		ImageOffsetY:        10,
	}
	w := newTestWriter()
	serializeFill(w, orig)
	r := newTestReader(w.data)
	got := deserializeFill(r, &style.SolidFill{})
	tf, ok := got.(*style.TextureFill)
	if !ok {
		t.Fatalf("got fill type %T, want *style.TextureFill", got)
	}
	if tf.ImageWidth != orig.ImageWidth {
		t.Errorf("ImageWidth = %d, want %d", tf.ImageWidth, orig.ImageWidth)
	}
	if tf.ImageHeight != orig.ImageHeight {
		t.Errorf("ImageHeight = %d, want %d", tf.ImageHeight, orig.ImageHeight)
	}
	if tf.PreserveAspectRatio != orig.PreserveAspectRatio {
		t.Errorf("PreserveAspectRatio = %v, want %v", tf.PreserveAspectRatio, orig.PreserveAspectRatio)
	}
	if tf.WrapMode != orig.WrapMode {
		t.Errorf("WrapMode = %v, want %v", tf.WrapMode, orig.WrapMode)
	}
	if tf.ImageOffsetX != orig.ImageOffsetX {
		t.Errorf("ImageOffsetX = %d, want %d", tf.ImageOffsetX, orig.ImageOffsetX)
	}
	if tf.ImageOffsetY != orig.ImageOffsetY {
		t.Errorf("ImageOffsetY = %d, want %d", tf.ImageOffsetY, orig.ImageOffsetY)
	}
	if len(tf.ImageData) != len(orig.ImageData) {
		t.Errorf("ImageData length = %d, want %d", len(tf.ImageData), len(orig.ImageData))
	}
}

func TestDeserializeFill_Texture_DefaultWrapMode(t *testing.T) {
	// No Fill.WrapMode → defaults to Tile.
	r := newTestReader(map[string]any{
		"Fill": "Texture",
	})
	got := deserializeFill(r, &style.SolidFill{})
	tf, ok := got.(*style.TextureFill)
	if !ok {
		t.Fatalf("got fill type %T, want *style.TextureFill", got)
	}
	if tf.WrapMode != style.WrapModeTile {
		t.Errorf("WrapMode = %v, want WrapModeTile", tf.WrapMode)
	}
}

func TestDeserializeFill_Texture_NoPanic_NoImageData(t *testing.T) {
	// Loading a Texture fill with no image data should not panic (stub/graceful load).
	r := newTestReader(map[string]any{
		"Fill":            "Texture",
		"Fill.ImageWidth": 140,
		"Fill.ImageHeight": 170,
		"Fill.PreserveAspectRatio": true,
		"Fill.WrapMode":   "Clamp",
		"Fill.ImageOffsetX": 20,
		"Fill.ImageOffsetY": 0,
	})
	got := deserializeFill(r, &style.SolidFill{})
	if _, ok := got.(*style.TextureFill); !ok {
		t.Fatalf("got fill type %T, want *style.TextureFill", got)
	}
	// No panic — pass.
}

// ── parseLineStyle "Custom" ───────────────────────────────────────────────────

func TestParseLineStyle_Custom(t *testing.T) {
	// LineStyle.Custom is a GDI+ rendering concept only; we map it to Solid
	// since no FRX DashPattern attribute is serialized.
	got := parseLineStyle("Custom")
	if got != style.LineStyleSolid {
		t.Errorf("parseLineStyle(Custom) = %v, want Solid", got)
	}
}

// ── formatPathGradientStyle / parsePathGradientStyle ─────────────────────────

func TestFormatParsePathGradientStyle(t *testing.T) {
	cases := []struct {
		s     style.PathGradientStyle
		name  string
	}{
		{style.PathGradientElliptic, "Elliptic"},
		{style.PathGradientRectangular, "Rectangular"},
	}
	for _, tc := range cases {
		got := formatPathGradientStyle(tc.s)
		if got != tc.name {
			t.Errorf("formatPathGradientStyle(%v) = %q, want %q", tc.s, got, tc.name)
		}
		back := parsePathGradientStyle(tc.name)
		if back != tc.s {
			t.Errorf("parsePathGradientStyle(%q) = %v, want %v", tc.name, back, tc.s)
		}
	}
	// Unknown defaults to Elliptic.
	if parsePathGradientStyle("unknown") != style.PathGradientElliptic {
		t.Error("unknown pathGradient style should default to Elliptic")
	}
}

// ── formatWrapMode / parseWrapMode ────────────────────────────────────────────

func TestFormatParseWrapMode(t *testing.T) {
	cases := []struct {
		w    style.WrapMode
		name string
	}{
		{style.WrapModeTile, "Tile"},
		{style.WrapModeTileFlipX, "TileFlipX"},
		{style.WrapModeTileFlipY, "TileFlipY"},
		{style.WrapModeTileFlipXY, "TileFlipXY"},
		{style.WrapModeClamp, "Clamp"},
	}
	for _, tc := range cases {
		got := formatWrapMode(tc.w)
		if got != tc.name {
			t.Errorf("formatWrapMode(%v) = %q, want %q", tc.w, got, tc.name)
		}
		back := parseWrapMode(tc.name)
		if back != tc.w {
			t.Errorf("parseWrapMode(%q) = %v, want %v", tc.name, back, tc.w)
		}
	}
	// Unknown defaults to Tile.
	if parseWrapMode("unknown") != style.WrapModeTile {
		t.Error("unknown WrapMode should default to Tile")
	}
}
