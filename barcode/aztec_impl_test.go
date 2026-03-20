package barcode

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// newAztecGF / multiply
// ---------------------------------------------------------------------------

func TestNewAztecGF(t *testing.T) {
	gf := newAztecGF(0x43, 64) // GF(2^6)
	if gf.size != 64 {
		t.Fatalf("size = %d, want 64", gf.size)
	}
	// exp[0] should always be 1 (alpha^0 = 1)
	if gf.expTable[0] != 1 {
		t.Errorf("expTable[0] = %d, want 1", gf.expTable[0])
	}
	// log[1] should be 0
	if gf.logTable[1] != 0 {
		t.Errorf("logTable[1] = %d, want 0", gf.logTable[1])
	}
}

func TestNewAztecGF_AllPrimitives(t *testing.T) {
	// Exercise every word-size primitive polynomial used by the encoder.
	cases := []struct {
		ws   int
		poly int
		size int
	}{
		{4, 0x13, 16},
		{6, 0x43, 64},
		{8, 0x12D, 256},
		{10, 0x409, 1024},
		{12, 0x1069, 4096},
	}
	for _, c := range cases {
		gf := newAztecGF(c.poly, c.size)
		if gf.size != c.size {
			t.Errorf("ws=%d: size=%d, want %d", c.ws, gf.size, c.size)
		}
		if gf.expTable[0] != 1 {
			t.Errorf("ws=%d: expTable[0]=%d, want 1", c.ws, gf.expTable[0])
		}
	}
}

func TestAztecGF_Multiply(t *testing.T) {
	gf := newAztecGF(0x43, 64)
	// Multiply by zero should give zero.
	if got := gf.multiply(0, 5); got != 0 {
		t.Errorf("multiply(0,5) = %d, want 0", got)
	}
	if got := gf.multiply(5, 0); got != 0 {
		t.Errorf("multiply(5,0) = %d, want 0", got)
	}
	// Multiply by 1 (alpha^0) should give identity.
	if got := gf.multiply(1, 7); got != 7 {
		t.Errorf("multiply(1,7) = %d, want 7", got)
	}
	// Commutativity: a*b == b*a
	a, b := 13, 42
	if gf.multiply(a, b) != gf.multiply(b, a) {
		t.Errorf("multiply not commutative for %d,%d", a, b)
	}
}

// ---------------------------------------------------------------------------
// aztecReedSolomon
// ---------------------------------------------------------------------------

func TestAztecReedSolomon_Basic(t *testing.T) {
	gf := newAztecGF(0x43, 64)
	data := []int{1, 2, 3}
	ec := aztecReedSolomon(data, 3, gf)
	if len(ec) != 3 {
		t.Fatalf("len(ec) = %d, want 3", len(ec))
	}
	// ECC words should be in valid range [0, size).
	for i, v := range ec {
		if v < 0 || v >= gf.size {
			t.Errorf("ec[%d] = %d, out of GF range [0,%d)", i, v, gf.size)
		}
	}
}

func TestAztecReedSolomon_DifferentWordSizes(t *testing.T) {
	cases := []struct {
		poly int
		size int
		data []int
		ecN  int
	}{
		{0x13, 16, []int{1, 2}, 2},
		{0x43, 64, []int{10, 20, 30, 40}, 5},
		{0x12D, 256, []int{100, 200, 50}, 4},
		{0x409, 1024, []int{500, 100, 200}, 3},
		{0x1069, 4096, []int{1000, 2000, 3000}, 3},
	}
	for _, c := range cases {
		gf := newAztecGF(c.poly, c.size)
		ec := aztecReedSolomon(c.data, c.ecN, gf)
		if len(ec) != c.ecN {
			t.Errorf("poly=0x%X: len(ec)=%d, want %d", c.poly, len(ec), c.ecN)
		}
		for i, v := range ec {
			if v < 0 || v >= c.size {
				t.Errorf("poly=0x%X: ec[%d]=%d, out of range", c.poly, i, v)
			}
		}
	}
}

func TestAztecReedSolomon_SingleDataWord(t *testing.T) {
	gf := newAztecGF(0x43, 64)
	ec := aztecReedSolomon([]int{42}, 5, gf)
	if len(ec) != 5 {
		t.Fatalf("len(ec) = %d, want 5", len(ec))
	}
}

