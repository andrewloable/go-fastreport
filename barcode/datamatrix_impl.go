package barcode

// datamatrix_impl.go — pure-Go DataMatrix ECC 200 encoder.
//
// Ported from FastReport.Base/Barcode/BarcodeDatamatrix.cs (originally derived
// from iText / Paulo Soares, MPL 1.1).
//
// Implements GetMatrix() for DataMatrixBarcode and GS1DatamatrixBarcode.

import "fmt"

// -----------------------------------------------------------------------
// DataMatrix constants
// -----------------------------------------------------------------------

const (
	dmLatchB256   = byte(231)
	dmLatchEdifact = byte(240)
	dmLatchX12    = byte(238)
	dmLatchText   = byte(239)
	dmLatchC40    = byte(230)
	dmUnlatch     = byte(254)
	dmUpperCase   = byte(235)
)

const (
	dmSetX12 = "\r*> 0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	dmShiftedC40AndText = "!\"#$%&'()*+,-./:;<=>?@[\\]^_"
	dmBaseC40            = " 0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	dmShiftedC40         = "`abcdefghijklmnopqrstuvwxyz{|}~\u007f"
	dmBaseText           = " 0123456789abcdefghijklmnopqrstuvwxyz"
	dmShiftedText        = "`ABCDEFGHIJKLMNOPQRSTUVWXYZ{|}~\u007f"
)

// -----------------------------------------------------------------------
// DmParams — symbol geometry
// -----------------------------------------------------------------------

type dmParams struct {
	height        int
	width         int
	heightSection int
	widthSection  int
	dataSize      int
	dataBlock     int
	errorBlock    int
}

// dmSizes is the DataMatrix ECC200 symbol table (30 entries).
var dmSizes = []dmParams{
	{10, 10, 10, 10, 3, 3, 5},
	{12, 12, 12, 12, 5, 5, 7},
	{8, 18, 8, 18, 5, 5, 7},
	{14, 14, 14, 14, 8, 8, 10},
	{8, 32, 8, 16, 10, 10, 11},
	{16, 16, 16, 16, 12, 12, 12},
	{12, 26, 12, 26, 16, 16, 14},
	{18, 18, 18, 18, 18, 18, 14},
	{20, 20, 20, 20, 22, 22, 18},
	{12, 36, 12, 18, 22, 22, 18},
	{22, 22, 22, 22, 30, 30, 20},
	{16, 36, 16, 18, 32, 32, 24},
	{24, 24, 24, 24, 36, 36, 24},
	{26, 26, 26, 26, 44, 44, 28},
	{16, 48, 16, 24, 49, 49, 28},
	{32, 32, 16, 16, 62, 62, 36},
	{36, 36, 18, 18, 86, 86, 42},
	{40, 40, 20, 20, 114, 114, 48},
	{44, 44, 22, 22, 144, 144, 56},
	{48, 48, 24, 24, 174, 174, 68},
	{52, 52, 26, 26, 204, 102, 42},
	{64, 64, 16, 16, 280, 140, 56},
	{72, 72, 18, 18, 368, 92, 36},
	{80, 80, 20, 20, 456, 114, 48},
	{88, 88, 22, 22, 576, 144, 56},
	{96, 96, 24, 24, 696, 174, 68},
	{104, 104, 26, 26, 816, 136, 56},
	{120, 120, 20, 20, 1050, 175, 68},
	{132, 132, 22, 22, 1304, 163, 62},
	{144, 144, 24, 24, 1558, 156, 62},
}

// -----------------------------------------------------------------------
// dmEncoder — top-level encode state
// -----------------------------------------------------------------------

type dmEncoder struct {
	height int
	width  int
	image  []byte  // packed-bit image row-major, (width+7)/8 bytes per row
	place  []int16 // placement array
}

// dmGenerate encodes text into a DataMatrix and returns the bit image.
// Returns (image, height, width, error).
func dmGenerate(text []byte) ([]byte, int, int, error) {
	e := &dmEncoder{}
	data := make([]byte, 2500)
	textOffset := 0
	textSize := len(text)
	extCount := 0

	// Handle FNC1 prefix.
	if len(text) > 0 && text[0] == 232 {
		data[0] = 232
		textOffset++
		textSize--
		extCount = 1
	}

	var dm dmParams
	var enc int

	if e.height == 0 || e.width == 0 {
		last := dmSizes[len(dmSizes)-1]
		enc = dmGetEncodation(text, textOffset, textSize, data, extCount, last.dataSize-extCount, false, false)
		if enc < 0 {
			return nil, 0, 0, fmt.Errorf("datamatrix: text is too large for any symbol")
		}
		enc += extCount
		found := false
		for k := 0; k < len(dmSizes); k++ {
			if dmSizes[k].dataSize >= enc {
				dm = dmSizes[k]
				found = true
				break
			}
		}
		if !found {
			return nil, 0, 0, fmt.Errorf("datamatrix: text is too large")
		}
		e.height = dm.height
		e.width = dm.width
	} else {
		found := false
		for k := 0; k < len(dmSizes); k++ {
			if e.height == dmSizes[k].height && e.width == dmSizes[k].width {
				dm = dmSizes[k]
				found = true
				break
			}
		}
		if !found {
			return nil, 0, 0, fmt.Errorf("datamatrix: invalid symbol size %dx%d", e.height, e.width)
		}
		enc = dmGetEncodation(text, textOffset, textSize, data, extCount, dm.dataSize-extCount, true, true)
		if enc < 0 {
			return nil, 0, 0, fmt.Errorf("datamatrix: text is too large for selected symbol")
		}
		enc += extCount
	}

	xByte := (dm.width + 7) / 8
	e.image = make([]byte, xByte*dm.height)
	dmMakePadding(data, enc, dm.dataSize-enc)
	e.place = dmDoPlacement(dm.height-(dm.height/dm.heightSection*2), dm.width-(dm.width/dm.widthSection*2))
	full := dm.dataSize + ((dm.dataSize+2)/dm.dataBlock)*dm.errorBlock
	dmGenerateECC(data, dm.dataSize, dm.dataBlock, dm.errorBlock)
	dmDraw(e.image, data, full, dm, e.place)

	return e.image, dm.height, dm.width, nil
}

