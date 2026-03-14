package utils

import (
	"image/color"
	"testing"
)

// ── BoolFromString ────────────────────────────────────────────────────────────

func TestBoolFromString(t *testing.T) {
	trueInputs := []string{"true", "True", "TRUE", "  true  ", "TrUe"}
	for _, s := range trueInputs {
		if !BoolFromString(s) {
			t.Errorf("BoolFromString(%q) expected true", s)
		}
	}

	falseInputs := []string{"false", "False", "FALSE", "0", "1", "", "yes", "no"}
	for _, s := range falseInputs {
		if BoolFromString(s) {
			t.Errorf("BoolFromString(%q) expected false", s)
		}
	}
}

// ── BoolToString ──────────────────────────────────────────────────────────────

func TestBoolToString(t *testing.T) {
	if BoolToString(true) != "true" {
		t.Errorf("BoolToString(true) = %q, want %q", BoolToString(true), "true")
	}
	if BoolToString(false) != "false" {
		t.Errorf("BoolToString(false) = %q, want %q", BoolToString(false), "false")
	}
}

// ── IntFromString ─────────────────────────────────────────────────────────────

func TestIntFromString(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"0", 0},
		{"42", 42},
		{"-7", -7},
		{"  100  ", 100},
		{"", 0},
		{"abc", 0},
		{"3.14", 0},
	}
	for _, tc := range tests {
		got := IntFromString(tc.input)
		if got != tc.want {
			t.Errorf("IntFromString(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// ── IntToString ───────────────────────────────────────────────────────────────

func TestIntToString(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{42, "42"},
		{-7, "-7"},
	}
	for _, tc := range tests {
		got := IntToString(tc.input)
		if got != tc.want {
			t.Errorf("IntToString(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ── Float32FromString ─────────────────────────────────────────────────────────

func TestFloat32FromString(t *testing.T) {
	tests := []struct {
		input string
		want  float32
	}{
		{"0", 0},
		{"3.14", 3.14},
		{"-1.5", -1.5},
		{"  2.5  ", 2.5},
		{"", 0},
		{"abc", 0},
	}
	for _, tc := range tests {
		got := Float32FromString(tc.input)
		// Allow a small tolerance for float32 precision.
		diff := got - tc.want
		if diff < 0 {
			diff = -diff
		}
		if diff > 1e-5 {
			t.Errorf("Float32FromString(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

// ── Float32ToString ───────────────────────────────────────────────────────────

func TestFloat32ToString(t *testing.T) {
	tests := []struct {
		input float32
		want  string
	}{
		{0, "0"},
		{1.5, "1.5"},
		{-3.14, "-3.14"},
	}
	for _, tc := range tests {
		got := Float32ToString(tc.input)
		if got != tc.want {
			t.Errorf("Float32ToString(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFloat32RoundTrip(t *testing.T) {
	values := []float32{0, 1, -1, 3.14, 100.5, -0.001}
	for _, v := range values {
		s := Float32ToString(v)
		got := Float32FromString(s)
		diff := got - v
		if diff < 0 {
			diff = -diff
		}
		if diff > 1e-5 {
			t.Errorf("round-trip failed for %v: got %v via %q", v, got, s)
		}
	}
}

// ── RGBAFromString / RGBAToString ─────────────────────────────────────────────

func TestRGBAFromString_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  color.RGBA
	}{
		{"#FF0000", color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"#FFFF0000", color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"#F00", color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"-65536", color.RGBA{R: 255, G: 0, B: 0, A: 255}}, // .NET Color.Red.ToArgb()
	}
	for _, tc := range tests {
		got := RGBAFromString(tc.input)
		if got != tc.want {
			t.Errorf("RGBAFromString(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestRGBAFromString_Invalid(t *testing.T) {
	// Invalid input should return zero value, not panic.
	got := RGBAFromString("notacolor")
	if got != (color.RGBA{}) {
		t.Errorf("RGBAFromString(invalid) = %v, want zero RGBA", got)
	}
}

func TestRGBAToString(t *testing.T) {
	c := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	got := RGBAToString(c)
	if got != "#FFFF0000" {
		t.Errorf("RGBAToString(%v) = %q, want %q", c, got, "#FFFF0000")
	}
}

func TestRGBARoundTrip(t *testing.T) {
	colors := []color.RGBA{
		{R: 0, G: 0, B: 0, A: 0},
		{R: 255, G: 255, B: 255, A: 255},
		{R: 1, G: 2, B: 3, A: 4},
	}
	for _, c := range colors {
		s := RGBAToString(c)
		got := RGBAFromString(s)
		if got != c {
			t.Errorf("round-trip failed for %v: got %v via %q", c, got, s)
		}
	}
}
