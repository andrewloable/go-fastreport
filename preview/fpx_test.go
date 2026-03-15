package preview_test

import (
	"bytes"
	"encoding/gob"
	"errors"
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/style"
)

// failWriter rejects all writes.
type failWriter struct{}

func (f *failWriter) Write(p []byte) (int, error) { return 0, errors.New("write failed") }

func buildTestPages() *preview.PreparedPages {
	pp := preview.New()

	// Blob
	pp.BlobStore.Add("img1", []byte{0x89, 0x50, 0x4E, 0x47})

	// Page 1
	pp.AddPage(595, 842, 1)
	pp.AddBand(&preview.PreparedBand{ //nolint
		Name:   "DataBand1",
		Top:    100,
		Height: 20,
		Objects: []preview.PreparedObject{
			{
				Name:      "Text1",
				Kind:      preview.ObjectTypeText,
				Left:      10,
				Top:       0,
				Width:     200,
				Height:    20,
				Text:      "Hello FPX",
				BlobIdx:   -1,
				Font:      style.Font{Name: "Arial", Size: 12},
				TextColor: color.RGBA{0, 0, 0, 255},
				FillColor: color.RGBA{255, 255, 255, 255},
				WordWrap:  true,
			},
			{
				Name:      "Pic1",
				Kind:      preview.ObjectTypePicture,
				Left:      220,
				Top:       0,
				Width:     50,
				Height:    20,
				BlobIdx:   0,
				IsBarcode: true,
				BarcodeModules: [][]bool{
					{true, false, true},
					{false, true, false},
				},
			},
		},
	})

	// Page 2 with watermark
	pp.AddPage(595, 842, 2)
	pg2 := pp.CurrentPage()
	pg2.Watermark = &preview.PreparedWatermark{
		Enabled:       true,
		Text:          "DRAFT",
		Font:          style.Font{Name: "Arial", Size: 48},
		TextColor:     color.RGBA{200, 200, 200, 128},
		TextRotation:  preview.WatermarkTextRotationForwardDiagonal,
		ShowTextOnTop: false,
		ImageBlobIdx:  -1,
	}

	// Bookmark
	pp.AddBookmark("Chapter1", 150)

	// Outline
	pp.Outline.Add("Chapter 1", 0, 0)
	pp.Outline.Add("Section 1.1", 0, 100)
	pp.Outline.LevelUp()
	pp.Outline.LevelUp()

	return pp
}

