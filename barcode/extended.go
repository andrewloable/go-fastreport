package barcode

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	boombarcode "github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
)

// extended.go implements the barcode types not yet present in barcode.go:
//   - GS1-128 / GS1 DataBar   (GS1Barcode)
//   - IntelligentMail (IMb)   (IntelligentMailBarcode)
//   - MSI Modified Plessey    (MSIBarcode)
//   - MaxiCode                (MaxiCodeBarcode)
//   - Pharmacode              (PharmacodeBarcode)
//   - Plessey                 (PlesseyBarcode)
//   - POSTNET                 (PostNetBarcode)
//   - Swiss QR Code           (SwissQRBarcode)

// ── Additional BarcodeType constants ─────────────────────────────────────────

const (
	BarcodeTypeIntelligentMail BarcodeType = "IntelligentMail"
	BarcodeTypePharmacode      BarcodeType = "Pharmacode"
	BarcodeTypePlessey         BarcodeType = "Plessey"
	BarcodeTypePostNet         BarcodeType = "PostNet"
	BarcodeTypeSwissQR         BarcodeType = "SwissQR"
)

// ── GS1Barcode ────────────────────────────────────────────────────────────────

// GS1Barcode encodes GS1-128 (Code128 with GS1 application identifiers).
// The FNC1 start character is prepended automatically when encoding.
type GS1Barcode struct {
	BaseBarcodeImpl
}

// NewGS1Barcode creates a GS1Barcode.
func NewGS1Barcode() *GS1Barcode {
	return &GS1Barcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeGS1_128)}
}

// DefaultValue returns a sample GS1-128 value with an AI (01 = GTIN).
func (g *GS1Barcode) DefaultValue() string { return "(01)12345678901231" }

// Encode encodes text as GS1-128 using Code128 with the GS1 FNC1 prefix.
// Application identifiers in parentheses are stripped before encoding.
func (g *GS1Barcode) Encode(text string) error {
	// Strip parenthesised AI notation: "(01)123" → "01123".
	cleaned := stripGS1Parens(text)
	// Prepend GS1 FNC1 start (represented as "\x1d" = ASCII GS / Group Separator).
	encoded := "\u00f1" + cleaned // \u00f1 is FNC1 in Code128B context
	bc, err := code128.Encode(encoded)
	if err != nil {
		// Fall back to plain text.
		bc, err = code128.Encode(cleaned)
		if err != nil {
			return fmt.Errorf("gs1 encode: %w", err)
		}
	}
	g.encodedText = text
	g.encoded = bc
	return nil
}

// Render returns the GS1-128 barcode as a raster image.
func (g *GS1Barcode) Render(width, height int) (image.Image, error) {
	if g.encoded == nil {
		if err := g.Encode(g.encodedText); err != nil {
			return nil, err
		}
	}
	return boombarcode.Scale(g.encoded, width, height)
}

func stripGS1Parens(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '(' {
			// Skip until closing paren.
			for i < len(s) && s[i] != ')' {
				i++
			}
		} else {
			sb.WriteByte(s[i])
		}
	}
	return sb.String()
}

// ── IntelligentMailBarcode ────────────────────────────────────────────────────

// IntelligentMailBarcode represents the USPS Intelligent Mail barcode (IMb).
// Full encoding requires a 20/25/29/31-digit string; this stub stores the value
// for round-trip fidelity and renders a placeholder.
type IntelligentMailBarcode struct {
	BaseBarcodeImpl
}

// NewIntelligentMailBarcode creates an IntelligentMailBarcode.
func NewIntelligentMailBarcode() *IntelligentMailBarcode {
	return &IntelligentMailBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeIntelligentMail)}
}

// DefaultValue returns a sample 31-digit IMb value.
func (b *IntelligentMailBarcode) DefaultValue() string { return "01234567094987654321012345678" }

// Encode validates the IMb digit string (must be 20, 25, 29, or 31 digits).
func (b *IntelligentMailBarcode) Encode(text string) error {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, text)
	switch len(digits) {
	case 20, 25, 29, 31:
		b.encodedText = text
		return nil
	default:
		return fmt.Errorf("intelligentmail: expected 20/25/29/31 digits, got %d", len(digits))
	}
}

// Render is implemented in intelligentmail.go.

// ── MSIBarcode ────────────────────────────────────────────────────────────────

// MSICheckDigit specifies which check-digit algorithm MSI uses.
type MSICheckDigit int

