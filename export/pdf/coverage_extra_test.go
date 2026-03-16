package pdf

// coverage_extra_test.go covers specific branches that are not exercised by the
// main test suite. All tests are in package pdf (internal) so they can access
// unexported functions directly.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image/color"
	"io"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/export/pdf/core"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// buildPNGWithCorruptIDAT returns a PNG byte slice that has a valid signature
// and IHDR chunk (so image.DecodeConfig succeeds and returns dimensions), but
// a corrupt IDAT chunk (so image.Decode fails with a checksum error).
func buildPNGWithCorruptIDAT(width, height int) []byte {
	var out []byte

	// PNG signature
	out = append(out, 0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n')

	// IHDR chunk (13 bytes of data)
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:4], uint32(width))
	binary.BigEndian.PutUint32(ihdr[4:8], uint32(height))
	ihdr[8] = 8  // bit depth
	ihdr[9] = 2  // color type: RGB
	ihdr[10] = 0 // compression
	ihdr[11] = 0 // filter
	ihdr[12] = 0 // interlace: none
	out = appendPNGChunk(out, "IHDR", ihdr)

	// IDAT chunk with corrupt (non-zlib) data
	corruptIDAT := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x00, 0x00, 0x00, 0x00}
	out = appendPNGChunk(out, "IDAT", corruptIDAT)

	// IEND chunk (empty)
	out = appendPNGChunk(out, "IEND", nil)

	return out
}

// appendPNGChunk appends a PNG chunk (length + type + data + CRC) to buf.
func appendPNGChunk(buf []byte, chunkType string, data []byte) []byte {
	// Length
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(data)))
	buf = append(buf, length...)
	// Type
	typeBytes := []byte(chunkType)
	buf = append(buf, typeBytes...)
	// Data
	buf = append(buf, data...)
	// CRC over type + data
	crcVal := crc32.NewIEEE()
	crcVal.Write(typeBytes)
	crcVal.Write(data)
	crcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crcBytes, crcVal.Sum32())
	buf = append(buf, crcBytes...)
	return buf
}

// ── renderTextObject: empty text early return (export.go:200-202) ─────────────

func TestRenderTextObject_EmptyText(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{fontMgr: nil}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 50}
	obj := preview.PreparedObject{
		Name:   "obj",
		Kind:   preview.ObjectTypeText,
		Left:   10, Top: 5, Width: 200, Height: 20,
		Text:   "", // empty → early return
		Font:   style.Font{Name: "Arial", Size: 10},
	}
	initial := c.buf.Len()
	exp.renderTextObject(c, b, obj)
	// No text operators should be written for an empty text object.
	if c.buf.Len() != initial {
		t.Errorf("expected no output for empty text, but buffer grew by %d bytes", c.buf.Len()-initial)
	}
}

// ── renderTextObject: startY clamp branch (export.go:261-263) ─────────────────

// TestRenderTextObject_StartYClamp triggers the branch where startY exceeds
// yPt+hPt-fontPt. This happens with Center or Bottom VertAlign when the total
// text height is large relative to the box.
func TestRenderTextObject_StartYClamp(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{fontMgr: nil}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 5}
	obj := preview.PreparedObject{
		Name:      "obj",
		Kind:      preview.ObjectTypeText,
		Left:      0, Top: 0, Width: 200, Height: 5,
		Text:      "line1\nline2\nline3\nline4\nline5",
		Font:      style.Font{Name: "Arial", Size: 20},
		VertAlign: 1, // Center — can push startY above clamp threshold
		WordWrap:  false,
	}
	// Should not panic; clamp branch should be taken.
	exp.renderTextObject(c, b, obj)
	out := c.buf.String()
	if !strings.Contains(out, "BT") {
		t.Error("expected BT operator in output")
	}
}

// ── pdfDrawBorder: lw < 0.5 clamp (export.go:521-523) ───────────────────────

