package barcode_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// TestDrawBarcode2D_3x3_ImageDimensions verifies that a small 3x3 matrix
// produces an image with the exact requested pixel dimensions and that
// dark/light modules map to black/white pixels.
func TestDrawBarcode2D_3x3_ImageDimensions(t *testing.T) {
	matrix := [][]bool{
		{true, false, true},
		{false, true, false},
		{true, false, true},
	}
	const w, h = 90, 90
	img := barcode.DrawBarcode2D(matrix, 3, 3, w, h)
	if img == nil {
		t.Fatal("DrawBarcode2D returned nil")
	}
	bounds := img.Bounds()
	if bounds.Dx() != w || bounds.Dy() != h {
		t.Fatalf("expected image %dx%d, got %dx%d", w, h, bounds.Dx(), bounds.Dy())
	}

	// Each module is 30x30 pixels (90/3). Verify centre pixel of each module.
	type check struct {
		row, col int
		dark     bool
	}
	checks := []check{
		{0, 0, true}, {0, 1, false}, {0, 2, true},
		{1, 0, false}, {1, 1, true}, {1, 2, false},
		{2, 0, true}, {2, 1, false}, {2, 2, true},
	}
	for _, c := range checks {
		// centre of module (row, col)
		px := c.col*30 + 15
		py := c.row*30 + 15
		r, g, b, a := img.At(px, py).RGBA()
		isBlack := r == 0 && g == 0 && b == 0 && a > 0
		isLight := a == 0 // transparent = light module
		if c.dark && !isBlack {
			t.Errorf("module (%d,%d) should be dark but pixel (%d,%d) is not black", c.row, c.col, px, py)
		}
		if !c.dark && !isLight {
			t.Errorf("module (%d,%d) should be light but pixel (%d,%d) is not transparent", c.row, c.col, px, py)
		}
	}
}

// TestDrawBarcode2D_21x21_QRLike verifies rendering a 21x21 matrix
// (the size of a QR version 1) into a 210x210 image.
func TestDrawBarcode2D_21x21_QRLike(t *testing.T) {
	const size = 21
	const imgSize = 210
	matrix := make([][]bool, size)
	for r := range matrix {
		matrix[r] = make([]bool, size)
		for c := range matrix[r] {
			// Checkerboard pattern
			matrix[r][c] = (r+c)%2 == 0
		}
	}
	img := barcode.DrawBarcode2D(matrix, size, size, imgSize, imgSize)
	if img == nil {
		t.Fatal("DrawBarcode2D returned nil")
	}
	bounds := img.Bounds()
	if bounds.Dx() != imgSize || bounds.Dy() != imgSize {
		t.Fatalf("expected %dx%d, got %dx%d", imgSize, imgSize, bounds.Dx(), bounds.Dy())
	}

	// Each module is 10x10 pixels. Spot-check a few modules.
	moduleSize := imgSize / size
	spotChecks := [][2]int{{0, 0}, {0, 1}, {10, 10}, {20, 20}, {5, 15}}
	for _, sc := range spotChecks {
		r, c := sc[0], sc[1]
		expectDark := (r+c)%2 == 0
		px := c*moduleSize + moduleSize/2
		py := r*moduleSize + moduleSize/2
		rv, gv, bv, av := img.At(px, py).RGBA()
		isBlack := rv == 0 && gv == 0 && bv == 0 && av > 0
		isLight := av == 0
		if expectDark && !isBlack {
			t.Errorf("module (%d,%d) expected dark, pixel (%d,%d) not black", r, c, px, py)
		}
		if !expectDark && !isLight {
			t.Errorf("module (%d,%d) expected light, pixel (%d,%d) is not transparent", r, c, px, py)
		}
	}
}

// TestDrawBarcode2D_NilMatrix verifies that passing a nil matrix
// returns a valid (all-transparent) image of the requested size when
// rows/cols are zero.
func TestDrawBarcode2D_NilMatrix(t *testing.T) {
	img := barcode.DrawBarcode2D(nil, 0, 0, 50, 50)
	if img == nil {
		t.Fatal("expected non-nil image for nil matrix")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Fatalf("expected 50x50, got %dx%d", bounds.Dx(), bounds.Dy())
	}
	// The image should be entirely transparent (no background fill).
	for y := 0; y < 50; y += 10 {
		for x := 0; x < 50; x += 10 {
			_, _, _, a := img.At(x, y).RGBA()
			if a != 0 {
				t.Fatalf("pixel (%d,%d) should be transparent, got alpha %d", x, y, a)
			}
		}
	}
}

