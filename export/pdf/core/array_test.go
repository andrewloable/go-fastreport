package core

import (
	"bytes"
	"testing"
)

func TestArray_Type(t *testing.T) {
	a := NewArray()
	if a.Type() != TypeArray {
		t.Fatalf("expected TypeArray, got %q", a.Type())
	}
}

func TestArray_NewArray_WithItems(t *testing.T) {
	a := NewArray(NewInt(1), NewInt(2), NewInt(3))
	if a.Len() != 3 {
		t.Fatalf("expected Len=3, got %d", a.Len())
	}
}

func TestArray_Add(t *testing.T) {
	a := NewArray()
	ret := a.Add(NewInt(10))
	if ret != a {
		t.Fatal("Add should return the receiver")
	}
	if a.Len() != 1 {
		t.Fatalf("expected Len=1, got %d", a.Len())
	}
}

func TestArray_WriteTo_Empty(t *testing.T) {
	a := NewArray()
	var buf bytes.Buffer
	n, err := a.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "[ ]"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if n != int64(len(want)) {
		t.Fatalf("byte count: got %d want %d", n, int64(len(want)))
	}
}

func TestArray_WriteTo_SingleItem(t *testing.T) {
	a := NewArray(NewInt(42))
	var buf bytes.Buffer
	_, err := a.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "[ 42 ]"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestArray_WriteTo_MultipleItems(t *testing.T) {
	a := NewArray(NewInt(1), NewInt(2), NewInt(3))
	var buf bytes.Buffer
	_, err := a.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "[ 1 2 3 ]"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestArray_WriteTo_Mixed(t *testing.T) {
	a := NewArray(NewName("Foo"), NewBoolean(true), &Null{})
	var buf bytes.Buffer
	_, err := a.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "[ /Foo true null ]"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestArray_WriteTo_ByteCountMatches(t *testing.T) {
	a := NewArray(NewInt(7), NewInt(8))
	var buf bytes.Buffer
	n, _ := a.WriteTo(&buf)
	if n != int64(buf.Len()) {
		t.Fatalf("byte count mismatch: WriteTo=%d buf.Len=%d", n, buf.Len())
	}
}

func TestArray_Len_Empty(t *testing.T) {
	a := NewArray()
	if a.Len() != 0 {
		t.Fatalf("expected 0, got %d", a.Len())
	}
}
