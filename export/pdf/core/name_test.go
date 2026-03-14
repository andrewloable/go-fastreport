package core

import (
	"bytes"
	"testing"
)

func TestName_Type(t *testing.T) {
	n := NewName("Type")
	if n.Type() != TypeName {
		t.Fatalf("expected TypeName, got %q", n.Type())
	}
}

func TestNewName(t *testing.T) {
	n := NewName("FlateDecode")
	if n.Value != "FlateDecode" {
		t.Fatalf("expected 'FlateDecode', got %q", n.Value)
	}
}

func TestName_WriteTo_Empty(t *testing.T) {
	n := NewName("")
	var buf bytes.Buffer
	nn, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected empty output for empty name, got %q", buf.String())
	}
	if nn != 0 {
		t.Fatalf("expected 0 bytes returned for empty name, got %d", nn)
	}
}

func TestName_WriteTo_AlphaNumeric(t *testing.T) {
	n := NewName("FlateDecode")
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "/FlateDecode"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestName_WriteTo_SpecialChars(t *testing.T) {
	// Space (0x20) should be encoded as #20
	n := NewName("my name")
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "/my#20name"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestName_WriteTo_NonAscii(t *testing.T) {
	// Hash (#) character (0x23) should be encoded as #23
	n := NewName("a#b")
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "/a#23b"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestName_WriteTo_AllDigitsAndLetters(t *testing.T) {
	n := NewName("ABC123xyz")
	var buf bytes.Buffer
	_, err := n.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "/ABC123xyz"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestName_WriteTo_ByteCountMatches(t *testing.T) {
	n := NewName("Type")
	var buf bytes.Buffer
	nn, _ := n.WriteTo(&buf)
	if nn != int64(buf.Len()) {
		t.Fatalf("byte count mismatch: WriteTo=%d buf.Len=%d", nn, buf.Len())
	}
}

func TestIsNameRegular(t *testing.T) {
	cases := []struct {
		c    byte
		want bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{' ', false},
		{'/', false},
		{'#', false},
		{'(', false},
	}
	for _, tc := range cases {
		got := isNameRegular(tc.c)
		if got != tc.want {
			t.Errorf("isNameRegular(%q) = %v, want %v", tc.c, got, tc.want)
		}
	}
}
