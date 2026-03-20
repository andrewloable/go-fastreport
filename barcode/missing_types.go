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

// Encode validates and stores a 7- or 8-digit EAN-8 value.
func (e *EAN8Barcode) Encode(text string) error {
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("ean8: only digits allowed, found %q", ch)
		}
	}
	if len(text) < 7 {
		return fmt.Errorf("ean8: expected at least 7 digits, got %d", len(text))
	}
	e.encodedText = text
	return nil
}

// Render renders the EAN-8 barcode using the native pattern-based renderer.
func (e *EAN8Barcode) Render(width, height int) (image.Image, error) {
	if e.encodedText == "" {
		return nil, fmt.Errorf("ean8: Encode must be called before Render")
	}
	pattern, err := e.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, e.encodedText, width, height, true, e.GetWideBarRatio()), nil
}

// ── UPCABarcode ───────────────────────────────────────────────────────────────

// UPCABarcode implements the 11/12-digit UPC-A barcode symbology.
type UPCABarcode struct {
	BaseBarcodeImpl
}

// NewUPCABarcode creates a UPCABarcode.
func NewUPCABarcode() *UPCABarcode {
	return &UPCABarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeUPCA),
	}
}

// DefaultValue returns a sample 11-digit UPC-A value.
func (u *UPCABarcode) DefaultValue() string { return "01234567890" }

// Encode validates and stores the UPC-A value.
func (u *UPCABarcode) Encode(text string) error {
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("upca: only digits allowed, found %q", ch)
		}
	}
	u.encodedText = text
	return nil
}

// Render renders the UPC-A barcode using the native pattern-based renderer.
func (u *UPCABarcode) Render(width, height int) (image.Image, error) {
	if u.encodedText == "" {
		return nil, fmt.Errorf("upca: Encode must be called before Render")
	}
	pattern, err := u.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, u.encodedText, width, height, true, u.GetWideBarRatio()), nil
}

// ── UPCEBarcode ───────────────────────────────────────────────────────────────

// UPCEBarcode implements the UPC-E (zero-suppressed) barcode symbology.
// Rendered using the UPCE0 pattern encoder.
type UPCEBarcode struct{ BaseBarcodeImpl }

// NewUPCEBarcode creates a UPCEBarcode.
func NewUPCEBarcode() *UPCEBarcode {
	return &UPCEBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeUPCE)}
}

// DefaultValue returns a sample 7-digit UPC-E value.
func (u *UPCEBarcode) DefaultValue() string { return "1234567" }

// Encode validates and stores a UPC-E value (6-8 digits).
func (u *UPCEBarcode) Encode(text string) error {
	text = strings.TrimSpace(text)
	if len(text) < 6 || len(text) > 8 {
		return fmt.Errorf("upce: expected 6-8 digits, got %d", len(text))
	}
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("upce: only digits allowed, found %q", ch)
		}
	}
	u.encodedText = text
	return nil
}

// GetPattern generates the UPC-E bar pattern (delegates to UPCE0 pattern).
func (u *UPCEBarcode) GetPattern() (string, error) {
	helper := &UPCE0Barcode{}
	helper.encodedText = u.encodedText
	return helper.GetPattern()
}

// GetWideBarRatio returns the wide bar ratio for UPC-E.
func (u *UPCEBarcode) GetWideBarRatio() float32 { return 2 }

// Render renders the UPC-E barcode using the native pattern-based renderer.
func (u *UPCEBarcode) Render(width, height int) (image.Image, error) {
	if u.encodedText == "" {
		return nil, fmt.Errorf("upce: Encode must be called before Render")
	}
	pattern, err := u.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, u.encodedText, width, height, true, u.GetWideBarRatio()), nil
}

// ── Code93ExtendedBarcode ─────────────────────────────────────────────────────

// Code93ExtendedBarcode implements Code 93 Extended (full ASCII) symbology.
type Code93ExtendedBarcode struct {
	BaseBarcodeImpl
}

// NewCode93ExtendedBarcode creates a Code93ExtendedBarcode.
func NewCode93ExtendedBarcode() *Code93ExtendedBarcode {
	return &Code93ExtendedBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode93Extended),
	}
}

// DefaultValue returns a sample Code 93 Extended value.
func (c *Code93ExtendedBarcode) DefaultValue() string { return "CODE93" }

