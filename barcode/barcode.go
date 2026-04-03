// Package barcode implements barcode rendering support for go-fastreport.
// It provides the BarcodeBase interface and the BarcodeObject report element.
package barcode

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/andrewloable/go-fastreport/expr"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
)

// -----------------------------------------------------------------------
// BarcodeType registry
// -----------------------------------------------------------------------

// BarcodeType identifies a barcode symbology.
type BarcodeType string

const (
	BarcodeTypeCode128            BarcodeType = "Code128"
	BarcodeTypeCode128A           BarcodeType = "Code128A"
	BarcodeTypeCode128B           BarcodeType = "Code128B"
	BarcodeTypeCode128C           BarcodeType = "Code128C"
	BarcodeTypeCode39             BarcodeType = "Code39"
	BarcodeTypeCode39Extended     BarcodeType = "Code39Extended"
	BarcodeTypeCode93             BarcodeType = "Code93"
	BarcodeTypeCode93Extended     BarcodeType = "Code93Extended"
	BarcodeTypeCode2of5           BarcodeType = "2of5"
	BarcodeTypeCode2of5Industrial BarcodeType = "2of5Industrial"
	BarcodeTypeCode2of5Matrix     BarcodeType = "2of5Matrix"
	BarcodeTypeCodabar            BarcodeType = "Codabar"
	BarcodeTypeEAN13              BarcodeType = "EAN13"
	BarcodeTypeEAN8               BarcodeType = "EAN8"
	BarcodeTypeUPCA               BarcodeType = "UPCA"
	BarcodeTypeUPCE               BarcodeType = "UPCE"
	BarcodeTypeMSI                BarcodeType = "MSI"
	BarcodeTypeQR                 BarcodeType = "QR"
	BarcodeTypeDataMatrix         BarcodeType = "DataMatrix"
	BarcodeTypeAztec              BarcodeType = "Aztec"
	BarcodeTypeMaxiCode           BarcodeType = "MaxiCode"
	BarcodeTypePDF417             BarcodeType = "PDF417"
	BarcodeTypeGS1_128            BarcodeType = "GS1-128"
	BarcodeTypeITF14              BarcodeType = "ITF14"
	BarcodeTypeDeutscheIdentcode  BarcodeType = "DeutscheIdentcode"
	BarcodeTypeDeutscheLeitcode   BarcodeType = "DeutscheLeitcode"
	BarcodeTypeSupplement2        BarcodeType = "Supplement2"
	BarcodeTypeSupplement5        BarcodeType = "Supplement5"
	BarcodeTypeJapanPost4State    BarcodeType = "JapanPost4State"
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
	// ratioMin and ratioMax are the clamping bounds for WideBarRatio.
	// C# LinearBarcodeBase.cs:40-41: internal float ratioMin, ratioMax.
	// Set in each barcode type's constructor; zero means "no bound".
	ratioMin float32
	ratioMax float32
	// showText is propagated from BarcodeObject.ShowText before CalcBounds.
	// Used by 2D barcodes to add font height to the bounds.
	showText bool
}

// SetWideBarRatio overrides the WideBarRatio for this barcode instance.
// Called during FRX deserialization when Barcode.WideBarRatio is present.
// Applies clamping to ratioMin/ratioMax per C# LinearBarcodeBase.cs:66-73.
func (b *BaseBarcodeImpl) SetWideBarRatio(v float32) {
	b.wideBarRatioOverride = v
	if b.ratioMin != 0 && b.wideBarRatioOverride < b.ratioMin {
		b.wideBarRatioOverride = b.ratioMin
	}
	if b.ratioMax != 0 && b.wideBarRatioOverride > b.ratioMax {
		b.wideBarRatioOverride = b.ratioMax
	}
}

// WBROverride returns the FRX-deserialized WideBarRatio override (0 if not set).
func (b *BaseBarcodeImpl) WBROverride() float32 { return b.wideBarRatioOverride }

// SetBarcodeColor sets the bar color for this barcode instance.
// Implements the colorSetter interface used during FRX deserialization.
// C# ref: LinearBarcodeBase.Color property (LinearBarcodeBase.cs:69).
func (b *BaseBarcodeImpl) SetBarcodeColor(c color.RGBA) { b.Color = c }

// clampedWBR returns the effective WideBarRatio: the override if non-zero,
// otherwise typeDefault; the chosen value is then clamped to [ratioMin, ratioMax].
// C# LinearBarcodeBase.WideBarRatio getter always returns the already-clamped value.
// This helper is called by each concrete GetWideBarRatio() implementation.
func (b *BaseBarcodeImpl) clampedWBR(typeDefault float32) float32 {
	v := typeDefault
	if b.wideBarRatioOverride != 0 {
		v = b.wideBarRatioOverride
	}
	if b.ratioMin != 0 && v < b.ratioMin {
		v = b.ratioMin
	}
	if b.ratioMax != 0 && v > b.ratioMax {
		v = b.ratioMax
	}
	return v
}

