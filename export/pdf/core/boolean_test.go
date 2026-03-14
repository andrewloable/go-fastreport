package core

import (
	"bytes"
	"testing"
)

func TestBoolean_Type(t *testing.T) {
	b := NewBoolean(true)
	if b.Type() != TypeBoolean {
		t.Fatalf("expected TypeBoolean, got %q", b.Type())
	}
}

func TestNewBoolean_True(t *testing.T) {
	b := NewBoolean(true)
	if !b.Value {
		t.Fatal("expected Value=true")
	}
}

func TestNewBoolean_False(t *testing.T) {
	b := NewBoolean(false)
	if b.Value {
		t.Fatal("expected Value=false")
	}
}

func TestBoolean_WriteTo_True(t *testing.T) {
	b := NewBoolean(true)
	var buf bytes.Buffer
	n, err := b.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "true" {
		t.Fatalf("got %q want %q", buf.String(), "true")
	}
	if n != 4 {
		t.Fatalf("byte count: got %d want 4", n)
	}
}

func TestBoolean_WriteTo_False(t *testing.T) {
	b := NewBoolean(false)
	var buf bytes.Buffer
	n, err := b.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "false" {
		t.Fatalf("got %q want %q", buf.String(), "false")
	}
	if n != 5 {
		t.Fatalf("byte count: got %d want 5", n)
	}
}

func TestBoolean_WriteTo_ByteCountMatches(t *testing.T) {
	for _, v := range []bool{true, false} {
		b := NewBoolean(v)
		var buf bytes.Buffer
		nn, _ := b.WriteTo(&buf)
		if nn != int64(buf.Len()) {
			t.Fatalf("byte count mismatch for %v: WriteTo=%d buf.Len=%d", v, nn, buf.Len())
		}
	}
}
