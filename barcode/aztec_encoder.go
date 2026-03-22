// aztec_encoder.go — Full ZXing-compatible Aztec barcode encoder.
//
// This is a direct port of the ZXing.Net Aztec encoder from:
//   original-dotnet/FastReport.Base/Barcode/Aztec/
//
// Ported classes (in order of dependency):
//   GenericGF        (GenericGF.cs)
//   GenericGFPoly    (GenericGFPoly.cs)
//   ReedSolomonEncoder (ReedSolomonEncoder.cs)
//   BitArray         (BitArray.cs)
//   BitMatrix        (BitMatrix.cs)
//   Token/SimpleToken/BinaryShiftToken (Token.cs, SimpleToken.cs, BinaryShiftToken.cs)
//   State            (State.cs)
//   HighLevelEncoder (HighLevelEncoder.cs)
//   Encoder          (Encoder.cs)
//
// The public entry point is encodeAztecFull(data []byte, minECCPercent, userSpecifiedLayers int)
// which returns a [][]bool bit matrix.
package barcode

import "fmt"

// ─────────────────────────────────────────────────────────────────────────────
// aztecGenericGF — GF(2^m) Galois Field
// Ref: GenericGF.cs
// ─────────────────────────────────────────────────────────────────────────────

type aztecGenericGF struct {
	expTable      []int
	logTable      []int
	zero          *aztecGenericGFPoly
	one           *aztecGenericGFPoly
	size          int
	primitive     int
	generatorBase int
}

func newAztecGenericGF(primitive, size, genBase int) *aztecGenericGF {
	gf := &aztecGenericGF{
		primitive:     primitive,
		size:          size,
		generatorBase: genBase,
		expTable:      make([]int, size),
		logTable:      make([]int, size),
	}
	x := 1
	for i := 0; i < size; i++ {
		gf.expTable[i] = x
		x <<= 1
		if x >= size {
			x ^= primitive
			x &= size - 1
		}
	}
	for i := 0; i < size-1; i++ {
		gf.logTable[gf.expTable[i]] = i
	}
	// logTable[0] is 0 but should never be used
	gf.zero = &aztecGenericGFPoly{field: gf, coefficients: []int{0}}
	gf.one = &aztecGenericGFPoly{field: gf, coefficients: []int{1}}
	return gf
}

// Pre-constructed field instances matching C# static fields.
var (
	aztecGFParam    = newAztecGenericGF(0x13, 16, 1)    // x^4+x+1        (AZTEC_PARAM)
	aztecGFData6    = newAztecGenericGF(0x43, 64, 1)    // x^6+x+1        (AZTEC_DATA_6 / MAXICODE_FIELD_64)
	aztecGFData8    = newAztecGenericGF(0x12D, 256, 1)  // x^8+x^5+x^3+x^2+1 (DATA_MATRIX_FIELD_256 / AZTEC_DATA_8)
	aztecGFData10   = newAztecGenericGF(0x409, 1024, 1) // x^10+x^3+1     (AZTEC_DATA_10)
	aztecGFData12   = newAztecGenericGF(0x1069, 4096, 1) // x^12+x^6+x^5+x^3+1 (AZTEC_DATA_12)
)

func (gf *aztecGenericGF) exp(a int) int { return gf.expTable[a] }

func (gf *aztecGenericGF) log(a int) int {
	// panic on 0 to match C# behaviour
	return gf.logTable[a]
}

func (gf *aztecGenericGF) inverse(a int) int {
	return gf.expTable[gf.size-gf.logTable[a]-1]
}

func (gf *aztecGenericGF) multiply(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return gf.expTable[(gf.logTable[a]+gf.logTable[b])%(gf.size-1)]
}

// addOrSubtract is XOR in GF(2^m). Static method in C#.
func aztecGFAddOrSubtract(a, b int) int { return a ^ b }

func (gf *aztecGenericGF) buildMonomial(degree, coefficient int) *aztecGenericGFPoly {
	if coefficient == 0 {
		return gf.zero
	}
	coeffs := make([]int, degree+1)
	coeffs[0] = coefficient
	return newAztecGenericGFPoly(gf, coeffs)
}

// ─────────────────────────────────────────────────────────────────────────────
// aztecGenericGFPoly — polynomial over a GF
// Ref: GenericGFPoly.cs
// ─────────────────────────────────────────────────────────────────────────────

type aztecGenericGFPoly struct {
	field        *aztecGenericGF
	coefficients []int // most significant coefficient first
}

func newAztecGenericGFPoly(field *aztecGenericGF, coefficients []int) *aztecGenericGFPoly {
	if len(coefficients) == 0 {
		panic("aztecGenericGFPoly: empty coefficients")
	}
	coeffLen := len(coefficients)
	if coeffLen > 1 && coefficients[0] == 0 {
		firstNonZero := 1
		for firstNonZero < coeffLen && coefficients[firstNonZero] == 0 {
			firstNonZero++
		}
		if firstNonZero == coeffLen {
			return &aztecGenericGFPoly{field: field, coefficients: []int{0}}
		}
		c := make([]int, coeffLen-firstNonZero)
		copy(c, coefficients[firstNonZero:])
		return &aztecGenericGFPoly{field: field, coefficients: c}
	}
	c := make([]int, len(coefficients))
	copy(c, coefficients)
	return &aztecGenericGFPoly{field: field, coefficients: c}
}

