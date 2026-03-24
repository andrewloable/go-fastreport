// qr.go implements a pure-Go QR Code encoder, ported from the ZXing-derived
// C# implementation in original-dotnet/FastReport.Base/Barcode/QRCode/.
//
// The encoder follows the JISX0510:2004 / ISO 18004 standard:
//   - Mode selection: Numeric, Alphanumeric, Byte (default), Kanji
//   - Error correction levels: L, M, Q, H
//   - Reed-Solomon error correction over GF(256) (primitive 0x011D)
//   - Format and version information encoding
//   - Data/ECC byte interleaving for multi-block versions
//   - All 8 mask patterns with penalty scoring
//   - Matrix construction: finder patterns, alignment patterns, timing patterns
//
// The public entry point is encodeQR which returns a [][]bool module matrix.
package barcode

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ── QR Error Correction Level ─────────────────────────────────────────────────

// qrECLevel represents a QR error-correction level (L/M/Q/H).
type qrECLevel struct {
	ordinal int // 0=L, 1=M, 2=Q, 3=H
	bits    int // 2-bit value encoded in format information
}

var (
	qrECL = qrECLevel{0, 0x01} // L – ~7% correction
	qrECM = qrECLevel{1, 0x00} // M – ~15% correction
	qrECQ = qrECLevel{2, 0x03} // Q – ~25% correction
	qrECH = qrECLevel{3, 0x02} // H – ~30% correction
)

func qrECLevelFromString(s string) qrECLevel {
	switch s {
	case "L":
		return qrECL
	case "Q":
		return qrECQ
	case "H":
		return qrECH
	default:
		return qrECM
	}
}

// ── QR Mode ───────────────────────────────────────────────────────────────────

type qrMode int

const (
	qrModeNumeric      qrMode = 1
	qrModeAlphanumeric qrMode = 2
	qrModeByte         qrMode = 4
	qrModeKanji        qrMode = 8
)

// characterCountBits returns the number of bits used to encode the character
// count for a given mode and QR version (1-9, 10-26, 27-40).
func (m qrMode) characterCountBits(version int) int {
	type counts [3]int
	table := map[qrMode]counts{
		qrModeNumeric:      {10, 12, 14},
		qrModeAlphanumeric: {9, 11, 13},
		qrModeByte:         {8, 16, 16},
		qrModeKanji:        {8, 10, 12},
	}
	c := table[m]
	switch {
	case version <= 9:
		return c[0]
	case version <= 26:
		return c[1]
	default:
		return c[2]
	}
}

// ── Alphanumeric table ────────────────────────────────────────────────────────

var qrAlphanumericTable = [96]int{
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, // 0-15
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, // 16-31
	36, -1, -1, -1, 37, 38, -1, -1, -1, -1, 39, 40, -1, 41, 42, 43, // 32-47 (SP, $, %, *, +, -, ., /)
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 44, -1, -1, -1, -1, -1, // 48-63 (digits, :)
	-1, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, // 64-79 (A-O)
	25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, -1, -1, -1, -1, -1, // 80-95 (P-Z)
}

func qrGetAlphanumericCode(c rune) int {
	if c >= 0 && c < rune(len(qrAlphanumericTable)) {
		return qrAlphanumericTable[c]
	}
	return -1
}

// ── Mode selection ────────────────────────────────────────────────────────────

func qrChooseMode(content string) qrMode {
	hasNumeric := false
	hasAlphanumeric := false
	for _, c := range content {
		if c >= '0' && c <= '9' {
			hasNumeric = true
		} else if qrGetAlphanumericCode(c) != -1 {
			hasAlphanumeric = true
		} else {
			return qrModeByte
		}
	}
	if hasAlphanumeric {
		return qrModeAlphanumeric
	}
	if hasNumeric {
		return qrModeNumeric
	}
	return qrModeByte
}

// ── BitVector ─────────────────────────────────────────────────────────────────

// qrBitVector is a dynamic bit buffer, MSB-first.
type qrBitVector struct {
	data       []byte
	sizeInBits int
}

func newQRBitVector() *qrBitVector {
	return &qrBitVector{data: make([]byte, 32)}
}

func (bv *qrBitVector) size() int { return bv.sizeInBits }

func (bv *qrBitVector) sizeInBytes() int { return (bv.sizeInBits + 7) >> 3 }

func (bv *qrBitVector) at(index int) int {
	v := int(bv.data[index>>3]) & 0xff
	return (v >> (7 - (index & 0x7))) & 1
}

func (bv *qrBitVector) appendBit(bit int) {
	numBitsInLastByte := bv.sizeInBits & 0x7
	if numBitsInLastByte == 0 {
		bv.appendByte(0)
		bv.sizeInBits -= 8
	}
	bv.data[bv.sizeInBits>>3] |= byte(bit << (7 - numBitsInLastByte))
	bv.sizeInBits++
}

func (bv *qrBitVector) appendBits(value, numBits int) {
	numBitsLeft := numBits
	for numBitsLeft > 0 {
		if (bv.sizeInBits&0x7) == 0 && numBitsLeft >= 8 {
			newByte := (value >> (numBitsLeft - 8)) & 0xff
			bv.appendByte(newByte)
			numBitsLeft -= 8
		} else {
			bit := (value >> (numBitsLeft - 1)) & 1
			bv.appendBit(bit)
			numBitsLeft--
		}
	}
}

func (bv *qrBitVector) appendBitVector(other *qrBitVector) {
	for i := 0; i < other.size(); i++ {
		bv.appendBit(other.at(i))
	}
}

func (bv *qrBitVector) xorWith(other *qrBitVector) {
	sizeBytes := (bv.sizeInBits + 7) >> 3
	for i := 0; i < sizeBytes; i++ {
		bv.data[i] ^= other.data[i]
	}
}

func (bv *qrBitVector) appendByte(value int) {
	if (bv.sizeInBits >> 3) == len(bv.data) {
		newData := make([]byte, len(bv.data)<<1)
		copy(newData, bv.data)
		bv.data = newData
	}
	bv.data[bv.sizeInBits>>3] = byte(value)
	bv.sizeInBits += 8
}

// ── ByteMatrix ────────────────────────────────────────────────────────────────

// qrByteMatrix is a 2-D grid of int8 values. -1 = empty, 0 = light, 1 = dark.
type qrByteMatrix struct {
	bytes  [][]int8
	width  int
	height int
}

func newQRByteMatrix(w, h int) *qrByteMatrix {
	b := make([][]int8, h)
	for i := range b {
		b[i] = make([]int8, w)
	}
	return &qrByteMatrix{bytes: b, width: w, height: h}
}

func (m *qrByteMatrix) get(x, y int) int8 { return m.bytes[y][x] }
func (m *qrByteMatrix) set(x, y, v int)   { m.bytes[y][x] = int8(v) }
func (m *qrByteMatrix) clear(v int8) {
	for y := range m.bytes {
		for x := range m.bytes[y] {
			m.bytes[y][x] = v
		}
	}
}

