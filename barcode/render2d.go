// render2d.go implements 2D barcode rendering.
//
// Ported from C# Barcode2DBase.DrawBarcode.
// A 2D barcode is represented as a boolean module grid; each true cell is a
// dark module rendered as a filled rectangle.
package barcode

import (
	"image"
	"image/color"
	"image/draw"
)

// Matrix2DProvider is implemented by 2D barcode types that can supply a
// boolean module grid for rendering by DrawBarcode2D.
type Matrix2DProvider interface {
	// GetMatrix returns (matrix[row][col], rows, cols).
	// matrix[r][c] == true means dark module at row r, column c.
	GetMatrix() (matrix [][]bool, rows, cols int)
}

// DrawBarcode2D renders a boolean module matrix to an image.
// Ported from C# Barcode2DBase.DrawBarcode which iterates over the module
// grid and fills each dark cell with a filled rectangle.
func DrawBarcode2D(matrix [][]bool, rows, cols, width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	if rows <= 0 || cols <= 0 || width <= 0 || height <= 0 {
		return img
	}

	black := image.NewUniform(color.Black)
	for r := range rows {
		if r >= len(matrix) {
			break
		}
		for c := range cols {
			if c >= len(matrix[r]) || !matrix[r][c] {
				continue
			}
			x0 := c * width / cols
			x1 := (c + 1) * width / cols
			y0 := r * height / rows
			y1 := (r + 1) * height / rows
			if x1 <= x0 {
				x1 = x0 + 1
			}
			if y1 <= y0 {
				y1 = y0 + 1
			}
			draw.Draw(img, image.Rect(x0, y0, x1, y1), black, image.Point{}, draw.Src)
		}
	}
	return img
}
