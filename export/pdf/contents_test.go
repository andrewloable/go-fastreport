package pdf

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewContents(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)

	if c == nil {
		t.Fatal("NewContents returned nil")
	}
	if c.obj == nil {
		t.Error("Contents has nil indirect object")
	}
	if c.stream == nil {
		t.Error("Contents has nil stream")
	}
}

func TestContents_Write(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	data := []byte("BT /F1 12 Tf ET")
	c.Write(data)

	if c.buf.Len() == 0 {
		t.Error("Write did not append data to buffer")
	}
}

func TestContents_WriteString(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	c.WriteString("q\n")
	c.WriteString("Q\n")

	if c.buf.String() != "q\nQ\n" {
		t.Errorf("WriteString produced unexpected content: %q", c.buf.String())
	}
}

func TestContents_Finalize(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	c.WriteString("BT ET")
	c.Finalize()

	if string(c.stream.Data) != "BT ET" {
		t.Errorf("Finalize did not copy buffer to stream data, got %q", c.stream.Data)
	}
}

func TestContents_Finalize_Empty(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	c.Finalize()

	// Should not panic and stream.Data should be empty slice
	if len(c.stream.Data) != 0 {
		t.Errorf("expected empty stream data, got %d bytes", len(c.stream.Data))
	}
}

func TestContents_Write_Multiple(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	c.Write([]byte("q "))
	c.Write([]byte("Q"))
	c.Finalize()

	if string(c.stream.Data) != "q Q" {
		t.Errorf("expected 'q Q', got %q", c.stream.Data)
	}
}

func TestContents_NotCompressed(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)

	if c.stream.Compressed {
		t.Error("Contents stream should not be compressed by default")
	}
}

func TestContents_InOutput(t *testing.T) {
	w := NewWriter()
	pages := NewPages(w)
	page := NewPage(w, pages, 595, 842)

	contents := page.Contents()
	contents.WriteString("BT /F1 12 Tf (Hello) Tj ET\n")
	contents.Finalize()

	_ = NewCatalog(w, pages)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "BT /F1 12 Tf (Hello) Tj ET") {
		t.Error("output missing content stream data")
	}
}

func TestContents_Finalize_IsolatesBuffer(t *testing.T) {
	w := NewWriter()
	c := NewContents(w)
	c.WriteString("initial")
	c.Finalize()

	originalData := make([]byte, len(c.stream.Data))
	copy(originalData, c.stream.Data)

	// Modifying the buffer after Finalize should not affect stream.Data
	c.WriteString(" more")
	if string(c.stream.Data) != string(originalData) {
		t.Error("Finalize should produce an independent copy of the buffer")
	}
}
