package csv_test

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"

	csvexp "github.com/andrewloable/go-fastreport/export/csv"
	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
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

func exportCSV(t *testing.T, pp *preview.PreparedPages, opts ...func(*csvexp.Exporter)) string {
	t.Helper()
	exp := csvexp.NewExporter()
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

func TestCSVExporter_Defaults(t *testing.T) {
	exp := csvexp.NewExporter()
	if exp.Separator != ',' {
		t.Errorf("default Separator: want ',', got %q", exp.Separator)
	}
	if exp.Quote != '"' {
		t.Errorf("default Quote: want '\"', got %q", exp.Quote)
	}
	if !exp.HeaderRow {
		t.Error("default HeaderRow should be true")
	}
	if exp.FileExtension() != ".csv" {
		t.Errorf("FileExtension: want .csv, got %s", exp.FileExtension())
	}
	if exp.Name() != "CSV" {
		t.Errorf("Name: want CSV, got %s", exp.Name())
	}
}

func TestCSVExporter_NilPages_ReturnsError(t *testing.T) {
	exp := csvexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(nil, &buf); err == nil {
		t.Error("expected error for nil PreparedPages")
	}
}

func TestCSVExporter_EmptyPages_NoOutput(t *testing.T) {
	pp := preview.New()
	exp := csvexp.NewExporter()
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export on empty pages: %v", err)
	}
	// Empty pages → nothing to export → empty output
	if buf.Len() != 0 {
		t.Errorf("expected empty output for empty pages, got %q", buf.String())
	}
}

func TestCSVExporter_SinglePage_NoObjects_EmptyOutput(t *testing.T) {
	pp := buildPages(1, []string{"Header"})
	out := exportCSV(t, pp)
	// Band has no text objects → CSV writer has nothing to write → empty
	if out != "" {
		t.Errorf("expected empty output for band with no text objects, got %q", out)
	}
}

// ── Text object rendering ─────────────────────────────────────────────────────

func TestCSVExporter_TextObjects_WithHeaderRow(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "Col1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Alice",
		},
		{
			Name: "Col2", Kind: preview.ObjectTypeText,
			Left: 110, Top: 0, Width: 100, Height: 20,
			Text: "30",
		},
	})
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = true })

	r := csv.NewReader(strings.NewReader(out))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv.ReadAll: %v", err)
	}
	// Header row + 1 data row
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (header+data), got %d", len(rows))
	}
	// Header should be the object names
	if rows[0][0] != "Col1" || rows[0][1] != "Col2" {
		t.Errorf("header: want [Col1 Col2], got %v", rows[0])
	}
	// Data row should be the text values
	if rows[1][0] != "Alice" || rows[1][1] != "30" {
		t.Errorf("data: want [Alice 30], got %v", rows[1])
	}
}

func TestCSVExporter_TextObjects_WithoutHeaderRow(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "F1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "Hello",
		},
		{
			Name: "F2", Kind: preview.ObjectTypeText,
			Left: 110, Top: 0, Width: 100, Height: 20,
			Text: "World",
		},
	})
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = false })

	r := csv.NewReader(strings.NewReader(out))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv.ReadAll: %v", err)
	}
	// No header row → only 1 data row
	if len(rows) != 1 {
		t.Fatalf("expected 1 row (no header), got %d", len(rows))
	}
	if rows[0][0] != "Hello" || rows[0][1] != "World" {
		t.Errorf("data: want [Hello World], got %v", rows[0])
	}
}

func TestCSVExporter_ObjectsSortedLeftToRight(t *testing.T) {
	// Objects placed out of order (right before left): CSV should sort them left-to-right
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
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = false })

	r := csv.NewReader(strings.NewReader(out))
	rows, _ := r.ReadAll()
	if len(rows) == 0 {
		t.Fatal("expected at least 1 row")
	}
	if rows[0][0] != "A" || rows[0][1] != "B" {
		t.Errorf("want [A B] (sorted left-to-right), got %v", rows[0])
	}
}

