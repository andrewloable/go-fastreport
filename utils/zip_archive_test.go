package utils

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

// TestZipArchive_empty verifies that a brand-new archive with no entries
// produces valid (non-nil, non-empty) ZIP bytes.
func TestZipArchive_empty(t *testing.T) {
	za := NewZipArchive()
	b, err := za.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty ZIP bytes for empty archive")
	}
	// Verify it is a parseable ZIP (empty central directory is valid).
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatalf("zip.NewReader error: %v", err)
	}
	if len(r.File) != 0 {
		t.Fatalf("expected 0 files, got %d", len(r.File))
	}
}

// TestZipArchive_addEntry verifies that AddEntry stores the correct content.
func TestZipArchive_addEntry(t *testing.T) {
	const (
		name    = "hello.txt"
		content = "hello world"
	)
	za := NewZipArchive()
	if err := za.AddEntry(name, []byte(content)); err != nil {
		t.Fatalf("AddEntry error: %v", err)
	}
	b, err := za.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatalf("zip.NewReader error: %v", err)
	}
	if len(r.File) != 1 {
		t.Fatalf("expected 1 file, got %d", len(r.File))
	}
	f := r.File[0]
	if f.Name != name {
		t.Errorf("expected name %q, got %q", name, f.Name)
	}
	rc, err := f.Open()
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer rc.Close()
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if string(got) != content {
		t.Errorf("expected content %q, got %q", content, string(got))
	}
}

// TestZipArchive_addEntryFromStream verifies that AddEntryFromStream stores
// the correct content when the source is an io.Reader.
func TestZipArchive_addEntryFromStream(t *testing.T) {
	const (
		name    = "stream.txt"
		content = "hello world"
	)
	za := NewZipArchive()
	if err := za.AddEntryFromStream(name, bytes.NewReader([]byte(content))); err != nil {
		t.Fatalf("AddEntryFromStream error: %v", err)
	}
	b, err := za.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatalf("zip.NewReader error: %v", err)
	}
	if len(r.File) != 1 {
		t.Fatalf("expected 1 file, got %d", len(r.File))
	}
	f := r.File[0]
	if f.Name != name {
		t.Errorf("expected name %q, got %q", name, f.Name)
	}
	rc, err := f.Open()
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer rc.Close()
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if string(got) != content {
		t.Errorf("expected content %q, got %q", content, string(got))
	}
}

// TestZipArchive_multipleEntries verifies that all entries added are present
// in the final archive with the correct names and contents.
func TestZipArchive_multipleEntries(t *testing.T) {
	entries := []struct {
		name    string
		content string
	}{
		{"a.txt", "alpha"},
		{"b/c.txt", "beta"},
		{"d.bin", "gamma"},
	}

	za := NewZipArchive()
	for _, e := range entries {
		if err := za.AddEntry(e.name, []byte(e.content)); err != nil {
			t.Fatalf("AddEntry(%q) error: %v", e.name, err)
		}
	}
	b, err := za.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatalf("zip.NewReader error: %v", err)
	}
	if len(r.File) != len(entries) {
		t.Fatalf("expected %d files, got %d", len(entries), len(r.File))
	}

	// Build a map for order-independent lookup.
	byName := make(map[string]string, len(r.File))
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("Open(%q) error: %v", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("ReadAll(%q) error: %v", f.Name, err)
		}
		byName[f.Name] = string(data)
	}

	for _, e := range entries {
		got, ok := byName[e.name]
		if !ok {
			t.Errorf("entry %q not found in archive", e.name)
			continue
		}
		if got != e.content {
			t.Errorf("entry %q: expected %q, got %q", e.name, e.content, got)
		}
	}
}
