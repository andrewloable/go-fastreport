// Package barcode implements barcode rendering support for go-fastreport.
// It provides the BarcodeBase interface and the BarcodeObject report element.
package barcode

import (
	"fmt"
	"image"
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
	BarcodeTypeCode128              BarcodeType = "Code128"
	BarcodeTypeCode128A             BarcodeType = "Code128A"
	BarcodeTypeCode128B             BarcodeType = "Code128B"
	BarcodeTypeCode128C             BarcodeType = "Code128C"
	BarcodeTypeCode39               BarcodeType = "Code39"
	BarcodeTypeCode39Extended       BarcodeType = "Code39Extended"
	BarcodeTypeCode93               BarcodeType = "Code93"
	BarcodeTypeCode93Extended       BarcodeType = "Code93Extended"
	BarcodeTypeCode2of5             BarcodeType = "2of5"
	BarcodeTypeCode2of5Industrial   BarcodeType = "2of5Industrial"
	BarcodeTypeCode2of5Matrix       BarcodeType = "2of5Matrix"
	BarcodeTypeCodabar              BarcodeType = "Codabar"
	BarcodeTypeEAN13                BarcodeType = "EAN13"
	BarcodeTypeEAN8                 BarcodeType = "EAN8"
	BarcodeTypeUPCA                 BarcodeType = "UPCA"
	BarcodeTypeUPCE                 BarcodeType = "UPCE"
	BarcodeTypeMSI                  BarcodeType = "MSI"
	BarcodeTypeQR                   BarcodeType = "QR"
	BarcodeTypeDataMatrix           BarcodeType = "DataMatrix"
	BarcodeTypeAztec                BarcodeType = "Aztec"
	BarcodeTypeMaxiCode             BarcodeType = "MaxiCode"
	BarcodeTypePDF417               BarcodeType = "PDF417"
	BarcodeTypeGS1_128              BarcodeType = "GS1-128"
	BarcodeTypeITF14                BarcodeType = "ITF14"
	BarcodeTypeDeutscheIdentcode    BarcodeType = "DeutscheIdentcode"
	BarcodeTypeDeutscheLeitcode     BarcodeType = "DeutscheLeitcode"
	BarcodeTypeSupplement2          BarcodeType = "Supplement2"
	BarcodeTypeSupplement5          BarcodeType = "Supplement5"
	BarcodeTypeJapanPost4State      BarcodeType = "JapanPost4State"
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
	// CalcBounds returns the natural (width, height) of the barcode in module units
	// after encoding. Returns (0, 0) if not encoded or not supported (keeps FRX dims).
	CalcBounds() (float32, float32)
	// EncodedText returns the canonical display text stored after Encode.
	// For GS1 types this includes the "(01)" prefix and checksum digit.
	EncodedText() string
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
	// wideBarRatioOverride is set by FRX deserialization when Barcode.WideBarRatio
	// is present in the FRX file. A zero value means "use the type's default".
	// C# LinearBarcodeBase: barcode.WideBarRatio = Reader.ReadFloat("Barcode.WideBarRatio", 2)
	wideBarRatioOverride float32
}

// SetWideBarRatio overrides the WideBarRatio for this barcode instance.
// Called during FRX deserialization when Barcode.WideBarRatio is present.
func (b *BaseBarcodeImpl) SetWideBarRatio(v float32) { b.wideBarRatioOverride = v }

// WBROverride returns the FRX-deserialized WideBarRatio override (0 if not set).
func (b *BaseBarcodeImpl) WBROverride() float32 { return b.wideBarRatioOverride }