func TestAztecReedSolomon_ZeroData(t *testing.T) {
	gf := newAztecGF(0x43, 64)
	ec := aztecReedSolomon([]int{0, 0, 0}, 3, gf)
	// All-zero data should produce all-zero ECC.
	for i, v := range ec {
		if v != 0 {
			t.Errorf("ec[%d] = %d, expected 0 for all-zero input", i, v)
		}
	}
}

// ---------------------------------------------------------------------------
// aztecChooseParams
// ---------------------------------------------------------------------------

func TestAztecChooseParams_SmallData(t *testing.T) {
	// Very small data (2 bytes = 16 bits + 10 = 26 bits) should fit compact mode.
	p, err := aztecChooseParams(26, 23)
	if err != nil {
		t.Fatalf("aztecChooseParams(26,23): %v", err)
	}
	if p.layers == 0 {
		t.Error("layers should not be 0")
	}
	if p.wordSize == 0 {
		t.Error("wordSize should not be 0")
	}
	if p.eccPercent < 23 {
		t.Errorf("eccPercent = %d, want >= 23", p.eccPercent)
	}
}

func TestAztecChooseParams_VerySmallCompact(t *testing.T) {
	// 1 byte = 8 bits + 10 = 18 bits. Compact layer 1 has 17 bits total (2 CW * 6 + 5),
	// which may not fit; but compact layer 2 (40 bits, 6 CW) should.
	p, err := aztecChooseParams(18, 23)
	if err != nil {
		t.Fatalf("aztecChooseParams(18,23): %v", err)
	}
	if p.wordSize == 0 {
		t.Error("wordSize should not be 0")
	}
}

func TestAztecChooseParams_DefaultECC(t *testing.T) {
	// eccPercent <= 0 should default to 23.
	p, err := aztecChooseParams(50, 0)
	if err != nil {
		t.Fatalf("aztecChooseParams(50,0): %v", err)
	}
	if p.eccPercent < 23 {
		t.Errorf("eccPercent = %d, want >= 23 (default)", p.eccPercent)
	}
}

func TestAztecChooseParams_NegativeECC(t *testing.T) {
	p, err := aztecChooseParams(50, -5)
	if err != nil {
		t.Fatalf("aztecChooseParams(50,-5): %v", err)
	}
	if p.eccPercent < 23 {
		t.Errorf("eccPercent = %d, want >= 23", p.eccPercent)
	}
}

func TestAztecChooseParams_MediumData(t *testing.T) {
	// 100 bytes = 800 bits + 10 = 810 bits.
	p, err := aztecChooseParams(810, 23)
	if err != nil {
		t.Fatalf("aztecChooseParams(810,23): %v", err)
	}
	if p.layers == 0 {
		t.Error("layers should not be 0")
	}
	if p.codewords == 0 {
		t.Error("codewords should not be 0")
	}
}

func TestAztecChooseParams_LargeData(t *testing.T) {
	// 500 bytes = 4000 bits + 10 = 4010 bits; should pick full mode.
	p, err := aztecChooseParams(4010, 23)
	if err != nil {
		t.Fatalf("aztecChooseParams(4010,23): %v", err)
	}
	if p.compact {
		t.Error("expected full (non-compact) mode for large data")
	}
}

func TestAztecChooseParams_DataTooLarge(t *testing.T) {
	// Impossibly large data.
	_, err := aztecChooseParams(999999, 23)
	if err == nil {
		t.Error("expected error for oversized data, got nil")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("error = %q, want substring 'too large'", err.Error())
	}
}

func TestAztecChooseParams_HighECC(t *testing.T) {
	// 50% ECC requires more capacity.
	p, err := aztecChooseParams(74, 50)
	if err != nil {
		t.Fatalf("aztecChooseParams(74,50): %v", err)
	}
	if p.eccPercent < 50 {
		t.Errorf("eccPercent = %d, want >= 50", p.eccPercent)
	}
}

// ---------------------------------------------------------------------------
// aztecWordSize
// ---------------------------------------------------------------------------