// dmDraw renders the encoded data + ECC into the bit image.
func dmDraw(image []byte, data []byte, dataSize int, dm dmParams, place []int16) {
	xByte := (dm.width + 7) / 8

	// Clear image.
	for k := range image {
		image[k] = 0
	}

	// Dotted horizontal line (top of each section).
	for i := 0; i < dm.height; i += dm.heightSection {
		for j := 0; j < dm.width; j += 2 {
			dmSetBit(image, j, i, xByte)
		}
	}

	// Solid horizontal line (bottom of each section).
	for i := dm.heightSection - 1; i < dm.height; i += dm.heightSection {
		for j := 0; j < dm.width; j++ {
			dmSetBit(image, j, i, xByte)
		}
	}

	// Solid vertical line (left of each section).
	for i := 0; i < dm.width; i += dm.widthSection {
		for j := 0; j < dm.height; j++ {
			dmSetBit(image, i, j, xByte)
		}
	}

	// Dotted vertical line (right of each section).
	for i := dm.widthSection - 1; i < dm.width; i += dm.widthSection {
		for j := 1; j < dm.height; j += 2 {
			dmSetBit(image, i, j, xByte)
		}
	}

	// Place data bits.
	p := 0
	for ys := 0; ys < dm.height; ys += dm.heightSection {
		for y := 1; y < dm.heightSection-1; y++ {
			for xs := 0; xs < dm.width; xs += dm.widthSection {
				for x := 1; x < dm.widthSection-1; x++ {
					z := int(place[p])
					p++
					if z == 1 || (z > 1 && ((int(data[z/8-1])&0xff)&(128>>(z%8))) != 0) {
						dmSetBit(image, x+xs, y+ys, xByte)
					}
				}
			}
		}
	}
}

func dmSetBit(image []byte, x, y, xByte int) {
	image[y*xByte+x/8] |= byte(128 >> uint(x&7))
}

// -----------------------------------------------------------------------
// Padding
// -----------------------------------------------------------------------

func dmMakePadding(data []byte, position, count int) {
	if count <= 0 {
		return
	}
	data[position] = 129
	position++
	count--
	for count > 0 {
		t := 129 + (((position + 1) * 149) % 253) + 1
		if t > 254 {
			t -= 254
		}
		data[position] = byte(t)
		position++
		count--
	}
}

// -----------------------------------------------------------------------
// Encodation selection
// -----------------------------------------------------------------------

func dmIsDigit(c int) bool { return c >= '0' && c <= '9' }

func dmGetEncodation(text []byte, textOffset, textSize int, data []byte, dataOffset, dataSize int, sizeFixed, firstMatch bool) int {
	if dataSize < 0 {
		return -1
	}

	e1 := [6]int{}
	e1[0] = dmAsciiEncodation(text, textOffset, textSize, data, dataOffset, dataSize, -1)
	if firstMatch && e1[0] >= 0 {
		return e1[0]
	}
	e1[1] = dmC40OrTextEncodation(text, textOffset, textSize, data, dataOffset, dataSize, true, -1, -1, dataOffset)
	if firstMatch && e1[1] >= 0 {
		return e1[1]
	}
	e1[2] = dmC40OrTextEncodation(text, textOffset, textSize, data, dataOffset, dataSize, false, -1, -1, dataOffset)
	if firstMatch && e1[2] >= 0 {
		return e1[2]
	}
	e1[3] = dmB256Encodation(text, textOffset, textSize, data, dataOffset, dataSize, -1, dataOffset)
	if firstMatch && e1[3] >= 0 {
		return e1[3]
	}
	e1[4] = dmX12Encodation(text, textOffset, textSize, data, dataOffset, dataSize, -1, dataOffset)
	if firstMatch && e1[4] >= 0 {
		return e1[4]
	}
	e1[5] = dmEdifactEncodation(text, textOffset, textSize, data, dataOffset, dataSize, -1, -1, dataOffset, sizeFixed)
	if firstMatch && e1[5] >= 0 {
		return e1[5]
	}

	if e1[0] < 0 && e1[1] < 0 && e1[2] < 0 && e1[3] < 0 && e1[4] < 0 && e1[5] < 0 {
		return -1
	}

	// Pick shortest.
	j := 0
	best := 99999
	for k := 0; k < 6; k++ {
		if e1[k] >= 0 && e1[k] < best {
			best = e1[k]
			j = k
		}
	}

	switch j {
	case 0:
		return dmAsciiEncodation(text, textOffset, textSize, data, dataOffset, dataSize, -1)
	case 1:
		return dmC40OrTextEncodation(text, textOffset, textSize, data, dataOffset, dataSize, true, -1, -1, dataOffset)
	case 2:
		return dmC40OrTextEncodation(text, textOffset, textSize, data, dataOffset, dataSize, false, -1, -1, dataOffset)
	case 3:
		return dmB256Encodation(text, textOffset, textSize, data, dataOffset, dataSize, -1, dataOffset)
	case 4:
		return dmX12Encodation(text, textOffset, textSize, data, dataOffset, dataSize, -1, dataOffset)
	default:
		return dmEdifactEncodation(text, textOffset, textSize, data, dataOffset, dataSize, -1, -1, dataOffset, sizeFixed)
	}
}

