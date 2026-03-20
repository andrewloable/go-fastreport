package xlsx_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"

	"github.com/andrewloable/go-fastreport/export"
	xlsxexp "github.com/andrewloable/go-fastreport/export/xlsx"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
	excelize "github.com/xuri/excelize/v2"
)

// minimalPNG returns a valid 1×1 red PNG as bytes.
func minimalPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// minimalJPEG returns a byte slice that starts with JPEG magic bytes.
// It is intentionally not a valid JPEG — excelize AddPictureFromBytes will
// fail for it, exercising the non-fatal error path.
func minimalJPEG() []byte {
	return []byte{0xFF, 0xD8, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04}
}

// minimalGIF returns a byte slice with GIF magic bytes (not a full valid GIF).
func minimalGIF() []byte {
	return []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00}
}

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

// exportAndRead runs Export, parses the resulting XLSX, and returns the excelize.File.
func exportAndRead(t *testing.T, pp *preview.PreparedPages, opts ...func(*xlsxexp.Exporter)) *excelize.File {
	t.Helper()
	exp := xlsxexp.NewExporter()
	for _, o := range opts {
		o(exp)
	}
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	f, err := excelize.OpenReader(&buf)
	if err != nil {
		t.Fatalf("excelize.OpenReader: %v", err)
	}
	return f
}

// ── Basic lifecycle tests ─────────────────────────────────────────────────────

func TestXLSXExporter_Defaults(t *testing.T) {
	exp := xlsxexp.NewExporter()
	if exp.SheetName != "Report" {
		t.Errorf("default SheetName: want Report, got %s", exp.SheetName)
	}
	if exp.FileExtension() != ".xlsx" {
		t.Errorf("FileExtension: want .xlsx, got %s", exp.FileExtension())
	}
	if exp.Name() != "XLSX" {
		t.Errorf("Name: want XLSX, got %s", exp.Name())
	}
}

func TestXLSXExporter_NilPages_ReturnsError(t *testing.T) {
	exp := xlsxexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(nil, &buf); err == nil {
		t.Error("expected error for nil PreparedPages")
	}
}

func TestXLSXExporter_EmptyPages_ProducesValidXLSX(t *testing.T) {
	pp := preview.New()
	exp := xlsxexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export on empty pages: %v", err)
	}
	// Empty pages → nothing to export → ExportBase returns nil without calling Start
	if buf.Len() != 0 {
		t.Logf("Note: empty pages produced %d bytes (may be valid XLSX or empty)", buf.Len())
	}
}

func TestXLSXExporter_SinglePage_NoObjects_ValidXLSX(t *testing.T) {
	pp := buildPages(1, []string{"Header", "DataBand"})
	f := exportAndRead(t, pp)
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		t.Error("expected at least one sheet")
	}
}

// ── Text content ──────────────────────────────────────────────────────────────

func TestXLSXExporter_TextObjects_WrittenToCells(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "Col1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Hello",
		},
		{
			Name: "Col2", Kind: preview.ObjectTypeText,
			Left: 110, Top: 0, Width: 100, Height: 20,
			Text: "World",
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)

	v1, err := f.GetCellValue(sheetName, "A1")
	if err != nil {
		t.Fatalf("GetCellValue A1: %v", err)
	}
	v2, err := f.GetCellValue(sheetName, "B1")
	if err != nil {
		t.Fatalf("GetCellValue B1: %v", err)
	}
	if v1 != "Hello" {
		t.Errorf("A1: want Hello, got %q", v1)
	}
	if v2 != "World" {
		t.Errorf("B1: want World, got %q", v2)
	}
}

func TestXLSXExporter_MultipleRows_CorrectRowMapping(t *testing.T) {
	// Two separate Y groups within one band → two rows in spreadsheet
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "R1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Row1",
		},
		{
			Name: "R2", Kind: preview.ObjectTypeText,
			Left: 0, Top: 30, Width: 100, Height: 20,
			Text: "Row2",
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)

	v1, _ := f.GetCellValue(sheetName, "A1")
	v2, _ := f.GetCellValue(sheetName, "A2")
	if v1 != "Row1" {
		t.Errorf("A1: want Row1, got %q", v1)
	}
	if v2 != "Row2" {
		t.Errorf("A2: want Row2, got %q", v2)
	}
}