// TestDrawBarcode2D_EmptyMatrix verifies that an empty (non-nil) matrix
// with positive rows/cols still returns a valid transparent image.
func TestDrawBarcode2D_EmptyMatrix(t *testing.T) {
	matrix := [][]bool{} // non-nil but empty
	img := barcode.DrawBarcode2D(matrix, 3, 3, 60, 60)
	if img == nil {
		t.Fatal("expected non-nil image for empty matrix")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 60 || bounds.Dy() != 60 {
		t.Fatalf("expected 60x60, got %dx%d", bounds.Dx(), bounds.Dy())
	}
	// Should be all-transparent since matrix has no actual rows to render.
	_, _, _, a := img.At(30, 30).RGBA()
	if a != 0 {
		t.Error("centre pixel should be transparent for empty matrix")
	}
}

// TestDrawBarcode2D_JaggedMatrix verifies that a jagged matrix
// (rows with varying column lengths) does not panic and renders correctly
// for the columns that exist.
func TestDrawBarcode2D_JaggedMatrix(t *testing.T) {
	matrix := [][]bool{
		{true, true, true, true},
		{true},                        // only 1 column, cols says 4
		{true, false, true},           // 3 columns
		{true, true, true, true},
	}
	img := barcode.DrawBarcode2D(matrix, 4, 4, 80, 80)
	if img == nil {
		t.Fatal("expected non-nil image for jagged matrix")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 80 || bounds.Dy() != 80 {
		t.Fatalf("expected 80x80, got %dx%d", bounds.Dx(), bounds.Dy())
	}
	// Module size = 80/4 = 20px. Row 1 only has col 0 dark.
	// Centre of module (1,0) should be black.
	r0, g0, b0, _ := img.At(10, 30).RGBA()
	if r0 != 0 || g0 != 0 || b0 != 0 {
		t.Error("module (1,0) should be dark for jagged row")
	}
	// Centre of module (1,1) should be transparent (beyond row's column count).
	_, _, _, a1 := img.At(30, 30).RGBA()
	if a1 != 0 {
		t.Error("module (1,1) should be transparent for jagged row (column missing)")
	}
}

// TestDrawBarcode2D_RowsColsLargerThanMatrix verifies behavior when
// rows/cols exceed the actual matrix dimensions. The function should
// render what exists and leave the rest white.
func TestDrawBarcode2D_RowsColsLargerThanMatrix(t *testing.T) {
	matrix := [][]bool{
		{true, true},
		{true, true},
	}
	// Claim 5x5 but only provide 2x2 data.
	img := barcode.DrawBarcode2D(matrix, 5, 5, 100, 100)
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Fatalf("expected 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}
	// Module size = 100/5 = 20px.
	// Module (0,0) should be dark.
	r0, g0, b0, _ := img.At(10, 10).RGBA()
	if r0 != 0 || g0 != 0 || b0 != 0 {
		t.Error("module (0,0) should be dark")
	}
	// Module (3,3) should be transparent (row 3 doesn't exist in matrix).
	_, _, _, a3 := img.At(70, 70).RGBA()
	if a3 != 0 {
		t.Error("module (3,3) should be transparent (beyond matrix)")
	}
}

// TestDrawBarcode2D_NegativeDimensions verifies that negative width/height
// still returns an image without panicking.
func TestDrawBarcode2D_NegativeDimensions(t *testing.T) {
	matrix := [][]bool{{true}}
	img := barcode.DrawBarcode2D(matrix, 1, 1, -10, -10)
	if img == nil {
		t.Fatal("expected non-nil image for negative dimensions")
	}
}

// TestDrawBarcode2D_NonSquare verifies rendering into a non-square image
// where width != height. Modules should be rectangular.
func TestDrawBarcode2D_NonSquare(t *testing.T) {
	matrix := [][]bool{
		{true, false},
		{false, true},
	}
	const w, h = 200, 100
	img := barcode.DrawBarcode2D(matrix, 2, 2, w, h)
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != w || bounds.Dy() != h {
		t.Fatalf("expected %dx%d, got %dx%d", w, h, bounds.Dx(), bounds.Dy())
	}
	// Module (0,0): x=0..100, y=0..50. Centre = (50, 25). Should be dark.
	r0, g0, b0, _ := img.At(50, 25).RGBA()
	if r0 != 0 || g0 != 0 || b0 != 0 {
		t.Error("module (0,0) should be dark in non-square image")
	}
	// Module (0,1): x=100..200, y=0..50. Centre = (150, 25). Should be transparent (light).
	_, _, _, a1 := img.At(150, 25).RGBA()
	if a1 != 0 {
		t.Error("module (0,1) should be transparent in non-square image")
	}
	// Module (1,1): x=100..200, y=50..100. Centre = (150, 75). Should be dark.
	r2, g2, b2, _ := img.At(150, 75).RGBA()
	if r2 != 0 || g2 != 0 || b2 != 0 {
		t.Error("module (1,1) should be dark in non-square image")
	}
}

// TestDrawBarcode2D_AllDark verifies that a fully-dark matrix fills
// the entire image with black pixels.
func TestDrawBarcode2D_AllDark(t *testing.T) {
	matrix := [][]bool{
		{true, true},
		{true, true},
	}
	img := barcode.DrawBarcode2D(matrix, 2, 2, 100, 100)
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	// Sample several points; all should be black.
	for _, pt := range [][2]int{{0, 0}, {25, 25}, {50, 50}, {75, 75}, {99, 99}} {
		r, g, b, _ := img.At(pt[0], pt[1]).RGBA()
		if r != 0 || g != 0 || b != 0 {
			t.Errorf("pixel (%d,%d) should be black in all-dark matrix", pt[0], pt[1])
		}
	}
}

// TestDrawBarcode2D_AllLight verifies that a fully-light matrix produces
// an all-transparent image.
func TestDrawBarcode2D_AllLight(t *testing.T) {
	matrix := [][]bool{
		{false, false},
		{false, false},
	}
	img := barcode.DrawBarcode2D(matrix, 2, 2, 100, 100)
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	for _, pt := range [][2]int{{0, 0}, {25, 25}, {50, 50}, {75, 75}, {99, 99}} {
		_, _, _, a := img.At(pt[0], pt[1]).RGBA()
		if a != 0 {
			t.Errorf("pixel (%d,%d) should be transparent in all-light matrix", pt[0], pt[1])
		}
	}
}

// TestDrawBarcode2D_SingleModule verifies a 1x1 matrix fills the entire image.
func TestDrawBarcode2D_SingleModule(t *testing.T) {
	matrix := [][]bool{{true}}
	img := barcode.DrawBarcode2D(matrix, 1, 1, 50, 50)
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	// The single dark module spans the whole image.
	for _, pt := range [][2]int{{0, 0}, {25, 25}, {49, 49}} {
		r, g, b, _ := img.At(pt[0], pt[1]).RGBA()
		if r != 0 || g != 0 || b != 0 {
			t.Errorf("pixel (%d,%d) should be black for single dark module", pt[0], pt[1])
		}
	}
}

// TestDrawBarcode2D_SmallOutputSize verifies rendering when the output
// is smaller than the module count (some modules map to 0 px width,
// triggering the x1 = x0+1 / y1 = y0+1 guard).
func TestDrawBarcode2D_SmallOutputSize(t *testing.T) {
	const size = 10
	matrix := make([][]bool, size)
	for r := range matrix {
		matrix[r] = make([]bool, size)
		for c := range matrix[r] {
			matrix[r][c] = true
		}
	}
	// 3x3 image for a 10x10 matrix — many modules map to the same pixel.
	img := barcode.DrawBarcode2D(matrix, size, size, 3, 3)
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 3 || bounds.Dy() != 3 {
		t.Fatalf("expected 3x3, got %dx%d", bounds.Dx(), bounds.Dy())
	}
	// All pixels should be black since all modules are dark.
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r != 0 || g != 0 || b != 0 {
				t.Errorf("pixel (%d,%d) should be black", x, y)
			}
		}
	}
}

// TestMatrix2DProvider_Interface verifies the Matrix2DProvider interface
// can be satisfied by a simple implementation and used with DrawBarcode2D.
func TestMatrix2DProvider_Interface(t *testing.T) {
	p := &testMatrix2DProvider{
		data: [][]bool{
			{true, false},
			{false, true},
		},
		rows: 2,
		cols: 2,
	}

	// Verify interface compliance.
	var _ barcode.Matrix2DProvider = p

	matrix, rows, cols := p.GetMatrix()
	img := barcode.DrawBarcode2D(matrix, rows, cols, 100, 100)
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	// Module (0,0) dark, (0,1) light (transparent).
	r0, g0, b0, a0 := img.At(25, 25).RGBA()
	_, _, _, a1 := img.At(75, 25).RGBA()
	if r0 != 0 || g0 != 0 || b0 != 0 || a0 == 0 {
		t.Error("module (0,0) should be dark (opaque black)")
	}
	if a1 != 0 {
		t.Error("module (0,1) should be transparent (light)")
	}
}

// testMatrix2DProvider is a minimal Matrix2DProvider for testing.
type testMatrix2DProvider struct {
	data [][]bool
	rows int
	cols int
}

func (p *testMatrix2DProvider) GetMatrix() ([][]bool, int, int) {
	return p.data, p.rows, p.cols
}

// TestDrawBarcode2D_ModuleBoundaryPixels verifies that pixels right at
// module boundaries are assigned to the correct module.
func TestDrawBarcode2D_ModuleBoundaryPixels(t *testing.T) {
	matrix := [][]bool{
		{true, false},
		{false, true},
	}
	// 100x100 image, 2x2 modules → each module is 50x50.
	img := barcode.DrawBarcode2D(matrix, 2, 2, 100, 100)
	if img == nil {
		t.Fatal("expected non-nil image")
	}

	isBlack := func(x, y int) bool {
		r, g, b, a := img.At(x, y).RGBA()
		return r == 0 && g == 0 && b == 0 && a > 0
	}
	isTransparent := func(x, y int) bool {
		_, _, _, av := img.At(x, y).RGBA()
		return av == 0
	}

	// Last pixel of dark module (0,0) at x=49, y=49 should be black.
	if !isBlack(49, 49) {
		t.Error("pixel (49,49) should be black (last pixel of module 0,0)")
	}
	// First pixel of light module (0,1) at x=50, y=0 should be transparent.
	if !isTransparent(50, 0) {
		t.Error("pixel (50,0) should be transparent (first pixel of module 0,1)")
	}
	// First pixel of dark module (1,1) at x=50, y=50 should be black.
	if !isBlack(50, 50) {
		t.Error("pixel (50,50) should be black (first pixel of module 1,1)")
	}

	// Verify light module (0,1) pixel is transparent.
	_, _, _, a := img.At(75, 25).RGBA()
	if a != 0 {
		t.Errorf("expected transparent pixel in light module, got alpha %d", a)
	}
}

// TestDrawBarcode2D_WidthHeightOnePixel verifies edge case of 1x1 output.
func TestDrawBarcode2D_WidthHeightOnePixel(t *testing.T) {
	matrix := [][]bool{{true}}
	img := barcode.DrawBarcode2D(matrix, 1, 1, 1, 1)
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 1 || bounds.Dy() != 1 {
		t.Fatalf("expected 1x1, got %dx%d", bounds.Dx(), bounds.Dy())
	}
	r, g, b, _ := img.At(0, 0).RGBA()
	if r != 0 || g != 0 || b != 0 {
		t.Error("single pixel should be black for dark module")
	}
}

// TestDrawBarcode2D_NilMatrixPositiveRowsCols verifies that a nil matrix
// with positive rows/cols doesn't panic.
func TestDrawBarcode2D_NilMatrixPositiveRowsCols(t *testing.T) {
	img := barcode.DrawBarcode2D(nil, 5, 5, 100, 100)
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	// Should be all transparent since matrix is nil and loop breaks immediately.
	_, _, _, a := img.At(50, 50).RGBA()
	if a != 0 {
		t.Error("centre pixel should be transparent for nil matrix with positive rows/cols")
	}
}
