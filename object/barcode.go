package object

import (
	"fmt"

	"github.com/andrewloable/go-fastreport/report"
)

// BarcodeObject is a report object that renders a barcode (1D or 2D).
// It is the Go equivalent of FastReport.BarcodeObject.
type BarcodeObject struct {
	report.ReportComponentBase

	// text is the data to encode in the barcode.
	text string
	// barcodeType is the barcode symbology name (e.g. "QR Code", "Code128").
	barcodeType string
	// showText controls whether the human-readable text is printed below.
	showText bool
	// autoSize resizes the object to fit the barcode content when true.
	autoSize bool
	// allowExpressions enables bracket-expression evaluation in text.
	allowExpressions bool
	// dataColumn is the data source column bound to this object.
	dataColumn string
	// expression is an expression that evaluates to the barcode value.
	expression string
	// savedText holds the pre-engine-pass text for SaveState/RestoreState.
	savedText string
}

// NewBarcodeObject creates a BarcodeObject with defaults (ShowText=true, AutoSize=true).
// Default Text is "12345678" (mirrors C# BarcodeObject constructor which calls
// Barcode.GetDefaultValue() on the default Barcode39 instance).
func NewBarcodeObject() *BarcodeObject {
	return &BarcodeObject{
		ReportComponentBase: *report.NewReportComponentBase(),
		text:                "12345678",
		showText:            true,
		autoSize:            true,
	}
}

// Text returns the barcode data string.
func (b *BarcodeObject) Text() string { return b.text }

// SetText sets the barcode data string.
func (b *BarcodeObject) SetText(v string) { b.text = v }

// BarcodeType returns the symbology name (e.g. "QR Code", "Code128").
func (b *BarcodeObject) BarcodeType() string { return b.barcodeType }

// SetBarcodeType sets the symbology name.
func (b *BarcodeObject) SetBarcodeType(v string) { b.barcodeType = v }

// ShowText returns whether the human-readable text is displayed.
func (b *BarcodeObject) ShowText() bool { return b.showText }

// SetShowText sets the ShowText flag.
func (b *BarcodeObject) SetShowText(v bool) { b.showText = v }

// AutoSize returns whether the object resizes to fit the barcode.
func (b *BarcodeObject) AutoSize() bool { return b.autoSize }

// SetAutoSize sets the AutoSize flag.
func (b *BarcodeObject) SetAutoSize(v bool) { b.autoSize = v }

// AllowExpressions returns whether bracket expressions are evaluated in Text.
func (b *BarcodeObject) AllowExpressions() bool { return b.allowExpressions }

// SetAllowExpressions sets the AllowExpressions flag.
func (b *BarcodeObject) SetAllowExpressions(v bool) { b.allowExpressions = v }

// DataColumn returns the data source column that provides the barcode value.
func (b *BarcodeObject) DataColumn() string { return b.dataColumn }

// SetDataColumn sets the data column binding.
func (b *BarcodeObject) SetDataColumn(s string) { b.dataColumn = s }

// Expression returns the expression that evaluates to the barcode value.
func (b *BarcodeObject) Expression() string { return b.expression }

// SetExpression sets the barcode value expression.
func (b *BarcodeObject) SetExpression(s string) { b.expression = s }

// Serialize writes BarcodeObject properties that differ from defaults.
func (b *BarcodeObject) Serialize(w report.Writer) error {
	if err := b.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if b.text != "" {
		w.WriteStr("Text", b.text)
	}
	if b.barcodeType != "" {
		w.WriteStr("Barcode", b.barcodeType)
	}
	if !b.showText {
		w.WriteBool("ShowText", false)
	}
	if !b.autoSize {
		w.WriteBool("AutoSize", false)
	}
	if b.allowExpressions {
		w.WriteBool("AllowExpressions", true)
	}
	if b.dataColumn != "" {
		w.WriteStr("DataColumn", b.dataColumn)
	}
	if b.expression != "" {
		w.WriteStr("Expression", b.expression)
	}
	return nil
}

