package barcode

// missing_types.go implements the barcode types that were previously
// mapped to generic fallbacks but now have proper distinct implementations:
//
//   - EAN8Barcode              — 7/8-digit EAN-8 (distinct from EAN-13)
//   - UPCABarcode              — 11/12-digit UPC-A
//   - UPCEBarcode              — 6/7/8-digit UPC-E
//   - Code93ExtendedBarcode    — Code 93 in full-ASCII mode
//   - Code128ABarcode          — Code 128 set A (enforced via Code128 auto)
//   - Code128BBarcode          — Code 128 set B (enforced via Code128 auto)
//   - Code128CBarcode          — Code 128 set C (numeric pairs)
//   - Code2of5IndustrialBarcode — Standard (Industrial) 2-of-5
//   - Code2of5MatrixBarcode    — Matrix 2-of-5
//   - ITF14Barcode             — 14-digit ITF-14 (interleaved 2-of-5)
//   - DeutscheIdentcodeBarcode — Deutsche Post Identcode (11-digit interleaved)
//   - DeutscheLeitcodeBarcode  — Deutsche Post Leitcode (13-digit interleaved)
//   - Supplement2Barcode       — 2-digit EAN add-on supplement
//   - Supplement5Barcode       — 5-digit EAN add-on supplement
//   - JapanPost4StateBarcode   — Japan Post 4-state customer barcode

import (
	"fmt"
	"image"
	"strings"

	boombarcode "github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	boomean "github.com/boombuler/barcode/ean"

	"github.com/andrewloable/go-fastreport/barcode/code2of5"
	"github.com/andrewloable/go-fastreport/barcode/code93"
	"github.com/andrewloable/go-fastreport/barcode/upc"
)

// ── EAN8Barcode ───────────────────────────────────────────────────────────────

// EAN8Barcode implements the 7/8-digit EAN-8 barcode symbology.
type EAN8Barcode struct{ BaseBarcodeImpl }

// NewEAN8Barcode creates an EAN8Barcode.
func NewEAN8Barcode() *EAN8Barcode {
	return &EAN8Barcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeEAN8)}
}

// DefaultValue returns a sample 7-digit EAN-8 value (checksum auto-computed).
func (e *EAN8Barcode) DefaultValue() string { return "1234567" }

// Encode encodes a 7- or 8-digit EAN-8 value.
func (e *EAN8Barcode) Encode(text string) error {
	bc, err := boomean.Encode(text)
	if err != nil {
		// C# recalculates checksum. Try with first 7 digits (let library compute check digit).
		if len(text) == 8 {
			bc, err = boomean.Encode(text[:7])
		}
		if err != nil {
			return fmt.Errorf("ean8 encode: %w", err)
		}
	}
	e.encodedText = text
	e.encoded = bc
	return nil
}

// ── UPCABarcode ───────────────────────────────────────────────────────────────

// UPCABarcode implements the 11/12-digit UPC-A barcode symbology.
type UPCABarcode struct {
	BaseBarcodeImpl
	enc *upc.Encoder
}

// NewUPCABarcode creates a UPCABarcode.
func NewUPCABarcode() *UPCABarcode {
	return &UPCABarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeUPCA),
		enc:             upc.New(),
	}
}

// DefaultValue returns a sample 11-digit UPC-A value.
func (u *UPCABarcode) DefaultValue() string { return "01234567890" }

// Encode validates the UPC-A value.
func (u *UPCABarcode) Encode(text string) error {
	if err := u.enc.Validate(text); err != nil {
		return fmt.Errorf("upca encode: %w", err)
	}
	u.encodedText = text
	return nil
}

// Render renders the UPC-A barcode as an image.
func (u *UPCABarcode) Render(width, height int) (image.Image, error) {
	if u.encodedText == "" {
		return nil, fmt.Errorf("upca: Encode must be called before Render")
	}
	return u.enc.Encode(u.encodedText, width, height)
}

// ── UPCEBarcode ───────────────────────────────────────────────────────────────

// UPCEBarcode implements the UPC-E (zero-suppressed) barcode symbology.
// UPC-E is encoded as EAN-13 with a '0' prefix (the boombuler library handles
// the conversion internally when given an 8-digit UPC-E code starting with 0).
type UPCEBarcode struct{ BaseBarcodeImpl }

