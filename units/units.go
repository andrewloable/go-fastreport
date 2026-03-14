// Package units provides constants and utilities for converting between
// report measurement units and screen pixels.
//
// Internal measurements are stored in screen pixels (96 DPI). Use the
// conversion constants to convert to/from other units:
//
//	valueInPixels := valueInMM * units.Millimeters
//	valueInMM := valueInPixels / units.Millimeters
package units

// Pixel-per-unit conversion constants.
const (
	// Millimeters is the number of pixels per millimeter.
	Millimeters float32 = 3.78

	// Centimeters is the number of pixels per centimeter.
	Centimeters float32 = 37.8

	// Inches is the number of pixels per inch.
	Inches float32 = 96

	// TenthsOfInch is the number of pixels per tenth of an inch.
	TenthsOfInch float32 = 9.6

	// HundrethsOfInch is the number of pixels per hundredth of an inch.
	HundrethsOfInch float32 = 0.96
)

// PageUnits enumerates the measurement units used for page dimensions.
type PageUnits int

const (
	// PageUnitsMillimeters specifies measurements in millimeters.
	PageUnitsMillimeters PageUnits = iota
	// PageUnitsCentimeters specifies measurements in centimeters.
	PageUnitsCentimeters
	// PageUnitsInches specifies measurements in inches.
	PageUnitsInches
	// PageUnitsHundrethsOfInch specifies measurements in hundredths of an inch.
	PageUnitsHundrethsOfInch
)

// pixelsPerUnit returns the pixels-per-unit factor for the given PageUnits.
func pixelsPerUnit(u PageUnits) float32 {
	switch u {
	case PageUnitsMillimeters:
		return Millimeters
	case PageUnitsCentimeters:
		return Centimeters
	case PageUnitsInches:
		return Inches
	case PageUnitsHundrethsOfInch:
		return HundrethsOfInch
	default:
		return Millimeters
	}
}

// Convert converts value from one PageUnits to another.
//
//	pixels := Convert(25.4, PageUnitsMillimeters, PageUnitsInches) // ≈ 1.0
func Convert(value float32, from, to PageUnits) float32 {
	if from == to {
		return value
	}
	pixels := value * pixelsPerUnit(from)
	return pixels / pixelsPerUnit(to)
}

// ToPixels converts value in the given units to pixels.
func ToPixels(value float32, u PageUnits) float32 {
	return value * pixelsPerUnit(u)
}

// FromPixels converts a pixel value to the given units.
func FromPixels(pixels float32, u PageUnits) float32 {
	return pixels / pixelsPerUnit(u)
}
