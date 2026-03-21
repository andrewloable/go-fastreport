package image_test

// image_formats_test.go — tests for the new image format features ported from
// C# ImageExport.cs: JPEG quality, GIF, BMP, TIFF, multi-frame TIFF,
// monochrome TIFF, combined-page mode, DPI/resolution scaling, and
// SeparateFiles flag.

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"testing"

	imgexport "github.com/andrewloable/go-fastreport/export/image"
	"github.com/andrewloable/go-fastreport/preview"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

// failWriter always returns an error on Write.
type failWriter struct{}

func (f failWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("intentional write failure")
}

// ── default format is JPEG (C# ImageExport constructor line 715) ──────────────

func TestImageExport_DefaultFormatIsJPEG(t *testing.T) {
	exp := imgexport.NewExporter()
	if exp.Format != imgexport.ImageFormatJPEG {
		t.Errorf("default Format = %v, want ImageFormatJPEG", exp.Format)
	}
}

func TestImageExport_DefaultJpegQuality(t *testing.T) {
	exp := imgexport.NewExporter()
	if exp.JpegQuality != 100 {
		t.Errorf("default JpegQuality = %d, want 100", exp.JpegQuality)
	}
}

func TestImageExport_DefaultResolution(t *testing.T) {
	exp := imgexport.NewExporter()
	if exp.ResolutionX != imgexport.DefaultDPI {
		t.Errorf("default ResolutionX = %d, want %d", exp.ResolutionX, imgexport.DefaultDPI)
	}
	if exp.ResolutionY != imgexport.DefaultDPI {
		t.Errorf("default ResolutionY = %d, want %d", exp.ResolutionY, imgexport.DefaultDPI)
	}
}

func TestImageExport_DefaultSeparateFiles(t *testing.T) {
	exp := imgexport.NewExporter()
	if !exp.SeparateFiles {
		t.Error("default SeparateFiles should be true")
	}
}

// ── JPEG output ───────────────────────────────────────────────────────────────

func TestImageExport_ProducesJPEG(t *testing.T) {
	pp := singlePageReport(200, 150, 30)
	exp := imgexport.NewExporter()
	// Default is already JPEG; set explicitly for clarity.
	exp.Format = imgexport.ImageFormatJPEG
	exp.JpegQuality = 80

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export JPEG: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("JPEG output is empty")
	}
	if _, err := jpeg.Decode(&buf); err != nil {
		t.Errorf("output is not valid JPEG: %v", err)
	}
}

func TestImageExport_JPEG_LowQuality(t *testing.T) {
	pp := singlePageReport(100, 80, 20)
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatJPEG
	exp.JpegQuality = 1 // minimum quality

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export JPEG q=1: %v", err)
	}
	if _, err := jpeg.Decode(bytes.NewReader(buf.Bytes())); err != nil {
		t.Errorf("low-quality JPEG output invalid: %v", err)
	}
}

func TestImageExport_JPEG_ZeroQualityClampedTo100(t *testing.T) {
	// Quality == 0 should clamp to 100 (encodeImage guard).
	pp := singlePageReport(50, 50, 10)
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatJPEG
	exp.JpegQuality = 0 // invalid → clamped to 100

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export JPEG q=0: %v", err)
	}
	if _, err := jpeg.Decode(bytes.NewReader(buf.Bytes())); err != nil {
		t.Errorf("clamped-quality JPEG output invalid: %v", err)
	}
}

// ── GIF output ────────────────────────────────────────────────────────────────

func TestImageExport_ProducesGIF(t *testing.T) {
	pp := singlePageReport(100, 80, 20)
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatGIF

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export GIF: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("GIF output is empty")
	}
	// GIF files start with "GIF8" magic.
	if len(buf.Bytes()) < 4 || string(buf.Bytes()[:4]) != "GIF8" {
		t.Errorf("output does not start with GIF8 magic: %q", string(buf.Bytes()[:6]))
	}
}

// ── BMP output ────────────────────────────────────────────────────────────────

func TestImageExport_ProducesBMP(t *testing.T) {
	pp := singlePageReport(100, 80, 20)
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatBMP

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export BMP: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("BMP output is empty")
	}
	if _, err := bmp.Decode(&buf); err != nil {
		t.Errorf("output is not valid BMP: %v", err)
	}
}

// ── TIFF output ───────────────────────────────────────────────────────────────