// Render returns an error indicating the barcode type must provide its own Render.
// Concrete types override this by implementing Render via pattern or matrix rendering.
func (b *BaseBarcodeImpl) Render(width, height int) (image.Image, error) {
	if b.encodedText == "" {
		return nil, fmt.Errorf("barcode not encoded — call Encode first")
	}
	// Fallback: return a placeholder image. Concrete types should override.
	return placeholderImage(width, height), nil
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

// CalcBounds returns (0, 0) by default; concrete types override as needed.
func (b *BaseBarcodeImpl) CalcBounds() (float32, float32) { return 0, 0 }

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

// Encode validates and stores text for Code128 encoding.
func (c *Code128Barcode) Encode(text string) error {
	if text == "" {
		return fmt.Errorf("code128: text must not be empty")
	}
	c.encodedText = text
	return nil
}

// Render renders the Code128 barcode using the native pattern-based renderer.
func (c *Code128Barcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code128: Encode must be called before Render")
	}
	pattern, err := c.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, c.encodedText, width, height, true, c.GetWideBarRatio()), nil
}

// DefaultValue returns "12345678" matching C# BarcodeBase.GetDefaultValue() (no override in Barcode128.cs).
func (c *Code128Barcode) DefaultValue() string { return "12345678" }

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

// Encode validates and stores text for Code39 encoding.
func (c *Code39Barcode) Encode(text string) error {
	if !c.AllowExtended {
		for _, ch := range text {
			if ch >= 'a' && ch <= 'z' {
				return fmt.Errorf("code39: lowercase not allowed without AllowExtended, found %q", ch)
			}
		}
	}
	c.encodedText = text
	return nil
}

// Render renders the Code39 barcode using the native pattern-based renderer.
func (c *Code39Barcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code39: Encode must be called before Render")
	}
	pattern, err := c.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, c.encodedText, width, height, true, c.GetWideBarRatio()), nil
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

// Encode validates and stores text for QR encoding.
func (q *QRBarcode) Encode(text string) error {
	q.encodedText = text
	return nil
}

