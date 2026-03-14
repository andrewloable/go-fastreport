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