func TestImageExport_ProducesTIFF(t *testing.T) {
	pp := singlePageReport(100, 80, 20)
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatTIFF

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export TIFF: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("TIFF output is empty")
	}
	if _, err := tiff.Decode(&buf); err != nil {
		t.Errorf("output is not valid TIFF: %v", err)
	}
}

func TestImageExport_TIFF_Monochrome(t *testing.T) {
	pp := singlePageReport(60, 50, 15)
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatTIFF
	exp.MonochromeTiff = true

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export TIFF monochrome: %v", err)
	}
	if _, err := tiff.Decode(bytes.NewReader(buf.Bytes())); err != nil {
		t.Errorf("monochrome TIFF output invalid: %v", err)
	}
}

// ── multi-frame TIFF ──────────────────────────────────────────────────────────

func TestImageExport_MultiFrameTIFF(t *testing.T) {
	// Three pages → three TIFF frames concatenated in the output stream.
	pp := preview.New()
	for i := 0; i < 3; i++ {
		pp.AddPage(100, 80, i+1)
		_ = pp.AddBand(&preview.PreparedBand{
			Name:   "Band",
			Top:    0,
			Height: 20,
		})
	}

	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatTIFF
	exp.MultiFrameTiff = true

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export multi-frame TIFF: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("multi-frame TIFF output is empty")
	}
	// At least the first frame must be a decodable TIFF.
	if _, err := tiff.Decode(bytes.NewReader(buf.Bytes())); err != nil {
		t.Errorf("multi-frame TIFF first frame invalid: %v", err)
	}
}

func TestImageExport_MultiFrameTIFF_Monochrome(t *testing.T) {
	pp := preview.New()
	for i := 0; i < 2; i++ {
		pp.AddPage(60, 50, i+1)
		_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})
	}

	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatTIFF
	exp.MultiFrameTiff = true
	exp.MonochromeTiff = true

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export multi-frame monochrome TIFF: %v", err)
	}
	if _, err := tiff.Decode(bytes.NewReader(buf.Bytes())); err != nil {
		t.Errorf("multi-frame monochrome TIFF first frame invalid: %v", err)
	}
}

func TestImageExport_MultiFrameTIFF_EmptyFrames(t *testing.T) {
	// Zero pages → no TIFF frames; Finish should succeed with empty output.
	pp := preview.New()
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatTIFF
	exp.MultiFrameTiff = true

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export multi-frame TIFF empty: %v", err)
	}
	// No pages → no output (ExportBase returns early for zero pages).
}

// ── combined (non-separate) pages ────────────────────────────────────────────

func TestImageExport_CombinedPages_PNG(t *testing.T) {
	// Two pages combined into one tall PNG.
	pp := twoPageReport()
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatPNG
	exp.SeparateFiles = false

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export combined PNG: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("combined PNG output invalid: %v", err)
	}
	// Combined height should be sum of both page heights (300 + 300 = 600).
	if img.Bounds().Dy() != 600 {
		t.Errorf("combined image height = %d, want 600", img.Bounds().Dy())
	}
	if img.Bounds().Dx() != 400 {
		t.Errorf("combined image width = %d, want 400", img.Bounds().Dx())
	}
}

func TestImageExport_CombinedPages_WithPadding(t *testing.T) {
	pp := twoPageReport()
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatPNG
	exp.SeparateFiles = false
	exp.PaddingNonSeparatePages = 10 // 10px top + 10px bottom per page

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export combined PNG with padding: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("combined padded PNG invalid: %v", err)
	}
	// 2 pages × (300 height + 10 top + 10 bottom) = 640.
	wantH := 2 * (300 + 10 + 10)
	if img.Bounds().Dy() != wantH {
		t.Errorf("combined padded height = %d, want %d", img.Bounds().Dy(), wantH)
	}
}

func TestImageExport_CombinedPages_JPEG(t *testing.T) {
	pp := twoPageReport()
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatJPEG
	exp.SeparateFiles = false
	exp.JpegQuality = 90

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export combined JPEG: %v", err)
	}
	if _, err := jpeg.Decode(bytes.NewReader(buf.Bytes())); err != nil {
		t.Errorf("combined JPEG output invalid: %v", err)
	}
}

func TestImageExport_CombinedPages_Empty(t *testing.T) {
	// Zero pages combined → no output, no error.
	pp := preview.New()
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatPNG
	exp.SeparateFiles = false

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export combined empty: %v", err)
	}
}

// ── DPI / resolution scaling ──────────────────────────────────────────────────

