// render.go implements 1D (linear) barcode rendering.
//
// Ported from C# LinearBarcodeBase.cs — DoLines + DrawBarcode.
// The pattern string encodes each bar as a single character:
//
//	'0'-'3' = white space (modules[0..3] width)
//	'5'-'8' = black bar   (modules[0..3] width)
//	'9'     = half-height black bar (PostNet short bar)
//	'A'-'D' = long black bar, extends 7 units into text area (EAN/UPC guard)
//	'E'     = tracker bar — middle 1/3 height (Intelligent Mail)
//	'F'     = ascender bar — top 2/3 height (Intelligent Mail)
//	'G'     = descender bar — bottom 1/3→bottom (Intelligent Mail)
package barcode

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// BarLineType classifies the visual appearance of a bar in a linear barcode pattern.
type BarLineType int

const (
	BarLineWhite          BarLineType = iota // white space
	BarLineBlack                             // full-height black bar
	BarLineBlackHalf                         // 2/5 height from bottom (PostNet)
	BarLineBlackLong                         // full height + 7 units into text area (EAN/UPC guard)
	BarLineBlackTracker                      // middle 1/3 (Intelligent Mail)
	BarLineBlackAscender                     // top 2/3 (Intelligent Mail)
	BarLineBlackDescender                    // bottom 1/3→bottom (Intelligent Mail)
)

// PatternProvider is implemented by linear barcode types that generate a
// pattern string for rendering by DrawLinearBarcode.
type PatternProvider interface {
	GetPattern() string
	GetWideBarRatio() float32
}

// MakeModules computes the 4-element module-width array from wideBarRatio.
// Ported from C# LinearBarcodeBase.MakeModules().
//
//	modules[0] = 1            (narrow)
//	modules[1] = wideBarRatio (wide)
//	modules[2] = wideBarRatio * 1.5
//	modules[3] = wideBarRatio * 2
func MakeModules(wideBarRatio float32) [4]float32 {
	var m [4]float32
	m[0] = 1
	m[1] = m[0] * wideBarRatio
	m[2] = m[1] * 1.5
	m[3] = m[1] * 2
	return m
}

// OneBarProps returns the pixel width and line type for a single pattern
// character. Ported from C# LinearBarcodeBase.OneBarProps().
func OneBarProps(code byte, modules [4]float32) (width float32, lt BarLineType, err error) {
	switch code {
	case '0':
		return modules[0], BarLineWhite, nil
	case '1':
		return modules[1], BarLineWhite, nil
	case '2':
		return modules[2], BarLineWhite, nil
	case '3':
		return modules[3], BarLineWhite, nil
	case '5':
		return modules[0], BarLineBlack, nil
	case '6':
		return modules[1], BarLineBlack, nil
	case '7':
		return modules[2], BarLineBlack, nil
	case '8':
		return modules[3], BarLineBlack, nil
	case '9':
		return modules[0], BarLineBlackHalf, nil
	case 'A':
		return modules[0], BarLineBlackLong, nil
	case 'B':
		return modules[1], BarLineBlackLong, nil
	case 'C':
		return modules[2], BarLineBlackLong, nil
	case 'D':
		return modules[3], BarLineBlackLong, nil
	case 'E':
		return modules[1], BarLineBlackTracker, nil
	case 'F':
		return modules[1], BarLineBlackAscender, nil
	case 'G':
		return modules[1], BarLineBlackDescender, nil
	default:
		return 0, BarLineWhite, fmt.Errorf("barcode: unknown pattern code %q", code)
	}
}

// GetPatternWidth returns the total width in module units for a pattern string.
func GetPatternWidth(pattern string, modules [4]float32) float32 {
	var total float32
	for i := range len(pattern) {
		w, _, err := OneBarProps(pattern[i], modules)
		if err == nil {
			total += w
		}
	}
	return total
}

// CheckSumModulo10 computes a mod-10 check digit and appends it to data.
// Ported from C# LinearBarcodeBase.CheckSumModulo10().
func CheckSumModulo10(data string) string {
	sum := 0
	fak := len(data)
	for i := range len(data) {
		digit := int(data[i] - '0')
		if fak%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
		fak--
	}
	if sum%10 == 0 {
		return data + "0"
	}
	return data + fmt.Sprintf("%d", 10-(sum%10))
}