// -----------------------------------------------------------------------
// ASCII encodation
// -----------------------------------------------------------------------

func dmAsciiEncodation(text []byte, textOffset, textLength int, data []byte, dataOffset, dataLength, symbolIndex int) int {
	textIndex := textOffset
	dataIndex := dataOffset
	textEnd := textOffset + textLength
	dataEnd := dataOffset + dataLength

	for textIndex < textEnd {
		c := int(text[textIndex]) & 0xff
		textIndex++

		if dmIsDigit(c) && symbolIndex < 0 && textIndex < textEnd && dmIsDigit(int(text[textIndex])&0xff) {
			data[dataIndex] = byte((c-'0')*10 + int(text[textIndex]&0xff) - '0' + 130)
			dataIndex++
			textIndex++
		} else if c == 232 {
			data[dataIndex] = 232
			dataIndex++
		} else if c > 127 {
			if dataIndex+1 >= dataEnd {
				return -1
			}
			data[dataIndex] = dmUpperCase
			dataIndex++
			data[dataIndex] = byte(c - 128 + 1)
			dataIndex++
		} else {
			data[dataIndex] = byte(c + 1)
			dataIndex++
		}
	}
	return dataIndex - dataOffset
}

// -----------------------------------------------------------------------
// Base256 encodation
// -----------------------------------------------------------------------

func dmB256Encodation(text []byte, textOffset, textLength int, data []byte, dataOffset, dataLength, prevEnc, origDataOffset int) int {
	if textLength == 0 {
		return 0
	}

	simulatedDataOffset := dataOffset
	minRequiredDataIncrement := 0

	if prevEnc != 3 { // 3 = Base256
		if textLength < 250 && textLength+2 > dataLength {
			return -1
		}
		if textLength >= 250 && textLength+3 > dataLength {
			return -1
		}
		data[dataOffset] = dmLatchB256
	}

	if textLength < 250 {
		data[simulatedDataOffset+1] = byte(textLength)
		if prevEnc != 3 {
			minRequiredDataIncrement = 2
		}
	} else if textLength == 250 && prevEnc == 3 {
		data[simulatedDataOffset+1] = byte(textLength/250 + 249)
		for i := dataOffset + 1; i > simulatedDataOffset+2; i-- {
			data[i] = data[i-1]
		}
		data[simulatedDataOffset+2] = byte(textLength % 250)
		minRequiredDataIncrement = 1
	} else {
		data[simulatedDataOffset+1] = byte(textLength/250 + 249)
		data[simulatedDataOffset+2] = byte(textLength % 250)
		if prevEnc != 3 {
			minRequiredDataIncrement = 3
		}
	}

	copyLen := textLength
	if prevEnc == 3 {
		copyLen = 1
	}
	copy(data[minRequiredDataIncrement+dataOffset:], text[textOffset:textOffset+copyLen])

	jStart := dataOffset + 1
	if prevEnc == 3 {
		jStart = dataOffset
	}
	jEnd := minRequiredDataIncrement + copyLen + dataOffset
	for j := jStart; j < jEnd; j++ {
		dmRandomizationAlgorithm255(data, j)
	}

	if prevEnc == 3 {
		dmRandomizationAlgorithm255(data, simulatedDataOffset+1)
	}

	return copyLen + dataOffset + minRequiredDataIncrement - origDataOffset
}

func dmRandomizationAlgorithm255(data []byte, j int) {
	c := int(data[j]) & 0xff
	prn := 149*(j+1)%255 + 1
	tv := c + prn
	if tv > 255 {
		tv -= 256
	}
	data[j] = byte(tv)
}

// -----------------------------------------------------------------------
// X12 encodation
// -----------------------------------------------------------------------

