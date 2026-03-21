// qr_datastructs_internal_test.go — internal package tests for QR code data
// structures: qrBitVector, qrByteMatrix, blockPair (inline), qrECLevel.
//
// Verified against C# source:
//   original-dotnet/FastReport.Base/Barcode/QRCode/BitVector.cs
//   original-dotnet/FastReport.Base/Barcode/QRCode/ByteMatrix.cs
//   original-dotnet/FastReport.Base/Barcode/QRCode/BlockPair.cs
//   original-dotnet/FastReport.Base/Barcode/QRCode/ByteArray.cs
//   original-dotnet/FastReport.Base/Barcode/QRCode/ErrorCorrectionLevel.cs
package barcode

import "testing"

// ── qrECLevel ─────────────────────────────────────────────────────────────────

// TestQRECLevel_Ordinals verifies that L/M/Q/H have the correct ordinal values
// as defined in C# ErrorCorrectionLevel.cs (L=0, M=1, Q=2, H=3).
func TestQRECLevel_Ordinals(t *testing.T) {
	tests := []struct {
		name    string
		level   qrECLevel
		ordinal int
		bits    int
	}{
		{"L", qrECL, 0, 0x01},
		{"M", qrECM, 1, 0x00},
		{"Q", qrECQ, 2, 0x03},
		{"H", qrECH, 3, 0x02},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.level.ordinal != tt.ordinal {
				t.Errorf("ordinal: got %d, want %d", tt.level.ordinal, tt.ordinal)
			}
			if tt.level.bits != tt.bits {
				t.Errorf("bits: got %d (0x%02X), want %d (0x%02X)",
					tt.level.bits, tt.level.bits, tt.bits, tt.bits)
			}
		})
	}
}

// TestQRECLevel_Distinct verifies the four EC levels all differ from each other.
func TestQRECLevel_Distinct(t *testing.T) {
	levels := []qrECLevel{qrECL, qrECM, qrECQ, qrECH}
	for i := 0; i < len(levels); i++ {
		for j := i + 1; j < len(levels); j++ {
			if levels[i].ordinal == levels[j].ordinal {
				t.Errorf("EC levels %d and %d share ordinal %d", i, j, levels[i].ordinal)
			}
			if levels[i].bits == levels[j].bits {
				t.Errorf("EC levels %d and %d share bits 0x%02X", i, j, levels[i].bits)
			}
		}
	}
}

// TestQRECLevelFromString_Uppercase verifies all four uppercase letter inputs.
// C# uses a switch on enum; Go uses a string switch. All four recognised values
// must round-trip correctly (go-fastreport-6uh4c).
func TestQRECLevelFromString_Uppercase(t *testing.T) {
	tests := []struct {
		input string
		want  qrECLevel
	}{
		{"L", qrECL},
		{"M", qrECM},
		{"Q", qrECQ},
		{"H", qrECH},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := qrECLevelFromString(tt.input)
			if got != tt.want {
				t.Errorf("qrECLevelFromString(%q) = {ord:%d bits:0x%02X}, want {ord:%d bits:0x%02X}",
					tt.input, got.ordinal, got.bits, tt.want.ordinal, tt.want.bits)
			}
		})
	}
}

// TestQRECLevelFromString_DefaultFallbackIsM verifies that unrecognised input
// falls back to M. "M" is the practical default used throughout the Go port
// (NewQRBarcode default after ErrorCorrection="L" is overridden in tests).
func TestQRECLevelFromString_DefaultFallbackIsM(t *testing.T) {
	cases := []string{"", "X", "l", "m", "q", "h", "LOW", "HIGH"}
	for _, s := range cases {
		got := qrECLevelFromString(s)
		if got != qrECM {
			t.Errorf("qrECLevelFromString(%q) = %v, want qrECM (fallback)", s, got)
		}
	}
}

// ── qrBitVector — bit-packing order ──────────────────────────────────────────

