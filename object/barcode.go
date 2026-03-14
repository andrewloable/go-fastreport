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

// ZipCodeObject renders a US POSTNET zip code barcode.
// It is the Go equivalent of FastReport.ZipCodeObject.
type ZipCodeObject struct {
	report.ReportComponentBase

	// text is the zip code string to render.
	text string
	// allowExpressions enables bracket-expression evaluation in text.
	allowExpressions bool
}

// NewZipCodeObject creates a ZipCodeObject with defaults.
func NewZipCodeObject() *ZipCodeObject {
	return &ZipCodeObject{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// Text returns the zip code data string.
func (z *ZipCodeObject) Text() string { return z.text }

// SetText sets the zip code data string.
func (z *ZipCodeObject) SetText(v string) { z.text = v }

// AllowExpressions returns whether bracket expressions are evaluated.
func (z *ZipCodeObject) AllowExpressions() bool { return z.allowExpressions }

// SetAllowExpressions sets the AllowExpressions flag.
func (z *ZipCodeObject) SetAllowExpressions(v bool) { z.allowExpressions = v }

// Serialize writes ZipCodeObject properties that differ from defaults.
func (z *ZipCodeObject) Serialize(w report.Writer) error {
	if err := z.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if z.text != "" {
		w.WriteStr("Text", z.text)
	}
	if z.allowExpressions {
		w.WriteBool("AllowExpressions", true)
	}
	return nil
}

// Deserialize reads ZipCodeObject properties.
func (z *ZipCodeObject) Deserialize(r report.Reader) error {
	if err := z.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	z.text = r.ReadStr("Text", "")
	z.allowExpressions = r.ReadBool("AllowExpressions", false)
	return nil
}
