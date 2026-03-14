package utils

import (
	"image/color"
	"testing"
)

func TestParseColor_Hex6(t *testing.T) {
	tests := []struct {
		input string
		want  color.RGBA
	}{
		{"#FF0000", color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"#00FF00", color.RGBA{R: 0, G: 255, B: 0, A: 255}},
		{"#0000FF", color.RGBA{R: 0, G: 0, B: 255, A: 255}},
		{"#000000", color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		{"#FFFFFF", color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		// lowercase is also valid hex
		{"#ff8800", color.RGBA{R: 255, G: 136, B: 0, A: 255}},
	}
	for _, tc := range tests {
		got, err := ParseColor(tc.input)
		if err != nil {
			t.Errorf("ParseColor(%q) error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseColor(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParseColor_Hex8(t *testing.T) {
	tests := []struct {
		input string
		want  color.RGBA
	}{
		{"#FFFF0000", color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"#80FF0000", color.RGBA{R: 255, G: 0, B: 0, A: 128}},
		{"#00000000", color.RGBA{R: 0, G: 0, B: 0, A: 0}},
	}
	for _, tc := range tests {
		got, err := ParseColor(tc.input)
		if err != nil {
			t.Errorf("ParseColor(%q) error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseColor(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParseColor_Hex3(t *testing.T) {
	tests := []struct {
		input string
		want  color.RGBA
	}{
		{"#F00", color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"#0F0", color.RGBA{R: 0, G: 255, B: 0, A: 255}},
		{"#FFF", color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"#000", color.RGBA{R: 0, G: 0, B: 0, A: 255}},
	}
	for _, tc := range tests {
		got, err := ParseColor(tc.input)
		if err != nil {
			t.Errorf("ParseColor(%q) error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseColor(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParseColor_DecimalARGB(t *testing.T) {
	// -65536 is the .NET int32 representation of Color.Red (ARGB = 0xFFFF0000)
	got, err := ParseColor("-65536")
	if err != nil {
		t.Fatalf("ParseColor(\"-65536\") error: %v", err)
	}
	want := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if got != want {
		t.Errorf("ParseColor(\"-65536\") = %v, want %v", got, want)
	}

	// 0 → transparent black
	got, err = ParseColor("0")
	if err != nil {
		t.Fatalf("ParseColor(\"0\") error: %v", err)
	}
	if got != (color.RGBA{}) {
		t.Errorf("ParseColor(\"0\") = %v, want zero RGBA", got)
	}
}

func TestParseColor_Whitespace(t *testing.T) {
	got, err := ParseColor("  #FF0000  ")
	if err != nil {
		t.Fatalf("ParseColor with whitespace error: %v", err)
	}
	want := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if got != want {
		t.Errorf("ParseColor with whitespace = %v, want %v", got, want)
	}
}

func TestParseColor_Errors(t *testing.T) {
	bad := []string{
		"",
		"#",
		"#12",
		"#1234",
		"#12345",
		"#1234567",
		"#ZZZZZZ",
		"notacolor",
		"red",
	}
	for _, s := range bad {
		_, err := ParseColor(s)
		if err == nil {
			t.Errorf("ParseColor(%q) expected error, got nil", s)
		}
	}
}

func TestFormatColor(t *testing.T) {
	tests := []struct {
		c    color.RGBA
		want string
	}{
		{color.RGBA{R: 255, G: 0, B: 0, A: 255}, "#FFFF0000"},
		{color.RGBA{R: 0, G: 0, B: 0, A: 0}, "#00000000"},
		{color.RGBA{R: 255, G: 255, B: 255, A: 255}, "#FFFFFFFF"},
		{color.RGBA{R: 0, G: 128, B: 0, A: 255}, "#FF008000"},
	}
	for _, tc := range tests {
		got := FormatColor(tc.c)
		if got != tc.want {
			t.Errorf("FormatColor(%v) = %q, want %q", tc.c, got, tc.want)
		}
	}
}

func TestColorEqual(t *testing.T) {
	a := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	b := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	c := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	if !ColorEqual(a, b) {
		t.Error("expected ColorEqual true for identical colors")
	}
	if ColorEqual(a, c) {
		t.Error("expected ColorEqual false for different colors")
	}
}

func TestPredefinedColors(t *testing.T) {
	if ColorTransparent.A != 0 {
		t.Errorf("ColorTransparent should have A=0, got A=%d", ColorTransparent.A)
	}
	if ColorBlack != (color.RGBA{R: 0, G: 0, B: 0, A: 255}) {
		t.Errorf("ColorBlack wrong: %v", ColorBlack)
	}
	if ColorWhite != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Errorf("ColorWhite wrong: %v", ColorWhite)
	}
	if ColorRed != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Errorf("ColorRed wrong: %v", ColorRed)
	}
	if ColorGreen != (color.RGBA{R: 0, G: 128, B: 0, A: 255}) {
		t.Errorf("ColorGreen wrong: %v", ColorGreen)
	}
	if ColorBlue != (color.RGBA{R: 0, G: 0, B: 255, A: 255}) {
		t.Errorf("ColorBlue wrong: %v", ColorBlue)
	}
}

func TestParseFormatRoundTrip(t *testing.T) {
	colors := []color.RGBA{
		ColorTransparent,
		ColorBlack,
		ColorWhite,
		ColorRed,
		ColorGreen,
		ColorBlue,
		{R: 1, G: 2, B: 3, A: 4},
	}
	for _, c := range colors {
		s := FormatColor(c)
		got, err := ParseColor(s)
		if err != nil {
			t.Errorf("ParseColor(FormatColor(%v)) error: %v", c, err)
			continue
		}
		if got != c {
			t.Errorf("round-trip failed: FormatColor(%v) = %q, ParseColor = %v", c, s, got)
		}
	}
}