func dmX12Encodation(text []byte, textOffset, textLength int, data []byte, dataOffset, dataLength, symbolIndex, origDataOffset int) int {
	if textLength == 0 {
		return 0
	}

	x := make([]byte, textLength)
	count := 0

	for ti := 0; ti < textLength; ti++ {
		i := indexOf(dmSetX12, rune(text[ti+textOffset]))
		if i >= 0 {
			x[ti] = byte(i)
			count++
		} else {
			x[ti] = 100
			if count >= 6 {
				count -= count / 3 * 3
			}
			for k := 0; k < count; k++ {
				x[ti-k-1] = 100
			}
			count = 0
		}
	}

	if count >= 6 {
		count -= count / 3 * 3
	}
	for k := 0; k < count; k++ {
		x[textLength-k-1] = 100
	}

	textIndex := 0
	dataIndex := 0
	c := byte(0)

	for textIndex < textLength {
		c = x[textIndex]
		if dataIndex > dataLength {
			break
		}

		if c < 40 {
			if textIndex == 0 || x[textIndex-1] > 40 {
				data[dataOffset+dataIndex] = dmLatchX12
				dataIndex++
			}
			if dataIndex+2 > dataLength {
				break
			}
			n := 1600*int(x[textIndex]) + 40*int(x[textIndex+1]) + int(x[textIndex+2]) + 1
			data[dataOffset+dataIndex] = byte(n / 256)
			dataIndex++
			data[dataOffset+dataIndex] = byte(n)
			dataIndex++
			textIndex += 3
		} else {
			if symbolIndex <= 0 {
				if textIndex > 0 && x[textIndex-1] < 40 {
					data[dataOffset+dataIndex] = dmUnlatch
					dataIndex++
				}
			}
			i := dmAsciiEncodation(text, textOffset+textIndex, 1, data, dataOffset+dataIndex, dataLength, -1)
			if i < 0 {
				return -1
			}
			if data[dataOffset+dataIndex] == dmUpperCase {
				dataIndex++
			}
			dataIndex++
			textIndex++
		}
	}

	c = 100
	if textLength > 0 {
		c = x[textLength-1]
	}

	if textIndex != textLength {
		return -1
	}

	if c < 40 {
		data[dataOffset+dataIndex] = dmUnlatch
		dataIndex++
	}

	if dataIndex > dataLength {
		return -1
	}

	return dataIndex + dataOffset - origDataOffset
}

// -----------------------------------------------------------------------
// Edifact encodation
// -----------------------------------------------------------------------

func dmEdifactEncodation(text []byte, textOffset, textLength int, data []byte, dataOffset, dataLength, symbolIndex, prevEnc, origDataOffset int, sizeFixed bool) int {
	if textLength == 0 {
		return 0
	}

	textIndex := 0
	dataIndex := 0
	edi := 0
	pedi := 18
	ascii := true
	dataSize := dataOffset + dataLength

	for textIndex < textLength {
		c := int(text[textIndex+textOffset]) & 0xff

		if ((c&0xe0) == 0x40 || (c&0xe0) == 0x20) && c != '_' {
			if ascii {
				if dataIndex+1 > dataLength {
					break
				}
				data[dataOffset+dataIndex] = dmLatchEdifact
				dataIndex++
				ascii = false
			}
			c &= 0x3f
			edi |= c << uint(pedi)
			if pedi == 0 {
				if dataIndex+3 > dataLength {
					break
				}
				data[dataOffset+dataIndex] = byte(edi >> 16)
				dataIndex++
				data[dataOffset+dataIndex] = byte(edi >> 8)
				dataIndex++
				data[dataOffset+dataIndex] = byte(edi)
				dataIndex++
				edi = 0
				pedi = 18
			} else {
				pedi -= 6
			}
		} else {
			if !ascii {
				edi |= ('_' & 0x3f) << uint(pedi)
				if dataIndex+3-pedi/8 > dataLength {
					break
				}
				data[dataOffset+dataIndex] = byte(edi >> 16)
				dataIndex++
				if pedi <= 12 {
					data[dataOffset+dataIndex] = byte(edi >> 8)
					dataIndex++
				}
				if pedi <= 6 {
					data[dataOffset+dataIndex] = byte(edi)
					dataIndex++
				}
				ascii = true
				pedi = 18
				edi = 0
			}
			if dmIsDigit(c) && textOffset+textIndex > 0 && dmIsDigit(int(text[textOffset+textIndex-1])&0xff) &&
				prevEnc == 6 && data[dataOffset-1] >= 49 && data[dataOffset-1] <= 58 { // 6 = Edifact
				data[dataOffset+dataIndex-1] = byte((int(text[textOffset-1]&0xff)-'0')*10 + c - '0' + 130)
				dataIndex--
			} else {
				i := dmAsciiEncodation(text, textOffset+textIndex, 1, data, dataOffset+dataIndex, dataLength, -1)
				if i < 0 {
					return -1
				}
				if data[dataOffset+dataIndex] == dmUpperCase {
					dataIndex++
				}
				dataIndex++
			}
		}
		textIndex++
	}

	if textIndex != textLength {
		return -1
	}

	if !sizeFixed && (symbolIndex == len(text)-1 || symbolIndex < 0) {
		dataSize = 2<<28 // int.MaxValue equivalent
		for i := 0; i < len(dmSizes); i++ {
			if dmSizes[i].dataSize >= dataOffset+dataIndex+(3-pedi/6) {
				dataSize = dmSizes[i].dataSize
				break
			}
		}
	}

	if dataSize-dataOffset-dataIndex <= 2 && pedi >= 6 {
		if pedi != 18 && dataIndex+2-pedi/8 > dataLength {
			return -1
		}

		if pedi <= 12 {
			val := byte((edi >> 18) & 0x3f)
			if (val & 0x20) == 0 {
				val |= 0x40
			}
			data[dataOffset+dataIndex] = val + 1
			dataIndex++
		}

		if pedi <= 6 {
			val := byte((edi >> 12) & 0x3f)
			if (val & 0x20) == 0 {
				val |= 0x40
			}
			data[dataOffset+dataIndex] = val + 1
			dataIndex++
		}
	} else if !ascii {
		edi |= ('_' & 0x3f) << uint(pedi)
		if dataIndex+3-pedi/8 > dataLength {
			return -1
		}
		data[dataOffset+dataIndex] = byte(edi >> 16)
		dataIndex++
		if pedi <= 12 {
			data[dataOffset+dataIndex] = byte(edi >> 8)
			dataIndex++
		}
		if pedi <= 6 {
			data[dataOffset+dataIndex] = byte(edi)
			dataIndex++
		}
	}

	return dataIndex + dataOffset - origDataOffset
}

