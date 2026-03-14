package units_test

import (
	"math"
	"testing"

	"github.com/loabletech/go-fastreport/units"
)

func almostEqual(a, b, tolerance float32) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

func TestConstants(t *testing.T) {
	tests := []struct {
		name  string
		got   float32
		want  float32
	}{
		{"Millimeters", units.Millimeters, 3.78},
		{"Centimeters", units.Centimeters, 37.8},
		{"Inches", units.Inches, 96},
		{"TenthsOfInch", units.TenthsOfInch, 9.6},
		{"HundrethsOfInch", units.HundrethsOfInch, 0.96},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestConvertSameUnit(t *testing.T) {
	for _, u := range []units.PageUnits{
		units.PageUnitsMillimeters,
		units.PageUnitsCentimeters,
		units.PageUnitsInches,
		units.PageUnitsHundrethsOfInch,
	} {
		got := units.Convert(100, u, u)
		if got != 100 {
			t.Errorf("Convert(100, %v, %v) = %v, want 100", u, u, got)
		}
	}
}

func TestConvertMMToInches(t *testing.T) {
	// 25.4mm = 1 inch
	got := units.Convert(25.4, units.PageUnitsMillimeters, units.PageUnitsInches)
	want := float32(1.0)
	if !almostEqual(got, want, 0.001) {
		t.Errorf("Convert(25.4, MM, Inches) = %v, want ~%v", got, want)
	}
}

func TestConvertInchesToMM(t *testing.T) {
	// 1 inch = 25.4mm
	got := units.Convert(1.0, units.PageUnitsInches, units.PageUnitsMillimeters)
	want := float32(25.4)
	if !almostEqual(got, want, 0.01) {
		t.Errorf("Convert(1, Inches, MM) = %v, want ~%v", got, want)
	}
}

func TestConvertMMToCM(t *testing.T) {
	// 10mm = 1cm
	got := units.Convert(10, units.PageUnitsMillimeters, units.PageUnitsCentimeters)
	want := float32(1.0)
	if !almostEqual(got, want, 0.001) {
		t.Errorf("Convert(10, MM, CM) = %v, want ~%v", got, want)
	}
}

func TestConvertRoundTrip(t *testing.T) {
	orig := float32(210) // A4 width in mm
	inInches := units.Convert(orig, units.PageUnitsMillimeters, units.PageUnitsInches)
	backToMM := units.Convert(inInches, units.PageUnitsInches, units.PageUnitsMillimeters)
	if !almostEqual(orig, backToMM, 0.01) {
		t.Errorf("round-trip MM→in→MM: got %v, want %v", backToMM, orig)
	}
}

func TestConvertZero(t *testing.T) {
	got := units.Convert(0, units.PageUnitsMillimeters, units.PageUnitsInches)
	if got != 0 {
		t.Errorf("Convert(0, ...) = %v, want 0", got)
	}
}

func TestConvertNegative(t *testing.T) {
	pos := units.Convert(10, units.PageUnitsMillimeters, units.PageUnitsInches)
	neg := units.Convert(-10, units.PageUnitsMillimeters, units.PageUnitsInches)
	if !almostEqual(pos, -neg, 0.001) {
		t.Errorf("negative conversion asymmetric: pos=%v, neg=%v", pos, neg)
	}
}

func TestToPixels(t *testing.T) {
	tests := []struct {
		value float32
		unit  units.PageUnits
		want  float32
	}{
		{1, units.PageUnitsMillimeters, 3.78},
		{1, units.PageUnitsCentimeters, 37.8},
		{1, units.PageUnitsInches, 96},
		{1, units.PageUnitsHundrethsOfInch, 0.96},
	}
	for _, tt := range tests {
		got := units.ToPixels(tt.value, tt.unit)
		if !almostEqual(got, tt.want, 0.001) {
			t.Errorf("ToPixels(%v, %v) = %v, want %v", tt.value, tt.unit, got, tt.want)
		}
	}
}

func TestFromPixels(t *testing.T) {
	tests := []struct {
		pixels float32
		unit   units.PageUnits
		want   float32
	}{
		{3.78, units.PageUnitsMillimeters, 1},
		{37.8, units.PageUnitsCentimeters, 1},
		{96, units.PageUnitsInches, 1},
		{0.96, units.PageUnitsHundrethsOfInch, 1},
	}
	for _, tt := range tests {
		got := units.FromPixels(tt.pixels, tt.unit)
		if !almostEqual(got, tt.want, 0.001) {
			t.Errorf("FromPixels(%v, %v) = %v, want %v", tt.pixels, tt.unit, got, tt.want)
		}
	}
}

func TestConvertUnknownUnit(t *testing.T) {
	// An unknown PageUnits value falls through to the default (Millimeters).
	unknown := units.PageUnits(99)
	// Converting 1 unknown unit to MM should give 1 (same factor = Millimeters).
	got := units.Convert(1, unknown, unknown)
	if got != 1 {
		t.Errorf("Convert with unknown unit same-to-same = %v, want 1", got)
	}
	// Converting 1 unknown to inches should equal mm→inches
	gotUnknown := units.Convert(1, unknown, units.PageUnitsInches)
	gotMM := units.Convert(1, units.PageUnitsMillimeters, units.PageUnitsInches)
	if !almostEqual(gotUnknown, gotMM, 0.001) {
		t.Errorf("unknown unit does not fall back to mm: got %v, want %v", gotUnknown, gotMM)
	}
}

func TestToFromPixelsRoundTrip(t *testing.T) {
	for _, u := range []units.PageUnits{
		units.PageUnitsMillimeters,
		units.PageUnitsCentimeters,
		units.PageUnitsInches,
		units.PageUnitsHundrethsOfInch,
	} {
		orig := float32(42.5)
		pixels := units.ToPixels(orig, u)
		got := units.FromPixels(pixels, u)
		if !almostEqual(got, orig, float32(math.SmallestNonzeroFloat32)*1000) {
			t.Errorf("round-trip ToPixels/FromPixels unit=%v: got %v, want %v", u, got, orig)
		}
	}
}