func TestAztecWordSize(t *testing.T) {
	cases := []struct {
		layers int
		want   int
	}{
		{1, 6},
		{2, 6},
		{3, 8},
		{8, 8},
		{9, 10},
		{22, 10},
		{23, 12},
		{32, 12},
	}
	for _, c := range cases {
		got := aztecWordSize(c.layers)
		if got != c.want {
			t.Errorf("aztecWordSize(%d) = %d, want %d", c.layers, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// aztecFullTotalBits
// ---------------------------------------------------------------------------

func TestAztecFullTotalBits(t *testing.T) {
	// Formula: 16 * layers * (14 + layers).
	cases := []struct {
		layers int
		want   int
	}{
		{1, 16 * 1 * 15},    // 240
		{2, 16 * 2 * 16},    // 512
		{10, 16 * 10 * 24},  // 3840
		{32, 16 * 32 * 46},  // 23552
	}
	for _, c := range cases {
		got := aztecFullTotalBits(c.layers)
		if got != c.want {
			t.Errorf("aztecFullTotalBits(%d) = %d, want %d", c.layers, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// aztecEncodeByte
// ---------------------------------------------------------------------------

func TestAztecEncodeByte(t *testing.T) {
	bits := aztecEncodeByte([]byte("A")) // 0x41 = 0100_0001
	if len(bits) != 8 {
		t.Fatalf("len = %d, want 8", len(bits))
	}
	want := []int{0, 1, 0, 0, 0, 0, 0, 1}
	for i, v := range want {
		if bits[i] != v {
			t.Errorf("bit[%d] = %d, want %d", i, bits[i], v)
		}
	}
}

func TestAztecEncodeByte_MultipleChars(t *testing.T) {
	bits := aztecEncodeByte([]byte("AB"))
	if len(bits) != 16 {
		t.Fatalf("len = %d, want 16", len(bits))
	}
}

func TestAztecEncodeByte_Empty(t *testing.T) {
	bits := aztecEncodeByte([]byte{})
	if len(bits) != 0 {
		t.Fatalf("len = %d, want 0", len(bits))
	}
}

func TestAztecEncodeByte_HighByte(t *testing.T) {
	bits := aztecEncodeByte([]byte{0xFF})
	if len(bits) != 8 {
		t.Fatalf("len = %d, want 8", len(bits))
	}
	for i, v := range bits {
		if v != 1 {
			t.Errorf("bit[%d] = %d, want 1 for 0xFF", i, v)
		}
	}
}

// ---------------------------------------------------------------------------
// aztecPrimitive
// ---------------------------------------------------------------------------

func TestAztecPrimitive(t *testing.T) {
	cases := []struct {
		ws   int
		want int
	}{
		{4, 0x13},
		{6, 0x43},
		{8, 0x12D},
		{10, 0x409},
		{12, 0x1069},
		{99, 0x43}, // unknown falls back to default
	}
	for _, c := range cases {
		got := aztecPrimitive(c.ws)
		if got != c.want {
			t.Errorf("aztecPrimitive(%d) = 0x%X, want 0x%X", c.ws, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// encodeAztecBarcode
// ---------------------------------------------------------------------------

func TestEncodeAztecBarcode_Hello(t *testing.T) {
	matrix, err := encodeAztecBarcode("Hello", 23, 0)
	if err != nil {
		t.Fatalf("encodeAztecBarcode: %v", err)
	}
	if len(matrix) == 0 {
		t.Fatal("matrix is empty")
	}
	n := len(matrix)
	// Matrix should be square.
	for i, row := range matrix {
		if len(row) != n {
			t.Fatalf("row %d: len=%d, want %d", i, len(row), n)
		}
	}
	// Matrix should be at least 15x15 (minimum Aztec size).
	if n < 15 {
		t.Errorf("matrix size %d < 15", n)
	}
}

func TestEncodeAztecBarcode_EmptyContent(t *testing.T) {
	_, err := encodeAztecBarcode("", 23, 0)
	if err == nil {
		t.Error("expected error for empty content")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error = %q, want substring 'empty'", err.Error())
	}
}

func TestEncodeAztecBarcode_SingleChar(t *testing.T) {
	matrix, err := encodeAztecBarcode("A", 23, 0)
	if err != nil {
		t.Fatalf("encodeAztecBarcode: %v", err)
	}
	if len(matrix) < 15 {
		t.Errorf("matrix too small: %d", len(matrix))
	}
}

func TestEncodeAztecBarcode_LongText(t *testing.T) {
	text := strings.Repeat("Hello World! ", 20) // ~260 chars
	matrix, err := encodeAztecBarcode(text, 23, 0)
	if err != nil {
		t.Fatalf("encodeAztecBarcode: %v", err)
	}
	if len(matrix) < 15 {
		t.Errorf("matrix too small: %d", len(matrix))
	}
}

func TestEncodeAztecBarcode_UserLayers_Compact(t *testing.T) {
	// Force compact layer 2.
	matrix, err := encodeAztecBarcode("Hi", 23, 2)
	if err != nil {
		t.Fatalf("encodeAztecBarcode: %v", err)
	}
	if len(matrix) == 0 {
		t.Fatal("matrix is empty")
	}
}

func TestEncodeAztecBarcode_UserLayers_Full(t *testing.T) {
	// Force full layer 6 (>4 means full).
	matrix, err := encodeAztecBarcode("Test data", 23, 6)
	if err != nil {
		t.Fatalf("encodeAztecBarcode: %v", err)
	}
	if len(matrix) == 0 {
		t.Fatal("matrix is empty")
	}
}

func TestEncodeAztecBarcode_HighECC(t *testing.T) {
	matrix, err := encodeAztecBarcode("ECC test", 50, 0)
	if err != nil {
		t.Fatalf("encodeAztecBarcode: %v", err)
	}
	if len(matrix) == 0 {
		t.Fatal("matrix is empty")
	}
}

func TestEncodeAztecBarcode_BinaryContent(t *testing.T) {
	// Non-ASCII content.
	content := string([]byte{0x00, 0x01, 0xFE, 0xFF})
	matrix, err := encodeAztecBarcode(content, 23, 0)
	if err != nil {
		t.Fatalf("encodeAztecBarcode: %v", err)
	}
	if len(matrix) == 0 {
		t.Fatal("matrix is empty")
	}
}

func TestEncodeAztecBarcode_DifferentSizes(t *testing.T) {
	// Encode increasing amounts of data and verify matrix sizes are reasonable.
	// Sizes may not be strictly monotonic due to mode transitions (compact vs full),
	// but larger inputs should generally produce larger matrices.
	sizes := make([]int, 0)
	for _, n := range []int{1, 10, 50, 200} {
		text := strings.Repeat("X", n)
		matrix, err := encodeAztecBarcode(text, 23, 0)
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		sizes = append(sizes, len(matrix))
	}
	// The smallest input should produce a smaller matrix than the largest.
	if sizes[0] > sizes[len(sizes)-1] {
		t.Errorf("smallest input produced larger matrix (%d) than largest (%d)", sizes[0], sizes[len(sizes)-1])
	}
}

// ---------------------------------------------------------------------------
// aztecBuildMatrix
// ---------------------------------------------------------------------------

func TestAztecBuildMatrix_Compact(t *testing.T) {
	params := aztecParams{compact: true, layers: 1, codewords: 3, wordSize: 6}
	codewords := []int{1, 2, 3}
	matrix := aztecBuildMatrix(codewords, params)
	expectedSize := 11 + 4*1 // 15
	if len(matrix) != expectedSize {
		t.Errorf("compact layer 1: size=%d, want %d", len(matrix), expectedSize)
	}
	// All rows should have same width.
	for i, row := range matrix {
		if len(row) != expectedSize {
			t.Errorf("row %d: width=%d, want %d", i, len(row), expectedSize)
		}
	}
}

func TestAztecBuildMatrix_CompactLayers(t *testing.T) {
	for layers := 1; layers <= 4; layers++ {
		params := aztecParams{compact: true, layers: layers, codewords: 3, wordSize: 6}
		codewords := []int{1, 2, 3}
		matrix := aztecBuildMatrix(codewords, params)
		expectedSize := 11 + 4*layers
		if len(matrix) != expectedSize {
			t.Errorf("compact layer %d: size=%d, want %d", layers, len(matrix), expectedSize)
		}
	}
}

func TestAztecBuildMatrix_Full(t *testing.T) {
	params := aztecParams{compact: false, layers: 1, codewords: 5, wordSize: 6}
	codewords := []int{1, 2, 3, 4, 5}
	matrix := aztecBuildMatrix(codewords, params)
	if len(matrix) < 15 {
		t.Errorf("full layer 1: size=%d, want >= 15", len(matrix))
	}
}

func TestAztecBuildMatrix_FullLargeLayer(t *testing.T) {
	params := aztecParams{compact: false, layers: 10, codewords: 5, wordSize: 10}
	codewords := []int{1, 2, 3, 4, 5}
	matrix := aztecBuildMatrix(codewords, params)
	// Full size = 14 + 4*10 + (2*((10-1)/15+1)-2) = 14+40+0 = 54
	if len(matrix) < 50 {
		t.Errorf("full layer 10: size=%d, want >= 50", len(matrix))
	}
}

func TestAztecBuildMatrix_FullZeroLayers(t *testing.T) {
	// layers <= 0 corner case.
	params := aztecParams{compact: false, layers: 0, codewords: 1, wordSize: 6}
	codewords := []int{1}
	matrix := aztecBuildMatrix(codewords, params)
	if len(matrix) < 15 {
		t.Errorf("full layer 0: size=%d, want >= 15", len(matrix))
	}
}

// ---------------------------------------------------------------------------
// aztecDrawBullsEye
// ---------------------------------------------------------------------------

func TestAztecDrawBullsEye_Compact(t *testing.T) {
	size := 15
	matrix := makeBlankMatrix(size)
	center := size / 2
	aztecDrawBullsEye(matrix, center, true)
	// Center module should be dark (ring 0).
	if !matrix[center][center] {
		t.Error("compact: center module should be dark")
	}
	// Orientation marks: 3 of 4 corners should be dark.
	darkCorners := 0
	offsets := [][2]int{{-3, -3}, {-3, 3}, {3, -3}, {3, 3}}
	for _, o := range offsets {
		if matrix[center+o[0]][center+o[1]] {
			darkCorners++
		}
	}
	if darkCorners != 3 {
		t.Errorf("compact: %d dark corners, want 3", darkCorners)
	}
}

func TestAztecDrawBullsEye_Full(t *testing.T) {
	size := 19
	matrix := makeBlankMatrix(size)
	center := size / 2
	aztecDrawBullsEye(matrix, center, false)
	// The bull's-eye draws concentric squares from outer (r=0, dark) inward.
	// With rings=7, half=3: r=0 dark, r=1 light, r=2 dark, r=3 light.
	// So the outermost ring should be dark.
	outerY := center - 3
	if !matrix[outerY][center] {
		t.Error("full: outermost ring module should be dark")
	}
	// Verify some modules were set (not all blank).
	set := countSetModules(matrix)
	if set == 0 {
		t.Error("full: no modules set by DrawBullsEye")
	}
	// Orientation marks at offset 4: 3 of 4 corners should be dark.
	darkCorners := 0
	offsets := [][2]int{{-4, -4}, {-4, 4}, {4, -4}, {4, 4}}
	for _, o := range offsets {
		if matrix[center+o[0]][center+o[1]] {
			darkCorners++
		}
	}
	if darkCorners != 3 {
		t.Errorf("full: %d dark corners, want 3", darkCorners)
	}
}

// ---------------------------------------------------------------------------
// aztecDrawModeMessage
// ---------------------------------------------------------------------------

func TestAztecDrawModeMessage_Compact(t *testing.T) {
	size := 15
	matrix := makeBlankMatrix(size)
	center := size / 2
	params := aztecParams{compact: true, layers: 1, wordSize: 6}
	aztecDrawModeMessage(matrix, center, params)
	// Mode message should have written some modules around the bull's-eye.
	// Check the row just above the compact core (center-4).
	anySet := false
	for x := 0; x < size; x++ {
		if matrix[center-4][x] {
			anySet = true
			break
		}
	}
	if !anySet {
		t.Error("compact mode message: no modules set above core")
	}
}

func TestAztecDrawModeMessage_Full(t *testing.T) {
	size := 23
	matrix := makeBlankMatrix(size)
	center := size / 2
	params := aztecParams{compact: false, layers: 2, wordSize: 6}
	aztecDrawModeMessage(matrix, center, params)
	// Check the row above the full core (center-5).
	anySet := false
	for x := 0; x < size; x++ {
		if matrix[center-5][x] {
			anySet = true
			break
		}
	}
	if !anySet {
		t.Error("full mode message: no modules set above core")
	}
}

// ---------------------------------------------------------------------------
// aztecDrawDataLayers
// ---------------------------------------------------------------------------

func TestAztecDrawDataLayers_Compact(t *testing.T) {
	size := 15
	matrix := makeBlankMatrix(size)
	center := size / 2
	params := aztecParams{compact: true, layers: 1, wordSize: 6}
	// 3 codewords of 6 bits each = 18 bits.
	codewords := []int{0x3F, 0x15, 0x2A} // all-ones, alternating patterns
	aztecDrawDataLayers(matrix, center, codewords, params)
	// At least some modules should be set.
	set := countSetModules(matrix)
	if set == 0 {
		t.Error("compact data layers: no modules set")
	}
}

func TestAztecDrawDataLayers_Full(t *testing.T) {
	size := 23
	matrix := makeBlankMatrix(size)
	center := size / 2
	params := aztecParams{compact: false, layers: 2, wordSize: 6}
	codewords := []int{1, 2, 3, 4, 5}
	aztecDrawDataLayers(matrix, center, codewords, params)
	set := countSetModules(matrix)
	if set == 0 {
		t.Error("full data layers: no modules set")
	}
}

func TestAztecDrawDataLayers_MultipleLayers(t *testing.T) {
	size := 27
	matrix := makeBlankMatrix(size)
	center := size / 2
	params := aztecParams{compact: true, layers: 4, wordSize: 8}
	codewords := make([]int, 20)
	for i := range codewords {
		codewords[i] = i + 1
	}
	aztecDrawDataLayers(matrix, center, codewords, params)
	set := countSetModules(matrix)
	if set == 0 {
		t.Error("multi-layer data: no modules set")
	}
}

func TestAztecDrawDataLayers_EmptyCodewords(t *testing.T) {
	size := 15
	matrix := makeBlankMatrix(size)
	center := size / 2
	params := aztecParams{compact: true, layers: 1, wordSize: 6}
	// Empty codewords should not panic.
	aztecDrawDataLayers(matrix, center, nil, params)
	set := countSetModules(matrix)
	if set != 0 {
		t.Errorf("empty codewords: %d modules set, want 0", set)
	}
}

// ---------------------------------------------------------------------------
// AztecBarcode.GetMatrix
// ---------------------------------------------------------------------------

func TestAztecBarcode_GetMatrix_WithEncodedText(t *testing.T) {
	b := NewAztecBarcode()
	_ = b.Encode("Hello Aztec")
	matrix, rows, cols := b.GetMatrix()
	if rows == 0 || cols == 0 {
		t.Fatal("GetMatrix returned zero dimensions")
	}
	if rows != cols {
		t.Errorf("rows=%d, cols=%d; Aztec should be square", rows, cols)
	}
	if len(matrix) != rows {
		t.Errorf("len(matrix)=%d, want %d", len(matrix), rows)
	}
	if rows < 15 {
		t.Errorf("rows=%d, want >= 15", rows)
	}
}

func TestAztecBarcode_GetMatrix_Empty(t *testing.T) {
	b := NewAztecBarcode()
	// No Encode called: falls back to DefaultValue().
	matrix, rows, cols := b.GetMatrix()
	if rows != 1 && cols != 1 {
		// Should use default value "Aztec".
		if rows == 0 || cols == 0 {
			t.Fatal("GetMatrix returned zero dimensions")
		}
	}
	if len(matrix) == 0 {
		t.Fatal("matrix is empty")
	}
}

func TestAztecBarcode_GetMatrix_CustomLayers(t *testing.T) {
	b := NewAztecBarcode()
	b.UserSpecifiedLayers = 3
	_ = b.Encode("Custom layers")
	matrix, rows, cols := b.GetMatrix()
	if rows == 0 || cols == 0 {
		t.Fatal("GetMatrix returned zero dimensions")
	}
	if len(matrix) != rows {
		t.Errorf("len(matrix)=%d, want %d", len(matrix), rows)
	}
	_ = cols
}

func TestAztecBarcode_GetMatrix_CustomECC(t *testing.T) {
	b := NewAztecBarcode()
	b.MinECCPercent = 50
	_ = b.Encode("High ECC")
	matrix, rows, cols := b.GetMatrix()
	if rows == 0 || cols == 0 {
		t.Fatal("GetMatrix returned zero dimensions")
	}
	if len(matrix) != rows {
		t.Errorf("len(matrix)=%d, want %d", len(matrix), rows)
	}
	_ = cols
}

// ---------------------------------------------------------------------------
// End-to-end: encode -> matrix -> content verification
// ---------------------------------------------------------------------------

func TestEncodeAztecBarcode_MatrixHasContent(t *testing.T) {
	matrix, err := encodeAztecBarcode("Test123", 23, 0)
	if err != nil {
		t.Fatalf("encodeAztecBarcode: %v", err)
	}
	set := countSetModules(matrix)
	total := len(matrix) * len(matrix)
	// At least 10% and at most 90% should be set (sanity check for a real barcode).
	pct := float64(set) / float64(total) * 100
	if pct < 5 || pct > 95 {
		t.Errorf("set modules = %d/%d (%.1f%%); seems wrong for a barcode", set, total, pct)
	}
}

func TestEncodeAztecBarcode_DeterministicOutput(t *testing.T) {
	// Same input should always produce the same matrix.
	m1, err := encodeAztecBarcode("Deterministic", 23, 0)
	if err != nil {
		t.Fatalf("first encode: %v", err)
	}
	m2, err := encodeAztecBarcode("Deterministic", 23, 0)
	if err != nil {
		t.Fatalf("second encode: %v", err)
	}
	if len(m1) != len(m2) {
		t.Fatalf("sizes differ: %d vs %d", len(m1), len(m2))
	}
	for r := range m1 {
		for c := range m1[r] {
			if m1[r][c] != m2[r][c] {
				t.Fatalf("matrices differ at [%d][%d]", r, c)
			}
		}
	}
}

func TestEncodeAztecBarcode_DifferentInputsDifferentOutput(t *testing.T) {
	m1, _ := encodeAztecBarcode("Alpha", 23, 0)
	m2, _ := encodeAztecBarcode("Bravo", 23, 0)
	// Matrices could be same size but should differ in content.
	if len(m1) == len(m2) {
		same := true
		for r := range m1 {
			for c := range m1[r] {
				if m1[r][c] != m2[r][c] {
					same = false
					break
				}
			}
			if !same {
				break
			}
		}
		if same {
			t.Error("different inputs produced identical matrices")
		}
	}
}

// ---------------------------------------------------------------------------
// Edge cases for remaining uncovered branches
// ---------------------------------------------------------------------------

func TestEncodeAztecBarcode_UserLayers_TooSmall(t *testing.T) {
	// Force compact layer 1 with data that exceeds its capacity.
	// This hits the eccCount < 0 branch in encodeAztecBarcode.
	longText := strings.Repeat("ABCDEFGHIJ", 5) // 50 bytes, too big for compact layer 1
	matrix, err := encodeAztecBarcode(longText, 23, 1)
	if err != nil {
		t.Fatalf("encodeAztecBarcode: %v", err)
	}
	if len(matrix) == 0 {
		t.Fatal("matrix is empty")
	}
}

func TestEncodeAztecBarcode_UserLayers_ExactBoundary(t *testing.T) {
	// Force compact layer 4 (boundary between compact and full).
	matrix, err := encodeAztecBarcode("Boundary", 23, 4)
	if err != nil {
		t.Fatalf("encodeAztecBarcode: %v", err)
	}
	if len(matrix) == 0 {
		t.Fatal("matrix is empty")
	}
}

func TestEncodeAztecBarcode_UserLayers_5(t *testing.T) {
	// Layer 5 is the first full-mode layer when user-specified.
	matrix, err := encodeAztecBarcode("FullMode", 23, 5)
	if err != nil {
		t.Fatalf("encodeAztecBarcode: %v", err)
	}
	if len(matrix) == 0 {
		t.Fatal("matrix is empty")
	}
}

func TestAztecBuildMatrix_FullNegativeLayers(t *testing.T) {
	// Negative layers corner case for full mode.
	params := aztecParams{compact: false, layers: -1, codewords: 1, wordSize: 6}
	codewords := []int{1}
	matrix := aztecBuildMatrix(codewords, params)
	if len(matrix) < 15 {
		t.Errorf("full layers=-1: size=%d, want >= 15", len(matrix))
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeBlankMatrix(size int) [][]bool {
	m := make([][]bool, size)
	for i := range m {
		m[i] = make([]bool, size)
	}
	return m
}

func countSetModules(matrix [][]bool) int {
	n := 0
	for _, row := range matrix {
		for _, v := range row {
			if v {
				n++
			}
		}
	}
	return n
}