// Encode stores the text for Code 93 Extended rendering.
func (c *Code93ExtendedBarcode) Encode(text string) error {
	c.encodedText = text
	return nil
}

// Render renders the Code 93 Extended barcode using the native pattern-based renderer.
func (c *Code93ExtendedBarcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code93ext: Encode must be called before Render")
	}
	pattern, err := c.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, c.encodedText, width, height, true, c.GetWideBarRatio()), nil
}

// ── Code128ABarcode ───────────────────────────────────────────────────────────

// Code128ABarcode implements Code 128 set A (control chars + uppercase ASCII).
// Uses the native Code128 pattern encoder for rendering.
type Code128ABarcode struct{ BaseBarcodeImpl }

// NewCode128ABarcode creates a Code128ABarcode.
func NewCode128ABarcode() *Code128ABarcode {
	return &Code128ABarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode128A)}
}

// DefaultValue returns a sample value valid for Code 128A.
func (c *Code128ABarcode) DefaultValue() string { return "CODE128A" }

// Encode validates and stores the text for Code 128A encoding.
func (c *Code128ABarcode) Encode(text string) error {
	if text == "" {
		return fmt.Errorf("code128a: text must not be empty")
	}
	c.encodedText = text
	return nil
}

// Render renders the Code 128A barcode using the native pattern-based renderer.
func (c *Code128ABarcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code128a: Encode must be called before Render")
	}
	pattern, err := c.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, c.encodedText, width, height, true, c.GetWideBarRatio()), nil
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

// Encode validates and stores the text for Code 128B encoding.
func (c *Code128BBarcode) Encode(text string) error {
	if text == "" {
		return fmt.Errorf("code128b: text must not be empty")
	}
	c.encodedText = text
	return nil
}

// Render renders the Code 128B barcode using the native pattern-based renderer.
func (c *Code128BBarcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code128b: Encode must be called before Render")
	}
	pattern, err := c.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, c.encodedText, width, height, true, c.GetWideBarRatio()), nil
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

// Encode validates and stores the text for Code 128C encoding.
// Code 128C requires an even number of digits; a leading '0' is prepended if needed.
func (c *Code128CBarcode) Encode(text string) error {
	c.encodedText = text
	return nil
}

// Render renders the Code 128C barcode using the native pattern-based renderer.
func (c *Code128CBarcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code128c: Encode must be called before Render")
	}
	pattern, err := c.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, c.encodedText, width, height, true, c.GetWideBarRatio()), nil
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

// Render renders Standard (Industrial) 2-of-5 barcode using the native pattern-based renderer.
func (c *Code2of5IndustrialBarcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code2of5industrial: Encode must be called before Render")
	}
	pattern, err := c.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, c.encodedText, width, height, true, c.GetWideBarRatio()), nil
}

// ── Code2of5MatrixBarcode ─────────────────────────────────────────────────────

// Code2of5MatrixBarcode implements Matrix 2-of-5 symbology.
// Matrix 2-of-5 uses a dedicated bar-width encoding pattern distinct from
// Standard and Interleaved 2-of-5.
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

// Render renders Matrix 2-of-5 barcode using the native pattern-based renderer.
func (c *Code2of5MatrixBarcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code2of5matrix: Encode must be called before Render")
	}
	pattern, err := c.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, c.encodedText, width, height, true, c.GetWideBarRatio()), nil
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

// Render renders the ITF-14 barcode using the native pattern-based renderer.
func (i *ITF14Barcode) Render(width, height int) (image.Image, error) {
	if i.encodedText == "" {
		return nil, fmt.Errorf("itf14: Encode must be called before Render")
	}
	pattern, err := i.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, i.encodedText, width, height, true, i.GetWideBarRatio()), nil
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

// Render renders Deutsche Post Identcode using the native pattern-based renderer.
func (d *DeutscheIdentcodeBarcode) Render(width, height int) (image.Image, error) {
	if d.encodedText == "" {
		return nil, fmt.Errorf("identcode: Encode must be called before Render")
	}
	pattern, err := d.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, d.encodedText, width, height, true, d.GetWideBarRatio()), nil
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

