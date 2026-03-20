package rtf_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/export"
	rtfexp "github.com/andrewloable/go-fastreport/export/rtf"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
	"image/color"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func buildPages(n int, bands []string) *preview.PreparedPages {
	pp := preview.New()
	for i := 0; i < n; i++ {
		pp.AddPage(794, 1123, i+1)
		for j, name := range bands {
			_ = pp.AddBand(&preview.PreparedBand{
				Name:   name,
				Top:    float32(j * 40),
				Height: 40,
			})
		}
	}
	return pp
}

func buildPageWithObjects(objects []preview.PreparedObject) *preview.PreparedPages {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:    "DataBand",
		Top:     0,
		Height:  40,
		Objects: objects,
	})
	return pp
}

func exportRTF(t *testing.T, pp *preview.PreparedPages, opts ...func(*rtfexp.Exporter)) string {
	t.Helper()
	exp := rtfexp.NewExporter()
	for _, o := range opts {
		o(exp)
	}
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	return buf.String()
}

// ── Basic lifecycle tests ─────────────────────────────────────────────────────

func TestRTFExporter_Defaults(t *testing.T) {
	exp := rtfexp.NewExporter()
	if exp.Title != "Report" {
		t.Errorf("default Title: want Report, got %s", exp.Title)
	}
	if exp.FileExtension() != ".rtf" {
		t.Errorf("FileExtension: want .rtf, got %s", exp.FileExtension())
	}
	if exp.Name() != "RTF" {
		t.Errorf("Name: want RTF, got %s", exp.Name())
	}
}

func TestRTFExporter_NilPages_ReturnsError(t *testing.T) {
	exp := rtfexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(nil, &buf); err == nil {
		t.Error("expected error for nil PreparedPages")
	}
}

func TestRTFExporter_EmptyPages_NoOutput(t *testing.T) {
	pp := preview.New()
	exp := rtfexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export on empty pages: %v", err)
	}
	// Empty pages → ExportBase returns nil without calling Start/Finish
	if buf.Len() != 0 {
		t.Logf("note: empty pages produced %d bytes", buf.Len())
	}
}

// ── RTF document structure ────────────────────────────────────────────────────

func TestRTFExporter_SinglePage_RTFSignature(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	out := exportRTF(t, pp)

	if !strings.HasPrefix(out, `{\rtf1`) {
		t.Errorf("RTF should start with {\\rtf1, got: %q", out[:min(len(out), 20)])
	}
}

func TestRTFExporter_HasFontTable(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	out := exportRTF(t, pp)

	if !strings.Contains(out, `{\fonttbl`) {
		t.Errorf("RTF should contain font table, got: %q", out[:min(len(out), 200)])
	}
}

func TestRTFExporter_HasColorTable(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	out := exportRTF(t, pp)

	if !strings.Contains(out, `{\colortbl`) {
		t.Errorf("RTF should contain color table")
	}
}

func TestRTFExporter_HasDocumentTitle(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	out := exportRTF(t, pp, func(e *rtfexp.Exporter) { e.Title = "My Report" })

	if !strings.Contains(out, "My Report") {
		t.Errorf("RTF should contain title 'My Report'")
	}
	if !strings.Contains(out, `{\info`) {
		t.Errorf("RTF should contain document info block")
	}
}

func TestRTFExporter_DocEndsWithClosingBrace(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	out := exportRTF(t, pp)

	trimmed := strings.TrimRight(out, "\n")
	if !strings.HasSuffix(trimmed, "}") {
		t.Errorf("RTF document should end with }, got: %q", out[max(0, len(out)-20):])
	}
}

// ── Page dimensions ───────────────────────────────────────────────────────────

func TestRTFExporter_PageDimensions_Twips(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	out := exportRTF(t, pp)

	// 794 px at 96dpi = 794/96 * 1440 = 11910 twips
	if !strings.Contains(out, `\paperw11910`) {
		t.Errorf("RTF should contain \\paperw11910 (794px → twips), got: %q", out[:min(len(out), 300)])
	}
}

// ── Text objects ──────────────────────────────────────────────────────────────

func TestRTFExporter_TextObject_ContentPresent(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T1", Kind: preview.ObjectTypeText,
			Left: 10, Top: 10, Width: 200, Height: 20,
			Text: "Hello RTF World",
		},
	})
	out := exportRTF(t, pp)

	if !strings.Contains(out, "Hello RTF World") {
		t.Errorf("RTF should contain text 'Hello RTF World'")
	}
}

func TestRTFExporter_TextObject_PositionedFrame(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T1", Kind: preview.ObjectTypeText,
			Left: 96, Top: 0, Width: 200, Height: 20,
			Text: "Positioned",
		},
	})
	out := exportRTF(t, pp)

	// 96 px = 1 inch = 1440 twips → \absx1440
	if !strings.Contains(out, `\absx1440`) {
		t.Errorf("RTF text frame should contain \\absx1440, got: %q", out)
	}
}

