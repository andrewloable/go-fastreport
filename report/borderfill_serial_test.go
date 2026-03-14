package report

import (
	"image/color"
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