// Render renders Deutsche Post Leitcode using the native pattern-based renderer.
func (d *DeutscheLeitcodeBarcode) Render(width, height int) (image.Image, error) {
	if d.encodedText == "" {
		return nil, fmt.Errorf("leitcode: Encode must be called before Render")
	}
	pattern, err := d.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, d.encodedText, width, height, true, d.GetWideBarRatio()), nil
}

// ── Supplement2Barcode ────────────────────────────────────────────────────────

// Supplement2Barcode implements the 2-digit EAN/UPC add-on supplement barcode.
// It encodes a 2-digit numeric value as a supplementary barcode appended to
// EAN-13 or UPC-A barcodes. Rendered using the native EAN supplement pattern encoder.
type Supplement2Barcode struct{ BaseBarcodeImpl }

// NewSupplement2Barcode creates a Supplement2Barcode.
func NewSupplement2Barcode() *Supplement2Barcode {
	return &Supplement2Barcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeSupplement2)}
}

// DefaultValue returns a sample 2-digit supplement value.
func (s *Supplement2Barcode) DefaultValue() string { return "53" }

// Encode validates and stores the 2-digit supplement value.
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
	s.encodedText = text
	return nil
}

// Render renders the Supplement 2 barcode using the native pattern-based renderer.
func (s *Supplement2Barcode) Render(width, height int) (image.Image, error) {
	if s.encodedText == "" {
		return nil, fmt.Errorf("supplement2: Encode must be called before Render")
	}
	pattern, err := s.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, s.encodedText, width, height, false, s.GetWideBarRatio()), nil
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

// Encode validates and stores the 5-digit supplement value.
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
	s.encodedText = text
	return nil
}

// Render renders the Supplement 5 barcode using the native pattern-based renderer.
func (s *Supplement5Barcode) Render(width, height int) (image.Image, error) {
	if s.encodedText == "" {
		return nil, fmt.Errorf("supplement5: Encode must be called before Render")
	}
	pattern, err := s.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, s.encodedText, width, height, false, s.GetWideBarRatio()), nil
}

// ── JapanPost4StateBarcode ────────────────────────────────────────────────────

// JapanPost4StateBarcode implements the Japan Post Customer Barcode (4-state).
// The barcode encodes a Japanese postal address using 4 bar heights.
// This implementation renders the value as a Code 128 barcode pattern as a
// visual approximation. The encoding accepts the alphanumeric postal code with hyphens.
type JapanPost4StateBarcode struct{ BaseBarcodeImpl }

// ── Types referenced by dedicated implementation files ────────────────────────

// Code39ExtendedBarcode implements Code 39 in full-ASCII mode.
type Code39ExtendedBarcode struct{ BaseBarcodeImpl }

// NewCode39ExtendedBarcode creates a Code39ExtendedBarcode.
func NewCode39ExtendedBarcode() *Code39ExtendedBarcode {
	return &Code39ExtendedBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode39Extended)}
}

// DefaultValue returns a sample Code 39 Extended value.
func (c *Code39ExtendedBarcode) DefaultValue() string { return "abc-1234" }

// Encode stores text for Code 39 Extended rendering.
func (c *Code39ExtendedBarcode) Encode(text string) error {
	c.encodedText = text
	return nil
}

// Render renders the Code 39 Extended barcode using the native pattern-based renderer.
func (c *Code39ExtendedBarcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code39ext: Encode must be called before Render")
	}
	pattern, err := c.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, c.encodedText, width, height, true, c.GetWideBarRatio()), nil
}

// UPCE0Barcode implements UPC-E number system 0.
type UPCE0Barcode struct{ BaseBarcodeImpl }

// NewUPCE0Barcode creates a UPCE0Barcode.
func NewUPCE0Barcode() *UPCE0Barcode {
	return &UPCE0Barcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeUPCE)}
}

// DefaultValue returns a sample UPC-E0 value.
func (u *UPCE0Barcode) DefaultValue() string { return "01234565" }

// Encode stores text for UPC-E0 rendering.
func (u *UPCE0Barcode) Encode(text string) error {
	u.encodedText = text
	return nil
}

// Render renders the UPC-E0 barcode using the native pattern-based renderer.
func (u *UPCE0Barcode) Render(width, height int) (image.Image, error) {
	if u.encodedText == "" {
		return nil, fmt.Errorf("upce0: Encode must be called before Render")
	}
	pattern, err := u.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, u.encodedText, width, height, true, u.GetWideBarRatio()), nil
}

