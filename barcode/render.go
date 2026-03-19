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

	// zoom = targetWidth / (originalWidth * 1.25) — matches C#'s 1.25 padding factor.
	zoom := float32(width) / (originalWidth * 1.25)

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

	// Centre the barcode horizontally.
	totalW := originalWidth * zoom
	leftOff := (float32(width) - totalW) / 2
	if leftOff < 0 {
		leftOff = 0
	}

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