// ── GF(256) field ─────────────────────────────────────────────────────────────

// qrGF256 holds the log and anti-log tables for GF(256) with primitive 0x011D.
type qrGF256 struct {
	expTable [256]int
	logTable [256]int
}

func newQRGF256() *qrGF256 {
	gf := &qrGF256{}
	x := 1
	for i := 0; i < 256; i++ {
		gf.expTable[i] = x
		x <<= 1
		if x >= 0x100 {
			x ^= 0x011D
		}
	}
	for i := 0; i < 255; i++ {
		gf.logTable[gf.expTable[i]] = i
	}
	return gf
}

func (gf *qrGF256) multiply(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return gf.expTable[(gf.logTable[a]+gf.logTable[b])%255]
}

func (gf *qrGF256) exp(a int) int { return gf.expTable[a] }

// ── Reed-Solomon polynomial over GF(256) ─────────────────────────────────────

// qrGF256Poly represents a polynomial with GF(256) coefficients.
// Coefficients are stored MSB-first: coefficients[0] is the leading term.
type qrGF256Poly struct {
	gf           *qrGF256
	coefficients []int
}

func newQRGF256Poly(gf *qrGF256, coefficients []int) *qrGF256Poly {
	if len(coefficients) > 1 && coefficients[0] == 0 {
		first := 0
		for first < len(coefficients) && coefficients[first] == 0 {
			first++
		}
		if first == len(coefficients) {
			return &qrGF256Poly{gf: gf, coefficients: []int{0}}
		}
		coefficients = coefficients[first:]
	}
	c := make([]int, len(coefficients))
	copy(c, coefficients)
	return &qrGF256Poly{gf: gf, coefficients: c}
}

func (p *qrGF256Poly) degree() int  { return len(p.coefficients) - 1 }
func (p *qrGF256Poly) isZero() bool { return p.coefficients[0] == 0 }
func (p *qrGF256Poly) getCoefficient(degree int) int {
	return p.coefficients[len(p.coefficients)-1-degree]
}

func (p *qrGF256Poly) addOrSubtract(other *qrGF256Poly) *qrGF256Poly {
	if p.isZero() {
		return other
	}
	if other.isZero() {
		return p
	}
	small := p.coefficients
	large := other.coefficients
	if len(small) > len(large) {
		small, large = large, small
	}
	result := make([]int, len(large))
	diff := len(large) - len(small)
	copy(result, large[:diff])
	for i := diff; i < len(large); i++ {
		result[i] = small[i-diff] ^ large[i]
	}
	return newQRGF256Poly(p.gf, result)
}

func (p *qrGF256Poly) multiply(other *qrGF256Poly) *qrGF256Poly {
	if p.isZero() || other.isZero() {
		return newQRGF256Poly(p.gf, []int{0})
	}
	a := p.coefficients
	b := other.coefficients
	product := make([]int, len(a)+len(b)-1)
	for i, ac := range a {
		for j, bc := range b {
			product[i+j] ^= p.gf.multiply(ac, bc)
		}
	}
	return newQRGF256Poly(p.gf, product)
}

func (p *qrGF256Poly) multiplyByMonomial(degree, coefficient int) *qrGF256Poly {
	if coefficient == 0 {
		return newQRGF256Poly(p.gf, []int{0})
	}
	product := make([]int, len(p.coefficients)+degree)
	for i, c := range p.coefficients {
		product[i] = p.gf.multiply(c, coefficient)
	}
	return newQRGF256Poly(p.gf, product)
}

// divide returns [quotient, remainder].
func (p *qrGF256Poly) divide(other *qrGF256Poly) (*qrGF256Poly, *qrGF256Poly) {
	zero := newQRGF256Poly(p.gf, []int{0})
	quotient := zero
	remainder := p

	denomLeading := other.getCoefficient(other.degree())
	// inverse of denomLeading in GF(256): gf.expTable[255 - gf.logTable[x]]
	invDenomLeading := p.gf.expTable[255-p.gf.logTable[denomLeading]]

	for remainder.degree() >= other.degree() && !remainder.isZero() {
		degDiff := remainder.degree() - other.degree()
		scale := p.gf.multiply(remainder.getCoefficient(remainder.degree()), invDenomLeading)
		term := other.multiplyByMonomial(degDiff, scale)
		iterQuotient := newQRGF256Poly(p.gf, func() []int {
			c := make([]int, degDiff+1)
			c[0] = scale
			return c
		}())
		quotient = quotient.addOrSubtract(iterQuotient)
		remainder = remainder.addOrSubtract(term)
	}
	return quotient, remainder
}

// ── Reed-Solomon encoder ──────────────────────────────────────────────────────

// qrReedSolomon encodes toEncode[0:dataBytes] and writes ecBytes EC bytes at
// toEncode[dataBytes:dataBytes+ecBytes].
type qrReedSolomon struct {
	gf               *qrGF256
	cachedGenerators []*qrGF256Poly
}

func newQRReedSolomon(gf *qrGF256) *qrReedSolomon {
	rs := &qrReedSolomon{gf: gf}
	rs.cachedGenerators = append(rs.cachedGenerators, newQRGF256Poly(gf, []int{1}))
	return rs
}

func (rs *qrReedSolomon) buildGenerator(degree int) *qrGF256Poly {
	for len(rs.cachedGenerators) <= degree {
		last := rs.cachedGenerators[len(rs.cachedGenerators)-1]
		d := len(rs.cachedGenerators)
		next := last.multiply(newQRGF256Poly(rs.gf, []int{1, rs.gf.exp(d - 1)}))
		rs.cachedGenerators = append(rs.cachedGenerators, next)
	}
	return rs.cachedGenerators[degree]
}

func (rs *qrReedSolomon) encode(toEncode []int, ecBytes int) {
	dataBytes := len(toEncode) - ecBytes
	generator := rs.buildGenerator(ecBytes)
	infoCoeffs := make([]int, dataBytes)
	copy(infoCoeffs, toEncode[:dataBytes])
	info := newQRGF256Poly(rs.gf, infoCoeffs)
	info = info.multiplyByMonomial(ecBytes, 1)
	_, remainder := info.divide(generator)
	coeffs := remainder.coefficients
	numZero := ecBytes - len(coeffs)
	for i := 0; i < numZero; i++ {
		toEncode[dataBytes+i] = 0
	}
	copy(toEncode[dataBytes+numZero:], coeffs)
}

// ── QR Version data tables ────────────────────────────────────────────────────

// qrVersionInfo holds capacity information for one QR version and one EC level.
type qrVersionInfo struct {
	totalCodewords int
	ecCodewords    int // EC codewords per block (same for all blocks in a version/level)
	blocks         []qrBlock
}

