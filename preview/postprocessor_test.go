package preview_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
)

func TestNewPostprocessor(t *testing.T) {
	pp := preview.New()
	proc := preview.NewPostprocessor(pp)
	if proc == nil {
		t.Fatal("NewPostprocessor returned nil")
	}
}

func TestProcess_EmptyPages(t *testing.T) {
	pp := preview.New()
	proc := preview.NewPostprocessor(pp)
	proc.Process() // should not panic
}

func TestProcess_NoBands(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	proc := preview.NewPostprocessor(pp)
	proc.Process() // should not panic
}

func TestProcess_MacroSubstitution(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)
	// page 1 has text with macros
	_ = pp.GetPage(0) // keep current on page 1 — AddPage left curPage=1 (page 2)
	// Go back: use SetAddPageAction trick or just AddBand will fail. Use Save/Load approach.
	// Simpler: build separate pp with page 1 active.
	pp2 := preview.New()
	pp2.AddPage(595, 842, 1)
	err := pp2.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 20,
		Objects: []preview.PreparedObject{
			{Kind: preview.ObjectTypeText, Text: "Total:[TotalPages] Count:[PageCount] Page:[Page]"},
		},
	})
	if err != nil {
		t.Fatalf("AddBand: %v", err)
	}
	pp2.AddPage(595, 842, 2)

	proc := preview.NewPostprocessor(pp2)
	proc.Process()

	pg := pp2.GetPage(0)
	txt := pg.Bands[0].Objects[0].Text
	want := "Total:2 Count:2 Page:1"
	if txt != want {
		t.Errorf("text = %q, want %q", txt, want)
	}
}

func TestProcess_MacroSubstitution_Page2(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pp.AddPage(595, 842, 2)
	// AddBand goes to current page (page 2)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 20,
		Objects: []preview.PreparedObject{
			{Kind: preview.ObjectTypeText, Text: "[Page] of [TotalPages]"},
		},
	})

	proc := preview.NewPostprocessor(pp)
	proc.Process()

	pg := pp.GetPage(1)
	txt := pg.Bands[0].Objects[0].Text
	want := "2 of 2"
	if txt != want {
		t.Errorf("text = %q, want %q", txt, want)
	}
}

func TestProcess_MacroSubstitution_SystemVariableTokens(t *testing.T) {
	// The engine injects "[PAGE#]" and "[TOTALPAGES#]" as literal bracket
	// tokens into text objects. The postprocessor must replace them with the
	// actual page/total strings, matching C# PreparedPagePostprocessor behavior.
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 20,
		Objects: []preview.PreparedObject{
			{Kind: preview.ObjectTypeText, Text: "Page [PAGE#] of [TOTALPAGES#]"},
		},
	})
	pp.AddPage(595, 842, 2)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b2",
		Top:    0,
		Height: 20,
		Objects: []preview.PreparedObject{
			{Kind: preview.ObjectTypeText, Text: "[PAGE#]/[TOTALPAGES#]"},
		},
	})
	pp.AddPage(595, 842, 3)

	preview.NewPostprocessor(pp).Process()

	// Page 1
	pg1 := pp.GetPage(0)
	if txt := pg1.Bands[0].Objects[0].Text; txt != "Page 1 of 3" {
		t.Errorf("page 1 text = %q, want %q", txt, "Page 1 of 3")
	}
	// Page 2
	pg2 := pp.GetPage(1)
	if txt := pg2.Bands[0].Objects[0].Text; txt != "2/3" {
		t.Errorf("page 2 text = %q, want %q", txt, "2/3")
	}
}