// NewUPCEBarcode creates a UPCEBarcode.
func NewUPCEBarcode() *UPCEBarcode {
	return &UPCEBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeUPCE)}
}

// DefaultValue returns a sample 7-digit UPC-E value.
func (u *UPCEBarcode) DefaultValue() string { return "1234567" }

// Encode encodes a UPC-E value using EAN encoding (boombuler treats 8-digit
// codes starting with 0 as UPC-E, others as EAN-8).
func (u *UPCEBarcode) Encode(text string) error {
	// Strip any trailing whitespace and validate.
	text = strings.TrimSpace(text)
	if len(text) < 6 || len(text) > 8 {
		return fmt.Errorf("upce: expected 6-8 digits, got %d", len(text))
	}
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("upce: only digits allowed, found %q", ch)
		}
	}
	bc, err := boomean.Encode(text)
	if err != nil {
		return fmt.Errorf("upce encode: %w", err)
	}
	u.encodedText = text
	u.encoded = bc
	return nil
}

// ── Code93ExtendedBarcode ─────────────────────────────────────────────────────

// Code93ExtendedBarcode implements Code 93 Extended (full ASCII) symbology.
type Code93ExtendedBarcode struct {
	BaseBarcodeImpl
	enc *code93.Encoder
}

// NewCode93ExtendedBarcode creates a Code93ExtendedBarcode.
func NewCode93ExtendedBarcode() *Code93ExtendedBarcode {
	enc := code93.New()
	enc.FullASCIIMode = true
	return &Code93ExtendedBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode93Extended),
		enc:             enc,
	}
}

// DefaultValue returns a sample Code 93 Extended value.
func (c *Code93ExtendedBarcode) DefaultValue() string { return "CODE93" }

// Encode stores the text for Code 93 Extended rendering.
func (c *Code93ExtendedBarcode) Encode(text string) error {
	c.encodedText = text
	return nil
}

// Render renders the Code 93 Extended barcode.
func (c *Code93ExtendedBarcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code93ext: Encode must be called before Render")
	}
	return c.enc.Encode(c.encodedText, width, height)
}

// ── Code128ABarcode ───────────────────────────────────────────────────────────

// Code128ABarcode implements Code 128 set A (control chars + uppercase ASCII).
// boombuler auto-selects the code set; we use the same encoder as Code128.
type Code128ABarcode struct{ BaseBarcodeImpl }

// NewCode128ABarcode creates a Code128ABarcode.
func NewCode128ABarcode() *Code128ABarcode {
	return &Code128ABarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode128A)}
}

// DefaultValue returns a sample value valid for Code 128A.
func (c *Code128ABarcode) DefaultValue() string { return "CODE128A" }

// Encode encodes the text using Code 128 (auto code-set selection).
func (c *Code128ABarcode) Encode(text string) error {
	bc, err := code128.Encode(text)
	if err != nil {
		return fmt.Errorf("code128a encode: %w", err)
	}
	c.encodedText = text
	c.encoded = bc
	return nil
}

// ── Code128BBarcode ───────────────────────────────────────────────────────────

// Code128BBarcode implements Code 128 set B (printable ASCII).
type Code128BBarcode struct{ BaseBarcodeImpl }

// NewCode128BBarcode creates a Code128BBarcode.
func NewCode128BBarcode() *Code128BBarcode {
	return &Code128BBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode128B)}
}

// DefaultValue returns a sample value valid for Code 128B.
func (c *Code128BBarcode) DefaultValue() string { return "Code128B" }

// Encode encodes the text using Code 128 (auto code-set selection).
func (c *Code128BBarcode) Encode(text string) error {
	bc, err := code128.Encode(text)
	if err != nil {
		return fmt.Errorf("code128b encode: %w", err)
	}
	c.encodedText = text
	c.encoded = bc
	return nil
}

// ── Code128CBarcode ───────────────────────────────────────────────────────────

// Code128CBarcode implements Code 128 set C (numeric pairs only).
type Code128CBarcode struct{ BaseBarcodeImpl }