// qrBlock describes one Reed-Solomon block group.
type qrBlock struct {
	count        int
	dataPerBlock int
}

// numDataCodewords returns total data codewords across all blocks.
func (v qrVersionInfo) numDataCodewords() int {
	n := 0
	for _, b := range v.blocks {
		n += b.count * b.dataPerBlock
	}
	return n
}

// numBlocks returns total block count.
func (v qrVersionInfo) numBlocks() int {
	n := 0
	for _, b := range v.blocks {
		n += b.count
	}
	return n
}

// ecPerBlock returns EC codewords per block.
// ecCodewords is stored as the per-block value (matching C# ECBlocks.ECCodewordsPerBlock).
// Returns 0 when blocks is empty (degenerate/uninitialised entry).
func (v qrVersionInfo) ecPerBlock() int {
	if len(v.blocks) == 0 {
		return 0
	}
	return v.ecCodewords
}

// qrVersionTable[i] is version i+1 info [L, M, Q, H].
// Source: ISO 18004:2006 Table 9.
var qrVersionTable [40][4]qrVersionInfo

func init() {
	type ec = qrBlock
	type vrow = [4]qrVersionInfo
	// Version 1
	qrVersionTable[0] = vrow{
		{26, 7, []ec{{1, 19}}},
		{26, 10, []ec{{1, 16}}},
		{26, 13, []ec{{1, 13}}},
		{26, 17, []ec{{1, 9}}},
	}
	// Version 2
	qrVersionTable[1] = vrow{
		{44, 10, []ec{{1, 34}}},
		{44, 16, []ec{{1, 28}}},
		{44, 22, []ec{{1, 22}}},
		{44, 28, []ec{{1, 16}}},
	}
	// Version 3
	qrVersionTable[2] = vrow{
		{70, 15, []ec{{1, 55}}},
		{70, 26, []ec{{1, 44}}},
		{70, 18, []ec{{2, 17}}},
		{70, 22, []ec{{2, 13}}},
	}
	// Version 4
	qrVersionTable[3] = vrow{
		{100, 20, []ec{{1, 80}}},
		{100, 18, []ec{{2, 32}}},
		{100, 26, []ec{{2, 24}}},
		{100, 16, []ec{{4, 9}}},
	}
	// Version 5
	qrVersionTable[4] = vrow{
		{134, 26, []ec{{1, 108}}},
		{134, 24, []ec{{2, 43}}},
		{134, 18, []ec{{2, 15}, {2, 16}}},
		{134, 22, []ec{{2, 11}, {2, 12}}},
	}
	// Version 6
	qrVersionTable[5] = vrow{
		{172, 18, []ec{{2, 68}}},
		{172, 16, []ec{{4, 27}}},
		{172, 24, []ec{{4, 19}}},
		{172, 28, []ec{{4, 15}}},
	}
	// Version 7
	qrVersionTable[6] = vrow{
		{196, 20, []ec{{2, 78}}},
		{196, 18, []ec{{4, 31}}},
		{196, 18, []ec{{2, 14}, {4, 15}}},
		{196, 26, []ec{{4, 13}, {1, 14}}},
	}
	// Version 8
	qrVersionTable[7] = vrow{
		{242, 24, []ec{{2, 97}}},
		{242, 22, []ec{{2, 38}, {2, 39}}},
		{242, 22, []ec{{4, 18}, {2, 19}}},
		{242, 26, []ec{{4, 14}, {2, 15}}},
	}
	// Version 9
	qrVersionTable[8] = vrow{
		{292, 30, []ec{{2, 116}}},
		{292, 22, []ec{{3, 36}, {2, 37}}},
		{292, 20, []ec{{4, 16}, {4, 17}}},
		{292, 24, []ec{{4, 12}, {4, 13}}},
	}
	// Version 10
	qrVersionTable[9] = vrow{
		{346, 18, []ec{{2, 68}, {2, 69}}},
		{346, 26, []ec{{4, 43}, {1, 44}}},
		{346, 24, []ec{{6, 19}, {2, 20}}},
		{346, 28, []ec{{6, 15}, {2, 16}}},
	}
	// Version 11
	qrVersionTable[10] = vrow{
		{404, 20, []ec{{4, 81}}},
		{404, 30, []ec{{1, 50}, {4, 51}}},
		{404, 28, []ec{{4, 22}, {4, 23}}},
		{404, 24, []ec{{3, 12}, {8, 13}}},
	}
	// Version 12
	qrVersionTable[11] = vrow{
		{466, 24, []ec{{2, 92}, {2, 93}}},
		{466, 22, []ec{{6, 36}, {2, 37}}},
		{466, 26, []ec{{4, 20}, {6, 21}}},
		{466, 28, []ec{{7, 14}, {4, 15}}},
	}
	// Version 13
	qrVersionTable[12] = vrow{
		{532, 26, []ec{{4, 107}}},
		{532, 22, []ec{{8, 37}, {1, 38}}},
		{532, 24, []ec{{8, 20}, {4, 21}}},
		{532, 22, []ec{{12, 11}, {4, 12}}},
	}
	// Version 14
	qrVersionTable[13] = vrow{
		{581, 30, []ec{{3, 115}, {1, 116}}},
		{581, 24, []ec{{4, 40}, {5, 41}}},
		{581, 20, []ec{{11, 16}, {5, 17}}},
		{581, 24, []ec{{11, 12}, {5, 13}}},
	}
	// Version 15
	qrVersionTable[14] = vrow{
		{655, 22, []ec{{5, 87}, {1, 88}}},
		{655, 24, []ec{{5, 41}, {5, 42}}},
		{655, 30, []ec{{5, 24}, {7, 25}}},
		{655, 24, []ec{{11, 12}, {7, 13}}},
	}
	// Version 16
	qrVersionTable[15] = vrow{
		{733, 24, []ec{{5, 98}, {1, 99}}},
		{733, 28, []ec{{7, 45}, {3, 46}}},
		{733, 24, []ec{{15, 19}, {2, 20}}},
		{733, 30, []ec{{3, 15}, {13, 16}}},
	}
	// Version 17
	qrVersionTable[16] = vrow{
		{815, 28, []ec{{1, 107}, {5, 108}}},
		{815, 28, []ec{{10, 46}, {1, 47}}},
		{815, 28, []ec{{1, 22}, {15, 23}}},
		{815, 28, []ec{{2, 14}, {17, 15}}},
	}
	// Version 18
	qrVersionTable[17] = vrow{
		{901, 30, []ec{{5, 120}, {1, 121}}},
		{901, 26, []ec{{9, 43}, {4, 44}}},
		{901, 28, []ec{{17, 22}, {1, 23}}},
		{901, 28, []ec{{2, 14}, {19, 15}}},
	}
	// Version 19
	qrVersionTable[18] = vrow{
		{991, 28, []ec{{3, 113}, {4, 114}}},
		{991, 26, []ec{{3, 44}, {11, 45}}},
		{991, 26, []ec{{17, 21}, {4, 22}}},
		{991, 26, []ec{{9, 13}, {16, 14}}},
	}
	// Version 20
	qrVersionTable[19] = vrow{
		{1085, 28, []ec{{3, 107}, {5, 108}}},
		{1085, 26, []ec{{3, 41}, {13, 42}}},
		{1085, 30, []ec{{15, 24}, {5, 25}}},
		{1085, 28, []ec{{15, 15}, {10, 16}}},
	}
	// Version 21
	qrVersionTable[20] = vrow{
		{1156, 28, []ec{{4, 116}, {4, 117}}},
		{1156, 26, []ec{{17, 42}}},
		{1156, 28, []ec{{17, 22}, {6, 23}}},
		{1156, 30, []ec{{19, 16}, {6, 17}}},
	}
	// Version 22
	qrVersionTable[21] = vrow{
		{1258, 28, []ec{{2, 111}, {7, 112}}},
		{1258, 28, []ec{{17, 46}}},
		{1258, 30, []ec{{7, 24}, {16, 25}}},
		{1258, 24, []ec{{34, 13}}},
	}
	// Version 23
	qrVersionTable[22] = vrow{
		{1364, 30, []ec{{4, 121}, {5, 122}}},
		{1364, 28, []ec{{4, 47}, {14, 48}}},
		{1364, 30, []ec{{11, 24}, {14, 25}}},
		{1364, 30, []ec{{16, 15}, {14, 16}}},
	}
	// Version 24
	qrVersionTable[23] = vrow{
		{1474, 30, []ec{{6, 117}, {4, 118}}},
		{1474, 28, []ec{{6, 45}, {14, 46}}},
		{1474, 30, []ec{{11, 24}, {16, 25}}},
		{1474, 30, []ec{{30, 16}, {2, 17}}},
	}
	// Version 25
	qrVersionTable[24] = vrow{
		{1588, 26, []ec{{8, 106}, {4, 107}}},
		{1588, 28, []ec{{8, 47}, {13, 48}}},
		{1588, 30, []ec{{7, 24}, {22, 25}}},
		{1588, 30, []ec{{22, 15}, {13, 16}}},
	}
	// Version 26
	qrVersionTable[25] = vrow{
		{1706, 28, []ec{{10, 114}, {2, 115}}},
		{1706, 28, []ec{{19, 46}, {4, 47}}},
		{1706, 28, []ec{{28, 22}, {6, 23}}},
		{1706, 30, []ec{{33, 16}, {4, 17}}},
	}
	// Version 27
	qrVersionTable[26] = vrow{
		{1828, 30, []ec{{8, 122}, {4, 123}}},
		{1828, 28, []ec{{22, 45}, {3, 46}}},
		{1828, 30, []ec{{8, 23}, {26, 24}}},
		{1828, 30, []ec{{12, 15}, {28, 16}}},
	}
	// Version 28
	qrVersionTable[27] = vrow{
		{1921, 30, []ec{{3, 117}, {10, 118}}},
		{1921, 28, []ec{{3, 45}, {23, 46}}},
		{1921, 30, []ec{{4, 24}, {31, 25}}},
		{1921, 30, []ec{{11, 15}, {31, 16}}},
	}
	// Version 29
	qrVersionTable[28] = vrow{
		{2051, 30, []ec{{7, 116}, {7, 117}}},
		{2051, 28, []ec{{21, 45}, {7, 46}}},
		{2051, 30, []ec{{1, 23}, {37, 24}}},
		{2051, 30, []ec{{19, 15}, {26, 16}}},
	}
	// Version 30
	qrVersionTable[29] = vrow{
		{2185, 30, []ec{{5, 115}, {10, 116}}},
		{2185, 28, []ec{{19, 47}, {10, 48}}},
		{2185, 30, []ec{{15, 24}, {25, 25}}},
		{2185, 30, []ec{{23, 15}, {25, 16}}},
	}
	// Version 31
	qrVersionTable[30] = vrow{
		{2323, 30, []ec{{13, 115}, {3, 116}}},
		{2323, 28, []ec{{2, 46}, {29, 47}}},
		{2323, 30, []ec{{42, 24}, {1, 25}}},
		{2323, 30, []ec{{23, 15}, {28, 16}}},
	}
	// Version 32
	qrVersionTable[31] = vrow{
		{2465, 30, []ec{{17, 115}}},
		{2465, 28, []ec{{10, 46}, {23, 47}}},
		{2465, 30, []ec{{10, 24}, {35, 25}}},
		{2465, 30, []ec{{19, 15}, {35, 16}}},
	}
	// Version 33
	qrVersionTable[32] = vrow{
		{2611, 30, []ec{{17, 115}, {1, 116}}},
		{2611, 28, []ec{{14, 46}, {21, 47}}},
		{2611, 30, []ec{{29, 24}, {19, 25}}},
		{2611, 30, []ec{{11, 15}, {46, 16}}},
	}
	// Version 34
	qrVersionTable[33] = vrow{
		{2761, 30, []ec{{13, 115}, {6, 116}}},
		{2761, 28, []ec{{14, 46}, {23, 47}}},
		{2761, 30, []ec{{44, 24}, {7, 25}}},
		{2761, 30, []ec{{59, 16}, {1, 17}}},
	}
	// Version 35
	qrVersionTable[34] = vrow{
		{2876, 30, []ec{{12, 121}, {7, 122}}},
		{2876, 28, []ec{{12, 47}, {26, 48}}},
		{2876, 30, []ec{{39, 24}, {14, 25}}},
		{2876, 30, []ec{{22, 15}, {41, 16}}},
	}
	// Version 36
	qrVersionTable[35] = vrow{
		{3034, 30, []ec{{6, 121}, {14, 122}}},
		{3034, 28, []ec{{6, 47}, {34, 48}}},
		{3034, 30, []ec{{46, 24}, {10, 25}}},
		{3034, 30, []ec{{2, 15}, {64, 16}}},
	}
	// Version 37
	qrVersionTable[36] = vrow{
		{3196, 30, []ec{{17, 122}, {4, 123}}},
		{3196, 28, []ec{{29, 46}, {14, 47}}},
		{3196, 30, []ec{{49, 24}, {10, 25}}},
		{3196, 30, []ec{{24, 15}, {46, 16}}},
	}
	// Version 38
	qrVersionTable[37] = vrow{
		{3362, 30, []ec{{4, 122}, {18, 123}}},
		{3362, 28, []ec{{13, 46}, {32, 47}}},
		{3362, 30, []ec{{48, 24}, {14, 25}}},
		{3362, 30, []ec{{42, 15}, {32, 16}}},
	}
	// Version 39
	qrVersionTable[38] = vrow{
		{3532, 30, []ec{{20, 117}, {4, 118}}},
		{3532, 28, []ec{{40, 47}, {7, 48}}},
		{3532, 30, []ec{{43, 24}, {22, 25}}},
		{3532, 30, []ec{{10, 15}, {67, 16}}},
	}
	// Version 40
	qrVersionTable[39] = vrow{
		{3706, 30, []ec{{19, 118}, {6, 119}}},
		{3706, 28, []ec{{18, 47}, {31, 48}}},
		{3706, 30, []ec{{34, 24}, {34, 25}}},
		{3706, 30, []ec{{20, 15}, {61, 16}}},
	}
}

