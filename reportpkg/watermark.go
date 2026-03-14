package reportpkg

import (
	"image/color"

	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
)

// WatermarkTextRotation controls the angle of watermark text.
type WatermarkTextRotation int

const (
	// WatermarkTextRotationHorizontal renders text horizontally.
	WatermarkTextRotationHorizontal WatermarkTextRotation = iota
	// WatermarkTextRotationVertical renders text vertically.
	WatermarkTextRotationVertical
	// WatermarkTextRotationForwardDiagonal renders text at a forward diagonal (default).
	WatermarkTextRotationForwardDiagonal
	// WatermarkTextRotationBackwardDiagonal renders text at a backward diagonal.
	WatermarkTextRotationBackwardDiagonal
)

// WatermarkImageSize controls how a watermark image is sized.
type WatermarkImageSize int

const (
	// WatermarkImageSizeNormal displays the image at its original size.
	WatermarkImageSizeNormal WatermarkImageSize = iota
	// WatermarkImageSizeCenter centres the image without scaling.
	WatermarkImageSizeCenter
	// WatermarkImageSizeStretch stretches the image to fill the page.
	WatermarkImageSizeStretch
	// WatermarkImageSizeZoom scales the image proportionally (default).
	WatermarkImageSizeZoom
	// WatermarkImageSizeTile tiles the image across the page.
	WatermarkImageSizeTile
)

// Watermark defines an optional background (or foreground) overlay printed on
// every page of a ReportPage. It is the Go equivalent of FastReport.Watermark.
type Watermark struct {
	// Enabled controls whether the watermark is rendered.
	Enabled bool

	// ── Text watermark ───────────────────────────────────────────────────────

	// Text is the watermark text. An empty string disables text rendering.
	Text string
	// Font is the font used to render the watermark text.
	Font style.Font
	// TextRotation controls the angle of the watermark text.
	TextRotation WatermarkTextRotation
	// ShowTextOnTop renders the text on top of page objects (true) or behind them (false).
	ShowTextOnTop bool
	// TextFillColor is the color of the watermark text (RGBA).
	TextFillColor color.RGBA

	// ── Image watermark ──────────────────────────────────────────────────────

	// ImageData holds raw encoded bytes of the watermark image (nil = no image).
	ImageData []byte
	// ImageSize controls how the watermark image is sized on the page.
	ImageSize WatermarkImageSize
	// ImageTransparency is the image opacity (0.0 = opaque, 1.0 = invisible).
	ImageTransparency float32
	// ShowImageOnTop renders the image on top of page objects (true) or behind (false).
	ShowImageOnTop bool
}

// NewWatermark creates a Watermark with sensible defaults matching FastReport .NET.
func NewWatermark() *Watermark {
	return &Watermark{
		Font: style.Font{
			Name: "Arial",
			Size: 60,
		},
		TextRotation:  WatermarkTextRotationForwardDiagonal,
		ShowTextOnTop: true,
		ImageSize:     WatermarkImageSizeZoom,
	}
}

// Serialize writes Watermark properties that differ from defaults.
// Properties are written with a "Watermark." prefix to match the FRX format.
func (wm *Watermark) Serialize(w report.Writer) {
	if !wm.Enabled {
		return // nothing to write if not enabled
	}
	w.WriteBool("Watermark.Enabled", true)
	if wm.Text != "" {
		w.WriteStr("Watermark.Text", wm.Text)
	}
	def := NewWatermark()
	if wm.Font != def.Font {
		w.WriteStr("Watermark.Font", style.FontToStr(wm.Font))
	}
	if wm.TextRotation != def.TextRotation {
		w.WriteInt("Watermark.TextRotation", int(wm.TextRotation))
	}
	if !wm.ShowTextOnTop {
		w.WriteBool("Watermark.ShowTextOnTop", false)
	}
	if wm.ImageSize != def.ImageSize {
		w.WriteInt("Watermark.ImageSize", int(wm.ImageSize))
	}
	if wm.ImageTransparency != 0 {
		w.WriteFloat("Watermark.ImageTransparency", wm.ImageTransparency)
	}
	if wm.ShowImageOnTop {
		w.WriteBool("Watermark.ShowImageOnTop", true)
	}
}

// Deserialize reads Watermark properties from an FRReader.
// Properties are expected with a "Watermark." prefix.
func (wm *Watermark) Deserialize(r report.Reader) {
	wm.Enabled = r.ReadBool("Watermark.Enabled", false)
	wm.Text = r.ReadStr("Watermark.Text", "")
	if fs := r.ReadStr("Watermark.Font", ""); fs != "" {
		wm.Font = style.FontFromStr(fs)
	}
	wm.TextRotation = WatermarkTextRotation(r.ReadInt("Watermark.TextRotation", int(WatermarkTextRotationForwardDiagonal)))
	wm.ShowTextOnTop = r.ReadBool("Watermark.ShowTextOnTop", true)
	wm.ImageSize = WatermarkImageSize(r.ReadInt("Watermark.ImageSize", int(WatermarkImageSizeZoom)))
	wm.ImageTransparency = r.ReadFloat("Watermark.ImageTransparency", 0)
	wm.ShowImageOnTop = r.ReadBool("Watermark.ShowImageOnTop", false)
	if cs := r.ReadStr("Watermark.TextFill.Color", ""); cs != "" {
		if c, err := utils.ParseColor(cs); err == nil {
			wm.TextFillColor = c
		}
	}
}
