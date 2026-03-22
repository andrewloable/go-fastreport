package object

import (
	"encoding/base64"

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

// formatSizeMode converts SizeMode to its FRX string name (PictureBoxSizeMode enum).
func formatSizeMode(m SizeMode) string {
	switch m {
	case SizeModeNormal:
		return "Normal"
	case SizeModeStretchImage:
		return "StretchImage"
	case SizeModeAutoSize:
		return "AutoSize"
	case SizeModeCenterImage:
		return "CenterImage"
	default:
		return "Zoom"
	}
}

// parseSizeMode converts an FRX string to SizeMode (handles both names and ints).
func parseSizeMode(s string) SizeMode {
	switch s {
	case "Normal", "0":
		return SizeModeNormal
	case "StretchImage", "1":
		return SizeModeStretchImage
	case "AutoSize", "2":
		return SizeModeAutoSize
	case "CenterImage", "3":
		return SizeModeCenterImage
	default:
		return SizeModeZoom
	}
}

// formatImageAlign converts ImageAlign to its FRX string name.
// C# uses underscore-separated names: Top_Left, Center_Center, etc.
func formatImageAlign(a ImageAlign) string {
	switch a {
	case ImageAlignTopLeft:
		return "Top_Left"
	case ImageAlignTopCenter:
		return "Top_Center"
	case ImageAlignTopRight:
		return "Top_Right"
	case ImageAlignCenterLeft:
		return "Center_Left"
	case ImageAlignCenterCenter:
		return "Center_Center"
	case ImageAlignCenterRight:
		return "Center_Right"
	case ImageAlignBottomLeft:
		return "Bottom_Left"
	case ImageAlignBottomCenter:
		return "Bottom_Center"
	case ImageAlignBottomRight:
		return "Bottom_Right"
	default:
		return "None"
	}
}

// parseImageAlign converts an FRX string to ImageAlign (handles both names and ints).
func parseImageAlign(s string) ImageAlign {
	switch s {
	case "Top_Left", "1":
		return ImageAlignTopLeft
	case "Top_Center", "2":
		return ImageAlignTopCenter
	case "Top_Right", "3":
		return ImageAlignTopRight
	case "Center_Left", "4":
		return ImageAlignCenterLeft
	case "Center_Center", "5":
		return ImageAlignCenterCenter
	case "Center_Right", "6":
		return ImageAlignCenterRight
	case "Bottom_Left", "7":
		return ImageAlignBottomLeft
	case "Bottom_Center", "8":
		return ImageAlignBottomCenter
	case "Bottom_Right", "9":
		return ImageAlignBottomRight
	default:
		return ImageAlignNone
	}
}

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
	// savedPicState holds state saved by SaveState for RestoreState.
	savedPicState *pictureObjectSavedState
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

// savedSizeMode holds the pre-engine-pass SizeMode for SaveState/RestoreState.
// Mirrors C# PictureObjectBase.saveSizeMode field (PictureObjectBase.cs line 750-752).
type pictureObjectSavedState struct {
	sizeMode SizeMode
}

// GetExpressions returns the list of expressions used by this PictureObjectBase
// for pre-compilation by the report engine.
// Mirrors C# PictureObjectBase.GetExpressions (PictureObjectBase.cs line 875-894).
func (p *PictureObjectBase) GetExpressions() []string {
	exprs := p.ReportComponentBase.GetExpressions()
	if p.dataColumn != "" {
		exprs = append(exprs, p.dataColumn)
	}
	if p.imageSourceExpr != "" {
		// Strip enclosing brackets if present — C# strips them too.
		expr := p.imageSourceExpr
		if len(expr) > 2 && expr[0] == '[' && expr[len(expr)-1] == ']' {
			expr = expr[1 : len(expr)-1]
		}
		exprs = append(exprs, expr)
	}
	return exprs
}

// SaveState saves the current SizeMode (in addition to the ReportComponentBase
// state) so RestoreState can undo engine-pass changes.
// Mirrors C# PictureObjectBase.SaveState (PictureObjectBase.cs line 748-752).
func (p *PictureObjectBase) SaveState() {
	p.ReportComponentBase.SaveState()
	p.savedPicState = &pictureObjectSavedState{sizeMode: p.sizeMode}
}

// RestoreState restores the SizeMode saved by SaveState.
// Mirrors C# PictureObjectBase.RestoreState (PictureObjectBase.cs line 722-727).
func (p *PictureObjectBase) RestoreState() {
	p.ReportComponentBase.RestoreState()
	if p.savedPicState != nil {
		p.sizeMode = p.savedPicState.sizeMode
		p.savedPicState = nil
	}
}

// GetData evaluates the DataColumn or ImageSourceExpression binding using the
// provided calc function. When DataColumn is set, the evaluated string is
// stored as the new ImageLocation. When ImageSourceExpression is set, the
// expression is evaluated and stored as DataColumn if it is a column reference,
// or as ImageLocation otherwise.
// Mirrors C# PictureObjectBase behaviour: DataColumn → GetColumnValue → image bytes;
// ImageSourceExpression → Calc → resolve as DataColumn or path.
func (p *PictureObjectBase) GetData(calc func(string) (any, error)) {
	if p.dataColumn != "" {
		val, err := calc("[" + p.dataColumn + "]")
		if err == nil && val != nil {
			// Convert to string for image location / URL. Actual image loading
			// happens in the engine/exporter layer.
			if s, ok := val.(string); ok {
				p.imageLocation = s
			}
		}
	} else if p.imageSourceExpr != "" {
		val, err := calc(p.imageSourceExpr)
		if err == nil && val != nil {
			if s, ok := val.(string); ok {
				p.imageLocation = s
			}
		}
	}
}

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
		w.WriteStr("SizeMode", formatSizeMode(p.sizeMode))
	}
	if p.imageAlign != ImageAlignNone {
		w.WriteStr("ImageAlign", formatImageAlign(p.imageAlign))
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
	p.sizeMode = parseSizeMode(r.ReadStr("SizeMode", "Zoom"))
	p.imageAlign = parseImageAlign(r.ReadStr("ImageAlign", "None"))
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
	// FRX stores inline image data as a base64-encoded "Image" attribute.
	if imgStr := r.ReadStr("Image", ""); imgStr != "" {
		if decoded, err := base64.StdEncoding.DecodeString(imgStr); err == nil {
			p.imageData = decoded
		}
	}
	// Detect format from the "ImageFormat" attribute when present.
	switch r.ReadStr("ImageFormat", "") {
	case "Png":
		p.imageFormat = ImageFormatPng
	case "Jpeg":
		p.imageFormat = ImageFormatJpeg
	case "Gif":
		p.imageFormat = ImageFormatGif
	case "Bmp":
		p.imageFormat = ImageFormatBmp
	}
	return nil
}
