package image_test

// image_htmltext_test.go — tests targeting renderHtmlText (0% coverage) and
// remaining uncovered branches in renderObject and drawPictureObject.

import (
	"bytes"
	"image/color"
	"testing"

	imgexport "github.com/andrewloable/go-fastreport/export/image"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// ── renderHtmlText via TextRenderType=1 (HtmlTags) ──────────────────────────

// TestRenderHtmlText_Basic exercises the basic renderHtmlText path via
// TextRenderType=1 (textRenderTypeHtmlTags). This is the only way to reach
// renderHtmlText from the public API.
func TestRenderHtmlText_Basic(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "ht1", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 120, Height: 40,
			Text:           "<b>Bold text</b> and <i>italic</i>",
			Font:           style.Font{Size: 10},
			TextRenderType: 1, // textRenderTypeHtmlTags
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_HtmlParagraph exercises TextRenderType=2 (HtmlParagraph).
func TestRenderHtmlText_HtmlParagraph(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "ht2", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 120, Height: 40,
			Text:           "<b>Paragraph</b> mode",
			Font:           style.Font{Size: 10},
			TextRenderType: 2, // textRenderTypeHtmlParagraph
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_VertAlignCenter exercises the VertAlign=1 (center) branch
// inside renderHtmlText.
func TestRenderHtmlText_VertAlignCenter(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htvc", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 120, Height: 60,
			Text:           "<b>Centered</b>",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
			VertAlign:      1, // center
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_VertAlignBottom exercises the VertAlign=2 (bottom) branch.
func TestRenderHtmlText_VertAlignBottom(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htvb", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 120, Height: 60,
			Text:           "<b>Bottom</b>",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
			VertAlign:      2, // bottom
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_HorzAlignCenter exercises the HorzAlign=1 (center) branch
// in the run-drawing loop of renderHtmlText.
func TestRenderHtmlText_HorzAlignCenter(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "hthc", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 120, Height: 40,
			Text:           "<b>Center aligned</b>",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
			HorzAlign:      1, // center
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_HorzAlignRight exercises the HorzAlign=2 (right) branch.
func TestRenderHtmlText_HorzAlignRight(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "hthr", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 120, Height: 40,
			Text:           "<b>Right aligned</b>",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
			HorzAlign:      2, // right
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_MultiLine exercises the multi-line HTML path, including
// the "empty plain line" branch (totalVisualLines++ when plain=="").
// A <br> tag in HTML creates an explicit empty line.
func TestRenderHtmlText_MultiLine(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "html", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 150, Height: 80,
			Text:           "First line<br/>Second line<br/>Third line",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_EmptyText exercises the early-return branch when the
// HTML produces zero lines (empty string).
func TestRenderHtmlText_EmptyText(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htempty", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 120, Height: 40,
			Text:           "",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_BoldRunWithFontSize exercises the run.Font.Size > 0 branch
// (runs with explicit font sizes get their own face).
func TestRenderHtmlText_BoldRunWithFontSize(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htbs", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 150, Height: 40,
			Text:           `<font size="14">Large</font> and <font size="8">small</font>`,
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_OverflowsHeight exercises the "startY > y+h → break" branch
// by rendering many lines into a very short box.
func TestRenderHtmlText_OverflowsHeight(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htof", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 150, Height: 10, // very short
			Text:           "Line1<br/>Line2<br/>Line3<br/>Line4<br/>Line5<br/>Line6",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_RightAlignOverflow exercises the "dotX < x → dotX = x"
// clamp by using a very wide run with right alignment in a narrow box.
func TestRenderHtmlText_RightAlignOverflow(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htrof", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 20, Height: 40, // narrow box
			Text:           "<b>VeryVeryLongTextThatOverflows</b>",
			Font:           style.Font{Size: 12},
			TextRenderType: 1,
			HorzAlign:      2, // right → dotX may go negative → clamped to x
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_RunsExceedWidth exercises the "dotX >= x+w → break" inside
// the per-run draw loop by rendering a long text with many runs.
func TestRenderHtmlText_RunsExceedWidth(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htrew", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 30, Height: 40, // very narrow
			Text:           "<b>Word1</b> <i>Word2</i> Word3 <b>Word4</b> Word5",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_ZeroFontSize exercises the "run.Font.Size <= 0 → use baseFace"
// branch. The base font has size=0 which defaults to 10pt inside the exporter.
func TestRenderHtmlText_ZeroFontSize(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htzfs", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 150, Height: 40,
			Text:           "<b>Zero size</b> base font",
			Font:           style.Font{Size: 0}, // triggers fontPt=10 fallback
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_HtmlObjectKind exercises renderHtmlText via ObjectTypeHtml
// combined with TextRenderType=1. ObjectTypeHtml falls into the
// ObjectTypeText|ObjectTypeHtml case in the switch, then the TextRenderType
// check routes to renderHtmlText.
func TestRenderHtmlText_HtmlObjectKind(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "hthtml", Kind: preview.ObjectTypeHtml,
			Left: 5, Top: 5, Width: 150, Height: 40,
			Text:           "<b>Html kind</b> with render type",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
			FillColor:      color.RGBA{R: 240, G: 248, B: 255, A: 255},
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_WithBoldItalicRuns tests mixed bold+italic HTML runs.
func TestRenderHtmlText_WithBoldItalicRuns(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htbi", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 150, Height: 40,
			Text:           "Normal <b>bold</b> <i>italic</i> <b><i>bold-italic</i></b>",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_TextColor exercises the non-zero TextColor path in
// renderHtmlText (tc.A != 0 → use obj.TextColor as default).
func TestRenderHtmlText_TextColor(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "httc", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 150, Height: 40,
			Text:           "<b>Colored</b> text",
			Font:           style.Font{Size: 10},
			TextColor:      color.RGBA{R: 200, G: 50, B: 50, A: 255},
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_EmptyHtmlLinesProducesNoOutput exercises the case where
// htmlLines has zero elements (empty HTML string → early return from renderHtmlText).
// This is separate from empty Text: even a whitespace-only Text produces no runs.
func TestRenderHtmlText_WhitespaceOnly(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htws", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 150, Height: 40,
			Text:           "   ",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_OutOfBounds exercises the bounds.Empty() early-return
// for an HTML text object placed outside the page.
func TestRenderHtmlText_OutOfBounds(t *testing.T) {
	pp := preview.New()
	pp.AddPage(100, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "Band1", Top: 0, Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name: "htoob", Kind: preview.ObjectTypeText,
				Left: 500, Top: 500, Width: 80, Height: 30,
				Text:           "<b>Out of bounds</b>",
				Font:           style.Font{Size: 10},
				TextRenderType: 1,
			},
		},
	})
	exportOK(t, pp)
}

// ── renderObject: ObjectTypeDigitalSignature ──────────────────────────────────

// TestRenderObject_DigitalSignature_WithText exercises the digital signature
// rendering path in renderObject, including the dashed border and centered text.
func TestRenderObject_DigitalSignature_WithText(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "ds1", Kind: preview.ObjectTypeDigitalSignature,
			Left: 10, Top: 10, Width: 120, Height: 50,
			Text:      "Sign Here",
			Font:      style.Font{Size: 10},
			TextColor: color.RGBA{A: 255},
		},
	})
	exportOK(t, pp)
}

// TestRenderObject_DigitalSignature_EmptyText exercises the default
// "Digital Signature" label when Text is empty.
func TestRenderObject_DigitalSignature_EmptyText(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "ds2", Kind: preview.ObjectTypeDigitalSignature,
			Left: 10, Top: 10, Width: 120, Height: 50,
			Text: "", // → default label
			Font: style.Font{Size: 10},
		},
	})
	exportOK(t, pp)
}

// TestRenderObject_DigitalSignature_ZeroFontSize exercises the fontPt=10 fallback
// when Font.Size == 0 for a digital signature object.
func TestRenderObject_DigitalSignature_ZeroFontSize(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "ds3", Kind: preview.ObjectTypeDigitalSignature,
			Left: 10, Top: 10, Width: 120, Height: 50,
			Text: "Zero Size",
			Font: style.Font{Size: 0},
		},
	})
	exportOK(t, pp)
}

// TestRenderObject_DigitalSignature_ZeroAlphaTextColor exercises the tc.A==0
// fallback to black in the digital signature text rendering branch.
func TestRenderObject_DigitalSignature_ZeroAlphaTextColor(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name:      "ds4",
			Kind:      preview.ObjectTypeDigitalSignature,
			Left:      10,
			Top:       10,
			Width:     120,
			Height:    50,
			Text:      "Black text",
			Font:      style.Font{Size: 10},
			TextColor: color.RGBA{A: 0}, // → fallback to black
		},
	})
	exportOK(t, pp)
}

// TestRenderObject_DigitalSignature_OutOfBounds exercises the bounds.Empty()
// early-return for a digital signature placed outside the page.
func TestRenderObject_DigitalSignature_OutOfBounds(t *testing.T) {
	pp := preview.New()
	pp.AddPage(100, 100, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "Band1", Top: 0, Height: 100,
		Objects: []preview.PreparedObject{
			{
				Name: "dsoob", Kind: preview.ObjectTypeDigitalSignature,
				Left: 500, Top: 500, Width: 100, Height: 50,
				Text: "OOB",
			},
		},
	})
	exportOK(t, pp)
}

// TestRenderObject_DigitalSignature_SmallPage exercises the digital signature
// on a small page to ensure dashed border edge handling works correctly.
func TestRenderObject_DigitalSignature_SmallPage(t *testing.T) {
	pp := preview.New()
	pp.AddPage(60, 40, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "Band1", Top: 0, Height: 40,
		Objects: []preview.PreparedObject{
			{
				Name: "dssmall", Kind: preview.ObjectTypeDigitalSignature,
				Left: 2, Top: 2, Width: 55, Height: 35,
				Text: "Sign",
				Font: style.Font{Size: 8},
			},
		},
	})
	exportOK(t, pp)
}

// ── renderObject: ObjectTypeBarcode ──────────────────────────────────────────

// TestRenderObject_Barcode_NoModules exercises the ObjectTypeBarcode case (or the
// default fallthrough if barcode has no dedicated case). Either way, the code
// path in renderObject for a barcode object must run without panicking.
func TestRenderObject_Barcode_NoModules(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "bc1", Kind: preview.ObjectTypeBarcode,
			Left: 5, Top: 5, Width: 80, Height: 40,
			IsBarcode: true,
			// BarcodeModules is nil → no vector rendering
		},
	})
	exportOK(t, pp)
}