func TestFPX_RoundTrip(t *testing.T) {
	original := buildTestPages()

	var buf bytes.Buffer
	if err := original.Save(&buf); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := preview.Load(&buf)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Page count
	if loaded.Count() != original.Count() {
		t.Errorf("Count: got %d, want %d", loaded.Count(), original.Count())
	}

	// Page 1 bands
	pg := loaded.GetPage(0)
	if pg == nil {
		t.Fatal("GetPage(0) = nil")
	}
	if pg.PageNo != 1 {
		t.Errorf("Page[0].PageNo = %d, want 1", pg.PageNo)
	}
	if len(pg.Bands) != 1 {
		t.Fatalf("Page[0].Bands len = %d, want 1", len(pg.Bands))
	}
	band := pg.Bands[0]
	if band.Name != "DataBand1" {
		t.Errorf("band.Name = %q", band.Name)
	}
	if len(band.Objects) != 2 {
		t.Fatalf("band.Objects len = %d, want 2", len(band.Objects))
	}

	// Text object
	txt := band.Objects[0]
	if txt.Text != "Hello FPX" {
		t.Errorf("txt.Text = %q, want Hello FPX", txt.Text)
	}
	if txt.Font.Name != "Arial" {
		t.Errorf("txt.Font.Name = %q", txt.Font.Name)
	}
	if !txt.WordWrap {
		t.Error("txt.WordWrap should be true")
	}

	// Barcode object
	bc := band.Objects[1]
	if !bc.IsBarcode {
		t.Error("bc.IsBarcode should be true")
	}
	if len(bc.BarcodeModules) != 2 {
		t.Errorf("BarcodeModules rows = %d, want 2", len(bc.BarcodeModules))
	}
	if len(bc.BarcodeModules[0]) != 3 {
		t.Errorf("BarcodeModules[0] len = %d, want 3", len(bc.BarcodeModules[0]))
	}
	if bc.BarcodeModules[0][0] != true || bc.BarcodeModules[0][1] != false {
		t.Errorf("BarcodeModules[0] = %v, want [true false true]", bc.BarcodeModules[0])
	}

	// Page 2 watermark
	pg2 := loaded.GetPage(1)
	if pg2 == nil {
		t.Fatal("GetPage(1) = nil")
	}
	if pg2.Watermark == nil {
		t.Fatal("Page[1].Watermark = nil, want non-nil")
	}
	wm := pg2.Watermark
	if wm.Text != "DRAFT" {
		t.Errorf("Watermark.Text = %q, want DRAFT", wm.Text)
	}
	if wm.TextRotation != preview.WatermarkTextRotationForwardDiagonal {
		t.Errorf("Watermark.TextRotation = %v", wm.TextRotation)
	}

	// BlobStore
	if loaded.BlobStore.Count() != 1 {
		t.Errorf("BlobStore.Count = %d, want 1", loaded.BlobStore.Count())
	}
	blob := loaded.BlobStore.Get(0)
	if len(blob) != 4 || blob[0] != 0x89 {
		t.Errorf("BlobStore[0] = %v", blob)
	}

	// Bookmark
	bk := loaded.Bookmarks.Find("Chapter1")
	if bk == nil {
		t.Fatal("Bookmark Chapter1 not found")
	}
	if bk.OffsetY != 150 {
		t.Errorf("Bookmark.OffsetY = %v, want 150", bk.OffsetY)
	}

	// Outline
	root := loaded.Outline.Root
	if len(root.Children) != 1 {
		t.Fatalf("Outline root children = %d, want 1", len(root.Children))
	}
	ch1 := root.Children[0]
	if ch1.Text != "Chapter 1" {
		t.Errorf("Outline[0].Text = %q", ch1.Text)
	}
	if len(ch1.Children) != 1 || ch1.Children[0].Text != "Section 1.1" {
		t.Errorf("Outline[0].Children = %v", ch1.Children)
	}
}

func TestFPX_EmptyRoundTrip(t *testing.T) {
	original := preview.New()

	var buf bytes.Buffer
	if err := original.Save(&buf); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := preview.Load(&buf)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Count() != 0 {
		t.Errorf("Count = %d, want 0", loaded.Count())
	}
}

func TestFPX_BadMagic(t *testing.T) {
	// Write garbage
	buf := bytes.NewBufferString("XXXX garbage data that is not a valid gob stream")
	_, err := preview.Load(buf)
	if err == nil {
		t.Error("Load with bad magic: expected error, got nil")
	}
}

func TestFPX_Save_Error(t *testing.T) {
	pp := preview.New()
	err := pp.Save(&failWriter{})
	if err == nil {
		t.Error("Save with failing writer: expected error")
	}
}

func TestFPX_Load_WrongMagic(t *testing.T) {
	// Encode a gob-valid struct with just a Magic field containing a wrong value.
	// Gob matches fields by name, so fpxFile.Magic will be set to "WRONG" while
	// all other fields remain at their zero values — triggering the magic check.
	type fakeHeader struct {
		Magic string
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(fakeHeader{Magic: "WRONG_MAGIC_STRING"}); err != nil {
		t.Fatalf("gob encode: %v", err)
	}
	_, err := preview.Load(&buf)
	if err == nil {
		t.Error("Load with wrong magic: expected error, got nil")
	}
}

func TestFPX_BorderWithLines_RoundTrip(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 20,
		Objects: []preview.PreparedObject{
			{
				Kind: preview.ObjectTypeText,
				Text: "bordered",
				Border: style.Border{
					Lines: [4]*style.BorderLine{
						{Color: color.RGBA{0, 0, 0, 255}, Style: style.LineStyleSolid, Width: 1},
						nil, nil, nil,
					},
				},
			},
		},
	})

	var buf bytes.Buffer
	if err := pp.Save(&buf); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := preview.Load(&buf)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	line := loaded.GetPage(0).Bands[0].Objects[0].Border.Lines[0]
	if line == nil {
		t.Fatal("border line 0 should not be nil after round-trip")
	}
	if line.Width != 1 {
		t.Errorf("line.Width = %v, want 1", line.Width)
	}
}