// -----------------------------------------------------------------------
// C40 / Text encodation
// -----------------------------------------------------------------------

func dmC40OrTextEncodation(text []byte, textOffset, textLength int, data []byte, dataOffset, dataLength int, c40 bool, symbolIndex, prevEnc, origDataOffset int) int {
	if textLength == 0 {
		return 0
	}

	var basic, shift2, shift3 string
	shift2 = dmShiftedC40AndText
	if c40 {
		basic = dmBaseC40
		shift3 = dmShiftedC40
	} else {
		basic = dmBaseText
		shift3 = dmShiftedText
	}

	mode := 2 // C40
	if !c40 {
		mode = 4 // Text
	}

	if symbolIndex != -1 {
		prevMode := -1
		if prevEnc == mode {
			prevMode = 1
		}
		return dmAsciiEncodation(text, textOffset, 1, data, dataOffset, dataLength, prevMode)
	}

	textIndex := 0
	dataIndex := 0

	if c40 {
		data[dataOffset+dataIndex] = dmLatchC40
	} else {
		data[dataOffset+dataIndex] = dmLatchText
	}
	dataIndex++

	encodedChars := make([]int, textLength*4+10)
	encIndex := 0
	last0 := 0
	last1 := 0

	for textIndex < textLength {
		if encIndex%3 == 0 {
			last0 = textIndex
			last1 = encIndex
		}

		c := int(text[textOffset+textIndex]) & 0xff
		textIndex++

		if c > 127 {
			c -= 128
			encodedChars[encIndex] = 1
			encIndex++
			encodedChars[encIndex] = 30
			encIndex++
		}

		if idx := indexOf(basic, rune(c)); idx >= 0 {
			encodedChars[encIndex] = idx + 3
			encIndex++
		} else if c < 32 {
			encodedChars[encIndex] = 0
			encIndex++
			encodedChars[encIndex] = c
			encIndex++
		} else if idx := indexOf(shift2, rune(c)); idx >= 0 {
			encodedChars[encIndex] = 1
			encIndex++
			encodedChars[encIndex] = idx
			encIndex++
		} else if idx := indexOf(shift3, rune(c)); idx >= 0 {
			encodedChars[encIndex] = 2
			encIndex++
			encodedChars[encIndex] = idx
			encIndex++
		}
	}

	if encIndex%3 != 0 {
		textIndex = last0
		encIndex = last1
	}

	if encIndex/3*2 > dataLength-2 {
		return -1
	}

	for i := 0; i < encIndex; i += 3 {
		a := 1600*encodedChars[i] + 40*encodedChars[i+1] + encodedChars[i+2] + 1
		data[dataOffset+dataIndex] = byte(a / 256)
		dataIndex++
		data[dataOffset+dataIndex] = byte(a)
		dataIndex++
	}

	if dataLength-dataIndex > 2 {
		data[dataOffset+dataIndex] = dmUnlatch
		dataIndex++
	}

	if symbolIndex < 0 && textLength > textIndex {
		i := dmAsciiEncodation(text, textOffset+textIndex, textLength-textIndex, data, dataOffset+dataIndex, dataLength-dataIndex, -1)
		return i + dataIndex + dataOffset - origDataOffset
	}

	return dataIndex + dataOffset - origDataOffset
}

// indexOf returns the index of r in s, or -1 if not found.
func indexOf(s string, r rune) int {
	for i, ch := range s {
		if ch == r {
			return i
		}
	}
	return -1
}

// -----------------------------------------------------------------------
// Placement algorithm (ECC200)
// -----------------------------------------------------------------------

// dmPlacementCache caches placement arrays keyed by nrow*1000+ncol.
var dmPlacementCache = map[int][]int16{}

func dmDoPlacement(nrow, ncol int) []int16 {
	key := nrow*1000 + ncol
	if pc, ok := dmPlacementCache[key]; ok {
		return pc
	}
	p := &dmPlacementState{
		nrow:  nrow,
		ncol:  ncol,
		array: make([]int16, nrow*ncol),
	}
	p.ecc200()
	dmPlacementCache[key] = p.array
	return p.array
}

type dmPlacementState struct {
	nrow  int
	ncol  int
	array []int16
}

func (p *dmPlacementState) module(row, col, chr, bit int) {
	if row < 0 {
		row += p.nrow
		col += 4 - ((p.nrow + 4) % 8)
	}
	if col < 0 {
		col += p.ncol
		row += 4 - ((p.ncol + 4) % 8)
	}
	p.array[row*p.ncol+col] = int16(8*chr + bit)
}