func TestCSVExporter_ObjectsGroupedByY(t *testing.T) {
	// Two distinct Y positions → two CSV rows
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "Row1Col1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "R1C1",
		},
		{
			Name: "Row1Col2", Kind: preview.ObjectTypeText,
			Left: 110, Top: 0, Width: 100, Height: 20,
			Text: "R1C2",
		},
		{
			Name: "Row2Col1", Kind: preview.ObjectTypeText,
			Left: 0, Top: 30, Width: 100, Height: 20,
			Text: "R2C1",
		},
	})
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = false })

	r := csv.NewReader(strings.NewReader(out))
	r.FieldsPerRecord = -1 // allow variable column counts per row
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv.ReadAll: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (2 Y groups), got %d: %q", len(rows), out)
	}
	if rows[0][0] != "R1C1" || rows[0][1] != "R1C2" {
		t.Errorf("row 0: want [R1C1 R1C2], got %v", rows[0])
	}
	if rows[1][0] != "R2C1" {
		t.Errorf("row 1: want [R2C1], got %v", rows[1])
	}
}

// ── CheckBox objects ──────────────────────────────────────────────────────────

func TestCSVExporter_CheckBox_Checked(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:    "CB", Kind: preview.ObjectTypeCheckBox,
			Left:    0, Top: 0, Width: 20, Height: 20,
			Checked: true,
		},
	})
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = false })
	if !strings.Contains(out, "true") {
		t.Errorf("checked checkbox: expected 'true', got %q", out)
	}
}

func TestCSVExporter_CheckBox_Unchecked(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:    "CB", Kind: preview.ObjectTypeCheckBox,
			Left:    0, Top: 0, Width: 20, Height: 20,
			Checked: false,
		},
	})
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = false })
	if !strings.Contains(out, "false") {
		t.Errorf("unchecked checkbox: expected 'false', got %q", out)
	}
}

func TestCSVExporter_CheckBox_TextTrue(t *testing.T) {
	// Text="true" also evaluates as checked
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name:    "CB", Kind: preview.ObjectTypeCheckBox,
			Left:    0, Top: 0, Width: 20, Height: 20,
			Text:    "true",
			Checked: false,
		},
	})
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = false })
	if !strings.Contains(out, "true") {
		t.Errorf("checkbox text=true: expected 'true', got %q", out)
	}
}

// ── Non-text objects skipped ──────────────────────────────────────────────────

func TestCSVExporter_NonTextObjects_Skipped(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "Pic", Kind: preview.ObjectTypePicture,
			Left: 0, Top: 0, Width: 80, Height: 80,
			BlobIdx: -1,
		},
		{
			Name: "Line", Kind: preview.ObjectTypeLine,
			Left: 0, Top: 0, Width: 100, Height: 1,
		},
		{
			Name: "Shape", Kind: preview.ObjectTypeShape,
			Left: 0, Top: 0, Width: 50, Height: 50,
		},
	})
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = false })
	// All non-text → empty output
	if out != "" {
		t.Errorf("non-text objects: expected empty CSV, got %q", out)
	}
}

// ── HTML and RTF objects treated as text ──────────────────────────────────────

func TestCSVExporter_HtmlObject_IncludedAsText(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "HtmlObj", Kind: preview.ObjectTypeHtml,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "<b>bold</b>",
		},
	})
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = false })
	if !strings.Contains(out, "<b>bold</b>") {
		t.Errorf("HTML object: expected raw HTML text, got %q", out)
	}
}

func TestCSVExporter_RTFObject_IncludedAsText(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "RtfObj", Kind: preview.ObjectTypeRTF,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: `{\rtf1 Hello}`,
		},
	})
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = false })
	if !strings.Contains(out, `{\rtf1 Hello}`) {
		t.Errorf("RTF object: expected raw RTF text, got %q", out)
	}
}

// ── Multiple pages ────────────────────────────────────────────────────────────

func TestCSVExporter_MultiplePages_AllData(t *testing.T) {
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
					Text: "Row",
				},
			},
		})
	}
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = false })

	r := csv.NewReader(strings.NewReader(out))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv.ReadAll: %v", err)
	}
	// 3 pages × 1 band × 1 object = 3 rows
	if len(rows) != 3 {
		t.Errorf("3 pages: expected 3 rows, got %d", len(rows))
	}
}