// Deserialize reads BarcodeObject properties.
func (b *BarcodeObject) Deserialize(r report.Reader) error {
	if err := b.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	b.text = r.ReadStr("Text", b.text)
	b.barcodeType = r.ReadStr("Barcode", "")
	b.showText = r.ReadBool("ShowText", true)
	b.autoSize = r.ReadBool("AutoSize", true)
	b.allowExpressions = r.ReadBool("AllowExpressions", false)
	b.dataColumn = r.ReadStr("DataColumn", "")
	b.expression = r.ReadStr("Expression", "")
	return nil
}

// savedText holds the text saved by SaveState so RestoreState can undo
// expression evaluation. Stored as a separate unexported field because
// BarcodeObject has no embedding that provides this; mirrors C# BarcodeObject.savedText.

// GetExpressions returns the list of expressions used by this BarcodeObject
// for pre-compilation by the report engine.
// Mirrors C# BarcodeObject.GetExpressions (BarcodeObject.cs line 557-576).
func (b *BarcodeObject) GetExpressions() []string {
	exprs := b.ReportComponentBase.GetExpressions()
	if b.dataColumn != "" {
		exprs = append(exprs, b.dataColumn)
	}
	if b.expression != "" {
		exprs = append(exprs, b.expression)
	}
	return exprs
}

// SaveState saves the current Text so RestoreState can undo engine-pass changes.
// Mirrors C# BarcodeObject.SaveState (BarcodeObject.cs line 579-583).
func (b *BarcodeObject) SaveState() {
	b.ReportComponentBase.SaveState()
	b.savedText = b.text
}

// RestoreState restores the Text saved by SaveState.
// Mirrors C# BarcodeObject.RestoreState (BarcodeObject.cs line 586-590).
func (b *BarcodeObject) RestoreState() {
	b.ReportComponentBase.RestoreState()
	b.text = b.savedText
}

// GetData evaluates the DataColumn or Expression binding using the provided
// calc function and updates the Text value accordingly.
// Mirrors C# BarcodeObject.GetData / GetDataShared (BarcodeObject.cs line 593-627).
func (b *BarcodeObject) GetData(calc func(string) (any, error)) {
	if b.dataColumn != "" {
		val, err := calc("[" + b.dataColumn + "]")
		if err == nil && val != nil {
			b.text = toString(val)
		} else {
			b.text = ""
		}
	} else if b.expression != "" {
		val, err := calc(b.expression)
		if err == nil && val != nil {
			b.text = toString(val)
		} else {
			b.text = ""
		}
	}
}

// ── ZipCodeObject ─────────────────────────────────────────────────────────────

// ZipCodeObject default values matching FastReport .NET ZipCodeObject constructor.
// C# Units.Centimeters = 37.8 px, Units.Millimeters = 3.78 px (Utils/Units.cs).
// ZipCodeObject.cs: segmentWidth = Units.Centimeters * 0.5f = 18.9
//                   segmentHeight = Units.Centimeters * 1    = 37.8
//                   spacing       = Units.Centimeters * 0.9f = 34.02
//                   text          = "123456"
//                   Border.Width  = 3
const (
	zipDefaultSegmentWidth  float32 = 18.9
	zipDefaultSegmentHeight float32 = 37.8
	zipDefaultSpacing       float32 = 34.02
	zipDefaultText                  = "123456"
	zipDefaultSegmentCount          = 6
)

// ZipCodeObject renders a Russian postal (GOST R 51506-99) zip code barcode.
// It is the Go equivalent of FastReport.ZipCodeObject, which complies with
// GOST R 51506-99.
type ZipCodeObject struct {
	report.ReportComponentBase

	// text is the zip code value.
	text string
	// dataColumn is the data source column bound to this object.
	dataColumn string
	// expression is an expression that evaluates to the zip code value.
	expression string
	// segmentWidth is the width of a single digit segment in pixels.
	// Default: 18.9 px (Units.Centimeters * 0.5f).
	segmentWidth float32
	// segmentHeight is the height of a single digit segment in pixels.
	// Default: 37.8 px (Units.Centimeters * 1).
	segmentHeight float32
	// spacing is the spacing between segment origins in pixels.
	// Default: 34.02 px (Units.Centimeters * 0.9f).
	spacing float32
	// segmentCount is the number of zip code digit segments (default 6).
	segmentCount int
	// showMarkers controls whether reference markers are drawn (default true).
	showMarkers bool
	// showGrid controls whether the digit grid is drawn (default true).
	showGrid bool
}