const (
	MSICheckDigitNone   MSICheckDigit = iota
	MSICheckDigitMod10                 // single Mod-10
	MSICheckDigitMod1010               // double Mod-10
	MSICheckDigitMod11                 // Mod-11
)

// MSIBarcode implements the MSI Modified Plessey barcode.
type MSIBarcode struct {
	BaseBarcodeImpl
	CheckDigit MSICheckDigit
}

// NewMSIBarcode creates an MSIBarcode.
func NewMSIBarcode() *MSIBarcode {
	return &MSIBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeMSI),
		CheckDigit:      MSICheckDigitMod10,
	}
}

// DefaultValue returns a sample MSI value.
func (b *MSIBarcode) DefaultValue() string { return "12345" }

// Encode validates and stores the MSI value (digits only).
func (b *MSIBarcode) Encode(text string) error {
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("msi: only digits allowed, found %q", ch)
		}
	}
	b.encodedText = text
	return nil
}

// Render draws the MSI barcode as a 1-D bar pattern image.
func (b *MSIBarcode) Render(width, height int) (image.Image, error) {
	if b.encodedText == "" {
		return nil, fmt.Errorf("msi: not encoded")
	}
	// Build bar pattern: MSI uses 4 bars per digit.
	bits := msiEncode(b.encodedText, b.CheckDigit)
	return renderBitPattern(bits, width, height, color.Black, color.White), nil
}

// msiEncode converts digits to MSI bar bits (0=narrow, 1=wide).
// Each digit encodes as 4 bit-pairs (8 elements total). Start/stop bars added.
func msiEncode(digits string, cd MSICheckDigit) []bool {
	if cd == MSICheckDigitMod10 {
		digits += msiMod10(digits)
	}
	// Start: narrow-narrow (2 bars).
	var bits []bool
	bits = append(bits, true, true) // start bar
	for _, d := range digits {
		n := int(d - '0')
		for bit := 3; bit >= 0; bit-- {
			if (n>>bit)&1 == 1 {
				// Wide bar = narrow+wide = 110 , narrow space = 0
				bits = append(bits, true, true, false)
			} else {
				// Narrow bar = narrow+narrow = 10
				bits = append(bits, true, false, false)
			}
		}
	}
	// Stop: narrow, wide, narrow = 100+110+10
	bits = append(bits, true, false, false, true, true, false)
	return bits
}

func msiMod10(digits string) string {
	sum := 0
	odd := true
	for i := len(digits) - 1; i >= 0; i-- {
		d := int(digits[i] - '0')
		if odd {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		odd = !odd
	}
	check := (10 - (sum % 10)) % 10
	return fmt.Sprintf("%d", check)
}

// ── MaxiCodeBarcode ───────────────────────────────────────────────────────────

// MaxiCodeBarcode represents the MaxiCode 2D barcode used by UPS.
// This is a complex 2D symbol; the stub stores the data for round-trip fidelity.
type MaxiCodeBarcode struct {
	BaseBarcodeImpl
	Mode int // MaxiCode mode (2-6)
}

// NewMaxiCodeBarcode creates a MaxiCodeBarcode in mode 2 (structured carrier).
func NewMaxiCodeBarcode() *MaxiCodeBarcode {
	return &MaxiCodeBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeMaxiCode),
		Mode:            4,
	}
}

// DefaultValue returns a sample MaxiCode payload.
func (b *MaxiCodeBarcode) DefaultValue() string { return "MaxiCode Test" }

// Encode and Render are implemented in maxicode.go.

// ── PharmacodeBarcode ─────────────────────────────────────────────────────────

// PharmacodeBarcode implements the Pharmacode (pharmaceutical) barcode.
// Pharmacode encodes an integer 3–131070 as alternating wide/narrow bars.
type PharmacodeBarcode struct {
	BaseBarcodeImpl
	TwoTrack bool // use two-track Pharmacode variant
}

// NewPharmacodeBarcode creates a PharmacodeBarcode.
func NewPharmacodeBarcode() *PharmacodeBarcode {
	return &PharmacodeBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypePharmacode)}
}

// DefaultValue returns a sample Pharmacode value.
func (b *PharmacodeBarcode) DefaultValue() string { return "1234" }

// Encode validates and stores the Pharmacode value (integer 3–131070).
func (b *PharmacodeBarcode) Encode(text string) error {
	var n int
	_, err := fmt.Sscanf(text, "%d", &n)
	if err != nil || n < 3 || n > 131070 {
		return fmt.Errorf("pharmacode: value must be integer 3–131070, got %q", text)
	}
	b.encodedText = text
	return nil
}