// SetShowText propagates the BarcodeObject.ShowText flag to the barcode implementation.
// Used by 2D barcodes to include font height in CalcBounds.
func (b *BaseBarcodeImpl) SetShowText(v bool) { b.showText = v }

// GetShowText returns the showText flag.
func (b *BaseBarcodeImpl) GetShowText() bool { return b.showText }

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

// DefaultValue returns the C# BarcodeBase.GetDefaultValue() default: "12345678".
func (b *BaseBarcodeImpl) DefaultValue() string { return "12345678" }

// CalcBounds returns (0, 0) by default; concrete types override as needed.
func (b *BaseBarcodeImpl) CalcBounds() (float32, float32) { return 0, 0 }

// EncodedText returns the validated text stored by the last Encode call.
func (b *BaseBarcodeImpl) EncodedText() string { return b.encodedText }

// -----------------------------------------------------------------------
// Concrete simple barcodes (linear 1-D)
// -----------------------------------------------------------------------

// Code128Barcode implements Code128 symbology.
type Code128Barcode struct {
	BaseBarcodeImpl
	// AutoEncode enables automatic Code128 subcode selection (A/B/C).
	// C# Barcode128.AutoEncode default is true (Barcode128.cs:591).
	AutoEncode bool
}

// NewCode128Barcode creates a Code128Barcode.
func NewCode128Barcode() *Code128Barcode {
	return &Code128Barcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeCode128),
		AutoEncode:      true, // C# Barcode128 default
	}
}

// SetAutoEncode implements autoEncodeSetter for FRX deserialization.
// C# Barcode.AutoEncode attribute controls automatic Code128 subcode selection.
func (c *Code128Barcode) SetAutoEncode(v bool) { c.AutoEncode = v }

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
	return drawLinearBarcodeColored(pattern, c.encodedText, width, height, true, c.GetWideBarRatio(), c.Color), nil
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
// C# LinearBarcodeBase default: calcCheckSum=true (LinearBarcodeBase.cs:637).
// C# Barcode39 constructor: ratioMin=2, ratioMax=3 (Barcode39.cs:137-138).
func NewCode39Barcode() *Code39Barcode {
	b := newBaseBarcodeImpl(BarcodeTypeCode39)
	b.ratioMin = 2
	b.ratioMax = 3
	return &Code39Barcode{
		BaseBarcodeImpl: b,
		CalcChecksum:    true,
	}
}

// SetCalcCheckSum implements calcCheckSumSetter for FRX deserialization.
// C# Barcode.CalcCheckSum attribute controls whether a check digit is appended.
func (c *Code39Barcode) SetCalcCheckSum(v bool) { c.CalcChecksum = v }

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
	return drawLinearBarcodeColored(pattern, c.encodedText, width, height, true, c.GetWideBarRatio(), c.Color), nil
}

// DefaultValue returns the C# BarcodeBase.GetDefaultValue() default: "12345678".
func (c *Code39Barcode) DefaultValue() string { return "12345678" }

// QRBarcode implements QR Code 2D symbology.
type QRBarcode struct {
	BaseBarcodeImpl
	// ErrorCorrection is the QR error correction level (L/M/Q/H).
	// C# BarcodeQR.ErrorCorrection default is M (BarcodeQR.cs).
	ErrorCorrection string
	// QuietZone adds a 4-module white border around the QR matrix.
	// C# BarcodeQR.QuietZone default is true (BarcodeQR.cs:902).
	QuietZone bool
	// Encoding is the QR character encoding (UTF8, ISO8859_1, Shift_JIS, etc.).
	// C# QRCodeEncoding enum, default UTF8 (BarcodeQR.cs:153).
	Encoding string
	// ShowMarker controls whether finder-pattern markers are drawn.
	// C# BarcodeQR.showMarker, default false (BarcodeQR.cs:281).
	ShowMarker bool
	// Shape is the module shape: Rectangle, Circle, Diamond, RoundedSquare,
	// PillHorizontal, PillVertical, Plus, Hexagon, Star, Snowflake.
	// C# QrModuleShape enum, default Rectangle (BarcodeQR.cs:173).
	Shape string
	// UseThinModules adds a 10% inset to non-finder modules.
	// C# BarcodeQR.UseThinModules default is false.
	UseThinModules bool
	// Angle is the rotation angle in degrees applied to rotational module shapes
	// (Hexagon, Star, Snowflake). Has no effect on non-rotational shapes.
	// C# BarcodeQR.Angle, default 0 (BarcodeQR.cs:198).
	Angle int
}

// NewQRBarcode creates a QRBarcode.
func NewQRBarcode() *QRBarcode {
	return &QRBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeQR),
		ErrorCorrection: "L",         // C# BarcodeQR.cs:143 [DefaultValue(QRCodeErrorCorrection.L)]
		QuietZone:       true,        // C# BarcodeQR default
		Encoding:        "UTF8",      // C# QRCodeEncoding.UTF8 default
		Shape:           "Rectangle", // C# QrModuleShape.Rectangle default
	}
}

// Encode validates and stores text for QR encoding.
func (q *QRBarcode) Encode(text string) error {
	q.encodedText = normalizeSwissQRPayload(text)
	return nil
}