// Render renders the QR barcode using the native matrix-based renderer.
func (q *QRBarcode) Render(width, height int) (image.Image, error) {
	if q.encodedText == "" {
		return nil, fmt.Errorf("qr: Encode must be called before Render")
	}
	matrix, rows, cols := q.GetMatrix()
	return DrawBarcode2D(matrix, rows, cols, width, height), nil
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

// Horizontal returns left + right padding.
func (p Padding) Horizontal() float32 { return p.Left + p.Right }

// Vertical returns top + bottom padding.
func (p Padding) Vertical() float32 { return p.Top + p.Bottom }

// NewBarcodeObject creates a BarcodeObject with defaults.
// AutoSize defaults to true, matching C# BarcodeObject (BarcodeObject.cs:689).
func NewBarcodeObject() *BarcodeObject {
	return &BarcodeObject{
		ReportComponentBase: *report.NewReportComponentBase(),
		Barcode:             NewCode128Barcode(),
		autoSize:            true,
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

// UpdateAutoSize resizes the BarcodeObject to fit the barcode's natural dimensions.
// Mirrors C# BarcodeObject.UpdateAutoSize() (BarcodeObject.cs:390-412) and
// C# LinearBarcodeBase.CalcBounds() ShowText extra-padding logic (LinearBarcodeBase.cs:435-452).
// Called after Encode() so that CalcBounds() can use the encoded data.
func (b *BarcodeObject) UpdateAutoSize() {
	if b.Barcode == nil {
		return
	}
	w, h := b.Barcode.CalcBounds()
	if w == 0 && h == 0 {
		return // barcode type doesn't support CalcBounds, keep FRX dimensions
	}

	// Apply ShowText extra padding for linear (1-D) barcodes only.
	// C# LinearBarcodeBase.CalcBounds() measures the human-readable text with GDI+
	// and adds symmetric padding when the text is wider than the bar area.
	// h == 0 is the Go convention for linear barcodes (2-D barcodes return height > 0).
	if b.showText && h == 0 {
		displayText := b.Barcode.EncodedText()
		if displayText != "" {
			// Undo the 1.25 scaling applied by CalcBounds to recover raw bar width.
			barWidth := w / 1.25
			// Approximate Arial 8pt GDI+ MeasureString at 96 DPI.
			// C#: using (Graphics g = ...) txtWidth = g.MeasureString(text, Font, 100000).Width
			// Average character advance for Arial Regular 8pt at 96 DPI ≈ fontPx × 0.542.
			const arialAvgWidthFactor = 0.542
			fontPx := float32(8.0) * 96.0 / 72.0 // barcode default: Arial 8pt
			txtWidth := float32(len(displayText)) * fontPx * arialAvgWidthFactor
			if barWidth < txtWidth {
				extra := (txtWidth-barWidth)/2 + 2
				w = (barWidth + 2*extra) * 1.25
			}
		}
	}

	w *= b.zoom
	h *= b.zoom
	if !b.autoSize {
		return
	}
	if b.angle == 0 || b.angle == 180 {
		b.SetWidth(w + b.padding.Horizontal())
		if h > 0 {
			b.SetHeight(h + b.padding.Vertical())
		}
	} else if b.angle == 90 || b.angle == 270 {
		b.SetHeight(w + b.padding.Vertical())
		if h > 0 {
			b.SetWidth(h + b.padding.Horizontal())
		}
	}
}

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
	if !b.autoSize {
		w.WriteBool("AutoSize", false)
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
	// FRX uses Barcode="QR Code" (display name) while our internal format
	// uses Barcode.Type="QR". Try the internal key first, then fall back to
	// the FRX display-name attribute.
	if t := r.ReadStr("Barcode.Type", ""); t != "" {
		b.Barcode = NewBarcodeByType(BarcodeType(t))
	} else if name := r.ReadStr("Barcode", ""); name != "" {
		b.Barcode = NewBarcodeByName(name)
	}
	// Read barcode-specific properties (prefixed with "Barcode.").
	if b.Barcode != nil {
		if qrbc, ok := b.Barcode.(*QRBarcode); ok {
			if ec := r.ReadStr("Barcode.ErrorCorrection", ""); ec != "" {
				qrbc.ErrorCorrection = ec
			}
		}
		// FRX stores the WideBarRatio override as Barcode.WideBarRatio="2.25".
		// C# LinearBarcodeBase: barcode.WideBarRatio = Reader.ReadFloat("Barcode.WideBarRatio", 2)
		// Go: we use 0 as the "not-set" sentinel and only override when explicitly present.
		if wbr := r.ReadFloat("Barcode.WideBarRatio", 0); wbr != 0 {
			type wbrSetter interface{ SetWideBarRatio(float32) }
			if setter, ok := b.Barcode.(wbrSetter); ok {
				setter.SetWideBarRatio(wbr)
			}
		}
	}
	b.angle = r.ReadInt("Angle", 0)
	b.autoSize = r.ReadBool("AutoSize", true)
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

// Encode validates and stores a 12 or 13 digit EAN-13 barcode value.
func (e *EAN13Barcode) Encode(text string) error {
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("ean13: only digits allowed, found %q", ch)
		}
	}
	if len(text) < 12 {
		return fmt.Errorf("ean13: expected at least 12 digits, got %d", len(text))
	}
	e.encodedText = text
	return nil
}

// Render renders the EAN-13 barcode using the native pattern-based renderer.
func (e *EAN13Barcode) Render(width, height int) (image.Image, error) {
	if e.encodedText == "" {
		return nil, fmt.Errorf("ean13: Encode must be called before Render")
	}
	pattern, err := e.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, e.encodedText, width, height, true, e.GetWideBarRatio()), nil
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

// Encode validates and stores text for Aztec encoding.
func (a *AztecBarcode) Encode(text string) error {
	if a.UserSpecifiedLayers > 32 {
		return fmt.Errorf("aztec: UserSpecifiedLayers must be 0-32, got %d", a.UserSpecifiedLayers)
	}
	a.encodedText = text
	return nil
}

// Render renders the Aztec barcode using the native matrix-based renderer.
// The Aztec matrix is generated by the aztec subpackage encoder.
func (a *AztecBarcode) Render(width, height int) (image.Image, error) {
	if a.encodedText == "" {
		return nil, fmt.Errorf("aztec: Encode must be called before Render")
	}
	matrix, rows, cols := a.GetMatrix()
	return DrawBarcode2D(matrix, rows, cols, width, height), nil
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

// Encode validates and stores text for PDF417 encoding.
func (p *PDF417Barcode) Encode(text string) error {
	if p.SecurityLevel < 0 || p.SecurityLevel > 8 {
		return fmt.Errorf("pdf417: SecurityLevel must be 0-8, got %d", p.SecurityLevel)
	}
	p.encodedText = text
	return nil
}

// Render renders the PDF417 barcode using the native matrix-based renderer.
func (p *PDF417Barcode) Render(width, height int) (image.Image, error) {
	if p.encodedText == "" {
		return nil, fmt.Errorf("pdf417: Encode must be called before Render")
	}
	matrix, rows, cols := p.GetMatrix()
	return DrawBarcode2D(matrix, rows, cols, width, height), nil
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

// Render renders the Code 93 barcode using the native pattern-based renderer.
func (c *Code93Barcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code93: Encode must be called before Render")
	}
	pattern, err := c.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, c.encodedText, width, height, true, c.GetWideBarRatio()), nil
}

// DefaultValue returns "12345678" matching C# BarcodeBase.GetDefaultValue() (no override in Barcode93.cs).
func (c *Code93Barcode) DefaultValue() string { return "12345678" }

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

// Encode validates and stores the text for later rendering.
func (c *Code2of5Barcode) Encode(text string) error {
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("code2of5: only digits allowed, found %q", ch)
		}
	}
	c.encodedText = text
	return nil
}