// Render draws the Pharmacode barcode pattern.
func (b *PharmacodeBarcode) Render(width, height int) (image.Image, error) {
	if b.encodedText == "" {
		return nil, fmt.Errorf("pharmacode: not encoded")
	}
	var n int
	fmt.Sscanf(b.encodedText, "%d", &n) //nolint:errcheck
	bits := pharmacodeEncode(n)
	return renderBitPattern(bits, width, height, color.Black, color.White), nil
}

// pharmacodeEncode encodes an integer as Pharmacode bar widths.
// Returns true for wide bar, false for narrow bar.
func pharmacodeEncode(n int) []bool {
	var bars []bool
	for n > 0 {
		if n%2 == 0 {
			bars = append(bars, true) // wide
			n = (n - 2) / 2
		} else {
			bars = append(bars, false) // narrow
			n = (n - 1) / 2
		}
	}
	// Reverse (MSB first).
	for i, j := 0, len(bars)-1; i < j; i, j = i+1, j-1 {
		bars[i], bars[j] = bars[j], bars[i]
	}
	return bars
}

// ── PlesseyBarcode ────────────────────────────────────────────────────────────

// PlesseyBarcode implements the Plessey barcode (hex digit encoding).
type PlesseyBarcode struct {
	BaseBarcodeImpl
}

// NewPlesseyBarcode creates a PlesseyBarcode.
func NewPlesseyBarcode() *PlesseyBarcode {
	return &PlesseyBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypePlessey)}
}

// DefaultValue returns a sample Plessey value.
func (b *PlesseyBarcode) DefaultValue() string { return "1234" }

// Encode validates that text contains only hex digits (0-9, A-F).
func (b *PlesseyBarcode) Encode(text string) error {
	for _, ch := range strings.ToUpper(text) {
		if !((ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F')) {
			return fmt.Errorf("plessey: only hex digits allowed, found %q", ch)
		}
	}
	b.encodedText = strings.ToUpper(text)
	return nil
}

// Render draws the Plessey barcode as a 1-D bar image with CRC.
func (b *PlesseyBarcode) Render(width, height int) (image.Image, error) {
	if b.encodedText == "" {
		return nil, fmt.Errorf("plessey: not encoded")
	}
	bits, err := plesseyEncode(b.encodedText)
	if err != nil {
		return nil, err
	}
	return renderBitPattern(bits, width, height, color.Black, color.White), nil
}

// plesseyEncode generates the Plessey bar/space bit array including CRC.
// Ported from FastReport BarcodePlessey.cs / ZXing PlesseyWriter.
func plesseyEncode(text string) ([]bool, error) {
	const alphabet = "0123456789ABCDEF"

	// Lookup table: index of each character in alphabet.
	indexOf := func(ch rune) (int, bool) {
		for i, a := range alphabet {
			if a == ch {
				return i, true
			}
		}
		return 0, false
	}

	// Bar-width tables (unit widths alternating bar/space starting with bar).
	startWidths := []int{14, 11, 14, 11, 5, 20, 14, 11}
	endWidths := []int{20, 5, 20, 5, 14, 11, 14, 11}
	terminationWidths := []int{25}
	numberWidths := [16][]int{
		{5, 20, 5, 20, 5, 20, 5, 20},     // 0
		{14, 11, 5, 20, 5, 20, 5, 20},    // 1
		{5, 20, 14, 11, 5, 20, 5, 20},    // 2
		{14, 11, 14, 11, 5, 20, 5, 20},   // 3
		{5, 20, 5, 20, 14, 11, 5, 20},    // 4
		{14, 11, 5, 20, 14, 11, 5, 20},   // 5
		{5, 20, 14, 11, 14, 11, 5, 20},   // 6
		{14, 11, 14, 11, 14, 11, 5, 20},  // 7
		{5, 20, 5, 20, 5, 20, 14, 11},    // 8
		{14, 11, 5, 20, 5, 20, 14, 11},   // 9
		{5, 20, 14, 11, 5, 20, 14, 11},   // A
		{14, 11, 14, 11, 5, 20, 14, 11},  // B
		{5, 20, 5, 20, 14, 11, 14, 11},   // C
		{14, 11, 5, 20, 14, 11, 14, 11},  // D
		{5, 20, 14, 11, 14, 11, 14, 11},  // E
		{14, 11, 14, 11, 14, 11, 14, 11}, // F
	}
	crcGrid := []byte{1, 1, 1, 1, 0, 1, 0, 0, 1}
	crc0Widths := []int{5, 20}
	crc1Widths := []int{14, 11}

	n := len(text)
	// Calculate the total bit array capacity.
	codeWidth := 100 + 100 + n*100 + 25*8 + 25 + 100 + 100
	result := make([]bool, codeWidth)

	// CRC buffer: 4 bits per character + 8 bits CRC remainder.
	crcBuf := make([]byte, 4*n+8)
	crcPos := 0

	appendPattern := func(pos int, widths []int, startBlack bool) int {
		black := startBlack
		added := 0
		for _, w := range widths {
			for j := 0; j < w; j++ {
				if pos+j < len(result) {
					result[pos+j] = black
				}
			}
			pos += w
			added += w
			black = !black
		}
		return added
	}

	pos := 100
	// Start pattern.
	pos += appendPattern(pos, startWidths, true)

	// Data + CRC buffer population.
	for _, ch := range text {
		idx, ok := indexOf(ch)
		if !ok {
			return nil, fmt.Errorf("plessey: invalid character %q", ch)
		}
		pos += appendPattern(pos, numberWidths[idx], true)
		crcBuf[crcPos] = byte(idx & 1)
		crcBuf[crcPos+1] = byte((idx >> 1) & 1)
		crcBuf[crcPos+2] = byte((idx >> 2) & 1)
		crcBuf[crcPos+3] = byte((idx >> 3) & 1)
		crcPos += 4
	}

	// CRC polynomial division.
	for i := 0; i < 4*n; i++ {
		if crcBuf[i] != 0 {
			for j := 0; j < 9; j++ {
				crcBuf[i+j] ^= crcGrid[j]
			}
		}
	}

	// Append CRC bits.
	for i := 0; i < 8; i++ {
		if crcBuf[n*4+i] == 0 {
			pos += appendPattern(pos, crc0Widths, true)
		} else {
			pos += appendPattern(pos, crc1Widths, true)
		}
	}

	// Termination bar.
	pos += appendPattern(pos, terminationWidths, true)
	// End pattern.
	appendPattern(pos, endWidths, false)

	// Trim result to actual used length.
	return result[:pos+sum(endWidths)], nil
}

func sum(s []int) int {
	n := 0
	for _, v := range s {
		n += v
	}
	return n
}

// ── PostNetBarcode ────────────────────────────────────────────────────────────

// PostNetBarcode implements the USPS POSTNET barcode.
// POSTNET encodes ZIP+4+delivery-point codes as tall/short bar patterns.
type PostNetBarcode struct {
	BaseBarcodeImpl
}

// NewPostNetBarcode creates a PostNetBarcode.
func NewPostNetBarcode() *PostNetBarcode {
	return &PostNetBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypePostNet)}
}