func TestProcess_Duplicates_Clear(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []preview.PreparedObject{
			{Name: "field1", Kind: preview.ObjectTypeText, Top: 0, Height: 20, Text: "ABC", Duplicates: preview.DuplicatesClear},
			{Name: "field1", Kind: preview.ObjectTypeText, Top: 20, Height: 20, Text: "ABC", Duplicates: preview.DuplicatesClear},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Text != "ABC" {
		t.Errorf("first text = %q, want ABC", objs[0].Text)
	}
	if objs[1].Text != "" {
		t.Errorf("duplicate should be cleared, got %q", objs[1].Text)
	}
}

func TestProcess_Duplicates_Hide(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []preview.PreparedObject{
			{Name: "f2", Kind: preview.ObjectTypeText, Top: 0, Width: 100, Height: 20, Text: "XYZ", Duplicates: preview.DuplicatesHide},
			{Name: "f2", Kind: preview.ObjectTypeText, Top: 20, Width: 100, Height: 20, Text: "XYZ", Duplicates: preview.DuplicatesHide},
		},
	})

	preview.NewPostprocessor(pp).Process()

	o := pp.GetPage(0).Bands[0].Objects[1]
	if o.Width != 0 || o.Height != 0 || o.Text != "" {
		t.Errorf("hidden obj: width=%v height=%v text=%q, want 0,0,\"\"", o.Width, o.Height, o.Text)
	}
}

func TestProcess_Duplicates_Merge(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 60,
		Objects: []preview.PreparedObject{
			{Name: "f3", Kind: preview.ObjectTypeText, Top: 0, Width: 100, Height: 20, Text: "merged", Duplicates: preview.DuplicatesMerge},
			{Name: "f3", Kind: preview.ObjectTypeText, Top: 20, Width: 100, Height: 20, Text: "merged", Duplicates: preview.DuplicatesMerge},
			{Name: "f3", Kind: preview.ObjectTypeText, Top: 40, Width: 100, Height: 20, Text: "merged", Duplicates: preview.DuplicatesMerge},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Height != 60 {
		t.Errorf("merged height = %v, want 60", objs[0].Height)
	}
	if objs[1].Width != 0 || objs[2].Width != 0 {
		t.Error("merged others should have width 0")
	}
}

func TestProcess_Duplicates_DifferentText_NoChange(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []preview.PreparedObject{
			{Name: "f4", Kind: preview.ObjectTypeText, Top: 0, Height: 20, Text: "A", Duplicates: preview.DuplicatesClear},
			{Name: "f4", Kind: preview.ObjectTypeText, Top: 20, Height: 20, Text: "B", Duplicates: preview.DuplicatesClear},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[0].Text != "A" || objs[1].Text != "B" {
		t.Error("different text objects should not be modified")
	}
}

func TestProcess_Duplicates_NotAdjacent_NoChange(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 60,
		Objects: []preview.PreparedObject{
			// absBottom of obj1 = 0+0+20=20; absTop of obj2 = 0+25=25; gap=5 > 0.5
			{Name: "f5", Kind: preview.ObjectTypeText, Top: 0, Height: 20, Text: "same", Duplicates: preview.DuplicatesClear},
			{Name: "f5", Kind: preview.ObjectTypeText, Top: 25, Height: 20, Text: "same", Duplicates: preview.DuplicatesClear},
		},
	})

	preview.NewPostprocessor(pp).Process()

	objs := pp.GetPage(0).Bands[0].Objects
	if objs[1].Text != "same" {
		t.Errorf("non-adjacent duplicate should not be cleared, got %q", objs[1].Text)
	}
}

func TestProcess_Duplicates_Show_NoChange(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []preview.PreparedObject{
			{Name: "f6", Kind: preview.ObjectTypeText, Top: 0, Height: 20, Text: "show", Duplicates: preview.DuplicatesShow},
			{Name: "f6", Kind: preview.ObjectTypeText, Top: 20, Height: 20, Text: "show", Duplicates: preview.DuplicatesShow},
		},
	})

	preview.NewPostprocessor(pp).Process()

	if txt := pp.GetPage(0).Bands[0].Objects[1].Text; txt != "show" {
		t.Errorf("DuplicatesShow should not modify text, got %q", txt)
	}
}

