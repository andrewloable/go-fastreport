// Package barcode implements barcode rendering support for go-fastreport.
// It provides the BarcodeBase interface and the BarcodeObject report element.
package barcode

import (
	"fmt"
	"image"
	"image/color"

	boombarcode "github.com/boombuler/barcode"
	"github.com/boombuler/barcode/aztec"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/code39"
	"github.com/boombuler/barcode/ean"
	"github.com/boombuler/barcode/pdf417"
	"github.com/boombuler/barcode/qr"

	"github.com/andrewloable/go-fastreport/barcode/codabar"
	"github.com/andrewloable/go-fastreport/barcode/code2of5"
	"github.com/andrewloable/go-fastreport/barcode/code93"
	"github.com/andrewloable/go-fastreport/barcode/datamatrix"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
)

// -----------------------------------------------------------------------
// BarcodeType registry
// -----------------------------------------------------------------------

// BarcodeType identifies a barcode symbology.
type BarcodeType string

const (
	BarcodeTypeCode128    BarcodeType = "Code128"
	BarcodeTypeCode39     BarcodeType = "Code39"
	BarcodeTypeCode93     BarcodeType = "Code93"
	BarcodeTypeCode2of5   BarcodeType = "2of5"
	BarcodeTypeCodabar    BarcodeType = "Codabar"
	BarcodeTypeEAN13      BarcodeType = "EAN13"
	BarcodeTypeEAN8       BarcodeType = "EAN8"
	BarcodeTypeUPCA       BarcodeType = "UPCA"
	BarcodeTypeUPCE       BarcodeType = "UPCE"
	BarcodeTypeMSI        BarcodeType = "MSI"
	BarcodeTypeQR         BarcodeType = "QR"
	BarcodeTypeDataMatrix BarcodeType = "DataMatrix"
	BarcodeTypeAztec      BarcodeType = "Aztec"
	BarcodeTypeMaxiCode   BarcodeType = "MaxiCode"
	BarcodeTypePDF417     BarcodeType = "PDF417"
	BarcodeTypeGS1_128    BarcodeType = "GS1-128"
)

// -----------------------------------------------------------------------
// BarcodeBase interface
// -----------------------------------------------------------------------

// BarcodeBase is the interface implemented by all barcode symbologies.
// It is the Go equivalent of FastReport.Barcode.BarcodeBase.
type BarcodeBase interface {
	// Type returns the barcode symbology identifier.
	Type() BarcodeType
	// Encode validates and encodes text into the barcode symbol.
	// Returns an error when text is invalid for the symbology.
	Encode(text string) error
	// DefaultValue returns a valid sample value for this symbology.
	DefaultValue() string
}

// -----------------------------------------------------------------------
// BaseBarcodeImpl — shared state for concrete barcodes
// -----------------------------------------------------------------------

// BaseBarcodeImpl holds the properties common to all barcode types.
type BaseBarcodeImpl struct {
	// Color is the bar colour (default black).
	Color color.RGBA
	// Font is used to print the human-readable text below the barcode.
	Font style.Font
	// encodedText is set by Encode.
	encodedText string
	// barcodeType is the symbology.
	barcodeType BarcodeType
	// encoded holds the boombuler Barcode result after Encode.
	encoded boombarcode.Barcode
}

// Render scales the encoded barcode to the given pixel dimensions and returns
// an image.Image. Returns an error if Encode has not been called yet.
func (b *BaseBarcodeImpl) Render(width, height int) (image.Image, error) {
	if b.encoded == nil {
		return nil, fmt.Errorf("barcode not encoded — call Encode first")
	}
	img, err := boombarcode.Scale(b.encoded, width, height)
	if err != nil {
		return nil, fmt.Errorf("scale barcode: %w", err)
	}
	return img, nil
}

// newBaseBarcodeImpl creates a BaseBarcodeImpl for the given type.
func newBaseBarcodeImpl(t BarcodeType) BaseBarcodeImpl {
	return BaseBarcodeImpl{
		Color:       color.RGBA{A: 255}, // black
		Font:        style.Font{Name: "Arial", Size: 8},
		barcodeType: t,
	}
}