func (p *dmPlacementState) utah(row, col, chr int) {
	p.module(row-2, col-2, chr, 0)
	p.module(row-2, col-1, chr, 1)
	p.module(row-1, col-2, chr, 2)
	p.module(row-1, col-1, chr, 3)
	p.module(row-1, col, chr, 4)
	p.module(row, col-2, chr, 5)
	p.module(row, col-1, chr, 6)
	p.module(row, col, chr, 7)
}

func (p *dmPlacementState) corner1(chr int) {
	p.module(p.nrow-1, 0, chr, 0)
	p.module(p.nrow-1, 1, chr, 1)
	p.module(p.nrow-1, 2, chr, 2)
	p.module(0, p.ncol-2, chr, 3)
	p.module(0, p.ncol-1, chr, 4)
	p.module(1, p.ncol-1, chr, 5)
	p.module(2, p.ncol-1, chr, 6)
	p.module(3, p.ncol-1, chr, 7)
}

func (p *dmPlacementState) corner2(chr int) {
	p.module(p.nrow-3, 0, chr, 0)
	p.module(p.nrow-2, 0, chr, 1)
	p.module(p.nrow-1, 0, chr, 2)
	p.module(0, p.ncol-4, chr, 3)
	p.module(0, p.ncol-3, chr, 4)
	p.module(0, p.ncol-2, chr, 5)
	p.module(0, p.ncol-1, chr, 6)
	p.module(1, p.ncol-1, chr, 7)
}

func (p *dmPlacementState) corner3(chr int) {
	p.module(p.nrow-3, 0, chr, 0)
	p.module(p.nrow-2, 0, chr, 1)
	p.module(p.nrow-1, 0, chr, 2)
	p.module(0, p.ncol-2, chr, 3)
	p.module(0, p.ncol-1, chr, 4)
	p.module(1, p.ncol-1, chr, 5)
	p.module(2, p.ncol-1, chr, 6)
	p.module(3, p.ncol-1, chr, 7)
}

func (p *dmPlacementState) corner4(chr int) {
	p.module(p.nrow-1, 0, chr, 0)
	p.module(p.nrow-1, p.ncol-1, chr, 1)
	p.module(0, p.ncol-3, chr, 2)
	p.module(0, p.ncol-2, chr, 3)
	p.module(0, p.ncol-1, chr, 4)
	p.module(1, p.ncol-3, chr, 5)
	p.module(1, p.ncol-2, chr, 6)
	p.module(1, p.ncol-1, chr, 7)
}

func (p *dmPlacementState) ecc200() {
	// Initialize array.
	for k := range p.array {
		p.array[k] = 0
	}

	chr := 1
	row := 4
	col := 0

	for row < p.nrow || col < p.ncol {
		// Corner cases.
		if row == p.nrow && col == 0 {
			p.corner1(chr)
			chr++
		}
		if row == p.nrow-2 && col == 0 && p.ncol%4 != 0 {
			p.corner2(chr)
			chr++
		}
		if row == p.nrow-2 && col == 0 && p.ncol%8 == 4 {
			p.corner3(chr)
			chr++
		}
		if row == p.nrow+4 && col == 2 && p.ncol%8 == 0 {
			p.corner4(chr)
			chr++
		}

		// Sweep up.
		for row >= 0 && col < p.ncol {
			if row < p.nrow && col >= 0 && p.array[row*p.ncol+col] == 0 {
				p.utah(row, col, chr)
				chr++
			}
			row -= 2
			col += 2
		}
		row++
		col += 3

		// Sweep down.
		for row < p.nrow && col >= 0 {
			if row >= 0 && col < p.ncol && p.array[row*p.ncol+col] == 0 {
				p.utah(row, col, chr)
				chr++
			}
			row += 2
			col -= 2
		}
		row += 3
		col++
	}

	// Fix lower-right corner if untouched.
	if p.array[p.nrow*p.ncol-1] == 0 {
		p.array[p.nrow*p.ncol-1] = 1
		p.array[p.nrow*p.ncol-p.ncol-2] = 1
	}
}

// -----------------------------------------------------------------------
// Reed-Solomon ECC generation
// -----------------------------------------------------------------------

var dmLog = []int{
	0, 255, 1, 240, 2, 225, 241, 53, 3, 38, 226, 133, 242, 43, 54, 210,
	4, 195, 39, 114, 227, 106, 134, 28, 243, 140, 44, 23, 55, 118, 211, 234,
	5, 219, 196, 96, 40, 222, 115, 103, 228, 78, 107, 125, 135, 8, 29, 162,
	244, 186, 141, 180, 45, 99, 24, 49, 56, 13, 119, 153, 212, 199, 235, 91,
	6, 76, 220, 217, 197, 11, 97, 184, 41, 36, 223, 253, 116, 138, 104, 193,
	229, 86, 79, 171, 108, 165, 126, 145, 136, 34, 9, 74, 30, 32, 163, 84,
	245, 173, 187, 204, 142, 81, 181, 190, 46, 88, 100, 159, 25, 231, 50, 207,
	57, 147, 14, 67, 120, 128, 154, 248, 213, 167, 200, 63, 236, 110, 92, 176,
	7, 161, 77, 124, 221, 102, 218, 95, 198, 90, 12, 152, 98, 48, 185, 179,
	42, 209, 37, 132, 224, 52, 254, 239, 117, 233, 139, 22, 105, 27, 194, 113,
	230, 206, 87, 158, 80, 189, 172, 203, 109, 175, 166, 62, 127, 247, 146, 66,
	137, 192, 35, 252, 10, 183, 75, 216, 31, 83, 33, 73, 164, 144, 85, 170,
	246, 65, 174, 61, 188, 202, 205, 157, 143, 169, 82, 72, 182, 215, 191, 251,
	47, 178, 89, 151, 101, 94, 160, 123, 26, 112, 232, 21, 51, 238, 208, 131,
	58, 69, 148, 18, 15, 16, 68, 17, 121, 149, 129, 19, 155, 59, 249, 70,
	214, 250, 168, 71, 201, 156, 64, 60, 237, 130, 111, 20, 93, 122, 177, 150,
}

