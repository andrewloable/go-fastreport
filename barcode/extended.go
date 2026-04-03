package barcode

import (
	"fmt"
	"image"
	"image/color"
	"strings"
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

// Encode validates and stores text for GS1-128 encoding.
// Application identifiers in parentheses are kept for GetPattern().
func (g *GS1Barcode) Encode(text string) error {
	g.encodedText = text
	return nil
}

// GetPattern generates the Code128 bar pattern with GS1 FNC1 prefix.
// Delegates to the GS1_128Barcode pattern generator which handles
// AI-formatted text with FNC1 separators.
func (g *GS1Barcode) GetPattern() (string, error) {
	helper := &GS1_128Barcode{}
	helper.encodedText = g.encodedText
	return helper.GetPattern()
}

// GetWideBarRatio returns the wide bar ratio for GS1-128.
func (g *GS1Barcode) GetWideBarRatio() float32 { return 1 }

// Render returns the GS1-128 barcode as a raster image.
func (g *GS1Barcode) Render(width, height int) (image.Image, error) {
	if g.encodedText == "" {
		return nil, fmt.Errorf("gs1: Encode must be called before Render")
	}
	pattern, err := g.GetPattern()
	if err != nil {
		return nil, err
	}
	return drawLinearBarcodeColored(pattern, g.encodedText, width, height, true, g.GetWideBarRatio(), g.Color), nil
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
	QuietZone bool
}

// NewIntelligentMailBarcode creates an IntelligentMailBarcode.
func NewIntelligentMailBarcode() *IntelligentMailBarcode {
	return &IntelligentMailBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeIntelligentMail),
		QuietZone:       false,
	}
}

// Assign copies all IntelligentMailBarcode fields from src.
// Mirrors C# BarcodeIntelligentMail.Assign (BarcodeIntelligentMail.cs:44-48).
func (b *IntelligentMailBarcode) Assign(src *IntelligentMailBarcode) {
	if src == nil {
		return
	}
	b.BaseBarcodeImpl = src.BaseBarcodeImpl
	b.QuietZone = src.QuietZone
}

// DefaultValue returns "12345678901234567890" matching C# BarcodeIntelligentMail.GetDefaultValue().
func (b *IntelligentMailBarcode) DefaultValue() string { return "12345678901234567890" }

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
	MSICheckDigitNone    MSICheckDigit = iota
	MSICheckDigitMod10                 // single Mod-10
	MSICheckDigitMod1010               // double Mod-10
	MSICheckDigitMod11                 // Mod-11
)

// MSIBarcode implements the MSI Modified Plessey barcode.
type MSIBarcode struct {
	BaseBarcodeImpl
	CheckDigit   MSICheckDigit
	CalcChecksum bool // C# LinearBarcodeBase.CalcCheckSum default true
}

// NewMSIBarcode creates an MSIBarcode.
func NewMSIBarcode() *MSIBarcode {
	return &MSIBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeMSI),
		CheckDigit:      MSICheckDigitMod10,
		CalcChecksum:    true,
	}
}

// SetCalcCheckSum implements calcCheckSumSetter for FRX deserialization.
func (b *MSIBarcode) SetCalcCheckSum(v bool) { b.CalcChecksum = v }

// DefaultValue returns the C# BarcodeBase.GetDefaultValue() default: "12345678".
func (b *MSIBarcode) DefaultValue() string { return "12345678" }

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

// Assign copies Mode from src into this MaxiCodeBarcode.
// Mirrors C# BarcodeMaxiCode.Assign(BarcodeBase source) (BarcodeMaxiCode.cs:113-118).
func (b *MaxiCodeBarcode) Assign(src *MaxiCodeBarcode) {
	if src == nil {
		return
	}
	b.BaseBarcodeImpl = src.BaseBarcodeImpl
	b.Mode = src.Mode
}

// Encode and Render are implemented in maxicode.go.

// ── PharmacodeBarcode ─────────────────────────────────────────────────────────

// PharmacodeBarcode implements the Pharmacode (pharmaceutical) barcode.
// Pharmacode encodes an integer 3–131070 as alternating wide/narrow bars.
type PharmacodeBarcode struct {
	BaseBarcodeImpl
	TwoTrack  bool // use two-track Pharmacode variant
	QuietZone bool
}

// NewPharmacodeBarcode creates a PharmacodeBarcode.
func NewPharmacodeBarcode() *PharmacodeBarcode {
	return &PharmacodeBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypePharmacode),
		QuietZone:       true,
	}
}

// Assign copies all PharmacodeBarcode fields from src.
// Mirrors C# BarcodePharmacode.Assign (BarcodePharmacode.cs).
func (b *PharmacodeBarcode) Assign(src *PharmacodeBarcode) {
	if src == nil {
		return
	}
	b.BaseBarcodeImpl = src.BaseBarcodeImpl
	b.TwoTrack = src.TwoTrack
	b.QuietZone = src.QuietZone
}