// Type implements BarcodeBase.
func (b *BaseBarcodeImpl) Type() BarcodeType { return b.barcodeType }

// DefaultValue returns a generic sample string.
func (b *BaseBarcodeImpl) DefaultValue() string { return "12345" }

// EncodedText returns the validated text stored by the last Encode call.
func (b *BaseBarcodeImpl) EncodedText() string { return b.encodedText }

// -----------------------------------------------------------------------
// Concrete simple barcodes (linear 1-D)
// -----------------------------------------------------------------------

// Code128Barcode implements Code128 symbology.
type Code128Barcode struct{ BaseBarcodeImpl }

// NewCode128Barcode creates a Code128Barcode.
func NewCode128Barcode() *Code128Barcode {
	return &Code128Barcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode128)}
}

// Encode encodes text using Code128 symbology.
func (c *Code128Barcode) Encode(text string) error {
	bc, err := code128.Encode(text)
	if err != nil {
		return fmt.Errorf("code128 encode: %w", err)
	}
	c.encodedText = text
	c.encoded = bc
	return nil
}

// DefaultValue returns a sample Code128 value.
func (c *Code128Barcode) DefaultValue() string { return "CODE128" }

// Code39Barcode implements Code39 (3-of-9) symbology.
type Code39Barcode struct {
	BaseBarcodeImpl
	// AllowExtended enables Code39 Extended (full ASCII).
	AllowExtended bool
	// CalcChecksum adds a modulo-43 checksum character.
	CalcChecksum bool
}

// NewCode39Barcode creates a Code39Barcode.
func NewCode39Barcode() *Code39Barcode {
	return &Code39Barcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode39)}
}

// Encode encodes text using Code39 symbology.
func (c *Code39Barcode) Encode(text string) error {
	bc, err := code39.Encode(text, c.CalcChecksum, c.AllowExtended)
	if err != nil {
		return fmt.Errorf("code39 encode: %w", err)
	}
	c.encodedText = text
	c.encoded = bc
	return nil
}

// DefaultValue returns a sample Code39 value.
func (c *Code39Barcode) DefaultValue() string { return "CODE39" }

// QRBarcode implements QR Code 2D symbology.
type QRBarcode struct {
	BaseBarcodeImpl
	// ErrorCorrection is the QR error correction level (L/M/Q/H).
	ErrorCorrection string // default "M"
}

// NewQRBarcode creates a QRBarcode.
func NewQRBarcode() *QRBarcode {
	return &QRBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeQR),
		ErrorCorrection: "M",
	}
}

// Encode encodes text as a QR code using the configured error correction level.
func (q *QRBarcode) Encode(text string) error {
	level := qr.M
	switch q.ErrorCorrection {
	case "L":
		level = qr.L
	case "Q":
		level = qr.Q
	case "H":
		level = qr.H
	}
	bc, err := qr.Encode(text, level, qr.Auto)
	if err != nil {
		return fmt.Errorf("qr encode: %w", err)
	}
	q.encodedText = text
	q.encoded = bc
	return nil
}

// DefaultValue returns a sample QR value.
func (q *QRBarcode) DefaultValue() string { return "https://example.com" }

// -----------------------------------------------------------------------
// BarcodeObject
// -----------------------------------------------------------------------

// BarcodeObject is a report element that renders a barcode.
// It is the Go equivalent of FastReport.BarcodeObject.
type BarcodeObject struct {
	report.ReportComponentBase

	// Barcode is the symbology implementation.
	Barcode BarcodeBase
	// angle is the rotation in degrees.
	angle int
	// autoSize adjusts the object bounds to fit the barcode.
	autoSize bool
	// dataColumn is the data source column that provides the barcode text.
	dataColumn string
	// expression is a report expression used as the barcode text.
	expression string
	// text is the static barcode text (used when expression/dataColumn are empty).
	text string
	// showText shows the human-readable text below the bars.
	showText bool
	// padding is the interior padding around the barcode.
	padding Padding
	// zoom scales the barcode (default 1.0).
	zoom float32
	// hideIfNoData hides the object when text evaluates to empty.
	hideIfNoData bool
	// noDataText is shown when hideIfNoData is false and text is empty.
	noDataText string
	// allowExpressions enables bracket-expression evaluation.
	allowExpressions bool
	// brackets is the expression delimiter (default "[,]").
	brackets string
}