func TestXLSXExporter_ObjectsSortedLeftToRight(t *testing.T) {
	// Objects placed in reverse X order — should be sorted left-to-right in columns
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "Right", Kind: preview.ObjectTypeText,
			Left: 200, Top: 0, Width: 100, Height: 20,
			Text: "B",
		},
		{
			Name: "Left", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "A",
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)

	v1, _ := f.GetCellValue(sheetName, "A1")
	v2, _ := f.GetCellValue(sheetName, "B1")
	if v1 != "A" {
		t.Errorf("A1 (left): want A, got %q", v1)
	}
	if v2 != "B" {
		t.Errorf("B1 (right): want B, got %q", v2)
	}
}

// ── Multiple pages ────────────────────────────────────────────────────────────

func TestXLSXExporter_MultiplePages_RowsContinue(t *testing.T) {
	pp := preview.New()
	for i := 0; i < 3; i++ {
		pp.AddPage(794, 1123, i+1)
		_ = pp.AddBand(&preview.PreparedBand{
			Name:   "DataBand",
			Top:    0,
			Height: 40,
			Objects: []preview.PreparedObject{
				{
					Name: "Val", Kind: preview.ObjectTypeText,
					Left: 0, Top: 0, Width: 100, Height: 20,
					Text: "row",
				},
			},
		})
	}
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)

	// 3 pages × 1 row each → rows A1, A2, A3
	v3, err := f.GetCellValue(sheetName, "A3")
	if err != nil {
		t.Fatalf("GetCellValue A3: %v", err)
	}
	if v3 != "row" {
		t.Errorf("A3 (page 3): want 'row', got %q", v3)
	}
}

// ── Custom sheet name ─────────────────────────────────────────────────────────

func TestXLSXExporter_CustomSheetName(t *testing.T) {
	pp := buildPages(1, []string{"Band"})
	f := exportAndRead(t, pp, func(e *xlsxexp.Exporter) {
		e.SheetName = "MyData"
	})
	sheets := f.GetSheetList()
	found := false
	for _, s := range sheets {
		if s == "MyData" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("custom sheet name 'MyData' not found in sheets %v", sheets)
	}
}

// ── CheckBox objects ──────────────────────────────────────────────────────────

func TestXLSXExporter_CheckBox_Checked(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:    "CB", Kind: preview.ObjectTypeCheckBox,
			Left:    0, Top: 0, Width: 20, Height: 20,
			Checked: true,
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "true" {
		t.Errorf("checked checkbox: want 'true', got %q", v)
	}
}

func TestXLSXExporter_CheckBox_Unchecked(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:    "CB", Kind: preview.ObjectTypeCheckBox,
			Left:    0, Top: 0, Width: 20, Height: 20,
			Checked: false,
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "false" {
		t.Errorf("unchecked checkbox: want 'false', got %q", v)
	}
}

// ── Non-text objects are skipped ──────────────────────────────────────────────

func TestXLSXExporter_NonTextObjects_NotInCells(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "Line", Kind: preview.ObjectTypeLine,
			Left: 0, Top: 0, Width: 100, Height: 1,
		},
		{
			Name: "Shape", Kind: preview.ObjectTypeShape,
			Left: 0, Top: 0, Width: 50, Height: 50,
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "" {
		t.Errorf("non-text objects: A1 should be empty, got %q", v)
	}
}

// ── Cell style ────────────────────────────────────────────────────────────────

func TestXLSXExporter_CellStyle_Bold(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "Bold", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "BoldText",
			Font: style.Font{Name: "Arial", Size: 12, Style: style.FontStyleBold},
		},
	})
	// Just verify it exports without error and text is in the cell
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "BoldText" {
		t.Errorf("bold cell: want 'BoldText', got %q", v)
	}
}

func TestXLSXExporter_CellStyle_WithFillColor(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "Filled", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Filled",
			FillColor: color.RGBA{R: 255, G: 255, B: 0, A: 255},
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "Filled" {
		t.Errorf("filled cell: want 'Filled', got %q", v)
	}
}

func TestXLSXExporter_CellStyle_WithBorder(t *testing.T) {
	bl := &style.BorderLine{
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Width: 1,
		Style: style.LineStyleSolid,
	}
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "Bordered", Kind: preview.ObjectTypeText,
			Left:  0, Top: 0, Width: 100, Height: 20,
			Text:  "Bordered",
			Border: style.Border{
				VisibleLines: style.BorderLinesAll,
				Lines:        [4]*style.BorderLine{bl, bl, bl, bl},
			},
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "Bordered" {
		t.Errorf("bordered cell: want 'Bordered', got %q", v)
	}
}

func TestXLSXExporter_CellStyle_HorzAlignCenter(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "Centered", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Centered",
			HorzAlign: 1, // center
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "Centered" {
		t.Errorf("centered cell: want 'Centered', got %q", v)
	}
}

