// Package utils provides image loading, resizing, and encoding helpers
// equivalent to FastReport's Utils/ImageHelper.cs.
package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"os"
	"strings"

	_ "golang.org/x/image/bmp" // register BMP decoder for ICO-embedded BMP
	xdraw "golang.org/x/image/draw"
)

// ImageFormat identifies the output encoding for ImageToBytes.
type ImageFormat int

const (
	ImageFormatPNG  ImageFormat = iota
	ImageFormatJPEG             // quality 90
)

// SizeMode mirrors object.SizeMode for use in ResizeImage without creating an import cycle.
type SizeMode int

const (
	SizeModeNormal       SizeMode = iota // original size, clipped
	SizeModeStretchImage                  // fill bounds, distort if needed
	SizeModeAutoSize                      // (caller handles; treated as normal here)
	SizeModeCenterImage                   // centered, no scaling
	SizeModeZoom                          // proportional fit within bounds
)

// LoadImage loads an image from:
//   - a file path
//   - a data URI  (data:image/...;base64,<b64>)
//   - an HTTP/HTTPS URL
//   - raw base64 (no prefix, tried last)
func LoadImage(source string) (image.Image, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, fmt.Errorf("LoadImage: empty source")
	}

	// Data URI
	if strings.HasPrefix(source, "data:") {
		return loadFromDataURI(source)
	}

	// HTTP / HTTPS
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return loadFromURL(source)
	}

	// File path
	if _, err := os.Stat(source); err == nil {
		return loadFromFile(source)
	}

	// Fallback: attempt raw base64
	return loadFromBase64(source)
}

// BytesToImage decodes raw bytes (PNG or JPEG) into an image.Image.
func BytesToImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// imageToBytesWriter is a package-level hook so tests can inject a custom
// io.Writer into ImageToBytes to exercise its encode-error paths.
// In production it always returns nil, causing ImageToBytes to use its own
// internal bytes.Buffer. A non-nil value is used by tests only.
var imageToBytesWriter func() io.Writer

// ImageToBytes encodes an image.Image to PNG or JPEG bytes.
func ImageToBytes(img image.Image, fmt ImageFormat) ([]byte, error) {
	var buf bytes.Buffer
	var w io.Writer = &buf
	if imageToBytesWriter != nil {
		w = imageToBytesWriter()
	}
	switch fmt {
	case ImageFormatJPEG:
		if err := jpeg.Encode(w, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, err
		}
	default:
		if err := png.Encode(w, img); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// ResizeImage scales src to fit within the target (dstW × dstH) box according
// to mode. The returned image is exactly dstW×dstH (with transparent/white
// padding where required).
func ResizeImage(src image.Image, dstW, dstH int, mode SizeMode) image.Image {
	if src == nil || dstW <= 0 || dstH <= 0 {
		return src
	}
	srcB := src.Bounds()
	srcW := srcB.Dx()
	srcH := srcB.Dy()

	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	// Fill with white background.
	fillRect(dst, color.White)

	switch mode {
	case SizeModeStretchImage:
		// Stretch to fill, ignoring aspect ratio.
		scaleDraw(src, srcB, dst, dst.Bounds())

	case SizeModeCenterImage:
		// Center at original size.
		offX := (dstW - srcW) / 2
		offY := (dstH - srcH) / 2
		dstRect := image.Rect(offX, offY, offX+srcW, offY+srcH).
			Intersect(dst.Bounds())
		srcOffX := 0
		srcOffY := 0
		if offX < 0 {
			srcOffX = -offX
		}
		if offY < 0 {
			srcOffY = -offY
		}
		srcClip := image.Rect(srcOffX, srcOffY, srcOffX+dstRect.Dx(), srcOffY+dstRect.Dy())
		scaleDraw(src, srcClip, dst, dstRect)

	case SizeModeZoom:
		// Proportional fit within bounds.
		scaleX := float64(dstW) / float64(srcW)
		scaleY := float64(dstH) / float64(srcH)
		scale := scaleX
		if scaleY < scale {
			scale = scaleY
		}
		fitW := int(float64(srcW) * scale)
		fitH := int(float64(srcH) * scale)
		offX := (dstW - fitW) / 2
		offY := (dstH - fitH) / 2
		scaleDraw(src, srcB, dst, image.Rect(offX, offY, offX+fitW, offY+fitH))

	default: // SizeModeNormal, SizeModeAutoSize
		// Draw at original size, clipped to dst bounds.
		clip := srcB.Intersect(dst.Bounds())
		scaleDraw(src, clip, dst, clip)
	}
	return dst
}

// ApplyGrayscale converts img to grayscale using the NTSC luminance weights
// (R×0.299 + G×0.587 + B×0.114), matching C# ImageHelper.GetGrayscaleBitmap.
// The alpha channel is preserved unchanged.
// C# ref: FastReport.Base/Utils/ImageHelper.cs GetGrayscaleBitmap
func ApplyGrayscale(src image.Image) image.Image {
	if src == nil {
		return nil
	}
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA()
			// RGBA() returns alpha-premultiplied 16-bit values; >>8 gives 8-bit.
			rf := float64(r>>8) * 0.299
			gf := float64(g>>8) * 0.587
			bf := float64(bl>>8) * 0.114
			lum := uint8(math.Round(rf + gf + bf))
			dst.SetNRGBA(x, y, color.NRGBA{R: lum, G: lum, B: lum, A: uint8(a >> 8)})
		}
	}
	return dst
}