// Padding holds interior spacing around the barcode.
type Padding struct {
	Left, Top, Right, Bottom float32
}

// NewBarcodeObject creates a BarcodeObject with defaults.
func NewBarcodeObject() *BarcodeObject {
	return &BarcodeObject{
		ReportComponentBase: *report.NewReportComponentBase(),
		Barcode:             NewCode128Barcode(),
		showText:            true,
		zoom:                1.0,
		allowExpressions:    true,
		brackets:            "[,]",
	}
}

// --- Property accessors ---

// Angle returns the rotation in degrees.
func (b *BarcodeObject) Angle() int { return b.angle }

// SetAngle sets the rotation.
func (b *BarcodeObject) SetAngle(a int) { b.angle = a }

// AutoSize returns whether bounds auto-adjust.
func (b *BarcodeObject) AutoSize() bool { return b.autoSize }

// SetAutoSize sets auto-size.
func (b *BarcodeObject) SetAutoSize(v bool) { b.autoSize = v }

// DataColumn returns the data-bound column name.
func (b *BarcodeObject) DataColumn() string { return b.dataColumn }

// SetDataColumn sets the data column.
func (b *BarcodeObject) SetDataColumn(s string) { b.dataColumn = s }

// Expression returns the barcode text expression.
func (b *BarcodeObject) Expression() string { return b.expression }

// SetExpression sets the expression.
func (b *BarcodeObject) SetExpression(s string) { b.expression = s }

// Text returns the static barcode text.
func (b *BarcodeObject) Text() string { return b.text }

// SetText sets the static text.
func (b *BarcodeObject) SetText(s string) { b.text = s }

// ShowText returns whether human-readable text is shown.
func (b *BarcodeObject) ShowText() bool { return b.showText }

// SetShowText sets show-text.
func (b *BarcodeObject) SetShowText(v bool) { b.showText = v }

// Padding returns interior padding.
func (b *BarcodeObject) Padding() Padding { return b.padding }

// SetPadding sets interior padding.
func (b *BarcodeObject) SetPadding(p Padding) { b.padding = p }

// Zoom returns the barcode scale factor.
func (b *BarcodeObject) Zoom() float32 { return b.zoom }

// SetZoom sets the zoom factor.
func (b *BarcodeObject) SetZoom(z float32) { b.zoom = z }

// HideIfNoData returns whether the object hides when text is empty.
func (b *BarcodeObject) HideIfNoData() bool { return b.hideIfNoData }

// SetHideIfNoData sets hide-if-no-data.
func (b *BarcodeObject) SetHideIfNoData(v bool) { b.hideIfNoData = v }

// NoDataText returns the text shown when data is empty.
func (b *BarcodeObject) NoDataText() string { return b.noDataText }

// SetNoDataText sets the no-data text.
func (b *BarcodeObject) SetNoDataText(s string) { b.noDataText = s }

// AllowExpressions returns whether expressions are evaluated in text.
func (b *BarcodeObject) AllowExpressions() bool { return b.allowExpressions }

// SetAllowExpressions sets allow-expressions.
func (b *BarcodeObject) SetAllowExpressions(v bool) { b.allowExpressions = v }

// Brackets returns the expression delimiter.
func (b *BarcodeObject) Brackets() string { return b.brackets }

// SetBrackets sets the delimiter.
func (b *BarcodeObject) SetBrackets(s string) { b.brackets = s }