// NewZipCodeObject creates a ZipCodeObject with defaults matching FastReport .NET.
// C# ZipCodeObject constructor (ZipCodeObject.cs line 362-378):
//
//	segmentWidth  = Units.Centimeters * 0.5f  // 18.9 px
//	segmentHeight = Units.Centimeters * 1     // 37.8 px
//	spacing       = Units.Centimeters * 0.9f  // 34.02 px
//	segmentCount  = 6
//	showMarkers   = true
//	showGrid      = true
//	text          = "123456"
func NewZipCodeObject() *ZipCodeObject {
	return &ZipCodeObject{
		ReportComponentBase: *report.NewReportComponentBase(),
		text:                zipDefaultText,
		segmentWidth:        zipDefaultSegmentWidth,
		segmentHeight:       zipDefaultSegmentHeight,
		spacing:             zipDefaultSpacing,
		segmentCount:        zipDefaultSegmentCount,
		showMarkers:         true,
		showGrid:            true,
	}
}

// Text returns the zip code string.
func (z *ZipCodeObject) Text() string { return z.text }

// SetText sets the zip code string.
func (z *ZipCodeObject) SetText(v string) { z.text = v }

// DataColumn returns the data source column name.
func (z *ZipCodeObject) DataColumn() string { return z.dataColumn }

// SetDataColumn sets the data source column name.
func (z *ZipCodeObject) SetDataColumn(v string) { z.dataColumn = v }

// Expression returns the zip code expression.
func (z *ZipCodeObject) Expression() string { return z.expression }

// SetExpression sets the zip code expression.
func (z *ZipCodeObject) SetExpression(v string) { z.expression = v }

// SegmentWidth returns the width of one digit segment in pixels.
func (z *ZipCodeObject) SegmentWidth() float32 { return z.segmentWidth }

// SetSegmentWidth sets the segment width.
func (z *ZipCodeObject) SetSegmentWidth(v float32) { z.segmentWidth = v }

// SegmentHeight returns the height of one digit segment in pixels.
func (z *ZipCodeObject) SegmentHeight() float32 { return z.segmentHeight }

// SetSegmentHeight sets the segment height.
func (z *ZipCodeObject) SetSegmentHeight(v float32) { z.segmentHeight = v }

// Spacing returns the spacing between digit segment origins in pixels.
func (z *ZipCodeObject) Spacing() float32 { return z.spacing }

// SetSpacing sets the spacing.
func (z *ZipCodeObject) SetSpacing(v float32) { z.spacing = v }

// SegmentCount returns the number of digit segments (default 6).
func (z *ZipCodeObject) SegmentCount() int { return z.segmentCount }

// SetSegmentCount sets the number of digit segments.
func (z *ZipCodeObject) SetSegmentCount(v int) { z.segmentCount = v }

// ShowMarkers returns whether reference markers are drawn.
func (z *ZipCodeObject) ShowMarkers() bool { return z.showMarkers }

// SetShowMarkers sets the ShowMarkers flag.
func (z *ZipCodeObject) SetShowMarkers(v bool) { z.showMarkers = v }

// ShowGrid returns whether the digit grid is drawn.
func (z *ZipCodeObject) ShowGrid() bool { return z.showGrid }

// SetShowGrid sets the ShowGrid flag.
func (z *ZipCodeObject) SetShowGrid(v bool) { z.showGrid = v }