// MakeLong converts black bars ('5'-'8') to long bars ('A'-'D') so they extend
// into the text area. Used for EAN/UPC guard bars.
// Ported from C# LinearBarcodeBase.MakeLong().
func MakeLong(pattern string) string {
	b := make([]byte, len(pattern))
	for i := range len(pattern) {
		c := pattern[i]
		if c >= '5' && c <= '8' {
			c = c - '5' + 'A'
		}
		b[i] = c
	}
	return string(b)
}

// DrawLinearBarcode renders a linear barcode pattern to an image.
// Ported from C# LinearBarcodeBase.DrawBarcode + DoLines.
func DrawLinearBarcode(pattern, text string, width, height int, showText bool, wideBarRatio float32) image.Image {
	if width <= 0 || height <= 0 || len(pattern) == 0 {
		return image.NewRGBA(image.Rect(0, 0, max(width, 1), max(height, 1)))
	}
	if wideBarRatio <= 0 {
		wideBarRatio = 2.0
	}

	modules := MakeModules(wideBarRatio)
	originalWidth := GetPatternWidth(pattern, modules)
	if originalWidth <= 0 {
		return image.NewRGBA(image.Rect(0, 0, width, height))
	}

	// C# LinearBarcodeBase.DrawBarcode():
	//   originalWidth = CalcBounds().Width / 1.25  (= barWidth + extra1 + extra2)
	//   zoom = displayWidth / originalWidth
	//
	// CalcBounds().Width = (barWidth + extras) * 1.25. When the engine renders at
	// w*3 (3× resolution), displayWidth ≈ CalcBounds().Width * 3.
	//
	// In Go, CalcBounds already bakes the 1.25 factor into the object Width, and
	// the engine passes that width * 3 as our `width` parameter. So:
	//   width ≈ (barWidth + extras) * 1.25 * 3
	// C# equivalent originalWidth = (barWidth + extras)
	// C# zoom = width / (barWidth + extras)
	//
	// We don't know extras here, so we use barWidth directly. The bars will fill
	// the full image width — identical to C# for barcodes without extras (Code128,
	// Code39, etc.) and slightly stretched for EAN/UPC (where extras exist as quiet
	// zones). This matches C#'s bar-filling behavior.
	zoom := float32(width) / originalWidth

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	// Bar area height (in pattern units).
	barAreaH := float32(height) / zoom
	const fontHeight float32 = 14 // C# default FontHeight ~14px at 8pt Arial
	if showText && text != "" {
		barAreaH -= fontHeight
		if barAreaH < 1 {
			barAreaH = 1
		}
	}

	// Position bars left-aligned (matching C# barArea.Left * zoom = 0 for most types).
	// C# offsets by extra1 * zoom for EAN/UPC quiet zones, but we fill the full width
	// since CalcBounds already sized the object to include quiet zones.
	var leftOff float32

	black := image.NewUniform(color.Black)
	curX := leftOff

	for i := range len(pattern) {
		barW, lt, err := OneBarProps(pattern[i], modules)
		if err != nil {
			continue
		}
		scaledW := barW * zoom

		var y0, y1 float32
		switch lt {
		case BarLineWhite:
			curX += scaledW
			continue
		case BarLineBlack:
			y0, y1 = 0, barAreaH*zoom
		case BarLineBlackHalf:
			end := barAreaH * zoom
			shortH := end * 2 / 5
			y0, y1 = end-shortH, end
		case BarLineBlackLong:
			y0 = 0
			y1 = barAreaH * zoom
			if showText {
				y1 += 7 * zoom
			}
		case BarLineBlackTracker:
			y0 = barAreaH * zoom * 1 / 3
			y1 = barAreaH * zoom * 2 / 3
		case BarLineBlackAscender:
			y0 = 0
			y1 = barAreaH * zoom * 2 / 3
		case BarLineBlackDescender:
			y0 = barAreaH * zoom * 1 / 3
			y1 = barAreaH * zoom
		}

		px0 := int(math.Round(float64(curX)))
		px1 := int(math.Round(float64(curX + scaledW)))
		py0 := int(math.Round(float64(y0)))
		py1 := int(math.Round(float64(y1)))

		if px1 <= px0 {
			px1 = px0 + 1
		}
		if py1 <= py0 {
			py1 = py0 + 1
		}
		if px1 > width {
			px1 = width
		}
		if py1 > height {
			py1 = height
		}

		draw.Draw(img, image.Rect(px0, py0, px1, py1), black, image.Point{}, draw.Src)
		curX += scaledW
	}

	if showText && text != "" {
		textTop := int(math.Round(float64(barAreaH * zoom)))
		textH := height - textTop
		if textH > 0 {
			drawLinearText(img, text, 0, textTop, width, textH)
		}
	}
	return img
}