func TestRenderObject_Barcode_WithModules(t *testing.T) {
	// Provide a simple 3x3 barcode module matrix.
	modules := [][]bool{
		{true, false, true},
		{false, true, false},
		{true, false, true},
	}
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "bc2", Kind: preview.ObjectTypeBarcode,
			Left: 5, Top: 5, Width: 60, Height: 60,
			IsBarcode:      true,
			BarcodeModules: modules,
		},
	})
	exportOK(t, pp)
}

// ── drawPictureObject: remaining uncovered branches ───────────────────────────

// TestDrawPictureObject_NilPPViaPublicAPI exercises the pp==nil early-return
// path of drawPictureObject indirectly: create a picture object with a valid
// BlobIdx but no PreparedPages set on the exporter. Since Export sets e.pp,
// we must bypass it; the nil-pp case is fully exercised in the internal tests,
// so this test just confirms the public Export path doesn't panic on BlobIdx=-1.
func TestDrawPictureObject_BlobIdxNegativeViaExport(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "picN", Kind: preview.ObjectTypePicture,
			Left: 5, Top: 5, Width: 60, Height: 40,
			BlobIdx: -1,
		},
	})
	exp := imgexport.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
}

// TestRenderHtmlText_FillColorTransparent exercises the transparent fill branch
// (FillColor.A == 0) for an HTML text object, which draws white background.
func TestRenderHtmlText_FillColorTransparent(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "httrans", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 120, Height: 40,
			Text:           "<b>Transparent fill</b>",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
			FillColor:      color.RGBA{A: 0}, // transparent → white fill branch
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_MonoFont exercises selectFace within renderHtmlText using
// a monospace base font.
func TestRenderHtmlText_MonoFont(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htmono", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 150, Height: 40,
			Text:           "<b>Mono bold</b> text",
			Font:           style.Font{Size: 10, Name: "Courier New"},
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_RTFWithHtmlRenderType exercises an RTF object that also has
// TextRenderType=1. RTF stripping happens first in renderObject, then the text
// goes through the HTML text path.
func TestRenderHtmlText_RTFWithHtmlRenderType(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "rtfht", Kind: preview.ObjectTypeRTF,
			Left: 5, Top: 5, Width: 150, Height: 40,
			Text:           `{\rtf1\ansi Hello world}`,
			Font:           style.Font{Size: 10},
			TextRenderType: 1, // after RTF stripping, text goes to renderHtmlText
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_ConsecutiveBreaks exercises the "empty HtmlLine" path where
// two consecutive <br/> tags produce a line with no runs. This covers:
//   - plain == "" branch (line 458: totalVisualLines++ for blank line)
//   - len(renderRuns) == 0 branch (line 515: empty line → advance Y and continue)
func TestRenderHtmlText_ConsecutiveBreaks(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htcb", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 150, Height: 80,
			// Two consecutive <br/> produce an empty middle line.
			Text:           "Line one<br/><br/>Line three",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_LeadingBreak exercises the case where the HTML starts with
// a <br/> so the first HtmlLine is completely empty (no runs at all).
func TestRenderHtmlText_LeadingBreak(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htlb", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 150, Height: 60,
			Text:           "<br/>Content after break",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}

// TestRenderHtmlText_OnlyBreaks exercises the edge case of only <br/> tags,
// producing multiple empty lines.
func TestRenderHtmlText_OnlyBreaks(t *testing.T) {
	pp := pageWithObjectBand([]preview.PreparedObject{
		{
			Name: "htob", Kind: preview.ObjectTypeText,
			Left: 5, Top: 5, Width: 150, Height: 60,
			Text:           "<br/><br/><br/>",
			Font:           style.Font{Size: 10},
			TextRenderType: 1,
		},
	})
	exportOK(t, pp)
}