// DefaultValue returns a sample 5-digit ZIP code.
func (b *PostNetBarcode) DefaultValue() string { return "90210" }

// Encode validates POSTNET input (5, 9, or 11 digits).
func (b *PostNetBarcode) Encode(text string) error {
	digits := strings.ReplaceAll(text, "-", "")
	for _, ch := range digits {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("postnet: only digits allowed, found %q", ch)
		}
	}
	switch len(digits) {
	case 5, 9, 11:
		b.encodedText = digits
		return nil
	default:
		return fmt.Errorf("postnet: expected 5, 9, or 11 digits, got %d", len(digits))
	}
}

// Render draws the POSTNET tall/short bar pattern.
func (b *PostNetBarcode) Render(width, height int) (image.Image, error) {
	if b.encodedText == "" {
		return nil, fmt.Errorf("postnet: not encoded")
	}
	bits := postnetEncode(b.encodedText)
	return renderBitPattern(bits, width, height, color.Black, color.White), nil
}

// postnetEncode converts digits to a POSTNET tall/short bit pattern.
// true = tall bar (half-height top extension), false = short bar.
func postnetEncode(digits string) []bool {
	// POSTNET digit encoding (5 bars each, sum=2 tall bars).
	postnetDigits := [10][5]bool{
		{false, false, true, true, false}, // 0 — non-standard, sum=2
		{false, false, false, true, true}, // 1
		{false, false, true, false, true}, // 2
		{false, false, true, true, false}, // 3 (same as 0 by spec; 0 is 11000)
		{false, true, false, false, true}, // 4
		{false, true, false, true, false}, // 5
		{false, true, true, false, false}, // 6
		{true, false, false, false, true}, // 7
		{true, false, false, true, false}, // 8
		{true, false, true, false, false}, // 9
	}
	// Override 0 correctly: tall bars at positions 0,1 (11000).
	postnetDigits[0] = [5]bool{true, true, false, false, false}

	var bits []bool
	bits = append(bits, true) // start frame bar
	sum := 0
	for _, d := range digits {
		n := int(d - '0')
		bars := postnetDigits[n]
		for _, b := range bars {
			bits = append(bits, b)
			if b {
				sum++
			}
		}
	}
	// Check digit: make total tall bars divisible by 10.
	check := (10 - (sum % 10)) % 10
	for _, b := range postnetDigits[check] {
		bits = append(bits, b)
	}
	bits = append(bits, true) // end frame bar
	return bits
}