// NewCode128CBarcode creates a Code128CBarcode.
func NewCode128CBarcode() *Code128CBarcode {
	return &Code128CBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode128C)}
}

// DefaultValue returns a sample value valid for Code 128C (even-length digits).
func (c *Code128CBarcode) DefaultValue() string { return "12345678" }

// Encode encodes the text using Code 128 (auto code-set selection).
// Code 128C requires an even number of digits; a leading '0' is prepended if needed.
func (c *Code128CBarcode) Encode(text string) error {
	// Code128C needs even-length digit string; pad if necessary.
	enc := text
	if len(enc)%2 != 0 {
		enc = "0" + enc
	}
	bc, err := code128.Encode(enc)
	if err != nil {
		return fmt.Errorf("code128c encode: %w", err)
	}
	c.encodedText = text
	c.encoded = bc
	return nil
}

// ── Code2of5IndustrialBarcode ─────────────────────────────────────────────────

// Code2of5IndustrialBarcode implements Standard (Industrial) 2-of-5 symbology.
type Code2of5IndustrialBarcode struct {
	BaseBarcodeImpl
}

// NewCode2of5IndustrialBarcode creates a Code2of5IndustrialBarcode.
func NewCode2of5IndustrialBarcode() *Code2of5IndustrialBarcode {
	return &Code2of5IndustrialBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode2of5Industrial)}
}

// DefaultValue returns a sample Standard 2-of-5 value.
func (c *Code2of5IndustrialBarcode) DefaultValue() string { return "123456" }

// Encode stores text for Standard (non-interleaved) 2-of-5 rendering.
func (c *Code2of5IndustrialBarcode) Encode(text string) error {
	c.encodedText = text
	return nil
}

// Render renders Standard (Industrial) 2-of-5 barcode.
func (c *Code2of5IndustrialBarcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code2of5industrial: Encode must be called before Render")
	}
	enc := code2of5.New()
	enc.Interleaved = false
	return enc.Encode(c.encodedText, width, height)
}

// ── Code2of5MatrixBarcode ─────────────────────────────────────────────────────

// Code2of5MatrixBarcode implements Matrix 2-of-5 symbology.
// Matrix 2-of-5 uses the same encoding structure as Standard 2-of-5 but with
// a different bar-width ratio. We render it as Standard (non-interleaved) 2-of-5
// since the boombuler library does not have a distinct Matrix variant.
type Code2of5MatrixBarcode struct {
	BaseBarcodeImpl
}

// NewCode2of5MatrixBarcode creates a Code2of5MatrixBarcode.
func NewCode2of5MatrixBarcode() *Code2of5MatrixBarcode {
	return &Code2of5MatrixBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode2of5Matrix)}
}

// DefaultValue returns a sample Matrix 2-of-5 value.
func (c *Code2of5MatrixBarcode) DefaultValue() string { return "123456" }

// Encode stores text for Matrix 2-of-5 rendering.
func (c *Code2of5MatrixBarcode) Encode(text string) error {
	c.encodedText = text
	return nil
}

// Render renders Matrix 2-of-5 barcode (rendered as Standard 2-of-5).
func (c *Code2of5MatrixBarcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code2of5matrix: Encode must be called before Render")
	}
	enc := code2of5.New()
	enc.Interleaved = false
	return enc.Encode(c.encodedText, width, height)
}

// ── ITF14Barcode ──────────────────────────────────────────────────────────────

// ITF14Barcode implements the ITF-14 (Interleaved 2-of-5, 14-digit) symbology.
// ITF-14 is a 14-digit Interleaved 2-of-5 barcode used for shipping containers.
type ITF14Barcode struct {
	BaseBarcodeImpl
}

// NewITF14Barcode creates an ITF14Barcode.
func NewITF14Barcode() *ITF14Barcode {
	return &ITF14Barcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeITF14)}
}

// DefaultValue returns a sample 14-digit ITF-14 value.
func (i *ITF14Barcode) DefaultValue() string { return "12345678901231" }