// BarcodeType returns the type string of the current barcode symbology.
func (b *BarcodeObject) BarcodeType() BarcodeType {
	if b.Barcode == nil {
		return ""
	}
	return b.Barcode.Type()
}

// --- Serialization ---

// Serialize writes BarcodeObject properties that differ from defaults.
func (b *BarcodeObject) Serialize(w report.Writer) error {
	if err := b.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if b.Barcode != nil {
		w.WriteStr("Barcode.Type", string(b.Barcode.Type()))
	}
	if b.angle != 0 {
		w.WriteInt("Angle", b.angle)
	}
	if b.autoSize {
		w.WriteBool("AutoSize", true)
	}
	if b.dataColumn != "" {
		w.WriteStr("DataColumn", b.dataColumn)
	}
	if b.expression != "" {
		w.WriteStr("Expression", b.expression)
	}
	if b.text != "" {
		w.WriteStr("Text", b.text)
	}
	if !b.showText {
		w.WriteBool("ShowText", false)
	}
	if b.zoom != 1.0 {
		w.WriteFloat("Zoom", b.zoom)
	}
	if b.hideIfNoData {
		w.WriteBool("HideIfNoData", true)
	}
	if b.noDataText != "" {
		w.WriteStr("NoDataText", b.noDataText)
	}
	if !b.allowExpressions {
		w.WriteBool("AllowExpressions", false)
	}
	if b.brackets != "[,]" {
		w.WriteStr("Brackets", b.brackets)
	}
	return nil
}

// Deserialize reads BarcodeObject properties.
func (b *BarcodeObject) Deserialize(r report.Reader) error {
	if err := b.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	if t := r.ReadStr("Barcode.Type", ""); t != "" {
		b.Barcode = NewBarcodeByType(BarcodeType(t))
	}
	b.angle = r.ReadInt("Angle", 0)
	b.autoSize = r.ReadBool("AutoSize", false)
	b.dataColumn = r.ReadStr("DataColumn", "")
	b.expression = r.ReadStr("Expression", "")
	b.text = r.ReadStr("Text", "")
	b.showText = r.ReadBool("ShowText", true)
	b.zoom = r.ReadFloat("Zoom", 1.0)
	b.hideIfNoData = r.ReadBool("HideIfNoData", false)
	b.noDataText = r.ReadStr("NoDataText", "")
	b.allowExpressions = r.ReadBool("AllowExpressions", true)
	b.brackets = r.ReadStr("Brackets", "[,]")
	return nil
}

// -----------------------------------------------------------------------
// EAN13Barcode
// -----------------------------------------------------------------------

// EAN13Barcode implements EAN-13 symbology.
type EAN13Barcode struct{ BaseBarcodeImpl }

// NewEAN13Barcode creates an EAN13Barcode.
func NewEAN13Barcode() *EAN13Barcode {
	return &EAN13Barcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeEAN13)}
}

// Encode encodes a 12 or 13 digit EAN-13 barcode value.
func (e *EAN13Barcode) Encode(text string) error {
	bc, err := ean.Encode(text)
	if err != nil {
		return fmt.Errorf("ean13 encode: %w", err)
	}
	e.encodedText = text
	e.encoded = bc
	return nil
}

// DefaultValue returns a sample EAN-13 value.
func (e *EAN13Barcode) DefaultValue() string { return "590123412345" }

// -----------------------------------------------------------------------
// AztecBarcode
// -----------------------------------------------------------------------

// AztecBarcode implements Aztec 2D symbology.
type AztecBarcode struct {
	BaseBarcodeImpl
	// MinECCPercent is the minimum error correction percentage (default 23).
	MinECCPercent int
	// UserSpecifiedLayers configures compact/full Aztec layers (0 = auto).
	UserSpecifiedLayers int
}

// NewAztecBarcode creates an AztecBarcode.
func NewAztecBarcode() *AztecBarcode {
	return &AztecBarcode{
		BaseBarcodeImpl:     newBaseBarcodeImpl(BarcodeTypeAztec),
		MinECCPercent:       23,
		UserSpecifiedLayers: 0,
	}
}