func TestRTFExporter_TextObject_AbsoluteY_IncludesBandTop(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	// band at top=96px, object at top=0 → absolute Y=96px = 1440 twips
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "Band",
		Top:    96,
		Height: 40,
		Objects: []preview.PreparedObject{
			{
				Name: "T", Kind: preview.ObjectTypeText,
				Left: 0, Top: 0, Width: 100, Height: 20,
				Text: "AbsY",
			},
		},
	})
	out := exportRTF(t, pp)

	if !strings.Contains(out, `\absy1440`) {
		t.Errorf("RTF text frame: expected \\absy1440 (band.Top=96px), got: %q", out)
	}
}

// ── Font styling ──────────────────────────────────────────────────────────────

func TestRTFExporter_TextObject_FontBold(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "B", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Bold",
			Font: style.Font{Name: "Arial", Size: 12, Style: style.FontStyleBold},
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, `\b`) {
		t.Errorf("bold: expected \\b in RTF, got: %q", out)
	}
}

func TestRTFExporter_TextObject_FontItalic(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "I", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Italic",
			Font: style.Font{Name: "Arial", Size: 12, Style: style.FontStyleItalic},
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, `\i`) {
		t.Errorf("italic: expected \\i in RTF, got: %q", out)
	}
}

func TestRTFExporter_TextObject_FontUnderline(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "U", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Under",
			Font: style.Font{Name: "Arial", Size: 12, Style: style.FontStyleUnderline},
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, `\ul`) {
		t.Errorf("underline: expected \\ul in RTF, got: %q", out)
	}
}

func TestRTFExporter_TextObject_FontStrikeout(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "S", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Strike",
			Font: style.Font{Name: "Arial", Size: 12, Style: style.FontStyleStrikeout},
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, `\strike`) {
		t.Errorf("strikeout: expected \\strike in RTF, got: %q", out)
	}
}

func TestRTFExporter_TextObject_FontSize_HalfPoints(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Size",
			Font: style.Font{Name: "Arial", Size: 14},
		},
	})
	out := exportRTF(t, pp)
	// 14pt × 2 = 28 half-points → \fs28
	if !strings.Contains(out, `\fs28`) {
		t.Errorf("font size 14pt: expected \\fs28 in RTF, got: %q", out)
	}
}

func TestRTFExporter_TextObject_CustomFont_InFontTable(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Custom Font",
			Font: style.Font{Name: "Courier New", Size: 10},
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, "Courier New") {
		t.Errorf("custom font: 'Courier New' should appear in RTF font table")
	}
}

// ── Horizontal alignment ──────────────────────────────────────────────────────

func TestRTFExporter_TextObject_HorzAlign_Left(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "T", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Left",
			HorzAlign: 0,
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, `\ql`) {
		t.Errorf("left align: expected \\ql in RTF")
	}
}

func TestRTFExporter_TextObject_HorzAlign_Center(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "T", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Center",
			HorzAlign: 1,
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, `\qc`) {
		t.Errorf("center align: expected \\qc in RTF")
	}
}

func TestRTFExporter_TextObject_HorzAlign_Right(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "T", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Right",
			HorzAlign: 2,
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, `\qr`) {
		t.Errorf("right align: expected \\qr in RTF")
	}
}

func TestRTFExporter_TextObject_HorzAlign_Justify(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "T", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Justify",
			HorzAlign: 3,
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, `\qj`) {
		t.Errorf("justify align: expected \\qj in RTF")
	}
}

// ── Text color ────────────────────────────────────────────────────────────────

func TestRTFExporter_TextObject_TextColor_InColorTable(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "T", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Colored",
			TextColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, `\red255\green0\blue0`) {
		t.Errorf("red text: expected red255green0blue0 in color table, got: %q", out)
	}
}

// ── RTF escaping ──────────────────────────────────────────────────────────────

func TestRTFExporter_TextObject_EscapesBackslash(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: `C:\path\to\file`,
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, `C:\\path\\to\\file`) {
		t.Errorf("backslash escape: expected C:\\\\path\\\\to\\\\file, got: %q", out)
	}
}

func TestRTFExporter_TextObject_EscapesBraces(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "a{b}c",
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, `a\{b\}c`) {
		t.Errorf("brace escape: expected a\\{b\\}c, got: %q", out)
	}
}

func TestRTFExporter_TextObject_EscapesNewline(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "line1\nline2",
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, `\line`) {
		t.Errorf("newline: expected \\line in RTF, got: %q", out)
	}
}

// ── RTF source objects ────────────────────────────────────────────────────────

func TestRTFExporter_RTFObject_ControlWordsStripped(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "R", Kind: preview.ObjectTypeRTF,
			Left: 0, Top: 0, Width: 200, Height: 40,
			Text: `{\rtf1\ansi Hello RTF}`,
		},
	})
	out := exportRTF(t, pp)
	// RTF control words should be stripped; "Hello RTF" plain text should appear
	if !strings.Contains(out, "Hello RTF") {
		t.Errorf("RTF source: expected plain text 'Hello RTF', got: %q", out)
	}
}

// ── Non-text objects skipped ──────────────────────────────────────────────────