// qrDimension returns the module dimension for a QR version: 17 + 4*version.
func qrDimension(version int) int { return 17 + 4*version }

// ── Alignment pattern center coordinates ─────────────────────────────────────

// qrAlignmentCenters returns the alignment pattern center coordinates for version.
// These match Version.cs POSITION_ADJUSTMENT_PATTERN_COORDINATE_TABLE.
var qrAlignmentCentersTable = [40][]int{
	{},                             // v1
	{6, 18},                        // v2
	{6, 22},                        // v3
	{6, 26},                        // v4
	{6, 30},                        // v5
	{6, 34},                        // v6
	{6, 22, 38},                    // v7
	{6, 24, 42},                    // v8
	{6, 26, 46},                    // v9
	{6, 28, 50},                    // v10
	{6, 30, 54},                    // v11
	{6, 32, 58},                    // v12
	{6, 34, 62},                    // v13
	{6, 26, 46, 66},                // v14
	{6, 26, 48, 70},                // v15
	{6, 26, 50, 74},                // v16
	{6, 30, 54, 78},                // v17
	{6, 30, 56, 82},                // v18
	{6, 30, 58, 86},                // v19
	{6, 34, 62, 90},                // v20
	{6, 28, 50, 72, 94},            // v21
	{6, 26, 50, 74, 98},            // v22
	{6, 30, 54, 78, 102},           // v23
	{6, 28, 54, 80, 106},           // v24
	{6, 32, 58, 84, 110},           // v25
	{6, 30, 58, 86, 114},           // v26
	{6, 34, 62, 90, 118},           // v27
	{6, 26, 50, 74, 98, 122},       // v28
	{6, 30, 54, 78, 102, 126},      // v29
	{6, 26, 52, 78, 104, 130},      // v30
	{6, 30, 56, 82, 108, 134},      // v31
	{6, 34, 60, 86, 112, 138},      // v32
	{6, 30, 58, 86, 114, 142},      // v33
	{6, 34, 62, 90, 118, 146},      // v34
	{6, 30, 54, 78, 102, 126, 150}, // v35
	{6, 24, 50, 76, 102, 128, 154}, // v36
	{6, 28, 54, 80, 106, 132, 158}, // v37
	{6, 32, 58, 84, 110, 136, 162}, // v38
	{6, 26, 54, 82, 110, 138, 166}, // v39
	{6, 30, 58, 86, 114, 142, 170}, // v40
}

