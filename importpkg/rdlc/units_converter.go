// Package rdlc provides an importer for Microsoft RDL/RDLC report definitions.
//
// It is the Go equivalent of FastReport.Import.RDL (original-dotnet/
// FastReport.Base/Import/RDL/RDLImport.cs, UnitsConverter.cs, SizeUnits.cs,
// and ImportTable.cs).
//
// Usage:
//
//	rpt := reportpkg.NewReport()
//	imp := rdlc.New()
//	if err := imp.LoadReportFromFile(rpt, "report.rdlc"); err != nil {
//	    log.Fatal(err)
//	}
package rdlc

import (
	"strconv"
	"strings"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
	"image/color"
)

// ── Size unit string constants ────────────────────────────────────────────────
//
// C# ref: original-dotnet/FastReport.Base/Import/RDL/SizeUnits.cs

const (
	sizeUnitMM   = "mm"
	sizeUnitCM   = "cm"
	sizeUnitIn   = "in"
	sizeUnitPt   = "pt"
	sizeUnitPica = "pc"
)

// Millimeter conversion factors (mm per unit).
// C# ref: SizeUnitsM constants.
const (
	mmPerCM   = float32(10)
	mmPerIn   = float32(25.4)
	mmPerPt   = float32(0.3528)
	mmPerPica = float32(4.2336)
)

// Pixels per unit — kept here so the converter is self-contained.
// C# ref: SizeUnitsP — references FastReport.Utils.Units.*
const (
	pxPerMM   = float32(3.78)   // units.Millimeters
	pxPerCM   = float32(37.8)   // units.Centimeters
	pxPerIn   = float32(96)     // units.Inches
	pxPerPt   = pxPerMM * mmPerPt
	pxPerPica = pxPerMM * mmPerPica
)

// ── Helper: parse a raw float from an RDL size string ────────────────────────

// sizeToFloat strips the given unit suffix and parses the remaining float.
// C# ref: UnitsConverter.SizeToFloat
func sizeToFloat(s, unit string) float32 {
	trimmed := strings.TrimSpace(strings.Replace(s, unit, "", 1))
	f, err := strconv.ParseFloat(trimmed, 32)
	if err != nil {
		return 0
	}
	return float32(f)
}

// ── Public converters ─────────────────────────────────────────────────────────

// sizeToPixels converts an RDL size string (e.g. "2.54cm", "1in", "12pt")
// to pixels.
// C# ref: UnitsConverter.SizeToPixels
func sizeToPixels(s string) float32 {
	s = strings.TrimSpace(s)
	switch {
	case strings.Contains(s, sizeUnitMM):
		return sizeToFloat(s, sizeUnitMM) * pxPerMM
	case strings.Contains(s, sizeUnitCM):
		return sizeToFloat(s, sizeUnitCM) * pxPerCM
	case strings.Contains(s, sizeUnitIn):
		return sizeToFloat(s, sizeUnitIn) * pxPerIn
	case strings.Contains(s, sizeUnitPt):
		return sizeToFloat(s, sizeUnitPt) * pxPerPt
	case strings.Contains(s, sizeUnitPica):
		return sizeToFloat(s, sizeUnitPica) * pxPerPica
	}
	return 0
}

// sizeToMillimeters converts an RDL size string to millimeters.
// C# ref: UnitsConverter.SizeToMillimeters
func sizeToMillimeters(s string) float32 {
	s = strings.TrimSpace(s)
	switch {
	case strings.Contains(s, sizeUnitMM):
		return sizeToFloat(s, sizeUnitMM)
	case strings.Contains(s, sizeUnitCM):
		return sizeToFloat(s, sizeUnitCM) * mmPerCM
	case strings.Contains(s, sizeUnitIn):
		return sizeToFloat(s, sizeUnitIn) * mmPerIn
	case strings.Contains(s, sizeUnitPt):
		return sizeToFloat(s, sizeUnitPt) * mmPerPt
	case strings.Contains(s, sizeUnitPica):
		return sizeToFloat(s, sizeUnitPica) * mmPerPica
	}
	return 0
}