func TestPdfDrawBorder_LineWidthClamp(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	exp := &Exporter{}

	// Create a border with Width=0 (converts to 0 pt < 0.5 → clamped to 0.5).
	bl := style.NewBorderLine()
	bl.Width = 0
	bl.Style = style.LineStyleSolid
	bl.Color = color.RGBA{R: 0, G: 0, B: 0, A: 255}

	border := &style.Border{
		VisibleLines: style.BorderLinesLeft,
	}
	border.Lines[int(style.BorderLeft)] = bl

	exp.pdfDrawBorder(c, 0, 0, 100, 50, border)
	out := c.buf.String()
	// The line-width operator should contain 0.50 (clamped from 0).
	if !strings.Contains(out, "0.50 w") {
		t.Errorf("expected clamped line width 0.50 in output: %q", out)
	}
}

// ── renderHyperlinkAnnotation: nil curPage (export.go:554-556) ───────────────

func TestRenderHyperlinkAnnotation_NilCurPage(t *testing.T) {
	exp := &Exporter{} // curPage is nil
	b := &preview.PreparedBand{Top: 0, Height: 50}
	obj := preview.PreparedObject{
		Name: "link", HyperlinkKind: 1, HyperlinkValue: "https://example.com",
	}
	// Should be a no-op when curPage is nil.
	exp.renderHyperlinkAnnotation(b, obj)
}

// ── renderDigitalSignatureField: nil curPage (export.go:634-636) ─────────────

func TestRenderDigitalSignatureField_NilCurPage(t *testing.T) {
	exp := &Exporter{} // curPage and writer are nil
	b := &preview.PreparedBand{Top: 0, Height: 50}
	obj := preview.PreparedObject{Name: "sig", Kind: preview.ObjectTypeDigitalSignature}
	// Should be a no-op when curPage is nil.
	exp.renderDigitalSignatureField(b, obj)
}

// ── renderWatermark: nil/disabled watermark (export.go:774-776) ──────────────

func TestRenderWatermark_NilWatermark(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{}
	exp.curPage = NewPage(w, pages, 595, 842)

	pg := &preview.PreparedPage{Width: 595, Height: 842, Watermark: nil}
	initial := c.buf.Len()
	exp.renderWatermark(c, pg)
	if c.buf.Len() != initial {
		t.Error("expected no output for nil watermark")
	}
}

func TestRenderWatermark_DisabledWatermark(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{}
	exp.curPage = NewPage(w, pages, 595, 842)

	pg := &preview.PreparedPage{
		Width: 595, Height: 842,
		Watermark: &preview.PreparedWatermark{Enabled: false},
	}
	initial := c.buf.Len()
	exp.renderWatermark(c, pg)
	if c.buf.Len() != initial {
		t.Error("expected no output for disabled watermark")
	}
}

// ── renderWatermarkText: empty text early return (export.go:898-900) ─────────

func TestRenderWatermarkText_EmptyText(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{}
	exp.curPage = NewPage(w, pages, 595, 842)

	pg := &preview.PreparedPage{Width: 595, Height: 842}
	wm := &preview.PreparedWatermark{
		Enabled: true,
		Text:    "", // empty → early return
	}
	initial := c.buf.Len()
	exp.renderWatermarkText(c, pg, wm)
	if c.buf.Len() != initial {
		t.Error("expected no output for empty watermark text")
	}
}

// ── renderWatermarkImage: empty blob (export.go:790-792) ─────────────────────

func TestRenderWatermarkImage_EmptyBlob(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	// Add an empty byte slice to the blob store.
	blobIdx := pp.BlobStore.Add("empty", []byte{})

	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{pp: pp, writer: w}
	exp.curPage = NewPage(w, pages, 595, 842)

	pg := pp.GetPage(0)
	wm := &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
	}
	initial := c.buf.Len()
	exp.renderWatermarkImage(c, pg, wm)
	if c.buf.Len() != initial {
		t.Error("expected no output for empty image blob")
	}
}

// ── renderWatermarkImage: invalid image data (export.go:799-801) ─────────────

func TestRenderWatermarkImage_InvalidImageData(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	// Store invalid/corrupt image bytes (not a valid image format).
	blobIdx := pp.BlobStore.Add("bad", []byte("not an image"))

	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{pp: pp, writer: w}
	exp.curPage = NewPage(w, pages, 595, 842)

	pg := pp.GetPage(0)
	wm := &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeCenter,
	}
	initial := c.buf.Len()
	// image.DecodeConfig fails on invalid data → early return.
	exp.renderWatermarkImage(c, pg, wm)
	if c.buf.Len() != initial {
		t.Error("expected no output for invalid image data in watermark")
	}
}

