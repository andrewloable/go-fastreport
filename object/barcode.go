package object

import (
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
}

// NewBarcodeObject creates a BarcodeObject with defaults (ShowText=true, AutoSize=true).
func NewBarcodeObject() *BarcodeObject {
	return &BarcodeObject{
		ReportComponentBase: *report.NewReportComponentBase(),
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
	return nil
}

// Deserialize reads BarcodeObject properties.
func (b *BarcodeObject) Deserialize(r report.Reader) error {
	if err := b.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	b.text = r.ReadStr("Text", "")
	b.barcodeType = r.ReadStr("Barcode", "")
	b.showText = r.ReadBool("ShowText", true)
	b.autoSize = r.ReadBool("AutoSize", true)
	b.allowExpressions = r.ReadBool("AllowExpressions", false)
	return nil
}

// ── ZipCodeObject ─────────────────────────────────────────────────────────────

// ZipCodeObject renders a Russian postal (GOST R 51506-99) or US POSTNET zip code.
// It is the Go equivalent of FastReport.ZipCodeObject.
type ZipCodeObject struct {
	report.ReportComponentBase

	// text is the zip code value.
	text string
	// dataColumn is the data source column bound to this object.
	dataColumn string
	// expression is an expression that evaluates to the zip code value.
	expression string
	// segmentWidth is the width of a single digit segment in pixels.
	segmentWidth float32
	// segmentHeight is the height of a single digit segment in pixels.
	segmentHeight float32
	// spacing is the spacing between segment origins in pixels.
	spacing float32
	// segmentCount is the number of zip code digit segments (default 6).
	segmentCount int
	// showMarkers controls whether reference markers are drawn (default true).
	showMarkers bool
	// showGrid controls whether the digit grid is drawn (default true).
	showGrid bool
}

// NewZipCodeObject creates a ZipCodeObject with defaults matching FastReport .NET.
func NewZipCodeObject() *ZipCodeObject {
	return &ZipCodeObject{
		ReportComponentBase: *report.NewReportComponentBase(),
		segmentCount:        6,
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
func (z *ZipCodeObject) Serialize(w report.Writer) error {
	if err := z.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if z.text != "" {
		w.WriteStr("Text", z.text)
	}
	if z.dataColumn != "" {
		w.WriteStr("DataColumn", z.dataColumn)
	}
	if z.expression != "" {
		w.WriteStr("Expression", z.expression)
	}
	if z.segmentWidth != 0 {
		w.WriteFloat("SegmentWidth", z.segmentWidth)
	}
	if z.segmentHeight != 0 {
		w.WriteFloat("SegmentHeight", z.segmentHeight)
	}
	if z.spacing != 0 {
		w.WriteFloat("Spacing", z.spacing)
	}
	if z.segmentCount != 6 {
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

// Deserialize reads ZipCodeObject properties.
func (z *ZipCodeObject) Deserialize(r report.Reader) error {
	if err := z.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	z.text = r.ReadStr("Text", "")
	z.dataColumn = r.ReadStr("DataColumn", "")
	z.expression = r.ReadStr("Expression", "")
	z.segmentWidth = r.ReadFloat("SegmentWidth", 0)
	z.segmentHeight = r.ReadFloat("SegmentHeight", 0)
	z.spacing = r.ReadFloat("Spacing", 0)
	z.segmentCount = r.ReadInt("SegmentCount", 6)
	z.showMarkers = r.ReadBool("ShowMarkers", true)
	z.showGrid = r.ReadBool("ShowGrid", true)
	return nil
}