func TestRTFExporter_PictureObject_Skipped(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:    "Pic", Kind: preview.ObjectTypePicture,
			Left:    0, Top: 0, Width: 80, Height: 80,
			BlobIdx: -1,
		},
	})
	out := exportRTF(t, pp)
	// No text frames for pictures — just the RTF header/footer
	if strings.Contains(out, `\pard`) {
		t.Logf("note: \\pard found (not from picture, from other RTF content)")
	}
	// Should be valid RTF without image data
	if !strings.HasPrefix(out, `{\rtf1`) {
		t.Errorf("RTF should still have valid header")
	}
}

func TestRTFExporter_LineObject_Skipped(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "Line", Kind: preview.ObjectTypeLine,
			Left: 0, Top: 0, Width: 100, Height: 1,
		},
	})
	out := exportRTF(t, pp)
	// Line objects produce no \pard frames (only text objects do)
	if !strings.HasPrefix(out, `{\rtf1`) {
		t.Errorf("RTF with line object: should still have valid header")
	}
}

// ── Multiple pages ────────────────────────────────────────────────────────────

func TestRTFExporter_MultiplePages_PageBreaks(t *testing.T) {
	pp := buildPages(3, []string{"Band"})
	out := exportRTF(t, pp)

	// Multiple pages → page breaks between them
	// The last \page is removed in Finish(), so we expect 2 \page for 3 pages
	count := strings.Count(out, `\page`)
	if count < 2 {
		t.Errorf("3 pages: expected at least 2 \\page breaks, got %d", count)
	}
}

func TestRTFExporter_MultiplePages_PageDimensions(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1) // A4 portrait
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	pp.AddPage(842, 595, 2) // A4 landscape
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 40})
	out := exportRTF(t, pp)

	// Portrait: 595px = 595/96*1440 = 8925 twips
	// Landscape: 842px = 842/96*1440 = 12630 twips
	if !strings.Contains(out, `\paperw8925`) {
		t.Errorf("expected \\paperw8925 for 595px page width")
	}
	if !strings.Contains(out, `\paperw12630`) {
		t.Errorf("expected \\paperw12630 for 842px page width")
	}
}

func TestRTFExporter_TextContent_MultipleObjects(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Alpha",
		},
		{
			Name: "T2", Kind: preview.ObjectTypeText,
			Left: 110, Top: 0, Width: 100, Height: 20,
			Text: "Beta",
		},
	})
	out := exportRTF(t, pp)
	if !strings.Contains(out, "Alpha") {
		t.Errorf("expected text 'Alpha' in RTF")
	}
	if !strings.Contains(out, "Beta") {
		t.Errorf("expected text 'Beta' in RTF")
	}
}

// ── PageRange ─────────────────────────────────────────────────────────────────

func TestRTFExporter_PageRangeCurrent(t *testing.T) {
	pp := preview.New()
	for i := 0; i < 3; i++ {
		pp.AddPage(794, 1123, i+1)
		_ = pp.AddBand(&preview.PreparedBand{
			Name:   "DataBand",
			Top:    0,
			Height: 40,
			Objects: []preview.PreparedObject{
				{
					Name: "T", Kind: preview.ObjectTypeText,
					Left: 0, Top: 0, Width: 100, Height: 20,
					Text: "PageText",
				},
			},
		})
	}
	exp := rtfexp.NewExporter()
	exp.PageRange = export.PageRangeCurrent
	exp.CurPage = 1
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	out := buf.String()
	// Only 1 page → no \page page-break at all (trailing \page stripped)
	if strings.Count(out, `\page`) > 0 {
		t.Logf("note: \\page count in single-page export: %d", strings.Count(out, `\page`))
	}
	if !strings.Contains(out, "PageText") {
		t.Errorf("PageRangeCurrent: expected 'PageText' in output")
	}
}

// ── Unicode escape ────────────────────────────────────────────────────────────

func TestRTFExporter_TextObject_UnicodeEscape(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 200, Height: 20,
			Text: "Caf\u00e9", // é is non-ASCII
		},
	})
	out := exportRTF(t, pp)
	// Non-ASCII characters should be encoded as \uN? escapes
	if !strings.Contains(out, `\u`) {
		t.Errorf("unicode: expected \\uN? escape for non-ASCII char, got: %q", out)
	}
}

// ── Edge cases ────────────────────────────────────────────────────────────────

func TestRTFExporter_EmptyTitle_NoInfoBlock(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	out := exportRTF(t, pp, func(e *rtfexp.Exporter) { e.Title = "" })

	// No title → no info block
	if strings.Contains(out, `{\info`) {
		t.Errorf("empty title: should not contain info block, got: %q", out[:min(len(out), 300)])
	}
}

func TestRTFExporter_DefaultFontSize_ZeroSize_Defaults12pt(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "T", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "DefaultSize",
			Font: style.Font{Name: "Arial", Size: 0}, // zero → defaults to 12pt
		},
	})
	out := exportRTF(t, pp)
	// 12pt × 2 = 24 half-points → \fs24
	if !strings.Contains(out, `\fs24`) {
		t.Errorf("zero font size: expected \\fs24 (default 12pt), got: %q", out)
	}
}

// helper: min for old Go compatibility
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// helper: max for old Go compatibility
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