// Encode encodes text as an Aztec barcode.
func (a *AztecBarcode) Encode(text string) error {
	bc, err := aztec.Encode([]byte(text), a.MinECCPercent, a.UserSpecifiedLayers)
	if err != nil {
		return fmt.Errorf("aztec encode: %w", err)
	}
	a.encodedText = text
	a.encoded = bc
	return nil
}

// DefaultValue returns a sample Aztec value.
func (a *AztecBarcode) DefaultValue() string { return "Aztec" }

// -----------------------------------------------------------------------
// PDF417Barcode
// -----------------------------------------------------------------------

// PDF417Barcode implements PDF417 2D symbology.
type PDF417Barcode struct {
	BaseBarcodeImpl
	// Columns is the number of data columns (0 = auto).
	Columns int
	// SecurityLevel is the error correction level 0-8 (default 2).
	SecurityLevel int
}

// NewPDF417Barcode creates a PDF417Barcode.
func NewPDF417Barcode() *PDF417Barcode {
	return &PDF417Barcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypePDF417),
		Columns:         0,
		SecurityLevel:   2,
	}
}

// Encode encodes text as a PDF417 barcode.
func (p *PDF417Barcode) Encode(text string) error {
	bc, err := pdf417.Encode(text, byte(p.SecurityLevel))
	if err != nil {
		return fmt.Errorf("pdf417 encode: %w", err)
	}
	p.encodedText = text
	p.encoded = bc
	return nil
}

// DefaultValue returns a sample PDF417 value.
func (p *PDF417Barcode) DefaultValue() string { return "PDF417" }

// -----------------------------------------------------------------------
// Code93Barcode — Code 93 linear symbology
// -----------------------------------------------------------------------

// Code93Barcode implements Code 93 symbology.
type Code93Barcode struct {
	BaseBarcodeImpl
	// IncludeChecksum adds checksum characters (default: true).
	IncludeChecksum bool
	// FullASCIIMode enables full ASCII encoding (default: false).
	FullASCIIMode bool
}

// NewCode93Barcode creates a Code93Barcode with defaults.
func NewCode93Barcode() *Code93Barcode {
	return &Code93Barcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode93),
		IncludeChecksum: true,
	}
}

// Encode stores the text for later rendering.
func (c *Code93Barcode) Encode(text string) error {
	c.encodedText = text
	return nil
}

// Render renders the Code 93 barcode to an image of the given size.
func (c *Code93Barcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code93: Encode must be called before Render")
	}
	enc := code93.New()
	enc.IncludeChecksum = c.IncludeChecksum
	enc.FullASCIIMode = c.FullASCIIMode
	return enc.Encode(c.encodedText, width, height)
}

// DefaultValue returns a sample Code 93 value.
func (c *Code93Barcode) DefaultValue() string { return "CODE93" }

// -----------------------------------------------------------------------
// Code2of5Barcode — 2-of-5 / ITF linear symbology
// -----------------------------------------------------------------------

// Code2of5Barcode implements the 2-of-5 (interleaved) barcode symbology.
type Code2of5Barcode struct {
	BaseBarcodeImpl
	// Interleaved selects Interleaved 2-of-5 (default: true).
	Interleaved bool
}

// NewCode2of5Barcode creates a Code2of5Barcode with defaults.
func NewCode2of5Barcode() *Code2of5Barcode {
	return &Code2of5Barcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode2of5),
		Interleaved:     true,
	}
}

// Encode stores the text for later rendering.
func (c *Code2of5Barcode) Encode(text string) error {
	c.encodedText = text
	return nil
}

// Render renders the 2-of-5 barcode to an image of the given size.
func (c *Code2of5Barcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code2of5: Encode must be called before Render")
	}
	enc := code2of5.New()
	enc.Interleaved = c.Interleaved
	return enc.Encode(c.encodedText, width, height)
}