// drawLinearText draws centred text below the bars, matching C# DrawString behaviour.
func drawLinearText(img *image.RGBA, text string, x0, y0, areaW, areaH int) {
	face := basicfont.Face7x13
	metrics := face.Metrics()
	advance := font.MeasureString(face, text)
	textW := advance.Ceil()
	tx := x0 + (areaW-textW)/2
	if tx < x0 {
		tx = x0
	}
	ascent := metrics.Ascent.Ceil()
	descent := metrics.Descent.Ceil()
	ty := y0 + (areaH+ascent-descent)/2
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: face,
		Dot:  fixed.P(tx, ty),
	}
	d.DrawString(text)
}

// ── Custom text drawing for EAN/UPC ──────────────────────────────────────────

// CustomTextDrawFunc is a callback for barcode types that need custom text
// positioning (e.g. EAN-8, EAN-13, UPC-A). It receives the image, the display
// text, the text-area rectangle (y0, height), and the zoom factor + modules
// so it can compute per-group positions from the pattern.
type CustomTextDrawFunc func(img *image.RGBA, text string, textTop, textH int, zoom float32, modules [4]float32)

// DrawLinearBarcodeCustomText renders a linear barcode with a custom text
// drawing function. The customDraw callback replaces the default centered text.
// Ported from C# LinearBarcodeBase.DrawBarcode + per-type DrawText overrides.
func DrawLinearBarcodeCustomText(pattern, text string, width, height int, showText bool, wideBarRatio float32, customDraw CustomTextDrawFunc) image.Image {
	if width <= 0 || height <= 0 || len(pattern) == 0 {
		return image.NewRGBA(image.Rect(0, 0, max(width, 1), max(height, 1)))
	}
	if wideBarRatio <= 0 {
		wideBarRatio = 2.0
	}

	modules := MakeModules(wideBarRatio)
	originalWidth := GetPatternWidth(pattern, modules)
	if originalWidth <= 0 {
		return image.NewRGBA(image.Rect(0, 0, width, height))
	}

	zoom := float32(width) / originalWidth

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	barAreaH := float32(height) / zoom
	const fontHeight float32 = 14
	if showText && text != "" {
		barAreaH -= fontHeight
		if barAreaH < 1 {
			barAreaH = 1
		}
	}

	var leftOff float32
	black := image.NewUniform(color.Black)
	curX := leftOff

	for i := range len(pattern) {
		barW, lt, err := OneBarProps(pattern[i], modules)
		if err != nil {
			continue
		}
		scaledW := barW * zoom

		var y0, y1 float32
		switch lt {
		case BarLineWhite:
			curX += scaledW
			continue
		case BarLineBlack:
			y0, y1 = 0, barAreaH*zoom
		case BarLineBlackHalf:
			end := barAreaH * zoom
			shortH := end * 2 / 5
			y0, y1 = end-shortH, end
		case BarLineBlackLong:
			y0 = 0
			y1 = barAreaH * zoom
			if showText {
				y1 += 7 * zoom
			}
		case BarLineBlackTracker:
			y0 = barAreaH * zoom * 1 / 3
			y1 = barAreaH * zoom * 2 / 3
		case BarLineBlackAscender:
			y0 = 0
			y1 = barAreaH * zoom * 2 / 3
		case BarLineBlackDescender:
			y0 = barAreaH * zoom * 1 / 3
			y1 = barAreaH * zoom
		}

		px0 := int(math.Round(float64(curX)))
		px1 := int(math.Round(float64(curX + scaledW)))
		py0 := int(math.Round(float64(y0)))
		py1 := int(math.Round(float64(y1)))

		if px1 <= px0 {
			px1 = px0 + 1
		}
		if py1 <= py0 {
			py1 = py0 + 1
		}
		if px1 > width {
			px1 = width
		}
		if py1 > height {
			py1 = height
		}

		draw.Draw(img, image.Rect(px0, py0, px1, py1), black, image.Point{}, draw.Src)
		curX += scaledW
	}

	if showText && text != "" {
		textTop := int(math.Round(float64(barAreaH * zoom)))
		textH := height - textTop
		if textH > 0 {
			if customDraw != nil {
				customDraw(img, text, textTop, textH, zoom, modules)
			} else {
				drawLinearText(img, text, 0, textTop, width, textH)
			}
		}
	}
	return img
}