// Serialize writes ZipCodeObject properties that differ from defaults.
// Mirrors C# ZipCodeObject.Serialize (ZipCodeObject.cs line 295-320):
// only writes attributes that differ from the default instance.
func (z *ZipCodeObject) Serialize(w report.Writer) error {
	if err := z.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if z.text != zipDefaultText {
		w.WriteStr("Text", z.text)
	}
	if z.dataColumn != "" {
		w.WriteStr("DataColumn", z.dataColumn)
	}
	if z.expression != "" {
		w.WriteStr("Expression", z.expression)
	}
	if z.segmentWidth != zipDefaultSegmentWidth {
		w.WriteFloat("SegmentWidth", z.segmentWidth)
	}
	if z.segmentHeight != zipDefaultSegmentHeight {
		w.WriteFloat("SegmentHeight", z.segmentHeight)
	}
	if z.spacing != zipDefaultSpacing {
		w.WriteFloat("Spacing", z.spacing)
	}
	if z.segmentCount != zipDefaultSegmentCount {
		w.WriteInt("SegmentCount", z.segmentCount)
	}
	if !z.showMarkers {
		w.WriteBool("ShowMarkers", false)
	}
	if !z.showGrid {
		w.WriteBool("ShowGrid", false)
	}
	return nil
}

// Deserialize reads ZipCodeObject properties, using C# defaults as fallback
// when attributes are absent from the FRX file.
func (z *ZipCodeObject) Deserialize(r report.Reader) error {
	if err := z.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	z.text = r.ReadStr("Text", zipDefaultText)
	z.dataColumn = r.ReadStr("DataColumn", "")
	z.expression = r.ReadStr("Expression", "")
	z.segmentWidth = r.ReadFloat("SegmentWidth", zipDefaultSegmentWidth)
	z.segmentHeight = r.ReadFloat("SegmentHeight", zipDefaultSegmentHeight)
	z.spacing = r.ReadFloat("Spacing", zipDefaultSpacing)
	z.segmentCount = r.ReadInt("SegmentCount", zipDefaultSegmentCount)
	z.showMarkers = r.ReadBool("ShowMarkers", true)
	z.showGrid = r.ReadBool("ShowGrid", true)
	return nil
}

// Assign copies all ZipCodeObject properties from src.
// Mirrors C# ZipCodeObject.Assign (ZipCodeObject.cs:247-263).
func (z *ZipCodeObject) Assign(src *ZipCodeObject) {
	if src == nil {
		return
	}
	z.ReportComponentBase.Assign(&src.ReportComponentBase)
	z.segmentWidth = src.segmentWidth
	z.segmentHeight = src.segmentHeight
	z.spacing = src.spacing
	z.segmentCount = src.segmentCount
	z.showMarkers = src.showMarkers
	z.showGrid = src.showGrid
	z.dataColumn = src.dataColumn
	z.expression = src.expression
	z.text = src.text
}

// GetExpressions returns the list of expressions that need to be evaluated
// by the report engine. Mirrors C# ZipCodeObject.GetExpressions()
// (ZipCodeObject.cs line 325-335).
func (z *ZipCodeObject) GetExpressions() []string {
	exprs := z.ReportComponentBase.GetExpressions()
	if z.dataColumn != "" {
		exprs = append(exprs, z.dataColumn)
	}
	if z.expression != "" {
		exprs = append(exprs, z.expression)
	}
	return exprs
}

// GetData evaluates the DataColumn or Expression binding using the provided
// calc function and updates the Text value accordingly.
// Mirrors C# ZipCodeObject.GetData / GetDataShared (ZipCodeObject.cs line 338-356).
func (z *ZipCodeObject) GetData(calc func(string) (any, error)) {
	if z.dataColumn != "" {
		val, err := calc("[" + z.dataColumn + "]")
		if err == nil && val != nil {
			z.text = toString(val)
		} else {
			z.text = ""
		}
	} else if z.expression != "" {
		val, err := calc(z.expression)
		if err == nil && val != nil {
			z.text = toString(val)
		} else {
			z.text = ""
		}
	}
}

// toString converts an arbitrary value to its string representation.
func toString(v any) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		return fmt.Sprintf("%v", v)
	}
}