var dmAlog = []int{
	1, 2, 4, 8, 16, 32, 64, 128, 45, 90, 180, 69, 138, 57, 114, 228,
	229, 231, 227, 235, 251, 219, 155, 27, 54, 108, 216, 157, 23, 46, 92, 184,
	93, 186, 89, 178, 73, 146, 9, 18, 36, 72, 144, 13, 26, 52, 104, 208,
	141, 55, 110, 220, 149, 7, 14, 28, 56, 112, 224, 237, 247, 195, 171, 123,
	246, 193, 175, 115, 230, 225, 239, 243, 203, 187, 91, 182, 65, 130, 41, 82,
	164, 101, 202, 185, 95, 190, 81, 162, 105, 210, 137, 63, 126, 252, 213, 135,
	35, 70, 140, 53, 106, 212, 133, 39, 78, 156, 21, 42, 84, 168, 125, 250,
	217, 159, 19, 38, 76, 152, 29, 58, 116, 232, 253, 215, 131, 43, 86, 172,
	117, 234, 249, 223, 147, 11, 22, 44, 88, 176, 77, 154, 25, 50, 100, 200,
	189, 87, 174, 113, 226, 233, 255, 211, 139, 59, 118, 236, 245, 199, 163, 107,
	214, 129, 47, 94, 188, 85, 170, 121, 242, 201, 191, 83, 166, 97, 194, 169,
	127, 254, 209, 143, 51, 102, 204, 181, 71, 142, 49, 98, 196, 165, 103, 206,
	177, 79, 158, 17, 34, 68, 136, 61, 122, 244, 197, 167, 99, 198, 161, 111,
	222, 145, 15, 30, 60, 120, 240, 205, 183, 67, 134, 33, 66, 132, 37, 74,
	148, 5, 10, 20, 40, 80, 160, 109, 218, 153, 31, 62, 124, 248, 221, 151,
	3, 6, 12, 24, 48, 96, 192, 173, 119, 238, 241, 207, 179, 75, 150, 1,
}

var dmPoly5 = []int{228, 48, 15, 111, 62}
var dmPoly7 = []int{23, 68, 144, 134, 240, 92, 254}
var dmPoly10 = []int{28, 24, 185, 166, 223, 248, 116, 255, 110, 61}
var dmPoly11 = []int{175, 138, 205, 12, 194, 168, 39, 245, 60, 97, 120}
var dmPoly12 = []int{41, 153, 158, 91, 61, 42, 142, 213, 97, 178, 100, 242}
var dmPoly14 = []int{156, 97, 192, 252, 95, 9, 157, 119, 138, 45, 18, 186, 83, 185}
var dmPoly18 = []int{83, 195, 100, 39, 188, 75, 66, 61, 241, 213, 109, 129, 94, 254, 225, 48, 90, 188}
var dmPoly20 = []int{15, 195, 244, 9, 233, 71, 168, 2, 188, 160, 153, 145, 253, 79, 108, 82, 27, 174, 186, 172}
var dmPoly24 = []int{52, 190, 88, 205, 109, 39, 176, 21, 155, 197, 251, 223, 155, 21, 5, 172, 254, 124, 12, 181, 184, 96, 50, 193}
var dmPoly28 = []int{211, 231, 43, 97, 71, 96, 103, 174, 37, 151, 170, 53, 75, 34, 249, 121, 17, 138, 110, 213, 141, 136, 120, 151, 233, 168, 93, 255}
var dmPoly36 = []int{245, 127, 242, 218, 130, 250, 162, 181, 102, 120, 84, 179, 220, 251, 80, 182, 229, 18, 2, 4, 68, 33, 101, 137, 95, 119, 115, 44, 175, 184, 59, 25, 225, 98, 81, 112}
var dmPoly42 = []int{77, 193, 137, 31, 19, 38, 22, 153, 247, 105, 122, 2, 245, 133, 242, 8, 175, 95, 100, 9, 167, 105, 214, 111, 57, 121, 21, 1, 253, 57, 54, 101, 248, 202, 69, 50, 150, 177, 226, 5, 9, 5}
var dmPoly48 = []int{245, 132, 172, 223, 96, 32, 117, 22, 238, 133, 238, 231, 205, 188, 237, 87, 191, 106, 16, 147, 118, 23, 37, 90, 170, 205, 131, 88, 120, 100, 66, 138, 186, 240, 82, 44, 176, 87, 187, 147, 160, 175, 69, 213, 92, 253, 225, 19}
var dmPoly56 = []int{175, 9, 223, 238, 12, 17, 220, 208, 100, 29, 175, 170, 230, 192, 215, 235, 150, 159, 36, 223, 38, 200, 132, 54, 228, 146, 218, 234, 117, 203, 29, 232, 144, 238, 22, 150, 201, 117, 62, 207, 164, 13, 137, 245, 127, 67, 247, 28, 155, 43, 203, 107, 233, 53, 143, 46}
var dmPoly62 = []int{242, 93, 169, 50, 144, 210, 39, 118, 202, 188, 201, 189, 143, 108, 196, 37, 185, 112, 134, 230, 245, 63, 197, 190, 250, 106, 185, 221, 175, 64, 114, 71, 161, 44, 147, 6, 27, 218, 51, 63, 87, 10, 40, 130, 188, 17, 163, 31, 176, 170, 4, 107, 232, 7, 94, 166, 224, 124, 86, 47, 11, 204}
var dmPoly68 = []int{220, 228, 173, 89, 251, 149, 159, 56, 89, 33, 147, 244, 154, 36, 73, 127, 213, 136, 248, 180, 234, 197, 158, 177, 68, 122, 93, 213, 15, 160, 227, 236, 66, 139, 153, 185, 202, 167, 179, 25, 220, 232, 96, 210, 231, 136, 223, 239, 181, 241, 59, 52, 172, 25, 49, 232, 211, 189, 64, 54, 108, 153, 132, 63, 96, 103, 82, 186}