// drawStringInRange draws text centered between pixel positions x1 and x2 in
// the text area. This mirrors C# LinearBarcodeBase.DrawString(g, x1, x2, s).
// x1 and x2 are in pattern-unit coordinates (pre-zoom).
func drawStringInRange(img *image.RGBA, s string, x1px, x2px, textTop, textH int) {
	if len(s) == 0 {
		return
	}
	face := basicfont.Face7x13
	metrics := face.Metrics()
	advance := font.MeasureString(face, s)
	textW := advance.Ceil()

	// Center text between x1px and x2px.
	tx := x1px + (x2px-x1px-textW)/2
	ascent := metrics.Ascent.Ceil()
	descent := metrics.Descent.Ceil()
	ty := textTop + (textH+ascent-descent)/2

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: face,
		Dot:  fixed.P(tx, ty),
	}
	d.DrawString(s)
}

// getWidthPx returns the pixel width of a pattern substring at the given zoom.
// This mirrors C# LinearBarcodeBase.GetWidth(pattern.Substring(start, len)).
func getWidthPx(pattern string, start, length int, zoom float32, modules [4]float32) int {
	if start < 0 || start+length > len(pattern) {
		return 0
	}
	w := GetPatternWidth(pattern[start:start+length], modules)
	return int(math.Round(float64(w * zoom)))
}

// ── EAN-8 custom text drawing ────────────────────────────────────────────────

// EAN8DrawText draws the EAN-8 human-readable text split into two 4-digit
// groups positioned under the left and right halves of the barcode.
// Ported from C# BarcodeEAN8.DrawText (BarcodeEAN.cs:76-86).
//
// EAN-8 pattern structure: 3 + 16 + 5 + 16 + 3 = 43 characters
//   - "A0A"       (3) — start guard
//   - 4×4=16      (16) — left 4 digits (charset A)
//   - "0A0A0"     (5) — centre guard
//   - 4×4=16      (16) — right 4 digits (charset C)
//   - "A0A"       (3) — stop guard
func EAN8DrawText(pattern string) CustomTextDrawFunc {
	return func(img *image.RGBA, text string, textTop, textH int, zoom float32, modules [4]float32) {
		// Ensure we have the 8-digit display text.
		barData := eanSetLen(text, 7)
		barData = eanChecksum(barData) // now 8 digits

		// Left group: digits 0-3, positioned between start guard and centre guard.
		// C#: x1 = GetWidth(pattern[0:3]), x2 = GetWidth(pattern[0:3+16+1])
		x1 := getWidthPx(pattern, 0, 3, zoom, modules)
		x2 := getWidthPx(pattern, 0, 3+16+1, zoom, modules)
		drawStringInRange(img, barData[0:4], x1, x2, textTop, textH)

		// Right group: digits 4-7, positioned between centre guard and stop guard.
		// C#: x1 = GetWidth(pattern[0:3+16+5-1]), x2 = GetWidth(pattern[0:3+16+5+16])
		x1 = getWidthPx(pattern, 0, 3+16+5-1, zoom, modules)
		x2 = getWidthPx(pattern, 0, 3+16+5+16, zoom, modules)
		drawStringInRange(img, barData[4:8], x1, x2, textTop, textH)
	}
}

// ── EAN-13 custom text drawing ───────────────────────────────────────────────