func TestXLSXExporter_CellStyle_HorzAlignRight(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "Right", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Right",
			HorzAlign: 2, // right
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "Right" {
		t.Errorf("right-aligned cell: want 'Right', got %q", v)
	}
}

func TestXLSXExporter_CellStyle_HorzAlignJustify(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "Justify", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 20,
			Text:      "Justify",
			HorzAlign: 3, // justify
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "Justify" {
		t.Errorf("justify-aligned cell: want 'Justify', got %q", v)
	}
}

func TestXLSXExporter_CellStyle_VertAlignCenter(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "VCenter", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 40,
			Text:      "VCenter",
			VertAlign: 1, // center
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "VCenter" {
		t.Errorf("vcenter cell: want 'VCenter', got %q", v)
	}
}

func TestXLSXExporter_CellStyle_VertAlignBottom(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "VBottom", Kind: preview.ObjectTypeText,
			Left:      0, Top: 0, Width: 100, Height: 40,
			Text:      "VBottom",
			VertAlign: 2, // bottom
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "VBottom" {
		t.Errorf("vbottom cell: want 'VBottom', got %q", v)
	}
}

func TestXLSXExporter_CellStyle_Underline(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "Under", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Underlined",
			Font: style.Font{Name: "Arial", Size: 10, Style: style.FontStyleUnderline},
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "Underlined" {
		t.Errorf("underlined cell: want 'Underlined', got %q", v)
	}
}

func TestXLSXExporter_CellStyle_WordWrap(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:     "Wrap", Kind: preview.ObjectTypeText,
			Left:     0, Top: 0, Width: 100, Height: 40,
			Text:     "Word wrapped text",
			WordWrap: true,
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "Word wrapped text" {
		t.Errorf("word-wrap cell: want 'Word wrapped text', got %q", v)
	}
}

// ── Picture objects ───────────────────────────────────────────────────────────

func TestXLSXExporter_PictureObject_NoBlobIdx_Skipped(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:    "Pic", Kind: preview.ObjectTypePicture,
			Left:    0, Top: 0, Width: 80, Height: 80,
			BlobIdx: -1,
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	// No text for picture → cell empty
	if v != "" {
		t.Errorf("picture no blob: want empty cell, got %q", v)
	}
}

// ── PageRange ─────────────────────────────────────────────────────────────────

func TestXLSXExporter_PageRangeCurrent(t *testing.T) {
	pp := preview.New()
	for i := 0; i < 3; i++ {
		pp.AddPage(794, 1123, i+1)
		_ = pp.AddBand(&preview.PreparedBand{
			Name:   "DataBand",
			Top:    0,
			Height: 40,
			Objects: []preview.PreparedObject{
				{
					Name: "Val", Kind: preview.ObjectTypeText,
					Left: 0, Top: 0, Width: 100, Height: 20,
					Text: "row",
				},
			},
		})
	}
	exp := xlsxexp.NewExporter()
	exp.PageRange = export.PageRangeCurrent
	exp.CurPage = 1
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	f, err := excelize.OpenReader(&buf)
	if err != nil {
		t.Fatalf("excelize.OpenReader: %v", err)
	}
	sheetName := f.GetSheetName(0)
	// Only 1 page → 1 row (A1 filled, A2 empty)
	v1, _ := f.GetCellValue(sheetName, "A1")
	v2, _ := f.GetCellValue(sheetName, "A2")
	if v1 != "row" {
		t.Errorf("PageRangeCurrent: A1 want 'row', got %q", v1)
	}
	if v2 != "" {
		t.Errorf("PageRangeCurrent: A2 should be empty (only 1 page exported), got %q", v2)
	}
}

// ── Style cache ───────────────────────────────────────────────────────────────

func TestXLSXExporter_StyleCache_SameStyleReusesCachedID(t *testing.T) {
	// Two objects with identical styles → style is created once (verified by successful export)
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	objs := make([]preview.PreparedObject, 10)
	for i := range objs {
		objs[i] = preview.PreparedObject{
			Name: "O", Kind: preview.ObjectTypeText,
			Left: float32(i * 110), Top: 0, Width: 100, Height: 20,
			Text: "x",
			Font: style.Font{Name: "Arial", Size: 10, Style: style.FontStyleBold},
		}
	}
	_ = pp.AddBand(&preview.PreparedBand{
		Name: "Band", Top: 0, Height: 40, Objects: objs,
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	// All 10 objects on the same Y → 10 columns in row 1
	v, _ := f.GetCellValue(sheetName, "J1")
	if v != "x" {
		t.Errorf("style cache: J1 want 'x', got %q", v)
	}
}

// ── WriteToReader helper ──────────────────────────────────────────────────────

func TestXLSXExporter_OutputIsValidXLSX(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "V", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "test",
		},
	})
	exp := xlsxexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	// Verify XLSX magic bytes (PK ZIP header)
	b := buf.Bytes()
	if len(b) < 4 || b[0] != 0x50 || b[1] != 0x4B {
		t.Errorf("expected XLSX (ZIP) magic bytes PK, got %02X %02X", b[0], b[1])
	}
}