// Render renders the 2-of-5 barcode using the native pattern-based renderer.
func (c *Code2of5Barcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("code2of5: Encode must be called before Render")
	}
	pattern, err := c.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, c.encodedText, width, height, true, c.GetWideBarRatio()), nil
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

// Render renders the Codabar barcode using the native pattern-based renderer.
func (c *CodabarBarcode) Render(width, height int) (image.Image, error) {
	if c.encodedText == "" {
		return nil, fmt.Errorf("codabar: Encode must be called before Render")
	}
	pattern, err := c.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcode(pattern, c.encodedText, width, height, true, c.GetWideBarRatio()), nil
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

// Render renders the DataMatrix barcode using the native matrix-based renderer.
func (d *DataMatrixBarcode) Render(width, height int) (image.Image, error) {
	if d.encodedText == "" {
		return nil, fmt.Errorf("datamatrix: Encode must be called before Render")
	}
	matrix, rows, cols := d.GetMatrix()
	return DrawBarcode2D(matrix, rows, cols, width, height), nil
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
	case BarcodeTypeCode128A:
		return NewCode128ABarcode()
	case BarcodeTypeCode128B:
		return NewCode128BBarcode()
	case BarcodeTypeCode128C:
		return NewCode128CBarcode()
	case BarcodeTypeGS1_128:
		return NewGS1Barcode()
	case BarcodeTypeCode39:
		return NewCode39Barcode()
	case BarcodeTypeQR:
		return NewQRBarcode()
	case BarcodeTypeEAN13:
		return NewEAN13Barcode()
	case BarcodeTypeEAN8:
		return NewEAN8Barcode()
	case BarcodeTypeUPCA:
		return NewUPCABarcode()
	case BarcodeTypeUPCE:
		return NewUPCEBarcode()
	case BarcodeTypeAztec:
		return NewAztecBarcode()
	case BarcodeTypePDF417:
		return NewPDF417Barcode()
	case BarcodeTypeCode93:
		return NewCode93Barcode()
	case BarcodeTypeCode93Extended:
		return NewCode93ExtendedBarcode()
	case BarcodeTypeCode2of5:
		return NewCode2of5Barcode()
	case BarcodeTypeCode2of5Industrial:
		return NewCode2of5IndustrialBarcode()
	case BarcodeTypeCode2of5Matrix:
		return NewCode2of5MatrixBarcode()
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
	case BarcodeTypeITF14:
		return NewITF14Barcode()
	case BarcodeTypeDeutscheIdentcode:
		return NewDeutscheIdentcodeBarcode()
	case BarcodeTypeDeutscheLeitcode:
		return NewDeutscheLeitcodeBarcode()
	case BarcodeTypeSupplement2:
		return NewSupplement2Barcode()
	case BarcodeTypeSupplement5:
		return NewSupplement5Barcode()
	case BarcodeTypeJapanPost4State:
		return NewJapanPost4StateBarcode()
	case BarcodeTypeGS1DataBarOmni:
		return NewGS1DataBarOmniBarcode()
	case BarcodeTypeGS1DataBarStacked:
		return NewGS1DataBarStackedBarcode()
	case BarcodeTypeGS1DataBarStackedOmni:
		return NewGS1DataBarStackedOmniBarcode()
	case BarcodeTypeGS1DataBarLimited:
		return NewGS1DataBarLimitedBarcode()
	default:
		return NewCode128Barcode()
	}
}

