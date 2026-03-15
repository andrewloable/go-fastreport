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
	"net/http"
	"os"
	"strings"

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

// ImageToBytes encodes an image.Image to PNG or JPEG bytes.
func ImageToBytes(img image.Image, fmt ImageFormat) ([]byte, error) {
	var buf bytes.Buffer
	switch fmt {
	case ImageFormatJPEG:
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, err
		}
	default:
		if err := png.Encode(&buf, img); err != nil {
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