// EAN13DrawText draws the EAN-13 human-readable text: first digit outside
// the left guard, then two 6-digit groups under the left and right halves.
// Ported from C# BarcodeEAN13.DrawText (BarcodeEAN.cs:130-142).
//
// EAN-13 pattern structure: 3 + 24 + 5 + 24 + 3 = 59 characters
//   - "A0A"       (3) — start guard
//   - 6×4=24      (24) — left 6 digits (charset A/B per parity)
//   - "0A0A0"     (5) — centre guard
//   - 6×4=24      (24) — right 6 digits (charset C)
//   - "A0A"       (3) — stop guard
func EAN13DrawText(pattern string) CustomTextDrawFunc {
	return func(img *image.RGBA, text string, textTop, textH int, zoom float32, modules [4]float32) {
		// Ensure we have the 13-digit display text.
		barData := eanSetLen(text, 12)
		barData = eanChecksum(barData) // now 13 digits

		// First digit: drawn outside the left start guard.
		// C#: DrawString(g, -8, -2, barData.Substring(0, 1))
		x1 := int(math.Round(float64(-8 * zoom)))
		x2 := int(math.Round(float64(-2 * zoom)))
		drawStringInRange(img, barData[0:1], x1, x2, textTop, textH)

		// Left group: digits 1-6, positioned between start guard and centre guard.
		// C#: x1 = GetWidth(pattern[0:3]), x2 = GetWidth(pattern[0:3+24+1])
		x1 = getWidthPx(pattern, 0, 3, zoom, modules)
		x2 = getWidthPx(pattern, 0, 3+24+1, zoom, modules)
		drawStringInRange(img, barData[1:7], x1, x2, textTop, textH)

		// Right group: digits 7-12, positioned between centre guard and stop guard.
		// C#: x1 = GetWidth(pattern[0:3+24+5-1]), x2 = GetWidth(pattern[0:3+24+5+24])
		x1 = getWidthPx(pattern, 0, 3+24+5-1, zoom, modules)
		x2 = getWidthPx(pattern, 0, 3+24+5+24, zoom, modules)
		drawStringInRange(img, barData[7:13], x1, x2, textTop, textH)
	}
}

// ── UPC-A custom text drawing ─────────────────────────────────────────────────

// UPCADrawText draws the UPC-A human-readable text with the first digit
// outside the left guard, two 5-digit groups under the left and right halves,
// and the check digit outside the right guard.
// Ported from C# BarcodeUPC_A.DrawText (BarcodeUPC.cs:109-125).
//
// UPC-A pattern structure: 3 + 4(long) + 20 + 5 + 20 + 4(long) + 3 = 59 characters
//
//	"A0A"          (3) — start guard
//	1×4=4 (long)   (4) — first digit (long bars via MakeLong)
//	5×4=20         (20) — left 5 digits
//	"0A0A0"        (5) — centre guard
//	5×4=20         (20) — right 5 digits
//	1×4=4 (long)   (4) — last digit (long bars via MakeLong)
//	"A0A"          (3) — stop guard
func UPCADrawText(pattern string) CustomTextDrawFunc {
	return func(img *image.RGBA, text string, textTop, textH int, zoom float32, modules [4]float32) {
		// Compute the 12-digit display text (11 data + 1 check).
		barData := eanSetLen(text, 11)
		barData = CheckSumModulo10(barData) // now 12 digits

		// First digit: drawn outside the left start guard in the quiet zone.
		// C#: DrawString(g, -8, -2, barData.Substring(0, 1), true)
		x1 := int(math.Round(float64(-8 * zoom)))
		x2 := int(math.Round(float64(-2 * zoom)))
		drawStringInRange(img, barData[0:1], x1, x2, textTop, textH)

		// Left group: digits 1-5, positioned between first long digit and centre guard.
		// C# pattern parts: 7 + 20 + 5 + 20 + 7
		// C#: x1 = GetWidth(pattern[0:7]), x2 = GetWidth(pattern[0:7+20])
		x1 = getWidthPx(pattern, 0, 7, zoom, modules)
		x2 = getWidthPx(pattern, 0, 7+20, zoom, modules)
		drawStringInRange(img, barData[1:6], x1, x2, textTop, textH)

		// Right group: digits 6-10, positioned between centre guard and last long digit.
		// C#: x1 = GetWidth(pattern[0:7+20+5]), x2 = GetWidth(pattern[0:7+20+5+20])
		x1 = getWidthPx(pattern, 0, 7+20+5, zoom, modules)
		x2 = getWidthPx(pattern, 0, 7+20+5+20, zoom, modules)
		drawStringInRange(img, barData[6:11], x1, x2, textTop, textH)

		// Check digit: drawn outside the right stop guard.
		// C#: x1 = GetWidth(pattern) + 1, x2 = x1 + 7
		patW := GetPatternWidth(pattern, modules)
		x1 = int(math.Round(float64((patW + 1) * zoom)))
		x2 = int(math.Round(float64((patW + 8) * zoom)))
		drawStringInRange(img, barData[11:12], x1, x2, textTop, textH)
	}
}

