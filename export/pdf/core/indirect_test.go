package core

import (
	"bytes"
	"strings"
	"testing"
)

func TestIndirectObject_Type(t *testing.T) {
	o := &IndirectObject{}
	if o.Type() != TypeIndirect {
		t.Fatalf("expected TypeIndirect, got %q", o.Type())
	}
}

func TestIndirectObject_WriteTo_Basic(t *testing.T) {
	o := &IndirectObject{
		Number:     1,
		Generation: 0,
		Value:      NewBoolean(true),
	}
	var buf bytes.Buffer
	n, err := o.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "1 0 obj\ntrue\nendobj\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if n != int64(len(want)) {
		t.Fatalf("byte count: got %d want %d", n, int64(len(want)))
	}
}

func TestIndirectObject_WriteTo_NilValue(t *testing.T) {
	o := &IndirectObject{Number: 5, Generation: 2, Value: nil}
	var buf bytes.Buffer
	_, err := o.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.HasPrefix(got, "5 2 obj\n") {
		t.Fatalf("unexpected output: %q", got)
	}
	if !strings.HasSuffix(got, "\nendobj\n") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestIndirectObject_Reference(t *testing.T) {
	o := &IndirectObject{Number: 7, Generation: 0}
	got := o.Reference()
	want := "7 0 R"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestIndirectObject_WriteTo_WithDictionary(t *testing.T) {
	d := NewDictionary()
	d.Add("Type", NewName("Page"))
	o := &IndirectObject{Number: 2, Generation: 0, Value: d}
	var buf bytes.Buffer
	_, err := o.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "/Type /Page") {
		t.Fatalf("expected dictionary content in output, got %q", got)
	}
}

func TestCountWriter(t *testing.T) {
	var buf bytes.Buffer
	cw := &countWriter{w: &buf}
	payload := []byte("hello")
	n, err := cw.Write(payload)
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("expected 5 got %d", n)
	}
	if cw.n != 5 {
		t.Fatalf("cumulative count: expected 5 got %d", cw.n)
	}
	cw.Write([]byte(" world"))
	if cw.n != 11 {
		t.Fatalf("cumulative count after second write: expected 11 got %d", cw.n)
	}
}