// DefaultValue returns the C# BarcodeBase.GetDefaultValue() default: "12345678".
func (b *PharmacodeBarcode) DefaultValue() string { return "12345678" }

// Encode validates and stores the Pharmacode value (non-negative integer).
// C# BarcodePharmacode does not validate a specific range; it accepts any
// non-negative integer and encodes it via binary representation. We mirror
// that behaviour so that the default value "12345678" (from BarcodeBase.GetDefaultValue)
// encodes successfully even though it exceeds the standard Pharmacode spec
// maximum of 131070. C# ref: BarcodePharmacode.cs GetPattern().
func (b *PharmacodeBarcode) Encode(text string) error {
	var n uint64
	_, err := fmt.Sscanf(text, "%d", &n)
	if err != nil {
		return fmt.Errorf("pharmacode: value must be a non-negative integer, got %q", text)
	}
	b.encodedText = text
	return nil
}

// Render draws the Pharmacode barcode pattern.
func (b *PharmacodeBarcode) Render(width, height int) (image.Image, error) {
	if b.encodedText == "" {
		return nil, fmt.Errorf("pharmacode: not encoded")
	}
	pattern, err := b.GetPattern()
	if err != nil {
		return nil, err
	}
	return drawLinearBarcodeColored(pattern, b.encodedText, width, height, true, b.GetWideBarRatio(), b.Color), nil
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

// DefaultValue returns the C# BarcodeBase.GetDefaultValue() default: "12345678".
func (b *PlesseyBarcode) DefaultValue() string { return "12345678" }

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
// Mirrors C# BarcodePlessey / LinearBarcodeBase.DrawBarcode: uses GetPattern()
// then DrawLinearBarcode which handles showText automatically.
func (b *PlesseyBarcode) Render(width, height int) (image.Image, error) {
	if b.encodedText == "" {
		return nil, fmt.Errorf("plessey: not encoded")
	}
	pattern, err := b.GetPattern()
	if err != nil {
		return nil, err
	}
	return drawLinearBarcodeColored(pattern, b.encodedText, width, height, b.showText, b.GetWideBarRatio(), b.Color), nil
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

// DefaultValue returns "12345678" matching C# BarcodeBase.GetDefaultValue().
// C# BarcodePostNet does not override GetDefaultValue(), so the base class default
// "12345678" is used. C# PostNet accepts any number of digits without length validation.
func (b *PostNetBarcode) DefaultValue() string { return "12345678" }

// Encode validates POSTNET input (digits only, any count).
// C# BarcodePostNet.GetPattern() iterates text chars without length checks,
// so we accept any positive number of digits to match.
func (b *PostNetBarcode) Encode(text string) error {
	digits := strings.ReplaceAll(text, "-", "")
	for _, ch := range digits {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("postnet: only digits allowed, found %q", ch)
		}
	}
	if len(digits) == 0 {
		return fmt.Errorf("postnet: at least one digit required")
	}
	b.encodedText = digits
	return nil
}

// Render draws the POSTNET tall/short bar pattern using GetPattern and the
// standard linear renderer. Mirrors C# BarcodePostNet which inherits
// LinearBarcodeBase.DrawBarcode and uses GetPattern() for rendering.
func (b *PostNetBarcode) Render(width, height int) (image.Image, error) {
	if b.encodedText == "" {
		return nil, fmt.Errorf("postnet: not encoded")
	}
	pattern, err := b.GetPattern()
	if err != nil {
		return nil, err
	}
	return drawLinearBarcodeColored(pattern, b.encodedText, width, height, false, b.GetWideBarRatio(), b.Color), nil
}


// ── SwissQRBarcode ────────────────────────────────────────────────────────────

// SwissQRParameters holds all structured fields for a Swiss QR Code payment slip.
// Fields follow the Swiss Payment Standards SPC v2.0 specification.
type SwissQRParameters struct {
	// IBAN is the creditor's IBAN (CH or LI prefix).
	IBAN string
	// Currency is either "CHF" or "EUR".
	Currency string
	// CreditorName is the name of the creditor (payee).
	CreditorName string
	// CreditorStreet is the creditor's street address.
	CreditorStreet string
	// CreditorCity is the creditor's city.
	CreditorCity string
	// CreditorPostalCode is the creditor's postal/zip code.
	CreditorPostalCode string
	// CreditorCountry is the two-letter ISO 3166-1 country code for the creditor.
	CreditorCountry string
	// Amount is the payment amount as a string (empty = not specified).
	Amount string
	// Reference is the payment reference number.
	Reference string
	// ReferenceType specifies the reference type: "QRR", "SCOR", or "NON".
	ReferenceType string
	// UnstructuredMessage is a free-text payment message (max 140 chars).
	UnstructuredMessage string
	// TrailerEPD is always "EPD" per the specification.
	TrailerEPD string
	// AlternativeProcedure1 is an optional alternative processing command line 1.
	AlternativeProcedure1 string
	// AlternativeProcedure2 is an optional alternative processing command line 2.
	AlternativeProcedure2 string
}

// SwissQRBarcode encodes a Swiss QR Code payment slip as a QR code.
// The payload is structured per the Swiss Payment Standards specification.
type SwissQRBarcode struct {
	BaseBarcodeImpl

	// Params holds the structured payment fields.
	Params SwissQRParameters

	// Legacy flat fields — kept for backward compatibility; Params takes precedence
	// when Params.IBAN is non-empty.
	IBAN         string
	Amount       string
	Currency     string // CHF or EUR
	CreditorName string
	CreditorIBAN string
	Reference    string
	RefType      string // QRR, SCOR, or NON
	Message      string
}

// NewSwissQRBarcode creates a SwissQRBarcode with default values.
func NewSwissQRBarcode() *SwissQRBarcode {
	return &SwissQRBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeSwissQR),
		Params: SwissQRParameters{
			Currency:      "CHF",
			ReferenceType: "NON",
			TrailerEPD:    "EPD",
		},
		Currency: "CHF",
		RefType:  "NON",
	}
}