// TestQRBitVector_InitialState verifies the zero-value initial state after
// construction, matching C# BitVector() constructor (sizeInBits=0, array
// pre-allocated to DEFAULT_SIZE_IN_BYTES=32).
func TestQRBitVector_InitialState(t *testing.T) {
	bv := newQRBitVector()
	if bv.size() != 0 {
		t.Errorf("initial size: got %d, want 0", bv.size())
	}
	if bv.sizeInBytes() != 0 {
		t.Errorf("initial sizeInBytes: got %d, want 0", bv.sizeInBytes())
	}
	if len(bv.data) != 32 {
		t.Errorf("initial data capacity: got %d, want 32", len(bv.data))
	}
}

// TestQRBitVector_AppendBit_MSBFirst verifies MSB-first bit-packing order.
// C# BitVector.appendBit sets bit at position (7 - numBitsInLastByte) in the
// current byte. Appending bit 1 followed by bit 0 should yield byte 0x80.
func TestQRBitVector_AppendBit_MSBFirst(t *testing.T) {
	bv := newQRBitVector()
	bv.appendBit(1) // bit index 0 → position 7 within first byte → 0x80
	bv.appendBit(0) // bit index 1 → position 6 within first byte
	if bv.size() != 2 {
		t.Fatalf("size after 2 appendBit: got %d, want 2", bv.size())
	}
	// at(0) must return the first appended bit.
	if bv.at(0) != 1 {
		t.Errorf("at(0) = %d, want 1", bv.at(0))
	}
	if bv.at(1) != 0 {
		t.Errorf("at(1) = %d, want 0", bv.at(1))
	}
	// Internal byte must be 0x80 (only MSB set, second bit is 0).
	if bv.data[0] != 0x80 {
		t.Errorf("data[0] = 0x%02X, want 0x80", bv.data[0])
	}
}

// TestQRBitVector_AppendBit_SingleBitZero verifies that appending a zero bit
// leaves the first byte at 0x00 (no bits set) with size 1.
func TestQRBitVector_AppendBit_SingleBitZero(t *testing.T) {
	bv := newQRBitVector()
	bv.appendBit(0)
	if bv.size() != 1 {
		t.Fatalf("size = %d, want 1", bv.size())
	}
	if bv.at(0) != 0 {
		t.Errorf("at(0) = %d, want 0", bv.at(0))
	}
	if bv.data[0] != 0x00 {
		t.Errorf("data[0] = 0x%02X, want 0x00", bv.data[0])
	}
}

// TestQRBitVector_AppendBits_EightOnes mirrors C# BitVector example
// "appendBits(0xff, 8) adds 11111111".
func TestQRBitVector_AppendBits_EightOnes(t *testing.T) {
	bv := newQRBitVector()
	bv.appendBits(0xff, 8)
	if bv.size() != 8 {
		t.Fatalf("size = %d, want 8", bv.size())
	}
	for i := 0; i < 8; i++ {
		if bv.at(i) != 1 {
			t.Errorf("at(%d) = %d, want 1", i, bv.at(i))
		}
	}
	if bv.data[0] != 0xff {
		t.Errorf("data[0] = 0x%02X, want 0xFF", bv.data[0])
	}
}

// TestQRBitVector_AppendBits_ZeroNibble mirrors C# example
// "appendBits(0x00, 4) adds 0000".
func TestQRBitVector_AppendBits_ZeroNibble(t *testing.T) {
	bv := newQRBitVector()
	bv.appendBits(0x00, 4)
	if bv.size() != 4 {
		t.Fatalf("size = %d, want 4", bv.size())
	}
	for i := 0; i < 4; i++ {
		if bv.at(i) != 0 {
			t.Errorf("at(%d) = %d, want 0", i, bv.at(i))
		}
	}
}