func (p *aztecGenericGFPoly) degree() int { return len(p.coefficients) - 1 }

func (p *aztecGenericGFPoly) isZero() bool { return p.coefficients[0] == 0 }

func (p *aztecGenericGFPoly) getCoefficient(degree int) int {
	return p.coefficients[len(p.coefficients)-1-degree]
}

func (p *aztecGenericGFPoly) addOrSubtract(other *aztecGenericGFPoly) *aztecGenericGFPoly {
	if p.isZero() {
		return other
	}
	if other.isZero() {
		return p
	}
	smaller := p.coefficients
	larger := other.coefficients
	if len(smaller) > len(larger) {
		smaller, larger = larger, smaller
	}
	sumDiff := make([]int, len(larger))
	lengthDiff := len(larger) - len(smaller)
	copy(sumDiff, larger[:lengthDiff])
	for i := lengthDiff; i < len(larger); i++ {
		sumDiff[i] = aztecGFAddOrSubtract(smaller[i-lengthDiff], larger[i])
	}
	return newAztecGenericGFPoly(p.field, sumDiff)
}

func (p *aztecGenericGFPoly) multiplyPoly(other *aztecGenericGFPoly) *aztecGenericGFPoly {
	if p.isZero() || other.isZero() {
		return p.field.zero
	}
	a := p.coefficients
	b := other.coefficients
	product := make([]int, len(a)+len(b)-1)
	for i, ac := range a {
		for j, bc := range b {
			product[i+j] = aztecGFAddOrSubtract(product[i+j], p.field.multiply(ac, bc))
		}
	}
	return newAztecGenericGFPoly(p.field, product)
}

func (p *aztecGenericGFPoly) multiplyScalar(scalar int) *aztecGenericGFPoly {
	if scalar == 0 {
		return p.field.zero
	}
	if scalar == 1 {
		return p
	}
	product := make([]int, len(p.coefficients))
	for i, c := range p.coefficients {
		product[i] = p.field.multiply(c, scalar)
	}
	return newAztecGenericGFPoly(p.field, product)
}

func (p *aztecGenericGFPoly) multiplyByMonomial(degree, coefficient int) *aztecGenericGFPoly {
	if coefficient == 0 {
		return p.field.zero
	}
	product := make([]int, len(p.coefficients)+degree)
	for i, c := range p.coefficients {
		product[i] = p.field.multiply(c, coefficient)
	}
	return newAztecGenericGFPoly(p.field, product)
}

// divide returns [quotient, remainder].
func (p *aztecGenericGFPoly) divide(other *aztecGenericGFPoly) [2]*aztecGenericGFPoly {
	quotient := p.field.zero
	remainder := p

	denominatorLeadingTerm := other.getCoefficient(other.degree())
	inverseDenominatorLeadingTerm := p.field.inverse(denominatorLeadingTerm)

	for remainder.degree() >= other.degree() && !remainder.isZero() {
		degreeDifference := remainder.degree() - other.degree()
		scale := p.field.multiply(remainder.getCoefficient(remainder.degree()), inverseDenominatorLeadingTerm)
		term := other.multiplyByMonomial(degreeDifference, scale)
		iterationQuotient := p.field.buildMonomial(degreeDifference, scale)
		quotient = quotient.addOrSubtract(iterationQuotient)
		remainder = remainder.addOrSubtract(term)
	}
	return [2]*aztecGenericGFPoly{quotient, remainder}
}

// ─────────────────────────────────────────────────────────────────────────────
// aztecRSEncoder — Reed-Solomon encoder over a GenericGF
// Ref: ReedSolomonEncoder.cs
// ─────────────────────────────────────────────────────────────────────────────

type aztecRSEncoder struct {
	field           *aztecGenericGF
	cachedGenerators []*aztecGenericGFPoly
}

func newAztecRSEncoder(field *aztecGenericGF) *aztecRSEncoder {
	e := &aztecRSEncoder{field: field}
	e.cachedGenerators = append(e.cachedGenerators,
		newAztecGenericGFPoly(field, []int{1}))
	return e
}

func (e *aztecRSEncoder) buildGenerator(degree int) *aztecGenericGFPoly {
	if degree >= len(e.cachedGenerators) {
		last := e.cachedGenerators[len(e.cachedGenerators)-1]
		for d := len(e.cachedGenerators); d <= degree; d++ {
			next := last.multiplyPoly(newAztecGenericGFPoly(e.field,
				[]int{1, e.field.exp(d - 1 + e.field.generatorBase)}))
			e.cachedGenerators = append(e.cachedGenerators, next)
			last = next
		}
	}
	return e.cachedGenerators[degree]
}