func dmGetPoly(nc int) []int {
	switch nc {
	case 5:
		return dmPoly5
	case 7:
		return dmPoly7
	case 10:
		return dmPoly10
	case 11:
		return dmPoly11
	case 12:
		return dmPoly12
	case 14:
		return dmPoly14
	case 18:
		return dmPoly18
	case 20:
		return dmPoly20
	case 24:
		return dmPoly24
	case 28:
		return dmPoly28
	case 36:
		return dmPoly36
	case 42:
		return dmPoly42
	case 48:
		return dmPoly48
	case 56:
		return dmPoly56
	case 62:
		return dmPoly62
	case 68:
		return dmPoly68
	}
	return nil
}

func dmReedSolomonBlock(wd []byte, nd int, ncout []byte, nc int, c []int) {
	for i := 0; i <= nc; i++ {
		ncout[i] = 0
	}
	for i := 0; i < nd; i++ {
		k := int(ncout[0]^wd[i]) & 0xff
		for j := 0; j < nc; j++ {
			v := byte(0)
			if k != 0 {
				v = byte(dmAlog[(dmLog[k]+dmLog[c[nc-j-1]])%255])
			}
			ncout[j] = ncout[j+1] ^ v
		}
	}
}

func dmGenerateECC(wd []byte, nd, dataBlock, nc int) {
	blocks := (nd + 2) / dataBlock
	buf := make([]byte, 256)
	ecc := make([]byte, 256)
	c := dmGetPoly(nc)
	if c == nil {
		return
	}
	for b := 0; b < blocks; b++ {
		p := 0
		for n := b; n < nd; n += blocks {
			buf[p] = wd[n]
			p++
		}
		dmReedSolomonBlock(buf, p, ecc, nc, c)
		p = 0
		for n := b; n < nc*blocks; n += blocks {
			wd[nd+n] = ecc[p]
			p++
		}
	}
}

// -----------------------------------------------------------------------
// GetMatrix implementation for DataMatrixBarcode
// -----------------------------------------------------------------------

// GetMatrix encodes the barcode text and returns a boolean matrix where
// true = dark module, false = light module.
// Returns (matrix, rows, cols). Returns (nil, 0, 0) on error.
func (d *DataMatrixBarcode) GetMatrix() ([][]bool, int, int) {
	text := d.encodedText
	if text == "" {
		return nil, 0, 0
	}
	return dmGetMatrix([]byte(text))
}

// GetMatrix encodes the GS1 DataMatrix barcode (with FNC1 prefix) and returns
// a boolean matrix where true = dark module, false = light module.
// Returns (matrix, rows, cols). Returns (nil, 0, 0) on error.
func (g *GS1DatamatrixBarcode) GetMatrix() ([][]bool, int, int) {
	text := g.encodedText
	if text == "" {
		return nil, 0, 0
	}
	// GS1 DataMatrix prepends FNC1 character (byte 232 = 0xE8).
	data := make([]byte, 1+len(text))
	data[0] = 232
	copy(data[1:], []byte(text))
	return dmGetMatrix(data)
}

// dmGetMatrix is the shared implementation for both DataMatrix types.
func dmGetMatrix(data []byte) ([][]bool, int, int) {
	imgBytes, height, width, err := dmGenerate(data)
	if err != nil {
		return nil, 0, 0
	}
	xByte := (width + 7) / 8
	matrix := make([][]bool, height)
	for y := 0; y < height; y++ {
		row := make([]bool, width)
		for x := 0; x < width; x++ {
			b := int(imgBytes[y*xByte+x/8]) & 0xff
			b <<= uint(x % 8)
			row[x] = (b & 0x80) != 0
		}
		matrix[y] = row
	}
	return matrix, height, width
}