// ── UPC-E0 custom text drawing ────────────────────────────────────────────────

// UPCE0DrawText draws the UPC-E0 human-readable text with "0" outside the
// left guard, the 6 data digits under the bars, and the check digit outside
// the right guard.
// Ported from C# BarcodeUPC_E0.DrawText (BarcodeUPC.cs:27-39).
//
// UPC-E0 pattern structure: 3 + 24 + 6 = 33 characters
//
//	"A0A"          (3) — start guard
//	6×4=24         (24) — 6 data digits
//	"0A0A0A"       (6) — stop guard
func UPCE0DrawText(pattern string) CustomTextDrawFunc {
	return func(img *image.RGBA, text string, textTop, textH int, zoom float32, modules [4]float32) {
		// Compute the 7-digit display text (6 data + 1 check).
		barData := eanSetLen(text, 6)
		barData = CheckSumModulo10(barData) // now 7 digits

		// "0" digit: drawn outside the left start guard.
		// C#: DrawString(g, -8, -2, "0", true)
		x1 := int(math.Round(float64(-8 * zoom)))
		x2 := int(math.Round(float64(-2 * zoom)))
		drawStringInRange(img, "0", x1, x2, textTop, textH)

		// Data digits: all 6 digits under the bars between start and stop guard.
		// C# pattern parts: 3 + 24 + 6
		// C#: x1 = GetWidth(pattern[0:3]), x2 = GetWidth(pattern[0:3+24])
		x1 = getWidthPx(pattern, 0, 3, zoom, modules)
		x2 = getWidthPx(pattern, 0, 3+24, zoom, modules)
		drawStringInRange(img, barData[0:6], x1, x2, textTop, textH)

		// Check digit: drawn outside the right stop guard.
		// C#: x1 = GetWidth(pattern) + 1, x2 = x1 + 7
		patW := GetPatternWidth(pattern, modules)
		x1 = int(math.Round(float64((patW + 1) * zoom)))
		x2 = int(math.Round(float64((patW + 8) * zoom)))
		drawStringInRange(img, barData[6:7], x1, x2, textTop, textH)
	}
}

// ── ITF-14 bearer bar rendering ───────────────────────────────────────────────

// ITF14FormatDisplayText formats a 14-digit ITF-14 string with spaces for
// human-readable display below the barcode.
// Ported from C# BarcodeITF14.DrawText (Barcode2of5.cs:401-403):
//
//	data.Insert(1, " ").Insert(4, " ").Insert(10, " ").Insert(16, " ")
//
// Example: "12345678901231" → "1 23 45678 90123 1"
// Cumulative-insert trace on 14-char input:
//
//	Insert at 1:  groups from orig [0:1], [1:14]     → orig[0] + " " + …
//	Insert at 4:  groups from orig [0:1], [1:3], [3:14]
//	Insert at 10: groups from orig [0:1], [1:3], [3:8], [8:14]
//	Insert at 16: groups from orig [0:1], [1:3], [3:8], [8:13], [13:14]
func ITF14FormatDisplayText(text string) string {
	if len(text) < 14 {
		return text
	}
	// Original digit groups (C# trace, positions are on the running string):
	//   Insert at 1:  split after orig index 0       → group sizes: 1
	//   Insert at 4:  split after orig index 2       → group sizes: 1, 2
	//   Insert at 10: split after orig index 7       → group sizes: 1, 2, 5
	//   Insert at 16: split after orig index 12      → group sizes: 1, 2, 5, 5, 1
	return text[0:1] + " " + text[1:3] + " " + text[3:8] + " " + text[8:13] + " " + text[13:14]
}