// Render renders the QR barcode using the native matrix-based renderer.
// For Swiss QR payloads (text starting with "SPC") the module shape is forced
// to Rectangle and UseThinModules is disabled, matching C# BarcodeQR.Draw2DBarcode
// (BarcodeQR.cs:311–315). After the QR modules are drawn the Swiss cross overlay
// is applied at the image centre per C# Barcode2DBase.DrawBarcode (Barcode2DBase.cs:22–29).
func (q *QRBarcode) Render(width, height int) (image.Image, error) {
	if q.encodedText == "" {
		return nil, fmt.Errorf("qr: Encode must be called before Render")
	}
	shape := q.Shape
	useThin := q.UseThinModules
	if isSwissQRPayload(q.encodedText) {
		// C# BarcodeQR.Draw2DBarcode: force Rectangle shape and no thin modules.
		shape = "Rectangle"
		useThin = false
	}
	matrix, rows, cols := q.GetMatrix()
	img := DrawQRCode2D(matrix, rows, cols, width, height, shape, useThin, q.QuietZone, q.Angle)
	if isSwissQRPayload(q.encodedText) {
		rgba, ok := img.(*image.RGBA)
		if !ok {
			// Convert to RGBA so we can overlay the cross.
			rgba = image.NewRGBA(img.Bounds())
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					rgba.Set(x, y, img.At(x, y))
				}
			}
		}
		DrawSwissCross(rgba, width, height)
		return rgba, nil
	}
	return img, nil
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
	// trim trims leading/trailing whitespace from barcode text before encoding.
	// C# LinearBarcodeBase.Trim (default=true).
	trim bool
	// horzAlign is the horizontal alignment within the object bounds.
	// C# BarcodeObject.HorzAlign, default Left (BarcodeObject.cs:119).
	horzAlign BarcodeHorzAlign
	// showMarker controls whether finder markers are shown (e.g. QR corner squares).
	// C# BarcodeObject.ShowMarker, default false (BarcodeObject.cs:329).
	showMarker bool
	// asBitmap forces rendering as a bitmap image instead of vector graphics.
	// C# BarcodeObject.AsBitmap, default false (BarcodeObject.cs:318).
	asBitmap bool
}

// BarcodeHorzAlign is the horizontal alignment of the barcode within its bounds.
// C# BarcodeObject.Alignment enum (BarcodeObject.cs:60).
type BarcodeHorzAlign int

const (
	// BarcodeHorzAlignLeft aligns the barcode to the left edge (default).
	BarcodeHorzAlignLeft BarcodeHorzAlign = iota
	// BarcodeHorzAlignCenter centers the barcode horizontally.
	BarcodeHorzAlignCenter
	// BarcodeHorzAlignRight aligns the barcode to the right edge.
	BarcodeHorzAlignRight
)

// barcodeHorzAlignStr converts a BarcodeHorzAlign to its FRX enum string.
func barcodeHorzAlignStr(a BarcodeHorzAlign) string {
	switch a {
	case BarcodeHorzAlignCenter:
		return "Center"
	case BarcodeHorzAlignRight:
		return "Right"
	default:
		return "Left"
	}
}