// encode fills toEncode[dataBytes:] with ecBytes error correction codewords.
// toEncode must have length dataBytes+ecBytes.
func (e *aztecRSEncoder) encode(toEncode []int, ecBytes int) {
	dataBytes := len(toEncode) - ecBytes
	generator := e.buildGenerator(ecBytes)

	infoCoeffs := make([]int, dataBytes)
	copy(infoCoeffs, toEncode[:dataBytes])

	info := newAztecGenericGFPoly(e.field, infoCoeffs)
	info = info.multiplyByMonomial(ecBytes, 1)

	remainder := info.divide(generator)[1]
	coefficients := remainder.coefficients
	numZeroCoefficients := ecBytes - len(coefficients)
	for i := 0; i < numZeroCoefficients; i++ {
		toEncode[dataBytes+i] = 0
	}
	copy(toEncode[dataBytes+numZeroCoefficients:], coefficients)
}

// ─────────────────────────────────────────────────────────────────────────────
// aztecBitArray — compact bit array (LSB-first within each int32 word)
// Ref: BitArray.cs
// ─────────────────────────────────────────────────────────────────────────────

type aztecBitArray struct {
	bits []int32
	size int
}

func newAztecBitArray() *aztecBitArray { return &aztecBitArray{bits: make([]int32, 1)} }

func (a *aztecBitArray) Size() int { return a.size }

func (a *aztecBitArray) get(i int) bool {
	return (a.bits[i>>5]&(1<<(i&0x1F))) != 0
}

func (a *aztecBitArray) set(i int) {
	a.bits[i>>5] |= 1 << (i & 0x1F)
}

func (a *aztecBitArray) ensureCapacity(sz int) {
	if sz > len(a.bits)<<5 {
		newBits := make([]int32, (sz+31)>>5)
		copy(newBits, a.bits)
		a.bits = newBits
	}
}

func (a *aztecBitArray) appendBit(bit bool) {
	a.ensureCapacity(a.size + 1)
	if bit {
		a.bits[a.size>>5] |= 1 << (a.size & 0x1F)
	}
	a.size++
}