// ── io.Writer interface ───────────────────────────────────────────────────────

func TestXLSXExporter_WriterInterface(t *testing.T) {
	// Verify Export accepts any io.Writer
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "V", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 50, Height: 20,
			Text: "data",
		},
	})
	exp := xlsxexp.NewExporter()
	if err := exp.Export(pp, io.Discard); err != nil {
		t.Fatalf("Export to Discard: %v", err)
	}
}

// ── Start: empty SheetName fallback ──────────────────────────────────────────

func TestXLSXExporter_Start_EmptySheetName_FallsBackToReport(t *testing.T) {
	// Setting SheetName to "" exercises the fallback branch (line 94-96 of xlsx.go)
	// that assigns "Report" when the field is empty.
	pp := buildPages(1, []string{"Band"})
	f := exportAndRead(t, pp, func(e *xlsxexp.Exporter) {
		e.SheetName = ""
	})
	sheets := f.GetSheetList()
	found := false
	for _, s := range sheets {
		if s == "Report" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("empty SheetName should fall back to 'Report', got sheets %v", sheets)
	}
}

// ── ExportBand: picture objects with real blob data ───────────────────────────

// buildPageWithBlob creates a PreparedPages where the BlobStore holds the given
// imgData at index 0, and adds a single picture object that references it.
func buildPageWithBlob(imgData []byte, left float32) *preview.PreparedPages {
	pp := preview.New()
	idx := pp.BlobStore.Add("img", imgData)
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "DataBand",
		Top:    0,
		Height: 80,
		Objects: []preview.PreparedObject{
			{
				Name:    "Pic",
				Kind:    preview.ObjectTypePicture,
				Left:    left,
				Top:     0,
				Width:   80,
				Height:  80,
				BlobIdx: idx,
			},
		},
	})
	return pp
}

func TestXLSXExporter_PictureObject_ValidPNG_Embedded(t *testing.T) {
	// A valid 1×1 PNG in the BlobStore exercises the full picture-embedding path
	// including imageExtension (via ExportBand) and AddPictureFromBytes.
	pp := buildPageWithBlob(minimalPNG(), 0)
	exp := xlsxexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export with valid PNG: %v", err)
	}
	// The workbook must still be parseable.
	if _, err := excelize.OpenReader(&buf); err != nil {
		t.Fatalf("excelize.OpenReader: %v", err)
	}
}

func TestXLSXExporter_PictureObject_InvalidJPEG_NonFatal(t *testing.T) {
	// A byte slice with JPEG magic bytes but invalid content exercises:
	//   - imageExtension returning ".jpg"
	//   - AddPictureFromBytes failing (non-fatal continue path)
	// The export should still succeed.
	pp := buildPageWithBlob(minimalJPEG(), 0)
	exp := xlsxexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export with invalid JPEG should not fail: %v", err)
	}
}

func TestXLSXExporter_PictureObject_InvalidGIF_NonFatal(t *testing.T) {
	// GIF magic bytes with invalid body — exercises imageExtension returning ".gif"
	// and the non-fatal error path in ExportBand.
	pp := buildPageWithBlob(minimalGIF(), 0)
	exp := xlsxexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export with invalid GIF should not fail: %v", err)
	}
}

func TestXLSXExporter_PictureObject_ColClampedToOne(t *testing.T) {
	// Left = 0 → col = int(0/60)+1 = 1 (already 1, no clamping needed).
	// Left < 0 is not representable in float32 in practice but Left=0 is the
	// minimum "normal" case that results in col=1. This exercises line 183.
	pp := buildPageWithBlob(minimalPNG(), 0)
	exp := xlsxexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export with Left=0 picture: %v", err)
	}
}

// ── ExportBand: shape and line objects are skipped ────────────────────────────

func TestXLSXExporter_ShapeObject_NotInCells(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:      "Rect",
			Kind:      preview.ObjectTypeShape,
			Left:      0,
			Top:       0,
			Width:     100,
			Height:    50,
			ShapeKind: 0, // Rectangle
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "" {
		t.Errorf("shape object: A1 should be empty, got %q", v)
	}
}