// TestQRBitVector_AppendBits_SingleBitZero mirrors C# example
// "appendBits(0x00, 1) adds 0".
func TestQRBitVector_AppendBits_SingleBitZeroValue(t *testing.T) {
	bv := newQRBitVector()
	bv.appendBits(0x00, 1)
	if bv.size() != 1 {
		t.Fatalf("size = %d, want 1", bv.size())
	}
	if bv.at(0) != 0 {
		t.Errorf("at(0) = %d, want 0", bv.at(0))
	}
}

// TestQRBitVector_AppendBits_MSBFirst_Value verifies that appendBits encodes
// MSB-first: appending 0b1011 (=11) in 4 bits should give bits 1,0,1,1.
func TestQRBitVector_AppendBits_MSBFirst_Value(t *testing.T) {
	bv := newQRBitVector()
	bv.appendBits(0b1011, 4) // = 11 decimal
	want := []int{1, 0, 1, 1}
	for i, w := range want {
		if bv.at(i) != w {
			t.Errorf("at(%d) = %d, want %d", i, bv.at(i), w)
		}
	}
}

// TestQRBitVector_AppendBits_TwoBytes verifies that appending a 16-bit value
// produces two correctly packed bytes.
func TestQRBitVector_AppendBits_TwoBytes(t *testing.T) {
	bv := newQRBitVector()
	bv.appendBits(0xABCD, 16)
	if bv.size() != 16 {
		t.Fatalf("size = %d, want 16", bv.size())
	}
	if bv.data[0] != 0xAB {
		t.Errorf("data[0] = 0x%02X, want 0xAB", bv.data[0])
	}
	if bv.data[1] != 0xCD {
		t.Errorf("data[1] = 0x%02X, want 0xCD", bv.data[1])
	}
}

// TestQRBitVector_AppendBitVector verifies that appendBitVector copies all bits
// from another vector in order (same as C# BitVector.appendBitVector which
// iterates from 0 to size calling appendBit(other.at(i))).
func TestQRBitVector_AppendBitVector(t *testing.T) {
	src := newQRBitVector()
	src.appendBits(0b10110100, 8) // = 0xB4

	dst := newQRBitVector()
	dst.appendBits(0b11001100, 8) // = 0xCC
	dst.appendBitVector(src)

	if dst.size() != 16 {
		t.Fatalf("size after appendBitVector: got %d, want 16", dst.size())
	}
	if dst.data[0] != 0xCC {
		t.Errorf("data[0] = 0x%02X, want 0xCC", dst.data[0])
	}
	if dst.data[1] != 0xB4 {
		t.Errorf("data[1] = 0x%02X, want 0xB4", dst.data[1])
	}
}

// TestQRBitVector_XorWith verifies byte-wise XOR matching C# BitVector.xor().
// C# xor iterates over (sizeInBits+7)>>3 bytes and XORs array[i] with other.array[i].
func TestQRBitVector_XorWith(t *testing.T) {
	a := newQRBitVector()
	a.appendBits(0xFF, 8)

	b := newQRBitVector()
	b.appendBits(0x0F, 8)

	a.xorWith(b)

	// 0xFF ^ 0x0F = 0xF0
	if a.data[0] != 0xF0 {
		t.Errorf("after XOR: data[0] = 0x%02X, want 0xF0", a.data[0])
	}
	if a.size() != 8 {
		t.Errorf("size unchanged after XOR: got %d, want 8", a.size())
	}
}

// TestQRBitVector_XorWith_Identity verifies X XOR X = 0.
func TestQRBitVector_XorWith_Identity(t *testing.T) {
	a := newQRBitVector()
	a.appendBits(0xA5, 8)

	b := newQRBitVector()
	b.appendBits(0xA5, 8)

	a.xorWith(b)

	if a.data[0] != 0x00 {
		t.Errorf("X XOR X: data[0] = 0x%02X, want 0x00", a.data[0])
	}
}