// appendBits appends the least-significant numBits of value, MSB first.
func (a *aztecBitArray) appendBits(value, numBits int) {
	a.ensureCapacity(a.size + numBits)
	for numBitsLeft := numBits; numBitsLeft > 0; numBitsLeft-- {
		a.appendBit(((value >> (numBitsLeft - 1)) & 0x01) == 1)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// aztecBitMatrix — 2D bit matrix, indexed as [x, y] (column, row)
// Ref: BitMatrix.cs
// ─────────────────────────────────────────────────────────────────────────────

type aztecBitMatrix struct {
	width   int
	height  int
	rowSize int
	bits    []int32
}

func newAztecBitMatrix(dimension int) *aztecBitMatrix {
	rowSize := (dimension + 31) >> 5
	return &aztecBitMatrix{
		width:   dimension,
		height:  dimension,
		rowSize: rowSize,
		bits:    make([]int32, rowSize*dimension),
	}
}

// get returns the bit at column x, row y.
func (m *aztecBitMatrix) get(x, y int) bool {
	offset := y*m.rowSize + (x >> 5)
	return (int(uint(m.bits[offset])>>(x&0x1f))&1) != 0
}

// set sets the bit at column x, row y to true.
func (m *aztecBitMatrix) set(x, y int) {
	offset := y*m.rowSize + (x >> 5)
	m.bits[offset] |= 1 << (x & 0x1f)
}

// ─────────────────────────────────────────────────────────────────────────────
// Token types
// Ref: Token.cs, SimpleToken.cs, BinaryShiftToken.cs
// ─────────────────────────────────────────────────────────────────────────────

type aztecToken interface {
	previous() aztecToken
	appendTo(bitArray *aztecBitArray, text []byte)
}

// aztecSimpleToken holds a value and a bit count.
type aztecSimpleToken struct {
	prev     aztecToken
	value    int16
	bitCount int16
}

var aztecTokenEmpty aztecToken = &aztecSimpleToken{prev: nil, value: 0, bitCount: 0}

func (t *aztecSimpleToken) previous() aztecToken { return t.prev }

func (t *aztecSimpleToken) appendTo(ba *aztecBitArray, text []byte) {
	ba.appendBits(int(t.value), int(t.bitCount))
}

func aztecTokenAdd(prev aztecToken, value, bitCount int) aztecToken {
	return &aztecSimpleToken{prev: prev, value: int16(value), bitCount: int16(bitCount)}
}

func aztecTokenAddBinaryShift(prev aztecToken, start, byteCount int) aztecToken {
	return &aztecBinaryShiftToken{prev: prev, binaryShiftStart: int16(start), binaryShiftByteCount: int16(byteCount)}
}

// aztecBinaryShiftToken encodes a run of raw bytes.
type aztecBinaryShiftToken struct {
	prev                 aztecToken
	binaryShiftStart     int16
	binaryShiftByteCount int16
}

func (t *aztecBinaryShiftToken) previous() aztecToken { return t.prev }

func (t *aztecBinaryShiftToken) appendTo(ba *aztecBitArray, text []byte) {
	count := int(t.binaryShiftByteCount)
	for i := 0; i < count; i++ {
		if i == 0 || (i == 31 && count <= 62) {
			ba.appendBits(31, 5) // BINARY_SHIFT
			if count > 62 {
				ba.appendBits(count-31, 16)
			} else if i == 0 {
				lo := count
				if lo > 31 {
					lo = 31
				}
				ba.appendBits(lo, 5)
			} else {
				ba.appendBits(count-31, 5)
			}
		}
		ba.appendBits(int(text[int(t.binaryShiftStart)+i]), 8)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// HighLevelEncoder tables and constants
// Ref: HighLevelEncoder.cs
// ─────────────────────────────────────────────────────────────────────────────

const (
	aztecModeUpper = 0 // 5 bits
	aztecModeLower = 1 // 5 bits
	aztecModeDigit = 2 // 4 bits
	aztecModeMixed = 3 // 5 bits
	aztecModePunct = 4 // 5 bits
)

// aztecLatchTable[from][to] — high 16 bits = number of bits, low 16 bits = the latch bits.
var aztecLatchTable = [5][5]int{
	{ // FROM UPPER
		0,
		(5 << 16) + 28,                    // UPPER -> LOWER
		(5 << 16) + 30,                    // UPPER -> DIGIT
		(5 << 16) + 29,                    // UPPER -> MIXED
		(10 << 16) + (29 << 5) + 30,       // UPPER -> MIXED -> PUNCT
	},
	{ // FROM LOWER
		(9 << 16) + (30 << 4) + 14,        // LOWER -> DIGIT -> UPPER
		0,
		(5 << 16) + 30,                    // LOWER -> DIGIT
		(5 << 16) + 29,                    // LOWER -> MIXED
		(10 << 16) + (29 << 5) + 30,       // LOWER -> MIXED -> PUNCT
	},
	{ // FROM DIGIT
		(4 << 16) + 14,                    // DIGIT -> UPPER
		(9 << 16) + (14 << 5) + 28,        // DIGIT -> UPPER -> LOWER
		0,
		(9 << 16) + (14 << 5) + 29,        // DIGIT -> UPPER -> MIXED
		(14 << 16) + (14 << 10) + (29 << 5) + 30, // DIGIT -> UPPER -> MIXED -> PUNCT
	},
	{ // FROM MIXED
		(5 << 16) + 29,                    // MIXED -> UPPER
		(5 << 16) + 28,                    // MIXED -> LOWER
		(10 << 16) + (29 << 5) + 30,       // MIXED -> UPPER -> DIGIT
		0,
		(5 << 16) + 30,                    // MIXED -> PUNCT
	},
	{ // FROM PUNCT
		(5 << 16) + 31,                    // PUNCT -> UPPER
		(10 << 16) + (31 << 5) + 28,       // PUNCT -> UPPER -> LOWER
		(10 << 16) + (31 << 5) + 30,       // PUNCT -> UPPER -> DIGIT
		(10 << 16) + (31 << 5) + 29,       // PUNCT -> UPPER -> MIXED
		0,
	},
}

// aztecCharMap[mode][char] — encoding value for char in mode. 0 means no mapping.
var aztecCharMap [5][256]int

// aztecShiftTable[mode][mode] — shift code. -1 means no shift available.
var aztecShiftTable [6][6]int

func init() {
	// Initialize shift table to -1
	for i := range aztecShiftTable {
		for j := range aztecShiftTable[i] {
			aztecShiftTable[i][j] = -1
		}
	}

	// UPPER mode
	aztecCharMap[aztecModeUpper][' '] = 1
	for c := 'A'; c <= 'Z'; c++ {
		aztecCharMap[aztecModeUpper][c] = int(c-'A') + 2
	}

	// LOWER mode
	aztecCharMap[aztecModeLower][' '] = 1
	for c := 'a'; c <= 'z'; c++ {
		aztecCharMap[aztecModeLower][c] = int(c-'a') + 2
	}

	// DIGIT mode
	aztecCharMap[aztecModeDigit][' '] = 1
	for c := '0'; c <= '9'; c++ {
		aztecCharMap[aztecModeDigit][c] = int(c-'0') + 2
	}
	aztecCharMap[aztecModeDigit][','] = 12
	aztecCharMap[aztecModeDigit]['.'] = 13

	// MIXED mode
	mixedTable := []int{
		'\x00', ' ', 1, 2, 3, 4, 5, 6, 7, '\b', '\t', '\n', 11, '\f', '\r',
		27, 28, 29, 30, 31, '@', '\\', '^', '_', '`', '|', '~', 127,
	}
	for i, ch := range mixedTable {
		aztecCharMap[aztecModeMixed][ch] = i
	}

	// PUNCT mode
	punctTable := []int{
		'\x00', '\r', '\x00', '\x00', '\x00', '\x00', '!', '\'', '#', '$', '%', '&', '\'',
		'(', ')', '*', '+', ',', '-', '.', '/', ':', ';', '<', '=', '>', '?',
		'[', ']', '{', '}',
	}
	for i, ch := range punctTable {
		if ch > 0 {
			aztecCharMap[aztecModePunct][ch] = i
		}
	}

	// Shift table
	aztecShiftTable[aztecModeUpper][aztecModePunct] = 0
	aztecShiftTable[aztecModeLower][aztecModePunct] = 0
	aztecShiftTable[aztecModeLower][aztecModeUpper] = 28
	aztecShiftTable[aztecModeMixed][aztecModePunct] = 0
	aztecShiftTable[aztecModeDigit][aztecModePunct] = 0
	aztecShiftTable[aztecModeDigit][aztecModeUpper] = 15
}

// ─────────────────────────────────────────────────────────────────────────────
// aztecState — encoding state for the dynamic programming HLE algorithm
// Ref: State.cs
// ─────────────────────────────────────────────────────────────────────────────

type aztecState struct {
	token                aztecToken
	mode                 int
	binaryShiftByteCount int
	bitCount             int
}

var aztecInitialState = &aztecState{
	token:    aztecTokenEmpty,
	mode:     aztecModeUpper,
	bitCount: 0,
}

func (s *aztecState) latchAndAppend(mode, value int) *aztecState {
	bc := s.bitCount
	token := s.token
	if mode != s.mode {
		latch := aztecLatchTable[s.mode][mode]
		token = aztecTokenAdd(token, latch&0xFFFF, latch>>16)
		bc += latch >> 16
	}
	latchModeBitCount := 5
	if mode == aztecModeDigit {
		latchModeBitCount = 4
	}
	token = aztecTokenAdd(token, value, latchModeBitCount)
	return &aztecState{token: token, mode: mode, bitCount: bc + latchModeBitCount}
}

func (s *aztecState) shiftAndAppend(mode, value int) *aztecState {
	token := s.token
	thisModeBitCount := 5
	if s.mode == aztecModeDigit {
		thisModeBitCount = 4
	}
	token = aztecTokenAdd(token, aztecShiftTable[s.mode][mode], thisModeBitCount)
	token = aztecTokenAdd(token, value, 5)
	return &aztecState{token: token, mode: s.mode, bitCount: s.bitCount + thisModeBitCount + 5}
}

func (s *aztecState) addBinaryShiftChar(index int) *aztecState {
	token := s.token
	mode := s.mode
	bc := s.bitCount
	if s.mode == aztecModePunct || s.mode == aztecModeDigit {
		latch := aztecLatchTable[mode][aztecModeUpper]
		token = aztecTokenAdd(token, latch&0xFFFF, latch>>16)
		bc += latch >> 16
		mode = aztecModeUpper
	}
	var deltaBitCount int
	switch {
	case s.binaryShiftByteCount == 0 || s.binaryShiftByteCount == 31:
		deltaBitCount = 18
	case s.binaryShiftByteCount == 62:
		deltaBitCount = 9
	default:
		deltaBitCount = 8
	}
	result := &aztecState{
		token:                token,
		mode:                 mode,
		binaryShiftByteCount: s.binaryShiftByteCount + 1,
		bitCount:             bc + deltaBitCount,
	}
	if result.binaryShiftByteCount == 2047+31 {
		result = result.endBinaryShift(index + 1)
	}
	return result
}

func (s *aztecState) endBinaryShift(index int) *aztecState {
	if s.binaryShiftByteCount == 0 {
		return s
	}
	token := aztecTokenAddBinaryShift(s.token, index-s.binaryShiftByteCount, s.binaryShiftByteCount)
	return &aztecState{token: token, mode: s.mode, bitCount: s.bitCount}
}

func (s *aztecState) isBetterThanOrEqualTo(other *aztecState) bool {
	mySize := s.bitCount + (aztecLatchTable[s.mode][other.mode] >> 16)
	if other.binaryShiftByteCount > 0 &&
		(s.binaryShiftByteCount == 0 || s.binaryShiftByteCount > other.binaryShiftByteCount) {
		mySize += 10
	}
	return mySize <= other.bitCount
}

func (s *aztecState) toBitArray(text []byte) *aztecBitArray {
	// collect token chain in reverse, then play forward
	var symbols []aztecToken
	for tok := s.endBinaryShift(len(text)).token; tok != nil; tok = tok.previous() {
		symbols = append(symbols, tok)
	}
	ba := newAztecBitArray()
	for i := len(symbols) - 1; i >= 0; i-- {
		symbols[i].appendTo(ba, text)
	}
	return ba
}

// ─────────────────────────────────────────────────────────────────────────────
// HighLevelEncoder — dynamic programming encoder
// Ref: HighLevelEncoder.cs
// ─────────────────────────────────────────────────────────────────────────────

func aztecHighLevelEncode(text []byte) *aztecBitArray {
	states := []*aztecState{aztecInitialState}
	for index := 0; index < len(text); index++ {
		var pairCode int
		var nextChar int
		if index+1 < len(text) {
			nextChar = int(text[index+1])
		}
		switch text[index] {
		case '\r':
			if nextChar == '\n' {
				pairCode = 2
			}
		case '.':
			if nextChar == ' ' {
				pairCode = 3
			}
		case ',':
			if nextChar == ' ' {
				pairCode = 4
			}
		case ':':
			if nextChar == ' ' {
				pairCode = 5
			}
		}
		if pairCode > 0 {
			states = aztecUpdateStateListForPair(states, index, pairCode)
			index++
		} else {
			states = aztecUpdateStateListForChar(states, index, text)
		}
	}
	// find minimum bit-count state
	minState := states[0]
	for _, st := range states[1:] {
		if st.bitCount < minState.bitCount {
			minState = st
		}
	}
	return minState.toBitArray(text)
}

func aztecUpdateStateListForChar(states []*aztecState, index int, text []byte) []*aztecState {
	var result []*aztecState
	for _, st := range states {
		result = aztecUpdateStateForChar(st, index, text, result)
	}
	return aztecSimplifyStates(result)
}

func aztecUpdateStateForChar(state *aztecState, index int, text []byte, result []*aztecState) []*aztecState {
	ch := int(text[index] & 0xFF)
	charInCurrentTable := aztecCharMap[state.mode][ch] > 0
	var stateNoBinary *aztecState
	for mode := 0; mode <= aztecModePunct; mode++ {
		charInMode := aztecCharMap[mode][ch]
		if charInMode > 0 {
			if stateNoBinary == nil {
				stateNoBinary = state.endBinaryShift(index)
			}
			if !charInCurrentTable || mode == state.mode || mode == aztecModeDigit {
				result = append(result, stateNoBinary.latchAndAppend(mode, charInMode))
			}
			if !charInCurrentTable && aztecShiftTable[state.mode][mode] >= 0 {
				result = append(result, stateNoBinary.shiftAndAppend(mode, charInMode))
			}
		}
	}
	if state.binaryShiftByteCount > 0 || aztecCharMap[state.mode][ch] == 0 {
		result = append(result, state.addBinaryShiftChar(index))
	}
	return result
}

func aztecUpdateStateListForPair(states []*aztecState, index, pairCode int) []*aztecState {
	var result []*aztecState
	for _, st := range states {
		result = aztecUpdateStateForPair(st, index, pairCode, result)
	}
	return aztecSimplifyStates(result)
}

func aztecUpdateStateForPair(state *aztecState, index, pairCode int, result []*aztecState) []*aztecState {
	stateNoBinary := state.endBinaryShift(index)
	result = append(result, stateNoBinary.latchAndAppend(aztecModePunct, pairCode))
	if state.mode != aztecModePunct {
		result = append(result, stateNoBinary.shiftAndAppend(aztecModePunct, pairCode))
	}
	if pairCode == 3 || pairCode == 4 {
		digitState := stateNoBinary.
			latchAndAppend(aztecModeDigit, 16-pairCode).
			latchAndAppend(aztecModeDigit, 1)
		result = append(result, digitState)
	}
	if state.binaryShiftByteCount > 0 {
		result = append(result, state.addBinaryShiftChar(index).addBinaryShiftChar(index+1))
	}
	return result
}

func aztecSimplifyStates(states []*aztecState) []*aztecState {
	var result []*aztecState
	for _, newState := range states {
		add := true
		var filtered []*aztecState
		for _, oldState := range result {
			if oldState.isBetterThanOrEqualTo(newState) {
				add = false
				filtered = append(filtered, oldState)
			} else if newState.isBetterThanOrEqualTo(oldState) {
				// drop oldState
			} else {
				filtered = append(filtered, oldState)
			}
		}
		result = filtered
		if add {
			result = append(result, newState)
		}
	}
	return result
}

// ─────────────────────────────────────────────────────────────────────────────
// Aztec Encoder — top-level encode logic
// Ref: Encoder.cs
// ─────────────────────────────────────────────────────────────────────────────

// WORD_SIZE[layers] — bits per codeword for each layer count (1-indexed, index 0 unused).
var aztecWordSizeTable = []int{
	4,  // unused index 0
	6, 6, 8, 8, 8, 8, 8, 8, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
	12, 12, 12, 12, 12, 12, 12, 12, 12, 12,
}

const (
	aztecDefaultECPercent    = 33
	aztecMaxNbBits           = 32
	aztecMaxNbBitsCompact    = 4
)

func aztecTotalBitsInLayer(layers int, compact bool) int {
	base := 88
	if !compact {
		base = 112
	}
	return (base + 16*layers) * layers
}

func aztecGetGF(wordSize int) *aztecGenericGF {
	switch wordSize {
	case 4:
		return aztecGFParam
	case 6:
		return aztecGFData6
	case 8:
		return aztecGFData8
	case 10:
		return aztecGFData10
	case 12:
		return aztecGFData12
	}
	return nil
}

// aztecStuffBits inserts stuffing bits to avoid all-0 or all-1 codewords.
// Ref: Encoder.cs stuffBits()
func aztecStuffBits(bits *aztecBitArray, wordSize int) *aztecBitArray {
	out := newAztecBitArray()
	n := bits.Size()
	mask := (1 << wordSize) - 2
	for i := 0; i < n; i += wordSize {
		word := 0
		for j := 0; j < wordSize; j++ {
			if i+j >= n || bits.get(i+j) {
				word |= 1 << (wordSize - 1 - j)
			}
		}
		if (word & mask) == mask {
			out.appendBits(word&mask, wordSize)
			i-- // re-process the last bit
		} else if (word & mask) == 0 {
			out.appendBits(word|1, wordSize)
			i-- // re-process the last bit
		} else {
			out.appendBits(word, wordSize)
		}
	}
	return out
}

// aztecBitsToWords converts stuffedBits to an int array of totalWords words.
func aztecBitsToWords(stuffedBits *aztecBitArray, wordSize, totalWords int) []int {
	message := make([]int, totalWords)
	n := stuffedBits.Size() / wordSize
	for i := 0; i < n; i++ {
		value := 0
		for j := 0; j < wordSize; j++ {
			if stuffedBits.get(i*wordSize + j) {
				value |= 1 << (wordSize - j - 1)
			}
		}
		message[i] = value
	}
	return message
}

// aztecGenerateCheckWords appends RS check words to bitArray so the total is
// totalBits bits, using wordSize-bit codewords. Returns the extended BitArray.
func aztecGenerateCheckWords(bitArray *aztecBitArray, totalBits, wordSize int) *aztecBitArray {
	messageSizeInWords := bitArray.Size() / wordSize
	rs := newAztecRSEncoder(aztecGetGF(wordSize))
	totalWords := totalBits / wordSize
	messageWords := aztecBitsToWords(bitArray, wordSize, totalWords)
	rs.encode(messageWords, totalWords-messageSizeInWords)

	startPad := totalBits % wordSize
	messageBits := newAztecBitArray()
	messageBits.appendBits(0, startPad)
	for _, mw := range messageWords {
		messageBits.appendBits(mw, wordSize)
	}
	return messageBits
}

// aztecGenerateModeMessage generates the mode message BitArray.
// Ref: Encoder.cs generateModeMessage()
func aztecGenerateModeMessage(compact bool, layers, messageSizeInWords int) *aztecBitArray {
	modeMessage := newAztecBitArray()
	if compact {
		modeMessage.appendBits(layers-1, 2)
		modeMessage.appendBits(messageSizeInWords-1, 6)
		modeMessage = aztecGenerateCheckWords(modeMessage, 28, 4)
	} else {
		modeMessage.appendBits(layers-1, 5)
		modeMessage.appendBits(messageSizeInWords-1, 11)
		modeMessage = aztecGenerateCheckWords(modeMessage, 40, 4)
	}
	return modeMessage
}

// aztecEncoderDrawBullsEye draws the concentric square bulls-eye pattern.
// Ref: Encoder.cs drawBullsEye()
func aztecEncoderDrawBullsEye(matrix *aztecBitMatrix, center, size int) {
	for i := 0; i < size; i += 2 {
		for j := center - i; j <= center+i; j++ {
			matrix.set(j, center-i)
			matrix.set(j, center+i)
			matrix.set(center-i, j)
			matrix.set(center+i, j)
		}
	}
	matrix.set(center-size, center-size)
	matrix.set(center-size+1, center-size)
	matrix.set(center-size, center-size+1)
	matrix.set(center+size, center-size)
	matrix.set(center+size, center-size+1)
	matrix.set(center+size, center+size-1)
}

// aztecEncoderDrawModeMessage draws the mode message around the bulls-eye.
// Ref: Encoder.cs drawModeMessage()
func aztecEncoderDrawModeMessage(matrix *aztecBitMatrix, compact bool, matrixSize int, modeMessage *aztecBitArray) {
	center := matrixSize / 2
	if compact {
		for i := 0; i < 7; i++ {
			offset := center - 3 + i
			if modeMessage.get(i) {
				matrix.set(offset, center-5)
			}
			if modeMessage.get(i + 7) {
				matrix.set(center+5, offset)
			}
			if modeMessage.get(20 - i) {
				matrix.set(offset, center+5)
			}
			if modeMessage.get(27 - i) {
				matrix.set(center-5, offset)
			}
		}
	} else {
		for i := 0; i < 10; i++ {
			offset := center - 5 + i + i/5
			if modeMessage.get(i) {
				matrix.set(offset, center-7)
			}
			if modeMessage.get(i + 10) {
				matrix.set(center+7, offset)
			}
			if modeMessage.get(29 - i) {
				matrix.set(offset, center+7)
			}
			if modeMessage.get(39 - i) {
				matrix.set(center-7, offset)
			}
		}
	}
}

// encodeAztecFull is the main entry point replacing the simplified encoder.
// It returns a [][]bool matrix (row-major: matrix[row][col]).
// Ref: Encoder.cs encode()
func encodeAztecFull(data []byte, minECCPercent, userSpecifiedLayers int) ([][]bool, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("aztec: data must not be empty")
	}

	// High-level encode
	bits := aztecHighLevelEncode(data)

	eccBits := bits.Size()*minECCPercent/100 + 11
	totalSizeBits := bits.Size() + eccBits

	var compact bool
	var layers int
	var totalBitsInLayer int
	var wordSize int
	var stuffedBits *aztecBitArray

	if userSpecifiedLayers != 0 {
		compact = userSpecifiedLayers < 0
		layers = userSpecifiedLayers
		if layers < 0 {
			layers = -layers
		}
		maxLayers := aztecMaxNbBits
		if compact {
			maxLayers = aztecMaxNbBitsCompact
		}
		if layers > maxLayers {
			return nil, fmt.Errorf("aztec: illegal value %d for layers", userSpecifiedLayers)
		}
		totalBitsInLayer = aztecTotalBitsInLayer(layers, compact)
		wordSize = aztecWordSizeTable[layers]
		usableBitsInLayers := totalBitsInLayer - (totalBitsInLayer % wordSize)
		stuffedBits = aztecStuffBits(bits, wordSize)
		if stuffedBits.Size()+eccBits > usableBitsInLayers {
			return nil, fmt.Errorf("aztec: data too large for user-specified layer")
		}
		if compact && stuffedBits.Size() > wordSize*64 {
			return nil, fmt.Errorf("aztec: data too large for user-specified layer (compact)")
		}
	} else {
		wordSize = 0
		stuffedBits = nil
		found := false
		for i := 0; i <= aztecMaxNbBits; i++ {
			compact = i <= 3
			if compact {
				layers = i + 1
			} else {
				layers = i
			}
			totalBitsInLayer = aztecTotalBitsInLayer(layers, compact)
			if totalSizeBits > totalBitsInLayer {
				continue
			}
			if wordSize != aztecWordSizeTable[layers] {
				wordSize = aztecWordSizeTable[layers]
				stuffedBits = aztecStuffBits(bits, wordSize)
			}
			if stuffedBits == nil {
				continue
			}
			usableBitsInLayers := totalBitsInLayer - (totalBitsInLayer % wordSize)
			if compact && stuffedBits.Size() > wordSize*64 {
				continue
			}
			if stuffedBits.Size()+eccBits <= usableBitsInLayers {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("aztec: data too large for an Aztec code")
		}
	}

	messageBits := aztecGenerateCheckWords(stuffedBits, totalBitsInLayer, wordSize)

	messageSizeInWords := stuffedBits.Size() / wordSize
	modeMessage := aztecGenerateModeMessage(compact, layers, messageSizeInWords)

	// allocate symbol
	baseMatrixSize := 11 + layers*4
	if !compact {
		baseMatrixSize = 14 + layers*4
	}

	alignmentMap := make([]int, baseMatrixSize)
	var matrixSize int
	if compact {
		matrixSize = baseMatrixSize
		for i := range alignmentMap {
			alignmentMap[i] = i
		}
	} else {
		matrixSize = baseMatrixSize + 1 + 2*((baseMatrixSize/2-1)/15)
		origCenter := baseMatrixSize / 2
		center := matrixSize / 2
		for i := 0; i < origCenter; i++ {
			newOffset := i + i/15
			alignmentMap[origCenter-i-1] = center - newOffset - 1
			alignmentMap[origCenter+i] = center + newOffset + 1
		}
	}

	matrix := newAztecBitMatrix(matrixSize)

	// draw data bits
	// Ref: Encoder.cs lines 174-201
	rowOffset := 0
	for i := 0; i < layers; i++ {
		var rowSize int
		if compact {
			rowSize = (layers-i)*4 + 9
		} else {
			rowSize = (layers-i)*4 + 12
		}
		for j := 0; j < rowSize; j++ {
			columnOffset := j * 2
			for k := 0; k < 2; k++ {
				if messageBits.get(rowOffset + columnOffset + k) {
					matrix.set(alignmentMap[i*2+k], alignmentMap[i*2+j])
				}
				if messageBits.get(rowOffset + rowSize*2 + columnOffset + k) {
					matrix.set(alignmentMap[i*2+j], alignmentMap[baseMatrixSize-1-i*2-k])
				}
				if messageBits.get(rowOffset + rowSize*4 + columnOffset + k) {
					matrix.set(alignmentMap[baseMatrixSize-1-i*2-k], alignmentMap[baseMatrixSize-1-i*2-j])
				}
				if messageBits.get(rowOffset + rowSize*6 + columnOffset + k) {
					matrix.set(alignmentMap[baseMatrixSize-1-i*2-j], alignmentMap[i*2+k])
				}
			}
		}
		rowOffset += rowSize * 8
	}

	// draw mode message
	aztecEncoderDrawModeMessage(matrix, compact, matrixSize, modeMessage)

	// draw alignment marks (bulls-eye + reference grid)
	if compact {
		aztecEncoderDrawBullsEye(matrix, matrixSize/2, 5)
	} else {
		aztecEncoderDrawBullsEye(matrix, matrixSize/2, 7)
		for i, j := 0, 0; i < baseMatrixSize/2-1; i, j = i+15, j+16 {
			for k := (matrixSize / 2) & 1; k < matrixSize; k += 2 {
				matrix.set(matrixSize/2-j, k)
				matrix.set(matrixSize/2+j, k)
				matrix.set(k, matrixSize/2-j)
				matrix.set(k, matrixSize/2+j)
			}
		}
	}

	// Convert aztecBitMatrix to [][]bool (row-major: result[row][col] = matrix.get(col, row))
	result := make([][]bool, matrixSize)
	for row := 0; row < matrixSize; row++ {
		result[row] = make([]bool, matrixSize)
		for col := 0; col < matrixSize; col++ {
			result[row][col] = matrix.get(col, row)
		}
	}
	return result, nil
}