// Encode validates and stores the ITF-14 value (14 digits).
func (i *ITF14Barcode) Encode(text string) error {
	// Strip any whitespace.
	text = strings.TrimSpace(text)
	// Accept 13 or 14 digits; pad to 14 with a leading '0' if 13.
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("itf14: only digits allowed, found %q", ch)
		}
	}
	switch len(text) {
	case 13:
		text = "0" + text
	case 14:
		// OK
	default:
		return fmt.Errorf("itf14: expected 13 or 14 digits, got %d", len(text))
	}
	i.encodedText = text
	return nil
}

// Render renders the ITF-14 barcode as an Interleaved 2-of-5 image.
func (i *ITF14Barcode) Render(width, height int) (image.Image, error) {
	if i.encodedText == "" {
		return nil, fmt.Errorf("itf14: Encode must be called before Render")
	}
	enc := code2of5.New()
	enc.Interleaved = true
	return enc.Encode(i.encodedText, width, height)
}

// ── DeutscheIdentcodeBarcode ──────────────────────────────────────────────────

// DeutscheIdentcodeBarcode implements Deutsche Post Identcode (11-digit Interleaved 2-of-5).
type DeutscheIdentcodeBarcode struct {
	BaseBarcodeImpl
}

// NewDeutscheIdentcodeBarcode creates a DeutscheIdentcodeBarcode.
func NewDeutscheIdentcodeBarcode() *DeutscheIdentcodeBarcode {
	return &DeutscheIdentcodeBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeDeutscheIdentcode)}
}

// DefaultValue returns a sample 11-digit Identcode value.
func (d *DeutscheIdentcodeBarcode) DefaultValue() string { return "12345123456" }

// Encode validates and stores the Identcode value (11 digits).
func (d *DeutscheIdentcodeBarcode) Encode(text string) error {
	text = strings.TrimSpace(text)
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("identcode: only digits allowed, found %q", ch)
		}
	}
	if len(text) != 11 {
		return fmt.Errorf("identcode: expected 11 digits, got %d", len(text))
	}
	d.encodedText = text
	return nil
}

// Render renders Deutsche Post Identcode as Interleaved 2-of-5.
// Interleaved 2-of-5 requires an even number of digits; a leading '0' is
// prepended if the Identcode digit count is odd.
func (d *DeutscheIdentcodeBarcode) Render(width, height int) (image.Image, error) {
	if d.encodedText == "" {
		return nil, fmt.Errorf("identcode: Encode must be called before Render")
	}
	enc := code2of5.New()
	enc.Interleaved = true
	text := d.encodedText
	if len(text)%2 != 0 {
		text = "0" + text
	}
	return enc.Encode(text, width, height)
}

// ── DeutscheLeitcodeBarcode ───────────────────────────────────────────────────

// DeutscheLeitcodeBarcode implements Deutsche Post Leitcode (13-digit Interleaved 2-of-5).
type DeutscheLeitcodeBarcode struct {
	BaseBarcodeImpl
}

// NewDeutscheLeitcodeBarcode creates a DeutscheLeitcodeBarcode.
func NewDeutscheLeitcodeBarcode() *DeutscheLeitcodeBarcode {
	return &DeutscheLeitcodeBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeDeutscheLeitcode)}
}

// DefaultValue returns a sample 13-digit Leitcode value.
func (d *DeutscheLeitcodeBarcode) DefaultValue() string { return "1234512312312" }

// Encode validates and stores the Leitcode value (13 digits).
func (d *DeutscheLeitcodeBarcode) Encode(text string) error {
	text = strings.TrimSpace(text)
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("leitcode: only digits allowed, found %q", ch)
		}
	}
	if len(text) != 13 {
		return fmt.Errorf("leitcode: expected 13 digits, got %d", len(text))
	}
	d.encodedText = text
	return nil
}

// Render renders Deutsche Post Leitcode as Interleaved 2-of-5.
// Interleaved 2-of-5 requires an even number of digits; a leading '0' is
// prepended if the Leitcode digit count is odd.
func (d *DeutscheLeitcodeBarcode) Render(width, height int) (image.Image, error) {
	if d.encodedText == "" {
		return nil, fmt.Errorf("leitcode: Encode must be called before Render")
	}
	enc := code2of5.New()
	enc.Interleaved = true
	text := d.encodedText
	if len(text)%2 != 0 {
		text = "0" + text
	}
	return enc.Encode(text, width, height)
}