// TestQRBitVector_SizeInBytes verifies the sizeInBytes() helper (same formula
// as C# sizeInBytes() = (sizeInBits + 7) >> 3).
func TestQRBitVector_SizeInBytes(t *testing.T) {
	tests := []struct {
		numBits   int
		wantBytes int
	}{
		{0, 0},
		{1, 1},
		{7, 1},
		{8, 1},
		{9, 2},
		{15, 2},
		{16, 2},
		{17, 3},
	}
	for _, tt := range tests {
		bv := newQRBitVector()
		for i := 0; i < tt.numBits; i++ {
			bv.appendBit(i & 1)
		}
		got := bv.sizeInBytes()
		if got != tt.wantBytes {
			t.Errorf("sizeInBytes after %d bits: got %d, want %d", tt.numBits, got, tt.wantBytes)
		}
	}
}

// TestQRBitVector_AppendByte_Grow verifies dynamic growth: appending more than
// 32 bytes causes the underlying slice to double (matching C# appendByte
// which doubles the array when sizeInBits>>3 == array.Length).
func TestQRBitVector_AppendByte_Grow(t *testing.T) {
	bv := newQRBitVector()
	// Append 33 bytes (264 bits) to force at least one doubling.
	for i := 0; i < 33; i++ {
		bv.appendBits(int(i), 8)
	}
	if bv.size() != 33*8 {
		t.Errorf("size = %d, want %d", bv.size(), 33*8)
	}
	// Verify first byte preserved correctly.
	if bv.data[0] != 0x00 {
		t.Errorf("data[0] = 0x%02X, want 0x00", bv.data[0])
	}
	// Verify last byte.
	if bv.data[32] != 32 {
		t.Errorf("data[32] = 0x%02X, want 0x20", bv.data[32])
	}
}

// ── qrByteMatrix — coordinate convention ─────────────────────────────────────

// TestQRByteMatrix_Init verifies that a freshly created matrix has correct
// width and height and all cells are zero (default int8 zero value), matching
// C# ByteMatrix constructor which allocates bytes[height][width].
func TestQRByteMatrix_Init(t *testing.T) {
	m := newQRByteMatrix(5, 7)
	if m.width != 5 {
		t.Errorf("width = %d, want 5", m.width)
	}
	if m.height != 7 {
		t.Errorf("height = %d, want 7", m.height)
	}
	// All cells default to zero.
	for y := 0; y < 7; y++ {
		for x := 0; x < 5; x++ {
			if m.get(x, y) != 0 {
				t.Errorf("get(%d,%d) = %d, want 0", x, y, m.get(x, y))
			}
		}
	}
}

// TestQRByteMatrix_GetSet_XColYRow verifies the coordinate convention:
// x = column, y = row, stored as bytes[y][x].
// This matches C# ByteMatrix.get_Renamed(x, y) = bytes[y][x].
func TestQRByteMatrix_GetSet_XColYRow(t *testing.T) {
	m := newQRByteMatrix(3, 3)
	// Set (col=2, row=0) = 1  and  (col=0, row=2) = -1.
	m.set(2, 0, 1)
	m.set(0, 2, -1)

	if m.get(2, 0) != 1 {
		t.Errorf("get(2,0) = %d, want 1", m.get(2, 0))
	}
	if m.get(0, 2) != -1 {
		t.Errorf("get(0,2) = %d, want -1", m.get(0, 2))
	}
	// Verify internal storage: bytes[row][col].
	if m.bytes[0][2] != 1 {
		t.Errorf("bytes[0][2] = %d, want 1", m.bytes[0][2])
	}
	if m.bytes[2][0] != -1 {
		t.Errorf("bytes[2][0] = %d, want -1", m.bytes[2][0])
	}
	// Neighbours untouched.
	if m.get(0, 0) != 0 {
		t.Errorf("get(0,0) = %d, want 0 (untouched)", m.get(0, 0))
	}
}

