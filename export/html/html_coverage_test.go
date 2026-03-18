package html_test

// html_coverage_test.go — targeted tests that cover the remaining branches in html.go:
//
//  • ExportPageBegin:       scale <= 0 → clamp to 1.
//  • renderWatermarkText:   WatermarkTextRotationVertical (rotDeg=90),
//                           WatermarkTextRotationBackwardDiagonal (rotDeg=45),
//                           zero-value TextColor (alpha=0).
//  • renderWatermarkImage:  blob data is empty → early return (len(imgData)==0).
//  • ExportBand:            scale <= 0 → clamp to 1.
//  • ExportPageEnd:         scale <= 0 → clamp to 1 (watermark-on-top path).

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/export/html"
	"github.com/andrewloable/go-fastreport/preview"
)

// ── helpers ────────────────────────────────────────────────────────────────────

// exportHTMLWith runs Export on pp using exp and returns the HTML output string.
func exportHTMLWith(t *testing.T, exp *html.Exporter, pp *preview.PreparedPages) string {
	t.Helper()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	return buf.String()
}

// singlePageWithBand returns a PreparedPages with one page and one empty band.
func singlePageWithBand() *preview.PreparedPages {
	pp := preview.New()
	pp.AddPage(594, 841, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 30})
	return pp
}

// singlePageWithWatermark returns a PreparedPages whose page has a watermark.
func singlePageWithWatermark(wm *preview.PreparedWatermark) *preview.PreparedPages {
	pp := singlePageWithBand()
	pp.GetPage(0).Watermark = wm
	return pp
}

// ── ExportPageBegin: scale <= 0 ────────────────────────────────────────────────

func TestExporter_ZeroScale_ExportPageBegin(t *testing.T) {
	// Scale=0 should be clamped to 1.0 in ExportPageBegin, so the page div should
	// show the raw page dimensions (594 × 841 at scale 1) plus the +3 C# offset.
	pp := singlePageWithBand()
	exp := html.NewExporter()
	exp.Scale = 0 // triggers the `scale <= 0 → scale = 1` branch

	out := exportHTMLWith(t, exp, pp)
	// The page div must still be present even when Scale==0.
	// New structure uses frpage0 class (not class="page").
	if !strings.Contains(out, `class="frpage0"`) {
		t.Error("ZeroScale: expected frpage0 div in output")
	}
	// Width must be page width + 3 (C# +3 offset): 594+3=597px.
	if !strings.Contains(out, "597px") {
		t.Errorf("ZeroScale: expected 597px page width, got:\n%s", out)
	}
}

func TestExporter_NegativeScale_ExportPageBegin(t *testing.T) {
	// Same as zero scale: negative clamps to 1.
	pp := singlePageWithBand()
	exp := html.NewExporter()
	exp.Scale = -2

	out := exportHTMLWith(t, exp, pp)
	// Width must be page width + 3 (C# +3 offset): 594+3=597px.
	if !strings.Contains(out, "597px") {
		t.Errorf("NegativeScale: expected 597px, got:\n%s", out)
	}
}

// ── renderWatermarkText: Vertical rotation ─────────────────────────────────────

func TestExporter_WatermarkText_Vertical(t *testing.T) {
	wm := &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "VERTICAL",
		TextColor:    color.RGBA{R: 0, G: 0, B: 128, A: 200},
		TextRotation: preview.WatermarkTextRotationVertical,
		ImageBlobIdx: -1,
	}
	pp := singlePageWithWatermark(wm)
	exp := html.NewExporter()
	out := exportHTMLWith(t, exp, pp)

	if !strings.Contains(out, "VERTICAL") {
		t.Error("Vertical watermark: text not found in output")
	}
	// rotDeg=90 → transform:rotate(90deg)
	if !strings.Contains(out, "rotate(90deg)") {
		t.Errorf("Vertical watermark: expected rotate(90deg), got:\n%s", out)
	}
}

// ── renderWatermarkText: BackwardDiagonal rotation ────────────────────────────

func TestExporter_WatermarkText_BackwardDiagonal(t *testing.T) {
	wm := &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "BACKWARD",
		TextColor:    color.RGBA{R: 100, G: 50, B: 0, A: 180},
		TextRotation: preview.WatermarkTextRotationBackwardDiagonal,
		ImageBlobIdx: -1,
	}
	pp := singlePageWithWatermark(wm)
	exp := html.NewExporter()
	out := exportHTMLWith(t, exp, pp)

	if !strings.Contains(out, "BACKWARD") {
		t.Error("BackwardDiagonal watermark: text not found")
	}
	// rotDeg=45 → transform:rotate(45deg)
	if !strings.Contains(out, "rotate(45deg)") {
		t.Errorf("BackwardDiagonal watermark: expected rotate(45deg), got:\n%s", out)
	}
}

// ── renderWatermarkText: zero-value TextColor (alpha == 0) ────────────────────

func TestExporter_WatermarkText_ZeroAlphaColor(t *testing.T) {
	// TextColor.A == 0 → rgba(0,0,0,0.00) — no special branch but exercises
	// the rgba() format path with a zero-alpha colour.
	wm := &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "TRANSPARENT",
		TextColor:    color.RGBA{R: 0, G: 0, B: 0, A: 0}, // zero value
		TextRotation: preview.WatermarkTextRotationHorizontal,
		ImageBlobIdx: -1,
	}
	pp := singlePageWithWatermark(wm)
	exp := html.NewExporter()
	out := exportHTMLWith(t, exp, pp)

	if !strings.Contains(out, "TRANSPARENT") {
		t.Error("ZeroAlpha watermark: text not found")
	}
	// The rendered colour should be rgba(0,0,0,0.00).
	if !strings.Contains(out, "rgba(0,0,0,0.00)") {
		t.Errorf("ZeroAlpha watermark: expected rgba(0,0,0,0.00), got:\n%s", out)
	}
}