// ── Matrix construction helpers ───────────────────────────────────────────────

// Static patterns.
var qrFinderPattern = [7][7]int{
	{1, 1, 1, 1, 1, 1, 1},
	{1, 0, 0, 0, 0, 0, 1},
	{1, 0, 1, 1, 1, 0, 1},
	{1, 0, 1, 1, 1, 0, 1},
	{1, 0, 1, 1, 1, 0, 1},
	{1, 0, 0, 0, 0, 0, 1},
	{1, 1, 1, 1, 1, 1, 1},
}

var qrAlignmentPattern = [5][5]int{
	{1, 1, 1, 1, 1},
	{1, 0, 0, 0, 1},
	{1, 0, 1, 0, 1},
	{1, 0, 0, 0, 1},
	{1, 1, 1, 1, 1},
}

// Type info coordinates (left-top corner and surrounding).
var qrTypeInfoCoords = [15][2]int{
	{8, 0}, {8, 1}, {8, 2}, {8, 3}, {8, 4}, {8, 5}, {8, 7}, {8, 8},
	{7, 8}, {5, 8}, {4, 8}, {3, 8}, {2, 8}, {1, 8}, {0, 8},
}

const (
	qrVersionInfoPoly = 0x1f25 // 1 1111 0010 0101
	qrTypeInfoPoly    = 0x537
	qrTypeInfoMask    = 0x5412
)

// findMSBSet returns the 1-based position of the highest set bit.
func qrFindMSBSet(v int) int {
	n := 0
	for v != 0 {
		v >>= 1
		n++
	}
	return n
}

// calculateBCHCode computes the BCH remainder of value divided by poly.
func qrCalcBCH(value, poly int) int {
	msb := qrFindMSBSet(poly)
	value <<= msb - 1
	for qrFindMSBSet(value) >= msb {
		value ^= poly << (qrFindMSBSet(value) - msb)
	}
	return value
}

// embedFinderPattern places a 7×7 finder at (xStart, yStart).
func qrEmbedFinder(m *qrByteMatrix, xStart, yStart int) {
	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			m.set(xStart+x, yStart+y, qrFinderPattern[y][x])
		}
	}
}

// embedBasicPatterns places finder patterns, separators, dark module,
// alignment patterns, and timing patterns.
func qrEmbedBasicPatterns(version int, m *qrByteMatrix) {
	w := m.width
	// Finder patterns at three corners.
	qrEmbedFinder(m, 0, 0)
	qrEmbedFinder(m, w-7, 0)
	qrEmbedFinder(m, 0, w-7)

	// Horizontal separators (8 zeros).
	for x := 0; x < 8; x++ {
		m.set(x, 7, 0)     // top-left
		m.set(w-8+x, 7, 0) // top-right
		m.set(x, w-8, 0)   // bottom-left
	}
	// Vertical separators (7 zeros).
	for y := 0; y < 7; y++ {
		m.set(7, y, 0)     // top-left
		m.set(w-8, y, 0)   // top-right
		m.set(7, w-7+y, 0) // bottom-left
	}

	// Dark module (always 1) at column 8, row = size-8.
	m.set(8, w-8, 1)

	// Alignment patterns (version >= 2).
	if version >= 2 {
		coords := qrAlignmentCentersTable[version-1]
		for _, cy := range coords {
			for _, cx := range coords {
				if m.get(cx, cy) == -1 {
					// Place 5×5 alignment pattern centered at (cx, cy).
					for y := 0; y < 5; y++ {
						for x := 0; x < 5; x++ {
							if m.get(cx-2+x, cy-2+y) == -1 {
								m.set(cx-2+x, cy-2+y, qrAlignmentPattern[y][x])
							}
						}
					}
				}
			}
		}
	}

	// Timing patterns (row 6 and column 6, starting from 8 to w-9).
	for i := 8; i < w-8; i++ {
		bit := (i + 1) % 2 // alternating 1,0,1,0...
		if m.get(i, 6) == -1 {
			m.set(i, 6, bit)
		}
		if m.get(6, i) == -1 {
			m.set(6, i, bit)
		}
	}
}

