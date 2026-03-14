package core

import (
	"bytes"
	"strings"
	"testing"
)

func TestDictionary_Type(t *testing.T) {
	d := NewDictionary()
	if d.Type() != TypeDictionary {
		t.Fatalf("expected TypeDictionary, got %q", d.Type())
	}
}

func TestDictionary_AddAndGet(t *testing.T) {
	d := NewDictionary()
	d.Add("Type", NewName("Page"))
	got := d.Get("Type")
	if got == nil {
		t.Fatal("expected non-nil value")
	}
	n, ok := got.(*Name)
	if !ok {
		t.Fatalf("expected *Name, got %T", got)
	}
	if n.Value != "Page" {
		t.Fatalf("expected 'Page', got %q", n.Value)
	}
}

func TestDictionary_GetMissing(t *testing.T) {
	d := NewDictionary()
	if d.Get("Missing") != nil {
		t.Fatal("expected nil for missing key")
	}
}

func TestDictionary_Len(t *testing.T) {
	d := NewDictionary()
	if d.Len() != 0 {
		t.Fatalf("expected 0, got %d", d.Len())
	}
	d.Add("A", NewNull())
	d.Add("B", NewNull())
	if d.Len() != 2 {
		t.Fatalf("expected 2, got %d", d.Len())
	}
}

func TestDictionary_AddReplace(t *testing.T) {
	d := NewDictionary()
	d.Add("Key", NewInt(1))
	d.Add("Key", NewInt(99))
	if d.Len() != 1 {
		t.Fatalf("expected 1 entry after replace, got %d", d.Len())
	}
	n := d.Get("Key").(*Numeric)
	if int(n.Value) != 99 {
		t.Fatalf("expected 99, got %v", n.Value)
	}
}

func TestDictionary_WriteTo_Empty(t *testing.T) {
	d := NewDictionary()
	var buf bytes.Buffer
	n, err := d.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "<< >>"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if n != int64(len(want)) {
		t.Fatalf("byte count: got %d want %d", n, int64(len(want)))
	}
}

func TestDictionary_WriteTo_OneEntry(t *testing.T) {
	d := NewDictionary()
	d.Add("Type", NewName("Catalog"))
	var buf bytes.Buffer
	_, err := d.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "/Type /Catalog") {
		t.Fatalf("expected /Type /Catalog in output, got %q", got)
	}
}

func TestDictionary_WriteTo_MultipleEntries_OrderPreserved(t *testing.T) {
	d := NewDictionary()
	d.Add("A", NewInt(1))
	d.Add("B", NewInt(2))
	d.Add("C", NewInt(3))
	var buf bytes.Buffer
	_, err := d.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	posA := strings.Index(got, "/A")
	posB := strings.Index(got, "/B")
	posC := strings.Index(got, "/C")
	if posA < 0 || posB < 0 || posC < 0 {
		t.Fatalf("missing keys in output: %q", got)
	}
	if !(posA < posB && posB < posC) {
		t.Fatalf("insertion order not preserved: %q", got)
	}
}

func TestDictionary_WriteTo_CountMatchesLen(t *testing.T) {
	d := NewDictionary()
	d.Add("X", NewBoolean(false))
	var buf bytes.Buffer
	n, _ := d.WriteTo(&buf)
	if n != int64(buf.Len()) {
		t.Fatalf("byte count mismatch: WriteTo returned %d, buf.Len()=%d", n, buf.Len())
	}
}

// NewNull is a helper used in tests across files.
func NewNull() *Null { return &Null{} }
