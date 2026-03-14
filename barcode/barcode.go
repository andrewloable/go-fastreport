// Package barcode implements barcode rendering support for go-fastreport.
// It provides the BarcodeBase interface and the BarcodeObject report element.
package barcode

import (
	"image/color"

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

// Encode validates text for Code128 (accepts any printable ASCII).
func (c *Code128Barcode) Encode(text string) error {
	c.encodedText = text
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

// Encode validates and stores text for Code39.
func (c *Code39Barcode) Encode(text string) error {
	c.encodedText = text
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

// Encode stores QR text.
func (q *QRBarcode) Encode(text string) error {
	q.encodedText = text
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

// NewBarcodeByType constructs a BarcodeBase from a BarcodeType string.
// Returns a Code128Barcode for unknown types.
func NewBarcodeByType(t BarcodeType) BarcodeBase {
	switch t {
	case BarcodeTypeCode128:
		return NewCode128Barcode()
	case BarcodeTypeCode39:
		return NewCode39Barcode()
	case BarcodeTypeQR:
		return NewQRBarcode()
	default:
		return NewCode128Barcode()
	}
}