// embedTypeInfo encodes the EC level and mask pattern into the format info bits.
func qrEmbedTypeInfo(ecLevel qrECLevel, maskPattern int, m *qrByteMatrix) {
	bv := newQRBitVector()
	typeInfo := (ecLevel.bits << 3) | maskPattern
	bv.appendBits(typeInfo, 5)
	bchCode := qrCalcBCH(typeInfo, qrTypeInfoPoly)
	bv.appendBits(bchCode, 10)
	// XOR with mask.
	maskBv := newQRBitVector()
	maskBv.appendBits(qrTypeInfoMask, 15)
	bv.xorWith(maskBv)

	w := m.width
	for i := 0; i < bv.size(); i++ {
		bit := bv.at(bv.size() - 1 - i)
		x1, y1 := qrTypeInfoCoords[i][0], qrTypeInfoCoords[i][1]
		m.set(x1, y1, bit)
		if i < 8 {
			// Right-top corner.
			m.set(w-i-1, 8, bit)
		} else {
			// Left-bottom corner.
			m.set(8, w-7+(i-8), bit)
		}
	}
}

// maybeEmbedVersionInfo places version info bits for version >= 7.
func qrMaybeEmbedVersionInfo(version int, m *qrByteMatrix) {
	if version < 7 {
		return
	}
	bv := newQRBitVector()
	bv.appendBits(version, 6)
	bchCode := qrCalcBCH(version, qrVersionInfoPoly)
	bv.appendBits(bchCode, 12)

	w := m.width
	bitIndex := 6*3 - 1
	for i := 0; i < 6; i++ {
		for j := 0; j < 3; j++ {
			bit := bv.at(bitIndex)
			bitIndex--
			m.set(i, w-11+j, bit)
			m.set(w-11+j, i, bit)
		}
	}
}

// qrDataMaskBit returns true if the data bit at (x,y) should be flipped for maskPattern.
func qrDataMaskBit(maskPattern, x, y int) bool {
	var intermediate int
	switch maskPattern {
	case 0:
		intermediate = (y + x) & 0x1
	case 1:
		intermediate = y & 0x1
	case 2:
		intermediate = x % 3
	case 3:
		intermediate = (y + x) % 3
	case 4:
		intermediate = ((y >> 1) + (x / 3)) & 0x1
	case 5:
		tmp := y * x
		intermediate = (tmp & 0x1) + (tmp % 3)
	case 6:
		tmp := y * x
		intermediate = ((tmp & 0x1) + (tmp % 3)) & 0x1
	case 7:
		tmp := y * x
		intermediate = ((tmp % 3) + ((y + x) & 0x1)) & 0x1
	}
	return intermediate == 0
}

// embedDataBits places the final data+EC bits into the matrix using the zig-zag
// pattern specified in JISX0510:2004 section 8.7.
func qrEmbedDataBits(dataBits *qrBitVector, maskPattern int, m *qrByteMatrix) error {
	bitIndex := 0
	direction := -1
	x := m.width - 1
	y := m.height - 1
	for x > 0 {
		if x == 6 {
			x--
		}
		for y >= 0 && y < m.height {
			for i := 0; i < 2; i++ {
				xx := x - i
				if m.get(xx, y) != -1 {
					continue
				}
				var bit int
				if bitIndex < dataBits.size() {
					bit = dataBits.at(bitIndex)
					bitIndex++
				}
				if maskPattern != -1 && qrDataMaskBit(maskPattern, xx, y) {
					bit ^= 1
				}
				m.set(xx, y, bit)
			}
			y += direction
		}
		direction = -direction
		y += direction
		x -= 2
	}
	if bitIndex != dataBits.size() {
		return fmt.Errorf("qr: not all bits consumed: %d/%d", bitIndex, dataBits.size())
	}
	return nil
}

// ── Mask penalty scoring ──────────────────────────────────────────────────────

func qrPenaltyRule1(m *qrByteMatrix) int {
	penalty := 0
	for isHoriz := 0; isHoriz < 2; isHoriz++ {
		var iLimit, jLimit int
		if isHoriz == 1 {
			iLimit, jLimit = m.height, m.width
		} else {
			iLimit, jLimit = m.width, m.height
		}
		for i := 0; i < iLimit; i++ {
			same := 1
			prev := -1
			for j := 0; j < jLimit; j++ {
				var bit int
				if isHoriz == 1 {
					bit = int(m.bytes[i][j])
				} else {
					bit = int(m.bytes[j][i])
				}
				if bit == prev {
					same++
					if same == 5 {
						penalty += 3
					} else if same > 5 {
						penalty++
					}
				} else {
					same = 1
					prev = bit
				}
			}
		}
	}
	return penalty
}

func qrPenaltyRule2(m *qrByteMatrix) int {
	penalty := 0
	for y := 0; y < m.height-1; y++ {
		for x := 0; x < m.width-1; x++ {
			v := m.bytes[y][x]
			if v == m.bytes[y][x+1] && v == m.bytes[y+1][x] && v == m.bytes[y+1][x+1] {
				penalty += 3
			}
		}
	}
	return penalty
}

func qrPenaltyRule3(m *qrByteMatrix) int {
	penalty := 0
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			a := m.bytes[y]
			if x+6 < m.width &&
				a[x] == 1 && a[x+1] == 0 && a[x+2] == 1 && a[x+3] == 1 && a[x+4] == 1 && a[x+5] == 0 && a[x+6] == 1 &&
				((x+10 < m.width && a[x+7] == 0 && a[x+8] == 0 && a[x+9] == 0 && a[x+10] == 0) ||
					(x-4 >= 0 && a[x-1] == 0 && a[x-2] == 0 && a[x-3] == 0 && a[x-4] == 0)) {
				penalty += 40
			}
			if y+6 < m.height {
				col := func(row, col int) int8 { return m.bytes[row][col] }
				if col(y, x) == 1 && col(y+1, x) == 0 && col(y+2, x) == 1 && col(y+3, x) == 1 && col(y+4, x) == 1 && col(y+5, x) == 0 && col(y+6, x) == 1 &&
					((y+10 < m.height && col(y+7, x) == 0 && col(y+8, x) == 0 && col(y+9, x) == 0 && col(y+10, x) == 0) ||
						(y-4 >= 0 && col(y-1, x) == 0 && col(y-2, x) == 0 && col(y-3, x) == 0 && col(y-4, x) == 0)) {
					penalty += 40
				}
			}
		}
	}
	return penalty
}