func TestImageExport_Resolution192DPI(t *testing.T) {
	// 192 DPI = 2× the default 96 DPI → image should be twice as large.
	pp := singlePageReport(200, 150, 30)
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatPNG
	exp.ResolutionX = 192
	exp.ResolutionY = 192

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export 192 DPI: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("192 DPI PNG invalid: %v", err)
	}
	// At 192 DPI (2× scale), dimensions should be 400×300.
	if img.Bounds().Dx() != 400 {
		t.Errorf("192 DPI width = %d, want 400", img.Bounds().Dx())
	}
	if img.Bounds().Dy() != 300 {
		t.Errorf("192 DPI height = %d, want 300", img.Bounds().Dy())
	}
}

func TestImageExport_Resolution_ZeroFallback(t *testing.T) {
	// ResolutionX = 0 → effectiveScale uses 1.0 fallback (dpiRatio <= 0 guard).
	pp := singlePageReport(200, 150, 30)
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatPNG
	exp.ResolutionX = 0 // invalid → fallback to 1.0

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export zero DPI: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("zero DPI PNG invalid: %v", err)
	}
	if img.Bounds().Dx() != 200 {
		t.Errorf("zero DPI width = %d, want 200", img.Bounds().Dx())
	}
}

func TestImageExport_Resolution_AndScale(t *testing.T) {
	// 192 DPI + Scale 0.5 → effective scale = (192/96) × 0.5 = 1.0.
	pp := singlePageReport(200, 150, 30)
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatPNG
	exp.ResolutionX = 192
	exp.ResolutionY = 192
	exp.Scale = 0.5

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export 192 DPI + 0.5 scale: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("192 DPI + 0.5 scale PNG invalid: %v", err)
	}
	// Effective scale = 1.0 → unchanged size.
	if img.Bounds().Dx() != 200 {
		t.Errorf("192 DPI + 0.5 scale width = %d, want 200", img.Bounds().Dx())
	}
}

// ── PNG format ────────────────────────────────────────────────────────────────

func TestImageExport_PNG_ExplicitFormat(t *testing.T) {
	pp := singlePageReport(80, 60, 15)
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatPNG

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export PNG explicit: %v", err)
	}
	if _, err := png.Decode(bytes.NewReader(buf.Bytes())); err != nil {
		t.Errorf("explicit PNG output invalid: %v", err)
	}
}

// ── error paths ───────────────────────────────────────────────────────────────

func TestImageExport_CombinedPages_WriteError(t *testing.T) {
	// failWriter causes encodeImage to fail inside Finish (combined mode).
	pp := twoPageReport()
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatPNG
	exp.SeparateFiles = false

	if err := exp.Export(pp, failWriter{}); err == nil {
		t.Error("expected error from failing writer in combined mode, got nil")
	}
}

func TestImageExport_MultiFrameTIFF_WriteError(t *testing.T) {
	// failWriter causes encodeMultiFrameTIFF to fail inside Finish.
	pp := preview.New()
	pp.AddPage(50, 40, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 10})

	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatTIFF
	exp.MultiFrameTiff = true

	if err := exp.Export(pp, failWriter{}); err == nil {
		t.Error("expected error from failing writer in multi-frame TIFF, got nil")
	}
}

// ── effectiveScale: Scale <= 0 fallback ──────────────────────────────────────

func TestImageExport_NegativeScale_FallsBackToOne(t *testing.T) {
	// Scale <= 0 → effectiveScale uses 1.0 fallback.
	pp := singlePageReport(100, 80, 20)
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatPNG
	exp.Scale = -1.0 // invalid → fallback to 1.0

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export negative scale: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("negative scale PNG invalid: %v", err)
	}
	// Effective scale = 1.0 (Scale=-1 clamped, ResolutionX=96/96=1.0) → unchanged.
	if img.Bounds().Dx() != 100 {
		t.Errorf("negative scale width = %d, want 100", img.Bounds().Dx())
	}
}

// ── stitchPages: negative padding clamped to 0 ───────────────────────────────

func TestImageExport_CombinedPages_NegativePadding(t *testing.T) {
	// Negative PaddingNonSeparatePages is clamped to 0.
	pp := twoPageReport()
	exp := imgexport.NewExporter()
	exp.Format = imgexport.ImageFormatPNG
	exp.SeparateFiles = false
	exp.PaddingNonSeparatePages = -5 // negative → clamped to 0

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export combined negative padding: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("negative padding PNG invalid: %v", err)
	}
	// With pad=0, height = 300 + 300 = 600.
	if img.Bounds().Dy() != 600 {
		t.Errorf("negative padding height = %d, want 600", img.Bounds().Dy())
	}
}