// TestQRByteMatrix_Clear verifies that clear() fills every cell with the given
// value, matching C# ByteMatrix.clear(sbyte value_Renamed).
func TestQRByteMatrix_Clear(t *testing.T) {
	m := newQRByteMatrix(4, 4)
	m.set(1, 1, 1)
	m.set(3, 2, -1)

	m.clear(-1)

	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if m.get(x, y) != -1 {
				t.Errorf("after clear(-1): get(%d,%d) = %d, want -1", x, y, m.get(x, y))
			}
		}
	}

	m.clear(0)

	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if m.get(x, y) != 0 {
				t.Errorf("after clear(0): get(%d,%d) = %d, want 0", x, y, m.get(x, y))
			}
		}
	}
}

// TestQRByteMatrix_AsymmetricDimensions verifies correct handling of non-square
// matrices (width != height).
func TestQRByteMatrix_AsymmetricDimensions(t *testing.T) {
	m := newQRByteMatrix(10, 3) // 10 wide, 3 tall
	m.set(9, 2, 1)
	if m.get(9, 2) != 1 {
		t.Errorf("get(9,2) = %d, want 1", m.get(9, 2))
	}
	m.set(0, 0, -1)
	if m.get(0, 0) != -1 {
		t.Errorf("get(0,0) = %d, want -1", m.get(0, 0))
	}
}

// ── blockPair (inline struct) ─────────────────────────────────────────────────

// TestQRInterleave_BlockPair_DataAndEC verifies that qrInterleave correctly
// builds blockPair structs with data and EC fields corresponding to C#
// BlockPair.DataBytes / BlockPair.ErrorCorrectionBytes.
// We test indirectly via qrInterleave output byte count and correctness.
func TestQRInterleave_BlockPair_DataAndEC(t *testing.T) {
	// Build a minimal valid bit vector of numDataBytes bytes.
	// Use version 1-L: numTotalBytes=26, numDataBytes=19, numRSBlocks=1.
	numTotalBytes := 26
	numDataBytes := 19
	numRSBlocks := 1

	gf := newQRGF256()
	bits := newQRBitVector()
	for i := 0; i < numDataBytes; i++ {
		bits.appendBits(i+1, 8)
	}
	// Pad to full byte range for the interleave function.
	for bits.sizeInBytes() < numDataBytes {
		bits.appendBits(0, 8)
	}

	result, err := qrInterleave(bits, numTotalBytes, numDataBytes, numRSBlocks, gf)
	if err != nil {
		t.Fatalf("qrInterleave: %v", err)
	}
	if result.sizeInBytes() != numTotalBytes {
		t.Errorf("interleaved bytes = %d, want %d", result.sizeInBytes(), numTotalBytes)
	}
	// The first 19 bytes in the result should match the original data bytes
	// (single block, so no interleaving required).
	for i := 0; i < numDataBytes; i++ {
		got := int(result.data[i]) & 0xff
		want := i + 1
		if got != want {
			t.Errorf("result byte[%d] = %d, want %d", i, got, want)
		}
	}
}

// ── qrBitVector integration: appendBit across byte boundary ──────────────────

// TestQRBitVector_CrossByteBoundary verifies that bits are correctly packed
// when crossing byte boundaries (the appendBit expansion path in C# is:
// if numBitsInLastByte == 0: appendByte(0); sizeInBits -= 8; then OR the bit).
func TestQRBitVector_CrossByteBoundary(t *testing.T) {
	bv := newQRBitVector()
	// Fill first byte completely: 10101010 = 0xAA
	for i := 0; i < 8; i++ {
		bv.appendBit((i + 1) & 1) // bits: 1,0,1,0,1,0,1,0
	}
	// Now append one more bit (1) — crosses into byte 1.
	bv.appendBit(1)
	if bv.size() != 9 {
		t.Fatalf("size = %d, want 9", bv.size())
	}
	if bv.data[0] != 0xAA {
		t.Errorf("data[0] = 0x%02X, want 0xAA", bv.data[0])
	}
	// The 9th bit (index 8) is 1, stored at MSB of data[1] = 0x80.
	if bv.data[1] != 0x80 {
		t.Errorf("data[1] = 0x%02X, want 0x80", bv.data[1])
	}
}