func TestProcess_Duplicates_NoName_Skipped(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []preview.PreparedObject{
			{Name: "", Kind: preview.ObjectTypeText, Top: 0, Height: 20, Text: "X", Duplicates: preview.DuplicatesClear},
			{Name: "", Kind: preview.ObjectTypeText, Top: 20, Height: 20, Text: "X", Duplicates: preview.DuplicatesClear},
		},
	})

	preview.NewPostprocessor(pp).Process()

	if txt := pp.GetPage(0).Bands[0].Objects[1].Text; txt != "X" {
		t.Errorf("unnamed objects should not be cleared, got %q", txt)
	}
}

func TestProcess_Duplicates_NonTextKind_Skipped(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 40,
		Objects: []preview.PreparedObject{
			{Name: "pic", Kind: preview.ObjectTypePicture, Top: 0, Height: 20, Duplicates: preview.DuplicatesClear},
			{Name: "pic", Kind: preview.ObjectTypePicture, Top: 20, Height: 20, Duplicates: preview.DuplicatesClear},
		},
	})

	preview.NewPostprocessor(pp).Process() // should not panic
}

// ── InitialPageNumber TotalPages fix ─────────────────────────────────────────

// TestProcess_TotalPages_InitialPageNumber verifies that [TotalPages] macro
// accounts for InitialPageNumber (i.e. is logicalTotal = firstPage.PageNo + count - 1).
// C# reference: PreparedPages.cs line 283: macroValues["TotalPages#"] = count + InitialPageNumber - 1.
func TestProcess_TotalPages_InitialPageNumber(t *testing.T) {
	pp := preview.New()
	// Simulate InitialPageNumber=3 with 2 pages (PageNo 3,4)
	pp.AddPage(595, 842, 3)
	pg1 := pp.CurrentPage()
	pg1.Bands = append(pg1.Bands, &preview.PreparedBand{
		Objects: []preview.PreparedObject{
			{Name: "txt", Text: "[TotalPages]"},
		},
	})

	pp.AddPage(595, 842, 4)
	pg2 := pp.CurrentPage()
	pg2.Bands = append(pg2.Bands, &preview.PreparedBand{
		Objects: []preview.PreparedObject{
			{Name: "txt", Text: "[TotalPages]"},
		},
	})

	preview.NewPostprocessor(pp).Process()

	// With InitialPageNumber=3 and count=2: logicalTotal = 3 + 2 - 1 = 4
	want := "4"
	got := pp.GetPage(0).Bands[0].Objects[0].Text
	if got != want {
		t.Errorf("page1 TotalPages = %q, want %q", got, want)
	}
	got = pp.GetPage(1).Bands[0].Objects[0].Text
	if got != want {
		t.Errorf("page2 TotalPages = %q, want %q", got, want)
	}
}

// TestProcess_WatermarkMacro verifies that [TotalPages] and [Page] macros
// are replaced in the PreparedPage.Watermark.Text field.
// C# reference: PreparedPagePostprocessor.cs line 224 calls ExtractMacros on watermark.
func TestProcess_WatermarkMacro(t *testing.T) {
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	pg := pp.CurrentPage()
	pg.Watermark = &preview.PreparedWatermark{
		Enabled: true,
		Text:    "Page [PAGE#] of [TOTALPAGES#]",
	}

	pp.AddPage(595, 842, 2)
	pg2 := pp.CurrentPage()
	pg2.Watermark = &preview.PreparedWatermark{
		Enabled: true,
		Text:    "Page [PAGE#] of [TOTALPAGES#]",
	}

	preview.NewPostprocessor(pp).Process()

	// 2 pages, InitialPageNumber=1 → logicalTotal = 2
	got := pp.GetPage(0).Watermark.Text
	if got != "Page 1 of 2" {
		t.Errorf("page1 watermark = %q, want \"Page 1 of 2\"", got)
	}
	got = pp.GetPage(1).Watermark.Text
	if got != "Page 2 of 2" {
		t.Errorf("page2 watermark = %q, want \"Page 2 of 2\"", got)
	}
}