// sizeToInt converts an RDL size string to an integer, stripping the given
// unit and truncating.
// C# ref: UnitsConverter.SizeToInt
func sizeToInt(s, unit string) int {
	trimmed := strings.TrimSpace(strings.Replace(s, unit, "", 1))
	n, err := strconv.ParseFloat(trimmed, 32)
	if err != nil {
		return 0
	}
	return int(n)
}

// booleanToBool converts the RDL "true"/"false" string to bool (case-insensitive).
// C# ref: UnitsConverter.BooleanToBool
func booleanToBool(s string) bool {
	return strings.EqualFold(s, "true")
}

// convertColor converts an RDL color name or hex string to color.RGBA.
// C# ref: UnitsConverter.ConvertColor — calls Color.FromName which supports
// named HTML colors (e.g. "Red") as well as hex strings (e.g. "#FF0000").
// We delegate to utils.ParseColor which handles both forms.
func convertColor(s string) color.RGBA {
	c, err := utils.ParseColor(s)
	if err != nil {
		return color.RGBA{A: 255}
	}
	return c
}

// convertFontStyle converts an RDL FontStyle string to style.FontStyle.
// RDL only uses "Italic" for the FontStyle element; "Bold" is a separate
// FontWeight element. This function maps that RDL convention.
// C# ref: UnitsConverter.ConvertFontStyle
func convertFontStyle(s string) style.FontStyle {
	if s == "Italic" {
		return style.FontStyleItalic
	}
	return style.FontStyleRegular
}

// convertFontSize parses an RDL FontSize string (e.g. "12pt") to float32.
// C# ref: UnitsConverter.ConvertFontSize
func convertFontSize(s string) float32 {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, sizeUnitPt, "", 1)
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 32)
	if err != nil {
		return 10
	}
	return float32(f)
}

// convertTextAlign converts an RDL TextAlign value to object.HorzAlign.
// C# ref: UnitsConverter.ConvertTextAlign
func convertTextAlign(s string) object.HorzAlign {
	switch s {
	case "Center":
		return object.HorzAlignCenter
	case "Right":
		return object.HorzAlignRight
	}
	return object.HorzAlignLeft
}

// convertVerticalAlign converts an RDL VerticalAlign value to object.VertAlign.
// C# ref: UnitsConverter.ConvertVerticalAlign
func convertVerticalAlign(s string) object.VertAlign {
	switch s {
	case "Middle":
		return object.VertAlignCenter
	case "Bottom":
		return object.VertAlignBottom
	}
	return object.VertAlignTop
}

// convertWritingMode converts an RDL WritingMode string to a rotation angle in
// degrees.
// C# ref: UnitsConverter.ConvertWritingMode
func convertWritingMode(s string) int {
	if s == "tb-rl" {
		return 90
	}
	return 0
}

// convertBorderStyle converts an RDL BorderStyle string to style.LineStyle.
// C# ref: UnitsConverter.ConvertBorderStyle
func convertBorderStyle(s string) style.LineStyle {
	switch s {
	case "Dotted":
		return style.LineStyleDot
	case "Dashed":
		return style.LineStyleDash
	case "Double":
		return style.LineStyleDouble
	}
	return style.LineStyleSolid
}

// convertSizing converts an RDL Sizing string to object.SizeMode.
// C# ref: UnitsConverter.ConvertSizing
//
// RDL Sizing → PictureBoxSizeMode → object.SizeMode mapping:
//   "AutoSize" → AutoSize → SizeModeAutoSize
//   "Fit"      → StretchImage → SizeModeStretchImage
//   "Clip"     → Normal → SizeModeNormal
//   default    → Zoom → SizeModeZoom
func convertSizing(s string) object.SizeMode {
	switch s {
	case "AutoSize":
		return object.SizeModeAutoSize
	case "Fit":
		return object.SizeModeStretchImage
	case "Clip":
		return object.SizeModeNormal
	}
	return object.SizeModeZoom
}