// DefaultValue returns a sample 2-of-5 value.
func (c *Code2of5Barcode) DefaultValue() string { return "12345670" }

// -----------------------------------------------------------------------
// CodabarBarcode — Codabar linear symbology
// -----------------------------------------------------------------------

// CodabarBarcode implements Codabar symbology.
type CodabarBarcode struct {
	BaseBarcodeImpl
}

// NewCodabarBarcode creates a CodabarBarcode.
func NewCodabarBarcode() *CodabarBarcode {
	return &CodabarBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCodabar)}
}

// Encode stores the text for later rendering.
func (c *CodabarBarcode) Encode(text string) error {
	c.encodedText = text
	return nil
}

// Render renders the Codabar barcode to an image of the given size.
func (c *CodabarBarcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("codabar: Encode must be called before Render")
	}
	enc := codabar.New()
	return enc.Encode(c.encodedText, width, height)
}

// DefaultValue returns a sample Codabar value.
func (c *CodabarBarcode) DefaultValue() string { return "A12345B" }

// -----------------------------------------------------------------------
// DataMatrixBarcode — DataMatrix 2D symbology
// -----------------------------------------------------------------------

// DataMatrixBarcode implements DataMatrix 2D symbology.
type DataMatrixBarcode struct {
	BaseBarcodeImpl
}

// NewDataMatrixBarcode creates a DataMatrixBarcode.
func NewDataMatrixBarcode() *DataMatrixBarcode {
	return &DataMatrixBarcode{BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeDataMatrix)}
}

// Encode stores the text for later rendering.
func (d *DataMatrixBarcode) Encode(text string) error {
	d.encodedText = text
	return nil
}

// Render renders the DataMatrix barcode to an image of the given size.
func (d *DataMatrixBarcode) Render(width, height int) (image.Image, error) {
	if d.encodedText == "" {
		return nil, fmt.Errorf("datamatrix: Encode must be called before Render")
	}
	enc := datamatrix.New()
	return enc.Encode(d.encodedText, width, height)
}

// DefaultValue returns a sample DataMatrix value.
func (d *DataMatrixBarcode) DefaultValue() string { return "DataMatrix" }

// -----------------------------------------------------------------------
// Factory
// -----------------------------------------------------------------------

// NewBarcodeByType constructs a BarcodeBase from a BarcodeType string.
// Returns a Code128Barcode for unknown types.
func NewBarcodeByType(t BarcodeType) BarcodeBase {
	switch t {
	case BarcodeTypeCode128:
		return NewCode128Barcode()
	case BarcodeTypeGS1_128:
		return NewGS1Barcode()
	case BarcodeTypeCode39:
		return NewCode39Barcode()
	case BarcodeTypeQR:
		return NewQRBarcode()
	case BarcodeTypeEAN13, BarcodeTypeEAN8, BarcodeTypeUPCA, BarcodeTypeUPCE:
		return NewEAN13Barcode()
	case BarcodeTypeAztec:
		return NewAztecBarcode()
	case BarcodeTypePDF417:
		return NewPDF417Barcode()
	case BarcodeTypeCode93:
		return NewCode93Barcode()
	case BarcodeTypeCode2of5:
		return NewCode2of5Barcode()
	case BarcodeTypeCodabar:
		return NewCodabarBarcode()
	case BarcodeTypeDataMatrix:
		return NewDataMatrixBarcode()
	case BarcodeTypeMSI:
		return NewMSIBarcode()
	case BarcodeTypeMaxiCode:
		return NewMaxiCodeBarcode()
	case BarcodeTypeIntelligentMail:
		return NewIntelligentMailBarcode()
	case BarcodeTypePharmacode:
		return NewPharmacodeBarcode()
	case BarcodeTypePlessey:
		return NewPlesseyBarcode()
	case BarcodeTypePostNet:
		return NewPostNetBarcode()
	case BarcodeTypeSwissQR:
		return NewSwissQRBarcode()
	default:
		return NewCode128Barcode()
	}
}