func qrPenaltyRule4(m *qrByteMatrix) int {
	dark := 0
	total := m.width * m.height
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			if m.bytes[y][x] == 1 {
				dark++
			}
		}
	}
	// Mirror C# MaskUtil.applyMaskPenaltyRule4:
	//   Math.Abs((int)(darkRatio * 100 - 50)) / 5 * 10
	// The subtraction happens in floating-point before the truncating cast,
	// which differs from int(ratio) - 50 when ratio is non-integer.
	// Reference: MaskUtil.cs:118
	darkRatio := float64(dark) / float64(total)
	diff := int(darkRatio*100 - 50)
	if diff < 0 {
		diff = -diff
	}
	return (diff / 5) * 10
}

func qrCalcPenalty(m *qrByteMatrix) int {
	return qrPenaltyRule1(m) + qrPenaltyRule2(m) + qrPenaltyRule3(m) + qrPenaltyRule4(m)
}

// ── Matrix builder ────────────────────────────────────────────────────────────

func qrBuildMatrix(dataBits *qrBitVector, ecLevel qrECLevel, version, maskPattern int, m *qrByteMatrix) error {
	m.clear(-1)
	qrEmbedBasicPatterns(version, m)
	qrEmbedTypeInfo(ecLevel, maskPattern, m)
	qrMaybeEmbedVersionInfo(version, m)
	return qrEmbedDataBits(dataBits, maskPattern, m)
}

func qrChooseMaskPattern(finalBits *qrBitVector, ecLevel qrECLevel, version int, m *qrByteMatrix) (int, error) {
	bestPenalty := int(^uint(0) >> 1)
	bestPattern := -1
	for p := 0; p < 8; p++ {
		if err := qrBuildMatrix(finalBits, ecLevel, version, p, m); err != nil {
			return -1, err
		}
		penalty := qrCalcPenalty(m)
		if penalty < bestPenalty {
			bestPenalty = penalty
			bestPattern = p
		}
	}
	return bestPattern, nil
}

// ── Data encoding ─────────────────────────────────────────────────────────────

func qrAppendModeInfo(mode qrMode, bv *qrBitVector) {
	bv.appendBits(int(mode), 4)
}

func qrAppendLengthInfo(numLetters, version int, mode qrMode, bv *qrBitVector) error {
	numBits := mode.characterCountBits(version)
	maxLetters := (1 << numBits) - 1
	if numLetters > maxLetters {
		return fmt.Errorf("qr: too many characters for version %d mode %v", version, mode)
	}
	bv.appendBits(numLetters, numBits)
	return nil
}

func qrAppendNumericBytes(content string, bv *qrBitVector) {
	i := 0
	n := len(content)
	for i < n {
		d1 := int(content[i] - '0')
		if i+2 < n {
			d2 := int(content[i+1] - '0')
			d3 := int(content[i+2] - '0')
			bv.appendBits(d1*100+d2*10+d3, 10)
			i += 3
		} else if i+1 < n {
			d2 := int(content[i+1] - '0')
			bv.appendBits(d1*10+d2, 7)
			i += 2
		} else {
			bv.appendBits(d1, 4)
			i++
		}
	}
}

func qrAppendAlphanumericBytes(content string, bv *qrBitVector) error {
	runes := []rune(content)
	i := 0
	for i < len(runes) {
		c1 := qrGetAlphanumericCode(runes[i])
		if c1 == -1 {
			return fmt.Errorf("qr: invalid alphanumeric char %q", runes[i])
		}
		if i+1 < len(runes) {
			c2 := qrGetAlphanumericCode(runes[i+1])
			if c2 == -1 {
				return fmt.Errorf("qr: invalid alphanumeric char %q", runes[i+1])
			}
			bv.appendBits(c1*45+c2, 11)
			i += 2
		} else {
			bv.appendBits(c1, 6)
			i++
		}
	}
	return nil
}

func qrAppend8BitBytes(content string, bv *qrBitVector, charset string) {
	if charset == "ISO8859_1" || charset == "ISO-8859-1" {
		// ISO-8859-1: each rune maps to a single byte (truncate to low 8 bits).
		for _, r := range content {
			bv.appendBits(int(r&0xFF), 8)
		}
		return
	}
	// Default: ISO-8859-1 byte encoding for runes < 256, UTF-8 fallback otherwise.
	for _, r := range content {
		if r < 256 {
			bv.appendBits(int(r), 8)
		} else {
			// Fall back to UTF-8 for characters outside Latin-1.
			var buf [utf8.UTFMax]byte
			n := utf8.EncodeRune(buf[:], r)
			for i := 0; i < n; i++ {
				bv.appendBits(int(buf[i]), 8)
			}
		}
	}
}

func qrAppendBytes(content string, mode qrMode, bv *qrBitVector, charset string) error {
	switch mode {
	case qrModeNumeric:
		qrAppendNumericBytes(content, bv)
	case qrModeAlphanumeric:
		return qrAppendAlphanumericBytes(content, bv)
	case qrModeByte:
		qrAppend8BitBytes(content, bv, charset)
	default:
		// Kanji mode not implemented (fallback to Byte).
		qrAppend8BitBytes(content, bv, charset)
	}
	return nil
}

// qrTerminateBits pads dataBits to numDataBytes as per JISX0510:2004 §8.4.8-9.
func qrTerminateBits(numDataBytes int, bv *qrBitVector) error {
	capacity := numDataBytes << 3
	if bv.size() > capacity {
		return fmt.Errorf("qr: data bits exceed capacity")
	}
	// Up to 4 termination zero bits.
	for i := 0; i < 4 && bv.size() < capacity; i++ {
		bv.appendBit(0)
	}
	// Byte-align.
	for bv.size()%8 != 0 {
		bv.appendBit(0)
	}
	// Fill remaining bytes with alternating 0xEC / 0x11.
	numPaddingBytes := numDataBytes - bv.sizeInBytes()
	for i := 0; i < numPaddingBytes; i++ {
		if i%2 == 0 {
			bv.appendBits(0xec, 8)
		} else {
			bv.appendBits(0x11, 8)
		}
	}
	return nil
}

// qrGetBlockDataAndECBytes returns (dataBytes, ecBytes) for blockID in a multi-block setup.
func qrGetBlockSizes(numTotalBytes, numDataBytes, numRSBlocks, blockID int) (int, int) {
	numRSInGroup2 := numTotalBytes % numRSBlocks
	numRSInGroup1 := numRSBlocks - numRSInGroup2
	numTotalInGroup1 := numTotalBytes / numRSBlocks
	numTotalInGroup2 := numTotalInGroup1 + 1
	numDataInGroup1 := numDataBytes / numRSBlocks
	numDataInGroup2 := numDataInGroup1 + 1
	numECInGroup1 := numTotalInGroup1 - numDataInGroup1
	// numECInGroup2 := numTotalInGroup2 - numDataInGroup2 (must equal numECInGroup1)
	_ = numTotalInGroup2
	if blockID < numRSInGroup1 {
		return numDataInGroup1, numECInGroup1
	}
	return numDataInGroup2, numECInGroup1
}