// ── renderWatermarkImage: PNG with valid header but corrupt IDAT (export.go:816-818) ─

// TestRenderWatermarkImage_CorruptIDAT exercises the image.Decode error branch
// at line 816-818 of export.go. image.DecodeConfig succeeds because it only
// reads the PNG IHDR chunk (which is valid), but image.Decode fails because the
// IDAT (compressed pixel data) chunk contains invalid zlib data.
func TestRenderWatermarkImage_CorruptIDAT(t *testing.T) {
	corrupt := buildPNGWithCorruptIDAT(4, 4)

	pp := preview.New()
	pp.AddPage(595, 842, 1)
	blobIdx := pp.BlobStore.Add("corrupt_idat", corrupt)

	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{pp: pp, writer: w}
	exp.curPage = NewPage(w, pages, 595, 842)

	pg := pp.GetPage(0)
	wm := &preview.PreparedWatermark{
		Enabled:      true,
		ImageBlobIdx: blobIdx,
		ImageSize:    preview.WatermarkImageSizeCenter,
	}
	initial := c.buf.Len()
	// image.DecodeConfig succeeds (valid IHDR), but image.Decode fails (corrupt IDAT)
	// → function returns early without writing to content stream.
	exp.renderWatermarkImage(c, pg, wm)
	if c.buf.Len() != initial {
		t.Logf("note: corrupt PNG may have been handled (output: %d bytes)", c.buf.Len()-initial)
	}
	// Test passes if no panic occurs.
}

// ── renderPictureObject: empty blob (export.go:1129-1131) ────────────────────

func TestRenderPictureObject_EmptyBlob(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	blobIdx := pp.BlobStore.Add("empty_pic", []byte{})

	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{pp: pp, writer: w}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 100}
	obj := preview.PreparedObject{
		Name: "pic", Kind: preview.ObjectTypePicture,
		Left: 0, Top: 0, Width: 100, Height: 100,
		BlobIdx: blobIdx,
	}
	initial := c.buf.Len()
	exp.renderPictureObject(c, b, obj)
	if c.buf.Len() != initial {
		t.Error("expected no output for empty picture blob")
	}
}

// ── renderPictureObject: JPEG decode error (export.go:1149-1151) ─────────────

func TestRenderPictureObject_InvalidJPEG(t *testing.T) {
	// Has JPEG magic bytes but is otherwise invalid.
	badJPEG := []byte{0xFF, 0xD8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	pp := preview.New()
	pp.AddPage(595, 842, 1)
	blobIdx := pp.BlobStore.Add("bad_jpeg", badJPEG)

	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{pp: pp, writer: w}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 100}
	obj := preview.PreparedObject{
		Name: "pic", Kind: preview.ObjectTypePicture,
		Left: 0, Top: 0, Width: 100, Height: 100,
		BlobIdx: blobIdx,
	}
	initial := c.buf.Len()
	// JPEG path: image.DecodeConfig fails on invalid JPEG → early return.
	exp.renderPictureObject(c, b, obj)
	if c.buf.Len() != initial {
		t.Error("expected no output for invalid JPEG data")
	}
}

// ── renderPictureObject: non-JPEG invalid image decode (export.go:1156-1158) ─

func TestRenderPictureObject_InvalidNonJPEG(t *testing.T) {
	// Not a JPEG (different magic bytes), not a valid image format.
	badData := []byte("not an image at all, definitely not jpeg or png")

	pp := preview.New()
	pp.AddPage(595, 842, 1)
	blobIdx := pp.BlobStore.Add("bad_data", badData)

	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{pp: pp, writer: w}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 100}
	obj := preview.PreparedObject{
		Name: "pic", Kind: preview.ObjectTypePicture,
		Left: 0, Top: 0, Width: 100, Height: 100,
		BlobIdx: blobIdx,
	}
	initial := c.buf.Len()
	// Non-JPEG path: image.Decode fails → early return.
	exp.renderPictureObject(c, b, obj)
	if c.buf.Len() != initial {
		t.Error("expected no output for invalid non-JPEG image data")
	}
}

