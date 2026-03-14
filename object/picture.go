package object

import (
	"github.com/andrewloable/go-fastreport/report"
)

// SizeMode controls how an image is scaled within a PictureObject.
// It mirrors System.Windows.Forms.PictureBoxSizeMode.
type SizeMode int

const (
	// SizeModeNormal displays the image at its original size, clipped if needed.
	SizeModeNormal SizeMode = iota
	// SizeModeStretchImage stretches the image to fill the object bounds.
	SizeModeStretchImage
	// SizeModeAutoSize resizes the object to fit the image.
	SizeModeAutoSize
	// SizeModeCenterImage centres the image without scaling.
	SizeModeCenterImage
	// SizeModeZoom scales the image proportionally to fit within bounds (default).
	SizeModeZoom
)

// ImageAlign controls fixed-position alignment of the image within the object.
type ImageAlign int

const (
	ImageAlignNone         ImageAlign = iota
	ImageAlignTopLeft
	ImageAlignTopCenter
	ImageAlignTopRight
	ImageAlignCenterLeft
	ImageAlignCenterCenter
	ImageAlignCenterRight
	ImageAlignBottomLeft
	ImageAlignBottomCenter
	ImageAlignBottomRight
)

// PictureObjectBase is the base for image-displaying report objects.
// It is the Go equivalent of FastReport.PictureObjectBase.
type PictureObjectBase struct {
	report.ReportComponentBase

	angle               int
	dataColumn          string
	grayscale           bool
	imageLocation       string
	imageSourceExpr     string
	maxHeight           float32
	maxWidth            float32
	padding             Padding
	sizeMode            SizeMode // default SizeModeZoom
	imageAlign          ImageAlign
	showErrorImage      bool
}

// NewPictureObjectBase creates a PictureObjectBase with defaults.
func NewPictureObjectBase() *PictureObjectBase {
	return &PictureObjectBase{
		ReportComponentBase: *report.NewReportComponentBase(),
		sizeMode:            SizeModeZoom,
	}
}

// Angle returns the rotation angle in degrees.
func (p *PictureObjectBase) Angle() int { return p.angle }

// SetAngle sets the rotation angle.
func (p *PictureObjectBase) SetAngle(a int) { p.angle = a }

// DataColumn returns the data source column that provides the image bytes.
func (p *PictureObjectBase) DataColumn() string { return p.dataColumn }

// SetDataColumn sets the data column binding.
func (p *PictureObjectBase) SetDataColumn(s string) { p.dataColumn = s }

// Grayscale returns whether the image is rendered in grayscale.
func (p *PictureObjectBase) Grayscale() bool { return p.grayscale }

// SetGrayscale sets grayscale rendering.
func (p *PictureObjectBase) SetGrayscale(v bool) { p.grayscale = v }

// ImageLocation returns a URL or file path for the image.
func (p *PictureObjectBase) ImageLocation() string { return p.imageLocation }

// SetImageLocation sets the image URL or file path.
func (p *PictureObjectBase) SetImageLocation(s string) { p.imageLocation = s }

// ImageSourceExpression returns an expression that evaluates to an image path/URL.
func (p *PictureObjectBase) ImageSourceExpression() string { return p.imageSourceExpr }

// SetImageSourceExpression sets the image-source expression.
func (p *PictureObjectBase) SetImageSourceExpression(s string) { p.imageSourceExpr = s }

// MaxHeight returns the maximum display height in pixels (0 = no limit).
func (p *PictureObjectBase) MaxHeight() float32 { return p.maxHeight }

// SetMaxHeight sets the maximum height.
func (p *PictureObjectBase) SetMaxHeight(v float32) { p.maxHeight = v }

// MaxWidth returns the maximum display width in pixels (0 = no limit).
func (p *PictureObjectBase) MaxWidth() float32 { return p.maxWidth }

// SetMaxWidth sets the maximum width.
func (p *PictureObjectBase) SetMaxWidth(v float32) { p.maxWidth = v }

// Padding returns interior padding around the image.
func (p *PictureObjectBase) Padding() Padding { return p.padding }

// SetPadding sets the interior padding.
func (p *PictureObjectBase) SetPadding(pd Padding) { p.padding = pd }

// SizeMode returns how the image is scaled.
func (p *PictureObjectBase) SizeMode() SizeMode { return p.sizeMode }

// SetSizeMode sets the size mode.
func (p *PictureObjectBase) SetSizeMode(m SizeMode) { p.sizeMode = m }

// ImageAlign returns the fixed-position alignment of the image.
func (p *PictureObjectBase) ImageAlign() ImageAlign { return p.imageAlign }

// SetImageAlign sets the image alignment.
func (p *PictureObjectBase) SetImageAlign(a ImageAlign) { p.imageAlign = a }

// ShowErrorImage returns whether a placeholder is shown when the image fails to load.
func (p *PictureObjectBase) ShowErrorImage() bool { return p.showErrorImage }

// SetShowErrorImage sets the show-error-image flag.
func (p *PictureObjectBase) SetShowErrorImage(v bool) { p.showErrorImage = v }