// ApplyTransparency multiplies the alpha channel of every pixel by (1 - transparency),
// matching C# ImageHelper.GetTransparentBitmap where Matrix33 = 1 - transparency.
// transparency 0.0 → fully opaque (no change); 1.0 → fully invisible.
// C# ref: FastReport.Base/Utils/ImageHelper.cs GetTransparentBitmap
func ApplyTransparency(src image.Image, transparency float32) image.Image {
	if src == nil {
		return nil
	}
	if transparency <= 0 {
		return src
	}
	factor := 1.0 - float64(transparency)
	if factor < 0 {
		factor = 0
	}
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA()
			newA := uint8(float64(a>>8) * factor)
			dst.SetNRGBA(x, y, color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(bl >> 8),
				A: newA,
			})
		}
	}
	return dst
}

// ── internal helpers ──────────────────────────────────────────────────────────

func loadFromDataURI(uri string) (image.Image, error) {
	// Format: data:<mediatype>[;base64],<data>
	comma := strings.Index(uri, ",")
	if comma < 0 {
		return nil, fmt.Errorf("loadFromDataURI: malformed data URI")
	}
	header := uri[:comma]
	body := uri[comma+1:]
	isBase64 := strings.Contains(header, ";base64")
	var raw []byte
	if isBase64 {
		var err error
		raw, err = base64.StdEncoding.DecodeString(body)
		if err != nil {
			// Try URL-safe variant.
			raw, err = base64.URLEncoding.DecodeString(body)
			if err != nil {
				return nil, fmt.Errorf("loadFromDataURI: base64 decode: %w", err)
			}
		}
	} else {
		raw = []byte(body)
	}
	return BytesToImage(raw)
}

func loadFromURL(url string) (image.Image, error) {
	resp, err := http.Get(url) //nolint:gosec // URL comes from report designer, not user input
	if err != nil {
		return nil, fmt.Errorf("loadFromURL %q: %w", url, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("loadFromURL %q: read body: %w", url, err)
	}
	return BytesToImage(data)
}

func loadFromFile(path string) (image.Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loadFromFile %q: %w", path, err)
	}
	return BytesToImage(data)
}

func loadFromBase64(s string) (image.Image, error) {
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		raw, err = base64.URLEncoding.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("loadFromBase64: not a valid file path or base64 string")
		}
	}
	return BytesToImage(raw)
}

func fillRect(dst *image.NRGBA, c color.Color) {
	b := dst.Bounds()
	draw.Draw(dst, b, &image.Uniform{C: c}, image.Point{}, draw.Src)
}

