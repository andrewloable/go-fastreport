package pdf

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewPages(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)

	if pages == nil {
		t.Fatal("NewPages returned nil")
	}
	if pages.Count() != 0 {
		t.Errorf("expected 0 pages, got %d", pages.Count())
	}
}

func TestPages_AddPage(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	page := NewPage(w, pages, 595, 842)

	if pages.Count() != 1 {
		t.Errorf("expected 1 page after AddPage, got %d", pages.Count())
	}
	if pages.pageList[0] != page {
		t.Error("pageList[0] does not match the added page")
	}
}

func TestPages_AddPage_Multiple(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	for i := 0; i < 3; i++ {
		NewPage(w, pages, 595, 842)
	}
	if pages.Count() != 3 {
		t.Errorf("expected 3 pages, got %d", pages.Count())
	}
}

func TestPages_OutputContainsType(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "/Type /Pages") {
		t.Error("output missing /Type /Pages")
	}
}

func TestPages_OutputContainsCount(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	NewPage(w, pages, 595, 842)
	NewPage(w, pages, 595, 842)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "/Count 2") {
		t.Error("output missing /Count 2 for pages tree")
	}
}

func TestPages_OutputContainsKids(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	NewPage(w, pages, 595, 842)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "/Kids") {
		t.Error("output missing /Kids")
	}
}

func TestNewPage(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	page := NewPage(w, pages, 595.28, 841.89)

	if page == nil {
		t.Fatal("NewPage returned nil")
	}
	if page.Width != 595.28 {
		t.Errorf("expected Width 595.28, got %f", page.Width)
	}
	if page.Height != 841.89 {
		t.Errorf("expected Height 841.89, got %f", page.Height)
	}
}

func TestNewPage_OutputContainsType(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	NewPage(w, pages, 595, 842)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "/Type /Page") {
		t.Error("output missing /Type /Page")
	}
}

func TestNewPage_OutputContainsMediaBox(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	NewPage(w, pages, 595, 842)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "/MediaBox") {
		t.Error("output missing /MediaBox")
	}
}

func TestNewPage_OutputContainsResources(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	NewPage(w, pages, 595, 842)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "/Resources") {
		t.Error("output missing /Resources")
	}
}

func TestNewPage_OutputContainsParent(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	NewPage(w, pages, 595, 842)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "/Parent") {
		t.Error("output missing /Parent reference")
	}
}

func TestNewPage_OutputContainsContents(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	NewPage(w, pages, 595, 842)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "/Contents") {
		t.Error("output missing /Contents reference")
	}
}

func TestPage_Contents(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	page := NewPage(w, pages, 595, 842)

	c := page.Contents()
	if c == nil {
		t.Fatal("Contents() returned nil")
	}
}

// TestPages_KidsUsesIndirectRef verifies that /Kids entries are written as
// proper PDF indirect references ("N G R"), NOT as name-encoded strings like
// "/1#200#20R".  This guards against the bug where core.NewName(obj.Reference())
// was used instead of core.NewRef(obj).
func TestPages_KidsUsesIndirectRef(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	NewPage(w, pages, 595, 842)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()

	// A correct indirect reference looks like "2 0 R" (digits, space, digit, space, R)
	if !strings.Contains(output, " 0 R") {
		t.Error("output does not contain proper indirect references (N G R format)")
	}
	// A name-encoded reference would contain "#20" (hex-encoded space) — this must NOT appear
	if strings.Contains(output, "#20") {
		t.Error("output contains #20 (hex-encoded space), indicating NewName was used for an indirect reference instead of NewRef")
	}
}

// TestPage_ParentUsesIndirectRef checks that /Parent is a proper indirect reference.
func TestPage_ParentUsesIndirectRef(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	NewPage(w, pages, 595, 842)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()

	// /Parent must be followed by an indirect reference like "1 0 R"
	if !strings.Contains(output, "/Parent ") {
		t.Error("output missing /Parent entry")
	}
	if strings.Contains(output, "#20") {
		t.Error("output contains hex-encoded space in reference (indirect ref used as Name)")
	}
}

// TestPage_ContentsRefUsesIndirectRef checks that /Contents links the content stream by
// proper indirect reference.
func TestPage_ContentsRefUsesIndirectRef(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	page := NewPage(w, pages, 595, 842)
	page.Contents().WriteString("q Q")
	page.Contents().Finalize()
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()

	if !strings.Contains(output, "/Contents ") {
		t.Error("output missing /Contents entry")
	}
	if strings.Contains(output, "#20") {
		t.Error("output contains hex-encoded space — indirect ref encoded as Name")
	}
}

// TestCatalog_PagesRefUsesIndirectRef checks /Pages in the catalog uses "N G R" form.
func TestCatalog_PagesRefUsesIndirectRef(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()

	if !strings.Contains(output, "/Pages ") {
		t.Error("output missing /Pages entry in catalog")
	}
	if strings.Contains(output, "#20") {
		t.Error("output contains hex-encoded space in catalog /Pages reference")
	}
}