// ── Custom separator ──────────────────────────────────────────────────────────

func TestCSVExporter_CustomSeparator_Tab(t *testing.T) {
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "A", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "X",
		},
		{
			Name: "B", Kind: preview.ObjectTypeText,
			Left: 110, Top: 0, Width: 100, Height: 20,
			Text: "Y",
		},
	})
	out := exportCSV(t, pp, func(e *csvexp.Exporter) {
		e.Separator = '\t'
		e.HeaderRow = false
	})
	// TSV: X\tY
	if !strings.Contains(out, "X\tY") {
		t.Errorf("tab separator: expected X\\tY in %q", out)
	}
}

// ── PageRange support ─────────────────────────────────────────────────────────

func TestCSVExporter_PageRangeCurrent(t *testing.T) {
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
					Text: "page",
				},
			},
		})
	}
	exp := csvexp.NewExporter()
	exp.PageRange = export.PageRangeCurrent
	exp.CurPage = 2
	exp.HeaderRow = false
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	r := csv.NewReader(&buf)
	rows, _ := r.ReadAll()
	// Only the current page → 1 row
	if len(rows) != 1 {
		t.Errorf("PageRangeCurrent: expected 1 row, got %d", len(rows))
	}
}

func TestCSVExporter_PageRangeCustom(t *testing.T) {
	pp := preview.New()
	for i := 0; i < 5; i++ {
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
	exp := csvexp.NewExporter()
	exp.PageRange = export.PageRangeCustom
	exp.PageNumbers = "1,3"
	exp.HeaderRow = false
	var buf bytes.Buffer
	if err := exp.Export(pp, &buf); err != nil {
		t.Fatalf("Export: %v", err)
	}
	r := csv.NewReader(&buf)
	rows, _ := r.ReadAll()
	// Pages 1 and 3 → 2 rows
	if len(rows) != 2 {
		t.Errorf("PageRangeCustom (1,3): expected 2 rows, got %d", len(rows))
	}
}

// ── Object name fallback to band name in header ───────────────────────────────

func TestCSVExporter_Header_ObjectNameEmpty_FallsBackToBandName(t *testing.T) {
	pp := preview.New()
	pp.AddPage(794, 1123, 1)
	_ = pp.AddBand(&preview.PreparedBand{
		Name:   "MyBand",
		Top:    0,
		Height: 40,
		Objects: []preview.PreparedObject{
			{
				Name: "", // empty name → should fall back to band name
				Kind: preview.ObjectTypeText,
				Left: 0, Top: 0, Width: 100, Height: 20,
				Text: "value",
			},
		},
	})
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = true })

	r := csv.NewReader(strings.NewReader(out))
	rows, _ := r.ReadAll()
	if len(rows) < 1 {
		t.Fatal("expected at least header row")
	}
	if rows[0][0] != "MyBand" {
		t.Errorf("empty object name: header should be band name 'MyBand', got %q", rows[0][0])
	}
}

// ── Epsilon grouping ──────────────────────────────────────────────────────────

func TestCSVExporter_YEpsilonGrouping_SameRowWithinEpsilon(t *testing.T) {
	// Objects at Y=0 and Y=0.5 (within 1px epsilon) → same row
	pp := buildPageWithObjects([]preview.PreparedObject{
		{
			Name: "A", Kind: preview.ObjectTypeText,
			Left: 0, Top: 0, Width: 100, Height: 20,
			Text: "First",
		},
		{
			Name: "B", Kind: preview.ObjectTypeText,
			Left: 110, Top: 0.5, Width: 100, Height: 20,
			Text: "Second",
		},
	})
	out := exportCSV(t, pp, func(e *csvexp.Exporter) { e.HeaderRow = false })

	r := csv.NewReader(strings.NewReader(out))
	rows, _ := r.ReadAll()
	// Should be in the same row
	if len(rows) != 1 {
		t.Errorf("within epsilon: expected 1 row, got %d: %q", len(rows), out)
	}
	if len(rows[0]) != 2 {
		t.Errorf("within epsilon: expected 2 columns in row, got %d: %v", len(rows[0]), rows[0])
	}
}