// qrInterleave interleaves data and EC bytes from multiple blocks.
func qrInterleave(bits *qrBitVector, numTotalBytes, numDataBytes, numRSBlocks int, gf *qrGF256) (*qrBitVector, error) {
	rs := newQRReedSolomon(gf)

	type blockPair struct {
		data []int
		ec   []int
	}
	blocks := make([]blockPair, 0, numRSBlocks)
	offset := 0
	maxData := 0
	maxEC := 0

	for i := 0; i < numRSBlocks; i++ {
		dataN, ecN := qrGetBlockSizes(numTotalBytes, numDataBytes, numRSBlocks, i)
		data := make([]int, dataN)
		for j := 0; j < dataN; j++ {
			data[j] = int(bits.data[offset+j]) & 0xff
		}
		toEncode := make([]int, dataN+ecN)
		copy(toEncode, data)
		rs.encode(toEncode, ecN)
		ec := toEncode[dataN:]
		blocks = append(blocks, blockPair{data: data, ec: ec})
		if dataN > maxData {
			maxData = dataN
		}
		if ecN > maxEC {
			maxEC = ecN
		}
		offset += dataN
	}

	result := newQRBitVector()
	for i := 0; i < maxData; i++ {
		for _, bp := range blocks {
			if i < len(bp.data) {
				result.appendBits(bp.data[i], 8)
			}
		}
	}
	for i := 0; i < maxEC; i++ {
		for _, bp := range blocks {
			if i < len(bp.ec) {
				result.appendBits(bp.ec[i], 8)
			}
		}
	}
	if result.sizeInBytes() != numTotalBytes {
		return nil, fmt.Errorf("qr: interleave byte count mismatch %d != %d",
			result.sizeInBytes(), numTotalBytes)
	}
	return result, nil
}

// ── Public entry point ────────────────────────────────────────────────────────

// encodeQR encodes content as a QR code and returns the module matrix.
// Each element matrix[row][col] is true for a dark module.
// The matrix is always square (rows == cols == dimension).
// charset controls byte encoding: "ISO8859_1" or "ISO-8859-1" forces single-byte
// ISO-8859-1 encoding; any other value (including "" and "UTF8") uses default
// UTF-8 encoding. See C# BarcodeQR.cs GetEncoding().
func encodeQR(content string, ecLevel qrECLevel, charset string) ([][]bool, error) {
	if content == "" {
		return nil, fmt.Errorf("qr: content must not be empty")
	}

	gf := newQRGF256()

	// Step 1: Choose mode.
	mode := qrChooseMode(content)

	// Step 2: Encode data bits.
	dataBits := newQRBitVector()
	if err := qrAppendBytes(content, mode, dataBits, charset); err != nil {
		return nil, err
	}
	numInputBytes := dataBits.sizeInBytes()

	// Step 3: Find smallest version that fits.
	version := -1
	var vi qrVersionInfo
	for v := 1; v <= 40; v++ {
		info := qrVersionTable[v-1][ecLevel.ordinal]
		numDataBytes := info.numDataCodewords()
		if numDataBytes >= numInputBytes+3 {
			version = v
			vi = info
			break
		}
	}
	if version == -1 {
		return nil, fmt.Errorf("qr: content too large to encode")
	}

	numDataBytes := vi.numDataCodewords()
	numTotalBytes := vi.totalCodewords
	numRSBlocks := vi.numBlocks()

	// Step 4: Build header + data bit vector.
	headerAndData := newQRBitVector()
	qrAppendModeInfo(mode, headerAndData)
	numLetters := numDataBytes
	if mode != qrModeByte {
		numLetters = len([]rune(content))
	} else {
		// For byte mode, count encoded bytes.
		numLetters = dataBits.sizeInBytes()
	}
	if err := qrAppendLengthInfo(numLetters, version, mode, headerAndData); err != nil {
		return nil, err
	}
	headerAndData.appendBitVector(dataBits)

	// Step 5: Terminate and pad.
	if err := qrTerminateBits(numDataBytes, headerAndData); err != nil {
		return nil, err
	}

	// Step 6: Interleave with EC.
	finalBits, err := qrInterleave(headerAndData, numTotalBytes, numDataBytes, numRSBlocks, gf)
	if err != nil {
		return nil, err
	}

	// Step 7: Choose best mask pattern.
	dim := qrDimension(version)
	m := newQRByteMatrix(dim, dim)
	maskPattern, err := qrChooseMaskPattern(finalBits, ecLevel, version, m)
	if err != nil {
		return nil, err
	}

	// Step 8: Build final matrix.
	if err := qrBuildMatrix(finalBits, ecLevel, version, maskPattern, m); err != nil {
		return nil, err
	}

	// Convert int8 matrix to [][]bool.
	matrix := make([][]bool, dim)
	for row := 0; row < dim; row++ {
		matrix[row] = make([]bool, dim)
		for col := 0; col < dim; col++ {
			matrix[row][col] = m.bytes[row][col] == 1
		}
	}
	return matrix, nil
}

// GetMatrix encodes b.encodedText as a QR code and returns (matrix, rows, cols).
// Implements Matrix2DProvider for QRBarcode.
// When QuietZone is true, a 4-module white border is added per C# BarcodeQR.cs:851.
func (b *QRBarcode) GetMatrix() ([][]bool, int, int) {
	text := b.encodedText
	if text == "" {
		text = b.DefaultValue()
	}
	ecLevel := qrECLevelFromString(b.ErrorCorrection)
	if isSwissQRPayload(text) {
		ecLevel = qrECM
	}
	matrix, err := encodeQR(text, ecLevel, b.Encoding)
	if err != nil || len(matrix) == 0 {
		// Return a 1×1 fallback so callers never receive nil.
		return [][]bool{{true}}, 1, 1
	}
	if b.QuietZone {
		// Add a 4-module quiet zone border (all false/white) around the matrix.
		// C# BarcodeQR.cs:845: quiet = QuietZone ? 4 : 0
		const quiet = 4
		n := len(matrix)
		newSize := n + 2*quiet
		bordered := make([][]bool, newSize)
		for i := range bordered {
			bordered[i] = make([]bool, newSize)
		}
		for r, row := range matrix {
			copy(bordered[r+quiet][quiet:], row)
		}
		return bordered, newSize, newSize
	}
	n := len(matrix)
	return matrix, n, n
}

func isSwissQRPayload(text string) bool {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return strings.HasPrefix(normalized, "SPC")
}

// normalizeSwissQRPayload normalises line endings in a Swiss QR payload to \n
// as required by the Swiss Payment Standards (SPS). The payload structure is
// preserved unchanged — only \r\n → \n and standalone \r → \n are converted.
// Mirrors C# behaviour: the QR encoder receives the text verbatim; Go additionally
// strips \r so the encoded byte count matches the SPS-mandated \n-separated format.
func normalizeSwissQRPayload(text string) string {
	if !isSwissQRPayload(text) {
		return text
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return text
}