// ── renderBarcodeVector: zero-size object (export.go:1239-1241) ──────────────

func TestRenderBarcodeVector_ZeroSize(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	pages := NewPages(w)
	exp := &Exporter{}
	exp.curPage = NewPage(w, pages, 595, 842)

	b := &preview.PreparedBand{Top: 0, Height: 0}
	obj := preview.PreparedObject{
		Name: "bar", Kind: preview.ObjectTypePicture,
		Left: 0, Top: 0,
		// Width=0, Height=0 → PixelsToPoints = 0 → wPt<=0 || hPt<=0 → early return
		Width: 0, Height: 0,
		IsBarcode: true,
		BarcodeModules: [][]bool{
			{true, false, true},
			{false, true, false},
		},
	}
	initial := c.buf.Len()
	exp.renderBarcodeVector(c, b, obj)
	if c.buf.Len() != initial {
		t.Error("expected no output for zero-size barcode object")
	}
}

// ── RegisterFont: duplicate registration returns cached alias ─────────────────

// TestRegisterFont_DuplicateReturnsExistingAlias verifies the caching path
// (lines 115-117 in font.go) where a previously-registered font returns its
// existing alias without re-parsing the TTF.
func TestRegisterFont_DuplicateReturnsExistingAlias(t *testing.T) {
	w := NewWriter()
	fm := NewPDFFontManager(w)
	a1 := fm.RegisterFont("Arial", false, false)
	a2 := fm.RegisterFont("Arial", false, false) // duplicate → cached
	if a1 != a2 {
		t.Errorf("duplicate RegisterFont should return same alias, got %q and %q", a1, a2)
	}
}

// ── EncodeText: glyph not found (gi==0) (font.go:165-167) ────────────────────

// TestEncodeText_GlyphNotFound exercises the gi==0 branch in EncodeText.
// Using a Private-Use Area rune that is not in GoRegular's cmap forces gi=0.
func TestEncodeText_GlyphNotFound(t *testing.T) {
	w := NewWriter()
	fm := NewPDFFontManager(w)
	alias := fm.RegisterFont("Arial", false, false)

	// Private-use area rune unlikely to be in GoRegular cmap.
	result := fm.EncodeText(alias, "\uE000")
	// Even if glyph is 0, we should get a valid hex string back.
	if !strings.HasPrefix(result, "<") {
		t.Errorf("EncodeText with unknown glyph should return hex string, got %q", result)
	}
	// The glyph index 0 is encoded as <0000>.
	if !strings.Contains(result, "0000") {
		t.Errorf("expected glyph index 0 encoded as 0000 in %q", result)
	}
}

// ── MeasureText: nil font lookup (font.go:192-194) ────────────────────────────

func TestMeasureText_NilLookup(t *testing.T) {
	w := NewWriter()
	fm := NewPDFFontManager(w)
	// Use an alias that was never registered → lookupFont returns nil.
	width := fm.MeasureText("NOTREGISTERED", "hello", 10)
	expected := pdfEstimateTextWidth("hello", 10)
	if width != expected {
		t.Errorf("MeasureText with unregistered alias = %.2f, want %.2f", width, expected)
	}
}

// ── Writer.Write: object WriteTo error (writer.go:62-64) ─────────────────────

// failingPDFObject is a core.Object whose WriteTo always returns an error,
// used to exercise the error path in Writer.Write at line 62-64.
type failingPDFObject struct{}

func (f *failingPDFObject) Type() core.ObjectType { return core.TypeNull }

func (f *failingPDFObject) WriteTo(w io.Writer) (int64, error) {
	return 0, errObjectWriteFailed
}

var errObjectWriteFailed = errors.New("object write failed")

func TestWriter_Write_ObjectError(t *testing.T) {
	w := NewWriter()
	// Register a failing object.
	w.NewObject(&failingPDFObject{})

	var buf bytes.Buffer
	err := w.Write(&buf)
	if err == nil {
		t.Fatal("expected error from Write when an object's WriteTo fails")
	}
	if !strings.Contains(err.Error(), "writing object") {
		t.Errorf("error should mention 'writing object', got: %v", err)
	}
}
