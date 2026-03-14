package image_test

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	imgexport "github.com/andrewloable/go-fastreport/export/image"
	"github.com/andrewloable/go-fastreport/preview"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func singlePageReport(width, height float32, bandHeight float32) *preview.PreparedPages {
	pp := preview.New()
	pp.AddPage(width, height, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "DataBand1",
		Top:    0,
		Height: bandHeight,
	})
	return pp
}

func twoPageReport() *preview.PreparedPages {
	pp := preview.New()
	for i := 0; i < 2; i++ {
		pp.AddPage(400, 300, i+1)
		_ = pp.AddBand(&preview.PreparedBand{
			Name:   "DataBand",
			Top:    float32(i * 10),
			Height: 20,
		})
	}
	return pp
}

// ── basic output ──────────────────────────────────────────────────────────────

func TestImageExport_ProducesPNG(t *testing.T) {
	pp := singlePageReport(400, 300, 50)
	exp := imgexport.NewExporter()

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("output is empty")
	}

	// Verify output is a valid PNG.
	_, err := png.Decode(&buf)
	if err != nil {
		t.Errorf("output is not valid PNG: %v", err)
	}
}

func TestImageExport_ImageDimensions(t *testing.T) {
	pp := singlePageReport(500, 400, 100)
	exp := imgexport.NewExporter()

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("PNG decode: %v", err)
	}

	if img.Bounds().Dx() != 500 {
		t.Errorf("image width = %d, want 500", img.Bounds().Dx())
	}
	if img.Bounds().Dy() != 400 {
		t.Errorf("image height = %d, want 400", img.Bounds().Dy())
	}
}

func TestImageExport_BackgroundIsWhite(t *testing.T) {
	pp := singlePageReport(100, 100, 10)
	exp := imgexport.NewExporter()

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	img, _ := png.Decode(bytes.NewReader(buf.Bytes()))
	// Bottom-right pixel should be white (not covered by the 10px band).
	r, g, b, _ := img.At(99, 99).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("background pixel is not white: r=%d g=%d b=%d", r>>8, g>>8, b>>8)
	}
}

// ── scale ─────────────────────────────────────────────────────────────────────

func TestImageExport_Scale(t *testing.T) {
	pp := singlePageReport(200, 150, 30)
	exp := imgexport.NewExporter()
	exp.Scale = 2.0

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("PNG decode: %v", err)
	}

	if img.Bounds().Dx() != 400 {
		t.Errorf("scaled width = %d, want 400", img.Bounds().Dx())
	}
	if img.Bounds().Dy() != 300 {
		t.Errorf("scaled height = %d, want 300", img.Bounds().Dy())
	}
}

// ── multi-page (sequential PNG output) ───────────────────────────────────────

func TestImageExport_MultiPage_NonEmpty(t *testing.T) {
	pp := twoPageReport()
	exp := imgexport.NewExporter()

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("multi-page output should not be empty")
	}
}

// ── nil/empty pages ───────────────────────────────────────────────────────────

func TestImageExport_NilPages_Error(t *testing.T) {
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(nil, &buf); err == nil {
		t.Error("expected error for nil PreparedPages")
	}
}

func TestImageExport_EmptyPages_NoError(t *testing.T) {
	pp := preview.New()
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Errorf("unexpected error for empty pages: %v", err)
	}
}

// ── custom colors ─────────────────────────────────────────────────────────────

func TestImageExport_CustomBackgroundColor(t *testing.T) {
	pp := singlePageReport(100, 100, 10)
	exp := imgexport.NewExporter()
	exp.BackgroundColor = color.RGBA{R: 0, G: 128, B: 0, A: 255} // green

	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}

	img, _ := png.Decode(bytes.NewReader(buf.Bytes()))
	// Bottom-right pixel should be green (not covered by 10px band).
	r, g, b, _ := img.At(99, 99).RGBA()
	if r != 0 || g>>8 != 128 || b != 0 {
		t.Errorf("background pixel not green: r=%d g=%d b=%d", r>>8, g>>8, b>>8)
	}
}

// ── page range ────────────────────────────────────────────────────────────────

func TestImageExport_PageRangeCurrent(t *testing.T) {
	pp := twoPageReport()
	exp := imgexport.NewExporter()
	exp.PageRange = 1 // PageRangeCurrent
	exp.CurPage = 1   // first page

	var buf1 bytes.Buffer
	if err := exp.Export(pp, &buf1); err != nil {
		t.Fatalf("Export page 1: %v", err)
	}
	if buf1.Len() == 0 {
		t.Error("expected non-empty output for single page")
	}
}

// ── accessor defaults ─────────────────────────────────────────────────────────

func TestImageExport_DefaultScale(t *testing.T) {
	exp := imgexport.NewExporter()
	if exp.Scale != 1.0 {
		t.Errorf("Scale default = %v, want 1.0", exp.Scale)
	}
}

func TestImageExport_DefaultBackground(t *testing.T) {
	exp := imgexport.NewExporter()
	want := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if exp.BackgroundColor != want {
		t.Errorf("BackgroundColor = %v, want %v", exp.BackgroundColor, want)
	}
}