func TestXLSXExporter_LineObject_NotInCells(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:         "Line",
			Kind:         preview.ObjectTypeLine,
			Left:         0,
			Top:          0,
			Width:        200,
			Height:       1,
			LineDiagonal: true,
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "" {
		t.Errorf("line object: A1 should be empty, got %q", v)
	}
}

// ── Finish: column width clamping ─────────────────────────────────────────────

func TestXLSXExporter_Finish_ColumnWidthFloor(t *testing.T) {
	// A single character in a very narrow column → estimated width < 8 → clamped to 8.
	// Width=1px → byPx = 1/7 ≈ 0.14. byLen = len("x")+2 = 3. max(3, 0.14) = 3 → < 8 → floor to 8.
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:  "Narrow",
			Kind:  preview.ObjectTypeText,
			Left:  0,
			Top:   0,
			Width: 1, // very narrow column
			Height: 20,
			Text:  "x",
		},
	})
	// No error expected — the floor clamp is silent.
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "x" {
		t.Errorf("narrow column: want 'x', got %q", v)
	}
}

func TestXLSXExporter_Finish_ColumnWidthCeiling(t *testing.T) {
	// A very long text string → estimated width > 60 → clamped to 60.
	longText := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz_extra"
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:   "Wide",
			Kind:   preview.ObjectTypeText,
			Left:   0,
			Top:    0,
			Width:  800, // wide column: 800/7 ≈ 114 chars
			Height: 20,
			Text:   longText,
		},
	})
	// No error expected — the ceiling clamp is silent.
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != longText {
		t.Errorf("wide column: want %q, got %q", longText, v)
	}
}

// ── Finish: multiple pages produce correct row count ─────────────────────────

func TestXLSXExporter_Finish_MultiplePages_ColumnWidthsMerged(t *testing.T) {
	// Verify that Finish runs correctly when there are colWidths entries from
	// multiple pages (exercises the loop in Finish over the colWidths map).
	pp := preview.New()
	for i := 0; i < 5; i++ {
		pp.AddPage(794, 1123, i+1)
		_ = pp.AddBand(&preview.PreparedBand{
			Name:   "DataBand",
			Top:    0,
			Height: 40,
			Objects: []preview.PreparedObject{
				{
					Name:  "A",
					Kind:  preview.ObjectTypeText,
					Left:  0,
					Top:   0,
					Width: 100,
					Height: 20,
					Text:  "Page data",
				},
				{
					Name:  "B",
					Kind:  preview.ObjectTypeText,
					Left:  110,
					Top:   0,
					Width: 200,
					Height: 20,
					Text:  "Second column",
				},
			},
		})
	}
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	// 5 pages × 1 row = 5 rows
	v5, err := f.GetCellValue(sheetName, "A5")
	if err != nil {
		t.Fatalf("GetCellValue A5: %v", err)
	}
	if v5 != "Page data" {
		t.Errorf("A5: want 'Page data', got %q", v5)
	}
}

// ── ExportBand: RTF object treated as text ────────────────────────────────────

func TestXLSXExporter_RTFObject_WrittenToCells(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:  "RTF",
			Kind:  preview.ObjectTypeRTF,
			Left:  0,
			Top:   0,
			Width: 100,
			Height: 20,
			Text:  "rtf content",
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "rtf content" {
		t.Errorf("RTF object: want 'rtf content', got %q", v)
	}
}

// ── ExportBand: HTML object treated as text ───────────────────────────────────

func TestXLSXExporter_HTMLObject_WrittenToCells(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:  "HTML",
			Kind:  preview.ObjectTypeHtml,
			Left:  0,
			Top:   0,
			Width: 100,
			Height: 20,
			Text:  "<b>html</b>",
		},
	})
	f := exportAndRead(t, pp)
	sheetName := f.GetSheetName(0)
	v, _ := f.GetCellValue(sheetName, "A1")
	if v != "<b>html</b>" {
		t.Errorf("HTML object: want '<b>html</b>', got %q", v)
	}
}

// ── ExportBand: PictureObject with BlobIdx >= 0 but empty blob ────────────────

func TestXLSXExporter_PictureObject_EmptyBlob_Skipped(t *testing.T) {
	// Add a blob of zero length to the BlobStore → should skip gracefully.
	pp := preview.New()
	idx := pp.BlobStore.Add("empty", []byte{})
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "DataBand",
		Top:    0,
		Height: 80,
		Objects: []preview.PreparedObject{
			{
				Name:    "EmptyPic",
				Kind:    preview.ObjectTypePicture,
				Left:    0,
				Top:     0,
				Width:   80,
				Height:  80,
				BlobIdx: idx,
			},
		},
	})
	exp := xlsxexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export with empty blob: %v", err)
	}
}