// ── SwissQRBarcode ────────────────────────────────────────────────────────────

// SwissQRBarcode encodes a Swiss QR Code payment slip as a QR code.
// The payload is structured per the Swiss Payment Standards specification.
type SwissQRBarcode struct {
	BaseBarcodeImpl

	// Payment fields (subset of Swiss QR specification).
	IBAN          string
	Amount        string
	Currency      string // CHF or EUR
	CreditorName  string
	CreditorIBAN  string
	Reference     string
	RefType       string // QRR, SCOR, or NON
	Message       string
}

// NewSwissQRBarcode creates a SwissQRBarcode.
func NewSwissQRBarcode() *SwissQRBarcode {
	return &SwissQRBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeSwissQR),
		Currency:        "CHF",
		RefType:         "NON",
	}
}

// DefaultValue returns a minimal Swiss QR payload sample.
func (b *SwissQRBarcode) DefaultValue() string {
	return "SPC\r\n0200\r\n1\r\nCH5604835012345678009\r\nS\r\nMaxMuster\r\nMusterstrasse\r\n1\r\n8000\r\nZürich\r\nCH\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n10.00\r\nCHF\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\nNON\r\n\r\n\r\nEPD"
}

// Encode encodes the Swiss QR payload string as a QR code.
func (b *SwissQRBarcode) Encode(text string) error {
	if text == "" {
		text = b.buildPayload()
	}
	bc, err := qr.Encode(text, qr.M, qr.Auto)
	if err != nil {
		return fmt.Errorf("swissqr encode: %w", err)
	}
	b.encodedText = text
	b.encoded = bc
	return nil
}

// Render renders the Swiss QR Code.
func (b *SwissQRBarcode) Render(width, height int) (image.Image, error) {
	if b.encoded == nil {
		if err := b.Encode(b.encodedText); err != nil {
			return nil, err
		}
	}
	return boombarcode.Scale(b.encoded, width, height)
}

// buildPayload assembles the structured Swiss QR payload from individual fields.
func (b *SwissQRBarcode) buildPayload() string {
	currency := b.Currency
	if currency == "" {
		currency = "CHF"
	}
	refType := b.RefType
	if refType == "" {
		refType = "NON"
	}
	return strings.Join([]string{
		"SPC", "0200", "1",
		b.IBAN, "S",
		b.CreditorName, "", "", "", "", "CH",
		"", "", "", "", "", "", "",
		b.Amount, currency,
		"", "", "", "", "", "", "",
		refType, b.Reference, "",
		b.Message, "EPD",
	}, "\r\n")
}

// ── helpers ───────────────────────────────────────────────────────────────────

// placeholderImage returns a uniform light-grey image of the given dimensions.
// Used when a barcode type does not have a full renderer yet.
func placeholderImage(width, height int) image.Image {
	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 50
	}
	img := image.NewGray(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetGray(x, y, color.Gray{Y: 200})
		}
	}
	return img
}

// renderBitPattern renders a 1-D bar pattern (true=bar, false=space) as a
// raster image scaled to width×height. Bars are evenly distributed.
func renderBitPattern(bits []bool, width, height int, barColor, spaceColor color.Color) image.Image {
	if len(bits) == 0 || width <= 0 || height <= 0 {
		return placeholderImage(width, height)
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	barW := float64(width) / float64(len(bits))
	for i, bar := range bits {
		x0 := int(float64(i) * barW)
		x1 := int(float64(i+1) * barW)
		if x1 > width {
			x1 = width
		}
		c := spaceColor
		if bar {
			c = barColor
		}
		for y := 0; y < height; y++ {
			for x := x0; x < x1; x++ {
				img.Set(x, y, c)
			}
		}
	}
	return img
}