// UPCE1Barcode implements UPC-E number system 1.
type UPCE1Barcode struct{ BaseBarcodeImpl }

// NewUPCE1Barcode creates a UPCE1Barcode.
func NewUPCE1Barcode() *UPCE1Barcode {
	return &UPCE1Barcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeUPCE)}
}

// DefaultValue returns a sample UPC-E1 value.
func (u *UPCE1Barcode) DefaultValue() string { return "11234565" }

// Encode stores text for UPC-E1 rendering.
func (u *UPCE1Barcode) Encode(text string) error {
	u.encodedText = text
	return nil
}

// Render renders the UPC-E1 barcode using the native pattern-based renderer.
func (u *UPCE1Barcode) Render(width, height int) (image.Image, error) {
	if u.encodedText == "" {
		return nil, fmt.Errorf("upce1: Encode must be called before Render")
	}
	pattern, err := u.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, u.encodedText, width, height, true, u.GetWideBarRatio()), nil
}

// GS1_128Barcode implements GS1-128 (formerly EAN-128).
type GS1_128Barcode struct{ BaseBarcodeImpl }

// NewGS1_128Barcode creates a GS1_128Barcode.
func NewGS1_128Barcode() *GS1_128Barcode {
	return &GS1_128Barcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeGS1_128)}
}

// DefaultValue returns a sample GS1-128 value.
func (g *GS1_128Barcode) DefaultValue() string { return "(01)12345678901231" }

// Encode stores text for GS1-128 rendering via GetPattern.
func (g *GS1_128Barcode) Encode(text string) error {
	g.encodedText = text
	return nil
}

// Render renders the GS1-128 barcode using the native pattern-based renderer.
func (g *GS1_128Barcode) Render(width, height int) (image.Image, error) {
	if g.encodedText == "" {
		return nil, fmt.Errorf("gs1_128: Encode must be called before Render")
	}
	pattern, err := g.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, g.encodedText, width, height, true, g.GetWideBarRatio()), nil
}

// GS1DatamatrixBarcode implements GS1 DataMatrix.
type GS1DatamatrixBarcode struct{ BaseBarcodeImpl }

// NewGS1DatamatrixBarcode creates a GS1DatamatrixBarcode.
func NewGS1DatamatrixBarcode() *GS1DatamatrixBarcode {
	return &GS1DatamatrixBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeDataMatrix)}
}

// DefaultValue returns a sample GS1 DataMatrix value.
func (g *GS1DatamatrixBarcode) DefaultValue() string { return "(01)12345678901231" }

// Encode stores text for GS1 DataMatrix rendering via GetMatrix.
func (g *GS1DatamatrixBarcode) Encode(text string) error {
	g.encodedText = text
	return nil
}

// Render renders the GS1 DataMatrix barcode using the native matrix-based renderer.
func (g *GS1DatamatrixBarcode) Render(width, height int) (image.Image, error) {
	if g.encodedText == "" {
		return nil, fmt.Errorf("gs1datamatrix: Encode must be called before Render")
	}
	matrix, rows, cols := g.GetMatrix()
	return DrawBarcode2D(matrix, rows, cols, width, height), nil
}

// NewJapanPost4StateBarcode creates a JapanPost4StateBarcode.
func NewJapanPost4StateBarcode() *JapanPost4StateBarcode {
	return &JapanPost4StateBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeJapanPost4State)}
}

// DefaultValue returns a sample Japan Post 4-state value.
func (j *JapanPost4StateBarcode) DefaultValue() string { return "597-8615-5-7-6" }

// Encode validates and stores the postal code for rendering.
func (j *JapanPost4StateBarcode) Encode(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("japanpost4state: text must not be empty")
	}
	j.encodedText = text
	return nil
}

// Render renders the Japan Post 4-state barcode using Code128 pattern as a visual approximation.
func (j *JapanPost4StateBarcode) Render(width, height int) (image.Image, error) {
	if j.encodedText == "" {
		return nil, fmt.Errorf("japanpost4state: Encode must be called before Render")
	}
	// Strip hyphens for encoding.
	cleaned := strings.ReplaceAll(j.encodedText, "-", "")
	helper := &Code128Barcode{}
	helper.encodedText = cleaned
	pattern, err := helper.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, j.encodedText, width, height, false, helper.GetWideBarRatio()), nil
}