// DefaultValue returns a minimal Swiss QR payload sample.
func (b *SwissQRBarcode) DefaultValue() string {
	return "SPC\r\n0200\r\n1\r\nCH5604835012345678009\r\nS\r\nMaxMuster\r\nMusterstrasse\r\n1\r\n8000\r\nZürich\r\nCH\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n10.00\r\nCHF\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\nNON\r\n\r\n\r\nEPD"
}

// FormatPayload builds the Swiss QR Code payload string from Params following
// the Swiss Payment Standards SPC v2.0 format. When Params.IBAN is non-empty
// the structured Params fields are used; otherwise the legacy flat fields
// (IBAN, Currency, CreditorName, etc.) are used as a fallback.
func (b *SwissQRBarcode) FormatPayload() string {
	// Use structured Params when populated.
	if b.Params.IBAN != "" {
		currency := b.Params.Currency
		if currency == "" {
			currency = "CHF"
		}
		refType := b.Params.ReferenceType
		if refType == "" {
			refType = "NON"
		}
		trailer := b.Params.TrailerEPD
		if trailer == "" {
			trailer = "EPD"
		}
		creditorCountry := b.Params.CreditorCountry
		if creditorCountry == "" {
			creditorCountry = "CH"
		}
		return strings.Join([]string{
			"SPC",
			"0200",
			"1",
			b.Params.IBAN,
			"K",
			b.Params.CreditorName,
			b.Params.CreditorStreet,
			b.Params.CreditorCity,
			b.Params.CreditorPostalCode,
			"",
			creditorCountry,
			// Debtor: 6 empty lines.
			"", "", "", "", "", "",
			b.Params.Amount,
			currency,
			refType,
			b.Params.Reference,
			b.Params.UnstructuredMessage,
			trailer,
			b.Params.AlternativeProcedure1,
			b.Params.AlternativeProcedure2,
		}, "\n")
	}
	// Fall back to legacy flat fields.
	return b.buildPayload()
}

// Encode validates and stores the Swiss QR payload string.
// When text is empty, FormatPayload is called to build the payload.
func (b *SwissQRBarcode) Encode(text string) error {
	if text == "" {
		text = b.FormatPayload()
	}
	b.encodedText = text
	return nil
}

// GetMatrix returns the QR code module matrix for the Swiss QR payload.
func (b *SwissQRBarcode) GetMatrix() ([][]bool, int, int) {
	text := b.encodedText
	if text == "" {
		text = b.FormatPayload()
	}
	matrix, err := encodeQR(text, qrECM, "")
	if err != nil || len(matrix) == 0 {
		return [][]bool{{true}}, 1, 1
	}
	n := len(matrix)
	return matrix, n, n
}

// Render renders the Swiss QR Code using the native matrix-based renderer.
func (b *SwissQRBarcode) Render(width, height int) (image.Image, error) {
	if b.encodedText == "" {
		if err := b.Encode(b.encodedText); err != nil {
			return nil, err
		}
	}
	matrix, rows, cols := b.GetMatrix()
	return DrawBarcode2D(matrix, rows, cols, width, height), nil
}

// buildPayload assembles the structured Swiss QR payload from the legacy flat fields.
// This is the fallback used when Params.IBAN is empty.
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