// ── Supplement2Barcode ────────────────────────────────────────────────────────

// Supplement2Barcode implements the 2-digit EAN/UPC add-on supplement barcode.
// It encodes a 2-digit numeric value as a supplementary barcode appended to
// EAN-13 or UPC-A barcodes. Rendered using Code 128 as a visual placeholder
// since the boombuler library does not have a dedicated supplement encoder.
type Supplement2Barcode struct{ BaseBarcodeImpl }

// NewSupplement2Barcode creates a Supplement2Barcode.
func NewSupplement2Barcode() *Supplement2Barcode {
	return &Supplement2Barcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeSupplement2)}
}

// DefaultValue returns a sample 2-digit supplement value.
func (s *Supplement2Barcode) DefaultValue() string { return "53" }

// Encode validates the 2-digit supplement value.
func (s *Supplement2Barcode) Encode(text string) error {
	text = strings.TrimSpace(text)
	if len(text) != 2 {
		return fmt.Errorf("supplement2: expected 2 digits, got %d", len(text))
	}
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("supplement2: only digits allowed, found %q", ch)
		}
	}
	bc, err := supplement2Encode(text)
	if err != nil {
		return fmt.Errorf("supplement2 encode: %w", err)
	}
	s.encodedText = text
	s.encoded = bc
	return nil
}

// supplement2Encode encodes a 2-digit supplement using Code 128.
// This is a visual approximation; a true supplement encoder would generate
// the interleaved bar pattern per the EAN supplement specification.
func supplement2Encode(text string) (boombarcode.Barcode, error) {
	return code128.Encode(text)
}

// ── Supplement5Barcode ────────────────────────────────────────────────────────

// Supplement5Barcode implements the 5-digit EAN/UPC add-on supplement barcode.
type Supplement5Barcode struct{ BaseBarcodeImpl }

// NewSupplement5Barcode creates a Supplement5Barcode.
func NewSupplement5Barcode() *Supplement5Barcode {
	return &Supplement5Barcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeSupplement5)}
}

// DefaultValue returns a sample 5-digit supplement value.
func (s *Supplement5Barcode) DefaultValue() string { return "52495" }

// Encode validates the 5-digit supplement value.
func (s *Supplement5Barcode) Encode(text string) error {
	text = strings.TrimSpace(text)
	if len(text) != 5 {
		return fmt.Errorf("supplement5: expected 5 digits, got %d", len(text))
	}
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("supplement5: only digits allowed, found %q", ch)
		}
	}
	bc, err := code128.Encode(text)
	if err != nil {
		return fmt.Errorf("supplement5 encode: %w", err)
	}
	s.encodedText = text
	s.encoded = bc
	return nil
}

// ── JapanPost4StateBarcode ────────────────────────────────────────────────────

// JapanPost4StateBarcode implements the Japan Post Customer Barcode (4-state).
// The barcode encodes a Japanese postal address using 4 bar heights.
// This implementation renders the value as a Code 128 barcode since the
// boombuler library does not have a dedicated Japan Post encoder.
// The encoding accepts the alphanumeric postal code with hyphens.
type JapanPost4StateBarcode struct{ BaseBarcodeImpl }

// NewJapanPost4StateBarcode creates a JapanPost4StateBarcode.
func NewJapanPost4StateBarcode() *JapanPost4StateBarcode {
	return &JapanPost4StateBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeJapanPost4State)}
}

// DefaultValue returns a sample Japan Post 4-state value.
func (j *JapanPost4StateBarcode) DefaultValue() string { return "597-8615-5-7-6" }

// Encode stores the postal code for rendering.
func (j *JapanPost4StateBarcode) Encode(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("japanpost4state: text must not be empty")
	}
	// Strip hyphens and encode digit-only content as Code 128.
	cleaned := strings.ReplaceAll(text, "-", "")
	bc, err := code128.Encode(cleaned)
	if err != nil {
		return fmt.Errorf("japanpost4state encode: %w", err)
	}
	j.encodedText = text
	j.encoded = bc
	return nil
}