// barcodeDisplayNames maps FRX display names (as used by FastReport .NET in the
// Barcode="..." attribute) to internal BarcodeType constants.
var barcodeDisplayNames = map[string]BarcodeType{
	"2/5 Interleaved":                     BarcodeTypeCode2of5,
	"2/5 Industrial":                      BarcodeTypeCode2of5Industrial,
	"2/5 Matrix":                          BarcodeTypeCode2of5Matrix,
	"Codabar":                             BarcodeTypeCodabar,
	"Code128":                             BarcodeTypeCode128,
	"Code128 A":                           BarcodeTypeCode128A,
	"Code128 B":                           BarcodeTypeCode128B,
	"Code128 C":                           BarcodeTypeCode128C,
	"Code39":                              BarcodeTypeCode39,
	"Code39 Extended":                     BarcodeTypeCode39,
	"Code93":                              BarcodeTypeCode93,
	"Code93 Extended":                     BarcodeTypeCode93Extended,
	"EAN8":                                BarcodeTypeEAN8,
	"EAN13":                               BarcodeTypeEAN13,
	"MSI":                                 BarcodeTypeMSI,
	"PostNet":                             BarcodeTypePostNet,
	"UPC-A":                               BarcodeTypeUPCA,
	"UPC-E0":                              BarcodeTypeUPCE,
	"UPC-E1":                              BarcodeTypeUPCE,
	"PDF417":                              BarcodeTypePDF417,
	"Datamatrix":                          BarcodeTypeDataMatrix,
	"QR Code":                             BarcodeTypeQR,
	"Aztec":                               BarcodeTypeAztec,
	"Plessey":                             BarcodeTypePlessey,
	"GS1-128 (UCC/EAN-128)":               BarcodeTypeGS1_128,
	"Pharmacode":                          BarcodeTypePharmacode,
	"Intelligent Mail (USPS)":             BarcodeTypeIntelligentMail,
	"MaxiCode":                            BarcodeTypeMaxiCode,
	"ITF-14":                              BarcodeTypeITF14,
	"Deutsche Identcode":                  BarcodeTypeDeutscheIdentcode,
	"Deutsche Leitcode":                   BarcodeTypeDeutscheLeitcode,
	"Japan Post 4 State Code":             BarcodeTypeJapanPost4State,
	"Supplement 2":                        BarcodeTypeSupplement2,
	"Supplement 5":                        BarcodeTypeSupplement5,
	"GS1 DataBar Omnidirectional":         BarcodeTypeGS1DataBarOmni,
	"GS1 DataBar Limited":                 BarcodeTypeGS1DataBarLimited,
	"GS1 DataBar Stacked":                 BarcodeTypeGS1DataBarStacked,
	"GS1 DataBar Stacked Omnidirectional": BarcodeTypeGS1DataBarStackedOmni,
	"GS1 Datamatrix":                      BarcodeTypeDataMatrix,
	"SwissQR":                             BarcodeTypeSwissQR,
}

// NewBarcodeByName constructs a BarcodeBase from an FRX display name
// (e.g. "QR Code", "Code128", "EAN13"). Falls back to Code128 for unknown names.
func NewBarcodeByName(name string) BarcodeBase {
	if t, ok := barcodeDisplayNames[name]; ok {
		return NewBarcodeByType(t)
	}
	return NewCode128Barcode()
}