// DrawLinearBarcodeITF14 renders an ITF-14 barcode with bearer bars.
// It calls DrawLinearBarcodeCustomText for the bars and text, then overlays
// the horizontal bearer bars (always drawn) and optional vertical bearer bars.
//
// Ported from C# BarcodeITF14.DrawBarcode (Barcode2of5.cs:427-476):
//   - Bearer bar thickness = WideBarRatio * 2 * zoom
//   - Horizontal bars: top and bottom, spanning the full barArea width
//   - Vertical bars: left and right, spanning the full barArea height
//     (only when drawVerticalBearerBars is true)
//
// The barArea in C# is the bar-only region (excludes text area). In our
// renderer the bar area occupies the top portion of the image; the text
// area (if any) is below it.
func DrawLinearBarcodeITF14(pattern, displayText string, width, height int,
	showText bool, wideBarRatio float32, drawVertBearerBars bool) image.Image {

	img := DrawLinearBarcodeCustomText(
		pattern, displayText, width, height, showText, wideBarRatio, nil,
	)

	rgba, ok := img.(*image.RGBA)
	if !ok {
		// Convert to RGBA if the renderer returned a different type.
		bounds := img.Bounds()
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	}

	if width <= 0 || height <= 0 || len(pattern) == 0 {
		return rgba
	}
	if wideBarRatio <= 0 {
		wideBarRatio = 2.0
	}

	modules := MakeModules(wideBarRatio)
	originalWidth := GetPatternWidth(pattern, modules)
	if originalWidth <= 0 {
		return rgba
	}
	zoom := float32(width) / originalWidth

	// Bearer bar thickness: WideBarRatio * 2 * zoom.
	// C# Barcode2of5.cs:450: float bearerWidth = WideBarRatio * 2 * zoom
	bearerThickness := int(math.Round(float64(wideBarRatio * 2 * zoom)))
	if bearerThickness < 1 {
		bearerThickness = 1
	}

	// Bar area height (pixels): same calculation as DrawLinearBarcode.
	const fontHeight float32 = 14
	barAreaH := float32(height) / zoom
	if showText && displayText != "" {
		barAreaH -= fontHeight
		if barAreaH < 1 {
			barAreaH = 1
		}
	}
	barAreaPx := int(math.Round(float64(barAreaH * zoom)))
	if barAreaPx > height {
		barAreaPx = height
	}

	black := color.Black

	// Draw horizontal bearer bars: top edge and bottom edge of the bar area.
	// C# Barcode2of5.cs:462-463:
	//   g.DrawLine(pen, x0, y01-0.5F, x1, y01-0.5F)   // top
	//   g.DrawLine(pen, x0, y11, x1, y11)               // bottom
	// y01 = bearerWidth/2, y11 = barArea.Bottom*zoom - bearerWidth/2
	// So top bar centre = bearerThickness/2, bottom bar centre = barAreaPx - bearerThickness/2.

	// Top bearer bar.
	topY0 := 0
	topY1 := bearerThickness
	if topY1 > barAreaPx {
		topY1 = barAreaPx
	}
	for y := topY0; y < topY1; y++ {
		for x := 0; x < width; x++ {
			rgba.Set(x, y, black)
		}
	}

	// Bottom bearer bar.
	botY1 := barAreaPx
	botY0 := barAreaPx - bearerThickness
	if botY0 < 0 {
		botY0 = 0
	}
	for y := botY0; y < botY1; y++ {
		for x := 0; x < width; x++ {
			rgba.Set(x, y, black)
		}
	}

	// Vertical bearer bars (optional).
	// C# Barcode2of5.cs:464-468:
	//   if (this.drawVerticalBearerBars)
	//   { g.DrawLine(pen, x01-0.5F, y0, x01-0.5F, y1)   // left
	//     g.DrawLine(pen, x11, y0, x11, y1) }             // right
	if drawVertBearerBars {
		// Left vertical bar.
		leftX1 := bearerThickness
		if leftX1 > width {
			leftX1 = width
		}
		for y := 0; y < barAreaPx; y++ {
			for x := 0; x < leftX1; x++ {
				rgba.Set(x, y, black)
			}
		}

		// Right vertical bar.
		rightX0 := width - bearerThickness
		if rightX0 < 0 {
			rightX0 = 0
		}
		for y := 0; y < barAreaPx; y++ {
			for x := rightX0; x < width; x++ {
				rgba.Set(x, y, black)
			}
		}
	}

	return rgba
}
