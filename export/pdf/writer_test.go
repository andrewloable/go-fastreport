package pdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/export/pdf/core"
)

func TestNewWriter(t *testing.T) {
	w := NewWriter()
	if w == nil {
		t.Fatal("NewWriter returned nil")
	}
	if w.nextObjNum != 1 {
		t.Errorf("expected nextObjNum=1, got %d", w.nextObjNum)
	}
}

func TestWriter_NewObject(t *testing.T) {
	w := NewWriter()
	obj1 := w.NewObject(core.NewName("Test"))
	obj2 := w.NewObject(core.NewName("Test2"))

	if obj1.Number != 1 {
		t.Errorf("expected first object number 1, got %d", obj1.Number)
	}
	if obj2.Number != 2 {
		t.Errorf("expected second object number 2, got %d", obj2.Number)
	}
	if len(w.objects) != 2 {
		t.Errorf("expected 2 objects registered, got %d", len(w.objects))
	}
}

func TestWriter_NewObject_Generation(t *testing.T) {
	w := NewWriter()
	obj := w.NewObject(core.NewInt(42))
	if obj.Generation != 0 {
		t.Errorf("expected generation 0, got %d", obj.Generation)
	}
}

func TestWriter_Write_Header(t *testing.T) {
	w := NewWriter()
	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.HasPrefix(output, "%PDF-1.4\n") {
		t.Errorf("output does not start with PDF header, got: %q", output[:min(20, len(output))])
	}
}

func TestWriter_Write_EOF(t *testing.T) {
	w := NewWriter()
	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	// "%%EOF" is the PDF end-of-file marker
	eofMarker := "%" + "%EOF"
	if !strings.Contains(output, eofMarker) {
		t.Error("output does not contain the PDF end-of-file marker")
	}
}

func TestWriter_Write_Xref(t *testing.T) {
	w := NewWriter()
	w.NewObject(core.NewName("Foo"))
	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "xref") {
		t.Error("output does not contain xref table")
	}
	if !strings.Contains(output, "startxref") {
		t.Error("output does not contain startxref")
	}
}

func TestWriter_Write_Trailer_WithCatalogAndInfo(t *testing.T) {
	w := NewWriter()

	pages := NewPages(w)
	_ = NewCatalog(w, pages)
	_ = NewInfo(w)

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "/Root") {
		t.Error("trailer missing /Root")
	}
	if !strings.Contains(output, "/Info") {
		t.Error("trailer missing /Info")
	}
	if !strings.Contains(output, "/Size") {
		t.Error("trailer missing /Size")
	}
}

func TestWriter_Write_Trailer_NoCatalogNoInfo(t *testing.T) {
	w := NewWriter()
	w.NewObject(core.NewBoolean(true))

	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if strings.Contains(output, "/Root") {
		t.Error("trailer should not have /Root when no catalog set")
	}
	if strings.Contains(output, "/Info") {
		t.Error("trailer should not have /Info when no info set")
	}
}

func TestWriter_Write_XrefFreeEntry(t *testing.T) {
	w := NewWriter()
	w.NewObject(&core.Null{})
	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "0000000000 65535 f") {
		t.Error("xref free entry for object 0 missing")
	}
}

func TestWriter_Write_ObjectsInOutput(t *testing.T) {
	w := NewWriter()
	w.NewObject(core.NewName("TestObj"))
	var buf bytes.Buffer
	if err := w.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "1 0 obj") {
		t.Error("output missing object header '1 0 obj'")
	}
	if !strings.Contains(output, "endobj") {
		t.Error("output missing 'endobj'")
	}
}

func TestWriter_SetCatalog_SetInfo(t *testing.T) {
	w := NewWriter()
	obj := w.NewObject(core.NewName("CatalogTest"))
	w.setCatalog(obj)
	if w.catalog != obj {
		t.Error("setCatalog did not set catalog")
	}

	obj2 := w.NewObject(core.NewName("InfoTest"))
	w.setInfo(obj2)
	if w.info != obj2 {
		t.Error("setInfo did not set info")
	}
}

func TestWriter_MultipleObjects_SequentialNumbers(t *testing.T) {
	w := NewWriter()
	for i := 0; i < 5; i++ {
		obj := w.NewObject(core.NewInt(i))
		if obj.Number != i+1 {
			t.Errorf("expected object number %d, got %d", i+1, obj.Number)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
