package core

import (
	"bytes"
	"testing"
)

func TestNull_Type(t *testing.T) {
	n := &Null{}
	if n.Type() != TypeNull {
		t.Fatalf("expected TypeNull, got %q", n.Type())
	}
}

func TestNull_WriteTo(t *testing.T) {
	n := &Null{}
	var buf bytes.Buffer
	nn, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "null" {
		t.Fatalf("got %q want %q", buf.String(), "null")
	}
	if nn != 4 {
		t.Fatalf("byte count: got %d want 4", nn)
	}
}

func TestNull_WriteTo_ByteCountMatches(t *testing.T) {
	n := &Null{}
	var buf bytes.Buffer
	nn, _ := n.WriteTo(&buf)
	if nn != int64(buf.Len()) {
		t.Fatalf("byte count mismatch: WriteTo=%d buf.Len=%d", nn, buf.Len())
	}
}