// Serialize writes PictureObjectBase properties that differ from defaults.
func (p *PictureObjectBase) Serialize(w report.Writer) error {
	if err := p.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if p.angle != 0 {
		w.WriteInt("Angle", p.angle)
	}
	if p.dataColumn != "" {
		w.WriteStr("DataColumn", p.dataColumn)
	}
	if p.grayscale {
		w.WriteBool("Grayscale", true)
	}
	if p.imageLocation != "" {
		w.WriteStr("ImageLocation", p.imageLocation)
	}
	if p.imageSourceExpr != "" {
		w.WriteStr("ImageSourceExpression", p.imageSourceExpr)
	}
	if p.maxHeight != 0 {
		w.WriteFloat("MaxHeight", p.maxHeight)
	}
	if p.maxWidth != 0 {
		w.WriteFloat("MaxWidth", p.maxWidth)
	}
	if p.padding != (Padding{}) {
		w.WriteStr("Padding", paddingToStr(p.padding))
	}
	if p.sizeMode != SizeModeZoom {
		w.WriteInt("SizeMode", int(p.sizeMode))
	}
	if p.imageAlign != ImageAlignNone {
		w.WriteInt("ImageAlign", int(p.imageAlign))
	}
	if p.showErrorImage {
		w.WriteBool("ShowErrorImage", true)
	}
	return nil
}

// Deserialize reads PictureObjectBase properties.
func (p *PictureObjectBase) Deserialize(r report.Reader) error {
	if err := p.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	p.angle = r.ReadInt("Angle", 0)
	p.dataColumn = r.ReadStr("DataColumn", "")
	p.grayscale = r.ReadBool("Grayscale", false)
	p.imageLocation = r.ReadStr("ImageLocation", "")
	p.imageSourceExpr = r.ReadStr("ImageSourceExpression", "")
	p.maxHeight = r.ReadFloat("MaxHeight", 0)
	p.maxWidth = r.ReadFloat("MaxWidth", 0)
	if s := r.ReadStr("Padding", ""); s != "" {
		p.padding = strToPadding(s)
	}
	p.sizeMode = SizeMode(r.ReadInt("SizeMode", int(SizeModeZoom)))
	p.imageAlign = ImageAlign(r.ReadInt("ImageAlign", 0))
	p.showErrorImage = r.ReadBool("ShowErrorImage", false)
	return nil
}

// -----------------------------------------------------------------------
// PictureObject
// -----------------------------------------------------------------------

// ImageFormat identifies the encoded format of image bytes.
type ImageFormat int

const (
	ImageFormatUnknown ImageFormat = iota
	ImageFormatPng
	ImageFormatJpeg
	ImageFormatGif
	ImageFormatBmp
	ImageFormatTiff
	ImageFormatSvg
)

// PictureObject displays a raster or vector image.
// It is the Go equivalent of FastReport.PictureObject.
type PictureObject struct {
	PictureObjectBase

	// imageData holds the raw encoded image bytes.
	imageData []byte
	// imageFormat is the format of imageData.
	imageFormat ImageFormat
	// tile tiles the image across the object bounds.
	tile bool
	// transparency is the global alpha (0.0 = opaque, 1.0 = invisible).
	transparency float32
}

// NewPictureObject creates a PictureObject with defaults.
func NewPictureObject() *PictureObject {
	return &PictureObject{
		PictureObjectBase: *NewPictureObjectBase(),
	}
}

// ImageData returns the raw encoded image bytes (nil if none loaded).
func (p *PictureObject) ImageData() []byte { return p.imageData }

// SetImageData sets the raw image bytes and resets the format to Unknown.
func (p *PictureObject) SetImageData(data []byte) {
	p.imageData = data
	p.imageFormat = ImageFormatUnknown
}

// SetImageDataWithFormat sets raw image bytes and their known format.
func (p *PictureObject) SetImageDataWithFormat(data []byte, fmt ImageFormat) {
	p.imageData = data
	p.imageFormat = fmt
}

// ImageFormat returns the format of the stored image bytes.
func (p *PictureObject) ImageFormat() ImageFormat { return p.imageFormat }

// Tile returns whether the image is tiled across the object bounds.
func (p *PictureObject) Tile() bool { return p.tile }

// SetTile sets the tile flag.
func (p *PictureObject) SetTile(v bool) { p.tile = v }

// Transparency returns the global transparency (0.0 = fully opaque).
func (p *PictureObject) Transparency() float32 { return p.transparency }

// SetTransparency sets the transparency (0.0–1.0).
func (p *PictureObject) SetTransparency(v float32) { p.transparency = v }

// HasImage reports whether any image data or location is set.
func (p *PictureObject) HasImage() bool {
	return len(p.imageData) > 0 || p.imageLocation != "" || p.dataColumn != ""
}

// Serialize writes PictureObject properties that differ from defaults.
func (p *PictureObject) Serialize(w report.Writer) error {
	if err := p.PictureObjectBase.Serialize(w); err != nil {
		return err
	}
	if p.tile {
		w.WriteBool("Tile", true)
	}
	if p.transparency != 0 {
		w.WriteFloat("Transparency", p.transparency)
	}
	return nil
}

// Deserialize reads PictureObject properties.
func (p *PictureObject) Deserialize(r report.Reader) error {
	if err := p.PictureObjectBase.Deserialize(r); err != nil {
		return err
	}
	p.tile = r.ReadBool("Tile", false)
	p.transparency = r.ReadFloat("Transparency", 0)
	return nil
}