func scaleDraw(src image.Image, srcRect image.Rectangle, dst *image.NRGBA, dstRect image.Rectangle) {
	if srcRect.Empty() || dstRect.Empty() {
		return
	}
	xdraw.CatmullRom.Scale(dst, dstRect, src, srcRect, xdraw.Over, nil)
}

// DecodeICOImage extracts and decodes the first image from ICO format data.
// ICO files embed BMP or PNG image data. C# System.Drawing.Image.FromStream
// handles ICO natively; this provides equivalent Go decoding.
// Returns nil if the data is not valid ICO or cannot be decoded.
func DecodeICOImage(data []byte) image.Image {
	if len(data) < 22 { // 6 header + 16 directory entry minimum
		return nil
	}
	// ICO magic: reserved=0, type=1.
	if data[0] != 0 || data[1] != 0 || data[2] != 1 || data[3] != 0 {
		return nil
	}
	// Read data offset and size from the first directory entry (starts at byte 6).
	dataSize := int(data[14]) | int(data[15])<<8 | int(data[16])<<16 | int(data[17])<<24
	dataOffset := int(data[18]) | int(data[19])<<8 | int(data[20])<<16 | int(data[21])<<24
	if dataOffset < 0 || dataOffset >= len(data) || dataSize <= 0 || dataOffset+dataSize > len(data) {
		return nil
	}
	imgData := data[dataOffset : dataOffset+dataSize]
	// Embedded PNG? (magic: 89 50 4E 47)
	if len(imgData) >= 8 && imgData[0] == 0x89 && imgData[1] == 0x50 {
		if img, _, err := image.Decode(bytes.NewReader(imgData)); err == nil {
			return img
		}
	}
	// Otherwise embedded BMP DIB (BITMAPINFOHEADER).
	if len(imgData) < 40 {
		return nil
	}
	dibHeaderSize := int(imgData[0]) | int(imgData[1])<<8 | int(imgData[2])<<16 | int(imgData[3])<<24
	bitsPerPixel := int(imgData[14]) | int(imgData[15])<<8
	colorUsed := int(imgData[32]) | int(imgData[33])<<8 | int(imgData[34])<<16 | int(imgData[35])<<24
	paletteSize := 0
	if bitsPerPixel <= 8 {
		if colorUsed == 0 {
			colorUsed = 1 << uint(bitsPerPixel)
		}
		paletteSize = colorUsed * 4
	}
	bmpHeaderSize := 14
	pixelOffset := bmpHeaderSize + dibHeaderSize + paletteSize
	// ICO BMPs store double-height (image + AND mask). Halve the height.
	icoHeight := int(imgData[8]) | int(imgData[9])<<8 | int(imgData[10])<<16 | int(imgData[11])<<24
	halfHeight := icoHeight / 2
	// Build a full BMP file: 14-byte file header + modified DIB.
	modImgData := make([]byte, len(imgData))
	copy(modImgData, imgData)
	modImgData[8] = byte(halfHeight)
	modImgData[9] = byte(halfHeight >> 8)
	modImgData[10] = byte(halfHeight >> 16)
	modImgData[11] = byte(halfHeight >> 24)
	fileSize := bmpHeaderSize + len(modImgData)
	bmpFile := make([]byte, fileSize)
	bmpFile[0] = 'B'
	bmpFile[1] = 'M'
	bmpFile[2] = byte(fileSize)
	bmpFile[3] = byte(fileSize >> 8)
	bmpFile[4] = byte(fileSize >> 16)
	bmpFile[5] = byte(fileSize >> 24)
	bmpFile[10] = byte(pixelOffset)
	bmpFile[11] = byte(pixelOffset >> 8)
	bmpFile[12] = byte(pixelOffset >> 16)
	bmpFile[13] = byte(pixelOffset >> 24)
	copy(bmpFile[bmpHeaderSize:], modImgData)
	img, _, err := image.Decode(bytes.NewReader(bmpFile))
	if err != nil {
		return nil
	}
	return img
}
