package object_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/object"
)

// -----------------------------------------------------------------------
// PictureObjectBase
// -----------------------------------------------------------------------

func TestNewPictureObjectBase_Defaults(t *testing.T) {
	p := object.NewPictureObjectBase()
	if p == nil {
		t.Fatal("NewPictureObjectBase returned nil")
	}
	if p.Angle() != 0 {
		t.Errorf("Angle default = %d, want 0", p.Angle())
	}
	if p.DataColumn() != "" {
		t.Errorf("DataColumn default = %q, want empty", p.DataColumn())
	}
	if p.Grayscale() {
		t.Error("Grayscale should default to false")
	}
	if p.ImageLocation() != "" {
		t.Errorf("ImageLocation default = %q, want empty", p.ImageLocation())
	}
	if p.MaxHeight() != 0 {
		t.Errorf("MaxHeight default = %v, want 0", p.MaxHeight())
	}
	if p.SizeMode() != object.SizeModeZoom {
		t.Errorf("SizeMode default = %d, want Zoom", p.SizeMode())
	}
	if p.ImageAlign() != object.ImageAlignNone {
		t.Errorf("ImageAlign default = %d, want None", p.ImageAlign())
	}
	if p.ShowErrorImage() {
		t.Error("ShowErrorImage should default to false")
	}
}

func TestPictureObjectBase_Angle(t *testing.T) {
	p := object.NewPictureObjectBase()
	p.SetAngle(90)
	if p.Angle() != 90 {
		t.Errorf("Angle = %d, want 90", p.Angle())
	}
}

func TestPictureObjectBase_DataColumn(t *testing.T) {
	p := object.NewPictureObjectBase()
	p.SetDataColumn("Photo")
	if p.DataColumn() != "Photo" {
		t.Errorf("DataColumn = %q, want Photo", p.DataColumn())
	}
}

func TestPictureObjectBase_Grayscale(t *testing.T) {
	p := object.NewPictureObjectBase()
	p.SetGrayscale(true)
	if !p.Grayscale() {
		t.Error("Grayscale should be true")
	}
}

func TestPictureObjectBase_ImageLocation(t *testing.T) {
	p := object.NewPictureObjectBase()
	p.SetImageLocation("http://example.com/img.png")
	if p.ImageLocation() != "http://example.com/img.png" {
		t.Errorf("ImageLocation = %q", p.ImageLocation())
	}
}

func TestPictureObjectBase_ImageSourceExpression(t *testing.T) {
	p := object.NewPictureObjectBase()
	p.SetImageSourceExpression("[ImagePath]")
	if p.ImageSourceExpression() != "[ImagePath]" {
		t.Errorf("ImageSourceExpression = %q", p.ImageSourceExpression())
	}
}

func TestPictureObjectBase_MaxDimensions(t *testing.T) {
	p := object.NewPictureObjectBase()
	p.SetMaxWidth(200)
	p.SetMaxHeight(150)
	if p.MaxWidth() != 200 || p.MaxHeight() != 150 {
		t.Errorf("Max dims = (%v,%v), want (200,150)", p.MaxWidth(), p.MaxHeight())
	}
}

func TestPictureObjectBase_Padding(t *testing.T) {
	p := object.NewPictureObjectBase()
	pad := object.Padding{Left: 5, Top: 5, Right: 5, Bottom: 5}
	p.SetPadding(pad)
	if p.Padding() != pad {
		t.Errorf("Padding = %+v, want %+v", p.Padding(), pad)
	}
}

func TestPictureObjectBase_SizeMode(t *testing.T) {
	modes := []object.SizeMode{
		object.SizeModeNormal,
		object.SizeModeStretchImage,
		object.SizeModeAutoSize,
		object.SizeModeCenterImage,
		object.SizeModeZoom,
	}
	for _, m := range modes {
		p := object.NewPictureObjectBase()
		p.SetSizeMode(m)
		if p.SizeMode() != m {
			t.Errorf("SizeMode = %d, want %d", p.SizeMode(), m)
		}
	}
}