// parseBarcodeHorzAlign converts an FRX enum string to BarcodeHorzAlign.
func parseBarcodeHorzAlign(s string) BarcodeHorzAlign {
	switch s {
	case "Center":
		return BarcodeHorzAlignCenter
	case "Right":
		return BarcodeHorzAlignRight
	default:
		return BarcodeHorzAlignLeft
	}
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
// C# BarcodeObject constructor: Barcode = new Barcode39() (BarcodeObject.cs:688).
func NewBarcodeObject() *BarcodeObject {
	b := &BarcodeObject{
		ReportComponentBase: *report.NewReportComponentBase(),
		Barcode:             NewCode39Barcode(),
		autoSize:            true,
		showText:            true,
		showMarker:          true,  // C# BarcodeObject.cs:695 ShowMarker = true
		hideIfNoData:        true,  // C# BarcodeObject.cs:696 HideIfNoData = true
		zoom:                1.0,
		allowExpressions:    false, // C# BarcodeObject.cs:232 [DefaultValue(false)]
		brackets:            "[,]",
		trim:                true, // C# LinearBarcodeBase default
	}
	b.text = b.Barcode.DefaultValue() // C# BarcodeObject.cs:697 Text = Barcode.GetDefaultValue()
	return b
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

// Trim returns whether whitespace is trimmed from barcode text before encoding.
func (b *BarcodeObject) Trim() bool { return b.trim }

// SetTrim sets the trim flag.
func (b *BarcodeObject) SetTrim(v bool) { b.trim = v }

// HorzAlign returns the horizontal alignment of the barcode within its bounds.
func (b *BarcodeObject) HorzAlign() BarcodeHorzAlign { return b.horzAlign }

// SetHorzAlign sets the horizontal alignment.
func (b *BarcodeObject) SetHorzAlign(a BarcodeHorzAlign) { b.horzAlign = a }

// ShowMarker returns whether finder-pattern markers are shown.
func (b *BarcodeObject) ShowMarker() bool { return b.showMarker }

// SetShowMarker sets the show-marker flag.
func (b *BarcodeObject) SetShowMarker(v bool) { b.showMarker = v }

// AsBitmap returns whether the barcode is forced to render as a bitmap.
func (b *BarcodeObject) AsBitmap() bool { return b.asBitmap }

// SetAsBitmap sets the as-bitmap flag.
func (b *BarcodeObject) SetAsBitmap(v bool) { b.asBitmap = v }

// AllowExpressions returns whether expressions are evaluated in text.
func (b *BarcodeObject) AllowExpressions() bool { return b.allowExpressions }

// SetAllowExpressions sets allow-expressions.
func (b *BarcodeObject) SetAllowExpressions(v bool) { b.allowExpressions = v }

// Brackets returns the expression delimiter.
func (b *BarcodeObject) Brackets() string { return b.brackets }

// SetBrackets sets the delimiter.
func (b *BarcodeObject) SetBrackets(s string) { b.brackets = s }

// GetExpressions returns all expression strings from this BarcodeObject for
// dependency tracking and pre-compilation by the report engine.
// Mirrors C# BarcodeObject.GetExpressions() (BarcodeObject.cs:557–576):
//  1. Base expressions: Hyperlink.Expression and Bookmark from ReportComponentBase.
//  2. DataColumn if non-empty.
//  3. Expression if non-empty; otherwise, when AllowExpressions is true and
//     Brackets is non-empty, extract every bracket-delimited expression from Text.
func (b *BarcodeObject) GetExpressions() []string {
	var expressions []string

	// 1. Base: collect Hyperlink.Expression and Bookmark (C# ReportComponentBase.GetExpressions).
	if h := b.Hyperlink(); h != nil && h.Expression != "" {
		expressions = append(expressions, h.Expression)
	}
	if bk := b.Bookmark(); bk != "" {
		expressions = append(expressions, bk)
	}

	// 2. DataColumn takes priority for data binding.
	if b.dataColumn != "" {
		expressions = append(expressions, b.dataColumn)
	}

	// 3. Expression or bracket-extracted expressions from Text.
	// C# BarcodeObject.GetExpressions():
	//   if Expression != "" → add Expression
	//   else if AllowExpressions && Brackets != "" → GetExpressions(Text, open, close)
	if b.expression != "" {
		expressions = append(expressions, b.expression)
	} else if b.allowExpressions && b.brackets != "" {
		// Brackets is stored as "open,close" e.g. "[,]".
		// Split on the first comma only to support multi-char bracket sequences.
		parts := strings.SplitN(b.brackets, ",", 2)
		if len(parts) == 2 {
			open, close := parts[0], parts[1]
			for _, tok := range expr.ParseWithBrackets(b.text, open, close) {
				if tok.IsExpr {
					expressions = append(expressions, tok.Value)
				}
			}
		}
	}

	return expressions
}

// CreateSwissQR initializes the object as a QR barcode carrying a Swiss QR payload.
// Mirrors the C# BarcodeObject.CreateSwissQR convenience method.
func (b *BarcodeObject) CreateSwissQR(params SwissQRParameters) {
	swiss := NewSwissQRBarcode()
	swiss.Params = params

	b.Barcode = NewQRBarcode()
	b.text = swiss.FormatPayload()
	b.showText = false
}

// UpdateAutoSize resizes the BarcodeObject to fit the barcode's natural dimensions,
// then applies horizontal alignment via RelocateAlign().
// Mirrors C# BarcodeObject.UpdateAutoSize() (BarcodeObject.cs:390-412) and
// C# LinearBarcodeBase.CalcBounds() ShowText extra-padding logic (LinearBarcodeBase.cs:435-452).
// Called after Encode() so that CalcBounds() can use the encoded data.
func (b *BarcodeObject) UpdateAutoSize() {
	if b.Barcode == nil {
		return
	}
	// Propagate ShowText flag to the barcode implementation so 2D CalcBounds
	// can include font height. C# Barcode2DBase.CalcBounds() reads showText
	// and adds FontHeight when true.
	type showTextSetter interface{ SetShowText(bool) }
	if sts, ok := b.Barcode.(showTextSetter); ok {
		sts.SetShowText(b.showText)
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
			// Floor the result to match C# GDI+ MeasureString which returns
			// precise per-glyph measurements; our average-based approximation
			// tends to overshoot by a fractional amount.
			const arialAvgWidthFactor = 0.542
			fontPx := float32(8.0) * 96.0 / 72.0 // barcode default: Arial 8pt
			txtWidth := float32(math.Floor(float64(float32(len(displayText)) * fontPx * arialAvgWidthFactor)))
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

	// Save original bounds for RelocateAlign().
	// C# BarcodeObject.DrawBarcode(): origRect set before UpdateAutoSize() call.
	origLeft := b.Left()
	origWidth := b.Width()

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

	// RelocateAlign: shift Left to honour HorzAlign within original bounds.
	// C# BarcodeObject.RelocateAlign() (BarcodeObject.cs:417-435).
	switch b.horzAlign {
	case BarcodeHorzAlignCenter:
		b.SetLeft(origLeft + (origWidth-b.Width())/2)
	case BarcodeHorzAlignRight:
		b.SetLeft(origLeft + origWidth - b.Width())
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
	if !b.hideIfNoData {
		w.WriteBool("HideIfNoData", false)
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
	// Padding: C# BarcodeObject.cs:533.
	if b.padding.Left != 0 || b.padding.Top != 0 || b.padding.Right != 0 || b.padding.Bottom != 0 {
		w.WriteFloat("Padding.Left", b.padding.Left)
		w.WriteFloat("Padding.Top", b.padding.Top)
		w.WriteFloat("Padding.Right", b.padding.Right)
		w.WriteFloat("Padding.Bottom", b.padding.Bottom)
	}
	// HorzAlign: default Left (BarcodeObject.cs:119).
	if b.horzAlign != BarcodeHorzAlignLeft {
		w.WriteStr("HorzAlign", barcodeHorzAlignStr(b.horzAlign))
	}
	// ShowMarker: default true (BarcodeObject.cs:695).
	if !b.showMarker {
		w.WriteBool("ShowMarker", false)
	}
	// AsBitmap: default false (BarcodeObject.cs:318).
	if b.asBitmap {
		w.WriteBool("AsBitmap", true)
	}
	// Barcode-type-specific properties.
	if b.Barcode != nil {
		switch bc := b.Barcode.(type) {
		case *IntelligentMailBarcode:
			if bc.QuietZone {
				w.WriteBool("Barcode.QuietZone", true)
			}
		case *ITF14Barcode:
			// DrawVerticalBearerBars: default true; only write when false.
			// C# BarcodeITF14.Serialize (Barcode2of5.cs:405-412).
			if !bc.DrawVerticalBearerBars {
				w.WriteBool("Barcode.DrawVerticalBearerBars", false)
			}
		case *MaxiCodeBarcode:
			// Mode: MaxiCode encoding mode 2-6; default 4.
			// C# BarcodeMaxiCode.Serialize (BarcodeMaxiCode.cs:127).
			if bc.Mode != 4 {
				w.WriteInt("Barcode.Mode", bc.Mode)
			}
		}
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
		b.text = b.Barcode.DefaultValue() // reset text default to new type's default
	} else if name := r.ReadStr("Barcode", ""); name != "" {
		b.Barcode = NewBarcodeByName(name)
		b.text = b.Barcode.DefaultValue() // reset text default to new type's default
	}
	// Read barcode-specific properties (prefixed with "Barcode.").
	if b.Barcode != nil {
		switch bc := b.Barcode.(type) {
		case *QRBarcode:
			// ErrorCorrection: level string ("L"/"M"/"Q"/"H"), default "M".
			if ec := r.ReadStr("Barcode.ErrorCorrection", ""); ec != "" {
				bc.ErrorCorrection = ec
			}
			// QuietZone default is true per C# BarcodeQR.cs:902.
			bc.QuietZone = r.ReadBool("Barcode.QuietZone", true)
			// Encoding: QR character set (UTF8, ISO8859_1, …), default "UTF8".
			// C# QRCodeEncoding enum (BarcodeQR.cs:153).
			if enc := r.ReadStr("Barcode.Encoding", ""); enc != "" {
				bc.Encoding = enc
			}
			// ShowMarker: QR-specific marker display, default false.
			bc.ShowMarker = r.ReadBool("Barcode.ShowMarker", false)
			// Shape: module shape ("Rectangle"/"Circle"), default "Rectangle".
			if sh := r.ReadStr("Barcode.Shape", ""); sh != "" {
				bc.Shape = sh
			}
			bc.UseThinModules = r.ReadBool("Barcode.UseThinModules", false)
			// Angle: rotation for rotational shapes (Hexagon, Star, Snowflake), default 0.
			// C# BarcodeQR.Angle (BarcodeQR.cs:198).
			bc.Angle = r.ReadInt("Barcode.Angle", 0)
		case *AztecBarcode:
			// ErrorCorrection for Aztec is an integer percentage (default 33).
			// C# BarcodeAztec.ErrorCorrectionPercent serialised as prefix+"ErrorCorrection"
			// using WriteInt (BarcodeAztec.cs:86).
			bc.MinECCPercent = r.ReadInt("Barcode.ErrorCorrection", 33)
		case *PDF417Barcode:
			// C# BarcodePDF417 properties (BarcodePDF417.cs:1478-1496).
			bc.AspectRatio = r.ReadFloat("Barcode.AspectRatio", 0.5)
			bc.Columns = r.ReadInt("Barcode.Columns", 0)
			bc.Rows = r.ReadInt("Barcode.Rows", 0)
			bc.CodePage = r.ReadInt("Barcode.CodePage", 437)
			if cm := r.ReadStr("Barcode.CompactionMode", ""); cm != "" {
				bc.CompactionMode = cm
			}
			if ec := r.ReadStr("Barcode.ErrorCorrection", ""); ec != "" {
				bc.ErrorCorrection = ec
			}
			// PixelSize is serialised as "Width, Height" by C# WriteValue(Size).
			// We read Width and Height separately using common FRX patterns.
			// C# default: Size{2,8} (BarcodePDF417.cs:1551).
			bc.PixelSizeWidth = r.ReadInt("Barcode.PixelSize.Width", 2)
			bc.PixelSizeHeight = r.ReadInt("Barcode.PixelSize.Height", 8)
		case *DataMatrixBarcode:
			// C# BarcodeDatamatrix properties (BarcodeDatamatrix.cs:1060-1073).
			if ss := r.ReadStr("Barcode.SymbolSize", ""); ss != "" {
				bc.SymbolSize = ss
			}
			if enc := r.ReadStr("Barcode.Encoding", ""); enc != "" {
				bc.Encoding = enc
			}
			bc.CodePage = r.ReadInt("Barcode.CodePage", 1252)
			bc.PixelSize = r.ReadInt("Barcode.PixelSize", 3)
			bc.AutoEncode = r.ReadBool("Barcode.AutoEncode", true)
		case *MaxiCodeBarcode:
			// Mode: MaxiCode encoding mode 2-6, default 4.
			// C# BarcodeMaxiCode.cs:43: Mode = 4; serialised as WriteInt (line 127).
			bc.Mode = r.ReadInt("Barcode.Mode", 4)
		case *ITF14Barcode:
			// DrawVerticalBearerBars: default true.
			// C# BarcodeITF14.drawVerticalBearerBars=true (Barcode2of5.cs:332).
			bc.DrawVerticalBearerBars = r.ReadBool("Barcode.DrawVerticalBearerBars", true)
		case *DeutscheIdentcodeBarcode:
			// PrintCheckSum: controls whether the check digit appears in display text.
			// C# BarcodeDeutscheIdentcode serialises this as Barcode.DrawVerticalBearerBars
			// (a naming quirk in the C# codebase, Barcode2of5.cs:183).
			// Default true per C# constructor (Barcode2of5.cs:194).
			bc.PrintCheckSum = r.ReadBool("Barcode.DrawVerticalBearerBars", true)
		case *DeutscheLeitcodeBarcode:
			// PrintCheckSum: controls whether the check digit appears in display text.
			// C# BarcodeDeutscheLeitcode serialises this as Barcode.DrawVerticalBearerBars
			// (a naming quirk in the C# codebase, Barcode2of5.cs:244).
			// Default true per C# constructor (Barcode2of5.cs:316).
			bc.PrintCheckSum = r.ReadBool("Barcode.DrawVerticalBearerBars", true)
		case *CodabarBarcode:
			// StartChar/StopChar: C# BarcodeCodabar.cs:63,71 (default A/B).
			if s := r.ReadStr("Barcode.StartChar", ""); s != "" && len(s) == 1 {
				bc.StartChar = s[0]
			}
			if s := r.ReadStr("Barcode.StopChar", ""); s != "" && len(s) == 1 {
				bc.StopChar = s[0]
			}
		case *PharmacodeBarcode:
			bc.QuietZone = r.ReadBool("Barcode.QuietZone", true)
		case *IntelligentMailBarcode:
			bc.QuietZone = r.ReadBool("Barcode.QuietZone", false)
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
		// CalcCheckSum: C# LinearBarcodeBase default is true (LinearBarcodeBase.cs:637).
		// FRX writes the attribute only when it differs from the default.
		calcCS := r.ReadBool("Barcode.CalcCheckSum", true)
		type calcCheckSumSetter interface{ SetCalcCheckSum(bool) }
		if setter, ok := b.Barcode.(calcCheckSumSetter); ok {
			setter.SetCalcCheckSum(calcCS)
		}
		// AutoEncode: C# Barcode128.AutoEncode default is true (Barcode128.cs:591).
		type autoEncodeSetter interface{ SetAutoEncode(bool) }
		if setter, ok := b.Barcode.(autoEncodeSetter); ok {
			setter.SetAutoEncode(r.ReadBool("Barcode.AutoEncode", true))
		}
		// Barcode.Color: bar color for linear barcodes (default black).
		// C# LinearBarcodeBase.Color (LinearBarcodeBase.cs:69).
		// FRX attribute: Barcode.Color="Blue" or Barcode.Color="101, 67, 33".
		if cs := r.ReadStr("Barcode.Color", ""); cs != "" {
			type colorSetter interface{ SetBarcodeColor(color.RGBA) }
			if setter, ok := b.Barcode.(colorSetter); ok {
				if c, err := utils.ParseColor(cs); err == nil {
					setter.SetBarcodeColor(c)
				}
			}
		}
	}
	b.angle = r.ReadInt("Angle", 0)
	b.autoSize = r.ReadBool("AutoSize", true)
	b.dataColumn = r.ReadStr("DataColumn", "")
	b.expression = r.ReadStr("Expression", "")
	b.text = r.ReadStr("Text", b.text) // default: Barcode.GetDefaultValue() set in constructor
	b.showText = r.ReadBool("ShowText", true)
	b.zoom = r.ReadFloat("Zoom", 1.0)
	b.hideIfNoData = r.ReadBool("HideIfNoData", true) // C# BarcodeObject default true
	b.noDataText = r.ReadStr("NoDataText", "")
	b.allowExpressions = r.ReadBool("AllowExpressions", true)
	b.brackets = r.ReadStr("Brackets", "[,]")
	// Trim: C# LinearBarcodeBase default is true (LinearBarcodeBase.cs:638).
	b.trim = r.ReadBool("Barcode.Trim", true)
	// Padding: C# BarcodeObject.cs:533.
	b.padding.Left = r.ReadFloat("Padding.Left", 0)
	b.padding.Top = r.ReadFloat("Padding.Top", 0)
	b.padding.Right = r.ReadFloat("Padding.Right", 0)
	b.padding.Bottom = r.ReadFloat("Padding.Bottom", 0)
	// HorzAlign: default Left (BarcodeObject.cs:119).
	b.horzAlign = parseBarcodeHorzAlign(r.ReadStr("HorzAlign", "Left"))
	// ShowMarker: default true (BarcodeObject.cs:695).
	b.showMarker = r.ReadBool("ShowMarker", true)
	// AsBitmap: default false (BarcodeObject.cs:318).
	b.asBitmap = r.ReadBool("AsBitmap", false)
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

// Render renders the EAN-13 barcode using the native pattern-based renderer
// with custom text positioning: first digit outside left guard, then two
// groups of 6 digits under the left and right halves.
func (e *EAN13Barcode) Render(width, height int) (image.Image, error) {
	if e.encodedText == "" {
		return nil, fmt.Errorf("ean13: Encode must be called before Render")
	}
	pattern, err := e.GetPattern()
	if err != nil {
		return nil, err
	}
	return DrawLinearBarcodeCustomText(pattern, e.encodedText, width, height, true, e.GetWideBarRatio(), EAN13DrawText(pattern)), nil
}

// DefaultValue returns a sample EAN-13 value.
func (e *EAN13Barcode) DefaultValue() string { return "590123412345" }

// -----------------------------------------------------------------------
// AztecBarcode
// -----------------------------------------------------------------------

// AztecBarcode implements Aztec 2D symbology.
type AztecBarcode struct {
	BaseBarcodeImpl
	// MinECCPercent is the minimum error correction percentage.
	// C# BarcodeAztec.ErrorCorrectionPercent default is 33 (BarcodeAztec.cs:35).
	MinECCPercent int
	// UserSpecifiedLayers configures compact/full Aztec layers (0 = auto).
	UserSpecifiedLayers int
}

// NewAztecBarcode creates an AztecBarcode.
func NewAztecBarcode() *AztecBarcode {
	return &AztecBarcode{
		BaseBarcodeImpl:     newBaseBarcodeImpl(BarcodeTypeAztec),
		MinECCPercent:       33, // C# BarcodeAztec.cs:35 — default is 33
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
	// C# BarcodePDF417.Columns default is 0 (BarcodePDF417.cs:624).
	Columns int
	// Rows is the number of data rows (0 = auto).
	// C# BarcodePDF417.Rows default is 0 (BarcodePDF417.cs:639).
	Rows int
	// SecurityLevel is the error correction level 0-8 (default 2).
	SecurityLevel int
	// AspectRatio controls width:height ratio when Columns/Rows are 0.
	// C# BarcodePDF417.AspectRatio default is 0.5 (BarcodePDF417.cs:609).
	AspectRatio float32
	// CodePage is the code page for text conversion.
	// C# BarcodePDF417.CodePage default is 437 (BarcodePDF417.cs:663).
	CodePage int
	// CompactionMode controls encoding compaction: "Auto", "Text", "Numeric", "Binary".
	// C# PDF417CompactionMode enum, default Auto (BarcodePDF417.cs:673).
	CompactionMode string
	// ErrorCorrection is the error correction level: "Auto", "L0"–"L8".
	// C# PDF417ErrorCorrection enum, default Auto (BarcodePDF417.cs:649).
	ErrorCorrection string
	// PixelSizeWidth is the horizontal pixel size.
	// C# BarcodePDF417.PixelSize.Width default is 2 (BarcodePDF417.cs:1551).
	PixelSizeWidth int
	// PixelSizeHeight is the vertical pixel size.
	// C# BarcodePDF417.PixelSize.Height default is 8 (BarcodePDF417.cs:1551).
	PixelSizeHeight int
}

// NewPDF417Barcode creates a PDF417Barcode.
func NewPDF417Barcode() *PDF417Barcode {
	return &PDF417Barcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypePDF417),
		Columns:         0,
		Rows:            0,
		SecurityLevel:   2,
		AspectRatio:     0.5,
		CodePage:        437,
		CompactionMode:  "Auto",
		ErrorCorrection: "Auto",
		PixelSizeWidth:  2,
		PixelSizeHeight: 8,
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
	return drawLinearBarcodeColored(pattern, c.encodedText, width, height, true, c.GetWideBarRatio(), c.Color), nil
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
	Interleaved  bool
	CalcChecksum bool // C# LinearBarcodeBase.CalcCheckSum default true
}

// NewCode2of5Barcode creates a Code2of5Barcode with defaults.
// C# Barcode2of5Interleaved constructor: ratioMin=2, ratioMax=3 (Barcode2of5.cs:78-79).
func NewCode2of5Barcode() *Code2of5Barcode {
	b := newBaseBarcodeImpl(BarcodeTypeCode2of5)
	b.ratioMin = 2
	b.ratioMax = 3
	return &Code2of5Barcode{
		BaseBarcodeImpl: b,
		Interleaved:     true,
		CalcChecksum:    true,
	}
}

// SetCalcCheckSum implements calcCheckSumSetter for FRX deserialization.
func (c *Code2of5Barcode) SetCalcCheckSum(v bool) { c.CalcChecksum = v }

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
	return drawLinearBarcodeColored(pattern, c.encodedText, width, height, true, c.GetWideBarRatio(), c.Color), nil
}

// DefaultValue returns a sample 2-of-5 value.
func (c *Code2of5Barcode) DefaultValue() string { return "12345670" }

// -----------------------------------------------------------------------
// CodabarBarcode — Codabar linear symbology
// -----------------------------------------------------------------------

// CodabarBarcode implements Codabar symbology.
type CodabarBarcode struct {
	BaseBarcodeImpl
	// StartChar is the start character (A, B, C, or D).
	// C# BarcodeCodabar.cs:63 [DefaultValue(CodabarChar.A)].
	StartChar byte
	// StopChar is the stop character (A, B, C, or D).
	// C# BarcodeCodabar.cs:71 [DefaultValue(CodabarChar.B)].
	StopChar byte
}

// NewCodabarBarcode creates a CodabarBarcode.
// C# BarcodeCodabar constructor: ratioMin=2, ratioMax=3 (BarcodeCodabar.cs:141-142).
func NewCodabarBarcode() *CodabarBarcode {
	b := newBaseBarcodeImpl(BarcodeTypeCodabar)
	b.ratioMin = 2
	b.ratioMax = 3
	return &CodabarBarcode{
		BaseBarcodeImpl: b,
		StartChar:       'A', // C# default: CodabarChar.A
		StopChar:        'B', // C# default: CodabarChar.B
	}
}

// Assign copies all CodabarBarcode fields from src.
// Mirrors C# BarcodeCodabar.Assign (BarcodeCodabar.cs).
func (c *CodabarBarcode) Assign(src *CodabarBarcode) {
	if src == nil {
		return
	}
	c.BaseBarcodeImpl = src.BaseBarcodeImpl
	c.StartChar = src.StartChar
	c.StopChar = src.StopChar
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
	return drawLinearBarcodeColored(pattern, c.encodedText, width, height, true, c.GetWideBarRatio(), c.Color), nil
}

// DefaultValue returns the C# BarcodeBase.GetDefaultValue() default: "12345678".
func (c *CodabarBarcode) DefaultValue() string { return "12345678" }

// -----------------------------------------------------------------------
// DataMatrixBarcode — DataMatrix 2D symbology
// -----------------------------------------------------------------------

// DataMatrixBarcode implements DataMatrix 2D symbology.
type DataMatrixBarcode struct {
	BaseBarcodeImpl
	// SymbolSize is the DataMatrix symbol size: "Auto", "10x10", etc.
	// C# DatamatrixSymbolSize enum, default Auto (BarcodeDatamatrix.cs:326).
	SymbolSize string
	// Encoding is the DataMatrix encoding mode: "Auto", "Ascii", "C40", etc.
	// C# DatamatrixEncoding enum, default Auto (BarcodeDatamatrix.cs:336).
	Encoding string
	// CodePage is the code page for text conversion.
	// C# BarcodeDatamatrix.CodePage default is 1252 (BarcodeDatamatrix.cs:350).
	CodePage int
	// PixelSize is the pixel size for rendering.
	// C# BarcodeDatamatrix.PixelSize default is 3 (BarcodeDatamatrix.cs:360).
	PixelSize int
	// AutoEncode enables automatic encoding mode selection.
	// C# BarcodeDatamatrix.AutoEncode default is true (BarcodeDatamatrix.cs:370).
	AutoEncode bool
}

// NewDataMatrixBarcode creates a DataMatrixBarcode.
func NewDataMatrixBarcode() *DataMatrixBarcode {
	return &DataMatrixBarcode{
		BaseBarcodeImpl: newBaseBarcodeImpl(BarcodeTypeDataMatrix),
		SymbolSize:      "Auto",
		Encoding:        "Auto",
		CodePage:        1252,
		PixelSize:       3,
		AutoEncode:      true,
	}
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
		return NewGS1_128Barcode()
	case BarcodeTypeCode39:
		return NewCode39Barcode()
	case BarcodeTypeCode39Extended:
		return NewCode39ExtendedBarcode()
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
