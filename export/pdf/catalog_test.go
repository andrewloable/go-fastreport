package pdf

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewCatalog(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	cat := NewCatalog(w, pages)

	if cat == nil {
		t.Fatal("NewCatalog returned nil")
	}
	if cat.obj == nil {
		t.Error("Catalog has nil indirect object")
	}
	if cat.pages != pages {
		t.Error("Catalog did not store pages reference")
	}
}

func TestNewCatalog_RegisteredAsRoot(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	_ = NewCatalog(w, pages)

	if w.catalog == nil {
		t.Error("writer catalog was not set after NewCatalog")
	}
}

func TestNewCatalog_OutputContainsType(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "/Type /Catalog") {
		t.Error("output missing /Type /Catalog")
	}
}

func TestNewCatalog_OutputContainsPages(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "/Pages") {
		t.Error("output missing /Pages reference in catalog")
	}
}

func TestNewCatalog_OutputContainsMarkInfo(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "MarkInfo") {
		t.Error("output missing MarkInfo")
	}
}

func TestNewInfo(t *testing.T) {
	w := NewWriter()
	info := NewInfo(w)

	if info == nil {
		t.Fatal("NewInfo returned nil")
	}
	if info.obj == nil {
		t.Error("Info has nil indirect object")
	}
}

func TestNewInfo_RegisteredAsInfo(t *testing.T) {
	w := NewWriter()
	_ = NewInfo(w)

	if w.info == nil {
		t.Error("writer info was not set after NewInfo")
	}
}

func TestNewInfo_DefaultCreatorProducer(t *testing.T) {
	w := NewWriter()
	info := NewInfo(w)

	if info.Creator != "go-fastreport" {
		t.Errorf("expected Creator 'go-fastreport', got %q", info.Creator)
	}
	if info.Producer != "go-fastreport" {
		t.Errorf("expected Producer 'go-fastreport', got %q", info.Producer)
	}
}

func TestInfo_SetTitle(t *testing.T) {
	w := NewWriter()
	info := NewInfo(w)
	info.SetTitle("My Report")

	if info.Title != "My Report" {
		t.Errorf("expected Title 'My Report', got %q", info.Title)
	}

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if !strings.Contains(buf.String(), "/Title") {
		t.Error("output missing /Title")
	}
}

func TestInfo_SetTitle_Empty(t *testing.T) {
	w := NewWriter()
	info := NewInfo(w)
	info.SetTitle("")

	// Empty title should not add /Title to the dictionary
	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	// Just check it doesn't panic and Title field is empty
	if info.Title != "" {
		t.Errorf("expected empty Title, got %q", info.Title)
	}
}

func TestInfo_SetAuthor(t *testing.T) {
	w := NewWriter()
	info := NewInfo(w)
	info.SetAuthor("Jane Doe")

	if info.Author != "Jane Doe" {
		t.Errorf("expected Author 'Jane Doe', got %q", info.Author)
	}

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if !strings.Contains(buf.String(), "/Author") {
		t.Error("output missing /Author")
	}
}

func TestInfo_SetAuthor_Empty(t *testing.T) {
	w := NewWriter()
	info := NewInfo(w)
	info.SetAuthor("")
	if info.Author != "" {
		t.Errorf("expected empty Author, got %q", info.Author)
	}
}

func TestInfo_SetCreator(t *testing.T) {
	w := NewWriter()
	info := NewInfo(w)
	info.SetCreator("MyApp")

	if info.Creator != "MyApp" {
		t.Errorf("expected Creator 'MyApp', got %q", info.Creator)
	}
}

func TestInfo_SetProducer(t *testing.T) {
	w := NewWriter()
	info := NewInfo(w)
	info.SetProducer("MyProducer")

	if info.Producer != "MyProducer" {
		t.Errorf("expected Producer 'MyProducer', got %q", info.Producer)
	}
}

func TestInfo_SetSubject(t *testing.T) {
	w := NewWriter()
	info := NewInfo(w)
	info.SetSubject("Quarterly Report")

	if info.Subject != "Quarterly Report" {
		t.Errorf("expected Subject 'Quarterly Report', got %q", info.Subject)
	}

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if !strings.Contains(buf.String(), "/Subject") {
		t.Error("output missing /Subject")
	}
}

func TestInfo_SetSubject_Empty(t *testing.T) {
	w := NewWriter()
	info := NewInfo(w)
	info.SetSubject("")

	if info.Subject != "" {
		t.Errorf("expected empty Subject, got %q", info.Subject)
	}
}

func TestInfo_SetKeywords(t *testing.T) {
	w := NewWriter()
	info := NewInfo(w)
	info.SetKeywords("finance sales 2024")

	if info.Keywords != "finance sales 2024" {
		t.Errorf("expected Keywords 'finance sales 2024', got %q", info.Keywords)
	}

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if !strings.Contains(buf.String(), "/Keywords") {
		t.Error("output missing /Keywords")
	}
}

func TestInfo_SetKeywords_Empty(t *testing.T) {
	w := NewWriter()
	info := NewInfo(w)
	info.SetKeywords("")

	if info.Keywords != "" {
		t.Errorf("expected empty Keywords, got %q", info.Keywords)
	}
}

// TestCatalog_Version verifies the catalog includes /Version matching the PDF version name.
// Both C# PdfName and Go core.Name hex-encode non-alphanumeric chars, so "1.5" is
// written as "/1#2E5" (dot → #2E).  This matches C# PdfCatalog behaviour exactly.
func TestCatalog_Version(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	// The Name "1.5" encodes the dot as #2E (matching C# PdfName behaviour)
	if !strings.Contains(output, "/Version") {
		t.Error("catalog missing /Version key")
	}
	if !strings.Contains(output, "/1#2E5") {
		t.Errorf("expected version name '/1#2E5' (1.5 with dot hex-encoded), output does not contain it")
	}
}