func TestPictureObjectBase_ImageAlign(t *testing.T) {
	aligns := []object.ImageAlign{
		object.ImageAlignNone,
		object.ImageAlignTopLeft,
		object.ImageAlignCenterCenter,
		object.ImageAlignBottomRight,
	}
	for _, a := range aligns {
		p := object.NewPictureObjectBase()
		p.SetImageAlign(a)
		if p.ImageAlign() != a {
			t.Errorf("ImageAlign = %d, want %d", p.ImageAlign(), a)
		}
	}
}

func TestPictureObjectBase_ShowErrorImage(t *testing.T) {
	p := object.NewPictureObjectBase()
	p.SetShowErrorImage(true)
	if !p.ShowErrorImage() {
		t.Error("ShowErrorImage should be true")
	}
}

func TestPictureObjectBase_InheritsVisible(t *testing.T) {
	p := object.NewPictureObjectBase()
	if !p.Visible() {
		t.Error("PictureObjectBase should inherit Visible=true")
	}
}

// -----------------------------------------------------------------------
// PictureObject
// -----------------------------------------------------------------------

func TestNewPictureObject_Defaults(t *testing.T) {
	p := object.NewPictureObject()
	if p == nil {
		t.Fatal("NewPictureObject returned nil")
	}
	if p.ImageData() != nil {
		t.Error("ImageData should default to nil")
	}
	if p.Tile() {
		t.Error("Tile should default to false")
	}
	if p.Transparency() != 0 {
		t.Errorf("Transparency default = %v, want 0", p.Transparency())
	}
	if p.HasImage() {
		t.Error("HasImage should be false when no data/location set")
	}
}

func TestPictureObject_ImageData(t *testing.T) {
	p := object.NewPictureObject()
	data := []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic bytes
	p.SetImageData(data)
	if len(p.ImageData()) != 4 {
		t.Errorf("ImageData len = %d, want 4", len(p.ImageData()))
	}
	if p.ImageFormat() != object.ImageFormatUnknown {
		t.Error("Format should reset to Unknown after SetImageData")
	}
}

func TestPictureObject_ImageDataWithFormat(t *testing.T) {
	p := object.NewPictureObject()
	data := []byte{0xFF, 0xD8, 0xFF} // JPEG magic
	p.SetImageDataWithFormat(data, object.ImageFormatJpeg)
	if p.ImageFormat() != object.ImageFormatJpeg {
		t.Error("ImageFormat should be Jpeg")
	}
}

func TestPictureObject_HasImage_FromData(t *testing.T) {
	p := object.NewPictureObject()
	p.SetImageData([]byte{1, 2, 3})
	if !p.HasImage() {
		t.Error("HasImage should be true when imageData is set")
	}
}

func TestPictureObject_HasImage_FromLocation(t *testing.T) {
	p := object.NewPictureObject()
	p.SetImageLocation("/img/logo.png")
	if !p.HasImage() {
		t.Error("HasImage should be true when imageLocation is set")
	}
}

func TestPictureObject_HasImage_FromDataColumn(t *testing.T) {
	p := object.NewPictureObject()
	p.SetDataColumn("Products.Image")
	if !p.HasImage() {
		t.Error("HasImage should be true when dataColumn is set")
	}
}

func TestPictureObject_Tile(t *testing.T) {
	p := object.NewPictureObject()
	p.SetTile(true)
	if !p.Tile() {
		t.Error("Tile should be true")
	}
}

func TestPictureObject_Transparency(t *testing.T) {
	p := object.NewPictureObject()
	p.SetTransparency(0.5)
	if p.Transparency() != 0.5 {
		t.Errorf("Transparency = %v, want 0.5", p.Transparency())
	}
}

func TestPictureObject_InheritsSizeMode(t *testing.T) {
	p := object.NewPictureObject()
	if p.SizeMode() != object.SizeModeZoom {
		t.Errorf("SizeMode default = %d, want Zoom", p.SizeMode())
	}
}