// ── renderWatermarkImage: empty blob data ──────────────────────────────────────

func TestExporter_WatermarkImage_EmptyBlobData(t *testing.T) {
	// Add a blob with empty bytes — BlobStore.Add stores it, but Get returns the
	// empty slice. renderWatermarkImage must return early without emitting anything.
	pp := preview.New()
	pp.AddPage(594, 841, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 30})

	// Add an empty blob; Get will return a non-nil but zero-length slice.
	idx := pp.BlobStore.Add("empty-img", []byte{})
	// Sanity-check: Get returns the empty slice (not nil).
	if data := pp.BlobStore.Get(idx); data == nil {
		t.Skip("BlobStore.Get returned nil for empty blob — skipping")
	}

	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:      true,
		Text:         "",  // no text watermark
		ImageBlobIdx: idx, // points to empty blob
		ImageSize:    preview.WatermarkImageSizeCenter,
	}

	exp := html.NewExporter()
	out := exportHTMLWith(t, exp, pp)

	// The output should NOT contain background-image since blob data is empty.
	if strings.Contains(out, "background-image") {
		t.Error("EmptyBlobData: expected no background-image for empty blob")
	}
}

// ── ExportBand: scale <= 0 ────────────────────────────────────────────────────

func TestExporter_ZeroScale_ExportBand(t *testing.T) {
	// Scale=0 must be clamped to 1.0 inside ExportBand.
	// With flat rendering (no band wrappers), the page div should still appear.
	// We verify the page structure is valid with clamped scale.
	pp := preview.New()
	pp.AddPage(594, 841, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "MyBand", Top: 100, Height: 40})

	exp := html.NewExporter()
	exp.Scale = 0

	out := exportHTMLWith(t, exp, pp)

	// With clamped scale=1 and flat rendering, page div must be present.
	if !strings.Contains(out, `class="frpage0"`) {
		t.Errorf("ZeroScale ExportBand: expected frpage0 div, got:\n%s", out)
	}
	// Width must use clamped scale=1: 594+3=597px.
	if !strings.Contains(out, "597px") {
		t.Errorf("ZeroScale ExportBand: expected 597px page width, got:\n%s", out)
	}
}

// ── ExportPageEnd: scale <= 0 with ShowImageOnTop watermark ──────────────────

func TestExporter_ZeroScale_ExportPageEnd_WatermarkOnTop(t *testing.T) {
	// ExportPageEnd renders watermark-on-top. With Scale=0 clamped to 1,
	// the page dimensions should appear as the raw page size.
	pp := preview.New()
	pp.AddPage(400, 600, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "Band", Top: 0, Height: 30})

	// Add image blob so ShowImageOnTop branch is taken.
	idx := pp.BlobStore.Add("wm-img", []byte{0x89, 0x50, 0x4E, 0x47}) // PNG-like header
	pg := pp.GetPage(0)
	pg.Watermark = &preview.PreparedWatermark{
		Enabled:        true,
		Text:           "ONTOP",
		TextColor:      color.RGBA{R: 255, G: 0, B: 0, A: 200},
		TextRotation:   preview.WatermarkTextRotationHorizontal,
		ShowTextOnTop:  true,
		ShowImageOnTop: true,
		ImageBlobIdx:   idx,
		ImageSize:      preview.WatermarkImageSizeStretch,
	}

	exp := html.NewExporter()
	exp.Scale = 0 // triggers scale <= 0 branch in ExportPageEnd

	out := exportHTMLWith(t, exp, pp)

	// With clamped scale=1, page width 400 stays as 400.00px.
	if !strings.Contains(out, "400.00px") {
		t.Errorf("ZeroScale ExportPageEnd: expected 400.00px, got:\n%s", out)
	}
	// The on-top watermark text should be present.
	if !strings.Contains(out, "ONTOP") {
		t.Error("ZeroScale ExportPageEnd: expected watermark text ONTOP in output")
	}
}

// ── ExportPageEnd: ShowTextOnTop with Vertical rotation ───────────────────────

func TestExporter_ExportPageEnd_WatermarkTextOnTop_Vertical(t *testing.T) {
	wm := &preview.PreparedWatermark{
		Enabled:       true,
		Text:          "TOPTXT",
		TextColor:     color.RGBA{R: 0, G: 128, B: 0, A: 180},
		TextRotation:  preview.WatermarkTextRotationVertical,
		ShowTextOnTop: true,
		ImageBlobIdx:  -1,
	}
	pp := singlePageWithWatermark(wm)
	exp := html.NewExporter()
	out := exportHTMLWith(t, exp, pp)

	if !strings.Contains(out, "TOPTXT") {
		t.Error("ExportPageEnd ShowTextOnTop: text not found")
	}
	if !strings.Contains(out, "rotate(90deg)") {
		t.Errorf("ExportPageEnd ShowTextOnTop Vertical: expected rotate(90deg), got:\n%s", out)
	}
}
